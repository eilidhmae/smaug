package persist

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/types"
)

// G3 — SaveStances. Covers A9/A10/A11 from plan-phase6-stances-olc.md.

func saveToBuffer(t *testing.T, target *[types.MAX_STANCE]combat.StanceInfo) string {
	t.Helper()
	var buf bytes.Buffer
	if err := writeStances(&buf, target); err != nil {
		t.Fatalf("writeStances: %v", err)
	}
	return buf.String()
}

func TestSaveStances_EmptyTable_WritesHeaderAndEnd(t *testing.T) {
	var empty [types.MAX_STANCE]combat.StanceInfo
	out := saveToBuffer(t, &empty)
	// Must end with "End\n".
	if !strings.HasSuffix(out, "End\n") {
		t.Errorf("output does not end with 'End\\n': tail=%q", out[len(out)-10:])
	}
	// Must have 12 StartStance lines (one per MAX_STANCE slot).
	count := strings.Count(out, "StartStance")
	if count != types.MAX_STANCE {
		t.Errorf("StartStance count = %d, want %d", count, types.MAX_STANCE)
	}
	// Each stance block is StartStance then EndStance then blank line.
	if !strings.Contains(out, "EndStance\n\n") {
		t.Errorf("output missing EndStance+blank-line separator")
	}
	// No optional keys on a fully-default table.
	// Prefix each key with "\n" to avoid matching "StartStance\t" as "Stance\t".
	for _, k := range []string{"Attacks\t", "Class\t", "DamDone\t", "DamTaken\t", "Dodge\t",
		"Dual\t", "Immune\t", "Other\t", "Parry\t", "Percent\t",
		"Race\t", "Resist\t", "Self\t", "Special\t", "Stance\t",
		"Suscept\t", "Wait\t", "Weight\t",
	} {
		if strings.Contains(out, "\n"+k) {
			t.Errorf("default table should NOT emit %q; output=\n%s", k, out)
		}
	}
}

func TestSaveStances_FullyPopulatedStance_EmitsAllKeys(t *testing.T) {
	var tbl [types.MAX_STANCE]combat.StanceInfo
	tbl[types.STANCE_DRAGON] = combat.StanceInfo{
		NumAttacks: 2, Class: 7, DamDone: 150, DamTaken: 90,
		Dodge: 10, Dual: 1, Immune: 5,
		Others:         "enter room",
		Parry:          5,
		SpecialPercent: 25, Race: 3, Resist: 3,
		Self: "enter self", SpecialMove: 1,
		Prereq:  [2]int{types.STANCE_TIGER, 0},
		Suscept: 2, Wait: 12, MaxWeight: 500,
	}
	out := saveToBuffer(t, &tbl)
	// Extract the Dragon block.
	dragonStart := strings.Index(out, "StartStance\tDragon")
	if dragonStart < 0 {
		t.Fatal("Dragon block not found")
	}
	dragonEnd := strings.Index(out[dragonStart:], "EndStance")
	dragonBlock := out[dragonStart : dragonStart+dragonEnd]

	wantKeys := []string{
		"Attacks\t2\n",
		"Class\t7\n",
		"DamDone\t150\n",
		"DamTaken\t90\n",
		"Dodge\t10\n",
		"Dual\t1\n",
		"Immune\t5\n",
		"Other\tenter room~\n",
		"Parry\t5\n",
		"Percent\t25\n",
		"Race\t3\n",
		"Resist\t3\n",
		"Self\tenter self~\n",
		"Special\tNone\n",
		"Stance\tTiger None\n",
		"Suscept\t2\n",
		"Wait\t12\n",
		"Weight\t500\n",
	}
	for _, k := range wantKeys {
		if !strings.Contains(dragonBlock, k) {
			t.Errorf("Dragon block missing %q; block=\n%s", k, dragonBlock)
		}
	}
}

func TestSaveStances_ConditionalWriteInvariants(t *testing.T) {
	cases := []struct {
		name   string
		set    func(*combat.StanceInfo)
		key    string
		expect bool
	}{
		// NumAttacks: != 0
		{"NumAttacks zero → no Attacks", func(s *combat.StanceInfo) { s.NumAttacks = 0 }, "Attacks\t", false},
		{"NumAttacks negative → emits Attacks", func(s *combat.StanceInfo) { s.NumAttacks = -3 }, "Attacks\t", true},
		{"NumAttacks positive → emits Attacks", func(s *combat.StanceInfo) { s.NumAttacks = 3 }, "Attacks\t", true},
		// Class: > 0
		{"Class zero → no Class", func(s *combat.StanceInfo) { s.Class = 0 }, "Class\t", false},
		{"Class negative → no Class", func(s *combat.StanceInfo) { s.Class = -1 }, "Class\t", false},
		{"Class positive → emits Class", func(s *combat.StanceInfo) { s.Class = 1 }, "Class\t", true},
		// DamDone: > 0
		{"DamDone zero → no", func(s *combat.StanceInfo) { s.DamDone = 0 }, "DamDone\t", false},
		{"DamDone negative → no", func(s *combat.StanceInfo) { s.DamDone = -1 }, "DamDone\t", false},
		{"DamDone positive → yes", func(s *combat.StanceInfo) { s.DamDone = 50 }, "DamDone\t", true},
		// Dodge: != 0
		{"Dodge zero → no", func(s *combat.StanceInfo) { s.Dodge = 0 }, "Dodge\t", false},
		{"Dodge negative → yes", func(s *combat.StanceInfo) { s.Dodge = -5 }, "Dodge\t", true},
		{"Dodge positive → yes", func(s *combat.StanceInfo) { s.Dodge = 5 }, "Dodge\t", true},
		// Dual: != 0
		{"Dual zero → no", func(s *combat.StanceInfo) { s.Dual = 0 }, "Dual\t", false},
		{"Dual 1 → yes", func(s *combat.StanceInfo) { s.Dual = 1 }, "Dual\t", true},
		// Immune: > 0
		{"Immune 0 → no", func(s *combat.StanceInfo) { s.Immune = 0 }, "Immune\t", false},
		{"Immune positive → yes", func(s *combat.StanceInfo) { s.Immune = 5 }, "Immune\t", true},
		// Others: != ""
		{"Others empty → no", func(s *combat.StanceInfo) { s.Others = "" }, "Other\t", false},
		{"Others nonempty → yes", func(s *combat.StanceInfo) { s.Others = "x" }, "Other\t", true},
		// Parry: != 0
		{"Parry 0 → no", func(s *combat.StanceInfo) { s.Parry = 0 }, "Parry\t", false},
		{"Parry nonzero → yes", func(s *combat.StanceInfo) { s.Parry = -3 }, "Parry\t", true},
		// SpecialPercent: > 0
		{"SpecialPercent 0 → no", func(s *combat.StanceInfo) { s.SpecialPercent = 0 }, "Percent\t", false},
		{"SpecialPercent positive → yes", func(s *combat.StanceInfo) { s.SpecialPercent = 50 }, "Percent\t", true},
		// Race: > 0
		{"Race 0 → no", func(s *combat.StanceInfo) { s.Race = 0 }, "Race\t", false},
		{"Race positive → yes", func(s *combat.StanceInfo) { s.Race = 3 }, "Race\t", true},
		// Resist: > 0
		{"Resist 0 → no", func(s *combat.StanceInfo) { s.Resist = 0 }, "Resist\t", false},
		{"Resist positive → yes", func(s *combat.StanceInfo) { s.Resist = 1 }, "Resist\t", true},
		// Self: != ""
		{"Self empty → no", func(s *combat.StanceInfo) { s.Self = "" }, "Self\t", false},
		{"Self nonempty → yes", func(s *combat.StanceInfo) { s.Self = "x" }, "Self\t", true},
		// SpecialMove: > 0
		{"SpecialMove 0 → no", func(s *combat.StanceInfo) { s.SpecialMove = 0 }, "Special\t", false},
		{"SpecialMove positive → yes", func(s *combat.StanceInfo) { s.SpecialMove = 1 }, "Special\t", true},
		// Prereq[0]: > 0
		{"Prereq[0] 0 → no Stance", func(s *combat.StanceInfo) { s.Prereq = [2]int{0, 0} }, "Stance\t", false},
		{"Prereq[0] positive → yes Stance", func(s *combat.StanceInfo) { s.Prereq = [2]int{types.STANCE_TIGER, 0} }, "Stance\t", true},
		// Suscept: > 0
		{"Suscept 0 → no", func(s *combat.StanceInfo) { s.Suscept = 0 }, "Suscept\t", false},
		{"Suscept positive → yes", func(s *combat.StanceInfo) { s.Suscept = 1 }, "Suscept\t", true},
		// Wait: > 0
		{"Wait 0 → no", func(s *combat.StanceInfo) { s.Wait = 0 }, "Wait\t", false},
		{"Wait positive → yes", func(s *combat.StanceInfo) { s.Wait = 5 }, "Wait\t", true},
		// MaxWeight: > 0
		{"MaxWeight 0 → no Weight", func(s *combat.StanceInfo) { s.MaxWeight = 0 }, "Weight\t", false},
		{"MaxWeight positive → yes Weight", func(s *combat.StanceInfo) { s.MaxWeight = 100 }, "Weight\t", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var tbl [types.MAX_STANCE]combat.StanceInfo
			tc.set(&tbl[types.STANCE_DRAGON])
			out := saveToBuffer(t, &tbl)
			// Extract dragon block.
			start := strings.Index(out, "StartStance\tDragon")
			end := start + strings.Index(out[start:], "EndStance")
			block := out[start:end]
			// Prefix the lookup with a newline so "Stance\t" doesn't match
			// "StartStance\t".
			has := strings.Contains(block, "\n"+tc.key)
			if has != tc.expect {
				t.Errorf("key %q presence=%v, want %v; block=\n%s", tc.key, has, tc.expect, block)
			}
		})
	}
}

func TestSaveStances_RoundTripMatchesLoad(t *testing.T) {
	withSavedStances(t, func() {
		savedPath := StancePath
		t.Cleanup(func() { StancePath = savedPath })
		// Load the full fixture.
		path := filepath.Join("testdata", "stances_full.dat")
		if err := LoadStancesInto(&combat.StanceIndex, path); err != nil {
			t.Fatalf("load: %v", err)
		}
		// Snapshot the loaded table.
		loaded := combat.StanceIndex

		// Save to a buffer, then reload into a fresh table.
		var buf bytes.Buffer
		if err := writeStances(&buf, &combat.StanceIndex); err != nil {
			t.Fatalf("write: %v", err)
		}
		var fresh [types.MAX_STANCE]combat.StanceInfo
		sc := NewScanner(&buf, "round-trip")
		for {
			word := sc.ReadWord()
			if word == "" || word == "End" {
				break
			}
			if word == "StartStance" {
				readStanceBlock(sc, &fresh)
			}
		}

		// Compare the DRAGON and TIGER slots byte-for-byte.
		for _, idx := range []int{types.STANCE_DRAGON, types.STANCE_TIGER} {
			if !reflect.DeepEqual(loaded[idx], fresh[idx]) {
				t.Errorf("round-trip diverged at stance %d:\nloaded=%+v\nfresh=%+v",
					idx, loaded[idx], fresh[idx])
			}
		}
	})
}

func TestSaveStances_StancePrereqNames_BothWritten(t *testing.T) {
	var tbl [types.MAX_STANCE]combat.StanceInfo
	tbl[types.STANCE_DRAGON] = combat.StanceInfo{
		Prereq: [2]int{types.STANCE_TIGER, types.STANCE_SWALLOW},
	}
	out := saveToBuffer(t, &tbl)
	if !strings.Contains(out, "Stance\tTiger Swallow\n") {
		t.Errorf("expected 'Stance\\tTiger Swallow\\n' in output; got:\n%s", out)
	}
}

func TestSaveStances_StancePrereqSecondaryZero_WritesNone(t *testing.T) {
	var tbl [types.MAX_STANCE]combat.StanceInfo
	tbl[types.STANCE_DRAGON] = combat.StanceInfo{
		Prereq: [2]int{types.STANCE_TIGER, 0},
	}
	out := saveToBuffer(t, &tbl)
	if !strings.Contains(out, "Stance\tTiger None\n") {
		t.Errorf("expected 'Stance\\tTiger None\\n' in output; got:\n%s", out)
	}
}

func TestSaveStances_NilTargetRejected(t *testing.T) {
	if err := SaveStances(nil, "/tmp/unused.dat"); err == nil {
		t.Error("SaveStances(nil, ...) should return an error")
	}
}

func TestSaveStances_WritesToPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "stances.dat")
	var tbl [types.MAX_STANCE]combat.StanceInfo
	tbl[types.STANCE_DRAGON] = combat.StanceInfo{NumAttacks: 7}
	if err := SaveStances(&tbl, path); err != nil {
		t.Fatalf("SaveStances: %v", err)
	}
	// Load it back.
	var got [types.MAX_STANCE]combat.StanceInfo
	if err := LoadStancesInto(&got, path); err != nil {
		t.Fatalf("LoadStancesInto: %v", err)
	}
	if got[types.STANCE_DRAGON].NumAttacks != 7 {
		t.Errorf("round-trip via disk: got Attacks=%d, want 7", got[types.STANCE_DRAGON].NumAttacks)
	}
}

// TestSaveStances_FileMode pins 0600 — security adversary 2026-04-19.
func TestSaveStances_FileMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stances.dat")
	var tbl [types.MAX_STANCE]combat.StanceInfo
	if err := SaveStances(&tbl, path); err != nil {
		t.Fatalf("SaveStances: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("stances.dat mode = %#o, want 0o600", mode)
	}
}
