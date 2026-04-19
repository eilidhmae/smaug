// Package act — polymorph (morph subsystem) commands and admin tools.
//
// This file contains the user-facing command wrappers (morph / unmorph /
// morphset / morphstat / morphcreate / morphdestroy). The stat-application
// primitives (DoMorph / DoUnmorph / SendMorphMessage / MakeCharMorph /
// CanMorph / UnmorphAll / DoMorphChar / DoUnmorphChar) live in
// internal/handler/polymorph.go so both this package and the mudprog
// package (mpmorph / mpunmorph bodies) can reach them without an import
// cycle. See plan-phase6-polymorph.md §4.7.
package act

import (
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// MorphFilePath is the on-disk location of db/system/morph.dat. Written
// once at boot; read by DoMorphset "save" / DoMorphcreate / DoMorphdestroy
// to persist the morph table. Mirrors act.HolidayFilePath (holidays) and
// act.PlanesFilePath (planes) at their respective set-at-boot declaration
// sites. Set-at-boot, read-from-game-loop — no synchronization needed.
var MorphFilePath string

// DoMorph is the immortal `morph <vnum>` / `morph <vnum> <target>` command.
// Mirrors C do_imm_morph at src/polymorph.c:2676-2723. Trust gate is the
// registered Level in boot.go (LEVEL_IMMORTAL); inside the handler we
// additionally check the standard get_trust comparison when a victim is
// named (C :2713-2718).
func DoMorph(ch *types.CharData, argument string) {
	if ch == nil {
		return
	}
	if ch.IsNPC() {
		ch.Send("Only player characters can use this command.\n\r")
		return
	}
	arg, rest := util.OneArgument(argument)
	arg2, _ := util.OneArgument(rest)

	vnum, err := strconv.Atoi(arg)
	if err != nil || vnum == 0 {
		ch.Send("Syntax: morph <vnum>\n\r")
		return
	}
	if WorldRef == nil {
		return
	}
	m := WorldRef.GetMorphVnum(vnum)
	if m == nil {
		ch.Sendf("No such morph %d exists.\n\r", vnum)
		return
	}
	if arg2 == "" {
		handler.DoMorphChar(WorldRef, ch, m)
		ch.Send("Done.\n\r")
		return
	}
	victim := handler.GetCharWorld(WorldRef, ch, arg2)
	if victim == nil {
		ch.Send("No one like that in all the realms.\n\r")
		return
	}
	if !victim.IsNPC() && ch.GetTrust() < victim.GetTrust() {
		ch.Send("You can't do that!\n\r")
		return
	}
	handler.DoMorphChar(WorldRef, victim, m)
	ch.Send("Done.\n\r")
}

// DoUnmorph is the immortal `unmorph [target]` command. Mirrors C
// do_imm_unmorph at src/polymorph.c:2729-2753.
func DoUnmorph(ch *types.CharData, argument string) {
	if ch == nil {
		return
	}
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		handler.DoUnmorphChar(ch)
		ch.Send("Done.\n\r")
		return
	}
	if WorldRef == nil {
		return
	}
	victim := handler.GetCharWorld(WorldRef, ch, arg)
	if victim == nil {
		ch.Send("No one like that in all the realms.\n\r")
		return
	}
	if !victim.IsNPC() && ch.GetTrust() < victim.GetTrust() {
		ch.Send("You can't do that!\n\r")
		return
	}
	handler.DoUnmorphChar(victim)
	ch.Send("Done.\n\r")
}

// DoMorphstat — pretty-print a morph's fields. Mirrors C do_morphstat at
// src/polymorph.c:969-1160. Subcommands:
//
//   - `morphstat list`           → one-line-per-morph summary
//   - `morphstat <name-or-vnum>` → full restriction + enhancement dump
//   - `morphstat <name> help`    → description + help text
//   - `morphstat <name> desc`    → same as help
func DoMorphstat(ch *types.CharData, argument string) {
	if ch == nil {
		return
	}
	if ch.IsNPC() {
		ch.Send("Mob's can't morphstat\n\r")
		return
	}
	arg, rest := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Morphstat what?\n\r")
		return
	}
	if WorldRef == nil {
		return
	}
	if strings.EqualFold(arg, "list") {
		if len(WorldRef.Morphs) == 0 {
			ch.Send("No morph's currently exist.\n\r")
			return
		}
		for i, m := range WorldRef.Morphs {
			ch.Sendf("[%2d]   Name:  %-13s    Vnum:  %4d  Used:  %3d\n\r",
				i+1, m.Name, m.Vnum, m.Used)
		}
		return
	}
	m := resolveMorphArg(arg)
	if m == nil {
		ch.Send("No such morph exists.\n\r")
		return
	}
	if rest == "" {
		morphstatFields(ch, m)
		return
	}
	second, _ := util.OneArgument(rest)
	if strings.EqualFold(second, "help") || strings.EqualFold(second, "desc") {
		morphstatHelp(ch, m)
		return
	}
	ch.Send("Syntax: morphstat <morph>\n\r")
	ch.Send("Syntax: morphstat <morph> <help/desc>\n\r")
}

// resolveMorphArg looks up a morph by vnum (if arg is numeric) else by
// name. Both lookups are case-insensitive for name.
func resolveMorphArg(arg string) *types.MorphData {
	if WorldRef == nil {
		return nil
	}
	if v, err := strconv.Atoi(arg); err == nil {
		return WorldRef.GetMorphVnum(v)
	}
	return WorldRef.GetMorph(arg)
}

// morphstatFields emits the C-fidelity "Morph Restrictions + Enhancements
// + Affects" panel. Colors dropped for readability in Go port; structure
// matches src/polymorph.c:1017-1117.
func morphstatFields(ch *types.CharData, m *types.MorphData) {
	divider := "[----------------------------------------------------------------------------]\n\r"
	ch.Sendf("  Morph Name: %-20s  Vnum: %4d\n\r", m.Name, m.Vnum)
	ch.Send(divider)
	ch.Send("                           Morph Restrictions\n\r")
	ch.Send(divider)
	ch.Sendf("  Classes Allowed   : %s\n\r", morphClassNames(m.Class))
	ch.Sendf("  Races Not Allowed : %s\n\r", morphRaceNames(m.Race))
	ch.Sendf("  Sex:  %s   Pkill:   %s   Time From:   %d   Time To:    %d\n\r",
		morphSexLabel(m.Sex), morphPkillLabel(m.PKill), m.TimeFrom, m.TimeTo)
	ch.Sendf("  Day From:  %d  Day To:  %d\n\r", m.DayFrom, m.DayTo)
	ch.Sendf("  Level:  %d       Casting Allowed   : %s\n\r",
		m.Level, noCastLabel(m.NoCast))
	ch.Sendf("  USAGES:  Mana:  %d  Move:  %d  Hp:  %d  Favour:  %d  Glory:  %d\n\r",
		m.ManaUsed, m.MoveUsed, m.HpUsed, m.FavourUsed, m.GloryUsed)
	ch.Sendf("           Blood:  %d\n\r", m.BloodUsed)
	ch.Sendf("  Obj1: %d  Objuse1: %s   Obj2: %d  Objuse2: %s   Obj3: %d  Objuse3: %s\n\r",
		m.Obj[0], yesNoLabel(m.ObjUse[0]),
		m.Obj[1], yesNoLabel(m.ObjUse[1]),
		m.Obj[2], yesNoLabel(m.ObjUse[2]))
	ch.Sendf("  Timer: %d\n\r", m.Timer)
	ch.Send(divider)
	ch.Send("                       Enhancements to the Player\n\r")
	ch.Send(divider)
	ch.Sendf("  Str: %2d )( Int: %2d )( Wis: %2d )( Dex: %2d )( Con: %2d )( Cha: %2d )( Lck: %2d )\n\r",
		m.Str, m.Int, m.Wis, m.Dex, m.Con, m.Cha, m.Lck)
	ch.Sendf("  Save versus: %d %d %d %d %d       Dodge: %d  Parry: %d  Tumble: %d\n\r",
		m.SavingPoisonDeath, m.SavingWand, m.SavingParaPetri,
		m.SavingBreath, m.SavingSpellStaff, m.Dodge, m.Parry, m.Tumble)
	ch.Sendf("  Hps     : %s    Blood  : %s    Mana   : %s    Move      : %s\n\r",
		m.Hit, m.Blood, m.Mana, m.Move)
	ch.Sendf("  Damroll : %s    Hitroll: %s    AC     : %d\n\r",
		m.Damroll, m.Hitroll, m.AC)
	ch.Send(divider)
	ch.Send("                          Affects to the Player\n\r")
	ch.Send(divider)
	ch.Sendf("  Affected by: %s\n\r", m.AffectedBy.String())
	ch.Sendf("  Immune     : %d\n\r", m.Immune)
	ch.Sendf("  Susceptible: %d\n\r", m.Suscept)
	ch.Sendf("  Resistant  : %d\n\r", m.Resistant)
	ch.Sendf("  Skills     : %s\n\r", m.Skills)
	ch.Send(divider)
	ch.Send("                     Prevented affects to the Player\n\r")
	ch.Send(divider)
	ch.Sendf("  Not affected by: %s\n\r", m.NoAffectedBy.String())
	ch.Sendf("  Not Immune     : %d\n\r", m.NoImmune)
	ch.Sendf("  Not Susceptible: %d\n\r", m.NoSuscept)
	ch.Sendf("  Not Resistant  : %d\n\r", m.NoResistant)
	ch.Sendf("  Not Skills     : %s\n\r", m.NoSkills)
	ch.Send(divider)
	ch.Send("\n\r")
}

func morphstatHelp(ch *types.CharData, m *types.MorphData) {
	divider := "[----------------------------------------------------------------------------]\n\r"
	ch.Sendf("  Morph Name  : %-20s\n\r", m.Name)
	ch.Sendf("  Default Pos : %d\n\r", m.DefPos)
	ch.Sendf("  Keywords    : %s\n\r", m.KeyWords)
	sd := m.ShortDesc
	if sd == "" {
		sd = "(none set)"
	}
	ch.Sendf("  Shortdesc   : %s\n\r", sd)
	ld := m.LongDesc
	if ld == "" {
		ld = "(none set)\n\r"
	}
	ch.Sendf("  Longdesc    : %s", ld)
	ch.Sendf("  Morphself   : %s\n\r", m.MorphSelf)
	ch.Sendf("  Morphother  : %s\n\r", m.MorphOther)
	ch.Sendf("  UnMorphself : %s\n\r", m.UnmorphSelf)
	ch.Sendf("  UnMorphother: %s\n\r", m.UnmorphOther)
	ch.Send(divider)
	ch.Sendf("                                  Help:\n\r%s\n\r", m.Help)
	ch.Send(divider)
	ch.Sendf("                               Description:\n\r%s\n\r", m.Description)
	ch.Send(divider)
}

func morphSexLabel(sex int) string {
	switch sex {
	case types.SEX_MALE:
		return "male"
	case types.SEX_FEMALE:
		return "female"
	default:
		return "neutral"
	}
}

func morphPkillLabel(p int) string {
	switch p {
	case types.ONLY_PKILL:
		return "YES"
	case types.ONLY_PEACEFULL:
		return "NO"
	default:
		return "n/a"
	}
}

func noCastLabel(b bool) string {
	if b {
		return "NO"
	}
	return "yes"
}

func yesNoLabel(b bool) string {
	if b {
		return "YES"
	}
	return "no"
}

func morphClassNames(mask int) string {
	if WorldRef == nil || mask == 0 {
		return ""
	}
	var out []string
	for i, c := range WorldRef.Classes {
		if c == nil {
			continue
		}
		if mask&(1<<uint(i)) != 0 {
			out = append(out, c.WhoName)
		}
	}
	return strings.Join(out, " ")
}

func morphRaceNames(mask int) string {
	if WorldRef == nil || mask == 0 {
		return ""
	}
	var out []string
	for i, r := range WorldRef.Races {
		if r == nil {
			continue
		}
		if mask&(1<<uint(i)) != 0 {
			out = append(out, r.Name)
		}
	}
	return strings.Join(out, " ")
}

// DoMorphcreate — allocate a blank morph, assign a fresh vnum ≥1000,
// append to world.Morphs. Mirrors C do_morphcreate at src/polymorph.c:
// 2313-2362. The optional `copy` subcommand clones an existing morph;
// not implemented in this port (deferred; typical usage is morphset
// field-by-field after create).
func DoMorphcreate(ch *types.CharData, argument string) {
	if ch == nil {
		return
	}
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Usage: morphcreate <name>\n\r")
		return
	}
	if WorldRef == nil {
		return
	}
	// Reject duplicates by name (C doesn't, but the downstream resolver
	// uses first-match and the admin UI would be surprising otherwise).
	if existing := WorldRef.GetMorph(arg); existing != nil {
		ch.Sendf("Morph %q already exists (vnum %d).\n\r", existing.Name, existing.Vnum)
		return
	}
	m := persist.MorphDefaults()
	m.Name = util.SmashTilde(arg)
	// Assign vnum via the same counter that SetupMorphVnum maintains.
	m.Vnum = WorldRef.MorphVnumCounter
	WorldRef.MorphVnumCounter++
	if m.Vnum < 1000 {
		// First create on a tree whose setup_morph_vnum ran on an
		// empty table — counter starts at 1000, so this branch is
		// essentially dead code. Defensive.
		m.Vnum = 1000
		WorldRef.MorphVnumCounter = 1001
	}
	WorldRef.Morphs = append(WorldRef.Morphs, m)
	ch.Sendf("Morph %s created with vnum %d.\n\r", m.Name, m.Vnum)
}

// DoMorphdestroy — find morph, force-unmorph all users, remove from
// table, save. Mirrors C do_morphdestroy at src/polymorph.c:2370-2392.
// C semantics: unmorph_all runs BEFORE the UNLINK so no PC is left
// pointing at a freed morph pointer.
func DoMorphdestroy(ch *types.CharData, argument string) {
	if ch == nil {
		return
	}
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Usage: morphdestroy <name-or-vnum>\n\r")
		return
	}
	if WorldRef == nil {
		return
	}
	m := resolveMorphArg(arg)
	if m == nil {
		ch.Sendf("Unknown morph %s.\n\r", arg)
		return
	}
	handler.UnmorphAll(WorldRef, m)
	// Splice out by identity.
	for i, mm := range WorldRef.Morphs {
		if mm == m {
			WorldRef.Morphs = append(WorldRef.Morphs[:i], WorldRef.Morphs[i+1:]...)
			break
		}
	}
	ch.Send("Morph deleted.\n\r")
	if MorphFilePath != "" {
		if err := persist.SaveMorphs(MorphFilePath, WorldRef.Morphs); err != nil {
			util.Bug("DoMorphdestroy: SaveMorphs: %v", err)
		}
	}
}

// DoMorphset — field-editor command. Mirrors C do_morphset at
// src/polymorph.c:72-960 (889 LOC of if/else-if per-field dispatch).
//
// Scope split: G4a ships the dispatcher + subcommand switch + `save`/
// `list`/`?` help + the 5-field reduced surface (name, vnum, shortdesc,
// level, ac) that proves the pattern. G4b extends to the full field
// coverage (~70 fields). Marker comments flag where G4b inserts new
// branches.
//
// Arg layout for value-setting: `morphset <morph> <field> <value>`
// (three-arg form). The C substate / `on` mode that enables a two-arg
// repeat-edit path is NOT ported (the Go port uses flat per-command
// invocations).
func DoMorphset(ch *types.CharData, argument string) {
	if ch == nil {
		return
	}
	if ch.IsNPC() {
		ch.Send("Mob's can't morphset\n\r")
		return
	}
	arg1, rest1 := util.OneArgument(argument)
	arg2, rest2 := util.OneArgument(rest1)
	arg3, _ := util.OneArgument(rest2)

	// `morphset save` — writes the full table to disk.
	if strings.EqualFold(arg1, "save") {
		if WorldRef == nil || MorphFilePath == "" {
			ch.Send("Morph data not writable (no table path).\n\r")
			return
		}
		if err := persist.SaveMorphs(MorphFilePath, WorldRef.Morphs); err != nil {
			ch.Sendf("Save failed: %v\n\r", err)
			util.Bug("DoMorphset: SaveMorphs: %v", err)
			return
		}
		ch.Send("Morph data saved.\n\r")
		return
	}

	if arg1 == "" || arg2 == "" || arg1 == "?" {
		morphsetSyntax(ch)
		return
	}

	m := resolveMorphArg(arg1)
	if m == nil {
		ch.Send("That morph does not exist.\n\r")
		return
	}

	value := -1
	if v, err := strconv.Atoi(arg3); err == nil {
		value = v
	}

	if !applyMorphsetField(ch, m, arg2, arg3, value) {
		morphsetSyntax(ch)
	}
}

// morphsetSyntax prints the help/syntax block. Mirrors src/polymorph.c:
// 210-241.
func morphsetSyntax(ch *types.CharData) {
	ch.Send("Syntax: morphset <morph> <field>  <value>\n\r")
	ch.Send("Syntax: morphset save\n\r")
	ch.Send("\n\r")
	ch.Send("Field being one of:\n\r")
	ch.Send("-------------------------------------------------\n\r")
	ch.Send("  ac, affected, blood, bloodused, cha, class, con, damroll, dayto,\n\r")
	ch.Send("  dayfrom, deity, description, defpos, dex, dodge,\n\r")
	ch.Send("  favourused, gloryused, help, hitroll, hp, hpused, immune,\n\r")
	ch.Send("  int, str, keyword, lck, level, long, mana, manaused,\n\r")
	ch.Send("  morphother, morphself, move, moveused, name, noaffected,\n\r")
	ch.Send("  nocast, noimmune, noresistant, noskill, nosusceptible,\n\r")
	ch.Send("  obj1, obj2, obj3, objuse1, objuse2, objuse3, parry,\n\r")
	ch.Send("  pkill, race, resistant, sav1, sav2, sav3, sav4, sav5,\n\r")
	ch.Send("  sex, short, skills, susceptible, timefrom, timer, timeto,\n\r")
	ch.Send("  tumble, unmorphother, unmorphself, wis.\n\r")
	ch.Send("-------------------------------------------------\n\r")
}

// applyMorphsetField dispatches one field-setter. Returns true if the
// field name was recognized (regardless of whether the value was valid
// — range errors print their own message and return true). Returns false
// on unknown field, letting the caller print the syntax block.
//
// The grid of fields mirrors src/polymorph.c:276-959. G4b pattern: add
// case "<fieldname>" cases in alphabetic clusters, mirroring C's
// if/else-if order where it matters (e.g. the `obj`/`objuse` prefix
// handlers must come before generic setters).
func applyMorphsetField(ch *types.CharData, m *types.MorphData, field, arg3 string, value int) bool {
	f := strings.ToLower(field)
	switch f {
	// --- G4a minimum surface (pattern-prover) -------------------------
	case "name":
		if arg3 == "" {
			ch.Send("Name must be non-empty.\n\r")
			return true
		}
		m.Name = util.SmashTilde(arg3)
	case "short":
		m.ShortDesc = util.SmashTilde(arg3)
	case "level":
		if value < 0 || value > types.LEVEL_IMMORTAL {
			ch.Sendf("Level range is 0 to %d.\n\r", types.LEVEL_IMMORTAL)
			return true
		}
		m.Level = value
	case "ac":
		if value > 500 || value < -500 {
			ch.Send("Ac range is -500 to 500.\n\r")
			return true
		}
		m.AC = value

	// --- G4b attributes (polymorph.c:276-349) -------------------------
	case "str":
		if !rangeOK10(ch, value, "Strength") {
			return true
		}
		m.Str = value
	case "int":
		if !rangeOK10(ch, value, "Intelligence") {
			return true
		}
		m.Int = value
	case "wis":
		if !rangeOK10(ch, value, "Wisdom") {
			return true
		}
		m.Wis = value
	case "dex":
		if !rangeOK10(ch, value, "Dexterity") {
			return true
		}
		m.Dex = value
	case "con":
		if !rangeOK10(ch, value, "Constitution") {
			return true
		}
		m.Con = value
	case "cha":
		if !rangeOK10(ch, value, "Charisma") {
			return true
		}
		m.Cha = value
	case "lck":
		if !rangeOK10(ch, value, "Luck") {
			return true
		}
		m.Lck = value

	// --- defpos / sex / pkill -----------------------------------------
	case "defpos":
		if value < 0 || value > types.POS_STANDING {
			ch.Sendf("Position range is 0 to %d.\n\r", types.POS_STANDING)
			return true
		}
		m.DefPos = value
	case "sex":
		if (value < 0 || value > 2) && value != -1 {
			ch.Send("Sex must be a value from 0 to 2.\n\r")
			return true
		}
		m.Sex = value
	case "pkill":
		switch strings.ToLower(arg3) {
		case "pkill":
			m.PKill = types.ONLY_PKILL
		case "peace":
			m.PKill = types.ONLY_PEACEFULL
		case "none":
			m.PKill = 0
		default:
			ch.Send("Usuage: morphset <morph> pkill [pkill|peace|none]\n\r")
		}

	// --- resource usage (polymorph.c:374-427) --------------------------
	case "bloodused":
		if !rangeOK(ch, value, 0, 60, "Blood used") {
			return true
		}
		m.BloodUsed = value
	case "manaused":
		if !rangeOK(ch, value, 0, 2000, "Mana used") {
			return true
		}
		m.ManaUsed = value
	case "moveused":
		if !rangeOK(ch, value, 0, 2000, "Move used") {
			return true
		}
		m.MoveUsed = value
	case "hpused":
		if !rangeOK(ch, value, 0, 2000, "Hp used") {
			return true
		}
		m.HpUsed = value
	case "favourused":
		if !rangeOK(ch, value, 0, 2000, "Favour used") {
			return true
		}
		m.FavourUsed = value
	case "gloryused":
		if !rangeOK(ch, value, 0, 2000, "Glory used") {
			return true
		}
		m.GloryUsed = value

	// --- time/day windows (polymorph.c:428-463) ------------------------
	case "timeto":
		if !rangeOK(ch, value, 0, 23, "Timeto") {
			return true
		}
		m.TimeTo = value
	case "timefrom":
		if !rangeOK(ch, value, 0, 23, "Timefrom") {
			return true
		}
		m.TimeFrom = value
	case "dayto":
		if !rangeOK(ch, value, 0, 31, "Dayto") {
			return true
		}
		m.DayTo = value
	case "dayfrom":
		if !rangeOK(ch, value, 0, 31, "Dayfrom") {
			return true
		}
		m.DayFrom = value

	// --- saves (polymorph.c:464-508) ----------------------------------
	case "sav1", "savepoison":
		if !rangeOK(ch, value, -30, 30, "Saving throw") {
			return true
		}
		m.SavingPoisonDeath = value
	case "sav2", "savewand":
		if !rangeOK(ch, value, -30, 30, "Saving throw") {
			return true
		}
		m.SavingWand = value
	case "sav3", "savepara":
		if !rangeOK(ch, value, -30, 30, "Saving throw") {
			return true
		}
		m.SavingParaPetri = value
	case "sav4", "savebreath":
		if !rangeOK(ch, value, -30, 30, "Saving throw") {
			return true
		}
		m.SavingBreath = value
	case "sav5", "savestaff":
		if !rangeOK(ch, value, -30, 30, "Saving throw") {
			return true
		}
		m.SavingSpellStaff = value

	// --- timer (polymorph.c:509-517) ----------------------------------
	case "timer":
		if value < -1 || value == 0 {
			ch.Send("Timer must be -1 (None) or greater than 0.\n\r")
			return true
		}
		m.Timer = value

	// --- dice-string fields (polymorph.c:518-544, :555-568) -----------
	case "hp":
		m.Hit = zeroAsEmpty(arg3)
	case "mana":
		m.Mana = zeroAsEmpty(arg3)
	case "move":
		m.Move = zeroAsEmpty(arg3)
	case "blood":
		m.Blood = zeroAsEmpty(arg3)
	case "hitroll":
		m.Hitroll = zeroAsEmpty(arg3)
	case "damroll":
		m.Damroll = zeroAsEmpty(arg3)

	// --- dodge/parry/tumble (polymorph.c:569-620) ---------------------
	case "dodge":
		if !rangeOK(ch, value, -100, 100, "Dodge") {
			return true
		}
		m.Dodge = value
	case "parry":
		if !rangeOK(ch, value, -100, 100, "Dodge") {
			return true
		}
		m.Parry = value
	case "tumble":
		if !rangeOK(ch, value, -100, 100, "Dodge") {
			return true
		}
		m.Tumble = value

	// --- obj / objuse (polymorph.c:578-652) ---------------------------
	case "obj1", "obj2", "obj3":
		idx := int(f[3] - '0')
		if WorldRef != nil {
			if _, ok := WorldRef.ObjIndex[value]; !ok {
				ch.Send("No such vnum.\n\r")
				return true
			}
		}
		m.Obj[idx-1] = value
	case "objuse1", "objuse2", "objuse3":
		idx := int(f[6] - '0')
		m.ObjUse[idx-1] = value != 0

	// --- RIS scalars (polymorph.c:653-735) ----------------------------
	case "immune":
		m.Immune = value
	case "resistant":
		m.Resistant = value
	case "susceptible", "suscept":
		m.Suscept = value
	case "noimmune":
		m.NoImmune = value
	case "noresistant":
		m.NoResistant = value
	case "nosusceptible", "nosuscept":
		m.NoSuscept = value

	// --- affects BitVector (polymorph.c:736-760) ----------------------
	case "affected":
		bv, err := types.ParseBitVector(arg3)
		if err != nil {
			ch.Sendf("Bad bitvector: %v\n\r", err)
			return true
		}
		m.AffectedBy = bv
	case "noaffected":
		bv, err := types.ParseBitVector(arg3)
		if err != nil {
			ch.Sendf("Bad bitvector: %v\n\r", err)
			return true
		}
		m.NoAffectedBy = bv

	// --- long-form strings (polymorph.c:761-860) ----------------------
	case "keyword":
		m.KeyWords = util.SmashTilde(arg3)
	case "deity":
		m.Deity = util.SmashTilde(arg3)
	case "skills":
		m.Skills = util.SmashTilde(arg3)
	case "noskill", "noskills":
		m.NoSkills = util.SmashTilde(arg3)
	case "morphself":
		m.MorphSelf = util.SmashTilde(arg3)
	case "morphother":
		m.MorphOther = util.SmashTilde(arg3)
	case "unmorphself":
		m.UnmorphSelf = util.SmashTilde(arg3)
	case "unmorphother":
		m.UnmorphOther = util.SmashTilde(arg3)
	case "long":
		m.LongDesc = util.SmashTilde(arg3)

	// --- class / race setters (polymorph.c:820-889) -------------------
	// Space-separated list of class who_names / race names.
	case "class":
		mask := 0
		for _, tok := range strings.Fields(arg3) {
			if WorldRef != nil {
				for i, c := range WorldRef.Classes {
					if c != nil && strings.EqualFold(c.WhoName, tok) {
						mask |= 1 << uint(i)
						break
					}
				}
			}
		}
		m.Class = mask
	case "race":
		mask := 0
		for _, tok := range strings.Fields(arg3) {
			if WorldRef != nil {
				for i, r := range WorldRef.Races {
					if r != nil && strings.EqualFold(r.Name, tok) {
						mask |= 1 << uint(i)
						break
					}
				}
			}
		}
		m.Race = mask

	// --- nocast / description / help (polymorph.c:890-960) ------------
	case "nocast":
		m.NoCast = value != 0
	case "description":
		m.Description = util.SmashTilde(arg3)
	case "help":
		m.Help = util.SmashTilde(arg3)

	default:
		return false
	}
	// Acknowledge the change for any branch that fell through without
	// its own user-facing message. Most C branches are silent on
	// success too, so we mirror that.
	return true
}

func rangeOK(ch *types.CharData, value, lo, hi int, label string) bool {
	if value < lo || value > hi {
		ch.Sendf("%s must be a value from %d to %d.\n\r", label, lo, hi)
		return false
	}
	return true
}

func rangeOK10(ch *types.CharData, value int, label string) bool {
	return rangeOK(ch, value, -10, 10, label)
}

func zeroAsEmpty(s string) string {
	if s == "0" {
		return ""
	}
	return s
}
