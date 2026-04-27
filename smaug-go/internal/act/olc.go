package act

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// StartEditingFunc is set from main to avoid circular dependency with game package.
// Written once at boot before the game loop starts; read only from the game loop goroutine. Safe without synchronization.
var StartEditingFunc func(ch *types.CharData, text string)

// CopyBufferFunc and StopEditingFunc are the companion seams to
// StartEditingFunc — set from boot alongside it. EditorSave closures set by
// DoRedit (and future editor callers) reach game.CopyBuffer / game.StopEditing
// through these function variables to avoid the otherwise-circular
// game -> act -> game import path.
var (
	CopyBufferFunc  func(ch *types.CharData) string
	StopEditingFunc func(ch *types.CharData)
)

// ReditDispMenuFunc opens the interactive redit menu for a descriptor.
// Wired from boot (game.ReditDispMenu); act cannot import game directly,
// so the seam breaks the cycle the same way StartEditingFunc does.
// Called from DoRedit(ch, "") — the no-arg path enters the menu.
// Plan plan-phase6-olc-redit.md §G3.
var ReditDispMenuFunc func(d *types.DescriptorData)

// OeditDispMenuFunc is the oedit counterpart to ReditDispMenuFunc. Wired
// from boot (game.OeditDispMenu). Wave 1 only declares the seam — Wave 2+
// will populate game.OeditDispMenu and the DoOedit no-arg path will call
// this function to enter the CON_OEDIT menu.
// Plan plan-phase6-olc-oedit.md §G3.
var OeditDispMenuFunc func(d *types.DescriptorData)

// RenamePlayerFileFunc is the seam to persist.RenamePlayerFile, wired
// at boot. Tests can override to record calls / inject errors without
// touching disk. Set once at boot before the game loop starts; read
// only from the game loop goroutine. Safe without synchronization.
//
// Plan: plan-phase6-quickwins-blank-pcrename.md §D4b / §G6.
var RenamePlayerFileFunc func(oldName, newName string) error

// MeditDispMenuFunc is the medit counterpart to ReditDispMenuFunc /
// OeditDispMenuFunc. Wired from boot (game.MeditDispMenu) in Wave 2 — Wave
// 1 declares the seam but deliberately does NOT wire it in boot.go because
// the Wave 1 stub meditParse does not call it (so a nil seam is safe).
// Wave 2 lands the real game.MeditDispMenu renderer and adds the boot
// assignment. The eventual DoMedit no-arg menu-entry path will call this
// function to enter the CON_MEDIT menu (NPC or PC branch selected inside
// the renderer via victim.IsNPC()).
// Plan plan-phase6-olc-medit.md §G4 / §G14.
var MeditDispMenuFunc func(d *types.DescriptorData)

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
	room := ch.InRoom
	if arg == "" {
		// No subcommand → enter the interactive menu. Builders who
		// prefer the flat path keep using `redit name foo` etc.
		// Plan plan-phase6-olc-redit.md §G3 / §A5.
		if ch.Desc == nil {
			ch.Send("No descriptor.\n\r")
			return
		}
		ch.Desc.Olc = &types.OlcData{
			Mode:   types.REDIT_MAIN_MENU,
			Vnum:   room.Vnum,
			Target: room,
		}
		ch.Desc.Connected = int(types.CON_REDIT)
		if ReditDispMenuFunc != nil {
			ReditDispMenuFunc(ch.Desc)
		}
		return
	}

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
		targetRoom := room
		ch.EditorSave = func(c *types.CharData) {
			if CopyBufferFunc != nil {
				targetRoom.Description = CopyBufferFunc(c)
			}
			if StopEditingFunc != nil {
				StopEditingFunc(c)
			}
			// Trailing newline so the post-CON_PLAYING prompt lands on
			// a fresh line (plan § G4 prompt caveat).
			c.Send("\n\r")
		}
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

	case "ed":
		// Launch string editor seeded with existing description (if the keyword
		// already matches) or empty for a new extra-descr.
		kw := strings.TrimSpace(rest)
		if kw == "" {
			ch.Send("Usage: redit ed <keyword>\n\r")
			return
		}
		var existing *types.ExtraDescrData
		for _, ed := range room.ExtraDescr {
			if strings.EqualFold(ed.Keyword, kw) {
				existing = ed
				break
			}
		}
		if existing == nil {
			existing = &types.ExtraDescrData{Keyword: kw}
			room.ExtraDescr = append(room.ExtraDescr, existing)
		}
		ch.Substate = types.SUB_ROOM_EXTRA
		ch.InterEditing = kw
		targetExtra := existing
		ch.EditorSave = func(c *types.CharData) {
			if CopyBufferFunc != nil {
				targetExtra.Description = CopyBufferFunc(c)
			}
			if StopEditingFunc != nil {
				StopEditingFunc(c)
			}
			c.Send("\n\r")
		}
		if StartEditingFunc != nil {
			StartEditingFunc(ch, existing.Description)
		} else {
			ch.Sendf("Extra description '%s' ready for editing.\n\r", kw)
		}

	case "rmed":
		kw := strings.TrimSpace(rest)
		if kw == "" {
			ch.Send("Usage: redit rmed <keyword>\n\r")
			return
		}
		before := len(room.ExtraDescr)
		filtered := make([]*types.ExtraDescrData, 0, before)
		for _, ed := range room.ExtraDescr {
			if !strings.EqualFold(ed.Keyword, kw) {
				filtered = append(filtered, ed)
			}
		}
		if len(filtered) == before {
			ch.Sendf("No extra description matches '%s'.\n\r", kw)
			return
		}
		room.ExtraDescr = filtered
		ch.Sendf("Extra description '%s' removed.\n\r", kw)

	case "exit":
		editExit(ch, rest)

	case "bexit":
		editBidirExit(ch, rest)

	case "exflags":
		editExitFlags(ch, rest)

	case "exname":
		editExitKeyword(ch, rest)

	case "exkey":
		editExitKey(ch, rest)

	case "teledelay":
		val, err := strconv.Atoi(strings.TrimSpace(rest))
		if err != nil || val < 0 {
			ch.Send("Teledelay must be a non-negative number.\n\r")
			return
		}
		room.TeleDelay = val
		ch.Sendf("Teledelay set to %d.\n\r", val)

	case "televnum":
		val, err := strconv.Atoi(strings.TrimSpace(rest))
		if err != nil || val < 0 {
			ch.Send("Televnum must be a non-negative number.\n\r")
			return
		}
		room.TeleVnum = val
		ch.Sendf("Televnum set to %d.\n\r", val)

	case "tunnel":
		val, err := strconv.Atoi(strings.TrimSpace(rest))
		if err != nil || val < 0 {
			ch.Send("Tunnel must be a non-negative number.\n\r")
			return
		}
		room.Tunnel = val
		ch.Sendf("Tunnel set to %d.\n\r", val)

	case "rlist":
		DoRlist(ch, rest)

	default:
		ch.Send("Redit what? (name, desc, sector, flags, exdesc, ed, rmed, exit, bexit, exflags, exname, exkey, teledelay, televnum, tunnel, rlist)\n\r")
	}
}

// parseDirection parses a direction name prefix (north/east/south/west/up/down).
// Returns -1 if not found.
func parseDirection(s string) int {
	dirNames := []string{"north", "east", "south", "west", "up", "down"}
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return -1
	}
	for i, name := range dirNames {
		if strings.HasPrefix(name, s) {
			return i
		}
	}
	return -1
}

// reverseDir returns the opposite direction constant.
func reverseDir(dir int) int {
	rev := []int{2, 3, 0, 1, 5, 4}
	if dir < 0 || dir >= len(rev) {
		return -1
	}
	return rev[dir]
}

// editBidirExit creates a bidirectional exit: dig to destination + auto-add reverse.
// Differs from rdig: does not create a new room, links to an EXISTING room.
func editBidirExit(ch *types.CharData, args string) {
	dir, rest := util.OneArgument(args)
	dirNum := parseDirection(dir)
	if dirNum < 0 {
		ch.Send("Usage: redit bexit <direction> <vnum>\n\r")
		return
	}
	vnum, err := strconv.Atoi(strings.TrimSpace(rest))
	if err != nil || vnum <= 0 {
		ch.Send("Usage: redit bexit <direction> <vnum>\n\r")
		return
	}
	dest := WorldRef.GetRoom(vnum)
	if dest == nil {
		ch.Sendf("Room %d does not exist.\n\r", vnum)
		return
	}
	room := ch.InRoom

	// Forward exit
	var fwd *types.ExitData
	for _, ex := range room.Exits {
		if ex.Direction == dirNum {
			fwd = ex
			break
		}
	}
	if fwd == nil {
		fwd = &types.ExitData{Direction: dirNum}
		room.Exits = append(room.Exits, fwd)
	}
	fwd.ToRoom = dest
	fwd.Vnum = vnum
	fwd.RVnum = vnum

	// Reverse exit
	rev := reverseDir(dirNum)
	if rev >= 0 {
		var back *types.ExitData
		for _, ex := range dest.Exits {
			if ex.Direction == rev {
				back = ex
				break
			}
		}
		if back == nil {
			back = &types.ExitData{Direction: rev}
			dest.Exits = append(dest.Exits, back)
		}
		back.ToRoom = room
		back.Vnum = room.Vnum
		back.RVnum = room.Vnum
	}

	ch.Sendf("Bidirectional exit created to room %d.\n\r", vnum)
}

// exitFlagBits maps flag names to EX_* bits.
var exitFlagBits = map[string]uint32{
	"isdoor":     types.EX_ISDOOR,
	"closed":     types.EX_CLOSED,
	"locked":     types.EX_LOCKED,
	"secret":     types.EX_SECRET,
	"pickproof":  types.EX_PICKPROOF,
	"hidden":     types.EX_HIDDEN,
	"nomob":      types.EX_NOMOB,
	"nopassdoor": types.EX_NOPASSDOOR,
	"nopass":     types.EX_NOPASSDOOR,
	"bashed":     types.EX_BASHED,
	"bashproof":  types.EX_BASHPROOF,
}

func editExitFlags(ch *types.CharData, args string) {
	dir, rest := util.OneArgument(args)
	dirNum := parseDirection(dir)
	if dirNum < 0 {
		ch.Send("Usage: redit exflags <direction> <flag>\n\r")
		return
	}
	flagName := strings.ToLower(strings.TrimSpace(rest))
	bit, ok := exitFlagBits[flagName]
	if !ok {
		ch.Send("Valid flags: isdoor, closed, locked, secret, pickproof, hidden, nomob, nopassdoor, bashed, bashproof\n\r")
		return
	}
	ex := ch.InRoom.GetExit(dirNum)
	if ex == nil {
		ch.Sendf("No exit %s.\n\r", dir)
		return
	}
	ex.ExitInfo ^= int(bit)
	ch.Sendf("Exit flag '%s' toggled.\n\r", flagName)
}

func editExitKeyword(ch *types.CharData, args string) {
	dir, rest := util.OneArgument(args)
	dirNum := parseDirection(dir)
	if dirNum < 0 {
		ch.Send("Usage: redit exname <direction> <keyword>\n\r")
		return
	}
	ex := ch.InRoom.GetExit(dirNum)
	if ex == nil {
		ch.Sendf("No exit %s.\n\r", dir)
		return
	}
	ex.Keyword = strings.TrimSpace(rest)
	ch.Sendf("Exit keyword set to '%s'.\n\r", ex.Keyword)
}

func editExitKey(ch *types.CharData, args string) {
	dir, rest := util.OneArgument(args)
	dirNum := parseDirection(dir)
	if dirNum < 0 {
		ch.Send("Usage: redit exkey <direction> <vnum>\n\r")
		return
	}
	vnum, err := strconv.Atoi(strings.TrimSpace(rest))
	if err != nil {
		ch.Send("Usage: redit exkey <direction> <vnum>\n\r")
		return
	}
	ex := ch.InRoom.GetExit(dirNum)
	if ex == nil {
		ch.Sendf("No exit %s.\n\r", dir)
		return
	}
	ex.Key = vnum
	ch.Sendf("Exit key vnum set to %d.\n\r", vnum)
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
		Vnum:        vnum,
		Name:        name,
		ShortDescr:  name,
		Description: util.Capitalize(name) + " is here.",
		ItemType:    types.ITEM_TRASH,
		Level:       1,
		Weight:      1,
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

// DoSaveArea implements the 'savearea' command: save the current room's area
// to disk. The mechanical save (path validation, tmp-file write, atomic
// install) lives in writeAreaToDisk (olc_area_save.go) and is shared with
// DoFoldarea. Plan plan-phase6-foldarea.md §G2.
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
	if area.Filename == "" {
		ch.Send("This area has no filename.\n\r")
		return
	}

	if err := writeAreaToDisk(area); err != nil {
		// Path-containment / filename rejection surfaces as the C-faithful
		// "Invalid area filename" message; other failures get the underlying
		// error text so a mid-save disk failure is still actionable.
		if err.Error() == "invalid area filename" {
			ch.Send("Invalid area filename.\n\r")
			return
		}
		ch.Sendf("Error saving area: %v\n\r", err)
		return
	}

	ch.Sendf("Area '%s' saved to %s.\n\r", area.Name, area.Filename)
}
