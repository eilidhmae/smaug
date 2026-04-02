package act

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// setupItemUseWorld creates a world with skills for item-use tests.
func setupItemUseWorld() *world.World {
	w := world.New("../../db")
	w.Skills = make([]*types.SkillType, 20)
	w.Skills[5] = &types.SkillType{
		Name:         "cure light",
		Type:         types.SKILL_SPELL,
		SpellFunName: "spell_cure_light",
	}
	w.Skills[8] = &types.SkillType{
		Name:         "armor",
		Type:         types.SKILL_SPELL,
		SpellFunName: "spell_armor",
	}
	WorldRef = w
	return w
}

// makeEquippedObj creates an object equipped at the given wear location.
func makeEquippedObj(ch *types.CharData, name, shortDescr string, itemType int, values [6]int, wearLoc int) *types.ObjData {
	obj := &types.ObjData{
		Name:       name,
		ShortDescr: shortDescr,
		ItemType:   itemType,
		Value:      values,
		CarriedBy:  ch,
		WearLoc:    wearLoc,
	}
	ch.Carrying = append(ch.Carrying, obj)
	return obj
}

// --- DoQuaff tests ---

func TestDoQuaff_NoArg(t *testing.T) {
	setupItemUseWorld()
	ch := newTestCharWithDesc()
	DoQuaff(ch, "")
	// Should send "Quaff what?"
}

func TestDoQuaff_ItemNotFound(t *testing.T) {
	setupItemUseWorld()
	ch := newTestCharWithDesc()
	DoQuaff(ch, "potion")
	// Should send "You do not have that item."
}

func TestDoQuaff_WrongItemType(t *testing.T) {
	setupItemUseWorld()
	ch := newTestCharWithDesc()
	makeCarriedObj(ch, "sword", "a sword", types.ITEM_WEAPON, [6]int{})
	DoQuaff(ch, "sword")
	if len(ch.Carrying) != 1 {
		t.Error("non-potion should not be consumed")
	}
}

func TestDoQuaff_Success(t *testing.T) {
	setupItemUseWorld()
	ch := newTestCharWithDesc()
	// Potion: level 10, spell slot 5 (cure light), no other spells
	makeCarriedObj(ch, "potion healing", "a healing potion", types.ITEM_POTION, [6]int{10, 5, 0, 0, 0, 0})

	DoQuaff(ch, "potion")
	if len(ch.Carrying) != 0 {
		t.Error("potion should be consumed after quaffing")
	}
}

func TestDoQuaff_MultipleSpells(t *testing.T) {
	setupItemUseWorld()
	ch := newTestCharWithDesc()
	// Potion with spell slots 5 and 8
	makeCarriedObj(ch, "potion multi", "a multi potion", types.ITEM_POTION, [6]int{10, 5, 8, 0, 0, 0})

	DoQuaff(ch, "potion")
	if len(ch.Carrying) != 0 {
		t.Error("potion should be consumed after quaffing")
	}
}

func TestDoQuaff_InvalidSpellSlot(t *testing.T) {
	setupItemUseWorld()
	ch := newTestCharWithDesc()
	// Potion with out-of-range spell slot
	makeCarriedObj(ch, "potion broken", "a broken potion", types.ITEM_POTION, [6]int{10, 99, 0, 0, 0, 0})

	DoQuaff(ch, "potion")
	// Should not panic, potion consumed
	if len(ch.Carrying) != 0 {
		t.Error("potion should be consumed even with invalid spell slot")
	}
}

// --- DoRecite tests ---

func TestDoRecite_NoArg(t *testing.T) {
	setupItemUseWorld()
	ch := newTestCharWithDesc()
	DoRecite(ch, "")
	// Should send "Recite what?"
}

func TestDoRecite_ItemNotFound(t *testing.T) {
	setupItemUseWorld()
	ch := newTestCharWithDesc()
	DoRecite(ch, "scroll")
	// Should send "You do not have that item."
}

func TestDoRecite_WrongItemType(t *testing.T) {
	setupItemUseWorld()
	ch := newTestCharWithDesc()
	makeCarriedObj(ch, "sword", "a sword", types.ITEM_WEAPON, [6]int{})
	DoRecite(ch, "sword")
	if len(ch.Carrying) != 1 {
		t.Error("non-scroll should not be consumed")
	}
}

func TestDoRecite_SuccessSelf(t *testing.T) {
	setupItemUseWorld()
	ch := newTestCharWithDesc()
	makeCarriedObj(ch, "scroll recall", "a scroll of recall", types.ITEM_SCROLL, [6]int{10, 5, 0, 0, 0, 0})

	DoRecite(ch, "scroll")
	if len(ch.Carrying) != 0 {
		t.Error("scroll should be consumed after reciting")
	}
}

func TestDoRecite_SuccessTarget(t *testing.T) {
	setupItemUseWorld()
	ch := newTestCharWithDesc()
	victim := newTestCharWithDesc()
	victim.Name = "Victim"
	victim.InRoom = ch.InRoom
	ch.InRoom.People = append(ch.InRoom.People, ch, victim)

	makeCarriedObj(ch, "scroll cure", "a scroll of cure light", types.ITEM_SCROLL, [6]int{10, 5, 0, 0, 0, 0})

	DoRecite(ch, "scroll Victim")
	if len(ch.Carrying) != 0 {
		t.Error("scroll should be consumed after reciting on target")
	}
}

func TestDoRecite_TargetNotFound(t *testing.T) {
	setupItemUseWorld()
	ch := newTestCharWithDesc()
	ch.InRoom.People = append(ch.InRoom.People, ch)
	makeCarriedObj(ch, "scroll cure", "a scroll of cure", types.ITEM_SCROLL, [6]int{10, 5, 0, 0, 0, 0})

	DoRecite(ch, "scroll nobody")
	// Should send "They aren't here." and not consume scroll
	if len(ch.Carrying) != 1 {
		t.Error("scroll should not be consumed when target not found")
	}
}

// --- DoBrandish tests ---

func TestDoBrandish_NoStaff(t *testing.T) {
	setupItemUseWorld()
	ch := newTestCharWithDesc()
	DoBrandish(ch, "")
	// Should send "You are not holding a staff."
}

func TestDoBrandish_NoCharges(t *testing.T) {
	setupItemUseWorld()
	ch := newTestCharWithDesc()
	// Staff with 0 charges remaining, equipped at WEAR_HOLD
	makeEquippedObj(ch, "staff oak", "an oak staff", types.ITEM_STAFF, [6]int{10, 5, 0, 5, 0, 0}, types.WEAR_HOLD)

	DoBrandish(ch, "")
	// Should send "The staff has no charges remaining."
}

func TestDoBrandish_Success(t *testing.T) {
	setupItemUseWorld()
	ch := newTestCharWithDesc()
	ch.InRoom.People = append(ch.InRoom.People, ch)

	// Staff: level 10, max 5, current 3, spell slot 5 (cure light)
	staff := makeEquippedObj(ch, "staff oak", "an oak staff", types.ITEM_STAFF, [6]int{10, 5, 3, 5, 0, 0}, types.WEAR_HOLD)

	DoBrandish(ch, "")
	if staff.Value[2] != 2 {
		t.Errorf("expected charges to decrease to 2, got %d", staff.Value[2])
	}
}

func TestDoBrandish_LastCharge(t *testing.T) {
	setupItemUseWorld()
	ch := newTestCharWithDesc()
	ch.InRoom.People = append(ch.InRoom.People, ch)

	// Staff with 1 charge remaining
	makeEquippedObj(ch, "staff oak", "an oak staff", types.ITEM_STAFF, [6]int{10, 5, 1, 5, 0, 0}, types.WEAR_HOLD)

	DoBrandish(ch, "")
	// Staff should be destroyed (extracted) — no longer in carrying
	found := false
	for _, obj := range ch.Carrying {
		if obj.ItemType == types.ITEM_STAFF {
			found = true
		}
	}
	if found {
		t.Error("staff should be destroyed when last charge is used")
	}
}

// --- DoZap tests ---

func TestDoZap_NoWand(t *testing.T) {
	setupItemUseWorld()
	ch := newTestCharWithDesc()
	DoZap(ch, "")
	// Should send "You are not holding a wand."
}

func TestDoZap_NoCharges(t *testing.T) {
	setupItemUseWorld()
	ch := newTestCharWithDesc()
	makeEquippedObj(ch, "wand silver", "a silver wand", types.ITEM_WAND, [6]int{10, 5, 0, 5, 0, 0}, types.WEAR_HOLD)

	DoZap(ch, "")
	// Should send "The wand has no charges remaining."
}

func TestDoZap_SuccessSelf(t *testing.T) {
	setupItemUseWorld()
	ch := newTestCharWithDesc()
	wand := makeEquippedObj(ch, "wand silver", "a silver wand", types.ITEM_WAND, [6]int{10, 5, 3, 5, 0, 0}, types.WEAR_HOLD)

	DoZap(ch, "")
	if wand.Value[2] != 2 {
		t.Errorf("expected charges to decrease to 2, got %d", wand.Value[2])
	}
}

func TestDoZap_SuccessTarget(t *testing.T) {
	setupItemUseWorld()
	ch := newTestCharWithDesc()
	victim := newTestCharWithDesc()
	victim.Name = "Victim"
	victim.InRoom = ch.InRoom
	ch.InRoom.People = append(ch.InRoom.People, ch, victim)

	wand := makeEquippedObj(ch, "wand silver", "a silver wand", types.ITEM_WAND, [6]int{10, 5, 3, 5, 0, 0}, types.WEAR_HOLD)

	DoZap(ch, "Victim")
	if wand.Value[2] != 2 {
		t.Errorf("expected charges to decrease to 2, got %d", wand.Value[2])
	}
}

func TestDoZap_TargetFighting(t *testing.T) {
	setupItemUseWorld()
	ch := newTestCharWithDesc()
	mob := newTestCharWithDesc()
	mob.Name = "Goblin"
	mob.InRoom = ch.InRoom
	ch.InRoom.People = append(ch.InRoom.People, ch, mob)
	ch.Fighting = &types.FightData{Who: mob}

	wand := makeEquippedObj(ch, "wand silver", "a silver wand", types.ITEM_WAND, [6]int{10, 5, 3, 5, 0, 0}, types.WEAR_HOLD)

	// No argument, but fighting — should target fighting opponent
	DoZap(ch, "")
	if wand.Value[2] != 2 {
		t.Errorf("expected charges to decrease to 2, got %d", wand.Value[2])
	}
}

func TestDoZap_LastCharge(t *testing.T) {
	setupItemUseWorld()
	ch := newTestCharWithDesc()
	makeEquippedObj(ch, "wand silver", "a silver wand", types.ITEM_WAND, [6]int{10, 5, 1, 5, 0, 0}, types.WEAR_HOLD)

	DoZap(ch, "")
	found := false
	for _, obj := range ch.Carrying {
		if obj.ItemType == types.ITEM_WAND {
			found = true
		}
	}
	if found {
		t.Error("wand should be destroyed when last charge is used")
	}
}

func TestDoZap_TargetNotFound(t *testing.T) {
	setupItemUseWorld()
	ch := newTestCharWithDesc()
	ch.InRoom.People = append(ch.InRoom.People, ch)
	wand := makeEquippedObj(ch, "wand silver", "a silver wand", types.ITEM_WAND, [6]int{10, 5, 3, 5, 0, 0}, types.WEAR_HOLD)

	DoZap(ch, "nobody")
	// Should send "They aren't here." and not use charges
	if wand.Value[2] != 3 {
		t.Errorf("wand charges should not decrease when target not found, got %d", wand.Value[2])
	}
}
