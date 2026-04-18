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
| 5 Tier 6 (combat depth — multi-attack + prof bonus + stance) | `plan-combat-depth.md` (plan + completion) | — |
| 5 Tier 7 (dammessage gaps — was_in_room + PCFLAG_GAG + poison prefix) | `plan-dammessage-gaps.md` (plan + completion) | — |
| 5 Tier 8 (player-config commands — save / afk / title / password + pagelen) | `plan-player-config.md` (plan + completion) | — |
| 5 Tier 9 (communication channels — immtalk / gtell / auction stub) | `plan-channels.md` (plan + completion) | — |
| 5 Tier 10 (Deaf persistence fix + DoChannels toggle command) | `plan-do-channels.md` (plan + completion) | — |
| 5 Tier 11 (Timer subsystem + IsAttackSuppressed wiring + TIMER_RECENTFIGHT + DoQuit gate) | `plan-timer-subsystem.md` (plan + completion) | — |
| 5 Tier 12 (editor `/s` save — `EditorSave` callback + `CON_PLAYING` transition) | `plan-editor-save.md` (plan + completion) | — |
| 5 Tier 13 (`DoGag` standalone toggle) | `plan-do-gag.md` (plan + completion) | — |
| 5 Tier 14 (`rollD20` function-variable seam — eliminates combat test flake) | `plan-rollD20-seam.md` (plan + completion) | — |
| 5 Tier 15 (Tranche A quick wins — `DoBio`/`DoDescription`, `[AFK]` on who, `%X` prompt, retcode guards, testclient idiom polish, publicAll derivation) | — (ran without dedicated plan doc) | CHANGELOG.md 2026-04-18 entry |
| 5 Tier 16 (Tranche B — stances loader + DoMset stance branch + TIMER_DO_FUN callback registry + 7 mudprog if-checks + wordlist match + Oprog/Rprog CMD port) | `plan-tranche-b.md` (plan + completion) | — |
| 5 Tier 17 (Tranche C — quality/fidelity: `util.Act` per-call AType color + 33-site migration, `update_aris` audit correction, `PLR_BLANK` blank-line emission, XP-on-skill-gain + adept-cap message) | `plan-tranche-c.md` (plan + completion) | — |

**Phase 5 is complete (2026-04-18).** All 17 tiers + Tranches A/B/C landed; last commit `26db5f9`. **Phase 6 planning underway** — see `smaug-go/doc/phase6-roadmap.md` for the full candidate inventory, dependency graph, and recommended 6-wave execution order. The roadmap was externally audited 2026-04-18; 21 factual corrections applied (several "missing schema" claims were wrong — `MorphData`, `CharData.X/Y/Map/Sector`, `RoomIndexData.Plane`, `WEAR_MISSILE_WIELD`, `ITEM_PROJECTILE/QUIVER` are already defined).

**Phase 6 plans authored (pending external adversary review before execution):**

| Plan | Status | Scope |
|---|---|---|
| `plan-phase6-arena.md` | audited 2026-04-18 | Arena PvP — 7 groups, 15 criteria. Open Q6 (ROOM_ARENA area-data flag absent on vnums 10366-10382) is a confirmed prereq. |
| `plan-phase6-hotboot.md` | design-exploration only | Wave 0 design doc (not executable). Two Go designs evaluated with PoCs in `/tmp/` (both PASS). Recommends Design A (`syscall.Exec` + FD inheritance, ~65ms pause, seamless). Executable rewrite follows after Open-Q1 (Windows support) resolves. |
| `plan-phase6-starmap.md` | authored | `look sky` port — 3 groups, 13 criteria. Constellation table preserved verbatim from C. |
| `plan-phase6-holidays.md` | authored | Holiday CRUD + persistence + `month_name[]` port — 4 groups, 13 criteria. Fixes 2 latent C bugs (day/month `<=1` rejection; 1-indexed file format). |
| `plan-phase6-marriage.md` | authored | marry/divorce/rings — 7 groups, 12 criteria. Cleans up `PCData.Spouse` orphan; fixes `SavePlayer` Spouse asymmetry. 4 open questions pending human input. |

**Remaining Phase-6 prerequisites:** none blocking Wave 1. All 5 authored plans lack independent-adversary passes (manager subagents in this environment have no `Agent` tool for dispatching adversaries — self-review substituted). Small follow-up items in `TODO.md` Active — deferred mudprog ifchecks (`isflagged`/`istagged`), per-player color customization, `DoBlank` toggle, `DoConfig` surface.

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
