# TODO

## Recommended next steps (ordered)

The four 2026-04-17 audit plans (`plan-combat-depth.md`, `plan-dammessage-gaps.md`, `plan-player-config.md`, `plan-channels.md`) all landed 2026-04-17 — see Done section below.

**Five new plans adversary-verified (2026-04-17 wave 2), ready to execute.** Each has its own doc under `smaug-go/doc/` with gap inventory, task groups, acceptance criteria, and resolved adversary concerns.

1. ~~**`plan-do-channels.md`** (P1) — LANDED 2026-04-17. G1 `Deaf` persistence fix + G2-G5 `DoChannels` command (30-entry toggle table, grouped display, `+all`/`-all`, end-to-end save/load round-trip). See Done below + CHANGELOG entries.~~
2. ~~**`plan-timer-subsystem.md`** (P1) — LANDED 2026-04-18. Generic `handler.AddTimer` / `GetTimer` / `GetTimerPtr` / `RemoveTimer` / `ExtractTimer` / `DecrementTimers` subsystem + wiring into `ViolenceUpdate` (per-char decrement at 3s PULSE_VIOLENCE cadence), `IsAttackSuppressed` refactor, `MultiHit` `TIMER_RECENTFIGHT` set, `DoQuit` gate. Two commits: `8d5675b` (G1) + `062f4a7` (G2-G5).~~
3. **`plan-editor-save.md`** (P2) — Fix `/s` stuck-state bug (blocks `bio`/`description` and Phase-6 OLC substates). Option-C design: call-site assignment of `ch.EditorSave`, no signature churn.
4. **`plan-do-gag.md`** (P2) — Ship `DoGag` standalone toggle mirroring `DoAfk` (conditional pattern, directional messages). ~10 LOC.
5. **`plan-rollD20-seam.md`** (P2) — Convert `rollD20` to function-variable seam; fix ~5% flake in `TestOneHitFull_ExplicitWieldUsed`.

**Still deferred:**

6. **Interactive OLC substates** (P2): `CON_OEDITING` / `CON_MEDITING`. **Unblocked** once `plan-editor-save.md` lands. Harness has the `WithPrompt` seam ready (Tier 5).
7. **Hotboot/copyover** (P2, infrastructure): Go-native design required — C's `exec()` + fd-inheritance doesn't translate. Design pass first (save-all-state + graceful restart + auto-reconnect handshake), then implementation.
8. **Big optional systems** (P3): overland, housing, polymorph, archery, arena, dragon flight, planes, holidays, star maps. Each is large and self-contained; pick by demand signal, not order.

---

## Active

### High-impact combat gaps (from 2026-04-17 audit — P0)

**→ See `smaug-go/doc/plan-combat-depth.md` for the full plan (9 task groups, adversary-verified). LANDED 2026-04-17.**

- [x] Port PC multi-attack skills (second_attack … seventh_attack) into combat loop (plan G1–G5)
- [x] Port weapon proficiency bonus into `OneHit` (plan G6)
- [x] Apply `ch.Stance` in combat — NPC num_attacks stacking, PC GM bonus loop, dam_done/dam_taken multipliers (plan G7)

Follow-ups queued from plan-combat-depth.md:
- [x] `handler.AddTimer` subsystem + `TIMER_RECENTFIGHT` wiring (LANDED 2026-04-18 via plan-timer-subsystem.md)

Follow-ups queued from plan-timer-subsystem.md:
- [ ] `TIMER_DO_FUN` callback dispatch — `AddTimer` accepts a `doFun string` parameter and stores it, but `DecrementTimers` drops the timer silently on expiry without dispatching. Needs a string-to-function registry (analogous to the spell registry) and a hook in the expiry branch.
- [ ] `TIMER_PKILLED` persistence via `PTimer` line in `SavePlayer` (C `save.c:546` load + `save.c:1863` read — plan v1 adversary corrected the file citation from `db.c` to `save.c`). No Go code currently sets `TIMER_PKILLED` so nothing persists; add alongside PKilled mechanics port.
- [ ] `TIMER_NUISANCE` / `TIMER_SHOVEDRAG` wiring — these timer types are defined but no setter or consumer exists (the underlying commands are not ported).
- [ ] Mudprog `timerskilled` / `asupressed` / `pkadrenalized` if-check wiring (`internal/mudprog/ifcheck.go:888`) — subsystem is ready; consumer glue is trivial once the if-check bodies are written.
- [ ] Deity prayer gate on `TIMER_RECENTFIGHT` — C `src/deity.c:1498` blocks prayer when a player recently PK-fought; Go deity system has the hook point but no timer check yet.
- [ ] Wiz-stat display of remaining timer counts (C `src/act_wiz.c:2552-2554`) — useful for immortals debugging timer state.
- [ ] Manager process note: two separate workers inadvertently ran `git checkout -- combat.go` during mutation verification and destroyed uncommitted work. Ban `git checkout` / `git reset --hard` / `git stash` in future worker and adversary prompts; use `Edit` round-trips for mutation verification instead.
- [ ] Devoted-clan favor penalty in `WeaponProfBonusCheck` (C fight.c:1312-1313)
- [ ] Per-round move-cost tracking (C fight.c:1149-1171)
- [ ] `db/system/stances.dat` loader — currently `StanceIndex` is hard-coded in `combat/stance_index.go`
- [ ] PC practice-stance flow — `PCData.Stances[]` counter never increments today, so GM-bonus path is unreachable for existing players
- [ ] Review `DoCircle` (`act/skills3.go:88-99`) and `DoHitall` (`act/skills3.go:294`) for explicit retcode handling now that `OneHit` returns `int`

### Player-visible command gaps (P1)

**→ See `smaug-go/doc/plan-player-config.md` for the player commands (4 groups) and `smaug-go/doc/plan-channels.md` for the channels (5 groups). Both adversary-verified.**

- [x] Register `password`, `title`, `afk`, `save` (plan groups G1–G4). LANDED 2026-04-17.
- [x] Optional `pagelen` alias (plan G5). LANDED 2026-04-17.
- [x] Communication channels: `immtalk`, `gtell`, `auction`-helper + stub (`plan-channels.md` G1–G4). LANDED 2026-04-17. Full auction subsystem (G5 in the plan) + `music`/`newbiechat`/`racetalk`/`wartalk`/`counciltalk`/`guildtalk` deferred to Phase 6 — see Phase 6 candidates below.
- [ ] `bio`/`description` deferred — blocked on `editor.go` `/s` bug where descriptor stays in `CON_EDITING` permanently (`plan-player-config.md` R6).

Follow-ups queued from plan-channels.md:
- [ ] Full `do_auction` state machine (list / bid / stop / noauction list / item escrow / gold handling / auction tick in `update.c`). Plan called out as G5; deferred to Phase 6. C refs: `act_obj.c:3775+`, `update.c:2886-3286`. `BroadcastAuction` helper already in place.
- [ ] `music` / `newbiechat` / `racetalk` / `wartalk` / `counciltalk` / `guildtalk` channel commands — each follows the `DoImmtalk` / `DoGtell` template now that the open-coded pattern is established. Revisit the "factor `talk_channel`?" question once 5+ channels need uniform filtering (see plan-channels.md Open Questions).
- [ ] Per-AT_ color preservation (`AT_IMMORT` / `AT_GTELL` / `AT_GOSSIP` / `AT_CLANTALK`) — currently all channel commands use `DoClantalk`-style `&Y`/`&G`/`&D`; fidelity pass tracked alongside the existing `util.Act` per-recipient color audit finding.
- [ ] Alias prefix-matching (`":hi"` with no space) — `util.OneArgument` splits on whitespace, so a cmdWord of `":hi"` doesn't match the `":"` registration. Deliberate scope cut for P1; a separate interpreter change is required to support intra-token prefix matching.
- [x] `DoChannels` toggle command — LANDED 2026-04-17 (see Done below + `smaug-go/doc/plan-do-channels.md`).

Follow-ups queued from plan-do-channels.md:
- [ ] Per-entry immortal-section trust gates in `DoChannels` no-arg display — currently `channels.go` uses a blanket `IsImmortal()` for the whole Immortal section. C gates `muse` on `sysdata.muse_level`, `log` on `sysdata.log_level`, `high` on `sysdata.think_level`, and `bug` on a hardcoded level 57. Needs a `sysdata` port first.
- [ ] `channels.go:247` `publicAll` slice duplicates part of `channelToggleTable` — derive one from the other when the next channel gets added to the public set, to avoid drift.

Follow-ups queued from plan-player-config.md:
- [ ] R1: `persist/player.go:218` reads `Bio` but nothing writes it — add `Bio` field to the saver when G6 (`bio`/`description`) unblocks.
- [ ] R5: `update_aris` not called before `save_char_obj` in `DoSave` (low-impact for manual save — follow-up).
- [ ] R6: `internal/game/editor.go:156-160` — `/s` does not call `StopEditing`; descriptor stays in `CON_EDITING` forever. Fix required before `DoBio` / `DoDescription` / Phase-6 OLC substates.
- [ ] R7: AFK `[AFK]` indicator on `do_who` listings.
- [ ] R8: audit claim about `ban.go` honoring AFK is incorrect — tracked for audit-doc correction.

### Persistence gaps

- [ ] Add `SaveClan` / `SaveDeity` in `persist/subsystems.go` when the first mutation command ships
- [ ] Persistent clan storeroom contents across reboots (Tier 2 deferral)
- [ ] Per-player "last read" index for note filtering (Tier 2 deferral — C uses `pcdata->last_note`)

### Go idiom polish

- [ ] Replace `math/rand` with `util.NumberRange` in `act/quest.go`, `act/cmds2.go`
- [ ] Change `...interface{}` to `...any` in `testclient/client.go:265`
- [ ] `testclient/client.go:71,84`: switch to `bytes.ToLower`/`bytes.Index` on `[]byte`
- [ ] `testclient/client.go:222`: promote 4KB scratch to a `Client` field

### Mudprog depth gaps (not load-bearing but worth tracking)

From `phase5-tier3-completed.md` known deferrals + code-level TODOs:

- [ ] `OprogCommandTrigger` / `RprogCommandTrigger` — port C `rprog_wordlist_check` wordlist-match logic (`mudprog/oprog.go:216–220`)
- [ ] `mpPeace` per-target argument (currently always room-wide)
- [ ] `OprogDamageTrigger` — C fires once per weapon-damage event; Go fires per worn item
- [ ] `mpsleep` in a false-if branch does not queue (minor divergence)
- [ ] `MPROG_SPEECH` "p " prefix for phrase-match (`mudprog/triggers.go:49`)
- [ ] `mortinroom` / `mortinworld` use exact match (C uses `nifty_is_name` prefix match) — `mudprog/ifcheck.go:803,831`
- [ ] `mpslay` leaves no corpse — decide whether to match C's `raw_kill` which DOES call `make_corpse` (`mudprog/commands.go:386`)
- [ ] `mpmorph` / `mpunmorph` — wire once morph subsystem exists
- [ ] `mphate` — richer hate-list (currently single-slot `Hating`)
- [ ] `mpapply` / `mpapplyb` — auth state machine not ported
- [ ] Sleep queue uses raw mob pointers — a generation counter is the proper defensive fix (`mudprog/sleep.go:61`)
- [ ] `timeskilled`, `leverpos`, `isflagged`, `istagged`, `pkadrenalized`, `asupressed`, `areamulti`, `multi`, `objtype` if-checks (`mudprog/ifcheck.go:885–890`)
- [ ] `util/act.go:35,69` — replace local `actCanSee`/`actCanSeeObj` with a shared canonical port when available

### Combat / damage-message gaps

**→ See `smaug-go/doc/plan-dammessage-gaps.md` for the full plan (3 task groups, adversary-verified). LANDED 2026-04-17.**

- [x] Port `was_in_room` swap (plan G1)
- [x] Port `PCFLAG_GAG` self-suppress (plan G2)
- [x] Port `is_wielding_poisoned` prefix (plan G3) — must check both `WEAR_WIELD` and `WEAR_DUAL_WIELD` for obj identity

Follow-ups queued from plan-dammessage-gaps.md:
- [ ] Port `DoGag` player command (`act_info.c:5795` — ~15 LOC; gag flag is currently test-only)
- [ ] Per-recipient color code preservation (`AT_ACTION` / `AT_HIT` / `AT_HITME`) in `util.Act` — audit-flagged, out of scope of the dam-message gaps but tracked here
- [ ] Fix pre-existing flake in `TestOneHitFull_ExplicitWieldUsed` (`combat_test.go:1981`): `rollD20` has 1/20 chance to roll 0 which always misses regardless of bonuses, causing a 1-in-10 single-call full-suite failure. Not introduced by dam-message gaps (reproduced at HEAD 273039e). Fix by stubbing `rollD20` via the existing `numberPercent` pattern, or by looping until a non-zero roll in the test.

### Usability polish (from audit U1–U3 + phase notes)

- [ ] `PLR_COMPACT` blank-line suppression on descriptor flush (Tier 2 deferral)
- [ ] `XP-on-skill-gain` plumbing (Tier 1 → Tier 4 deferral)
- [ ] "Fully learned" message when a skill reaches its adept cap
- [ ] Per-language phoneme substitution tables (C `LCNV_DATA`) — scrambler is currently a simple rotation

### Documentation correction

- [ ] Reconcile test-count methodology across tier docs (Tier 3 reports 2,205; Tier 4 reports 1,914; actual `func Test*` is 1,962). Pick one methodology, annotate.
- [ ] `game/prompt.go:63` — wire XP-to-next-level into `%x` token (currently prints "0")

---

## Phase 6 candidates (large, self-contained)

Roughly ordered by player-visibility/impact:

### Infrastructure

- [ ] **Hotboot / copyover** — seamless server restart. C uses `exec()` + fd inheritance; Go needs a different design (save state + graceful restart + auto-reconnect handshake). C ref: `src/hotboot.c`.
- [ ] DNS resolution — `net.LookupAddr` for host display
- [ ] Web status page — embedded HTTP server (stdlib)
- [ ] MXP protocol parsing — C `protocol.c` has it; Go has no equivalent

### Game systems

- [ ] **Overland maps** (~3,752 C LOC) — 1000×1000 tile maps, ANSI rendering, landmarks. C: `overland.c`
- [ ] **Player housing** (~2,853 C LOC) — apartments, room customization, guests. C: `house.c`
- [ ] **Polymorph** (~2,753 C LOC) — form shifting with stat mods. MVP stub exists; full system needed. C: `polymorph.c`
- [ ] **Archery** (~1,362 C LOC) — ranged combat, arrow lodging. `DoFire` currently aliases `DoThrow`. C: `archery.c`
- [ ] **Combat stances** (~1,022 C LOC) — 10 fighting postures. Skill stub exists; stance not applied in combat (see P0 above). C: `stances.c`
- [ ] **Dragon flight** (~945 C LOC) — coordinate-based flying. C: `dragonflight.c`
- [ ] **Arena PvP** (~358 C LOC) — challenge/accept isolated combat. C: `arena.c`
- [ ] **Planes** (~298 C LOC) — multi-planar system. C: `planes.c`
- [ ] **Holidays** (~416 C LOC) — calendar events. C: `holidays.c`
- [ ] **Star maps** (~226 C LOC) — celestial display. C: `starmap.c`
- [ ] **Marriage** — C: `marry.c`

### OLC / builder

- [ ] Interactive `CON_OEDITING` / `CON_MEDITING` substates (deferred Tier 4 → Tier 5 → Phase 6)
- [ ] Editable mudprog editors (currently inspector-only)
- [ ] `foldarea` — area vnum repack (low-reward, high-risk — Phase-6 builder tooling)

### Content

- [ ] Skills not yet ported: `bloodlet`, `pounce`, `broach`
- [ ] Clan commands: `promote`, `demote`, `induct`, `outcast`, `bestow`
- [ ] Councils: all commands (not implemented as player-facing yet)
- [ ] Deities: full prayer, favor beyond `mpFavor`, deity-specific effects

### Test infrastructure (nice-to-have)

- [ ] Golden-file diff support in `testclient`
- [ ] CLI scenario runner (`cmd/testclient/main.go`) — scripted YAML playback
- [ ] MCCP client-side zlib reader in testclient (hook already in place)
- [ ] MSDP variable exposure in testclient (currently blind-strips IAC)
- [ ] `cmd/smaug` + `internal/boot` tests migrate to `t.TempDir()` — still share legacy `testdata/player` path
- [ ] End-to-end same-pulse cleanup test driving through `pulse()`
- [ ] `game.drainQueries` dead nil-guard cleanup
- [ ] Nanny-protocol prompt table (centralize exact-string nanny prompts)
- [ ] Mutation-verify the 2026-04-17 batch-1 fixes per CLAUDE.md TDD convention (spot-checked during review; formal mutation pass deferred)

---

## Done

### 2026-04-18

- [x] **Timer subsystem — landed (plan-timer-subsystem.md G1-G7).** Generic `handler.AddTimer` / `GetTimer` / `GetTimerPtr` / `RemoveTimer` / `ExtractTimer` / `DecrementTimers` in `internal/handler/timer.go` (G1, commit `8d5675b`, ~140 LOC + 20 tests). Consumer wiring (G2-G5, commit `062f4a7`): per-char decrement pass at top of `combat.ViolenceUpdate` (ALL chars, not just fighters — matches C `fight.c:382-430` at PULSE_VIOLENCE 3s cadence, NOT PULSE_TICK 70s — plan adversary caught the 23x cadence bug); `IsAttackSuppressed` refactored to `handler.GetTimerPtr` with `Value == -1` OR `Count >= 1` semantics (previously functionally dead — type existed, read path existed, nothing ever wrote); `MultiHit` sets `TIMER_RECENTFIGHT` Count=11 on both PCs in the PC-vs-PC block AFTER the `PLR_NICE` early-return; `DoQuit` gates on the timer with `"Your adrenaline is pumping too hard to quit now!"`. 8 G2-G5 tests + `TestIsAttackSuppressed_CountOneBoundary` to close the `>= 1` boundary gap caught by first adversary. `DecrementTimers` exported (not lowercase per plan prose — caller is `combat` package). `Value == -1` treated as universal permanent-timer sentinel, matches C `fight.c:74-98` generalized. G6 updated `mudprog/ifcheck.go:888` comment to reference the new handler path for future `pkadrenalized`/`asupressed` wiring. `go test -count=3 ./...` green. See plan completion record at the bottom of `smaug-go/doc/plan-timer-subsystem.md`.

### 2026-04-17

- [x] Run `.claude/agents/manager.md` Startup Protocol
- [x] Run full post-Phase-5 audit (4 parallel adversaries: phase completeness, Go idiom, security+usability, C↔Go) — see `smaug-go/doc/audit-2026-04-17.md`
- [x] `doc/phases.md` Phase 3 deliverables rewritten to match what actually shipped
- [x] `OprogCommandTrigger` known-deferral already documented at `phase5-tier3-completed.md:85` — audit correction applied
- [x] **Batch 1 — security + determinism fixes** (commit `9b789a7`):
  - [x] `TestSpellFarsight_Success` deterministic (PC victim bypasses NPC save gate)
  - [x] `mpForce` uses `InterpretWithTrustCap(victim, cmd, mob.GetTrust())`
  - [x] `DoSaveArea` atomic write via tmp + rename
  - [x] `MaxAliases = 50` cap in `DoAlias` create path
  - [x] Boot fail-loud on missing classes/races/skills
  - [x] Brute-force disconnect message wording
  - [x] Positive-direction tests (mpForce + SpellFarsight NPC path) with mutation verification; `savesSpellStaffFn` seam in `magic.go`
- [x] Audit commits landed: `720278d` (docs) + `9b789a7` (code)
- [x] **Player-config commands P1 — landed.** Four commands + one alias (`save` G1, `afk` G2, `title` G3, `password` G4, `pagelen` G5) plus two util prereqs (`util.CaseArgument`, `util.SmashColorToken`). Divergence from C: `do_password` requires `<old> <new> <again>` (C takes only `<new> <again>` with the old-pwd check commented out); min length bumped to 6. `act.BcryptCost` synced from `game.BcryptCost` in `boot.Boot`. New files: `internal/act/playercfg.go`, `internal/act/playercfg_test.go`. 29 new tests (28 in `act` + 1 boot sync test); all mutation-verified. `go test -count=3 ./...` green. See plan completion record at the bottom of `smaug-go/doc/plan-player-config.md`.
- [x] **Communication channels P1 — landed.** Four in-scope groups from `plan-channels.md` (G1 shared infra + PLR_SILENCE opportunistic; G2 DoImmtalk; G3 DoGtell + concurrent-login testclient helper; G4 `BroadcastAuction` + stubbed `DoAuction`). `handler.IsSameGroup` added with flat-graph semantics matching C exactly (no chain-collapse). `PLR_SILENCE` sender gate added to `DoTell` / `DoYell` / `DoShout` / `DoGossip` as a latent-bug fix. `DoImmtalk` does not auto-clear the deaf bit on a deaf sender (matches C: `xREMOVE_BIT` at `act_comm.c:514` is unreachable after the sender returns on `:511`). Boot registers 5 entries (`immtalk`, `:`, `gtell`, `;`, `auction`); alias requires a space after `:`/`;`. Added `Harness.QuickLoginTwo` in `internal/testclient` — the first test helper to drive two simultaneously-logged-in clients. Full auction subsystem, per-AT_ color fidelity, `DoChannels` toggle, and `": hi"`-style intra-token prefix matching deferred (see follow-ups). 26 new tests, every gate mutation-verified. `go test -count=3 ./...` green. See plan completion record at the bottom of `smaug-go/doc/plan-channels.md`.
- [x] **`Deaf` bitvector persistence — bug fix landed (plan-do-channels.md G1).** `SavePlayer` now emits `Deaf <bitvec>` adjacent to `Act`/`AffectedBy`, matching the `LoadPlayer` case already present at `player.go:189`. Previously: `Deaf` was read on load but never written, silently dropping player channel preferences on every logout. Shipped standalone per plan phasing so the persistence fix lands before the `DoChannels` command (G2-G5) that would routinely exercise it. Two new tests in `player_test.go`: round-trip preserves two distinct channel bits; empty `Deaf` is not emitted (forward-compat with existing saves). Mutation-verified. `go test -count=3 ./...` green.
- [x] **`DoChannels` toggle command — landed (plan-do-channels.md G2-G5).** 30-entry `channelToggleTable` in C order with the non-obvious `muse → CHANNEL_HIGHGOD` mapping; NPC + `PLR_SILENCE` gates; C-exact error strings (`"Channels -channel or +channel?\n\r"` / `"Set or clear which channel?\n\r"` / `"Ok.\n\r"`); no-arg grouped display with `+NAME`/`-name` markers and per-section gates (auction > trust 4, clan/council/guild on `PCData` predicates, avatar on Level >= LEVEL_HERO, immortal section on `IsImmortal()`); `+all`/`-all` toggles exactly the C public set (RACETALK, AUCTION, CHAT, QUEST, WARTALK, PRAY, TRAFFIC, MUSIC, ASK, YELL, plus AVTALK gated on `Level >= LEVEL_IMMORTAL` — adversary-corrected from plan v1's incorrect `IS_HERO`). `pray` omitted from display (C commented-out) but toggleable individually. 16 new tests with 5 mutations + adversary-run 6th mutation (all caught); end-to-end `TestDoChannels_RoundTripPersistsViaSave` proves G1+G2 compose through real `persist.SavePlayer` + `LoadPlayer`. Registered in `internal/boot/boot.go` adjacent to Tier 9 channel commands. Files affected: `internal/act/channels.go` (+~200), `internal/act/channels_test.go` (+~290), `internal/boot/boot.go` (+1). `go test -count=3 ./...` green. See plan completion record at the bottom of `smaug-go/doc/plan-do-channels.md`.
