package util

import (
	crand "crypto/rand"
	"math/rand/v2"
	"strconv"
	"strings"
)

// rng is the package-level random number generator, seeded from crypto/rand.
var rng *rand.Rand

func init() {
	var seed [32]byte
	if _, err := crand.Read(seed[:]); err != nil {
		panic("util: failed to seed RNG from crypto/rand: " + err.Error())
	}
	rng = rand.New(rand.NewChaCha8(seed))
}

// NumberRange returns a random integer in the inclusive range [from, to].
// If from >= to it returns from. Mirrors number_range() from db.c.
func NumberRange(from, to int) int {
	if to-from < 1 {
		return from
	}
	return rng.IntN(to-from+1) + from
}

// NumberPercent returns a random integer in the range [1, 100].
// Mirrors number_percent() from db.c.
func NumberPercent() int {
	return rng.IntN(100) + 1
}

// NumberDoor returns a random integer in the range [0, 9], covering the
// ten possible exit directions (N/S/E/W/U/D + diagonals).
// Mirrors number_door() from db.c.
func NumberDoor() int {
	return rng.IntN(10)
}

// NumberBits returns a random non-negative integer using only the lowest
// 'width' bits (i.e. in range [0, 2^width - 1]).
// Mirrors number_bits() from db.c.
func NumberBits(width int) int {
	if width <= 0 {
		return 0
	}
	return int(rng.Uint64() & uint64((1<<width)-1))
}

// DiceRoll rolls 'number' dice each with 'size' sides and returns the sum.
// Mirrors dice() from db.c.
func DiceRoll(number, size int) int {
	switch size {
	case 0:
		return 0
	case 1:
		return number
	}
	sum := 0
	for i := 0; i < number; i++ {
		sum += NumberRange(1, size)
	}
	return sum
}

// NumberFuzzy returns number-1, number, or number+1 at random, but never
// less than 1. Mirrors number_fuzzy() from db.c.
func NumberFuzzy(number int) int {
	switch NumberBits(2) {
	case 0:
		number--
	case 3:
		number++
	}
	return UMAX(1, number)
}

// DiceParse evaluates a SMAUG SmaugAff Duration/Modifier, skill DiceFormula,
// or morph stat expression. Empty string returns 0; parse errors return 0
// (matches C dice_parse at src/magic.c:1059-1066 — silently zero out the
// field so bitvector names like "sanctuary" fall through to the caller's
// BitVector handling).
//
// Patterns handled:
//
//   - ""                                   → 0
//   - plain signed int:  "12", "-3"
//   - level:             "l", "level", "i"
//   - scaling:           "l*15", "level*2", "l/2"
//   - arithmetic:        "l+25", "l-10", "1+(l/17)"
//   - parens:            "(l*3)+25", "-(l*4+50)"
//   - dice:              "NdM", "NdM+K", "3d8+(l-6)", "1d8+(l/3)"
//
// Grammar:
//
//	expr   = term ( (+|-) term )*
//	term   = factor ( (*|/|d) factor )*
//	factor = NUMBER | IDENT | '-' factor | '(' expr ')'
//
// C reference: src/magic.c rd_parse + dice_parse. The C implementation
// takes a `ch` parameter for RNG seeding; Go uses the package rand source
// seeded at init time so `ch` is not needed.
func DiceParse(s string, level int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	p := &diceExprParser{src: s, pos: 0, level: level}
	v, ok := p.parseExpr()
	if !ok || !p.eof() {
		return 0
	}
	return v
}

// diceExprParser is a recursive-descent evaluator for the DiceParse grammar.
type diceExprParser struct {
	src   string
	pos   int
	level int
}

func (p *diceExprParser) eof() bool {
	p.skipSpace()
	return p.pos >= len(p.src)
}

func (p *diceExprParser) skipSpace() {
	for p.pos < len(p.src) && (p.src[p.pos] == ' ' || p.src[p.pos] == '\t') {
		p.pos++
	}
}

func (p *diceExprParser) peek() byte {
	p.skipSpace()
	if p.pos >= len(p.src) {
		return 0
	}
	return p.src[p.pos]
}

// parseExpr: term ( (+|-) term )*
func (p *diceExprParser) parseExpr() (int, bool) {
	lhs, ok := p.parseTerm()
	if !ok {
		return 0, false
	}
	for {
		c := p.peek()
		if c != '+' && c != '-' {
			return lhs, true
		}
		p.pos++
		rhs, ok := p.parseTerm()
		if !ok {
			return 0, false
		}
		if c == '+' {
			lhs += rhs
		} else {
			lhs -= rhs
		}
	}
}

// parseTerm: factor ( (*|/|d) factor )*
func (p *diceExprParser) parseTerm() (int, bool) {
	lhs, ok := p.parseFactor()
	if !ok {
		return 0, false
	}
	for {
		c := p.peek()
		if c != '*' && c != '/' && c != 'd' {
			return lhs, true
		}
		if c == 'd' {
			if p.pos+1 >= len(p.src) {
				return lhs, true
			}
			next := p.src[p.pos+1]
			if !(next >= '0' && next <= '9') && next != '(' {
				return lhs, true
			}
		}
		p.pos++
		rhs, ok := p.parseFactor()
		if !ok {
			return 0, false
		}
		switch c {
		case '*':
			lhs *= rhs
		case '/':
			if rhs == 0 {
				return 0, false
			}
			lhs /= rhs
		case 'd':
			if lhs <= 0 || rhs <= 0 {
				lhs = 0
			} else {
				lhs = DiceRoll(lhs, rhs)
			}
		}
	}
}

// parseFactor: NUMBER | IDENT | '-' factor | '(' expr ')'
func (p *diceExprParser) parseFactor() (int, bool) {
	p.skipSpace()
	if p.pos >= len(p.src) {
		return 0, false
	}
	c := p.src[p.pos]
	if c == '(' {
		p.pos++
		v, ok := p.parseExpr()
		if !ok {
			return 0, false
		}
		p.skipSpace()
		if p.pos >= len(p.src) || p.src[p.pos] != ')' {
			return 0, false
		}
		p.pos++
		return v, true
	}
	if c == '-' {
		p.pos++
		v, ok := p.parseFactor()
		if !ok {
			return 0, false
		}
		return -v, true
	}
	if c == '+' {
		p.pos++
		return p.parseFactor()
	}
	if c >= '0' && c <= '9' {
		return p.parseNumber()
	}
	if diceIdentStart(c) {
		return p.parseIdent()
	}
	return 0, false
}

func (p *diceExprParser) parseNumber() (int, bool) {
	start := p.pos
	for p.pos < len(p.src) && p.src[p.pos] >= '0' && p.src[p.pos] <= '9' {
		p.pos++
	}
	n, err := strconv.Atoi(p.src[start:p.pos])
	if err != nil {
		return 0, false
	}
	return n, true
}

func (p *diceExprParser) parseIdent() (int, bool) {
	start := p.pos
	for p.pos < len(p.src) && diceIdentCont(p.src[p.pos]) {
		p.pos++
	}
	ident := strings.ToLower(p.src[start:p.pos])
	switch ident {
	case "l", "level", "i":
		return p.level, true
	}
	return 0, false
}

func diceIdentStart(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func diceIdentCont(c byte) bool {
	return diceIdentStart(c) || (c >= '0' && c <= '9')
}
