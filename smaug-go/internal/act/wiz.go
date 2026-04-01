package act

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/command"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// CmdRegistry is set from main to allow force/at commands to interpret.
var CmdRegistry *command.Registry

// sendToPager sends text through the pager if enabled, else to normal output.
// This is a local version to avoid circular imports with the game package.
func sendToPager(ch *types.CharData, text string) {
	if ch == nil || ch.Desc == nil {
		return
	}
	if ch.IsNPC() || ch.PCData == nil || (ch.PCData.Flags&int(types.PCFLAG_PAGERON)) == 0 {
		ch.Send(text)
		return
	}
	ch.Desc.WriteToPager(text)
}

// DoMstat implements the 'mstat' command: show detailed mob/character stats.
func DoMstat(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Stat whom?\n\r")
		return
	}

	victim := handler.GetCharWorld(WorldRef, ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}

	var sb strings.Builder

	fmt.Fprintf(&sb, "&W--- Mobile/Character Stat: %s ---&D\n\r", victim.Name)
	fmt.Fprintf(&sb, "Name: %-20s  Short: %s\n\r", victim.Name, victim.ShortDescr)
	if victim.LongDescr != "" {
		fmt.Fprintf(&sb, "Long: %s", victim.LongDescr)
	}

	raceName := fmt.Sprintf("%d", victim.Race)
	if victim.Race >= 0 && victim.Race < len(WorldRef.Races) && WorldRef.Races[victim.Race] != nil {
		raceName = WorldRef.Races[victim.Race].Name
	}
	className := fmt.Sprintf("%d", victim.Class)
	if victim.Class >= 0 && victim.Class < len(WorldRef.Classes) && WorldRef.Classes[victim.Class] != nil {
		className = WorldRef.Classes[victim.Class].WhoName
	}

	fmt.Fprintf(&sb, "Level: %-3d  Race: %-10s  Class: %-10s  Sex: %d\n\r",
		victim.Level, raceName, className, victim.Sex)
	fmt.Fprintf(&sb, "Hp: %d/%d  Mana: %d/%d  Move: %d/%d\n\r",
		victim.Hit, victim.MaxHit, victim.Mana, victim.MaxMana, victim.Move, victim.MaxMove)
	fmt.Fprintf(&sb, "Str: %d  Int: %d  Wis: %d  Dex: %d  Con: %d  Cha: %d  Lck: %d\n\r",
		victim.GetCurrStr(), victim.GetCurrInt(), victim.GetCurrWis(),
		victim.GetCurrDex(), victim.GetCurrCon(), victim.GetCurrCha(), victim.GetCurrLck())
	fmt.Fprintf(&sb, "Hitroll: %d  Damroll: %d  Armor: %d\n\r",
		victim.Hitroll, victim.Damroll, victim.Armor)
	fmt.Fprintf(&sb, "Position: %d  Alignment: %d  Gold: %d  Exp: %d\n\r",
		victim.Position, victim.Alignment, victim.Gold, victim.Exp)
	fmt.Fprintf(&sb, "Bare dice: %dd%d  Thac0: %d  Attacks: %d\n\r",
		victim.BareNumDie, victim.BareSizeDie, victim.MobThac0, victim.NumAttacks)

	if victim.InRoom != nil {
		fmt.Fprintf(&sb, "Room: %d (%s)\n\r", victim.InRoom.Vnum, victim.InRoom.Name)
	}

	if victim.IsNPC() && victim.IndexData != nil {
		fmt.Fprintf(&sb, "Vnum: %d  Count: %d\n\r", victim.IndexData.Vnum, victim.IndexData.Count)
	}

	fmt.Fprintf(&sb, "Act flags: %s\n\r", victim.Act.String())
	fmt.Fprintf(&sb, "Affected_by: %s\n\r", victim.AffectedBy.String())

	if len(victim.Affects) > 0 {
		fmt.Fprintf(&sb, "Affects:\n\r")
		for _, aff := range victim.Affects {
			name := "unknown"
			if aff.Type >= 0 && aff.Type < len(WorldRef.Skills) && WorldRef.Skills[aff.Type] != nil {
				name = WorldRef.Skills[aff.Type].Name
			}
			fmt.Fprintf(&sb, "  %-18s: loc %d mod %d dur %d bv %s\n\r",
				name, aff.Location, aff.Modifier, aff.Duration, aff.BitVector.String())
		}
	}

	fmt.Fprintf(&sb, "Carrying: %d items (%d weight)\n\r", victim.CarryNumber, victim.CarryWeight)

	if victim.Fighting != nil {
		fmt.Fprintf(&sb, "Fighting: %s\n\r", victim.Fighting.Who.Name)
	}

	sendToPager(ch, sb.String())
}

// DoOstat implements the 'ostat' command: show detailed object stats.
func DoOstat(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Stat what object?\n\r")
		return
	}

	obj := handler.GetObjWorld(WorldRef, ch, arg)
	if obj == nil {
		ch.Send("Nothing like that in the game.\n\r")
		return
	}

	var sb strings.Builder

	fmt.Fprintf(&sb, "&W--- Object Stat: %s ---&D\n\r", obj.ShortDescr)
	fmt.Fprintf(&sb, "Name: %s\n\r", obj.Name)
	fmt.Fprintf(&sb, "Short: %s\n\r", obj.ShortDescr)
	if obj.Description != "" {
		fmt.Fprintf(&sb, "Long: %s\n\r", obj.Description)
	}

	if obj.IndexData != nil {
		fmt.Fprintf(&sb, "Vnum: %d  Count: %d\n\r", obj.IndexData.Vnum, obj.IndexData.Count)
	}

	fmt.Fprintf(&sb, "Type: %d  Level: %d  Weight: %d  Timer: %d\n\r",
		obj.ItemType, obj.Level, obj.Weight, obj.Timer)
	fmt.Fprintf(&sb, "Cost: %d gold  Wear: %d  WearLoc: %d\n\r",
		obj.GoldCost, obj.WearFlags, obj.WearLoc)
	fmt.Fprintf(&sb, "Extra flags: %s\n\r", obj.ExtraFlags.String())
	fmt.Fprintf(&sb, "Values: %d %d %d %d %d %d\n\r",
		obj.Value[0], obj.Value[1], obj.Value[2],
		obj.Value[3], obj.Value[4], obj.Value[5])

	if obj.CarriedBy != nil {
		fmt.Fprintf(&sb, "Carried by: %s\n\r", obj.CarriedBy.Name)
	}
	if obj.InRoom != nil {
		fmt.Fprintf(&sb, "In room: %d (%s)\n\r", obj.InRoom.Vnum, obj.InRoom.Name)
	}
	if obj.InObj != nil {
		fmt.Fprintf(&sb, "In object: %s\n\r", obj.InObj.ShortDescr)
	}

	if len(obj.Affects) > 0 {
		fmt.Fprintf(&sb, "Affects:\n\r")
		for _, aff := range obj.Affects {
			fmt.Fprintf(&sb, "  loc %d mod %d\n\r", aff.Location, aff.Modifier)
		}
	}

	if len(obj.Contents) > 0 {
		fmt.Fprintf(&sb, "Contents: %d items\n\r", len(obj.Contents))
	}

	sendToPager(ch, sb.String())
}

// DoRstat implements the 'rstat' command: show detailed room stats.
func DoRstat(ch *types.CharData, argument string) {
	var room *types.RoomIndexData

	if argument == "" {
		room = ch.InRoom
	} else {
		// Try to find room by vnum or current room
		room = ch.InRoom
	}

	if room == nil {
		ch.Send("You are not in a room.\n\r")
		return
	}

	var sb strings.Builder

	fmt.Fprintf(&sb, "&W--- Room Stat: %s ---&D\n\r", room.Name)
	fmt.Fprintf(&sb, "Vnum: %d  Sector: %d  Light: %d\n\r",
		room.Vnum, room.SectorType, room.Light)
	fmt.Fprintf(&sb, "Room flags: %s\n\r", room.RoomFlags.String())
	if room.Description != "" {
		fmt.Fprintf(&sb, "Description:\n\r%s", room.Description)
	}
	if room.Area != nil {
		fmt.Fprintf(&sb, "Area: %s\n\r", room.Area.Name)
	}

	// Exits
	dirNames := []string{"north", "east", "south", "west", "up", "down",
		"northeast", "northwest", "southeast", "southwest"}
	for _, ex := range room.Exits {
		if ex == nil {
			continue
		}
		dir := "?"
		if ex.Direction >= 0 && ex.Direction < len(dirNames) {
			dir = dirNames[ex.Direction]
		}
		dest := -1
		if ex.ToRoom != nil {
			dest = ex.ToRoom.Vnum
		}
		fmt.Fprintf(&sb, "Exit %-9s → %d  Key: %d  Flags: %d",
			dir, dest, ex.Key, ex.ExitInfo)
		if ex.Keyword != "" {
			fmt.Fprintf(&sb, "  Keyword: %s", ex.Keyword)
		}
		fmt.Fprintf(&sb, "\n\r")
	}

	// People
	if len(room.People) > 0 {
		fmt.Fprintf(&sb, "People: ")
		for i, rch := range room.People {
			if i > 0 {
				fmt.Fprintf(&sb, ", ")
			}
			fmt.Fprintf(&sb, "%s", rch.Name)
		}
		fmt.Fprintf(&sb, "\n\r")
	}

	// Objects
	if len(room.Contents) > 0 {
		fmt.Fprintf(&sb, "Objects: ")
		for i, obj := range room.Contents {
			if i > 0 {
				fmt.Fprintf(&sb, ", ")
			}
			fmt.Fprintf(&sb, "%s", obj.ShortDescr)
		}
		fmt.Fprintf(&sb, "\n\r")
	}

	// Extra descriptions
	if len(room.ExtraDescr) > 0 {
		fmt.Fprintf(&sb, "Extra descs: ")
		for i, ed := range room.ExtraDescr {
			if i > 0 {
				fmt.Fprintf(&sb, ", ")
			}
			fmt.Fprintf(&sb, "%s", ed.Keyword)
		}
		fmt.Fprintf(&sb, "\n\r")
	}

	sendToPager(ch, sb.String())
}

// --- Movement Commands ---

// DoGoto implements the 'goto' command: teleport to a room, mob, or player.
func DoGoto(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Goto where?\n\r")
		return
	}

	// Try vnum first
	if vnum, err := strconv.Atoi(arg); err == nil {
		room := WorldRef.GetRoom(vnum)
		if room != nil {
			teleportTo(ch, room)
			return
		}
		ch.Sendf("No room with vnum %d.\n\r", vnum)
		return
	}

	// Try character name
	victim := handler.GetCharWorld(WorldRef, ch, arg)
	if victim != nil && victim.InRoom != nil {
		teleportTo(ch, victim.InRoom)
		return
	}

	ch.Send("No such location.\n\r")
}

// teleportTo moves an immortal to a room with bamf messages.
func teleportTo(ch *types.CharData, dest *types.RoomIndexData) {
	if ch.InRoom != nil {
		msg := ch.Name + " leaves in a swirling mist."
		if ch.PCData != nil && ch.PCData.BamfOut != "" {
			msg = ch.PCData.BamfOut
		}
		for _, rch := range ch.InRoom.People {
			if rch != ch && rch.Desc != nil {
				rch.Sendf("%s\n\r", msg)
			}
		}
		handler.CharFromRoom(ch)
	}
	handler.CharToRoom(ch, dest)

	msg := ch.Name + " appears in a swirling mist."
	if ch.PCData != nil && ch.PCData.BamfIn != "" {
		msg = ch.PCData.BamfIn
	}
	for _, rch := range dest.People {
		if rch != ch && rch.Desc != nil {
			rch.Sendf("%s\n\r", msg)
		}
	}

	DoLook(ch, "auto")
}

// DoTransfer implements the 'transfer' command: move a character to you.
func DoTransfer(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Transfer whom?\n\r")
		return
	}

	if ch.InRoom == nil {
		return
	}

	if strings.EqualFold(arg, "all") {
		for _, d := range WorldRef.Descriptors {
			if d.Character != nil && d.Character != ch && d.Character.InRoom != nil {
				victim := d.Character
				handler.CharFromRoom(victim)
				handler.CharToRoom(victim, ch.InRoom)
				victim.Sendf("%s has transferred you.\n\r", ch.Name)
				DoLook(victim, "auto")
			}
		}
		ch.Send("All online players transferred.\n\r")
		return
	}

	victim := handler.GetCharWorld(WorldRef, ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}
	if victim == ch {
		ch.Send("That's you.\n\r")
		return
	}

	handler.CharFromRoom(victim)
	handler.CharToRoom(victim, ch.InRoom)
	victim.Sendf("%s has transferred you.\n\r", ch.Name)
	DoLook(victim, "auto")
	ch.Sendf("%s transferred.\n\r", victim.Name)
}

// DoAt implements the 'at' command: execute a command at another location.
func DoAt(ch *types.CharData, argument string) {
	arg, rest := util.OneArgument(argument)
	if arg == "" || rest == "" {
		ch.Send("At where what?\n\r")
		return
	}

	var dest *types.RoomIndexData

	// Try vnum
	if vnum, err := strconv.Atoi(arg); err == nil {
		dest = WorldRef.GetRoom(vnum)
	}
	// Try character
	if dest == nil {
		victim := handler.GetCharWorld(WorldRef, ch, arg)
		if victim != nil {
			dest = victim.InRoom
		}
	}

	if dest == nil {
		ch.Send("No such location.\n\r")
		return
	}

	origRoom := ch.InRoom
	handler.CharFromRoom(ch)
	handler.CharToRoom(ch, dest)

	if CmdRegistry != nil {
		CmdRegistry.Interpret(ch, rest)
	}

	// Return if still alive and in the destination
	if ch.InRoom == dest {
		handler.CharFromRoom(ch)
		handler.CharToRoom(ch, origRoom)
	}
}

// DoBamfin implements the 'bamfin' command: set arrival message.
func DoBamfin(ch *types.CharData, argument string) {
	if ch.PCData == nil {
		return
	}
	if argument == "" {
		ch.Sendf("Your bamfin is: %s\n\r", ch.PCData.BamfIn)
		return
	}
	ch.PCData.BamfIn = argument
	ch.Sendf("Bamfin set to: %s\n\r", argument)
}

// DoBamfout implements the 'bamfout' command: set departure message.
func DoBamfout(ch *types.CharData, argument string) {
	if ch.PCData == nil {
		return
	}
	if argument == "" {
		ch.Sendf("Your bamfout is: %s\n\r", ch.PCData.BamfOut)
		return
	}
	ch.PCData.BamfOut = argument
	ch.Sendf("Bamfout set to: %s\n\r", argument)
}

// --- Action Commands ---

// DoForce implements the 'force' command: compel a character to execute a command.
func DoForce(ch *types.CharData, argument string) {
	arg, rest := util.OneArgument(argument)
	if arg == "" || rest == "" {
		ch.Send("Force whom to do what?\n\r")
		return
	}

	if strings.EqualFold(arg, "all") {
		for _, d := range WorldRef.Descriptors {
			if d.Character != nil && d.Character != ch &&
				d.Character.GetTrust() < ch.GetTrust() {
				if CmdRegistry != nil {
					CmdRegistry.Interpret(d.Character, rest)
				}
			}
		}
		ch.Send("Force all done.\n\r")
		return
	}

	victim := handler.GetCharWorld(WorldRef, ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}
	if victim == ch {
		ch.Send("Force yourself?\n\r")
		return
	}
	if !victim.IsNPC() && victim.GetTrust() >= ch.GetTrust() {
		ch.Send("You can't force that.\n\r")
		return
	}

	if CmdRegistry != nil {
		CmdRegistry.Interpret(victim, rest)
	}
	ch.Send("Ok.\n\r")
}

// DoPeace implements the 'peace' command: stop all fighting in the room.
func DoPeace(ch *types.CharData, argument string) {
	if ch.InRoom == nil {
		return
	}
	for _, rch := range ch.InRoom.People {
		if rch.Fighting != nil {
			combat.StopFighting(rch, true)
		}
		rch.Position = types.POS_STANDING
	}
	ch.Send("Peace has been restored.\n\r")
}

// DoPurge implements the 'purge' command: remove NPCs/objects from room.
func DoPurge(ch *types.CharData, argument string) {
	if ch.InRoom == nil {
		return
	}

	if argument == "" {
		// Purge all NPCs and objects in room
		for i := len(ch.InRoom.People) - 1; i >= 0; i-- {
			rch := ch.InRoom.People[i]
			if rch.IsNPC() {
				handler.ExtractChar(WorldRef, rch, true)
			}
		}
		for i := len(ch.InRoom.Contents) - 1; i >= 0; i-- {
			handler.ExtractObj(WorldRef, ch.InRoom.Contents[i])
		}
		ch.Send("Room purged.\n\r")
		return
	}

	arg, _ := util.OneArgument(argument)

	// Try NPC
	victim := handler.GetCharRoom(ch, arg)
	if victim != nil {
		if !victim.IsNPC() {
			ch.Send("You can only purge NPCs.\n\r")
			return
		}
		handler.ExtractChar(WorldRef, victim, true)
		ch.Send("Ok.\n\r")
		return
	}

	// Try object
	obj := handler.GetObjHere(ch, arg)
	if obj != nil {
		handler.ExtractObj(WorldRef, obj)
		ch.Send("Ok.\n\r")
		return
	}

	ch.Send("Nothing like that here.\n\r")
}

// DoRestore implements the 'restore' command: restore HP/mana/move to max.
func DoRestore(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)

	if strings.EqualFold(arg, "all") {
		for _, d := range WorldRef.Descriptors {
			if d.Character != nil {
				v := d.Character
				v.Hit = v.MaxHit
				v.Mana = v.MaxMana
				v.Move = v.MaxMove
			}
		}
		ch.Send("All players restored.\n\r")
		return
	}

	if arg == "" {
		ch.Send("Restore whom?\n\r")
		return
	}

	victim := handler.GetCharWorld(WorldRef, ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}

	victim.Hit = victim.MaxHit
	victim.Mana = victim.MaxMana
	victim.Move = victim.MaxMove
	victim.Send("You feel fully restored!\n\r")
	if victim != ch {
		ch.Sendf("%s restored.\n\r", victim.Name)
	}
}

// DoAdvance implements the 'advance' command: set a player's level.
func DoAdvance(ch *types.CharData, argument string) {
	arg, rest := util.OneArgument(argument)
	if arg == "" || rest == "" {
		ch.Send("Usage: advance <player> <level>\n\r")
		return
	}

	victim := handler.GetCharRoom(ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}
	if victim.IsNPC() {
		ch.Send("Not on NPCs.\n\r")
		return
	}

	level, err := strconv.Atoi(strings.TrimSpace(rest))
	if err != nil || level < 1 || level > types.MAX_LEVEL {
		ch.Sendf("Level must be 1 to %d.\n\r", types.MAX_LEVEL)
		return
	}
	if level > ch.GetTrust() {
		ch.Send("You can't advance someone beyond your own trust.\n\r")
		return
	}

	victim.Level = level
	victim.Sendf("Your level has been set to %d!\n\r", level)
	if victim != ch {
		ch.Sendf("%s advanced to level %d.\n\r", victim.Name, level)
	}
}

// DoSlay implements the 'slay' command: instant kill.
func DoSlay(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Slay whom?\n\r")
		return
	}

	victim := handler.GetCharRoom(ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}
	if victim == ch {
		ch.Send("Suicide is a mortal sin.\n\r")
		return
	}
	if !victim.IsNPC() && victim.GetTrust() >= ch.GetTrust() {
		ch.Send("You can't slay that player.\n\r")
		return
	}

	ch.Sendf("You slay %s in cold blood!\n\r", victim.Name)
	victim.Send("You have been SLAIN!\n\r")

	if victim.IsNPC() {
		combat.MakeCorpse(WorldRef, victim)
		handler.ExtractChar(WorldRef, victim, true)
	} else {
		victim.Hit = 1
		victim.Mana = 1
		victim.Move = 1
		victim.Position = types.POS_RESTING
	}
}

// --- Info Commands ---

// DoMfind implements the 'mfind' command: find mob templates by name.
func DoMfind(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Find what mob?\n\r")
		return
	}

	var sb strings.Builder
	found := false
	for vnum, idx := range WorldRef.MobIndex {
		if util.IsName(arg, idx.PlayerName) {
			fmt.Fprintf(&sb, "[%5d] %s\n\r", vnum, idx.ShortDescr)
			found = true
		}
	}
	if !found {
		ch.Send("No mobs found.\n\r")
		return
	}
	sendToPager(ch, sb.String())
}

// DoOfind implements the 'ofind' command: find object templates by name.
func DoOfind(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Find what object?\n\r")
		return
	}

	var sb strings.Builder
	found := false
	for vnum, idx := range WorldRef.ObjIndex {
		if util.IsName(arg, idx.Name) {
			fmt.Fprintf(&sb, "[%5d] %s\n\r", vnum, idx.ShortDescr)
			found = true
		}
	}
	if !found {
		ch.Send("No objects found.\n\r")
		return
	}
	sendToPager(ch, sb.String())
}

// DoMwhere implements the 'mwhere' command: find mob instances in world.
func DoMwhere(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Find what mob?\n\r")
		return
	}

	var sb strings.Builder
	found := false
	for _, mob := range WorldRef.Characters {
		if !mob.IsNPC() {
			continue
		}
		if !util.IsName(arg, mob.Name) {
			continue
		}
		roomName := "nowhere"
		roomVnum := 0
		if mob.InRoom != nil {
			roomName = mob.InRoom.Name
			roomVnum = mob.InRoom.Vnum
		}
		mobVnum := 0
		if mob.IndexData != nil {
			mobVnum = mob.IndexData.Vnum
		}
		fmt.Fprintf(&sb, "[%5d] %-20s in [%5d] %s\n\r",
			mobVnum, mob.ShortDescr, roomVnum, roomName)
		found = true
	}
	if !found {
		ch.Send("No mobs found.\n\r")
		return
	}
	sendToPager(ch, sb.String())
}

// DoOwhere implements the 'owhere' command: find object instances in world.
func DoOwhere(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Find what object?\n\r")
		return
	}

	var sb strings.Builder
	found := false
	for _, obj := range WorldRef.Objects {
		if !util.IsName(arg, obj.Name) {
			continue
		}
		location := "somewhere"
		if obj.CarriedBy != nil {
			location = fmt.Sprintf("carried by %s", obj.CarriedBy.Name)
		} else if obj.InRoom != nil {
			location = fmt.Sprintf("in [%d] %s", obj.InRoom.Vnum, obj.InRoom.Name)
		} else if obj.InObj != nil {
			location = fmt.Sprintf("in %s", obj.InObj.ShortDescr)
		}
		vnum := 0
		if obj.IndexData != nil {
			vnum = obj.IndexData.Vnum
		}
		fmt.Fprintf(&sb, "[%5d] %-20s %s\n\r", vnum, obj.ShortDescr, location)
		found = true
	}
	if !found {
		ch.Send("No objects found.\n\r")
		return
	}
	sendToPager(ch, sb.String())
}

// DoUsers implements the 'users' command: show connected users.
func DoUsers(ch *types.CharData, argument string) {
	var sb strings.Builder
	fmt.Fprintf(&sb, "&WDesc  Connected  Name          Host&D\n\r")
	fmt.Fprintf(&sb, "----  ---------  ----          ----\n\r")
	count := 0
	for i, d := range WorldRef.Descriptors {
		name := "(none)"
		if d.Character != nil {
			name = d.Character.Name
		} else if d.User != "" {
			name = d.User
		}
		state := "playing"
		if d.Connected != types.CON_PLAYING {
			state = fmt.Sprintf("state %d", d.Connected)
		}
		fmt.Fprintf(&sb, "%4d  %-9s  %-12s  %s\n\r", i, state, name, d.Host)
		count++
	}
	fmt.Fprintf(&sb, "\n\r%d user(s).\n\r", count)
	sendToPager(ch, sb.String())
}

// --- Control Commands ---

// DoInvis implements the 'invis' command: toggle immortal invisibility.
func DoInvis(ch *types.CharData, argument string) {
	if ch.PCData == nil {
		return
	}

	arg := strings.TrimSpace(argument)
	if arg != "" {
		level, err := strconv.Atoi(arg)
		if err != nil || level < 2 || level > ch.GetTrust() {
			ch.Sendf("Valid range is 2 to %d.\n\r", ch.GetTrust())
			return
		}
		ch.PCData.WizInvis = level
		ch.Act.Set(types.PLR_WIZINVIS)
		ch.Sendf("Wizinvis level set to %d.\n\r", level)
		return
	}

	if ch.Act.IsSet(types.PLR_WIZINVIS) {
		ch.Act.Toggle(types.PLR_WIZINVIS)
		ch.PCData.WizInvis = 0
		ch.Send("You are now visible.\n\r")
	} else {
		ch.Act.Set(types.PLR_WIZINVIS)
		ch.PCData.WizInvis = ch.GetTrust()
		ch.Sendf("You are now invisible to level %d.\n\r", ch.PCData.WizInvis)
	}
}

// DoHolylight implements the 'holylight' command: toggle see-all mode.
func DoHolylight(ch *types.CharData, argument string) {
	if ch.IsNPC() {
		return
	}
	if ch.Act.IsSet(types.PLR_HOLYLIGHT) {
		ch.Act.Toggle(types.PLR_HOLYLIGHT)
		ch.Send("Holy light mode off.\n\r")
	} else {
		ch.Act.Set(types.PLR_HOLYLIGHT)
		ch.Send("Holy light mode on.\n\r")
	}
}

// DoFreeze implements the 'freeze' command: prevent a player from acting.
func DoFreeze(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Freeze whom?\n\r")
		return
	}

	victim := handler.GetCharWorld(WorldRef, ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}
	if victim.IsNPC() {
		ch.Send("Not on NPCs.\n\r")
		return
	}
	if victim.GetTrust() >= ch.GetTrust() {
		ch.Send("You can't do that.\n\r")
		return
	}

	if victim.Act.IsSet(types.PLR_FREEZE) {
		victim.Act.Toggle(types.PLR_FREEZE)
		victim.Send("You can play again.\n\r")
		ch.Sendf("%s unfrozen.\n\r", victim.Name)
	} else {
		victim.Act.Set(types.PLR_FREEZE)
		victim.Send("You have been frozen!\n\r")
		ch.Sendf("%s frozen.\n\r", victim.Name)
	}
}

// DoSilence implements the 'silence' command: prevent channel use.
func DoSilence(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Silence whom?\n\r")
		return
	}

	victim := handler.GetCharWorld(WorldRef, ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}
	if victim.IsNPC() {
		ch.Send("Not on NPCs.\n\r")
		return
	}
	if victim.GetTrust() >= ch.GetTrust() {
		ch.Send("You can't do that.\n\r")
		return
	}

	if victim.Act.IsSet(types.PLR_SILENCE) {
		victim.Act.Toggle(types.PLR_SILENCE)
		victim.Send("You can use channels again.\n\r")
		ch.Sendf("%s unsilenced.\n\r", victim.Name)
	} else {
		victim.Act.Set(types.PLR_SILENCE)
		victim.Send("You have been silenced!\n\r")
		ch.Sendf("%s silenced.\n\r", victim.Name)
	}
}

// --- System Commands ---

// DoEcho implements the 'echo' command: send message to all players.
func DoEcho(ch *types.CharData, argument string) {
	if argument == "" {
		ch.Send("Echo what?\n\r")
		return
	}

	for _, d := range WorldRef.Descriptors {
		if d.Character != nil {
			d.Character.Sendf("%s\n\r", argument)
		}
	}
}

// DoRecho implements the 'recho' command: send message to current room.
func DoRecho(ch *types.CharData, argument string) {
	if argument == "" {
		ch.Send("Recho what?\n\r")
		return
	}
	if ch.InRoom == nil {
		return
	}

	for _, rch := range ch.InRoom.People {
		if rch.Desc != nil {
			rch.Sendf("%s\n\r", argument)
		}
	}
}
