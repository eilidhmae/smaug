package persist

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// ClassNameLookup, if non-nil, resolves a class who_name to its class-table
// index. Set at boot; left nil in isolated unit tests (in which case Class
// tokens in morph records are silently dropped with a bug log).
var ClassNameLookup func(name string) int

// RaceNameLookup, if non-nil, resolves a race name to its race-table index.
// Same lifecycle as ClassNameLookup.
var RaceNameLookup func(name string) int

// LoadMorphs reads a SMAUG morph.dat file into a slice. Missing file returns
// `(nil, nil)` (matches C load_morphs at src/polymorph.c:1757-1814:
// missing file is non-fatal, bug-logged, and boot continues).
//
// Malformed records log via util.Bug and are skipped (the record is
// abandoned at the unknown key and the loader continues with the next
// `Morph <name>` block). File terminator is `#END`; EOF before `#END` is
// treated as implicit terminator (matches C's `feof (fp) ? "#END"`
// short-circuit at polymorph.c:1776).
//
// Plan plan-phase6-polymorph.md §G1.
func LoadMorphs(path string) ([]*types.MorphData, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	sc := NewScanner(f, path)
	list := make([]*types.MorphData, 0)
	pendingMorphs = pendingMorphs[:0]
	for {
		word := sc.ReadWord()
		if word == "" || word == "#END" {
			if len(pendingMorphs) > 0 {
				list = append(list, pendingMorphs...)
				pendingMorphs = pendingMorphs[:0]
			}
			return list, nil
		}
		if word == "Morph" {
			m := readMorphBlock(sc, path)
			if m != nil {
				list = append(list, m)
			}
			if len(pendingMorphs) > 0 {
				list = append(list, pendingMorphs...)
				pendingMorphs = pendingMorphs[:0]
			}
			continue
		}
		util.Bug("LoadMorphs: %s: unknown top-level token %q", path, word)
	}
}

// readMorphBlock reads one morph record's key/value pairs up to End.
// The leading `Morph <name>` token was already consumed by the caller.
// Mirrors C fread_morph at src/polymorph.c:1820-2013.
//
// The first word AFTER "Morph" is the morph's Name. Then a loop of keys
// terminates on "End". Unknown keys cause the record to be abandoned
// (C free_morph's early-bail policy at :2004-2010 — "better than possibly
// having the memory messed up").
func readMorphBlock(sc *Scanner, path string) *types.MorphData {
	name := sc.ReadWord()
	if name == "" || name == "End" {
		return nil
	}
	m := newMorphWithDefaults()
	m.Name = name
	for {
		word := sc.ReadWord()
		if word == "" {
			util.Bug("LoadMorphs: %s: unexpected EOF inside morph %q", path, m.Name)
			return m
		}
		if word == "End" {
			return m
		}
		if !morphKey(sc, m, word) {
			util.Bug("Fread_morph: no match: %s (morph %q)", word, m.Name)
			// Abandon record — matches C :2004. Drain tokens to a
			// structural boundary so the orphan value doesn't bleed
			// over into the next top-level word.
			for {
				w := sc.ReadWord()
				if w == "" || w == "End" || w == "Morph" || w == "#END" {
					// Can't push back — but the outer caller loops
					// by reading the next word, which will be the
					// next Morph/#END or EOF. If we stopped on Morph,
					// we've consumed its header; readMorphBlock for
					// the next record will still pick up a Name from
					// ReadWord (the morph's name). Actually NO — we
					// just consumed "Morph" here. Handle by reading
					// the next morph inline.
					if w == "Morph" {
						// The outer loop expects to consume "Morph"
						// itself. Recurse once to read the recovery
						// morph record, then we need a way to append
						// it. Simplest: stash a drained-morph pointer
						// on a package-level seam... no. Instead,
						// return via a sentinel by setting fields on
						// m that the outer loop checks. Cleanest: we
						// actually recurse and LoadMorphs appends m
						// only if non-nil — so we need to push the
						// recovered morph into the outer list via a
						// different path.
						//
						// Pragmatic fix: return the recovered morph
						// through a closure variable. See the rework
						// in LoadMorphs below.
						drained := readMorphBlock(sc, path)
						if drained != nil {
							pendingMorphs = append(pendingMorphs, drained)
						}
					}
					return nil
				}
			}
		}
	}
}

// pendingMorphs is a transient queue used when recovering from an
// unknown-key-abandoned record: the recovery scanner consumes the
// next "Morph" header itself, so it must stash any follow-on records
// here for the outer LoadMorphs to pick up. Cleared at LoadMorphs
// entry. Not safe for concurrent load — matches C's single-threaded
// load semantics.
var pendingMorphs []*types.MorphData

// morphKey dispatches one key-value pair for a morph record. Returns
// true if the key was recognized. Mirrors the case/KEY dispatch in C
// fread_morph at src/polymorph.c:1844-1999.
func morphKey(sc *Scanner, m *types.MorphData, word string) bool {
	switch word {
	// A
	case "Armor":
		m.AC = sc.ReadNumber()
	case "Affected":
		m.AffectedBy = sc.ReadBitvector()
	// B
	case "Blood":
		m.Blood = sc.ReadString()
	case "BloodUsed":
		m.BloodUsed = sc.ReadNumber()
	// C
	case "Charisma":
		m.Cha = sc.ReadNumber()
	case "Class":
		// String of space-separated class who_names; resolve each to
		// the class index and OR (1<<index) into Class. Mirrors
		// polymorph.c:1856-1871.
		s := sc.ReadString()
		for _, tok := range strings.Fields(s) {
			if ClassNameLookup != nil {
				idx := ClassNameLookup(tok)
				if idx >= 0 {
					m.Class |= 1 << uint(idx)
				}
			}
		}
	case "Constitution":
		m.Con = sc.ReadNumber()
	// D
	case "Damroll":
		m.Damroll = sc.ReadString()
	case "DayFrom":
		m.DayFrom = sc.ReadNumber()
	case "DayTo":
		m.DayTo = sc.ReadNumber()
	case "Defpos":
		m.DefPos = sc.ReadNumber()
	case "Description":
		m.Description = sc.ReadString()
	case "Dexterity":
		m.Dex = sc.ReadNumber()
	case "Dodge":
		m.Dodge = sc.ReadNumber()
	// F
	case "FavourUsed":
		m.FavourUsed = sc.ReadNumber()
	// G
	case "GloryUsed":
		m.GloryUsed = sc.ReadNumber()
	// H
	case "Help":
		m.Help = sc.ReadString()
	case "Hit":
		m.Hit = sc.ReadString()
	case "Hitroll":
		m.Hitroll = sc.ReadString()
	case "HpUsed":
		m.HpUsed = sc.ReadNumber()
	// I
	case "Intelligence":
		m.Int = sc.ReadNumber()
	case "Immune":
		m.Immune = sc.ReadNumber()
	// K
	case "Keywords":
		m.KeyWords = sc.ReadString()
	// L
	case "Level":
		m.Level = sc.ReadNumber()
	case "Longdesc":
		m.LongDesc = sc.ReadString()
	case "Luck":
		m.Lck = sc.ReadNumber()
	// M
	case "Mana":
		m.Mana = sc.ReadString()
	case "ManaUsed":
		m.ManaUsed = sc.ReadNumber()
	case "MorphOther":
		m.MorphOther = sc.ReadString()
	case "MorphSelf":
		m.MorphSelf = sc.ReadString()
	case "Move":
		// C bug preserved: polymorph.c:1916 writes `morph->morph_self`
		// instead of `morph->move` under "Move" key (same function
		// is assigned to both). We read into `Move` (the field users
		// actually expect to hold the move expression) — divergence
		// from C that matches what the fwrite side emits.
		m.Move = sc.ReadString()
	case "MoveUsed":
		m.MoveUsed = sc.ReadNumber()
	// N
	case "NoAffected":
		m.NoAffectedBy = sc.ReadBitvector()
	case "NoImmune":
		m.NoImmune = sc.ReadNumber()
	case "NoResistant":
		m.NoResistant = sc.ReadNumber()
	case "NoSkills":
		m.NoSkills = sc.ReadString()
	case "NoSuscept":
		m.NoSuscept = sc.ReadNumber()
	case "NoCast":
		m.NoCast = sc.ReadNumber() != 0
	// O
	case "Objs":
		m.Obj[0] = sc.ReadNumber()
		m.Obj[1] = sc.ReadNumber()
		m.Obj[2] = sc.ReadNumber()
	case "Objuse":
		m.ObjUse[0] = sc.ReadNumber() != 0
		m.ObjUse[1] = sc.ReadNumber() != 0
		m.ObjUse[2] = sc.ReadNumber() != 0
	// P
	case "Parry":
		m.Parry = sc.ReadNumber()
	case "Pkill":
		m.PKill = sc.ReadNumber()
	// R
	case "Race":
		s := sc.ReadString()
		for _, tok := range strings.Fields(s) {
			if RaceNameLookup != nil {
				idx := RaceNameLookup(tok)
				if idx >= 0 {
					m.Race |= 1 << uint(idx)
				}
			}
		}
	case "Resistant":
		m.Resistant = sc.ReadNumber()
	// S
	case "SaveBreath":
		m.SavingBreath = sc.ReadNumber()
	case "SavePara":
		m.SavingParaPetri = sc.ReadNumber()
	case "SavePoison":
		m.SavingPoisonDeath = sc.ReadNumber()
	case "SaveSpell":
		m.SavingSpellStaff = sc.ReadNumber()
	case "SaveWand":
		m.SavingWand = sc.ReadNumber()
	case "Sex":
		m.Sex = sc.ReadNumber()
	case "ShortDesc":
		m.ShortDesc = sc.ReadString()
	case "Skills":
		m.Skills = sc.ReadString()
	case "Strength":
		m.Str = sc.ReadNumber()
	case "Suscept":
		m.Suscept = sc.ReadNumber()
	// T
	case "TimeFrom":
		m.TimeFrom = sc.ReadNumber()
	case "TimeTo":
		m.TimeTo = sc.ReadNumber()
	case "Tumble":
		m.Tumble = sc.ReadNumber()
	// U
	case "UnmorphOther":
		m.UnmorphOther = sc.ReadString()
	case "UnmorphSelf":
		m.UnmorphSelf = sc.ReadString()
	case "Used":
		m.Used = sc.ReadNumber()
	// V
	case "Vnum":
		m.Vnum = sc.ReadNumber()
	// W
	case "Wisdom":
		m.Wis = sc.ReadNumber()
	default:
		return false
	}
	return true
}

// newMorphWithDefaults returns a fresh MorphData initialized with the
// defaults from C morph_defaults at src/polymorph.c:2068-2140.
func newMorphWithDefaults() *types.MorphData {
	return &types.MorphData{
		Sex:      -1,
		TimeFrom: -1,
		TimeTo:   -1,
		DayFrom:  -1,
		DayTo:    -1,
		DefPos:   types.POS_STANDING,
		Timer:    -1,
	}
}

// MorphDefaults returns a zero-valued MorphData with the C-default
// sentinel fields populated. Used by DoMorphcreate and by the loader.
func MorphDefaults() *types.MorphData {
	return newMorphWithDefaults()
}

// worldMorphs is the minimal interface SetupMorphVnum needs from *world.World
// (avoids an import cycle: persist can't import world).
type worldMorphs interface {
	GetMorphs() []*types.MorphData
	SetMorphVnumCounter(int)
	GetMorphVnumCounter() int
}

// SetupMorphVnum assigns fresh vnums (≥1000) to every morph that has
// vnum==0, and updates the world's MorphVnumCounter so future
// morphcreate calls hand out unique values. Mirrors C setup_morph_vnum
// at src/polymorph.c:2630-2652.
//
// Algorithm: find max(existing vnum, 999)+1 as the starting point. Then
// assign vnums sequentially to any morph with vnum==0. Final counter
// is one past the last assigned vnum.
//
// The world parameter is `any` to avoid an import cycle with the world
// package — callers pass *world.World and we type-assert to a small
// interface. If the argument doesn't implement the interface, we bail
// silently (defensive: tests may pass nil).
func SetupMorphVnum(w any) {
	wm, ok := w.(worldMorphs)
	if !ok || wm == nil {
		return
	}
	morphs := wm.GetMorphs()
	vnum := wm.GetMorphVnumCounter()
	for _, m := range morphs {
		if m != nil && m.Vnum > vnum {
			vnum = m.Vnum
		}
	}
	if vnum < 1000 {
		vnum = 1000
	} else {
		vnum++
	}
	for _, m := range morphs {
		if m != nil && m.Vnum == 0 {
			m.Vnum = vnum
			vnum++
		}
	}
	wm.SetMorphVnumCounter(vnum)
}

// SaveMorphs writes the morph table to path in the C-compatible format.
// Mirrors C save_morphs at src/polymorph.c:1586-1605 + fwrite_morph at
// :1613-1748. Only non-zero/non-default fields are emitted (per C's
// "why waste disk-space" convention).
//
// Terminator: `#END\n`.
func SaveMorphs(path string, list []*types.MorphData) error {
	// 0o600 — private to the smaug user (hotboot precedent).
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, m := range list {
		if m == nil {
			continue
		}
		if err := writeMorph(f, m); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(f, "#END"); err != nil {
		return err
	}
	return nil
}

// writeMorph emits one morph record in C fwrite_morph format.
func writeMorph(w io.Writer, m *types.MorphData) error {
	fmt.Fprintf(w, "Morph           \t%s\n", m.Name)
	if m.Obj[0] != 0 || m.Obj[1] != 0 || m.Obj[2] != 0 {
		fmt.Fprintf(w, "Objs  \t\t\t%d %d %d\n", m.Obj[0], m.Obj[1], m.Obj[2])
	}
	if m.ObjUse[0] || m.ObjUse[1] || m.ObjUse[2] {
		fmt.Fprintf(w, "Objuse\t\t\t%d %d %d\n",
			boolToInt(m.ObjUse[0]), boolToInt(m.ObjUse[1]), boolToInt(m.ObjUse[2]))
	}
	if m.Vnum != 0 {
		fmt.Fprintf(w, "Vnum              %d\n", m.Vnum)
	}
	if m.Blood != "" {
		fmt.Fprintf(w, "Blood    \t        %s~\n", m.Blood)
	}
	if m.Damroll != "" {
		fmt.Fprintf(w, "Damroll         \t%s~\n", m.Damroll)
	}
	if m.DefPos != types.POS_STANDING {
		fmt.Fprintf(w, "Defpos            %d\n", m.DefPos)
	}
	if m.Description != "" {
		fmt.Fprintf(w, "Description     \t%s~\n", m.Description)
	}
	if m.Help != "" {
		fmt.Fprintf(w, "Help            \t%s~\n", m.Help)
	}
	if m.Hit != "" {
		fmt.Fprintf(w, "Hit             \t%s~\n", m.Hit)
	}
	if m.Hitroll != "" {
		fmt.Fprintf(w, "Hitroll         \t%s~\n", m.Hitroll)
	}
	if m.KeyWords != "" {
		fmt.Fprintf(w, "Keywords        \t%s~\n", m.KeyWords)
	}
	if m.LongDesc != "" {
		fmt.Fprintf(w, "Longdesc        \t%s~\n", m.LongDesc)
	}
	if m.Mana != "" {
		fmt.Fprintf(w, "Mana            \t%s~\n", m.Mana)
	}
	if m.MorphOther != "" {
		fmt.Fprintf(w, "MorphOther      \t%s~\n", m.MorphOther)
	}
	if m.MorphSelf != "" {
		fmt.Fprintf(w, "MorphSelf       \t%s~\n", m.MorphSelf)
	}
	if m.Move != "" {
		fmt.Fprintf(w, "Move            \t%s~\n", m.Move)
	}
	if m.NoSkills != "" {
		fmt.Fprintf(w, "NoSkills         %s~\n", m.NoSkills)
	}
	if m.ShortDesc != "" {
		fmt.Fprintf(w, "ShortDesc       \t%s~\n", m.ShortDesc)
	}
	if m.Skills != "" {
		fmt.Fprintf(w, "Skills          \t%s~\n", m.Skills)
	}
	if m.UnmorphOther != "" {
		fmt.Fprintf(w, "UnmorphOther    \t%s~\n", m.UnmorphOther)
	}
	if m.UnmorphSelf != "" {
		fmt.Fprintf(w, "UnmorphSelf     \t%s~\n", m.UnmorphSelf)
	}
	if !m.AffectedBy.IsEmpty() {
		fmt.Fprintf(w, "Affected        \t%s\n", m.AffectedBy.String())
	}
	if m.Class != 0 {
		fmt.Fprintf(w, "Class           \t%s~\n", morphClassString(m.Class))
	}
	if m.Immune != 0 {
		fmt.Fprintf(w, "Immune          \t%d\n", m.Immune)
	}
	if !m.NoAffectedBy.IsEmpty() {
		fmt.Fprintf(w, "NoAffected      \t%s\n", m.NoAffectedBy.String())
	}
	if m.NoImmune != 0 {
		fmt.Fprintf(w, "NoImmune        \t%d\n", m.NoImmune)
	}
	if m.NoResistant != 0 {
		fmt.Fprintf(w, "NoResistant     \t%d\n", m.NoResistant)
	}
	if m.NoSuscept != 0 {
		fmt.Fprintf(w, "NoSuscept       \t%d\n", m.NoSuscept)
	}
	if m.Race != 0 {
		fmt.Fprintf(w, "Race        \t%s~\n", morphRaceString(m.Race))
	}
	if m.Resistant != 0 {
		fmt.Fprintf(w, "Resistant       \t%d\n", m.Resistant)
	}
	if m.Suscept != 0 {
		fmt.Fprintf(w, "Suscept        \t%d\n", m.Suscept)
	}
	if m.Used != 0 {
		fmt.Fprintf(w, "Used       \t%d\n", m.Used)
	}
	if m.Sex != -1 {
		fmt.Fprintf(w, "Sex\t\t%d\n", m.Sex)
	}
	if m.PKill != 0 {
		fmt.Fprintf(w, "Pkill\t\t%d\n", m.PKill)
	}
	if m.TimeFrom != -1 {
		fmt.Fprintf(w, "TimeFrom\t\t%d\n", m.TimeFrom)
	}
	if m.TimeTo != -1 {
		fmt.Fprintf(w, "TimeTo\t\t%d\n", m.TimeTo)
	}
	if m.DayFrom != -1 {
		fmt.Fprintf(w, "DayFrom\t\t%d\n", m.DayFrom)
	}
	if m.DayTo != -1 {
		fmt.Fprintf(w, "DayTo\t\t%d\n", m.DayTo)
	}
	if m.BloodUsed != 0 {
		fmt.Fprintf(w, "BloodUsed\t%d\n", m.BloodUsed)
	}
	if m.ManaUsed != 0 {
		fmt.Fprintf(w, "ManaUsed\t\t%d\n", m.ManaUsed)
	}
	if m.MoveUsed != 0 {
		fmt.Fprintf(w, "MoveUsed\t\t%d\n", m.MoveUsed)
	}
	if m.HpUsed != 0 {
		fmt.Fprintf(w, "HpUsed\t\t%d\n", m.HpUsed)
	}
	if m.FavourUsed != 0 {
		fmt.Fprintf(w, "FavourUsed\t%d\n", m.FavourUsed)
	}
	if m.GloryUsed != 0 {
		fmt.Fprintf(w, "GloryUsed\t%d\n", m.GloryUsed)
	}
	if m.AC != 0 {
		fmt.Fprintf(w, "Armor        \t%d\n", m.AC)
	}
	if m.Cha != 0 {
		fmt.Fprintf(w, "Charisma        \t%d\n", m.Cha)
	}
	if m.Con != 0 {
		fmt.Fprintf(w, "Constitution    \t%d\n", m.Con)
	}
	if m.Dex != 0 {
		fmt.Fprintf(w, "Dexterity       \t%d\n", m.Dex)
	}
	if m.Dodge != 0 {
		fmt.Fprintf(w, "Dodge        \t%d\n", m.Dodge)
	}
	if m.Int != 0 {
		fmt.Fprintf(w, "Intelligence    \t%d\n", m.Int)
	}
	if m.Lck != 0 {
		fmt.Fprintf(w, "Luck        \t%d\n", m.Lck)
	}
	if m.Level != 0 {
		fmt.Fprintf(w, "Level        \t%d\n", m.Level)
	}
	if m.Parry != 0 {
		fmt.Fprintf(w, "Parry        \t%d\n", m.Parry)
	}
	if m.SavingBreath != 0 {
		fmt.Fprintf(w, "SaveBreath      \t%d\n", m.SavingBreath)
	}
	if m.SavingParaPetri != 0 {
		fmt.Fprintf(w, "SavePara        \t%d\n", m.SavingParaPetri)
	}
	if m.SavingPoisonDeath != 0 {
		fmt.Fprintf(w, "SavePoison      \t%d\n", m.SavingPoisonDeath)
	}
	if m.SavingSpellStaff != 0 {
		fmt.Fprintf(w, "SaveSpell       \t%d\n", m.SavingSpellStaff)
	}
	if m.SavingWand != 0 {
		fmt.Fprintf(w, "SaveWand        \t%d\n", m.SavingWand)
	}
	if m.Str != 0 {
		fmt.Fprintf(w, "Strength        \t%d\n", m.Str)
	}
	if m.Tumble != 0 {
		fmt.Fprintf(w, "Tumble          \t%d\n", m.Tumble)
	}
	if m.Wis != 0 {
		fmt.Fprintf(w, "Wisdom          \t%d\n", m.Wis)
	}
	if m.NoCast {
		fmt.Fprintf(w, "NoCast          \t1\n")
	}
	if _, err := fmt.Fprintf(w, "End\n\n"); err != nil {
		return err
	}
	return nil
}

// morphClassString formats the Class bitmask as a space-separated list of
// class who_names. Reverse of the Class loader; uses ClassNameFormatter if
// available, else emits numeric tokens. The emitted form round-trips via
// ClassNameLookup on reload.
var ClassNameFormatter func(int) string
var RaceNameFormatter func(int) string

func morphClassString(mask int) string {
	var out []string
	for i := 0; i < 32; i++ {
		if mask&(1<<uint(i)) == 0 {
			continue
		}
		if ClassNameFormatter != nil {
			if s := ClassNameFormatter(i); s != "" {
				out = append(out, s)
				continue
			}
		}
		// Fallback: emit numeric index as `#N` so the reloader (which
		// looks up by class who_name) silently drops the bit. Tests
		// that exercise Class round-trip must register a
		// ClassNameFormatter.
		out = append(out, fmt.Sprintf("#%d", i))
	}
	return strings.Join(out, " ")
}

func morphRaceString(mask int) string {
	var out []string
	for i := 0; i < 32; i++ {
		if mask&(1<<uint(i)) == 0 {
			continue
		}
		if RaceNameFormatter != nil {
			if s := RaceNameFormatter(i); s != "" {
				out = append(out, s)
				continue
			}
		}
		out = append(out, fmt.Sprintf("#%d", i))
	}
	return strings.Join(out, " ")
}
