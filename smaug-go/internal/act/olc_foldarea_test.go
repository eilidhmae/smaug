package act

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// --- Wave 3 (G4): DoFoldarea tests ---
//
// Plan plan-phase6-foldarea.md §D2 + §G4 + §A6-A8. Mirrors C do_foldarea
// at src/build.c:8055-8081.

func TestDoFoldarea_NoArgument(t *testing.T) {
	ch, client, _, _, cleanup := setupSaveAreaWorld(t)
	defer cleanup()

	DoFoldarea(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Fold what?") {
		t.Errorf("expected 'Fold what?' for empty arg, got: %q", out)
	}
}

func TestDoFoldarea_NoSuchArea(t *testing.T) {
	ch, client, _, _, cleanup := setupSaveAreaWorld(t)
	defer cleanup()

	DoFoldarea(ch, "ghost.are")
	out := readOutput(ch, client)
	if !strings.Contains(out, "No such area exists") {
		t.Errorf("expected 'No such area exists.', got: %q", out)
	}
}

func TestDoFoldarea_HappyPath(t *testing.T) {
	ch, client, area, tmpDir, cleanup := setupSaveAreaWorld(t)
	defer cleanup()

	DoFoldarea(ch, area.Filename)
	out := readOutput(ch, client)

	if !strings.Contains(out, "Folding area") {
		t.Errorf("expected 'Folding area...', got: %q", out)
	}
	if !strings.Contains(out, "Done") {
		t.Errorf("expected 'Done', got: %q", out)
	}
	livePath := filepath.Join(tmpDir, "area", area.Filename)
	if _, err := os.Stat(livePath); err != nil {
		t.Fatalf("expected live file at %s: %v", livePath, err)
	}
}

func TestDoFoldarea_CaseInsensitive(t *testing.T) {
	ch, client, area, _, cleanup := setupSaveAreaWorld(t)
	defer cleanup()

	area.Filename = "MyZone.are"
	DoFoldarea(ch, "myzone.are")
	out := readOutput(ch, client)

	if strings.Contains(out, "No such area") {
		t.Errorf("expected case-insensitive match, got: %q", out)
	}
	if !strings.Contains(out, "Done") {
		t.Errorf("expected 'Done' on success, got: %q", out)
	}
}

func TestDoFoldarea_TrustGate(t *testing.T) {
	ch, client, area, _, cleanup := setupSaveAreaWorld(t)
	defer cleanup()

	// Drop the char's trust below LEVEL_IMMORTAL.
	ch.Level = 1
	ch.Trust = 0
	DoFoldarea(ch, area.Filename)
	out := readOutput(ch, client)

	if !strings.Contains(out, "Huh?") {
		t.Errorf("expected 'Huh?' for non-immortal, got: %q", out)
	}
}

func TestDoFoldarea_BakRotationApplies(t *testing.T) {
	ch, client, area, tmpDir, cleanup := setupSaveAreaWorld(t)
	defer cleanup()

	livePath := filepath.Join(tmpDir, "area", area.Filename)
	bakPath := livePath + ".bak"

	// First fold.
	DoFoldarea(ch, area.Filename)
	_ = readOutput(ch, client)
	if _, err := os.Stat(bakPath); !os.IsNotExist(err) {
		t.Errorf("first fold: expected NO .bak, stat err=%v", err)
	}

	v1, err := os.ReadFile(livePath)
	if err != nil {
		t.Fatalf("read v1: %v", err)
	}

	// Second fold with mutated content.
	area.Name = "Test Area v2"
	DoFoldarea(ch, area.Filename)
	_ = readOutput(ch, client)

	bakContent, err := os.ReadFile(bakPath)
	if err != nil {
		t.Fatalf("expected .bak after second fold: %v", err)
	}
	if string(bakContent) != string(v1) {
		t.Errorf(".bak should equal v1 after second fold")
	}
}

// --- Wave 3 (G5): DoUnfoldarea tests ---
//
// Plan §D3 + §G5 + §A9-A10. The Go port scopes the C semantic DOWN to
// a "use hotboot" guidance message because internal/persist/area.go's
// loadAreaFile is not re-entrant (would corrupt w.Areas + index maps).

func TestDoUnfoldarea_NoArgument(t *testing.T) {
	ch, client, _, _, cleanup := setupSaveAreaWorld(t)
	defer cleanup()

	DoUnfoldarea(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Unfold what?") {
		t.Errorf("expected 'Unfold what?', got: %q", out)
	}
}

func TestDoUnfoldarea_TrustGate(t *testing.T) {
	ch, client, _, _, cleanup := setupSaveAreaWorld(t)
	defer cleanup()

	ch.Level = 1
	ch.Trust = 0
	DoUnfoldarea(ch, "myzone.are")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Huh?") {
		t.Errorf("expected 'Huh?' for non-immortal, got: %q", out)
	}
}

func TestDoUnfoldarea_PrintsGuidance(t *testing.T) {
	ch, client, _, _, cleanup := setupSaveAreaWorld(t)
	defer cleanup()

	DoUnfoldarea(ch, "myzone.are")
	out := readOutput(ch, client)

	for _, want := range []string{
		"not supported",
		"hotboot",
		"restart",
	} {
		if !strings.Contains(strings.ToLower(out), want) {
			t.Errorf("expected guidance to contain %q, got: %q", want, out)
		}
	}
}

func TestDoUnfoldarea_DoesNotCallLoader(t *testing.T) {
	ch, client, _, _, cleanup := setupSaveAreaWorld(t)
	defer cleanup()

	areasBefore := len(WorldRef.Areas)
	DoUnfoldarea(ch, "test.are")
	_ = readOutput(ch, client)
	areasAfter := len(WorldRef.Areas)

	if areasBefore != areasAfter {
		t.Errorf("WorldRef.Areas count changed (%d -> %d) — loader was invoked",
			areasBefore, areasAfter)
	}
}

// Compile-time dependency check.
var _ = func() *types.AreaData { return nil }
