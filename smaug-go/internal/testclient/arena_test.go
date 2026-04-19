// Phase 6 Arena PvP — end-to-end testclient scenarios. See
// plan-phase6-arena.md §G7.
//
// The fixture area (testdata/area/tier5test.are) does not carry the
// arena rooms. Each scenario below installs a pair of arena-flagged
// rooms via `h.Query` before driving the challenge commands — that
// keeps the arena feature exercised end-to-end via the real command
// interpreter, real network pipeline, and real charUpdate tick.
package testclient

import (
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/act"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// injectArenaRooms adds two ROOM_ARENA-flagged rooms at the arena vnum
// range so DoAccept's pickArenaRoom has valid targets. Safe to call
// multiple times; second call is a no-op (skips if already present).
func injectArenaRooms(t *testing.T, h *Harness) {
	t.Helper()
	h.Query(func(w *world.World) {
		for i, v := 0, types.ROOM_VNUM_ARENA_MIN; i < 2 && v <= types.ROOM_VNUM_ARENA_MAX; i, v = i+1, v+1 {
			if w.Rooms[v] != nil {
				continue
			}
			r := &types.RoomIndexData{
				Vnum:        v,
				Name:        "Arena",
				Description: "A sandy combat arena.\n\r",
				SectorType:  types.SECT_FIELD,
			}
			r.RoomFlags.Set(types.ROOM_ARENA)
			w.Rooms[v] = r
		}
		// And also ensure ROOM_VNUM_ALTAR is loaded for the victory
		// loser-teleport path.
		if w.Rooms[types.ROOM_VNUM_ALTAR] == nil {
			alt := &types.RoomIndexData{
				Vnum:        types.ROOM_VNUM_ALTAR,
				Name:        "Altar",
				Description: "A polished altar stands here.\n\r",
				SectorType:  types.SECT_INSIDE,
			}
			w.Rooms[types.ROOM_VNUM_ALTAR] = alt
		}
	})
}

// TestArena_ChallengeAndDecline drives the happy decline path via two
// concurrent sessions. Alice challenges Bob; Bob declines; arena state
// is cleared.
func TestArena_ChallengeAndDecline(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))
	injectArenaRooms(t, h)
	// Clean arena state in case a prior test in the same binary
	// polluted the package-level flags.
	act.ResetArenaState()
	defer act.ResetArenaState()

	aClient, bClient := h.QuickLoginTwo(t, "Alice", "Bob")
	defer aClient.Close()
	defer bClient.Close()

	// Both PCs default to Level 1 from QuickLogin. The level-5 victim
	// gate inside DoChallenge would bounce Alice's challenge; elevate
	// Bob to L10 via the admin hook.
	var alice, bob *types.CharData
	h.Query(func(w *world.World) {
		alice = findByName(w, "Alice")
		bob = findByName(w, "Bob")
		if alice != nil {
			alice.Level = 10
		}
		if bob != nil {
			bob.Level = 10
		}
	})
	if alice == nil || bob == nil {
		t.Fatalf("both PCs must be in-world: alice=%v bob=%v", alice, bob)
	}

	aClient.Send("challenge Bob")
	aOut := aClient.ReadUntil("Your challenge has been sent", 3*time.Second)
	if !strings.Contains(aOut, "Your challenge has been sent") {
		t.Fatalf("challenger should see confirmation; got %q", aOut)
	}
	bOut := bClient.ReadUntil("You have been challenged by Alice", 3*time.Second)
	if !strings.Contains(bOut, "You have been challenged by Alice") {
		t.Fatalf("victim should see challenge notice; got %q", bOut)
	}

	if !act.ArenaIsChallenge() {
		t.Error("arenaState.IsChallenge should be true after challenge")
	}

	bClient.Send("decline Alice")
	dOut := bClient.ReadUntil("choosen not to die", 3*time.Second)
	if !strings.Contains(dOut, "choosen not to die") {
		t.Errorf("decliner should see confirmation; got %q", dOut)
	}
	cOut := aClient.ReadUntil("decline your invitation", 3*time.Second)
	if !strings.Contains(cOut, "decline your invitation") {
		t.Errorf("challenger should see decline notice; got %q", cOut)
	}

	if act.ArenaIsChallenge() {
		t.Error("arenaState.IsChallenge should be false after decline")
	}
}

// TestArena_ChallengeAndWithdraw drives challenger-side cancellation.
func TestArena_ChallengeAndWithdraw(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))
	injectArenaRooms(t, h)
	act.ResetArenaState()
	defer act.ResetArenaState()

	aClient, bClient := h.QuickLoginTwo(t, "Alidraw", "Bobdraw")
	defer aClient.Close()
	defer bClient.Close()

	h.Query(func(w *world.World) {
		if a := findByName(w, "Alidraw"); a != nil {
			a.Level = 10
		}
		if b := findByName(w, "Bobdraw"); b != nil {
			b.Level = 10
		}
	})

	aClient.Send("challenge Bobdraw")
	_ = aClient.ReadUntil("Your challenge has been sent", 3*time.Second)
	_ = bClient.ReadUntil("You have been challenged by", 3*time.Second)
	if !act.ArenaIsChallenge() {
		t.Fatal("IsChallenge should be true")
	}

	aClient.Send("withdraw Bobdraw")
	wOut := aClient.ReadUntil("withdraw your challenge", 3*time.Second)
	if !strings.Contains(wOut, "withdraw your challenge") {
		t.Errorf("challenger should see withdraw confirmation; got %q", wOut)
	}
	bOut := bClient.ReadUntil("withdrawn their challenge", 3*time.Second)
	if !strings.Contains(bOut, "withdrawn their challenge") {
		t.Errorf("target should see withdraw notice; got %q", bOut)
	}

	if act.ArenaIsChallenge() {
		t.Error("IsChallenge should be false after withdraw")
	}
	var alice *types.CharData
	h.Query(func(w *world.World) {
		alice = findByName(w, "Alidraw")
	})
	if alice != nil && alice.Act.IsSet(types.ACT_CHALLENGER) {
		t.Error("challenger's ACT_CHALLENGER should be cleared after withdraw")
	}
}

// TestArena_ChallengeAndAccept_TeleportsBoth pins the happy-path accept
// teleport through to the arena room. No kill — combat-victory is
// exercised by combat package unit tests. Here we verify the
// accept-time state transitions land end-to-end.
func TestArena_ChallengeAndAccept_TeleportsBoth(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))
	injectArenaRooms(t, h)
	act.ResetArenaState()
	defer act.ResetArenaState()

	aClient, bClient := h.QuickLoginTwo(t, "Aliaccept", "Bobaccept")
	defer aClient.Close()
	defer bClient.Close()

	h.Query(func(w *world.World) {
		if a := findByName(w, "Aliaccept"); a != nil {
			a.Level = 10
		}
		if b := findByName(w, "Bobaccept"); b != nil {
			b.Level = 10
		}
	})

	aClient.Send("challenge Bobaccept")
	_ = aClient.ReadUntil("Your challenge has been sent", 3*time.Second)
	_ = bClient.ReadUntil("You have been challenged by", 3*time.Second)

	bClient.Send("accept Aliaccept")
	acceptAck := bClient.ReadUntil("plop onto the ground", 3*time.Second)
	if !strings.Contains(acceptAck, "plop onto the ground") {
		t.Errorf("accepter should see plop message; got %q", acceptAck)
	}

	var alice, bob *types.CharData
	h.Query(func(w *world.World) {
		alice = findByName(w, "Aliaccept")
		bob = findByName(w, "Bobaccept")
	})

	if alice == nil || bob == nil {
		t.Fatal("both PCs should be in-world")
	}
	if alice.InRoom == nil || !alice.InRoom.RoomFlags.IsSet(types.ROOM_ARENA) {
		t.Errorf("Aliaccept should be in ROOM_ARENA; got %+v", alice.InRoom)
	}
	if bob.InRoom == nil || !bob.InRoom.RoomFlags.IsSet(types.ROOM_ARENA) {
		t.Errorf("Bobaccept should be in ROOM_ARENA; got %+v", bob.InRoom)
	}
	if !alice.Act.IsSet(types.PLR_SILENCE) || !bob.Act.IsSet(types.PLR_SILENCE) {
		t.Error("both PCs should have PLR_SILENCE set")
	}
	// Accepter is bob (ch), so bob gets ACT_CHALLENGED.
	if !bob.Act.IsSet(types.ACT_CHALLENGED) {
		t.Error("accepter (Bobaccept) should have ACT_CHALLENGED set")
	}
	if !act.ArenaIsBusy() {
		t.Error("arenaState.IsBusy should be true after accept")
	}
	if act.ArenaIsChallenge() {
		t.Error("arenaState.IsChallenge should be cleared after accept")
	}
}
