package testclient

import (
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// These E2E tests cover plan-phase6-olc-medit.md §G15 — the full
// `medit <vnum>` round-trip from nanny-driven login, through
// CON_MEDIT menu entry, through a simple-field set, and back out.
// They mirror the oedit precedent at oedit_test.go.
//
// The shared testdata area (`../../cmd/smaug/testdata`) ships rooms
// but no mob prototypes at predictable vnums, so every scenario seeds
// a MobIndexData into the live world via h.Query before driving the
// client. That keeps the scenarios deterministic.

// seedMedit6000 injects a fresh NPC mob prototype at vnum 6000 into
// the live world. Callers should h.Query this before typing
// `medit 6000`.
func seedMedit6000(h *Harness) {
	h.Query(func(w *world.World) {
		idx := &types.MobIndexData{
			Vnum:        6000,
			PlayerName:  "rat",
			ShortDescr:  "a tiny rat",
			LongDescr:   "A tiny rat scurries here.",
			Description: "It is small and gray.",
			Level:       2,
			Position:    types.POS_STANDING,
			DefPosition: types.POS_STANDING,
		}
		idx.Act.Set(types.ACT_IS_NPC)
		w.MobIndex[6000] = idx
	})
}

// TestTestclient_MeditMenuEntryAndQuit exercises plan §G15 scenario 1:
// an immortal types `medit 6000`, sees the NPC main menu (pins each
// expected field label), types `Q`, and is returned to CON_PLAYING
// with the "Exiting editor." confirmation. Pins the full loop-dispatch
// path: act.DoMedit → act.MeditDispMenuFunc seam → game.MeditDispMenu
// emits NPC-menu text → client reads menu → client sends Q →
// game.meditParse case MEDIT_NPC_MAIN_MENU handles Q → cleanupOlc →
// CON_PLAYING.
func TestTestclient_MeditMenuEntryAndQuit(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))
	seedMedit6000(h)

	c := h.NewCharacter(t, CharSpec{
		Name:     "Meditor",
		Password: defaultQuickLoginPassword,
		Trust:    types.LEVEL_IMMORTAL,
	})
	defer c.Close()

	c.Send("medit 6000")
	out := c.ReadUntil("Enter choice", 3*time.Second)
	for _, want := range []string{
		"Mob Number",
		"Sex",
		"Name",
		"Shortdesc",
		"Longdesc",
		"Class",
		"Race",
		"Level",
		"Quit",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("NPC menu missing %q; got:\n%s", want, out)
		}
	}

	c.Send("Q")
	ack := c.ReadUntil("Exiting editor", 3*time.Second)
	if !strings.Contains(ack, "Exiting editor") {
		t.Errorf("expected 'Exiting editor' on Q; got %q", ack)
	}

	// Post-Quit: descriptor must be back in CON_PLAYING.
	h.Query(func(w *world.World) {
		for _, d := range w.Descriptors {
			if d.Character != nil && strings.EqualFold(d.Character.Name, "Meditor") {
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

// TestTestclient_MeditSetField drives plan §G15 scenario 2: enter the
// menu, navigate to digit `2` (NAME) per the NPC main-menu digit
// table, type a new name, verify the menu redisplays with the mutation
// landed and the prototype's PlayerName updated through the
// ACT_PROTOTYPE mirror.
func TestTestclient_MeditSetField(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))
	seedMedit6000(h)

	c := h.NewCharacter(t, CharSpec{
		Name:     "Mnamer",
		Password: defaultQuickLoginPassword,
		Trust:    types.LEVEL_IMMORTAL,
	})
	defer c.Close()

	c.Send("medit 6000")
	_ = c.ReadUntil("Enter choice", 3*time.Second)

	// NPC digit 2 → MEDIT_NAME (per plan §232 NPC menu table).
	c.Send("2")
	_ = c.ReadUntil("Enter name", 3*time.Second)

	c.Send("renamedrat")
	// The menu redisplay carries the new name.
	out := c.ReadUntil("Enter choice", 3*time.Second)
	if !strings.Contains(out, "renamedrat") {
		t.Errorf("new name not in menu redisplay; got %q", out)
	}

	// Verify on-world via Query — the prototype mirror must have
	// fired (ACT_PROTOTYPE is set on the wrapper by DoMedit so name
	// arms dual-write through victim.IndexData).
	h.Query(func(w *world.World) {
		idx, ok := w.MobIndex[6000]
		if !ok || idx == nil {
			t.Fatalf("mob 6000 missing")
		}
		if idx.PlayerName != "renamedrat" {
			t.Errorf("idx.PlayerName = %q, want renamedrat", idx.PlayerName)
		}
	})

	c.Send("Q")
	_ = c.ReadUntil("Exiting editor", 3*time.Second)
}

// TestTestclient_MeditInvalidDigitRedisplays pins the equivalent of
// the oedit redisplay-on-bad-input contract: garbage input at the
// main menu must re-render the menu and keep Connected at CON_MEDIT
// (not drop out to CON_PLAYING).
func TestTestclient_MeditInvalidDigitRedisplays(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))
	seedMedit6000(h)

	c := h.NewCharacter(t, CharSpec{
		Name:     "Mbadinput",
		Password: defaultQuickLoginPassword,
		Trust:    types.LEVEL_IMMORTAL,
	})
	defer c.Close()

	c.Send("medit 6000")
	_ = c.ReadUntil("Enter choice", 3*time.Second)

	// `~` is outside any documented digit; the dispatcher's default
	// arm re-renders the NPC main menu per medit_parse.go:371-374.
	c.Send("~")
	out := c.ReadUntil("Enter choice", 3*time.Second)
	if !strings.Contains(out, "Mob Number") {
		t.Errorf("expected NPC menu re-render after bad input; got %q", out)
	}

	// Still in CON_MEDIT; Q should exit cleanly.
	c.Send("Q")
	exit := c.ReadUntil("Exiting editor", 3*time.Second)
	if !strings.Contains(exit, "Exiting editor") {
		t.Errorf("Q did not produce exit message after invalid input; got %q", exit)
	}
}
