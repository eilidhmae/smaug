package testclient

import (
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// TestBio_RoundTripThroughEditor exercises DoBio end-to-end through the
// real string editor and the real SavePlayer pipeline:
//
//  1. Log in as a new character.
//  2. Type `bio`; confirm the editor prompt appears.
//  3. Type a line of text.
//  4. Type `/s` to save — drives the EditorSave closure installed by DoBio.
//  5. Assert ch.PCData.Bio is updated (in-memory check via h.Query).
//  6. Log out via `save` + disconnect, log back in.
//  7. Assert the loaded character has the same bio.
//
// This is the end-to-end acceptance for plan tranche-A item 1. Until
// this test passes, the `bio` command is inert — the DoBio closure
// writes to PCData.Bio but without SavePlayer wiring that gets lost on
// logout (TODO R1 from plan-player-config.md).
func TestBio_RoundTripThroughEditor(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))

	c := h.NewCharacter(t, CharSpec{Name: "Biowriter", Password: "secret"})

	// Drain any post-login banner so `bio`'s editor prompt is clean.
	_ = c.ReadFor(100 * time.Millisecond)

	expectedBio := "Lived in the mountains for seven years."
	c.Send("bio")
	// StartEditing sends "Begin entering your text now..." plus a "> " prompt.
	c.ReadUntil("Begin entering your text now", 2*time.Second)

	c.Send(expectedBio)
	// Editor echoes a "> " for the next line.
	c.ReadUntil("> ", 2*time.Second)

	c.Send("/s")
	// Give the save closure a moment to land and CON_PLAYING to return.
	_ = c.ReadFor(300 * time.Millisecond)

	// Assert in-memory state.
	var found *types.CharData
	h.Query(func(w *world.World) {
		for _, d := range w.Descriptors {
			if d.Character != nil && strings.EqualFold(d.Character.Name, "Biowriter") {
				found = d.Character
				return
			}
		}
	})
	if found == nil {
		t.Fatal("Biowriter not in-world after /s")
	}
	if !strings.Contains(found.PCData.Bio, expectedBio) {
		t.Errorf("in-memory Bio = %q, want to contain %q", found.PCData.Bio, expectedBio)
	}

	// Save via `save` command (Level gate: need >= 2). Trust shim: bump
	// level before save to clear the level<2 block in DoSave.
	h.Query(func(w *world.World) {
		for _, d := range w.Descriptors {
			if d.Character != nil && strings.EqualFold(d.Character.Name, "Biowriter") {
				d.Character.Level = 2
				return
			}
		}
	})
	c.Send("save")
	c.ReadUntil("Saved", 2*time.Second)
	c.Send("quit")
	c.Close()
	// Wait for the server to flush the quit.
	time.Sleep(200 * time.Millisecond)

	// Re-login. Password was set by NewCharacter (defaults to "secret").
	c2 := h.Login(t, "Biowriter", "secret")
	defer c2.Close()

	var reloaded *types.CharData
	h.Query(func(w *world.World) {
		for _, d := range w.Descriptors {
			if d.Character != nil && strings.EqualFold(d.Character.Name, "Biowriter") {
				reloaded = d.Character
				return
			}
		}
	})
	if reloaded == nil {
		t.Fatal("Biowriter not in-world after relogin")
	}
	if !strings.Contains(reloaded.PCData.Bio, expectedBio) {
		t.Errorf("reloaded Bio = %q, want to contain %q", reloaded.PCData.Bio, expectedBio)
	}
}
