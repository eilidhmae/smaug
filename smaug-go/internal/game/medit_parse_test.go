package game

import (
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/command"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// newWorldStubForMedit returns a minimal world adequate for processInput
// dispatch tests. /tmp/test is a placeholder DataDir; processInput does
// not touch the filesystem.
func newWorldStubForMedit() *world.World {
	return world.New("/tmp/test")
}

// newRegistryStubForMedit returns an empty command registry — sufficient
// for the CON_MEDIT dispatch path which never reaches CON_PLAYING during
// Wave 1 tests.
func newRegistryStubForMedit() *command.Registry {
	return command.NewRegistry()
}

// meditTestRig parallels oeditTestRig but stashes a *CharData (NPC by
// default — Act|ACT_IS_NPC) on Olc.Target and sets Connected = CON_MEDIT.
// Wave 1 uses this only for the skeleton + stub tests; Wave 2+ will drive
// the full mode-dispatch through the same harness once the menu content
// is wired.
type meditTestRig struct {
	d      *types.DescriptorData
	victim *types.CharData
	mu     sync.Mutex
	sink   []byte
	done   chan struct{}
}

// newMeditHarness builds an NPC-victim harness (default). For PC-victim
// tests the caller flips victim.Act to clear ACT_IS_NPC after construction.
func newMeditHarness(t *testing.T) *meditTestRig {
	t.Helper()
	server, client := net.Pipe()
	rig := &meditTestRig{done: make(chan struct{})}
	t.Cleanup(func() {
		server.Close()
		client.Close()
		<-rig.done
	})

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

	// Victim is an NPC prototype — ACT_IS_NPC set on the Act flag. This
	// matches the shape medit_parse will dispatch against in Wave 2's
	// NPC-main-menu branch. PC-victim tests flip Act back to clear the
	// bit after newMeditHarness returns.
	victim := &types.CharData{Name: "a test mob", Level: 10}
	victim.Act.Set(types.ACT_IS_NPC)

	d.Olc = &types.OlcData{
		Mode:   types.MEDIT_NPC_MAIN_MENU,
		Vnum:   1234,
		Target: victim,
	}
	d.Connected = int(types.CON_MEDIT)
	rig.d = d
	rig.victim = victim
	return rig
}

func (r *meditTestRig) readBuf(t *testing.T) string {
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

// --- G4 skeleton tests ---

// TestMeditParse_Quit_CleansUpAndReturnsToPlaying pins the Wave 1 Q branch:
// reuses the mode-agnostic cleanupOlc helper so descriptor state after a Q
// matches the redit/oedit analog exactly (Connected=CON_PLAYING, Olc=nil,
// Substate=SUB_NONE) and emits "Exiting editor." to the descriptor buffer.
func TestMeditParse_Quit_CleansUpAndReturnsToPlaying(t *testing.T) {
	rig := newMeditHarness(t)

	meditParse(rig.d, "Q")

	if rig.d.Connected != int(types.CON_PLAYING) {
		t.Errorf("Connected = %d, want CON_PLAYING", rig.d.Connected)
	}
	if rig.d.Olc != nil {
		t.Errorf("Olc should be nil after Q; got %+v", rig.d.Olc)
	}
	if rig.d.Character.Substate != types.SUB_NONE {
		t.Errorf("Substate = %d, want SUB_NONE", rig.d.Character.Substate)
	}
	out := rig.readBuf(t)
	if !strings.Contains(out, "Exiting editor.") {
		t.Errorf("expected 'Exiting editor.' in output, got: %q", out)
	}
}

// TestMeditParse_Quit_LowercaseQ pins the case-insensitive Q matching,
// mirroring redit / oedit main-menu behavior.
func TestMeditParse_Quit_LowercaseQ(t *testing.T) {
	rig := newMeditHarness(t)
	meditParse(rig.d, "q")
	if rig.d.Connected != int(types.CON_PLAYING) {
		t.Errorf("Connected = %d, want CON_PLAYING after lowercase q", rig.d.Connected)
	}
	if rig.d.Olc != nil {
		t.Error("Olc should be nil after lowercase q")
	}
}

// TestMeditParse_Wave2Dispatch_DigitOneSetsSex pins the Wave-2 replacement
// for the Wave-1 "unimplemented stub" test. Wave 1 made input "1" emit a
// placeholder; Wave 2's NPC main-menu dispatcher MUST transition to
// MEDIT_SEX per plan §230-269 / C omedit.c:1065-1067.
func TestMeditParse_Wave2Dispatch_DigitOneSetsSex(t *testing.T) {
	rig := newMeditHarness(t)

	meditParse(rig.d, "1")

	if rig.d.Connected != int(types.CON_MEDIT) {
		t.Errorf("Connected = %d, want CON_MEDIT (session still open)", rig.d.Connected)
	}
	if rig.d.Olc == nil {
		t.Fatal("Olc cleared — should still be in CON_MEDIT after digit 1")
	}
	if rig.d.Olc.Mode != types.MEDIT_SEX {
		t.Errorf("Mode = %d, want MEDIT_SEX after digit 1 on NPC menu", rig.d.Olc.Mode)
	}
	_ = rig.readBuf(t) // drain output buffer to keep harness tidy
}

// TestMeditParse_DefensiveNilOlc verifies the defensive branch that fires
// when d.Olc is nil. Should restore CON_PLAYING without panicking. Does
// NOT consult cleanupOlc (which would clear Olc that's already nil — no
// visible effect) — the branch directly restores Connected.
func TestMeditParse_DefensiveNilOlc(t *testing.T) {
	rig := newMeditHarness(t)
	rig.d.Olc = nil

	meditParse(rig.d, "anything")

	if rig.d.Connected != int(types.CON_PLAYING) {
		t.Errorf("Connected = %d, want CON_PLAYING after nil-Olc recovery", rig.d.Connected)
	}
}

// TestMeditParse_DefensiveWrongTargetType pins the type-assert fallback.
// If Target is not *CharData (e.g. somehow an ObjIndexData got stashed),
// cleanupOlc fires and the session closes gracefully. Matches oedit's
// wrong-type branch shape.
func TestMeditParse_DefensiveWrongTargetType(t *testing.T) {
	rig := newMeditHarness(t)
	rig.d.Olc.Target = &types.ObjIndexData{Vnum: 999} // wrong type

	meditParse(rig.d, "anything")

	if rig.d.Connected != int(types.CON_PLAYING) {
		t.Errorf("Connected = %d, want CON_PLAYING after wrong-type recovery", rig.d.Connected)
	}
	if rig.d.Olc != nil {
		t.Errorf("Olc should be cleared on wrong-type recovery; got %+v", rig.d.Olc)
	}
}

// TestLoop_ConMeditDispatchesToMeditParse is the integration pin for the
// loop.go dispatch arm. We construct a real GameLoop, queue a "Q" input
// on a descriptor in CON_MEDIT, run processInput once, and verify the
// descriptor cleaned up — which only happens if meditParse was reached
// (the nanny default would disconnect with "Unexpected state" instead).
// Mutation gate (plan §G4): swapping the case label to CON_OEDIT routes
// input to the nanny default; this test fails because Olc would still be
// non-nil and Connected would not be CON_PLAYING.
func TestLoop_ConMeditDispatchesToMeditParse(t *testing.T) {
	server, client := net.Pipe()
	t.Cleanup(func() {
		server.Close()
		client.Close()
	})
	// Drain client side so writes don't block.
	go func() {
		buf := make([]byte, 4096)
		for {
			if _, err := client.Read(buf); err != nil {
				return
			}
		}
	}()

	w := newWorldStubForMedit()
	reg := newRegistryStubForMedit()
	incoming := make(chan *types.DescriptorData, 4)
	g := NewGameLoop(w, reg, incoming)

	d := types.NewDescriptor(server)
	ch := &types.CharData{Name: "Builder", Level: 100, Trust: 100}
	ch.Desc = d
	d.Character = ch
	victim := &types.CharData{Name: "a test mob", Level: 10}
	victim.Act.Set(types.ACT_IS_NPC)
	d.Olc = &types.OlcData{
		Mode:   types.MEDIT_NPC_MAIN_MENU,
		Vnum:   1234,
		Target: victim,
	}
	d.Connected = int(types.CON_MEDIT)

	w.Descriptors = append(w.Descriptors, d)
	d.InputQueue <- "Q"

	g.processInput()

	if d.Connected != int(types.CON_PLAYING) {
		t.Errorf("Connected = %d, want CON_PLAYING after Q via processInput dispatch (loop arm likely missing)", d.Connected)
	}
	if d.Olc != nil {
		t.Errorf("Olc should be nil after Q; got %+v (loop arm likely routed to nanny)", d.Olc)
	}
}

// -------------- worldPcLookup (plan §G3) --------------

func TestWorldPcLookup_FindsConnectedPc(t *testing.T) {
	w := newWorldStubForMedit()
	prev := worldRef
	worldRef = w
	defer func() { worldRef = prev }()

	pc := &types.CharData{Name: "Eilidh"}
	d := &types.DescriptorData{Connected: int(types.CON_PLAYING), Character: pc}
	w.Descriptors = append(w.Descriptors, d)

	got := worldPcLookup("Eilidh")
	if got != pc {
		t.Errorf("worldPcLookup(Eilidh) = %v, want %v", got, pc)
	}
}

func TestWorldPcLookup_IgnoresLinkdead(t *testing.T) {
	w := newWorldStubForMedit()
	prev := worldRef
	worldRef = w
	defer func() { worldRef = prev }()

	pc := &types.CharData{Name: "Eilidh"}
	// Connected != CON_PLAYING: simulates linkdead / mid-login state.
	d := &types.DescriptorData{Connected: int(types.CON_GET_NAME), Character: pc}
	w.Descriptors = append(w.Descriptors, d)

	if got := worldPcLookup("Eilidh"); got != nil {
		t.Errorf("linkdead descriptor should not match; got %v", got)
	}
}

func TestWorldPcLookup_IgnoresNilCharacter(t *testing.T) {
	w := newWorldStubForMedit()
	prev := worldRef
	worldRef = w
	defer func() { worldRef = prev }()

	d := &types.DescriptorData{Connected: int(types.CON_PLAYING), Character: nil}
	w.Descriptors = append(w.Descriptors, d)

	if got := worldPcLookup("Eilidh"); got != nil {
		t.Errorf("nil-Character descriptor should not match; got %v", got)
	}
}

func TestWorldPcLookup_CaseInsensitive(t *testing.T) {
	w := newWorldStubForMedit()
	prev := worldRef
	worldRef = w
	defer func() { worldRef = prev }()

	pc := &types.CharData{Name: "Eilidh"}
	d := &types.DescriptorData{Connected: int(types.CON_PLAYING), Character: pc}
	w.Descriptors = append(w.Descriptors, d)

	if got := worldPcLookup("eilidh"); got != pc {
		t.Errorf("case-insensitive lookup failed; got %v", got)
	}
	if got := worldPcLookup("EILIDH"); got != pc {
		t.Errorf("case-insensitive lookup failed; got %v", got)
	}
}

func TestWorldPcLookup_NoMatch(t *testing.T) {
	w := newWorldStubForMedit()
	prev := worldRef
	worldRef = w
	defer func() { worldRef = prev }()

	pc := &types.CharData{Name: "Eilidh"}
	d := &types.DescriptorData{Connected: int(types.CON_PLAYING), Character: pc}
	w.Descriptors = append(w.Descriptors, d)

	if got := worldPcLookup("Bob"); got != nil {
		t.Errorf("non-matching name should return nil; got %v", got)
	}
}

func TestWorldPcLookup_NilWorld(t *testing.T) {
	prev := worldRef
	worldRef = nil
	defer func() { worldRef = prev }()

	if got := worldPcLookup("Eilidh"); got != nil {
		t.Errorf("nil worldRef should return nil; got %v", got)
	}
}
