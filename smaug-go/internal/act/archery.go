package act

import (
	"fmt"

	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// Archery subsystem — Phase 6 port of src/archery.c.
//
// Public commands: DoDraw / DoFire / DoDislodge.
// Helpers: FindQuiver, FindProjectile.
// Internal hit core: projectileHit, rangedAttack, scanForVictim,
// rangedGotTarget.
//
// Plan: smaug-go/doc/plan-phase6-archery.md. TDD pin tests live in
// archery_test.go. Three C bugs preserved verbatim (plan §Open Q):
//   - do_draw WEAR_DUAL_WIELD omission in hand-count (archery.c:142-149).
//   - do_dislodge arm damage 3*val[1]..2*val[2] asymmetry vs rib 3/3 / leg 2/2 (archery.c:234).
//   - do_fire structurally-dead victim==ch check on never-assigned pointer (archery.c:1255,1272).
//
// Out of scope: mob_fire (follow-up plan), quiver auto-reload (not in C),
// combat-loop ranged-attack hook (C's one_hit never reads WEAR_MISSILE_WIELD).

// FindQuiver returns the most-recently-carried visible open ITEM_QUIVER
// from ch.Carrying. Ports src/archery.c:94-107 find_quiver; C walks
// ch->last_carrying backwards — in Go we iterate the Carrying slice from
// tail to head for equivalent semantics.
func FindQuiver(ch *types.CharData) *types.ObjData {
	for i := len(ch.Carrying) - 1; i >= 0; i-- {
		obj := ch.Carrying[i]
		if !archeryCanSeeObj(ch, obj) {
			continue
		}
		if obj.ItemType == types.ITEM_QUIVER && int(types.CONT_CLOSED)&obj.Value[1] == 0 {
			return obj
		}
	}
	return nil
}

// FindProjectile returns the most-recently-added visible ITEM_PROJECTILE
// inside quiver. Ports src/archery.c:109-122 find_projectile.
func FindProjectile(ch *types.CharData, quiver *types.ObjData) *types.ObjData {
	if quiver == nil {
		return nil
	}
	for i := len(quiver.Contents) - 1; i >= 0; i-- {
		obj := quiver.Contents[i]
		if !archeryCanSeeObj(ch, obj) {
			continue
		}
		if obj.ItemType == types.ITEM_PROJECTILE {
			return obj
		}
	}
	return nil
}

// archeryCanSeeObj is a reduced visibility check for the archery
// helpers. Mirrors util.actCanSeeObj (private): ITEM_INVIS hidden unless
// observer has AFF_DETECT_INVIS. All other ch.Carrying objects are
// visible to the carrier.
func archeryCanSeeObj(ch *types.CharData, obj *types.ObjData) bool {
	if ch == nil || obj == nil {
		return true
	}
	if obj.ExtraFlags.IsSet(types.ITEM_INVIS) &&
		!ch.AffectedBy.IsSet(types.AFF_DETECT_INVIS) {
		return false
	}
	return true
}

// archeryCanSee is the char-visibility counterpart. Mirrors util.actCanSee
// — AFF_INVISIBLE hidden unless observer has AFF_DETECT_INVIS.
func archeryCanSee(observer, target *types.CharData) bool {
	if observer == nil || target == nil {
		return true
	}
	if observer == target {
		return true
	}
	if target.AffectedBy.IsSet(types.AFF_INVISIBLE) &&
		!observer.AffectedBy.IsSet(types.AFF_DETECT_INVIS) {
		return false
	}
	return true
}

// SeparateObj is the Go stand-in for C's separate_obj (handler.c) which
// splits a stacked-object group into a singleton before in-place
// mutations. The Go port does not currently implement object stacking;
// SeparateObj is a no-op. Kept as a named call site so future stacking
// work has an anchor.
//
// TODO: when object stacking lands, implement the split semantics here.
func SeparateObj(obj *types.ObjData) {
	// Intentionally empty — no stacking in Go port.
	_ = obj
}

// --- G3: DoDraw ---------------------------------------------------------

// DoDraw moves a projectile from a worn quiver into WEAR_HOLD. Ports
// src/archery.c:125-200. C bug preserved: WEAR_DUAL_WIELD is NOT counted
// in the hand-count check (:142-149).
func DoDraw(ch *types.CharData, argument string) {
	_ = argument // C ignores the argument entirely.

	bow := handler.GetEqChar(ch, types.WEAR_MISSILE_WIELD)
	if bow == nil {
		ch.Send("You are not wielding a missile weapon!\n\r")
		return
	}

	quiver := FindQuiver(ch)
	if quiver == nil {
		ch.Send("You aren't wearing a quiver where you can get to it!\n\r")
		return
	}

	handCount := 0
	if handler.GetEqChar(ch, types.WEAR_LIGHT) != nil {
		handCount++
	}
	if handler.GetEqChar(ch, types.WEAR_SHIELD) != nil {
		handCount++
	}
	if handler.GetEqChar(ch, types.WEAR_HOLD) != nil {
		handCount++
	}
	if handler.GetEqChar(ch, types.WEAR_WIELD) != nil {
		handCount++
	}
	// C BUG preserved: WEAR_DUAL_WIELD is deliberately absent from this
	// count (src/archery.c:142-149). See plan Open-Q summary.
	if handCount > 1 {
		ch.Send("You need a free hand to draw with.\n\r")
		return
	}

	// Redundant with handCount when HOLD is the one occupied slot, but
	// port-verbatim — this is the second gate in C.
	if handler.GetEqChar(ch, types.WEAR_HOLD) != nil {
		ch.Send("Your hand is not empty!\n\r")
		return
	}

	arrow := FindProjectile(ch, quiver)
	if arrow == nil {
		ch.Send("Your quiver is empty!!\n\r")
		return
	}

	SeparateObj(arrow)

	// Ammo-type gate: bow.Value[5] (accepted PROJ_* kind) must equal
	// arrow.Value[4] (arrow's ammo type). C :176-182.
	if bow.Value[5] != arrow.Value[4] {
		ch.Send("You drew the wrong projectile type for this weapon!\n\r")
		// C pulls the arrow from the quiver via obj_from_obj then
		// obj_to_char — but in our flow the arrow was never detached
		// (separate_obj is a no-op in Go). Match C's observable
		// behaviour by ensuring the arrow stays inside the quiver.
		if arrow.InObj != quiver {
			handler.ObjFromChar(arrow)
			handler.ObjToObj(arrow, quiver)
		}
		return
	}

	ch.Wait = types.PULSE_VIOLENCE

	drawRoomMsg := fmt.Sprintf("$n draws %s from $p.", arrow.ShortDescr)
	drawCharMsg := fmt.Sprintf("You draw %s from $p.", arrow.ShortDescr)
	util.Act(types.AT_ACTION, drawRoomMsg, ch, nil, quiver, nil, types.TO_ROOM)
	util.Act(types.AT_ACTION, drawCharMsg, ch, nil, quiver, nil, types.TO_CHAR)

	handler.ObjFromObj(arrow)
	handler.ObjToChar(arrow, ch)

	// Equip at WEAR_HOLD. C calls wear_obj(ch, arrow, TRUE, -1) which
	// autodetects via arrow's wear-flags; projectile arrows are expected
	// to carry ITEM_HOLD. In Go we directly slot WEAR_HOLD to keep the
	// port simple and deterministic — arrows are not expected to have
	// multi-slot wear flags.
	arrow.WearLoc = types.WEAR_HOLD
}

// --- G4: DoDislodge ------------------------------------------------------

// DoDislodge removes a lodged arrow from the actor's body, dealing
// self-damage. Ports src/archery.c:203-255. Rib/arm/leg scanned in order;
// first non-nil wins. Three C formulae preserved verbatim, including
// the arm's 3*val[1]..2*val[2] asymmetry (:234 — rib is 3/3, leg is 2/2).
func DoDislodge(ch *types.CharData, argument string) {
	if argument == "" {
		ch.Send("Dislodge what?\n\r")
		return
	}
	// C quirk: argument is consumed only for the empty-check; actual
	// arrow located by slot scan, not by name match. Port-verbatim.

	if arrow := handler.GetEqChar(ch, types.WEAR_LODGE_RIB); arrow != nil {
		util.Act(types.AT_CARNAGE, "With a wrenching pull, you dislodge $p from your chest.",
			ch, nil, arrow, nil, types.TO_CHAR)
		util.Act(types.AT_CARNAGE, "$N winces in pain as $e dislodges $p from $s chest.",
			ch, nil, arrow, nil, types.TO_ROOM)
		handler.UnequipChar(ch, arrow)
		arrow.WearFlags &^= int(types.ITEM_LODGE_RIB)
		arrow.ExtraFlags.Remove(types.ITEM_LODGED)
		dam := util.NumberRange(3*arrow.Value[1], 3*arrow.Value[2])
		if WorldRef != nil {
			combat.Damage(WorldRef, ch, ch, dam, types.TYPE_UNDEFINED)
		} else {
			ch.Hit -= dam
		}
		return
	}
	if arrow := handler.GetEqChar(ch, types.WEAR_LODGE_ARM); arrow != nil {
		util.Act(types.AT_CARNAGE, "With a tug you dislodge $p from your arm.",
			ch, nil, arrow, nil, types.TO_CHAR)
		util.Act(types.AT_CARNAGE, "$N winces in pain as $e dislodges $p from $s arm.",
			ch, nil, arrow, nil, types.TO_ROOM)
		handler.UnequipChar(ch, arrow)
		arrow.WearFlags &^= int(types.ITEM_LODGE_ARM)
		arrow.ExtraFlags.Remove(types.ITEM_LODGED)
		// C BUG preserved: arm uses 3*val[1]..2*val[2] (rib: 3/3, leg: 2/2).
		// src/archery.c:234. See plan Open-Q summary.
		dam := util.NumberRange(3*arrow.Value[1], 2*arrow.Value[2])
		if WorldRef != nil {
			combat.Damage(WorldRef, ch, ch, dam, types.TYPE_UNDEFINED)
		} else {
			ch.Hit -= dam
		}
		return
	}
	if arrow := handler.GetEqChar(ch, types.WEAR_LODGE_LEG); arrow != nil {
		util.Act(types.AT_CARNAGE, "With a tug you dislodge $p from your leg.",
			ch, nil, arrow, nil, types.TO_CHAR)
		util.Act(types.AT_CARNAGE, "$N winces in pain as $e dislodges $p from $s leg.",
			ch, nil, arrow, nil, types.TO_ROOM)
		handler.UnequipChar(ch, arrow)
		arrow.WearFlags &^= int(types.ITEM_LODGE_LEG)
		arrow.ExtraFlags.Remove(types.ITEM_LODGED)
		dam := util.NumberRange(2*arrow.Value[1], 2*arrow.Value[2])
		if WorldRef != nil {
			combat.Damage(WorldRef, ch, ch, dam, types.TYPE_UNDEFINED)
		} else {
			ch.Hit -= dam
		}
		return
	}
	ch.Send("You have nothing lodged in your body.\n\r")
}

// --- G5: DoFire, rangedAttack, scanForVictim, rangedGotTarget, projectileHit ---

// rc_* return codes — internal to archery. Match combat.rNONE family.
const (
	rcNONE      = 0
	rcCHAR_DIED = 1
	rcVICT_DIED = 2
)

// maxFight mirrors the C helper max_fight at src/fight.c:281. In stock C
// it returns a level-scaled or flat cap — since Go doesn't currently
// have the helper, use a reasonable default of 3 (matches SMAUG's
// unmodified MAX_FIGHT). Extend when max_fight formally ports.
//
// TODO: port full max_fight from src/fight.c during combat-depth follow-up.
func maxFight(ch *types.CharData) int {
	_ = ch
	return 3
}

// isSafe is a simplified port of src/fight.c is_safe — returns true if
// ch cannot attack vch. The Go port does not currently have a full
// is_safe; provide a minimal gate covering the cases the archery path
// cares about: ROOM_SAFE on the victim's room, attacker == victim.
//
// TODO: replace with a shared IsSafe helper when that ports.
func isSafe(ch, vch *types.CharData, checkFriendly bool) bool {
	_ = checkFriendly
	if ch == nil || vch == nil {
		return true
	}
	if ch == vch {
		return true
	}
	if vch.InRoom != nil && vch.InRoom.RoomFlags.IsSet(types.ROOM_SAFE) {
		return true
	}
	return false
}

// DoFire is the archery fire command. Ports src/archery.c:1252-1329.
// Replaces the Tier-5 MVP throw-passthrough.
//
// C bug preserved verbatim: `victim` is declared and initialized NULL
// at archery.c:1255, and the `victim == ch` branch at :1272 is
// structurally dead (never fires). Ported as-is — if "none"/"self" arg
// is given, the message fires; the victim==ch branch is effectively a
// no-op but kept for fidelity.
func DoFire(ch *types.CharData, argument string) {
	bow := handler.GetEqChar(ch, types.WEAR_MISSILE_WIELD)
	if bow == nil {
		ch.Send("But you are not wielding a missile weapon!!\n\r")
		return
	}

	arg, _ := util.OneArgument(argument)
	if arg == "" && ch.Fighting == nil {
		ch.Send("Fire at whom or what?\n\r")
		return
	}

	// C BUG preserved: victim is never assigned before this check
	// (src/archery.c:1255,1272). The `victim == ch` clause is dead; the
	// string-matched arm handles "none"/"self" correctly.
	var victim *types.CharData = nil
	if arg == "none" || arg == "self" || victim == ch {
		ch.Send("How exactly did you plan on firing at yourself?\n\r")
		return
	}

	arrow := handler.GetEqChar(ch, types.WEAR_HOLD)
	if arrow == nil {
		ch.Send("You are not holding a projectile!\n\r")
		return
	}
	if arrow.ItemType != types.ITEM_PROJECTILE {
		ch.Send("You are not holding a projectile!\n\r")
		return
	}

	// Max-dist clamped 1..10 (URANGE).
	maxDist := bow.Value[4]
	if maxDist < 1 {
		maxDist = 1
	}
	if maxDist > 10 {
		maxDist = 10
	}

	// Ammo gate. src/archery.c:1295 uses bow.Value[5] != arrow.Value[4].
	// (mob_fire at :1352 uses [4]!=[5] — asymmetry noted in plan Q12;
	// mob_fire is out of scope for this plan.)
	if bow.Value[5] != arrow.Value[4] {
		var msg string
		switch bow.Value[5] {
		case types.PROJ_BOLT:
			msg = "You have no bolts...\n\r"
		case types.PROJ_ARROW:
			msg = "You have no arrows...\n\r"
		case types.PROJ_DART:
			msg = "You have no darts...\n\r"
		case types.PROJ_STONE:
			msg = "You have no slingstones...\n\r"
		default:
			msg = "You have nothing to fire...\n\r"
		}
		ch.Send(msg)
		return
	}

	ch.Wait = 6 // wait-state; C uses 6 explicitly, not PULSE_VIOLENCE.

	rangedAttack(ch, argument, bow, arrow, types.TYPE_HIT+arrow.Value[3], maxDist)
}

// rangedAttack ports src/archery.c:873-1249 ranged_attack.
func rangedAttack(ch *types.CharData, argument string, weapon, projectile *types.ObjData, dt, rng int) int {
	arg, rest := util.OneArgument(argument)
	arg1, _ := util.OneArgument(rest)

	if arg == "" {
		ch.Send("Where?  At who?\n\r")
		return rcNONE
	}

	var victim *types.CharData
	var pexit *types.ExitData
	dir := -1

	pexit = findDoor(ch, arg)
	if pexit == nil {
		victim = handler.GetCharRoom(ch, arg)
		if victim == nil {
			ch.Send("Aim in what direction?\n\r")
			return rcNONE
		}
		if ch.Fighting != nil && ch.Fighting.Who == victim {
			ch.Send("They are too close to release that type of attack!\n\r")
			return rcNONE
		}
	} else {
		dir = pexit.Direction
	}

	if victim == nil && ch.InRoom != nil {
		if ch.InRoom.RoomFlags.IsSet(types.ROOM_PRIVATE) ||
			ch.InRoom.RoomFlags.IsSet(types.ROOM_SOLITARY) {
			ch.Send("You cannot perform a ranged attack from a private room.\n\r")
			return rcNONE
		}
		if ch.InRoom.Tunnel > 0 && len(ch.InRoom.People) >= ch.InRoom.Tunnel {
			ch.Send("This room is too cramped to perform such an attack.\n\r")
			return rcNONE
		}
	}

	if pexit != nil && pexit.ToRoom == nil {
		ch.Send("Are you expecting to fire through a wall!?\n\r")
		return rcNONE
	}

	if pexit != nil && pexit.ExitInfo&int(types.EX_CLOSED) != 0 {
		if pexit.ExitInfo&int(types.EX_SECRET) != 0 {
			ch.Send("Are you expecting to fire through a wall!?\n\r")
		} else {
			ch.Send("Are you expecting to fire through a door!?\n\r")
		}
		return rcNONE
	}

	var vch *types.CharData
	if pexit != nil && arg1 != "" {
		vch = scanForVictim(ch, pexit, arg1)
		if vch == nil {
			ch.Send("You cannot see your target.\n\r")
			return rcNONE
		}
		if vch.InRoom != nil && vch.InRoom.RoomFlags.IsSet(types.ROOM_NOMISSILE) {
			ch.Send("You can't get a clean shot off.\n\r")
			return rcNONE
		}
		if vch.NumFighting > maxFight(vch) {
			ch.Send("There is too much activity there for you to get a clear shot.\n\r")
			return rcNONE
		}
	}
	if vch != nil {
		if !vch.IsNPC() && !ch.IsNPC() && ch.Act.IsSet(types.PLR_NICE) {
			ch.Send("Your too nice to do that!\n\r")
			return rcNONE
		}
		if isSafe(ch, vch, true) {
			return rcNONE
		}
	}

	wasInRoom := ch.InRoom

	if projectile != nil {
		SeparateObj(projectile)
		if pexit != nil {
			if weapon != nil {
				util.Act(types.AT_ACTION, "You fire $p $T.", ch, nil, projectile, dirName(dir), types.TO_CHAR)
				util.Act(types.AT_ACTION, "$n fires $p $T.", ch, nil, projectile, dirName(dir), types.TO_ROOM)
			} else {
				util.Act(types.AT_ACTION, "You throw $p $T.", ch, nil, projectile, dirName(dir), types.TO_CHAR)
				util.Act(types.AT_ACTION, "$n throw $p $T.", ch, nil, projectile, dirName(dir), types.TO_ROOM)
			}
		} else {
			if weapon != nil {
				util.Act(types.AT_ACTION, "You fire $p at $N.", ch, victim, projectile, nil, types.TO_CHAR)
				util.Act(types.AT_ACTION, "$n fires $p at $N.", ch, victim, projectile, nil, types.TO_NOTVICT)
				util.Act(types.AT_ACTION, "$n fires $p at you!", ch, victim, projectile, nil, types.TO_VICT)
			} else {
				util.Act(types.AT_ACTION, "You throw $p at $N.", ch, victim, projectile, nil, types.TO_CHAR)
				util.Act(types.AT_ACTION, "$n throws $p at $N.", ch, victim, projectile, nil, types.TO_NOTVICT)
				util.Act(types.AT_ACTION, "$n throws $p at you!", ch, victim, projectile, nil, types.TO_VICT)
			}
		}
	}

	// Same-room victim — immediate resolution.
	if victim != nil {
		return rangedGotTarget(ch, victim, weapon, projectile, 0, dt)
	}

	// Exit-chain walk — projectile flies room-to-room.
	if pexit == nil {
		return rcNONE
	}
	dist := 0
	for dist <= rng {
		handler.CharFromRoom(ch)
		handler.CharToRoom(ch, pexit.ToRoom)

		if pexit.ExitInfo&int(types.EX_CLOSED) != 0 {
			if projectile != nil {
				util.Act(types.AT_ACTION, "$p strikes a door in the distance.",
					ch, nil, projectile, nil, types.TO_CHAR)
			}
			break
		}

		if vch == nil && ch.InRoom != nil {
			for _, candidate := range ch.InRoom.People {
				if candidate == ch {
					continue
				}
				npcCh := ch.IsNPC()
				npcVch := candidate.IsNPC()
				if ((npcCh && !npcVch) || (!npcCh && npcVch)) && util.NumberBits(1) == 0 {
					vch = candidate
					break
				}
			}
			if vch != nil && isSafe(ch, vch, false) {
				handler.CharFromRoom(ch)
				handler.CharToRoom(ch, wasInRoom)
				return rcNONE
			}
		}

		if vch != nil && ch.InRoom == vch.InRoom {
			handler.CharFromRoom(ch)
			handler.CharToRoom(ch, wasInRoom)
			return rangedGotTarget(ch, vch, weapon, projectile, dist, dt)
		}

		if dist == rng {
			if projectile != nil {
				util.Act(types.AT_ACTION, "$p flies in and falls harmlessly to the ground here.",
					ch, nil, projectile, nil, types.TO_ROOM)
				if projectile.InObj != nil {
					handler.ObjFromObj(projectile)
				}
				if projectile.CarriedBy != nil {
					handler.ObjFromChar(projectile)
				}
				if ch.InRoom != nil {
					handler.ObjToRoom(projectile, ch.InRoom)
				}
			}
			break
		}

		next := ch.InRoom.GetExit(dir)
		if next == nil {
			if projectile != nil {
				util.Act(types.AT_ACTION, "$p strikes the wall and falls harmlessly to the ground.",
					ch, nil, projectile, nil, types.TO_ROOM)
				if projectile.InObj != nil {
					handler.ObjFromObj(projectile)
				}
				if projectile.CarriedBy != nil {
					handler.ObjFromChar(projectile)
				}
				if ch.InRoom != nil {
					handler.ObjToRoom(projectile, ch.InRoom)
				}
			}
			break
		}
		pexit = next
		dist++
	}

	handler.CharFromRoom(ch)
	handler.CharToRoom(ch, wasInRoom)
	return rcNONE
}

// rangedGotTarget is the projectile-vs-victim handoff. Ports
// src/archery.c:640-783 ranged_got_target.
func rangedGotTarget(ch, victim *types.CharData, weapon, projectile *types.ObjData, dist, dt int) int {
	if ch.InRoom != nil && ch.InRoom.RoomFlags.IsSet(types.ROOM_SAFE) {
		if projectile != nil && WorldRef != nil {
			util.Act(types.AT_ACTION, "A godly presence smites $p!", ch, nil, projectile, nil, types.TO_ROOM)
			handler.ExtractObj(WorldRef, projectile)
		}
		return rcNONE
	}

	// Skill check gate. C at :735 uses number_percent() > 50 || can_use_skill;
	// Go collapses all archery gsns to gsnMissileWeapons. Simplification:
	// always attempt the shot when seamed percent passes.
	if util.NumberPercent() > 50 {
		return projectileHit(ch, victim, weapon, projectile, dist, dt)
	}

	// Fail — no hit, 50/50 extract/drop.
	if projectile != nil {
		if util.NumberPercent() < 50 {
			if WorldRef != nil {
				handler.ExtractObj(WorldRef, projectile)
			}
		} else {
			if projectile.CarriedBy != nil {
				handler.ObjFromChar(projectile)
			}
			if victim.InRoom != nil {
				handler.ObjToRoom(projectile, victim.InRoom)
			}
		}
	}
	if WorldRef != nil {
		combat.Damage(WorldRef, ch, victim, 0, dt)
	}
	return rcNONE
}

// projectileHit ports src/archery.c:262-635 projectile_hit.
//
// Simplifications vs C (Go port scope):
//   - THAC0 for PCs uses `20 - ch.Level` (matches internal/combat/combat.go).
//   - RIS bitmap check (C :435-493) is omitted — Go has no public ris_damage.
//   - Weapon-spell iteration (C :608-629) is omitted.
//   - `learn_from_failure` on miss is omitted.
func projectileHit(ch, victim *types.CharData, wield, projectile *types.ObjData, dist, dt int) int {
	if projectile == nil {
		return rcNONE
	}

	projBonus := 0
	isProjOrWeap := projectile.ItemType == types.ITEM_PROJECTILE ||
		projectile.ItemType == types.ITEM_WEAPON
	if isProjOrWeap {
		dt = types.TYPE_HIT + projectile.Value[3]
		if wield != nil {
			projBonus = util.NumberRange(wield.Value[1], wield.Value[2])
		}
	} else {
		dt = types.TYPE_UNDEFINED
	}

	// Dead victim early exit.
	if victim.Position == types.POS_DEAD || victim.Hit <= 0 {
		if WorldRef != nil {
			handler.ExtractObj(WorldRef, projectile)
		}
		return rcVICT_DIED
	}

	profBonus, _ := combat.WeaponProfBonusCheck(ch, wield)

	if dt == types.TYPE_UNDEFINED {
		dt = types.TYPE_HIT
		if wield != nil && wield.ItemType == types.ITEM_MISSILE_WEAPON {
			dt += wield.Value[3]
		}
	}

	var thac0 int
	if ch.IsNPC() {
		thac0 = ch.MobThac0
	} else {
		thac0 = 20 - ch.Level
	}
	thac0 = thac0 - ch.Hitroll + (dist * 2)

	victimAC := util.UMAX(-19, victim.Armor/10)
	if !archeryCanSeeObj(victim, projectile) {
		victimAC++
	}
	if !archeryCanSee(ch, victim) {
		victimAC -= 4
	}
	victimAC += profBonus

	diceroll := archeryRollD20()
	if diceroll == 0 || (diceroll != 19 && diceroll < thac0-victimAC) {
		// Miss. 50/50 extract or drop.
		if util.NumberPercent() < 50 {
			if WorldRef != nil {
				handler.ExtractObj(WorldRef, projectile)
			}
		} else {
			if projectile.CarriedBy != nil {
				handler.ObjFromChar(projectile)
			}
			if victim.InRoom != nil {
				handler.ObjToRoom(projectile, victim.InRoom)
			}
		}
		if WorldRef != nil {
			combat.Damage(WorldRef, ch, victim, 0, dt)
		}
		return rcNONE
	}

	// Hit-zone damage.
	pchance := util.NumberRange(1, 10)
	var dam int
	switch {
	case pchance <= 3:
		dam = util.NumberRange(projectile.Value[1], projectile.Value[2]) + projBonus
	case pchance <= 6:
		dam = util.NumberRange(2*projectile.Value[1], 2*projectile.Value[2]) + projBonus
	default:
		dam = util.NumberRange(3*projectile.Value[1], 3*projectile.Value[2]) + projBonus
	}
	dam += ch.Damroll
	if profBonus > 0 {
		dam += profBonus / 4
	}

	switch victim.Position {
	case types.POS_BERSERK:
		dam = int(1.2 * float64(dam))
	case types.POS_AGGRESSIVE:
		dam = int(1.1 * float64(dam))
	case types.POS_DEFENSIVE:
		dam = int(0.85 * float64(dam))
	case types.POS_EVASIVE:
		dam = int(0.8 * float64(dam))
	}
	if victim.Position <= types.POS_SLEEPING {
		dam *= 2
	}
	if dam <= 0 {
		dam = 1
	}

	var retcode int
	if WorldRef != nil {
		retcode = combat.Damage(WorldRef, ch, victim, dam, dt)
	}
	if retcode != rcNONE {
		if projectile.Value[5] == types.PROJ_STONE {
			if WorldRef != nil {
				handler.ExtractObj(WorldRef, projectile)
			}
			return retcode
		}
		if victim.Hit <= 0 {
			if WorldRef != nil {
				handler.ExtractObj(WorldRef, projectile)
			}
			return rcVICT_DIED
		}
		lodgeProjectile(projectile, victim, pchance)
		return retcode
	}
	if ch.Hit <= 0 {
		if WorldRef != nil {
			handler.ExtractObj(WorldRef, projectile)
		}
		return rcCHAR_DIED
	}
	if victim.Hit <= 0 {
		if WorldRef != nil {
			handler.ExtractObj(WorldRef, projectile)
		}
		return rcVICT_DIED
	}

	// Survived hit: lodge when damage landed and not a stone.
	if dam > 0 && projectile.Value[5] != types.PROJ_STONE {
		lodgeProjectile(projectile, victim, pchance)
		return rcNONE
	}

	// dam == 0 — 50/50 extract/drop.
	if util.NumberPercent() < 50 {
		if WorldRef != nil {
			handler.ExtractObj(WorldRef, projectile)
		}
	} else {
		if projectile.CarriedBy != nil {
			handler.ObjFromChar(projectile)
		}
		if victim.InRoom != nil {
			handler.ObjToRoom(projectile, victim.InRoom)
		}
	}
	return rcNONE
}

// lodgeProjectile equips the projectile in one of victim's WEAR_LODGE_*
// slots based on hit-zone pchance, sets ITEM_LODGED and ITEM_LODGE_*.
// Ports src/archery.c:550-575.
func lodgeProjectile(projectile *types.ObjData, victim *types.CharData, pchance int) {
	if projectile.CarriedBy != nil {
		handler.ObjFromChar(projectile)
	}
	if projectile.InRoom != nil {
		handler.ObjFromRoom(projectile)
	}
	handler.ObjToChar(projectile, victim)
	projectile.ExtraFlags.Set(types.ITEM_LODGED)
	switch {
	case pchance <= 3:
		projectile.WearFlags |= int(types.ITEM_LODGE_ARM)
		projectile.WearLoc = types.WEAR_LODGE_ARM
	case pchance <= 6:
		projectile.WearFlags |= int(types.ITEM_LODGE_LEG)
		projectile.WearLoc = types.WEAR_LODGE_LEG
	default:
		projectile.WearFlags |= int(types.ITEM_LODGE_RIB)
		projectile.WearLoc = types.WEAR_LODGE_RIB
	}
}

// scanForVictim walks an exit chain looking for a named character.
// Ports src/archery.c:789-868.
func scanForVictim(ch *types.CharData, pexit *types.ExitData, name string) *types.CharData {
	if ch == nil || pexit == nil {
		return nil
	}
	if ch.AffectedBy.IsSet(types.AFF_BLIND) {
		return nil
	}
	wasInRoom := ch.InRoom
	maxDist := 8
	if ch.Level < 50 {
		maxDist--
	}
	if ch.Level < 40 {
		maxDist--
	}
	if ch.Level < 30 {
		maxDist--
	}

	dist := 1
	cur := pexit
	for dist <= maxDist {
		if cur == nil || cur.ToRoom == nil {
			break
		}
		if cur.ExitInfo&int(types.EX_CLOSED) != 0 {
			break
		}

		handler.CharFromRoom(ch)
		handler.CharToRoom(ch, cur.ToRoom)

		if found := handler.GetCharRoom(ch, name); found != nil && found != ch {
			handler.CharFromRoom(ch)
			handler.CharToRoom(ch, wasInRoom)
			return found
		}

		if ch.InRoom != nil {
			switch ch.InRoom.SectorType {
			case types.SECT_AIR:
				if util.NumberPercent() < 80 {
					dist++
				}
			case types.SECT_INSIDE, types.SECT_FIELD, types.SECT_UNDERGROUND:
				dist++
			case types.SECT_FOREST, types.SECT_CITY, types.SECT_DESERT, types.SECT_HILLS:
				dist += 2
			case types.SECT_WATER_SWIM, types.SECT_WATER_NOSWIM:
				dist += 3
			case types.SECT_MOUNTAIN, types.SECT_UNDERWATER, types.SECT_OCEANFLOOR:
				dist += 4
			default:
				dist++
			}
		} else {
			dist++
		}
		if dist >= maxDist {
			break
		}

		cur = ch.InRoom.GetExit(cur.Direction)
	}

	handler.CharFromRoom(ch)
	handler.CharToRoom(ch, wasInRoom)
	return nil
}

// archeryRollD20 is a function-variable seam matching combat.rollD20.
// Tests stub this to deterministically exercise miss/hit paths.
var archeryRollD20 = func() int {
	for {
		v := util.NumberBits(5)
		if v < 20 {
			return v
		}
	}
}

// dirName maps direction index to name. -1 or out-of-range → "somewhere".
func dirName(dir int) string {
	names := []string{"north", "east", "south", "west", "up", "down",
		"northeast", "northwest", "southeast", "southwest"}
	if dir < 0 || dir >= len(names) {
		return "somewhere"
	}
	return names[dir]
}
