package util

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// TestMProgFlagNames_HasAllCEntries asserts the table has 52 entries
// (indices 0..51) matching the C mprog_flags[] table at src/build.c:366-374.
func TestMProgFlagNames_HasAllCEntries(t *testing.T) {
	if got := len(MProgFlagNames); got != 52 {
		t.Fatalf("len(MProgFlagNames) = %d, want 52", got)
	}
	// Spot-check a handful of indices that pin the C order.
	cases := []struct {
		idx  int
		name string
	}{
		{0, "act"},
		{1, "speech"},
		{2, "rand"},
		{3, "fight"},
		{4, "death"},
		{5, "hitprcnt"},
		{6, "entry"},
		{7, "greet"},
		{8, "allgreet"},
		{15, "sac"},
		{16, "look"},
		{31, "use"},
		{32, "load"},
		{33, "login"},
		{34, "void"},
		{35, "tell"},
		{36, "imminfo"},
		{37, "greetinfight"},
		{38, "move"},
		{39, "command"},
		{40, "sell"},
		{41, "emote"},
		{42, "r1"},
		{51, "r10"},
	}
	for _, c := range cases {
		if MProgFlagNames[c.idx] != c.name {
			t.Errorf("MProgFlagNames[%d] = %q, want %q", c.idx, MProgFlagNames[c.idx], c.name)
		}
	}
}

func TestGetMpFlag_KnownNames(t *testing.T) {
	cases := []struct {
		name string
		want int64
	}{
		{"act", types.MPROG_ACT},
		{"speech", types.MPROG_SPEECH},
		{"rand", types.MPROG_RAND},
		{"fight", types.MPROG_FIGHT},
		{"death", types.MPROG_DEATH},
		{"hitprcnt", types.MPROG_HITPRCNT},
		{"entry", types.MPROG_ENTRY},
		{"greet", types.MPROG_GREET},
		{"allgreet", types.MPROG_ALL_GREET},
		{"all_greet", types.MPROG_ALL_GREET},
		{"give", types.MPROG_GIVE},
		{"bribe", types.MPROG_BRIBE},
		{"hour", types.MPROG_HOUR},
		{"time", types.MPROG_TIME},
		{"wear", types.MPROG_WEAR},
		{"remove", types.MPROG_REMOVE},
		{"sac", types.MPROG_SAC},
		{"look", types.MPROG_LOOK},
		{"exa", types.MPROG_EXA},
		{"examine", types.MPROG_EXA},
		{"zap", types.MPROG_ZAP},
		{"get", types.MPROG_GET},
		{"drop", types.MPROG_DROP},
		{"damage", types.MPROG_DAMAGE},
		{"repair", types.MPROG_REPAIR},
		{"randiw", types.MPROG_RANDIW},
		{"speechiw", types.MPROG_SPEECHIW},
		{"pull", types.MPROG_PULL},
		{"push", types.MPROG_PUSH},
		{"sleep", types.MPROG_SLEEP},
		{"rest", types.MPROG_REST},
		{"leave", types.MPROG_LEAVE},
		{"script", types.MPROG_SCRIPT},
		{"use", types.MPROG_USE},
		{"login", types.MPROG_LOGIN},
		{"void", types.MPROG_VOID},
		{"tell", types.MPROG_TELL},
		{"imminfo", types.MPROG_IMMINFO},
		{"command", types.MPROG_CMD},
		{"cmd", types.MPROG_CMD},
		{"sell", types.MPROG_SELL},
		{"enter", types.MPROG_ENTER}, // alias of entry
	}
	for _, c := range cases {
		got, ok := GetMpFlag(c.name)
		if !ok {
			t.Errorf("GetMpFlag(%q): ok=false, want true", c.name)
			continue
		}
		if got != c.want {
			t.Errorf("GetMpFlag(%q) = 0x%x, want 0x%x", c.name, got, c.want)
		}
	}
}

func TestGetMpFlag_StripSuffix(t *testing.T) {
	for _, in := range []string{"act_prog", "actprog", "ACT_PROG", "Act_Prog"} {
		got, ok := GetMpFlag(in)
		if !ok || got != types.MPROG_ACT {
			t.Errorf("GetMpFlag(%q) = (0x%x, %v), want (0x%x, true)", in, got, ok, types.MPROG_ACT)
		}
	}
}

func TestGetMpFlag_Unknown(t *testing.T) {
	// Empty, garbage, and unported (load / greetinfight / move / emote / r*).
	for _, in := range []string{"", "  ", "bogus", "xyzzy", "load", "greetinfight", "move", "emote", "r1", "r10"} {
		got, ok := GetMpFlag(in)
		if ok || got != 0 {
			t.Errorf("GetMpFlag(%q) = (0x%x, %v), want (0, false)", in, got, ok)
		}
	}
}

func TestGetMpFlag_CaseInsensitive(t *testing.T) {
	for _, in := range []string{"act", "ACT", "Act", "aCt"} {
		got, ok := GetMpFlag(in)
		if !ok || got != types.MPROG_ACT {
			t.Errorf("GetMpFlag(%q) = (0x%x, %v), want (0x%x, true)", in, got, ok, types.MPROG_ACT)
		}
	}
}
