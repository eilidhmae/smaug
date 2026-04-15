// Tests for Phase 5 Tier 4 G3 Group A — simple affect-variant spells.
package magic

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

func TestSpellPassDoor_AddsAffect(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9400, Name: "Temple"}
	ch := newCaster("Wizard", 20)
	handler.CharToRoom(ch, room)

	SpellPassDoor(w, 0, ch.Level, ch, ch)

	if !ch.AffectedBy.IsSet(types.AFF_PASS_DOOR) {
		t.Error("AFF_PASS_DOOR should be set after pass_door")
	}
}

func TestSpellPassDoor_AlreadyAffected(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9401, Name: "Temple"}
	ch, client := newCasterWithDesc("Wizard", 20)
	handler.CharToRoom(ch, room)
	ch.AffectedBy.Set(types.AFF_PASS_DOOR)
	affectsBefore := len(ch.Affects)

	SpellPassDoor(w, 0, ch.Level, ch, ch)

	out := readOutput(ch, client)
	if !strings.Contains(out, "already") {
		t.Errorf("expected 'already' in output, got %q", out)
	}
	if len(ch.Affects) != affectsBefore {
		t.Error("PassDoor should not stack when already affected")
	}
}

func TestSpellPassDoor_MagicImmune(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9402, Name: "Temple"}
	ch, client := newCasterWithDesc("Wizard", 20)
	handler.CharToRoom(ch, room)
	victim := &types.CharData{Name: "golem", Level: 10, Immune: int(types.RIS_MAGIC)}
	victim.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(victim, room)

	SpellPassDoor(w, 0, ch.Level, ch, victim)

	out := readOutput(ch, client)
	if !strings.Contains(out, "immune") {
		t.Errorf("expected immunity message, got %q", out)
	}
	if victim.AffectedBy.IsSet(types.AFF_PASS_DOOR) {
		t.Error("magic-immune victim should not gain AFF_PASS_DOOR")
	}
}

func TestSpellFarsight_Success(t *testing.T) {
	w := newMagicWorld()
	casterRoom := &types.RoomIndexData{Vnum: 9410, Name: "Caster's Room"}
	targetRoom := &types.RoomIndexData{Vnum: 9411, Name: "Dragon's Lair", Description: "A dark cavern."}
	ch, client := newCasterWithDesc("Scryer", 30)
	handler.CharToRoom(ch, casterRoom)
	victim := &types.CharData{Name: "dragon", Level: 10}
	victim.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(victim, targetRoom)

	SpellFarsight(w, 0, ch.Level, ch, victim)

	out := readOutput(ch, client)
	if !strings.Contains(out, "Dragon's Lair") {
		t.Errorf("expected target room name in output, got %q", out)
	}
}

func TestSpellFarsight_BlockedByNoAstral(t *testing.T) {
	w := newMagicWorld()
	casterRoom := &types.RoomIndexData{Vnum: 9412, Name: "Caster's Room"}
	targetRoom := &types.RoomIndexData{Vnum: 9413, Name: "Warded Sanctum"}
	targetRoom.RoomFlags.Set(types.ROOM_NO_ASTRAL)
	ch, client := newCasterWithDesc("Scryer", 30)
	handler.CharToRoom(ch, casterRoom)
	victim := &types.CharData{Name: "priest", Level: 10}
	handler.CharToRoom(victim, targetRoom)

	SpellFarsight(w, 0, ch.Level, ch, victim)

	out := readOutput(ch, client)
	if strings.Contains(out, "Warded Sanctum") {
		t.Errorf("NO_ASTRAL room should block farsight, got %q", out)
	}
	if !strings.Contains(out, "fail") {
		t.Errorf("expected failure message, got %q", out)
	}
}

func TestSpellVentriloquate_ListenersReceive(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9420, Name: "Tavern"}
	ch := newCaster("Trickster", 30)
	handler.CharToRoom(ch, room)
	victim := &types.CharData{Name: "bob", ShortDescr: "bob"}
	victim.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(victim, room)
	listener, client := newCasterWithDesc("listener", 5)
	handler.CharToRoom(listener, room)

	SpellVentriloquate(w, 0, ch.Level, ch, victim)

	out := readOutput(listener, client)
	if !strings.Contains(out, "bob") && !strings.Contains(out, "Bob") {
		t.Errorf("listener output %q should mention bob", out)
	}
}

func TestSpellRemoveInvis_StripsInvis(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9430, Name: "Plaza"}
	ch := newCaster("Mage", 20)
	handler.CharToRoom(ch, room)
	victim := &types.CharData{Name: "sneak", Level: 5}
	victim.AffectedBy.Set(types.AFF_INVISIBLE)
	handler.CharToRoom(victim, room)

	SpellRemoveInvis(w, 0, ch.Level, ch, victim)

	if victim.AffectedBy.IsSet(types.AFF_INVISIBLE) {
		t.Error("remove_invis should strip AFF_INVISIBLE")
	}
}

func TestSpellRemoveInvis_NotInvisible(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9431, Name: "Plaza"}
	ch, client := newCasterWithDesc("Mage", 20)
	handler.CharToRoom(ch, room)
	victim := &types.CharData{Name: "normal", Level: 5}
	handler.CharToRoom(victim, room)

	SpellRemoveInvis(w, 0, ch.Level, ch, victim)

	out := readOutput(ch, client)
	if !strings.Contains(out, "not invisible") {
		t.Errorf("expected 'not invisible' message, got %q", out)
	}
}

func TestSpellRemoveTrap_Success(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9440, Name: "Trapped Room"}
	ch, client := newCasterWithDesc("Rogue", 50) // high level = high success
	handler.CharToRoom(ch, room)
	trap := &types.ObjData{
		Name:       "trap",
		ShortDescr: "a spike trap",
		ItemType:   types.ITEM_TRAP,
	}
	room.Contents = append(room.Contents, trap)
	trap.InRoom = room

	// Retry a few times — chance is 75+level/4 ≈ 87% at level 50, non-deterministic.
	var outcome string
	landed := false
	for i := 0; i < 20; i++ {
		SpellRemoveTrap(w, 0, ch.Level, ch, ch)
		outcome = readOutput(ch, client)
		if strings.Contains(outcome, "successfully") {
			landed = true
			break
		}
		// Restore trap between attempts
		if len(room.Contents) == 0 {
			room.Contents = append(room.Contents, trap)
			trap.InRoom = room
		}
	}
	if !landed {
		t.Errorf("remove_trap should land at level 50 within 20 attempts, last out=%q", outcome)
	}
}

func TestSpellRemoveTrap_NoTrap(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9441, Name: "Empty Room"}
	ch, client := newCasterWithDesc("Rogue", 20)
	handler.CharToRoom(ch, room)

	SpellRemoveTrap(w, 0, ch.Level, ch, ch)
	out := readOutput(ch, client)
	if !strings.Contains(out, "can't find") && !strings.Contains(out, "cannot find") {
		t.Errorf("expected 'can't find' message, got %q", out)
	}
}

func TestSpellRegistry_GroupA(t *testing.T) {
	for _, name := range []string{
		"spell_pass_door", "spell_farsight", "spell_ventriloquate",
		"spell_remove_invis", "spell_remove_trap",
	} {
		if FindSpellFunc(name) == nil {
			t.Errorf("Group A spell %q should be registered", name)
		}
	}
}
