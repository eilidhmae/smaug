// Package act implements player-facing commands.
package act

import (
	"fmt"
	"strings"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/mudprog"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// WorldRef holds a reference to the world for commands that need it.
// Set during boot.
// Written once at boot before the game loop starts; read only from the game loop goroutine. Safe without synchronization.
var WorldRef *world.World

// DoLook implements the 'look' command.
func DoLook(ch *types.CharData, argument string) {
	if ch.InRoom == nil {
		ch.Send("You are nowhere!\n\r")
		return
	}

	room := ch.InRoom

	if argument == "" || strings.EqualFold(argument, "auto") {
		// Look at the room
		ch.Sendf("&W%s&D\n\r", room.Name)
		if room.Description != "" {
			ch.Send(room.Description)
		}

		// Show exits
		showExits(ch, room)

		// Show objects in room
		for _, obj := range room.Contents {
			ch.Sendf("&G%s&D\n\r", obj.Description)
		}

		// Show characters in room
		for _, rch := range room.People {
			if rch == ch {
				continue
			}
			if rch.IsNPC() {
				ch.Sendf("&Y%s&D\n\r", rch.LongDescr)
			} else {
				ch.Sendf("&Y%s %s&D\n\r", rch.Name, rch.PCData.Title)
			}
		}
		return
	}

	// Look at something specific
	arg, _ := util.OneArgument(argument)

	// Check for looking at a character
	for _, rch := range room.People {
		if util.IsName(arg, rch.Name) {
			ch.Sendf("You look at %s.\n\r", rch.ShortDescr)
			if rch.Description != "" {
				ch.Send(rch.Description)
			}
			return
		}
	}

	// Check extra descriptions
	for _, ed := range room.ExtraDescr {
		if util.IsName(arg, ed.Keyword) {
			ch.Send(ed.Description)
			return
		}
	}

	// Check objects
	for _, obj := range room.Contents {
		if util.IsName(arg, obj.Name) {
			ch.Sendf("%s\n\r", obj.Description)
			mudprog.OprogLookTrigger(ch, obj)
			return
		}
	}

	// Check inventory
	for _, obj := range ch.Carrying {
		if util.IsName(arg, obj.Name) {
			ch.Sendf("%s\n\r", obj.Description)
			mudprog.OprogLookTrigger(ch, obj)
			return
		}
	}

	ch.Send("You do not see that here.\n\r")
}

// DoExamine implements the 'examine' command: look + show contents.
func DoExamine(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Examine what?\n\r")
		return
	}

	// First do a normal look
	DoLook(ch, arg)

	// Then show container contents or drink level
	obj := handler.GetObjHere(ch, arg)
	if obj == nil {
		return
	}
	mudprog.OprogExamineTrigger(ch, obj)

	switch obj.ItemType {
	case types.ITEM_DRINK_CON:
		if obj.Value[1] <= 0 {
			ch.Send("It is empty.\n\r")
		} else {
			ch.Sendf("It contains some liquid.\n\r")
		}

	case types.ITEM_CONTAINER, types.ITEM_CORPSE_NPC, types.ITEM_CORPSE_PC:
		ch.Sendf("%s contains:\n\r", util.Capitalize(obj.ShortDescr))
		if len(obj.Contents) == 0 {
			ch.Send("  Nothing.\n\r")
		} else {
			for _, item := range obj.Contents {
				ch.Sendf("  %s\n\r", item.ShortDescr)
			}
		}
	}
}

func showExits(ch *types.CharData, room *types.RoomIndexData) {
	dirNames := []string{"north", "east", "south", "west", "up", "down",
		"northeast", "northwest", "southeast", "southwest"}

	holylight := !ch.IsNPC() && ch.PCData != nil && ch.Act.IsSet(types.PLR_HOLYLIGHT)
	var exits []string
	for _, ex := range room.Exits {
		if ex.ToRoom == nil {
			continue
		}
		if ex.ExitInfo&int(types.EX_CLOSED) != 0 {
			continue // don't show closed exits in brief
		}
		if !holylight && ex.ExitInfo&int(types.EX_SECRET|types.EX_HIDDEN) != 0 {
			continue
		}
		if ex.Direction >= 0 && ex.Direction < len(dirNames) {
			exits = append(exits, dirNames[ex.Direction])
		}
	}

	if len(exits) == 0 {
		ch.Send("&DExits: none\n\r")
	} else {
		ch.Sendf("&DExits: %s\n\r", strings.Join(exits, " "))
	}
}

// DoScore implements the 'score' command.
func DoScore(ch *types.CharData, argument string) {
	// Race and class names
	raceName := "Unknown"
	if ch.Race >= 0 && ch.Race < len(WorldRef.Races) && WorldRef.Races[ch.Race] != nil {
		raceName = WorldRef.Races[ch.Race].Name
	}
	className := "Unknown"
	if ch.Class >= 0 && ch.Class < len(WorldRef.Classes) && WorldRef.Classes[ch.Class] != nil {
		className = WorldRef.Classes[ch.Class].WhoName
	}

	sexName := "Neutral"
	switch ch.Sex {
	case types.SEX_MALE:
		sexName = "Male"
	case types.SEX_FEMALE:
		sexName = "Female"
	}

	ch.Sendf("&W----- Score for %s -----&D\n\r", ch.Name)
	if ch.PCData != nil && ch.PCData.Title != "" {
		ch.Sendf("Title: %s %s\n\r", ch.Name, ch.PCData.Title)
	}
	ch.Sendf("Level: %-3d  Race: %-12s  Class: %-12s  Sex: %s\n\r",
		ch.Level, raceName, className, sexName)
	ch.Sendf("Hp: %d/%d  Mana: %d/%d  Move: %d/%d\n\r",
		ch.Hit, ch.MaxHit, ch.Mana, ch.MaxMana, ch.Move, ch.MaxMove)
	ch.Send("\n\r")
	ch.Sendf("Str: %-3d  Int: %-3d  Wis: %-3d  Dex: %-3d  Con: %-3d  Cha: %-3d  Lck: %-3d\n\r",
		ch.GetCurrStr(), ch.GetCurrInt(), ch.GetCurrWis(),
		ch.GetCurrDex(), ch.GetCurrCon(), ch.GetCurrCha(), ch.GetCurrLck())
	ch.Sendf("Hitroll: %-3d  Damroll: %-3d  Armor: %d\n\r",
		ch.Hitroll, ch.Damroll, ch.Armor)
	ch.Sendf("Alignment: %-5d  ", ch.Alignment)

	// Alignment descriptor
	switch {
	case ch.Alignment > 900:
		ch.Send("(Angelic)")
	case ch.Alignment > 700:
		ch.Send("(Saintly)")
	case ch.Alignment > 350:
		ch.Send("(Good)")
	case ch.Alignment > 100:
		ch.Send("(Kind)")
	case ch.Alignment > -100:
		ch.Send("(Neutral)")
	case ch.Alignment > -350:
		ch.Send("(Mean)")
	case ch.Alignment > -700:
		ch.Send("(Evil)")
	case ch.Alignment > -900:
		ch.Send("(Demonic)")
	default:
		ch.Send("(Satanic)")
	}
	ch.Send("\n\r")

	ch.Sendf("Gold: %d  Exp: %d\n\r", ch.Gold, ch.Exp)

	if ch.PCData != nil {
		ch.Sendf("Pkills: %d  Pdeaths: %d  Mkills: %d  Mdeaths: %d\n\r",
			ch.PCData.PKills, ch.PCData.PDeaths, ch.PCData.MKills, ch.PCData.MDeaths)
	}

	// Items carried
	ch.Sendf("Items: %d  Weight: %d\n\r", ch.CarryNumber, ch.CarryWeight)

	// Wimpy
	if ch.Wimpy > 0 {
		ch.Sendf("Wimpy set to %d hit points.\n\r", ch.Wimpy)
	}

	// Affects
	if len(ch.Affects) > 0 {
		ch.Send("\n\r&WActive affects:&D\n\r")
		for _, aff := range ch.Affects {
			name := "unknown"
			if aff.Type >= 0 && aff.Type < len(WorldRef.Skills) && WorldRef.Skills[aff.Type] != nil {
				name = WorldRef.Skills[aff.Type].Name
			}
			if aff.Duration >= 0 {
				ch.Sendf("  %-18s: %d rounds remaining\n\r", name, aff.Duration)
			} else {
				ch.Sendf("  %-18s: permanent\n\r", name)
			}
		}
	}
}

// DoWeather implements the 'weather' command.
func DoWeather(ch *types.CharData, argument string) {
	if ch.InRoom == nil {
		ch.Send("You can't see the weather from here.\n\r")
		return
	}
	if ch.InRoom.RoomFlags.IsSet(types.ROOM_INDOORS) {
		ch.Send("You can't see the weather indoors.\n\r")
		return
	}

	sky := "cloudless"
	if WorldRef != nil {
		switch WorldRef.TimeInfo.Sunlight {
		case types.SUN_DARK:
			sky = "dark and "
		case types.SUN_RISE:
			sky = "dawn and "
		case types.SUN_LIGHT:
			sky = "bright and "
		case types.SUN_SET:
			sky = "dusk and "
		}

		switch WorldRef.WeatherInfo.Sky {
		case types.SKY_CLOUDLESS:
			sky += "cloudless"
		case types.SKY_CLOUDY:
			sky += "cloudy"
		case types.SKY_RAINING:
			sky += "rainy"
		case types.SKY_LIGHTNING:
			sky += "stormy with lightning"
		}

		ch.Sendf("The sky is %s.\n\r", sky)
	} else {
		ch.Send("The sky is clear.\n\r")
	}
}

// DoWho implements the 'who' command.
func DoWho(ch *types.CharData, argument string) {
	if WorldRef == nil {
		ch.Send("Error: no world reference.\n\r")
		return
	}
	count := 0
	ch.Send("&W--- Players Online ---&D\n\r")
	for _, d := range WorldRef.Descriptors {
		if d.Connected != types.CON_PLAYING || d.Character == nil {
			continue
		}
		rch := d.Character
		if rch.IsNPC() {
			continue
		}
		count++
		title := ""
		if rch.PCData != nil {
			title = rch.PCData.Title
		}
		// AFK marker (C act_info.c:3686 and :4306 — "[AFK] " prefix before
		// the character's name when PLR_AFK is set on a non-NPC).
		afk := ""
		if rch.Act.IsSet(types.PLR_AFK) {
			afk = "[AFK] "
		}
		ch.Sendf("[%2d] %s%s %s\n\r", rch.Level, afk, rch.Name, title)
	}
	ch.Sendf("\n\r%d player%s online.\n\r", count, plural(count))
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// SaveFunc is set by main to allow saving players from the act package.
// This avoids a circular dependency between act and game.
// Written once at boot before the game loop starts; read only from the game loop goroutine. Safe without synchronization.
var SaveFunc func(ch *types.CharData)

// DoQuit implements the 'quit' command.
func DoQuit(ch *types.CharData, argument string) {
	if ch.Position == types.POS_FIGHTING {
		ch.Send("No way! You are fighting.\n\r")
		return
	}

	if !ch.IsNPC() && handler.GetTimer(ch, types.TIMER_RECENTFIGHT) > 0 {
		ch.Send("Your adrenaline is pumping too hard to quit now!\n\r")
		return
	}

	// Save the player before quitting
	if SaveFunc != nil {
		SaveFunc(ch)
		ch.Send("Your character has been saved.\n\r")
	}

	ch.Send("Your surroundings begin to fade as you slowly slip into a deep sleep ...\n\r")
	// The game loop will handle the actual disconnection
	if ch.Desc != nil {
		ch.Desc.Connected = -1 // signal to close
	}
}

// DoSay implements the 'say' command.
func DoSay(ch *types.CharData, argument string) {
	if argument == "" {
		ch.Send("Say what?\n\r")
		return
	}
	if ch.InRoom != nil && ch.InRoom.RoomFlags.IsSet(types.ROOM_SILENCE) {
		ch.Send("You can't do that here.\n\r")
		return
	}
	ch.Sendf("&CYou say '%s'&D\n\r", argument)
	if ch.InRoom != nil {
		for _, rch := range ch.InRoom.People {
			if rch != ch && rch.Desc != nil {
				rch.Sendf("&C%s says '%s'&D\n\r", ch.Name, translateFor(ch, rch, argument))
			}
		}
	}
	mudprog.OprogSpeechTrigger(ch, argument)
	mudprog.RprogSpeechTrigger(ch, argument)
}

// DoNorth etc. — movement commands
func DoNorth(ch *types.CharData, argument string)     { MoveChar(ch, types.DIR_NORTH) }
func DoEast(ch *types.CharData, argument string)      { MoveChar(ch, types.DIR_EAST) }
func DoSouth(ch *types.CharData, argument string)     { MoveChar(ch, types.DIR_SOUTH) }
func DoWest(ch *types.CharData, argument string)      { MoveChar(ch, types.DIR_WEST) }
func DoUp(ch *types.CharData, argument string)        { MoveChar(ch, types.DIR_UP) }
func DoDown(ch *types.CharData, argument string)      { MoveChar(ch, types.DIR_DOWN) }
func DoNortheast(ch *types.CharData, argument string) { MoveChar(ch, types.DIR_NORTHEAST) }
func DoNorthwest(ch *types.CharData, argument string) { MoveChar(ch, types.DIR_NORTHWEST) }
func DoSoutheast(ch *types.CharData, argument string) { MoveChar(ch, types.DIR_SOUTHEAST) }
func DoSouthwest(ch *types.CharData, argument string) { MoveChar(ch, types.DIR_SOUTHWEST) }

func MoveChar(ch *types.CharData, dir int) {
	if ch.InRoom == nil {
		ch.Send("You are nowhere!\n\r")
		return
	}

	if ch.Position < types.POS_STANDING {
		if ch.Position == types.POS_FIGHTING {
			ch.Send("You can't move while fighting! Try 'flee'.\n\r")
		} else {
			ch.Send("You need to stand up first.\n\r")
		}
		return
	}

	exit := ch.InRoom.GetExit(dir)
	if exit == nil || exit.ToRoom == nil {
		ch.Send("Alas, you cannot go that way.\n\r")
		return
	}

	if exit.ExitInfo&int(types.EX_CLOSED) != 0 {
		ch.Send("The door is closed.\n\r")
		return
	}

	if ch.IsNPC() && exit.ExitInfo&int(types.EX_NOMOB) != 0 {
		return
	}

	dest := exit.ToRoom
	if ch.IsNPC() && dest.RoomFlags.IsSet(types.ROOM_NO_MOB) {
		ch.Send("You can't go there.\n\r")
		return
	}
	if dir == types.DIR_DOWN && dest.RoomFlags.IsSet(types.ROOM_NOFLOOR) &&
		!ch.AffectedBy.IsSet(types.AFF_FLYING) {
		ch.Send("You can't fly.\n\r")
		return
	}
	if dest.RoomFlags.IsSet(types.ROOM_SOLITARY) && len(dest.People) > 0 {
		ch.Send("That room is too private for more than one person.\n\r")
		return
	}
	// ROOM_PRIVATE: C permits two occupants, third is refused.
	// See `src/handler.c:3264` (`room_is_private`).
	if dest.RoomFlags.IsSet(types.ROOM_PRIVATE) && len(dest.People) >= 2 {
		ch.Send("That room is private right now.\n\r")
		return
	}

	dirNames := []string{"north", "east", "south", "west", "up", "down",
		"northeast", "northwest", "southeast", "southwest"}
	revDir := []int{2, 3, 0, 1, 5, 4, 9, 8, 7, 6}

	// Leave message
	if ch.InRoom != nil {
		for _, rch := range ch.InRoom.People {
			if rch != ch && rch.Desc != nil {
				rch.Sendf("%s leaves %s.\n\r", ch.Name, dirNames[dir])
			}
		}
	}

	// Remove from old room
	oldRoom := ch.InRoom
	// Room-prog LEAVE fires BEFORE CharFromRoom so the prog still sees ch
	// in room.People. C calls it at the same point in move_char.
	mudprog.RprogLeaveTrigger(ch, oldRoom)
	removeFromRoom(ch, oldRoom)

	// Add to new room
	addToRoom(ch, dest)
	// Room-prog ENTER fires after ch is in the destination room, next to
	// where mob-prog GREET and obj-prog GREET would fire.
	mudprog.RprogEnterTrigger(ch)

	// Arrive message
	if dir < len(revDir) {
		revName := dirNames[revDir[dir]]
		for _, rch := range dest.People {
			if rch != ch && rch.Desc != nil {
				rch.Sendf("%s arrives from the %s.\n\r", ch.Name, revName)
			}
		}
	}

	// Auto-look
	DoLook(ch, "")

	// Mob-prog GREET / ALL_GREET fires on arrival for any non-fighting,
	// standing NPC in the destination room.
	mudprog.TrigGreet(ch)

	// Object-prog GREET fires after arrival for any floor-resident object
	// with a GREET prog.
	mudprog.OprogGreetTrigger(ch)

	// ROOM_DEATH: entering a death trap kills the character.
	if dest.RoomFlags.IsSet(types.ROOM_DEATH) && !ch.IsNPC() {
		applyRoomDeath(ch)
	}
}

// applyRoomDeath handles a player walking into a ROOM_DEATH trap.
// Matches minimum death behavior: announce death, reset to the temple
// with 1 HP/Mana/Move, position resting. Full corpse/XP-loss treatment
// would be a Phase-6 polish item.
func applyRoomDeath(ch *types.CharData) {
	if ch.InRoom != nil {
		for _, rch := range ch.InRoom.People {
			if rch != ch && rch.Desc != nil {
				rch.Sendf("%s screams and dies!\n\r", ch.Name)
			}
		}
	}
	ch.Send("You are dead! The gods have mercy and resurrect you.\n\r")
	ch.Hit = 1
	ch.Mana = 1
	ch.Move = 1
	ch.Position = types.POS_RESTING
	if WorldRef != nil {
		if temple := WorldRef.GetRoom(types.ROOM_VNUM_TEMPLE); temple != nil && ch.InRoom != nil {
			removeFromRoom(ch, ch.InRoom)
			addToRoom(ch, temple)
			DoLook(ch, "")
		}
	}
}

func removeFromRoom(ch *types.CharData, room *types.RoomIndexData) {
	if room == nil {
		return
	}
	for i, p := range room.People {
		if p == ch {
			room.People = append(room.People[:i], room.People[i+1:]...)
			break
		}
	}
	ch.WasInRoom = room
	ch.InRoom = nil
}

func addToRoom(ch *types.CharData, room *types.RoomIndexData) {
	ch.InRoom = room
	room.People = append(room.People, ch)
}

// DoCommands lists all available commands.
func DoCommands(ch *types.CharData, argument string) {
	ch.Send("Available commands: look, quit, say, score, who, commands\n\r")
	ch.Send("Movement: north, south, east, west, up, down, ne, nw, se, sw\n\r")
}

// DoHelp implements basic help.
func DoHelp(ch *types.CharData, argument string) {
	if argument == "" {
		ch.Send("Type 'commands' for a list of commands.\n\r")
		return
	}

	if WorldRef != nil {
		arg := strings.ToUpper(argument)
		for _, help := range WorldRef.Helps {
			if strings.Contains(strings.ToUpper(help.Keyword), arg) {
				ch.Sendf("&W%s&D\n\r%s\n\r", help.Keyword, help.Text)
				return
			}
		}
	}

	ch.Sendf("No help found for '%s'.\n\r", argument)
}

// DoInventory implements the 'inventory' command.
func DoInventory(ch *types.CharData, argument string) {
	ch.Send("You are carrying:\n\r")
	if len(ch.Carrying) == 0 {
		ch.Send("     Nothing.\n\r")
		return
	}
	for _, obj := range ch.Carrying {
		if obj.WearLoc == types.WEAR_NONE {
			ch.Sendf("     %s\n\r", obj.ShortDescr)
		}
	}
}

// DoEquipment implements the 'equipment' command.
func DoEquipment(ch *types.CharData, argument string) {
	ch.Send("You are using:\n\r")
	found := false
	for _, obj := range ch.Carrying {
		if obj.WearLoc != types.WEAR_NONE {
			ch.Sendf("  <%s>  %s\n\r", wearLocName(obj.WearLoc), obj.ShortDescr)
			found = true
		}
	}
	if !found {
		ch.Send("     Nothing.\n\r")
	}
}

func wearLocName(loc int) string {
	names := map[int]string{
		types.WEAR_LIGHT:      "used as light",
		types.WEAR_FINGER_L:   "left finger",
		types.WEAR_FINGER_R:   "right finger",
		types.WEAR_NECK_1:     "around neck",
		types.WEAR_NECK_2:     "around neck",
		types.WEAR_BODY:       "on body",
		types.WEAR_HEAD:       "on head",
		types.WEAR_LEGS:       "on legs",
		types.WEAR_FEET:       "on feet",
		types.WEAR_HANDS:      "on hands",
		types.WEAR_ARMS:       "on arms",
		types.WEAR_SHIELD:     "as shield",
		types.WEAR_ABOUT:      "about body",
		types.WEAR_WAIST:      "around waist",
		types.WEAR_WRIST_L:    "left wrist",
		types.WEAR_WRIST_R:    "right wrist",
		types.WEAR_WIELD:      "wielded",
		types.WEAR_HOLD:       "held",
		types.WEAR_DUAL_WIELD: "dual wielded",
		types.WEAR_EARS:       "on ears",
		types.WEAR_EYES:       "over eyes",
		types.WEAR_BACK:       "on back",
		types.WEAR_FACE:       "on face",
		types.WEAR_ANKLE_L:    "left ankle",
		types.WEAR_ANKLE_R:    "right ankle",
	}
	if n, ok := names[loc]; ok {
		return n
	}
	return fmt.Sprintf("wear loc %d", loc)
}
