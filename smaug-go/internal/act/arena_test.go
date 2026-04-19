package act

import (
	"net"
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

// setupArenaWorld creates a world with one temple room (where challengers
// live pre-accept) and two arena-flagged rooms so DoAccept has somewhere
// to teleport. Also resets arenaState for isolation between tests.
func setupArenaWorld() (room, arena1, arena2 *types.RoomIndexData) {
	w := setupCommWorld()
	room = &types.RoomIndexData{Vnum: 6200, Name: "Staging"}
	w.Rooms[6200] = room
	arena1 = &types.RoomIndexData{Vnum: types.ROOM_VNUM_ARENA_MIN, Name: "Arena A"}
	arena1.RoomFlags.Set(types.ROOM_ARENA)
	w.Rooms[arena1.Vnum] = arena1
	arena2 = &types.RoomIndexData{Vnum: types.ROOM_VNUM_ARENA_MIN + 1, Name: "Arena B"}
	arena2.RoomFlags.Set(types.ROOM_ARENA)
	w.Rooms[arena2.Vnum] = arena2
	// Make sure altar + temple exist so DoLook doesn't choke if called.
	temple := &types.RoomIndexData{Vnum: types.ROOM_VNUM_TEMPLE, Name: "Temple"}
	w.Rooms[types.ROOM_VNUM_TEMPLE] = temple
	altar := &types.RoomIndexData{Vnum: types.ROOM_VNUM_ALTAR, Name: "Altar"}
	w.Rooms[types.ROOM_VNUM_ALTAR] = altar
	ResetArenaState()
	return room, arena1, arena2
}

// addPC places `ch` in `room`, registers with the world + descriptor
// table so GetCharWorld finds them. Returns the client conn for
// readOutput. The PC is Level 10 by default (> 5 gate).
func addPC(room *types.RoomIndexData, name string) (*types.CharData, net.Conn) {
	ch, client := makeTestChar(name)
	handler.CharToRoom(ch, room)
	WorldRef.AddChar(ch)
	addPlayingDescriptor(ch)
	return ch, client
}

// drainOutput consumes any pending output so subsequent readOutput
// captures only what a specific call emits. Safe on zero-output state.
func drainOutput(ch *types.CharData, client net.Conn) {
	_ = readOutput(ch, client)
}

// ============================================================
// G2 — DoChallenge
// ============================================================

func TestDoChallenge_EmptyArg(t *testing.T) {
	room, _, _ := setupArenaWorld()
	ch, client := addPC(room, "Alice")
	defer client.Close()
	DoChallenge(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "You must specify who you want to challenge") {
		t.Errorf("expected empty-arg message; got %q", out)
	}
	if ArenaIsChallenge() {
		t.Error("arenaState.IsChallenge should be false after rejected challenge")
	}
}

func TestDoChallenge_AlreadyChallenging(t *testing.T) {
	room, _, _ := setupArenaWorld()
	SetArenaIsChallenge(true)
	ch, client := addPC(room, "Alice")
	defer client.Close()
	DoChallenge(ch, "Bob")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Someone has been challenged wait a few moments") {
		t.Errorf("expected is_challenge message; got %q", out)
	}
}

func TestDoChallenge_ArenaBusy(t *testing.T) {
	room, _, _ := setupArenaWorld()
	SetArenaIsBusy(true)
	ch, client := addPC(room, "Alice")
	defer client.Close()
	DoChallenge(ch, "Bob")
	out := readOutput(ch, client)
	if !strings.Contains(out, "arena is being used") {
		t.Errorf("expected arena_is_busy message; got %q", out)
	}
}

func TestDoChallenge_VictimNotFound(t *testing.T) {
	room, _, _ := setupArenaWorld()
	ch, client := addPC(room, "Alice")
	defer client.Close()
	DoChallenge(ch, "Ghost")
	out := readOutput(ch, client)
	if !strings.Contains(out, "seems to be gone from the realms") {
		t.Errorf("expected not-found message; got %q", out)
	}
	if !strings.Contains(out, "Ghost") {
		t.Errorf("expected argument to appear ($t); got %q", out)
	}
}

func TestDoChallenge_ChIsAlreadyChallenged(t *testing.T) {
	room, _, _ := setupArenaWorld()
	ch, aClient := addPC(room, "Alice")
	defer aClient.Close()
	_, bClient := addPC(room, "Bob")
	defer bClient.Close()
	ch.Act.Set(types.ACT_CHALLENGED)
	DoChallenge(ch, "Bob")
	out := readOutput(ch, aClient)
	if !strings.Contains(out, "You have been challenged already") {
		t.Errorf("expected challenged-already message; got %q", out)
	}
}

func TestDoChallenge_VictimIsChallenger(t *testing.T) {
	room, _, _ := setupArenaWorld()
	ch, aClient := addPC(room, "Alice")
	defer aClient.Close()
	vic, bClient := addPC(room, "Bob")
	defer bClient.Close()
	vic.Act.Set(types.ACT_CHALLENGER)
	DoChallenge(ch, "Bob")
	out := readOutput(ch, aClient)
	if !strings.Contains(out, "has already challenged someone else") {
		t.Errorf("expected already-challenged-someone message; got %q", out)
	}
}

func TestDoChallenge_VictimAFK(t *testing.T) {
	room, _, _ := setupArenaWorld()
	ch, aClient := addPC(room, "Alice")
	defer aClient.Close()
	vic, bClient := addPC(room, "Bob")
	defer bClient.Close()
	vic.Act.Set(types.PLR_AFK)
	DoChallenge(ch, "Bob")
	out := readOutput(ch, aClient)
	if !strings.Contains(out, "is AFK at the moment") {
		t.Errorf("expected AFK message; got %q", out)
	}
}

func TestDoChallenge_VictimIsNPC(t *testing.T) {
	room, _, _ := setupArenaWorld()
	ch, aClient := addPC(room, "Alice")
	defer aClient.Close()
	mob, bClient := addPC(room, "Orc")
	defer bClient.Close()
	// Mark as NPC: set ACT_IS_NPC.
	mob.Act.Set(types.ACT_IS_NPC)
	DoChallenge(ch, "Orc")
	out := readOutput(ch, aClient)
	if !strings.Contains(out, "You can only challenge players or yourself") {
		t.Errorf("expected NPC-only message; got %q", out)
	}
}

func TestDoChallenge_ChIsNPC(t *testing.T) {
	room, _, _ := setupArenaWorld()
	ch, aClient := addPC(room, "Mob")
	defer aClient.Close()
	_, bClient := addPC(room, "Bob")
	defer bClient.Close()
	ch.Act.Set(types.ACT_IS_NPC)
	DoChallenge(ch, "Bob")
	out := readOutput(ch, aClient)
	if !strings.Contains(out, "You can only challenge players or yourself") {
		t.Errorf("expected NPC-only message; got %q", out)
	}
}

func TestDoChallenge_VictimImmortalChMortal(t *testing.T) {
	room, _, _ := setupArenaWorld()
	ch, aClient := addPC(room, "Alice")
	defer aClient.Close()
	vic, bClient := addPC(room, "Zeus")
	defer bClient.Close()
	vic.Level = types.LEVEL_IMMORTAL
	DoChallenge(ch, "Zeus")
	out := readOutput(ch, aClient)
	if !strings.Contains(out, "higher being") {
		t.Errorf("expected higher-being message; got %q", out)
	}
	if !strings.Contains(out, "laugh at you") {
		t.Errorf("expected laugh-at-you follow-up; got %q", out)
	}
}

func TestDoChallenge_VictimLowLevel(t *testing.T) {
	room, _, _ := setupArenaWorld()
	ch, aClient := addPC(room, "Alice")
	defer aClient.Close()
	vic, bClient := addPC(room, "Bob")
	defer bClient.Close()
	vic.Level = 5
	DoChallenge(ch, "Bob")
	out := readOutput(ch, aClient)
	if !strings.Contains(out, "not experienced enough") {
		t.Errorf("expected low-level message; got %q", out)
	}
}

func TestDoChallenge_VictimFighting(t *testing.T) {
	room, _, _ := setupArenaWorld()
	ch, aClient := addPC(room, "Alice")
	defer aClient.Close()
	vic, bClient := addPC(room, "Bob")
	defer bClient.Close()
	other, cClient := addPC(room, "Carl")
	defer cClient.Close()
	vic.Fighting = &types.FightData{Who: other}
	DoChallenge(ch, "Bob")
	out := readOutput(ch, aClient)
	if !strings.Contains(out, "in combat right now") {
		t.Errorf("expected fighting message; got %q", out)
	}
}

func TestDoChallenge_SelfArg(t *testing.T) {
	// "self" with no such PC in-world goes through get_char_world
	// first and hits the not-found branch. No state mutation. The
	// explicit "self" check at C arena.c:177 is functionally dead
	// code for typical play — it only fires if a PC literally named
	// "Self" exists. That case is pinned by
	// TestDoChallenge_SelfArgAfterFound below.
	room, _, _ := setupArenaWorld()
	ch, _ := addPC(room, "Alice")
	DoChallenge(ch, "self")
	if ArenaIsChallenge() {
		t.Error("self arg should not set IsChallenge")
	}
	if ch.Act.IsSet(types.ACT_CHALLENGER) {
		t.Error("self arg should not set ACT_CHALLENGER")
	}
}

func TestDoChallenge_MirrorMatch(t *testing.T) {
	// "mirror match" in C flows through get_char_world first, so the
	// explicit string check at arena.c:177 is typically dead code —
	// get_char_world returns NULL on non-existent name, firing the
	// "gone from the realms" branch instead. We preserve the order
	// verbatim, so a "mirror match" arg with no such player does NOT
	// reach the silent branch. What matters: no state mutation.
	room, _, _ := setupArenaWorld()
	ch, _ := addPC(room, "Alice")
	DoChallenge(ch, "mirror match")
	if ArenaIsChallenge() {
		t.Error("mirror-match should not set IsChallenge")
	}
	if ch.Act.IsSet(types.ACT_CHALLENGER) {
		t.Error("mirror-match should not set ACT_CHALLENGER")
	}
}

// TestDoChallenge_SelfArgAfterFound verifies that the silent "self"
// branch fires when get_char_world DOES resolve "self" to a player —
// a PC named "Self" would satisfy this unlikely corner. Pins that
// reaching the branch short-circuits without setting state.
func TestDoChallenge_SelfArgAfterFound(t *testing.T) {
	room, _, _ := setupArenaWorld()
	ch, aClient := addPC(room, "Alice")
	defer aClient.Close()
	// Ensure get_char_world(ch, "self") succeeds: create a player
	// called "self". The silent branch then short-circuits.
	_, bClient := addPC(room, "Self")
	defer bClient.Close()
	drainOutput(ch, aClient)
	DoChallenge(ch, "self")
	if ArenaIsChallenge() {
		t.Error("self arg should not set IsChallenge")
	}
	if ch.Act.IsSet(types.ACT_CHALLENGER) {
		t.Error("self arg should not set ACT_CHALLENGER")
	}
	out := readOutput(ch, aClient)
	if out != "" {
		t.Errorf("self arg should be silent once get_char_world resolves; got %q", out)
	}
}

func TestDoChallenge_Success(t *testing.T) {
	room, _, _ := setupArenaWorld()
	ch, aClient := addPC(room, "Alice")
	defer aClient.Close()
	vic, bClient := addPC(room, "Bob")
	defer bClient.Close()

	DoChallenge(ch, "Bob")

	if !ch.Act.IsSet(types.ACT_CHALLENGER) {
		t.Error("expected ch.ACT_CHALLENGER after successful challenge")
	}
	if !ArenaIsChallenge() {
		t.Error("expected arenaState.IsChallenge = true")
	}
	if handler.GetTimer(ch, types.TIMER_CHALLENGE) != 5 {
		t.Errorf("expected TIMER_CHALLENGE count=5; got %d",
			handler.GetTimer(ch, types.TIMER_CHALLENGE))
	}
	aOut := readOutput(ch, aClient)
	if !strings.Contains(aOut, "Your challenge has been sent") {
		t.Errorf("challenger should see confirmation; got %q", aOut)
	}
	if !strings.Contains(aOut, "Alice") {
		t.Errorf("challenger's confirmation should mention their name; got %q", aOut)
	}
	bOut := readOutput(vic, bClient)
	if !strings.Contains(bOut, "You have been challenged by Alice") {
		t.Errorf("victim should see challenge notice; got %q", bOut)
	}
	if !strings.Contains(bOut, "Type accept Alice") {
		t.Errorf("victim should see accept instructions; got %q", bOut)
	}
}

// ============================================================
// G3 — DoAccept
// ============================================================

func TestDoAccept_NPC(t *testing.T) {
	room, _, _ := setupArenaWorld()
	ch, client := addPC(room, "Orc")
	defer client.Close()
	ch.Act.Set(types.ACT_IS_NPC)
	DoAccept(ch, "Alice")
	out := readOutput(ch, client)
	if !strings.Contains(out, "new kinda mobile to accept challenges") {
		t.Errorf("expected NPC-accept message; got %q", out)
	}
}

func TestDoAccept_EmptyArg(t *testing.T) {
	room, _, _ := setupArenaWorld()
	ch, client := addPC(room, "Bob")
	defer client.Close()
	DoAccept(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "You must be specify who challenged you") {
		t.Errorf("expected empty-arg message; got %q", out)
	}
}

func TestDoAccept_VictimNotFound(t *testing.T) {
	room, _, _ := setupArenaWorld()
	ch, client := addPC(room, "Bob")
	defer client.Close()
	DoAccept(ch, "Ghost")
	out := readOutput(ch, client)
	if !strings.Contains(out, "seems to be gone from the realms") {
		t.Errorf("expected not-found; got %q", out)
	}
}

func TestDoAccept_VictimIsNPC(t *testing.T) {
	room, _, _ := setupArenaWorld()
	ch, aClient := addPC(room, "Bob")
	defer aClient.Close()
	vic, bClient := addPC(room, "Orc")
	defer bClient.Close()
	vic.Act.Set(types.ACT_IS_NPC)
	DoAccept(ch, "Orc")
	out := readOutput(ch, aClient)
	if !strings.Contains(out, "I dont think I can let you die like that") {
		t.Errorf("expected NPC-victim message; got %q", out)
	}
}

func TestDoAccept_SelfAccept(t *testing.T) {
	room, _, _ := setupArenaWorld()
	ch, client := addPC(room, "Bob")
	defer client.Close()
	DoAccept(ch, "Bob")
	out := readOutput(ch, client)
	if !strings.Contains(out, "I bet you think your funny eh") {
		t.Errorf("expected self-accept message; got %q", out)
	}
}

func TestDoAccept_VictimNotChallenger(t *testing.T) {
	room, _, _ := setupArenaWorld()
	ch, aClient := addPC(room, "Bob")
	defer aClient.Close()
	_, bClient := addPC(room, "Alice")
	defer bClient.Close()
	DoAccept(ch, "Alice")
	out := readOutput(ch, aClient)
	if !strings.Contains(out, "drank that much ale") {
		t.Errorf("expected not-challenger message; got %q", out)
	}
}

func TestDoAccept_Success(t *testing.T) {
	room, arena1, arena2 := setupArenaWorld()
	// Pre-state: Alice challenged Bob.
	challenger, cClient := addPC(room, "Alice")
	defer cClient.Close()
	accepter, aClient := addPC(room, "Bob")
	defer aClient.Close()
	challenger.Act.Set(types.ACT_CHALLENGER)
	handler.AddTimer(challenger, types.TIMER_CHALLENGE, 5, "", 0)
	SetArenaIsChallenge(true)
	drainOutput(challenger, cClient)
	drainOutput(accepter, aClient)

	DoAccept(accepter, "Alice")

	// ch is the accepter (Bob); ch gets ACT_CHALLENGED.
	if !accepter.Act.IsSet(types.ACT_CHALLENGED) {
		t.Error("accepter should have ACT_CHALLENGED set")
	}
	if !accepter.Act.IsSet(types.PLR_SILENCE) {
		t.Error("accepter should have PLR_SILENCE set")
	}
	if !challenger.Act.IsSet(types.PLR_SILENCE) {
		t.Error("challenger should have PLR_SILENCE set")
	}
	if !ArenaIsBusy() {
		t.Error("arenaState.IsBusy should be true")
	}
	if ArenaIsChallenge() {
		t.Error("arenaState.IsChallenge should be cleared")
	}
	if handler.HasTimer(challenger, types.TIMER_CHALLENGE) {
		t.Error("TIMER_CHALLENGE should be removed from challenger")
	}
	// Both PCs should now be in some arena room (room1/room2 are random
	// between arena1/arena2; we only care they landed in ROOM_ARENA).
	if accepter.InRoom == nil || !accepter.InRoom.RoomFlags.IsSet(types.ROOM_ARENA) {
		t.Errorf("accepter not in arena room: %+v", accepter.InRoom)
	}
	if challenger.InRoom == nil || !challenger.InRoom.RoomFlags.IsSet(types.ROOM_ARENA) {
		t.Errorf("challenger not in arena room: %+v", challenger.InRoom)
	}
	_ = arena1
	_ = arena2
}

// ============================================================
// G4 — DoDecline + DoWithdraw
// ============================================================

func TestDoDecline_EmptyArg(t *testing.T) {
	room, _, _ := setupArenaWorld()
	SetArenaIsChallenge(true)
	ch, client := addPC(room, "Bob")
	defer client.Close()
	DoDecline(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "You must specify who you want to challenge") {
		t.Errorf("expected challenge-syntax error; got %q", out)
	}
}

func TestDoDecline_NoChallenge(t *testing.T) {
	room, _, _ := setupArenaWorld()
	ch, client := addPC(room, "Bob")
	defer client.Close()
	DoDecline(ch, "Alice")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Nobody hase even challenged") {
		t.Errorf("expected no-challenge message; got %q", out)
	}
}

func TestDoDecline_ArenaBusy(t *testing.T) {
	room, _, _ := setupArenaWorld()
	SetArenaIsChallenge(true)
	SetArenaIsBusy(true)
	ch, client := addPC(room, "Bob")
	defer client.Close()
	DoDecline(ch, "Alice")
	out := readOutput(ch, client)
	if !strings.Contains(out, "arena is in use") {
		t.Errorf("expected arena-busy message; got %q", out)
	}
}

func TestDoDecline_VictimNotFound(t *testing.T) {
	room, _, _ := setupArenaWorld()
	SetArenaIsChallenge(true)
	ch, client := addPC(room, "Bob")
	defer client.Close()
	DoDecline(ch, "Ghost")
	out := readOutput(ch, client)
	if !strings.Contains(out, "seems to be gone from the realms") {
		t.Errorf("expected not-found; got %q", out)
	}
}

func TestDoDecline_VictimNotChallenger(t *testing.T) {
	room, _, _ := setupArenaWorld()
	SetArenaIsChallenge(true)
	ch, aClient := addPC(room, "Bob")
	defer aClient.Close()
	_, bClient := addPC(room, "Alice")
	defer bClient.Close()
	DoDecline(ch, "Alice")
	out := readOutput(ch, aClient)
	if !strings.Contains(out, "hasn't challenged you") {
		t.Errorf("expected not-challenger message; got %q", out)
	}
}

func TestDoDecline_NPCorSelf(t *testing.T) {
	room, _, _ := setupArenaWorld()
	SetArenaIsChallenge(true)
	ch, client := addPC(room, "Alice")
	defer client.Close()
	ch.Act.Set(types.ACT_CHALLENGER)
	DoDecline(ch, "Alice")
	out := readOutput(ch, client)
	if !strings.Contains(out, "How did that happen") {
		t.Errorf("expected self-decline message; got %q", out)
	}
}

func TestDoDecline_Success(t *testing.T) {
	room, _, _ := setupArenaWorld()
	SetArenaIsChallenge(true)
	decliner, dClient := addPC(room, "Bob")
	defer dClient.Close()
	challenger, cClient := addPC(room, "Alice")
	defer cClient.Close()
	challenger.Act.Set(types.ACT_CHALLENGER)

	DoDecline(decliner, "Alice")

	if ArenaIsChallenge() {
		t.Error("arenaState.IsChallenge should be false after decline")
	}
	// C bug preserved: ch (decliner) gets ACT_CHALLENGER cleared; the
	// decliner never had it, so this is a no-op cosmetic. What matters
	// is that the challenger's flag is LEFT SET — which is the C bug.
	if !challenger.Act.IsSet(types.ACT_CHALLENGER) {
		t.Error("challenger's ACT_CHALLENGER should still be set (C bug preserved)")
	}
	cOut := readOutput(challenger, cClient)
	if !strings.Contains(cOut, "decline your invitation to death") {
		t.Errorf("challenger should get decline notice; got %q", cOut)
	}
	dOut := readOutput(decliner, dClient)
	if !strings.Contains(dOut, "choosen not to die") {
		t.Errorf("decliner should get confirmation; got %q", dOut)
	}
}

// --- DoWithdraw ---

func TestDoWithdraw_NPC(t *testing.T) {
	room, _, _ := setupArenaWorld()
	ch, client := addPC(room, "Orc")
	defer client.Close()
	ch.Act.Set(types.ACT_IS_NPC)
	DoWithdraw(ch, "Alice")
	out := readOutput(ch, client)
	if out != "" {
		t.Errorf("NPC withdraw should be silent; got %q", out)
	}
}

func TestDoWithdraw_NotChallenger(t *testing.T) {
	room, _, _ := setupArenaWorld()
	ch, client := addPC(room, "Alice")
	defer client.Close()
	DoWithdraw(ch, "Bob")
	out := readOutput(ch, client)
	if !strings.Contains(out, "You haven't challenged anyone") {
		t.Errorf("expected not-challenger message; got %q", out)
	}
}

func TestDoWithdraw_VictimNotFound(t *testing.T) {
	room, _, _ := setupArenaWorld()
	ch, client := addPC(room, "Alice")
	defer client.Close()
	ch.Act.Set(types.ACT_CHALLENGER)
	DoWithdraw(ch, "Ghost")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Does it look like they are here") {
		t.Errorf("expected not-found message; got %q", out)
	}
}

func TestDoWithdraw_ArenaBusy(t *testing.T) {
	room, _, _ := setupArenaWorld()
	SetArenaIsBusy(true)
	ch, aClient := addPC(room, "Alice")
	defer aClient.Close()
	_, bClient := addPC(room, "Bob")
	defer bClient.Close()
	ch.Act.Set(types.ACT_CHALLENGER)
	DoWithdraw(ch, "Bob")
	out := readOutput(ch, aClient)
	if !strings.Contains(out, "how does that work again") {
		t.Errorf("expected arena-busy message; got %q", out)
	}
}

func TestDoWithdraw_VictimNPC(t *testing.T) {
	room, _, _ := setupArenaWorld()
	ch, aClient := addPC(room, "Alice")
	defer aClient.Close()
	vic, bClient := addPC(room, "Orc")
	defer bClient.Close()
	ch.Act.Set(types.ACT_CHALLENGER)
	vic.Act.Set(types.ACT_IS_NPC)
	DoWithdraw(ch, "Orc")
	out := readOutput(ch, aClient)
	if !strings.Contains(out, "brain dead how many drugs") {
		t.Errorf("expected brain-dead message; got %q", out)
	}
}

func TestDoWithdraw_Success(t *testing.T) {
	room, _, _ := setupArenaWorld()
	SetArenaIsChallenge(true)
	challenger, cClient := addPC(room, "Alice")
	defer cClient.Close()
	target, tClient := addPC(room, "Bob")
	defer tClient.Close()
	challenger.Act.Set(types.ACT_CHALLENGER)
	handler.AddTimer(challenger, types.TIMER_CHALLENGE, 5, "", 0)

	DoWithdraw(challenger, "Bob")

	if ArenaIsChallenge() {
		t.Error("arenaState.IsChallenge should be false after withdraw")
	}
	if challenger.Act.IsSet(types.ACT_CHALLENGER) {
		t.Error("challenger's ACT_CHALLENGER should be cleared")
	}
	if handler.HasTimer(challenger, types.TIMER_CHALLENGE) {
		t.Error("TIMER_CHALLENGE should be removed after withdraw")
	}
	cOut := readOutput(challenger, cClient)
	if !strings.Contains(cOut, "withdraw your challenge") {
		t.Errorf("challenger should see confirmation; got %q", cOut)
	}
	tOut := readOutput(target, tClient)
	if !strings.Contains(tOut, "withdrawn their challenge") {
		t.Errorf("target should see withdraw notice; got %q", tOut)
	}
}

// pickArenaRoom — no valid arena rooms means nil return.
func TestPickArenaRoom_NoArenaFlaggedRooms(t *testing.T) {
	_ = setupCommWorld() // clean world, no arena rooms
	ResetArenaState()
	if r := pickArenaRoom(); r != nil {
		t.Errorf("pickArenaRoom with no arena-flagged rooms should be nil; got %+v", r)
	}
}
