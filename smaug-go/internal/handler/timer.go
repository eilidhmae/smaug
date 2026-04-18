package handler

import (
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// TimerFunc is the signature of a timer callback — identical to
// `command.CmdFunc` so existing command handlers can be registered as
// do_fun targets without a wrapper. Mirrors C's `DO_FUN`.
type TimerFunc func(ch *types.CharData, argument string)

// timerRegistry maps DoFun names (as persisted in TimerData.DoFun) to
// their Go implementations. Populated at boot by
// `internal/boot/boot.go`. Stateless in production; the G2 tests reset
// it via ClearTimerRegistry + snapshot/restore.
var timerRegistry = map[string]TimerFunc{}

// RegisterTimerFunc installs a callback under the given name. A nil fn
// purges any existing entry — this is how tests clean up between runs,
// and it mirrors the C `do_fun = NULL` convention.
func RegisterTimerFunc(name string, fn TimerFunc) {
	if name == "" {
		return
	}
	if fn == nil {
		delete(timerRegistry, name)
		return
	}
	timerRegistry[name] = fn
}

// LookupTimerFunc returns the registered TimerFunc, or nil when absent.
func LookupTimerFunc(name string) TimerFunc {
	if name == "" {
		return nil
	}
	return timerRegistry[name]
}

// ClearTimerRegistry empties the registry. Intended for test setup; the
// production code path never calls this.
func ClearTimerRegistry() {
	for k := range timerRegistry {
		delete(timerRegistry, k)
	}
}

// AddTimer upsert-adds a timer to ch.Timers. If a timer with the same Type
// already exists, Count / DoFun / Value are overwritten in place; otherwise
// a new *TimerData is appended. Mirrors C `add_timer` in `handler.c:5125`.
// Nil-safe on ch.
func AddTimer(ch *types.CharData, tType, count int, doFun string, value int) {
	if ch == nil {
		return
	}
	for _, t := range ch.Timers {
		if t == nil {
			continue
		}
		if t.Type == tType {
			t.Count = count
			t.DoFun = doFun
			t.Value = value
			return
		}
	}
	ch.Timers = append(ch.Timers, &types.TimerData{
		Type:  tType,
		Count: count,
		DoFun: doFun,
		Value: value,
	})
}

// GetTimer returns the Count of the timer with the given Type, or 0 when
// no such timer exists. Mirrors C `get_timer` in `handler.c:5160`.
// Nil-safe on ch.
func GetTimer(ch *types.CharData, tType int) int {
	if ch == nil {
		return 0
	}
	for _, t := range ch.Timers {
		if t == nil {
			continue
		}
		if t.Type == tType {
			return t.Count
		}
	}
	return 0
}

// GetTimerPtr returns a pointer to the timer with the given Type, or nil
// when absent. Mirrors C `get_timerptr` in `handler.c:5148`.
// Nil-safe on ch.
func GetTimerPtr(ch *types.CharData, tType int) *types.TimerData {
	if ch == nil {
		return nil
	}
	for _, t := range ch.Timers {
		if t == nil {
			continue
		}
		if t.Type == tType {
			return t
		}
	}
	return nil
}

// RemoveTimer removes any timer matching the given Type. No-op when absent.
// Mirrors C `remove_timer` in `handler.c:5185`. Nil-safe on ch.
func RemoveTimer(ch *types.CharData, tType int) {
	if ch == nil || len(ch.Timers) == 0 {
		return
	}
	kept := ch.Timers[:0]
	for _, t := range ch.Timers {
		if t == nil {
			continue
		}
		if t.Type == tType {
			continue
		}
		kept = append(kept, t)
	}
	ch.Timers = kept
}

// ExtractTimer removes a specific *TimerData from ch.Timers by pointer
// identity. No-op if t is nil or not in the slice. Mirrors C
// `extract_timer` in `handler.c:5171`. Nil-safe on ch.
func ExtractTimer(ch *types.CharData, t *types.TimerData) {
	if ch == nil || t == nil || len(ch.Timers) == 0 {
		return
	}
	kept := ch.Timers[:0]
	for _, cur := range ch.Timers {
		if cur == t {
			continue
		}
		kept = append(kept, cur)
	}
	ch.Timers = kept
}

// DecrementTimers is the per-violence-pulse decrement helper. Each
// non-permanent timer has its Count decremented by 1; timers whose Count
// reaches 0 are dropped. Timers with Value == -1 are permanent and never
// decrement (matches C `fight.c:74-98` for TIMER_ASUPRESSED; the
// plan-timer-subsystem generalizes the invariant to all timer types).
//
// Expiry dispatch (plan-tranche-b.md G2, mirrors C fight.c:415-427):
// when a TIMER_DO_FUN timer expires, we look up its DoFun name in the
// timer registry and invoke it with `ch.Substate = t.Value` during the
// call (substate restored afterwards). If the callback re-registers a
// timer of the same type (via AddTimer), the re-extended timer is
// preserved — matches C's `if (timer->count > 0) continue` at
// fight.c:425. Empty DoFun names drop silently; unknown names log a
// util.Bug and drop.
//
// SCOPE CUT — mid-decrement intercept: the C combat-aborts-skill path
// at src/fight.c:386-398 (fire timer with SUB_TIMER_DO_ABORT when the
// char is fighting) is NOT implemented here; no Go skill command
// currently sets a TIMER_DO_FUN, so the intercept is dead code. When
// the first such command ports, add the intercept at the top of this
// loop before the Value==-1 check.
//
// SCOPE CUT — interp intercept: the C command-interpreter abort path
// at src/interp.c:713-733 is similarly deferred for the same reason.
//
// Expected caller is `combat.ViolenceUpdate` at PULSE_VIOLENCE cadence.
// Exported (rather than unexported as originally planned) because the
// combat package imports handler; an unexported name would not be
// reachable from the caller. Nil-safe on ch.
func DecrementTimers(ch *types.CharData) {
	if ch == nil || len(ch.Timers) == 0 {
		return
	}
	// Snapshot the slice because a DoFun callback may call AddTimer on
	// the same ch, mutating ch.Timers mid-iteration. Build `kept` in a
	// fresh backing array so slice aliasing between kept and ch.Timers
	// cannot corrupt callback-appended entries.
	snap := make([]*types.TimerData, len(ch.Timers))
	copy(snap, ch.Timers)
	originalLen := len(ch.Timers)
	kept := make([]*types.TimerData, 0, len(ch.Timers))
	for _, t := range snap {
		if t == nil {
			continue
		}
		if t.Value == -1 {
			kept = append(kept, t)
			continue
		}
		t.Count--
		if t.Count > 0 {
			kept = append(kept, t)
			continue
		}
		// Expired.
		if t.Type == types.TIMER_DO_FUN && t.DoFun != "" {
			dispatchExpiredDoFun(ch, t)
			// C fight.c:425-426: if the do_fun re-extended the
			// timer (Count > 0 after the callback), keep it.
			if t.Count > 0 {
				kept = append(kept, t)
				continue
			}
		} else if t.Type == types.TIMER_DO_FUN && t.DoFun == "" {
			// Empty DoFun: drop silently, no BUG log.
		}
		// Default drop.
	}
	// If callbacks appended NEW timer entries (a different Type — same
	// type is upsert-mutated in place by AddTimer), they appear at
	// ch.Timers[originalLen:]. Merge them into `kept` so they survive.
	if len(ch.Timers) > originalLen {
		kept = append(kept, ch.Timers[originalLen:]...)
	}
	ch.Timers = kept
}

// dispatchExpiredDoFun handles the TIMER_DO_FUN expiry branch:
//  1. Save ch.Substate
//  2. Set ch.Substate = t.Value (C fight.c:421 `ch->substate = timer->value`)
//  3. Invoke the registered callback; if unknown, log util.Bug
//  4. Restore ch.Substate
func dispatchExpiredDoFun(ch *types.CharData, t *types.TimerData) {
	fn := LookupTimerFunc(t.DoFun)
	if fn == nil {
		util.Bug("DecrementTimers: unknown TIMER_DO_FUN %q", t.DoFun)
		return
	}
	savedSubstate := ch.Substate
	ch.Substate = t.Value
	fn(ch, "")
	ch.Substate = savedSubstate
}
