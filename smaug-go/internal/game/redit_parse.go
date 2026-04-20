// Package game — interactive room editor (CON_REDIT substate).
//
// This file ports src/oredit.c redit_parse at :541-1035 plus the
// cleanup_olc helper. The pulse loop dispatches input addressed to a
// descriptor in CON_REDIT directly here (see loop.go processInput); the
// nanny is NOT invoked for menu input. Mirrors C src/smaug.c:1641-1652.
//
// Plan: plan-phase6-olc-redit.md §G5-G10.
package game

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// worldRef is the game-package world pointer used by the menu-driven
// editor to resolve exit target rooms. Wired from boot alongside the
// existing seams (act.WorldRef, mudprog.WorldRef). Not exported because
// game callers pass *world.World explicitly where they can.
var worldRef *world.World

// SetWorldRef is the boot-time seam. Called from Boot after world load.
// Separate from NewGameLoop so tests can set a custom world for
// redit_parse tests without standing up the full loop.
func SetWorldRef(w *world.World) {
	worldRef = w
}

// reditParse is the top-level CON_REDIT dispatcher. Keyed on
// d.Olc.Mode, it mutates the room under edit and either returns (stay in
// same mode) or transitions to a new mode + redisplays the appropriate
// menu.
//
// The caller (processInput) has already validated d.Connected ==
// CON_REDIT. If d.Olc is nil (should not happen — stale state bug
// elsewhere), cleanup and drop back to CON_PLAYING defensively.
func reditParse(d *types.DescriptorData, arg string) {
	if d == nil {
		return
	}
	if d.Olc == nil || d.Character == nil {
		// Defensive: impossible state; restore playable.
		d.Connected = int(types.CON_PLAYING)
		return
	}
	room, ok := d.Olc.Target.(*types.RoomIndexData)
	if !ok || room == nil {
		// Target lost (should not happen). Restore playable.
		cleanupOlc(d)
		return
	}

	switch d.Olc.Mode {
	case types.REDIT_MAIN_MENU:
		reditHandleMainMenu(d, room, arg)

	case types.REDIT_NAME:
		if strings.TrimSpace(arg) == "" {
			ReditDispMenu(d)
			return
		}
		room.Name = util.SmashTilde(arg)
		olcLog(d, "ROOM", "Changed name to %s", room.Name)
		ReditDispMenu(d)

	case types.REDIT_FLAGS:
		reditHandleFlags(d, room, arg)

	case types.REDIT_SECTOR:
		reditHandleSector(d, room, arg)

	case types.REDIT_TUNNEL:
		n := parseIntDefault(arg, 0)
		room.Tunnel = uRange(0, n, 1000)
		olcLog(d, "ROOM", "Changed tunnel amount to %d", room.Tunnel)
		ReditDispMenu(d)

	case types.REDIT_TELEDELAY:
		room.TeleDelay = parseIntDefault(arg, 0)
		olcLog(d, "ROOM", "Changed teleportation delay to %d", room.TeleDelay)
		ReditDispMenu(d)

	case types.REDIT_TELEVNUM:
		n := parseIntDefault(arg, 0)
		room.TeleVnum = uRange(1, n, types.MAX_VNUM)
		olcLog(d, "ROOM", "Changed teleportation vnum to %d", room.TeleVnum)
		ReditDispMenu(d)

	case types.REDIT_EXIT_MENU:
		reditHandleExitMenu(d, room, arg)

	case types.REDIT_EXIT_EDIT:
		reditHandleExitEdit(d, room, arg)

	case types.REDIT_EXIT_ADD:
		reditHandleExitAdd(d, room, arg)

	case types.REDIT_EXIT_ADD_VNUM:
		reditHandleExitAddVnum(d, room, arg)

	case types.REDIT_EXIT_DELETE:
		reditHandleExitDelete(d, room, arg)

	case types.REDIT_EXIT_VNUM:
		reditHandleExitVnum(d, arg)

	case types.REDIT_EXIT_KEY:
		reditHandleExitKey(d, arg)

	case types.REDIT_EXIT_KEYWORD:
		reditHandleExitKeyword(d, arg)

	case types.REDIT_EXIT_DESC:
		reditHandleExitDesc(d, arg)

	case types.REDIT_EXIT_FLAGS:
		reditHandleExitFlags(d, arg)

	case types.REDIT_EXTRADESC_MENU:
		reditHandleExtradescMenu(d, room, arg)

	case types.REDIT_EXTRADESC_CHOICE:
		reditHandleExtradescChoice(d, room, arg)

	case types.REDIT_EXTRADESC_KEY:
		reditHandleExtradescKey(d, arg)

	case types.REDIT_EXTRADESC_DELETE:
		reditHandleExtradescDelete(d, room, arg)

	default:
		util.Bug("reditParse: unhandled mode %d", d.Olc.Mode)
		ReditDispMenu(d)
	}
}

// reditHandleMainMenu handles input at the top-level menu.
// Mirrors C redit_parse case REDIT_MAIN_MENU at src/oredit.c:575-633.
func reditHandleMainMenu(d *types.DescriptorData, room *types.RoomIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		ReditDispMenu(d)
		return
	}
	switch strings.ToLower(arg)[:1] {
	case "q":
		cleanupOlc(d)
		d.WriteToBuffer("Exiting editor.\n\r")
		return
	case "1":
		d.WriteToBuffer("Enter room name:\n\r| ")
		d.Olc.Mode = types.REDIT_NAME
	case "2":
		// Route through the line editor, then re-enter CON_REDIT after
		// /s fires. Reuses EditorSave closure pattern from the flat
		// DoRedit path (Tier 12).
		ch := d.Character
		ch.Substate = types.SUB_ROOM_DESC
		targetRoom := room
		ch.EditorSave = func(c *types.CharData) {
			if c == nil || c.Desc == nil {
				return
			}
			targetRoom.Description = CopyBuffer(c)
			StopEditing(c)
			// StopEditing just set Connected = CON_PLAYING. Re-set
			// CON_REDIT so the next line returns to the menu handler.
			c.Desc.Connected = int(types.CON_REDIT)
			olcLog(c.Desc, "ROOM", "Edited room description")
			ReditDispMenu(c.Desc)
		}
		d.WriteToBuffer("Enter room description:\n\r")
		if room.Description == "" {
			room.Description = ""
		}
		StartEditing(ch, room.Description)
	case "3":
		reditDispFlagMenu(d)
	case "4":
		reditDispSectorMenu(d)
	case "5":
		d.WriteToBuffer("How many people can fit in the room? ")
		d.Olc.Mode = types.REDIT_TUNNEL
	case "6":
		d.WriteToBuffer("How long before people are teleported out? ")
		d.Olc.Mode = types.REDIT_TELEDELAY
	case "7":
		d.WriteToBuffer("Where are they teleported to? ")
		d.Olc.Mode = types.REDIT_TELEVNUM
	case "a":
		reditDispExitMenu(d)
	case "b":
		reditDispExtradescMenu(d)
	default:
		d.WriteToBuffer("Invalid choice!\n\r")
		ReditDispMenu(d)
	}
}

// reditHandleFlags toggles one or more room-flag bits.
// Mirrors C redit_parse case REDIT_FLAGS at src/oredit.c:647-683.
// Two input modes:
//   - integer 0 → exit back to main menu
//   - integer 1..32 → toggle that 1-indexed bit
//   - word list → per word, look up via getRoomFlagBit and toggle
func reditHandleFlags(d *types.DescriptorData, room *types.RoomIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		reditDispFlagMenu(d)
		return
	}
	// Try integer first (whole-input is a number).
	if n, err := strconv.Atoi(arg); err == nil {
		if n == 0 {
			ReditDispMenu(d)
			return
		}
		if n < 1 || n > 32 {
			d.WriteToBuffer("Invalid flag, try again: ")
			return
		}
		bit := n - 1
		room.RoomFlags.Toggle(bit)
		action := "Removed"
		if room.RoomFlags.IsSet(bit) {
			action = "Added"
		}
		label := "(unnamed)"
		if bit < len(roomFlagNames) {
			label = roomFlagNames[bit]
		}
		olcLog(d, "ROOM", "%s the room flag %s", action, label)
		reditDispFlagMenu(d)
		return
	}
	// Word list: toggle each. Unknown words are skipped with a log note.
	words := strings.Fields(arg)
	for _, w := range words {
		bit := getRoomFlagBit(w)
		if bit < 0 {
			continue
		}
		room.RoomFlags.Toggle(bit)
		action := "Removed"
		if room.RoomFlags.IsSet(bit) {
			action = "Added"
		}
		olcLog(d, "ROOM", "%s the room flag %s", action, roomFlagNames[bit])
	}
	reditDispFlagMenu(d)
}

// reditHandleSector sets room.SectorType to a 0..SECT_MAX-1 integer.
// Rejects SECT_DUNNO and out-of-range values by redisplaying the menu.
// Mirrors C REDIT_SECTOR at src/oredit.c:685-696.
func reditHandleSector(d *types.DescriptorData, room *types.RoomIndexData, arg string) {
	n, err := strconv.Atoi(strings.TrimSpace(arg))
	if err != nil || n < 0 || n >= types.SECT_MAX || n == types.SECT_DUNNO {
		d.WriteToBuffer("Invalid choice!\n\r")
		reditDispSectorMenu(d)
		return
	}
	room.SectorType = n
	olcLog(d, "ROOM", "Changed sector to %s", sectorKeywords[n])
	ReditDispMenu(d)
}

// reditHandleExitMenu interprets input at the exit-list menu.
// Mirrors C REDIT_EXIT_MENU at src/oredit.c:716-742.
func reditHandleExitMenu(d *types.DescriptorData, room *types.RoomIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		reditDispExitMenu(d)
		return
	}
	switch strings.ToLower(arg)[:1] {
	case "a":
		d.Olc.Mode = types.REDIT_EXIT_ADD
		reditDispExitDirs(d)
	case "r":
		d.Olc.Mode = types.REDIT_EXIT_DELETE
		d.WriteToBuffer("Delete which exit? ")
	case "q":
		d.Olc.Spare = nil
		ReditDispMenu(d)
	default:
		n, err := strconv.Atoi(arg)
		if err != nil || n < 1 || n > len(room.Exits) {
			reditDispExitMenu(d)
			return
		}
		pexit := room.Exits[n-1]
		d.Olc.Spare = pexit
		reditDispExitEdit(d)
	}
}

// reditHandleExitEdit handles the 1/2/3/4/5/6/Q selection on a specific
// exit. Mirrors C REDIT_EXIT_EDIT at src/oredit.c:744-778.
func reditHandleExitEdit(d *types.DescriptorData, _ *types.RoomIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		reditDispExitEdit(d)
		return
	}
	switch strings.ToLower(arg)[:1] {
	case "q":
		d.Olc.Spare = nil
		reditDispExitMenu(d)
		return
	case "1":
		d.WriteToBuffer("This option can only be changed by remaking the exit.\n\r")
		reditDispExitEdit(d)
	case "2":
		d.Olc.Mode = types.REDIT_EXIT_VNUM
		d.WriteToBuffer("Which room does this exit go to? ")
	case "3":
		d.Olc.Mode = types.REDIT_EXIT_KEY
		d.WriteToBuffer("What is the vnum of the key to this exit? ")
	case "4":
		d.Olc.Mode = types.REDIT_EXIT_KEYWORD
		d.WriteToBuffer("What is the keyword to this exit? ")
	case "5":
		reditDispExitFlagMenu(d)
	case "6":
		d.Olc.Mode = types.REDIT_EXIT_DESC
		d.WriteToBuffer("Description:\n\r] ")
	default:
		reditDispExitEdit(d)
	}
}

// reditHandleExitAdd validates the direction input then prompts for the
// to-vnum. Mirrors C REDIT_EXIT_ADD at src/oredit.c:792-817.
func reditHandleExitAdd(d *types.DescriptorData, room *types.RoomIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		reditDispExitDirs(d)
		return
	}
	var dir int
	if n, err := strconv.Atoi(arg); err == nil {
		if n < types.DIR_NORTH || n > types.DIR_SOMEWHERE {
			d.WriteToBuffer("Invalid direction, try again: ")
			return
		}
		dir = n
	} else {
		dir = -1
		for i := 0; i <= types.DIR_SOMEWHERE; i++ {
			if strings.HasPrefix(directionName(i), strings.ToLower(arg)) {
				dir = i
				break
			}
		}
		if dir < 0 {
			d.WriteToBuffer("Invalid direction, try again: ")
			return
		}
		// Check existing exit in that direction.
		if room.GetExit(dir) != nil {
			d.WriteToBuffer("An exit in that direction already exists.\n\r")
			reditDispExitMenu(d)
			return
		}
	}
	d.Olc.TempNum = dir
	d.Olc.Mode = types.REDIT_EXIT_ADD_VNUM
	d.WriteToBuffer("Which room does this exit go to? ")
}

// reditHandleExitAddVnum validates the destination vnum, makes the exit,
// and transitions to REDIT_EXIT_EDIT so the builder can tune it.
// Mirrors C REDIT_EXIT_ADD_VNUM at src/oredit.c:819-840.
func reditHandleExitAddVnum(d *types.DescriptorData, room *types.RoomIndexData, arg string) {
	n, err := strconv.Atoi(strings.TrimSpace(arg))
	if err != nil {
		d.WriteToBuffer("Non-existant room.\n\r")
		d.Olc.Mode = types.REDIT_EXIT_MENU
		reditDispExitMenu(d)
		return
	}
	target := worldRoomLookup(n)
	if target == nil {
		d.WriteToBuffer("Non-existant room.\n\r")
		d.Olc.Mode = types.REDIT_EXIT_MENU
		reditDispExitMenu(d)
		return
	}
	pexit := &types.ExitData{
		Direction:   d.Olc.TempNum,
		ToRoom:      target,
		Vnum:        n,
		RVnum:       n,
		Keyword:     "",
		Description: "",
		Key:         -1,
		ExitInfo:    0,
	}
	room.Exits = append(room.Exits, pexit)
	d.Olc.Spare = pexit
	olcLog(d, "ROOM", "Added %s exit to %d", directionName(pexit.Direction), n)
	d.Olc.Mode = types.REDIT_EXIT_EDIT
	reditDispExitEdit(d)
}

// reditHandleExitDelete removes an exit by 1-indexed slot number.
// Mirrors C REDIT_EXIT_DELETE at src/oredit.c:842-858.
func reditHandleExitDelete(d *types.DescriptorData, room *types.RoomIndexData, arg string) {
	n, err := strconv.Atoi(strings.TrimSpace(arg))
	if err != nil {
		d.WriteToBuffer("Exit must be specified in a number.\n\r")
		reditDispExitMenu(d)
		return
	}
	if n < 1 || n > len(room.Exits) {
		d.WriteToBuffer("That exit does not exist.\n\r")
		reditDispExitMenu(d)
		return
	}
	idx := n - 1
	gone := room.Exits[idx]
	olcLog(d, "ROOM", "Removed %s exit", directionName(gone.Direction))
	room.Exits = append(room.Exits[:idx], room.Exits[idx+1:]...)
	reditDispExitMenu(d)
}

// reditHandleExitVnum sets the destination vnum for the currently-
// selected exit. Unlike C src/oredit.c:860-875 (which sets pexit->vnum
// only), Go updates BOTH exit.Vnum AND exit.ToRoom so the exit isn't
// left silently pointing at the prior room. Documented C-bug fix
// (plan §Open Q4; adversary resolved 2026-04-18).
func reditHandleExitVnum(d *types.DescriptorData, arg string) {
	pexit, ok := d.Olc.Spare.(*types.ExitData)
	if !ok || pexit == nil {
		reditDispExitMenu(d)
		return
	}
	n, err := strconv.Atoi(strings.TrimSpace(arg))
	if err != nil || n < 0 || n > types.MAX_VNUM {
		d.WriteToBuffer("Invalid room number, try again : ")
		return
	}
	target := worldRoomLookup(n)
	if target == nil {
		d.WriteToBuffer("That room does not exist, try again: ")
		return
	}
	pexit.Vnum = n
	pexit.RVnum = n
	pexit.ToRoom = target // Go fix — C leaves this stale (C bug).
	olcLog(d, "ROOM", "%s exit vnum changed to %d", directionName(pexit.Direction), n)
	reditDispExitMenu(d)
}

// reditHandleExitKey stores the key vnum on the selected exit.
// Mirrors C REDIT_EXIT_KEY at src/oredit.c:884-894.
func reditHandleExitKey(d *types.DescriptorData, arg string) {
	pexit, ok := d.Olc.Spare.(*types.ExitData)
	if !ok || pexit == nil {
		reditDispExitMenu(d)
		return
	}
	n, err := strconv.Atoi(strings.TrimSpace(arg))
	if err != nil || n < 0 || n > types.MAX_VNUM {
		d.WriteToBuffer("Invalid vnum, try again: ")
		return
	}
	pexit.Key = n
	olcLog(d, "ROOM", "%s key vnum is now %d", directionName(pexit.Direction), n)
	reditDispExitEdit(d)
}

// reditHandleExitKeyword stores the keyword(s) on the selected exit.
// Mirrors C REDIT_EXIT_KEYWORD at src/oredit.c:877-882.
func reditHandleExitKeyword(d *types.DescriptorData, arg string) {
	pexit, ok := d.Olc.Spare.(*types.ExitData)
	if !ok || pexit == nil {
		reditDispExitMenu(d)
		return
	}
	pexit.Keyword = util.SmashTilde(strings.TrimSpace(arg))
	olcLog(d, "ROOM", "Changed %s keyword to %s", directionName(pexit.Direction), pexit.Keyword)
	reditDispExitEdit(d)
}

// reditHandleExitDesc sets the one-line exit description.
// Mirrors C REDIT_EXIT_DESC at src/oredit.c:780-790.
func reditHandleExitDesc(d *types.DescriptorData, arg string) {
	pexit, ok := d.Olc.Spare.(*types.ExitData)
	if !ok || pexit == nil {
		reditDispExitMenu(d)
		return
	}
	arg = strings.TrimSpace(arg)
	if arg == "" {
		pexit.Description = ""
	} else {
		pexit.Description = util.SmashTilde(arg) + "\n\r"
	}
	olcLog(d, "ROOM", "Changed %s description to %s", directionName(pexit.Direction), arg)
	reditDispExitEdit(d)
}

// reditHandleExitFlags toggles an exit-flag bit by 1-indexed position.
// Mirrors C REDIT_EXIT_FLAGS at src/oredit.c:896-919.
func reditHandleExitFlags(d *types.DescriptorData, arg string) {
	pexit, ok := d.Olc.Spare.(*types.ExitData)
	if !ok || pexit == nil {
		reditDispExitMenu(d)
		return
	}
	n, err := strconv.Atoi(strings.TrimSpace(arg))
	if err != nil {
		d.WriteToBuffer("That's not a valid choice!\n\r")
		reditDispExitFlagMenu(d)
		return
	}
	if n == 0 {
		reditDispExitEdit(d)
		return
	}
	if n < 0 || n > types.MAX_EXFLAG+1 {
		d.WriteToBuffer("That's not a valid choice!\n\r")
		reditDispExitFlagMenu(d)
		return
	}
	bit := n - 1
	if exitFlagReserved(bit) {
		d.WriteToBuffer("That's not a valid choice!\n\r")
		reditDispExitFlagMenu(d)
		return
	}
	pexit.ExitInfo ^= 1 << bit
	label := "(unnamed)"
	if bit < len(exitFlagLabels) {
		label = exitFlagLabels[bit]
	}
	action := "Removed"
	if pexit.ExitInfo&(1<<bit) != 0 {
		action = "Added"
	}
	olcLog(d, "ROOM", "%s %s to %s exit", action, label, directionName(pexit.Direction))
	reditDispExitFlagMenu(d)
}

// reditHandleExtradescMenu interprets A/R/<number>/Q at the extradesc
// list. Mirrors C REDIT_EXTRADESC_MENU at src/oredit.c:982-1020.
func reditHandleExtradescMenu(d *types.DescriptorData, room *types.RoomIndexData, arg string) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		reditDispExtradescMenu(d)
		return
	}
	switch strings.ToLower(arg)[:1] {
	case "q":
		ReditDispMenu(d)
		return
	case "a":
		ed := &types.ExtraDescrData{Keyword: "", Description: ""}
		room.ExtraDescr = append(room.ExtraDescr, ed)
		d.Olc.Spare = ed
		olcLog(d, "ROOM", "Added new exdesc")
		reditDispExtradescChoice(d)
		return
	case "r":
		d.Olc.Mode = types.REDIT_EXTRADESC_DELETE
		d.WriteToBuffer("Delete which extra description? ")
		return
	default:
		n, err := strconv.Atoi(arg)
		if err != nil || n < 1 || n > len(room.ExtraDescr) {
			reditDispExtradescMenu(d)
			return
		}
		ed := room.ExtraDescr[n-1]
		d.Olc.Spare = ed
		reditDispExtradescChoice(d)
	}
}

// reditHandleExtradescChoice handles the 1/2/Q selection for a
// specific extradesc. Mirrors C REDIT_EXTRADESC_CHOICE at
// src/oredit.c:937-966. Q junks an empty extradesc (no keyword).
func reditHandleExtradescChoice(d *types.DescriptorData, room *types.RoomIndexData, arg string) {
	ed, ok := d.Olc.Spare.(*types.ExtraDescrData)
	if !ok || ed == nil {
		reditDispExtradescMenu(d)
		return
	}
	arg = strings.TrimSpace(arg)
	if arg == "" {
		reditDispExtradescChoice(d)
		return
	}
	switch strings.ToLower(arg)[:1] {
	case "q":
		// C junks when keyword OR description is empty. We match C
		// (keyword empty OR description empty) — either gate fires.
		if ed.Keyword == "" || ed.Description == "" {
			d.WriteToBuffer("No keyword and/or description, junking...\n\r")
			for i, e := range room.ExtraDescr {
				if e == ed {
					room.ExtraDescr = append(room.ExtraDescr[:i], room.ExtraDescr[i+1:]...)
					break
				}
			}
		}
		d.Olc.Spare = nil
		reditDispExtradescMenu(d)
	case "1":
		d.Olc.Mode = types.REDIT_EXTRADESC_KEY
		d.WriteToBuffer("Keywords, seperated by spaces: ")
	case "2":
		ch := d.Character
		ch.Substate = types.SUB_ROOM_EXTRA
		targetEd := ed
		ch.EditorSave = func(c *types.CharData) {
			if c == nil || c.Desc == nil {
				return
			}
			targetEd.Description = CopyBuffer(c)
			StopEditing(c)
			c.Desc.Connected = int(types.CON_REDIT)
			// Restore the extradesc-choice sub-menu so builder is
			// still positioned on this ed.
			c.Desc.Olc.Spare = targetEd
			olcLog(c.Desc, "ROOM", "Edit description for exdesc %s", targetEd.Keyword)
			reditDispExtradescChoice(c.Desc)
		}
		d.WriteToBuffer("Enter new extradesc description:\n\r")
		StartEditing(ch, ed.Description)
	default:
		reditDispExtradescChoice(d)
	}
}

// reditHandleExtradescKey sets the keyword(s) on the selected extradesc.
// Mirrors C REDIT_EXTRADESC_KEY at src/oredit.c:968-980.
func reditHandleExtradescKey(d *types.DescriptorData, arg string) {
	ed, ok := d.Olc.Spare.(*types.ExtraDescrData)
	if !ok || ed == nil {
		reditDispExtradescMenu(d)
		return
	}
	old := ed.Keyword
	ed.Keyword = util.SmashTilde(strings.TrimSpace(arg))
	olcLog(d, "ROOM", "Changed exkey %s to %s", old, ed.Keyword)
	reditDispExtradescChoice(d)
}

// reditHandleExtradescDelete removes an extradesc by 1-indexed slot.
// Mirrors C REDIT_EXTRADESC_DELETE at src/oredit.c:921-935.
func reditHandleExtradescDelete(d *types.DescriptorData, room *types.RoomIndexData, arg string) {
	n, err := strconv.Atoi(strings.TrimSpace(arg))
	if err != nil {
		d.WriteToBuffer("Not found, try again: ")
		return
	}
	if n < 1 || n > len(room.ExtraDescr) {
		d.WriteToBuffer("Not found, try again: ")
		return
	}
	gone := room.ExtraDescr[n-1]
	olcLog(d, "ROOM", "Deleted exdesc %s", gone.Keyword)
	room.ExtraDescr = append(room.ExtraDescr[:n-1], room.ExtraDescr[n:]...)
	reditDispExtradescMenu(d)
}

// cleanupOlc disposes the OLC session and returns the descriptor to
// CON_PLAYING. Idempotent — safe to call whether or not an editor
// session is active. Mirrors C cleanup_olc at src/olc.c (referenced
// from src/oredit.c:586).
func cleanupOlc(d *types.DescriptorData) {
	if d == nil {
		return
	}
	d.Olc = nil
	d.Connected = int(types.CON_PLAYING)
	if d.Character != nil {
		d.Character.Substate = types.SUB_NONE
	}
}

// olcLog emits a structured build-log line. Mirrors C olc_log at
// src/oredit.c:171-210 — the target label selects the appropriate
// ROOM/OBJ/MOB prefix so redit/oedit/medit share the same helper. Callers
// pass the edit-target type as the second arg ("ROOM" for redit, "OBJ"
// for oedit, "MOB" for medit).
// Plan: plan-phase6-olc-oedit.md §G2 (generalized from the redit-only
// predecessor).
func olcLog(d *types.DescriptorData, target, format string, args ...any) {
	if d == nil || d.Character == nil {
		return
	}
	var vnum int
	if d.Olc != nil {
		vnum = d.Olc.Vnum
	}
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	trust := 0
	if d.Character != nil {
		trust = d.Character.GetTrust()
	}
	util.LogStringPlus(
		fmt.Sprintf("Log %s: %s(%d): %s", d.Character.Name, target, vnum, msg),
		types.LOG_BUILD, trust)
}

// worldRoomLookup is the seam the parser uses to resolve room vnums.
// Points at game.worldRef.GetRoom at boot time; test code can override
// directly. Kept as a package-level function variable so redit_parse
// tests don't need to stand up a full world.
var worldRoomLookup = func(vnum int) *types.RoomIndexData {
	if worldRef == nil {
		return nil
	}
	return worldRef.GetRoom(vnum)
}

// uRange clamps v to [low, high]. Mirrors C URANGE macro.
func uRange(low, v, high int) int {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}

// parseIntDefault returns the integer parsed from s, or def if parsing
// fails. Mirrors C atoi semantics for menu prompts.
func parseIntDefault(s string, def int) int {
	if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return n
	}
	return def
}
