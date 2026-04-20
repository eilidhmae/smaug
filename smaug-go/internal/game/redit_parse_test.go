package game

import (
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// reditTestRig wraps a descriptor connected to a pipe; reads drain into
// a mutex-protected buffer so readBuf() can inspect accumulated output.
type reditTestRig struct {
	d    *types.DescriptorData
	room *types.RoomIndexData
	mu   sync.Mutex
	sink []byte
	done chan struct{}
}

// newReditHarness spins up an in-process descriptor + reader so tests
// can drive reditParse and observe menu emissions via readBuf. Returns
// the rig (use rig.d / rig.room / rig.readBuf).
func newReditHarness(t *testing.T) *reditTestRig {
	t.Helper()
	server, client := net.Pipe()
	rig := &reditTestRig{done: make(chan struct{})}
	t.Cleanup(func() {
		server.Close()
		client.Close()
		<-rig.done
	})

	// Drain client side into rig.sink.
	go func() {
		defer close(rig.done)
		buf := make([]byte, 4096)
		for {
			n, err := client.Read(buf)
			if n > 0 {
				rig.mu.Lock()
				rig.sink = append(rig.sink, buf[:n]...)
				rig.mu.Unlock()
			}
			if err == io.EOF || err != nil {
				return
			}
		}
	}()

	d := types.NewDescriptor(server)
	ch := &types.CharData{Name: "Builder", Level: 100, Trust: 100}
	ch.Desc = d
	d.Character = ch
	room := &types.RoomIndexData{
		Vnum:        1000,
		Name:        "The Workshop",
		Description: "A plain workshop.\n\r",
		SectorType:  types.SECT_INSIDE,
	}
	d.Olc = &types.OlcData{
		Mode:   types.REDIT_MAIN_MENU,
		Vnum:   room.Vnum,
		Target: room,
	}
	d.Connected = int(types.CON_REDIT)
	rig.d = d
	rig.room = room
	return rig
}

// readBuf flushes the descriptor's output buffer to the pipe, waits
// briefly for the reader goroutine to accumulate bytes, then returns
// and clears the accumulated contents.
func (r *reditTestRig) readBuf(t *testing.T) string {
	t.Helper()
	if err := r.d.FlushOutput(); err != nil {
		t.Fatalf("FlushOutput: %v", err)
	}
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		r.mu.Lock()
		got := len(r.sink) > 0
		r.mu.Unlock()
		if got {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	r.mu.Lock()
	s := string(r.sink)
	r.sink = nil
	r.mu.Unlock()
	return s
}

// strconvItoa: wrapper so test body reads cleanly.
func strconvItoa(n int) string { return strconv.Itoa(n) }

// captureLog diverts the stdlib logger output while fn runs, returning
// whatever was logged. Used to pin olcLog's format string; restores the
// original writer on return.
func captureLog(fn func()) string {
	var buf strings.Builder
	prior := log.Writer()
	priorFlags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	defer func() {
		log.SetOutput(prior)
		log.SetFlags(priorFlags)
	}()
	fn()
	return buf.String()
}

// --- olcLog target-label tests (plan-phase6-olc-oedit.md §G2) ---

// TestOlcLog_EmitsRoomPrefix pins the redit caller path: target="ROOM"
// produces "ROOM(vnum)" in the log line. Mutation gate: swapping the
// format-string argument order in olcLog flips this to "123(ROOM)".
func TestOlcLog_EmitsRoomPrefix(t *testing.T) {
	rig := newReditHarness(t)
	out := captureLog(func() {
		olcLog(rig.d, "ROOM", "Changed name to %s", "Foo")
	})
	if !strings.Contains(out, "ROOM(1000)") {
		t.Errorf("olcLog output missing ROOM(1000); got %q", out)
	}
	if !strings.Contains(out, "Changed name to Foo") {
		t.Errorf("olcLog output missing formatted body; got %q", out)
	}
}

// TestOlcLog_CustomTarget pins the generalization: target="OBJ" produces
// "OBJ(vnum)". Confirms the target-label parameter is threaded all the way
// through to the formatted log line.
func TestOlcLog_CustomTarget(t *testing.T) {
	rig := newReditHarness(t)
	rig.d.Olc.Vnum = 1234 // OBJ vnum
	out := captureLog(func() {
		olcLog(rig.d, "OBJ", "Changed type to %s", "weapon")
	})
	if !strings.Contains(out, "OBJ(1234)") {
		t.Errorf("olcLog output missing OBJ(1234); got %q", out)
	}
	if !strings.Contains(out, "Changed type to weapon") {
		t.Errorf("olcLog output missing formatted body; got %q", out)
	}
}

// ----------------------------------------------------------------------

// TestReditDispMenu_ContainsAllFields pins A5 main-menu output. Plan §G4.
func TestReditDispMenu_ContainsAllFields(t *testing.T) {
	rig := newReditHarness(t)
	rig.room.RoomFlags.Set(types.ROOM_DARK)
	rig.room.Tunnel = 5
	rig.room.TeleDelay = 10
	rig.room.TeleVnum = 2000

	ReditDispMenu(rig.d)

	out := rig.readBuf(t)
	for _, want := range []string{
		"Room number : [", "1000", "Workshop",
		"1", "Name", "2", "Description", "3", "Room flags",
		"4", "Sector type", "Inside", "5", "Tunnel", "6", "TeleDelay",
		"7", "TeleVnum", "A", "Exit menu", "B", "Extra descriptions",
		"Q", "Quit",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("menu missing %q; got:\n%s", want, out)
		}
	}
	if rig.d.Olc.Mode != types.REDIT_MAIN_MENU {
		t.Errorf("Olc.Mode = %d, want REDIT_MAIN_MENU", rig.d.Olc.Mode)
	}
}

// TestReditParse_Quit_CleansUpAndReturnsToPlaying pins A14.
func TestReditParse_Quit_CleansUpAndReturnsToPlaying(t *testing.T) {
	rig := newReditHarness(t)

	reditParse(rig.d, "Q")

	if rig.d.Connected != int(types.CON_PLAYING) {
		t.Errorf("Connected = %d, want CON_PLAYING", rig.d.Connected)
	}
	if rig.d.Olc != nil {
		t.Errorf("Olc should be nil after Q; got %+v", rig.d.Olc)
	}
	if rig.d.Character.Substate != types.SUB_NONE {
		t.Errorf("Substate = %d, want SUB_NONE", rig.d.Character.Substate)
	}
}

// TestReditParse_MainMenu_InvalidChoice pins A15.
func TestReditParse_MainMenu_InvalidChoice(t *testing.T) {
	rig := newReditHarness(t)

	reditParse(rig.d, "z")

	if rig.d.Connected != int(types.CON_REDIT) {
		t.Errorf("Connected = %d, want CON_REDIT (still)", rig.d.Connected)
	}
	out := rig.readBuf(t)
	if !strings.Contains(out, "Invalid choice") {
		t.Errorf("expected 'Invalid choice'; got %q", out)
	}
	if !strings.Contains(out, "Enter choice") {
		t.Errorf("expected menu redisplay; got %q", out)
	}
}

// TestReditParse_Name_SetsRoomName pins A7.
func TestReditParse_Name_SetsRoomName(t *testing.T) {
	rig := newReditHarness(t)

	reditParse(rig.d, "1")
	if rig.d.Olc.Mode != types.REDIT_NAME {
		t.Fatalf("Mode = %d, want REDIT_NAME", rig.d.Olc.Mode)
	}
	reditParse(rig.d, "A Fancy Workshop")
	if rig.room.Name != "A Fancy Workshop" {
		t.Errorf("room.Name = %q, want 'A Fancy Workshop'", rig.room.Name)
	}
	if rig.d.Olc.Mode != types.REDIT_MAIN_MENU {
		t.Errorf("Mode = %d, want REDIT_MAIN_MENU after set", rig.d.Olc.Mode)
	}
}

// TestReditParse_Name_TildeSmashed.
func TestReditParse_Name_TildeSmashed(t *testing.T) {
	rig := newReditHarness(t)
	reditParse(rig.d, "1")
	reditParse(rig.d, "A~B")
	if strings.ContainsRune(rig.room.Name, '~') {
		t.Errorf("name contains tilde: %q", rig.room.Name)
	}
}

// TestReditParse_Tunnel_Clamps pins A11.
func TestReditParse_Tunnel_Clamps(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"5", 5},
		{"-5", 0},
		{"9999", 1000},
		{"0", 0},
	}
	for _, tc := range cases {
		rig := newReditHarness(t)
		reditParse(rig.d, "5")
		reditParse(rig.d, tc.in)
		if rig.room.Tunnel != tc.want {
			t.Errorf("tunnel %s → %d, want %d", tc.in, rig.room.Tunnel, tc.want)
		}
	}
}

// TestReditParse_Teledelay.
func TestReditParse_Teledelay(t *testing.T) {
	rig := newReditHarness(t)
	reditParse(rig.d, "6")
	reditParse(rig.d, "42")
	if rig.room.TeleDelay != 42 {
		t.Errorf("TeleDelay = %d, want 42", rig.room.TeleDelay)
	}
}

// TestReditParse_Televnum_Clamps.
func TestReditParse_Televnum_Clamps(t *testing.T) {
	rig := newReditHarness(t)
	reditParse(rig.d, "7")
	reditParse(rig.d, "0")
	if rig.room.TeleVnum != 1 {
		t.Errorf("TeleVnum clamp low = %d, want 1", rig.room.TeleVnum)
	}
}

// TestReditParse_Flags_NumberToggle pins A9.
func TestReditParse_Flags_NumberToggle(t *testing.T) {
	rig := newReditHarness(t)
	reditParse(rig.d, "3")
	reditParse(rig.d, "1")
	if !rig.room.RoomFlags.IsSet(types.ROOM_DARK) {
		t.Error("expected ROOM_DARK set after first toggle")
	}
	reditParse(rig.d, "1")
	if rig.room.RoomFlags.IsSet(types.ROOM_DARK) {
		t.Error("expected ROOM_DARK cleared after second toggle")
	}
}

// TestReditParse_Flags_WordToggle pins A9.
func TestReditParse_Flags_WordToggle(t *testing.T) {
	rig := newReditHarness(t)
	reditParse(rig.d, "3")
	reditParse(rig.d, "dark silence")
	if !rig.room.RoomFlags.IsSet(types.ROOM_DARK) {
		t.Error("expected ROOM_DARK set")
	}
	if !rig.room.RoomFlags.IsSet(types.ROOM_SILENCE) {
		t.Error("expected ROOM_SILENCE set")
	}
}

// TestReditParse_Flags_ZeroExits.
func TestReditParse_Flags_ZeroExits(t *testing.T) {
	rig := newReditHarness(t)
	reditParse(rig.d, "3")
	if rig.d.Olc.Mode != types.REDIT_FLAGS {
		t.Fatalf("Mode = %d, want REDIT_FLAGS", rig.d.Olc.Mode)
	}
	reditParse(rig.d, "0")
	if rig.d.Olc.Mode != types.REDIT_MAIN_MENU {
		t.Errorf("Mode = %d, want REDIT_MAIN_MENU after 0", rig.d.Olc.Mode)
	}
}

// TestReditParse_Sector_Valid pins A10.
func TestReditParse_Sector_Valid(t *testing.T) {
	rig := newReditHarness(t)
	reditParse(rig.d, "4")
	reditParse(rig.d, strconvItoa(types.SECT_DESERT))
	if rig.room.SectorType != types.SECT_DESERT {
		t.Errorf("SectorType = %d, want SECT_DESERT", rig.room.SectorType)
	}
}

// TestReditParse_Sector_RejectsDunno.
func TestReditParse_Sector_RejectsDunno(t *testing.T) {
	rig := newReditHarness(t)
	rig.room.SectorType = types.SECT_CITY
	reditParse(rig.d, "4")
	reditParse(rig.d, strconvItoa(types.SECT_DUNNO))
	if rig.room.SectorType != types.SECT_CITY {
		t.Errorf("SectorType = %d, want unchanged (SECT_CITY)", rig.room.SectorType)
	}
	if rig.d.Olc.Mode != types.REDIT_SECTOR {
		t.Errorf("Mode = %d, want still REDIT_SECTOR", rig.d.Olc.Mode)
	}
}

// TestReditParse_ExitFlow_AddAndEdit pins A12.
func TestReditParse_ExitFlow_AddAndEdit(t *testing.T) {
	rig := newReditHarness(t)
	other := &types.RoomIndexData{Vnum: 2001, Name: "Other"}
	w := &world.World{Rooms: map[int]*types.RoomIndexData{1000: rig.room, 2001: other}}
	SetWorldRef(w)
	t.Cleanup(func() { SetWorldRef(nil) })

	reditParse(rig.d, "A")
	reditParse(rig.d, "A")
	reditParse(rig.d, "0")
	reditParse(rig.d, "2001")

	if len(rig.room.Exits) != 1 {
		t.Fatalf("exits count = %d, want 1", len(rig.room.Exits))
	}
	ex := rig.room.Exits[0]
	if ex.Direction != types.DIR_NORTH {
		t.Errorf("exit direction = %d, want DIR_NORTH", ex.Direction)
	}
	if ex.ToRoom != other {
		t.Error("exit.ToRoom != other room")
	}
	if ex.Vnum != 2001 {
		t.Errorf("exit.Vnum = %d, want 2001", ex.Vnum)
	}
}

// TestReditParse_ExitVnum_SetsBothFields — C-bug fix (Open Q4):
// setting vnum also updates ToRoom so the exit doesn't point at the
// stale room.
func TestReditParse_ExitVnum_SetsBothFields(t *testing.T) {
	rig := newReditHarness(t)
	roomA := &types.RoomIndexData{Vnum: 3000, Name: "A"}
	roomB := &types.RoomIndexData{Vnum: 3001, Name: "B"}
	w := &world.World{Rooms: map[int]*types.RoomIndexData{1000: rig.room, 3000: roomA, 3001: roomB}}
	SetWorldRef(w)
	t.Cleanup(func() { SetWorldRef(nil) })

	pexit := &types.ExitData{Direction: types.DIR_EAST, ToRoom: roomA, Vnum: 3000, RVnum: 3000}
	rig.room.Exits = append(rig.room.Exits, pexit)

	reditParse(rig.d, "A")
	reditParse(rig.d, "1")
	reditParse(rig.d, "2")
	reditParse(rig.d, "3001")

	if pexit.Vnum != 3001 {
		t.Errorf("exit.Vnum = %d, want 3001", pexit.Vnum)
	}
	if pexit.ToRoom != roomB {
		t.Errorf("exit.ToRoom = %v, want roomB (bug fix)", pexit.ToRoom)
	}
}

// TestReditParse_ExitDelete.
func TestReditParse_ExitDelete(t *testing.T) {
	rig := newReditHarness(t)
	rig.room.Exits = append(rig.room.Exits,
		&types.ExitData{Direction: types.DIR_NORTH, Vnum: 100},
		&types.ExitData{Direction: types.DIR_EAST, Vnum: 101},
	)

	reditParse(rig.d, "A")
	reditParse(rig.d, "R")
	reditParse(rig.d, "1")

	if len(rig.room.Exits) != 1 {
		t.Fatalf("exits count = %d, want 1", len(rig.room.Exits))
	}
	if rig.room.Exits[0].Direction != types.DIR_EAST {
		t.Errorf("remaining exit direction = %d, want DIR_EAST", rig.room.Exits[0].Direction)
	}
}

// TestReditParse_ExitFlags_Toggle.
func TestReditParse_ExitFlags_Toggle(t *testing.T) {
	rig := newReditHarness(t)
	rig.room.Exits = append(rig.room.Exits, &types.ExitData{Direction: types.DIR_NORTH})

	reditParse(rig.d, "A")
	reditParse(rig.d, "1")
	reditParse(rig.d, "5")
	reditParse(rig.d, "1")

	if rig.room.Exits[0].ExitInfo&1 == 0 {
		t.Errorf("expected isdoor bit set; ExitInfo = %d", rig.room.Exits[0].ExitInfo)
	}
}

// TestReditParse_Extradesc_AddKeyQuitJunksIfEmpty pins A13.
func TestReditParse_Extradesc_AddKeyQuitJunksIfEmpty(t *testing.T) {
	rig := newReditHarness(t)

	reditParse(rig.d, "B")
	reditParse(rig.d, "A")
	if len(rig.room.ExtraDescr) != 1 {
		t.Fatalf("after add: count = %d, want 1", len(rig.room.ExtraDescr))
	}
	reditParse(rig.d, "Q")
	if len(rig.room.ExtraDescr) != 0 {
		t.Errorf("after Q on empty ed: count = %d, want 0", len(rig.room.ExtraDescr))
	}
}

// TestReditParse_Extradesc_KeywordThenQuitKeeps — keyword+description set means Q keeps.
func TestReditParse_Extradesc_KeywordThenQuitKeeps(t *testing.T) {
	rig := newReditHarness(t)
	reditParse(rig.d, "B")
	reditParse(rig.d, "A")
	reditParse(rig.d, "1")
	reditParse(rig.d, "foo bar")
	rig.room.ExtraDescr[0].Description = "something"
	reditParse(rig.d, "Q")

	if len(rig.room.ExtraDescr) != 1 {
		t.Errorf("after keyword-set + Q: count = %d, want 1", len(rig.room.ExtraDescr))
	}
	if rig.room.ExtraDescr[0].Keyword != "foo bar" {
		t.Errorf("keyword = %q, want 'foo bar'", rig.room.ExtraDescr[0].Keyword)
	}
}

// TestCleanupOlc_ResetsDescriptorAndCharacter pins G10.
func TestCleanupOlc_ResetsDescriptorAndCharacter(t *testing.T) {
	rig := newReditHarness(t)
	rig.d.Character.Substate = types.SUB_ROOM_DESC

	cleanupOlc(rig.d)

	if rig.d.Olc != nil {
		t.Error("Olc should be nil")
	}
	if rig.d.Connected != int(types.CON_PLAYING) {
		t.Errorf("Connected = %d, want CON_PLAYING", rig.d.Connected)
	}
	if rig.d.Character.Substate != types.SUB_NONE {
		t.Errorf("Substate = %d, want SUB_NONE", rig.d.Character.Substate)
	}
}

// TestCleanupOlc_Idempotent.
func TestCleanupOlc_Idempotent(t *testing.T) {
	rig := newReditHarness(t)
	cleanupOlc(rig.d)
	cleanupOlc(rig.d)
	if rig.d.Olc != nil {
		t.Error("Olc should still be nil")
	}
}

// TestReditParse_NilDescriptor.
func TestReditParse_NilDescriptor(t *testing.T) {
	reditParse(nil, "anything")
}

// TestReditParse_NilOlc_FallsBackToPlaying.
func TestReditParse_NilOlc_FallsBackToPlaying(t *testing.T) {
	rig := newReditHarness(t)
	rig.d.Olc = nil
	reditParse(rig.d, "anything")
	if rig.d.Connected != int(types.CON_PLAYING) {
		t.Errorf("Connected = %d, want CON_PLAYING", rig.d.Connected)
	}
}

// TestReditParse_ExitAdd_RejectsDuplicate — word-form direction that
// already exists returns to exit menu without creating.
func TestReditParse_ExitAdd_RejectsDuplicate(t *testing.T) {
	rig := newReditHarness(t)
	rig.room.Exits = append(rig.room.Exits, &types.ExitData{Direction: types.DIR_NORTH, Vnum: 999})

	reditParse(rig.d, "A")
	reditParse(rig.d, "A")
	reditParse(rig.d, "north")

	if len(rig.room.Exits) != 1 {
		t.Errorf("exits count = %d, want 1 (no new exit added)", len(rig.room.Exits))
	}
}

// TestReditParse_ExitMenu_NumericEditsExit.
func TestReditParse_ExitMenu_NumericEditsExit(t *testing.T) {
	rig := newReditHarness(t)
	ex := &types.ExitData{Direction: types.DIR_NORTH, Vnum: 500}
	rig.room.Exits = append(rig.room.Exits, ex)

	reditParse(rig.d, "A")
	reditParse(rig.d, "1")

	if rig.d.Olc.Mode != types.REDIT_EXIT_EDIT {
		t.Errorf("Mode = %d, want REDIT_EXIT_EDIT", rig.d.Olc.Mode)
	}
	if rig.d.Olc.Spare != ex {
		t.Error("Spare should point to the selected exit")
	}
}

// TestReditParse_ExitKey — "3" + vnum sets key.
func TestReditParse_ExitKey(t *testing.T) {
	rig := newReditHarness(t)
	ex := &types.ExitData{Direction: types.DIR_NORTH, Vnum: 500}
	rig.room.Exits = append(rig.room.Exits, ex)

	reditParse(rig.d, "A")
	reditParse(rig.d, "1")
	reditParse(rig.d, "3")
	reditParse(rig.d, "42")

	if ex.Key != 42 {
		t.Errorf("exit.Key = %d, want 42", ex.Key)
	}
}

// TestReditParse_ExitKeyword — "4" + text sets keyword.
func TestReditParse_ExitKeyword(t *testing.T) {
	rig := newReditHarness(t)
	ex := &types.ExitData{Direction: types.DIR_NORTH, Vnum: 500}
	rig.room.Exits = append(rig.room.Exits, ex)

	reditParse(rig.d, "A")
	reditParse(rig.d, "1")
	reditParse(rig.d, "4")
	reditParse(rig.d, "wooden door")

	if ex.Keyword != "wooden door" {
		t.Errorf("keyword = %q, want 'wooden door'", ex.Keyword)
	}
}

// TestReditParse_ExitDesc — "6" + text sets description with \n\r.
func TestReditParse_ExitDesc(t *testing.T) {
	rig := newReditHarness(t)
	ex := &types.ExitData{Direction: types.DIR_NORTH, Vnum: 500}
	rig.room.Exits = append(rig.room.Exits, ex)

	reditParse(rig.d, "A")
	reditParse(rig.d, "1")
	reditParse(rig.d, "6")
	reditParse(rig.d, "A sturdy door")

	if !strings.HasPrefix(ex.Description, "A sturdy door") {
		t.Errorf("description = %q, want prefix 'A sturdy door'", ex.Description)
	}
}

// TestReditParse_ExitEdit_QReturnsToExitMenu.
func TestReditParse_ExitEdit_QReturnsToExitMenu(t *testing.T) {
	rig := newReditHarness(t)
	rig.room.Exits = append(rig.room.Exits, &types.ExitData{Direction: types.DIR_NORTH, Vnum: 500})

	reditParse(rig.d, "A")
	reditParse(rig.d, "1")
	if rig.d.Olc.Mode != types.REDIT_EXIT_EDIT {
		t.Fatalf("Mode = %d, want REDIT_EXIT_EDIT", rig.d.Olc.Mode)
	}
	reditParse(rig.d, "Q")
	if rig.d.Olc.Mode != types.REDIT_EXIT_MENU {
		t.Errorf("Mode = %d, want REDIT_EXIT_MENU after Q", rig.d.Olc.Mode)
	}
	if rig.d.Olc.Spare != nil {
		t.Error("Spare should be nil after Q")
	}
}

// TestReditParse_DescEditorCallback_ReturnsToConRedit pins plan §A8.
// When the builder picks "2" at the main menu, edits the description,
// and saves via /s, the callback MUST restore CON_REDIT — NOT leave
// the descriptor in CON_PLAYING. The callback here is synthesized by
// the REDIT_MAIN_MENU branch in reditHandleMainMenu; this test drives
// that branch then manually invokes the callback to simulate /s
// (without running the full line-editor FSM).
func TestReditParse_DescEditorCallback_ReturnsToConRedit(t *testing.T) {
	rig := newReditHarness(t)

	// Enter desc editor. reditHandleMainMenu sets up the editor
	// session via StartEditing; the callback is stashed on EditorSave.
	reditParse(rig.d, "2")

	// The character should now be in CON_EDITING with EditorSave set.
	if rig.d.Connected != int(types.CON_EDITING) {
		t.Fatalf("after '2': Connected = %d, want CON_EDITING", rig.d.Connected)
	}
	if rig.d.Character.EditorSave == nil {
		t.Fatal("EditorSave callback was not installed")
	}

	// Seed an Editor buffer so CopyBuffer returns something non-empty.
	rig.d.Character.Editor = &types.EditorData{NumLines: 1}
	rig.d.Character.Editor.Lines[0] = "A new description line."

	// Simulate the /s path: game.EditBuffer's '/s' branch first
	// transitions CON_EDITING→CON_PLAYING, then invokes the callback.
	// We reproduce that contract here directly.
	rig.d.Connected = int(types.CON_PLAYING)
	save := rig.d.Character.EditorSave
	rig.d.Character.EditorSave = nil
	save(rig.d.Character)

	// Verdict: callback re-set Connected to CON_REDIT.
	if rig.d.Connected != int(types.CON_REDIT) {
		t.Errorf("after /s: Connected = %d, want CON_REDIT (callback must re-set)",
			rig.d.Connected)
	}
	if !strings.Contains(rig.room.Description, "A new description") {
		t.Errorf("room.Description = %q, want saved buffer", rig.room.Description)
	}
}

// TestReditParse_ExitMenu_QReturnsToMain.
func TestReditParse_ExitMenu_QReturnsToMain(t *testing.T) {
	rig := newReditHarness(t)
	reditParse(rig.d, "A")
	if rig.d.Olc.Mode != types.REDIT_EXIT_MENU {
		t.Fatalf("Mode = %d, want REDIT_EXIT_MENU", rig.d.Olc.Mode)
	}
	reditParse(rig.d, "Q")
	if rig.d.Olc.Mode != types.REDIT_MAIN_MENU {
		t.Errorf("Mode = %d, want REDIT_MAIN_MENU after Q", rig.d.Olc.Mode)
	}
}

// TestRoomFlagsString — renders set flags space-separated; empty when no flags.
func TestRoomFlagsString(t *testing.T) {
	var bv types.BitVector
	if s := roomFlagsString(bv); s != "" {
		t.Errorf("empty bv = %q, want empty", s)
	}
	bv.Set(types.ROOM_DARK)
	bv.Set(types.ROOM_INDOORS)
	s := roomFlagsString(bv)
	if !strings.Contains(s, "dark") || !strings.Contains(s, "indoors") {
		t.Errorf("got %q, want both 'dark' and 'indoors'", s)
	}
}

// TestGetRoomFlagBit — known lookup, unknown returns -1.
func TestGetRoomFlagBit(t *testing.T) {
	if getRoomFlagBit("dark") != int(types.ROOM_DARK) {
		t.Errorf("dark bit = %d, want %d", getRoomFlagBit("dark"), types.ROOM_DARK)
	}
	if getRoomFlagBit("NOMOB") != int(types.ROOM_NO_MOB) {
		t.Errorf("case-insensitive fail: NOMOB → %d", getRoomFlagBit("NOMOB"))
	}
	if getRoomFlagBit("bogus") != -1 {
		t.Error("bogus flag should return -1")
	}
	if getRoomFlagBit("") != -1 {
		t.Error("empty should return -1")
	}
}

// TestSectorName — known sectors + ??? fallback.
func TestSectorName(t *testing.T) {
	if sectorName(types.SECT_INSIDE) != "Inside" {
		t.Errorf("SECT_INSIDE = %q, want 'Inside'", sectorName(types.SECT_INSIDE))
	}
	if sectorName(9999) != "???!" {
		t.Errorf("unknown sector = %q, want '???!'", sectorName(9999))
	}
}
