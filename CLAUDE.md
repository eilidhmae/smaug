# SMAUG MUD

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
- `smaug-go/doc/phases.md` — All 4 implementation phases with detailed deliverables and verification criteria
- `smaug-go/doc/phase1-completed.md` — Phase 1 record (per-package breakdown with files, structs, functions, test results)
- `smaug-go/doc/phase1-remaining.md` — Phase 1 status (complete)
- `smaug-go/doc/phase2-completed.md` — Phase 2 record (completed work so far)
- `smaug-go/doc/phase2-remaining.md` — Phase 2 task breakdown with priority order
- `smaug-go/doc/phase3-plan.md` — Phase 3 implementation plan: 12 task groups with dependencies and execution order

## Current Status

**Phase 1 complete. Phase 2 complete.** 55 source files, 27 test files, 455 test cases — all passing. Boot loads 1,909 rooms, 4,299 exits, 505 mob templates, 821 obj templates, 406 mob instances, 710 obj instances, 1,603 helps, 325 skills/spells, 17 classes, 15 races.

### What works
- TCP server with goroutine-per-connection I/O
- Full login flow: returning players load from saved files with password verification
- Character creation: new player flow with name confirm, password, sex/class/race selection
- Player save on quit with automatic directory creation, periodic autosave (every 5 min)
- Player inventory, equipment, and affects saved/loaded from player file
- Single-threaded game loop at 4 pulses/second
- Command interpreter with prefix matching
- Room navigation (10 directions) with auto-look through real loaded rooms
- Commands: look, quit, say, score, who, help, commands, inventory, equipment, get, drop, put, give, wear, remove, sacrifice, kill, flee, tell, reply, yell, gossip, emote, open, close, unlock, lock, consider, where, time, cast
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
- Game updates: HP/mana/move regen per tick, affect duration countdown with wear-off, corpse decay, NPC wandering, violence update
- Combat system: StartFighting, StopFighting, ViolenceUpdate, OneHit (thac0 + d20 vs AC), Damage, MakeCorpse, kill/flee commands, XP gain on kill, gold in corpses, position checks
- Communication: tell/reply (private), yell (area), gossip (global), emote (room)
- Door commands: open, close, lock, unlock with key matching
- Movement position check: must be standing to move
- Enhanced info: consider (level comparison), where (find in area), time (game clock)
- Magic system: 12 spells (3 heals, 2 damage, 3 buffs, 4 debuffs), cast command with mana/target/save, spell function registry
- Test suite: 27 files, 450 cases covering types, util, net, persist, command, world, handler, game, act, combat, magic (mutation-verified)

### What's next (Phase 3: Advanced Systems)
See `smaug-go/doc/phase3-plan.md` for the implementation plan: 12 task groups covering pager, Phase 2 deferred items, subsystem loaders, stat/immortal commands, shops, spells/skills, OLC, clans/boards, MUD Progs, protocol support, and area save.

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
  cmd/smaug/main.go                  # Entry point
  internal/
    types/                            # Core data structures, enums, constants (17 files)
    util/                             # String, dice, logging helpers (3 files)
    world/                            # Mutable game state container (1 file)
    persist/                          # File format I/O — scanner, area, classes, races, player, skills (6 files)
    net/                              # TCP server + color processing (2 files)
    game/                             # Game loop + nanny state machine + login/creation flow (1 file)
    command/                          # Command registry + interpreter (1 file)
    act/                              # Player commands (1 file)
    handler/                          # Entity manipulation: create mob/obj, room placement, area resets (2 files)
    combat/                           # (future) Combat system
    magic/                            # (future) Spell/skill system
    mudprog/                          # (future) MUD program interpreter
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

### Backfilling tests for existing code
When adding tests for code that already exists, verify the tests actually exercise the code by temporarily breaking the implementation (e.g., change a return value, flip a condition, short-circuit a function), confirming the test fails, then reverting the breakage and confirming the test passes. This proves the test is not vacuous.

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

### Running tests
```bash
cd smaug-go
go test ./...                    # all tests
go test ./internal/types/        # one package
go test -v ./internal/util/      # verbose
go test -run TestBitVector ./internal/types/  # one test
```
