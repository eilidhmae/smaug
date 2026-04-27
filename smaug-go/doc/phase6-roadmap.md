# Phase 6 Roadmap

**Status:** Planning begun 2026-04-18. Phase 5 complete (17 tiers landed; Tranche C most-recent at `26db5f9`). This document is the meta-plan; individual first-cut plans live in sibling `plan-phase6-*.md` files and are dispatched as independent manager lineages.

---

## Phase 6 Statement of Intent

Phase 5 closed out the audit-driven completeness work: combat depth (multi-attack cascade + prof bonus + stance application), damage messages, player-config commands, communication channels, `DoChannels`, timer subsystem, editor `/s` save, `DoGag`, `rollD20` seam, and Tranches A/B/C (util.Act color, stances loader, timer registry, mudprog ifchecks, wordlist match, PLR_BLANK, XP-on-skill-gain, Tranche A quick wins). The port is at feature parity with the C baseline for all *mandatory* play paths.

Phase 6 delivers the *optional* large-surface systems that were deferred because they are self-contained — a player can log in, fight, craft, chat, and explore without any of them — but each is a recognizable SMAUG feature whose absence makes the port feel incomplete.

**Done for Phase 6 means:**

1. Hotboot / copyover shipped (a Go-native design, not a mechanical port of `execl()` + fd inheritance). This is the one infrastructure item that affects every future change — every subsequent feature ships atop it.
2. Arena PvP shipped end-to-end: challenge / accept / decline / withdraw, two players teleport into isolated arena rooms, fight to resolution, get teleported out.
3. Housing shipped: players buy, furnish, and share apartment rooms that persist across reboot.
4. Overland maps shipped: 3x 1000x1000 tile sector maps, ANSI rendering, landmark/entrance/reset data, `coords`/`survey`/`landmarks` commands.
5. Polymorph shipped: `morph` / `unmorph` commands, `morphset` / `morphstat` admin, per-form stat overrides, persistence.
6. Archery shipped: `WEAR_MISSILE_WIELD` / quivers / projectiles / `draw` / `fire` / `dislodge`, arrows lodge in victims, ranged hit-roll.
7. Dragon flight shipped (optional; depends on Overland): `call` / `release` / `fly` / `land`, coordinate-based movement, NPC dragon piloted across the map.
8. Planes, Starmaps, Holidays, Marriage shipped as small standalone ports — each a recognized SMAUG feature with a <500 C-LOC surface.
9. Interactive OLC substates (`CON_OEDIT` / `CON_MEDIT` / `CON_REDIT`) shipped: the nanny dispatches `oedit` / `medit` / `redit` menu-driven sessions, not just flat-attribute `oset` / `mset` / `rset` setters.
10. Deferred content gaps from Phase 5 TODO.md closed: full `do_auction` state machine, remaining communication channels (`music` / `newbiechat` / `racetalk` / `wartalk` / `counciltalk` / `guildtalk`), clan commands (`promote` / `demote` / `induct` / `outcast` / `bestow`), skills not yet ported (`bloodlet` / `pounce` / `broach`).

**Explicit non-goals for Phase 6:**

- No new gameplay invented. Every landed item must cite a C source line.
- No rewrite of already-shipped subsystems unless a Phase 6 feature forces it.
- No performance optimization pass; that's a separate future phase.
- Web status page, MXP protocol parsing, DNS resolution (the "infrastructure polish" trio) are out of Phase 6 — tracked but not gated here.

---

## Candidate Inventory

All C-LOC counts verified 2026-04-18 via `wc -l src/<file>.c`. Summary header:

| System | C LOC | C source | Category | Scope signal |
|---|---|---|---|---|
| Hotboot | 854 | `src/hotboot.c` | Infrastructure | Invasive (redesign) |
| Overland | 3752 | `src/overland.c` | Game systems | Largest; depends on nothing |
| Housing | 2853 | `src/house.c` | Game systems | Big; depends on persist |
| Polymorph | 2753 | `src/polymorph.c` | Game systems | Big; depends on combat hooks |
| Archery | 1362 | `src/archery.c` | Game systems | Adds WEAR_* slots + item kinds |
| Combat stances OLC | (included in `stances.c` 1022 but Tranche B already ported the combat half) | `src/stances.c` | Game systems | Small, bounded |
| Dragon flight | 945 | `src/dragonflight.c` | Game systems | Depends on Overland |
| Arena PvP | 358 | `src/arena.c` | Game systems | Smallest; depends on combat only |
| Polymorph + morphset + morphstat | included in Polymorph 2753 | `src/polymorph.c` | Game systems | — |
| Planes | 298 | `src/planes.c` | Game systems | Small; trivial ifdef |
| Holidays | 416 | `src/holidays.c` | Game systems | Small |
| Star maps | 226 | `src/starmap.c` | Game systems | Smallest; pure render |
| Marriage | 362 | `src/marry.c` | Game systems | Small; ifdef MARRIAGE |
| Interactive OLC `CON_OEDIT` / `CON_MEDIT` / `CON_REDIT` | 2160 + 2280 + 1036 = 5476 | `src/ooedit.c` + `src/omedit.c` + `src/oredit.c` | OLC | Largest per-component |
| Editable mudprog editors | (part of OLC) | `src/mpxset.c` et al. | OLC | Small |
| `foldarea` / `unfoldarea` + `.bak` rotation for `savearea` | ~50 C LOC + reuse Phase-3 serializer | `src/build.c:7346,8036,8055` | OLC | **LANDED 2026-04-26** |
| Area vnum repack (`renumber_area`) | — | `src/renumber.c` | OLC | High-risk; deferred (was conflated with `foldarea` in earlier roadmap) |
| Full `do_auction` state machine | ~400 C LOC | `src/act_obj.c` + `src/update.c` | Content | Bounded |
| Channels `music` / `newbiechat` / `racetalk` / `wartalk` / `counciltalk` / `guildtalk` | ~50 C LOC each | `src/act_comm.c` | Content | Trivial (template exists) |
| Clan commands (`promote` / `demote` / `induct` / `outcast` / `bestow`) | ~300 C LOC total | `src/clans.c` | Content | Small |
| Councils (player-facing) | — | `src/clans.c` / `src/deity.c` | Content | Unmeasured |
| Skills: `bloodlet`, `pounce`, `broach` | ~200 C LOC | `src/skills.c` | Content | Small |
| Deities: full prayer + favor | ~400 C LOC | `src/deity.c` | Content | Bounded |
| Testclient: golden-file diff / YAML scenario runner / MCCP zlib / MSDP exposure | — | new | Test infrastructure | Small |

**Note on LOC method:** `wc -l` counts raw source including comments and blank lines. The 854-line `src/hotboot.c` has ~350 LOC of actual code; the ~40-line copyright banner inflates every file by ~2%. Use C LOC as a scope signal, not a budget.

---

## Per-System Scope, Dependencies, Plan-Doc Filenames

Each entry: scope description + ordering rationale + dependencies + proposed plan-doc filename.

### Infrastructure

#### Hotboot / copyover — `plan-phase6-hotboot.md`

- **Scope:** Save all character state and descriptor connection state to a hotboot file, restart the Go process, reload state, reattach descriptors to their still-open TCP sockets. Players experience ~2s pause instead of disconnect.
- **Key C sources:** `src/hotboot.c:599` (`do_hotboot`), `:744` (`hotboot_recover`), `:73` (`save_mobile` — world save), plus area reload.
- **Why Go design differs from C:** C uses `execl()` to replace its process image while keeping open FDs. Go has no equivalent — `os.Exec` replaces the process but we'd lose all `net.Conn` goroutine state. Two viable Go designs: (a) save/shutdown/restart cleanly and have clients reconnect with a cookie (loses the "seamless" promise), or (b) `syscall.Exec`-equivalent + pass socket FDs via `os.StartProcess` with `Files` — but this fights Go's `net.Listener` abstraction. **Design pass required before executable plan.**
- **Dependencies:** None. Unblocks nothing else mandatorily, but every Phase 6 feature that ships *before* hotboot increases the upgrade cost once hotboot lands (more state to serialize).
- **Ordering:** **First, but as a design-exploration plan, not an executable plan.** Prototype the design; if the syscall.Exec+Files path works, promote to executable. Otherwise, re-scope to "save + restart + reconnect cookie."

#### Web status page / MXP / DNS

- **Out of Phase 6.** Noted; not gated on Phase 6 completion.

### Game systems — player-facing

#### Arena PvP — `plan-phase6-arena.md` — **FIRST-CUT PLAN ATTACHED**

- **Scope:** 4 commands: `challenge <player>`, `accept <player>`, `decline <player>`, `withdraw <player>`. Uses `ACT_CHALLENGED` / `ACT_CHALLENGER` flags (already defined). Teleports both players into random arena rooms (vnums 10366-10382 in default data) on accept; sets `PLR_SILENCE` for the duration; tracks global `is_challenge` / `arena_is_busy` state.
- **Key C sources:** `src/arena.c:97-358` (all four commands in 358 lines total).
- **Why good first-cut:** Uses only already-shipped primitives (`GetCharWorld`, `CharFromRoom`/`CharToRoom`, `ACT_CHALLENGED` flags, `PLR_AFK` / `PLR_SILENCE` / `ACT_IS_NPC` checks, existing combat loop). Exercises the `QuickLoginTwo` testclient helper from Tier 9 — drives the first end-to-end two-player PvP scenario in the test suite.
- **Dependencies:** None. All primitives landed in Phases 2-5.
- **Ordering:** **Can ship immediately in parallel with Hotboot design pass.** Lowest risk, smallest surface, most satisfying demo (two players punch each other in a telnet session).

#### Combat Stances OLC completion — `plan-phase6-stances-olc.md`

- **Scope:** `do_stset` full stance-table editor, `do_ststat` stat display, `can_use_stance` prerequisite checks, full `StanceInfo` struct extension (resist/immune/suscept/dodge/parry/max_weight/dual_wield/wait/prerequisite), persistence of non-combat fields via `fwrite_stance`.
- **Key C sources:** `src/stances.c:199-1022`. Tranche B (landed 2026-04-18) read-and-discards the non-combat fields on load; this plan finishes the job.
- **Dependencies:** Tranche B G1 (shipped). `DoMset <victim> <stance-name> <value>` (shipped). Stance data file format (partially shipped).
- **Ordering:** After Hotboot (so stance file persistence benefits from hotboot). Small scope — ~200 Go LOC.

#### Starmaps — `plan-phase6-starmap.md`

- **Scope:** `look sky` port — pure-render terminal image of sun/moon/constellations driven by `time_info.hour`/`.day`/`.month` and `weather.precip`. No persistence; no player mutation.
- **Key C sources:** `src/starmap.c:87-226` (`look_sky`). One call-site in `src/act_info.c:1498` — the C `do_look` dispatches to `look_sky(ch)` when the outdoor player types `look sky`. The current Go `DoLook` (`internal/act/info.go:21-99`) has **no `sky` branch at all** — typing `look sky` falls through to the generic "You do not see that here." path. A new `case strings.EqualFold(arg, "sky"):` branch in `DoLook` plus a new `LookSky(ch)` function are the entry points this plan must add. No reuse of `DoWeather` — that command only describes weather, not constellations.
- **Dependencies:** `types.TimeInfoData` (shipped), `weather.precip` (shipped). Constellation ANSI color codes (shipped).
- **Ordering:** Second-smallest; excellent "warm-up" for a manager lineage new to the codebase.

#### Holidays — `plan-phase6-holidays.md`

- **Scope:** `do_holidays` player command (list holiday table), `do_saveholiday` / `do_setholiday` immortal commands (CRUD on `HOLIDAY_DATA`), `load_holidays` / `save_holidays` persistence to `system/holidays.dat`. Note: the C `get_holiday` function is defined but never called — no actual announce-on-time-tick machinery needs porting (it isn't in C either).
- **Key C sources:** `src/holidays.c:87-416`. Wired into `do_load` admin path at `src/db.c:693`.
- **Dependencies:** `month_name[]` array (not yet in Go; need a Go equivalent — `const/months.go` or `types/time.go` extension). Scanner is shipped. `TimeInfoData.Month` is shipped.
- **Ordering:** After Starmap (both are small). Introduces the holiday data file format.

#### Marriage — `plan-phase6-marriage.md`

- **Scope:** 3 immortal commands: `marry <p1> <p2>`, `divorce <p1> <p2>`, `rings <p1> <p2>` (creates a wedding-ring object). Canonicalises on the already-defined `CharData.Spouse string` field (`character.go:212`, read by LoadPlayer) and removes the orphan `PCData.Spouse` duplicate; adds the matching `SavePlayer` writer that C has but Go lacks. C has a commented-out level-10-minimum check at `src/marry.c:128-132` — Marriage plan resolved Q1 as omit (structurally unreachable).
- **Key C sources:** `src/marry.c:79-362`. `#ifdef MARRIAGE` gated in C; port unconditionally.
- **Dependencies:** Both `CharData.Spouse` (canonical, read by LoadPlayer at `persist/player.go:285`) and `PCData.Spouse` (orphan, never read/written) are defined today; the marriage plan's G0 removes the orphan and G0b adds the missing SavePlayer writer. Ring vnums in C are `OBJ_VNUM_DIAMOND_RING = 100` and `OBJ_VNUM_WEDDING_BAND = 101` (`src/mud.h:2071-2072`). **Earlier roadmap text said `STEEL_RING` — that was wrong; corrected 2026-04-18.** Neither vnum is present in shipped `.are` files — Marriage plan Open Question 2.
- **Ordering:** After Holidays. Orphan-field cleanup + SavePlayer writer + new commands; no net new schema additions.

#### Planes — `plan-phase6-planes.md`

- **Scope:** `do_plist` / `do_pstat` / `do_pset` commands. `PLANE_DATA` list with name-only payload (298 C LOC, most of which is banner + whitespace — actual code is ~150 LOC). The `room.Plane` back-reference already exists as a struct field (`types/room.go:26`); this plan populates it via the planes loader and `check_planes` orphan-assignment.
- **Key C sources:** `src/planes.c:51-298`.
- **Dependencies:** `RoomIndexData.Plane *PlaneData` is already present at `internal/types/room.go:26`; `PlaneData` struct is defined at `room.go:112` (audit 2026-04-18). Persistence at `system/planes.dat` still to wire.
- **Ordering:** Low priority — feature is essentially cosmetic (named groupings of rooms). Ship when building-tool completeness pressure appears.

#### Polymorph — `plan-phase6-polymorph.md`

- **Scope:** `MORPH_DATA` struct + per-morph stat overrides (hp / mana / move / hitroll / damroll as DiceString). `do_morph` / `do_unmorph` / `do_morphset` / `do_morphstat` / `do_morphcreate` / `do_morphdestroy` commands. Mudprog `mpmorph` / `mpunmorph` — hook points landed, bodies not. Combat hooks (stat application on morph, removal on unmorph). Persistence of morph table + per-player current morph state.
- **Key C sources:** `src/polymorph.c:72-2753`. `MORPH_DATA` defined in `src/mud.h`.
- **Dependencies:** `CharData.Morph` is already `*CharMorph` at `internal/types/character.go:66`; `MorphData` (the template struct) is fully defined at `character.go:306` with all C fields present. **Schema is done; no struct additions needed.** `Morph` table (new loader, new file). Stat-override application path in combat (new hook in `handler.AffectModify` or equivalent). Per-player current-morph persistence (new `#MORPH` block in playerfile).
- **Ordering:** Mid-Phase-6. Polymorph is a canonical SMAUG feature but high scope; ship after hotboot and arena.

#### Archery — `plan-phase6-archery.md`

- **Scope:** New `WEAR_LODGE_RIB` / `WEAR_LODGE_ARM` / `WEAR_LODGE_LEG` slots (arrow-lodge-in-victim); `WEAR_MISSILE_WIELD`, `ITEM_PROJECTILE`, `ITEM_QUIVER` already defined in Go (audit 2026-04-18). `do_draw` / `do_fire` / `do_dislodge` commands. Ranged hit-roll path. Arrow-lodges-in-victim mechanic.
- **Key C sources:** `src/archery.c:125-1362`. Entry points `do_draw` / `do_dislodge` / `do_fire` cited above. `do_fire` is currently aliased to `do_throw` in Go.
- **Dependencies:** WEAR_* slot enum extension (`internal/types/enums.go`) — `WEAR_MISSILE_WIELD = 21` already exists (audit 2026-04-18); need to add `WEAR_LODGE_RIB` / `WEAR_LODGE_ARM` / `WEAR_LODGE_LEG` (C `src/mud.h:2448-2450`, gated behind `#ifdef ENABLE_ARCHERY`). **Item types are already present**: `ITEM_PROJECTILE` and `ITEM_QUIVER` at `enums.go:629-630`. Combat loop hook for ranged attack (need to verify if Go's `OneHit` has a WEAR_MISSILE_WIELD path — probably doesn't).
- **Ordering:** Mid-Phase-6, after polymorph. Cross-cutting change to `types/` and combat.

#### Housing — `plan-phase6-housing.md`

- **Scope:** `HOME_DATA` struct with `vnum[MAX_HOUSE_ROOMS]` array (room-ownership mapping). `do_house` / `do_gohome` / `do_residence` / `do_accessories` / `do_homebuy` / `do_sellhouse` commands. Per-room description customization via editor (builds on `EditorSave` callback pattern from Tier 12). Persistence at `system/houses.dat` + per-home room file. Guest access list.
- **Key C sources:** `src/house.c:60-2853`. Entry points listed above.
- **Dependencies:** `EditorSave` callback pattern (shipped Tier 12). Room ownership model on `RoomIndexData` (needs new `OwnedBy string` field). Persistence — one of the larger new formats in Phase 6.
- **Ordering:** Late Phase-6. 2853 C LOC is the second-largest game system; only overland is larger.

#### Dragon flight — `plan-phase6-dragonflight.md`

- **Scope:** `do_call` / `do_release` / `do_fly` / `do_land` / `do_landing_sites` / `do_setlanding` commands. NPC dragon paired with a PC rider; coordinate-based movement across overland map. Uses `ch->map` / `ch->x` / `ch->y` fields. **Go-state correction (2026-04-18):** the coordinate fields (`X`, `Y`, `Map`, `Sector`) already exist on `CharData` at `internal/types/character.go:226-229`; what's unset is their *population* (no overland loader, no sector-type lookup, no `get_terrain` equivalent). Shipping dragonflight still requires Overland first, but for the structural reason (no map data to move through), not for the schema reason.
- **Key C sources:** `src/dragonflight.c:359-945`.
- **Dependencies:** **Overland maps** (hard blocker — every dragonflight command reads `ch->map`).
- **Ordering:** **Cannot start until Overland lands.** Parallelize design work with Overland implementation but do not dispatch a dragonflight execution plan until Overland has shipped.

#### Overland — `plan-phase6-overland.md`

- **Scope:** 3x 1000x1000 sector map files (`map1.raw` / `map2.raw` / `map3.raw`). `ENTRANCE_DATA` / `LANDMARK_DATA` / `MAPRESET_DATA` load+save. `do_survey` / `do_coords` / `do_landmarks` / `do_setmark` / `do_setexit` / `do_mapresets` / `do_mreset` / `do_mapedit`. Sector-type lookup. ANSI rendering of a view window around the player.
- **Key C sources:** `src/overland.c:839-3363`. 3752 C LOC — the single largest game system left.
- **Dependencies:** `CharData.Map`/`X`/`Y`/`Sector` schema fields are already present at `internal/types/character.go:226-229` (audit 2026-04-18). What's missing is the population layer: a new raw-binary file format for maps (NOT the text `.are` format — fixed-size sector byte per tile), `get_terrain`/`map_names`/sector-type lookup, and the set of `do_survey` / `do_coords` / etc. commands. Compatibility with existing `db/maps/` data if present.
- **Ordering:** Late Phase-6. Blocks Dragonflight. Consider splitting into two plans (loader + display, then editor+reset).

### OLC / builder

#### Interactive OLC substates — `plan-phase6-olc-oedit.md`, `plan-phase6-olc-medit.md`, `plan-phase6-olc-redit.md`

- **Scope per editor:** A `CON_OEDIT` / `CON_MEDIT` / `CON_REDIT` substate for the nanny; menu-driven editing of a target vnum with per-field subcommand dispatch. `oedit_parse` / `medit_parse` / `redit_parse` in C each run 1-2k LOC of menu state machine.
- **Key C sources:** `src/ooedit.c` (2160), `src/omedit.c` (2280), `src/oredit.c` (1036). Dispatch at `src/smaug.c:1641-1652`.
- **Unblocked by:** Tier 12 `EditorSave` callback pattern (for the text-editor-within-menu "description" path). `WithPrompt` seam in the testclient harness (Tier 5) ready for prompt-tracking in tests.
- **Dependencies:** No external. Blocks *nothing* — the flat `oset` / `mset` / `rset` commands (Phase 4b) remain. This is pure builder UX improvement.
- **Ordering:** Parallel to overland/housing. Ship all three editors together or interleave; they share the nanny-dispatch pattern.

#### Editable mudprog editors — `plan-phase6-olc-mpedit.md` — **LANDED 2026-04-26**

- **Status:** LANDED 2026-04-26 across 5 waves. Closes Phase-6 Wave-D's interactive-OLC stripe.
- **Scope shipped:** `mpedit` / `opedit` / `rpedit` accept the C-faithful `<victim> <command> [number] <program> <value>` shape with full add/delete/insert/edit/list arms; EditorSave-callback integration (no `CON_MPEDIT` substate per design); Q1/Q2 rpedit insert C-bug fix applied; 52-entry `MProgFlagNames` table mirrors C `mprog_flags[]`; `progEditOpenEditor` shared closure handles Tier-12 `/s` round-trip with optional progtypes-bitmask rebuild on edit. A1-A24 covered, M1-M14 mutation gates verified.
- **Plan:** `smaug-go/doc/plan-phase6-olc-mpedit.md` — see §Completion Record for per-wave commit hashes.

#### `foldarea` / `unfoldarea` — `plan-phase6-foldarea.md` — **LANDED 2026-04-26**

- **Scope shipped:** `foldarea <filename>` saves a named area to disk via the shared `writeAreaToDisk` helper (which both `savearea` and `foldarea` use). `.bak` rotation added at the same seam — the live file is rotated to `<file>.bak` before each save, mirroring C `fold_area` at `src/build.c:7369-7370`. `unfoldarea` ships scoped DOWN to a "use hotboot" guidance message because `internal/persist/area.go:48-82`'s `loadAreaFile` is not re-entrant. Boot regs at `LEVEL_IMMORTAL` / `POS_DEAD`. **Note:** the original roadmap entry conflated `foldarea` (save-by-filename) with vnum repacking (`renumber_area` in `src/renumber.c`) — they are separate concerns; vnum repack remains deferred.
- **Plan:** `smaug-go/doc/plan-phase6-foldarea.md` — see §Completion Record for per-wave commit hashes.

### Content gaps (carried over from Phase 5 TODO.md)

#### Full `do_auction` state machine — `plan-phase6-auction.md`

- **Scope:** `DoAuction <item>` start auction; `DoAuction bid <price>` / `stop` / `list`. Item escrow during auction. Gold transfer. Per-pulse tick in `update.go` to advance auction phases. `BroadcastAuction` helper already shipped (Tier 9).
- **Key C sources:** `src/act_obj.c:3775+` (the `#else` branch `do_auction` — non-gold-silver-copper variant), `src/update.c:2927-3286` (`auction_update` at line 2927 under `#ifdef ENABLE_GOLD_SILVER_COPPER`; the non-GSC `auction_update` starts at line 3183). **Audit note (2026-04-18):** the earlier citation `update.c:2886-3286` was mid-function inside `reboot_update`; corrected to the actual `auction_update` start.
- **Dependencies:** `BroadcastAuction` (shipped). `update.go` auction tick slot (new).
- **Ordering:** After any of the big game systems land. Small-to-medium scope.

#### Remaining channels — `plan-phase6-channels-extra.md`

- **Scope:** `music` / `newbiechat` / `racetalk` / `wartalk` / `counciltalk` / `guildtalk` — each follows the `DoImmtalk` / `DoGtell` template (Tier 9). Template is established; port is mechanical.
- **Dependencies:** None new.
- **Ordering:** Anytime. Trivial.

#### Clan officer commands — `plan-phase6-clan-officer.md`

- **Scope:** `promote` / `demote` / `induct` / `outcast` / `bestow`. Requires clan-leader privilege check. Persistence via `SaveClan`.
- **Dependencies:** `SaveClan` in `persist/subsystems.go` (currently tracked as follow-up in TODO.md).
- **Ordering:** After `SaveClan` ships. Small scope.

#### Missing skills — `plan-phase6-skills.md`

- **Scope:** `bloodlet`, `pounce`, `broach`.
- **Dependencies:** None new.
- **Ordering:** Anytime. One-per-plan or bundled.

#### Deities full prayer — `plan-phase6-deity-prayer.md`

- **Scope:** Full `do_pray` prayer; per-deity favor modifiers beyond `mpFavor`; deity-specific effects (resurrection, smite, etc.).
- **Ordering:** Mid-Phase-6.

### Test infrastructure

Queued in TODO.md already. Each is a small delta. Not blocking.

---

## Recommended Execution Order

Rationale: Hotboot's design pass must start early (it influences every other state-save schema), but execution of small self-contained ports (Arena, Starmap, Holidays, Marriage) should happen in parallel to keep momentum. The biggest systems (Overland, Housing) ship last because they are rewrites of large surface area and benefit from a stable hotboot already being in place. Dragonflight must wait on Overland.

**Wave 0 — Design only:**
1. Hotboot design doc (`plan-phase6-hotboot.md` — design-exploration, not executable).

**Wave 1 — Small self-contained (parallel, any manager may pick):**
2. **Arena** (`plan-phase6-arena.md`) — **first-cut plan attached, ready to dispatch.**
3. Starmap (`plan-phase6-starmap.md`).
4. Holidays (`plan-phase6-holidays.md`).
5. Marriage (`plan-phase6-marriage.md`).
6. Planes (`plan-phase6-planes.md`).
7. Missing skills (`plan-phase6-skills.md`).
8. Remaining channels (`plan-phase6-channels-extra.md`).

**Wave 2 — Infrastructure + mid-scope game systems:**
9. Hotboot executable plan (whichever design from Wave 0 won).
10. Combat Stances OLC completion (`plan-phase6-stances-olc.md`).
11. Full `do_auction` state machine (`plan-phase6-auction.md`).
12. Clan officer commands (`plan-phase6-clan-officer.md`) — after `SaveClan` lands.

**Wave 3 — Interactive OLC:**
13. `CON_REDIT` (simplest of the three; redit already has more subcommand support in Go than the others).
14. `CON_OEDIT` / `CON_MEDIT` (parallel after redit proves the nanny-dispatch pattern).
15. Editable mudprog editors.

**Wave 4 — Large game systems:**
16. Polymorph.
17. Archery.
18. Housing.
19. Overland.

**Wave 5 — Overland-dependent:**
20. Dragonflight.

**Wave 6 — Content polish:**
21. Deity prayer + favor.
22. Test infrastructure polish (golden-file, scenario runner, etc.).

Each wave's items are independent. Within a wave, assign one item per manager lineage; the orchestrator can fan out multiple managers in parallel.

**Soft-shared footprint within Wave 1 (audit note 2026-04-18):** every Wave 1 plan registers new commands in `internal/boot/boot.go`. These are additive, one-line-per-command registrations, so N-way parallel merges converge cleanly in practice — but the orchestrator should sequence merges serially (not concurrent writes to the same file) to avoid conflict artifacts. Starmap is the only Wave 1 item that modifies an *existing* file beyond boot-registration (it adds a `sky` branch to `DoLook` in `internal/act/info.go`); no other Wave 1 item touches `info.go`, so this is safe.

---

## Gaps in the Existing Go Port Relevant to Phase 6

1. **No `OwnedBy` room ownership model** — housing needs per-room "owner player name" or equivalent. Current `RoomIndexData` has no such field. (Confirmed 2026-04-18 audit.)
2. ~~No `Morph` field on `CharData`~~ **Correction (2026-04-18 audit):** `Morph *CharMorph` is already present at `internal/types/character.go:66`. The `MorphData` struct (morph template) is defined at `character.go:306` with 30+ fields including Name, ShortDesc, LongDesc, Damroll, Hit, Hitroll, Mana, Move, AffectedBy, Class, Race, Obj[3], Vnum, etc. The schema scaffolding is in place; what's missing is the Morph-table loader/saver, the command set (`do_morph` / `do_unmorph` / `do_morphset` / etc.), the `affect_modify`-style stat application path, and the mudprog `mpmorph` / `mpunmorph` bodies.
3. ~~No `Map` / `X` / `Y` coordinate fields on `CharData`~~ **Correction (2026-04-18 audit):** `CharData` already has `X`, `Y`, `Map`, `Sector int` fields at `internal/types/character.go:226-229` under the `// Overland` comment, plus `MapData *MapData` on `RoomIndexData` at `room.go:25`. The field schema is in place; what's missing is the overland loader, renderer, and command set.
4. ~~No `Plane` back-reference on `RoomIndexData`~~ **Correction (2026-04-18 audit):** `Plane *PlaneData` is already present at `internal/types/room.go:26`, and `PlaneData` struct is defined at `room.go:112`. What's missing is the planes loader, the command set (`do_plist` / `do_pstat` / `do_pset`), and `check_planes` wiring.
5. ~~No `Spouse` field on `PCData`~~ **Correction (2026-04-18 audit):** `PCData.Spouse string` is already present at `internal/types/pcdata.go:130`. What's missing is the command set (`do_marry` / `do_divorce` / `do_rings`), persistence emit/read for the field in `SavePlayer`/`LoadPlayer`, and the wedding-ring object vnum constants.
6. **`act.WorldRef` singleton pattern** — works for the current scope but hotboot's world-save path must serialize through this or take the world explicitly. Some current callers would need a refactor. Auditable during hotboot design pass.
7. **No `month_name[]` equivalent in Go** — holidays needs this (also nice-to-have for `DoTime`).
8. **Nanny dispatch has no `CON_OEDIT` / `CON_MEDIT` / `CON_REDIT` branches** — `internal/game/loop.go:261` is the `CON_EDITING` case; the dispatch-input switch only has `CON_PLAYING` and `CON_EDITING` arms with a `default: g.nanny(d, line)` fallthrough. The constants `CON_REDIT`, `CON_OEDIT`, `CON_MEDIT` are DEFINED in `types/enums.go:89-91` (iota-assigned values 21/22/23) but the input-dispatch switch has no cases for them, so menu-state input never reaches a dedicated handler.
9. **No WEAR_LODGE_RIB / WEAR_LODGE_ARM / WEAR_LODGE_LEG slots in Go** — archery-only concern. Note: `WEAR_MISSILE_WIELD` (value 21) IS already defined at `types/enums.go:787`; only the three arrow-lodged-in-victim slots are missing. In C these three slots live behind `#ifdef ENABLE_ARCHERY` at `src/mud.h:2448-2450`.
10. ~~No new-item-type slots for quivers / projectiles~~ **Correction (2026-04-18 audit):** `ITEM_PROJECTILE` and `ITEM_QUIVER` are already defined at `types/enums.go:629-630`. No new item-type constants are needed for archery — only the C-side behavior (ammunition tracking, quiver autoload) remains to port.
11. **No raw-binary file format in `persist/`** — Scanner is line-oriented text. Overland maps are raw binary. Needs a new loader shape.
12. **No hotboot-capable graceful shutdown path** — `net/server.go` tears down all connections on shutdown. Hotboot needs a different path that preserves sockets.

None of these gaps block Wave 0-1. Each is a per-feature prerequisite that the relevant plan doc will enumerate.

---

## Risk Analysis

**Low risk (small surface, all primitives exist):**
- Arena (358 LOC, first-cut plan attached)
- Starmap (226 LOC)
- Planes (298 LOC)
- Marriage (362 LOC)
- Holidays (416 LOC)
- Remaining channels (~50 LOC each)
- Missing skills (~200 LOC total)

**Medium risk (needs new schema fields or cross-package changes):**
- Combat stances OLC (needs `StanceInfo` extension)
- Full auction state machine (needs `update.go` hook)
- Interactive OLC substates (needs nanny dispatch; per-editor ~1k Go LOC)
- Polymorph (`Morph` field already present; needs Morph-table loader + combat stat-application hook + command set)
- Archery (needs `WEAR_LODGE_*` slots; `WEAR_MISSILE_WIELD`, `ITEM_PROJECTILE`, `ITEM_QUIVER` already defined; needs combat ranged-attack hook)
- Clan officer commands (needs SaveClan)

**High risk (invasive / protocol-cross-cutting):**
- **Hotboot** — C design not portable; requires fresh design pass. Risk: picking the wrong design wastes significant implementation work.
- **Overland** — 3752 C LOC, new binary file format, plus sector-type cross-compatibility. (Coordinate fields `X`/`Y`/`Map`/`Sector` already defined on `CharData` — the large surface is the loader + renderer + command set + map editor, not the schema.) Multiple plan docs may be required.
- **Housing** — 2853 C LOC, new persistence schema, interacts with every room-ownership path. Invasive.

**Dragonflight** is medium-scope (945 LOC) but high-risk *because it blocks on Overland*. Slip in Overland slips Dragonflight.

---

## Cross-Plan Dependencies and Ordering Constraints

Summary table for the orchestrator dispatching future Phase 6 managers.

| Plan | Hard blockers | Soft blockers (preferred order) |
|---|---|---|
| `plan-phase6-arena.md` | None | — |
| `plan-phase6-starmap.md` | None | — |
| `plan-phase6-holidays.md` | None | `month_name[]` port (bundle in the plan) |
| `plan-phase6-marriage.md` | None | `PCData` schema change — bundle other PCData additions if scheduled |
| `plan-phase6-planes.md` | None | `RoomIndexData.Plane` field addition |
| `plan-phase6-skills.md` | None | — |
| `plan-phase6-channels-extra.md` | None | — |
| `plan-phase6-hotboot.md` (design) | None | Run early; design informs every later state-serialization decision |
| `plan-phase6-hotboot.md` (executable) | Hotboot design complete | — |
| `plan-phase6-stances-olc.md` | Tranche B G1 (shipped) | Bundle with any `StanceInfo` consumer |
| `plan-phase6-auction.md` | None | `update.go` tick-slot refactor if helpful |
| `plan-phase6-clan-officer.md` | `SaveClan` port | — |
| `plan-phase6-olc-redit.md` | Tier 12 `EditorSave` (shipped) | Ship first of the three editors — smallest |
| `plan-phase6-olc-oedit.md` | redit nanny-dispatch pattern | — |
| `plan-phase6-olc-medit.md` | redit nanny-dispatch pattern | — |
| `plan-phase6-olc-mpedit.md` | Interactive OLC substates | — |
| `plan-phase6-polymorph.md` | None (new schema work bundled) | Hotboot shipped (for graceful re-deployment) |
| `plan-phase6-archery.md` | None | `WEAR_*` slot enum extension |
| `plan-phase6-housing.md` | None | Hotboot shipped (persistence complexity increases without it) |
| `plan-phase6-overland.md` | None | Hotboot shipped (map state must survive hotboot) |
| `plan-phase6-dragonflight.md` | **Overland shipped** | — |
| `plan-phase6-deity-prayer.md` | None | — |

**Ordering heuristic:** "Prefer shipping Hotboot before Overland / Housing because their persistence scope should land once the hotboot format is settled." But Arena / Starmap / Holidays / Marriage / Planes / Stances-OLC / Auction / Channels-extra / Skills are all unblocked and can be dispatched the moment a manager is free.

---

## Open Questions for the Orchestrator / Human

1. **Hotboot design:** syscall.Exec + socket-FD passing vs save+restart+reconnect-cookie. Human should weight "seamless experience" vs "operational simplicity." Design plan will present both with tradeoffs.
2. **Overland splitting:** one 3752-LOC plan or two plans (loader+display, then editor+reset)? Recommendation: split.
3. **Archery prerequisites:** does Go need `WEAR_MISSILE_WIELD` as a new equipment slot, or can it reuse `WEAR_WIELD` with an item flag? C uses a distinct slot. Recommendation: match C.
4. **Marriage level-10 gate:** C has commented-out code; port it (uncommented) or omit? Recommendation: omit (unreachable C code is unreachable intent).
5. **Planes priority:** is this feature worth porting at all? It's mostly cosmetic (named grouping of rooms) and the C source shows very little actual usage. Can defer indefinitely.
6. **`foldarea`:** RESOLVED 2026-04-26 — landed as save-by-filename + `.bak` rotation (see §`foldarea` / `unfoldarea`). Vnum repack (`renumber_area`) remains deferred separately.
7. **Starmap constellations:** the C data table has FIXED constellation positions (static arrays in `src/starmap.c:59-79`). Go port should preserve them verbatim. Confirm policy of "no gameplay invented" — this table IS the gameplay.

---

## Appendix: File Index for Future Plan Authors

Every Phase 6 plan doc should follow the shape of `plan-tranche-b.md` / `plan-tranche-c.md`:

1. **Problem** — state the gap with C citations.
2. **C Reference (authoritative)** — line-cited entry points, key functions, algorithm notes.
3. **Go Current State** — what's shipped, what's stubbed, what's missing.
4. **Go Design** — approach chosen + alternatives rejected with reasoning.
5. **Task Groups** — test-first + file paths + mutation-verify steps using `Edit`-only revert.
6. **Acceptance Criteria** — mechanically verifiable conditions for "done."
7. **Scope Cuts / Deferrals** — what is explicitly NOT in this plan.
8. **Open Questions** — pre-dispatch blockers requiring human input.
9. **Risk Analysis** — what could go wrong; mitigations.
10. **Adversary Verification Notes** (appended after plan adversary pass).
11. **Completion Record** (appended after work lands).

Reference plans, from simplest to most complex:

- `plan-tranche-b.md` (combat stances + timer dispatch + 7 ifchecks + wordlist match — 5 task groups, 18 acceptance criteria)
- `plan-tranche-c.md` (util.Act color per-call + 33-site migration + PLR_BLANK + XP-on-skill-gain + adept-cap — 8 task groups, 10 acceptance criteria)
- `plan-channels.md` (immtalk/gtell/auction-stub — 4 task groups, 18 acceptance criteria)
- `plan-combat-depth.md` (multi-attack cascade + prof bonus + stance — 9 task groups, 12 acceptance criteria)

Each of those is a manager-lineage-sized unit of work. Future Phase 6 plans should aim for the same granularity.
