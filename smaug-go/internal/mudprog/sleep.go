package mudprog

import (
	"github.com/eilidhmae/smaug/internal/types"
)

// sleepQueue holds suspended mudprogs waiting for their timer to elapse.
//
// The game is single-threaded: the game loop processes input, runs updates
// (including SleepUpdate), and flushes output all in one goroutine. No mutex
// is needed. Appends from re-entrant mpsleep calls during resume (see
// SleepUpdate) are safe because we rebuild the slice via the kept[:0] pattern
// on a local variable before reassigning.
//
// Mirrors C's `first_mpsleep`/`last_mpsleep` doubly-linked list in update.c
// (reimplemented here as a slice because Go doesn't need intrusive lists).
var sleepQueue []*types.MProgSleepData

// SleepAdd appends a suspended mudprog to the queue.
func SleepAdd(sd *types.MProgSleepData) {
	if sd == nil {
		return
	}
	sleepQueue = append(sleepQueue, sd)
}

// SleepUpdate is called once per pulse from the game loop. It decrements each
// entry's Timer; entries whose Timer reaches 0 have their ComList resumed via
// Driver() and are then removed from the queue.
//
// Re-entrancy: a resumed mudprog may itself call mpsleep, which will append
// a new entry via SleepAdd. To keep that safe, we iterate the original queue
// by length captured at entry, build `kept` in place using the `kept[:0]`
// pattern, and then append anything new (added during resume) to the end.
//
// Mirrors C's `mpsleep_update` in src/update.c.
func SleepUpdate() {
	if len(sleepQueue) == 0 {
		return
	}

	// Snapshot the pre-resume queue so new entries appended during resume
	// (via re-entrant mpsleep) are not iterated this tick.
	snapshot := sleepQueue
	sleepQueue = nil

	kept := snapshot[:0]
	for _, sd := range snapshot {
		if sd == nil {
			continue
		}
		sd.Timer--
		if sd.Timer > 0 {
			kept = append(kept, sd)
			continue
		}
		// Defensive guard (partial fix): the sleep queue holds raw *CharData /
		// *RoomIndexData pointers that can be extracted between enqueue and
		// resume, which would nil-panic inside Driver's command calls. Skip
		// entries whose mob was extracted (no InRoom and no IndexData handle).
		// TODO(tier4): track a generation counter on extracted mobs / rooms /
		// objects so we can drop stale sleep entries deterministically instead
		// of relying on this heuristic.
		if sd.Mob == nil || (sd.Mob.InRoom == nil && sd.Mob.IndexData == nil) {
			continue
		}
		// Timer elapsed: resume the program. During this call, re-entrant
		// mpsleep invocations will append to the (currently reassigned) live
		// sleepQueue; we merge those back below.
		Driver(sd.ComList, sd.Mob, sd.Actor, sd.Obj, sd.Victim, sd.Target, sd.SingleStep)
	}

	// Any entries queued during resume live in sleepQueue now.
	// Prepend the survivors of this tick.
	if len(sleepQueue) == 0 {
		sleepQueue = kept
	} else {
		sleepQueue = append(kept, sleepQueue...)
	}
}

// SleepReset clears the sleep queue. Intended for tests.
func SleepReset() {
	sleepQueue = nil
}
