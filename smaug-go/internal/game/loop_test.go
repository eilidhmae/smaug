package game

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/eilidhmae/smaug/internal/act"
	"github.com/eilidhmae/smaug/internal/command"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func newTestLoop() *GameLoop {
	w := world.New("/tmp/test")
	reg := command.NewRegistry()
	incoming := make(chan *types.DescriptorData, 10)
	return NewGameLoop(w, reg, incoming)
}

// --- isValidName tests ---

func TestIsValidName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"normal", "Gandalf", true},
		{"lowercase", "gandalf", true},
		{"uppercase", "GANDALF", true},
		{"min_length", "Bob", true},
		{"max_length", "Abcdefghijkl", true},
		{"too_short", "Ab", false},
		{"too_long", "Abcdefghijklm", false},
		{"empty", "", false},
		{"has_space", "Gan dalf", false},
		{"has_number", "Gandalf1", false},
		{"has_dash", "Gan-dalf", false},
		{"has_underscore", "Gan_dalf", false},
		{"single_char", "A", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidName(tt.input)
			if got != tt.want {
				t.Errorf("isValidName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// --- createNewCharacter tests ---

func TestCreateNewCharacter(t *testing.T) {
	g := newTestLoop()
	ch := g.createNewCharacter("Gandalf")

	if ch.Name != "Gandalf" {
		t.Errorf("Name = %q, want %q", ch.Name, "Gandalf")
	}
	if ch.Level != 1 {
		t.Errorf("Level = %d, want 1", ch.Level)
	}
	if ch.Hit != 20 || ch.MaxHit != 20 {
		t.Errorf("Hit = %d/%d, want 20/20", ch.Hit, ch.MaxHit)
	}
	if ch.PCData == nil {
		t.Fatal("PCData is nil")
	}
	if ch.PCData.Title != "the newbie" {
		t.Errorf("Title = %q, want %q", ch.PCData.Title, "the newbie")
	}
	if ch.PCData.Filename != "gandalf" {
		t.Errorf("Filename = %q, want %q", ch.PCData.Filename, "gandalf")
	}
	if !ch.Act.IsSet(types.PLR_ANSI) {
		t.Error("PLR_ANSI not set")
	}
	if !ch.Act.IsSet(types.PLR_AUTOEXIT) {
		t.Error("PLR_AUTOEXIT not set")
	}
	if ch.PermStr != 13 {
		t.Errorf("PermStr = %d, want 13", ch.PermStr)
	}
	if ch.Position != types.POS_STANDING {
		t.Errorf("Position = %d, want POS_STANDING", ch.Position)
	}
}

// --- applyRaceBonuses tests ---

func TestApplyRaceBonuses(t *testing.T) {
	g := newTestLoop()
	ch := g.createNewCharacter("Test")

	race := &types.RaceData{
		StrPlus: 2,
		DexPlus: -1,
		WisPlus: 0,
		IntPlus: 1,
		ConPlus: 3,
		ChaPlus: -2,
		LckPlus: 0,
		Resist:  42,
		Suscept: 7,
	}
	race.Affected.Set(types.AFF_INFRARED)

	g.applyRaceBonuses(ch, race)

	if ch.PermStr != 15 {
		t.Errorf("PermStr = %d, want 15 (13+2)", ch.PermStr)
	}
	if ch.PermDex != 12 {
		t.Errorf("PermDex = %d, want 12 (13-1)", ch.PermDex)
	}
	if ch.PermInt != 14 {
		t.Errorf("PermInt = %d, want 14 (13+1)", ch.PermInt)
	}
	if ch.PermCon != 16 {
		t.Errorf("PermCon = %d, want 16 (13+3)", ch.PermCon)
	}
	if ch.PermCha != 11 {
		t.Errorf("PermCha = %d, want 11 (13-2)", ch.PermCha)
	}
	if ch.Resistant != 42 {
		t.Errorf("Resistant = %d, want 42", ch.Resistant)
	}
	if ch.Susceptible != 7 {
		t.Errorf("Susceptible = %d, want 7", ch.Susceptible)
	}
	if !ch.AffectedBy.IsSet(types.AFF_INFRARED) {
		t.Error("AFF_INFRARED not set")
	}
}

// --- NewGameLoop tests ---

func TestNewGameLoop(t *testing.T) {
	g := newTestLoop()

	if g.world == nil {
		t.Error("world is nil")
	}
	if g.cmdReg == nil {
		t.Error("cmdReg is nil")
	}
	if g.pulseArea != types.PULSE_AREA {
		t.Errorf("pulseArea = %d, want %d", g.pulseArea, types.PULSE_AREA)
	}
	if g.pulseTick != types.PULSE_TICK {
		t.Errorf("pulseTick = %d, want %d", g.pulseTick, types.PULSE_TICK)
	}
}

// --- flushOutput broken pipe test ---

func TestFlushOutput_BrokenPipe(t *testing.T) {
	g := newTestLoop()

	// Create a descriptor with an already-closed connection
	server, client := net.Pipe()
	client.Close() // Close the client end
	d := types.NewDescriptor(server)
	d.WriteToBuffer("Hello, world!\n\r")
	g.world.Descriptors = append(g.world.Descriptors, d)

	g.flushOutput()

	if d.Connected != -1 {
		t.Errorf("d.Connected = %d, want -1 (should be marked for cleanup after broken pipe)", d.Connected)
	}
	server.Close()
}

// --- closeDescriptor saves player test ---

func TestCloseDescriptor_SavesPlayer(t *testing.T) {
	tmpDir := t.TempDir()
	w := world.New(tmpDir)
	reg := command.NewRegistry()
	incoming := make(chan *types.DescriptorData, 10)
	g := NewGameLoop(w, reg, incoming)

	room := &types.RoomIndexData{Vnum: 21001, Name: "Temple"}
	w.Rooms[21001] = room

	server, client := net.Pipe()
	defer client.Close()
	d := types.NewDescriptor(server)
	ch := &types.CharData{
		Name:     "Savetest",
		Level:    5,
		Position: types.POS_STANDING,
		Hit:      50, MaxHit: 100,
		Mana: 30, MaxMana: 50,
		Move: 60, MaxMove: 80,
		PCData: &types.PCData{
			Pwd:      "pass",
			Title:    "the Tester",
			Prompt:   "<%hhp> ",
			Filename: "savetest",
			PagerLen: 24,
		},
	}
	d.Character = ch
	ch.Desc = d
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	g.closeDescriptor(d)

	// Verify player file was saved
	playerPath := persist.PlayerFilePath(tmpDir, "Savetest")
	if _, err := os.Stat(playerPath); os.IsNotExist(err) {
		t.Errorf("player file not saved at %s", playerPath)
	}

	// Verify character removed from world
	if len(w.Characters) != 0 {
		t.Errorf("world.Characters = %d, want 0", len(w.Characters))
	}

	// Suppress unused import
	_ = filepath.Base
}

// --- acceptNewConnections tests ---

func TestAcceptNewConnections(t *testing.T) {
	w := world.New("/tmp/test")
	reg := command.NewRegistry()
	incoming := make(chan *types.DescriptorData, 10)
	g := NewGameLoop(w, reg, incoming)

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()
	d := types.NewDescriptor(server)
	incoming <- d

	g.acceptNewConnections()

	if len(w.Descriptors) != 1 {
		t.Errorf("expected 1 descriptor, got %d", len(w.Descriptors))
	}
	if !d.HasOutput() {
		t.Error("new connection should have received greeting")
	}
}

func TestAcceptNewConnections_Multiple(t *testing.T) {
	w := world.New("/tmp/test")
	reg := command.NewRegistry()
	incoming := make(chan *types.DescriptorData, 10)
	g := NewGameLoop(w, reg, incoming)

	for i := 0; i < 3; i++ {
		s, c := net.Pipe()
		defer s.Close()
		defer c.Close()
		incoming <- types.NewDescriptor(s)
	}

	g.acceptNewConnections()

	if len(w.Descriptors) != 3 {
		t.Errorf("expected 3 descriptors, got %d", len(w.Descriptors))
	}
}

func TestAcceptNewConnections_Empty(t *testing.T) {
	w := world.New("/tmp/test")
	reg := command.NewRegistry()
	incoming := make(chan *types.DescriptorData, 10)
	g := NewGameLoop(w, reg, incoming)

	// No connections waiting
	g.acceptNewConnections()

	if len(w.Descriptors) != 0 {
		t.Errorf("expected 0 descriptors, got %d", len(w.Descriptors))
	}
}

// --- cleanupDescriptors tests ---

func TestCleanupDescriptors_RemovesClosed(t *testing.T) {
	tmpDir := t.TempDir()
	w := world.New(tmpDir)
	reg := command.NewRegistry()
	incoming := make(chan *types.DescriptorData, 10)
	g := NewGameLoop(w, reg, incoming)

	s1, c1 := net.Pipe()
	defer c1.Close()
	d1 := types.NewDescriptor(s1)
	d1.Connected = types.CON_PLAYING

	s2, c2 := net.Pipe()
	defer c2.Close()
	d2 := types.NewDescriptor(s2)
	d2.Connected = -1 // marked for cleanup

	w.Descriptors = append(w.Descriptors, d1, d2)

	g.cleanupDescriptors()

	if len(w.Descriptors) != 1 {
		t.Errorf("expected 1 descriptor after cleanup, got %d", len(w.Descriptors))
	}
}

func TestCleanupDescriptors_NilConn(t *testing.T) {
	tmpDir := t.TempDir()
	w := world.New(tmpDir)
	reg := command.NewRegistry()
	incoming := make(chan *types.DescriptorData, 10)
	g := NewGameLoop(w, reg, incoming)

	d := &types.DescriptorData{
		Conn:       nil,
		Connected:  types.CON_PLAYING,
		InputQueue: make(chan string, 10),
	}
	w.Descriptors = append(w.Descriptors, d)

	g.cleanupDescriptors()

	if len(w.Descriptors) != 0 {
		t.Errorf("descriptor with nil conn should be cleaned up, got %d", len(w.Descriptors))
	}
}

// --- processInput tests ---

func TestProcessInput_Playing(t *testing.T) {
	w := world.New("/tmp/test")
	reg := command.NewRegistry()

	// Register a simple test command
	called := false
	reg.Register(&command.Command{
		Name: "testcmd",
		DoFun: func(ch *types.CharData, argument string) {
			called = true
		},
		Position: types.POS_DEAD,
		Level:    0,
	})

	incoming := make(chan *types.DescriptorData, 10)
	g := NewGameLoop(w, reg, incoming)

	room := &types.RoomIndexData{Vnum: 30000, Name: "Test Room"}
	w.Rooms[30000] = room

	ch := &types.CharData{
		Name:     "Tester",
		Level:    5,
		Position: types.POS_STANDING,
		PCData:   &types.PCData{Prompt: "> "},
	}
	d := &types.DescriptorData{
		Character:  ch,
		Connected:  types.CON_PLAYING,
		InputQueue: make(chan string, 10),
	}
	ch.Desc = d
	ch.InRoom = room
	w.Descriptors = append(w.Descriptors, d)

	d.InputQueue <- "testcmd"

	g.processInput()

	if !called {
		t.Error("command should have been dispatched")
	}
}

func TestProcessInput_Editing(t *testing.T) {
	w := world.New("/tmp/test")
	reg := command.NewRegistry()
	incoming := make(chan *types.DescriptorData, 10)
	g := NewGameLoop(w, reg, incoming)

	ch := &types.CharData{
		Name:   "Builder",
		Level:  types.LEVEL_IMMORTAL,
		PCData: &types.PCData{Prompt: "> "},
	}
	d := &types.DescriptorData{
		Character:  ch,
		Connected:  types.CON_EDITING,
		InputQueue: make(chan string, 10),
	}
	ch.Desc = d
	ch.Editor = &types.EditorData{}

	w.Descriptors = append(w.Descriptors, d)

	d.InputQueue <- "some text"

	g.processInput()

	// Editor should have appended the line
	if ch.Editor.NumLines != 1 {
		t.Errorf("expected 1 line in editor, got %d", ch.Editor.NumLines)
	}
}

func TestProcessInput_PagerTakesPriority(t *testing.T) {
	w := world.New("/tmp/test")
	reg := command.NewRegistry()
	incoming := make(chan *types.DescriptorData, 10)
	g := NewGameLoop(w, reg, incoming)

	ch := &types.CharData{
		Name:   "Pager",
		Level:  5,
		PCData: &types.PCData{Prompt: "> ", PagerLen: 10},
	}
	d := &types.DescriptorData{
		Character:  ch,
		Connected:  types.CON_PLAYING,
		InputQueue: make(chan string, 10),
	}
	ch.Desc = d
	w.Descriptors = append(w.Descriptors, d)

	// Put data in pager
	d.WriteToPager("Line1\nLine2\n")

	d.InputQueue <- "q"

	g.processInput()

	// Pager cmd should be set to 'q'
	if d.GetPagerCmd() != 'q' {
		t.Errorf("pager cmd should be 'q', got %c", d.GetPagerCmd())
	}
}

func TestProcessInput_NoInput(t *testing.T) {
	w := world.New("/tmp/test")
	reg := command.NewRegistry()
	incoming := make(chan *types.DescriptorData, 10)
	g := NewGameLoop(w, reg, incoming)

	ch := &types.CharData{
		Name:   "Idle",
		Level:  5,
		PCData: &types.PCData{Prompt: "> "},
	}
	d := &types.DescriptorData{
		Character:  ch,
		Connected:  types.CON_PLAYING,
		InputQueue: make(chan string, 10),
	}
	ch.Desc = d
	w.Descriptors = append(w.Descriptors, d)

	// No input — should not crash
	g.processInput()
}

// TestProcessInput_ClosedQueue_MarksDisconnect covers the race fixed for
// Phase 5 Tier 5 G3: when the network readLoop closes InputQueue (client
// disconnected), processInput must mark the descriptor dead so the next
// cleanupDescriptors run saves the player and removes the descriptor.
// Before the fix, the zero-value from the closed channel was dispatched
// as an empty command on every pulse, leading to infinite spurious
// prompt writes and a reliance on TCP write failure to trigger cleanup —
// which raced under CPU load.
func TestProcessInput_ClosedQueue_MarksDisconnect(t *testing.T) {
	w := world.New("/tmp/test")
	reg := command.NewRegistry()
	incoming := make(chan *types.DescriptorData, 10)
	g := NewGameLoop(w, reg, incoming)

	// Count how many times any command would be interpreted. Closed-queue
	// reads must NOT dispatch commands.
	dispatched := 0
	reg.Register(&command.Command{
		Name: "noop",
		DoFun: func(ch *types.CharData, argument string) {
			dispatched++
		},
		Position: types.POS_DEAD,
		Level:    0,
	})

	room := &types.RoomIndexData{Vnum: 30000, Name: "Test Room"}
	w.Rooms[30000] = room

	ch := &types.CharData{
		Name:     "Dropped",
		Level:    5,
		Position: types.POS_STANDING,
		PCData:   &types.PCData{Prompt: "> "},
	}
	d := &types.DescriptorData{
		Character:  ch,
		Connected:  types.CON_PLAYING,
		InputQueue: make(chan string, 10),
	}
	ch.Desc = d
	ch.InRoom = room
	w.Descriptors = append(w.Descriptors, d)

	// Simulate the readLoop closing the queue because the client
	// disconnected.
	close(d.InputQueue)

	// Run processInput several times to prove the state is stable: after
	// the first detection, subsequent pulses must not re-dispatch.
	for i := 0; i < 3; i++ {
		g.processInput()
	}

	if d.Connected != -1 {
		t.Errorf("d.Connected = %d, want -1 (descriptor with closed InputQueue must be marked for cleanup)", d.Connected)
	}
	if dispatched != 0 {
		t.Errorf("closed-queue reads dispatched %d commands; want 0", dispatched)
	}
}

// --- nanny tests ---

func TestNanny_GetName_EmptyInput(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	d := types.NewDescriptor(s)
	d.Connected = types.CON_GET_NAME

	g.nanny(d, "")
	// Should prompt again
	if !d.HasOutput() {
		t.Error("empty name should re-prompt")
	}
}

func TestNanny_GetName_InvalidName(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	d := types.NewDescriptor(s)
	d.Connected = types.CON_GET_NAME

	g.nanny(d, "ab") // too short
	if !d.HasOutput() {
		t.Error("invalid name should produce error message")
	}
}

func TestNanny_GetName_NewPlayer(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	d := types.NewDescriptor(s)
	d.Connected = types.CON_GET_NAME

	g.nannyGetName(d, "Newguy")

	if d.Connected != types.CON_CONFIRM_NEW_NAME {
		t.Errorf("new player should go to CON_CONFIRM_NEW_NAME, got %d", d.Connected)
	}
	if d.User != "Newguy" {
		t.Errorf("d.User = %q, want 'Newguy'", d.User)
	}
}

// drainPipe reads and discards all data from a net.Conn in the background.
func drainPipe(c net.Conn) {
	go func() {
		buf := make([]byte, 4096)
		for {
			_, err := c.Read(buf)
			if err != nil {
				return
			}
		}
	}()
}

func TestNanny_ConfirmNewName_Yes(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	drainPipe(c)
	d := types.NewDescriptor(s)
	d.Connected = types.CON_CONFIRM_NEW_NAME
	d.User = "Testguy"

	g.nannyConfirmNewName(d, "Y")

	if d.Connected != types.CON_GET_NEW_PASSWORD {
		t.Errorf("expected CON_GET_NEW_PASSWORD, got %d", d.Connected)
	}
}

func TestNanny_ConfirmNewName_No(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	d := types.NewDescriptor(s)
	d.Connected = types.CON_CONFIRM_NEW_NAME
	d.User = "Testguy"

	g.nannyConfirmNewName(d, "N")

	if d.Connected != types.CON_GET_NAME {
		t.Errorf("expected CON_GET_NAME, got %d", d.Connected)
	}
	if d.User != "" {
		t.Errorf("d.User should be empty, got %q", d.User)
	}
}

func TestNanny_ConfirmNewName_Empty(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	d := types.NewDescriptor(s)
	d.Connected = types.CON_CONFIRM_NEW_NAME
	d.User = "Testguy"

	g.nannyConfirmNewName(d, "")

	// Should re-prompt without changing state
	if !d.HasOutput() {
		t.Error("empty confirm should re-prompt")
	}
}

func TestNanny_ConfirmNewName_InvalidInput(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	d := types.NewDescriptor(s)
	d.Connected = types.CON_CONFIRM_NEW_NAME
	d.User = "Testguy"

	g.nannyConfirmNewName(d, "maybe")

	// Should ask for Yes or No
	if !d.HasOutput() {
		t.Error("invalid input should prompt for Yes or No")
	}
}

func TestNanny_GetNewPassword_TooShort(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	drainPipe(c)
	d := types.NewDescriptor(s)
	d.Connected = types.CON_GET_NEW_PASSWORD
	d.User = "Testguy"

	g.nannyGetNewPassword(d, "abc")

	// Should stay in same state
	if d.Connected != types.CON_GET_NEW_PASSWORD {
		t.Errorf("short password should stay in CON_GET_NEW_PASSWORD, got %d", d.Connected)
	}
}

func TestNanny_GetNewPassword_ContainsTilde(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	drainPipe(c)
	d := types.NewDescriptor(s)
	d.Connected = types.CON_GET_NEW_PASSWORD
	d.User = "Testguy"

	g.nannyGetNewPassword(d, "pass~word")

	if d.Connected != types.CON_GET_NEW_PASSWORD {
		t.Errorf("tilde password should stay in CON_GET_NEW_PASSWORD, got %d", d.Connected)
	}
}

func TestNanny_GetNewPassword_Valid(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	drainPipe(c)
	d := types.NewDescriptor(s)
	d.Connected = types.CON_GET_NEW_PASSWORD
	d.User = "Testguy"

	g.nannyGetNewPassword(d, "secret123")

	if d.Connected != types.CON_CONFIRM_NEW_PASSWORD {
		t.Errorf("valid password should go to CON_CONFIRM_NEW_PASSWORD, got %d", d.Connected)
	}
	if d.Character == nil {
		t.Fatal("character should be created")
	}
	if !strings.HasPrefix(d.Character.PCData.Pwd, "$2") {
		t.Errorf("password should be bcrypt hash starting with $2, got %q", d.Character.PCData.Pwd)
	}
}

func TestNanny_ConfirmNewPassword_Mismatch(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	drainPipe(c)
	d := types.NewDescriptor(s)
	d.Connected = types.CON_CONFIRM_NEW_PASSWORD

	ch := g.createNewCharacter("Testguy")
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.MinCost)
	ch.PCData.Pwd = string(hash)
	ch.Desc = d
	d.Character = ch

	g.nannyConfirmNewPassword(d, "wrong")

	if d.Connected != types.CON_GET_NEW_PASSWORD {
		t.Errorf("mismatched password should go back to CON_GET_NEW_PASSWORD, got %d", d.Connected)
	}
}

func TestNanny_ConfirmNewPassword_NilChar(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	drainPipe(c)
	d := types.NewDescriptor(s)
	d.Connected = types.CON_CONFIRM_NEW_PASSWORD
	d.Character = nil

	g.nannyConfirmNewPassword(d, "anything")

	if d.Connected != -1 {
		t.Errorf("nil character should disconnect, got %d", d.Connected)
	}
}

func TestNanny_ConfirmNewPassword_Match(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	drainPipe(c)
	d := types.NewDescriptor(s)
	d.Connected = types.CON_CONFIRM_NEW_PASSWORD

	ch := g.createNewCharacter("Testguy")
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.MinCost)
	ch.PCData.Pwd = string(hash)
	ch.Desc = d
	d.Character = ch

	g.nannyConfirmNewPassword(d, "secret123")

	if d.Connected != types.CON_GET_NEW_SEX {
		t.Errorf("matching password should go to CON_GET_NEW_SEX, got %d", d.Connected)
	}
}

func TestNanny_GetNewSex(t *testing.T) {
	g := newTestLoop()
	// Add at least one class for the menu
	g.world.Classes = make([]*types.ClassType, 20)
	g.world.Classes[0] = &types.ClassType{WhoName: "Mage"}

	tests := []struct {
		input string
		sex   int
	}{
		{"M", types.SEX_MALE},
		{"m", types.SEX_MALE},
		{"F", types.SEX_FEMALE},
		{"f", types.SEX_FEMALE},
		{"N", types.SEX_NEUTRAL},
		{"n", types.SEX_NEUTRAL},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			s, c := net.Pipe()
			defer s.Close()
			defer c.Close()
			d := types.NewDescriptor(s)
			ch := g.createNewCharacter("Sextest")
			ch.Desc = d
			d.Character = ch

			g.nannyGetNewSex(d, tt.input)

			if ch.Sex != tt.sex {
				t.Errorf("input %q: sex = %d, want %d", tt.input, ch.Sex, tt.sex)
			}
			if d.Connected != types.CON_GET_NEW_CLASS {
				t.Errorf("input %q: should go to CON_GET_NEW_CLASS, got %d", tt.input, d.Connected)
			}
		})
	}
}

func TestNanny_GetNewSex_Invalid(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	d := types.NewDescriptor(s)
	ch := g.createNewCharacter("Sextest")
	ch.Desc = d
	d.Character = ch

	g.nannyGetNewSex(d, "X")

	// Should not change state
	if d.Connected == types.CON_GET_NEW_CLASS {
		t.Error("invalid sex should not advance state")
	}
}

func TestNanny_GetNewSex_Empty(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	d := types.NewDescriptor(s)
	ch := g.createNewCharacter("Sextest")
	ch.Desc = d
	d.Character = ch

	g.nannyGetNewSex(d, "")

	// Should re-prompt
	if !d.HasOutput() {
		t.Error("empty sex input should re-prompt")
	}
}

func TestNanny_GetOldPassword_Wrong(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	drainPipe(c)
	d := types.NewDescriptor(s)

	ch := &types.CharData{
		Name:   "Oldplayer",
		PCData: &types.PCData{Pwd: "correct"},
	}
	d.Character = ch
	ch.Desc = d

	g.nannyGetOldPassword(d, "wrong")

	if d.Connected != -1 {
		t.Errorf("wrong password should disconnect, got %d", d.Connected)
	}
	if d.Character != nil {
		t.Error("character should be nil after wrong password")
	}
}

func TestNanny_GetOldPassword_Correct(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	drainPipe(c)
	d := types.NewDescriptor(s)
	d.Host = "localhost"

	ch := &types.CharData{
		Name:   "Oldplayer",
		PCData: &types.PCData{Pwd: "correct"},
	}
	d.Character = ch
	ch.Desc = d

	g.nannyGetOldPassword(d, "correct")

	if d.Connected != types.CON_READ_MOTD {
		t.Errorf("correct password should go to CON_READ_MOTD, got %d", d.Connected)
	}
	if ch.PCData.RecentSite != "localhost" {
		t.Errorf("RecentSite = %q, want 'localhost'", ch.PCData.RecentSite)
	}
	// Legacy password should have been migrated to bcrypt
	if !strings.HasPrefix(ch.PCData.Pwd, "$2") {
		t.Errorf("legacy password should be migrated to bcrypt, got %q", ch.PCData.Pwd)
	}
}

func TestNanny_GetOldPassword_NilCharacter(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	drainPipe(c)
	d := types.NewDescriptor(s)
	d.Character = nil

	g.nannyGetOldPassword(d, "anything")

	if d.Connected != -1 {
		t.Errorf("nil character should disconnect, got %d", d.Connected)
	}
}

func TestNanny_DefaultState(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	d := types.NewDescriptor(s)
	d.Connected = 999 // unknown state

	g.nanny(d, "test")

	if d.Connected != -1 {
		t.Errorf("unknown state should disconnect, got %d", d.Connected)
	}
}

func TestNanny_GetNewClass(t *testing.T) {
	g := newTestLoop()
	g.world.Classes = make([]*types.ClassType, 20)
	g.world.Classes[0] = &types.ClassType{WhoName: "Mage"}
	g.world.Classes[1] = &types.ClassType{WhoName: "Warrior"}
	g.world.Races = make([]*types.RaceData, 20)
	g.world.Races[0] = &types.RaceData{Name: "Human"}

	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	d := types.NewDescriptor(s)
	ch := g.createNewCharacter("Classtest")
	ch.Desc = d
	d.Character = ch

	g.nannyGetNewClass(d, "Mage")

	if ch.Class != 0 {
		t.Errorf("class = %d, want 0 (Mage)", ch.Class)
	}
	if d.Connected != types.CON_GET_NEW_RACE {
		t.Errorf("should go to CON_GET_NEW_RACE, got %d", d.Connected)
	}
}

func TestNanny_GetNewClass_Invalid(t *testing.T) {
	g := newTestLoop()
	g.world.Classes = make([]*types.ClassType, 20)
	g.world.Classes[0] = &types.ClassType{WhoName: "Mage"}

	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	d := types.NewDescriptor(s)
	ch := g.createNewCharacter("Classtest")
	ch.Desc = d
	d.Character = ch

	g.nannyGetNewClass(d, "Paladin")

	// Should not advance
	if d.Connected == types.CON_GET_NEW_RACE {
		t.Error("invalid class should not advance to race selection")
	}
}

func TestNanny_GetNewClass_Empty(t *testing.T) {
	g := newTestLoop()
	g.world.Classes = make([]*types.ClassType, 20)
	g.world.Classes[0] = &types.ClassType{WhoName: "Mage"}

	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	d := types.NewDescriptor(s)
	ch := g.createNewCharacter("Classtest")
	ch.Desc = d
	d.Character = ch

	g.nannyGetNewClass(d, "")

	// Should re-show class menu
	if !d.HasOutput() {
		t.Error("empty class input should show menu")
	}
}

func TestNanny_GetNewRace(t *testing.T) {
	g := newTestLoop()
	g.world.Classes = make([]*types.ClassType, 20)
	g.world.Classes[0] = &types.ClassType{WhoName: "Mage"}
	g.world.Races = make([]*types.RaceData, 20)
	g.world.Races[0] = &types.RaceData{Name: "Human"}

	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	d := types.NewDescriptor(s)
	ch := g.createNewCharacter("Racetest")
	ch.Desc = d
	ch.Class = 0
	d.Character = ch

	g.nannyGetNewRace(d, "Human")

	if ch.Race != 0 {
		t.Errorf("race = %d, want 0 (Human)", ch.Race)
	}
	if d.Connected != types.CON_READ_MOTD {
		t.Errorf("should go to CON_READ_MOTD, got %d", d.Connected)
	}
}

func TestNanny_GetNewRace_Invalid(t *testing.T) {
	g := newTestLoop()
	g.world.Classes = make([]*types.ClassType, 20)
	g.world.Classes[0] = &types.ClassType{WhoName: "Mage"}
	g.world.Races = make([]*types.RaceData, 20)
	g.world.Races[0] = &types.RaceData{Name: "Human"}

	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	d := types.NewDescriptor(s)
	ch := g.createNewCharacter("Racetest")
	ch.Desc = d
	d.Character = ch

	g.nannyGetNewRace(d, "Vulcan")

	// Should not advance
	if d.Connected == types.CON_READ_MOTD {
		t.Error("invalid race should not advance")
	}
}

func TestNanny_GetNewRace_ClassRestriction(t *testing.T) {
	g := newTestLoop()
	g.world.Classes = make([]*types.ClassType, 20)
	g.world.Classes[0] = &types.ClassType{WhoName: "Mage"}
	g.world.Races = make([]*types.RaceData, 20)
	g.world.Races[0] = &types.RaceData{
		Name:             "Orc",
		ClassRestriction: 1, // restricts class 0 (Mage)
	}

	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	d := types.NewDescriptor(s)
	ch := g.createNewCharacter("Racetest")
	ch.Desc = d
	ch.Class = 0 // Mage
	d.Character = ch

	g.nannyGetNewRace(d, "Orc")

	// Should not advance due to class restriction
	if d.Connected == types.CON_READ_MOTD {
		t.Error("class-restricted race should not be selectable")
	}
}

func TestNanny_GetNewRace_Empty(t *testing.T) {
	g := newTestLoop()
	g.world.Classes = make([]*types.ClassType, 20)
	g.world.Classes[0] = &types.ClassType{WhoName: "Mage"}
	g.world.Races = make([]*types.RaceData, 20)
	g.world.Races[0] = &types.RaceData{Name: "Human"}

	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	d := types.NewDescriptor(s)
	ch := g.createNewCharacter("Racetest")
	ch.Desc = d
	d.Character = ch

	g.nannyGetNewRace(d, "")

	// Should re-show menu
	if !d.HasOutput() {
		t.Error("empty race input should show menu")
	}
}

// --- enterGame tests ---

func TestEnterGame(t *testing.T) {
	tmpDir := t.TempDir()
	w := world.New(tmpDir)
	reg := command.NewRegistry()
	incoming := make(chan *types.DescriptorData, 10)
	g := NewGameLoop(w, reg, incoming)

	room := &types.RoomIndexData{Vnum: types.ROOM_VNUM_TEMPLE, Name: "Temple", Description: "A holy temple.\n\r"}
	w.Rooms[types.ROOM_VNUM_TEMPLE] = room

	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	d := types.NewDescriptor(s)

	ch := g.createNewCharacter("Enterer")
	ch.Desc = d
	d.Character = ch

	g.enterGame(d)

	if d.Connected != types.CON_PLAYING {
		t.Errorf("should be CON_PLAYING, got %d", d.Connected)
	}
	if ch.InRoom != room {
		t.Error("character should be in temple room")
	}
	if ch.Position != types.POS_STANDING {
		t.Errorf("position = %d, want POS_STANDING", ch.Position)
	}
}

func TestEnterGame_NilCharacter(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	d := types.NewDescriptor(s)
	d.Character = nil

	g.enterGame(d)

	if d.Connected != -1 {
		t.Errorf("nil character should disconnect, got %d", d.Connected)
	}
}

func TestEnterGame_HomeVnum(t *testing.T) {
	tmpDir := t.TempDir()
	w := world.New(tmpDir)
	reg := command.NewRegistry()
	incoming := make(chan *types.DescriptorData, 10)
	g := NewGameLoop(w, reg, incoming)

	homeRoom := &types.RoomIndexData{Vnum: 5000, Name: "Home Room", Description: "Your home.\n\r"}
	w.Rooms[5000] = homeRoom

	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	d := types.NewDescriptor(s)

	ch := g.createNewCharacter("Homer")
	ch.HomeVnum = 5000
	ch.Desc = d
	d.Character = ch

	g.enterGame(d)

	if ch.InRoom != homeRoom {
		t.Error("character with HomeVnum should start in home room")
	}
}

// --- SavePlayer tests ---

func TestSavePlayer_NilChar(t *testing.T) {
	g := newTestLoop()
	// Should not panic
	g.SavePlayer(nil)
}

func TestSavePlayer_NPC(t *testing.T) {
	g := newTestLoop()
	ch := &types.CharData{Name: "Guard"}
	ch.Act.Set(types.ACT_IS_NPC)
	// Should not save
	g.SavePlayer(ch)
}

func TestSavePlayer_NoPCData(t *testing.T) {
	g := newTestLoop()
	ch := &types.CharData{Name: "NoPC"}
	// Should not save
	g.SavePlayer(ch)
}

func TestSavePlayer_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	w := world.New(tmpDir)
	reg := command.NewRegistry()
	incoming := make(chan *types.DescriptorData, 10)
	g := NewGameLoop(w, reg, incoming)

	ch := &types.CharData{
		Name:   "Atomica",
		Level:  1,
		PCData: &types.PCData{Pwd: "test"},
	}

	g.SavePlayer(ch)

	// The final file should exist
	playerPath := persist.PlayerFilePath(tmpDir, ch.Name)
	if _, err := os.Stat(playerPath); os.IsNotExist(err) {
		t.Errorf("player file %s does not exist after save", playerPath)
	}

	// The temp file should NOT exist
	tmpPath := playerPath + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Errorf("temp file %s still exists after save", tmpPath)
	}
}

// --- pulse tests ---

func TestPulse(t *testing.T) {
	tmpDir := t.TempDir()
	w := world.New(tmpDir)
	reg := command.NewRegistry()
	incoming := make(chan *types.DescriptorData, 10)
	g := NewGameLoop(w, reg, incoming)

	// Run a single pulse — should not crash
	g.pulse()
}

// --- flushOutput with pager ---

func TestFlushOutput_WithPager(t *testing.T) {
	w := world.New("/tmp/test")
	reg := command.NewRegistry()
	incoming := make(chan *types.DescriptorData, 10)
	g := NewGameLoop(w, reg, incoming)

	s, c := net.Pipe()
	defer c.Close()
	d := types.NewDescriptor(s)

	ch := &types.CharData{
		Name:   "PagerFlush",
		Level:  5,
		PCData: &types.PCData{PagerLen: 10},
	}
	d.Character = ch
	ch.Desc = d

	// Write enough data to trigger paging
	var sb strings.Builder
	for i := 0; i < 5; i++ {
		sb.WriteString("Short line\n\r")
	}
	d.WriteToPager(sb.String())

	w.Descriptors = append(w.Descriptors, d)

	// Read output in background to prevent blocking
	go func() {
		buf := make([]byte, 4096)
		for {
			_, err := c.Read(buf)
			if err != nil {
				return
			}
		}
	}()

	g.flushOutput()
	s.Close()
}

// --- Password hashing and brute force tests ---

func TestPasswordHash_Bcrypt(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	drainPipe(c)
	d := types.NewDescriptor(s)
	d.Connected = types.CON_GET_NEW_PASSWORD
	d.User = "Hashtest"

	g.nannyGetNewPassword(d, "secret123")

	if d.Character == nil {
		t.Fatal("character should be created")
	}
	pwd := d.Character.PCData.Pwd
	if !strings.HasPrefix(pwd, "$2") {
		t.Fatalf("stored password should be bcrypt hash (start with $2), got %q", pwd)
	}
	// Verify the hash actually matches the input
	if err := bcrypt.CompareHashAndPassword([]byte(pwd), []byte("secret123")); err != nil {
		t.Errorf("bcrypt hash should match original password: %v", err)
	}
	// Verify wrong password does not match
	if err := bcrypt.CompareHashAndPassword([]byte(pwd), []byte("wrongpass")); err == nil {
		t.Error("bcrypt hash should NOT match wrong password")
	}
}

func TestPasswordHash_LegacyMigration(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	drainPipe(c)
	d := types.NewDescriptor(s)
	d.Host = "localhost"

	// Set up a character with a legacy plaintext password
	ch := &types.CharData{
		Name:   "Legacy",
		PCData: &types.PCData{Pwd: "oldplain"},
	}
	d.Character = ch
	ch.Desc = d

	// Login with correct plaintext password
	g.nannyGetOldPassword(d, "oldplain")

	if d.Connected != types.CON_READ_MOTD {
		t.Fatalf("correct legacy password should succeed, got state %d", d.Connected)
	}
	// Password should now be migrated to bcrypt
	if !strings.HasPrefix(ch.PCData.Pwd, "$2") {
		t.Errorf("legacy password should be migrated to bcrypt, got %q", ch.PCData.Pwd)
	}
	// Verify the new hash matches the original password
	if err := bcrypt.CompareHashAndPassword([]byte(ch.PCData.Pwd), []byte("oldplain")); err != nil {
		t.Errorf("migrated hash should match original password: %v", err)
	}
}

func TestPasswordHash_BcryptLogin(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	drainPipe(c)
	d := types.NewDescriptor(s)
	d.Host = "localhost"

	// Set up a character with a bcrypt-hashed password
	hash, _ := bcrypt.GenerateFromPassword([]byte("bcryptpwd"), bcrypt.MinCost)
	ch := &types.CharData{
		Name:   "Modern",
		PCData: &types.PCData{Pwd: string(hash)},
	}
	d.Character = ch
	ch.Desc = d

	g.nannyGetOldPassword(d, "bcryptpwd")

	if d.Connected != types.CON_READ_MOTD {
		t.Errorf("correct bcrypt password should succeed, got state %d", d.Connected)
	}
}

func TestBruteForceProtection(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	drainPipe(c)

	hash, _ := bcrypt.GenerateFromPassword([]byte("realpassword"), bcrypt.MinCost)

	// Attempt 1: wrong password
	d := types.NewDescriptor(s)
	d.Host = "localhost"
	ch1 := &types.CharData{Name: "Bruteforce", PCData: &types.PCData{Pwd: string(hash)}}
	d.Character = ch1
	ch1.Desc = d

	g.nannyGetOldPassword(d, "wrong1")
	if d.FailedAttempts != 1 {
		t.Errorf("after 1st failed attempt, FailedAttempts = %d, want 1", d.FailedAttempts)
	}
	if d.Connected != -1 {
		t.Errorf("wrong password should disconnect, got %d", d.Connected)
	}

	// Attempt 2: simulate reconnect on same descriptor (reset state but keep FailedAttempts)
	d.Connected = types.CON_GET_OLD_PASSWORD
	ch2 := &types.CharData{Name: "Bruteforce", PCData: &types.PCData{Pwd: string(hash)}}
	d.Character = ch2
	ch2.Desc = d

	g.nannyGetOldPassword(d, "wrong2")
	if d.FailedAttempts != 2 {
		t.Errorf("after 2nd failed attempt, FailedAttempts = %d, want 2", d.FailedAttempts)
	}

	// Attempt 3: should disconnect with brute force message
	d.Connected = types.CON_GET_OLD_PASSWORD
	ch3 := &types.CharData{Name: "Bruteforce", PCData: &types.PCData{Pwd: string(hash)}}
	d.Character = ch3
	ch3.Desc = d

	g.nannyGetOldPassword(d, "wrong3")
	if d.FailedAttempts != 3 {
		t.Errorf("after 3rd failed attempt, FailedAttempts = %d, want 3", d.FailedAttempts)
	}
	if d.Connected != -1 {
		t.Errorf("3 failed attempts should disconnect, got %d", d.Connected)
	}
}

// TestNanny_GetOldPassword_LockoutMessage covers audit finding U1 (2026-04-17):
// when a player hits the password-failure limit, the server must emit a
// user-visible explanation before marking the descriptor dead. Without the
// message a legitimate mistyper gets silently dropped and cannot tell the
// difference between a server fault and a lockout.
func TestNanny_GetOldPassword_LockoutMessage(t *testing.T) {
	g := newTestLoop()
	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()

	// Collect everything the server writes to the client side of the pipe.
	got := make(chan string, 1)
	go func() {
		var sb strings.Builder
		buf := make([]byte, 4096)
		for {
			n, err := c.Read(buf)
			if n > 0 {
				sb.Write(buf[:n])
			}
			if err != nil {
				got <- sb.String()
				return
			}
		}
	}()

	hash, _ := bcrypt.GenerateFromPassword([]byte("realpassword"), bcrypt.MinCost)
	d := types.NewDescriptor(s)
	d.Host = "localhost"
	d.FailedAttempts = 2 // one more failure trips the lockout
	ch := &types.CharData{Name: "Lockout", PCData: &types.PCData{Pwd: string(hash)}}
	d.Character = ch
	ch.Desc = d

	g.nannyGetOldPassword(d, "stillwrong")

	// Flush pending output before closing, mirroring the pulse sequence
	// (processInput -> flushOutput -> cleanupDescriptors).
	if err := d.FlushOutput(); err != nil {
		t.Fatalf("FlushOutput error: %v", err)
	}
	s.Close()

	out := <-got
	if !strings.Contains(out, "Too many password failures") {
		t.Errorf("expected lockout message containing %q, got %q",
			"Too many password failures", out)
	}
	if d.Connected != -1 {
		t.Errorf("d.Connected = %d, want -1 (marked for cleanup after lockout)", d.Connected)
	}
	if d.FailedAttempts != 3 {
		t.Errorf("FailedAttempts = %d, want 3 (lockout threshold unchanged)", d.FailedAttempts)
	}
}

// --- Adversary-follow-up coverage tests ---

func TestNanny_GetNewClass_Banned(t *testing.T) {
	g := newTestLoop()
	g.world.Classes = make([]*types.ClassType, 20)
	g.world.Classes[0] = &types.ClassType{WhoName: "Mage"}
	g.world.Races = make([]*types.RaceData, 20)
	g.world.Races[0] = &types.RaceData{Name: "Human"}
	g.world.Bans = append(g.world.Bans, &types.BanData{
		Name: "mage", Type: types.BAN_CLASS, Level: 10,
	})

	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	d := types.NewDescriptor(s)
	ch := g.createNewCharacter("Bantest")
	ch.Desc = d
	d.Character = ch

	g.nannyGetNewClass(d, "Mage")

	if d.Connected == types.CON_GET_NEW_RACE {
		t.Error("banned class should not advance to race selection")
	}
	if ch.Class == 0 {
		// Class was set to 0 — verify by checking it wasn't kept.
		// The test char defaults to Class=0 already, so we check the state
		// didn't advance instead.
	}
}

func TestNanny_GetNewRace_Banned(t *testing.T) {
	g := newTestLoop()
	g.world.Classes = make([]*types.ClassType, 20)
	g.world.Classes[0] = &types.ClassType{WhoName: "Mage"}
	g.world.Races = make([]*types.RaceData, 20)
	g.world.Races[0] = &types.RaceData{Name: "Troll"}
	g.world.Bans = append(g.world.Bans, &types.BanData{
		Name: "troll", Type: types.BAN_RACE, Level: 0, // 0 = block everyone
	})

	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	d := types.NewDescriptor(s)
	ch := g.createNewCharacter("Bantest")
	ch.Desc = d
	d.Character = ch
	d.Connected = types.CON_GET_NEW_RACE

	g.nannyGetNewRace(d, "Troll")

	if d.Connected == types.CON_READ_MOTD {
		t.Error("banned race should not advance to MOTD")
	}
}

func TestEnterGame_MailGreetingShown(t *testing.T) {
	g := newTestLoop()
	g.world.Rooms[types.ROOM_VNUM_TEMPLE] = &types.RoomIndexData{
		Vnum: types.ROOM_VNUM_TEMPLE, Name: "Temple",
	}
	board := &types.BoardData{
		Notes: []*types.NoteData{
			{Sender: "Someone", Subject: "Hello", ToList: "Mailer"},
		},
	}
	g.world.Boards = append(g.world.Boards, board)

	origRef := act.WorldRef
	act.WorldRef = g.world
	defer func() { act.WorldRef = origRef }()

	s, c := net.Pipe()
	defer s.Close()
	defer c.Close()
	d := types.NewDescriptor(s)
	ch := &types.CharData{
		Name: "Mailer", Position: types.POS_STANDING,
		PCData: &types.PCData{},
	}
	ch.Desc = d
	d.Character = ch

	got := make(chan string, 1)
	go func() {
		buf := make([]byte, 8192)
		n, _ := c.Read(buf)
		got <- string(buf[:n])
	}()

	g.enterGame(d)
	_ = d.FlushOutput()

	out := <-got
	if !strings.Contains(out, "1 note") {
		t.Errorf("expected mailbox greeting; got %q", out)
	}
}

// TestNewGameLoop_PulseSaveInitialized pins D1 — pulseSave field must
// be initialized to types.PULSE_SAVE in NewGameLoop. Pre-fix the field
// was zero-valued, so the autosave decrement-to-zero fired on the very
// first pulse after boot instead of after PULSE_SAVE ticks.
//
// Plan: plan-phase6-quickwins-blank-pcrename.md §G1 / §A1.
func TestNewGameLoop_PulseSaveInitialized(t *testing.T) {
	g := newTestLoop()
	if g.pulseSave != types.PULSE_SAVE {
		t.Errorf("pulseSave = %d; want %d (types.PULSE_SAVE)", g.pulseSave, types.PULSE_SAVE)
	}
}
