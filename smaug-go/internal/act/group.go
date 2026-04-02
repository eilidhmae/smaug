package act

import (
	"fmt"
	"strings"

	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// DoFollow implements the 'follow' command.
func DoFollow(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Follow whom?\n\r")
		return
	}

	victim := handler.GetCharRoom(ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}

	if victim == ch {
		// Follow self means stop following
		if ch.Master == nil {
			ch.Send("You aren't following anyone.\n\r")
			return
		}
		ch.Master = nil
		ch.Leader = nil
		ch.Send("You stop following.\n\r")
		return
	}

	if ch.Master == victim {
		ch.Sendf("You are already following %s.\n\r", victim.Name)
		return
	}

	// Stop following old master first
	if ch.Master != nil {
		ch.Master = nil
		ch.Leader = nil
	}

	ch.Master = victim
	ch.Leader = victim
	ch.Sendf("You now follow %s.\n\r", victim.Name)
	victim.Sendf("%s now follows you.\n\r", ch.Name)
}

// DoGroup implements the 'group' command.
func DoGroup(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)

	if arg == "" {
		// Show group info
		leader := ch.Leader
		if leader == nil {
			leader = ch
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "&W%s's group:&D\n\r", leader.Name)
		fmt.Fprintf(&sb, "  %-15s  %5d/%5d hp  %5d/%5d mana  %5d/%5d move\n\r",
			leader.Name, leader.Hit, leader.MaxHit,
			leader.Mana, leader.MaxMana, leader.Move, leader.MaxMove)

		if ch.InRoom != nil {
			for _, rch := range ch.InRoom.People {
				if rch != leader && rch.Leader == ch {
					fmt.Fprintf(&sb, "  %-15s  %5d/%5d hp  %5d/%5d mana  %5d/%5d move\n\r",
						rch.Name, rch.Hit, rch.MaxHit,
						rch.Mana, rch.MaxMana, rch.Move, rch.MaxMove)
				}
			}
		}

		ch.Send(sb.String())
		return
	}

	victim := handler.GetCharRoom(ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}

	if victim == ch {
		ch.Send("You can't group yourself.\n\r")
		return
	}

	if victim.Master != ch {
		ch.Sendf("%s is not following you.\n\r", victim.Name)
		return
	}

	// Toggle: if already in group, remove
	if victim.Leader == ch {
		victim.Leader = nil
		ch.Sendf("%s is removed from your group.\n\r", victim.Name)
		victim.Send("You have been removed from the group.\n\r")
		return
	}

	victim.Leader = ch
	ch.Sendf("%s joins your group.\n\r", victim.Name)
	victim.Send("You join the group.\n\r")
}

// DoOrder implements the 'order' command.
func DoOrder(ch *types.CharData, argument string) {
	arg1, rest := util.OneArgument(argument)
	if arg1 == "" {
		ch.Send("Order whom to do what?\n\r")
		return
	}
	if rest == "" {
		ch.Send("Order them to do what?\n\r")
		return
	}

	if strings.EqualFold(arg1, "all") {
		// Order all followers in room
		found := false
		if ch.InRoom != nil {
			for _, rch := range ch.InRoom.People {
				if rch != ch && rch.Master == ch && rch.Desc != nil {
					rch.Desc.InputQueue <- rest
					found = true
				}
			}
		}
		if found {
			ch.Send("Ok.\n\r")
		} else {
			ch.Send("You have no followers here.\n\r")
		}
		return
	}

	victim := handler.GetCharRoom(ch, arg1)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}

	if victim == ch {
		ch.Send("You can't order yourself.\n\r")
		return
	}

	if victim.Master != ch {
		ch.Sendf("%s is not following you.\n\r", victim.Name)
		return
	}

	if victim.Desc != nil {
		victim.Desc.InputQueue <- rest
		ch.Send("Ok.\n\r")
	} else {
		ch.Send("They can't receive orders right now.\n\r")
	}
}

// DoAssist implements the 'assist' command.
func DoAssist(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Assist whom?\n\r")
		return
	}

	victim := handler.GetCharRoom(ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}

	if victim == ch {
		ch.Send("You can't assist yourself.\n\r")
		return
	}

	if victim.Fighting == nil {
		ch.Sendf("%s isn't fighting anyone.\n\r", victim.Name)
		return
	}

	if ch.Fighting != nil {
		ch.Send("You're already fighting!\n\r")
		return
	}

	ch.Sendf("You assist %s!\n\r", victim.Name)
	combat.StartFighting(ch, victim.Fighting.Who)
}
