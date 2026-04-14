package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func setupRepairWorld() *world.World {
	w := world.New("/tmp/test")
	WorldRef = w
	return w
}

// newRepairKeeper builds a fixer NPC with a RepairData configured to accept
// ARMOR, WEAPON, and WAND, at 100% profit_fix (cost == base).
func newRepairKeeper(w *world.World, room *types.RoomIndexData) *types.CharData {
	idx := &types.MobIndexData{
		Vnum:       9001,
		PlayerName: "smith",
		ShortDescr: "Thrain the smith",
		RShop: &types.RepairData{
			Keeper:    9001,
			FixType:   [types.MAX_FIX]int{types.ITEM_ARMOR, types.ITEM_WEAPON, types.ITEM_WAND},
			ProfitFix: 1000,
			ShopType:  0,
		},
	}
	keeper := &types.CharData{
		Name:       "smith",
		ShortDescr: "Thrain the smith",
		IndexData:  idx,
		Position:   types.POS_STANDING,
		Level:      50,
	}
	keeper.Act.Set(types.ACT_IS_NPC)
	keeper.InRoom = room
	room.People = append(room.People, keeper)
	return keeper
}

func TestFindFixer_Present(t *testing.T) {
	w := setupRepairWorld()
	room := &types.RoomIndexData{Vnum: 7100, Name: "Smithy"}
	w.Rooms[7100] = room

	ch, client := makeTestChar("Visitor")
	defer client.Close()
	handler.CharToRoom(ch, room)
	newRepairKeeper(w, room)

	if got := findFixer(ch); got == nil {
		t.Fatal("findFixer should find keeper")
	}
}

func TestFindFixer_Absent(t *testing.T) {
	_ = setupRepairWorld()
	room := &types.RoomIndexData{Vnum: 7101, Name: "Empty"}
	ch, client := makeTestChar("Lonely")
	defer client.Close()
	handler.CharToRoom(ch, room)
	if got := findFixer(ch); got != nil {
		t.Error("findFixer should return nil in empty room")
	}
}

func TestComputeRepairCost_Armor(t *testing.T) {
	w := setupRepairWorld()
	room := &types.RoomIndexData{Vnum: 7102, Name: "Smithy"}
	w.Rooms[7102] = room
	keeper := newRepairKeeper(w, room)

	obj := &types.ObjData{ItemType: types.ITEM_ARMOR, GoldCost: 100}
	obj.Value[0] = 3 // current
	obj.Value[1] = 8 // max

	cost := computeRepairCost(keeper, obj)
	// cost = 100 * 1000 / 1000 = 100; delta = 5; total = 500
	if cost != 500 {
		t.Errorf("armor cost = %d; want 500", cost)
	}
}

func TestComputeRepairCost_WeaponPristine(t *testing.T) {
	w := setupRepairWorld()
	room := &types.RoomIndexData{Vnum: 7103, Name: "Smithy"}
	w.Rooms[7103] = room
	keeper := newRepairKeeper(w, room)

	obj := &types.ObjData{ItemType: types.ITEM_WEAPON, GoldCost: 50}
	obj.Value[0] = types.INIT_WEAPON_CONDITION // at max

	if got := computeRepairCost(keeper, obj); got != repairCostPristine {
		t.Errorf("pristine weapon should return pristine sentinel; got %d", got)
	}
}

func TestComputeRepairCost_UnsupportedType(t *testing.T) {
	w := setupRepairWorld()
	room := &types.RoomIndexData{Vnum: 7104, Name: "Smithy"}
	w.Rooms[7104] = room
	keeper := newRepairKeeper(w, room)

	obj := &types.ObjData{ItemType: types.ITEM_FOOD, GoldCost: 10}
	if got := computeRepairCost(keeper, obj); got != repairCostUnrepairable {
		t.Errorf("food should be unrepairable; got %d", got)
	}
}

func TestDoRepair_NoKeeper(t *testing.T) {
	_ = setupRepairWorld()
	room := &types.RoomIndexData{Vnum: 7105, Name: "Empty"}
	ch, client := makeTestChar("Wanderer")
	defer client.Close()
	handler.CharToRoom(ch, room)

	DoRepair(ch, "anything")
	out := readOutput(ch, client)
	if !strings.Contains(out, "can't do that here") {
		t.Errorf("expected 'can't do that here', got: %q", out)
	}
}

func TestDoRepair_InsufficientGold(t *testing.T) {
	w := setupRepairWorld()
	room := &types.RoomIndexData{Vnum: 7106, Name: "Smithy"}
	w.Rooms[7106] = room
	newRepairKeeper(w, room)

	ch, client := makeTestChar("PoorHero")
	defer client.Close()
	handler.CharToRoom(ch, room)
	ch.Gold = 10 // not enough

	idx := &types.ObjIndexData{Vnum: 8001, Name: "rusty sword", ShortDescr: "a rusty sword",
		WearFlags: int(types.ITEM_TAKE)}
	w.ObjIndex[8001] = idx
	obj := handler.CreateObject(w, idx, 1)
	obj.ItemType = types.ITEM_WEAPON
	obj.GoldCost = 100
	obj.Value[0] = 2 // very damaged
	handler.ObjToChar(obj, ch)

	DoRepair(ch, "sword")
	out := readOutput(ch, client)
	if !strings.Contains(out, "can't afford") {
		t.Errorf("expected 'can't afford', got: %q", out)
	}
	// Gold unchanged
	if ch.Gold != 10 {
		t.Errorf("gold should be untouched; got %d", ch.Gold)
	}
	// Condition unchanged
	if obj.Value[0] != 2 {
		t.Errorf("condition should be untouched; got %d", obj.Value[0])
	}
}

func TestDoRepair_Success(t *testing.T) {
	w := setupRepairWorld()
	room := &types.RoomIndexData{Vnum: 7107, Name: "Smithy"}
	w.Rooms[7107] = room
	keeper := newRepairKeeper(w, room)

	ch, client := makeTestChar("RichHero")
	defer client.Close()
	handler.CharToRoom(ch, room)
	ch.Gold = 10000

	idx := &types.ObjIndexData{Vnum: 8002, Name: "dented plate", ShortDescr: "a dented plate",
		WearFlags: int(types.ITEM_TAKE)}
	w.ObjIndex[8002] = idx
	obj := handler.CreateObject(w, idx, 1)
	obj.ItemType = types.ITEM_ARMOR
	obj.GoldCost = 100
	obj.Value[0] = 3
	obj.Value[1] = 8
	handler.ObjToChar(obj, ch)

	kGoldBefore := keeper.Gold
	DoRepair(ch, "plate")
	out := readOutput(ch, client)

	if !strings.Contains(out, "repairs") {
		t.Errorf("expected success message, got: %q", out)
	}
	if obj.Value[0] != 8 {
		t.Errorf("armor should be restored to max; got %d", obj.Value[0])
	}
	if ch.Gold != 10000-500 {
		t.Errorf("gold should be deducted by 500; got %d", ch.Gold)
	}
	if keeper.Gold != kGoldBefore+500 {
		t.Errorf("keeper should receive the gold; got %d", keeper.Gold)
	}
}

func TestDoAppraise_Pristine(t *testing.T) {
	w := setupRepairWorld()
	room := &types.RoomIndexData{Vnum: 7108, Name: "Smithy"}
	w.Rooms[7108] = room
	newRepairKeeper(w, room)

	ch, client := makeTestChar("Owner")
	defer client.Close()
	handler.CharToRoom(ch, room)

	idx := &types.ObjIndexData{Vnum: 8003, Name: "shiny sword", ShortDescr: "a shiny sword",
		WearFlags: int(types.ITEM_TAKE)}
	w.ObjIndex[8003] = idx
	obj := handler.CreateObject(w, idx, 1)
	obj.ItemType = types.ITEM_WEAPON
	obj.GoldCost = 200
	obj.Value[0] = types.INIT_WEAPON_CONDITION
	handler.ObjToChar(obj, ch)

	DoAppraise(ch, "sword")
	out := readOutput(ch, client)
	if !strings.Contains(out, "looks fine") {
		t.Errorf("expected 'looks fine', got: %q", out)
	}
}

func TestDoAppraise_EstimateCost(t *testing.T) {
	w := setupRepairWorld()
	room := &types.RoomIndexData{Vnum: 7109, Name: "Smithy"}
	w.Rooms[7109] = room
	newRepairKeeper(w, room)

	ch, client := makeTestChar("Owner")
	defer client.Close()
	handler.CharToRoom(ch, room)

	idx := &types.ObjIndexData{Vnum: 8004, Name: "cracked plate", ShortDescr: "a cracked plate",
		WearFlags: int(types.ITEM_TAKE)}
	w.ObjIndex[8004] = idx
	obj := handler.CreateObject(w, idx, 1)
	obj.ItemType = types.ITEM_ARMOR
	obj.GoldCost = 100
	obj.Value[0] = 2
	obj.Value[1] = 8
	handler.ObjToChar(obj, ch)

	DoAppraise(ch, "plate")
	out := readOutput(ch, client)
	if !strings.Contains(out, "600 gold") {
		t.Errorf("expected cost estimate 600, got: %q", out)
	}
}
