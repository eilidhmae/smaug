package magic

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func newMagicWorld() *world.World {
	w := world.New("/tmp/test")
	return w
}

func newCaster(name string, level int) *types.CharData {
	return &types.CharData{
		Name:     name,
		Level:    level,
		Position: types.POS_STANDING,
		Hit:      100, MaxHit: 100,
		Mana: 200, MaxMana: 200,
		Move: 100, MaxMove: 100,
		PermStr: 15, PermInt: 15, PermWis: 15,
		PermDex: 14, PermCon: 15, PermCha: 11, PermLck: 13,
		PCData: &types.PCData{},
	}
}

func TestSpellRegistry(t *testing.T) {
	fn := FindSpellFunc("spell_magic_missile")
	if fn == nil {
		t.Error("spell_magic_missile should be registered")
	}

	fn = FindSpellFunc("spell_nonexistent")
	if fn != nil {
		t.Error("nonexistent spell should return nil")
	}
}

func TestSpellCureLight(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9100, Name: "Temple"}
	ch := newCaster("Healer", 10)
	handler.CharToRoom(ch, room)
	ch.Hit = 50

	SpellCureLight(w, 0, ch.Level, ch, ch)

	if ch.Hit <= 50 {
		t.Errorf("Hit = %d, should have increased from 50", ch.Hit)
	}
	if ch.Hit > ch.MaxHit {
		t.Errorf("Hit = %d, should not exceed MaxHit %d", ch.Hit, ch.MaxHit)
	}
}

func TestSpellMagicMissile(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9101, Name: "Arena"}

	ch := newCaster("Wizard", 10)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	victim := &types.CharData{
		Name: "target", Level: 5, Position: types.POS_STANDING,
		Hit: 100, MaxHit: 100, Armor: 100,
	}
	victim.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	SpellMagicMissile(w, 0, ch.Level, ch, victim)

	if victim.Hit >= 100 {
		t.Errorf("victim.Hit = %d, should have taken damage", victim.Hit)
	}
}

func TestSpellArmor(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9102, Name: "Temple"}
	ch := newCaster("Caster", 10)
	handler.CharToRoom(ch, room)

	armorBefore := ch.Armor
	SpellArmor(w, 0, ch.Level, ch, ch)

	if ch.Armor >= armorBefore {
		t.Errorf("Armor = %d, should have decreased (improved) from %d", ch.Armor, armorBefore)
	}
	if len(ch.Affects) != 1 {
		t.Errorf("Affects = %d, want 1", len(ch.Affects))
	}
}

func TestSpellPoison(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9103, Name: "Arena"}

	ch := newCaster("Wizard", 20)
	handler.CharToRoom(ch, room)

	victim := &types.CharData{
		Name: "target", Level: 5, Position: types.POS_STANDING,
		Hit: 100, MaxHit: 100, PermStr: 15,
	}
	victim.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(victim, room)

	// Run many times — poison has a save, but at level 20 vs 5 should usually land
	landed := false
	for i := 0; i < 50; i++ {
		victim.Affects = nil
		victim.ModStr = 0
		victim.AffectedBy.Clear()
		SpellPoison(w, 0, ch.Level, ch, victim)
		if victim.AffectedBy.IsSet(types.AFF_POISON) {
			landed = true
			break
		}
	}
	if !landed {
		t.Error("poison should land at least once in 50 attempts (level 20 vs 5)")
	}
}

func TestSpellSanctuary(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9104, Name: "Temple"}
	ch := newCaster("Priest", 15)
	handler.CharToRoom(ch, room)

	SpellSanctuary(w, 0, ch.Level, ch, ch)

	if !ch.AffectedBy.IsSet(types.AFF_SANCTUARY) {
		t.Error("AFF_SANCTUARY should be set after sanctuary spell")
	}
}

func TestSavesSpellStaff(t *testing.T) {
	// High level victim vs low level spell = high save chance
	victim := &types.CharData{Level: 20}
	saved := 0
	for i := 0; i < 100; i++ {
		if SavesSpellStaff(5, victim) {
			saved++
		}
	}
	if saved < 50 {
		t.Errorf("high-level victim saved %d/100 vs low-level spell, expected mostly saves", saved)
	}
}

func TestFindSpellByName(t *testing.T) {
	w := newMagicWorld()
	// Add a skill to world
	sk := &types.SkillType{
		Name:         "magic missile",
		SpellFunName: "spell_magic_missile",
		Type:         types.SKILL_SPELL,
		Target:       types.TAR_CHAR_OFFENSIVE,
		MinMana:      10,
	}
	w.Skills = append(w.Skills, sk)

	sn := FindSpellByName(w, "magic missile")
	if sn < 0 {
		t.Error("FindSpellByName should find 'magic missile'")
	}

	sn = FindSpellByName(w, "nonexistent")
	if sn >= 0 {
		t.Error("FindSpellByName should return -1 for unknown spell")
	}
}
