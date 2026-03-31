package act

import (
	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// DoKill implements the 'kill' command.
func DoKill(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Kill whom?\n\r")
		return
	}

	victim := handler.GetCharRoom(ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}

	if victim == ch {
		ch.Send("You hit yourself. Ouch!\n\r")
		return
	}

	if !victim.IsNPC() {
		ch.Send("You must MURDER a player.\n\r")
		return
	}

	if ch.Fighting != nil {
		ch.Send("You do the best you can!\n\r")
		return
	}

	combat.StartFighting(ch, victim)
	ch.Sendf("You attack %s!\n\r", victim.ShortDescr)
}

// DoFlee implements the 'flee' command.
func DoFlee(ch *types.CharData, argument string) {
	if ch.Fighting == nil {
		ch.Send("You aren't fighting anyone.\n\r")
		return
	}

	if ch.Move <= 0 {
		ch.Send("You're too exhausted to flee!\n\r")
		return
	}

	wasInRoom := ch.InRoom

	for attempt := 0; attempt < 8; attempt++ {
		door := util.NumberRange(0, 9)
		exit := ch.InRoom.GetExit(door)
		if exit == nil || exit.ToRoom == nil {
			continue
		}
		if exit.ExitInfo&int(types.EX_CLOSED) != 0 {
			continue
		}

		MoveChar(ch, door)

		if ch.InRoom != wasInRoom {
			// Fled successfully
			combat.StopFighting(ch, true)
			ch.Send("You flee from combat!\n\r")
			return
		}
	}

	ch.Send("You attempt to flee but can't escape!\n\r")
}
