package combat_test

// G6 scenario test: mob creation + kill round-trip via the testclient
// harness. Exercises mcreate (immortal-only) to spawn a fresh kill
// target, then `kill <keyword>` to initiate combat, and finally drains
// the output to confirm at least one damage-message verb or the
// killed/XP ack made it back to the client.

import (
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/testclient"
	"github.com/eilidhmae/smaug/internal/types"
)

// tier5FixtureDir is the path to the G4 tier5 fixture, resolved from
// the combat package's test cwd.
const tier5FixtureDir = "../testclient/testdata"

// TestScenario_MobCreateAndKill verifies that an immortal can create a
// mob via mcreate and kill it, capturing the damage-message output
// from combat.DamMessage.
func TestScenario_MobCreateAndKill(t *testing.T) {
	h := testclient.Start(t, testclient.WithDataDir(tier5FixtureDir))

	c := h.NewCharacter(t, testclient.CharSpec{
		Name:     "Fighter",
		Password: "secret",
		Trust:    types.LEVEL_IMMORTAL,
	})

	// mcreate <vnum> <name> creates a new mob template and places an
	// instance in the caller's room. Use a vnum not already in the
	// fixture (9100/9101/9102 are taken).
	c.Send("mcreate 9500 rat")
	c.ReadUntil("created", 2 * time.Second)

	// Now a "rat" mob is in the temple. Kill it. Combat runs on the
	// violence pulse (every 3 seconds at 4 pulses/sec), so drain for
	// long enough to catch at least one round — usually one is enough
	// to drop the freshly-minted level-1 mob.
	c.Send("kill rat")
	out := c.ReadFor(5 * time.Second)

	// Damage-message verbs from combat/dammessage.go. Any of these
	// (or the kill-ack "receive" or "DEAD") proves the combat pipeline
	// reached the client.
	verbs := []string{
		"miss", "hit", "pound", "slash", "crush", "punch",
		"bruise", "strike", "thrash", "scratch",
		"receive", // "You receive N experience points." on kill
		"DEAD",
	}
	lowered := strings.ToLower(out)
	for _, v := range verbs {
		if strings.Contains(lowered, strings.ToLower(v)) {
			return
		}
	}
	t.Fatalf("no damage message in combat output; got:\n%s", out)
}
