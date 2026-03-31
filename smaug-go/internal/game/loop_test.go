package game

import (
	"net"
	"os"
	"path/filepath"
	"testing"

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
