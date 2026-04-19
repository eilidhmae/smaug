//go:build !windows

// do_hotboot — initiate a copyover. Ports src/hotboot.c:599-743.
//
// High-level sequence (plan-phase6-hotboot.md §Chosen Design):
//   1. Gate: Level >= LEVEL_ASCENDANT, nobody fighting, nobody in
//      CON_EDITING, no hotboot already running.
//   2. Drop pre-login / mid-nanny descriptors with "Sorry, we are
//      rebooting, come back later.\r\n" and close them.
//   3. SaveWorld → <dataDir>/hotboot/mobfile.dat + per-room *.objdat.
//   4. For each CON_PLAYING descriptor: flip ch.PCData.Hotboot = true,
//      SavePlayer, flush output, emit farewell.
//   5. PauseForHotboot → returns listener FD, per-desc FDs, sessions.
//   6. SaveHotbootSessions → hotboot.dat.
//   7. Build argv with --hotboot-recover <ln_fd> <d_fd>:<idx> ...
//   8. execSelf → syscall.Exec(os.Executable(), argv, os.Environ()).
//
// On any error AFTER PauseForHotboot (including non-panic return from
// execSelf — which is impossible in production but reachable in tests),
// call ResumeFromPause to restart accept, clear HotbootInProgress, log,
// and return. Mirrors C's `dl_handle = dlopen(...)` fallback.

package act

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
)

// HotbootPauseFunc is set by boot to wire net.Server.PauseForHotboot
// without act importing internal/net (avoids the act→net dependency).
// Returns (listener file, per-desc files, sessions, err).
var HotbootPauseFunc func(descriptors []*types.DescriptorData, port int) (*os.File, []*os.File, []persist.HotbootSession, error)

// HotbootResumeFunc is the sibling seam — called on exec failure.
var HotbootResumeFunc func(lnFile *os.File) error

// HotbootPort carries the listener port so DoHotboot can stamp each
// session row. Set at boot.
var HotbootPort int

// execSelf is the syscall.Exec seam. Production never sees a return
// (syscall.Exec replaces the process image on success). Tests stub
// this to observe state just before the "pretend exec". A nil return
// is only reachable in tests; production code treats nil and error
// identically — calls ResumeFromPause, clears HotbootInProgress, logs,
// returns.
var execSelf = func(argv0 string, argv, envv []string) error {
	return syscall.Exec(argv0, argv, envv)
}

// clearCloexec removes the FD_CLOEXEC flag from an inherited FD so it
// survives syscall.Exec. Go's net.File() / TCPListener.File() set the
// flag by default; without clearing it the kernel closes the FD during
// exec and the hotboot recovery path sees an invalid descriptor. This
// is the F_SETFD 0 call referenced in plan-phase6-hotboot.md §Chosen
// Design step 5.
func clearCloexec(fd uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_FCNTL, fd, syscall.F_SETFD, 0)
	if errno != 0 {
		return errno
	}
	return nil
}

// hotbootNannyDropMsg is the exact string C writes to dropped mid-nanny
// descriptors. Match byte-for-byte for C-parity.
const hotbootNannyDropMsg = "\n\rSorry, we are rebooting, come back later.\n\r"

// hotbootFarewellMsg is emitted to each CON_PLAYING descriptor before
// the save+exec. C src/hotboot.c:707 equivalent.
const hotbootFarewellMsg = "\n\rYou feel a wrenching sensation in the pit of your stomach...\n\r"

// DoHotboot is the LEVEL_ASCENDANT immortal command that initiates a
// copyover. See file header for the full sequence.
func DoHotboot(ch *types.CharData, argument string) {
	if ch == nil {
		return
	}
	if WorldRef == nil {
		ch.Send("Hotboot unavailable: world not wired.\n\r")
		return
	}
	w := WorldRef

	// --- Gate 1: trust level ---
	if ch.GetTrust() < types.LEVEL_ASCENDANT {
		ch.Send("You are not sufficiently trusted.\n\r")
		return
	}

	// --- Gate 2: nobody in combat. Scan ALL descriptors, not just ch
	// (C src/hotboot.c:610-619 — the global descriptor walk). ---
	for _, d := range w.Descriptors {
		if d == nil || d.Character == nil {
			continue
		}
		if d.Character.Fighting != nil {
			ch.Send("A player is in combat.\n\r")
			return
		}
	}

	// --- Gate 3: nobody in the editor. C src/hotboot.c:619-626. ---
	// Extended beyond C to also refuse hotboot while any descriptor is
	// in an OLC substate (CON_REDIT / CON_OEDIT / CON_MEDIT). The Go
	// port's OLC sessions carry `d.Olc *OlcData` — a process-local
	// pointer to the room/obj/mob under edit plus an `EditorSave`
	// closure. Neither survives syscall.Exec (heap is replaced; func
	// pointers are process-scoped). If we let the hotboot proceed, the
	// player would land in CON_PLAYING in the child with their edits
	// silently discarded — worse than a clear "try again after you
	// finish" refusal. Per plan-phase6-hotboot.md §Chosen Design
	// (revised 2026-04-19 post-adversary review): refuse to hide OLC
	// session loss behind an invisible drop.
	for _, d := range w.Descriptors {
		if d == nil {
			continue
		}
		switch d.Connected {
		case types.CON_EDITING:
			ch.Send("A player is in the editor.\n\r")
			return
		case types.CON_REDIT, types.CON_OEDIT, types.CON_MEDIT:
			ch.Send("A player is in OLC. Try again once they finish.\n\r")
			return
		}
	}

	// --- Gate 4: concurrency guard. ---
	if w.SysData.HotbootInProgress {
		ch.Send("Hotboot already in progress.\n\r")
		return
	}
	w.SysData.HotbootInProgress = true

	// --- Drop mid-nanny / pre-login descriptors. C src/hotboot.c:668-680.
	// C-BUG DIVERGENCE (documented): C :675 writes the "Sorry, we are
	// rebooting" nanny message to ch->desc (the initiator's own
	// descriptor), spamming the initiator N times and leaving the
	// dropped descriptors silent. Go writes to d (the descriptor being
	// dropped) — matches the evident intent. plan-phase6-hotboot.md
	// §Chosen Design step 2. ---
	var kept []*types.DescriptorData
	for _, d := range w.Descriptors {
		if d == nil {
			continue
		}
		// Drop if not fully logged in. In Go's CON_* enum CON_PLAYING==0,
		// so "< CON_PLAYING" never fires; the Character==nil gate is
		// what catches mid-nanny. OLC states (CON_REDIT etc.) preserve
		// their Character pointer, so they stay through the drop.
		if d.Character == nil {
			if d.Conn != nil {
				d.Conn.Write([]byte(hotbootNannyDropMsg))
				d.Conn.Close()
			}
			continue
		}
		kept = append(kept, d)
	}
	w.Descriptors = kept

	// --- Save world state (NPCs + floor objects). ---
	if err := persist.SaveWorld(w, w.DataDir); err != nil {
		log.Printf("[hotboot] SaveWorld failed: %v", err)
		w.SysData.HotbootInProgress = false
		return
	}

	// --- Save each player. Session rows come from PauseForHotboot
	// below (it has direct access to the conn→FD mapping); we don't
	// build them here. This loop just flips the Hotboot flag on each
	// PC, saves their pfile, and emits the farewell. ---
	for _, d := range w.Descriptors {
		if d == nil || d.Character == nil {
			continue
		}
		if d.Connected != types.CON_PLAYING {
			continue
		}
		if d.Character.PCData != nil {
			d.Character.PCData.Hotboot = true
		}
		if SaveFunc != nil {
			SaveFunc(d.Character)
		}
		d.WriteToBuffer(hotbootFarewellMsg)
		_ = d.FlushOutput()
	}

	// --- Pause the server: stops accept, halts readLoops, returns
	// dup'd FDs we can pass to the exec'd child. Builds sessions. ---
	if HotbootPauseFunc == nil {
		log.Printf("[hotboot] HotbootPauseFunc not wired — refusing")
		w.SysData.HotbootInProgress = false
		return
	}
	lnFile, descFiles, sessions, err := HotbootPauseFunc(w.Descriptors, HotbootPort)
	if err != nil {
		log.Printf("[hotboot] PauseForHotboot failed: %v", err)
		w.SysData.HotbootInProgress = false
		return
	}

	// --- Persist session rows (so the child can correlate FDs with
	// pfiles + saved rooms). ---
	if err := persist.SaveHotbootSessions(w.DataDir, sessions); err != nil {
		log.Printf("[hotboot] SaveHotbootSessions failed: %v", err)
		// Attempt to resume; sessions file is unrecoverable but the
		// server can at least accept new connections. Close the
		// dup'd descriptor FDs (originals were already torn down by
		// PauseForHotboot) — each open *os.File here holds a kernel
		// FD that would otherwise leak until process exit.
		if HotbootResumeFunc != nil {
			if rerr := HotbootResumeFunc(lnFile); rerr != nil {
				log.Printf("[hotboot] ResumeFromPause also failed: %v", rerr)
			}
		}
		for _, f := range descFiles {
			f.Close()
		}
		w.SysData.HotbootInProgress = false
		return
	}

	// --- Build argv. ---
	exe, err := os.Executable()
	if err != nil {
		log.Printf("[hotboot] os.Executable: %v", err)
		if HotbootResumeFunc != nil {
			_ = HotbootResumeFunc(lnFile)
		}
		w.SysData.HotbootInProgress = false
		return
	}
	exeAbs, err := filepath.EvalSymlinks(exe)
	if err != nil {
		// Fall back to the un-resolved path; os.Executable() normally
		// already returns an absolute path.
		exeAbs = exe
	}

	// Clear FD_CLOEXEC on the listener and each descriptor FD so they
	// survive syscall.Exec. Go's (*os.File).Fd() / (*net.TCPListener).File()
	// set O_CLOEXEC by default (since Go 1.11), which means the kernel
	// would close them during the exec and the child's argv-passed FD
	// numbers would dangle. Mirrors the explicit F_SETFD 0 call in
	// plan-phase6-hotboot.md §Chosen Design step 5. Discovered by the
	// G6 integration test: prior versions of this path failed in the
	// recovered child with `FileListener: fcntl: bad file descriptor`.
	if err := clearCloexec(lnFile.Fd()); err != nil {
		log.Printf("[hotboot] clearCloexec(lnFD=%d): %v", lnFile.Fd(), err)
		if HotbootResumeFunc != nil {
			_ = HotbootResumeFunc(lnFile)
		}
		for _, f := range descFiles {
			f.Close()
		}
		w.SysData.HotbootInProgress = false
		return
	}
	for _, f := range descFiles {
		if err := clearCloexec(f.Fd()); err != nil {
			log.Printf("[hotboot] clearCloexec(descFD=%d): %v", f.Fd(), err)
			if HotbootResumeFunc != nil {
				_ = HotbootResumeFunc(lnFile)
			}
			for _, g := range descFiles {
				g.Close()
			}
			w.SysData.HotbootInProgress = false
			return
		}
	}

	lnFD := lnFile.Fd()
	argv := []string{
		exeAbs,
		"-port", fmt.Sprintf("%d", HotbootPort),
		"-data", w.DataDir,
		"--hotboot-recover", fmt.Sprintf("%d", lnFD),
	}
	for i, f := range descFiles {
		argv = append(argv, fmt.Sprintf("%d:%d", f.Fd(), i))
	}

	log.Printf("[hotboot] exec %s with %d session FDs", exeAbs, len(descFiles))

	// --- Exec. On success, never returns. On failure (or in test
	// stubbing), treat any return as a rollback trigger. ---
	if err := execSelf(exeAbs, argv, os.Environ()); err != nil {
		log.Printf("[hotboot] exec failed: %v", err)
	} else {
		// Nil return is only possible under test stubbing; production
		// cannot reach here.
		log.Printf("[hotboot] execSelf returned nil (test stub or impossible production path)")
	}

	// --- Rollback path. ---
	if HotbootResumeFunc != nil {
		if err := HotbootResumeFunc(lnFile); err != nil {
			log.Printf("[hotboot] ResumeFromPause failed: %v", err)
		}
	}
	// Close the dup'd descriptor FDs (they were duplicates; originals
	// were already torn down in PauseForHotboot).
	for _, f := range descFiles {
		f.Close()
	}
	w.SysData.HotbootInProgress = false
}
