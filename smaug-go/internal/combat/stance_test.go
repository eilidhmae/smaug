package combat

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

// --- G7 — Stance application in combat ---

// Install a synthetic StanceIndex entry for the test (preserves the
// real defaults for unrelated stances).
func withStanceIndex(t *testing.T, stance int, info StanceInfo) {
	t.Helper()
	saved := StanceIndex[stance]
	t.Cleanup(func() { StanceIndex[stance] = saved })
	StanceIndex[stance] = info
}

// PC with Stance == DRAGON, mastery >= STANCE_GRAND_MASTER (200), and
// NumAttacks=1 for DRAGON → one extra OneHit before the cascade.
func TestMultiHit_StanceGM_BonusAttacks(t *testing.T) {
	stubGsnsForCascade(t)
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{NumAttacks: 1, DamDone: 100, DamTaken: 100})
	calls := stubCascade(t, func() int { return 100 }) // no cascade firing
	_, _, ch, victim := setupCascade(t, 30)
	ch.Stance = types.STANCE_DRAGON
	ch.PCData.Stances[types.STANCE_DRAGON] = types.STANCE_GRAND_MASTER

	MultiHit(nil, ch, victim, types.TYPE_UNDEFINED)

	// Primary (1) + Dragon GM bonus (1) = 2.
	if *calls != 2 {
		t.Errorf("GM Dragon stance: OneHit calls = %d, want 2", *calls)
	}
}

// Same as above but mastery is below GM: no bonus.
func TestMultiHit_StanceNonGM_NoBonus(t *testing.T) {
	stubGsnsForCascade(t)
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{NumAttacks: 1, DamDone: 100, DamTaken: 100})
	calls := stubCascade(t, func() int { return 100 })
	_, _, ch, victim := setupCascade(t, 30)
	ch.Stance = types.STANCE_DRAGON
	ch.PCData.Stances[types.STANCE_DRAGON] = types.STANCE_GRAND_MASTER - 1 // 199

	MultiHit(nil, ch, victim, types.TYPE_UNDEFINED)

	if *calls != 1 {
		t.Errorf("non-GM Dragon: OneHit calls = %d, want 1 (no bonus)", *calls)
	}
}

// Either side in STANCE_MONKEY suppresses the GM bonus.
func TestMultiHit_MonkeyStanceSuppressesGMBonus(t *testing.T) {
	stubGsnsForCascade(t)
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{NumAttacks: 1, DamDone: 100, DamTaken: 100})
	calls := stubCascade(t, func() int { return 100 })
	_, _, ch, victim := setupCascade(t, 30)
	ch.Stance = types.STANCE_DRAGON
	ch.PCData.Stances[types.STANCE_DRAGON] = types.STANCE_GRAND_MASTER
	victim.Stance = types.STANCE_MONKEY

	MultiHit(nil, ch, victim, types.TYPE_UNDEFINED)

	if *calls != 1 {
		t.Errorf("victim monkey: OneHit calls = %d, want 1 (bonus suppressed)", *calls)
	}
}

// NPC stance adds to NumAttacks (C fight.c:1044).
func TestMultiHit_NPCStanceAddsNumAttacks(t *testing.T) {
	stubGsnsForCascade(t)
	withStanceIndex(t, types.STANCE_TIGER, StanceInfo{NumAttacks: 2, DamDone: 100, DamTaken: 100})
	calls := stubCascade(t, func() int { return 100 })
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 10200, Name: "NPC Arena"}
	w.Rooms[10200] = room

	ch := newFighter("NPC", 30)
	ch.Act.Set(types.ACT_IS_NPC)
	ch.NumAttacks = 1
	ch.Stance = types.STANCE_TIGER
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	victim := newFighter("Victim", 1)
	victim.Act.Set(types.ACT_IS_NPC)
	victim.Hit = 1_000_000
	victim.MaxHit = 1_000_000
	handler.CharToRoom(victim, room)
	w.AddChar(victim)
	StartFighting(ch, victim)

	MultiHit(w, ch, victim, types.TYPE_UNDEFINED)

	// Primary + tiger's 2 extra = 3 total.
	if *calls != 3 {
		t.Errorf("NPC tiger stance: OneHit calls = %d, want 3", *calls)
	}
}

// --- applyStanceDamage unit tests ---

// Monkey on either side is a no-op.
func TestApplyStanceDamage_MonkeyNoOp(t *testing.T) {
	ch := &types.CharData{Stance: types.STANCE_MONKEY}
	victim := &types.CharData{Stance: types.STANCE_DRAGON}
	got := applyStanceDamage(ch, victim, 100)
	if got != 100 {
		t.Errorf("ch STANCE_MONKEY: dam = %d, want 100 (no change)", got)
	}

	ch2 := &types.CharData{Stance: types.STANCE_DRAGON}
	victim2 := &types.CharData{Stance: types.STANCE_MONKEY}
	got = applyStanceDamage(ch2, victim2, 100)
	if got != 100 {
		t.Errorf("victim STANCE_MONKEY: dam = %d, want 100 (no change)", got)
	}
}

// Attacker stance with DamDone=120 and mastery=200:
//
//	dam *= 120/100 = 120
//	temp = 200/200 = 1.0, but min 0.5 → eff=200 in our integer form
//	dam *= 200/200 = unchanged
//
// 100 * 120/100 = 120; * 200/200 = 120.
func TestApplyStanceDamage_DamDoneAmplifies(t *testing.T) {
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{NumAttacks: 0, DamDone: 120, DamTaken: 0})

	ch := &types.CharData{Stance: types.STANCE_DRAGON, PCData: &types.PCData{}}
	ch.PCData.Stances[types.STANCE_DRAGON] = 200
	victim := &types.CharData{Stance: types.STANCE_NONE}

	got := applyStanceDamage(ch, victim, 100)
	if got != 120 {
		t.Errorf("DamDone=120 mastery=200: dam = %d, want 120", got)
	}
}

// Mastery below 100 clamps to 100 (C: min 0.5 == 100/200).
// DamDone=100 mastery=50 → clamped eff=100, dam = 100 * 100/100 * 100/200 = 50.
func TestApplyStanceDamage_MasteryClampedAtHalf(t *testing.T) {
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{NumAttacks: 0, DamDone: 100, DamTaken: 0})

	ch := &types.CharData{Stance: types.STANCE_DRAGON, PCData: &types.PCData{}}
	ch.PCData.Stances[types.STANCE_DRAGON] = 50 // below 100 → clamped
	victim := &types.CharData{Stance: types.STANCE_NONE}

	got := applyStanceDamage(ch, victim, 100)
	// dam = 100 (DamDone 100%) * 100/200 (clamped mastery) = 50
	if got != 50 {
		t.Errorf("mastery=50 DamDone=100: dam = %d, want 50 (mastery clamped to 100/200)", got)
	}
}

// Victim stance with DamTaken>100 reduces dam: at DamTaken=200 mastery=200,
// dam = 100 * 200/100 = 200, then dam /= (200/200=1.0) == dam * 200/200 = 200.
// Hmm. Let's pick a clean case: victim DamTaken=200 mastery=200 → half damage
// because the C code is counterintuitive. Actually: dam *= dam_taken/100 = *2,
// then dam /= temp_dam (1.0) → stays *2. That's NOT "defensive". Re-read C.
//
// C fight.c:2586-2592:
//   dam *= (dam_taken / 100.0)    // e.g. 120 → *1.2
//   temp_dam = pcdata->stances[stance] / 200.0
//   if (temp_dam < 0.5) temp_dam = 0.5
//   dam /= temp_dam               // higher mastery = smaller dam
//
// So DamTaken=120 mastery=200 → dam *= 1.2 /= 1.0 = *1.2 (takes 20% more)
// DamTaken=120 mastery=50 → dam *= 1.2 /= 0.5 = *2.4. The field name is
// misleading — high DamTaken is WORSE for the victim.
//
// Test: DamTaken=80 mastery=200 means dam *= 0.8 /= 1.0 = 80.
func TestApplyStanceDamage_DamTakenReducesDamage(t *testing.T) {
	withStanceIndex(t, types.STANCE_CRAB, StanceInfo{NumAttacks: 0, DamDone: 0, DamTaken: 80})

	ch := &types.CharData{Stance: types.STANCE_NONE}
	victim := &types.CharData{Stance: types.STANCE_CRAB, PCData: &types.PCData{}}
	victim.PCData.Stances[types.STANCE_CRAB] = 200

	got := applyStanceDamage(ch, victim, 100)
	// dam = 100 * 80/100 = 80; / 1.0 (mastery=200/200) = 80.
	if got != 80 {
		t.Errorf("DamTaken=80 mastery=200: dam = %d, want 80", got)
	}
}

// NPC victim uses ch.IndexData.Stances (not PCData).
func TestApplyStanceDamage_NPCReadsIndexData(t *testing.T) {
	withStanceIndex(t, types.STANCE_DRAGON, StanceInfo{NumAttacks: 0, DamDone: 120, DamTaken: 0})

	// NPC attacker: needs IndexData with Stances[DRAGON] set.
	idx := &types.MobIndexData{Vnum: 9999}
	idx.Stances[types.STANCE_DRAGON] = 200
	ch := &types.CharData{
		Stance:    types.STANCE_DRAGON,
		IndexData: idx,
	}
	ch.Act.Set(types.ACT_IS_NPC)
	victim := &types.CharData{Stance: types.STANCE_NONE}

	got := applyStanceDamage(ch, victim, 100)
	// Same math as PC test: 100 * 120/100 * 200/200 = 120.
	if got != 120 {
		t.Errorf("NPC IndexData: dam = %d, want 120", got)
	}
}

// stanceMastery handles missing PCData / IndexData cleanly.
func TestStanceMastery_NilPaths(t *testing.T) {
	pc := &types.CharData{PCData: nil}
	if got := stanceMastery(pc, types.STANCE_DRAGON); got != 0 {
		t.Errorf("nil PCData: mastery = %d, want 0", got)
	}
	npc := &types.CharData{IndexData: nil}
	npc.Act.Set(types.ACT_IS_NPC)
	if got := stanceMastery(npc, types.STANCE_DRAGON); got != 0 {
		t.Errorf("nil IndexData: mastery = %d, want 0", got)
	}
	// Out-of-range stance guards.
	pc2 := &types.CharData{PCData: &types.PCData{}}
	if got := stanceMastery(pc2, types.STANCE_NONE); got != 0 {
		t.Errorf("STANCE_NONE: mastery = %d, want 0", got)
	}
	if got := stanceMastery(pc2, types.MAX_STANCE); got != 0 {
		t.Errorf("MAX_STANCE: mastery = %d, want 0", got)
	}
}
