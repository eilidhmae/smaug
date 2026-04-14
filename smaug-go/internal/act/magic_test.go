package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

func TestDoCast_NPC(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Guard")
	defer client.Close()
	ch.Act.Set(types.ACT_IS_NPC)

	DoCast(ch, "magic missile")
	out := readOutput(ch, client)
	// NPC returns early, no output
	if out != "" {
		t.Errorf("NPC should get no output, got: %q", out)
	}
}

func TestDoCast_NoArg(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Caster")
	defer client.Close()

	DoCast(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Cast which") {
		t.Errorf("expected 'Cast which', got: %q", out)
	}
}

func TestDoCast_UnknownSpell(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Caster")
	defer client.Close()

	DoCast(ch, "nonexistentspell")
	out := readOutput(ch, client)
	if !strings.Contains(out, "don't know") {
		t.Errorf("expected 'don't know' message, got: %q", out)
	}
}

func TestDoCast_NotEnoughMana(t *testing.T) {
	w := setupWizWorld()
	// Create a simple spell in the world skills
	w.Skills = make([]*types.SkillType, 500)
	w.Skills[1] = &types.SkillType{
		Name:        "magic missile",
		Type:        types.SKILL_SPELL,
		MinMana:     50,
		Target:      types.TAR_CHAR_OFFENSIVE,
		SpellFunName: "spell_magic_missile",
	}

	room := &types.RoomIndexData{Vnum: 9200, Name: "Test"}
	w.Rooms[9200] = room

	ch, client := makeTestChar("Caster")
	defer client.Close()
	ch.Mana = 10 // Not enough
	handler.CharToRoom(ch, room)

	DoCast(ch, "'magic missile'")
	out := readOutput(ch, client)
	if !strings.Contains(out, "enough mana") {
		t.Errorf("expected 'enough mana' message, got: %q", out)
	}
}

func TestDoCast_OffensiveNoTarget(t *testing.T) {
	w := setupWizWorld()
	w.Skills = make([]*types.SkillType, 500)
	w.Skills[1] = &types.SkillType{
		Name:        "magic missile",
		Type:        types.SKILL_SPELL,
		MinMana:     10,
		Target:      types.TAR_CHAR_OFFENSIVE,
		SpellFunName: "spell_magic_missile",
	}

	room := &types.RoomIndexData{Vnum: 9201, Name: "Test"}
	w.Rooms[9201] = room

	ch, client := makeTestChar("Caster")
	defer client.Close()
	ch.Mana = 100
	handler.CharToRoom(ch, room)

	DoCast(ch, "'magic missile'")
	out := readOutput(ch, client)
	if !strings.Contains(out, "on whom") {
		t.Errorf("expected 'on whom' message, got: %q", out)
	}
}

// --- Tier 2: ROOM_NO_MAGIC guard ---

func TestDoCast_NoMagicRoom(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 9210, Name: "Dead Magic Zone"}
	room.RoomFlags.Set(types.ROOM_NO_MAGIC)
	w.Rooms[9210] = room

	ch, client := makeTestChar("Caster")
	defer client.Close()
	ch.Trust = 0
	ch.Level = 10
	ch.Mana = 100
	handler.CharToRoom(ch, room)

	DoCast(ch, "'magic missile' dummy")
	out := readOutput(ch, client)
	if !strings.Contains(strings.ToLower(out), "blocking your magic") {
		t.Errorf("expected 'blocking your magic' in NO_MAGIC room, got: %q", out)
	}
	// Mana must not be deducted — the guard fires before dispatch.
	if ch.Mana != 100 {
		t.Errorf("Mana should be untouched in NO_MAGIC; got %d", ch.Mana)
	}
}

func TestDoCast_NoMagicRoomImmortalBypass(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 9211, Name: "Dead Magic Zone"}
	room.RoomFlags.Set(types.ROOM_NO_MAGIC)
	w.Rooms[9211] = room

	ch, client := makeTestChar("God")
	defer client.Close()
	ch.Trust = types.LEVEL_IMMORTAL
	ch.Mana = 100
	handler.CharToRoom(ch, room)

	// No spells defined → will fall through to "don't know" rather than NO_MAGIC.
	DoCast(ch, "'magic missile' dummy")
	out := readOutput(ch, client)
	if strings.Contains(strings.ToLower(out), "blocking your magic") {
		t.Errorf("immortal should bypass NO_MAGIC, got: %q", out)
	}
}
