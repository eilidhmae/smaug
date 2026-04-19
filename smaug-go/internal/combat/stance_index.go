package combat

import "github.com/eilidhmae/smaug/internal/types"

// StanceInfo captures C's `struct stance_data` (src/mud.h, referenced from
// src/stances.c). The first three fields (NumAttacks/DamDone/DamTaken) are
// the combat-critical subset that Tranche B G1 (plan-tranche-b.md) shipped;
// the remainder is the Phase-6 OLC-facing extension added by
// plan-phase6-stances-olc.md (§D1) — resist/immune/suscept bitsets, dodge/
// parry deltas, dual-wield-forbid, max-weight gate, wait-state, class/race
// restrictions, special-move table index, prerequisite chain, and the
// entry messages.
//
// When Go ports C behaviour verbatim (including preserved C bugs) the
// inline comment points at the C line number for cross-reference.
//
// Value semantics:
//   - NumAttacks: added to NPC NumAttacks total in MultiHit's NPC branch,
//     and forms the count for the PC GM-bonus-attack loop.
//   - DamDone (percent): dam *= DamDone/100.0 — so >100 amplifies.
//   - DamTaken (percent): victim's dam /= DamTaken/100.0 — so <100 means
//     the victim takes *more* damage in this stance (the C code is
//     counterintuitive).
//   - Dodge / Parry: signed percent deltas, -50..50.
//   - Dual: non-zero forbids entry while WEAR_DUAL_WIELD is equipped.
//     (C `dual_wield` semantics.)
//   - MaxWeight: 0 means no ceiling; else stance is refused when
//     ch.CarryWeight > MaxWeight.
//   - Wait: WAIT_STATE on entry (0..25 pulses).
//   - Resist / Immune / Suscept: RIS_* bitsets toggled on/off by
//     UpdateStances against CharData.StanceResistant / StanceImmune /
//     StanceSusceptible. These are plain `int` to match the CharData
//     fields; callers setting them from `types.RIS_FIRE` etc. must
//     convert with `int(types.RIS_FIRE)`.
//   - Class / Race: bitmask intended, but C uses `IS_SET(mask, index)`
//     against a bare class/race index — observationally broken. Port
//     preserves the bug; see CanUseStance for the detail.
//   - SpecialMove / SpecialPercent: stub in both C and Go
//     (get_special_number returns 0). Feature unfinished upstream.
//   - Prereq[0]: primary prerequisite stance — must be GM
//     (>= STANCE_GRAND_MASTER = 200).
//   - Prereq[1]: optional secondary prerequisite. 0 = none.
//   - Self / Others: entry messages (AT_STANCE color).
type StanceInfo struct {
	// Combat-critical (Tranche B G1)
	NumAttacks int
	DamDone    int // percent, 0 means "no modifier"
	DamTaken   int // percent, 0 means "no modifier"

	// Phase-6 OLC extension — stances.c STANCE_DATA coverage
	Dodge          int    // signed percent, -50..50
	Parry          int    // signed percent, -50..50
	Dual           int    // non-zero = forbid dual-wield while in stance (C: dual_wield)
	MaxWeight      int    // 0 = no ceiling
	Wait           int    // WAIT_STATE on entry, 0..25
	Resist         int    // RIS_* bitset
	Immune         int    // RIS_* bitset
	Suscept        int    // RIS_* bitset
	Class          int    // class_restrictions bitmask (C bug — see CanUseStance)
	Race           int    // race_restrictions bitmask (same bug)
	SpecialMove    int    // index; stub in both C and Go (get_special_number returns 0)
	SpecialPercent int    // 0..100
	Prereq         [2]int // [0] primary prereq GM; [1] optional secondary
	Self           string // first-person entry message (AT_STANCE)
	Others         string // third-person room entry message
}

// StanceIndex is indexed by STANCE_NONE..STANCE_SWALLOW. Entry 0
// (STANCE_NONE) is zero so ch.Stance == STANCE_NONE skips the block.
var StanceIndex = [types.MAX_STANCE]StanceInfo{
	types.STANCE_NONE:     {NumAttacks: 0, DamDone: 0, DamTaken: 0},
	types.STANCE_NORMAL:   {NumAttacks: 0, DamDone: 100, DamTaken: 100},
	types.STANCE_VIPER:    {NumAttacks: 0, DamDone: 110, DamTaken: 90},
	types.STANCE_CRANE:    {NumAttacks: 0, DamDone: 100, DamTaken: 110},
	types.STANCE_CRAB:     {NumAttacks: 0, DamDone: 90, DamTaken: 120},
	types.STANCE_MONGOOSE: {NumAttacks: 1, DamDone: 100, DamTaken: 100},
	types.STANCE_BULL:     {NumAttacks: 0, DamDone: 120, DamTaken: 90},
	types.STANCE_MANTIS:   {NumAttacks: 0, DamDone: 110, DamTaken: 100},
	types.STANCE_DRAGON:   {NumAttacks: 1, DamDone: 120, DamTaken: 100},
	types.STANCE_TIGER:    {NumAttacks: 1, DamDone: 110, DamTaken: 110},
	types.STANCE_MONKEY:   {NumAttacks: 0, DamDone: 0, DamTaken: 0}, // suppresses
	types.STANCE_SWALLOW:  {NumAttacks: 1, DamDone: 100, DamTaken: 100},
}

// GetStanceName returns the canonical-cased name for a STANCE_* index, or
// the empty string if idx is out of range. Reverse of
// persist.GetStanceNumber. Mirrors C get_stance_name at
// src/stances.c:163-195.
func GetStanceName(idx int) string {
	switch idx {
	case types.STANCE_TIGER:
		return "Tiger"
	case types.STANCE_SWALLOW:
		return "Swallow"
	case types.STANCE_DRAGON:
		return "Dragon"
	case types.STANCE_MONKEY:
		return "Monkey"
	case types.STANCE_MANTIS:
		return "Mantis"
	case types.STANCE_VIPER:
		return "Viper"
	case types.STANCE_CRANE:
		return "Crane"
	case types.STANCE_CRAB:
		return "Crab"
	case types.STANCE_MONGOOSE:
		return "Mongoose"
	case types.STANCE_BULL:
		return "Bull"
	case types.STANCE_NORMAL:
		return "Normal"
	case types.STANCE_NONE:
		return "None"
	default:
		return ""
	}
}

// GetStanceMastery returns the caller's mastery level (0..200) for their
// current stance. Mirrors C get_stance_mastery at src/stances.c:151-160:
// STANCE_NONE → 0; NPC reads ch.IndexData.Stances[...]; PC reads
// ch.PCData.Stances[...]. Nil-safe on both sides.
func GetStanceMastery(ch *types.CharData) int {
	if ch == nil || ch.Stance == types.STANCE_NONE {
		return 0
	}
	if ch.Stance < 0 || ch.Stance >= types.MAX_STANCE {
		return 0
	}
	if ch.IsNPC() {
		if ch.IndexData == nil {
			return 0
		}
		return ch.IndexData.Stances[ch.Stance]
	}
	if ch.PCData == nil {
		return 0
	}
	return ch.PCData.Stances[ch.Stance]
}

// GetSpecialName is the Go port of C get_special_name at
// src/stances.c:1018-1022. The C implementation returns "None"
// unconditionally — the "special moves" feature is unfinished upstream.
func GetSpecialName(idx int) string { return "None" }

// GetSpecialNumber is the Go port of C get_special_number at
// src/stances.c:1012-1016. The C implementation returns 0
// unconditionally — the "special moves" feature is unfinished upstream.
func GetSpecialNumber(name string) int { return 0 }
