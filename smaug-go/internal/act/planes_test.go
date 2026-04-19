package act

import (
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// setupPlanesWorld creates a fresh world, installs it at WorldRef, and
// returns it. Call t.Cleanup to restore the previous WorldRef if the
// test cares (most planes tests are self-contained).
func setupPlanesWorld(t *testing.T) *world.World {
	t.Helper()
	w := world.New("/tmp/test-planes")
	prev := WorldRef
	WorldRef = w
	t.Cleanup(func() { WorldRef = prev })
	return w
}

// --- DoPlist ----------------------------------------------------------

func TestDoPlist_EmptySlicePrintsHeaderOnly(t *testing.T) {
	setupPlanesWorld(t)
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPlist(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Planes:") {
		t.Errorf("expected 'Planes:' header; got %q", out)
	}
	if !strings.Contains(out, "-------") {
		t.Errorf("expected '-------' separator; got %q", out)
	}
}

func TestDoPlist_TwoPlanesLists(t *testing.T) {
	w := setupPlanesWorld(t)
	w.Planes = []*types.PlaneData{
		{Name: "Astral"},
		{Name: "Ethereal"},
	}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPlist(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Astral") {
		t.Errorf("expected 'Astral' in output; got %q", out)
	}
	if !strings.Contains(out, "Ethereal") {
		t.Errorf("expected 'Ethereal' in output; got %q", out)
	}
}

func TestDoPlist_NilWorldRefNoPanic(t *testing.T) {
	prev := WorldRef
	WorldRef = nil
	t.Cleanup(func() { WorldRef = prev })
	ch, client := makeTestChar("Tester")
	defer client.Close()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("DoPlist with nil WorldRef panicked: %v", r)
		}
	}()
	DoPlist(ch, "")
	out := readOutput(ch, client)
	// Even with nil WorldRef, the header emits.
	if !strings.Contains(out, "Planes:") {
		t.Errorf("expected header even on nil WorldRef; got %q", out)
	}
}

func TestDoPlist_IgnoresArgument(t *testing.T) {
	w := setupPlanesWorld(t)
	w.Planes = []*types.PlaneData{{Name: "Astral"}}
	ch1, c1 := makeTestChar("A")
	defer c1.Close()
	ch2, c2 := makeTestChar("B")
	defer c2.Close()

	DoPlist(ch1, "")
	out1 := readOutput(ch1, c1)
	DoPlist(ch2, "some nonsense arg")
	out2 := readOutput(ch2, c2)

	if out1 != out2 {
		t.Errorf("output differs with/without argument:\n1=%q\n2=%q", out1, out2)
	}
}

// --- DoPstat ----------------------------------------------------------

func TestDoPstat_EmptyArgPromptsStatWhich(t *testing.T) {
	setupPlanesWorld(t)
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPstat(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Stat which plane?") {
		t.Errorf("expected 'Stat which plane?' prompt; got %q", out)
	}
}

func TestDoPstat_UnknownPlaneSamePrompt(t *testing.T) {
	w := setupPlanesWorld(t)
	w.Planes = []*types.PlaneData{{Name: "Astral"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPstat(ch, "nonexistent")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Stat which plane?") {
		t.Errorf("expected 'Stat which plane?' for unknown plane; got %q", out)
	}
}

func TestDoPstat_ExactMatchPrintsName(t *testing.T) {
	w := setupPlanesWorld(t)
	w.Planes = []*types.PlaneData{{Name: "Astral"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPstat(ch, "astral")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Name: Astral") {
		t.Errorf("expected 'Name: Astral'; got %q", out)
	}
}

func TestDoPstat_PrefixMatchPrintsName(t *testing.T) {
	w := setupPlanesWorld(t)
	w.Planes = []*types.PlaneData{{Name: "Astral"}, {Name: "Ethereal"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPstat(ch, "eth")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Name: Ethereal") {
		t.Errorf("expected 'Name: Ethereal' for prefix 'eth'; got %q", out)
	}
}

func TestDoPstat_CaseInsensitive(t *testing.T) {
	w := setupPlanesWorld(t)
	w.Planes = []*types.PlaneData{{Name: "Astral"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPstat(ch, "ASTRAL")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Name: Astral") {
		t.Errorf("expected 'Name: Astral' for ASTRAL; got %q", out)
	}
}

func TestDoPstat_ExactBeforePrefix(t *testing.T) {
	w := setupPlanesWorld(t)
	// Astra is shorter; Astral has Astra as prefix. An exact match on
	// "astra" must resolve to Astra, not Astral.
	w.Planes = []*types.PlaneData{{Name: "Astra"}, {Name: "Astral"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPstat(ch, "astra")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Name: Astra\n") {
		t.Errorf("expected exact match 'Name: Astra'; got %q", out)
	}
}

// --- DoPset — syntax / save ------------------------------------------

func TestDoPset_EmptyArgPrintsSyntax(t *testing.T) {
	setupPlanesWorld(t)
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPset(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Syntax: pset <plane> create") {
		t.Errorf("expected syntax header; got %q", out)
	}
	if !strings.Contains(out, "pset save") {
		t.Errorf("expected 'pset save' line; got %q", out)
	}
	if !strings.Contains(out, "name") {
		t.Errorf("expected 'name' field mention; got %q", out)
	}
}

func TestDoPset_SaveSubcommand(t *testing.T) {
	w := setupPlanesWorld(t)
	w.Planes = []*types.PlaneData{{Name: "Astral"}}
	tmp := t.TempDir()
	prevPath := PlanesFilePath
	PlanesFilePath = filepath.Join(tmp, "planes.dat")
	t.Cleanup(func() { PlanesFilePath = prevPath })

	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPset(ch, "save")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Planes saved.") {
		t.Errorf("expected 'Planes saved.' ack; got %q", out)
	}

	// File must exist with the Astral block.
	data, err := os.ReadFile(PlanesFilePath)
	if err != nil {
		t.Fatalf("saved file unreadable: %v", err)
	}
	if !strings.Contains(string(data), "Name      Astral~") {
		t.Errorf("saved file missing Astral block; got %s", string(data))
	}
}

func TestDoPset_SaveErrorStillEmitsMessage(t *testing.T) {
	setupPlanesWorld(t)
	prevPath := PlanesFilePath
	PlanesFilePath = "/nonexistent/dir/planes.dat"
	t.Cleanup(func() { PlanesFilePath = prevPath })

	var bugCount atomic.Int64
	prevSink := util.BugSink
	util.BugSink = func(msg string) { bugCount.Add(1) }
	t.Cleanup(func() { util.BugSink = prevSink })

	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPset(ch, "save")
	out := readOutput(ch, client)

	// User sees ack regardless of I/O error (matches C behavior).
	if !strings.Contains(out, "Planes saved.") {
		t.Errorf("expected 'Planes saved.' even on error; got %q", out)
	}
	if bugCount.Load() == 0 {
		t.Error("expected util.Bug call on I/O error")
	}
}

// --- DoPset — create --------------------------------------------------

func TestDoPset_CreateNewPlane(t *testing.T) {
	w := setupPlanesWorld(t)
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPset(ch, "astral create")
	out := readOutput(ch, client)

	if len(w.Planes) != 1 {
		t.Fatalf("expected 1 plane after create; got %d", len(w.Planes))
	}
	if w.Planes[0].Name != "astral" {
		t.Errorf("stored name = %q, want %q (OneArgument lowercases)", w.Planes[0].Name, "astral")
	}
	if !strings.Contains(out, "Plane created.") {
		t.Errorf("expected 'Plane created.'; got %q", out)
	}
}

func TestDoPset_CreateDuplicate(t *testing.T) {
	w := setupPlanesWorld(t)
	w.Planes = []*types.PlaneData{{Name: "astral"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPset(ch, "astral create")
	out := readOutput(ch, client)

	if len(w.Planes) != 1 {
		t.Errorf("duplicate create appended: len = %d", len(w.Planes))
	}
	if !strings.Contains(out, "Plane already exists.") {
		t.Errorf("expected 'Plane already exists.'; got %q", out)
	}
}

func TestDoPset_CreatePrefixMatch(t *testing.T) {
	w := setupPlanesWorld(t)
	ch, client := makeTestChar("Tester")
	defer client.Close()

	// "c" is a prefix of "create" — C str_prefix semantics.
	DoPset(ch, "astral c")
	out := readOutput(ch, client)

	if len(w.Planes) != 1 {
		t.Fatalf("expected create from prefix 'c'; len = %d", len(w.Planes))
	}
	if !strings.Contains(out, "Plane created.") {
		t.Errorf("expected 'Plane created.' on prefix match; got %q", out)
	}
}

func TestDoPset_CreateCaseInsensitiveOp(t *testing.T) {
	w := setupPlanesWorld(t)
	ch, client := makeTestChar("Tester")
	defer client.Close()

	// OneArgument lowercases arg2 to "create", so CREATE works.
	DoPset(ch, "astral CREATE")
	out := readOutput(ch, client)

	if len(w.Planes) != 1 {
		t.Fatalf("expected create from 'CREATE'; len = %d", len(w.Planes))
	}
	if !strings.Contains(out, "Plane created.") {
		t.Errorf("expected 'Plane created.' on uppercase CREATE; got %q", out)
	}
}

// --- DoPset — delete --------------------------------------------------

func TestDoPset_DeleteNoOp(t *testing.T) {
	w := setupPlanesWorld(t)
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPset(ch, "astral delete")
	out := readOutput(ch, client)

	if len(w.Planes) != 0 {
		t.Errorf("slice mutated on delete-of-nonexistent: len = %d", len(w.Planes))
	}
	if !strings.Contains(out, "Plane doesn't exist.") {
		t.Errorf("expected 'Plane doesn't exist.'; got %q", out)
	}
}

func TestDoPset_DeleteRemovesAndReassigns(t *testing.T) {
	w := setupPlanesWorld(t)
	astral := &types.PlaneData{Name: "astral"}
	ethereal := &types.PlaneData{Name: "ethereal"}
	w.Planes = []*types.PlaneData{astral, ethereal}
	room := &types.RoomIndexData{Vnum: 3001, Plane: ethereal}
	w.Rooms[3001] = room

	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPset(ch, "ethereal delete")
	out := readOutput(ch, client)

	if len(w.Planes) != 1 {
		t.Fatalf("expected 1 plane after delete; got %d", len(w.Planes))
	}
	if w.Planes[0].Name != "astral" {
		t.Errorf("surviving plane = %q, want astral", w.Planes[0].Name)
	}
	if w.Rooms[3001].Plane != astral {
		t.Errorf("room 3001 not reassigned to astral after Ethereal delete")
	}
	if !strings.Contains(out, "Plane deleted.") {
		t.Errorf("expected 'Plane deleted.'; got %q", out)
	}
}

func TestDoPset_DeleteLastPlaneRebuildsPrime(t *testing.T) {
	w := setupPlanesWorld(t)
	astral := &types.PlaneData{Name: "astral"}
	w.Planes = []*types.PlaneData{astral}
	room := &types.RoomIndexData{Vnum: 3001, Plane: astral}
	w.Rooms[3001] = room

	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPset(ch, "astral delete")
	out := readOutput(ch, client)

	if len(w.Planes) != 1 {
		t.Fatalf("expected 1 plane (rebuilt Prime Material); got %d", len(w.Planes))
	}
	if w.Planes[0].Name != "Prime Material" {
		t.Errorf("expected 'Prime Material' as sole plane after last-delete; got %q", w.Planes[0].Name)
	}
	if w.Rooms[3001].Plane == nil || w.Rooms[3001].Plane.Name != "Prime Material" {
		t.Errorf("room 3001 not reassigned to Prime Material")
	}
	if !strings.Contains(out, "Plane deleted.") {
		t.Errorf("expected 'Plane deleted.'; got %q", out)
	}
}

func TestDoPset_DeleteFirstPlane(t *testing.T) {
	w := setupPlanesWorld(t)
	astral := &types.PlaneData{Name: "astral"}
	ethereal := &types.PlaneData{Name: "ethereal"}
	w.Planes = []*types.PlaneData{astral, ethereal}

	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPset(ch, "astral delete")
	_ = readOutput(ch, client)

	if len(w.Planes) != 1 {
		t.Fatalf("expected len 1 after first-plane delete; got %d", len(w.Planes))
	}
	if w.Planes[0].Name != "ethereal" {
		t.Errorf("expected surviving plane 'ethereal'; got %q", w.Planes[0].Name)
	}
}

func TestDoPset_DeleteMiddlePlane(t *testing.T) {
	// Post-audit: pins splice-delete correctness for mid-index targets.
	w := setupPlanesWorld(t)
	astral := &types.PlaneData{Name: "astral"}
	ethereal := &types.PlaneData{Name: "ethereal"}
	astral2 := &types.PlaneData{Name: "astral2"}
	w.Planes = []*types.PlaneData{astral, ethereal, astral2}
	room := &types.RoomIndexData{Vnum: 3001, Plane: ethereal}
	w.Rooms[3001] = room

	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPset(ch, "ethereal delete")
	_ = readOutput(ch, client)

	if len(w.Planes) != 2 {
		t.Fatalf("expected len 2; got %d", len(w.Planes))
	}
	if w.Planes[0] != astral || w.Planes[1] != astral2 {
		t.Errorf("slice order wrong after mid delete: got %+v", w.Planes)
	}
	if w.Rooms[3001].Plane != astral {
		t.Errorf("room 3001 should reassign to w.Planes[0] (astral); got %v", w.Rooms[3001].Plane)
	}
}

func TestDoPset_DeleteLastPlaneSymmetric(t *testing.T) {
	w := setupPlanesWorld(t)
	astral := &types.PlaneData{Name: "astral"}
	ethereal := &types.PlaneData{Name: "ethereal"}
	w.Planes = []*types.PlaneData{astral, ethereal}

	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPset(ch, "ethereal delete")
	_ = readOutput(ch, client)

	if len(w.Planes) != 1 || w.Planes[0] != astral {
		t.Errorf("expected surviving [astral]; got %+v", w.Planes)
	}
}

// --- DoPset — name ---------------------------------------------------

func TestDoPset_RenameChangesName(t *testing.T) {
	w := setupPlanesWorld(t)
	w.Planes = []*types.PlaneData{{Name: "astral"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPset(ch, "astral name Astrolabe")
	out := readOutput(ch, client)

	if w.Planes[0].Name != "Astrolabe" {
		t.Errorf("expected renamed to 'Astrolabe'; got %q", w.Planes[0].Name)
	}
	if !strings.Contains(out, "Name changed.") {
		t.Errorf("expected 'Name changed.'; got %q", out)
	}
}

func TestDoPset_RenameToDuplicate(t *testing.T) {
	w := setupPlanesWorld(t)
	w.Planes = []*types.PlaneData{{Name: "astral"}, {Name: "ethereal"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPset(ch, "astral name Ethereal")
	out := readOutput(ch, client)

	if w.Planes[0].Name != "astral" {
		t.Errorf("expected name unchanged; got %q", w.Planes[0].Name)
	}
	if !strings.Contains(out, "Another plane has that name.") {
		t.Errorf("expected 'Another plane has that name.'; got %q", out)
	}
}

func TestDoPset_RenameSmashTilde(t *testing.T) {
	w := setupPlanesWorld(t)
	w.Planes = []*types.PlaneData{{Name: "astral"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPset(ch, "astral name Funky~Name")
	_ = readOutput(ch, client)

	if w.Planes[0].Name != "Funky-Name" {
		t.Errorf("expected SmashTilde 'Funky-Name'; got %q", w.Planes[0].Name)
	}
}

func TestDoPset_RenameMultiWord(t *testing.T) {
	w := setupPlanesWorld(t)
	w.Planes = []*types.PlaneData{{Name: "astral"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPset(ch, "astral name Prime Material")
	out := readOutput(ch, client)

	if w.Planes[0].Name != "Prime Material" {
		t.Errorf("expected 'Prime Material'; got %q", w.Planes[0].Name)
	}
	if !strings.Contains(out, "Name changed.") {
		t.Errorf("expected 'Name changed.'; got %q", out)
	}
}

// --- DoPset — fallthrough --------------------------------------------

func TestDoPset_UnknownOpRePrintsSyntax(t *testing.T) {
	w := setupPlanesWorld(t)
	w.Planes = []*types.PlaneData{{Name: "astral"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPset(ch, "astral wombat")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Syntax: pset <plane> create") {
		t.Errorf("expected syntax fallthrough; got %q", out)
	}
}

// --- Integration: DoPset + persist.LoadPlanes round-trip --------------

func TestDoPset_CreateSaveLoadRoundTrip(t *testing.T) {
	w := setupPlanesWorld(t)
	tmp := t.TempDir()
	prevPath := PlanesFilePath
	PlanesFilePath = filepath.Join(tmp, "planes.dat")
	t.Cleanup(func() { PlanesFilePath = prevPath })

	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoPset(ch, "astral create")
	DoPset(ch, "astral name Astral Plane")
	DoPset(ch, "save")
	_ = readOutput(ch, client)

	reloaded, err := persist.LoadPlanes(PlanesFilePath)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if len(reloaded) != 1 || reloaded[0].Name != "Astral Plane" {
		t.Errorf("reload mismatch: %+v", reloaded)
	}
	// Original world slice still intact.
	if len(w.Planes) != 1 || w.Planes[0].Name != "Astral Plane" {
		t.Errorf("in-memory slice mismatch: %+v", w.Planes)
	}
}
