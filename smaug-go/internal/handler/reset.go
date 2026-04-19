package handler

import (
	"log"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// ResetArea processes all resets for an area, instantiating mobs and objects.
// Equivalent to C reset_area() in reset.c.
func ResetArea(w *world.World, area *types.AreaData) {
	var lastMob *types.CharData
	var lastObj *types.ObjData
	var lastRoom *types.RoomIndexData
	mobLevel := 0

	for _, reset := range area.Resets {
		switch reset.Command {
		case 'M':
			resetMobile(w, reset, &lastMob, &lastRoom, &mobLevel)
		case 'G':
			resetGive(w, reset, lastMob, mobLevel)
		case 'E':
			resetEquip(w, reset, lastMob, mobLevel)
		case 'O':
			resetObject(w, reset, &lastObj, &lastRoom)
		case 'P':
			resetPut(w, reset, lastObj)
		case 'D':
			resetDoor(w, reset)
		case 'H':
			resetHide(w, reset, lastObj, lastRoom)
		case 'R':
			// Randomize exits — skip for now
		case 'T':
			// Traps — skip for now
		case 'B':
			// Bit manipulation — skip for now
		}
	}
}

// resetMobile handles 'M' reset: spawn a mob in a room.
func resetMobile(w *world.World, reset *types.ResetData, lastMob **types.CharData, lastRoom **types.RoomIndexData, mobLevel *int) {
	*lastMob = nil

	idx := w.GetMobIndex(reset.Arg1)
	if idx == nil {
		util.Bug("resetMobile: mob vnum %d not found", reset.Arg1)
		return
	}

	room := w.GetRoom(reset.Arg3)
	if room == nil {
		util.Bug("resetMobile: room vnum %d not found", reset.Arg3)
		return
	}

	// Check max count: arg2 is the maximum number allowed
	if idx.Count >= reset.Arg2 {
		return
	}

	mob := CreateMobile(w, idx)
	CharToRoom(mob, room)
	SetSentinelHome(mob, room)

	*lastMob = mob
	*lastRoom = room
	*mobLevel = util.URANGE(0, mob.Level-2, types.LEVEL_AVATAR)
}

// resetGive handles 'G' reset: give an object to the last spawned mob.
func resetGive(w *world.World, reset *types.ResetData, lastMob *types.CharData, mobLevel int) {
	if lastMob == nil {
		return
	}

	idx := w.GetObjIndex(reset.Arg1)
	if idx == nil {
		util.Bug("resetGive: obj vnum %d not found", reset.Arg1)
		return
	}

	level := util.NumberFuzzy(mobLevel)
	if level > types.LEVEL_AVATAR {
		level = types.LEVEL_AVATAR
	}

	// Shop mobs get items at a higher level with ITEM_INVENTORY
	if lastMob.IndexData != nil && lastMob.IndexData.Shop != nil {
		level = util.UMAX(level, lastMob.Level)
	}

	obj := CreateObject(w, idx, level)

	if lastMob.IndexData != nil && lastMob.IndexData.Shop != nil {
		obj.ExtraFlags.Set(types.ITEM_INVENTORY)
	}

	ObjToChar(obj, lastMob)
}

// resetEquip handles 'E' reset: equip an object on the last spawned mob.
func resetEquip(w *world.World, reset *types.ResetData, lastMob *types.CharData, mobLevel int) {
	if lastMob == nil {
		return
	}

	idx := w.GetObjIndex(reset.Arg1)
	if idx == nil {
		util.Bug("resetEquip: obj vnum %d not found", reset.Arg1)
		return
	}

	level := util.NumberFuzzy(mobLevel)
	if level > types.LEVEL_AVATAR {
		level = types.LEVEL_AVATAR
	}

	obj := CreateObject(w, idx, level)
	EquipChar(lastMob, obj, reset.Arg3)
}

// resetObject handles 'O' reset: place an object in a room.
func resetObject(w *world.World, reset *types.ResetData, lastObj **types.ObjData, lastRoom **types.RoomIndexData) {
	*lastObj = nil

	idx := w.GetObjIndex(reset.Arg1)
	if idx == nil {
		util.Bug("resetObject: obj vnum %d not found", reset.Arg1)
		return
	}

	room := w.GetRoom(reset.Arg3)
	if room == nil {
		util.Bug("resetObject: room vnum %d not found", reset.Arg3)
		return
	}

	// Skip if object already in room (check by vnum)
	for _, existing := range room.Contents {
		if existing.IndexData != nil && existing.IndexData.Vnum == idx.Vnum {
			*lastObj = existing
			*lastRoom = room
			return
		}
	}

	level := util.NumberFuzzy(idx.Level)
	if level > types.LEVEL_AVATAR {
		level = types.LEVEL_AVATAR
	}

	obj := CreateObject(w, idx, level)
	// Room objects have no sale value
	obj.GoldCost = 0
	obj.SilverCost = 0
	obj.CopperCost = 0

	ObjToRoom(obj, room)
	*lastObj = obj
	*lastRoom = room
}

// resetPut handles 'P' reset: put an object inside a container.
func resetPut(w *world.World, reset *types.ResetData, lastObj *types.ObjData) {
	idx := w.GetObjIndex(reset.Arg1)
	if idx == nil {
		util.Bug("resetPut: obj vnum %d not found", reset.Arg1)
		return
	}

	// Find the target container
	var container *types.ObjData
	if reset.Arg3 > 0 {
		// Find by vnum in world
		targetIdx := w.GetObjIndex(reset.Arg3)
		if targetIdx == nil {
			util.Bug("resetPut: container vnum %d not found", reset.Arg3)
			return
		}
		container = findObjByIndex(w, targetIdx)
	} else {
		container = lastObj
	}

	if container == nil {
		return
	}

	// Skip if object already in container
	for _, existing := range container.Contents {
		if existing.IndexData != nil && existing.IndexData.Vnum == idx.Vnum {
			return
		}
	}

	level := util.NumberFuzzy(util.UMAX(idx.Level, container.Level))
	if level > types.LEVEL_AVATAR {
		level = types.LEVEL_AVATAR
	}

	obj := CreateObject(w, idx, level)
	ObjToObj(obj, container)
}

// resetDoor handles 'D' reset: set door state.
func resetDoor(w *world.World, reset *types.ResetData) {
	room := w.GetRoom(reset.Arg1)
	if room == nil {
		return
	}

	exit := room.GetExit(reset.Arg2)
	if exit == nil {
		return
	}

	switch reset.Arg3 {
	case 0: // Open
		exit.ExitInfo &^= int(types.EX_CLOSED)
		exit.ExitInfo &^= int(types.EX_LOCKED)
	case 1: // Closed
		exit.ExitInfo |= int(types.EX_CLOSED)
		exit.ExitInfo &^= int(types.EX_LOCKED)
	case 2: // Locked
		exit.ExitInfo |= int(types.EX_CLOSED)
		exit.ExitInfo |= int(types.EX_LOCKED)
	}
}

// resetHide handles 'H' reset: hide an object.
func resetHide(w *world.World, reset *types.ResetData, lastObj *types.ObjData, lastRoom *types.RoomIndexData) {
	var obj *types.ObjData
	if reset.Arg1 > 0 {
		idx := w.GetObjIndex(reset.Arg1)
		if idx != nil {
			obj = findObjByIndex(w, idx)
		}
	} else {
		obj = lastObj
	}

	if obj != nil {
		obj.ExtraFlags.Set(types.ITEM_HIDDEN)
	}
}

// findObjByIndex finds the first object instance in the world matching an index.
func findObjByIndex(w *world.World, idx *types.ObjIndexData) *types.ObjData {
	for _, obj := range w.Objects {
		if obj.IndexData == idx {
			return obj
		}
	}
	return nil
}

// ResetAllAreas processes resets for all loaded areas.
func ResetAllAreas(w *world.World) {
	mobCount := 0
	objCount := 0
	beforeMobs := len(w.Characters)
	beforeObjs := len(w.Objects)

	for _, area := range w.Areas {
		ResetArea(w, area)
	}

	mobCount = len(w.Characters) - beforeMobs
	objCount = len(w.Objects) - beforeObjs
	log.Printf("Area resets: %d mobs and %d objects instantiated.", mobCount, objCount)
}
