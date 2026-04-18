# TODO

## Recommended next steps (ordered)

Four plans are adversary-verified and ready to execute (2026-04-17). Each has its own doc under `smaug-go/doc/` with gap inventory, task groups (G1, G2…), acceptance criteria, and open questions.

1. **`plan-combat-depth.md`** (P0) — multi-attack cascade (second…seventh_attack), weapon proficiency bonus, stance application. Acceptance: a level-30 PC warrior with `second_attack`/`third_attack` learned gets the expected extra attacks per round.
2. **`plan-dammessage-gaps.md`** (P1) — `was_in_room` swap, `PCFLAG_GAG` self-suppress, poisoned-weapon prefix. Pairs naturally with the combat-depth plan (same package).
3. **`plan-player-config.md`** (P1) — `save`, `afk`, `title`, `password` (+ optional `pagelen` alias). Closes four TODO items in one session.
4. **`plan-channels.md`** (P1) — `immtalk`, `gtell`, `auction`-helper + stub. Unblocks multi-player coordination. Full auction system deferred to Phase 6.
5. **Interactive OLC substates** (P2): `CON_OEDITING` / `CON_MEDITING`. Blocked on an `editor.go` fix (`/s` doesn't call `StopEditing` — see `plan-player-config.md` R6). Harness has the `WithPrompt` seam ready (Tier 5).
6. **Hotboot/copyover** (P2, infrastructure): Go-native design required — C's `exec()` + fd-inheritance doesn't translate. Design pass first (save-all-state + graceful restart + auto-reconnect handshake), then implementation.
7. **Big optional systems** (P3): overland, housing, polymorph, archery, arena, dragon flight, planes, holidays, star maps. Each is large and self-contained; pick by demand signal, not order.

---

## Active

### High-impact combat gaps (from 2026-04-17 audit — P0)

**→ See `smaug-go/doc/plan-combat-depth.md` for the full plan (9 task groups, adversary-verified). LANDED 2026-04-17.**

- [x] Port PC multi-attack skills (second_attack … seventh_attack) into combat loop (plan G1–G5)
- [x] Port weapon proficiency bonus into `OneHit` (plan G6)
- [x] Apply `ch.Stance` in combat — NPC num_attacks stacking, PC GM bonus loop, dam_done/dam_taken multipliers (plan G7)

Follow-ups queued from plan-combat-depth.md:
- [ ] `handler.AddTimer` subsystem + `TIMER_RECENTFIGHT` wiring (deferred from G5)
- [ ] Devoted-clan favor penalty in `WeaponProfBonusCheck` (C fight.c:1312-1313)
- [ ] Per-round move-cost tracking (C fight.c:1149-1171)
- [ ] `db/system/stances.dat` loader — currently `StanceIndex` is hard-coded in `combat/stance_index.go`
- [ ] PC practice-stance flow — `PCData.Stances[]` counter never increments today, so GM-bonus path is unreachable for existing players
- [ ] Review `DoCircle` (`act/skills3.go:88-99`) and `DoHitall` (`act/skills3.go:294`) for explicit retcode handling now that `OneHit` returns `int`

### Player-visible command gaps (P1)

**→ See `smaug-go/doc/plan-player-config.md` for the player commands (4 groups) and `smaug-go/doc/plan-channels.md` for the channels (5 groups). Both adversary-verified.**

- [ ] Register `password`, `title`, `afk`, `save` (plan groups G1–G4)
- [ ] Optional `pagelen` alias (plan G5)
- [ ] Communication channels: `immtalk`, `gtell`, `auction`-helper + stub (`plan-channels.md` G1–G4). Full auction system and `music`/`newbiechat`/`racetalk`/`wartalk`/`counciltalk`/`guildtalk` deferred to Phase 6.
- [ ] `bio`/`description` deferred — blocked on `editor.go` `/s` bug where descriptor stays in `CON_EDITING` permanently (`plan-player-config.md` R6).

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

**→ See `smaug-go/doc/plan-dammessage-gaps.md` for the full plan (3 task groups, adversary-verified). Pairs with combat-depth.**

- [ ] Port `was_in_room` swap (plan G1)
- [ ] Port `PCFLAG_GAG` self-suppress (plan G2)
- [ ] Port `is_wielding_poisoned` prefix (plan G3) — must check both `WEAR_WIELD` and `WEAR_DUAL_WIELD` for obj identity

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
