package mudprog_test

// G6 scenario test: walking into a room with a greet_prog mob fires
// the prog and the player sees its output. The tier5 fixture's mob
// 9101 lives in room 21004 (South Garden, south of Temple) and has a
// 100%-chance greet_prog that emits "The shopkeeper eyes you warily."

import (
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/testclient"
)

// tier5FixtureDir is the path to the G4 tier5 fixture, resolved from
// the mudprog package's test cwd.
const tier5FixtureDir = "../testclient/testdata"

// TestScenario_GreetProgFires verifies that walking into a room with a
// mob that has a greet_prog triggers the prog and the player sees its
// output.
func TestScenario_GreetProgFires(t *testing.T) {
	h := testclient.Start(t, testclient.WithDataDir(tier5FixtureDir))

	c := h.QuickLogin(t, "Greeter")

	// From ROOM_VNUM_TEMPLE (21001), the D2 (south) exit leads to
	// 21004, the South Garden, where mob 9101 (greeter) is reset.
	c.Send("south")
	out := c.ReadFor(1 * time.Second)

	if !strings.Contains(out, "The shopkeeper eyes you warily.") {
		t.Fatalf("greet_prog output not observed; got:\n%s", out)
	}
}
