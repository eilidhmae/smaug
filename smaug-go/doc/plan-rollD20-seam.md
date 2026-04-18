# Plan: `rollD20` Function-Variable Seam — Eliminate Combat Test Flake

**Status:** Landed 2026-04-18. Adversary verdict PASS. All five acceptance criteria satisfied.
**Priority:** P2 — ~1/20 spurious failure in a single test. Fix is ~10 LOC.
**Scope:** `internal/combat/combat.go`, `internal/combat/combat_test.go`. Optionally `internal/combat/profbonus_test.go`.

---

## Problem

`TestOneHitFull_ExplicitWieldUsed` at `internal/combat/combat_test.go:1981-2029` flakes ~5% of runs. It calls `oneHitFull()` twice with `Hitroll=999` (intended to "always hit"), then asserts `offhandDam > primaryDam`. But `rollD20()` at `combat.go:767-775` can return `0` with probability 1/20, and `combat.go:477` treats `diceroll == 0` as an automatic miss (critical miss — C-faithful per `src/fight.c:1541`).

Failure scenarios:
- Offhand call rolls 0 → `offhandDam = 0` → `0 <= primaryDam` assertion fails (≈4.75%).
- Both roll 0 → both zero → fails (≈0.25%).
- Total: ≈5% per run.

The flake is **not** a production bug. `rollD20() ∈ [0, 19]` and `0 = auto-miss` are intentional and match C behavior.

## C Reference

- `src/fight.c:1538` — `while ((diceroll = number_bits(5)) >= 20);` yields `[0, 19]`.
- `src/fight.c:1541` — miss condition includes `diceroll == 0`.

Go port is faithful. No divergence.

## Existing Seam Pattern

Three function variables already provide RNG/dispatch seams in `combat.go`:

| Line | Variable | Purpose |
|---|---|---|
| 135 | `numberPercent` | Stubbed in tests at 1310, 1571, 1932 |
| 140-142 | `oneHit` | Spied for `MultiHit` cascade |
| 147-149 | `oneHitOffhand` | Swapped for dual-wield tests |

Canonical save/restore pattern (from `stubCascade`, `combat_test.go:1302-1310`):
```go
savedPct := numberPercent
t.Cleanup(func() { numberPercent = savedPct })
numberPercent = pctFn
```

## Design

Convert `rollD20` from a package-local function to a package-local function variable. Single call site (`combat.go:474`) makes this low-risk.

```go
var rollD20 = func() int {
    for {
        v := util.NumberBits(5)
        if v < 20 { return v }
    }
}
```

Tests stub by saving and assigning:
```go
saved := rollD20
t.Cleanup(func() { rollD20 = saved })
rollD20 = func() int { return 10 }  // guaranteed non-crit-miss, non-crit-hit
```

## Task Groups

### G1 — Convert `rollD20` to function variable

- File: `internal/combat/combat.go`, lines 767-775.
- Change `func rollD20() int { ... }` → `var rollD20 = func() int { ... }`.
- Grep to confirm only one call site exists (`combat.go:474`); no change needed at the call site.
- **Test first:** Add `TestRollD20_IsSeam` — save `rollD20`, replace with stub returning 42, call it, assert 42, restore. Mutation-verify: revert to `func` declaration → test fails to compile.

### G2 — Stub `rollD20` in the flaky test

- File: `internal/combat/combat_test.go`, function `TestOneHitFull_ExplicitWieldUsed` at line 1981.
- Add at top of test body:
  ```go
  savedRoll := rollD20
  t.Cleanup(func() { rollD20 = savedRoll })
  rollD20 = func() int { return 10 }  // non-zero, non-19 — normal hit path
  ```
- Replace the existing `ch.Hitroll = 999 // always hit` comment (verified at `combat_test.go:1987`) with: `ch.Hitroll = 999 // large positive hitroll; combined with rollD20 stub above, guarantees hit`.
- **Mutation-verify:** Remove the stub → run `go test -count=100 -run TestOneHitFull_ExplicitWieldUsed ./internal/combat/...` — expect ≥1 failure (false-pass probability at `-count=100` is ≈0.6%; adversary flagged that `-count=30` is 21% false-pass at a 5% flake rate, insufficiently sensitive). Replace stub → run same command — expect 100 passes.

### G3 — Audit adjacent tests for the same flake

- Confirm `TestOneHit_ProfBonus_HigherLearnedDealsMoreDamage` (`profbonus_test.go:254`) is statistically robust (2000 rounds absorb variance; adversary noted the existing comment acknowledges the seam gap).
- **Decision:** Do not stub it. 2000 rounds make the failure probability negligible; stubbing defeats the purpose of a statistical test. Leave a `// Deliberately unstubbed — tests statistical distribution; stubbing would defeat the test.` comment to prevent future confusion.
- Scan every other `_test.go` in `internal/combat/` for calls to `OneHit` / `oneHitFull` that do NOT retry and DO assert on exact damage. Flag each. Expected: none beyond the already-flagged `TestOneHitFull_ExplicitWieldUsed`.

### G4 — Verify fix doesn't regress Hitroll=999 math

- The test relies on `Hitroll=999` producing `thac0 - victimAC` so negative that any non-zero roll hits. Adversary verified: at level 50, `thac0 = -30`; with `Hitroll=999`, `thac0 = -1029`; `victimAC` capped at -19; so `thac0 - victimAC = -1010`; any roll ≥ 1 passes. Stubbing to 10 stays well within this band.

## Acceptance Criteria

A1. `rollD20` is a `var` at `combat.go:767` (or equivalent); tests can save/restore.
A2. `TestOneHitFull_ExplicitWieldUsed` runs 100 times with `go test -count=100` without failure (`-count=30` is insufficiently sensitive — 21% false-pass rate at a 5% flake).
A3. No other combat test regresses (running `go test -count=3 ./internal/combat/...` is green).
A4. `profbonus_test.go:254` comment documents its deliberate unstubbed status.
A5. Production behavior unchanged: `OneHit` / `oneHitFull` use `rollD20` via the var, which points to the same impl.

## Scope Cuts

- **No test reshape.** Do not rewrite the test to loop-until-hit or widen the assertion. The seam approach fixes the root cause.
- **No seam for `numberBits`.** `rollD20` is the right granularity — only combat hit-rolls use this shape.
- **No C parity audit of other RNG uses.** Scope is just this flake.

## Known Constraints

- **`t.Parallel()` incompatible.** The save/restore seam pattern mutates a package-level var. Running tests with `t.Parallel()` concurrently would race. None of the existing seam-using tests (`numberPercent`, `oneHit`, `oneHitOffhand`) are marked `t.Parallel()` — this plan inherits that constraint. Document in a seam-declaration comment so future test authors don't unknowingly break it.

## Open Questions

1. Should the stub be `func() int { return 10 }` or `return 5`? Either works; `10` avoids confusion with the critical-hit value 19. No functional difference.
2. Should `rollD20` be exported (`RollD20`)? No — it's only needed for in-package testing. Keep lowercase.

## Risk

- **Very low.** Function-variable seams are established pattern in this package; zero production-path behavior change; one test changes.

## Adversary-Resolved Concerns (2026-04-17 plan review)

1. **Mutation-verify `-count=30` insufficiently sensitive** — bumped to `-count=100` (false-pass rate 0.6% vs 21% at 30).
2. **"Always hit" comment misleading** — replacement wording specified in G2.
3. **`t.Parallel()` constraint undocumented** — added a "Known Constraints" section.

---

## Completion record (2026-04-18)

Landed as planned. `rollD20` at `internal/combat/combat.go:784` is now a package-local `var = func() int { ... }` rather than `func rollD20() int`, matching the three existing seams in the file (`numberPercent`, `oneHit`, `oneHitOffhand`). Single call site at `combat.go:487` needed no change — identifier resolution is identical. Seam comment on the var declaration documents the no-`t.Parallel()` constraint (Known Constraints § of the plan).

`TestOneHitFull_ExplicitWieldUsed` (`combat_test.go:1981`) now saves `rollD20`, installs a stub returning `10` (normal hit band — non-zero auto-miss, non-19 crit-hit), and restores via `t.Cleanup`. `Hitroll=999` comment updated to the plan-specified wording. A new `TestRollD20_IsSeam` at the end of `combat_test.go` pins the save/restore pattern for future authors.

`profbonus_test.go:255` gained the plan-specified documenting comment (`// Deliberately unstubbed — tests statistical distribution; stubbing would defeat the test.`). The 2000-round `TestOneHit_ProfBonus_HigherLearnedDealsMoreDamage` stays unstubbed on purpose — per adversary analysis, statistical robustness absorbs the 1/20 auto-miss variance.

Worker TDD cycle: `TestRollD20_IsSeam` red before G1 (compile error on `rollD20 = func...` against a `func` declaration); green after conversion. Pre-fix flake observed at 6/100 on one probe run before the stub landed; post-fix `-count=100` runs green deterministically. Mutation verified by changing the stub from `10` to `0` (auto-miss sentinel) with `Edit`: test fails with `offhand dam=0 should exceed primary dam=0`; Edit back to `10`: green.

Adversary PASS. All 5 acceptance criteria satisfied. One pre-existing unrelated flake (`TestViolenceUpdate_DualWield` — does not stub `rollD20`) observed once by the adversary on their first `-count=3` run; five subsequent runs were green, confirming it's a pre-existing statistical flake in an unrelated test, not a regression from this PR. Tracked for a separate future seam application if the flake recurs.

`go build ./...` clean. `go test -count=3 ./...` green across all 15 packages. Not committed yet — this landing batches with `plan-do-gag.md` and `plan-editor-save.md` (three independent plans executed in parallel, one commit each at landing).
