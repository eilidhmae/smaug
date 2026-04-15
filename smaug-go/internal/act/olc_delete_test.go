package act

import (
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

// --- DoRdelete ---

func TestDoRdelete_NoArg(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	DoRdelete(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Usage") {
		t.Errorf("expected usage, got: %q", out)
	}
}

func TestDoRdelete_BadVnum(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	DoRdelete(ch, "99999")
	out := readOutput(ch, client)
	if !strings.Contains(out, "does not exist") {
		t.Errorf("expected does-not-exist, got: %q", out)
	}
}

func TestDoRdelete_RequiresConfirm(t *testing.T) {
	w := setupOlcWorld()
	target := &types.RoomIndexData{Vnum: 9500, Name: "Target"}
	here := &types.RoomIndexData{Vnum: 9501, Name: "Here"}
	w.Rooms[9500] = target
	w.Rooms[9501] = here
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = here

	DoRdelete(ch, "9500")
	out := readOutput(ch, client)
	if !strings.Contains(out, "again within") {
		t.Errorf("expected confirm-prompt, got: %q", out)
	}
	if _, ok := w.Rooms[9500]; !ok {
		t.Error("room should still exist after first invocation")
	}
}

func TestDoRdelete_ConfirmCommits(t *testing.T) {
	w := setupOlcWorld()
	target := &types.RoomIndexData{Vnum: 9502, Name: "Target"}
	here := &types.RoomIndexData{Vnum: 9503, Name: "Here"}
	w.Rooms[9502] = target
	w.Rooms[9503] = here
	// Another room with an exit to target, to verify scrub
	peer := &types.RoomIndexData{Vnum: 9504, Name: "Peer"}
	peer.Exits = []*types.ExitData{{Direction: types.DIR_NORTH, ToRoom: target}}
	w.Rooms[9504] = peer
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = here

	DoRdelete(ch, "9502")
	_ = readOutput(ch, client)
	DoRdelete(ch, "9502")
	out := readOutput(ch, client)

	if !strings.Contains(out, "deleted") {
		t.Errorf("expected deleted message, got: %q", out)
	}
	if _, ok := w.Rooms[9502]; ok {
		t.Error("room should be removed")
	}
	if len(peer.Exits) != 0 {
		t.Errorf("peer exits should be scrubbed, got %d", len(peer.Exits))
	}
}

func TestDoRdelete_CannotDeleteCurrentRoom(t *testing.T) {
	w := setupOlcWorld()
	here := &types.RoomIndexData{Vnum: 9510, Name: "Here"}
	w.Rooms[9510] = here
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = here

	DoRdelete(ch, "9510")
	out := readOutput(ch, client)
	if !strings.Contains(out, "standing in") {
		t.Errorf("expected standing-in guard, got: %q", out)
	}
}

func TestDoRdelete_ConfirmExpires(t *testing.T) {
	w := setupOlcWorld()
	target := &types.RoomIndexData{Vnum: 9520, Name: "Target"}
	here := &types.RoomIndexData{Vnum: 9521, Name: "Here"}
	w.Rooms[9520] = target
	w.Rooms[9521] = here
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = here

	// Override clock to simulate stale confirmation.
	base := time.Now()
	nowFunc = func() time.Time { return base }
	defer func() { nowFunc = func() time.Time { return time.Now() } }()

	DoRdelete(ch, "9520")
	_ = readOutput(ch, client)

	// Advance past window
	nowFunc = func() time.Time { return base.Add(deleteConfirmWindow + time.Second) }
	DoRdelete(ch, "9520")
	out := readOutput(ch, client)
	if !strings.Contains(out, "again within") {
		t.Errorf("stale confirmation should re-prompt, got: %q", out)
	}
	if _, ok := w.Rooms[9520]; !ok {
		t.Error("room should still exist")
	}
}

// --- DoOdelete ---

func TestDoOdelete_ScrubLiveInstances(t *testing.T) {
	w := setupOlcWorld()
	idx := &types.ObjIndexData{Vnum: 9600, Name: "widget"}
	w.ObjIndex[9600] = idx
	room := &types.RoomIndexData{Vnum: 9601, Name: "Here"}
	w.Rooms[9601] = room

	// Make a live instance in the room.
	obj := handler.CreateObject(w, idx, 1)
	handler.ObjToRoom(obj, room)

	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoOdelete(ch, "9600")
	_ = readOutput(ch, client)
	DoOdelete(ch, "9600")
	out := readOutput(ch, client)

	if !strings.Contains(out, "deleted") {
		t.Errorf("expected deleted, got: %q", out)
	}
	if _, ok := w.ObjIndex[9600]; ok {
		t.Error("obj index should be removed")
	}
	if len(room.Contents) != 0 {
		t.Errorf("room contents should be scrubbed, got %d", len(room.Contents))
	}
}

func TestDoOdelete_BadVnum(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	DoOdelete(ch, "77777")
	out := readOutput(ch, client)
	if !strings.Contains(out, "does not exist") {
		t.Errorf("expected does-not-exist, got: %q", out)
	}
}

// --- DoMdelete ---

func TestDoMdelete_ScrubLiveInstances(t *testing.T) {
	w := setupOlcWorld()
	idx := &types.MobIndexData{Vnum: 9700, PlayerName: "rat", ShortDescr: "a rat"}
	idx.Act.Set(types.ACT_IS_NPC)
	w.MobIndex[9700] = idx
	room := &types.RoomIndexData{Vnum: 9701, Name: "Here"}
	w.Rooms[9701] = room

	mob := handler.CreateMobile(w, idx)
	handler.CharToRoom(mob, room)

	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoMdelete(ch, "9700")
	_ = readOutput(ch, client)
	DoMdelete(ch, "9700")
	out := readOutput(ch, client)

	if !strings.Contains(out, "deleted") {
		t.Errorf("expected deleted, got: %q", out)
	}
	if _, ok := w.MobIndex[9700]; ok {
		t.Error("mob index should be removed")
	}
	// The live mob should no longer be in the room
	for _, c := range room.People {
		if c.IndexData == idx {
			t.Error("live mob instance should be extracted")
		}
	}
}

func TestDoMdelete_BadVnum(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	DoMdelete(ch, "88888")
	out := readOutput(ch, client)
	if !strings.Contains(out, "does not exist") {
		t.Errorf("expected does-not-exist, got: %q", out)
	}
}

// --- Mortal guard ---

func TestOLC_DeleteMortalReject(t *testing.T) {
	_ = setupOlcWorld()
	mortal, client := makeTestChar("Mortal")
	defer client.Close()
	mortal.Level = 1

	cases := []struct {
		name string
		fn   func(ch *types.CharData, arg string)
	}{
		{"rdelete", DoRdelete},
		{"odelete", DoOdelete},
		{"mdelete", DoMdelete},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.fn(mortal, "1")
			out := readOutput(mortal, client)
			if !strings.Contains(out, "Huh?") {
				t.Errorf("expected Huh?, got: %q", out)
			}
		})
	}
}
