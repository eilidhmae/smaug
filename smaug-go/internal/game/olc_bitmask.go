// Package game — shared bitmask-editor helper used by medit / oedit OLC
// flag-toggle arms.
//
// olcBitmaskEdit drives a one-line input into a named bitmask field on
// the victim / object under edit. The shape is shared across the
// medit arms (MEDIT_NPC_FLAGS, MEDIT_PC_FLAGS, MEDIT_AFF_FLAGS,
// MEDIT_PCDATA_FLAGS, MEDIT_PARTS, MEDIT_RESISTANT/IMMUNE/SUSCEPTIBLE)
// and — once the oedit back-wire lands (plan-phase6-olc-medit.md §G14
// back-wire) — the oedit OEDIT_AFFECT_MODIFIER arm for APPLY_AFFECT,
// APPLY_RESISTANT, APPLY_IMMUNE, APPLY_SUSCEPTIBLE.
//
// The helper intentionally takes a string table name (not a pointer
// to the table slice) so callers can pin the wire-up to a well-known
// symbolic table without exposing internal table layout. Lookup is
// dispatched through lookupBitmaskTable.
//
// Plan: plan-phase6-olc-medit.md §G9 (A17, A18, A19).
package game

import (
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// lookupBitmaskTable returns the flag-name slice for a given symbolic
// table name. Returns nil for unknown table names; callers can treat
// the nil return as an internal misconfiguration.
//
// Table names (uppercase, C-style) match plan-phase6-olc-medit.md §447:
//
//	"ACT_FLAGS"  → util.ActflagNames   (NPC act flags)
//	"PLR_FLAGS"  → util.PlrflagNames   (PC player flags)
//	"AFF_FLAGS"  → util.AffflagNames   (affect flags)
//	"PCFLAG"     → util.PcflagNames    (PC-data flags)
//	"PART"       → util.PartflagNames  (body-part flags)
//	"RIS"        → util.RisflagNames   (resist / immune / suscept)
func lookupBitmaskTable(tableName string) []string {
	switch tableName {
	case "ACT_FLAGS":
		return util.ActflagNames
	case "PLR_FLAGS":
		return util.PlrflagNames
	case "AFF_FLAGS":
		return util.AffflagNames
	case "PCFLAG":
		return util.PcflagNames
	case "PART":
		return util.PartflagNames
	case "RIS":
		return util.RisflagNames
	default:
		return nil
	}
}

// olcBitmaskEdit parses a single builder input line against the named
// flag table and toggles the corresponding bit(s) in *current.
//
// Semantics (A19 — consistent across all wires):
//
//   - Empty input, "done", or "quit" (case-insensitive) → return true
//     with no mutation. The caller uses the return value to detect the
//     exit request and redisplays the main menu appropriate to the
//     victim.
//   - Numeric single-token input: C-parity. `atoi` the token; if the
//     parsed value is 0 return false (C-parity with MEDIT_PARTS / PC /
//     AFF / PCDATA arms which treat `0` as "end of input"). Otherwise
//     subtract 1 from the number (C offset convention) and toggle
//     that bit. Out-of-range → write "Invalid flag, try again: " and
//     return false without toggling.
//   - Token-stream input (space-separated keywords): for each token,
//     look up the flag name in the table; if found toggle the bit, if
//     not found write "Invalid flag, try again: " and abort the rest
//     of the line without toggling any further bits. Numeric tokens
//     inside a multi-token stream use the offset convention too.
//
// Returns true when the caller should exit the mode (empty / done /
// quit); false when the caller should stay in the mode (either the
// toggle succeeded or input was rejected, in which case the helper
// already emitted the error message and a re-prompt is expected).
//
// olcLog entries — "Added" / "Removed" the flag name — are emitted per
// toggle, matching C's xTOGGLE_BIT + olc_log pattern at
// src/omedit.c:1494-1509 etc.
func olcBitmaskEdit(
	d *types.DescriptorData,
	tableName string,
	current *int,
	arg string,
) bool {
	if d == nil || current == nil {
		return true
	}
	table := lookupBitmaskTable(tableName)
	if table == nil {
		// Misconfiguration — treat as exit to avoid stalling.
		return true
	}

	arg = strings.TrimSpace(arg)

	// Exit semantics — A19.
	switch strings.ToLower(arg) {
	case "", "done", "quit":
		return true
	}

	// Single-token numeric: C-parity shortcut per omedit.c:1485-1491
	// (MEDIT_PC_FLAGS) etc. If arg is a bare integer, handle it
	// specially so "0" means "exit" (not toggle bit -1).
	if !strings.ContainsAny(arg, " \t") {
		if n, err := strconv.Atoi(arg); err == nil {
			if n == 0 {
				return true // C: if (number == 0) break;
			}
			n-- // C offset: number -= 1
			if n < 0 || n >= len(table) {
				d.WriteToBuffer("Invalid flag, try again: ")
				return false
			}
			*current ^= 1 << n
			added := *current&(1<<n) != 0
			logBitmaskToggle(d, tableName, n, table, added)
			return false
		}
	}

	// Token-stream parse: iterate space-separated tokens.
	toggled := false
	for _, tok := range strings.Fields(arg) {
		var bit int
		if n, err := strconv.Atoi(tok); err == nil {
			// Numeric token inside a stream still uses the C offset.
			if n == 0 {
				continue
			}
			bit = n - 1
			if bit < 0 || bit >= len(table) {
				d.WriteToBuffer("Invalid flag, try again: ")
				return false
			}
		} else {
			bit = lookupFlagInTable(table, tok)
			if bit < 0 {
				d.WriteToBuffer("Invalid flag, try again: ")
				return false
			}
		}
		*current ^= 1 << bit
		toggled = true
		added := *current&(1<<bit) != 0
		logBitmaskToggle(d, tableName, bit, table, added)
	}
	_ = toggled // no-op retained for readability; absence of toggle returns false below
	return false
}

// olcBitmaskEditBitVector is the BitVector-field variant. Medit's
// ACT_* / PLR_* / AFF_* fields on CharData are BitVector (128-bit)
// rather than plain int. Shape is identical to olcBitmaskEdit but
// Set/Clear/Toggle go through BitVector methods.
func olcBitmaskEditBitVector(
	d *types.DescriptorData,
	tableName string,
	current *types.BitVector,
	arg string,
) bool {
	if d == nil || current == nil {
		return true
	}
	table := lookupBitmaskTable(tableName)
	if table == nil {
		return true
	}

	arg = strings.TrimSpace(arg)
	switch strings.ToLower(arg) {
	case "", "done", "quit":
		return true
	}

	if !strings.ContainsAny(arg, " \t") {
		if n, err := strconv.Atoi(arg); err == nil {
			if n == 0 {
				return true
			}
			n--
			if n < 0 || n >= len(table) {
				d.WriteToBuffer("Invalid flag, try again: ")
				return false
			}
			current.Toggle(n)
			added := current.IsSet(n)
			logBitmaskToggle(d, tableName, n, table, added)
			return false
		}
	}

	for _, tok := range strings.Fields(arg) {
		var bit int
		if n, err := strconv.Atoi(tok); err == nil {
			if n == 0 {
				continue
			}
			bit = n - 1
			if bit < 0 || bit >= len(table) {
				d.WriteToBuffer("Invalid flag, try again: ")
				return false
			}
		} else {
			bit = lookupFlagInTable(table, tok)
			if bit < 0 {
				d.WriteToBuffer("Invalid flag, try again: ")
				return false
			}
		}
		current.Toggle(bit)
		added := current.IsSet(bit)
		logBitmaskToggle(d, tableName, bit, table, added)
	}
	return false
}

// lookupFlagInTable returns the bit index for a flag name in the
// given table, or -1 if not found. Empty / reserved names ("", "r1",
// "r2") are intentionally NOT matched even when the input happens to
// be literally "r1"/"r2" — the C lookup functions return the index
// for them because they are positional placeholders, but the OLC
// editor rejects them. Case-insensitive.
func lookupFlagInTable(table []string, name string) int {
	for i, n := range table {
		if n == "" {
			continue
		}
		if strings.EqualFold(n, name) {
			return i
		}
	}
	return -1
}

// logBitmaskToggle emits a consistent olcLog entry for a toggle.
func logBitmaskToggle(d *types.DescriptorData, tableName string, bit int, table []string, added bool) {
	name := ""
	if bit >= 0 && bit < len(table) {
		name = table[bit]
	}
	if name == "" {
		name = "(reserved)"
	}
	verb := "Removed"
	if added {
		verb = "Added"
	}
	olcLog(d, "MOB", "%s %s flag %s", verb, tableName, name)
}
