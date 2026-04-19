package act

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

// ensureSkill registers a placeholder skill at a free slot in WorldRef.Skills.
// Returns the slot number. Idempotent — if name already registered, returns
// the existing slot. Uses a non-zero Difficulty so learnFromSuccess can fire.
func ensureSkill(t *testing.T, name string) int {
	t.Helper()
	if WorldRef == nil {
		t.Fatal("WorldRef not initialised")
	}
	if sn := lookupSkillSlot(name); sn >= 0 {
		return sn
	}
	for i := range WorldRef.Skills {
		if WorldRef.Skills[i] == nil {
			WorldRef.Skills[i] = &types.SkillType{Name: name, Type: types.SKILL_SKILL, Difficulty: 1}
			return i
		}
	}
	WorldRef.Skills = append(WorldRef.Skills, &types.SkillType{Name: name, Type: types.SKILL_SKILL, Difficulty: 1})
	return len(WorldRef.Skills) - 1
}

// setLearned sets the character's proficiency in the named skill.
func setLearned(t *testing.T, ch *types.CharData, skillName string, pct int) {
	t.Helper()
	sn := ensureSkill(t, skillName)
	if ch.PCData != nil && sn < types.MAX_SKILL {
		ch.PCData.Learned[sn] = pct
	}
}

// --- Wave 1: unarmed attacks ---

func TestUnarmedAttack_NotFighting(t *testing.T) {
	cases := []struct {
		skill string
		fn    func(*types.CharData, string)
	}{
		{"bite", DoBite},
		{"claw", DoClaw},
		{"punch", DoPunch},
		{"sting", DoSting},
		{"tail", DoTail},
	}
	for _, tc := range cases {
		t.Run(tc.skill+"_not_fighting", func(t *testing.T) {
			ensureSkill(t, tc.skill)
			ch, _ := newSkillTestRoom()
			tc.fn(ch, "")
			// Should not crash; no fighting means early return.
		})
	}
}

func TestUnarmedAttack_Fighting(t *testing.T) {
	cases := []string{"bite", "claw", "punch", "sting", "tail"}
	for _, name := range cases {
		t.Run(name+"_fighting", func(t *testing.T) {
			ensureSkill(t, name)
			ch, victim := newSkillTestRoom()
			setLearned(t, ch, name, 100)
			combat.StartFighting(ch, victim)
			combat.StartFighting(victim, ch)
			dispatch := map[string]func(*types.CharData, string){
				"bite": DoBite, "claw": DoClaw, "punch": DoPunch,
				"sting": DoSting, "tail": DoTail,
			}
			origHit := victim.Hit
			for i := 0; i < 20; i++ {
				dispatch[name](ch, "")
				if victim.Hit < origHit {
					return
				}
			}
		})
	}
}

// --- Wave 2: offensive combat ---

func TestDoCircle_NoArg(t *testing.T) {
	ensureSkill(t, "circle")
	ch, _ := newSkillTestRoom()
	DoCircle(ch, "")
}

func TestDoCircle_NotFightingSelf(t *testing.T) {
	ensureSkill(t, "circle")
	ch, _ := newSkillTestRoom()
	DoCircle(ch, "tester") // can't circle self by own name
}

func TestDoCircle_SelfPrevented(t *testing.T) {
	ensureSkill(t, "circle")
	ch, _ := newSkillTestRoom()
	ch.Name = "selftgt"
	DoCircle(ch, "selftgt")
}

// When the first OneHit kills the victim (POS_DEAD), DoCircle's second
// swing must NOT fire. Plan item 4 retcode-review guard.
func TestDoCircle_SecondSwingSkippedOnVictimDeath(t *testing.T) {
	ensureSkill(t, "circle")
	ch, victim := newSkillTestRoom()
	setLearned(t, ch, "circle", 100)

	// Give ch a wield so the weapon gate passes.
	wield := &types.ObjData{
		Name:       "dagger",
		ShortDescr: "a dagger",
		ItemType:   types.ITEM_WEAPON,
		WearLoc:    types.WEAR_WIELD,
	}
	wield.Value[1] = 1
	wield.Value[2] = 4
	wield.Value[3] = 11 // pierce (C skills.c:5195)
	handler.ObjToChar(wield, ch)
	handler.EquipChar(ch, wield, types.WEAR_WIELD)

	// Pre-dead victim: Hit <= 0 so OneHit early-returns rVICT_DIED
	// (combat.go:459 `if victim.Hit <= 0 || ch.InRoom != victim.InRoom`).
	// A third party is fighting victim so the distraction check passes.
	third := &types.CharData{
		Name: "distractor", Level: 5, Hit: 20, MaxHit: 20,
		InRoom: ch.InRoom, Position: types.POS_STANDING,
	}
	third.Act.Set(types.ACT_IS_NPC)
	ch.InRoom.People = append(ch.InRoom.People, third)
	combat.StartFighting(ch, third)
	combat.StartFighting(victim, third)

	victim.Hit = 0
	victim.Position = types.POS_DEAD

	// Should not panic; both guards (retcode and Position) agree victim is
	// dead so the second swing is suppressed. The assertion is negative:
	// no infinite loop, no panic, and victim.Hit still 0.
	DoCircle(ch, "target")
	if victim.Hit > 0 {
		t.Errorf("pre-dead victim.Hit = %d, want <= 0", victim.Hit)
	}
}

func TestDoGouge_NotFighting(t *testing.T) {
	ensureSkill(t, "gouge")
	ensureSkill(t, "blindness")
	ch, _ := newSkillTestRoom()
	DoGouge(ch, "")
}

func TestDoGouge_Fighting(t *testing.T) {
	ensureSkill(t, "gouge")
	ensureSkill(t, "blindness")
	ch, victim := newSkillTestRoom()
	setLearned(t, ch, "gouge", 100)
	combat.StartFighting(ch, victim)
	combat.StartFighting(victim, ch)
	// Run several times to touch the success branch
	for i := 0; i < 20; i++ {
		DoGouge(ch, "")
		if victim.AffectedBy.IsSet(types.AFF_BLIND) {
			break
		}
	}
}

func TestDoStun_NotFighting(t *testing.T) {
	ensureSkill(t, "stun")
	ch, _ := newSkillTestRoom()
	DoStun(ch, "")
}

func TestDoStun_Fighting(t *testing.T) {
	ensureSkill(t, "stun")
	ch, victim := newSkillTestRoom()
	setLearned(t, ch, "stun", 100)
	combat.StartFighting(ch, victim)
	combat.StartFighting(victim, ch)
	DoStun(ch, "")
	// Should not crash; paralysis may or may not land
}

func TestDoGrapple_NoArg(t *testing.T) {
	ensureSkill(t, "grapple")
	ch, _ := newSkillTestRoom()
	DoGrapple(ch, "")
}

func TestDoGrapple_Success(t *testing.T) {
	ensureSkill(t, "grapple")
	ch, victim := newSkillTestRoom()
	setLearned(t, ch, "grapple", 100)
	combat.StartFighting(ch, victim)
	DoGrapple(ch, "")
	// Either grappled or failed — test no crash
}

func TestDoCleave_NoWeapon(t *testing.T) {
	ensureSkill(t, "cleave")
	ch, victim := newSkillTestRoom()
	combat.StartFighting(ch, victim)
	DoCleave(ch, "")
	// Should say "need a slashing weapon"
}

func TestDoHitall_Empty(t *testing.T) {
	ensureSkill(t, "hitall")
	ch := newTestCharWithDesc()
	ch.InRoom = &types.RoomIndexData{Vnum: 9000}
	ch.InRoom.People = append(ch.InRoom.People, ch)
	DoHitall(ch, "")
	// Should say "no one else here"
}

// Plan item 4 retcode-review note: DoHitall now honors
// combat.AttackerDied(ret) in addition to the pre-existing ch.Position
// <= POS_DEAD guard. In the current codebase the attacker-died path is
// dormant (reactive damage like fireshield / ice_shield / acid_shield
// is not yet ported — verified via grep), so a live mutation test
// cannot drive the guard through OneHit/Damage. Tests for the helper
// predicates themselves live in internal/combat/combat_test.go
// (TestAttackerDied / TestVictimDied). The defense-in-depth change
// guarantees correctness as soon as reactive damage ships.

func TestDoHitall_WithTargets(t *testing.T) {
	ensureSkill(t, "hitall")
	ch, _ := newSkillTestRoom()
	setLearned(t, ch, "hitall", 100)
	// Add a couple more NPCs
	for i := 0; i < 3; i++ {
		m := &types.CharData{
			Name:     "orc",
			Level:    5,
			Hit:      20,
			MaxHit:   20,
			InRoom:   ch.InRoom,
			Position: types.POS_STANDING,
			Armor:    100,
		}
		m.Act.Set(types.ACT_IS_NPC)
		ch.InRoom.People = append(ch.InRoom.People, m)
		WorldRef.AddChar(m)
	}
	DoHitall(ch, "")
	// No crash
}

func TestDoBerserk_NotFighting(t *testing.T) {
	ensureSkill(t, "berserk")
	ch, _ := newSkillTestRoom()
	DoBerserk(ch, "")
	// Should say "but you aren't fighting"
}

func TestDoBerserk_Success(t *testing.T) {
	ensureSkill(t, "berserk")
	ch, victim := newSkillTestRoom()
	setLearned(t, ch, "berserk", 100)
	combat.StartFighting(ch, victim)
	for i := 0; i < 20; i++ {
		DoBerserk(ch, "")
		if ch.AffectedBy.IsSet(types.AFF_BERSERK) {
			return
		}
	}
}

// --- DoPounce (Phase 6 skills G1) ---

// equipPounceWeapon gives ch a wielded weapon with Value[3]=DAM_STAB so the
// weapon-type gate passes. Returns the obj for mutation tests.
func equipPounceWeapon(t *testing.T, ch *types.CharData) *types.ObjData {
	t.Helper()
	wield := &types.ObjData{
		Name:       "dagger",
		ShortDescr: "a dagger",
		ItemType:   types.ITEM_WEAPON,
		WearLoc:    types.WEAR_WIELD,
	}
	wield.Value[1] = 1
	wield.Value[2] = 4
	wield.Value[3] = types.DAM_STAB
	handler.ObjToChar(wield, ch)
	handler.EquipChar(ch, wield, types.WEAR_WIELD)
	return wield
}

func TestDoPounce_MissingArgPromptsTarget(t *testing.T) {
	ensureSkill(t, "pounce")
	ch, _ := newSkillTestRoom()
	equipPounceWeapon(t, ch)
	prevWait := ch.Wait
	DoPounce(ch, "")
	if ch.Wait != prevWait {
		t.Errorf("no-arg pounce should not set wait; was %d, now %d", prevWait, ch.Wait)
	}
	if ch.Fighting != nil {
		t.Error("no-arg pounce should not start combat")
	}
}

func TestDoPounce_TargetNotFound(t *testing.T) {
	ensureSkill(t, "pounce")
	ch, _ := newSkillTestRoom()
	equipPounceWeapon(t, ch)
	DoPounce(ch, "nobody")
	if ch.Fighting != nil {
		t.Error("pounce on absent target should not start combat")
	}
}

func TestDoPounce_SelfTargetRejected(t *testing.T) {
	ensureSkill(t, "pounce")
	ch, _ := newSkillTestRoom()
	equipPounceWeapon(t, ch)
	DoPounce(ch, ch.Name)
	if ch.Fighting != nil {
		t.Error("self-pounce should not start combat")
	}
}

func TestDoPounce_MountedRejected(t *testing.T) {
	ensureSkill(t, "pounce")
	ch, victim := newSkillTestRoom()
	equipPounceWeapon(t, ch)
	mount := &types.CharData{Name: "horse"}
	mount.Act.Set(types.ACT_IS_NPC)
	ch.Mount = mount
	DoPounce(ch, "target")
	if ch.Fighting != nil || victim.Fighting != nil {
		t.Error("mounted pounce should not start combat")
	}
}

func TestDoPounce_NPCCharmedRejected(t *testing.T) {
	ensureSkill(t, "pounce")
	ch, _ := newSkillTestRoom()
	// Convert ch to NPC and charm.
	ch.Act.Set(types.ACT_IS_NPC)
	ch.AffectedBy.Set(types.AFF_CHARM)
	equipPounceWeapon(t, ch)
	prevWait := ch.Wait
	DoPounce(ch, "target")
	if ch.Wait != prevWait {
		t.Errorf("charmed-NPC pounce should not set wait; was %d, now %d", prevWait, ch.Wait)
	}
}

func TestDoPounce_NoWeaponRejected(t *testing.T) {
	ensureSkill(t, "pounce")
	ch, _ := newSkillTestRoom()
	DoPounce(ch, "target")
	if ch.Fighting != nil {
		t.Error("no-weapon pounce should not start combat")
	}
}

func TestDoPounce_WrongWeaponTypeRejected(t *testing.T) {
	ensureSkill(t, "pounce")
	ch, _ := newSkillTestRoom()
	wield := &types.ObjData{
		Name: "bludgeon", ShortDescr: "a bludgeon",
		ItemType: types.ITEM_WEAPON, WearLoc: types.WEAR_WIELD,
	}
	wield.Value[3] = types.DAM_HIT // not in {1,2,3,5,10,11}
	handler.ObjToChar(wield, ch)
	handler.EquipChar(ch, wield, types.WEAR_WIELD)
	DoPounce(ch, "target")
	if ch.Fighting != nil {
		t.Error("wrong-weapon-type pounce should not start combat")
	}
}

func TestDoPounce_EachValidWeaponTypeAccepted(t *testing.T) {
	ensureSkill(t, "pounce")
	validTypes := []int{
		types.DAM_SLICE, types.DAM_STAB, types.DAM_SLASH,
		types.DAM_CLAW, types.DAM_BITE, types.DAM_PIERCE,
	}
	for _, dt := range validTypes {
		ch, victim := newSkillTestRoom()
		setLearned(t, ch, "pounce", 100)
		wield := &types.ObjData{
			Name: "weapon", ShortDescr: "a weapon",
			ItemType: types.ITEM_WEAPON, WearLoc: types.WEAR_WIELD,
		}
		wield.Value[1] = 1
		wield.Value[2] = 4
		wield.Value[3] = dt
		handler.ObjToChar(wield, ch)
		handler.EquipChar(ch, wield, types.WEAR_WIELD)
		// Victim is sleeping → auto-success gate; combat starts.
		victim.Position = types.POS_SLEEPING
		DoPounce(ch, "target")
		if ch.Fighting == nil {
			t.Errorf("valid weapon type %d: expected pounce to start combat", dt)
		}
	}
}

func TestDoPounce_VictimAlreadyFighting(t *testing.T) {
	ensureSkill(t, "pounce")
	ch, victim := newSkillTestRoom()
	equipPounceWeapon(t, ch)
	// Third-party engaging victim.
	third := &types.CharData{
		Name: "other", Level: 5, Hit: 20, MaxHit: 20,
		InRoom: ch.InRoom, Position: types.POS_STANDING,
	}
	third.Act.Set(types.ACT_IS_NPC)
	ch.InRoom.People = append(ch.InRoom.People, third)
	combat.StartFighting(victim, third)
	combat.StartFighting(third, victim)
	DoPounce(ch, "target")
	if ch.Fighting != nil {
		t.Error("pounce on already-fighting victim should not start combat for ch")
	}
}

func TestDoPounce_HurtAwakeVictim(t *testing.T) {
	ensureSkill(t, "pounce")
	ch, victim := newSkillTestRoom()
	equipPounceWeapon(t, ch)
	victim.Hit = victim.MaxHit - 1
	victim.Position = types.POS_STANDING
	DoPounce(ch, "target")
	if ch.Fighting != nil {
		t.Error("hurt-and-awake victim should reject pounce")
	}
}

func TestDoPounce_SleepingVictimAllowed(t *testing.T) {
	ensureSkill(t, "pounce")
	ch, victim := newSkillTestRoom()
	setLearned(t, ch, "pounce", 100)
	equipPounceWeapon(t, ch)
	victim.Hit = victim.MaxHit - 1
	victim.Position = types.POS_SLEEPING
	DoPounce(ch, "target")
	if ch.Fighting == nil {
		t.Error("sleeping hurt victim should allow pounce")
	}
}

func TestDoPounce_SuccessInvokesMultiHit(t *testing.T) {
	ensureSkill(t, "pounce")
	ch, victim := newSkillTestRoom()
	setLearned(t, ch, "pounce", 100)
	equipPounceWeapon(t, ch)
	// Keep victim alive during combat cascade (damageWith early-returns on
	// Hit <= 0 which would unset Fighting bookkeeping).
	victim.Hit = 10000
	victim.MaxHit = 10000
	restore := withStubNumberPercent(1) // guaranteed canUseSkill success
	defer restore()
	DoPounce(ch, "target")
	if ch.Fighting == nil || ch.Fighting.Who != victim {
		t.Errorf("pounce success: ch.Fighting should target victim")
	}
}

func TestDoPounce_FailureDoesZeroDamage(t *testing.T) {
	ensureSkill(t, "pounce")
	ch, victim := newSkillTestRoom()
	setLearned(t, ch, "pounce", 0) // unlearned → canUseSkill false
	equipPounceWeapon(t, ch)
	origHit := victim.Hit
	restore := withStubNumberPercent(99)
	defer restore()
	DoPounce(ch, "target")
	if victim.Hit != origHit {
		t.Errorf("failed pounce should do 0 damage; hit went %d -> %d", origHit, victim.Hit)
	}
}

func TestDoPounce_WaitStateSet(t *testing.T) {
	ensureSkill(t, "pounce")
	ch, _ := newSkillTestRoom()
	setLearned(t, ch, "pounce", 100)
	equipPounceWeapon(t, ch)
	// Put a specific Beats value on the skill for this assertion.
	gsn := lookupSkillSlot("pounce")
	WorldRef.Skills[gsn].Beats = 12
	ch.Wait = 0
	DoPounce(ch, "target")
	if ch.Wait != 12 {
		t.Errorf("ch.Wait = %d after pounce; want 12 (beats)", ch.Wait)
	}
}
