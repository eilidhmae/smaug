package util

import (
	"math"
	"testing"
)

func TestAdvatoi_BareInt(t *testing.T) {
	if got := Advatoi("123"); got != 123 {
		t.Errorf("Advatoi(123) = %d, want 123", got)
	}
	if got := Advatoi(""); got != 0 {
		t.Errorf("Advatoi(empty) = %d, want 0", got)
	}
	if got := Advatoi("abc"); got != 0 {
		t.Errorf("Advatoi(abc) = %d, want 0", got)
	}
	if got := Advatoi("0"); got != 0 {
		t.Errorf("Advatoi(0) = %d, want 0", got)
	}
}

func TestAdvatoi_KSuffix(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"14k", 14000},
		{"14K", 14000},
		{"14k42", 14420},   // 14*1000 + 4*100 + 2*10
		{"14k1234", 14123}, // drops final digit (multiplier floor at 1)
		{"1k", 1000},
		{"0k", 0},
	}
	for _, c := range cases {
		if got := Advatoi(c.in); got != c.want {
			t.Errorf("Advatoi(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestAdvatoi_MSuffix(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"23m", 23000000},
		{"23M", 23000000},
		{"23m5", 23500000},
		{"1m", 1000000},
	}
	for _, c := range cases {
		if got := Advatoi(c.in); got != c.want {
			t.Errorf("Advatoi(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestAdvatoi_RejectBadChar(t *testing.T) {
	if got := Advatoi("14q"); got != 0 {
		t.Errorf("Advatoi(14q) = %d, want 0", got)
	}
	if got := Advatoi("14k42x"); got != 0 {
		t.Errorf("Advatoi(14k42x) = %d, want 0", got)
	}
}

func TestParseBet_BareDigit(t *testing.T) {
	if got := ParseBet(1000, "500"); got != 500 {
		t.Errorf("ParseBet(1000, 500) = %d, want 500", got)
	}
	if got := ParseBet(1000, "14k"); got != 14000 {
		t.Errorf("ParseBet(1000, 14k) = %d, want 14000", got)
	}
}

func TestParseBet_PercentDefault(t *testing.T) {
	if got := ParseBet(1000, "+"); got != 1250 {
		t.Errorf("ParseBet(1000, +) = %d, want 1250", got)
	}
}

func TestParseBet_PercentN(t *testing.T) {
	if got := ParseBet(1000, "+50"); got != 1500 {
		t.Errorf("ParseBet(1000, +50) = %d, want 1500", got)
	}
	if got := ParseBet(10000, "+100"); got != 20000 {
		t.Errorf("ParseBet(10000, +100) = %d, want 20000", got)
	}
}

func TestParseBet_MultiplyDefault(t *testing.T) {
	if got := ParseBet(1000, "*"); got != 2000 {
		t.Errorf("ParseBet(1000, *) = %d, want 2000", got)
	}
	if got := ParseBet(1000, "x"); got != 2000 {
		t.Errorf("ParseBet(1000, x) = %d, want 2000", got)
	}
}

func TestParseBet_MultiplyN(t *testing.T) {
	if got := ParseBet(1000, "x10"); got != 10000 {
		t.Errorf("ParseBet(1000, x10) = %d, want 10000", got)
	}
	if got := ParseBet(1000, "*5"); got != 5000 {
		t.Errorf("ParseBet(1000, *5) = %d, want 5000", got)
	}
}

func TestParseBet_Empty(t *testing.T) {
	if got := ParseBet(1000, ""); got != 0 {
		t.Errorf("ParseBet(1000, empty) = %d, want 0", got)
	}
}

func TestParseBet_Unknown(t *testing.T) {
	if got := ParseBet(1000, "abc"); got != 0 {
		t.Errorf("ParseBet(1000, abc) = %d, want 0", got)
	}
}

// Security adversary 2026-04-19: a crafted `*N` whose product wraps int
// would otherwise produce a deceptively-small bet that passes the
// auction's 2-billion ceiling check. ParseBet must return 0 instead of
// wrapping.
func TestParseBet_MultiplyOverflowReturnsZero(t *testing.T) {
	// 10000 * 1844674407370960 wraps to a small positive on 64-bit int.
	if got := ParseBet(10000, "*1844674407370960"); got != 0 {
		t.Errorf("ParseBet(10000, *huge) = %d, want 0 (overflow guard)", got)
	}
	if got := ParseBet(10000, "x1844674407370960"); got != 0 {
		t.Errorf("ParseBet(10000, xhuge) = %d, want 0 (overflow guard)", got)
	}
}

func TestParseBet_PercentOverflowReturnsZero(t *testing.T) {
	// 1000 * (100 + huge) — the `100+n` step alone wraps.
	if got := ParseBet(1000, "+9223372036854775000"); got != 0 {
		t.Errorf("ParseBet(1000, +huge) = %d, want 0 (overflow guard)", got)
	}
}

func TestMulOverflows(t *testing.T) {
	cases := []struct {
		a, b int
		want bool
	}{
		{0, math.MaxInt, false},
		{math.MaxInt, 0, false},
		{-1, math.MaxInt, false}, // negative operands are out-of-contract; treated as no-overflow.
		{2, math.MaxInt / 2, false},
		{2, math.MaxInt/2 + 1, true},
		{10000, 1844674407370960, true},
	}
	for _, c := range cases {
		if got := mulOverflows(c.a, c.b); got != c.want {
			t.Errorf("mulOverflows(%d, %d) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}
