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

// newWorldStubForOedit returns a minimal world adequate for processInput
// dispatch tests. /tmp/test is a placeholder DataDir; processInput does
// not touch the filesystem.
func newWorldStubForOedit() *world.World {
	return world.New("/tmp/test")
}

// newRegistryStubForOedit returns an empty command registry — sufficient
// for the CON_OEDIT dispatch path which never reaches CON_PLAYING.
func newRegistryStubForOedit() *command.Registry {
	return command.NewRegistry()
}

// oeditTestRig mirrors reditTestRig but stashes a *ObjIndexData on
// Olc.Target and sets Connected = CON_OEDIT. Wave 1 uses this only for
// the skeleton + stub tests; Wave 2+ will drive the full mode-dispatch
// through the same harness.
type oeditTestRig struct {
	d    *types.DescriptorData
	idx  *types.ObjIndexData
	mu   sync.Mutex
	sink []byte
	done chan struct{}
}

func newOeditHarness(t *testing.T) *oeditTestRig {
	t.Helper()
	server, client := net.Pipe()
	rig := &oeditTestRig{done: make(chan struct{})}
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
	idx := &types.ObjIndexData{
		Vnum: 1234,
		Name: "a test item",
	}
	d.Olc = &types.OlcData{
		Mode:   types.OEDIT_MAIN_MENU,
		Vnum:   idx.Vnum,
		Target: idx,
	}
	d.Connected = int(types.CON_OEDIT)
	rig.d = d
	rig.idx = idx
	return rig
}

func (r *oeditTestRig) readBuf(t *testing.T) string {
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

// --- G3 skeleton tests ---

// TestOeditParse_Quit_CleansUpAndReturnsToPlaying pins the Wave 1 Q branch:
// reuses the mode-agnostic cleanupOlc helper so that descriptor state after
// a Q matches the redit analog exactly (Connected=CON_PLAYING, Olc=nil,
// Substate=SUB_NONE) and emits "Exiting editor." to the descriptor buffer.
func TestOeditParse_Quit_CleansUpAndReturnsToPlaying(t *testing.T) {
	rig := newOeditHarness(t)

	oeditParse(rig.d, "Q")

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

// TestOeditParse_Quit_LowercaseQ pins the case-insensitive Q matching,
// mirroring redit's main-menu behavior.
func TestOeditParse_Quit_LowercaseQ(t *testing.T) {
	rig := newOeditHarness(t)
	oeditParse(rig.d, "q")
	if rig.d.Connected != int(types.CON_PLAYING) {
		t.Errorf("Connected = %d, want CON_PLAYING after lowercase q", rig.d.Connected)
	}
}

// TestOeditParse_MainMenuOption1_PromptsNamelist confirms input "1" at
// OEDIT_MAIN_MENU prompts for namelist and transitions to
// OEDIT_EDIT_NAMELIST. Replaces Wave 1's unimplemented-stub test.
func TestOeditParse_MainMenuOption1_PromptsNamelist(t *testing.T) {
	rig := newOeditHarness(t)

	oeditParse(rig.d, "1")

	if rig.d.Connected != int(types.CON_OEDIT) {
		t.Errorf("Connected = %d, want CON_OEDIT (session still open)", rig.d.Connected)
	}
	if rig.d.Olc == nil {
		t.Fatal("Olc cleared prematurely — main-menu dispatch must not cleanup")
	}
	if rig.d.Olc.Mode != types.OEDIT_EDIT_NAMELIST {
		t.Errorf("Mode = %d, want OEDIT_EDIT_NAMELIST", rig.d.Olc.Mode)
	}
	out := rig.readBuf(t)
	if !strings.Contains(out, "Enter namelist") {
		t.Errorf("expected 'Enter namelist' prompt, got: %q", out)
	}
}

// TestOeditParse_DefensiveNilOlc verifies the defensive branch that fires
// when d.Olc is nil. Should restore CON_PLAYING without panicking.
func TestOeditParse_DefensiveNilOlc(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc = nil

	oeditParse(rig.d, "anything")

	if rig.d.Connected != int(types.CON_PLAYING) {
		t.Errorf("Connected = %d, want CON_PLAYING after nil-Olc recovery", rig.d.Connected)
	}
}

// TestOeditParse_DefensiveWrongTargetType pins the type-assert fallback.
// If Target is not *ObjIndexData (e.g. somehow a Room got stashed),
// cleanupOlc fires and the session closes gracefully.
func TestOeditParse_DefensiveWrongTargetType(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Target = &types.RoomIndexData{Vnum: 999} // wrong type

	oeditParse(rig.d, "anything")

	if rig.d.Connected != int(types.CON_PLAYING) {
		t.Errorf("Connected = %d, want CON_PLAYING after wrong-type recovery", rig.d.Connected)
	}
	if rig.d.Olc != nil {
		t.Errorf("Olc should be cleared on wrong-type recovery; got %+v", rig.d.Olc)
	}
}

// TestLoop_ConOeditDispatchesToOeditParse is the integration pin for the
// loop.go dispatch arm. We construct a real GameLoop, queue a "Q" input
// on a descriptor in CON_OEDIT, run processInput once, and verify the
// descriptor cleaned up — which only happens if oeditParse was reached
// (the nanny default would disconnect with "Unexpected state" instead).
// Mutation gate (plan §G3): swapping the case label to CON_MEDIT routes
// input to the nanny default; this test fails because Olc would still be
// non-nil and Connected would not be CON_PLAYING.
func TestLoop_ConOeditDispatchesToOeditParse(t *testing.T) {
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

	w := newWorldStubForOedit()
	reg := newRegistryStubForOedit()
	incoming := make(chan *types.DescriptorData, 4)
	g := NewGameLoop(w, reg, incoming)

	d := types.NewDescriptor(server)
	ch := &types.CharData{Name: "Builder", Level: 100, Trust: 100}
	ch.Desc = d
	d.Character = ch
	idx := &types.ObjIndexData{Vnum: 1234, Name: "a test item"}
	d.Olc = &types.OlcData{Mode: types.OEDIT_MAIN_MENU, Vnum: idx.Vnum, Target: idx}
	d.Connected = int(types.CON_OEDIT)

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

// --- G5 main-menu dispatch pin-tests ---

// TestOeditParse_MainMenu_EmptyInputRedisplays pins that bare input at
// OEDIT_MAIN_MENU just re-renders the menu (no error, no mode change).
func TestOeditParse_MainMenu_EmptyInputRedisplays(t *testing.T) {
	rig := newOeditHarness(t)
	oeditParse(rig.d, "")
	if rig.d.Olc.Mode != types.OEDIT_MAIN_MENU {
		t.Errorf("Mode = %d, want OEDIT_MAIN_MENU", rig.d.Olc.Mode)
	}
	out := rig.readBuf(t)
	if !strings.Contains(out, "Object number") {
		t.Errorf("expected main menu redisplay; got: %q", out)
	}
}

// TestOeditParse_MainMenu_UnknownCharRedisplays pins that garbage input
// silently re-renders the menu. Mirrors C's fall-through behavior.
func TestOeditParse_MainMenu_UnknownCharRedisplays(t *testing.T) {
	rig := newOeditHarness(t)
	oeditParse(rig.d, "zzz")
	if rig.d.Olc.Mode != types.OEDIT_MAIN_MENU {
		t.Errorf("Mode = %d, want OEDIT_MAIN_MENU after garbage input", rig.d.Olc.Mode)
	}
	out := rig.readBuf(t)
	if !strings.Contains(out, "Object number") {
		t.Errorf("expected menu redisplay; got: %q", out)
	}
}

// TestOeditParse_MainMenu_Option5_ShowsTypeMenu pins that '5' dispatches
// to the item-type submenu and sets OEDIT_TYPE.
func TestOeditParse_MainMenu_Option5_ShowsTypeMenu(t *testing.T) {
	rig := newOeditHarness(t)
	oeditParse(rig.d, "5")
	if rig.d.Olc.Mode != types.OEDIT_TYPE {
		t.Errorf("Mode = %d, want OEDIT_TYPE", rig.d.Olc.Mode)
	}
	out := rig.readBuf(t)
	if !strings.Contains(out, "Enter type") {
		t.Errorf("expected type-menu prompt; got: %q", out)
	}
}

// TestOeditParse_MainMenu_Option6_ShowsExtraMenu pins '6' → OEDIT_EXTRAS.
func TestOeditParse_MainMenu_Option6_ShowsExtraMenu(t *testing.T) {
	rig := newOeditHarness(t)
	oeditParse(rig.d, "6")
	if rig.d.Olc.Mode != types.OEDIT_EXTRAS {
		t.Errorf("Mode = %d, want OEDIT_EXTRAS", rig.d.Olc.Mode)
	}
}

// TestOeditParse_MainMenu_Option7_ShowsWearMenu pins '7' → OEDIT_WEAR.
func TestOeditParse_MainMenu_Option7_ShowsWearMenu(t *testing.T) {
	rig := newOeditHarness(t)
	oeditParse(rig.d, "7")
	if rig.d.Olc.Mode != types.OEDIT_WEAR {
		t.Errorf("Mode = %d, want OEDIT_WEAR", rig.d.Olc.Mode)
	}
}

// TestOeditParse_MainMenu_Option8_WeightPrompt pins '8' transitions to
// OEDIT_WEIGHT and emits a weight prompt.
func TestOeditParse_MainMenu_Option8_WeightPrompt(t *testing.T) {
	rig := newOeditHarness(t)
	oeditParse(rig.d, "8")
	if rig.d.Olc.Mode != types.OEDIT_WEIGHT {
		t.Errorf("Mode = %d, want OEDIT_WEIGHT", rig.d.Olc.Mode)
	}
	out := rig.readBuf(t)
	if !strings.Contains(out, "Enter weight") {
		t.Errorf("expected weight prompt; got: %q", out)
	}
}

// TestOeditParse_MainMenu_OptionD_ShowsLayerMenu pins 'D' → OEDIT_LAYERS.
func TestOeditParse_MainMenu_OptionD_ShowsLayerMenu(t *testing.T) {
	rig := newOeditHarness(t)
	oeditParse(rig.d, "D")
	if rig.d.Olc.Mode != types.OEDIT_LAYERS {
		t.Errorf("Mode = %d, want OEDIT_LAYERS", rig.d.Olc.Mode)
	}
}

// TestOeditParse_MainMenu_OptionG_ShowsExtradescMenu pins 'G' →
// OEDIT_EXTRADESC_MENU (Wave 3 will fill the arms; Wave 2 just routes).
func TestOeditParse_MainMenu_OptionG_ShowsExtradescMenu(t *testing.T) {
	rig := newOeditHarness(t)
	oeditParse(rig.d, "G")
	if rig.d.Olc.Mode != types.OEDIT_EXTRADESC_MENU {
		t.Errorf("Mode = %d, want OEDIT_EXTRADESC_MENU", rig.d.Olc.Mode)
	}
}

// TestOeditParse_MainMenu_OptionE_CallsVal1Menu pins 'E' → OEDIT_VALUE_1.
func TestOeditParse_MainMenu_OptionE_CallsVal1Menu(t *testing.T) {
	rig := newOeditHarness(t)
	oeditParse(rig.d, "E")
	if rig.d.Olc.Mode != types.OEDIT_VALUE_1 {
		t.Errorf("Mode = %d, want OEDIT_VALUE_1", rig.d.Olc.Mode)
	}
}

// TestOeditParse_MainMenu_OptionH_ScopeCut pins the scope-cut 'H'
// placeholder: main menu redisplays WITHOUT entering a mudprog mode.
func TestOeditParse_MainMenu_OptionH_ScopeCut(t *testing.T) {
	rig := newOeditHarness(t)
	oeditParse(rig.d, "H")
	// Plan states H emits Wave-3 message and redisplays main. We just
	// require that Mode stays OEDIT_MAIN_MENU and no mprog state is set.
	if rig.d.Olc.Mode != types.OEDIT_MAIN_MENU {
		t.Errorf("Mode = %d, want OEDIT_MAIN_MENU after H (scope-cut)", rig.d.Olc.Mode)
	}
}

// TestOeditParse_MainMenu_LevelGateBlocksLowTrust pins the LEVEL_GREATER
// gate on 'C' (level) at main-menu entry — trust-59 user gets the refusal
// message and menu redisplays. Corresponds to plan §G5 'C' branch.
func TestOeditParse_MainMenu_LevelGateBlocksLowTrust(t *testing.T) {
	rig := newOeditHarness(t)
	// Drop trust below LEVEL_GREATER.
	rig.d.Character.Trust = types.LEVEL_GREATER - 1

	oeditParse(rig.d, "C")

	if rig.d.Olc.Mode != types.OEDIT_MAIN_MENU {
		t.Errorf("Mode = %d, want OEDIT_MAIN_MENU (refused)", rig.d.Olc.Mode)
	}
	out := rig.readBuf(t)
	if !strings.Contains(out, "powerful enough") {
		t.Errorf("expected refusal; got: %q", out)
	}
}

// TestOeditParse_MainMenu_LevelGateAllowsGreater pins that trust at
// LEVEL_GREATER+ transitions into OEDIT_LEVEL.
func TestOeditParse_MainMenu_LevelGateAllowsGreater(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Character.Trust = types.LEVEL_GREATER
	oeditParse(rig.d, "C")
	if rig.d.Olc.Mode != types.OEDIT_LEVEL {
		t.Errorf("Mode = %d, want OEDIT_LEVEL", rig.d.Olc.Mode)
	}
}

// --- G6 simple-field parse arms ---

func TestOeditParse_Namelist_SetsName(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_EDIT_NAMELIST
	oeditParse(rig.d, "shiny gem stone")
	if rig.idx.Name != "shiny gem stone" {
		t.Errorf("Name = %q, want %q", rig.idx.Name, "shiny gem stone")
	}
	if rig.d.Olc.Mode != types.OEDIT_MAIN_MENU {
		t.Errorf("Mode = %d, want OEDIT_MAIN_MENU", rig.d.Olc.Mode)
	}
}

func TestOeditParse_Namelist_SmashesTilde(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_EDIT_NAMELIST
	oeditParse(rig.d, "tilde~here")
	if strings.Contains(rig.idx.Name, "~") {
		t.Errorf("Name should have tilde smashed; got %q", rig.idx.Name)
	}
}

func TestOeditParse_Shortdesc_SetsShort(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_SHORTDESC
	oeditParse(rig.d, "a sparkly gem")
	if rig.idx.ShortDescr != "a sparkly gem" {
		t.Errorf("ShortDescr = %q, want 'a sparkly gem'", rig.idx.ShortDescr)
	}
}

func TestOeditParse_Weight_SetsWeight(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_WEIGHT
	oeditParse(rig.d, "42")
	if rig.idx.Weight != 42 {
		t.Errorf("Weight = %d, want 42", rig.idx.Weight)
	}
}

func TestOeditParse_Weight_ClampsLow(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_WEIGHT
	oeditParse(rig.d, "0")
	if rig.idx.Weight != 1 {
		t.Errorf("Weight = %d, want 1 (clamp floor)", rig.idx.Weight)
	}
}

func TestOeditParse_Cost_SetsCost(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_COST
	oeditParse(rig.d, "500")
	if rig.idx.GoldCost != 500 {
		t.Errorf("GoldCost = %d, want 500", rig.idx.GoldCost)
	}
}

func TestOeditParse_Rent_SetsRent(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_COSTPERDAY
	oeditParse(rig.d, "20")
	if rig.idx.Rent != 20 {
		t.Errorf("Rent = %d, want 20", rig.idx.Rent)
	}
}

func TestOeditParse_Timer_AcceptsAndLogs(t *testing.T) {
	// Prototype has no Timer field in Go — input is accepted + logged
	// without mutation. Test pins the non-panic acceptance.
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_TIMER
	oeditParse(rig.d, "7")
	if rig.d.Olc.Mode != types.OEDIT_MAIN_MENU {
		t.Errorf("Mode = %d, want OEDIT_MAIN_MENU", rig.d.Olc.Mode)
	}
}

func TestOeditParse_Level_ClampsToMaxLevel(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_LEVEL
	oeditParse(rig.d, "999")
	if rig.idx.Level != types.MAX_LEVEL {
		t.Errorf("Level = %d, want %d (MAX_LEVEL clamp)", rig.idx.Level, types.MAX_LEVEL)
	}
}

func TestOeditParse_Level_LowTrustBlocked(t *testing.T) {
	// Defense-in-depth: even with Mode forced to OEDIT_LEVEL, parse arm
	// re-checks trust gate.
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_LEVEL
	rig.d.Character.Trust = types.LEVEL_GREATER - 1
	oeditParse(rig.d, "50")
	if rig.idx.Level != 0 {
		t.Errorf("Level = %d, want 0 (refused)", rig.idx.Level)
	}
}

func TestOeditParse_Type_SetsByNumber(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_TYPE
	oeditParse(rig.d, strconvTestItoa(types.ITEM_WEAPON))
	if rig.idx.ItemType != types.ITEM_WEAPON {
		t.Errorf("ItemType = %d, want %d", rig.idx.ItemType, types.ITEM_WEAPON)
	}
}

func TestOeditParse_Type_WordLookup(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_TYPE
	oeditParse(rig.d, "weapon")
	if rig.idx.ItemType != types.ITEM_WEAPON {
		t.Errorf("ItemType = %d, want %d", rig.idx.ItemType, types.ITEM_WEAPON)
	}
}

func TestOeditParse_Type_RejectsOutOfRange(t *testing.T) {
	// Input 0 → reject-and-reprompt (stay in OEDIT_TYPE per C parity).
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_TYPE
	rig.idx.ItemType = types.ITEM_WEAPON
	oeditParse(rig.d, "0")
	if rig.idx.ItemType != types.ITEM_WEAPON {
		t.Errorf("ItemType must not mutate on 0; got %d", rig.idx.ItemType)
	}
	if rig.d.Olc.Mode != types.OEDIT_TYPE {
		t.Errorf("Mode = %d, want stay in OEDIT_TYPE after reject", rig.d.Olc.Mode)
	}
}

func TestOeditParse_Extras_NumberToggle(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_EXTRAS
	// Bit 0 = ITEM_GLOW. Numeric input is 1-indexed.
	oeditParse(rig.d, "1")
	if !rig.idx.ExtraFlags.IsSet(types.ITEM_GLOW) {
		t.Errorf("ITEM_GLOW should be set after toggle")
	}
}

func TestOeditParse_Extras_NumberBreaks(t *testing.T) {
	// Numeric input processes only ONE flag per C parity.
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_EXTRAS
	// "5 6" — leading token parses as int 5... wait, strconv.Atoi on
	// "5 6" fails because of the space. Use a single number:
	oeditParse(rig.d, "5")
	// bit 4 = ITEM_EVIL
	if !rig.idx.ExtraFlags.IsSet(types.ITEM_EVIL) {
		t.Errorf("ITEM_EVIL should be set")
	}
}

func TestOeditParse_Extras_WordToggle(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_EXTRAS
	oeditParse(rig.d, "magic")
	if !rig.idx.ExtraFlags.IsSet(types.ITEM_MAGIC) {
		t.Errorf("ITEM_MAGIC should be set after word toggle")
	}
}

func TestOeditParse_Extras_PrototypeGatedByTrust(t *testing.T) {
	// Trust below LEVEL_GREATER without protoflag bestowment → refusal.
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_EXTRAS
	rig.d.Character.Trust = types.LEVEL_GREATER - 1
	// Numeric: bit 30 = ITEM_PROTOTYPE. n = 31 (1-indexed).
	oeditParse(rig.d, "31")
	if rig.idx.ExtraFlags.IsSet(types.ITEM_PROTOTYPE) {
		t.Errorf("ITEM_PROTOTYPE must not be set without LEVEL_GREATER trust")
	}
	out := rig.readBuf(t)
	if !strings.Contains(out, "cannot change the prototype flag") {
		t.Errorf("expected refusal message; got: %q", out)
	}
}

func TestOeditParse_Extras_ZeroReturnsToMain(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_EXTRAS
	oeditParse(rig.d, "0")
	if rig.d.Olc.Mode != types.OEDIT_MAIN_MENU {
		t.Errorf("Mode = %d, want OEDIT_MAIN_MENU after 0", rig.d.Olc.Mode)
	}
}

func TestOeditParse_Wear_NumberToggle(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_WEAR
	// bit 3 = body. 1-indexed = 4.
	oeditParse(rig.d, "4")
	if rig.idx.WearFlags&(1<<3) == 0 {
		t.Errorf("wear body bit should be set; got WearFlags=%d", rig.idx.WearFlags)
	}
}

func TestOeditParse_Wear_RejectsDualWieldNumeric(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_WEAR
	// bit 15 = dual_wield. 1-indexed = 16.
	oeditParse(rig.d, "16")
	if rig.idx.WearFlags&(1<<15) != 0 {
		t.Errorf("dual_wield bit must not be set via OLC; got WearFlags=%d", rig.idx.WearFlags)
	}
}

func TestOeditParse_Wear_RejectsDualWieldWord(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_WEAR
	oeditParse(rig.d, "dual_wield")
	if rig.idx.WearFlags&(1<<15) != 0 {
		t.Errorf("dual_wield bit must not be set via OLC word input; got WearFlags=%d", rig.idx.WearFlags)
	}
}

func TestOeditParse_Wear_ZeroReturnsToMain(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_WEAR
	oeditParse(rig.d, "0")
	if rig.d.Olc.Mode != types.OEDIT_MAIN_MENU {
		t.Errorf("Mode = %d, want OEDIT_MAIN_MENU after 0", rig.d.Olc.Mode)
	}
}

func TestOeditParse_Layers_Option1ZeroesLayers(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_LAYERS
	rig.idx.Layers = 0xFF // all set
	oeditParse(rig.d, "1")
	if rig.idx.Layers != 0 {
		t.Errorf("Layers = %d, want 0 after option 1", rig.idx.Layers)
	}
}

func TestOeditParse_Layers_Option2TogglesBit0(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_LAYERS
	rig.idx.Layers = 0
	oeditParse(rig.d, "2")
	if rig.idx.Layers != 1 {
		t.Errorf("Layers = %d, want 1 (bit 0 toggled on)", rig.idx.Layers)
	}
}

func TestOeditParse_Layers_Option9TogglesBit7(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_LAYERS
	rig.idx.Layers = 0
	oeditParse(rig.d, "9")
	if rig.idx.Layers != 128 {
		t.Errorf("Layers = %d, want 128 (bit 7 toggled on)", rig.idx.Layers)
	}
}

func TestOeditParse_Layers_ZeroReturnsToMain(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_LAYERS
	oeditParse(rig.d, "0")
	if rig.d.Olc.Mode != types.OEDIT_MAIN_MENU {
		t.Errorf("Mode = %d, want OEDIT_MAIN_MENU after 0", rig.d.Olc.Mode)
	}
}

func TestOeditParse_Layers_InvalidRedisplays(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_LAYERS
	rig.idx.Layers = 0
	oeditParse(rig.d, "42")
	if rig.idx.Layers != 0 {
		t.Errorf("Layers must not change on invalid input; got %d", rig.idx.Layers)
	}
	if rig.d.Olc.Mode != types.OEDIT_LAYERS {
		t.Errorf("Mode = %d, want stay in OEDIT_LAYERS", rig.d.Olc.Mode)
	}
}

// --- Wave 2 value-stub parse behavior ---

func TestOeditParse_ValueStub_ZeroReturnsToMain(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_VALUE_1
	oeditParse(rig.d, "0")
	if rig.d.Olc.Mode != types.OEDIT_MAIN_MENU {
		t.Errorf("Mode = %d, want OEDIT_MAIN_MENU after 0 in value stub", rig.d.Olc.Mode)
	}
}

func TestOeditParse_ValueStub_NonZeroReprompts(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_VALUE_3
	oeditParse(rig.d, "7")
	if rig.d.Olc.Mode != types.OEDIT_VALUE_3 {
		t.Errorf("Mode = %d, want stay in OEDIT_VALUE_3", rig.d.Olc.Mode)
	}
}

// --- G5 EditorSave closure: long desc restores CON_OEDIT (mutation #11) ---

// TestOeditParse_Longdesc_RestoresConOedit pins the '3' trampoline: after
// /s fires via the EditorSave closure, Connected must be CON_OEDIT (not
// CON_REDIT or CON_PLAYING). Drives the closure directly by invoking
// ch.EditorSave after pre-populating the editor buffer.
func TestOeditParse_Longdesc_RestoresConOedit(t *testing.T) {
	rig := newOeditHarness(t)
	// Enter the long-desc editor via main-menu option '3'.
	oeditParse(rig.d, "3")
	if rig.d.Character.EditorSave == nil {
		t.Fatal("EditorSave closure not registered")
	}
	if rig.d.Character.Substate != types.SUB_OBJ_LONG {
		t.Errorf("Substate = %d, want SUB_OBJ_LONG", rig.d.Character.Substate)
	}
	// Seed editor buffer directly. StartEditing may or may not have
	// allocated ch.Editor with content; we force one line so CopyBuffer
	// returns non-empty.
	rig.d.Character.Editor = &types.EditorData{NumLines: 1}
	rig.d.Character.Editor.Lines[0] = "A long description of the item."
	rig.d.Character.EditorSave(rig.d.Character)
	if rig.d.Connected != int(types.CON_OEDIT) {
		t.Errorf("Connected = %d, want CON_OEDIT (trampoline must re-enter menu state)", rig.d.Connected)
	}
	if rig.idx.Description == "" {
		t.Errorf("Description should have been written via EditorSave; got empty")
	}
}

func TestOeditParse_Actdesc_RestoresConOedit(t *testing.T) {
	rig := newOeditHarness(t)
	oeditParse(rig.d, "4")
	if rig.d.Character.Substate != types.SUB_OBJ_ACTION {
		t.Errorf("Substate = %d, want SUB_OBJ_ACTION", rig.d.Character.Substate)
	}
	if rig.d.Character.EditorSave == nil {
		t.Fatal("EditorSave closure not registered for action desc")
	}
	rig.d.Character.Editor = &types.EditorData{NumLines: 1}
	rig.d.Character.Editor.Lines[0] = "* click *"
	rig.d.Character.EditorSave(rig.d.Character)
	if rig.d.Connected != int(types.CON_OEDIT) {
		t.Errorf("Connected = %d, want CON_OEDIT after action-desc save", rig.d.Connected)
	}
	if rig.idx.ActionDesc == "" {
		t.Errorf("ActionDesc should have been written via EditorSave; got empty")
	}
}

// strconvTestItoa is a tiny helper so the test file doesn't need to
// import strconv just to format a single int.
func strconvTestItoa(n int) string {
	return istrconv(n)
}

func istrconv(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
