package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func setupRepairWorld() (*world.World, *types.RoomIndexData) {
	w := world.New("/tmp/test")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 9500, Name: "Forge"}
	return w, room
}

func makeFixer(room *types.RoomIndexData, profit, shopType int, fixTypes ...int) *types.CharData {
	rshop := &types.RepairData{
		Keeper:    3001,
		ProfitFix: profit,
		ShopType:  shopType,
	}
	for i, ft := range fixTypes {
		if i >= len(rshop.FixType) {
			break
		}
		rshop.FixType[i] = ft
	}
	mob := &types.CharData{
		Name: "smith", ShortDescr: "the smith",
		Position:  types.POS_STANDING,
		IndexData: &types.MobIndexData{Vnum: 3001, RShop: rshop},
	}
	mob.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(mob, room)
	return mob
}

func TestDoRepair_NoKeeper(t *testing.T) {
	_, room := setupRepairWorld()
	ch, client := makeTestChar("Cust")
	defer client.Close()
	handler.CharToRoom(ch, room)
	DoRepair(ch, "sword")
	out := readOutput(ch, client)
	if !strings.Contains(out, "can't do that here") {
		t.Errorf("expected no-keeper msg; got %q", out)
	}
}

func TestDoRepair_WrongItemType(t *testing.T) {
	_, room := setupRepairWorld()
	_ = makeFixer(room, 100, types.SHOP_FIX, types.ITEM_WEAPON, types.ITEM_ARMOR)
	ch, client := makeTestChar("Cust")
	defer client.Close()
	handler.CharToRoom(ch, room)
	obj := &types.ObjData{
		Name: "apple", ShortDescr: "an apple",
		ItemType: types.ITEM_FOOD, GoldCost: 10, WearLoc: types.WEAR_NONE,
	}
	handler.ObjToChar(obj, ch)
	DoRepair(ch, "apple")
	out := readOutput(ch, client)
	if !strings.Contains(out, "can't do anything") {
		t.Errorf("expected type-refusal; got %q", out)
	}
}

func TestDoRepair_InsufficientGold(t *testing.T) {
	_, room := setupRepairWorld()
	_ = makeFixer(room, 100, types.SHOP_FIX, types.ITEM_WEAPON, types.ITEM_ARMOR)
	ch, client := makeTestChar("Cust")
	defer client.Close()
	ch.Gold = 0
	handler.CharToRoom(ch, room)
	obj := &types.ObjData{
		Name: "sword", ShortDescr: "a sword",
		ItemType: types.ITEM_WEAPON, GoldCost: 1000, WearLoc: types.WEAR_NONE,
	}
	handler.ObjToChar(obj, ch)
	DoRepair(ch, "sword")
	out := readOutput(ch, client)
	if !strings.Contains(out, "can't afford") {
		t.Errorf("expected affordability msg; got %q", out)
	}
}

func TestDoRepair_SuccessfulRepair(t *testing.T) {
	_, room := setupRepairWorld()
	keeper := makeFixer(room, 100, types.SHOP_FIX, types.ITEM_WEAPON, types.ITEM_ARMOR)
	ch, client := makeTestChar("Cust")
	defer client.Close()
	ch.Gold = 5000
	handler.CharToRoom(ch, room)
	obj := &types.ObjData{
		Name: "sword", ShortDescr: "a sword",
		ItemType: types.ITEM_WEAPON, GoldCost: 400, WearLoc: types.WEAR_NONE,
		Value: [6]int{5, 0, 0, 0, 0, 0}, // damaged
	}
	handler.ObjToChar(obj, ch)

	DoRepair(ch, "sword")
	out := readOutput(ch, client)

	if !strings.Contains(out, "repair") {
		t.Errorf("expected repair confirmation; got %q", out)
	}
	if ch.Gold >= 5000 {
		t.Errorf("gold should have been deducted; still %d", ch.Gold)
	}
	if keeper.Gold == 0 {
		t.Errorf("keeper should have received payment")
	}
	if obj.Value[0] != initWeaponCondition {
		t.Errorf("weapon should be restored to condition %d; got %d", initWeaponCondition, obj.Value[0])
	}
}

func TestDoRepair_UndamagedItemRefused(t *testing.T) {
	_, room := setupRepairWorld()
	_ = makeFixer(room, 100, types.SHOP_FIX, types.ITEM_WEAPON, types.ITEM_ARMOR)
	ch, client := makeTestChar("Cust")
	defer client.Close()
	ch.Gold = 5000
	handler.CharToRoom(ch, room)
	obj := &types.ObjData{
		Name: "sword", ShortDescr: "a sword",
		ItemType: types.ITEM_WEAPON, GoldCost: 400, WearLoc: types.WEAR_NONE,
		Value: [6]int{initWeaponCondition, 0, 0, 0, 0, 0}, // pristine
	}
	handler.ObjToChar(obj, ch)

	DoRepair(ch, "sword")
	out := readOutput(ch, client)
	if !strings.Contains(out, "looks fine") {
		t.Errorf("expected 'looks fine' refusal for pristine weapon; got %q", out)
	}
	if ch.Gold != 5000 {
		t.Errorf("no charge should apply to pristine items; gold=%d", ch.Gold)
	}
}

func TestDoRepair_RechargerRefillsStaves(t *testing.T) {
	_, room := setupRepairWorld()
	_ = makeFixer(room, 100, types.SHOP_RECHARGE, types.ITEM_STAFF, types.ITEM_WAND)
	ch, client := makeTestChar("Mage")
	defer client.Close()
	ch.Gold = 10000
	handler.CharToRoom(ch, room)
	staff := &types.ObjData{
		Name: "staff", ShortDescr: "a staff",
		ItemType: types.ITEM_STAFF, GoldCost: 800, WearLoc: types.WEAR_NONE,
	}
	staff.Value[1] = 10 // max charges
	staff.Value[2] = 2  // current charges
	handler.ObjToChar(staff, ch)

	DoRepair(ch, "staff")
	_ = readOutput(ch, client)

	if staff.Value[2] != 10 {
		t.Errorf("staff should be fully recharged; charges = %d", staff.Value[2])
	}
}

func TestRepairDamageDelta_Armor(t *testing.T) {
	cases := []struct {
		name     string
		cur, max int
		want     int
	}{
		{"pristine", 20, 20, 0},
		{"damaged 5", 15, 20, 5},
		{"fully broken", 0, 20, 20},
		// Per C `shops.c:504`: undamaged when value[0] >= value[1].
		{"value0 above max treated as pristine", 25, 20, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			obj := &types.ObjData{
				ItemType: types.ITEM_ARMOR,
				Value:    [6]int{tc.cur, tc.max, 0, 0, 0, 0},
			}
			if got := repairDamageDelta(obj); got != tc.want {
				t.Errorf("delta = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestRepairDamageDelta_Wand(t *testing.T) {
	obj := &types.ObjData{
		ItemType: types.ITEM_WAND,
		Value:    [6]int{0, 10, 3, 0, 0, 0}, // max 10, current 3
	}
	if got := repairDamageDelta(obj); got != 7 {
		t.Errorf("wand delta = %d, want 7", got)
	}
	obj.Value[2] = 10
	if got := repairDamageDelta(obj); got != 0 {
		t.Errorf("full wand should be undamaged; got %d", got)
	}
}

func TestRepairDamageDelta_UnknownTypeReturnsNegative(t *testing.T) {
	obj := &types.ObjData{ItemType: types.ITEM_FOOD}
	if got := repairDamageDelta(obj); got >= 0 {
		t.Errorf("unsupported item type should return < 0; got %d", got)
	}
}

func TestDoRepair_DamagedArmorRepaired(t *testing.T) {
	_, room := setupRepairWorld()
	keeper := makeFixer(room, 100, types.SHOP_FIX, types.ITEM_ARMOR)
	ch, client := makeTestChar("Cust")
	defer client.Close()
	ch.Gold = 10000
	handler.CharToRoom(ch, room)
	armor := &types.ObjData{
		Name: "shield", ShortDescr: "a shield",
		ItemType: types.ITEM_ARMOR, GoldCost: 200, WearLoc: types.WEAR_NONE,
		Value: [6]int{5, 20, 0, 0, 0, 0}, // damaged (current 5, max 20)
	}
	handler.ObjToChar(armor, ch)

	DoRepair(ch, "shield")
	_ = readOutput(ch, client)
	if armor.Value[0] != 20 {
		t.Errorf("armor Value[0] should be restored to max 20; got %d", armor.Value[0])
	}
	if keeper.Gold == 0 {
		t.Errorf("keeper should have collected payment")
	}
}

func TestDoAppraise_QuotesWithoutCharging(t *testing.T) {
	_, room := setupRepairWorld()
	_ = makeFixer(room, 100, types.SHOP_FIX, types.ITEM_WEAPON, types.ITEM_ARMOR)
	ch, client := makeTestChar("Cust")
	defer client.Close()
	ch.Gold = 500
	handler.CharToRoom(ch, room)
	obj := &types.ObjData{
		Name: "sword", ShortDescr: "a sword",
		ItemType: types.ITEM_WEAPON, GoldCost: 400, WearLoc: types.WEAR_NONE,
		Value: [6]int{5, 0, 0, 0, 0, 0}, // damaged so appraise has something to quote
	}
	handler.ObjToChar(obj, ch)

	DoAppraise(ch, "sword")
	out := readOutput(ch, client)
	if ch.Gold != 500 {
		t.Errorf("appraise should not charge; gold = %d", ch.Gold)
	}
	if !strings.Contains(out, "cost") {
		t.Errorf("expected cost quote; got %q", out)
	}
}
