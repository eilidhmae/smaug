package act

import (
	"fmt"
	"strings"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

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
