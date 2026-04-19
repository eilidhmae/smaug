package act

import (
	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// gsnHide / gsnSneak are resolved once at boot via ResolveGSNs so the
// XP-on-skill-gain hot path in learnFromSuccess can silence the
// "You gain N experience points" message for stealth skills by an
// integer compare. C src/skills.c:1661 uses gsn_hide / gsn_sneak for
// the same suppression (the player would otherwise reveal their hide
// attempts to themselves even when Fighting is false). Left at -1 in
// unit tests where ResolveGSNs has not been called — the silentSkill
// branch simply never fires, which matches C on a fresh boot.
var (
	gsnHide  = -1
	gsnSneak = -1
)

// ResolveGSNs caches the gsn slot numbers act cares about for hot-path
// integer compares. Called once from boot.Boot after the skill table is
// loaded and persist.SkillNameLookup is wired. Safe to call multiple
// times (idempotent). Missing skills leave the cached slot at -1 so
// the suppression check silently fails safe.
func ResolveGSNs() {
	if combat.LookupSkillSlotHook == nil {
		return
	}
	gsnHide = combat.LookupSkillSlotHook("hide")
	gsnSneak = combat.LookupSkillSlotHook("sneak")
}

// CanUseSkill is the exported seam for combat.CanUseSkillHook. Thin
// wrapper so the combat package (which cannot import act) can still
// consult the same skill-check logic used by act commands.
func CanUseSkill(ch *types.CharData, percent int, gsn int) bool {
	return canUseSkill(ch, percent, gsn)
}

// LearnFromSuccess is the exported seam for combat.LearnFromSuccessHook.
func LearnFromSuccess(ch *types.CharData, gsn int) {
	learnFromSuccess(ch, gsn)
}

// LearnFromFailure is the exported seam for combat.LearnFromFailureHook.
func LearnFromFailure(ch *types.CharData, gsn int) {
	learnFromFailure(ch, gsn)
}

// canUseSkill checks if ch can use a skill at the given percent roll.
// For NPCs, 85% base success. For PCs, compares against learned skill %.
func canUseSkill(ch *types.CharData, percent int, gsn int) bool {
	if ch.IsNPC() {
		return percent < 85
	}
	if ch.PCData == nil || gsn < 0 || gsn >= types.MAX_SKILL {
		return false
	}
	return percent < ch.PCData.Learned[gsn]
}

// numberPercent is an indirection over util.NumberPercent so tests can
// replace it with a deterministic stub. Production code reads the real RNG.
var numberPercent = util.NumberPercent

// learnFromSuccess improves skill proficiency on successful use.
// Mirrors learn_from_success in src/skills.c:1621 — uses
// chance = learned + 5*difficulty and honours the per-class adept cap.
func learnFromSuccess(ch *types.CharData, gsn int) {
	if ch.IsNPC() || ch.PCData == nil || gsn < 0 || gsn >= types.MAX_SKILL {
		return
	}
	// C guard: only skills the player has already touched improve from use.
	if ch.PCData.Learned[gsn] <= 0 {
		return
	}
	if WorldRef == nil || gsn >= len(WorldRef.Skills) {
		return
	}
	skill := WorldRef.Skills[gsn]
	if skill == nil {
		return
	}
	learned := ch.PCData.Learned[gsn]
	adept := 100
	if ch.Class >= 0 && ch.Class < types.MAX_CLASS {
		adept = skill.SkillAdept[ch.Class]
	}
	if learned >= adept {
		return
	}
	chance := learned + 5*skill.Difficulty
	roll := numberPercent()
	gain := 0
	switch {
	case roll >= chance:
		gain = 2
	case chance-roll <= 25:
		gain = 1
	}
	if gain > 0 {
		ch.PCData.Learned[gsn] = util.UMIN(learned+gain, adept)
		ch.Sendf("You have become better at %s! (%d%%)\n\r", skill.Name, ch.PCData.Learned[gsn])

		// XP-on-gain — mirrors C src/skills.c:1641-1669.
		skLvl := skill.SkillLevel[ch.Class]
		if skLvl == 0 {
			skLvl = ch.Level
		}
		var xpGain int
		if ch.PCData.Learned[gsn] == adept {
			// Adept-cap branch (C src/skills.c:1644-1652) — player has
			// just fully learned the skill. Larger XP grant + a special
			// "now an adept" message in AT_WHITE.
			xpGain = 1000 * skLvl
			switch ch.Class {
			case types.CLASS_MAGE:
				xpGain *= 5
			case types.CLASS_CLERIC:
				xpGain *= 2
			}
			ch.Sendf("&WYou are now an adept of %s!  You gain %d bonus experience!\n\r&D",
				skill.Name, xpGain)
		} else {
			// Normal-gain branch (C src/skills.c:1654-1667). Silent
			// during combat and for hide/sneak (C :1661 guard — the
			// player would otherwise reveal their attempts to
			// themselves).
			xpGain = 20 * skLvl
			switch ch.Class {
			case types.CLASS_MAGE:
				xpGain *= 6
			case types.CLASS_CLERIC:
				xpGain *= 3
			}
			fighting := ch.Fighting != nil
			silentSkill := (gsnHide >= 0 && gsn == gsnHide) ||
				(gsnSneak >= 0 && gsn == gsnSneak)
			if !fighting && !silentSkill {
				ch.Sendf("You gain %d experience points from your success!\n\r", xpGain)
			}
		}
		ch.Exp += xpGain
	}
}

// learnFromFailure improves skill proficiency on failed use (slower).
// Mirrors learn_from_failure in src/skills.c:1658 — gain is capped at adept-1.
func learnFromFailure(ch *types.CharData, gsn int) {
	if ch.IsNPC() || ch.PCData == nil || gsn < 0 || gsn >= types.MAX_SKILL {
		return
	}
	// C guard: only skills the player has already touched improve from use.
	if ch.PCData.Learned[gsn] <= 0 {
		return
	}
	if WorldRef == nil || gsn >= len(WorldRef.Skills) {
		return
	}
	skill := WorldRef.Skills[gsn]
	if skill == nil {
		return
	}
	learned := ch.PCData.Learned[gsn]
	adept := 100
	if ch.Class >= 0 && ch.Class < types.MAX_CLASS {
		adept = skill.SkillAdept[ch.Class]
	}
	if learned >= adept-1 {
		return
	}
	chance := learned + 5*skill.Difficulty
	// Matches C src/skills.c:1682 — any roll within 25 of chance gains,
	// regardless of whether the roll beat chance. Previously included
	// an extra `roll < chance` conjunct that suppressed legitimate gains.
	if chance-numberPercent() > 25 {
		return
	}
	ch.PCData.Learned[gsn] = util.UMIN(learned+1, adept-1)
}

// --- Combat Skills ---

// DoBackstab implements the 'backstab' command.
func DoBackstab(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Backstab whom?\n\r")
		return
	}

	victim := handler.GetCharRoom(ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}
	if victim == ch {
		ch.Send("How can you backstab yourself?\n\r")
		return
	}
	if ch.Fighting != nil {
		ch.Send("You're too busy fighting!\n\r")
		return
	}

	wield := handler.GetEqChar(ch, types.WEAR_WIELD)
	if wield == nil {
		ch.Send("You need to wield a weapon to backstab.\n\r")
		return
	}

	gsn := lookupSkillSlot("backstab")
	if canUseSkill(ch, util.NumberPercent(), gsn) || victim.Position == types.POS_SLEEPING {
		learnFromSuccess(ch, gsn)
		// Backstab does weapon damage * level/10 multiplier
		mult := util.UMAX(2, ch.Level/10)
		dam := util.NumberRange(wield.Value[1], wield.Value[2]) * mult
		dam += ch.Damroll
		combat.Damage(WorldRef, ch, victim, dam, types.TYPE_HIT)
	} else {
		learnFromFailure(ch, gsn)
		combat.Damage(WorldRef, ch, victim, 0, types.TYPE_HIT)
	}
}

// DoBash implements the 'bash' command.
func DoBash(ch *types.CharData, argument string) {
	if ch.Fighting == nil {
		ch.Send("You aren't fighting anyone.\n\r")
		return
	}
	victim := ch.Fighting.Who

	gsn := lookupSkillSlot("bash")
	if canUseSkill(ch, util.NumberPercent(), gsn) {
		learnFromSuccess(ch, gsn)
		dam := util.NumberRange(1, ch.Level)
		victim.Position = types.POS_SITTING
		victim.Wait += 2 * types.PULSE_VIOLENCE
		combat.Damage(WorldRef, ch, victim, dam, types.TYPE_HIT)
		victim.Send("You are knocked down!\n\r")
		ch.Sendf("You bash %s!\n\r", victim.Name)
	} else {
		learnFromFailure(ch, gsn)
		combat.Damage(WorldRef, ch, victim, 0, types.TYPE_HIT)
		ch.Send("You miss your bash.\n\r")
	}
}

// DoKick implements the 'kick' command (combat skill).
func DoKickSkill(ch *types.CharData, argument string) {
	if ch.Fighting == nil {
		ch.Send("You aren't fighting anyone.\n\r")
		return
	}
	victim := ch.Fighting.Who

	gsn := lookupSkillSlot("kick")
	if canUseSkill(ch, util.NumberPercent(), gsn) {
		learnFromSuccess(ch, gsn)
		dam := util.NumberRange(1, ch.Level)
		combat.Damage(WorldRef, ch, victim, dam, types.TYPE_HIT)
	} else {
		learnFromFailure(ch, gsn)
		combat.Damage(WorldRef, ch, victim, 0, types.TYPE_HIT)
	}
}

// DoDisarm implements the 'disarm' command.
func DoDisarm(ch *types.CharData, argument string) {
	if ch.Fighting == nil {
		ch.Send("You aren't fighting anyone.\n\r")
		return
	}
	victim := ch.Fighting.Who

	if handler.GetEqChar(ch, types.WEAR_WIELD) == nil {
		ch.Send("You must wield a weapon to disarm.\n\r")
		return
	}

	victimWield := handler.GetEqChar(victim, types.WEAR_WIELD)
	if victimWield == nil {
		ch.Send("Your opponent is not wielding a weapon.\n\r")
		return
	}

	gsn := lookupSkillSlot("disarm")
	if canUseSkill(ch, util.NumberPercent(), gsn) {
		learnFromSuccess(ch, gsn)
		handler.UnequipChar(victim, victimWield)
		handler.ObjFromChar(victimWield)
		if victim.InRoom != nil {
			handler.ObjToRoom(victimWield, victim.InRoom)
		}
		ch.Sendf("You disarm %s!\n\r", victim.Name)
		victim.Send("Your weapon is knocked from your grasp!\n\r")
	} else {
		learnFromFailure(ch, gsn)
		ch.Send("You failed to disarm.\n\r")
	}
}

// DoRescue implements the 'rescue' command.
func DoRescue(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Rescue whom?\n\r")
		return
	}

	victim := handler.GetCharRoom(ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}
	if victim == ch {
		ch.Send("You can't rescue yourself.\n\r")
		return
	}
	if victim.Fighting == nil {
		ch.Send("They aren't fighting anyone.\n\r")
		return
	}

	attacker := victim.Fighting.Who
	gsn := lookupSkillSlot("rescue")

	if canUseSkill(ch, util.NumberPercent(), gsn) {
		learnFromSuccess(ch, gsn)
		combat.StopFighting(victim, false)
		combat.StopFighting(attacker, false)
		combat.StartFighting(ch, attacker)
		combat.StartFighting(attacker, ch)
		ch.Sendf("You rescue %s!\n\r", victim.Name)
		victim.Sendf("%s rescues you!\n\r", ch.Name)
	} else {
		learnFromFailure(ch, gsn)
		ch.Send("You fail the rescue.\n\r")
	}
}

// --- Stealth Skills ---

// DoSneak implements the 'sneak' command.
func DoSneak(ch *types.CharData, argument string) {
	handler.AffectStrip(ch, lookupSkillSlot("sneak"))

	gsn := lookupSkillSlot("sneak")
	if canUseSkill(ch, util.NumberPercent(), gsn) {
		learnFromSuccess(ch, gsn)
		aff := &types.AffectData{
			Type:     gsn,
			Duration: ch.Level,
			Location: types.APPLY_NONE,
		}
		aff.BitVector.Set(types.AFF_SNEAK)
		handler.AffectToChar(ch, aff)
		ch.Send("You begin to move silently.\n\r")
	} else {
		learnFromFailure(ch, gsn)
		ch.Send("You fail to move silently.\n\r")
	}
}

// DoHide implements the 'hide' command.
func DoHide(ch *types.CharData, argument string) {
	if ch.AffectedBy.IsSet(types.AFF_HIDE) {
		ch.AffectedBy.Toggle(types.AFF_HIDE)
	}

	gsn := lookupSkillSlot("hide")
	if canUseSkill(ch, util.NumberPercent(), gsn) {
		learnFromSuccess(ch, gsn)
		ch.AffectedBy.Set(types.AFF_HIDE)
		ch.Send("You attempt to hide.\n\r")
	} else {
		learnFromFailure(ch, gsn)
		ch.Send("You fail to hide.\n\r")
	}
}

// DoSteal implements the 'steal' command.
func DoSteal(ch *types.CharData, argument string) {
	arg1, arg2 := util.OneArgument(argument)
	if arg1 == "" || arg2 == "" {
		ch.Send("Steal what from whom?\n\r")
		return
	}

	victim := handler.GetCharRoom(ch, arg2)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}
	if victim == ch {
		ch.Send("That's pointless.\n\r")
		return
	}

	gsn := lookupSkillSlot("steal")

	if !canUseSkill(ch, util.NumberPercent(), gsn) {
		learnFromFailure(ch, gsn)
		ch.Send("Oops.\n\r")
		victim.Sendf("%s tried to steal from you!\n\r", ch.Name)
		return
	}

	learnFromSuccess(ch, gsn)

	if arg1 == "gold" || arg1 == "coins" {
		amount := victim.Gold / 10
		if amount <= 0 {
			ch.Send("You couldn't get any gold.\n\r")
			return
		}
		ch.Gold += amount
		victim.Gold -= amount
		ch.Sendf("You steal %d gold coins.\n\r", amount)
		return
	}

	obj := handler.GetObjCarry(victim, arg1)
	if obj == nil {
		ch.Send("You can't find it.\n\r")
		return
	}

	handler.ObjFromChar(obj)
	obj.CarriedBy = ch
	obj.WearLoc = types.WEAR_NONE
	ch.Carrying = append(ch.Carrying, obj)
	ch.Sendf("You steal %s.\n\r", obj.ShortDescr)
}

// DoPick implements the 'pick' command: pick a lock.
func DoPick(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Pick what?\n\r")
		return
	}

	gsn := lookupSkillSlot("pick lock")

	// Try door
	exit := findDoor(ch, arg)
	if exit != nil {
		if exit.ExitInfo&int(types.EX_ISDOOR) == 0 {
			ch.Send("That's not a door.\n\r")
			return
		}
		if exit.ExitInfo&int(types.EX_LOCKED) == 0 {
			ch.Send("It's not locked.\n\r")
			return
		}
		if exit.ExitInfo&int(types.EX_PICKPROOF) != 0 {
			ch.Send("You failed.\n\r")
			learnFromFailure(ch, gsn)
			return
		}

		if canUseSkill(ch, util.NumberPercent(), gsn) {
			learnFromSuccess(ch, gsn)
			exit.ExitInfo &^= int(types.EX_LOCKED)
			ch.Send("*Click*\n\r")
		} else {
			learnFromFailure(ch, gsn)
			ch.Send("You failed.\n\r")
		}
		return
	}

	ch.Send("You see nothing to pick.\n\r")
}

// DoBroach implements the 'broach' command: forcefully break a lock. Distinct
// from pick (thief lock-picking) — broach is a brute-force alternative that
// clears EX_LOCKED on both the forward exit and its reverse when the check
// succeeds. src/skills.c:3960.
//
// C predicate bug fix (plan Open Question 2, src/skills.c:3987-3990): C reads
//
//	if (!CLOSED || !LOCKED || PICKPROOF || can_use_skill) { fail; }
//
// which is a tautology on unlocked/open doors AND inverts the can_use_skill
// sense (fails on success). We port the reconstructed intent:
//
//	success := CLOSED && LOCKED && !PICKPROOF && can_use_skill
//
// Also drops the C calls to adjust_favor (deity subsystem unported; plan
// Scope Cuts) and check_room_for_traps (trap subsystem unported).
//
// Color: C calls set_char_color(AT_DGREEN, ch) once and then send_to_char
// inherits the color. Go has no per-descriptor color state; we prefix each
// send with the &g (dark green) inline token to preserve the visual intent
// (plan Open Question 3, Option A).
func DoBroach(ch *types.CharData, argument string) {
	// NPC + AFF_CHARM early-return (C :3967-3971).
	if ch.IsNPC() && ch.AffectedBy.IsSet(types.AFF_CHARM) {
		ch.Send("&gYou can't concentrate enough for that.\n\r")
		return
	}

	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("&gAttempt this in which direction?\n\r")
		return
	}

	// Mount gate (C :3978-3982).
	if ch.Mount != nil {
		ch.Send("&gYou should really dismount first.\n\r")
		return
	}

	gsn := lookupSkillSlot("broach")

	// WAIT_STATE uses skill_table[gsn_broach]->beats (= 24 per skills.dat). C
	// sets this unconditionally before the find-door branch so a bogus
	// direction still spends the wait.
	if WorldRef != nil && gsn >= 0 && gsn < len(WorldRef.Skills) && WorldRef.Skills[gsn] != nil {
		ch.Wait = WorldRef.Skills[gsn].Beats
	}

	exit := findDoor(ch, arg)
	if exit == nil {
		// C's find_door(ch, arg, TRUE) is silent on failure; C then falls
		// through to the trailing "Your attempt fails." send.
		ch.Send("&gYour attempt fails.\n\r")
		return
	}

	closed := exit.ExitInfo&int(types.EX_CLOSED) != 0
	locked := exit.ExitInfo&int(types.EX_LOCKED) != 0
	pickproof := exit.ExitInfo&int(types.EX_PICKPROOF) != 0
	skillPass := canUseSkill(ch, numberPercent(), gsn)

	if !(closed && locked && !pickproof && skillPass) {
		ch.Send("&gYour attempt fails.\n\r")
		learnFromFailure(ch, gsn)
		// TODO: adjust_favor / check_room_for_traps — unported subsystems.
		return
	}

	// Success: clear EX_LOCKED on forward exit and, when the reverse exit
	// points back to ch's room, clear it there too (C :4001-4005).
	exit.ExitInfo &^= int(types.EX_LOCKED)
	if exit.ReverseExit != nil && exit.ReverseExit.ToRoom == ch.InRoom {
		exit.ReverseExit.ExitInfo &^= int(types.EX_LOCKED)
	}
	ch.Send("&gYou successfully broach the exit...\n\r")
	learnFromSuccess(ch, gsn)
	// TODO: adjust_favor(ch, 9, 1) — deity subsystem unported.
	// TODO: check_room_for_traps(ch, TRAP_PICK | trap_door[pexit->vdir]) —
	// trap dispatcher unported.
}

// --- Utility Skills ---

// DoScan implements the 'scan' command: look in adjacent rooms.
func DoScan(ch *types.CharData, argument string) {
	if ch.InRoom == nil {
		return
	}

	ch.Send("You scan your surroundings...\n\r")

	dirNames := []string{"north", "east", "south", "west", "up", "down"}
	for dir := 0; dir < 6; dir++ {
		exit := ch.InRoom.GetExit(dir)
		if exit == nil || exit.ToRoom == nil {
			continue
		}
		if exit.ExitInfo&int(types.EX_CLOSED) != 0 {
			continue
		}

		room := exit.ToRoom
		if len(room.People) > 0 {
			ch.Sendf("  %s:\n\r", util.Capitalize(dirNames[dir]))
			for _, rch := range room.People {
				if rch.IsNPC() {
					ch.Sendf("    %s\n\r", rch.ShortDescr)
				} else {
					ch.Sendf("    %s\n\r", rch.Name)
				}
			}
		}
	}
}

// DoAid implements the 'aid' command: help an incapacitated character.
func DoAid(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Aid whom?\n\r")
		return
	}

	victim := handler.GetCharRoom(ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}

	if victim.Position > types.POS_STUNNED {
		ch.Send("They don't need your aid.\n\r")
		return
	}
	if victim.Hit >= 1 {
		ch.Send("They don't need your aid.\n\r")
		return
	}

	gsn := lookupSkillSlot("aid")
	if canUseSkill(ch, util.NumberPercent(), gsn) {
		learnFromSuccess(ch, gsn)
		victim.Hit = 1
		victim.Position = types.POS_RESTING
		ch.Sendf("You aid %s.\n\r", victim.Name)
		victim.Send("Someone aids you!\n\r")
	} else {
		learnFromFailure(ch, gsn)
		ch.Send("You fail.\n\r")
	}
}

// DoRecall implements the 'recall' command: return to the temple.
func DoRecall(ch *types.CharData, argument string) {
	if ch.Fighting != nil {
		ch.Send("You can't recall while fighting!\n\r")
		return
	}

	if ch.InRoom != nil && ch.InRoom.RoomFlags.IsSet(types.ROOM_NO_RECALL) {
		ch.Send("You failed.\n\r")
		return
	}

	room := WorldRef.GetRoom(types.ROOM_VNUM_TEMPLE)
	if room == nil {
		ch.Send("You are completely lost.\n\r")
		return
	}

	if ch.InRoom == room {
		ch.Send("You are already there.\n\r")
		return
	}

	ch.Move /= 2
	if ch.InRoom != nil {
		for _, rch := range ch.InRoom.People {
			if rch != ch && rch.Desc != nil {
				rch.Sendf("%s disappears in a flash of light.\n\r", ch.Name)
			}
		}
		handler.CharFromRoom(ch)
	}
	handler.CharToRoom(ch, room)
	ch.Send("You recall to the temple!\n\r")
	DoLook(ch, "auto")
}
