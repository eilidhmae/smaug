package world

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

func TestNew_MapsInitialized(t *testing.T) {
	w := New("/tmp/testdata")

	if w.Rooms == nil {
		t.Fatal("Rooms map is nil")
	}
	if w.MobIndex == nil {
		t.Fatal("MobIndex map is nil")
	}
	if w.ObjIndex == nil {
		t.Fatal("ObjIndex map is nil")
	}
	if w.DataDir != "/tmp/testdata" {
		t.Fatalf("expected DataDir %q, got %q", "/tmp/testdata", w.DataDir)
	}
}

func TestGetRoom(t *testing.T) {
	w := New("/tmp")
	room := &types.RoomIndexData{}
	w.Rooms[3001] = room

	got := w.GetRoom(3001)
	if got != room {
		t.Fatal("expected to get the room we added")
	}

	got = w.GetRoom(9999)
	if got != nil {
		t.Fatal("expected nil for unknown vnum")
	}
}

func TestGetMobIndex(t *testing.T) {
	w := New("/tmp")
	mob := &types.MobIndexData{}
	w.MobIndex[1001] = mob

	got := w.GetMobIndex(1001)
	if got != mob {
		t.Fatal("expected to get the mob index we added")
	}

	got = w.GetMobIndex(9999)
	if got != nil {
		t.Fatal("expected nil for unknown vnum")
	}
}

func TestGetObjIndex(t *testing.T) {
	w := New("/tmp")
	obj := &types.ObjIndexData{}
	w.ObjIndex[2001] = obj

	got := w.GetObjIndex(2001)
	if got != obj {
		t.Fatal("expected to get the obj index we added")
	}

	got = w.GetObjIndex(9999)
	if got != nil {
		t.Fatal("expected nil for unknown vnum")
	}
}

func TestAddRemoveChar(t *testing.T) {
	w := New("/tmp")
	a := &types.CharData{Name: "Alice"}
	b := &types.CharData{Name: "Bob"}
	c := &types.CharData{Name: "Charlie"}

	w.AddChar(a)
	w.AddChar(b)
	w.AddChar(c)

	if len(w.Characters) != 3 {
		t.Fatalf("expected 3 characters, got %d", len(w.Characters))
	}

	// Remove the middle one
	w.RemoveChar(b)
	if len(w.Characters) != 2 {
		t.Fatalf("expected 2 characters after removal, got %d", len(w.Characters))
	}
	for _, ch := range w.Characters {
		if ch == b {
			t.Fatal("Bob should have been removed")
		}
	}

	// Remove non-existent (should not panic)
	w.RemoveChar(&types.CharData{Name: "Nobody"})
	if len(w.Characters) != 2 {
		t.Fatalf("expected 2 characters after no-op removal, got %d", len(w.Characters))
	}
}

func TestAddRemoveObj(t *testing.T) {
	w := New("/tmp")
	a := &types.ObjData{}
	b := &types.ObjData{}
	c := &types.ObjData{}

	w.AddObj(a)
	w.AddObj(b)
	w.AddObj(c)

	if len(w.Objects) != 3 {
		t.Fatalf("expected 3 objects, got %d", len(w.Objects))
	}

	// Remove the middle one
	w.RemoveObj(b)
	if len(w.Objects) != 2 {
		t.Fatalf("expected 2 objects after removal, got %d", len(w.Objects))
	}
	for _, o := range w.Objects {
		if o == b {
			t.Fatal("object b should have been removed")
		}
	}

	// Remove non-existent (should not panic)
	w.RemoveObj(&types.ObjData{})
	if len(w.Objects) != 2 {
		t.Fatalf("expected 2 objects after no-op removal, got %d", len(w.Objects))
	}
}

func TestFixExits_ResolvesVnums(t *testing.T) {
	w := New("/tmp")
	room1 := &types.RoomIndexData{Vnum: 100}
	room2 := &types.RoomIndexData{Vnum: 200}
	room3 := &types.RoomIndexData{Vnum: 300}

	// room1 has an exit to room2 by vnum (ToRoom not yet set)
	room1.Exits = []*types.ExitData{
		{Direction: types.DIR_NORTH, Vnum: 200, ToRoom: nil},
	}
	// room2 has an exit to room3 by vnum
	room2.Exits = []*types.ExitData{
		{Direction: types.DIR_SOUTH, Vnum: 100, ToRoom: nil},
		{Direction: types.DIR_EAST, Vnum: 300, ToRoom: nil},
	}

	w.Rooms[100] = room1
	w.Rooms[200] = room2
	w.Rooms[300] = room3

	w.FixExits()

	if room1.Exits[0].ToRoom != room2 {
		t.Error("room1 north exit should point to room2")
	}
	if room2.Exits[0].ToRoom != room1 {
		t.Error("room2 south exit should point to room1")
	}
	if room2.Exits[1].ToRoom != room3 {
		t.Error("room2 east exit should point to room3")
	}
}

func TestFixExits_BrokenVnum(t *testing.T) {
	w := New("/tmp")
	room1 := &types.RoomIndexData{Vnum: 100}
	room1.Exits = []*types.ExitData{
		{Direction: types.DIR_NORTH, Vnum: 99999, ToRoom: nil}, // non-existent destination
	}
	w.Rooms[100] = room1

	// Should not panic; broken exit remains nil
	w.FixExits()

	if room1.Exits[0].ToRoom != nil {
		t.Error("broken exit should remain nil")
	}
}

func TestFixExits_AlreadyResolved(t *testing.T) {
	w := New("/tmp")
	room1 := &types.RoomIndexData{Vnum: 100}
	room2 := &types.RoomIndexData{Vnum: 200}

	// Exit already has ToRoom set (should not be re-resolved)
	room1.Exits = []*types.ExitData{
		{Direction: types.DIR_NORTH, Vnum: 200, ToRoom: room2},
	}
	w.Rooms[100] = room1
	w.Rooms[200] = room2

	w.FixExits()

	if room1.Exits[0].ToRoom != room2 {
		t.Error("already-resolved exit should remain pointing to room2")
	}
}

func TestFixExits_ZeroVnum(t *testing.T) {
	w := New("/tmp")
	room1 := &types.RoomIndexData{Vnum: 100}
	room1.Exits = []*types.ExitData{
		{Direction: types.DIR_NORTH, Vnum: 0, ToRoom: nil}, // Vnum 0 should not be resolved
	}
	w.Rooms[100] = room1

	w.FixExits()

	if room1.Exits[0].ToRoom != nil {
		t.Error("exit with vnum 0 should not be resolved")
	}
}

func TestFixExits_NoExits(t *testing.T) {
	w := New("/tmp")
	room := &types.RoomIndexData{Vnum: 100}
	w.Rooms[100] = room

	// Should not panic on room with no exits
	w.FixExits()
}

func TestAddObj_Multiple(t *testing.T) {
	w := New("/tmp")
	objs := make([]*types.ObjData, 5)
	for i := range objs {
		objs[i] = &types.ObjData{Name: "obj"}
		w.AddObj(objs[i])
	}
	if len(w.Objects) != 5 {
		t.Fatalf("expected 5 objects, got %d", len(w.Objects))
	}
	// Remove from the front
	w.RemoveObj(objs[0])
	if len(w.Objects) != 4 {
		t.Fatalf("expected 4 objects after remove, got %d", len(w.Objects))
	}
	if w.Objects[0] == objs[0] {
		t.Error("first object should have been removed")
	}
}

func TestRemoveObj_Last(t *testing.T) {
	w := New("/tmp")
	a := &types.ObjData{Name: "a"}
	b := &types.ObjData{Name: "b"}
	w.AddObj(a)
	w.AddObj(b)

	// Remove last element
	w.RemoveObj(b)
	if len(w.Objects) != 1 {
		t.Fatalf("expected 1 object, got %d", len(w.Objects))
	}
	if w.Objects[0] != a {
		t.Error("remaining object should be 'a'")
	}
}

func TestRemoveChar_First(t *testing.T) {
	w := New("/tmp")
	a := &types.CharData{Name: "a"}
	b := &types.CharData{Name: "b"}
	c := &types.CharData{Name: "c"}
	w.AddChar(a)
	w.AddChar(b)
	w.AddChar(c)

	w.RemoveChar(a)
	if len(w.Characters) != 2 {
		t.Fatalf("expected 2 characters, got %d", len(w.Characters))
	}
	if w.Characters[0] != b {
		t.Error("first element should be 'b' after removing 'a'")
	}
}

func TestGetMobIndex_NotFound(t *testing.T) {
	w := New("/tmp")
	if w.GetMobIndex(12345) != nil {
		t.Error("should return nil for non-existent mob index")
	}
}

func TestGetObjIndex_NotFound(t *testing.T) {
	w := New("/tmp")
	if w.GetObjIndex(12345) != nil {
		t.Error("should return nil for non-existent obj index")
	}
}
