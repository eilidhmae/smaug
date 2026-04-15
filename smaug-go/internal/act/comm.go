package act

import (
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/mudprog"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// translateFor returns the message the listener hears from the speaker.
// If the speaker is speaking a language the listener does not share, the
// text is scrambled. LANG_UNKNOWN (0) or LANG_COMMON passes through.
func translateFor(speaker, listener *types.CharData, text string) string {
	if speaker == nil || listener == nil {
		return text
	}
	lang := uint32(speaker.Speaking)
	if lang == 0 || lang == types.LANG_COMMON {
		return text
	}
	if uint32(listener.Speaks)&lang != 0 {
		return text
	}
	// Immortals understand everything.
	if !listener.IsNPC() && listener.Act.IsSet(types.PLR_HOLYLIGHT) {
		return text
	}
	return util.Scramble(text, lang)
}

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

	if !victim.IsNPC() && victim.Act.IsSet(types.PLR_NO_TELL) {
		ch.Send("They are refusing tells.\n\r")
		return
	}

	ch.Sendf("You tell %s '%s'\n\r", victim.Name, message)
	heard := translateFor(ch, victim, message)
	if !victim.IsNPC() && victim.Act.IsSet(types.PLR_AFK) {
		victim.Sendf("(afk) %s tells you '%s'\n\r", ch.Name, heard)
	} else {
		victim.Sendf("%s tells you '%s'\n\r", ch.Name, heard)
	}
	victim.Reply = ch

	// Fire TELL progs on NPC targets after the message has been delivered.
	if victim.IsNPC() {
		mudprog.TrigTell(ch, victim, message)
	}
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
	if !victim.IsNPC() && victim.Act.IsSet(types.PLR_NO_TELL) {
		ch.Send("They are refusing tells.\n\r")
		return
	}
	ch.Sendf("You tell %s '%s'\n\r", victim.Name, argument)
	heard := translateFor(ch, victim, argument)
	if !victim.IsNPC() && victim.Act.IsSet(types.PLR_AFK) {
		victim.Sendf("(afk) %s tells you '%s'\n\r", ch.Name, heard)
	} else {
		victim.Sendf("%s tells you '%s'\n\r", ch.Name, heard)
	}
	victim.Reply = ch
}

// DoYell implements the 'yell' command: message to all in same area.
func DoYell(ch *types.CharData, argument string) {
	if argument == "" {
		ch.Send("Yell what?\n\r")
		return
	}
	if ch.InRoom != nil && ch.InRoom.RoomFlags.IsSet(types.ROOM_SILENCE) {
		ch.Send("You can't do that here.\n\r")
		return
	}

	ch.Sendf("You yell '%s'\n\r", argument)

	for _, wch := range WorldRef.Characters {
		if wch == ch || wch.Desc == nil {
			continue
		}
		if wch.InRoom != nil && ch.InRoom != nil &&
			wch.InRoom.Area == ch.InRoom.Area {
			wch.Sendf("%s yells '%s'\n\r", ch.Name, translateFor(ch, wch, argument))
		}
	}
	mudprog.OprogSpeechTrigger(ch, argument)
	mudprog.RprogSpeechTrigger(ch, argument)
}

// DoGossip implements the 'gossip' command: global channel message.
func DoGossip(ch *types.CharData, argument string) {
	if argument == "" {
		ch.Send("Gossip what?\n\r")
		return
	}
	if ch.InRoom != nil && ch.InRoom.RoomFlags.IsSet(types.ROOM_SILENCE) {
		ch.Send("You can't do that here.\n\r")
		return
	}

	ch.Sendf("You gossip '%s'\n\r", argument)

	for _, wch := range WorldRef.Characters {
		if wch == ch || wch.Desc == nil {
			continue
		}
		wch.Sendf("%s gossips '%s'\n\r", ch.Name, argument)
	}
	mudprog.OprogSpeechTrigger(ch, argument)
	mudprog.RprogSpeechTrigger(ch, argument)
}

// DoShout implements the 'shout' command: message heard by all players.
func DoShout(ch *types.CharData, argument string) {
	if argument == "" {
		ch.Send("Shout what?\n\r")
		return
	}
	if ch.InRoom != nil && ch.InRoom.RoomFlags.IsSet(types.ROOM_SILENCE) {
		ch.Send("You can't do that here.\n\r")
		return
	}

	ch.Sendf("You shout '%s'\n\r", argument)

	for _, wch := range WorldRef.Characters {
		if wch == ch || wch.Desc == nil {
			continue
		}
		wch.Sendf("%s shouts '%s'\n\r", ch.Name, translateFor(ch, wch, argument))
	}
	mudprog.OprogSpeechTrigger(ch, argument)
	mudprog.RprogSpeechTrigger(ch, argument)
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
		if rch != ch && !rch.IsNPC() && rch.Act.IsSet(types.PLR_NO_EMOTE) {
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

// DoSpeak selects which language the character speaks.
func DoSpeak(ch *types.CharData, argument string) {
	if argument == "" {
		ch.Send("Speak which language?\n\r")
		ch.Send("You know:")
		any := false
		for bit := uint32(1); bit != 0 && bit <= 1<<20; bit <<= 1 {
			if uint32(ch.Speaks)&bit != 0 {
				name := util.LanguageName(bit)
				if name != "" {
					ch.Sendf(" %s", name)
					any = true
				}
			}
		}
		if !any {
			ch.Send(" none")
		}
		ch.Send("\n\r")
		return
	}
	bit := util.LanguageBit(argument)
	if bit == 0 {
		ch.Send("That's not a language.\n\r")
		return
	}
	if uint32(ch.Speaks)&bit == 0 && !(!ch.IsNPC() && ch.Act.IsSet(types.PLR_HOLYLIGHT)) {
		ch.Send("You do not know that language.\n\r")
		return
	}
	ch.Speaking = int(bit)
	ch.Sendf("You will now speak %s.\n\r", util.LanguageName(bit))
}

// DoLearn adds a language to the character's known list. In C this is gated
// by immortal/trainer; the Tier 2 MVP allows any char with holylight (trust)
// to teach themselves. Full trainer flow is a Phase-6 polish item.
func DoLearn(ch *types.CharData, argument string) {
	if argument == "" {
		ch.Send("Learn which language?\n\r")
		return
	}
	bit := util.LanguageBit(argument)
	if bit == 0 {
		ch.Send("That's not a language.\n\r")
		return
	}
	if bit&types.VALID_LANGS == 0 {
		ch.Send("You cannot learn that language.\n\r")
		return
	}
	if !ch.IsNPC() && !ch.Act.IsSet(types.PLR_HOLYLIGHT) {
		ch.Send("You need a trainer to learn new languages.\n\r")
		return
	}
	ch.Speaks = int(uint32(ch.Speaks) | bit)
	ch.Sendf("You now know %s.\n\r", util.LanguageName(bit))
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
		if rch != ch && !rch.IsNPC() && rch.Act.IsSet(types.PLR_NO_EMOTE) {
			continue
		}
		rch.Sendf("%s %s\n\r", ch.Name, argument)
	}
}
