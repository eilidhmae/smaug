// Package game — oedit per-item-type value dispatchers.
//
// Ports src/ooedit.c oedit_disp_val{1..6}_menu at :623-849 and the
// corresponding OEDIT_VALUE_{1..6} parse arms at :1535-1733.
//
// Each "disp" helper branches on idx.ItemType and either emits a prompt or
// dispatches into a sub-menu (lever-flags, container-flags, liquid-type,
// weapon-type, spells-name prompt). Each "parse" arm clamps per-type and
// advances to the next value slot, or re-displays the current slot's
// sub-menu when input is a flag-toggle.
//
// Go-port divergences (vs C at ooedit.c):
//   - C mirror-writes to pIndexData->value[] when IS_OBJ_STAT(obj,
//     ITEM_PROTOTYPE); Go operates on *ObjIndexData directly (see plan
//     §Go Design). No mirror-write needed.
//   - C's skill_lookup(arg) resolves spell names to sn. Go has
//     `combat.lookupSkillSlot` but it's in a sibling package and not
//     reachable from `game/` without a seam. Per plan §Q5 Option 2:
//     numeric-only spell input is accepted; spell-name word input is
//     rejected with a reprompt. Follow-up tracked in TODO.md.
//   - Container-flags and lever-flags label tables live here.
//   - ITEM_GOLD / ITEM_SILVER / ITEM_COPPER (GSC variant) — scope-cut
//     per plan; ITEM_MONEY branch preserved (stock C).
//
// Plan: plan-phase6-olc-oedit.md §G7.
package game

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
)

// --- Lever / trigger flag table ---

// trigFlagNames is the 29-entry trigger-flag table used by ITEM_LEVER/
// ITEM_SWITCH menus. Mirrors C `trig_flags[]` indexed 0..28. Index maps
// to bit position in idx.Value[0] for lever/switch triggers.
var trigFlagNames = []string{
	"up",           // 0
	"unlock",       // 1
	"lock",         // 2
	"d_north",      // 3
	"d_south",      // 4
	"d_east",       // 5
	"d_west",       // 6
	"d_up",         // 7
	"d_down",       // 8
	"door",         // 9
	"container",    // 10
	"open",         // 11
	"close",        // 12
	"passage",      // 13
	"oload",        // 14
	"mload",        // 15
	"teleport",     // 16
	"teleportall",  // 17
	"teleportplus", // 18
	"death",        // 19
	"cast",         // 20
	"fakeblade",    // 21
	"rand4",        // 22
	"rand6",        // 23
	"trapdoor",     // 24
	"anotheroom",   // 25
	"usedial",      // 26
	"absolutevnum", // 27
	"showroomdesc", // 28
}

// --- Container flag table ---

// containerFlagNames is the 5-entry container-flag table. Mirrors C
// `container_flags[]`. Bits land in idx.Value[1].
var containerFlagNames = []string{
	"closeable", // 0 = CONT_CLOSEABLE
	"pickproof", // 1 = CONT_PICKPROOF
	"closed",    // 2 = CONT_CLOSED
	"locked",    // 3 = CONT_LOCKED
	"eatkey",    // 4 = CONT_EATKEY
}

// --- Liquid table (for DRINK_CON / FOUNTAIN) ---

// liquidNames is the LIQ_MAX-entry label table. Mirrors C `liq_table[].liq_name`.
// A leaner version focused on the menu display; the full liq_table in C
// has color + proof + sugar fields that aren't needed here.
var liquidNames = []string{
	"water",
	"beer",
	"red wine",
	"ale",
	"dark ale",
	"whisky",
	"lemonade",
	"firebreather",
	"local specialty",
	"slime mold juice",
	"milk",
	"tea",
	"coffee",
	"blood",
	"salt water",
	"cola",
	"wine",
	"fine wine",
}

// --- Weapon-type menu ---

// weaponTypeNames mirrors a subset of C attack_table for the weapon-type
// prompt. Index = damtype value written to idx.Value[3] when ItemType ==
// ITEM_WEAPON. Names match attackTable in internal/combat/dammessage.go.
var weaponTypeNames = []string{
	"hit",      // 0
	"slice",    // 1
	"stab",     // 2
	"slash",    // 3
	"whip",     // 4
	"claw",     // 5
	"blast",    // 6
	"pound",    // 7
	"crush",    // 8
	"grep",     // 9
	"bite",     // 10
	"pierce",   // 11
	"suction",  // 12
	"chop",     // 13
	"sting",    // 14
	"smash",    // 15
	"shock",    // 16
	"scratch",  // 17
	"peck-pi",  // 18 (peck pierce)
	"peck-bl",  // 19 (peck blunt)
	"chop-pi",  // 20
	"sweep",    // 21
	"wavepush", // 22
	"flame",    // 23
}

// --- Flag-list helper ---

// flagListString returns a space-separated string of labels where the
// corresponding bit in v is set. Mirrors C `flag_string(v, names)`.
func flagListString(v int, names []string) string {
	var out []string
	for i, n := range names {
		if v&(1<<i) != 0 {
			out = append(out, n)
		}
	}
	if len(out) == 0 {
		return "(none)"
	}
	return strings.Join(out, " ")
}

// --- Disp helpers (per-slot per-item-type) ---

// oeditDispContainerFlagsMenu renders the 5-option container-flags menu.
// Mirrors C :407-424. Sets Mode = OEDIT_VALUE_2.
func oeditDispContainerFlagsMenu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	idx, ok := d.Olc.Target.(*types.ObjIndexData)
	if !ok || idx == nil {
		return
	}
	var sb strings.Builder
	for i, n := range containerFlagNames {
		fmt.Fprintf(&sb, "&g%d&w) %s\n\r", i+1, n)
	}
	fmt.Fprintf(&sb, "Container flags: &c%s&w\n\rEnter flag, 0 to quit : ",
		flagListString(idx.Value[1], containerFlagNames))
	d.WriteToBuffer(sb.String())
	d.Olc.Mode = types.OEDIT_VALUE_2
}

// oeditDispLeverFlagsMenu renders the 29-option trigger-flags menu.
// Mirrors C :429-444. Sets Mode = OEDIT_VALUE_1.
func oeditDispLeverFlagsMenu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	idx, ok := d.Olc.Target.(*types.ObjIndexData)
	if !ok || idx == nil {
		return
	}
	var sb strings.Builder
	for i, n := range trigFlagNames {
		fmt.Fprintf(&sb, "&g%2d&w) %s\n\r", i+1, n)
	}
	fmt.Fprintf(&sb, "Lever flags: &c%s&w\n\rEnter flag, 0 to quit: ",
		flagListString(idx.Value[0], trigFlagNames))
	d.WriteToBuffer(sb.String())
	d.Olc.Mode = types.OEDIT_VALUE_1
}

// oeditDispLiquidMenu renders the LIQ_MAX-entry liquid-type list in 3
// columns. Mirrors C oedit_liquid_type at :547-564. Sets Mode = OEDIT_VALUE_3.
func oeditDispLiquidMenu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	var sb strings.Builder
	col := 0
	for i, n := range liquidNames {
		fmt.Fprintf(&sb, " &w%2d&g) &c%-20.20s ", i, n)
		col++
		if col%3 == 0 {
			sb.WriteString("\n\r")
		}
	}
	if col%3 != 0 {
		sb.WriteString("\n\r")
	}
	sb.WriteString("\n\rEnter liquid type : ")
	d.WriteToBuffer(sb.String())
	d.Olc.Mode = types.OEDIT_VALUE_3
}

// oeditDispWeaponMenu renders the 24-entry attack-type menu in 2 columns.
// Mirrors C oedit_disp_weapon_menu at :597-614. Sets Mode = OEDIT_VALUE_4.
func oeditDispWeaponMenu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	var sb strings.Builder
	col := 0
	for i, n := range weaponTypeNames {
		fmt.Fprintf(&sb, "&g%2d&w) %-15.15s ", i, n)
		col++
		if col%2 == 0 {
			sb.WriteString("\n\r")
		}
	}
	if col%2 != 0 {
		sb.WriteString("\n\r")
	}
	sb.WriteString("\n\rEnter weapon type : ")
	d.WriteToBuffer(sb.String())
	d.Olc.Mode = types.OEDIT_VALUE_4
}

// oeditDispSpellsMenu emits the one-line spell-name prompt. Mirrors C
// oedit_disp_spells_menu at :617-620. The mode is set by the caller.
func oeditDispSpellsMenu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	d.WriteToBuffer("Enter the name of the spell (or numeric sn, -1 for none): ")
}

// --- Per-slot menu dispatchers ---

// oeditDispVal1Menu dispatches on idx.ItemType for Value[0]. Mirrors C
// oedit_disp_val1_menu at :623-688.
func oeditDispVal1Menu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	idx, ok := d.Olc.Target.(*types.ObjIndexData)
	if !ok || idx == nil {
		return
	}
	d.Olc.Mode = types.OEDIT_VALUE_1
	switch idx.ItemType {
	case types.ITEM_LIGHT:
		// values 0 and 1 unused; jump to val3 per C :632.
		oeditDispVal3Menu(d)
	case types.ITEM_SALVE, types.ITEM_PILL, types.ITEM_SCROLL,
		types.ITEM_WAND, types.ITEM_STAFF, types.ITEM_POTION:
		d.WriteToBuffer("Spell level : ")
	case types.ITEM_MISSILE_WEAPON, types.ITEM_WEAPON:
		d.WriteToBuffer("Condition : ")
	case types.ITEM_ARMOR:
		d.WriteToBuffer("Current AC : ")
	case types.ITEM_PIPE, types.ITEM_CONTAINER, types.ITEM_DRINK_CON,
		types.ITEM_FOUNTAIN:
		d.WriteToBuffer("Capacity : ")
	case types.ITEM_FOOD:
		d.WriteToBuffer("Hours to fill stomach : ")
	case types.ITEM_MONEY:
		d.WriteToBuffer("Amount of Gold coins : ")
	case types.ITEM_HERB:
		// Value 0 unused, skip to val2.
		oeditDispVal2Menu(d)
	case types.ITEM_LEVER, types.ITEM_SWITCH:
		oeditDispLeverFlagsMenu(d)
	case types.ITEM_TRAP:
		d.WriteToBuffer("Charges: ")
	default:
		// No value-0 surface for this item-type; bail back to main.
		OeditDispMenu(d)
	}
}

// oeditDispVal2Menu dispatches on idx.ItemType for Value[1]. Mirrors C
// oedit_disp_val2_menu at :691-749.
func oeditDispVal2Menu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	idx, ok := d.Olc.Target.(*types.ObjIndexData)
	if !ok || idx == nil {
		return
	}
	d.Olc.Mode = types.OEDIT_VALUE_2
	switch idx.ItemType {
	case types.ITEM_PILL, types.ITEM_SCROLL, types.ITEM_POTION:
		oeditDispSpellsMenu(d)
	case types.ITEM_SALVE, types.ITEM_HERB:
		d.WriteToBuffer("Charges: ")
	case types.ITEM_PIPE:
		d.WriteToBuffer("Number of draws: ")
	case types.ITEM_WAND, types.ITEM_STAFF:
		d.WriteToBuffer("Max number of charges : ")
	case types.ITEM_WEAPON:
		d.WriteToBuffer("Number of damage dice : ")
	case types.ITEM_FOOD:
		d.WriteToBuffer("Condition: ")
	case types.ITEM_CONTAINER:
		oeditDispContainerFlagsMenu(d)
	case types.ITEM_DRINK_CON, types.ITEM_FOUNTAIN:
		d.WriteToBuffer("Quantity : ")
	case types.ITEM_ARMOR:
		d.WriteToBuffer("Original AC: ")
	case types.ITEM_LEVER, types.ITEM_SWITCH:
		// TRIG_CAST bit 20 → if set, spell name; else Vnum. Per C :740-743.
		if idx.Value[0]&(1<<20) != 0 {
			oeditDispSpellsMenu(d)
		} else {
			d.WriteToBuffer("Vnum: ")
		}
	default:
		OeditDispMenu(d)
	}
}

// oeditDispVal3Menu dispatches on idx.ItemType for Value[2]. Mirrors C
// oedit_disp_val3_menu at :752-784.
func oeditDispVal3Menu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	idx, ok := d.Olc.Target.(*types.ObjIndexData)
	if !ok || idx == nil {
		return
	}
	d.Olc.Mode = types.OEDIT_VALUE_3
	switch idx.ItemType {
	case types.ITEM_LIGHT:
		d.WriteToBuffer("Number of hours (0 = burnt, -1 is infinite) : ")
	case types.ITEM_PILL, types.ITEM_SCROLL, types.ITEM_POTION:
		oeditDispSpellsMenu(d)
	case types.ITEM_WAND, types.ITEM_STAFF:
		d.WriteToBuffer("Number of charges remaining : ")
	case types.ITEM_WEAPON:
		d.WriteToBuffer("Size of damage dice : ")
	case types.ITEM_CONTAINER:
		d.WriteToBuffer("Vnum of key to open container (-1 for no key) : ")
	case types.ITEM_DRINK_CON, types.ITEM_FOUNTAIN:
		oeditDispLiquidMenu(d)
	default:
		OeditDispMenu(d)
	}
}

// oeditDispVal4Menu dispatches on idx.ItemType for Value[3]. Mirrors C
// oedit_disp_val4_menu at :787-811.
func oeditDispVal4Menu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	idx, ok := d.Olc.Target.(*types.ObjIndexData)
	if !ok || idx == nil {
		return
	}
	d.Olc.Mode = types.OEDIT_VALUE_4
	switch idx.ItemType {
	case types.ITEM_SCROLL, types.ITEM_POTION, types.ITEM_WAND, types.ITEM_STAFF:
		oeditDispSpellsMenu(d)
	case types.ITEM_WEAPON:
		oeditDispWeaponMenu(d)
	case types.ITEM_DRINK_CON, types.ITEM_FOUNTAIN, types.ITEM_FOOD:
		d.WriteToBuffer("Poisoned (0 = not poisoned) : ")
	default:
		OeditDispMenu(d)
	}
}

// oeditDispVal5Menu dispatches on idx.ItemType for Value[4]. Mirrors C
// oedit_disp_val5_menu at :814-833.
func oeditDispVal5Menu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	idx, ok := d.Olc.Target.(*types.ObjIndexData)
	if !ok || idx == nil {
		return
	}
	d.Olc.Mode = types.OEDIT_VALUE_5
	switch idx.ItemType {
	case types.ITEM_SALVE:
		oeditDispSpellsMenu(d)
	case types.ITEM_FOOD:
		d.WriteToBuffer("Food value: ")
	case types.ITEM_MISSILE_WEAPON:
		d.WriteToBuffer("Range: ")
	default:
		OeditDispMenu(d)
	}
}

// oeditDispVal6Menu dispatches on idx.ItemType for Value[5]. Mirrors C
// oedit_disp_val6_menu at :836-849.
func oeditDispVal6Menu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	idx, ok := d.Olc.Target.(*types.ObjIndexData)
	if !ok || idx == nil {
		return
	}
	d.Olc.Mode = types.OEDIT_VALUE_6
	switch idx.ItemType {
	case types.ITEM_SALVE:
		oeditDispSpellsMenu(d)
	default:
		OeditDispMenu(d)
	}
}

// --- Parse arms ---

// parseIntArg parses arg as a decimal integer. Returns (n, true) on
// success, (0, false) on failure. Tiny helper so each value-parse arm
// stays compact.
func parseIntArg(arg string) (int, bool) {
	if v, err := strconv.Atoi(strings.TrimSpace(arg)); err == nil {
		return v, true
	}
	return 0, false
}

// oeditHandleValue1 ports C OEDIT_VALUE_1 at :1535-1564. Dispatches on
// idx.ItemType. For LEVER/SWITCH, numeric input toggles trig-flag bits in
// Value[0] (n=0 → advance to val2; 1..29 → toggle bit n-1). Default:
// integer set of Value[0] + advance to val2.
func oeditHandleValue1(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	n, _ := parseIntArg(arg)
	switch idx.ItemType {
	case types.ITEM_LEVER, types.ITEM_SWITCH:
		if n < 0 || n > 29 {
			oeditDispLeverFlagsMenu(d)
			return
		}
		if n == 0 {
			oeditDispVal2Menu(d)
			return
		}
		idx.Value[0] ^= 1 << (n - 1)
		olcLog(d, "OBJ", "Changed v0 to %d", idx.Value[0])
		oeditDispVal1Menu(d)
	default:
		idx.Value[0] = n
		olcLog(d, "OBJ", "Changed v0 to %d", idx.Value[0])
		oeditDispVal2Menu(d)
	}
}

// oeditHandleValue2 ports C OEDIT_VALUE_2 at :1566-1616. Branches per
// item_type. Per plan §Q5 Option 2: spell-name word input is rejected
// (numeric sn only). Container numeric input toggles bit in Value[1].
func oeditHandleValue2(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	n, isNum := parseIntArg(arg)
	switch idx.ItemType {
	case types.ITEM_PILL, types.ITEM_SCROLL, types.ITEM_POTION:
		if !isNum {
			d.WriteToBuffer("Spell name lookup not yet wired; enter numeric sn: ")
			return
		}
		idx.Value[1] = n
		olcLog(d, "OBJ", "Changed v1 to %d", idx.Value[1])
		oeditDispVal3Menu(d)
	case types.ITEM_LEVER, types.ITEM_SWITCH:
		// TRIG_CAST bit 20 → spell-name (numeric-only port), else plain int.
		if idx.Value[0]&(1<<20) != 0 && !isNum {
			d.WriteToBuffer("Spell name lookup not yet wired; enter numeric sn: ")
			return
		}
		idx.Value[1] = n
		olcLog(d, "OBJ", "Changed v1 to %d", idx.Value[1])
		oeditDispVal3Menu(d)
	case types.ITEM_CONTAINER:
		if n < 0 || n > 61 { // MAX_OLC_ITEMS_LIST per C src/olc.h:48
			oeditDispContainerFlagsMenu(d)
			return
		}
		if n == 0 {
			oeditDispVal3Menu(d)
			return
		}
		idx.Value[1] ^= 1 << (n - 1)
		olcLog(d, "OBJ", "Changed v1 to %d", idx.Value[1])
		oeditDispVal2Menu(d)
	default:
		idx.Value[1] = n
		olcLog(d, "OBJ", "Changed v1 to %d", idx.Value[1])
		oeditDispVal3Menu(d)
	}
}

// oeditHandleValue3 ports C OEDIT_VALUE_3 at :1618-1651. Applies URANGE
// per-item-type clamps. Scroll/Potion/Pill take numeric sn (spell-name
// word rejected per Q5). Advances to val4.
func oeditHandleValue3(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	n, isNum := parseIntArg(arg)
	var lo, hi int
	switch idx.ItemType {
	case types.ITEM_SCROLL, types.ITEM_POTION, types.ITEM_PILL:
		if !isNum {
			d.WriteToBuffer("Spell name lookup not yet wired; enter numeric sn: ")
			return
		}
		lo, hi = -1, 32000
	case types.ITEM_WEAPON:
		lo, hi = 0, 100
	case types.ITEM_DRINK_CON, types.ITEM_FOUNTAIN:
		lo, hi = 0, types.LIQ_MAX
	default:
		lo, hi = -32000, 32000
	}
	idx.Value[2] = uRange(lo, n, hi)
	olcLog(d, "OBJ", "Changed v2 to %d", idx.Value[2])
	oeditDispVal4Menu(d)
}

// oeditHandleValue4 ports C OEDIT_VALUE_4 at :1653-1686. Weapon attack-
// type range-REJECTS (no clamp — re-prompts on out-of-range per C :1670).
// Scroll/Potion/Wand/Staff accept numeric sn only.
func oeditHandleValue4(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	n, isNum := parseIntArg(arg)
	var lo, hi int
	switch idx.ItemType {
	case types.ITEM_PILL, types.ITEM_SCROLL, types.ITEM_POTION,
		types.ITEM_WAND, types.ITEM_STAFF:
		if !isNum {
			d.WriteToBuffer("Spell name lookup not yet wired; enter numeric sn: ")
			return
		}
		lo, hi = -1, 32000
	case types.ITEM_WEAPON:
		lo, hi = 0, types.MAX_ATTACK_TYPE-1
		if n < lo || n > hi {
			oeditDispVal4Menu(d)
			return
		}
	default:
		lo, hi = -32000, 32000
	}
	idx.Value[3] = uRange(lo, n, hi)
	olcLog(d, "OBJ", "Changed v3 to %d", idx.Value[3])
	oeditDispVal5Menu(d)
}

// oeditHandleValue5 ports C OEDIT_VALUE_5 at :1688-1712. Salve takes
// numeric sn; Food clamped [0, 32000]; others [-32000, 32000]. Advances
// to val6.
func oeditHandleValue5(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	n, isNum := parseIntArg(arg)
	var lo, hi int
	switch idx.ItemType {
	case types.ITEM_SALVE:
		if !isNum {
			d.WriteToBuffer("Spell name lookup not yet wired; enter numeric sn: ")
			return
		}
		lo, hi = -1, 32000
	case types.ITEM_FOOD:
		lo, hi = 0, 32000
	default:
		lo, hi = -32000, 32000
	}
	idx.Value[4] = uRange(lo, n, hi)
	olcLog(d, "OBJ", "Changed v4 to %d", idx.Value[4])
	oeditDispVal6Menu(d)
}

// oeditHandleValue6 ports C OEDIT_VALUE_6 at :1714-1733. Salve takes
// numeric sn; others [-32000, 32000]. Returns to main menu.
func oeditHandleValue6(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	n, isNum := parseIntArg(arg)
	var lo, hi int
	switch idx.ItemType {
	case types.ITEM_SALVE:
		if !isNum {
			d.WriteToBuffer("Spell name lookup not yet wired; enter numeric sn: ")
			return
		}
		lo, hi = -1, 32000
	default:
		lo, hi = -32000, 32000
	}
	idx.Value[5] = uRange(lo, n, hi)
	olcLog(d, "OBJ", "Changed v5 to %d", idx.Value[5])
	OeditDispMenu(d)
}
