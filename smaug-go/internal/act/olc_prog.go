package act

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// parseProgTriggerName delegates to util.GetMpFlag (single source of truth
// for the mprog keyword table). Kept as a thin wrapper for backward compat
// with inspector callers.
func parseProgTriggerName(s string) (int64, bool) {
	return util.GetMpFlag(s)
}

// showProg prints a mudprog in a human-readable block.
func showProg(ch *types.CharData, p *types.MProgData) {
	ch.Sendf("Trigger : 0x%x\n\r", p.Type)
	if p.ArgList != "" {
		ch.Sendf("Args    : %s\n\r", p.ArgList)
	}
	ch.Sendf("Script  :\n\r%s\n\r", p.ComList)
	ch.Send("---\n\r")
}

// findProgsByTrigger returns all progs whose Type bitmask includes bit.
func findProgsByTrigger(progs []*types.MProgData, bit int64) []*types.MProgData {
	var out []*types.MProgData
	for _, p := range progs {
		if p.Type&bit != 0 {
			out = append(out, p)
		}
	}
	return out
}

// DoMpedit implements the 'mpedit' command.
// Usage: mpedit <vnum> [trigger]
// If trigger omitted, list all progs on the mob prototype.
// If provided, show progs matching that trigger.
func DoMpedit(ch *types.CharData, argument string) {
	mudprogEdit(ch, argument, "mob")
}

// DoOpedit implements the 'opedit' command (obj-prog inspector).
func DoOpedit(ch *types.CharData, argument string) {
	mudprogEdit(ch, argument, "obj")
}

// DoRpedit implements the 'rpedit' command (room-prog inspector).
func DoRpedit(ch *types.CharData, argument string) {
	mudprogEdit(ch, argument, "room")
}

func mudprogEdit(ch *types.CharData, argument, kind string) {
	if ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}
	vnumArg, rest := util.OneArgument(argument)
	if vnumArg == "" {
		ch.Sendf("Usage: %sedit <vnum> [trigger]\n\r", progEditCmdPrefix(kind))
		return
	}
	vnum, err := strconv.Atoi(vnumArg)
	if err != nil || vnum <= 0 {
		ch.Send("Vnum must be a positive number.\n\r")
		return
	}

	progs, label, ok := lookupProgs(vnum, kind)
	if !ok {
		ch.Sendf("%s %d does not exist.\n\r", strings.Title(kind), vnum)
		return
	}

	rest = strings.TrimSpace(rest)
	if rest == "" {
		if len(progs) == 0 {
			ch.Sendf("%s %d has no %s-progs.\n\r", label, vnum, kind)
			return
		}
		var sb strings.Builder
		fmt.Fprintf(&sb, "%s %d has %d %s-prog(s):\n\r", label, vnum, len(progs), kind)
		for i, p := range progs {
			trigName := firstTriggerName(p.Type)
			fmt.Fprintf(&sb, "  [%d] trigger=%s args=%q\n\r", i, trigName, p.ArgList)
		}
		ch.Send(sb.String())
		return
	}

	bit, ok := parseProgTriggerName(rest)
	if !ok {
		ch.Sendf("Unknown trigger '%s'.\n\r", rest)
		return
	}
	matches := findProgsByTrigger(progs, bit)
	if len(matches) == 0 {
		ch.Sendf("No %s-prog on %s %d with trigger %s.\n\r", kind, label, vnum, rest)
		return
	}
	for _, p := range matches {
		showProg(ch, p)
	}
	ch.Send("(read-only inspector; full string-editor integration is a follow-up)\n\r")
}

func progEditCmdPrefix(kind string) string {
	switch kind {
	case "mob":
		return "mp"
	case "obj":
		return "op"
	case "room":
		return "rp"
	}
	return ""
}

func lookupProgs(vnum int, kind string) ([]*types.MProgData, string, bool) {
	switch kind {
	case "mob":
		idx, ok := WorldRef.MobIndex[vnum]
		if !ok {
			return nil, "Mob", false
		}
		return idx.MudProgs, "Mob", true
	case "obj":
		idx, ok := WorldRef.ObjIndex[vnum]
		if !ok {
			return nil, "Object", false
		}
		return idx.MudProgs, "Object", true
	case "room":
		r := WorldRef.GetRoom(vnum)
		if r == nil {
			return nil, "Room", false
		}
		return r.MudProgs, "Room", true
	}
	return nil, "", false
}

// firstTriggerName returns a readable name for the first bit set in mask.
// Delegates to util.FirstMProgFlagName (single source of truth) and falls
// back to a hex rendering when no known bit is set.
func firstTriggerName(mask int64) string {
	if name := util.FirstMProgFlagName(mask); name != "" {
		return name
	}
	return fmt.Sprintf("0x%x", mask)
}
