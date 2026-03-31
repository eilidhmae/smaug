package handler

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// GetCharRoom finds a character in the same room by name.
// Supports "N.name" syntax for finding the Nth match.
// Two-phase search: exact match first, then prefix match.
func GetCharRoom(ch *types.CharData, argument string) *types.CharData {
	if ch.InRoom == nil {
		return nil
	}

	number, arg := util.NumberArgument(argument)
	if strings.EqualFold(arg, "self") {
		return ch
	}

	// Phase 1: exact match
	count := 0
	for _, rch := range ch.InRoom.People {
		if util.IsNameExact(arg, rch.Name) {
			count++
			if count == number {
				return rch
			}
		}
	}

	// Phase 2: prefix match
	count = 0
	for _, rch := range ch.InRoom.People {
		if util.IsName(arg, rch.Name) {
			count++
			if count == number {
				return rch
			}
		}
	}

	return nil
}

// GetCharWorld finds a character anywhere in the game world by name.
// Checks current room first, then searches the full world character list.
// Two-phase search: exact match first, then prefix match.
func GetCharWorld(w *world.World, ch *types.CharData, argument string) *types.CharData {
	// Try room first
	if found := GetCharRoom(ch, argument); found != nil {
		return found
	}

	number, arg := util.NumberArgument(argument)
	if strings.EqualFold(arg, "self") {
		return ch
	}

	// Phase 1: exact match in world
	count := 0
	for _, wch := range w.Characters {
		if util.IsNameExact(arg, wch.Name) {
			count++
			if count == number {
				return wch
			}
		}
	}

	// Phase 2: prefix match in world
	count = 0
	for _, wch := range w.Characters {
		if util.IsName(arg, wch.Name) {
			count++
			if count == number {
				return wch
			}
		}
	}

	return nil
}

// GetObjCarry finds an object in a character's inventory (not equipped).
// Supports "N.name" syntax.
// Two-phase search: exact match first, then prefix match.
func GetObjCarry(ch *types.CharData, argument string) *types.ObjData {
	number, arg := util.NumberArgument(argument)

	// Phase 1: exact match
	count := 0
	for _, obj := range ch.Carrying {
		if obj.WearLoc != types.WEAR_NONE {
			continue
		}
		if util.IsNameExact(arg, obj.Name) {
			count++
			if count == number {
				return obj
			}
		}
	}

	// Phase 2: prefix match
	count = 0
	for _, obj := range ch.Carrying {
		if obj.WearLoc != types.WEAR_NONE {
			continue
		}
		if util.IsName(arg, obj.Name) {
			count++
			if count == number {
				return obj
			}
		}
	}

	return nil
}

// GetObjWear finds an object in a character's equipped slots.
// Supports "N.name" syntax.
// Two-phase search: exact match first, then prefix match.
func GetObjWear(ch *types.CharData, argument string) *types.ObjData {
	number, arg := util.NumberArgument(argument)

	// Phase 1: exact match
	count := 0
	for _, obj := range ch.Carrying {
		if obj.WearLoc == types.WEAR_NONE {
			continue
		}
		if util.IsNameExact(arg, obj.Name) {
			count++
			if count == number {
				return obj
			}
		}
	}

	// Phase 2: prefix match
	count = 0
	for _, obj := range ch.Carrying {
		if obj.WearLoc == types.WEAR_NONE {
			continue
		}
		if util.IsName(arg, obj.Name) {
			count++
			if count == number {
				return obj
			}
		}
	}

	return nil
}

// GetObjHere finds an object in the room, character inventory, or equipment.
// Precedence: room > inventory > equipped.
func GetObjHere(ch *types.CharData, argument string) *types.ObjData {
	// Check room contents
	if ch.InRoom != nil {
		if obj := GetObjList(ch.InRoom.Contents, argument); obj != nil {
			return obj
		}
	}

	// Check inventory
	if obj := GetObjCarry(ch, argument); obj != nil {
		return obj
	}

	// Check equipment
	return GetObjWear(ch, argument)
}

// GetObjWorld finds an object anywhere in the game world.
// Checks current location first, then searches the full world object list.
func GetObjWorld(w *world.World, ch *types.CharData, argument string) *types.ObjData {
	// Check local first
	if obj := GetObjHere(ch, argument); obj != nil {
		return obj
	}

	number, arg := util.NumberArgument(argument)

	// Phase 1: exact match in world
	count := 0
	for _, obj := range w.Objects {
		if util.IsNameExact(arg, obj.Name) {
			count++
			if count == number {
				return obj
			}
		}
	}

	// Phase 2: prefix match in world
	count = 0
	for _, obj := range w.Objects {
		if util.IsName(arg, obj.Name) {
			count++
			if count == number {
				return obj
			}
		}
	}

	return nil
}

// GetObjList finds an object in a list by name.
// Two-phase search: exact match first, then prefix match.
func GetObjList(list []*types.ObjData, argument string) *types.ObjData {
	number, arg := util.NumberArgument(argument)

	// Phase 1: exact match
	count := 0
	for _, obj := range list {
		if util.IsNameExact(arg, obj.Name) {
			count++
			if count == number {
				return obj
			}
		}
	}

	// Phase 2: prefix match
	count = 0
	for _, obj := range list {
		if util.IsName(arg, obj.Name) {
			count++
			if count == number {
				return obj
			}
		}
	}

	return nil
}
