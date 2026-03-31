package persist

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// LoadAreas loads all areas listed in area.lst from the given directory.
func LoadAreas(w *world.World, areaDir string) error {
	listPath := filepath.Join(areaDir, "area.lst")
	data, err := os.ReadFile(listPath)
	if err != nil {
		return fmt.Errorf("LoadAreas: cannot read %s: %w", listPath, err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "$" {
			break
		}
		areaPath := filepath.Join(areaDir, line)
		if err := loadAreaFile(w, areaPath); err != nil {
			util.Bug("LoadAreas: error loading %s: %v", areaPath, err)
			// Continue loading other areas, like the C code does during boot.
		}
	}
	return nil
}

// loadAreaFile loads a single .are file.
func loadAreaFile(w *world.World, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("loadAreaFile: cannot open %s: %w", filename, err)
	}
	defer f.Close()

	sc := NewScanner(f, filename)

	var area *types.AreaData

	for {
		letter := sc.ReadLetter()
		if letter == 0 {
			break // EOF
		}
		if letter != '#' {
			util.Bug("loadAreaFile: %s:%d: '#' not found, got '%c' — skipping to next section", filename, sc.Line(), letter)
			// Skip forward to find the next '#' at start of a line
			skipSection(sc, "recovery")
			continue
		}

		word := sc.ReadWord()
		if word == "$" {
			break
		}

		switch strings.ToUpper(word) {
		case "AREA":
			area = &types.AreaData{
				Filename: filepath.Base(filename),
			}
			area.Name = sc.ReadStringNoHash()
			w.Areas = append(w.Areas, area)
			w.TopArea++

		case "AUTHOR":
			if area != nil {
				area.Author = sc.ReadStringNoHash()
			} else {
				sc.ReadStringNoHash()
			}

		case "RANGES":
			if area != nil {
				loadRanges(sc, area)
			}

		case "FLAGS":
			if area != nil {
				area.Flags = sc.ReadNumber()
			} else {
				sc.ReadNumber()
			}

		case "ECONOMY":
			if area != nil {
				area.HighEconomy = sc.ReadNumber()
				area.LowEconomy = sc.ReadNumber()
			} else {
				sc.ReadNumber()
				sc.ReadNumber()
			}

		case "RESETMSG":
			if area != nil {
				area.ResetMsg = sc.ReadStringNoHash()
			} else {
				sc.ReadStringNoHash()
			}

		case "MOBILES":
			if area == nil {
				util.Bug("loadAreaFile: %s: #MOBILES before #AREA", filename)
				return fmt.Errorf("loadAreaFile: MOBILES before AREA")
			}
			loadMobiles(w, sc, area)

		case "OBJECTS":
			if area == nil {
				util.Bug("loadAreaFile: %s: #OBJECTS before #AREA", filename)
				return fmt.Errorf("loadAreaFile: OBJECTS before AREA")
			}
			loadObjects(w, sc, area)

		case "ROOMS":
			if area == nil {
				util.Bug("loadAreaFile: %s: #ROOMS before #AREA", filename)
				return fmt.Errorf("loadAreaFile: ROOMS before AREA")
			}
			loadRooms(w, sc, area)

		case "RESETS":
			if area == nil {
				util.Bug("loadAreaFile: %s: #RESETS before #AREA", filename)
				return fmt.Errorf("loadAreaFile: RESETS before AREA")
			}
			loadResets(sc, area)

		case "SHOPS":
			if area == nil {
				util.Bug("loadAreaFile: %s: #SHOPS before #AREA", filename)
				return fmt.Errorf("loadAreaFile: SHOPS before AREA")
			}
			loadShops(w, sc)

		case "REPAIRS":
			if area == nil {
				util.Bug("loadAreaFile: %s: #REPAIRS before #AREA", filename)
				return fmt.Errorf("loadAreaFile: REPAIRS before AREA")
			}
			loadRepairs(w, sc)

		case "SPECIALS":
			if area == nil {
				util.Bug("loadAreaFile: %s: #SPECIALS before #AREA", filename)
				return fmt.Errorf("loadAreaFile: SPECIALS before AREA")
			}
			loadSpecials(w, sc)

		case "VERSION":
			if area != nil {
				area.Version = sc.ReadNumber()
			} else {
				sc.ReadNumber()
			}

		case "SPELLLIMIT":
			if area != nil {
				area.SpellLimit = sc.ReadNumber()
			} else {
				sc.ReadNumber()
			}

		case "CLIMATE", "NEIGHBOR", "HELPS", "MUDPROGS", "OBJPROGS",
			"CREDITS", "CONTINENT":
			// Skip unsupported sections by reading to the next '#' marker.
			skipSection(sc, word)

		default:
			util.Bug("loadAreaFile: %s:%d: unknown section '%s'", filename, sc.Line(), word)
			skipSection(sc, word)
		}
	}

	if area != nil {
		util.LogString(fmt.Sprintf("%-14s: Rooms: %5d - %-5d Objs: %5d - %-5d Mobs: %5d - %d",
			area.Filename,
			area.LowRVnum, area.HiRVnum,
			area.LowOVnum, area.HiOVnum,
			area.LowMVnum, area.HiMVnum))
	}

	return nil
}

// loadRanges reads the #RANGES section: 4 numbers followed by a $ line.
func loadRanges(sc *Scanner, area *types.AreaData) {
	area.LowSoftRange = sc.ReadNumber()
	area.HiSoftRange = sc.ReadNumber()
	area.LowHardRange = sc.ReadNumber()
	area.HiHardRange = sc.ReadNumber()
	// Consume the terminating "$" line.
	for {
		w := sc.ReadWord()
		if w == "$" || w == "" {
			break
		}
	}
}

// skipSection consumes input until the next '#' marker is found (which is
// left in the stream for the main loop to handle), or EOF.
func skipSection(sc *Scanner, sectionName string) {
	// Skip an unrecognised or unsupported section by consuming bytes until we
	// find a '#' at the start of a line (i.e. after a newline). We unread the
	// '#' so the main dispatch loop can process it.
	atLineStart := false
	for {
		b, err := sc.readByte()
		if err != nil {
			return // EOF
		}
		if b == '#' && atLineStart {
			sc.unreadByte()
			return
		}
		if b == '\n' {
			sc.line++
			atLineStart = true
		} else {
			atLineStart = false
		}
	}
}

// ---------------------------------------------------------------------------
// Mobiles
// ---------------------------------------------------------------------------

func loadMobiles(w *world.World, sc *Scanner, area *types.AreaData) {
	for {
		letter := sc.ReadLetter()
		if letter != '#' {
			util.Bug("loadMobiles: %s:%d: '#' not found", sc.File(), sc.Line())
			return
		}

		vnum := sc.ReadNumber()
		if vnum == 0 {
			break
		}

		if w.MobIndex[vnum] != nil {
			util.Bug("loadMobiles: vnum %d duplicated in %s", vnum, sc.File())
		}

		mob := &types.MobIndexData{Vnum: vnum}

		// Track vnum ranges.
		if area.LowMVnum == 0 || vnum < area.LowMVnum {
			area.LowMVnum = vnum
		}
		if vnum > area.HiMVnum {
			area.HiMVnum = vnum
		}

		mob.PlayerName = sc.ReadString()
		mob.ShortDescr = sc.ReadString()
		mob.LongDescr = sc.ReadString()
		mob.Description = sc.ReadString()

		// Capitalize first char of long_descr and description.
		mob.LongDescr = capitalizeFirst(mob.LongDescr)
		mob.Description = capitalizeFirst(mob.Description)

		// Line 1: act affected_by alignment letter
		// These are read as individual numbers, not whole-line bitvectors.
		actFlags := sc.ReadNumber()
		mob.Act = types.BitVectorFromInt(uint32(actFlags))
		mob.Act.Set(types.ACT_IS_NPC)
		affFlags := sc.ReadNumber()
		mob.AffectedBy = types.BitVectorFromInt(uint32(affFlags))
		mob.Alignment = sc.ReadNumber()
		mobLetter := sc.ReadLetter()

		// Line 2: level thac0 ac hitnodice d hitsizedice + hitplus damnodice d damsizedice + damplus
		mob.Level = sc.ReadNumber()
		mob.MobThac0 = sc.ReadNumber()
		mob.AC = sc.ReadNumber()
		mob.HitNoDice = sc.ReadNumber()
		sc.ReadLetter() // 'd'
		mob.HitSizeDice = sc.ReadNumber()
		sc.ReadLetter() // '+'
		mob.HitPlus = sc.ReadNumber()
		mob.DamNoDice = sc.ReadNumber()
		sc.ReadLetter() // 'd'
		mob.DamSizeDice = sc.ReadNumber()
		sc.ReadLetter() // '+'
		mob.DamPlus = sc.ReadNumber()

		// Gold/exp line. Format depends on ENABLE_GOLD_SILVER_COPPER:
		// Without: gold exp
		// With:    exp gold silver copper
		// We detect by reading the first number: if the rest of the gold
		// line has 3 more numbers, it's the multi-currency format.
		// Since fread_number skips newlines, we read via ReadToEOL to get the line.
		goldLine := sc.ReadToEOL()
		goldFields := strings.Fields(goldLine)
		if len(goldFields) >= 4 {
			// Multi-currency: exp gold silver copper
			mob.Exp = atoi(goldFields[0])
			mob.Gold = atoi(goldFields[1])
			mob.Silver = atoi(goldFields[2])
			mob.Copper = atoi(goldFields[3])
		} else if len(goldFields) >= 2 {
			// Standard: gold exp
			mob.Gold = atoi(goldFields[0])
			mob.Exp = atoi(goldFields[1])
		} else if len(goldFields) == 1 {
			// Single number — likely multi-currency with exp on this line
			// and gold/silver/copper on next
			mob.Exp = atoi(goldFields[0])
			mob.Gold = sc.ReadNumber()
			mob.Silver = sc.ReadNumber()
			mob.Copper = sc.ReadNumber()
		}

		// Position line: position defposition sex
		mob.Position = convertPosition(sc.ReadNumber())
		mob.DefPosition = convertPosition(sc.ReadNumber())
		mob.Sex = sc.ReadNumber()

		if mobLetter == 'C' || mobLetter == 'V' {
			// Complex mob: stats, saves, race line, combat line
			mob.PermStr = sc.ReadNumber()
			mob.PermInt = sc.ReadNumber()
			mob.PermWis = sc.ReadNumber()
			mob.PermDex = sc.ReadNumber()
			mob.PermCon = sc.ReadNumber()
			mob.PermCha = sc.ReadNumber()
			mob.PermLck = sc.ReadNumber()

			mob.SavingPoisonDeath = sc.ReadNumber()
			mob.SavingWand = sc.ReadNumber()
			mob.SavingParaPetri = sc.ReadNumber()
			mob.SavingBreath = sc.ReadNumber()
			mob.SavingSpellStaff = sc.ReadNumber()

			// Race/class line: race class height weight speaks speaking numattacks
			raceLine := sc.ReadToEOL()
			parseRaceLine(raceLine, mob)

			// Combat line: hitroll damroll xflags resistant immune susceptible attacks defenses
			combatLine := sc.ReadToEOL()
			parseCombatLine(combatLine, mob)
		} else {
			// Simple mob defaults.
			mob.PermStr = 13
			mob.PermInt = 13
			mob.PermWis = 13
			mob.PermDex = 13
			mob.PermCon = 13
			mob.PermCha = 13
			mob.PermLck = 13
			mob.Class = 3
		}

		if mobLetter == 'V' {
			// Very complex mob: stances
			for i := 0; i < types.MAX_STANCE; i++ {
				mob.Stances[i] = sc.ReadNumber()
			}
		} else {
			// Set basic stances based on level.
			setDefaultStances(mob)
		}

		// Check for mudprogs (lines starting with '>').
		loadMobProgs(sc, mob)

		w.MobIndex[vnum] = mob
		w.TopMob++
	}
}

// parseRaceLine parses "race class height weight speaks speaking numattacks".
func parseRaceLine(line string, mob *types.MobIndexData) {
	fields := strings.Fields(line)
	if len(fields) >= 1 {
		mob.Race = atoi(fields[0])
	}
	if len(fields) >= 2 {
		mob.Class = atoi(fields[1])
	}
	if len(fields) >= 3 {
		mob.Height = atoi(fields[2])
	}
	if len(fields) >= 4 {
		mob.Weight = atoi(fields[3])
	}
	if len(fields) >= 5 {
		mob.Speaks = atoi(fields[4])
	}
	if len(fields) >= 6 {
		mob.Speaking = atoi(fields[5])
	}
	if len(fields) >= 7 {
		mob.NumAttacks = atoi(fields[6])
	}
	if mob.Speaks == 0 {
		mob.Speaks = int(types.LANG_COMMON)
	}
	if mob.Speaking == 0 {
		mob.Speaking = int(types.LANG_COMMON)
	}
}

// parseCombatLine parses "hitroll damroll xflags resistant immune susceptible attacks defenses".
// Note: in the non-XBI C code, attacks and defenses are single ints read via sscanf.
// In the XBI code they would be bitvectors. We parse as ints since the area files
// use the non-XBI format (space-separated ints on one line).
func parseCombatLine(line string, mob *types.MobIndexData) {
	fields := strings.Fields(line)
	if len(fields) >= 1 {
		mob.Hitroll = atoi(fields[0])
	}
	if len(fields) >= 2 {
		mob.Damroll = atoi(fields[1])
	}
	if len(fields) >= 3 {
		mob.XFlags = atoi(fields[2])
	}
	if len(fields) >= 4 {
		mob.Resistant = atoi(fields[3])
	}
	if len(fields) >= 5 {
		mob.Immune = atoi(fields[4])
	}
	if len(fields) >= 6 {
		mob.Susceptible = atoi(fields[5])
	}
	if len(fields) >= 7 {
		mob.Attacks = types.BitVectorFromInt(parseUint32(fields[6]))
	}
	if len(fields) >= 8 {
		mob.Defenses = types.BitVectorFromInt(parseUint32(fields[7]))
	}
}

// setDefaultStances sets basic stances based on mob level, matching C behavior.
func setDefaultStances(mob *types.MobIndexData) {
	var val int
	switch {
	case mob.Level < 10:
		val = 25
	case mob.Level < 20:
		val = 50
	case mob.Level < 30:
		val = 100
	case mob.Level < 40:
		val = 150
	case mob.Level < 50:
		val = 175
	default:
		val = 200
	}
	for i := 0; i < int(types.BASIC_STANCE); i++ {
		mob.Stances[i] = val
	}
}

// loadMobProgs reads optional mob programs (lines starting with '>') after mob data.
func loadMobProgs(sc *Scanner, mob *types.MobIndexData) {
	for {
		letter := sc.ReadLetter()
		if letter == 0 {
			return
		}
		if letter != '>' {
			sc.unreadByte()
			return
		}
		// Read prog type word (e.g. "speech_prog")
		progType := sc.ReadWord()
		// Read arglist (tilde-terminated)
		argList := sc.ReadString()
		// Read comlist (tilde-terminated)
		comList := sc.ReadString()

		prog := &types.MProgData{
			Type:    mprogNameToType(progType),
			ArgList: argList,
			ComList: comList,
		}
		mob.MudProgs = append(mob.MudProgs, prog)
		mob.ProgTypes.Set(prog.Type)

		// After comlist, check for '|' separator
		sep := sc.ReadLetter()
		if sep != '|' {
			sc.unreadByte()
		}
	}
}

// ---------------------------------------------------------------------------
// Objects
// ---------------------------------------------------------------------------

func loadObjects(w *world.World, sc *Scanner, area *types.AreaData) {
	for {
		letter := sc.ReadLetter()
		if letter != '#' {
			if letter == 0 {
				return // EOF
			}
			util.Bug("loadObjects: %s:%d: '#' not found, got '%c' — skipping to next object", sc.File(), sc.Line(), letter)
			skipSection(sc, "obj-recovery")
			continue
		}

		vnum := sc.ReadNumber()
		if vnum == 0 {
			break
		}

		if w.ObjIndex[vnum] != nil {
			util.Bug("loadObjects: vnum %d duplicated in %s", vnum, sc.File())
		}

		obj := &types.ObjIndexData{Vnum: vnum}

		if area.LowOVnum == 0 || vnum < area.LowOVnum {
			area.LowOVnum = vnum
		}
		if vnum > area.HiOVnum {
			area.HiOVnum = vnum
		}

		obj.Name = sc.ReadString()
		obj.ShortDescr = sc.ReadString()
		obj.Description = sc.ReadString()
		obj.ActionDesc = sc.ReadString()

		obj.Description = capitalizeFirst(obj.Description)

		// Line 1: item_type extra_flags wear_flags [layers [level]]
		typeLine := sc.ReadToEOL()
		parseObjTypeLine(typeLine, obj)

		// Line 2: value[0..5]
		valLine := sc.ReadToEOL()
		parseObjValues(valLine, obj)

		// Line 3: weight cost [silver copper] [rent]
		costLine := sc.ReadToEOL()
		parseObjCostLine(costLine, obj)

		// Read optional trailing sections: E (extra descr), A (affect), > (prog)
		loadObjExtras(sc, obj)

		w.ObjIndex[vnum] = obj
		w.TopObj++
	}
}

// parseObjTypeLine parses "item_type extra_flags wear_flags [layers [level]]".
func parseObjTypeLine(line string, obj *types.ObjIndexData) {
	fields := strings.Fields(line)
	if len(fields) >= 1 {
		obj.ItemType = atoi(fields[0])
	}
	if len(fields) >= 2 {
		obj.ExtraFlags = types.BitVectorFromInt(uint32(atoi(fields[1])))
	}
	if len(fields) >= 3 {
		obj.WearFlags = atoi(fields[2])
	}
	if len(fields) >= 4 {
		obj.Layers = atoi(fields[3])
	}
	if len(fields) >= 5 {
		obj.Level = atoi(fields[4])
	}
}

// parseObjCostLine parses "weight cost [silver copper] [rent]".
func parseObjCostLine(line string, obj *types.ObjIndexData) {
	fields := strings.Fields(line)
	if len(fields) >= 1 {
		obj.Weight = atoi(fields[0])
		if obj.Weight < 1 {
			obj.Weight = 1
		}
	}
	if len(fields) >= 2 {
		obj.GoldCost = atoi(fields[1])
	}
	if len(fields) == 3 {
		// weight cost rent
		obj.Rent = atoi(fields[2])
	} else if len(fields) >= 5 {
		// weight gold silver copper rent
		obj.SilverCost = atoi(fields[2])
		obj.CopperCost = atoi(fields[3])
		obj.Rent = atoi(fields[4])
	} else if len(fields) == 4 {
		// weight gold silver copper (no rent)
		obj.SilverCost = atoi(fields[2])
		obj.CopperCost = atoi(fields[3])
	}
}

// parseObjValues parses "value[0] value[1] ... value[5]".
func parseObjValues(line string, obj *types.ObjIndexData) {
	fields := strings.Fields(line)
	for i := 0; i < len(fields) && i < 6; i++ {
		obj.Value[i] = atoi(fields[i])
	}
}

// parseObjCosts parses the remainder after weight and first cost.
// This handles the simple format: "cost rent" where we already read weight and cost.
// The cost rest contains: "rent" (and rent is unused).

func loadObjExtras(sc *Scanner, obj *types.ObjIndexData) {
	for {
		letter := sc.ReadLetter()
		if letter == 0 {
			return
		}

		switch letter {
		case 'E':
			ed := &types.ExtraDescrData{
				Keyword:     sc.ReadString(),
				Description: sc.ReadString(),
			}
			obj.ExtraDescr = append(obj.ExtraDescr, ed)

		case 'A':
			aff := &types.AffectData{
				Type:     -1,
				Duration: -1,
				Location: sc.ReadNumber(),
				Modifier: sc.ReadNumber(),
			}
			obj.Affects = append(obj.Affects, aff)

		case '>':
			// Object mudprog
			progType := sc.ReadWord()
			argList := sc.ReadString()
			comList := sc.ReadString()
			prog := &types.MProgData{
				Type:    mprogNameToType(progType),
				ArgList: argList,
				ComList: comList,
			}
			obj.MudProgs = append(obj.MudProgs, prog)
			obj.ProgTypes.Set(prog.Type)
			// Consume '|' separator if present
			sep := sc.ReadLetter()
			if sep != '|' {
				sc.unreadByte()
			}

		default:
			sc.unreadByte()
			return
		}
	}
}

// ---------------------------------------------------------------------------
// Rooms
// ---------------------------------------------------------------------------

func loadRooms(w *world.World, sc *Scanner, area *types.AreaData) {
	for {
		letter := sc.ReadLetter()
		if letter != '#' {
			util.Bug("loadRooms: %s:%d: '#' not found", sc.File(), sc.Line())
			return
		}

		vnum := sc.ReadNumber()
		if vnum == 0 {
			break
		}

		if w.Rooms[vnum] != nil {
			util.Bug("loadRooms: vnum %d duplicated in %s", vnum, sc.File())
		}

		room := &types.RoomIndexData{
			Vnum: vnum,
			Area: area,
		}

		if area.LowRVnum == 0 || vnum < area.LowRVnum {
			area.LowRVnum = vnum
		}
		if vnum > area.HiRVnum {
			area.HiRVnum = vnum
		}

		room.Name = sc.ReadString()
		room.Description = sc.ReadString()

		// Line: (unused) room_flags sector_type [tele_delay tele_vnum tunnel [max_weight]]
		// Read the entire line to avoid ReadNumber consuming the newline.
		flagLine := sc.ReadToEOL()
		parseRoomFlagLine(flagLine, room)

		// Read room contents: exits (D), extra descrs (E), mudprogs (>), end (S)
		loadRoomContents(sc, room, vnum)

		w.Rooms[vnum] = room
		w.TopRoom++
	}
}

// parseRoomFlags parses "sector_type [tele_delay tele_vnum tunnel [max_weight]]".
// parseRoomFlagLine parses "(unused) room_flags sector_type [tele_delay tele_vnum tunnel [max_weight]]".
func parseRoomFlagLine(line string, room *types.RoomIndexData) {
	fields := strings.Fields(line)
	// field 0: unused (area number in old format)
	if len(fields) >= 2 {
		room.RoomFlags = types.BitVectorFromInt(uint32(atoi(fields[1])))
	}
	if len(fields) >= 3 {
		room.SectorType = atoi(fields[2])
	}
	if len(fields) >= 4 {
		room.TeleDelay = atoi(fields[3])
	}
	if len(fields) >= 5 {
		room.TeleVnum = atoi(fields[4])
	}
	if len(fields) >= 6 {
		room.Tunnel = atoi(fields[5])
	}
	if len(fields) >= 7 {
		room.MaxWeight = atoi(fields[6])
	}
}

func loadRoomContents(sc *Scanner, room *types.RoomIndexData, vnum int) {
	for {
		letter := sc.ReadLetter()
		if letter == 0 {
			return
		}

		switch letter {
		case 'S':
			return // End of room

		case 'D':
			door := sc.ReadNumber()
			if door < 0 || door > 10 {
				util.Bug("loadRooms: vnum %d bad door %d", vnum, door)
				continue
			}
			exit := &types.ExitData{
				Direction: door,
				OrigDoor:  door,
			}
			exit.Description = sc.ReadString()
			exit.Keyword = sc.ReadString()

			// Line: locks key to_room [distance [pulltype pull]]
			exitLine := sc.ReadToEOL()
			parseExitLine(exitLine, exit)

			room.Exits = append(room.Exits, exit)

		case 'E':
			ed := &types.ExtraDescrData{
				Keyword:     sc.ReadString(),
				Description: sc.ReadString(),
			}
			room.ExtraDescr = append(room.ExtraDescr, ed)

		case 'M':
			// Map data - skip for now
			sc.ReadNumber() // map vnum
			sc.ReadNumber() // x
			sc.ReadNumber() // y
			sc.ReadLetter() // entry

		case '>':
			// Room mudprog
			progType := sc.ReadWord()
			argList := sc.ReadString()
			comList := sc.ReadString()
			prog := &types.MProgData{
				Type:    mprogNameToType(progType),
				ArgList: argList,
				ComList: comList,
			}
			room.MudProgs = append(room.MudProgs, prog)
			room.ProgTypes.Set(prog.Type)
			sep := sc.ReadLetter()
			if sep != '|' {
				sc.unreadByte()
			}

		default:
			util.Bug("loadRooms: vnum %d has unexpected flag '%c'", vnum, letter)
			return
		}
	}
}

// parseExitLine parses "locks key to_room [distance [pulltype pull [x y]]]".
func parseExitLine(line string, exit *types.ExitData) {
	fields := strings.Fields(line)
	locks := 0
	if len(fields) >= 1 {
		locks = atoi(fields[0])
	}
	if len(fields) >= 2 {
		exit.Key = atoi(fields[1])
	}
	if len(fields) >= 3 {
		exit.Vnum = atoi(fields[2])
	}
	if len(fields) >= 4 {
		exit.Distance = atoi(fields[3])
	}
	if len(fields) >= 5 {
		exit.PullType = atoi(fields[4])
	}
	if len(fields) >= 6 {
		exit.Pull = atoi(fields[5])
	}

	// Convert lock values to exit flags, matching C code.
	switch locks {
	case 1:
		exit.ExitInfo = int(types.EX_ISDOOR)
	case 2:
		exit.ExitInfo = int(types.EX_ISDOOR | types.EX_PICKPROOF)
	default:
		exit.ExitInfo = locks
	}
}

// ---------------------------------------------------------------------------
// Resets
// ---------------------------------------------------------------------------

func loadResets(sc *Scanner, area *types.AreaData) {
	for {
		letter := sc.ReadLetter()
		if letter == 0 {
			return
		}
		if letter == 'S' {
			return
		}
		if letter == '*' {
			sc.ReadToEOL()
			continue
		}

		extra := sc.ReadNumber()
		arg1 := sc.ReadNumber()
		arg2 := sc.ReadNumber()
		arg3 := 0
		if letter != 'G' && letter != 'R' {
			arg3 = sc.ReadNumber()
		}
		sc.ReadToEOL()

		reset := &types.ResetData{
			Command: letter,
			Extra:   extra,
			Arg1:    arg1,
			Arg2:    arg2,
			Arg3:    arg3,
		}
		area.Resets = append(area.Resets, reset)
	}
}

// ---------------------------------------------------------------------------
// Shops
// ---------------------------------------------------------------------------

func loadShops(w *world.World, sc *Scanner) {
	for {
		keeper := sc.ReadNumber()
		if keeper == 0 {
			break
		}

		shop := &types.ShopData{Keeper: keeper}
		for i := 0; i < types.MAX_TRADE; i++ {
			shop.BuyType[i] = sc.ReadNumber()
		}
		shop.ProfitBuy = sc.ReadNumber()
		shop.ProfitSell = sc.ReadNumber()

		// Clamp values.
		if shop.ProfitBuy < shop.ProfitSell+5 {
			shop.ProfitBuy = shop.ProfitSell + 5
		}
		if shop.ProfitBuy > 1000 {
			shop.ProfitBuy = 1000
		}
		if shop.ProfitSell < 0 {
			shop.ProfitSell = 0
		}
		if shop.ProfitSell > shop.ProfitBuy-5 {
			shop.ProfitSell = shop.ProfitBuy - 5
		}

		shop.OpenHour = sc.ReadNumber()
		shop.CloseHour = sc.ReadNumber()
		sc.ReadToEOL()

		// Link shop to mob index.
		mob := w.GetMobIndex(keeper)
		if mob != nil {
			mob.Shop = shop
		} else {
			util.Bug("loadShops: mob %d not found for shop", keeper)
		}

		w.Shops = append(w.Shops, shop)
	}
}

// ---------------------------------------------------------------------------
// Repairs
// ---------------------------------------------------------------------------

func loadRepairs(w *world.World, sc *Scanner) {
	for {
		keeper := sc.ReadNumber()
		if keeper == 0 {
			break
		}

		rshop := &types.RepairData{Keeper: keeper}
		for i := 0; i < types.MAX_FIX; i++ {
			rshop.FixType[i] = sc.ReadNumber()
		}
		rshop.ProfitFix = sc.ReadNumber()
		rshop.ShopType = sc.ReadNumber()
		rshop.OpenHour = sc.ReadNumber()
		rshop.CloseHour = sc.ReadNumber()
		sc.ReadToEOL()

		mob := w.GetMobIndex(keeper)
		if mob != nil {
			mob.RShop = rshop
		} else {
			util.Bug("loadRepairs: mob %d not found for repair shop", keeper)
		}

		w.Repairs = append(w.Repairs, rshop)
	}
}

// ---------------------------------------------------------------------------
// Specials
// ---------------------------------------------------------------------------

func loadSpecials(w *world.World, sc *Scanner) {
	for {
		letter := sc.ReadLetter()
		if letter == 0 {
			return
		}

		switch letter {
		case 'S':
			return
		case '*':
			sc.ReadToEOL()
		case 'M':
			mobVnum := sc.ReadNumber()
			specFun := sc.ReadWord()
			sc.ReadToEOL()

			mob := w.GetMobIndex(mobVnum)
			if mob != nil {
				mob.SpecFun = specFun
			} else {
				util.Bug("loadSpecials: mob %d not found for spec %s", mobVnum, specFun)
			}
		default:
			util.Bug("loadSpecials: unknown letter '%c'", letter)
			sc.ReadToEOL()
		}
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// convertPosition translates old-style position codes to the current values,
// matching the C code's switch statement.
func convertPosition(pos int) int {
	if pos >= 100 {
		return pos - 100
	}
	switch pos {
	case 5:
		return 6
	case 6:
		return 8
	case 7:
		return 9
	case 8:
		return 12
	case 9:
		return 13
	case 10:
		return 14
	case 11:
		return 15
	default:
		return pos
	}
}

// capitalizeFirst upper-cases the first byte of a string if it is a letter.
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	if s[0] >= 'a' && s[0] <= 'z' {
		return string(s[0]-32) + s[1:]
	}
	return s
}

// atoi converts a string to int, returning 0 on error.
func atoi(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

// parseUint32 converts a string to uint32, returning 0 on error.
func parseUint32(s string) uint32 {
	n, _ := strconv.ParseUint(strings.TrimSpace(s), 10, 32)
	return uint32(n)
}

// mprogNameToType converts a mudprog type name like "speech_prog" to its
// MPROG_* constant bit value.
func mprogNameToType(name string) int {
	name = strings.ToLower(strings.TrimSpace(name))
	switch name {
	case "act_prog":
		return types.MPROG_ACT
	case "speech_prog":
		return types.MPROG_SPEECH
	case "rand_prog":
		return types.MPROG_RAND
	case "fight_prog":
		return types.MPROG_FIGHT
	case "death_prog":
		return types.MPROG_DEATH
	case "hitprcnt_prog":
		return types.MPROG_HITPRCNT
	case "entry_prog":
		return types.MPROG_ENTRY
	case "greet_prog":
		return types.MPROG_GREET
	case "all_greet_prog":
		return types.MPROG_ALL_GREET
	case "give_prog":
		return types.MPROG_GIVE
	case "bribe_prog":
		return types.MPROG_BRIBE
	case "hour_prog":
		return types.MPROG_HOUR
	case "time_prog":
		return types.MPROG_TIME
	case "wear_prog":
		return types.MPROG_WEAR
	case "remove_prog":
		return types.MPROG_REMOVE
	case "sac_prog":
		return types.MPROG_SAC
	case "look_prog":
		return types.MPROG_LOOK
	case "exa_prog":
		return types.MPROG_EXA
	case "zap_prog":
		return types.MPROG_ZAP
	case "get_prog":
		return types.MPROG_GET
	case "drop_prog":
		return types.MPROG_DROP
	case "damage_prog":
		return types.MPROG_DAMAGE
	case "repair_prog":
		return types.MPROG_REPAIR
	case "randiw_prog":
		return types.MPROG_RANDIW
	case "speechiw_prog":
		return types.MPROG_SPEECHIW
	case "pull_prog":
		return types.MPROG_PULL
	case "push_prog":
		return types.MPROG_PUSH
	case "sleep_prog":
		return types.MPROG_SLEEP
	case "rest_prog":
		return types.MPROG_REST
	case "leave_prog":
		return types.MPROG_LEAVE
	case "script_prog":
		return types.MPROG_SCRIPT
	case "use_prog":
		return types.MPROG_USE
	default:
		util.Bug("mprogNameToType: unknown prog type '%s'", name)
		return 0
	}
}
