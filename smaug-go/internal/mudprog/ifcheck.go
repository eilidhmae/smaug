package mudprog

import (
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// wearLocNames mirrors C item_w_flags in build.c:147, indexed by WEAR_* location.
var wearLocNames = []string{
	"take", "finger", "neck", "neck", "neck", "body", "head", "legs", "feet",
	"hands", "arms", "shield", "about", "waist", "wrist", "wrist", "wield",
	"hold", "dual", "ears", "eyes", "missile", "back", "face", "ankle", "ankle",
}

// dirFromKeyword resolves a north/n/south/etc. keyword or a door keyword
// on any exit to an ExitData, or nil. Matches C find_door behaviour.
func dirFromKeyword(ch *types.CharData, arg string) *types.ExitData {
	if ch == nil || ch.InRoom == nil {
		return nil
	}
	arg = strings.TrimSpace(strings.ToLower(arg))
	dirNames := []string{"north", "east", "south", "west", "up", "down",
		"northeast", "northwest", "southeast", "southwest", "somewhere"}
	for i, n := range dirNames {
		if n == arg || (len(arg) == 1 && strings.HasPrefix(n, arg)) {
			return ch.InRoom.GetExit(i)
		}
	}
	// Match by keyword
	for _, ex := range ch.InRoom.Exits {
		if ex.Keyword != "" && strings.Contains(strings.ToLower(ex.Keyword), arg) {
			return ex
		}
	}
	return nil
}

// resolveObj picks out an object by variable name ($o, $p).
func resolveObj(varStr string, obj, target *types.ObjData) *types.ObjData {
	switch strings.TrimSpace(varStr) {
	case "$o":
		return obj
	case "$p":
		return target
	}
	return nil
}

// DoIfCheck evaluates a single if-check expression.
// Format: "ifchk( $var ) [operator value]"
// Returns true if the condition is met.
func DoIfCheck(check string, mob *types.CharData, actor *types.CharData,
	obj *types.ObjData, victim *types.CharData, target *types.ObjData) bool {

	check = strings.TrimSpace(check)
	if check == "" {
		return false
	}

	// Parse: "checkname($x) [op val]" or just "checkname($x)"
	parenIdx := strings.Index(check, "(")
	if parenIdx < 0 {
		return false
	}
	checkName := strings.TrimSpace(check[:parenIdx])
	rest := check[parenIdx+1:]

	closeIdx := strings.Index(rest, ")")
	if closeIdx < 0 {
		return false
	}
	argStr := strings.TrimSpace(rest[:closeIdx])
	rest = strings.TrimSpace(rest[closeIdx+1:])

	// Resolve the variable argument to a character and/or object
	chk := resolveChar(argStr, mob, actor, victim)
	chkObj := resolveObj(argStr, obj, target)

	// Parse operator and value. C copies only operator chars (==, !=, <, >, etc.)
	// into opr; everything after goes into rval. If rest starts with a
	// non-operator alphanumeric character, the entire rest becomes the value
	// with the default op (matches C where opr stays empty and rval captures all).
	// We also preserve the original space-split behaviour so that unknown-op
	// tests (e.g. "?? 20") still fall through to compareInt's default branch.
	op := "=="
	valStr := ""
	if rest != "" {
		if isOpChar(rest[0]) {
			i := 0
			for i < len(rest) && isOpChar(rest[i]) {
				i++
			}
			op = rest[:i]
			valStr = strings.TrimSpace(rest[i:])
		} else if isLetter(rest[0]) {
			// Non-operator word (e.g. direction name) — take whole rest as value.
			valStr = rest
		} else {
			// Punctuation-like tokens (e.g. "?? 20") — preserve old space-split.
			parts := strings.SplitN(rest, " ", 2)
			op = parts[0]
			if len(parts) >= 2 {
				valStr = strings.TrimSpace(parts[1])
			}
		}
	}
	val, _ := strconv.Atoi(valStr)

	switch strings.ToLower(checkName) {
	case "rand":
		return util.NumberPercent() <= val

	case "ispc":
		return chk != nil && !chk.IsNPC()
	case "isnpc":
		return chk != nil && chk.IsNPC()
	case "isevil":
		return chk != nil && chk.Alignment < -350
	case "isgood":
		return chk != nil && chk.Alignment > 350
	case "isneutral":
		return chk != nil && chk.Alignment >= -350 && chk.Alignment <= 350
	case "isimmort":
		return chk != nil && chk.IsImmortal()
	case "isfight":
		return chk != nil && chk.Fighting != nil
	case "ischarmed":
		// C IS_AFFECTED checks AFF_CHARM || AFF_POSSESS (mud_prog.c).
		return chk != nil && (chk.AffectedBy.IsSet(types.AFF_CHARM) ||
			chk.AffectedBy.IsSet(types.AFF_POSSESS))
	case "isflying":
		return chk != nil && chk.AffectedBy.IsSet(types.AFF_FLYING)
	case "isinvis":
		return chk != nil && chk.AffectedBy.IsSet(types.AFF_INVISIBLE)
	case "isaffected":
		return chk != nil && chk.AffectedBy.IsSet(val)

	case "level":
		if chk == nil {
			return false
		}
		// C mud_prog.c:1128 uses get_trust(chkchar), not raw ->level, so that
		// immortal trust-level gates ("if level($n) >= 60") read the effective
		// trust (including NPC cap and PC trust override).
		return compareInt(chk.GetTrust(), op, val)
	case "hp":
		if chk == nil {
			return false
		}
		return compareInt(chk.Hit, op, val)
	case "hppcnt":
		if chk == nil || chk.MaxHit == 0 {
			return false
		}
		return compareInt(chk.Hit*100/chk.MaxHit, op, val)
	case "mana":
		if chk == nil {
			return false
		}
		return compareInt(chk.Mana, op, val)
	case "gold":
		if chk == nil {
			return false
		}
		return compareInt(chk.Gold, op, val)
	case "sex":
		if chk == nil {
			return false
		}
		return compareInt(chk.Sex, op, val)
	case "position":
		if chk == nil {
			return false
		}
		return compareInt(chk.Position, op, val)
	case "class":
		if chk == nil {
			return false
		}
		// C mud_prog.c:1144 compares class_table[ch->class]->who_name as a
		// STRING — the common area-builder idiom is `if class($n) == mag`.
		// Fall back to numeric comparison when the value parses as an int
		// (supports legacy numeric-class progs) or the class table is missing.
		if _, numErr := strconv.Atoi(strings.TrimSpace(valStr)); numErr == nil {
			return compareInt(chk.Class, op, val)
		}
		name := classNameLookup(chk.Class)
		if name == "" {
			return false
		}
		return matchStr(name, op, valStr)
	case "race":
		if chk == nil {
			return false
		}
		// C mud_prog.c:1150 compares race_table[ch->race]->name as a STRING.
		if _, numErr := strconv.Atoi(strings.TrimSpace(valStr)); numErr == nil {
			return compareInt(chk.Race, op, val)
		}
		name := raceNameLookup(chk.Race)
		if name == "" {
			return false
		}
		return matchStr(name, op, valStr)
	case "alignment":
		if chk == nil {
			return false
		}
		return compareInt(chk.Alignment, op, val)
	case "str":
		if chk == nil {
			return false
		}
		return compareInt(chk.GetCurrStr(), op, val)
	case "int":
		if chk == nil {
			return false
		}
		return compareInt(chk.GetCurrInt(), op, val)
	case "wis":
		if chk == nil {
			return false
		}
		return compareInt(chk.GetCurrWis(), op, val)
	case "dex":
		if chk == nil {
			return false
		}
		return compareInt(chk.GetCurrDex(), op, val)
	case "con":
		if chk == nil {
			return false
		}
		return compareInt(chk.GetCurrCon(), op, val)

	case "name":
		if chk == nil {
			return false
		}
		return matchStr(chk.Name, op, valStr)

	case "isinroom":
		if mob == nil || mob.InRoom == nil {
			return false
		}
		return mob.InRoom.Vnum == val

	// ---------- Priority A: obj/mob lookups (count-based) ----------
	case "mobinroom":
		if mob == nil || mob.InRoom == nil {
			return false
		}
		vnum, _ := strconv.Atoi(argStr)
		count := 0
		for _, p := range mob.InRoom.People {
			if p.IsNPC() && p.IndexData != nil && p.IndexData.Vnum == vnum {
				count++
			}
		}
		return compareInt(count, op, val)

	case "mobinarea":
		if mob == nil || mob.InRoom == nil || WorldRef == nil {
			return false
		}
		vnum, _ := strconv.Atoi(argStr)
		count := 0
		for _, c := range WorldRef.Characters {
			if c.IsNPC() && c.IndexData != nil && c.IndexData.Vnum == vnum &&
				c.InRoom != nil && c.InRoom.Area == mob.InRoom.Area {
				count++
			}
		}
		return compareInt(count, op, val)

	case "mobinworld":
		if WorldRef == nil {
			return false
		}
		vnum, _ := strconv.Atoi(argStr)
		count := 0
		for _, c := range WorldRef.Characters {
			if c.IsNPC() && c.IndexData != nil && c.IndexData.Vnum == vnum {
				count++
			}
		}
		return compareInt(count, op, val)

	case "objinworld":
		if WorldRef == nil {
			return false
		}
		vnum, _ := strconv.Atoi(argStr)
		count := 0
		for _, o := range WorldRef.Objects {
			if o.IndexData != nil && o.IndexData.Vnum == vnum {
				count++
			}
		}
		return compareInt(count, op, val)

	case "ovnumhere":
		if mob == nil || mob.InRoom == nil {
			return false
		}
		vnum, _ := strconv.Atoi(argStr)
		count := 0
		for _, o := range mob.Carrying {
			if o.IndexData != nil && o.IndexData.Vnum == vnum {
				count += maxInt(1, o.Count)
			}
		}
		for _, o := range mob.InRoom.Contents {
			if o.IndexData != nil && o.IndexData.Vnum == vnum {
				count += maxInt(1, o.Count)
			}
		}
		return compareInt(count, op, val)

	case "otypehere":
		if mob == nil || mob.InRoom == nil {
			return false
		}
		typ, _ := strconv.Atoi(argStr)
		count := 0
		for _, o := range mob.Carrying {
			if o.ItemType == typ {
				count += maxInt(1, o.Count)
			}
		}
		for _, o := range mob.InRoom.Contents {
			if o.ItemType == typ {
				count += maxInt(1, o.Count)
			}
		}
		return compareInt(count, op, val)

	case "ovnumroom":
		if mob == nil || mob.InRoom == nil {
			return false
		}
		vnum, _ := strconv.Atoi(argStr)
		count := 0
		for _, o := range mob.InRoom.Contents {
			if o.IndexData != nil && o.IndexData.Vnum == vnum {
				count += maxInt(1, o.Count)
			}
		}
		return compareInt(count, op, val)

	case "otyperoom":
		if mob == nil || mob.InRoom == nil {
			return false
		}
		typ, _ := strconv.Atoi(argStr)
		count := 0
		for _, o := range mob.InRoom.Contents {
			if o.ItemType == typ {
				count += maxInt(1, o.Count)
			}
		}
		return compareInt(count, op, val)

	case "ovnumcarry":
		if mob == nil {
			return false
		}
		vnum, _ := strconv.Atoi(argStr)
		count := 0
		for _, o := range mob.Carrying {
			if o.IndexData != nil && o.IndexData.Vnum == vnum {
				count += maxInt(1, o.Count)
			}
		}
		return compareInt(count, op, val)

	case "otypecarry":
		if mob == nil {
			return false
		}
		typ, _ := strconv.Atoi(argStr)
		count := 0
		for _, o := range mob.Carrying {
			if o.ItemType == typ {
				count += maxInt(1, o.Count)
			}
		}
		return compareInt(count, op, val)

	case "ovnumwear":
		if mob == nil {
			return false
		}
		vnum, _ := strconv.Atoi(argStr)
		count := 0
		for _, o := range mob.Carrying {
			if o.WearLoc != types.WEAR_NONE && o.IndexData != nil && o.IndexData.Vnum == vnum {
				count += maxInt(1, o.Count)
			}
		}
		return compareInt(count, op, val)

	case "otypewear":
		if mob == nil {
			return false
		}
		typ, _ := strconv.Atoi(argStr)
		count := 0
		for _, o := range mob.Carrying {
			if o.WearLoc != types.WEAR_NONE && o.ItemType == typ {
				count += maxInt(1, o.Count)
			}
		}
		return compareInt(count, op, val)

	case "ovnuminv":
		if mob == nil {
			return false
		}
		vnum, _ := strconv.Atoi(argStr)
		count := 0
		for _, o := range mob.Carrying {
			if o.WearLoc == types.WEAR_NONE && o.IndexData != nil && o.IndexData.Vnum == vnum {
				count += maxInt(1, o.Count)
			}
		}
		return compareInt(count, op, val)

	case "otypeinv":
		if mob == nil {
			return false
		}
		typ, _ := strconv.Atoi(argStr)
		count := 0
		for _, o := range mob.Carrying {
			if o.WearLoc == types.WEAR_NONE && o.ItemType == typ {
				count += maxInt(1, o.Count)
			}
		}
		return compareInt(count, op, val)

	case "objval0":
		if chkObj == nil {
			return false
		}
		return compareInt(chkObj.Value[0], op, val)
	case "objval1":
		if chkObj == nil {
			return false
		}
		return compareInt(chkObj.Value[1], op, val)
	case "objval2":
		if chkObj == nil {
			return false
		}
		return compareInt(chkObj.Value[2], op, val)
	case "objval3":
		if chkObj == nil {
			return false
		}
		return compareInt(chkObj.Value[3], op, val)
	case "objval4":
		if chkObj == nil {
			return false
		}
		return compareInt(chkObj.Value[4], op, val)
	case "objval5":
		if chkObj == nil {
			return false
		}
		return compareInt(chkObj.Value[5], op, val)

	case "wearing":
		if chk == nil {
			return false
		}
		target := strings.ToLower(strings.TrimSpace(valStr))
		for _, o := range chk.Carrying {
			if o.CarriedBy == chk && o.WearLoc > types.WEAR_NONE && o.WearLoc < len(wearLocNames) {
				if strings.EqualFold(target, wearLocNames[o.WearLoc]) {
					return true
				}
			}
		}
		return false

	case "wearingvnum":
		if chk == nil {
			return false
		}
		vnum, err := strconv.Atoi(strings.TrimSpace(valStr))
		if err != nil {
			return false
		}
		for _, o := range chk.Carrying {
			if o.CarriedBy == chk && o.WearLoc > types.WEAR_NONE &&
				o.IndexData != nil && o.IndexData.Vnum == vnum {
				return true
			}
		}
		return false

	case "carryingvnum":
		if chk == nil {
			return false
		}
		vnum, err := strconv.Atoi(strings.TrimSpace(valStr))
		if err != nil {
			return false
		}
		return carryingVnumVisit(chk, chk.Carrying, vnum)

	// ---------- Priority B: char state ----------
	case "cansee":
		if chk == nil {
			return false
		}
		return canSeeIfCheck(mob, chk)

	case "ispacifist":
		return chk != nil && chk.IsNPC() && chk.Act.IsSet(types.ACT_PACIFIST)

	case "isriding":
		return chk != nil && chk.Mount == mob

	case "ismounted":
		return chk != nil && chk.Position == types.POS_MOUNTED

	case "ismorphed":
		return chk != nil && chk.Morph != nil

	case "isnuisance":
		return chk != nil && !chk.IsNPC() && chk.PCData != nil && chk.PCData.Nuisance != nil

	case "ispkill":
		if chk == nil || chk.IsNPC() || chk.PCData == nil {
			return false
		}
		return (uint32(chk.PCData.Flags) & types.PCFLAG_DEADLY) != 0

	case "isthief":
		return chk != nil && !chk.IsNPC() && chk.Act.IsSet(types.PLR_THIEF)

	case "isattacker":
		return chk != nil && !chk.IsNPC() && chk.Act.IsSet(types.PLR_ATTACKER)

	case "iskiller":
		return chk != nil && !chk.IsNPC() && chk.Act.IsSet(types.PLR_KILLER)

	case "ismobinvis":
		return chk != nil && chk.IsNPC() && chk.Act.IsSet(types.ACT_MOBINVIS)

	case "mobinvislevel":
		if chk == nil || !chk.IsNPC() {
			return false
		}
		return compareInt(chk.MobInvis, op, val)

	case "drunk":
		if chk == nil || chk.IsNPC() || chk.PCData == nil {
			return false
		}
		return compareInt(chk.PCData.Condition[types.COND_DRUNK], op, val)

	case "hostdesc":
		if chk == nil || chk.IsNPC() || chk.Desc == nil || chk.Desc.Host == "" {
			return false
		}
		return matchStr(chk.Desc.Host, op, valStr)

	case "waitstate":
		if chk == nil || chk.IsNPC() || chk.Wait == 0 {
			return false
		}
		return compareInt(chk.Wait, op, val)

	case "favor":
		if chk == nil || chk.IsNPC() || chk.PCData == nil || chk.PCData.Favor == 0 {
			return false
		}
		return compareInt(chk.PCData.Favor, op, val)

	case "hps":
		if chk == nil {
			return false
		}
		return compareInt(chk.Hit, op, val)

	case "lck":
		if chk == nil {
			return false
		}
		return compareInt(chk.GetCurrLck(), op, val)

	case "cha":
		if chk == nil {
			return false
		}
		return compareInt(chk.GetCurrCha(), op, val)

	case "numfighting":
		if chk == nil {
			return false
		}
		return compareInt(chk.NumFighting-1, op, val)

	case "weight":
		if chk == nil {
			return false
		}
		return compareInt(chk.CarryWeight, op, val)

	// ---------- Priority C: room/area/exit state ----------
	case "inroom":
		if chk == nil || chk.InRoom == nil {
			return false
		}
		return compareInt(chk.InRoom.Vnum, op, val)

	case "wasinroom":
		if chk == nil || chk.WasInRoom == nil {
			return false
		}
		return compareInt(chk.WasInRoom.Vnum, op, val)

	case "inarea":
		if chk == nil || chk.InRoom == nil || chk.InRoom.Area == nil {
			return false
		}
		return matchStr(chk.InRoom.Area.Filename, op, valStr)

	case "indoors":
		if chk == nil || chk.InRoom == nil {
			return false
		}
		// C IS_OUTSIDE returns false when ROOM_INDOORS flag is set, so a room
		// is indoors if it has the flag OR is sector INSIDE.
		return chk.InRoom.RoomFlags.IsSet(types.ROOM_INDOORS) ||
			chk.InRoom.SectorType == types.SECT_INSIDE

	case "nomagic":
		return chk != nil && chk.InRoom != nil && chk.InRoom.RoomFlags.IsSet(types.ROOM_NO_MAGIC)
	case "safe":
		return chk != nil && chk.InRoom != nil && chk.InRoom.RoomFlags.IsSet(types.ROOM_SAFE)
	case "nosummon":
		return chk != nil && chk.InRoom != nil && chk.InRoom.RoomFlags.IsSet(types.ROOM_NO_SUMMON)
	case "noastral":
		return chk != nil && chk.InRoom != nil && chk.InRoom.RoomFlags.IsSet(types.ROOM_NO_ASTRAL)
	case "nosupplicate":
		return chk != nil && chk.InRoom != nil && chk.InRoom.RoomFlags.IsSet(types.ROOM_NOSUPPLICATE)
	case "norecall":
		return chk != nil && chk.InRoom != nil && chk.InRoom.RoomFlags.IsSet(types.ROOM_NO_RECALL)

	case "ispassage":
		return dirFromKeyword(chk, valStr) != nil
	case "isopen":
		ex := dirFromKeyword(chk, valStr)
		if ex == nil {
			return false
		}
		return (uint32(ex.ExitInfo) & types.EX_CLOSED) == 0
	case "islocked":
		ex := dirFromKeyword(chk, valStr)
		if ex == nil {
			return false
		}
		return (uint32(ex.ExitInfo) & types.EX_LOCKED) != 0

	// ---------- Priority D: social/political ----------
	case "clan":
		if chk == nil || chk.IsNPC() || chk.PCData == nil || chk.PCData.Clan == nil {
			return false
		}
		return matchStr(chk.PCData.Clan.Name, op, valStr)

	case "council":
		if chk == nil || chk.IsNPC() || chk.PCData == nil || chk.PCData.Council == nil {
			return false
		}
		return matchStr(chk.PCData.Council.Name, op, valStr)

	case "deity":
		if chk == nil || chk.IsNPC() || chk.PCData == nil || chk.PCData.Deity == nil {
			return false
		}
		return matchStr(chk.PCData.Deity.Name, op, valStr)

	case "guild":
		// C checks IS_GUILDED(chkchar); we map to clan with ClanType indicating guild.
		// Fall back: if clan present, compare clan name (mirrors C code path).
		if chk == nil || chk.IsNPC() || chk.PCData == nil || chk.PCData.Clan == nil {
			return false
		}
		return matchStr(chk.PCData.Clan.Name, op, valStr)

	case "clantype":
		if chk == nil || chk.IsNPC() || chk.PCData == nil || chk.PCData.Clan == nil {
			return false
		}
		return compareInt(chk.PCData.Clan.ClanType, op, val)

	case "isleader":
		if chk == nil || chk.IsNPC() || WorldRef == nil {
			return false
		}
		clan := findClanByName(valStr)
		if clan == nil {
			return false
		}
		return strings.EqualFold(chk.Name, clan.Leader) ||
			strings.EqualFold(chk.Name, clan.Number1) ||
			strings.EqualFold(chk.Name, clan.Number2)

	case "isclanleader":
		if chk == nil || chk.IsNPC() || WorldRef == nil {
			return false
		}
		clan := findClanByName(valStr)
		if clan == nil {
			return false
		}
		return strings.EqualFold(chk.Name, clan.Leader)

	case "isclan1":
		if chk == nil || chk.IsNPC() || WorldRef == nil {
			return false
		}
		clan := findClanByName(valStr)
		if clan == nil {
			return false
		}
		return strings.EqualFold(chk.Name, clan.Number1)

	case "isclan2":
		if chk == nil || chk.IsNPC() || WorldRef == nil {
			return false
		}
		clan := findClanByName(valStr)
		if clan == nil {
			return false
		}
		return strings.EqualFold(chk.Name, clan.Number2)

	case "isdevoted":
		return chk != nil && !chk.IsNPC() && chk.PCData != nil && chk.PCData.Deity != nil

	case "rank":
		if chk == nil || chk.IsNPC() || chk.PCData == nil {
			return false
		}
		return matchStr(chk.PCData.Rank, op, valStr)

	// ---------- Priority E: misc ----------
	case "economy":
		// `economy(idx)` where idx == 0 means mob's current room.
		idx, _ := strconv.Atoi(argStr)
		var room *types.RoomIndexData
		if idx == 0 {
			if mob == nil || mob.InRoom == nil {
				return false
			}
			room = mob.InRoom
		} else if WorldRef != nil {
			room = WorldRef.GetRoom(idx)
		}
		if room == nil || room.Area == nil {
			return false
		}
		lhs := room.Area.LowEconomy
		if room.Area.HighEconomy > 0 {
			lhs += 1000000000
		}
		return compareInt(lhs, op, val)

	case "time":
		if WorldRef == nil {
			return false
		}
		return compareInt(WorldRef.TimeInfo.Hour, op, val)

	case "number":
		if chk != nil {
			if !chk.IsNPC() {
				return false
			}
			var lhs int
			if chk == mob {
				lhs = chk.Gold
			} else if chk.IndexData != nil {
				lhs = chk.IndexData.Vnum
			}
			return compareInt(lhs, op, val)
		}
		if chkObj != nil && chkObj.IndexData != nil {
			return compareInt(chkObj.IndexData.Vnum, op, val)
		}
		return false

	case "mortinroom":
		// TODO(tier4): C uses nifty_is_name which is a whitespace-split
		// prefix match against the name list; here we only do exact-name
		// compare. Swap to util.IsName when/if a prefix-aware variant is
		// ported (mirrors C is_name_prefix).
		if mob == nil || mob.InRoom == nil {
			return false
		}
		for _, p := range mob.InRoom.People {
			if !p.IsNPC() && p.GetTrust() < types.LEVEL_IMMORTAL &&
				strings.EqualFold(p.Name, argStr) {
				return true
			}
		}
		return false

	case "mortinarea":
		if mob == nil || mob.InRoom == nil || WorldRef == nil {
			return false
		}
		for _, c := range WorldRef.Characters {
			if !c.IsNPC() && c.InRoom != nil && c.InRoom.Area == mob.InRoom.Area &&
				c.GetTrust() < types.LEVEL_IMMORTAL && strings.EqualFold(c.Name, argStr) {
				return true
			}
		}
		return false

	case "mortinworld":
		// TODO(tier4): same nifty_is_name prefix-match gap as mortinroom.
		if WorldRef == nil {
			return false
		}
		for _, d := range WorldRef.Descriptors {
			if d.Connected == types.CON_PLAYING && d.Character != nil &&
				d.Character.GetTrust() < types.LEVEL_IMMORTAL &&
				strings.EqualFold(d.Character.Name, argStr) {
				return true
			}
		}
		return false

	case "mortcount":
		room := roomForCountCheck(mob, argStr)
		if room == nil {
			return false
		}
		count := 0
		for _, p := range room.People {
			if !p.IsNPC() && p.GetTrust() < types.LEVEL_IMMORTAL {
				count++
			}
		}
		return compareInt(count, op, val)

	case "mobcount":
		room := roomForCountCheck(mob, argStr)
		if room == nil {
			return false
		}
		// C starts count at -1
		count := -1
		for _, p := range room.People {
			if p.IsNPC() {
				count++
			}
		}
		return compareInt(count, op, val)

	case "charcount":
		room := roomForCountCheck(mob, argStr)
		if room == nil {
			return false
		}
		// C starts count at -1; counts NPCs and non-immortal PCs.
		count := -1
		for _, p := range room.People {
			if (!p.IsNPC() && p.GetTrust() < types.LEVEL_IMMORTAL) || p.IsNPC() {
				count++
			}
		}
		return compareInt(count, op, val)

	// TODO(tier3): timeskilled — needs MobIndexData.Killed counter
	// TODO(tier3): leverpos — needs lever/switch trigger data on ObjData (Value[0] bitflag TRIG_UP)
	// TODO(tier3): isflagged / istagged — needs VariableData get_tag implementation
	// TODO(tier3): pkadrenalized / asupressed — handler.GetTimer(TIMER_PKADRENALINE / TIMER_ASUPRESSED) > 0 / timer.Value == -1; subsystem now present (handler/timer.go), callers to be wired in a follow-up
	// TODO(tier3): areamulti / multi — require host-descriptor matching across all chars (host comparison is implemented but requires every PC to have a descriptor; untested without integration fixture)
	// TODO(tier3): objtype — duplicate of obj type check; rarely used in stock areas

	default:
		return false
	}
}

// isOpChar reports whether c is a C-style operator character
// (==, !=, <, >, <=, >=, /, !/). Mirrors C isoperator().
func isOpChar(c byte) bool {
	switch c {
	case '=', '!', '<', '>', '/':
		return true
	}
	return false
}

// isLetter reports whether c starts a word token (a letter or digit).
func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// maxInt returns the larger of a and b.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// carryingVnumVisit walks an object list recursively looking for a vnum match.
// Mirrors C carryingvnum_visit.
func carryingVnumVisit(ch *types.CharData, list []*types.ObjData, vnum int) bool {
	for _, o := range list {
		if o.IndexData != nil && o.IndexData.Vnum == vnum {
			return true
		}
		if len(o.Contents) > 0 {
			if carryingVnumVisit(ch, o.Contents, vnum) {
				return true
			}
		}
	}
	return false
}

// roomForCountCheck returns the room for count-style ifchecks. If cvar is 0
// or empty, uses the mob's current room.
func roomForCountCheck(mob *types.CharData, cvar string) *types.RoomIndexData {
	rvnum, _ := strconv.Atoi(strings.TrimSpace(cvar))
	if rvnum == 0 {
		if mob != nil {
			return mob.InRoom
		}
		return nil
	}
	if WorldRef == nil {
		return nil
	}
	return WorldRef.GetRoom(rvnum)
}

// canSeeIfCheck is a trimmed visibility check used by the cansee ifcheck.
// Mirrors C can_see (handler.c:3388-3409): HOLYLIGHT short-circuits; blindness
// and room-darkness (unless AFF_INFRARED) block sight; wizinvis, invisible,
// and hide follow their C gating rules.
func canSeeIfCheck(observer, target *types.CharData) bool {
	if observer == nil || target == nil {
		return true
	}
	if observer == target {
		return true
	}
	if observer.Act.IsSet(types.PLR_HOLYLIGHT) {
		return true
	}
	// C handler.c:3394: AFF_BLIND defeats sight (unless AFF_TRUESIGHT, which
	// the Go port has not yet wired through).
	if observer.AffectedBy.IsSet(types.AFF_BLIND) {
		return false
	}
	// C handler.c:3397: dark room without infrared defeats sight.
	if observer.InRoom != nil && roomIsDark(observer.InRoom) &&
		!observer.AffectedBy.IsSet(types.AFF_INFRARED) {
		return false
	}
	if target.PCData != nil && target.PCData.WizInvis > 0 {
		if observer.GetTrust() < target.PCData.WizInvis {
			return false
		}
	}
	if target.AffectedBy.IsSet(types.AFF_INVISIBLE) &&
		!observer.AffectedBy.IsSet(types.AFF_DETECT_INVIS) {
		return false
	}
	if target.AffectedBy.IsSet(types.AFF_HIDE) &&
		!observer.AffectedBy.IsSet(types.AFF_DETECT_HIDDEN) &&
		target.Fighting == nil {
		return false
	}
	return true
}

// roomIsDark mirrors C room_is_dark (handler.c:3183): explicit light overrides
// a dark flag; otherwise ROOM_DARK or SUN_SET/SUN_DARK (outside INSIDE/CITY)
// darken the room.
func roomIsDark(room *types.RoomIndexData) bool {
	if room == nil {
		return false
	}
	if room.Light > 0 {
		return false
	}
	if room.RoomFlags.IsSet(types.ROOM_LIGHT) {
		return false
	}
	if room.RoomFlags.IsSet(types.ROOM_DARK) {
		return true
	}
	if room.SectorType == types.SECT_INSIDE || room.SectorType == types.SECT_CITY {
		return false
	}
	if WorldRef != nil && (WorldRef.TimeInfo.Sunlight == types.SUN_SET ||
		WorldRef.TimeInfo.Sunlight == types.SUN_DARK) {
		return true
	}
	return false
}

// classNameLookup resolves a class index to its WhoName via WorldRef.Classes.
// Returns "" if the index is out of range or the table is not loaded.
func classNameLookup(idx int) string {
	if WorldRef == nil {
		return ""
	}
	if idx < 0 || idx >= len(WorldRef.Classes) {
		return ""
	}
	c := WorldRef.Classes[idx]
	if c == nil {
		return ""
	}
	return c.WhoName
}

// raceNameLookup resolves a race index to its Name via WorldRef.Races.
// Returns "" if the index is out of range or the table is not loaded.
func raceNameLookup(idx int) string {
	if WorldRef == nil {
		return ""
	}
	if idx < 0 || idx >= len(WorldRef.Races) {
		return ""
	}
	r := WorldRef.Races[idx]
	if r == nil {
		return ""
	}
	return r.Name
}

// findClanByName looks up a clan by name across WorldRef.
func findClanByName(name string) *types.ClanData {
	if WorldRef == nil {
		return nil
	}
	name = strings.TrimSpace(name)
	for _, c := range WorldRef.Clans {
		if strings.EqualFold(c.Name, name) {
			return c
		}
	}
	return nil
}

func resolveChar(varStr string, mob, actor, victim *types.CharData) *types.CharData {
	varStr = strings.TrimSpace(varStr)
	switch varStr {
	case "$i":
		return mob
	case "$n":
		return actor
	case "$t":
		return victim
	case "$r":
		if mob != nil && mob.InRoom != nil && len(mob.InRoom.People) > 0 {
			return mob.InRoom.People[0]
		}
		return nil
	default:
		return actor
	}
}

func compareInt(lhs int, op string, rhs int) bool {
	switch op {
	case "==", "=":
		return lhs == rhs
	case "!=":
		return lhs != rhs
	case ">":
		return lhs > rhs
	case "<":
		return lhs < rhs
	case ">=":
		return lhs >= rhs
	case "<=":
		return lhs <= rhs
	default:
		return lhs == rhs
	}
}

func matchStr(lhs, op, rhs string) bool {
	switch op {
	case "==", "=":
		return strings.EqualFold(lhs, rhs)
	case "!=":
		return !strings.EqualFold(lhs, rhs)
	case "/":
		return strings.Contains(strings.ToLower(lhs), strings.ToLower(rhs))
	case "!/":
		return !strings.Contains(strings.ToLower(lhs), strings.ToLower(rhs))
	default:
		return strings.EqualFold(lhs, rhs)
	}
}
