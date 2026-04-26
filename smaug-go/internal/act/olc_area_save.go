package act

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// Plan plan-phase6-foldarea.md §G1 + §G2 + §G3.
//
// resolveAreaFilePath is the path-containment guard formerly inlined at the
// top of DoSaveArea (olc.go:817-841 prior to extraction). It rejects empty
// filenames, absolute paths, and any cleaned path that escapes
// <dataDir>/area/ via "..". On success it returns the cleaned absolute path
// where the area file should live.
//
// dataDir is normally WorldRef.DataDir; passed explicitly so tests can
// exercise the helper without standing up a full world.
func resolveAreaFilePath(dataDir, filename string) (string, error) {
	if filename == "" {
		return "", errors.New("area filename is empty")
	}
	if filepath.IsAbs(filename) {
		return "", fmt.Errorf("area filename must be relative: %q", filename)
	}
	areaDir := filepath.Clean(filepath.Join(dataDir, "area"))
	cleaned := filepath.Clean(filepath.Join(areaDir, filename))
	base := areaDir + string(filepath.Separator)
	// Allow exactly the area dir prefix; reject anything that escapes.
	if !strings.HasPrefix(cleaned+string(filepath.Separator), base) {
		return "", fmt.Errorf("area filename escapes area dir: %q", filename)
	}
	return cleaned, nil
}

// writeAreaToDisk is the shared save helper used by both DoSaveArea and
// DoFoldarea (Wave 3). It performs:
//
//  1. Filename + path-containment validation via resolveAreaFilePath.
//  2. Write to <path>.tmp via persist.SaveArea.
//  3. .bak rotation: if <path> exists, rename it to <path>.bak before
//     installing the new tmp file. Mirrors C fold_area at
//     src/build.c:7369-7370. Best-effort — rotation failure is logged
//     but does not abort the save.
//  4. Atomic rename <path>.tmp → <path>.
//
// Returns nil on success, or an error suitable for surfacing to the user.
// The caller decides how to phrase the user-facing message.
func writeAreaToDisk(area *types.AreaData) error {
	if area == nil {
		return errors.New("area is nil")
	}
	if area.Filename == "" {
		return errors.New("area has no filename")
	}
	path, err := resolveAreaFilePath(WorldRef.DataDir, area.Filename)
	if err != nil {
		util.Bug("writeAreaToDisk: rejected filename %q: %v", area.Filename, err)
		return errors.New("invalid area filename")
	}
	tmpPath := path + ".tmp"

	// 1. Write to tmp file.
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create tmp: %w", err)
	}
	if err := persist.SaveArea(f, WorldRef, area); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write area: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close tmp: %w", err)
	}

	// 2. .bak rotation. Rotate-if-exists (a brand-new area's first save has
	//    no live file to rotate). Failure is logged via util.Bug but does
	//    NOT abort the save — the tmp file is the source of truth and the
	//    operator can recover from disk via the tmp→live rename below.
	//    Mirrors C fold_area at src/build.c:7369-7370.
	if _, statErr := os.Stat(path); statErr == nil {
		bakPath := path + ".bak"
		if rerr := os.Rename(path, bakPath); rerr != nil {
			util.Bug("writeAreaToDisk: .bak rotation failed for %q: %v", path, rerr)
			// fall through; tmp→live rename below will overwrite the live file.
		}
	}

	// 3. Atomic install.
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("install: %w", err)
	}
	return nil
}
