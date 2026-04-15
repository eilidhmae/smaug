package mudprog

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// setupSleepTest returns a mob+room+actor with a pipe-backed desc.
// The returned client must be Close()d; read output with readOutput().
func setupSleepTest(t *testing.T) (mob, actor *types.CharData, room *types.RoomIndexData, cleanup func() string) {
	t.Helper()
	SleepReset()

	room = &types.RoomIndexData{Vnum: 9500, Name: "Sleep Test Room"}
	mob = makeNPC("sleepmob")
	mob.InRoom = room
	room.People = append(room.People, mob)

	ch, client := makeTestChar("Watcher")
	ch.InRoom = room
	room.People = append(room.People, ch)

	cleanup = func() string {
		out := readOutput(ch, client)
		client.Close()
		return out
	}
	return mob, ch, room, cleanup
}

func TestSleep_DefersExecution(t *testing.T) {
	mob, actor, _, finish := setupSleepTest(t)

	// "mpsleep 2" should suspend execution; the following mpecho must NOT run this tick.
	prog := "mpecho before sleep\nmpsleep 2\nmpecho after sleep"
	Driver(prog, mob, actor, nil, nil, nil, false)

	// Check that only "before sleep" is in the queue (not "after sleep") right now.
	// We do this by flushing and reading what's been buffered to date.
	out := finish()
	if !strings.Contains(out, "before sleep") {
		t.Errorf("expected 'before sleep' to execute immediately, got %q", out)
	}
	if strings.Contains(out, "after sleep") {
		t.Errorf("'after sleep' must not execute until timer elapses, got %q", out)
	}
	if len(sleepQueue) != 1 {
		t.Fatalf("expected 1 queued sleep entry, got %d", len(sleepQueue))
	}
	if sleepQueue[0].Timer != 2 {
		t.Errorf("expected Timer=2, got %d", sleepQueue[0].Timer)
	}
}

func TestSleep_FiresAfterTimerElapses(t *testing.T) {
	mob, actor, _, finish := setupSleepTest(t)

	prog := "mpsleep 2\nmpecho delayed hello"
	Driver(prog, mob, actor, nil, nil, nil, false)

	// Tick 1: should NOT fire yet.
	SleepUpdate()
	if len(sleepQueue) != 1 {
		t.Fatalf("after 1 SleepUpdate, queue should still have entry, got %d", len(sleepQueue))
	}

	// Tick 2: should fire and drain.
	SleepUpdate()
	if len(sleepQueue) != 0 {
		t.Fatalf("after 2 SleepUpdates, queue should be empty, got %d", len(sleepQueue))
	}

	out := finish()
	if !strings.Contains(out, "delayed hello") {
		t.Errorf("expected 'delayed hello' after resume, got %q", out)
	}
}

func TestSleep_ReentrantMpsleepQueuesNewEntry(t *testing.T) {
	mob, actor, _, finish := setupSleepTest(t)

	// First mpsleep defers; the resumed script itself contains another mpsleep.
	prog := "mpsleep 1\nmpecho first resume\nmpsleep 2\nmpecho second resume"
	Driver(prog, mob, actor, nil, nil, nil, false)

	// Queue has 1 entry.
	if len(sleepQueue) != 1 {
		t.Fatalf("expected 1 entry after initial Driver call, got %d", len(sleepQueue))
	}

	// Tick 1: first sleep fires, runs "mpecho first resume", then hits nested mpsleep 2
	// which should queue a NEW entry (with "mpecho second resume" as its ComList).
	SleepUpdate()

	if len(sleepQueue) != 1 {
		t.Fatalf("after first resume, expected exactly 1 new queued entry, got %d", len(sleepQueue))
	}
	if sleepQueue[0].Timer != 2 {
		t.Errorf("new entry should have Timer=2, got %d", sleepQueue[0].Timer)
	}

	// Tick 2 and 3: decrement + fire.
	SleepUpdate()
	SleepUpdate()
	if len(sleepQueue) != 0 {
		t.Fatalf("queue should be empty after second resume, got %d", len(sleepQueue))
	}

	out := finish()
	if !strings.Contains(out, "first resume") {
		t.Errorf("expected 'first resume' in output, got %q", out)
	}
	if !strings.Contains(out, "second resume") {
		t.Errorf("expected 'second resume' after re-entrant resume, got %q", out)
	}
}

func TestSleep_ZeroTicksDefaultsLikeC(t *testing.T) {
	mob, actor, _, finish := setupSleepTest(t)

	// C: timer<1 -> progbug and reset to 4. Follow C.
	prog := "mpsleep 0\nmpecho eventually"
	Driver(prog, mob, actor, nil, nil, nil, false)

	if len(sleepQueue) != 1 {
		t.Fatalf("mpsleep 0 should still queue (with default timer), got %d entries", len(sleepQueue))
	}
	if sleepQueue[0].Timer != 4 {
		t.Errorf("mpsleep 0 should default Timer to 4 (matching C), got %d", sleepQueue[0].Timer)
	}

	// Must NOT fire immediately.
	out := finish()
	if strings.Contains(out, "eventually") {
		t.Errorf("'eventually' must not fire immediately on mpsleep 0, got %q", out)
	}
}

func TestSleep_MissingArgDefaultsTo4(t *testing.T) {
	mob, actor, _, _ := setupSleepTest(t)

	prog := "mpsleep\nmpecho after default"
	Driver(prog, mob, actor, nil, nil, nil, false)

	if len(sleepQueue) != 1 {
		t.Fatalf("mpsleep with no arg should queue, got %d", len(sleepQueue))
	}
	if sleepQueue[0].Timer != 4 {
		t.Errorf("mpsleep with no arg should default Timer to 4, got %d", sleepQueue[0].Timer)
	}
}

func TestSleep_PreservesDriverArgs(t *testing.T) {
	mob, actor, _, _ := setupSleepTest(t)

	obj := &types.ObjData{Name: "sword", ShortDescr: "a sword"}
	victim := makeNPC("victim")
	target := &types.ObjData{Name: "shield", ShortDescr: "a shield"}

	prog := "mpsleep 3\nmpecho done"
	Driver(prog, mob, actor, obj, victim, target, true)

	if len(sleepQueue) != 1 {
		t.Fatalf("expected 1 queued entry, got %d", len(sleepQueue))
	}
	sd := sleepQueue[0]
	if sd.Mob != mob || sd.Actor != actor || sd.Obj != obj || sd.Victim != victim || sd.Target != target {
		t.Error("driver args (mob/actor/obj/victim/target) must be preserved on the queue")
	}
	if !sd.SingleStep {
		t.Error("SingleStep must be preserved on the queue")
	}
	if sd.Type != types.MP_MOB {
		t.Errorf("Type should default to MP_MOB (%d), got %d", types.MP_MOB, sd.Type)
	}
}

func TestSleep_ResetClearsQueue(t *testing.T) {
	mob, actor, _, _ := setupSleepTest(t)

	Driver("mpsleep 2\nmpecho hi", mob, actor, nil, nil, nil, false)
	if len(sleepQueue) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(sleepQueue))
	}
	SleepReset()
	if len(sleepQueue) != 0 {
		t.Errorf("SleepReset should clear the queue, got %d", len(sleepQueue))
	}
}

func TestSleep_UpdateOnEmptyQueueIsNoop(t *testing.T) {
	SleepReset()
	// Must not panic.
	SleepUpdate()
	SleepUpdate()
	if len(sleepQueue) != 0 {
		t.Errorf("empty queue must remain empty after SleepUpdate, got %d", len(sleepQueue))
	}
}

func TestSleep_InsideFalseIfDoesNotQueue(t *testing.T) {
	// mpsleep inside a false-branch if should NOT fire (it's skipped like any other cmd).
	// C actually handles mpsleep BEFORE if/else evaluation, meaning it always fires. We
	// document this Go divergence: we treat mpsleep as a regular command, gated by if-state.
	// This matches SMAUG user expectation in practice (mpsleep inside a conditional).
	mob, actor, _, _ := setupSleepTest(t)

	prog := "if level($n) > 999\nmpsleep 5\nmpecho skipped\nendif\nmpecho end"
	Driver(prog, mob, actor, nil, nil, nil, false)

	if len(sleepQueue) != 0 {
		t.Errorf("mpsleep inside false-branch if should NOT queue (Go semantics), got %d entries", len(sleepQueue))
	}
}
