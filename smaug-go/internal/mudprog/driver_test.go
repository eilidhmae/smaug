package mudprog

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

func TestDriver_SimpleCommand(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 9000, Name: "Test"}
	mob := &types.CharData{
		Name:     "TestMob",
		Level:    10,
		Position: types.POS_STANDING,
		InRoom:   room,
	}
	mob.Act.Set(types.ACT_IS_NPC)
	room.People = append(room.People, mob)

	actor := &types.CharData{
		Name:  "Player",
		Level: 5,
		Desc:  &types.DescriptorData{},
	}
	actor.Desc.Character = actor

	// Run a simple echo command
	Driver("mpecho Hello from mudprog!", mob, actor, nil, nil, nil, false)
	// Just verify it doesn't crash
}

func TestDriver_IfElse(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 9001, Name: "Test"}
	mob := &types.CharData{
		Name:     "TestMob",
		Level:    10,
		Position: types.POS_STANDING,
		InRoom:   room,
	}
	mob.Act.Set(types.ACT_IS_NPC)
	room.People = append(room.People, mob)

	actor := &types.CharData{
		Name:  "Player",
		Level: 20,
		Desc:  &types.DescriptorData{},
	}
	actor.Desc.Character = actor

	// If level > 10, echo "high level"
	prog := "if level($n) > 10\nmpecho High level!\nelse\nmpecho Low level!\nendif"
	Driver(prog, mob, actor, nil, nil, nil, false)
	// Should execute "High level!" path — just verify no crash
}

func TestDriver_NestedIf(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 9002, Name: "Test"}
	mob := &types.CharData{
		Name:     "TestMob",
		Level:    10,
		Position: types.POS_STANDING,
		InRoom:   room,
	}
	mob.Act.Set(types.ACT_IS_NPC)
	room.People = append(room.People, mob)

	actor := &types.CharData{Name: "Player", Level: 20}

	prog := "if ispc($n)\nif level($n) > 15\nmpecho High PC!\nendif\nendif"
	Driver(prog, mob, actor, nil, nil, nil, false)
}

func TestDriver_Break(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 9003, Name: "Test"}
	mob := &types.CharData{
		Name:     "TestMob",
		InRoom:   room,
		Position: types.POS_STANDING,
	}
	mob.Act.Set(types.ACT_IS_NPC)
	room.People = append(room.People, mob)

	prog := "mpecho line1\nbreak\nmpecho should not run"
	Driver(prog, mob, nil, nil, nil, nil, false)
}

func TestDriver_MaxNest(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 9004, Name: "Test"}
	mob := &types.CharData{
		Name:     "TestMob",
		InRoom:   room,
		Position: types.POS_STANDING,
	}
	mob.Act.Set(types.ACT_IS_NPC)

	// Reset nest counter
	progNest = 0

	// Call driver recursively — should not stack overflow
	for i := 0; i < 10; i++ {
		Driver("mpecho test", mob, nil, nil, nil, nil, false)
	}
}
