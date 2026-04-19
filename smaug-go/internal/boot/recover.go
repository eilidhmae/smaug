//go:build !windows

// Hotboot recovery path. Ports src/hotboot.c:744-854 (hotboot_recover).
//
// BootRecover is the "hotboot-aware" sibling of Boot. It is invoked by
// cmd/smaug/main.go's argv-branch when `--hotboot-recover <ln_fd>
// <d1_fd>:<idx> ...` is present. Unlike Boot, BootRecover:
//
//   1. Does NOT bind a fresh listener — it inherits the old one via the
//      supplied lnFD (dup'd across syscall.Exec and not CLOEXEC'd).
//   2. Calls persist.LoadWorld after the normal bootDB, overlaying the
//      saved NPC/obj state onto the freshly-reset world. LoadWorld unlinks
//      mobfile.dat + *.objdat after parsing (C parity — prevents a crash
//      mid-recovery from looping).
//   3. Reads hotboot.dat independently (descFDSessions carries only FD +
//      session_idx) and re-pairs each inherited connection FD with the
//      saved player by name, loads the pfile, places in the saved room (or
//      ROOM_VNUM_TEMPLE fallback), and emits the "Time resumes its normal
//      flow" + puff-of-smoke welcome. Clears ch.PCData.Hotboot after the
//      welcome so next-save doesn't carry the stale flag.
//   4. Unlinks hotboot.dat (mobfile.dat and objdat already unlinked by
//      LoadWorld).
//
// The returned triple (*command.Registry, *game.GameLoop, *net.Server)
// mirrors Boot's contract plus a ready-to-run server that already holds
// the recovered descriptors.
//
// Build-tagged !windows because syscall.Exec + (*net.TCPListener).File()
// + net.FileConn are not meaningfully usable on Windows. Sibling
// recover_windows.go stub exists for GOOS=windows builds.

package boot

import (
	"bufio"
	"fmt"
	"log"
	gonet "net"
	"os"
	"path/filepath"

	"github.com/eilidhmae/smaug/internal/command"
	"github.com/eilidhmae/smaug/internal/game"
	smaugnet "github.com/eilidhmae/smaug/internal/net"
	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// FDSession pairs an inherited file-descriptor number with its session
// index in the child argv. The child parses argv tokens like "4:0" into
// {FD:4, SessionIdx:0}; BootRecover uses SessionIdx to correlate the FD
// with the matching HotbootSession row read from hotboot.dat.
type FDSession struct {
	FD         uintptr
	SessionIdx int
}

// hotbootWelcomePrefix is emitted to every resumed descriptor before the
// puff-of-smoke social. Matches C src/hotboot.c:827.
const hotbootWelcomePrefix = "\r\nTime resumes its normal flow.\r\n"

// hotbootPuffOfSmoke is the puff-of-smoke social text. Ported from the
// C inline `act(AT_MAGIC, "A puff of ethereal smoke dissipates around
// you!", ...)` pair at src/hotboot.c:838-839. The TO_CHAR line is what
// the recovered player sees; the TO_ROOM line is shown to other people in
// the room (not replicated here — the game loop re-emits it naturally as
// the character is placed via ordinary room-add logic).
const hotbootPuffOfSmoke = "A puff of ethereal smoke dissipates around you!\r\n"

// BootRecover is the hotboot-aware alternative to Boot. It reuses the
// shared bootDB pre-wire, then rehydrates saved NPC/obj state, wraps the
// inherited listener FD, and restores one descriptor per FDSession.
//
// The returned *net.Server is already running (accept loop spawned)
// against the inherited listener. Each recovered descriptor is added to
// w.Descriptors and has a readLoop goroutine feeding its InputQueue.
// Callers call gameLoop.Run(ctx) to drive pulses.
func BootRecover(
	w *world.World,
	dataDir string,
	opts BootOpts,
	lnFD uintptr,
	descFDSessions []FDSession,
) (*command.Registry, *game.GameLoop, *smaugnet.Server, error) {
	if w == nil {
		return nil, nil, nil, fmt.Errorf("BootRecover: nil world")
	}

	// --- Build the server + incoming channel up front so Boot's wiring
	// can reference the right chan. We will wrap the inherited listener
	// below and hand it to the server via StartOnListener. ---
	server := smaugnet.NewServer()

	// --- Shared Boot path: loads areas, wires callbacks, registers
	// commands. Same as a cold start; the world is a blank slate at this
	// point (in-memory — the pre-exec parent's world is gone). ---
	cmdReg, gameLoop, err := Boot(w, dataDir, server.Incoming, opts)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("BootRecover: Boot: %w", err)
	}

	// --- Overlay saved NPC + floor-object state on top of the fresh
	// area resets. LoadWorld unlinks mobfile.dat and each *.objdat after
	// a successful parse (C parity). ---
	if err := persist.LoadWorld(w, dataDir); err != nil {
		return nil, nil, nil, fmt.Errorf("BootRecover: LoadWorld: %w", err)
	}

	// --- Read the per-descriptor session rows. descFDSessions (from the
	// child argv) correlates an inherited FD with a row index here. ---
	sessions, err := persist.LoadHotbootSessions(dataDir)
	if err != nil {
		// Missing file is a hard error at this point — we already have
		// FDs in hand but no way to map them to pfiles. Log and fall
		// through: the caller still gets a running server, but inherited
		// FDs will be closed by the session loop below (nothing claims
		// them).
		util.Bug("BootRecover: LoadHotbootSessions: %v", err)
		sessions = nil
	}

	// --- Wrap the inherited listener FD. net.FileListener dups the fd
	// internally; the *os.File we pass in retains its own fd which we
	// close via lnFile.Close() after the wrap. ---
	lnFile := os.NewFile(lnFD, "hotboot-listener")
	if lnFile == nil {
		return cmdReg, gameLoop, nil, fmt.Errorf("BootRecover: os.NewFile(%d) returned nil", lnFD)
	}
	ln, err := gonet.FileListener(lnFile)
	if err != nil {
		lnFile.Close()
		return cmdReg, gameLoop, nil, fmt.Errorf("BootRecover: FileListener: %w", err)
	}
	// FileListener dups the fd; we can close our side of the *os.File.
	lnFile.Close()

	if err := server.StartOnListener(ln); err != nil {
		return cmdReg, gameLoop, server, fmt.Errorf("BootRecover: StartOnListener: %w", err)
	}

	// --- Restore each descriptor. Iterate descFDSessions (the source of
	// truth for which FDs were inherited); for each, look up the matching
	// hotboot.dat row by index. ---
	restored := 0
	for _, fs := range descFDSessions {
		if restoreDescriptor(w, gameLoop, server, dataDir, fs, sessions) {
			restored++
		}
	}
	log.Printf("[hotboot] recovery complete: %d of %d sessions restored", restored, len(descFDSessions))

	// --- Clean up the session manifest. mobfile.dat + *.objdat were
	// already unlinked inside LoadWorld; we own the hotboot.dat unlink. ---
	hotbootDat := filepath.Join(dataDir, "hotboot", "hotboot.dat")
	if err := os.Remove(hotbootDat); err != nil && !os.IsNotExist(err) {
		util.Bug("BootRecover: remove %s: %v", hotbootDat, err)
	}

	// --- Clear the global hotboot flag now that recovery is complete.
	// This was set to true pre-exec in the parent; the child inherits a
	// zero-valued SysData after fresh boot, so in practice this is a no-op
	// defensive reset (tests may have tampered with it). ---
	w.SysData.HotbootInProgress = false

	return cmdReg, gameLoop, server, nil
}

// restoreDescriptor rebuilds one descriptor from an inherited FD +
// matching hotboot.dat row. Returns true when the character is fully
// linked and welcomed; false on any failure along the way. Failures
// close the connection silently (C parity — src/hotboot.c:822-824).
// The bool return lets BootRecover emit an operator-facing "N of M
// restored" summary line without cluttering util.Bug with a noisy
// counter.
func restoreDescriptor(
	w *world.World,
	gameLoop *game.GameLoop,
	server *smaugnet.Server,
	dataDir string,
	fs FDSession,
	sessions []persist.HotbootSession,
) bool {
	// --- Wrap the inherited FD into a net.Conn. ---
	connFile := os.NewFile(fs.FD, fmt.Sprintf("hotboot-conn-%d", fs.SessionIdx))
	if connFile == nil {
		util.Bug("BootRecover: os.NewFile(%d) returned nil for session %d", fs.FD, fs.SessionIdx)
		return false
	}
	conn, err := gonet.FileConn(connFile)
	// FileConn dups the fd; we can close our side of the *os.File.
	connFile.Close()
	if err != nil {
		util.Bug("BootRecover: FileConn for session %d: %v", fs.SessionIdx, err)
		return false
	}

	// --- Look up the matching session row. ---
	if fs.SessionIdx < 0 || fs.SessionIdx >= len(sessions) {
		util.Bug("BootRecover: session idx %d out of range (have %d rows)", fs.SessionIdx, len(sessions))
		conn.Close()
		return false
	}
	sess := sessions[fs.SessionIdx]

	// --- Allocate descriptor. ---
	d := &types.DescriptorData{
		Conn:               conn,
		Host:               sess.Host,
		Port:               sess.Port,
		Connected:          types.CON_PLAYING,
		InputQueue:         make(chan string, 100),
		ScrLen:             24,
		TelnetState:        &types.TelnetState{Width: 80, Height: 24},
		ResumedFromHotboot: true,
	}
	d.ColorFunc = smaugnet.ProcessColors

	// --- Load the pfile. ---
	pfilePath := persist.PlayerFilePath(dataDir, sess.Name)
	if pfilePath == "" {
		util.Bug("BootRecover: invalid player name %q for session %d", sess.Name, fs.SessionIdx)
		conn.Close()
		return false
	}
	f, err := os.Open(pfilePath)
	if err != nil {
		// Missing pfile → close silently. C parity: src/hotboot.c:822-824
		// writes a "Somehow, your character was lost" message then
		// close_socket(d, FALSE). We omit the message (the plan required
		// silent-close + no-panic; matches the "closes connection" test).
		util.Bug("BootRecover: open pfile %s: %v", pfilePath, err)
		conn.Close()
		return false
	}
	ch, err := persist.LoadPlayerWithWorld(f, pfilePath, func(v int) *types.ObjIndexData {
		return w.GetObjIndex(v)
	})
	f.Close()
	if err != nil || ch == nil {
		util.Bug("BootRecover: LoadPlayer %s: %v", pfilePath, err)
		conn.Close()
		return false
	}

	// --- Place in saved room, or temple fallback (C parity — note C's
	// hotboot.c:829-830 falls to TEMPLE, not LIMBO, for PCs). ---
	room := w.GetRoom(sess.RoomVnum)
	if room == nil {
		util.Bug("BootRecover: saved room vnum %d not found for %s; falling back to TEMPLE",
			sess.RoomVnum, sess.Name)
		room = w.GetRoom(types.ROOM_VNUM_TEMPLE)
	}
	if room != nil {
		ch.InRoom = room
		room.People = append(room.People, ch)
	}

	// --- Link the descriptor to the character and vice-versa. ---
	ch.Desc = d
	d.Character = ch
	w.AddChar(ch)
	w.Descriptors = append(w.Descriptors, d)

	// --- Spawn read-loop (server-managed; matches fresh-connection
	// path in acceptLoop). ---
	go restoreReadLoop(d, conn)

	// --- Welcome. C src/hotboot.c:827 + 838 emit two lines: the flow
	// message goes straight to the descriptor via write_to_descriptor;
	// the puff-of-smoke comes from act(AT_MAGIC, "A puff of ethereal
	// smoke dissipates around you!", ..., TO_CHAR). ---
	d.WriteToBuffer(hotbootWelcomePrefix)
	d.WriteToBuffer(hotbootPuffOfSmoke)
	_ = d.FlushOutput()

	// --- Clear the hotboot flag now that the player is re-established.
	// Set by DoHotboot pre-exec and carried through the pfile; clearing
	// here prevents next-save from persisting a stale true. ---
	if ch.PCData != nil {
		ch.PCData.Hotboot = false
	}

	_ = gameLoop // reserved for future session-binding hooks.
	return true
}

// restoreReadLoop mirrors (*Server).readLoop's body — line-scan the conn
// and pipe into the descriptor's InputQueue. Duplicated here (rather than
// reusing the server's unexported method) so tests can observe the loop
// in isolation without coupling to the server goroutine lifecycle.
//
// When the scanner returns (EOF, scan error, or connection close), the
// InputQueue is closed so the game-loop's range-read unblocks.
func restoreReadLoop(d *types.DescriptorData, conn gonet.Conn) {
	defer close(d.InputQueue)

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1024), 1024)
	for scanner.Scan() {
		line := stripTelnet(scanner.Bytes())
		line = stripCarriageReturn(line)
		select {
		case d.InputQueue <- string(line):
		default:
			// Full queue — drop. Production buffer is 100; overflow
			// during recovery means a misbehaving client.
			log.Printf("[hotboot-recover] InputQueue full on %s; dropping line", d.Host)
		}
	}
}

// stripTelnet is a local copy of internal/net's stripTelnetIAC. Kept
// private to the boot package so we don't take a runtime dep on
// internal/net's unexported symbols.
func stripTelnet(data []byte) []byte {
	const iac byte = 255
	out := make([]byte, 0, len(data))
	i := 0
	for i < len(data) {
		if data[i] == iac && i+1 < len(data) {
			cmd := data[i+1]
			switch {
			case cmd >= 251 && cmd <= 254:
				if i+2 < len(data) {
					i += 3
				} else {
					i = len(data)
				}
			case cmd == iac:
				out = append(out, iac)
				i += 2
			default:
				i += 2
			}
			continue
		}
		out = append(out, data[i])
		i++
	}
	return out
}

func stripCarriageReturn(data []byte) []byte {
	out := make([]byte, 0, len(data))
	for _, b := range data {
		if b != '\r' {
			out = append(out, b)
		}
	}
	return out
}
