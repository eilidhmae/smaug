package main

import (
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/testclient"
)

// These tests exercise the full server boot + telnet dialogue path end to
// end. The heavy lifting (harness boot, nanny flow, telnet/ANSI-aware
// reads) lives in internal/testclient; this file is a thin wrapper that
// covers cmd/smaug's specific integration scenarios using the shared
// helpers.

func TestIntegration_ServerBoot(t *testing.T) {
	h := testclient.Start(t)

	c := h.Dial(t)
	output := c.ReadUntil("what name", 3*time.Second)
	if !strings.Contains(strings.ToLower(output), "name") {
		t.Errorf("greeting doesn't ask for name: %q", output)
	}
}

func TestIntegration_CharacterCreation(t *testing.T) {
	h := testclient.Start(t)

	c := h.NewCharacter(t, testclient.CharSpec{Name: "Testchar", Password: "secret123"})

	// Should be in the temple room.
	c.Send("look")
	c.ReadUntil("Temple", 2*time.Second)
}

func TestIntegration_Commands(t *testing.T) {
	h := testclient.Start(t)

	c := h.NewCharacter(t, testclient.CharSpec{Name: "Cmdtest", Password: "secret123"})

	tests := []struct {
		cmd    string
		expect string
	}{
		{"score", "Cmdtest"},
		{"who", "Cmdtest"},
		{"inventory", "carrying"},
		{"equipment", "using"},
		{"time", "hour"},
		{"commands", "look"},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			c.Send(tt.cmd)
			c.ReadUntil(tt.expect, 2*time.Second)
		})
	}
}

func TestIntegration_Communication(t *testing.T) {
	h := testclient.Start(t)

	c := h.NewCharacter(t, testclient.CharSpec{Name: "Talker", Password: "secret123"})

	c.Send("say hello world")
	c.ReadUntil("hello world", 2*time.Second)

	c.Send("emote waves")
	c.ReadUntil("waves", 2*time.Second)
}

func TestIntegration_InvalidName(t *testing.T) {
	h := testclient.Start(t)

	c := h.Dial(t)
	c.ReadUntil("what name", 3*time.Second)

	// Name too short.
	c.Send("ab")
	c.ReadUntil("Illegal name", 2*time.Second)

	// Name with numbers.
	c.Send("test123")
	c.ReadUntil("Illegal name", 2*time.Second)

	// Valid name works.
	c.Send("Validname")
	c.ReadUntil("Did I get that right", 2*time.Second)
}

func TestIntegration_BadPassword(t *testing.T) {
	h := testclient.Start(t)

	c := h.Dial(t)
	c.ReadUntil("what name", 3*time.Second)

	c.Send("Passtest")
	c.ReadUntil("Did I get that right", 2*time.Second)
	c.Send("y")
	c.ReadUntil("password", 2*time.Second)

	// Password too short.
	c.Send("abc")
	c.ReadUntil("at least five", 2*time.Second)

	// Valid password.
	c.Send("goodpass")
	c.ReadUntil("retype", 2*time.Second)

	// Wrong confirmation.
	c.Send("different")
	c.ReadUntil("don't match", 2*time.Second)
}

func TestIntegration_Quit(t *testing.T) {
	h := testclient.Start(t)

	c := h.NewCharacter(t, testclient.CharSpec{Name: "Quitter", Password: "secret123"})

	c.Send("quit")
	// Connection should close eventually; ReadFor drains everything
	// until the remote close unblocks the read loop.
	_ = c.ReadFor(3 * time.Second)
}

func TestIntegration_Help(t *testing.T) {
	h := testclient.Start(t)

	c := h.NewCharacter(t, testclient.CharSpec{Name: "Helper", Password: "secret123"})

	c.Send("help")
	// Should show something — just verify no crash.
	_ = c.ReadFor(1 * time.Second)
}

// Verify multiple simultaneous connections work.
func TestIntegration_MultipleConnections(t *testing.T) {
	h := testclient.Start(t)

	c1 := h.NewCharacter(t, testclient.CharSpec{Name: "Playerx", Password: "secret123"})
	c2 := h.NewCharacter(t, testclient.CharSpec{Name: "Playery", Password: "secret123"})

	// Both should be able to use commands.
	c1.Send("who")
	c1.ReadUntil("Playerx", 2*time.Second)

	c2.Send("who")
	c2.ReadUntil("Playery", 2*time.Second)
}
