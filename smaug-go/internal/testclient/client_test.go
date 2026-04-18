package testclient

import (
	"bytes"
	"net"
	"regexp"
	"strings"
	"testing"
	"time"
)

// --- stripIAC tests ---

func TestStripIAC_Passthrough(t *testing.T) {
	in := []byte("Hello, world!")
	clean, consumed := stripIAC(in)
	if string(clean) != "Hello, world!" {
		t.Errorf("clean = %q, want %q", clean, "Hello, world!")
	}
	if consumed != len(in) {
		t.Errorf("consumed = %d, want %d", consumed, len(in))
	}
}

func TestStripIAC_WillWontDoDont(t *testing.T) {
	// IAC WILL ECHO (255 251 1), IAC WONT ECHO (255 252 1),
	// IAC DO TERMTYPE (255 253 24), IAC DONT TERMTYPE (255 254 24).
	tests := []struct {
		name string
		in   []byte
	}{
		{"WILL", []byte{255, 251, 1}},
		{"WONT", []byte{255, 252, 1}},
		{"DO", []byte{255, 253, 24}},
		{"DONT", []byte{255, 254, 24}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			full := append([]byte("pre"), tt.in...)
			full = append(full, []byte("post")...)
			clean, consumed := stripIAC(full)
			if string(clean) != "prepost" {
				t.Errorf("clean = %q, want %q", clean, "prepost")
			}
			if consumed != len(full) {
				t.Errorf("consumed = %d, want %d", consumed, len(full))
			}
		})
	}
}

func TestStripIAC_Subnegotiation(t *testing.T) {
	// IAC SB TERMTYPE IS 'x' 'y' IAC SE
	in := []byte{'a', 255, 250, 24, 0, 'x', 'y', 255, 240, 'b'}
	clean, consumed := stripIAC(in)
	if string(clean) != "ab" {
		t.Errorf("clean = %q, want %q", clean, "ab")
	}
	if consumed != len(in) {
		t.Errorf("consumed = %d, want %d", consumed, len(in))
	}
}

func TestStripIAC_Literal(t *testing.T) {
	// IAC IAC (255 255) → literal 0xFF.
	in := []byte{'a', 255, 255, 'b'}
	clean, consumed := stripIAC(in)
	if len(clean) != 3 || clean[0] != 'a' || clean[1] != 255 || clean[2] != 'b' {
		t.Errorf("clean = %v, want [a, 0xFF, b]", clean)
	}
	if consumed != len(in) {
		t.Errorf("consumed = %d, want %d", consumed, len(in))
	}
}

func TestStripIAC_Truncated(t *testing.T) {
	// IAC at end: no follow-up byte, must leave input unconsumed for next read.
	in := []byte{'a', 'b', 255}
	clean, consumed := stripIAC(in)
	if string(clean) != "ab" {
		t.Errorf("clean = %q, want %q", clean, "ab")
	}
	if consumed != 2 {
		t.Errorf("consumed = %d, want 2 (IAC must be preserved for next read)", consumed)
	}

	// IAC WILL at end, option byte missing
	in2 := []byte{'a', 255, 251}
	clean2, consumed2 := stripIAC(in2)
	if string(clean2) != "a" {
		t.Errorf("clean2 = %q, want %q", clean2, "a")
	}
	if consumed2 != 1 {
		t.Errorf("consumed2 = %d, want 1", consumed2)
	}

	// IAC SB not yet closed with IAC SE
	in3 := []byte{'a', 255, 250, 24, 0, 'x'}
	clean3, consumed3 := stripIAC(in3)
	if string(clean3) != "a" {
		t.Errorf("clean3 = %q, want %q", clean3, "a")
	}
	if consumed3 != 1 {
		t.Errorf("consumed3 = %d, want 1 (SB sequence must be preserved for next read)", consumed3)
	}
}

// --- stripANSI tests ---

func TestStripANSI_Passthrough(t *testing.T) {
	in := []byte("Hello, world!")
	out := stripANSI(in)
	if string(out) != "Hello, world!" {
		t.Errorf("out = %q, want %q", out, "Hello, world!")
	}
}

func TestStripANSI_Colors(t *testing.T) {
	in := []byte("\x1b[31mred\x1b[0m")
	out := stripANSI(in)
	if string(out) != "red" {
		t.Errorf("out = %q, want %q", out, "red")
	}
}

func TestStripANSI_CursorAndOSC(t *testing.T) {
	// CSI with cursor-position (H), and OSC title set (ESC ] 0 ; title BEL).
	in := []byte("\x1b[2Jfoo\x1b]0;title\x07bar")
	out := stripANSI(in)
	if string(out) != "foobar" {
		t.Errorf("out = %q, want %q", out, "foobar")
	}
}

// TestStripANSI_Truncated: a lone ESC at the end of buffer is dropped. This
// is a deliberate simplification — MUDs don't split escapes across reads at
// 4KB granularity in practice.
func TestStripANSI_Truncated(t *testing.T) {
	in := []byte("abc\x1b")
	out := stripANSI(in)
	if string(out) != "abc" {
		t.Errorf("out = %q, want %q (trailing ESC should be dropped)", out, "abc")
	}
}

// --- Client tests (using net.Pipe) ---

func TestClient_SendReadUntil(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	go func() {
		_, _ = serverConn.Write([]byte("Welcome, hello there.\n"))
	}()

	c := newTestClient(t, clientConn)
	got := c.ReadUntil("hello", 1*time.Second)
	if !strings.Contains(strings.ToLower(got), "hello") {
		t.Errorf("got = %q, want to contain %q", got, "hello")
	}
}

func TestClient_ReadUntilTimeout(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	c := newTestClient(t, clientConn)
	_, err := c.readUntilErr("never", 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline") {
		t.Errorf("err = %v, want timeout", err)
	}
}

func TestClient_ReadMatch_Groups(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	go func() {
		_, _ = serverConn.Write([]byte("hp=42/100 mv=33"))
	}()

	c := newTestClient(t, clientConn)
	re := regexp.MustCompile(`hp=(\d+)/(\d+)`)
	m := c.ReadMatch(re, 1*time.Second)
	if len(m) < 3 {
		t.Fatalf("expected 3 submatches, got %v", m)
	}
	if m[1] != "42" || m[2] != "100" {
		t.Errorf("got %q/%q, want 42/100", m[1], m[2])
	}
}

func TestClient_ReadToPrompt_Default(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	go func() {
		_, _ = serverConn.Write([]byte("You see a room.\n> "))
	}()

	c := newTestClient(t, clientConn)
	got := c.ReadToPrompt(1 * time.Second)
	if !strings.HasSuffix(got, "> ") {
		t.Errorf("got = %q, want suffix '> '", got)
	}
	if !strings.Contains(got, "room") {
		t.Errorf("got = %q, want to contain 'room'", got)
	}
}

func TestClient_WithPrompt_Override(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	go func() {
		_, _ = serverConn.Write([]byte("shell output\n$ "))
	}()

	c := newTestClient(t, clientConn).WithPrompt(`\$\s$`)
	got := c.ReadToPrompt(1 * time.Second)
	if !strings.HasSuffix(got, "$ ") {
		t.Errorf("got = %q, want suffix '$ '", got)
	}
}

func TestClient_StrippedIntegration(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	go func() {
		// IAC WILL ECHO + ANSI red + text + prompt.
		payload := []byte{255, 251, 1}
		payload = append(payload, []byte("\x1b[31mHello\x1b[0m, > ")...)
		_, _ = serverConn.Write(payload)
	}()

	c := newTestClient(t, clientConn)
	got := c.ReadToPrompt(1 * time.Second)
	if got != "Hello, > " {
		t.Errorf("got = %q, want %q", got, "Hello, > ")
	}
}

// TestClient_Send verifies Send writes line+"\n" to the underlying conn.
func TestClient_Send(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	type readResult struct {
		data []byte
		err  error
	}
	got := make(chan readResult, 1)
	go func() {
		buf := make([]byte, 32)
		n, err := serverConn.Read(buf)
		got <- readResult{data: buf[:n], err: err}
	}()

	c := newTestClient(t, clientConn)
	c.Send("hello")

	select {
	case r := <-got:
		if r.err != nil {
			t.Fatalf("peer read error: %v", r.err)
		}
		if string(r.data) != "hello\n" {
			t.Errorf("peer got %q, want %q", string(r.data), "hello\n")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("peer did not receive write within 1s")
	}
}

// TestClient_SendErr_ClosedPeer verifies sendErr returns an I/O error when
// the peer has closed the connection. sendErr is the error-returning
// helper that Send wraps; testing it directly lets us assert failure
// handling without triggering t.Fatalf.
func TestClient_SendErr_ClosedPeer(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	// Close peer immediately so the subsequent Write returns an error.
	serverConn.Close()
	defer clientConn.Close()

	c := newTestClient(t, clientConn)
	if err := c.sendErr("data"); err == nil {
		t.Fatal("sendErr to closed peer should return error, got nil")
	}
}

// first Read delivers a bare trailing IAC byte that stripIAC cannot yet
// consume, and a later Read supplies the rest of the WILL ECHO triple plus
// payload. The final buffer must contain 'C' with all IAC bytes stripped.
func TestClient_IACSplitAcrossReads(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	c := newTestClient(t, clientConn)

	// Write the first chunk in a goroutine — server pipe unbuffered.
	go func() {
		_, _ = serverConn.Write([]byte{'A', 'B', 0xFF})
	}()
	// Drain until "AB" is visible — confirms the truncated IAC was held
	// in rawTail and "AB" reached the buffer.
	got := c.ReadUntil("AB", 1*time.Second)
	if !strings.Contains(got, "AB") {
		t.Fatalf("first read buffer = %q, want to contain 'AB'", got)
	}

	// Write the rest: complete IAC WILL ECHO plus a 'C' payload byte.
	go func() {
		_, _ = serverConn.Write([]byte{0xFB, 0x01, 'C'})
	}()
	got2 := c.ReadUntil("C", 1*time.Second)
	// Must contain the 'C' — and it must NOT contain any IAC bytes
	// (0xFF, 0xFB, or 0x01 raw option byte).
	if !strings.Contains(got2, "C") {
		t.Fatalf("second read buffer = %q, want to contain 'C'", got2)
	}
	for _, b := range []byte(got2) {
		if b == 0xFF || b == 0xFB {
			t.Errorf("second read buffer has IAC byte 0x%02X: %q", b, got2)
		}
	}
}

// newTestClient builds a Client around an arbitrary net.Conn for unit tests
// (no real server involved). Defaults: stripANSI on, stripIAC on, default prompt.
func newTestClient(t *testing.T, conn net.Conn) *Client {
	t.Helper()
	return &Client{
		conn:      conn,
		t:         t,
		stripANSI: true,
		stripIAC:  true,
	}
}

// TestClient_ScratchBufferReused verifies that the read scratch slice is
// allocated lazily once and persists across fillOnce calls — a regression
// guard against reintroducing per-call `make([]byte, 4096)`. Uses the
// identity of the underlying array via &c.scratch[0] since slices of the
// same backing array compare that way.
func TestClient_ScratchBufferReused(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	c := newTestClient(t, clientConn)
	if c.scratch != nil {
		t.Fatal("scratch must be nil before first fillOnce")
	}

	// Pump one byte through to trigger allocation.
	go func() { _, _ = serverConn.Write([]byte("X")) }()
	_, _ = c.readUntilErr("X", 500*time.Millisecond)

	if c.scratch == nil {
		t.Fatal("scratch must be non-nil after first fillOnce")
	}
	if len(c.scratch) != readScratchSize {
		t.Errorf("scratch len = %d, want %d", len(c.scratch), readScratchSize)
	}
	firstPtr := &c.scratch[0]

	// Do a second round-trip. Scratch must NOT be reallocated.
	go func() { _, _ = serverConn.Write([]byte("Y")) }()
	_, _ = c.readUntilErr("Y", 500*time.Millisecond)

	if &c.scratch[0] != firstPtr {
		t.Error("scratch must be reused across fillOnce calls, not reallocated")
	}
}

// TestClient_ReadUntil_CaseInsensitive_AfterBytesRefactor verifies that
// the bytes.ToLower / bytes.Index refactor preserved the original
// case-insensitive match semantics.
func TestClient_ReadUntil_CaseInsensitive_AfterBytesRefactor(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	go func() { _, _ = serverConn.Write([]byte("Some MIXED Case Output")) }()
	c := newTestClient(t, clientConn)
	got := c.ReadUntil("mixed", 500*time.Millisecond)
	if !bytes.Contains([]byte(got), []byte("MIXED")) {
		t.Errorf("got = %q, want to contain 'MIXED'", got)
	}
}
