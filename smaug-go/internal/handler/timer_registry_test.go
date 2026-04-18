package handler

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// resetTimerRegistry wipes the timer-function registry and installs a
// t.Cleanup restoring the zeroed state. Called from every G2 test so
// parallel subtests never collide on the package-level map.
func resetTimerRegistry(t *testing.T) {
	t.Helper()
	saved := snapshotTimerRegistry()
	ClearTimerRegistry()
	t.Cleanup(func() {
		ClearTimerRegistry()
		for name, fn := range saved {
			RegisterTimerFunc(name, fn)
		}
	})
}

func snapshotTimerRegistry() map[string]TimerFunc {
	out := make(map[string]TimerFunc)
	for k, v := range timerRegistry {
		out[k] = v
	}
	return out
}

// capturedBugs wires util.Bug into a per-test slice.
func capturedBugs(t *testing.T) *[]string {
	t.Helper()
	bugs := []string{}
	prev := util.BugSink
	util.SetBugSink(func(msg string) { bugs = append(bugs, msg) })
	t.Cleanup(func() { util.SetBugSink(prev) })
	return &bugs
}

func TestRegisterTimerFunc_LookupRoundTrip(t *testing.T) {
	resetTimerRegistry(t)
	called := 0
	RegisterTimerFunc("do_spy", func(ch *types.CharData, arg string) {
		called++
	})
	fn := LookupTimerFunc("do_spy")
	if fn == nil {
		t.Fatal("LookupTimerFunc returned nil for registered name")
	}
	fn(nil, "")
	if called != 1 {
		t.Errorf("registered fn not invoked, called=%d", called)
	}
}

func TestLookupTimerFunc_UnknownReturnsNil(t *testing.T) {
	resetTimerRegistry(t)
	if fn := LookupTimerFunc("nonexistent"); fn != nil {
		t.Errorf("unknown name should return nil, got %v", fn)
	}
}

func TestRegisterTimerFunc_NilFuncPurges(t *testing.T) {
	resetTimerRegistry(t)
	RegisterTimerFunc("do_spy", func(_ *types.CharData, _ string) {})
	RegisterTimerFunc("do_spy", nil) // convention: nil purges
	if fn := LookupTimerFunc("do_spy"); fn != nil {
		t.Errorf("nil re-register should purge, got non-nil")
	}
}

func TestDecrementTimers_ExpiryDispatchesKnownDoFun(t *testing.T) {
	resetTimerRegistry(t)
	ch := &types.CharData{}
	var seenArg string
	var substateDuring int
	RegisterTimerFunc("do_spy", func(c *types.CharData, arg string) {
		seenArg = arg
		substateDuring = c.Substate
	})
	ch.Substate = 42 // prior state; should be restored after dispatch
	AddTimer(ch, types.TIMER_DO_FUN, 1, "do_spy", 17)

	DecrementTimers(ch)

	if seenArg != "" {
		t.Errorf("arg on dispatch = %q, want empty", seenArg)
	}
	if substateDuring != 17 {
		t.Errorf("substate during call = %d, want 17 (timer.Value)", substateDuring)
	}
	if len(ch.Timers) != 0 {
		t.Errorf("timer should be dropped after dispatch, got len=%d", len(ch.Timers))
	}
}

func TestDecrementTimers_ExpiryRestoresSubstate(t *testing.T) {
	resetTimerRegistry(t)
	ch := &types.CharData{}
	RegisterTimerFunc("do_spy", func(_ *types.CharData, _ string) {
		// callback does not touch Substate
	})
	ch.Substate = 99
	AddTimer(ch, types.TIMER_DO_FUN, 1, "do_spy", 17)

	DecrementTimers(ch)

	if ch.Substate != 99 {
		t.Errorf("substate after dispatch = %d, want 99 (restored)", ch.Substate)
	}
}

func TestDecrementTimers_ExpiryDispatchesUnknownDoFunLogsBug(t *testing.T) {
	resetTimerRegistry(t)
	bugs := capturedBugs(t)
	ch := &types.CharData{}
	AddTimer(ch, types.TIMER_DO_FUN, 1, "does_not_exist", 5)

	DecrementTimers(ch)

	if len(ch.Timers) != 0 {
		t.Errorf("expired unknown-DoFun timer should still drop, got len=%d", len(ch.Timers))
	}
	if len(*bugs) == 0 {
		t.Errorf("expected a BUG log for unknown do_fun name, got none")
	}
}

func TestDecrementTimers_EmptyDoFunDropsSilently(t *testing.T) {
	resetTimerRegistry(t)
	bugs := capturedBugs(t)
	ch := &types.CharData{}
	AddTimer(ch, types.TIMER_DO_FUN, 1, "", 0)

	DecrementTimers(ch)

	if len(ch.Timers) != 0 {
		t.Errorf("empty-DoFun timer should drop, got len=%d", len(ch.Timers))
	}
	if len(*bugs) != 0 {
		t.Errorf("empty DoFun should not BUG, got %v", *bugs)
	}
}

func TestDecrementTimers_DoFunThatReExtends(t *testing.T) {
	resetTimerRegistry(t)
	ch := &types.CharData{}
	RegisterTimerFunc("do_reextend", func(c *types.CharData, _ string) {
		AddTimer(c, types.TIMER_DO_FUN, 5, "do_reextend", 0)
	})
	AddTimer(ch, types.TIMER_DO_FUN, 1, "do_reextend", 0)

	DecrementTimers(ch)

	if len(ch.Timers) != 1 {
		t.Fatalf("re-extended timer should persist, got len=%d", len(ch.Timers))
	}
	if ch.Timers[0].Count != 5 {
		t.Errorf("re-extended Count = %d, want 5", ch.Timers[0].Count)
	}
}

func TestDecrementTimers_DoFunNilChar(t *testing.T) {
	// Ensure the expiry branch nil-checks before dispatch.
	resetTimerRegistry(t)
	DecrementTimers(nil)
}

// Coverage for the aliasing concern raised during the G2 adversary
// self-review: a DoFun callback that appends a NEW-type timer during
// iteration must leave BOTH the new timer AND any surviving original
// timers intact.
func TestDecrementTimers_DoFunAppendsNewTypeTimer(t *testing.T) {
	resetTimerRegistry(t)
	ch := &types.CharData{}
	RegisterTimerFunc("do_spawn", func(c *types.CharData, _ string) {
		// Append a different-type timer while we're expiring.
		AddTimer(c, types.TIMER_RECENTFIGHT, 7, "", 0)
	})
	// Sibling TIMER_ASUPRESSED timer should survive through the decrement.
	AddTimer(ch, types.TIMER_ASUPRESSED, 5, "", 0)
	AddTimer(ch, types.TIMER_DO_FUN, 1, "do_spawn", 0)

	DecrementTimers(ch)

	// Expect: TIMER_ASUPRESSED (count=4) + TIMER_RECENTFIGHT (count=7).
	haveAsup, haveRecent := false, false
	for _, tm := range ch.Timers {
		switch tm.Type {
		case types.TIMER_ASUPRESSED:
			haveAsup = true
			if tm.Count != 4 {
				t.Errorf("TIMER_ASUPRESSED count = %d, want 4", tm.Count)
			}
		case types.TIMER_RECENTFIGHT:
			haveRecent = true
			if tm.Count != 7 {
				t.Errorf("TIMER_RECENTFIGHT count = %d, want 7", tm.Count)
			}
		}
	}
	if !haveAsup {
		t.Error("TIMER_ASUPRESSED should survive")
	}
	if !haveRecent {
		t.Error("TIMER_RECENTFIGHT (callback-added) should be present")
	}
}

func TestClearTimerRegistry(t *testing.T) {
	resetTimerRegistry(t)
	RegisterTimerFunc("do_a", func(_ *types.CharData, _ string) {})
	RegisterTimerFunc("do_b", func(_ *types.CharData, _ string) {})
	ClearTimerRegistry()
	if fn := LookupTimerFunc("do_a"); fn != nil {
		t.Errorf("after Clear, do_a should be nil")
	}
	if fn := LookupTimerFunc("do_b"); fn != nil {
		t.Errorf("after Clear, do_b should be nil")
	}
}
