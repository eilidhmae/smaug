package game

import (
	"github.com/eilidhmae/smaug/internal/act"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

// clearAuctionOnDisconnect is the Go-originated defensive clear that
// fires when a descriptor closes without going through DoQuit — for
// example peer reset, SIGKILL, or any hot-path drop before the auction
// gate. Complements the DoQuit gate ported from C
// `do_quit` at src/act_comm.c:2883-2890. See plan-phase6-auction.md §D8.
//
// The seller-drop path cancels the auction and deposits the item in the
// seller's last room (so it's not lost); any distinct buyer gets
// refunded their standing bid. The buyer-drop path rolls the auction
// back to the starting bet and seller-as-placeholder-buyer so a
// subsequent valid bidder can still win. History ring is preserved
// across both cases so the spam guard survives.
func (g *GameLoop) clearAuctionOnDisconnect(ch *types.CharData) {
	if g.world == nil || g.world.Auction == nil {
		return
	}
	auc := g.world.Auction
	if auc.Item == nil {
		return
	}

	if auc.Seller == ch {
		// Seller dropout: cancel + return item to ch's last room (if any)
		// + refund distinct buyer.
		act.BroadcastAuction("The auction has been cancelled — the seller has left the game.")
		if ch.InRoom != nil {
			handler.ObjToRoom(auc.Item, ch.InRoom)
		}
		if auc.Buyer != nil && auc.Buyer != ch {
			auc.Buyer.Gold += auc.Bet
			auc.Buyer.Send("The auction was cancelled. Your bid has been returned.\n\r")
		}
		auc.Item = nil
		auc.Seller = nil
		auc.Buyer = nil
		auc.Bet = 0
		auc.Going = 0
		auc.Pulse = 0
		auc.Starting = 0
		return
	}

	if auc.Buyer == ch {
		// Buyer dropout: rewind to starting bet, buyer=seller placeholder,
		// restart the Going counter so late-arriving bidders can still
		// reach case-1/2/3 progression.
		act.BroadcastAuction("The top bidder has left the game. Bidding re-opens at the starting price.")
		auc.Buyer = auc.Seller
		auc.Bet = auc.Starting
		auc.Going = 0
		auc.Pulse = types.PULSE_AUCTION
	}
}
