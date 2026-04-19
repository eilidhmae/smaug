package game

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/act"
	"github.com/eilidhmae/smaug/internal/command"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// newAuctionTestLoop builds a fresh loop with act.WorldRef wired so
// BroadcastAuction and DoAuction see the same world the tick does.
func newAuctionTestLoop(t *testing.T) *GameLoop {
	t.Helper()
	w := world.New("/tmp/test-auction")
	w.TimeInfo.Hour = 12
	act.WorldRef = w
	reg := command.NewRegistry()
	incoming := make(chan *types.DescriptorData, 4)
	return NewGameLoop(w, reg, incoming)
}

func makeTickWeapon(short string) (*types.ObjData, *types.ObjIndexData) {
	idx := &types.ObjIndexData{Vnum: 20000, Name: "sword", ShortDescr: short, ItemType: types.ITEM_WEAPON}
	obj := &types.ObjData{
		Name: "sword", ShortDescr: short, ItemType: types.ITEM_WEAPON,
		GoldCost: 2000, IndexData: idx, WearLoc: types.WEAR_NONE,
	}
	return obj, idx
}

// Placeholder chars — we need concrete CharData values; the tick reads
// .Name, .InRoom, .Gold, .Send* methods. Keep minimal.
func makeTickChar(name string) *types.CharData {
	return &types.CharData{Name: name, ShortDescr: name}
}

// Empty auction: tick should be a no-op — no broadcast, no crash.
func TestAuctionUpdate_NoItem_NoBroadcast(t *testing.T) {
	g := newAuctionTestLoop(t)
	// Auction is initialized zero-valued by world.New per G2.
	g.auctionUpdate() // must not panic
	if g.world.Auction.Item != nil {
		t.Errorf("Item should stay nil")
	}
}

// No item + history entries age: after 18 ticks (6 × AUCTION_MEM) the
// oldest entry gets popped, timer resets.
func TestAuctionUpdate_NoItem_HistoryDecay(t *testing.T) {
	g := newAuctionTestLoop(t)
	idx := &types.ObjIndexData{Vnum: 55}
	g.world.Auction.History[0] = idx

	// 18 ticks = 6 * AUCTION_MEM (3).
	for i := 0; i < 6*types.AUCTION_MEM-1; i++ {
		g.auctionUpdate()
		if g.world.Auction.History[0] == nil {
			t.Fatalf("history cleared too early at tick %d", i)
		}
	}
	// The 18th tick should pop.
	g.auctionUpdate()
	if g.world.Auction.History[0] != nil {
		t.Errorf("History[0] should be nil after 18 empty-auction ticks; got %v", g.world.Auction.History[0])
	}
	if g.world.Auction.HistTimer != 0 {
		t.Errorf("HistTimer should reset to 0; got %d", g.world.Auction.HistTimer)
	}
}

func TestAuctionUpdate_Going1(t *testing.T) {
	g := newAuctionTestLoop(t)
	obj, _ := makeTickWeapon("a sword")
	g.world.Auction.Item = obj
	g.world.Auction.Starting = 100
	g.world.Auction.Bet = 1000

	g.auctionUpdate()
	if g.world.Auction.Going != 1 {
		t.Errorf("Going = %d, want 1", g.world.Auction.Going)
	}
}

func TestAuctionUpdate_Going2(t *testing.T) {
	g := newAuctionTestLoop(t)
	obj, _ := makeTickWeapon("a sword")
	g.world.Auction.Item = obj
	g.world.Auction.Going = 1
	g.world.Auction.Starting = 100
	g.world.Auction.Bet = 1000

	g.auctionUpdate()
	if g.world.Auction.Going != 2 {
		t.Errorf("Going = %d, want 2", g.world.Auction.Going)
	}
}

// Sold happy path: Buyer != Seller, Bet > 0.
// Item goes to buyer, seller gets 90%, area gets 10% tax, invariant holds.
func TestAuctionUpdate_Sold_HappyPath(t *testing.T) {
	g := newAuctionTestLoop(t)
	area := &types.AreaData{Name: "test"}
	room := &types.RoomIndexData{Vnum: 777, Area: area}
	g.world.Rooms[777] = room

	seller := makeTickChar("Seller")
	seller.InRoom = room
	seller.Gold = 0
	buyer := makeTickChar("Buyer")
	buyer.InRoom = room

	obj, _ := makeTickWeapon("a sword")
	g.world.Auction.Item = obj
	g.world.Auction.Seller = seller
	g.world.Auction.Buyer = buyer
	g.world.Auction.Bet = 10000
	g.world.Auction.Starting = 1000
	g.world.Auction.Going = 2

	g.auctionUpdate()

	if g.world.Auction.Item != nil {
		t.Errorf("Item should clear after sale; got %v", g.world.Auction.Item)
	}
	if obj.CarriedBy != buyer {
		t.Errorf("Item should be transferred to buyer; CarriedBy = %v", obj.CarriedBy)
	}
	// pay = 10000 * 9/10 = 9000, tax = 10000 - 9000 = 1000.
	if seller.Gold != 9000 {
		t.Errorf("Seller.Gold = %d, want 9000", seller.Gold)
	}
	if area.LowEconomy != 1000 {
		t.Errorf("area.LowEconomy = %d, want 1000 (tax)", area.LowEconomy)
	}
}

// Not-sold branch: Buyer == Seller means no real bid. Seller reclaims.
func TestAuctionUpdate_NotSold(t *testing.T) {
	g := newAuctionTestLoop(t)
	area := &types.AreaData{Name: "test"}
	room := &types.RoomIndexData{Vnum: 778, Area: area}
	g.world.Rooms[778] = room

	seller := makeTickChar("Seller")
	seller.InRoom = room
	seller.Gold = 500

	obj, _ := makeTickWeapon("a sword")
	obj.GoldCost = 1000 // 5% of 1000 = 50 tax
	g.world.Auction.Item = obj
	g.world.Auction.Seller = seller
	g.world.Auction.Buyer = seller
	g.world.Auction.Bet = 0
	g.world.Auction.Starting = 0
	g.world.Auction.Going = 2

	g.auctionUpdate()

	if g.world.Auction.Item != nil {
		t.Errorf("Item should clear")
	}
	if obj.CarriedBy != seller {
		t.Errorf("Item should return to seller; CarriedBy = %v", obj.CarriedBy)
	}
	if seller.Gold != 450 {
		t.Errorf("Seller.Gold = %d, want 450 (500-50)", seller.Gold)
	}
}

// Unsold with seller broke: tax floors at 0.
func TestAuctionUpdate_NotSold_TaxFloor(t *testing.T) {
	g := newAuctionTestLoop(t)
	area := &types.AreaData{Name: "test"}
	room := &types.RoomIndexData{Vnum: 779, Area: area}
	g.world.Rooms[779] = room

	seller := makeTickChar("Seller")
	seller.InRoom = room
	seller.Gold = 10 // cannot pay 50 tax

	obj, _ := makeTickWeapon("a sword")
	obj.GoldCost = 1000
	g.world.Auction.Item = obj
	g.world.Auction.Seller = seller
	g.world.Auction.Buyer = seller
	g.world.Auction.Going = 2

	g.auctionUpdate()

	if seller.Gold != 0 {
		t.Errorf("Seller.Gold = %d, want 0 (floored)", seller.Gold)
	}
}

// Bet math invariant: pay + tax = bet even for odd values.
func TestAuctionUpdate_Sold_OddBet_NoGoldLeak(t *testing.T) {
	g := newAuctionTestLoop(t)
	area := &types.AreaData{Name: "test"}
	room := &types.RoomIndexData{Vnum: 780, Area: area}
	g.world.Rooms[780] = room

	seller := makeTickChar("Seller")
	seller.InRoom = room
	buyer := makeTickChar("Buyer")
	buyer.InRoom = room

	obj, _ := makeTickWeapon("a sword")
	g.world.Auction.Item = obj
	g.world.Auction.Seller = seller
	g.world.Auction.Buyer = buyer
	g.world.Auction.Bet = 11
	g.world.Auction.Starting = 1
	g.world.Auction.Going = 2

	g.auctionUpdate()

	total := seller.Gold + area.LowEconomy + area.HighEconomy*1000000000
	if total != 11 {
		t.Errorf("pay+tax = %d, want 11 (no gold leak invariant)", total)
	}
}

// Defensive: Buyer nil, Bet > 0 — C logs and zeroes bet, then not-sold.
func TestAuctionUpdate_Sold_NilBuyer_Recovers(t *testing.T) {
	g := newAuctionTestLoop(t)
	area := &types.AreaData{Name: "test"}
	room := &types.RoomIndexData{Vnum: 781, Area: area}
	g.world.Rooms[781] = room

	seller := makeTickChar("Seller")
	seller.InRoom = room

	obj, _ := makeTickWeapon("a sword")
	g.world.Auction.Item = obj
	g.world.Auction.Seller = seller
	g.world.Auction.Buyer = nil
	g.world.Auction.Bet = 10000
	g.world.Auction.Going = 2

	// Should not panic.
	g.auctionUpdate()

	if g.world.Auction.Item != nil {
		t.Errorf("Item should clear after defensive resolve")
	}
}

// --- closeDescriptor disconnect tests ---

func TestCloseDescriptor_SellerDropout_CancelsAuction(t *testing.T) {
	g := newAuctionTestLoop(t)
	room := &types.RoomIndexData{Vnum: 900}
	g.world.Rooms[900] = room

	seller := makeTickChar("Seller")
	seller.InRoom = room
	buyer := makeTickChar("Buyer")
	buyer.Gold = 0

	obj, _ := makeTickWeapon("a sword")
	g.world.Auction.Item = obj
	g.world.Auction.Seller = seller
	g.world.Auction.Buyer = buyer
	g.world.Auction.Bet = 5000

	g.clearAuctionOnDisconnect(seller)

	if g.world.Auction.Item != nil {
		t.Errorf("Auction.Item should be nil after seller drop; got %v", g.world.Auction.Item)
	}
	if buyer.Gold != 5000 {
		t.Errorf("buyer should be refunded 5000; has %d", buyer.Gold)
	}
	// Item deposited in seller's last room.
	found := false
	for _, o := range room.Contents {
		if o == obj {
			found = true
		}
	}
	if !found {
		t.Errorf("obj should be in seller's last room after drop")
	}
}

func TestCloseDescriptor_BuyerDropout_RevertsBet(t *testing.T) {
	g := newAuctionTestLoop(t)
	room := &types.RoomIndexData{Vnum: 901}
	g.world.Rooms[901] = room

	seller := makeTickChar("Seller")
	seller.InRoom = room
	buyer := makeTickChar("Buyer")
	buyer.InRoom = room

	obj, _ := makeTickWeapon("a sword")
	g.world.Auction.Item = obj
	g.world.Auction.Seller = seller
	g.world.Auction.Buyer = buyer
	g.world.Auction.Bet = 20000
	g.world.Auction.Starting = 1000
	g.world.Auction.Going = 2

	g.clearAuctionOnDisconnect(buyer)

	if g.world.Auction.Buyer != seller {
		t.Errorf("Buyer should revert to seller; got %v", g.world.Auction.Buyer)
	}
	if g.world.Auction.Bet != 1000 {
		t.Errorf("Bet = %d, want 1000 (starting)", g.world.Auction.Bet)
	}
	if g.world.Auction.Going != 0 {
		t.Errorf("Going should reset to 0; got %d", g.world.Auction.Going)
	}
	if g.world.Auction.Pulse != types.PULSE_AUCTION {
		t.Errorf("Pulse should reset to PULSE_AUCTION (%d); got %d", types.PULSE_AUCTION, g.world.Auction.Pulse)
	}
}

func TestCloseDescriptor_PreservesHistory(t *testing.T) {
	g := newAuctionTestLoop(t)
	room := &types.RoomIndexData{Vnum: 902}
	g.world.Rooms[902] = room

	seller := makeTickChar("Seller")
	seller.InRoom = room

	obj, _ := makeTickWeapon("a sword")
	g.world.Auction.Item = obj
	g.world.Auction.Seller = seller
	g.world.Auction.Buyer = seller

	historyMarker := &types.ObjIndexData{Vnum: 42}
	g.world.Auction.History[0] = historyMarker
	g.world.Auction.HistTimer = 3

	g.clearAuctionOnDisconnect(seller)

	if g.world.Auction.History[0] != historyMarker {
		t.Errorf("History should be preserved; got %v", g.world.Auction.History[0])
	}
}

// Quick smoke: disconnect of a non-participant doesn't clear the auction.
func TestCloseDescriptor_NoAuction_Noop(t *testing.T) {
	g := newAuctionTestLoop(t)
	room := &types.RoomIndexData{Vnum: 903}
	g.world.Rooms[903] = room

	seller := makeTickChar("Seller")
	seller.InRoom = room
	other := makeTickChar("Other")
	other.InRoom = room

	obj, _ := makeTickWeapon("a sword")
	g.world.Auction.Item = obj
	g.world.Auction.Seller = seller
	g.world.Auction.Buyer = seller

	g.clearAuctionOnDisconnect(other)
	if g.world.Auction.Item != obj {
		t.Errorf("Item should be preserved when non-participant drops")
	}
}

// Simple sanity: GameLoop.NewGameLoop initializes pulseAuction.
func TestGameLoop_PulseAuctionInitialized(t *testing.T) {
	g := newTestLoop()
	if g.pulseAuction != types.PULSE_AUCTION {
		t.Errorf("pulseAuction = %d, want PULSE_AUCTION (%d)", g.pulseAuction, types.PULSE_AUCTION)
	}
}
