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
