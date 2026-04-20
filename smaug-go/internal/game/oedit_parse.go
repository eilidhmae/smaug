// Package game — interactive object editor (CON_OEDIT substate).
//
// This file ports src/ooedit.c oedit_parse at :1172-2160. The pulse loop
// dispatches input addressed to a descriptor in CON_OEDIT directly here
// (see loop.go processInput); the nanny is NOT invoked for menu input.
// Mirrors C src/smaug.c:1641-1652 (same routing pattern as CON_REDIT).
//
// Wave 2 scope: main-menu dispatch (G5) + simple-field arms (G6).
// Per-item-type value bodies (G7), extradesc add/edit/delete (G8), affect
// menus (G9), and DoOedit menu-entry (G10) land in Wave 3.
//
// Plan: plan-phase6-olc-oedit.md §G3 / §G5 / §G6.
package game

import (
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// worldObjLookup is the seam the parser uses to resolve object vnums.
// Points at worldRef.GetObjIndex at boot time (via SetWorldRef shared
// with redit_parse.go); tests can override directly.
var worldObjLookup = func(vnum int) *types.ObjIndexData {
	if worldRef == nil {
		return nil
	}
	return worldRef.GetObjIndex(vnum)
}

// oeditParse is the top-level CON_OEDIT dispatcher. Keyed on d.Olc.Mode,
// it mutates the object under edit and either returns (stay in same mode)
// or transitions to a new mode + redisplays the appropriate menu.
//
// The caller (processInput) has already validated d.Connected ==
// CON_OEDIT. If d.Olc is nil or d.Olc.Target is not an *ObjIndexData
// (should not happen — stale state bug elsewhere), cleanup and drop
// back to CON_PLAYING defensively.
func oeditParse(d *types.DescriptorData, arg string) {
	if d == nil {
		return
	}
	if d.Olc == nil || d.Character == nil {
		// Defensive: impossible state; restore playable.
		d.Connected = int(types.CON_PLAYING)
		return
	}
	idx, ok := d.Olc.Target.(*types.ObjIndexData)
	if !ok || idx == nil {
		// Target lost (should not happen). Restore playable.
		cleanupOlc(d)
		return
	}

	switch d.Olc.Mode {
	case types.OEDIT_MAIN_MENU:
		oeditHandleMainMenu(d, idx, arg)

	case types.OEDIT_EDIT_NAMELIST:
		oeditHandleNamelist(d, idx, arg)

	case types.OEDIT_SHORTDESC:
		oeditHandleShortdesc(d, idx, arg)

	case types.OEDIT_LONGDESC:
		// Editor-driven; defensive only — input arrives via /s callback.
		d.WriteToBuffer("Use /s to save the editor.\n\r")

	case types.OEDIT_ACTDESC:
		// Editor-driven; defensive only.
		d.WriteToBuffer("Use /s to save the editor.\n\r")

	case types.OEDIT_TYPE:
		oeditHandleType(d, idx, arg)

	case types.OEDIT_EXTRAS:
		oeditHandleExtras(d, idx, arg)

	case types.OEDIT_WEAR:
		oeditHandleWear(d, idx, arg)

	case types.OEDIT_WEIGHT:
		oeditHandleWeight(d, idx, arg)

	case types.OEDIT_COST:
		oeditHandleCost(d, idx, arg)

	case types.OEDIT_COSTPERDAY:
		oeditHandleRent(d, idx, arg)

	case types.OEDIT_TIMER:
		oeditHandleTimer(d, idx, arg)

	case types.OEDIT_LEVEL:
		oeditHandleLevel(d, idx, arg)

	case types.OEDIT_LAYERS:
		oeditHandleLayers(d, idx, arg)

	case types.OEDIT_VALUE_1:
		oeditHandleValue1(d, idx, arg)
	case types.OEDIT_VALUE_2:
		oeditHandleValue2(d, idx, arg)
	case types.OEDIT_VALUE_3:
		oeditHandleValue3(d, idx, arg)
	case types.OEDIT_VALUE_4:
		oeditHandleValue4(d, idx, arg)
	case types.OEDIT_VALUE_5:
		oeditHandleValue5(d, idx, arg)
	case types.OEDIT_VALUE_6:
		oeditHandleValue6(d, idx, arg)

	case types.OEDIT_AFFECT_MENU:
		oeditHandleAffectMenu(d, idx, arg)
	case types.OEDIT_AFFECT_LOCATION:
		oeditHandleAffectLocation(d, idx, arg)
	case types.OEDIT_AFFECT_MODIFIER:
		oeditHandleAffectModifier(d, idx, arg)
	case types.OEDIT_AFFECT_REMOVE:
		oeditHandleAffectRemove(d, idx, arg)
	case types.OEDIT_AFFECT_RIS:
		// Unreachable in shipped C per §C Bug Catalog #1. Preserve as
		// dead path — redisplay main defensively.
		util.Bug("oeditParse: OEDIT_AFFECT_RIS unreachable path")
		OeditDispMenu(d)

	case types.OEDIT_EXTRADESC_MENU:
		oeditHandleExtradescMenu(d, idx, arg)
	case types.OEDIT_EXTRADESC_CHOICE:
		oeditHandleExtradescChoice(d, idx, arg)
	case types.OEDIT_EXTRADESC_KEY:
		oeditHandleExtradescKey(d, idx, arg)
	case types.OEDIT_EXTRADESC_DESCRIPTION:
		// Editor-driven; defensive only.
		d.WriteToBuffer("Use /s to save the editor.\n\r")
	case types.OEDIT_EXTRADESC_DELETE:
		oeditHandleExtradescDelete(d, idx, arg)

	default:
		util.Bug("oeditParse: unhandled mode %d", d.Olc.Mode)
		OeditDispMenu(d)
	}
}

// --- G5 main-menu dispatch ---

// oeditHandleMainMenu interprets one keystroke at the top-level menu.
// Mirrors C oedit_parse case OEDIT_MAIN_MENU at src/ooedit.c:1207-1338.
func oeditHandleMainMenu(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		OeditDispMenu(d)
		return
	}
	switch strings.ToUpper(arg)[:1] {
	case "Q":
		cleanupOlc(d)
		d.WriteToBuffer("Exiting editor.\n\r")
		return
	case "1":
		d.WriteToBuffer("Enter namelist : ")
		d.Olc.Mode = types.OEDIT_EDIT_NAMELIST
	case "2":
		d.WriteToBuffer("Enter short desc : ")
		d.Olc.Mode = types.OEDIT_SHORTDESC
	case "3":
		// Long desc — route through the line editor and re-enter
		// CON_OEDIT after /s fires. Same trampoline shape as redit
		// case "2" at redit_parse.go:163-184.
		ch := d.Character
		ch.Substate = types.SUB_OBJ_LONG
		targetIdx := idx
		ch.EditorSave = func(c *types.CharData) {
			if c == nil || c.Desc == nil {
				return
			}
			targetIdx.Description = CopyBuffer(c)
			StopEditing(c)
			c.Desc.Connected = int(types.CON_OEDIT)
			olcLog(c.Desc, "OBJ", "Edited long description")
			OeditDispMenu(c.Desc)
		}
		d.WriteToBuffer("Enter long desc:\n\r")
		StartEditing(ch, idx.Description)
		d.Olc.Mode = types.OEDIT_LONGDESC
	case "4":
		// Action desc — Go-port divergence from C (C uses inline
		// single-line input; Go uses the text editor for multi-line
		// fidelity per plan §SUB_OBJ_ACTION rationale).
		ch := d.Character
		ch.Substate = types.SUB_OBJ_ACTION
		targetIdx := idx
		ch.EditorSave = func(c *types.CharData) {
			if c == nil || c.Desc == nil {
				return
			}
			targetIdx.ActionDesc = CopyBuffer(c)
			StopEditing(c)
			c.Desc.Connected = int(types.CON_OEDIT)
			olcLog(c.Desc, "OBJ", "Edited action description")
			OeditDispMenu(c.Desc)
		}
		d.WriteToBuffer("Enter action desc:\n\r")
		StartEditing(ch, idx.ActionDesc)
		d.Olc.Mode = types.OEDIT_ACTDESC
	case "5":
		oeditDispTypeMenu(d)
	case "6":
		oeditDispExtraMenu(d)
	case "7":
		oeditDispWearMenu(d)
	case "8":
		d.WriteToBuffer("Enter weight : ")
		d.Olc.Mode = types.OEDIT_WEIGHT
	case "9":
		d.WriteToBuffer("Enter cost : ")
		d.Olc.Mode = types.OEDIT_COST
	case "A":
		d.WriteToBuffer("Enter rent : ")
		d.Olc.Mode = types.OEDIT_COSTPERDAY
	case "B":
		d.WriteToBuffer("Enter timer : ")
		d.Olc.Mode = types.OEDIT_TIMER
	case "C":
		// Level gate per C oedit_parse :1230-1234 — only LEVEL_GREATER+
		// can edit object level.
		if d.Character.GetTrust() < types.LEVEL_GREATER {
			d.WriteToBuffer("You aren't powerful enough to set the level.\n\r")
			OeditDispMenu(d)
			return
		}
		d.WriteToBuffer("Enter level : ")
		d.Olc.Mode = types.OEDIT_LEVEL
	case "D":
		// Layerable precheck per C ooedit.c:1261-1275. D only opens the
		// layer menu when the object has at least one of BODY/ABOUT/ARMS/
		// FEET/HANDS/LEGS/WAIST wear bits set; otherwise refuse with
		// C-verbatim message and stay on main menu.
		const layerableMask = (1 << 3) | // ITEM_WEAR_BODY
			(1 << 10) | // ITEM_WEAR_ABOUT
			(1 << 8) | // ITEM_WEAR_ARMS
			(1 << 6) | // ITEM_WEAR_FEET
			(1 << 7) | // ITEM_WEAR_HANDS
			(1 << 5) | // ITEM_WEAR_LEGS
			(1 << 11) // ITEM_WEAR_WAIST
		if idx.WearFlags&layerableMask == 0 {
			d.WriteToBuffer("The wear location of this object is not layerable.\n\r")
			OeditDispMenu(d)
			return
		}
		oeditDispLayerMenu(d)
	case "E":
		oeditDispVal1Menu(d)
	case "F":
		oeditDispPromptApplyMenu(d)
	case "G":
		oeditDispExtradescMenu(d)
	default:
		// Match C behavior: invalid char silently re-renders the menu.
		OeditDispMenu(d)
	}
}

// --- G6 simple-field parse arms ---

// oeditHandleNamelist sets idx.Name from arg. SmashTilde because Name
// lands in the .are file format. Mirrors C OEDIT_EDIT_NAMELIST at
// src/ooedit.c:1340-1342.
func oeditHandleNamelist(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		OeditDispMenu(d)
		return
	}
	idx.Name = util.SmashTilde(arg)
	olcLog(d, "OBJ", "Changed name to %s", idx.Name)
	OeditDispMenu(d)
}

// oeditHandleShortdesc sets idx.ShortDescr. Mirrors C OEDIT_SHORTDESC at
// src/ooedit.c:1349-1352.
func oeditHandleShortdesc(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		OeditDispMenu(d)
		return
	}
	idx.ShortDescr = util.SmashTilde(arg)
	olcLog(d, "OBJ", "Changed short to %s", idx.ShortDescr)
	OeditDispMenu(d)
}

// oeditHandleType validates and sets idx.ItemType. Accepts numeric or
// case-insensitive type-name. Mirrors C OEDIT_TYPE at
// src/ooedit.c:1344-1348. Out-of-range rejection re-prompts WITHOUT mode
// change (preserve C's stay-in-OEDIT_TYPE behavior).
func oeditHandleType(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		oeditDispTypeMenu(d)
		return
	}
	var n int
	if v, err := strconv.Atoi(arg); err == nil {
		n = v
	} else {
		n = getOtype(arg)
	}
	if n < 1 || n > types.MAX_ITEM_TYPE {
		d.WriteToBuffer("Invalid choice, try again : ")
		return
	}
	idx.ItemType = n
	olcLog(d, "OBJ", "Changed type to %s", oTypeName(n))
	OeditDispMenu(d)
}

// oeditHandleExtras toggles one or more extra-flag bits on idx.ExtraFlags.
// Mirrors C OEDIT_EXTRAS at src/ooedit.c:1361-1408.
//
// Modes:
//   - integer 0 → exit back to main menu
//   - integer 1..MAX_ITEM_FLAG → toggle that 1-indexed bit and stop (C
//     processes only ONE flag per numeric input — preserved)
//   - word list → per word, look up via getOflag and toggle (processes ALL)
//
// ITEM_PROTOTYPE (bit 30) toggle requires GetTrust() >= LEVEL_GREATER OR
// is_name("protoflag", bestowments). Pinned by mutation gate #6.
func oeditHandleExtras(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		oeditDispExtraMenu(d)
		return
	}
	// Try integer first (whole-input is a number).
	if n, err := strconv.Atoi(arg); err == nil {
		if n == 0 {
			OeditDispMenu(d)
			return
		}
		if n < 1 || n > types.MAX_ITEM_FLAG {
			d.WriteToBuffer("Invalid flag, try again: ")
			return
		}
		bit := n - 1
		if !canEditExtraFlag(d.Character, bit) {
			d.WriteToBuffer("You cannot change the prototype flag.\n\r")
			oeditDispExtraMenu(d)
			return
		}
		idx.ExtraFlags.Toggle(bit)
		action := "Removed"
		if idx.ExtraFlags.IsSet(bit) {
			action = "Added"
		}
		label := "(unnamed)"
		if bit < len(oFlagNames) {
			label = oFlagNames[bit]
		}
		olcLog(d, "OBJ", "%s the extra flag %s", action, label)
		oeditDispExtraMenu(d)
		return
	}
	// Word list: toggle each. Unknown words are skipped silently.
	words := strings.Fields(arg)
	for _, w := range words {
		bit := getOflag(w)
		if bit < 0 {
			continue
		}
		if !canEditExtraFlag(d.Character, bit) {
			d.WriteToBuffer("You cannot change the prototype flag.\n\r")
			continue
		}
		idx.ExtraFlags.Toggle(bit)
		action := "Removed"
		if idx.ExtraFlags.IsSet(bit) {
			action = "Added"
		}
		olcLog(d, "OBJ", "%s the extra flag %s", action, oFlagNames[bit])
	}
	oeditDispExtraMenu(d)
}

// canEditExtraFlag returns true if ch is allowed to toggle the given
// extra-flag bit. Only ITEM_PROTOTYPE is gated. Mirrors C check at
// src/ooedit.c:1389-1392.
func canEditExtraFlag(ch *types.CharData, bit int) bool {
	if bit != types.ITEM_PROTOTYPE {
		return true
	}
	if ch == nil {
		return false
	}
	if ch.GetTrust() >= types.LEVEL_GREATER {
		return true
	}
	if ch.PCData != nil && util.IsName("protoflag", ch.PCData.Bestowments) {
		return true
	}
	return false
}

// oeditHandleWear toggles wear-flag bits on idx.WearFlags. ITEM_DUAL_WIELD
// (bit 15) is REJECTED per C parity at src/ooedit.c:895-917 — the menu
// skips it and the parser refuses numeric and word input for it.
//
// Modes:
//   - integer 0 → exit back to main menu
//   - integer 1..ITEM_WEAR_MAX+1 → toggle that 1-indexed bit (reject 16)
//   - word list → per word, lookup via getWflag and toggle each
func oeditHandleWear(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		oeditDispWearMenu(d)
		return
	}
	if n, err := strconv.Atoi(arg); err == nil {
		if n == 0 {
			OeditDispMenu(d)
			return
		}
		if n < 1 || n > types.ITEM_WEAR_MAX+1 {
			d.WriteToBuffer("Invalid flag, try again: ")
			return
		}
		bit := n - 1
		if bit == 15 {
			// ITEM_DUAL_WIELD — combat-only, refuse.
			d.WriteToBuffer("Cannot set dual_wield via OLC.\n\r")
			oeditDispWearMenu(d)
			return
		}
		idx.WearFlags ^= 1 << bit
		action := "Removed"
		if idx.WearFlags&(1<<bit) != 0 {
			action = "Added"
		}
		label := "(unnamed)"
		if bit < len(wFlagNames) {
			label = wFlagNames[bit]
		}
		olcLog(d, "OBJ", "%s the wear flag %s", action, label)
		oeditDispWearMenu(d)
		return
	}
	words := strings.Fields(arg)
	for _, w := range words {
		bit := getWflag(w)
		if bit < 0 {
			// Includes "dual_wield" — getWflag rejects it explicitly.
			continue
		}
		idx.WearFlags ^= 1 << bit
		action := "Removed"
		if idx.WearFlags&(1<<bit) != 0 {
			action = "Added"
		}
		olcLog(d, "OBJ", "%s the wear flag %s", action, wFlagNames[bit])
	}
	oeditDispWearMenu(d)
}

// oeditHandleWeight sets idx.Weight. Clamped [1, 1000]. Per C
// OEDIT_WEIGHT at src/ooedit.c.
func oeditHandleWeight(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	n := parseIntDefault(arg, 0)
	idx.Weight = uRange(1, n, 1000)
	olcLog(d, "OBJ", "Changed weight to %d", idx.Weight)
	OeditDispMenu(d)
}

// oeditHandleCost sets idx.GoldCost. Clamped [0, 100000].
func oeditHandleCost(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	n := parseIntDefault(arg, 0)
	idx.GoldCost = uRange(0, n, 100000)
	olcLog(d, "OBJ", "Changed cost to %d", idx.GoldCost)
	OeditDispMenu(d)
}

// oeditHandleRent sets idx.Rent. Clamped [0, 100000].
func oeditHandleRent(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	n := parseIntDefault(arg, 0)
	idx.Rent = uRange(0, n, 100000)
	olcLog(d, "OBJ", "Changed rent to %d", idx.Rent)
	OeditDispMenu(d)
}

// oeditHandleTimer sets idx.Layers' companion field — but ObjIndexData has
// no Timer field (it's per-instance ObjData.Timer). Go-port simplification:
// the prototype timer prompt is accepted but discarded with a logged note,
// matching the existing flat DoOset omission. Documented divergence from C.
func oeditHandleTimer(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	n := parseIntDefault(arg, 0)
	clamped := uRange(0, n, 1000)
	// Prototype has no Timer field in Go (per-instance only). Log the
	// intent without mutation; flat DoOset behaves identically.
	olcLog(d, "OBJ", "Timer prompt accepted (%d) — prototype has no timer field", clamped)
	_ = idx
	OeditDispMenu(d)
}

// oeditHandleLevel sets idx.Level. Clamped [0, MAX_LEVEL]. Defense-in-depth
// re-checks the LEVEL_GREATER trust gate at parse time (main-menu dispatch
// already gates entry; this catches any direct CON_OEDIT input bypass).
func oeditHandleLevel(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	if d.Character.GetTrust() < types.LEVEL_GREATER {
		d.WriteToBuffer("You aren't powerful enough to set the level.\n\r")
		OeditDispMenu(d)
		return
	}
	n := parseIntDefault(arg, 0)
	idx.Level = uRange(0, n, types.MAX_LEVEL)
	olcLog(d, "OBJ", "Changed level to %d", idx.Level)
	OeditDispMenu(d)
}

// oeditHandleLayers parses a 0-9 selection from oeditDispLayerMenu and
// sets idx.Layers per the C table at src/ooedit.c:1500-1526.
//
//	0 → exit to main
//	1 → idx.Layers = 0 (Nothing)
//	2..9 → toggle bit (n-2): 1, 2, 4, 8, 16, 32, 64, 128
func oeditHandleLayers(d *types.DescriptorData, idx *types.ObjIndexData, arg string) {
	n, err := strconv.Atoi(strings.TrimSpace(arg))
	if err != nil {
		d.WriteToBuffer("Invalid selection, try again: ")
		return
	}
	if n == 0 {
		OeditDispMenu(d)
		return
	}
	if n == 1 {
		idx.Layers = 0
		olcLog(d, "OBJ", "Cleared layers")
		oeditDispLayerMenu(d)
		return
	}
	if n < 2 || n > 9 {
		d.WriteToBuffer("Invalid selection, try again: ")
		return
	}
	bit := 1 << (n - 2)
	idx.Layers ^= bit
	action := "Removed"
	if idx.Layers&bit != 0 {
		action = "Added"
	}
	olcLog(d, "OBJ", "%s layer %s", action, layerLabels[n])
	oeditDispLayerMenu(d)
}
