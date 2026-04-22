// Package game — interactive mob/character editor (CON_MEDIT substate) menu
// renderers.
//
// Wave 2 ports src/omedit.c medit_disp_menu / medit_disp_npc_menu /
// medit_disp_pc_menu plus the 16 submenu renderers (sex, pos, default_pos,
// attack, defense, spec, class, race, save, affect, npc_flags, pc_flags,
// aff_flags, pcdata_flags, parts, ris) into Go. The dispatcher MeditDispMenu
// mirrors C :814-825 — branches on victim.IsNPC() (the Go analog of C's
// IS_NPC macro on CharData).
//
// Submenu renderers in Wave 2 are display-only scaffolds — the PARSE arms
// that consume their input land in later waves (G7 simple fields,
// G8 stats, G9 bitmask editors, G10 affect editor, G11 save editor,
// G12 class/race, G13 password). Each submenu writes its current state +
// the input hint; the parse-arm stubs simply re-render the main menu when
// a non-dispatched mode receives input.
//
// The ANSI screen-clear sequence C emits at the top of each menu
// ("50\x1B[;H\x1B[2J") is deliberately omitted per redit/oedit precedent —
// cosmetic only and breaks testclient determinism.
//
// Plan: plan-phase6-olc-medit.md §G2 / §G3.
package game

import (
	"fmt"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// MeditDispMenu is the single entry point the rest of the codebase calls
// to redisplay the medit main menu for a descriptor. It dispatches to the
// NPC or PC sub-renderer based on the current victim pointer stashed in
// d.Olc.Target. Mirrors C medit_disp_menu at src/omedit.c:814-825.
//
// Defensive: if d, d.Olc, or the target is absent / the wrong type, the
// call is a no-op. In valid flow this cannot happen (meditParse's
// type-assertion guard fires first), but the menu entry point can be
// reached from elsewhere (e.g. the eventual DoMedit no-arg path) and so
// double-checks.
func MeditDispMenu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	victim, ok := d.Olc.Target.(*types.CharData)
	if !ok || victim == nil {
		return
	}
	if victim.IsNPC() {
		meditDispNpcMenu(d)
		return
	}
	meditDispPcMenu(d)
}

// sexLabel returns the human-readable sex name used in menu headers.
// Matches C's ternary chain at omedit.c:847 / :894.
func sexLabel(sex int) string {
	switch sex {
	case types.SEX_MALE:
		return "male"
	case types.SEX_FEMALE:
		return "female"
	default:
		return "neutral"
	}
}

// positionName returns the C-parity position label. Index mirrors
// src/const.c position_names[] (POS_DEAD..POS_DRAG).
func positionName(pos int) string {
	names := []string{
		"dead", "mortally wounded", "incapacitated", "stunned",
		"sleeping", "berserk", "resting", "aggressive",
		"sitting", "fighting", "defensive", "evasive",
		"standing", "mounted", "shove", "drag",
	}
	if pos < 0 || pos >= len(names) {
		return "unknown"
	}
	return names[pos]
}

// meditDispNpcMenu renders the NPC main menu. Ports C
// src/omedit.c:827-885 verbatim in label content; cosmetic ANSI clear and
// color tags are dropped. Fields read from victim.IndexData for dice
// (NPC prototype data) and from victim directly for per-instance stats.
//
// All label strings in this function are load-bearing for test assertions
// (A5 regex scan) — do not rename without updating medit_wave2_test.go.
func meditDispNpcMenu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	victim, ok := d.Olc.Target.(*types.CharData)
	if !ok || victim == nil {
		return
	}

	vnum := d.Olc.Vnum
	if victim.IndexData != nil && victim.IndexData.Vnum != 0 {
		vnum = victim.IndexData.Vnum
	}
	d.WriteToBuffer(fmt.Sprintf("-- Mob Number:  [%d]\n\r", vnum))

	sdesc := ""
	ldesc := ""
	if victim.IndexData != nil {
		sdesc = victim.IndexData.ShortDescr
		ldesc = victim.IndexData.LongDescr
	}
	if sdesc == "" {
		sdesc = victim.ShortDescr
	}
	if ldesc == "" {
		ldesc = victim.LongDescr
	}
	if sdesc == "" {
		sdesc = "(none set)"
	}
	if ldesc == "" {
		ldesc = "(none set)"
	}

	d.WriteToBuffer(fmt.Sprintf("1) Sex: %-7s          2) Name: %s\n\r",
		sexLabel(victim.Sex), victim.Name))
	d.WriteToBuffer(fmt.Sprintf("3) Shortdesc: %s\n\r", sdesc))
	d.WriteToBuffer(fmt.Sprintf("4) Longdesc:-\n\r%s\n\r", ldesc))
	desc := victim.Description
	if desc == "" {
		desc = "(none set)"
	}
	d.WriteToBuffer(fmt.Sprintf("5) Description:-\n\r%s\n\r", desc))

	d.WriteToBuffer(fmt.Sprintf("6) Class: [%d], 7) Race:   [%d]\n\r",
		victim.Class, victim.Race))
	d.WriteToBuffer(fmt.Sprintf(
		"8) Level:       [%5d], 9) Alignment:    [%5d], A) Strength: [%5d]\n\r",
		victim.Level, victim.Alignment, victim.GetCurrStr()))
	d.WriteToBuffer(fmt.Sprintf(
		"B) Intelligence:[%5d], C) Wisdom:       [%5d], D) Dexterity:[%5d]\n\r",
		victim.GetCurrInt(), victim.GetCurrWis(), victim.GetCurrDex()))
	d.WriteToBuffer(fmt.Sprintf(
		"E) Constitution:[%5d], F) Charisma:     [%5d], G) Luck:     [%5d]\n\r",
		victim.GetCurrCon(), victim.GetCurrCha(), victim.GetCurrLck()))

	var damNum, damSize, damPlus, hitNum, hitSize, hitPlus int
	if victim.IndexData != nil {
		damNum = victim.IndexData.DamNoDice
		damSize = victim.IndexData.DamSizeDice
		damPlus = victim.IndexData.DamPlus
		hitNum = victim.IndexData.HitNoDice
		hitSize = victim.IndexData.HitSizeDice
		hitPlus = victim.IndexData.HitPlus
	}
	d.WriteToBuffer(fmt.Sprintf(
		"H) DamNumDice:  [%5d], I) DamSizeDice:  [%5d], J) DamPlus:  [%5d]\n\r",
		damNum, damSize, damPlus))
	d.WriteToBuffer(fmt.Sprintf(
		"K) HitNumDice:  [%5d], L) HitSizeDice:  [%5d], M) HitPlus:  [%5d]\n\r",
		hitNum, hitSize, hitPlus))
	d.WriteToBuffer(fmt.Sprintf("N) Gold:     [%8d], O) Spec: %s\n\r",
		victim.Gold, victim.SpecFun))

	d.WriteToBuffer("P) Saving Throws\n\r")
	d.WriteToBuffer(fmt.Sprintf("R) Resistant   : %d\n\r", victim.Resistant))
	d.WriteToBuffer(fmt.Sprintf("S) Immune      : %d\n\r", victim.Immune))
	d.WriteToBuffer(fmt.Sprintf("T) Susceptible : %d\n\r", victim.Susceptible))
	d.WriteToBuffer(fmt.Sprintf("U) Position    : %s\n\r", positionName(victim.Position)))
	d.WriteToBuffer(fmt.Sprintf("V) Attacks     : %s\n\r", victim.Attacks.String()))
	d.WriteToBuffer(fmt.Sprintf("W) Defenses    : %s\n\r", victim.Defenses.String()))
	d.WriteToBuffer(fmt.Sprintf("X) Body Parts  : 0x%x\n\r", victim.XFlags))
	d.WriteToBuffer(fmt.Sprintf("Y) Act Flags   : %s\n\r", victim.Act.String()))
	d.WriteToBuffer(fmt.Sprintf("Z) Affected    : %s\n\r", victim.AffectedBy.String()))
	d.WriteToBuffer("Q) Quit\n\r")
	d.WriteToBuffer("Enter choice : ")

	d.Olc.Mode = types.MEDIT_NPC_MAIN_MENU
}

// meditDispPcMenu renders the PC main menu. Ports C
// src/omedit.c:887-945 verbatim in label content; cosmetic tags dropped.
// Reads PC-specific fields from victim.PCData (conditions + favor).
//
// All label strings are load-bearing for test assertions (A6 regex scan).
func meditDispPcMenu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil || d.Character == nil {
		return
	}
	victim, ok := d.Olc.Target.(*types.CharData)
	if !ok || victim == nil {
		return
	}
	ch := d.Character

	d.WriteToBuffer(fmt.Sprintf("1) Sex: %-7s           2) Name: %s\n\r",
		sexLabel(victim.Sex), victim.Name))
	desc := victim.Description
	if desc == "" {
		desc = "(none set)"
	}
	d.WriteToBuffer(fmt.Sprintf("3) Description:-\n\r%s\n\r", desc))

	d.WriteToBuffer(fmt.Sprintf("4) Class: [%d],  5) Race:   [%d]\n\r",
		victim.Class, victim.Race))
	d.WriteToBuffer(fmt.Sprintf(
		"6) Level:       [%5d],  7) Alignment:    [%5d],  8) Strength:  [%5d]\n\r",
		victim.Level, victim.Alignment, victim.GetCurrStr()))
	d.WriteToBuffer(fmt.Sprintf(
		"9) Intelligence:[%5d],  A) Wisdom:       [%5d],  B) Dexterity: [%5d]\n\r",
		victim.GetCurrInt(), victim.GetCurrWis(), victim.GetCurrDex()))
	d.WriteToBuffer(fmt.Sprintf(
		"C) Constitution:[%5d],  D) Charisma:     [%5d],  E) Luck:      [%5d]\n\r",
		victim.GetCurrCon(), victim.GetCurrCha(), victim.GetCurrLck()))
	d.WriteToBuffer(fmt.Sprintf(
		"F) Hps:   [%5d/%5d],  G) Mana:   [%5d/%5d],  H) Move:[%5d/%-5d]\n\r",
		victim.Hit, victim.MaxHit, victim.Mana, victim.MaxMana, victim.Move, victim.MaxMove))
	d.WriteToBuffer(fmt.Sprintf(
		"I) Gold:  [%11d],  J) Mentalstate:  [%5d],  K) Emotional: [%5d]\n\r",
		victim.Gold, victim.MentalState, victim.EmotionalState))

	thirst, full, drunk, favor := 0, 0, 0, 0
	if victim.PCData != nil {
		// C uses COND_THIRST=0, COND_FULL=1, COND_DRUNK=2.
		if len(victim.PCData.Condition) > 0 {
			thirst = victim.PCData.Condition[0]
		}
		if len(victim.PCData.Condition) > 1 {
			full = victim.PCData.Condition[1]
		}
		if len(victim.PCData.Condition) > 2 {
			drunk = victim.PCData.Condition[2]
		}
		favor = victim.PCData.Favor
	}
	d.WriteToBuffer(fmt.Sprintf(
		"L) Thirst:      [%5d],  M) Full:         [%5d],  N) Drunk:     [%5d]\n\r",
		thirst, full, drunk))
	d.WriteToBuffer(fmt.Sprintf("O) Favor:       [%5d]\n\r", favor))
	d.WriteToBuffer("P) Saving Throws\n\r")
	d.WriteToBuffer(fmt.Sprintf("R) Resistant   : %d\n\r", victim.Resistant))
	d.WriteToBuffer(fmt.Sprintf("S) Immune      : %d\n\r", victim.Immune))
	d.WriteToBuffer(fmt.Sprintf("T) Susceptible : %d\n\r", victim.Susceptible))
	d.WriteToBuffer(fmt.Sprintf("U) Position    : %s\n\r", positionName(victim.Position)))
	d.WriteToBuffer(fmt.Sprintf("V) Act Flags   : %s\n\r", victim.Act.String()))
	pcFlags := 0
	if victim.PCData != nil {
		pcFlags = victim.PCData.Flags
	}
	d.WriteToBuffer(fmt.Sprintf("W) PC Flags    : 0x%x\n\r", pcFlags))
	d.WriteToBuffer(fmt.Sprintf("X) Affected    : %s\n\r", victim.AffectedBy.String()))

	deityName := "None"
	if victim.PCData != nil && victim.PCData.Deity != nil {
		deityName = victim.PCData.Deity.Name
	}
	d.WriteToBuffer(fmt.Sprintf("Y) Deity       : %s\n\r", deityName))

	// Trust-gated rows per C :927-937.
	if ch.GetTrust() >= types.LEVEL_GOD {
		clanLabel := "Clan"
		clanName := "None"
		if victim.PCData != nil && victim.PCData.Clan != nil {
			clanName = victim.PCData.Clan.Name
		}
		d.WriteToBuffer(fmt.Sprintf("Z) %-12s: %s\n\r", clanLabel, clanName))
	}
	if ch.GetTrust() >= types.LEVEL_SUB_IMPLEM {
		cName := "None"
		if victim.PCData != nil && victim.PCData.Council != nil {
			cName = victim.PCData.Council.Name
		}
		d.WriteToBuffer(fmt.Sprintf("=) Council     : %s\n\r", cName))
	}

	d.WriteToBuffer("Q) Quit\n\r")
	d.WriteToBuffer("Enter choice : ")

	d.Olc.Mode = types.MEDIT_PC_MAIN_MENU
}

// --- G3 submenu renderers ---
//
// Each renderer prints:
//   1. A header identifying the field being edited.
//   2. Current state (value or flag set).
//   3. An input hint (valid values / keywords).
//
// The PARSE side of each submenu is filled in by later waves per plan
// §G7-§G13. The render side is "display-only scaffold" in Wave 2.

// meditDispSexMenu renders the sex picker.
func meditDispSexMenu(d *types.DescriptorData) {
	victim := meditVictim(d)
	current := -1
	if victim != nil {
		current = victim.Sex
	}
	d.WriteToBuffer("Sex:\n\r")
	d.WriteToBuffer("  0) Neutral\n\r")
	d.WriteToBuffer("  1) Male\n\r")
	d.WriteToBuffer("  2) Female\n\r")
	d.WriteToBuffer(fmt.Sprintf("Current: %d (%s)\n\rEnter choice: ",
		current, sexLabelOrNone(current)))
}

// sexLabelOrNone is sexLabel with an "(unset)" branch for the submenu.
func sexLabelOrNone(sex int) string {
	if sex < 0 {
		return "unset"
	}
	return sexLabel(sex)
}

// meditDispPosMenu renders the position picker.
func meditDispPosMenu(d *types.DescriptorData) {
	victim := meditVictim(d)
	cur := -1
	if victim != nil {
		cur = victim.Position
	}
	d.WriteToBuffer("Position:\n\r")
	for i, name := range []string{
		"dead", "mortally wounded", "incapacitated", "stunned",
		"sleeping", "berserk", "resting", "aggressive",
		"sitting", "fighting", "defensive", "evasive",
		"standing", "mounted", "shove", "drag",
	} {
		d.WriteToBuffer(fmt.Sprintf("  %2d) %s\n\r", i, name))
	}
	d.WriteToBuffer(fmt.Sprintf("Current: %s\n\rEnter choice: ", positionName(cur)))
}

// meditDispDefaultPosMenu renders the default-position picker (NPC-only).
func meditDispDefaultPosMenu(d *types.DescriptorData) {
	victim := meditVictim(d)
	cur := -1
	if victim != nil {
		cur = victim.DefPosition
	}
	d.WriteToBuffer("Default Position:\n\r")
	d.WriteToBuffer(fmt.Sprintf("Current: %s\n\r", positionName(cur)))
	d.WriteToBuffer("Enter position number: ")
}

// meditDispAttackMenu renders the NPC attack-flag picker. Full flag
// tables land with G9's bitmask editor; Wave 2 prints current mask only.
func meditDispAttackMenu(d *types.DescriptorData) {
	victim := meditVictim(d)
	attacks := ""
	if victim != nil {
		attacks = victim.Attacks.String()
	}
	d.WriteToBuffer("Attacks:\n\r")
	d.WriteToBuffer(fmt.Sprintf("Current: %s\n\r", attacks))
	d.WriteToBuffer("Enter attack-flag toggle (done to exit): ")
}

// meditDispDefenseMenu renders the NPC defense-flag picker.
func meditDispDefenseMenu(d *types.DescriptorData) {
	victim := meditVictim(d)
	defenses := ""
	if victim != nil {
		defenses = victim.Defenses.String()
	}
	d.WriteToBuffer("Defenses:\n\r")
	d.WriteToBuffer(fmt.Sprintf("Current: %s\n\r", defenses))
	d.WriteToBuffer("Enter defense-flag toggle (done to exit): ")
}

// meditDispSpecMenu renders the NPC spec_fun name prompt.
func meditDispSpecMenu(d *types.DescriptorData) {
	victim := meditVictim(d)
	cur := ""
	if victim != nil {
		cur = victim.SpecFun
	}
	d.WriteToBuffer("Spec Fun:\n\r")
	d.WriteToBuffer(fmt.Sprintf("Current: %s\n\r", cur))
	d.WriteToBuffer("Enter spec function name (blank to clear): ")
}

// meditDispClassMenu renders the class picker. Full class_table lookup
// lands with G12 — Wave 2 prints current class index + numeric hint.
func meditDispClassMenu(d *types.DescriptorData) {
	victim := meditVictim(d)
	cur := -1
	if victim != nil {
		cur = victim.Class
	}
	d.WriteToBuffer("Class:\n\r")
	d.WriteToBuffer(fmt.Sprintf("Current: %d\n\r", cur))
	d.WriteToBuffer("Enter class number: ")
}

// meditDispRaceMenu renders the race picker.
func meditDispRaceMenu(d *types.DescriptorData) {
	victim := meditVictim(d)
	cur := -1
	if victim != nil {
		cur = victim.Race
	}
	d.WriteToBuffer("Race:\n\r")
	d.WriteToBuffer(fmt.Sprintf("Current: %d\n\r", cur))
	d.WriteToBuffer("Enter race number: ")
}

// meditDispSaveMenu renders the 5-way save-throw sub-dispatcher. Full
// editing lands in G11; Wave 2 prints the menu.
func meditDispSaveMenu(d *types.DescriptorData) {
	victim := meditVictim(d)
	var p, w, pp, br, ss int
	if victim != nil {
		p = victim.SavingPoisonDeath
		w = victim.SavingWand
		pp = victim.SavingParaPetri
		br = victim.SavingBreath
		ss = victim.SavingSpellStaff
	}
	d.WriteToBuffer("Saving Throws:\n\r")
	d.WriteToBuffer(fmt.Sprintf("  1) Poison/Death     [%d]\n\r", p))
	d.WriteToBuffer(fmt.Sprintf("  2) Wand             [%d]\n\r", w))
	d.WriteToBuffer(fmt.Sprintf("  3) Paralysis/Petri  [%d]\n\r", pp))
	d.WriteToBuffer(fmt.Sprintf("  4) Breath           [%d]\n\r", br))
	d.WriteToBuffer(fmt.Sprintf("  5) Spell/Staff      [%d]\n\r", ss))
	d.WriteToBuffer("  Q) Quit\n\rEnter choice: ")
}

// meditDispAffectMenu renders the add-affect-menu entry. Full editor
// lands in G10.
func meditDispAffectMenu(d *types.DescriptorData) {
	d.WriteToBuffer("Affect editor:\n\r")
	d.WriteToBuffer("  A) Add affect\n\r  R) Remove affect\n\r  Q) Quit\n\r")
	d.WriteToBuffer("Enter choice: ")
}

// renderBitmaskTable renders a flag-name picker for a bitmask editor:
// lists each flag (1-based) with current-set markers, plus the helper's
// input hint. Shared by all six G9 flag submenus (NPC/PC/AFF/PCDATA/
// PARTS/RIS). Back-wired 2026-04-21 (plan §G9); the Wave-2 scaffolds
// only rendered mask hex strings because the name tables had not yet
// landed.
func renderBitmaskTable(d *types.DescriptorData, header string, names []string, isSet func(i int) bool) {
	d.WriteToBuffer(header + "\n\r")
	cols := 3
	for i, name := range names {
		if name == "" {
			continue
		}
		marker := " "
		if isSet(i) {
			marker = "*"
		}
		d.WriteToBuffer(fmt.Sprintf(" [%s] %2d) %-14s", marker, i+1, name))
		if (i+1)%cols == 0 {
			d.WriteToBuffer("\n\r")
		}
	}
	if len(names)%cols != 0 {
		d.WriteToBuffer("\n\r")
	}
	d.WriteToBuffer("Enter flag name or 1-based index (done to exit): ")
}

// meditDispNpcFlagsMenu renders the ACT_* bitmask editor.
func meditDispNpcFlagsMenu(d *types.DescriptorData) {
	victim := meditVictim(d)
	isSet := func(i int) bool {
		return victim != nil && victim.Act.IsSet(i)
	}
	d.WriteToBuffer("Act Flags (NPC) — '*' = set:\n\r")
	renderBitmaskTable(d, "", util.ActflagNames, isSet)
}

// meditDispPcFlagsMenu renders the PLR_* bitmask editor.
func meditDispPcFlagsMenu(d *types.DescriptorData) {
	victim := meditVictim(d)
	isSet := func(i int) bool {
		return victim != nil && victim.Act.IsSet(i)
	}
	d.WriteToBuffer("PC Flags (PLR_*) — '*' = set:\n\r")
	renderBitmaskTable(d, "", util.PlrflagNames, isSet)
}

// meditDispAffFlagsMenu renders the AFF_* bitmask editor.
func meditDispAffFlagsMenu(d *types.DescriptorData) {
	victim := meditVictim(d)
	isSet := func(i int) bool {
		return victim != nil && victim.AffectedBy.IsSet(i)
	}
	d.WriteToBuffer("Affect Flags (AFF_*) — '*' = set:\n\r")
	renderBitmaskTable(d, "", util.AffflagNames, isSet)
}

// meditDispPcdataFlagsMenu renders the PCFLAG_* bitmask editor.
func meditDispPcdataFlagsMenu(d *types.DescriptorData) {
	victim := meditVictim(d)
	mask := 0
	if victim != nil && victim.PCData != nil {
		mask = victim.PCData.Flags
	}
	isSet := func(i int) bool { return mask&(1<<i) != 0 }
	d.WriteToBuffer("PCData Flags — '*' = set:\n\r")
	renderBitmaskTable(d, "", util.PcflagNames, isSet)
}

// meditDispPartsMenu renders the PART_* bitmask editor.
func meditDispPartsMenu(d *types.DescriptorData) {
	victim := meditVictim(d)
	mask := 0
	if victim != nil {
		mask = victim.XFlags
	}
	isSet := func(i int) bool { return mask&(1<<i) != 0 }
	d.WriteToBuffer("Body Parts — '*' = set:\n\r")
	renderBitmaskTable(d, "", util.PartflagNames, isSet)
}

// meditDispRisMenu renders the RIS_* bitmask editor. RIS shares one
// table across three fields (Resistant/Immune/Susceptible); this
// renderer shows all three current masks then the shared flag table.
// Back-wired 2026-04-21 (plan §G9).
func meditDispRisMenu(d *types.DescriptorData) {
	victim := meditVictim(d)
	var r, im, su int
	if victim != nil {
		r = victim.Resistant
		im = victim.Immune
		su = victim.Susceptible
	}
	d.WriteToBuffer("Resistant / Immune / Susceptible (RIS_*):\n\r")
	d.WriteToBuffer(fmt.Sprintf("  Resistant   : %s\n\r", util.FlagString(r, util.RisflagNames)))
	d.WriteToBuffer(fmt.Sprintf("  Immune      : %s\n\r", util.FlagString(im, util.RisflagNames)))
	d.WriteToBuffer(fmt.Sprintf("  Susceptible : %s\n\r", util.FlagString(su, util.RisflagNames)))
	// Pick the current-mode's target mask for the '*' markers.
	mode := d.Olc.Mode
	var focus int
	switch mode {
	case types.MEDIT_IMMUNE:
		focus = im
	case types.MEDIT_SUSCEPTIBLE:
		focus = su
	default:
		focus = r
	}
	isSet := func(i int) bool { return focus&(1<<i) != 0 }
	renderBitmaskTable(d, "", util.RisflagNames, isSet)
}

// meditVictim is a small helper that does the repeated type-assert dance.
// Returns nil if the descriptor or target is absent / wrong type.
func meditVictim(d *types.DescriptorData) *types.CharData {
	if d == nil || d.Olc == nil {
		return nil
	}
	v, ok := d.Olc.Target.(*types.CharData)
	if !ok {
		return nil
	}
	return v
}
