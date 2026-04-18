// Auction channel helper + stubbed do_auction command.
//
// Per plan-channels.md G4: the full `do_auction` state machine
// (list/bid/stop, escrow, tick) is ~250 LOC of auction-house logic and is
// deferred to Phase 6. What lands here is the narrow-but-reusable
// `BroadcastAuction` helper (C: `talk_auction` at act_comm.c:4309) plus a
// `DoAuction` stub that echoes a "closed" message so the command visible in
// `commands` doesn't look broken.
package act

import (
	"github.com/eilidhmae/smaug/internal/types"
)

// auctionMinTrust is the trust threshold C `talk_auction` applies to
// receivers (`act_comm.c:4322`). Pulled out as a named constant rather than
// a magic literal so the mapping back to C is obvious.
const auctionMinTrust = 5

// BroadcastAuction delivers `message` as an AUCTION: channel line to every
// currently-playing descriptor that passes the C-matching filter chain:
// Trust >= 5, Deaf[AUCTION] not set, current room is NOT ROOM_SILENCE.
// Intended for wiring up the Phase-6 auction subsystem once it lands;
// exported now so tests (and any early immortal-helper tools) can drive it.
func BroadcastAuction(message string) {
	if WorldRef == nil {
		return
	}
	for _, d := range WorldRef.Descriptors {
		if d == nil || d.Connected != types.CON_PLAYING || d.Character == nil {
			continue
		}
		ch := d.Character
		if ch.GetTrust() < auctionMinTrust {
			continue
		}
		if ch.Deaf.IsSet(types.CHANNEL_AUCTION) {
			continue
		}
		if ch.InRoom != nil && ch.InRoom.RoomFlags.IsSet(types.ROOM_SILENCE) {
			continue
		}
		ch.Sendf("Auction: %s\n\r", message)
	}
}

// DoAuction is the Phase-5 stub for the `auction` command. The full auction
// system (list/bid/stop/escrow/tick) is tracked as a Phase-6 deliverable;
// landing a stub here keeps the command from being visibly missing when
// players type `commands`. See plan-channels.md G5.
func DoAuction(ch *types.CharData, argument string) {
	if ch == nil {
		return
	}
	if ch.IsNPC() {
		ch.Send("Huh?\n\r")
		return
	}
	ch.Send("The auction house is currently closed. (See the Phase-6 roadmap.)\n\r")
}
