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

**Status:** Complete. All 12 task groups done. See `phase3-completed.md` for full record.

**Deliverables:**

### Protocol Support (`net/`)
- `mccp.go` — MCCP2 compression via `compress/zlib` (stdlib)
- `msdp.go` — MUD Server Data Protocol variable reporting
- `mssp.go` — MUD Server Status Protocol for listing services

### MUD Programs (`mudprog/`)
- `driver.go` — mprog_driver interpreter with if/else/endif, nested ifs
- `triggers.go` — All trigger types: act, greet, all_greet, speech, random, fight, death, hitprcnt, entry, give, bribe, hour, time, wear, remove, sac, look, exa, zap, get, drop, damage, repair, randiw, speechiw, pull, push, sleep, rest, leave, script, use
- Variable substitution ($n, $N, $t, $T, $i, $I, $e, $E, $j, $J, $k, $K, $o, $O, $p, $P, $r, $R)
- If-checks: level, hp, mana, goldamt, class, race, etc.
- Sleeping programs (delayed execution)
- Object progs and room progs (not just mob progs)

### Immortal Commands (`act/wiz.go`)
- goto, transfer, at, bamfin, bamfout
- slay, purge, advance, restore
- mset, oset, rset (set mob/obj/room attributes)
- mstat, ostat, rstat (stat mob/obj/room)
- force, snoop, switch, return
- peace, wizlock, shutdown, reboot
- log, freeze, silence, hell
- invis, holylight
- mfind, ofind, rfind (find by name)
- mwhere, owhere (find in world)

### OLC — Online Creation (`olc/` or `act/`)
- redit — room editor (name, desc, flags, sector, exits, extra descs, progs)
- oedit — object editor (type, flags, values, affects, extra descs, progs)
- medit — mob editor (stats, flags, attacks, defenses, progs)
- Port from `build.c`, `oredit.c`, `omedit.c`, `ooedit.c`
- Area save (write modified areas back to `.are` files)

### Remaining Spells/Skills (~100+)
- All spell functions from `magic.c`
- All skill functions from `skills.c`
- Weapon proficiency checks
- Skill improvement on use

### Subsystems
- Clans: join, leave, promote, demote, clan talk, clan storeroom, all commands
- Councils: similar to clans
- Deities: prayer, favor tracking, deity-specific effects
- Shops: buy, sell, list, value — NPC shop interaction
- Repair shops: repair, estimate
- Quest system: quest request, complete, buy rewards
- Boards and notes: read, write, remove, post, mail
- Pager: page long output with --more-- prompts
- String editor: for building descriptions in OLC
- Languages: speak, learn, language scrambling

**Verification:** Builder can log in, create areas with OLC, add mob progs, test them. Players can use shops, join clans, interact with all game systems.

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
- Minimal telnet test client for repeatable interactive testing
- Performance profiling and optimization
- Stress testing (100+ concurrent connections)
