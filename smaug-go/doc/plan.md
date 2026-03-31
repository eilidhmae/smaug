# SMAUG MUD: C to Go Port — Full Plan

## Context

SMAUG (Simulated Medieval Adventure Multi-User Game) is a text-based MMORPG built on the Diku → Merc → SMAUG lineage, dating back to 1990. The current codebase at `/home/eilidh/src/smaug/src/` is ~100 C source files (~30,000+ lines) with extensive conditional compilation (`#ifdef`) for optional features.

The goal is a **pure Go port** (no Cgo) that:
- Maintains compatibility with existing area files (`.are` format)
- Maintains compatibility with existing player save files
- Preserves game mechanics faithfully
- Supports the same telnet protocols (MCCP, MSDP, MSSP)
- Produces a single static binary with zero external dependencies

---

## Original C Architecture Summary

### Data Structures (`mud.h`, ~6,700 lines)
- **CHAR_DATA** — Characters (players + NPCs) with stats, combat data, affects, inventory, equipment
- **PC_DATA** — Player-only data (skills[], clan, deity, quest, bank, aliases)
- **OBJ_DATA** / **OBJ_INDEX_DATA** — Object instances and templates
- **ROOM_INDEX_DATA** — Rooms with exits, characters, objects, effects
- **MOB_INDEX_DATA** — NPC templates with combat stats, AI, shops
- **AREA_DATA** — Zones containing rooms/mobs/objects with resets
- **DESCRIPTOR_DATA** — Network connections with I/O buffers, protocol state
- **SKILL_TYPE** — Spell/skill definitions with function pointers
- **EXT_BV** — 128-bit bitvectors (4x uint32) used for flags throughout

### Networking (`smaug.c`)
- BSD `select()` multiplexing on multiple listening ports
- Telnet protocol with MCCP (zlib), MSDP, MSSP, MXP support
- Input pipeline: `recv()` → `inbuf` → `read_from_buffer()` → `incomm` → `interpret()`
- Output pipeline: `write_to_buffer()` → `outbuf` → `flush_buffer()` → `send()`
- Descriptor state machine for login/creation flow

### Game Loop (`smaug.c` `game_loop()`)
- Synchronous, pulse-based: 4 pulses/second (250ms per cycle)
- Each pulse: accept connections → read input → process commands → game updates → flush output
- Update timers: PULSE_VIOLENCE (3 pulses), PULSE_MOBILE (4), PULSE_TICK (70*4=280), PULSE_AREA (60*4=240)

### Game Systems
- **Combat** (`fight.c`) — `one_hit()` per attack, hit roll vs AC, damage types, dual wield, special attacks
- **Magic/Skills** (`magic.c`, `skills.c`) — Function pointer dispatch, skill_table from `skills.dat`
- **Movement** (`act_move.c`) — Sector costs, doors/locks, traps
- **MUD Programs** (`mud_prog.c`) — Scripting on mobs/rooms/objects with if/else/endif
- **Overland** (`overland.c`) — 3 maps of 1000x1000 tiles
- **OLC** (`build.c`) — Online creation for building areas
- Plus: quests, clans, deities, shops, banks, dragon flight, arena, archery, stances, weather, etc.

### Data Persistence
- Custom text file formats throughout: `.are` area files, player files, `skills.dat`, `.class`, `.race`, clan/deity/council files
- The `fread_word`/`fread_string`/`fread_number` functions are the universal parsers

---

## Go Design Decisions

### 1. Data Model Translation

| C Pattern | Go Replacement | Rationale |
|-----------|---------------|-----------|
| Intrusive linked lists (next/prev) | Slices `[]*T` and maps `map[int]*T` | Idiomatic Go; GC handles memory |
| `EXT_BV` (4x uint32) | `type BitVector [4]uint32` | Binary-compatible for file I/O |
| `SPELL_FUN *spell_fun` | `type SpellFunc func(...)` + `map[string]SpellFunc` | Registry pattern replaces function pointer tables |
| `DO_FUN *do_fun` | `type CmdFunc func(ch *CharData, arg string)` + registry | Same pattern |
| `gsn_backstab` globals | Package-level `var GsnBackstab int` | Assigned at boot by scanning skill table |
| `void *vo` in spells | `any` with type assertion | Same semantics, type-safe at assertion point |
| `#ifdef` conditional features | Always included; config flags for optional ones | Simplifies code, avoids build complexity |
| `STRALLOC/STRFREE` | Plain Go `string` | GC eliminates this entire subsystem |
| C global linked lists | Fields on a `World` struct | Explicit dependency, testable |
| Extraction queues | Same pattern | Still needed to avoid mutation-during-iteration |

### 2. Networking: Goroutine-Per-Connection

```
[read goroutine] → inputQueue channel → [game loop goroutine] → outputBuf → [flush to net.Conn]
```

- One read goroutine per connection handles telnet IAC parsing, sends complete lines to a buffered channel
- Game loop goroutine drains channels non-blockingly each pulse, processes commands, writes to output buffer
- Flush at end of each pulse: output buffer → MCCP compress (if negotiated) → `net.Conn.Write()`
- MCCP2 via Go's `compress/zlib` (stdlib, pure Go)
- **Single-threaded game state mutation** — no locks on game data

### 3. Game Loop: Single Goroutine, Pulse-Based

```
time.Ticker(250ms) → each tick:
  1. Accept new connections (drain channel)
  2. Process input from all descriptors (non-blocking channel read)
  3. Game updates (violence, mobile, char, obj, weather, area resets)
  4. Flush output to all descriptors
  5. Clean up dead descriptors
```

### 4. File Format Compatibility

**Read existing SMAUG formats as-is.** No conversion to JSON/protobuf. The `.are` text format is the standard for area builders. Player files must round-trip identically for seamless migration.

A shared `Scanner` type (port of `fread_word`, `fread_string`, `fread_number`, `fread_bitvector`) is used by all file loaders.

### 5. The `act()` Function

Port faithfully — this is one of the most-called functions. Formats messages with token substitution (`$n` = char name, `$N` = victim name, `$e` = he/she/it, `$p` = object short, etc.) and routes to TO_CHAR/TO_VICT/TO_ROOM/TO_NOTVICT targets.

---

## Go Project Structure

```
smaug-go/
  go.mod                              # github.com/eilidhmae/smaug
  cmd/
    smaug/main.go                     # Entry point: boot, listen, game loop
  internal/
    types/                             # Pure data: structs, enums, constants (17 files)
      bitvector.go                     # BitVector [4]uint32, bit ops
      character.go                     # CharData, FightData, HHFData, CharMorph, MorphData
      pcdata.go                        # PCData, KilledData, NuisanceData, AliasData, NoteData
      object.go                        # ObjData, ObjIndexData, ExtraDescrData
      room.go                          # RoomIndexData, ExitData, MapData, PlaneData
      area.go                          # AreaData, ResetData, WeatherData, NeighborData
      mob_index.go                     # MobIndexData
      affect.go                        # AffectData, SmaugAff
      skill.go                         # SkillType
      descriptor.go                    # DescriptorData (goroutine-safe output buffering)
      clan.go                          # ClanData, CouncilData, DeityData
      shop.go                          # ShopData, RepairData
      system.go                        # SystemData, TimeInfoData, ClassType, RaceData
      misc.go                          # VariableData, HelpData, BoardData, BanData, SocialType, etc.
      mudprog.go                       # MProgData, MProgActList, MProgSleepData
      constants.go                     # All #define constants from mud.h
      enums.go                         # All typedef enums as int iota constants
    world/                             # Mutable game state (replaces C globals)
      world.go                         # World struct: indexes, entities, config, methods
    net/                               # Networking
      server.go                        # TCP listener, accept loop, read goroutines
      color.go                         # &-code → ANSI escape processing
      telnet.go                        # (future) Full telnet negotiation
      mccp.go                          # (future) MCCP2 compression
      msdp.go                          # (future) MSDP variables
      mssp.go                          # (future) MSSP status
    game/                              # Game loop and tick dispatch
      loop.go                          # Pulse-based loop, nanny(), enterGame()
      update.go                        # (future) violence, mobile, char, obj updates
    command/                           # Command interpretation
      interpret.go                     # Registry, Find, Interpret with position checks
    act/                               # Player commands
      info.go                          # look, score, who, quit, say, movement, etc.
      comm.go                          # (future) tell, yell, gossip, channels
      obj.go                           # (future) get, drop, wear, remove, etc.
      wiz.go                           # (future) immortal commands
    combat/                            # (future) fight.c port
    magic/                             # (future) spell/skill system
    mudprog/                           # (future) MUD program interpreter
    handler/                           # (future) entity manipulation (char/obj to/from room)
    overland/                          # (future) overland map system
    persist/                           # File I/O
      scanner.go                       # Shared fread_word/fread_string/fread_number
      area.go                          # .are file parser (mobs, objects, rooms, resets, shops)
      player.go                        # (future) Player save/load
      skills.go                        # (future) skills.dat parser
      classes.go                       # (future) .class file loader
      races.go                         # (future) .race file loader
    util/                              # Shared helpers
      strings.go                       # OneArgument, IsName, Capitalize, etc.
      dice.go                          # NumberRange, DiceRoll, NumberPercent, etc.
      log.go                           # Bug, LogString
```

---

## C → Go File Mapping

| C File | Purpose | Go Package/File |
|--------|---------|-----------------|
| `mud.h` (6700 lines) | All structs, enums, constants | `types/` (17 files) |
| `smaug.c` | Main loop, network I/O | `game/loop.go`, `net/server.go` |
| `db.c` | Boot sequence, area loading | `cmd/smaug/main.go`, `persist/area.go` |
| `handler.c` | Entity CRUD | `handler/` (future) |
| `interp.c` | Command dispatch | `command/interpret.go` |
| `save.c` | Player save/load | `persist/player.go` (future) |
| `update.c` | Game tick updates | `game/update.go` (future) |
| `fight.c` | Combat | `combat/` (future) |
| `magic.c` | Spells | `magic/` (future) |
| `skills.c` | Skills | `magic/` (future) |
| `act_comm.c` | Communication | `act/comm.go` (future) |
| `act_info.c` | Information | `act/info.go` |
| `act_move.c` | Movement | `act/info.go` (basic), `act/move.go` (future full) |
| `act_obj.c` | Object commands | `act/obj.go` (future) |
| `act_wiz.c` | Immortal commands | `act/wiz.go` (future) |
| `mud_prog.c` | MUD programs | `mudprog/` (future) |
| `mud_comm.c` | MUD prog commands | `mudprog/` (future) |
| `build.c` | OLC building | `olc/` (future) |
| `tables.c` | Function lookups | `magic/registry.go`, `command/interpret.go` |
| `const.c` | Attribute tables | `types/constants.go` |
| `protocol.c` | Telnet protocols | `net/telnet.go` (future) |
| `color.c` | ANSI colors | `net/color.go` |
| `overland.c` | Overland maps | `overland/` (future) |
| `quest.c` | Quests | `act/` (future) |
| `clans.c` | Clans | `act/` (future), `persist/clans.go` (future) |
| `deity.c` | Deities | `act/` (future), `persist/deity.go` (future) |
| `shops.c` | Shops | `act/` (future) |

## Data File Mapping

| Path | Format | Go Loader |
|------|--------|-----------|
| `db/area/*.are` | SMAUG area format | `persist/area.go` ✅ |
| `db/area/area.lst` | Area file list | `persist/area.go` ✅ |
| `db/classes/*.class` | Class definitions | `persist/classes.go` (future) |
| `db/classes/class.lst` | Class file list | `persist/classes.go` (future) |
| `db/races/*.race` | Race definitions | `persist/races.go` (future) |
| `db/races/race.lst` | Race file list | `persist/races.go` (future) |
| `db/player/<letter>/<Name>` | Player saves | `persist/player.go` (future) |
| `db/clans/` | Clan data | `persist/clans.go` (future) |
| `db/deity/` | Deity data | `persist/deity.go` (future) |
| `db/councils/` | Council data | `persist/councils.go` (future) |
| `db/boards/` | Board/note data | `persist/boards.go` (future) |
| `db/gods/` | Immortal building data | `persist/gods.go` (future) |
| `db/hotboot/` | Hotboot state | (future) |
| `db/maps/` | Overland maps | `overland/` (future) |

---

## Testing Strategy

**TDD is mandatory.** All new code gets tests written first. Existing code backfilled with mutation-verified tests. See `CLAUDE.md` at repo root for workflow details.

### Current State (as of 2026-03-30)

18 test files, 383 test cases — all passing. 6 mutations tested and detected. All `internal/` packages have test coverage (only `cmd/smaug` has no tests — it's just wiring).

| Test File | Package | Cases | Coverage |
|-----------|---------|-------|----------|
| `bitvector_test.go` | `types/` | ~60 | All BitVector methods, boundary bits, serialization |
| `attributes_test.go` | `types/` | 8 | All 7 attribute bonus tables + length validation |
| `strings_test.go` | `util/` | ~50 | All string functions, edge cases |
| `dice_test.go` | `util/` | ~35 | Statistical distribution tests over 10k trials |
| `color_test.go` | `net/` | ~40 | All color codes, ANSI on/off, escapes |
| `scanner_test.go` | `persist/` | ~45 | All Scanner methods, tilde strings, pipe-OR, flags |
| `area_test.go` | `persist/` | 5 | Integration: objects, rooms, C-mobs, V-mobs, spell objects |
| `classes_test.go` | `persist/` | 1 | LoadClasses with Warrior fixture |
| `races_test.go` | `persist/` | 1 | LoadRaces with Human fixture |
| `player_test.go` | `persist/` | 2 | LoadPlayer fixture + SaveLoadRoundTrip |
| `skills_test.go` | `persist/` | 1 | LoadSkills: fireball, backstab, sanctuary with affects |
| `helps_test.go` | `persist/` | 1 | loadHelps: 3 entries, level/keyword/text |
| `interpret_test.go` | `command/` | ~9 | Find (exact/prefix/trust), Interpret dispatch/position |
| `world_test.go` | `world/` | ~6 | CRUD operations on World |
| `handler_test.go` | `handler/` | 49 | CreateMobile, CreateObject, placement, ResetArea, obj removal, extraction, affects, find functions |
| `loop_test.go` | `game/` | 17 | isValidName, createNewCharacter, applyRaceBonuses, NewGameLoop |
| `prompt_test.go` | `game/` | 5 | FormatPrompt: standard, all tokens, default, no PCData, unknown |
| `info_test.go` | `act/` | 17 | DoLook, DoScore, DoSay, DoWho, DoInventory, DoEquipment, DoQuit, moveChar, showExits |

### Test Types

1. **Unit tests** (done) — BitVector ops, file scanner, string utils, dice rolls, color processing, command registry, world CRUD, prompt formatting, name validation
2. **File round-trip tests** (done) — Player SaveLoadRoundTrip
3. **Area load integration tests** (done) — 5 tests with testdata fixtures for objects, rooms, mobs, spell objects
4. **Command output tests** (done) — `net.Pipe()` pattern capturing DoLook, DoScore, DoSay, DoWho, DoInventory, DoEquipment, moveChar, showExits
5. **Entity creation tests** (done) — CreateMobile, CreateObject, CharToRoom/FromRoom, ObjToRoom/Char/Obj, EquipChar, ResetArea
6. **Handler tests** (done) — ObjFromChar/Room/Obj, UnequipChar, ExtractObj/Char, AffectToChar/Remove/Strip/Join/Modify, GetCharRoom/World, GetObjCarry/Wear/Here/World
7. **Attribute table tests** (done) — All 7 bonus tables with spot-checks at key stat values
8. **Combat math** (Phase 2) — Fix RNG seed, verify `one_hit` calculations match C formulas exactly
9. **MUD prog tests** (Phase 3) — Test `.are` files with known triggers, verify behavior
10. **Stress tests** (Phase 4) — 100+ concurrent telnet connections, measure memory and latency

### Mutation Verification

When backfilling tests for existing code, always verify by:
1. Break the implementation (change a return value, flip a condition)
2. Run tests — they must fail
3. Revert the break
4. Run tests — they must pass

This proves the test actually exercises the code path.
