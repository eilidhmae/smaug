package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// setupAuctionWorld configures a world, an auction-open daytime, and
// returns a room handle callers can populate with characters + items.
func setupAuctionWorld(t *testing.T) (*types.RoomIndexData, func()) {
	t.Helper()
	w := setupCommWorld()
	w.TimeInfo = types.TimeInfoData{Hour: 12}
	room := &types.RoomIndexData{Vnum: 8000, Name: "Auction Hall"}
	w.Rooms[8000] = room
	return room, func() {
		// Reset auction between tests to avoid cross-test leaks.
		w.Auction = &types.AuctionData{}
		w.NoAuction = nil
	}
}

// makeWeapon returns a weapon instance (ITEM_WEAPON, value[1]/[2] for dmg
// min/max) with vnum-100 prototype and a stable ShortDescr.
func makeWeapon(vnum int, shortDescr string) (*types.ObjData, *types.ObjIndexData) {
	idx := &types.ObjIndexData{
		Vnum:       vnum,
		Name:       "sword",
		ShortDescr: shortDescr,
		ItemType:   types.ITEM_WEAPON,
		Level:      1,
	}
	obj := &types.ObjData{
		Name:       "sword",
		ShortDescr: shortDescr,
		ItemType:   types.ITEM_WEAPON,
		Level:      1,
		Value:      [6]int{0, 5, 10, 0, 0, 0},
		IndexData:  idx,
		WearLoc:    types.WEAR_NONE,
	}
	return obj, idx
}

func giveObj(obj *types.ObjData, ch *types.CharData) {
	obj.CarriedBy = ch
	obj.WearLoc = types.WEAR_NONE
	ch.Carrying = append(ch.Carrying, obj)
}

// --- Time / level / NPC gates ---

func TestDoAuction_LevelTooLow(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()

	ch, client := makeMortalInRoom(room, "Newbie")
	defer client.Close()
	ch.Level = 2

	DoAuction(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "at least level three") {
		t.Errorf("expected level-3 refusal; got %q", out)
	}
}

func TestDoAuction_NightHours_MortalBlocked(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	WorldRef.TimeInfo.Hour = 20

	ch, client := makeMortalInRoom(room, "NightOwl")
	defer client.Close()
	ch.Level = 10

	DoAuction(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "auctioneer works between") {
		t.Errorf("expected night-hours refusal; got %q", out)
	}
}

func TestDoAuction_NightHours_ImmortalBypass(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	WorldRef.TimeInfo.Hour = 20

	ch, client := makeImmortalInRoom(room, "Zeus")
	defer client.Close()

	DoAuction(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "nothing being auctioned") {
		t.Errorf("immortal should bypass time gate and see info; got %q", out)
	}
}

func TestDoAuction_NightHours_ItemExists_Allowed(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	WorldRef.TimeInfo.Hour = 20

	ch, client := makeMortalInRoom(room, "NightOwl")
	defer client.Close()
	ch.Level = 10

	// Start an auction directly in world state.
	obj, _ := makeWeapon(9001, "a shiny sword")
	WorldRef.Auction.Item = obj
	WorldRef.Auction.Seller = ch
	WorldRef.Auction.Buyer = ch

	DoAuction(ch, "")
	out := readOutput(ch, client)
	if strings.Contains(out, "auctioneer works between") {
		t.Errorf("item-exists bypass should fire; got %q", out)
	}
	// The info block's `Object '%s'` uses obj.Name; weapons emit a damage
	// line but not the ShortDescr. Any info output at all (e.g., "Object"
	// or "Damage is") confirms the time gate was bypassed.
	if !strings.Contains(out, "Damage is") && !strings.Contains(out, "Object '") {
		t.Errorf("info should appear under item-exists bypass; got %q", out)
	}
}

// --- Info branch ---

func TestDoAuction_Empty_NoItem(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	ch, client := makeMortalInRoom(room, "Passerby")
	defer client.Close()
	ch.Level = 10

	DoAuction(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "nothing being auctioned right now") {
		t.Errorf("expected empty-auction info; got %q", out)
	}
}

func TestDoAuction_Empty_WithItem_NoBids(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	ch, client := makeMortalInRoom(room, "Passerby")
	defer client.Close()
	ch.Level = 10
	obj, _ := makeWeapon(9002, "a rusty sword")
	WorldRef.Auction.Item = obj
	WorldRef.Auction.Seller = ch
	WorldRef.Auction.Buyer = ch
	WorldRef.Auction.Bet = 0

	DoAuction(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "No bids on this item have been received") {
		t.Errorf("expected no-bids info; got %q", out)
	}
}

func TestDoAuction_Empty_WithItem_WithBids(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	seller, sc := makeMortalInRoom(room, "Seller")
	defer sc.Close()
	seller.Level = 10
	buyer, bc := makeMortalInRoom(room, "Buyer")
	defer bc.Close()
	buyer.Level = 10

	obj, _ := makeWeapon(9003, "a glowing sword")
	WorldRef.Auction.Item = obj
	WorldRef.Auction.Seller = seller
	WorldRef.Auction.Buyer = buyer
	WorldRef.Auction.Bet = 500

	DoAuction(buyer, "")
	out := readOutput(buyer, bc)
	if !strings.Contains(out, "Current bid on this item is 500 gold") {
		t.Errorf("expected current-bid line; got %q", out)
	}
}

func TestDoAuction_Empty_Immortal_SeesSellerBuyer(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	seller, sc := makeMortalInRoom(room, "Sally")
	defer sc.Close()
	seller.Level = 10
	buyer, bc := makeMortalInRoom(room, "Bobby")
	defer bc.Close()
	buyer.Level = 10
	imm, ic := makeImmortalInRoom(room, "Zeus")
	defer ic.Close()

	obj, _ := makeWeapon(9004, "a sword")
	WorldRef.Auction.Item = obj
	WorldRef.Auction.Seller = seller
	WorldRef.Auction.Buyer = buyer
	WorldRef.Auction.Bet = 100
	WorldRef.Auction.Going = 1

	DoAuction(imm, "")
	out := readOutput(imm, ic)
	if !strings.Contains(out, "Seller: Sally") || !strings.Contains(out, "Bidder: Bobby") {
		t.Errorf("immortal should see seller/bidder names; got %q", out)
	}
	if !strings.Contains(out, "Round:") {
		t.Errorf("immortal should see Round:; got %q", out)
	}
}

func TestDoAuction_Empty_Mortal_NoSellerBuyer(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	seller, sc := makeMortalInRoom(room, "Sally")
	defer sc.Close()
	seller.Level = 10
	buyer, bc := makeMortalInRoom(room, "Bobby")
	defer bc.Close()
	buyer.Level = 10

	obj, _ := makeWeapon(9005, "a sword")
	WorldRef.Auction.Item = obj
	WorldRef.Auction.Seller = seller
	WorldRef.Auction.Buyer = buyer
	WorldRef.Auction.Bet = 100

	DoAuction(buyer, "")
	out := readOutput(buyer, bc)
	if strings.Contains(out, "Seller:") {
		t.Errorf("mortal must NOT see Seller: line; got %q", out)
	}
}

// --- Bid branch ---

func TestDoAuction_Bid_NoItem(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	ch, client := makeMortalInRoom(room, "Bidder")
	defer client.Close()
	ch.Level = 10

	DoAuction(ch, "bid 500")
	out := readOutput(ch, client)
	if !strings.Contains(out, "isn't anything being auctioned") {
		t.Errorf("expected no-item bid refusal; got %q", out)
	}
}

func TestDoAuction_Bid_ItemLevelTooHigh(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	ch, client := makeMortalInRoom(room, "LowLevel")
	defer client.Close()
	ch.Level = 5
	ch.Gold = 1000000

	seller, sc := makeMortalInRoom(room, "Seller")
	defer sc.Close()

	obj, _ := makeWeapon(9006, "a legendary sword")
	obj.Level = 50
	WorldRef.Auction.Item = obj
	WorldRef.Auction.Seller = seller
	WorldRef.Auction.Buyer = seller

	DoAuction(ch, "bid 20000")
	out := readOutput(ch, client)
	if !strings.Contains(out, "level is too high") {
		t.Errorf("expected item-level-too-high; got %q", out)
	}
}

func TestDoAuction_Bid_SelfBid(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	seller, sc := makeMortalInRoom(room, "Seller")
	defer sc.Close()
	seller.Level = 10
	seller.Gold = 1000000

	obj, _ := makeWeapon(9007, "a sword")
	WorldRef.Auction.Item = obj
	WorldRef.Auction.Seller = seller
	WorldRef.Auction.Buyer = seller

	DoAuction(seller, "bid 20000")
	out := readOutput(seller, sc)
	if !strings.Contains(out, "can't bid on your own item") {
		t.Errorf("expected self-bid refusal; got %q", out)
	}
}

func TestDoAuction_Bid_NoAmount(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	seller, sc := makeMortalInRoom(room, "Seller")
	defer sc.Close()
	ch, client := makeMortalInRoom(room, "Bidder")
	defer client.Close()
	ch.Level = 10

	obj, _ := makeWeapon(9008, "a sword")
	WorldRef.Auction.Item = obj
	WorldRef.Auction.Seller = seller
	WorldRef.Auction.Buyer = seller

	DoAuction(ch, "bid")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Bid how much") {
		t.Errorf("expected 'Bid how much?'; got %q", out)
	}
}

func TestDoAuction_Bid_BelowStarting(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	seller, sc := makeMortalInRoom(room, "Seller")
	defer sc.Close()
	ch, client := makeMortalInRoom(room, "Bidder")
	defer client.Close()
	ch.Level = 10
	ch.Gold = 1000000

	obj, _ := makeWeapon(9009, "a sword")
	WorldRef.Auction.Item = obj
	WorldRef.Auction.Seller = seller
	WorldRef.Auction.Buyer = seller
	WorldRef.Auction.Starting = 1000

	DoAuction(ch, "bid 500")
	out := readOutput(ch, client)
	if !strings.Contains(out, "higher than the starting bet") {
		t.Errorf("expected below-starting refusal; got %q", out)
	}
}

func TestDoAuction_Bid_BelowCurrentPlus10k(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	seller, sc := makeMortalInRoom(room, "Seller")
	defer sc.Close()
	ch, client := makeMortalInRoom(room, "Bidder")
	defer client.Close()
	ch.Level = 10
	ch.Gold = 1000000

	obj, _ := makeWeapon(9010, "a sword")
	WorldRef.Auction.Item = obj
	WorldRef.Auction.Seller = seller
	WorldRef.Auction.Buyer = seller
	WorldRef.Auction.Bet = 1000
	WorldRef.Auction.Starting = 1000

	DoAuction(ch, "bid 5000")
	out := readOutput(ch, client)
	if !strings.Contains(out, "at least bid 10000 coins over") {
		t.Errorf("expected +10k refusal; got %q", out)
	}
}

func TestDoAuction_Bid_ExceedsGold(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	seller, sc := makeMortalInRoom(room, "Seller")
	defer sc.Close()
	ch, client := makeMortalInRoom(room, "Poor")
	defer client.Close()
	ch.Level = 10
	ch.Gold = 100

	obj, _ := makeWeapon(9011, "a sword")
	WorldRef.Auction.Item = obj
	WorldRef.Auction.Seller = seller
	WorldRef.Auction.Buyer = seller
	WorldRef.Auction.Starting = 0
	WorldRef.Auction.Bet = 0

	DoAuction(ch, "bid 20000")
	out := readOutput(ch, client)
	if !strings.Contains(out, "don't have that much money") {
		t.Errorf("expected not-enough-gold; got %q", out)
	}
}

func TestDoAuction_Bid_ExceedsMax(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	seller, sc := makeMortalInRoom(room, "Seller")
	defer sc.Close()
	ch, client := makeMortalInRoom(room, "Plutocrat")
	defer client.Close()
	ch.Level = 10
	ch.Gold = 2147483647

	obj, _ := makeWeapon(9012, "a sword")
	WorldRef.Auction.Item = obj
	WorldRef.Auction.Seller = seller
	WorldRef.Auction.Buyer = seller
	WorldRef.Auction.Bet = 0
	WorldRef.Auction.Starting = 0

	DoAuction(ch, "bid 2100000000")
	out := readOutput(ch, client)
	if !strings.Contains(out, "over 2 billion") {
		t.Errorf("expected 2B cap; got %q", out)
	}
}

func TestDoAuction_Bid_Success(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	seller, sc := makeMortalInRoom(room, "Seller")
	defer sc.Close()
	ch, client := makeMortalInRoom(room, "Bidder")
	defer client.Close()
	ch.Level = 10
	ch.Gold = 30000

	obj, _ := makeWeapon(9013, "a sword")
	WorldRef.Auction.Item = obj
	WorldRef.Auction.Seller = seller
	WorldRef.Auction.Buyer = seller
	WorldRef.Auction.Bet = 0
	WorldRef.Auction.Starting = 0

	DoAuction(ch, "bid 20000")
	_ = readOutput(ch, client)

	if WorldRef.Auction.Buyer != ch {
		t.Errorf("Buyer should be ch; got %v", WorldRef.Auction.Buyer)
	}
	if WorldRef.Auction.Bet != 20000 {
		t.Errorf("Bet = %d, want 20000", WorldRef.Auction.Bet)
	}
	if WorldRef.Auction.Going != 0 {
		t.Errorf("Going should reset to 0; got %d", WorldRef.Auction.Going)
	}
	if WorldRef.Auction.Pulse != types.PULSE_AUCTION {
		t.Errorf("Pulse = %d, want PULSE_AUCTION (%d)", WorldRef.Auction.Pulse, types.PULSE_AUCTION)
	}
	if ch.Gold != 10000 {
		t.Errorf("Bidder gold = %d, want 10000", ch.Gold)
	}
}

func TestDoAuction_Bid_RefundsPreviousBuyer(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	seller, sc := makeMortalInRoom(room, "Seller")
	defer sc.Close()
	prev, pc := makeMortalInRoom(room, "Prev")
	defer pc.Close()
	prev.Level = 10
	prev.Gold = 0
	next, nc := makeMortalInRoom(room, "Next")
	defer nc.Close()
	next.Level = 10
	next.Gold = 30000

	obj, _ := makeWeapon(9014, "a sword")
	WorldRef.Auction.Item = obj
	WorldRef.Auction.Seller = seller
	WorldRef.Auction.Buyer = prev
	WorldRef.Auction.Bet = 5000
	WorldRef.Auction.Starting = 0

	DoAuction(next, "bid 20000")
	_ = readOutput(next, nc)

	if prev.Gold != 5000 {
		t.Errorf("prev bidder should be refunded 5000; has %d", prev.Gold)
	}
	if next.Gold != 10000 {
		t.Errorf("next bidder gold = %d, want 10000", next.Gold)
	}
}

func TestDoAuction_Bid_SelfPreviousBidder_NoDoubleRefund(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	seller, sc := makeMortalInRoom(room, "Seller")
	defer sc.Close()
	seller.Gold = 5000
	ch, client := makeMortalInRoom(room, "Bidder")
	defer client.Close()
	ch.Level = 10
	ch.Gold = 30000

	obj, _ := makeWeapon(9015, "a sword")
	WorldRef.Auction.Item = obj
	WorldRef.Auction.Seller = seller
	WorldRef.Auction.Buyer = seller
	WorldRef.Auction.Bet = 0

	DoAuction(ch, "bid 20000")
	_ = readOutput(ch, client)

	if seller.Gold != 5000 {
		t.Errorf("seller should NOT be refunded; has %d, want 5000", seller.Gold)
	}
}

func TestDoAuction_Bid_WrongKeyword(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	seller, sc := makeMortalInRoom(room, "Seller")
	defer sc.Close()
	ch, client := makeMortalInRoom(room, "Bidder")
	defer client.Close()
	ch.Level = 10
	ch.Gold = 30000

	obj, _ := makeWeapon(9016, "a sword")
	WorldRef.Auction.Item = obj
	WorldRef.Auction.Seller = seller
	WorldRef.Auction.Buyer = seller
	WorldRef.Auction.Bet = 0

	DoAuction(ch, "bid 20000 shield")
	out := readOutput(ch, client)
	if !strings.Contains(out, "not being auctioned right now") {
		t.Errorf("expected keyword-mismatch refusal; got %q", out)
	}
}

// --- Start-new branch ---

func TestDoAuction_Start_NotCarried(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	ch, client := makeMortalInRoom(room, "Empty")
	defer client.Close()
	ch.Level = 10

	DoAuction(ch, "sword 1000")
	out := readOutput(ch, client)
	if !strings.Contains(out, "aren't carrying that") {
		t.Errorf("expected not-carrying; got %q", out)
	}
}

func TestDoAuction_Start_InNoAuctionList(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	ch, client := makeMortalInRoom(room, "Seller")
	defer client.Close()
	ch.Level = 10

	obj, idx := makeWeapon(100, "a blacklisted sword")
	_ = idx
	giveObj(obj, ch)
	WorldRef.NoAuction = []int{100}

	DoAuction(ch, "sword 1000")
	out := readOutput(ch, client)
	if !strings.Contains(out, "cannot be auctioned") {
		t.Errorf("expected noauction refusal; got %q", out)
	}
}

func TestDoAuction_Start_ImmortalBypassesNoauc(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	ch, client := makeImmortalInRoom(room, "Zeus")
	defer client.Close()

	obj, _ := makeWeapon(100, "a sword")
	giveObj(obj, ch)
	WorldRef.NoAuction = []int{100}

	DoAuction(ch, "sword 1000")
	_ = readOutput(ch, client)

	if WorldRef.Auction.Item != obj {
		t.Errorf("immortal should bypass noauction; item = %v", WorldRef.Auction.Item)
	}
}

func TestDoAuction_Start_TypeMismatch(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	ch, client := makeMortalInRoom(room, "Modder")
	defer client.Close()
	ch.Level = 10

	obj, idx := makeWeapon(200, "a sword")
	idx.ItemType = types.ITEM_TREASURE // modded
	giveObj(obj, ch)

	DoAuction(ch, "sword 1000")
	out := readOutput(ch, client)
	if !strings.Contains(out, "too modified") {
		t.Errorf("expected too-modified refusal; got %q", out)
	}
}

func TestDoAuction_Start_Decaying(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	ch, client := makeMortalInRoom(room, "Seller")
	defer client.Close()
	ch.Level = 10

	obj, _ := makeWeapon(300, "a decaying sword")
	obj.Timer = 10
	giveObj(obj, ch)

	DoAuction(ch, "sword 1000")
	out := readOutput(ch, client)
	if !strings.Contains(out, "decaying") {
		t.Errorf("expected decaying refusal; got %q", out)
	}
}

func TestDoAuction_Start_ClanObject(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	ch, client := makeMortalInRoom(room, "Seller")
	defer client.Close()
	ch.Level = 10

	obj, _ := makeWeapon(400, "a clan sword")
	obj.ExtraFlags.Set(types.ITEM_CLANOBJECT)
	giveObj(obj, ch)

	DoAuction(ch, "sword 1000")
	out := readOutput(ch, client)
	if !strings.Contains(out, "clan items") {
		t.Errorf("expected clan refusal; got %q", out)
	}
}

func TestDoAuction_Start_Permanent(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	ch, client := makeMortalInRoom(room, "Seller")
	defer client.Close()
	ch.Level = 10

	obj, _ := makeWeapon(500, "a permanent sword")
	obj.ExtraFlags.Set(types.ITEM_PERMANENT)
	giveObj(obj, ch)

	DoAuction(ch, "sword 1000")
	out := readOutput(ch, client)
	if !strings.Contains(out, "cannot leave your possession") {
		t.Errorf("expected permanent refusal; got %q", out)
	}
}

func TestDoAuction_Start_HistoryCollision(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	ch, client := makeMortalInRoom(room, "Seller")
	defer client.Close()
	ch.Level = 10

	obj, idx := makeWeapon(600, "a sword")
	giveObj(obj, ch)
	WorldRef.Auction.History[0] = idx

	DoAuction(ch, "sword 1000")
	out := readOutput(ch, client)
	if !strings.Contains(out, "auctioned recently") {
		t.Errorf("expected history-collision refusal; got %q", out)
	}
}

func TestDoAuction_Start_NegativeMinBet(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	ch, client := makeMortalInRoom(room, "Seller")
	defer client.Close()
	ch.Level = 10

	obj, _ := makeWeapon(700, "a sword")
	giveObj(obj, ch)

	DoAuction(ch, "sword -10")
	out := readOutput(ch, client)
	if !strings.Contains(out, "less than 0 gold") {
		t.Errorf("expected negative-min-bet refusal; got %q", out)
	}
}

func TestDoAuction_Start_InvalidMinBet(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	ch, client := makeMortalInRoom(room, "Seller")
	defer client.Close()
	ch.Level = 10

	obj, _ := makeWeapon(701, "a sword")
	giveObj(obj, ch)

	DoAuction(ch, "sword garbage")
	out := readOutput(ch, client)
	if !strings.Contains(out, "must input a number") {
		t.Errorf("expected must-input-number; got %q", out)
	}
}

func TestDoAuction_Start_DisallowedItemType(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	ch, client := makeMortalInRoom(room, "Seller")
	defer client.Close()
	ch.Level = 10

	idx := &types.ObjIndexData{
		Vnum:       800,
		Name:       "key",
		ShortDescr: "a key",
		ItemType:   types.ITEM_KEY,
	}
	obj := &types.ObjData{
		Name:       "key",
		ShortDescr: "a key",
		ItemType:   types.ITEM_KEY,
		IndexData:  idx,
		WearLoc:    types.WEAR_NONE,
	}
	giveObj(obj, ch)

	DoAuction(ch, "key 1000")
	out := readOutput(ch, client)
	if !strings.Contains(out, "cannot auction") {
		t.Errorf("expected cannot-auction-Xs; got %q", out)
	}
}

func TestDoAuction_Start_HappyPath_Weapon(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	ch, client := makeMortalInRoom(room, "Seller")
	defer client.Close()
	ch.Level = 10

	obj, idx := makeWeapon(900, "a shiny sword")
	giveObj(obj, ch)

	DoAuction(ch, "sword 1000")
	_ = readOutput(ch, client)

	if WorldRef.Auction.Item != obj {
		t.Errorf("Item not set on happy path")
	}
	if WorldRef.Auction.Seller != ch || WorldRef.Auction.Buyer != ch {
		t.Errorf("Seller/Buyer should both be ch; got seller=%v buyer=%v",
			WorldRef.Auction.Seller, WorldRef.Auction.Buyer)
	}
	if WorldRef.Auction.Starting != 1000 {
		t.Errorf("Starting = %d, want 1000", WorldRef.Auction.Starting)
	}
	// starting > 0 ⇒ Bet preloaded to starting (act_obj.c:4222-4223).
	if WorldRef.Auction.Bet != 1000 {
		t.Errorf("Bet preload = %d, want 1000", WorldRef.Auction.Bet)
	}
	if WorldRef.Auction.History[0] != idx {
		t.Errorf("History[0] not set to idx; got %v", WorldRef.Auction.History[0])
	}
	// Item should be removed from ch.Carrying.
	for _, o := range ch.Carrying {
		if o == obj {
			t.Errorf("obj still in ch.Carrying after auction start")
		}
	}
}

func TestDoAuction_Start_HistoryRotation(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	ch, client := makeMortalInRoom(room, "Seller")
	defer client.Close()
	ch.Level = 10

	// Preseed history: [A, B, nil].
	a := &types.ObjIndexData{Vnum: 901}
	b := &types.ObjIndexData{Vnum: 902}
	WorldRef.Auction.History[0] = a
	WorldRef.Auction.History[1] = b

	obj, idx := makeWeapon(903, "a new sword")
	giveObj(obj, ch)

	DoAuction(ch, "sword 0")
	_ = readOutput(ch, client)

	if WorldRef.Auction.History[0] != idx {
		t.Errorf("History[0] not rotated to new entry; got %v", WorldRef.Auction.History[0])
	}
	if WorldRef.Auction.History[1] != a {
		t.Errorf("History[1] = %v, want a", WorldRef.Auction.History[1])
	}
	if WorldRef.Auction.History[2] != b {
		t.Errorf("History[2] = %v, want b", WorldRef.Auction.History[2])
	}
}

func TestDoAuction_Start_WhileInProgress_MortalWaitState(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	ch, client := makeMortalInRoom(room, "Seller")
	defer client.Close()
	ch.Level = 10

	// Existing auction.
	existing, _ := makeWeapon(1000, "auctioneer's current item")
	WorldRef.Auction.Item = existing

	// Seller's second item.
	obj, _ := makeWeapon(1001, "another sword")
	giveObj(obj, ch)

	DoAuction(ch, "sword 1000")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Try again later") {
		t.Errorf("expected try-again; got %q", out)
	}
	if ch.Wait != types.PULSE_VIOLENCE {
		t.Errorf("ch.Wait = %d, want PULSE_VIOLENCE (%d)", ch.Wait, types.PULSE_VIOLENCE)
	}
}

func TestDoAuction_Start_WhileInProgress_ImmortalNoWaitState(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	ch, client := makeImmortalInRoom(room, "Zeus")
	defer client.Close()

	existing, _ := makeWeapon(1002, "current item")
	WorldRef.Auction.Item = existing
	obj, _ := makeWeapon(1003, "another sword")
	giveObj(obj, ch)
	ch.Wait = 0

	DoAuction(ch, "sword 1000")
	_ = readOutput(ch, client)
	if ch.Wait != 0 {
		t.Errorf("immortal Wait should stay 0; got %d", ch.Wait)
	}
}

// --- Stop branch (immortal) ---

func TestDoAuction_Stop_Immortal_NoActiveAuction(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	ch, client := makeImmortalInRoom(room, "Zeus")
	defer client.Close()

	DoAuction(ch, "stop")
	out := readOutput(ch, client)
	if !strings.Contains(out, "no auction to stop") {
		t.Errorf("expected no-auction message; got %q", out)
	}
}

func TestDoAuction_Stop_Immortal_Active_ReturnsItem(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	imm, ic := makeImmortalInRoom(room, "Zeus")
	defer ic.Close()
	seller, sc := makeMortalInRoom(room, "Seller")
	defer sc.Close()
	buyer, bc := makeMortalInRoom(room, "Buyer")
	defer bc.Close()
	buyer.Gold = 0

	obj, _ := makeWeapon(1100, "a sword")
	WorldRef.Auction.Item = obj
	WorldRef.Auction.Seller = seller
	WorldRef.Auction.Buyer = buyer
	WorldRef.Auction.Bet = 5000

	DoAuction(imm, "stop")
	_ = readOutput(imm, ic)
	_ = readOutput(buyer, bc)

	if WorldRef.Auction.Item != nil {
		t.Errorf("Item should be nil after stop; got %v", WorldRef.Auction.Item)
	}
	if obj.CarriedBy != seller {
		t.Errorf("Item should return to seller; CarriedBy = %v", obj.CarriedBy)
	}
	if buyer.Gold != 5000 {
		t.Errorf("Buyer should be refunded 5000; has %d", buyer.Gold)
	}
}

func TestDoAuction_Stop_Immortal_SameBuyerAndSeller_NoRefund(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	imm, ic := makeImmortalInRoom(room, "Zeus")
	defer ic.Close()
	seller, sc := makeMortalInRoom(room, "Seller")
	defer sc.Close()
	seller.Gold = 0

	obj, _ := makeWeapon(1200, "a sword")
	WorldRef.Auction.Item = obj
	WorldRef.Auction.Seller = seller
	WorldRef.Auction.Buyer = seller
	WorldRef.Auction.Bet = 5000

	DoAuction(imm, "stop")
	_ = readOutput(imm, ic)

	if seller.Gold != 0 {
		t.Errorf("seller should NOT be refunded when buyer==seller; gold=%d", seller.Gold)
	}
}

// --- DoQuit gate (G8) ---

// Note: DoQuit blocking is tested with a live auction. Full closeDescriptor
// defensive-clear tests live in internal/game/auction_disconnect_test.go
// (sibling package, direct access to g.closeDescriptor).

func TestDoQuit_BlockedWhileAuctioning_Seller(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	ch, client := makeMortalInRoom(room, "Seller")
	defer client.Close()
	ch.Level = 10
	ch.Position = types.POS_STANDING

	obj, _ := makeWeapon(1300, "a sword")
	WorldRef.Auction.Item = obj
	WorldRef.Auction.Seller = ch

	DoQuit(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Wait until you have bought/sold") {
		t.Errorf("expected auction quit-gate; got %q", out)
	}
	if ch.Desc.Connected == -1 {
		t.Errorf("ch.Desc.Connected should NOT be -1 after blocked quit")
	}
}

func TestDoQuit_BlockedWhileAuctioning_Buyer(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	seller, sc := makeMortalInRoom(room, "Seller")
	defer sc.Close()
	ch, client := makeMortalInRoom(room, "Buyer")
	defer client.Close()
	ch.Level = 10
	ch.Position = types.POS_STANDING

	obj, _ := makeWeapon(1301, "a sword")
	WorldRef.Auction.Item = obj
	WorldRef.Auction.Seller = seller
	WorldRef.Auction.Buyer = ch

	DoQuit(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Wait until you have bought/sold") {
		t.Errorf("expected auction quit-gate; got %q", out)
	}
}

func TestDoQuit_NotBlocked_NoActiveAuction(t *testing.T) {
	room, cleanup := setupAuctionWorld(t)
	defer cleanup()
	ch, client := makeMortalInRoom(room, "Free")
	defer client.Close()
	ch.Level = 10
	ch.Position = types.POS_STANDING
	// No Auction.Item set.

	DoQuit(ch, "")
	out := readOutput(ch, client)
	if strings.Contains(out, "Wait until you have bought/sold") {
		t.Errorf("gate should not fire without active auction; got %q", out)
	}
}
