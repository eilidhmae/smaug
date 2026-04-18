package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// TestLearnFromSuccess_AwardsXP_NormalBranch pins the normal-gain XP
// formula (C src/skills.c:1656 — 20 * skLvl with class multipliers).
// With CLASS_WARRIOR (no multiplier) and skLvl=10, one successful gain
// should add 200 XP and emit the gain message.
func TestLearnFromSuccess_AwardsXP_NormalBranch(t *testing.T) {
	const gsn = 510
	restore := installLearnTestSkill(t, gsn, 5, 95, "xpnorm")
	defer restore()
	restoreRNG := withStubNumberPercent(99) // forces gain=2
	defer restoreRNG()

	ch, client := makeTestChar("Warrior")
	defer client.Close()
	ch.Class = types.CLASS_WARRIOR
	ch.Level = 10
	ch.PCData.Learned[gsn] = 10
	ch.Exp = 1000

	learnFromSuccess(ch, gsn)

	if ch.Exp-1000 != 200 {
		t.Errorf("Exp delta = %d, want 200 (20 * skLvl 10)", ch.Exp-1000)
	}
	got := readOutput(ch, client)
	if !strings.Contains(got, "You gain 200 experience points") {
		t.Errorf("expected gain message; got %q", got)
	}
}

// TestLearnFromSuccess_ClassMultiplier_Mage pins the CLASS_MAGE ×6
// multiplier on normal-gain XP (C src/skills.c:1657).
func TestLearnFromSuccess_ClassMultiplier_Mage(t *testing.T) {
	const gsn = 511
	restore := installLearnTestSkill(t, gsn, 5, 95, "xpmage")
	defer restore()
	restoreRNG := withStubNumberPercent(99)
	defer restoreRNG()

	ch, client := makeTestChar("Mage")
	defer client.Close()
	ch.Class = types.CLASS_MAGE
	ch.Level = 10
	ch.PCData.Learned[gsn] = 10
	ch.Exp = 0

	learnFromSuccess(ch, gsn)

	if ch.Exp != 1200 {
		t.Errorf("mage normal-gain XP = %d, want 1200 (20 * 10 * 6)", ch.Exp)
	}
}

// TestLearnFromSuccess_ClassMultiplier_Cleric pins the CLASS_CLERIC ×3
// multiplier on normal-gain XP (C src/skills.c:1658).
func TestLearnFromSuccess_ClassMultiplier_Cleric(t *testing.T) {
	const gsn = 512
	restore := installLearnTestSkill(t, gsn, 5, 95, "xpcleric")
	defer restore()
	restoreRNG := withStubNumberPercent(99)
	defer restoreRNG()

	ch, client := makeTestChar("Cleric")
	defer client.Close()
	ch.Class = types.CLASS_CLERIC
	ch.Level = 10
	ch.PCData.Learned[gsn] = 10
	ch.Exp = 0

	learnFromSuccess(ch, gsn)

	if ch.Exp != 600 {
		t.Errorf("cleric normal-gain XP = %d, want 600 (20 * 10 * 3)", ch.Exp)
	}
}

// TestLearnFromSuccess_FightingSuppressesMessage verifies that XP is
// awarded but the "You gain X experience points" message is withheld
// when the player is actively fighting — matches C src/skills.c:1661.
func TestLearnFromSuccess_FightingSuppressesMessage(t *testing.T) {
	const gsn = 513
	restore := installLearnTestSkill(t, gsn, 5, 95, "xpfight")
	defer restore()
	restoreRNG := withStubNumberPercent(99)
	defer restoreRNG()

	ch, client := makeTestChar("Combat")
	defer client.Close()
	ch.Class = types.CLASS_WARRIOR
	ch.Level = 10
	ch.PCData.Learned[gsn] = 10
	// Simulate being in a fight — any non-nil attacker record is enough.
	opponent := &types.CharData{Name: "Foe"}
	ch.Fighting = &types.FightData{Who: opponent}

	learnFromSuccess(ch, gsn)

	if ch.Exp == 0 {
		t.Error("XP should still be awarded while Fighting")
	}
	if ch.Exp != 200 {
		t.Errorf("Fighting XP = %d, want 200", ch.Exp)
	}
	got := readOutput(ch, client)
	if strings.Contains(got, "experience points from your success") {
		t.Errorf("XP message should NOT be emitted while Fighting; got %q", got)
	}
}

// TestLearnFromSuccess_HideSneakSuppressMessage verifies the
// silent-skill branch (C src/skills.c:1661 — gsn == gsn_hide / gsn_sneak).
// XP must still be awarded, only the message is suppressed.
func TestLearnFromSuccess_HideSneakSuppressMessage(t *testing.T) {
	const gsn = 514
	restore := installLearnTestSkill(t, gsn, 5, 95, "hide")
	defer restore()
	restoreRNG := withStubNumberPercent(99)
	defer restoreRNG()

	// Force the package-level gsnHide to our test slot.
	prev := gsnHide
	gsnHide = gsn
	defer func() { gsnHide = prev }()

	ch, client := makeTestChar("Shadow")
	defer client.Close()
	ch.Class = types.CLASS_WARRIOR
	ch.Level = 10
	ch.PCData.Learned[gsn] = 10

	learnFromSuccess(ch, gsn)

	if ch.Exp != 200 {
		t.Errorf("Hide XP = %d, want 200", ch.Exp)
	}
	got := readOutput(ch, client)
	if strings.Contains(got, "experience points from your success") {
		t.Errorf("XP message should NOT be emitted for silent skills; got %q", got)
	}
}

// TestLearnFromSuccess_AdeptCap_Message_And_LargerXP pins the adept-cap
// branch (C src/skills.c:1644-1652): "You are now an adept of X!"
// message in AT_WHITE + 1000*skLvl XP with ×5/×2 class multipliers.
func TestLearnFromSuccess_AdeptCap_Message_And_LargerXP(t *testing.T) {
	const gsn = 515
	restore := installLearnTestSkill(t, gsn, 5, 95, "adept")
	defer restore()
	restoreRNG := withStubNumberPercent(99) // roll pushes gain=2; 94+2 clamps to 95
	defer restoreRNG()

	ch, client := makeTestChar("Master")
	defer client.Close()
	ch.Class = types.CLASS_WARRIOR
	ch.Level = 10
	ch.PCData.Learned[gsn] = 94 // one gain away from adept=95

	learnFromSuccess(ch, gsn)

	if ch.PCData.Learned[gsn] != 95 {
		t.Fatalf("Learned = %d, want 95 (clamped to adept)", ch.PCData.Learned[gsn])
	}
	// Warrior: 1000 * 10 = 10000 (no multiplier).
	if ch.Exp != 10000 {
		t.Errorf("adept-cap XP = %d, want 10000 (1000 * skLvl 10)", ch.Exp)
	}
	got := readOutput(ch, client)
	if !strings.Contains(got, "You are now an adept of adept") {
		t.Errorf("expected adept-cap message; got %q", got)
	}
	if !strings.Contains(got, "10000 bonus experience") {
		t.Errorf("expected bonus XP in message; got %q", got)
	}
	// The &W color token must appear (rendered via ProcessColors at the
	// net layer; here we only assert the raw token made it through).
	if !strings.Contains(got, "&W") {
		t.Errorf("adept-cap message should carry &W color token; got %q", got)
	}
}

// TestLearnFromSuccess_AdeptCap_Mage_Multiplier pins ×5 mage multiplier
// on the adept-cap branch (C src/skills.c:1647).
func TestLearnFromSuccess_AdeptCap_Mage_Multiplier(t *testing.T) {
	const gsn = 516
	restore := installLearnTestSkill(t, gsn, 5, 95, "mageadept")
	defer restore()
	restoreRNG := withStubNumberPercent(99)
	defer restoreRNG()

	ch, client := makeTestChar("Archmage")
	defer client.Close()
	ch.Class = types.CLASS_MAGE
	ch.Level = 10
	ch.PCData.Learned[gsn] = 94

	learnFromSuccess(ch, gsn)

	if ch.Exp != 50000 {
		t.Errorf("mage adept-cap XP = %d, want 50000 (1000 * 10 * 5)", ch.Exp)
	}
}

// TestLearnFromSuccess_AdeptCap_Cleric_Multiplier pins ×2 cleric
// multiplier on the adept-cap branch (C src/skills.c:1649).
func TestLearnFromSuccess_AdeptCap_Cleric_Multiplier(t *testing.T) {
	const gsn = 517
	restore := installLearnTestSkill(t, gsn, 5, 95, "clericadept")
	defer restore()
	restoreRNG := withStubNumberPercent(99)
	defer restoreRNG()

	ch, client := makeTestChar("Highpriest")
	defer client.Close()
	ch.Class = types.CLASS_CLERIC
	ch.Level = 10
	ch.PCData.Learned[gsn] = 94

	learnFromSuccess(ch, gsn)

	if ch.Exp != 20000 {
		t.Errorf("cleric adept-cap XP = %d, want 20000 (1000 * 10 * 2)", ch.Exp)
	}
}

// TestLearnFromSuccess_AdeptCap_MutuallyExclusiveWithNormalBranch pins
// that the normal-gain "You gain X experience points" message is NOT
// emitted when the adept-cap branch fires (C src/skills.c:1642-1654
// if/else). Two messages would be a double-count bug.
func TestLearnFromSuccess_AdeptCap_MutuallyExclusiveWithNormalBranch(t *testing.T) {
	const gsn = 518
	restore := installLearnTestSkill(t, gsn, 5, 95, "exclusive")
	defer restore()
	restoreRNG := withStubNumberPercent(99)
	defer restoreRNG()

	ch, client := makeTestChar("Capper")
	defer client.Close()
	ch.Class = types.CLASS_WARRIOR
	ch.Level = 10
	ch.PCData.Learned[gsn] = 94

	learnFromSuccess(ch, gsn)

	got := readOutput(ch, client)
	if strings.Contains(got, "experience points from your success") {
		t.Errorf("normal-gain message should NOT fire on adept cap; got %q", got)
	}
	if !strings.Contains(got, "You are now an adept") {
		t.Errorf("adept-cap message SHOULD fire; got %q", got)
	}
	// Only one xpGain should be recorded, not both summed.
	if ch.Exp != 10000 {
		t.Errorf("Exp = %d, want exactly 10000 (only adept branch counted)", ch.Exp)
	}
}

// TestLearnFromSuccess_SkillLevelZeroFallsBackToCharLevel pins the
// C src/skills.c:1641 behavior — if skill.SkillLevel[class] is 0 the
// formula uses ch.Level instead. Without this, low-level skills the
// player is using at a high level would award zero XP.
func TestLearnFromSuccess_SkillLevelZeroFallsBackToCharLevel(t *testing.T) {
	const gsn = 519
	restore := installLearnTestSkill(t, gsn, 5, 95, "falllevel")
	defer restore()
	// installLearnTestSkill leaves SkillLevel[class] = 0 by default.
	restoreRNG := withStubNumberPercent(99)
	defer restoreRNG()

	ch, client := makeTestChar("Tester")
	defer client.Close()
	ch.Class = types.CLASS_WARRIOR
	ch.Level = 15
	ch.PCData.Learned[gsn] = 10

	learnFromSuccess(ch, gsn)

	// skLvl falls back to ch.Level=15 → 20 * 15 = 300.
	if ch.Exp != 300 {
		t.Errorf("fallback XP = %d, want 300 (20 * ch.Level 15)", ch.Exp)
	}
	_ = client
}
