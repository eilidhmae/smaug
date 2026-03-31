package act

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// makeTestChar creates a CharData with a net.Pipe-backed descriptor for output capture.
func makeTestChar(name string) (*types.CharData, net.Conn) {
	server, client := net.Pipe()
	d := &types.DescriptorData{
		Conn:       server,
		InputQueue: make(chan string, 10),
		Connected:  int(types.CON_PLAYING),
	}
	ch := &types.CharData{
		Name:       name,
		ShortDescr: name,
		Level:      10,
		Position:   types.POS_STANDING,
		Hit:        100,
		MaxHit:     100,
		Mana:       50,
		MaxMana:    50,
		Move:       80,
		MaxMove:    80,
		PermStr:    15,
		PermInt:    13,
		PermWis:    12,
		PermDex:    14,
		PermCon:    15,
		PermCha:    11,
		PermLck:    13,
		Gold:       500,
		Desc:       d,
		PCData: &types.PCData{
			Title: "the wizard",
		},
	}
	d.Character = ch
	return ch, client
}

// readOutput flushes the descriptor and reads the output.
func readOutput(ch *types.CharData, client net.Conn) string {
	if !ch.Desc.HasOutput() {
		return ""
	}
	result := make(chan string, 1)
	go func() {
		buf := make([]byte, 8192)
		client.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, _ := client.Read(buf)
		result <- string(buf[:n])
	}()
	_ = ch.Desc.FlushOutput()
	return <-result
}

func setupTestWorld() *world.World {
	w := world.New("/tmp/test")

	// Create temple room
	temple := &types.RoomIndexData{
		Vnum:        21001,
		Name:        "The Temple of Mota",
		Description: "You are in the temple.\n\r",
		SectorType:  types.SECT_INSIDE,
	}
	// Create town square
	square := &types.RoomIndexData{
		Vnum:        21002,
		Name:        "Town Square",
		Description: "The town square bustles with activity.\n\r",
		SectorType:  types.SECT_CITY,
		ExtraDescr: []*types.ExtraDescrData{
			{Keyword: "fountain", Description: "A marble fountain sprays water.\n\r"},
		},
	}

	// Connect with exits
	temple.Exits = []*types.ExitData{
		{Direction: types.DIR_NORTH, ToRoom: square, Vnum: 21002},
	}
	square.Exits = []*types.ExitData{
		{Direction: types.DIR_SOUTH, ToRoom: temple, Vnum: 21001},
	}

	w.Rooms[21001] = temple
	w.Rooms[21002] = square

	WorldRef = w
	return w
}

func TestDoLook_Room(t *testing.T) {
	w := setupTestWorld()
	ch, client := makeTestChar("Gandalf")
	defer client.Close()

	room := w.GetRoom(21001)
	ch.InRoom = room
	room.People = append(room.People, ch)

	DoLook(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "The Temple of Mota") {
		t.Errorf("output missing room name, got: %q", out)
	}
	if !strings.Contains(out, "You are in the temple.") {
		t.Errorf("output missing room description, got: %q", out)
	}
	if !strings.Contains(out, "north") {
		t.Errorf("output missing exit 'north', got: %q", out)
	}
}

func TestDoLook_AtCharacter(t *testing.T) {
	w := setupTestWorld()
	ch, client := makeTestChar("Gandalf")
	defer client.Close()

	room := w.GetRoom(21001)
	ch.InRoom = room
	room.People = append(room.People, ch)

	// Add NPC to room
	npc := &types.CharData{
		Name:        "guard",
		ShortDescr:  "a temple guard",
		Description: "The guard stands alert.\n\r",
	}
	npc.Act.Set(types.ACT_IS_NPC)
	npc.InRoom = room
	room.People = append(room.People, npc)

	DoLook(ch, "guard")
	out := readOutput(ch, client)

	if !strings.Contains(out, "a temple guard") {
		t.Errorf("output missing NPC short descr, got: %q", out)
	}
}

func TestDoLook_ExtraDesc(t *testing.T) {
	w := setupTestWorld()
	ch, client := makeTestChar("Gandalf")
	defer client.Close()

	room := w.GetRoom(21002)
	ch.InRoom = room
	room.People = append(room.People, ch)

	DoLook(ch, "fountain")
	out := readOutput(ch, client)

	if !strings.Contains(out, "marble fountain") {
		t.Errorf("output missing extra desc, got: %q", out)
	}
}

func TestDoLook_NotFound(t *testing.T) {
	setupTestWorld()
	ch, client := makeTestChar("Gandalf")
	defer client.Close()

	room := &types.RoomIndexData{Vnum: 99999, Name: "Empty Room"}
	ch.InRoom = room

	DoLook(ch, "nonexistent")
	out := readOutput(ch, client)

	if !strings.Contains(out, "You do not see that here") {
		t.Errorf("expected 'not see' message, got: %q", out)
	}
}

func TestDoScore(t *testing.T) {
	setupTestWorld()
	ch, client := makeTestChar("Gandalf")
	defer client.Close()

	DoScore(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Gandalf") {
		t.Errorf("output missing name, got: %q", out)
	}
	if !strings.Contains(out, "100/100") {
		t.Errorf("output missing HP, got: %q", out)
	}
	if !strings.Contains(out, "500") {
		t.Errorf("output missing gold, got: %q", out)
	}
}

func TestDoSay(t *testing.T) {
	setupTestWorld()
	ch, client := makeTestChar("Gandalf")
	defer client.Close()

	room := &types.RoomIndexData{Vnum: 99998, Name: "Test"}
	ch.InRoom = room
	room.People = []*types.CharData{ch}

	DoSay(ch, "Hello World")
	out := readOutput(ch, client)

	if !strings.Contains(out, "You say 'Hello World'") {
		t.Errorf("output missing say message, got: %q", out)
	}
}

func TestDoSay_Empty(t *testing.T) {
	setupTestWorld()
	ch, client := makeTestChar("Gandalf")
	defer client.Close()

	DoSay(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Say what?") {
		t.Errorf("expected 'Say what?' prompt, got: %q", out)
	}
}

func TestDoWho(t *testing.T) {
	w := setupTestWorld()
	ch, client := makeTestChar("Gandalf")
	defer client.Close()

	// Add to descriptors
	w.Descriptors = append(w.Descriptors, ch.Desc)

	DoWho(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Gandalf") {
		t.Errorf("output missing player name, got: %q", out)
	}
	if !strings.Contains(out, "the wizard") {
		t.Errorf("output missing title, got: %q", out)
	}
	if !strings.Contains(out, "1 player") {
		t.Errorf("output missing player count, got: %q", out)
	}
}

func TestDoInventory_Empty(t *testing.T) {
	setupTestWorld()
	ch, client := makeTestChar("Gandalf")
	defer client.Close()

	DoInventory(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Nothing") {
		t.Errorf("expected 'Nothing' message, got: %q", out)
	}
}

func TestDoInventory_WithItems(t *testing.T) {
	setupTestWorld()
	ch, client := makeTestChar("Gandalf")
	defer client.Close()

	obj := &types.ObjData{
		Name:       "magic staff",
		ShortDescr: "a gnarled staff",
		WearLoc:    types.WEAR_NONE,
	}
	ch.Carrying = append(ch.Carrying, obj)

	DoInventory(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "a gnarled staff") {
		t.Errorf("output missing item, got: %q", out)
	}
}

func TestDoEquipment_Empty(t *testing.T) {
	setupTestWorld()
	ch, client := makeTestChar("Gandalf")
	defer client.Close()

	DoEquipment(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Nothing") {
		t.Errorf("expected 'Nothing' message, got: %q", out)
	}
}

func TestDoEquipment_Worn(t *testing.T) {
	setupTestWorld()
	ch, client := makeTestChar("Gandalf")
	defer client.Close()

	obj := &types.ObjData{
		Name:       "iron shield",
		ShortDescr: "an iron shield",
		WearLoc:    types.WEAR_SHIELD,
	}
	ch.Carrying = append(ch.Carrying, obj)

	DoEquipment(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "an iron shield") {
		t.Errorf("output missing equipped item, got: %q", out)
	}
	if !strings.Contains(out, "shield") {
		t.Errorf("output missing wear location, got: %q", out)
	}
}

func TestDoQuit_Fighting(t *testing.T) {
	setupTestWorld()
	ch, client := makeTestChar("Gandalf")
	defer client.Close()

	ch.Position = types.POS_FIGHTING

	DoQuit(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "fighting") {
		t.Errorf("expected fighting message, got: %q", out)
	}
	// Should NOT be disconnected
	if ch.Desc.Connected == -1 {
		t.Error("should not disconnect while fighting")
	}
}

func TestMoveChar(t *testing.T) {
	w := setupTestWorld()
	ch, client := makeTestChar("Gandalf")
	defer client.Close()

	temple := w.GetRoom(21001)
	ch.InRoom = temple
	temple.People = append(temple.People, ch)

	MoveChar(ch, types.DIR_NORTH)
	out := readOutput(ch, client)

	if ch.InRoom != w.GetRoom(21002) {
		t.Error("character did not move to Town Square")
	}
	if !strings.Contains(out, "Town Square") {
		t.Errorf("output missing destination room name, got: %q", out)
	}
}

func TestMoveChar_Blocked(t *testing.T) {
	setupTestWorld()
	ch, client := makeTestChar("Gandalf")
	defer client.Close()

	room := &types.RoomIndexData{Vnum: 99997, Name: "Dead End"}
	ch.InRoom = room

	MoveChar(ch, types.DIR_EAST)
	out := readOutput(ch, client)

	if !strings.Contains(out, "cannot go that way") {
		t.Errorf("expected blocked message, got: %q", out)
	}
	if ch.InRoom != room {
		t.Error("character should not have moved")
	}
}

func TestShowExits(t *testing.T) {
	w := setupTestWorld()
	ch, client := makeTestChar("Gandalf")
	defer client.Close()

	room := w.GetRoom(21001)
	ch.InRoom = room

	showExits(ch, room)
	out := readOutput(ch, client)

	if !strings.Contains(out, "north") {
		t.Errorf("output missing 'north' exit, got: %q", out)
	}
}

func TestShowExits_None(t *testing.T) {
	setupTestWorld()
	ch, client := makeTestChar("Gandalf")
	defer client.Close()

	room := &types.RoomIndexData{Vnum: 99996, Name: "Sealed Room"}
	ch.InRoom = room

	showExits(ch, room)
	out := readOutput(ch, client)

	if !strings.Contains(out, "none") {
		t.Errorf("expected 'none' exits, got: %q", out)
	}
}
