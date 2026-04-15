package act

import (
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/magic"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// objItemCastSpell looks up and invokes the spell function for a given skill slot.
// Callers are expected to have already checked roomSuppressesMagic before
// consuming the item — otherwise a potion in a no-magic room would be
// destroyed without effect.
func objItemCastSpell(sn int, level int, ch *types.CharData, victim *types.CharData) {
	if WorldRef == nil || sn <= 0 || sn >= len(WorldRef.Skills) {
		return
	}
	sk := WorldRef.Skills[sn]
	if sk == nil {
		return
	}
	spellFn := magic.FindSpellFunc(sk.SpellFunName)
	if spellFn == nil {
		return
	}
	spellFn(WorldRef, sn, level, ch, victim)
}

// noMagicSuppresses reports whether the caller should abort before consuming
// an item or charge because ch's room (or its area) suppresses magic. Matches
// C `src/magic.c:obj_cast_spell` gate.
func noMagicSuppresses(ch *types.CharData) bool {
	if ch.InRoom != nil && roomSuppressesMagic(ch.InRoom) {
		ch.Send("Nothing seems to happen.\n\r")
		return true
	}
	return false
}

// findHeldItemType finds an equipped item of the given type at WEAR_HOLD.
func findHeldItemType(ch *types.CharData, itemType int) *types.ObjData {
	for _, obj := range ch.Carrying {
		if obj.WearLoc == types.WEAR_HOLD && obj.ItemType == itemType {
			return obj
		}
	}
	return nil
}

// DoQuaff implements the 'quaff' command — drink a potion to cast its spells on self.
func DoQuaff(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Quaff what?\n\r")
		return
	}

	obj := handler.GetObjCarry(ch, arg)
	if obj == nil {
		ch.Send("You do not have that item.\n\r")
		return
	}

	if obj.ItemType != types.ITEM_POTION {
		ch.Send("You can only quaff potions.\n\r")
		return
	}

	if noMagicSuppresses(ch) {
		return
	}

	ch.Sendf("You quaff %s.\n\r", obj.ShortDescr)
	if ch.InRoom != nil {
		for _, rch := range ch.InRoom.People {
			if rch != ch && rch.Desc != nil {
				rch.Sendf("%s quaffs %s.\n\r", ch.Name, obj.ShortDescr)
			}
		}
	}

	objItemCastSpell(obj.Value[1], obj.Value[0], ch, ch)
	objItemCastSpell(obj.Value[2], obj.Value[0], ch, ch)
	objItemCastSpell(obj.Value[3], obj.Value[0], ch, ch)

	handler.ExtractObj(WorldRef, obj)
}

// DoRecite implements the 'recite' command — read a scroll to cast its spells on a target.
func DoRecite(ch *types.CharData, argument string) {
	arg1, rest := util.OneArgument(argument)
	if arg1 == "" {
		ch.Send("Recite what?\n\r")
		return
	}

	obj := handler.GetObjCarry(ch, arg1)
	if obj == nil {
		ch.Send("You do not have that item.\n\r")
		return
	}

	if obj.ItemType != types.ITEM_SCROLL {
		ch.Send("You can only recite scrolls.\n\r")
		return
	}

	if noMagicSuppresses(ch) {
		return
	}

	arg2, _ := util.OneArgument(rest)
	var victim *types.CharData
	if arg2 == "" {
		victim = ch
	} else {
		victim = handler.GetCharRoom(ch, arg2)
		if victim == nil {
			ch.Send("They aren't here.\n\r")
			return
		}
	}

	ch.Sendf("You recite %s.\n\r", obj.ShortDescr)
	if ch.InRoom != nil {
		for _, rch := range ch.InRoom.People {
			if rch != ch && rch.Desc != nil {
				rch.Sendf("%s recites %s.\n\r", ch.Name, obj.ShortDescr)
			}
		}
	}

	objItemCastSpell(obj.Value[1], obj.Value[0], ch, victim)
	objItemCastSpell(obj.Value[2], obj.Value[0], ch, victim)
	objItemCastSpell(obj.Value[3], obj.Value[0], ch, victim)

	handler.ExtractObj(WorldRef, obj)
}

// DoBrandish implements the 'brandish' command — wave a staff to cast its spell on all in room.
func DoBrandish(ch *types.CharData, argument string) {
	staff := findHeldItemType(ch, types.ITEM_STAFF)
	if staff == nil {
		ch.Send("You are not holding a staff.\n\r")
		return
	}

	if staff.Value[2] <= 0 {
		ch.Send("The staff has no charges remaining.\n\r")
		return
	}

	if noMagicSuppresses(ch) {
		return
	}

	staff.Value[2]--

	ch.Sendf("You brandish %s.\n\r", staff.ShortDescr)
	if ch.InRoom != nil {
		for _, rch := range ch.InRoom.People {
			if rch != ch && rch.Desc != nil {
				rch.Sendf("%s brandishes %s.\n\r", ch.Name, staff.ShortDescr)
			}
		}
	}

	// Cast spell on everyone in the room
	if ch.InRoom != nil {
		for _, rch := range ch.InRoom.People {
			objItemCastSpell(staff.Value[3], staff.Value[0], ch, rch)
		}
	}

	if staff.Value[2] <= 0 {
		ch.Send("The staff crumbles to dust.\n\r")
		handler.ExtractObj(WorldRef, staff)
	}
}

// DoZap implements the 'zap' command — use a wand to cast its spell on a target.
func DoZap(ch *types.CharData, argument string) {
	wand := findHeldItemType(ch, types.ITEM_WAND)
	if wand == nil {
		ch.Send("You are not holding a wand.\n\r")
		return
	}

	if wand.Value[2] <= 0 {
		ch.Send("The wand has no charges remaining.\n\r")
		return
	}

	if noMagicSuppresses(ch) {
		return
	}

	arg, _ := util.OneArgument(argument)
	var victim *types.CharData
	if arg != "" {
		victim = handler.GetCharRoom(ch, arg)
		if victim == nil {
			ch.Send("They aren't here.\n\r")
			return
		}
	} else if ch.Fighting != nil {
		victim = ch.Fighting.Who
	} else {
		victim = ch
	}

	wand.Value[2]--

	ch.Sendf("You zap %s with %s.\n\r", victim.Name, wand.ShortDescr)
	if ch.InRoom != nil {
		for _, rch := range ch.InRoom.People {
			if rch != ch && rch.Desc != nil {
				rch.Sendf("%s zaps %s with %s.\n\r", ch.Name, victim.Name, wand.ShortDescr)
			}
		}
	}

	objItemCastSpell(wand.Value[3], wand.Value[0], ch, victim)

	if wand.Value[2] <= 0 {
		ch.Send("The wand crumbles to dust.\n\r")
		handler.ExtractObj(WorldRef, wand)
	}
}
