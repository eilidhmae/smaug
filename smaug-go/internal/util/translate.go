package util

import "strings"

// langEntry binds a canonical language name to its LANG_* bit. The slice is
// ordered deterministically so prefix matches are stable — Go map iteration is
// randomized, so a map would make "go" nondeterministic between goblin/gith/
// gnome/god.
type langEntry struct {
	name string
	bit  uint32
}

var langTable = []langEntry{
	{"common", 1 << 0},
	{"elven", 1 << 1},
	{"dwarven", 1 << 2},
	{"pixie", 1 << 3},
	{"ogre", 1 << 4},
	{"orcish", 1 << 5},
	{"trollish", 1 << 6},
	{"rodent", 1 << 7},
	{"insectoid", 1 << 8},
	{"mammal", 1 << 9},
	{"reptile", 1 << 10},
	{"dragon", 1 << 11},
	{"spiritual", 1 << 12},
	{"magical", 1 << 13},
	{"goblin", 1 << 14},
	{"god", 1 << 15},
	{"ancient", 1 << 16},
	{"halfling", 1 << 17},
	{"clan", 1 << 18},
	{"gith", 1 << 19},
	{"gnome", 1 << 20},
}

// LanguageBit returns the LANG_* bit matching name (case-insensitive). Exact
// match wins; otherwise the first prefix match in declaration order wins. The
// slice ordering makes this deterministic regardless of Go map iteration.
func LanguageBit(name string) uint32 {
	name = strings.ToLower(name)
	for _, e := range langTable {
		if e.name == name {
			return e.bit
		}
	}
	for _, e := range langTable {
		if strings.HasPrefix(e.name, name) {
			return e.bit
		}
	}
	return 0
}

// LanguageName returns the canonical name for a LANG_* bit, or "" if unknown.
func LanguageName(bit uint32) string {
	for _, e := range langTable {
		if e.bit == bit {
			return e.name
		}
	}
	return ""
}

// Scramble produces a deterministic pseudo-translation of text for listeners
// who don't share the speaker's language. Preserves non-letter characters
// (spaces, punctuation, color codes prefixed with '&') and word lengths.
// Matches C scramble() intent without depending on rand so tests are stable.
func Scramble(text string, langBit uint32) string {
	if langBit == 0 {
		return text
	}
	// Rotate letters by a per-language offset derived from the bit index.
	rot := 0
	for b := langBit; b > 1; b >>= 1 {
		rot++
	}
	rot = (rot*7 + 3) % 26
	var b strings.Builder
	b.Grow(len(text))
	skipNext := false
	for i, r := range text {
		if skipNext {
			b.WriteRune(r)
			skipNext = false
			continue
		}
		if r == '&' && i+1 < len(text) {
			b.WriteRune(r)
			skipNext = true
			continue
		}
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(rune('a' + (int(r-'a')+rot)%26))
		case r >= 'A' && r <= 'Z':
			b.WriteRune(rune('A' + (int(r-'A')+rot)%26))
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
