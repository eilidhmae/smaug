package combat

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// G2 — hooks exposed to the combat package so it can call the act-package
// skill-check helpers without importing act (which would cycle).

// When the hook is nil, canUseSkillHook falls through to a safe default
// (NPCs always succeed; PCs always fail — matches a "no skill table yet"
// environment).
func TestCanUseSkillHook_NilFallthrough(t *testing.T) {
	// Save + restore the hook so this test doesn't leak global state.
	saved := CanUseSkillHook
	t.Cleanup(func() { CanUseSkillHook = saved })
	CanUseSkillHook = nil

	npc := &types.CharData{}
	npc.Act.Set(types.ACT_IS_NPC)
	if !canUseSkill(npc, 50, 0) {
		t.Error("nil hook: NPC fallthrough should return true")
	}

	pc := &types.CharData{PCData: &types.PCData{}}
	if canUseSkill(pc, 50, 0) {
		t.Error("nil hook: PC fallthrough should return false")
	}
}

// When installed, canUseSkillHook is called with the exact (ch, percent, gsn)
// tuple we passed in.
func TestCanUseSkillHook_Installed(t *testing.T) {
	saved := CanUseSkillHook
	t.Cleanup(func() { CanUseSkillHook = saved })

	var gotPct, gotGsn int
	var gotCh *types.CharData
	CanUseSkillHook = func(ch *types.CharData, percent int, gsn int) bool {
		gotCh = ch
		gotPct = percent
		gotGsn = gsn
		return true
	}
	pc := &types.CharData{PCData: &types.PCData{}}
	if !canUseSkill(pc, 42, 7) {
		t.Error("installed hook returning true should be honored")
	}
	if gotCh != pc || gotPct != 42 || gotGsn != 7 {
		t.Errorf("hook received (%v, %d, %d); want (pc, 42, 7)", gotCh, gotPct, gotGsn)
	}
}

// learnFromSuccess / learnFromFailure hooks are nil-safe.
func TestLearnHooks_NilSafe(t *testing.T) {
	savedS := LearnFromSuccessHook
	savedF := LearnFromFailureHook
	t.Cleanup(func() {
		LearnFromSuccessHook = savedS
		LearnFromFailureHook = savedF
	})
	LearnFromSuccessHook = nil
	LearnFromFailureHook = nil

	// Should not panic when hooks are nil.
	ch := &types.CharData{}
	learnFromSuccess(ch, 0)
	learnFromFailure(ch, 0)
}

// Installed learn hooks see the exact args.
func TestLearnHooks_Installed(t *testing.T) {
	savedS := LearnFromSuccessHook
	savedF := LearnFromFailureHook
	t.Cleanup(func() {
		LearnFromSuccessHook = savedS
		LearnFromFailureHook = savedF
	})

	var successCalled, failureCalled bool
	var successGsn, failureGsn int
	LearnFromSuccessHook = func(_ *types.CharData, gsn int) {
		successCalled = true
		successGsn = gsn
	}
	LearnFromFailureHook = func(_ *types.CharData, gsn int) {
		failureCalled = true
		failureGsn = gsn
	}
	ch := &types.CharData{}
	learnFromSuccess(ch, 11)
	learnFromFailure(ch, 22)
	if !successCalled || successGsn != 11 {
		t.Errorf("success hook: called=%v gsn=%d; want true, 11", successCalled, successGsn)
	}
	if !failureCalled || failureGsn != 22 {
		t.Errorf("failure hook: called=%v gsn=%d; want true, 22", failureCalled, failureGsn)
	}
}

// LookupSkillSlotHook is nil-safe and when installed gets the given name.
func TestLookupSkillSlotHook_NilSafe(t *testing.T) {
	saved := LookupSkillSlotHook
	t.Cleanup(func() { LookupSkillSlotHook = saved })
	LookupSkillSlotHook = nil

	if got := lookupSkillSlot("anything"); got != -1 {
		t.Errorf("nil hook should return -1; got %d", got)
	}
}

func TestLookupSkillSlotHook_Installed(t *testing.T) {
	saved := LookupSkillSlotHook
	t.Cleanup(func() { LookupSkillSlotHook = saved })

	var gotName string
	LookupSkillSlotHook = func(name string) int {
		gotName = name
		return 99
	}
	if got := lookupSkillSlot("dual wield"); got != 99 {
		t.Errorf("hook should have returned 99; got %d", got)
	}
	if gotName != "dual wield" {
		t.Errorf("hook received %q; want %q", gotName, "dual wield")
	}
}

// ResolveGSNs looks up every combat-relevant gsn name via the installed
// hook and caches the result. Asserts the exact name set that boot will
// need resolved at startup: the 6 multi-attack skills (second..seventh),
// dual_wield, berserk, backstab, circle, pounce, and the 7 weapon-prof
// skills. If a caller adds a new one, this test will start missing it.
func TestResolveGSNs_CallsHookForAllExpectedNames(t *testing.T) {
	saved := LookupSkillSlotHook
	t.Cleanup(func() { LookupSkillSlotHook = saved })

	expected := []string{
		"second attack", "third attack", "fourth attack",
		"fifth attack", "sixth attack", "seventh attack",
		"dual wield", "berserk", "backstab", "circle", "pounce",
		"pugilism", "long blades", "short blades",
		"flexible arms", "talonous arms", "bludgeons", "missile weapons",
	}
	calls := map[string]bool{}
	LookupSkillSlotHook = func(name string) int {
		calls[name] = true
		return -1
	}
	ResolveGSNs()
	for _, name := range expected {
		if !calls[name] {
			t.Errorf("ResolveGSNs did not look up %q", name)
		}
	}
}

// ResolveGSNs correctly caches the returned gsn ints onto the package
// globals that MultiHit + OneHit will consult on the hot path.
func TestResolveGSNs_CachesReturnedGsns(t *testing.T) {
	saved := LookupSkillSlotHook
	t.Cleanup(func() { LookupSkillSlotHook = saved })

	// Hand-pick distinct ints so we can detect cross-assignment bugs.
	lookups := map[string]int{
		"second attack":   11,
		"third attack":    12,
		"fourth attack":   13,
		"fifth attack":    14,
		"sixth attack":    15,
		"seventh attack":  16,
		"dual wield":      17,
		"berserk":         18,
		"backstab":        19,
		"circle":          20,
		"pounce":          21,
		"pugilism":        30,
		"long blades":     31,
		"short blades":    32,
		"flexible arms":   33,
		"talonous arms":   34,
		"bludgeons":       35,
		"missile weapons": 36,
	}
	LookupSkillSlotHook = func(name string) int {
		if v, ok := lookups[name]; ok {
			return v
		}
		return -1
	}
	ResolveGSNs()
	checks := []struct {
		got  int
		want int
		name string
	}{
		{gsnSecondAttack, 11, "gsnSecondAttack"},
		{gsnThirdAttack, 12, "gsnThirdAttack"},
		{gsnFourthAttack, 13, "gsnFourthAttack"},
		{gsnFifthAttack, 14, "gsnFifthAttack"},
		{gsnSixthAttack, 15, "gsnSixthAttack"},
		{gsnSeventhAttack, 16, "gsnSeventhAttack"},
		{gsnDualWield, 17, "gsnDualWield"},
		{gsnBerserk, 18, "gsnBerserk"},
		{gsnBackstab, 19, "gsnBackstab"},
		{gsnCircle, 20, "gsnCircle"},
		{gsnPounce, 21, "gsnPounce"},
		{gsnPugilism, 30, "gsnPugilism"},
		{gsnLongBlades, 31, "gsnLongBlades"},
		{gsnShortBlades, 32, "gsnShortBlades"},
		{gsnFlexibleArms, 33, "gsnFlexibleArms"},
		{gsnTalonousArms, 34, "gsnTalonousArms"},
		{gsnBludgeons, 35, "gsnBludgeons"},
		{gsnMissileWeapons, 36, "gsnMissileWeapons"},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d", c.name, c.got, c.want)
		}
	}
}

// ResolveGSNs is nil-hook-safe.
func TestResolveGSNs_NilHookSafe(t *testing.T) {
	saved := LookupSkillSlotHook
	t.Cleanup(func() { LookupSkillSlotHook = saved })
	LookupSkillSlotHook = nil
	ResolveGSNs() // must not panic
}
