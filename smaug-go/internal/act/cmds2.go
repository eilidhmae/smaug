package act

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// DoSplit divides coins evenly among ch and grouped followers in the same
// room, giving the remainder to ch. Mirrors src/act_comm.c:do_split
// (ENABLE_GOLD_SILVER_COPPER branch) since the Go port tracks all three
// coin types. Coin type defaults to gold when omitted.
//
// Usage: split <amount> [gold|silver|copper]
func DoSplit(ch *types.CharData, argument string) {
	arg, rest := util.OneArgument(argument)
	coinArg, _ := util.OneArgument(rest)
	if arg == "" {
		ch.Send("Split <amount> [gold|silver|copper]\n\r")
		return
	}

	if !util.IsNumber(arg) {
		ch.Send("Split how much?\n\r")
		return
	}

	// Pick coin bucket — default gold. Accept full name + common short forms.
	const (
		coinGold = iota
		coinSilver
		coinCopper
	)
	coinType := coinGold
	switch strings.ToLower(coinArg) {
	case "", "gold":
		coinType = coinGold
	case "silver", "silv":
		coinType = coinSilver
	case "copper", "cop":
		coinType = coinCopper
	default:
		ch.Sendf("%s is not a valid coin type.\n\r", coinArg)
		return
	}

	amount := 0
	fmt.Sscanf(arg, "%d", &amount)

	if amount < 0 {
		ch.Send("Your group wouldn't like that.\n\r")
		return
	}
	if amount == 0 {
		ch.Send("You hand out zero coins, but no one notices.\n\r")
		return
	}

	// Per-bucket balance check + purse selector.
	var purse *int
	var coinName string
	switch coinType {
	case coinGold:
		purse = &ch.Gold
		coinName = "gold"
	case coinSilver:
		purse = &ch.Silver
		coinName = "silver"
	case coinCopper:
		purse = &ch.Copper
		coinName = "copper"
	}
	if *purse < amount {
		ch.Sendf("You don't have that much %s.\n\r", coinName)
		return
	}

	// Count group members in the same room (includes ch).
	groupLeader := ch.Leader
	if groupLeader == nil {
		groupLeader = ch
	}

	members := 0
	var mates []*types.CharData
	if ch.InRoom != nil {
		for _, gch := range ch.InRoom.People {
			if gch == ch {
				members++
				continue
			}
			gLeader := gch.Leader
			if gLeader == nil {
				gLeader = gch
			}
			if gLeader == groupLeader {
				members++
				mates = append(mates, gch)
			}
		}
	} else {
		members = 1
	}

	if members < 2 {
		ch.Send("Just keep it all.\n\r")
		return
	}

	share := amount / members
	extra := amount % members
	if share == 0 {
		ch.Send("Don't even bother, cheapskate.\n\r")
		return
	}

	*purse -= amount
	*purse += share + extra

	ch.Sendf("You split %d %s coins.  Your share is %d %s coins.\n\r",
		amount, coinName, share+extra, coinName)

	msg := fmt.Sprintf("$n splits %d %s coins.  Your share is %d %s coins.",
		amount, coinName, share, coinName)
	for _, gch := range mates {
		util.Act(msg, ch, gch, nil, nil, types.TO_VICT)
		switch coinType {
		case coinGold:
			gch.Gold += share
		case coinSilver:
			gch.Silver += share
		case coinCopper:
			gch.Copper += share
		}
	}
}

// DoLight activates an ITEM_LIGHT (torch/lantern/etc.) carried by ch.
// Mirrors src/misc.c:1917 — C checks obj->value[1] > 0 for ITEM_LIGHT's
// "has fuel" guard, and sets PIPE_LIT on value[3] when lit. We skip C's
// tinder requirement as a documented MVP simplification.
func DoLight(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Light what?\n\r")
		return
	}

	obj := handler.GetObjCarry(ch, arg)
	if obj == nil {
		ch.Send("You aren't carrying that.\n\r")
		return
	}

	if obj.ItemType != types.ITEM_LIGHT {
		ch.Send("You can't light that.\n\r")
		return
	}

	// C misc.c:1917: `if (obj->value[1] > 0)` — Value[1] is the fuel/hours
	// gate. A negative value is used elsewhere in SMAUG to mean "permanent"
	// so we allow <0 through as well.
	if obj.Value[1] == 0 {
		ch.Send("You can't light that.\n\r")
		return
	}

	// Check/set the lit bit on Value[3].
	if obj.Value[3]&int(types.PIPE_LIT) != 0 {
		ch.Send("It's already lit.\n\r")
		return
	}
	obj.Value[3] |= int(types.PIPE_LIT)

	util.Act("You carefully light $p.", ch, nil, obj, nil, types.TO_CHAR)
	util.Act("$n carefully lights $p.", ch, nil, obj, nil, types.TO_ROOM)

	// If worn as light, increment room light counter.
	if obj.WearLoc == types.WEAR_LIGHT && ch.InRoom != nil {
		ch.InRoom.Light++
	}
}

// DoThrow hurls an object from inventory. MVP: the object leaves ch's
// inventory and either lands in a random adjacent room (if any exists) or
// clatters to the floor of the current room. The C original (skills.c:6355)
// is commented-out pseudo-code with bugs; we implement the simpler, safe
// behaviour described in the tier-4 plan.
func DoThrow(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Throw what?\n\r")
		return
	}

	obj := handler.GetObjCarry(ch, arg)
	if obj == nil {
		ch.Send("You aren't carrying that.\n\r")
		return
	}

	if ch.InRoom == nil {
		ch.Send("You have nowhere to throw it.\n\r")
		return
	}

	// Collect live exits.
	var exits []*types.ExitData
	for _, ex := range ch.InRoom.Exits {
		if ex != nil && ex.ToRoom != nil {
			exits = append(exits, ex)
		}
	}

	if len(exits) == 0 {
		// No exits — object just clatters to the floor.
		util.Act("You throw $p, and it clatters to the floor.",
			ch, nil, obj, nil, types.TO_CHAR)
		util.Act("$n throws $p, which clatters to the floor.",
			ch, nil, obj, nil, types.TO_ROOM)
		handler.ObjFromChar(obj)
		handler.ObjToRoom(obj, ch.InRoom)
		return
	}

	// Pick a random exit and send the object there.
	pick := exits[rand.Intn(len(exits))]
	dest := pick.ToRoom

	util.Act("You throw $p through the exit.", ch, nil, obj, nil, types.TO_CHAR)
	util.Act("$n throws $p through an exit.", ch, nil, obj, nil, types.TO_ROOM)
	handler.ObjFromChar(obj)
	handler.ObjToRoom(obj, dest)

	// Announce arrival to occupants of the destination room.
	for _, p := range dest.People {
		if p.Desc != nil {
			p.Sendf("%s clatters in from an exit.\n\r", obj.ShortDescr)
		}
	}
}

// MaxAliases caps per-player alias count. Runtime-only guard against a script
// looping DoAlias forever (unbounded memory growth + O(n) tax on every command
// dispatch via FindAlias). Not enforced on load — legacy player files with
// more entries still parse.
const MaxAliases = 50

// DoAlias lists or creates player aliases. Maps to src/alias.c:do_alias.
// Usage:
//
//	alias                  — list
//	alias <name>           — show/delete (delete when no expansion follows)
//	alias <name> <expand>  — create or modify
func DoAlias(ch *types.CharData, argument string) {
	if ch.IsNPC() || ch.PCData == nil {
		return
	}

	if strings.ContainsRune(argument, '~') {
		ch.Send("Command not acceptable, cannot use the ~ character.\n\r")
		return
	}

	arg, rest := util.OneArgument(argument)

	if arg == "" {
		if len(ch.PCData.Aliases) == 0 {
			ch.Send("You have no aliases defined.\n\r")
			return
		}
		ch.Sendf("%-20s %s\n\r", "Alias", "What it does")
		for _, a := range ch.PCData.Aliases {
			ch.Sendf("%-20s %s\n\r", a.Name, a.Cmd)
		}
		return
	}

	// Lookup existing.
	var existing *types.AliasData
	for _, a := range ch.PCData.Aliases {
		if strings.EqualFold(a.Name, arg) {
			existing = a
			break
		}
	}

	if rest == "" {
		// No expansion → delete.
		if existing == nil {
			ch.Send("That alias does not exist.\n\r")
			return
		}
		removeAlias(ch, existing)
		ch.Send("Deleted alias.\n\r")
		return
	}

	// Create or modify.
	if existing == nil {
		if len(ch.PCData.Aliases) >= MaxAliases {
			ch.Send("You have too many aliases. Remove some with 'unalias <name>' first.\n\r")
			return
		}
		ch.PCData.Aliases = append(ch.PCData.Aliases,
			&types.AliasData{Name: arg, Cmd: rest})
		ch.Send("Created alias.\n\r")
		return
	}
	existing.Cmd = rest
	ch.Send("Modified alias.\n\r")
}

// DoUnalias removes a player alias by name.
func DoUnalias(ch *types.CharData, argument string) {
	if ch.IsNPC() || ch.PCData == nil {
		return
	}
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Unalias what?\n\r")
		return
	}
	for _, a := range ch.PCData.Aliases {
		if strings.EqualFold(a.Name, arg) {
			removeAlias(ch, a)
			ch.Send("Deleted alias.\n\r")
			return
		}
	}
	ch.Send("That alias does not exist.\n\r")
}

func removeAlias(ch *types.CharData, target *types.AliasData) {
	out := ch.PCData.Aliases[:0]
	for _, a := range ch.PCData.Aliases {
		if a != target {
			out = append(out, a)
		}
	}
	ch.PCData.Aliases = out
}

// FindAlias returns the alias whose name has `command` as a case-insensitive
// prefix, or nil. Matches C alias.c:42 semantics (`!str_prefix(argument,
// pal->name)`): typing "g" fires an alias named "get".
func FindAlias(ch *types.CharData, command string) *types.AliasData {
	if ch == nil || ch.PCData == nil || command == "" {
		return nil
	}
	cmd := strings.ToLower(command)
	for _, a := range ch.PCData.Aliases {
		if strings.HasPrefix(strings.ToLower(a.Name), cmd) {
			return a
		}
	}
	return nil
}

// DoAreas lists all loaded areas. Pager-aware. Mirrors src/act_info.c:do_areas.
// MVP: ignores soft-range range filters and "old" keyword — just tabulates.
func DoAreas(ch *types.CharData, argument string) {
	if WorldRef == nil {
		ch.Send("No areas loaded.\n\r")
		return
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "\n\r&c&GAuthor          &Y|&G              Area                     &Y| &GRecommended &Y|\n\r")
	fmt.Fprintf(&sb, "-------------------+---------------------------------------+-------------+\n\r")
	for _, a := range WorldRef.Areas {
		if a == nil {
			continue
		}
		author := a.Author
		if author == "" {
			author = "(unknown)"
		}
		name := a.Name
		if name == "" {
			name = a.Filename
		}
		fmt.Fprintf(&sb, "&c&G%-18.18s &Y| &z%-37s &Y|&W %4d - %-4d &Y|\n\r",
			author, name, a.LowSoftRange, a.HiSoftRange)
	}
	ch.Send(sb.String())
}

// DoAltscore prints a compact single-screen score variant.
// Maps to src/player.c:do_altscore.
func DoAltscore(ch *types.CharData, argument string) {
	if ch.IsNPC() {
		DoScore(ch, argument)
		return
	}

	title := ""
	if ch.PCData != nil {
		title = ch.PCData.Title
	}

	raceName := "Unknown"
	className := "Unknown"
	if WorldRef != nil {
		if ch.Race >= 0 && ch.Race < len(WorldRef.Races) && WorldRef.Races[ch.Race] != nil {
			raceName = WorldRef.Races[ch.Race].Name
		}
		if ch.Class >= 0 && ch.Class < len(WorldRef.Classes) && WorldRef.Classes[ch.Class] != nil {
			className = WorldRef.Classes[ch.Class].WhoName
		}
	}

	ch.Sendf("\n\r&G%s %s.&D\n\r", ch.Name, title)
	ch.Send("&g----------------------------------------------------------------------------&D\n\r")
	ch.Sendf("&gLevel: &W%-3d &gRace: &W%-10.10s &gClass: &W%-10.10s&D\n\r",
		ch.Level, raceName, className)
	ch.Sendf("&gHP: &W%d/%d  &gMana: &W%d/%d  &gMove: &W%d/%d&D\n\r",
		ch.Hit, ch.MaxHit, ch.Mana, ch.MaxMana, ch.Move, ch.MaxMove)
	ch.Sendf("&gStr: &W%d &gInt: &W%d &gWis: &W%d &gDex: &W%d &gCon: &W%d &gCha: &W%d &gLck: &W%d&D\n\r",
		ch.GetCurrStr(), ch.GetCurrInt(), ch.GetCurrWis(),
		ch.GetCurrDex(), ch.GetCurrCon(), ch.GetCurrCha(), ch.GetCurrLck())
	ch.Sendf("&gHitroll: &W%-3d &gDamroll: &W%-3d &gArmor: &W%d  &gAlign: &W%d&D\n\r",
		ch.Hitroll, ch.Damroll, ch.Armor, ch.Alignment)
	ch.Sendf("&gGold: &W%d  &gSilver: &W%d  &gCopper: &W%d  &gExp: &W%d&D\n\r",
		ch.Gold, ch.Silver, ch.Copper, ch.Exp)
	if ch.PCData != nil {
		ch.Sendf("&gPkills: &W%d &gPdeaths: &W%d &gMkills: &W%d &gMdeaths: &W%d&D\n\r",
			ch.PCData.PKills, ch.PCData.PDeaths, ch.PCData.MKills, ch.PCData.MDeaths)
	}
	if len(ch.Affects) > 0 {
		ch.Sendf("&gAffects: &W%d active&D\n\r", len(ch.Affects))
	}
}

// DoColor toggles ANSI colour output. MVP: simple PLR_ANSI toggle; matches
// the Descriptor's ColorFunc gating (see types/descriptor.go:FlushOutput).
// Usage: color           — toggle
//
//	color on        — enable
//	color off       — disable
func DoColor(ch *types.CharData, argument string) {
	if ch.IsNPC() {
		ch.Send("Only PCs can change colors.\n\r")
		return
	}
	arg, _ := util.OneArgument(argument)
	switch arg {
	case "on", "yes", "1":
		ch.Act.Set(types.PLR_ANSI)
		ch.Send("ANSI color is now ON.\n\r")
	case "off", "no", "0":
		ch.Act.Remove(types.PLR_ANSI)
		ch.Send("ANSI color is now OFF.\n\r")
	default:
		if ch.Act.IsSet(types.PLR_ANSI) {
			ch.Act.Remove(types.PLR_ANSI)
			ch.Send("ANSI color is now OFF.\n\r")
		} else {
			ch.Act.Set(types.PLR_ANSI)
			ch.Send("ANSI color is now ON.\n\r")
		}
	}
}

// DoCompress toggles MCCP2 compression state on the descriptor. MVP: flips
// the flag and announces. The full telnet negotiation with the client lives
// in net/protocol.go; this command is the user-facing toggle, matching
// src/mccp.c:do_compress.
func DoCompress(ch *types.CharData, argument string) {
	if ch.Desc == nil || ch.Desc.TelnetState == nil {
		ch.Send("No connection state.\n\r")
		return
	}
	if ch.Desc.TelnetState.MCCPEnabled {
		ch.Desc.TelnetState.MCCPEnabled = false
		ch.Send("MCCP compression DISABLED.\n\r")
	} else {
		ch.Desc.TelnetState.MCCPEnabled = true
		ch.Send("MCCP compression ENABLED.\n\r")
	}
}
