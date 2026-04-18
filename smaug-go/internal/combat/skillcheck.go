package combat

import "github.com/eilidhmae/smaug/internal/types"

// Skill-check hooks wired from boot. `combat` cannot import `act` directly
// (that package already depends on combat), so we publish function-variable
// seams and let boot fill them in at startup. Each is nil-safe and
// degrades gracefully: NPCs always succeed on canUseSkill, PCs always fail,
// the learn-from-* hooks no-op, and lookupSkillSlot returns -1.
//
// Mirrors the existing pattern used for HitprcntHook / VoidHook /
// ObjDamageHook / RfightHook / DeathRoomHook already in this package.
var (
	// CanUseSkillHook reports whether the skill identified by gsn succeeds
	// at the given percent roll. Ports act.canUseSkill — NPC 85% default,
	// PC compares percent against ch.PCData.Learned[gsn].
	CanUseSkillHook func(ch *types.CharData, percent int, gsn int) bool
	// LearnFromSuccessHook improves a PC's proficiency on a successful use.
	LearnFromSuccessHook func(ch *types.CharData, gsn int)
	// LearnFromFailureHook improves a PC's proficiency on a failed use
	// (slower than success — capped at adept-1 in act's impl).
	LearnFromFailureHook func(ch *types.CharData, gsn int)
	// LookupSkillSlotHook resolves a skill name to a gsn int. Resolved
	// once at boot and cached into combat-package globals (see
	// resolveCombatGSNs) to keep the hot path free of per-attack lookups.
	LookupSkillSlotHook func(name string) int
)

// canUseSkill is a thin wrapper around the hook that gives a safe default
// when no hook is installed (test environments, early boot).
func canUseSkill(ch *types.CharData, percent int, gsn int) bool {
	if CanUseSkillHook != nil {
		return CanUseSkillHook(ch, percent, gsn)
	}
	// Fallback mirrors act.canUseSkill's coarse default: NPC succeeds if
	// percent < 85, PC always fails (no Learned table reachable here).
	if ch != nil && ch.IsNPC() {
		return percent < 85
	}
	return false
}

// learnFromSuccess forwards to the installed hook or silently no-ops.
func learnFromSuccess(ch *types.CharData, gsn int) {
	if LearnFromSuccessHook != nil {
		LearnFromSuccessHook(ch, gsn)
	}
}

// learnFromFailure forwards to the installed hook or silently no-ops.
func learnFromFailure(ch *types.CharData, gsn int) {
	if LearnFromFailureHook != nil {
		LearnFromFailureHook(ch, gsn)
	}
}

// lookupSkillSlot forwards to the installed hook or returns -1.
// Callers should generally resolve gsn names once at boot (see
// ResolveGSNs) rather than calling this from a hot path.
func lookupSkillSlot(name string) int {
	if LookupSkillSlotHook != nil {
		return LookupSkillSlotHook(name)
	}
	return -1
}

// Resolved gsn slots cached at boot. -1 means "skill not in the table";
// callers must handle that (e.g. treat second-attack as 0% chance for
// a PC if gsnSecondAttack == -1).
var (
	gsnSecondAttack  = -1
	gsnThirdAttack   = -1
	gsnFourthAttack  = -1
	gsnFifthAttack   = -1
	gsnSixthAttack   = -1
	gsnSeventhAttack = -1
	gsnDualWield     = -1
	gsnBerserk       = -1
	gsnBackstab      = -1
	gsnCircle        = -1
	gsnPounce        = -1

	gsnPugilism       = -1
	gsnLongBlades     = -1
	gsnShortBlades    = -1
	gsnFlexibleArms   = -1
	gsnTalonousArms   = -1
	gsnBludgeons      = -1
	gsnMissileWeapons = -1
)

// ResolveGSNs looks up every gsn name combat cares about via
// LookupSkillSlotHook. Called once by boot after the skill table is
// loaded and LookupSkillSlotHook is wired. Missing skills leave the
// cached slot at -1 (skill-specific code must guard).
func ResolveGSNs() {
	if LookupSkillSlotHook == nil {
		return
	}
	gsnSecondAttack = LookupSkillSlotHook("second attack")
	gsnThirdAttack = LookupSkillSlotHook("third attack")
	gsnFourthAttack = LookupSkillSlotHook("fourth attack")
	gsnFifthAttack = LookupSkillSlotHook("fifth attack")
	gsnSixthAttack = LookupSkillSlotHook("sixth attack")
	gsnSeventhAttack = LookupSkillSlotHook("seventh attack")
	gsnDualWield = LookupSkillSlotHook("dual wield")
	gsnBerserk = LookupSkillSlotHook("berserk")
	gsnBackstab = LookupSkillSlotHook("backstab")
	gsnCircle = LookupSkillSlotHook("circle")
	gsnPounce = LookupSkillSlotHook("pounce")

	gsnPugilism = LookupSkillSlotHook("pugilism")
	gsnLongBlades = LookupSkillSlotHook("long blades")
	gsnShortBlades = LookupSkillSlotHook("short blades")
	gsnFlexibleArms = LookupSkillSlotHook("flexible arms")
	gsnTalonousArms = LookupSkillSlotHook("talonous arms")
	gsnBludgeons = LookupSkillSlotHook("bludgeons")
	gsnMissileWeapons = LookupSkillSlotHook("missile weapons")
}
