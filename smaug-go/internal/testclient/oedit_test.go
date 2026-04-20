package testclient

import (
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/act"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// These E2E tests cover plan-phase6-olc-oedit.md §G12 — the full
// `oedit <vnum>` round-trip from nanny-driven login, through CON_OEDIT
// menu entry, to /s-backed text editor round-trip and back out. They
// mirror the redit precedent at redit_test.go.
//
// The shared testdata area (`../../cmd/smaug/testdata`) ships rooms but
// no objects, so every scenario seeds an ObjIndexData into the live
// world via h.Query before driving the client. That keeps the scenarios
// deterministic and independent of area-file shape.

// seedOedit4000 injects a fresh prototype at vnum 4000 into the live
// world. Callers should h.Query this before typing `oedit 4000`.
func seedOedit4000(h *Harness) {
	h.Query(func(w *world.World) {
		w.ObjIndex[4000] = &types.ObjIndexData{
			Vnum:        4000,
			Name:        "proto",
			ShortDescr:  "a prototype widget",
			Description: "A prototype widget lies here.",
			ItemType:    types.ITEM_TRASH,
			Level:       1,
			Weight:      1,
		}
	})
}

// TestTestclient_OeditMenuEntryAndQuit exercises plan §G12 scenario 1:
// an immortal types `oedit 4000`, sees the main menu (pins each
// expected field label), types `Q`, and is returned to CON_PLAYING
// with the "Exiting editor." confirmation. Pins the full loop-dispatch
// path: act.DoOedit → act.OeditDispMenuFunc seam → game.OeditDispMenu
// emits menu text → client reads menu → client sends Q →
// game.oeditParse case OEDIT_MAIN_MENU handles Q → cleanupOlc →
// CON_PLAYING.
func TestTestclient_OeditMenuEntryAndQuit(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))
	seedOedit4000(h)

	c := h.NewCharacter(t, CharSpec{
		Name:     "Oeditor",
		Password: defaultQuickLoginPassword,
		Trust:    types.LEVEL_IMMORTAL,
	})
	defer c.Close()

	c.Send("oedit 4000")
	// Menu redisplays every field + menu options. Match on stable
	// literals. No trailing "> " prompt sentinel — read until the
	// "Enter choice" terminator (same contract as redit).
	out := c.ReadUntil("Enter choice", 3*time.Second)
	for _, want := range []string{
		"Object number",
		"Name",
		"Short desc",
		"Long desc",
		"Type",
		"Extra flags",
		"Wear flags",
		"Weight",
		"Cost",
		"Values",
		"Extra descriptions menu",
		"Quit",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("menu missing %q; got:\n%s", want, out)
		}
	}

	c.Send("Q")
	ack := c.ReadUntil("Exiting editor", 3*time.Second)
	if !strings.Contains(ack, "Exiting editor") {
		t.Errorf("expected 'Exiting editor' on Q; got %q", ack)
	}

	// Post-Quit: descriptor must be back in CON_PLAYING. Probe via
	// Query so the assertion runs on the loop goroutine.
	h.Query(func(w *world.World) {
		for _, d := range w.Descriptors {
			if d.Character != nil && strings.EqualFold(d.Character.Name, "Oeditor") {
				if d.Connected != int(types.CON_PLAYING) {
					t.Errorf("post-Q Connected = %d, want CON_PLAYING", d.Connected)
				}
				if d.Olc != nil {
					t.Errorf("post-Q Olc must be nil; got %+v", d.Olc)
				}
			}
		}
	})
}

// TestTestclient_OeditMenuSetName drives plan §G12 scenario 2:
// immortal enters the menu, types `1` (namelist), types a new name,
// and the menu redisplays with the mutated name. Pins A7.
func TestTestclient_OeditMenuSetName(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))
	seedOedit4000(h)

	c := h.NewCharacter(t, CharSpec{
		Name:     "Onamer",
		Password: defaultQuickLoginPassword,
		Trust:    types.LEVEL_IMMORTAL,
	})
	defer c.Close()

	c.Send("oedit 4000")
	_ = c.ReadUntil("Enter choice", 3*time.Second)

	c.Send("1")
	_ = c.ReadUntil("Enter namelist", 3*time.Second)

	c.Send("renamedproto")
	// Menu redisplay; match on the new name.
	out := c.ReadUntil("Enter choice", 3*time.Second)
	if !strings.Contains(out, "renamedproto") {
		t.Errorf("new name not in menu redisplay; got %q", out)
	}

	// Verify on-world via Query — menu redisplay might contain the
	// old text via race; confirm the source of truth.
	h.Query(func(w *world.World) {
		idx, ok := w.ObjIndex[4000]
		if !ok || idx == nil {
			t.Fatalf("object 4000 missing")
		}
		if idx.Name != "renamedproto" {
			t.Errorf("idx.Name = %q, want renamedproto", idx.Name)
		}
	})

	c.Send("Q")
	_ = c.ReadUntil("Exiting editor", 3*time.Second)
}

// TestTestclient_OeditInvalidChoiceRedisplays pins A22: garbage input
// at the main menu must re-render the menu and keep Connected at
// CON_OEDIT (not drop out to CON_PLAYING).
func TestTestclient_OeditInvalidChoiceRedisplays(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))
	seedOedit4000(h)

	c := h.NewCharacter(t, CharSpec{
		Name:     "Obador",
		Password: defaultQuickLoginPassword,
		Trust:    types.LEVEL_IMMORTAL,
	})
	defer c.Close()

	c.Send("oedit 4000")
	_ = c.ReadUntil("Enter choice", 3*time.Second)

	c.Send("zzz")
	out := c.ReadUntil("Enter choice", 3*time.Second)
	if !strings.Contains(out, "Object number") {
		t.Errorf("expected menu re-render after bad input; got %q", out)
	}

	// Still in CON_OEDIT; Q should exit cleanly.
	c.Send("Q")
	exit := c.ReadUntil("Exiting editor", 3*time.Second)
	if !strings.Contains(exit, "Exiting editor") {
		t.Errorf("Q did not produce exit message after invalid input; got %q", exit)
	}
}

// TestTestclient_OeditTypeSubmenu pins A10: `5` enters the item-type
// sub-menu; numeric input (ITEM_WEAPON=5) mutates idx.ItemType and
// returns to main. Uses the numeric path (Word-path depends on the
// oTypeName table and `getOtype` helper — both exercised at unit
// level in game/oedit_parse_test.go).
func TestTestclient_OeditTypeSubmenu(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))
	seedOedit4000(h)

	c := h.NewCharacter(t, CharSpec{
		Name:     "Otyper",
		Password: defaultQuickLoginPassword,
		Trust:    types.LEVEL_IMMORTAL,
	})
	defer c.Close()

	c.Send("oedit 4000")
	_ = c.ReadUntil("Enter choice", 3*time.Second)

	c.Send("5")
	_ = c.ReadUntil("Enter type", 3*time.Second)

	// ITEM_WEAPON = 5 in the SMAUG iota (verified via types.enums.go).
	c.Send("5")
	out := c.ReadUntil("Enter choice", 3*time.Second)
	if !strings.Contains(out, "weapon") {
		t.Errorf("expected 'weapon' in menu redisplay; got %q", out)
	}

	h.Query(func(w *world.World) {
		idx := w.ObjIndex[4000]
		if idx == nil {
			t.Fatal("obj 4000 missing")
		}
		if idx.ItemType != types.ITEM_WEAPON {
			t.Errorf("idx.ItemType = %d, want ITEM_WEAPON (%d)", idx.ItemType, types.ITEM_WEAPON)
		}
	})

	c.Send("Q")
	_ = c.ReadUntil("Exiting editor", 3*time.Second)
}

// TestTestclient_OeditLongdescRoundtrip pins A8: `3` enters the line
// editor; `/s` transitions through CON_EDITING back into CON_OEDIT
// (NOT CON_PLAYING), writes the typed text into idx.Description, and
// redisplays the main menu. This exercises the EditorSave trampoline
// at oedit_parse.go case "3" — mutation gate #11 (swap Connected to
// CON_REDIT/CON_PLAYING in the closure) is caught here.
func TestTestclient_OeditLongdescRoundtrip(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))
	seedOedit4000(h)

	c := h.NewCharacter(t, CharSpec{
		Name:     "Olonger",
		Password: defaultQuickLoginPassword,
		Trust:    types.LEVEL_IMMORTAL,
	})
	defer c.Close()

	c.Send("oedit 4000")
	_ = c.ReadUntil("Enter choice", 3*time.Second)

	c.Send("3")
	// StartEditing prints the "Enter long desc:" prompt plus editor
	// banner. Read past whichever token lands first.
	_ = c.ReadUntil("Enter long desc", 3*time.Second)

	c.Send("A freshly edited prototype description.")
	c.Send("/s")

	// Post-save: CON_OEDIT trampoline re-renders the main menu. The
	// new description must appear.
	out := c.ReadUntil("Enter choice", 5*time.Second)
	if !strings.Contains(out, "freshly edited prototype") {
		t.Errorf("new long desc not in menu redisplay; got %q", out)
	}

	h.Query(func(w *world.World) {
		idx := w.ObjIndex[4000]
		if idx == nil {
			t.Fatal("obj 4000 missing")
		}
		if !strings.Contains(idx.Description, "freshly edited prototype") {
			t.Errorf("idx.Description = %q; missing 'freshly edited prototype'", idx.Description)
		}
		// Pin: Connected must be CON_OEDIT post-/s (mutation gate
		// #11 target). The menu re-render in the test above already
		// implies CON_OEDIT (otherwise the input would have been
		// treated as a command), but hard-assert here for clarity.
		for _, d := range w.Descriptors {
			if d.Character != nil && strings.EqualFold(d.Character.Name, "Olonger") {
				if d.Connected != int(types.CON_OEDIT) {
					t.Errorf("post-/s Connected = %d, want CON_OEDIT", d.Connected)
				}
			}
		}
	})

	c.Send("Q")
	_ = c.ReadUntil("Exiting editor", 3*time.Second)
}

// TestTestclient_OeditAreaSaveGuardsPathTraversal proves the G11
// path-containment guard at DoSaveArea plumbs correctly end-to-end
// (Wave 1 unit tests in act/olc_test.go already pin the core logic).
// We mutate `area.Filename` via harness Query to inject a traversal
// attempt, then drive `asave` via the logged-in immortal — the server
// must reject without creating any file outside `<DataDir>/area/`.
func TestTestclient_OeditAreaSaveGuardsPathTraversal(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))

	c := h.NewCharacter(t, CharSpec{
		Name:     "Asaver",
		Password: defaultQuickLoginPassword,
		Trust:    types.LEVEL_IMMORTAL,
	})
	defer c.Close()

	// Inject the traversal target into the area the character's
	// current room belongs to. The shipped testdata area is "Test Area"
	// at `test_boot.are`; we clobber its Filename in-place with a
	// malicious path, then drive asave.
	//
	// Verified invariant: DoSaveArea consults ch.InRoom.Area.Filename.
	var originalFilename string
	h.Query(func(w *world.World) {
		for _, d := range w.Descriptors {
			if d.Character != nil && strings.EqualFold(d.Character.Name, "Asaver") {
				if d.Character.InRoom == nil || d.Character.InRoom.Area == nil {
					t.Fatal("Asaver not in a room with an Area")
				}
				originalFilename = d.Character.InRoom.Area.Filename
				d.Character.InRoom.Area.Filename = "../../etc/passwd"
			}
		}
	})

	c.Send("savearea")
	out := c.ReadUntil("Invalid area filename", 3*time.Second)
	if !strings.Contains(out, "Invalid area filename") {
		t.Errorf("expected rejection, got %q", out)
	}

	// Restore for any subsequent test in the same package.
	h.Query(func(w *world.World) {
		for _, d := range w.Descriptors {
			if d.Character != nil && strings.EqualFold(d.Character.Name, "Asaver") {
				if d.Character.InRoom != nil && d.Character.InRoom.Area != nil {
					d.Character.InRoom.Area.Filename = originalFilename
				}
			}
		}
	})

	// Keep linker happy if `act` is otherwise unused.
	_ = act.WorldRef
}
