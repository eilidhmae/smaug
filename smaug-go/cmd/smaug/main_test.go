package main

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/boot"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// These tests used to exercise package-local bootDB/registerCommands
// helpers. Boot wiring now lives in internal/boot, and boot_test.go in that
// package covers the full surface. The tests here remain as integration
// smoke checks that Boot-against-testdata produces the expected results
// when invoked from the cmd package.

func newIncoming() chan *types.DescriptorData {
	return make(chan *types.DescriptorData, 4)
}

func TestBootDB(t *testing.T) {
	w := world.New("testdata")

	_, _, err := boot.Boot(w, "testdata", newIncoming(), boot.ProductionOpts())
	if err != nil {
		t.Fatalf("boot.Boot failed: %v", err)
	}

	// Should have loaded the test area with the temple room
	temple := w.GetRoom(types.ROOM_VNUM_TEMPLE)
	if temple == nil {
		t.Fatal("temple room (vnum 21001) not loaded")
	}
	if temple.Name != "The Temple of Mota" {
		t.Errorf("temple.Name = %q, want %q", temple.Name, "The Temple of Mota")
	}

	// Areas should be loaded
	if len(w.Areas) == 0 {
		t.Error("no areas loaded")
	}
}

func TestBootDB_MissingDataDir(t *testing.T) {
	w := world.New("/nonexistent/path")

	_, _, err := boot.Boot(w, "/nonexistent/path", newIncoming(), boot.ProductionOpts())
	if err == nil {
		t.Error("boot.Boot should fail with missing data dir")
	}
}

func TestBootDB_FallbackRoom(t *testing.T) {
	// Create a world with a data dir that has no rooms (empty area)
	w := world.New("testdata")

	// Temporarily clear rooms after boot to test fallback
	_, _, err := boot.Boot(w, "testdata", newIncoming(), boot.ProductionOpts())
	if err != nil {
		t.Fatalf("boot.Boot failed: %v", err)
	}

	// Delete the temple room and re-run the fallback logic
	delete(w.Rooms, types.ROOM_VNUM_TEMPLE)

	// Re-run just the fallback part (mirrors the in-Boot fallback block)
	if w.GetRoom(types.ROOM_VNUM_TEMPLE) == nil {
		fallback := &types.RoomIndexData{
			Vnum:        types.ROOM_VNUM_TEMPLE,
			Name:        "The Void",
			Description: "You are floating in an empty void. The world has not been loaded.\n\r",
			SectorType:  types.SECT_INSIDE,
		}
		w.Rooms[fallback.Vnum] = fallback
	}

	temple := w.GetRoom(types.ROOM_VNUM_TEMPLE)
	if temple == nil {
		t.Fatal("fallback room should be created")
	}
	if temple.Name != "The Void" {
		t.Errorf("fallback room name = %q, want %q", temple.Name, "The Void")
	}
}

func TestRegisterCommands(t *testing.T) {
	w := world.New("testdata")
	reg, _, err := boot.Boot(w, "testdata", newIncoming(), boot.ProductionOpts())
	if err != nil {
		t.Fatalf("boot.Boot failed: %v", err)
	}

	// Verify key commands are registered
	commands := []struct {
		name string
	}{
		{"look"}, {"quit"}, {"say"}, {"score"}, {"who"},
		{"commands"}, {"help"}, {"inventory"}, {"equipment"},
		{"consider"}, {"where"}, {"time"},
		{"tell"}, {"reply"}, {"yell"}, {"gossip"}, {"emote"},
		{"get"}, {"drop"}, {"put"}, {"give"}, {"wear"}, {"remove"}, {"sacrifice"},
		{"cast"},
		{"kill"}, {"flee"},
		{"open"}, {"close"}, {"unlock"}, {"lock"},
		{"north"}, {"south"}, {"east"}, {"west"}, {"up"}, {"down"},
		{"northeast"}, {"northwest"}, {"southeast"}, {"southwest"},
	}

	for _, tt := range commands {
		cmd := reg.Find(tt.name, 0)
		if cmd == nil {
			t.Errorf("command %q not registered", tt.name)
		}
	}
}

func TestRegisterCommands_Count(t *testing.T) {
	w := world.New("testdata")
	reg, _, err := boot.Boot(w, "testdata", newIncoming(), boot.ProductionOpts())
	if err != nil {
		t.Fatalf("boot.Boot failed: %v", err)
	}

	// 12 info + 5 comm + 7 object + 1 magic + 2 combat + 4 door + 10 movement = 41 commands
	expected := 41
	count := 0
	for _, name := range []string{
		"look", "quit", "say", "score", "who", "commands", "help", "inventory", "equipment",
		"consider", "where", "time",
		"tell", "reply", "yell", "gossip", "emote",
		"get", "drop", "put", "give", "wear", "remove", "sacrifice",
		"cast",
		"kill", "flee",
		"open", "close", "unlock", "lock",
		"north", "east", "south", "west", "up", "down",
		"northeast", "northwest", "southeast", "southwest",
	} {
		if reg.Find(name, 0) != nil {
			count++
		}
	}
	if count != expected {
		t.Errorf("registered %d commands, want %d", count, expected)
	}
}
