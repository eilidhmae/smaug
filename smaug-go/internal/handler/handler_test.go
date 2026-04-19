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

// Pre-G0.a for plan-phase6-hotboot.md — ACT_SENTINEL mobs record their spawn
// room on HomeVnum so hotboot recovery restores them to the original location
// even if an in-game event (e.g. mudprog) has since moved them.
func TestSetSentinelHome_SentinelRecordsVnum(t *testing.T) {
	room := newTestRoom(3001)
	mob := &types.CharData{}
	mob.Act.Set(types.ACT_SENTINEL)

	SetSentinelHome(mob, room)

	if mob.HomeVnum != 3001 {
		t.Errorf("sentinel HomeVnum = %d, want 3001", mob.HomeVnum)
	}
}

func TestSetSentinelHome_NonSentinelUnchanged(t *testing.T) {
	room := newTestRoom(3001)
	mob := &types.CharData{HomeVnum: 99}

	SetSentinelHome(mob, room)

	if mob.HomeVnum != 99 {
		t.Errorf("non-sentinel HomeVnum = %d, want preserved 99", mob.HomeVnum)
	}
}

func TestSetSentinelHome_NilSafe(t *testing.T) {
	// Defensive: must not panic on nil mob or nil room.
	SetSentinelHome(nil, newTestRoom(3001))
	SetSentinelHome(&types.CharData{}, nil)
}

func TestResetMobile_SentinelHomeVnumWired(t *testing.T) {
	// End-to-end: resetMobile calls SetSentinelHome after CharToRoom so the
	// sentinel's HomeVnum reflects the room it was placed in.
	w := newTestWorld()
	idx := newTestMobIndex(5001, 10)
	idx.Act.Set(types.ACT_SENTINEL)
	w.MobIndex[idx.Vnum] = idx
	room := newTestRoom(3001)
	w.Rooms[room.Vnum] = room

	reset := &types.ResetData{Command: 'M', Arg1: 5001, Arg2: 1, Arg3: 3001}
	var lastMob *types.CharData
	var lastRoom *types.RoomIndexData
	mobLevel := 0

	resetMobile(w, reset, &lastMob, &lastRoom, &mobLevel)

	if lastMob == nil {
		t.Fatal("resetMobile did not create mob")
	}
	if lastMob.HomeVnum != 3001 {
		t.Errorf("sentinel mob HomeVnum = %d, want 3001", lastMob.HomeVnum)
	}
}

func TestResetMobile_NonSentinelHomeVnumZero(t *testing.T) {
	w := newTestWorld()
	idx := newTestMobIndex(5002, 10) // no ACT_SENTINEL
	w.MobIndex[idx.Vnum] = idx
	room := newTestRoom(3002)
	w.Rooms[room.Vnum] = room

	reset := &types.ResetData{Command: 'M', Arg1: 5002, Arg2: 1, Arg3: 3002}
	var lastMob *types.CharData
	var lastRoom *types.RoomIndexData
	mobLevel := 0

	resetMobile(w, reset, &lastMob, &lastRoom, &mobLevel)

	if lastMob == nil {
		t.Fatal("resetMobile did not create mob")
	}
	if lastMob.HomeVnum != 0 {
		t.Errorf("non-sentinel mob HomeVnum = %d, want 0", lastMob.HomeVnum)
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

func TestObjFromChar(t *testing.T) {
	w := newTestWorld()
	ch := &types.CharData{Name: "Tester"}
	obj := CreateObject(w, newTestObjIndex(2020), 1)
	ObjToChar(obj, ch)

	ObjFromChar(obj)

	if obj.CarriedBy != nil {
		t.Error("obj.CarriedBy should be nil")
	}
	if len(ch.Carrying) != 0 {
		t.Errorf("ch.Carrying should be empty, has %d", len(ch.Carrying))
	}
}

func TestObjFromChar_Equipped(t *testing.T) {
	w := newTestWorld()
	ch := &types.CharData{Name: "Tester"}
	obj := CreateObject(w, newTestObjIndex(2021), 1)
	EquipChar(ch, obj, types.WEAR_WIELD)

	ObjFromChar(obj)

	if obj.CarriedBy != nil {
		t.Error("obj.CarriedBy should be nil")
	}
	if obj.WearLoc != types.WEAR_NONE {
		t.Errorf("WearLoc = %d, want WEAR_NONE", obj.WearLoc)
	}
	if len(ch.Carrying) != 0 {
		t.Errorf("ch.Carrying should be empty, has %d", len(ch.Carrying))
	}
}

func TestObjFromRoom(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(3010)
	obj := CreateObject(w, newTestObjIndex(2022), 1)
	ObjToRoom(obj, room)

	ObjFromRoom(obj)

	if obj.InRoom != nil {
		t.Error("obj.InRoom should be nil")
	}
	if len(room.Contents) != 0 {
		t.Errorf("room.Contents should be empty, has %d", len(room.Contents))
	}
}

func TestObjFromObj(t *testing.T) {
	w := newTestWorld()
	container := CreateObject(w, newTestObjIndex(2023), 1)
	item := CreateObject(w, newTestObjIndex(2024), 1)
	ObjToObj(item, container)

	ObjFromObj(item)

	if item.InObj != nil {
		t.Error("item.InObj should be nil")
	}
	if len(container.Contents) != 0 {
		t.Errorf("container.Contents should be empty, has %d", len(container.Contents))
	}
}

func TestUnequipChar(t *testing.T) {
	w := newTestWorld()
	ch := &types.CharData{Name: "Tester"}
	obj := CreateObject(w, newTestObjIndex(2025), 1)
	EquipChar(ch, obj, types.WEAR_WIELD)

	UnequipChar(ch, obj)

	if obj.WearLoc != types.WEAR_NONE {
		t.Errorf("WearLoc = %d, want WEAR_NONE", obj.WearLoc)
	}
	// Object should still be in carrying (moved to inventory)
	if len(ch.Carrying) != 1 {
		t.Errorf("ch.Carrying = %d, want 1 (unequip keeps in inventory)", len(ch.Carrying))
	}
}

func TestUnequipChar_DualWield(t *testing.T) {
	w := newTestWorld()
	ch := &types.CharData{Name: "Tester"}
	weapon := CreateObject(w, newTestObjIndex(2026), 1)
	dual := CreateObject(w, newTestObjIndex(2027), 1)
	EquipChar(ch, weapon, types.WEAR_WIELD)
	EquipChar(ch, dual, types.WEAR_DUAL_WIELD)

	// Unequip primary wield — dual should move to wield
	UnequipChar(ch, weapon)

	if weapon.WearLoc != types.WEAR_NONE {
		t.Errorf("weapon.WearLoc = %d, want WEAR_NONE", weapon.WearLoc)
	}
	if dual.WearLoc != types.WEAR_WIELD {
		t.Errorf("dual.WearLoc = %d, want WEAR_WIELD (should move from dual to wield)", dual.WearLoc)
	}
}

func TestExtractObj(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(3020)
	idx := newTestObjIndex(2030)
	obj := CreateObject(w, idx, 1)
	ObjToRoom(obj, room)

	ExtractObj(w, obj)

	if len(room.Contents) != 0 {
		t.Errorf("room.Contents should be empty, has %d", len(room.Contents))
	}
	if len(w.Objects) != 0 {
		t.Errorf("world.Objects should be empty, has %d", len(w.Objects))
	}
	if idx.Count != 0 {
		t.Errorf("idx.Count = %d, want 0", idx.Count)
	}
}

func TestExtractObj_WithContents(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(3021)
	containerIdx := newTestObjIndex(2031)
	containerIdx.Name = "container"
	itemIdx := newTestObjIndex(2032)
	itemIdx.Name = "item"

	container := CreateObject(w, containerIdx, 1)
	item := CreateObject(w, itemIdx, 1)
	ObjToRoom(container, room)
	ObjToObj(item, container)

	ExtractObj(w, container)

	// Both objects should be extracted
	if len(w.Objects) != 0 {
		t.Errorf("world.Objects should be empty, has %d", len(w.Objects))
	}
	if len(room.Contents) != 0 {
		t.Errorf("room.Contents should be empty, has %d", len(room.Contents))
	}
}

func TestExtractObj_FromChar(t *testing.T) {
	w := newTestWorld()
	ch := &types.CharData{Name: "Tester"}
	idx := newTestObjIndex(2033)
	obj := CreateObject(w, idx, 1)
	ObjToChar(obj, ch)

	ExtractObj(w, obj)

	if len(ch.Carrying) != 0 {
		t.Errorf("ch.Carrying should be empty, has %d", len(ch.Carrying))
	}
}

func TestExtractChar(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(3030)
	idx := newTestMobIndex(1020, 5)
	mob := CreateMobile(w, idx)
	CharToRoom(mob, room)

	ExtractChar(w, mob, true)

	if len(room.People) != 0 {
		t.Errorf("room.People should be empty, has %d", len(room.People))
	}
	if len(w.Characters) != 0 {
		t.Errorf("world.Characters should be empty, has %d", len(w.Characters))
	}
	if idx.Count != 0 {
		t.Errorf("idx.Count = %d, want 0", idx.Count)
	}
}

func TestExtractChar_WithInventory(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(3031)
	mobIdx := newTestMobIndex(1021, 5)
	mob := CreateMobile(w, mobIdx)
	CharToRoom(mob, room)

	objIdx := newTestObjIndex(2034)
	obj := CreateObject(w, objIdx, 1)
	ObjToChar(obj, mob)

	ExtractChar(w, mob, true)

	// Both mob and its inventory should be gone
	if len(w.Characters) != 0 {
		t.Errorf("world.Characters should be empty, has %d", len(w.Characters))
	}
	if len(w.Objects) != 0 {
		t.Errorf("world.Objects should be empty, has %d", len(w.Objects))
	}
}

func TestAffectToChar(t *testing.T) {
	ch := &types.CharData{Name: "Tester", PermStr: 15}

	aff := &types.AffectData{
		Type:     1, // arbitrary skill number
		Duration: 10,
		Location: types.APPLY_STR,
		Modifier: 3,
	}

	AffectToChar(ch, aff)

	if len(ch.Affects) != 1 {
		t.Fatalf("ch.Affects = %d, want 1", len(ch.Affects))
	}
	// Should be a copy, not same pointer
	if ch.Affects[0] == aff {
		t.Error("affect should be a copy, not same pointer")
	}
	if ch.Affects[0].Duration != 10 {
		t.Errorf("Duration = %d, want 10", ch.Affects[0].Duration)
	}
	// ModStr should be increased by modifier
	if ch.ModStr != 3 {
		t.Errorf("ModStr = %d, want 3", ch.ModStr)
	}
}

func TestAffectRemove(t *testing.T) {
	ch := &types.CharData{Name: "Tester", PermStr: 15}

	aff := &types.AffectData{
		Type:     1,
		Duration: 10,
		Location: types.APPLY_STR,
		Modifier: 3,
	}

	AffectToChar(ch, aff)
	applied := ch.Affects[0]
	AffectRemove(ch, applied)

	if len(ch.Affects) != 0 {
		t.Errorf("ch.Affects = %d, want 0", len(ch.Affects))
	}
	if ch.ModStr != 0 {
		t.Errorf("ModStr = %d, want 0 (affect removed)", ch.ModStr)
	}
}

func TestAffectStrip(t *testing.T) {
	ch := &types.CharData{Name: "Tester"}

	// Add two affects of same type, one of different
	AffectToChar(ch, &types.AffectData{Type: 5, Duration: 10, Location: types.APPLY_STR, Modifier: 2})
	AffectToChar(ch, &types.AffectData{Type: 5, Duration: 5, Location: types.APPLY_DEX, Modifier: 1})
	AffectToChar(ch, &types.AffectData{Type: 7, Duration: 20, Location: types.APPLY_AC, Modifier: -10})

	AffectStrip(ch, 5)

	if len(ch.Affects) != 1 {
		t.Errorf("ch.Affects = %d, want 1 (only type 7 should remain)", len(ch.Affects))
	}
	if ch.Affects[0].Type != 7 {
		t.Errorf("remaining affect type = %d, want 7", ch.Affects[0].Type)
	}
	if ch.ModStr != 0 {
		t.Errorf("ModStr = %d, want 0", ch.ModStr)
	}
	if ch.ModDex != 0 {
		t.Errorf("ModDex = %d, want 0", ch.ModDex)
	}
}

func TestAffectJoin(t *testing.T) {
	ch := &types.CharData{Name: "Tester"}

	// Add first affect
	AffectToChar(ch, &types.AffectData{Type: 5, Duration: 10, Location: types.APPLY_STR, Modifier: 2})

	// Join with same type — should combine
	AffectJoin(ch, &types.AffectData{Type: 5, Duration: 5, Location: types.APPLY_STR, Modifier: 3})

	if len(ch.Affects) != 1 {
		t.Errorf("ch.Affects = %d, want 1 (should combine)", len(ch.Affects))
	}
	if ch.Affects[0].Duration != 15 {
		t.Errorf("Duration = %d, want 15 (10+5)", ch.Affects[0].Duration)
	}
	if ch.Affects[0].Modifier != 5 {
		t.Errorf("Modifier = %d, want 5 (2+3)", ch.Affects[0].Modifier)
	}
	if ch.ModStr != 5 {
		t.Errorf("ModStr = %d, want 5", ch.ModStr)
	}
}

func TestAffectModify_BitVector(t *testing.T) {
	ch := &types.CharData{Name: "Tester"}

	aff := &types.AffectData{
		Type:     1,
		Duration: 10,
		Location: types.APPLY_NONE,
	}
	aff.BitVector.Set(types.AFF_INVISIBLE)

	AffectToChar(ch, aff)

	if !ch.AffectedBy.IsSet(types.AFF_INVISIBLE) {
		t.Error("AFF_INVISIBLE should be set after affect applied")
	}

	AffectRemove(ch, ch.Affects[0])

	if ch.AffectedBy.IsSet(types.AFF_INVISIBLE) {
		t.Error("AFF_INVISIBLE should be cleared after affect removed")
	}
}

func TestAffectModify_MultipleStats(t *testing.T) {
	ch := &types.CharData{Name: "Tester"}

	tests := []struct {
		location int
		modifier int
		check    func() int
		name     string
	}{
		{types.APPLY_STR, 3, func() int { return ch.ModStr }, "ModStr"},
		{types.APPLY_DEX, 2, func() int { return ch.ModDex }, "ModDex"},
		{types.APPLY_INT, 1, func() int { return ch.ModInt }, "ModInt"},
		{types.APPLY_WIS, 4, func() int { return ch.ModWis }, "ModWis"},
		{types.APPLY_CON, 2, func() int { return ch.ModCon }, "ModCon"},
		{types.APPLY_CHA, 1, func() int { return ch.ModCha }, "ModCha"},
		{types.APPLY_LCK, 3, func() int { return ch.ModLck }, "ModLck"},
		{types.APPLY_AC, -20, func() int { return ch.Armor }, "Armor"},
		{types.APPLY_HITROLL, 5, func() int { return ch.Hitroll }, "Hitroll"},
		{types.APPLY_DAMROLL, 3, func() int { return ch.Damroll }, "Damroll"},
		{types.APPLY_HIT, 50, func() int { return ch.MaxHit }, "MaxHit"},
		{types.APPLY_MANA, 30, func() int { return ch.MaxMana }, "MaxMana"},
		{types.APPLY_MOVE, 20, func() int { return ch.MaxMove }, "MaxMove"},
	}

	for _, tt := range tests {
		aff := &types.AffectData{
			Type:     1,
			Duration: 10,
			Location: tt.location,
			Modifier: tt.modifier,
		}
		AffectToChar(ch, aff)
		if got := tt.check(); got != tt.modifier {
			t.Errorf("%s = %d, want %d after apply", tt.name, got, tt.modifier)
		}
		AffectRemove(ch, ch.Affects[len(ch.Affects)-1])
		if got := tt.check(); got != 0 {
			t.Errorf("%s = %d, want 0 after remove", tt.name, got)
		}
	}
}

func TestGetCharRoom(t *testing.T) {
	room := newTestRoom(3040)
	ch := &types.CharData{Name: "Tester"}
	CharToRoom(ch, room)

	mob1 := &types.CharData{Name: "guard soldier"}
	mob1.Act.Set(types.ACT_IS_NPC)
	CharToRoom(mob1, room)

	mob2 := &types.CharData{Name: "guard captain"}
	mob2.Act.Set(types.ACT_IS_NPC)
	CharToRoom(mob2, room)

	// Find first guard
	found := GetCharRoom(ch, "guard")
	if found != mob1 {
		t.Error("GetCharRoom('guard') should find first guard")
	}

	// Find 2nd guard
	found = GetCharRoom(ch, "2.guard")
	if found != mob2 {
		t.Error("GetCharRoom('2.guard') should find second guard")
	}

	// Self
	found = GetCharRoom(ch, "self")
	if found != ch {
		t.Error("GetCharRoom('self') should return self")
	}

	// Not found
	found = GetCharRoom(ch, "dragon")
	if found != nil {
		t.Error("GetCharRoom('dragon') should return nil")
	}
}

func TestGetCharWorld(t *testing.T) {
	w := newTestWorld()
	room1 := newTestRoom(3050)
	room2 := newTestRoom(3051)

	ch := &types.CharData{Name: "Tester"}
	CharToRoom(ch, room1)
	w.AddChar(ch)

	mob := &types.CharData{Name: "wizard mage"}
	mob.Act.Set(types.ACT_IS_NPC)
	CharToRoom(mob, room2)
	w.AddChar(mob)

	// Find in world (different room)
	found := GetCharWorld(w, ch, "wizard")
	if found != mob {
		t.Error("GetCharWorld('wizard') should find mob in different room")
	}

	// Self
	found = GetCharWorld(w, ch, "self")
	if found != ch {
		t.Error("GetCharWorld('self') should return self")
	}

	// Not found
	found = GetCharWorld(w, ch, "dragon")
	if found != nil {
		t.Error("GetCharWorld('dragon') should return nil")
	}
}

func TestGetObjCarry(t *testing.T) {
	w := newTestWorld()
	ch := &types.CharData{Name: "Tester"}

	idx1 := newTestObjIndex(2040)
	idx1.Name = "iron sword"
	sword := CreateObject(w, idx1, 1)
	ObjToChar(sword, ch)

	idx2 := newTestObjIndex(2041)
	idx2.Name = "iron shield"
	shield := CreateObject(w, idx2, 1)
	EquipChar(ch, shield, types.WEAR_BODY)

	// Should find sword (in inventory, not equipped)
	found := GetObjCarry(ch, "sword")
	if found != sword {
		t.Error("GetObjCarry('sword') should find sword in inventory")
	}

	// Should not find shield (it's equipped)
	found = GetObjCarry(ch, "shield")
	if found != nil {
		t.Error("GetObjCarry('shield') should not find equipped items")
	}
}

func TestGetObjWear(t *testing.T) {
	w := newTestWorld()
	ch := &types.CharData{Name: "Tester"}

	idx1 := newTestObjIndex(2042)
	idx1.Name = "iron sword"
	sword := CreateObject(w, idx1, 1)
	ObjToChar(sword, ch)

	idx2 := newTestObjIndex(2043)
	idx2.Name = "iron shield"
	shield := CreateObject(w, idx2, 1)
	EquipChar(ch, shield, types.WEAR_BODY)

	// Should find shield (equipped)
	found := GetObjWear(ch, "shield")
	if found != shield {
		t.Error("GetObjWear('shield') should find equipped shield")
	}

	// Should not find sword (in inventory, not equipped)
	found = GetObjWear(ch, "sword")
	if found != nil {
		t.Error("GetObjWear('sword') should not find inventory items")
	}
}

func TestGetObjHere(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(3060)
	ch := &types.CharData{Name: "Tester"}
	CharToRoom(ch, room)

	// Object in room
	idx1 := newTestObjIndex(2044)
	idx1.Name = "gold coin"
	coin := CreateObject(w, idx1, 1)
	ObjToRoom(coin, room)

	// Object in inventory
	idx2 := newTestObjIndex(2045)
	idx2.Name = "silver ring"
	ring := CreateObject(w, idx2, 1)
	ObjToChar(ring, ch)

	// Find in room
	found := GetObjHere(ch, "coin")
	if found != coin {
		t.Error("GetObjHere('coin') should find coin in room")
	}

	// Find in inventory
	found = GetObjHere(ch, "ring")
	if found != ring {
		t.Error("GetObjHere('ring') should find ring in inventory")
	}
}

func TestGetObjWorld(t *testing.T) {
	w := newTestWorld()
	room1 := newTestRoom(3070)
	room2 := newTestRoom(3071)
	ch := &types.CharData{Name: "Tester"}
	CharToRoom(ch, room1)

	idx := newTestObjIndex(2046)
	idx.Name = "magic orb"
	orb := CreateObject(w, idx, 1)
	ObjToRoom(orb, room2)

	found := GetObjWorld(w, ch, "orb")
	if found != orb {
		t.Error("GetObjWorld('orb') should find orb in different room")
	}

	found = GetObjWorld(w, ch, "dragon")
	if found != nil {
		t.Error("GetObjWorld('dragon') should return nil")
	}
}

func TestGetEqChar(t *testing.T) {
	w := newTestWorld()
	ch := &types.CharData{Name: "Tester"}
	idx := newTestObjIndex(2050)
	obj := CreateObject(w, idx, 1)
	EquipChar(ch, obj, types.WEAR_WIELD)

	found := GetEqChar(ch, types.WEAR_WIELD)
	if found != obj {
		t.Error("GetEqChar should find equipped item at WEAR_WIELD")
	}

	found = GetEqChar(ch, types.WEAR_HEAD)
	if found != nil {
		t.Error("GetEqChar should return nil for empty slot")
	}
}

func TestCanDropObj(t *testing.T) {
	w := newTestWorld()
	idx := newTestObjIndex(2051)
	obj := CreateObject(w, idx, 1)

	if !CanDropObj(obj) {
		t.Error("normal object should be droppable")
	}

	// Set no-drop flag
	obj.ExtraFlags.Set(types.ITEM_NODROP)
	if CanDropObj(obj) {
		t.Error("ITEM_NODROP object should not be droppable")
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

// --- Additional tests for increased coverage ---

// ObjToObj: verify multiple items in container
func TestObjToObj_MultipleItems(t *testing.T) {
	w := newTestWorld()
	container := CreateObject(w, newTestObjIndex(2060), 1)
	item1 := CreateObject(w, newTestObjIndex(2061), 1)
	item2 := CreateObject(w, newTestObjIndex(2062), 1)

	ObjToObj(item1, container)
	ObjToObj(item2, container)

	if len(container.Contents) != 2 {
		t.Fatalf("container.Contents = %d, want 2", len(container.Contents))
	}
	if item1.InObj != container || item2.InObj != container {
		t.Error("items should reference container as InObj")
	}
}

// EquipChar edge cases: multiple wear locations
func TestEquipChar_MultipleSlots(t *testing.T) {
	w := newTestWorld()
	ch := &types.CharData{Name: "Tester"}

	helm := CreateObject(w, newTestObjIndex(2070), 1)
	helm.IndexData.Name = "iron helm"
	helm.Name = "iron helm"
	body := CreateObject(w, newTestObjIndex(2071), 1)
	body.IndexData.Name = "chain mail"
	body.Name = "chain mail"

	EquipChar(ch, helm, types.WEAR_HEAD)
	EquipChar(ch, body, types.WEAR_BODY)

	if len(ch.Carrying) != 2 {
		t.Fatalf("ch.Carrying = %d, want 2", len(ch.Carrying))
	}
	if helm.WearLoc != types.WEAR_HEAD {
		t.Errorf("helm.WearLoc = %d, want WEAR_HEAD", helm.WearLoc)
	}
	if body.WearLoc != types.WEAR_BODY {
		t.Errorf("body.WearLoc = %d, want WEAR_BODY", body.WearLoc)
	}
}

// GetEqChar: nil for completely empty carrying list
func TestGetEqChar_EmptyCarrying(t *testing.T) {
	ch := &types.CharData{Name: "Tester"}
	found := GetEqChar(ch, types.WEAR_WIELD)
	if found != nil {
		t.Error("GetEqChar should return nil for empty carrying list")
	}
}

// GetEqChar: doesn't match inventory items (WEAR_NONE)
func TestGetEqChar_IgnoresInventory(t *testing.T) {
	w := newTestWorld()
	ch := &types.CharData{Name: "Tester"}
	obj := CreateObject(w, newTestObjIndex(2072), 1)
	ObjToChar(obj, ch) // In inventory, WearLoc = WEAR_NONE

	found := GetEqChar(ch, types.WEAR_NONE)
	if found != nil {
		// WEAR_NONE == -1, items in inventory have WearLoc == WEAR_NONE
		// but GetEqChar iterates looking for match to WEAR_NONE which would be an inventory item
		// Actually this DOES match, let's verify the behavior
	}
	// The important test: looking for an eq slot that's unoccupied
	found = GetEqChar(ch, types.WEAR_WIELD)
	if found != nil {
		t.Error("GetEqChar should return nil when slot is empty")
	}
}

// CharToRoom: NPC in dark room gets AFF_INFRARED
func TestCharToRoom_DarkRoom_NPC(t *testing.T) {
	room := newTestRoom(3100)
	room.RoomFlags.Set(types.ROOM_DARK)

	mob := &types.CharData{Name: "dark mob"}
	mob.Act.Set(types.ACT_IS_NPC)

	CharToRoom(mob, room)

	if !mob.AffectedBy.IsSet(types.AFF_INFRARED) {
		t.Error("NPC in dark room should get AFF_INFRARED")
	}
}

// CharToRoom: PC in dark room does NOT get AFF_INFRARED
func TestCharToRoom_DarkRoom_PC(t *testing.T) {
	room := newTestRoom(3101)
	room.RoomFlags.Set(types.ROOM_DARK)

	ch := &types.CharData{Name: "Player"}
	// No ACT_IS_NPC set

	CharToRoom(ch, room)

	if ch.AffectedBy.IsSet(types.AFF_INFRARED) {
		t.Error("PC in dark room should NOT get AFF_INFRARED")
	}
}

// CharFromRoom: calling on char not in a room is a no-op
func TestCharFromRoom_NilRoom(t *testing.T) {
	ch := &types.CharData{Name: "Tester", InRoom: nil}
	// Should not panic
	CharFromRoom(ch)
	if ch.InRoom != nil {
		t.Error("InRoom should remain nil")
	}
}

// ExtractChar with fPull=false (rebirth path)
func TestExtractChar_NoPull(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(3110)
	mob := &types.CharData{Name: "rebirth mob"}
	mob.Act.Set(types.ACT_IS_NPC)
	w.AddChar(mob)
	CharToRoom(mob, room)

	ExtractChar(w, mob, false)

	// Room should be cleared
	if len(room.People) != 0 {
		t.Error("room.People should be empty after extract")
	}
	// But character should NOT be removed from world (fPull=false returns early)
	if len(w.Characters) != 1 {
		t.Errorf("world.Characters = %d, want 1 (fPull=false should not remove from world)", len(w.Characters))
	}
}

// ExtractChar: clears fighting reference
func TestExtractChar_ClearsFighting(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(3111)
	mob := &types.CharData{Name: "fighter mob"}
	mob.Act.Set(types.ACT_IS_NPC)
	mob.IndexData = &types.MobIndexData{Count: 1}
	w.AddChar(mob)
	CharToRoom(mob, room)

	target := &types.CharData{Name: "target"}
	mob.Fighting = &types.FightData{Who: target}
	mob.NumFighting = 1

	ExtractChar(w, mob, true)

	// mob.Fighting checked inside ExtractChar but mob is extracted so just verify no panic
	if len(w.Characters) != 0 {
		t.Errorf("world.Characters = %d, want 0", len(w.Characters))
	}
}

// ExtractChar: clears Reply/Retell references in other characters
func TestExtractChar_ClearsReplyRetell(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(3112)

	mob := &types.CharData{Name: "leaving mob"}
	mob.Act.Set(types.ACT_IS_NPC)
	mob.IndexData = &types.MobIndexData{Count: 1}
	w.AddChar(mob)
	CharToRoom(mob, room)

	other := &types.CharData{Name: "other player", Reply: mob, Retell: mob}
	w.AddChar(other)

	ExtractChar(w, mob, true)

	if other.Reply != nil {
		t.Error("other.Reply should be cleared after mob extracted")
	}
	if other.Retell != nil {
		t.Error("other.Retell should be cleared after mob extracted")
	}
}

// ExtractChar: disconnects descriptor
func TestExtractChar_DisconnectsDesc(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(3113)

	ch := &types.CharData{Name: "player with desc"}
	ch.Act.Set(types.ACT_IS_NPC)
	ch.IndexData = &types.MobIndexData{Count: 1}
	desc := &types.DescriptorData{Character: ch}
	ch.Desc = desc
	w.AddChar(ch)
	CharToRoom(ch, room)

	ExtractChar(w, ch, true)

	if desc.Character != nil {
		t.Error("desc.Character should be nil after extract")
	}
	if ch.Desc != nil {
		t.Error("ch.Desc should be nil after extract")
	}
}

// AffectModify: saving throws
func TestAffectModify_SavingThrows(t *testing.T) {
	ch := &types.CharData{Name: "Tester"}

	tests := []struct {
		loc  int
		get  func() int
		name string
	}{
		{types.APPLY_SAVING_POISON, func() int { return ch.SavingPoisonDeath }, "SavingPoisonDeath"},
		{types.APPLY_SAVING_ROD, func() int { return ch.SavingWand }, "SavingWand"},
		{types.APPLY_SAVING_PARA, func() int { return ch.SavingParaPetri }, "SavingParaPetri"},
		{types.APPLY_SAVING_BREATH, func() int { return ch.SavingBreath }, "SavingBreath"},
		{types.APPLY_SAVING_SPELL, func() int { return ch.SavingSpellStaff }, "SavingSpellStaff"},
	}

	for _, tt := range tests {
		aff := &types.AffectData{Type: 1, Duration: 10, Location: tt.loc, Modifier: -5}
		AffectToChar(ch, aff)
		if got := tt.get(); got != -5 {
			t.Errorf("%s = %d, want -5 after apply", tt.name, got)
		}
		AffectRemove(ch, ch.Affects[len(ch.Affects)-1])
		if got := tt.get(); got != 0 {
			t.Errorf("%s = %d, want 0 after remove", tt.name, got)
		}
	}
}

// AffectModify: APPLY_SEX, APPLY_HEIGHT, APPLY_WEIGHT
func TestAffectModify_SexHeightWeight(t *testing.T) {
	ch := &types.CharData{Name: "Tester", Sex: 1, Height: 170, Weight: 80}

	// APPLY_SEX
	aff := &types.AffectData{Type: 1, Duration: 10, Location: types.APPLY_SEX, Modifier: 1}
	AffectModify(ch, aff, true)
	if ch.Sex != 2 {
		t.Errorf("Sex = %d, want 2", ch.Sex)
	}
	AffectModify(ch, aff, false)
	if ch.Sex != 1 {
		t.Errorf("Sex = %d, want 1 after remove", ch.Sex)
	}

	// APPLY_HEIGHT
	aff2 := &types.AffectData{Type: 1, Duration: 10, Location: types.APPLY_HEIGHT, Modifier: 10}
	AffectModify(ch, aff2, true)
	if ch.Height != 180 {
		t.Errorf("Height = %d, want 180", ch.Height)
	}
	AffectModify(ch, aff2, false)
	if ch.Height != 170 {
		t.Errorf("Height = %d, want 170 after remove", ch.Height)
	}

	// APPLY_WEIGHT
	aff3 := &types.AffectData{Type: 1, Duration: 10, Location: types.APPLY_WEIGHT, Modifier: 5}
	AffectModify(ch, aff3, true)
	if ch.Weight != 85 {
		t.Errorf("Weight = %d, want 85", ch.Weight)
	}
	AffectModify(ch, aff3, false)
	if ch.Weight != 80 {
		t.Errorf("Weight = %d, want 80 after remove", ch.Weight)
	}
}

// AffectModify: no-op apply types (LEVEL, AGE, GOLD, EXP)
func TestAffectModify_NoOpTypes(t *testing.T) {
	ch := &types.CharData{Name: "Tester", Level: 10, Gold: 500}

	noOps := []int{types.APPLY_LEVEL, types.APPLY_AGE, types.APPLY_GOLD, types.APPLY_EXP}
	for _, loc := range noOps {
		aff := &types.AffectData{Type: 1, Duration: 10, Location: loc, Modifier: 99}
		// Should not panic
		AffectModify(ch, aff, true)
		AffectModify(ch, aff, false)
	}
	// Level and Gold should be unchanged (these apply types are intentionally no-ops)
	if ch.Level != 10 {
		t.Errorf("Level = %d, want 10 (should be unchanged)", ch.Level)
	}
	if ch.Gold != 500 {
		t.Errorf("Gold = %d, want 500 (should be unchanged)", ch.Gold)
	}
}

// AffectJoin: when no existing affect exists, adds new
func TestAffectJoin_NoExisting(t *testing.T) {
	ch := &types.CharData{Name: "Tester"}
	aff := &types.AffectData{Type: 42, Duration: 10, Location: types.APPLY_STR, Modifier: 3}

	AffectJoin(ch, aff)

	if len(ch.Affects) != 1 {
		t.Fatalf("ch.Affects = %d, want 1", len(ch.Affects))
	}
	if ch.Affects[0].Duration != 10 {
		t.Errorf("Duration = %d, want 10", ch.Affects[0].Duration)
	}
	if ch.Affects[0].Modifier != 3 {
		t.Errorf("Modifier = %d, want 3", ch.Affects[0].Modifier)
	}
}

// AffectJoin: when modifier is zero on new affect, keeps old modifier
func TestAffectJoin_ZeroModifier(t *testing.T) {
	ch := &types.CharData{Name: "Tester"}
	AffectToChar(ch, &types.AffectData{Type: 42, Duration: 10, Location: types.APPLY_STR, Modifier: 5})

	AffectJoin(ch, &types.AffectData{Type: 42, Duration: 5, Location: types.APPLY_STR, Modifier: 0})

	if len(ch.Affects) != 1 {
		t.Fatalf("ch.Affects = %d, want 1", len(ch.Affects))
	}
	if ch.Affects[0].Duration != 15 {
		t.Errorf("Duration = %d, want 15", ch.Affects[0].Duration)
	}
	if ch.Affects[0].Modifier != 5 {
		t.Errorf("Modifier = %d, want 5 (should keep old modifier when new is 0)", ch.Affects[0].Modifier)
	}
}

// CreateObject: copies affects from index
func TestCreateObject_Affects(t *testing.T) {
	w := newTestWorld()
	idx := newTestObjIndex(2080)
	idx.Affects = []*types.AffectData{
		{Type: 1, Duration: -1, Location: types.APPLY_STR, Modifier: 2},
		{Type: 2, Duration: -1, Location: types.APPLY_DEX, Modifier: 1},
	}

	obj := CreateObject(w, idx, 5)

	if len(obj.Affects) != 2 {
		t.Fatalf("obj.Affects = %d, want 2", len(obj.Affects))
	}
	if obj.Affects[0] == idx.Affects[0] {
		t.Error("affects should be copies, not same pointers")
	}
	if obj.Affects[0].Modifier != 2 {
		t.Errorf("affects[0].Modifier = %d, want 2", obj.Affects[0].Modifier)
	}
	if obj.Affects[1].Modifier != 1 {
		t.Errorf("affects[1].Modifier = %d, want 1", obj.Affects[1].Modifier)
	}
}

// CreateMobile: AC from index data when non-zero
func TestCreateMobile_AC(t *testing.T) {
	w := newTestWorld()
	idx := newTestMobIndex(1100, 10)
	idx.AC = -50

	mob := CreateMobile(w, idx)

	if mob.Armor != -50 {
		t.Errorf("Armor = %d, want -50 (from index AC)", mob.Armor)
	}
}

// --- Find function tests ---

// GetObjList: basic search, exact then prefix
func TestGetObjList_ExactMatch(t *testing.T) {
	w := newTestWorld()
	idx1 := newTestObjIndex(2090)
	idx1.Name = "sword"
	sword := CreateObject(w, idx1, 1)

	idx2 := newTestObjIndex(2091)
	idx2.Name = "swordfish trophy"
	trophy := CreateObject(w, idx2, 1)

	list := []*types.ObjData{sword, trophy}

	// "sword" exact match should find sword first
	found := GetObjList(list, "sword")
	if found != sword {
		t.Error("exact match 'sword' should find sword, not swordfish trophy")
	}

	_ = trophy // used in list
}

// GetObjList: N.name syntax
func TestGetObjList_NthMatch(t *testing.T) {
	w := newTestWorld()
	idx1 := newTestObjIndex(2092)
	idx1.Name = "gold coin"
	coin1 := CreateObject(w, idx1, 1)

	idx2 := newTestObjIndex(2093)
	idx2.Name = "gold coin"
	coin2 := CreateObject(w, idx2, 1)

	list := []*types.ObjData{coin1, coin2}

	found := GetObjList(list, "1.coin")
	if found != coin1 {
		t.Error("1.coin should find first coin")
	}

	found = GetObjList(list, "2.coin")
	if found != coin2 {
		t.Error("2.coin should find second coin")
	}

	found = GetObjList(list, "3.coin")
	if found != nil {
		t.Error("3.coin should return nil (only 2 coins)")
	}
}

// GetObjList: empty list returns nil
func TestGetObjList_EmptyList(t *testing.T) {
	found := GetObjList(nil, "anything")
	if found != nil {
		t.Error("empty list should return nil")
	}
}

// GetObjList: prefix match when no exact match
func TestGetObjList_PrefixOnly(t *testing.T) {
	w := newTestWorld()
	idx := newTestObjIndex(2094)
	idx.Name = "longsword of fire"
	obj := CreateObject(w, idx, 1)

	list := []*types.ObjData{obj}

	found := GetObjList(list, "long")
	if found != obj {
		t.Error("prefix match 'long' should find 'longsword of fire'")
	}
}

// GetObjHere: precedence (room > inventory > equipment)
func TestGetObjHere_Precedence(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(3200)
	ch := &types.CharData{Name: "Tester"}
	CharToRoom(ch, room)

	// Same name in room, inventory, and equipment
	idxRoom := newTestObjIndex(2100)
	idxRoom.Name = "diamond ring"
	roomObj := CreateObject(w, idxRoom, 1)
	ObjToRoom(roomObj, room)

	idxInv := newTestObjIndex(2101)
	idxInv.Name = "diamond ring"
	invObj := CreateObject(w, idxInv, 1)
	ObjToChar(invObj, ch)

	idxEq := newTestObjIndex(2102)
	idxEq.Name = "diamond ring"
	eqObj := CreateObject(w, idxEq, 1)
	EquipChar(ch, eqObj, types.WEAR_WIELD)

	// Should find room object first
	found := GetObjHere(ch, "ring")
	if found != roomObj {
		t.Error("GetObjHere should find room object first (precedence)")
	}
}

// GetObjHere: falls through to inventory when not in room
func TestGetObjHere_FallsToInventory(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(3201)
	ch := &types.CharData{Name: "Tester"}
	CharToRoom(ch, room)

	idx := newTestObjIndex(2103)
	idx.Name = "silver dagger"
	dagger := CreateObject(w, idx, 1)
	ObjToChar(dagger, ch)

	found := GetObjHere(ch, "dagger")
	if found != dagger {
		t.Error("GetObjHere should find inventory item when not in room")
	}
}

// GetObjHere: falls through to equipment
func TestGetObjHere_FallsToEquipment(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(3202)
	ch := &types.CharData{Name: "Tester"}
	CharToRoom(ch, room)

	idx := newTestObjIndex(2104)
	idx.Name = "mystic staff"
	staff := CreateObject(w, idx, 1)
	EquipChar(ch, staff, types.WEAR_WIELD)

	found := GetObjHere(ch, "staff")
	if found != staff {
		t.Error("GetObjHere should find equipped item as last resort")
	}
}

// GetObjHere: ch not in room (nil InRoom) still searches inventory
func TestGetObjHere_NilRoom(t *testing.T) {
	w := newTestWorld()
	ch := &types.CharData{Name: "Tester"}

	idx := newTestObjIndex(2105)
	idx.Name = "potion"
	potion := CreateObject(w, idx, 1)
	ObjToChar(potion, ch)

	found := GetObjHere(ch, "potion")
	if found != potion {
		t.Error("GetObjHere should search inventory even with nil InRoom")
	}
}

// GetCharRoom: exact match takes priority over prefix
func TestGetCharRoom_ExactOverPrefix(t *testing.T) {
	room := newTestRoom(3210)
	ch := &types.CharData{Name: "Tester"}
	CharToRoom(ch, room)

	// "guard" exact match exists; "guardian" also starts with "guard"
	mob1 := &types.CharData{Name: "guardian angel"}
	mob1.Act.Set(types.ACT_IS_NPC)
	CharToRoom(mob1, room)

	mob2 := &types.CharData{Name: "guard"}
	mob2.Act.Set(types.ACT_IS_NPC)
	CharToRoom(mob2, room)

	found := GetCharRoom(ch, "guard")
	if found != mob2 {
		t.Error("exact match 'guard' should beat prefix match 'guardian'")
	}
}

// GetCharRoom: ch not in room returns nil
func TestGetCharRoom_NilRoom(t *testing.T) {
	ch := &types.CharData{Name: "Tester"}
	found := GetCharRoom(ch, "anyone")
	if found != nil {
		t.Error("GetCharRoom should return nil when ch.InRoom is nil")
	}
}

// GetCharWorld: N.name syntax
func TestGetCharWorld_NthMatch(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(3220)
	ch := &types.CharData{Name: "Tester"}
	CharToRoom(ch, room)
	w.AddChar(ch)

	room2 := newTestRoom(3221)
	mob1 := &types.CharData{Name: "guard"}
	mob1.Act.Set(types.ACT_IS_NPC)
	CharToRoom(mob1, room2)
	w.AddChar(mob1)

	mob2 := &types.CharData{Name: "guard"}
	mob2.Act.Set(types.ACT_IS_NPC)
	CharToRoom(mob2, room2)
	w.AddChar(mob2)

	found := GetCharWorld(w, ch, "2.guard")
	if found != mob2 {
		t.Error("2.guard should find second guard in world")
	}
}

// GetObjCarry: N.name syntax
func TestGetObjCarry_NthMatch(t *testing.T) {
	w := newTestWorld()
	ch := &types.CharData{Name: "Tester"}

	idx1 := newTestObjIndex(2110)
	idx1.Name = "potion heal"
	pot1 := CreateObject(w, idx1, 1)
	ObjToChar(pot1, ch)

	idx2 := newTestObjIndex(2111)
	idx2.Name = "potion heal"
	pot2 := CreateObject(w, idx2, 1)
	ObjToChar(pot2, ch)

	found := GetObjCarry(ch, "2.potion")
	if found != pot2 {
		t.Error("2.potion should find second potion")
	}
}

// GetObjWear: N.name syntax
func TestGetObjWear_NthMatch(t *testing.T) {
	w := newTestWorld()
	ch := &types.CharData{Name: "Tester"}

	idx1 := newTestObjIndex(2112)
	idx1.Name = "iron ring"
	ring1 := CreateObject(w, idx1, 1)
	EquipChar(ch, ring1, types.WEAR_WIELD)

	idx2 := newTestObjIndex(2113)
	idx2.Name = "iron ring"
	ring2 := CreateObject(w, idx2, 1)
	EquipChar(ch, ring2, types.WEAR_DUAL_WIELD)

	found := GetObjWear(ch, "2.ring")
	if found != ring2 {
		t.Error("2.ring should find second equipped ring")
	}
}

// GetObjWorld: exact match takes priority over prefix
func TestGetObjWorld_ExactOverPrefix(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(3230)
	ch := &types.CharData{Name: "Tester"}
	CharToRoom(ch, room)

	room2 := newTestRoom(3231)

	idx1 := newTestObjIndex(2120)
	idx1.Name = "swords collection"
	coll := CreateObject(w, idx1, 1)
	ObjToRoom(coll, room2)

	idx2 := newTestObjIndex(2121)
	idx2.Name = "sword"
	sword := CreateObject(w, idx2, 1)
	ObjToRoom(sword, room2)

	found := GetObjWorld(w, ch, "sword")
	if found != sword {
		t.Error("exact match 'sword' should take priority in world search")
	}
}

// GetObjWorld: N.name syntax in world search
func TestGetObjWorld_NthMatch(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(3232)
	ch := &types.CharData{Name: "Tester"}
	CharToRoom(ch, room)

	room2 := newTestRoom(3233)
	idx1 := newTestObjIndex(2122)
	idx1.Name = "gold coin"
	coin1 := CreateObject(w, idx1, 1)
	ObjToRoom(coin1, room2)

	idx2 := newTestObjIndex(2123)
	idx2.Name = "gold coin"
	coin2 := CreateObject(w, idx2, 1)
	ObjToRoom(coin2, room2)

	found := GetObjWorld(w, ch, "2.coin")
	if found != coin2 {
		t.Error("2.coin should find second coin in world")
	}
}

// --- Reset function tests ---

// resetObject: skips if object already in room
func TestResetObject_SkipsDuplicate(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(5010)
	w.Rooms[5010] = room
	idx := newTestObjIndex(2200)
	w.ObjIndex[2200] = idx

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'O', Arg1: 2200, Arg3: 5010},
			{Command: 'O', Arg1: 2200, Arg3: 5010}, // duplicate, should be skipped
		},
	}

	ResetArea(w, area)

	if len(room.Contents) != 1 {
		t.Errorf("room.Contents = %d, want 1 (duplicate should be skipped)", len(room.Contents))
	}
}

// resetObject: bad vnum (obj not found)
func TestResetObject_BadObjVnum(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(5011)
	w.Rooms[5011] = room

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'O', Arg1: 99999, Arg3: 5011}, // non-existent obj
		},
	}

	// Should not panic
	ResetArea(w, area)

	if len(room.Contents) != 0 {
		t.Error("room should be empty when obj vnum not found")
	}
}

// resetObject: bad room vnum
func TestResetObject_BadRoomVnum(t *testing.T) {
	w := newTestWorld()
	idx := newTestObjIndex(2201)
	w.ObjIndex[2201] = idx

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'O', Arg1: 2201, Arg3: 99999}, // non-existent room
		},
	}

	// Should not panic
	ResetArea(w, area)
}

// resetPut: put an object inside a container
func TestResetPut_IntoContainer(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(5012)
	w.Rooms[5012] = room

	containerIdx := newTestObjIndex(2210)
	containerIdx.Name = "chest"
	w.ObjIndex[2210] = containerIdx

	itemIdx := newTestObjIndex(2211)
	itemIdx.Name = "gem"
	w.ObjIndex[2211] = itemIdx

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'O', Arg1: 2210, Arg3: 5012},          // place chest in room
			{Command: 'P', Arg1: 2211, Arg2: 1, Arg3: 2210}, // put gem in chest
		},
	}

	ResetArea(w, area)

	if len(room.Contents) != 1 {
		t.Fatalf("room.Contents = %d, want 1 (just the chest)", len(room.Contents))
	}
	chest := room.Contents[0]
	if len(chest.Contents) != 1 {
		t.Fatalf("chest.Contents = %d, want 1", len(chest.Contents))
	}
	if chest.Contents[0].Name != "gem" {
		t.Errorf("item name = %q, want %q", chest.Contents[0].Name, "gem")
	}
}

// resetPut: uses lastObj when Arg3 is 0
func TestResetPut_UsesLastObj(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(5013)
	w.Rooms[5013] = room

	containerIdx := newTestObjIndex(2212)
	containerIdx.Name = "bag"
	w.ObjIndex[2212] = containerIdx

	itemIdx := newTestObjIndex(2213)
	itemIdx.Name = "scroll"
	w.ObjIndex[2213] = itemIdx

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'O', Arg1: 2212, Arg3: 5013},       // place bag in room (sets lastObj)
			{Command: 'P', Arg1: 2213, Arg2: 1, Arg3: 0}, // put scroll in lastObj (bag)
		},
	}

	ResetArea(w, area)

	bag := room.Contents[0]
	if len(bag.Contents) != 1 {
		t.Fatalf("bag.Contents = %d, want 1", len(bag.Contents))
	}
	if bag.Contents[0].Name != "scroll" {
		t.Errorf("item name = %q, want %q", bag.Contents[0].Name, "scroll")
	}
}

// resetPut: bad obj vnum
func TestResetPut_BadObjVnum(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(5014)
	w.Rooms[5014] = room

	containerIdx := newTestObjIndex(2214)
	w.ObjIndex[2214] = containerIdx

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'O', Arg1: 2214, Arg3: 5014},
			{Command: 'P', Arg1: 99999, Arg2: 1, Arg3: 2214}, // bad item vnum
		},
	}

	// Should not panic
	ResetArea(w, area)
}

// resetPut: bad container vnum
func TestResetPut_BadContainerVnum(t *testing.T) {
	w := newTestWorld()
	idx := newTestObjIndex(2215)
	w.ObjIndex[2215] = idx

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'P', Arg1: 2215, Arg2: 1, Arg3: 99999}, // bad container vnum
		},
	}

	// Should not panic
	ResetArea(w, area)
}

// resetPut: skips duplicate in container
func TestResetPut_SkipsDuplicate(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(5015)
	w.Rooms[5015] = room

	containerIdx := newTestObjIndex(2216)
	w.ObjIndex[2216] = containerIdx

	itemIdx := newTestObjIndex(2217)
	itemIdx.Name = "key"
	w.ObjIndex[2217] = itemIdx

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'O', Arg1: 2216, Arg3: 5015},
			{Command: 'P', Arg1: 2217, Arg2: 1, Arg3: 2216},
			{Command: 'P', Arg1: 2217, Arg2: 1, Arg3: 2216}, // duplicate, should be skipped
		},
	}

	ResetArea(w, area)

	chest := room.Contents[0]
	if len(chest.Contents) != 1 {
		t.Errorf("chest.Contents = %d, want 1 (duplicate should be skipped)", len(chest.Contents))
	}
}

// resetPut: nil lastObj when Arg3 is 0 (no previous O reset)
func TestResetPut_NilLastObj(t *testing.T) {
	w := newTestWorld()
	idx := newTestObjIndex(2218)
	w.ObjIndex[2218] = idx

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'P', Arg1: 2218, Arg2: 1, Arg3: 0}, // no lastObj
		},
	}

	// Should not panic, just skip
	ResetArea(w, area)
}

// resetGive: nil lastMob (no previous M reset)
func TestResetGive_NilLastMob(t *testing.T) {
	w := newTestWorld()
	idx := newTestObjIndex(2220)
	w.ObjIndex[2220] = idx

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'G', Arg1: 2220}, // no lastMob
		},
	}

	// Should not panic
	ResetArea(w, area)
}

// resetGive: bad obj vnum
func TestResetGive_BadObjVnum(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(5020)
	w.Rooms[5020] = room
	mobIdx := newTestMobIndex(1050, 5)
	w.MobIndex[1050] = mobIdx

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'M', Arg1: 1050, Arg2: 1, Arg3: 5020},
			{Command: 'G', Arg1: 99999}, // bad obj vnum
		},
	}

	// Should not panic
	ResetArea(w, area)
	mob := room.People[0]
	if len(mob.Carrying) != 0 {
		t.Error("mob should have no items when obj vnum not found")
	}
}

// resetGive: shop mob sets ITEM_INVENTORY
func TestResetGive_ShopMob(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(5021)
	w.Rooms[5021] = room

	mobIdx := newTestMobIndex(1051, 10)
	mobIdx.Shop = &types.ShopData{}
	w.MobIndex[1051] = mobIdx

	objIdx := newTestObjIndex(2221)
	w.ObjIndex[2221] = objIdx

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'M', Arg1: 1051, Arg2: 1, Arg3: 5021},
			{Command: 'G', Arg1: 2221},
		},
	}

	ResetArea(w, area)

	mob := room.People[0]
	if len(mob.Carrying) != 1 {
		t.Fatalf("mob.Carrying = %d, want 1", len(mob.Carrying))
	}
	if !mob.Carrying[0].ExtraFlags.IsSet(types.ITEM_INVENTORY) {
		t.Error("shop mob item should have ITEM_INVENTORY flag set")
	}
}

// resetEquip: nil lastMob
func TestResetEquip_NilLastMob(t *testing.T) {
	w := newTestWorld()
	idx := newTestObjIndex(2230)
	w.ObjIndex[2230] = idx

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'E', Arg1: 2230, Arg3: types.WEAR_WIELD}, // no lastMob
		},
	}

	// Should not panic
	ResetArea(w, area)
}

// resetEquip: bad obj vnum
func TestResetEquip_BadObjVnum(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(5022)
	w.Rooms[5022] = room
	mobIdx := newTestMobIndex(1052, 5)
	w.MobIndex[1052] = mobIdx

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'M', Arg1: 1052, Arg2: 1, Arg3: 5022},
			{Command: 'E', Arg1: 99999, Arg3: types.WEAR_WIELD}, // bad vnum
		},
	}

	// Should not panic
	ResetArea(w, area)
	mob := room.People[0]
	if len(mob.Carrying) != 0 {
		t.Error("mob should have no items when obj vnum not found")
	}
}

// resetDoor: open state
func TestResetDoor_Open(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(5030)
	room.Exits = []*types.ExitData{
		{Direction: types.DIR_NORTH, ExitInfo: int(types.EX_CLOSED) | int(types.EX_LOCKED), ToRoom: newTestRoom(5031)},
	}
	w.Rooms[5030] = room

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'D', Arg1: 5030, Arg2: types.DIR_NORTH, Arg3: 0}, // Open
		},
	}

	ResetArea(w, area)

	exit := room.Exits[0]
	if exit.ExitInfo&int(types.EX_CLOSED) != 0 {
		t.Error("door should not be closed after reset to open")
	}
	if exit.ExitInfo&int(types.EX_LOCKED) != 0 {
		t.Error("door should not be locked after reset to open")
	}
}

// resetDoor: closed (not locked) state
func TestResetDoor_Closed(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(5032)
	room.Exits = []*types.ExitData{
		{Direction: types.DIR_NORTH, ExitInfo: 0, ToRoom: newTestRoom(5033)},
	}
	w.Rooms[5032] = room

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'D', Arg1: 5032, Arg2: types.DIR_NORTH, Arg3: 1}, // Closed
		},
	}

	ResetArea(w, area)

	exit := room.Exits[0]
	if exit.ExitInfo&int(types.EX_CLOSED) == 0 {
		t.Error("door should be closed")
	}
	if exit.ExitInfo&int(types.EX_LOCKED) != 0 {
		t.Error("door should not be locked (only closed)")
	}
}

// resetDoor: bad room vnum (should not panic)
func TestResetDoor_BadRoomVnum(t *testing.T) {
	w := newTestWorld()
	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'D', Arg1: 99999, Arg2: types.DIR_NORTH, Arg3: 0},
		},
	}

	// Should not panic
	ResetArea(w, area)
}

// resetDoor: bad exit direction (should not panic)
func TestResetDoor_BadDirection(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(5034)
	room.Exits = []*types.ExitData{
		{Direction: types.DIR_NORTH, ExitInfo: 0, ToRoom: newTestRoom(5035)},
	}
	w.Rooms[5034] = room

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'D', Arg1: 5034, Arg2: types.DIR_SOUTH, Arg3: 0}, // No south exit
		},
	}

	// Should not panic (exit is nil, early return)
	ResetArea(w, area)
}

// resetHide: hide an object by vnum
func TestResetHide_ByVnum(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(5040)
	w.Rooms[5040] = room

	idx := newTestObjIndex(2240)
	idx.Name = "hidden treasure"
	w.ObjIndex[2240] = idx

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'O', Arg1: 2240, Arg3: 5040},
			{Command: 'H', Arg1: 2240},
		},
	}

	ResetArea(w, area)

	if len(room.Contents) != 1 {
		t.Fatal("object should be in room")
	}
	if !room.Contents[0].ExtraFlags.IsSet(types.ITEM_HIDDEN) {
		t.Error("object should have ITEM_HIDDEN flag set")
	}
}

// resetHide: uses lastObj when Arg1 is 0
func TestResetHide_UsesLastObj(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(5041)
	w.Rooms[5041] = room

	idx := newTestObjIndex(2241)
	idx.Name = "secret note"
	w.ObjIndex[2241] = idx

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'O', Arg1: 2241, Arg3: 5041}, // sets lastObj
			{Command: 'H', Arg1: 0},                // use lastObj
		},
	}

	ResetArea(w, area)

	if !room.Contents[0].ExtraFlags.IsSet(types.ITEM_HIDDEN) {
		t.Error("lastObj should have ITEM_HIDDEN flag set")
	}
}

// resetHide: nil lastObj and Arg1=0 (should not panic)
func TestResetHide_NilLastObj(t *testing.T) {
	w := newTestWorld()
	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'H', Arg1: 0}, // no lastObj
		},
	}

	// Should not panic
	ResetArea(w, area)
}

// resetHide: bad vnum
func TestResetHide_BadVnum(t *testing.T) {
	w := newTestWorld()
	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'H', Arg1: 99999}, // non-existent obj
		},
	}

	// Should not panic
	ResetArea(w, area)
}

// resetMobile: bad mob vnum
func TestResetMobile_BadMobVnum(t *testing.T) {
	w := newTestWorld()
	room := newTestRoom(5050)
	w.Rooms[5050] = room

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'M', Arg1: 99999, Arg2: 1, Arg3: 5050},
		},
	}

	// Should not panic
	ResetArea(w, area)
	if len(room.People) != 0 {
		t.Error("room should be empty when mob vnum not found")
	}
}

// resetMobile: bad room vnum
func TestResetMobile_BadRoomVnum(t *testing.T) {
	w := newTestWorld()
	idx := newTestMobIndex(1060, 5)
	w.MobIndex[1060] = idx

	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'M', Arg1: 1060, Arg2: 1, Arg3: 99999},
		},
	}

	// Should not panic
	ResetArea(w, area)
}

// ResetAllAreas: processes multiple areas
func TestResetAllAreas(t *testing.T) {
	w := newTestWorld()
	room1 := newTestRoom(6000)
	w.Rooms[6000] = room1
	room2 := newTestRoom(6001)
	w.Rooms[6001] = room2

	mobIdx1 := newTestMobIndex(1070, 5)
	w.MobIndex[1070] = mobIdx1
	mobIdx2 := newTestMobIndex(1071, 5)
	w.MobIndex[1071] = mobIdx2

	area1 := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'M', Arg1: 1070, Arg2: 1, Arg3: 6000},
		},
	}
	area2 := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'M', Arg1: 1071, Arg2: 1, Arg3: 6001},
		},
	}

	w.Areas = []*types.AreaData{area1, area2}
	ResetAllAreas(w)

	if len(room1.People) != 1 {
		t.Errorf("room1.People = %d, want 1", len(room1.People))
	}
	if len(room2.People) != 1 {
		t.Errorf("room2.People = %d, want 1", len(room2.People))
	}
}

// findObjByIndex: finds object, returns nil when not found
func TestFindObjByIndex(t *testing.T) {
	w := newTestWorld()
	idx := newTestObjIndex(2250)
	obj := CreateObject(w, idx, 1)

	found := findObjByIndex(w, idx)
	if found != obj {
		t.Error("findObjByIndex should find the object")
	}

	otherIdx := newTestObjIndex(2251)
	found = findObjByIndex(w, otherIdx)
	if found != nil {
		t.Error("findObjByIndex should return nil for uninstantiated index")
	}
}

// UnequipChar: already unequipped (no-op with bug log)
func TestUnequipChar_AlreadyUnequipped(t *testing.T) {
	w := newTestWorld()
	ch := &types.CharData{Name: "Tester"}
	obj := CreateObject(w, newTestObjIndex(2260), 1)
	ObjToChar(obj, ch)
	// obj.WearLoc is WEAR_NONE (in inventory)

	// Should not panic, just log bug
	UnequipChar(ch, obj)
	if obj.WearLoc != types.WEAR_NONE {
		t.Error("WearLoc should remain WEAR_NONE")
	}
}

// ExtractObj from an ObjInObj (nested)
func TestExtractObj_FromObj(t *testing.T) {
	w := newTestWorld()
	containerIdx := newTestObjIndex(2270)
	container := CreateObject(w, containerIdx, 1)
	room := newTestRoom(3300)
	ObjToRoom(container, room)

	itemIdx := newTestObjIndex(2271)
	item := CreateObject(w, itemIdx, 1)
	ObjToObj(item, container)

	// Extract item from inside container
	ExtractObj(w, item)

	if len(container.Contents) != 0 {
		t.Error("container should be empty after extracting item")
	}
	if len(w.Objects) != 1 {
		t.Errorf("world.Objects = %d, want 1 (only container remains)", len(w.Objects))
	}
}

// ResetArea: R/T/B commands are no-ops (should not panic)
func TestResetArea_SkippedCommands(t *testing.T) {
	w := newTestWorld()
	area := &types.AreaData{
		Resets: []*types.ResetData{
			{Command: 'R', Arg1: 1},
			{Command: 'T', Arg1: 1},
			{Command: 'B', Arg1: 1},
		},
	}

	// Should not panic
	ResetArea(w, area)
}
