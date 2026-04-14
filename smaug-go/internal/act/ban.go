package act

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// DoBan handles the "ban" command for adding, listing, and removing bans.
// Subcommands: list, site <addr>, class <name> [lvl], race <name> [lvl],
// remove <name>.
func DoBan(ch *types.CharData, argument string) {
	arg1, rest := util.OneArgument(argument)

	switch strings.ToLower(arg1) {
	case "", "list":
		if len(WorldRef.Bans) == 0 {
			ch.Send("No bans.\r\n")
			return
		}
		var buf strings.Builder
		buf.WriteString("Bans:\r\n")
		for _, ban := range WorldRef.Bans {
			display := ban.Name
			if ban.Prefix {
				display = display + "*"
			}
			if ban.Suffix {
				display = "*" + display
			}
			kind := banKindLabel(ban)
			buf.WriteString(fmt.Sprintf("  [%-5s] %-30s  (by %s on %s)\r\n",
				kind, display, ban.BanBy, ban.BanTime))
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
		ch.Send(fmt.Sprintf("Ban on %s added.\r\n", pattern))

	case "class":
		name, rest2 := util.OneArgument(rest)
		if name == "" {
			ch.Send("Usage: ban class <name> [level]\r\n")
			return
		}
		level := 0
		if rest2 != "" {
			if v, err := strconv.Atoi(strings.TrimSpace(rest2)); err == nil {
				level = v
			}
		}
		ban := &types.BanData{
			Name:    name,
			BanBy:   ch.Name,
			BanTime: time.Now().Format("2006-01-02"),
			Type:    types.BAN_CLASS,
			Level:   level,
		}
		WorldRef.Bans = append(WorldRef.Bans, ban)
		ch.Send(fmt.Sprintf("Class ban on %s added.\r\n", name))

	case "race":
		name, rest2 := util.OneArgument(rest)
		if name == "" {
			ch.Send("Usage: ban race <name> [level]\r\n")
			return
		}
		level := 0
		if rest2 != "" {
			if v, err := strconv.Atoi(strings.TrimSpace(rest2)); err == nil {
				level = v
			}
		}
		ban := &types.BanData{
			Name:    name,
			BanBy:   ch.Name,
			BanTime: time.Now().Format("2006-01-02"),
			Type:    types.BAN_RACE,
			Level:   level,
		}
		WorldRef.Bans = append(WorldRef.Bans, ban)
		ch.Send(fmt.Sprintf("Race ban on %s added.\r\n", name))

	case "remove":
		name, _ := util.OneArgument(rest)
		if name == "" {
			ch.Send("Usage: ban remove <name>\r\n")
			return
		}

		nameLower := strings.ToLower(name)
		for i, ban := range WorldRef.Bans {
			if strings.EqualFold(ban.Name, nameLower) {
				WorldRef.Bans = append(WorldRef.Bans[:i], WorldRef.Bans[i+1:]...)
				ch.Send(fmt.Sprintf("Ban on %s removed.\r\n", name))
				return
			}
		}
		ch.Send("That ban was not found.\r\n")

	default:
		ch.Send("Usage: ban [site <addr>|class <name> [lvl]|race <name> [lvl]|remove <name>|list]\r\n")
	}
}

func banKindLabel(ban *types.BanData) string {
	switch ban.BanType() {
	case types.BAN_CLASS:
		return "class"
	case types.BAN_RACE:
		return "race"
	default:
		return "site"
	}
}

// IsClassBanned reports whether the named class is currently banned for a
// character of the given level. A ban's Level field is the minimum level a
// character must have to bypass it — mortals below that level are refused.
func IsClassBanned(w *world.World, className string, level int) *types.BanData {
	if w == nil {
		return nil
	}
	for _, ban := range w.Bans {
		if ban.BanType() != types.BAN_CLASS {
			continue
		}
		if strings.EqualFold(ban.Name, className) && level < ban.Level {
			return ban
		}
	}
	return nil
}

// IsRaceBanned mirrors IsClassBanned for BAN_RACE entries.
func IsRaceBanned(w *world.World, raceName string, level int) *types.BanData {
	if w == nil {
		return nil
	}
	for _, ban := range w.Bans {
		if ban.BanType() != types.BAN_RACE {
			continue
		}
		if strings.EqualFold(ban.Name, raceName) && level < ban.Level {
			return ban
		}
	}
	return nil
}

// CheckBans checks whether a site matches any site-ban in the world's ban list.
// Returns the matching BanData or nil if no match. Class and race bans are
// ignored here — use IsClassBanned / IsRaceBanned for those.
func CheckBans(w *world.World, site string) *types.BanData {
	siteLower := strings.ToLower(site)
	for _, ban := range w.Bans {
		if ban.BanType() != types.BAN_SITE {
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
