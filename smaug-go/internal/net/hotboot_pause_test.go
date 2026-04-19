//go:build !windows

package net

import (
	gonet "net"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/types"
)

// helper: start a server on 127.0.0.1:0 and return it + the bound port.
func startTestServer(t *testing.T) (*Server, int) {
	t.Helper()
	ln, err := gonet.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	s := NewServer()
	if err := s.StartOnListener(ln); err != nil {
		t.Fatalf("StartOnListener: %v", err)
	}
	port := ln.Addr().(*gonet.TCPAddr).Port
	return s, port
}

func TestPauseForHotboot_StopsAcceptStopsReadLoops(t *testing.T) {
	s, port := startTestServer(t)
	defer s.Stop()

	// Open one live client connection and wait for the server to pick it up.
	client, err := gonet.Dial("tcp", gonet.JoinHostPort("127.0.0.1", itoa(port)))
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer client.Close()

	var desc *types.DescriptorData
	select {
	case desc = <-s.Incoming:
	case <-time.After(1 * time.Second):
		t.Fatal("server never reported the incoming descriptor")
	}
	desc.Character = &types.CharData{Name: "Alice"}
	desc.Connected = types.CON_PLAYING

	// Pause.
	lnFile, descFiles, _, err := s.PauseForHotboot([]*types.DescriptorData{desc}, port)
	if err != nil {
		t.Fatalf("PauseForHotboot: %v", err)
	}
	if lnFile == nil {
		t.Fatal("lnFile nil")
	}
	if len(descFiles) != 1 {
		t.Fatalf("descFiles = %d, want 1", len(descFiles))
	}
	defer lnFile.Close()
	for _, f := range descFiles {
		defer f.Close()
	}

	// Prove the accept goroutine stopped: dial a second client and show it
	// does NOT end up in s.Incoming within a short window. (The kernel may
	// queue the SYN via the listener backlog; the test is that the Go
	// goroutine does not complete the Accept().)
	client2, err2 := gonet.DialTimeout("tcp", gonet.JoinHostPort("127.0.0.1", itoa(port)), 100*time.Millisecond)
	// Either dial fails (listener's accept goroutine gone → backlog may
	// still accept at kernel level but Accept() never runs) or succeeds
	// but never produces an Incoming event. We test the latter.
	if err2 == nil {
		defer client2.Close()
		select {
		case <-s.Incoming:
			t.Fatal("paused server delivered a second incoming; accept not halted")
		case <-time.After(150 * time.Millisecond):
			// good
		}
	}

	// Prove readLoops stopped: the channel created by the original
	// readLoop should be closed (defer close(desc.InputQueue) fires).
	select {
	case _, ok := <-desc.InputQueue:
		if ok {
			// Channel still open with data — acceptable if data was
			// already buffered; but nothing was sent. Expect close.
			t.Error("InputQueue had unexpected data after pause")
		}
		// channel closed — readLoop exited
	case <-time.After(500 * time.Millisecond):
		t.Error("InputQueue never closed after pause")
	}
}

func TestPauseForHotboot_ReturnsFDs(t *testing.T) {
	s, port := startTestServer(t)
	defer s.Stop()

	c1, err := gonet.Dial("tcp", gonet.JoinHostPort("127.0.0.1", itoa(port)))
	if err != nil {
		t.Fatal(err)
	}
	defer c1.Close()
	c2, err := gonet.Dial("tcp", gonet.JoinHostPort("127.0.0.1", itoa(port)))
	if err != nil {
		t.Fatal(err)
	}
	defer c2.Close()

	var descs []*types.DescriptorData
	for len(descs) < 2 {
		select {
		case d := <-s.Incoming:
			d.Character = &types.CharData{Name: "x"}
			d.Connected = types.CON_PLAYING
			descs = append(descs, d)
		case <-time.After(1 * time.Second):
			t.Fatalf("only got %d descriptors", len(descs))
		}
	}

	lnFile, descFiles, sessions, err := s.PauseForHotboot(descs, port)
	if err != nil {
		t.Fatalf("PauseForHotboot: %v", err)
	}
	defer lnFile.Close()
	for _, f := range descFiles {
		defer f.Close()
	}
	if lnFile == nil {
		t.Error("lnFile nil")
	}
	if len(descFiles) != 2 {
		t.Errorf("descFiles = %d, want 2", len(descFiles))
	}
	if len(sessions) != 2 {
		t.Errorf("sessions = %d, want 2", len(sessions))
	}
	for i, s := range sessions {
		if s.FDIndex != i {
			t.Errorf("session[%d].FDIndex = %d, want %d", i, s.FDIndex, i)
		}
		if s.Port != port {
			t.Errorf("session[%d].Port = %d, want %d", i, s.Port, port)
		}
	}
}

func TestResumeFromPause_RestartsAccept(t *testing.T) {
	s, port := startTestServer(t)
	defer s.Stop()

	c1, err := gonet.Dial("tcp", gonet.JoinHostPort("127.0.0.1", itoa(port)))
	if err != nil {
		t.Fatal(err)
	}
	defer c1.Close()

	var desc *types.DescriptorData
	select {
	case desc = <-s.Incoming:
	case <-time.After(1 * time.Second):
		t.Fatal("never got first desc")
	}
	desc.Character = &types.CharData{Name: "x"}
	desc.Connected = types.CON_PLAYING

	lnFile, descFiles, _, err := s.PauseForHotboot([]*types.DescriptorData{desc}, port)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range descFiles {
		defer f.Close()
	}

	// Resume — hands ownership of lnFile to the server.
	if err := s.ResumeFromPause(lnFile); err != nil {
		t.Fatalf("ResumeFromPause: %v", err)
	}

	// New connection should be accepted and delivered.
	c2, err := gonet.Dial("tcp", gonet.JoinHostPort("127.0.0.1", itoa(port)))
	if err != nil {
		t.Fatalf("post-resume Dial: %v", err)
	}
	defer c2.Close()
	select {
	case <-s.Incoming:
		// good
	case <-time.After(1 * time.Second):
		t.Error("no Incoming after ResumeFromPause")
	}
}

// itoa avoids pulling in strconv just for tests.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
