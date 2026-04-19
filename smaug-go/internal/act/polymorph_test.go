package act

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// newMorphTestWorld returns a minimal *world.World for morph command tests.
// Populates two morphs so GetMorph/GetMorphVnum have targets.
func newMorphTestWorld() *world.World {
	w := world.New("")
	w.Morphs = []*types.MorphData{
		{Name: "wolf", Vnum: 1000, Level: 10},
		{Name: "bear", Vnum: 1001, Level: 20},
	}
	w.MorphVnumCounter = 1002
	return w
}

// newMorphPC returns an immortal PC with a capture-buffer descriptor.
func newMorphPC() *types.CharData {
	return &types.CharData{
		Name:  "Imm",
		Level: types.LEVEL_IMMORTAL,
		Trust: types.LEVEL_IMMORTAL,
		Hit:   100, MaxHit: 100,
		Mana: 100, MaxMana: 100,
		Move: 100, MaxMove: 100,
		PCData: &types.PCData{PagerLen: 24},
		Desc:   &types.DescriptorData{},
	}
}

// --- DoMorph ---------------------------------------------------------------

func TestDoMorph_AppliesToSelfByVnum(t *testing.T) {
	prev := WorldRef
	defer func() { WorldRef = prev }()
	WorldRef = newMorphTestWorld()
	ch := newMorphPC()
	DoMorph(ch, "1000")
	if ch.Morph == nil {
		t.Fatal("ch.Morph should be set after morph 1000")
	}
	if ch.Morph.Morph != WorldRef.Morphs[0] {
		t.Errorf("ch.Morph.Morph should point at wolf")
	}
	out := ch.Desc.BufferedOutput()
	if !strings.Contains(out, "Done.") {
		t.Errorf("expected 'Done.' confirmation; got %q", out)
	}
}

func TestDoMorph_NonExistentVnum(t *testing.T) {
	prev := WorldRef
	defer func() { WorldRef = prev }()
	WorldRef = newMorphTestWorld()
	ch := newMorphPC()
	DoMorph(ch, "9999")
	if ch.Morph != nil {
		t.Error("ch.Morph should be nil for unknown vnum")
	}
	out := ch.Desc.BufferedOutput()
	if !strings.Contains(out, "No such morph 9999") {
		t.Errorf("expected 'No such morph 9999' message; got %q", out)
	}
}

func TestDoMorph_NonNumericArgRejected(t *testing.T) {
	prev := WorldRef
	defer func() { WorldRef = prev }()
	WorldRef = newMorphTestWorld()
	ch := newMorphPC()
	DoMorph(ch, "wolf")
	if ch.Morph != nil {
		t.Error("DoMorph should require numeric vnum")
	}
	out := ch.Desc.BufferedOutput()
	if !strings.Contains(out, "Syntax: morph") {
		t.Errorf("expected syntax message; got %q", out)
	}
}

func TestDoMorph_NPCBlocked(t *testing.T) {
	prev := WorldRef
	defer func() { WorldRef = prev }()
	WorldRef = newMorphTestWorld()
	ch := newMorphPC()
	ch.Act.Set(types.ACT_IS_NPC)
	DoMorph(ch, "1000")
	if ch.Morph != nil {
		t.Error("NPC should be rejected")
	}
	out := ch.Desc.BufferedOutput()
	if !strings.Contains(out, "Only player characters") {
		t.Errorf("expected NPC-block message; got %q", out)
	}
}

// --- DoUnmorph -------------------------------------------------------------

func TestDoUnmorph_ClearsSelf(t *testing.T) {
	prev := WorldRef
	defer func() { WorldRef = prev }()
	WorldRef = newMorphTestWorld()
	ch := newMorphPC()
	DoMorph(ch, "1000")
	if ch.Morph == nil {
		t.Fatal("setup: ch should be morphed")
	}
	ch.Desc.ResetBufferedOutput()
	DoUnmorph(ch, "")
	if ch.Morph != nil {
		t.Error("ch.Morph should be nil after unmorph")
	}
	out := ch.Desc.BufferedOutput()
	if !strings.Contains(out, "Done.") {
		t.Errorf("expected 'Done.' confirmation; got %q", out)
	}
}

// --- DoMorphstat -----------------------------------------------------------

func TestDoMorphstat_List(t *testing.T) {
	prev := WorldRef
	defer func() { WorldRef = prev }()
	WorldRef = newMorphTestWorld()
	ch := newMorphPC()
	DoMorphstat(ch, "list")
	out := ch.Desc.BufferedOutput()
	if !strings.Contains(out, "wolf") || !strings.Contains(out, "bear") {
		t.Errorf("list output should contain both morphs; got %q", out)
	}
}

func TestDoMorphstat_ListEmpty(t *testing.T) {
	prev := WorldRef
	defer func() { WorldRef = prev }()
	w := world.New("")
	WorldRef = w
	ch := newMorphPC()
	DoMorphstat(ch, "list")
	out := ch.Desc.BufferedOutput()
	if !strings.Contains(out, "No morph's currently exist") {
		t.Errorf("empty-list message missing; got %q", out)
	}
}

func TestDoMorphstat_ByName(t *testing.T) {
	prev := WorldRef
	defer func() { WorldRef = prev }()
	WorldRef = newMorphTestWorld()
	ch := newMorphPC()
	DoMorphstat(ch, "wolf")
	out := ch.Desc.BufferedOutput()
	if !strings.Contains(out, "Morph Name: wolf") {
		t.Errorf("morphstat output should include 'Morph Name: wolf'; got %q", out)
	}
	if !strings.Contains(out, "Vnum: 1000") {
		t.Errorf("morphstat output should include 'Vnum: 1000'; got %q", out)
	}
}

func TestDoMorphstat_ByVnum(t *testing.T) {
	prev := WorldRef
	defer func() { WorldRef = prev }()
	WorldRef = newMorphTestWorld()
	ch := newMorphPC()
	DoMorphstat(ch, "1001")
	out := ch.Desc.BufferedOutput()
	if !strings.Contains(out, "bear") {
		t.Errorf("morphstat 1001 should include bear; got %q", out)
	}
}

func TestDoMorphstat_Unknown(t *testing.T) {
	prev := WorldRef
	defer func() { WorldRef = prev }()
	WorldRef = newMorphTestWorld()
	ch := newMorphPC()
	DoMorphstat(ch, "nonexistent")
	out := ch.Desc.BufferedOutput()
	if !strings.Contains(out, "No such morph exists") {
		t.Errorf("expected 'No such morph exists'; got %q", out)
	}
}

// --- DoMorphcreate ---------------------------------------------------------

func TestDoMorphcreate_AssignsFreshVnum(t *testing.T) {
	prev := WorldRef
	defer func() { WorldRef = prev }()
	WorldRef = newMorphTestWorld()
	ch := newMorphPC()
	DoMorphcreate(ch, "eagle")
	if len(WorldRef.Morphs) != 3 {
		t.Fatalf("Morphs len = %d, want 3", len(WorldRef.Morphs))
	}
	eagle := WorldRef.GetMorph("eagle")
	if eagle == nil {
		t.Fatal("eagle not in table after create")
	}
	if eagle.Vnum < 1000 {
		t.Errorf("eagle Vnum = %d, want >= 1000 (SetupMorphVnum rule)", eagle.Vnum)
	}
	if eagle.Vnum != 1002 {
		t.Errorf("eagle Vnum = %d, want 1002 (next after 1001)", eagle.Vnum)
	}
}

func TestDoMorphcreate_EmptyArgRejected(t *testing.T) {
	prev := WorldRef
	defer func() { WorldRef = prev }()
	WorldRef = newMorphTestWorld()
	ch := newMorphPC()
	DoMorphcreate(ch, "")
	if len(WorldRef.Morphs) != 2 {
		t.Errorf("Morphs table should be unchanged on empty arg; got len=%d", len(WorldRef.Morphs))
	}
	out := ch.Desc.BufferedOutput()
	if !strings.Contains(out, "Usage: morphcreate") {
		t.Errorf("expected usage message; got %q", out)
	}
}

func TestDoMorphcreate_DuplicateNameRejected(t *testing.T) {
	prev := WorldRef
	defer func() { WorldRef = prev }()
	WorldRef = newMorphTestWorld()
	ch := newMorphPC()
	DoMorphcreate(ch, "wolf")
	if len(WorldRef.Morphs) != 2 {
		t.Errorf("duplicate name should not create; got len=%d", len(WorldRef.Morphs))
	}
	out := ch.Desc.BufferedOutput()
	if !strings.Contains(out, "already exists") {
		t.Errorf("expected 'already exists' message; got %q", out)
	}
}

// --- DoMorphdestroy --------------------------------------------------------

func TestDoMorphdestroy_RemovesFromTable(t *testing.T) {
	prev := WorldRef
	defer func() { WorldRef = prev }()
	WorldRef = newMorphTestWorld()
	MorphFilePath = filepath.Join(t.TempDir(), "morph.dat")
	ch := newMorphPC()
	DoMorphdestroy(ch, "wolf")
	if len(WorldRef.Morphs) != 1 {
		t.Fatalf("Morphs len = %d, want 1", len(WorldRef.Morphs))
	}
	if WorldRef.Morphs[0].Name != "bear" {
		t.Errorf("remaining morph = %q, want bear", WorldRef.Morphs[0].Name)
	}
	out := ch.Desc.BufferedOutput()
	if !strings.Contains(out, "Morph deleted") {
		t.Errorf("expected delete confirmation; got %q", out)
	}
}

func TestDoMorphdestroy_UnmorphsActiveUsers(t *testing.T) {
	prev := WorldRef
	defer func() { WorldRef = prev }()
	WorldRef = newMorphTestWorld()
	MorphFilePath = filepath.Join(t.TempDir(), "morph.dat")

	victim := &types.CharData{
		Name:   "victim",
		Level:  10,
		PCData: &types.PCData{},
	}
	WorldRef.Characters = append(WorldRef.Characters, victim)
	// Manually apply via handler (import avoided by direct DoMorph).
	// Instead, use DoMorph flow: make victim immortal-trust so it
	// bypasses. For this test, just set Morph directly.
	victim.Morph = &types.CharMorph{Morph: WorldRef.Morphs[0]}

	ch := newMorphPC()
	DoMorphdestroy(ch, "wolf")
	if victim.Morph != nil {
		t.Errorf("victim should have been unmorphed; Morph still set")
	}
}

func TestDoMorphdestroy_Unknown(t *testing.T) {
	prev := WorldRef
	defer func() { WorldRef = prev }()
	WorldRef = newMorphTestWorld()
	ch := newMorphPC()
	DoMorphdestroy(ch, "ghost")
	out := ch.Desc.BufferedOutput()
	if !strings.Contains(out, "Unknown morph") {
		t.Errorf("expected 'Unknown morph' message; got %q", out)
	}
}

// --- DoMorphset ------------------------------------------------------------

func TestDoMorphset_SavePersists(t *testing.T) {
	prev := WorldRef
	defer func() { WorldRef = prev }()
	WorldRef = newMorphTestWorld()
	MorphFilePath = filepath.Join(t.TempDir(), "morph.dat")
	ch := newMorphPC()
	DoMorphset(ch, "save")
	out := ch.Desc.BufferedOutput()
	if !strings.Contains(out, "Morph data saved") {
		t.Errorf("expected save confirmation; got %q", out)
	}
	data, err := os.ReadFile(MorphFilePath)
	if err != nil {
		t.Fatalf("read morph.dat: %v", err)
	}
	if !strings.Contains(string(data), "wolf") || !strings.Contains(string(data), "bear") {
		t.Errorf("saved file missing morphs: %q", string(data))
	}
}

func TestDoMorphset_SetLevelPersists(t *testing.T) {
	prev := WorldRef
	defer func() { WorldRef = prev }()
	WorldRef = newMorphTestWorld()
	MorphFilePath = filepath.Join(t.TempDir(), "morph.dat")
	ch := newMorphPC()
	DoMorphset(ch, "wolf level 42")
	if WorldRef.Morphs[0].Level != 42 {
		t.Errorf("wolf Level = %d, want 42", WorldRef.Morphs[0].Level)
	}
	DoMorphset(ch, "save")
	// Reload and verify.
	reloaded, err := persist.LoadMorphs(MorphFilePath)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if len(reloaded) == 0 {
		t.Fatal("reload empty")
	}
	var wolf *types.MorphData
	for _, m := range reloaded {
		if m.Name == "wolf" {
			wolf = m
			break
		}
	}
	if wolf == nil {
		t.Fatal("wolf not in reloaded table")
	}
	if wolf.Level != 42 {
		t.Errorf("reloaded wolf.Level = %d, want 42", wolf.Level)
	}
}

func TestDoMorphset_LevelOutOfRange(t *testing.T) {
	prev := WorldRef
	defer func() { WorldRef = prev }()
	WorldRef = newMorphTestWorld()
	ch := newMorphPC()
	DoMorphset(ch, "wolf level 999")
	if WorldRef.Morphs[0].Level == 999 {
		t.Error("level out-of-range should be rejected")
	}
	out := ch.Desc.BufferedOutput()
	if !strings.Contains(out, "Level range") {
		t.Errorf("expected range message; got %q", out)
	}
}

func TestDoMorphset_SyntaxHelp(t *testing.T) {
	prev := WorldRef
	defer func() { WorldRef = prev }()
	WorldRef = newMorphTestWorld()
	ch := newMorphPC()
	DoMorphset(ch, "?")
	out := ch.Desc.BufferedOutput()
	if !strings.Contains(out, "Field being one of") {
		t.Errorf("expected field list help; got %q", out)
	}
}

func TestDoMorphset_StrRange(t *testing.T) {
	prev := WorldRef
	defer func() { WorldRef = prev }()
	WorldRef = newMorphTestWorld()
	ch := newMorphPC()
	DoMorphset(ch, "wolf str 3")
	if WorldRef.Morphs[0].Str != 3 {
		t.Errorf("wolf.Str = %d, want 3", WorldRef.Morphs[0].Str)
	}
	DoMorphset(ch, "wolf str 99")
	if WorldRef.Morphs[0].Str != 3 {
		t.Errorf("wolf.Str = %d, want 3 (99 out of range must be rejected)", WorldRef.Morphs[0].Str)
	}
}

func TestDoMorphset_DiceString(t *testing.T) {
	prev := WorldRef
	defer func() { WorldRef = prev }()
	WorldRef = newMorphTestWorld()
	ch := newMorphPC()
	DoMorphset(ch, "wolf hp 2d10+5")
	if WorldRef.Morphs[0].Hit != "2d10+5" {
		t.Errorf("wolf.Hit = %q, want 2d10+5", WorldRef.Morphs[0].Hit)
	}
	// "0" should clear the field (C :521-522 convention).
	DoMorphset(ch, "wolf hp 0")
	if WorldRef.Morphs[0].Hit != "" {
		t.Errorf(`wolf.Hit after "hp 0" = %q, want ""`, WorldRef.Morphs[0].Hit)
	}
}
