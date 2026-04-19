package mudprog

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// setupMpMorphWorld returns a world with a single morph, one NPC mob,
// and one PC victim — the standard mpmorph test bed.
func setupMpMorphWorld(t *testing.T) (w *world.World, mob, victim *types.CharData) {
	t.Helper()
	w = world.New("")
	room := &types.RoomIndexData{Vnum: 3001, Name: "Test Room"}
	w.Rooms[3001] = room
	m := &types.MorphData{Name: "wolf", Vnum: 1000, Level: 10}
	w.Morphs = []*types.MorphData{m}

	mob = makeNPC("trickster")
	mob.InRoom = room
	room.People = append(room.People, mob)

	victim = &types.CharData{
		Name:  "Victim",
		Level: 10,
		Hit:   100, MaxHit: 100,
		Mana: 100, MaxMana: 100,
		Move: 100, MaxMove: 100,
		Position: types.POS_STANDING,
		PCData:   &types.PCData{},
		InRoom:   room,
	}
	room.People = append(room.People, victim)
	w.Characters = []*types.CharData{mob, victim}

	prev := WorldRef
	t.Cleanup(func() { WorldRef = prev })
	WorldRef = w
	return
}

func TestMpMorph_ByName_AppliesMorph(t *testing.T) {
	_, mob, victim := setupMpMorphWorld(t)
	mpMorph(mob, "Victim wolf")
	if victim.Morph == nil {
		t.Fatal("victim should be morphed")
	}
	if victim.Morph.Morph.Name != "wolf" {
		t.Errorf("morph name = %q, want wolf", victim.Morph.Morph.Name)
	}
}

func TestMpMorph_ByVnum_AppliesMorph(t *testing.T) {
	_, mob, victim := setupMpMorphWorld(t)
	mpMorph(mob, "Victim 1000")
	if victim.Morph == nil {
		t.Fatal("victim should be morphed (by vnum)")
	}
}

func TestMpMorph_MissingArgsNoop(t *testing.T) {
	_, mob, victim := setupMpMorphWorld(t)
	mpMorph(mob, "")
	if victim.Morph != nil {
		t.Error("empty args should noop")
	}
	mpMorph(mob, "Victim")
	if victim.Morph != nil {
		t.Error("single-arg should noop (need target + morph)")
	}
}

func TestMpMorph_UnknownVictimNoop(t *testing.T) {
	_, mob, _ := setupMpMorphWorld(t)
	mpMorph(mob, "Nobody wolf")
	// No panic — just a noop. Check via a fresh fetch.
	if mob.Morph != nil {
		t.Error("mob itself should not be morphed")
	}
}

func TestMpMorph_UnknownMorphNoop(t *testing.T) {
	_, mob, victim := setupMpMorphWorld(t)
	mpMorph(mob, "Victim nonexistent")
	if victim.Morph != nil {
		t.Error("unknown morph should noop")
	}
}

func TestMpMorph_AlreadyMorphedSkips(t *testing.T) {
	w, mob, victim := setupMpMorphWorld(t)
	other := &types.MorphData{Name: "bear", Vnum: 1001, Level: 10}
	w.Morphs = append(w.Morphs, other)
	victim.Morph = &types.CharMorph{Morph: other}
	mpMorph(mob, "Victim wolf")
	// Original morph still attached — mpmorph does not swap.
	if victim.Morph == nil || victim.Morph.Morph != other {
		t.Error("already-morphed victim should retain original morph (no stacking)")
	}
}

func TestMpMorph_PCCallerBlocked(t *testing.T) {
	w, _, victim := setupMpMorphWorld(t)
	// Create a second character as a PC (has Desc).
	caller := &types.CharData{
		Name:   "Caller",
		InRoom: victim.InRoom,
		PCData: &types.PCData{},
		Level:  10,
		Desc:   &types.DescriptorData{},
	}
	w.Characters = append(w.Characters, caller)
	victim.InRoom.People = append(victim.InRoom.People, caller)
	mpMorph(caller, "Victim wolf")
	if victim.Morph != nil {
		t.Error("PC caller should be blocked from mpmorph")
	}
}

// --- mpUnmorph ------------------------------------------------------------

func TestMpUnmorph_ClearsMorph(t *testing.T) {
	_, mob, victim := setupMpMorphWorld(t)
	victim.Morph = &types.CharMorph{Morph: WorldRef.Morphs[0]}
	mpUnmorph(mob, "Victim")
	if victim.Morph != nil {
		t.Error("victim should be unmorphed")
	}
}

func TestMpUnmorph_NotMorphedNoop(t *testing.T) {
	_, mob, victim := setupMpMorphWorld(t)
	mpUnmorph(mob, "Victim")
	if victim.Morph != nil {
		t.Error("noop on not-morphed victim")
	}
}

func TestMpUnmorph_MissingArgsNoop(t *testing.T) {
	_, mob, victim := setupMpMorphWorld(t)
	victim.Morph = &types.CharMorph{Morph: WorldRef.Morphs[0]}
	mpUnmorph(mob, "")
	if victim.Morph == nil {
		t.Error("empty args should noop; victim still morphed")
	}
}
