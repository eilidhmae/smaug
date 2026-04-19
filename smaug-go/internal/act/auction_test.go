package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

// --- BroadcastAuction (plan-channels.md G4) ---

// Happy path: trusted, non-deaf, non-silenced room gets "Auction: …".
func TestBroadcastAuction_HappyPath(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 7200, Name: "Room"}
	w.Rooms[7200] = room

	ch, client := makeMortalInRoom(room, "Buyer")
	defer client.Close()
	ch.Trust = auctionMinTrust // trusted mortal

	BroadcastAuction("Sword of Smiting, 100 gold!")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Auction: Sword of Smiting, 100 gold!") {
		t.Errorf("trusted non-deaf receiver should see auction; got %q", out)
	}
}

// Trust < 5: not delivered.
func TestBroadcastAuction_TrustGate(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 7201, Name: "Room"}
	w.Rooms[7201] = room

	ch, client := makeMortalInRoom(room, "Newbie")
	defer client.Close()
	ch.Level = 1
	ch.Trust = auctionMinTrust - 1 // below threshold

	BroadcastAuction("Hi")
	out := readOutput(ch, client)
	if strings.Contains(out, "Hi") {
		t.Errorf("low-trust receiver must NOT see auction; got %q", out)
	}
}

// Deaf[AUCTION]: not delivered.
func TestBroadcastAuction_DeafFilter(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 7202, Name: "Room"}
	w.Rooms[7202] = room

	ch, client := makeMortalInRoom(room, "QuietFan")
	defer client.Close()
	ch.Trust = auctionMinTrust + 5
	ch.Deaf.Set(types.CHANNEL_AUCTION)

	BroadcastAuction("Loud sale")
	out := readOutput(ch, client)
	if strings.Contains(out, "Loud sale") {
		t.Errorf("deaf receiver must NOT see auction; got %q", out)
	}
}

// ROOM_SILENCE: receiver's room silenced — not delivered.
func TestBroadcastAuction_RoomSilence(t *testing.T) {
	w := setupCommWorld()
	silent := &types.RoomIndexData{Vnum: 7203, Name: "Library"}
	silent.RoomFlags.Set(types.ROOM_SILENCE)
	w.Rooms[7203] = silent

	ch, client := makeMortalInRoom(silent, "Scholar")
	defer client.Close()
	ch.Trust = auctionMinTrust + 5

	BroadcastAuction("Noise")
	out := readOutput(ch, client)
	if strings.Contains(out, "Noise") {
		t.Errorf("receiver in ROOM_SILENCE must NOT see auction; got %q", out)
	}
}

// Non-CON_PLAYING descriptors skipped (nanny stage, etc.).
func TestBroadcastAuction_OnlyPlaying(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 7204, Name: "Room"}
	w.Rooms[7204] = room

	ch, client := makeMortalInRoom(room, "Pending")
	defer client.Close()
	ch.Trust = auctionMinTrust + 5
	// Simulate mid-nanny state (still has desc in world, but not PLAYING).
	ch.Desc.Connected = types.CON_GET_NAME

	BroadcastAuction("hi")
	out := readOutput(ch, client)
	if strings.Contains(out, "hi") {
		t.Errorf("non-CON_PLAYING must NOT receive auction; got %q", out)
	}
}

// Nil-safety: WorldRef nil must not panic.
func TestBroadcastAuction_NilWorldNoPanic(t *testing.T) {
	saved := WorldRef
	defer func() { WorldRef = saved }()
	WorldRef = nil

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("BroadcastAuction panicked on nil WorldRef: %v", r)
		}
	}()
	BroadcastAuction("nothing")
}

// --- DoAuction state machine ---

// NPC caller: silent return — matches C `do_auction` at
// src/act_obj.c:3791-3792 (early return on IS_NPC with no message).
func TestDoAuction_NPCGuard(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 7211, Name: "Room"}
	w.Rooms[7211] = room

	mob, mClient := makeTestChar("Mob")
	defer mClient.Close()
	mob.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(mob, room)
	w.AddChar(mob)
	addPlayingDescriptor(mob)

	DoAuction(mob, "list")
	out := readOutput(mob, mClient)
	if out != "" {
		t.Errorf("NPC caller should return silently; got %q", out)
	}
}
