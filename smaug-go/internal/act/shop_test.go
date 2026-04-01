package act

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

func newShopRoom() (*types.CharData, *types.CharData) {
	room := &types.RoomIndexData{Vnum: 5000, Name: "Shop"}

	shop := &types.ShopData{
		Keeper:     100,
		ProfitBuy:  150,
		ProfitSell: 50,
		BuyType:    [5]int{types.ITEM_WEAPON, types.ITEM_ARMOR, 0, 0, 0},
	}
	keeperIdx := &types.MobIndexData{
		Vnum:       100,
		PlayerName: "shopkeeper",
		ShortDescr: "a shopkeeper",
		Shop:       shop,
	}
	keeper := &types.CharData{
		Name:       "shopkeeper",
		ShortDescr: "a shopkeeper",
		Level:      30,
		Position:   types.POS_STANDING,
		Gold:       10000,
		InRoom:     room,
		IndexData:  keeperIdx,
	}
	keeper.Act.Set(types.ACT_IS_NPC)

	// Give keeper some items to sell
	sword := &types.ObjData{
		Name:       "sword steel",
		ShortDescr: "a steel sword",
		ItemType:   types.ITEM_WEAPON,
		GoldCost:   100,
		CarriedBy:  keeper,
		WearLoc:    types.WEAR_NONE,
	}
	keeper.Carrying = append(keeper.Carrying, sword)

	room.People = append(room.People, keeper)

	ch := newTestCharWithDesc()
	ch.InRoom = room
	ch.Gold = 500
	room.People = append(room.People, ch)

	return ch, keeper
}

func TestDoBuy_Success(t *testing.T) {
	ch, keeper := newShopRoom()
	goldBefore := ch.Gold

	DoBuy(ch, "sword")

	if ch.Gold >= goldBefore {
		t.Error("gold should decrease after buying")
	}
	// Player should have the sword
	found := false
	for _, o := range ch.Carrying {
		if o.Name == "sword steel" {
			found = true
		}
	}
	if !found {
		t.Error("player should have the sword after buying")
	}
	// Keeper should not have it
	for _, o := range keeper.Carrying {
		if o.Name == "sword steel" {
			t.Error("keeper should not have the sword after selling")
		}
	}
}

func TestDoBuy_CantAfford(t *testing.T) {
	ch, _ := newShopRoom()
	ch.Gold = 0

	DoBuy(ch, "sword")

	// Should still have 0 gold and no sword
	if ch.Gold != 0 {
		t.Error("gold should not change when can't afford")
	}
}

func TestDoBuy_NoKeeper(t *testing.T) {
	ch := newTestCharWithDesc()
	DoBuy(ch, "sword")
	// Should say "can't do that here"
}

func TestDoSell_Success(t *testing.T) {
	ch, keeper := newShopRoom()
	// Give player a weapon to sell
	shield := &types.ObjData{
		Name:       "shield wooden",
		ShortDescr: "a wooden shield",
		ItemType:   types.ITEM_ARMOR,
		GoldCost:   50,
		CarriedBy:  ch,
		WearLoc:    types.WEAR_NONE,
	}
	ch.Carrying = append(ch.Carrying, shield)

	goldBefore := ch.Gold
	DoSell(ch, "shield")

	if ch.Gold <= goldBefore {
		t.Error("gold should increase after selling")
	}
	// Shield should be with keeper now
	found := false
	for _, o := range keeper.Carrying {
		if o.Name == "shield wooden" {
			found = true
		}
	}
	if !found {
		t.Error("keeper should have the shield after buying it")
	}
}

func TestDoSell_WontBuyType(t *testing.T) {
	ch, _ := newShopRoom()
	// Food is not in the shop's buy types
	food := &types.ObjData{
		Name:       "bread",
		ShortDescr: "a loaf of bread",
		ItemType:   types.ITEM_FOOD,
		GoldCost:   5,
		CarriedBy:  ch,
		WearLoc:    types.WEAR_NONE,
	}
	ch.Carrying = append(ch.Carrying, food)

	goldBefore := ch.Gold
	DoSell(ch, "bread")

	if ch.Gold != goldBefore {
		t.Error("gold should not change when shop won't buy the type")
	}
}

func TestDoList(t *testing.T) {
	ch, _ := newShopRoom()
	DoList(ch, "")
	// Just verify it doesn't crash
}

func TestDoValue(t *testing.T) {
	ch, _ := newShopRoom()
	sword := &types.ObjData{
		Name:       "sword old",
		ShortDescr: "an old sword",
		ItemType:   types.ITEM_WEAPON,
		GoldCost:   80,
		CarriedBy:  ch,
		WearLoc:    types.WEAR_NONE,
	}
	ch.Carrying = append(ch.Carrying, sword)

	DoValue(ch, "old")
	// Just verify it doesn't crash
}

func TestGetShopCost(t *testing.T) {
	keeperIdx := &types.MobIndexData{
		Shop: &types.ShopData{ProfitBuy: 150},
	}
	keeper := &types.CharData{IndexData: keeperIdx}
	obj := &types.ObjData{GoldCost: 100}

	cost := getShopCost(keeper, obj)
	if cost != 150 {
		t.Errorf("expected 150, got %d", cost)
	}
}

func TestGetSellPrice(t *testing.T) {
	keeperIdx := &types.MobIndexData{
		Shop: &types.ShopData{ProfitSell: 50},
	}
	keeper := &types.CharData{IndexData: keeperIdx}
	obj := &types.ObjData{GoldCost: 100}

	price := getSellPrice(keeper, obj)
	if price != 50 {
		t.Errorf("expected 50, got %d", price)
	}
}
