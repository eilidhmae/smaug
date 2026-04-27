package act

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
)

// WorldPcLookup resolves a connected-PC name to its *CharData by walking
// WorldRef.Descriptors. Public so DoMedit (and future PC-targeting
// commands) can invoke it without importing the game package.
//
// Mirrors C get_char_world() (handler.c:1894-) for the PC subset only.
// Returns the first descriptor whose Connected == CON_PLAYING and whose
// Character.Name matches case-insensitively. Linkdead (Connected !=
// CON_PLAYING) and nil-Character descriptors are skipped.
//
// Tests can override directly without standing up a full world.
//
// See internal/game/medit_parse.go:worldPcLookup for the parallel
// game-package seam used internally by medit's interactive arms.
//
// Plan: plan-phase6-quickwins-blank-pcrename.md §D3 / §G3.
var WorldPcLookup = func(name string) *types.CharData {
	if WorldRef == nil {
		return nil
	}
	for _, d := range WorldRef.Descriptors {
		if d == nil || d.Character == nil {
			continue
		}
		if d.Connected != int(types.CON_PLAYING) {
			continue
		}
		if strings.EqualFold(d.Character.Name, name) {
			return d.Character
		}
	}
	return nil
}
