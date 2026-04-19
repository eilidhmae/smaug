//go:build !windows

// Hotboot pause/resume seam. Ports the subset of src/hotboot.c:599-743
// responsible for stopping new connections and active read-loops without
// closing the underlying sockets, so the FDs can be inherited across
// syscall.Exec. Separate file (not server.go) for build-tag containment
// per plan-phase6-hotboot.md §Platform.

package net

import (
	"fmt"
	gonet "net"
	"os"

	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
)

// PauseForHotboot stops the accept goroutine, halts readLoops on every
// active descriptor, and returns the listener FD plus per-descriptor FDs
// as *os.File handles without closing the underlying sockets. Callers
// own the returned files and must either syscall.Exec with them
// inherited, or call ResumeFromPause to restore full operation.
//
// The descriptors argument is the live descriptor slice (world.Descriptors).
// Session rows are built in parallel — the returned []persist.HotbootSession
// has one entry per non-skipped descriptor (Character != nil && Connected
// == CON_PLAYING), with FDIndex matching the position in the returned FD
// slice. Pre-login / mid-nanny descriptors are skipped entirely (caller
// drops them before calling PauseForHotboot).
//
// On error, any files already dup'd are closed and the accept goroutine
// is NOT restarted — the caller should invoke ResumeFromPause anyway to
// reset internal state.
func (s *Server) PauseForHotboot(descriptors []*types.DescriptorData, port int) (*os.File, []*os.File, []persist.HotbootSession, error) {
	s.doneMu.Lock()
	select {
	case <-s.done:
		// Already stopped (previous pause still in effect).
	default:
		close(s.done)
	}
	s.doneMu.Unlock()

	// --- Dup the listener FD. File() duplicates the fd; closing the
	// original listener will NOT close the returned file. ---
	tcpLn, ok := s.listener.(*gonet.TCPListener)
	if !ok {
		return nil, nil, nil, fmt.Errorf("PauseForHotboot: listener is %T, need *net.TCPListener", s.listener)
	}
	lnFile, err := tcpLn.File()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("PauseForHotboot: listener.File: %w", err)
	}
	// Close the original listener to wake acceptLoop out of Accept().
	s.listener.Close()

	// --- Dup each descriptor's connection FD and build session rows. ---
	var descFiles []*os.File
	var sessions []persist.HotbootSession
	cleanup := func() {
		lnFile.Close()
		for _, f := range descFiles {
			f.Close()
		}
	}
	for _, d := range descriptors {
		if d == nil || d.Character == nil || d.Connected != types.CON_PLAYING {
			continue
		}
		tcpConn, ok := d.Conn.(*gonet.TCPConn)
		if !ok {
			cleanup()
			return nil, nil, nil, fmt.Errorf("PauseForHotboot: descriptor conn is %T, need *net.TCPConn", d.Conn)
		}
		f, err := tcpConn.File()
		if err != nil {
			cleanup()
			return nil, nil, nil, fmt.Errorf("PauseForHotboot: conn.File: %w", err)
		}
		// Close original conn to wake readLoop out of Scan(). The dup
		// in f survives — that's what the child will inherit across exec.
		tcpConn.Close()
		roomVnum := 0
		if d.Character.InRoom != nil {
			roomVnum = d.Character.InRoom.Vnum
		}
		ansi := d.Character.Act.IsSet(types.PLR_ANSI)
		sessions = append(sessions, persist.HotbootSession{
			FDIndex:   len(descFiles),
			RoomVnum:  roomVnum,
			Port:      port,
			IdleTicks: d.Idle,
			Ansi:      ansi,
			Name:      d.Character.Name,
			Host:      d.Host,
		})
		descFiles = append(descFiles, f)
	}
	return lnFile, descFiles, sessions, nil
}

// ResumeFromPause restarts the accept goroutine after a failed hotboot
// attempt. It re-wraps the previously-dup'd listener FD via net.FileListener
// so new connections resume. Pre-pause descriptor connections are NOT
// restarted — they closed when PauseForHotboot closed the originals.
// In practice, if DoHotboot gets as far as PauseForHotboot and the
// subsequent syscall.Exec fails, the MUD is in a degraded state: accept
// resumes for NEW clients but existing PLAYING-state connections are
// gone. This mirrors C's fallback where dlclose + dl_open fallback
// leaves dropped descriptors as-is.
func (s *Server) ResumeFromPause(lnFile *os.File) error {
	if lnFile == nil {
		return fmt.Errorf("ResumeFromPause: nil listener file")
	}
	// Wait for the prior acceptLoop goroutine (started by
	// StartOnListener, woken by PauseForHotboot's listener.Close()) to
	// fully exit. Without this wait the next-line rewrite of
	// s.listener would race the old goroutine's pending Accept() read
	// of s.listener — go test -race catches this deterministically.
	s.acceptWG.Wait()

	ln, err := gonet.FileListener(lnFile)
	if err != nil {
		return fmt.Errorf("ResumeFromPause: FileListener: %w", err)
	}
	// FileListener dups the fd again and wraps that; we can close
	// the original *os.File — the new listener owns its own dup.
	lnFile.Close()

	s.doneMu.Lock()
	s.done = make(chan struct{})
	s.doneMu.Unlock()
	s.listener = ln
	s.acceptWG.Add(1)
	go s.acceptLoop()
	return nil
}
