package magic_test

// G6 scenario test: cast sanctuary on self via the testclient harness
// and verify the AFF_SANCTUARY bit is set on the live character
// inspected through Harness.Query. Query is more reliable than parsing
// `score`/`affects` output text — it reads the world state directly.

import (
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/testclient"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// tier5FixtureDir is the path to the G4 tier5 fixture, resolved from
// the magic package's test cwd.
const tier5FixtureDir = "../testclient/testdata"

// TestScenario_CastSanctuary verifies that an immortal can cast
// sanctuary on self and see the affect in their score/affects list.
func TestScenario_CastSanctuary(t *testing.T) {
	h := testclient.Start(t, testclient.WithDataDir(tier5FixtureDir))

	c := h.NewCharacter(t, testclient.CharSpec{
		Name:     "Mage",
		Password: "secret",
		Trust:    types.LEVEL_IMMORTAL,
	})

	// Sanctuary is a defensive spell; with no arg2 it targets self.
	c.Send("cast sanctuary")
	_ = c.ReadFor(1 * time.Second)

	// Verify via Query — more reliable than parsing output text.
	var hasSanct bool
	var foundChar bool
	h.Query(func(w *world.World) {
		for _, d := range w.Descriptors {
			if d.Character == nil {
				continue
			}
			if strings.EqualFold(d.Character.Name, "Mage") {
				foundChar = true
				hasSanct = d.Character.AffectedBy.IsSet(types.AFF_SANCTUARY)
				return
			}
		}
	})
	if !foundChar {
		t.Fatal("Mage character not found in world descriptors")
	}
	if !hasSanct {
		t.Fatal("AFF_SANCTUARY not set after `cast sanctuary`")
	}
}
