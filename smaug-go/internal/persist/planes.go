package persist

import (
	"fmt"
	"os"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// LoadPlanes reads db/system/planes.dat into a slice. Three outcomes the
// caller must distinguish:
//
//   - Missing file: returns `(nil, nil)` with no bug log. First-boot case;
//     CheckPlanes will seed "Prime Material".
//   - Empty but valid file (e.g. `#END\n` stub): returns a non-nil
//     zero-length slice, nil err. Data-tree-present-but-no-planes case;
//     CheckPlanes still needs to seed a default because len==0.
//   - Populated file: returns one *PlaneData per `#PLANE`/`End` block.
//
// Mirrors C `load_planes` at src/planes.c:232-269 and `read_plane` at
// :190-230. Unlike C, this loader is tilde-tolerant: the `Name` value may
// be either tilde-terminated (SMAUG convention; what SavePlanes emits) or
// space/newline-terminated (what C save_planes actually writes — a latent
// C format mismatch per src/planes.c:181). Both round-trip cleanly.
//
// Malformed blocks log via util.Bug and continue recovering where
// possible, matching the C "fread_to_eol on unknown field" convention.
func LoadPlanes(path string) ([]*types.PlaneData, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	sc := NewScanner(f, path)
	list := make([]*types.PlaneData, 0)
	for {
		c := sc.ReadLetter()
		if c == 0 {
			// EOF without #END — return whatever we have.
			return list, nil
		}
		if c != '#' {
			util.Bug("LoadPlanes: %s: expected '#', got %q", path, c)
			return list, nil
		}
		word := sc.ReadWord()
		switch word {
		case "END":
			return list, nil
		case "PLANE":
			if p := readPlaneBlock(sc, path); p != nil {
				list = append(list, p)
			}
		default:
			util.Bug("LoadPlanes: %s: unknown section %q", path, word)
			return list, nil
		}
	}
}

// readPlaneBlock reads one `#PLANE` block up to `End`. Returns nil when
// the block has no Name (matches intent of C `read_plane` which accepts
// but would yield a zero-name plane — Go rejects with a bug log).
func readPlaneBlock(sc *Scanner, path string) *types.PlaneData {
	p := &types.PlaneData{}
	for {
		word := sc.ReadWord()
		if word == "" {
			util.Bug("LoadPlanes: %s: unexpected EOF inside #PLANE block", path)
			if p.Name == "" {
				return nil
			}
			return p
		}
		switch word {
		case "End":
			if p.Name == "" {
				util.Bug("LoadPlanes: %s: #PLANE block has no Name", path)
				return nil
			}
			return p
		case "Name":
			// Tilde-tolerant: read the rest of the line, strip whitespace
			// and any trailing tilde. Accepts both C-format
			// (`Name      Astral\n`) and SMAUG-format (`Name      Astral~\n`).
			line := sc.ReadToEOL()
			line = strings.TrimSpace(line)
			line = strings.TrimSuffix(line, "~")
			line = strings.TrimSpace(line)
			p.Name = line
		default:
			util.Bug("LoadPlanes: %s: unknown key %q in #PLANE block", path, word)
			sc.ReadToEOL()
		}
	}
}

// SavePlanes truncate-writes planes to path in the SMAUG #PLANE block
// format. Emits one `#PLANE`/`Name <name>~`/`End` block per plane, then
// a terminating `#END\n`. String fields are tilde-terminated, which
// diverges from C `save_planes` (src/planes.c:181 writes no tilde) but
// converges with SMAUG's other string-field convention (clans, deities,
// holidays). LoadPlanes accepts both formats.
//
// Failure returns the error after attempting to close the file. Caller
// is expected to log via util.Bug — the command handler DoPset emits
// "Planes saved." unconditionally to match C's no-error-to-user
// convention.
func SavePlanes(path string, list []*types.PlaneData) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, p := range list {
		if p == nil {
			continue
		}
		if _, err := fmt.Fprintf(f, "#PLANE\nName      %s~\nEnd\n\n", p.Name); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(f, "#END"); err != nil {
		return err
	}
	return nil
}

// CheckPlanes ensures at least one plane exists and reassigns every room
// whose Plane pointer is nil (or points at the about-to-be-deleted
// `deleted` argument) to the first plane. Mirrors C `check_planes` at
// src/planes.c:283-298. Called at boot (after LoadPlanes) and inside
// DoPset delete (before unlinking the deleted plane).
//
// If w.Planes is empty, a "Prime Material" plane is appended (matches C
// `build_prime_plane` at src/planes.c:271-281).
//
// Passing `deleted = nil` is valid (boot path); every nil-Plane room
// gets first_plane. Passing a non-nil `deleted` reassigns rooms that
// already pointed at that plane — DoPset delete calls with the slice
// already spliced, so `w.Planes[0]` is guaranteed to be a different
// pointer than the freed one.
//
// Room iteration uses Go's randomized map walk. That is safe here
// because every room is assigned the same value; if a future change
// scopes assignment by vnum range, switch to a sorted-key iteration.
func CheckPlanes(w *world.World, deleted *types.PlaneData) {
	if w == nil {
		return
	}
	if len(w.Planes) == 0 {
		w.Planes = append(w.Planes, &types.PlaneData{Name: "Prime Material"})
	}
	firstPlane := w.Planes[0]
	for _, room := range w.Rooms {
		if room == nil {
			continue
		}
		if room.Plane == nil || room.Plane == deleted {
			room.Plane = firstPlane
		}
	}
}
