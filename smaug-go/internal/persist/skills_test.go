package persist

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func TestLoadSkills(t *testing.T) {
	w := world.New("testdata")

	err := LoadSkills(w, "testdata/test_skills.dat")
	if err != nil {
		t.Fatalf("LoadSkills: %v", err)
	}

	if len(w.Skills) != 3 {
		t.Fatalf("Skills count = %d, want 3", len(w.Skills))
	}

	// Test fireball
	fb := w.Skills[0]
	if fb.Name != "fireball" {
		t.Errorf("Skills[0].Name = %q, want %q", fb.Name, "fireball")
	}
	if fb.Type != types.SKILL_SPELL {
		t.Errorf("fireball.Type = %d, want SKILL_SPELL (%d)", fb.Type, types.SKILL_SPELL)
	}
	if fb.Target != 1 {
		t.Errorf("fireball.Target = %d, want 1", fb.Target)
	}
	if fb.Slot != 26 {
		t.Errorf("fireball.Slot = %d, want 26", fb.Slot)
	}
	if fb.MinMana != 15 {
		t.Errorf("fireball.MinMana = %d, want 15", fb.MinMana)
	}
	if fb.Beats != 12 {
		t.Errorf("fireball.Beats = %d, want 12", fb.Beats)
	}
	if fb.NounDamage != "fireball" {
		t.Errorf("fireball.NounDamage = %q, want %q", fb.NounDamage, "fireball")
	}
	if fb.SpellFunName != "spell_fireball" {
		t.Errorf("fireball.SpellFunName = %q, want %q", fb.SpellFunName, "spell_fireball")
	}
	if fb.MissChar != "Your fireball goes wide!" {
		t.Errorf("fireball.MissChar = %q", fb.MissChar)
	}
	if fb.Info != 5 {
		t.Errorf("fireball.Info = %d, want 5", fb.Info)
	}

	// Test backstab (skill, not spell)
	bs := w.Skills[1]
	if bs.Name != "backstab" {
		t.Errorf("Skills[1].Name = %q, want %q", bs.Name, "backstab")
	}
	if bs.Type != types.SKILL_SKILL {
		t.Errorf("backstab.Type = %d, want SKILL_SKILL (%d)", bs.Type, types.SKILL_SKILL)
	}
	if bs.SkillFunName != "do_backstab" {
		t.Errorf("backstab.SkillFunName = %q, want %q", bs.SkillFunName, "do_backstab")
	}
	if bs.Beats != 24 {
		t.Errorf("backstab.Beats = %d, want 24", bs.Beats)
	}

	// Test sanctuary (has affect)
	sanc := w.Skills[2]
	if sanc.Name != "sanctuary" {
		t.Errorf("Skills[2].Name = %q, want %q", sanc.Name, "sanctuary")
	}
	if sanc.MinMana != 75 {
		t.Errorf("sanctuary.MinMana = %d, want 75", sanc.MinMana)
	}
	if sanc.MsgOff != "The white aura around your body fades." {
		t.Errorf("sanctuary.MsgOff = %q", sanc.MsgOff)
	}
	if len(sanc.Affects) != 1 {
		t.Fatalf("sanctuary.Affects count = %d, want 1", len(sanc.Affects))
	}
	aff := sanc.Affects[0]
	if aff.Duration != "l*5+20" {
		t.Errorf("affect.Duration = %q, want %q", aff.Duration, "l*5+20")
	}
	if aff.Location != 0 {
		t.Errorf("affect.Location = %d, want 0", aff.Location)
	}
	if aff.Modifier != "0" {
		t.Errorf("affect.Modifier = %q, want %q", aff.Modifier, "0")
	}
	if aff.BitVector != 19 {
		t.Errorf("affect.BitVector = %d, want 19", aff.BitVector)
	}
}
