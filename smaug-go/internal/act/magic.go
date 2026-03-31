package act

import (
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/magic"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// DoCast implements the 'cast' command.
func DoCast(ch *types.CharData, argument string) {
	if ch.IsNPC() {
		return
	}

	arg1, arg2 := util.OneArgument(argument)
	if arg1 == "" {
		ch.Send("Cast which what where?\n\r")
		return
	}

	sn := magic.FindSpellByName(WorldRef, arg1)
	if sn < 0 {
		ch.Send("You don't know any spells of that name.\n\r")
		return
	}

	skill := WorldRef.Skills[sn]
	if skill == nil {
		ch.Send("Error: spell data not found.\n\r")
		return
	}

	// Mana cost
	mana := skill.MinMana
	if mana <= 0 {
		mana = 10
	}

	if ch.Mana < mana {
		ch.Send("You don't have enough mana.\n\r")
		return
	}

	// Resolve target
	var victim *types.CharData

	switch skill.Target {
	case types.TAR_CHAR_OFFENSIVE:
		if arg2 == "" {
			// Default to current fighting target
			if ch.Fighting != nil {
				victim = ch.Fighting.Who
			} else {
				ch.Send("Cast the spell on whom?\n\r")
				return
			}
		} else {
			victim = handler.GetCharRoom(ch, arg2)
			if victim == nil {
				ch.Send("They aren't here.\n\r")
				return
			}
		}

	case types.TAR_CHAR_DEFENSIVE, types.TAR_CHAR_SELF:
		if arg2 == "" {
			victim = ch
		} else {
			victim = handler.GetCharRoom(ch, arg2)
			if victim == nil {
				ch.Send("They aren't here.\n\r")
				return
			}
		}

	case types.TAR_IGNORE:
		victim = ch

	default:
		victim = ch
	}

	// Deduct mana
	ch.Mana -= mana

	// Find and call spell function
	spellFn := magic.FindSpellFunc(skill.SpellFunName)
	if spellFn == nil {
		ch.Sendf("You cast '%s' but nothing happens.\n\r", skill.Name)
		return
	}

	ch.Sendf("You cast '%s'.\n\r", skill.Name)

	// Notify room
	if ch.InRoom != nil {
		for _, rch := range ch.InRoom.People {
			if rch != ch && rch != victim && rch.Desc != nil {
				rch.Sendf("%s utters the words, '%s'.\n\r", ch.Name, skill.Name)
			}
		}
	}

	spellFn(WorldRef, sn, ch.Level, ch, victim)
}
