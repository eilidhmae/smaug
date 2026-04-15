package act

import (
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// findFixer locates an NPC with an RShop entry in ch's current room.
func findFixer(ch *types.CharData) *types.CharData {
	if ch.InRoom == nil {
		return nil
	}
	for _, rch := range ch.InRoom.People {
		if rch.IsNPC() && rch.IndexData != nil && rch.IndexData.RShop != nil {
			return rch
		}
	}
	return nil
}

// repairAccepts returns true if the keeper fixes this item type, matching
// `src/shops.c:get_repaircost` — only items listed in FixType qualify.
func repairAccepts(keeper *types.CharData, obj *types.ObjData) bool {
	if keeper.IndexData == nil || keeper.IndexData.RShop == nil {
		return false
	}
	for _, t := range keeper.IndexData.RShop.FixType {
		if t == obj.ItemType {
			return true
		}
	}
	return false
}

// INIT_WEAPON_CONDITION mirrors C src/handler.c — weapons start at condition 12.
const initWeaponCondition = 12

// repairDamageDelta returns how damaged the item is on an integer scale.
// 0 means "undamaged, refuse to charge" (C returns -2 with "looks fine").
// Negative means "not repairable" (C returns -1).
//
// Semantics mirror `src/shops.c:495`:
//   - ARMOR: value[0] is current condition, value[1] is max; damage lowers
//     value[0], so damaged means value[0] < value[1].
//   - WEAPON: value[0] is current condition, capped at initWeaponCondition;
//     anything equal to init is pristine.
//   - WAND/STAFF: value[1] is max charges, value[2] is current charges.
func repairDamageDelta(obj *types.ObjData) int {
	switch obj.ItemType {
	case types.ITEM_ARMOR:
		if obj.Value[0] >= obj.Value[1] {
			return 0
		}
		return obj.Value[1] - obj.Value[0]
	case types.ITEM_WEAPON:
		if obj.Value[0] == initWeaponCondition {
			return 0
		}
		return initWeaponCondition - obj.Value[0]
	case types.ITEM_WAND, types.ITEM_STAFF:
		if obj.Value[2] >= obj.Value[1] {
			return 0
		}
		return obj.Value[1] - obj.Value[2]
	}
	return -1
}

// repairCost mirrors C `src/shops.c:503` — cost scales with damage delta.
// Minimum 1 gold on any chargeable repair.
func repairCost(keeper *types.CharData, obj *types.ObjData) int {
	delta := repairDamageDelta(obj)
	if delta <= 0 {
		return 0
	}
	profit := keeper.IndexData.RShop.ProfitFix
	if profit <= 0 {
		profit = 100
	}
	cost := obj.GoldCost * profit * delta / 1000
	if cost < 1 {
		cost = 1
	}
	return cost
}

// DoRepair — estimate or execute a repair/recharge at an RShop keeper.
// Usage: repair           (list items the keeper can fix)
//        repair <item>    (carry out the repair, paying in gold)
//        repair estimate <item> (quote without paying)
func DoRepair(ch *types.CharData, argument string) {
	keeper := findFixer(ch)
	if keeper == nil {
		ch.Send("You can't do that here.\n\r")
		return
	}

	arg1, arg2 := util.OneArgument(argument)
	if arg1 == "" {
		listRepairable(ch, keeper)
		return
	}

	estimate := false
	if arg1 == "estimate" || arg1 == "appraise" {
		estimate = true
		arg1 = arg2
		if arg1 == "" {
			ch.Send("Estimate what?\n\r")
			return
		}
	}

	obj := handler.GetObjCarry(ch, arg1)
	if obj == nil {
		ch.Sendf("%s tells you 'You don't have that.'\n\r", keeper.ShortDescr)
		return
	}
	if !repairAccepts(keeper, obj) {
		ch.Sendf("%s tells you 'I can't do anything with %s.'\n\r", keeper.ShortDescr, obj.ShortDescr)
		return
	}
	if repairDamageDelta(obj) == 0 {
		ch.Sendf("%s tells you '%s looks fine to me!'\n\r", keeper.ShortDescr, obj.ShortDescr)
		return
	}

	cost := repairCost(keeper, obj)
	if estimate {
		ch.Sendf("%s tells you 'It will cost %d gold to repair %s.'\n\r",
			keeper.ShortDescr, cost, obj.ShortDescr)
		return
	}

	if ch.Gold < cost {
		ch.Sendf("%s tells you 'You can't afford it — %d gold.'\n\r", keeper.ShortDescr, cost)
		return
	}

	ch.Gold -= cost
	keeper.Gold += cost

	// Apply repair. Mirrors `src/shops.c` restoration block.
	switch obj.ItemType {
	case types.ITEM_WAND, types.ITEM_STAFF:
		obj.Value[2] = obj.Value[1]
	case types.ITEM_WEAPON:
		obj.Value[0] = initWeaponCondition
	case types.ITEM_ARMOR:
		obj.Value[0] = obj.Value[1]
	}

	verb := "repairs"
	if keeper.IndexData.RShop.ShopType == types.SHOP_RECHARGE {
		verb = "recharges"
	}
	ch.Sendf("%s %s %s for %d gold.\n\r", keeper.ShortDescr, verb, obj.ShortDescr, cost)
}

// DoAppraise aliases `repair estimate` for legacy muscle memory.
func DoAppraise(ch *types.CharData, argument string) {
	if argument == "" {
		ch.Send("Appraise what?\n\r")
		return
	}
	DoRepair(ch, "estimate "+argument)
}

func listRepairable(ch *types.CharData, keeper *types.CharData) {
	found := false
	for _, obj := range ch.Carrying {
		if obj.WearLoc != types.WEAR_NONE {
			continue
		}
		if !repairAccepts(keeper, obj) || repairDamageDelta(obj) == 0 {
			continue
		}
		if !found {
			ch.Sendf("%s tells you 'I can work on:'\n\r", keeper.ShortDescr)
			found = true
		}
		ch.Sendf("  %-30s  %d gold\n\r", obj.ShortDescr, repairCost(keeper, obj))
	}
	if !found {
		ch.Sendf("%s tells you 'Bring me something I can work on.'\n\r", keeper.ShortDescr)
	}
}
