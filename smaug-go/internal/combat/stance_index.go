package combat

import "github.com/eilidhmae/smaug/internal/types"

// StanceInfo captures the subset of C's `struct stance_data` that affects
// combat directly — extra attacks per round, damage-done multiplier, and
// damage-taken multiplier. SMAUG's `db/system/stances.dat` is an empty
// stub in this repo, so the defaults here are derived from reasonable
// SMAUG 2.0 conventions (and verified against the C stances.c zeroing
// block in load_stances at stances.c:322-344). When/if the data file is
// populated, load_stances will override these values.
//
// Values chosen:
//   - Tiger/Dragon: aggressive stances with extra attacks and high damage
//   - Crab: defensive — reduced damage taken
//   - Monkey: special — zeroes out damage and attack bonuses (C fight.c
//     explicitly suppresses the stance block when either combatant is in
//     STANCE_MONKEY)
//   - Normal: a middle stance with no modifiers
//   - Swallow/Mongoose/Bull/Viper/Crane/Mantis: varied mild modifiers
//
// Fields:
//   - NumAttacks: added to NPC NumAttacks total in MultiHit's NPC branch,
//     and forms the count for the PC GM-bonus-attack loop.
//   - DamDone (percent): dam *= DamDone/100.0 — so >100 amplifies.
//   - DamTaken (percent): victim's dam /= DamTaken/100.0 — so <100 means
//     the victim takes *more* damage in this stance (the C code is
//     counterintuitive).
type StanceInfo struct {
	NumAttacks int
	DamDone    int // percent, 0 means "no modifier"
	DamTaken   int // percent, 0 means "no modifier"
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
