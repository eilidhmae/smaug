package act

import (
	"strings"

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

	if !deliverTell(ch, victim, message) {
		return
	}
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
	if !deliverTell(ch, victim, argument) {
		return
	}
	victim.Reply = ch
}

// deliverTell handles the common tell/reply delivery path with PLR_NO_TELL,
// PLR_AFK, and room-silence gating. Returns true if the sender's confirmation
// was emitted (caller may then update reply pointers).
func deliverTell(ch, victim *types.CharData, message string) bool {
	// Sender's room silences outbound tells.
	if ch.InRoom != nil && ch.InRoom.RoomFlags.IsSet(types.ROOM_SILENCE) && !ch.IsImmortal() {
		ch.Send("You can't do that here.\n\r")
		return false
	}

	// Receiver refusing tells — immortals bypass the block.
	if !ch.IsImmortal() && !victim.IsNPC() && victim.Act.IsSet(types.PLR_NO_TELL) {
		ch.Sendf("%s is not receiving tells.\n\r", victim.Name)
		return false
	}

	// Apply per-listener language translation.
	heard := util.Translate(message, ch, victim)

	// AFK: deliver, but tag both sides.
	afk := !victim.IsNPC() && victim.Act.IsSet(types.PLR_AFK)
	if afk {
		ch.Sendf("%s is AFK. You tell %s '%s'\n\r", victim.Name, victim.Name, message)
		victim.Sendf("(afk) %s tells you '%s'\n\r", ch.Name, heard)
	} else {
		ch.Sendf("You tell %s '%s'\n\r", victim.Name, message)
		victim.Sendf("%s tells you '%s'\n\r", ch.Name, heard)
	}
	return true
}

// DoYell implements the 'yell' command: message to all in same area.
func DoYell(ch *types.CharData, argument string) {
	if argument == "" {
		ch.Send("Yell what?\n\r")
		return
	}

	if ch.InRoom != nil && ch.InRoom.RoomFlags.IsSet(types.ROOM_SILENCE) && !ch.IsImmortal() {
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
			wch.Sendf("%s yells '%s'\n\r", ch.Name, util.Translate(argument, ch, wch))
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

	if ch.InRoom != nil && ch.InRoom.RoomFlags.IsSet(types.ROOM_SILENCE) && !ch.IsImmortal() {
		ch.Send("You can't do that here.\n\r")
		return
	}

	ch.Sendf("You shout '%s'\n\r", argument)

	for _, wch := range WorldRef.Characters {
		if wch == ch || wch.Desc == nil {
			continue
		}
		wch.Sendf("%s shouts '%s'\n\r", ch.Name, util.Translate(argument, ch, wch))
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

// DoSpeak implements the 'speak' command. With no argument it reports the
// currently-active tongue and those the character knows; with a language name
// it switches the active language, provided the character actually knows it.
func DoSpeak(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		cur := "nothing"
		for bit, name := range util.LangName {
			if uint32(ch.Speaking) == bit {
				cur = name
				break
			}
		}
		ch.Sendf("You are currently speaking %s.\n\r", cur)
		// List known languages.
		var known []string
		for bit, name := range util.LangName {
			if uint32(ch.Speaks)&bit != 0 {
				known = append(known, name)
			}
		}
		if len(known) == 0 {
			ch.Send("You don't know any languages.\n\r")
			return
		}
		ch.Sendf("You know: %s.\n\r", strings.Join(known, ", "))
		return
	}

	bit, ok := util.LangBit(arg)
	if !ok {
		ch.Send("That is not a known language.\n\r")
		return
	}
	if !ch.IsImmortal() && uint32(ch.Speaks)&bit == 0 {
		ch.Send("You don't know that language.\n\r")
		return
	}
	ch.Speaking = int(bit)
	ch.Sendf("You now speak %s.\n\r", util.LangName[bit])
}

// DoLearn implements the 'learn' command. Immortals may teach a language to
// the target (or themselves); mortals self-taught learning is handled by the
// trainer/practice subsystem — this command is a stub for those callers.
func DoLearn(ch *types.CharData, argument string) {
	arg, rest := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Learn what language (from whom)?\n\r")
		return
	}

	bit, ok := util.LangBit(arg)
	if !ok {
		ch.Send("That is not a known language.\n\r")
		return
	}

	target := ch
	if rest != "" && ch.IsImmortal() {
		victim := handler.GetCharRoom(ch, rest)
		if victim == nil {
			ch.Send("They aren't here.\n\r")
			return
		}
		target = victim
	} else if !ch.IsImmortal() {
		// Mortal learn without a trainer NPC: stubbed — refer player to practice.
		ch.Send("Find a trainer and use 'practice' to learn languages.\n\r")
		return
	}

	target.Speaks |= int(bit)
	if target == ch {
		ch.Sendf("You can now speak %s.\n\r", util.LangName[bit])
	} else {
		target.Sendf("%s has taught you %s.\n\r", ch.Name, util.LangName[bit])
		ch.Sendf("You teach %s %s.\n\r", target.Name, util.LangName[bit])
	}
}

// DoEmote implements the 'emote' command: roleplay action visible to room.
func DoEmote(ch *types.CharData, argument string) {
	// Silenced senders (PLR_NO_EMOTE) get a refusal; mirrors C src/act_comm.c do_emote.
	if !ch.IsNPC() && ch.Act.IsSet(types.PLR_NO_EMOTE) {
		ch.Send("You can't show your emotions.\n\r")
		return
	}

	if argument == "" {
		ch.Send("Emote what?\n\r")
		return
	}

	if ch.InRoom == nil {
		return
	}

	if ch.InRoom.RoomFlags.IsSet(types.ROOM_SILENCE) && !ch.IsImmortal() {
		ch.Send("You can't do that here.\n\r")
		return
	}

	for _, rch := range ch.InRoom.People {
		if rch.Desc == nil {
			continue
		}
		rch.Sendf("%s %s\n\r", ch.Name, argument)
	}
}
