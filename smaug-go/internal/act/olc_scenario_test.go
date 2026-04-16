package act_test

// G6 scenario test: OLC redit subcommand round-trip. Exercises the
// testclient harness end-to-end against the tier5 fixture: an immortal
// edits the current room's sector via `redit sector N` and then
// observes the new value via `rstat`. The interactive `redit desc`
// string-editor path is out of scope for this scenario.

import (
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/testclient"
	"github.com/eilidhmae/smaug/internal/types"
)

// tier5FixtureDir is the path to the G4 tier5 fixture, resolved from
// the act package's test cwd (go test runs in the package dir).
const tier5FixtureDir = "../testclient/testdata"

// TestScenario_RedIt_SubcommandRoundTrip verifies that an immortal can
// edit a room via non-interactive redit subcommands and see the change
// reflected via rstat.
func TestScenario_RedIt_SubcommandRoundTrip(t *testing.T) {
	h := testclient.Start(t, testclient.WithDataDir(tier5FixtureDir))

	c := h.NewCharacter(t, testclient.CharSpec{
		Name:     "Oimmort",
		Password: "secret",
		Trust:    types.LEVEL_IMMORTAL,
	})

	// Newly-created character is placed at ROOM_VNUM_TEMPLE (21001),
	// whose default sector in the fixture is 0 (inside). Flip to 1.
	c.Send("redit sector 1")
	c.ReadUntil("Sector set to 1", 2*time.Second)

	c.Send("rstat")
	out := c.ReadFor(500 * time.Millisecond)
	if !strings.Contains(out, "Sector: 1") {
		t.Fatalf("rstat output missing new sector; got:\n%s", out)
	}
}
