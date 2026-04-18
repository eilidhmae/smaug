package testclient

import (
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// TestGtell_TwoClients_Concurrent drives the first concurrent-login
// scenario in the test suite: two characters log in, form a group via
// `follow` + `group`, then one gtells and the other receives. See
// plan-channels.md G3 and the QuickLoginTwo helper docstring in login.go.
//
// Why this matters: until now no test has driven two simultaneous client
// sessions against one harness. The channel commands (gtell / immtalk)
// can't be fully end-to-end tested without it — unit tests against
// DoGtell directly don't exercise the command-table registration or the
// real network pipeline.
func TestGtell_TwoClients_Concurrent(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))

	// Log both in concurrently. Both start at the temple (ROOM_VNUM_TEMPLE).
	leaderClient, followerClient := h.QuickLoginTwo(t, "Leadgt", "Folgt")
	defer leaderClient.Close()
	defer followerClient.Close()

	// Form a group. Go port's DoFollow sets both Master *and* Leader
	// (see internal/act/group.go:50-51), which differs from C where
	// `group` toggles Leader separately — a pre-existing divergence
	// unrelated to this plan. `follow` alone is enough to make
	// IsSameGroup return true.
	followerClient.Send("follow Leadgt")
	_ = followerClient.ReadUntil("follow", time.Second)

	// Sanity: verify the Leader field was set via h.Query to avoid a
	// false-positive if grouping silently diverged.
	var leader, follower *types.CharData
	h.Query(func(w *world.World) {
		leader = findByName(w, "Leadgt")
		follower = findByName(w, "Folgt")
	})
	if leader == nil || follower == nil {
		t.Fatalf("both characters must be in-world: leader=%v follower=%v", leader, follower)
	}
	if follower.Leader != leader {
		t.Fatalf("follower.Leader not set to leader pointer (got %v); grouping did not take hold", follower.Leader)
	}

	// Leader gtells. Follower must receive the group-tell line.
	leaderClient.Send("gtell assemble for combat")

	got := followerClient.ReadUntil("assemble for combat", 3*time.Second)
	if !strings.Contains(got, "assemble for combat") {
		t.Errorf("follower did not receive gtell payload within 3s; got %q", got)
	}
	if !strings.Contains(got, "Leadgt") {
		t.Errorf("follower should see sender name 'Leadgt'; got %q", got)
	}
	if !strings.Contains(got, "tells the group") {
		t.Errorf("follower should see 'tells the group' phrase; got %q", got)
	}
}

// findByName returns the live CharData matching the given name (case-
// insensitive) by walking world descriptors. Must be called from inside
// an h.Query block so the read is serialized with the game loop.
func findByName(w *world.World, name string) *types.CharData {
	for _, d := range w.Descriptors {
		if d.Character != nil && strings.EqualFold(d.Character.Name, name) {
			return d.Character
		}
	}
	return nil
}
