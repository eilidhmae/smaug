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
