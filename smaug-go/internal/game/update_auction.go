package game

import (
	"fmt"

	"github.com/eilidhmae/smaug/internal/act"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// auctionUpdate mirrors C auction_update at src/update.c:3183-3326.
// Called every PULSE_AUCTION (9 seconds) from the main pulse loop.
//
// Branch map:
//  1. No item: age the history ring every `6 * AUCTION_MEM` pulses,
//     pop oldest non-nil entry.
//  2. Going advances: 0 → 1 → 2 → 3.
//     - case 1 / case 2: broadcast "going once/twice" lines.
//     - case 3: resolve — sold (90% pay / 10% tax / buyer gets item) or
//     not-sold (seller reclaims / 5% tax on item.Cost, floor at 0).
//  3. Clear item at end of case 3.
func (g *GameLoop) auctionUpdate() {
	if g.world == nil || g.world.Auction == nil {
		return
	}
	auc := g.world.Auction

	if auc.Item == nil {
		auctionAgeHistory(auc)
		return
	}

	auc.Going++
	switch auc.Going {
	case 1, 2:
		auctionBroadcastGoing(auc)
	case 3:
		auctionResolve(auc)
	}
}

// auctionAgeHistory ports the no-item history-decay branch at
// update.c:3188-3205.
func auctionAgeHistory(auc *types.AuctionData) {
	if types.AUCTION_MEM <= 0 {
		return
	}
	if auc.History[0] == nil {
		return
	}
	auc.HistTimer++
	if auc.HistTimer != 6*types.AUCTION_MEM {
		return
	}
	// Pop the newest trailing non-nil entry (C walks from high index
	// downward and nils the first hit it encounters).
	for i := types.AUCTION_MEM - 1; i >= 0; i-- {
		if auc.History[i] != nil {
			auc.History[i] = nil
			auc.HistTimer = 0
			break
		}
	}
}

// auctionBroadcastGoing ports update.c:3210-3222.
func auctionBroadcastGoing(auc *types.AuctionData) {
	stage := "once"
	if auc.Going == 2 {
		stage = "twice"
	}
	var msg string
	if auc.Bet > auc.Starting {
		msg = fmt.Sprintf("%s: going %s for %s.",
			auc.Item.ShortDescr, stage, util.NumPunct(auc.Bet))
	} else {
		msg = fmt.Sprintf("%s: going %s (bid not received yet).",
			auc.Item.ShortDescr, stage)
	}
	act.BroadcastAuction(msg)
}

// auctionResolve ports the case-3 resolution at update.c:3224-3324.
func auctionResolve(auc *types.AuctionData) {
	// Defensive: if buyer went nil during the auction (C defensive log
	// at update.c:3225-3230), zero the bet and fall through to the
	// not-sold branch.
	if auc.Buyer == nil && auc.Bet != 0 {
		util.Bug("Auction code reached SOLD, with NULL buyer, but %d gold bid", auc.Bet)
		auc.Bet = 0
	}

	sold := auc.Bet > 0 && auc.Buyer != nil && auc.Buyer != auc.Seller

	if sold {
		auctionResolveSold(auc)
	} else {
		auctionResolveUnsold(auc)
	}

	auc.Item = nil
	auc.Buyer = nil
	auc.Seller = nil
	auc.Bet = 0
	auc.Going = 0
	auc.Pulse = 0
	auc.Starting = 0
}

func auctionResolveSold(auc *types.AuctionData) {
	buyerName := auc.Buyer.Name
	if auc.Buyer.IsNPC() {
		buyerName = auc.Buyer.ShortDescr
	}
	msg := fmt.Sprintf("%s sold to %s for %s.",
		auc.Item.ShortDescr, buyerName, util.NumPunct(auc.Bet))
	act.BroadcastAuction(msg)

	util.Act(types.AT_ACTION,
		"The auctioneer materializes before you, and hands you $p.",
		auc.Buyer, nil, auc.Item, nil, types.TO_CHAR)
	util.Act(types.AT_ACTION,
		"The auctioneer materializes before $n, and hands $m $p.",
		auc.Buyer, nil, auc.Item, nil, types.TO_ROOM)

	// C has a carry-weight cap here (update.c:3247-3262); Go has not
	// modeled CanCarryW (plan §Q4). Always give to buyer.
	handler.ObjToChar(auc.Item, auc.Buyer)

	// Integer-arithmetic tax split. C uses (int)(bet*0.9)+(int)(bet*0.1)
	// which can drift 1 gold on certain bet values; plan §Q7 fixes by
	// computing pay first and letting tax absorb the remainder. So
	// `pay + tax == bet` invariant holds for any positive bet.
	pay := auc.Bet * 9 / 10
	tax := auc.Bet - pay
	if auc.Seller != nil {
		auc.Seller.Gold += pay
		if auc.Seller.InRoom != nil {
			handler.BoostEconomy(auc.Seller.InRoom.Area, tax)
		}
		auc.Seller.Sendf("The auctioneer pays you %s gold, charging an auction fee of %s.\n\r",
			util.NumPunct(pay), util.NumPunct(tax))
		auc.Seller.Sendf("%s\n\r", msg)
	}
}

func auctionResolveUnsold(auc *types.AuctionData) {
	msg := fmt.Sprintf("No bids received for %s - removed from auction.",
		auc.Item.ShortDescr)
	act.BroadcastAuction(msg)

	if auc.Seller != nil {
		util.Act(types.AT_ACTION,
			"The auctioneer appears before you to return $p to you.",
			auc.Seller, nil, auc.Item, nil, types.TO_CHAR)
		util.Act(types.AT_ACTION,
			"The auctioneer appears before $n to return $p to $m.",
			auc.Seller, nil, auc.Item, nil, types.TO_ROOM)

		// C carry-weight branch (update.c:3293-3308); Go: always return
		// to seller per plan §Q4.
		handler.ObjToChar(auc.Item, auc.Seller)

		tax := auc.Item.GoldCost * 5 / 100
		if auc.Seller.InRoom != nil {
			handler.BoostEconomy(auc.Seller.InRoom.Area, tax)
		}
		auc.Seller.Sendf("The auctioneer charges you an auction fee of %s.\n\r",
			util.NumPunct(tax))
		// C: if seller's gold would go negative, zero it (update.c:3317-3320).
		if auc.Seller.Gold-tax < 0 {
			auc.Seller.Gold = 0
		} else {
			auc.Seller.Gold -= tax
		}
	}
}
