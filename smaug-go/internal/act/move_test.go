package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func setupMoveWorld() *world.World {
	w := world.New("/tmp/test")
	WorldRef = w
	return w
}

func TestDoOpen(t *testing.T) {
	_ = setupMoveWorld()
	room := &types.RoomIndexData{Vnum: 7100, Name: "Test Room"}
	room.Exits = []*types.ExitData{
		{Direction: types.DIR_NORTH, Keyword: "door", ExitInfo: int(types.EX_ISDOOR | types.EX_CLOSED),
			ToRoom: &types.RoomIndexData{Vnum: 7101, Name: "Other Room"}},
	}

	ch, client := makeTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	DoOpen(ch, "door")
	output := readOutput(ch, client)

	exit := room.Exits[0]
	if exit.ExitInfo&int(types.EX_CLOSED) != 0 {
		t.Error("door should be open after DoOpen")
	}
	if !strings.Contains(strings.ToLower(output), "open") {
		t.Errorf("expected open message, got: %q", output)
	}
}

func TestDoClose(t *testing.T) {
	_ = setupMoveWorld()
	room := &types.RoomIndexData{Vnum: 7102, Name: "Test Room"}
	room.Exits = []*types.ExitData{
		{Direction: types.DIR_NORTH, Keyword: "door", ExitInfo: int(types.EX_ISDOOR),
			ToRoom: &types.RoomIndexData{Vnum: 7103, Name: "Other Room"}},
	}

	ch, client := makeTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	DoClose(ch, "door")
	_ = readOutput(ch, client)

	exit := room.Exits[0]
	if exit.ExitInfo&int(types.EX_CLOSED) == 0 {
		t.Error("door should be closed after DoClose")
	}
}

func TestDoOpen_Locked(t *testing.T) {
	_ = setupMoveWorld()
	room := &types.RoomIndexData{Vnum: 7104, Name: "Test Room"}
	room.Exits = []*types.ExitData{
		{Direction: types.DIR_NORTH, Keyword: "door",
			ExitInfo: int(types.EX_ISDOOR | types.EX_CLOSED | types.EX_LOCKED),
			ToRoom:   &types.RoomIndexData{Vnum: 7105, Name: "Other Room"}},
	}

	ch, client := makeTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	DoOpen(ch, "door")
	output := readOutput(ch, client)

	// Should remain closed — it's locked
	exit := room.Exits[0]
	if exit.ExitInfo&int(types.EX_CLOSED) == 0 {
		t.Error("locked door should remain closed")
	}
	if !strings.Contains(strings.ToLower(output), "locked") {
		t.Errorf("expected 'locked' message, got: %q", output)
	}
}

func TestDoUnlock(t *testing.T) {
	w := setupMoveWorld()
	room := &types.RoomIndexData{Vnum: 7106, Name: "Test Room"}
	room.Exits = []*types.ExitData{
		{Direction: types.DIR_NORTH, Keyword: "door", Key: 5000,
			ExitInfo: int(types.EX_ISDOOR | types.EX_CLOSED | types.EX_LOCKED),
			ToRoom:   &types.RoomIndexData{Vnum: 7107, Name: "Other Room"}},
	}

	ch, client := makeTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	// Give player the key
	keyIdx := &types.ObjIndexData{Vnum: 5000, Name: "iron key", ShortDescr: "an iron key",
		ItemType: types.ITEM_KEY, WearFlags: int(types.ITEM_TAKE)}
	w.ObjIndex[5000] = keyIdx
	key := handler.CreateObject(w, keyIdx, 1)
	handler.ObjToChar(key, ch)

	DoUnlock(ch, "door")
	_ = readOutput(ch, client)

	exit := room.Exits[0]
	if exit.ExitInfo&int(types.EX_LOCKED) != 0 {
		t.Error("door should be unlocked after DoUnlock with key")
	}
}

func TestDoLock(t *testing.T) {
	w := setupMoveWorld()
	room := &types.RoomIndexData{Vnum: 7108, Name: "Test Room"}
	room.Exits = []*types.ExitData{
		{Direction: types.DIR_NORTH, Keyword: "door", Key: 5001,
			ExitInfo: int(types.EX_ISDOOR | types.EX_CLOSED),
			ToRoom:   &types.RoomIndexData{Vnum: 7109, Name: "Other Room"}},
	}

	ch, client := makeTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	keyIdx := &types.ObjIndexData{Vnum: 5001, Name: "brass key", ShortDescr: "a brass key",
		ItemType: types.ITEM_KEY, WearFlags: int(types.ITEM_TAKE)}
	w.ObjIndex[5001] = keyIdx
	key := handler.CreateObject(w, keyIdx, 1)
	handler.ObjToChar(key, ch)

	DoLock(ch, "door")
	_ = readOutput(ch, client)

	exit := room.Exits[0]
	if exit.ExitInfo&int(types.EX_LOCKED) == 0 {
		t.Error("door should be locked after DoLock with key")
	}
}

func TestMoveChar_SittingCantMove(t *testing.T) {
	_ = setupMoveWorld()
	room1 := &types.RoomIndexData{Vnum: 7110, Name: "Room 1"}
	room2 := &types.RoomIndexData{Vnum: 7111, Name: "Room 2"}
	room1.Exits = []*types.ExitData{
		{Direction: types.DIR_NORTH, ToRoom: room2},
	}

	ch, client := makeTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room1)
	ch.Position = types.POS_SITTING

	MoveChar(ch, types.DIR_NORTH)
	output := readOutput(ch, client)

	if ch.InRoom != room1 {
		t.Error("sitting character should not be able to move")
	}
	if !strings.Contains(strings.ToLower(output), "stand") || !strings.Contains(strings.ToLower(output), "first") {
		t.Errorf("expected 'stand up first' message, got: %q", output)
	}
}
