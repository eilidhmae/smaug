package mudprog

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

func TestDoIfCheck_Rand(t *testing.T) {
	// rand($i) should succeed ~val% of the time
	trueCount := 0
	for i := 0; i < 100; i++ {
		if DoIfCheck("rand($i) == 50", nil, nil, nil, nil, nil) {
			trueCount++
		}
	}
	// Should be roughly 50% but we just verify it's not always true or false
	if trueCount == 0 || trueCount == 100 {
		t.Errorf("rand check seems broken: %d/100 true", trueCount)
	}
}

func TestDoIfCheck_IsPC(t *testing.T) {
	pc := &types.CharData{Name: "Player"}
	npc := &types.CharData{Name: "Mob"}
	npc.Act.Set(types.ACT_IS_NPC)

	if !DoIfCheck("ispc($n)", nil, pc, nil, nil, nil) {
		t.Error("ispc should return true for PC")
	}
	if DoIfCheck("ispc($n)", nil, npc, nil, nil, nil) {
		t.Error("ispc should return false for NPC")
	}
}

func TestDoIfCheck_Level(t *testing.T) {
	ch := &types.CharData{Name: "Test", Level: 20}

	if !DoIfCheck("level($n) > 10", nil, ch, nil, nil, nil) {
		t.Error("level 20 > 10 should be true")
	}
	if DoIfCheck("level($n) > 30", nil, ch, nil, nil, nil) {
		t.Error("level 20 > 30 should be false")
	}
	if !DoIfCheck("level($n) == 20", nil, ch, nil, nil, nil) {
		t.Error("level 20 == 20 should be true")
	}
}

func TestDoIfCheck_IsEvil(t *testing.T) {
	evil := &types.CharData{Name: "Evil", Alignment: -500}
	good := &types.CharData{Name: "Good", Alignment: 500}

	if !DoIfCheck("isevil($n)", nil, evil, nil, nil, nil) {
		t.Error("alignment -500 should be evil")
	}
	if DoIfCheck("isevil($n)", nil, good, nil, nil, nil) {
		t.Error("alignment 500 should not be evil")
	}
}

func TestDoIfCheck_HpPcnt(t *testing.T) {
	ch := &types.CharData{Name: "Test", Hit: 50, MaxHit: 100}

	if !DoIfCheck("hppcnt($n) == 50", nil, ch, nil, nil, nil) {
		t.Error("50/100 hp should be 50%")
	}
	if !DoIfCheck("hppcnt($n) < 75", nil, ch, nil, nil, nil) {
		t.Error("50% < 75% should be true")
	}
}

func TestDoIfCheck_IsFight(t *testing.T) {
	fighting := &types.CharData{
		Name:     "Fighter",
		Fighting: &types.FightData{Who: &types.CharData{Name: "Target"}},
	}
	idle := &types.CharData{Name: "Idle"}

	if !DoIfCheck("isfight($n)", nil, fighting, nil, nil, nil) {
		t.Error("fighting char should be isfight")
	}
	if DoIfCheck("isfight($n)", nil, idle, nil, nil, nil) {
		t.Error("idle char should not be isfight")
	}
}

func TestDoIfCheck_Name(t *testing.T) {
	ch := &types.CharData{Name: "Gandalf"}

	if !DoIfCheck("name($n) == Gandalf", nil, ch, nil, nil, nil) {
		t.Error("name should match")
	}
	if DoIfCheck("name($n) == Frodo", nil, ch, nil, nil, nil) {
		t.Error("name should not match")
	}
}
