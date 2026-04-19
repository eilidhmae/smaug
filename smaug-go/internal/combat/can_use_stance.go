package combat

import "github.com/eilidhmae/smaug/internal/types"

// CanUseStance decides whether ch may enter newStance. Mirrors C
// can_use_stance at src/stances.c:236-281.
//
// Flow (C verbatim, with the preserved bugs flagged inline):
//  1. NPC: ability = (ch.IndexData != nil && ch.IndexData.Stances[new] > 0).
//  2. PC:
//     a. temp = StanceIndex[new].Prereq[0].
//     b. temp <= 0           → ability = true (no prereq).
//     c. Stances[temp] >= GM → recurse on Prereq[1]:
//     · Prereq[1] <= 0          → true
//     · Stances[Prereq[1]] >= GM→ true
//     · else                   → false
//     d. else                → false.
//  3. Post-gates (override any earlier TRUE):
//     · MaxWeight > 0 && ch.CarryWeight > MaxWeight → false
//     · Dual != 0 && ch has WEAR_DUAL_WIELD equipped → false
//     · IS_SET(Class, ch.Class) → false       // C bug: see below
//     · IS_SET(Race,  ch.Race ) → false       // C bug: see below
//  4. Return ability.
//
// Preserved C bug (plan-phase6-stances-olc.md Q4): C uses
// `IS_SET(mask, index)` where index is a bare class/race number, not a
// bit. `IS_SET(a,b) → ((a)&(b))` — so the test is `mask & index`, which
// is almost always 0. Port preserves this rather than silently fixing it,
// because no shipped stance populates Class/Race (the setters are also
// no-ops in C) and a forward-port to `mask & (1<<index)` would be a
// gameplay change, not a fidelity repair. Pinned by
// TestCanUseStance_ClassRestriction_CBug_PreservesMaskANDIndex.
func CanUseStance(ch *types.CharData, newStance int) bool {
	if ch == nil {
		return false
	}
	if newStance < 0 || newStance >= types.MAX_STANCE {
		return false
	}
	info := StanceIndex[newStance]

	var ability bool
	if ch.IsNPC() {
		ability = ch.IndexData != nil && ch.IndexData.Stances[newStance] > 0
	} else {
		temp := info.Prereq[0]
		switch {
		case temp <= 0:
			ability = true
		case ch.PCData != nil && temp < types.MAX_STANCE && ch.PCData.Stances[temp] >= types.STANCE_GRAND_MASTER:
			t2 := info.Prereq[1]
			switch {
			case t2 <= 0:
				ability = true
			case t2 < types.MAX_STANCE && ch.PCData.Stances[t2] >= types.STANCE_GRAND_MASTER:
				ability = true
			default:
				ability = false
			}
		default:
			ability = false
		}
	}

	// Post-gates — apply after the PC/NPC prereq block. These override
	// any earlier TRUE.
	if info.MaxWeight > 0 && ch.CarryWeight > info.MaxWeight {
		ability = false
	}
	if info.Dual != 0 && charHasDualWield(ch) {
		ability = false
	}
	// C bug preserved: `IS_SET(mask, index)` = `mask & index`, not
	// `mask & (1<<index)`. See plan §Q4.
	if info.Class != 0 && (info.Class&ch.Class) != 0 {
		ability = false
	}
	if info.Race != 0 && (info.Race&ch.Race) != 0 {
		ability = false
	}
	return ability
}

// charHasDualWield returns true when ch has an object equipped at
// WEAR_DUAL_WIELD. Inlined rather than calling handler.GetEqChar to
// avoid a combat → handler import.
func charHasDualWield(ch *types.CharData) bool {
	for _, obj := range ch.Carrying {
		if obj != nil && obj.WearLoc == types.WEAR_DUAL_WIELD {
			return true
		}
	}
	return false
}
