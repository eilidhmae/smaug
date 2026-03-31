# Phase 1: Completed Work

This document records what was implemented during Phase 1.

## Summary

Phase 1 goal: "Connect via telnet, log in, walk between rooms, see descriptions."

**Status**: Complete. Full login flow with character creation and returning player support. Players save on quit and reload with password verification. Area resets populate rooms with NPCs and objects. Help system loaded. Prompt system with token substitution. Single static binary with zero dependencies.

**Stats**: 36 source files (~8,891 lines), 17 test files (~3,636 lines), 352 test cases — all passing.

**Boot results (26 area files)**: 1,909 rooms, 4,299 exits, 505 mob templates, 821 object templates, 406 mob instances, 710 object instances, 1,603 help entries, 325 skills/spells, 17 classes, 15 races.

---

## Packages Implemented

### `internal/types/` — Core Data Structures (17 files, ~3,873 lines)

All major C structs from `mud.h` (~6,700 lines) have been ported to Go:

| File | Contents |
|------|----------|
| `bitvector.go` | `BitVector [4]uint32` — 128-bit extended bitvector matching C `EXT_BV`. Methods: `IsSet`, `Set`, `Remove`, `Toggle`, `Clear`, `IsEmpty`, `Equal`, `Or`, `And`, `HasAny`, `SetFrom`, `String`, `ParseBitVector`, `BitVectorFromInt`. |
| `constants.go` | All `#define` constants from `mud.h`: game limits, pulse timing, level hierarchy, language/RIS/body part/autosave/spell/trap/exit/wear/container flags, class constants, mob/obj/room vnums, ACT bits, PCFLAG bits, area flags, color codes, trigger flags, etc. (791 lines) |
| `enums.go` | All `typedef enum` types as untyped `int` iota constants: log types, stances, ret types, connection states, substates, sun/sky, to_types, color types, races, sexes, affected_by, attacks, defenses, item types, item extra flags, room flags, directions, sectors, wear locations, apply types, positions, styles, player flags, timer types, channels, conditions, trap types, clan types, variable types, spell damage/action/power/class/save types, damage types, etc. (1,123 lines) |
| `character.go` | `CharData` (all fields from C `char_data`), `FightData`, `HHFData`, `EditorData`, `TimerData`, `CharMorph`, `MorphData`. Helper methods: `IsNPC()`, `IsImmortal()`, `GetTrust()`, `Send()`, `Sendf()`, `GetCurrStr/Int/Wis/Dex/Con/Cha/Lck()`. |
| `pcdata.go` | `PCData` (all fields from C `pc_data`), `KilledData`, `NuisanceData`, `AliasData`, `NoteData`. |
| `object.go` | `ObjData`, `ObjIndexData`, `ExtraDescrData`. |
| `room.go` | `RoomIndexData` (with `GetExit()` helper), `ExitData`, `MapData`, `PlaneData`. |
| `area.go` | `AreaData`, `ResetData`, `WeatherData`, `NeighborData`. |
| `mob_index.go` | `MobIndexData`. |
| `affect.go` | `AffectData`, `SmaugAff`. |
| `skill.go` | `SkillType` with all fields for spell/skill definitions. |
| `descriptor.go` | `DescriptorData` with `net.Conn`, goroutine-safe output buffering (`sync.Mutex`), `InputQueue` channel, `TelnetState`, `ColorFunc` callback for ANSI processing on flush, `NewDescriptor()`, `WriteToBuffer()`, `WriteToBufferf()`, `FlushOutput()` (applies color processing if `ColorFunc` set), `HasOutput()`. |
| `clan.go` | `ClanData`, `CouncilData`, `DeityData`. |
| `shop.go` | `ShopData`, `RepairData`. |
| `system.go` | `SystemData`, `TimeInfoData`, `ClassType`, `RaceData`. |
| `misc.go` | `VariableData`, `HelpData`, `BoardData`, `BanData`, `SocialType`, `CmdType`, `AuctionData`, `LiqType`, attribute bonus types (`StrAppType`, `IntAppType`, etc.). |
| `mudprog.go` | `MProgData`, `MProgActList`, `MProgSleepData`, `MPROG_*` trigger type constants. |

### `internal/util/` — Utility Functions (3 files, 263 lines)

| File | Functions |
|------|-----------|
| `strings.go` | `OneArgument`, `SmashTilde`, `Capitalize`, `IsName` (prefix match), `IsNameExact`, `IsNumber`, `NumberArgument`, `UMIN`, `UMAX`, `URANGE` |
| `dice.go` | `NumberRange`, `NumberPercent`, `NumberDoor`, `NumberBits`, `DiceRoll`, `NumberFuzzy` — seeded from `crypto/rand` via `ChaCha8` |
| `log.go` | `Bug`, `LogString`, `LogStringPlus` |

### `internal/world/` — Game State Container (1 file, ~155 lines)

`World` struct holds all mutable game state replacing C globals:
- Index maps: `Rooms`, `MobIndex`, `ObjIndex` (`map[int]*Type`)
- Entity slices: `Characters`, `Objects`, `Descriptors`
- Data tables: `Areas`, `Skills`, `Classes`, `Races`, `Clans`, `Councils`, `Deities`, `Shops`, `Repairs`, `Boards`, `Socials`, `Bans`, `Helps`
- Config: `SysData`, `TimeInfo`, `Auction`
- Extraction queues: `ExtractChars`, `ExtractObjs`
- Methods: `New()`, `GetRoom()`, `GetMobIndex()`, `GetObjIndex()`, `AddChar()`, `RemoveChar()`, `AddObj()`, `RemoveObj()`, `FixExits()`

### `internal/persist/` — File Format I/O (6 files, ~2,200 lines)

| File | Contents |
|------|----------|
| `scanner.go` | `Scanner` wrapping `bufio.Reader` with file/line tracking. Methods: `ReadWord`, `ReadString` (tilde-terminated), `ReadNumber` (with pipe-OR support), `ReadToEOL`, `ReadBitvector`, `ReadLetter`, `ReadStringNoHash`, `ReadFlag` (number or letter-based flag format), `Errorf`. |
| `area.go` | `LoadAreas(w, areaDir)` — reads `area.lst`, loads each `.are` file. `loadAreaFile` dispatches `#SECTION` headers with skip-and-recover on parse errors. Section loaders: `loadMobiles` (S/C/V formats), `loadObjects` (whole-line parsing, spell-name peek detection), `loadRooms` (whole-line flag parsing, exits, extra descs, map data, progs), `loadResets`, `loadShops`, `loadRepairs`, `loadSpecials`, `loadHelps` (level + keyword + text until `$` terminator). |
| `classes.go` | `LoadClasses(w, classDir)` — reads `class.lst`, loads each `.class` file. |
| `races.go` | `LoadRaces(w, raceDir)` — reads `race.lst`, loads each `.race` file. |
| `player.go` | `LoadPlayer(r, filename)` — reads `#PLAYER` section with 60+ keywords. `SavePlayer(w, ch)` — writes matching format. `PlayerFilePath(dataDir, name)`. |
| `skills.go` | `LoadSkills(w, filename)` — reads `skills.dat` with `#SKILL` blocks. 325 skills loaded from real data. |

### `internal/net/` — Networking (2 files, 226 lines)

| File | Contents |
|------|----------|
| `server.go` | `Server` with TCP listener, `Incoming` channel, accept loop, read goroutine per connection with telnet IAC stripping. |
| `color.go` | `ProcessColors(text, ansiEnabled)` — `&R/&G/&B/&W/&Y/&C/&P/&O` → ANSI escapes. |

### `internal/game/` — Game Loop + Login + Prompt (3 files, ~750 lines)

| File | Contents |
|------|----------|
| `loop.go` | `GameLoop` with pulse counters. `pulse()`: accept → input → updates → flush → cleanup. `nanny()` state machine with 9 states: `CON_GET_NAME` (checks existing player file), `CON_GET_OLD_PASSWORD` (plaintext verify), `CON_CONFIRM_NEW_NAME`, `CON_GET_NEW_PASSWORD`, `CON_CONFIRM_NEW_PASSWORD`, `CON_GET_NEW_SEX`, `CON_GET_NEW_CLASS` (menu from loaded classes), `CON_GET_NEW_RACE` (menu filtered by class restriction, applies racial stat bonuses), `CON_READ_MOTD`. `enterGame()` places loaded/created character in room. `SavePlayer()` with auto-directory creation. `isValidName()`, `createNewCharacter()`, `applyRaceBonuses()`. |
| `prompt.go` | `FormatPrompt(ch)` — expands prompt tokens: `%h/%H` (hp), `%m/%M` (mana), `%v/%V` (move), `%g` (gold), `%a` (alignment), `%x` (xp), `%r` (room name), `%%` (literal). Default prompt for empty/nil. Sent after each command and on entering game. |

### `internal/command/` — Command Interpreter (1 file, 105 lines)

`Registry` with `commands map[string]*Command` and `sorted []*Command`. `Register()`, `Find()` (exact then prefix match with trust check), `Interpret()` (parse via `OneArgument`, position check with SMAUG-style messages, dispatch).

### `internal/act/` — Player Commands (1 file, ~380 lines)

Commands: `DoLook` (room, character, extra desc, object), `DoScore`, `DoWho`, `DoQuit` (save + disconnect), `DoSay`, movement (10 directions), `DoCommands`, `DoHelp` (searches loaded help data), `DoInventory`, `DoEquipment`. `SaveFunc` callback for persistence without circular dependency.

### `internal/handler/` — Entity Manipulation (2 files, ~494 lines)

| File | Contents |
|------|----------|
| `handler.go` | `CreateMobile(w, idx)`, `CreateObject(w, idx, level)`, `CharToRoom`, `CharFromRoom`, `ObjToRoom`, `ObjToChar`, `ObjToObj`, `EquipChar`. |
| `reset.go` | `ResetArea(w, area)` — M/G/E/O/P/D/H reset commands. `ResetAllAreas(w)`. |

### `cmd/smaug/main.go` — Entry Point (1 file, ~175 lines)

`bootDB()`: `LoadAreas` → `FixExits` → `LoadClasses` → `LoadRaces` → `LoadSkills` → `ResetAllAreas`. Registers 21 commands. Wires `SaveFunc` for quit-save.

---

## Test Suite

**17 test files, ~3,636 lines, 352 test cases — all passing.**

| File | Lines | Tests | What's Covered |
|------|-------|-------|----------------|
| `types/bitvector_test.go` | 435 | ~60 | All BitVector methods, boundary bits, serialization |
| `util/strings_test.go` | 264 | ~50 | All string functions, edge cases |
| `util/dice_test.go` | 215 | ~35 | Statistical distribution tests over 10k trials |
| `net/color_test.go` | 201 | ~40 | All color codes, ANSI on/off, escapes |
| `persist/scanner_test.go` | 317 | ~45 | All Scanner methods, tilde strings, pipe-OR, flags |
| `persist/area_test.go` | ~280 | 5 | Integration: objects, rooms, C-mobs, V-mobs, spell objects |
| `persist/classes_test.go` | ~50 | 1 | LoadClasses with Warrior fixture |
| `persist/races_test.go` | ~50 | 1 | LoadRaces with Human fixture |
| `persist/player_test.go` | ~170 | 2 | LoadPlayer + SaveLoadRoundTrip |
| `persist/skills_test.go` | ~100 | 1 | LoadSkills: fireball, backstab, sanctuary with affects |
| `persist/helps_test.go` | ~60 | 1 | loadHelps: 3 entries (MOTD, RULES, IMOTD), level/keyword/text |
| `command/interpret_test.go` | 202 | ~9 | Find, Interpret with net.Pipe() output capture |
| `world/world_test.go` | 136 | ~6 | CRUD operations on World |
| `handler/handler_test.go` | 444 | 18 | CreateMobile, CreateObject, placement, ResetArea, door state |
| `game/loop_test.go` | ~130 | 17 | isValidName (13 cases), createNewCharacter, applyRaceBonuses, NewGameLoop |
| `game/prompt_test.go` | ~100 | 5 | FormatPrompt: standard, all tokens, default, no PCData, unknown |
| `act/info_test.go` | ~330 | 17 | DoLook (4), DoScore, DoSay (2), DoWho, DoInventory (2), DoEquipment (2), DoQuit, moveChar (2), showExits (2) |

### Mutation Verification Results

6 mutations tested, all detected by the test suite:

| # | File | Mutation | Test That Caught It |
|---|------|----------|-------------------|
| 1 | `bitvector.go` | `IsSet` → always `false` | `TestBitVector_Toggle/bit_127` |
| 2 | `strings.go` | `Capitalize` → identity function | `TestCapitalize/mixed_case` |
| 3 | `dice.go` | `NumberRange` → always returns `from` | Statistical variance check |
| 4 | `color.go` | `&&` outputs `&&` instead of `&` | `TestProcessColors_EdgeCases` |
| 5 | `scanner.go` | Negative sign ignored | `TestReadNumber_Negative` |
| 6 | `world.go` | `GetRoom` → always `nil` | `TestGetRoom` |
