// Tests for Phase 5 Tier 4 G3 Group D — unique-mechanic spells.
package magic

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

func TestSpellAcidBlast_DealsDamage(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9700, Name: "Arena"}
	ch := newCaster("Mage", 30)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)
	victim := &types.CharData{
		Name: "target", Level: 5, Hit: 500, MaxHit: 500,
		Position: types.POS_STANDING, Armor: 100,
	}
	victim.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	SpellAcidBlast(w, 0, ch.Level, ch, victim)

	if victim.Hit >= 500 {
		t.Errorf("acid_blast should damage victim (Hit=%d)", victim.Hit)
	}
}

func TestSpellKnock_OpensLockedDoor(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9710, Name: "Corridor"}
	exit := &types.ExitData{
		Direction: types.DIR_NORTH,
		Keyword:   "door",
		ExitInfo:  int(types.EX_ISDOOR) | int(types.EX_CLOSED) | int(types.EX_LOCKED),
	}
	room.Exits = append(room.Exits, exit)
	ch := newCaster("Thief", 20)
	handler.CharToRoom(ch, room)

	SpellKnock(w, 0, ch.Level, ch, ch)

	if exit.ExitInfo&int(types.EX_LOCKED) != 0 {
		t.Error("knock should unlock the door")
	}
}

func TestSpellKnock_PickproofBlocked(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9711, Name: "Corridor"}
	exit := &types.ExitData{
		Direction: types.DIR_NORTH,
		Keyword:   "door",
		ExitInfo:  int(types.EX_ISDOOR) | int(types.EX_CLOSED) | int(types.EX_LOCKED) | int(types.EX_PICKPROOF),
	}
	room.Exits = append(room.Exits, exit)
	ch, client := newCasterWithDesc("Thief", 20)
	handler.CharToRoom(ch, room)

	SpellKnock(w, 0, ch.Level, ch, ch)

	if exit.ExitInfo&int(types.EX_LOCKED) == 0 {
		t.Error("pickproof door should remain locked")
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "failed") {
		t.Errorf("expected failure message, got %q", out)
	}
}

func TestSpellRecharge_NoWandInInventory(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9720, Name: "Lab"}
	ch, client := newCasterWithDesc("Mage", 30)
	handler.CharToRoom(ch, room)

	SpellRecharge(w, 0, ch.Level, ch, ch)

	out := readOutput(ch, client)
	if !strings.Contains(out, "nothing") {
		t.Errorf("expected 'nothing to recharge', got %q", out)
	}
}

func TestSpellRecharge_OverchargedBurns(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9721, Name: "Lab"}
	ch, client := newCasterWithDesc("Mage", 30)
	handler.CharToRoom(ch, room)
	wand := &types.ObjData{
		Name:       "wand",
		ShortDescr: "a crackling wand",
		ItemType:   types.ITEM_WAND,
		Value:      [6]int{0, 5, 5, 0, 0, 0}, // value[1]==value[2] → burn path
	}
	handler.ObjToChar(wand, ch)

	SpellRecharge(w, 0, ch.Level, ch, ch)
	out := readOutput(ch, client)
	if !strings.Contains(out, "flames") {
		t.Errorf("fully charged wand should burst into flames, got %q", out)
	}
}

func TestSpellAnimateDead_CreatesFollower(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9730, Name: "Crypt"}
	// Register the animated-corpse mob index.
	corpseIdx := &types.MobIndexData{
		Vnum:       types.MOB_VNUM_ANIMATED_CORPSE,
		PlayerName: "animated corpse",
		ShortDescr: "an animated corpse",
	}
	w.MobIndex[types.MOB_VNUM_ANIMATED_CORPSE] = corpseIdx
	ch := newCaster("Necromancer", 30)
	handler.CharToRoom(ch, room)

	corpse := &types.ObjData{
		Name:       "corpse",
		ShortDescr: "the corpse of a rat",
		ItemType:   types.ITEM_CORPSE_NPC,
	}
	room.Contents = append(room.Contents, corpse)
	corpse.InRoom = room

	SpellAnimateDead(w, 0, ch.Level, ch, ch)

	// Look for the animated corpse mob in the room.
	foundMob := false
	for _, p := range room.People {
		if p == ch {
			continue
		}
		if p.Master == ch && p.AffectedBy.IsSet(types.AFF_CHARM) {
			foundMob = true
		}
	}
	if !foundMob {
		t.Error("animated corpse should be a charmed follower in the room")
	}
}

func TestSpellAnimateDead_NoCorpse(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9731, Name: "Plaza"}
	ch, client := newCasterWithDesc("Necromancer", 30)
	handler.CharToRoom(ch, room)

	SpellAnimateDead(w, 0, ch.Level, ch, ch)
	out := readOutput(ch, client)
	if !strings.Contains(out, "cannot find") {
		t.Errorf("expected 'cannot find' message, got %q", out)
	}
}

func TestSpellEnergyDrain_DamagesAndHeals(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9740, Name: "Arena"}
	ch := newCaster("Necromancer", 30)
	ch.Hit = 50
	ch.MaxHit = 200
	handler.CharToRoom(ch, room)
	w.AddChar(ch)
	victim := &types.CharData{
		Name: "target", Level: 10, Hit: 500, MaxHit: 500, MaxMana: 100, Mana: 100, MaxMove: 100, Move: 100,
		Position: types.POS_STANDING, Armor: 100, Exp: 10000,
	}
	victim.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	// Run a few times because SavesSpellStaff can cancel the spell.
	landed := false
	for i := 0; i < 20; i++ {
		victim.Hit = 500
		victim.Mana = 100
		victim.Move = 100
		ch.Hit = 50
		SpellEnergyDrain(w, 0, ch.Level, ch, victim)
		if victim.Hit < 500 && victim.Mana < 100 && victim.Move < 100 {
			landed = true
			break
		}
	}
	if !landed {
		t.Error("energy_drain should damage + drain mana/move at least once in 20 attempts")
	}
}

func TestSpellEnergyDrain_MagicImmune(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9741, Name: "Arena"}
	ch, client := newCasterWithDesc("Necromancer", 30)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)
	victim := &types.CharData{Name: "golem", Level: 10, Hit: 500, MaxHit: 500, Immune: int(types.RIS_MAGIC)}
	victim.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(victim, room)

	SpellEnergyDrain(w, 0, ch.Level, ch, victim)
	out := readOutput(ch, client)
	if !strings.Contains(out, "immune") {
		t.Errorf("expected immunity message, got %q", out)
	}
	if victim.Hit != 500 {
		t.Error("magic-immune victim should take no damage")
	}
}

func TestSpellCallLightning_IndoorsFails(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9750, Name: "Hall"}
	room.RoomFlags.Set(types.ROOM_INDOORS)
	ch, client := newCasterWithDesc("Druid", 30)
	handler.CharToRoom(ch, room)
	w.WeatherInfo.Sky = types.SKY_RAINING

	SpellCallLightning(w, 0, ch.Level, ch, ch)

	out := readOutput(ch, client)
	if !strings.Contains(out, "out of doors") {
		t.Errorf("expected 'out of doors' message, got %q", out)
	}
}

func TestSpellCallLightning_NoBadWeatherFails(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9751, Name: "Field", SectorType: types.SECT_FIELD}
	ch, client := newCasterWithDesc("Druid", 30)
	handler.CharToRoom(ch, room)
	w.WeatherInfo.Sky = types.SKY_CLOUDLESS

	SpellCallLightning(w, 0, ch.Level, ch, ch)
	out := readOutput(ch, client)
	if !strings.Contains(out, "bad weather") {
		t.Errorf("expected 'bad weather' message, got %q", out)
	}
}

func TestSpellCallLightning_OutdoorsAndRainingHits(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9752, Name: "Field", SectorType: types.SECT_FIELD}
	ch := breathCasterNPC(30)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)
	w.WeatherInfo.Sky = types.SKY_RAINING

	victim := &types.CharData{Name: "target", Level: 5, Hit: 500, MaxHit: 500, Position: types.POS_STANDING, Armor: 100}
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	SpellCallLightning(w, 0, ch.Level, ch, victim)

	if victim.Hit >= 500 {
		t.Errorf("call_lightning should damage victim (Hit=%d)", victim.Hit)
	}
}

func TestSpellControlWeather_ModifiesWeather(t *testing.T) {
	w := newMagicWorld()
	weather := &types.WeatherData{}
	area := &types.AreaData{Name: "Test Area", Weather: weather}
	room := &types.RoomIndexData{Vnum: 9760, Name: "Field", Area: area}
	ch, client := newCasterWithDesc("Druid", 30)
	handler.CharToRoom(ch, room)

	SpellControlWeather(w, 0, ch.Level, ch, ch)
	out := readOutput(ch, client)
	// Some vector should have changed.
	changed := weather.TempVector != 0 || weather.PrecipVector != 0 || weather.WindVector != 0
	if !changed {
		t.Errorf("control_weather should nudge at least one vector, got %+v, out=%q", weather, out)
	}
	if !strings.Contains(out, "weather grows") {
		t.Errorf("expected 'weather grows' message, got %q", out)
	}
}

func TestSpellControlWeather_NoWeatherData(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9761, Name: "Void"}
	ch, client := newCasterWithDesc("Druid", 30)
	handler.CharToRoom(ch, room)

	SpellControlWeather(w, 0, ch.Level, ch, ch)
	out := readOutput(ch, client)
	if !strings.Contains(out, "does not respond") {
		t.Errorf("expected 'does not respond' message, got %q", out)
	}
}

func TestSpellRegistry_GroupD(t *testing.T) {
	for _, name := range []string{
		"spell_acid_blast", "spell_knock", "spell_recharge",
		"spell_animate_dead", "spell_energy_drain",
		"spell_call_lightning", "spell_control_weather",
	} {
		if FindSpellFunc(name) == nil {
			t.Errorf("Group D spell %q should be registered", name)
		}
	}
}
