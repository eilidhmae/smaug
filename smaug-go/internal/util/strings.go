// Package util provides general-purpose utility functions ported from the
// SMAUG MUD codebase (handler.c, interp.c, db.c).
package util

import (
	"strconv"
	"strings"
	"unicode"
)

// OneArgument extracts the first word from argument and returns it (lowercased)
// along with the remaining string. Single- or double-quoted phrases are treated
// as a single argument. This mirrors the C one_argument() from interp.c.
func OneArgument(argument string) (first, rest string) {
	argument = strings.TrimLeftFunc(argument, unicode.IsSpace)
	if argument == "" {
		return "", ""
	}

	// Check for quoted argument.
	delim := ' '
	if argument[0] == '\'' || argument[0] == '"' {
		delim = rune(argument[0])
		argument = argument[1:]
	}

	var buf strings.Builder
	i := 0
	for i < len(argument) {
		r := rune(argument[i])
		if r == delim {
			i++
			break
		}
		buf.WriteRune(unicode.ToLower(r))
		i++
	}

	rest = strings.TrimLeftFunc(argument[i:], unicode.IsSpace)
	return buf.String(), rest
}

// SmashTilde replaces all tilde characters (~) with dashes (-).
// Tildes serve as string terminators in SMAUG area files, so player-entered
// text must have them removed before being written to disk.
func SmashTilde(str string) string {
	return strings.ReplaceAll(str, "~", "-")
}

// Capitalize returns the string with its first letter uppercased and the
// rest lowercased, matching the SMAUG capitalize() from db.c.
func Capitalize(str string) string {
	if str == "" {
		return ""
	}
	lower := strings.ToLower(str)
	runes := []rune(lower)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// IsName checks whether str is a prefix match of any word in the
// space-separated namelist (case-insensitive). This corresponds to
// is_name_prefix() in handler.c.
func IsName(str, namelist string) bool {
	if str == "" {
		return false
	}
	str = strings.ToLower(str)
	remaining := namelist
	for remaining != "" {
		var name string
		name, remaining = OneArgument(remaining)
		if name == "" {
			break
		}
		// Prefix match: str is a prefix of name.
		if strings.HasPrefix(name, str) {
			return true
		}
	}
	return false
}

// IsNameExact checks whether str exactly matches any word in the
// space-separated namelist (case-insensitive). This corresponds to
// is_name() in handler.c.
func IsNameExact(str, namelist string) bool {
	if str == "" {
		return false
	}
	str = strings.ToLower(str)
	remaining := namelist
	for remaining != "" {
		var name string
		name, remaining = OneArgument(remaining)
		if name == "" {
			break
		}
		if name == str {
			return true
		}
	}
	return false
}

// IsNumber returns true if the string represents an integer (optionally
// preceded by a minus sign). Matches is_number() from interp.c.
func IsNumber(str string) bool {
	if str == "" {
		return false
	}
	start := 0
	if str[0] == '-' {
		if len(str) == 1 {
			return false
		}
		start = 1
	}
	for i := start; i < len(str); i++ {
		if str[i] < '0' || str[i] > '9' {
			return false
		}
	}
	return true
}

// NumberArgument parses strings of the form "3.sword" into a count and the
// argument portion. If no dot-prefix is present the count defaults to 1.
// Mirrors number_argument() from interp.c.
func NumberArgument(argument string) (number int, arg string) {
	dot := strings.IndexByte(argument, '.')
	if dot == -1 {
		return 1, argument
	}
	n, err := strconv.Atoi(argument[:dot])
	if err != nil {
		return 1, argument
	}
	return n, argument[dot+1:]
}

// UMIN returns the smaller of a and b.
func UMIN(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// UMAX returns the larger of a and b.
func UMAX(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// URANGE clamps b to the range [a, c].
func URANGE(a, b, c int) int {
	return UMAX(a, UMIN(b, c))
}
