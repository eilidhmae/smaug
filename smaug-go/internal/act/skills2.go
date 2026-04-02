package act

import (
	"fmt"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// padName pads a name to the given width with spaces.
func padName(name string, width int) string {
	if len(name) >= width {
		return name
	}
	return name + strings.Repeat(" ", width-len(name))
}

// DoSkills lists the player's known non-spell skills with proficiency.
func DoSkills(ch *types.CharData, argument string) {
	if ch.IsNPC() || ch.PCData == nil {
		return
	}

	var buf strings.Builder
	found := false
	for i, sk := range WorldRef.Skills {
		if sk == nil || sk.Type != types.SKILL_SKILL {
			continue
		}
		if ch.PCData.Learned[i] <= 0 {
			continue
		}
		if !found {
			buf.WriteString("Your skills:\n\r")
			found = true
		}
		fmt.Fprintf(&buf, "  %s%3d%%\n\r", padName(sk.Name, 20), ch.PCData.Learned[i])
	}

	if !found {
		ch.Send("You have no skills learned.\n\r")
		return
	}
	ch.Send(buf.String())
}

// DoSpells lists the player's known spells with proficiency.
func DoSpells(ch *types.CharData, argument string) {
	if ch.IsNPC() || ch.PCData == nil {
		return
	}

	var buf strings.Builder
	found := false
	for i, sk := range WorldRef.Skills {
		if sk == nil || sk.Type != types.SKILL_SPELL {
			continue
		}
		if ch.PCData.Learned[i] <= 0 {
			continue
		}
		if !found {
			buf.WriteString("Your spells:\n\r")
			found = true
		}
		fmt.Fprintf(&buf, "  %s%3d%%\n\r", padName(sk.Name, 20), ch.PCData.Learned[i])
	}

	if !found {
		ch.Send("You have no spells learned.\n\r")
		return
	}
	ch.Send(buf.String())
}

// DoPractice implements the 'practice' command.
// With no argument, lists all known skills/spells with proficiency.
// With an argument, spends a practice session to improve the named skill.
func DoPractice(ch *types.CharData, argument string) {
	if ch.IsNPC() || ch.PCData == nil {
		return
	}

	arg, _ := util.OneArgument(argument)

	if arg == "" {
		var buf strings.Builder
		found := false
		for i, sk := range WorldRef.Skills {
			if sk == nil || ch.PCData.Learned[i] <= 0 {
				continue
			}
			if !found {
				buf.WriteString("Skills/Spells:\n\r")
				found = true
			}
			fmt.Fprintf(&buf, "  %s%3d%%\n\r", padName(sk.Name, 20), ch.PCData.Learned[i])
		}
		if !found {
			ch.Send("You have no skills or spells.\n\r")
		} else {
			fmt.Fprintf(&buf, "\nYou have %d practice sessions remaining.\n\r", ch.Practice)
			ch.Send(buf.String())
		}
		return
	}

	// Find the skill by name
	sn := -1
	arg = strings.ToLower(arg)
	for i, sk := range WorldRef.Skills {
		if sk != nil && strings.HasPrefix(strings.ToLower(sk.Name), arg) {
			if ch.PCData.Learned[i] > 0 {
				sn = i
				break
			}
		}
	}

	if sn < 0 {
		ch.Send("No such skill or spell.\n\r")
		return
	}

	if ch.Practice <= 0 {
		ch.Send("You have no more practice sessions.\n\r")
		return
	}

	sk := WorldRef.Skills[sn]
	adept := 75
	if sk.SkillAdept[ch.Class] > 0 {
		adept = sk.SkillAdept[ch.Class]
	}

	if ch.PCData.Learned[sn] >= adept {
		ch.Sendf("You are already adept at %s.\n\r", sk.Name)
		return
	}

	ch.Practice--
	improve := 5 + util.NumberRange(1, 10)
	ch.PCData.Learned[sn] += improve
	if ch.PCData.Learned[sn] > adept {
		ch.PCData.Learned[sn] = adept
	}

	ch.Sendf("You practice %s. (%d%%)\n\r", sk.Name, ch.PCData.Learned[sn])
}
