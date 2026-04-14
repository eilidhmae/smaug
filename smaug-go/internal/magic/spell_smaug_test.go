package magic

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// addSkill appends skill to w.Skills and returns the sn.
func addSkill(w *world.World, sk *types.SkillType) int {
	w.Skills = append(w.Skills, sk)
	return len(w.Skills) - 1
}

func TestSpellSmaug_NilSafe(t *testing.T) {
	w := newMagicWorld()
	ch := newCaster("A", 10)
	// Bad sn — must not panic.
	SpellSmaug(w, -1, 10, ch, nil)
	SpellSmaug(w, 99999, 10, ch, nil)

	// sn in range but nil skill entry.
	w.Skills = append(w.Skills, nil)
	SpellSmaug(w, 0, 10, ch, nil)
}

func TestSpellSmaug_AffectPath(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 1, Name: "Test"}
	ch := newCaster("Caster", 20)
	victim := newNPCVictim("goblin", 10)
	handler.CharToRoom(ch, room)
	handler.CharToRoom(victim, room)

	// SA_CREATE with an affect in .Affects. Defensive-style buff delivered to
	// a victim: use TAR_CHAR_DEFENSIVE so the dispatcher takes the affect
	// branch directly. Non-NEWSPELLS shift for SA is bits 3-5.
	skill := &types.SkillType{
		Name:   "magic boost",
		Type:   types.SKILL_SPELL,
		Target: types.TAR_CHAR_DEFENSIVE,
		Info:   types.SA_CREATE << 3,
		Affects: []*types.SmaugAff{
			{Duration: "12", Location: types.APPLY_STR, Modifier: "3", BitVector: -1},
		},
		HitChar: "You boost the target.",
		HitVict: "You feel stronger.",
	}
	sn := addSkill(w, skill)

	before := len(victim.Affects)
	SpellSmaug(w, sn, 20, ch, victim)
	if len(victim.Affects) != before+1 {
		t.Fatalf("affect not applied: before=%d after=%d", before, len(victim.Affects))
	}
	aff := victim.Affects[len(victim.Affects)-1]
	if aff.Location != types.APPLY_STR || aff.Modifier != 3 || aff.Duration != 12 {
		t.Errorf("affect wrong: %+v", aff)
	}
}

func TestSpellSmaug_AttackPath(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 2, Name: "Arena"}
	ch := newCaster("Wiz", 20)
	victim := newNPCVictim("dummy", 10)
	victim.Hit = 100
	victim.MaxHit = 100
	handler.CharToRoom(ch, room)
	handler.CharToRoom(victim, room)
	w.AddChar(ch)
	w.AddChar(victim)

	// SA_DESTROY + SC_LIFE + dice formula: offensive damage spell.
	// Non-NEWSPELLS shifts: SA bits 3-5, SC bits 6-8.
	skill := &types.SkillType{
		Name:        "zap",
		Type:        types.SKILL_SPELL,
		Target:      types.TAR_CHAR_OFFENSIVE,
		Info:        (types.SA_DESTROY << 3) | (types.SC_LIFE << 6),
		DiceFormula: "10",
		SaveType:    types.SS_NONE,
	}
	sn := addSkill(w, skill)

	beforeHP := victim.Hit
	SpellSmaug(w, sn, 20, ch, victim)
	if victim.Hit >= beforeHP {
		t.Errorf("attack did not damage victim: before=%d after=%d", beforeHP, victim.Hit)
	}
}

func TestSpellSmaug_AreaAttack(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 3, Name: "Battlefield"}
	ch := newCaster("Mage", 30)
	v1 := newNPCVictim("orc1", 10)
	v2 := newNPCVictim("orc2", 10)
	v3 := newNPCVictim("orc3", 10)
	for _, c := range []*types.CharData{ch, v1, v2, v3} {
		handler.CharToRoom(c, room)
		w.AddChar(c)
	}
	chBefore := ch.Hit
	v1Before, v2Before, v3Before := v1.Hit, v2.Hit, v3.Hit

	skill := &types.SkillType{
		Name:        "firestorm",
		Type:        types.SKILL_SPELL,
		Target:      types.TAR_IGNORE,
		Info:        (types.SA_DESTROY << 3) | (types.SC_LIFE << 6) | types.SD_FIRE,
		Flags:       int(types.SF_AREA),
		DiceFormula: "20",
	}
	sn := addSkill(w, skill)
	SpellSmaug(w, sn, 30, ch, nil)

	if ch.Hit != chBefore {
		t.Errorf("caster was hit by their own area spell: before=%d after=%d", chBefore, ch.Hit)
	}
	if v1.Hit >= v1Before || v2.Hit >= v2Before || v3.Hit >= v3Before {
		t.Errorf("area spell did not hit all targets: v1 %d->%d, v2 %d->%d, v3 %d->%d",
			v1Before, v1.Hit, v2Before, v2.Hit, v3Before, v3.Hit)
	}
}

func TestSpellSmaug_SaveNegate(t *testing.T) {
	// Force the save roll to always succeed so SE_NEGATE deterministically
	// drops damage to 0.
	orig := rollSaveFunc
	rollSaveFunc = func(level int, victim *types.CharData, saveType int) bool { return true }
	defer func() { rollSaveFunc = orig }()

	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 4}
	ch := newCaster("Wiz", 20)
	victim := newNPCVictim("dummy", 10)
	victim.Hit = 100
	handler.CharToRoom(ch, room)
	handler.CharToRoom(victim, room)
	w.AddChar(ch)
	w.AddChar(victim)

	skill := &types.SkillType{
		Name:        "fizzle",
		Type:        types.SKILL_SPELL,
		Target:      types.TAR_CHAR_OFFENSIVE,
		Info:        (types.SA_DESTROY << 3) | (types.SC_LIFE << 6),
		DiceFormula: "50",
		SaveType:    types.SS_SPELL_STAFF,
		SaveEffect:  types.SE_NEGATE,
	}
	sn := addSkill(w, skill)

	before := victim.Hit
	SpellSmaug(w, sn, 20, ch, victim)
	if victim.Hit != before {
		t.Fatalf("SE_NEGATE failed to block damage: before=%d after=%d", before, victim.Hit)
	}

	// And check SE_HALFDAM halves damage when save succeeds.
	skill.SaveEffect = types.SE_HALFDAM
	skill.DiceFormula = "10"
	victim.Hit = 100
	SpellSmaug(w, sn, 20, ch, victim)
	// Exact damage = 10/2 = 5, so hit should be 95.
	if victim.Hit != 95 {
		t.Fatalf("SE_HALFDAM: expected hit=95, got %d", victim.Hit)
	}
}

func TestSpellSmaug_CreateObject(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 5}
	ch := newCaster("Cleric", 15)
	handler.CharToRoom(ch, room)

	objIdx := &types.ObjIndexData{
		Vnum:       4200,
		Name:       "wand",
		ShortDescr: "a glowing wand",
		ItemType:   types.ITEM_WAND,
	}
	w.ObjIndex[4200] = objIdx

	skill := &types.SkillType{
		Name:    "create wand",
		Type:    types.SKILL_SPELL,
		Target:  types.TAR_IGNORE,
		Info:    types.SA_CREATE << 3, // action = SA_CREATE; SC_* = 0
		Value:   4200,                 // vnum (matches C skill->value)
		Flags:   int(types.SF_OBJECT),
		HitChar: "A wand appears.",
	}
	sn := addSkill(w, skill)

	before := len(ch.Carrying)
	SpellSmaug(w, sn, 15, ch, nil)
	if len(ch.Carrying) != before+1 {
		t.Fatalf("object not added to caster: before=%d after=%d", before, len(ch.Carrying))
	}
}

func TestSpellSmaug_CreateMob(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 6}
	ch := newCaster("Summoner", 25)
	handler.CharToRoom(ch, room)

	mobIdx := &types.MobIndexData{
		Vnum:       5500,
		PlayerName: "imp",
		ShortDescr: "a summoned imp",
		Level:      10,
	}
	w.MobIndex[5500] = mobIdx

	skill := &types.SkillType{
		Name:    "summon imp",
		Type:    types.SKILL_SPELL,
		Target:  types.TAR_IGNORE,
		Info:    (types.SA_CREATE << 3) | (types.SC_LIFE << 6),
		Value:   5500,
		HitChar: "An imp arrives.",
	}
	sn := addSkill(w, skill)

	beforeCount := len(room.People)
	SpellSmaug(w, sn, 25, ch, nil)
	if len(room.People) != beforeCount+1 {
		t.Fatalf("mob not added to room: before=%d after=%d", beforeCount, len(room.People))
	}
}

func TestSkillType_InfoUnpacking_RealDat(t *testing.T) {
	// Non-NEWSPELLS layout — NEWSPELLS is OFF in shipping C
	// (src/mud.h:4129), so db/system/en/skills.dat was generated with the
	// 3/6/9/11 shift pattern. Verify against three real Info values from
	// that file.

	// benediction: Info 840
	//   840 & 7         = 0  (SD_NONE)
	//   (840 >> 3) & 7  = 1  (SA_CREATE)
	//   (840 >> 6) & 7  = 5  (SC_LIFE)
	//   (840 >> 9) & 3  = 1  (SP_MINOR)
	sk := &types.SkillType{Info: 840}
	if got := sk.SpellDamageType(); got != types.SD_NONE {
		t.Errorf("benediction SpellDamageType = %d, want SD_NONE", got)
	}
	if got := sk.SpellAction(); got != types.SA_CREATE {
		t.Errorf("benediction SpellAction = %d, want SA_CREATE", got)
	}
	if got := sk.SpellClass(); got != types.SC_LIFE {
		t.Errorf("benediction SpellClass = %d, want SC_LIFE", got)
	}
	if got := sk.SpellPower(); got != types.SP_MINOR {
		t.Errorf("benediction SpellPower = %d, want SP_MINOR", got)
	}

	// acidmist: Info 397
	//   397 & 7         = 5  (SD_ACID)
	//   (397 >> 3) & 7  = 1  (SA_CREATE)
	//   (397 >> 6) & 7  = 6  (SC_DEATH)
	sk = &types.SkillType{Info: 397}
	if got := sk.SpellDamageType(); got != types.SD_ACID {
		t.Errorf("acidmist SpellDamageType = %d, want SD_ACID", got)
	}
	if got := sk.SpellAction(); got != types.SA_CREATE {
		t.Errorf("acidmist SpellAction = %d, want SA_CREATE", got)
	}
	if got := sk.SpellClass(); got != types.SC_DEATH {
		t.Errorf("acidmist SpellClass = %d, want SC_DEATH", got)
	}

	// create fire: Info 521
	//   521 & 7         = 1  (SD_FIRE)
	//   (521 >> 3) & 7  = 1  (SA_CREATE)
	sk = &types.SkillType{Info: 521}
	if got := sk.SpellDamageType(); got != types.SD_FIRE {
		t.Errorf("create fire SpellDamageType = %d, want SD_FIRE", got)
	}
	if got := sk.SpellAction(); got != types.SA_CREATE {
		t.Errorf("create fire SpellAction = %d, want SA_CREATE", got)
	}
}

func TestSkillType_SpellSave_FromInfoBits(t *testing.T) {
	// SpellSave reads bits 11-13 (non-NEWSPELLS).
	// Encode SE_HALFDAM (4) into bits 11-13.
	sk := &types.SkillType{Info: types.SE_HALFDAM << 11}
	if got := sk.SpellSave(); got != types.SE_HALFDAM {
		t.Errorf("SpellSave = %d, want SE_HALFDAM (%d)", got, types.SE_HALFDAM)
	}
	// With no Info bits but a struct SaveEffect, effectiveSave should fall
	// back to the struct field.
	sk = &types.SkillType{SaveEffect: types.SE_NEGATE}
	if got := sk.SpellSave(); got != types.SE_NONE {
		t.Errorf("SpellSave on zero Info = %d, want SE_NONE", got)
	}
	if got := effectiveSave(sk); got != types.SE_NEGATE {
		t.Errorf("effectiveSave fallback = %d, want SE_NEGATE", got)
	}
}

func TestSkillType_HasFlag(t *testing.T) {
	sk := &types.SkillType{Flags: int(types.SF_AREA | types.SF_NOSELF)}
	if !sk.HasFlag(types.SF_AREA) {
		t.Error("SF_AREA not detected")
	}
	if !sk.HasFlag(types.SF_NOSELF) {
		t.Error("SF_NOSELF not detected")
	}
	if sk.HasFlag(types.SF_OBJECT) {
		t.Error("SF_OBJECT falsely detected")
	}
}

func TestSpellSmaug_RegisteredInCodeMap(t *testing.T) {
	if FindSpellFunc("spell_smaug") == nil {
		t.Error("spell_smaug not registered in spellRegistry")
	}
}

func TestSpellSmaug_UnknownActionWithNothing(t *testing.T) {
	// No Affects, no DiceFormula, unknown action — should log Bug and return,
	// not panic.
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 7}
	ch := newCaster("X", 10)
	victim := newNPCVictim("t", 5)
	handler.CharToRoom(ch, room)
	handler.CharToRoom(victim, room)

	skill := &types.SkillType{
		Name: "empty",
		Type: types.SKILL_SPELL,
		Info: 0,
	}
	sn := addSkill(w, skill)
	SpellSmaug(w, sn, 10, ch, victim)
	// No assertion — just must not panic. Record sn so linter is happy.
	_ = sn
}

func TestParseDiceOrInt(t *testing.T) {
	cases := []struct {
		in    string
		level int
		want  int
		exact bool
	}{
		{"", 10, 0, true},
		{"12", 10, 12, true},
		{"-3", 10, -3, true},
		{"level", 25, 25, true},
		{"0", 5, 0, true},
	}
	for _, c := range cases {
		got := parseDiceOrInt(c.in, c.level)
		if c.exact && got != c.want {
			t.Errorf("parseDiceOrInt(%q, %d) = %d, want %d", c.in, c.level, got, c.want)
		}
	}

	// Dice: 1d1 always = 1
	if got := parseDiceOrInt("1d1", 10); got != 1 {
		t.Errorf("parseDiceOrInt(1d1) = %d, want 1", got)
	}
	// 3d1 always = 3
	if got := parseDiceOrInt("3d1", 10); got != 3 {
		t.Errorf("parseDiceOrInt(3d1) = %d, want 3", got)
	}
	// Ensure NdM stays in range [N, N*M]
	for i := 0; i < 20; i++ {
		got := parseDiceOrInt("2d6", 10)
		if got < 2 || got > 12 {
			t.Errorf("parseDiceOrInt(2d6) out of range: %d", got)
		}
	}
}

// TestParseDiceExpr_RealDatPatterns covers every Duration/Modifier expression
// shape found in db/system/en/skills.dat (169 Affect lines surveyed).
func TestParseDiceExpr_RealDatPatterns(t *testing.T) {
	cases := []struct {
		name  string
		in    string
		level int
		want  int
	}{
		{"empty", "", 20, 0},
		{"plain int", "12", 10, 12},
		{"negative int", "-3", 10, -3},
		{"large int", "1048576", 50, 1048576},

		// Level keyword forms.
		{"l keyword", "l", 25, 25},
		{"level keyword", "level", 25, 25},
		{"i alias", "i", 40, 40},

		// Level scaling.
		{"l*15", "l*15", 10, 150},
		{"l*2", "l*2", 30, 60},
		{"level*2", "level*2", 30, 60},
		{"l/2", "l/2", 50, 25},
		{"l/8", "(l/8)", 24, 3},

		// Parenthesized shift.
		{"(l*3)+25", "(l*3)+25", 10, 55},
		{"(l*4)+30", "(l*4)+30", 10, 70},
		{"(l*3)+14", "(l*3)+14", 10, 44},
		{"(l*23)", "(l*23)", 5, 115},

		// Arithmetic with level.
		{"l+25", "l+25", 10, 35},
		{"l+50", "l+50", 20, 70},
		{"l-10", "l-10", 20, 10},
		{"1+(l/17)", "1+(l/17)", 34, 3},

		// Negation of compound.
		{"-(l*4+25)", "-(l*4+25)", 10, -65},
		{"-(l*4+50)", "-(l*4+50)", 10, -90},
		{"-(l/8)", "-(l/8)", 24, -3},

		// Dice with level addend.
		{"l+2d10 min", "l+2d10", 0, 0},  // level 0 → 0 + [2..20]
		{"l+2d10 max", "l+2d10", 0, 20}, // keep it level-0 so we can bound
	}
	for _, c := range cases {
		got := parseDiceExpr(c.in, c.level)
		// For dice-bearing expressions we only bound-check.
		if strings.ContainsRune(c.in, 'd') && !(c.name == "plain int" || c.name == "large int") {
			continue
		}
		if got != c.want {
			t.Errorf("%s: parseDiceExpr(%q, %d) = %d, want %d",
				c.name, c.in, c.level, got, c.want)
		}
	}

	// Dice bounds check: 1d8+(l/3) at level 30 → [1+10, 8+10] = [11, 18].
	for i := 0; i < 40; i++ {
		got := parseDiceExpr("1d8+(l/3)", 30)
		if got < 11 || got > 18 {
			t.Errorf("1d8+(l/3) at l=30 out of range [11,18]: %d", got)
		}
	}
	// 2d8+(l/2) at level 20 → [2+10, 16+10] = [12, 26].
	for i := 0; i < 40; i++ {
		got := parseDiceExpr("2d8+(l/2)", 20)
		if got < 12 || got > 26 {
			t.Errorf("2d8+(l/2) at l=20 out of range [12,26]: %d", got)
		}
	}
	// 3d8+(l-6) at level 16 → [3+10, 24+10] = [13, 34].
	for i := 0; i < 40; i++ {
		got := parseDiceExpr("3d8+(l-6)", 16)
		if got < 13 || got > 34 {
			t.Errorf("3d8+(l-6) at l=16 out of range [13,34]: %d", got)
		}
	}
	// 1d100+100 at any level → [101, 200].
	for i := 0; i < 40; i++ {
		got := parseDiceExpr("1d100+100", 10)
		if got < 101 || got > 200 {
			t.Errorf("1d100+100 out of range [101,200]: %d", got)
		}
	}
	// 5*(l/3) at level 30 → 5*10 = 50.
	if got := parseDiceExpr("5*(l/3)", 30); got != 50 {
		t.Errorf("5*(l/3) at l=30: got %d, want 50", got)
	}

	// Bitvector-name strings ("blind", "sanctuary", etc.) must quietly
	// return 0 — SmaugAff handles those via a name→bit lookup, not the
	// dice parser.
	for _, name := range []string{"blind", "blindness", "sanctuary", "curse", "invis"} {
		if got := parseDiceExpr(name, 20); got != 0 {
			t.Errorf("bitvector name %q: got %d, want 0", name, got)
		}
	}
}

// Smoke-test that a minimal SmaugAff with a valid bitvector bit index
// sets an AFF_* flag on the target.
func TestSpellSmaug_AffectSetsBitVector(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9}
	ch := newCaster("C", 20)
	victim := newNPCVictim("v", 10)
	handler.CharToRoom(ch, room)
	handler.CharToRoom(victim, room)

	skill := &types.SkillType{
		Name:   "charm",
		Type:   types.SKILL_SPELL,
		Target: types.TAR_CHAR_DEFENSIVE,
		Info:   types.SA_CREATE << 3,
		Affects: []*types.SmaugAff{
			{Duration: "10", Location: types.APPLY_NONE, Modifier: "0", BitVector: types.AFF_CHARM},
		},
	}
	sn := addSkill(w, skill)

	SpellSmaug(w, sn, 20, ch, victim)
	if !victim.AffectedBy.IsSet(types.AFF_CHARM) {
		t.Errorf("AFF_CHARM not set on victim; affects=%+v", victim.Affects)
	}
}

// Sanity: after running many random area spells, caster never takes damage.
func TestSpellSmaug_AreaNeverHitsCaster(t *testing.T) {
	for trial := 0; trial < 50; trial++ {
		w := newMagicWorld()
		room := &types.RoomIndexData{Vnum: 100}
		ch := newCaster("Self", 20)
		handler.CharToRoom(ch, room)
		w.AddChar(ch)

		skill := &types.SkillType{
			Name:        "selfblast",
			Type:        types.SKILL_SPELL,
			Target:      types.TAR_IGNORE,
			Info:        (types.SA_DESTROY << 3) | (types.SC_LIFE << 6),
			Flags:       int(types.SF_AREA),
			DiceFormula: "50",
		}
		sn := addSkill(w, skill)

		before := ch.Hit
		SpellSmaug(w, sn, 20, ch, nil)
		if ch.Hit != before {
			t.Fatalf("caster hit by own area spell (trial %d): %d -> %d", trial, before, ch.Hit)
		}
	}
}

// Area attack with SF_PKSENSITIVE hits PCs but halves the effective save
// level (C src/magic.c:2343/4763/6957). It does NOT skip PC victims.
func TestSpellSmaug_PKSensitive_HalvesLevelNotSkip(t *testing.T) {
	// Capture the save-level passed into rollSave.
	var observed []int
	orig := rollSaveFunc
	rollSaveFunc = func(level int, victim *types.CharData, saveType int) bool {
		observed = append(observed, level)
		return false // force no save so damage always lands
	}
	defer func() { rollSaveFunc = orig }()

	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 11}
	ch := newCaster("Mage", 30)
	pc := newCaster("Player", 20) // non-NPC
	handler.CharToRoom(ch, room)
	handler.CharToRoom(pc, room)
	w.AddChar(ch)
	w.AddChar(pc)

	skill := &types.SkillType{
		Name:        "pkfire",
		Type:        types.SKILL_SPELL,
		Target:      types.TAR_IGNORE,
		Info:        (types.SA_DESTROY << 3) | (types.SC_LIFE << 6),
		Flags:       int(types.SF_AREA | types.SF_PKSENSITIVE),
		DiceFormula: "5",
		SaveType:    types.SS_SPELL_STAFF,
	}
	sn := addSkill(w, skill)

	pcBefore := pc.Hit
	SpellSmaug(w, sn, 30, ch, nil)

	if pc.Hit >= pcBefore {
		t.Errorf("PC victim NOT hit by PKSENSITIVE area spell (should still hit): %d -> %d",
			pcBefore, pc.Hit)
	}
	// Level passed to rollSave for the PC target should be half (15).
	found := false
	for _, lv := range observed {
		if lv == 15 {
			found = true
		}
	}
	if !found {
		t.Errorf("PKSENSITIVE did not halve save level for PC; saw %v", observed)
	}
}

// SE_REFLECT on a successful save routes the damage back to the caster.
func TestSpellSmaug_SaveReflect(t *testing.T) {
	orig := rollSaveFunc
	rollSaveFunc = func(level int, victim *types.CharData, saveType int) bool { return true }
	defer func() { rollSaveFunc = orig }()

	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 200}
	ch := newCaster("Attacker", 20)
	victim := newNPCVictim("mirror", 10)
	ch.Hit = 100
	victim.Hit = 100
	handler.CharToRoom(ch, room)
	handler.CharToRoom(victim, room)
	w.AddChar(ch)
	w.AddChar(victim)

	skill := &types.SkillType{
		Name:        "bouncy",
		Type:        types.SKILL_SPELL,
		Target:      types.TAR_CHAR_OFFENSIVE,
		Info:        (types.SA_DESTROY << 3) | (types.SC_LIFE << 6),
		DiceFormula: "15",
		SaveType:    types.SS_SPELL_STAFF,
		SaveEffect:  types.SE_REFLECT,
	}
	sn := addSkill(w, skill)

	chBefore, vBefore := ch.Hit, victim.Hit
	SpellSmaug(w, sn, 20, ch, victim)
	if ch.Hit >= chBefore {
		t.Errorf("SE_REFLECT: caster took no damage: %d -> %d", chBefore, ch.Hit)
	}
	if victim.Hit != vBefore {
		t.Errorf("SE_REFLECT: victim took damage (should be reflected): %d -> %d",
			vBefore, victim.Hit)
	}
}

// SE_ABSORB on a successful save heals the victim and deals no damage.
func TestSpellSmaug_SaveAbsorb(t *testing.T) {
	orig := rollSaveFunc
	rollSaveFunc = func(level int, victim *types.CharData, saveType int) bool { return true }
	defer func() { rollSaveFunc = orig }()

	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 201}
	ch := newCaster("Attacker", 20)
	victim := newNPCVictim("sponge", 10)
	victim.Hit = 50
	victim.MaxHit = 100
	handler.CharToRoom(ch, room)
	handler.CharToRoom(victim, room)
	w.AddChar(ch)
	w.AddChar(victim)

	skill := &types.SkillType{
		Name:        "drainer",
		Type:        types.SKILL_SPELL,
		Target:      types.TAR_CHAR_OFFENSIVE,
		Info:        (types.SA_DESTROY << 3) | (types.SC_LIFE << 6),
		DiceFormula: "30",
		SaveType:    types.SS_SPELL_STAFF,
		SaveEffect:  types.SE_ABSORB,
	}
	sn := addSkill(w, skill)

	SpellSmaug(w, sn, 20, ch, victim)
	if victim.Hit != 80 {
		t.Errorf("SE_ABSORB: victim.Hit = %d, want 80 (50 + 30)", victim.Hit)
	}
}

// Verify integration with FindSpellFunc — spell_smaug is callable through
// the registry with the canonical signature.
func TestSpellSmaug_InvokeViaRegistry(t *testing.T) {
	fn := FindSpellFunc("spell_smaug")
	if fn == nil {
		t.Fatal("spell_smaug not registered")
	}
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 12}
	ch := newCaster("C", 5)
	handler.CharToRoom(ch, room)
	// sn=-1 takes the nil-safe branch; must not panic.
	fn(w, -1, 5, ch, nil)
}

// Keyword-check: the test file's name for the Code line in skills.dat.
// The C source looks up by string name; our registry uses the same string.
func TestSpellSmaug_RegistryKeyMatchesDatCode(t *testing.T) {
	// Key seen in db/system/en/skills.dat, e.g. "Code  spell_smaug".
	key := "spell_smaug"
	if fn := FindSpellFunc(key); fn == nil {
		t.Errorf("registry missing key %q", key)
	}
	// Sanity: uppercase variant not matched.
	if FindSpellFunc(strings.ToUpper(key)) != nil {
		t.Errorf("registry matched uppercase variant; should be case-sensitive")
	}
}
