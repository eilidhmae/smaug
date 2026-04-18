package combat

import (
	"github.com/eilidhmae/smaug/internal/types"
)

// WeaponProfBonusCheck ports the non-ENABLE_WEAPONPROF branch of C
// src/fight.c:1256-1317 weapon_prof_bonus_check. Returns the bonus to
// apply (to victim_ac before the hit roll, and to dam on hit) and the
// resolved prof gsn (used for learnFromFailure on miss). PCs only,
// level > 5, wield required. Returns (0, -1) for NPCs, low-level PCs,
// or unarmed attackers.
//
// Bonus formula: (LEARNED(ch, prof_gsn) - 50) / 10. Negative when a PC
// is below 50% learned — SMAUG deliberately penalizes unlearned weapons.
func WeaponProfBonusCheck(ch *types.CharData, wield *types.ObjData) (int, int) {
	if ch.IsNPC() || ch.Level <= 5 || wield == nil {
		return 0, -1
	}
	profGsn := profGsnForWeapon(wield)
	if profGsn == -1 {
		return 0, -1
	}
	if ch.PCData == nil || profGsn >= types.MAX_SKILL {
		return 0, -1
	}
	learned := ch.PCData.Learned[profGsn]
	bonus := (learned - 50) / 10
	return bonus, profGsn
}

// profGsnForWeapon maps wield.Value[3] (the DAM_* damage type) to the
// corresponding weapon-proficiency gsn. Mirrors C's switch in
// weapon_prof_bonus_check exactly. Returns -1 for unrecognized types
// (C's `default:` arm).
func profGsnForWeapon(wield *types.ObjData) int {
	switch wield.Value[3] {
	case types.DAM_HIT, types.DAM_SUCTION, types.DAM_BITE, types.DAM_BLAST:
		return gsnPugilism
	case types.DAM_SLASH, types.DAM_SLICE:
		return gsnLongBlades
	case types.DAM_PIERCE, types.DAM_STAB:
		return gsnShortBlades
	case types.DAM_WHIP:
		return gsnFlexibleArms
	case types.DAM_CLAW:
		return gsnTalonousArms
	case types.DAM_POUND, types.DAM_CRUSH:
		return gsnBludgeons
	case types.DAM_BOLT, types.DAM_ARROW, types.DAM_DART,
		types.DAM_STONE, types.DAM_PEA:
		return gsnMissileWeapons
	}
	return -1
}
