// Package game — interactive object editor (CON_OEDIT substate).
//
// This file ports src/ooedit.c oedit_parse at :1172-2160. The pulse loop
// dispatches input addressed to a descriptor in CON_OEDIT directly here
// (see loop.go processInput); the nanny is NOT invoked for menu input.
// Mirrors C src/smaug.c:1641-1652 (same routing pattern as CON_REDIT).
//
// Wave 1 scope: skeleton + CON_OEDIT loop arm only. Menu rendering + field
// mutation branches land in Waves 2+ per plan-phase6-olc-oedit.md §G4-G10.
//
// Plan: plan-phase6-olc-oedit.md §G3.
package game

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
)

// worldObjLookup is the seam the parser uses to resolve object vnums.
// Points at worldRef.GetObjIndex at boot time (via SetWorldRef shared
// with redit_parse.go); tests can override directly.
var worldObjLookup = func(vnum int) *types.ObjIndexData {
	if worldRef == nil {
		return nil
	}
	return worldRef.GetObjIndex(vnum)
}

// oeditParse is the top-level CON_OEDIT dispatcher. Keyed on d.Olc.Mode,
// it mutates the object under edit and either returns (stay in same mode)
// or transitions to a new mode + redisplays the appropriate menu.
//
// The caller (processInput) has already validated d.Connected ==
// CON_OEDIT. If d.Olc is nil or d.Olc.Target is not an *ObjIndexData
// (should not happen — stale state bug elsewhere), cleanup and drop
// back to CON_PLAYING defensively.
//
// Wave 1: only "Q"/"q" is wired; every other input emits a placeholder
// stub message. Waves 2+ fill the dispatch arms.
func oeditParse(d *types.DescriptorData, arg string) {
	if d == nil {
		return
	}
	if d.Olc == nil || d.Character == nil {
		// Defensive: impossible state; restore playable.
		d.Connected = int(types.CON_PLAYING)
		return
	}
	idx, ok := d.Olc.Target.(*types.ObjIndexData)
	if !ok || idx == nil {
		// Target lost (should not happen). Restore playable.
		cleanupOlc(d)
		return
	}
	_ = idx // will be used by mode branches in Wave 2+

	trimmed := strings.TrimSpace(arg)
	if trimmed == "Q" || trimmed == "q" {
		d.WriteToBuffer("Exiting editor.\n\r")
		cleanupOlc(d)
		return
	}

	// Wave 1 stub — Wave 2 replaces this with the OEDIT_* dispatch switch.
	d.WriteToBuffer("Oedit menu not yet implemented — Q to quit.\n\r")
}
