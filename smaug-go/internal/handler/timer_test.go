package handler

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

func TestAddTimer_NewChar(t *testing.T) {
	ch := &types.CharData{}
	AddTimer(ch, types.TIMER_RECENTFIGHT, 11, "", 0)
	if len(ch.Timers) != 1 {
		t.Fatalf("want len 1, got %d", len(ch.Timers))
	}
	tm := ch.Timers[0]
	if tm.Type != types.TIMER_RECENTFIGHT {
		t.Errorf("want Type=%d, got %d", types.TIMER_RECENTFIGHT, tm.Type)
	}
	if tm.Count != 11 {
		t.Errorf("want Count=11, got %d", tm.Count)
	}
	if tm.DoFun != "" {
		t.Errorf("want DoFun='', got %q", tm.DoFun)
	}
	if tm.Value != 0 {
		t.Errorf("want Value=0, got %d", tm.Value)
	}
}

func TestAddTimer_Upsert(t *testing.T) {
	ch := &types.CharData{}
	AddTimer(ch, types.TIMER_RECENTFIGHT, 11, "", 0)
	AddTimer(ch, types.TIMER_RECENTFIGHT, 5, "foo", 42)
	if len(ch.Timers) != 1 {
		t.Fatalf("want len 1 (upsert), got %d", len(ch.Timers))
	}
	tm := ch.Timers[0]
	if tm.Count != 5 {
		t.Errorf("want Count=5, got %d", tm.Count)
	}
	if tm.DoFun != "foo" {
		t.Errorf("want DoFun='foo', got %q", tm.DoFun)
	}
	if tm.Value != 42 {
		t.Errorf("want Value=42, got %d", tm.Value)
	}
}

func TestAddTimer_DifferentTypes(t *testing.T) {
	ch := &types.CharData{}
	AddTimer(ch, types.TIMER_RECENTFIGHT, 11, "", 0)
	AddTimer(ch, types.TIMER_ASUPRESSED, 3, "", -1)
	if len(ch.Timers) != 2 {
		t.Fatalf("want len 2, got %d", len(ch.Timers))
	}
}

func TestGetTimer_Present(t *testing.T) {
	ch := &types.CharData{}
	AddTimer(ch, types.TIMER_RECENTFIGHT, 7, "", 0)
	if got := GetTimer(ch, types.TIMER_RECENTFIGHT); got != 7 {
		t.Errorf("want 7, got %d", got)
	}
}

func TestGetTimer_Absent(t *testing.T) {
	ch := &types.CharData{}
	if got := GetTimer(ch, types.TIMER_RECENTFIGHT); got != 0 {
		t.Errorf("want 0, got %d", got)
	}
}

func TestGetTimer_NilChar(t *testing.T) {
	if got := GetTimer(nil, types.TIMER_RECENTFIGHT); got != 0 {
		t.Errorf("want 0, got %d", got)
	}
}

func TestGetTimerPtr_Present(t *testing.T) {
	ch := &types.CharData{}
	AddTimer(ch, types.TIMER_RECENTFIGHT, 7, "fn", 9)
	p := GetTimerPtr(ch, types.TIMER_RECENTFIGHT)
	if p == nil {
		t.Fatalf("want non-nil")
	}
	if p.Count != 7 || p.DoFun != "fn" || p.Value != 9 || p.Type != types.TIMER_RECENTFIGHT {
		t.Errorf("unexpected fields: %+v", *p)
	}
}

func TestGetTimerPtr_Absent(t *testing.T) {
	ch := &types.CharData{}
	if p := GetTimerPtr(ch, types.TIMER_RECENTFIGHT); p != nil {
		t.Errorf("want nil, got %+v", p)
	}
}

func TestGetTimerPtr_NilChar(t *testing.T) {
	if p := GetTimerPtr(nil, types.TIMER_RECENTFIGHT); p != nil {
		t.Errorf("want nil, got %+v", p)
	}
}

func TestRemoveTimer_Present(t *testing.T) {
	ch := &types.CharData{}
	AddTimer(ch, types.TIMER_RECENTFIGHT, 5, "", 0)
	AddTimer(ch, types.TIMER_ASUPRESSED, 3, "", -1)
	RemoveTimer(ch, types.TIMER_RECENTFIGHT)
	if len(ch.Timers) != 1 {
		t.Fatalf("want len 1, got %d", len(ch.Timers))
	}
	if ch.Timers[0].Type != types.TIMER_ASUPRESSED {
		t.Errorf("want TIMER_ASUPRESSED remaining, got Type=%d", ch.Timers[0].Type)
	}
}

func TestRemoveTimer_Absent(t *testing.T) {
	ch := &types.CharData{}
	AddTimer(ch, types.TIMER_RECENTFIGHT, 5, "", 0)
	before := len(ch.Timers)
	RemoveTimer(ch, types.TIMER_ASUPRESSED)
	if len(ch.Timers) != before {
		t.Errorf("slice changed: was %d now %d", before, len(ch.Timers))
	}
}

func TestRemoveTimer_NilChar(t *testing.T) {
	RemoveTimer(nil, types.TIMER_RECENTFIGHT)
}

func TestExtractTimer_Present(t *testing.T) {
	ch := &types.CharData{}
	AddTimer(ch, types.TIMER_RECENTFIGHT, 5, "", 0)
	AddTimer(ch, types.TIMER_ASUPRESSED, 3, "", -1)
	ptr := GetTimerPtr(ch, types.TIMER_RECENTFIGHT)
	if ptr == nil {
		t.Fatal("setup: timer ptr should be non-nil")
	}
	ExtractTimer(ch, ptr)
	if len(ch.Timers) != 1 {
		t.Fatalf("want len 1, got %d", len(ch.Timers))
	}
	if ch.Timers[0].Type != types.TIMER_ASUPRESSED {
		t.Errorf("want TIMER_ASUPRESSED remaining, got Type=%d", ch.Timers[0].Type)
	}
}

func TestExtractTimer_NilTimer(t *testing.T) {
	ch := &types.CharData{}
	AddTimer(ch, types.TIMER_RECENTFIGHT, 5, "", 0)
	ExtractTimer(ch, nil)
	if len(ch.Timers) != 1 {
		t.Errorf("slice unexpectedly changed: %d", len(ch.Timers))
	}
}

func TestExtractTimer_NotInList(t *testing.T) {
	ch := &types.CharData{}
	AddTimer(ch, types.TIMER_RECENTFIGHT, 5, "", 0)
	stranger := &types.TimerData{Type: types.TIMER_PKILLED, Count: 1}
	ExtractTimer(ch, stranger)
	if len(ch.Timers) != 1 {
		t.Errorf("slice unexpectedly changed: %d", len(ch.Timers))
	}
}

func TestDecrementTimers_Empty(t *testing.T) {
	ch := &types.CharData{}
	DecrementTimers(ch)
}

func TestDecrementTimers_NilChar(t *testing.T) {
	DecrementTimers(nil)
}

func TestDecrementTimers_CountsDown(t *testing.T) {
	ch := &types.CharData{}
	AddTimer(ch, types.TIMER_RECENTFIGHT, 3, "", 0)

	DecrementTimers(ch)
	if len(ch.Timers) != 1 || ch.Timers[0].Count != 2 {
		t.Fatalf("after 1st: want len=1 count=2, got len=%d count=%d", len(ch.Timers), countOrZero(ch))
	}

	DecrementTimers(ch)
	if len(ch.Timers) != 1 || ch.Timers[0].Count != 1 {
		t.Fatalf("after 2nd: want len=1 count=1, got len=%d count=%d", len(ch.Timers), countOrZero(ch))
	}

	DecrementTimers(ch)
	if len(ch.Timers) != 0 {
		t.Fatalf("after 3rd: want len=0 (expired), got len=%d count=%d", len(ch.Timers), countOrZero(ch))
	}
}

func TestDecrementTimers_PermanentValueMinus1(t *testing.T) {
	ch := &types.CharData{}
	AddTimer(ch, types.TIMER_ASUPRESSED, 5, "", -1)
	for i := 0; i < 10; i++ {
		DecrementTimers(ch)
	}
	if len(ch.Timers) != 1 {
		t.Fatalf("want len=1, got %d", len(ch.Timers))
	}
	if ch.Timers[0].Count != 5 {
		t.Errorf("want Count=5 (permanent), got %d", ch.Timers[0].Count)
	}
}

func TestDecrementTimers_MixedPermanentAndNormal(t *testing.T) {
	ch := &types.CharData{}
	AddTimer(ch, types.TIMER_ASUPRESSED, 5, "", -1)
	AddTimer(ch, types.TIMER_RECENTFIGHT, 1, "", 0)

	DecrementTimers(ch)
	if len(ch.Timers) != 1 {
		t.Fatalf("want len=1 (normal expired), got %d", len(ch.Timers))
	}
	if ch.Timers[0].Type != types.TIMER_ASUPRESSED {
		t.Errorf("want TIMER_ASUPRESSED persisting, got Type=%d", ch.Timers[0].Type)
	}
	if ch.Timers[0].Count != 5 {
		t.Errorf("want permanent Count=5, got %d", ch.Timers[0].Count)
	}
}

func countOrZero(ch *types.CharData) int {
	if len(ch.Timers) == 0 {
		return 0
	}
	return ch.Timers[0].Count
}
