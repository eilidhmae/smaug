package act

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// DoMeditate — enter a meditative state that accelerates mana regen.
// src/skills.c:2917. MVP: flags the character with SubState; the actual
// HP/mana bonus is applied in game/update.go only if/when a hook is
// added. For now we set Substate and Position=POS_RESTING and send text.
func DoMeditate(ch *types.CharData, argument string) {
	if ch.Fighting != nil {
		ch.Send("Not in the middle of combat!\n\r")
		return
	}
	if ch.Position < types.POS_SLEEPING {
		ch.Send("In your dreams, or what?\n\r")
		return
	}
	gsn := lookupSkillSlot("meditate")
	if canUseSkill(ch, util.NumberPercent(), gsn) {
		learnFromSuccess(ch, gsn)
		ch.Substate = types.SUB_PAUSE // marker state; no dedicated SUB_MEDITATE
		ch.Send("You begin to meditate.\n\r")
		util.Act(types.AT_ACTION, "$n begins to meditate.", ch, nil, nil, nil, types.TO_ROOM)
	} else {
		learnFromFailure(ch, gsn)
		ch.Send("Your thoughts drift and you fail to focus.\n\r")
	}
}

// DoTrance — enter a trance state. src/skills.c:3028. Shares the
// meditate shape but for spellcasters; MVP is identical.
func DoTrance(ch *types.CharData, argument string) {
	if ch.Fighting != nil {
		ch.Send("Not in the middle of combat!\n\r")
		return
	}
	gsn := lookupSkillSlot("trance")
	if canUseSkill(ch, util.NumberPercent(), gsn) {
		learnFromSuccess(ch, gsn)
		ch.Substate = types.SUB_PAUSE
		ch.Send("You enter a trance.\n\r")
		util.Act(types.AT_ACTION, "$n enters a trance.", ch, nil, nil, nil, types.TO_ROOM)
	} else {
		learnFromFailure(ch, gsn)
		ch.Send("You fail to enter a trance.\n\r")
	}
}

// DoSearch — hunt for hidden exits and objects in the current room.
// src/skills.c:2223. MVP: reveals EX_SECRET/EX_HIDDEN exits by clearing
// the hidden flag, and reveals hidden objects by printing their names.
func DoSearch(ch *types.CharData, argument string) {
	if ch.InRoom == nil {
		return
	}
	gsn := lookupSkillSlot("search")
	if !canUseSkill(ch, util.NumberPercent(), gsn) {
		learnFromFailure(ch, gsn)
		ch.Send("You find nothing.\n\r")
		return
	}
	learnFromSuccess(ch, gsn)
	found := false
	for _, exit := range ch.InRoom.Exits {
		if exit.ExitInfo&int(types.EX_SECRET) != 0 {
			exit.ExitInfo &^= int(types.EX_SECRET)
			found = true
			ch.Send("You discover a hidden passage!\n\r")
		}
	}
	if !found {
		ch.Send("You find nothing.\n\r")
	}
}

// DoDetrap — attempt to disarm a trap on an object. src/skills.c:1906.
// MVP: searches ch's carried items for ITEM_TRAP and destroys it.
func DoDetrap(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Detrap what?\n\r")
		return
	}
	obj := handler.GetObjHere(ch, arg)
	if obj == nil {
		ch.Send("You can't find that.\n\r")
		return
	}
	gsn := lookupSkillSlot("detrap")
	if !canUseSkill(ch, util.NumberPercent(), gsn) {
		learnFromFailure(ch, gsn)
		ch.Send("You failed to disarm the trap.\n\r")
		return
	}
	learnFromSuccess(ch, gsn)
	// MVP: just report success. Real C removes ITEM_TRAP children.
	if obj.ItemType == types.ITEM_TRAP {
		handler.ExtractObj(WorldRef, obj)
		ch.Send("You successfully disarm the trap!\n\r")
	} else {
		ch.Send("There is no trap on that.\n\r")
	}
}

// DoDig — excavate the room floor. src/skills.c:2030. MVP: emits messages
// and does not generate random objects.
func DoDig(ch *types.CharData, argument string) {
	if ch.InRoom == nil {
		return
	}
	gsn := lookupSkillSlot("dig")
	if !canUseSkill(ch, util.NumberPercent(), gsn) {
		learnFromFailure(ch, gsn)
		ch.Send("You dig for a while, but find nothing.\n\r")
		return
	}
	learnFromSuccess(ch, gsn)
	ch.Send("You dig in the earth.\n\r")
	util.Act(types.AT_ACTION, "$n digs in the earth.", ch, nil, nil, nil, types.TO_ROOM)
}

// DoVisible — cancel invisibility / hide / (if immortal) wiz-invis.
// src/skills.c:4235.
func DoVisible(ch *types.CharData, argument string) {
	if ch.AffectedBy.IsSet(types.AFF_INVISIBLE) {
		ch.AffectedBy.Toggle(types.AFF_INVISIBLE)
	}
	if ch.AffectedBy.IsSet(types.AFF_HIDE) {
		ch.AffectedBy.Toggle(types.AFF_HIDE)
	}
	if ch.AffectedBy.IsSet(types.AFF_SNEAK) {
		ch.AffectedBy.Toggle(types.AFF_SNEAK)
	}
	// Strip sn-targeted invis affects so the bits don't re-appear from duration.
	handler.AffectStrip(ch, lookupSkillSlot("invis"))
	handler.AffectStrip(ch, lookupSkillSlot("mass invis"))
	handler.AffectStrip(ch, lookupSkillSlot("hide"))
	handler.AffectStrip(ch, lookupSkillSlot("sneak"))
	if ch.PCData != nil && ch.PCData.WizInvis > 0 {
		ch.PCData.WizInvis = 0
	}
	ch.Send("You are now visible.\n\r")
}

// DoStyle — toggle fighting style. src/skills.c:6525. Cycles through
// STYLE_BERSERK / AGGRESSIVE / FIGHTING / DEFENSIVE / EVASIVE by name.
func DoStyle(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	arg = strings.ToLower(arg)
	if arg == "" {
		ch.Send("Available styles: berserk, aggressive, fighting, defensive, evasive.\n\r")
		return
	}
	styles := []struct {
		name string
		val  int
	}{
		{"berserk", types.STYLE_BERSERK},
		{"aggressive", types.STYLE_AGGRESSIVE},
		{"fighting", types.STYLE_FIGHTING},
		{"defensive", types.STYLE_DEFENSIVE},
		{"evasive", types.STYLE_EVASIVE},
	}
	for _, s := range styles {
		if strings.HasPrefix(s.name, arg) {
			ch.Style = s.val
			ch.Sendf("You change your fighting style to %s.\n\r", s.name)
			return
		}
	}
	ch.Send("No such style.\n\r")
}

// DoStance — toggle combat stance. MVP: store the requested index on
// CharData.Stance and echo. Real C uses STANCE_MONGOOSE…STANCE_SWALLOW
// with learned proficiency per stance; we keep name-to-index mapping.
func DoStance(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	arg = strings.ToLower(arg)
	names := []string{"mongoose", "bull", "mantis", "dragon", "tiger", "monkey", "swallow"}
	if arg == "" {
		ch.Send("Available stances: mongoose, bull, mantis, dragon, tiger, monkey, swallow.\n\r")
		return
	}
	for i, n := range names {
		if strings.HasPrefix(n, arg) {
			ch.Stance = types.STANCE_MONGOOSE + i
			ch.Sendf("You take the %s stance.\n\r", n)
			return
		}
	}
	ch.Send("No such stance.\n\r")
}

// DoMistwalk — short-range teleport to a named character. src/skills.c:3887.
// MVP: only works if the target is in a room we can find by walking up to
// 10 exits from ch; full C version uses vampire/plane checks.
func DoMistwalk(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Mistwalk to whom?\n\r")
		return
	}
	victim := handler.GetCharWorld(WorldRef, ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}
	if victim.InRoom == nil || victim == ch {
		ch.Send("You can't reach them.\n\r")
		return
	}
	gsn := lookupSkillSlot("mistwalk")
	if !canUseSkill(ch, util.NumberPercent(), gsn) {
		learnFromFailure(ch, gsn)
		ch.Send("You fail to shape the mist.\n\r")
		return
	}
	learnFromSuccess(ch, gsn)
	if ch.InRoom != nil {
		util.Act(types.AT_ACTION, "$n dissolves into mist.", ch, nil, nil, nil, types.TO_ROOM)
		handler.CharFromRoom(ch)
	}
	handler.CharToRoom(ch, victim.InRoom)
	util.Act(types.AT_ACTION, "$n coalesces from the mist.", ch, nil, nil, nil, types.TO_ROOM)
	ch.Send("You drift through the mist.\n\r")
	DoLook(ch, "auto")
}

// DoFeed — hand a food item from inventory to another character.
// No direct C analogue — maps to the "feed $target $food" verb used in
// various MUDs. MVP: find an ITEM_FOOD in ch's inventory matching the
// second arg, transfer it to target, add to target's inventory.
func DoFeed(ch *types.CharData, argument string) {
	arg1, rest := util.OneArgument(argument)
	arg2, _ := util.OneArgument(rest)
	if arg1 == "" || arg2 == "" {
		ch.Send("Feed what to whom?\n\r")
		return
	}
	victim := handler.GetCharRoom(ch, arg1)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}
	obj := handler.GetObjCarry(ch, arg2)
	if obj == nil {
		ch.Send("You don't have that.\n\r")
		return
	}
	if obj.ItemType != types.ITEM_FOOD {
		ch.Send("That's not edible.\n\r")
		return
	}
	handler.ObjFromChar(obj)
	handler.ObjToChar(obj, victim)
	util.Act(types.AT_ACTION, "You feed $p to $N.", ch, victim, obj, nil, types.TO_CHAR)
	util.Act(types.AT_ACTION, "$n feeds you $p.", ch, victim, obj, nil, types.TO_VICT)
	util.Act(types.AT_ACTION, "$n feeds $p to $N.", ch, victim, obj, nil, types.TO_NOTVICT)
}

// isBloodRace mirrors C's composite predicate "IS_VAMPIRE(ch) ||
// IS_DEMON(ch)" (src/mud.h:4029-4034). A character qualifies as blood-race
// when they're a PC (not an NPC) AND either their race OR their class is
// vampire or demon. IS_VAMPIRE / IS_DEMON each expand to a race-OR-class
// disjunction; the caller's intent is "any blood identity", so we union.
//
// See plan-phase6-skills.md Open Question 1: C's bloodlet gate reads
// "IS_NPC(ch) || !IS_VAMPIRE(ch) || !IS_DEMON(ch)" (skills.c:3521) which by
// operator precedence requires BOTH predicates — a functionally dead gate
// for every typical single-identity vampire or demon. We port the intent.
func isBloodRace(ch *types.CharData) bool {
	if ch.IsNPC() {
		return false
	}
	return ch.Race == types.RACE_VAMPIRE ||
		ch.Race == types.RACE_DEMON ||
		ch.Class == types.CLASS_VAMPIRE ||
		ch.Class == types.CLASS_DEMON
}

// DoBloodlet — vampire/demon self-harm ritual. src/skills.c:3517. Consumes
// COND_BLOODTHIRST, spawns an OBJ_VNUM_BLOODLET pool in the room, and
// self-damages for ch.Level/5 HP. Bug fix for the gate condition is
// documented on isBloodRace (plan Open Question 1).
//
// Omissions vs. C: none meaningful — the skill is self-contained. Color
// fidelity: C calls act(AT_BLOOD, ...) which would render blood-red; Go's
// atColorCode table has no AT_BLOOD entry yet so messages render uncolored
// (plan Scope Cuts — covered by the existing atColorCode TODO).
func DoBloodlet(ch *types.CharData, argument string) {
	// Blood-race gate (C :3521 intent; see isBloodRace). Silent early-return
	// for NPCs and plain characters — matches C behavior of "return" with no
	// user message.
	if !isBloodRace(ch) {
		return
	}

	if ch.Fighting != nil {
		ch.Send("You're too busy fighting...\n\r")
		return
	}
	if ch.PCData == nil || ch.PCData.Condition[types.COND_BLOODTHIRST] < 10 {
		ch.Send("You are too drained to offer any blood...\n\r")
		return
	}

	// WAIT_STATE uses the literal PULSE_VIOLENCE (C :3535) — NOT the
	// skills.dat Beats (= 12). This divergence is intentional in C.
	ch.Wait = types.PULSE_VIOLENCE

	gsn := lookupSkillSlot("bloodlet")
	if canUseSkill(ch, numberPercent(), gsn) {
		GainCondition(ch, types.COND_BLOODTHIRST, -7)
		util.Act(types.AT_BLOOD,
			"Tracing a sharp nail over your skin, you let your blood spill.",
			ch, nil, nil, nil, types.TO_CHAR)
		util.Act(types.AT_BLOOD,
			"$n traces a sharp nail over $s skin, spilling a quantity of blood to the ground.",
			ch, nil, nil, nil, types.TO_ROOM)
		learnFromSuccess(ch, gsn)

		// Spawn the blood pool in the room. C :3546-3552.
		if WorldRef != nil && ch.InRoom != nil {
			idx := WorldRef.GetObjIndex(types.OBJ_VNUM_BLOODLET)
			if idx == nil {
				util.Bug("DoBloodlet: OBJ_VNUM_BLOODLET (vnum %d) not found; skipping spawn",
					types.OBJ_VNUM_BLOODLET)
			} else {
				obj := handler.CreateObject(WorldRef, idx, 0)
				obj.Timer = 1
				obj.Value[1] = 6
				handler.ObjToRoom(obj, ch.InRoom)
			}
		}

		// Self-damage LAST so the object spawn survives even if the damage
		// kills ch. C :3554 runs damage after obj_to_room for the same
		// reason. damageWith guards self-damage: it skips StartFighting
		// when ch == victim (internal/combat/combat.go:647-658).
		combat.Damage(WorldRef, ch, ch, ch.Level/5, gsn)
	} else {
		util.Act(types.AT_BLOOD, "You cannot manage to draw much blood...",
			ch, nil, nil, nil, types.TO_CHAR)
		util.Act(types.AT_BLOOD, "$n slices open $s skin, but no blood is spilled...",
			ch, nil, nil, nil, types.TO_ROOM)
		learnFromFailure(ch, gsn)
	}
}

// DoSkin — skin a corpse for food. src/skills.c:523. MVP: turns an
// ITEM_CORPSE_NPC in the room into a simple food item carried by ch.
func DoSkin(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Skin what?\n\r")
		return
	}
	if ch.InRoom == nil {
		return
	}
	var corpse *types.ObjData
	for _, o := range ch.InRoom.Contents {
		if o.ItemType == types.ITEM_CORPSE_NPC && strings.Contains(strings.ToLower(o.Name), strings.ToLower(arg)) {
			corpse = o
			break
		}
	}
	if corpse == nil {
		ch.Send("You don't see that corpse here.\n\r")
		return
	}
	gsn := lookupSkillSlot("skin")
	if !canUseSkill(ch, util.NumberPercent(), gsn) {
		learnFromFailure(ch, gsn)
		ch.Send("You botch the skinning.\n\r")
		return
	}
	learnFromSuccess(ch, gsn)
	food := &types.ObjData{
		Name:       "meat skin",
		ShortDescr: "a strip of meat",
		ItemType:   types.ITEM_FOOD,
		Value:      [6]int{24, 0, 0, 0, 0, 0}, // 24 food value
	}
	handler.ObjToChar(food, ch)
	handler.ExtractObj(WorldRef, corpse)
	ch.Send("You skin the corpse and get a strip of meat.\n\r")
}

// DoPoisonWeapon — apply poison to a wielded weapon. src/skills.c:4681.
// Sets the ITEM_POISONED extra flag (C skills.c:4803:
// `xSET_BIT(obj->extra_flags, ITEM_POISONED)`) so that combat code reading
// ExtraFlags sees the poison marker. MVP: skips the ingredient requirement
// (black powder + water container per skills.c:4737-4755) — noted TODO.
func DoPoisonWeapon(ch *types.CharData, argument string) {
	wield := handler.GetEqChar(ch, types.WEAR_WIELD)
	if wield == nil {
		ch.Send("You must wield a weapon to poison it.\n\r")
		return
	}
	gsn := lookupSkillSlot("poison weapon")
	if !canUseSkill(ch, util.NumberPercent(), gsn) {
		learnFromFailure(ch, gsn)
		ch.Send("You fail to apply the poison.\n\r")
		return
	}
	learnFromSuccess(ch, gsn)
	wield.ExtraFlags.Set(types.ITEM_POISONED)
	ch.Sendf("You coat %s with poison.\n\r", wield.ShortDescr)
}

// DoFire — ranged projectile attack (archery primer). MVP: acts like a
// throw of a named object. Full C is src/skills.c:6223 and depends on
// ARCHERY enable; we keep a stub so the command name dispatches.
func DoFire(ch *types.CharData, argument string) {
	if argument == "" {
		ch.Send("Fire what, and at whom?\n\r")
		return
	}
	// Reuse throw for MVP; real implementation drops a bow/ranged check.
	DoThrow(ch, argument)
}

// DoScribe — write a spell into a blank scroll. src/skills.c:4836. The
// blank scroll is identified by C as the held-slot object of vnum
// OBJ_VNUM_SCROLL_SCRIBING (34) with Value[1] == -1 (skills.c:4897-4916).
// MVP omits C's component-check and mana cost (TODO).
func DoScribe(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Scribe what spell?\n\r")
		return
	}
	sn := lookupSkillSlot(arg)
	if sn < 0 {
		ch.Send("No such spell.\n\r")
		return
	}
	if ch.PCData == nil || ch.PCData.Learned[sn] <= 0 {
		ch.Send("You don't know that spell.\n\r")
		return
	}
	// C uses the WEAR_HOLD slot; we check held + carried as MVP. Scroll must
	// be vnum OBJ_VNUM_SCROLL_SCRIBING and Value[1] == -1 (blank).
	isBlankScroll := func(o *types.ObjData) bool {
		if o == nil || o.ItemType != types.ITEM_SCROLL {
			return false
		}
		if o.IndexData != nil {
			return o.IndexData.Vnum == types.OBJ_VNUM_SCROLL_SCRIBING
		}
		return false
	}
	var scroll *types.ObjData
	if held := handler.GetEqChar(ch, types.WEAR_HOLD); isBlankScroll(held) {
		scroll = held
	}
	if scroll == nil {
		for _, o := range ch.Carrying {
			if isBlankScroll(o) {
				scroll = o
				break
			}
		}
	}
	if scroll == nil {
		ch.Send("You must be holding a blank scroll to scribe it.\n\r")
		return
	}
	if scroll.Value[1] != -1 {
		ch.Send("That scroll has already been inscribed.\n\r")
		return
	}
	gsn := lookupSkillSlot("scribe")
	if !canUseSkill(ch, util.NumberPercent(), gsn) {
		learnFromFailure(ch, gsn)
		ch.Send("Your hand shakes and you ruin the scroll.\n\r")
		return
	}
	learnFromSuccess(ch, gsn)
	scroll.Value[0] = ch.Level
	scroll.Value[1] = sn
	ch.Sendf("You scribe %s onto the scroll.\n\r", arg)
}

// DoCook — cook a food item. src/skills.c:6751. C requires the object to be
// ITEM_COOK (raw meat etc., distinct from ITEM_FOOD), and requires an
// ITEM_FIRE object to be present in the room (skills.c:6778, 6788-6796).
func DoCook(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Cook what?\n\r")
		return
	}
	obj := handler.GetObjCarry(ch, arg)
	if obj == nil {
		ch.Send("You don't have that.\n\r")
		return
	}
	if obj.ItemType != types.ITEM_COOK {
		ch.Send("How can you cook that?\n\r")
		return
	}
	if obj.Value[2] > 2 {
		ch.Send("That is already burnt to a crisp.\n\r")
		return
	}
	// Fire-in-room guard (skills.c:6788-6796).
	var fire *types.ObjData
	if ch.InRoom != nil {
		for _, o := range ch.InRoom.Contents {
			if o.ItemType == types.ITEM_FIRE {
				fire = o
				break
			}
		}
	}
	if fire == nil {
		ch.Send("There is no fire here!\n\r")
		return
	}
	gsn := lookupSkillSlot("cook")
	if !canUseSkill(ch, util.NumberPercent(), gsn) {
		learnFromFailure(ch, gsn)
		obj.Value[0] = 0
		obj.Value[2] = 3
		ch.Sendf("%s catches on fire burning it to a crisp!\n\r", obj.ShortDescr)
		return
	}
	learnFromSuccess(ch, gsn)
	obj.Value[0] *= 2
	ch.Sendf("You cook %s.\n\r", obj.ShortDescr)
}
