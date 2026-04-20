// Package game — oedit menu rendering helpers.
//
// Ports src/ooedit.c menu-display functions (oedit_disp_*). Each function
// writes color-tagged menu text to the descriptor's output buffer; the
// descriptor's ColorFunc translates the tags on flush. Line endings use
// Go-convention "\n\r" to match the rest of the port and the redit
// precedent.
//
// The ANSI screen-clear sequence C emits at the top of each menu
// ("50\x1B[;H\x1B[2J") is deliberately omitted per redit Q1 precedent —
// cosmetic only and breaks testclient determinism.
//
// Wave 2 ships: main menu, type menu, extras menu, wear menu, layer menu,
// extradesc menu, plus per-item-type value-menu STUBS (G7 fills in Wave 3).
//
// Plan: plan-phase6-olc-oedit.md §G4.
package game

import (
	"fmt"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
)

// oTypeNames is the human-readable item-type label table. Index = ITEM_*
// constant value (iota starting at 0 = ITEM_NONE). Mirrors the C
// `o_types[]` table referenced by oedit_disp_type_menu at
// src/ooedit.c:852-868. Order pinned by `internal/types/enums.go:572-642`
// — adding entries to that iota block requires extending this table.
var oTypeNames = []string{
	"none",           // 0 ITEM_NONE
	"light",          // 1 ITEM_LIGHT
	"scroll",         // 2
	"wand",           // 3
	"staff",          // 4
	"weapon",         // 5
	"fireweapon",     // 6
	"missile",        // 7
	"treasure",       // 8
	"armor",          // 9
	"potion",         // 10
	"worn",           // 11
	"furniture",      // 12
	"trash",          // 13
	"oldtrap",        // 14
	"container",      // 15
	"note",           // 16
	"drink-con",      // 17
	"key",            // 18
	"food",           // 19
	"money",          // 20
	"pen",            // 21
	"boat",           // 22
	"corpse-npc",     // 23
	"corpse-pc",      // 24
	"fountain",       // 25
	"pill",           // 26
	"blood",          // 27
	"bloodstain",     // 28
	"scraps",         // 29
	"pipe",           // 30
	"herb-con",       // 31
	"herb",           // 32
	"incense",        // 33
	"fire",           // 34
	"book",           // 35
	"switch",         // 36
	"lever",          // 37
	"pullchain",      // 38
	"button",         // 39
	"dial",           // 40
	"rune",           // 41
	"runepouch",      // 42
	"match",          // 43
	"trap",           // 44
	"map",            // 45
	"portal",         // 46
	"paper",          // 47
	"tinder",         // 48
	"lockpick",       // 49
	"spike",          // 50
	"disease",        // 51
	"oil",            // 52
	"fuel",           // 53
	"puddle",         // 54
	"abacus",         // 55
	"missile-weapon", // 56
	"projectile",     // 57
	"quiver",         // 58
	"shovel",         // 59
	"salve",          // 60
	"cook",           // 61
	"keyring",        // 62
	"odor",           // 63
	"chance",         // 64
	"piece",          // 65
	"housekey",       // 66
	"journal",        // 67
	"drink-mix",      // 68 ITEM_DRINK_MIX
}

// oTypeName returns the printable label for an ITEM_* value, or "???" for
// out-of-range. Used by OeditDispMenu for the Type field display.
func oTypeName(t int) string {
	if t < 0 || t >= len(oTypeNames) {
		return "???"
	}
	return oTypeNames[t]
}

// getOtype maps a case-insensitive item-type name to its ITEM_* constant.
// Returns -1 if not found. Mirrors C get_otype at src/build.c.
func getOtype(name string) int {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return -1
	}
	for i, n := range oTypeNames {
		if n == name {
			return i
		}
	}
	return -1
}

// oFlagNames is the extra-flag bit label table. Index = bit position in
// ObjIndexData.ExtraFlags; matches the iota in `internal/types/enums.go:649-694`.
// Mirrors C `o_flags[]` at src/build.c. Adding to that iota requires
// extending this table.
var oFlagNames = []string{
	"glow",         // 0
	"hum",          // 1
	"dark",         // 2
	"loyal",        // 3
	"evil",         // 4
	"invis",        // 5
	"magic",        // 6
	"nodrop",       // 7
	"bless",        // 8
	"antigood",     // 9
	"antievil",     // 10
	"antineutral",  // 11
	"noremove",     // 12
	"inventory",    // 13
	"antimage",     // 14
	"antithief",    // 15
	"antiwarrior",  // 16
	"anticleric",   // 17
	"organic",      // 18
	"metal",        // 19
	"donation",     // 20
	"clanobject",   // 21
	"clancorpse",   // 22
	"antivampire",  // 23
	"antidruid",    // 24
	"hidden",       // 25
	"poisoned",     // 26
	"covering",     // 27
	"deathrot",     // 28
	"buried",       // 29
	"prototype",    // 30
	"nolocate",     // 31
	"groundrot",    // 32
	"lootable",     // 33
	"permanent",    // 34
	"multi-invoke", // 35
	"deathdrop",    // 36
	"skinned",      // 37
	"nofill",       // 38
	"blackened",    // 39
	"noscavange",   // 40
	"lodged",       // 41
}

// getOflag returns the bit position for a case-insensitive extra-flag
// name, or -1 if not found. Mirrors C get_oflag at src/build.c.
func getOflag(name string) int {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return -1
	}
	for i, n := range oFlagNames {
		if n == name {
			return i
		}
	}
	return -1
}

// objExtraFlagsString renders the currently-set extra-flag labels as a
// space-separated string. Used by OeditDispMenu and oeditDispExtraMenu.
func objExtraFlagsString(bv types.BitVector) string {
	var out []string
	for i, n := range oFlagNames {
		if bv.IsSet(i) {
			out = append(out, n)
		}
	}
	return strings.Join(out, " ")
}

// wFlagNames is the wear-flag bit label table. Index = bit position in
// ObjIndexData.WearFlags. Mirrors C `w_flags[]` at src/build.c. Bit 15
// (ITEM_DUAL_WIELD) is intentionally listed for save fidelity but is
// SKIPPED by oeditDispWearMenu and rejected by the OEDIT_WEAR parse arm
// per C parity (dual-wield is a combat-set bit, not editable).
var wFlagNames = []string{
	"take",          // 0
	"finger",        // 1
	"neck",          // 2
	"body",          // 3
	"head",          // 4
	"legs",          // 5
	"feet",          // 6
	"hands",         // 7
	"arms",          // 8
	"shield",        // 9
	"about",         // 10
	"waist",         // 11
	"wrist",         // 12
	"wield",         // 13
	"hold",          // 14
	"dual_wield",    // 15 — SKIPPED by menu (C parity)
	"ears",          // 16
	"eyes",          // 17
	"missile_wield", // 18
	"back",          // 19
	"face",          // 20
	"ankle",         // 21
	"lodge_rib",     // 22
	"lodge_arm",     // 23
	"lodge_leg",     // 24
}

// getWflag returns the bit position for a case-insensitive wear-flag
// name, or -1 if not found. Mirrors C get_wflag at src/build.c. Returns
// -1 for "dual_wield" too (cannot be set via OLC).
func getWflag(name string) int {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" || name == "dual_wield" {
		return -1
	}
	for i, n := range wFlagNames {
		if n == name {
			return i
		}
	}
	return -1
}

// objWearFlagsString renders the currently-set wear-flag labels as a
// space-separated string.
func objWearFlagsString(wf int) string {
	var out []string
	for i, n := range wFlagNames {
		if wf&(1<<i) != 0 {
			out = append(out, n)
		}
	}
	return strings.Join(out, " ")
}

// layerLabels is the 9-option layer table. Mirrors C oedit_disp_layer_menu
// at src/ooedit.c:449-466. Index 1 is "Nothing" (clears Layers); 2..9 set
// progressive bits 1<<0 through 1<<7 per the C table.
var layerLabels = []string{
	"",                // 0 — unused (menu is 1-indexed)
	"Nothing",         // 1
	"Silk Shirt",      // 2 → bit 0
	"Leather Vest",    // 3 → bit 1
	"Light Chainmail", // 4 → bit 2
	"Leather Jacket",  // 5 → bit 3
	"Light Cloak",     // 6 → bit 4
	"Loose Cloak",     // 7 → bit 5
	"Cape",            // 8 → bit 6
	"Magical Effects", // 9 → bit 7
}

// OeditDispMenu renders the main object-editor menu. Mirrors C
// oedit_disp_menu at src/ooedit.c:921-985.
//
// 17 rendered options (1-9, A-G, Q). H (mprog) is parser-dispatched only
// — NOT rendered in the menu — per audit-corrected plan §G4.
//
// Sets OlcData.Mode = OEDIT_MAIN_MENU.
func OeditDispMenu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	idx, ok := d.Olc.Target.(*types.ObjIndexData)
	if !ok || idx == nil {
		return
	}

	extras := objExtraFlagsString(idx.ExtraFlags)
	if extras == "" {
		extras = "(none)"
	}
	wear := objWearFlagsString(idx.WearFlags)
	if wear == "" {
		wear = "(none)"
	}

	var sb strings.Builder
	fmt.Fprintf(&sb,
		"&w-- Object number : [&c%d&w]\n\r"+
			"&g1&w) Name        : &O%s\n\r"+
			"&g2&w) Short desc  : &O%s\n\r"+
			"&g3&w) Long desc   : &O%s\n\r"+
			"&g4&w) Action desc : &O%s\n\r"+
			"&g5&w) Type        : &c%s\n\r"+
			"&g6&w) Extra flags : &c%s\n\r"+
			"&g7&w) Wear flags  : &c%s\n\r"+
			"&g8&w) Weight      : &c%d\n\r"+
			"&g9&w) Cost        : &c%d\n\r"+
			"&gA&w) Rent        : &c%d\n\r"+
			"&gB&w) Timer       : &c%d\n\r"+
			"&gC&w) Level       : &c%d\n\r"+
			"&gD&w) Layers      : &c%d\n\r"+
			"&gE&w) Values      : &c%d %d %d %d %d %d\n\r"+
			"&gF&w) Affect menu\n\r"+
			"&gG&w) Extra descriptions menu\n\r"+
			"&gQ&w) Quit\n\r"+
			"Enter choice : ",
		d.Olc.Vnum,
		idx.Name,
		idx.ShortDescr,
		idx.Description,
		idx.ActionDesc,
		oTypeName(idx.ItemType),
		extras,
		wear,
		idx.Weight,
		idx.GoldCost,
		idx.Rent,
		0, // timer is per-instance in Go; prototype has no Timer field
		idx.Level,
		idx.Layers,
		idx.Value[0], idx.Value[1], idx.Value[2],
		idx.Value[3], idx.Value[4], idx.Value[5],
	)

	d.WriteToBuffer(sb.String())
	d.Olc.Mode = types.OEDIT_MAIN_MENU
}

// oeditDispTypeMenu renders the item-type sub-menu. 3 columns. Mirrors C
// oedit_disp_type_menu at src/ooedit.c:852-868.
//
// Sets OlcData.Mode = OEDIT_TYPE.
func oeditDispTypeMenu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	var sb strings.Builder
	col := 0
	for i := 1; i <= types.MAX_ITEM_TYPE; i++ {
		fmt.Fprintf(&sb, "&g%2d&w) %-15.15s ", i, oTypeName(i))
		col++
		if col%3 == 0 {
			sb.WriteString("\n\r")
		}
	}
	if col%3 != 0 {
		sb.WriteString("\n\r")
	}
	sb.WriteString("\n\rEnter type : ")

	d.WriteToBuffer(sb.String())
	d.Olc.Mode = types.OEDIT_TYPE
}

// oeditDispExtraMenu renders the extra-flags sub-menu in 2 columns plus
// the current set. Mirrors C oedit_disp_extra_menu at src/ooedit.c:871-890.
//
// Sets OlcData.Mode = OEDIT_EXTRAS.
func oeditDispExtraMenu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	idx, ok := d.Olc.Target.(*types.ObjIndexData)
	if !ok || idx == nil {
		return
	}
	var sb strings.Builder
	col := 0
	for i, n := range oFlagNames {
		fmt.Fprintf(&sb, "&g%2d&w) %-20.20s ", i+1, n)
		col++
		if col%2 == 0 {
			sb.WriteString("\n\r")
		}
	}
	if col%2 != 0 {
		sb.WriteString("\n\r")
	}
	current := objExtraFlagsString(idx.ExtraFlags)
	if current == "" {
		current = "(none)"
	}
	fmt.Fprintf(&sb, "\n\rExtra flags: &c%s&w\n\rEnter flags, 0 to quit : ", current)

	d.WriteToBuffer(sb.String())
	d.Olc.Mode = types.OEDIT_EXTRAS
}

// oeditDispWearMenu renders the wear-flags sub-menu. ITEM_DUAL_WIELD bit
// (15) is SKIPPED per C parity at src/ooedit.c:895-917. 2 columns.
//
// Sets OlcData.Mode = OEDIT_WEAR.
func oeditDispWearMenu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	idx, ok := d.Olc.Target.(*types.ObjIndexData)
	if !ok || idx == nil {
		return
	}
	var sb strings.Builder
	col := 0
	for i, n := range wFlagNames {
		if i == 15 { // ITEM_DUAL_WIELD — SKIPPED
			continue
		}
		fmt.Fprintf(&sb, "&g%2d&w) %-20.20s ", i+1, n)
		col++
		if col%2 == 0 {
			sb.WriteString("\n\r")
		}
	}
	if col%2 != 0 {
		sb.WriteString("\n\r")
	}
	current := objWearFlagsString(idx.WearFlags)
	if current == "" {
		current = "(none)"
	}
	fmt.Fprintf(&sb, "\n\rWear flags: &c%s&w\n\rEnter flags, 0 to quit : ", current)

	d.WriteToBuffer(sb.String())
	d.Olc.Mode = types.OEDIT_WEAR
}

// oeditDispLayerMenu renders the 9-option layer list. Mirrors C
// oedit_disp_layer_menu at src/ooedit.c:449-466.
//
// Sets OlcData.Mode = OEDIT_LAYERS.
func oeditDispLayerMenu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	var sb strings.Builder
	for i := 1; i < len(layerLabels); i++ {
		fmt.Fprintf(&sb, "&g%d&w) %s\n\r", i, layerLabels[i])
	}
	sb.WriteString("\n\rEnter layer : ")

	d.WriteToBuffer(sb.String())
	d.Olc.Mode = types.OEDIT_LAYERS
}

// oeditDispExtradescMenu renders the extradesc list + A/R/Q. Analogous to
// reditDispExtradescMenu in redit_menu.go. Wave 2 only renders the menu;
// add/edit/delete arms wire in Wave 3 (G8).
//
// Sets OlcData.Mode = OEDIT_EXTRADESC_MENU.
func oeditDispExtradescMenu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	idx, ok := d.Olc.Target.(*types.ObjIndexData)
	if !ok || idx == nil {
		return
	}

	var sb strings.Builder
	for i, ed := range idx.ExtraDescr {
		fmt.Fprintf(&sb, "&g%2d&w) Keyword: &O%s\n\r", i+1, ed.Keyword)
	}
	if len(idx.ExtraDescr) > 0 {
		sb.WriteString("\n\r")
	}
	sb.WriteString("&gA&w) Add a new description\n\r")
	sb.WriteString("&gR&w) Remove a description\n\r")
	sb.WriteString("&gQ&w) Quit\n\r")
	sb.WriteString("\n\rEnter choice: ")

	d.WriteToBuffer(sb.String())
	d.Olc.Mode = types.OEDIT_EXTRADESC_MENU
}

// --- Wave 2 value-menu STUBS — Wave 3 (G7) ports the per-item-type
// dispatch tables at src/ooedit.c:623-849 with full prompts. For now,
// each stub renders a placeholder + cancel-on-0 prompt and sets the
// corresponding OEDIT_VALUE_N mode so the parser arm can recognize the
// state and emit the same Wave-3 placeholder.

func oeditDispVal1Menu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	idx, ok := d.Olc.Target.(*types.ObjIndexData)
	if !ok || idx == nil {
		return
	}
	d.WriteToBufferf("Value 1 editing for type %s -- Wave 3 (enter 0 to cancel) : ",
		oTypeName(idx.ItemType))
	d.Olc.Mode = types.OEDIT_VALUE_1
}

func oeditDispVal2Menu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	idx, ok := d.Olc.Target.(*types.ObjIndexData)
	if !ok || idx == nil {
		return
	}
	d.WriteToBufferf("Value 2 editing for type %s -- Wave 3 (enter 0 to cancel) : ",
		oTypeName(idx.ItemType))
	d.Olc.Mode = types.OEDIT_VALUE_2
}

func oeditDispVal3Menu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	idx, ok := d.Olc.Target.(*types.ObjIndexData)
	if !ok || idx == nil {
		return
	}
	d.WriteToBufferf("Value 3 editing for type %s -- Wave 3 (enter 0 to cancel) : ",
		oTypeName(idx.ItemType))
	d.Olc.Mode = types.OEDIT_VALUE_3
}

func oeditDispVal4Menu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	idx, ok := d.Olc.Target.(*types.ObjIndexData)
	if !ok || idx == nil {
		return
	}
	d.WriteToBufferf("Value 4 editing for type %s -- Wave 3 (enter 0 to cancel) : ",
		oTypeName(idx.ItemType))
	d.Olc.Mode = types.OEDIT_VALUE_4
}

func oeditDispVal5Menu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	idx, ok := d.Olc.Target.(*types.ObjIndexData)
	if !ok || idx == nil {
		return
	}
	d.WriteToBufferf("Value 5 editing for type %s -- Wave 3 (enter 0 to cancel) : ",
		oTypeName(idx.ItemType))
	d.Olc.Mode = types.OEDIT_VALUE_5
}

func oeditDispVal6Menu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	idx, ok := d.Olc.Target.(*types.ObjIndexData)
	if !ok || idx == nil {
		return
	}
	d.WriteToBufferf("Value 6 editing for type %s -- Wave 3 (enter 0 to cancel) : ",
		oTypeName(idx.ItemType))
	d.Olc.Mode = types.OEDIT_VALUE_6
}
