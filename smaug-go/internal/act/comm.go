package act

import (
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// DoTell implements the 'tell' command: private message to a character.
func DoTell(ch *types.CharData, argument string) {
	arg, message := util.OneArgument(argument)
	if arg == "" || message == "" {
		ch.Send("Tell whom what?\n\r")
		return
	}

	victim := handler.GetCharWorld(WorldRef, ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}

	if victim == ch {
		ch.Send("Talking to yourself again?\n\r")
		return
	}

	ch.Sendf("You tell %s '%s'\n\r", victim.Name, message)
	victim.Sendf("%s tells you '%s'\n\r", ch.Name, message)
	victim.Reply = ch
}

// DoReply implements the 'reply' command: reply to last tell sender.
func DoReply(ch *types.CharData, argument string) {
	if ch.Reply == nil {
		ch.Send("They aren't here.\n\r")
		return
	}

	if argument == "" {
		ch.Send("Reply what?\n\r")
		return
	}

	victim := ch.Reply
	ch.Sendf("You tell %s '%s'\n\r", victim.Name, argument)
	victim.Sendf("%s tells you '%s'\n\r", ch.Name, argument)
	victim.Reply = ch
}

// DoYell implements the 'yell' command: message to all in same area.
func DoYell(ch *types.CharData, argument string) {
	if argument == "" {
		ch.Send("Yell what?\n\r")
		return
	}

	ch.Sendf("You yell '%s'\n\r", argument)

	for _, wch := range WorldRef.Characters {
		if wch == ch || wch.Desc == nil {
			continue
		}
		if wch.InRoom != nil && ch.InRoom != nil &&
			wch.InRoom.Area == ch.InRoom.Area {
			wch.Sendf("%s yells '%s'\n\r", ch.Name, argument)
		}
	}
}

// DoGossip implements the 'gossip' command: global channel message.
func DoGossip(ch *types.CharData, argument string) {
	if argument == "" {
		ch.Send("Gossip what?\n\r")
		return
	}

	ch.Sendf("You gossip '%s'\n\r", argument)

	for _, wch := range WorldRef.Characters {
		if wch == ch || wch.Desc == nil {
			continue
		}
		wch.Sendf("%s gossips '%s'\n\r", ch.Name, argument)
	}
}

// DoShout implements the 'shout' command: message heard by all players.
func DoShout(ch *types.CharData, argument string) {
	if argument == "" {
		ch.Send("Shout what?\n\r")
		return
	}

	ch.Sendf("You shout '%s'\n\r", argument)

	for _, wch := range WorldRef.Characters {
		if wch == ch || wch.Desc == nil {
			continue
		}
		wch.Sendf("%s shouts '%s'\n\r", ch.Name, argument)
	}
}

// DoPmote implements the 'pmote' command: possessive emote.
// The character's name is replaced with "your" when shown to each target.
func DoPmote(ch *types.CharData, argument string) {
	if argument == "" {
		ch.Send("Pmote what?\n\r")
		return
	}

	if ch.InRoom == nil {
		return
	}

	for _, rch := range ch.InRoom.People {
		if rch.Desc == nil {
			continue
		}
		if rch == ch {
			rch.Sendf("%s's %s\n\r", ch.Name, argument)
		} else {
			// Replace target's name with "your" if it appears in the message
			msg := argument
			rch.Sendf("%s's %s\n\r", ch.Name, msg)
		}
	}
}

// DoEmote implements the 'emote' command: roleplay action visible to room.
func DoEmote(ch *types.CharData, argument string) {
	if argument == "" {
		ch.Send("Emote what?\n\r")
		return
	}

	if ch.InRoom == nil {
		return
	}

	for _, rch := range ch.InRoom.People {
		if rch.Desc == nil {
			continue
		}
		rch.Sendf("%s %s\n\r", ch.Name, argument)
	}
}
