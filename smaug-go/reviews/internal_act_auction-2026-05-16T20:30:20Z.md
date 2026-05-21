# Adversary Review

**Target**: `internal/act/auction.go`
**Timestamp**: 2026-05-16T20:30:20Z
**Model**: local-mlx-dscoder/mlx-community/deepseek-coder-33b-instruct-hf-4bit-mlx

This file seems to be a part of a game server written in Go. It contains the source code for the auction system of the game. The code is divided into several sections, each handling a specific aspect of the auction process.

```go
// Full `do_auction` state machine, non-GSC variant, ports
// `src/act_obj.c:3775-4240` (do_auction) with per-branch helpers.
// Auction tick lives in `internal/game/update_auction.go` (the per-pulse, pulseTick, pulseMobile, pulseSave peers).

// Shared reader: BroadcastAuction is called from both the command and the tick; it's the `talk_auction` helper (`src/act_comm.c:4309`) and retains the Tier-9 filter chain verbatim.
package act

import (
	"fmt"
	"strings"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

const auctionMinTrust = 5 // auctionMinTrust
const auctionLevelFloor = 3 // auctionLevelFloor
const auctionHourOpen = 9 // auctionHourOpen
const auctionHourClose = 18 // auctionHourClose
const auctionMaxBid = 2000000000 // auctionMaxBid
const auctionMinIncrement = 10000 // auctionMinIncrement

// `do_auction` (non-GSC) at `src/act_obj.c:3775-4240`.
// Four intents:
// 1. `auction` with no args shows current item info or "Nothing auctioned."
// 2. `auction stop` (immortal only) cancels + refunds
// 3. `auction bid <amount> [item-keyword]` places a bid
// 4. `auction <item> [min-bet]` starts a new auction for a carried item
func DoAuction(ch *types.CharData, argument string) {
	if ch == nil {
		return
	}
	arg1, rest1 := util.OneArgument(argument)
	arg2, rest2 := util.OneArgument(rest1)
	arg3, _ := util.OneArgument(rest2)

	// NPC silent return at `act_obj.c:3791-3792`.
	if ch.IsNPC() {
		return
	}

	// Level gate at `act_obj.c:3794-3798`.
	if ch.Level < auctionLevelFloor {
		ch.Send("You must be at least level three to the auction...\n\r")
		return
	}

	// Time gate at `act_obj.c:3801-3808`.
	if WorldRef != nil && WorldRef.Auction != nil {
		hour := WorldRef.TimeInfo.Hour
		if (hour > auctionHourClose || hour < auctionHourOpen) && WorldRef.Auction.Item == nil && !ch.IsImmortal() {
			ch.Send("\n\rThe auctioneer works between the hours of 9AM and 6PM\n\r")
			return
		}
	}

	// Empty arg1 shows info block at `act_obj.c:3810-3955`.
	if arg1 == "" {
		auctionInfo(ch)
		return
	}

	// Immortal `stop` at `act_obj.c:3957-3982`.
	if ch.IsImmortal() && strings.EqualFold(arg1, "stop") {
		auctionStop(ch)
		return
	}

	// Bid at `act_obj.c:3983-4077`.
	if strings.EqualFold(arg1, "bid") {
		auctionBid(ch, arg2, arg3)
		return
	}

	// Start new auction at `act_obj.c:4078-4240`.
	auctionStart(ch, arg1, arg2)
}

// auctionInfo at `act_obj.c:3810-3955`.
// Show either current auction details or "Nothing auctioned.".
func auctionInfo(ch *types.CharData) {
	auc := WorldRef.Auction
	if auc == nil || auc.Item == nil {
		ch.Send("\n\rThere is nothing being auctioned right now. What would you like to auction?\n\r")
		return
	}

	obj := auc.Item
	ch.Sendf("\n\rCurrent bid on this item is %s gold.\n\r", util.NumPunct(auc.Bet))
	if auc.Bet == 0 {
		ch.Send("No bids on this item have been received.\n\r")
	}

	// Core item description at `act_obj.c:3827-3835`.
	ch.Sendf("Object '%s' is %s, special properties: %s\n\rIts weight is %d, value is %d, and level is %d.\n\r", obj.Name, aoran(itemTypeName(obj.ItemType)), auctionExtraBitName(obj.ExtraFlags), obj.Weight, obj.GoldCost, obj.Level)

	// Wear-location line at `act_obj.c:3836-3838`.
	if obj.ItemType != types.ITEM_LIGHT && obj.WearFlags > 1 {
		ch.Sendf("Item's wear location: %s\n\r", auctionWearLocString(obj.WearFlags))
	}

	// Per-item-type info block at `act_obj.c:3842-3920`.
	auctionItemTypeInfo(ch, obj)

	// Contents for containers at `act_obj.c:3927-3934`.
	if (obj.ItemType == types.ITEM_CONTAINER || obj.ItemType == types.ITEM_KEYRING || obj.ItemType == types.ITEM_QUIVER) && len(obj.Contents) > 0 {
		ch.Send("Contents:\n\r")
		for _, c := range obj.Contents {
			ch.Sendf("%s\n\r", c.ShortDescr)
		}
	}

	// Immortal seller/buyer disclosure at `act_obj.c:3936-3944`.
	if ch.IsImmortal() {
		sellerName := "nobody"
		buyerName := "nobody"
		if auc.Seller != nil {
			sellerName = auc.Seller.Name
		}
		if auc.Buyer != nil {
			buyerName = auc.Buyer.Name
		}
		ch.Sendf("Seller: %s. Bidder: %s. Round: %d.\n\r", sellerName, buyerName, auc.Going+1)
		ch.Sendf("Time left in round: %d.\n\r", auc.Pulse)
	}
}

// auctionStop at `act_obj.c:3957-3982`.
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

	// Refund any distinct bidder at `act_obj.c:3974-3978`.
	if auc.Buyer != nil && auc.Buyer != auc.Seller {
		auc.Buyer.Gold += auc.Bet
		auc.Buyer.Send("Your money has been returned.\n\r")
	}

	auc.Item = nil
	auc.Seller = nil
	auc.Buyer = nil
	auc.Bet = 0
	auc.Going = 0
	auc.Pulse = 0
	auc.Starting = 0
}

// auctionBid at `act_obj.c:3983-4077`.
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

	newBet := util.ParseBet(auc.Bet, arg2)
	if newBet < auc.Starting {
		ch.Send("You must place a bid that is higher than the starting bet.\n\r")
		return
	}

	// +10000 step at `act_obj.c:4024`.
	if newBet < auc.Bet || newBet < (auc.Bet+auctionMinIncrement) {
		ch.Send("You must at least bid 10000 coins over the current bid.\n\r")
		return
	}

	if newBet > ch.Gold {
		ch.Send("You don't have that much money!\n\r")
		return
	}

	if newBet > auctionMaxBid {
		ch.Send("You can't bid over 2 billion coins.\n\r")
		return
	}

	// Item-keyword match at `act_obj.c:4045-4050`.
	if arg3 != "" && !util.IsName(arg3, auc.Item.Name) {
		ch.Send("That item is not being auctioned right now.\n\r")
		return
	}

	// Refund previous distinct bidder at `act_obj.c:4054-4055`.
	if auc.Buyer != nil && auc.Buyer != auc.Seller {
		auc.Buyer.Gold += auc.Bet
	}

	ch.Gold -= newBet
	auc.Buyer = ch
	auc.Bet = newBet
	auc.Going = 0
	auc.Pulse = 0
	msg := fmt.Sprintf("A bid of %s gold has been received on %s.", util.NumPunct(newBet), auc.Item.ShortDescr)
	BroadcastAuction(msg)
}

// auctionStart at `act_obj.c:4078-4240`.
func auctionStart(ch *types.CharData, arg1, arg2 string) {
	obj := handler.GetObjCarry(ch, arg1)
	if obj == nil {
		ch.Send("You aren't carrying that.\n\r")
		return
	}

	// NoAuction blacklist at `act_obj.c:4090-4096`.
	if !ch.IsImmortal() {
		for _, vnum := range WorldRef.NoAuction {
			if obj.IndexData != nil && obj.IndexData.Vnum == vnum {
				ch.Send("That item cannot be auctioned.\n\r")
				return
			}
		}
	}

	// Type mismatch at `act_obj.c:4099-4104`.
	if obj.IndexData != nil && obj.ItemType != obj.IndexData.ItemType {
		ch.Send("This object is too modified to auction. Please contact an Immortal.\n\r")
		return
	}

	// Decaying items at `act_obj.c:4107-4111`.
	if obj.Timer > 0 {
		ch.Send("You can't auction objects that are decaying.\n\r")
		return
	}

	// Clan/permanent at `act_obj.c:4113-4123`.
	if obj.ExtraFlags.IsSet(types.ITEM_CLANOBJECT) {
		ch.Send("You can't auction clan items.\n\r")
		return
	}
	if obj.ExtraFlags.IsSet(types.ITEM_PERMANENT) {
		ch.Send("This item cannot leave your possession.\n\r")
		return
	}

	// History collision at `act_obj.c:4125-4134`.
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

	// Min-bet parsing at `act_obj.c:4137-4155`.
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
	if minBet != -1 {
		minBet = atoiSafe(arg2)
	}
	if minBet < 0 {
		ch.Send("You can't auction something for less than 0 gold!\n\r")
		return
	}

	// Second auction guard at `act_obj.c:4232-4239`.
	if auc.Item != nil {
		util.Act(types.AT_TELL, "Try again later-$p is being auctioned right now!!", ch, nil, auc.Item, nil, types.TO_CHAR)
		if !ch.IsImmortal() {
			ch.Wait = types.PULSE_VIOLENCE
		}
		return
	}

	// Item-type whitelist switch at `act_obj.c:4158-4231`.
	if !auctionTypeAllowed(obj.ItemType) {
		util.Act(types.AT_TELL, "You cannot auction $Ts.", ch, nil, nil, itemTypeName(obj.ItemType), types.TO_CHAR)
		return
	}

	handler.ObjFromChar(obj)
	auc.Item = obj
	auc.Bet = 0
	auc.Buyer = ch
	auc.Seller = ch
	auc.Pulse = types.PULSE_AUCTION
	auc.Going = 0
	auc.Starting = minBet

	// History ring rotation at `act_obj.c:4210-4216`.
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
	msg := fmt.Sprintf("A new item is being auctioned: %s at %d gold.", obj.ShortDescr, auc.Starting)
	BroadcastAuction(msg)
}

// atoiSafe is strconv.Atoi with a zero-fallback.
func atoiSafe(s string) int {
	n := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			n = n*10 + int(r-'0')
		}
	}
	return n
}

// auctionTypeAllowed mirrors the C case-list at `act_obj.c:4169-4196`.
func auctionTypeAllowed(t int) bool {
	switch t {
	case types.ITEM_PAPER, types.ITEM_LIGHT, types.ITEM_TREASURE, types.ITEM_POTION, types.ITEM_KEYRING, types.ITEM_QUIVER, types.ITEM_DRINK_CON, types.ITEM_FOOD, types.ITEM_COOK, types.ITEM_PEN, types.ITEM_BOAT, types.ITEM_PILL, types.ITEM_PIPE, types.ITEM_HERB_CON, types.ITEM_INCENSE, types.ITEM_FIRE, types.ITEM_RUNEPOUCH, types.ITEM_MAP, types.ITEM_BOOK, types.ITEM_RUNE, types.ITEM_MATCH, types.ITEM_HERB, types.ITEM_WEAPON, types.ITEM_MISSILE_WEAPON, types.ITEM_ARMOR, types.ITEM_STAFF, types.ITEM_WAND, types.ITEM_SCROLL:
		return true
	}
	return false
}

// auctionItemTypeInfo at `act_obj.c:3842-3920` for the G4b info block.
func auctionItemTypeInfo(ch *types.CharData, obj *types.ObjData) {
	switch obj.ItemType {
	case types.ITEM_CONTAINER, types.ITEM_KEYRING, types.ITEM_QUIVER:
		v := obj.Value[0]
		var cap string
		switch {
		case v < 76:
			cap = "has a small capacity"
		case v < 150:
			cap = "has a small to medium capacity"
		case v < 300:
			cap = "has a medium capacity"
		case v < 500:
			cap = "has a medium to large capacity"
		case v < 751:
			cap = "has a large capacity"
		default:
			cap = "has a giant capacity"
		}
		ch.Sendf("%s appears to %s.\n\r", capitalize(obj.ShortDescr), cap)
	case types.ITEM_PILL, types.ITEM_SCROLL, types.ITEM_POTION:
		var b strings.Builder
		fmt.Fprintf(&b, "Level %d spells of:", obj.Value[0])
		for i := 1; i <= 3; i++ {
			name := auctionSpellName(obj.Value[i])
			if name != "" {
				fmt.Fprintf(&b, "'%s'", name)
			}
		}
		b.WriteString(".\n\r")
		ch.Send(b.String())
	case types.ITEM_WAND, types.ITEM_STAFF:
		var b strings.Builder
		fmt.Fprintf(&b, "Has %d(%d) charges of level %d", obj.Value[1], obj.Value[2], obj.Value[0])
		if name := auctionSpellName(obj.Value[3]); name != "" {
			fmt.Fprintf(&b, "'%s'", name)
		}
		b.WriteString(".\n\r")
		ch.Send(b.String())
	case types.ITEM_MISSILE_WEAPON, types.ITEM_WEAPON:
		avg := (obj.Value[1] + obj.Value[2]) / 2
		poison := ""
		if obj.ExtraFlags.IsSet(types.ITEM_POISONED) {
			poison = "\n\rThis weapon is poisoned."
		}
		ch.Sendf("Damage is %d to %d(average %d).%s\n\r", obj.Value[1], obj.Value[2], avg, poison)
	case types.ITEM_ARMOR:
		ch.Sendf("Armor class is %d.\n\r", obj.Value[0])
	}
}

// auctionSpellName resolves a spell-index `sn` to a printable name, or an empty string if out of range/nil.
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

// itemTypeName returns the human-readable item-type label used in the C `item_type_name()`.
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
		return "missileweapon"
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
		return "drinkcontainer"
	case types.ITEM_KEY:
		return "key"
	case types.ITEM_FOOD:
		return "food"
	case types.ITEM_MONEY:
		return "money"
	case types.ITEM_BOAT:
		return "boat"
	case types.ITEM_CORPSE_NPC:
		return "NPCcorpse"
	case types.ITEM_CORPSE_PC:
		return "PCcorpse"
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
		return "herbcontainer"
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
		return "runepouch"
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
		return "cookedfood"
	case types.ITEM_PEN:
		return "pen"
	}
	return "item"
}

// auctionExtraBitName returns the special-properties phrase used in the auction info block.
func auctionExtraBitName(flags types.BitVector) string {
	if flags == (types.BitVector{}) {
		return "none"
	}
	var parts []string
	named := []struct {
		bit int
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
	return strings.Join(parts, "")
}

// auctionWearLocString summarizes the `w_flags` output at `act_obj.c:3838`.
func auctionWearLocString(wf int) string {
	var parts []string
	bits := uint32(wf)
	locs := []struct {
		bit uint32
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
	return strings.Join(parts, "")
}

// capitalize returns s with the first rune upper-cased; mirrors C's `capitalize()` helper.
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

// aoran returns "a" or "an" based on the first letter of s; mirrors C's `aoran()` helper.
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
```
