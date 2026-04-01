package act

import (
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// findKeeper finds a shopkeeper NPC in the character's room.
func findKeeper(ch *types.CharData) *types.CharData {
	if ch.InRoom == nil {
		return nil
	}
	for _, rch := range ch.InRoom.People {
		if rch.IsNPC() && rch.IndexData != nil && rch.IndexData.Shop != nil {
			return rch
		}
	}
	return nil
}

// getShopCost calculates buy price based on shop profit margin.
func getShopCost(keeper *types.CharData, obj *types.ObjData) int {
	if keeper.IndexData == nil || keeper.IndexData.Shop == nil {
		return obj.GoldCost
	}
	profit := keeper.IndexData.Shop.ProfitBuy
	if profit <= 0 {
		profit = 120 // default 120%
	}
	return obj.GoldCost * profit / 100
}

// getSellPrice calculates sell price based on shop profit margin.
func getSellPrice(keeper *types.CharData, obj *types.ObjData) int {
	if keeper.IndexData == nil || keeper.IndexData.Shop == nil {
		return obj.GoldCost / 2
	}
	profit := keeper.IndexData.Shop.ProfitSell
	if profit <= 0 {
		profit = 50 // default 50%
	}
	cost := obj.GoldCost * profit / 100
	if cost < 1 {
		cost = 1
	}
	return cost
}

// shopBuysType returns true if the shop buys items of the given type.
func shopBuysType(shop *types.ShopData, itemType int) bool {
	for _, bt := range shop.BuyType {
		if bt == itemType {
			return true
		}
	}
	return false
}

// DoBuy implements the 'buy' command: purchase an item from a shopkeeper.
func DoBuy(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Buy what?\n\r")
		return
	}

	keeper := findKeeper(ch)
	if keeper == nil {
		ch.Send("You can't do that here.\n\r")
		return
	}

	// Find the item in the keeper's inventory
	obj := handler.GetObjCarry(keeper, arg)
	if obj == nil {
		ch.Sendf("%s tells you 'I don't sell that — try \"list\"'.\n\r", keeper.ShortDescr)
		return
	}

	cost := getShopCost(keeper, obj)
	if ch.Gold < cost {
		ch.Sendf("%s tells you 'You can't afford it. It costs %d gold.'.\n\r",
			keeper.ShortDescr, cost)
		return
	}

	ch.Gold -= cost
	keeper.Gold += cost

	// Transfer object to buyer
	handler.ObjFromChar(obj)
	obj.CarriedBy = ch
	obj.WearLoc = types.WEAR_NONE
	ch.Carrying = append(ch.Carrying, obj)

	ch.Sendf("You buy %s for %d gold.\n\r", obj.ShortDescr, cost)
	ch.Sendf("%s tells you 'Pleasure doing business with you.'.\n\r", keeper.ShortDescr)
}

// DoSell implements the 'sell' command: sell an item to a shopkeeper.
func DoSell(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Sell what?\n\r")
		return
	}

	keeper := findKeeper(ch)
	if keeper == nil {
		ch.Send("You can't do that here.\n\r")
		return
	}

	obj := handler.GetObjCarry(ch, arg)
	if obj == nil {
		ch.Send("You don't have that item.\n\r")
		return
	}

	if keeper.IndexData.Shop != nil && !shopBuysType(keeper.IndexData.Shop, obj.ItemType) {
		ch.Sendf("%s tells you 'I don't buy that kind of thing.'.\n\r", keeper.ShortDescr)
		return
	}

	cost := getSellPrice(keeper, obj)

	ch.Gold += cost

	// Transfer to keeper
	handler.ObjFromChar(obj)
	obj.CarriedBy = keeper
	obj.WearLoc = types.WEAR_NONE
	keeper.Carrying = append(keeper.Carrying, obj)

	ch.Sendf("You sell %s for %d gold.\n\r", obj.ShortDescr, cost)
}

// DoList implements the 'list' command: show items for sale.
func DoList(ch *types.CharData, argument string) {
	keeper := findKeeper(ch)
	if keeper == nil {
		ch.Send("You can't do that here.\n\r")
		return
	}

	found := false
	for _, obj := range keeper.Carrying {
		if obj.WearLoc != types.WEAR_NONE {
			continue
		}
		cost := getShopCost(keeper, obj)
		if !found {
			ch.Sendf("&W[Cost    ] Item&D\n\r")
			found = true
		}
		ch.Sendf("[%-8d] %s\n\r", cost, obj.ShortDescr)
	}
	if !found {
		ch.Sendf("%s tells you 'I have nothing for sale right now.'.\n\r", keeper.ShortDescr)
	}
}

// DoValue implements the 'value' command: see what a shopkeeper will pay.
func DoValue(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Value what?\n\r")
		return
	}

	keeper := findKeeper(ch)
	if keeper == nil {
		ch.Send("You can't do that here.\n\r")
		return
	}

	obj := handler.GetObjCarry(ch, arg)
	if obj == nil {
		ch.Send("You don't have that item.\n\r")
		return
	}

	if keeper.IndexData.Shop != nil && !shopBuysType(keeper.IndexData.Shop, obj.ItemType) {
		ch.Sendf("%s tells you 'I don't buy that kind of thing.'.\n\r", keeper.ShortDescr)
		return
	}

	cost := getSellPrice(keeper, obj)
	ch.Sendf("%s tells you 'I'd give you %d gold for %s.'.\n\r",
		keeper.ShortDescr, cost, obj.ShortDescr)
}
