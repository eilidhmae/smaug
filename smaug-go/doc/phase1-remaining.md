# Phase 1: Remaining Work

This document details what still needs to be done to complete Phase 1.

Phase 1 goal: "Connect via telnet, log in, walk between rooms, see descriptions."

The server loads 1,909 rooms, 17 classes, and 15 races. Player save/load is implemented. The remaining items wire player persistence into the login flow, add character creation, and populate rooms.

---

## Completed (this phase)

- ~~Wire up area file loading~~ — `persist.LoadAreas()` called from `bootDB()`
- ~~Fix exits~~ — `world.FixExits()` resolves vnum→room pointers (4,299 exits)
- ~~Wire color processing~~ — `ColorFunc` on `DescriptorData`, ANSI codes render
- ~~Area loader integration tests~~ — 5 tests with testdata fixtures
- ~~Makefile + README~~ — `make build`, `make test`, `make check`, `make run`
- ~~Spell-name object values~~ — peek-based detection, stores in `SpellNames[6]string`
- ~~Class file loading~~ — `persist.LoadClasses()`, 17 classes loaded from real data (TDD)
- ~~Race file loading~~ — `persist.LoadRaces()`, 15 races loaded from real data (TDD)
- ~~Player save/load~~ — `persist.LoadPlayer()` / `persist.SavePlayer()`, round-trip tested (TDD)

---

## 1. Wire Player Save/Load into Login Flow (HIGH PRIORITY)

`persist.LoadPlayer()` and `persist.SavePlayer()` exist and are tested, but not wired into the game yet. `enterGame()` still creates fresh characters.

**Tasks:**
- In `nanny()` `CON_GET_NAME`: check if `persist.PlayerFilePath(w.DataDir, name)` exists
  - If exists: load player with `persist.LoadPlayer()`, prompt for password
  - If not: start character creation flow (or simplified new-player flow)
- In `nanny()` `CON_GET_OLD_PASSWORD`: verify password against `ch.PCData.Pwd`
- In `DoQuit`: call `persist.SavePlayer()` before disconnect
- Create player save directory if needed (`db/player/<letter>/`)

## 2. Character Creation Flow (HIGH PRIORITY)

Currently skipped — any name + any password enters the game.

**Tasks:**
- Implement `CON_CONFIRM_NEW_NAME`, `CON_GET_NEW_PASSWORD`, `CON_CONFIRM_NEW_PASSWORD`, `CON_GET_NEW_SEX`, `CON_GET_NEW_RACE`, `CON_GET_NEW_CLASS` states in `nanny()`
- Use loaded `RaceData` and `ClassType` for menus
- Set starting stats based on race/class bonuses

## 3. Mob/Object Instantiation on Area Reset (MEDIUM PRIORITY)

After loading area files, process resets to place mob and object instances in rooms.

**Tasks:**
- Create `handler/char.go` with `CreateMobile(mobIndex)` and `CharToRoom(ch, room)`
- Create `handler/obj.go` with `CreateObject(objIndex, level)` and `ObjToRoom(obj, room)`
- Process reset commands (M, O, P, G, E, D, H) during boot

This is the bridge between "area files loaded" and "NPCs/objects visible in rooms."

## 4. Skills/Spells Data Loading (MEDIUM PRIORITY)

**Tasks:**
- Create `persist/skills.go` — load `skills.dat` into `world.Skills`
- Assign GSN variables (`GsnBackstab`, etc.)
- Needed for spell-name resolution in object SpellNames and for Phase 2 combat

## 5. System Data Loading (LOW PRIORITY)

**Tasks:**
- Create `persist/system.go` — load system data file into `world.SysData`

## 6. Help File Loading (LOW PRIORITY)

**Tasks:**
- Parse `#HELPS` section in area files (currently skipped by `skipSection`)
- Populate `world.Helps` for the `DoHelp` command

## 7. Prompt System (LOW PRIORITY)

**Tasks:**
- Replace bare prompt with player's configured prompt string
- Implement token substitution: `%h`/`%H` (hp), `%m`/`%M` (mana), `%v`/`%V` (move), etc.
- Send prompt after each command output

---

## 8. Tests for New Code (ONGOING — TDD REQUIRED)

All new code must follow TDD. See `CLAUDE.md` at repo root for the full TDD workflow.

**Current test coverage** (as of 2026-03-30):
- `types/bitvector_test.go` — 435 lines, ~60 tests
- `util/strings_test.go` — 264 lines, ~50 tests
- `util/dice_test.go` — 215 lines, ~35 tests
- `net/color_test.go` — 201 lines, ~40 tests
- `persist/scanner_test.go` — 317 lines, ~45 tests
- `persist/area_test.go` — ~280 lines, 5 tests (objects, rooms, C-mobs, V-mobs, spell objects)
- `persist/classes_test.go` — ~50 lines, 1 test (Warrior fixture)
- `persist/races_test.go` — ~50 lines, 1 test (Human fixture)
- `persist/player_test.go` — ~170 lines, 2 tests (load fixture + save/load round-trip)
- `command/interpret_test.go` — 202 lines, ~9 tests
- `world/world_test.go` — 136 lines, ~6 tests

**Packages still needing tests:**
- `game/loop.go` — hard to unit test, but `enterGame()` could be tested
- `act/info.go` — commands could be tested with `net.Pipe()` output capture

---

## Priority Order

1. **Wire player save/load into login flow** — persistence end-to-end
2. **Character creation** — new player flow with race/class selection
3. **Mob/obj instantiation** — populate rooms with NPCs and items
4. **Skills data loading** — needed for Phase 2 combat
5. Everything else (system data, helps, prompts)
