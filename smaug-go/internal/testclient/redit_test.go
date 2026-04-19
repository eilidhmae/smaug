package testclient

import (
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/types"
)

// TestTestclient_ReditMenuEntryAndQuit exercises plan §G12 scenario 1:
// an immortal types `redit` with no args, sees the main menu, types Q,
// and is returned to CON_PLAYING. Pins the full loop-dispatch path:
// act.DoRedit → act.ReditDispMenuFunc seam → game.ReditDispMenu emits
// menu text → client reads menu → client sends Q → game.reditParse
// case REDIT_MAIN_MENU handles Q → cleanupOlc → CON_PLAYING.
func TestTestclient_ReditMenuEntryAndQuit(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))

	c := h.NewCharacter(t, CharSpec{
		Name:     "Reditor",
		Password: defaultQuickLoginPassword,
		Trust:    types.LEVEL_IMMORTAL,
	})
	defer c.Close()

	c.Send("redit")
	// Menu redisplays every field + menu options. We match on stable
	// literals. The menu has no trailing `> ` sentinel (per plan G12
	// audit note), so we read until the "Enter choice" prompt instead.
	out := c.ReadUntil("Enter choice", 3*time.Second)
	for _, want := range []string{
		"Room number",
		"Name",
		"Description",
		"Exit menu",
		"Extra descriptions",
		"Quit",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("menu missing %q; got:\n%s", want, out)
		}
	}

	// Q exits cleanly.
	c.Send("Q")
	ack := c.ReadUntil("Exiting editor", 3*time.Second)
	if !strings.Contains(ack, "Exiting editor") {
		t.Errorf("expected 'Exiting editor' on Q; got %q", ack)
	}
}

// TestTestclient_ReditMenuSetName drives plan §G12 — 1 + text sets the
// room name via the menu path.
func TestTestclient_ReditMenuSetName(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))

	c := h.NewCharacter(t, CharSpec{
		Name:     "Renamer",
		Password: defaultQuickLoginPassword,
		Trust:    types.LEVEL_IMMORTAL,
	})
	defer c.Close()

	c.Send("redit")
	_ = c.ReadUntil("Enter choice", 3*time.Second)

	c.Send("1")
	_ = c.ReadUntil("Enter room name", 3*time.Second)

	c.Send("The Edited Room")
	// After setting, the menu redisplays. Match on the new name.
	out := c.ReadUntil("Enter choice", 3*time.Second)
	if !strings.Contains(out, "The Edited Room") {
		t.Errorf("new room name not in menu redisplay; got %q", out)
	}
}

// TestTestclient_ReditInvalidChoiceRedisplays pins A15 E2E — bad input
// does NOT drop out of CON_REDIT; the menu is redisplayed.
func TestTestclient_ReditInvalidChoiceRedisplays(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))

	c := h.NewCharacter(t, CharSpec{
		Name:     "Bador",
		Password: defaultQuickLoginPassword,
		Trust:    types.LEVEL_IMMORTAL,
	})
	defer c.Close()

	c.Send("redit")
	_ = c.ReadUntil("Enter choice", 3*time.Second)

	c.Send("zzz")
	out := c.ReadUntil("Enter choice", 3*time.Second)
	if !strings.Contains(out, "Invalid choice") {
		t.Errorf("expected 'Invalid choice' on bad input; got %q", out)
	}
	// Confirm we're still in the menu — send a valid Q and expect the
	// clean exit message.
	c.Send("Q")
	exit := c.ReadUntil("Exiting editor", 3*time.Second)
	if !strings.Contains(exit, "Exiting editor") {
		t.Errorf("Q did not produce exit message after invalid input; got %q", exit)
	}
}
