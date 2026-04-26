package act

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
)

// DoFoldarea implements the 'foldarea' command: save a named area to disk
// by filename, with .bak rotation. Mirrors C do_foldarea at
// src/build.c:8055-8081. Plan plan-phase6-foldarea.md §D2.
//
// Syntax: foldarea <filename>
//
// Looks up the area in WorldRef.Areas by case-insensitive Filename match
// and delegates to writeAreaToDisk (which performs path validation, .bak
// rotation, and the atomic tmp→live rename — all shared with DoSaveArea).
func DoFoldarea(ch *types.CharData, argument string) {
	if ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}
	arg := strings.TrimSpace(argument)
	if arg == "" {
		ch.Send("Fold what?\n\r")
		return
	}
	var found *types.AreaData
	for _, area := range WorldRef.Areas {
		if strings.EqualFold(area.Filename, arg) {
			found = area
			break
		}
	}
	if found == nil {
		ch.Send("No such area exists.\n\r")
		return
	}
	ch.Send("Folding area...\n\r")
	if err := writeAreaToDisk(found); err != nil {
		if err.Error() == "invalid area filename" {
			ch.Send("Invalid area filename.\n\r")
			return
		}
		ch.Sendf("Error folding area: %v\n\r", err)
		return
	}
	ch.Send("Done.\n\r")
}

// DoUnfoldarea implements the 'unfoldarea' command. The full C semantic
// (load an .are file post-boot) is unsafe in the Go port: the area
// loader at internal/persist/area.go:48-82 is not re-entrant — it
// unconditionally appends to w.Areas and re-populates the index maps
// without removing prior entries, which would corrupt the world for any
// already-loaded area. Until a safe reload path exists, direct builders
// to use the existing hotboot facility, which reloads the entire world
// atomically.
//
// C reference (with Thoric's own "Use of this command is not recommended"
// warning at src/build.c:8027-8035): src/build.c:8036-8052.
//
// Plan plan-phase6-foldarea.md §D3 — explicit scope-down.
func DoUnfoldarea(ch *types.CharData, argument string) {
	if ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}
	arg := strings.TrimSpace(argument)
	if arg == "" {
		ch.Send("Unfold what?\n\r")
		return
	}
	ch.Send("Post-boot area reload is not supported in the Go port.\n\r")
	ch.Send("Use 'hotboot' to reload the entire world atomically, or\n\r")
	ch.Send("restart the server to pick up changes to a single area.\n\r")
}
