package act

import (
	"fmt"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// DoAset implements the 'aset' command: set fields on an area.
func DoAset(ch *types.CharData, argument string) {
	if ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}

	areaArg, rest := util.OneArgument(argument)
	if areaArg == "" {
		ch.Send("Syntax: aset <area> <field> <value>\n\r")
		return
	}

	field, value := util.OneArgument(rest)
	if field == "" {
		ch.Send("Syntax: aset <area> <field> <value>\n\rValid fields: name author resetmsg low_r_vnum hi_r_vnum low_m_vnum hi_m_vnum low_o_vnum hi_o_vnum resetfreq\n\r")
		return
	}

	area := findAreaByPrefix(areaArg)
	if area == nil {
		ch.Send("No such area.\n\r")
		return
	}

	field = strings.ToLower(field)

	switch field {
	case "name":
		area.Name = value
		ch.Sendf("Area name set to %s.\n\r", value)

	case "author":
		area.Author = value
		ch.Sendf("Area author set to %s.\n\r", value)

	case "resetmsg":
		area.ResetMsg = value
		ch.Sendf("Area resetmsg set to %s.\n\r", value)

	case "low_r_vnum":
		val := parseIntOrZero(value)
		area.LowRVnum = val
		ch.Sendf("Area low_r_vnum set to %d.\n\r", val)

	case "hi_r_vnum":
		val := parseIntOrZero(value)
		area.HiRVnum = val
		ch.Sendf("Area hi_r_vnum set to %d.\n\r", val)

	case "low_m_vnum":
		val := parseIntOrZero(value)
		area.LowMVnum = val
		ch.Sendf("Area low_m_vnum set to %d.\n\r", val)

	case "hi_m_vnum":
		val := parseIntOrZero(value)
		area.HiMVnum = val
		ch.Sendf("Area hi_m_vnum set to %d.\n\r", val)

	case "low_o_vnum":
		val := parseIntOrZero(value)
		area.LowOVnum = val
		ch.Sendf("Area low_o_vnum set to %d.\n\r", val)

	case "hi_o_vnum":
		val := parseIntOrZero(value)
		area.HiOVnum = val
		ch.Sendf("Area hi_o_vnum set to %d.\n\r", val)

	case "resetfreq":
		val := parseIntOrZero(value)
		area.ResetFrequency = val
		ch.Sendf("Area resetfreq set to %d.\n\r", val)

	default:
		ch.Send("Valid fields: name author resetmsg low_r_vnum hi_r_vnum low_m_vnum hi_m_vnum low_o_vnum hi_o_vnum resetfreq\n\r")
	}
}

// DoAstat implements the 'astat' command: display area statistics.
func DoAstat(ch *types.CharData, argument string) {
	if ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}

	var area *types.AreaData

	arg := strings.TrimSpace(argument)
	if arg == "" {
		if ch.InRoom == nil {
			ch.Send("You are not in a room.\n\r")
			return
		}
		area = ch.InRoom.Area
		if area == nil {
			ch.Send("This room is not in an area.\n\r")
			return
		}
	} else {
		area = findAreaByPrefix(arg)
		if area == nil {
			ch.Send("No such area.\n\r")
			return
		}
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Name: %s  Author: %s  Filename: %s\n\r", area.Name, area.Author, area.Filename)
	fmt.Fprintf(&sb, "Low/Hi Room: %d-%d  Low/Hi Mob: %d-%d  Low/Hi Obj: %d-%d\n\r",
		area.LowRVnum, area.HiRVnum, area.LowMVnum, area.HiMVnum, area.LowOVnum, area.HiOVnum)
	fmt.Fprintf(&sb, "Age: %d  Players: %d  Reset Freq: %d\n\r", area.Age, area.NPlayer, area.ResetFrequency)
	fmt.Fprintf(&sb, "Reset Message: %s\n\r", area.ResetMsg)
	fmt.Fprintf(&sb, "Flags: %d\n\r", area.Flags)
	ch.Send(sb.String())
}

// findAreaByPrefix searches WorldRef.Areas for an area whose name
// starts with the given prefix (case-insensitive).
func findAreaByPrefix(prefix string) *types.AreaData {
	if WorldRef == nil {
		return nil
	}
	lower := strings.ToLower(prefix)
	for _, area := range WorldRef.Areas {
		if strings.HasPrefix(strings.ToLower(area.Name), lower) {
			return area
		}
	}
	return nil
}
