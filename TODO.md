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

6. **Interactive OLC substates** (P2): `CON_OEDITING` / `CON_MEDITING`. **`CON_REDIT` LANDED 2026-04-19** (`plan-phase6-olc-redit.md`) — nanny-dispatch pattern + `OlcData`-on-descriptor allocation + loop CON_REDIT arm are now proven and ready for `oedit`/`medit` to inherit. Remaining work: `plan-phase6-olc-oedit.md` + `plan-phase6-olc-medit.md` (unauthored — soft-blocked on olc-redit which has now landed, so these can be drafted).
7. **`DoBio` / `DoDescription`** (P2): **Unblocked** now that `plan-editor-save.md` landed — wire via the `EditorSave` callback pattern demonstrated in `DoRedit desc` / `ed`.
8. ~~**Hotboot/copyover** (P2, infrastructure): Q1 resolved 2026-04-19 — Linux/Unix only with `//go:build !windows` guard.~~ **LANDED 2026-04-19** via `plan-phase6-hotboot.md`. 8 task groups (pre-G0.a/b + G0–G7), all 18 acceptance criteria satisfied, 16 mutation gates verified, 57 new tests including `TestHotboot_EndToEnd` integration (~9.4s measured real hotboot pause). 6 Windows stubs keep `GOOS=windows go build ./...` clean (pinned by new `make windows-build-check` target in `smaug-go/Makefile`). Windows-native hotboot remains out of scope; dedicated `plan-phase6-hotboot-windows.md` tracked if demand emerges. See `smaug-go/doc/plan-phase6-hotboot.md` §Completion Record for per-criterion test citations.
9. **Marriage**: DEFERRED 2026-04-19 (human decision — park; revisit on player demand signal). Q2 vnum-100/101 collision not worth forcing a resolution without a use case. **Note:** when marriage is revisited it will be a **rewrite** per `smaug-go/doc/post-phase6-vision.md` §2 (poly-capable, per-coupling agency), not a port of the C two-person `Spouse` field.
10. **Big optional systems** (P3): overland, housing, polymorph, archery, arena, dragon flight, planes, holidays, star maps. Each is large and self-contained; pick by demand signal, not order.

---

## Active

### Post-hotboot adversary review follow-ups (2026-04-19)

Deferred from `plan-phase6-hotboot.md` §Post-Execution Adversary Review — non-blocking for Phase 6 closure but actionable:

- [ ] **Pre-existing race in `TestBio_RoundTripThroughEditor`** (`internal/testclient/bio_test.go:61`). Test goroutine reads `ch.PCData.Bio` while the game loop's `DoBio` `EditorSave` closure writes the same field. Caught by `go test -race -count=1 ./internal/testclient/`. Default `go test -count=3 ./...` is clean; race only surfaces with `-race`. Fix shape: add a `sync.Mutex` guard on `PCData.Bio` access, OR read from a snapshot the test explicitly requests. Not introduced by hotboot — surfaces any time the testclient package runs under `-race`.
- [ ] **Supervisor / Docker deployment notes** for `smaug-go/doc/plan.md` — `systemd Type=exec` recommended, `--hotboot-recover` must NOT be passed by operator restart scripts, PID-1 container caveats. Ops adversary flagged as LOW-severity doc gap.
- [ ] **`/proc/<pid>/cmdline` threat-model note** — argv FD integers are world-readable on Linux during the ~10s recovery window; same-UID-only attack surface. Documentation entry in `smaug-go/doc/plan.md` security section.
- [ ] **`os.Executable()` → `syscall.Exec` TOCTOU** — symlink swap between `EvalSymlinks` resolution and `syscall.Exec` execution. Mitigated in practice by binary-directory write permissions; worth a deployment-note mention. LOW.

### Phase 6 post-landing audit (2026-04-19)

Three parallel adversary audits — functionality/C-parity, security, and Go idiom — over all 13 landed Phase 6 commits. Functionality audit was PASS. Four fixes landed in this audit pass; the rest are queued below.

**Landed in audit pass:**
- [x] **HIGH (security)**: `internal/act/holidays.go:177,253,263` — `DoSetHoliday` now applies `util.SmashTilde` to `name` / `announce` / create-name. An immortal could otherwise inject `~`-terminated tokens into `holidays.dat` and corrupt subsequent `LoadHolidays` parses. Three pin tests in `internal/act/holidays_test.go` (`TestDoSetHoliday_NameSmashesTilde`, `_AnnounceSmashesTilde`, `_CreateSmashesTildeInName`).
- [x] **HIGH (security)**: 5 persist writers switched from `os.Create` (default `0o644` under standard umask) to `os.OpenFile(..., 0o600)` — `holidays.go:149`, `planes.go:117`, `stances_save.go:55`, `morphs.go:403`, `subsystems.go:230` (SaveClanFile). Pin tests in each `*_test.go` (`TestSaveHolidays_FileMode`, `TestSavePlanes_FileMode`, `TestSaveStances_FileMode`, `TestSaveMorphs_FileMode`, `TestSaveClanFile_FileMode`). Matches the 2026-04-19 hotboot precedent.
- [x] **MEDIUM (security)**: `internal/util/parsebet.go:78,88` — `ParseBet` `+N%` and `*N` paths now reject overflowing products via new `mulOverflows` helper. Crafted `*1844674407370960` would otherwise wrap to a small valid bid and pass the auction's 2-billion ceiling check. Pin tests `TestParseBet_MultiplyOverflowReturnsZero`, `TestParseBet_PercentOverflowReturnsZero`, `TestMulOverflows`.
- [x] **SHOULD-FIX (idiom)**: `internal/persist/subsystems.go:164-217` — `SaveClan` now uses an `errWriter` wrapper that short-circuits on the first failed `Fprintf` and surfaces the error. Prior implementation discarded all 39 write returns and silently returned nil on broken-pipe / disk-full. Pin test `TestSaveClan_PropagatesWriteError` (mutation-verified).

**Queued for follow-up (no fix this pass):**
- [ ] **MEDIUM (security, future-risk)**: `internal/act/olc.go:799` — `filepath.Join(WorldRef.DataDir, "area", area.Filename)` has no path-containment check. Safe today (filename comes from loaded `.are` files), but becomes a path-traversal risk when a future `oedit`/`medit` plan lets immortals set `area.Filename` interactively. Add `strings.HasPrefix(filepath.Clean(path), filepath.Clean(filepath.Join(...)))` containment guard at that point. Same forward-risk applies to `SaveClanFile` if a future `do_setclan` exposes `Filename` editing.
- [ ] **TODO (idiom)**: `internal/combat/combat.go:449`, `internal/act/archery.go:887`, `internal/game/redit_menu.go:234` — three private copies of the same direction-name lookup. Will diverge further as `oedit`/`medit` editors land. Consolidate into a single `util.DirectionName(dir int) string` helper. Three lookups, ~15 LOC saved + drift risk eliminated.
- [ ] **TODO (idiom)**: `internal/persist/holidays.go:43`, `planes.go:42`, `morphs.go:44` — `make([]*T, 0)` allocates a non-nil empty slice when `var list []*T` would do; the `nil`-vs-empty distinction is load-bearing for the missing-file vs empty-file semantics in these loaders. Cosmetic, but misleading next to the empty-slice contract. Switch to `var` form.
- [ ] **TODO (idiom)**: `internal/game/redit_parse.go:25` (`worldRef`) — third package-level world-pointer seam after `act.WorldRef` and `mudprog.WorldRef`. Each new OLC editor (`oedit`, `medit`) will need its own. Consider passing `*world.World` through `reditParse` instead of via `SetWorldRef`. Would also let `worldRoomLookup` (line 718) drop the seam-of-a-seam indirection. Defer the refactor to when the second OLC editor lands.

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
- [x] Full `do_auction` state machine — **LANDED 2026-04-19** via `plan-phase6-auction.md`. 10 task groups, 23 acceptance criteria all satisfied, 8 mutation gates verified, +86 tests.
- [x] `music` / `newbiechat` / `racetalk` / `wartalk` / `counciltalk` / `guildtalk` channel commands — **LANDED 2026-04-18 via `plan-phase6-channels-extra.md`.** Shared `talkChannel` helper in `internal/act/channels.go` + 6 thin wrappers. 6 groups, 18 criteria all satisfied, 7 mutation gates verified via `Edit` round-trips. C's `||`-bug in wartalk deaf-exception corrected to `&&` (with both WARTALK and YELL exceptions carried for future callers). Tier 9 channels (`DoImmtalk`/`DoGtell`/`DoClantalk`/`DoAuction`) deliberately not retrofitted — separate plan.

  Follow-ups from plan-phase6-channels-extra.md (all deferred per plan scope cuts):
  - [ ] Uniform `PLR_WIZINVIS` "(level) " preamble across all 10 channels.
  - [ ] Uniform `ROOM_LOGSPEECH` file-append across all 10 channels.
  - [ ] Uniform `is_ignoring` receiver filter (helper needed; `PCData.Ignored []string` exists).
  - [ ] Uniform `AFLAG_SILENCE` area-flag check (constant exists; no caller today).
  - [ ] Retrofit Tier 9 channels onto `talkChannel` when a next wave of channels (quest/ask/muse/think/avtalk) lands.
  - [ ] Per-`AT_` color fidelity for all channel commands — `util.Act` per-call color seam landed Tranche C; channels still use inline `&Y/&G/&D`.
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

- [x] ~~Add `SaveClan`~~ / Add `SaveDeity` in `persist/subsystems.go` when the first mutation command ships. **`SaveClan` LANDED 2026-04-19** via `plan-phase6-clan-officer.md` G0 (`SaveClan(w io.Writer)` + `SaveClanFile(dir, c)` helpers, byte-for-byte C parity). `SaveDeity` still pending; land alongside first deity-mutation command.
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
- [ ] Re-verify unclosed factual scope brackets in `phase6-roadmap.md`: the citations `src/overland.c:839-3363`, `src/stances.c:199-1022`, `src/house.c:60-2853` use a start-line that isn't always a function boundary. Low priority — these are "read this region" brackets, not precise entry points.
- [ ] Cleanup: remove orphan `ObjData.ExtraDescr2 string` at `internal/types/object.go:62` (marked as "Marriage extra descr" in comment but Marriage plan uses `ExtraDescr []*ExtraDescrData` instead, matching C's linked-list shape).
- [ ] Follow-up (deferred by plan-phase6-holidays.md): `DoTime` calendar styling upgrade — switch to `"the Month of %s"` with `types.MonthName(TimeInfo.Month)`. ~10-line change in `internal/act/info2.go:95-126`.
- [ ] Follow-up (deferred): `DoLoad <subsystem>` umbrella admin command. C has it at `src/act_wiz.c`; Go port has no equivalent. Low priority — boot-time reload via restart covers most cases.
- [ ] Follow-up (deferred): `DoCset` admin command to tune `SysData.MaxHoliday` / `DaysPerMonth` / `MonthsPerYear` at runtime. C has `src/act_wiz.c:8141-8146`. Holidays plan hard-codes `MaxHoliday=32` default; `DoCset` lands with a future sysdata-loader plan.
- [ ] Follow-up (deferred): `LoadSysdata`/`SaveSysdata` — port `db/system/sysdata.dat` loader. Separate plan when pressure arrives.
- [ ] Cleanup scheduled: `/tmp/hotboot-poc-a/` and `/tmp/hotboot-poc-b/` throwaway PoCs. Remove after the hotboot executable plan re-verifies PoC outcomes and its first integration test lands green.
- [x] `game/prompt.go:63` — wire XP-to-next-level into `%x` / `%X` tokens. **LANDED 2026-04-18 (Tranche A item 3).** `%x` already printed current Exp; `%X` was the stubbed-zero — now returns `exp_level(ch, level+1) - ch.Exp` via a new `game.PromptExpBase` seam (wired in boot from `w.Classes[ch.Class].ExpBase`; NPCs use 1000 per C `handler.c:107-112`). Clamped non-negative. Seam + clamp + boot-wire tests in `internal/game/prompt_test.go` and `internal/boot/boot_test.go`.

---

## Phase 6 candidates (large, self-contained)

**Phase 6 planning begun 2026-04-18 — see `smaug-go/doc/phase6-roadmap.md`** for the full inventory, dependency graph, risk matrix, and recommended 6-wave execution order. First executable plan: `plan-phase6-arena.md`. All other systems below have proposed plan-doc filenames queued in the roadmap's § Cross-Plan Dependencies table.

**Status of authored Phase-6 plans (Wave A landed 2026-04-18):**
- [x] Phase 6 roadmap — audited 2026-04-18 via `audit-roadmap` lineage; 21 factual corrections applied (several "missing schema" claims were wrong — `MorphData`/`CharData.X/Y/Map/Sector`/`RoomIndexData.Plane`/`PCData.Spouse`/`ITEM_PROJECTILE/QUIVER`/`WEAR_MISSILE_WIELD` are already defined).
- [x] `plan-phase6-arena.md` — drafted + audited 2026-04-18 via `audit-arena` lineage; 6 fact corrections applied. Open Q6 (ROOM_ARENA flag absence on vnums 10366-10382) confirmed — prereq area-data edit required.
- [x] `plan-phase6-hotboot.md` — drafted 2026-04-18 as DESIGN-EXPLORATION (not executable); **audited 2026-04-18 (PASS with notes)** via `audit-hotboot` lineage. 28 C + 19 Go citations verified. Windows double-lock strengthened; `syscall.Dup2` → `syscall.Dup3(old, new, 0)` needed for linux/arm64.
- [x] `plan-phase6-starmap.md` — drafted 2026-04-18; **audited 2026-04-18 (PASS)** via `audit-starmap` lineage. 3 groups, 13 criteria. 3 fidelity corrections applied: critical row-6 constellation transcription error (`"C*"` → `"c."` per starmap.c:66); precip-bounds citation `update.c:3377-3378` → `:3446-3452`; `DoWeather` flag-only-vs-flag-OR-sector pattern. Ready to execute.
- [x] `plan-phase6-holidays.md` — drafted 2026-04-18; **audited 2026-04-18 (CONCERNS)** via `audit-holidays` lineage. 4 groups, 13 criteria (+ A14/A15 pending Q6). Fixes 2 latent C bugs. Audit corrected 2 material C claims: `get_holiday` has 2 live callers in `src/timezone.c` under `#ifdef ENABLE_HOLIDAYS` (plan claimed zero — season-tick-announce machinery may need porting); `do_load holiday` command doesn't exist in C (fabricated claim removed).
- [x] `plan-phase6-marriage.md` — drafted 2026-04-18; **audited 2026-04-18 (CONCERNS)** via `audit-marriage` lineage. Plan is self-consistent at G1-G5 (5 groups, not 7), 12 criteria. Discovered `PCData.Spouse` orphan; `SavePlayer` Spouse asymmetry. Audit uncovered **vnum-100/101 collision** with `newgate.are` — original nil-guard strategy invalid; Q2 resolution space broadened to 7 options. Re-audit required after Q2 human input.
- [x] `plan-phase6-planes.md` — drafted 2026-04-18 via `phase6-planes` lineage; **audited 2026-04-18 (CONCERNS)** via `audit-planes` lineage. 3 groups, 13 criteria. Audit caught algorithmic bug in `DoPset delete` slice-splice + `one_argument` quote-support misstatement.
- [x] `plan-phase6-channels-extra.md` — drafted 2026-04-18 via `phase6-channels-extra` lineage; **audited 2026-04-18 (CONCERNS)** via `audit-channels-extra` lineage. 6 groups, 18 criteria. 5 sites of `util.TranslateFor` misattribution corrected to `translateFor` (`internal/act/comm.go:13`). `db/councils/council.lst` verified empty → `newbiechat` immortal-only by default (matches C). Executable without re-audit.
- [x] `plan-phase6-skills.md` — drafted 2026-04-18 via `phase6-skills` lineage; **audited 2026-04-18 (PASS with clarifications)** via `audit-skills` lineage. 3 groups, 16 criteria, parallel-safe. Q1 reachability refined: `IS_VAMPIRE`/`IS_DEMON` macros expand to race-OR-class disjunctions, so gate IS reachable for dual-identity characters while functionally dead for typical single-identity. Option A (fix-the-bug) recommendation stands.

**Status of Wave 2 plans (authored 2026-04-18 Wave C, pending external audit):**
- [x] `plan-phase6-stances-olc.md` — authored 2026-04-18 via `phase6-stances-olc` lineage; **audited 2026-04-18 (CONCERNS)** via `audit-stances-olc` lineage (first dispatch stalled at 600s stream watchdog; re-dispatched with time-boxed prompt). 5 groups (G1-G5), 20 criteria, 10 open questions. **Correctness-critical audit finding**: §D3 `UpdateStances` mixes `int` CharData fields with `uint32` RIS_* constants — cast required at fixture sites, compile-error risk at G2 without inline note. Citation corrections: §Q6 RisflagNames has 22 bits not 20 (terminal `magic`/`paralysis`); §Q9 `stance` already registered at `boot.go:521`. Q8 resolved (`world.go:76` DataDir). 4 in-plan edits applied.
- [x] `plan-phase6-auction.md` — authored 2026-04-18 via `phase6-auction` lineage; **audited 2026-04-18 (CONCERNS)** via `audit-auction` lineage. 10 groups, 23 criteria, 8 open questions. Audit applied 4 factual corrections: carry-weight NOT dead code (reachable during 27s window); C has active DoQuit gate at `act_comm.c:2883-2890` (plan claimed passive `mud.h` comment only); `ms_find_obj` is generalized drunk/mental-state check at `handler.c:2941` (not mud-school hook); C ships full `do_noauction` admin command at `act_wiz.c:11120-11174` (load-only port deferred). Pre-existing `pulseSave` uninitialized bug at `loop.go:68-94` flagged.
- [x] `plan-phase6-clan-officer.md` — authored 2026-04-18 via `phase6-clan-officer` lineage; **audited 2026-04-18 (PASS after 3 edits)** via `audit-clan-officer` lineage. **Scope-correction independently verified** (`grep -rn "do_promote\|do_demote" src/` zero hits). 5 groups (+1 conditional), 14 criteria, 4 open questions. Audit added missing `Abbrev` to loader-missing-keys list; flagged legacy-PKills index bug (Go `[0]` vs C `[6]` at `fread_clan:435`); neutralized one-sided Q1 bias.

**Status of Wave D plans (authored + audited 2026-04-18):**
- [x] `plan-phase6-olc-redit.md` — authored 2026-04-18 via `phase6-olc-redit` lineage; **audited 2026-04-18 (PASS minor CONCERNS)** via `audit-olc-redit` lineage. 13 groups, 18 criteria, 6 open questions. First of three OLC-substate plans. `OlcData` on `DescriptorData`. C bug at `oredit.c:860-875` (REDIT_EXIT_VNUM asymmetric) — plan fixes in Go. Audit verified 11/11 C + 8/8 Go citations; applied 2 surgical edits (G5 EditorSave closure re-sets CON_REDIT AFTER StopEditingFunc; G12 promptless-menu substring-match note).
- [x] `plan-phase6-polymorph.md` — authored 2026-04-18 via `phase6-polymorph` lineage; **audited 2026-04-18 (PASS with CONCERNS, self-review)** via `audit-polymorph` lineage. 7 groups (G0-G6), 20 criteria, 7 open questions. Schema complete (confirmed via 5-field spot-check). Direct stat mutation in new `internal/handler/polymorph.go` (NOT AffectModify). G4 morphset 889 LOC (not 700) → G4a/G4b split mandatory. AT_MORPH + COND_BLOODTHIRST + parseDiceExpr already exist in Go; G0 simplified to promotion. UpdateAris dropped (Go no-update_aris invariant).
- [x] `plan-phase6-archery.md` — authored 2026-04-18 via `phase6-archery` lineage; **audited 2026-04-18 (CONCERNS)** via `audit-archery` lineage. 6 groups, 15 criteria. 3 C bugs preserved verbatim (do_draw WEAR_DUAL_WIELD omission, do_dislodge arm asymmetry, do_fire dead victim==ch). OMIT combat-loop hook confirmed. Schema-gap corrections (ROOM_NOMISSILE, PLR_NICE already present; ITEM_WEAR_MAX needs 21→24 bump). Audit-discovered: mob_fire value-index asymmetry; lodged-arrow bypass via `remove`.

**Open items blocking Phase 6 plan execution:**
- [ ] External adversary pass with a real `Agent`-tool-equipped harness for the starmap/holidays plans (Wave A self-review only originally — starmap audited 2026-04-18, holidays audited 2026-04-18 via `audit-holidays` lineage; both had self-review fallback). All 14 Phase-6 plans have been externally audited as of 2026-04-18 but every audit fell back to structured self-review. Re-run when `Agent` tool becomes available to `manager` subagents (see tooling investigation follow-up).
- [ ] Resolve `plan-phase6-hotboot.md` Open Question 1 (Windows support) — Design A is Linux/macOS only; audit strengthened to double-lock (`syscall.Exec` absent AND `(*net.TCPConn).File()` fd not usable on other processes per Go stdlib). A Windows build adds Design B fallback behind `//go:build windows` and roughly doubles implementation cost. Human decision needed before executable-hotboot rewrite.
- [ ] Resolve `plan-phase6-marriage.md` Q2 vnum-collision — `newgate.are` registers non-ring objects at vnums 100/101 (candelabra + magical spring); 7 resolution options (new vnums, remove newgate entries, runtime shape validation, boot-time synthetic registration, etc.). Plan recommends new vnums as smallest operational risk.
- [ ] Resolve `plan-phase6-marriage.md` Q1 (level-10 gate — dead-code omit), Q3 (C-bug policy), Q4 (one-ring anomaly), Q5 (message-typo policy) — plan has recommendations; human confirmation preferred.
- [ ] Reconcile upstream-doc drift: `CLAUDE.md` / `TODO.md` / `phase6-roadmap.md` referenced marriage plan as "7 groups G0/G0b/G7" but plan is self-consistent at G1-G5. CLAUDE.md already updated 2026-04-18; roadmap line 122 still references G0/G0b pattern — cosmetic only, not a functional blocker.
- [ ] Resolve `plan-phase6-arena.md` Open Questions 1-7 — Q3 (arena-room selection filter), Q4 (challenge command level), Q6 (ROOM_ARENA flag) are the highest-priority before worker dispatch.

**Audit follow-ups (Wave B audits, 2026-04-18):**
- [ ] Hotboot pre-G1 prerequisite: wire `HomeVnum` on sentinel-mob creation. `internal/handler/handler.go:14 CreateMobile` does not populate `HomeVnum` from `idx.Act & ACT_SENTINEL`. C `src/hotboot.c:86-94` saves `mob->home_vnum` for sentinel mobs; Go currently saves 0 → recovery restores to LIMBO. Add `if util.XIsSet(mob.Act, ACT_SENTINEL) { mob.HomeVnum = room.Vnum }` in `CreateMobile` (or equivalent char_to_room path). Blocks G1 of the hotboot executable plan.
- [ ] Hotboot plan: pfile-schema migration policy — document as an Operations constraint (rolling-restart only with matching schema) in the executable plan.
- [ ] Hotboot plan: partial-line readLoop buffering at exec time — pause readLoops + drain pending `InputQueue` before FD extraction; add to executable plan's G3 implementation notes.
- [ ] Hotboot plan: mid-login descriptor-close C bug policy — C `src/hotboot.c:675` writes farewell to `ch->desc` (probable typo for `d`). Decide mirror vs fix in executable plan's G3.
- [ ] Hotboot plan: linux/arm64 portability — use `syscall.Dup3(old, new, 0)` instead of `syscall.Dup2`; same semantics when flags==0, available on all Linux architectures.
- [ ] Hotboot plan: G6 integration-test harness budget — build-tag `//go:build integration` so it doesn't run on every `go test ./...`. Realistic 1-2 day budget.
- [ ] Hotboot plan: re-estimate LOC after executable-plan task expansion (current ~555 Go LOC estimate likely low).
- [x] ~~Planes plan Q5 (roadmap + plan): feature is cosmetic~~ — resolved 2026-04-19. Landed as planned (see CHANGELOG 2026-04-19). Schema-already-exists argument held up; three commands + loader + saver + boot-pass shipped.
- [x] ~~Planes plan Q9 (added post-audit): SmashTilde ordering~~ — resolved 2026-04-19. SmashTilde applied BEFORE duplicate-check per audit recommendation.
- [ ] Planes plan follow-up (deferred): decide whether `DoPset` syntax help should include a quoted-form example (`pset "Prime Material" delete`). Plan recommends YES — one extra line, makes multi-word addressability discoverable. Not in scope of 2026-04-19 landing.
- [x] ~~Planes plan G2 mutation-verify: DeleteFirstPlane / DeleteMiddlePlane / DeleteLastPlane~~ — resolved 2026-04-19. All three index-position tests shipped; `TestDoPset_DeleteMiddlePlane` pins the audit-corrected splice-delete algorithm. Mutation #8 (drop `CheckPlanes` after splice) fails 3 tests including the middle-index one.
- [x] ~~Skills plan Q1~~ — resolved 2026-04-18 during execution. Option A shipped. See plan Completion Record + CHANGELOG entry.
- [x] ~~Skills plan G3 pre-flight: verify `combat.Damage(w, ch, ch, ...)` self-target safety~~ — resolved 2026-04-18. Readiness vet confirmed `damageWith` guard at `internal/combat/combat.go:647-658` is safe (skips `StartFighting` when `ch == victim`; `victim.Hit <= 0` early-returns). `TestDoBloodlet_SuccessSelfDamages` pins this behaviour (ch.Hit=100 → 96 after level-20 bloodlet self-damage).
- [x] ~~Skills plan: add one-line scope-cut entry noting C `#ifdef OVERLANDCODE` 3-arg `obj_to_room` fork at `src/skills.c:3549-3553`~~ — resolved 2026-04-18. Go port uses 2-arg `handler.ObjToRoom(obj, ch.InRoom)` unconditionally per existing project convention. Noted inline.
- [ ] (Optional) Ship a stock "Newbie Council" definition — `db/councils/council.lst` is empty in stock data; `newbiechat` is immortal-only by default in both C and Go. Low-priority polish if non-empty default desired.

**Audit/author follow-ups (Wave C, 2026-04-18):**
- [ ] Starmap G3 fixture: spot-check `readOutput` / `setupInfo3World` helpers in `internal/act/info3_test.go` before writing G3 assertions on `&Y|`/`&W@` substrings — verify no test fixture installs a stripping `ColorFunc`. If so, tests must configure passthrough descriptor.
- [ ] Starmap G1 nice-to-have: upgrade `TestRenderStarmap_EclipseAtNoonRendersMoon` from substring-presence to full row-byte pinning on rows 3/4/5 at `hour=12, day=0, month=0`. Catches subtle eclipse-condition inversions.
- [ ] Holidays plan author follow-up: add A14 (`GetHoliday` one-liner) + A15 (`DoTime` holiday-today suffix) per Q6 scope expansion; add scope-cut entry for `season_update` announce-tick pending Go tick-hook grep.
- [ ] Grep Go codebase for `season_update` / season-tick / pulse-calendar hook before executing holidays plan. If exists, land the hour-0 `echo_to_all(day.Announce)` broadcast per C `src/timezone.c:617-631`. If absent, log as deferred under `game.PulseUpdate` or season-loop owner.
- [ ] Re-audit methodology sweep: grep every "zero callers" / "never called" claim across remaining Phase-6 plan docs against C source — holidays plan had two such claims wrong; possible sibling plans have similar errors.
- [ ] Stances-olc plan: resolve Open Questions Q1-Q10 before worker dispatch (most have recommended answers; Q2/Q3/Q4 ask whether to preserve C bugs verbatim — default YES).
- [ ] Stances-olc follow-up (after plan lands): `stance-ris-wire` — wire `CharData.StanceResistant/Immune/Susceptible` into combat R/I/S resolution path. Data populated by `UpdateStances` but not yet consumed.
- [ ] Stances-olc follow-up (after plan lands): richer `class_string`/`race_string` formatter so `DoSTstat` prints human names instead of raw hex masks.
- [ ] Stances-olc follow-up (after plan lands): add `STANCEFLAGS` help-file entry (referenced in error paths).
- [x] Auction plan Q4 — **RESOLVED 2026-04-19** as plan recommended default: always `ObjToChar` (skip weight check). Deferred to future `auction-carry-weight-cap` follow-up when a weight-cap subsystem lands.
- [x] Auction plan G2 — **RESOLVED 2026-04-19**: `db/system/noauction.dat` not present in stock tree (verified at boot — `Loaded 0 noauction entries`). Loader is permissive on missing file.
- [x] Auction plan dispatch — **RESOLVED 2026-04-19**: plan executed end-to-end in a single session (plan-recommended split ignored since all G0-G9 completed in one pass).
- [ ] Auction × Hotboot cross-reference: mid-auction state (`world.Auction`) serialization policy for hotboot. Current recommendation: do not serialize (match C); hotboot cancels active auctions.
- [ ] **`auction-carry-weight-cap` follow-up** — when a weight-cap subsystem lands, restore C checks at `update.c:3247-3262` (sold branch) and `:3293-3308` (unsold branch). Today: sold → always `ObjToChar(buyer)`; unsold → always `ObjToChar(seller)`. Seller could be over carrying capacity (27s auction window allows weight acquisition).
- [ ] **`economy-helpers-unify` follow-up** — `mudprog/commands.go` retains package-private `boostEconomy`/`lowerEconomy` helpers while `handler/economy.go` now exposes public `BoostEconomy`/`LowerEconomy`. Low-priority future-tidy to consolidate.
- [x] Clan-officer plan Q1 (human input): include G5 or defer? **RESOLVED 2026-04-19**: Option 2 (defer) chosen during landing. Officers can induct/outcast but not promote; rank management gated behind future `plan-phase6-setclan.md`.
- [ ] Clan-officer plan Q2 (low-priority): add `IsPkill()` method on `CharData` (6 call sites) or keep inline bit check. Landed with package-private `isPkill` helper in `act/clan_officer.go`; promote to `CharData.IsPkill()` method if a 7th+ call site appears.
- [x] Clan-officer plan Q3: `SaveClanFile` clan-dir seam. **RESOLVED 2026-04-19** as Option A — `act.ClanDir` package var set during `Boot()`, parallel to `act.PlanesFilePath`.
- [x] Clan-officer plan Q4: pfile persistence path. **RESOLVED 2026-04-19** — `DoInduct` and `DoOutcast` call `SaveFunc(victim)` (wired to `GameLoop.SavePlayer` at `boot.go:92`), matching the pattern used by `DoSave` and other existing commands.
- [ ] Follow-up plan: `plan-phase6-bestowarea.md` — port `do_bestowarea` (`src/act_wiz.c:7004-7076`), 50-LOC immortal command. Adjacent to `do_bestow`; deliberately out of clan-officer scope.
- [ ] Follow-up plan: `plan-phase6-setclan.md` — full `do_setclan` port (30+ field setters including rank management — leader/number1/number2 promotion). ~600 C LOC surface. Elevated priority now that clan-officer G5 deferred: officers cannot currently promote peers.
- [ ] Clan-officer follow-up: per-clan roster file (`save_member_lists` / `add_member` / `remove_member`) — display nicety; `clan.Members` counter sufficient for officer mechanics. Enables `claninfo` to show member list.
- [ ] Clan-officer follow-up: `add_loginmsg` subsystem — Go has no equivalent; `DoOutcast` currently logs via `util.Bug` when outcasting linkdead victim. Future queue-on-disk-and-flush-on-login matches C `do_outcast:1323`.
- [ ] Clan-officer display-path follow-up: `DoClanInfo` at `act/clan.go:67` reads `clan.PKills[0]` (per-level range slot). C's own display is split — `pkills[0]` at `clans.c:1948` and `pkills[6]` at `:2142`. Consider showing both range-slot-0 and cumulative-slot-6, or switch to `[6]` for the headline total. Cosmetic only.
- [ ] **Environment gap (persistent across 3 waves)**: `manager` subagents report `Agent` tool unavailable despite spec listing it. Every 2026-04-18 Phase-6 adversary audit fell back to structured self-review. Raise with human to establish workflow (e.g., human-triggered external adversary pass) before Wave-1 execution begins. **NEW 2026-04-18**: `audit-stances-olc` first dispatch also stalled at stream watchdog (600s no progress) before writing any drafts — re-dispatched successfully with time-boxed prompt instructing early-write-then-refine.

**Audit follow-ups (Wave C audits, 2026-04-18):**
- [x] Auction plan Q4 sign-off — **RESOLVED 2026-04-19** with landing: "always ObjToChar" default shipped; unsold-branch carry-weight gap tracked as `auction-carry-weight-cap` follow-up above.
- [x] Auction plan Q8 disconnect broadcast wording — **RESOLVED 2026-04-19** with landing: "The auction has been cancelled — the seller has left the game." (seller drop) and "The top bidder has left the game. Bidding re-opens at the starting price." (buyer drop). Go-originated wording documented inline at `internal/game/auction_disconnect.go`.
- [ ] **`auction-do-noauction-admin` follow-up** — port `do_noauction` immortal CRUD (C `act_wiz.c:11120-11174`) listing + toggling + saving no-auction vnums. Load-only path shipped in `plan-phase6-auction.md` landing (2026-04-19). Low-priority — the list is rarely mutated in practice.
- [ ] Fix pre-existing `pulseSave` uninitialized field at `internal/game/loop.go:68` — `NewGameLoop` at `:83-94` does not set it, so autosave fires on the very first pulse after boot. Add `pulseSave: types.PULSE_SAVE` to the struct literal. Discovered during auction audit; pre-existing bug, kept out of auction landing commit per readiness-vet scope gate.
- [x] Clan-officer Q1 human decision — **RESOLVED 2026-04-19**: Option 2 (defer G5).
- [ ] Investigate replacing `util.Bug` fallback in `DoOutcast` linkdead path with proper `add_loginmsg` equivalent — candidates: `internal/notes` (closest semantic fit) or new `PCFLAG_WAS_OUTCAST` login-hook. Follow-up after clan-officer has shipped (which it now has, 2026-04-19).
- [x] Clan-officer loader legacy PKills index-0 vs index-6 bug — **FIXED 2026-04-19** via `plan-phase6-clan-officer.md` G0. Loader now stores into `[6]` to match C `fread_clan:434-435`; `TestLoadClan_LegacyPKillsIndex6` pins the correction.
- [ ] Clan-officer per-plan adversary-pairing against C's ~35 `send_to_char` strings in induct/outcast/bestow is needed — current self-review covered structural match; dedicated string-pairing adversary pass queued.
- [ ] Stances-OLC Q2 human decision — preserve class/race setter no-ops verbatim (default C-fidelity) vs. implement proper `IS_SET(mask, 1<<class)` setters + matching fix to `CanUseStance`. Stance file is empty in stock data so likely safe to preserve bug.
- [ ] Stances-OLC Q6 final — cross-check `src/tables.c` `ris_flags` literal table at G4 implementation time; confirm no SMAUG-variant-specific bits beyond `RIS_PARALYSIS` (bit 21).
- [ ] Stances-OLC G2 test mitigation — swap `TestDoStance_Set`'s `"dragon"` → `"viper"` at `skills4_test.go:112` during G2 commit (default-zero Prereq on VIPER means test passes under new CanUseStance gate).
- [ ] Stances-OLC G4 RisflagNames length audit — dedicated test iterating every `RIS_*` constant asserting `RisflagNames[log2(RIS_X)] == "<expected>"` to detect table drift.

**Audit/author follow-ups (Wave D, 2026-04-18):**
- [ ] OLC-redit Open Questions 1-6: screen-clear emit, `OlcData` field location (Desc vs Char), color-code path, REDIT_EXIT_VNUM fix-vs-port policy, olc_log scope, explicit-vs-defer menu redisplay. All have recommended answers; Q4 (fix-not-port) matches Tranche C precedent.
- [ ] OLC-redit G12 testclient prompt-tracking: `ReditDispMenu` has no trailing `> ` sentinel — scenarios must match menu-text substrings (e.g. `"A) Exits"`). If flaky, consider emitting synthetic trailing sentinel; deferred to implementation.
- [ ] OLC-redit G1-G13 execution — 18 acceptance criteria; mutation matrix ≥8. Apply REDIT_EXIT_VNUM fix (update both `exit.Vnum` AND `exit.ToRoom` via `GetRoom(number)`) as documented divergence; log CHANGELOG entry.
- [ ] Polymorph Open Questions Q1-Q7 — most are G0 reconnaissance (dice_parse wrap/promote, UpdateAris no-op, separate_obj/extract_obj Go equivalents, stock morph.dat parse verification, COND_BLOODTHIRST array shape, boot ordering, SetupMorphVnum collisions). Q1/Q2 resolved inline by audit.
- [ ] Polymorph G1 worker recon: verify `db.c:941` places `load_morphs()` after area load, mirror ordering in `internal/boot/boot.go`.
- [ ] Polymorph G1 worker recon: verify `db/system/morph.dat` stock file exists and parses with G1 loader; handle unexpected tokens with non-fatal warn-and-skip.
- [ ] Polymorph G2 worker prompt: grep for Go equivalents of C `separate_obj` / `extract_obj` before porting DoMorphChar obj-consumption path.
- [ ] Polymorph G2 worker prompt: explicitly include C-fidelity resource-leak note — `do_morph_char` deducts hp/mana/move/blood/favour/glory as each check runs (not atomically); objs already extracted are not refunded if a later check fails. Preserve check-and-spend ordering.
- [ ] Polymorph G4 execution: adopt mandatory G4a/G4b split (morphset is 889 LOC not 700). G4a = structure + create/destroy/stat + minimal field set; G4b = full field coverage.
- [ ] Polymorph G0-G6 execution — ~1500 Go LOC estimated.
- [x] Archery G1 corrected scope: drop `ROOM_NOMISSILE` and `PLR_NICE` items (both present); add `ITEM_WEAR_MAX` bump from 21 to 24 at `internal/types/constants.go:588` (LANDED 2026-04-19 — both bumps in place: `ITEM_WEAR_MAX=24`, `MAX_WEAR=29`).
- [x] Archery Open Questions Q1-Q13 — all resolved in plan completion record (LANDED 2026-04-19).
- [ ] `archery-ux-lodge-remove-policy`: document lodged-arrow `remove` vs `dislodge` policy — C does not set `ITEM_NOREMOVE` on lodged arrows, so players can `remove` to bypass dislodge damage. Ship-state preserves bypass verbatim; decide whether to close in follow-up (would be new gameplay).
- [ ] `archery-followup-mob-fire`: port `mob_fire` NPC autonomous firing (`src/archery.c:1335-1362`). Needs mob-AI update-pulse hook. Resolve the value-index asymmetry at that time (`mob_fire:1352` uses `bow->value[4] != arrow->value[5]` vs `do_fire:1295` uses `[5]!=[4]` — port verbatim or reconcile).
- [ ] `archery-followup-is-safe`: port full `src/fight.c is_safe` — current Go archery `isSafe(ch, vch, checkFriendly)` is a minimal 3-branch gate (nil/self/ROOM_SAFE). Full C also checks SAFE_* stances, group membership, NPC protect, etc.
- [ ] `archery-followup-max-fight`: port full `src/fight.c:281 max_fight` — currently returns flat `3`.
- [ ] `archery-followup-separate-obj`: implement `handler.SeparateObj` when object stacking lands — currently a no-op (Go port has no object stacking yet).
- [ ] `archery-followup-ris-in-projectile-hit`: wire RIS bitmap check (C `src/archery.c:435-493`) once Go gets a public `ris_damage` helper.
- [ ] `archery-followup-weapon-spell-in-projectile-hit`: wire `APPLY_WEAPONSPELL` iteration (C `:608-629`) once skill_table spell-fun dispatch is callable from `act/`.
- [ ] `archery-followup-learn-from-failure-on-miss`: wire `learn_from_failure` on archery miss once the combat package exports a safe hook.

### Infrastructure

- [x] **Hotboot / copyover** (854 C LOC) — **LANDED 2026-04-19** via `plan-phase6-hotboot.md`. C ref: `src/hotboot.c`. See `plan-phase6-hotboot.md` §Completion Record; CHANGELOG 2026-04-19 entry.
- [ ] DNS resolution — `net.LookupAddr` for host display (out of Phase 6)
- [ ] Web status page — embedded HTTP server (out of Phase 6)
- [ ] MXP protocol parsing — C `protocol.c` has it; Go has no equivalent (out of Phase 6)

### Game systems

- [ ] **Arena PvP** (358 C LOC) — `plan-phase6-arena.md` drafted + audited 2026-04-18. Challenge / accept / decline / withdraw + teleport + victory branch. 7 task groups, 15 criteria. `TIMER_CHALLENGE=8` slot, AddTimer signature corrected, `ROOM_VNUM_ALTAR` already-exists confirmed.
- [ ] **Star maps** (226 C LOC) — `plan-phase6-starmap.md` drafted 2026-04-18. 3 task groups, 13 criteria. Pure render over shared constellation data.
- [ ] **Planes** (298 C LOC) — `plan-phase6-planes.md` drafted 2026-04-18 via `phase6-planes` lineage. 3 groups, ~20 criteria. `RoomIndexData.Plane` + `PlaneData` struct pre-exist; port is persistence + commands + `CheckPlanes` orphan-assign. External adversary queued.
- [ ] **Holidays** (416 C LOC) — `plan-phase6-holidays.md` drafted 2026-04-18. 4 task groups, 13 criteria. Bundles `month_name[]` port.
- [ ] **Marriage** (362 C LOC) — `plan-phase6-marriage.md` drafted 2026-04-18. 7 task groups, 12 criteria. Canonicalises `CharData.Spouse` (removes `PCData.Spouse` orphan) and fixes `SavePlayer` Spouse asymmetry. Correct ring vnums are `OBJ_VNUM_DIAMOND_RING=100` / `OBJ_VNUM_WEDDING_BAND=101` (roadmap's prior `STEEL_RING` naming was wrong).
- [ ] **Combat stances OLC** — `plan-phase6-stances-olc.md` drafted 2026-04-18 via `phase6-stances-olc` lineage. 5 groups, 20 criteria, 10 open questions. Extends `StanceInfo` 3→19 fields; adds `SaveStances`; ports `can_use_stance`+`update_stances`; rewrites `DoStance` (4-stance gap fix); adds `DoSTstat`/`DoSTset`. Preserves 4 C bugs; fixes 1 cosmetic. External adversary queued.
- [x] **Full `do_auction` state machine** — **LANDED 2026-04-19** via `plan-phase6-auction.md`. Completed Tier-9 `BroadcastAuction` stub by porting C `act_obj.c:3775-4240` + `update.c:3183-3326`. 10 groups, 23 criteria. Follow-ups: `auction-do-noauction-admin` + `auction-carry-weight-cap` + `economy-helpers-unify` + pre-existing `pulseSave` init bug (flagged, out of scope).
- [ ] **Archery** (1362 C LOC) — `plan-phase6-archery.md` drafted + audited 2026-04-18 via `phase6-archery` + `audit-archery` lineages. 6 groups, 15 criteria. 3 C bugs preserved verbatim. OMIT combat-loop hook. ITEM_WEAR_MAX bump 21→24 folded into G1.
- [ ] **Polymorph** (2753 C LOC) — `plan-phase6-polymorph.md` drafted + audited 2026-04-18 via `phase6-polymorph` + `audit-polymorph` lineages. 7 groups, 20 criteria. Schema complete. Direct stat mutation (not AffectModify). G4 morphset 889 LOC → G4a/G4b split mandatory.
- [ ] **Player housing** (2853 C LOC) — `plan-phase6-housing.md` to draft (Wave B+). New persistence schema + `RoomIndexData.OwnedBy` field. Prefer after hotboot executable lands.
- [ ] **Overland maps** (3752 C LOC) — `plan-phase6-overland.md` to draft (Wave B+). `CharData.X/Y/Map/Sector` already defined; Go port is map file format + loader + renderer + commands (`do_survey`/`coords`/`landmarks`/`setmark`/`setexit`/`mapresets`/`mreset`/`mapedit`). Split recommended.
- [ ] **Dragon flight** (945 C LOC) — `plan-phase6-dragonflight.md` to draft. **Blocks on Overland shipping** (every command reads `ch.Map`/`X`/`Y`).

### OLC / builder

- [ ] **Interactive `CON_REDIT` substate** — `plan-phase6-olc-redit.md` drafted + audited 2026-04-18 via `phase6-olc-redit` + `audit-olc-redit` lineages. 13 groups, 18 criteria, PASS with minor CONCERNS. First of three OLC plans; proves nanny-dispatch pattern for oedit/medit to inherit. OlcData on DescriptorData.
- [ ] **Interactive `CON_OEDITING` substate** — `plan-phase6-olc-oedit.md` to draft. Depends on redit nanny-dispatch pattern.
- [ ] **Interactive `CON_MEDITING` substate** — `plan-phase6-olc-medit.md` to draft. Depends on redit pattern.
- [ ] **Editable mudprog editors** (currently inspector-only) — `plan-phase6-olc-mpedit.md` to draft. Depends on interactive OLC substates.
- [ ] `foldarea` — area vnum repack (low-reward, high-risk; defer pending explicit builder demand).

### Content

- [x] ~~**Skills not yet ported** (`bloodlet`, `pounce`, `broach`)~~ — **LANDED 2026-04-18** (plan-phase6-skills.md G1-G3). See CHANGELOG.md entry and plan Completion Record.
- [x] ~~**Clan officer commands** (`induct`, `outcast`, `bestow`)~~ — **LANDED 2026-04-19** via `plan-phase6-clan-officer.md`. Five task groups (G0 SaveClan + loader extension, G1 isClanOfficer/DoInduct, G2 DoInduct edges + persist seam, G3 DoOutcast + echoToPKers, G4 DoBestow + boot wire); 14 acceptance criteria all satisfied. Q1 decision was Option 2 (DEFER G5) — rank management to future `plan-phase6-setclan.md`. Pre-existing `PKills[0]→[6]` loader bug fixed in-scope. See CHANGELOG.md 2026-04-19 entry + plan Completion Record.
- [x] **Extra channels** (`music` / `newbiechat` / `racetalk` / `wartalk` / `counciltalk` / `guildtalk`) — **LANDED 2026-04-18 via `plan-phase6-channels-extra.md`.** Shared `talkChannel` helper + 6 thin wrappers; Tier 9 channels not retrofitted (separate plan).
- [ ] Councils: all commands (not implemented as player-facing yet).
- [ ] **Deities: full prayer, favor beyond `mpFavor`, deity-specific effects** — `plan-phase6-deity-prayer.md` to draft. ~400 C LOC.

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

### Holidays follow-ups (from `plan-phase6-holidays.md`, landed 2026-04-19)

- [ ] Season-tick holiday announce — port C `season_update` at `src/timezone.c:617-631` which broadcasts `day->announce` via `echo_to_all(AT_IMMORT, day->announce, ECHOTAR_ALL)` on hour-0 of any holiday day. Go-side `weatherUpdate` at `internal/game/update.go:135-154` has no `echo_to_all`/`season_update` equivalent; landing requires a Go-side global-broadcast primitive or a new tick-hook layer. Deferred from Q6 resolution.
- [ ] `DoTime` calendar styling upgrade — use `types.MonthName(TimeInfo.Month)` for `"the Month of %s"` output matching C `do_time` at `src/act_info.c:2493`. Currently generic `"month %d"`. Bundle with a broader DoTime polish pass (ordinal day, weekday if ported).
- [ ] `DoLoad` umbrella admin command — C has no `do_load holiday` runtime reload (verified during holidays audit); `saveholiday` covers the in-memory→disk direction, but no disk→in-memory reload exists. Future Phase-6 general admin-reload subsystem could cover skills/races/holidays/planes together.
- [ ] `DoCset` / `sysdata.dat` loader-saver — `SysData.MaxHoliday`/`MonthsPerYear`/`DaysPerMonth` are boot-time-defaulted but not persisted. C `cset max-holidays N` at `act_wiz.c:8141-8146` tunes at runtime. Deferred to a future plan.
- [ ] `setholiday name` uniqueness check on rename — C doesn't enforce this (verified); Go port mirrors C. Latent UX bug: renaming A to match existing B's name leaves B unreachable via name lookup. No data-integrity risk — both entries still load/save correctly. Flag for future UX polish.

---

## Done

### 2026-04-18

- [x] **Phase 6 Skills — `bloodlet` / `pounce` / `broach` landed (plan-phase6-skills.md G1-G3).** All 16 acceptance criteria satisfied. Three parallel-safe player commands from `src/skills.c`:2607 (do_pounce), :3517 (do_bloodlet), :3960 (do_broach) + `isBloodRace` helper + 3 `boot.go` registrations. Both latent C bugs fixed per plan Option-A with in-code citations: bloodlet's operator-precedence gate (`!IS_NPC && (vampire || demon)` instead of the C form that required BOTH), broach's tautological+inverted predicate (`CLOSED && LOCKED && !PICKPROOF && can_use_skill` instead of the C `||` form). Bloodlet Wait uses literal `PULSE_VIOLENCE` not skill Beats (intentional C divergence at `:3535`). `gsnPounce` combat-side already shipped (`combat/skillcheck.go:110`); `OBJ_VNUM_BLOODLET` prototype already in shipped area data (`limbo.are2:806`). 41 tests added (14 Pounce + 11 Broach + 16 Bloodlet/isBloodRace). 5 mutation-verify cycles via `Edit`-only (weapon-type set, broach predicate, bloodthirst delta, Wait constant, reverse-exit lock removal). `go build ./...`, `go vet ./...`, `go test -count=3 ./...` all green. Deliberate omissions tracked as shared follow-ups: `is_safe`/`check_attacker`/`check_illegal_pk` / `adjust_favor` / `check_room_for_traps` / `set_char_color` / AT_BLOOD color — all match existing Go skill convention. Files: `internal/act/skills.go` (DoBroach), `skills3.go` (DoPounce), `skills4.go` (DoBloodlet + isBloodRace), `skills_test.go` (+11), `skills3_test.go` (+14), `skills4_test.go` (+16), `boot.go` (+3 regs). See plan Completion Record at bottom of `smaug-go/doc/plan-phase6-skills.md`.

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
