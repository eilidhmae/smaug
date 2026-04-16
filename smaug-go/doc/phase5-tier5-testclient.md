# Phase 5 — Tier 5: Test Client

## Completion (2026-04-15)

All 7 task groups landed 2026-04-15. 15 Go packages pass `go test -count=1 ./...`. The reusable `internal/testclient` harness + the `internal/boot` consolidation are now the canonical entry points for both production server boot and scenario-driven test coverage — `cmd/smaug/main.go` shrunk from ~500 lines to ~70 and the private harness in `cmd/smaug/integration_test.go` shrunk from 373 lines to 151.

- **G1 — Extract boot to a reusable package.** `internal/boot/` now owns `Boot(w, dataDir, incoming, opts)`, which wires all **19** cross-package callbacks identified in the plan (`act.WorldRef`, `act.SaveFunc`, `act.CmdRegistry`, `act.StartEditingFunc`, `act.ShutdownFunc`, `act.DisconnectFunc`, `mudprog.CmdRegistry`, `mudprog.WorldRef`, `combat.WorldRef`, `combat.HitprcntHook`, `combat.VoidHook`, `combat.ObjDamageHook`, `combat.RfightHook`, `combat.DeathRoomHook`, `cmdReg.SocialFallback`, `cmdReg.ObjCommandHook`, `cmdReg.RoomCommandHook`, `persist.SkillNameLookup`, `persist.SkillGetter`). `ProductionOpts()` preserves legacy behavior; `TestOpts()` lowers `game.BcryptCost` to `bcrypt.MinCost` and installs a Shutdowns-channel `ShutdownFunc`. Five tests including the mechanical `TestMainGoHasNoCallbackWires` that reads `cmd/smaug/main.go` as source and regex-rejects any callback assignment outside `boot.Boot`. Production also migrates to `server.StartOnListener(ln net.Listener)`; the racy `Start(port int)` path is retired.
- **G2 — Core `testclient` package.** `internal/testclient/` ships `Harness` (`Start`/`Dial`/`Query`/`Shutdowns`) and `Client` (`Send`/`ReadUntil`/`ReadMatch`/`ReadToPrompt`/`ReadFor`/`WithPrompt`/`Close`). Read pipeline strips telnet IAC sequences (including partial IAC carry across reads), ANSI escapes, and bare CRs. Package-level `harnessMu sync.Mutex` serializes tests to protect the package-global hook vars (`act.WorldRef`, `mudprog.WorldRef`, `combat.*Hook`, etc.) from the `t.Parallel()` hazard flagged in the plan's Open Questions. ~19 unit tests plus 4 harness tests.
- **G3 — Login helpers.** `NewCharacter(t, CharSpec{Name, Password, Sex, Race, Class, Trust})`, `Login(t, name, pwd)`, and `QuickLogin(t, name)` land the client at the in-game prompt. All three canonicalize names (first-letter upper / rest lower) so scenarios can pass lowercase and still round-trip through persistence. 5 login tests.
- **G4 — Minimal fixture data.** `internal/testclient/testdata/` is a ~7KB tree: 1 area file (`tier5test.are`) with 6 rooms (Temple 21001, NORECALL 21002, greet-prog room 21004, banker-keyword room 21005, pebble room 21006), 3 mob templates (basic / greet-prog / banker-by-keyword), 1 pebble object, Warrior class, Human race, ~10 skills, ~5 socials. Boot time ~3ms. Per-test `t.TempDir()` copy of the fixture isolates concurrent packages on disk.
- **G5 — Migrate existing integration tests.** 9 `TestIntegration_*` scenarios preserved (ServerBoot, CharacterCreation, Commands, Communication, InvalidName, BadPassword, Quit, Help, MultipleConnections); harness file dropped from 373 lines of private helpers to 151 lines of calls into `testclient` APIs. `cmd/smaug/main_test.go` rewritten to route through `boot.Boot`.
- **G6 — New scenario tests (one per interactive surface).** Five real-telnet scenarios, each in its own package:
  - `internal/act/olc_scenario_test.go::TestScenario_RedIt_SubcommandRoundTrip`
  - `internal/act/mortal_scenario_test.go::TestScenario_AliasExpansion`
  - `internal/combat/scenario_test.go::TestScenario_MobCreateAndKill`
  - `internal/magic/scenario_test.go::TestScenario_CastSanctuary`
  - `internal/mudprog/scenario_test.go::TestScenario_GreetProgFires`

  The `greet_prog` scenario caught a real dormant wiring bug in `act/info.go::MoveChar` — the greet trigger had been silently unfired since Phase 3. Fix is C-faithful (mirrors `src/act_move.c::move_char`). A second latent bug was fixed in `game.processInput`: a closed `InputQueue` channel previously fell through as a silent no-op; it now marks the descriptor dead so `cleanupDescriptors` can save + remove on the next pulse.
- **G7 — Docs.** This doc (plan kept above, completion appended here), `smaug-go/doc/phases.md` Tier 5 entry, and the `CLAUDE.md` phase-records row all updated in this tier.

### Intentional deviations / gaps

- **ACT_BANKER fixture gap.** The `.are` mob Act-flag parser reads a single 32-bit int, but `ACT_BANKER = 1<<42` is outside that range. The fixture banker mob instead uses the keyword `"banker"` so `DoBank` can find it; the real flag wiring is a separate area-file-format extension. TODO noted in `internal/testclient/testdata_test.go`.
- **`teleportTo` does NOT fire greet.** Matches C `src/build.c::do_goto` (no `mprog_greet_trigger` call there). Only `MoveChar` fires greet, matching C `src/act_move.c::move_char`. The scenario test asserts greet on room movement, not on goto.
- **Interactive OLC substates (`CON_OEDITING` / `CON_MEDITING`) still deferred.** The plan's G6 table explicitly pushed interactive substates to a follow-up; the non-interactive `redit` subcommand path is covered by `TestScenario_RedIt_SubcommandRoundTrip`.
- **`TestSpellFarsight_Success`** was historically RNG-seed flaky; observed stable under `-count=3` after Tier 5 work. Not caused by this tier; monitor.
- **Pre-existing ordering gaps in `MoveChar`** surfaced during greet work: `RprogEnterTrigger` fires before `DoLook` in Go but after in C, and `DoLook` is called with `""` where C uses `"auto"`. Not fixed — pre-existing, out of Tier 5 scope.
- **Shared test-data dir between `cmd/smaug` and `internal/boot`** — both packages run `RemoveAll` on `cmd/smaug/testdata/player` during `go test ./...`, which runs packages in parallel. `testclient` is now isolated via `t.TempDir`; `cmd/smaug` and `internal/boot` still share the legacy path. No currently-observed failures, but a latent race worth flagging.
- **Dead nil guard in `game.drainQueries`.** `if g.queryQueue == nil { return }` is unreachable since `NewGameLoop` always allocates — cosmetic cleanup candidate.
- **Same-pulse cleanup end-to-end untested.** `TestProcessInput_ClosedQueue_MarksDisconnect` verifies `Connected=-1` is set on the closed-queue path, but does not drive through `pulse()` to prove `cleanupDescriptors` saves+removes on the same tick. The unit-level coverage is sufficient for the fix; end-to-end behavior was observed empirically during flake-repro.

### Adversary rounds that mattered

- **G1.** Initial `TestMainGoHasNoCallbackWires` missed 5 of the 19 callbacks; regex was too narrow. Also caught a TOCTOU: `net.Listen(:0)` → `Close` → `Start(port)` has a window where another process can grab the port. Fix was `StartOnListener(ln net.Listener)` as recommended in the plan's Open Questions — the racy port-number API is now retired.
- **G2.** `Harness.Query` hung on shutdown (post-Cancel `Invoke` deadlocked against a drained queue); `WithVerbose` was a no-op on first pass; `ReadMatch` double-evaluated its regex against the live buffer causing duplicate matches. All three fixed before land.
- **G3.** Cross-package fixture contention (`cmd/smaug`, `internal/boot`, `internal/testclient` all wanted to own `testdata/player`); `QuickLogin` failed on all-lowercase names because the nanny stores canonical-case; `testdata_test.go` needed the same `harnessMu` discipline when it booted a Harness to check fixture counts.
- **G4.** `ProgTypes` fixture bit-index was off by one — mob loaded with wrong triggers; greet mob fired on `random` instead of `greet` until corrected.
- **G6.** Scope creep attempt: a first pass of the `greet_prog` scenario tried to route the test through `teleportTo` (faster than `MoveChar`). Rejected — C `do_goto` never fires greet. The scenario now walks the player in via a real direction move, which is what revealed the dormant wiring bug.

### Follow-ups queued for Phase 6

- Interactive `CON_OEDITING` / `CON_MEDITING` substates (deferred again; the harness's `WithPrompt` is ready for them).
- ACT_BANKER area-file format extension (64-bit flag field or overflow byte).
- `MoveChar` ordering parity with C (`RprogEnterTrigger` before `DoLook`, `DoLook` arg `"auto"`).
- Shared `cmd/smaug/testdata/player` race — migrate `cmd/smaug` + `internal/boot` tests to `t.TempDir()` like `testclient` did.
- End-to-end same-pulse cleanup test driving through `pulse()` rather than just the input path.
- `game.drainQueries` dead nil-guard cleanup.
- Nanny-protocol prompt table (flagged in plan Open Questions) — tests still depend on exact-string nanny prompts.
- MCCP / MSDP testclient hooks (plan Open Questions — `Client` has the seam).

## Goal

A reusable in-process test client that drives a real telnet session against an ephemeral game server, used from any `_test.go` in the module. It unblocks the interactive work deferred from Tier 4 (OLC `CON_OEDITING`/`CON_MEDITING` substates, editable mudprog editors) and the Phase 6 candidates (hotboot, overland, housing, stances) where unit tests only prove a function fires — not that a player sees the right thing.

After Tier 5:

1. `internal/testclient` is importable from any test file in the module.
2. Any test can boot a full game server on a random port against a minimal fixture `db/`, connect one or more clients, send scripted input, and assert on captured output.
3. Captured output is clean: telnet IAC bytes and (optionally) ANSI escapes are stripped before matching.
4. Login flows — new character creation and returning-player auth — are single-call helpers.
5. The 9 existing end-to-end tests in `cmd/smaug/integration_test.go` run on the new harness.
6. At least one real-telnet scenario test exists per major interactive surface (`act`, `combat`, `magic`, `mudprog`, `act/olc`) so future regressions surface immediately.

## Baseline (what exists today)

`cmd/smaug/integration_test.go` (373 lines) ships a working but private harness:

- `testServer(t)` — boots a full server on a random port against `testdata/`, wires all callbacks (`act.WorldRef`, `act.SaveFunc`, `mudprog.CmdRegistry`, `act.StartEditingFunc`, …), and registers `t.Cleanup` for loop cancel + player-file removal.
- `mudClient{conn, buf, t}` — raw-byte client with `dial`, `send`, `readUntil` (case-insensitive substring), `readFor` (duration drain).
- `createCharacter(t, port, name, password)` — drives the full new-player nanny walk to the Temple prompt.
- 9 scenarios: server boot, character creation, generic commands, communication, invalid name, bad password, quit, help, multiple connections.

It works. The gaps:

| Gap | Impact |
|-----|--------|
| Lives in `package main` (private to `cmd/smaug`) | Unusable from `act/`, `combat/`, `magic/`, `mudprog/` tests |
| `bootDB` and `registerCommands` are package-private to `cmd/smaug/main.go` | Harness duplicates wiring; drift risk |
| No telnet IAC handling in read path | Any future test asserting on bytes near an IAC sequence will flake; MCCP/MSDP tests are impossible |
| No ANSI strip | Prompt assertions fail because the prompt is colored; most output matching has to be case-insensitive substring to paper over color codes |
| No prompt detection | `readUntil` needs a magic substring for every scenario; tests can't naturally wait for a turn to end |
| No regex matcher | Can't assert on structured output ("damage in range X–Y") |
| Only new-character login helper | Returning-player auth is hand-rolled per test |
| No golden-file support | Scenario diff is per-test, per-hand |
| Fixture data at `cmd/smaug/testdata/` is already trimmed (only `area/classes/races/system/`) but not tuned for scenario tests | Boot is workable, but the fixture lacks a shop, a greet-mudprog mob, and a pickable ground object that G6 scenarios will exercise |

## Task groups

### G1 — Extract boot to a reusable package

Boot wiring currently lives in two places — `cmd/smaug/main.go` (production) and `cmd/smaug/integration_test.go:testServer` (test). Consolidate into `internal/boot/`:

- `Boot(w *world.World, dataDir string, opts BootOpts) (*command.Registry, *game.GameLoop, error)` — loads area/classes/races/skills, registers commands, wires every cross-package callback. Production `main.go` + `bootDB` together wire **19** callbacks: the 15 in the main wiring block plus `act.ShutdownFunc` (`main.go:91`), `act.DisconnectFunc` (`main.go:104`), `persist.SkillNameLookup` (inside `bootDB`, `main.go:185`), `persist.SkillGetter` (`main.go:193`). Full list: `act.WorldRef`, `act.SaveFunc`, `act.CmdRegistry`, `act.StartEditingFunc`, `act.ShutdownFunc`, `act.DisconnectFunc`, `mudprog.CmdRegistry`, `mudprog.WorldRef`, `combat.WorldRef`, `combat.HitprcntHook`, `combat.VoidHook`, `combat.ObjDamageHook`, `combat.RfightHook`, `combat.DeathRoomHook`, `cmdReg.SocialFallback`, `cmdReg.ObjCommandHook`, `cmdReg.RoomCommandHook`, `persist.SkillNameLookup`, `persist.SkillGetter`. The existing `integration_test.go` `testServer` wires only 7 — a latent drift bug G1 also fixes.
- **`BootOpts`** allows overriding the callbacks that call `os.Exit` or close a conn:
  - `ShutdownFunc` — production saves every connected player then `os.Exit(0|2)`. Test default: records the request on a buffered channel (`opts.Shutdowns chan ShutdownRequest`) AND calls `loop.Cancel()` so `gameLoop.Run` returns; the test can drain the channel via `Harness.Shutdowns()` to assert "a shutdown happened". Tests that don't read the channel are fine — the buffered channel absorbs the request and the loop still shuts down cleanly.
  - `DisconnectFunc` — safe in both modes; default behavior closes the conn and is fine for tests. Included in `BootOpts` only for symmetry / future test-only instrumentation.
- Production `cmd/smaug/main.go` calls `boot.Boot(w, dataDir, boot.ProductionOpts())` then `server.StartOnListener(ln)` + `loop.Run`. **Production also migrates to `StartOnListener`** (see Open Questions: TOCTOU) so the racy `Start(port int)` path is fully retired rather than left as dead code.
- Testclient harness calls the same `boot.Boot()` with `boot.TestOpts()`.

Import-graph note: `internal/boot` must import `act`, `combat`, `mudprog`, `persist`, `game`, `command`, `world`, `handler`. To avoid cycles, nothing else may import `internal/boot`. The `BootOpts` type lives in `internal/boot`; callers either construct it via helpers (`ProductionOpts`, `TestOpts`) or accept the exported struct.

**Acceptance:** `cmd/smaug/main.go` post-refactor contains exactly one call to `boot.Boot` and no package-level assignments to `act.*`, `combat.*`, `mudprog.*`, `persist.SkillNameLookup`, `persist.SkillGetter`, or `cmdReg.*` hooks outside `boot.Boot`. Verified mechanically: a test in `boot/boot_test.go` reads `cmd/smaug/main.go` as source and regex-rejects those assignments outside of comments. `main.go` uses `server.StartOnListener` (new API on `internal/net/server.go`); no caller remains on `server.Start(port int)`.

**TDD:** `boot/boot_test.go` — Boot returns a loop that can service a single command via in-memory descriptor; asserts expected area/command counts loaded from a test fixture.

### G2 — Core `testclient` package

Create `internal/testclient/`:

```go
type Harness struct { Port int; /* world/loop are unexported; use Query */ }
func Start(t *testing.T, opts ...Option) *Harness
func (h *Harness) Dial(t *testing.T) *Client
func (h *Harness) Query(fn func(*world.World))  // serialized against the loop goroutine
func (h *Harness) Shutdowns() <-chan ShutdownRequest

type Client struct { ... }
func (c *Client) Send(line string)
func (c *Client) ReadUntil(substr string, timeout time.Duration) string
func (c *Client) ReadMatch(re *regexp.Regexp, timeout time.Duration) []string
func (c *Client) ReadToPrompt(timeout time.Duration) string
func (c *Client) ReadFor(d time.Duration) string
func (c *Client) WithPrompt(pattern string) *Client  // override prompt detection for OLC/editor substates
func (c *Client) Close()
```

`Query` is mandatory for assertions that touch world state (character affects, room descriptions, object inventories). Direct exported `World` access would race with the game loop goroutine. `Query` posts `fn` onto the loop's input queue and blocks until it runs; the loop processes one `fn` per pulse so assertions observe a consistent snapshot.

`WithPrompt` is included in v1 (not deferred) because G6's mudprog scenario and future OLC-substate work both need it. The default pattern is `"> "` after ANSI strip; `redit desc` / string-editor substate uses `"> "` too but emits different surrounding text — callers that care override the pattern explicitly.

Options cover: custom data dir, custom port (default `:0`), disable ANSI strip, disable IAC strip, verbose logging.

Read pipeline: raw bytes → telnet IAC strip (consume IAC … commands silently) → optional ANSI strip → match buffer. Captured output in failure messages always includes the raw buffer so prompt regressions are debuggable.

Prompt detection: look for the last two chars being `"> "` after ANSI strip, or any configurable custom pattern. Configurable because OLC substates will use a different prompt shape.

**Acceptance:** every method has a failing test first, passes with implementation. ANSI and IAC handling verified against synthesized byte streams (no server boot needed for those tests).

**TDD:** `testclient/client_test.go` uses `net.Pipe()` or a throwaway TCP server for Send/ReadUntil/IAC/ANSI tests. `testclient/harness_test.go` verifies Start/Dial/Cleanup against the real game loop.

### G3 — Login helpers

```go
func (h *Harness) NewCharacter(t *testing.T, spec CharSpec) *Client
func (h *Harness) Login(t *testing.T, name, password string) *Client
func (h *Harness) QuickLogin(t *testing.T, name string) *Client  // creates defaults if missing

type CharSpec struct {
    Name, Password string
    Sex, Race, Class string  // defaults: "m", "human", "warrior"
    Trust            int     // 0 = normal mortal; LEVEL_IMMORTAL (54) required for G6 OLC + combat scenarios
}
```

`QuickLogin` is for scenarios that don't care about creation details — it creates with default sex/race/class/password if no save file exists, else logs in. G6's OLC and combat scenarios use `NewCharacter` with `Trust: types.LEVEL_IMMORTAL` because `redit`/`rstat`/`mcreate`/`kill`-on-peaceful all require immortal gates in `registerCommands`. Setting `Trust` after character creation writes directly to `ch.Trust` (bypasses `Level` via `GetTrust()`) and must be persisted into the player save so subsequent `Login` returns with the same trust.

**Acceptance:** both helpers leave the client positioned at the in-game prompt (`ReadToPrompt` succeeds immediately). Returning-player login works after a prior `NewCharacter` + `Close` round-trip (persistence verified).

**TDD:** `testclient/login_test.go` — NewCharacter lands in Temple; Login round-trips after save; QuickLogin both branches.

### G4 — Minimal fixture data

`internal/testclient/testdata/` with enough to boot:

- Single `.are` file: ~6 rooms including Temple (`ROOM_VNUM_TEMPLE`), one NORECALL room, one with a mob carrying a greet mudprog whose `greet_prog` emits a known string (e.g. `"The shopkeeper eyes you warily."`) so G6 can assert on it, one with a mob whose `MobIndexData.Shop` is non-nil and `ACT_BANKER` set (Go has no `ACT_SHOPKEEPER` constant — shops are defined by the `Shop` field), one with a free pickable object (e.g. pebble, for the alias scenario in G6) and a vnum range with headroom so `mcreate 9999` in G6 doesn't collide with fixture mobs.
- Minimal `classes/class.lst` + `warrior.class`, `races/race.lst` + `human.race`.
- `system/en/skills.dat` trimmed to ~10 skills/spells needed by scenario tests (fireball, sanctuary, sleep, bite, cleave, berserk, poison_weapon, meditate, hitall, detrap).
- Empty `helps/`, `boards/`, `player/` (player dir recreated per test).
- Target: boot in <500ms per `go test` invocation (requires bcrypt-cost override; see Open Questions).

**Acceptance:** `go test ./internal/testclient/ -count=1` completes in <5s for the full package test suite on a developer laptop (6 scenario boots × ~500ms + overhead). Revisit if CI proves slower.

**TDD:** `testclient/harness_test.go` asserts the fixture area loads with expected vnums and mob/obj counts.

### G5 — Migrate existing integration tests

Port the 9 `cmd/smaug/integration_test.go` cases onto `internal/testclient` APIs. After migration:

- Delete `mudClient`, `testServer`, `dial`, `createCharacter`, `truncate` from `cmd/smaug/integration_test.go`.
- File either stays (as a thin smoke test using `testclient` + the real `db/`) or is removed entirely. Decide during migration.

**Acceptance:** `go test ./cmd/smaug/... -count=1` still passes AND 9 `Test*` functions remain (in `cmd/smaug/integration_test.go` or migrated into `internal/testclient/smoke_test.go`), each calling `testclient` APIs and covering the same 9 scenarios (ServerBoot, CharacterCreation, Commands, Communication, InvalidName, BadPassword, Quit, Help, MultipleConnections). Dropping any of the 9 is not acceptable — scenario coverage must be preserved.

### G6 — New scenario tests (one per interactive surface)

Proof the harness unblocks work it was built for:

| Package | Scenario | Assert |
|---------|----------|--------|
| `act/olc` | `redit sector 1`, verify via `rstat` | Non-interactive subcommand round-trips. Interactive `desc` (CON_EDITING substate, `olc.go:52` → `editor.go:74`) is deferred until the harness has a per-substate prompt override — add after G2 if needed, or fold into Tier 4 follow-up work |
| `combat` | `mcreate 9999 testmob` (registered `main.go:460`, spawns a live instance in the room), then `kill testmob`, capture damage messages | `DamMessage` output matches expected verb tier |
| `magic` | Cast `sanctuary` on self, inspect affects via `score` (lists active affects) | Affect appears with correct name and duration (no standalone `affect` command exists — use `score` output or read `ch.Affects` directly via the Harness query API, see G2) |
| `mudprog` | Walk into a room with a greet_prog mob | Mob's greet script output is captured on the client |
| `act` (mortal) | `alias g get pebble`, then `g`, pickup of fixture pebble confirmed | Alias expansion works through the interpreter (G4 fixture must include a free object on the ground) |

These are the acceptance tests for the whole tier — if they pass, the harness is load-bearing.

**TDD:** each scenario gets a failing test first. Some will expose harness gaps (e.g., prompt shape in OLC substate differs) — fix forward in G2 options.

### G7 — Docs

- This doc (plan section → replaced with progress section on completion, matching Tier 4 doc shape).
- Update `smaug-go/doc/phases.md`: Phase 5 Tier 5 entry with scope + status.
- Update `CLAUDE.md`: row in the phase-records index table.
- Short "how to write a scenario test" section added to this doc on completion, or folded into `plan.md` Testing Strategy.

## Critical files

**Create:**

- `smaug-go/internal/boot/boot.go` + `boot_test.go` — G1.
- `smaug-go/internal/testclient/client.go` + `client_test.go` — G2.
- `smaug-go/internal/testclient/harness.go` + `harness_test.go` — G2.
- `smaug-go/internal/testclient/login.go` + `login_test.go` — G3.
- `smaug-go/internal/testclient/ansi.go` (or fold into client.go) — G2.
- `smaug-go/internal/testclient/telnet.go` (IAC strip) — G2.
- `smaug-go/internal/testclient/testdata/` fixture tree — G4.
- `smaug-go/internal/act/olc_scenario_test.go`, `smaug-go/internal/act/mortal_scenario_test.go` (both inside `internal/act/` — "act/olc" in G6 is shorthand for the OLC-commands subset of `act`, not a separate package), `smaug-go/internal/combat/scenario_test.go`, `smaug-go/internal/magic/scenario_test.go`, `smaug-go/internal/mudprog/scenario_test.go` — G6.

**Modify:**

- `smaug-go/cmd/smaug/main.go` — delegate boot to `internal/boot`.
- `smaug-go/cmd/smaug/integration_test.go` — migrate to `internal/testclient` (G5); may shrink to a smoke test or be deleted.

**Reference:**

- `smaug-go/cmd/smaug/integration_test.go` — baseline to extract.
- `smaug-go/internal/game/loop.go` — nanny states, prompt construction.
- `smaug-go/internal/net/color.go` — ANSI codes emitted by server (for the strip-table in testclient).
- `smaug-go/internal/net/server.go` — telnet IAC sequences emitted (for the strip table).

## Reused utilities

- Existing `testServer`, `mudClient`, `createCharacter` — starting point for G1/G2/G3 extraction (don't rewrite; port-and-extend).
- `net.Pipe()` — already widely used for per-command output tests; keep for unit tests, add Harness for end-to-end.
- Existing `testdata/` layout in `persist/` tests — style reference for G4 fixtures.

## Verification

1. `go test -count=1 ./...` passes across all 13+ packages.
2. `go test -count=1 ./internal/testclient/...` completes in <5s on a developer laptop (fast feedback loop).
3. 9 migrated integration tests (G5) pass with no behavior change.
4. 5 new scenario tests (G6) pass.
5. `cmd/smaug/main.go` and `cmd/smaug/integration_test.go` share the same boot path (`internal/boot`). Verified by diff — no duplicated callback wiring.
6. Mutation-verify every new function in G1–G4 per CLAUDE.md's TDD mandate (not merely one-per-surface for G6). Break an implementation → test fails → revert → test passes.

## Acceptance criteria (mechanical)

- [ ] `internal/boot.Boot(w, dataDir)` exists; main.go + testclient both call it.
- [ ] `internal/testclient` package exists with Harness, Client, login helpers.
- [ ] Telnet IAC bytes stripped from captured output; verified by a test with synthesized IAC sequences.
- [ ] ANSI escapes stripped when `StripANSI` option is on (default); verified by test.
- [ ] `ReadToPrompt` succeeds after any command that returns to in-game prompt; verified across login, look, score.
- [ ] `ReadMatch(regexp)` returns capture groups; verified by test.
- [ ] 9 `cmd/smaug/integration_test.go` cases pass after migration.
- [ ] 5 new scenario tests pass, one per interactive surface in G6.
- [ ] Harness boots against the minimal fixture in <500ms.
- [ ] `go test ./internal/testclient/... -count=1` under 5s on a developer laptop.

## Open questions / follow-ups

- **Golden files:** deferred. Substring + regex matchers cover the scenarios G6 enumerates. Add golden-file diff later if a scenario genuinely wants a large fixed block.
- **CLI scenario runner:** out of scope for Tier 5. A future `cmd/testclient/main.go` that plays back scripted YAML scenarios could be useful for ad-hoc manual verification, but building it now is scope creep — every scenario the user cares about is a Go test. Revisit for Phase 6 only if a concrete need emerges.
- **IAC handling depth:** v1 strips IAC sequences blindly. When Phase 6 touches MSDP/MSSP tests, extend `testclient/telnet.go` to parse and expose captured variables.
- **MCCP compressed streams:** out of scope for v1 (server emits raw by default). Phase 6 MCCP tests will need a client-side zlib reader; leave the hook in `Client` so it can slot in.
- **Nanny-protocol coupling:** `NewCharacter` drives the creation nanny with exact-string matching at ~7 stages (name → confirm → password → retype → sex → class → race → MOTD). Any change to the nanny's prompts breaks every scenario test silently — they just time out. Consider extracting the nanny steps into a table (`{prompt, response}` pairs) so format drift is localized and obvious. Not a blocker for Tier 5, but file as a follow-up once Tier 5 lands.
- **Fixture drift:** the minimal fixture in G4 will need maintenance when area file format changes. Document the fixture schema so future contributors don't break it silently.
- **Per-test data isolation (filesystem):** G4's `player/` dir cleanup is per-test. Each Harness gets its own `t.TempDir()` with the fixture tree copied (or `go:embed` + extract) so parallel tests don't step on each other's player files. Fixture delivery mechanism (copy vs `go:embed` vs symlink) deliberately left to the G4 worker; all three work, tradeoffs are speed vs simplicity.
- **Per-test data isolation (package globals):** `t.TempDir()` does **not** solve the harder problem — `act.WorldRef`, `mudprog.WorldRef`, `combat.WorldRef`, `act.CmdRegistry`, and the `combat.*Hook` vars are all package-level globals. Two Harness instances running under `t.Parallel()` will race on these. Either disallow `t.Parallel()` inside testclient tests (document + enforce via a serializing mutex in Harness.Start) or redesign the globals — out of scope for Tier 5; recommend the mutex.
- **Port TOCTOU:** the baseline's Listen(:0) → close → `server.Start(port)` pattern (`integration_test.go:49–57`) has a race window where another process can claim the port. `net/server.go:Start(port int)` takes a port number rather than a bound listener. Tier 5 should either (a) add `server.StartOnListener(ln net.Listener)` and use that, or (b) document the race as accepted and retry on bind failure. Recommend (a); it's a tiny addition to `net/server.go` and it's the correct fix.
- **Goroutine cleanup window:** `gameLoop.Run(ctx)` only returns after its next ticker fire post-cancel — up to 250ms. If the next test's Harness.Start begins before the prior loop's goroutine exits, they overlap briefly and share package-global WorldRef state. The serializing mutex above covers this; verify it releases only after `wg.Wait()` returns.
- **Telnet IAC negotiation:** the server declares `telnetEchoOff = IAC WILL ECHO` at `game/loop.go:41` but only *sends* it at the password prompt (`loop.go:278,362,416`), not on initial connect. V1 blind-strips any IAC bytes without replying `IAC DONT ECHO`. Current tests pass because no server-side logic is conditional on the client's DONT response — confirm this stays true after Phase 6 MSDP/MCCP work. If/when the server conditions behavior on the response, extend `testclient/telnet.go` to emit IAC DONT/DO replies.
- **bcrypt cost:** login/NewCharacter pay a full bcrypt hash per call. Calls live in `game/loop.go:310,389,414` (`GenerateFromPassword`, `CompareHashAndPassword`, `DefaultCost`). Not parameterized today. If G4's <3s target is tight in CI, expose a package-level `BcryptCost` var in `game/loop.go` defaulting to `bcrypt.DefaultCost` and override to `bcrypt.MinCost` (4 → ~100× faster) in `TestOpts()`. Add as an explicit prerequisite of G4 if target misses.
