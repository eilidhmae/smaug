package types

import "testing"

// AuctionData defaults cleanly — zero-value has empty history ring
// and no timer, matching C's `auction_update` early-return behavior.
func TestAuctionData_ZeroHistoryRing(t *testing.T) {
	a := AuctionData{}
	for i, h := range a.History {
		if h != nil {
			t.Errorf("AuctionData{}.History[%d] = %v, want nil", i, h)
		}
	}
	if a.HistTimer != 0 {
		t.Errorf("AuctionData{}.HistTimer = %d, want 0", a.HistTimer)
	}
}

// History ring size pinned to AUCTION_MEM — mutation-verify anchor.
func TestAuctionData_FixedSize(t *testing.T) {
	a := AuctionData{}
	if len(a.History) != AUCTION_MEM {
		t.Errorf("len(History) = %d, want AUCTION_MEM (%d)", len(a.History), AUCTION_MEM)
	}
}
