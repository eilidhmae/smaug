package act

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

func setupAreaWorld() (*types.CharData, net.Conn, *types.AreaData, func()) {
	w := setupOlcWorld()
	area := &types.AreaData{
		Name:           "Test Area",
		Author:         "Tester",
		Filename:       "test.are",
		ResetMsg:       "The area resets.",
		LowRVnum:       100,
		HiRVnum:        199,
		LowMVnum:       200,
		HiMVnum:        299,
		LowOVnum:       300,
		HiOVnum:        399,
		ResetFrequency: 15,
		Age:            5,
		NPlayer:        2,
		Flags:          0,
	}
	w.Areas = append(w.Areas, area)

	ch, client := makeImmTestChar("Builder")
	room := &types.RoomIndexData{Vnum: 100, Name: "Area Room", Area: area}
	w.Rooms[100] = room
	handler.CharToRoom(ch, room)

	return ch, client, area, func() { client.Close() }
}

// --- DoAset tests ---

func TestDoAset_NoArg(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoAset(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Syntax") {
		t.Errorf("expected syntax help, got: %q", out)
	}
}

func TestDoAset_AreaNotFound(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoAset(ch, "nonexistent name Foo")
	out := readOutput(ch, client)
	if !strings.Contains(out, "No such area") {
		t.Errorf("expected 'No such area', got: %q", out)
	}
}

func TestDoAset_NoField(t *testing.T) {
	ch, client, _, cleanup := setupAreaWorld()
	defer cleanup()

	DoAset(ch, "test")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Syntax") || !strings.Contains(out, "Valid fields") {
		t.Errorf("expected syntax or valid fields, got: %q", out)
	}
}

func TestDoAset_Fields(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		check   func(area *types.AreaData) bool
		wantOut string
	}{
		{"name", "test name New Name", func(a *types.AreaData) bool { return a.Name == "New Name" }, "name set to"},
		{"author", "test author Eilidh", func(a *types.AreaData) bool { return a.Author == "Eilidh" }, "author set to"},
		{"resetmsg", "test resetmsg The wind blows.", func(a *types.AreaData) bool { return a.ResetMsg == "The wind blows." }, "resetmsg set to"},
		{"low_r_vnum", "test low_r_vnum 50", func(a *types.AreaData) bool { return a.LowRVnum == 50 }, "low_r_vnum set to"},
		{"hi_r_vnum", "test hi_r_vnum 150", func(a *types.AreaData) bool { return a.HiRVnum == 150 }, "hi_r_vnum set to"},
		{"low_m_vnum", "test low_m_vnum 60", func(a *types.AreaData) bool { return a.LowMVnum == 60 }, "low_m_vnum set to"},
		{"hi_m_vnum", "test hi_m_vnum 160", func(a *types.AreaData) bool { return a.HiMVnum == 160 }, "hi_m_vnum set to"},
		{"low_o_vnum", "test low_o_vnum 70", func(a *types.AreaData) bool { return a.LowOVnum == 70 }, "low_o_vnum set to"},
		{"hi_o_vnum", "test hi_o_vnum 170", func(a *types.AreaData) bool { return a.HiOVnum == 170 }, "hi_o_vnum set to"},
		{"resetfreq", "test resetfreq 20", func(a *types.AreaData) bool { return a.ResetFrequency == 20 }, "resetfreq set to"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ch, client, area, cleanup := setupAreaWorld()
			defer cleanup()

			DoAset(ch, tc.args)
			out := readOutput(ch, client)

			if !strings.Contains(strings.ToLower(out), tc.wantOut) {
				t.Errorf("expected output containing %q, got: %q", tc.wantOut, out)
			}
			if !tc.check(area) {
				t.Errorf("field check failed for %s", tc.name)
			}
		})
	}
}

func TestDoAset_BadField(t *testing.T) {
	ch, client, _, cleanup := setupAreaWorld()
	defer cleanup()

	DoAset(ch, "test badfield 5")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Valid fields") {
		t.Errorf("expected valid fields, got: %q", out)
	}
}

// --- DoAstat tests ---

func TestDoAstat_NoArg_CurrentArea(t *testing.T) {
	ch, client, area, cleanup := setupAreaWorld()
	defer cleanup()

	DoAstat(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, area.Name) {
		t.Errorf("expected area name %q in output, got: %q", area.Name, out)
	}
	if !strings.Contains(out, area.Author) {
		t.Errorf("expected author %q in output, got: %q", area.Author, out)
	}
	if !strings.Contains(out, area.Filename) {
		t.Errorf("expected filename %q in output, got: %q", area.Filename, out)
	}
}

func TestDoAstat_NamedArea(t *testing.T) {
	ch, client, area, cleanup := setupAreaWorld()
	defer cleanup()

	DoAstat(ch, "test")
	out := readOutput(ch, client)

	if !strings.Contains(out, area.Name) {
		t.Errorf("expected area name %q in output, got: %q", area.Name, out)
	}
}

func TestDoAstat_AreaNotFound(t *testing.T) {
	ch, client, _, cleanup := setupAreaWorld()
	defer cleanup()

	DoAstat(ch, "nonexistent")
	out := readOutput(ch, client)
	if !strings.Contains(out, "No such area") {
		t.Errorf("expected 'No such area', got: %q", out)
	}
}

func TestDoAstat_NoRoom(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = nil

	DoAstat(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "not in") {
		t.Errorf("expected 'not in' message, got: %q", out)
	}
}

func TestDoAstat_OutputContainsVnumRanges(t *testing.T) {
	ch, client, _, cleanup := setupAreaWorld()
	defer cleanup()

	DoAstat(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "100") || !strings.Contains(out, "199") {
		t.Errorf("expected room vnum range in output, got: %q", out)
	}
	if !strings.Contains(out, "Reset Freq") {
		t.Errorf("expected 'Reset Freq' in output, got: %q", out)
	}
}

// --- DoSaveArea tests ---

// setupSaveAreaWorld creates an area world rooted in a temp directory so
// DoSaveArea can write real files without touching the repo.
func setupSaveAreaWorld(t *testing.T) (*types.CharData, net.Conn, *types.AreaData, string, func()) {
	t.Helper()
	ch, client, area, cleanup := setupAreaWorld()

	tmpDir := t.TempDir()
	areaDir := filepath.Join(tmpDir, "area")
	if err := os.MkdirAll(areaDir, 0o755); err != nil {
		cleanup()
		t.Fatalf("failed to create area dir: %v", err)
	}
	WorldRef.DataDir = tmpDir

	return ch, client, area, tmpDir, cleanup
}

func TestDoSaveArea_WritesFileAndCleansTmp(t *testing.T) {
	ch, client, area, tmpDir, cleanup := setupSaveAreaWorld(t)
	defer cleanup()

	DoSaveArea(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(strings.ToLower(out), "saved") {
		t.Errorf("expected success message, got: %q", out)
	}

	path := filepath.Join(tmpDir, "area", area.Filename)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected area file at %s: %v", path, err)
	}

	tmpPath := path + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Errorf("expected no .tmp file after successful save, stat err=%v", err)
	}
}

// --- Path-containment guard tests (plan-phase6-olc-oedit.md §G11) ---

// TestDoSaveArea_RejectsPathTraversal pins the §G11 guard against directory-
// traversal attempts via area.Filename (e.g. "../../etc/passwd"). Without
// the guard, filepath.Join would escape the area directory and write into
// the wider filesystem. Mutation gate: drop the HasPrefix check and this
// test fails because the filename is silently accepted.
func TestDoSaveArea_RejectsPathTraversal(t *testing.T) {
	ch, client, area, tmpDir, cleanup := setupSaveAreaWorld(t)
	defer cleanup()

	area.Filename = "../../etc/passwd"

	DoSaveArea(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Invalid area filename") {
		t.Errorf("expected rejection message, got: %q", out)
	}
	// The traversal target must NOT exist on disk relative to tmpDir.
	if _, err := os.Stat(filepath.Join(tmpDir, "..", "..", "etc", "passwd.tmp")); err == nil {
		t.Errorf("path traversal succeeded — a file was written outside the area dir")
	}
}

// TestDoSaveArea_RejectsAbsolutePath pins the absolute-path branch of the
// guard. filepath.Clean preserves leading slashes, so a filename like
// "/tmp/evil.are" must be rejected by the HasPrefix check (the joined
// path resolves outside the area dir entirely).
func TestDoSaveArea_RejectsAbsolutePath(t *testing.T) {
	ch, client, area, _, cleanup := setupSaveAreaWorld(t)
	defer cleanup()

	// Use an absolute path that is definitely outside the temp area dir.
	evilPath := filepath.Join(t.TempDir(), "evil.are")
	area.Filename = evilPath

	DoSaveArea(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Invalid area filename") {
		t.Errorf("expected rejection message for absolute path, got: %q", out)
	}
	// The evil target must NOT have been created.
	if _, err := os.Stat(evilPath); err == nil {
		t.Errorf("absolute path traversal succeeded — file was created at %s", evilPath)
	}
	if _, err := os.Stat(evilPath + ".tmp"); err == nil {
		t.Errorf("absolute path .tmp file was created at %s.tmp", evilPath)
	}
}

// TestDoSaveArea_AllowsNormalFilename confirms the guard is permissive for
// well-formed filenames — "foo.are" must pass through and write into the
// area dir as before. Regression-guards against a too-strict guard.
func TestDoSaveArea_AllowsNormalFilename(t *testing.T) {
	ch, client, area, tmpDir, cleanup := setupSaveAreaWorld(t)
	defer cleanup()

	area.Filename = "foo.are"

	DoSaveArea(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(strings.ToLower(out), "saved") {
		t.Errorf("expected success for normal filename, got: %q", out)
	}
	expected := filepath.Join(tmpDir, "area", "foo.are")
	if _, err := os.Stat(expected); err != nil {
		t.Errorf("expected file at %s after save: %v", expected, err)
	}
}

func TestDoSaveArea_LeavesLiveFileOnError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("cannot simulate permission failure as root")
	}
	ch, client, area, tmpDir, cleanup := setupSaveAreaWorld(t)
	defer cleanup()

	// Seed a known-good live file.
	path := filepath.Join(tmpDir, "area", area.Filename)
	originalContents := []byte("ORIGINAL CONTENT - DO NOT CORRUPT\n")
	if err := os.WriteFile(path, originalContents, 0o644); err != nil {
		t.Fatalf("failed to seed live file: %v", err)
	}

	// Make the area dir read-only so os.Create of the .tmp fails.
	areaDir := filepath.Join(tmpDir, "area")
	if err := os.Chmod(areaDir, 0o555); err != nil {
		t.Fatalf("failed to chmod area dir: %v", err)
	}
	defer os.Chmod(areaDir, 0o755) //nolint:errcheck // test cleanup

	DoSaveArea(ch, "")
	_ = readOutput(ch, client)

	// The live file must be untouched.
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("live file gone after failed save: %v", err)
	}
	if string(got) != string(originalContents) {
		t.Errorf("live file was corrupted on failed save: got %q, want %q", got, originalContents)
	}

	// No lingering .tmp file.
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("expected no .tmp file after failed save, stat err=%v", err)
	}
}
