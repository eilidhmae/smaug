package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/act"
	"github.com/eilidhmae/smaug/internal/game"
	"github.com/eilidhmae/smaug/internal/mudprog"
	smaugnet "github.com/eilidhmae/smaug/internal/net"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// testServer starts a full game server on a random port and returns
// the port, a cancel function, and a WaitGroup that completes when
// the game loop exits.
func testServer(t *testing.T) (int, context.CancelFunc, *sync.WaitGroup) {
	t.Helper()

	// Remove stale player files from prior runs before booting.
	os.RemoveAll(filepath.Join("testdata", "player"))

	w := world.New("testdata")
	act.WorldRef = w

	if err := bootDB(w); err != nil {
		t.Fatalf("bootDB: %v", err)
	}

	cmdReg := registerCommands()
	server := smaugnet.NewServer()
	gameLoop := game.NewGameLoop(w, cmdReg, server.Incoming)

	act.SaveFunc = func(ch *types.CharData) { gameLoop.SavePlayer(ch) }
	act.CmdRegistry = cmdReg
	cmdReg.SocialFallback = act.CheckSocial
	mudprog.CmdRegistry = cmdReg
	mudprog.WorldRef = w
	act.StartEditingFunc = game.StartEditing

	// Listen on port 0 to get a random free port.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	// Close this listener — we'll use server.Start with the specific port.
	listener.Close()

	if err := server.Start(port); err != nil {
		t.Fatalf("server.Start: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		gameLoop.Run(ctx)
	}()

	t.Cleanup(func() {
		cancel()
		wg.Wait()
		server.Stop()
	})

	// Clean up any player save files from previous runs.
	playerDir := filepath.Join("testdata", "player")
	t.Cleanup(func() {
		os.RemoveAll(playerDir)
	})

	// Give the server a moment to start accepting.
	time.Sleep(100 * time.Millisecond)

	return port, cancel, &wg
}

// mudClient wraps a TCP connection to the MUD for test convenience.
// Uses raw byte reading instead of line-based scanning because MUD
// output uses \n\r line endings and prompts don't end with newlines.
type mudClient struct {
	conn net.Conn
	buf  string // accumulated unprocessed output
	t    *testing.T
}

func dial(t *testing.T, port int) *mudClient {
	t.Helper()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 2*time.Second)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return &mudClient{
		conn: conn,
		t:    t,
	}
}

// send writes a line to the server (appends \n).
func (c *mudClient) send(line string) {
	c.t.Helper()
	c.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	_, err := fmt.Fprintf(c.conn, "%s\n", line)
	if err != nil {
		c.t.Fatalf("send %q: %v", line, err)
	}
}

// readUntil reads raw bytes until the accumulated output contains substr,
// or times out. Returns all accumulated output up to and including the match.
func (c *mudClient) readUntil(substr string, timeout time.Duration) string {
	c.t.Helper()
	deadline := time.Now().Add(timeout)
	tmp := make([]byte, 4096)

	for {
		// Check if we already have a match in the buffer
		lower := strings.ToLower(c.buf)
		if idx := strings.Index(lower, strings.ToLower(substr)); idx >= 0 {
			result := c.buf
			c.buf = ""
			return result
		}

		if time.Now().After(deadline) {
			c.t.Fatalf("readUntil %q: timeout (buf has %d bytes: %q)",
				substr, len(c.buf), truncate(c.buf, 200))
			return ""
		}

		c.conn.SetReadDeadline(deadline)
		n, err := c.conn.Read(tmp)
		if n > 0 {
			c.buf += string(tmp[:n])
		}
		if err != nil {
			// Check one more time after the final read
			if strings.Contains(strings.ToLower(c.buf), strings.ToLower(substr)) {
				result := c.buf
				c.buf = ""
				return result
			}
			c.t.Fatalf("readUntil %q: %v (buf: %q)", substr, err, truncate(c.buf, 200))
			return ""
		}
	}
}

// readFor reads all available output for the given duration.
func (c *mudClient) readFor(d time.Duration) string {
	c.t.Helper()
	deadline := time.Now().Add(d)
	tmp := make([]byte, 4096)
	for {
		c.conn.SetReadDeadline(deadline)
		n, err := c.conn.Read(tmp)
		if n > 0 {
			c.buf += string(tmp[:n])
		}
		if err != nil {
			break
		}
	}
	result := c.buf
	c.buf = ""
	return result
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// createCharacter drives the full character creation flow and returns
// a client positioned at the game prompt.
func createCharacter(t *testing.T, port int, name, password string) *mudClient {
	t.Helper()
	c := dial(t, port)

	// Wait for greeting
	c.readUntil("what name", 3*time.Second)

	// Enter name
	c.send(name)
	c.readUntil("Did I get that right", 2*time.Second)

	// Confirm name
	c.send("y")
	c.readUntil("password", 2*time.Second)

	// Enter password
	c.send(password)
	c.readUntil("retype", 2*time.Second)

	// Confirm password
	c.send(password)
	c.readUntil("gender", 2*time.Second)

	// Select sex
	c.send("m")
	c.readUntil("class", 2*time.Second)

	// Select class — use first available
	c.send("warrior")
	c.readUntil("race", 2*time.Second)

	// Select race
	c.send("human")
	c.readUntil("Press Enter", 2*time.Second)

	// Press enter for MOTD
	c.send("")
	c.readUntil("Welcome", 3*time.Second)

	return c
}

func TestIntegration_ServerBoot(t *testing.T) {
	port, _, _ := testServer(t)

	c := dial(t, port)
	output := c.readUntil("what name", 3*time.Second)
	if !strings.Contains(strings.ToLower(output), "name") {
		t.Errorf("greeting doesn't ask for name: %q", truncate(output, 100))
	}
}

func TestIntegration_CharacterCreation(t *testing.T) {
	port, _, _ := testServer(t)

	c := createCharacter(t, port, "Testchar", "secret123")

	// Should be in the temple room
	c.send("look")
	c.readUntil("Temple", 2*time.Second)
}

func TestIntegration_Commands(t *testing.T) {
	port, _, _ := testServer(t)

	c := createCharacter(t, port, "Cmdtest", "secret123")

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
			c.send(tt.cmd)
			c.readUntil(tt.expect, 2*time.Second)
		})
	}
}

func TestIntegration_Communication(t *testing.T) {
	port, _, _ := testServer(t)

	c := createCharacter(t, port, "Talker", "secret123")

	c.send("say hello world")
	c.readUntil("hello world", 2*time.Second)

	c.send("emote waves")
	c.readUntil("waves", 2*time.Second)
}

func TestIntegration_InvalidName(t *testing.T) {
	port, _, _ := testServer(t)

	c := dial(t, port)
	c.readUntil("what name", 3*time.Second)

	// Name too short
	c.send("ab")
	c.readUntil("Illegal name", 2*time.Second)

	// Name with numbers
	c.send("test123")
	c.readUntil("Illegal name", 2*time.Second)

	// Valid name works
	c.send("Validname")
	c.readUntil("Did I get that right", 2*time.Second)
}

func TestIntegration_BadPassword(t *testing.T) {
	port, _, _ := testServer(t)

	c := dial(t, port)
	c.readUntil("what name", 3*time.Second)

	c.send("Passtest")
	c.readUntil("Did I get that right", 2*time.Second)
	c.send("y")
	c.readUntil("password", 2*time.Second)

	// Password too short
	c.send("abc")
	c.readUntil("at least five", 2*time.Second)

	// Valid password
	c.send("goodpass")
	c.readUntil("retype", 2*time.Second)

	// Wrong confirmation
	c.send("different")
	c.readUntil("don't match", 2*time.Second)
}

func TestIntegration_Quit(t *testing.T) {
	port, _, _ := testServer(t)

	c := createCharacter(t, port, "Quitter", "secret123")

	c.send("quit")
	// Connection should close eventually
	c.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	buf := make([]byte, 4096)
	for {
		_, err := c.conn.Read(buf)
		if err != nil {
			break // expected: connection closed
		}
	}
}

func TestIntegration_Help(t *testing.T) {
	port, _, _ := testServer(t)

	c := createCharacter(t, port, "Helper", "secret123")

	c.send("help")
	// Should show something — just verify no crash
	_ = c.readFor(1 * time.Second)
}

// Verify multiple simultaneous connections work.
func TestIntegration_MultipleConnections(t *testing.T) {
	port, _, _ := testServer(t)

	c1 := createCharacter(t, port, "Playerx", "secret123")
	c2 := createCharacter(t, port, "Playery", "secret123")

	// Both should be able to use commands
	c1.send("who")
	c1.readUntil("Playerx", 2*time.Second)

	c2.send("who")
	c2.readUntil("Playery", 2*time.Second)
}
