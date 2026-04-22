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
		// Menu-entry path (plan-phase6-olc-oedit.md §G10): `oedit <vnum>`
		// with no subcommand enters the interactive CON_OEDIT substate.
		// The flat subcommand form (`oedit <vnum> name foo`) is
		// preserved — only the no-subcommand branch routes into the menu.
		//
		// Go-port divergence from C: C `do_ooedit` (src/ooedit.c:109-208)
		// takes an OBJECT name/vnum of a LIVE INSTANCE; Go port has
		// always operated on the prototype by vnum (matches flat DoOedit
		// + DoOset).
		if ch.Desc == nil {
			// NPC or disconnected — defensive refusal matching DoRedit.
			ch.Send("No descriptor.\n\r")
			return
		}
		ch.Desc.Olc = &types.OlcData{
			Mode:   types.OEDIT_MAIN_MENU,
			Vnum:   vnum,
			Target: idx,
		}
		ch.Desc.Connected = int(types.CON_OEDIT)
		if OeditDispMenuFunc != nil {
			OeditDispMenuFunc(ch.Desc)
		}
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
	case "show":
		// Retained flat summary — formerly the no-arg default, now
		// available as an explicit subcommand. Menu-entry (the no-arg
		// path) replaced it per plan-phase6-olc-oedit.md §G10 / Q2.
		oeditShow(ch, idx)
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

// DoMedit implements the 'medit' command: edit a mob prototype.
// Usage:
//
//	medit <vnum>                 — open the interactive CON_MEDIT menu
//	medit <vnum> show            — flat one-shot summary (Wave 1-pre form)
//	medit <vnum> <sub> [args...] — flat single-field set (legacy MVP path)
//
// Plan: plan-phase6-olc-medit.md §G14. The no-arg branch enters the
// interactive editor. The `show` subcommand re-binds what the no-arg
// branch did pre-Wave-5 so the flat summary is still reachable.
//
// Scope cut: PC-by-name argument resolution (`medit <playername>`) is
// deferred to a future wave that lands the `worldPcLookup` seam.
// Today only NPC vnum is recognized — non-numeric arguments fall
// through to the existing "Vnum must be a positive number." rejection.
func DoMedit(ch *types.CharData, argument string) {
	if ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}
	vnumArg, rest := util.OneArgument(argument)
	if vnumArg == "" {
		ch.Send("Usage: medit <vnum> [subcommand] [args]\n\r")
		ch.Send("With no subcommand: opens the interactive editor.\n\r")
		ch.Send("Subcommands: create, show, name, short, long, desc, level, stats, hp, mana, mv, armor, hitroll, damroll, align, flags, race, class, sex, attacks, defenses, gold, xp, position, affected\n\r")
		return
	}
	vnum, err := strconv.Atoi(vnumArg)
	if err != nil {
		// Plan §G14 scope cut: PC-by-name lookup not yet supported
		// (worldPcLookup seam unland). Distinguish "non-numeric" (likely
		// a PC name) from "numeric but invalid (vnum<=0)" using the same
		// err the outer Atoi already produced — re-parsing the same
		// string would always re-fail the same way (LOW #4).
		ch.Send("PC editing by name not yet supported; pass an NPC vnum.\n\r")
		return
	}
	if vnum <= 0 {
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
		// Plan §G14: no-arg → enter interactive CON_MEDIT menu.
		// Mirrors the oedit Wave 4 pattern at olc_interactive.go case ""
		// for `oedit <vnum>`.
		if ch.Desc == nil {
			// NPC or disconnected — defensive refusal matching DoOedit.
			ch.Send("No descriptor.\n\r")
			return
		}
		// Double-edit guard: refuse if this descriptor is already in
		// CON_MEDIT. A second `medit <vnum>` while editing would clobber
		// the in-flight Olc state. C `do_omedit` (omedit.c:203-210) does
		// a global descriptor scan; here we only need to guard the
		// current descriptor since each descriptor owns its own Olc.
		if ch.Desc.Connected == int(types.CON_MEDIT) && ch.Desc.Olc != nil {
			ch.Send("You are already editing a mob. Type Q to exit first.\n\r")
			return
		}
		// Wrap the prototype in a bare CharData so the menu renderers
		// (which type-assert OlcData.Target to *CharData) can read fields
		// and the parse arms (which dual-write through victim.IndexData
		// when ACT_PROTOTYPE is set) propagate edits back to the
		// prototype.
		//
		// Wave-5 follow-up HIGH #1: do NOT use handler.CreateMobile
		// here. CreateMobile bumps idx.Count and appends to
		// w.Characters; cleanupOlc only resets descriptor state, so
		// every menu-entry would permanently leak one phantom mob —
		// blocking area resets that gate on `idx.Count >= reset.Arg2`
		// (handler/reset.go:62) and growing w.Characters with roomless
		// orphans iterated by every pulse (game/update.go).
		//
		// We bare-allocate and copy ONLY the fields the menu renderers
		// (game/medit_menu.go) read for display, plus IndexData and
		// ACT_PROTOTYPE so the existing dual-write arms (game/medit_arms.go)
		// continue to mirror edits to the prototype. Vital fields
		// (Hit/MaxHit/Mana/MaxMove/AC/etc.) are intentionally NOT
		// copied — the NPC main menu reads dice from victim.IndexData,
		// not victim.MaxHit, and direct stat edits on a prototype are
		// flat-mset territory. Any zero-valued display field that
		// surprises a future reader is a UX cosmetic; correctness lives
		// at victim.IndexData.
		//
		// C reference: do_omedit (omedit.c:109-232) operates on a LIVE
		// mob instance from get_char_world, never wrapping the
		// prototype. The Go port's wrapper-around-prototype model is a
		// deliberate divergence — see plan-phase6-olc-medit.md §G14.
		victim := &types.CharData{
			Name:              idx.PlayerName,
			ShortDescr:        idx.ShortDescr,
			LongDescr:         idx.LongDescr,
			Description:       idx.Description,
			IndexData:         idx,
			Sex:               idx.Sex,
			Level:             idx.Level,
			Position:          idx.Position,
			DefPosition:       idx.DefPosition,
			Race:              idx.Race,
			Class:             idx.Class,
			Alignment:         idx.Alignment,
			Hitroll:           idx.Hitroll,
			Damroll:           idx.Damroll,
			Gold:              idx.Gold,
			Silver:            idx.Silver,
			Copper:            idx.Copper,
			Exp:               idx.Exp,
			SpecFun:           idx.SpecFun,
			XFlags:            idx.XFlags,
			Immune:            idx.Immune,
			Resistant:         idx.Resistant,
			Susceptible:       idx.Susceptible,
			Attacks:           idx.Attacks,
			Defenses:          idx.Defenses,
			Speaks:            idx.Speaks,
			Speaking:          idx.Speaking,
			AffectedBy:        idx.AffectedBy,
			PermStr:           idx.PermStr,
			PermInt:           idx.PermInt,
			PermWis:           idx.PermWis,
			PermDex:           idx.PermDex,
			PermCon:           idx.PermCon,
			PermCha:           idx.PermCha,
			PermLck:           idx.PermLck,
			SavingPoisonDeath: idx.SavingPoisonDeath,
			SavingWand:        idx.SavingWand,
			SavingParaPetri:   idx.SavingParaPetri,
			SavingBreath:      idx.SavingBreath,
			SavingSpellStaff:  idx.SavingSpellStaff,
		}
		// Act bitvector is the canonical edit target — copy from idx
		// then set ACT_PROTOTYPE so the existing arms' dual-write
		// through victim.IndexData kicks in (Wave 3 G7 / G8 mirror
		// pattern). ACT_IS_NPC must remain set so victim.IsNPC()
		// returns true and the dispatcher routes to the NPC menu.
		victim.Act = idx.Act
		victim.Act.Set(types.ACT_IS_NPC)
		victim.Act.Set(types.ACT_PROTOTYPE)
		ch.Desc.Olc = &types.OlcData{
			Mode:   types.MEDIT_NPC_MAIN_MENU,
			Vnum:   vnum,
			Target: victim,
		}
		ch.Desc.Connected = int(types.CON_MEDIT)
		if MeditDispMenuFunc != nil {
			MeditDispMenuFunc(ch.Desc)
		}
	case "show":
		// Wave 5: the old default pre-G14 summary, re-bound under an
		// explicit subcommand so flat inspection is still reachable.
		// Mirrors `oedit <vnum> show` (olc_interactive.go case "show").
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
