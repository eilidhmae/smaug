package act

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/combat"
)

// TestResolveGSNs_CallsHookForHideSneak verifies that act.ResolveGSNs
// asks combat.LookupSkillSlotHook for the two stealth gsn names used by
// learnFromSuccess for its silent-skill suppression branch. Matches
// plan-tranche-c.md G7 acceptance criterion A5.
func TestResolveGSNs_CallsHookForHideSneak(t *testing.T) {
	saved := combat.LookupSkillSlotHook
	t.Cleanup(func() { combat.LookupSkillSlotHook = saved })

	calls := map[string]bool{}
	combat.LookupSkillSlotHook = func(name string) int {
		calls[name] = true
		return -1
	}

	ResolveGSNs()

	if !calls["hide"] {
		t.Error("ResolveGSNs did not look up \"hide\"")
	}
	if !calls["sneak"] {
		t.Error("ResolveGSNs did not look up \"sneak\"")
	}
}

// TestResolveGSNs_CachesReturnedGsns verifies the gsn lookups are cached
// into the package-private gsnHide / gsnSneak vars that the
// learnFromSuccess silent-skill guard reads.
func TestResolveGSNs_CachesReturnedGsns(t *testing.T) {
	saved := combat.LookupSkillSlotHook
	savedHide, savedSneak := gsnHide, gsnSneak
	t.Cleanup(func() {
		combat.LookupSkillSlotHook = saved
		gsnHide, gsnSneak = savedHide, savedSneak
	})

	combat.LookupSkillSlotHook = func(name string) int {
		switch name {
		case "hide":
			return 77
		case "sneak":
			return 88
		}
		return -1
	}

	ResolveGSNs()

	if gsnHide != 77 {
		t.Errorf("gsnHide = %d, want 77", gsnHide)
	}
	if gsnSneak != 88 {
		t.Errorf("gsnSneak = %d, want 88", gsnSneak)
	}
}

// TestResolveGSNs_NilHook_NoPanic verifies the early-return guard when
// boot has not wired LookupSkillSlotHook yet (tests that construct
// act-package state directly without calling Boot).
func TestResolveGSNs_NilHook_NoPanic(t *testing.T) {
	saved := combat.LookupSkillSlotHook
	t.Cleanup(func() { combat.LookupSkillSlotHook = saved })

	combat.LookupSkillSlotHook = nil
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ResolveGSNs with nil hook panicked: %v", r)
		}
	}()
	ResolveGSNs()
}
