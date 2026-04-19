package combat

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// G2 — CanUseStance. Covers A5 from plan-phase6-stances-olc.md.
//
// Helper: mirror the existing withStanceIndex pattern in stance_test.go.

func TestCanUseStance_NPC_IndexDataStancesPositive_TRUE(t *testing.T) {
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{})
	npc := &types.CharData{}
	npc.Act.Set(types.ACT_IS_NPC)
	npc.IndexData = &types.MobIndexData{}
	npc.IndexData.Stances[types.STANCE_DRAGON] = 1
	if !CanUseStance(npc, types.STANCE_DRAGON) {
		t.Error("expected TRUE for NPC with positive IndexData.Stances")
	}
}

func TestCanUseStance_NPC_IndexDataStancesZero_FALSE(t *testing.T) {
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{})
	npc := &types.CharData{}
	npc.Act.Set(types.ACT_IS_NPC)
	npc.IndexData = &types.MobIndexData{}
	// Stances[DRAGON] defaults to 0
	if CanUseStance(npc, types.STANCE_DRAGON) {
		t.Error("expected FALSE for NPC with zero IndexData.Stances")
	}
}

func TestCanUseStance_NPC_NilIndexData_FALSE(t *testing.T) {
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{})
	npc := &types.CharData{}
	npc.Act.Set(types.ACT_IS_NPC)
	if CanUseStance(npc, types.STANCE_DRAGON) {
		t.Error("expected FALSE for NPC with nil IndexData")
	}
}

func TestCanUseStance_PC_NoPrereq_TRUE(t *testing.T) {
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{}) // Prereq[0] = 0
	pc := &types.CharData{PCData: &types.PCData{}}
	if !CanUseStance(pc, types.STANCE_DRAGON) {
		t.Error("PC with no-prereq stance should succeed")
	}
}

func TestCanUseStance_PC_Prereq0_MasteredAt200_TRUE(t *testing.T) {
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{Prereq: [2]int{types.STANCE_TIGER, 0}})
	pc := &types.CharData{PCData: &types.PCData{}}
	pc.PCData.Stances[types.STANCE_TIGER] = 200
	if !CanUseStance(pc, types.STANCE_DRAGON) {
		t.Error("PC with TIGER=200 should get DRAGON")
	}
}

func TestCanUseStance_PC_Prereq0_Below200_FALSE(t *testing.T) {
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{Prereq: [2]int{types.STANCE_TIGER, 0}})
	pc := &types.CharData{PCData: &types.PCData{}}
	pc.PCData.Stances[types.STANCE_TIGER] = 199
	if CanUseStance(pc, types.STANCE_DRAGON) {
		t.Error("PC with TIGER=199 should NOT get DRAGON")
	}
}

func TestCanUseStance_PC_Prereq0and1_BothMastered_TRUE(t *testing.T) {
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{Prereq: [2]int{types.STANCE_TIGER, types.STANCE_SWALLOW}})
	pc := &types.CharData{PCData: &types.PCData{}}
	pc.PCData.Stances[types.STANCE_TIGER] = 200
	pc.PCData.Stances[types.STANCE_SWALLOW] = 200
	if !CanUseStance(pc, types.STANCE_DRAGON) {
		t.Error("both prereqs GM should succeed")
	}
}

func TestCanUseStance_PC_Prereq0Mastered_Prereq1NotMastered_FALSE(t *testing.T) {
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{Prereq: [2]int{types.STANCE_TIGER, types.STANCE_SWALLOW}})
	pc := &types.CharData{PCData: &types.PCData{}}
	pc.PCData.Stances[types.STANCE_TIGER] = 200
	pc.PCData.Stances[types.STANCE_SWALLOW] = 150
	if CanUseStance(pc, types.STANCE_DRAGON) {
		t.Error("only Prereq[0] at GM should fail")
	}
}

func TestCanUseStance_MaxWeightExceeded_FALSE(t *testing.T) {
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{MaxWeight: 100})
	pc := &types.CharData{PCData: &types.PCData{}, CarryWeight: 200}
	if CanUseStance(pc, types.STANCE_DRAGON) {
		t.Error("carry_weight > max_weight should fail")
	}
}

func TestCanUseStance_MaxWeightEqual_TRUE_BoundaryOff(t *testing.T) {
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{MaxWeight: 100})
	pc := &types.CharData{PCData: &types.PCData{}, CarryWeight: 100}
	// C uses `>`, not `>=`.
	if !CanUseStance(pc, types.STANCE_DRAGON) {
		t.Error("carry_weight == max_weight should pass (C uses > not >=)")
	}
}

func TestCanUseStance_DualForbidWithDualWielded_FALSE(t *testing.T) {
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{Dual: 1})
	pc := &types.CharData{PCData: &types.PCData{}}
	obj := &types.ObjData{WearLoc: types.WEAR_DUAL_WIELD}
	pc.Carrying = append(pc.Carrying, obj)
	if CanUseStance(pc, types.STANCE_DRAGON) {
		t.Error("Dual=1 with WEAR_DUAL_WIELD equipped should fail")
	}
}

func TestCanUseStance_DualForbidWithoutDualWielded_TRUE(t *testing.T) {
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{Dual: 1})
	pc := &types.CharData{PCData: &types.PCData{}}
	if !CanUseStance(pc, types.STANCE_DRAGON) {
		t.Error("Dual=1 without dual-wield equipped should pass")
	}
}

func TestCanUseStance_DualAllowWithDualWielded_TRUE(t *testing.T) {
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{Dual: 0})
	pc := &types.CharData{PCData: &types.PCData{}}
	obj := &types.ObjData{WearLoc: types.WEAR_DUAL_WIELD}
	pc.Carrying = append(pc.Carrying, obj)
	if !CanUseStance(pc, types.STANCE_DRAGON) {
		t.Error("Dual=0 should ignore dual-wield")
	}
}

// C bug: IS_SET(Class, ch.Class) = Class & ch.Class — not Class & (1<<index).
// Pinned to guard the preserve-C-behaviour decision (Q4).
func TestCanUseStance_ClassRestriction_CBug_PreservesMaskANDIndex(t *testing.T) {
	// Case A: mask=1, index=1 → 1&1=1 ≠ 0 → blocked.
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{Class: 1})
	blocked := &types.CharData{PCData: &types.PCData{}, Class: 1}
	if CanUseStance(blocked, types.STANCE_DRAGON) {
		t.Error("Class=1, ch.Class=1 should be blocked by C's mask&index quirk")
	}
	// Case B: mask=2, index=1 → 2&1=0 → allowed.
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{Class: 2})
	allowed := &types.CharData{PCData: &types.PCData{}, Class: 1}
	if !CanUseStance(allowed, types.STANCE_DRAGON) {
		t.Error("Class=2, ch.Class=1 should be allowed (2&1==0)")
	}
}

func TestCanUseStance_RaceRestriction_CBug_SameAsClass(t *testing.T) {
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{Race: 4})
	blocked := &types.CharData{PCData: &types.PCData{}, Race: 4}
	if CanUseStance(blocked, types.STANCE_DRAGON) {
		t.Error("Race=4, ch.Race=4 should be blocked")
	}
	allowed := &types.CharData{PCData: &types.PCData{}, Race: 1}
	if !CanUseStance(allowed, types.STANCE_DRAGON) {
		t.Error("Race=4, ch.Race=1 should be allowed (4&1==0)")
	}
}

func TestCanUseStance_NilChar_FALSE(t *testing.T) {
	if CanUseStance(nil, types.STANCE_DRAGON) {
		t.Error("nil ch should return FALSE")
	}
}

func TestCanUseStance_OutOfRangeStance_FALSE(t *testing.T) {
	pc := &types.CharData{PCData: &types.PCData{}}
	if CanUseStance(pc, types.MAX_STANCE+1) {
		t.Error("out-of-range new stance should return FALSE")
	}
	if CanUseStance(pc, -1) {
		t.Error("negative new stance should return FALSE")
	}
}
