package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// buildRoomChain creates a linear chain of rooms connected east→west.
// Returns the slice of rooms (index 0 is westernmost).
func buildRoomChain(w *world.World, count int) []*types.RoomIndexData {
	rooms := make([]*types.RoomIndexData, count)
	for i := 0; i < count; i++ {
		rooms[i] = &types.RoomIndexData{
			Vnum: 8000 + i,
			Name: "Room",
			Area: &types.AreaData{},
		}
		w.Rooms[rooms[i].Vnum] = rooms[i]
	}
	// Link east/west
	for i := 0; i < count-1; i++ {
		eastExit := &types.ExitData{Direction: 1, ToRoom: rooms[i+1]} // east
		westExit := &types.ExitData{Direction: 3, ToRoom: rooms[i]}   // west
		rooms[i].Exits = append(rooms[i].Exits, eastExit)
		rooms[i+1].Exits = append(rooms[i+1].Exits, westExit)
	}
	// Set same area for all rooms
	area := &types.AreaData{}
	for _, r := range rooms {
		r.Area = area
	}
	return rooms
}

// --- BFSFindPath tests ---

func TestBFSFindPath_NilRooms(t *testing.T) {
	dir := BFSFindPath(nil, nil, 100)
	if dir != BFS_ERROR {
		t.Errorf("expected BFS_ERROR, got %d", dir)
	}
}

func TestBFSFindPath_AlreadyThere(t *testing.T) {
	w := world.New("/tmp/test")
	rooms := buildRoomChain(w, 3)

	dir := BFSFindPath(rooms[0], rooms[0], 100)
	if dir != BFS_ALREADY_THERE {
		t.Errorf("expected BFS_ALREADY_THERE, got %d", dir)
	}
}

func TestBFSFindPath_DirectNeighbor(t *testing.T) {
	w := world.New("/tmp/test")
	rooms := buildRoomChain(w, 3)

	// Room 0 → Room 1 should be east (dir 1)
	dir := BFSFindPath(rooms[0], rooms[1], 100)
	if dir != 1 {
		t.Errorf("expected direction 1 (east), got %d", dir)
	}

	// Room 1 → Room 0 should be west (dir 3)
	dir = BFSFindPath(rooms[1], rooms[0], 100)
	if dir != 3 {
		t.Errorf("expected direction 3 (west), got %d", dir)
	}
}

func TestBFSFindPath_MultipleHops(t *testing.T) {
	w := world.New("/tmp/test")
	rooms := buildRoomChain(w, 5)

	// Room 0 → Room 4 should start east
	dir := BFSFindPath(rooms[0], rooms[4], 100)
	if dir != 1 {
		t.Errorf("expected direction 1 (east), got %d", dir)
	}
}

func TestBFSFindPath_MaxDistance(t *testing.T) {
	w := world.New("/tmp/test")
	rooms := buildRoomChain(w, 10)

	// With maxDist=3, can't reach room 5 from room 0
	dir := BFSFindPath(rooms[0], rooms[5], 3)
	if dir != BFS_NO_PATH {
		t.Errorf("expected BFS_NO_PATH, got %d", dir)
	}

	// But can reach room 2
	dir = BFSFindPath(rooms[0], rooms[2], 3)
	if dir != 1 {
		t.Errorf("expected direction 1 (east), got %d", dir)
	}
}

func TestBFSFindPath_Disconnected(t *testing.T) {
	w := world.New("/tmp/test")
	area := &types.AreaData{}
	room1 := &types.RoomIndexData{Vnum: 7000, Name: "Island1", Area: area}
	room2 := &types.RoomIndexData{Vnum: 7001, Name: "Island2", Area: area}
	w.Rooms[7000] = room1
	w.Rooms[7001] = room2

	dir := BFSFindPath(room1, room2, 100)
	if dir != BFS_NO_PATH {
		t.Errorf("expected BFS_NO_PATH, got %d", dir)
	}
}

func TestBFSFindPath_DifferentAreas(t *testing.T) {
	room1 := &types.RoomIndexData{Vnum: 6000, Name: "Area1", Area: &types.AreaData{}}
	room2 := &types.RoomIndexData{Vnum: 6001, Name: "Area2", Area: &types.AreaData{}}

	dir := BFSFindPath(room1, room2, 100)
	if dir != BFS_NO_PATH {
		t.Errorf("expected BFS_NO_PATH for different areas, got %d", dir)
	}
}

func TestBFSFindPath_BranchingPath(t *testing.T) {
	w := world.New("/tmp/test")
	area := &types.AreaData{}

	// Build a T-shaped graph:
	//   center → north → target
	//   center → east → deadend
	center := &types.RoomIndexData{Vnum: 5000, Name: "Center", Area: area}
	north := &types.RoomIndexData{Vnum: 5001, Name: "North", Area: area}
	east := &types.RoomIndexData{Vnum: 5002, Name: "East", Area: area}
	target := &types.RoomIndexData{Vnum: 5003, Name: "Target", Area: area}

	center.Exits = []*types.ExitData{
		{Direction: 0, ToRoom: north}, // north
		{Direction: 1, ToRoom: east},  // east
	}
	north.Exits = []*types.ExitData{
		{Direction: 0, ToRoom: target}, // north
		{Direction: 2, ToRoom: center}, // south
	}
	east.Exits = []*types.ExitData{
		{Direction: 3, ToRoom: center}, // west
	}
	target.Exits = []*types.ExitData{
		{Direction: 2, ToRoom: north}, // south
	}

	w.Rooms[5000] = center
	w.Rooms[5001] = north
	w.Rooms[5002] = east
	w.Rooms[5003] = target

	// center → target should go north (dir 0)
	dir := BFSFindPath(center, target, 100)
	if dir != 0 {
		t.Errorf("expected direction 0 (north), got %d", dir)
	}
}

// --- DoTrack tests ---

func TestDoTrack_NoArg(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Tracker")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 9000, Name: "Room"}

	DoTrack(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Whom are you trying to track") {
		t.Errorf("expected track usage, got: %q", out)
	}
}

func TestDoTrack_NotFound(t *testing.T) {
	w := setupWizWorld()
	ch, client := makeTestChar("Tracker")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9000, Name: "Room"}
	w.Rooms[9000] = room
	handler.CharToRoom(ch, room)

	DoTrack(ch, "nobody")
	out := readOutput(ch, client)
	if !strings.Contains(out, "can't find") {
		t.Errorf("expected not found message, got: %q", out)
	}
}

func TestDoTrack_SameRoom(t *testing.T) {
	w := setupWizWorld()
	area := &types.AreaData{}
	room := &types.RoomIndexData{Vnum: 9000, Name: "Room", Area: area}
	w.Rooms[9000] = room

	ch, client := makeTestChar("Tracker")
	defer client.Close()
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	mob := &types.CharData{Name: "rabbit", ShortDescr: "a rabbit"}
	mob.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(mob, room)
	w.AddChar(mob)

	DoTrack(ch, "rabbit")
	out := readOutput(ch, client)
	if !strings.Contains(out, "already in the same room") {
		t.Errorf("expected same room message, got: %q", out)
	}
}

func TestDoTrack_Found(t *testing.T) {
	w := setupWizWorld()
	rooms := buildRoomChain(w, 3)

	ch, client := makeTestChar("Tracker")
	defer client.Close()
	handler.CharToRoom(ch, rooms[0])
	w.AddChar(ch)

	mob := &types.CharData{Name: "dragon", ShortDescr: "a dragon"}
	mob.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(mob, rooms[2])
	w.AddChar(mob)

	DoTrack(ch, "dragon")
	out := readOutput(ch, client)
	if !strings.Contains(out, "east") {
		t.Errorf("expected 'east' direction, got: %q", out)
	}
}

// --- HuntVictim tests ---

func TestHuntVictim_NilHunting(t *testing.T) {
	w := setupWizWorld()
	ch := &types.CharData{Name: "mob", Position: types.POS_STANDING}
	ch.Act.Set(types.ACT_IS_NPC)
	room := &types.RoomIndexData{Vnum: 9000, Name: "Room"}
	handler.CharToRoom(ch, room)

	// Should not panic with nil hunting
	HuntVictim(w, ch)
}

func TestHuntVictim_MovesTowardTarget(t *testing.T) {
	w := setupWizWorld()
	rooms := buildRoomChain(w, 3)

	hunter := &types.CharData{
		Name:     "wolf",
		Position: types.POS_STANDING,
		Level:    10,
		Move:     100,
		MaxMove:  100,
		Hit:      100,
		MaxHit:   100,
	}
	hunter.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(hunter, rooms[0])
	w.AddChar(hunter)

	victim := &types.CharData{Name: "prey", Position: types.POS_STANDING}
	handler.CharToRoom(victim, rooms[2])
	w.AddChar(victim)

	hunter.Hunting = &types.HHFData{Name: "prey", Who: victim}

	startRoom := hunter.InRoom
	HuntVictim(w, hunter)

	if hunter.InRoom == startRoom {
		t.Error("expected hunter to move, but stayed in same room")
	}
}

func TestHuntVictim_AlreadyInRoom(t *testing.T) {
	w := setupWizWorld()
	area := &types.AreaData{}
	room := &types.RoomIndexData{Vnum: 9000, Name: "Room", Area: area}
	w.Rooms[9000] = room

	hunter := &types.CharData{
		Name:     "wolf",
		Position: types.POS_STANDING,
		Level:    10,
	}
	hunter.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(hunter, room)
	w.AddChar(hunter)

	victim := &types.CharData{Name: "prey", Position: types.POS_STANDING}
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	hunter.Hunting = &types.HHFData{Name: "prey", Who: victim}

	HuntVictim(w, hunter)

	// Hunter should stay in room (found prey)
	if hunter.InRoom != room {
		t.Error("expected hunter to stay in room with prey")
	}
}
