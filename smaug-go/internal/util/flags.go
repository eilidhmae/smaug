package util

import "strings"

// RisflagNames maps bit position to the RIS_* flag name used by
// `stset immune/resist/susceptible`. Derived from C's ris_flags table
// in src/stances.c (variant) and types/constants.go:196-217 which
// defines 22 RIS_* bits (0..21) terminating in "magic"+"paralysis".
//
// Used by FlagString for rendering bitsets to name lists and by
// GetRisflag for parsing stset input. Order MUST match the bit
// layout — index i corresponds to `1 << i`.
var RisflagNames = []string{
	// bits 0-7
	"fire", "cold", "electricity", "energy",
	"blunt", "pierce", "slash", "acid",
	// bits 8-13
	"poison", "drain", "sleep", "charm",
	"hold", "nonmagic",
	// bits 14-19 (plus1..plus6)
	"plus1", "plus2", "plus3", "plus4", "plus5", "plus6",
	// bits 20-21 (audit-corrected 2026-04-18 — constants.go:196-217
	// defines 22 RIS_* bits; terminal names are "magic" + "paralysis",
	// not "plus7" as an earlier draft had it)
	"magic", "paralysis",
}

// GetRisflag returns the bit index (0..21) for a RIS_* flag name, or -1
// if the name is unrecognized. Case-insensitive. Mirrors C's
// `get_risflag` in src/act_wiz.c.
func GetRisflag(name string) int {
	for i, n := range RisflagNames {
		if strings.EqualFold(n, name) {
			return i
		}
	}
	return -1
}

// FlagString renders a bitset to a space-separated list of flag names.
// Returns "none" when no bit is set. Mirrors C's `flag_string` in db.c.
func FlagString(flags int, names []string) string {
	var parts []string
	for i, name := range names {
		if flags&(1<<i) != 0 {
			parts = append(parts, name)
		}
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, " ")
}
