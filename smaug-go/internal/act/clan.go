package act

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// findClan finds a clan by name (exact or prefix).
func findClan(name string) *types.ClanData {
	if WorldRef == nil {
		return nil
	}
	name = strings.ToLower(name)
	for _, c := range WorldRef.Clans {
		if strings.EqualFold(c.Name, name) {
			return c
		}
	}
	for _, c := range WorldRef.Clans {
		if strings.HasPrefix(strings.ToLower(c.Name), name) {
			return c
		}
	}
	return nil
}

// DoClans implements the 'clans' command: list all clans.
func DoClans(ch *types.CharData, argument string) {
	if WorldRef == nil || len(WorldRef.Clans) == 0 {
		ch.Send("There are no clans.\n\r")
		return
	}

	ch.Send("&W--- Clans ---&D\n\r")
	for _, c := range WorldRef.Clans {
		ch.Sendf("  %-25s Leader: %-12s Members: %d\n\r",
			c.Name, c.Leader, c.Members)
	}
}

// DoClanInfo implements the 'claninfo' command: show details about a clan.
func DoClanInfo(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Which clan?\n\r")
		return
	}

	clan := findClan(arg)
	if clan == nil {
		ch.Send("No such clan.\n\r")
		return
	}

	ch.Sendf("&W--- %s ---&D\n\r", clan.Name)
	ch.Sendf("Leader:  %s\n\r", clan.Leader)
	if clan.Number1 != "" {
		ch.Sendf("First:   %s\n\r", clan.Number1)
	}
	if clan.Number2 != "" {
		ch.Sendf("Second:  %s\n\r", clan.Number2)
	}
	ch.Sendf("Members: %d\n\r", clan.Members)
	ch.Sendf("PKills:  %d  PDeaths: %d\n\r", clan.PKills[0], clan.PDeaths[0])
	ch.Sendf("MKills:  %d  MDeaths: %d\n\r", clan.MKills, clan.MDeaths)
	if clan.Motto != "" {
		ch.Sendf("Motto:   %s\n\r", clan.Motto)
	}
	if clan.Description != "" {
		ch.Send(clan.Description)
	}
}

// DoClantalk implements the 'clantalk' command: clan-only chat channel.
func DoClantalk(ch *types.CharData, argument string) {
	if ch.IsNPC() || ch.PCData == nil || ch.PCData.Clan == nil {
		ch.Send("You aren't in a clan.\n\r")
		return
	}
	if argument == "" {
		ch.Send("Clantalk what?\n\r")
		return
	}

	clan := ch.PCData.Clan
	ch.Sendf("&G[%s] You clantalk '%s'&D\n\r", clan.Name, argument)

	for _, d := range WorldRef.Descriptors {
		if d.Character != nil && d.Character != ch &&
			d.Character.PCData != nil && d.Character.PCData.Clan == clan {
			d.Character.Sendf("&G[%s] %s clantalks '%s'&D\n\r",
				clan.Name, ch.Name, argument)
		}
	}
}

// DoClanJoin implements the 'join' command: join a clan (simplified).
func DoClanJoin(ch *types.CharData, argument string) {
	if ch.IsNPC() || ch.PCData == nil {
		return
	}
	if ch.PCData.Clan != nil {
		ch.Send("You are already in a clan. You must leave first.\n\r")
		return
	}

	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Join which clan?\n\r")
		return
	}

	clan := findClan(arg)
	if clan == nil {
		ch.Send("No such clan.\n\r")
		return
	}

	ch.PCData.Clan = clan
	ch.PCData.ClanName = clan.Name
	clan.Members++
	ch.Sendf("You join %s!\n\r", clan.Name)
}

// DoClanLeave implements the 'leave' command: leave your clan.
func DoClanLeave(ch *types.CharData, argument string) {
	if ch.IsNPC() || ch.PCData == nil || ch.PCData.Clan == nil {
		ch.Send("You aren't in a clan.\n\r")
		return
	}

	clan := ch.PCData.Clan
	clan.Members--
	ch.Sendf("You leave %s.\n\r", clan.Name)
	ch.PCData.Clan = nil
	ch.PCData.ClanName = ""
}

// --- Deity Commands ---

// DoDeities implements the 'deities' command: list all deities.
func DoDeities(ch *types.CharData, argument string) {
	if WorldRef == nil || len(WorldRef.Deities) == 0 {
		ch.Send("There are no deities.\n\r")
		return
	}

	ch.Send("&W--- Deities ---&D\n\r")
	for _, d := range WorldRef.Deities {
		ch.Sendf("  %-20s Worshippers: %d  Alignment: %d\n\r",
			d.Name, d.Worshippers, d.Alignment)
	}
}

// DoDevote implements the 'devote' command: devote to a deity.
func DoDevote(ch *types.CharData, argument string) {
	if ch.IsNPC() || ch.PCData == nil {
		return
	}

	arg, _ := util.OneArgument(argument)
	if arg == "" {
		if ch.PCData.Deity != nil {
			ch.Sendf("You are devoted to %s.\n\r", ch.PCData.Deity.Name)
		} else {
			ch.Send("You are not devoted to any deity.\n\r")
		}
		return
	}

	if strings.EqualFold(arg, "none") {
		if ch.PCData.Deity != nil {
			ch.Sendf("You renounce your devotion to %s.\n\r", ch.PCData.Deity.Name)
			ch.PCData.Deity.Worshippers--
			ch.PCData.Deity = nil
			ch.PCData.DeityName = ""
		} else {
			ch.Send("You aren't devoted to anyone.\n\r")
		}
		return
	}

	if ch.PCData.Deity != nil {
		ch.Sendf("You are already devoted to %s. Devote none first.\n\r", ch.PCData.Deity.Name)
		return
	}

	var deity *types.DeityData
	for _, d := range WorldRef.Deities {
		if util.IsName(arg, d.Name) {
			deity = d
			break
		}
	}
	if deity == nil {
		ch.Send("No such deity.\n\r")
		return
	}

	ch.PCData.Deity = deity
	ch.PCData.DeityName = deity.Name
	deity.Worshippers++
	ch.Sendf("You devote yourself to %s!\n\r", deity.Name)
}

// --- Board/Note Commands ---

// DoNote implements the 'note' command: read/list/write/post/remove notes.
func DoNote(ch *types.CharData, argument string) {
	arg, rest := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Note what? (list, read <#>, write, post, remove <#>)\n\r")
		return
	}

	// Find a board in the room
	var board *types.BoardData
	if ch.InRoom != nil && WorldRef != nil {
		for _, b := range WorldRef.Boards {
			for _, obj := range ch.InRoom.Contents {
				if obj.IndexData != nil && obj.IndexData.Vnum == b.BoardObj {
					board = b
					break
				}
			}
			if board != nil {
				break
			}
		}
	}

	// Allow note commands even without a board object for simplicity
	if board == nil && WorldRef != nil && len(WorldRef.Boards) > 0 {
		board = WorldRef.Boards[0]
	}

	if board == nil {
		ch.Send("There is no board here.\n\r")
		return
	}

	switch strings.ToLower(arg) {
	case "list":
		if len(board.Notes) == 0 {
			ch.Send("There are no notes.\n\r")
			return
		}
		for i, note := range board.Notes {
			ch.Sendf("[%2d] %s: %s\n\r", i+1, note.Sender, note.Subject)
		}

	case "read":
		num := 0
		if rest != "" {
			num, _ = util.NumberArgument(rest)
		}
		if num < 1 || num > len(board.Notes) {
			ch.Sendf("Valid range is 1 to %d.\n\r", len(board.Notes))
			return
		}
		note := board.Notes[num-1]
		ch.Sendf("&W[%d] %s: %s&D\n\r", num, note.Sender, note.Subject)
		ch.Sendf("Date: %s  To: %s\n\r", note.Date, note.ToList)
		ch.Send(note.Text)
		ch.Send("\n\r")

	case "write":
		// Start a note
		ch.PNote = &types.NoteData{
			Sender: ch.Name,
		}
		ch.Send("Enter the subject of your note: ")
		// In a full implementation this would enter a special nanny state.
		// For now, simplified.
		if rest != "" {
			ch.PNote.Subject = rest
			ch.Send("Enter the recipients (all for everyone): ")
		}

	case "post":
		if ch.PNote == nil {
			ch.Send("You have no note in progress. Use 'note write' first.\n\r")
			return
		}
		if ch.PNote.Subject == "" {
			ch.PNote.Subject = "(no subject)"
		}
		if ch.PNote.ToList == "" {
			ch.PNote.ToList = "all"
		}
		board.Notes = append(board.Notes, ch.PNote)
		ch.Sendf("Note posted: %s\n\r", ch.PNote.Subject)
		ch.PNote = nil

	case "remove":
		num := 0
		if rest != "" {
			num, _ = util.NumberArgument(rest)
		}
		if num < 1 || num > len(board.Notes) {
			ch.Sendf("Valid range is 1 to %d.\n\r", len(board.Notes))
			return
		}
		board.Notes = append(board.Notes[:num-1], board.Notes[num:]...)
		ch.Sendf("Note %d removed.\n\r", num)

	default:
		ch.Send("Note what? (list, read <#>, write, post, remove <#>)\n\r")
	}
}

// --- Snoop Command (moved from wiz for dependency reasons) ---

// DoSnoop implements the 'snoop' command: watch another player's output.
func DoSnoop(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Snoop whom?\n\r")
		return
	}

	victim := handler.GetCharWorld(WorldRef, ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}

	if victim == ch {
		// Cancel all snoops
		for _, d := range WorldRef.Descriptors {
			if d.SnoopBy == ch.Desc {
				d.SnoopBy = nil
			}
		}
		ch.Send("All snoops cancelled.\n\r")
		return
	}

	if victim.Desc == nil {
		ch.Send("No descriptor to snoop.\n\r")
		return
	}

	if !victim.IsNPC() && victim.GetTrust() >= ch.GetTrust() {
		ch.Send("You can't snoop that.\n\r")
		return
	}

	if victim.Desc.SnoopBy != nil {
		ch.Send("They are already being snooped.\n\r")
		return
	}

	victim.Desc.SnoopBy = ch.Desc
	ch.Sendf("You begin snooping %s.\n\r", victim.Name)
}
