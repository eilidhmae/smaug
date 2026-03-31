package handler

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func newTestWorld() *world.World {
	return world.New("/tmp/testdata")
}

func newTestMobIndex(vnum, level int) *types.MobIndexData {
	return &types.MobIndexData{
		Vnum:        vnum,
		PlayerName:  "test mob",
		ShortDescr:  "a test mob",
		LongDescr:   "A test mob is here.\n\r",
		Description: "It looks like a test mob.\n\r",
		Level:       level,
		Sex:         types.SEX_MALE,
		Race:        types.RACE_HUMAN,
		Position:    types.POS_STANDING,
		DefPosition: types.POS_STANDING,
		HitNoDice:   5,
		HitSizeDice: 10,
		HitPlus:     50,
		DamNoDice:   2,
		DamSizeDice: 6,
		Gold:        100,
		PermStr:     15,
		PermInt:     12,
		PermWis:     12,
		PermDex:     14,
		PermCon:     15,
		PermCha:     10,
		PermLck:     13,
	}
}

func newTestObjIndex(vnum int) *types.ObjIndexData {
	return &types.ObjIndexData{
		Vnum:        vnum,
		Name:        "test sword",
		ShortDescr:  "a test sword",
		Description: "A test sword lies here.",
		ItemType:    types.ITEM_WEAPON,
		Weight:      5,
		GoldCost:    100,
		Value:       [6]int{12, 4, 8, 0, 0, 0},
	}
}

func newTestRoom(vnum int) *types.RoomIndexData {
	return &types.RoomIndexData{
		Vnum:        vnum,
		Name:        "Test Room",
		Description: "A test room.\n\r",
		SectorType:  types.SECT_INSIDE,
	}
}

func TestCreateMobile(t *testing.T) {
	w := newTestWorld()
	idx := newTestMobIndex(1000, 10)
	idx.Act.Set(types.ACT_IS_NPC)

	mob := CreateMobile(w, idx)

	if mob.Name != "test mob" {
		t.Errorf("Name = %q, want %q", mob.Name, "test mob")
	}
	if !mob.Act.IsSet(types.ACT_IS_NPC) {
		t.Error("ACT_IS_NPC not set")
	}
	if mob.IndexData != idx {
		t.Error("IndexData not set")
	}
	if mob.MaxHit < 1 {
		t.Errorf("MaxHit = %d, want > 0", mob.MaxHit)
	}
	if mob.Hit != mob.MaxHit {
		t.Errorf("Hit = %d, want MaxHit = %d", mob.Hit, mob.MaxHit)
	}
	if mob.PermStr != 15 {
		t.Errorf("PermStr = %d, want 15", mob.PermStr)
	}
	if mob.BareNumDie != 2 {
		t.Errorf("BareNumDie = %d, want 2", mob.BareNumDie)
	}
	if mob.BareSizeDie != 6 {
		t.Errorf("BareSizeDie = %d, want 6", mob.BareSizeDie)
	}
	if idx.Count != 1 {
		t.Errorf("idx.Count = %d, want 1", idx.Count)
	}
	// Should be in world character list
	if len(w.Characters) != 1 || w.Characters[0] != mob {
		t.Error("mob not added to world Characters")
	}
}

func TestCreateMobile_HitDice(t *testing.T) {
	w := newTestWorld()
	idx := newTestMobIndex(1001, 5)
	idx.HitNoDice = 3
	idx.HitSizeDice = 8
	idx.HitPlus = 20

	mob := CreateMobile(w, idx)

	// HP should be between HitPlus+HitNoDice (23) and HitNoDice*HitSizeDice+HitPlus (44)
	if mob.MaxHit < 23 || mob.MaxHit > 44 {
		t.Errorf("MaxHit = %d, expected range [23, 44]", mob.MaxHit)
	}
}

func TestCreateMobile_NoDice(t *testing.T) {
	w := newTestWorld()
	idx := newTestMobIndex(1002, 10)
	idx.HitNoDice = 0
	idx.HitSizeDice = 0

	mob := CreateMobile(w, idx)

	// With no hit dice, HP is level*8 + random (level*level/4 to level*level)
	// level=10 (fuzzy, so 9-11): min ~72 + 20 = 92, max ~88 + 121 = 209
	if mob.MaxHit < 1 {
		t.Errorf("MaxHit = %d, want > 0", mob.MaxHit)
	}
}

func TestCreateObject(t *testing.T) {
	w := newTestWorld()
	idx := newTestObjIndex(2000)

	obj := CreateObject(w, idx, 5)

	if obj.Name != "test sword" {
		t.Errorf("Name = %q, want %q", obj.Name, "test sword")
	}
	if obj.IndexData != idx {
		t.Error("IndexData not set")
	}
	if obj.Level != 5 {
		t.Errorf("Level = %d, want 5", obj.Level)
	}
	if obj.WearLoc != types.WEAR_NONE {
		t.Errorf("WearLoc = %d, want WEAR_NONE", obj.WearLoc)
	}
	if obj.ItemType != types.ITEM_WEAPON {
		t.Errorf("ItemType = %d, want ITEM_WEAPON", obj.ItemType)
	}
	if idx.Count != 1 {
		t.Errorf("idx.Count = %d, want 1", idx.Count)
	}
	if len(w.Objects) != 1 || w.Objects[0] != obj {
		t.Error("obj not added to world Objects")
	}
}

func TestCreateObject_ExtraDescrs(t *testing.T) {
	w := newTestWorld()
	idx := newTestObjIndex(2001)
	idx.ExtraDescr = []*types.ExtraDescrData{
		{Keyword: "runes", Description: "Ancient runes glow faintly."},
	}

	obj := CreateObject(w, idx, 1)

	if len(obj.ExtraDescr) != 1 {
		t.Fatalf("ExtraDescr count = %d, want 1", len(obj.ExtraDescr))
	}
	if obj.ExtraDescr[0].Keyword != "runes" {
		t.Errorf("ExtraDescr keyword = %q, want %q", obj.ExtraDescr[0].Keyword, "runes")
	}
	// Verify it's a copy, not the same pointer
	if obj.ExtraDescr[0] == idx.ExtraDescr[0] {
		t.Error("ExtraDescr should be a copy, not same pointer")
	}
}

func TestCharToRoom(t *testing.T) {
	room := newTestRoom(3000)
	ch := &types.CharData{Name: "Tester"}

	CharToRoom(ch, room)

	if ch.InRoom != room {
		t.Error("ch.InRoom not set")
	}
	if len(room.People) != 1 || room.People[0] != ch {
		t.Error("ch not in room.People")
	}
}

func TestCharFromRoom(t *testing.T) {
	room := newTestRoom(3001)
	ch := &types.CharData{Name: "Tester"}
	CharToRoom(ch, room)

	CharFromRoom(ch)

	if ch.InRoom != nil {
		t.Error("ch.InRoom should be nil after CharFromRoom")
	}
	if ch.WasInRoom != room {
		t.Error("ch.WasInRoom should be set to old room")
	}
	if len(room.People) != 0 {
		t.Errorf("room.People should be empty, has %d", len(room.People))
	}
}

func TestObjToRoom(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(3002)
	idx := newTestObjIndex(2002)
	obj := CreateObject(w, idx, 1)

	ObjToRoom(obj, room)

	if obj.InRoom != room {
		t.Error("obj.InRoom not set")
	}
	if len(room.Contents) != 1 || room.Contents[0] != obj {
		t.Error("obj not in room.Contents")
	}
}

func TestObjToChar(t *testing.T) {
	w := newTestWorld()
	ch := &types.CharData{Name: "Tester"}
	idx := newTestObjIndex(2003)
	obj := CreateObject(w, idx, 1)

	ObjToChar(obj, ch)

	if obj.CarriedBy != ch {
		t.Error("obj.CarriedBy not set")
	}
	if len(ch.Carrying) != 1 || ch.Carrying[0] != obj {
		t.Error("obj not in ch.Carrying")
	}
}

func TestObjToObj(t *testing.T) {
	w := newTestWorld()
	container := CreateObject(w, newTestObjIndex(2004), 1)
	item := CreateObject(w, newTestObjIndex(2005), 1)

	ObjToObj(item, container)

	if item.InObj != container {
		t.Error("item.InObj not set")
	}
	if len(container.Contents) != 1 || container.Contents[0] != item {
		t.Error("item not in container.Contents")
	}
}

func TestEquipChar(t *testing.T) {
	w := newTestWorld()
	ch := &types.CharData{Name: "Tester"}
	idx := newTestObjIndex(2006)
	obj := CreateObject(w, idx, 1)

	EquipChar(ch, obj, types.WEAR_WIELD)

	if obj.WearLoc != types.WEAR_WIELD {
		t.Errorf("obj.WearLoc = %d, want WEAR_WIELD", obj.WearLoc)
	}
	if obj.CarriedBy != ch {
		t.Error("obj.CarriedBy not set")
	}
	if len(ch.Carrying) != 1 {
		t.Error("obj not in ch.Carrying")
	}
}

func TestResetArea_MobSpawn(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(5000)
	w.Rooms[5000] = room
	idx := newTestMobIndex(1010, 5)
	w.MobIndex[1010] = idx

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'M', Arg1: 1010, Arg2: 1, Arg3: 5000},
		},
	}

	ResetArea(w, area)

	if len(room.People) != 1 {
		t.Fatalf("room.People = %d, want 1", len(room.People))
	}
	if room.People[0].Name != "test mob" {
		t.Errorf("mob name = %q, want %q", room.People[0].Name, "test mob")
	}
}

func TestResetArea_MobMaxCount(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(5001)
	w.Rooms[5001] = room
	idx := newTestMobIndex(1011, 5)
	w.MobIndex[1011] = idx

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'M', Arg1: 1011, Arg2: 1, Arg3: 5001},
			{Command: 'M', Arg1: 1011, Arg2: 1, Arg3: 5001}, // Should be skipped (max=1)
		},
	}

	ResetArea(w, area)

	if len(room.People) != 1 {
		t.Errorf("room.People = %d, want 1 (second reset should be skipped)", len(room.People))
	}
}

func TestResetArea_ObjectInRoom(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(5002)
	w.Rooms[5002] = room
	idx := newTestObjIndex(2010)
	w.ObjIndex[2010] = idx

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'O', Arg1: 2010, Arg3: 5002},
		},
	}

	ResetArea(w, area)

	if len(room.Contents) != 1 {
		t.Fatalf("room.Contents = %d, want 1", len(room.Contents))
	}
	if room.Contents[0].Name != "test sword" {
		t.Errorf("obj name = %q, want %q", room.Contents[0].Name, "test sword")
	}
	// Room objects should have zero cost
	if room.Contents[0].GoldCost != 0 {
		t.Errorf("GoldCost = %d, want 0 for room object", room.Contents[0].GoldCost)
	}
}

func TestResetArea_GiveToMob(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(5003)
	w.Rooms[5003] = room
	mobIdx := newTestMobIndex(1012, 5)
	w.MobIndex[1012] = mobIdx
	objIdx := newTestObjIndex(2011)
	w.ObjIndex[2011] = objIdx

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'M', Arg1: 1012, Arg2: 1, Arg3: 5003},
			{Command: 'G', Arg1: 2011},
		},
	}

	ResetArea(w, area)

	if len(room.People) != 1 {
		t.Fatal("mob not spawned")
	}
	mob := room.People[0]
	if len(mob.Carrying) != 1 {
		t.Fatalf("mob.Carrying = %d, want 1", len(mob.Carrying))
	}
	if mob.Carrying[0].Name != "test sword" {
		t.Errorf("obj name = %q, want %q", mob.Carrying[0].Name, "test sword")
	}
}

func TestResetArea_EquipOnMob(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(5004)
	w.Rooms[5004] = room
	mobIdx := newTestMobIndex(1013, 5)
	w.MobIndex[1013] = mobIdx
	objIdx := newTestObjIndex(2012)
	w.ObjIndex[2012] = objIdx

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'M', Arg1: 1013, Arg2: 1, Arg3: 5004},
			{Command: 'E', Arg1: 2012, Arg3: types.WEAR_WIELD},
		},
	}

	ResetArea(w, area)

	mob := room.People[0]
	if len(mob.Carrying) != 1 {
		t.Fatalf("mob.Carrying = %d, want 1", len(mob.Carrying))
	}
	if mob.Carrying[0].WearLoc != types.WEAR_WIELD {
		t.Errorf("WearLoc = %d, want WEAR_WIELD", mob.Carrying[0].WearLoc)
	}
}

func TestResetArea_DoorState(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(5005)
	room.Exits = []*types.ExitData{
		{Direction: types.DIR_NORTH, ExitInfo: 0, ToRoom: newTestRoom(5006)},
	}
	w.Rooms[5005] = room

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'D', Arg1: 5005, Arg2: types.DIR_NORTH, Arg3: 2}, // Locked
		},
	}

	ResetArea(w, area)

	exit := room.Exits[0]
	if exit.ExitInfo&int(types.EX_CLOSED) == 0 {
		t.Error("door should be closed")
	}
	if exit.ExitInfo&int(types.EX_LOCKED) == 0 {
		t.Error("door should be locked")
	}
}

func TestInterpolate(t *testing.T) {
	// Level 0 should return low
	if v := interpolate(0, 100, -100); v != 100 {
		t.Errorf("interpolate(0, 100, -100) = %d, want 100", v)
	}
	// Level LEVEL_AVATAR should return high
	if v := interpolate(types.LEVEL_AVATAR, 100, -100); v != -100 {
		t.Errorf("interpolate(LEVEL_AVATAR, 100, -100) = %d, want -100", v)
	}
}
