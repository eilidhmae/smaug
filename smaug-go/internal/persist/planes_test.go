package persist

import (
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// captureBug installs a test BugSink that counts bug-log invocations.
// Returns a cleanup-func the caller t.Cleanup()s.
func captureBug() (*atomic.Int64, func()) {
	var count atomic.Int64
	prev := util.BugSink
	util.BugSink = func(msg string) { count.Add(1) }
	return &count, func() { util.BugSink = prev }
}

// --- LoadPlanes tests -------------------------------------------------

func TestLoadPlanes_MissingFileReturnsNilNil(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.dat")
	list, err := LoadPlanes(path)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if list != nil {
		t.Fatalf("expected nil slice, got %v", list)
	}
}

func TestLoadPlanes_EmptyFileReturnsEmpty(t *testing.T) {
	list, err := LoadPlanes("testdata/planes_empty.dat")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if list == nil {
		t.Fatal("expected non-nil empty slice (stub file case), got nil")
	}
	if len(list) != 0 {
		t.Fatalf("expected len 0, got %d", len(list))
	}
}

func TestLoadPlanes_TildeTerminatedRoundTrip(t *testing.T) {
	list, err := LoadPlanes("testdata/planes_two.dat")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 planes, got %d", len(list))
	}
	if list[0].Name != "Astral" {
		t.Errorf("list[0].Name = %q, want %q", list[0].Name, "Astral")
	}
	if list[1].Name != "Ethereal" {
		t.Errorf("list[1].Name = %q, want %q", list[1].Name, "Ethereal")
	}
}

func TestLoadPlanes_CFormatTolerance(t *testing.T) {
	list, err := LoadPlanes("testdata/planes_cformat.dat")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 planes (c-format), got %d", len(list))
	}
	if list[0].Name != "Astral" {
		t.Errorf("list[0].Name = %q, want %q", list[0].Name, "Astral")
	}
	if list[1].Name != "Ethereal" {
		t.Errorf("list[1].Name = %q, want %q", list[1].Name, "Ethereal")
	}
}

func TestLoadPlanes_BlockWithNoNameDropped(t *testing.T) {
	path := filepath.Join(t.TempDir(), "noname.dat")
	content := "#PLANE\nEnd\n\n#PLANE\nName      Astral~\nEnd\n\n#END\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	bugCount, restore := captureBug()
	t.Cleanup(restore)

	list, err := LoadPlanes(path)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 plane (noname dropped), got %d", len(list))
	}
	if list[0].Name != "Astral" {
		t.Errorf("surviving plane name = %q, want Astral", list[0].Name)
	}
	if bugCount.Load() == 0 {
		t.Error("expected at least one bug-log for noname block")
	}
}

func TestLoadPlanes_UnknownSectionStopsParsing(t *testing.T) {
	list, err := LoadPlanes("testdata/planes_malformed.dat")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	// malformed fixture: one good PLANE, one Garbage key (bug-log but
	// plane still loads), then #BOGUS which stops parsing.
	if len(list) != 1 {
		t.Fatalf("expected 1 plane (Astral), got %d: %+v", len(list), list)
	}
	if list[0].Name != "Astral" {
		t.Errorf("list[0].Name = %q, want Astral", list[0].Name)
	}
}

func TestLoadPlanes_UnknownKeyInBlockSkipsLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "unknownkey.dat")
	content := "#PLANE\nName      Astral~\nGarbage   payload\nEnd\n\n#END\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	bugCount, restore := captureBug()
	t.Cleanup(restore)

	list, err := LoadPlanes(path)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(list) != 1 || list[0].Name != "Astral" {
		t.Fatalf("expected [Astral], got %+v", list)
	}
	if bugCount.Load() == 0 {
		t.Error("expected bug-log for Garbage key")
	}
}

func TestLoadPlanes_NonHashLeaderBugsAndStops(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nohash.dat")
	content := "XYZ oops\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	bugCount, restore := captureBug()
	t.Cleanup(restore)

	list, err := LoadPlanes(path)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected empty list, got %+v", list)
	}
	if bugCount.Load() == 0 {
		t.Error("expected bug-log for non-hash leader")
	}
}

func TestLoadPlanes_ShippedStubIsNoOp(t *testing.T) {
	shippedPath := "../../../db/system/planes.dat"
	if _, err := os.Stat(shippedPath); err != nil {
		t.Skipf("shipped stub not reachable from %s: %v", shippedPath, err)
	}
	bugCount, restore := captureBug()
	t.Cleanup(restore)

	list, err := LoadPlanes(shippedPath)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if list == nil {
		t.Fatal("expected non-nil empty slice on #END-only stub, got nil")
	}
	if len(list) != 0 {
		t.Fatalf("expected len 0 from shipped stub, got %d: %+v", len(list), list)
	}
	if bugCount.Load() != 0 {
		t.Errorf("expected no bug-log on shipped stub, got %d", bugCount.Load())
	}
}

// --- SavePlanes tests -------------------------------------------------

func TestSavePlanes_EmptyListWritesTerminatorOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.dat")
	if err := SavePlanes(path, nil); err != nil {
		t.Fatalf("save: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "#END\n" {
		t.Errorf("content = %q, want %q", string(data), "#END\n")
	}
	// round-trip
	reloaded, err := LoadPlanes(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if len(reloaded) != 0 {
		t.Errorf("reload len = %d, want 0", len(reloaded))
	}
}

func TestSavePlanes_TwoPlanesWritesTwoBlocks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "two.dat")
	input := []*types.PlaneData{
		{Name: "Astral"},
		{Name: "Ethereal"},
	}
	if err := SavePlanes(path, input); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "#PLANE\nName      Astral~\nEnd\n") {
		t.Errorf("missing Astral block; got:\n%s", content)
	}
	if !strings.Contains(content, "#PLANE\nName      Ethereal~\nEnd\n") {
		t.Errorf("missing Ethereal block; got:\n%s", content)
	}
	if !strings.HasSuffix(content, "#END\n") {
		t.Errorf("missing #END terminator; got:\n%s", content)
	}
	// round-trip
	reloaded, err := LoadPlanes(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(reloaded) != 2 || reloaded[0].Name != "Astral" || reloaded[1].Name != "Ethereal" {
		t.Errorf("reload mismatch: %+v", reloaded)
	}
}

func TestSavePlanes_NameWithSpacesPreserved(t *testing.T) {
	path := filepath.Join(t.TempDir(), "prime.dat")
	input := []*types.PlaneData{{Name: "Prime Material"}}
	if err := SavePlanes(path, input); err != nil {
		t.Fatal(err)
	}
	reloaded, err := LoadPlanes(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(reloaded) != 1 || reloaded[0].Name != "Prime Material" {
		t.Errorf("got %+v, want [Prime Material]", reloaded)
	}
}

func TestSavePlanes_FailureReturnsError(t *testing.T) {
	// A path inside a nonexistent directory — os.Create should fail.
	path := filepath.Join(t.TempDir(), "nonexistent-subdir", "planes.dat")
	if err := SavePlanes(path, []*types.PlaneData{{Name: "Astral"}}); err == nil {
		t.Error("expected error, got nil")
	}
}

// --- CheckPlanes tests ------------------------------------------------

func makeTestWorld(rooms ...int) *world.World {
	w := world.New("/tmp/test")
	for _, v := range rooms {
		w.Rooms[v] = &types.RoomIndexData{Vnum: v}
	}
	return w
}

func TestCheckPlanes_EmptySliceAppendsPrimeMaterial(t *testing.T) {
	w := makeTestWorld(3001)
	CheckPlanes(w, nil)
	if len(w.Planes) != 1 {
		t.Fatalf("len(Planes) = %d, want 1", len(w.Planes))
	}
	if w.Planes[0].Name != "Prime Material" {
		t.Errorf("Planes[0].Name = %q, want %q", w.Planes[0].Name, "Prime Material")
	}
	if w.Rooms[3001].Plane != w.Planes[0] {
		t.Errorf("room 3001 not assigned to Prime Material plane")
	}
}

func TestCheckPlanes_NilOrphansAssigned(t *testing.T) {
	w := makeTestWorld(3001, 3002)
	astral := &types.PlaneData{Name: "Astral"}
	w.Planes = []*types.PlaneData{astral}
	w.Rooms[3001].Plane = nil
	w.Rooms[3002].Plane = astral

	CheckPlanes(w, nil)

	if w.Rooms[3001].Plane != astral {
		t.Errorf("room 3001 not reassigned to Astral")
	}
	if w.Rooms[3002].Plane != astral {
		t.Errorf("room 3002 lost Astral assignment")
	}
}

func TestCheckPlanes_DeletedMatchReassigned(t *testing.T) {
	w := makeTestWorld(3001, 3002)
	astral := &types.PlaneData{Name: "Astral"}
	ethereal := &types.PlaneData{Name: "Ethereal"}
	w.Planes = []*types.PlaneData{astral, ethereal}
	w.Rooms[3001].Plane = astral
	w.Rooms[3002].Plane = ethereal

	// Simulate mid-delete: Ethereal is about to be unlinked.
	CheckPlanes(w, ethereal)

	if w.Rooms[3001].Plane != astral {
		t.Errorf("room 3001 should still point at Astral, got %v", w.Rooms[3001].Plane)
	}
	if w.Rooms[3002].Plane != astral {
		t.Errorf("room 3002 should be reassigned to Astral (first_plane), got %v", w.Rooms[3002].Plane)
	}
}

func TestCheckPlanes_NoopWhenEverythingAssigned(t *testing.T) {
	w := makeTestWorld(3001, 3002)
	astral := &types.PlaneData{Name: "Astral"}
	w.Planes = []*types.PlaneData{astral}
	w.Rooms[3001].Plane = astral
	w.Rooms[3002].Plane = astral

	before := len(w.Planes)
	CheckPlanes(w, nil)
	after := len(w.Planes)

	if before != after {
		t.Errorf("CheckPlanes mutated slice length: %d → %d", before, after)
	}
	if w.Rooms[3001].Plane != astral || w.Rooms[3002].Plane != astral {
		t.Errorf("Noop mutated room assignments")
	}
}

func TestCheckPlanes_StableAcrossMultipleCalls(t *testing.T) {
	w := makeTestWorld(3001)
	CheckPlanes(w, nil)
	snapshotLen := len(w.Planes)
	snapshotName := w.Planes[0].Name
	snapshotRoomPlane := w.Rooms[3001].Plane

	CheckPlanes(w, nil)

	if len(w.Planes) != snapshotLen {
		t.Errorf("second call changed slice length: %d → %d", snapshotLen, len(w.Planes))
	}
	if w.Planes[0].Name != snapshotName {
		t.Errorf("second call changed plane[0].Name: %q → %q", snapshotName, w.Planes[0].Name)
	}
	if w.Rooms[3001].Plane != snapshotRoomPlane {
		t.Errorf("second call changed room.Plane pointer")
	}
}

func TestCheckPlanes_NilWorldNoPanic(t *testing.T) {
	// Defensive — CheckPlanes should handle nil world gracefully.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("CheckPlanes(nil, nil) panicked: %v", r)
		}
	}()
	CheckPlanes(nil, nil)
}

// TestSavePlanes_FileMode pins 0600 — security adversary 2026-04-19.
func TestSavePlanes_FileMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "planes.dat")
	if err := SavePlanes(path, []*types.PlaneData{{Name: "Prime Material"}}); err != nil {
		t.Fatalf("SavePlanes: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("planes.dat mode = %#o, want 0o600", mode)
	}
}
