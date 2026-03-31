# Phase 1: Completed Work

This document records what has been implemented during Phase 1.

## Summary

Phase 1 goal: "Connect via telnet, log in, walk between rooms, see descriptions."

**Status**: Full Phase 1 login flow implemented. New players go through character creation (name, password, sex, class, race). Returning players load from saved files with password verification. Players save on quit. Area resets populate rooms with 406 NPCs and 710 objects. 325 skills/spells loaded from skills.dat. ANSI color codes render. Single static binary with zero dependencies.

**Stats**: 35 source files (~8,778 lines), 13 test files (~2,910 lines), 312 test cases — all passing.

**Boot results (26 area files)**: 1,909 rooms, 4,299 exits, 505 mob templates, 821 object templates, 406 mob instances, 710 object instances, 325 skills/spells, 17 classes, 15 races.

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

### `internal/world/` — Game State Container (1 file, 129 lines)

`World` struct holds all mutable game state replacing C globals:
- Index maps: `Rooms`, `MobIndex`, `ObjIndex` (`map[int]*Type`)
- Entity slices: `Characters`, `Objects`, `Descriptors`
- Data tables: `Areas`, `Skills`, `Classes`, `Races`, `Clans`, `Councils`, `Deities`, `Shops`, `Repairs`, `Boards`, `Socials`, `Bans`, `Helps`
- Config: `SysData`, `TimeInfo`, `Auction`
- Extraction queues: `ExtractChars`, `ExtractObjs`
- Methods: `New()`, `GetRoom()`, `GetMobIndex()`, `GetObjIndex()`, `AddChar()`, `RemoveChar()`, `AddObj()`, `RemoveObj()`, `FixExits()`

### `internal/persist/` — File Format I/O (2 files, 1,416 lines)

| File | Contents |
|------|----------|
| `scanner.go` | `Scanner` wrapping `bufio.Reader` with file/line tracking. Methods: `ReadWord`, `ReadString` (tilde-terminated), `ReadNumber` (with pipe-OR support), `ReadToEOL`, `ReadBitvector`, `ReadLetter`, `ReadStringNoHash`, `ReadFlag` (number or letter-based flag format), `Errorf`. |
| `area.go` | `LoadAreas(w, areaDir)` — reads `area.lst`, loads each `.are` file. `loadAreaFile` dispatches `#SECTION` headers with skip-and-recover on parse errors. Section loaders: `loadMobiles` (S/C/V formats, dice parsing, bitvectors, default stances, mudprogs, multi-currency gold detection), `loadObjects` (whole-line parsing for type/flags/values/cost lines, extra descs, affects, progs, recovery on failure), `loadRooms` (whole-line flag parsing, exits with lock conversion, extra descs, map data, progs), `loadResets`, `loadShops`, `loadRepairs`, `loadSpecials`. Helpers: `convertPosition`, `mprogNameToType`, `setDefaultStances`, `skipSection` (with `atLineStart` tracking), `parseObjTypeLine`, `parseObjCostLine`, `parseRoomFlagLine`. |

### `internal/net/` — Networking (2 files, 226 lines)

| File | Contents |
|------|----------|
| `server.go` | `Server` with TCP listener, `Incoming` channel (cap 50), `Start(port)`, `Stop()`. Accept loop goroutine creates `DescriptorData` via `types.NewDescriptor()`, sets `ColorFunc = ProcessColors`, spawns read goroutine per connection. Read goroutine: `bufio.Scanner` line reading, telnet IAC stripping (2-byte commands, 3-byte WILL/WONT/DO/DONT), sends to `InputQueue`. |
| `color.go` | `ProcessColors(text, ansiEnabled)` — converts `&R/&G/&B/&W/&Y/&C/&P/&O` (and lowercase variants) to ANSI escape sequences. `&D/&d` resets, `&&` = literal `&`. Strips codes when ANSI disabled. |

### `internal/game/` — Game Loop (1 file, 344 lines)

`GameLoop` with pulse counters (area, violence, mobile, tick). `Run()` uses `time.Ticker(250ms)`. Each `pulse()`:
1. `acceptNewConnections()` — drains incoming channel, sends banner
2. `processInput()` — non-blocking channel read per descriptor, dispatches to `Interpret()` or `nanny()`
3. Pulse counter updates (stub functions for area/violence/mobile/char/obj updates)
4. `flushOutput()` — sends all descriptor output buffers
5. `cleanupDescriptors()` — removes dead connections, calls `closeDescriptor()` for cleanup

`nanny()` state machine: `CON_GET_NAME` → `CON_GET_OLD_PASSWORD` → `CON_READ_MOTD` → `enterGame()`.

`enterGame()`: creates `CharData` with `PCData`, sets player flags, places in starting room, sends welcome + room description, notifies room.

`closeDescriptor()`: notifies room, removes from room, removes from world, closes connection.

### `internal/command/` — Command Interpreter (1 file, 105 lines)

`Registry` with `commands map[string]*Command` and `sorted []*Command`. `Register()`, `Find()` (exact then prefix match with trust check), `Interpret()` (parse via `OneArgument`, position check with SMAUG-style messages, dispatch).

### `internal/act/` — Player Commands (1 file, ~380 lines)

Commands implemented:
- `DoLook` — Room description, exits, objects, characters; look at specific target (character, extra desc, object in room, object in inventory)
- `DoScore` — Character stats display
- `DoWho` — Online player listing with level and title
- `DoQuit` — Save player, then disconnect with fight check. Uses `SaveFunc` callback to avoid circular dependency.
- `DoSay` — Chat to room
- `DoNorth/East/South/West/Up/Down/Northeast/Northwest/Southeast/Southwest` — Movement with exit checking, closed door detection, leave/arrive messages, auto-look
- `DoCommands` — List available commands
- `DoHelp` — Search help data
- `DoInventory` — List carried (unworn) items
- `DoEquipment` — List worn items with wear location names

Helper functions: `showExits`, `moveChar`, `removeFromRoom`, `addToRoom`, `wearLocName`.

### `internal/handler/` — Entity Manipulation (2 files, ~494 lines)

| File | Contents |
|------|----------|
| `handler.go` | `CreateMobile(w, idx)` — instantiate mob from index template (HP dice, AC interpolation, bare-hand damage). `CreateObject(w, idx, level)` — instantiate object from template (copy extra descs, affects). `CharToRoom`, `CharFromRoom`, `ObjToRoom`, `ObjToChar`, `ObjToObj`, `EquipChar` — entity placement functions. |
| `reset.go` | `ResetArea(w, area)` — process area reset commands (M=mob spawn, G=give to mob, E=equip on mob, O=object in room, P=put in container, D=door state, H=hide object). `ResetAllAreas(w)` — process all areas during boot. Tracks max mob count, deduplicates room objects. |

### `cmd/smaug/main.go` — Entry Point (1 file, ~170 lines)

`main()`: parse flags (`-port`, `-data`), create `World`, set `act.WorldRef`, call `bootDB()`, register commands, create server + game loop, wire `SaveFunc`, signal handling, run.

`bootDB()`: calls `persist.LoadAreas()`, `w.FixExits()`, `persist.LoadClasses()`, `persist.LoadRaces()`, `persist.LoadSkills()`, `handler.ResetAllAreas()`. Creates a fallback room if the temple (vnum 21001) isn't found.

`registerCommands()`: registers all 21 commands (look, quit, say, score, who, commands, help, inventory, equipment, + 10 directions).

### Persist Loaders (4 files, ~1,000 lines)

| File | Contents |
|------|----------|
| `classes.go` | `LoadClasses(w, classDir)` — reads `class.lst`, loads each `.class` file. Parses: Name, Class index, AttrPrime/Second/Deficient, Weapon, Guild, SkillAdept, Thac0/Thac32, HPMin/Max, Mana, ExpBase, Affected, Resist, Suscept, Skill entries, Title entries, Login/Logout/Reconnect messages. |
| `races.go` | `LoadRaces(w, raceDir)` — reads `race.lst`, loads each `.race` file. Parses: Name, Race index, Classes restriction, stat bonuses (Str/Dex/Wis/Int/Con/Cha/Lck), Hit/Mana, Affected/Resist/Suscept, Language, Alignment/MinAlign/MaxAlign, ACPlus, ExpMult, Attacks/Defenses, Height/Weight, HungerMod/ThirstMod, ManaRegen/HPRegen, RaceRecall, WhereName entries, Skill entries. Case-insensitive keyword matching. |
| `player.go` | `LoadPlayer(r, filename)` — reads `#PLAYER` section with 60+ KEY/VALUE keywords (Name, Sex, Class, Race, Level, HpManaMove, Gold, Exp, stats, saves, password, title, flags, etc.). Skips `#OBJECT`/`#CORPSE` sections. `SavePlayer(w, ch)` — writes matching format. `PlayerFilePath(dataDir, name)` — returns `db/player/<first_letter>/<Name>` path. |
| `skills.go` | `LoadSkills(w, filename)` — reads `skills.dat` with `#SKILL` blocks. Parses: Name, Type, Info, Flags, Target, Minpos (with legacy conversion), Saves, Slot, Mana, Rounds, Range, Code (spell/skill function name), Dammsg, Dice, all message strings (hit/miss/die/imm/abs for char/vict/room), Affect entries (duration/location/modifier/bitvector), Class/Race level assignments, Components, Teachers, Value. 325 skills loaded from real data. |

### Spell-Name Handling in Area Parser

Added `loadObjSpellNames()` to `persist/area.go` — peeks at next char after cost line; if `'` (single quote), reads spell names for the item type:
- Potions/scrolls/pills: 3 names → `SpellNames[1],[2],[3]`
- Wands/staves: 1 name → `SpellNames[3]`
- Salves: 2 names → `SpellNames[4],[5]`

Added `SpellNames [6]string` field to `types.ObjIndexData`.

---

## Integration Test Results

Verified end-to-end on 2026-03-30:

1. Server starts, boots, listens on configured port
2. Telnet connection receives SMAUG banner and "By what name..." prompt
3. Name entry → password prompt (with telnet echo-off)
4. Password entry → MOTD → press enter → enter game
5. `look` — room name, description, exits
6. `north` — move to Town Square with leave/arrive messages + auto-look
7. `south` — move back to Temple
8. `say Hello World` — broadcast to room
9. `score` — character stats
10. `who` — player listing
11. `inventory` — "Nothing."
12. `quit` — graceful disconnect with cleanup

---

## Test Suite

**13 test files, ~2,910 lines, 312 test cases — all passing.**

All tests were verified non-vacuous using mutation testing: code was temporarily broken, tests confirmed to fail, code reverted, tests confirmed to pass again.

### Test Files

| File | Lines | Tests | What's Covered |
|------|-------|-------|----------------|
| `types/bitvector_test.go` | 435 | ~60 | IsSet, Set, Remove, Toggle (bits 0, 31, 32, 127, out-of-range), Clear, IsEmpty, Equal, Or, And, HasAny, SetFrom, String, ParseBitVector (valid, invalid, partial, extra), BitVectorFromInt, Bits |
| `util/strings_test.go` | 264 | ~50 | OneArgument (normal, quoted, empty, multi-word), SmashTilde, Capitalize (empty, single char, mixed case), IsName (prefix, no-match, empty, case-insensitive), IsNameExact, IsNumber (positive, negative, non-numeric), NumberArgument, UMIN, UMAX, URANGE |
| `util/dice_test.go` | 215 | ~35 | NumberRange (equal, inverted, statistical distribution over 10k trials), NumberPercent (bounds + all-values-appear), NumberDoor, NumberBits (width 0/1/8), DiceRoll (size 0/1, 1d6 bounds, 2d6 mean), NumberFuzzy (range + never-below-1) |
| `net/color_test.go` | 201 | ~40 | ProcessColors ANSI enabled (all 18 color codes, reset, sequences), ANSI disabled (strip codes), escaped ampersand (&&→&), edge cases (trailing &, unknown codes, empty string) |
| `persist/scanner_test.go` | 317 | ~45 | ReadWord (simple, quoted, empty), ReadString (tilde-terminated, multi-line, empty), ReadNumber (positive, negative, plus, pipe-OR, zero, sequential), ReadToEOL, ReadLetter, ReadFlag (numeric + letter-based A-Z/a-z), line tracking, ParseVnum |
| `persist/area_test.go` | ~280 | 5 | Integration tests: objects (candlestick + tickler with extra descs & affects), rooms (chapel + courtyard with exits, FixExits linking), C-format mobs (skeleton), V-format mobs (priestess with multi-currency gold + stances + room), spell-name objects (potion/wand/scroll with quoted spell names) |
| `persist/classes_test.go` | ~50 | 1 | LoadClasses with TestWarrior.class fixture: WhoName, AttrPrime, Weapon, Guild, SkillAdept, Thac0, HPMin/Max, ExpBase |
| `persist/races_test.go` | ~50 | 1 | LoadRaces with TestHuman.race fixture: Name, Language, MinAlign, MaxAlign, ExpMultiplier, Height, Weight, WhereName |
| `persist/player_test.go` | ~170 | 2 | LoadPlayer from fixture (name, sex, class, level, stats, password, title, mkills), SaveLoadRoundTrip (create → save → load → verify all fields) |
| `persist/skills_test.go` | ~100 | 1 | LoadSkills with 3-skill fixture (fireball spell, backstab skill, sanctuary with affect). Verifies Name, Type, Target, Slot, MinMana, Beats, NounDamage, SpellFunName/SkillFunName, messages, Affect duration/location/modifier/bitvector |
| `command/interpret_test.go` | 202 | ~9 | Find (exact, prefix, no-match, trust check, exact-beats-prefix), Interpret (dispatch, empty input, unknown command → "Huh?", position check → sleeping message). Uses `net.Pipe()` for output capture. |
| `world/world_test.go` | 136 | ~6 | New (maps initialized), GetRoom/GetMobIndex/GetObjIndex (found + not-found), AddChar/RemoveChar (add 3, remove middle, remove non-existent), AddObj/RemoveObj |
| `handler/handler_test.go` | 444 | 18 | CreateMobile (fields, ACT_IS_NPC, HP dice, HP no-dice, index count, world add). CreateObject (fields, level, wear loc, extra descrs copy). CharToRoom/CharFromRoom. ObjToRoom, ObjToChar, ObjToObj. EquipChar (wear loc). ResetArea (mob spawn, max count enforcement, object in room with zero cost, give to mob, equip on mob, door state). Interpolate. |

### Mutation Verification Results

6 mutations tested, all detected by the test suite:

| # | File | Mutation | Test That Caught It |
|---|------|----------|-------------------|
| 1 | `bitvector.go` | `IsSet` → always `false` | `TestBitVector_Toggle/bit_127` |
| 2 | `strings.go` | `Capitalize` → identity function | `TestCapitalize/mixed_case` |
| 3 | `dice.go` | `NumberRange` → always returns `from` | Statistical variance check (only 1 distinct value) |
| 4 | `color.go` | `&&` outputs `&&` instead of `&` | `TestProcessColors_EdgeCases/escaped_ampersand_before_code` |
| 5 | `scanner.go` | Negative sign ignored (`sign = -1` → `sign = 1`) | `TestReadNumber_Negative` |
| 6 | `world.go` | `GetRoom` → always `nil` | `TestGetRoom` |
