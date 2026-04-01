package types

import (
	"net"
	"testing"
)

// ---------------------------------------------------------------------------
// NewDescriptor
// ---------------------------------------------------------------------------

func TestNewDescriptor(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	d := NewDescriptor(server)

	if d.Conn != server {
		t.Error("Conn not set correctly")
	}
	if d.Connected != int(CON_GET_NAME) {
		t.Errorf("Connected = %d, want %d (CON_GET_NAME)", d.Connected, CON_GET_NAME)
	}
	if d.InputQueue == nil {
		t.Error("InputQueue is nil")
	}
	if cap(d.InputQueue) != 100 {
		t.Errorf("InputQueue capacity = %d, want 100", cap(d.InputQueue))
	}
	if d.ScrLen != 24 {
		t.Errorf("ScrLen = %d, want 24", d.ScrLen)
	}
	if d.TelnetState == nil {
		t.Error("TelnetState is nil")
	}
	if d.TelnetState.Width != 80 {
		t.Errorf("TelnetState.Width = %d, want 80", d.TelnetState.Width)
	}
	if d.TelnetState.Height != 24 {
		t.Errorf("TelnetState.Height = %d, want 24", d.TelnetState.Height)
	}
}

// ---------------------------------------------------------------------------
// WriteToBuffer / HasOutput
// ---------------------------------------------------------------------------

func TestDescriptor_WriteToBuffer_HasOutput(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	d := NewDescriptor(server)

	if d.HasOutput() {
		t.Error("new descriptor should not have output")
	}

	d.WriteToBuffer("hello")
	if !d.HasOutput() {
		t.Error("should have output after WriteToBuffer")
	}
}

func TestDescriptor_WriteToBufferf(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	d := NewDescriptor(server)

	d.WriteToBufferf("You see %d %s.", 3, "orcs")
	if !d.HasOutput() {
		t.Error("should have output after WriteToBufferf")
	}
}

// ---------------------------------------------------------------------------
// FlushOutput
// ---------------------------------------------------------------------------

func TestDescriptor_FlushOutput_Empty(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	d := NewDescriptor(server)

	// Flushing with no output should be a no-op and return nil.
	if err := d.FlushOutput(); err != nil {
		t.Errorf("FlushOutput on empty buffer returned error: %v", err)
	}
}

func TestDescriptor_FlushOutput_WritesToConn(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	d := NewDescriptor(server)
	d.WriteToBuffer("test output")

	// Read from the client side in a goroutine.
	done := make(chan string, 1)
	go func() {
		buf := make([]byte, 256)
		n, _ := client.Read(buf)
		done <- string(buf[:n])
	}()

	if err := d.FlushOutput(); err != nil {
		t.Fatalf("FlushOutput error: %v", err)
	}

	got := <-done
	if got != "test output" {
		t.Errorf("flushed output = %q, want %q", got, "test output")
	}

	// Buffer should be empty after flush.
	if d.HasOutput() {
		t.Error("should not have output after FlushOutput")
	}
}

func TestDescriptor_FlushOutput_WithColorFunc(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	d := NewDescriptor(server)
	d.ColorFunc = func(text string, ansiEnabled bool) string {
		return "[colored]" + text
	}

	d.WriteToBuffer("raw text")

	done := make(chan string, 1)
	go func() {
		buf := make([]byte, 256)
		n, _ := client.Read(buf)
		done <- string(buf[:n])
	}()

	if err := d.FlushOutput(); err != nil {
		t.Fatalf("FlushOutput error: %v", err)
	}

	got := <-done
	if got != "[colored]raw text" {
		t.Errorf("flushed output = %q, want %q", got, "[colored]raw text")
	}
}

func TestDescriptor_FlushOutput_ColorFuncAnsiFlag(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	d := NewDescriptor(server)

	var receivedAnsi bool
	d.ColorFunc = func(text string, ansiEnabled bool) string {
		receivedAnsi = ansiEnabled
		return text
	}

	// No character set: ansi should default to true (d.Character == nil).
	d.WriteToBuffer("x")
	go func() {
		buf := make([]byte, 256)
		client.Read(buf)
	}()
	d.FlushOutput()
	if !receivedAnsi {
		t.Error("expected ansiEnabled=true when Character is nil")
	}

	// With character that has PLR_ANSI set.
	ch := &CharData{}
	ch.Act.Set(PLR_ANSI)
	d.Character = ch

	d.WriteToBuffer("y")
	go func() {
		buf := make([]byte, 256)
		client.Read(buf)
	}()
	d.FlushOutput()
	if !receivedAnsi {
		t.Error("expected ansiEnabled=true when PLR_ANSI is set")
	}

	// With character that does NOT have PLR_ANSI.
	ch2 := &CharData{}
	d.Character = ch2

	receivedAnsi = true // Reset to detect change.
	d.WriteToBuffer("z")
	go func() {
		buf := make([]byte, 256)
		client.Read(buf)
	}()
	d.FlushOutput()
	if receivedAnsi {
		t.Error("expected ansiEnabled=false when PLR_ANSI is not set")
	}
}

// ---------------------------------------------------------------------------
// Multiple WriteToBuffer calls accumulate
// ---------------------------------------------------------------------------

func TestDescriptor_WriteToBuffer_Accumulates(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	d := NewDescriptor(server)
	d.WriteToBuffer("one")
	d.WriteToBuffer("two")
	d.WriteToBuffer("three")

	done := make(chan string, 1)
	go func() {
		buf := make([]byte, 256)
		n, _ := client.Read(buf)
		done <- string(buf[:n])
	}()

	d.FlushOutput()
	got := <-done
	if got != "onetwothree" {
		t.Errorf("accumulated output = %q, want %q", got, "onetwothree")
	}
}

// ---------------------------------------------------------------------------
// Pager: WriteToPager / HasPagerData / GetPagerData / ClearPager
// ---------------------------------------------------------------------------

func TestDescriptor_Pager_Lifecycle(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	d := NewDescriptor(server)

	// Initially no pager data.
	if d.HasPagerData() {
		t.Error("new descriptor should not have pager data")
	}
	if data := d.GetPagerData(); data != "" {
		t.Errorf("GetPagerData() = %q, want empty", data)
	}

	// Write to pager.
	d.WriteToPager("page 1\r\n")
	if !d.HasPagerData() {
		t.Error("should have pager data after WriteToPager")
	}

	d.WriteToPager("page 2\r\n")
	want := "page 1\r\npage 2\r\n"
	if got := d.GetPagerData(); got != want {
		t.Errorf("GetPagerData() = %q, want %q", got, want)
	}

	// PagePoint
	if d.GetPagePoint() != 0 {
		t.Errorf("initial page point = %d, want 0", d.GetPagePoint())
	}
	d.SetPagePoint(10)
	if d.GetPagePoint() != 10 {
		t.Errorf("page point = %d, want 10", d.GetPagePoint())
	}

	// PagerCmd
	if d.GetPagerCmd() != 0 {
		t.Errorf("initial pager cmd = %d, want 0", d.GetPagerCmd())
	}
	d.SetPagerCmd('c')
	if d.GetPagerCmd() != 'c' {
		t.Errorf("pager cmd = %c, want c", d.GetPagerCmd())
	}

	// ClearPager resets everything.
	d.ClearPager()
	if d.HasPagerData() {
		t.Error("should not have pager data after ClearPager")
	}
	if d.GetPagePoint() != 0 {
		t.Errorf("page point after clear = %d, want 0", d.GetPagePoint())
	}
	if d.GetPagerCmd() != 0 {
		t.Errorf("pager cmd after clear = %d, want 0", d.GetPagerCmd())
	}
}
