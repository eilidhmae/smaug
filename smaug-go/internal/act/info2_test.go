package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func setupInfo2World() *world.World {
	w := world.New("/tmp/test")
	WorldRef = w
	return w
}

func TestDoConsider_WeakerMob(t *testing.T) {
	_ = setupInfo2World()
	room := &types.RoomIndexData{Vnum: 7200, Name: "Arena"}
	ch, client := makeTestChar("Tester")
	defer client.Close()
	ch.Level = 20
	handler.CharToRoom(ch, room)

	mob := &types.CharData{Name: "rat", ShortDescr: "a rat", Level: 1}
	mob.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(mob, room)

	DoConsider(ch, "rat")
	output := readOutput(ch, client)

	if !strings.Contains(strings.ToLower(output), "not worth") &&
		!strings.Contains(strings.ToLower(output), "no contest") &&
		!strings.Contains(strings.ToLower(output), "easy") {
		t.Errorf("expected weak-mob message, got: %q", output)
	}
}

func TestDoConsider_StrongerMob(t *testing.T) {
	_ = setupInfo2World()
	room := &types.RoomIndexData{Vnum: 7201, Name: "Arena"}
	ch, client := makeTestChar("Tester")
	defer client.Close()
	ch.Level = 5
	handler.CharToRoom(ch, room)

	mob := &types.CharData{Name: "dragon", ShortDescr: "a dragon", Level: 50}
	mob.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(mob, room)

	DoConsider(ch, "dragon")
	output := readOutput(ch, client)

	if !strings.Contains(strings.ToLower(output), "death") && !strings.Contains(strings.ToLower(output), "run") {
		t.Errorf("expected death wish message for much stronger mob, got: %q", output)
	}
}

func TestDoConsider_NoArg(t *testing.T) {
	_ = setupInfo2World()
	room := &types.RoomIndexData{Vnum: 7202, Name: "Arena"}
	ch, client := makeTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	DoConsider(ch, "")
	output := readOutput(ch, client)

	if !strings.Contains(strings.ToLower(output), "consider") {
		t.Errorf("expected usage message, got: %q", output)
	}
}

func TestDoWhere(t *testing.T) {
	w := setupInfo2World()
	area := &types.AreaData{Name: "Test Area"}
	room1 := &types.RoomIndexData{Vnum: 7210, Name: "Room One", Area: area}
	room2 := &types.RoomIndexData{Vnum: 7211, Name: "Room Two", Area: area}
	w.Rooms[7210] = room1
	w.Rooms[7211] = room2

	ch, client := makeTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room1)
	w.AddChar(ch)

	mob := &types.CharData{Name: "guard soldier", ShortDescr: "a guard"}
	mob.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(mob, room2)
	w.AddChar(mob)

	DoWhere(ch, "guard")
	output := readOutput(ch, client)

	if !strings.Contains(output, "guard") && !strings.Contains(output, "Room Two") {
		t.Errorf("expected guard location, got: %q", output)
	}
}

func TestDoWhere_NotFound(t *testing.T) {
	w := setupInfo2World()
	room := &types.RoomIndexData{Vnum: 7212, Name: "Room"}
	w.Rooms[7212] = room
	ch, client := makeTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	DoWhere(ch, "dragon")
	output := readOutput(ch, client)

	if !strings.Contains(strings.ToLower(output), "no one") {
		t.Errorf("expected 'no one' message, got: %q", output)
	}
}

func TestDoTime(t *testing.T) {
	_ = setupInfo2World()
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoTime(ch, "")
	output := readOutput(ch, client)

	if !strings.Contains(output, "hour") || !strings.Contains(output, "day") {
		t.Errorf("expected time display with hour and day, got: %q", output)
	}
}
