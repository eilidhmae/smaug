# Plan: Timer Subsystem — `handler.AddTimer` + `TIMER_RECENTFIGHT` + `TIMER_ASUPRESSED` Wiring

**Status:** Planned (2026-04-17). Adversary-verified research: CONCERNS (wrong file `db.c` vs `save.c` cited for `TIMER_PKILLED` persistence — corrected here; also confirmed `IsAttackSuppressed` is functionally dead in production).
**Priority:** P1 — unblocks quit-during-combat gate, activates already-wired `IsAttackSuppressed`, clears `TODO.md` follow-up from combat-depth G5.
**Scope:** New file `internal/handler/timer.go`. Modifications to `internal/game/update.go`, `internal/combat/combat.go`, `internal/act/playercfg.go` (or quit command file), plus tests.

---

## Problem

Three overlapping gaps:

1. **No generic timer subsystem in Go.** C has `add_timer` / `remove_timer` / `get_timer` / `get_timerptr` at `src/handler.c:5125-5195`. Go has the `TimerData` type (`internal/types/character.go:260-266`) and `CharData.Timers []*TimerData` (character.go:62), but the handler functions don't exist.
2. **`IsAttackSuppressed` is functionally dead.** `internal/combat/combat.go:152-167` reads `ch.Timers`, but nothing in production code ever writes to it. The test at `combat_test.go:1844` populates timers directly to prove the read path works, but in a real game session `IsAttackSuppressed` always returns `false`.
3. **`TIMER_RECENTFIGHT` never set.** Quit during PC-vs-PC combat is not blocked (C `src/act_comm.c:2875` blocks with "Your adrenaline is pumping too hard..."). Go's DoQuit gate is open.

## C Reference (authoritative, corrected for adversary finding)

- **`TIMER_DATA` struct** — `src/mud.h:2561-2569`.
- **Timer type enum** (8 values) — `src/mud.h:2555-2559`. Go's `enums.go:898-905` matches order and numeric values (iota-derived; `TIMER_ASUPRESSED = 6` in both).
- **Handler functions** — `src/handler.c:5125-5195`:
  - `add_timer(ch, type, count, do_fun, value)` at 5125 — upsert by type.
  - `get_timerptr(ch, type)` at 5148.
  - `get_timer(ch, type)` at 5160.
  - `extract_timer(ch, timer)` at 5171.
  - `remove_timer(ch, type)` at 5185.
- **Decrement loop** — `src/fight.c:382-430` in `violence_update`. Each pulse, every char's timers decrement; expired timers dispatch by type.
- **`TIMER_RECENTFIGHT` set** — `src/fight.c:980-986`, during `multi_hit`, on mutual PC-vs-PC damage (honors `PLR_NICE` early-return).
- **`TIMER_RECENTFIGHT` read** — `src/act_comm.c:2875` (quit gate); `src/deity.c:1498` (prayer gate); `src/act_wiz.c:2552-2554` (wiz stat display); `src/mud_prog.c:1433` (`timerskilled` if-check).
- **`TIMER_ASUPRESSED` gate** — `src/fight.c:74-98` (`is_attack_supressed`).
- **`TIMER_PKILLED` persistence** — `src/save.c:546` (save as `PTimer`), `src/save.c:1863` (load). **Adversary correction: worker cited `db.c` — the correct file is `save.c`.**

## Go Current State

- `CharData.Timers []*TimerData` exists but is written only by test code (`combat_test.go:1844`). No production writers.
- `IsAttackSuppressed` (combat.go:158) scans manually — correct logic, but without writers it always returns `false`.
- `charUpdate` (internal/game/update.go:167-242) has NO timer loop. Has an affect-duration loop (216-231); the timer loop would live adjacent.
- Mudprog `timerskilled` / `asupressed` / `pkadrenalized` if-checks at `internal/mudprog/ifcheck.go:888` are deferred with a TODO pointing at this gap.

## Design — Option C (Hybrid)

Ship the full generic subsystem with all four handler functions and a decrement loop in `charUpdate`. Wire consumer gates for `TIMER_RECENTFIGHT` (combat set + quit check) and `TIMER_ASUPRESSED` (use the new `GetTimer` in `IsAttackSuppressed`). Leave other timer types defined but unwired; follow-up work can consume them without touching infrastructure.

Rejected alternatives:
- **Option A (full subsystem + all consumers):** Out of scope. Too many follow-ups in one PR.
- **Option B (minimal `RecentFightTimer int` field):** Breaks C fidelity. Every future timer type adds another typed field. Doesn't unblock mudprog if-checks.

## Persistence

- Only `TIMER_PKILLED` is persisted in C (via `PTimer` in save.c). Other timers are transient (lost on logout).
- This plan does NOT add char-timer persistence. `TIMER_PKILLED` is not currently set by any Go code, so nothing persists. If PKilled mechanics land later, persistence lands alongside.
- Rationale: transient is correct for `TIMER_RECENTFIGHT` (in-session adrenaline cooldown) and `TIMER_ASUPRESSED` (combat disable during stun).

## Task Groups

### G1 — New file `internal/handler/timer.go`

- Functions to implement:
  ```go
  func AddTimer(ch *types.CharData, tType, count int, doFun string, value int)
  func GetTimer(ch *types.CharData, tType int) int         // returns count, 0 if not present
  func GetTimerPtr(ch *types.CharData, tType int) *types.TimerData  // nil if not present
  func RemoveTimer(ch *types.CharData, tType int)
  func ExtractTimer(ch *types.CharData, t *types.TimerData)
  ```
- `AddTimer` upsert: scan `ch.Timers`, if existing match on `Type`, overwrite `Count`/`DoFun`/`Value` in place. Else append a fresh `*TimerData`.
- `RemoveTimer` deletes by `Type`; uses slice-delete pattern that preserves order (or ignore order — C's linked list has no ordering invariant).
- **Test first:** `handler/timer_test.go` with cases:
  - `TestAddTimer_New` — empty Timers, after call len == 1, fields set.
  - `TestAddTimer_Upsert` — same type twice, len stays 1, fields updated.
  - `TestAddTimer_DifferentTypes` — two types, len == 2.
  - `TestGetTimer_Present` — returns count.
  - `TestGetTimer_Absent` — returns 0.
  - `TestGetTimerPtr_Present` / `TestGetTimerPtr_Absent`.
  - `TestRemoveTimer_Present` — removes, len decrements.
  - `TestRemoveTimer_Absent` — no-op, no panic.
  - `TestExtractTimer_NilTimer` — no panic.
- Mutation-verify each.

### G2 — Decrement loop in `violenceUpdate` (NOT `charUpdate`)

**Adversary-critical correction.** C decrements timers at PULSE_VIOLENCE (3-second cadence). Go's `charUpdate` fires at PULSE_TICK (70-second cadence per `internal/game/loop.go:170-178`). Putting the decrement loop in `charUpdate` would make `TIMER_RECENTFIGHT` count 11 last ~770 seconds instead of ~33 seconds — a 23x overestimate. That is unacceptable, not "drift."

- File: `internal/combat/violence.go` (or extract to a new helper — `violenceUpdate` at `internal/game/loop.go:157-162` calls `combat.ViolenceUpdate`). The decrement loop belongs inside that path, running once per violence pulse per char.
- Implementation placement: add a `decrementTimers(ch *types.CharData)` helper in `internal/handler/timer.go` and call it from `ViolenceUpdate`'s per-char loop. Keeps the subsystem self-contained.
- Loop body:
  ```go
  func decrementTimers(ch *types.CharData) {
      if len(ch.Timers) == 0 { return }
      kept := ch.Timers[:0]
      for _, t := range ch.Timers {
          if t == nil { continue }
          // Permanent suppress: Value == -1 never decrements. C uses this
          // only for TIMER_ASUPRESSED but the invariant is universal — any
          // timer with Value == -1 is permanent until explicitly removed.
          if t.Value == -1 {
              kept = append(kept, t)
              continue
          }
          t.Count--
          if t.Count > 0 {
              kept = append(kept, t)
              continue
          }
          // Expired — no dispatch in this plan. TIMER_DO_FUN callbacks are
          // a follow-up (doFun string registry not yet built).
      }
      ch.Timers = kept
  }
  ```
- The compact-in-place filter (`kept := ch.Timers[:0]`) is safe: the range copies the slice header at entry, so writes into `kept` do not corrupt the ongoing iteration.
- **Precondition:** Value == -1 means permanent. This is enforced UNIFORMLY across all timer types (not just `TIMER_ASUPRESSED`), matching C semantics at `fight.c:400-430`. Callers that set Value == -1 are responsible for explicit `RemoveTimer` to clear.
- **Test first:**
  - `TestDecrementTimers_Empty` — nil/empty slice, no panic.
  - `TestDecrementTimers_CountsDown` — count 3, 3 calls, removed on the third.
  - `TestDecrementTimers_PermanentValueMinus1` — `Value=-1`, 10 calls, timer persists.
  - `TestViolenceUpdate_DecrementsEveryChar` — integration: two chars with timers, one `ViolenceUpdate`, both decrement.
  - Mutation-verify: flip `> 0` to `>= 0`, check count-3-one-call test fails (timer should persist).

### G3 — Update `IsAttackSuppressed` to use `GetTimer`

- File: `internal/combat/combat.go:152-167`.
- Replace manual scan with:
  ```go
  func IsAttackSuppressed(ch *types.CharData) bool {
      if ch == nil { return false }
      t := handler.GetTimerPtr(ch, types.TIMER_ASUPRESSED)
      if t == nil { return false }
      if t.Value == -1 { return true }
      return t.Count >= 1
  }
  ```
- Now matches C `fight.c:74-98` semantics exactly (permanent-suppress via `Value == -1`).
- **Test first:** Existing `combat_test.go:1844` test should still pass. Add:
  - `TestIsAttackSuppressed_PermanentValue` — timer with `Value=-1`, `Count=0`, returns true.
  - `TestIsAttackSuppressed_ZeroCount` — timer with `Value=0`, `Count=0`, returns false.
- Mutation-verify the `Value == -1` branch.

### G4 — Set `TIMER_RECENTFIGHT` in `MultiHit`

**Adversary-critical correction.** `MultiHit` at `internal/combat/combat.go:185-193` already has the PLR_NICE early-return and the `!ch.IsNPC() && !victim.IsNPC()` gate. The plan's original snippet duplicated both guards AFTER `oneHit`, where the victim could already be dead/extracted. The correct placement is INSIDE the existing PC-vs-PC block, BEFORE the primary swing.

- File: `internal/combat/combat.go:189-193` — extend the existing PC-vs-PC block.
- Replace the comment "TIMER_RECENTFIGHT is not set here because the timer subsystem is not yet ported" with the live set calls.
- New block shape:
  ```go
  if !ch.IsNPC() && !victim.IsNPC() {
      if ch.Act.IsSet(types.PLR_NICE) {
          return rNONE
      }
      handler.AddTimer(ch, types.TIMER_RECENTFIGHT, 11, "", 0)
      handler.AddTimer(victim, types.TIMER_RECENTFIGHT, 11, "", 0)
  }
  ```
- **No new guard**; extending the existing guard. The `AddTimer` calls happen only on the non-nice PC-vs-PC path, before any damage resolution. Victim is guaranteed alive at this point.
- **Import:** add `github.com/eilidhmae/smaug/internal/handler` to `combat.go` if not already present (grep confirms it is — used elsewhere in combat).
- **Test first:**
  - `TestMultiHit_PCvsPCSetsRecentFightOnBoth` — two PCs, trigger `MultiHit`, assert both have `GetTimer(TIMER_RECENTFIGHT) > 0`.
  - `TestMultiHit_PCvsNPCDoesNotSet` — PC attacks mob, neither gets the timer.
  - `TestMultiHit_PLR_NICE_NoTimer` — attacker has PLR_NICE, MultiHit returns rNONE early, no timer set.
  - Mutation-verify: remove one of the two `AddTimer` calls, confirm the "both PCs" test fails on the missing side.

### G5 — Gate `DoQuit` on `TIMER_RECENTFIGHT`

- File: `internal/act/info.go:337` (verified — `DoQuit` lives here; no `rent` or `camp` command exists).
- Add after the existing `POS_FIGHTING` gate at `info.go:339-341`, before any save/cleanup:
  ```go
  if !ch.IsNPC() && handler.GetTimer(ch, types.TIMER_RECENTFIGHT) > 0 {
      ch.Send("Your adrenaline is pumping too hard to quit now!\n\r")
      return
  }
  ```
- Line-ending `\n\r` matches existing `ch.Send` convention in `info.go` (adversary caught this — `\r\n` was wrong in v1 plan).
- **Test first:** `TestDoQuit_BlockedByRecentFight` — PC with timer, call DoQuit, assert still `CON_PLAYING` and message sent. Clear timer, call again, assert quit proceeds.
- **Mutation-verify** the gate by removing the `if` block — test fails.

### G6 — Remove mudprog deferral comments (optional cleanup)

- `internal/mudprog/ifcheck.go:888` — update or remove the TODO; the subsystem now exists. Actual wiring of `timerskilled` / `asupressed` if-checks can land in a follow-up; just update the comment to reference that the infrastructure is ready.

### G7 — Documentation corrections

- This plan file already corrects the `db.c` → `save.c` citation.
- Update `TODO.md` combat-depth follow-up section:
  - Change "handler.AddTimer subsystem + TIMER_RECENTFIGHT wiring (deferred from G5)" → move to Done on landing.
- Append completion record to this plan.

## Acceptance Criteria

A1. All five handler timer functions exist in `internal/handler/timer.go` with passing tests.
A2. `AddTimer` is upsert (same type twice → len stays 1, fields updated).
A3. `violenceUpdate` (NOT `charUpdate`) decrements all timers at PULSE_VIOLENCE cadence; expired non-permanent timers are removed.
A4. Permanent suppress (`Value == -1`) does not decrement.
A5. `IsAttackSuppressed` uses `GetTimerPtr` + `Value == -1` branch.
A6. `MultiHit` sets `TIMER_RECENTFIGHT` on both PCs during mutual combat (gated by `PLR_NICE`).
A7. `DoQuit` blocks with C-exact message when timer > 0.
A8. `go test -count=3 ./...` green across all packages.
A9. `IsAttackSuppressed` verified live (not just test-only) via a combat integration test.

## Scope Cuts (follow-ups tracked in `TODO.md`)

- `TIMER_DO_FUN` callback dispatch (name-to-function resolver).
- `TIMER_PKILLED` persistence (no consumer yet).
- `TIMER_NUISANCE` / `TIMER_SHOVEDRAG` wiring (commands not ported).
- Mudprog `timerskilled` / `asupressed` / `pkadrenalized` if-checks.
- Deity prayer gate on `TIMER_RECENTFIGHT` (`src/deity.c:1498`).
- Wiz stat display of timer remaining.

## Open Questions

1. ~~**Decrement cadence.**~~ **RESOLVED by adversary.** Decrement lives in `violenceUpdate` (3-second PULSE_VIOLENCE cadence), matching C `fight.c:382-430`. NOT in `charUpdate` (70-second cadence).
2. **Slice mutation during iteration.** The decrement loop uses the `kept := ch.Timers[:0]` compact-in-place pattern. Safer than index-delete. Verified idiomatic.
3. **Double-decrement risk.** `violenceUpdate` fires once per PULSE_VIOLENCE per char — no multi-decrement risk. Test covers single-pulse semantics.
4. **Expiry callback for `TIMER_DO_FUN`.** Deferred. `doFun string` parameter is accepted by `AddTimer` and stored on `TimerData`, but NEVER dispatched in this plan — stub. Requires a string-to-function registry (similar to spells). Explicit follow-up; until then, any timer with `Type == TIMER_DO_FUN` expires silently.
5. **Permanent-timer invariant.** `Value == -1` is a universal permanent marker (matches C — `fight.c:74-98` checks `timer->value == -1` without type-checking). G2 loop treats it as type-agnostic.
6. **Test fixture isolation.** `combat_test.go` already uses fresh `CharData` per test (verified — no shared fixtures). G4 tests inherit that isolation.

## Risk

- **Medium.** Touches the combat hot loop (`MultiHit`) and the violence pulse (`violenceUpdate`). Regression could affect combat performance or cause timer leaks. Mitigations: (1) comprehensive unit tests on `handler/timer.go`; (2) integration test covering combat→timer-set→violenceUpdate→timer-expire; (3) mutation testing on the compact-in-place loop (easy to write a subtle off-by-one).
- Adversary correctly flagged that `IsAttackSuppressed` is dead-in-production today. Shipping G3-G4 means the behavior becomes live — watch for surprise test failures in mudprog or combat suites that relied on the always-false behavior.

## Adversary-Resolved Concerns (2026-04-17 plan review)

1. **CRITICAL — cadence bug** (`charUpdate` at 70s vs `violenceUpdate` at 3s, 23x overestimate). Decrement moved to `violenceUpdate`. Plan v1 was wrong.
2. **PLR_NICE timer-set placement bug** (duplicate guards, set-after-damage risk). Fixed: extend the existing PC-vs-PC block at `combat.go:189-193` instead of adding a new block after `oneHit`.
3. **Permanent-timer `Value == -1` invariant** — clarified as UNIVERSAL, not `TIMER_ASUPRESSED`-only. Loop simplified.
4. **`doFun string` stored-but-not-dispatched** — explicit in Open Question 4; stub until string registry exists.
5. **`DoQuit` location** verified at `internal/act/info.go:337`; line-ending fixed to `\n\r`.
6. **Import cycle check** — `combat` already imports `handler`; no new cycle.
