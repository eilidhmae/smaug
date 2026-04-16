package testclient

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/act"
	"github.com/eilidhmae/smaug/internal/world"
)

// CharSpec describes a character to create or log in as. Zero values
// for Sex/Race/Class are replaced with sensible defaults
// ("m"/"human"/"warrior"). Trust != 0 is applied after creation so
// scenarios that need immortal gates can drive OLC/combat commands.
type CharSpec struct {
	Name     string
	Password string
	Sex      string
	Race     string
	Class    string
	Trust    int
}

const (
	// nannyPromptTimeout is the per-step timeout used while driving the
	// nanny state machine. The loop runs at 4 pulses/sec, so worst-case
	// round-trip latency is a couple hundred ms — 3s is plenty of headroom
	// while still surfacing real bugs fast.
	nannyPromptTimeout = 3 * time.Second

	// defaultQuickLoginPassword is used by QuickLogin for both creation
	// and returning-player login paths.
	defaultQuickLoginPassword = "secret123"
)

// NewCharacter drives the full new-character nanny flow and returns a
// Client positioned at the in-game prompt. Fails the test on any
// unexpected nanny response. If spec.Trust is non-zero, Trust is
// applied to the live character and the player save file is rewritten
// so the value persists through relogin.
func (h *Harness) NewCharacter(t *testing.T, spec CharSpec) *Client {
	t.Helper()

	if spec.Sex == "" {
		spec.Sex = "m"
	}
	if spec.Class == "" {
		spec.Class = "warrior"
	}
	if spec.Race == "" {
		spec.Race = "human"
	}

	c := h.Dial(t)
	c.ReadUntil("what name", nannyPromptTimeout)

	steps := []struct{ Prompt, Response string }{
		{"what name", spec.Name},
		{"Did I get that right", "y"},
		{"password", spec.Password},
		{"retype", spec.Password},
		{"gender", spec.Sex},
		{"class", spec.Class},
		{"race", spec.Race},
		{"Press Enter", ""},
	}
	// The first "what name" is already consumed by the greeting read
	// above; send the name and pick up the flow from the next prompt.
	c.Send(steps[0].Response)
	for _, s := range steps[1:] {
		c.ReadUntil(s.Prompt, nannyPromptTimeout)
		c.Send(s.Response)
	}
	c.ReadUntil("Welcome", nannyPromptTimeout)

	if spec.Trust != 0 {
		applyTrust(t, h, spec.Name, spec.Trust)
	}

	return c
}

// applyTrust sets ch.Trust on the named live character and persists the
// change via act.SaveFunc (the package-level hook wired by boot.Boot).
// Runs on the game-loop goroutine via h.Query so we don't race with
// pulse processing.
func applyTrust(t *testing.T, h *Harness, name string, trust int) {
	t.Helper()
	var found bool
	h.Query(func(w *world.World) {
		for _, d := range w.Descriptors {
			if d.Character != nil && strings.EqualFold(d.Character.Name, name) {
				d.Character.Trust = trust
				if act.SaveFunc != nil {
					act.SaveFunc(d.Character)
				}
				found = true
				return
			}
		}
	})
	if !found {
		t.Fatalf("applyTrust: character %q not found in world descriptors", name)
	}
}

// Login dials a new connection and drives the returning-player nanny
// flow: greeting → name → password → MOTD enter → in-game prompt.
// Fails the test if the player file does not exist or the server
// rejects the login.
func (h *Harness) Login(t *testing.T, name, password string) *Client {
	t.Helper()

	c := h.Dial(t)
	c.ReadUntil("what name", nannyPromptTimeout)

	steps := []struct{ Prompt, Response string }{
		{"what name", name},
		{"Password", password},
		{"Press Enter", ""},
	}
	c.Send(steps[0].Response)
	for _, s := range steps[1:] {
		c.ReadUntil(s.Prompt, nannyPromptTimeout)
		c.Send(s.Response)
	}
	c.ReadUntil("Welcome", nannyPromptTimeout)
	return c
}

// QuickLogin creates the character if no save file exists for name,
// otherwise logs in. Uses defaultQuickLoginPassword for both branches
// so successive calls in the same test stay in sync. Sex/Race/Class
// default to m/human/warrior.
//
// The supplied name is canonicalized to match the server's own
// normalization (Capitalize first letter, lowercase the rest) before
// building the save path and before driving the nanny. Callers can pass
// any case variant ("trusty", "TRUSTY", "tRuStY") and hit the same
// on-disk record.
func (h *Harness) QuickLogin(t *testing.T, name string) *Client {
	t.Helper()

	if name == "" {
		t.Fatal("QuickLogin: empty name")
	}
	name = canonicalName(name)
	// SMAUG stores player files at <dataDir>/player/<lowercase-initial>/<Name>.
	path := filepath.Join(h.dataDir, "player", strings.ToLower(name[:1]), name)
	if _, err := os.Stat(path); err == nil {
		return h.Login(t, name, defaultQuickLoginPassword)
	}
	return h.NewCharacter(t, CharSpec{
		Name:     name,
		Password: defaultQuickLoginPassword,
	})
}

// canonicalName mirrors the server's name normalization in
// game.nannyGetName: uppercase first byte, lowercase the rest. Only
// ASCII names are legal (isValidName in the game package restricts to
// [a-zA-Z]{3,12}), so single-byte slicing is safe. An empty input
// returns "" — callers must guard against empty names before calling.
func canonicalName(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}
