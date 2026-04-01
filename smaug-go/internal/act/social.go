package act

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// CheckSocial handles social commands as a fallback from the interpreter.
// Returns true if a matching social was found and executed.
func CheckSocial(ch *types.CharData, cmd string, argument string) bool {
	if WorldRef == nil {
		return false
	}

	cmd = strings.ToLower(cmd)
	var social *types.SocialType

	for _, s := range WorldRef.Socials {
		if strings.EqualFold(s.Name, cmd) || strings.HasPrefix(strings.ToLower(s.Name), cmd) {
			social = s
			break
		}
	}

	if social == nil {
		return false
	}

	if ch.InRoom == nil {
		return false
	}

	arg, _ := util.OneArgument(argument)

	if arg == "" {
		// No target
		ch.Sendf("%s\n\r", socialSub(social.CharNoArg, ch, nil))
		for _, rch := range ch.InRoom.People {
			if rch != ch && rch.Desc != nil {
				rch.Sendf("%s\n\r", socialSub(social.OthersNoArg, ch, nil))
			}
		}
		return true
	}

	victim := handler.GetCharRoom(ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return true
	}

	if victim == ch {
		// Self-targeted
		ch.Sendf("%s\n\r", socialSub(social.CharAuto, ch, ch))
		for _, rch := range ch.InRoom.People {
			if rch != ch && rch.Desc != nil {
				rch.Sendf("%s\n\r", socialSub(social.OthersAuto, ch, ch))
			}
		}
		return true
	}

	// Target found
	ch.Sendf("%s\n\r", socialSub(social.CharFound, ch, victim))
	victim.Sendf("%s\n\r", socialSub(social.VictFound, ch, victim))
	for _, rch := range ch.InRoom.People {
		if rch != ch && rch != victim && rch.Desc != nil {
			rch.Sendf("%s\n\r", socialSub(social.OthersFound, ch, victim))
		}
	}
	return true
}

// socialSub does basic variable substitution for social messages.
// $n = actor name, $N = target name, $e/$E = he/she, $m/$M = him/her, $s/$S = his/her
func socialSub(msg string, ch *types.CharData, victim *types.CharData) string {
	if msg == "" {
		return ""
	}

	r := strings.NewReplacer(
		"$n", ch.Name,
		"$e", pronoun(ch, "he", "she", "it"),
		"$m", pronoun(ch, "him", "her", "it"),
		"$s", pronoun(ch, "his", "her", "its"),
	)
	result := r.Replace(msg)

	if victim != nil {
		r2 := strings.NewReplacer(
			"$N", victim.Name,
			"$E", pronoun(victim, "he", "she", "it"),
			"$M", pronoun(victim, "him", "her", "it"),
			"$S", pronoun(victim, "his", "her", "its"),
		)
		result = r2.Replace(result)
	}

	return result
}

func pronoun(ch *types.CharData, male, female, neutral string) string {
	switch ch.Sex {
	case types.SEX_MALE:
		return male
	case types.SEX_FEMALE:
		return female
	default:
		return neutral
	}
}
