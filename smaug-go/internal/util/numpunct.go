package util

import (
	"strconv"
	"strings"
)

// NumPunct formats an integer with commas as thousands separators.
// Mirrors C num_punct() in src/comm.c — used throughout the auction
// subsystem to format gold amounts in broadcast messages. Negative
// numbers are handled: `-1234` → `"-1,234"`.
//
// Go stdlib has no built-in thousands-separator formatter (the stdlib
// `fmt.Sprintf("%d", n)` emits raw digits); adding a dependency on
// `golang.org/x/text/message` solely for this would be overkill.
func NumPunct(n int) string {
	if n < 0 {
		return "-" + NumPunct(-n)
	}
	s := strconv.Itoa(n)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	first := len(s) % 3
	if first > 0 {
		b.WriteString(s[:first])
		if len(s) > first {
			b.WriteByte(',')
		}
	}
	for i := first; i < len(s); i += 3 {
		b.WriteString(s[i : i+3])
		if i+3 < len(s) {
			b.WriteByte(',')
		}
	}
	return b.String()
}
