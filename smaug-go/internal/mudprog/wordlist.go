package mudprog

import "strings"

// wordlistMatch ports C's `rprog_wordlist_check` / `oprog_wordlist_check`
// keyword-matching algorithm at src/mud_prog.c:4109-4178 and :3813-3880.
//
// Algorithm:
//  1. Lower-case arglist and input.
//  2. If arglist starts with "p " the remainder is treated as ONE phrase
//     and matched as a single substring with word-boundary checks.
//  3. Otherwise split arglist on whitespace; any one keyword matching
//     (with word-boundary checks) wins.
//  4. Word boundary — left: start-of-text or preceded by ' '. Right:
//     end-of-text or followed by ' ', '\n', '\r'. Other characters (tab,
//     punctuation) do NOT count as word boundaries, matching C which
//     only checks ' ', '\n', '\r', '\0'.
//
// Returns false for empty arglist or empty input (defensive; callers
// short-circuit before us but the helper is self-contained).
func wordlistMatch(arglist, input string) bool {
	list := strings.ToLower(strings.TrimSpace(arglist))
	if list == "" {
		return false
	}
	text := strings.ToLower(input)
	if text == "" {
		return false
	}
	// "p " prefix: treat remainder as ONE phrase.
	if strings.HasPrefix(list, "p ") {
		phrase := list[2:]
		return containsAtWordBoundary(text, phrase)
	}
	for _, kw := range strings.Fields(list) {
		if containsAtWordBoundary(text, kw) {
			return true
		}
	}
	return false
}

// containsAtWordBoundary reports whether needle appears in text at a
// position where both its left side (start-of-text or preceded by ' ')
// and right side (end-of-text or followed by ' ', '\n', '\r') satisfy
// C's `rprog_wordlist_check` word-boundary check.
//
// Empty needle is treated as non-matching (matches C which would require
// `strstr` to find something — and `strstr(x, "")` returns x, but the
// boundary logic wouldn't consider a zero-length find a "word". The
// rprog caller walks a non-empty keyword list; the guard is defensive.
func containsAtWordBoundary(text, needle string) bool {
	if needle == "" {
		return false
	}
	idx := 0
	for idx <= len(text)-len(needle) {
		found := strings.Index(text[idx:], needle)
		if found < 0 {
			return false
		}
		start := idx + found
		// Left boundary: start-of-text or preceded by ' '.
		leftOk := start == 0 || text[start-1] == ' '
		// Right boundary: end-of-text or followed by ' ', '\n', '\r'.
		end := start + len(needle)
		rightOk := end == len(text) ||
			text[end] == ' ' ||
			text[end] == '\n' ||
			text[end] == '\r'
		if leftOk && rightOk {
			return true
		}
		// Advance past this non-boundary match.
		idx = start + 1
	}
	return false
}
