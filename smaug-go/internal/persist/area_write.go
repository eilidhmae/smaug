package persist

import (
	"fmt"
	"io"
	"sort"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// SaveArea writes an area to a writer in SMAUG .are format.
func SaveArea(w io.Writer, wld *world.World, area *types.AreaData) error {
	// #AREA
	fmt.Fprintf(w, "#AREA\n%s~\n\n", area.Name)

	// #AUTHOR
	if area.Author != "" {
		fmt.Fprintf(w, "#AUTHOR\n%s~\n\n", area.Author)
	}

	// #RANGES
	fmt.Fprintf(w, "#RANGES\n%d %d %d %d\n$\n\n",
		area.LowSoftRange, area.HiSoftRange, area.LowHardRange, area.HiHardRange)

	// #RESETMSG
	if area.ResetMsg != "" {
		fmt.Fprintf(w, "#RESETMSG\n%s~\n\n", area.ResetMsg)
	}

	// #FLAGS
	fmt.Fprintf(w, "#FLAGS\n%d\n\n", area.Flags)

	// #ECONOMY
	fmt.Fprintf(w, "#ECONOMY\n%d %d\n\n", area.HighEconomy, area.LowEconomy)

	// #VERSION
	if area.Version > 0 {
		fmt.Fprintf(w, "#VERSION\n%d\n\n", area.Version)
	}

	// #MOBILES
	if err := saveMobiles(w, wld, area); err != nil {
		return err
	}

	// #OBJECTS
	if err := saveObjects(w, wld, area); err != nil {
		return err
	}

	// #ROOMS
	if err := saveRooms(w, wld, area); err != nil {
		return err
	}

	// #RESETS
	if err := saveResets(w, area); err != nil {
		return err
	}

	// #SHOPS
	if err := saveShops(w, wld, area); err != nil {
		return err
	}

	// End marker
	fmt.Fprintf(w, "#$\n")

	return nil
}

func saveMobiles(w io.Writer, wld *world.World, area *types.AreaData) error {
	fmt.Fprintf(w, "#MOBILES\n")

	// Collect mob vnums in this area's range, sorted
	vnums := make([]int, 0)
	for vnum := range wld.MobIndex {
		if vnum >= area.LowMVnum && vnum <= area.HiMVnum {
			vnums = append(vnums, vnum)
		}
	}
	sort.Ints(vnums)

	for _, vnum := range vnums {
		idx := wld.MobIndex[vnum]
		fmt.Fprintf(w, "#%d\n", vnum)
		fmt.Fprintf(w, "%s~\n", idx.PlayerName)
		fmt.Fprintf(w, "%s~\n", idx.ShortDescr)
		fmt.Fprintf(w, "%s~\n", idx.LongDescr)
		fmt.Fprintf(w, "%s~\n", idx.Description)
		fmt.Fprintf(w, "%s %s %d S\n",
			idx.Act.String(), idx.AffectedBy.String(), idx.Alignment)
		fmt.Fprintf(w, "%d %d %d %dd%d+%d %dd%d+%d\n",
			idx.Level, idx.MobThac0, idx.AC,
			idx.HitNoDice, idx.HitSizeDice, idx.HitPlus,
			idx.DamNoDice, idx.DamSizeDice, idx.DamPlus)
		fmt.Fprintf(w, "%d 0 0\n", idx.Gold)
		fmt.Fprintf(w, "%d %d %d\n", idx.Position, idx.DefPosition, idx.Sex)

		// Save mudprogs
		for _, prog := range idx.MudProgs {
			fmt.Fprintf(w, ">%d %s~\n%s~\n|\n", prog.Type, prog.ArgList, prog.ComList)
		}
	}

	fmt.Fprintf(w, "#0\n\n")
	return nil
}

func saveObjects(w io.Writer, wld *world.World, area *types.AreaData) error {
	fmt.Fprintf(w, "#OBJECTS\n")

	vnums := make([]int, 0)
	for vnum := range wld.ObjIndex {
		if vnum >= area.LowOVnum && vnum <= area.HiOVnum {
			vnums = append(vnums, vnum)
		}
	}
	sort.Ints(vnums)

	for _, vnum := range vnums {
		idx := wld.ObjIndex[vnum]
		fmt.Fprintf(w, "#%d\n", vnum)
		fmt.Fprintf(w, "%s~\n", idx.Name)
		fmt.Fprintf(w, "%s~\n", idx.ShortDescr)
		fmt.Fprintf(w, "%s~\n", idx.Description)
		fmt.Fprintf(w, "%s~\n", idx.ActionDesc)
		fmt.Fprintf(w, "%d %s %d\n", idx.ItemType, idx.ExtraFlags.String(), idx.WearFlags)
		fmt.Fprintf(w, "%d %d %d %d %d %d\n",
			idx.Value[0], idx.Value[1], idx.Value[2],
			idx.Value[3], idx.Value[4], idx.Value[5])
		fmt.Fprintf(w, "%d %d 0\n", idx.Weight, idx.GoldCost)

		// Extra descriptions
		for _, ed := range idx.ExtraDescr {
			fmt.Fprintf(w, "E\n%s~\n%s~\n", ed.Keyword, ed.Description)
		}

		// Affects
		for _, aff := range idx.Affects {
			fmt.Fprintf(w, "A\n%d %d\n", aff.Location, aff.Modifier)
		}

		// Mudprogs
		for _, prog := range idx.MudProgs {
			fmt.Fprintf(w, ">%d %s~\n%s~\n|\n", prog.Type, prog.ArgList, prog.ComList)
		}
	}

	fmt.Fprintf(w, "#0\n\n")
	return nil
}

func saveRooms(w io.Writer, wld *world.World, area *types.AreaData) error {
	fmt.Fprintf(w, "#ROOMS\n")

	vnums := make([]int, 0)
	for vnum, room := range wld.Rooms {
		if room.Area == area {
			vnums = append(vnums, vnum)
		}
	}
	sort.Ints(vnums)

	for _, vnum := range vnums {
		room := wld.Rooms[vnum]
		fmt.Fprintf(w, "#%d\n", vnum)
		fmt.Fprintf(w, "%s~\n", room.Name)
		fmt.Fprintf(w, "%s~\n", room.Description)
		fmt.Fprintf(w, "0 %s %d\n", room.RoomFlags.String(), room.SectorType)

		// Exits
		for _, exit := range room.Exits {
			if exit == nil {
				continue
			}
			destVnum := 0
			if exit.ToRoom != nil {
				destVnum = exit.ToRoom.Vnum
			} else {
				destVnum = exit.RVnum
			}
			fmt.Fprintf(w, "D%d\n", exit.Direction)
			fmt.Fprintf(w, "%s~\n", exit.Description)
			fmt.Fprintf(w, "%s~\n", exit.Keyword)
			fmt.Fprintf(w, "%d %d %d\n", exit.ExitInfo, exit.Key, destVnum)
		}

		// Extra descriptions
		for _, ed := range room.ExtraDescr {
			fmt.Fprintf(w, "E\n%s~\n%s~\n", ed.Keyword, ed.Description)
		}

		// Mudprogs
		for _, prog := range room.MudProgs {
			fmt.Fprintf(w, ">%d %s~\n%s~\n|\n", prog.Type, prog.ArgList, prog.ComList)
		}

		fmt.Fprintf(w, "S\n")
	}

	fmt.Fprintf(w, "#0\n\n")
	return nil
}

func saveResets(w io.Writer, area *types.AreaData) error {
	fmt.Fprintf(w, "#RESETS\n")

	for _, reset := range area.Resets {
		fmt.Fprintf(w, "%c %d %d %d %d\n",
			reset.Command, reset.Extra, reset.Arg1, reset.Arg2, reset.Arg3)
	}

	fmt.Fprintf(w, "S\n\n")
	return nil
}

func saveShops(w io.Writer, wld *world.World, area *types.AreaData) error {
	fmt.Fprintf(w, "#SHOPS\n")

	for vnum := area.LowMVnum; vnum <= area.HiMVnum; vnum++ {
		idx, ok := wld.MobIndex[vnum]
		if !ok || idx.Shop == nil {
			continue
		}
		shop := idx.Shop
		fmt.Fprintf(w, "%d ", vnum)
		for i := 0; i < types.MAX_TRADE; i++ {
			fmt.Fprintf(w, "%d ", shop.BuyType[i])
		}
		fmt.Fprintf(w, "%d %d %d %d\n",
			shop.ProfitBuy, shop.ProfitSell, shop.OpenHour, shop.CloseHour)
	}

	fmt.Fprintf(w, "0\n\n")
	return nil
}
