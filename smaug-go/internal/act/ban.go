package act

import (
	"fmt"
	"strings"
	"time"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// DoBan handles the "ban" command for adding, listing, and removing site bans.
func DoBan(ch *types.CharData, argument string) {
	arg1, rest := util.OneArgument(argument)

	switch strings.ToLower(arg1) {
	case "", "list":
		if len(WorldRef.Bans) == 0 {
			ch.Send("No bans.\r\n")
			return
		}
		var buf strings.Builder
		buf.WriteString("Banned sites:\r\n")
		for _, ban := range WorldRef.Bans {
			display := ban.Name
			if ban.Prefix {
				display = display + "*"
			}
			if ban.Suffix {
				display = "*" + display
			}
			buf.WriteString(fmt.Sprintf("  %-30s  (by %s on %s)\r\n", display, ban.BanBy, ban.BanTime))
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

	case "remove":
		site, _ := util.OneArgument(rest)
		if site == "" {
			ch.Send("Usage: ban remove <address>\r\n")
			return
		}

		siteLower := strings.ToLower(site)
		for i, ban := range WorldRef.Bans {
			if strings.EqualFold(ban.Name, siteLower) {
				WorldRef.Bans = append(WorldRef.Bans[:i], WorldRef.Bans[i+1:]...)
				ch.Send(fmt.Sprintf("Ban on %s removed.\r\n", site))
				return
			}
		}
		ch.Send("That ban was not found.\r\n")

	default:
		ch.Send("Usage: ban [site <address> | remove <address> | list]\r\n")
	}
}

// CheckBans checks whether a site matches any ban in the world's ban list.
// Returns the matching BanData or nil if no match.
func CheckBans(w *world.World, site string) *types.BanData {
	siteLower := strings.ToLower(site)
	for _, ban := range w.Bans {
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
