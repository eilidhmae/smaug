package act

import (
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// repairCostCodes returned by computeRepairCost. Negative codes mirror C's
// get_repaircost sentinel values ("-1" unrepairable, "-2" already pristine).
const (
	repairCostUnrepairable = -1
	repairCostPristine     = -2
)

// findFixer scans ch.InRoom for an NPC with an attached RepairData. Returns
// nil when none is present. Mirrors C find_fixer at src/shops.c:202.
func findFixer(ch *types.CharData) *types.CharData {
	if ch == nil || ch.InRoom == nil {
		return nil
	}
	for _, rch := range ch.InRoom.People {
		if rch.IsNPC() && rch.IndexData != nil && rch.IndexData.RShop != nil {
			return rch
		}
	}
	return nil
}

// computeRepairCost mirrors C's get_repaircost: returns the gold price of
// restoring obj to full condition at keeper, or a negative sentinel when
// repair is impossible / not needed. The formula is:
//
//	baseCost = obj.GoldCost * keeper.RShop.ProfitFix / 1000
//	scaled  *= (max - current) condition delta per item type
//
// Condition tracking follows C: value[0]=current condition, value[1]=max
// for ARMOR / WAND / STAFF, and value[0]=current for WEAPON with
// INIT_WEAPON_CONDITION as the ceiling.
func computeRepairCost(keeper *types.CharData, obj *types.ObjData) int {
	if keeper == nil || obj == nil {
		return repairCostUnrepairable
	}
	if keeper.IndexData == nil || keeper.IndexData.RShop == nil {
		return repairCostUnrepairable
	}
	r := keeper.IndexData.RShop

	found := false
	for _, t := range r.FixType {
		if t == obj.ItemType {
			found = true
			break
		}
	}
	if !found {
		return repairCostUnrepairable
	}

	base := obj.GoldCost
	if base == 0 && obj.IndexData != nil {
		base = obj.IndexData.GoldCost
	}
	if base <= 0 {
		base = 1
	}
	profit := r.ProfitFix
	if profit <= 0 {
		profit = 100
	}
	cost := base * profit / 1000
	if cost <= 0 {
		cost = 1
	}

	switch obj.ItemType {
	case types.ITEM_ARMOR:
		if obj.Value[0] >= obj.Value[1] {
			return repairCostPristine
		}
		cost *= obj.Value[1] - obj.Value[0]
	case types.ITEM_WEAPON:
		if obj.Value[0] >= types.INIT_WEAPON_CONDITION {
			return repairCostPristine
		}
		cost *= types.INIT_WEAPON_CONDITION - obj.Value[0]
	case types.ITEM_WAND, types.ITEM_STAFF:
		if obj.Value[2] >= obj.Value[1] {
			return repairCostPristine
		}
		cost *= obj.Value[1] - obj.Value[2]
	default:
		return repairCostUnrepairable
	}
	if cost <= 0 {
		cost = 1
	}
	return cost
}

// restoreCondition brings obj's condition fields back to their maxima, mirror
// of the per-item-type assignments in C's repair_one_obj.
func restoreCondition(obj *types.ObjData) {
	switch obj.ItemType {
	case types.ITEM_ARMOR:
		obj.Value[0] = obj.Value[1]
	case types.ITEM_WEAPON:
		obj.Value[0] = types.INIT_WEAPON_CONDITION
	case types.ITEM_WAND, types.ITEM_STAFF:
		obj.Value[2] = obj.Value[1]
	}
}

// DoAppraise implements the 'appraise' command: estimate the repair cost of a
// carried item without spending gold.
func DoAppraise(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Appraise what?\n\r")
		return
	}
	keeper := findFixer(ch)
	if keeper == nil {
		ch.Send("You can't do that here.\n\r")
		return
	}
	obj := handler.GetObjCarry(ch, arg)
	if obj == nil {
		ch.Sendf("%s tells you 'You don't have that item.'\n\r", keeper.ShortDescr)
		return
	}
	cost := computeRepairCost(keeper, obj)
	switch cost {
	case repairCostUnrepairable:
		ch.Sendf("%s tells you, 'Sorry, I can't do anything with %s.'\n\r",
			keeper.ShortDescr, obj.ShortDescr)
	case repairCostPristine:
		ch.Sendf("%s tells you, '%s looks fine to me!'\n\r",
			keeper.ShortDescr, obj.ShortDescr)
	default:
		ch.Sendf("%s tells you, 'It will cost %d gold to repair %s.'\n\r",
			keeper.ShortDescr, cost, obj.ShortDescr)
	}
}

// DoRepair implements the 'repair' command: pay a fixer keeper to restore a
// carried item's condition. Mirrors C do_repair / repair_one_obj.
func DoRepair(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Repair what?\n\r")
		return
	}
	keeper := findFixer(ch)
	if keeper == nil {
		ch.Send("You can't do that here.\n\r")
		return
	}
	obj := handler.GetObjCarry(ch, arg)
	if obj == nil {
		ch.Sendf("%s tells you 'You don't have that item.'\n\r", keeper.ShortDescr)
		return
	}
	if !handler.CanDropObj(obj) {
		ch.Sendf("You can't let go of %s.\n\r", obj.ShortDescr)
		return
	}
	cost := computeRepairCost(keeper, obj)
	switch cost {
	case repairCostUnrepairable:
		ch.Sendf("%s tells you, 'Sorry, I can't do anything with %s.'\n\r",
			keeper.ShortDescr, obj.ShortDescr)
		return
	case repairCostPristine:
		ch.Sendf("%s tells you, '%s looks fine to me!'\n\r",
			keeper.ShortDescr, obj.ShortDescr)
		return
	}
	if ch.Gold < cost {
		ch.Sendf("%s tells you, 'It costs %d gold to repair %s, which I see you can't afford.'\n\r",
			keeper.ShortDescr, cost, obj.ShortDescr)
		return
	}

	ch.Gold -= cost
	keeper.Gold += cost
	restoreCondition(obj)
	ch.Sendf("%s repairs %s for %d gold.\n\r",
		keeper.ShortDescr, obj.ShortDescr, cost)
}
