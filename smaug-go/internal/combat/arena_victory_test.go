package combat

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// setupArenaVictoryWorld wires a world with the three rooms the branch
// needs (arena, altar, temple) plus two PCs placed in the arena room.
// Returns winner, loser, plus the three rooms so tests can assert
// teleport destinations.
func setupArenaVictoryWorld(t *testing.T) (*world.World, *types.CharData, *types.CharData, *types.RoomIndexData, *types.RoomIndexData, *types.RoomIndexData) {
	t.Helper()
	w := world.New("/tmp/test")
	WorldRef = w

	arena := &types.RoomIndexData{Vnum: types.ROOM_VNUM_ARENA_MIN, Name: "Arena"}
	arena.RoomFlags.Set(types.ROOM_ARENA)
	w.Rooms[arena.Vnum] = arena
	altar := &types.RoomIndexData{Vnum: types.ROOM_VNUM_ALTAR, Name: "Altar"}
	w.Rooms[altar.Vnum] = altar
	temple := &types.RoomIndexData{Vnum: types.ROOM_VNUM_TEMPLE, Name: "Temple"}
	w.Rooms[temple.Vnum] = temple

	winner := newFighter("Alice", 10)
	winner.PCData = &types.PCData{}
	loser := newFighter("Bob", 10)
	loser.PCData = &types.PCData{}
	handler.CharToRoom(winner, arena)
	handler.CharToRoom(loser, arena)
	w.AddChar(winner)
	w.AddChar(loser)
	return w, winner, loser, arena, altar, temple
}

func TestArenaVictory_NoFireIfVictimNotDead(t *testing.T) {
	_, winner, loser, _, _, _ := setupArenaVictoryWorld(t)
	loser.Position = types.POS_FIGHTING
	if ArenaVictoryCheck(winner, loser) {
		t.Error("should return false when victim not dead")
	}
}

func TestArenaVictory_NoFireIfNPC(t *testing.T) {
	_, winner, loser, _, _, _ := setupArenaVictoryWorld(t)
	loser.Position = types.POS_DEAD
	loser.Act.Set(types.ACT_IS_NPC)
	if ArenaVictoryCheck(winner, loser) {
		t.Error("should return false when victim is NPC")
	}
	// Reset and test ch-is-NPC branch.
	loser.Act.Remove(types.ACT_IS_NPC)
	winner.Act.Set(types.ACT_IS_NPC)
	if ArenaVictoryCheck(winner, loser) {
		t.Error("should return false when attacker is NPC")
	}
}

func TestArenaVictory_NoFireIfNotInArena(t *testing.T) {
	w, winner, loser, _, _, _ := setupArenaVictoryWorld(t)
	// Move loser to a non-arena room.
	nonArena := &types.RoomIndexData{Vnum: 4242, Name: "Plain"}
	w.Rooms[4242] = nonArena
	handler.CharFromRoom(loser)
	handler.CharToRoom(loser, nonArena)
	loser.Position = types.POS_DEAD
	if ArenaVictoryCheck(winner, loser) {
		t.Error("should return false when victim not in ROOM_ARENA")
	}
}

func TestArenaVictory_TeleportAndHeal(t *testing.T) {
	_, winner, loser, _, altar, temple := setupArenaVictoryWorld(t)
	// Damage both so the heal is observable.
	winner.Hit = 5
	winner.Mana = 0
	winner.Move = 0
	loser.Hit = -50
	loser.Mana = 0
	loser.Move = 0
	loser.Position = types.POS_DEAD

	if !ArenaVictoryCheck(winner, loser) {
		t.Fatal("ArenaVictoryCheck should fire for PvP arena death")
	}
	if loser.InRoom != altar {
		t.Errorf("loser should be at altar; got %+v", loser.InRoom)
	}
	if winner.InRoom != temple {
		t.Errorf("winner should be at temple; got %+v", winner.InRoom)
	}
	if winner.Hit != winner.MaxHit {
		t.Errorf("winner.Hit = %d, want MaxHit %d", winner.Hit, winner.MaxHit)
	}
	if loser.Hit != loser.MaxHit {
		t.Errorf("loser.Hit = %d, want MaxHit %d", loser.Hit, loser.MaxHit)
	}
	if winner.Mana != winner.MaxMana || loser.Mana != loser.MaxMana {
		t.Errorf("mana should be restored: winner=%d/%d loser=%d/%d",
			winner.Mana, winner.MaxMana, loser.Mana, loser.MaxMana)
	}
	if winner.Move != winner.MaxMove || loser.Move != loser.MaxMove {
		t.Errorf("move should be restored: winner=%d/%d loser=%d/%d",
			winner.Move, winner.MaxMove, loser.Move, loser.MaxMove)
	}
}

func TestArenaVictory_AKillsADeathsIncremented(t *testing.T) {
	_, winner, loser, _, _, _ := setupArenaVictoryWorld(t)
	loser.Position = types.POS_DEAD
	winner.PCData.AKills = 4
	loser.PCData.ADeaths = 2
	if !ArenaVictoryCheck(winner, loser) {
		t.Fatal("branch should fire")
	}
	if winner.PCData.AKills != 5 {
		t.Errorf("winner.AKills = %d, want 5", winner.PCData.AKills)
	}
	if loser.PCData.ADeaths != 3 {
		t.Errorf("loser.ADeaths = %d, want 3", loser.PCData.ADeaths)
	}
}

func TestArenaVictory_AllFlagsClearedBoth(t *testing.T) {
	_, winner, loser, _, _, _ := setupArenaVictoryWorld(t)
	loser.Position = types.POS_DEAD
	// Pre-set all 6 bits.
	winner.Act.Set(types.ACT_CHALLENGER)
	winner.Act.Set(types.ACT_CHALLENGED)
	winner.Act.Set(types.PLR_SILENCE)
	loser.Act.Set(types.ACT_CHALLENGER)
	loser.Act.Set(types.ACT_CHALLENGED)
	loser.Act.Set(types.PLR_SILENCE)

	if !ArenaVictoryCheck(winner, loser) {
		t.Fatal("branch should fire")
	}

	for _, p := range []*types.CharData{winner, loser} {
		if p.Act.IsSet(types.ACT_CHALLENGER) {
			t.Errorf("%s: ACT_CHALLENGER not cleared", p.Name)
		}
		if p.Act.IsSet(types.ACT_CHALLENGED) {
			t.Errorf("%s: ACT_CHALLENGED not cleared", p.Name)
		}
		if p.Act.IsSet(types.PLR_SILENCE) {
			t.Errorf("%s: PLR_SILENCE not cleared", p.Name)
		}
	}
}

func TestArenaVictory_IsBusyClearedAfter(t *testing.T) {
	_, winner, loser, _, _, _ := setupArenaVictoryWorld(t)
	loser.Position = types.POS_DEAD
	var gotBusyCall bool
	ArenaIsBusyFunc = func(v bool) {
		if v {
			t.Error("unexpected SetArenaIsBusy(true) call")
		}
		gotBusyCall = true
	}
	defer func() { ArenaIsBusyFunc = nil }()

	if !ArenaVictoryCheck(winner, loser) {
		t.Fatal("branch should fire")
	}
	if !gotBusyCall {
		t.Error("ArenaIsBusyFunc(false) should have been invoked")
	}
}

func TestArenaVictory_DoLookFuncInvokedForBoth(t *testing.T) {
	_, winner, loser, _, _, _ := setupArenaVictoryWorld(t)
	loser.Position = types.POS_DEAD
	var lookCalls []string
	DoLookFunc = func(c *types.CharData, arg string) {
		lookCalls = append(lookCalls, c.Name+":"+arg)
	}
	defer func() { DoLookFunc = nil }()

	if !ArenaVictoryCheck(winner, loser) {
		t.Fatal("branch should fire")
	}
	// Loser-look fires first (C ordering), then winner-look.
	if len(lookCalls) != 2 {
		t.Fatalf("expected 2 DoLookFunc calls, got %d: %v", len(lookCalls), lookCalls)
	}
	if lookCalls[0] != "Bob:auto" || lookCalls[1] != "Alice:auto" {
		t.Errorf("DoLookFunc call order wrong: %v", lookCalls)
	}
}

// TestArenaVictory_FallbackAltarToTemple — if the altar vnum isn't
// loaded (builder data gap), loser falls back to temple instead of
// becoming roomless. Pins Open Q5.
func TestArenaVictory_FallbackAltarToTemple(t *testing.T) {
	w, winner, loser, _, _, temple := setupArenaVictoryWorld(t)
	delete(w.Rooms, types.ROOM_VNUM_ALTAR)
	loser.Position = types.POS_DEAD
	if !ArenaVictoryCheck(winner, loser) {
		t.Fatal("branch should fire")
	}
	if loser.InRoom != temple {
		t.Errorf("loser should fall back to temple; got %+v", loser.InRoom)
	}
}

// TestArenaVictory_NoCorpseSpawned asserts the arena branch ran via
// Damage and no corpse was left in the arena room. Uses the real
// Damage hook so we exercise the death-path short-circuit.
func TestArenaVictory_NoCorpseSpawned(t *testing.T) {
	w, winner, loser, arena, _, _ := setupArenaVictoryWorld(t)
	loser.Hit = 1
	// Deal enough damage to send POS_DEAD via normal Damage path.
	ret := Damage(w, winner, loser, 1000, types.TYPE_HIT)
	if ret != rVICT_DIED {
		t.Errorf("Damage retcode = %d, want rVICT_DIED", ret)
	}
	for _, o := range arena.Contents {
		if o.ItemType == types.ITEM_CORPSE_PC || o.ItemType == types.ITEM_CORPSE_NPC {
			t.Fatalf("arena room should have no corpse; found %s", o.ShortDescr)
		}
	}
	// Loser teleported out of arena.
	if loser.InRoom == arena {
		t.Errorf("loser should be teleported out of arena; still in %+v", loser.InRoom)
	}
}

// TestArenaVictory_DebuffsStripped pins the debuff-strip pass. We use
// synthetic gsns (set via the package vars) — can't rely on
// LookupSkillSlotHook without a full skill table.
func TestArenaVictory_DebuffsStripped(t *testing.T) {
	_, winner, loser, _, _, _ := setupArenaVictoryWorld(t)
	loser.Position = types.POS_DEAD

	// Stub gsns so the strip pass has something to strip. Snapshot and
	// restore after so other tests are unaffected.
	origPoison := gsnPoison
	gsnPoison = 1
	defer func() { gsnPoison = origPoison }()

	// Add a matching affect to the loser.
	loser.Affects = append(loser.Affects, &types.AffectData{Type: 1, Duration: 10})
	loser.Affects = append(loser.Affects, &types.AffectData{Type: 99, Duration: 10}) // unrelated

	if !ArenaVictoryCheck(winner, loser) {
		t.Fatal("branch should fire")
	}
	for _, a := range loser.Affects {
		if a.Type == 1 {
			t.Error("poison affect should have been stripped")
		}
	}
	// Non-arena debuff remains.
	found := false
	for _, a := range loser.Affects {
		if a.Type == 99 {
			found = true
		}
	}
	if !found {
		t.Error("non-arena affect should remain")
	}
}
