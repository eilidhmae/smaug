// Package combat implements the SMAUG combat system.
package combat

import (
	"fmt"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// Return codes from damage/combat functions.
const (
	rNONE      = 0
	rVICT_DIED = 2
	rCHAR_DIED = 1
	rBOTH_DIED = 3
)

// AttackerDied returns true if the given retcode indicates the attacker
// (ch) died during the call — rCHAR_DIED or rBOTH_DIED. Used by skills
// that cascade OneHit/Damage over multiple victims (DoHitall) so the
// attacker's death (e.g. from reactive damage like fireshield) stops
// the cascade immediately. C-equivalent: skills.c:5334
// `global_retcode == rCHAR_DIED || global_retcode == rBOTH_DIED`.
func AttackerDied(retcode int) bool {
	return retcode == rCHAR_DIED || retcode == rBOTH_DIED
}

// VictimDied returns true if the retcode indicates the victim (vch)
// died during the call — rVICT_DIED or rBOTH_DIED. Sibling helper to
// AttackerDied; exported for the same cross-package reason.
func VictimDied(retcode int) bool {
	return retcode == rVICT_DIED || retcode == rBOTH_DIED
}

// Hook variables populated at boot to avoid an import cycle with mudprog.
// When nil, combat silently skips the trigger fire.
var (
	// HitprcntHook is called after HP has been reduced on an NPC victim to
	// allow MPROG_HITPRCNT progs to fire. Signature: (mob, attacker).
	HitprcntHook func(*types.CharData, *types.CharData)
	// VoidHook is called on a room whenever a PC leaves (via death/flee).
	VoidHook func(*types.RoomIndexData)
	// ObjDamageHook is called on each worn obj on the victim whenever damage
	// lands, so obj-prog MPROG_DAMAGE progs can fire. Signature: (attacker, obj).
	ObjDamageHook func(*types.CharData, *types.ObjData)
	// RfightHook is called on the initial StartFighting transition so a room
	// can fire its RFIGHT progs. Signature: (ch).
	RfightHook func(*types.CharData)
	// DeathRoomHook is called when a victim dies so the room can fire its
	// RDEATH progs before the corpse is generated. Signature: (victim).
	DeathRoomHook func(*types.CharData)
)

// StartFighting initiates combat between ch and victim.
func StartFighting(ch *types.CharData, victim *types.CharData) {
	if ch.Fighting != nil {
		return
	}
	ch.Fighting = &types.FightData{
		Who: victim,
	}
	ch.NumFighting = 1
	ch.Position = types.POS_FIGHTING
	// Room-prog RFIGHT trigger: fires once when combat begins in the room.
	if RfightHook != nil {
		RfightHook(ch)
	}
}

// StopFighting ends combat for a character. If fBoth is true, also stops
// anyone fighting ch.
func StopFighting(ch *types.CharData, fBoth bool) {
	ch.Fighting = nil
	ch.NumFighting = 0
	if ch.Position == types.POS_FIGHTING {
		ch.Position = types.POS_STANDING
	}

	if fBoth {
		// Find anyone fighting ch and stop them too
		if ch.InRoom != nil {
			for _, rch := range ch.InRoom.People {
				if rch.Fighting != nil && rch.Fighting.Who == ch {
					rch.Fighting = nil
					rch.NumFighting = 0
					if rch.Position == types.POS_FIGHTING {
						rch.Position = types.POS_STANDING
					}
				}
			}
		}
	}

	// After combat ends, a room that ended up with only NPCs (e.g., the PC
	// fled or was extracted) should see their VOID progs fire.
	if VoidHook != nil && ch.InRoom != nil {
		VoidHook(ch.InRoom)
	}
}

// ViolenceUpdate processes one combat round for all fighting characters.
func ViolenceUpdate(w *world.World) {
	// Per-char timer decrement. Runs every PULSE_VIOLENCE for every
	// character (not just fighters) — matches C fight.c:382-430. Timers
	// with Value == -1 are permanent and never decrement; expired
	// timers are removed by handler.DecrementTimers.
	for _, ch := range w.Characters {
		if ch == nil {
			continue
		}
		handler.DecrementTimers(ch)
	}

	for _, ch := range w.Characters {
		if ch.Fighting == nil || ch.InRoom == nil {
			continue
		}

		// Can't fight if incapacitated or worse, or at 0 HP
		if ch.Hit <= 0 || ch.Position <= types.POS_INCAP {
			StopFighting(ch, false)
			continue
		}

		victim := ch.Fighting.Who
		if victim == nil || victim.InRoom != ch.InRoom {
			StopFighting(ch, false)
			continue
		}

		// Dispatch the full attack group for this round. MultiHit handles the
		// cascade, NPC NumAttacks, dual-wield, etc.
		MultiHit(w, ch, victim, types.TYPE_UNDEFINED)

		// Wimpy auto-flee check (post-round; runs only if ch still alive and
		// still fighting someone it can see leave from).
		if ch.Fighting != nil && !ch.IsNPC() && ch.Wimpy > 0 && ch.Hit > 0 && ch.Hit <= ch.Wimpy {
			ch.Send("You wimp out and attempt to flee!\n\r")
			// Try each direction
			for dir := 0; dir <= types.DIR_DOWN; dir++ {
				if ch.InRoom == nil {
					break
				}
				exit := ch.InRoom.GetExit(dir)
				if exit != nil && exit.ToRoom != nil && exit.ExitInfo&int(types.EX_CLOSED) == 0 {
					StopFighting(ch, true)
					handler.CharFromRoom(ch)
					handler.CharToRoom(ch, exit.ToRoom)
					ch.Sendf("You flee %s!\n\r", dirName(dir))
					break
				}
			}
		}
	}
}

// numberPercent is an indirection over util.NumberPercent so tests can
// replace it with a deterministic stub. Production reads the real RNG.
// Mirrors the existing pattern in act/skills.go.
var numberPercent = util.NumberPercent

// oneHit is the indirection used by MultiHit so tests can install a spy
// (e.g. to count calls or assert the exact arg tuple). Production reads
// the real OneHit. Swap via the exported field in tests.
var oneHit = func(w *world.World, ch, victim *types.CharData, dt int) int {
	return OneHit(w, ch, victim, dt)
}

// oneHitOffhand is the same indirection but routes through the offhand
// weapon (WEAR_DUAL_WIELD). Tests swap this alongside `oneHit` to verify
// dual-wield alternation without touching real combat math.
var oneHitOffhand = func(w *world.World, ch, victim *types.CharData, dt int) int {
	wield := handler.GetEqChar(ch, types.WEAR_DUAL_WIELD)
	return oneHitFull(w, ch, victim, dt, wield)
}

// IsAttackSuppressed reports whether ch has TIMER_ASUPRESSED active,
// indicating they're in a temporary no-attack grace window. Mirrors C
// fight.c:74-92 is_attack_supressed.
func IsAttackSuppressed(ch *types.CharData) bool {
	if ch == nil {
		return false
	}
	t := handler.GetTimerPtr(ch, types.TIMER_ASUPRESSED)
	if t == nil {
		return false
	}
	if t.Value == -1 {
		return true
	}
	return t.Count >= 1
}

// MultiHit dispatches a full round of attacks from ch against victim.
// Mirrors C fight.c:972 multi_hit. The order is:
//
//  1. PLR_NICE early-return for PC-vs-PC (C fight.c:982).
//  2. Attack-suppress early-return (TIMER_ASUPRESSED).
//  3. ACT_NOATTACK early-return for NPCs.
//  4. Primary swing.
//  5. Short-circuit on single-hit skills (backstab/circle/pounce) and on
//     target change / retcode.
//  6. AFF_BERSERK extra hit (C fight.c:1004-1008).
//  7. Dual-wield learned roll with dual_bonus + low-move penalty.
//  8. NPC path: NumAttacks loop (C fight.c:1034), returns.
//  9. PC path: 6-tier cascade second..seventh_attack (C fight.c:1071-1141).
//
// Returns rNONE, rVICT_DIED, rCHAR_DIED, or rBOTH_DIED.
func MultiHit(w *world.World, ch, victim *types.CharData, dt int) int {
	// Mutual PC-vs-PC combat sets TIMER_RECENTFIGHT on both participants
	// for 11 violence pulses (~33s), unless the attacker is wearing the
	// PLR_NICE flag which suppresses multi-hit entirely. Matches C
	// fight.c:982-986.
	if !ch.IsNPC() && !victim.IsNPC() {
		if ch.Act.IsSet(types.PLR_NICE) {
			return rNONE
		}
		handler.AddTimer(ch, types.TIMER_RECENTFIGHT, 11, "", 0)
		handler.AddTimer(victim, types.TIMER_RECENTFIGHT, 11, "", 0)
	}

	// Attack-suppressed — skip the round entirely (C fight.c:988).
	if IsAttackSuppressed(ch) {
		return rNONE
	}

	// ACT_NOATTACK mob — skip the round entirely (C fight.c:991).
	if ch.IsNPC() && ch.Act.IsSet(types.ACT_NOATTACK) {
		return rNONE
	}

	// Primary swing.
	retcode := oneHit(w, ch, victim, dt)
	if retcode != rNONE {
		return retcode
	}
	if ch.Fighting == nil || ch.Fighting.Who != victim {
		return rNONE
	}

	// Single-hit skills never cascade (C fight.c:997).
	if (gsnBackstab != -1 && dt == gsnBackstab) ||
		(gsnCircle != -1 && dt == gsnCircle) ||
		(gsnPounce != -1 && dt == gsnPounce) {
		return rNONE
	}

	// Berserk extra hit (C fight.c:1004-1008). NPCs get 100% chance; PCs
	// roll against LEARNED(berserk) * 6 / 2. AFF_BERSERK must be set.
	if ch.AffectedBy.IsSet(types.AFF_BERSERK) {
		berserkChance := 100
		if !ch.IsNPC() {
			berserkChance = 0
			if ch.PCData != nil && gsnBerserk != -1 && gsnBerserk < types.MAX_SKILL {
				berserkChance = ch.PCData.Learned[gsnBerserk] * 6 / 2
			}
		}
		if numberPercent() < berserkChance {
			retcode = oneHit(w, ch, victim, dt)
			if retcode != rNONE {
				return retcode
			}
			if ch.Fighting == nil || ch.Fighting.Who != victim {
				return rNONE
			}
		}
	}

	// Dual-wield learned roll (C fight.c:1010-1026). On success fire an
	// extra OneHit (which will swing the offhand once G8's alternation
	// lands) and compute dual_bonus for the G3 cascade tiers. NPCs use
	// level directly; PCs use Learned[gsn_dual_wield].
	dualBonus := 0
	if handler.GetEqChar(ch, types.WEAR_DUAL_WIELD) != nil {
		var chance int
		if ch.IsNPC() {
			dualBonus = ch.Level / 10
			chance = ch.Level
		} else if ch.PCData != nil && gsnDualWield != -1 && gsnDualWield < types.MAX_SKILL {
			learned := ch.PCData.Learned[gsnDualWield]
			dualBonus = learned / 10
			chance = learned
		}
		if numberPercent() < chance {
			learnFromSuccess(ch, gsnDualWield)
			// G8: the dual-wield bonus swing uses the OFFHAND weapon
			// (WEAR_DUAL_WIELD), not WEAR_WIELD. C achieves this via a
			// static dual_flip bool; we pass the wield explicitly.
			retcode = oneHitOffhand(w, ch, victim, dt)
			if retcode != rNONE {
				return retcode
			}
			if ch.Fighting == nil || ch.Fighting.Who != victim {
				return rNONE
			}
		} else {
			learnFromFailure(ch, gsnDualWield)
		}
	}

	// Low-move penalty (C fight.c:1028-1029). Independent of whether the
	// learned roll succeeded — any character below 10 move gets -20 to the
	// cascade tier chance.
	if ch.Move < 10 {
		dualBonus = -20
	}

	// NPC predetermined number of attacks (C fight.c:1034). Primary swing
	// already consumed one slot, so loop from 1..NumAttacks-1. NPCs do not
	// run the PC cascade. Stance adds extra attacks unless either side is
	// STANCE_MONKEY (C fight.c:1042-1044).
	if ch.IsNPC() {
		tempAttacks := ch.NumAttacks
		if ch.Stance != types.STANCE_MONKEY && victim.Stance != types.STANCE_MONKEY &&
			ch.Stance > types.STANCE_NONE && ch.Stance < types.MAX_STANCE {
			tempAttacks += StanceIndex[ch.Stance].NumAttacks
		}
		if tempAttacks > 1 {
			for i := 1; i < tempAttacks; i++ {
				retcode = oneHit(w, ch, victim, dt)
				if retcode != rNONE {
					return retcode
				}
				if ch.Fighting == nil || ch.Fighting.Who != victim {
					return rNONE
				}
			}
		}
		return rNONE
	}

	// PC GM bonus-attack loop (C fight.c:1058-1069). When the PC's stance
	// mastery has reached grand-master (>= STANCE_GRAND_MASTER = 200) and
	// neither side is STANCE_MONKEY, fire the stance's NumAttacks worth of
	// extra OneHits before the cascade.
	if ch.Stance != types.STANCE_MONKEY && victim.Stance != types.STANCE_MONKEY &&
		ch.Stance > types.STANCE_NONE && ch.Stance < types.MAX_STANCE &&
		ch.PCData != nil && ch.PCData.Stances[ch.Stance] >= types.STANCE_GRAND_MASTER {
		for i := 0; i < StanceIndex[ch.Stance].NumAttacks; i++ {
			retcode = oneHit(w, ch, victim, dt)
			if retcode != rNONE {
				return retcode
			}
			if ch.Fighting == nil || ch.Fighting.Who != victim {
				return rNONE
			}
		}
	}

	// PC-only cascade. Each tier rolls numberPercent() against a
	// learned-derived chance; on success fire an extra OneHit, propagate
	// retcode, and learn-from-success. On failure, learn-from-failure.
	// Mirrors C fight.c:1071-1141.
	if retcode = cascadeTier(w, ch, victim, dt, gsnSecondAttack, dualBonus, tierSecond); retcode != rNONE {
		return retcode
	}
	if ch.Fighting == nil || ch.Fighting.Who != victim {
		return rNONE
	}
	if retcode = cascadeTier(w, ch, victim, dt, gsnThirdAttack, dualBonus, tierThird); retcode != rNONE {
		return retcode
	}
	if ch.Fighting == nil || ch.Fighting.Who != victim {
		return rNONE
	}
	if retcode = cascadeTier(w, ch, victim, dt, gsnFourthAttack, dualBonus, tierFourth); retcode != rNONE {
		return retcode
	}
	if ch.Fighting == nil || ch.Fighting.Who != victim {
		return rNONE
	}
	if retcode = cascadeTier(w, ch, victim, dt, gsnFifthAttack, dualBonus, tierFifth); retcode != rNONE {
		return retcode
	}
	if ch.Fighting == nil || ch.Fighting.Who != victim {
		return rNONE
	}
	if retcode = cascadeTier(w, ch, victim, dt, gsnSixthAttack, dualBonus, tierSixth); retcode != rNONE {
		return retcode
	}
	if ch.Fighting == nil || ch.Fighting.Who != victim {
		return rNONE
	}
	if retcode = cascadeTier(w, ch, victim, dt, gsnSeventhAttack, dualBonus, tierSeventh); retcode != rNONE {
		return retcode
	}

	return rNONE
}

// cascadeTier runs one multi-attack tier. Uses tier-specific math that
// matches the C expressions in fight.c:1071-1141 exactly:
//
//	second:  (LEARNED + dual_bonus) / 1.5
//	third:   (LEARNED + dual_bonus * 1.5) / 2
//	fourth:  (LEARNED + dual_bonus * 2) / 3
//	fifth:   (LEARNED + dual_bonus * 3) / 4
//	sixth:   (LEARNED + dual_bonus * 4) / 4
//	seventh: (LEARNED + dual_bonus * 5) / 4
//
// NPCs would use ch.Level here, but NPCs never reach this path (the NPC
// branch in MultiHit returns early). So we always compute PC chance.
//
// gsn == -1 means the skill is not in the table at all; treat the chance
// as 0 (skill can never fire). Still call learnFromFailure so boot-time
// mis-registration gets some signal — actually learnFromFailure no-ops
// when gsn is -1 via act's own guard, so the call is harmless.
func cascadeTier(w *world.World, ch, victim *types.CharData, dt int, gsn int, dualBonus int, tier tierSpec) int {
	chance := 0
	if gsn != -1 && ch.PCData != nil && gsn >= 0 && gsn < types.MAX_SKILL {
		chance = tier.compute(ch.PCData.Learned[gsn], dualBonus)
	}
	if numberPercent() < chance {
		learnFromSuccess(ch, gsn)
		retcode := oneHit(w, ch, victim, dt)
		if retcode != rNONE {
			return retcode
		}
		if ch.Fighting == nil || ch.Fighting.Who != victim {
			return rNONE
		}
	} else {
		learnFromFailure(ch, gsn)
	}
	return rNONE
}

// tierSpec encapsulates the tier-specific arithmetic so each cascade
// branch is a data-driven call rather than 6 copy-pasted blocks. All
// expressions below use integer truncation to match C's `(int)((...)/K)`
// behavior — SMAUG's C casts a double back to int.
type tierSpec struct {
	compute func(learned, dualBonus int) int
}

var (
	tierSecond  = tierSpec{compute: func(l, b int) int { return (l + b) * 2 / 3 }} // /1.5 == *2/3
	tierThird   = tierSpec{compute: func(l, b int) int { return (l + b*3/2) / 2 }}
	tierFourth  = tierSpec{compute: func(l, b int) int { return (l + b*2) / 3 }}
	tierFifth   = tierSpec{compute: func(l, b int) int { return (l + b*3) / 4 }}
	tierSixth   = tierSpec{compute: func(l, b int) int { return (l + b*4) / 4 }}
	tierSeventh = tierSpec{compute: func(l, b int) int { return (l + b*5) / 4 }}
)

// dirName returns the name of a direction.
func dirName(dir int) string {
	names := []string{"north", "east", "south", "west", "up", "down",
		"northeast", "northwest", "southeast", "southwest"}
	if dir >= 0 && dir < len(names) {
		return names[dir]
	}
	return "somewhere"
}

// OneHit resolves a single attack from ch against victim using the
// primary wielded weapon (WEAR_WIELD). Returns a retcode: rNONE on
// hit/miss, rVICT_DIED on kill or on an already-dead victim / out-of-
// room mismatch (mirrors C fight.c:1394).
//
// The dual-wield alternation (C's `static bool dual_flip`) is handled by
// MultiHit, which calls the package-private oneHitFull with the offhand
// weapon for the bonus swing. This avoids the cross-fighter race bug
// that C's static variable introduced.
func OneHit(w *world.World, ch *types.CharData, victim *types.CharData, dt int) int {
	wield := handler.GetEqChar(ch, types.WEAR_WIELD)
	return oneHitFull(w, ch, victim, dt, wield)
}

// oneHitFull is the underlying implementation. Takes `wield` explicitly
// so callers can select the primary or offhand weapon per C's dual-flip
// semantics, without any process-global state.
func oneHitFull(w *world.World, ch *types.CharData, victim *types.CharData, dt int, wield *types.ObjData) int {
	if victim.Hit <= 0 || ch.InRoom != victim.InRoom {
		return rVICT_DIED
	}

	// Weapon proficiency bonus (G6). C fight.c:1418-1420 + 1533 + 1544-1545
	// + 1566-1567. Applied to victim_ac before the hit roll, to dam on
	// hit, and to learn_from_failure on miss. PC-only, level > 5, wield
	// required — WeaponProfBonusCheck handles the gates.
	profBonus, profGsn := WeaponProfBonusCheck(ch, wield)

	// Calculate thac0
	var thac0 int
	if ch.IsNPC() {
		thac0 = ch.MobThac0
	} else {
		// Player thac0: interpolate from class table (simplified)
		thac0 = 20 - ch.Level
	}
	thac0 -= ch.Hitroll
	thac0 -= types.StrApp[ch.GetCurrStr()].ToHit

	// Victim AC (capped at -19). Prof bonus applies AFTER the cap to the
	// running AC value — C fight.c:1533 adds prof_bonus directly to
	// victim_ac without re-capping, so a high-prof PC can push victim_ac
	// below -19. Match C exactly.
	victimAC := util.UMAX(-19, victim.Armor/10) + profBonus

	// Roll d20
	diceroll := rollD20()

	// Hit/miss check
	if diceroll == 0 || (diceroll != 19 && diceroll < thac0-victimAC) {
		// Miss — if we had a prof weapon, learn-from-failure on miss
		// (C fight.c:1544-1545).
		if profGsn != -1 {
			learnFromFailure(ch, profGsn)
		}
		return Damage(w, ch, victim, 0, dt)
	}

	// Calculate damage
	var dam int
	if wield != nil {
		dam = util.NumberRange(wield.Value[1], wield.Value[2])
	} else {
		if ch.BareNumDie > 0 && ch.BareSizeDie > 0 {
			dam = util.DiceRoll(ch.BareNumDie, ch.BareSizeDie)
		} else {
			dam = util.NumberRange(1, 4)
		}
	}

	// Add damroll and str bonus
	dam += ch.Damroll
	dam += types.StrApp[ch.GetCurrStr()].ToDam
	// Prof bonus damage on hit (C fight.c:1566-1567). Integer truncation
	// matches C's (int)(prof_bonus/4) cast.
	if profBonus != 0 {
		dam += profBonus / 4
	}

	// Position multipliers (attacker)
	switch ch.Position {
	case types.POS_BERSERK:
		dam = dam * 6 / 5
	case types.POS_AGGRESSIVE:
		dam = dam * 11 / 10
	case types.POS_DEFENSIVE:
		dam = dam * 85 / 100
	case types.POS_EVASIVE:
		dam = dam * 4 / 5
	}

	// Sleeping victim = double damage
	if victim.Position < types.POS_SLEEPING {
		dam *= 2
	}

	if dam <= 0 {
		dam = 1
	}

	// Sanctuary halves damage
	if victim.AffectedBy.IsSet(types.AFF_SANCTUARY) {
		dam /= 2
	}

	// Stance damage multipliers (C fight.c:2549-2594). Suppressed when
	// either combatant is STANCE_MONKEY. Applied as: dam = dam * (dam_done
	// / 100) * max(stances[stance]/200, 0.5), then dam /= (dam_taken / 100)
	// * max(stances[stance]/200, 0.5) for the victim's side. We work in
	// integer math — the C code casts to float, but the final dam is an
	// int, so we integer-truncate at each step.
	dam = applyStanceDamage(ch, victim, dam)

	return damageWith(w, ch, victim, dam, dt, wield)
}

// applyStanceDamage applies the attacker/victim stance damage modifiers
// from C fight.c:2549-2594. Pure function — no IO, only ch.Stance /
// victim.Stance, .Stances[] mastery values, and StanceIndex data.
//
// Returns the modified dam. Guarantees: never returns <0; preserves the
// original dam unchanged when either side is STANCE_MONKEY, when the
// attacker's stance has dam_done == 0, or when the victim's stance has
// dam_taken == 0.
func applyStanceDamage(ch, victim *types.CharData, dam int) int {
	if ch.Stance == types.STANCE_MONKEY || victim.Stance == types.STANCE_MONKEY {
		return dam
	}
	// Attacker side.
	if ch.Stance > types.STANCE_NONE && ch.Stance < types.MAX_STANCE &&
		StanceIndex[ch.Stance].DamDone > 0 {
		dam = dam * StanceIndex[ch.Stance].DamDone / 100
		mastery := stanceMastery(ch, ch.Stance)
		// temp_dam = mastery / 200.0, min 0.5. In integer terms we multiply
		// by `max(mastery, 100)` and divide by 200.
		eff := mastery
		if eff < 100 {
			eff = 100
		}
		dam = dam * eff / 200
	}
	// Victim side.
	if victim.Stance > types.STANCE_NONE && victim.Stance < types.MAX_STANCE &&
		StanceIndex[victim.Stance].DamTaken > 0 {
		dam = dam * StanceIndex[victim.Stance].DamTaken / 100
		mastery := stanceMastery(victim, victim.Stance)
		eff := mastery
		if eff < 100 {
			eff = 100
		}
		// C does dam /= temp_dam, so we divide by eff/200 == dam * 200 / eff.
		if eff > 0 {
			dam = dam * 200 / eff
		}
	}
	if dam < 0 {
		dam = 0
	}
	return dam
}

// stanceMastery reads the per-character mastery level for a given stance.
// PC reads ch.PCData.Stances[stance]; NPC reads ch.pIndexData.Stances[stance]
// when the index is present. Returns 0 on a nil path.
func stanceMastery(ch *types.CharData, stance int) int {
	if stance <= types.STANCE_NONE || stance >= types.MAX_STANCE {
		return 0
	}
	if ch.IsNPC() {
		if ch.IndexData == nil {
			return 0
		}
		return ch.IndexData.Stances[stance]
	}
	if ch.PCData == nil {
		return 0
	}
	return ch.PCData.Stances[stance]
}

// Damage applies damage to a victim. Returns a retcode indicating if
// the victim died. Callers without a wielded-weapon context pass no obj;
// DamMessage will pick verbs from the attackTable/skill registry instead.
func Damage(w *world.World, ch *types.CharData, victim *types.CharData, dam int, dt int) int {
	return damageWith(w, ch, victim, dam, dt, nil)
}

// damageWith is the shared worker for OneHit and external callers. It threads
// the attacker's wielded obj (or nil) through to the damage-message dispatcher.
func damageWith(w *world.World, ch, victim *types.CharData, dam, dt int, obj *types.ObjData) int {
	if victim.Hit <= 0 {
		return rNONE
	}

	// Ensure fighting is set
	if ch.Fighting == nil && ch != victim {
		StartFighting(ch, victim)
	}
	if victim.Fighting == nil && victim != ch {
		StartFighting(victim, ch)
	}

	// Apply damage
	victim.Hit -= dam

	// Fire obj-prog DAMAGE on every worn/equipped item on victim whenever
	// damage lands. MVP: iterate carrying and pick items at a wear loc other
	// than WEAR_NONE. C-fidelity is weaker (C's damage_obj is per-hit on one
	// item) but this preserves test hooks on equipped obj progs.
	if dam > 0 && ObjDamageHook != nil {
		for _, obj := range victim.Carrying {
			if obj == nil || obj.WearLoc == types.WEAR_NONE {
				continue
			}
			ObjDamageHook(ch, obj)
		}
	}

	// Fire HITPRCNT progs on NPC victims now that HP has been reduced. Do
	// this before position/death processing so that progs (e.g., calling for
	// help, teleporting) can run while the mob is still alive.
	if victim.IsNPC() && victim.Hit > 0 && HitprcntHook != nil {
		HitprcntHook(victim, ch)
	}

	// Send damage messages via the SMAUG damage-message dispatcher. For
	// weapon hits (dt >= TYPE_HIT) we pass the attacker's wielded obj so
	// the message uses the weapon's short_descr as the attack word; for
	// skill/spell sn we pass nil (the skill's own strings take over).
	DamMessage(ch, victim, dam, dt, obj)

	// Update position and send status messages
	oldPos := victim.Position
	updatePos(victim)

	if victim.Position != oldPos && victim.Position <= types.POS_INCAP {
		switch victim.Position {
		case types.POS_INCAP:
			victim.Send("You are incapacitated and will slowly die, if not aided.\n\r")
			StopFighting(victim, true)
		case types.POS_STUNNED:
			victim.Send("You are stunned, but will probably recover.\n\r")
			StopFighting(victim, true)
		}
	}

	// Death check
	if victim.Position == types.POS_DEAD {
		// Arena branch: if this is a PvP death in a ROOM_ARENA room,
		// fire the arena victory path (heal/teleport/no corpse) and
		// skip the normal death processing. See
		// plan-phase6-arena.md §G6. C ref: fight.c:2861-2915.
		if ArenaVictoryCheck(ch, victim) {
			return rVICT_DIED
		}

		// Room-prog RDEATH fires before corpse generation / ExtractChar so
		// the prog sees the victim still in room.People.
		if DeathRoomHook != nil {
			DeathRoomHook(victim)
		}
		StopFighting(ch, true)

		// XP gain for killer (player killing NPC)
		if !ch.IsNPC() && victim.IsNPC() {
			xpGain := computeXP(ch, victim)
			ch.Exp += xpGain
			ch.Sendf("You receive %d experience points.\n\r", xpGain)
		}

		if victim.IsNPC() {
			MakeCorpse(w, victim)
			handler.ExtractChar(w, victim, true)
		} else {
			// PC death: reset to resting with 1 HP
			victim.Hit = 1
			victim.Mana = 1
			victim.Move = 1
			victim.Position = types.POS_RESTING
			victim.Send("You have been KILLED!\n\r")
			// Full death handling (corpse, XP loss, etc.) will be expanded later
		}

		return rVICT_DIED
	}

	return rNONE
}

// MakeCorpse creates a corpse object from a dead character and transfers
// their inventory into it.
func MakeCorpse(w *world.World, ch *types.CharData) {
	if ch.InRoom == nil {
		return
	}

	var corpse *types.ObjData

	if ch.IsNPC() {
		corpse = &types.ObjData{
			Name:        fmt.Sprintf("corpse %s", ch.Name),
			ShortDescr:  fmt.Sprintf("the corpse of %s", ch.ShortDescr),
			Description: fmt.Sprintf("The corpse of %s is lying here.", ch.ShortDescr),
			ItemType:    types.ITEM_CORPSE_NPC,
			Timer:       6,
			Weight:      ch.Weight,
		}
	} else {
		corpse = &types.ObjData{
			Name:        fmt.Sprintf("corpse %s", ch.Name),
			ShortDescr:  fmt.Sprintf("the corpse of %s", ch.Name),
			Description: fmt.Sprintf("The corpse of %s is lying here.", ch.Name),
			ItemType:    types.ITEM_CORPSE_PC,
			Timer:       40,
			Weight:      ch.Weight,
		}
	}

	w.AddObj(corpse)

	// Transfer inventory to corpse
	for len(ch.Carrying) > 0 {
		obj := ch.Carrying[len(ch.Carrying)-1]
		handler.ObjFromChar(obj)
		handler.ObjToObj(obj, corpse)
	}

	// Store gold in corpse value[0] for looting
	if ch.Gold > 0 {
		corpse.Value[0] = ch.Gold
		ch.Gold = 0
	}

	handler.ObjToRoom(corpse, ch.InRoom)
}

// updatePos sets position based on current HP.
func updatePos(ch *types.CharData) {
	if ch.Hit <= -ch.MaxHit {
		ch.Position = types.POS_DEAD
	} else if ch.Hit <= -ch.MaxHit/2 {
		ch.Position = types.POS_STUNNED
	} else if ch.Hit <= 0 {
		ch.Position = types.POS_INCAP
	}
}

// rollD20 returns a value 0-19 (equivalent to C's number_bits(5) clamped < 20).
// Declared as a var so tests can stub it via save/restore; see
// TestRollD20_IsSeam for the pattern. Tests that mutate this var MUST NOT
// use t.Parallel() — the mutation is package-global.
var rollD20 = func() int {
	for {
		v := util.NumberBits(5)
		if v < 20 {
			return v
		}
	}
}

// computeXP calculates XP gained for killing a victim.
// Based on victim's base XP scaled by level difference.
func computeXP(ch *types.CharData, victim *types.CharData) int {
	xp := victim.Exp
	if xp <= 0 {
		// Fallback: base XP from level
		xp = victim.Level * victim.Level * 10
	}

	// Level difference scaling
	diff := victim.Level - ch.Level
	switch {
	case diff >= 5:
		xp = xp * 3 / 2 // 150%
	case diff >= 1:
		xp = xp * 11 / 10 // 110%
	case diff >= -3:
		// Same range, no modifier
	case diff >= -8:
		xp = xp * 3 / 4 // 75%
	default:
		xp = xp / 4 // 25% for very low level mobs
	}

	if xp < 1 {
		xp = 1
	}
	return xp
}
