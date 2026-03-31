package game

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func newUpdateTestWorld() (*world.World, *GameLoop) {
	w := world.New("/tmp/test")
	g := &GameLoop{world: w}
	return w, g
}

func TestHitGain_NPC(t *testing.T) {
	mob := &types.CharData{Level: 10, Position: types.POS_STANDING}
	mob.Act.Set(types.ACT_IS_NPC)
	mob.MaxHit = 100
	mob.Hit = 80

	gain := hitGain(mob)
	if gain != 15 { // level * 3 / 2 = 15
		t.Errorf("hitGain NPC = %d, want 15", gain)
	}
}

func TestHitGain_PlayerStanding(t *testing.T) {
	ch := &types.CharData{Level: 10, Position: types.POS_STANDING, PermCon: 15}
	ch.PCData = &types.PCData{Condition: [4]int{0, 48, 48, 0}}
	ch.MaxHit = 100
	ch.Hit = 80

	gain := hitGain(ch)
	// UMIN(5, 10) = 5, standing gets no position bonus
	if gain != 5 {
		t.Errorf("hitGain player standing = %d, want 5", gain)
	}
}

func TestHitGain_PlayerSleeping(t *testing.T) {
	ch := &types.CharData{Level: 10, Position: types.POS_SLEEPING, PermCon: 15}
	ch.PCData = &types.PCData{Condition: [4]int{0, 48, 48, 0}}
	ch.MaxHit = 200
	ch.Hit = 100

	gain := hitGain(ch)
	// base 5, + con*2 = 30, total = 35
	if gain != 35 {
		t.Errorf("hitGain player sleeping = %d, want 35", gain)
	}
}

func TestHitGain_Poisoned(t *testing.T) {
	ch := &types.CharData{Level: 10, Position: types.POS_STANDING, PermCon: 15}
	ch.PCData = &types.PCData{Condition: [4]int{0, 48, 48, 0}}
	ch.AffectedBy.Set(types.AFF_POISON)
	ch.MaxHit = 100
	ch.Hit = 80

	gain := hitGain(ch)
	// base 5, poison /4 = 1
	if gain != 1 {
		t.Errorf("hitGain poisoned = %d, want 1", gain)
	}
}

func TestManaGain_PlayerSleeping(t *testing.T) {
	ch := &types.CharData{Level: 10, Position: types.POS_SLEEPING, PermInt: 15}
	ch.PCData = &types.PCData{Condition: [4]int{0, 48, 48, 0}}
	ch.MaxMana = 200
	ch.Mana = 100

	gain := manaGain(ch)
	// base UMIN(5, 10/2) = 5, + int*3 = 45, total = 50
	if gain != 50 {
		t.Errorf("manaGain player sleeping = %d, want 50", gain)
	}
}

func TestMoveGain_PlayerResting(t *testing.T) {
	ch := &types.CharData{Level: 10, Position: types.POS_RESTING, PermDex: 15}
	ch.PCData = &types.PCData{Condition: [4]int{0, 48, 48, 0}}
	ch.MaxMove = 200
	ch.Move = 100

	gain := moveGain(ch)
	// base UMAX(15, 2*10) = 20, + dex*2 = 30, total = 50
	if gain != 50 {
		t.Errorf("moveGain player resting = %d, want 50", gain)
	}
}

func TestCharUpdate_Regen(t *testing.T) {
	w, g := newUpdateTestWorld()
	room := &types.RoomIndexData{Vnum: 6000, Name: "Test Room"}
	w.Rooms[6000] = room

	ch := &types.CharData{
		Name:     "Tester",
		Level:    10,
		Position: types.POS_STANDING,
		PermCon:  15,
		PermInt:  15,
		PermDex:  15,
		Hit:      50,
		MaxHit:   100,
		Mana:     20,
		MaxMana:  100,
		Move:     30,
		MaxMove:  100,
		PCData:   &types.PCData{Condition: [4]int{0, 48, 48, 0}},
	}
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	g.charUpdate()

	if ch.Hit <= 50 {
		t.Errorf("Hit = %d, should have increased from 50", ch.Hit)
	}
	if ch.Mana <= 20 {
		t.Errorf("Mana = %d, should have increased from 20", ch.Mana)
	}
	if ch.Move <= 30 {
		t.Errorf("Move = %d, should have increased from 30", ch.Move)
	}
}

func TestCharUpdate_AffectExpiry(t *testing.T) {
	w, g := newUpdateTestWorld()
	room := &types.RoomIndexData{Vnum: 6001, Name: "Test Room"}
	w.Rooms[6001] = room

	ch := &types.CharData{
		Name:     "Tester",
		Level:    10,
		Position: types.POS_STANDING,
		PermStr:  15,
		Hit:      100,
		MaxHit:   100,
		Mana:     100,
		MaxMana:  100,
		Move:     100,
		MaxMove:  100,
		PCData:   &types.PCData{Condition: [4]int{0, 48, 48, 0}},
	}
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	// Add affect that will expire this tick
	handler.AffectToChar(ch, &types.AffectData{
		Type:     1,
		Duration: 1,
		Location: types.APPLY_STR,
		Modifier: 5,
	})

	if ch.ModStr != 5 {
		t.Fatalf("ModStr = %d, want 5 before update", ch.ModStr)
	}

	// Tick down duration
	g.charUpdate()

	// Duration should be 0 now, will be removed next tick
	if len(ch.Affects) != 1 || ch.Affects[0].Duration != 0 {
		t.Errorf("Affect duration should be 0 after first tick")
	}

	// Second tick removes it
	g.charUpdate()

	if len(ch.Affects) != 0 {
		t.Errorf("Affect should be removed after expiry, has %d", len(ch.Affects))
	}
	if ch.ModStr != 0 {
		t.Errorf("ModStr = %d, want 0 after affect expired", ch.ModStr)
	}
}

func TestObjUpdate_CorpseDecay(t *testing.T) {
	w, g := newUpdateTestWorld()
	room := &types.RoomIndexData{Vnum: 6002, Name: "Test Room"}
	w.Rooms[6002] = room

	idx := &types.ObjIndexData{Vnum: 9000, Name: "corpse of guard",
		ShortDescr: "the corpse of a guard", ItemType: types.ITEM_CORPSE_NPC}
	w.ObjIndex[9000] = idx
	corpse := handler.CreateObject(w, idx, 1)
	corpse.Timer = 1 // Will decay next tick
	handler.ObjToRoom(corpse, room)

	g.objUpdate()

	if len(room.Contents) != 0 {
		t.Errorf("room.Contents = %d, want 0 (corpse should have decayed)", len(room.Contents))
	}
}

func TestMobileUpdate_Wander(t *testing.T) {
	w, g := newUpdateTestWorld()
	room1 := &types.RoomIndexData{Vnum: 6010, Name: "Room 1"}
	room2 := &types.RoomIndexData{Vnum: 6011, Name: "Room 2"}
	room1.Exits = []*types.ExitData{
		{Direction: types.DIR_NORTH, ToRoom: room2},
	}
	w.Rooms[6010] = room1
	w.Rooms[6011] = room2

	mobIdx := &types.MobIndexData{
		Vnum: 7000, PlayerName: "wanderer", Level: 5,
		Position: types.POS_STANDING, DefPosition: types.POS_STANDING,
	}
	w.MobIndex[7000] = mobIdx

	mob := handler.CreateMobile(w, mobIdx)
	handler.CharToRoom(mob, room1)

	// Run many mobile updates — the mob may or may not move (random)
	// Just verify it doesn't crash and the mob is still in a valid room
	for i := 0; i < 100; i++ {
		g.mobileUpdate()
	}

	if mob.InRoom == nil {
		t.Error("mob should still be in a room after wander updates")
	}
}
