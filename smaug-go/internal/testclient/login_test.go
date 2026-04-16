package testclient

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// TestNewCharacter_LandsAtPrompt confirms NewCharacter drives the nanny
// flow end-to-end and returns a client that can issue commands.
func TestNewCharacter_LandsAtPrompt(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))

	c := h.NewCharacter(t, CharSpec{
		Name:     "Prompter",
		Password: "secret123",
	})

	c.Send("look")
	got := c.ReadUntil("Temple", 3*time.Second)
	if !strings.Contains(got, "Temple") {
		t.Errorf("look did not show Temple; got %q", got)
	}
}

// TestNewCharacter_Trust verifies the Trust field is applied after
// creation and round-trips through the player save file when the
// character logs back in.
func TestNewCharacter_Trust(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))

	name := "Trusty"
	password := "secret123"

	c := h.NewCharacter(t, CharSpec{
		Name:     name,
		Password: password,
		Trust:    types.LEVEL_IMMORTAL,
	})

	// Verify pre-save trust is set.
	var preTrust int
	h.Query(func(w *world.World) {
		for _, d := range w.Descriptors {
			if d.Character != nil && strings.EqualFold(d.Character.Name, name) {
				preTrust = d.Character.Trust
				return
			}
		}
	})
	if preTrust != types.LEVEL_IMMORTAL {
		t.Fatalf("pre-save Trust = %d, want %d", preTrust, types.LEVEL_IMMORTAL)
	}

	// Disconnect and wait for the save-on-cleanup pulse to run.
	c.Close()
	waitForSave(t, h, name)
	waitForDescriptorGone(t, h, name)

	// Relog and check Trust survives.
	c2 := h.Login(t, name, password)

	var postTrust int
	h.Query(func(w *world.World) {
		for _, d := range w.Descriptors {
			if d.Character != nil && strings.EqualFold(d.Character.Name, name) {
				postTrust = d.Character.Trust
				return
			}
		}
	})
	if postTrust != types.LEVEL_IMMORTAL {
		t.Errorf("post-reload Trust = %d, want %d", postTrust, types.LEVEL_IMMORTAL)
	}

	c2.Close()
}

// TestLogin_ReturningPlayer creates then reconnects, confirming the
// returning-player nanny path lands at the in-game prompt.
func TestLogin_ReturningPlayer(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))

	name := "Backagain"
	password := "secret123"

	c := h.NewCharacter(t, CharSpec{Name: name, Password: password})
	c.Close()

	// Wait briefly for the descriptor cleanup/save to run on a pulse.
	waitForSave(t, h, name)
	waitForDescriptorGone(t, h, name)

	c2 := h.Login(t, name, password)
	c2.Send("look")
	got := c2.ReadUntil("Temple", 3*time.Second)
	if !strings.Contains(got, "Temple") {
		t.Errorf("look after relog did not show Temple; got %q", got)
	}
	c2.Close()
}

// TestLogin_WrongPassword_Fails dials manually and verifies the server
// rejects a bad password on a returning login. We drive the nanny
// directly rather than going through Login so we can observe the "Wrong
// password" response without triggering t.Fatal.
func TestLogin_WrongPassword_Fails(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))

	name := "Wrongpw"
	password := "secret123"

	c := h.NewCharacter(t, CharSpec{Name: name, Password: password})
	c.Close()

	waitForSave(t, h, name)
	waitForDescriptorGone(t, h, name)

	c2 := h.Dial(t)
	c2.ReadUntil("what name", 3*time.Second)
	c2.Send(name)
	c2.ReadUntil("Password", 3*time.Second)
	c2.Send("notthepassword")

	got := c2.ReadUntil("Wrong password", 3*time.Second)
	if !strings.Contains(got, "Wrong password") {
		t.Errorf("expected 'Wrong password' response; got %q", got)
	}
	c2.Close()
}

// TestQuickLogin_CaseInsensitive proves QuickLogin canonicalizes the
// supplied name to match the server's own capitalization before writing
// or reading the save file. Passing a lowercased name on the first call
// must create a save at the canonical path; a later call with a
// differently-cased variant must reuse that save (take the Login branch)
// rather than trying to create a duplicate.
func TestQuickLogin_CaseInsensitive(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))

	// First call with all-lowercase input; the server will canonicalize
	// to "Mixcase" for storage.
	lower := "mixcase"
	canonical := "Mixcase"
	canonicalPath := filepath.Join(h.dataDir, "player", strings.ToLower(canonical[:1]), canonical)

	c := h.QuickLogin(t, lower)
	c.Send("look")
	c.ReadUntil("Temple", 3*time.Second)
	c.Close()

	waitForSave(t, h, canonical)
	waitForDescriptorGone(t, h, canonical)

	// The save MUST be at the canonical path — not at a path derived
	// from the raw lowercase input.
	if _, err := os.Stat(canonicalPath); err != nil {
		t.Fatalf("save file not at canonical path %s: %v", canonicalPath, err)
	}
	// And no stray lowercase-named file should exist.
	wrongPath := filepath.Join(h.dataDir, "player", "m", lower)
	if _, err := os.Stat(wrongPath); err == nil {
		t.Errorf("unexpected non-canonical save file at %s", wrongPath)
	}

	// Second call with mIxCaSe — also a non-canonical case — must hit
	// the Login branch, i.e. authenticate with the shared password
	// rather than trying to create a new character.
	c2 := h.QuickLogin(t, "mIxCaSe")
	c2.Send("look")
	c2.ReadUntil("Temple", 3*time.Second)
	c2.Close()
}

// TestQuickLogin_CreatesThenReuses covers the branch logic: first call
// creates a fresh character, second call reuses the existing save file.
func TestQuickLogin_CreatesThenReuses(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))

	name := "Newbie"
	playerFile := filepath.Join(h.dataDir, "player", strings.ToLower(name[:1]), name)

	// Precondition: no save file exists yet.
	if _, err := os.Stat(playerFile); !os.IsNotExist(err) {
		t.Fatalf("player file %s exists before first QuickLogin (err=%v)", playerFile, err)
	}

	c := h.QuickLogin(t, name)
	c.Send("look")
	c.ReadUntil("Temple", 3*time.Second)
	c.Close()

	waitForSave(t, h, name)
	waitForDescriptorGone(t, h, name)

	// Now the save file must exist.
	if _, err := os.Stat(playerFile); err != nil {
		t.Fatalf("player file %s missing after first QuickLogin: %v", playerFile, err)
	}

	c2 := h.QuickLogin(t, name)
	c2.Send("look")
	c2.ReadUntil("Temple", 3*time.Second)
	c2.Close()
}

// waitForSave polls for up to 3s until a player save file appears for
// name. Save happens on descriptor cleanup, which runs at the end of
// each pulse (250ms).
func waitForSave(t *testing.T, h *Harness, name string) {
	t.Helper()
	path := filepath.Join(h.dataDir, "player", strings.ToLower(name[:1]), name)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("player file %s did not appear within 3s", path)
}

// waitForDescriptorGone polls via h.Query until no descriptor in the
// world has a Character with the given name. Used after Close to make
// sure the server-side cleanup pulse has removed the player before we
// try to log in again (otherwise we hit "already playing").
func waitForDescriptorGone(t *testing.T, h *Harness, name string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		var stillPresent bool
		h.Query(func(w *world.World) {
			for _, d := range w.Descriptors {
				if d.Character != nil && strings.EqualFold(d.Character.Name, name) {
					stillPresent = true
					return
				}
			}
		})
		if !stillPresent {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("descriptor for %q still present after 3s", name)
}
