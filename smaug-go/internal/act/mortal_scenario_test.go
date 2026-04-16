package act_test

// G6 scenario test: mortal alias expansion. A mortal character
// navigates to the pebble room, registers an alias `g` -> `get pebble`,
// fires it, and confirms both the interpreter-visible output and the
// world-state-visible carry via Harness.Query.
//
// tier5FixtureDir is defined in olc_scenario_test.go (same package).

import (
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/testclient"
	"github.com/eilidhmae/smaug/internal/world"
)

// TestScenario_AliasExpansion verifies that mortal-level alias expands
// correctly through the interpreter.
func TestScenario_AliasExpansion(t *testing.T) {
	h := testclient.Start(t, testclient.WithDataDir(tier5FixtureDir))

	c := h.QuickLogin(t, "Aliaser")

	// Temple (21001) D4 (up) leads to 21006 (Upper Gallery), where a
	// pebble (obj vnum 9200) is reset on the floor.
	c.Send("up")
	_ = c.ReadFor(500 * time.Millisecond)

	// Register the alias, drain the ack.
	c.Send("alias g get pebble")
	_ = c.ReadFor(300 * time.Millisecond)

	// Fire the alias. Interpret expands `g` -> `get pebble`, invoking
	// DoGet which sends "You get a small pebble.\n\r" on success.
	c.Send("g")
	out := c.ReadFor(500 * time.Millisecond)
	if !(strings.Contains(out, "You get") || strings.Contains(strings.ToLower(out), "pebble")) {
		t.Fatalf("alias expansion did not produce pickup output; got:\n%s", out)
	}

	// Sanity: the character should now carry the pebble.
	var foundChar, carrying bool
	h.Query(func(w *world.World) {
		for _, d := range w.Descriptors {
			if d.Character == nil {
				continue
			}
			if strings.EqualFold(d.Character.Name, "Aliaser") {
				foundChar = true
				for _, obj := range d.Character.Carrying {
					if obj.IndexData != nil && obj.IndexData.Vnum == 9200 {
						carrying = true
						return
					}
				}
				return
			}
		}
	})
	if !foundChar {
		t.Fatal("Aliaser character not found in world descriptors")
	}
	if !carrying {
		t.Fatal("Aliaser is not carrying the pebble (vnum 9200) after alias expansion")
	}
}
