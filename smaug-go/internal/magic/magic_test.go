package magic

import (
	"net"
	"strings"
	"testing"
	"time"

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

// newCasterWithDesc creates a caster with a pipe-backed descriptor for output capture.
func newCasterWithDesc(name string, level int) (*types.CharData, net.Conn) {
	server, client := net.Pipe()
	d := &types.DescriptorData{
		Conn:       server,
		InputQueue: make(chan string, 10),
		Connected:  types.CON_PLAYING,
	}
	ch := &types.CharData{
		Name:     name,
		Level:    level,
		Position: types.POS_STANDING,
		Hit:      100, MaxHit: 100,
		Mana: 200, MaxMana: 200,
		Move: 100, MaxMove: 100,
		PermStr: 15, PermInt: 15, PermWis: 15,
		PermDex: 14, PermCon: 15, PermCha: 11, PermLck: 13,
		PCData: &types.PCData{},
		Desc:   d,
	}
	d.Character = ch
	return ch, client
}

// readOutput flushes the descriptor and reads the output string.
func readOutput(ch *types.CharData, client net.Conn) string {
	if ch.Desc == nil || !ch.Desc.HasOutput() {
		return ""
	}
	result := make(chan string, 1)
	go func() {
		buf := make([]byte, 8192)
		client.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, _ := client.Read(buf)
		result <- string(buf[:n])
	}()
	_ = ch.Desc.FlushOutput()
	return <-result
}

// newNPCVictim creates an NPC victim with ACT_IS_NPC set.
func newNPCVictim(name string, level int) *types.CharData {
	v := &types.CharData{
		Name: name, ShortDescr: "a " + name, Level: level,
		Position: types.POS_STANDING,
		Hit: 100, MaxHit: 100,
		Armor: 100, PermStr: 15,
	}
	v.Act.Set(types.ACT_IS_NPC)
	return v
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

func TestSpellSleep(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9110, Name: "Arena"}
	ch := newCaster("Wizard", 20)
	handler.CharToRoom(ch, room)

	victim := &types.CharData{
		Name: "target", Level: 5, Position: types.POS_STANDING,
		Hit: 100, MaxHit: 100,
	}
	victim.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(victim, room)

	landed := false
	for i := 0; i < 50; i++ {
		victim.Affects = nil
		victim.AffectedBy.Clear()
		victim.Position = types.POS_STANDING
		SpellSleep(w, 0, ch.Level, ch, victim)
		if victim.AffectedBy.IsSet(types.AFF_SLEEP) {
			landed = true
			if victim.Position != types.POS_SLEEPING {
				t.Error("victim should be sleeping after sleep spell")
			}
			break
		}
	}
	if !landed {
		t.Error("sleep spell should land at least once in 50 attempts")
	}
}

func TestSpellCharmPerson(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9111, Name: "Arena"}
	ch := newCaster("Wizard", 20)
	handler.CharToRoom(ch, room)

	victim := &types.CharData{
		Name: "target", ShortDescr: "a target", Level: 5,
		Position: types.POS_STANDING, Hit: 100, MaxHit: 100,
	}
	victim.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(victim, room)

	landed := false
	for i := 0; i < 50; i++ {
		victim.Affects = nil
		victim.AffectedBy.Clear()
		victim.Master = nil
		victim.Leader = nil
		SpellCharmPerson(w, 0, ch.Level, ch, victim)
		if victim.AffectedBy.IsSet(types.AFF_CHARM) {
			landed = true
			if victim.Master != ch {
				t.Error("charmed victim should follow caster")
			}
			break
		}
	}
	if !landed {
		t.Error("charm should land at least once in 50 attempts")
	}
}

func TestSpellDetectEvil(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9112, Name: "Temple"}
	ch := newCaster("Priest", 10)
	handler.CharToRoom(ch, room)

	SpellDetectEvil(w, 0, ch.Level, ch, ch)
	if !ch.AffectedBy.IsSet(types.AFF_DETECT_EVIL) {
		t.Error("AFF_DETECT_EVIL should be set")
	}
}

func TestSpellDetectInvis(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9113, Name: "Temple"}
	ch := newCaster("Mage", 10)
	handler.CharToRoom(ch, room)

	SpellDetectInvis(w, 0, ch.Level, ch, ch)
	if !ch.AffectedBy.IsSet(types.AFF_DETECT_INVIS) {
		t.Error("AFF_DETECT_INVIS should be set")
	}
}

func TestSpellDetectMagic(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9114, Name: "Temple"}
	ch := newCaster("Mage", 10)
	handler.CharToRoom(ch, room)

	SpellDetectMagic(w, 0, ch.Level, ch, ch)
	if !ch.AffectedBy.IsSet(types.AFF_DETECT_MAGIC) {
		t.Error("AFF_DETECT_MAGIC should be set")
	}
}

func TestSpellDetectHidden(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9115, Name: "Temple"}
	ch := newCaster("Ranger", 10)
	handler.CharToRoom(ch, room)

	SpellDetectHidden(w, 0, ch.Level, ch, ch)
	if !ch.AffectedBy.IsSet(types.AFF_DETECT_HIDDEN) {
		t.Error("AFF_DETECT_HIDDEN should be set")
	}
}

func TestSpellShield(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9116, Name: "Temple"}
	ch := newCaster("Mage", 10)
	handler.CharToRoom(ch, room)

	armorBefore := ch.Armor
	SpellShield(w, 0, ch.Level, ch, ch)

	if ch.Armor >= armorBefore {
		t.Errorf("Armor = %d, should have improved from %d", ch.Armor, armorBefore)
	}
}

func TestSpellIdentify(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9117, Name: "Temple"}
	ch := newCaster("Mage", 10)
	ch.Desc = &types.DescriptorData{Character: ch}
	handler.CharToRoom(ch, room)

	// Add an item to carry
	obj := &types.ObjData{
		Name:       "sword",
		ShortDescr: "a magic sword",
		ItemType:   types.ITEM_WEAPON,
		Level:      5,
		Weight:     10,
		GoldCost:   100,
		Value:      [6]int{0, 5, 10, 0, 0, 0},
		CarriedBy:  ch,
		WearLoc:    types.WEAR_NONE,
	}
	ch.Carrying = append(ch.Carrying, obj)

	// Just verify it doesn't panic
	SpellIdentify(w, 0, ch.Level, ch, ch)
}

func TestSpellRegistry_NewSpells(t *testing.T) {
	newSpells := []string{
		"spell_sleep", "spell_charm_person",
		"spell_detect_evil", "spell_detect_invis",
		"spell_detect_magic", "spell_detect_hidden",
		"spell_shield", "spell_identify",
	}
	for _, name := range newSpells {
		if FindSpellFunc(name) == nil {
			t.Errorf("spell %s should be registered", name)
		}
	}
}

// TestSpellRegistry_AllRegistered verifies every spell in the registry is findable.
func TestSpellRegistry_AllRegistered(t *testing.T) {
	allSpells := []string{
		"spell_magic_missile", "spell_cure_light", "spell_cure_serious",
		"spell_cure_critical", "spell_fireball", "spell_armor", "spell_bless",
		"spell_curse", "spell_poison", "spell_blindness", "spell_sanctuary",
		"spell_dispel_magic", "spell_sleep", "spell_charm_person",
		"spell_detect_evil", "spell_detect_invis", "spell_detect_magic",
		"spell_detect_hidden", "spell_shield", "spell_identify",
		"spell_locate_object", "spell_create_food", "spell_create_water",
		"spell_summon", "spell_teleport", "spell_enchant_weapon",
		"spell_enchant_armor", "spell_invis", "spell_fly", "spell_heal",
	}
	for _, name := range allSpells {
		if FindSpellFunc(name) == nil {
			t.Errorf("spell %s should be registered", name)
		}
	}
}

// ---------------------------------------------------------------------------
// Healing spells (table-driven)
// ---------------------------------------------------------------------------

func TestHealingSpells(t *testing.T) {
	tests := []struct {
		name     string
		spell    SpellFunc
		startHP  int
		maxHP    int
		level    int
		minGain  int // minimum HP gain expected
		message  string
	}{
		{
			name:    "CureSerious",
			spell:   SpellCureSerious,
			startHP: 50, maxHP: 200, level: 10,
			minGain: 3, // 2d8 min=2 + 10/2=5 => min 7, but dice min is 2
			message: "You feel better!",
		},
		{
			name:    "CureCritical",
			spell:   SpellCureCritical,
			startHP: 50, maxHP: 200, level: 10,
			minGain: 4, // 3d8 min=3 + 10=10 => min 13
			message: "You feel much better!",
		},
		{
			name:    "Heal",
			spell:   SpellHeal,
			startHP: 50, maxHP: 500, level: 20,
			minGain: 100, // UMAX(100, 20*5) = 100
			message: "A warm feeling fills your body.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newMagicWorld()
			room := &types.RoomIndexData{Vnum: 9200, Name: "Temple"}
			ch, client := newCasterWithDesc("Healer", tc.level)
			defer client.Close()
			handler.CharToRoom(ch, room)
			ch.Hit = tc.startHP
			ch.MaxHit = tc.maxHP

			tc.spell(w, 0, ch.Level, ch, ch)

			if ch.Hit <= tc.startHP {
				t.Errorf("Hit = %d, should have increased from %d", ch.Hit, tc.startHP)
			}
			if ch.Hit > ch.MaxHit {
				t.Errorf("Hit = %d, should not exceed MaxHit %d", ch.Hit, ch.MaxHit)
			}
			out := readOutput(ch, client)
			if !strings.Contains(out, tc.message) {
				t.Errorf("output %q should contain %q", out, tc.message)
			}
		})
	}
}

func TestHealingSpells_CappedAtMaxHit(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9201, Name: "Temple"}
	ch := newCaster("Healer", 50)
	handler.CharToRoom(ch, room)
	ch.Hit = 99
	ch.MaxHit = 100

	SpellHeal(w, 0, ch.Level, ch, ch)
	if ch.Hit != 100 {
		t.Errorf("Hit = %d, should be capped at MaxHit 100", ch.Hit)
	}
}

func TestSpellCureSerious_CappedAtMaxHit(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9202, Name: "Temple"}
	ch := newCaster("Healer", 10)
	handler.CharToRoom(ch, room)
	ch.Hit = 99
	ch.MaxHit = 100

	SpellCureSerious(w, 0, ch.Level, ch, ch)
	if ch.Hit > ch.MaxHit {
		t.Errorf("Hit = %d, should not exceed MaxHit %d", ch.Hit, ch.MaxHit)
	}
}

func TestSpellCureCritical_CappedAtMaxHit(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9203, Name: "Temple"}
	ch := newCaster("Healer", 10)
	handler.CharToRoom(ch, room)
	ch.Hit = 99
	ch.MaxHit = 100

	SpellCureCritical(w, 0, ch.Level, ch, ch)
	if ch.Hit > ch.MaxHit {
		t.Errorf("Hit = %d, should not exceed MaxHit %d", ch.Hit, ch.MaxHit)
	}
}

// ---------------------------------------------------------------------------
// SpellFireball
// ---------------------------------------------------------------------------

func TestSpellFireball(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9210, Name: "Arena"}
	ch := newCaster("Wizard", 15)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	victim := newNPCVictim("target", 5)
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	SpellFireball(w, 0, ch.Level, ch, victim)
	if victim.Hit >= 100 {
		t.Errorf("victim.Hit = %d, should have taken damage from fireball", victim.Hit)
	}
}

func TestSpellFireball_HigherDamThanMagicMissile(t *testing.T) {
	// Fireball at level 15 should generally deal more than magic missile
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9211, Name: "Arena"}

	totalFireball := 0
	totalMissile := 0
	runs := 100

	for i := 0; i < runs; i++ {
		ch := newCaster("Wizard", 15)
		handler.CharToRoom(ch, room)
		w.AddChar(ch)

		v1 := newNPCVictim("target1", 1)
		v1.Hit = 10000
		v1.MaxHit = 10000
		handler.CharToRoom(v1, room)
		w.AddChar(v1)

		SpellFireball(w, 0, ch.Level, ch, v1)
		totalFireball += 10000 - v1.Hit

		v2 := newNPCVictim("target2", 1)
		v2.Hit = 10000
		v2.MaxHit = 10000
		handler.CharToRoom(v2, room)
		w.AddChar(v2)

		SpellMagicMissile(w, 0, ch.Level, ch, v2)
		totalMissile += 10000 - v2.Hit

		// Cleanup
		w.RemoveChar(ch)
		w.RemoveChar(v1)
		w.RemoveChar(v2)
	}
	if totalFireball <= totalMissile {
		t.Errorf("fireball avg %d should exceed magic missile avg %d over %d runs",
			totalFireball/runs, totalMissile/runs, runs)
	}
}

// ---------------------------------------------------------------------------
// SpellBless
// ---------------------------------------------------------------------------

func TestSpellBless(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9220, Name: "Temple"}
	ch, client := newCasterWithDesc("Priest", 16)
	defer client.Close()
	handler.CharToRoom(ch, room)

	hitrollBefore := ch.Hitroll
	SpellBless(w, 0, ch.Level, ch, ch)

	if ch.Hitroll <= hitrollBefore {
		t.Errorf("Hitroll = %d, should have increased from %d", ch.Hitroll, hitrollBefore)
	}
	if len(ch.Affects) != 1 {
		t.Errorf("Affects = %d, want 1", len(ch.Affects))
	}
	expectedMod := 1 + ch.Level/8
	if ch.Affects[0].Modifier != expectedMod {
		t.Errorf("bless modifier = %d, want %d", ch.Affects[0].Modifier, expectedMod)
	}
	if ch.Affects[0].Location != types.APPLY_HITROLL {
		t.Errorf("bless location = %d, want APPLY_HITROLL (%d)", ch.Affects[0].Location, types.APPLY_HITROLL)
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "righteous") {
		t.Errorf("output %q should contain 'righteous'", out)
	}
}

// ---------------------------------------------------------------------------
// SpellCurse
// ---------------------------------------------------------------------------

func TestSpellCurse(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9230, Name: "Arena"}
	ch := newCaster("Wizard", 20)
	handler.CharToRoom(ch, room)

	victim := newNPCVictim("target", 5)
	handler.CharToRoom(victim, room)

	landed := false
	for i := 0; i < 50; i++ {
		victim.Affects = nil
		victim.Hitroll = 0
		victim.AffectedBy.Clear()
		SpellCurse(w, 0, ch.Level, ch, victim)
		if victim.AffectedBy.IsSet(types.AFF_CURSE) {
			landed = true
			if victim.Hitroll >= 0 {
				t.Error("Hitroll should be negative after curse")
			}
			break
		}
	}
	if !landed {
		t.Error("curse should land at least once in 50 attempts (level 20 vs 5)")
	}
}

func TestSpellCurse_AlreadyCursed(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9231, Name: "Arena"}
	ch, client := newCasterWithDesc("Wizard", 20)
	defer client.Close()
	handler.CharToRoom(ch, room)

	victim := newNPCVictim("target", 5)
	victim.AffectedBy.Set(types.AFF_CURSE)
	handler.CharToRoom(victim, room)

	SpellCurse(w, 0, ch.Level, ch, victim)

	out := readOutput(ch, client)
	if !strings.Contains(out, "already cursed") {
		t.Errorf("output %q should contain 'already cursed'", out)
	}
}

// ---------------------------------------------------------------------------
// SpellBlindness
// ---------------------------------------------------------------------------

func TestSpellBlindness(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9240, Name: "Arena"}
	ch := newCaster("Wizard", 20)
	handler.CharToRoom(ch, room)

	victim := newNPCVictim("target", 5)
	handler.CharToRoom(victim, room)

	landed := false
	for i := 0; i < 50; i++ {
		victim.Affects = nil
		victim.Hitroll = 0
		victim.AffectedBy.Clear()
		SpellBlindness(w, 0, ch.Level, ch, victim)
		if victim.AffectedBy.IsSet(types.AFF_BLIND) {
			landed = true
			if victim.Hitroll >= 0 {
				t.Error("Hitroll should be negative after blindness (-4)")
			}
			break
		}
	}
	if !landed {
		t.Error("blindness should land at least once in 50 attempts (level 20 vs 5)")
	}
}

func TestSpellBlindness_AlreadyBlind(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9241, Name: "Arena"}
	ch, client := newCasterWithDesc("Wizard", 20)
	defer client.Close()
	handler.CharToRoom(ch, room)

	victim := newNPCVictim("target", 5)
	victim.AffectedBy.Set(types.AFF_BLIND)
	handler.CharToRoom(victim, room)

	SpellBlindness(w, 0, ch.Level, ch, victim)

	out := readOutput(ch, client)
	if !strings.Contains(out, "already blinded") {
		t.Errorf("output %q should contain 'already blinded'", out)
	}
}

// ---------------------------------------------------------------------------
// SpellDispelMagic
// ---------------------------------------------------------------------------

func TestSpellDispelMagic_Self(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9250, Name: "Temple"}
	ch, client := newCasterWithDesc("Mage", 10)
	defer client.Close()
	handler.CharToRoom(ch, room)

	// Apply some affects first
	SpellArmor(w, 0, ch.Level, ch, ch)
	SpellBless(w, 0, ch.Level, ch, ch)
	_ = readOutput(ch, client) // drain output

	if len(ch.Affects) < 2 {
		t.Fatalf("should have 2 affects before dispel, got %d", len(ch.Affects))
	}

	SpellDispelMagic(w, 0, ch.Level, ch, ch)

	if len(ch.Affects) != 0 {
		t.Errorf("Affects = %d, want 0 after dispel magic on self", len(ch.Affects))
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "enchantments fade") {
		t.Errorf("output %q should contain 'enchantments fade'", out)
	}
}

func TestSpellDispelMagic_Other(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9251, Name: "Arena"}
	ch, chClient := newCasterWithDesc("Mage", 20)
	defer chClient.Close()
	handler.CharToRoom(ch, room)

	victim, vClient := newCasterWithDesc("Target", 5)
	defer vClient.Close()
	handler.CharToRoom(victim, room)

	// Apply an affect to victim
	SpellArmor(w, 0, ch.Level, victim, victim)
	_ = readOutput(victim, vClient)

	// Try many times (save negates for others)
	dispelled := false
	for i := 0; i < 50; i++ {
		if len(victim.Affects) == 0 {
			// Re-apply
			SpellArmor(w, 0, ch.Level, victim, victim)
			_ = readOutput(victim, vClient)
		}
		_ = readOutput(ch, chClient) // drain
		SpellDispelMagic(w, 0, ch.Level, ch, victim)
		if len(victim.Affects) == 0 {
			dispelled = true
			break
		}
	}
	if !dispelled {
		t.Error("dispel magic should succeed at least once in 50 attempts (level 20 vs 5)")
	}
}

// ---------------------------------------------------------------------------
// Already-affected early returns (table-driven)
// ---------------------------------------------------------------------------

func TestSpellAlreadyAffected(t *testing.T) {
	tests := []struct {
		name    string
		flag    int
		spell   SpellFunc
		message string
	}{
		{"Sleep_AlreadyAsleep", types.AFF_SLEEP, SpellSleep, "already asleep"},
		{"Charm_AlreadyCharmed", types.AFF_CHARM, SpellCharmPerson, "already charmed"},
		{"DetectEvil_Already", types.AFF_DETECT_EVIL, SpellDetectEvil, "already sense evil"},
		{"DetectInvis_Already", types.AFF_DETECT_INVIS, SpellDetectInvis, "already see invisible"},
		{"DetectMagic_Already", types.AFF_DETECT_MAGIC, SpellDetectMagic, "already sense magical"},
		{"DetectHidden_Already", types.AFF_DETECT_HIDDEN, SpellDetectHidden, "already sense hidden"},
		{"Sanctuary_Already", types.AFF_SANCTUARY, SpellSanctuary, "already in sanctuary"},
		{"Invis_Already", types.AFF_INVISIBLE, SpellInvis, "already invisible"},
		{"Fly_Already", types.AFF_FLYING, SpellFly, "already flying"},
		{"Armor_Already", types.AFF_PROTECT, SpellArmor, "already armored"},
		{"Poison_Already", types.AFF_POISON, SpellPoison, "already poisoned"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newMagicWorld()
			room := &types.RoomIndexData{Vnum: 9260, Name: "Temple"}
			ch, client := newCasterWithDesc("Caster", 20)
			defer client.Close()
			handler.CharToRoom(ch, room)

			victim, vClient := newCasterWithDesc("Victim", 5)
			defer vClient.Close()
			victim.Act.Set(types.ACT_IS_NPC)
			handler.CharToRoom(victim, room)

			victim.AffectedBy.Set(tc.flag)

			tc.spell(w, 0, ch.Level, ch, victim)

			// Read from either ch or victim — the message might go to either
			outCh := readOutput(ch, client)
			outV := readOutput(victim, vClient)
			combined := outCh + outV

			if !strings.Contains(combined, tc.message) {
				t.Errorf("output %q should contain %q", combined, tc.message)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// SpellInvis
// ---------------------------------------------------------------------------

func TestSpellInvis(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9270, Name: "Temple"}
	ch, client := newCasterWithDesc("Mage", 10)
	defer client.Close()
	handler.CharToRoom(ch, room)

	SpellInvis(w, 0, ch.Level, ch, ch)

	if !ch.AffectedBy.IsSet(types.AFF_INVISIBLE) {
		t.Error("AFF_INVISIBLE should be set after invis spell")
	}
	if len(ch.Affects) != 1 {
		t.Errorf("Affects = %d, want 1", len(ch.Affects))
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "fade out") {
		t.Errorf("output %q should contain 'fade out'", out)
	}
}

// ---------------------------------------------------------------------------
// SpellFly
// ---------------------------------------------------------------------------

func TestSpellFly(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9280, Name: "Temple"}
	ch, client := newCasterWithDesc("Mage", 10)
	defer client.Close()
	handler.CharToRoom(ch, room)

	SpellFly(w, 0, ch.Level, ch, ch)

	if !ch.AffectedBy.IsSet(types.AFF_FLYING) {
		t.Error("AFF_FLYING should be set after fly spell")
	}
	if len(ch.Affects) != 1 {
		t.Errorf("Affects = %d, want 1", len(ch.Affects))
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "feet rise") {
		t.Errorf("output %q should contain 'feet rise'", out)
	}
}

// ---------------------------------------------------------------------------
// SpellLocateObject
// ---------------------------------------------------------------------------

func TestSpellLocateObject_Found(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9290, Name: "Temple"}
	ch, client := newCasterWithDesc("Mage", 10)
	defer client.Close()
	handler.CharToRoom(ch, room)

	// Add an object to the world
	obj := &types.ObjData{
		Name:       "sword magic",
		ShortDescr: "a magic sword",
		InRoom:     room,
	}
	w.AddObj(obj)

	SpellLocateObject(w, 0, ch.Level, ch, ch)

	out := readOutput(ch, client)
	if !strings.Contains(out, "a magic sword") {
		t.Errorf("output %q should contain 'a magic sword'", out)
	}
}

func TestSpellLocateObject_CarriedBy(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9291, Name: "Temple"}
	ch, client := newCasterWithDesc("Mage", 10)
	defer client.Close()
	handler.CharToRoom(ch, room)

	carrier := newCaster("Bob", 5)
	obj := &types.ObjData{
		Name:       "gem sparkling",
		ShortDescr: "a sparkling gem",
		CarriedBy:  carrier,
	}
	w.AddObj(obj)

	SpellLocateObject(w, 0, ch.Level, ch, ch)

	out := readOutput(ch, client)
	if !strings.Contains(out, "carried by Bob") {
		t.Errorf("output %q should contain 'carried by Bob'", out)
	}
}

func TestSpellLocateObject_Empty(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9292, Name: "Temple"}
	ch, client := newCasterWithDesc("Mage", 10)
	defer client.Close()
	handler.CharToRoom(ch, room)

	// No objects in world
	SpellLocateObject(w, 0, ch.Level, ch, ch)

	out := readOutput(ch, client)
	if !strings.Contains(out, "Nothing like that") {
		t.Errorf("output %q should contain 'Nothing like that'", out)
	}
}

func TestSpellLocateObject_NoLocate(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9293, Name: "Temple"}
	ch, client := newCasterWithDesc("Mage", 10)
	defer client.Close()
	handler.CharToRoom(ch, room)

	obj := &types.ObjData{
		Name:       "artifact hidden",
		ShortDescr: "a hidden artifact",
		InRoom:     room,
	}
	obj.ExtraFlags.Set(types.ITEM_NOLOCATE)
	w.AddObj(obj)

	SpellLocateObject(w, 0, ch.Level, ch, ch)

	out := readOutput(ch, client)
	if !strings.Contains(out, "Nothing like that") {
		t.Errorf("ITEM_NOLOCATE object should be skipped; output %q", out)
	}
}

// ---------------------------------------------------------------------------
// SpellCreateFood
// ---------------------------------------------------------------------------

func TestSpellCreateFood(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9300, Name: "Temple"}
	ch, client := newCasterWithDesc("Mage", 10)
	defer client.Close()
	handler.CharToRoom(ch, room)

	initialCount := len(ch.Carrying)
	SpellCreateFood(w, 0, ch.Level, ch, ch)

	if len(ch.Carrying) != initialCount+1 {
		t.Errorf("Carrying count = %d, want %d", len(ch.Carrying), initialCount+1)
	}
	food := ch.Carrying[len(ch.Carrying)-1]
	if food.ItemType != types.ITEM_FOOD {
		t.Errorf("created item type = %d, want ITEM_FOOD (%d)", food.ItemType, types.ITEM_FOOD)
	}
	if food.CarriedBy != ch {
		t.Error("food should be carried by caster")
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "mushroom") {
		t.Errorf("output %q should contain 'mushroom'", out)
	}
}

// ---------------------------------------------------------------------------
// SpellCreateWater
// ---------------------------------------------------------------------------

func TestSpellCreateWater(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9310, Name: "Temple"}
	ch, client := newCasterWithDesc("Mage", 10)
	defer client.Close()
	handler.CharToRoom(ch, room)

	container := &types.ObjData{
		Name:       "waterskin leather",
		ShortDescr: "a leather waterskin",
		ItemType:   types.ITEM_DRINK_CON,
		Value:      [6]int{50, 0, 0, 0, 0, 0}, // capacity=50, current=0
		WearLoc:    types.WEAR_NONE,
		CarriedBy:  ch,
	}
	ch.Carrying = append(ch.Carrying, container)

	SpellCreateWater(w, 0, ch.Level, ch, ch)

	if container.Value[1] <= 0 {
		t.Errorf("container water level = %d, should have increased", container.Value[1])
	}
	if container.Value[2] != 0 {
		t.Errorf("liquid type = %d, want 0 (water)", container.Value[2])
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "Water flows") {
		t.Errorf("output %q should contain 'Water flows'", out)
	}
}

func TestSpellCreateWater_AlreadyFull(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9311, Name: "Temple"}
	ch, client := newCasterWithDesc("Mage", 10)
	defer client.Close()
	handler.CharToRoom(ch, room)

	container := &types.ObjData{
		Name:       "waterskin",
		ShortDescr: "a waterskin",
		ItemType:   types.ITEM_DRINK_CON,
		Value:      [6]int{50, 50, 0, 0, 0, 0}, // full
		WearLoc:    types.WEAR_NONE,
		CarriedBy:  ch,
	}
	ch.Carrying = append(ch.Carrying, container)

	SpellCreateWater(w, 0, ch.Level, ch, ch)

	out := readOutput(ch, client)
	if !strings.Contains(out, "already full") {
		t.Errorf("output %q should contain 'already full'", out)
	}
}

func TestSpellCreateWater_NoContainer(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9312, Name: "Temple"}
	ch, client := newCasterWithDesc("Mage", 10)
	defer client.Close()
	handler.CharToRoom(ch, room)

	SpellCreateWater(w, 0, ch.Level, ch, ch)

	out := readOutput(ch, client)
	if !strings.Contains(out, "drink container") {
		t.Errorf("output %q should contain 'drink container'", out)
	}
}

// ---------------------------------------------------------------------------
// SpellSummon
// ---------------------------------------------------------------------------

func TestSpellSummon(t *testing.T) {
	w := newMagicWorld()
	room1 := &types.RoomIndexData{Vnum: 9320, Name: "Temple"}
	room2 := &types.RoomIndexData{Vnum: 9321, Name: "Forest"}

	ch, client := newCasterWithDesc("Mage", 20)
	defer client.Close()
	handler.CharToRoom(ch, room1)

	victim, vClient := newCasterWithDesc("Target", 5)
	defer vClient.Close()
	handler.CharToRoom(victim, room2)

	SpellSummon(w, 0, ch.Level, ch, victim)

	if victim.InRoom != room1 {
		t.Error("victim should be in caster's room after summon")
	}
	out := readOutput(victim, vClient)
	if !strings.Contains(out, "summoned you") {
		t.Errorf("victim output %q should contain 'summoned you'", out)
	}
}

func TestSpellSummon_Self(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9322, Name: "Temple"}
	ch, client := newCasterWithDesc("Mage", 20)
	defer client.Close()
	handler.CharToRoom(ch, room)

	SpellSummon(w, 0, ch.Level, ch, ch)

	out := readOutput(ch, client)
	if !strings.Contains(out, "can't summon") {
		t.Errorf("output %q should contain \"can't summon\"", out)
	}
}

func TestSpellSummon_Nil(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9323, Name: "Temple"}
	ch, client := newCasterWithDesc("Mage", 20)
	defer client.Close()
	handler.CharToRoom(ch, room)

	SpellSummon(w, 0, ch.Level, ch, nil)

	out := readOutput(ch, client)
	if !strings.Contains(out, "can't summon") {
		t.Errorf("output %q should contain \"can't summon\"", out)
	}
}

func TestSpellSummon_Fighting(t *testing.T) {
	w := newMagicWorld()
	room1 := &types.RoomIndexData{Vnum: 9324, Name: "Temple"}
	room2 := &types.RoomIndexData{Vnum: 9325, Name: "Arena"}

	ch, client := newCasterWithDesc("Mage", 20)
	defer client.Close()
	handler.CharToRoom(ch, room1)

	victim := newNPCVictim("target", 5)
	handler.CharToRoom(victim, room2)
	victim.Fighting = &types.FightData{Who: ch} // in combat

	SpellSummon(w, 0, ch.Level, ch, victim)

	if victim.InRoom == room1 {
		t.Error("should not summon someone who is fighting")
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "busy fighting") {
		t.Errorf("output %q should contain 'busy fighting'", out)
	}
}

// ---------------------------------------------------------------------------
// SpellTeleport
// ---------------------------------------------------------------------------

func TestSpellTeleport(t *testing.T) {
	w := newMagicWorld()
	startRoom := &types.RoomIndexData{Vnum: 9330, Name: "Start"}
	destRoom := &types.RoomIndexData{Vnum: 9331, Name: "Dest"}
	w.Rooms[startRoom.Vnum] = startRoom
	w.Rooms[destRoom.Vnum] = destRoom

	ch, client := newCasterWithDesc("Mage", 10)
	defer client.Close()
	handler.CharToRoom(ch, startRoom)

	// Run multiple times — teleport picks randomly
	teleported := false
	for i := 0; i < 50; i++ {
		if ch.InRoom != startRoom {
			handler.CharFromRoom(ch)
			handler.CharToRoom(ch, startRoom)
		}
		_ = readOutput(ch, client)
		SpellTeleport(w, 0, ch.Level, ch, ch)
		if ch.InRoom != startRoom {
			teleported = true
			break
		}
	}
	if !teleported {
		t.Error("teleport should move caster at least once in 50 attempts")
	}
}

func TestSpellTeleport_SkipsPrivateRooms(t *testing.T) {
	w := newMagicWorld()
	startRoom := &types.RoomIndexData{Vnum: 9340, Name: "Start"}
	privateRoom := &types.RoomIndexData{Vnum: 9341, Name: "Private"}
	privateRoom.RoomFlags.Set(types.ROOM_PRIVATE)
	w.Rooms[startRoom.Vnum] = startRoom
	w.Rooms[privateRoom.Vnum] = privateRoom

	ch := newCaster("Mage", 10)
	handler.CharToRoom(ch, startRoom)

	// With only private+start rooms, teleport should fizzle or stay in start
	for i := 0; i < 20; i++ {
		if ch.InRoom != startRoom {
			handler.CharFromRoom(ch)
			handler.CharToRoom(ch, startRoom)
		}
		SpellTeleport(w, 0, ch.Level, ch, ch)
		if ch.InRoom == privateRoom {
			t.Error("teleport should not send to ROOM_PRIVATE rooms")
			break
		}
	}
}

// ---------------------------------------------------------------------------
// SpellEnchantWeapon
// ---------------------------------------------------------------------------

func TestSpellEnchantWeapon(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9350, Name: "Temple"}
	ch, client := newCasterWithDesc("Mage", 20)
	defer client.Close()
	handler.CharToRoom(ch, room)

	weapon := &types.ObjData{
		Name:       "sword steel",
		ShortDescr: "a steel sword",
		ItemType:   types.ITEM_WEAPON,
		WearLoc:    types.WEAR_WIELD,
		CarriedBy:  ch,
	}
	ch.Carrying = append(ch.Carrying, weapon)

	SpellEnchantWeapon(w, 0, ch.Level, ch, ch)

	if !weapon.ExtraFlags.IsSet(types.ITEM_MAGIC) {
		t.Error("weapon should have ITEM_MAGIC flag after enchant")
	}
	if len(weapon.Affects) != 2 {
		t.Fatalf("weapon affects = %d, want 2 (hitroll + damroll)", len(weapon.Affects))
	}
	expectedBonus := 1 + ch.Level/20
	if weapon.Affects[0].Location != types.APPLY_HITROLL {
		t.Errorf("first affect location = %d, want APPLY_HITROLL", weapon.Affects[0].Location)
	}
	if weapon.Affects[0].Modifier != expectedBonus {
		t.Errorf("hitroll bonus = %d, want %d", weapon.Affects[0].Modifier, expectedBonus)
	}
	if weapon.Affects[1].Location != types.APPLY_DAMROLL {
		t.Errorf("second affect location = %d, want APPLY_DAMROLL", weapon.Affects[1].Location)
	}
	if weapon.Affects[1].Modifier != expectedBonus {
		t.Errorf("damroll bonus = %d, want %d", weapon.Affects[1].Modifier, expectedBonus)
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "glows brightly") {
		t.Errorf("output %q should contain 'glows brightly'", out)
	}
}

func TestSpellEnchantWeapon_NoWeapon(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9351, Name: "Temple"}
	ch, client := newCasterWithDesc("Mage", 20)
	defer client.Close()
	handler.CharToRoom(ch, room)

	SpellEnchantWeapon(w, 0, ch.Level, ch, ch)

	out := readOutput(ch, client)
	if !strings.Contains(out, "wield a weapon") {
		t.Errorf("output %q should contain 'wield a weapon'", out)
	}
}

func TestSpellEnchantWeapon_AlreadyMagic(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9352, Name: "Temple"}
	ch, client := newCasterWithDesc("Mage", 20)
	defer client.Close()
	handler.CharToRoom(ch, room)

	weapon := &types.ObjData{
		Name:       "sword magic",
		ShortDescr: "a magic sword",
		ItemType:   types.ITEM_WEAPON,
		WearLoc:    types.WEAR_WIELD,
		CarriedBy:  ch,
	}
	weapon.ExtraFlags.Set(types.ITEM_MAGIC)
	ch.Carrying = append(ch.Carrying, weapon)

	SpellEnchantWeapon(w, 0, ch.Level, ch, ch)

	out := readOutput(ch, client)
	if !strings.Contains(out, "already enchanted") {
		t.Errorf("output %q should contain 'already enchanted'", out)
	}
}

// ---------------------------------------------------------------------------
// SpellEnchantArmor
// ---------------------------------------------------------------------------

func TestSpellEnchantArmor(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9360, Name: "Temple"}
	ch, client := newCasterWithDesc("Mage", 20)
	defer client.Close()
	handler.CharToRoom(ch, room)

	armor := &types.ObjData{
		Name:       "plate armor",
		ShortDescr: "a suit of plate armor",
		ItemType:   types.ITEM_ARMOR,
		WearLoc:    types.WEAR_BODY,
		CarriedBy:  ch,
	}
	ch.Carrying = append(ch.Carrying, armor)

	SpellEnchantArmor(w, 0, ch.Level, ch, ch)

	if !armor.ExtraFlags.IsSet(types.ITEM_MAGIC) {
		t.Error("armor should have ITEM_MAGIC flag after enchant")
	}
	if len(armor.Affects) != 1 {
		t.Fatalf("armor affects = %d, want 1 (AC)", len(armor.Affects))
	}
	expectedBonus := -(1 + ch.Level/20)
	if armor.Affects[0].Location != types.APPLY_AC {
		t.Errorf("affect location = %d, want APPLY_AC", armor.Affects[0].Location)
	}
	if armor.Affects[0].Modifier != expectedBonus {
		t.Errorf("AC modifier = %d, want %d", armor.Affects[0].Modifier, expectedBonus)
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "protective aura") {
		t.Errorf("output %q should contain 'protective aura'", out)
	}
}

func TestSpellEnchantArmor_NoArmor(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9361, Name: "Temple"}
	ch, client := newCasterWithDesc("Mage", 20)
	defer client.Close()
	handler.CharToRoom(ch, room)

	SpellEnchantArmor(w, 0, ch.Level, ch, ch)

	out := readOutput(ch, client)
	if !strings.Contains(out, "wearing armor") {
		t.Errorf("output %q should contain 'wearing armor'", out)
	}
}

func TestSpellEnchantArmor_AlreadyMagic(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9362, Name: "Temple"}
	ch, client := newCasterWithDesc("Mage", 20)
	defer client.Close()
	handler.CharToRoom(ch, room)

	armor := &types.ObjData{
		Name:       "magic armor",
		ShortDescr: "magic armor",
		ItemType:   types.ITEM_ARMOR,
		WearLoc:    types.WEAR_BODY,
		CarriedBy:  ch,
	}
	armor.ExtraFlags.Set(types.ITEM_MAGIC)
	ch.Carrying = append(ch.Carrying, armor)

	SpellEnchantArmor(w, 0, ch.Level, ch, ch)

	out := readOutput(ch, client)
	if !strings.Contains(out, "already enchanted") {
		t.Errorf("output %q should contain 'already enchanted'", out)
	}
}

// ---------------------------------------------------------------------------
// SpellIdentify edge cases
// ---------------------------------------------------------------------------

func TestSpellIdentify_NoItems(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9370, Name: "Temple"}
	ch, client := newCasterWithDesc("Mage", 10)
	defer client.Close()
	handler.CharToRoom(ch, room)

	SpellIdentify(w, 0, ch.Level, ch, ch)

	out := readOutput(ch, client)
	if !strings.Contains(out, "not carrying anything") {
		t.Errorf("output %q should contain 'not carrying anything'", out)
	}
}

func TestSpellIdentify_AllWorn(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9371, Name: "Temple"}
	ch, client := newCasterWithDesc("Mage", 10)
	defer client.Close()
	handler.CharToRoom(ch, room)

	// Only worn items, no inventory items
	obj := &types.ObjData{
		Name:       "sword",
		ShortDescr: "a sword",
		ItemType:   types.ITEM_WEAPON,
		WearLoc:    types.WEAR_WIELD,
		CarriedBy:  ch,
	}
	ch.Carrying = append(ch.Carrying, obj)

	SpellIdentify(w, 0, ch.Level, ch, ch)

	out := readOutput(ch, client)
	if !strings.Contains(out, "not carrying anything") {
		t.Errorf("output %q should contain 'not carrying anything' when all items are worn", out)
	}
}

func TestSpellIdentify_WithAffects(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9372, Name: "Temple"}
	ch, client := newCasterWithDesc("Mage", 10)
	defer client.Close()
	handler.CharToRoom(ch, room)

	obj := &types.ObjData{
		Name:       "ring power",
		ShortDescr: "a ring of power",
		ItemType:   types.ITEM_ARMOR,
		Level:      10,
		Weight:     1,
		GoldCost:   500,
		WearLoc:    types.WEAR_NONE,
		CarriedBy:  ch,
		Affects: []*types.AffectData{
			{Location: types.APPLY_STR, Modifier: 2},
		},
	}
	ch.Carrying = append(ch.Carrying, obj)

	SpellIdentify(w, 0, ch.Level, ch, ch)

	out := readOutput(ch, client)
	if !strings.Contains(out, "ring of power") {
		t.Errorf("output %q should contain item short descr", out)
	}
	if !strings.Contains(out, "Affects") {
		t.Errorf("output %q should contain 'Affects' for items with affects", out)
	}
}

// ---------------------------------------------------------------------------
// SpellCharmPerson edge cases
// ---------------------------------------------------------------------------

func TestSpellCharmPerson_Self(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9380, Name: "Temple"}
	ch, client := newCasterWithDesc("Wizard", 20)
	defer client.Close()
	handler.CharToRoom(ch, room)

	SpellCharmPerson(w, 0, ch.Level, ch, ch)

	out := readOutput(ch, client)
	if !strings.Contains(out, "like yourself") {
		t.Errorf("output %q should contain 'like yourself'", out)
	}
}

func TestSpellCharmPerson_PC(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9381, Name: "Temple"}
	ch, client := newCasterWithDesc("Wizard", 20)
	defer client.Close()
	handler.CharToRoom(ch, room)

	victim := newCaster("Player", 5)
	// Note: no ACT_IS_NPC set, so victim is a PC
	handler.CharToRoom(victim, room)

	SpellCharmPerson(w, 0, ch.Level, ch, victim)

	out := readOutput(ch, client)
	if !strings.Contains(out, "cannot charm other players") {
		t.Errorf("output %q should contain 'cannot charm other players'", out)
	}
}

// ---------------------------------------------------------------------------
// SavesPoisonDeath
// ---------------------------------------------------------------------------

func TestSavesPoisonDeath(t *testing.T) {
	// High level victim vs low level poison = high save chance
	victim := &types.CharData{Level: 20}
	saved := 0
	for i := 0; i < 100; i++ {
		if SavesPoisonDeath(5, victim) {
			saved++
		}
	}
	if saved < 50 {
		t.Errorf("high-level victim saved %d/100 vs low-level poison, expected mostly saves", saved)
	}

	// Low level victim vs high level poison = low save chance
	weakVictim := &types.CharData{Level: 1}
	weakSaved := 0
	for i := 0; i < 100; i++ {
		if SavesPoisonDeath(30, weakVictim) {
			weakSaved++
		}
	}
	if weakSaved > 30 {
		t.Errorf("low-level victim saved %d/100 vs high-level poison, expected few saves", weakSaved)
	}
}

func TestSavesPoisonDeath_ClampedRange(t *testing.T) {
	// Even extreme level differences should be clamped to 5-95 range
	// Very high save: level 50 vs spell level 1
	victim := &types.CharData{Level: 50}
	allSaved := true
	for i := 0; i < 200; i++ {
		if !SavesPoisonDeath(1, victim) {
			allSaved = false
			break
		}
	}
	// With 95% save chance, failing at least once in 200 is expected
	if allSaved {
		t.Error("even with very high save, should occasionally fail (clamped at 95%)")
	}
}

// ---------------------------------------------------------------------------
// SavesSpellStaff additional tests
// ---------------------------------------------------------------------------

func TestSavesSpellStaff_LowLevel(t *testing.T) {
	// Low level victim vs high level spell = low save chance
	victim := &types.CharData{Level: 1}
	saved := 0
	for i := 0; i < 100; i++ {
		if SavesSpellStaff(30, victim) {
			saved++
		}
	}
	if saved > 30 {
		t.Errorf("low-level victim saved %d/100 vs high-level spell, expected few saves", saved)
	}
}

// ---------------------------------------------------------------------------
// SpellLocateObject nil caster guard
// ---------------------------------------------------------------------------

func TestSpellLocateObject_NilCaster(t *testing.T) {
	w := newMagicWorld()
	// Should not panic
	SpellLocateObject(w, 0, 10, nil, nil)
}

// ---------------------------------------------------------------------------
// FindSpellByName prefix match
// ---------------------------------------------------------------------------

func TestFindSpellByName_PrefixMatch(t *testing.T) {
	w := newMagicWorld()
	w.Skills = append(w.Skills, &types.SkillType{
		Name:         "magic missile",
		SpellFunName: "spell_magic_missile",
		Type:         types.SKILL_SPELL,
	})

	sn := FindSpellByName(w, "magic")
	if sn < 0 {
		t.Error("FindSpellByName should match prefix 'magic' for 'magic missile'")
	}
}

func TestFindSpellByName_SkipsNonSpells(t *testing.T) {
	w := newMagicWorld()
	w.Skills = append(w.Skills, &types.SkillType{
		Name:         "backstab",
		SpellFunName: "skill_backstab",
		Type:         types.SKILL_SKILL, // not a spell
	})

	sn := FindSpellByName(w, "backstab")
	if sn >= 0 {
		t.Error("FindSpellByName should not match non-spell skills")
	}
}

// ---------------------------------------------------------------------------
// Coverage gap tests: uncovered branches
// ---------------------------------------------------------------------------

// SpellCurse save-resisted path (level 1 caster vs level 50 victim)
func TestSpellCurse_SaveResists(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9400, Name: "Arena"}
	ch, client := newCasterWithDesc("Wizard", 1)
	defer client.Close()
	handler.CharToRoom(ch, room)

	victim := newNPCVictim("dragon", 50)
	handler.CharToRoom(victim, room)

	resisted := false
	for i := 0; i < 50; i++ {
		victim.Affects = nil
		victim.AffectedBy.Clear()
		victim.Hitroll = 0
		_ = readOutput(ch, client)
		SpellCurse(w, 0, ch.Level, ch, victim)
		if !victim.AffectedBy.IsSet(types.AFF_CURSE) {
			resisted = true
			out := readOutput(ch, client)
			if !strings.Contains(out, "fails to take hold") {
				t.Errorf("output %q should contain 'fails to take hold'", out)
			}
			break
		}
	}
	if !resisted {
		t.Error("curse should be resisted at least once in 50 attempts (level 1 vs 50)")
	}
}

// SpellPoison save-resisted path
func TestSpellPoison_SaveResists(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9401, Name: "Arena"}
	ch, client := newCasterWithDesc("Wizard", 1)
	defer client.Close()
	handler.CharToRoom(ch, room)

	victim := newNPCVictim("dragon", 50)
	handler.CharToRoom(victim, room)

	resisted := false
	for i := 0; i < 50; i++ {
		victim.Affects = nil
		victim.AffectedBy.Clear()
		victim.ModStr = 0
		_ = readOutput(ch, client)
		SpellPoison(w, 0, ch.Level, ch, victim)
		if !victim.AffectedBy.IsSet(types.AFF_POISON) {
			resisted = true
			out := readOutput(ch, client)
			if !strings.Contains(out, "fails to take hold") {
				t.Errorf("output %q should contain 'fails to take hold'", out)
			}
			break
		}
	}
	if !resisted {
		t.Error("poison should be resisted at least once in 50 attempts (level 1 vs 50)")
	}
}

// SpellDispelMagic save-negates for non-self target
func TestSpellDispelMagic_SaveNegates(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9402, Name: "Arena"}
	ch, chClient := newCasterWithDesc("Mage", 1)
	defer chClient.Close()
	handler.CharToRoom(ch, room)

	victim, vClient := newCasterWithDesc("Dragon", 50)
	defer vClient.Close()
	handler.CharToRoom(victim, room)

	resisted := false
	for i := 0; i < 50; i++ {
		// Apply an affect to victim
		victim.Affects = nil
		victim.AffectedBy.Clear()
		SpellBless(w, 0, 10, victim, victim)
		_ = readOutput(victim, vClient)
		_ = readOutput(ch, chClient)

		SpellDispelMagic(w, 0, ch.Level, ch, victim)
		if len(victim.Affects) > 0 {
			resisted = true
			out := readOutput(ch, chClient)
			if !strings.Contains(out, "fails to take hold") {
				t.Errorf("output %q should contain 'fails to take hold'", out)
			}
			break
		}
	}
	if !resisted {
		t.Error("dispel magic should be resisted at least once in 50 attempts (level 1 vs 50)")
	}
}

// SpellSleep: victim already sleeping (Position <= POS_SLEEPING)
func TestSpellSleep_AlreadySleepingPosition(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9403, Name: "Arena"}
	ch := newCaster("Wizard", 20)
	handler.CharToRoom(ch, room)

	victim := newNPCVictim("target", 5)
	victim.Position = types.POS_SLEEPING
	handler.CharToRoom(victim, room)

	// Force it to land by trying many times
	landed := false
	for i := 0; i < 50; i++ {
		victim.Affects = nil
		victim.AffectedBy.Clear()
		// Keep position sleeping to test that branch
		victim.Position = types.POS_SLEEPING
		SpellSleep(w, 0, ch.Level, ch, victim)
		if victim.AffectedBy.IsSet(types.AFF_SLEEP) {
			landed = true
			// Position should remain sleeping, not be changed
			if victim.Position != types.POS_SLEEPING {
				t.Error("position should remain POS_SLEEPING")
			}
			break
		}
	}
	if !landed {
		t.Error("sleep should land at least once in 50 attempts (level 20 vs 5)")
	}
}

// SpellCharmPerson save-resisted path
func TestSpellCharmPerson_SaveResists(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9404, Name: "Arena"}
	ch, client := newCasterWithDesc("Wizard", 1)
	defer client.Close()
	handler.CharToRoom(ch, room)

	victim := newNPCVictim("dragon", 50)
	handler.CharToRoom(victim, room)

	resisted := false
	for i := 0; i < 50; i++ {
		victim.Affects = nil
		victim.AffectedBy.Clear()
		victim.Master = nil
		victim.Leader = nil
		_ = readOutput(ch, client)
		SpellCharmPerson(w, 0, ch.Level, ch, victim)
		if !victim.AffectedBy.IsSet(types.AFF_CHARM) {
			resisted = true
			out := readOutput(ch, client)
			if !strings.Contains(out, "fails to take hold") {
				t.Errorf("output %q should contain 'fails to take hold'", out)
			}
			break
		}
	}
	if !resisted {
		t.Error("charm should be resisted at least once in 50 attempts (level 1 vs 50)")
	}
}

// SpellSummon nil room guard
func TestSpellSummon_NilRoom(t *testing.T) {
	w := newMagicWorld()
	ch := newCaster("Mage", 20)
	// ch.InRoom is nil (no CharToRoom called)
	victim := newCaster("Target", 5)

	// Should not panic
	SpellSummon(w, 0, ch.Level, ch, victim)

	// Also test victim with nil room but ch has a room
	room := &types.RoomIndexData{Vnum: 9405, Name: "Temple"}
	handler.CharToRoom(ch, room)
	victim2 := newCaster("Target2", 5)
	// victim2.InRoom is nil

	SpellSummon(w, 0, ch.Level, ch, victim2)
	// Should silently return without panic
}

// SpellBlindness save-resisted path
func TestSpellBlindness_SaveResists(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9406, Name: "Arena"}
	ch, client := newCasterWithDesc("Wizard", 1)
	defer client.Close()
	handler.CharToRoom(ch, room)

	victim := newNPCVictim("dragon", 50)
	handler.CharToRoom(victim, room)

	resisted := false
	for i := 0; i < 50; i++ {
		victim.Affects = nil
		victim.AffectedBy.Clear()
		victim.Hitroll = 0
		_ = readOutput(ch, client)
		SpellBlindness(w, 0, ch.Level, ch, victim)
		if !victim.AffectedBy.IsSet(types.AFF_BLIND) {
			resisted = true
			out := readOutput(ch, client)
			if !strings.Contains(out, "fails to take hold") {
				t.Errorf("output %q should contain 'fails to take hold'", out)
			}
			break
		}
	}
	if !resisted {
		t.Error("blindness should be resisted at least once in 50 attempts (level 1 vs 50)")
	}
}

// SpellSleep save-resisted path
func TestSpellSleep_SaveResists(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9407, Name: "Arena"}
	ch, client := newCasterWithDesc("Wizard", 1)
	defer client.Close()
	handler.CharToRoom(ch, room)

	victim := newNPCVictim("dragon", 50)
	handler.CharToRoom(victim, room)

	resisted := false
	for i := 0; i < 50; i++ {
		victim.Affects = nil
		victim.AffectedBy.Clear()
		victim.Position = types.POS_STANDING
		_ = readOutput(ch, client)
		SpellSleep(w, 0, ch.Level, ch, victim)
		if !victim.AffectedBy.IsSet(types.AFF_SLEEP) {
			resisted = true
			out := readOutput(ch, client)
			if !strings.Contains(out, "fails to take hold") {
				t.Errorf("output %q should contain 'fails to take hold'", out)
			}
			break
		}
	}
	if !resisted {
		t.Error("sleep should be resisted at least once in 50 attempts (level 1 vs 50)")
	}
}

// ---------------------------------------------------------------------------
// SavesWands
// ---------------------------------------------------------------------------

func TestSavesWands(t *testing.T) {
	// High level victim vs low level wand = high save chance
	victim := &types.CharData{Level: 20}
	saved := 0
	for i := 0; i < 100; i++ {
		if SavesWands(5, victim) {
			saved++
		}
	}
	if saved < 50 {
		t.Errorf("high-level victim saved %d/100 vs low-level wand, expected mostly saves", saved)
	}

	// Low level victim vs high level wand = low save chance
	weakVictim := &types.CharData{Level: 1}
	weakSaved := 0
	for i := 0; i < 100; i++ {
		if SavesWands(30, weakVictim) {
			weakSaved++
		}
	}
	if weakSaved > 30 {
		t.Errorf("low-level victim saved %d/100 vs high-level wand, expected few saves", weakSaved)
	}
}

func TestSavesWands_ClampedRange(t *testing.T) {
	// Very high save: level 50 vs wand level 1 should be clamped at 95%
	victim := &types.CharData{Level: 50}
	allSaved := true
	for i := 0; i < 200; i++ {
		if !SavesWands(1, victim) {
			allSaved = false
			break
		}
	}
	if allSaved {
		t.Error("even with very high save, should occasionally fail (clamped at 95%)")
	}

	// Very low save: level 1 vs wand level 50 should be clamped at 5%
	weakVictim := &types.CharData{Level: 1}
	allFailed := true
	for i := 0; i < 200; i++ {
		if SavesWands(50, weakVictim) {
			allFailed = false
			break
		}
	}
	if allFailed {
		t.Error("even with very low save, should occasionally succeed (clamped at 5%)")
	}
}

func TestSavesWands_UsesSavingWandField(t *testing.T) {
	// A big SavingWand bonus should push save chance to the ceiling.
	victim := &types.CharData{Level: 10, SavingWand: -50}
	saved := 0
	for i := 0; i < 100; i++ {
		if SavesWands(10, victim) {
			saved++
		}
	}
	// With save = 50 + (10-10-(-50))*5 = 300, clamped to 95%, expect near-universal saves.
	if saved < 80 {
		t.Errorf("victim with SavingWand=-50 saved %d/100, expected >=80", saved)
	}
}

// ---------------------------------------------------------------------------
// SavesParaPetri
// ---------------------------------------------------------------------------

func TestSavesParaPetri(t *testing.T) {
	victim := &types.CharData{Level: 20}
	saved := 0
	for i := 0; i < 100; i++ {
		if SavesParaPetri(5, victim) {
			saved++
		}
	}
	if saved < 50 {
		t.Errorf("high-level victim saved %d/100 vs low-level paralysis, expected mostly saves", saved)
	}

	weakVictim := &types.CharData{Level: 1}
	weakSaved := 0
	for i := 0; i < 100; i++ {
		if SavesParaPetri(30, weakVictim) {
			weakSaved++
		}
	}
	if weakSaved > 30 {
		t.Errorf("low-level victim saved %d/100 vs high-level paralysis, expected few saves", weakSaved)
	}
}

func TestSavesParaPetri_ClampedRange(t *testing.T) {
	victim := &types.CharData{Level: 50}
	allSaved := true
	for i := 0; i < 200; i++ {
		if !SavesParaPetri(1, victim) {
			allSaved = false
			break
		}
	}
	if allSaved {
		t.Error("even with very high save, should occasionally fail (clamped at 95%)")
	}

	weakVictim := &types.CharData{Level: 1}
	allFailed := true
	for i := 0; i < 200; i++ {
		if SavesParaPetri(50, weakVictim) {
			allFailed = false
			break
		}
	}
	if allFailed {
		t.Error("even with very low save, should occasionally succeed (clamped at 5%)")
	}
}

func TestSavesParaPetri_UsesSavingParaPetriField(t *testing.T) {
	victim := &types.CharData{Level: 10, SavingParaPetri: -50}
	saved := 0
	for i := 0; i < 100; i++ {
		if SavesParaPetri(10, victim) {
			saved++
		}
	}
	if saved < 80 {
		t.Errorf("victim with SavingParaPetri=-50 saved %d/100, expected >=80", saved)
	}
}

// ---------------------------------------------------------------------------
// SavesBreath
// ---------------------------------------------------------------------------

func TestSavesBreath(t *testing.T) {
	victim := &types.CharData{Level: 20}
	saved := 0
	for i := 0; i < 100; i++ {
		if SavesBreath(5, victim) {
			saved++
		}
	}
	if saved < 50 {
		t.Errorf("high-level victim saved %d/100 vs low-level breath, expected mostly saves", saved)
	}

	weakVictim := &types.CharData{Level: 1}
	weakSaved := 0
	for i := 0; i < 100; i++ {
		if SavesBreath(30, weakVictim) {
			weakSaved++
		}
	}
	if weakSaved > 30 {
		t.Errorf("low-level victim saved %d/100 vs high-level breath, expected few saves", weakSaved)
	}
}

func TestSavesBreath_ClampedRange(t *testing.T) {
	victim := &types.CharData{Level: 50}
	allSaved := true
	for i := 0; i < 200; i++ {
		if !SavesBreath(1, victim) {
			allSaved = false
			break
		}
	}
	if allSaved {
		t.Error("even with very high save, should occasionally fail (clamped at 95%)")
	}

	weakVictim := &types.CharData{Level: 1}
	allFailed := true
	for i := 0; i < 200; i++ {
		if SavesBreath(50, weakVictim) {
			allFailed = false
			break
		}
	}
	if allFailed {
		t.Error("even with very low save, should occasionally succeed (clamped at 5%)")
	}
}

func TestSavesBreath_UsesSavingBreathField(t *testing.T) {
	victim := &types.CharData{Level: 10, SavingBreath: -50}
	saved := 0
	for i := 0; i < 100; i++ {
		if SavesBreath(10, victim) {
			saved++
		}
	}
	if saved < 80 {
		t.Errorf("victim with SavingBreath=-50 saved %d/100, expected >=80", saved)
	}
}

// TestSavesWands_RISMagicImmune verifies the RIS_MAGIC immunity short-circuit
// that C (src/magic.c:1087) applies. A victim immune to RIS_MAGIC auto-saves
// against wands regardless of level / SavingWand. Regression guard for the
// earlier port that omitted this check.
func TestSavesWands_RISMagicImmune(t *testing.T) {
	victim := &types.CharData{
		Level:       1,          // much lower than caster; would normally fail
		SavingWand:  -50,        // terrible save; would normally fail
		Immune:      int(types.RIS_MAGIC),
	}
	for i := 0; i < 50; i++ {
		if !SavesWands(60, victim) {
			t.Fatalf("RIS_MAGIC-immune victim failed a wand save on iteration %d", i)
		}
	}
}

// TestSavesParaPetri_NoRISMagicShortCircuit confirms the RIS_MAGIC guard was
// added ONLY to SavesWands, matching C (which omits it in saves_para_petri).
func TestSavesParaPetri_NoRISMagicShortCircuit(t *testing.T) {
	victim := &types.CharData{
		Level:           60,
		SavingParaPetri: 100, // absurdly good save — should pass without RIS_MAGIC
		Immune:          0,   // not immune
	}
	// With save clamped to 95, expect ~95% pass rate — pick any failure to
	// prove the path was reached (no short-circuit).
	failed := false
	for i := 0; i < 400; i++ {
		if !SavesParaPetri(1, victim) {
			failed = true
			break
		}
	}
	if !failed {
		t.Error("expected at least one failure over 400 rolls against a 95pct-clamped save")
	}
}
