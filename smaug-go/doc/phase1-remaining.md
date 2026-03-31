# Phase 1: Remaining Work

This document details what still needs to be done to complete Phase 1.

Phase 1 goal: "Connect via telnet, log in, walk between rooms, see descriptions."

The skeleton is working. The remaining items bring it from "demo" to "loads real game data and supports basic player persistence."

---

## 1. Wire Up Area File Loading (HIGH PRIORITY)

The area parser (`persist/area.go`) is implemented but not yet called from the boot sequence. Currently `bootDB()` in `main.go` creates 2 hardcoded test rooms.

**Tasks:**
- Replace the hardcoded rooms in `bootDB()` with a call to `persist.LoadAreas(w, areaDir+"/area")`
- Test against all `.are` files in `db/area/` (listed in `db/area/area.lst`)
- Implement `fix_exits()` equivalent in `world/` — after all areas load, resolve exit `Vnum` fields to actual `*RoomIndexData` pointers
- Debug and fix any parse errors on real area files (the parser was written from format documentation, not tested against every file variation)

**Files to modify:**
- `cmd/smaug/main.go` — replace `bootDB()` body
- `internal/world/world.go` — add `FixExits()` method

**Verification:** Server boots and reports loading hundreds of rooms/mobs/objects from the real area files.

## 2. Wire Up Color Processing (MEDIUM PRIORITY)

The color processor (`net/color.go` `ProcessColors()`) exists but is not called in the output path. Color codes like `&W`, `&D`, `&C` appear raw in telnet output.

**Tasks:**
- Call `ProcessColors()` in `DescriptorData.FlushOutput()` before writing to the socket
- Determine ANSI capability from the `PLR_ANSI` flag on the descriptor's character (default to true for simplicity)

**Files to modify:**
- `internal/types/descriptor.go` — modify `FlushOutput()` to call color processing

## 3. Player Save/Load (`persist/player.go`) (HIGH PRIORITY)

Currently, `enterGame()` creates a fresh character every login. No persistence.

**Tasks:**
- Create `persist/player.go` with `LoadPlayer(r io.Reader) (*CharData, error)` and `SavePlayer(w io.Writer, ch *CharData) error`
- Match the C save format from `save.c`: `fwrite_char`/`fread_char` with `KEY value` pairs inside `#PLAYER`/`#OBJECT`/`#AFFECT` sections
- Integrate into the nanny state machine:
  - On `CON_GET_NAME`: check if `db/player/<first letter>/<Name>` exists
  - If exists: load the player data, prompt for password
  - If not: start character creation flow (new name confirmation, sex, race, class)
- Integrate save on quit (`DoQuit`) and periodic autosave

**C reference:** `src/save.c` — `fwrite_char()`, `fread_char()`, `save_char_obj()`, `load_char_obj()`

**Files to create:**
- `internal/persist/player.go`

**Files to modify:**
- `internal/game/loop.go` — expand `nanny()` and `enterGame()` to load/create players
- `internal/act/info.go` — modify `DoQuit` to save before disconnect

## 4. Character Creation Flow (MEDIUM PRIORITY)

Currently skipped — any name + any password enters the game.

**Tasks:**
- Implement `CON_CONFIRM_NEW_NAME`, `CON_GET_NEW_PASSWORD`, `CON_CONFIRM_NEW_PASSWORD`, `CON_GET_NEW_SEX`, `CON_GET_NEW_RACE`, `CON_GET_NEW_CLASS` states in `nanny()`
- Use `RaceData` and `ClassType` from `world.Races` and `world.Classes` for menus
- Set starting stats based on race/class from the loaded `.race`/`.class` files

**Prerequisites:** Class/race file loading (see item 5).

## 5. Class and Race File Loading (MEDIUM PRIORITY)

**Tasks:**
- Create `persist/classes.go` — load `.class` files from `db/classes/` (listed in `db/classes/class.lst`)
- Create `persist/races.go` — load `.race` files from `db/races/` (listed in `db/races/race.lst`)
- Call from `bootDB()`

**C reference:** `src/db.c` — `load_class_file()`, `load_race_file()`

## 6. Skills/Spells Data Loading (LOW PRIORITY for Phase 1)

**Tasks:**
- Create `persist/skills.go` — load `skills.dat` into `world.Skills`
- Assign GSN variables (`GsnBackstab`, etc.)

Not strictly needed for Phase 1 walking-around demo, but needed before Phase 2 combat.

## 7. System Data Loading (LOW PRIORITY for Phase 1)

**Tasks:**
- Create `persist/system.go` — load system data file into `world.SysData`

## 8. Help File Loading (LOW PRIORITY)

**Tasks:**
- Load help entries from area files (the `help.are` file contains `#HELPS` section)
- Populate `world.Helps` for the `DoHelp` command

## 9. Prompt System (LOW PRIORITY)

**Tasks:**
- Replace the bare `> ` prompt with the player's configured prompt string
- Implement prompt token substitution: `%h` = current hp, `%H` = max hp, `%m` = mana, `%M` = max mana, `%v` = move, `%V` = max move, etc.
- Send prompt after each command output

**Files to modify:**
- `internal/game/loop.go` — send prompt after command processing

## 10. Mob Instantiation on Area Reset (LOW PRIORITY for Phase 1)

**Tasks:**
- After loading area files, process resets to place mob and object instances in rooms
- Create `handler/char.go` with `CreateMobile(mobIndex)` and `CharToRoom(ch, room)`
- Create `handler/obj.go` with `CreateObject(objIndex, level)` and `ObjToRoom(obj, room)`

This is the bridge between "area files loaded" and "mobs/objects visible in rooms."

---

## 11. Tests for New Code (ONGOING — TDD REQUIRED)

All new code must follow TDD. See `CLAUDE.md` at repo root for the full TDD workflow.

**Current test coverage** (as of 2026-03-30):
- `types/bitvector_test.go` — 435 lines, ~60 tests
- `util/strings_test.go` — 264 lines, ~50 tests
- `util/dice_test.go` — 215 lines, ~35 tests
- `net/color_test.go` — 201 lines, ~40 tests
- `persist/scanner_test.go` — 317 lines, ~45 tests
- `command/interpret_test.go` — 202 lines, ~9 tests
- `world/world_test.go` — 136 lines, ~6 tests

**Packages still needing tests:**
- `persist/area.go` — needs integration tests loading real `.are` snippets from `testdata/`
- `game/loop.go` — hard to unit test (game loop + nanny), but `enterGame()` could be tested
- `act/info.go` — commands could be tested with fake descriptor output capture (same `net.Pipe()` pattern as `command/interpret_test.go`)

**When writing new persist loaders** (player.go, classes.go, races.go), write the test first using known fixture data, then implement the loader.

---

## Priority Order

1. **Wire up area file loading** — makes the world real
2. **Fix exits** — makes navigation work across loaded areas
3. **Player save/load** — persistence (TDD: write test with fixture player file first)
4. **Color processing** — visual quality
5. **Class/race loading** — needed for character creation (TDD: write test with fixture .class/.race first)
6. **Character creation** — new player flow
7. **Mob/obj instantiation** — populate the world
8. **Area loader integration test** — load real chapel.are, verify room/mob/obj counts
9. Everything else
