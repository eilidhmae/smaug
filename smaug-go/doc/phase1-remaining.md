# Phase 1: Remaining Work

This document details what still needs to be done to complete Phase 1.

Phase 1 goal: "Connect via telnet, log in, walk between rooms, see descriptions."

**Status**: Phase 1 is nearly complete. The server boots with 1,909 rooms, 406 mobs, 710 objects, 325 skills, 17 classes, 15 races. Players can create new characters (sex/race/class selection), log in with saved characters (password verification), walk rooms, see NPCs, and save on quit.

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
- ~~Wire player save/load into login flow~~ — check for existing player file, password verify, save on quit
- ~~Character creation flow~~ — 6 nanny states (confirm name, password, sex, class, race), racial stat bonuses
- ~~Mob/object instantiation~~ — `handler` package with `CreateMobile`, `CreateObject`, area reset processing (406 mobs, 710 objects)
- ~~Skills data loading~~ — `persist.LoadSkills()`, 325 skills/spells from `skills.dat` (TDD)
- ~~Handler tests~~ — 18 tests covering CreateMobile, CreateObject, placement functions, reset processing

---

## Remaining (LOW PRIORITY)

### 1. System Data Loading

**Tasks:**
- Create `persist/system.go` — load system data file into `world.SysData`

### 2. Help File Loading

**Tasks:**
- Parse `#HELPS` section in area files (currently skipped by `skipSection`)
- Populate `world.Helps` for the `DoHelp` command

### 3. Prompt System

**Tasks:**
- Replace bare prompt with player's configured prompt string
- Implement token substitution: `%h`/`%H` (hp), `%m`/`%M` (mana), `%v`/`%V` (move), etc.
- Send prompt after each command output

### 4. GSN Variable Assignment

**Tasks:**
- After skills load, assign global skill number variables (`GsnBackstab`, etc.)
- Needed for Phase 2 combat and spell-name resolution in objects

---

## Tests (ONGOING)

**Current test coverage** (as of 2026-03-30):
- `types/bitvector_test.go` — 435 lines, ~60 tests
- `util/strings_test.go` — 264 lines, ~50 tests
- `util/dice_test.go` — 215 lines, ~35 tests
- `net/color_test.go` — 201 lines, ~40 tests
- `persist/scanner_test.go` — 317 lines, ~45 tests
- `persist/area_test.go` — ~280 lines, 5 tests
- `persist/classes_test.go` — ~50 lines, 1 test
- `persist/races_test.go` — ~50 lines, 1 test
- `persist/player_test.go` — ~170 lines, 2 tests
- `persist/skills_test.go` — ~100 lines, 1 test (3 skills with affects)
- `command/interpret_test.go` — 202 lines, ~9 tests
- `world/world_test.go` — 136 lines, ~6 tests
- `handler/handler_test.go` — 444 lines, 18 tests

**13 test files, ~2,910 lines, 312 test cases — all passing.**

**Packages still needing tests:**
- `game/loop.go` — hard to unit test (nanny state machine), but testable with mock descriptors
- `act/info.go` — commands testable with `net.Pipe()` output capture

---

## Priority Order

Phase 1 core functionality is complete. Remaining items are polish:
1. Help file loading (most user-visible)
2. Prompt system
3. System data loading
4. GSN variable assignment (Phase 2 prerequisite)
