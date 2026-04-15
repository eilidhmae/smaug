package act

import (
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/mudprog"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// DoGet implements the 'get' command: pick up objects from room or containers.
func DoGet(ch *types.CharData, argument string) {
	arg1, arg2 := util.OneArgument(argument)
	if arg1 == "" {
		ch.Send("Get what?\n\r")
		return
	}

	if arg2 == "" {
		// Get from room
		obj := handler.GetObjHere(ch, arg1)
		if obj == nil {
			ch.Sendf("I see no %s here.\n\r", arg1)
			return
		}

		if obj.WearFlags&int(types.ITEM_TAKE) == 0 {
			ch.Send("You can't take that.\n\r")
			return
		}

		if obj.InRoom != nil {
			handler.ObjFromRoom(obj)
		} else if obj.InObj != nil {
			handler.ObjFromObj(obj)
		}
		handler.ObjToChar(obj, ch)
		ch.Sendf("You get %s.\n\r", obj.ShortDescr)
		mudprog.OprogGetTrigger(ch, obj)
		return
	}

	// Get from container
	container := handler.GetObjHere(ch, arg2)
	if container == nil {
		ch.Sendf("I see no %s here.\n\r", arg2)
		return
	}

	if container.ItemType != types.ITEM_CONTAINER &&
		container.ItemType != types.ITEM_CORPSE_NPC &&
		container.ItemType != types.ITEM_CORPSE_PC {
		ch.Send("That's not a container.\n\r")
		return
	}

	obj := handler.GetObjList(container.Contents, arg1)
	if obj == nil {
		ch.Sendf("I see nothing like that in %s.\n\r", container.ShortDescr)
		return
	}

	if obj.WearFlags&int(types.ITEM_TAKE) == 0 {
		ch.Send("You can't take that.\n\r")
		return
	}

	handler.ObjFromObj(obj)
	handler.ObjToChar(obj, ch)
	ch.Sendf("You get %s from %s.\n\r", obj.ShortDescr, container.ShortDescr)
	mudprog.OprogGetTrigger(ch, obj)
}

// DoDrop implements the 'drop' command: drop objects to room.
func DoDrop(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Drop what?\n\r")
		return
	}

	obj := handler.GetObjCarry(ch, arg)
	if obj == nil {
		ch.Send("You do not have that item.\n\r")
		return
	}

	if ch.InRoom != nil && ch.InRoom.RoomFlags.IsSet(types.ROOM_NODROP) {
		ch.Send("A magical force stops you.\n\r")
		return
	}

	if !handler.CanDropObj(obj) {
		ch.Send("You can't let go of it.\n\r")
		return
	}

	handler.ObjFromChar(obj)
	handler.ObjToRoom(obj, ch.InRoom)
	ch.Sendf("You drop %s.\n\r", obj.ShortDescr)
	mudprog.OprogDropTrigger(ch, obj)
}

// DoPut implements the 'put' command: put object in container.
func DoPut(ch *types.CharData, argument string) {
	arg1, arg2 := util.OneArgument(argument)
	if arg1 == "" || arg2 == "" {
		ch.Send("Put what in what?\n\r")
		return
	}

	obj := handler.GetObjCarry(ch, arg1)
	if obj == nil {
		ch.Send("You do not have that item.\n\r")
		return
	}

	container := handler.GetObjHere(ch, arg2)
	if container == nil {
		ch.Sendf("I see no %s here.\n\r", arg2)
		return
	}

	if container.ItemType != types.ITEM_CONTAINER &&
		container.ItemType != types.ITEM_KEYRING &&
		container.ItemType != types.ITEM_QUIVER {
		ch.Send("That's not a container.\n\r")
		return
	}

	if obj == container {
		ch.Send("You can't fold it into itself.\n\r")
		return
	}

	if !handler.CanDropObj(obj) {
		ch.Send("You can't let go of it.\n\r")
		return
	}

	handler.ObjFromChar(obj)
	handler.ObjToObj(obj, container)
	ch.Sendf("You put %s in %s.\n\r", obj.ShortDescr, container.ShortDescr)
}

// DoGive implements the 'give' command: give object to character.
func DoGive(ch *types.CharData, argument string) {
	arg1, arg2 := util.OneArgument(argument)
	if arg1 == "" || arg2 == "" {
		ch.Send("Give what to whom?\n\r")
		return
	}

	obj := handler.GetObjCarry(ch, arg1)
	if obj == nil {
		ch.Send("You do not have that item.\n\r")
		return
	}

	victim := handler.GetCharRoom(ch, arg2)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}

	if victim == ch {
		ch.Send("Give to yourself? That's pointless.\n\r")
		return
	}

	if !handler.CanDropObj(obj) {
		ch.Send("You can't let go of it.\n\r")
		return
	}

	handler.ObjFromChar(obj)
	handler.ObjToChar(obj, victim)
	ch.Sendf("You give %s to %s.\n\r", obj.ShortDescr, victim.Name)
}

// DoWear implements the 'wear' command: wear/wield/hold objects.
func DoWear(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Wear, wield, or hold what?\n\r")
		return
	}

	obj := handler.GetObjCarry(ch, arg)
	if obj == nil {
		ch.Send("You do not have that item.\n\r")
		return
	}

	if msg := itemWearRestriction(ch, obj); msg != "" {
		ch.Send(msg)
		return
	}

	wearLoc := findWearLoc(obj)
	if wearLoc == types.WEAR_NONE {
		ch.Send("You can't wear, wield, or hold that.\n\r")
		return
	}

	// Check if slot already occupied, auto-remove
	existing := handler.GetEqChar(ch, wearLoc)
	if existing != nil {
		if existing.ExtraFlags.IsSet(types.ITEM_NOREMOVE) {
			ch.Sendf("You can't remove %s.\n\r", existing.ShortDescr)
			return
		}
		handler.UnequipChar(ch, existing)
		ch.Sendf("You stop using %s.\n\r", existing.ShortDescr)
	}

	// Move from inventory to equipped
	handler.ObjFromChar(obj)
	handler.EquipChar(ch, obj, wearLoc)
	ch.Sendf("You %s %s.\n\r", wearVerb(wearLoc), obj.ShortDescr)
	mudprog.OprogWearTrigger(ch, obj)
}

// DoRemove implements the 'remove' command: remove worn equipment.
func DoRemove(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Remove what?\n\r")
		return
	}

	obj := handler.GetObjWear(ch, arg)
	if obj == nil {
		ch.Send("You are not using that item.\n\r")
		return
	}

	if obj.ExtraFlags.IsSet(types.ITEM_NOREMOVE) {
		ch.Sendf("You can't remove %s.\n\r", obj.ShortDescr)
		return
	}

	handler.UnequipChar(ch, obj)
	ch.Sendf("You stop using %s.\n\r", obj.ShortDescr)
	mudprog.OprogRemoveTrigger(ch, obj)
}

// DoSacrifice implements the 'sacrifice' command: destroy object for gold.
func DoSacrifice(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Sacrifice what?\n\r")
		return
	}

	obj := handler.GetObjHere(ch, arg)
	if obj == nil {
		ch.Send("You can't find it.\n\r")
		return
	}

	if obj.WearFlags&int(types.ITEM_TAKE) == 0 {
		ch.Sendf("%s is not an acceptable sacrifice.\n\r", util.Capitalize(obj.ShortDescr))
		return
	}

	// Must be in room, not in inventory
	if obj.InRoom == nil {
		ch.Send("You can't sacrifice something you're carrying.\n\r")
		return
	}

	mudprog.OprogSacTrigger(ch, obj)
	ch.Gold++
	handler.ExtractObj(WorldRef, obj)
	ch.Sendf("The gods give you one gold coin for your sacrifice of %s.\n\r", obj.ShortDescr)
}

// itemWearRestriction returns a refusal message if ch can't wear obj due to
// ITEM_ANTI_* flags; empty string means no restriction.
func itemWearRestriction(ch *types.CharData, obj *types.ObjData) string {
	// Immortals bypass anti-restrictions (holylight treated as trust).
	if !ch.IsNPC() && ch.Act.IsSet(types.PLR_HOLYLIGHT) {
		return ""
	}
	align := ch.Alignment
	if obj.ExtraFlags.IsSet(types.ITEM_ANTI_EVIL) && align <= -350 {
		return "You are too evil to use that.\n\r"
	}
	if obj.ExtraFlags.IsSet(types.ITEM_ANTI_GOOD) && align >= 350 {
		return "You are too good to use that.\n\r"
	}
	if obj.ExtraFlags.IsSet(types.ITEM_ANTI_NEUTRAL) && align > -350 && align < 350 {
		return "You are too neutral to use that.\n\r"
	}
	switch ch.Class {
	case types.CLASS_MAGE:
		if obj.ExtraFlags.IsSet(types.ITEM_ANTI_MAGE) {
			return "Mages can't use that.\n\r"
		}
	case types.CLASS_CLERIC:
		if obj.ExtraFlags.IsSet(types.ITEM_ANTI_CLERIC) {
			return "Clerics can't use that.\n\r"
		}
	case types.CLASS_THIEF:
		if obj.ExtraFlags.IsSet(types.ITEM_ANTI_THIEF) {
			return "Thieves can't use that.\n\r"
		}
	case types.CLASS_WARRIOR:
		if obj.ExtraFlags.IsSet(types.ITEM_ANTI_WARRIOR) {
			return "Warriors can't use that.\n\r"
		}
	}
	return ""
}

// findWearLoc determines the wear location for an object based on its wear flags.
func findWearLoc(obj *types.ObjData) int {
	flags := uint32(obj.WearFlags)

	switch {
	case obj.ItemType == types.ITEM_LIGHT:
		return types.WEAR_LIGHT
	case flags&types.ITEM_WEAR_FINGER != 0:
		return types.WEAR_FINGER_L
	case flags&types.ITEM_WEAR_NECK != 0:
		return types.WEAR_NECK_1
	case flags&types.ITEM_WEAR_BODY != 0:
		return types.WEAR_BODY
	case flags&types.ITEM_WEAR_HEAD != 0:
		return types.WEAR_HEAD
	case flags&types.ITEM_WEAR_LEGS != 0:
		return types.WEAR_LEGS
	case flags&types.ITEM_WEAR_FEET != 0:
		return types.WEAR_FEET
	case flags&types.ITEM_WEAR_HANDS != 0:
		return types.WEAR_HANDS
	case flags&types.ITEM_WEAR_ARMS != 0:
		return types.WEAR_ARMS
	case flags&types.ITEM_WEAR_SHIELD != 0:
		return types.WEAR_SHIELD
	case flags&types.ITEM_WEAR_ABOUT != 0:
		return types.WEAR_ABOUT
	case flags&types.ITEM_WEAR_WAIST != 0:
		return types.WEAR_WAIST
	case flags&types.ITEM_WEAR_WRIST != 0:
		return types.WEAR_WRIST_L
	case flags&types.ITEM_WIELD != 0:
		return types.WEAR_WIELD
	case flags&types.ITEM_HOLD != 0:
		return types.WEAR_HOLD
	case flags&types.ITEM_WEAR_EARS != 0:
		return types.WEAR_EARS
	case flags&types.ITEM_WEAR_EYES != 0:
		return types.WEAR_EYES
	case flags&types.ITEM_WEAR_BACK != 0:
		return types.WEAR_BACK
	case flags&types.ITEM_WEAR_FACE != 0:
		return types.WEAR_FACE
	case flags&types.ITEM_WEAR_ANKLE != 0:
		return types.WEAR_ANKLE_L
	}
	return types.WEAR_NONE
}

// wearVerb returns the appropriate verb for equipping at a wear location.
func wearVerb(wearLoc int) string {
	switch wearLoc {
	case types.WEAR_WIELD, types.WEAR_DUAL_WIELD, types.WEAR_MISSILE_WIELD:
		return "wield"
	case types.WEAR_HOLD:
		return "hold"
	case types.WEAR_LIGHT:
		return "light"
	default:
		return "wear"
	}
}
