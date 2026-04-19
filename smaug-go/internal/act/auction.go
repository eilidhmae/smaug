// Full `do_auction` state machine — non-GSC variant, ports
// `src/act_obj.c:3775-4240` (do_auction) with per-branch helpers.
// Auction tick lives in internal/game/update_auction.go (the per-pulse
// counter is managed there so the dispatch pattern matches the existing
// pulseTick / pulseMobile / pulseSave peers).
//
// Shared reader: BroadcastAuction is called from both the command and
// the tick; it's the `talk_auction` helper (`src/act_comm.c:4309`) and
// retains the Tier-9 filter chain verbatim.
package act

import (
	"fmt"
	"strings"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// auctionMinTrust is the trust threshold C `talk_auction` applies to
// receivers (`act_comm.c:4322`). Pulled out as a named constant rather
// than a magic literal so the mapping back to C is obvious.
const auctionMinTrust = 5

// auctionLevelFloor ports C's `ch->level < 3` gate at act_obj.c:3794.
const auctionLevelFloor = 3

// auctionHourOpen / auctionHourClose port the time gate at act_obj.c:3801.
// C: `(time_info.hour > 18 || time_info.hour < 9)` — hours 9 AM to 6 PM
// inclusive.
const (
	auctionHourOpen  = 9
	auctionHourClose = 18
)

// auctionMaxBid ports the 2-billion ceiling at act_obj.c:4038.
const auctionMaxBid = 2000000000

// auctionMinIncrement ports the +10000 step at act_obj.c:4024.
const auctionMinIncrement = 10000

// BroadcastAuction delivers `message` as an AUCTION: channel line to every
// currently-playing descriptor that passes the C-matching filter chain:
// Trust >= 5, Deaf[AUCTION] not set, current room is NOT ROOM_SILENCE.
// Mirrors C `talk_auction` at src/act_comm.c:4309.
func BroadcastAuction(message string) {
	if WorldRef == nil {
		return
	}
	for _, d := range WorldRef.Descriptors {
		if d == nil || d.Connected != types.CON_PLAYING || d.Character == nil {
			continue
		}
		ch := d.Character
		if ch.GetTrust() < auctionMinTrust {
			continue
		}
		if ch.Deaf.IsSet(types.CHANNEL_AUCTION) {
			continue
		}
		if ch.InRoom != nil && ch.InRoom.RoomFlags.IsSet(types.ROOM_SILENCE) {
			continue
		}
		ch.Sendf("Auction: %s\n\r", message)
	}
}

// DoAuction ports `do_auction` (non-GSC) at src/act_obj.c:3775-4240. Four
// intents:
//  1. `auction` with no args — show current-item info or "nothing auctioned"
//  2. `auction stop` (immortal only) — cancel + refund
//  3. `auction bid <amount> [item-keyword]` — place a bid
//  4. `auction <item> [min-bet]` — start auctioning a carried item
func DoAuction(ch *types.CharData, argument string) {
	if ch == nil {
		return
	}

	arg1, rest1 := util.OneArgument(argument)
	arg2, rest2 := util.OneArgument(rest1)
	arg3, _ := util.OneArgument(rest2)

	// NPC silent return at act_obj.c:3791-3792.
	if ch.IsNPC() {
		return
	}

	// Level gate at act_obj.c:3794-3798.
	if ch.Level < auctionLevelFloor {
		ch.Send("You must be at least level three to use the auction...\n\r")
		return
	}

	// Time gate at act_obj.c:3801-3808. Bypassed if an auction is already
	// running OR caller is immortal.
	if WorldRef != nil && WorldRef.Auction != nil {
		hour := WorldRef.TimeInfo.Hour
		if (hour > auctionHourClose || hour < auctionHourOpen) &&
			WorldRef.Auction.Item == nil &&
			!ch.IsImmortal() {
			ch.Send("\n\rThe auctioneer works between the hours of 9 AM and 6 PM\n\r")
			return
		}
	}

	// Empty arg1 — show info block (act_obj.c:3810-3955).
	if arg1 == "" {
		auctionInfo(ch)
		return
	}

	// Immortal `stop` (act_obj.c:3957-3982).
	if ch.IsImmortal() && strings.EqualFold(arg1, "stop") {
		auctionStop(ch)
		return
	}

	// Bid (act_obj.c:3983-4077).
	if strings.EqualFold(arg1, "bid") {
		auctionBid(ch, arg2, arg3)
		return
	}

	// Start new auction (act_obj.c:4078-4240).
	auctionStart(ch, arg1, arg2)
}

// auctionInfo ports the arg1==empty branch at act_obj.c:3810-3955.
// Shows either current auction details or "nothing auctioned".
func auctionInfo(ch *types.CharData) {
	auc := WorldRef.Auction
	if auc == nil || auc.Item == nil {
		ch.Send("\n\rThere is nothing being auctioned right now.  What would you like to auction?\n\r")
		return
	}
	obj := auc.Item

	if auc.Bet > 0 {
		ch.Sendf("\n\rCurrent bid on this item is %s gold.\n\r", util.NumPunct(auc.Bet))
	} else {
		ch.Send("\n\rNo bids on this item have been received.\n\r")
	}

	// Core item description (act_obj.c:3827-3835).
	ch.Sendf("Object '%s' is %s, special properties: %s\n\rIts weight is %d, value is %d, and level is %d.\n\r",
		obj.Name,
		aoran(itemTypeName(obj.ItemType)),
		auctionExtraBitName(obj.ExtraFlags),
		obj.Weight,
		obj.GoldCost,
		obj.Level,
	)

	// Wear-location line (act_obj.c:3836-3838) — non-LIGHT items with
	// wear_flags beyond TAKE (WearFlags > 1 after the -1 subtract C uses).
	if obj.ItemType != types.ITEM_LIGHT && obj.WearFlags > 1 {
		ch.Sendf("Item's wear location: %s\n\r", auctionWearLocString(obj.WearFlags))
	}

	// Per-item-type info block (G4b — act_obj.c:3842-3920).
	auctionItemTypeInfo(ch, obj)

	// Contents for containers (act_obj.c:3927-3934).
	if (obj.ItemType == types.ITEM_CONTAINER ||
		obj.ItemType == types.ITEM_KEYRING ||
		obj.ItemType == types.ITEM_QUIVER) && len(obj.Contents) > 0 {
		ch.Send("Contents:\n\r")
		for _, c := range obj.Contents {
			ch.Sendf("  %s\n\r", c.ShortDescr)
		}
	}

	// Immortal seller/buyer disclosure (act_obj.c:3936-3944).
	if ch.IsImmortal() {
		sellerName := "nobody"
		buyerName := "nobody"
		if auc.Seller != nil {
			sellerName = auc.Seller.Name
		}
		if auc.Buyer != nil {
			buyerName = auc.Buyer.Name
		}
		ch.Sendf("Seller: %s.  Bidder: %s.  Round: %d.\n\r",
			sellerName, buyerName, auc.Going+1)
		ch.Sendf("Time left in round: %d.\n\r", auc.Pulse)
	}
}

// auctionStop ports the immortal `stop` branch at act_obj.c:3957-3982.
// Cancels the auction, returns item to seller, refunds distinct buyer.
func auctionStop(ch *types.CharData) {
	auc := WorldRef.Auction
	if auc == nil || auc.Item == nil {
		ch.Send("There is no auction to stop.\n\r")
		return
	}
	msg := fmt.Sprintf("Sale of %s has been stopped by an Immortal.", auc.Item.ShortDescr)
	BroadcastAuction(msg)

	if auc.Seller != nil {
		handler.ObjToChar(auc.Item, auc.Seller)
	}

	// Refund any distinct buyer (act_obj.c:3974-3978).
	if auc.Buyer != nil && auc.Buyer != auc.Seller {
		auc.Buyer.Gold += auc.Bet
		auc.Buyer.Send("Your money has been returned.\n\r")
	}

	// Clear active auction but preserve History ring + HistTimer.
	auc.Item = nil
	auc.Seller = nil
	auc.Buyer = nil
	auc.Bet = 0
	auc.Going = 0
	auc.Pulse = 0
	auc.Starting = 0
}

// auctionBid ports the bid branch at act_obj.c:3983-4077.
func auctionBid(ch *types.CharData, arg2, arg3 string) {
	auc := WorldRef.Auction
	if auc == nil || auc.Item == nil {
		ch.Send("There isn't anything being auctioned right now.\n\r")
		return
	}

	if ch.Level < auc.Item.Level {
		ch.Send("This object's level is too high for your use.\n\r")
		return
	}

	if ch == auc.Seller {
		ch.Send("You can't bid on your own item!\n\r")
		return
	}

	if arg2 == "" {
		ch.Send("Bid how much?\n\r")
		return
	}

	newbet := util.ParseBet(auc.Bet, arg2)

	if newbet < auc.Starting {
		ch.Send("You must place a bid that is higher than the starting bet.\n\r")
		return
	}

	// +10000 step (act_obj.c:4024).
	if newbet < auc.Bet || newbet < (auc.Bet+auctionMinIncrement) {
		ch.Send("You must at least bid 10000 coins over the current bid.\n\r")
		return
	}

	if newbet > ch.Gold {
		ch.Send("You don't have that much money!\n\r")
		return
	}

	if newbet > auctionMaxBid {
		ch.Send("You can't bid over 2 billion coins.\n\r")
		return
	}

	// Item-keyword match (act_obj.c:4045-4050).
	if arg3 != "" && !util.IsName(arg3, auc.Item.Name) {
		ch.Send("That item is not being auctioned right now.\n\r")
		return
	}

	// Refund previous distinct bidder (act_obj.c:4054-4055).
	if auc.Buyer != nil && auc.Buyer != auc.Seller {
		auc.Buyer.Gold += auc.Bet
	}

	ch.Gold -= newbet
	auc.Buyer = ch
	auc.Bet = newbet
	auc.Going = 0
	auc.Pulse = types.PULSE_AUCTION

	msg := fmt.Sprintf("A bid of %s gold has been received on %s.",
		util.NumPunct(newbet), auc.Item.ShortDescr)
	BroadcastAuction(msg)
}

// auctionStart ports the start-new branch at act_obj.c:4078-4240.
func auctionStart(ch *types.CharData, arg1, arg2 string) {
	// C: ms_find_obj — generalized drunk/mental-state check at
	// handler.c:2941. Go has not wired mental-state side-effects into any
	// object command; port as a no-op consistent with the existing
	// omission.

	auc := WorldRef.Auction

	obj := handler.GetObjCarry(ch, arg1)
	if obj == nil {
		ch.Send("You aren't carrying that.\n\r")
		return
	}

	// NoAuction blacklist (act_obj.c:4090-4096).
	if !ch.IsImmortal() {
		for _, vnum := range WorldRef.NoAuction {
			if obj.IndexData != nil && obj.IndexData.Vnum == vnum {
				ch.Send("That item cannot be auctioned.\n\r")
				return
			}
		}
	}

	// Type mismatch (act_obj.c:4099-4104).
	if obj.IndexData != nil && obj.ItemType != obj.IndexData.ItemType {
		ch.Send("This object is too modified to auction.  Please contact an immortal.\n\r")
		return
	}

	// Decaying items (act_obj.c:4107-4111).
	if obj.Timer > 0 {
		ch.Send("You can't auction objects that are decaying.\n\r")
		return
	}

	// Clan/permanent (act_obj.c:4113-4123).
	if obj.ExtraFlags.IsSet(types.ITEM_CLANOBJECT) {
		ch.Send("You can't auction clan items.\n\r")
		return
	}
	if obj.ExtraFlags.IsSet(types.ITEM_PERMANENT) {
		ch.Send("This item cannot leave your possession.\n\r")
		return
	}

	// History collision (act_obj.c:4125-4134).
	if obj.IndexData != nil {
		for i := 0; i < types.AUCTION_MEM; i++ {
			if auc.History[i] == nil {
				break
			}
			if auc.History[i] == obj.IndexData {
				ch.Send("Such an item has been auctioned recently, try again later.\n\r")
				return
			}
		}
	}

	// Min-bet parsing (act_obj.c:4137-4155).
	if arg2 == "" {
		arg2 = "0"
	}
	if !util.IsNumber(arg2) {
		ch.Send("You must input a number at which to start the auction.\n\r")
		return
	}
	minBet := 0
	for _, r := range arg2 {
		if r == '-' {
			minBet = -1
			break
		}
	}
	// The C path does atoi(arg2) — which handles leading '-' by returning
	// a negative; Go's IsNumber already rejects non-digit/sign, so parse
	// via util.ParseBet on the bare-digit path would lose the sign. Use
	// strconv below.
	if minBet != -1 {
		minBet = atoiSafe(arg2)
	}
	if minBet < 0 {
		ch.Send("You can't auction something for less than 0 gold!\n\r")
		return
	}

	// Second-auction guard (act_obj.c:4232-4239). C does this INSIDE the
	// switch default arm — after validating everything else, it checks
	// `auction->item == NULL`. We preserve the same ordering by placing
	// the guard right before the item-type whitelist switch: if there's
	// already an auction, bail out with the Try-again-later message.
	if auc.Item != nil {
		util.Act(types.AT_TELL,
			"Try again later - $p is being auctioned right now!",
			ch, nil, auc.Item, nil, types.TO_CHAR)
		if !ch.IsImmortal() {
			ch.Wait = types.PULSE_VIOLENCE
		}
		return
	}

	// Item-type whitelist switch (act_obj.c:4158-4231). Unsupported types
	// fall to the default "cannot auction" branch.
	if !auctionTypeAllowed(obj.ItemType) {
		util.Act(types.AT_TELL, "You cannot auction $Ts.",
			ch, nil, nil, itemTypeName(obj.ItemType), types.TO_CHAR)
		return
	}

	// C calls separate_obj(obj) here; Go has no object stacking (every
	// ObjData is count-1), so the call is a no-op.
	handler.ObjFromChar(obj)

	auc.Item = obj
	auc.Bet = 0
	auc.Buyer = ch
	auc.Seller = ch
	auc.Pulse = types.PULSE_AUCTION
	auc.Going = 0
	auc.Starting = minBet

	// History ring rotation (act_obj.c:4210-4216): shift right, put new
	// entry at index 0. C's memmove on an overlapping region is
	// equivalent to a backward-slice copy; Go's `copy(dst[1:], src[:N-1])`
	// is well-defined even when both are views of the same array.
	for i := types.AUCTION_MEM - 1; i > 0; i-- {
		auc.History[i] = auc.History[i-1]
	}
	if obj.IndexData != nil {
		auc.History[0] = obj.IndexData
	}
	auc.HistTimer = 0

	if auc.Starting > 0 {
		auc.Bet = auc.Starting
	}

	msg := fmt.Sprintf("A new item is being auctioned: %s at %d gold.",
		obj.ShortDescr, auc.Starting)
	BroadcastAuction(msg)
}

// atoiSafe is strconv.Atoi with a zero-fallback (matches C atoi).
func atoiSafe(s string) int {
	n := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			n = n*10 + int(r-'0')
		}
	}
	return n
}

// auctionTypeAllowed mirrors the C case-list at act_obj.c:4169-4196. Any
// item type not in this list cannot be auctioned (default arm).
func auctionTypeAllowed(t int) bool {
	switch t {
	case types.ITEM_PAPER,
		types.ITEM_LIGHT,
		types.ITEM_TREASURE,
		types.ITEM_POTION,
		types.ITEM_KEYRING,
		types.ITEM_QUIVER,
		types.ITEM_DRINK_CON,
		types.ITEM_FOOD,
		types.ITEM_COOK,
		types.ITEM_PEN,
		types.ITEM_BOAT,
		types.ITEM_PILL,
		types.ITEM_PIPE,
		types.ITEM_HERB_CON,
		types.ITEM_INCENSE,
		types.ITEM_FIRE,
		types.ITEM_RUNEPOUCH,
		types.ITEM_MAP,
		types.ITEM_BOOK,
		types.ITEM_RUNE,
		types.ITEM_MATCH,
		types.ITEM_HERB,
		types.ITEM_WEAPON,
		types.ITEM_MISSILE_WEAPON,
		types.ITEM_ARMOR,
		types.ITEM_STAFF,
		types.ITEM_WAND,
		types.ITEM_SCROLL:
		return true
	}
	return false
}

// auctionItemTypeInfo ports the per-item-type info switch at
// act_obj.c:3842-3920 for the G4b info block. Obscure cases collapse to
// the C default (no extra output).
func auctionItemTypeInfo(ch *types.CharData, obj *types.ObjData) {
	switch obj.ItemType {
	case types.ITEM_CONTAINER, types.ITEM_KEYRING, types.ITEM_QUIVER:
		v := obj.Value[0]
		var cap string
		switch {
		case v < 76:
			cap = "have a small capacity"
		case v < 150:
			cap = "have a small to medium capacity"
		case v < 300:
			cap = "have a medium capacity"
		case v < 500:
			cap = "have a medium to large capacity"
		case v < 751:
			cap = "have a large capacity"
		default:
			cap = "have a giant capacity"
		}
		ch.Sendf("%s appears to %s.\n\r", capitalize(obj.ShortDescr), cap)

	case types.ITEM_PILL, types.ITEM_SCROLL, types.ITEM_POTION:
		var b strings.Builder
		fmt.Fprintf(&b, "Level %d spells of:", obj.Value[0])
		for i := 1; i <= 3; i++ {
			name := auctionSpellName(obj.Value[i])
			if name != "" {
				fmt.Fprintf(&b, " '%s'", name)
			}
		}
		b.WriteString(".\n\r")
		ch.Send(b.String())

	case types.ITEM_WAND, types.ITEM_STAFF:
		var b strings.Builder
		fmt.Fprintf(&b, "Has %d(%d) charges of level %d",
			obj.Value[1], obj.Value[2], obj.Value[0])
		if name := auctionSpellName(obj.Value[3]); name != "" {
			fmt.Fprintf(&b, " '%s'", name)
		}
		b.WriteString(".\n\r")
		ch.Send(b.String())

	case types.ITEM_MISSILE_WEAPON, types.ITEM_WEAPON:
		avg := (obj.Value[1] + obj.Value[2]) / 2
		poison := ""
		if obj.ExtraFlags.IsSet(types.ITEM_POISONED) {
			poison = "\n\rThis weapon is poisoned."
		}
		ch.Sendf("Damage is %d to %d (average %d).%s\n\r",
			obj.Value[1], obj.Value[2], avg, poison)

	case types.ITEM_ARMOR:
		ch.Sendf("Armor class is %d.\n\r", obj.Value[0])
	}
}

// auctionSpellName resolves a spell-index `sn` to a printable name, or
// empty string if out of range / nil. Mirrors the `skill_table[sn]->name`
// access pattern at act_obj.c:3869/3876/3883/3899.
func auctionSpellName(sn int) string {
	if sn < 0 || WorldRef == nil {
		return ""
	}
	if sn >= len(WorldRef.Skills) {
		return ""
	}
	sk := WorldRef.Skills[sn]
	if sk == nil {
		return ""
	}
	return sk.Name
}

// itemTypeName returns the human-readable item-type label used in the C
// item_type_name() call. Go has no central lookup table; this is a
// minimal ported subset covering every type allowed by
// auctionTypeAllowed plus common fallbacks. Unknown types fall back to
// a generic phrasing matching C's item_type_name default.
func itemTypeName(t int) string {
	switch t {
	case types.ITEM_LIGHT:
		return "light"
	case types.ITEM_SCROLL:
		return "scroll"
	case types.ITEM_WAND:
		return "wand"
	case types.ITEM_STAFF:
		return "staff"
	case types.ITEM_WEAPON:
		return "weapon"
	case types.ITEM_MISSILE_WEAPON:
		return "missile weapon"
	case types.ITEM_TREASURE:
		return "treasure"
	case types.ITEM_ARMOR:
		return "armor"
	case types.ITEM_POTION:
		return "potion"
	case types.ITEM_FURNITURE:
		return "furniture"
	case types.ITEM_TRASH:
		return "trash"
	case types.ITEM_CONTAINER:
		return "container"
	case types.ITEM_DRINK_CON:
		return "drink container"
	case types.ITEM_KEY:
		return "key"
	case types.ITEM_FOOD:
		return "food"
	case types.ITEM_MONEY:
		return "money"
	case types.ITEM_BOAT:
		return "boat"
	case types.ITEM_CORPSE_NPC:
		return "NPC corpse"
	case types.ITEM_CORPSE_PC:
		return "PC corpse"
	case types.ITEM_FOUNTAIN:
		return "fountain"
	case types.ITEM_PILL:
		return "pill"
	case types.ITEM_BLOOD:
		return "blood"
	case types.ITEM_BLOODSTAIN:
		return "bloodstain"
	case types.ITEM_SCRAPS:
		return "scraps"
	case types.ITEM_PIPE:
		return "pipe"
	case types.ITEM_HERB_CON:
		return "herb container"
	case types.ITEM_HERB:
		return "herb"
	case types.ITEM_INCENSE:
		return "incense"
	case types.ITEM_FIRE:
		return "fire"
	case types.ITEM_BOOK:
		return "book"
	case types.ITEM_SWITCH:
		return "switch"
	case types.ITEM_LEVER:
		return "lever"
	case types.ITEM_PUDDLE:
		return "puddle"
	case types.ITEM_PAPER:
		return "paper"
	case types.ITEM_RUNE:
		return "rune"
	case types.ITEM_RUNEPOUCH:
		return "rune pouch"
	case types.ITEM_MATCH:
		return "match"
	case types.ITEM_TRAP:
		return "trap"
	case types.ITEM_MAP:
		return "map"
	case types.ITEM_PORTAL:
		return "portal"
	case types.ITEM_KEYRING:
		return "keyring"
	case types.ITEM_QUIVER:
		return "quiver"
	case types.ITEM_COOK:
		return "cooked food"
	case types.ITEM_PEN:
		return "pen"
	}
	return "item"
}

// auctionExtraBitName returns the special-properties phrase used in the
// auction info block. C's `extra_bit_name` walks the ExtraFlags bitvector
// and emits space-separated flag names. Go's BitVector provides IsSet
// but no name table local to `act`; a minimal set covers the common
// player-visible flags. Unknown flags collapse to "none" to match
// empty-bitvector output.
func auctionExtraBitName(flags types.BitVector) string {
	if flags == (types.BitVector{}) {
		return "none"
	}
	var parts []string
	named := []struct {
		bit  int
		name string
	}{
		{types.ITEM_GLOW, "glow"},
		{types.ITEM_HUM, "hum"},
		{types.ITEM_MAGIC, "magic"},
		{types.ITEM_INVIS, "invis"},
		{types.ITEM_BLESS, "bless"},
		{types.ITEM_POISONED, "poisoned"},
		{types.ITEM_HIDDEN, "hidden"},
		{types.ITEM_PERMANENT, "permanent"},
	}
	for _, n := range named {
		if flags.IsSet(n.bit) {
			parts = append(parts, n.name)
		}
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, " ")
}

// auctionWearLocString summarizes the `w_flags` output at act_obj.c:3838.
// A full flag_string port lives in util; this is a compact label set
// sufficient for the info line. ObjData.WearFlags is an int but the
// bitmask constants in types/constants.go are uint32, so we lift through
// uint32 for the bitwise AND.
func auctionWearLocString(wf int) string {
	var parts []string
	bits := uint32(wf)
	locs := []struct {
		bit  uint32
		name string
	}{
		{types.ITEM_WEAR_FINGER, "finger"},
		{types.ITEM_WEAR_NECK, "neck"},
		{types.ITEM_WEAR_BODY, "body"},
		{types.ITEM_WEAR_HEAD, "head"},
		{types.ITEM_WEAR_LEGS, "legs"},
		{types.ITEM_WEAR_FEET, "feet"},
		{types.ITEM_WEAR_HANDS, "hands"},
		{types.ITEM_WEAR_ARMS, "arms"},
		{types.ITEM_WEAR_SHIELD, "shield"},
		{types.ITEM_WEAR_ABOUT, "about"},
		{types.ITEM_WEAR_WAIST, "waist"},
		{types.ITEM_WEAR_WRIST, "wrist"},
		{types.ITEM_WIELD, "wield"},
		{types.ITEM_HOLD, "hold"},
	}
	for _, l := range locs {
		if bits&l.bit != 0 {
			parts = append(parts, l.name)
		}
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, " ")
}

// capitalize returns s with the first rune upper-cased; mirrors C's
// `capitalize()` helper used at act_obj.c:3848.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	if r[0] >= 'a' && r[0] <= 'z' {
		r[0] -= 32
	}
	return string(r)
}

// aoran returns "a" or "an" based on the first letter of s; mirrors C's
// `aoran()` helper used at act_obj.c:3830.
func aoran(s string) string {
	if s == "" {
		return "a"
	}
	switch s[0] {
	case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
		return "an"
	}
	return "a"
}
