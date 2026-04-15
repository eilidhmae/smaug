package act

import (
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// DoOedit implements the 'oedit' command: edit an object prototype non-interactively.
// Usage: oedit <vnum> <subcommand> [args...]
// Note: the C OLC enters an interactive CON_OEDIT substate. This MVP form
// dispatches each subcommand immediately, mirroring the pattern used by
// DoRedit. A true interactive substate is a documented follow-up.
func DoOedit(ch *types.CharData, argument string) {
	if ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}
	vnumArg, rest := util.OneArgument(argument)
	if vnumArg == "" {
		ch.Send("Usage: oedit <vnum> <subcommand> [args]\n\r")
		ch.Send("Subcommands: create, name, short, long, type, flags, wearflags, values, weight, cost, level, affects, ed, delete\n\r")
		return
	}
	vnum, err := strconv.Atoi(vnumArg)
	if err != nil || vnum <= 0 {
		ch.Send("Vnum must be a positive number.\n\r")
		return
	}
	sub, args := util.OneArgument(rest)
	sub = strings.ToLower(sub)

	// `create` is the one subcommand that works on a non-existing vnum.
	if sub == "create" {
		if _, exists := WorldRef.ObjIndex[vnum]; exists {
			ch.Sendf("Object %d already exists.\n\r", vnum)
			return
		}
		name := strings.TrimSpace(args)
		if name == "" {
			name = "new object"
		}
		idx := &types.ObjIndexData{
			Vnum:        vnum,
			Name:        name,
			ShortDescr:  name,
			Description: util.Capitalize(name) + " is here.",
			ItemType:    types.ITEM_TRASH,
			Level:       1,
			Weight:      1,
		}
		WorldRef.ObjIndex[vnum] = idx
		ch.Sendf("Object %d (%s) created.\n\r", vnum, name)
		return
	}
	if sub == "delete" {
		// Delegate to DoOdelete for confirmation semantics.
		DoOdelete(ch, vnumArg)
		return
	}

	idx, ok := WorldRef.ObjIndex[vnum]
	if !ok {
		ch.Sendf("Object %d does not exist. Use 'oedit %d create' to make one.\n\r", vnum, vnum)
		return
	}

	switch sub {
	case "":
		oeditShow(ch, idx)
	case "name":
		if args == "" {
			ch.Sendf("Current name: %s\n\r", idx.Name)
			return
		}
		idx.Name = args
		ch.Sendf("Name set to: %s\n\r", args)
	case "short":
		if args == "" {
			ch.Sendf("Current short: %s\n\r", idx.ShortDescr)
			return
		}
		idx.ShortDescr = args
		ch.Sendf("Short descr set to: %s\n\r", args)
	case "long":
		if args == "" {
			ch.Sendf("Current long: %s\n\r", idx.Description)
			return
		}
		idx.Description = args
		ch.Sendf("Long descr set to: %s\n\r", args)
	case "type":
		t, ok := itemTypeFromName(args)
		if !ok {
			ch.Send("Unknown item type.\n\r")
			return
		}
		idx.ItemType = t
		ch.Sendf("Type set to %s.\n\r", args)
	case "flags":
		val, err := strconv.Atoi(strings.TrimSpace(args))
		if err != nil {
			ch.Send("flags <bit-number>\n\r")
			return
		}
		idx.ExtraFlags.Toggle(val)
		ch.Sendf("Extra flag %d toggled.\n\r", val)
	case "wearflags":
		val, err := strconv.Atoi(strings.TrimSpace(args))
		if err != nil {
			ch.Send("wearflags <bit-number>\n\r")
			return
		}
		idx.WearFlags ^= 1 << val
		ch.Sendf("Wear flag %d toggled.\n\r", val)
	case "values", "value":
		slotArg, valArg := util.OneArgument(args)
		slot, err := strconv.Atoi(slotArg)
		if err != nil || slot < 0 || slot >= 6 {
			ch.Send("values <0-5> <value>\n\r")
			return
		}
		v, err := strconv.Atoi(strings.TrimSpace(valArg))
		if err != nil {
			ch.Send("values <0-5> <value>\n\r")
			return
		}
		idx.Value[slot] = v
		ch.Sendf("Value[%d] set to %d.\n\r", slot, v)
	case "weight":
		v, err := strconv.Atoi(strings.TrimSpace(args))
		if err != nil {
			ch.Send("weight <n>\n\r")
			return
		}
		idx.Weight = v
		ch.Sendf("Weight set to %d.\n\r", v)
	case "cost":
		v, err := strconv.Atoi(strings.TrimSpace(args))
		if err != nil {
			ch.Send("cost <n>\n\r")
			return
		}
		idx.GoldCost = v
		ch.Sendf("Cost set to %d gold.\n\r", v)
	case "level":
		v, err := strconv.Atoi(strings.TrimSpace(args))
		if err != nil {
			ch.Send("level <n>\n\r")
			return
		}
		idx.Level = v
		ch.Sendf("Level set to %d.\n\r", v)
	case "affects":
		oeditAffects(ch, idx, args)
	case "ed":
		kw := strings.TrimSpace(args)
		if kw == "" {
			ch.Send("ed <keyword>\n\r")
			return
		}
		idx.ExtraDescr = append(idx.ExtraDescr, &types.ExtraDescrData{Keyword: kw})
		ch.Sendf("Extra descr '%s' added (use oset to populate text).\n\r", kw)
	default:
		ch.Send("Unknown oedit subcommand.\n\r")
	}
}

func oeditShow(ch *types.CharData, idx *types.ObjIndexData) {
	ch.Sendf("Object %d: %s\n\r", idx.Vnum, idx.ShortDescr)
	ch.Sendf("  Name   : %s\n\r", idx.Name)
	ch.Sendf("  Long   : %s\n\r", idx.Description)
	ch.Sendf("  Type   : %d  Level: %d  Weight: %d  Cost: %d\n\r", idx.ItemType, idx.Level, idx.Weight, idx.GoldCost)
	ch.Sendf("  Values : %d %d %d %d %d %d\n\r", idx.Value[0], idx.Value[1], idx.Value[2], idx.Value[3], idx.Value[4], idx.Value[5])
}

func oeditAffects(ch *types.CharData, idx *types.ObjIndexData, args string) {
	sub, rest := util.OneArgument(args)
	switch strings.ToLower(sub) {
	case "add":
		locArg, modArg := util.OneArgument(rest)
		loc, err := strconv.Atoi(strings.TrimSpace(locArg))
		if err != nil {
			ch.Send("affects add <location> <modifier>\n\r")
			return
		}
		mod, err := strconv.Atoi(strings.TrimSpace(modArg))
		if err != nil {
			ch.Send("affects add <location> <modifier>\n\r")
			return
		}
		idx.Affects = append(idx.Affects, &types.AffectData{Location: loc, Modifier: mod, Duration: -1})
		ch.Sendf("Affect added (loc=%d, mod=%d).\n\r", loc, mod)
	case "del":
		n, err := strconv.Atoi(strings.TrimSpace(rest))
		if err != nil || n < 0 || n >= len(idx.Affects) {
			ch.Sendf("affects del <0-%d>\n\r", len(idx.Affects)-1)
			return
		}
		idx.Affects = append(idx.Affects[:n], idx.Affects[n+1:]...)
		ch.Send("Affect removed.\n\r")
	default:
		ch.Send("affects add | del\n\r")
	}
}

// itemTypeFromName maps a human name (or number) to an ITEM_* constant.
func itemTypeFromName(s string) (int, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return 0, false
	}
	if v, err := strconv.Atoi(s); err == nil {
		return v, true
	}
	names := map[string]int{
		"light":     types.ITEM_LIGHT,
		"scroll":    types.ITEM_SCROLL,
		"wand":      types.ITEM_WAND,
		"staff":     types.ITEM_STAFF,
		"weapon":    types.ITEM_WEAPON,
		"treasure":  types.ITEM_TREASURE,
		"armor":     types.ITEM_ARMOR,
		"potion":    types.ITEM_POTION,
		"worn":      types.ITEM_WORN,
		"furniture": types.ITEM_FURNITURE,
		"trash":     types.ITEM_TRASH,
		"container": types.ITEM_CONTAINER,
		"drink":     types.ITEM_DRINK_CON,
		"key":       types.ITEM_KEY,
		"food":      types.ITEM_FOOD,
		"money":     types.ITEM_MONEY,
		"boat":      types.ITEM_BOAT,
		"fountain":  types.ITEM_FOUNTAIN,
		"pill":      types.ITEM_PILL,
		"portal":    types.ITEM_PORTAL,
	}
	v, ok := names[s]
	return v, ok
}

// DoMedit implements the 'medit' command: edit a mob prototype non-interactively.
// Usage: medit <vnum> <subcommand> [args...]
func DoMedit(ch *types.CharData, argument string) {
	if ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}
	vnumArg, rest := util.OneArgument(argument)
	if vnumArg == "" {
		ch.Send("Usage: medit <vnum> <subcommand> [args]\n\r")
		ch.Send("Subcommands: create, name, short, long, desc, level, stats, hp, mana, mv, armor, hitroll, damroll, align, flags, race, class, sex, attacks, defenses, gold, xp, position, affected\n\r")
		return
	}
	vnum, err := strconv.Atoi(vnumArg)
	if err != nil || vnum <= 0 {
		ch.Send("Vnum must be a positive number.\n\r")
		return
	}
	sub, args := util.OneArgument(rest)
	sub = strings.ToLower(sub)

	if sub == "create" {
		if _, exists := WorldRef.MobIndex[vnum]; exists {
			ch.Sendf("Mob %d already exists.\n\r", vnum)
			return
		}
		name := strings.TrimSpace(args)
		if name == "" {
			name = "new mob"
		}
		idx := &types.MobIndexData{
			Vnum:        vnum,
			PlayerName:  name,
			ShortDescr:  name,
			LongDescr:   util.Capitalize(name) + " is standing here.\n\r",
			Level:       1,
			Position:    types.POS_STANDING,
			DefPosition: types.POS_STANDING,
		}
		idx.Act.Set(types.ACT_IS_NPC)
		WorldRef.MobIndex[vnum] = idx
		ch.Sendf("Mob %d (%s) created.\n\r", vnum, name)
		return
	}
	if sub == "delete" {
		DoMdelete(ch, vnumArg)
		return
	}

	idx, ok := WorldRef.MobIndex[vnum]
	if !ok {
		ch.Sendf("Mob %d does not exist. Use 'medit %d create' to make one.\n\r", vnum, vnum)
		return
	}

	switch sub {
	case "":
		meditShow(ch, idx)
	case "name":
		if args == "" {
			ch.Sendf("Current name: %s\n\r", idx.PlayerName)
			return
		}
		idx.PlayerName = args
		ch.Sendf("Name set to: %s\n\r", args)
	case "short":
		if args == "" {
			ch.Sendf("Current short: %s\n\r", idx.ShortDescr)
			return
		}
		idx.ShortDescr = args
		ch.Sendf("Short set to: %s\n\r", args)
	case "long":
		if args == "" {
			ch.Sendf("Current long: %s\n\r", idx.LongDescr)
			return
		}
		idx.LongDescr = args + "\n\r"
		ch.Sendf("Long set to: %s\n\r", args)
	case "desc":
		if args == "" {
			ch.Sendf("Current desc: %s\n\r", idx.Description)
			return
		}
		idx.Description = args
		ch.Send("Description set.\n\r")
	case "level":
		v, err := strconv.Atoi(strings.TrimSpace(args))
		if err != nil {
			ch.Send("level <n>\n\r")
			return
		}
		idx.Level = v
		ch.Sendf("Level set to %d.\n\r", v)
	case "stats":
		statName, valArg := util.OneArgument(args)
		v, err := strconv.Atoi(strings.TrimSpace(valArg))
		if err != nil {
			ch.Send("stats <str|int|wis|dex|con|cha|lck> <n>\n\r")
			return
		}
		v = clamp(v, 1, 25)
		switch strings.ToLower(statName) {
		case "str":
			idx.PermStr = v
		case "int":
			idx.PermInt = v
		case "wis":
			idx.PermWis = v
		case "dex":
			idx.PermDex = v
		case "con":
			idx.PermCon = v
		case "cha":
			idx.PermCha = v
		case "lck":
			idx.PermLck = v
		default:
			ch.Send("Unknown stat.\n\r")
			return
		}
		ch.Sendf("%s set to %d.\n\r", statName, v)
	case "hp":
		// Format: "N d M + P"
		f1, r1 := util.OneArgument(args)
		f2, r2 := util.OneArgument(r1)
		f3, _ := util.OneArgument(r2)
		n, _ := strconv.Atoi(f1)
		s, _ := strconv.Atoi(f2)
		p, _ := strconv.Atoi(f3)
		idx.HitNoDice, idx.HitSizeDice, idx.HitPlus = n, s, p
		ch.Sendf("Hp set to %dd%d+%d.\n\r", n, s, p)
	case "mana":
		v, err := strconv.Atoi(strings.TrimSpace(args))
		if err != nil {
			ch.Send("mana <n>\n\r")
			return
		}
		// MobIndexData has no per-mob mana; treat as TODO.
		_ = v
		ch.Send("Note: mob prototypes do not store mana directly; set on instance via mset mana.\n\r")
	case "mv":
		ch.Send("Note: mob prototypes do not store move directly; set on instance via mset move.\n\r")
	case "armor":
		v, err := strconv.Atoi(strings.TrimSpace(args))
		if err != nil {
			ch.Send("armor <n>\n\r")
			return
		}
		idx.AC = v
		ch.Sendf("AC set to %d.\n\r", v)
	case "hitroll":
		v, err := strconv.Atoi(strings.TrimSpace(args))
		if err != nil {
			ch.Send("hitroll <n>\n\r")
			return
		}
		idx.Hitroll = v
		ch.Sendf("Hitroll set to %d.\n\r", v)
	case "damroll":
		v, err := strconv.Atoi(strings.TrimSpace(args))
		if err != nil {
			ch.Send("damroll <n>\n\r")
			return
		}
		idx.Damroll = v
		ch.Sendf("Damroll set to %d.\n\r", v)
	case "align":
		v, err := strconv.Atoi(strings.TrimSpace(args))
		if err != nil {
			ch.Send("align <n>\n\r")
			return
		}
		idx.Alignment = clamp(v, -1000, 1000)
		ch.Sendf("Alignment set to %d.\n\r", idx.Alignment)
	case "flags":
		val, err := strconv.Atoi(strings.TrimSpace(args))
		if err != nil {
			ch.Send("flags <bit-number>\n\r")
			return
		}
		idx.Act.Toggle(val)
		ch.Sendf("Act flag %d toggled.\n\r", val)
	case "race":
		v, err := strconv.Atoi(strings.TrimSpace(args))
		if err != nil {
			ch.Send("race <n>\n\r")
			return
		}
		idx.Race = v
		ch.Sendf("Race set to %d.\n\r", v)
	case "class":
		v, err := strconv.Atoi(strings.TrimSpace(args))
		if err != nil {
			ch.Send("class <n>\n\r")
			return
		}
		idx.Class = v
		ch.Sendf("Class set to %d.\n\r", v)
	case "sex":
		v, ok := sexFromName(args)
		if !ok {
			ch.Send("sex <male|female|neutral|0|1|2>\n\r")
			return
		}
		idx.Sex = v
		ch.Sendf("Sex set to %d.\n\r", v)
	case "attacks":
		val, err := strconv.Atoi(strings.TrimSpace(args))
		if err != nil {
			ch.Send("attacks <bit-number>\n\r")
			return
		}
		idx.Attacks.Toggle(val)
		ch.Sendf("Attacks flag %d toggled.\n\r", val)
	case "defenses":
		val, err := strconv.Atoi(strings.TrimSpace(args))
		if err != nil {
			ch.Send("defenses <bit-number>\n\r")
			return
		}
		idx.Defenses.Toggle(val)
		ch.Sendf("Defenses flag %d toggled.\n\r", val)
	case "gold":
		v, err := strconv.Atoi(strings.TrimSpace(args))
		if err != nil {
			ch.Send("gold <n>\n\r")
			return
		}
		idx.Gold = v
		ch.Sendf("Gold set to %d.\n\r", v)
	case "xp":
		v, err := strconv.Atoi(strings.TrimSpace(args))
		if err != nil {
			ch.Send("xp <n>\n\r")
			return
		}
		idx.Exp = v
		ch.Sendf("Exp set to %d.\n\r", v)
	case "position":
		v, err := strconv.Atoi(strings.TrimSpace(args))
		if err != nil {
			ch.Send("position <n>\n\r")
			return
		}
		idx.Position = v
		idx.DefPosition = v
		ch.Sendf("Position set to %d.\n\r", v)
	case "affected":
		val, err := strconv.Atoi(strings.TrimSpace(args))
		if err != nil {
			ch.Send("affected <bit-number>\n\r")
			return
		}
		idx.AffectedBy.Toggle(val)
		ch.Sendf("AffectedBy flag %d toggled.\n\r", val)
	default:
		ch.Send("Unknown medit subcommand.\n\r")
	}
}

func meditShow(ch *types.CharData, idx *types.MobIndexData) {
	ch.Sendf("Mob %d: %s\n\r", idx.Vnum, idx.ShortDescr)
	ch.Sendf("  Name     : %s\n\r", idx.PlayerName)
	ch.Sendf("  Long     : %s\n\r", strings.TrimRight(idx.LongDescr, "\n\r"))
	ch.Sendf("  Level    : %d  Align: %d  Race: %d  Class: %d  Sex: %d\n\r", idx.Level, idx.Alignment, idx.Race, idx.Class, idx.Sex)
	ch.Sendf("  AC: %d  Hitroll: %d  Damroll: %d  Gold: %d  Exp: %d\n\r", idx.AC, idx.Hitroll, idx.Damroll, idx.Gold, idx.Exp)
	ch.Sendf("  HP: %dd%d+%d  Position: %d\n\r", idx.HitNoDice, idx.HitSizeDice, idx.HitPlus, idx.Position)
}

func sexFromName(s string) (int, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "neutral", "0":
		return types.SEX_NEUTRAL, true
	case "male", "1", "m":
		return types.SEX_MALE, true
	case "female", "2", "f":
		return types.SEX_FEMALE, true
	}
	return 0, false
}

