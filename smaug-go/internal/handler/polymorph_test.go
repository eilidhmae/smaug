package handler

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// newMorphTestPC returns a minimal PC with PCData populated and baseline
// stats at zero. Tests layer morph templates on top.
func newMorphTestPC() *types.CharData {
	return &types.CharData{
		Name:  "tester",
		Level: 20,
		Class: 0,
		Race:  0,
		Sex:   0,
		Hit:   100, MaxHit: 100,
		Mana: 100, MaxMana: 100,
		Move: 100, MaxMove: 100,
		PCData: &types.PCData{PagerLen: 24},
	}
}

// TestMakeCharMorph_CopiesStaticFields pins the static-copy contract: the
// "no-dice" fields (AC, Mod stats, Savings, Timer, RIS bitmasks) come
// from the morph template; the "dice" fields (Hitroll, Damroll, Hit,
// Mana, Move, Blood) start zero and are populated by DoMorph.
func TestMakeCharMorph_CopiesStaticFields(t *testing.T) {
	m := &types.MorphData{
		AC: -7, Str: 2, Int: -1, Timer: 5,
		Immune: 3, Resistant: 7,
	}
	cm := MakeCharMorph(m)
	if cm.Morph != m {
		t.Error("Morph pointer not set")
	}
	if cm.AC != -7 {
		t.Errorf("AC = %d, want -7", cm.AC)
	}
	if cm.Str != 2 {
		t.Errorf("Str = %d, want 2", cm.Str)
	}
	if cm.Timer != 5 {
		t.Errorf("Timer = %d, want 5", cm.Timer)
	}
	if cm.Immune != 3 {
		t.Errorf("Immune = %d, want 3", cm.Immune)
	}
	// Dice fields are zero.
	if cm.Hit != 0 || cm.Damroll != 0 || cm.Hitroll != 0 ||
		cm.Mana != 0 || cm.Move != 0 || cm.Blood != 0 {
		t.Errorf("dice fields nonzero: %+v", cm)
	}
}

// TestDoMorph_AppliesAllDeltas pins every field DoMorph touches.
func TestDoMorph_AppliesAllDeltas(t *testing.T) {
	ch := newMorphTestPC()
	ch.Armor = 100
	ch.ModStr = 0
	ch.Hitroll = 5
	ch.Damroll = 3
	m := &types.MorphData{
		Name:              "testform",
		Level:             20,
		AC:                -10,
		Str:               3,
		Int:               1,
		Wis:               -1,
		Dex:               2,
		Cha:               0,
		Lck:               1,
		SavingBreath:      -5,
		SavingParaPetri:   -3,
		SavingPoisonDeath: -2,
		SavingSpellStaff:  -1,
		SavingWand:        -4,
		Hit:               "10",
		Hitroll:           "2",
		Damroll:           "3",
		Mana:              "5",
		Move:              "8",
		Immune:            0x01,
		Resistant:         0x02,
		Suscept:           0x04,
		NoImmune:          0,
		NoResistant:       0,
		NoSuscept:         0,
	}
	DoMorph(ch, m)
	if ch.Morph == nil {
		t.Fatal("ch.Morph not attached")
	}
	if ch.Armor != 90 {
		t.Errorf("Armor = %d, want 90", ch.Armor)
	}
	if ch.ModStr != 3 || ch.ModInt != 1 || ch.ModWis != -1 ||
		ch.ModDex != 2 || ch.ModCha != 0 || ch.ModLck != 1 {
		t.Errorf("Mod stats wrong: str=%d int=%d wis=%d dex=%d cha=%d lck=%d",
			ch.ModStr, ch.ModInt, ch.ModWis, ch.ModDex, ch.ModCha, ch.ModLck)
	}
	if ch.SavingBreath != -5 {
		t.Errorf("SavingBreath = %d, want -5", ch.SavingBreath)
	}
	if ch.SavingWand != -4 {
		t.Errorf("SavingWand = %d, want -4", ch.SavingWand)
	}
	if ch.Hitroll != 7 {
		t.Errorf("Hitroll = %d, want 7 (5+2)", ch.Hitroll)
	}
	if ch.Damroll != 6 {
		t.Errorf("Damroll = %d, want 6 (3+3)", ch.Damroll)
	}
	if ch.Hit != 110 {
		t.Errorf("Hit = %d, want 110", ch.Hit)
	}
	if ch.Mana != 105 {
		t.Errorf("Mana = %d, want 105", ch.Mana)
	}
	if ch.Move != 108 {
		t.Errorf("Move = %d, want 108", ch.Move)
	}
	if ch.Immune != 1 || ch.Resistant != 2 || ch.Susceptible != 4 {
		t.Errorf("RIS = %d/%d/%d, want 1/2/4", ch.Immune, ch.Resistant, ch.Susceptible)
	}
	if m.Used != 1 {
		t.Errorf("morph.Used = %d, want 1", m.Used)
	}
}

// TestDoUnmorph_RoundTrip: apply then unmorph must restore every field
// to its pre-morph value. This is the high-leverage invariant — a drift
// in any of the 20+ fields compounds across morph cycles.
func TestDoUnmorph_RoundTrip(t *testing.T) {
	ch := newMorphTestPC()
	ch.Armor = 42
	ch.ModStr = 5
	ch.Hitroll = 3
	ch.Damroll = 2
	// Pre-morph RIS state uses bits NOT touched by the morph's
	// NoImmune/NoResistant/NoSuscept masks — otherwise the round-trip
	// is lossy (C semantics: DoUnmorph restores morph.Immune but does
	// NOT re-add bits stripped by morph.NoImmune; see C do_unmorph
	// at polymorph.c:1571-1574). The asymmetry is C-verbatim.
	ch.Immune = 0x100
	ch.Resistant = 0x200
	ch.Susceptible = 0x400
	ch.AffectedBy = types.BitVector{0x1111, 0x2222, 0, 0}

	pre := *ch
	preBV := ch.AffectedBy

	m := &types.MorphData{
		Level:             10,
		AC:                -10,
		Str:               3,
		Int:               -1,
		Wis:               2,
		Dex:               -2,
		Cha:               1,
		Lck:               -1,
		Con:               0, // CharData has no Con mod; not mutated
		SavingBreath:      -5,
		SavingParaPetri:   -3,
		SavingPoisonDeath: -2,
		SavingSpellStaff:  -1,
		SavingWand:        -4,
		Hit:               "10",
		Hitroll:           "2",
		Damroll:           "3",
		Mana:              "5",
		Move:              "8",
		Immune:            0x01,
		Resistant:         0x02,
		Suscept:           0x04,
		// NoImmune/NoResistant/NoSuscept intentionally zero — no strip
		// to lose on round-trip. A separate test pins the C asymmetry.
	}
	DoMorph(ch, m)
	DoUnmorph(ch)
	if ch.Morph != nil {
		t.Error("Morph pointer should be nil after DoUnmorph")
	}
	if ch.Armor != pre.Armor {
		t.Errorf("Armor: got %d, want %d", ch.Armor, pre.Armor)
	}
	if ch.ModStr != pre.ModStr {
		t.Errorf("ModStr: got %d, want %d", ch.ModStr, pre.ModStr)
	}
	if ch.Hitroll != pre.Hitroll {
		t.Errorf("Hitroll: got %d, want %d", ch.Hitroll, pre.Hitroll)
	}
	if ch.Damroll != pre.Damroll {
		t.Errorf("Damroll: got %d, want %d", ch.Damroll, pre.Damroll)
	}
	if ch.Hit != pre.Hit {
		t.Errorf("Hit: got %d, want %d", ch.Hit, pre.Hit)
	}
	if ch.Mana != pre.Mana {
		t.Errorf("Mana: got %d, want %d", ch.Mana, pre.Mana)
	}
	if ch.Move != pre.Move {
		t.Errorf("Move: got %d, want %d", ch.Move, pre.Move)
	}
	if ch.SavingBreath != pre.SavingBreath {
		t.Errorf("SavingBreath: got %d, want %d", ch.SavingBreath, pre.SavingBreath)
	}
	// Round-trip restores RIS masks.
	if ch.Immune != pre.Immune {
		t.Errorf("Immune: got %d, want %d", ch.Immune, pre.Immune)
	}
	if ch.Resistant != pre.Resistant {
		t.Errorf("Resistant: got %d, want %d", ch.Resistant, pre.Resistant)
	}
	if ch.Susceptible != pre.Susceptible {
		t.Errorf("Susceptible: got %d, want %d", ch.Susceptible, pre.Susceptible)
	}
	if !ch.AffectedBy.Equal(preBV) {
		t.Errorf("AffectedBy: got %v, want %v", ch.AffectedBy, preBV)
	}
}

// TestDoMorph_NoImmuneStrip_CBugPreserved: DoMorph strips bits via
// NoImmune, but DoUnmorph does NOT re-add them. Matches C
// do_unmorph at polymorph.c:1571-1574 which only REMOVE_BITs back
// morph.Immune and leaves no_immune strippings permanent. Documenting
// the asymmetry so future refactors can spot intentional divergence.
func TestDoMorph_NoImmuneStrip_CBugPreserved(t *testing.T) {
	ch := newMorphTestPC()
	ch.Immune = 0x80
	m := &types.MorphData{Level: 10, NoImmune: 0x80}
	DoMorph(ch, m)
	if ch.Immune != 0 {
		t.Errorf("after DoMorph: Immune = %d, want 0 (stripped)", ch.Immune)
	}
	DoUnmorph(ch)
	if ch.Immune != 0 {
		t.Errorf("after DoUnmorph: Immune = %d, want 0 (C does not re-add stripped bits)", ch.Immune)
	}
}

// TestDoMorph_TwoMorphsSwap: applying a second morph should first
// unmorph the first (C :1480-1481), not stack.
func TestDoMorph_TwoMorphsSwap(t *testing.T) {
	ch := newMorphTestPC()
	a := &types.MorphData{Level: 10, AC: -5, Str: 2}
	b := &types.MorphData{Level: 10, AC: -10, Str: 4}
	DoMorph(ch, a)
	DoMorph(ch, b)
	if ch.Morph == nil || ch.Morph.Morph != b {
		t.Fatalf("ch.Morph should point to b")
	}
	// Net delta should be b's, not a+b.
	if ch.Armor != -10 {
		t.Errorf("Armor = %d, want -10 (only b applied)", ch.Armor)
	}
	if ch.ModStr != 4 {
		t.Errorf("ModStr = %d, want 4 (only b applied)", ch.ModStr)
	}
}

// TestDoMorph_VampireUsesBlood: a PC with RACE_VAMPIRE routes to the
// bloodthirst condition path instead of Mana.
func TestDoMorph_VampireUsesBlood(t *testing.T) {
	ch := newMorphTestPC()
	ch.Race = types.RACE_VAMPIRE
	ch.PCData.Condition[types.COND_BLOODTHIRST] = 10
	ch.Mana = 50
	m := &types.MorphData{Level: 10, Blood: "5", Mana: "99"}
	DoMorph(ch, m)
	if ch.PCData.Condition[types.COND_BLOODTHIRST] != 15 {
		t.Errorf("bloodthirst = %d, want 15", ch.PCData.Condition[types.COND_BLOODTHIRST])
	}
	if ch.Mana != 50 {
		t.Errorf("Mana = %d, want 50 (vampire takes blood, skips mana)", ch.Mana)
	}
	DoUnmorph(ch)
	if ch.PCData.Condition[types.COND_BLOODTHIRST] != 10 {
		t.Errorf("bloodthirst after unmorph = %d, want 10",
			ch.PCData.Condition[types.COND_BLOODTHIRST])
	}
}

// TestDoMorph_HitClamp: Hit cannot exceed MaxMorphVital after adding
// the morph delta (C :1501-1503).
func TestDoMorph_HitClamp(t *testing.T) {
	ch := newMorphTestPC()
	ch.Hit = 32699
	m := &types.MorphData{Level: 10, Hit: "100"}
	DoMorph(ch, m)
	if ch.Hit != MaxMorphVital {
		t.Errorf("Hit = %d, want %d (clamped)", ch.Hit, MaxMorphVital)
	}
	if ch.Morph.Hit != 1 {
		t.Errorf("stored delta = %d, want 1 (clamp residual)", ch.Morph.Hit)
	}
	DoUnmorph(ch)
	if ch.Hit != 32699 {
		t.Errorf("Hit post-unmorph = %d, want 32699", ch.Hit)
	}
}

// TestDoMorphChar_InsufficientHpBails tests the C-verbatim message.
func TestDoMorphChar_InsufficientHpBails(t *testing.T) {
	ch := newMorphTestPC()
	ch.Hit = 5
	ch.Desc = newCaptureDesc()
	m := &types.MorphData{Level: 10, HpUsed: 50}
	w := world.New("")
	ok := DoMorphChar(w, ch, m)
	if ok {
		t.Error("DoMorphChar should return false on insufficient HP")
	}
	if ch.Morph != nil {
		t.Error("ch.Morph should not be attached on failure")
	}
	out := capturedOutput(ch.Desc)
	if !contains(out, "begin to transform") || !contains(out, "something goes wrong") {
		t.Errorf("expected 'begin to transform...something goes wrong' message, got %q", out)
	}
}

// TestDoMorphChar_ConsumesHpOnSuccess.
func TestDoMorphChar_ConsumesHpOnSuccess(t *testing.T) {
	ch := newMorphTestPC()
	ch.Hit = 100
	ch.Desc = newCaptureDesc()
	m := &types.MorphData{Level: 10, HpUsed: 20}
	w := world.New("")
	ok := DoMorphChar(w, ch, m)
	if !ok {
		t.Fatal("DoMorphChar should succeed")
	}
	if ch.Hit != 80 {
		t.Errorf("Hit = %d, want 80 (100-20)", ch.Hit)
	}
	if ch.Morph == nil {
		t.Error("ch.Morph should be attached on success")
	}
}

// TestDoMorphChar_AlreadyMorphedBails: can_morph-style gate. C :1268-1269
// sets canmorph=FALSE if ch.morph is non-nil.
func TestDoMorphChar_AlreadyMorphedBails(t *testing.T) {
	ch := newMorphTestPC()
	ch.Desc = newCaptureDesc()
	pre := &types.MorphData{Level: 10}
	DoMorph(ch, pre)
	m := &types.MorphData{Level: 10}
	w := world.New("")
	ok := DoMorphChar(w, ch, m)
	if ok {
		t.Error("DoMorphChar on already-morphed char should bail")
	}
	// The prior morph is still attached.
	if ch.Morph == nil || ch.Morph.Morph != pre {
		t.Error("prior morph should survive the failed second attempt")
	}
}

// TestDoUnmorphChar_EmitsUnmorphMessage.
func TestDoUnmorphChar_EmitsUnmorphMessage(t *testing.T) {
	ch := newMorphTestPC()
	ch.Desc = newCaptureDesc()
	m := &types.MorphData{
		Level:        10,
		MorphSelf:    "You shift.",
		UnmorphSelf:  "You return.",
		MorphOther:   "$n shifts.",
		UnmorphOther: "$n returns.",
	}
	DoMorph(ch, m)
	// Clear the buffer before the unmorph message.
	resetCaptureDesc(ch.Desc)
	DoUnmorphChar(ch)
	if ch.Morph != nil {
		t.Error("Morph should be nil after DoUnmorphChar")
	}
	out := capturedOutput(ch.Desc)
	if !contains(out, "You return.") {
		t.Errorf("unmorph message not emitted: %q", out)
	}
}

// TestUnmorphAll_ClearsAllUsers.
func TestUnmorphAll_ClearsAllUsers(t *testing.T) {
	w := world.New("")
	m := &types.MorphData{Level: 10, AC: -5}
	other := &types.MorphData{Level: 10, AC: -3}
	a := newMorphTestPC()
	a.Name = "alice"
	b := newMorphTestPC()
	b.Name = "bob"
	c := newMorphTestPC()
	c.Name = "carol"
	w.Characters = []*types.CharData{a, b, c}
	DoMorph(a, m)
	DoMorph(b, m)
	DoMorph(c, other)
	UnmorphAll(w, m)
	if a.Morph != nil {
		t.Error("alice should be unmorphed")
	}
	if b.Morph != nil {
		t.Error("bob should be unmorphed")
	}
	if c.Morph == nil {
		t.Error("carol wore a different morph; should remain morphed")
	}
}

// TestCanMorph_ImmortalBypassesAllGates.
func TestCanMorph_ImmortalBypassesAllGates(t *testing.T) {
	ch := newMorphTestPC()
	ch.Level = types.LEVEL_IMMORTAL
	ch.Trust = types.LEVEL_IMMORTAL
	m := &types.MorphData{Level: 999, Class: 0x01, Sex: 9}
	if !CanMorph(ch, m, false) {
		t.Error("immortal should bypass all CanMorph gates")
	}
}

// TestCanMorph_LevelGate.
func TestCanMorph_LevelGate(t *testing.T) {
	ch := newMorphTestPC()
	ch.Level = 10
	m := &types.MorphData{Level: 20}
	if CanMorph(ch, m, false) {
		t.Error("CanMorph should deny below-level PC")
	}
	m.Level = 5
	if !CanMorph(ch, m, false) {
		t.Error("CanMorph should allow at-or-above-level PC")
	}
}

// TestCanMorph_NoCastGateRespectsIsCast.
func TestCanMorph_NoCastGateRespectsIsCast(t *testing.T) {
	ch := newMorphTestPC()
	m := &types.MorphData{NoCast: true}
	if CanMorph(ch, m, true) {
		t.Error("no-cast morph should deny spell-cast entry")
	}
	if !CanMorph(ch, m, false) {
		t.Error("no-cast morph should allow immortal/mpmorph entry")
	}
}

// TestCanMorph_RaceMaskInverted documents the C bug-feature: morph.Race
// is a mask of DISALLOWED races (inverted vs Class).
func TestCanMorph_RaceMaskInverted(t *testing.T) {
	ch := newMorphTestPC()
	ch.Race = 3
	m := &types.MorphData{Race: 1 << 3} // disallow race 3
	if CanMorph(ch, m, false) {
		t.Error("Race bit set means DISALLOWED — CanMorph should deny")
	}
	m.Race = 1 << 2 // disallow race 2 only
	if !CanMorph(ch, m, false) {
		t.Error("Different-race bit set should allow race 3 to morph")
	}
}

// --- Capture helpers: tiny DescriptorData stand-in for collecting act output ---

func newCaptureDesc() *types.DescriptorData {
	return &types.DescriptorData{}
}

func resetCaptureDesc(d *types.DescriptorData) {
	d.ResetBufferedOutput()
}

func capturedOutput(d *types.DescriptorData) string {
	return d.BufferedOutput()
}

func contains(s, sub string) bool {
	return len(sub) == 0 || indexOf(s, sub) != -1
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
