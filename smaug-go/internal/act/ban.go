package act

import (
	"fmt"
	"strings"
	"time"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// DoBan handles the "ban" command for adding, listing, and removing bans.
// Supports site, class, and race bans.
func DoBan(ch *types.CharData, argument string) {
	arg1, rest := util.OneArgument(argument)

	switch strings.ToLower(arg1) {
	case "", "list":
		if len(WorldRef.Bans) == 0 {
			ch.Send("No bans.\r\n")
			return
		}
		var buf strings.Builder
		buf.WriteString("Current bans:\r\n")
		for _, ban := range WorldRef.Bans {
			display := ban.Name
			if ban.Prefix {
				display = display + "*"
			}
			if ban.Suffix {
				display = "*" + display
			}
			buf.WriteString(fmt.Sprintf("  %-8s %-30s  (by %s on %s)\r\n",
				banTypeName(ban.Type), display, ban.BanBy, ban.BanTime))
		}
		ch.Send(buf.String())

	case "site":
		pattern, _ := util.OneArgument(rest)
		if pattern == "" {
			ch.Send("Usage: ban site <address>\r\n")
			return
		}

		ban := &types.BanData{
			BanBy:   ch.Name,
			BanTime: time.Now().Format("2006-01-02"),
			Type:    types.BAN_SITE,
		}

		if strings.HasPrefix(pattern, "*") {
			ban.Suffix = true
			ban.Name = pattern[1:]
		} else if strings.HasSuffix(pattern, "*") {
			ban.Prefix = true
			ban.Name = pattern[:len(pattern)-1]
		} else {
			ban.Name = pattern
		}

		WorldRef.Bans = append(WorldRef.Bans, ban)
		ch.Send(fmt.Sprintf("Ban on site %s added.\r\n", pattern))

	case "class":
		name, lvlArg := util.OneArgument(rest)
		if name == "" {
			ch.Send("Usage: ban class <classname> [level]\r\n")
			return
		}
		level := 0
		if lvlArg != "" {
			fmt.Sscanf(lvlArg, "%d", &level)
		}
		ban := &types.BanData{
			BanBy:   ch.Name,
			BanTime: time.Now().Format("2006-01-02"),
			Type:    types.BAN_CLASS,
			Name:    strings.ToLower(name),
			Level:   level,
		}
		WorldRef.Bans = append(WorldRef.Bans, ban)
		ch.Send(fmt.Sprintf("Class %s banned below level %d.\r\n", name, level))

	case "race":
		name, lvlArg := util.OneArgument(rest)
		if name == "" {
			ch.Send("Usage: ban race <racename> [level]\r\n")
			return
		}
		level := 0
		if lvlArg != "" {
			fmt.Sscanf(lvlArg, "%d", &level)
		}
		ban := &types.BanData{
			BanBy:   ch.Name,
			BanTime: time.Now().Format("2006-01-02"),
			Type:    types.BAN_RACE,
			Name:    strings.ToLower(name),
			Level:   level,
		}
		WorldRef.Bans = append(WorldRef.Bans, ban)
		ch.Send(fmt.Sprintf("Race %s banned below level %d.\r\n", name, level))

	case "remove":
		target, _ := util.OneArgument(rest)
		if target == "" {
			ch.Send("Usage: ban remove <name>\r\n")
			return
		}

		for i, ban := range WorldRef.Bans {
			if strings.EqualFold(ban.Name, target) {
				WorldRef.Bans = append(WorldRef.Bans[:i], WorldRef.Bans[i+1:]...)
				ch.Send(fmt.Sprintf("Ban on %s removed.\r\n", target))
				return
			}
		}
		ch.Send("That ban was not found.\r\n")

	default:
		ch.Send("Usage: ban [site|class|race <name> | remove <name> | list]\r\n")
	}
}

func banTypeName(t int) string {
	switch t {
	case types.BAN_CLASS:
		return "class"
	case types.BAN_RACE:
		return "race"
	default:
		return "site"
	}
}

// CheckClassBan returns the ban record if a PC of the given class name is
// banned at their current level, else nil. Matches C `src/ban.c:1266`:
// characters with level strictly greater than ban.Level bypass the ban;
// characters at or below the threshold are blocked. Ban.Level == 0 is a
// global block at all levels.
func CheckClassBan(w *world.World, className string, level int) *types.BanData {
	name := strings.ToLower(className)
	for _, ban := range w.Bans {
		if ban.Type == types.BAN_CLASS && strings.EqualFold(ban.Name, name) {
			if ban.Level == 0 || level <= ban.Level {
				return ban
			}
		}
	}
	return nil
}

// CheckRaceBan mirrors CheckClassBan for race bans.
func CheckRaceBan(w *world.World, raceName string, level int) *types.BanData {
	name := strings.ToLower(raceName)
	for _, ban := range w.Bans {
		if ban.Type == types.BAN_RACE && strings.EqualFold(ban.Name, name) {
			if ban.Level == 0 || level <= ban.Level {
				return ban
			}
		}
	}
	return nil
}

// CheckBans checks whether a site matches any SITE ban in the world's ban list.
// Returns the matching BanData or nil if no match.
func CheckBans(w *world.World, site string) *types.BanData {
	siteLower := strings.ToLower(site)
	for _, ban := range w.Bans {
		if ban.Type != 0 && ban.Type != types.BAN_SITE {
			continue
		}
		nameLower := strings.ToLower(ban.Name)
		if ban.Prefix && strings.HasPrefix(siteLower, nameLower) {
			return ban
		}
		if ban.Suffix && strings.HasSuffix(siteLower, nameLower) {
			return ban
		}
		if !ban.Prefix && !ban.Suffix && siteLower == nameLower {
			return ban
		}
	}
	return nil
}
