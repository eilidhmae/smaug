// Tests for Phase 5 Tier 4 G3 Group C — teleport spells.
package magic

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

func TestSpellAstralWalk_Success(t *testing.T) {
	w := newMagicWorld()
	src := &types.RoomIndexData{Vnum: 9600, Name: "Source"}
	dst := &types.RoomIndexData{Vnum: 9601, Name: "Destination"}
	ch := newCaster("Traveler", 30)
	handler.CharToRoom(ch, src)
	victim := &types.CharData{Name: "friend", Level: 5}
	handler.CharToRoom(victim, dst)

	SpellAstralWalk(w, 0, ch.Level, ch, victim)

	if ch.InRoom != dst {
		t.Errorf("caster should be in dest room, in %v", ch.InRoom)
	}
}

func TestSpellAstralWalk_NoAstralBlocks(t *testing.T) {
	w := newMagicWorld()
	src := &types.RoomIndexData{Vnum: 9602, Name: "Source"}
	dst := &types.RoomIndexData{Vnum: 9603, Name: "Warded"}
	dst.RoomFlags.Set(types.ROOM_NO_ASTRAL)
	ch := newCaster("Traveler", 30)
	handler.CharToRoom(ch, src)
	victim := &types.CharData{Name: "friend", Level: 5}
	handler.CharToRoom(victim, dst)

	SpellAstralWalk(w, 0, ch.Level, ch, victim)

	if ch.InRoom != src {
		t.Error("caster should not have moved (NO_ASTRAL)")
	}
}

func TestSpellAstralWalk_NoRecallOnSelfBlocks(t *testing.T) {
	w := newMagicWorld()
	src := &types.RoomIndexData{Vnum: 9604, Name: "Sealed"}
	src.RoomFlags.Set(types.ROOM_NO_RECALL)
	dst := &types.RoomIndexData{Vnum: 9605, Name: "Destination"}
	ch := newCaster("Traveler", 30)
	handler.CharToRoom(ch, src)
	victim := &types.CharData{Name: "friend", Level: 5}
	handler.CharToRoom(victim, dst)

	SpellAstralWalk(w, 0, ch.Level, ch, victim)

	if ch.InRoom != src {
		t.Error("caster should not have moved (NO_RECALL on source)")
	}
}

func TestSpellGate_Success(t *testing.T) {
	w := newMagicWorld()
	src := &types.RoomIndexData{Vnum: 9610, Name: "Source"}
	dst := &types.RoomIndexData{Vnum: 9611, Name: "Destination"}
	ch := newCaster("Mage", 30)
	handler.CharToRoom(ch, src)
	victim := &types.CharData{Name: "friend", Level: 5}
	handler.CharToRoom(victim, dst)

	SpellGate(w, 0, ch.Level, ch, victim)

	if ch.InRoom != dst {
		t.Error("gate should teleport caster to victim's room")
	}
}

func TestSpellGate_NoSelfTarget(t *testing.T) {
	w := newMagicWorld()
	src := &types.RoomIndexData{Vnum: 9612, Name: "Source"}
	ch := newCaster("Mage", 30)
	handler.CharToRoom(ch, src)

	SpellGate(w, 0, ch.Level, ch, ch)

	if ch.InRoom != src {
		t.Error("self-gate should be a no-op")
	}
}

func TestSpellMistWalk_Success(t *testing.T) {
	w := newMagicWorld()
	src := &types.RoomIndexData{Vnum: 9620, Name: "Source"}
	dst := &types.RoomIndexData{Vnum: 9621, Name: "Destination"}
	ch := newCaster("Vampire", 30)
	handler.CharToRoom(ch, src)
	victim := &types.CharData{Name: "prey", Level: 5}
	handler.CharToRoom(victim, dst)

	SpellMistWalk(w, 0, ch.Level, ch, victim)

	if ch.InRoom != dst {
		t.Error("mist_walk should teleport caster to victim's room")
	}
}

func TestSpellTransport_SameRoom(t *testing.T) {
	w := newMagicWorld()
	room := &types.RoomIndexData{Vnum: 9630, Name: "Room"}
	ch, client := newCasterWithDesc("Mage", 30)
	handler.CharToRoom(ch, room)
	victim := &types.CharData{Name: "friend", Level: 5}
	handler.CharToRoom(victim, room)

	SpellTransport(w, 0, ch.Level, ch, victim)
	readOutput(ch, client) // drain; success path is "no-op, already here"
	// Caster should not have moved anywhere meaningful.
	if ch.InRoom != room {
		t.Error("caster should remain in original room when target is already present")
	}
}

func TestSpellTransport_Success(t *testing.T) {
	w := newMagicWorld()
	src := &types.RoomIndexData{Vnum: 9631, Name: "Source"}
	dst := &types.RoomIndexData{Vnum: 9632, Name: "Destination"}
	ch := newCaster("Mage", 30)
	handler.CharToRoom(ch, src)
	victim := &types.CharData{Name: "friend", Level: 5}
	handler.CharToRoom(victim, dst)

	SpellTransport(w, 0, ch.Level, ch, victim)

	if ch.InRoom != dst {
		t.Error("transport should teleport caster to victim's room")
	}
}

func TestSpellWordOfRecall_MovesToTemple(t *testing.T) {
	w := newMagicWorld()
	src := &types.RoomIndexData{Vnum: 9640, Name: "Remote"}
	temple := &types.RoomIndexData{Vnum: types.ROOM_VNUM_TEMPLE, Name: "Temple"}
	w.Rooms[src.Vnum] = src
	w.Rooms[temple.Vnum] = temple
	ch := newCaster("Caster", 20)
	handler.CharToRoom(ch, src)

	SpellWordOfRecall(w, 0, ch.Level, ch, ch)

	if ch.InRoom != temple {
		t.Errorf("caster should be in temple, got %v", ch.InRoom)
	}
}

func TestSpellWordOfRecall_NoRecallBlocks(t *testing.T) {
	w := newMagicWorld()
	src := &types.RoomIndexData{Vnum: 9641, Name: "Sealed"}
	src.RoomFlags.Set(types.ROOM_NO_RECALL)
	temple := &types.RoomIndexData{Vnum: types.ROOM_VNUM_TEMPLE, Name: "Temple"}
	w.Rooms[src.Vnum] = src
	w.Rooms[temple.Vnum] = temple
	ch := newCaster("Caster", 20)
	handler.CharToRoom(ch, src)

	SpellWordOfRecall(w, 0, ch.Level, ch, ch)

	if ch.InRoom != src {
		t.Error("NO_RECALL should prevent word_of_recall")
	}
}

func TestSpellGroupTeleport_MovesFollower(t *testing.T) {
	w := newMagicWorld()
	src := &types.RoomIndexData{Vnum: 9650, Name: "Caster's Room"}
	distant := &types.RoomIndexData{Vnum: 9651, Name: "Far Away"}
	ch := newCaster("Leader", 30)
	handler.CharToRoom(ch, src)
	w.AddChar(ch)
	follower := &types.CharData{Name: "pet", Level: 5}
	follower.Master = ch
	follower.Leader = ch
	handler.CharToRoom(follower, distant)
	w.AddChar(follower)

	SpellGroupTeleport(w, 0, ch.Level, ch, ch)

	if follower.InRoom != src {
		t.Errorf("follower should be pulled to caster's room, in %v", follower.InRoom)
	}
}

func TestSpellGroupTeleport_NoFollowers(t *testing.T) {
	w := newMagicWorld()
	src := &types.RoomIndexData{Vnum: 9652, Name: "Lonely"}
	ch, client := newCasterWithDesc("Loner", 30)
	handler.CharToRoom(ch, src)
	w.AddChar(ch)

	SpellGroupTeleport(w, 0, ch.Level, ch, ch)

	out := readOutput(ch, client)
	if out == "" {
		t.Error("expected a 'no followers' message")
	}
}

func TestSpellRegistry_GroupC(t *testing.T) {
	for _, name := range []string{
		"spell_astral_walk", "spell_gate", "spell_mist_walk",
		"spell_transport", "spell_word_of_recall", "spell_group_teleport",
	} {
		if FindSpellFunc(name) == nil {
			t.Errorf("Group C spell %q should be registered", name)
		}
	}
}
