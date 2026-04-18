# TODO

## Recommended next steps (ordered)

The four 2026-04-17 audit plans (`plan-combat-depth.md`, `plan-dammessage-gaps.md`, `plan-player-config.md`, `plan-channels.md`) all landed 2026-04-17 — see Done section below.

**Five 2026-04-17 wave-2 plans all landed.** Each has its own doc under `smaug-go/doc/` with gap inventory, task groups, acceptance criteria, and resolved adversary concerns.

1. ~~**`plan-do-channels.md`** (P1) — LANDED 2026-04-17. G1 `Deaf` persistence fix + G2-G5 `DoChannels` command (30-entry toggle table, grouped display, `+all`/`-all`, end-to-end save/load round-trip). See Done below + CHANGELOG entries.~~
2. ~~**`plan-timer-subsystem.md`** (P1) — LANDED 2026-04-18. Generic `handler.AddTimer` / `GetTimer` / `GetTimerPtr` / `RemoveTimer` / `ExtractTimer` / `DecrementTimers` subsystem + wiring into `ViolenceUpdate` (per-char decrement at 3s PULSE_VIOLENCE cadence), `IsAttackSuppressed` refactor, `MultiHit` `TIMER_RECENTFIGHT` set, `DoQuit` gate. Two commits: `8d5675b` (G1) + `062f4a7` (G2-G5).~~
3. ~~**`plan-editor-save.md`** (P2) — LANDED 2026-04-18. `EditorSave` callback field on `CharData`, Option-C call-site assignment in `DoRedit`, `/s` handler transitions to `CON_PLAYING` + one-shot callback invoke, `StopEditing` defensive nil. Unblocks `DoBio` / `DoDescription` and Phase-6 interactive OLC substates.~~
4. ~~**`plan-do-gag.md`** (P2) — LANDED 2026-04-18. `DoGag` standalone toggle mirroring `DoAfk` with directional messages; 4 mutation-verified tests; `POS_DEAD` / level 0; citation corrections in TODO.md + `plan-dammessage-gaps.md`.~~
5. ~~**`plan-rollD20-seam.md`** (P2) — LANDED 2026-04-18. `rollD20` converted to function-variable seam; `TestOneHitFull_ExplicitWieldUsed` now stubs to return `10` (normal hit band); deliberate-unstubbed comment on the 2000-round statistical test in `profbonus_test.go`. 100-count runs deterministic.~~

**Still deferred:**

6. **Interactive OLC substates** (P2): `CON_OEDITING` / `CON_MEDITING`. **Unblocked** now that `plan-editor-save.md` landed. Harness has the `WithPrompt` seam ready (Tier 5).
7. **`DoBio` / `DoDescription`** (P2): **Unblocked** now that `plan-editor-save.md` landed — wire via the `EditorSave` callback pattern demonstrated in `DoRedit desc` / `ed`.
8. **Hotboot/copyover** (P2, infrastructure): Go-native design required — C's `exec()` + fd-inheritance doesn't translate. Design pass first (save-all-state + graceful restart + auto-reconnect handshake), then implementation.
9. **Big optional systems** (P3): overland, housing, polymorph, archery, arena, dragon flight, planes, holidays, star maps. Each is large and self-contained; pick by demand signal, not order.

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
- [x] `TIMER_DO_FUN` callback dispatch (LANDED 2026-04-18 via plan-tranche-b.md G2). Registry + expiry dispatch with substate save/restore; mid-decrement + interp intercepts deferred until first skill command ports.
- [ ] `TIMER_PKILLED` persistence via `PTimer` line in `SavePlayer` (C `save.c:546` load + `save.c:1863` read — plan v1 adversary corrected the file citation from `db.c` to `save.c`). No Go code currently sets `TIMER_PKILLED` so nothing persists; add alongside PKilled mechanics port.
- [ ] `TIMER_NUISANCE` / `TIMER_SHOVEDRAG` wiring — these timer types are defined but no setter or consumer exists (the underlying commands are not ported).
- [x] Mudprog `timerskilled` / `asupressed` / `pkadrenalized` if-check wiring — LANDED 2026-04-18 via plan-tranche-b.md G3 (note: C name is `timeskilled` without the 'r' — external adversary caught this typo).
- [ ] Deity prayer gate on `TIMER_RECENTFIGHT` — C `src/deity.c:1498` blocks prayer when a player recently PK-fought; Go deity system has the hook point but no timer check yet.
- [ ] Wiz-stat display of remaining timer counts (C `src/act_wiz.c:2552-2554`) — useful for immortals debugging timer state.
- [ ] Manager process note: two separate workers inadvertently ran `git checkout -- combat.go` during mutation verification and destroyed uncommitted work. Ban `git checkout` / `git reset --hard` / `git stash` in future worker and adversary prompts; use `Edit` round-trips for mutation verification instead.
- [ ] Devoted-clan favor penalty in `WeaponProfBonusCheck` (C fight.c:1312-1313)
- [ ] Per-round move-cost tracking (C fight.c:1149-1171)
- [x] `db/system/stances.dat` loader — LANDED 2026-04-18 via plan-tranche-b.md G1. `internal/persist/stances.go`; wired in `boot.bootDB`. Non-combat fields read-and-discard pending Phase-6 `StanceInfo` extension.
- [x] PC practice-stance flow — LANDED 2026-04-18 via plan-tranche-b.md G1b. `mset <victim> <stance-name> <value>` now seeds PC mastery per C `build.c:3499-3537`. No player-facing grind command (C has none either — deliberate fidelity).
- [x] Review `DoCircle` (`act/skills3.go:88-99`) and `DoHitall` (`act/skills3.go:294`) for explicit retcode handling now that `OneHit` returns `int`. **LANDED 2026-04-18 (Tranche A item 4).** Added `combat.AttackerDied(int) bool` / `combat.VictimDied(int) bool` exported helpers; `DoHitall` now breaks on `combat.AttackerDied(ret) || ch.Position <= POS_DEAD`; `DoCircle` now guards the second swing on `!VictimDied(ret) && !AttackerDied(ret) && victim.Position > POS_DEAD && ch.Position > POS_DEAD`. Attacker-death path is dormant (no fireshield/ice_shield/acid_shield yet) — defense-in-depth for when reactive damage ships. Helper predicates mutation-verified via combat_test.go.

### Player-visible command gaps (P1)

**→ See `smaug-go/doc/plan-player-config.md` for the player commands (4 groups) and `smaug-go/doc/plan-channels.md` for the channels (5 groups). Both adversary-verified.**

- [x] Register `password`, `title`, `afk`, `save` (plan groups G1–G4). LANDED 2026-04-17.
- [x] Optional `pagelen` alias (plan G5). LANDED 2026-04-17.
- [x] Communication channels: `immtalk`, `gtell`, `auction`-helper + stub (`plan-channels.md` G1–G4). LANDED 2026-04-17. Full auction subsystem (G5 in the plan) + `music`/`newbiechat`/`racetalk`/`wartalk`/`counciltalk`/`guildtalk` deferred to Phase 6 — see Phase 6 candidates below.
- [x] `bio`/`description` — **LANDED 2026-04-18 (Tranche A item 1).** `DoBio` sets `SUB_PERSONAL_BIO` and wires an `EditorSave` closure that writes to `PCData.Bio` on `/s`; `DoDescription` does the same for `ch.Description`. NPC + nil-PCData + PCFLAG_NOBIO/NODESC guards. Registered in `internal/boot/boot.go` at `POS_DEAD`/level 0. `Bio` field now emitted by `SavePlayer` (closing R1 follow-up). E2E test in `internal/testclient/bio_test.go` drives `bio` → editor → `/s` → `save` → `quit` → relogin → verify preserved.

Follow-ups queued from plan-channels.md:
- [ ] Full `do_auction` state machine (list / bid / stop / noauction list / item escrow / gold handling / auction tick in `update.c`). Plan called out as G5; deferred to Phase 6. C refs: `act_obj.c:3775+`, `update.c:2886-3286`. `BroadcastAuction` helper already in place.
- [ ] `music` / `newbiechat` / `racetalk` / `wartalk` / `counciltalk` / `guildtalk` channel commands — each follows the `DoImmtalk` / `DoGtell` template now that the open-coded pattern is established. Revisit the "factor `talk_channel`?" question once 5+ channels need uniform filtering (see plan-channels.md Open Questions).
- [x] Per-AT_ color preservation in `util.Act` — LANDED 2026-04-18 via plan-tranche-c.md G1-G3. `util.Act` now takes a per-call `aType int` that drives `&X` color prefix + `&D` reset; all 33 callers migrated. Channel commands themselves still use inline `&Y`/`&G`/`&D` (call `ch.Send` directly, not `util.Act`) — factor-out is a Phase-6 follow-up when harmonizing channel coloring. `AT_CLANTALK` does NOT exist in C and was NOT added.
- [ ] Alias prefix-matching (`":hi"` with no space) — `util.OneArgument` splits on whitespace, so a cmdWord of `":hi"` doesn't match the `":"` registration. Deliberate scope cut for P1; a separate interpreter change is required to support intra-token prefix matching.
- [x] `DoChannels` toggle command — LANDED 2026-04-17 (see Done below + `smaug-go/doc/plan-do-channels.md`).

Follow-ups queued from plan-do-channels.md:
- [ ] Per-entry immortal-section trust gates in `DoChannels` no-arg display — currently `channels.go` uses a blanket `IsImmortal()` for the whole Immortal section. C gates `muse` on `sysdata.muse_level`, `log` on `sysdata.log_level`, `high` on `sysdata.think_level`, and `bug` on a hardcoded level 57. Needs a `sysdata` port first.
- [x] `channels.go:247` `publicAll` slice duplicates part of `channelToggleTable` — derive one from the other when the next channel gets added to the public set, to avoid drift. **LANDED 2026-04-18 (Tranche A item 6).** Added `publicAll bool` field to `channelToggleTable` entries; `+all`/`-all` handler iterates the table filtered on `publicAll == true`. AVTALK's level-gate remains separately applied (immortal-only). Tests verify derivation (adding a fake publicAll entry auto-enrolls) and the negative direction (private channels stay untouched).

Follow-ups queued from plan-player-config.md:
- [x] R1: `persist/player.go:218` reads `Bio` but nothing writes it — **LANDED 2026-04-18 (Tranche A item 1).** `SavePlayer` now emits `Bio      <text>~` when non-empty, gated through `util.SmashTilde`. Round-trip test + empty-bio-not-emitted test in `internal/persist/player_test.go`.
- [x] R5: ~~`update_aris` not called before `save_char_obj` in `DoSave`~~ — **Audit-corrected 2026-04-18 via plan-tranche-c.md G5.** Go uses incremental `handler.AffectModify` (internal/handler/handler.go:386-455) rather than C's rebuild-from-scratch `update_aris`; `AffectToChar` / `AffectRemove` adjust `ch.Hitroll`/`Damroll`/`Armor` inline so `SavePlayer` writes correct values verbatim. Architecturally not applicable. Regression guard: `internal/persist/player_affect_test.go:TestSaveLoadPlayer_AffectRemovalMaintainsStatInvariant`. See also `plan-player-config.md:112`.
- [x] R6: `internal/game/editor.go:156-160` — `/s` does not call `StopEditing`; descriptor stays in `CON_EDITING` forever. **LANDED 2026-04-18 via `plan-editor-save.md`** (Option-C callback pattern).
- [x] R7: AFK `[AFK]` indicator on `do_who` listings. **LANDED 2026-04-18 (Tranche A item 2).** `DoWho` now prepends `"[AFK] "` to PCs with `PLR_AFK` set (C `act_info.c:3686/4306`). Mutation-verified both directions.
- [ ] R8: audit claim about `ban.go` honoring AFK is incorrect — tracked for audit-doc correction.

### Persistence gaps

- [ ] Add `SaveClan` / `SaveDeity` in `persist/subsystems.go` when the first mutation command ships
- [ ] Persistent clan storeroom contents across reboots (Tier 2 deferral)
- [ ] Per-player "last read" index for note filtering (Tier 2 deferral — C uses `pcdata->last_note`)

### Go idiom polish

- [ ] Replace `math/rand` with `util.NumberRange` in `act/quest.go`, `act/cmds2.go`
- [x] Change `...interface{}` to `...any` in `testclient/client.go:265`. **LANDED 2026-04-18 (Tranche A item 5).**
- [x] `testclient/client.go:71,84`: switch to `bytes.ToLower`/`bytes.Index` on `[]byte`. **LANDED 2026-04-18 (Tranche A item 5).**
- [x] `testclient/client.go:222`: promote 4KB scratch to a `Client` field. **LANDED 2026-04-18 (Tranche A item 5).** Lazy-init on first `fillOnce`; mutation-verified (per-call alloc → test fails).

### Mudprog depth gaps (not load-bearing but worth tracking)

From `phase5-tier3-completed.md` known deferrals + code-level TODOs:

- [x] `OprogCommandTrigger` / `RprogCommandTrigger` — LANDED 2026-04-18 via plan-tranche-b.md G4. Full word-boundary algorithm + `"p "` phrase prefix in `internal/mudprog/wordlist.go`; Oprog iterates room-floor only per C.
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
- [ ] `isflagged` / `istagged` mudprog if-checks — port C `get_tag` at `src/mud_prog.c:2236+` + `VariableData` lookup by name/vnum (~60 LOC + tests). Scope cut from Tranche B.
- [ ] `TIMER_DO_FUN` mid-decrement intercept (C `fight.c:386-398`) — combat-aborts-skill; ports with the first skill command (`do_detrap` / `do_dig` / `do_search` / `do_mend` / `do_reading` / `do_cast`).
- [ ] `TIMER_DO_FUN` interp intercept (C `interp.c:713-733`) — new-command-aborts-skill; ports with the first skill command.
- [ ] When the first skill command (`do_detrap`, etc.) ports, register it via `handler.RegisterTimerFunc("do_detrap", act.DoDetrap)` in `boot.Boot`. The 6 canonical C names are listed in the new `boot.go` comment block.
- [ ] Switch `MPROG_SPEECH` / `SPEECHIW` / `TELL` triggers to the new `wordlistMatch` helper — currently `internal/mudprog/triggers.go:49` has an incomplete `"p"`-skip that over-fires. Small follow-up after auditing the existing `triggerMatches` `"p"` handling.
- [ ] Stance-table OLC (full `do_stset` with field editing, `do_ststat`, `fwrite_stance`) — Phase 6 OLC. Tranche B G1 loader read-and-discards non-combat fields (class/race restrictions, immune/resist/suscept, dodge/parry, max_weight, dual_wield, wait, prerequisite stance[]) pending `StanceInfo` extension.
- [ ] `can_use_stance` prerequisite checks in `DoStance` — depends on Phase-6 `StanceInfo` extension.
- [ ] Class/race stance restrictions + `max_weight` / `dual_wield` restrictions — same prerequisite.
- [ ] Extend `internal/util/act.go:atColorCode` table to additional AT_* codes (AT_DAMAGE, AT_FLEE, AT_STANCE, AT_CARNAGE, AT_HURT, AT_DYING, etc.) as Phase-6 content or commands using them land. Unknown AT values currently emit no color prefix — safe default.
- [ ] Per-player `PCData.Colorize[]` customization — C's `set_char_color` honors per-player overrides; Go has no `Colorize` field today. Phase-6 config item.
- [ ] `DoBlank` / `DoConfig +blank` toggle command — `PLR_BLANK` rendering landed (plan-tranche-c.md G6) and the flag persists via `ch.Act`, but there is no user-facing command to set/clear it. Mirror `DoGag` / `DoAfk` pattern.
- [x] `timeskilled`, `leverpos`, `pkadrenalized`, `asupressed`, `areamulti`, `multi`, `objtype` if-checks — LANDED 2026-04-18 via plan-tranche-b.md G3 (7 of 9). `isflagged` / `istagged` remain; need variable-subsystem port of C `get_tag` at `src/mud_prog.c:2236+`.
- [ ] `util/act.go:35,69` — replace local `actCanSee`/`actCanSeeObj` with a shared canonical port when available

### Combat / damage-message gaps

**→ See `smaug-go/doc/plan-dammessage-gaps.md` for the full plan (3 task groups, adversary-verified). LANDED 2026-04-17.**

- [x] Port `was_in_room` swap (plan G1)
- [x] Port `PCFLAG_GAG` self-suppress (plan G2)
- [x] Port `is_wielding_poisoned` prefix (plan G3) — must check both `WEAR_WIELD` and `WEAR_DUAL_WIELD` for obj identity

Follow-ups queued from plan-dammessage-gaps.md:
- [x] Port `DoGag` player command (`act_info.c:5585 do_config (gag branch at :5794)` — ~15 LOC). **LANDED 2026-04-18 via `plan-do-gag.md`** — standalone toggle; flag no longer test-only.
- [x] Per-recipient color code preservation (`AT_ACTION` / `AT_HIT` / `AT_HITME`) in `util.Act` — LANDED 2026-04-18 via plan-tranche-c.md G1-G3. Note: C applies the SAME `AType` to every recipient of one call (per-call, not per-recipient); two-call pattern (TO_CHAR with AT_HIT + TO_VICT with AT_HITME) matches C `fight.c:4586-4588` and now works as intended.
- [x] Fix pre-existing flake in `TestOneHitFull_ExplicitWieldUsed` (`combat_test.go:1981`). **LANDED 2026-04-18 via `plan-rollD20-seam.md`** — `rollD20` is now a function-variable seam; the flaky test stubs it to return `10`. `go test -count=100` deterministic.

### Usability polish (from audit U1–U3 + phase notes)

- [x] `PLR_BLANK` blank-line emission on descriptor flush — LANDED 2026-04-18 via plan-tranche-c.md G6. (Original entry mislabeled as `PLR_COMPACT`; C uses `PLR_BLANK`.) No user-facing toggle shipped; `DoBlank` follow-up queued below.
- [x] `XP-on-skill-gain` plumbing — LANDED 2026-04-18 via plan-tranche-c.md G7. 20×skLvl normal (×6 mage, ×3 cleric); silent during combat / `gsnHide` / `gsnSneak`.
- [x] "Fully learned" message when a skill reaches its adept cap — LANDED 2026-04-18 via plan-tranche-c.md G7. `"&WYou are now an adept of %s! You gain %d bonus experience!\n\r&D"`; 1000×skLvl XP (×5 mage, ×2 cleric); mutually exclusive with normal-gain branch.
- [ ] Per-language phoneme substitution tables (C `LCNV_DATA`) — scrambler is currently a simple rotation

### Documentation correction

- [ ] Reconcile test-count methodology across tier docs (Tier 3 reports 2,205; Tier 4 reports 1,914; actual `func Test*` is 1,962). Pick one methodology, annotate.
- [x] `game/prompt.go:63` — wire XP-to-next-level into `%x` / `%X` tokens. **LANDED 2026-04-18 (Tranche A item 3).** `%x` already printed current Exp; `%X` was the stubbed-zero — now returns `exp_level(ch, level+1) - ch.Exp` via a new `game.PromptExpBase` seam (wired in boot from `w.Classes[ch.Class].ExpBase`; NPCs use 1000 per C `handler.c:107-112`). Clamped non-negative. Seam + clamp + boot-wire tests in `internal/game/prompt_test.go` and `internal/boot/boot_test.go`.

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

- [x] **Editor `/s` fix — landed (plan-editor-save.md G1-G6).** Option C call-site-assignment pattern: `CharData.EditorSave func(*CharData)` field set by callers **before** invoking `StartEditingFunc`, `/s` handler now transitions `ch.Desc.Connected → CON_PLAYING` first, one-shot-clears `ch.EditorSave` before invoking it, and leaves cleanup to the callback (matches C `build.c:7004-7010`). `StopEditing` defensively nils the callback. Both `DoRedit desc` and `DoRedit ed` call sites in `internal/act/olc.go` set a closure that captures the target pointer (room or `*ExtraDescrData`), calls `CopyBufferFunc(ch)` to extract text, writes to the target, and calls `StopEditingFunc(ch)` + `ch.Send("\n\r")` for prompt spacing. Two new seams `CopyBufferFunc` / `StopEditingFunc` were required because `act` cannot import `game` (circular dep) — they follow the same pattern as the pre-existing `StartEditingFunc`. Boot wiring + nil-reset + post-boot non-nil assertions all in place; the post-boot assertions were a first-pass adversary finding (gap would have let silent wiring regressions through). Pre-fix bug: any builder who used `redit desc` would get permanently wedged in `CON_EDITING` with every subsequent line routed back to the editor buffer. 9 new tests across `internal/game/editor_test.go` (6 including a rewritten `TestEditBuffer_SaveCommand` stub) and `internal/act/olc_test.go` (3 incl. a round-trip that manually simulates `/s` because of the `act → game` import prohibition). 7 mutations exercised (field removal, transition delete, invocation delete, one-shot clear removal, wrong-target assign, both wire removals) — all caught, all Edit-based reverts, no destructive git. `StartEditing` / `StartEditingFunc` signatures unchanged; `boot_test.go:422` regex assertion stays green untouched. Unblocks `DoBio` / `DoDescription` and Phase-6 interactive OLC substates. Files affected: `internal/types/character.go` (+1 field + field test), `internal/game/editor.go` (+~14 LOC `/s` handler + `StopEditing` nil), `internal/game/editor_test.go` (+~200 LOC), `internal/act/olc.go` (+2 seams + 2 closure blocks), `internal/act/olc_test.go` (+~80 LOC), `internal/boot/boot.go` (+2 wires), `internal/boot/boot_test.go` (+2 nil-resets + 2 post-boot assertions). `go test -count=3 ./...` green. See plan completion record at the bottom of `smaug-go/doc/plan-editor-save.md`.

- [x] **`DoGag` — landed (plan-do-gag.md G1-G4).** Standalone no-arg toggle command in `internal/act/playercfg.go` mirroring `DoAfk` exactly: `IsNPC` gate, load-bearing nil-`PCData` guard (`IsNPC()` checks `Act.ACT_IS_NPC` only — not `PCData`), conditional `if IsSet { Remove; "Combat messages will no longer be gagged.\n\r"; return }; Set; "Combat messages will be gagged.\n\r"`. NOT XOR (plan caught XOR mutation blind spot). Line endings `\n\r` matching every other `ch.Send` in the file. Registered in `internal/boot/boot.go` adjacent to `afk`: `POS_DEAD`, level 0. 4 tests mutation-verified (TogglesOn, TogglesOff, NPCIsNoop, NilPCDataIsNoop); full mutation matrix from plan § G3 exercised (`|=` → `&^=`, `&^=` → `|=`, remove IsNPC guard, remove nil guard, swap messages) — all caught via Edit round-trips. Persistence round-trip works with no new code (`persist/player.go:243/520` already round-trips `PCData.Flags`). Doc citation corrections: `TODO.md` + `plan-dammessage-gaps.md:87` changed from `act_info.c:5795` to `act_info.c:5585 do_config (gag branch at :5794)` (C has no standalone `do_gag` — adversary plan-review correctly flagged the citation mislabel). Stale follow-up prose in `plan-dammessage-gaps.md:87` also refreshed — `DoGag` is no longer a follow-up item. Files affected: `internal/act/playercfg.go` (+~25 LOC), `internal/act/playercfg_test.go` (+4 tests), `internal/boot/boot.go` (+1 register). `go test -count=3 ./...` green. See plan completion record at the bottom of `smaug-go/doc/plan-do-gag.md`.

- [x] **`rollD20` seam — landed (plan-rollD20-seam.md G1-G4).** `internal/combat/combat.go:784` converted from `func rollD20() int` to `var rollD20 = func() int { ... }`, joining the three existing function-variable seams in the file (`numberPercent`, `oneHit`, `oneHitOffhand`). Zero production behavior change — single call site at `combat.go:487` resolves the same identifier. Seam doc comment records the no-`t.Parallel()` constraint. `TestOneHitFull_ExplicitWieldUsed` now saves `rollD20`, installs a stub returning `10` (normal hit band — non-zero auto-miss, non-19 crit-hit), and restores via `t.Cleanup`. `Hitroll=999` comment refreshed to plan-specified wording. `TestRollD20_IsSeam` pins save/restore pattern for future authors. `profbonus_test.go:255` gained a documenting comment explaining why `TestOneHit_ProfBonus_HigherLearnedDealsMoreDamage` is deliberately unstubbed (2000-round statistical test — stubbing would defeat its purpose; variance is absorbed). Worker mutation-verified by changing stub from `10` to `0` (auto-miss sentinel): test fails `offhand dam=0 should exceed primary dam=0`, Edit back to `10` → green. `go test -count=100 -run TestOneHitFull_ExplicitWieldUsed ./internal/combat/...` — 100/100 green. Pre-fix flake rate observed at ~5-6%. Files affected: `internal/combat/combat.go` (func → var), `internal/combat/combat_test.go` (+stub + `TestRollD20_IsSeam`), `internal/combat/profbonus_test.go` (+1 comment). `go test -count=3 ./...` green. See plan completion record at the bottom of `smaug-go/doc/plan-rollD20-seam.md`.

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
