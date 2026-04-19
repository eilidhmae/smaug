package util

import "testing"

func TestRisflagNames_Count(t *testing.T) {
	if got := len(RisflagNames); got != 22 {
		t.Errorf("len(RisflagNames) = %d, want 22 (matches constants.go:196-217)", got)
	}
}

func TestRisflagNames_TerminalEntries(t *testing.T) {
	if RisflagNames[20] != "magic" {
		t.Errorf("bit 20 name = %q, want magic", RisflagNames[20])
	}
	if RisflagNames[21] != "paralysis" {
		t.Errorf("bit 21 name = %q, want paralysis", RisflagNames[21])
	}
}

func TestGetRisflag_KnownBits(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"fire", 0},
		{"Fire", 0},
		{"FIRE", 0},
		{"cold", 1},
		{"paralysis", 21},
		{"magic", 20},
		{"plus6", 19},
		{"unknown", -1},
		{"", -1},
	}
	for _, tc := range cases {
		if got := GetRisflag(tc.in); got != tc.want {
			t.Errorf("GetRisflag(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestFlagString_Empty(t *testing.T) {
	if got := FlagString(0, RisflagNames); got != "none" {
		t.Errorf("FlagString(0) = %q, want none", got)
	}
}

func TestFlagString_SingleBit(t *testing.T) {
	if got := FlagString(1<<0, RisflagNames); got != "fire" {
		t.Errorf("FlagString(fire) = %q, want fire", got)
	}
	if got := FlagString(1<<21, RisflagNames); got != "paralysis" {
		t.Errorf("FlagString(paralysis) = %q, want paralysis", got)
	}
}

func TestFlagString_MultipleBits(t *testing.T) {
	got := FlagString((1<<0)|(1<<1)|(1<<5), RisflagNames)
	want := "fire cold pierce"
	if got != want {
		t.Errorf("FlagString(fire|cold|pierce) = %q, want %q", got, want)
	}
}
