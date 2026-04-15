// Tests for Phase 5 Tier 4 G3 Group B — breath / area attacks.
package magic

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

// breathCasterNPC builds an NPC caster with enough HP for meaningful breath dice.
func breathCasterNPC(level int) *types.CharData {
	c := &types.CharData{
		Name: "dragon", Level: level, Hit: 200, MaxHit: 200, Position: types.POS_STANDING,
	}
	c.Act.Set(types.ACT_IS_NPC)
	return c
}

// TestBreathSpells runs each of the six Group B spells and verifies they
// damage a PC victim in the caster's room (table-driven).
func TestBreathSpells(t *testing.T) {
	spells := []struct {
		name string
		fn   SpellFunc
	}{
		{"acid_breath", SpellAcidBreath},
		{"fire_breath", SpellFireBreath},
		{"frost_breath", SpellFrostBreath},
		{"gas_breath", SpellGasBreath},
		{"lightning_breath", SpellLightningBreath},
		{"earthquake", SpellEarthquake},
	}
	for _, sp := range spells {
		t.Run(sp.name, func(t *testing.T) {
			w := newMagicWorld()
			room := &types.RoomIndexData{Vnum: 9500, Name: "Lair"}
			ch := breathCasterNPC(30)
			handler.CharToRoom(ch, room)
			w.AddChar(ch)
			victim := &types.CharData{
				Name: "hero", Level: 10, Hit: 500, MaxHit: 500,
				Position: types.POS_STANDING, Armor: 100,
			}
			handler.CharToRoom(victim, room)
			w.AddChar(victim)

			sp.fn(w, 0, ch.Level, ch, victim)

			if victim.Hit >= 500 {
				t.Errorf("%s did not damage victim (Hit=%d)", sp.name, victim.Hit)
			}
		})
	}
}

func TestSpellEarthquake_SkipsFlying(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9510, Name: "Plaza"}
	ch := breathCasterNPC(30)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	flyer := &types.CharData{
		Name: "airwalker", Level: 10, Hit: 500, MaxHit: 500,
		Position: types.POS_STANDING, Armor: 100,
	}
	flyer.AffectedBy.Set(types.AFF_FLYING)
	handler.CharToRoom(flyer, room)
	w.AddChar(flyer)

	ground := &types.CharData{
		Name: "walker", Level: 10, Hit: 500, MaxHit: 500,
		Position: types.POS_STANDING, Armor: 100,
	}
	handler.CharToRoom(ground, room)
	w.AddChar(ground)

	SpellEarthquake(w, 0, ch.Level, ch, ground)

	if flyer.Hit != 500 {
		t.Errorf("flying target should be skipped (Hit=%d)", flyer.Hit)
	}
	if ground.Hit >= 500 {
		t.Errorf("ground target should take damage (Hit=%d)", ground.Hit)
	}
}

func TestSpellEarthquake_RoomSafeBlocks(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9511, Name: "Sanctuary"}
	room.RoomFlags.Set(types.ROOM_SAFE)
	ch, client := newCasterWithDesc("Priest", 30)
	handler.CharToRoom(ch, room)
	target := &types.CharData{Name: "target", Level: 5, Hit: 100, MaxHit: 100}
	target.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(target, room)

	SpellEarthquake(w, 0, ch.Level, ch, target)

	out := readOutput(ch, client)
	if !strings.Contains(out, "refuses") {
		t.Errorf("expected 'refuses' message in ROOM_SAFE, got %q", out)
	}
	if target.Hit != 100 {
		t.Errorf("target should be untouched in ROOM_SAFE (Hit=%d)", target.Hit)
	}
}

func TestSpellGasBreath_RoomSafeBlocks(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9512, Name: "Safe"}
	room.RoomFlags.Set(types.ROOM_SAFE)
	ch := breathCasterNPC(30)
	handler.CharToRoom(ch, room)
	target := &types.CharData{Name: "target", Level: 5, Hit: 100, MaxHit: 100}
	handler.CharToRoom(target, room)

	SpellGasBreath(w, 0, ch.Level, ch, target)

	if target.Hit != 100 {
		t.Errorf("ROOM_SAFE should block gas_breath (Hit=%d)", target.Hit)
	}
}

func TestSpellRegistry_GroupB(t *testing.T) {
	for _, name := range []string{
		"spell_acid_breath", "spell_fire_breath", "spell_frost_breath",
		"spell_gas_breath", "spell_lightning_breath", "spell_earthquake",
	} {
		if FindSpellFunc(name) == nil {
			t.Errorf("Group B spell %q should be registered", name)
		}
	}
}
