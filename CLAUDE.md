# SMAUG MUD

## Prime Directives (override all other rules)

1. **Assume the manager role defined in `@.claude/agents/manager.md` and follow its Prime Directives in full — they carry equal weight to the ones here.**
2. **On session start, read these foundation docs to resume context:**
   - `@smaug-go/doc/plan.md` — architecture, C→Go mapping, testing strategy
   - `@smaug-go/doc/phases.md` — phase roadmap and per-phase status (includes Phase-6 landed/pending/unauthored board)
   - The most recent phase record (see the index below)
3. **Consult other `smaug-go/doc/` files on demand via the index — do not preload the whole directory.**

---

SMAUG (Simulated Medieval Adventure Multi-User Game) — a text-only MMORPG on the Diku → Merc → SMAUG lineage. Active work is the pure-Go port in `smaug-go/`; original C in `src/` is reference only. Game data lives in `db/`. Current phase status is in `smaug-go/doc/phases.md`.

## Active Work: C → Go port

### Documentation index (`smaug-go/doc/`)

Read foundation docs every session; fetch phase records only when touching that phase; consult review docs on demand.

**Foundation:**

| File | Contents |
|---|---|
| `plan.md` | Architecture, C→Go mapping, design decisions, testing strategy |
| `phases.md` | All phases: deliverables, verification criteria, current status — authoritative Phase-6 board |
| `post-phase6-vision.md` | Post-Phase-6 modernization capture — SOGI gender, poly marriage, Heritage+Community (Daggerheart). **Read before touching `Sex` / `Race` / `Spouse` fields or `util.Act` pronoun tokens.** Captures user-stated direction; not a plan. |

**Phase records:** individual plan and completion docs live under `smaug-go/doc/` as `phaseN-*.md` and `plan-*.md`. Each Phase-6 plan file owns its own §Completion Record with C-bug notes, mutation gates, file deltas, and commit hash — consult on demand via `phases.md` §Phase 6. The `phase6-roadmap.md` holds the full candidate inventory, dependencies, and wave-based execution ordering.

**Reviews (on demand):** `security-review.md`, `go-idiom-review.md`, `audit-2026-04-17.md` — archived audits; load only when re-running a review.

**Current state:** Phase 5 complete 2026-04-18; Phase 6 in flight — see `phases.md` §Phase 6 for the landed / pending / unauthored board and `TODO.md` for small follow-ups.

## Build and run

```bash
cd smaug-go
go build -o smaug-go ./cmd/smaug/
./smaug-go -port 4000 -data ../db
# connect: telnet localhost 4000
```

Module: `github.com/eilidhmae/smaug`. Build all: `go build ./...`. Test all: `go test ./...`.

On fresh clone: `make install-hooks` (from `smaug-go/`) activates the pre-commit gate (gofmt + vet + unit tests + conflict markers). Bypass with `--no-verify`; skip individual gates with `SMAUG_SKIP_VET=1` / `SMAUG_SKIP_TESTS=1`.

## Conventions (not covered in `plan.md`)

- C enums → untyped `int` constants (not typed Go enums) so they interop cleanly with struct fields
- Struct fields use PascalCase; constant names match C exactly for cross-reference (`ACT_IS_NPC`, `PULSE_VIOLENCE`, `ROOM_VNUM_TEMPLE`)
- `types/` has no logic dependencies — everything else may import it
- `world.World` is passed explicitly, never a global. `act.WorldRef` is a package-level set-at-boot exception for commands.
- File loaders: `util.Bug()` and continue on bad data (don't abort — matches C behavior)
- No Cgo. Use pure Go / stdlib equivalents (e.g. `compress/zlib` for MCCP)

## TDD (mandatory)

`plan.md`'s Testing Strategy section points here for workflow details.

1. Write failing test → confirm red
2. Write minimum code to pass → confirm green
3. Refactor → confirm still green
4. Commit

For existing code, backfill tests and verify with **mutation testing**: stage working code (`git add`), mutate the implementation (flip a condition / change a return), run tests → must fail, revert (`git checkout`), run tests → must pass.

Tests live next to code (`foo.go` → `foo_test.go`). Fixtures go in `testdata/`. Aim for high coverage on new/modified code: `go test -cover ./...`.
