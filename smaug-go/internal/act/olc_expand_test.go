package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// --- redit ed / rmed ---

func TestDoRedit_EdNewKeyword(t *testing.T) {
	_ = setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 8000, Name: "Test"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "ed fountain")
	_ = readOutput(ch, client)

	if len(room.ExtraDescr) != 1 {
		t.Fatalf("expected 1 ed, got %d", len(room.ExtraDescr))
	}
	if room.ExtraDescr[0].Keyword != "fountain" {
		t.Errorf("expected keyword 'fountain', got %q", room.ExtraDescr[0].Keyword)
	}
	if ch.Substate != types.SUB_ROOM_EXTRA {
		t.Errorf("expected SUB_ROOM_EXTRA substate, got %d", ch.Substate)
	}
}

func TestDoRedit_EdNoKeyword(t *testing.T) {
	_ = setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 8001, Name: "Test"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "ed")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Usage") {
		t.Errorf("expected usage, got: %q", out)
	}
}

func TestDoRedit_Rmed(t *testing.T) {
	_ = setupOlcWorld()
	room := &types.RoomIndexData{
		Vnum: 8002, Name: "Test",
		ExtraDescr: []*types.ExtraDescrData{
			{Keyword: "foo", Description: "a foo"},
			{Keyword: "bar", Description: "a bar"},
		},
	}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "rmed foo")
	out := readOutput(ch, client)

	if len(room.ExtraDescr) != 1 {
		t.Fatalf("expected 1 ed, got %d", len(room.ExtraDescr))
	}
	if room.ExtraDescr[0].Keyword != "bar" {
		t.Errorf("wrong ed remained: %q", room.ExtraDescr[0].Keyword)
	}
	if !strings.Contains(out, "removed") {
		t.Errorf("expected confirmation, got: %q", out)
	}
}

func TestDoRedit_RmedNoMatch(t *testing.T) {
	_ = setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 8003, Name: "Test"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "rmed missing")
	out := readOutput(ch, client)
	if !strings.Contains(out, "No extra description") {
		t.Errorf("expected no match message, got: %q", out)
	}
}

// --- redit bexit ---

func TestDoRedit_BexitSuccess(t *testing.T) {
	w := setupOlcWorld()
	src := &types.RoomIndexData{Vnum: 8100, Name: "Src"}
	dst := &types.RoomIndexData{Vnum: 8101, Name: "Dst"}
	w.Rooms[8100] = src
	w.Rooms[8101] = dst
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = src

	DoRedit(ch, "bexit north 8101")
	out := readOutput(ch, client)

	if !strings.Contains(out, "8101") {
		t.Errorf("expected vnum in output, got: %q", out)
	}
	// Forward exit
	fwd := src.GetExit(types.DIR_NORTH)
	if fwd == nil || fwd.ToRoom != dst {
		t.Error("forward exit missing or wrong dest")
	}
	// Reverse exit
	back := dst.GetExit(types.DIR_SOUTH)
	if back == nil || back.ToRoom != src {
		t.Error("reverse exit missing or wrong dest")
	}
}

func TestDoRedit_BexitBadVnum(t *testing.T) {
	w := setupOlcWorld()
	src := &types.RoomIndexData{Vnum: 8110, Name: "Src"}
	w.Rooms[8110] = src
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = src

	DoRedit(ch, "bexit north 99999")
	out := readOutput(ch, client)
	if !strings.Contains(out, "does not exist") {
		t.Errorf("expected 'does not exist', got: %q", out)
	}
}

func TestDoRedit_BexitBadDir(t *testing.T) {
	_ = setupOlcWorld()
	src := &types.RoomIndexData{Vnum: 8111, Name: "Src"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = src

	DoRedit(ch, "bexit xyzzy 100")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Usage") {
		t.Errorf("expected usage, got: %q", out)
	}
}

// --- redit exflags ---

func TestDoRedit_ExflagsToggle(t *testing.T) {
	_ = setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 8200, Name: "Test"}
	room.Exits = []*types.ExitData{{Direction: types.DIR_NORTH}}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "exflags north closed")
	_ = readOutput(ch, client)

	if room.Exits[0].ExitInfo&int(types.EX_CLOSED) == 0 {
		t.Error("expected EX_CLOSED to be set")
	}
	DoRedit(ch, "exflags north closed")
	_ = readOutput(ch, client)
	if room.Exits[0].ExitInfo&int(types.EX_CLOSED) != 0 {
		t.Error("expected EX_CLOSED to be toggled off")
	}
}

func TestDoRedit_ExflagsBadName(t *testing.T) {
	_ = setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 8201, Name: "Test"}
	room.Exits = []*types.ExitData{{Direction: types.DIR_NORTH}}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "exflags north bogus")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Valid flags") {
		t.Errorf("expected flag list, got: %q", out)
	}
}

func TestDoRedit_ExflagsNoExit(t *testing.T) {
	_ = setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 8202, Name: "Test"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "exflags north closed")
	out := readOutput(ch, client)
	if !strings.Contains(out, "No exit") {
		t.Errorf("expected no-exit message, got: %q", out)
	}
}

// --- redit exname / exkey ---

func TestDoRedit_Exname(t *testing.T) {
	_ = setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 8300, Name: "Test"}
	room.Exits = []*types.ExitData{{Direction: types.DIR_NORTH}}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "exname north iron door")
	_ = readOutput(ch, client)

	if room.Exits[0].Keyword != "iron door" {
		t.Errorf("expected 'iron door', got %q", room.Exits[0].Keyword)
	}
}

func TestDoRedit_Exkey(t *testing.T) {
	_ = setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 8301, Name: "Test"}
	room.Exits = []*types.ExitData{{Direction: types.DIR_NORTH}}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "exkey north 7777")
	_ = readOutput(ch, client)

	if room.Exits[0].Key != 7777 {
		t.Errorf("expected key 7777, got %d", room.Exits[0].Key)
	}
}

func TestDoRedit_ExnameBadDir(t *testing.T) {
	_ = setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 8302, Name: "Test"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "exname xyzzy foo")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Usage") {
		t.Errorf("expected usage, got: %q", out)
	}
}

// --- redit teledelay / televnum / tunnel ---

func TestDoRedit_Teledelay(t *testing.T) {
	_ = setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 8400, Name: "Test"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "teledelay 15")
	_ = readOutput(ch, client)

	if room.TeleDelay != 15 {
		t.Errorf("expected TeleDelay 15, got %d", room.TeleDelay)
	}
}

func TestDoRedit_Televnum(t *testing.T) {
	_ = setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 8401, Name: "Test"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "televnum 3001")
	_ = readOutput(ch, client)

	if room.TeleVnum != 3001 {
		t.Errorf("expected TeleVnum 3001, got %d", room.TeleVnum)
	}
}

func TestDoRedit_Tunnel(t *testing.T) {
	_ = setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 8402, Name: "Test"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "tunnel 5")
	_ = readOutput(ch, client)

	if room.Tunnel != 5 {
		t.Errorf("expected Tunnel 5, got %d", room.Tunnel)
	}
}

func TestDoRedit_TeledelayNegative(t *testing.T) {
	_ = setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 8403, Name: "Test"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "teledelay -1")
	out := readOutput(ch, client)
	if !strings.Contains(out, "non-negative") {
		t.Errorf("expected rejection, got: %q", out)
	}
}

// --- redit rlist delegation ---

func TestDoRedit_Rlist(t *testing.T) {
	w := setupOlcWorld()
	w.Rooms[8500] = &types.RoomIndexData{Vnum: 8500, Name: "Alpha"}
	w.Rooms[8505] = &types.RoomIndexData{Vnum: 8505, Name: "Beta"}
	room := &types.RoomIndexData{Vnum: 8499, Name: "Here"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = room

	DoRedit(ch, "rlist 8500 8510")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Alpha") {
		t.Errorf("expected Alpha in rlist, got: %q", out)
	}
}
