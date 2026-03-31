# Phase 1: Remaining Work

This document details what still needs to be done to complete Phase 1.

Phase 1 goal: "Connect via telnet, log in, walk between rooms, see descriptions."

The server loads 1,909 rooms from 26 area files and players can walk around real SMAUG rooms. The remaining items add player persistence, character creation, populated rooms (NPCs/objects), and parser robustness.

---

## Completed (this phase)

- ~~Wire up area file loading~~ — `persist.LoadAreas()` called from `bootDB()`, loads all `.are` files
- ~~Fix exits~~ — `world.FixExits()` resolves vnum→room pointers (4,299 exits linked)
- ~~Wire color processing~~ — `ColorFunc` on `DescriptorData`, ANSI codes render in clients
- ~~Area loader integration tests~~ — 4 tests with testdata fixtures (objects, rooms, C-mobs, V-mobs)
- ~~Makefile + README~~ — `make build`, `make test`, `make check`, `make run`

---

## 1. Area Parser: Spell-Name Values (HIGH PRIORITY)

Potions, scrolls, wands, pills, and staves store spell names (in quotes) in their value fields. The parser currently reads all values as numbers and chokes on these. ~370 BUG warnings during boot, mostly from this.

**Tasks:**
- In `parseObjValues`, detect quoted spell names (e.g., `'heal' 'NONE' 'NONE'`)
- Look up spell names in the skill table (requires skills.dat to be loaded first, or store as strings and resolve later)
- Handle the two-line value format: numeric values on line 1, spell names on line 2

**C reference:** `src/db.c` — the object value reading checks `item_type` and calls `skill_lookup()` for spell-containing items

**Impact:** Fixes loading of all potions, scrolls, wands, pills, staves in all area files

## 2. Player Save/Load (`persist/player.go`) (HIGH PRIORITY)

Currently `enterGame()` creates a fresh character every login. No persistence.

**Tasks:**
- Create `persist/player.go` with `LoadPlayer(r io.Reader) (*CharData, error)` and `SavePlayer(w io.Writer, ch *CharData) error`
- Match the C save format from `save.c`: `KEY value` pairs inside `#PLAYER`/`#OBJECT`/`#AFFECT` sections
- Integrate into the nanny state machine:
  - On `CON_GET_NAME`: check if `db/player/<first letter>/<Name>` exists
  - If exists: load player data, prompt for password
  - If not: start character creation flow
- Integrate save on quit (`DoQuit`) and periodic autosave
- **TDD: write test with a fixture player file first**

**C reference:** `src/save.c` — `fwrite_char()`, `fread_char()`

## 3. Character Creation Flow (MEDIUM PRIORITY)

Currently skipped — any name + any password enters the game.

**Tasks:**
- Implement `CON_CONFIRM_NEW_NAME`, `CON_GET_NEW_PASSWORD`, `CON_CONFIRM_NEW_PASSWORD`, `CON_GET_NEW_SEX`, `CON_GET_NEW_RACE`, `CON_GET_NEW_CLASS` states in `nanny()`
- Use `RaceData` and `ClassType` from loaded `.race`/`.class` files for menus
- Set starting stats based on race/class

**Prerequisites:** Class/race file loading (item 4)

## 4. Class and Race File Loading (MEDIUM PRIORITY)

**Tasks:**
- Create `persist/classes.go` — load `.class` files from `db/classes/`
- Create `persist/races.go` — load `.race` files from `db/races/`
- Call from `bootDB()`
- **TDD: write test with fixture .class/.race files first**

**C reference:** `src/db.c` — `load_class_file()`, `load_race_file()`

## 5. Mob/Object Instantiation on Area Reset (MEDIUM PRIORITY)

After loading area files, process resets to place mob and object instances in rooms.

**Tasks:**
- Create `handler/char.go` with `CreateMobile(mobIndex)` and `CharToRoom(ch, room)`
- Create `handler/obj.go` with `CreateObject(objIndex, level)` and `ObjToRoom(obj, room)`
- Process reset commands (M, O, P, G, E, D, H) during boot

This is the bridge between "area files loaded" and "NPCs/objects visible in rooms."

## 6. Skills/Spells Data Loading (MEDIUM PRIORITY)

**Tasks:**
- Create `persist/skills.go` — load `skills.dat` into `world.Skills`
- Assign GSN variables (`GsnBackstab`, etc.)
- Needed for spell-name resolution in object values (item 1) and for Phase 2 combat

## 7. System Data Loading (LOW PRIORITY)

**Tasks:**
- Create `persist/system.go` — load system data file into `world.SysData`

## 8. Help File Loading (LOW PRIORITY)

**Tasks:**
- Parse `#HELPS` section in area files (currently skipped by `skipSection`)
- Populate `world.Helps` for the `DoHelp` command

## 9. Prompt System (LOW PRIORITY)

**Tasks:**
- Replace bare prompt with player's configured prompt string
- Implement token substitution: `%h`/`%H` (hp), `%m`/`%M` (mana), `%v`/`%V` (move), etc.
- Send prompt after each command output

---

## 10. Tests for New Code (ONGOING — TDD REQUIRED)

All new code must follow TDD. See `CLAUDE.md` at repo root for the full TDD workflow.

**Current test coverage** (as of 2026-03-30):
- `types/bitvector_test.go` — 435 lines, ~60 tests
- `util/strings_test.go` — 264 lines, ~50 tests
- `util/dice_test.go` — 215 lines, ~35 tests
- `net/color_test.go` — 201 lines, ~40 tests
- `persist/scanner_test.go` — 317 lines, ~45 tests
- `persist/area_test.go` — 233 lines, 4 tests (integration with testdata fixtures)
- `command/interpret_test.go` — 202 lines, ~9 tests
- `world/world_test.go` — 136 lines, ~6 tests

**Packages still needing tests:**
- `game/loop.go` — hard to unit test (game loop + nanny), but `enterGame()` could be tested
- `act/info.go` — commands could be tested with `net.Pipe()` output capture

**When writing new persist loaders**, write the test first using known fixture data, then implement the loader.

---

## Priority Order

1. **Spell-name values in objects** — fixes ~370 parse errors, loads all potions/scrolls/wands
2. **Player save/load** — persistence (TDD)
3. **Class/race loading** — needed for character creation (TDD)
4. **Character creation** — new player flow
5. **Skills data loading** — needed for spell resolution and Phase 2
6. **Mob/obj instantiation** — populate rooms with NPCs and items
7. Everything else (system data, helps, prompts)
