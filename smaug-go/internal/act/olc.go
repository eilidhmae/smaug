package act

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// StartEditingFunc is set from main to avoid circular dependency with game package.
// Written once at boot before the game loop starts; read only from the game loop goroutine. Safe without synchronization.
var StartEditingFunc func(ch *types.CharData, text string)

// --- Room editing ---

// DoRedit implements the 'redit' command: edit the current room.
func DoRedit(ch *types.CharData, argument string) {
	if ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}
	if ch.InRoom == nil {
		ch.Send("You are not in a room.\n\r")
		return
	}

	arg, rest := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Redit what? (name, desc, sector, flags, exdesc, exit)\n\r")
		return
	}

	room := ch.InRoom

	switch strings.ToLower(arg) {
	case "name":
		if rest == "" {
			ch.Sendf("Current name: %s\n\r", room.Name)
			return
		}
		room.Name = rest
		ch.Sendf("Room name set to: %s\n\r", rest)

	case "desc":
		ch.Substate = types.SUB_ROOM_DESC
		if StartEditingFunc != nil {
			StartEditingFunc(ch, room.Description)
		}

	case "sector":
		if rest == "" {
			ch.Sendf("Current sector: %d\n\r", room.SectorType)
			return
		}
		val, err := strconv.Atoi(rest)
		if err != nil || val < 0 || val >= types.SECT_MAX {
			ch.Sendf("Sector must be 0 to %d.\n\r", types.SECT_MAX-1)
			return
		}
		room.SectorType = val
		ch.Sendf("Sector set to %d.\n\r", val)

	case "flags":
		if rest == "" {
			ch.Sendf("Room flags: %s\n\r", room.RoomFlags.String())
			return
		}
		val, err := strconv.Atoi(rest)
		if err != nil {
			ch.Send("Flags must be a number.\n\r")
			return
		}
		room.RoomFlags.Toggle(val)
		ch.Sendf("Room flag %d toggled.\n\r", val)

	case "exdesc":
		kw, desc := util.OneArgument(rest)
		if kw == "" {
			ch.Send("Usage: redit exdesc <keyword> <description>\n\r")
			return
		}
		room.ExtraDescr = append(room.ExtraDescr, &types.ExtraDescrData{
			Keyword:     kw,
			Description: desc + "\n\r",
		})
		ch.Sendf("Extra description '%s' added.\n\r", kw)

	case "exit":
		editExit(ch, rest)

	default:
		ch.Send("Redit what? (name, desc, sector, flags, exdesc, exit)\n\r")
	}
}

func editExit(ch *types.CharData, args string) {
	dir, rest := util.OneArgument(args)
	if dir == "" {
		ch.Send("Usage: redit exit <direction> <vnum|delete|key|flags>\n\r")
		return
	}

	dirNames := []string{"north", "east", "south", "west", "up", "down"}
	dirNum := -1
	for i, name := range dirNames {
		if strings.HasPrefix(name, strings.ToLower(dir)) {
			dirNum = i
			break
		}
	}
	if dirNum < 0 {
		ch.Send("Invalid direction.\n\r")
		return
	}

	sub, val := util.OneArgument(rest)
	sub = strings.ToLower(sub)

	if sub == "delete" {
		// Remove exit
		exits := make([]*types.ExitData, 0, len(ch.InRoom.Exits))
		for _, ex := range ch.InRoom.Exits {
			if ex.Direction != dirNum {
				exits = append(exits, ex)
			}
		}
		ch.InRoom.Exits = exits
		ch.Sendf("Exit %s deleted.\n\r", dirNames[dirNum])
		return
	}

	vnum, err := strconv.Atoi(sub)
	if err != nil {
		ch.Send("Usage: redit exit <dir> <vnum>\n\r")
		return
	}

	destRoom := WorldRef.GetRoom(vnum)
	if destRoom == nil {
		ch.Sendf("Room %d does not exist.\n\r", vnum)
		return
	}

	// Find or create exit
	var exit *types.ExitData
	for _, ex := range ch.InRoom.Exits {
		if ex.Direction == dirNum {
			exit = ex
			break
		}
	}
	if exit == nil {
		exit = &types.ExitData{Direction: dirNum}
		ch.InRoom.Exits = append(ch.InRoom.Exits, exit)
	}

	exit.ToRoom = destRoom
	exit.RVnum = vnum
	_ = val
	ch.Sendf("Exit %s set to room %d.\n\r", dirNames[dirNum], vnum)
}

// --- Object creation ---

// DoOcreate implements the 'ocreate' command: create a new object template.
func DoOcreate(ch *types.CharData, argument string) {
	if ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}
	arg, rest := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Usage: ocreate <vnum> [name]\n\r")
		return
	}

	vnum, err := strconv.Atoi(arg)
	if err != nil || vnum < 1 {
		ch.Send("Vnum must be a positive number.\n\r")
		return
	}

	if _, exists := WorldRef.ObjIndex[vnum]; exists {
		ch.Sendf("Object vnum %d already exists.\n\r", vnum)
		return
	}

	name := "new object"
	if rest != "" {
		name = rest
	}

	idx := &types.ObjIndexData{
		Vnum:       vnum,
		Name:       name,
		ShortDescr: name,
		Description: util.Capitalize(name) + " is here.",
		ItemType:   types.ITEM_TRASH,
		Level:      1,
		Weight:     1,
	}
	WorldRef.ObjIndex[vnum] = idx

	// Create an instance and give to builder
	obj := handler.CreateObject(WorldRef, idx, ch.Level)
	obj.CarriedBy = ch
	obj.WearLoc = types.WEAR_NONE
	ch.Carrying = append(ch.Carrying, obj)

	ch.Sendf("Object %d (%s) created.\n\r", vnum, name)
}

// --- Mob creation ---

// DoMcreate implements the 'mcreate' command: create a new mob template.
func DoMcreate(ch *types.CharData, argument string) {
	if ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}
	arg, rest := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Usage: mcreate <vnum> [name]\n\r")
		return
	}

	vnum, err := strconv.Atoi(arg)
	if err != nil || vnum < 1 {
		ch.Send("Vnum must be a positive number.\n\r")
		return
	}

	if _, exists := WorldRef.MobIndex[vnum]; exists {
		ch.Sendf("Mob vnum %d already exists.\n\r", vnum)
		return
	}

	name := "new mob"
	if rest != "" {
		name = rest
	}

	idx := &types.MobIndexData{
		Vnum:        vnum,
		PlayerName:  name,
		ShortDescr:  name,
		LongDescr:   util.Capitalize(name) + " is standing here.\n\r",
		Description: "",
		Level:       1,
		Position:    types.POS_STANDING,
		DefPosition: types.POS_STANDING,
	}
	idx.Act.Set(types.ACT_IS_NPC)
	WorldRef.MobIndex[vnum] = idx

	// Create an instance in the room
	mob := handler.CreateMobile(WorldRef, idx)
	if ch.InRoom != nil {
		handler.CharToRoom(mob, ch.InRoom)
	}

	ch.Sendf("Mob %d (%s) created.\n\r", vnum, name)
}

// --- Room digging ---

// DoRdig implements the 'rdig' command: create a room and link exits.
func DoRdig(ch *types.CharData, argument string) {
	if ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}
	arg, rest := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Usage: rdig <direction> [vnum]\n\r")
		return
	}

	dirNames := []string{"north", "east", "south", "west", "up", "down"}
	reverseDir := []int{2, 3, 0, 1, 5, 4} // opposite directions
	dirNum := -1
	for i, name := range dirNames {
		if strings.HasPrefix(name, strings.ToLower(arg)) {
			dirNum = i
			break
		}
	}
	if dirNum < 0 {
		ch.Send("Invalid direction.\n\r")
		return
	}

	// Determine vnum
	vnum := 0
	if rest != "" {
		var err error
		vnum, err = strconv.Atoi(strings.TrimSpace(rest))
		if err != nil {
			ch.Send("Vnum must be a number.\n\r")
			return
		}
	} else {
		// Auto-assign next available vnum
		for v := 1; v < types.MAX_VNUM; v++ {
			if WorldRef.GetRoom(v) == nil {
				vnum = v
				break
			}
		}
	}

	if WorldRef.GetRoom(vnum) != nil {
		ch.Sendf("Room %d already exists.\n\r", vnum)
		return
	}

	// Create the new room
	newRoom := &types.RoomIndexData{
		Vnum:       vnum,
		Name:       "A newly dug room",
		SectorType: types.SECT_INSIDE,
	}
	if ch.InRoom != nil && ch.InRoom.Area != nil {
		newRoom.Area = ch.InRoom.Area
	}
	WorldRef.Rooms[vnum] = newRoom

	// Create exit from current room to new room
	if ch.InRoom != nil {
		ch.InRoom.Exits = append(ch.InRoom.Exits, &types.ExitData{
			Direction: dirNum,
			ToRoom:    newRoom,
			RVnum:     vnum,
		})
		// Create reverse exit
		newRoom.Exits = append(newRoom.Exits, &types.ExitData{
			Direction: reverseDir[dirNum],
			ToRoom:    ch.InRoom,
			RVnum:     ch.InRoom.Vnum,
		})
	}

	ch.Sendf("Room %d created %s.\n\r", vnum, dirNames[dirNum])
}

// --- Listing commands ---

// DoRlist implements the 'rlist' command: list rooms in a vnum range.
func DoRlist(ch *types.CharData, argument string) {
	if ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}
	arg1, arg2 := util.OneArgument(argument)
	low := 0
	high := 0

	if arg1 == "" {
		// Default: show rooms in current area
		if ch.InRoom != nil && ch.InRoom.Area != nil {
			low = ch.InRoom.Area.LowRVnum
			high = ch.InRoom.Area.HiRVnum
		} else {
			ch.Send("Usage: rlist [low] [high]\n\r")
			return
		}
	} else {
		low, _ = strconv.Atoi(arg1)
		if arg2 != "" {
			high, _ = strconv.Atoi(arg2)
		} else {
			high = low + 100
		}
	}

	var sb strings.Builder
	for v := low; v <= high; v++ {
		room := WorldRef.GetRoom(v)
		if room != nil {
			fmt.Fprintf(&sb, "[%5d] %s\n\r", v, room.Name)
		}
	}

	if sb.Len() == 0 {
		ch.Send("No rooms in that range.\n\r")
		return
	}
	sendToPager(ch, sb.String())
}

// DoOlist implements the 'olist' command: list objects in a vnum range.
func DoOlist(ch *types.CharData, argument string) {
	if ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}
	arg1, arg2 := util.OneArgument(argument)
	low := 0
	high := 0

	if arg1 == "" {
		if ch.InRoom != nil && ch.InRoom.Area != nil {
			low = ch.InRoom.Area.LowOVnum
			high = ch.InRoom.Area.HiOVnum
		} else {
			ch.Send("Usage: olist [low] [high]\n\r")
			return
		}
	} else {
		low, _ = strconv.Atoi(arg1)
		if arg2 != "" {
			high, _ = strconv.Atoi(arg2)
		} else {
			high = low + 100
		}
	}

	var sb strings.Builder
	for v := low; v <= high; v++ {
		if idx, ok := WorldRef.ObjIndex[v]; ok {
			fmt.Fprintf(&sb, "[%5d] %s\n\r", v, idx.ShortDescr)
		}
	}

	if sb.Len() == 0 {
		ch.Send("No objects in that range.\n\r")
		return
	}
	sendToPager(ch, sb.String())
}

// DoMlist implements the 'mlist' command: list mobs in a vnum range.
func DoMlist(ch *types.CharData, argument string) {
	if ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}
	arg1, arg2 := util.OneArgument(argument)
	low := 0
	high := 0

	if arg1 == "" {
		if ch.InRoom != nil && ch.InRoom.Area != nil {
			low = ch.InRoom.Area.LowMVnum
			high = ch.InRoom.Area.HiMVnum
		} else {
			ch.Send("Usage: mlist [low] [high]\n\r")
			return
		}
	} else {
		low, _ = strconv.Atoi(arg1)
		if arg2 != "" {
			high, _ = strconv.Atoi(arg2)
		} else {
			high = low + 100
		}
	}

	var sb strings.Builder
	for v := low; v <= high; v++ {
		if idx, ok := WorldRef.MobIndex[v]; ok {
			fmt.Fprintf(&sb, "[%5d] %s\n\r", v, idx.ShortDescr)
		}
	}

	if sb.Len() == 0 {
		ch.Send("No mobs in that range.\n\r")
		return
	}
	sendToPager(ch, sb.String())
}

// --- Area Save ---

// DoSaveArea implements the 'savearea' command: save an area to disk.
func DoSaveArea(ch *types.CharData, argument string) {
	if ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}
	if ch.InRoom == nil || ch.InRoom.Area == nil {
		ch.Send("You are not in an area.\n\r")
		return
	}

	area := ch.InRoom.Area
	filename := area.Filename
	if filename == "" {
		ch.Send("This area has no filename.\n\r")
		return
	}

	path := filepath.Join(WorldRef.DataDir, "area", filename)
	f, err := os.Create(path)
	if err != nil {
		ch.Sendf("Error saving area: %v\n\r", err)
		return
	}
	defer f.Close()

	if err := persist.SaveArea(f, WorldRef, area); err != nil {
		ch.Sendf("Error writing area: %v\n\r", err)
		return
	}

	ch.Sendf("Area '%s' saved to %s.\n\r", area.Name, filename)
}
