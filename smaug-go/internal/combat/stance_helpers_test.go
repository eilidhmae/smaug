package combat

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// G1 — StanceInfo extension + helpers
// Covers A1, A3, A4 from plan-phase6-stances-olc.md.

func TestStanceInfo_ZeroValueCompatibleWithDefaults(t *testing.T) {
	// Extending StanceInfo with new fields must not change the default
	// StanceIndex combat values (Tranche B regression guard).
	want := map[int]StanceInfo{
		types.STANCE_DRAGON:   {NumAttacks: 1, DamDone: 120, DamTaken: 100},
		types.STANCE_MONKEY:   {NumAttacks: 0, DamDone: 0, DamTaken: 0},
		types.STANCE_SWALLOW:  {NumAttacks: 1, DamDone: 100, DamTaken: 100},
		types.STANCE_MONGOOSE: {NumAttacks: 1, DamDone: 100, DamTaken: 100},
	}
	for idx, w := range want {
		got := StanceIndex[idx]
		if got.NumAttacks != w.NumAttacks || got.DamDone != w.DamDone || got.DamTaken != w.DamTaken {
			t.Errorf("StanceIndex[%d] combat tuple = (%d,%d,%d), want (%d,%d,%d)",
				idx, got.NumAttacks, got.DamDone, got.DamTaken,
				w.NumAttacks, w.DamDone, w.DamTaken)
		}
		// New fields must default to zero on the package-level table.
		if got.Dodge != 0 || got.Parry != 0 || got.Dual != 0 || got.MaxWeight != 0 ||
			got.Wait != 0 || got.Resist != 0 || got.Immune != 0 || got.Suscept != 0 ||
			got.Class != 0 || got.Race != 0 || got.SpecialMove != 0 ||
			got.SpecialPercent != 0 || got.Prereq[0] != 0 || got.Prereq[1] != 0 ||
			got.Self != "" || got.Others != "" {
			t.Errorf("StanceIndex[%d] has non-zero extended fields: %+v", idx, got)
		}
	}
}

func TestGetStanceName_Exhaustive(t *testing.T) {
	cases := []struct {
		idx  int
		want string
	}{
		{types.STANCE_NONE, "None"},
		{types.STANCE_NORMAL, "Normal"},
		{types.STANCE_VIPER, "Viper"},
		{types.STANCE_CRANE, "Crane"},
		{types.STANCE_CRAB, "Crab"},
		{types.STANCE_MONGOOSE, "Mongoose"},
		{types.STANCE_BULL, "Bull"},
		{types.STANCE_MANTIS, "Mantis"},
		{types.STANCE_DRAGON, "Dragon"},
		{types.STANCE_TIGER, "Tiger"},
		{types.STANCE_MONKEY, "Monkey"},
		{types.STANCE_SWALLOW, "Swallow"},
		{types.MAX_STANCE, ""},
		{-1, ""},
		{9999, ""},
	}
	for _, tc := range cases {
		if got := GetStanceName(tc.idx); got != tc.want {
			t.Errorf("GetStanceName(%d) = %q, want %q", tc.idx, got, tc.want)
		}
	}
}

func TestGetStanceMastery_NoneReturnsZero(t *testing.T) {
	ch := &types.CharData{Stance: types.STANCE_NONE}
	ch.PCData = &types.PCData{}
	ch.PCData.Stances[types.STANCE_DRAGON] = 200
	if got := GetStanceMastery(ch); got != 0 {
		t.Errorf("STANCE_NONE mastery = %d, want 0", got)
	}
}

func TestGetStanceMastery_PC_UsesPCData(t *testing.T) {
	ch := &types.CharData{Stance: types.STANCE_DRAGON}
	ch.PCData = &types.PCData{}
	ch.PCData.Stances[types.STANCE_DRAGON] = 150
	if got := GetStanceMastery(ch); got != 150 {
		t.Errorf("PC DRAGON mastery = %d, want 150", got)
	}
}

func TestGetStanceMastery_NPC_UsesIndexData(t *testing.T) {
	ch := &types.CharData{Stance: types.STANCE_DRAGON}
	ch.Act.Set(types.ACT_IS_NPC)
	ch.IndexData = &types.MobIndexData{}
	ch.IndexData.Stances[types.STANCE_DRAGON] = 180
	if got := GetStanceMastery(ch); got != 180 {
		t.Errorf("NPC DRAGON mastery = %d, want 180", got)
	}
}

func TestGetStanceMastery_NilIndexDataNilPCDataSafe(t *testing.T) {
	npc := &types.CharData{Stance: types.STANCE_DRAGON}
	npc.Act.Set(types.ACT_IS_NPC)
	// npc.IndexData == nil
	if got := GetStanceMastery(npc); got != 0 {
		t.Errorf("NPC nil IndexData = %d, want 0", got)
	}

	pc := &types.CharData{Stance: types.STANCE_DRAGON}
	// pc.PCData == nil
	if got := GetStanceMastery(pc); got != 0 {
		t.Errorf("PC nil PCData = %d, want 0", got)
	}

	if got := GetStanceMastery(nil); got != 0 {
		t.Errorf("nil ch = %d, want 0", got)
	}
}

func TestGetStanceMastery_OutOfRangeSafe(t *testing.T) {
	ch := &types.CharData{Stance: types.MAX_STANCE + 5}
	ch.PCData = &types.PCData{}
	if got := GetStanceMastery(ch); got != 0 {
		t.Errorf("out-of-range stance = %d, want 0", got)
	}
}

func TestGetSpecialName_AlwaysNone(t *testing.T) {
	for _, n := range []int{0, 1, 99, -1} {
		if got := GetSpecialName(n); got != "None" {
			t.Errorf("GetSpecialName(%d) = %q, want None", n, got)
		}
	}
}

func TestGetSpecialNumber_AlwaysZero(t *testing.T) {
	for _, n := range []string{"", "None", "fireball", "Whatever"} {
		if got := GetSpecialNumber(n); got != 0 {
			t.Errorf("GetSpecialNumber(%q) = %d, want 0", n, got)
		}
	}
}
