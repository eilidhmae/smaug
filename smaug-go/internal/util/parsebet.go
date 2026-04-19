package util

import "strconv"

// Advatoi mirrors the non-GSC C advatoi (src/bet.h:118-163): parses
// an integer string with optional `k` (×1000) or `m` (×1000000)
// multiplier. Digits after the multiplier scale down by a factor of
// ten each step — so "14k42" → 14*1000 + 4*100 + 2*10 = 14420.
//
// Returns 0 on any parse failure (bad character, k/m twice, etc.),
// matching C semantics.
func Advatoi(s string) int {
	i := 0
	number := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		number = number*10 + int(s[i]-'0')
		i++
	}

	multiplier := 0
	if i < len(s) {
		c := s[i]
		switch c {
		case 'K', 'k':
			multiplier = 1000
			number *= multiplier
			i++
		case 'M', 'm':
			multiplier = 1000000
			number *= multiplier
			i++
		case 0:
			// end of string (defensive — len bound already covers this)
		default:
			return 0
		}
	}

	// Fold in trailing digits that scale down by the multiplier.
	for i < len(s) && s[i] >= '0' && s[i] <= '9' && multiplier > 1 {
		multiplier /= 10
		number += int(s[i]-'0') * multiplier
		i++
	}

	// Any non-digit trailing character = parse failure.
	if i < len(s) && !(s[i] >= '0' && s[i] <= '9') {
		return 0
	}

	return number
}

// ParseBet mirrors C parsebet (src/bet.h:188-210). Four cases:
//   - leading digit: delegate to Advatoi (absolute bet, k/m aware)
//   - "+": add percent; "+<N>" adds N%, bare "+" defaults to +25%
//   - "*" / "x": multiply; "x<N>" / "*<N>" multiplies, bare defaults to ×2
//   - empty / unknown: 0
//
// `currentbet` is the current winning bid; used by the relative
// (+ / * / x) forms.
func ParseBet(currentBet int, s string) int {
	if s == "" {
		return 0
	}
	if s[0] >= '0' && s[0] <= '9' {
		return Advatoi(s)
	}
	if s[0] == '+' {
		if len(s) == 1 {
			return currentBet * 125 / 100
		}
		n, err := strconv.Atoi(s[1:])
		if err != nil {
			// C's atoi returns 0 on junk — (100 + 0) / 100 = currentBet.
			return currentBet
		}
		return currentBet * (100 + n) / 100
	}
	if s[0] == '*' || s[0] == 'x' {
		if len(s) == 1 {
			return currentBet * 2
		}
		n, err := strconv.Atoi(s[1:])
		if err != nil {
			return 0
		}
		return currentBet * n
	}
	return 0
}
