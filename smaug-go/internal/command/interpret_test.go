package command

import (
	"net"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/types"
)

// makeTestChar creates a CharData with a net.Pipe-backed descriptor so we can
// capture output. Returns the character and the client side of the pipe.
func makeTestChar(name string) (*types.CharData, net.Conn) {
	server, client := net.Pipe()
	d := &types.DescriptorData{
		Conn:       server,
		InputQueue: make(chan string, 10),
	}
	ch := &types.CharData{
		Name:     name,
		Desc:     d,
		Position: types.POS_STANDING,
		Level:    1,
	}
	d.Character = ch
	return ch, client
}

// readOutput flushes the descriptor and reads whatever was written.
// net.Pipe is synchronous, so we must read concurrently with the flush.
func readOutput(ch *types.CharData, client net.Conn) string {
	if !ch.Desc.HasOutput() {
		return ""
	}
	result := make(chan string, 1)
	go func() {
		buf := make([]byte, 4096)
		client.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, _ := client.Read(buf)
		result <- string(buf[:n])
	}()
	_ = ch.Desc.FlushOutput()
	return <-result
}

// ---------- Registry.Find tests ----------

func TestFind_ExactMatch(t *testing.T) {
	r := NewRegistry()
	r.Register(&Command{Name: "look", Position: 0, Level: 0})

	cmd := r.Find("look", 0)
	if cmd == nil {
		t.Fatal("expected to find 'look', got nil")
	}
	if cmd.Name != "look" {
		t.Fatalf("expected command name 'look', got %q", cmd.Name)
	}
}

func TestFind_PrefixMatch(t *testing.T) {
	r := NewRegistry()
	r.Register(&Command{Name: "look", Position: 0, Level: 0})

	cmd := r.Find("lo", 0)
	if cmd == nil {
		t.Fatal("expected prefix match for 'lo', got nil")
	}
	if cmd.Name != "look" {
		t.Fatalf("expected 'look', got %q", cmd.Name)
	}
}

func TestFind_NoMatch(t *testing.T) {
	r := NewRegistry()
	r.Register(&Command{Name: "look", Position: 0, Level: 0})

	cmd := r.Find("xyz", 0)
	if cmd != nil {
		t.Fatalf("expected nil for 'xyz', got %q", cmd.Name)
	}
}

func TestFind_TrustCheck(t *testing.T) {
	r := NewRegistry()
	r.Register(&Command{Name: "wizhelp", Position: 0, Level: 50})

	// Insufficient trust
	cmd := r.Find("wizhelp", 0)
	if cmd != nil {
		t.Fatal("expected nil when trust is insufficient")
	}

	// Sufficient trust
	cmd = r.Find("wizhelp", 50)
	if cmd == nil {
		t.Fatal("expected to find 'wizhelp' at trust 50")
	}
}

func TestFind_ExactMatchBeatsPrefixMatch(t *testing.T) {
	r := NewRegistry()
	r.Register(&Command{Name: "north", Position: 0, Level: 0})
	r.Register(&Command{Name: "northeast", Position: 0, Level: 0})

	// "north" should exact-match "north", not prefix-match "northeast"
	cmd := r.Find("north", 0)
	if cmd == nil {
		t.Fatal("expected to find 'north', got nil")
	}
	if cmd.Name != "north" {
		t.Fatalf("expected exact match 'north', got %q", cmd.Name)
	}

	// "northe" should prefix-match "northeast"
	cmd = r.Find("northe", 0)
	if cmd == nil {
		t.Fatal("expected to find 'northeast' via prefix, got nil")
	}
	if cmd.Name != "northeast" {
		t.Fatalf("expected prefix match 'northeast', got %q", cmd.Name)
	}
}

// ---------- Registry.Interpret tests ----------

func TestInterpret_DispatchesCommand(t *testing.T) {
	r := NewRegistry()
	dispatched := false
	r.Register(&Command{
		Name:     "test",
		Position: 0,
		Level:    0,
		DoFun: func(ch *types.CharData, argument string) {
			dispatched = true
		},
	})

	ch, client := makeTestChar("Tester")
	defer client.Close()
	defer ch.Desc.Conn.Close()

	r.Interpret(ch, "test")
	if !dispatched {
		t.Fatal("expected command to be dispatched")
	}
}

func TestInterpret_EmptyInput(t *testing.T) {
	r := NewRegistry()
	// Should not panic or produce output
	ch, client := makeTestChar("Tester")
	defer client.Close()
	defer ch.Desc.Conn.Close()

	r.Interpret(ch, "")
	r.Interpret(ch, "   ")

	out := readOutput(ch, client)
	if out != "" {
		t.Fatalf("expected no output for empty input, got %q", out)
	}
}

func TestInterpret_UnknownCommand(t *testing.T) {
	r := NewRegistry()
	ch, client := makeTestChar("Tester")
	defer client.Close()
	defer ch.Desc.Conn.Close()

	r.Interpret(ch, "xyzzy")

	out := readOutput(ch, client)
	if out != "Huh?\n\r" {
		t.Fatalf("expected %q, got %q", "Huh?\n\r", out)
	}
}

func TestInterpret_PositionCheck(t *testing.T) {
	r := NewRegistry()
	r.Register(&Command{
		Name:     "stand",
		Position: types.POS_STANDING,
		Level:    0,
		DoFun: func(ch *types.CharData, argument string) {
			ch.Send("You are already standing.\n\r")
		},
	})

	ch, client := makeTestChar("Sleeper")
	defer client.Close()
	defer ch.Desc.Conn.Close()

	ch.Position = types.POS_SLEEPING
	r.Interpret(ch, "stand")

	out := readOutput(ch, client)
	expected := "In your dreams, or what?\n\r"
	if out != expected {
		t.Fatalf("expected %q, got %q", expected, out)
	}
}
