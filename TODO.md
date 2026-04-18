# TODO

## Recommended next steps (ordered)

1. **Commit batch 1 fixes** once the coverage-gap follow-up worker lands and adversary-verifies (in-flight at the time of this writing).
2. **Combat depth** (P0, highest player-visibility gain): port multi-attack, weapon proficiency bonus, and stance application. The three are tightly coupled in C `fight.c` and natural to do as one unit. Acceptance: a level-30 PC warrior with `second_attack`/`third_attack` learned gets the expected extra attacks per round.
3. **Player-config commands** (P1, easy wins): `password`, `title`, `afk`, `save`. These are one-session tasks and close 4 TODO items.
4. **Communication channels** (P1): at least `immtalk`, `gtell`, `auction` — these three unblock multi-player coordination. Other channels (music/newbiechat/racetalk/wartalk) are lower priority.
5. **Interactive OLC substates** (P2): `CON_OEDITING` / `CON_MEDITING`. Harness has the `WithPrompt` seam ready (Tier 5). Unlocks a real builder experience.
6. **Hotboot/copyover** (P2, infrastructure): Go-native design required — C's `exec()` + fd-inheritance doesn't translate. Design pass first (save-all-state + graceful restart + auto-reconnect handshake), then implementation.
7. **Big optional systems** (P3): overland, housing, polymorph, archery, arena, dragon flight, planes, holidays, star maps. Each is large and self-contained; pick by demand signal, not order.

---

## Active

### High-impact combat gaps (from 2026-04-17 audit — P0)

- [ ] Port PC multi-attack skills (second_attack … seventh_attack) into combat loop
  - Reference: C `src/fight.c:1058–1132`
  - Target: `smaug-go/internal/combat/combat.go:ViolenceUpdate` / `OneHit`
  - Note: a level-30 PC currently gets 1 attack/round in Go; C gives 3–5
- [ ] Port weapon proficiency bonus into `OneHit`
  - Reference: C `src/fight.c:1385–1566` (`weapon_prof_bonus_check`)
  - Target: `smaug-go/internal/combat/combat.go`
- [ ] Apply `ch.Stance` in combat (currently stored/saved but unused)
  - Reference: C `src/fight.c:1058–1070`
  - Target: `smaug-go/internal/combat/combat.go`; `smaug-go/internal/act/skills4.go:183`

### Player-visible command gaps (P1)

- [ ] Register `password` command (change own password)
- [ ] Register `title` command (set own title)
- [ ] Register `afk` command (toggle `PLR_AFK` — flag already exists, `ban.go`/`comm.go` honor it)
- [ ] Register `save` command (manual save) — create `DoSave` calling the existing save path
- [ ] Register `bio`, `description`, `pagelen` (verify field plumbing first)
- [ ] Communication channels: `immtalk`, `gtell`, `auction` (top three); `music`, `newbiechat`, `racetalk`, `wartalk`, `counciltalk`, `guildtalk` (remainder)

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

From `combat/dammessage.go` TODOs:

- [ ] Port `was_in_room` swap (C `fight.c:4432–4438`) so dying/moved characters see messages from the right room
- [ ] Port `PCFLAG_GAG` handling (C `fight.c:4481–4486`) so gagged players are skipped
- [ ] Port `is_wielding_poisoned` prefix (C `fight.c:4496–4512`)

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
