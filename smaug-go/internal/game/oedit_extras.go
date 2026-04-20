// Package game — oedit extradesc + affect sub-machines.
//
// Ports:
//   - Extradesc CRUD: oedit_disp_extradesc_menu (:469-502),
//     oedit_disp_extra_choice (:504-514), and OEDIT_EXTRADESC_* parse
//     arms at :1933-2036.
//   - Affect CRUD: oedit_disp_prompt_apply_menu (:517-544),
//     oedit_disp_affect_menu (:569-591), and OEDIT_AFFECT_* parse arms
//     at :1735-1931.
//
// Go-port divergences (vs C at ooedit.c):
//   - Prototype-only operation (see plan §Go Design); C iterates both
//     pIndexData->first_extradesc and obj->first_extradesc, Go walks
//     idx.ExtraDescr only.
//   - Extradesc list uses a slice (idx.ExtraDescr), not a linked list.
//     Delete uses append-slice-splice instead of LINK/UNLINK.
//   - No junk-empty-on-Q at OEDIT_EXTRADESC_CHOICE — matches C ooedit
//     verbatim (C divergence from redit's junk-on-Q behavior, see plan
//     §G8 note).
//   - Affect-flag bitmask editing (APPLY_AFFECT / RIS) ships a
//     placeholder prompt per plan §Q4 until medit lands. Numeric 0
//     cancels; any other input is accepted as a single bit toggle but
//     the full flag menu is not rendered.
//   - SkillLookup word-input for spells is numeric-only per plan §Q5
//     Option 2 (same as oedit_values.go).
//
// Plan: plan-phase6-olc-oedit.md §G8 + §G9.
package game

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// --- Extradesc menu + parse arms ---

// oeditDispExtraChoice renders the per-extradesc sub-menu (keyword /
// description / quit). Mirrors C :504-514. Caller must have stashed the
// selected *ExtraDescrData on d.Olc.Spare.
func oeditDispExtraChoice(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	ed, ok := d.Olc.Spare.(*types.ExtraDescrData)
	if !ok || ed == nil {
		OeditDispMenu(d)
		return
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "&g1&w) Keyword: &O%s\n\r", ed.Keyword)
	fmt.Fprintf(&sb, "&g2&w) Description: \n\r&O%s&w\n\r", ed.Description)
	sb.WriteString("\n\rChange which option? ")
	d.WriteToBuffer(sb.String())
	d.Olc.Mode = types.OEDIT_EXTRADESC_CHOICE
}

// oeditHandleExtradescMenu handles OEDIT_EXTRADESC_MENU input. Mirrors C
// :1994-2035. A=add, R=delete, Q=back, numeric=edit-by-index.
func oeditHandleExtradescMenu(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		oeditDispExtradescMenu(d)
		return
	}
	upper := strings.ToUpper(arg)[:1]
	switch upper {
	case "Q":
		OeditDispMenu(d)
		return
	case "A":
		ed := &types.ExtraDescrData{Keyword: "", Description: ""}
		idx.ExtraDescr = append(idx.ExtraDescr, ed)
		d.Olc.Spare = ed
		olcLog(d, "OBJ", "Added new exdesc")
		oeditDispExtraChoice(d)
		return
	case "R":
		d.WriteToBuffer("Delete which extra description? ")
		d.Olc.Mode = types.OEDIT_EXTRADESC_DELETE
		return
	}
	if n, err := strconv.Atoi(arg); err == nil {
		ed := oeditFindExtradesc(idx, n)
		if ed == nil {
			d.WriteToBuffer("Not found, try again: ")
			return
		}
		d.Olc.Spare = ed
		oeditDispExtraChoice(d)
		return
	}
	oeditDispExtradescMenu(d)
}

// oeditHandleExtradescChoice handles OEDIT_EXTRADESC_CHOICE input.
// Mirrors C :1950-1973. 0=back, 1=edit keyword, 2=edit description.
// NO junk-on-Q — matches C ooedit behavior exactly (deliberate divergence
// from redit's junk-on-Q guard).
func oeditHandleExtradescChoice(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	n, err := strconv.Atoi(arg)
	if err != nil {
		oeditDispExtraChoice(d)
		return
	}
	switch n {
	case 0:
		oeditDispExtradescMenu(d)
		return
	case 1:
		d.WriteToBuffer("Enter keywords, separated by spaces: ")
		d.Olc.Mode = types.OEDIT_EXTRADESC_KEY
		return
	case 2:
		ed, ok := d.Olc.Spare.(*types.ExtraDescrData)
		if !ok || ed == nil {
			oeditDispExtradescMenu(d)
			return
		}
		ch := d.Character
		ch.Substate = types.SUB_OBJ_EXTRA
		targetEd := ed
		ch.EditorSave = func(c *types.CharData) {
			if c == nil || c.Desc == nil {
				return
			}
			targetEd.Description = CopyBuffer(c)
			StopEditing(c)
			c.Desc.Connected = int(types.CON_OEDIT)
			olcLog(c.Desc, "OBJ", "Changed exdesc description for %s", targetEd.Keyword)
			oeditDispExtraChoice(c.Desc)
		}
		d.WriteToBuffer("Enter new extra description - :\n\r")
		StartEditing(ch, ed.Description)
		d.Olc.Mode = types.OEDIT_EXTRADESC_DESCRIPTION
		return
	default:
		oeditDispExtraChoice(d)
	}
}

// oeditHandleExtradescKey handles OEDIT_EXTRADESC_KEY input. Mirrors C
// :1933-1944. Stores keyword on the pending extradesc (via Olc.Spare)
// and returns to the extradesc choice menu. SmashTilde because keyword
// lands in the .are file format.
func oeditHandleExtradescKey(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	ed, ok := d.Olc.Spare.(*types.ExtraDescrData)
	if !ok || ed == nil {
		oeditDispExtradescMenu(d)
		return
	}
	if arg == "" {
		oeditDispExtraChoice(d)
		return
	}
	cleaned := util.SmashTilde(arg)
	olcLog(d, "OBJ", "Changed exdesc %s to %s", ed.Keyword, cleaned)
	ed.Keyword = cleaned
	oeditDispExtraChoice(d)
}

// oeditHandleExtradescDelete handles OEDIT_EXTRADESC_DELETE input.
// Mirrors C :1975-1992. 1-indexed; splices out the slice entry.
func oeditHandleExtradescDelete(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	n, err := strconv.Atoi(arg)
	if err != nil {
		d.WriteToBuffer("Extra description not found, try again: ")
		return
	}
	if n < 1 || n > len(idx.ExtraDescr) {
		d.WriteToBuffer("Extra description not found, try again: ")
		return
	}
	gone := idx.ExtraDescr[n-1]
	olcLog(d, "OBJ", "Deleted exdesc %s", gone.Keyword)
	idx.ExtraDescr = append(idx.ExtraDescr[:n-1], idx.ExtraDescr[n:]...)
	// Clear stash if it was pointing at the deleted entry.
	if d.Olc.Spare == gone {
		d.Olc.Spare = nil
	}
	oeditDispExtradescMenu(d)
}

// oeditFindExtradesc returns the 1-indexed extradesc, or nil if out of
// range. Mirrors C oedit_find_extradesc (prototype walk only).
func oeditFindExtradesc(idx *types.ObjIndexData, n int) *types.ExtraDescrData {
	if idx == nil {
		return nil
	}
	if n < 1 || n > len(idx.ExtraDescr) {
		return nil
	}
	return idx.ExtraDescr[n-1]
}

// --- Affect menu + parse arms (G9 placeholder per plan §Q4) ---

// oeditDispPromptApplyMenu renders the affect list + A/R/Q options.
// Mirrors C :517-544 but operates on idx.Affects only (prototype, per
// plan §Go Design). Sets Mode = OEDIT_AFFECT_MENU.
func oeditDispPromptApplyMenu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	idx, ok := d.Olc.Target.(*types.ObjIndexData)
	if !ok || idx == nil {
		return
	}
	var sb strings.Builder
	sb.WriteString("\n\rAffects:\n\r")
	for i, paf := range idx.Affects {
		fmt.Fprintf(&sb, " &g%2d&w) location=%d modifier=%d\n\r",
			i+1, paf.Location, paf.Modifier)
	}
	if len(idx.Affects) == 0 {
		sb.WriteString(" (none)\n\r")
	}
	sb.WriteString(" \n\r &gA&w) Add an affect\n\r")
	sb.WriteString(" &gR&w) Remove an affect\n\r")
	sb.WriteString(" &gQ&w) Quit\n\r")
	sb.WriteString("\n\rEnter option or affect#: ")
	d.WriteToBuffer(sb.String())
	d.Olc.Mode = types.OEDIT_AFFECT_MENU
}

// oeditHandleAffectMenu handles OEDIT_AFFECT_MENU input. Mirrors C
// :1735-1777. A=add (enter location prompt), R=remove-by-index (or
// prompt), Q=back, numeric=edit-by-index (not yet wired — treat as
// out-of-scope numeric input, redisplay menu).
func oeditHandleAffectMenu(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		oeditDispPromptApplyMenu(d)
		return
	}
	upper := strings.ToUpper(arg)[:1]
	switch upper {
	case "Q":
		d.Olc.Spare = nil
		OeditDispMenu(d)
		return
	case "A":
		// Allocate a pending AffectData stashed on Spare.
		d.Olc.Spare = &types.AffectData{Type: -1, Duration: -1}
		d.WriteToBuffer("Enter APPLY location (numeric, 0 to cancel): ")
		d.Olc.Mode = types.OEDIT_AFFECT_LOCATION
		return
	case "R":
		// "R <n>" shortcut — or bare "R" prompts.
		rest := strings.TrimSpace(arg[1:])
		if rest != "" {
			if n, err := strconv.Atoi(rest); err == nil {
				affectRemoveIndex(idx, n)
				olcLog(d, "OBJ", "Removed affect #%d", n)
			}
			oeditDispPromptApplyMenu(d)
			return
		}
		d.WriteToBuffer("Remove which affect? ")
		d.Olc.Mode = types.OEDIT_AFFECT_REMOVE
		return
	}
	// Numeric → would edit by index. Not yet implemented (plan §Q11).
	// Redisplay menu.
	oeditDispPromptApplyMenu(d)
}

// oeditHandleAffectLocation handles OEDIT_AFFECT_LOCATION input. Mirrors
// C :1779-1819. Numeric 0 junks the pending affect. Otherwise validates
// range and rejects APPLY_EXT_AFFECT. Prompts for modifier next.
//
// Divergence (plan §Q5): word-input via get_atype is numeric-only here;
// builders must supply the numeric APPLY_* value.
func oeditHandleAffectLocation(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	paf, ok := d.Olc.Spare.(*types.AffectData)
	if !ok || paf == nil {
		oeditDispPromptApplyMenu(d)
		return
	}
	n, err := strconv.Atoi(arg)
	if err != nil {
		d.WriteToBuffer("Invalid location, try again: ")
		return
	}
	if n == 0 {
		// Junk the pending affect.
		d.Olc.Spare = nil
		oeditDispPromptApplyMenu(d)
		return
	}
	if n < 0 || n >= types.MAX_APPLY_TYPE || n == types.APPLY_EXT_AFFECT {
		d.WriteToBuffer("Invalid location, try again: ")
		return
	}
	paf.Location = n
	d.Olc.Mode = types.OEDIT_AFFECT_MODIFIER
	// Per plan §Q4: APPLY_AFFECT / APPLY_RESISTANT / APPLY_IMMUNE /
	// APPLY_SUSCEPTIBLE would normally render the medit aff-flags / ris
	// bitmask editor. Until medit lands, emit a placeholder.
	switch n {
	case types.APPLY_AFFECT:
		d.WriteToBuffer("Affect-flag editing is a medit-plan dependency; enter numeric modifier (0 to cancel): ")
	case types.APPLY_RESISTANT, types.APPLY_IMMUNE, types.APPLY_SUSCEPTIBLE:
		d.WriteToBuffer("RIS-flag editing is a medit-plan dependency; enter numeric modifier (0 to cancel): ")
	case types.APPLY_WEAPONSPELL, types.APPLY_WEARSPELL, types.APPLY_REMOVESPELL:
		d.WriteToBuffer("Spell-name lookup not yet wired; enter numeric sn: ")
	default:
		d.WriteToBuffer("\n\rModifier: ")
	}
}

// oeditHandleAffectModifier handles OEDIT_AFFECT_MODIFIER input. Mirrors
// C :1821-1914. Appends a new AffectData to idx.Affects with the
// location + modifier, then returns to the prompt-apply menu.
func oeditHandleAffectModifier(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	paf, ok := d.Olc.Spare.(*types.AffectData)
	if !ok || paf == nil {
		oeditDispPromptApplyMenu(d)
		return
	}
	n, err := strconv.Atoi(arg)
	if err != nil {
		d.WriteToBuffer("Invalid modifier; enter numeric (0 to cancel): ")
		return
	}
	if n == 0 {
		// Cancel pending affect.
		d.Olc.Spare = nil
		oeditDispPromptApplyMenu(d)
		return
	}
	paf.Modifier = n
	// Append to prototype-only list.
	idx.Affects = append(idx.Affects, paf)
	olcLog(d, "OBJ", "Added new affect: loc=%d mod=%d", paf.Location, paf.Modifier)
	d.Olc.Spare = nil
	oeditDispPromptApplyMenu(d)
}

// oeditHandleAffectRemove handles OEDIT_AFFECT_REMOVE input. Mirrors C
// :1926-1931.
func oeditHandleAffectRemove(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	n, err := strconv.Atoi(arg)
	if err != nil {
		oeditDispPromptApplyMenu(d)
		return
	}
	affectRemoveIndex(idx, n)
	olcLog(d, "OBJ", "Removed affect #%d", n)
	oeditDispPromptApplyMenu(d)
}

// affectRemoveIndex removes the 1-indexed affect; no-op on out-of-range.
func affectRemoveIndex(idx *types.ObjIndexData, n int) {
	if idx == nil {
		return
	}
	if n < 1 || n > len(idx.Affects) {
		return
	}
	idx.Affects = append(idx.Affects[:n-1], idx.Affects[n:]...)
}
