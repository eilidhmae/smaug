package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func setupSkillWorld() *world.World {
	w := setupWizWorld()
	// Add some skills
	w.Skills = make([]*types.SkillType, 20)
	w.Skills[1] = &types.SkillType{Name: "backstab", Type: types.SKILL_SKILL}
	w.Skills[2] = &types.SkillType{Name: "kick", Type: types.SKILL_SKILL}
	w.Skills[3] = &types.SkillType{Name: "magic missile", Type: types.SKILL_SPELL}
	w.Skills[4] = &types.SkillType{Name: "fireball", Type: types.SKILL_SPELL}
	w.Skills[5] = &types.SkillType{Name: "cure light", Type: types.SKILL_SPELL}
	return w
}

// --- DoSkills ---

func TestDoSkills_NoSkills(t *testing.T) {
	_ = setupSkillWorld()
	ch, client := makeTestChar("Warrior")
	defer client.Close()

	DoSkills(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "no skills") && !strings.Contains(out, "No skills") {
		t.Errorf("expected no skills message, got: %q", out)
	}
}

func TestDoSkills_WithSkills(t *testing.T) {
	_ = setupSkillWorld()
	ch, client := makeTestChar("Warrior")
	defer client.Close()
	ch.PCData.Learned[1] = 75 // backstab
	ch.PCData.Learned[2] = 50 // kick

	DoSkills(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "backstab") {
		t.Errorf("expected 'backstab' in output, got: %q", out)
	}
	if !strings.Contains(out, "kick") {
		t.Errorf("expected 'kick' in output, got: %q", out)
	}
	// Should NOT contain spells
	if strings.Contains(out, "magic missile") {
		t.Error("skills output should not contain spells")
	}
}

// --- DoSpells ---

func TestDoSpells_NoSpells(t *testing.T) {
	_ = setupSkillWorld()
	ch, client := makeTestChar("Warrior")
	defer client.Close()

	DoSpells(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "no spells") && !strings.Contains(out, "No spells") {
		t.Errorf("expected no spells message, got: %q", out)
	}
}

func TestDoSpells_WithSpells(t *testing.T) {
	_ = setupSkillWorld()
	ch, client := makeTestChar("Mage")
	defer client.Close()
	ch.PCData.Learned[3] = 90 // magic missile
	ch.PCData.Learned[4] = 60 // fireball

	DoSpells(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "magic missile") {
		t.Errorf("expected 'magic missile' in output, got: %q", out)
	}
	if !strings.Contains(out, "fireball") {
		t.Errorf("expected 'fireball' in output, got: %q", out)
	}
	// Should NOT contain skills
	if strings.Contains(out, "backstab") {
		t.Error("spells output should not contain skills")
	}
}

// --- DoPractice ---

func TestDoPractice_NoArg_ListsSkills(t *testing.T) {
	_ = setupSkillWorld()
	ch, client := makeTestChar("Warrior")
	defer client.Close()
	ch.PCData.Learned[1] = 50
	ch.PCData.Learned[3] = 30

	DoPractice(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "backstab") {
		t.Errorf("expected 'backstab' in practice list, got: %q", out)
	}
}

func TestDoPractice_NPC(t *testing.T) {
	_ = setupSkillWorld()
	ch, client := makeTestChar("Mob")
	defer client.Close()
	ch.Act.Set(types.ACT_IS_NPC)
	ch.PCData = nil

	DoPractice(ch, "")
	// Should not panic
	_ = readOutput(ch, client)
}

func TestDoPractice_PracticeSkill(t *testing.T) {
	_ = setupSkillWorld()
	ch, client := makeTestChar("Warrior")
	defer client.Close()
	ch.Practice = 5
	ch.PCData.Learned[1] = 50 // backstab at 50%

	DoPractice(ch, "backstab")
	out := readOutput(ch, client)
	if !strings.Contains(out, "practice") || !strings.Contains(strings.ToLower(out), "backstab") {
		t.Errorf("expected practice confirmation, got: %q", out)
	}
	if ch.PCData.Learned[1] <= 50 {
		t.Error("expected skill to improve after practice")
	}
	if ch.Practice >= 5 {
		t.Error("expected practice sessions to decrease")
	}
}

func TestDoPractice_NoPractices(t *testing.T) {
	_ = setupSkillWorld()
	ch, client := makeTestChar("Warrior")
	defer client.Close()
	ch.Practice = 0
	ch.PCData.Learned[1] = 50

	DoPractice(ch, "backstab")
	out := readOutput(ch, client)
	if !strings.Contains(out, "no practice") && !strings.Contains(out, "no more practice") {
		t.Errorf("expected no practices message, got: %q", out)
	}
}

func TestDoPractice_SkillNotKnown(t *testing.T) {
	_ = setupSkillWorld()
	ch, client := makeTestChar("Warrior")
	defer client.Close()
	ch.Practice = 5

	DoPractice(ch, "nosuchskill")
	out := readOutput(ch, client)
	if !strings.Contains(out, "No such skill") && !strings.Contains(strings.ToLower(out), "no such") {
		t.Errorf("expected not found message, got: %q", out)
	}
}
