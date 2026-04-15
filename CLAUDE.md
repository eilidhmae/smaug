# SMAUG MUD

## Prime Directives (override all other rules)

1. **You are the manager.**
2. **You always read the entire documents.**
3. **You always start with `@CLAUDE.md` and `@smaug-go/doc/*`.**
4. **You always keep these documents up-to-date as you work.**
5. **You always follow TDD.**
6. **You always use worker subagents.**
7. **Your worker subagents always follow TDD.**
8. **You keep the entire context, and give worker subagents only what they need.**
9. **You verify work by running adversary subagents — as many in parallel as appropriate for the scope.**
10. **You escalate adversary review as needed:**
    - For each adversary that does not return PASS, run another adversary with the same task.
    - If two adversaries disagree, or find different things, run a third adversary with the same task.
    - If you cannot get agreement between adversaries, step in yourself.
    - You may choose the best course of action from the findings, then run another adversary to challenge your chosen solution.
    - If the situation seems unresolvable, escalate to human input with detail about the problem.

---

This is a SMAUG (Simulated Medieval Adventure Multi-User Game) MUD — a purely text-based MMORPG built on the Diku → Merc → SMAUG lineage.

## Repository Layout

```
src/          # Original C codebase (~100 source files, ~30,000+ lines)
db/           # Game data files (areas, classes, races, players, clans, etc.)
  area/       # .are area files (rooms, mobs, objects, resets)
  classes/    # .class files
  races/      # .race files
  player/     # Player save files
  clans/      # Clan data
  deity/      # Deity data
  councils/   # Council data
  boards/     # Board/note data
  gods/       # Immortal building data
  maps/       # Overland map data
doc/          # Original C codebase documentation
smaug-go/     # Go port (in progress)
```

## Active Work: C to Go Port

The C codebase is being ported to pure Go (no Cgo). All work happens in `smaug-go/`.

**Read `smaug-go/doc/` for detailed documentation:**

- `smaug-go/doc/plan.md` — Full architectural plan: C architecture summary, Go design decisions, project structure, C→Go file mapping, data file mapping, testing strategy
- `smaug-go/doc/phases.md` — All 5 implementation phases with detailed deliverables and verification criteria (Phases 1–4 complete)
- `smaug-go/doc/phase1-completed.md` — Phase 1 record (per-package breakdown with files, structs, functions, test results)
- `smaug-go/doc/phase1-remaining.md` — Phase 1 status (complete)
- `smaug-go/doc/phase2-completed.md` — Phase 2 record (completed work so far)
- `smaug-go/doc/phase2-remaining.md` — Phase 2 task breakdown with priority order
- `smaug-go/doc/phase3-plan.md` — Phase 3 implementation plan: 12 task groups with dependencies and execution order
- `smaug-go/doc/phase3-completed.md` — Phase 3 record (all 12 task groups complete)
- `smaug-go/doc/phase4-plan.md` — Phase 4 plan: 12 task groups (5 quality + 7 features) — COMPLETE (G11 hotboot moved to Phase 5)
- `smaug-go/doc/phase4a-completed.md` — Phase 4a record (quality pass: G1–G5 complete)
- `smaug-go/doc/phase4b-completed.md` — Phase 4b record (features: G6–G10, G12 complete; G11 deferred)
- `smaug-go/doc/go-idiom-review.md` — Go idiom audit: 10 findings, all resolved
- `smaug-go/doc/security-review.md` — Security audit: 16 findings, 14 fixed
- `smaug-go/doc/phase5-tier1-foundation.md` — **Phase 5 Tier 1** plan: Act() dispatcher, spell_smaug, missing saves, skill-persistence + Silver/Copper save bugs, skill-learning formula
- `smaug-go/doc/phase5-tier1-completed.md` — Phase 5 Tier 1 record (G1–G6 complete; adversary + quorum verified)
- `smaug-go/doc/phase5-tier2-wiring.md` — **Phase 5 Tier 2** plan: honor already-defined flags, wire loaded-but-idle data
- `smaug-go/doc/phase5-tier2-completed.md` — Phase 5 Tier 2 record (G1–G9 complete; three-agent adversary quorum verified)
- `smaug-go/doc/phase5-tier3-mudprog.md` — **Phase 5 Tier 3** plan: ~86 missing if-checks, missing mob triggers, oprog + rprog subsystems, mpsleep runtime, ~35 missing mp commands
- `smaug-go/doc/phase5-tier3-completed.md` — Phase 5 Tier 3 record (G1–G6 complete; three-agent adversary quorum verified; 7 load-bearing findings fixed before land)
- `smaug-go/doc/phase5-tier4-content.md` — **Phase 5 Tier 4**: missing spells, combat/utility skills, damage-message dispatcher, missing immortal/mortal commands, OLC interactivity

## Current Status

**Phases 1–4 complete. Phase 5 Tiers 1–3 (foundation + flag-honoring + mudprog depth) complete.** 85 source files, 69 test files, 2,205 top-level test cases — all passing across 13 packages. Boot loads 1,909 rooms, 4,299 exits, 505 mob templates, 821 obj templates, 406 mob instances, 710 obj instances, 1,603 helps, 325 skills/spells, 17 classes, 15 races, 8 clans, 2 deities, 496 socials, 2 boards.

### What works
- TCP server with goroutine-per-connection I/O
- Full login flow: returning players load from saved files with password verification
- Character creation: new player flow with name confirm, password, sex/class/race selection
- Player save on quit with automatic directory creation, periodic autosave (every 5 min)
- Player inventory, equipment, and affects saved/loaded from player file
- Single-threaded game loop at 4 pulses/second
- Command interpreter with prefix matching
- Room navigation (10 directions) with auto-look through real loaded rooms
- Commands (135+): look, quit, say, score, who, help, commands, inventory, equipment, get, drop, put, give, wear, remove, sacrifice, kill, flee, tell, reply, yell, gossip, emote, open, close, unlock, lock, consider, where, time, cast, eat, drink, fill, empty, examine, shout, pmote, pager, weather, murder, wimpy, buy, sell, list, value, backstab, bash, kick, disarm, rescue, sneak, hide, steal, pick, scan, aid, recall, clans, claninfo, clantalk, join, leave, clandeposit, clanwithdraw, deities, devote, note, speak, learn, repair, appraise, mstat, ostat, rstat, goto, transfer, at, bamfin, bamfout, force, peace, purge, restore, advance, slay, mfind, ofind, mwhere, owhere, users, invis, holylight, freeze, silence, echo, recho, snoop, redit, ocreate, mcreate, rdig, rlist, olist, mlist, savearea, rest, sit, stand, sleep, wake, bank, ban, track, quest, practice, skills, spells, quaff, recite, brandish, zap, follow, group, order, assist, mount, dismount, mset, oset, rset, aset, astat
- Area file loading from `db/area/*.are` with skip-and-recover on parse errors
- Exit resolution (vnum → room pointer linking after all areas load)
- Area reset processing: mob/object instantiation (M/O/P/G/E/D/H reset commands)
- ANSI color processing wired into output flush (`ColorFunc` on descriptor)
- Spell-name object values (potions/scrolls/wands) via peek-based detection
- Class file loading (17 classes from `db/classes/`)
- Race file loading (15 races from `db/races/`)
- Skills/spells loading (325 from `db/system/en/skills.dat`)
- Player save/load (`persist.LoadPlayer` / `persist.SavePlayer`, round-trip tested)
- Handler package: CreateMobile, CreateObject, CharToRoom/FromRoom, ObjToRoom/Char/Obj, EquipChar, ObjFromChar/Room/Obj, UnequipChar, ExtractObj, ExtractChar
- Affect management: AffectToChar, AffectRemove, AffectStrip, AffectJoin, AffectModify (13 apply types + bitvector flags)
- Find functions: GetCharRoom, GetCharWorld, GetObjCarry, GetObjWear, GetObjHere, GetObjWorld (two-phase exact+prefix search)
- Attribute bonus tables: StrApp, IntApp, WisApp, DexApp, ConApp, ChaApp, LckApp (all 7 tables, 26 entries each)
- BitVector: added Not(), AndNot() methods
- Object commands: get (room + container), drop, put, give, wear (auto-detect location), remove, sacrifice
- Game updates: HP/mana/move regen per tick, affect duration countdown with wear-off, corpse decay, NPC wandering, violence update, aggressive mob attacks (aggrUpdate), NPC scavenging, area reset on timer
- Combat system: StartFighting, StopFighting, ViolenceUpdate, OneHit (thac0 + d20 vs AC), Damage, MakeCorpse, kill/flee/murder commands, XP gain on kill, gold in corpses, position checks, dual wield extra attack, wimpy auto-flee
- Communication: tell/reply (private), yell (area), shout (global), gossip (global), emote (room), pmote (possessive emote)
- Door commands: open, close, lock, unlock with key matching
- Movement position check: must be standing to move
- Enhanced info: consider (level comparison), where (find in area), time (game clock), enhanced score (alignment text, affects, kill stats), weather, examine (look + contents)
- Consume commands: eat (food/pills), drink (containers/fountains), fill (from fountain), empty (containers), condition tracking (hunger/thirst/drunk)
- Shops: buy, sell, list, value with FindKeeper, configurable profit margins, item type restrictions
- Magic system: 20 spells (3 heals, 2 damage, 3 buffs, 4 debuffs, 4 detects, sleep, charm, shield, identify), cast command with mana/target/save, spell function registry
- Pager: page long output with (C)ontinue/(N)on-stop/(R)efresh/(B)ack/(Q)uit, configurable page length, pager command to toggle on/off
- String editor: line-based text editor for OLC descriptions (/l /c /d /g /i /r /f /a /s commands), CON_EDITING integration in game loop
- Immortal commands (26): goto, transfer, at, bamfin/bamfout, force, peace, purge, restore, advance, slay, mfind, ofind, mwhere, owhere, users, invis, holylight, freeze, silence, echo, recho, snoop, mstat, ostat, rstat
- Combat skills (12): backstab, bash, kick, disarm, rescue, sneak, hide, steal, pick, scan, aid, recall — with skill improvement on use (learnFromSuccess/learnFromFailure)
- Magic system: 32 spells (3 heals + heal, 2 damage, 3 buffs, 4 debuffs, 4 detects, sleep, charm, shield, identify, locate object, create food/water, summon, teleport, enchant weapon/armor, invis, fly), cast command with mana/target/save, spell function registry
- Subsystem loaders: clans (8 loaded), deities (2 loaded), socials (496 loaded), boards (2 loaded) from db/ data files
- Social dispatch: 496 social commands as fallback in command interpreter with $n/$N/$e/$m/$s variable substitution
- Clan commands: clans, claninfo, clantalk, join, leave — with member tracking
- Deity commands: deities, devote (with worshipper tracking)
- Board/note commands: note list/read/write/post/remove
- MUD Programs: script interpreter with if/or/else/endif, 106 if-checks (Tier 3 added 77 across object/mob lookup, char state, room/exit state, social, misc), variable substitution ($n/$t/$o etc.), 13 mob + 19 obj + 13 room trigger types, full object-prog and room-prog subsystems via supermob pattern, mpsleep delayed-execution runtime, 44 mp commands (mpecho variants, mpmload/mpoload, mpat, mpadvance, mpslay, mprestore, mpbodybag, mpopen/close/fillpassage, mppeace, mpdeposit/mpwithdraw, mpapplyaffect, mpscatter, mpstrew, and more)
- OLC: redit (name/desc/sector/flags/exdesc/exit), ocreate, mcreate, rdig, rlist, olist, mlist, savearea, mset, oset, rset, aset, astat
- Area save: persist/area_write.go writes rooms, mobs, objects, resets, shops to .are files
- Protocol support: telnet negotiation helpers, MCCP2 zlib compression, MSDP variable reporting, MSSP server status
- Position commands: rest, sit, stand, sleep, wake with full state machine (fighting/mounted/sleep-affected guards)
- Banking system: deposit, withdraw, balance with ACT_BANKER NPC requirement
- Ban system: ban site (exact/prefix/suffix wildcards), ban list, ban remove, CheckBans on login, persist load/save
- Mob tracking/hunting: BFS pathfinding (map-based visited, no room flag pollution), do_track player command, HuntVictim NPC behavior in mobileUpdate
- Quest system: quest request/complete/list/buy/info/time/points, random mob-slay quests, quest point rewards, QuestUpdate in game loop
- Skill/spell info: practice (at trainer, spend sessions), skills (list non-spell skills), spells (list spells)
- Item-based spellcasting: quaff (potions), recite (scrolls), brandish (staves, area effect), zap (wands, targeted)
- Group system: follow, group, order (sends to follower input queue), assist (join ally's combat)
- Mount system: mount (ACT_MOUNTABLE NPC), dismount, POS_MOUNTED position state
- Test suite: 64 files, 1,445 top-level cases covering all 13 packages (types, util, net, persist, command, world, handler, game, act, combat, magic, mudprog)
- Phase 5 Tier 1 foundations: `util.Act()` message dispatcher with visibility gating (wiz-invis/holylight/invis/hide); `magic.SpellSmaug` data-driven spell dispatcher with non-NEWSPELLS bit layout, Target-first routing, SE_REFLECT/ABSORB, `parseDiceExpr` covering all real `l*N`/`(l*N)+M`/`NdM` forms; `SavesWands`/`SavesParaPetri`/`SavesBreath` (SavesWands honours RIS_MAGIC immunity); player-file learned proficiencies + Silver/Copper now persist; skill-learning formula uses `learned + 5*difficulty` with per-class adept cap and `learned <= 0` guard
- Integration tests: 9 end-to-end tests via programmatic TCP connections (server boot, char creation, commands, communication, multi-connection)
- Security hardening: bcrypt passwords, connection limits, input/output bounds, path traversal defense, brute force protection, atomic saves, trust caps

### What's next (Phase 5 — Depth-First Parity Closure)

Phase 4 is complete. Tier 1 landed 2026-04-13. Tiers 2 and 3 landed 2026-04-14. Remaining Tier 4 closes remaining parity gaps from the 2026-04 audit before optional breadth systems (overland, housing, polymorph, archery, stances, dragon flight, arena, planes, hotboot) in Phase 6.

- **Tier 1 — Foundation + correctness.** ✓ Complete. `Act()` dispatcher, `spell_smaug` data-driven dispatcher, three missing saves, skill/Silver/Copper persistence bugs fixed, skill-learning formula corrected. See `smaug-go/doc/phase5-tier1-completed.md`.
- **Tier 2 — Flag honoring + wire idle data.** ✓ Complete. Enforced ROOM_NO_MAGIC (+ AFLAG_NOMAGIC) / NO_SUMMON / NO_RECALL / NOFLOOR / DEATH / SILENCE / NO_MOB / PRIVATE / SOLITARY / NODROP, ITEM_ANTI_* wear restrictions, EX_SECRET/EX_HIDDEN exit hiding, EX_NOMOB, PLR_AFK/NO_TELL/NO_EMOTE; added language scrambler + speak/learn commands, repair/appraise commands (damage-aware cost), class/race bans with nanny enforcement, note recipient filtering + login-greeting mail count, clan storeroom deposit/withdraw (leader-gated). Three-agent adversary quorum reviewed all changes; all load-bearing findings addressed. See `smaug-go/doc/phase5-tier2-completed.md`.
- **Tier 3 — Mudprog depth.** ✓ Complete. 77 new if-checks (29→106), LOGIN/VOID/TELL/HOUR/TIME/SELL mob triggers + HITPRCNT wired, full object-prog subsystem (19 fire helpers via supermob pattern), full room-prog subsystem (13 helpers), mpsleep delayed-execution runtime, +35 mp commands (44 total). Two rounds of three-agent adversary quorum raised 10 load-bearing findings total — all fixed with TDD (prog-bit widening, command-hook precedence, indoors flag, mpScatter/mpDeposit/mpBodybag semantics, level→GetTrust, class/race string compare, cansee blind+dark checks). Two adversary claims (mpAdvance AVATAR cap, mpSlay raw_kill) turned out to match C correctly on closer inspection. See `smaug-go/doc/phase5-tier3-completed.md`.
- **Tier 4 — Content breadth.** Registry has 30 of ~101 C spells; 15 of ~43 combat/utility skills; no damage-message dispatcher (`new_dam_message` equivalent); 14 missing immortal commands; 9 missing mortal commands; OLC has 6 redit subcommands vs C's ~20 and no interactive oedit/medit/mpedit/opedit/rpedit. See `smaug-go/doc/phase5-tier4-content.md`.

Optional breadth systems (overland, housing, polymorph, archery, stances, dragon flight, arena, planes, holidays, star maps, hotboot) are now Phase 6 candidates.

## Building and Running the Go Port

```bash
cd smaug-go
go build -o smaug-go ./cmd/smaug/
./smaug-go -port 4000 -data ../db
```

Connect with: `telnet localhost 4000`

Build all packages: `go build ./...`

The Go module is `github.com/eilidhmae/smaug`.

## Go Port Architecture

### Key Design Principle: Single-Threaded Game State

The game loop runs in one goroutine. All game state mutation happens there. Network I/O runs in separate goroutines (one per connection) and communicates via channels. No locks on game data.

```
[read goroutine per connection] → inputQueue channel → [game loop] → outputBuf → [flush to socket]
```

### Data Model Translation

- C linked lists → Go slices and maps
- C `EXT_BV` (128-bit bitvector) → `types.BitVector [4]uint32`
- C function pointers → Go function values + registry maps
- C globals → fields on `world.World` struct (passed explicitly)
- C `STRALLOC/STRFREE` → plain Go strings (GC)
- All SMAUG file formats read/written as-is (no JSON/protobuf conversion)

### Go Project Structure

```
smaug-go/
  go.mod
  cmd/smaug/main.go                  # Entry point, command registration, bootDB
  internal/
    types/                            # Core data structures, enums, constants (17 files)
    util/                             # String, dice, logging, Act() message dispatcher (4 files)
    world/                            # Mutable game state container (1 file)
    persist/                          # File format I/O — scanner, area, classes, races, player, skills (6 files)
    net/                              # TCP server + color processing (2 files)
    game/                             # Game loop, nanny, pager, string editor, updates (6 files)
    command/                          # Command registry + interpreter with social fallback (1 file)
    act/                              # Player commands — info, combat, movement, objects, shops, wiz, clans, OLC, skills, socials (16 files)
    handler/                          # Entity manipulation: create mob/obj, room placement, area resets, find (3 files)
    combat/                           # Combat system: fighting, damage, corpses, dual wield (1 file)
    magic/                            # Spell/skill system: 32 hand-coded spells + spell_smaug data-driven dispatcher + cast command (2 files)
    mudprog/                          # MUD program interpreter: driver, if-checks, triggers, commands, variable substitution (4 files)
    overland/                         # (future) Overland maps
  doc/                                # Port documentation
```

## C Source Reference

Key C files to reference when porting (in `src/`):

| File | What it contains |
|------|-----------------|
| `mud.h` | All struct definitions, enums, `#define` constants (~6,700 lines) |
| `smaug.c` | `game_loop()`, network I/O, descriptor management |
| `db.c` | `boot_db()`, area file loading |
| `handler.c` | `char_to_room`, `obj_to_char`, `equip_char`, find functions |
| `interp.c` | `interpret()`, command hash table |
| `save.c` | Player file save/load format |
| `update.c` | `update_handler()`, game tick updates, regeneration |
| `fight.c` | `one_hit()`, `damage()`, combat resolution |
| `magic.c` | Spell casting and spell functions |
| `act_*.c` | Player commands by category |
| `mud_prog.c` | MUD program interpreter |
| `build.c` | OLC building system |
| `const.c` | Attribute bonus tables |
| `color.c` | Color code → ANSI mapping |

## Conventions

- All enums from C are untyped `int` constants (not typed Go enums) for easy interop with struct fields
- Struct fields use PascalCase but constant names match C names for cross-reference (`ACT_IS_NPC`, `PULSE_VIOLENCE`, `ROOM_VNUM_TEMPLE`)
- The `types/` package has no logic dependencies — everything else can import it
- `world.World` is passed explicitly, never as a global
- `act.WorldRef` is a package-level variable set at boot (pragmatic exception for commands)
- Error handling in file loaders: log with `util.Bug()` and continue (don't abort on bad data, matching C behavior)
- Avoid Cgo entirely — use pure Go or stdlib equivalents (e.g., `compress/zlib` for MCCP)

## Testing — Test Driven Development

**All new code must follow TDD.** For existing code that lacks tests, backfill tests before modifying.

### TDD workflow
1. Write a failing test that specifies the desired behaviour
2. Run the test, confirm it fails (red)
3. Write the minimum code to make the test pass (green)
4. Refactor if needed, confirm tests still pass
5. Commit

### Strive for code coverage
- Aim for high code coverage on all new and modified code
- Use `go test -cover ./...` to check coverage across packages
- Use `go test -coverprofile=coverage.out ./internal/package/` and `go tool cover -func=coverage.out` to identify untested functions
- Prioritise covering critical paths: combat, spell resolution, file I/O, command dispatch

### Mutation testing for existing code
When adding tests for code that already exists, use mutation testing to verify the tests are meaningful:
1. Stage your working code: `git add <filename>` to save changes you want to keep
2. Mutate the implementation (change a return value, flip a condition, short-circuit a function)
3. Run the tests — confirm they **fail** (proving the test catches the mutation)
4. Restore the original: `git checkout <filename>` to revert the mutation
5. Run the tests again — confirm they **pass**

### Test file placement
- Tests live next to the code they test: `foo.go` → `foo_test.go`
- Test data files go in `testdata/` directories within the package
- Use table-driven tests where there are multiple cases for the same function
- Use `t.Helper()` in shared test helpers

### What to test
- **Pure functions first** — `util/`, `types/bitvector.go`, `net/color.go`, `persist/scanner.go` are ideal: no I/O, no global state
- **File parsers** — use real `.are` snippets as `testdata/` fixtures, verify parsed structs match expected values
- **Command interpreter** — test Find() prefix matching, position checks
- **World operations** — test AddChar/RemoveChar, GetRoom, etc.
- **Commands** — use a fake descriptor with a buffer to capture output, verify command output

### Use Sub-Agents whenever possible
1. You read the full documents so that you have the full context.
2. Give the Sub-Agents explicit directions. Do not give them full context.

### Running tests
```bash
cd smaug-go
go test ./...                    # all tests
go test -cover ./...             # all tests with coverage summary
go test ./internal/types/        # one package
go test -v ./internal/util/      # verbose
go test -run TestBitVector ./internal/types/  # one test
go test -coverprofile=coverage.out ./internal/combat/  # coverage profile
go tool cover -func=coverage.out                       # show per-function coverage
```
