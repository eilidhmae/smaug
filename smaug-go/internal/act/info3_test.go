package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func setupInfo3World() *world.World {
	w := world.New("/tmp/test")
	WorldRef = w
	return w
}

// --- Direction commands (one-liners that call MoveChar) ---

func TestDoNorth(t *testing.T) {
	w := setupInfo3World()
	room1 := &types.RoomIndexData{Vnum: 8100, Name: "Start"}
	room2 := &types.RoomIndexData{Vnum: 8101, Name: "North Room"}
	room1.Exits = []*types.ExitData{{Direction: types.DIR_NORTH, ToRoom: room2}}
	w.Rooms[8100] = room1
	w.Rooms[8101] = room2

	ch, client := makeTestChar("Walker")
	defer client.Close()
	handler.CharToRoom(ch, room1)

	DoNorth(ch, "")
	_ = readOutput(ch, client)

	if ch.InRoom != room2 {
		t.Error("should have moved north")
	}
}

func TestDoSouth(t *testing.T) {
	w := setupInfo3World()
	room1 := &types.RoomIndexData{Vnum: 8102, Name: "Start"}
	room2 := &types.RoomIndexData{Vnum: 8103, Name: "South Room"}
	room1.Exits = []*types.ExitData{{Direction: types.DIR_SOUTH, ToRoom: room2}}
	w.Rooms[8102] = room1
	w.Rooms[8103] = room2

	ch, client := makeTestChar("Walker")
	defer client.Close()
	handler.CharToRoom(ch, room1)

	DoSouth(ch, "")
	_ = readOutput(ch, client)
	if ch.InRoom != room2 {
		t.Error("should have moved south")
	}
}

func TestDoEast(t *testing.T) {
	w := setupInfo3World()
	room1 := &types.RoomIndexData{Vnum: 8104, Name: "Start"}
	room2 := &types.RoomIndexData{Vnum: 8105, Name: "East Room"}
	room1.Exits = []*types.ExitData{{Direction: types.DIR_EAST, ToRoom: room2}}
	w.Rooms[8104] = room1
	w.Rooms[8105] = room2

	ch, client := makeTestChar("Walker")
	defer client.Close()
	handler.CharToRoom(ch, room1)

	DoEast(ch, "")
	_ = readOutput(ch, client)
	if ch.InRoom != room2 {
		t.Error("should have moved east")
	}
}

func TestDoWest(t *testing.T) {
	w := setupInfo3World()
	room1 := &types.RoomIndexData{Vnum: 8106, Name: "Start"}
	room2 := &types.RoomIndexData{Vnum: 8107, Name: "West Room"}
	room1.Exits = []*types.ExitData{{Direction: types.DIR_WEST, ToRoom: room2}}
	w.Rooms[8106] = room1
	w.Rooms[8107] = room2

	ch, client := makeTestChar("Walker")
	defer client.Close()
	handler.CharToRoom(ch, room1)

	DoWest(ch, "")
	_ = readOutput(ch, client)
	if ch.InRoom != room2 {
		t.Error("should have moved west")
	}
}

func TestDoUp(t *testing.T) {
	w := setupInfo3World()
	room1 := &types.RoomIndexData{Vnum: 8108, Name: "Start"}
	room2 := &types.RoomIndexData{Vnum: 8109, Name: "Up Room"}
	room1.Exits = []*types.ExitData{{Direction: types.DIR_UP, ToRoom: room2}}
	w.Rooms[8108] = room1
	w.Rooms[8109] = room2

	ch, client := makeTestChar("Walker")
	defer client.Close()
	handler.CharToRoom(ch, room1)

	DoUp(ch, "")
	_ = readOutput(ch, client)
	if ch.InRoom != room2 {
		t.Error("should have moved up")
	}
}

func TestDoDown(t *testing.T) {
	w := setupInfo3World()
	room1 := &types.RoomIndexData{Vnum: 8110, Name: "Start"}
	room2 := &types.RoomIndexData{Vnum: 8111, Name: "Down Room"}
	room1.Exits = []*types.ExitData{{Direction: types.DIR_DOWN, ToRoom: room2}}
	w.Rooms[8110] = room1
	w.Rooms[8111] = room2

	ch, client := makeTestChar("Walker")
	defer client.Close()
	handler.CharToRoom(ch, room1)

	DoDown(ch, "")
	_ = readOutput(ch, client)
	if ch.InRoom != room2 {
		t.Error("should have moved down")
	}
}

func TestDoNortheast(t *testing.T) {
	w := setupInfo3World()
	room1 := &types.RoomIndexData{Vnum: 8112, Name: "Start"}
	room2 := &types.RoomIndexData{Vnum: 8113, Name: "NE Room"}
	room1.Exits = []*types.ExitData{{Direction: types.DIR_NORTHEAST, ToRoom: room2}}
	w.Rooms[8112] = room1
	w.Rooms[8113] = room2

	ch, client := makeTestChar("Walker")
	defer client.Close()
	handler.CharToRoom(ch, room1)

	DoNortheast(ch, "")
	_ = readOutput(ch, client)
	if ch.InRoom != room2 {
		t.Error("should have moved northeast")
	}
}

func TestDoNorthwest(t *testing.T) {
	w := setupInfo3World()
	room1 := &types.RoomIndexData{Vnum: 8114, Name: "Start"}
	room2 := &types.RoomIndexData{Vnum: 8115, Name: "NW Room"}
	room1.Exits = []*types.ExitData{{Direction: types.DIR_NORTHWEST, ToRoom: room2}}
	w.Rooms[8114] = room1
	w.Rooms[8115] = room2

	ch, client := makeTestChar("Walker")
	defer client.Close()
	handler.CharToRoom(ch, room1)

	DoNorthwest(ch, "")
	_ = readOutput(ch, client)
	if ch.InRoom != room2 {
		t.Error("should have moved northwest")
	}
}

func TestDoSoutheast(t *testing.T) {
	w := setupInfo3World()
	room1 := &types.RoomIndexData{Vnum: 8116, Name: "Start"}
	room2 := &types.RoomIndexData{Vnum: 8117, Name: "SE Room"}
	room1.Exits = []*types.ExitData{{Direction: types.DIR_SOUTHEAST, ToRoom: room2}}
	w.Rooms[8116] = room1
	w.Rooms[8117] = room2

	ch, client := makeTestChar("Walker")
	defer client.Close()
	handler.CharToRoom(ch, room1)

	DoSoutheast(ch, "")
	_ = readOutput(ch, client)
	if ch.InRoom != room2 {
		t.Error("should have moved southeast")
	}
}

func TestDoSouthwest(t *testing.T) {
	w := setupInfo3World()
	room1 := &types.RoomIndexData{Vnum: 8118, Name: "Start"}
	room2 := &types.RoomIndexData{Vnum: 8119, Name: "SW Room"}
	room1.Exits = []*types.ExitData{{Direction: types.DIR_SOUTHWEST, ToRoom: room2}}
	w.Rooms[8118] = room1
	w.Rooms[8119] = room2

	ch, client := makeTestChar("Walker")
	defer client.Close()
	handler.CharToRoom(ch, room1)

	DoSouthwest(ch, "")
	_ = readOutput(ch, client)
	if ch.InRoom != room2 {
		t.Error("should have moved southwest")
	}
}

// --- DoExamine ---

func TestDoExamine_NoArg(t *testing.T) {
	_ = setupInfo3World()
	ch, client := makeTestChar("Tester")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 8200, Name: "Test"}

	DoExamine(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Examine what?") {
		t.Errorf("expected 'Examine what?', got: %q", out)
	}
}

func TestDoExamine_Container(t *testing.T) {
	w := setupInfo3World()
	room := &types.RoomIndexData{Vnum: 8201, Name: "Test Room"}
	w.Rooms[8201] = room

	ch, client := makeTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	container := &types.ObjData{
		Name:       "chest wooden",
		ShortDescr: "a wooden chest",
		ItemType:   types.ITEM_CONTAINER,
		InRoom:     room,
		WearLoc:    types.WEAR_NONE,
	}
	item := &types.ObjData{
		Name:       "ruby gem",
		ShortDescr: "a sparkling ruby",
		WearLoc:    types.WEAR_NONE,
	}
	container.Contents = append(container.Contents, item)
	room.Contents = append(room.Contents, container)

	DoExamine(ch, "chest")
	out := readOutput(ch, client)

	if !strings.Contains(out, "sparkling ruby") {
		t.Errorf("expected container contents, got: %q", out)
	}
}

func TestDoExamine_DrinkConEmpty(t *testing.T) {
	w := setupInfo3World()
	room := &types.RoomIndexData{Vnum: 8202, Name: "Test Room"}
	w.Rooms[8202] = room

	ch, client := makeTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	flask := &types.ObjData{
		Name:       "flask empty",
		ShortDescr: "an empty flask",
		ItemType:   types.ITEM_DRINK_CON,
		Value:      [6]int{10, 0, 0, 0, 0, 0},
		InRoom:     room,
		WearLoc:    types.WEAR_NONE,
	}
	room.Contents = append(room.Contents, flask)

	DoExamine(ch, "flask")
	out := readOutput(ch, client)

	if !strings.Contains(out, "empty") {
		t.Errorf("expected 'empty', got: %q", out)
	}
}

func TestDoExamine_DrinkConWithLiquid(t *testing.T) {
	w := setupInfo3World()
	room := &types.RoomIndexData{Vnum: 8203, Name: "Test Room"}
	w.Rooms[8203] = room

	ch, client := makeTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	flask := &types.ObjData{
		Name:       "flask water",
		ShortDescr: "a flask of water",
		ItemType:   types.ITEM_DRINK_CON,
		Value:      [6]int{10, 5, 0, 0, 0, 0},
		InRoom:     room,
		WearLoc:    types.WEAR_NONE,
	}
	room.Contents = append(room.Contents, flask)

	DoExamine(ch, "flask")
	out := readOutput(ch, client)

	if !strings.Contains(out, "liquid") {
		t.Errorf("expected 'liquid', got: %q", out)
	}
}

// --- DoCommands ---

func TestDoCommands(t *testing.T) {
	_ = setupInfo3World()
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoCommands(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "look") {
		t.Errorf("expected 'look' in commands, got: %q", out)
	}
}

// --- DoHelp ---

func TestDoHelp_NoArg(t *testing.T) {
	_ = setupInfo3World()
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoHelp(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "commands") {
		t.Errorf("expected 'commands' hint, got: %q", out)
	}
}

func TestDoHelp_Found(t *testing.T) {
	w := setupInfo3World()
	w.Helps = append(w.Helps, &types.HelpData{
		Keyword: "LOOK",
		Text:    "The look command shows you the room.",
	})

	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoHelp(ch, "look")
	out := readOutput(ch, client)

	if !strings.Contains(out, "look command") {
		t.Errorf("expected help text, got: %q", out)
	}
}

func TestDoHelp_NotFound(t *testing.T) {
	w := setupInfo3World()
	_ = w

	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoHelp(ch, "nonexistent")
	out := readOutput(ch, client)
	if !strings.Contains(out, "No help found") {
		t.Errorf("expected 'No help found', got: %q", out)
	}
}

// --- DoWeather ---

func TestDoWeather_NilRoom(t *testing.T) {
	_ = setupInfo3World()
	ch, client := makeTestChar("Tester")
	defer client.Close()
	ch.InRoom = nil

	DoWeather(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "can't see the weather") {
		t.Errorf("expected weather blocked, got: %q", out)
	}
}

func TestDoWeather_Indoors(t *testing.T) {
	_ = setupInfo3World()
	room := &types.RoomIndexData{Vnum: 8210, Name: "Indoor Room"}
	room.RoomFlags.Set(types.ROOM_INDOORS)

	ch, client := makeTestChar("Tester")
	defer client.Close()
	ch.InRoom = room

	DoWeather(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "indoors") {
		t.Errorf("expected indoors message, got: %q", out)
	}
}

func TestDoWeather_Outdoors(t *testing.T) {
	_ = setupInfo3World()
	room := &types.RoomIndexData{Vnum: 8211, Name: "Outdoor Room"}

	ch, client := makeTestChar("Tester")
	defer client.Close()
	ch.InRoom = room

	DoWeather(ch, "")
	out := readOutput(ch, client)
	// Should produce some weather output without crashing
	if out == "" {
		t.Error("expected some weather output")
	}
}

// --- DoPager ---

func TestDoPager_Toggle(t *testing.T) {
	_ = setupInfo3World()
	ch, client := makeTestChar("Tester")
	defer client.Close()

	// Initially off
	DoPager(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "enabled") {
		t.Errorf("expected 'enabled', got: %q", out)
	}

	// Toggle off
	DoPager(ch, "")
	out = readOutput(ch, client)
	if !strings.Contains(out, "disabled") {
		t.Errorf("expected 'disabled', got: %q", out)
	}
}

func TestDoPager_On(t *testing.T) {
	_ = setupInfo3World()
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPager(ch, "on")
	out := readOutput(ch, client)
	if !strings.Contains(out, "enabled") {
		t.Errorf("expected 'enabled', got: %q", out)
	}
	if ch.PCData.Flags&int(types.PCFLAG_PAGERON) == 0 {
		t.Error("pager flag should be set")
	}
}

func TestDoPager_Off(t *testing.T) {
	_ = setupInfo3World()
	ch, client := makeTestChar("Tester")
	defer client.Close()
	ch.PCData.Flags |= int(types.PCFLAG_PAGERON)

	DoPager(ch, "off")
	out := readOutput(ch, client)
	if !strings.Contains(out, "disabled") {
		t.Errorf("expected 'disabled', got: %q", out)
	}
}

func TestDoPager_SetLength(t *testing.T) {
	_ = setupInfo3World()
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPager(ch, "40")
	out := readOutput(ch, client)

	if ch.PCData.PagerLen != 40 {
		t.Errorf("pager length should be 40, got: %d", ch.PCData.PagerLen)
	}
	if !strings.Contains(out, "40") {
		t.Errorf("expected '40' in output, got: %q", out)
	}
}

func TestDoPager_InvalidLength(t *testing.T) {
	_ = setupInfo3World()
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPager(ch, "3")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Usage") {
		t.Errorf("expected usage message, got: %q", out)
	}
}

// --- DoShout ---

func TestDoShout_NoArg(t *testing.T) {
	_ = setupInfo3World()
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoShout(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Shout what?") {
		t.Errorf("expected 'Shout what?', got: %q", out)
	}
}

func TestDoShout_Broadcast(t *testing.T) {
	w := setupInfo3World()
	ch, chClient := makeTestChar("Shouter")
	defer chClient.Close()
	w.AddChar(ch)

	listener, listenerClient := makeTestChar("Listener")
	defer listenerClient.Close()
	w.AddChar(listener)

	DoShout(ch, "For the horde!")

	chOut := readOutput(ch, chClient)
	if !strings.Contains(chOut, "For the horde!") {
		t.Errorf("sender should see shout, got: %q", chOut)
	}

	listenerOut := readOutput(listener, listenerClient)
	if !strings.Contains(listenerOut, "For the horde!") {
		t.Errorf("listener should hear shout, got: %q", listenerOut)
	}
}

// --- DoPmote ---

func TestDoPmote_NoArg(t *testing.T) {
	_ = setupInfo3World()
	ch, client := makeTestChar("Tester")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 8220, Name: "Test"}

	DoPmote(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Pmote what?") {
		t.Errorf("expected 'Pmote what?', got: %q", out)
	}
}

func TestDoPmote_InRoom(t *testing.T) {
	_ = setupInfo3World()
	room := &types.RoomIndexData{Vnum: 8221, Name: "Test"}

	ch, chClient := makeTestChar("Actor")
	defer chClient.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	listener, listenerClient := makeTestChar("Watcher")
	defer listenerClient.Close()
	listener.InRoom = room
	room.People = append(room.People, listener)

	DoPmote(ch, "eyes glow red.")

	chOut := readOutput(ch, chClient)
	if !strings.Contains(chOut, "eyes glow red") {
		t.Errorf("sender should see pmote, got: %q", chOut)
	}

	listenerOut := readOutput(listener, listenerClient)
	if !strings.Contains(listenerOut, "Actor's eyes glow red") {
		t.Errorf("listener should see pmote, got: %q", listenerOut)
	}
}
