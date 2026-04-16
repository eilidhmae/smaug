package testclient

// G4 fixture smoke tests. These verify the data tree under
// internal/testclient/testdata/ boots cleanly via boot.Boot and contains
// the specific rooms/mobs/objects that G6 scenario tests rely on.
//
// This test does NOT exercise the testclient harness itself — it pokes
// the world directly after a bare Boot, keeping the assertions focused
// on the fixture rather than on network plumbing.
//
// Fixture scope / expected absences:
// The minimal G4 fixture deliberately omits clans/, deity/, and
// boards/boards.dat. boot.Boot logs WARNING lines for each on boot;
// these are expected and non-fatal (matches C behavior of continuing
// past missing optional subsystems). Scenario tests that need clans,
// deities, or bulletin boards must extend the fixture — do not treat
// the warnings as regressions.
//
// Package-globals guard: boot.Boot mutates cross-package globals
// (act.WorldRef, act.SaveFunc, mudprog.WorldRef, combat hooks, …) that
// Harness.Start also clobbers. bootFixture must hold harnessMu for the
// lifetime of the test, same as Harness.Start does — otherwise a
// parallel harness-based test running in the same process can corrupt
// those globals mid-boot. This file lives in package `testclient`
// (not `testclient_test`) so it can share the unexported lockHarness
// helper defined alongside Start in harness.go.

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/boot"
	smaugnet "github.com/eilidhmae/smaug/internal/net"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// fixtureDir resolves to internal/testclient/testdata relative to this
// test file's package cwd.
const fixtureDir = "testdata"

// Known vnums assigned by tier5test.are. These MUST stay in sync with
// the .are file — the whole point of the fixture is deterministic vnums.
const (
	vnumTemple    = types.ROOM_VNUM_TEMPLE // 21001, temple
	vnumNorecall  = 21002                  // north of temple, ROOM_NO_RECALL
	vnumEast      = 21003                  // empty plain room
	vnumSouth     = 21004                  // hosts greet-mob via reset
	vnumWest      = 21005                  // hosts banker-mob via reset
	vnumUp        = 21006                  // hosts pebble via reset

	mvnumFiller  = 9100
	mvnumGreeter = 9101
	mvnumBanker  = 9102

	ovnumPebble = 9200
)

// bootFixture boots the G4 fixture once and returns the world plus the
// observed boot duration. On failure, t.Fatal is called.
//
// Acquires harnessMu for the duration of the test so boot.Boot's
// clobbering of package-level globals (act.WorldRef, act.SaveFunc,
// mudprog.WorldRef, combat hooks, …) can't race with a concurrent
// harness-based test running in the same test binary.
func bootFixture(t *testing.T) (*world.World, time.Duration) {
	t.Helper()
	lockHarness(t)

	absDir, err := filepath.Abs(fixtureDir)
	if err != nil {
		t.Fatalf("resolve fixture dir: %v", err)
	}

	w := world.New(absDir)
	server := smaugnet.NewServer()

	start := time.Now()
	_, _, err = boot.Boot(w, absDir, server.Incoming, boot.ProductionOpts())
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("boot: %v", err)
	}
	// Document but do not enforce: Phase 5 Tier 5 plan target is <500ms;
	// <3s is acceptable on CI hardware.
	t.Logf("G4 fixture boot time: %s", elapsed)
	return w, elapsed
}

func TestFixture_Loads(t *testing.T) {
	w, _ := bootFixture(t)

	if len(w.Rooms) < 6 {
		t.Errorf("expected at least 6 rooms; got %d", len(w.Rooms))
	}
	if temple := w.GetRoom(types.ROOM_VNUM_TEMPLE); temple == nil {
		t.Fatalf("Temple room (vnum %d) not loaded", types.ROOM_VNUM_TEMPLE)
	}
}

// TestFixture_HoldsHarnessMu is a structural regression guard for the
// Phase 5 Tier 5 G3 fix: bootFixture MUST hold harnessMu so the
// package-level globals boot.Boot mutates cannot race against a
// parallel harness-based test. The test asserts the mutex is locked at
// the moment bootFixture returns by confirming TryLock fails.
func TestFixture_HoldsHarnessMu(t *testing.T) {
	_, _ = bootFixture(t)
	if harnessMu.TryLock() {
		harnessMu.Unlock()
		t.Fatal("harnessMu must be held by bootFixture; TryLock unexpectedly succeeded")
	}
}

func TestFixture_RoomsHaveCorrectVnums(t *testing.T) {
	w, _ := bootFixture(t)

	for _, vnum := range []int{vnumTemple, vnumNorecall, vnumEast, vnumSouth, vnumWest, vnumUp} {
		if r := w.GetRoom(vnum); r == nil {
			t.Errorf("room vnum %d missing from fixture", vnum)
		}
	}
}

func TestFixture_NorecallRoomFlag(t *testing.T) {
	w, _ := bootFixture(t)

	r := w.GetRoom(vnumNorecall)
	if r == nil {
		t.Fatalf("room %d missing", vnumNorecall)
	}
	if !r.RoomFlags.IsSet(types.ROOM_NO_RECALL) {
		t.Errorf("room %d should have ROOM_NO_RECALL set; flags bits=%v",
			vnumNorecall, r.RoomFlags.Bits())
	}
}

func TestFixture_MobTemplates(t *testing.T) {
	w, _ := bootFixture(t)

	if len(w.MobIndex) < 3 {
		t.Errorf("expected at least 3 mob templates; got %d", len(w.MobIndex))
	}

	t.Run("greet-mob has non-empty progs", func(t *testing.T) {
		mob := w.GetMobIndex(mvnumGreeter)
		if mob == nil {
			t.Fatalf("greet-mob vnum %d not found", mvnumGreeter)
		}
		if len(mob.MudProgs) == 0 {
			t.Errorf("greet-mob should have at least one mudprog; got 0")
		}
		// MPROG_GREET is defined as (1<<7) in types/mudprog.go, and
		// persist/area.go:513 sets ProgTypes via
		// bits.TrailingZeros64(uint64(prog.Type)), so the bit-index the
		// loader stores for a greet_prog is 7, not 0. Assert that.
		const mprogGreetBitIndex = 7
		if !mob.ProgTypes.IsSet(mprogGreetBitIndex) {
			t.Errorf("greet-mob ProgTypes missing MPROG_GREET (bit %d); bits=%v",
				mprogGreetBitIndex, mob.ProgTypes.Bits())
		}
		// Belt-and-suspenders: at least one parsed MudProg should have
		// Type == MPROG_GREET. This invariant holds regardless of
		// whether ProgTypes is populated (round-2 concern).
		foundGreet := false
		for _, p := range mob.MudProgs {
			if p.Type == types.MPROG_GREET {
				foundGreet = true
				break
			}
		}
		if !foundGreet {
			t.Errorf("greet-mob has no MudProgs entry with Type == MPROG_GREET (%d); got %+v",
				types.MPROG_GREET, mob.MudProgs)
		}
	})

	t.Run("banker-mob present", func(t *testing.T) {
		// TODO(G4): ACT_BANKER is bit 42, which cannot be set via the
		// simple 32-bit actFlags int in the area format the current
		// loader accepts. G6 assertions that need "banker" behavior
		// should match on the mob's keywords/short-descr, not on
		// ACT_BANKER. See fixture comment in tier5test.are.
		mob := w.GetMobIndex(mvnumBanker)
		if mob == nil {
			t.Fatalf("banker-mob vnum %d not found", mvnumBanker)
		}
		if !strings.Contains(strings.ToLower(mob.PlayerName), "banker") {
			t.Errorf("banker-mob keywords missing 'banker'; got %q", mob.PlayerName)
		}
	})
}

func TestFixture_ObjectTemplates(t *testing.T) {
	w, _ := bootFixture(t)

	if len(w.ObjIndex) < 1 {
		t.Errorf("expected at least 1 object template; got %d", len(w.ObjIndex))
	}
	obj := w.GetObjIndex(ovnumPebble)
	if obj == nil {
		t.Fatalf("pebble obj vnum %d not found", ovnumPebble)
	}
	if !strings.Contains(strings.ToLower(obj.Name), "pebble") {
		t.Errorf("pebble keywords missing 'pebble'; got %q", obj.Name)
	}
}

func TestFixture_ResetsPopulateRooms(t *testing.T) {
	w, _ := bootFixture(t)

	t.Run("pebble on ground in upper gallery", func(t *testing.T) {
		r := w.GetRoom(vnumUp)
		if r == nil {
			t.Fatalf("room %d missing", vnumUp)
		}
		found := false
		for _, o := range r.Contents {
			if o.IndexData != nil && o.IndexData.Vnum == ovnumPebble {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("pebble (vnum %d) not on the ground in room %d after reset",
				ovnumPebble, vnumUp)
		}
	})

	t.Run("greet-mob in south garden", func(t *testing.T) {
		r := w.GetRoom(vnumSouth)
		if r == nil {
			t.Fatalf("room %d missing", vnumSouth)
		}
		found := false
		for _, ch := range r.People {
			if ch.IndexData != nil && ch.IndexData.Vnum == mvnumGreeter {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("greet-mob (vnum %d) not in room %d after reset",
				mvnumGreeter, vnumSouth)
		}
	})

	t.Run("banker-mob in west vault", func(t *testing.T) {
		r := w.GetRoom(vnumWest)
		if r == nil {
			t.Fatalf("room %d missing", vnumWest)
		}
		found := false
		for _, ch := range r.People {
			if ch.IndexData != nil && ch.IndexData.Vnum == mvnumBanker {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("banker-mob (vnum %d) not in room %d after reset",
				mvnumBanker, vnumWest)
		}
	})
}
