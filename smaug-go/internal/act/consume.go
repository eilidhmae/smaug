package act

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// lookupSkillSlot finds a skill/spell slot number by name from the world skill table.
// Returns -1 if not found.
func lookupSkillSlot(name string) int {
	if WorldRef == nil {
		return -1
	}
	for i, sk := range WorldRef.Skills {
		if sk != nil && strings.EqualFold(sk.Name, name) {
			return i
		}
	}
	return -1
}

// GainCondition adjusts a character's condition value with bounds checking.
// Maps to C gain_condition() in update.c.
func GainCondition(ch *types.CharData, iCond int, value int) {
	if ch.IsNPC() || ch.PCData == nil {
		return
	}
	if ch.Level >= types.LEVEL_IMMORTAL {
		return
	}

	condition := ch.PCData.Condition[iCond]
	ch.PCData.Condition[iCond] = util.URANGE(0, condition+value, types.MAX_COND_VAL)

	if ch.PCData.Condition[iCond] == 0 {
		switch iCond {
		case types.COND_FULL:
			ch.Send("You are STARVING!\n\r")
		case types.COND_THIRST:
			ch.Send("You are DYING of THIRST!\n\r")
		case types.COND_DRUNK:
			// Sobers up — no warning
		}
	}
}

// DoEat implements the 'eat' command.
func DoEat(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Eat what?\n\r")
		return
	}

	obj := handler.GetObjCarry(ch, arg)
	if obj == nil {
		ch.Send("You do not have that item.\n\r")
		return
	}

	if !ch.IsImmortal() {
		if obj.ItemType != types.ITEM_FOOD && obj.ItemType != types.ITEM_PILL &&
			obj.ItemType != types.ITEM_COOK {
			ch.Send("That's not edible.\n\r")
			return
		}

		if !ch.IsNPC() && ch.PCData != nil && ch.PCData.Condition[types.COND_FULL] > 40 {
			ch.Send("You are too full to eat more.\n\r")
			return
		}
	}

	// Display eat message
	ch.Sendf("You eat %s.\n\r", obj.ShortDescr)
	if ch.InRoom != nil {
		for _, rch := range ch.InRoom.People {
			if rch != ch && rch.Desc != nil {
				rch.Sendf("%s eats %s.\n\r", ch.Name, obj.ShortDescr)
			}
		}
	}

	switch obj.ItemType {
	case types.ITEM_FOOD, types.ITEM_COOK:
		if !ch.IsNPC() && ch.PCData != nil {
			GainCondition(ch, types.COND_FULL, obj.Value[0])
			if ch.PCData.Condition[types.COND_FULL] > 40 {
				ch.Send("You are full.\n\r")
			}
		}
		// Check for poisoned food
		if obj.Value[3] != 0 {
			ch.Send("You feel very sick.\n\r")
			// Apply poison affect
			af := &types.AffectData{
				Type:     lookupSkillSlot("poison"),
				Duration: 2 * obj.Value[0],
				Location: types.APPLY_NONE,
				Modifier: 0,
			}
			af.BitVector.Set(types.AFF_POISON)
			handler.AffectJoin(ch, af)
		}

	case types.ITEM_PILL:
		// Pills are food that casts spells
		if !ch.IsNPC() && ch.PCData != nil {
			GainCondition(ch, types.COND_FULL, obj.Value[4])
		}
		// Spell casting from pills would be handled here when spell system is more complete
	}

	handler.ExtractObj(WorldRef, obj)
}

// DoDrink implements the 'drink' command.
func DoDrink(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)

	var obj *types.ObjData
	if arg == "" {
		// Default: find a fountain in the room
		if ch.InRoom != nil {
			for _, o := range ch.InRoom.Contents {
				if o.ItemType == types.ITEM_FOUNTAIN {
					obj = o
					break
				}
			}
		}
		if obj == nil {
			ch.Send("Drink what?\n\r")
			return
		}
	} else {
		obj = handler.GetObjHere(ch, arg)
		if obj == nil {
			ch.Send("You can't find that.\n\r")
			return
		}
	}

	if obj.ItemType != types.ITEM_DRINK_CON && obj.ItemType != types.ITEM_FOUNTAIN &&
		obj.ItemType != types.ITEM_BLOOD {
		ch.Send("You can't drink from that.\n\r")
		return
	}

	if obj.ItemType == types.ITEM_DRINK_CON || obj.ItemType == types.ITEM_BLOOD {
		if obj.Value[1] <= 0 {
			ch.Send("It is already empty.\n\r")
			return
		}
	}

	if !ch.IsNPC() && ch.PCData != nil && !ch.IsImmortal() {
		if ch.PCData.Condition[types.COND_THIRST] > 40 {
			ch.Send("Your stomach is too full to drink more.\n\r")
			return
		}
	}

	ch.Sendf("You drink from %s.\n\r", obj.ShortDescr)
	if ch.InRoom != nil {
		for _, rch := range ch.InRoom.People {
			if rch != ch && rch.Desc != nil {
				rch.Sendf("%s drinks from %s.\n\r", ch.Name, obj.ShortDescr)
			}
		}
	}

	if !ch.IsNPC() && ch.PCData != nil {
		switch obj.ItemType {
		case types.ITEM_FOUNTAIN:
			GainCondition(ch, types.COND_THIRST, types.MAX_COND_VAL)
		default:
			GainCondition(ch, types.COND_THIRST, 4)
			GainCondition(ch, types.COND_FULL, 1)
		}

		// Check for poison in drink
		if obj.Value[3] != 0 {
			ch.Send("You feel very sick.\n\r")
			af := &types.AffectData{
				Type:     lookupSkillSlot("poison"),
				Duration: obj.Value[3],
				Location: types.APPLY_NONE,
				Modifier: 0,
			}
			af.BitVector.Set(types.AFF_POISON)
			handler.AffectJoin(ch, af)
		}
	}

	// Decrement liquid amount for containers (not fountains)
	if obj.ItemType == types.ITEM_DRINK_CON || obj.ItemType == types.ITEM_BLOOD {
		obj.Value[1]--
		if obj.Value[1] <= 0 {
			obj.Value[1] = 0
		}
	}
}

// DoFill implements the 'fill' command: fill a container from a fountain.
func DoFill(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Fill what?\n\r")
		return
	}

	obj := handler.GetObjCarry(ch, arg)
	if obj == nil {
		ch.Send("You do not have that item.\n\r")
		return
	}

	if obj.ItemType != types.ITEM_DRINK_CON {
		ch.Send("You can't fill that.\n\r")
		return
	}

	if obj.Value[1] >= obj.Value[0] {
		ch.Send("It's already full.\n\r")
		return
	}

	// Find a fountain in the room
	var fountain *types.ObjData
	if ch.InRoom != nil {
		for _, o := range ch.InRoom.Contents {
			if o.ItemType == types.ITEM_FOUNTAIN {
				fountain = o
				break
			}
		}
	}
	if fountain == nil {
		ch.Send("There is no fountain here.\n\r")
		return
	}

	obj.Value[1] = obj.Value[0]   // fill to capacity
	obj.Value[2] = fountain.Value[2] // match liquid type
	obj.Value[3] = 0              // clear poison

	ch.Sendf("You fill %s from %s.\n\r", obj.ShortDescr, fountain.ShortDescr)
	if ch.InRoom != nil {
		for _, rch := range ch.InRoom.People {
			if rch != ch && rch.Desc != nil {
				rch.Sendf("%s fills %s from %s.\n\r", ch.Name, obj.ShortDescr, fountain.ShortDescr)
			}
		}
	}
}

// DoEmpty implements the 'empty' command: empty a drink container.
func DoEmpty(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Empty what?\n\r")
		return
	}

	obj := handler.GetObjCarry(ch, arg)
	if obj == nil {
		ch.Send("You do not have that item.\n\r")
		return
	}

	switch obj.ItemType {
	case types.ITEM_DRINK_CON:
		if obj.Value[1] <= 0 {
			ch.Send("It is already empty.\n\r")
			return
		}
		ch.Sendf("You empty %s.\n\r", obj.ShortDescr)
		if ch.InRoom != nil {
			for _, rch := range ch.InRoom.People {
				if rch != ch && rch.Desc != nil {
					rch.Sendf("%s empties %s.\n\r", ch.Name, obj.ShortDescr)
				}
			}
		}
		obj.Value[1] = 0

	default:
		ch.Send("You can't empty that.\n\r")
	}
}
