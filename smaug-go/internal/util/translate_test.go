package util

import (
	"strings"
	"testing"
)

func TestLanguageBit_PrefixDeterministic(t *testing.T) {
	// "g" should always return the first entry starting with g in declaration
	// order — i.e. "goblin" (bit 14). Go map iteration would randomize this,
	// so run many iterations to catch regressions to a map-backed table.
	want := LanguageBit("g")
	for i := 0; i < 1000; i++ {
		if got := LanguageBit("g"); got != want {
			t.Fatalf("LanguageBit('g') nondeterministic: got %d and %d", want, got)
		}
	}
	if want == 0 {
		t.Errorf("'g' should resolve to a language")
	}
}

func TestLanguageBit(t *testing.T) {
	cases := []struct {
		in  string
		out uint32
	}{
		{"common", 1 << 0},
		{"Elven", 1 << 1},
		{"DWARV", 1 << 2}, // prefix match
		{"garble", 0},
	}
	for _, tc := range cases {
		if got := LanguageBit(tc.in); got != tc.out {
			t.Errorf("LanguageBit(%q) = %d, want %d", tc.in, got, tc.out)
		}
	}
}

func TestScramble_DifferentLanguagesDifferentOutputs(t *testing.T) {
	elven := Scramble("hello world", 1<<1)
	orcish := Scramble("hello world", 1<<5)
	if elven == "hello world" {
		t.Errorf("elven scramble should alter input")
	}
	if elven == orcish {
		t.Errorf("different languages should produce different scrambles")
	}
}

func TestScramble_PreservesWordShape(t *testing.T) {
	orig := "hello, world!"
	got := Scramble(orig, 1<<2)
	if len(got) != len(orig) {
		t.Errorf("length changed: %q → %q", orig, got)
	}
	if !strings.Contains(got, ", ") || !strings.HasSuffix(got, "!") {
		t.Errorf("punctuation/spacing not preserved: %q", got)
	}
}

func TestScramble_ZeroBitNoOp(t *testing.T) {
	if Scramble("plain", 0) != "plain" {
		t.Error("zero lang bit should be no-op")
	}
}

func TestScramble_PreservesColorCodes(t *testing.T) {
	got := Scramble("&Whello&D", 1<<1)
	if !strings.Contains(got, "&W") || !strings.Contains(got, "&D") {
		t.Errorf("color codes lost: %q", got)
	}
}

func TestScramble_Deterministic(t *testing.T) {
	a := Scramble("test phrase", 1<<3)
	b := Scramble("test phrase", 1<<3)
	if a != b {
		t.Errorf("scramble not deterministic: %q vs %q", a, b)
	}
}
