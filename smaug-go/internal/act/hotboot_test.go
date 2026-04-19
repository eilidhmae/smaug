//go:build !windows

package act

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
)

// setupHotbootWorld returns a world with a DataDir under t.TempDir() so
// SaveWorld / SaveHotbootSessions land in a clean directory per test.
// Also seeds a room + ensures WorldRef is wired.
func setupHotbootWorld(t *testing.T) (*types.RoomIndexData, func()) {
	t.Helper()
	tmp := t.TempDir()
	w := setupCommWorld()
	w.DataDir = tmp
	w.SysData.HotbootInProgress = false
	w.Descriptors = nil
	w.Characters = nil
	room := &types.RoomIndexData{Vnum: 3001, Name: "Test Room"}
	w.Rooms[3001] = room

	// Save + restore package-level func seams so each test sees a clean
	// state. execSelf is restored regardless of whether the test touches it.
	prevExec := execSelf
	prevPause := HotbootPauseFunc
	prevResume := HotbootResumeFunc
	prevPort := HotbootPort
	prevSave := SaveFunc
	cleanup := func() {
		execSelf = prevExec
		HotbootPauseFunc = prevPause
		HotbootResumeFunc = prevResume
		HotbootPort = prevPort
		SaveFunc = prevSave
	}
	return room, cleanup
}

// makeImmortalHotbootChar creates a CharData with Trust=LEVEL_ASCENDANT
// so the trust gate passes, plus a PCData so Hotboot bool has somewhere
// to live.
func makeImmortalHotbootChar(name string, room *types.RoomIndexData) (*types.CharData, net.Conn) {
	ch, client := makeTestChar(name)
	ch.Trust = types.LEVEL_ASCENDANT
	ch.Level = types.LEVEL_ASCENDANT
	ch.PCData = &types.PCData{Title: "the implementor"}
	handler.CharToRoom(ch, room)
	WorldRef.AddChar(ch)
	addPlayingDescriptor(ch)
	return ch, client
}

// startDrain spawns a goroutine that consumes everything the server writes
// to `client` (the test-owned pipe end) so DoHotboot's farewell FlushOutput
// does not block. Returns a stop func the test should defer-call to end
// the drain goroutine; the stop func also returns whatever was read so
// far (for tests that want to inspect the farewell).
func startDrain(client net.Conn) (stop func() string) {
	done := make(chan struct{})
	var got []byte
	doneRead := make(chan struct{})
	go func() {
		defer close(doneRead)
		buf := make([]byte, 1024)
		for {
			client.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			n, err := client.Read(buf)
			if n > 0 {
				got = append(got, buf[:n]...)
			}
			select {
			case <-done:
				return
			default:
			}
			if err != nil && n == 0 {
				// Non-timeout error: bail.
				if !strings.Contains(err.Error(), "deadline") {
					return
				}
			}
		}
	}()
	return func() string {
		close(done)
		<-doneRead
		return string(got)
	}
}

// stubPauseSuccess installs a HotbootPauseFunc that returns a fake
// listener file (pipe read-end) and one fake desc file per CON_PLAYING
// descriptor, along with session rows. The returned channel fires once
// with the sessions actually handed back, so tests can assert.
func stubPauseSuccess(t *testing.T) <-chan []persist.HotbootSession {
	t.Helper()
	out := make(chan []persist.HotbootSession, 1)
	HotbootPauseFunc = func(descriptors []*types.DescriptorData, port int) (*os.File, []*os.File, []persist.HotbootSession, error) {
		// Fake listener FD — an os.Pipe read-end we own.
		lnR, _, _ := os.Pipe()
		var descFiles []*os.File
		var sessions []persist.HotbootSession
		for i, d := range descriptors {
			if d == nil || d.Character == nil || d.Connected != types.CON_PLAYING {
				continue
			}
			r, _, _ := os.Pipe()
			descFiles = append(descFiles, r)
			roomVnum := 0
			if d.Character.InRoom != nil {
				roomVnum = d.Character.InRoom.Vnum
			}
			sessions = append(sessions, persist.HotbootSession{
				FDIndex:  i,
				RoomVnum: roomVnum,
				Port:     port,
				Name:     d.Character.Name,
				Host:     d.Host,
			})
		}
		out <- sessions
		return lnR, descFiles, sessions, nil
	}
	return out
}

// ---- Test 1: BelowAscendantRejected ---------------------------------

func TestDoHotboot_BelowAscendantRejected(t *testing.T) {
	room, cleanup := setupHotbootWorld(t)
	defer cleanup()
	ch, client := makeImmortalHotbootChar("Imp", room)
	defer client.Close()

	// Demote below threshold.
	ch.Trust = types.LEVEL_GREATER
	ch.Level = types.LEVEL_GREATER

	called := false
	execSelf = func(_ string, _ []string, _ []string) error { called = true; return nil }

	DoHotboot(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(strings.ToLower(out), "not sufficiently trusted") {
		t.Errorf("expected trust-rejection message; got %q", out)
	}
	if called {
		t.Error("execSelf should not be invoked when trust gate rejects")
	}
	// And hotboot.dat must not exist.
	path := filepath.Join(WorldRef.DataDir, "hotboot", "hotboot.dat")
	if _, err := os.Stat(path); err == nil {
		t.Error("hotboot.dat should not exist after trust rejection")
	}
}

// ---- Test 2: AnyoneFightingRejected ---------------------------------

func TestDoHotboot_AnyoneFightingRejected(t *testing.T) {
	room, cleanup := setupHotbootWorld(t)
	defer cleanup()
	ch, client := makeImmortalHotbootChar("Imp", room)
	defer client.Close()

	// Seed a SECOND PC who is fighting. DoHotboot must scan all
	// descriptors, not just ch.
	other, otherClient := makeTestChar("Brawler")
	defer otherClient.Close()
	handler.CharToRoom(other, room)
	WorldRef.AddChar(other)
	addPlayingDescriptor(other)
	other.Fighting = &types.FightData{}

	called := false
	execSelf = func(_ string, _ []string, _ []string) error { called = true; return nil }

	DoHotboot(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "in combat") {
		t.Errorf("expected combat-reject message; got %q", out)
	}
	if called {
		t.Error("execSelf should not be invoked when combat gate rejects")
	}
}

// TestDoHotboot_NobodyFighting_GateSkipped is the mutation-catch twin
// of the test above. Together they pin the `!= nil` direction of the
// fighting gate: if the gate is flipped to `== nil`, THIS test (where
// no one is fighting) fails because Imp.Fighting==nil would trigger
// the gate and reject with "in combat", whereas a correctly-oriented
// gate lets the code proceed to the happy path.
func TestDoHotboot_NobodyFighting_GateSkipped(t *testing.T) {
	room, cleanup := setupHotbootWorld(t)
	defer cleanup()
	ch, client := makeImmortalHotbootChar("Imp", room)
	defer client.Close()
	stopDrain := startDrain(client)
	defer stopDrain()

	_ = stubPauseSuccess(t)
	HotbootResumeFunc = func(*os.File) error { return nil }
	HotbootPort = 4000
	SaveFunc = func(_ *types.CharData) {}

	execed := false
	execSelf = func(_ string, _ []string, _ []string) error {
		execed = true
		return nil
	}

	DoHotboot(ch, "")

	if !execed {
		t.Error("execSelf should be called when nobody is fighting (gate must NOT reject)")
	}
}

// ---- Test 3: AnyoneInEditorRejected ---------------------------------

func TestDoHotboot_AnyoneInEditorRejected(t *testing.T) {
	room, cleanup := setupHotbootWorld(t)
	defer cleanup()
	ch, client := makeImmortalHotbootChar("Imp", room)
	defer client.Close()

	// Seed a descriptor in CON_EDITING.
	other, otherClient := makeTestChar("Scribe")
	defer otherClient.Close()
	other.Desc.Connected = types.CON_EDITING
	handler.CharToRoom(other, room)
	WorldRef.AddChar(other)
	WorldRef.Descriptors = append(WorldRef.Descriptors, other.Desc)

	called := false
	execSelf = func(_ string, _ []string, _ []string) error { called = true; return nil }

	DoHotboot(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(strings.ToLower(out), "editor") {
		t.Errorf("expected editor-reject message; got %q", out)
	}
	if called {
		t.Error("execSelf should not run when editor gate rejects")
	}
}

// ---- Test 3b: OLC substates rejected (post-adversary, 2026-04-19) ----
// Plan-phase6-hotboot §Chosen Design gate 3 (revised): CON_REDIT /
// CON_OEDIT / CON_MEDIT must refuse hotboot because OlcData + the
// EditorSave closure are process-local — syscall.Exec would drop the
// session silently. Before the fix, an OLC-state descriptor passed the
// nanny-drop filter (Character != nil) but was then excluded from
// PauseForHotboot (Connected != CON_PLAYING), leaving its FD dangling.

func TestDoHotboot_RejectsAnyOLCSubstate(t *testing.T) {
	olcStates := []struct {
		name  string
		state int
	}{
		{"CON_REDIT", types.CON_REDIT},
		{"CON_OEDIT", types.CON_OEDIT},
		{"CON_MEDIT", types.CON_MEDIT},
	}
	for _, tc := range olcStates {
		t.Run(tc.name, func(t *testing.T) {
			room, cleanup := setupHotbootWorld(t)
			defer cleanup()
			ch, client := makeImmortalHotbootChar("Imp", room)
			defer client.Close()

			other, otherClient := makeTestChar("Builder")
			defer otherClient.Close()
			other.Desc.Connected = tc.state
			handler.CharToRoom(other, room)
			WorldRef.AddChar(other)
			WorldRef.Descriptors = append(WorldRef.Descriptors, other.Desc)

			called := false
			execSelf = func(_ string, _ []string, _ []string) error { called = true; return nil }

			DoHotboot(ch, "")
			out := readOutput(ch, client)
			if !strings.Contains(strings.ToLower(out), "olc") {
				t.Errorf("expected OLC-reject message; got %q", out)
			}
			if called {
				t.Errorf("execSelf should not run when OLC gate rejects (state=%s)", tc.name)
			}
		})
	}
}

// ---- Test 4: HotbootAlreadyInProgress --------------------------------

func TestDoHotboot_HotbootAlreadyInProgress(t *testing.T) {
	room, cleanup := setupHotbootWorld(t)
	defer cleanup()
	ch, client := makeImmortalHotbootChar("Imp", room)
	defer client.Close()

	WorldRef.SysData.HotbootInProgress = true

	called := false
	execSelf = func(_ string, _ []string, _ []string) error { called = true; return nil }

	DoHotboot(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(strings.ToLower(out), "already in progress") {
		t.Errorf("expected already-in-progress message; got %q", out)
	}
	if called {
		t.Error("execSelf should not run when concurrency gate rejects")
	}
}

// ---- Test 5: HappyPath_CallsSave -------------------------------------

func TestDoHotboot_HappyPath_CallsSave(t *testing.T) {
	room, cleanup := setupHotbootWorld(t)
	defer cleanup()
	ch, client := makeImmortalHotbootChar("Imp", room)
	defer client.Close()
	stopDrain := startDrain(client)
	defer stopDrain()

	_ = stubPauseSuccess(t)
	HotbootResumeFunc = func(*os.File) error { return nil }
	HotbootPort = 4000
	SaveFunc = func(_ *types.CharData) {}

	var capturedArgv []string
	var capturedHotbootFlag bool
	execSelf = func(_ string, argv []string, _ []string) error {
		capturedArgv = append(capturedArgv, argv...)
		// At this moment, mid-exec, HotbootInProgress should still be true.
		capturedHotbootFlag = WorldRef.SysData.HotbootInProgress
		return nil
	}

	DoHotboot(ch, "")

	// hotboot.dat exists.
	if _, err := os.Stat(filepath.Join(WorldRef.DataDir, "hotboot", "hotboot.dat")); err != nil {
		t.Errorf("hotboot.dat missing: %v", err)
	}
	// mobfile.dat exists.
	if _, err := os.Stat(filepath.Join(WorldRef.DataDir, "hotboot", "mobfile.dat")); err != nil {
		t.Errorf("mobfile.dat missing: %v", err)
	}
	// argv includes --hotboot-recover.
	joined := strings.Join(capturedArgv, " ")
	if !strings.Contains(joined, "--hotboot-recover") {
		t.Errorf("argv missing --hotboot-recover: %q", joined)
	}
	if !capturedHotbootFlag {
		t.Error("HotbootInProgress should be true at exec-time (captured in stub)")
	}
}

// ---- Test 6: HappyPath_FlipsHotbootFlag -----------------------------

func TestDoHotboot_HappyPath_FlipsHotbootFlag(t *testing.T) {
	room, cleanup := setupHotbootWorld(t)
	defer cleanup()
	ch, client := makeImmortalHotbootChar("Imp", room)
	defer client.Close()
	stopDrain := startDrain(client)
	defer stopDrain()

	_ = stubPauseSuccess(t)
	HotbootResumeFunc = func(*os.File) error { return nil }
	HotbootPort = 4000

	var flagsAtSave []bool
	SaveFunc = func(c *types.CharData) {
		if c.PCData != nil {
			flagsAtSave = append(flagsAtSave, c.PCData.Hotboot)
		}
	}

	execSelf = func(_ string, _ []string, _ []string) error { return nil }

	DoHotboot(ch, "")

	if len(flagsAtSave) == 0 {
		t.Fatal("SaveFunc never called")
	}
	for i, f := range flagsAtSave {
		if !f {
			t.Errorf("flagsAtSave[%d] = false, want true (PCData.Hotboot must be set BEFORE SavePlayer)", i)
		}
	}
}

// ---- Test 7: ExecFailure_ResumesServer -------------------------------

func TestDoHotboot_ExecFailure_ResumesServer(t *testing.T) {
	room, cleanup := setupHotbootWorld(t)
	defer cleanup()
	ch, client := makeImmortalHotbootChar("Imp", room)
	defer client.Close()
	stopDrain := startDrain(client)
	defer stopDrain()

	_ = stubPauseSuccess(t)
	HotbootPort = 4000
	SaveFunc = func(_ *types.CharData) {}

	resumed := false
	HotbootResumeFunc = func(_ *os.File) error {
		resumed = true
		return nil
	}

	execSelf = func(_ string, _ []string, _ []string) error {
		return os.ErrNotExist // any error
	}

	DoHotboot(ch, "")

	if !resumed {
		t.Error("ResumeFromPause should be called on exec failure")
	}
	if WorldRef.SysData.HotbootInProgress {
		t.Error("HotbootInProgress should be cleared after exec failure")
	}
}

// ---- Test 8: MidNannyDropsDescriptors --------------------------------

func TestDoHotboot_MidNannyDropsDescriptors(t *testing.T) {
	room, cleanup := setupHotbootWorld(t)
	defer cleanup()
	ch, client := makeImmortalHotbootChar("Imp", room)
	defer client.Close()
	stopDrain := startDrain(client)
	defer stopDrain()

	// Seed a CON_GET_NAME descriptor with NO character attached.
	// In real code this is a mid-nanny (pre-login) descriptor.
	nannyServer, nannyClient := net.Pipe()
	defer nannyClient.Close()
	nannyDesc := &types.DescriptorData{
		Conn:       nannyServer,
		Connected:  types.CON_GET_NAME,
		InputQueue: make(chan string, 1),
	}
	WorldRef.Descriptors = append(WorldRef.Descriptors, nannyDesc)

	// Buffer what the nanny-client receives on its end.
	got := make(chan []byte, 1)
	go func() {
		buf := make([]byte, 256)
		n, _ := nannyClient.Read(buf)
		got <- buf[:n]
	}()

	_ = stubPauseSuccess(t)
	HotbootResumeFunc = func(*os.File) error { return nil }
	HotbootPort = 4000
	SaveFunc = func(_ *types.CharData) {}
	execSelf = func(_ string, _ []string, _ []string) error { return nil }

	DoHotboot(ch, "")

	select {
	case msg := <-got:
		if !strings.Contains(string(msg), "Sorry, we are rebooting") {
			t.Errorf("dropped nanny did not receive expected message; got %q", string(msg))
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("nanny client never received drop message")
	}

	// hotboot.dat should contain ONE session (the playing descriptor),
	// not two. Reading it back via persist.LoadHotbootSessions.
	sessions, err := persist.LoadHotbootSessions(WorldRef.DataDir)
	if err != nil {
		t.Fatalf("LoadHotbootSessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Errorf("session count = %d, want 1", len(sessions))
	}
}
