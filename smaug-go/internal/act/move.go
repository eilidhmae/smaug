package act

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// findDoor finds an exit in the character's room by keyword or direction name.
func findDoor(ch *types.CharData, arg string) *types.ExitData {
	if ch.InRoom == nil {
		return nil
	}

	// Try direction name first
	dirNames := []string{"north", "east", "south", "west", "up", "down",
		"northeast", "northwest", "southeast", "southwest"}
	for i, name := range dirNames {
		if strings.HasPrefix(name, strings.ToLower(arg)) {
			return ch.InRoom.GetExit(i)
		}
	}

	// Try keyword match
	for _, exit := range ch.InRoom.Exits {
		if exit.Keyword != "" && util.IsName(arg, exit.Keyword) {
			return exit
		}
	}

	return nil
}

// hasKey returns true if the character is carrying an object with the given vnum.
func hasKey(ch *types.CharData, keyVnum int) bool {
	if keyVnum <= 0 {
		return false
	}
	for _, obj := range ch.Carrying {
		if obj.IndexData != nil && obj.IndexData.Vnum == keyVnum {
			return true
		}
	}
	return false
}

// DoOpen implements the 'open' command for doors.
func DoOpen(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Open what?\n\r")
		return
	}

	// Check for container in inventory/room first
	obj := handler.GetObjHere(ch, arg)
	if obj != nil && obj.ItemType == types.ITEM_CONTAINER {
		// Container open logic (simplified)
		if obj.Value[1]&int(types.CONT_CLOSED) == 0 {
			ch.Send("It's already open.\n\r")
			return
		}
		if obj.Value[1]&int(types.CONT_LOCKED) != 0 {
			ch.Send("It's locked.\n\r")
			return
		}
		obj.Value[1] &^= int(types.CONT_CLOSED)
		ch.Sendf("You open %s.\n\r", obj.ShortDescr)
		return
	}

	// Try door
	exit := findDoor(ch, arg)
	if exit == nil {
		ch.Send("You see no door there.\n\r")
		return
	}

	if exit.ExitInfo&int(types.EX_ISDOOR) == 0 {
		ch.Send("You can't do that.\n\r")
		return
	}

	if exit.ExitInfo&int(types.EX_CLOSED) == 0 {
		ch.Send("It's already open.\n\r")
		return
	}

	if exit.ExitInfo&int(types.EX_LOCKED) != 0 {
		ch.Send("It's locked.\n\r")
		return
	}

	exit.ExitInfo &^= int(types.EX_CLOSED)
	ch.Send("You open the door.\n\r")

	// Notify room
	for _, rch := range ch.InRoom.People {
		if rch != ch && rch.Desc != nil {
			rch.Sendf("%s opens a door.\n\r", ch.Name)
		}
	}
}

// DoClose implements the 'close' command for doors.
func DoClose(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Close what?\n\r")
		return
	}

	exit := findDoor(ch, arg)
	if exit == nil {
		ch.Send("You see no door there.\n\r")
		return
	}

	if exit.ExitInfo&int(types.EX_ISDOOR) == 0 {
		ch.Send("You can't do that.\n\r")
		return
	}

	if exit.ExitInfo&int(types.EX_CLOSED) != 0 {
		ch.Send("It's already closed.\n\r")
		return
	}

	exit.ExitInfo |= int(types.EX_CLOSED)
	ch.Send("You close the door.\n\r")

	for _, rch := range ch.InRoom.People {
		if rch != ch && rch.Desc != nil {
			rch.Sendf("%s closes a door.\n\r", ch.Name)
		}
	}
}

// DoUnlock implements the 'unlock' command for doors.
func DoUnlock(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Unlock what?\n\r")
		return
	}

	exit := findDoor(ch, arg)
	if exit == nil {
		ch.Send("You see no door there.\n\r")
		return
	}

	if exit.ExitInfo&int(types.EX_ISDOOR) == 0 {
		ch.Send("You can't do that.\n\r")
		return
	}

	if exit.ExitInfo&int(types.EX_LOCKED) == 0 {
		ch.Send("It's already unlocked.\n\r")
		return
	}

	if !hasKey(ch, exit.Key) {
		ch.Send("You lack the key.\n\r")
		return
	}

	exit.ExitInfo &^= int(types.EX_LOCKED)
	ch.Send("You unlock the door.\n\r")
}

// DoLock implements the 'lock' command for doors.
func DoLock(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Lock what?\n\r")
		return
	}

	exit := findDoor(ch, arg)
	if exit == nil {
		ch.Send("You see no door there.\n\r")
		return
	}

	if exit.ExitInfo&int(types.EX_ISDOOR) == 0 {
		ch.Send("You can't do that.\n\r")
		return
	}

	if exit.ExitInfo&int(types.EX_CLOSED) == 0 {
		ch.Send("It's not closed.\n\r")
		return
	}

	if exit.ExitInfo&int(types.EX_LOCKED) != 0 {
		ch.Send("It's already locked.\n\r")
		return
	}

	if !hasKey(ch, exit.Key) {
		ch.Send("You lack the key.\n\r")
		return
	}

	exit.ExitInfo |= int(types.EX_LOCKED)
	ch.Send("You lock the door.\n\r")
}
