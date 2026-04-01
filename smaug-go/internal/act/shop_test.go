package act

import (
	"strings"
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

func TestDoList_WithItems(t *testing.T) {
	ch, _ := newShopRoom()

	// Use makeTestChar to get a descriptor for output capture
	chWithDesc, client := makeTestChar("Buyer")
	defer client.Close()
	chWithDesc.InRoom = ch.InRoom
	chWithDesc.Gold = 500
	ch.InRoom.People = append(ch.InRoom.People, chWithDesc)

	DoList(chWithDesc, "")
	out := readOutput(chWithDesc, client)

	if !strings.Contains(out, "steel sword") {
		t.Errorf("expected item in list, got: %q", out)
	}
	// Cost should be 150 (100 * 150%)
	if !strings.Contains(out, "150") {
		t.Errorf("expected cost of 150, got: %q", out)
	}
}

func TestDoList_NoKeeper(t *testing.T) {
	_ = setupCommWorld()
	ch, client := makeTestChar("Buyer")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 8000, Name: "Empty Room"}

	DoList(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "can't do that here") {
		t.Errorf("expected 'can't do that here', got: %q", out)
	}
}

func TestDoList_EmptyShop(t *testing.T) {
	_ = setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8001, Name: "Empty Shop"}

	shop := &types.ShopData{ProfitBuy: 150, ProfitSell: 50}
	keeperIdx := &types.MobIndexData{Vnum: 200, PlayerName: "merchant", ShortDescr: "a merchant", Shop: shop}
	keeper := &types.CharData{
		Name: "merchant", ShortDescr: "a merchant",
		Level: 30, Position: types.POS_STANDING,
		InRoom: room, IndexData: keeperIdx,
	}
	keeper.Act.Set(types.ACT_IS_NPC)
	room.People = append(room.People, keeper)

	ch, client := makeTestChar("Buyer")
	defer client.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	DoList(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "nothing for sale") {
		t.Errorf("expected 'nothing for sale', got: %q", out)
	}
}

func TestDoValue_NoArg(t *testing.T) {
	_ = setupCommWorld()
	ch, client := makeTestChar("Seller")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 8002, Name: "Test"}

	DoValue(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Value what?") {
		t.Errorf("expected 'Value what?', got: %q", out)
	}
}

func TestDoValue_ShowsPrice(t *testing.T) {
	ch, _ := newShopRoom()
	chWithDesc, client := makeTestChar("Seller")
	defer client.Close()
	chWithDesc.InRoom = ch.InRoom
	ch.InRoom.People = append(ch.InRoom.People, chWithDesc)

	sword := &types.ObjData{
		Name: "axe battle", ShortDescr: "a battle axe",
		ItemType: types.ITEM_WEAPON, GoldCost: 200,
		CarriedBy: chWithDesc, WearLoc: types.WEAR_NONE,
	}
	chWithDesc.Carrying = append(chWithDesc.Carrying, sword)

	DoValue(chWithDesc, "axe")
	out := readOutput(chWithDesc, client)

	// Sell price: 200 * 50% = 100
	if !strings.Contains(out, "100") {
		t.Errorf("expected sell price of 100, got: %q", out)
	}
}

func TestDoValue_WontBuyType(t *testing.T) {
	ch, _ := newShopRoom()
	chWithDesc, client := makeTestChar("Seller")
	defer client.Close()
	chWithDesc.InRoom = ch.InRoom
	ch.InRoom.People = append(ch.InRoom.People, chWithDesc)

	food := &types.ObjData{
		Name: "bread", ShortDescr: "a loaf of bread",
		ItemType: types.ITEM_FOOD, GoldCost: 5,
		CarriedBy: chWithDesc, WearLoc: types.WEAR_NONE,
	}
	chWithDesc.Carrying = append(chWithDesc.Carrying, food)

	DoValue(chWithDesc, "bread")
	out := readOutput(chWithDesc, client)
	if !strings.Contains(out, "don't buy that kind") {
		t.Errorf("expected rejection message, got: %q", out)
	}
}

func TestDoValue_NoKeeper(t *testing.T) {
	_ = setupCommWorld()
	ch, client := makeTestChar("Seller")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 8003, Name: "Empty Room"}

	DoValue(ch, "sword")
	out := readOutput(ch, client)
	if !strings.Contains(out, "can't do that here") {
		t.Errorf("expected 'can't do that here', got: %q", out)
	}
}

func TestDoValue_DontHaveItem(t *testing.T) {
	ch, _ := newShopRoom()
	chWithDesc, client := makeTestChar("Seller")
	defer client.Close()
	chWithDesc.InRoom = ch.InRoom
	ch.InRoom.People = append(ch.InRoom.People, chWithDesc)

	DoValue(chWithDesc, "nonexistent")
	out := readOutput(chWithDesc, client)
	if !strings.Contains(out, "don't have") {
		t.Errorf("expected 'don't have', got: %q", out)
	}
}

func TestDoBuy_NoArg(t *testing.T) {
	_ = setupCommWorld()
	ch, client := makeTestChar("Buyer")
	defer client.Close()

	DoBuy(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Buy what?") {
		t.Errorf("expected 'Buy what?', got: %q", out)
	}
}

func TestDoBuy_ItemNotSold(t *testing.T) {
	ch, _ := newShopRoom()
	chWithDesc, client := makeTestChar("Buyer")
	defer client.Close()
	chWithDesc.InRoom = ch.InRoom
	chWithDesc.Gold = 10000
	ch.InRoom.People = append(ch.InRoom.People, chWithDesc)

	DoBuy(chWithDesc, "nonexistent")
	out := readOutput(chWithDesc, client)
	if !strings.Contains(out, "don't sell that") {
		t.Errorf("expected 'don't sell that', got: %q", out)
	}
}

func TestDoSell_NoArg(t *testing.T) {
	_ = setupCommWorld()
	ch, client := makeTestChar("Seller")
	defer client.Close()

	DoSell(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Sell what?") {
		t.Errorf("expected 'Sell what?', got: %q", out)
	}
}

func TestDoSell_NoKeeper(t *testing.T) {
	_ = setupCommWorld()
	ch, client := makeTestChar("Seller")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 8004, Name: "Empty Room"}

	DoSell(ch, "sword")
	out := readOutput(ch, client)
	if !strings.Contains(out, "can't do that here") {
		t.Errorf("expected 'can't do that here', got: %q", out)
	}
}

func TestDoSell_DontHaveItem(t *testing.T) {
	ch, _ := newShopRoom()
	chWithDesc, client := makeTestChar("Seller")
	defer client.Close()
	chWithDesc.InRoom = ch.InRoom
	ch.InRoom.People = append(ch.InRoom.People, chWithDesc)

	DoSell(chWithDesc, "nonexistent")
	out := readOutput(chWithDesc, client)
	if !strings.Contains(out, "don't have") {
		t.Errorf("expected 'don't have', got: %q", out)
	}
}

func TestGetShopCost_DefaultProfit(t *testing.T) {
	keeperIdx := &types.MobIndexData{
		Shop: &types.ShopData{ProfitBuy: 0}, // 0 means use default 120%
	}
	keeper := &types.CharData{IndexData: keeperIdx}
	obj := &types.ObjData{GoldCost: 100}

	cost := getShopCost(keeper, obj)
	if cost != 120 {
		t.Errorf("expected default 120, got %d", cost)
	}
}

func TestGetSellPrice_DefaultProfit(t *testing.T) {
	keeperIdx := &types.MobIndexData{
		Shop: &types.ShopData{ProfitSell: 0}, // 0 means use default 50%
	}
	keeper := &types.CharData{IndexData: keeperIdx}
	obj := &types.ObjData{GoldCost: 100}

	price := getSellPrice(keeper, obj)
	if price != 50 {
		t.Errorf("expected default 50, got %d", price)
	}
}

func TestGetSellPrice_MinimumOne(t *testing.T) {
	keeperIdx := &types.MobIndexData{
		Shop: &types.ShopData{ProfitSell: 1}, // 1% of 1 gold = 0, should clamp to 1
	}
	keeper := &types.CharData{IndexData: keeperIdx}
	obj := &types.ObjData{GoldCost: 1}

	price := getSellPrice(keeper, obj)
	if price < 1 {
		t.Errorf("expected minimum price of 1, got %d", price)
	}
}

func TestGetShopCost_NoShop(t *testing.T) {
	keeper := &types.CharData{IndexData: &types.MobIndexData{}}
	obj := &types.ObjData{GoldCost: 100}

	cost := getShopCost(keeper, obj)
	if cost != 100 {
		t.Errorf("expected raw cost 100 with no shop, got %d", cost)
	}
}

func TestGetSellPrice_NoShop(t *testing.T) {
	keeper := &types.CharData{IndexData: &types.MobIndexData{}}
	obj := &types.ObjData{GoldCost: 100}

	price := getSellPrice(keeper, obj)
	if price != 50 {
		t.Errorf("expected half cost 50 with no shop, got %d", price)
	}
}

func TestShopBuysType(t *testing.T) {
	shop := &types.ShopData{BuyType: [5]int{types.ITEM_WEAPON, types.ITEM_ARMOR, 0, 0, 0}}

	if !shopBuysType(shop, types.ITEM_WEAPON) {
		t.Error("shop should buy weapons")
	}
	if !shopBuysType(shop, types.ITEM_ARMOR) {
		t.Error("shop should buy armor")
	}
	if shopBuysType(shop, types.ITEM_FOOD) {
		t.Error("shop should not buy food")
	}
}
