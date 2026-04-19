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
| `plan-phase6-hotboot.md` | audited 2026-04-18 (PASS with notes) | Wave 0 design doc (not executable). Two Go designs evaluated with PoCs in `/tmp/` (both PASS). Recommends Design A (`syscall.Exec` + FD inheritance, ~65ms pause, seamless). 28 C citations and 19 Go-state claims verified; Windows infeasibility strengthened to double-lock (both `syscall.Exec` absent AND `(*net.TCPConn).File()` fd "not usable on other processes" per Go stdlib docs). `syscall.Dup2` (PoC) → `syscall.Dup3(old, new, 0)` needed for linux/arm64. Pre-G1 requirement: wire `HomeVnum` for `ACT_SENTINEL` mobs in `internal/handler/handler.go`. Executable rewrite follows after Open-Q1 (Windows support) resolves. |
| `plan-phase6-starmap.md` | audited 2026-04-18 (PASS) | `look sky` port — 3 groups, 13 criteria. Constellation table preserved verbatim from C. External adversary caught 3 fidelity corrections: critical row-6 transcription error in first fenced block (`"C*"` → `"c."` per starmap.c:66); precip-bounds citation `update.c:3377-3378` → `3446-3452`; `DoWeather` flag-only-vs-flag-OR-sector pattern. Ready to execute. |
| `plan-phase6-holidays.md` | audited 2026-04-18 (CONCERNS) | Holiday CRUD + persistence + `month_name[]` port — 4 groups, 13 criteria (+ A14/A15 pending Q6). Fixes 2 latent C bugs (day/month `<=1` rejection; 1-indexed file format). Audit corrected 2 material C claims: `get_holiday` has 2 live callers in `src/timezone.c` under `#ifdef ENABLE_HOLIDAYS` (plan claimed zero); no `do_load holiday` runtime-reload command exists (fabricated claim removed). New Q6: land `GetHoliday` + `DoTime` holiday-today suffix; season-tick announce deferred pending Go tick-hook grep. |
| `plan-phase6-marriage.md` | audited 2026-04-18 (CONCERNS) | marry/divorce/rings — 5 groups (G1-G5), 12 criteria. Cleans up `PCData.Spouse` orphan; fixes `SavePlayer` Spouse asymmetry. 5 open questions pending human input — critical: `newgate.are` vnum-100/101 collision invalidates original nil-guard strategy for `DoRings`; Q2 resolution broadened to 7 options. Re-audit after Q2 input. |
| `plan-phase6-planes.md` | audited 2026-04-18 (CONCERNS) | Planes CRUD — 3 groups, 13 criteria. Audit caught algorithmic bug in `DoPset delete` slice-splice (broken swap-to-end-then-remove-last — corrected via by-hand 3-index trace) + one factual misstatement (C `one_argument` quote support). All 25 C + 21 Go citations verified. One open follow-up: SmashTilde ordering in rename. |
| `plan-phase6-channels-extra.md` | audited 2026-04-18 (CONCERNS) | `music`/`newbiechat`/`racetalk`/`wartalk`/`counciltalk`/`guildtalk` — factor `talkChannel` helper + 6 thin wrappers. 6 groups, 18 criteria. Audit corrected `util.TranslateFor` misattribution at 5 code sites (actual: `translateFor` package-private in `internal/act/comm.go:13`) and verified `db/councils/council.lst` is empty in stock data — so `newbiechat` is immortal-only by default (matches C). Executable without re-audit. |
| `plan-phase6-skills.md` | audited 2026-04-18 (PASS) | bloodlet/pounce/broach — 3 groups, 16 criteria, parallel-safe. Fixes 2 latent C bugs (bloodlet operator-precedence gate refined to "functionally dead for typical characters" not strictly unreachable; broach predicate inversion) with documented Option-A policy. `gsnPounce` combat-side already shipped; OBJ_VNUM_BLOODLET prototype shipped. 5 open questions (all have recommended answers). One audit follow-up: confirm `combat.Damage(w, ch, ch, ...)` self-target safety before G3. |
| `plan-phase6-stances-olc.md` | audited 2026-04-18 (CONCERNS) | Combat Stances OLC completion — 5 groups (G1-G5), 20 criteria, 10 open questions. Extends `StanceInfo` from 3 to 19 fields; upgrades loader to store non-combat keys; adds `SaveStances` writer; ports `can_use_stance` + `update_stances`; rewrites `DoStance` state machine (picks up 4 currently-unreachable stances: viper/crane/crab/normal); adds `DoSTstat`/`DoSTset` OLC. Preserves 4 C bugs with pin-tests (class/race setter no-ops, dual inverted, class/race `IS_SET(mask, index)`); fixes 1 cosmetic C bug (`do_ststat` newline-every-4). Depends on Tranche B G1 (shipped). **Audit caught 1 correctness-critical finding + 2 citation corrections**: (a) §D3 `UpdateStances` mixes `int` CharData fields with `uint32` RIS_* constants — cast required at fixture/test sites, compile-error risk at G2 without inline note; (b) §Q6 inline RisflagNames draft table stops at "plus7" but `constants.go:196-217` defines 22 bits terminating in `magic`/`paralysis` (bits 20-21); (c) Q9 `stance` registration verified present at `boot.go:521` (POS_DEAD / Level 0). **Q8 resolved**: `World.DataDir` confirmed at `world.go:76`. 4 in-plan surgical edits applied. |
| `plan-phase6-auction.md` | audited 2026-04-18 (CONCERNS) | Full `do_auction` state machine (non-GSC variant). 10 task groups, 23 criteria, 8 open questions. Completes Tier-9 `BroadcastAuction` stub by porting C `act_obj.c:3775-4240` + `update.c:3183-3326`. Adds `pulseAuction` slot; ports `parsebet`/`advatoi`/`num_punct`; extends `AuctionData` with history ring + timer. **Audit corrected 4 factual errors** in 4 in-place edits: (a) anomaly #1 carry-weight check is NOT dead code (seller can acquire weight during 27s window); (b) anomaly #2 C has active `DoQuit` gate at `act_comm.c:2883-2890` (not just `mud.h` comments) — gate wording now ported verbatim; (c) anomaly #14 `ms_find_obj` is a generalized drunk/mental-state check (`handler.c:2941`), not a mud-school quest hook; (d) §D7 C ships `do_noauction` admin command at `act_wiz.c:11120-11174` — load-only port honestly tracked as deferred follow-up. Pre-existing `pulseSave` uninitialized bug at `loop.go:68-94` flagged. |
| `plan-phase6-clan-officer.md` | audited 2026-04-18 (PASS after 3 edits) | Clan officer commands — induct / outcast / bestow + `SaveClan` bundled as G0. 5 task groups (+1 conditional), 14 criteria, 4 open questions. **Scope-correction verified independently**: `do_promote`/`do_demote` do NOT exist in shipped SMAUG (roadmap error at `phase6-roadmap.md:205`); rank management via `do_setclan`, captured as optional G5 (Q1 — human decision pending). Audit applied 3 surgical edits: added `Abbrev` to the loader-missing-keys list (originally omitted); flagged pre-existing legacy-PKills index bug (Go stores to `[0]`, C stores to `[6]` at `fread_clan:435`) for in-plan fix; neutralized one-sided Q1 bias (Option 1 G5-include or Option 2 defer-to-setclan both defensible). Actual C LOC ~478 (roadmap ~300). |

**Remaining Phase-6 prerequisites:** none blocking Wave 1; Wave 2 blockers covered by per-plan open questions. **All Wave A/B/C plans audited as of 2026-04-18** (arena, hotboot, starmap, holidays, marriage, planes, channels-extra, skills, stances-olc, auction, clan-officer — 11 total). Manager subagent harness `Agent`-tool availability remains inconsistent — all 2026-04-18 audits/authorings fell back to structured self-review; tooling investigation tracked in TODO.md. Small follow-up items in `TODO.md` Active — deferred mudprog ifchecks (`isflagged`/`istagged`), per-player color customization, `DoBlank` toggle, `DoConfig` surface, SmashTilde-ordering decision for planes rename, `combat.Damage` self-target safety verification before bloodlet G3.

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
