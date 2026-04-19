package combat

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// G2 — UpdateStances. Covers A6 from plan-phase6-stances-olc.md.

func TestUpdateStances_EnteringAppliesResistImmuneSuscept(t *testing.T) {
	// Cast RIS constants (uint32) to int to match StanceInfo fields and
	// CharData.Stance* fields (both int).
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{
		Resist:  int(types.RIS_FIRE),
		Immune:  int(types.RIS_COLD),
		Suscept: int(types.RIS_ACID),
	})
	ch := &types.CharData{Stance: types.STANCE_DRAGON}
	UpdateStances(ch, true)
	if ch.StanceResistant != int(types.RIS_FIRE) {
		t.Errorf("StanceResistant = %#x, want %#x", ch.StanceResistant, types.RIS_FIRE)
	}
	if ch.StanceImmune != int(types.RIS_COLD) {
		t.Errorf("StanceImmune = %#x, want %#x", ch.StanceImmune, types.RIS_COLD)
	}
	if ch.StanceSusceptible != int(types.RIS_ACID) {
		t.Errorf("StanceSusceptible = %#x, want %#x", ch.StanceSusceptible, types.RIS_ACID)
	}
}

func TestUpdateStances_LeavingClearsResistImmuneSuscept_Idempotent(t *testing.T) {
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{
		Resist:  int(types.RIS_FIRE),
		Immune:  int(types.RIS_COLD),
		Suscept: int(types.RIS_ACID),
	})
	ch := &types.CharData{Stance: types.STANCE_DRAGON}
	UpdateStances(ch, true)
	UpdateStances(ch, false)
	if ch.StanceResistant != 0 || ch.StanceImmune != 0 || ch.StanceSusceptible != 0 {
		t.Errorf("after exit: Res=%#x Imm=%#x Sus=%#x; want all 0",
			ch.StanceResistant, ch.StanceImmune, ch.StanceSusceptible)
	}
	// Double-leave must stay at zero, no panic.
	UpdateStances(ch, false)
	if ch.StanceResistant != 0 || ch.StanceImmune != 0 || ch.StanceSusceptible != 0 {
		t.Error("idempotent leave violated")
	}
}

func TestUpdateStances_PreservesUnrelatedBits(t *testing.T) {
	// Pre-existing non-stance RIS bits (e.g., race resistances) must
	// survive UpdateStances exit. This verifies &^= semantics rather
	// than blanket zeroing.
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{Resist: int(types.RIS_FIRE)})
	ch := &types.CharData{Stance: types.STANCE_DRAGON}
	ch.StanceResistant = int(types.RIS_COLD) // some unrelated bit
	UpdateStances(ch, true)
	UpdateStances(ch, false)
	if ch.StanceResistant != int(types.RIS_COLD) {
		t.Errorf("unrelated RIS bit lost: got %#x want %#x",
			ch.StanceResistant, types.RIS_COLD)
	}
}

func TestUpdateStances_OutOfRangeSafe(t *testing.T) {
	ch := &types.CharData{Stance: types.MAX_STANCE + 1}
	// no panic, no mutation
	UpdateStances(ch, true)
	UpdateStances(ch, false)
	if ch.StanceResistant != 0 {
		t.Error("out-of-range stance should be no-op")
	}
}

func TestUpdateStances_NilCharSafe(t *testing.T) {
	UpdateStances(nil, true)
	UpdateStances(nil, false)
}
