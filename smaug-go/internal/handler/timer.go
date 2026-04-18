package handler

import "github.com/eilidhmae/smaug/internal/types"

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
// No expiry dispatch: TIMER_DO_FUN callbacks are a scope-cut follow-up
// (see plan-timer-subsystem.md Open Question 4).
//
// Expected caller is `combat.ViolenceUpdate` at PULSE_VIOLENCE cadence.
// Exported (rather than unexported as originally planned) because the
// combat package imports handler; an unexported name would not be
// reachable from the caller. Nil-safe on ch.
func DecrementTimers(ch *types.CharData) {
	if ch == nil || len(ch.Timers) == 0 {
		return
	}
	kept := ch.Timers[:0]
	for _, t := range ch.Timers {
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
		// Expired: drop. No dispatch.
	}
	ch.Timers = kept
}
