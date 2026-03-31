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

// ObjFromChar removes an object from a character's carrying list.
// If the object is equipped, it is unequipped first.
func ObjFromChar(obj *types.ObjData) {
	ch := obj.CarriedBy
	if ch == nil {
		util.Bug("ObjFromChar: nil CarriedBy for %s", obj.Name)
		return
	}

	// If equipped, unequip first
	if obj.WearLoc != types.WEAR_NONE {
		UnequipChar(ch, obj)
	}

	// Remove from carrying list
	for i, o := range ch.Carrying {
		if o == obj {
			ch.Carrying = append(ch.Carrying[:i], ch.Carrying[i+1:]...)
			break
		}
	}

	obj.CarriedBy = nil
	obj.InRoom = nil
}

// ObjFromRoom removes an object from a room's contents.
func ObjFromRoom(obj *types.ObjData) {
	room := obj.InRoom
	if room == nil {
		util.Bug("ObjFromRoom: nil InRoom for %s", obj.Name)
		return
	}

	for i, o := range room.Contents {
		if o == obj {
			room.Contents = append(room.Contents[:i], room.Contents[i+1:]...)
			break
		}
	}

	obj.CarriedBy = nil
	obj.InObj = nil
	obj.InRoom = nil
}

// ObjFromObj removes an object from inside a container object.
func ObjFromObj(obj *types.ObjData) {
	container := obj.InObj
	if container == nil {
		util.Bug("ObjFromObj: nil InObj for %s", obj.Name)
		return
	}

	for i, o := range container.Contents {
		if o == obj {
			container.Contents = append(container.Contents[:i], container.Contents[i+1:]...)
			break
		}
	}

	obj.InObj = nil
	obj.InRoom = nil
	obj.CarriedBy = nil
}

// UnequipChar removes equipment from a character's wear location.
// The object remains in the character's carrying list (moved to inventory).
func UnequipChar(ch *types.CharData, obj *types.ObjData) {
	if obj.WearLoc == types.WEAR_NONE {
		util.Bug("UnequipChar: already unequipped: %s", obj.Name)
		return
	}

	// If removing primary wield and dual wield exists, move dual to wield
	if obj.WearLoc == types.WEAR_WIELD {
		for _, o := range ch.Carrying {
			if o.WearLoc == types.WEAR_DUAL_WIELD {
				o.WearLoc = types.WEAR_WIELD
				break
			}
		}
	}

	obj.WearLoc = types.WEAR_NONE
}

// ExtractObj fully removes an object from the game world.
// Recursively extracts all contents.
func ExtractObj(w *world.World, obj *types.ObjData) {
	// Remove from wherever it is
	if obj.CarriedBy != nil {
		ObjFromChar(obj)
	} else if obj.InRoom != nil {
		ObjFromRoom(obj)
	} else if obj.InObj != nil {
		ObjFromObj(obj)
	}

	// Recursively extract contents
	for len(obj.Contents) > 0 {
		ExtractObj(w, obj.Contents[len(obj.Contents)-1])
	}

	// Remove from world object list
	w.RemoveObj(obj)

	// Update template count
	if obj.IndexData != nil {
		obj.IndexData.Count--
	}
}

// ExtractChar fully removes a character from the game world.
// If fPull is true, the character is permanently removed (not rebirth).
func ExtractChar(w *world.World, ch *types.CharData, fPull bool) {
	if ch.InRoom == nil {
		util.Bug("ExtractChar: nil InRoom for %s", ch.Name)
	}

	// Stop combat
	if ch.Fighting != nil {
		ch.Fighting = nil
		ch.NumFighting = 0
	}

	// Clear mount references
	if ch.Mount != nil {
		ch.Mount = nil
	}

	// Extract all carried objects
	for len(ch.Carrying) > 0 {
		ExtractObj(w, ch.Carrying[len(ch.Carrying)-1])
	}

	// Remove from room
	if ch.InRoom != nil {
		CharFromRoom(ch)
	}

	if !fPull {
		// Rebirth path — not implemented yet (Phase 2 combat death)
		return
	}

	// Update NPC template count
	if ch.IsNPC() && ch.IndexData != nil {
		ch.IndexData.Count--
	}

	// Clear reply/retell references in other characters
	for _, wch := range w.Characters {
		if wch.Reply == ch {
			wch.Reply = nil
		}
		if wch.Retell == ch {
			wch.Retell = nil
		}
	}

	// Remove from world character list
	w.RemoveChar(ch)

	// Disconnect if has descriptor
	if ch.Desc != nil {
		ch.Desc.Character = nil
		ch.Desc = nil
	}
}

// AffectModify applies or removes an affect's stat modifications on a character.
// If fAdd is true, the modifier is added; if false, it is subtracted.
func AffectModify(ch *types.CharData, aff *types.AffectData, fAdd bool) {
	mod := aff.Modifier
	if !fAdd {
		mod = -mod
	}

	switch aff.Location {
	case types.APPLY_NONE:
		// no stat change
	case types.APPLY_STR:
		ch.ModStr += mod
	case types.APPLY_DEX:
		ch.ModDex += mod
	case types.APPLY_INT:
		ch.ModInt += mod
	case types.APPLY_WIS:
		ch.ModWis += mod
	case types.APPLY_CON:
		ch.ModCon += mod
	case types.APPLY_CHA:
		ch.ModCha += mod
	case types.APPLY_LCK:
		ch.ModLck += mod
	case types.APPLY_SEX:
		ch.Sex += mod
	case types.APPLY_LEVEL:
		// not applied to mod directly in SMAUG
	case types.APPLY_AGE:
		// not applied to mod directly
	case types.APPLY_HEIGHT:
		ch.Height += mod
	case types.APPLY_WEIGHT:
		ch.Weight += mod
	case types.APPLY_MANA:
		ch.MaxMana += mod
	case types.APPLY_HIT:
		ch.MaxHit += mod
	case types.APPLY_MOVE:
		ch.MaxMove += mod
	case types.APPLY_GOLD:
		// not applied to mod
	case types.APPLY_EXP:
		// not applied to mod
	case types.APPLY_AC:
		ch.Armor += mod
	case types.APPLY_HITROLL:
		ch.Hitroll += mod
	case types.APPLY_DAMROLL:
		ch.Damroll += mod
	case types.APPLY_SAVING_POISON:
		ch.SavingPoisonDeath += mod
	case types.APPLY_SAVING_ROD:
		ch.SavingWand += mod
	case types.APPLY_SAVING_PARA:
		ch.SavingParaPetri += mod
	case types.APPLY_SAVING_BREATH:
		ch.SavingBreath += mod
	case types.APPLY_SAVING_SPELL:
		ch.SavingSpellStaff += mod
	}

	// Apply bitvector flags
	if !aff.BitVector.IsEmpty() {
		if fAdd {
			ch.AffectedBy = ch.AffectedBy.Or(aff.BitVector)
		} else {
			ch.AffectedBy = ch.AffectedBy.AndNot(aff.BitVector)
		}
	}
}

// AffectToChar adds a new affect to a character, applying its stat modifications.
func AffectToChar(ch *types.CharData, aff *types.AffectData) {
	newAff := &types.AffectData{
		Type:      aff.Type,
		Duration:  aff.Duration,
		Location:  aff.Location,
		Modifier:  aff.Modifier,
		BitVector: aff.BitVector,
	}
	ch.Affects = append(ch.Affects, newAff)
	AffectModify(ch, newAff, true)
}

// AffectRemove removes a specific affect from a character, reversing its stat modifications.
func AffectRemove(ch *types.CharData, aff *types.AffectData) {
	AffectModify(ch, aff, false)

	for i, a := range ch.Affects {
		if a == aff {
			ch.Affects = append(ch.Affects[:i], ch.Affects[i+1:]...)
			return
		}
	}
	util.Bug("AffectRemove: affect not found on %s", ch.Name)
}

// AffectStrip removes all affects of a given skill/spell type from a character.
func AffectStrip(ch *types.CharData, sn int) {
	for i := len(ch.Affects) - 1; i >= 0; i-- {
		if ch.Affects[i].Type == sn {
			AffectRemove(ch, ch.Affects[i])
		}
	}
}

// AffectJoin combines a new affect with an existing one of the same type,
// or adds it if no matching affect exists.
func AffectJoin(ch *types.CharData, aff *types.AffectData) {
	for _, old := range ch.Affects {
		if old.Type == aff.Type {
			// Combine duration and modifier
			aff.Duration = util.UMIN(1000000, aff.Duration+old.Duration)
			if aff.Modifier != 0 {
				aff.Modifier = util.UMIN(5000, aff.Modifier+old.Modifier)
			} else {
				aff.Modifier = old.Modifier
			}
			AffectRemove(ch, old)
			break
		}
	}
	AffectToChar(ch, aff)
}

// GetEqChar returns the object equipped at a specific wear location, or nil.
func GetEqChar(ch *types.CharData, wearLoc int) *types.ObjData {
	for _, obj := range ch.Carrying {
		if obj.WearLoc == wearLoc {
			return obj
		}
	}
	return nil
}

// CanDropObj returns true if an object can be dropped/given away.
// Objects with ITEM_NODROP flag cannot be dropped.
func CanDropObj(obj *types.ObjData) bool {
	return !obj.ExtraFlags.IsSet(types.ITEM_NODROP)
}

// interpolate does linear interpolation between low and high over 0..LEVEL_AVATAR range.
func interpolate(level, low, high int) int {
	if types.LEVEL_AVATAR <= 0 {
		return low
	}
	return low + (high-low)*level/types.LEVEL_AVATAR
}
