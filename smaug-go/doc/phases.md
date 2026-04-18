# Implementation Phases

This document defines all phases of the SMAUG C-to-Go port.

**Testing mandate:** All phases follow TDD. New code gets tests written first. Existing code gets backfill tests verified via mutation (break code → test fails → revert → test passes). See `CLAUDE.md` at repo root for full TDD workflow.

---

## Phase 1: Skeleton — Telnet Login + Walk Around

**Goal:** Connect via telnet, log in, walk between rooms, see descriptions.

**Status:** Complete. 36 source files (~8,891 lines), 17 test files (~3,636 lines), 352 test cases. Boot loads 1,909 rooms, 4,299 exits, 505 mob templates, 821 obj templates, 406 mob instances, 710 obj instances, 1,603 helps, 325 skills, 17 classes, 15 races.

**Deliverables (all complete):**
- All core data structures ported from `mud.h` (`types/` package)
- Utility functions (string manipulation, dice rolls, logging)
- File format scanner and area file parser (with `#HELPS` loading)
- Class/race/skills data file loaders
- World state container replacing C globals
- TCP server with goroutine-per-connection architecture
- Minimal telnet negotiation (echo on/off)
- ANSI color code processing
- Pulse-based game loop (4 Hz, single goroutine)
- Command interpreter with prefix matching
- Full login state machine: returning players (load + password verify) and new players (character creation with sex/race/class)
- Player file save/load in existing SMAUG format (save on quit, auto-create directories)
- Area resets (mob/object instantiation via handler package)
- Help file system (1,603 entries from area files)
- Prompt system with token substitution (`%h/%H`, `%m/%M`, `%v/%V`, `%g`, `%a`, `%x`, `%r`)
- Basic commands: look, quit, say, score, who, help, commands, inventory, equipment, movement (10 directions)

**Verification:** Telnet in, create a new character (name/password/sex/class/race) or load an existing one, walk around rooms loaded from real `.are` files, see room descriptions with NPCs and objects, use help system, save/quit/reload with all state preserved.

---

## Phase 2: Core Gameplay — Combat, Items, Skills

**Goal:** A playable game with full combat, inventory management, and basic spells.

**Status:** Complete. All 10 task groups done. 55 source files, 27 test files, 455 test cases. See `phase2-completed.md` for full record.

**Deliverables:**

### Handler Layer (`handler/`)
- `char_to_room`, `char_from_room`, `extract_char` with extraction queue
- `obj_to_char`, `obj_from_char`, `obj_to_room`, `obj_from_room`, `obj_to_obj`
- `equip_char`, `unequip_char`
- `affect_to_char`, `affect_remove`, `affect_strip`, `affect_join`
- Find functions: `get_char_room`, `get_char_world`, `get_obj_carry`, `get_obj_wear`, `get_obj_here`, `get_obj_world`

### Movement (`act/move.go`)
- Sector-based movement costs
- Door commands: open, close, lock, unlock, pick
- Traps on exits
- Flying/swimming/underwater checks
- Enter/leave portal support

### Object Commands (`act/obj.go`)
- get, drop, put, give, sacrifice
- wear, remove, hold, wield
- eat, drink, fill, empty
- examine, compare, brand

### Information Commands (expand `act/info.go`)
- Enhanced look (at characters, objects, in containers, in direction)
- examine, consider, where, time, weather
- Attribute score display matching C format

### Communication (`act/comm.go`)
- say, tell, reply, yell, shout, gossip
- group tell, clan talk, council talk
- emote, pmote
- Channels system with deaf flags

### Combat (`combat/`)
- `violence_update()` — per-pulse combat round processing
- `one_hit()` — single attack resolution (hit roll vs AC, damage calc)
- Multi-attack and dual wield
- Special mob attacks and defenses
- Flee and wimpy
- Death handling: corpse creation, gold/item transfer, experience loss
- Player killing rules and restrictions

### Magic (`magic/`)
- Spell casting: mana check, target resolution, saving throws
- ~20 initial spell implementations (cure light/serious/critical, magic missile, fireball, sanctuary, blindness, poison, sleep, charm, detect evil/invis/magic/hidden, armor, shield, bless, curse, dispel magic, identify)
- Affect application with duration tracking
- Affect tick-down and wear-off messages
- Spell component system

### Game Updates (`game/update.go`)
- `char_update()` — HP/mana/move regeneration, hunger/thirst, poison tick, affect duration
- `obj_update()` — item timers, decay, corpse rot
- `mobile_update()` — NPC wandering, scavenging, hunting
- `aggr_update()` — aggressive mob targeting
- Area reset execution (respawn mobs/objects on timer)

### Player Persistence
- Save on quit, periodic autosave
- Save inventory and equipment
- Save affects (active spells)
- Save quest progress

### Attribute Bonus Tables (`types/` or `const/`)
- Port `str_app`, `int_app`, `wis_app`, `dex_app`, `con_app`, `cha_app`, `lck_app` tables from `const.c`

**Verification:** Full gameplay loop — login, explore, pick up items, equip, fight mobs, cast spells, gain XP, level up, save/quit/reload with all state preserved.

---

## Phase 3: Advanced Systems — MUD Progs, OLC, Subsystems

**Goal:** Feature parity with C version for world interaction and building.

**Status:** Complete. All 12 task groups done. **`phase3-completed.md` is authoritative** — this section summarizes what actually shipped; the original plan was broader. Additional spell/skill coverage and some subsystem commands (e.g. full deity prayer, clan promote/demote, councils) continued to land in Phases 4 and 5.

**Deliverables as shipped:**

### Protocol Support (`net/`)
- `mccp.go` — MCCP2 compression via `compress/zlib` (stdlib)
- `msdp.go` — MUD Server Data Protocol variable reporting
- `mssp.go` — MUD Server Status Protocol for listing services

### MUD Programs (`mudprog/`)
- `driver.go` — mprog_driver interpreter with if/else/endif, nested ifs
- `triggers.go` — mob-prog triggers (act, greet, all_greet, speech, random, fight, death, hitprcnt, entry, give, bribe); deeper trigger coverage (obj/room progs, hour/time, void, sell, login, tell) landed in Phase 5 Tier 3
- Variable substitution ($n, $N, $t, $T, $i, $I, $e, $E, $j, $J, $k, $K, $o, $O, $p, $P, $r, $R)
- If-checks: ~25 in Phase 3; ~77 additional if-checks landed in Phase 5 Tier 3
- Object progs and room progs (Phase 5 Tier 3); sleeping programs (Phase 5 Tier 3)

### Immortal Commands (`act/wiz.go`)
- goto, transfer, at, bamfin, bamfout
- slay, purge, advance, restore
- mstat, ostat, rstat (stat mob/obj/room)
- force, peace
- mfind, ofind, mwhere, owhere
- invis, holylight, freeze, silence, snoop, echo, recho

  (`switch`/`return`, `wizlock`, `shutdown`, `reboot`, `wizhelp`, `aecho`, `hell`, `log`, `mpstat`/`opstat`/`rpstat`, `deny`, `pardon`, `disconnect`, `mortalize` all landed in Phase 5 Tier 4. `mset`/`oset`/`rset`/`aset`/`astat` landed in Phase 4b. `rfind` is not implemented.)

### OLC — Online Creation (`act/olc*.go`)
- redit — room editor (core subcommands in Phase 3; ~10 additional in Phase 5 Tier 4)
- Flat `mset`/`oset`/`rset` attribute setters (Phase 4b)
- Interactive `oedit`/`medit` flat-dispatch form (Phase 5 Tier 4)
- `mpedit`/`opedit`/`rpedit` as read-only inspectors (Phase 5 Tier 4)
- `rdelete`/`odelete`/`mdelete` with confirmation (Phase 5 Tier 4)
- Area save (write modified areas back to `.are` files) via `persist/area_write.go`

  (Interactive `CON_OEDITING`/`CON_MEDITING` substates and editable prog editors remain deferred to Phase 6.)

### Spells/Skills
- ~20 spells in Phase 3 (cure light/serious/critical, magic missile, fireball, sanctuary, bless, curse, poison, blindness, armor, shield, dispel magic, plus detect family, sleep, charm, identify, etc.)
- 12 combat/utility skills in Phase 3 (backstab, bash, kick, disarm, rescue, sneak, hide, steal, pick, scan, aid, recall)
- ~80 additional spells via Phase 5 Tiers 1/2/3/4 (`spell_smaug` data-driven dispatcher + breath/teleport/unique direct ports)
- ~27 additional skills in Phase 5 Tier 4 (unarmed, circle/gouge/stun/grapple/cleave/hitall/berserk, meditate/trance/search/detrap/dig, crafting, admin)
- Skill improvement on use (correct C-parity formula in Phase 5 Tier 1)
- **Not yet ported:** weapon proficiency bonus in `OneHit` (tracked in `TODO.md`)

### Subsystems
- Clans: list, info, clantalk, join, leave, deposit/withdraw (Phase 3 + Tier 2)
  (`promote`/`demote`/`induct`/`outcast`/`bestow` not implemented; tracked for Phase 6)
- Deities: list, devote (Phase 3)
  (Full prayer, favor tracking beyond mudprog `mpFavor`, deity-specific effects not implemented)
- Councils: not implemented as player-facing commands
- Shops: buy, sell, list, value — NPC shop interaction
- Repair shops: repair, estimate, appraise (Phase 5 Tier 2)
- Quest system: request, complete, list, buy, info, time, points (Phase 4b)
- Boards and notes: list, read, write, post, remove, mail targeting (Phase 3 + Tier 2)
- Pager: page long output with --more-- prompts
- String editor: for building descriptions in OLC
- Languages: speak, learn, language scrambling (Phase 5 Tier 2)
- Banking: deposit, withdraw, balance (Phase 4b)
- Bans: site, class, race (Phase 4b + Tier 2)

**Verification:** Builder can log in, create areas with OLC, add mob progs, test them. Players can use shops, join clans, interact with shipped game systems.

---

## Phase 4: Quality + Feature Systems

**Goal:** Quality hardening and feature gap closure.

**Status:** Complete. Phase 4a (quality: G1–G5) and Phase 4b (features: G6–G10, G12) done. G11 (hotboot) deferred to Phase 5. 78 source files, 62 test files, 1,372 test cases across 13 packages. See `phase4-plan.md`, `phase4a-completed.md`, `phase4b-completed.md` for full records.

**Deliverables (all complete):**

### Phase 4a — Quality
- G1: Documentation verification pass
- G2: Live integration testing (9 end-to-end tests)
- G3: Code coverage push (all packages at or near targets)
- G4: Go idiom audit (10 findings, all resolved)
- G5: Security vulnerability audit (16 findings, 14 fixed)

### Phase 4b — Features
- G6: OLC set commands — mset, oset, rset, aset, astat
- G7: Quest system — request/complete/list/buy/info/time/points, random mob-slay quests
- G8: Banking — deposit, withdraw, balance with ACT_BANKER NPC
- G9: Ban system — site bans with prefix/suffix wildcards, persist load/save
- G10: Mob tracking/hunting — BFS pathfinding, do_track, HuntVictim in mobileUpdate
- G12: Remaining player commands — position (rest/sit/stand/sleep/wake), skills/spells/practice, quaff/recite/brandish/zap, follow/group/order/assist, mount/dismount

**Verification:** All tests pass, server boots and accepts connections, all new commands work via telnet.

---

## Phase 5: Optional Systems (candidates)

**Goal:** Large, self-contained optional systems and infrastructure improvements.

**Candidates (~19,000+ lines of C total):**

### Infrastructure
- **Hotboot/copyover** — seamless server restart. C uses `exec()` + fd inheritance; Go needs a different approach (save state + graceful restart + auto-reconnect). From `src/hotboot.c`.
- DNS resolution — `net.LookupAddr` for hostnames
- Web status page — embedded HTTP server for game status

### Game Systems
- **Overland maps** (3,752 lines) — 1000x1000 tile maps, ANSI rendering, landmarks
- **Player housing** (2,853 lines) — apartments, room customization, guests
- **Polymorph** (2,753 lines) — form shifting with stat mods
- **Archery** (1,362 lines) — ranged combat, arrow lodging
- **Combat stances** (1,022 lines) — 10 fighting postures with stat effects
- **Dragon flight** (945 lines) — coordinate-based flying system
- **Arena PvP** (358 lines) — challenge/accept isolated combat
- **Planes** (298 lines) — multi-planar system
- **Holidays** (416 lines) — calendar events
- **Star maps** (226 lines) — celestial display

### Polish
- **Tier 5: Test client + boot extraction — Complete (2026-04-15).** `internal/testclient` harness (Harness/Client, IAC+ANSI strip, prompt detect, login helpers) plus `internal/boot` consolidation of all 19 cross-package callback wires. `cmd/smaug/main.go` reduced from ~500 to ~70 lines; `server.StartOnListener` replaces racy `Start(port int)`. 9 integration tests migrated + 5 new scenario tests (one per interactive surface: `act/olc`, `act` mortal, `combat`, `magic`, `mudprog`). Caught two real bugs (dormant `MoveChar` greet-prog wiring; silent close on `InputQueue`). 15 packages pass `go test -count=1 ./...`. See `phase5-tier5-testclient.md` for the full record and follow-ups queued for Phase 6.
- **Tier 6: Combat depth — Complete (2026-04-17).** Nine task groups (G1-G9) closing the 2026-04-17 audit's critical findings C1 (PC multi-attack cascade), C2 (weapon proficiency bonus), and C3 (stance application in combat). Extracted `MultiHit` out of `ViolenceUpdate` (G1); bridged `combat`↔`act` skill-check helpers via nil-safe hooks + boot-time GSN resolver for 18 combat-relevant skills (G2); ported the 6-tier PC cascade second…seventh_attack with C-faithful integer tier math (G3); gated dual-wield behind learned-roll with `dual_bonus` threading + low-move `-20` penalty (G4); added PLR_NICE / ACT_NOATTACK / AFF_BERSERK / TIMER_ASUPRESSED front-of-round gates (G5); ported non-ENABLE_WEAPONPROF `WeaponProfBonusCheck` and wired it into `OneHit`'s AC-before-roll / dam-on-hit / learn-on-miss paths (G6); integrated stance into the combat loop via synthetic `StanceIndex` defaults (stances.dat is a stub), NPC num_attacks stacking, PC grand-master bonus-attack loop, and dam_done/dam_taken multipliers with mastery-scaled boundaries (G7); replaced C's racy `static dual_flip` with explicit `oneHitFull(wield)` argument-passing so the dual-wield bonus swing uses the offhand weapon (G8); and updated all project docs (G9). All 12 acceptance criteria in `plan-combat-depth.md` verified by RNG-stubbed deterministic tests — no loose statistical checks. `go test -count=3 ./...` green across all 15 packages. See `plan-combat-depth.md` (completion appended at end) for the full record.
- **Tier 7: Damage-message gaps — Complete (2026-04-17).** Three task groups closing the inline TODOs in `combat/dammessage.go` against C `fight.c:4410-4596`: (G1) ported the `was_in_room` swap so a cross-room attacker is temporarily moved into the victim's room for message routing and restored via `defer` (panic-safe); (G2) ported `PCFLAG_GAG` self-suppress — zero-damage misses are silenced for the gagged side only (`gcflag` gates TO_CHAR, `gvflag` gates TO_VICT; TO_NOTVICT always delivers; positive damage never silenced); (G3) added `isWieldingPoisoned(ch, obj)` with the C-exact identity check (`obj == GetEqChar(WEAR_WIELD) || == GetEqChar(WEAR_DUAL_WIELD)` AND `ITEM_POISONED`) and a new branch emitting `"$n's poisoned <attack> ..."` using the generic attack_table entry (not obj.short_descr) per C. 14 new tests landed (2 G1 + 5 G2 + 7 G3) + the pre-existing `TestDamMessageDifferentRoomsEmitsToVictOnly` replaced with a fresh C-fidelity variant. `go test -count=3 ./...` green across all 15 packages. See `plan-dammessage-gaps.md` (completion appended at end) for the full record. Deferred: `DoGag` player command port and `util.Act` per-recipient color preservation — tracked in TODO.md.
- **Tier 9: Communication channels — Complete (2026-04-17).** Four in-scope groups from `plan-channels.md` unblocking cross-player coordination: (G1) shared infrastructure — added `handler.IsSameGroup(a, b *CharData) bool` porting C `act_comm.c:4293-4301` (leaderless-self, nil-safe, flat-graph semantics with a dedicated test that a 3-deep Leader chain does NOT imply same-group, matching C), plus the opportunistic sender-gate for `PLR_SILENCE` across `DoTell` / `DoYell` / `DoShout` / `DoGossip` (latent bug: `PLR_SILENCE` was defined and set by `DoSilence` but no communication command consulted it — now matches C `talk_channel:500` `"You can't <verb>."` for the channel commands and C `do_tell:1804` `"You can't do that."` for tells); (G2) `DoImmtalk` — C `act_comm.c:1327` port with mortal → `"Huh?"`, `PLR_SILENCE` sender gate, deaf-sender block with `"You don't have the immtalk channel turned on..."` (NOT clearing the deaf bit; C's `xREMOVE_BIT` at line 514 is unreachable when the sender returns on line 511), self-echo, and a broadcast walking `WorldRef.Descriptors` filtered on `CON_PLAYING && Trust >= LEVEL_IMMORTAL && !Deaf[IMMTALK]` with `DoClantalk`-style `&Y`/`&G`/`&D` color codes; (G3) `DoGtell` — C `act_comm.c:4217` port with empty-arg `"Tell your group what?"`, `PLR_NO_TELL` sender gate, `IsSameGroup(gch, ch)` filter over `WorldRef.Characters`, plus a new `QuickLoginTwo` helper in `internal/testclient` that drives the first concurrent two-client scenario in the test suite; (G4) `talk_auction` broadcast helper + stub `DoAuction` — `BroadcastAuction` exported for Phase 6 re-use with the full C `talk_auction:4309` filter chain (`CON_PLAYING && Trust >= 5 && !Deaf[AUCTION] && !InRoom.RoomFlags[ROOM_SILENCE]`), and `DoAuction` emits `"The auction house is currently closed. (See the Phase-6 roadmap.)"` so the command visible in `commands` doesn't look broken while the full auction subsystem is deferred. Alias parsing requires a space after `:`/`;` (`": hi"` works, `":hi"` does not — plan-documented deliberate scope cut; full prefix-within-token matching is an interpreter concern outside this plan). Boot registers five entries (`immtalk`, `:`, `gtell`, `;`, `auction`) using the same-name-second-entry pattern already used for `pager` / `pagelen`. 26 new tests across `internal/act/{channels,auction,comm}_test.go`, `internal/handler/group_test.go`, and `internal/testclient/channels_test.go`; each gate mutation-verified (`IsImmortal`, trust >= 51, deaf-not-cleared, `IsSameGroup`, auction trust >= 5, `PLR_SILENCE` on all four pre-existing comm commands). Flat-graph `IsSameGroup` covers the equivalence relation without recursing through chains, matching C exactly. Deferred: full auction state-machine + `music` / `newbiechat` / `racetalk` / `wartalk` / `counciltalk` / `guildtalk` channels (Phase 6). `go test -count=3 ./...` green across all 15 packages. See `plan-channels.md` (completion appended at end) for the full record.
- **Tier 8: Player-config commands — Complete (2026-04-17).** Five in-scope groups from `plan-player-config.md`: (G1) `DoSave` — C `act_comm.c:3277` port with NPC gate, level<2 gate ("You must be at least second level to save."), `ch.Wait=2`, `SaveFunc` invocation; (G2) `DoAfk` — C `act_info.c:6020` port toggling `PLR_AFK` with self-message + `util.Act(..., TO_CANSEE)` room broadcast; (G3) `DoTitle` — C `player.c:3138`/3112 port with `PCFLAG_NOTITLE` gate, 50-byte truncate, `SmashTilde` + `SmashColorToken`, and the `set_title` leading-space-for-alnum-start rule; (G4) `DoPassword` — C `act_info.c:5078` port with a deliberate divergence: Go requires `password <old> <new> <again>` (C takes only `<new> <again>` with the old-pwd check commented out) to align with the shipped bcrypt migration, and min length bumped from C's 5 to 6; (G5) `pagelen` registered in `internal/boot/boot.go` as a second name for the existing `DoPager` handler, verified via a new `Interpret`-level test. Two util prereqs ported to C fidelity: `util.CaseArgument` (C `interp.c:1170`) and `util.SmashColorToken` (C `db.c:4462`, `&`→`+` and `^`→`-`). `act.BcryptCost` added as a package-local seam synced from `game.BcryptCost` inside `boot.Boot` so `TestOpts()` lowers both in one call. 29 new test cases across `internal/act/playercfg_test.go` (28) and `internal/boot/boot_test.go` (1 BcryptCost-sync test) plus 2 new util tests; every gate mutation-verified (flip → red → revert → green). `bio` / `description` deferred on the plan's R6 — `internal/game/editor.go` `/s` path leaves the descriptor in `CON_EDITING` forever and must be fixed before editor-backed commands ship. `go test -count=3 ./...` green across all 15 packages. See `plan-player-config.md` (completion appended at end) for the full record.
- Performance profiling and optimization
- Stress testing (100+ concurrent connections)
