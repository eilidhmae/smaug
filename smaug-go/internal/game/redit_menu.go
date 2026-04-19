// Package game — redit menu rendering helpers.
//
// Ports src/oredit.c menu-display functions (redit_disp_*). Each function
// writes color-tagged menu text to the descriptor's output buffer; the
// descriptor's ColorFunc translates the tags on flush. Line endings use
// Go-convention "\n\r" to match the rest of the port.
//
// The ANSI screen-clear sequence C emits at the top of each menu
// ("50\x1B[;H\x1B[2J") is deliberately omitted — cosmetic in C, and the
// testclient harness does not interpret cursor-position control codes.
// Decision recorded in plan-phase6-olc-redit.md §Open Q1.
package game

import (
	"fmt"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
)

// sectorName returns the human-readable label for a sector constant.
// Mirrors the switch in C redit_disp_menu at src/oredit.c:430-446.
func sectorName(sect int) string {
	switch sect {
	case types.SECT_INSIDE:
		return "Inside"
	case types.SECT_CITY:
		return "City"
	case types.SECT_FIELD:
		return "Field"
	case types.SECT_FOREST:
		return "Forest"
	case types.SECT_HILLS:
		return "Hills"
	case types.SECT_MOUNTAIN:
		return "Mountains"
	case types.SECT_WATER_SWIM:
		return "Swim"
	case types.SECT_WATER_NOSWIM:
		return "Noswim"
	case types.SECT_UNDERWATER:
		return "Underwater"
	case types.SECT_AIR:
		return "Air"
	case types.SECT_DESERT:
		return "Desert"
	case types.SECT_OCEANFLOOR:
		return "Oceanfloor"
	case types.SECT_UNDERGROUND:
		return "Underground"
	default:
		return "???!"
	}
}

// sectorKeywords is the short-name lookup for the sector sub-menu.
// Order matches the iota in enums.go; SECT_DUNNO is skipped by the menu
// renderer (matches C redit_disp_sector_menu skip-continue at :407-408).
var sectorKeywords = []string{
	"inside",       // 0
	"city",         // 1
	"field",        // 2
	"forest",       // 3
	"hills",        // 4
	"mountains",    // 5
	"water_swim",   // 6
	"water_noswim", // 7
	"underwater",   // 8
	"air",          // 9
	"desert",       // 10
	"???!",         // 11 SECT_DUNNO (shown in C as a placeholder — kept for index alignment)
	"oceanfloor",   // 12
	"underground",  // 13
	"lava",         // 14
	"swamp",        // 15
}

// roomFlagNames is the authoritative room-flag-bit label table, sourced
// from src/build.c:92-105 (r_flags[]). Bit index = position in the slice.
// Labels are stable in the wire format so saved area files match C.
var roomFlagNames = []string{
	"dark", "death", "nomob", "indoors", "house", "neutral", "chaotic",
	"nomagic", "nolocate", "private", "safe", "solitary", "petshop",
	"norecall", "donation", "nodropall", "silence", "logspeech", "nodrop",
	"clanstoreroom", "nosummon", "noastral", "teleport", "teleshowdesc",
	"nofloor", "nosupplicate", "arena", "nomissile", "auction", "nohover",
	"prototype", "dnd", "_track_", "light", "nolog", "color", "nowhere",
	"noyell", "noquit", "notrack", "nosuppcorpse", "nosupprecall",
}

// getRoomFlagBit returns the bit position for a case-insensitive flag
// name, or -1 if not found. Mirrors C src/build.c get_rflag().
// Used by the REDIT_FLAGS word-list branch.
func getRoomFlagBit(name string) int {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return -1
	}
	for i, fn := range roomFlagNames {
		if fn == name {
			return i
		}
	}
	return -1
}

// roomFlagsString returns a space-separated list of set flag names for
// the given RoomFlags bitvector, matching C ext_flag_string(r_flags).
// Unknown/high bits beyond the table are omitted.
func roomFlagsString(bv types.BitVector) string {
	var out []string
	for i, fn := range roomFlagNames {
		if bv.IsSet(i) {
			out = append(out, fn)
		}
	}
	return strings.Join(out, " ")
}

// ReditDispMenu renders the main room-editor menu.
// Mirrors C redit_disp_menu at src/oredit.c:423-477.
func ReditDispMenu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	room, ok := d.Olc.Target.(*types.RoomIndexData)
	if !ok || room == nil {
		return
	}

	areaName := "None????"
	if room.Area != nil && room.Area.Name != "" {
		areaName = room.Area.Name
	}

	var sb strings.Builder
	fmt.Fprintf(&sb,
		"&w-- Room number : [&c%d&w]      Room area: [&c%-30.30s&w]\n\r"+
			"&g1&w) Name        : &O%s\n\r"+
			"&g2&w) Description :\n\r&O%s"+
			"&g3&w) Room flags  : &c%s\n\r"+
			"&g4&w) Sector type : &c%s\n\r"+
			"&g5&w) Tunnel      : &c%d\n\r"+
			"&g6&w) TeleDelay   : &c%d\n\r"+
			"&g7&w) TeleVnum    : &c%d\n\r"+
			"&gA&w) Exit menu\n\r"+
			"&gB&w) Extra descriptions menu\n\r"+
			"&gQ&w) Quit\n\r"+
			"Enter choice : ",
		d.Olc.Vnum,
		areaName,
		room.Name,
		room.Description,
		roomFlagsString(room.RoomFlags),
		sectorName(room.SectorType),
		room.Tunnel,
		room.TeleDelay,
		room.TeleVnum,
	)

	d.WriteToBuffer(sb.String())
	d.Olc.Mode = types.REDIT_MAIN_MENU
}

// reditDispFlagMenu renders the room-flag sub-menu. Two-column list of
// all r_flags labels with the current set shown. Mirrors
// src/oredit.c:366-389.
func reditDispFlagMenu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	room, ok := d.Olc.Target.(*types.RoomIndexData)
	if !ok || room == nil {
		return
	}

	var sb strings.Builder
	for i, fn := range roomFlagNames {
		fmt.Fprintf(&sb, "&g%2d&w) %-20.20s ", i+1, fn)
		if (i+1)%2 == 0 {
			sb.WriteString("\n\r")
		}
	}
	if len(roomFlagNames)%2 != 0 {
		sb.WriteString("\n\r")
	}
	fmt.Fprintf(&sb, "\n\rRoom flags: &c%s&w\n\rEnter room flags, 0 to quit : ",
		roomFlagsString(room.RoomFlags))

	d.WriteToBuffer(sb.String())
	d.Olc.Mode = types.REDIT_FLAGS
}

// reditDispSectorMenu renders the sector sub-menu.
// Mirrors src/oredit.c:399-420. Skips SECT_DUNNO (index 11) to match C.
func reditDispSectorMenu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}

	var sb strings.Builder
	col := 0
	for i := 0; i < types.SECT_MAX; i++ {
		if i == types.SECT_DUNNO {
			continue
		}
		label := sectorKeywords[i]
		fmt.Fprintf(&sb, "&g%2d&w) %-20.20s ", i, label)
		col++
		if col%2 == 0 {
			sb.WriteString("\n\r")
		}
	}
	if col%2 != 0 {
		sb.WriteString("\n\r")
	}
	sb.WriteString("\n\rEnter sector type : ")

	d.WriteToBuffer(sb.String())
	d.Olc.Mode = types.REDIT_SECTOR
}

// dirNames maps DIR_NORTH..DIR_DOWN to their human labels. DIR_SOMEWHERE
// is the special "portal" direction C keeps at index 10 — we skip it in
// the plan (no portal exits from OLC) but retain the name for the
// exit-add direction menu.
var dirNames = []string{
	"north", "east", "south", "west", "up", "down",
	"northeast", "northwest", "southeast", "southwest",
	"somewhere",
}

// directionName returns the label for a direction constant.
func directionName(dir int) string {
	if dir < 0 || dir >= len(dirNames) {
		return "???"
	}
	return dirNames[dir]
}

// reditDispExitMenu lists the room's exits and the A/R/Q options.
// Mirrors src/oredit.c:260-291.
func reditDispExitMenu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	room, ok := d.Olc.Target.(*types.RoomIndexData)
	if !ok || room == nil {
		return
	}

	d.Olc.Mode = types.REDIT_EXIT_MENU

	var sb strings.Builder
	for i, ex := range room.Exits {
		toVnum := 0
		if ex.ToRoom != nil {
			toVnum = ex.ToRoom.Vnum
		}
		kw := ex.Keyword
		if kw == "" {
			kw = "(none)"
		}
		fmt.Fprintf(&sb,
			"&g%2d&w) %-10.10s to %-5d.  Key: %d  Flags: %d  Keywords: %s.\n\r",
			i+1, directionName(ex.Direction), toVnum, ex.Key, ex.ExitInfo, kw)
	}
	if len(room.Exits) > 0 {
		sb.WriteString("\n\r")
	}
	sb.WriteString("&gA&w) Add a new exit\n\r")
	sb.WriteString("&gR&w) Remove an exit\n\r")
	sb.WriteString("&gQ&w) Quit\n\r")
	sb.WriteString("\n\rEnter choice: ")

	d.WriteToBuffer(sb.String())
}

// exitFlagLabels mirrors C ex_flags[] at src/tables.c; used by the exit
// flag display/toggle. Index = bit position (0..MAX_EXFLAG).
var exitFlagLabels = []string{
	"isdoor",     // 0
	"closed",     // 1
	"locked",     // 2
	"secret",     // 3
	"pickproof",  // 4
	"fly",        // 5
	"climb",      // 6
	"dig",        // 7
	"nomob",      // 8
	"window",     // 9
	"nopassdoor", // 10
	"nosee",      // 11
	"hidden",     // 12
	"passage",    // 13
	"portal",     // 14 EX_PORTAL — reserved/skipped
	"bashed",     // 15
	"bashproof",  // 16
	"reserved1",  // 17 EX_RES1 — reserved/skipped
	"can_walk",   // 18
	"can_climb",  // 19
	"can_fly",    // 20
	"can_swim",   // 21
	"can_crawl",  // 22
	"can_dig",    // 23
	"ocean",      // 24
	"floor",      // 25
	"reserved2",  // 26 EX_RES2 — reserved/skipped
	"no_lookin",  // 27
	"has_lock",   // 28
}

// exitFlagReserved returns true for bits C explicitly skips in the flag
// menu. Mirrors src/oredit.c:347-348 (EX_RES1, EX_RES2, EX_PORTAL).
func exitFlagReserved(bit int) bool {
	// EX_RES1 = 17, EX_RES2 = 26, EX_PORTAL = 14 per src/mud.h.
	return bit == 14 || bit == 17 || bit == 26
}

// exitFlagsString renders the currently-set exit flag labels as a
// space-separated string. Used by the exit-edit + exit-flag menus.
func exitFlagsString(exitInfo int) string {
	var out []string
	for i, label := range exitFlagLabels {
		if exitInfo&(1<<i) != 0 {
			out = append(out, label)
		}
	}
	return strings.Join(out, " ")
}

// reditDispExitEdit renders the per-exit field menu.
// Mirrors src/oredit.c:293-320.
func reditDispExitEdit(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	pexit, ok := d.Olc.Spare.(*types.ExitData)
	if !ok || pexit == nil {
		return
	}

	d.Olc.Mode = types.REDIT_EXIT_EDIT

	flags := exitFlagsString(pexit.ExitInfo)
	if flags == "" {
		flags = "(none)"
	}
	keyword := pexit.Keyword
	if keyword == "" {
		keyword = "(none)"
	}
	description := pexit.Description
	if description == "" {
		description = "(none)"
	}
	toVnum := -1
	if pexit.ToRoom != nil {
		toVnum = pexit.ToRoom.Vnum
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "&g1&w) Direction  : &c%s\n\r", directionName(pexit.Direction))
	fmt.Fprintf(&sb, "&g2&w) To Vnum    : &c%d\n\r", toVnum)
	fmt.Fprintf(&sb, "&g3&w) Key        : &c%d\n\r", pexit.Key)
	fmt.Fprintf(&sb, "&g4&w) Keyword    : &c%s\n\r", keyword)
	fmt.Fprintf(&sb, "&g5&w) Flags      : &c%s\n\r", flags)
	fmt.Fprintf(&sb, "&g6&w) Description: &c%s\n\r", description)
	sb.WriteString("&gQ&w) Quit\n\r")
	sb.WriteString("\n\rEnter choice: ")

	d.WriteToBuffer(sb.String())
}

// reditDispExitDirs lists DIR_NORTH..DIR_SOMEWHERE for the add-exit
// prompt. Mirrors src/oredit.c:322-334.
func reditDispExitDirs(d *types.DescriptorData) {
	if d == nil {
		return
	}
	var sb strings.Builder
	for i := 0; i <= types.DIR_SOMEWHERE; i++ {
		fmt.Fprintf(&sb, "&g%2d&w) %s\n\r", i, directionName(i))
	}
	sb.WriteString("\n\rChoose a direction: ")
	d.WriteToBuffer(sb.String())
}

// reditDispExitFlagMenu renders the exit-flag toggle sub-menu.
// Mirrors src/oredit.c:337-363.
func reditDispExitFlagMenu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	pexit, ok := d.Olc.Spare.(*types.ExitData)
	if !ok || pexit == nil {
		return
	}
	d.Olc.Mode = types.REDIT_EXIT_FLAGS

	var sb strings.Builder
	col := 0
	for i, label := range exitFlagLabels {
		if exitFlagReserved(i) {
			continue
		}
		fmt.Fprintf(&sb, "&g%2d&w) %-20.20s ", i+1, label)
		col++
		if col%2 == 0 {
			sb.WriteString("\n\r")
		}
	}
	if col%2 != 0 {
		sb.WriteString("\n\r")
	}
	current := exitFlagsString(pexit.ExitInfo)
	if current == "" {
		current = "(none)"
	}
	fmt.Fprintf(&sb, "\n\rExit flags: &c%s&w\n\rEnter room flags, 0 to quit: ", current)
	d.WriteToBuffer(sb.String())
}

// reditDispExtradescMenu lists extradescs + A/R/Q options.
// Mirrors src/oredit.c:234-257.
func reditDispExtradescMenu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	room, ok := d.Olc.Target.(*types.RoomIndexData)
	if !ok || room == nil {
		return
	}

	var sb strings.Builder
	for i, ed := range room.ExtraDescr {
		fmt.Fprintf(&sb, "&g%2d&w) Keyword: &O%s\n\r", i+1, ed.Keyword)
	}
	if len(room.ExtraDescr) > 0 {
		sb.WriteString("\n\r")
	}
	sb.WriteString("&gA&w) Add a new description\n\r")
	sb.WriteString("&gR&w) Remove a description\n\r")
	sb.WriteString("&gQ&w) Quit\n\r")
	sb.WriteString("\n\rEnter choice: ")

	d.WriteToBuffer(sb.String())
	d.Olc.Mode = types.REDIT_EXTRADESC_MENU
}

// reditDispExtradescChoice renders the per-extradesc edit sub-menu:
// keywords / description / quit. Mirrors C oedit_disp_extra_choice.
func reditDispExtradescChoice(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	ed, ok := d.Olc.Spare.(*types.ExtraDescrData)
	if !ok || ed == nil {
		return
	}

	keyword := ed.Keyword
	if keyword == "" {
		keyword = "(none)"
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "&g1&w) Keywords    : &c%s\n\r", keyword)
	sb.WriteString("&g2&w) Description : (use editor)\n\r")
	sb.WriteString("&gQ&w) Quit\n\r")
	sb.WriteString("\n\rEnter choice: ")

	d.WriteToBuffer(sb.String())
	d.Olc.Mode = types.REDIT_EXTRADESC_CHOICE
}
