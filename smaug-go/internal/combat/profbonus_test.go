package combat

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// --- G6 — Weapon proficiency bonus ---

// Helper: install a prof-gsn cache with distinct ints in the combat
// package globals so WeaponProfBonusCheck maps reliably from DAM_* to gsn.
// Matches the boot-time behavior but deterministic for tests.
func stubProfGsns(t *testing.T) {
	t.Helper()
	saved := struct {
		pu, lb, sb, fa, ta, bl, mw int
	}{
		gsnPugilism, gsnLongBlades, gsnShortBlades, gsnFlexibleArms,
		gsnTalonousArms, gsnBludgeons, gsnMissileWeapons,
	}
	t.Cleanup(func() {
		gsnPugilism = saved.pu
		gsnLongBlades = saved.lb
		gsnShortBlades = saved.sb
		gsnFlexibleArms = saved.fa
		gsnTalonousArms = saved.ta
		gsnBludgeons = saved.bl
		gsnMissileWeapons = saved.mw
	})
	gsnPugilism = 70
	gsnLongBlades = 71
	gsnShortBlades = 72
	gsnFlexibleArms = 73
	gsnTalonousArms = 74
	gsnBludgeons = 75
	gsnMissileWeapons = 76
}

// PC level 10 wielding a DAM_SLASH weapon, Learned[long_blades]=100 →
// bonus = (100 - 50) / 10 = 5. profGsn == gsnLongBlades.
func TestWeaponProfBonus_Slashing_Learned100(t *testing.T) {
	stubProfGsns(t)
	ch := newPCFighter(10)
	wield := &types.ObjData{Name: "sword"}
	wield.Value[3] = types.DAM_SLASH
	ch.PCData.Learned[gsnLongBlades] = 100

	bonus, profGsn := WeaponProfBonusCheck(ch, wield)
	if bonus != 5 {
		t.Errorf("bonus = %d, want 5 (Learned=100, (100-50)/10)", bonus)
	}
	if profGsn != gsnLongBlades {
		t.Errorf("profGsn = %d, want gsnLongBlades (%d)", profGsn, gsnLongBlades)
	}
}

// PC level 10 wielding DAM_STAB → short_blades, Learned=50 → bonus=0.
func TestWeaponProfBonus_ShortBlades_Learned50_Zero(t *testing.T) {
	stubProfGsns(t)
	ch := newPCFighter(10)
	wield := &types.ObjData{Name: "dagger"}
	wield.Value[3] = types.DAM_STAB
	ch.PCData.Learned[gsnShortBlades] = 50

	bonus, profGsn := WeaponProfBonusCheck(ch, wield)
	if bonus != 0 {
		t.Errorf("bonus = %d, want 0 (Learned=50, (50-50)/10)", bonus)
	}
	if profGsn != gsnShortBlades {
		t.Errorf("profGsn = %d, want gsnShortBlades (%d)", profGsn, gsnShortBlades)
	}
}

// PC level 10 wielding DAM_PIERCE, Learned[short_blades]=0 → bonus=-5
// (penalty: (0-50)/10). C formula is deliberately negative for unlearned.
func TestWeaponProfBonus_Unlearned_IsNegative(t *testing.T) {
	stubProfGsns(t)
	ch := newPCFighter(10)
	wield := &types.ObjData{Name: "rapier"}
	wield.Value[3] = types.DAM_PIERCE
	// Learned stays 0.

	bonus, _ := WeaponProfBonusCheck(ch, wield)
	if bonus != -5 {
		t.Errorf("bonus = %d, want -5 (Learned=0, (0-50)/10)", bonus)
	}
}

// Level 5 PC — gate is level > 5, so level=5 gets no bonus.
func TestWeaponProfBonus_BelowLevel6_NoBonus(t *testing.T) {
	stubProfGsns(t)
	ch := newPCFighter(5)
	wield := &types.ObjData{Name: "sword"}
	wield.Value[3] = types.DAM_SLASH
	ch.PCData.Learned[gsnLongBlades] = 100

	bonus, profGsn := WeaponProfBonusCheck(ch, wield)
	if bonus != 0 || profGsn != -1 {
		t.Errorf("level=5: got (%d, %d), want (0, -1)", bonus, profGsn)
	}
}

// Level 6 PC is exactly the first eligible level (C: level > 5).
func TestWeaponProfBonus_Level6_IsEligible(t *testing.T) {
	stubProfGsns(t)
	ch := newPCFighter(6)
	wield := &types.ObjData{Name: "sword"}
	wield.Value[3] = types.DAM_SLASH
	ch.PCData.Learned[gsnLongBlades] = 100

	bonus, profGsn := WeaponProfBonusCheck(ch, wield)
	if bonus != 5 || profGsn != gsnLongBlades {
		t.Errorf("level=6: got (%d, %d), want (5, %d)", bonus, profGsn, gsnLongBlades)
	}
}

// NPCs never get the bonus.
func TestWeaponProfBonus_NPC_NoBonus(t *testing.T) {
	stubProfGsns(t)
	ch := newFighter("Mob", 50)
	ch.Act.Set(types.ACT_IS_NPC)
	wield := &types.ObjData{Name: "sword"}
	wield.Value[3] = types.DAM_SLASH

	bonus, profGsn := WeaponProfBonusCheck(ch, wield)
	if bonus != 0 || profGsn != -1 {
		t.Errorf("NPC: got (%d, %d), want (0, -1)", bonus, profGsn)
	}
}

// Unarmed PC: wield == nil → no bonus.
func TestWeaponProfBonus_NoWield_NoBonus(t *testing.T) {
	stubProfGsns(t)
	ch := newPCFighter(50)

	bonus, profGsn := WeaponProfBonusCheck(ch, nil)
	if bonus != 0 || profGsn != -1 {
		t.Errorf("no-wield: got (%d, %d), want (0, -1)", bonus, profGsn)
	}
}

// All DAM_* types map to the correct prof gsn per C fight.c:1264-1300.
func TestWeaponProfBonus_AllDamTypesMap(t *testing.T) {
	stubProfGsns(t)
	cases := []struct {
		damType int
		want    int
		name    string
	}{
		{types.DAM_HIT, gsnPugilism, "DAM_HIT → pugilism"},
		{types.DAM_SUCTION, gsnPugilism, "DAM_SUCTION → pugilism"},
		{types.DAM_BITE, gsnPugilism, "DAM_BITE → pugilism"},
		{types.DAM_BLAST, gsnPugilism, "DAM_BLAST → pugilism"},
		{types.DAM_SLASH, gsnLongBlades, "DAM_SLASH → long_blades"},
		{types.DAM_SLICE, gsnLongBlades, "DAM_SLICE → long_blades"},
		{types.DAM_PIERCE, gsnShortBlades, "DAM_PIERCE → short_blades"},
		{types.DAM_STAB, gsnShortBlades, "DAM_STAB → short_blades"},
		{types.DAM_WHIP, gsnFlexibleArms, "DAM_WHIP → flexible_arms"},
		{types.DAM_CLAW, gsnTalonousArms, "DAM_CLAW → talonous_arms"},
		{types.DAM_POUND, gsnBludgeons, "DAM_POUND → bludgeons"},
		{types.DAM_CRUSH, gsnBludgeons, "DAM_CRUSH → bludgeons"},
		{types.DAM_BOLT, gsnMissileWeapons, "DAM_BOLT → missile_weapons"},
		{types.DAM_ARROW, gsnMissileWeapons, "DAM_ARROW → missile_weapons"},
		{types.DAM_DART, gsnMissileWeapons, "DAM_DART → missile_weapons"},
		{types.DAM_STONE, gsnMissileWeapons, "DAM_STONE → missile_weapons"},
		{types.DAM_PEA, gsnMissileWeapons, "DAM_PEA → missile_weapons"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			wield := &types.ObjData{}
			wield.Value[3] = c.damType
			got := profGsnForWeapon(wield)
			if got != c.want {
				t.Errorf("profGsnForWeapon(%d) = %d, want %d", c.damType, got, c.want)
			}
		})
	}
}

// When prof gsn is not resolved (e.g. name not in skill table), bonus=0.
// Defensive: guards against a missing "long blades" entry in skills.dat.
func TestWeaponProfBonus_UnresolvedGsn_NoBonus(t *testing.T) {
	// Temporarily unset gsnLongBlades so the switch returns -1 via the
	// fallthrough guard in WeaponProfBonusCheck.
	saved := gsnLongBlades
	t.Cleanup(func() { gsnLongBlades = saved })
	gsnLongBlades = -1

	ch := newPCFighter(50)
	wield := &types.ObjData{Name: "sword"}
	wield.Value[3] = types.DAM_SLASH
	ch.PCData.Learned[0] = 100 // irrelevant; gsn is -1

	bonus, profGsn := WeaponProfBonusCheck(ch, wield)
	if bonus != 0 || profGsn != -1 {
		t.Errorf("unresolved gsn: got (%d, %d), want (0, -1)", bonus, profGsn)
	}
}

// --- OneHit integration: prof bonus affects victim AC + damage ---

// Setup helper: PC with a DAM_SLASH wield, room, NPC victim, no-cascade
// world. Returns (w, room, ch, victim, wield).
func setupProfScenario(t *testing.T, learned int) (*world.World, *types.RoomIndexData, *types.CharData, *types.CharData, *types.ObjData) {
	t.Helper()
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 10100, Name: "Prof Arena"}
	w.Rooms[10100] = room

	ch := newPCFighter(50)
	ch.Hitroll = 0
	ch.Damroll = 0
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	wield := &types.ObjData{
		Name:     "longsword",
		WearLoc:  types.WEAR_WIELD,
		ItemType: types.ITEM_WEAPON,
	}
	wield.Value[1] = 1
	wield.Value[2] = 1 // fixed 1 so damage math is deterministic
	wield.Value[3] = types.DAM_SLASH
	handler.ObjToChar(wield, ch)

	ch.PCData.Learned[gsnLongBlades] = learned

	victim := newFighter("Victim", 1)
	victim.Act.Set(types.ACT_IS_NPC)
	victim.Hit = 1_000_000
	victim.MaxHit = 1_000_000
	victim.Armor = 0 // victim_ac = 0 without prof_bonus
	handler.CharToRoom(victim, room)
	w.AddChar(victim)
	StartFighting(ch, victim)

	return w, room, ch, victim, wield
}

// OneHit on hit applies prof bonus /4 to damage. Fixed weapon 1..1
// damage, Learned[long_blades]=100, bonus=5 → dam += 5/4 = 1 extra.
// Expected: base dam (1) + str + damroll + prof/4 = 1 + 0 + 0 + 1 = 2.
// Compare two OneHits over many rounds — prof bonus should add +1/hit.
//
// To keep this deterministic we test the prof-branch directly in code
// via WeaponProfBonusCheck(5) -> bonus==5; the integration here is that
// OneHit's dam computation calls WeaponProfBonusCheck. A precise count
// is impossible without RNG stubs on rollD20; instead we run enough
// rounds to confirm the high-Learned PC deals strictly more total damage
// than the Learned=0 baseline.
func TestOneHit_ProfBonus_HigherLearnedDealsMoreDamage(t *testing.T) {
	stubProfGsns(t)
	const rounds = 2000

	// Baseline: Learned=0 → bonus=-5 → dam penalty of 1 per hit, AC penalty of -5.
	wBase, _, chBase, vicBase, _ := setupProfScenario(t, 0)
	totalBase := 0
	for i := 0; i < rounds; i++ {
		startHP := vicBase.Hit
		OneHit(wBase, chBase, vicBase, types.TYPE_UNDEFINED)
		totalBase += (startHP - vicBase.Hit)
	}

	// High Learned=100 → bonus=+5 → dam bonus of 1 per hit, AC bonus too.
	wHi, _, chHi, vicHi, _ := setupProfScenario(t, 100)
	totalHi := 0
	for i := 0; i < rounds; i++ {
		startHP := vicHi.Hit
		OneHit(wHi, chHi, vicHi, types.TYPE_UNDEFINED)
		totalHi += (startHP - vicHi.Hit)
	}

	if totalHi <= totalBase {
		t.Errorf("Learned=100 total=%d should exceed Learned=0 total=%d", totalHi, totalBase)
	}
}

// OneHit miss path triggers learnFromFailure on the prof gsn.
func TestOneHit_ProfMissCallsLearnFailure(t *testing.T) {
	stubProfGsns(t)
	savedF := LearnFromFailureHook
	t.Cleanup(func() { LearnFromFailureHook = savedF })

	var failureGsns []int
	LearnFromFailureHook = func(_ *types.CharData, gsn int) {
		failureGsns = append(failureGsns, gsn)
	}

	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 10101, Name: "Miss Arena"}
	w.Rooms[10101] = room

	ch := newPCFighter(50)
	ch.Hitroll = -999 // guaranteed miss
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	wield := &types.ObjData{
		Name: "longsword", WearLoc: types.WEAR_WIELD, ItemType: types.ITEM_WEAPON,
	}
	wield.Value[1] = 1
	wield.Value[2] = 1
	wield.Value[3] = types.DAM_SLASH
	handler.ObjToChar(wield, ch)
	ch.PCData.Learned[gsnLongBlades] = 50 // bonus = 0, but profGsn set

	victim := newFighter("Target", 1)
	victim.Act.Set(types.ACT_IS_NPC)
	victim.Hit = 100000
	victim.Armor = -100 // very good AC
	handler.CharToRoom(victim, room)
	w.AddChar(victim)
	StartFighting(ch, victim)

	// Do several OneHits; some (ideally all) miss — confirm at least one
	// miss triggered learn-from-failure.
	missedAtLeastOnce := false
	for i := 0; i < 30; i++ {
		beforeHP := victim.Hit
		OneHit(w, ch, victim, types.TYPE_UNDEFINED)
		if victim.Hit == beforeHP {
			missedAtLeastOnce = true
		}
	}
	if !missedAtLeastOnce {
		t.Skip("RNG never produced a miss; cannot assert learn-from-failure path. Non-deterministic skip.")
	}

	found := false
	for _, g := range failureGsns {
		if g == gsnLongBlades {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("missed OneHit did not trigger learnFromFailure(gsnLongBlades); got %v", failureGsns)
	}
}
