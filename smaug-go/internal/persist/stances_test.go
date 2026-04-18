package persist

import (
	"path/filepath"
	"testing"

	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/types"
)

// cloneStanceIndex makes a deep copy we can restore after each test so that
// LoadStancesInto side effects don't leak across subtests.
func cloneStanceIndex() [types.MAX_STANCE]combat.StanceInfo {
	var out [types.MAX_STANCE]combat.StanceInfo
	out = combat.StanceIndex
	return out
}

// withSavedStances runs fn against a saved-and-restored combat.StanceIndex.
func withSavedStances(t *testing.T, fn func()) {
	t.Helper()
	saved := cloneStanceIndex()
	t.Cleanup(func() { combat.StanceIndex = saved })
	fn()
}

func TestGetStanceNumber(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"Tiger", types.STANCE_TIGER},
		{"tiger", types.STANCE_TIGER},
		{"TIGER", types.STANCE_TIGER},
		{"Swallow", types.STANCE_SWALLOW},
		{"Dragon", types.STANCE_DRAGON},
		{"Monkey", types.STANCE_MONKEY},
		{"Mantis", types.STANCE_MANTIS},
		{"Viper", types.STANCE_VIPER},
		{"Crane", types.STANCE_CRANE},
		{"Crab", types.STANCE_CRAB},
		{"Mongoose", types.STANCE_MONGOOSE},
		{"Bull", types.STANCE_BULL},
		{"None", types.STANCE_NONE},
		{"Normal", types.STANCE_NORMAL},
		{"nonesuch", -1},
		{"", -1},
	}
	for _, tc := range cases {
		if got := GetStanceNumber(tc.in); got != tc.want {
			t.Errorf("GetStanceNumber(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestLoadStances_EmptyFileKeepsDefaults(t *testing.T) {
	withSavedStances(t, func() {
		before := cloneStanceIndex()
		path := filepath.Join("testdata", "stances_empty.dat")
		if err := LoadStancesInto(&combat.StanceIndex, path); err != nil {
			t.Fatalf("LoadStancesInto: %v", err)
		}
		if combat.StanceIndex != before {
			t.Errorf("empty file mutated StanceIndex; expected unchanged")
		}
	})
}

func TestLoadStances_TwoStancesOverridesDefaults(t *testing.T) {
	withSavedStances(t, func() {
		before := cloneStanceIndex()
		path := filepath.Join("testdata", "stances_two.dat")
		if err := LoadStancesInto(&combat.StanceIndex, path); err != nil {
			t.Fatalf("LoadStancesInto: %v", err)
		}
		dragon := combat.StanceIndex[types.STANCE_DRAGON]
		if dragon.NumAttacks != 2 {
			t.Errorf("Dragon NumAttacks = %d, want 2", dragon.NumAttacks)
		}
		if dragon.DamDone != 150 {
			t.Errorf("Dragon DamDone = %d, want 150", dragon.DamDone)
		}
		if dragon.DamTaken != 90 {
			t.Errorf("Dragon DamTaken = %d, want 90", dragon.DamTaken)
		}
		tiger := combat.StanceIndex[types.STANCE_TIGER]
		if tiger.NumAttacks != 3 {
			t.Errorf("Tiger NumAttacks = %d, want 3", tiger.NumAttacks)
		}
		if tiger.DamDone != 120 {
			t.Errorf("Tiger DamDone = %d, want 120", tiger.DamDone)
		}
		if tiger.DamTaken != 110 {
			t.Errorf("Tiger DamTaken = %d, want 110", tiger.DamTaken)
		}
		// Unmentioned stances retain defaults.
		if combat.StanceIndex[types.STANCE_CRAB] != before[types.STANCE_CRAB] {
			t.Errorf("Crab mutated; expected defaults preserved: got %+v want %+v",
				combat.StanceIndex[types.STANCE_CRAB], before[types.STANCE_CRAB])
		}
	})
}

func TestLoadStances_TabSeparatedFixture(t *testing.T) {
	withSavedStances(t, func() {
		path := filepath.Join("testdata", "stances_tabs.dat")
		if err := LoadStancesInto(&combat.StanceIndex, path); err != nil {
			t.Fatalf("LoadStancesInto: %v", err)
		}
		if got := combat.StanceIndex[types.STANCE_DRAGON].NumAttacks; got != 2 {
			t.Errorf("Dragon NumAttacks (tabs) = %d, want 2", got)
		}
		if got := combat.StanceIndex[types.STANCE_DRAGON].DamDone; got != 150 {
			t.Errorf("Dragon DamDone (tabs) = %d, want 150", got)
		}
	})
}

func TestLoadStances_UnknownKeywordIgnored(t *testing.T) {
	withSavedStances(t, func() {
		before := cloneStanceIndex()
		path := filepath.Join("testdata", "stances_unknown_key.dat")
		if err := LoadStancesInto(&combat.StanceIndex, path); err != nil {
			t.Fatalf("LoadStancesInto: %v", err)
		}
		// Tiger block was opened but only BogusKey appeared; combat fields
		// should match the defaults because no recognized key was hit.
		if combat.StanceIndex[types.STANCE_TIGER] != before[types.STANCE_TIGER] {
			t.Errorf("unknown-key fixture mutated Tiger; got %+v want %+v",
				combat.StanceIndex[types.STANCE_TIGER], before[types.STANCE_TIGER])
		}
	})
}

func TestLoadStances_BadStanceNameSkipsBlock(t *testing.T) {
	withSavedStances(t, func() {
		before := cloneStanceIndex()
		path := filepath.Join("testdata", "stances_bad_name.dat")
		if err := LoadStancesInto(&combat.StanceIndex, path); err != nil {
			t.Fatalf("LoadStancesInto: %v", err)
		}
		// Nothing overridden; every slot equals defaults.
		if combat.StanceIndex != before {
			t.Errorf("bad stance name mutated StanceIndex")
		}
	})
}

func TestLoadStances_FileMissingIsNotFatal(t *testing.T) {
	withSavedStances(t, func() {
		before := cloneStanceIndex()
		path := filepath.Join("testdata", "does_not_exist.dat")
		if err := LoadStancesInto(&combat.StanceIndex, path); err != nil {
			t.Fatalf("LoadStancesInto missing file: %v", err)
		}
		if combat.StanceIndex != before {
			t.Errorf("missing file mutated StanceIndex")
		}
	})
}

func TestLoadStances_NilTargetIsNoop(t *testing.T) {
	if err := LoadStancesInto(nil, "testdata/stances_two.dat"); err != nil {
		t.Errorf("nil target should be no-op, got %v", err)
	}
}

func TestLoadStances_RealShippedStubLoads(t *testing.T) {
	// The stub at db/system/stances.dat contains just "End\n"; loader
	// must accept it without error and without mutating the defaults.
	withSavedStances(t, func() {
		before := cloneStanceIndex()
		// Resolve repo-rooted path via relative walk from this test file.
		path := filepath.Join("..", "..", "..", "db", "system", "stances.dat")
		if err := LoadStancesInto(&combat.StanceIndex, path); err != nil {
			t.Fatalf("LoadStancesInto stub: %v", err)
		}
		if combat.StanceIndex != before {
			t.Errorf("stub mutated StanceIndex")
		}
	})
}
