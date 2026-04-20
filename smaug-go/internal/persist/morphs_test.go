package persist

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// ---------------------------------------------------------------------------
// LoadMorphs — missing / empty / basic file
// ---------------------------------------------------------------------------

func TestLoadMorphs_MissingFile(t *testing.T) {
	morphs, err := LoadMorphs(filepath.Join(t.TempDir(), "no-such-file.dat"))
	if err != nil {
		t.Fatalf("LoadMorphs(missing) err = %v, want nil", err)
	}
	if morphs != nil {
		t.Errorf("LoadMorphs(missing) = %v, want nil", morphs)
	}
}

func TestLoadMorphs_EndOnlyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "morph.dat")
	if err := os.WriteFile(path, []byte("#END\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	morphs, err := LoadMorphs(path)
	if err != nil {
		t.Fatalf("LoadMorphs err = %v", err)
	}
	if len(morphs) != 0 {
		t.Errorf("LoadMorphs(#END-only) = %d morphs, want 0", len(morphs))
	}
}

func TestLoadMorphs_SingleMorph(t *testing.T) {
	path := filepath.Join(t.TempDir(), "morph.dat")
	content := `Morph           	wolf
Vnum              1000
ShortDesc       	a fierce wolf~
Longdesc        	A wolf pads silently through the grass.~
Keywords        	wolf fur fang~
MorphSelf       	You transform into a wolf.~
MorphOther      	$n drops to all fours as a wolf.~
UnmorphSelf     	You return to your normal form.~
UnmorphOther    	$n shimmers back to $s normal shape.~
Level        	25
Armor        	-10
Strength        	3
Dexterity       	2
Hit             	2d20+50~
Damroll         	1d4+2~
Hitroll         	1d2+1~
Mana            	0~
Move            	1d10+20~
End

#END
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	morphs, err := LoadMorphs(path)
	if err != nil {
		t.Fatalf("LoadMorphs err = %v", err)
	}
	if len(morphs) != 1 {
		t.Fatalf("LoadMorphs returned %d morphs, want 1", len(morphs))
	}
	m := morphs[0]
	if m.Name != "wolf" {
		t.Errorf("Name = %q, want wolf", m.Name)
	}
	if m.Vnum != 1000 {
		t.Errorf("Vnum = %d, want 1000", m.Vnum)
	}
	if m.Level != 25 {
		t.Errorf("Level = %d, want 25", m.Level)
	}
	if m.AC != -10 {
		t.Errorf("AC = %d, want -10", m.AC)
	}
	if m.Str != 3 {
		t.Errorf("Str = %d, want 3", m.Str)
	}
	if m.Dex != 2 {
		t.Errorf("Dex = %d, want 2", m.Dex)
	}
	if m.Hit != "2d20+50" {
		t.Errorf("Hit = %q, want 2d20+50", m.Hit)
	}
	if m.Damroll != "1d4+2" {
		t.Errorf("Damroll = %q, want 1d4+2", m.Damroll)
	}
	if m.MorphSelf != "You transform into a wolf." {
		t.Errorf("MorphSelf = %q", m.MorphSelf)
	}
	// Defaults
	if m.Sex != -1 {
		t.Errorf("Sex default = %d, want -1", m.Sex)
	}
	if m.DefPos != types.POS_STANDING {
		t.Errorf("DefPos default = %d, want POS_STANDING", m.DefPos)
	}
	if m.Timer != -1 {
		t.Errorf("Timer default = %d, want -1", m.Timer)
	}
}

func TestLoadMorphs_MultipleMorphs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "morph.dat")
	content := `Morph           	wolf
Vnum              1000
Level        	25
End

Morph           	bear
Vnum              1001
Level        	30
Armor        	-20
End

#END
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	morphs, err := LoadMorphs(path)
	if err != nil {
		t.Fatalf("LoadMorphs err = %v", err)
	}
	if len(morphs) != 2 {
		t.Fatalf("LoadMorphs returned %d morphs, want 2", len(morphs))
	}
	if morphs[0].Name != "wolf" || morphs[1].Name != "bear" {
		t.Errorf("names = [%q, %q], want [wolf bear]", morphs[0].Name, morphs[1].Name)
	}
	if morphs[1].AC != -20 {
		t.Errorf("bear AC = %d, want -20", morphs[1].AC)
	}
}

func TestLoadMorphs_ImplicitEndOnEOF(t *testing.T) {
	// C treats feof as implicit #END at polymorph.c:1776. Missing
	// trailing #END must not error out.
	path := filepath.Join(t.TempDir(), "morph.dat")
	content := `Morph           	wolf
Vnum              1000
End
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	morphs, err := LoadMorphs(path)
	if err != nil {
		t.Fatalf("LoadMorphs err = %v", err)
	}
	if len(morphs) != 1 {
		t.Fatalf("expected 1 morph, got %d", len(morphs))
	}
}

// ---------------------------------------------------------------------------
// SaveMorphs — basic shape
// ---------------------------------------------------------------------------

func TestSaveMorphs_EmitsEndTerminator(t *testing.T) {
	path := filepath.Join(t.TempDir(), "morph.dat")
	if err := SaveMorphs(path, []*types.MorphData{}); err != nil {
		t.Fatalf("SaveMorphs err = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "#END\n" {
		t.Errorf("empty save produced %q, want %q", string(data), "#END\n")
	}
}

func TestSaveMorphs_OnlyEmitsNonZeroFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "morph.dat")
	m := MorphDefaults()
	m.Name = "testform"
	m.Vnum = 1234
	m.Level = 10
	if err := SaveMorphs(path, []*types.MorphData{m}); err != nil {
		t.Fatalf("SaveMorphs err = %v", err)
	}
	data, _ := os.ReadFile(path)
	text := string(data)
	// Non-zero fields present
	if !strings.Contains(text, "Morph") || !strings.Contains(text, "testform") {
		t.Errorf("save missing Morph/testform line: %q", text)
	}
	if !strings.Contains(text, "Vnum") || !strings.Contains(text, "1234") {
		t.Errorf("save missing Vnum 1234: %q", text)
	}
	if !strings.Contains(text, "Level") {
		t.Errorf("save missing Level: %q", text)
	}
	// Zero/default fields must NOT appear
	if strings.Contains(text, "Armor") {
		t.Errorf("zero Armor should not appear in save: %q", text)
	}
	if strings.Contains(text, "Dexterity") {
		t.Errorf("zero Dexterity should not appear: %q", text)
	}
	// Sentinel defaults (Sex=-1, DefPos=STANDING, Timer=-1) must NOT appear
	if strings.Contains(text, "Sex") {
		t.Errorf("sentinel Sex=-1 should not appear: %q", text)
	}
	if strings.Contains(text, "Defpos") {
		t.Errorf("default Defpos=STANDING should not appear: %q", text)
	}
}

// ---------------------------------------------------------------------------
// Round-trip: save → load → equal
// ---------------------------------------------------------------------------

func TestMorphs_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "morph.dat")
	orig := MorphDefaults()
	orig.Name = "eagle"
	orig.Vnum = 1005
	orig.Level = 20
	orig.AC = -5
	orig.Str = 1
	orig.Dex = 4
	orig.Hit = "1d10+30"
	orig.Damroll = "1d2"
	orig.Hitroll = "1d3"
	orig.Mana = "50"
	orig.Move = "50"
	orig.MorphSelf = "You sprout feathers."
	orig.MorphOther = "$n sprouts feathers."
	orig.UnmorphSelf = "The feathers fade."
	orig.UnmorphOther = "$n's feathers fade."
	orig.Description = "A majestic eagle soars."
	orig.KeyWords = "eagle bird wings"
	orig.ShortDesc = "an eagle"
	orig.LongDesc = "A huge eagle is here."
	orig.Help = "The eagle form grants flight."
	orig.Skills = "fly soar"
	orig.NoSkills = "wield"
	orig.DayFrom = 5
	orig.DayTo = 10
	orig.TimeFrom = 6
	orig.TimeTo = 18
	orig.Sex = 0
	orig.HpUsed = 50
	orig.MoveUsed = 25
	orig.ManaUsed = 40
	orig.Obj[0] = 2000
	orig.Obj[1] = 2001
	orig.ObjUse[0] = true
	orig.Used = 42
	orig.Immune = 3
	orig.Resistant = 7
	orig.Suscept = 1
	orig.NoImmune = 2
	orig.NoResistant = 4
	orig.NoSuscept = 8
	orig.SavingBreath = -5
	orig.SavingParaPetri = -3
	orig.SavingPoisonDeath = -2
	orig.SavingSpellStaff = -1
	orig.SavingWand = -4
	orig.Parry = 5
	orig.Dodge = 10
	orig.Tumble = 3
	orig.NoCast = true

	if err := SaveMorphs(path, []*types.MorphData{orig}); err != nil {
		t.Fatalf("SaveMorphs err = %v", err)
	}
	loaded, err := LoadMorphs(path)
	if err != nil {
		t.Fatalf("LoadMorphs err = %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("LoadMorphs returned %d morphs, want 1", len(loaded))
	}
	got := loaded[0]
	checks := []struct {
		name string
		a, b any
	}{
		{"Name", orig.Name, got.Name},
		{"Vnum", orig.Vnum, got.Vnum},
		{"Level", orig.Level, got.Level},
		{"AC", orig.AC, got.AC},
		{"Str", orig.Str, got.Str},
		{"Dex", orig.Dex, got.Dex},
		{"Hit", orig.Hit, got.Hit},
		{"Damroll", orig.Damroll, got.Damroll},
		{"Hitroll", orig.Hitroll, got.Hitroll},
		{"Mana", orig.Mana, got.Mana},
		{"Move", orig.Move, got.Move},
		{"MorphSelf", orig.MorphSelf, got.MorphSelf},
		{"MorphOther", orig.MorphOther, got.MorphOther},
		{"UnmorphSelf", orig.UnmorphSelf, got.UnmorphSelf},
		{"UnmorphOther", orig.UnmorphOther, got.UnmorphOther},
		{"Description", orig.Description, got.Description},
		{"KeyWords", orig.KeyWords, got.KeyWords},
		{"ShortDesc", orig.ShortDesc, got.ShortDesc},
		{"LongDesc", orig.LongDesc, got.LongDesc},
		{"Help", orig.Help, got.Help},
		{"Skills", orig.Skills, got.Skills},
		{"NoSkills", orig.NoSkills, got.NoSkills},
		{"DayFrom", orig.DayFrom, got.DayFrom},
		{"DayTo", orig.DayTo, got.DayTo},
		{"TimeFrom", orig.TimeFrom, got.TimeFrom},
		{"TimeTo", orig.TimeTo, got.TimeTo},
		{"Sex", orig.Sex, got.Sex},
		{"HpUsed", orig.HpUsed, got.HpUsed},
		{"MoveUsed", orig.MoveUsed, got.MoveUsed},
		{"ManaUsed", orig.ManaUsed, got.ManaUsed},
		{"Obj[0]", orig.Obj[0], got.Obj[0]},
		{"Obj[1]", orig.Obj[1], got.Obj[1]},
		{"ObjUse[0]", orig.ObjUse[0], got.ObjUse[0]},
		{"Used", orig.Used, got.Used},
		{"Immune", orig.Immune, got.Immune},
		{"Resistant", orig.Resistant, got.Resistant},
		{"Suscept", orig.Suscept, got.Suscept},
		{"NoImmune", orig.NoImmune, got.NoImmune},
		{"NoResistant", orig.NoResistant, got.NoResistant},
		{"NoSuscept", orig.NoSuscept, got.NoSuscept},
		{"SavingBreath", orig.SavingBreath, got.SavingBreath},
		{"Parry", orig.Parry, got.Parry},
		{"Dodge", orig.Dodge, got.Dodge},
		{"Tumble", orig.Tumble, got.Tumble},
		{"NoCast", orig.NoCast, got.NoCast},
	}
	for _, c := range checks {
		if c.a != c.b {
			t.Errorf("%s: round-trip mismatch: orig=%v got=%v", c.name, c.a, c.b)
		}
	}
}

// TestLoadMorphs_AffectedByRoundTrip tests BitVector field round-trip.
func TestLoadMorphs_AffectedByRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "morph.dat")
	orig := MorphDefaults()
	orig.Name = "ghost"
	orig.AffectedBy[0] = 0x0001 | 0x0004 // arbitrary bits
	orig.AffectedBy[1] = 0x0100
	if err := SaveMorphs(path, []*types.MorphData{orig}); err != nil {
		t.Fatalf("SaveMorphs err = %v", err)
	}
	loaded, err := LoadMorphs(path)
	if err != nil || len(loaded) != 1 {
		t.Fatalf("load: err=%v n=%d", err, len(loaded))
	}
	if !loaded[0].AffectedBy.Equal(orig.AffectedBy) {
		t.Errorf("AffectedBy: got %v want %v", loaded[0].AffectedBy, orig.AffectedBy)
	}
}

// TestLoadMorphs_UnknownKeyAbandonsRecord documents C's "bail on unknown
// key" policy (polymorph.c:2004). Record with a bad key is dropped; the
// next valid Morph block still loads.
func TestLoadMorphs_UnknownKeyAbandonsRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "morph.dat")
	content := `Morph           	broken
Vnum              999
Gobbledygook	42
End

Morph           	good
Vnum              1001
End

#END
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	morphs, err := LoadMorphs(path)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	// The "broken" record is abandoned; "good" survives.
	if len(morphs) != 1 || morphs[0].Name != "good" {
		t.Errorf("expected 1 morph 'good', got %d (%v)", len(morphs),
			morphNames(morphs))
	}
}

func morphNames(ms []*types.MorphData) []string {
	out := []string{}
	for _, m := range ms {
		out = append(out, m.Name)
	}
	return out
}

// TestWriteMorph_VnumBeforeName_PfileOrderDoesNotApplyHere documents that
// the morph.dat writer emits Morph <name> header FIRST, then fields in
// a fixed order — no constraint like the #MorphData pfile Vnum-before-Name
// ordering (which IS constrained, tested in player_morph_test.go).
func TestWriteMorph_NameAppearsOnFirstLine(t *testing.T) {
	var buf bytes.Buffer
	m := MorphDefaults()
	m.Name = "dragon"
	m.Vnum = 2000
	if err := writeMorph(&buf, m); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(buf.String(), "\n")
	if !strings.HasPrefix(lines[0], "Morph") || !strings.Contains(lines[0], "dragon") {
		t.Errorf("first line = %q, want Morph ... dragon", lines[0])
	}
}

// ---------------------------------------------------------------------------
// SetupMorphVnum — plan §G1, mirrors C setup_morph_vnum at
// src/polymorph.c:2630-2652.
// ---------------------------------------------------------------------------

// mockWorldMorphs is a minimal worldMorphs implementation for testing
// SetupMorphVnum without importing world.
type mockWorldMorphs struct {
	morphs  []*types.MorphData
	counter int
}

func (m *mockWorldMorphs) GetMorphs() []*types.MorphData { return m.morphs }
func (m *mockWorldMorphs) SetMorphVnumCounter(v int)     { m.counter = v }
func (m *mockWorldMorphs) GetMorphVnumCounter() int      { return m.counter }

func TestSetupMorphVnum_EmptyTableSeeds1000(t *testing.T) {
	w := &mockWorldMorphs{}
	SetupMorphVnum(w)
	if w.counter != 1000 {
		t.Errorf("counter = %d, want 1000", w.counter)
	}
}

func TestSetupMorphVnum_AssignsToZeroVnums(t *testing.T) {
	w := &mockWorldMorphs{
		morphs: []*types.MorphData{
			{Name: "a", Vnum: 0},
			{Name: "b", Vnum: 1005},
			{Name: "c", Vnum: 0},
		},
	}
	SetupMorphVnum(w)
	// max(0,0,1005) = 1005; since >= 1000, start at 1005+1 = 1006.
	if w.morphs[0].Vnum != 1006 {
		t.Errorf("morph a Vnum = %d, want 1006", w.morphs[0].Vnum)
	}
	if w.morphs[2].Vnum != 1007 {
		t.Errorf("morph c Vnum = %d, want 1007", w.morphs[2].Vnum)
	}
	// Counter past last assignment.
	if w.counter != 1008 {
		t.Errorf("counter = %d, want 1008", w.counter)
	}
	// Existing non-zero vnum is unchanged.
	if w.morphs[1].Vnum != 1005 {
		t.Errorf("morph b Vnum changed to %d, want 1005", w.morphs[1].Vnum)
	}
}

func TestSetupMorphVnum_PreservesMaxAbove1000(t *testing.T) {
	w := &mockWorldMorphs{
		morphs: []*types.MorphData{{Name: "x", Vnum: 1500}, {Name: "y", Vnum: 0}},
	}
	SetupMorphVnum(w)
	if w.morphs[1].Vnum != 1501 {
		t.Errorf("y.Vnum = %d, want 1501", w.morphs[1].Vnum)
	}
}

func TestSetupMorphVnum_NilWorldNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SetupMorphVnum(nil) panicked: %v", r)
		}
	}()
	SetupMorphVnum(nil)
}

// TestSaveMorphs_FileMode pins 0600 — security adversary 2026-04-19.
func TestSaveMorphs_FileMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "morph.dat")
	if err := SaveMorphs(path, nil); err != nil {
		t.Fatalf("SaveMorphs: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("morph.dat mode = %#o, want 0o600", mode)
	}
}
