//go:build !windows

// Hotboot per-descriptor session rows. Ports the inline save loop in
// src/hotboot.c:672-740 (walks first_descriptor, writes one line per
// live player to hotboot.dat) and the corresponding inline read in
// src/hotboot.c:744-842.
//
// Format: one line per session, tab-delimited:
//
//     <fd_index>\t<room_vnum>\t<port>\t<idle_ticks>\t<ansi>\t<name>\t<host>\n
//
// Terminated by a `$` sentinel on its own line. Fields are read in a
// single strings.Split on `\t`; name/host are tolerated as-is after
// SmashTilde on the host at write time. If the file has no `$` sentinel
// (truncation mid-write), whatever parsed successfully is returned.
//
// C parity note: the C writer emits a space-separated single line with
// "%d %d %d %d %s %s" and closes with a blank line terminator. We use
// tabs + `$` sentinel to avoid name/host whitespace collisions without
// having to match C byte-exact — the hotboot.dat file is produced and
// consumed by the same Go binary in the same process lineage.

package persist

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/util"
)

// HotbootSession describes one descriptor's persistence row across a
// hotboot. FDIndex is the position of the corresponding FD in the child
// argv (see plan §Chosen Design step 6).
type HotbootSession struct {
	FDIndex   int    // position of this FD in the child argv
	RoomVnum  int    // character's room at save time
	Port      int    // listener port (same for all sessions, kept for symmetry)
	IdleTicks int    // CharData idle counter
	Ansi      bool   // d->can_compress-equivalent flag
	Name      string // character name (for logging only; recovery reads the pfile)
	Host      string // connecting host
}

// SaveHotbootSessions writes <dir>/hotboot/hotboot.dat. Auto-creates
// the hotboot directory. Tildes in Host are smashed to dashes.
func SaveHotbootSessions(dir string, sessions []HotbootSession) error {
	hotbootDir := filepath.Join(dir, "hotboot")
	if err := os.MkdirAll(hotbootDir, 0o755); err != nil {
		return fmt.Errorf("SaveHotbootSessions: mkdir: %w", err)
	}
	path := filepath.Join(hotbootDir, "hotboot.dat")
	// 0o600 — hotboot.dat carries player names + hosts; not
	// world-readable.
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("SaveHotbootSessions: create: %w", err)
	}
	defer f.Close()
	bw := bufio.NewWriter(f)
	for _, s := range sessions {
		ansi := 0
		if s.Ansi {
			ansi = 1
		}
		// Tab-delimited. SmashTilde on Host defends the format.
		fmt.Fprintf(bw, "%d\t%d\t%d\t%d\t%d\t%s\t%s\n",
			s.FDIndex, s.RoomVnum, s.Port, s.IdleTicks, ansi,
			util.SmashTilde(s.Name), util.SmashTilde(s.Host))
	}
	// Sentinel — guards against mid-truncation.
	fmt.Fprintf(bw, "$\n")
	return bw.Flush()
}

// LoadHotbootSessions reads <dir>/hotboot/hotboot.dat. Missing file
// returns (nil, error) with os.IsNotExist(err)==true. Malformed lines
// are logged via util.Bug and skipped; missing sentinel logs and returns
// whatever parsed successfully.
func LoadHotbootSessions(dir string) ([]HotbootSession, error) {
	path := filepath.Join(dir, "hotboot", "hotboot.dat")
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []HotbootSession
	sawSentinel := false
	sc := bufio.NewScanner(f)
	// Allow long lines (hostnames, character names can push default 64K).
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if line == "$" {
			sawSentinel = true
			break
		}
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 7 {
			util.Bug("LoadHotbootSessions: %s: malformed line %q (fields=%d)",
				path, line, len(parts))
			continue
		}
		fdIdx, err1 := strconv.Atoi(parts[0])
		room, err2 := strconv.Atoi(parts[1])
		port, err3 := strconv.Atoi(parts[2])
		idle, err4 := strconv.Atoi(parts[3])
		ansiVal, err5 := strconv.Atoi(parts[4])
		if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
			util.Bug("LoadHotbootSessions: %s: bad numeric fields in %q", path, line)
			continue
		}
		out = append(out, HotbootSession{
			FDIndex:   fdIdx,
			RoomVnum:  room,
			Port:      port,
			IdleTicks: idle,
			Ansi:      ansiVal != 0,
			Name:      parts[5],
			Host:      parts[6],
		})
	}
	if err := sc.Err(); err != nil {
		return out, fmt.Errorf("LoadHotbootSessions: scan: %w", err)
	}
	if !sawSentinel {
		util.Bug("LoadHotbootSessions: %s: missing $ sentinel", path)
	}
	return out, nil
}
