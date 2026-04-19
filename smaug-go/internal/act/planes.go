package act

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// PlanesFilePath is the absolute path to the planes data file. Written
// once at boot by boot.go so DoPset `save` can persist the in-memory
// slice back to disk. Package-level parity with act.WorldRef (set-at-boot,
// read-from-game-loop; no synchronization needed).
var PlanesFilePath string

// planeLookup mirrors C plane_lookup at src/planes.c:149-161: first pass
// is case-insensitive exact match; second pass is case-insensitive
// prefix match on stored names (so input "prim" resolves to
// "Prime Material"). Returns nil when nothing matches.
func planeLookup(list []*types.PlaneData, name string) *types.PlaneData {
	// Pass 1: exact case-insensitive match.
	for _, p := range list {
		if p != nil && strings.EqualFold(p.Name, name) {
			return p
		}
	}
	// Pass 2: case-insensitive prefix match (stored name starts with input).
	nameLower := strings.ToLower(name)
	for _, p := range list {
		if p != nil && strings.HasPrefix(strings.ToLower(p.Name), nameLower) {
			return p
		}
	}
	return nil
}

// DoPlist lists every plane, one per line. Player-visible (Level 0).
// Ports C do_plist at src/planes.c:50-59.
func DoPlist(ch *types.CharData, argument string) {
	if ch == nil {
		return
	}
	ch.Send("Planes:\n-------\n")
	if WorldRef == nil {
		return
	}
	for _, p := range WorldRef.Planes {
		if p == nil {
			continue
		}
		ch.Sendf("%s\n\r", p.Name)
	}
}

// DoPstat prints a single plane's stats. Immortal-gated at registration.
// Ports C do_pstat at src/planes.c:61-75. C emits `Name: %s\n`; Go port
// emits `Name: %s\n\r` to match the rest of Go's line-ending convention.
func DoPstat(ch *types.CharData, argument string) {
	if ch == nil {
		return
	}
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Stat which plane?\n\r")
		return
	}
	if WorldRef == nil {
		ch.Send("Stat which plane?\n\r")
		return
	}
	p := planeLookup(WorldRef.Planes, arg)
	if p == nil {
		ch.Send("Stat which plane?\n\r")
		return
	}
	ch.Sendf("Name: %s\n\r", p.Name)
}

// pSetSyntax emits the 7-line syntax-help output. Factored so both the
// empty-arg branch and the unmatched-op fallthrough can call it without
// recursing into DoPset.
func pSetSyntax(ch *types.CharData) {
	ch.Send("Syntax: pset <plane> create\n\r")
	ch.Send("        pset save\n\r")
	ch.Send("        pset <plane> delete\n\r")
	ch.Send("        pset <plane> <field> <value>\n\r")
	ch.Send("\n\r")
	ch.Send("  Where <field> is one of:\n\r")
	ch.Send("    name\n\r")
}

// DoPset implements the plane CRUD dispatcher. Immortal-gated at
// registration (LEVEL_GREATER per plan D6). Ports C do_pset at
// src/planes.c:77-147. See plan-phase6-planes.md §D5 for the full
// dispatch rationale and the mutation-verify table.
func DoPset(ch *types.CharData, argument string) {
	if ch == nil {
		return
	}
	if WorldRef == nil {
		pSetSyntax(ch)
		return
	}

	arg1, rest := util.OneArgument(argument)
	if arg1 == "" {
		pSetSyntax(ch)
		return
	}

	// `pset save` shortcut — persist and return. No further parsing.
	if strings.EqualFold(arg1, "save") {
		if err := persist.SavePlanes(PlanesFilePath, WorldRef.Planes); err != nil {
			util.Bug("DoPset save: %v", err)
		}
		ch.Send("Planes saved.\n\r")
		return
	}

	arg2, rest := util.OneArgument(rest)
	plane := planeLookup(WorldRef.Planes, arg1)

	// C `!str_prefix(mod, "create")` → arg2 is a prefix of "create".
	// Go idiom: strings.HasPrefix(target, prefix) with target = the
	// literal word and prefix = the user's token.
	arg2Lower := strings.ToLower(arg2)

	if arg2Lower != "" && strings.HasPrefix("create", arg2Lower) {
		if plane != nil {
			ch.Send("Plane already exists.\n\r")
			return
		}
		WorldRef.Planes = append(WorldRef.Planes, &types.PlaneData{Name: arg1})
		ch.Send("Plane created.\n\r")
		return
	}

	// Non-create ops require the plane to already exist.
	if plane == nil {
		ch.Send("Plane doesn't exist.\n\r")
		return
	}

	if arg2Lower != "" && strings.HasPrefix("delete", arg2Lower) {
		// Splice-delete by locating index — pointer-identity compare.
		idx := -1
		for i, p := range WorldRef.Planes {
			if p == plane {
				idx = i
				break
			}
		}
		if idx < 0 {
			// Should be unreachable (plane came from planeLookup on the
			// same slice), but guard anyway.
			ch.Send("Plane doesn't exist.\n\r")
			return
		}
		WorldRef.Planes = append(WorldRef.Planes[:idx], WorldRef.Planes[idx+1:]...)
		// After splice, `plane` is no longer in WorldRef.Planes. CheckPlanes
		// will reassign every room whose .Plane points at `plane` to
		// w.Planes[0] (or to a freshly-seeded Prime Material if the slice
		// is now empty).
		persist.CheckPlanes(WorldRef, plane)
		ch.Send("Plane deleted.\n\r")
		return
	}

	if arg2Lower != "" && strings.HasPrefix("name", arg2Lower) {
		newName := strings.TrimSpace(rest)
		if newName == "" {
			pSetSyntax(ch)
			return
		}
		newName = util.SmashTilde(newName)
		if planeLookup(WorldRef.Planes, newName) != nil {
			ch.Send("Another plane has that name.\n\r")
			return
		}
		plane.Name = newName
		ch.Send("Name changed.\n\r")
		return
	}

	// Fallthrough — unknown op. Re-emit syntax help without recursing.
	pSetSyntax(ch)
}
