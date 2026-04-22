// Package game — medit Wave 4 G10: full affect-list editor.
//
// C medit has NO affect-list editor of its own — the AFF_FLAGS bitmask
// (Wave 3 G9) covers the affected_by toggles, but C builders edit per-
// affect AffectData entries (location/modifier/duration/type) only via
// the flat `mset <victim> apply <args>` path. The Go port adds a
// menu-driven flow as a builder QoL enhancement, mirroring the OEDIT_*
// affect editor that already exists for objects.
//
// State machine:
//
//	MEDIT_AFFECT_MENU      → A/R/Q (add/remove/quit-to-main)
//	MEDIT_AFFECT_LOCATION  → numeric APPLY_* picker; 0 cancels;
//	                         APPLY_NONE / APPLY_EXT_AFFECT rejected
//	                         (matches oedit precedent + plan §665)
//	MEDIT_AFFECT_MODIFIER  → branches on staged Location:
//	                           scalar APPLY_* → numeric modifier
//	                           APPLY_AFFECT   → AFF_* bitmask via olcBitmaskEdit
//	                           APPLY_RESISTANT/IMMUNE/SUSCEPT → RIS_* bitmask
//	                           0 cancels in scalar path; "done"/"quit"/empty
//	                           commits the bitmask; "0" alone cancels bitmask too.
//	MEDIT_AFFECT_REMOVE    → numeric 1-based index; 0 cancels; OOB re-prompts
//
// The pending AffectData is stashed on Olc.Spare across the LOCATION →
// MODIFIER round-trip, mirroring the oedit pattern. On commit the
// helper appends to victim.Affects via handler.AffectToChar (which also
// applies the stat modification via AffectModify); on remove the helper
// calls handler.AffectRemove (which reverses the stat mod and splices
// the slice).
//
// Prototype mirror: MobIndexData has NO Affects field (verified at
// internal/types/mob_index.go) — unlike ObjIndexData which does. The
// affect list is therefore an instance-only mutation; the change does
// NOT propagate to subsequent NPC instantiations from the same prototype.
// Documented inline so a future MobIndexData.Affects extension can land
// the mirror without disturbing this code path.
//
// Plan: plan-phase6-olc-medit.md §G10 (A20, A21, A22).
package game

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

// meditDispAffectListMenu renders the affect list with 1-based indices,
// followed by the Add/Remove/Quit options. Mirrors oedit's
// oeditDispPromptApplyMenu shape but operates on victim.Affects.
//
// Sets Olc.Mode = MEDIT_AFFECT_MENU as a side effect so the parser knows
// where to dispatch the next input line.
func meditDispAffectListMenu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	victim, ok := d.Olc.Target.(*types.CharData)
	if !ok || victim == nil {
		return
	}
	var sb strings.Builder
	sb.WriteString("\n\rAffects:\n\r")
	if len(victim.Affects) == 0 {
		sb.WriteString(" (none)\n\r")
	}
	for i, paf := range victim.Affects {
		fmt.Fprintf(&sb, " %2d) location=%d modifier=%d duration=%d type=%d bv=0x%x\n\r",
			i+1, paf.Location, paf.Modifier, paf.Duration, paf.Type,
			affectBVMask(paf.BitVector))
	}
	sb.WriteString("\n\r A) Add an affect\n\r")
	sb.WriteString(" R) Remove an affect\n\r")
	sb.WriteString(" Q) Quit to main menu\n\r")
	sb.WriteString("\n\rEnter option: ")
	d.WriteToBuffer(sb.String())
	d.Olc.Mode = types.MEDIT_AFFECT_MENU
}

// affectBVMask returns the low 32 bits of a BitVector for display only.
// BitVector is [4]uint32; the AFF_* range fits in the first word for
// the foreseeable future.
func affectBVMask(bv types.BitVector) uint32 {
	return bv[0]
}

// meditArmAffectMenu handles MEDIT_AFFECT_MENU input. Mirrors oedit's
// oeditHandleAffectMenu shape but operates on a victim CharData.
//
//	A → stage a fresh AffectData on Olc.Spare, transition to LOCATION
//	R → transition to REMOVE (the parser-level handler reads index)
//	Q → return to NPC or PC main menu via npcOrPcMenu
//
// Anything else redisplays the menu. "R 3" shortcut matches oedit
// (numeric tail-arg removes immediately).
func meditArmAffectMenu(d *types.DescriptorData, victim *types.CharData, arg string) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		meditDispAffectListMenu(d)
		return
	}
	upper := strings.ToUpper(arg[:1])
	switch upper {
	case "Q":
		d.Olc.Spare = nil
		d.Olc.Mode = npcOrPcMenu(victim)
		MeditDispMenu(d)
		return
	case "A":
		// Stash a pending affect placeholder. Type/Duration default match
		// the oedit precedent: Type=-1, Duration=-1 (permanent).
		d.Olc.Spare = &types.AffectData{Type: -1, Duration: -1}
		d.WriteToBuffer("Enter APPLY location (numeric, 0 to cancel): ")
		d.Olc.Mode = types.MEDIT_AFFECT_LOCATION
		return
	case "R":
		rest := strings.TrimSpace(arg[1:])
		if rest != "" {
			if n, err := strconv.Atoi(rest); err == nil {
				meditAffectRemoveIndex(d, victim, n)
			}
			meditDispAffectListMenu(d)
			return
		}
		d.WriteToBuffer("Remove which affect? ")
		d.Olc.Mode = types.MEDIT_AFFECT_REMOVE
		return
	}
	meditDispAffectListMenu(d)
}

// meditArmAffectLocation handles MEDIT_AFFECT_LOCATION input. Numeric
// APPLY_* (1..MAX_APPLY_TYPE-1, excluding APPLY_EXT_AFFECT) is accepted;
// 0 cancels and discards the staged AffectData (matches oedit). Out-of-
// range or non-numeric re-prompts.
//
// Word-input (apply name → numeric) is deferred — `util.ApplyflagNames`
// does not yet exist (see follow-ups in TODO-updates.md).
func meditArmAffectLocation(d *types.DescriptorData, victim *types.CharData, arg string) {
	arg = strings.TrimSpace(arg)
	paf, ok := d.Olc.Spare.(*types.AffectData)
	if !ok || paf == nil {
		// Spare lost — abandon and redisplay.
		meditDispAffectListMenu(d)
		return
	}
	n, err := strconv.Atoi(arg)
	if err != nil {
		d.WriteToBuffer("Invalid location, try again: ")
		return
	}
	if n == 0 {
		// Cancel the staged affect.
		d.Olc.Spare = nil
		meditDispAffectListMenu(d)
		return
	}
	if n < 0 || n >= types.MAX_APPLY_TYPE || n == types.APPLY_EXT_AFFECT {
		d.WriteToBuffer("Invalid location, try again: ")
		return
	}
	paf.Location = n
	d.Olc.Mode = types.MEDIT_AFFECT_MODIFIER
	switch n {
	case types.APPLY_AFFECT:
		d.WriteToBuffer("Toggle AFF_* flags (name or 1-based index, 'done' to commit, 0 to cancel): ")
	case types.APPLY_RESISTANT, types.APPLY_IMMUNE, types.APPLY_SUSCEPTIBLE:
		d.WriteToBuffer("Toggle RIS_* flags (name or 1-based index, 'done' to commit, 0 to cancel): ")
	default:
		d.WriteToBuffer("\n\rModifier (0 cancels): ")
	}
}

// meditArmAffectModifier handles MEDIT_AFFECT_MODIFIER input. Branches
// on the staged Location:
//
//   - Scalar APPLY_*: numeric modifier; 0 cancels (matches oedit). On
//     non-zero, calls handler.AffectToChar (appends + applies stat mod).
//   - APPLY_AFFECT / APPLY_RESISTANT / APPLY_IMMUNE / APPLY_SUSCEPTIBLE:
//     dispatches into olcBitmaskEdit on a *staging* int. "done"/"quit"/
//     empty commits the accumulated modifier; "0" alone cancels.
//
// On commit the staged AffectData becomes a real entry on victim.Affects
// via handler.AffectToChar (which both appends AND applies the stat
// modification). The pending stash is cleared either way.
func meditArmAffectModifier(d *types.DescriptorData, victim *types.CharData, arg string) {
	arg = strings.TrimSpace(arg)
	paf, ok := d.Olc.Spare.(*types.AffectData)
	if !ok || paf == nil {
		meditDispAffectListMenu(d)
		return
	}

	// Bitmask APPLY_* types — dispatch through the shared helper.
	var bitmaskTable string
	switch paf.Location {
	case types.APPLY_AFFECT:
		bitmaskTable = "AFF_FLAGS"
	case types.APPLY_RESISTANT, types.APPLY_IMMUNE, types.APPLY_SUSCEPTIBLE:
		bitmaskTable = "RIS"
	}
	if bitmaskTable != "" {
		if arg == "0" {
			// Explicit cancel — drop the staged affect with no commit.
			d.Olc.Spare = nil
			meditDispAffectListMenu(d)
			return
		}
		done := olcBitmaskEdit(d, bitmaskTable, &paf.Modifier, arg)
		if !done {
			// Helper toggled and expects another line, OR rejected with
			// its own prompt. Stay in mode either way.
			return
		}
		meditAffectCommit(d, victim, paf)
		return
	}

	// Scalar path — numeric modifier (0 cancels).
	n, err := strconv.Atoi(arg)
	if err != nil {
		d.WriteToBuffer("Invalid modifier; enter numeric (0 to cancel): ")
		return
	}
	if n == 0 {
		d.Olc.Spare = nil
		meditDispAffectListMenu(d)
		return
	}
	paf.Modifier = n
	meditAffectCommit(d, victim, paf)
}

// meditAffectCommit appends the staged affect to victim.Affects, applies
// the stat-modify side effect via handler.AffectToChar, logs the change,
// clears the stash, and redisplays the affect-menu. NOTE: handler
// .AffectToChar copies the input AffectData into a fresh allocation
// before appending, so passing the staging pointer is safe.
//
// Prototype mirror is intentionally NOT applied here — MobIndexData has
// no Affects field today (see file-header note). When a future schema
// extension adds it, mirror under the established
// IsNPC()+ACT_PROTOTYPE+IndexData!=nil triple-gate.
func meditAffectCommit(d *types.DescriptorData, victim *types.CharData, paf *types.AffectData) {
	handler.AffectToChar(victim, paf)
	olcLog(d, "MOB", "Added affect: loc=%d mod=%d", paf.Location, paf.Modifier)
	d.Olc.Spare = nil
	if d.Olc != nil {
		d.Olc.Change = true
	}
	meditDispAffectListMenu(d)
}

// meditArmAffectRemove handles MEDIT_AFFECT_REMOVE input. Numeric 1-based
// index; 0 cancels back to the menu; OOB re-prompts.
func meditArmAffectRemove(d *types.DescriptorData, victim *types.CharData, arg string) {
	arg = strings.TrimSpace(arg)
	n, err := strconv.Atoi(arg)
	if err != nil || n == 0 {
		meditDispAffectListMenu(d)
		return
	}
	meditAffectRemoveIndex(d, victim, n)
	meditDispAffectListMenu(d)
}

// meditAffectRemoveIndex removes the 1-indexed affect from victim.Affects
// (and reverses the stat modification via handler.AffectRemove). No-op
// on out-of-range. Plan §G10 A22.
func meditAffectRemoveIndex(d *types.DescriptorData, victim *types.CharData, n int) {
	if victim == nil {
		return
	}
	if n < 1 || n > len(victim.Affects) {
		return
	}
	target := victim.Affects[n-1]
	handler.AffectRemove(victim, target)
	olcLog(d, "MOB", "Removed affect #%d (loc=%d mod=%d)", n, target.Location, target.Modifier)
	if d.Olc != nil {
		d.Olc.Change = true
	}
}
