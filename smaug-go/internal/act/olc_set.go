package act

import (
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// DoMset implements the 'mset' command: set fields on a mobile/character.
func DoMset(ch *types.CharData, argument string) {
	if ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}

	target, rest := util.OneArgument(argument)
	if target == "" {
		ch.Send("Syntax: mset <target> <field> <value>\n\r")
		return
	}

	field, value := util.OneArgument(rest)
	if field == "" {
		ch.Send("Syntax: mset <target> <field> <value>\n\r")
		return
	}

	victim := handler.GetCharRoom(ch, target)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}

	field = strings.ToLower(field)

	switch field {
	case "level":
		val := parseIntOrZero(value)
		victim.Level = val
		if val > 0 {
			victim.Trust = val
		}
		ch.Sendf("%s's level set to %d.\n\r", victim.Name, val)

	case "str":
		val := clamp(parseIntOrZero(value), 1, 25)
		victim.PermStr = val
		ch.Sendf("%s's str set to %d.\n\r", victim.Name, val)

	case "int":
		val := clamp(parseIntOrZero(value), 1, 25)
		victim.PermInt = val
		ch.Sendf("%s's int set to %d.\n\r", victim.Name, val)

	case "wis":
		val := clamp(parseIntOrZero(value), 1, 25)
		victim.PermWis = val
		ch.Sendf("%s's wis set to %d.\n\r", victim.Name, val)

	case "dex":
		val := clamp(parseIntOrZero(value), 1, 25)
		victim.PermDex = val
		ch.Sendf("%s's dex set to %d.\n\r", victim.Name, val)

	case "con":
		val := clamp(parseIntOrZero(value), 1, 25)
		victim.PermCon = val
		ch.Sendf("%s's con set to %d.\n\r", victim.Name, val)

	case "cha":
		val := clamp(parseIntOrZero(value), 1, 25)
		victim.PermCha = val
		ch.Sendf("%s's cha set to %d.\n\r", victim.Name, val)

	case "lck":
		val := clamp(parseIntOrZero(value), 1, 25)
		victim.PermLck = val
		ch.Sendf("%s's lck set to %d.\n\r", victim.Name, val)

	case "hp":
		val := parseIntOrZero(value)
		victim.MaxHit = val
		victim.Hit = val
		ch.Sendf("%s's hp set to %d.\n\r", victim.Name, val)

	case "mana":
		val := parseIntOrZero(value)
		victim.MaxMana = val
		victim.Mana = val
		ch.Sendf("%s's mana set to %d.\n\r", victim.Name, val)

	case "move":
		val := parseIntOrZero(value)
		victim.MaxMove = val
		victim.Move = val
		ch.Sendf("%s's move set to %d.\n\r", victim.Name, val)

	case "hitroll":
		val := parseIntOrZero(value)
		victim.Hitroll = val
		ch.Sendf("%s's hitroll set to %d.\n\r", victim.Name, val)

	case "damroll":
		val := parseIntOrZero(value)
		victim.Damroll = val
		ch.Sendf("%s's damroll set to %d.\n\r", victim.Name, val)

	case "gold":
		val := parseIntOrZero(value)
		victim.Gold = val
		ch.Sendf("%s's gold set to %d.\n\r", victim.Name, val)

	case "align":
		val := clamp(parseIntOrZero(value), -1000, 1000)
		victim.Alignment = val
		ch.Sendf("%s's align set to %d.\n\r", victim.Name, val)

	case "name":
		victim.Name = value
		ch.Sendf("%s's name set to %s.\n\r", victim.Name, value)

	case "short":
		victim.ShortDescr = value
		ch.Sendf("%s's short set to %s.\n\r", victim.Name, value)

	case "long":
		victim.LongDescr = value
		ch.Sendf("%s's long set to %s.\n\r", victim.Name, value)

	case "sex":
		switch strings.ToLower(value) {
		case "male":
			victim.Sex = 1
		case "female":
			victim.Sex = 2
		case "neutral":
			victim.Sex = 0
		default:
			ch.Send("Sex must be 'male', 'female', or 'neutral'.\n\r")
			return
		}
		ch.Sendf("%s's sex set to %s.\n\r", victim.Name, value)

	case "race":
		val := parseIntOrZero(value)
		victim.Race = val
		ch.Sendf("%s's race set to %d.\n\r", victim.Name, val)

	case "class":
		val := parseIntOrZero(value)
		victim.Class = val
		ch.Sendf("%s's class set to %d.\n\r", victim.Name, val)

	default:
		ch.Send("Valid fields: level str int wis dex con cha lck hp mana move hitroll damroll gold align name short long sex race class\n\r")
	}
}

// DoOset implements the 'oset' command: set fields on an object.
func DoOset(ch *types.CharData, argument string) {
	if ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}

	target, rest := util.OneArgument(argument)
	if target == "" {
		ch.Send("Syntax: oset <target> <field> <value>\n\r")
		return
	}

	field, value := util.OneArgument(rest)
	if field == "" {
		ch.Send("Syntax: oset <target> <field> <value>\n\r")
		return
	}

	obj := handler.GetObjHere(ch, target)
	if obj == nil {
		ch.Send("Nothing like that here.\n\r")
		return
	}

	field = strings.ToLower(field)

	switch field {
	case "type":
		val := parseIntOrZero(value)
		obj.ItemType = val
		ch.Sendf("%s's type set to %d.\n\r", obj.ShortDescr, val)

	case "name":
		obj.Name = value
		ch.Sendf("%s's name set to %s.\n\r", obj.ShortDescr, value)

	case "short":
		obj.ShortDescr = value
		ch.Sendf("%s's short set to %s.\n\r", obj.ShortDescr, value)

	case "long":
		obj.Description = value
		ch.Sendf("%s's long set to %s.\n\r", obj.ShortDescr, value)

	case "value0":
		val := parseIntOrZero(value)
		obj.Value[0] = val
		ch.Sendf("%s's value0 set to %d.\n\r", obj.ShortDescr, val)

	case "value1":
		val := parseIntOrZero(value)
		obj.Value[1] = val
		ch.Sendf("%s's value1 set to %d.\n\r", obj.ShortDescr, val)

	case "value2":
		val := parseIntOrZero(value)
		obj.Value[2] = val
		ch.Sendf("%s's value2 set to %d.\n\r", obj.ShortDescr, val)

	case "value3":
		val := parseIntOrZero(value)
		obj.Value[3] = val
		ch.Sendf("%s's value3 set to %d.\n\r", obj.ShortDescr, val)

	case "value4":
		val := parseIntOrZero(value)
		obj.Value[4] = val
		ch.Sendf("%s's value4 set to %d.\n\r", obj.ShortDescr, val)

	case "value5":
		val := parseIntOrZero(value)
		obj.Value[5] = val
		ch.Sendf("%s's value5 set to %d.\n\r", obj.ShortDescr, val)

	case "weight":
		val := parseIntOrZero(value)
		obj.Weight = val
		ch.Sendf("%s's weight set to %d.\n\r", obj.ShortDescr, val)

	case "cost":
		val := parseIntOrZero(value)
		obj.GoldCost = val
		ch.Sendf("%s's cost set to %d.\n\r", obj.ShortDescr, val)

	case "level":
		val := parseIntOrZero(value)
		obj.Level = val
		ch.Sendf("%s's level set to %d.\n\r", obj.ShortDescr, val)

	case "flags":
		val := parseIntOrZero(value)
		obj.ExtraFlags.Toggle(val)
		ch.Sendf("%s's flag %d toggled.\n\r", obj.ShortDescr, val)

	case "wearflags":
		val := parseIntOrZero(value)
		obj.WearFlags = val
		ch.Sendf("%s's wearflags set to %d.\n\r", obj.ShortDescr, val)

	default:
		ch.Send("Valid fields: type name short long value0 value1 value2 value3 value4 value5 weight cost level flags wearflags\n\r")
	}
}

// DoRset implements the 'rset' command: set fields on the current room.
func DoRset(ch *types.CharData, argument string) {
	if ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}

	if ch.InRoom == nil {
		ch.Send("You are not in a room.\n\r")
		return
	}

	field, value := util.OneArgument(argument)
	if field == "" {
		ch.Send("Syntax: rset <field> <value>\n\rValid fields: flags sector\n\r")
		return
	}

	field = strings.ToLower(field)

	switch field {
	case "flags":
		val := parseIntOrZero(value)
		ch.InRoom.RoomFlags.Toggle(val)
		ch.Sendf("Room flag %d toggled.\n\r", val)

	case "sector":
		val := parseIntOrZero(value)
		ch.InRoom.SectorType = val
		ch.Sendf("Room sector set to %d.\n\r", val)

	default:
		ch.Send("Valid fields: flags sector\n\r")
	}
}

// parseIntOrZero parses a string to int, returning 0 on error.
func parseIntOrZero(s string) int {
	val, _ := strconv.Atoi(s)
	return val
}

// clamp constrains val to the range [min, max].
func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
