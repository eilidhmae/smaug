// Package handler provides entity manipulation functions: creating mob/object
// instances from templates, and placing entities in rooms/on characters.
// This is the Go equivalent of handler.c and parts of db.c from the C codebase.
package handler

import (
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// CreateMobile instantiates a mob from its index template.
// Equivalent to C create_mobile() in db.c.
func CreateMobile(w *world.World, idx *types.MobIndexData) *types.CharData {
	mob := &types.CharData{
		Name:              idx.PlayerName,
		ShortDescr:        idx.ShortDescr,
		LongDescr:         idx.LongDescr,
		Description:       idx.Description,
		Level:             util.NumberFuzzy(idx.Level),
		Sex:               idx.Sex,
		Race:              idx.Race,
		Class:             idx.Class,
		Alignment:         idx.Alignment,
		Act:               idx.Act,
		AffectedBy:        idx.AffectedBy,
		Position:          idx.Position,
		DefPosition:       idx.DefPosition,
		Height:            idx.Height,
		Weight:            idx.Weight,
		Gold:              idx.Gold,
		Silver:            idx.Silver,
		Copper:            idx.Copper,
		Exp:               idx.Exp,
		Hitroll:           idx.Hitroll,
		Damroll:           idx.Damroll,
		NumAttacks:        idx.NumAttacks,
		Speaks:            idx.Speaks,
		Speaking:          idx.Speaking,
		Immune:            idx.Immune,
		Resistant:         idx.Resistant,
		Susceptible:       idx.Susceptible,
		Attacks:           idx.Attacks,
		Defenses:          idx.Defenses,
		PermStr:           idx.PermStr,
		PermInt:           idx.PermInt,
		PermWis:           idx.PermWis,
		PermDex:           idx.PermDex,
		PermCon:           idx.PermCon,
		PermCha:           idx.PermCha,
		PermLck:           idx.PermLck,
		SavingPoisonDeath: idx.SavingPoisonDeath,
		SavingWand:        idx.SavingWand,
		SavingParaPetri:   idx.SavingParaPetri,
		SavingBreath:      idx.SavingBreath,
		SavingSpellStaff:  idx.SavingSpellStaff,
		IndexData:         idx,
	}

	// Ensure ACT_IS_NPC is set
	mob.Act.Set(types.ACT_IS_NPC)

	// Copy stances
	mob.Stances = idx.Stances

	// Compute armor class
	if idx.AC != 0 {
		mob.Armor = idx.AC
	} else {
		mob.Armor = interpolate(mob.Level, 100, -100)
	}

	// Compute hit points
	if idx.HitNoDice > 0 && idx.HitSizeDice > 0 {
		mob.MaxHit = idx.HitNoDice*util.DiceRoll(1, idx.HitSizeDice) + idx.HitPlus
	} else {
		mob.MaxHit = mob.Level*8 + util.NumberRange(mob.Level*mob.Level/4, mob.Level*mob.Level)
	}
	if mob.MaxHit < 1 {
		mob.MaxHit = 1
	}
	mob.Hit = mob.MaxHit

	// Default mana/move
	mob.MaxMana = 100
	mob.Mana = mob.MaxMana
	mob.MaxMove = 100
	mob.Move = mob.MaxMove

	// Bare-hand damage dice (from dam dice in index)
	mob.BareNumDie = idx.DamNoDice
	mob.BareSizeDie = idx.DamSizeDice

	// Track count on index
	idx.Count++

	// Add to world
	w.AddChar(mob)

	return mob
}

// CreateObject instantiates an object from its index template.
// Equivalent to C create_object() in db.c.
func CreateObject(w *world.World, idx *types.ObjIndexData, level int) *types.ObjData {
	obj := &types.ObjData{
		IndexData:   idx,
		Name:        idx.Name,
		ShortDescr:  idx.ShortDescr,
		Description: idx.Description,
		ActionDesc:  idx.ActionDesc,
		ItemType:    idx.ItemType,
		ExtraFlags:  idx.ExtraFlags,
		WearFlags:   idx.WearFlags,
		Weight:      idx.Weight,
		GoldCost:    idx.GoldCost,
		SilverCost:  idx.SilverCost,
		CopperCost:  idx.CopperCost,
		Level:       level,
		WearLoc:     types.WEAR_NONE,
		Count:       1,
		Value:       idx.Value,
	}

	// Copy extra descriptions
	for _, ed := range idx.ExtraDescr {
		obj.ExtraDescr = append(obj.ExtraDescr, &types.ExtraDescrData{
			Keyword:     ed.Keyword,
			Description: ed.Description,
		})
	}

	// Copy affects
	for _, aff := range idx.Affects {
		obj.Affects = append(obj.Affects, &types.AffectData{
			Type:      aff.Type,
			Duration:  aff.Duration,
			Location:  aff.Location,
			Modifier:  aff.Modifier,
			BitVector: aff.BitVector,
		})
	}

	// Track count on index
	idx.Count++

	// Add to world
	w.AddObj(obj)

	return obj
}

// CharToRoom places a character in a room.
func CharToRoom(ch *types.CharData, room *types.RoomIndexData) {
	if room == nil {
		util.Bug("CharToRoom: nil room for %s", ch.Name)
		return
	}
	ch.InRoom = room
	room.People = append(room.People, ch)

	// If the room is dark, give infrared to NPCs
	if ch.IsNPC() && room.RoomFlags.IsSet(types.ROOM_DARK) {
		ch.AffectedBy.Set(types.AFF_INFRARED)
	}
}

// CharFromRoom removes a character from their current room.
func CharFromRoom(ch *types.CharData) {
	if ch.InRoom == nil {
		return
	}
	room := ch.InRoom
	for i, p := range room.People {
		if p == ch {
			room.People = append(room.People[:i], room.People[i+1:]...)
			break
		}
	}
	ch.WasInRoom = room
	ch.InRoom = nil
}

// ObjToRoom places an object in a room.
func ObjToRoom(obj *types.ObjData, room *types.RoomIndexData) {
	if room == nil {
		util.Bug("ObjToRoom: nil room for %s", obj.Name)
		return
	}
	obj.InRoom = room
	room.Contents = append(room.Contents, obj)
}

// ObjToChar gives an object to a character (inventory).
func ObjToChar(obj *types.ObjData, ch *types.CharData) {
	obj.CarriedBy = ch
	ch.Carrying = append(ch.Carrying, obj)
}

// ObjToObj puts an object inside another object (container).
func ObjToObj(obj *types.ObjData, container *types.ObjData) {
	obj.InObj = container
	container.Contents = append(container.Contents, obj)
}

// EquipChar equips an object in a specific wear location on a character.
func EquipChar(ch *types.CharData, obj *types.ObjData, wearLoc int) {
	obj.WearLoc = wearLoc
	obj.CarriedBy = ch
	ch.Carrying = append(ch.Carrying, obj)
}

// interpolate does linear interpolation between low and high over 0..LEVEL_AVATAR range.
func interpolate(level, low, high int) int {
	if types.LEVEL_AVATAR <= 0 {
		return low
	}
	return low + (high-low)*level/types.LEVEL_AVATAR
}
