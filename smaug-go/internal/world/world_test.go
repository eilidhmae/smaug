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
