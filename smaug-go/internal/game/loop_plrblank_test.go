package game

import (
	"bytes"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/types"
)

// readAllFromPipe drains the given connection into a buffer until the
// pipe is closed, returning the accumulated bytes. Safe to call after
// the producing side has called FlushOutput and then closed its conn.
func readAllFromPipe(c net.Conn) []byte {
	var mu sync.Mutex
	var out bytes.Buffer
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			n, err := c.Read(buf)
			if n > 0 {
				mu.Lock()
				out.Write(buf[:n])
				mu.Unlock()
			}
			if err != nil {
				return
			}
		}
	}()
	// Wait up to 500ms or until pipe closes.
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
	}
	mu.Lock()
	defer mu.Unlock()
	return out.Bytes()
}

// newPromptTestPlayer builds a minimal CharData + DescriptorData pair
// for prompt-emission tests. The descriptor has a piped Conn whose
// client side is returned for the test to drain after flush.
func newPromptTestPlayer(t *testing.T) (*types.DescriptorData, *types.CharData, net.Conn) {
	t.Helper()
	server, client := net.Pipe()
	t.Cleanup(func() {
		_ = server.Close()
		_ = client.Close()
	})
	d := types.NewDescriptor(server)
	ch := &types.CharData{
		Name:  "Blankt",
		Level: 5,
		Hit:   50, MaxHit: 100,
		Mana: 30, MaxMana: 50,
		Move: 60, MaxMove: 80,
		PCData: &types.PCData{
			Prompt:   "<%hhp> ",
			PagerLen: 24,
		},
	}
	d.Character = ch
	ch.Desc = d
	return d, ch, client
}

// flushAndRead flushes d's output buffer, then closes the server side
// of the pipe so the drain goroutine sees EOF and returns all bytes.
func flushAndRead(t *testing.T, d *types.DescriptorData, client net.Conn) string {
	t.Helper()
	done := make(chan []byte, 1)
	go func() {
		done <- readAllFromPipe(client)
	}()
	if err := d.FlushOutput(); err != nil && err != io.EOF {
		t.Fatalf("FlushOutput: %v", err)
	}
	// Close the server-side conn so the client drain sees EOF.
	if d.Conn != nil {
		_ = d.Conn.Close()
		d.Conn = nil
	}
	return string(<-done)
}

// TestWritePromptWithBlank_PLR_BLANK_EmitsBlankLine verifies that when
// PLR_BLANK is set on the character's Act flags, writePromptWithBlank
// emits a "\n\r" preceding the formatted prompt (matches C behavior at
// src/smaug.c:1359-1361 inside the display_prompt path).
func TestWritePromptWithBlank_PLR_BLANK_EmitsBlankLine(t *testing.T) {
	d, ch, client := newPromptTestPlayer(t)
	ch.Act.Set(types.PLR_BLANK)

	writePromptWithBlank(d)

	got := flushAndRead(t, d, client)
	if !strings.HasPrefix(got, "\n\r") {
		t.Errorf("with PLR_BLANK set, output should begin with %q; got %q", "\n\r", got)
	}
	if !strings.Contains(got, "<50hp>") {
		t.Errorf("prompt should still contain %q; got %q", "<50hp>", got)
	}
}

// TestWritePromptWithBlank_NoFlag_NoLeadingBlankLine verifies that
// without PLR_BLANK set, the prompt is emitted without the leading
// "\n\r" prefix (default SMAUG behavior).
func TestWritePromptWithBlank_NoFlag_NoLeadingBlankLine(t *testing.T) {
	d, _, client := newPromptTestPlayer(t)
	// Explicitly ensure PLR_BLANK is NOT set.

	writePromptWithBlank(d)

	got := flushAndRead(t, d, client)
	if strings.HasPrefix(got, "\n\r") {
		t.Errorf("without PLR_BLANK set, output should NOT begin with %q; got %q", "\n\r", got)
	}
	if !strings.HasPrefix(got, "<50hp>") {
		t.Errorf("output should start directly with the prompt; got %q", got)
	}
}

// TestWritePromptWithBlank_NilDescriptor_NoPanic verifies the helper is
// nil-safe — a nil descriptor produces no action and no panic. This
// guards the enterGame path where, on a broken connection, d may be
// partially cleaned up.
func TestWritePromptWithBlank_NilDescriptor_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("writePromptWithBlank(nil) panicked: %v", r)
		}
	}()
	writePromptWithBlank(nil)
}

// TestWritePromptWithBlank_NilCharacter_NoPanic verifies the helper
// tolerates a descriptor with no Character attached (this occurs during
// early login before enterGame has run).
func TestWritePromptWithBlank_NilCharacter_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("writePromptWithBlank with nil char panicked: %v", r)
		}
	}()
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()
	d := types.NewDescriptor(server)
	// Character deliberately unset.
	writePromptWithBlank(d)
}
