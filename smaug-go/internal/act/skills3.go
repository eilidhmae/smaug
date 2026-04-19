package act

import (
	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// unarmedAttack is the shared helper for bite/claw/punch/sting/tail.
// Ports the identical do_bite/do_claw/do_punch/do_sting/do_tail bodies at
// src/skills.c:3180–3371 — all five differ only in skill name and the
// flavour-text returned by who_fighting failure, which we collapse to a
// single "You aren't fighting anyone." message.
func unarmedAttack(ch *types.CharData, skillName string) {
	if ch.Fighting == nil {
		ch.Send("You aren't fighting anyone.\n\r")
		return
	}
	victim := ch.Fighting.Who
	gsn := lookupSkillSlot(skillName)
	// MVP: skip per-class level gating (C checks skill_level[class]). The
	// can_use_skill proficiency check already covers the common case where
	// a class that shouldn't have the skill has Learned=0.
	if canUseSkill(ch, util.NumberPercent(), gsn) {
		learnFromSuccess(ch, gsn)
		dam := util.NumberRange(1, ch.Level)
		combat.Damage(WorldRef, ch, victim, dam, gsn)
	} else {
		learnFromFailure(ch, gsn)
		combat.Damage(WorldRef, ch, victim, 0, gsn)
	}
}

// DoBite — natural unarmed attack. src/skills.c:3220.
func DoBite(ch *types.CharData, argument string) { unarmedAttack(ch, "bite") }

// DoClaw — natural unarmed attack. src/skills.c:3260.
func DoClaw(ch *types.CharData, argument string) { unarmedAttack(ch, "claw") }

// DoPunch — martial-arts unarmed attack. src/skills.c:3180.
func DoPunch(ch *types.CharData, argument string) { unarmedAttack(ch, "punch") }

// DoSting — natural unarmed attack. src/skills.c:3295.
func DoSting(ch *types.CharData, argument string) { unarmedAttack(ch, "sting") }

// DoTail — natural unarmed attack. src/skills.c:3335.
func DoTail(ch *types.CharData, argument string) { unarmedAttack(ch, "tail") }

// --- Wave 2 — Offensive combat ---

// DoCircle — reposition behind an engaged opponent for a piercing hit.
// src/skills.c:5136. MVP: no weapon-type check (C requires pierce/stab);
// we still require a wielded weapon; both multi-hit and the "distraction"
// num_fighting check are simplified to "victim must be fighting someone
// other than ch".
func DoCircle(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Circle around whom?\n\r")
		return
	}
	victim := handler.GetCharRoom(ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}
	if victim == ch {
		ch.Send("How can you sneak up on yourself?\n\r")
		return
	}
	if handler.GetEqChar(ch, types.WEAR_WIELD) == nil {
		ch.Send("You need to wield a piercing or stabbing weapon.\n\r")
		return
	}
	if ch.Fighting == nil {
		ch.Send("You can't circle when you aren't fighting.\n\r")
		return
	}
	if victim.Fighting == nil {
		ch.Send("You can't circle around a person who is not fighting.\n\r")
		return
	}
	if victim.Fighting.Who == ch {
		ch.Send("You can't circle around them without a distraction.\n\r")
		return
	}
	gsn := lookupSkillSlot("circle")
	if canUseSkill(ch, util.NumberPercent(), gsn) {
		learnFromSuccess(ch, gsn)
		// C fight.c:997-999: gsn_circle is one of the dt values that
		// short-circuits MultiHit's cascade — circle is meant to be a
		// single-hit skill. We preserve the historical Go MVP behavior of
		// firing two OneHits for the "multi" feel here in DoCircle itself;
		// MultiHit's cascade is explicitly suppressed for circle.
		//
		// Retcode propagation: the second OneHit must bail if the first
		// killed the victim OR killed the attacker (reactive damage
		// like fireshield is future work, but the hook is in place).
		ret := combat.OneHit(WorldRef, ch, victim, gsn)
		if !combat.VictimDied(ret) && !combat.AttackerDied(ret) &&
			victim.Position > types.POS_DEAD && ch.Position > types.POS_DEAD {
			_ = combat.OneHit(WorldRef, ch, victim, gsn)
		}
	} else {
		learnFromFailure(ch, gsn)
		combat.Damage(WorldRef, ch, victim, 0, gsn)
	}
}

// DoGouge — strike to blind the target. src/skills.c:1810. MVP: skip the
// dex-vs-dex chance modifier (combined into base percent) and the sysdata
// plr_vs_plr tweaks.
func DoGouge(ch *types.CharData, argument string) {
	if ch.Fighting == nil {
		ch.Send("You aren't fighting anyone.\n\r")
		return
	}
	victim := ch.Fighting.Who
	gsn := lookupSkillSlot("gouge")
	if canUseSkill(ch, util.NumberPercent(), gsn) {
		learnFromSuccess(ch, gsn)
		dam := util.NumberRange(5, ch.Level)
		combat.Damage(WorldRef, ch, victim, dam, gsn)
		// Apply blindness if target didn't die and isn't already blind.
		if victim.Position > types.POS_DEAD && !victim.AffectedBy.IsSet(types.AFF_BLIND) {
			aff := &types.AffectData{
				Type:     lookupSkillSlot("blindness"),
				Duration: 3 + ch.Level/15,
				Location: types.APPLY_HITROLL,
				Modifier: -6,
			}
			aff.BitVector.Set(types.AFF_BLIND)
			handler.AffectToChar(victim, aff)
			victim.Send("You can't see a thing!\n\r")
		}
	} else {
		learnFromFailure(ch, gsn)
		combat.Damage(WorldRef, ch, victim, 0, gsn)
	}
}

// DoStun — charge attack that paralyses on hit. src/skills.c:3425.
// MVP: no move-point cost, simplified dex/str chance adjustment.
func DoStun(ch *types.CharData, argument string) {
	if ch.Fighting == nil {
		ch.Send("You aren't fighting anyone.\n\r")
		return
	}
	victim := ch.Fighting.Who
	gsn := lookupSkillSlot("stun")
	if canUseSkill(ch, util.NumberPercent(), gsn) {
		learnFromSuccess(ch, gsn)
		util.Act(types.AT_HITME, "$N smashes into you, leaving you stunned!", victim, ch, nil, nil, types.TO_CHAR)
		util.Act(types.AT_ACTION, "You smash into $N, leaving $M stunned!", ch, victim, nil, nil, types.TO_CHAR)
		util.Act(types.AT_ACTION, "$n smashes into $N, leaving $M stunned!", ch, victim, nil, nil, types.TO_NOTVICT)
		if !victim.AffectedBy.IsSet(types.AFF_PARALYSIS) {
			aff := &types.AffectData{
				Type:     gsn,
				Duration: 3,
				Location: types.APPLY_AC,
				Modifier: 20,
			}
			aff.BitVector.Set(types.AFF_PARALYSIS)
			handler.AffectToChar(victim, aff)
			victim.Position = types.POS_STUNNED
		}
	} else {
		learnFromFailure(ch, gsn)
		util.Act(types.AT_ACTION, "You try to stun $N, but $E dodges out of the way.", ch, victim, nil, nil, types.TO_CHAR)
		util.Act(types.AT_HITME, "$n charges at you screaming, but you dodge out of the way.", ch, victim, nil, nil, types.TO_VICT)
	}
}

// DoPounce — stealth opening attack. src/skills.c:2607. Requires a wielded
// weapon with one of {DAM_SLICE, DAM_STAB, DAM_SLASH, DAM_CLAW, DAM_BITE,
// DAM_PIERCE} in Value[3], a non-fighting target that is either asleep or at
// full HP, and not mounted. The combat-side short-circuit for gsnPounce in
// MultiHit (internal/combat/combat.go:247) makes this a single-hit skill.
//
// Omissions vs. C:
//   - is_safe / check_attacker / check_illegal_pk — PK-legality subsystem not
//     ported in Go (matches DoBackstab/DoCircle convention).
//   - luck-derived percent modifier (C uses get_curr_lck adjustments) — Go
//     skills uniformly drop this.
func DoPounce(ch *types.CharData, argument string) {
	// NPC + AFF_CHARM early-return (C :2614-2618).
	if ch.IsNPC() && ch.AffectedBy.IsSet(types.AFF_CHARM) {
		ch.Send("You can't do that right now.\n\r")
		return
	}

	arg, _ := util.OneArgument(argument)

	// Mount gate (C :2622-2626). Checked before arg test, matching C source
	// order — if mounted, report mount issue even if no arg supplied.
	if ch.Mount != nil {
		ch.Send("You can't get close enough while mounted.\n\r")
		return
	}

	if arg == "" {
		ch.Send("Pounce on whom?\n\r")
		return
	}

	victim := handler.GetCharRoom(ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}
	if victim == ch {
		ch.Send("Pounce on yourself?\n\r")
		return
	}

	// Weapon-type gate (C :2649-2660). Value[3] is the damage-type index.
	wield := handler.GetEqChar(ch, types.WEAR_WIELD)
	if wield == nil ||
		(wield.Value[3] != types.DAM_SLICE &&
			wield.Value[3] != types.DAM_STAB &&
			wield.Value[3] != types.DAM_SLASH &&
			wield.Value[3] != types.DAM_CLAW &&
			wield.Value[3] != types.DAM_BITE &&
			wield.Value[3] != types.DAM_PIERCE) {
		ch.Send("You are not wielding an appropriate weapon type to effectively pounce.\n\r")
		return
	}

	// Victim-already-fighting gate (C :2662-2666).
	if victim.Fighting != nil {
		ch.Send("You cannot pounce on someone who is in combat.\n\r")
		return
	}

	// Hurt-and-awake gate (C :2668-2673). IS_AWAKE == Position > POS_SLEEPING.
	if victim.Hit < victim.MaxHit && victim.Position > types.POS_SLEEPING {
		util.Act(types.AT_PLAIN, "$N is hurt and suspicious ... you can't sneak up.",
			ch, victim, nil, nil, types.TO_CHAR)
		return
	}

	// WAIT_STATE uses skill_table[gsn_pounce]->beats (= 12 per skills.dat).
	gsn := lookupSkillSlot("pounce")
	if WorldRef != nil && gsn >= 0 && gsn < len(WorldRef.Skills) && WorldRef.Skills[gsn] != nil {
		ch.Wait = WorldRef.Skills[gsn].Beats
	}

	// Success gate (C :2682): succeed if victim is asleep (auto-success) OR
	// the skill check passes.
	if victim.Position <= types.POS_SLEEPING || canUseSkill(ch, numberPercent(), gsn) {
		learnFromSuccess(ch, gsn)
		combat.MultiHit(WorldRef, ch, victim, gsn)
	} else {
		learnFromFailure(ch, gsn)
		combat.Damage(WorldRef, ch, victim, 0, gsn)
	}
}

// DoGrapple — lock the target in a clinch. src/skills.c:1690. MVP: drop
// the PKill-only and NPC-restrictions (C requires IS_PKILL(ch) and
// !IS_NPC(victim)); allow anyone to grapple anyone in the same room. If
// successful, drop both parties' DEX by 2 and put them AFF_GRAPPLE.
func DoGrapple(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	var victim *types.CharData
	if ch.Fighting != nil {
		victim = ch.Fighting.Who
	} else if arg != "" {
		victim = handler.GetCharRoom(ch, arg)
		if victim == nil {
			ch.Send("They aren't here.\n\r")
			return
		}
	} else {
		ch.Send("Grapple whom?\n\r")
		return
	}
	if victim == ch {
		ch.Send("How can you sneak up on yourself?\n\r")
		return
	}
	if victim.AffectedBy.IsSet(types.AFF_GRAPPLE) {
		ch.Send("They are already grappled.\n\r")
		return
	}
	gsn := lookupSkillSlot("grapple")
	if !canUseSkill(ch, util.NumberPercent(), gsn) {
		ch.Send("You lost your balance.\n\r")
		learnFromFailure(ch, gsn)
		return
	}
	learnFromSuccess(ch, gsn)
	newAff := func() *types.AffectData {
		a := &types.AffectData{
			Type:     gsn,
			Duration: 2,
			Location: types.APPLY_DEX,
			Modifier: -2,
		}
		a.BitVector.Set(types.AFF_GRAPPLE)
		return a
	}
	handler.AffectToChar(victim, newAff())
	handler.AffectToChar(ch, newAff())
	ch.Sendf("You manage to grab hold of %s!\n\r", util.Capitalize(victim.Name))
	util.Act(types.AT_ACTION, "$n grabs hold of you!", ch, victim, nil, nil, types.TO_VICT)
	util.Act(types.AT_ACTION, "$n begins grappling with $N!", ch, victim, nil, nil, types.TO_NOTVICT)
	if ch.Fighting == nil && victim.InRoom == ch.InRoom {
		combat.StartFighting(ch, victim)
	}
	if victim.Fighting == nil && ch.InRoom == victim.InRoom {
		combat.StartFighting(victim, ch)
	}
}

// DoCleave — heavy-weapon slash. src/skills.c:3831. MVP: no weapon-type
// check (C requires slashing weapon); damage formula uses str-based scaling
// like C with a 20% crit chance for the "devastating blow" branch.
func DoCleave(ch *types.CharData, argument string) {
	if ch.Fighting == nil {
		ch.Send("You aren't fighting anyone.\n\r")
		return
	}
	victim := ch.Fighting.Who
	if handler.GetEqChar(ch, types.WEAR_WIELD) == nil {
		ch.Send("You need a slashing weapon.\n\r")
		return
	}
	gsn := lookupSkillSlot("cleave")
	if canUseSkill(ch, util.NumberPercent(), gsn) {
		var dam int
		if util.NumberPercent() <= 20 {
			ch.Send("You deal a devastating blow!\n\r")
			dam = util.NumberRange(11, 22)*ch.GetCurrStr() + 30
		} else {
			dam = util.NumberRange(9, 18)*ch.GetCurrStr() + 30
		}
		learnFromSuccess(ch, gsn)
		combat.Damage(WorldRef, ch, victim, dam, gsn)
	} else {
		learnFromFailure(ch, gsn)
		combat.Damage(WorldRef, ch, victim, 0, gsn)
	}
}

// DoHitall — attack every valid target in the room. src/skills.c:5291.
// MVP: no ROOM_SAFE check; cap targets at max(1, level/5). Group members are
// skipped per C skills.c:5314 (`is_same_group` — we approximate by comparing
// group leaders and direct master relationships).
func DoHitall(ch *types.CharData, argument string) {
	if ch.InRoom == nil || len(ch.InRoom.People) <= 1 {
		ch.Send("There's no one else here!\n\r")
		return
	}
	gsn := lookupSkillSlot("hitall")
	maxHits := util.UMAX(1, ch.Level/5)
	chLeader := ch.Leader
	if chLeader == nil {
		chLeader = ch
	}
	sameGroup := func(vch *types.CharData) bool {
		if vch == ch {
			return true
		}
		vLeader := vch.Leader
		if vLeader == nil {
			vLeader = vch
		}
		return vLeader == chLeader
	}
	// Copy the People slice — combat.Damage may mutate it.
	victims := make([]*types.CharData, 0, len(ch.InRoom.People))
	for _, vch := range ch.InRoom.People {
		if !sameGroup(vch) {
			victims = append(victims, vch)
		}
	}
	nvict := 0
	for _, vch := range victims {
		if vch.InRoom != ch.InRoom {
			continue // might have moved/died
		}
		if nvict >= maxHits {
			break
		}
		nvict++
		// Track the retcode so reactive damage (fireshield / ice_shield /
		// acid_shield, future work) on the attacker stops the cascade — C
		// skills.c:5334 breaks on rCHAR_DIED || rBOTH_DIED || char_died(ch).
		// We expose combat's retcode constants via the CharDiedRetcode /
		// BothDiedRetcode sentinel helpers in combat/retcode.go.
		var ret int
		if canUseSkill(ch, util.NumberPercent(), gsn) {
			ret = combat.OneHit(WorldRef, ch, vch, types.TYPE_UNDEFINED)
		} else {
			ret = combat.Damage(WorldRef, ch, vch, 0, types.TYPE_UNDEFINED)
		}
		if combat.AttackerDied(ret) || ch.Position <= types.POS_DEAD {
			break
		}
	}
	if nvict > 0 {
		learnFromSuccess(ch, gsn)
	}
}

// DoBerserk — self-rage buff. src/skills.c:5247.
func DoBerserk(ch *types.CharData, argument string) {
	if ch.Fighting == nil {
		ch.Send("But you aren't fighting!\n\r")
		return
	}
	if ch.AffectedBy.IsSet(types.AFF_BERSERK) {
		ch.Send("Your rage is already at its peak!\n\r")
		return
	}
	gsn := lookupSkillSlot("berserk")
	if !canUseSkill(ch, util.NumberPercent(), gsn) {
		ch.Send("You couldn't build up enough rage.\n\r")
		learnFromFailure(ch, gsn)
		return
	}
	aff := &types.AffectData{
		Type:     gsn,
		Duration: util.NumberRange(util.UMAX(1, ch.Level/5), util.UMAX(2, ch.Level*2/5)),
		Location: types.APPLY_STR,
		Modifier: 1,
	}
	aff.BitVector.Set(types.AFF_BERSERK)
	handler.AffectToChar(ch, aff)
	ch.Send("You start to lose control..\n\r")
	learnFromSuccess(ch, gsn)
}
