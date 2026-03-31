package act

import (
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// DoConsider implements the 'consider' command: compare your level to a mob.
func DoConsider(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Consider killing whom?\n\r")
		return
	}

	victim := handler.GetCharRoom(ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}

	if victim == ch {
		ch.Send("You think about yourself for a moment.\n\r")
		return
	}

	diff := victim.Level - ch.Level
	switch {
	case diff <= -10:
		ch.Send("Not worth the effort.\n\r")
	case diff <= -5:
		ch.Send("Should be no contest.\n\r")
	case diff <= -2:
		ch.Send("Easy.\n\r")
	case diff <= 1:
		ch.Send("A fair fight.\n\r")
	case diff <= 4:
		ch.Send("You would need some luck.\n\r")
	case diff <= 9:
		ch.Send("You would need a lot of luck!\n\r")
	default:
		ch.Send("Do you have a death wish?\n\r")
	}
}

// DoWhere implements the 'where' command: find characters in the area.
func DoWhere(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)

	if arg == "" {
		// Show all visible players in area
		ch.Send("Players near you:\n\r")
		found := false
		for _, wch := range WorldRef.Characters {
			if wch.IsNPC() || wch.InRoom == nil || ch.InRoom == nil {
				continue
			}
			if wch.InRoom.Area != ch.InRoom.Area {
				continue
			}
			ch.Sendf("  %-20s %s\n\r", wch.Name, wch.InRoom.Name)
			found = true
		}
		if !found {
			ch.Send("  No one.\n\r")
		}
		return
	}

	// Search for specific character in area
	found := false
	for _, wch := range WorldRef.Characters {
		if wch.InRoom == nil || ch.InRoom == nil {
			continue
		}
		if wch.InRoom.Area != ch.InRoom.Area {
			continue
		}
		if !util.IsName(arg, wch.Name) {
			continue
		}
		ch.Sendf("  %-20s %s\n\r", wch.Name, wch.InRoom.Name)
		found = true
	}
	if !found {
		ch.Send("No one by that name around here.\n\r")
	}
}

// DoTime implements the 'time' command: display game time.
func DoTime(ch *types.CharData, argument string) {
	if WorldRef == nil {
		ch.Send("Time is meaningless.\n\r")
		return
	}

	hour := WorldRef.TimeInfo.Hour
	day := WorldRef.TimeInfo.Day + 1
	month := WorldRef.TimeInfo.Month + 1
	year := WorldRef.TimeInfo.Year

	var timeOfDay string
	switch {
	case hour < 5:
		timeOfDay = "late at night"
	case hour < 9:
		timeOfDay = "early morning"
	case hour < 12:
		timeOfDay = "morning"
	case hour < 14:
		timeOfDay = "midday"
	case hour < 18:
		timeOfDay = "afternoon"
	case hour < 21:
		timeOfDay = "evening"
	default:
		timeOfDay = "night"
	}

	ch.Sendf("It is hour %d of the day, %s.\n\rDay %d of month %d, year %d.\n\r",
		hour, timeOfDay, day, month, year)
}
