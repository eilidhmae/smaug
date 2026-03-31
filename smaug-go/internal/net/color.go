package net

import "strings"

// ANSI escape sequence components.
const (
	ansiEsc   = "\033["
	ansiReset = ansiEsc + "0m"
)

// colorMap maps SMAUG &-codes to ANSI escape sequences.
var colorMap = map[byte]string{
	// Bright / bold foreground
	'R': ansiEsc + "1;31m", // bright red
	'G': ansiEsc + "1;32m", // bright green
	'Y': ansiEsc + "1;33m", // bright yellow
	'B': ansiEsc + "1;34m", // bright blue
	'P': ansiEsc + "1;35m", // bright magenta (pink)
	'C': ansiEsc + "1;36m", // bright cyan
	'W': ansiEsc + "1;37m", // bright white
	'O': ansiEsc + "0;33m", // orange (dark yellow)

	// Dark / normal foreground
	'r': ansiEsc + "0;31m", // dark red
	'g': ansiEsc + "0;32m", // dark green
	'y': ansiEsc + "0;33m", // dark yellow
	'b': ansiEsc + "0;34m", // dark blue
	'p': ansiEsc + "0;35m", // dark magenta
	'c': ansiEsc + "0;36m", // dark cyan
	'w': ansiEsc + "0;37m", // grey
	'o': ansiEsc + "0;33m", // dark yellow (same as 'y')

	// Reset
	'D': ansiReset,
	'd': ansiReset,
}

// ProcessColors converts SMAUG &-codes in text to ANSI escape sequences.
// If ansiEnabled is false the color codes are stripped instead.
func ProcessColors(text string, ansiEnabled bool) string {
	// Fast path: no ampersand means nothing to process.
	if !strings.Contains(text, "&") {
		return text
	}

	var b strings.Builder
	b.Grow(len(text))

	i := 0
	for i < len(text) {
		if text[i] == '&' && i+1 < len(text) {
			next := text[i+1]

			if next == '&' {
				// Escaped ampersand: && becomes literal &.
				b.WriteByte('&')
				i += 2
				continue
			}

			if ansi, ok := colorMap[next]; ok {
				if ansiEnabled {
					b.WriteString(ansi)
				}
				// If ANSI disabled we simply skip the two-character code.
				i += 2
				continue
			}
		}

		b.WriteByte(text[i])
		i++
	}

	return b.String()
}
