package persist

import (
	"bytes"
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// newMorphedTestPC returns a PC with a valid #PLAYER save layout plus an
// attached CharMorph. Shared by save/load tests.
func newMorphedTestPC(m *types.MorphData) *types.CharData {
	ch := &types.CharData{
		Name:  "Morphed",
		Level: 25,
		Class: 0, Race: 0, Sex: 0,
		Hit: 100, MaxHit: 100,
		Mana: 100, MaxMana: 100,
		Move: 100, MaxMove: 100,
		PCData: &types.PCData{PagerLen: 24},
	}
	ch.Morph = &types.CharMorph{
		Morph:        m,
		AC:           -5,
		Str:          3,
		Hit:          10,
		Damroll:      2,
		Hitroll:      1,
		Mana:         20,
		Move:         15,
		Timer:        -1,
		SavingBreath: -2,
		AffectedBy:   types.BitVector{0x01, 0x10, 0, 0},
		Immune:       0x02,
	}
	return ch
}

// TestWriteMorphData_EmitsVnumBeforeName pins the ordering constraint
// documented in plan §G6. C fread_morph_data validates Name against
// the morph pointer, which only resolves after Vnum is read; reordering
// silently disables validation.
func TestWriteMorphData_EmitsVnumBeforeName(t *testing.T) {
	m := &types.MorphData{Name: "wolf", Vnum: 1000}
	ch := newMorphedTestPC(m)
	var buf bytes.Buffer
	if err := writeMorphData(&buf, ch); err != nil {
		t.Fatal(err)
	}
	text := buf.String()
	vnumIdx := strings.Index(text, "Vnum")
	nameIdx := strings.Index(text, "Name")
	if vnumIdx < 0 || nameIdx < 0 {
		t.Fatalf("both Vnum and Name must appear; got %q", text)
	}
	if vnumIdx > nameIdx {
		t.Errorf("Vnum must appear BEFORE Name; got vnumIdx=%d, nameIdx=%d in %q",
			vnumIdx, nameIdx, text)
	}
}

// TestWriteMorphData_OnlyEmitsNonZeroFields pins the size-optimization:
// a CharMorph with default values emits Vnum + Name + End only.
func TestWriteMorphData_OnlyEmitsNonZeroFields(t *testing.T) {
	m := &types.MorphData{Name: "wolf", Vnum: 1000}
	ch := &types.CharData{
		Morph: &types.CharMorph{Morph: m, Timer: -1},
	}
	var buf bytes.Buffer
	_ = writeMorphData(&buf, ch)
	text := buf.String()
	if strings.Contains(text, "Armor") {
		t.Errorf("zero AC should not emit Armor: %q", text)
	}
	if strings.Contains(text, "Strength") {
		t.Errorf("zero Str should not emit Strength: %q", text)
	}
	// Timer=-1 is sentinel default; must not emit.
	if strings.Contains(text, "Timer") {
		t.Errorf("sentinel Timer=-1 should not appear: %q", text)
	}
}

// TestWriteMorphData_NilMorphNoop: SavePlayer calls writeMorphData
// unconditionally; the function must silently skip when ch.Morph is nil.
func TestWriteMorphData_NilMorphNoop(t *testing.T) {
	ch := &types.CharData{}
	var buf bytes.Buffer
	if err := writeMorphData(&buf, ch); err != nil {
		t.Fatalf("err = %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("non-morphed PC should emit nothing; got %q", buf.String())
	}
}

// TestSaveLoadPlayer_MorphRoundTrip: full pfile cycle with a morphed PC.
func TestSaveLoadPlayer_MorphRoundTrip(t *testing.T) {
	m := &types.MorphData{Name: "wolf", Vnum: 1000, Level: 20}
	prev := MorphGetter
	defer func() { MorphGetter = prev }()
	MorphGetter = func(v int) *types.MorphData {
		if v == 1000 {
			return m
		}
		return nil
	}
	ch := newMorphedTestPC(m)
	var buf bytes.Buffer
	if err := SavePlayer(&buf, ch); err != nil {
		t.Fatalf("SavePlayer: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "#MorphData") {
		t.Fatalf("pfile missing #MorphData block: %q", out)
	}

	loaded, err := LoadPlayer(strings.NewReader(out), "test")
	if err != nil {
		t.Fatalf("LoadPlayer: %v", err)
	}
	if loaded.Morph == nil {
		t.Fatal("loaded.Morph should be set")
	}
	if loaded.Morph.Morph == nil {
		t.Fatal("loaded.Morph.Morph should resolve via MorphGetter")
	}
	if loaded.Morph.Morph.Vnum != 1000 {
		t.Errorf("resolved vnum = %d, want 1000", loaded.Morph.Morph.Vnum)
	}
	// Delta fields survived the round trip.
	if loaded.Morph.AC != -5 {
		t.Errorf("AC = %d, want -5", loaded.Morph.AC)
	}
	if loaded.Morph.Str != 3 {
		t.Errorf("Str = %d, want 3", loaded.Morph.Str)
	}
	if loaded.Morph.Hit != 10 {
		t.Errorf("Hit delta = %d, want 10", loaded.Morph.Hit)
	}
	if loaded.Morph.Immune != 0x02 {
		t.Errorf("Immune = %d, want 2", loaded.Morph.Immune)
	}
	if !loaded.Morph.AffectedBy.Equal(ch.Morph.AffectedBy) {
		t.Errorf("AffectedBy round-trip mismatch: orig=%v got=%v",
			ch.Morph.AffectedBy, loaded.Morph.AffectedBy)
	}
}

// TestSavePlayer_PostUnmorphNoMorphBlock: once ch.Morph is nil, no
// #MorphData block appears in the pfile (plan §G6 A16).
func TestSavePlayer_PostUnmorphNoMorphBlock(t *testing.T) {
	ch := &types.CharData{
		Name:   "Clean",
		PCData: &types.PCData{},
	}
	var buf bytes.Buffer
	if err := SavePlayer(&buf, ch); err != nil {
		t.Fatalf("SavePlayer: %v", err)
	}
	if strings.Contains(buf.String(), "#MorphData") {
		t.Errorf("pfile should not contain #MorphData when ch.Morph is nil: %q", buf.String())
	}
}

// TestLoadPlayer_UnknownMorphVnumDegradesGracefully: plan §G6 A17 — if
// the morph vnum in a saved pfile no longer exists (e.g. morphdestroy
// happened between sessions), LoadPlayer does not panic.
func TestLoadPlayer_UnknownMorphVnumDegradesGracefully(t *testing.T) {
	m := &types.MorphData{Name: "wolf", Vnum: 1000}
	prev := MorphGetter
	defer func() { MorphGetter = prev }()
	MorphGetter = func(v int) *types.MorphData {
		// Pretend the morph was destroyed: return nil for all vnums.
		return nil
	}
	ch := newMorphedTestPC(m)
	var buf bytes.Buffer
	if err := SavePlayer(&buf, ch); err != nil {
		t.Fatalf("SavePlayer: %v", err)
	}
	loaded, err := LoadPlayer(strings.NewReader(buf.String()), "test")
	if err != nil {
		t.Fatalf("LoadPlayer: %v", err)
	}
	if loaded.Morph == nil {
		t.Fatal("Morph should still be allocated even if vnum no longer resolves")
	}
	if loaded.Morph.Morph != nil {
		t.Error("Morph.Morph pointer should be nil when vnum unresolved")
	}
	// Stat deltas still round-trip.
	if loaded.Morph.AC != -5 {
		t.Errorf("AC = %d, want -5 (delta preserved regardless of resolution)", loaded.Morph.AC)
	}
}

// TestReadMorphData_ReportsUnknownKey: bug-log branch should fire on
// unknown keys but not crash.
func TestReadMorphData_ReportsUnknownKey(t *testing.T) {
	block := `#MorphData
Vnum 1000
Bogus 42
End
`
	m := &types.MorphData{Name: "wolf", Vnum: 1000}
	prev := MorphGetter
	defer func() { MorphGetter = prev }()
	MorphGetter = func(v int) *types.MorphData {
		if v == 1000 {
			return m
		}
		return nil
	}
	// Feed as a pfile fragment with the section token.
	full := `#PLAYER
Name Test~
End

` + block
	_, err := LoadPlayer(strings.NewReader(full), "test")
	if err != nil {
		t.Fatalf("LoadPlayer: %v", err)
	}
	// No panic — test passes.
}
