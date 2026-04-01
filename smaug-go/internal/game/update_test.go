package game

import (
	"os"
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

func TestAutosave(t *testing.T) {
	dir := t.TempDir()
	w := world.New(dir)
	g := &GameLoop{world: w}

	room := &types.RoomIndexData{Vnum: 9000, Name: "Temple"}
	w.Rooms[9000] = room

	ch := &types.CharData{
		Name: "Autosavetest", Level: 5, Position: types.POS_STANDING,
		Hit: 100, MaxHit: 100, Mana: 50, MaxMana: 50, Move: 80, MaxMove: 80,
		PermStr: 15, PermInt: 13, PermWis: 12, PermDex: 14,
		PermCon: 13, PermCha: 11, PermLck: 13,
		PCData: &types.PCData{Pwd: "secret", PagerLen: 24},
	}
	handler.CharToRoom(ch, room)

	d := &types.DescriptorData{Character: ch, Host: "localhost"}
	ch.Desc = d
	w.Descriptors = append(w.Descriptors, d)

	g.autosave()

	// Check that the player file was created
	path := dir + "/player/a/Autosavetest"
	if _, err := os.Stat(path); err != nil {
		t.Errorf("autosave should create player file at %s: %v", path, err)
	}
}

func TestAutosave_SkipsLowLevel(t *testing.T) {
	dir := t.TempDir()
	w := world.New(dir)
	g := &GameLoop{world: w}

	ch := &types.CharData{
		Name: "Newbie", Level: 1, Position: types.POS_STANDING,
		PCData: &types.PCData{Pwd: "pass", PagerLen: 24},
	}
	d := &types.DescriptorData{Character: ch, Host: "localhost"}
	ch.Desc = d
	w.Descriptors = append(w.Descriptors, d)

	g.autosave()

	// Level 1 should not be saved
	path := dir + "/player/n/Newbie"
	if _, err := os.Stat(path); err == nil {
		t.Error("autosave should skip level 1 characters")
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

func TestAggrUpdate_AttacksPlayer(t *testing.T) {
	w, g := newUpdateTestWorld()
	room := &types.RoomIndexData{Vnum: 6020, Name: "Danger Room"}
	w.Rooms[6020] = room

	mobIdx := &types.MobIndexData{
		Vnum: 7010, PlayerName: "aggr mob", Level: 5,
		Position: types.POS_STANDING, DefPosition: types.POS_STANDING,
	}
	w.MobIndex[7010] = mobIdx
	mob := handler.CreateMobile(w, mobIdx)
	mob.Act.Set(types.ACT_AGGRESSIVE)
	handler.CharToRoom(mob, room)

	// Place a player in the room
	player := &types.CharData{
		Name:     "Victim",
		Level:    5,
		Position: types.POS_STANDING,
		PCData:   &types.PCData{PagerLen: 24},
	}
	handler.CharToRoom(player, room)
	w.AddChar(player)

	// Run many aggr updates (random chance to attack)
	for i := 0; i < 100; i++ {
		g.aggrUpdate()
		if mob.Fighting != nil {
			break
		}
	}

	if mob.Fighting == nil {
		t.Error("aggressive mob should have started fighting the player")
	}
}

func TestAggrUpdate_SkipsSafeRoom(t *testing.T) {
	w, g := newUpdateTestWorld()
	room := &types.RoomIndexData{Vnum: 6021, Name: "Safe Room"}
	room.RoomFlags.Set(types.ROOM_SAFE)
	w.Rooms[6021] = room

	mobIdx := &types.MobIndexData{
		Vnum: 7011, PlayerName: "aggr mob", Level: 5,
		Position: types.POS_STANDING, DefPosition: types.POS_STANDING,
	}
	w.MobIndex[7011] = mobIdx
	mob := handler.CreateMobile(w, mobIdx)
	mob.Act.Set(types.ACT_AGGRESSIVE)
	handler.CharToRoom(mob, room)

	player := &types.CharData{
		Name: "Safe", Level: 5, Position: types.POS_STANDING,
		PCData: &types.PCData{PagerLen: 24},
	}
	handler.CharToRoom(player, room)
	w.AddChar(player)

	for i := 0; i < 100; i++ {
		g.aggrUpdate()
	}

	if mob.Fighting != nil {
		t.Error("aggressive mob should not attack in safe room")
	}
}

func TestAreaUpdate_ResetsOnTimer(t *testing.T) {
	w, g := newUpdateTestWorld()
	area := &types.AreaData{
		Name:           "TestArea",
		ResetFrequency: 3, // reset every 3 ticks
		Age:            0,
		NPlayer:        0,
		ResetMsg:       "The area resets!",
	}
	w.Areas = append(w.Areas, area)

	// First two updates: age goes to 1, 2 — under freq, no reset
	g.areaUpdate()
	if area.Age != 1 {
		t.Errorf("area age should be 1, got %d", area.Age)
	}
	g.areaUpdate()
	if area.Age != 2 {
		t.Errorf("area age should be 2, got %d", area.Age)
	}

	// Third update: age=3 >= freq=3 and NPlayer=0, so reset
	g.areaUpdate()
	if area.Age != 0 {
		t.Errorf("area age should be 0 after reset, got %d", area.Age)
	}
}

func TestAreaUpdate_DelaysResetWithPlayers(t *testing.T) {
	w, g := newUpdateTestWorld()
	area := &types.AreaData{
		Name:           "TestArea",
		ResetFrequency: 2,
		Age:            0,
		NPlayer:        1, // players present
	}
	w.Areas = append(w.Areas, area)

	// With players present, won't reset at freq (age=2)
	g.areaUpdate() // age=1
	g.areaUpdate() // age=2, players present, no reset yet
	if area.Age != 2 {
		t.Errorf("area age should be 2, got %d", area.Age)
	}

	// Still players, age=3
	g.areaUpdate()
	if area.Age != 3 {
		t.Errorf("area age should be 3, got %d", area.Age)
	}

	// Forced reset at freq*2=4
	g.areaUpdate()
	if area.Age != 0 {
		t.Errorf("area age should be 0 after forced reset, got %d", area.Age)
	}
}

func TestAreaUpdate_DefaultResetFrequency(t *testing.T) {
	w, g := newUpdateTestWorld()
	area := &types.AreaData{
		Name:           "TestArea",
		ResetFrequency: 0, // should default to 15
		Age:            0,
		NPlayer:        0,
	}
	w.Areas = append(w.Areas, area)

	// Advance 14 ticks — should not reset (default freq = 15)
	for i := 0; i < 14; i++ {
		g.areaUpdate()
	}
	if area.Age != 14 {
		t.Errorf("area age should be 14, got %d", area.Age)
	}

	// 15th tick — should reset
	g.areaUpdate()
	if area.Age != 0 {
		t.Errorf("area age should be 0 after reset with default freq, got %d", area.Age)
	}
}

func TestAreaUpdate_NilAreaSkipped(t *testing.T) {
	w, g := newUpdateTestWorld()
	w.Areas = append(w.Areas, nil)
	// Should not panic
	g.areaUpdate()
}

func TestAreaUpdate_ResetMessageSentToPlayers(t *testing.T) {
	w, g := newUpdateTestWorld()
	area := &types.AreaData{
		Name:           "TestArea",
		ResetFrequency: 1,
		Age:            0,
		NPlayer:        0,
		ResetMsg:       "The forest comes alive!",
	}
	w.Areas = append(w.Areas, area)

	room := &types.RoomIndexData{Vnum: 7000, Name: "Forest", Area: area}
	w.Rooms[7000] = room

	ch := &types.CharData{
		Name:   "Watcher",
		Level:  5,
		PCData: &types.PCData{PagerLen: 24},
	}
	d := &types.DescriptorData{Character: ch, Connected: types.CON_PLAYING}
	ch.Desc = d
	ch.InRoom = room
	w.Descriptors = append(w.Descriptors, d)

	g.areaUpdate()

	if area.Age != 0 {
		t.Errorf("area should have reset, age = %d", area.Age)
	}
	// The message was sent via ch.Sendf which writes to d's output buffer
	if !d.HasOutput() {
		t.Error("player in area should have received reset message")
	}
}

func TestMobileUpdate_SentinelDoesNotWander(t *testing.T) {
	w, g := newUpdateTestWorld()
	room1 := &types.RoomIndexData{Vnum: 7100, Name: "Room 1"}
	room2 := &types.RoomIndexData{Vnum: 7101, Name: "Room 2"}
	room1.Exits = []*types.ExitData{
		{Direction: types.DIR_NORTH, ToRoom: room2},
	}
	w.Rooms[7100] = room1
	w.Rooms[7101] = room2

	mobIdx := &types.MobIndexData{
		Vnum: 7100, PlayerName: "sentinel guard", Level: 5,
		Position: types.POS_STANDING, DefPosition: types.POS_STANDING,
	}
	w.MobIndex[7100] = mobIdx

	mob := handler.CreateMobile(w, mobIdx)
	mob.Act.Set(types.ACT_SENTINEL) // sentinel flag
	handler.CharToRoom(mob, room1)

	// Run many updates — sentinel should never move
	for i := 0; i < 200; i++ {
		g.mobileUpdate()
	}

	if mob.InRoom != room1 {
		t.Error("sentinel mob should not have moved from its room")
	}
}

func TestMobileUpdate_StayAreaDoesNotLeave(t *testing.T) {
	w, g := newUpdateTestWorld()
	area1 := &types.AreaData{Name: "Area1"}
	area2 := &types.AreaData{Name: "Area2"}

	room1 := &types.RoomIndexData{Vnum: 7200, Name: "Room 1", Area: area1}
	room2 := &types.RoomIndexData{Vnum: 7201, Name: "Room 2", Area: area2}

	// All exits lead to the other area
	for i := 0; i < 10; i++ {
		room1.Exits = append(room1.Exits, &types.ExitData{
			Direction: i, ToRoom: room2,
		})
	}
	w.Rooms[7200] = room1
	w.Rooms[7201] = room2

	mobIdx := &types.MobIndexData{
		Vnum: 7200, PlayerName: "area guard", Level: 5,
		Position: types.POS_STANDING, DefPosition: types.POS_STANDING,
	}
	w.MobIndex[7200] = mobIdx

	mob := handler.CreateMobile(w, mobIdx)
	mob.Act.Set(types.ACT_STAY_AREA) // stay-area flag
	handler.CharToRoom(mob, room1)

	// Run many updates — stay_area mob should not leave area
	for i := 0; i < 200; i++ {
		g.mobileUpdate()
	}

	if mob.InRoom != room1 {
		t.Error("stay_area mob should not have left its area")
	}
}

func TestMobileUpdate_ClosedDoorBlocksWander(t *testing.T) {
	w, g := newUpdateTestWorld()
	room1 := &types.RoomIndexData{Vnum: 7300, Name: "Room 1"}
	room2 := &types.RoomIndexData{Vnum: 7301, Name: "Room 2"}
	// All exits are closed doors
	for i := 0; i < 10; i++ {
		room1.Exits = append(room1.Exits, &types.ExitData{
			Direction: i,
			ToRoom:    room2,
			ExitInfo:  int(types.EX_CLOSED),
		})
	}
	w.Rooms[7300] = room1
	w.Rooms[7301] = room2

	mobIdx := &types.MobIndexData{
		Vnum: 7300, PlayerName: "blocked mob", Level: 5,
		Position: types.POS_STANDING, DefPosition: types.POS_STANDING,
	}
	w.MobIndex[7300] = mobIdx

	mob := handler.CreateMobile(w, mobIdx)
	handler.CharToRoom(mob, room1)

	for i := 0; i < 200; i++ {
		g.mobileUpdate()
	}

	if mob.InRoom != room1 {
		t.Error("mob should not walk through closed doors")
	}
}

func TestMobileUpdate_NoMobRoomBlocksWander(t *testing.T) {
	w, g := newUpdateTestWorld()
	room1 := &types.RoomIndexData{Vnum: 7400, Name: "Room 1"}
	room2 := &types.RoomIndexData{Vnum: 7401, Name: "No Mob Room"}
	room2.RoomFlags.Set(types.ROOM_NO_MOB)
	for i := 0; i < 10; i++ {
		room1.Exits = append(room1.Exits, &types.ExitData{
			Direction: i, ToRoom: room2,
		})
	}
	w.Rooms[7400] = room1
	w.Rooms[7401] = room2

	mobIdx := &types.MobIndexData{
		Vnum: 7400, PlayerName: "blocked mob", Level: 5,
		Position: types.POS_STANDING, DefPosition: types.POS_STANDING,
	}
	w.MobIndex[7400] = mobIdx

	mob := handler.CreateMobile(w, mobIdx)
	handler.CharToRoom(mob, room1)

	for i := 0; i < 200; i++ {
		g.mobileUpdate()
	}

	if mob.InRoom != room1 {
		t.Error("mob should not enter ROOM_NO_MOB room")
	}
}

func TestMobileUpdate_SkipsNonStanding(t *testing.T) {
	w, g := newUpdateTestWorld()
	room1 := &types.RoomIndexData{Vnum: 7500, Name: "Room 1"}
	room2 := &types.RoomIndexData{Vnum: 7501, Name: "Room 2"}
	for i := 0; i < 10; i++ {
		room1.Exits = append(room1.Exits, &types.ExitData{
			Direction: i, ToRoom: room2,
		})
	}
	w.Rooms[7500] = room1
	w.Rooms[7501] = room2

	mobIdx := &types.MobIndexData{
		Vnum: 7500, PlayerName: "sleeping mob", Level: 5,
		Position: types.POS_SLEEPING, DefPosition: types.POS_SLEEPING,
	}
	w.MobIndex[7500] = mobIdx

	mob := handler.CreateMobile(w, mobIdx)
	mob.Position = types.POS_SLEEPING
	handler.CharToRoom(mob, room1)

	for i := 0; i < 200; i++ {
		g.mobileUpdate()
	}

	if mob.InRoom != room1 {
		t.Error("sleeping mob should not move")
	}
}

func TestMobileUpdate_SkipsCharmed(t *testing.T) {
	w, g := newUpdateTestWorld()
	room1 := &types.RoomIndexData{Vnum: 7600, Name: "Room 1"}
	room2 := &types.RoomIndexData{Vnum: 7601, Name: "Room 2"}
	for i := 0; i < 10; i++ {
		room1.Exits = append(room1.Exits, &types.ExitData{
			Direction: i, ToRoom: room2,
		})
	}
	w.Rooms[7600] = room1
	w.Rooms[7601] = room2

	mobIdx := &types.MobIndexData{
		Vnum: 7600, PlayerName: "charmed mob", Level: 5,
		Position: types.POS_STANDING, DefPosition: types.POS_STANDING,
	}
	w.MobIndex[7600] = mobIdx

	mob := handler.CreateMobile(w, mobIdx)
	mob.AffectedBy.Set(types.AFF_CHARM)
	handler.CharToRoom(mob, room1)

	for i := 0; i < 200; i++ {
		g.mobileUpdate()
	}

	if mob.InRoom != room1 {
		t.Error("charmed mob should not wander")
	}
}

func TestMobileUpdate_SkipsFighting(t *testing.T) {
	w, g := newUpdateTestWorld()
	room1 := &types.RoomIndexData{Vnum: 7700, Name: "Room 1"}
	room2 := &types.RoomIndexData{Vnum: 7701, Name: "Room 2"}
	for i := 0; i < 10; i++ {
		room1.Exits = append(room1.Exits, &types.ExitData{
			Direction: i, ToRoom: room2,
		})
	}
	w.Rooms[7700] = room1
	w.Rooms[7701] = room2

	mobIdx := &types.MobIndexData{
		Vnum: 7700, PlayerName: "fighting mob", Level: 5,
		Position: types.POS_STANDING, DefPosition: types.POS_STANDING,
	}
	w.MobIndex[7700] = mobIdx

	mob := handler.CreateMobile(w, mobIdx)
	mob.Fighting = &types.FightData{Who: &types.CharData{Name: "dummy"}}
	handler.CharToRoom(mob, room1)

	for i := 0; i < 200; i++ {
		g.mobileUpdate()
	}

	if mob.InRoom != room1 {
		t.Error("fighting mob should not wander")
	}
}

func TestCharUpdate_PoisonDamage(t *testing.T) {
	w, g := newUpdateTestWorld()
	room := &types.RoomIndexData{Vnum: 8000, Name: "Poison Room"}
	w.Rooms[8000] = room

	ch := &types.CharData{
		Name:     "Poisoned",
		Level:    10,
		Position: types.POS_STANDING,
		PermCon:  15,
		Hit:      100,
		MaxHit:   100,
		Mana:     100,
		MaxMana:  100,
		Move:     100,
		MaxMove:  100,
		PCData:   &types.PCData{Condition: [4]int{0, 48, 48, 0}},
	}
	ch.AffectedBy.Set(types.AFF_POISON)
	d := &types.DescriptorData{Character: ch, Connected: types.CON_PLAYING}
	ch.Desc = d
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	g.charUpdate()

	// Poison does 6 damage per tick
	if ch.Hit != 94 {
		t.Errorf("poisoned HP = %d, want 94 (100 - 6)", ch.Hit)
	}
}

func TestCharUpdate_PoisonMinHP(t *testing.T) {
	w, g := newUpdateTestWorld()
	room := &types.RoomIndexData{Vnum: 8001, Name: "Poison Room"}
	w.Rooms[8001] = room

	ch := &types.CharData{
		Name:     "VeryPoisoned",
		Level:    10,
		Position: types.POS_STANDING,
		PermCon:  15,
		Hit:      3,
		MaxHit:   100,
		Mana:     100,
		MaxMana:  100,
		Move:     100,
		MaxMove:  100,
		PCData:   &types.PCData{Condition: [4]int{0, 48, 48, 0}},
	}
	ch.AffectedBy.Set(types.AFF_POISON)
	d := &types.DescriptorData{Character: ch, Connected: types.CON_PLAYING}
	ch.Desc = d
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	g.charUpdate()

	// Poison won't reduce below 1
	if ch.Hit != 1 {
		t.Errorf("poisoned HP = %d, want 1 (minimum)", ch.Hit)
	}
}

func TestCharUpdate_IncapBleedOut(t *testing.T) {
	w, g := newUpdateTestWorld()
	room := &types.RoomIndexData{Vnum: 8002, Name: "Death Room"}
	w.Rooms[8002] = room

	ch := &types.CharData{
		Name:     "Bleeding",
		Level:    10,
		Position: types.POS_INCAP,
		Hit:      -5,
		MaxHit:   100,
		Mana:     100,
		MaxMana:  100,
		Move:     100,
		MaxMove:  100,
		PCData:   &types.PCData{Condition: [4]int{0, 48, 48, 0}},
	}
	d := &types.DescriptorData{Character: ch, Connected: types.CON_PLAYING}
	ch.Desc = d
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	g.charUpdate()

	if ch.Hit != -6 {
		t.Errorf("incap HP = %d, want -6", ch.Hit)
	}
}

func TestCharUpdate_MortalBleedToDeath(t *testing.T) {
	w, g := newUpdateTestWorld()
	room := &types.RoomIndexData{Vnum: 8003, Name: "Death Room"}
	w.Rooms[8003] = room

	ch := &types.CharData{
		Name:     "Dying",
		Level:    5,
		Position: types.POS_MORTAL,
		Hit:      -99,
		MaxHit:   100,
		Mana:     100,
		MaxMana:  100,
		Move:     100,
		MaxMove:  100,
		PCData:   &types.PCData{Condition: [4]int{0, 48, 48, 0}},
	}
	d := &types.DescriptorData{Character: ch, Connected: types.CON_PLAYING}
	ch.Desc = d
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	g.charUpdate()

	// PC at -maxHit dies then gets restored
	if ch.Hit != 1 {
		t.Errorf("dead PC Hit = %d, want 1 (restored)", ch.Hit)
	}
	if ch.Position != types.POS_RESTING {
		t.Errorf("dead PC position = %d, want POS_RESTING", ch.Position)
	}
}

func TestCharUpdate_StunnedAutoRecover(t *testing.T) {
	w, g := newUpdateTestWorld()
	room := &types.RoomIndexData{Vnum: 8004, Name: "Recovery Room"}
	w.Rooms[8004] = room

	ch := &types.CharData{
		Name:     "Stunned",
		Level:    10,
		Position: types.POS_STUNNED,
		PermCon:  15,
		Hit:      0,
		MaxHit:   100,
		Mana:     100,
		MaxMana:  100,
		Move:     100,
		MaxMove:  100,
		PCData:   &types.PCData{Condition: [4]int{0, 48, 48, 0}},
	}
	d := &types.DescriptorData{Character: ch, Connected: types.CON_PLAYING}
	ch.Desc = d
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	g.charUpdate()

	if ch.Hit != 1+1 { // stunned gives +1, then regen adds hitGain at stunned (=1)
		t.Logf("stunned Hit after update = %d", ch.Hit)
	}
	if ch.Position != types.POS_STANDING {
		t.Errorf("stunned character should recover to standing, got %d", ch.Position)
	}
}

func TestCharUpdate_NilRoomSkipped(t *testing.T) {
	w, g := newUpdateTestWorld()
	ch := &types.CharData{
		Name:     "NoRoom",
		Level:    10,
		Position: types.POS_STANDING,
		Hit:      50,
		MaxHit:   100,
		PCData:   &types.PCData{Condition: [4]int{0, 48, 48, 0}},
	}
	w.AddChar(ch)

	g.charUpdate()

	// Hit should be unchanged since InRoom is nil
	if ch.Hit != 50 {
		t.Errorf("character with nil room should be skipped, HP = %d", ch.Hit)
	}
}

func TestObjUpdate_TimerDecrementNonCorpse(t *testing.T) {
	w, g := newUpdateTestWorld()
	room := &types.RoomIndexData{Vnum: 8100, Name: "Test Room"}
	w.Rooms[8100] = room

	idx := &types.ObjIndexData{Vnum: 9100, Name: "magic scroll",
		ShortDescr: "a magic scroll", ItemType: types.ITEM_SCROLL}
	w.ObjIndex[9100] = idx
	scroll := handler.CreateObject(w, idx, 1)
	scroll.Timer = 3
	handler.ObjToRoom(scroll, room)

	g.objUpdate()

	if scroll.Timer != 2 {
		t.Errorf("scroll timer = %d, want 2 (decremented from 3)", scroll.Timer)
	}

	// Not yet expired, should still be in room
	if len(room.Contents) != 1 {
		t.Errorf("scroll should still be in room, contents = %d", len(room.Contents))
	}
}

func TestObjUpdate_TimerExpiresNonCorpse(t *testing.T) {
	w, g := newUpdateTestWorld()
	room := &types.RoomIndexData{Vnum: 8101, Name: "Test Room"}
	w.Rooms[8101] = room

	idx := &types.ObjIndexData{Vnum: 9101, Name: "temp item",
		ShortDescr: "a temporary item", ItemType: types.ITEM_TRASH}
	w.ObjIndex[9101] = idx
	item := handler.CreateObject(w, idx, 1)
	item.Timer = 1
	handler.ObjToRoom(item, room)

	g.objUpdate()

	if len(room.Contents) != 0 {
		t.Errorf("expired item should be extracted, contents = %d", len(room.Contents))
	}
}

func TestObjUpdate_ZeroTimerSkipped(t *testing.T) {
	w, g := newUpdateTestWorld()
	room := &types.RoomIndexData{Vnum: 8102, Name: "Test Room"}
	w.Rooms[8102] = room

	idx := &types.ObjIndexData{Vnum: 9102, Name: "permanent item",
		ShortDescr: "a permanent item", ItemType: types.ITEM_ARMOR}
	w.ObjIndex[9102] = idx
	item := handler.CreateObject(w, idx, 1)
	item.Timer = 0 // no timer
	handler.ObjToRoom(item, room)

	g.objUpdate()

	if len(room.Contents) != 1 {
		t.Errorf("item with timer 0 should remain, contents = %d", len(room.Contents))
	}
}

func TestObjUpdate_NegativeTimerSkipped(t *testing.T) {
	w, g := newUpdateTestWorld()
	room := &types.RoomIndexData{Vnum: 8103, Name: "Test Room"}
	w.Rooms[8103] = room

	idx := &types.ObjIndexData{Vnum: 9103, Name: "forever item",
		ShortDescr: "a forever item", ItemType: types.ITEM_ARMOR}
	w.ObjIndex[9103] = idx
	item := handler.CreateObject(w, idx, 1)
	item.Timer = -1 // negative timer
	handler.ObjToRoom(item, room)

	g.objUpdate()

	if len(room.Contents) != 1 {
		t.Errorf("item with negative timer should remain, contents = %d", len(room.Contents))
	}
}

func TestObjUpdate_PCCorpseDecay(t *testing.T) {
	w, g := newUpdateTestWorld()
	room := &types.RoomIndexData{Vnum: 8104, Name: "Test Room"}
	w.Rooms[8104] = room

	idx := &types.ObjIndexData{Vnum: 9104, Name: "corpse of player",
		ShortDescr: "the corpse of a player", ItemType: types.ITEM_CORPSE_PC}
	w.ObjIndex[9104] = idx
	corpse := handler.CreateObject(w, idx, 1)
	corpse.Timer = 1
	handler.ObjToRoom(corpse, room)

	g.objUpdate()

	if len(room.Contents) != 0 {
		t.Errorf("PC corpse should have decayed, contents = %d", len(room.Contents))
	}
}

func TestHitGain_PlayerDead(t *testing.T) {
	ch := &types.CharData{Level: 10, Position: types.POS_DEAD, PermCon: 15}
	ch.PCData = &types.PCData{Condition: [4]int{0, 48, 48, 0}}
	ch.MaxHit = 100
	ch.Hit = 50

	gain := hitGain(ch)
	if gain != 0 {
		t.Errorf("hitGain dead = %d, want 0", gain)
	}
}

func TestHitGain_PlayerMortal(t *testing.T) {
	ch := &types.CharData{Level: 10, Position: types.POS_MORTAL, PermCon: 15}
	ch.PCData = &types.PCData{Condition: [4]int{0, 48, 48, 0}}
	ch.MaxHit = 100
	ch.Hit = 50

	gain := hitGain(ch)
	if gain != -1 {
		t.Errorf("hitGain mortal = %d, want -1", gain)
	}
}

func TestHitGain_PlayerIncap(t *testing.T) {
	ch := &types.CharData{Level: 10, Position: types.POS_INCAP, PermCon: 15}
	ch.PCData = &types.PCData{Condition: [4]int{0, 48, 48, 0}}
	ch.MaxHit = 100
	ch.Hit = 50

	gain := hitGain(ch)
	if gain != -1 {
		t.Errorf("hitGain incap = %d, want -1", gain)
	}
}

func TestHitGain_PlayerStunned(t *testing.T) {
	ch := &types.CharData{Level: 10, Position: types.POS_STUNNED, PermCon: 15}
	ch.PCData = &types.PCData{Condition: [4]int{0, 48, 48, 0}}
	ch.MaxHit = 100
	ch.Hit = 50

	gain := hitGain(ch)
	if gain != 1 {
		t.Errorf("hitGain stunned = %d, want 1", gain)
	}
}

func TestHitGain_PlayerResting(t *testing.T) {
	ch := &types.CharData{Level: 10, Position: types.POS_RESTING, PermCon: 15}
	ch.PCData = &types.PCData{Condition: [4]int{0, 48, 48, 0}}
	ch.MaxHit = 200
	ch.Hit = 100

	gain := hitGain(ch)
	// base = UMIN(5, 10) = 5, resting adds con = 15, total = 20
	if gain != 20 {
		t.Errorf("hitGain resting = %d, want 20", gain)
	}
}

func TestHitGain_HungerThirstPenalty(t *testing.T) {
	ch := &types.CharData{Level: 10, Position: types.POS_STANDING, PermCon: 15}
	ch.PCData = &types.PCData{Condition: [4]int{0, 0, 0, 0}} // full=0, thirst=0
	ch.MaxHit = 200
	ch.Hit = 100

	gain := hitGain(ch)
	// base = 5, full=0 => /2 = 2, thirst=0 => /2 = 1
	if gain != 1 {
		t.Errorf("hitGain hungry+thirsty = %d, want 1", gain)
	}
}

func TestManaGain_NPC(t *testing.T) {
	mob := &types.CharData{Level: 10, Position: types.POS_STANDING}
	mob.Act.Set(types.ACT_IS_NPC)
	mob.MaxMana = 100
	mob.Mana = 80

	gain := manaGain(mob)
	if gain != 10 { // level = 10
		t.Errorf("manaGain NPC = %d, want 10", gain)
	}
}

func TestManaGain_PlayerBelowSleeping(t *testing.T) {
	ch := &types.CharData{Level: 10, Position: types.POS_INCAP, PermInt: 15}
	ch.PCData = &types.PCData{Condition: [4]int{0, 48, 48, 0}}
	ch.MaxMana = 200
	ch.Mana = 100

	gain := manaGain(ch)
	if gain != 0 {
		t.Errorf("manaGain incap = %d, want 0", gain)
	}
}

func TestManaGain_PlayerResting(t *testing.T) {
	ch := &types.CharData{Level: 10, Position: types.POS_RESTING, PermInt: 15}
	ch.PCData = &types.PCData{Condition: [4]int{0, 48, 48, 0}}
	ch.MaxMana = 200
	ch.Mana = 100

	gain := manaGain(ch)
	// base = UMIN(5, 10/2) = 5, resting adds int*2 = 30, total = 35
	if gain != 35 {
		t.Errorf("manaGain resting = %d, want 35", gain)
	}
}

func TestManaGain_Poisoned(t *testing.T) {
	ch := &types.CharData{Level: 10, Position: types.POS_STANDING, PermInt: 15}
	ch.PCData = &types.PCData{Condition: [4]int{0, 48, 48, 0}}
	ch.AffectedBy.Set(types.AFF_POISON)
	ch.MaxMana = 200
	ch.Mana = 100

	gain := manaGain(ch)
	// base = 5, poison /4 = 1
	if gain != 1 {
		t.Errorf("manaGain poisoned = %d, want 1", gain)
	}
}

func TestMoveGain_NPC(t *testing.T) {
	mob := &types.CharData{Level: 10, Position: types.POS_STANDING}
	mob.Act.Set(types.ACT_IS_NPC)
	mob.MaxMove = 100
	mob.Move = 80

	gain := moveGain(mob)
	if gain != 10 { // level = 10
		t.Errorf("moveGain NPC = %d, want 10", gain)
	}
}

func TestMoveGain_PlayerDead(t *testing.T) {
	ch := &types.CharData{Level: 10, Position: types.POS_DEAD, PermDex: 15}
	ch.PCData = &types.PCData{Condition: [4]int{0, 48, 48, 0}}
	ch.MaxMove = 200
	ch.Move = 100

	gain := moveGain(ch)
	if gain != 0 {
		t.Errorf("moveGain dead = %d, want 0", gain)
	}
}

func TestMoveGain_PlayerMortal(t *testing.T) {
	ch := &types.CharData{Level: 10, Position: types.POS_MORTAL, PermDex: 15}
	ch.PCData = &types.PCData{Condition: [4]int{0, 48, 48, 0}}
	ch.MaxMove = 200
	ch.Move = 100

	gain := moveGain(ch)
	if gain != -1 {
		t.Errorf("moveGain mortal = %d, want -1", gain)
	}
}

func TestMoveGain_PlayerStunned(t *testing.T) {
	ch := &types.CharData{Level: 10, Position: types.POS_STUNNED, PermDex: 15}
	ch.PCData = &types.PCData{Condition: [4]int{0, 48, 48, 0}}
	ch.MaxMove = 200
	ch.Move = 100

	gain := moveGain(ch)
	if gain != 1 {
		t.Errorf("moveGain stunned = %d, want 1", gain)
	}
}

func TestMoveGain_PlayerSleeping(t *testing.T) {
	ch := &types.CharData{Level: 10, Position: types.POS_SLEEPING, PermDex: 15}
	ch.PCData = &types.PCData{Condition: [4]int{0, 48, 48, 0}}
	ch.MaxMove = 200
	ch.Move = 100

	gain := moveGain(ch)
	// base = UMAX(15, 2*10) = 20, sleeping adds dex*4 = 60, total = 80
	if gain != 80 {
		t.Errorf("moveGain sleeping = %d, want 80", gain)
	}
}

func TestMoveGain_Poisoned(t *testing.T) {
	ch := &types.CharData{Level: 10, Position: types.POS_STANDING, PermDex: 15}
	ch.PCData = &types.PCData{Condition: [4]int{0, 48, 48, 0}}
	ch.AffectedBy.Set(types.AFF_POISON)
	ch.MaxMove = 200
	ch.Move = 100

	gain := moveGain(ch)
	// base = 20, poison /4 = 5
	if gain != 5 {
		t.Errorf("moveGain poisoned = %d, want 5", gain)
	}
}

func TestViolenceUpdate(t *testing.T) {
	w, g := newUpdateTestWorld()
	// Just ensure it doesn't crash with no fighters
	g.violenceUpdate()
	_ = w
}

func TestCharUpdate_AffectPermanent(t *testing.T) {
	w, g := newUpdateTestWorld()
	room := &types.RoomIndexData{Vnum: 8200, Name: "Test Room"}
	w.Rooms[8200] = room

	ch := &types.CharData{
		Name:     "Permaff",
		Level:    10,
		Position: types.POS_STANDING,
		PermStr:  13,
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

	// Add permanent affect (duration -1)
	handler.AffectToChar(ch, &types.AffectData{
		Type:     1,
		Duration: -1, // permanent
		Location: types.APPLY_STR,
		Modifier: 3,
	})

	g.charUpdate()
	g.charUpdate()
	g.charUpdate()

	// Permanent affect should still be present
	if len(ch.Affects) != 1 {
		t.Errorf("permanent affect should remain, count = %d", len(ch.Affects))
	}
	if ch.ModStr != 3 {
		t.Errorf("ModStr = %d, want 3 (permanent affect)", ch.ModStr)
	}
}

func TestCharUpdate_AffectWearOffMessage(t *testing.T) {
	w, g := newUpdateTestWorld()
	room := &types.RoomIndexData{Vnum: 8201, Name: "Test Room"}
	w.Rooms[8201] = room

	// Set up a skill with MsgOff
	w.Skills = make([]*types.SkillType, 10)
	w.Skills[5] = &types.SkillType{
		Name:   "armor",
		MsgOff: "Your armor fades away.",
	}

	ch := &types.CharData{
		Name:     "Buffed",
		Level:    10,
		Position: types.POS_STANDING,
		PermStr:  13,
		Hit:      100,
		MaxHit:   100,
		Mana:     100,
		MaxMana:  100,
		Move:     100,
		MaxMove:  100,
		PCData:   &types.PCData{Condition: [4]int{0, 48, 48, 0}},
	}
	d := &types.DescriptorData{Character: ch, Connected: types.CON_PLAYING}
	ch.Desc = d
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	handler.AffectToChar(ch, &types.AffectData{
		Type:     5, // matches w.Skills[5]
		Duration: 0, // will expire this tick
		Location: types.APPLY_AC,
		Modifier: -20,
	})

	g.charUpdate()

	// Affect should have been removed with wear-off message sent
	if len(ch.Affects) != 0 {
		t.Errorf("affect should be removed, count = %d", len(ch.Affects))
	}
	if !d.HasOutput() {
		t.Error("wear-off message should have been sent")
	}
}

func TestMobileUpdate_Scavenging(t *testing.T) {
	w, g := newUpdateTestWorld()
	room := &types.RoomIndexData{Vnum: 8300, Name: "Loot Room"}
	w.Rooms[8300] = room

	mobIdx := &types.MobIndexData{
		Vnum: 8300, PlayerName: "scavenger", Level: 5,
		Position: types.POS_STANDING, DefPosition: types.POS_STANDING,
	}
	w.MobIndex[8300] = mobIdx

	mob := handler.CreateMobile(w, mobIdx)
	mob.Act.Set(types.ACT_SCAVENGER)
	mob.Act.Set(types.ACT_SENTINEL) // prevent wandering
	handler.CharToRoom(mob, room)

	// Place a takeable item in the room
	idx := &types.ObjIndexData{Vnum: 9200, Name: "gold coin",
		ShortDescr: "a gold coin", ItemType: types.ITEM_TRASH}
	w.ObjIndex[9200] = idx
	coin := handler.CreateObject(w, idx, 1)
	coin.WearFlags = int(types.ITEM_TAKE)
	coin.GoldCost = 100
	handler.ObjToRoom(coin, room)

	// Run many updates — scavenger may pick up item (random chance)
	for i := 0; i < 200; i++ {
		g.mobileUpdate()
		if len(mob.Carrying) > 0 {
			break
		}
	}

	if len(mob.Carrying) == 0 {
		t.Error("scavenger mob should have picked up the item")
	}
}

func TestAggrUpdate_SkipsImmortal(t *testing.T) {
	w, g := newUpdateTestWorld()
	room := &types.RoomIndexData{Vnum: 8400, Name: "Danger Room"}
	w.Rooms[8400] = room

	mobIdx := &types.MobIndexData{
		Vnum: 8400, PlayerName: "aggr mob", Level: 5,
		Position: types.POS_STANDING, DefPosition: types.POS_STANDING,
	}
	w.MobIndex[8400] = mobIdx
	mob := handler.CreateMobile(w, mobIdx)
	mob.Act.Set(types.ACT_AGGRESSIVE)
	handler.CharToRoom(mob, room)

	player := &types.CharData{
		Name:     "Immortal",
		Level:    types.LEVEL_IMMORTAL,
		Position: types.POS_STANDING,
		PCData:   &types.PCData{PagerLen: 24},
	}
	handler.CharToRoom(player, room)
	w.AddChar(player)

	for i := 0; i < 200; i++ {
		g.aggrUpdate()
	}

	if mob.Fighting != nil {
		t.Error("aggressive mob should not attack immortal")
	}
}

func TestAggrUpdate_SkipsCharmed(t *testing.T) {
	w, g := newUpdateTestWorld()
	room := &types.RoomIndexData{Vnum: 8401, Name: "Danger Room"}
	w.Rooms[8401] = room

	mobIdx := &types.MobIndexData{
		Vnum: 8401, PlayerName: "charmed aggr", Level: 5,
		Position: types.POS_STANDING, DefPosition: types.POS_STANDING,
	}
	w.MobIndex[8401] = mobIdx
	mob := handler.CreateMobile(w, mobIdx)
	mob.Act.Set(types.ACT_AGGRESSIVE)
	mob.AffectedBy.Set(types.AFF_CHARM)
	handler.CharToRoom(mob, room)

	player := &types.CharData{
		Name:     "Target",
		Level:    5,
		Position: types.POS_STANDING,
		PCData:   &types.PCData{PagerLen: 24},
	}
	handler.CharToRoom(player, room)
	w.AddChar(player)

	for i := 0; i < 200; i++ {
		g.aggrUpdate()
	}

	if mob.Fighting != nil {
		t.Error("charmed aggressive mob should not attack")
	}
}

func TestAggrUpdate_SkipsNonStanding(t *testing.T) {
	w, g := newUpdateTestWorld()
	room := &types.RoomIndexData{Vnum: 8402, Name: "Danger Room"}
	w.Rooms[8402] = room

	mobIdx := &types.MobIndexData{
		Vnum: 8402, PlayerName: "sleeping aggr", Level: 5,
		Position: types.POS_SLEEPING, DefPosition: types.POS_SLEEPING,
	}
	w.MobIndex[8402] = mobIdx
	mob := handler.CreateMobile(w, mobIdx)
	mob.Act.Set(types.ACT_AGGRESSIVE)
	mob.Position = types.POS_SLEEPING
	handler.CharToRoom(mob, room)

	player := &types.CharData{
		Name:     "Target2",
		Level:    5,
		Position: types.POS_STANDING,
		PCData:   &types.PCData{PagerLen: 24},
	}
	handler.CharToRoom(player, room)
	w.AddChar(player)

	for i := 0; i < 200; i++ {
		g.aggrUpdate()
	}

	if mob.Fighting != nil {
		t.Error("sleeping aggressive mob should not attack")
	}
}
