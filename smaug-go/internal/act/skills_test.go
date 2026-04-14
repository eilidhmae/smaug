package act

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

func newSkillTestRoom() (*types.CharData, *types.CharData) {
	room := &types.RoomIndexData{Vnum: 8000, Name: "Arena"}

	ch := newTestCharWithDesc()
	ch.InRoom = room
	ch.Level = 20
	ch.Hit = 100
	ch.MaxHit = 100
	ch.Position = types.POS_STANDING
	room.People = append(room.People, ch)

	victim := &types.CharData{
		Name:       "target",
		ShortDescr: "a training dummy",
		Level:      10,
		Position:   types.POS_STANDING,
		Hit:        100,
		MaxHit:     100,
		Armor:      100,
		PermStr:    15,
		PermDex:    15,
		InRoom:     room,
	}
	victim.Act.Set(types.ACT_IS_NPC)
	room.People = append(room.People, victim)
	WorldRef.AddChar(victim)

	return ch, victim
}

func TestDoBackstab_NoArg(t *testing.T) {
	ch, _ := newSkillTestRoom()
	DoBackstab(ch, "")
	// Should say "Backstab whom?"
}

func TestDoBackstab_NoWeapon(t *testing.T) {
	ch, _ := newSkillTestRoom()
	DoBackstab(ch, "target")
	// Should say "You need to wield a weapon"
}

func TestDoBash_NotFighting(t *testing.T) {
	ch, _ := newSkillTestRoom()
	DoBash(ch, "")
	// Should say "You aren't fighting anyone"
}

func TestDoBash_InCombat(t *testing.T) {
	ch, victim := newSkillTestRoom()
	combat.StartFighting(ch, victim)
	combat.StartFighting(victim, ch)

	origHit := victim.Hit
	// Run bash many times — may succeed or fail
	for i := 0; i < 20; i++ {
		DoBash(ch, "")
		if victim.Hit < origHit {
			return // success path hit
		}
	}
	// Even if all misses, test shouldn't fail — just verify no crash
}

func TestDoKickSkill_InCombat(t *testing.T) {
	ch, victim := newSkillTestRoom()
	combat.StartFighting(ch, victim)

	DoKickSkill(ch, "")
	// Should not crash
}

func TestDoDisarm_NothingToDisarm(t *testing.T) {
	ch, victim := newSkillTestRoom()
	combat.StartFighting(ch, victim)

	DoDisarm(ch, "")
	// Should say victim has no weapon or ch needs a weapon
}

func TestDoRescue_NotFighting(t *testing.T) {
	ch, _ := newSkillTestRoom()
	DoRescue(ch, "target")
	// Should say target not fighting
}

func TestDoSneak(t *testing.T) {
	ch, _ := newSkillTestRoom()
	// Set high skill for reliable success
	if ch.PCData != nil {
		gsn := lookupSkillSlot("sneak")
		if gsn >= 0 && gsn < types.MAX_SKILL {
			ch.PCData.Learned[gsn] = 100
		}
	}

	DoSneak(ch, "")
	// May or may not succeed — just verify no crash
}

func TestDoHide(t *testing.T) {
	ch, _ := newSkillTestRoom()
	DoHide(ch, "")
	// Should not crash
}

func TestDoSteal_NoArgs(t *testing.T) {
	ch, _ := newSkillTestRoom()
	DoSteal(ch, "")
	// Should say "Steal what from whom?"
}

func TestDoScan(t *testing.T) {
	ch, _ := newSkillTestRoom()

	// Add an adjacent room
	room2 := &types.RoomIndexData{Vnum: 8001, Name: "Adjacent Room"}
	ch.InRoom.Exits = append(ch.InRoom.Exits, &types.ExitData{
		Direction: types.DIR_NORTH,
		ToRoom:    room2,
	})
	mob := &types.CharData{
		Name:       "guard",
		ShortDescr: "a guard",
		InRoom:     room2,
	}
	mob.Act.Set(types.ACT_IS_NPC)
	room2.People = append(room2.People, mob)

	DoScan(ch, "")
	// Should show the guard to the north
}

func TestDoAid_NotIncap(t *testing.T) {
	ch, victim := newSkillTestRoom()
	DoAid(ch, "target")
	_ = victim
	// Should say "don't need your aid"
}

func TestDoAid_Success(t *testing.T) {
	ch, victim := newSkillTestRoom()
	victim.Hit = -2
	victim.Position = types.POS_INCAP

	// Ensure the "aid" skill exists in the world skill table
	gsn := lookupSkillSlot("aid")
	if gsn < 0 {
		// Add it temporarily
		for gsn = 0; gsn < len(WorldRef.Skills); gsn++ {
			if WorldRef.Skills[gsn] == nil {
				WorldRef.Skills[gsn] = &types.SkillType{Name: "aid", Type: types.SKILL_SKILL}
				break
			}
		}
		if gsn >= len(WorldRef.Skills) {
			WorldRef.Skills = append(WorldRef.Skills, &types.SkillType{Name: "aid", Type: types.SKILL_SKILL})
			gsn = len(WorldRef.Skills) - 1
		}
	}
	if ch.PCData != nil && gsn >= 0 && gsn < types.MAX_SKILL {
		ch.PCData.Learned[gsn] = 100
	}

	// Try multiple times since skill can fail
	for i := 0; i < 50; i++ {
		DoAid(ch, "target")
		if victim.Hit >= 1 {
			break
		}
		victim.Hit = -2
		victim.Position = types.POS_INCAP
	}

	if victim.Hit < 1 {
		t.Error("aid should restore victim to 1 HP after enough attempts")
	}
}

func TestDoRecall(t *testing.T) {
	ch, _ := newSkillTestRoom()
	temple := WorldRef.GetRoom(types.ROOM_VNUM_TEMPLE)
	if temple == nil {
		// Create a temple for the test
		temple = &types.RoomIndexData{Vnum: types.ROOM_VNUM_TEMPLE, Name: "Temple"}
		WorldRef.Rooms[types.ROOM_VNUM_TEMPLE] = temple
	}

	DoRecall(ch, "")
	if ch.InRoom != temple {
		t.Error("recall should move player to temple")
	}
}

func TestCanUseSkill_NPC(t *testing.T) {
	if !canUseSkill(&types.CharData{Act: types.BitVector{1}}, 50, 0) {
		// NPC with percent < 85 should succeed
	}
	npc := &types.CharData{}
	npc.Act.Set(types.ACT_IS_NPC)
	if !canUseSkill(npc, 50, 0) {
		t.Error("NPC should succeed with percent 50 < 85")
	}
	if canUseSkill(npc, 90, 0) {
		t.Error("NPC should fail with percent 90 >= 85")
	}
}

// installLearnTestSkill places a stub SkillType into WorldRef.Skills at the
// given slot with the supplied difficulty and per-class adept cap. It returns
// a cleanup func that restores the previous slot value.
func installLearnTestSkill(t *testing.T, gsn, difficulty, adept int, name string) func() {
	t.Helper()
	for len(WorldRef.Skills) <= gsn {
		WorldRef.Skills = append(WorldRef.Skills, nil)
	}
	prev := WorldRef.Skills[gsn]
	sk := &types.SkillType{
		Name:       name,
		Type:       types.SKILL_SKILL,
		Difficulty: difficulty,
	}
	for i := 0; i < types.MAX_CLASS; i++ {
		sk.SkillAdept[i] = adept
	}
	WorldRef.Skills[gsn] = sk
	return func() { WorldRef.Skills[gsn] = prev }
}

// withStubNumberPercent swaps the package-level numberPercent with a stub
// that returns the given values in order (last value repeats). It returns
// a cleanup func restoring the original.
func withStubNumberPercent(values ...int) func() {
	prev := numberPercent
	idx := 0
	numberPercent = func() int {
		v := values[idx]
		if idx < len(values)-1 {
			idx++
		}
		return v
	}
	return func() { numberPercent = prev }
}

func TestLearnFromSuccess_FormulaTable(t *testing.T) {
	// Use a slot we control. Slot 500 is well inside MAX_SKILL (600) and
	// unlikely to collide with anything the test world might populate.
	const gsn = 500
	restore := installLearnTestSkill(t, gsn, 5, 95, "testskill")
	defer restore()

	tests := []struct {
		name      string
		learned   int
		roll      int
		wantDelta int
		wantMsg   bool
	}{
		// chance = learned + 5*5 = learned + 25
		{"roll>=chance gains 2", 10, 99, 2, true},           // chance 35, roll 99 -> gain 2
		{"chance-roll<=25 gains 1", 10, 20, 1, true},        // chance 35, diff 15 -> gain 1
		{"chance-roll>25 no gain", 10, 5, 0, false},         // chance 35, diff 30 -> no gain
		{"roll exactly chance gains 2", 10, 35, 2, true},    // roll>=chance path
		{"diff exactly 25 gains 1", 10, 10, 1, true},        // chance-roll==25 -> gain 1
		{"cap at adept on gain of 2", 94, 99, 1, true},      // 94+2=96 clamped to 95
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ch := newTestCharWithDesc()
			ch.PCData.Learned[gsn] = tc.learned
			restoreRNG := withStubNumberPercent(tc.roll)
			defer restoreRNG()

			learnFromSuccess(ch, gsn)
			got := ch.PCData.Learned[gsn] - tc.learned
			if got != tc.wantDelta {
				t.Errorf("delta = %d, want %d (learned %d -> %d)",
					got, tc.wantDelta, tc.learned, ch.PCData.Learned[gsn])
			}
		})
	}
}

func TestLearnFromSuccess_AtAdeptNoChange(t *testing.T) {
	const gsn = 501
	restore := installLearnTestSkill(t, gsn, 5, 95, "capped")
	defer restore()
	restoreRNG := withStubNumberPercent(99) // would gain if not capped
	defer restoreRNG()

	ch := newTestCharWithDesc()
	ch.PCData.Learned[gsn] = 95 // already at adept
	learnFromSuccess(ch, gsn)
	if ch.PCData.Learned[gsn] != 95 {
		t.Errorf("learned should stay 95 at adept cap, got %d", ch.PCData.Learned[gsn])
	}
}

func TestLearnFromSuccess_NPCNoChange(t *testing.T) {
	const gsn = 502
	restore := installLearnTestSkill(t, gsn, 5, 95, "npcskill")
	defer restore()
	restoreRNG := withStubNumberPercent(99)
	defer restoreRNG()

	ch := newTestCharWithDesc()
	ch.Act.Set(types.ACT_IS_NPC)
	ch.PCData.Learned[gsn] = 10
	learnFromSuccess(ch, gsn)
	if ch.PCData.Learned[gsn] != 10 {
		t.Errorf("NPC should not learn, got %d", ch.PCData.Learned[gsn])
	}
}

func TestLearnFromSuccess_NilSkillSlot(t *testing.T) {
	// gsn within range but WorldRef.Skills[gsn] is nil -> early return
	const gsn = 503
	for len(WorldRef.Skills) <= gsn {
		WorldRef.Skills = append(WorldRef.Skills, nil)
	}
	prev := WorldRef.Skills[gsn]
	WorldRef.Skills[gsn] = nil
	defer func() { WorldRef.Skills[gsn] = prev }()

	restoreRNG := withStubNumberPercent(99)
	defer restoreRNG()

	ch := newTestCharWithDesc()
	ch.PCData.Learned[gsn] = 10
	learnFromSuccess(ch, gsn)
	if ch.PCData.Learned[gsn] != 10 {
		t.Errorf("nil skill slot should not learn, got %d", ch.PCData.Learned[gsn])
	}
}

func TestLearnFromFailure_FormulaTable(t *testing.T) {
	const gsn = 504
	restore := installLearnTestSkill(t, gsn, 5, 95, "failskill")
	defer restore()

	tests := []struct {
		name      string
		learned   int
		roll      int
		wantDelta int
	}{
		// chance = learned + 25
		{"diff<=25 and roll<chance gains 1", 10, 20, 1}, // chance 35, diff 15 -> gain
		// C src/skills.c:1682 has only the "> 25 => return" guard. Any roll
		// within 25 of chance gains, even if it beat chance.
		{"roll>=chance still gains when within 25", 10, 50, 1}, // chance 35, diff=-15 -> gain
		{"diff>25 no gain", 10, 5, 0},                          // chance-roll = 30 -> no gain
		{"diff exactly 25 gains 1", 10, 10, 1},                 // chance-roll == 25
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ch := newTestCharWithDesc()
			ch.PCData.Learned[gsn] = tc.learned
			restoreRNG := withStubNumberPercent(tc.roll)
			defer restoreRNG()

			learnFromFailure(ch, gsn)
			got := ch.PCData.Learned[gsn] - tc.learned
			if got != tc.wantDelta {
				t.Errorf("delta = %d, want %d (learned %d -> %d)",
					got, tc.wantDelta, tc.learned, ch.PCData.Learned[gsn])
			}
		})
	}
}

func TestLearnFromFailure_AtAdeptMinusOneNoChange(t *testing.T) {
	const gsn = 505
	restore := installLearnTestSkill(t, gsn, 5, 95, "failcap")
	defer restore()
	restoreRNG := withStubNumberPercent(20) // would normally gain
	defer restoreRNG()

	ch := newTestCharWithDesc()
	ch.PCData.Learned[gsn] = 94 // adept-1
	learnFromFailure(ch, gsn)
	if ch.PCData.Learned[gsn] != 94 {
		t.Errorf("learned should stay at adept-1, got %d", ch.PCData.Learned[gsn])
	}
}

func TestLearnFromFailure_ClampAtAdeptMinusOne(t *testing.T) {
	const gsn = 506
	restore := installLearnTestSkill(t, gsn, 5, 95, "failclamp")
	defer restore()
	// learned=93, adept-1=94. chance = 93+25 = 118 -> both conditions easy.
	// Pick roll=100 so chance-roll=18 (<=25) and roll<chance -> gain 1.
	restoreRNG := withStubNumberPercent(100)
	defer restoreRNG()

	ch := newTestCharWithDesc()
	ch.PCData.Learned[gsn] = 93
	learnFromFailure(ch, gsn)
	if ch.PCData.Learned[gsn] != 94 {
		t.Errorf("should gain 1 up to adept-1 (94), got %d", ch.PCData.Learned[gsn])
	}
}

func TestLearnFromSuccess(t *testing.T) {
	// Regression: using the stubbed RNG and a known-good skill entry, a
	// chain of high rolls must keep pushing learned upward until it caps.
	const gsn = 507
	restore := installLearnTestSkill(t, gsn, 5, 95, "regress")
	defer restore()
	restoreRNG := withStubNumberPercent(99) // always "roll >= chance", gain 2
	defer restoreRNG()

	ch := newTestCharWithDesc()
	ch.PCData.Learned[gsn] = 50

	for i := 0; i < 200; i++ {
		learnFromSuccess(ch, gsn)
	}
	if ch.PCData.Learned[gsn] <= 50 {
		t.Errorf("skill should have improved from 50, got %d", ch.PCData.Learned[gsn])
	}
	if ch.PCData.Learned[gsn] > 95 {
		t.Errorf("skill should cap at adept=95, got %d", ch.PCData.Learned[gsn])
	}
}

func cleanupSkillTestRoom(ch *types.CharData, victim *types.CharData) {
	if victim != nil {
		handler.CharFromRoom(victim)
		WorldRef.RemoveChar(victim)
	}
}
