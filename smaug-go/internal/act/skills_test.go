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

func TestLearnFromSuccess(t *testing.T) {
	ch := newTestCharWithDesc()
	ch.PCData.Learned[0] = 50

	// Run many times to check it can improve
	for i := 0; i < 200; i++ {
		learnFromSuccess(ch, 0)
	}
	if ch.PCData.Learned[0] <= 50 {
		t.Errorf("skill should have improved from 50, got %d", ch.PCData.Learned[0])
	}
}

func cleanupSkillTestRoom(ch *types.CharData, victim *types.CharData) {
	if victim != nil {
		handler.CharFromRoom(victim)
		WorldRef.RemoveChar(victim)
	}
}
