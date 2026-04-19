# Phase 6 — Overland Builder

**Status:** Planned 2026-04-18; external adversary review recommended before dispatch.
**Priority:** Wave-2 / Medium.
**Scope:** Admin/immortal mutation commands on the overland — `setmark` (landmark CRUD), `setexit` (entrance CRUD), `mapresets` (listing), `mreset` (reset CRUD), `mapedit` (sector editing + floodfill + save). Writers for `landmarks.dat` / `entrances.dat` / `mapresets.dat` and the per-map `.raw` file. Landing-site persistence is out of scope (dragonflight).
**Depends on:** `plan-phase6-overland-loader.md` landed and all its criteria satisfied.
**Splits with:** loader plan (read-side) — no overlap; builder writes everything the loader reads.

Mutation verification for every task group in this plan uses `Edit` round-trips only. Banned project-wide: `git checkout` / `git restore` / `git reset --hard` / `git stash`.

---

## Problem

Once the loader plan ships, immortals can teleport onto the overland and survey / look / run `coords` — but the maps themselves cannot be edited without leaving the game, hand-editing binary `.raw` files, and rebooting. The C overland code (`src/overland.c:958-3550`) ships five mutation commands that let builders work entirely online:

- `setmark` — create / delete / edit landmarks at current coordinates.
- `setexit` — create / delete / edit overland entrances (link maps to each other or to regular zones).
- `mapresets` — dump the current mapreset table (read-side, but bundled with the mutation suite).
- `mreset` — add / delete mapresets at current coordinates.
- `mapedit` — toggle sector-edit mode, paint sectors by walking, floodfill, undo, save to disk, reload.

Also needed for round-trip workflows:

- Writers: `save_landmarks` (`:759`), `save_entrances` (`:1217`), `save_mapresets` (`:1671`), `save_map` (`:3319`).
- Editor session state: `PLR_MAPEDIT` flag, `pcdata->secedit` value, `substate == SUB_OVERLAND_DESC` for landmark descriptions.
- Descriptor menu substate for `mapedit` — this plan uses flat subcommands (matching C) rather than a `CON_MAPEDIT` interactive substate. The `redit` OLC plan (shipped as Wave-D `plan-phase6-olc-redit.md`) shows the substate pattern; overland `mapedit` does not need it.

Without this plan, every overland edit requires offline tooling. With it, worldbuilders can iterate live.

---

## C Reference (authoritative)

Line cites against `src/overland.c` unless noted.

### Landmark builder path

- `do_setmark` — `:958-1091`. Substate-aware landmark editor. Handles: `add`, `delete`, `distance N`, `desc` (multi-line editor), `isdesc` (toggle). `SUB_OVERLAND_DESC` is used to save an in-editor description back onto `dest_buf` (= `landmark`).
- `add_landmark( map, x, y )` — `:804-819`. Append to list; save.
- `delete_landmark( LANDMARK_DATA * )` — `:821-837`. Remove from list; save.
- `save_landmarks()` — `:759-787`. Writer.
- `check_landmark( map, x, y )` — `:789-802`. Linear lookup (used by both loader and builder).

### Entrance builder path

- `do_setexit` — `:1332-1483`. Subcommands: `create`, `delete`, `<vnum>` (link to regular zone), `<area-name>`, `map <mapname> <x> <y>` (link to another overland map).
- `add_entrance( tomap, onmap, hereX, hereY, thereX, thereY, vnum )` — `:1293-1312`. Seeds `prevsector` from current terrain, empty area string, appends, saves.
- `modify_entrance(...)` — `:1266-1291`. Mutate fields + save.
- `delete_entrance` — `:1314-1328`.
- `save_entrances()` — `:1217-1249`.
- `check_entrance` — `:1251-1264`.

### Mapreset builder path

- `do_mapresets` — `:1758-1785`. **Read-only dump.** Even though named *resets*, this command only lists — bundle into the builder plan since it pairs with `mreset`.
- `do_mreset` — `:1788-1959`. Subcommands: `add <object|mobile> <vnum-or-name>`, `delete <object|mobile> <vnum>`.
- `add_mapreset` — `:1701-1726`.
- `delete_mapreset` — `:1728-1742`.
- `check_mapreset(ch)` — `:1744-1755`. Linear lookup at ch's coords.
- `save_mapresets()` — `:1671-1699`.

### Mapedit + floodfill + save_map

- `do_mapedit` — `:3363-3550`. Toggle (no arg), `sector <type>`, `save [mapname]`, `fill <type>`, `undo`, `reload [confirm]`.
- `floodfill` — `:3204-3244`. Recursive 4-connected flood; stores every changed cell into the `undo` linked list. **Stack-recursive** — C comment warns "trying to floodfill a section that is too large will overflow the memory in short order and cause a rather uncool crash." Port as iterative BFS for safety.
- `unfloodfill` — `:3249-3270`. Walks the undo list and reverts; post-undo the list is re-armed as the redo path.
- `purgeundo` — `:3274-3287`. Clears the undo list; called between successful floodfills.
- `save_map( name, map )` — `:3319-3355`. Writes 3-byte RGB triples from `SectorTable` per tile. Lowercases the filename first. Output order matches loader: outer y, inner x.
- `reload_map` — `:3292-3311`. Zero the grid to SECT_OCEAN, call `load_mapfile` again. Dangerous — confirmation required at command level.
- `sect_edit` state — `CharData.PCData.SecEdit int` (need to add in Go); paints whenever PC is `PLR_MAPEDIT`-flagged and the tile under their feet is != `SECT_EXIT`. The paint happens in `display_map` at `:2255-2259` (loader path), not in a mutation command. The builder plan must ensure `DisplayMap` in the loader has this paint-in-place behavior — **loader G9 already ports this** (verify before starting here).

### PCData.SecEdit

- `pcdata->secedit` — `mud.h:3134`-ish (`int secedit`). Stores the currently-selected sector index for mapedit mode.
- Persistence: stored in pfile. C `save_char_obj` handles this; Go's `persist/player.go` does NOT currently save it — must add.

### `get_sectypes( char *sector )`

- `overland.c:3563-3571`. Takes a sector name string, returns the matching index or -1. Used by `mapedit sector <type>`, `mapedit fill <type>`, and historically by `redit` for sector-type edits. **Already ported?** Verify via `Grep "GetSectypes\|getSectypes" internal/`. If absent → G0 pre-req.

---

## Go Current State (gap analysis)

Assumes the loader plan has landed. Gap analysis relative to a post-loader state.

### Present (after loader lands)

- `internal/overland/` package with `SectorTable`, grid, geometry, display, `DoCoords`, `DoSurvey`, `DoLandmarks`.
- `internal/types/overland.go` with `LandmarkData`, `EntranceData`, `MapResetData`.
- `World.Landmarks / Entrances / MapResets / MapSectors`.
- `PLR_ONMAP`, `ACT_ONMAP`, `PLR_HOLYLIGHT`, `ROOM_MAP`, `ITEM_ONMAP`, `PLR_MAPEDIT` constants (loader G1+Q2+Q4).
- `persist.LoadLandmarks / LoadEntrances / LoadMapresets / LoadMapSectors`.
- `overland.WorldRef` set at boot.
- `game.EditorSave` callback pattern (Tier 12) — reusable for `setmark desc`.

### Missing

- No writers: `persist.SaveLandmarks / SaveEntrances / SaveMapresets / SaveMap`.
- No mutation commands: `DoSetmark / DoSetexit / DoMapresets / DoMreset / DoMapedit`.
- No `PCData.SecEdit int` field and no persistence for it.
- No `CheckLandmark / CheckEntrance / CheckMapreset / ModifyEntrance / AddLandmark / DeleteLandmark / AddEntrance / DeleteEntrance / AddMapreset / DeleteMapreset` helpers.
- No `GetSectypes` helper (likely — verify in G0).
- No `FloodFill / UnfloodFill / PurgeUndo` engine.
- No `PutTerr` exported sector mutator — `overland.MapSectorGrid` has it (loader G3) but package-private; builder needs public export or a package-internal call site.
- No `SUB_OVERLAND_DESC`-equivalent state flow for `setmark desc`.

---

## Go Design

### Package layout (additions to loader)

```
internal/
  overland/
    writers.go       // PutTerr public seam (if not already) + RefreshCharDisplay helper
    floodfill.go     // FloodFill engine + undo stack
    setmark.go       // DoSetmark
    setexit.go       // DoSetexit
    mreset.go        // DoMresets + DoMreset
    mapedit.go       // DoMapedit + save_map + reload_map
    setmark_test.go
    setexit_test.go
    mreset_test.go
    mapedit_test.go
    floodfill_test.go
  persist/
    overland_write.go   // SaveLandmarks / SaveEntrances / SaveMapresets / SaveMap
    overland_write_test.go
  types/
    pcdata.go        // add SecEdit int
  persist/
    player.go        // save/load SecEdit
```

### `PCData.SecEdit`

New integer on `PCData`. Default `SECT_OCEAN` (same as C default). Persisted in the player file as a single-line `Secedit N` key pattern. Matches the `pagelen` / `wizinvis` precedent.

### FloodFill engine

**Replace C's recursive flood with an iterative BFS** using a `[]struct{ x, y int }` work-queue. Rationale:

- C at `:3361` explicitly warns about stack overflow on large flood regions.
- Go's goroutine stacks are growable but default `--stacksize` is 8 MB; a 1000x1000 fill could still OOM on a small VPS.
- Iterative flood is a well-known safe pattern.

Preserve C's undo list semantics — every tile write goes onto the undo stack with its PRE-fill terrain. Undo walks the stack in order, re-writing terrains. Re-undo (`unfloodfill` on already-undone state) is a redo.

```go
// overland/floodfill.go
type floodUndoEntry struct {
    Map, X, Y int
    PrevTerr  int
}

type FloodUndoStack struct {
    entries []floodUndoEntry
}

func (g *MapSectorGrid) FloodFill(m, x, y, fill, existing int, undo *FloodUndoStack) error {
    if fill == existing {
        return errors.New("cannot floodfill identical terrain")
    }
    queue := []struct{ x, y int }{{x, y}}
    for len(queue) > 0 {
        p := queue[0]
        queue = queue[1:]
        if g.GetTerrain(m, p.x, p.y) != existing {
            continue
        }
        undo.entries = append(undo.entries, floodUndoEntry{m, p.x, p.y, existing})
        g.PutTerr(m, p.x, p.y, fill)
        queue = append(queue, struct{ x, y int }{p.x+1, p.y},
            struct{ x, y int }{p.x-1, p.y},
            struct{ x, y int }{p.x, p.y+1},
            struct{ x, y int }{p.x, p.y-1})
    }
    return nil
}
```

Per-PC or world-level undo? **World-level (package-global) `undoStack`.** C uses a global — preserve.

Cap: limit queue length to 1_000_000 before aborting with "Floodfill aborted — too large." Protects against accidental whole-map floods. (New invariant; C has no cap and relies on stack to blow up.)

### Writer file I/O

Each writer:

1. Write to `<target>.tmp` with a `os.O_CREATE | os.O_TRUNC | os.O_WRONLY` handle.
2. Rename to `<target>` atomically (Linux `rename(2)` is atomic for within-filesystem moves).
3. Missing-directory path returns nil error but logs `util.Bug` — same policy as loaders.

Matches the existing `persist/player.go:SavePlayer` pattern (`SaveAtomic` helper — verify at G0).

### `save_map` portability

Three bytes per tile, no BOM, no delimiters. Go: `bufio.Writer` + `WriteByte(r); WriteByte(g); WriteByte(b)` per tile. Lowercase filename (`strings.ToLower`). Join with `filepath.Join(world.DataDir, "maps", name+".raw")`.

### `DoSetmark` description-editor integration

The C `substate = SUB_OVERLAND_DESC` path opens a multi-line description editor, and the `/s` save dispatches a closure that updates `landmark->description` and calls `save_landmarks`. This is **exactly** the Tier 12 `EditorSave` callback pattern. Reuse it:

1. `DoSetmark desc` calls `StartEditing(ch, landmark.Description)` with an `EditorSave` closure:
   ```go
   ch.EditorSave = func(c *types.CharData) {
       landmark.Description = game.CopyBufferFunc(c)
       persist.SaveLandmarks(overland.WorldRef)
       c.SendToChar("Description set.\r\n")
   }
   ```
2. Editor `/s` handler transitions `CON_EDITING → CON_PLAYING` then fires the closure (already shipped in Tier 12).

No new descriptor substate is needed. The existing editor integration handles everything.

### `mapedit` vs interactive `CON_MAPEDIT`

**Keep `mapedit` as flat subcommands**, matching C:
- `mapedit` (no arg) → toggle `PLR_MAPEDIT`.
- `mapedit sector <type>`.
- `mapedit save [mapname]`.
- `mapedit fill <type>`.
- `mapedit undo`.
- `mapedit reload [confirm]`.
- `mapedit help`.

Reason: C's `do_mapedit` is simple subcommand dispatch; no nested menu state. Matches the OLC pattern we've used for `mset`/`oset`/`rset` (flat), not `redit`/`medit` (menu-driven). Keeping builder parity reduces surprise.

### Display-refresh after mutation

C's `do_setmark add` / `do_setexit create` / etc. do NOT auto-re-render the map after a mutation. The builder must type `look` to see the change. Preserve. (Q5 in Open Questions offers a UX improvement; recommended answer: preserve C behavior.)

### Persistence lifecycle

After every mutation: save the corresponding file immediately (C's pattern — `add_landmark` ends with `save_landmarks()`). **Trade-off:** file I/O on every keystroke-level mutation. Stock data is small (10 entrances, 1 landmark) — negligible. Preserve.

### Reload safety

`mapedit reload confirm` triggers a **disk-to-memory** re-read of the current map. Current in-memory edits are LOST. Require explicit `confirm` token (C does this at `:3409-3413`). Preserve.

**Race concern:** if a second immortal is mid-floodfill on the same map while reload runs, state diverges. No mitigation — overland is synchronous, single-goroutine-updated; the scenario is impossible in Go's game loop (same as C's single-thread model).

---

## Task Groups (test-first; mutation-verify via `Edit` round-trips only)

### G0 — Pre-flight audit

**Deliverable:** Confirm loader-plan prerequisites. Determine status of `GetSectypes`, `PutTerr` export, `PLR_MAPEDIT`, `PCData.SecEdit`.

**Files to touch:**
- None if all present.
- Otherwise flag gaps inline and block on the loader plan landing them.

**Tests:** None (meta).

### G1 — `PCData.SecEdit` + persistence

**Deliverable:** Add `SecEdit int` field to `PCData`. Default `SECT_OCEAN`. Save/load in `persist/player.go`.

**Files to touch:**
- `internal/types/pcdata.go` — add field.
- `internal/persist/player.go` — add `Secedit` key in save + load blocks (grep `Wizinvis` to find neighbors).
- `internal/persist/player_test.go` — round-trip test.

**Tests (first):**
1. `TestSecEditDefaultsToOcean` — fresh `PCData` zero-value has `SecEdit == SECT_OCEAN`. (Adjust: `int(0) == SECT_INSIDE` in new enum; need explicit default. Either redefine to `SECT_INSIDE` or ensure `NewPCData` helper sets `SecEdit = SECT_OCEAN`.)
2. `TestPlayerRoundTripPreservesSecedit` — save a char with `SecEdit = SECT_FOREST`, reload, assert same value.

**Mutation verify:**
- Drop the `Secedit` key from save → test 2 fails.

### G2 — Writers: `SaveLandmarks / SaveEntrances / SaveMapresets`

**Deliverable:** Three atomic (temp-file + rename) writers in `persist/overland_write.go`.

**Files to touch:**
- `internal/persist/overland_write.go` NEW.
- `internal/persist/overland_write_test.go` NEW.

**Tests (first):**
1. `TestSaveLandmarksRoundTrip` — seed 1 landmark, call `SaveLandmarks`, clear state, call `LoadLandmarks`, assert same data.
2. `TestSaveEntrancesRoundTrip` — same for entrances with all 7 fields populated.
3. `TestSaveMapresetsRoundTrip` — same for mapresets with `Type=TYPE_OBJECT` + `Type=TYPE_MOBILE`.
4. `TestSaveLandmarksEmitsEndSentinel` — after all records, file ends with `#END`.
5. `TestSaveEntrancesAtomicOnDiskFull` — use a 0-byte tmpfs (or mock) to force write failure mid-way; assert original file is untouched (the tmp+rename pattern protects it).
6. `TestSaveLandmarksWithEmptyDescription` — description is empty string; saves as `Description ~` (tilde immediately after key).
7. `TestSaveLandmarksSpecialChars` — description with embedded `&` color codes round-trips.
8. `TestSaveLandmarksFileMatchesStockFormatByteForByte` — seed the ONE stock landmark, save, diff against `db/maps/landmarks.dat`. Tab-separated fields; trailing blank lines.

**Mutation verify:**
- Drop the `#END` emission → test 4 fails.
- Change tab separators to spaces → test 8 fails.

### G3 — `DoSetmark`

**Deliverable:** Landmark CRUD command. Delegates `desc` subcommand to the existing `EditorSave` pattern.

**Files to touch:**
- `internal/overland/setmark.go` NEW.
- `internal/overland/setmark_test.go` NEW.
- `internal/boot/boot.go` — register.

**Tests (first):**
1. `TestDoSetmarkNPCRejected` — exact C message.
2. `TestDoSetmarkNoArgShowsUsage` — matches C's 5-line usage block.
3. `TestDoSetmarkAdd` — clear list, PC at (0, 500, 500), `setmark add` → list has 1 landmark at (0,500,500) with empty description, distance 0.
4. `TestDoSetmarkAddDuplicateRejected` — after an add, a second `setmark add` at the same spot → "There's already a landmark at this location."
5. `TestDoSetmarkDelete` — seed a landmark, `setmark delete` removes it; file is re-saved.
6. `TestDoSetmarkDeleteNoLandmark` — at an empty cell, `setmark delete` → "There is no landmark here."
7. `TestDoSetmarkDistance` — seed landmark; `setmark distance 20` sets `distance=20` and saves.
8. `TestDoSetmarkDistanceNegativeRejected` — `setmark distance -5` or `0` → "Distance must be at least 1."
9. `TestDoSetmarkDistanceNonNumericRejected` — `setmark distance forty` → "Distance must be a numeric amount."
10. `TestDoSetmarkIsdescToggle` — seed landmark `IsDesc=false`, `setmark isdesc` flips to true; again flips to false; each state change saves.
11. `TestDoSetmarkDescEntersEditor` — `setmark desc` opens editor; `ch.EditorSave` is non-nil; ch is in `CON_EDITING`.
12. `TestDoSetmarkDescSaveCallback` — simulate editor save: populate `ch.EditBuffer`, fire `ch.EditorSave(ch)`, assert landmark description is updated and `SaveLandmarks` was called.

**Mutation verify:**
- Remove the "already a landmark" duplicate-check → test 4 fails.
- Skip the `SaveLandmarks` call after distance change → add assertion in test 7 that the file on disk changed.

### G4 — `DoSetexit`

**Deliverable:** Entrance CRUD command.

**Files to touch:**
- `internal/overland/setexit.go` NEW.
- `internal/overland/setexit_test.go` NEW.
- `internal/boot/boot.go` — register.

**Tests (first):**
1. `TestDoSetexitNPCRejected`.
2. `TestDoSetexitRequiresPlrOnmap`.
3. `TestDoSetexitUsage` — 5-line usage match.
4. `TestDoSetexitCreate` — new entrance at ch's coords; `tomap = ch.Map`, `vnum = -1`, `prevsector = <current terrain>`, area = empty. Sector at ch's coords becomes `SECT_EXIT`.
5. `TestDoSetexitCreateDuplicateRejected` — already an entrance there.
6. `TestDoSetexitDelete` — removes entrance, restores `prevsector` in the grid.
7. `TestDoSetexitDeleteNoEntrance` — exact C error string.
8. `TestDoSetexitVnumLink` — `setexit 1000` where room 1000 exists → entrance points to that vnum, tomap=-1.
9. `TestDoSetexitVnumBadRoom` — `setexit 99999` → "No such room exists."
10. `TestDoSetexitMapLink` — `setexit map map2 250 250` → entrance's tomap=ACON_C2 (ordinal for map 2), therex=250, therey=250.
11. `TestDoSetexitMapLinkBadName` — `setexit map garbage 0 0` → "There isn't a map for 'garbage'."
12. `TestDoSetexitMapLinkOOB` — `setexit map map1 -1 0` / `1000 0` → OOB message matching C at `:1442-1449`.
13. `TestDoSetexitArea` — `setexit area The Keep` → entrance's area field = "The Keep".
14. `TestDoSetexitAreaEmptyArgRecursesToUsage` — `setexit area` (no arg) → calls `DoSetexit` with empty string → usage.

**Mutation verify:**
- Skip the `SECT_EXIT` paint on `create` → test 4 fails.
- Skip `prevsector` restore on `delete` → test 6 fails.

### G5 — `DoMapresets` (listing)

**Deliverable:** Read-only pager dump.

**Files to touch:**
- `internal/overland/mreset.go` NEW.
- `internal/overland/mreset_test.go` NEW.
- `internal/boot/boot.go`.

**Tests (first):**
1. `TestDoMapresetsEmpty` — zero resets → header + zero rows.
2. `TestDoMapresetsOneMobOneObject` — header + two rows (Mobile and Object types), columns match C's format (`%-7.7s %5d   %-25.25s %-10.10s %4dX %4dY`).
3. `TestDoMapresetsSortedByInsertionOrder` — insertion-order preservation (slices in Go give this for free but assert it).

**Mutation verify:**
- Truncate `objindex.ShortDescr` to 20 chars instead of 25 → test 2 fails on alignment.

### G6 — `DoMreset`

**Deliverable:** Mapreset mutator. `mreset add <object|mobile> <vnum-or-name>` / `mreset delete <object|mobile> <vnum>`.

**Files to touch:**
- `internal/overland/mreset.go`.
- `internal/overland/mreset_test.go`.

**Tests (first):**
1. `TestDoMresetAddObjectByVnum` — given a valid obj vnum, reset appended + saved.
2. `TestDoMresetAddObjectByName` — given an obj short-descr token match, found via lookup, reset appended.
3. `TestDoMresetAddMobileByVnum` — same for mob.
4. `TestDoMresetAddMobileByName` — same.
5. `TestDoMresetAddUnknownRejected` — `mreset add object 99999` → "No such object exists." (C line 1838-ish).
6. `TestDoMresetDeleteObjectVnum` — seed a reset, `mreset delete object 1234` → removed + saved.
7. `TestDoMresetDeleteNotFound` — delete for a vnum not in the list at these coords → exact C message.
8. `TestDoMresetRequiresPlrOnmap`.
9. `TestDoMresetNPCRejected`.
10. `TestDoMresetUsage` — no/invalid subcommand → 2-line usage.

**Mutation verify:**
- Skip name-lookup branch → test 2 fails.
- Remove `SaveMapresets` call after add → test 1 fails on re-read.

### G7 — `DoMapedit` sector + toggle

**Deliverable:** `mapedit` no-arg (toggle), `mapedit sector <type>`, `mapedit help`. Sector paint-on-walk is a loader-side behavior (loader G9) — this plan only exposes the on/off toggle and the current-sector selector.

**Files to touch:**
- `internal/overland/mapedit.go` NEW.
- `internal/overland/mapedit_test.go` NEW.
- `internal/boot/boot.go`.

**Tests (first):**
1. `TestDoMapeditToggleOn` — PC on map, no flag → `mapedit` sets `PLR_MAPEDIT`; message matches C.
2. `TestDoMapeditToggleOff` — PC on map, flag set → `mapedit` clears `PLR_MAPEDIT`.
3. `TestDoMapeditNPCRejected`.
4. `TestDoMapeditRequiresPlrOnmap`.
5. `TestDoMapeditHelpLists5Subcommands` — exact C help output (5 lines).
6. `TestDoMapeditSector` — `mapedit sector forest` → `ch.PCData.SecEdit = SECT_FOREST`; message matches C.
7. `TestDoMapeditSectorInvalid` — `mapedit sector xyzzy` → "Invalid sector type."
8. `TestDoMapeditSectorExitRejected` — `mapedit sector exit` → "You cannot place exits this way. Please use the setexit command for this."

**Mutation verify:**
- Invert the toggle branch → test 1 or 2 fails.

### G8 — `DoMapedit save` + `save_map`

**Deliverable:** `mapedit save [mapname]` writes the current grid to the per-map `.raw` file.

**Files to touch:**
- `internal/overland/mapedit.go`.
- `internal/persist/overland_write.go` — `SaveMap(dataDir, name string, grid *MapSectorGrid, mapNum int) error`.
- `internal/persist/overland_write_test.go`.
- `internal/overland/mapedit_test.go`.

**Tests (first):**
1. `TestSaveMapRoundTripsViaLoader` — mutate one tile in an in-memory grid, `SaveMap` to tmp, `LoadMapFile` back, assert tile equals new value.
2. `TestSaveMapAllOceanProducesStockByteContent` — save a fresh all-ocean grid to tmp; diff byte-for-byte against a fresh 1000x1000 ocean stock fixture (generate once, commit as fixture).
3. `TestSaveMapLowercasesFilename` — `SaveMap(..., "Map1", 0)` → writes `map1.raw`.
4. `TestDoMapeditSaveCurrentMap` — `mapedit save` (no arg) → saves the map the PC is on.
5. `TestDoMapeditSaveNamed` — `mapedit save map2` → saves map 2.
6. `TestDoMapeditSaveBadName` — `mapedit save garbage` → "There isn't a map for 'garbage'."

**Mutation verify:**
- Swap the byte-order per tile (G/R/B instead of R/G/B) → test 1 fails.
- Drop the `strings.ToLower` call → test 3 fails.

### G9 — Floodfill engine

**Deliverable:** Iterative-BFS `FloodFill` + `UnfloodFill` + `PurgeUndo` on the world-level grid. 1M-tile cap.

**Files to touch:**
- `internal/overland/floodfill.go` NEW.
- `internal/overland/floodfill_test.go` NEW.

**Tests (first):**
1. `TestFloodFillSameTerrainRejected` — fill=ocean, existing=ocean → returns "cannot floodfill identical terrain" error.
2. `TestFloodFillSingleTile` — 1x1 ocean tile surrounded by land, fill with grassland → exactly 1 tile changes.
3. `TestFloodFillConnectedRegion` — 10x10 ocean block, fill with grassland → all 100 tiles change.
4. `TestFloodFillDoesNotCrossBoundary` — L-shaped region, fill; non-connected tiles remain unchanged.
5. `TestFloodFillUndo` — fill, undo, grid state is restored.
6. `TestFloodFillUndoRedo` — fill, undo, undo again (redo), grid state matches the filled result.
7. `TestFloodFillPurgeUndoClearsStack` — after purge, subsequent undo is a no-op.
8. `TestFloodFillCapAborts` — synthesize a 1M+ connected region (use a small-dimensioned grid; set dimensions to 2000x2000 for the test harness) → fill aborts with cap-error, grid partially filled, undo list consistent.
9. `TestFloodFillBoundaryCondition` — tile at (0,0) with all its neighbors out-of-bounds → fills exactly one tile without walking off-grid.
10. `TestFloodFillUndoAfterReload` — fill, save, reload (via `ReloadMap`), undo is empty (purgeundo implicit on reload) — **FLAG as design decision**: C does NOT purge undo on reload (bug?); preserve verbatim? Recommended in Open Questions.

**Mutation verify:**
- Change BFS queue deque to LIFO (stack) → some pathological L-shape test fails ordering.
- Off-by-one in neighbor generation → test 9 fails.

### G10 — `DoMapedit fill / undo / reload`

**Deliverable:** Wire G9 into `mapedit fill <type>`, `mapedit undo`, `mapedit reload [confirm]`.

**Files to touch:**
- `internal/overland/mapedit.go`.
- `internal/overland/mapedit_test.go`.

**Tests (first):**
1. `TestDoMapeditFillNoArg` — "Floodfill with what???"
2. `TestDoMapeditFillValidSector` — seeds a 3x3 region at ch's coords, invokes `mapedit fill forest`, assert all 9 tiles changed, success message sent.
3. `TestDoMapeditFillInvalidSector` — "Invalid sector type."
4. `TestDoMapeditFillIdenticalTerrainRejected` — ch's current tile is ocean; `mapedit fill ocean` → "Cannot floodfill identical terrain type!"
5. `TestDoMapeditFillDisplaysMap` — after success, `display_map` is re-invoked (capture via hook or output-buffer spy).
6. `TestDoMapeditUndo` — after a fill, `mapedit undo` reverts and `display_map` re-invoked.
7. `TestDoMapeditReloadRequiresConfirm` — `mapedit reload` → "This is a dangerous command if used improperly.\r\nAre you sure about this? Confirm by typing: mapedit reload confirm".
8. `TestDoMapeditReloadConfirmed` — `mapedit reload confirm` → grid is reloaded from disk. Pre-seed a tmp-`.raw` with known content.

**Mutation verify:**
- Remove the confirm-token check → test 7 fails.
- Skip the re-display after fill → test 5 fails.

### G11 — Boot registrations + round-trip scenario

**Deliverable:** All five commands registered. End-to-end scenario: imm logs in, enters overland, adds a landmark, edits a sector, saves, reboots (simulated via re-boot of the same `World`), verifies persistence.

**Files to touch:**
- `internal/boot/boot.go` — five registrations.
- `internal/boot/boot_test.go` — non-nil assertions for the five commands.
- `internal/testclient/overland_builder_test.go` NEW.

**Tests (first):**
1. `TestBootRegistersBuilderCommands` — `command.Find("setmark")` through `command.Find("mapedit")` all non-nil.
2. **Integration** `TestBuilderAddLandmarkPersists` — via testclient: login, enter overland, `setmark add`, `setmark distance 30`, quit, re-boot, re-login, `landmarks` shows the new entry.
3. **Integration** `TestBuilderSetexitCreatePersists` — similar flow for `setexit create`.
4. **Integration** `TestBuilderMresetAddPersists` — similar for `mreset add object <vnum>`.
5. **Integration** `TestBuilderSectorEditSaveReload` — toggle mapedit, walk a few tiles (use `coords` to jump; each `coords` paints since `display_map` does the paint at ch's feet), `mapedit save`, re-boot, `mapedit` on the same PC, verify tiles persist.

**Mutation verify:**
- Drop the `setmark` registration → test 1 fails.

### G12 — Documentation + changelog

**Deliverable:** Update `CLAUDE.md` index. Append `CHANGELOG.md` entry. Mark builder follow-ups in `TODO.md`.

**Files to touch:**
- `CLAUDE.md` — add `plan-phase6-overland-builder.md`.
- `CHANGELOG.md`.
- `TODO.md`.

---

## Acceptance Criteria

- **A1.** After `Boot`, `command.Find("setmark")` / `setexit` / `mapresets` / `mreset` / `mapedit` all return non-nil handlers. (G11)
- **A2.** A legitimate builder workflow via testclient adds a landmark, changes its distance, writes a description via the editor, quits, reboots, and the changes persist in `landmarks.dat`. (G3, G11)
- **A3.** Adding / deleting an entrance paints / restores the `SECT_EXIT` sector on the grid. (G4)
- **A4.** `mreset add object <vnum>` at coords stamps a new mapreset and re-saves `mapresets.dat`; after reboot, the stored reset spawns the object via `ApplyMapResets` from the loader plan. (G6, loader G11)
- **A5.** `mapedit sector forest` + walking with `coords` paints the grid via `DisplayMap`'s paint-on-foot branch (loader G9). After `mapedit save`, tiles persist to `<mapN>.raw`. (G7, G8)
- **A6.** `mapedit fill <type>` on a connected region of ≤ 1 000 000 tiles completes without stack overflow, without going out-of-bounds, and allows `mapedit undo` to roll back. (G9, G10)
- **A7.** `mapedit reload` requires the `confirm` token. (G10)
- **A8.** `SaveLandmarks` / `SaveEntrances` / `SaveMapresets` are byte-compatible with their C counterparts on the stock fixture round-trip (G2 test 8).
- **A9.** `SaveMap` → `LoadMapFile` round-trip is byte-identity on an all-ocean grid. (G8)
- **A10.** `PCData.SecEdit` persists across save / load. (G1)
- **A11.** `mapedit sector exit` is blocked with the C-exact redirect message. (G7)
- **A12.** `setmark desc` uses the `EditorSave` callback (Tier 12) — no new CON_* substate introduced. (G3)
- **A13.** All writers are atomic: a simulated I/O failure partway through a save leaves the original file intact. (G2)
- **A14.** All 14 mutation-verify flips named in this plan produce a red test and the `Edit`-revert restores green.
- **A15.** `go test -count=3 ./...` green across all packages.

---

## Scope Cuts / Deferrals

- **Landing-site persistence** (`load_landing_sites` / `save_landing_sites`) — DRAGONFLIGHT-gated in C; not ported.
- **`foldarea` / vnum repack integration with overland** — out of scope. Tracked in roadmap.
- **Interactive `CON_MAPEDIT` substate** — not built; `mapedit` stays flat.
- **`mreset name` with spaces** — C uses `one_argument` (space-delimited, quote-support-inconsistent — see planes plan's Q on quotes). Port with space-delimited single token; deferred multi-word support tracked in TODO.
- **Auto-re-render after mutation** — C does not auto-render; preserve. (Q5)
- **Non-reset-type on load bug fix (map-3)** — that's the loader plan's Q1, not this one.
- **`check_random_mobs` / wander-mob spawn** — loader scope-cut; not this plan.
- **Multi-immortal concurrency on `mapedit`** — single-goroutine game loop makes it moot; no mutex needed.
- **Undo-list persistence across sessions** — C keeps undo in-process only; preserve. On reboot, undo is empty.

---

## Open Questions

### Q1. `reload_map` interaction with the floodfill undo stack

C's `reload_map` at `:3292-3311` re-zeroes the grid and re-reads the file, but does NOT call `purgeundo`. This means an `undo` after `reload` could partially re-apply pre-reload state, creating inconsistent grid sections. Options:

- **(a) Preserve verbatim** — document as a known bug.
- **(b) Call `PurgeUndo` at the top of `ReloadMap`** — safer; one-line fix.

**Recommended:** (b). Fidelity-to-buggy-C is not worth data corruption. Cite in code comment.

### Q2. `mapedit sector <type>` persistence

Does `SecEdit` persist across mapedit-mode toggles? C does (the field is on `PCData`; toggling `PLR_MAPEDIT` does not clear `secedit`). Preserve.

### Q3. `mreset add mobile <name>` name-lookup strategy

C uses `get_mob_index_name` (glob match against `name` / `short_descr` across all mob prototypes). Go has `world.GetMobIndex(vnum int)` but likely lacks a name-based equivalent. Options:

- **(a) Add a `world.FindMobIndexByName` helper** that linear-scans all prototypes.
- **(b) Reject `mreset add mobile <name>`** with "Name lookup not supported — use vnum."

**Recommended:** (a). It's a 15-line helper and preserves the builder workflow.

### Q4. Editor `/s` save for landmark description — callback order

Tier 12 editor `/s` transitions `CON_EDITING → CON_PLAYING` BEFORE firing `EditorSave`. The landmark save closure then re-stores nothing to `Connected` (it stays `CON_PLAYING`). Verify C does not have a substate-restore surprise — C's `do_setmark desc` saves `ch->substate = ch->tempnum` after the write. Go doesn't use `substate` for overland — tempnum-restore is N/A.

**Recommended:** Simply save description + call `SaveLandmarks`. No substate restore.

### Q5. Auto-re-render after mutations

C does NOT auto-redraw after `setmark add` etc. — builder must type `look`. Preserve? Alternative: call `DisplayMap(ch)` after every mutation. Convenience win; C fidelity loss.

**Recommended:** preserve C behavior. Builder ergonomics is a Phase-7 concern.

### Q6. `setexit` area-string validation

C's `setexit area <arg>` accepts any string with no area-existence check. A builder typing `setexit area Nonexistent Place` persists that text verbatim; the runtime never looks it up. Preserve? Or validate against `world.Areas`?

**Recommended:** preserve. The `area` field is annotation-only in C (unused by movement logic post-Phase-6 loader); validating would require exposing area-name search in `overland`.

### Q7. `SaveMap` temp-file pathing

Write `.raw.tmp` in the same directory as the target, then rename. Unix-atomic. Disk-full / permission errors surface as returned errors; G2 test 5 covers.

### Q8. Dependency: loader plan landed in full

This plan is **blocked on loader plan acceptance criteria A1-A17 being satisfied**. Specifically:

- A1 (all four loaders run) — required for G1 write round-trips to load via loader.
- A9 (color-run optimization) — required for G10's `mapedit fill` re-display test.
- A14 (mapreset spawn) — required for G6's add-mapreset-persists scenario.
- A15 (`DoLook` hook) — required for `mapedit save` → reload → visual-diff scenarios.

Any missing criterion from the loader → this plan pauses at G0.

---

## Risk Analysis

### R1. Floodfill stack overflow on huge connected regions

**Probability:** High if ported as recursive.
**Impact:** Panic; server restart.
**Mitigation:** G9 uses iterative BFS + 1M-tile cap. G9 test 8 exercises the cap.

### R2. Save-during-write data loss

**Probability:** Low.
**Impact:** Corrupted `.raw` or `landmarks.dat`.
**Mitigation:** Atomic tmp+rename pattern (G2, G8). G2 test 5 simulates partial-write failure.

### R3. `SaveMap` on a partially-reloaded grid writes inconsistent state

**Probability:** Low — `reload_map` is synchronous and single-goroutine.
**Impact:** Builders see "unexpected" data post-reload.
**Mitigation:** Q1 resolution (purge undo on reload). Integration test in G10.

### R4. `setmark desc` editor leaves the descriptor in `CON_EDITING` if `EditorSave` is never invoked (e.g. builder `/a` abort)

**Probability:** Medium. Builder types `/a` (abort) to cancel.
**Impact:** Landmark description is unchanged (correct); descriptor transitions via existing abort path (Tier 12 handles it).
**Mitigation:** G3 test 11 + 12 cover both save and abort. Add an abort-path test in G3.

### R5. `SectorType` vs `SECT_EXIT` conflicts on the grid after `setexit delete`

**Probability:** Medium.
**Impact:** If `prevsector` was never set (legacy file), restore paints `SECT_OCEAN` — builder's old terrain is gone.
**Mitigation:** G4 test 6 uses a seeded `prevsector = SECT_FOREST`; also add a regression test for `prevsector == SECT_OCEAN` default restore (matches C `:1198`).

### R6. `DoMreset add mobile <name>` name collision ambiguity

**Probability:** Medium — many NPCs share keywords.
**Impact:** Wrong mob is scheduled to spawn.
**Mitigation:** Q3 helper; on multi-match, take the first per C's linear-scan semantics. Builder can always fall back to vnum.

### R7. `mapedit reload confirm` races with in-flight floodfill

**Probability:** Zero (single goroutine).
**Impact:** N/A.
**Mitigation:** Document the invariant.

### R8. Writer produces trailing-whitespace diff with stock files

**Probability:** Medium.
**Impact:** G2 test 8 fails on byte-for-byte diff even though semantic content is right.
**Mitigation:** Inspect stock `landmarks.dat` byte layout during G2 writing — exact tab vs space choices, trailing newline count. Seed the fixture loader to strip trailing empty lines before comparison; OR write a "normalize" helper applied to both sides of the diff.

### R9. `PCData.SecEdit` zero-value means `SECT_INSIDE` (not `SECT_OCEAN`)

**Probability:** Certain (int zero-value).
**Impact:** First-time mapedit session paints `SECT_INSIDE` until builder explicitly chooses a sector. Matches C's undefined-init behavior if `pcdata` was `calloc`'d (which zeros — so `SECT_INSIDE` too). Preserve.
**Mitigation:** G1 test 1 asserts default explicitly. Document in `DoMapedit` help.

### R10. Removing a landmark invalidates in-flight `survey` pagination

**Probability:** Zero (single-goroutine; no pagination yield).
**Impact:** N/A.
**Mitigation:** None needed.

### R11. Writing `.raw` files while loader is mid-boot (test parallelism)

**Probability:** Low — Go tests are per-package and per-world.
**Impact:** G8 test 1 could race with a parallel-running G3 test.
**Mitigation:** G8 uses `t.TempDir()` for its `dataDir`; no shared path.

### R12. SaveMap on a running world overwrites unedited tiles

**Probability:** Certain — `SaveMap` dumps the whole grid.
**Impact:** If two builders on different continents both call `mapedit save` mid-session, map1 and map2 both re-write atomically; no cross-contamination.
**Mitigation:** None needed (single-continent per call).

---

## Status

Planned 2026-04-18; external adversary review recommended before dispatch.
