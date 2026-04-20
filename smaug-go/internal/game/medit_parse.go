// Package game — interactive mob/character editor (CON_MEDIT substate).
//
// This file ports src/omedit.c medit_parse at :985-2160. The pulse loop
// dispatches input addressed to a descriptor in CON_MEDIT directly here
// (see loop.go processInput); the nanny is NOT invoked for menu input.
// Mirrors C src/smaug.c (same routing pattern as CON_REDIT / CON_OEDIT).
//
// Wave 1 scope: skeleton + CON_MEDIT loop arm only. The victim is
// type-asserted out of d.Olc.Target as *types.CharData (medit edits BOTH
// NPC prototypes AND connected PCs — both use the same CharData struct).
// Menu content + field dispatch arms land in Waves 2+ per
// plan-phase6-olc-medit.md §G5-G14.
//
// Plan: plan-phase6-olc-medit.md §G4.
package game

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
)

// worldMobLookup is the seam the parser uses to resolve mob prototype
// vnums. Points at worldRef.GetMobIndex at boot time (worldRef /
// SetWorldRef are shared with redit_parse.go / oedit_parse.go); tests
// can override directly without standing up a full world.
var worldMobLookup = func(vnum int) *types.MobIndexData {
	if worldRef == nil {
		return nil
	}
	return worldRef.GetMobIndex(vnum)
}

// meditParse is the top-level CON_MEDIT dispatcher. Keyed on d.Olc.Mode,
// it mutates the victim (an NPC prototype OR a connected PC) and either
// returns (stay in same mode) or transitions to a new mode + redisplays
// the appropriate NPC / PC main menu.
//
// The caller (processInput) has already validated d.Connected ==
// CON_MEDIT. If d.Olc is nil, d.Character is nil, or d.Olc.Target is not
// a *CharData (stale state bug elsewhere), cleanupOlc fires defensively
// and the descriptor drops back to CON_PLAYING.
//
// Wave 1: only "Q"/"q" is wired; every other input emits a placeholder
// stub message. Waves 2+ fill the MEDIT_* dispatch arms — 64 modes,
// NPC-vs-PC ownership rules, bitmask editors, password trampoline.
func meditParse(d *types.DescriptorData, arg string) {
	if d == nil {
		return
	}
	if d.Olc == nil || d.Character == nil {
		// Defensive: impossible state; restore playable.
		d.Connected = int(types.CON_PLAYING)
		return
	}
	victim, ok := d.Olc.Target.(*types.CharData)
	if !ok || victim == nil {
		// Target lost / wrong type (should not happen). Restore playable.
		cleanupOlc(d)
		return
	}
	_ = victim // will be used by mode branches in Wave 2+

	trimmed := strings.TrimSpace(arg)
	if trimmed == "Q" || trimmed == "q" {
		d.WriteToBuffer("Exiting editor.\n\r")
		cleanupOlc(d)
		return
	}

	// Wave 1 stub — Wave 2 replaces this with the MEDIT_* dispatch switch.
	d.WriteToBuffer("Medit menu not yet implemented — Q to quit.\n\r")
}
