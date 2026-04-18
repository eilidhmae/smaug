# SMAUG MUD

## Prime Directives (override all other rules)

1. **Assume the manager role defined in `@.claude/agents/manager.md` and follow its Prime Directives in full — they carry equal weight to the ones here.**
2. **On session start, read these foundation docs to resume context:**
   - `@smaug-go/doc/plan.md` — architecture, C→Go mapping, testing strategy
   - `@smaug-go/doc/phases.md` — phase roadmap and overall status
   - The most recent phase record (see the index under *Active Work* below)
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
| `phases.md` | All 5 phases: deliverables, verification criteria, current status |

**Phase records:**

| Phase | Plan / scope | Completion record |
|---|---|---|
| 1 | — | `phase1-completed.md`, `phase1-remaining.md` |
| 2 | `phase2-remaining.md` | `phase2-completed.md` |
| 3 | `phase3-plan.md` | `phase3-completed.md` |
| 4 | `phase4-plan.md` | `phase4a-completed.md`, `phase4b-completed.md` |
| 5 Tier 1 (Act/spell_smaug/saves/persistence) | `phase5-tier1-foundation.md` | `phase5-tier1-completed.md` |
| 5 Tier 2 (flag honoring + idle-data wiring) | `phase5-tier2-wiring.md` | `phase5-tier2-completed.md` |
| 5 Tier 3 (mudprog depth) | `phase5-tier3-mudprog.md` | `phase5-tier3-completed.md` |
| 5 Tier 4 (content breadth) | `phase5-tier4-content.md` (plan + completion) | — |
| 5 Tier 5 (test client + boot extraction) | `phase5-tier5-testclient.md` (plan + completion) | — |

**Reviews:**

| File | Contents |
|---|---|
| `security-review.md` | 16 findings, 14 fixed |
| `go-idiom-review.md` | Go idiom audit: 10 findings, all resolved |
| `audit-2026-04-17.md` | Post-Phase-5 4-adversary audit: phase completeness, idiom, security/usability, C↔Go capability — verdict CONCERNS |

## Build and run

```bash
cd smaug-go
go build -o smaug-go ./cmd/smaug/
./smaug-go -port 4000 -data ../db
# connect: telnet localhost 4000
```

Module: `github.com/eilidhmae/smaug`. Build all: `go build ./...`. Test all: `go test ./...`.

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
