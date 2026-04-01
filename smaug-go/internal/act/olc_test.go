package act

import (
	"net"
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func setupOlcWorld() *world.World {
	w := world.New("/tmp/test")
	WorldRef = w
	return w
}

// makeImmTestChar creates a test character with immortal trust level for OLC tests.
func makeImmTestChar(name string) (*types.CharData, net.Conn) {
	ch, client := makeTestChar(name)
	ch.Level = types.LEVEL_IMMORTAL
	return ch, client
}

// --- DoRlist ---

func TestDoRlist_WithRange(t *testing.T) {
	w := setupOlcWorld()
	w.Rooms[100] = &types.RoomIndexData{Vnum: 100, Name: "Room Alpha"}
	w.Rooms[105] = &types.RoomIndexData{Vnum: 105, Name: "Room Beta"}
	w.Rooms[200] = &types.RoomIndexData{Vnum: 200, Name: "Room Gamma"}

	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoRlist(ch, "100 110")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Room Alpha") {
		t.Errorf("expected 'Room Alpha', got: %q", out)
	}
	if !strings.Contains(out, "Room Beta") {
		t.Errorf("expected 'Room Beta', got: %q", out)
	}
	if strings.Contains(out, "Room Gamma") {
		t.Errorf("should not contain 'Room Gamma' (outside range), got: %q", out)
	}
}

func TestDoRlist_Empty(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoRlist(ch, "9000 9100")
	out := readOutput(ch, client)

	if !strings.Contains(out, "No rooms in that range") {
		t.Errorf("expected 'No rooms in that range', got: %q", out)
	}
}

func TestDoRlist_NoArgNoArea(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 1, Name: "Test"}

	DoRlist(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Usage") {
		t.Errorf("expected usage message, got: %q", out)
	}
}

func TestDoRlist_DefaultArea(t *testing.T) {
	w := setupOlcWorld()
	area := &types.AreaData{Name: "Test Area", LowRVnum: 50, HiRVnum: 60}
	w.Rooms[55] = &types.RoomIndexData{Vnum: 55, Name: "Area Room", Area: area}

	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 50, Name: "Start", Area: area}

	DoRlist(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Area Room") {
		t.Errorf("expected 'Area Room', got: %q", out)
	}
}

func TestDoRlist_SingleVnum(t *testing.T) {
	w := setupOlcWorld()
	w.Rooms[300] = &types.RoomIndexData{Vnum: 300, Name: "Room 300"}
	w.Rooms[350] = &types.RoomIndexData{Vnum: 350, Name: "Room 350"}

	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	// When only one arg given, high = low + 100
	DoRlist(ch, "300")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Room 300") {
		t.Errorf("expected 'Room 300', got: %q", out)
	}
	if !strings.Contains(out, "Room 350") {
		t.Errorf("expected 'Room 350', got: %q", out)
	}
}

// --- DoOlist ---

func TestDoOlist_WithRange(t *testing.T) {
	w := setupOlcWorld()
	w.ObjIndex[100] = &types.ObjIndexData{Vnum: 100, Name: "sword", ShortDescr: "a steel sword"}
	w.ObjIndex[105] = &types.ObjIndexData{Vnum: 105, Name: "shield", ShortDescr: "an oak shield"}
	w.ObjIndex[200] = &types.ObjIndexData{Vnum: 200, Name: "potion", ShortDescr: "a healing potion"}

	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoOlist(ch, "100 110")
	out := readOutput(ch, client)

	if !strings.Contains(out, "steel sword") {
		t.Errorf("expected 'steel sword', got: %q", out)
	}
	if !strings.Contains(out, "oak shield") {
		t.Errorf("expected 'oak shield', got: %q", out)
	}
	if strings.Contains(out, "healing potion") {
		t.Errorf("should not contain 'healing potion' (outside range), got: %q", out)
	}
}

func TestDoOlist_Empty(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoOlist(ch, "9000 9100")
	out := readOutput(ch, client)

	if !strings.Contains(out, "No objects in that range") {
		t.Errorf("expected 'No objects in that range', got: %q", out)
	}
}

func TestDoOlist_NoArgNoArea(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 1, Name: "Test"}

	DoOlist(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Usage") {
		t.Errorf("expected usage message, got: %q", out)
	}
}

func TestDoOlist_DefaultArea(t *testing.T) {
	w := setupOlcWorld()
	area := &types.AreaData{Name: "Test Area", LowOVnum: 50, HiOVnum: 60}
	w.ObjIndex[55] = &types.ObjIndexData{Vnum: 55, Name: "gem", ShortDescr: "a shiny gem"}

	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 50, Name: "Start", Area: area}

	DoOlist(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "shiny gem") {
		t.Errorf("expected 'shiny gem', got: %q", out)
	}
}

// --- DoMlist ---

func TestDoMlist_WithRange(t *testing.T) {
	w := setupOlcWorld()
	w.MobIndex[100] = &types.MobIndexData{Vnum: 100, PlayerName: "guard", ShortDescr: "a city guard"}
	w.MobIndex[105] = &types.MobIndexData{Vnum: 105, PlayerName: "thief", ShortDescr: "a sneaky thief"}
	w.MobIndex[200] = &types.MobIndexData{Vnum: 200, PlayerName: "dragon", ShortDescr: "a red dragon"}

	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoMlist(ch, "100 110")
	out := readOutput(ch, client)

	if !strings.Contains(out, "city guard") {
		t.Errorf("expected 'city guard', got: %q", out)
	}
	if !strings.Contains(out, "sneaky thief") {
		t.Errorf("expected 'sneaky thief', got: %q", out)
	}
	if strings.Contains(out, "red dragon") {
		t.Errorf("should not contain 'red dragon' (outside range), got: %q", out)
	}
}

func TestDoMlist_Empty(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoMlist(ch, "9000 9100")
	out := readOutput(ch, client)

	if !strings.Contains(out, "No mobs in that range") {
		t.Errorf("expected 'No mobs in that range', got: %q", out)
	}
}

func TestDoMlist_NoArgNoArea(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 1, Name: "Test"}

	DoMlist(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Usage") {
		t.Errorf("expected usage message, got: %q", out)
	}
}

func TestDoMlist_DefaultArea(t *testing.T) {
	w := setupOlcWorld()
	area := &types.AreaData{Name: "Test Area", LowMVnum: 50, HiMVnum: 60}
	w.MobIndex[55] = &types.MobIndexData{Vnum: 55, PlayerName: "rat", ShortDescr: "a sewer rat"}

	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 50, Name: "Start", Area: area}

	DoMlist(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "sewer rat") {
		t.Errorf("expected 'sewer rat', got: %q", out)
	}
}

// --- DoOcreate ---

func TestDoOcreate_NoArg(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoOcreate(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Usage") {
		t.Errorf("expected usage, got: %q", out)
	}
}

func TestDoOcreate_Success(t *testing.T) {
	w := setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.Level = types.LEVEL_IMMORTAL

	DoOcreate(ch, "5000 magic wand")
	out := readOutput(ch, client)

	if !strings.Contains(out, "5000") {
		t.Errorf("expected vnum in output, got: %q", out)
	}
	if !strings.Contains(out, "magic wand") {
		t.Errorf("expected name in output, got: %q", out)
	}
	if _, exists := w.ObjIndex[5000]; !exists {
		t.Error("object template should be created in world")
	}
	found := false
	for _, o := range ch.Carrying {
		if strings.Contains(o.Name, "magic wand") {
			found = true
		}
	}
	if !found {
		t.Error("object instance should be in builder's inventory")
	}
}

func TestDoOcreate_DuplicateVnum(t *testing.T) {
	w := setupOlcWorld()
	w.ObjIndex[5001] = &types.ObjIndexData{Vnum: 5001, Name: "existing"}

	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoOcreate(ch, "5001 duplicate")
	out := readOutput(ch, client)

	if !strings.Contains(out, "already exists") {
		t.Errorf("expected 'already exists', got: %q", out)
	}
}

// --- DoMcreate ---

func TestDoMcreate_NoArg(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoMcreate(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Usage") {
		t.Errorf("expected usage, got: %q", out)
	}
}

func TestDoMcreate_Success(t *testing.T) {
	w := setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 7100, Name: "Build Room"}
	w.Rooms[7100] = room

	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	handler.CharToRoom(ch, room)

	DoMcreate(ch, "6000 fire elemental")
	out := readOutput(ch, client)

	if !strings.Contains(out, "6000") {
		t.Errorf("expected vnum in output, got: %q", out)
	}
	if _, exists := w.MobIndex[6000]; !exists {
		t.Error("mob template should be created in world")
	}
}

func TestDoMcreate_DuplicateVnum(t *testing.T) {
	w := setupOlcWorld()
	w.MobIndex[6001] = &types.MobIndexData{Vnum: 6001, PlayerName: "existing"}

	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoMcreate(ch, "6001 duplicate")
	out := readOutput(ch, client)

	if !strings.Contains(out, "already exists") {
		t.Errorf("expected 'already exists', got: %q", out)
	}
}

// --- DoRedit ---

func TestDoRedit_NoArg(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 7200, Name: "Build Room"}

	DoRedit(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Redit what?") {
		t.Errorf("expected 'Redit what?', got: %q", out)
	}
}

func TestDoRedit_SetName(t *testing.T) {
	_ = setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 7201, Name: "Old Name"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "name The Grand Hall")
	out := readOutput(ch, client)

	if room.Name != "The Grand Hall" {
		t.Errorf("room name should be changed, got: %q", room.Name)
	}
	if !strings.Contains(out, "The Grand Hall") {
		t.Errorf("expected confirmation, got: %q", out)
	}
}

func TestDoRedit_ShowName(t *testing.T) {
	_ = setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 7202, Name: "Current Name"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "name")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Current Name") {
		t.Errorf("expected current name shown, got: %q", out)
	}
}

func TestDoRedit_SetSector(t *testing.T) {
	_ = setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 7203, Name: "Test", SectorType: types.SECT_INSIDE}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "sector 2")
	out := readOutput(ch, client)

	if room.SectorType != 2 {
		t.Errorf("sector should be 2, got: %d", room.SectorType)
	}
	if !strings.Contains(out, "Sector set to 2") {
		t.Errorf("expected confirmation, got: %q", out)
	}
}

func TestDoRedit_AddExdesc(t *testing.T) {
	_ = setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 7204, Name: "Test"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "exdesc fountain A beautiful marble fountain.")
	out := readOutput(ch, client)

	if len(room.ExtraDescr) != 1 {
		t.Fatalf("expected 1 extra desc, got %d", len(room.ExtraDescr))
	}
	if room.ExtraDescr[0].Keyword != "fountain" {
		t.Errorf("expected keyword 'fountain', got: %q", room.ExtraDescr[0].Keyword)
	}
	if !strings.Contains(out, "fountain") {
		t.Errorf("expected confirmation, got: %q", out)
	}
}

func TestDoRedit_NilRoom(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = nil

	DoRedit(ch, "name test")
	out := readOutput(ch, client)
	if !strings.Contains(out, "not in a room") {
		t.Errorf("expected 'not in a room', got: %q", out)
	}
}

// --- DoRdig ---

func TestDoRdig_NoArg(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoRdig(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Usage") {
		t.Errorf("expected usage, got: %q", out)
	}
}

func TestDoRdig_Success(t *testing.T) {
	w := setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 7300, Name: "Start Room"}
	w.Rooms[7300] = room

	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	handler.CharToRoom(ch, room)

	DoRdig(ch, "north 7301")
	out := readOutput(ch, client)

	if !strings.Contains(out, "7301") {
		t.Errorf("expected vnum in output, got: %q", out)
	}
	newRoom := w.GetRoom(7301)
	if newRoom == nil {
		t.Fatal("new room should exist")
	}
	// Check exit from origin to new room
	found := false
	for _, ex := range room.Exits {
		if ex.Direction == types.DIR_NORTH && ex.ToRoom == newRoom {
			found = true
		}
	}
	if !found {
		t.Error("origin room should have north exit to new room")
	}
	// Check reverse exit
	foundReverse := false
	for _, ex := range newRoom.Exits {
		if ex.Direction == types.DIR_SOUTH && ex.ToRoom == room {
			foundReverse = true
		}
	}
	if !foundReverse {
		t.Error("new room should have south exit back to origin")
	}
}

// --- editExit ---

func TestEditExit_NoArg(t *testing.T) {
	_ = setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 7400, Name: "Test"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "exit")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Usage") {
		t.Errorf("expected usage, got: %q", out)
	}
}

func TestEditExit_SetVnum(t *testing.T) {
	w := setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 7401, Name: "Start"}
	dest := &types.RoomIndexData{Vnum: 7402, Name: "Destination"}
	w.Rooms[7401] = room
	w.Rooms[7402] = dest

	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "exit north 7402")
	out := readOutput(ch, client)

	if !strings.Contains(out, "7402") {
		t.Errorf("expected vnum in output, got: %q", out)
	}

	found := false
	for _, ex := range room.Exits {
		if ex.Direction == types.DIR_NORTH && ex.ToRoom == dest {
			found = true
		}
	}
	if !found {
		t.Error("exit should be created")
	}
}

func TestEditExit_Delete(t *testing.T) {
	_ = setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 7403, Name: "Test"}
	room.Exits = []*types.ExitData{
		{Direction: types.DIR_NORTH, ToRoom: &types.RoomIndexData{Vnum: 1}},
	}

	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "exit north delete")
	out := readOutput(ch, client)

	if !strings.Contains(out, "deleted") {
		t.Errorf("expected 'deleted', got: %q", out)
	}
	if len(room.Exits) != 0 {
		t.Errorf("exit should be removed, got %d exits", len(room.Exits))
	}
}

func TestEditExit_InvalidDirection(t *testing.T) {
	_ = setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 7404, Name: "Test"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "exit baddir 100")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Invalid direction") {
		t.Errorf("expected 'Invalid direction', got: %q", out)
	}
}

func TestEditExit_BadDestVnum(t *testing.T) {
	w := setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 7405, Name: "Test"}
	w.Rooms[7405] = room

	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "exit north 99999")
	out := readOutput(ch, client)
	if !strings.Contains(out, "does not exist") {
		t.Errorf("expected 'does not exist', got: %q", out)
	}
}

func TestDoRdig_ExistingVnum(t *testing.T) {
	w := setupOlcWorld()
	w.Rooms[7302] = &types.RoomIndexData{Vnum: 7302, Name: "Existing"}

	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 7303, Name: "Start"}

	DoRdig(ch, "north 7302")
	out := readOutput(ch, client)
	if !strings.Contains(out, "already exists") {
		t.Errorf("expected 'already exists', got: %q", out)
	}
}

func TestOLC_DefenseInDepth(t *testing.T) {
	_ = setupOlcWorld()

	// Create a mortal character (level 1, no trust override)
	mortal, client := makeTestChar("Mortal")
	defer client.Close()
	mortal.Level = 1
	mortal.InRoom = &types.RoomIndexData{Vnum: 100, Name: "Original Name"}

	tests := []struct {
		name string
		fn   func(ch *types.CharData, argument string)
		arg  string
	}{
		{"DoRedit", DoRedit, "name Hacked Room"},
		{"DoOcreate", DoOcreate, "9999 hacked object"},
		{"DoMcreate", DoMcreate, "9999 hacked mob"},
		{"DoRdig", DoRdig, "north 9999"},
		{"DoRlist", DoRlist, "100 200"},
		{"DoOlist", DoOlist, "100 200"},
		{"DoMlist", DoMlist, "100 200"},
		{"DoSaveArea", DoSaveArea, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.fn(mortal, tc.arg)
			out := readOutput(mortal, client)
			if !strings.Contains(out, "Huh?") {
				t.Errorf("%s should reject mortal with 'Huh?', got: %q", tc.name, out)
			}
		})
	}

	// Verify DoRedit didn't modify the room
	if mortal.InRoom.Name != "Original Name" {
		t.Errorf("room name should be unchanged, got: %q", mortal.InRoom.Name)
	}
}
