package act

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- Wave 1 (G1): resolveAreaFilePath helper tests ---
//
// Plan plan-phase6-foldarea.md §G1 + §A1. The helper encapsulates the
// path-containment guard previously inlined in DoSaveArea (olc.go:817-841).
// Tests pin the four reject cases and the two accept cases.

func TestResolveAreaFilePath_Empty(t *testing.T) {
	WorldRef = nil
	dir := t.TempDir()
	if _, err := resolveAreaFilePath(dir, ""); err == nil {
		t.Fatal("expected error for empty filename, got nil")
	}
}

func TestResolveAreaFilePath_Absolute(t *testing.T) {
	dir := t.TempDir()
	if _, err := resolveAreaFilePath(dir, "/etc/passwd"); err == nil {
		t.Fatal("expected error for absolute path, got nil")
	}
}

func TestResolveAreaFilePath_Traversal(t *testing.T) {
	dir := t.TempDir()
	if _, err := resolveAreaFilePath(dir, "../../etc/passwd"); err == nil {
		t.Fatal("expected error for traversal, got nil")
	}
}

func TestResolveAreaFilePath_Valid(t *testing.T) {
	dir := t.TempDir()
	got, err := resolveAreaFilePath(dir, "myzone.are")
	if err != nil {
		t.Fatalf("expected success, got err=%v", err)
	}
	want := filepath.Clean(filepath.Join(dir, "area", "myzone.are"))
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolveAreaFilePath_NestedSubpath(t *testing.T) {
	dir := t.TempDir()
	got, err := resolveAreaFilePath(dir, "sub/x.are")
	if err != nil {
		t.Fatalf("expected success for nested path, got err=%v", err)
	}
	want := filepath.Clean(filepath.Join(dir, "area", "sub", "x.are"))
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// --- Wave 1 (G2): writeAreaToDisk helper tests ---
//
// Plan §G2 + §A2. The shared save helper performs filename check + path
// containment + atomic tmp→live write. .bak rotation arrives in G3 (Wave 2);
// these G2 tests pin the pre-bak baseline so the G3 commit is a clean diff.

func TestWriteAreaToDisk_HappyPath(t *testing.T) {
	ch, _, area, tmpDir, cleanup := setupSaveAreaWorld(t)
	defer cleanup()
	_ = ch

	if err := writeAreaToDisk(area); err != nil {
		t.Fatalf("writeAreaToDisk failed: %v", err)
	}
	path := filepath.Join(tmpDir, "area", area.Filename)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected area file at %s: %v", path, err)
	}
	// Tmp must be cleaned up on success.
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("expected no .tmp after success, stat err=%v", err)
	}
}

func TestWriteAreaToDisk_NoFilename(t *testing.T) {
	_, _, area, _, cleanup := setupSaveAreaWorld(t)
	defer cleanup()

	area.Filename = ""
	err := writeAreaToDisk(area)
	if err == nil {
		t.Fatal("expected error for empty filename, got nil")
	}
	if !strings.Contains(err.Error(), "filename") {
		t.Errorf("expected 'filename' in error, got: %v", err)
	}
}

func TestWriteAreaToDisk_BadFilename(t *testing.T) {
	_, _, area, tmpDir, cleanup := setupSaveAreaWorld(t)
	defer cleanup()

	area.Filename = "../escape.are"
	err := writeAreaToDisk(area)
	if err == nil {
		t.Fatal("expected error for traversal filename, got nil")
	}
	// No file should have been created at the escape target.
	if _, err := os.Stat(filepath.Join(tmpDir, "..", "escape.are")); err == nil {
		t.Errorf("traversal succeeded — file written outside area dir")
	}
}

// --- Wave 2 (G3): .bak rotation tests ---
//
// Plan plan-phase6-foldarea.md §D1 + §G3 + §A3-A5. Mirrors C fold_area
// at src/build.c:7369-7370. Behavior:
//   - First save (no live file): no .bak created.
//   - Second save: live file rotated to <file>.bak; new content installed.
//   - Pre-existing .bak silently overwritten (POSIX rename semantics).
//   - Rotation failure logged via util.Bug, save still proceeds.

func TestWriteAreaToDisk_FirstSaveNoBak(t *testing.T) {
	_, _, area, tmpDir, cleanup := setupSaveAreaWorld(t)
	defer cleanup()

	if err := writeAreaToDisk(area); err != nil {
		t.Fatalf("writeAreaToDisk failed: %v", err)
	}
	livePath := filepath.Join(tmpDir, "area", area.Filename)
	bakPath := livePath + ".bak"
	if _, err := os.Stat(livePath); err != nil {
		t.Fatalf("expected live file at %s: %v", livePath, err)
	}
	if _, err := os.Stat(bakPath); !os.IsNotExist(err) {
		t.Errorf("expected NO .bak after first save (no prior live file), stat err=%v", err)
	}
}

func TestWriteAreaToDisk_SecondSaveCreatesBak(t *testing.T) {
	_, _, area, tmpDir, cleanup := setupSaveAreaWorld(t)
	defer cleanup()

	livePath := filepath.Join(tmpDir, "area", area.Filename)
	bakPath := livePath + ".bak"

	// Pre-seed a "v1" live file so the rotation has something to move.
	v1Marker := []byte("#AREA   v1-marker~\n\n\n#$\n")
	if err := os.WriteFile(livePath, v1Marker, 0o644); err != nil {
		t.Fatalf("seed v1: %v", err)
	}

	if err := writeAreaToDisk(area); err != nil {
		t.Fatalf("writeAreaToDisk failed: %v", err)
	}

	// .bak must contain v1 marker.
	bakContent, err := os.ReadFile(bakPath)
	if err != nil {
		t.Fatalf("expected .bak file at %s: %v", bakPath, err)
	}
	if !strings.Contains(string(bakContent), "v1-marker") {
		t.Errorf(".bak content does not contain v1 marker: %q", string(bakContent))
	}

	// Live file must NOT contain v1 marker (it's the new save).
	liveContent, err := os.ReadFile(livePath)
	if err != nil {
		t.Fatalf("expected live file: %v", err)
	}
	if strings.Contains(string(liveContent), "v1-marker") {
		t.Errorf("live file still contains v1 marker — rotation order wrong: %q", string(liveContent))
	}
	// Live file must contain the area name (proof persist.SaveArea wrote real content).
	if !strings.Contains(string(liveContent), area.Name) {
		t.Errorf("live file missing area name %q: %q", area.Name, string(liveContent))
	}
}

func TestWriteAreaToDisk_BakOverwritePreservesNewest(t *testing.T) {
	_, _, area, tmpDir, cleanup := setupSaveAreaWorld(t)
	defer cleanup()

	livePath := filepath.Join(tmpDir, "area", area.Filename)
	bakPath := livePath + ".bak"

	// Pre-seed both: an OLD .bak (from a previous save) and a v1 live file.
	if err := os.WriteFile(bakPath, []byte("vold-bak-marker"), 0o644); err != nil {
		t.Fatalf("seed old bak: %v", err)
	}
	if err := os.WriteFile(livePath, []byte("v1-live-marker"), 0o644); err != nil {
		t.Fatalf("seed v1: %v", err)
	}

	if err := writeAreaToDisk(area); err != nil {
		t.Fatalf("writeAreaToDisk failed: %v", err)
	}

	// .bak should now contain v1 (live was rotated). vold is gone.
	bakContent, err := os.ReadFile(bakPath)
	if err != nil {
		t.Fatalf("expected .bak: %v", err)
	}
	if !strings.Contains(string(bakContent), "v1-live-marker") {
		t.Errorf(".bak should contain v1 (rotated from live), got: %q", string(bakContent))
	}
	if strings.Contains(string(bakContent), "vold-bak-marker") {
		t.Errorf(".bak still contains vold marker — overwrite did not happen: %q", string(bakContent))
	}
}

func TestSaveArea_RoundTripWithBak(t *testing.T) {
	ch, client, area, tmpDir, cleanup := setupSaveAreaWorld(t)
	defer cleanup()

	livePath := filepath.Join(tmpDir, "area", area.Filename)
	bakPath := livePath + ".bak"

	// First save.
	DoSaveArea(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(strings.ToLower(out), "saved") {
		t.Fatalf("expected first save success, got: %q", out)
	}
	if _, err := os.Stat(bakPath); !os.IsNotExist(err) {
		t.Errorf("first save: expected NO .bak yet, stat err=%v", err)
	}

	// Capture v1 content.
	v1, err := os.ReadFile(livePath)
	if err != nil {
		t.Fatalf("read v1: %v", err)
	}

	// Second save: mutate area name so we can distinguish v1 vs v2.
	area.Name = "Test Area v2"
	DoSaveArea(ch, "")
	_ = readOutput(ch, client)

	bakContent, err := os.ReadFile(bakPath)
	if err != nil {
		t.Fatalf("expected .bak after second save: %v", err)
	}
	if string(bakContent) != string(v1) {
		t.Errorf(".bak content does not match v1 byte-for-byte\n  bak=%q\n   v1=%q", string(bakContent), string(v1))
	}
	liveContent, err := os.ReadFile(livePath)
	if err != nil {
		t.Fatalf("read live after second save: %v", err)
	}
	if !strings.Contains(string(liveContent), "Test Area v2") {
		t.Errorf("live file missing v2 marker after second save: %q", string(liveContent))
	}
}
