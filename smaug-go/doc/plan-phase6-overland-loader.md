# Phase 6 — Overland Loader

**Status:** Planned 2026-04-18; external adversary review recommended before dispatch.
**Priority:** Wave-1 / Large.
**Scope:** Read-side only — port the overland map/landmark/entrance loader, sector lookup table, coordinate utilities, and the three player-facing display commands (`coords` / `survey` / `landmarks`) plus the core `display_map` renderer.
**Splits with:** `plan-phase6-overland-builder.md` (all admin mutation commands — blocked on this plan landing).

Mutation verification for every task group in this plan uses `Edit` round-trips only. Banned project-wide: `git checkout` / `git restore` / `git reset --hard` / `git stash`.

---

## Problem

SMAUG ships three 1000x1000-tile overland maps with landmark / entrance / mapreset metadata. In C these are loaded at boot by `load_maps` (`src/overland.c:2006`) and drive `display_map` (`src/overland.c:2197`), `do_coords` (`:3131`), `do_survey` (`:839`), `do_landmarks` (`:936`). A player whose room is `ROOM_MAP`-flagged gets an ANSI tile render centered on their `(X,Y)` coordinates each time they look or move.

The Go port defines `CharData.X/Y/Map/Sector` (`internal/types/character.go:226-229`) and `ObjData.X/Y/Map` (`internal/types/object.go:93-95`) but nothing populates them and nothing renders the map. The `MapData` stub on `RoomIndexData` (`internal/types/room.go:25`, `:105-114`) has only two fields (`Vnum` + unused placeholder) and does not carry the per-room vnum-grid that C stores. The overland system is effectively absent; players who follow an entrance exit into an area keyed to `ROOM_VNUM_OVERLAND_MAP1 = 30000` see a normal room description with no map overlay.

Stock data is intact at `db/maps/`:

- `map1.raw`, `map2.raw`, `map3.raw` — 3 000 000 bytes each (1000 x 1000 x 3-byte RGB tiles).
- `landmarks.dat` — 1 landmark (pitch-black monolith at `map 0, 499, 498, distance 25`).
- `entrances.dat` — 10 entrances linking the three continents plus two regular-zone exits (target vnums `51499` and `50999` and `50499`).
- `mapresets.dat` — empty (`#END` only); the stock areas do not seed the overland with mobs.

This plan stands up the read path from those files to `coords` / `survey` / `landmarks` working for a logged-in player.

---

## C Reference (authoritative)

Line cites are against `src/overland.c` unless noted. `src/overland.h` owns the headers; `src/mud.h` owns `MAP_DATA`.

### Data structures

- `struct map_data` — `mud.h:6471-6477`. Per-room overland binding: `vnum`, `x`, `y`, `entry`. A `RoomIndexData` that represents an overland-map "bucket room" (one of `OVERLAND_MAP1=30000` / `OVERLAND_MAP2=30100` / `OVERLAND_MAP3=30200`) holds a `MAP_DATA *map` (`mud.h:3482`). Regular (non-map) rooms leave it NULL.
- `struct map_index_data` — `mud.h:6480-6486`. Keyed by map vnum, stores `map_of_vnums[49][81]` — a small display grid. **Audit correction 2026-04-19:** original claim "not used by stock overland code path" was wrong — `db.c:2463-2500` builds the MAP_INDEX_DATA linked list during area load, `db.c:6136` queries via `get_map_index`, and `mapout.c:78-141` constructs more entries. Accurate scope note: **not yet exercised by any Go-port code path**. This plan ignores it; defer to a future follow-up when a Go feature needs the index lookup.
- `struct mapreset_data` — `overland.h:68-77`. Fields: `next`, `prev`, `type` (TYPE_OBJECT/TYPE_MOBILE), `vnum`, `map` (short), `x`, `y`.
- `struct landmark_data` — `overland.h:79-89`. Fields: `next`, `prev`, `description` (tilde string), `distance` (int), `map`, `x`, `y`, `Isdesc` (bool — if true the landmark replaces the sector blurb in `display_map`).
- `struct entrance_data` — `overland.h:91-104`. Fields: `next`, `prev`, `area` (tilde string), `vnum`, `herex`, `herey`, `therex`, `therey`, `tomap`, `onmap`, `prevsector`.
- `struct sect_color_type` — `overland.h:106-117`. 9 fields per sector: sector index, ANSI color code, one-char symbol, description, `canpass` bool, move cost, graph1/graph2/graph3 RGB.

### Top-level boot sequence

- `load_maps( void )` — `:2006-2044`. Initializes `map_sector[MAP_MAX][MAX_X][MAX_Y]` to `SECT_OCEAN`, calls `load_mapfile` three times, then `load_entrances`, then `load_landmarks` (DRAGONFLIGHT gate skipped — we don't ship dragonflight). **Does NOT call `load_mapresets`** — the comment at `:2040-2042` explicitly says "Do not put the mapreset loading command here! It won't work at this stage!" Mapresets are loaded later, after `mob_index`/`obj_index` hashes are populated. In Go we load mapresets AFTER `LoadAreas` in boot order.
- `load_mapfile( char *mapfile, short mapnumber )` — `:1961-2003`. For y=0..MAX_Y-1, x=0..MAX_X-1 (outer y loop, inner x loop — note the order is reversed from natural row-major), read 3 bytes (`graph1`/`graph2`/`graph3`), linearly scan `sect_show[]` for a matching `{graph1,graph2,graph3}` triple; if none matches, store `SECT_OCEAN`. Uses `fopen(..., "r")` on a binary file — technically non-portable but works on Linux because the file is FD-level opaque; Go opens with `os.Open`.
- `putterr( short map, short x, short y, short terr )` — `:475-479`. Trivial `map_sector[map][x][y] = terr`.
- `get_terrain( short map, short x, short y )` — `:485-498`. Bounds-check `(x,y)` against `[0, MAX_X/MAX_Y)` and `map != -1`; return `map_sector[map][x][y]` or `-1` on out-of-bounds.

### Tables

- `map_filenames[]` — `:53-55`: `{"map1.raw", "map2.raw", "map3.raw"}`.
- `map_names[]` — `:57-59`: `{"Map 1", "Map 2", "Map 3"}`.
- `map_name[]` — `:61-63`: `{"map1", "map2", "map3"}`.
- `continents[]` — `:65-67`: `{"continent1", "continent2", "continent3"}`.
- `sect_types[]` — `:73-83`: the NAME table used by `get_sectypes`. **34 entries** (plus the optional 35th `"landing"` for DRAGONFLIGHT).
- `impass_message[SECT_MAX]` — `:88-126`: the blurb printed under `display_map` for a tile (both the "you're on X" positive path and the "can't go there" negative path).
- `sect_show[]` — `:137-179`: the 34-entry (35 with dragonflight) sector-color table. **Critical** — the loader reads RGB triples from `.raw` files and must match against this exact table. Row-by-row RGB mapping is in §Raw-file format below.
- `landmark_distances[]` — `:182-194`: 11 strings indexed by `iMes` 0..10 in `do_survey`.
- `random_mobs[SECT_MAX][25]` — `:199-472`: wander-mob vnum table used by `check_random_mobs`. **Out of scope for loader plan** — wander-mob spawn is a runtime behavior tied to `MoveChar`/`map_wander`; this plan ships the table as a load-time constant but does not wire `check_random_mobs`.

### Geometry helpers

- `distance( short chX, short chY, short lmX, short lmY )` — `:507-523`. Returns `sqrt( ((chX-lmX)^2 * (5.12/10.78)) + (chY-lmY)^2 )`. The font-aspect correction `5.12/10.78` makes the display radius look circular on a typical terminal font.
- `calc_angle( short chX, short chY, short lmX, short lmY, double *ipDistan )` — `:526-578`. Returns a 0..360 angle (degrees) from character to landmark, or `-1` if the two coincide. Uses atan on the opposite/adjacent leg ratio, then quadrant-corrects. Emits distance through the `ipDistan` out-param.
- `is_same_map( ch, victim )` — `:583-589`. Compare `map/x/y` equality.
- `fix_maps( ch, victim )` — `:599-642`. Copies `ch->map/x/y` + `PLR_ONMAP`/`ACT_ONMAP` flag onto `victim`.

### Landmark I/O

- `fread_landmark( LANDMARK_DATA *, FILE * )` — `:657-703`. Per-entry keys: `Coordinates` (reads 4 numbers: map, x, y, distance), `Description` (tilde string), `Isdesc` (int-as-bool). Terminator: `End`.
- `load_landmarks()` — `:705-757`. File layout: repeated `#LANDMARK\n<fields>\nEnd\n\n` blocks, terminated by `#END`. `*` prefix lines are comments. Missing file is non-fatal.
- `save_landmarks()` — `:759-787`. Emits the same layout (writes are for the builder plan).
- `check_landmark( map, x, y )` — `:789-802`. Linear scan for `map && x && y`-match; returns pointer or NULL.

### Entrance I/O

- `fread_entrance` — `:1096-1161`. Keys: `Area` (tilde), `Here` (2 numbers), `There` (2 numbers), `ToMap` (int), `OnMap` (int), `Vnum` (int), `Prevsector` (int). Terminator: `End`.
- `load_entrances()` — `:1163-1215`. Sets `enter->prevsector = SECT_OCEAN` before reading (legacy default). Missing file is non-fatal.
- `save_entrances()` — `:1217-1249`. Builder plan's concern.
- `check_entrance( map, x, y )` — `:1251-1264`. Linear scan matching `onmap && herex && herey`.

### Mapreset I/O

- `fread_mapreset` — `:1565-1610`. Keys: `Type` (int), `Vnum` (int), `Coordinates` (3 numbers: map, x, y). Terminator: `End`.
- `load_mapresets()` — `:1612-1669`. After loading, **immediately calls** `mapreset()` to spawn the entities. Must run AFTER mob/obj hashes populated — in Go, AFTER `persist.LoadAreas`.
- `mapreset()` — `:1488-1563`. For each reset, find the overland bucket room for the continent (`OVERLAND_MAP1`/2/3), create the obj or mob, attach it to that room, and stamp `map/x/y` on the entity plus `ITEM_ONMAP`/`ACT_ONMAP`. **Note C bug at `:1514-1518`**: the `case MAP_C3:` lacks a `break;`, so it falls through to `default:` which calls `bug()` and `continue`s — map-3 resets are silently dropped. Preserve or fix per §Open Questions.

### Character-facing commands

- `do_coords( ch, argument )` — `:3131-3182`. Immortal-only (`IS_IMMORTAL` check is implicit — the `PLR_ONMAP` guard + coordinate validation is enough; in practice it's the command-table permission that gates). Usage: `coords <x> <y>`. Validates bounds, sets `ch->x`, `ch->y`, optionally mount coords, then `do_look(ch, "auto")`. Reject if `!PLR_ONMAP`.
- `do_survey( ch, argument )` — `:839-933`. Iterates `first_landmark`, filters on `landmark->map == ch->map` and `!Isdesc`, computes `distance` and `calc_angle`, buckets angle into 8 compass directions, buckets distance into 11 iMes bands, prints "To the <dir>, <distance-msg>, <description>." for each visible landmark. Prints "Your survey of the area yields nothing special." if no landmark is in range. Immortals get bonus distance + coord output.
- `do_landmarks( ch, argument )` — `:936-955`. Tabular dump of every landmark to the pager. No filtering. Columns: Continent | Coordinates | Distance | Description.

### Map display

- `new_map_to_char( ch, startx, starty, endx, endy, radius )` — `:2064-2192`. The inner renderer. For each tile in the bounding box: compute `distance(ch.x, ch.y, x, y)` and black-out tiles beyond `radius` unless `PLR_HOLYLIGHT`. Scan `ch->in_room->first_person` to detect other PCs/NPCs/group members at `(x,y)`. Scan `ch->in_room->first_content` to detect objects at `(x,y)`. Overlay sprite rules in priority order: `@` (`&R` self / `&Y` self in group / `&B` NPC / `&P` other PC) > `$` (`&Y` object) > sector symbol. Sector color is only re-emitted when the sector changes (`lastsector != sector`) — this is the main performance optimization C comments at `:2047-2063` crow about.
- `display_map( ch )` — `:2197-2315`. Computes `(startx,starty,endx,endy)` based on `PLR_HOLYLIGHT` (full 74x28) or mortal (14-wide x 28-tall modulated by time-of-day and precip); applies mapedit sector substitution (`pcdata->secedit`); calls `new_map_to_char`; prints the continent banner, the sector blurb (landmark description if landmark `Isdesc` else `impass_message[sector]`); appends the immortal-only readout of sector name, coordinates, visible reset vnums, and `PLR_MAPEDIT` state.

### Movement / scene integration (OUT OF SCOPE for loader plan but documented for completeness)

- `process_exit( ch, map, x, y, dir )` — `:2585-3010`. The overland movement engine. Translates a direction press into `(x±1, y±1)` and handles wall/water/exit/sector transitions. **This plan does NOT port it**; the loader deliverable does not depend on overland movement. Players can be teleported in with `goto 30000` or `enter_map` and use `coords` to move.
- `find_continent`, `enter_map`, `leave_map`, `collect_followers` — `:3013-3128`. Entry/exit mechanics for the three overland bucket rooms. Out of scope for the loader (but `enter_map` is referenced by `display_map`'s invalid-map recovery at `:2211-2217`; we stub with a logged error + teleport to `ROOM_VNUM_TEMPLE` for now).

### DRAGONFLIGHT guards

`src/overland.c` has `#ifdef DRAGONFLIGHT` guards at `:48-51`, `:79-82`, `:123-125`, `:173-178`, `:2034-2037`, `:2201-2203`, `:2263-2291`. The Go port does not ship dragonflight. The loader mirrors the **non-dragonflight** view: 34 sectors, no `SECT_LANDING`, no `LANDING_DATA`, no landing-site file.

### MAP_DIR / filenames

- `MAP_DIR` is a `#define` in `mud.h` pointing to `"../db/maps/"` (run-relative). C concatenates `MAP_DIR + filename`.
- Go: use `filepath.Join(world.DataDir, "maps", filename)` per the precedent in `boot.go:243-345`.

---

## Go Current State (gap analysis)

### Present

- `CharData.X / Y / Map / Sector int` — `internal/types/character.go:226-229`. **Defined but no writer.** No boot code ever populates these.
- `ObjData.X / Y / Map int` — `internal/types/object.go:93-95`. Same — defined, unused.
- `RoomIndexData.MapData *MapData` — `internal/types/room.go:25`, plus the stub type at `:105-114` with a single field `Vnum int`. **Must be rewritten** to carry per-room `{vnum, x, y, entry}` matching C `MAP_DATA`.
- `SECT_*` enum — `internal/types/enums.go:549-567`. **Only 16 values** (`SECT_INSIDE` through `SECT_SWAMP` + `SECT_MAX`). This is the **non-`OVERLANDCODE`** C enum from `mud.h:2430-2434`, NOT the 34-value overland enum from `mud.h:2414-2426`. **Load will fail** without extending this enum — `sect_show[]` would index past `SECT_MAX`.
- `PLR_HOLYLIGHT` — `internal/types/enums.go:960`. Present.
- `persist.Scanner` — `internal/persist/scanner.go`. Generic line-oriented tilde-string parser. Works for `landmarks.dat` / `entrances.dat` / `mapresets.dat` but NOT for `.raw` (binary). New binary reader needed.
- `world.World.DataDir string` — `internal/world/world.go:76` + `:85`. The join root for the maps directory.
- `boot.Boot` — `internal/boot/boot.go:243-345`. Post-area hookup window. Overland load wires in here.

### Missing

- No `SECT_RIVER / SECT_JUNGLE / SECT_TUNDRA / SECT_ICE / SECT_OCEAN / SECT_SHORE / SECT_TREE / SECT_STONE / SECT_QUICKSAND / SECT_WALL / SECT_GLACIER / SECT_EXIT / SECT_TRAIL / SECT_BLANDS / SECT_GRASSLAND / SECT_SCRUB / SECT_BARREN / SECT_BRIDGE / SECT_ROAD`. Go has 16 sectors; C overland has 34.
- No `PLR_ONMAP / PLR_MAPEDIT` — absent from `enums.go`; add in G1. **Audit correction 2026-04-19:** `ACT_SENTINEL` (`constants.go:445`) and `ACT_ONMAP` (`constants.go:488`) are already defined — do not re-add. Original draft incorrectly listed both as missing. `PLR_ONMAP` and `PLR_MAPEDIT` remain genuinely missing.
- No `ITEM_ONMAP` — required for mapreset object stamping.
- No `ROOM_MAP` room flag — referenced by `atmob` / `atobj` (out of scope) and by the test harness for "is this the overland room?" check. Add.
- No `MAP_DIR` path seam, no `map_filenames[]` / `map_names[]` / `map_name[]` / `continents[]` tables.
- No `SectorInfo` table (Go name for `sect_show[]`).
- No `LandmarkData / EntranceData / MapResetData` types.
- No binary sector grid — equivalent of C `map_sector[MAP_MAX][MAX_X][MAX_Y]` (3 MB static array).
- No loaders (`persist.LoadLandmarks / LoadEntrances / LoadMapresets / LoadMapSectorGrid`).
- No `distance / calcAngle / getTerrain / putTerr / isSameMap / fixMaps` helpers.
- No `DoCoords / DoSurvey / DoLandmarks / DisplayMap` commands.
- No overland package — the port's convention (per `plan.md:152`) is `internal/overland/`.
- No command registrations in `boot.go` for `coords` / `survey` / `landmarks`.

---

## Go Design

### Package layout

```
internal/
  overland/
    sector.go         // SectorInfo table + constants; Sector/RGB helpers
    grid.go           // MapSectorGrid type (3 x 1000 x 1000 uint8) + getTerrain/putTerr
    geometry.go       // Distance/CalcAngle helpers
    world.go          // OverlandState struct + accessor helpers
    display.go        // DisplayMap + newMapToChar
    commands.go       // DoCoords / DoSurvey / DoLandmarks
  persist/
    overland.go       // LoadLandmarks / LoadEntrances / LoadMapresets / LoadMapSector
    overland_test.go
  types/
    room.go           // MapData rewrite: {Vnum int; X, Y int; Entry byte}
    character.go      // existing X/Y/Map/Sector fields are kept as-is
    overland.go       // LandmarkData / EntranceData / MapResetData struct defs
    constants.go      // MAX_X=1000, MAX_Y=1000, MAP_MAX=3, OVERLAND_MAP1/2/3
    enums.go          // extend SECT_* to 34 overland values (see §Sector enum)
```

Rationale for a new `internal/overland/` package (not folding into `act/` or `world/`):

- Clear boundary between "map-as-data" (`persist/`) and "map-as-gameplay" (`overland/`).
- Matches `plan.md:152` convention (which already reserves the name).
- Lets the builder plan (`plan-phase6-overland-builder.md`) add writers + mutation commands without touching the loader surface.
- Avoids `act` bloat — `act/commands.go` is already 1k+ LOC.

### Sector enum strategy (DECISION)

**Chosen:** extend the Go `SECT_*` enum in-place to the 34-value overland version.

Rejected alternative: parallel `overland.SectorType`. Would force every caller (`AreaData` sector-based movement, weather code, object drop rules) to convert between two enums.

The current Go Order is:
```
SECT_INSIDE, SECT_CITY, SECT_FIELD, SECT_FOREST, SECT_HILLS, SECT_MOUNTAIN,
SECT_WATER_SWIM, SECT_WATER_NOSWIM, SECT_UNDERWATER, SECT_AIR, SECT_DESERT,
SECT_DUNNO, SECT_OCEANFLOOR, SECT_UNDERGROUND, SECT_LAVA, SECT_SWAMP, SECT_MAX
```

**Note the mismatch with C-overland enum:** C overland has `SECT_AIR` at index 8 and `SECT_UNDERWATER` at 9 (swapped vs Go); `SECT_DUNNO` is at position 34 (last non-DRAGONFLIGHT) not 11; `SECT_RIVER` is at 11 in overland but absent in Go. Stock `.raw` data was generated against the **overland enum order** — indices in `.raw` are loaded by RGB match, not by index, so ordering is only observable through `get_sectypes` name lookup and through area files that store raw int sector values.

**Migration plan:**

1. Re-order the Go enum to match C overland: `SECT_INSIDE, SECT_CITY, SECT_FIELD, SECT_FOREST, SECT_HILLS, SECT_MOUNTAIN, SECT_WATER_SWIM, SECT_WATER_NOSWIM, SECT_AIR, SECT_UNDERWATER, SECT_DESERT, SECT_RIVER, SECT_OCEANFLOOR, SECT_UNDERGROUND, SECT_JUNGLE, SECT_SWAMP, SECT_TUNDRA, SECT_ICE, SECT_OCEAN, SECT_LAVA, SECT_SHORE, SECT_TREE, SECT_STONE, SECT_QUICKSAND, SECT_WALL, SECT_GLACIER, SECT_EXIT, SECT_TRAIL, SECT_BLANDS, SECT_GRASSLAND, SECT_SCRUB, SECT_BARREN, SECT_BRIDGE, SECT_ROAD, SECT_DUNNO, SECT_MAX`.
2. Audit every use-site of `SECT_SWAMP` / `SECT_LAVA` — these moved indices, so any area file or persistence layer that stores raw ints must be verified to round-trip.
3. `.are` files (stock SMAUG) store sector by integer. Old Go order had `SECT_LAVA=14`, new overland order puts `SECT_LAVA=19`. **Every stock area file's existing rooms will re-interpret their sector**. Mitigation in `persist/area.go`: build a compatibility map `oldGoIndex → overlandIndex` and translate on load for any area file that doesn't contain an overland-era sector. Policy decision flagged in §Open Questions Q3.

### Raw-file format (.raw)

Per `sect_show[]` at `overland.c:137-179`, each tile is 3 bytes RGB. For a 1000x1000 file that's exactly 3 000 000 bytes (matches stock on disk). The RGB triples in use:

| Sector (overland order) | RGB | Symbol | Color |
|---|---|---|---|
| INSIDE | 0,0,0 | ` ` | &x |
| CITY | 255,128,64 | `:` | &Y |
| FIELD | 141,215,1 | `+` | &G |
| FOREST | 0,108,47 | `+` | &g |
| HILLS | 140,102,54 | `^` | &O |
| MOUNTAIN | 152,152,152 | `^` | &w |
| WATER_SWIM | 89,242,251 | `~` | &C |
| WATER_NOSWIM | 67,114,251 | `~` | &B |
| AIR | 0,0,0 | `?` | &x |
| UNDERWATER | 0,0,0 | `?` | &x |
| DESERT | 241,228,145 | `~` | &Y |
| RIVER | 0,0,255 | `~` | &B |
| OCEANFLOOR | 0,0,0 | `?` | &x |
| UNDERGROUND | 0,0,0 | `?` | &x |
| JUNGLE | 70,149,52 | `*` | &g |
| SWAMP | 218,176,56 | `~` | &g |
| TUNDRA | 54,255,255 | `-` | &C |
| ICE | 133,177,252 | `=` | &W |
| OCEAN | 0,0,128 | `~` | &b |
| LAVA | 245,37,29 | `:` | &R |
| SHORE | 255,255,0 | `.` | &Y |
| TREE | 0,64,0 | `^` | &g |
| STONE | 128,128,128 | `^` | &W |
| QUICKSAND | 128,128,0 | `%` | &g |
| WALL | 255,0,255 | `I` | &P |
| GLACIER | 141,207,244 | `=` | &W |
| EXIT | 255,255,255 | `#` | &W |
| TRAIL | 128,64,0 | `:` | &O |
| BLANDS | 128,0,0 | `.` | &r |
| GRASSLAND | 83,202,2 | `.` | &G |
| SCRUB | 123,197,112 | `.` | &g |
| BARREN | 192,192,192 | `.` | &O |
| BRIDGE | 255,0,128 | `:` | &P |
| ROAD | 215,107,0 | `:` | &Y |

**Degenerate triples:** `(0,0,0)` is used by INSIDE, AIR, UNDERWATER, OCEANFLOOR, UNDERGROUND. Stock `.raw` files contain INSIDE rarely (just hand-placed rooms). The C loader picks the **first** match (INSIDE) on `(0,0,0)` — preserve this behavior. Build the lookup as a linear scan in enum order, not a map keyed on RGB.

**Loop order:** C uses outer-y/inner-x (`overland.c:1980-1998`). That means disk layout is row-major in y, with x the fast axis within a row. Matches typical screen raster order. Go: preserve byte-for-byte by reading in the same order.

### In-memory grid

```go
// types/overland.go (new file)
const (
    MAP_MAX = 3
    MAX_X   = 1000
    MAX_Y   = 1000
)

// MapSectorGrid holds the 3-continent sector grid. Indexed by [map][x][y].
// Values are Go SECT_* constants (see enums.go). Outside the grid,
// GetTerrain returns -1.
type MapSectorGrid struct {
    cells [MAP_MAX][MAX_X][MAX_Y]uint8
}
```

9 MB of zero-initialized memory at startup — equivalent to C's static array. Go GC does not move it; we can safely take long-lived pointers to elements if needed, but the plan uses getter/setter only.

### Doubly-linked lists → slices

C uses intrusive `first_X / last_X` + `next/prev` pointers. Go uses slices per `plan.md:60`:

```go
// types/overland.go
type LandmarkData struct {
    Map         int
    X, Y        int
    Distance    int
    Description string
    IsDesc      bool
}

type EntranceData struct {
    Area       string
    Vnum       int
    HereX, HereY   int
    ThereX, ThereY int
    ToMap      int
    OnMap      int
    PrevSector int
}

type MapResetData struct {
    Type int // TYPE_OBJECT=0 / TYPE_MOBILE=1
    Vnum int
    Map  int
    X, Y int
}
```

Owned by `world.World`:

```go
// world/world.go additions
type World struct {
    ...existing...
    Landmarks  []*LandmarkData
    Entrances  []*EntranceData
    MapResets  []*MapResetData
    MapSectors *overland.MapSectorGrid
}
```

Package-level `overland.WorldRef *world.World` set at boot, matching the `act.WorldRef` precedent (per `plan.md:134` ­­convention).

### MapData rewrite

```go
// types/room.go — replace existing stub at :105-114
type MapData struct {
    Vnum  int  // Which overland map this room bucket belongs to (OVERLAND_MAP1/2/3)
    X, Y  int  // Bucket coord (fixed at 499,499 for continent-center rooms)
    Entry byte // Display char (from C 'entry' field; not consumed by current display path)
}
```

`RoomIndexData.MapData` stays `*MapData`. Non-overland rooms have `nil`.

### Commands

```go
// overland/commands.go
func DoCoords(ch *types.CharData, argument string)     // C do_coords :3131
func DoSurvey(ch *types.CharData, argument string)     // C do_survey :839
func DoLandmarks(ch *types.CharData, argument string)  // C do_landmarks :936
```

Registered in `boot.go` Level 0 via `command.Register` (same pattern as `DoLook`).

`DoCoords` gate: reject if `IsNPC(ch)` or if `!ch.Act[PLR_ONMAP]`. Bounds-check `0 <= x < MAX_X` and same for y. Write `ch.X = x, ch.Y = y`; if `ch.Mount != nil` update mount too. Then call `DoLook(ch, "auto")`. **In this plan mount update is a no-op** because Phase 2 mount is wired but mount overland positioning is untested — flag R5.

### Display renderer

`overland.DisplayMap(ch *types.CharData)` ports `display_map` (`:2197`). Signature matches C 1:1. Called from a new hook in `act.DoLook` when the room has `RoomFlags[ROOM_MAP]`. Renderer uses `util.Act`-style `SendToChar` writes for the row-by-row emission — but the **color optimization** (only re-emit color when sector changes) requires direct output-buffer writes, not `act.Act`. Use `ch.SendToChar(string)` directly.

Default mortal window is 75-wide x 29-tall. Holylight gets 75x29 full. Mortal mod shrinks at sunset/sunrise (mod=4), night (mod=2), precip >1 weath_unit (mod -=1), and grows if holding a lit light at night (mod +=1). Port verbatim.

### `coords` flow without `process_exit`

This plan leaves `process_exit` UNPORTED — overland movement by `north`/`south`/etc. is deferred. Players move only via `coords <x> <y>` and via teleport / `goto`. Display is fresh on every `coords` (which calls `DoLook(ch, "auto")`). This limits the playable surface but meets all stated acceptance criteria.

### Boot wiring

In `boot.go` after `LoadAreas` (line 243) and before `LoadSkills`:

1. `LoadMapSectors(w, world.DataDir)` — loads the three `.raw` files.
2. `LoadEntrances(w, world.DataDir)` — parses `entrances.dat`.
3. `LoadLandmarks(w, world.DataDir)` — parses `landmarks.dat`.
4. After all other loaders (post-mob-index): `LoadMapResets(w, world.DataDir)` + `ApplyMapResets(w)`.
5. `overland.WorldRef = w`.
6. Register `DoCoords` / `DoSurvey` / `DoLandmarks` in the command table.

### Error handling

Missing `.raw` files: log via `util.Bug(...)` and **continue** with an all-ocean grid for that map (lets tests run without data). This differs from C which calls `shutdown_mud` + `exit(1)` at `:1975-1978` — the Go port prefers graceful degradation consistent with other loaders (e.g. `persist/area.go` `Bug()` on malformed records).

---

## Task Groups (test-first; mutation-verify via `Edit` round-trips only)

### G1 — Sector enum extension + sector table

**Deliverable:** Go `SECT_*` values match the C overland enum (34 + SECT_MAX). Add a `SectorInfo` struct + `overland.SectorTable[]` matching `sect_show[]`.

**Files to touch:**
- `internal/types/enums.go:549-567` — re-order + add 18 missing sectors.
- `internal/types/constants.go` — add `MAP_MAX = 3`, `MAX_X = 1000`, `MAX_Y = 1000`, `OVERLAND_MAP1 = 30000`, `OVERLAND_MAP2 = 30100`, `OVERLAND_MAP3 = 30200`.
- `internal/overland/sector.go` NEW — `type SectorInfo struct { Sector int; Color, Symbol, Desc string; CanPass bool; Move int; R, G, B byte }` + `var SectorTable [SECT_MAX]SectorInfo` populated from `overland.c:137-179`.
- `internal/overland/sector_test.go` NEW.

**Tests (first):**
1. `TestSectorTableLength` — `len(SectorTable) == SECT_MAX`, and `SECT_MAX == 34` (exactly).
2. `TestSectorTableIndicesMatchEnum` — for each row, `SectorTable[i].Sector == i`.
3. `TestSectorTableRGBMatchesC` — spot-check 6 rows: OCEAN (0,0,128), EXIT (255,255,255), GRASSLAND (83,202,2), LAVA (245,37,29), FOREST (0,108,47), BARREN (192,192,192). Values from `overland.c:137-179`.
4. `TestSectorEnumConstants` — `SECT_OCEAN == 18 && SECT_EXIT == 26 && SECT_ROAD == 33 && SECT_INSIDE == 0`. Mechanical check that the re-order landed.
5. `TestLegacyAreaFilesStillLoad` — load `db/area/airplane.are` and assert no panic, no `Bug` output, and at least one room's `SectorType` is within `[0, SECT_MAX)`. Catches index re-ordering damage to existing area files.

**Mutation verify:**
- Change `SECT_OCEAN` index in enums.go via `Edit` → test 4 fails → revert via `Edit`.
- Comment out `SECT_ROAD` row in `SectorTable` → test 1 fails → revert.

### G2 — Types: Landmark / Entrance / MapReset + MapData rewrite

**Deliverable:** `internal/types/overland.go` defines the three slice-payload structs. `MapData` in `room.go` is rewritten to `{Vnum, X, Y int; Entry byte}`. `World.Landmarks/Entrances/MapResets/MapSectors` slots added.

**Files to touch:**
- `internal/types/overland.go` NEW.
- `internal/types/room.go:105-114` — replace stub.
- `internal/world/world.go:76` area — add four new fields + constructor init.
- `internal/types/overland_test.go` NEW.

**Tests (first):**
1. `TestLandmarkDataZeroValue` — all fields zero-initialize cleanly.
2. `TestWorldOverlandFieldsInitializedEmpty` — freshly-constructed `World` has `Landmarks == nil`, `MapSectors == nil` (the grid is lazy — built by the loader).
3. `TestMapDataCarriesVnumXYEntry` — struct has the four named fields.

**Mutation verify:**
- Remove `Entry byte` field from `MapData` → test 3 fails.

### G3 — Binary map loader (`.raw` → `MapSectorGrid`)

**Deliverable:** `overland.NewMapSectorGrid()` + `overland.LoadMapFile(g *MapSectorGrid, mapNum int, path string) error`. All-ocean fallback on missing file. RGB-triple lookup in `SectorTable` order, `SECT_OCEAN` as unmatched default.

**Files to touch:**
- `internal/overland/grid.go` NEW — `MapSectorGrid`, `GetTerrain`, `PutTerr`, `LoadMapFile`.
- `internal/overland/grid_test.go` NEW.
- `internal/persist/overland.go` NEW — `LoadMapSectors(w *World, dataDir string) error` (wraps the three `overland.LoadMapFile` calls).

**Tests (first):**
1. `TestGetTerrainOutOfBoundsReturnsMinusOne` — `(-1, 0, 0)` / `(0, -1, 0)` / `(0, MAX_X, 0)` / `(0, 0, MAX_Y)` → -1.
2. `TestPutTerrGetTerrainRoundTrip` — set then get in a small fixture grid (no real file).
3. `TestLoadMapFileOcean` — generate a synthetic 12-byte `.raw` (2x2 at `(0,0,128)`), load into a 2x2 grid-wrapper helper, assert all 4 cells == `SECT_OCEAN`.
4. `TestLoadMapFileMixed` — synthetic 2x2 with one `(255,255,255)` / one `(141,215,1)` / two `(0,0,128)`, assert `SECT_EXIT` / `SECT_FIELD` / `SECT_OCEAN` in the right positions.
5. `TestLoadMapFileUnknownRGBDefaultsToOcean` — synthetic 1x1 with `(17,17,17)`, assert `SECT_OCEAN`.
6. `TestLoadMapFileMissingReturnsNilWithOceanGrid` — call with a non-existent path, assert nil error, grid all ocean, `util.Bug` was logged (capture via test hook).
7. `TestLoadMapFileDegenerate000MapsToInside` — synthetic 1x1 with `(0,0,0)`, assert `SECT_INSIDE` (first match in SectorTable order).
8. **Integration test** `TestLoadStockMap1Raw` — gated on `db/maps/map1.raw` existing; loads the real file, asserts `(499,499) == SECT_OCEAN` (center of stock empty map) and spot-checks a non-zero tile count (should be > 0 for any real map but could be all-ocean if stock is unpopulated; assert `>= 0`).

**Mutation verify:**
- Change the RGB-match loop to break on FIRST iteration → test 4 fails (FIELD won't be detected → shows as OCEAN).
- Flip the loop order (x outer, y inner) → test 4 fails because tile positions shift.

### G4 — Landmark / entrance / mapreset file parsers

**Deliverable:** `persist.LoadLandmarks / LoadEntrances / LoadMapresets`. Each reads the tilde-string format via `persist.Scanner`. Missing file → no-op (nil error, empty slice), per C.

**Files to touch:**
- `internal/persist/overland.go` — extend from G3 with three more loaders.
- `internal/persist/overland_test.go` NEW.
- `internal/persist/testdata/maps/landmarks.dat` NEW fixture (one `#LANDMARK`, one `#END`).
- `internal/persist/testdata/maps/entrances.dat` NEW (one `#ENTRANCE`, `#END`).
- `internal/persist/testdata/maps/mapresets.dat` NEW (`#END` only).

**Tests (first):**
1. `TestLoadLandmarksFixture` — parses one `#LANDMARK` with coords `0 499 498 25`, description "a pitch black monolith", IsDesc false. Values match `db/maps/landmarks.dat` exactly.
2. `TestLoadLandmarksIsDescSetsTrue` — fixture variant with `Isdesc 1`.
3. `TestLoadLandmarksMissingFileNoError` — pass a nonexistent path, assert nil error, zero landmarks.
4. `TestLoadLandmarksCommentLineSkipped` — fixture with a `*comment` line; parser must skip.
5. `TestLoadEntrancesFixture` — parses one entrance with all 7 keys, `prevsector` defaults to `SECT_OCEAN` when absent (per `:1198`).
6. `TestLoadEntrancesPrevsectorDefault` — fixture without `Prevsector` key → value is `SECT_OCEAN`.
7. `TestLoadMapresetsEmptyFile` — fixture with only `#END`, zero resets.
8. `TestLoadMapresetsFixture` — fixture with one `#RESET` with Type=1 / Vnum=1234 / Coordinates=1 100 200.
9. **Integration** `TestLoadStockEntrancesTenEntries` — load real `db/maps/entrances.dat`, assert exactly 10 entrances.
10. **Integration** `TestLoadStockLandmarksOneEntry` — load real file, assert one landmark with the monolith description.

**Mutation verify:**
- Change the `case 'I':` branch (Isdesc) in the landmark parser to ignore the key → test 2 fails.
- Remove the `Prevsector` default init → test 6 fails.

### G5 — Geometry helpers (`Distance`, `CalcAngle`)

**Deliverable:** `overland.Distance(chx, chy, lmx, lmy int) float64` and `overland.CalcAngle(chx, chy, lmx, lmy int) (angle float64, distance float64)`. Ports `src/overland.c:507-523` and `:526-578` verbatim.

**Files to touch:**
- `internal/overland/geometry.go` NEW.
- `internal/overland/geometry_test.go` NEW.

**Tests (first):**
1. `TestDistanceZero` — `Distance(5, 5, 5, 5) == 0`.
2. `TestDistancePureY` — `Distance(5, 10, 5, 15) == 5.0` (exact; y-only distance skips the font ratio).
3. `TestDistancePureX` — `Distance(0, 0, 10, 0) ≈ sqrt(100 * 5.12 / 10.78)`. Check within 1e-9.
4. `TestDistanceSymmetric` — `Distance(a,b,c,d) == Distance(c,d,a,b)`.
5. `TestCalcAngleSame` — `CalcAngle(5,5,5,5)` returns -1 angle.
6. `TestCalcAngleCardinals` — (0,0)→(0,-10) returns 0 (north), (0,0)→(10,0) returns 90 (east), (0,0)→(0,10) returns 180 (south), (0,0)→(-10,0) returns 270 (west).
7. `TestCalcAngleNE` — (0,0)→(10,-10) is in quadrant NE (45° range); the C path at `:564-565` computes `(90 + (90 - iFinal))` — for a 45-degree diagonal, result is 45.

**Mutation verify:**
- Change the font-ratio constant `5.12/10.78` to `1.0` → test 3 fails.
- Swap quadrant branches in `CalcAngle` → test 6 fails for one cardinal.

### G6 — Command: `DoCoords`

**Deliverable:** `overland.DoCoords` ported from `:3131-3182`. Register in `boot.go`.

**Files to touch:**
- `internal/overland/commands.go` NEW.
- `internal/overland/commands_test.go` NEW.
- `internal/boot/boot.go` — `command.Register("coords", overland.DoCoords, ...)`.

**Tests (first):**
1. `TestDoCoordsNPCRejected` — NPC invoker gets "NPCs cannot use this command." No state change.
2. `TestDoCoordsNotOnMapRejected` — PC without `PLR_ONMAP` → "This command can only be used from the overland maps."
3. `TestDoCoordsUsage` — on-map, empty arg → "Usage: coords <x> <y>".
4. `TestDoCoordsBoundsCheck` — `coords -1 0` / `coords 1000 0` / `coords 0 1000` each rejected with the exact-C "Valid x coordinates are 0 to 999." / similar-for-y.
5. `TestDoCoordsHappyPath` — PLR_ONMAP PC calls `coords 100 200`, assert `ch.X==100`, `ch.Y==200`, and `DoLook(ch, "auto")` fires (capture via a hook or via output-buffer spy).
6. `TestDoCoordsMountUpdated` — if `ch.Mount != nil`, mount X/Y copy. (Plan R5 — if the mount plumbing is too stubbed, defer this test as `t.Skip` with a TODO.)

**Mutation verify:**
- Remove the X bounds check → test 4 fails.
- Skip the `DoLook` call → test 5 fails.

### G7 — Command: `DoLandmarks`

**Deliverable:** `DoLandmarks` ported from `:936-955`. Dumps all landmarks to the player's output buffer. Column widths match C.

**Files to touch:**
- `internal/overland/commands.go`.
- `internal/overland/commands_test.go`.
- `internal/boot/boot.go` — register.

**Tests (first):**
1. `TestDoLandmarksEmpty` — zero landmarks → "No landmarks defined." No header.
2. `TestDoLandmarksSingle` — one landmark → header printed + one row with `Map 1` / `499X` / `498Y` / `25` / monolith description.
3. `TestDoLandmarksColumnAlignment` — two landmarks of different widths render their columns aligned (exact format: `%-10s  %-4dX %-4dY   %-4d       %s`).

**Mutation verify:**
- Swap map-name index to ch.Map (wrong) → test 2 fails (shows the viewer's continent, not the landmark's).

### G8 — Command: `DoSurvey`

**Deliverable:** `DoSurvey` ported from `:839-933`. Iterates landmarks on same map, filters `!IsDesc`, emits directional + distance blurb, extra imm-only info.

**Files to touch:** same as G7.

**Tests (first):**
1. `TestDoSurveyNoLandmarks` — output: "Your survey of the area yields nothing special."
2. `TestDoSurveyLandmarkOutOfRange` — landmark distance=10, player 100 away → nothing special.
3. `TestDoSurveyLandmarkInRange` — distance=50, landmark at (499,498), player at (499,499) → "Right here nearby, a pitch black monolith rises toward the sky."
4. `TestDoSurveyDirectionalBlurb` — landmark due north, dist 20 → "To the north, in the immediate area, <desc>." (iMes=10 since dist <= 1; use a closer landmark or adjust bucket).
5. `TestDoSurveyDistanceBuckets` — seed a landmark at various distances, assert the 11 iMes bands ("hundreds of miles away..." through "in the immediate area").
6. `TestDoSurveyIsDescLandmarkHidden` — landmark with `IsDesc=true` is skipped (it's a room description, not a real landmark).
7. `TestDoSurveyImmortalExtras` — high-trust viewer gets "Distance to landmark: N\r\nLandmark coordinates: Nx Ny" appended.
8. `TestDoSurveyWrongMapHidden` — landmark on map 2, player on map 0 → no output for that landmark.

**Mutation verify:**
- Delete the `!landmark.IsDesc` filter → test 6 fails.
- Change `dist <= landmark.distance` to `dist < landmark.distance` → test 3 fails at exact boundary (need a boundary-fixture to trigger).

### G9 — `DisplayMap` renderer + `newMapToChar`

**Deliverable:** `overland.DisplayMap` ports `:2197-2315`. `newMapToChar` ports `:2064-2192`. Renders a 75x29 (holylight) or mod-modulated (mortal) tile block with sprite overlays + continent banner + sector blurb. No integration with `DoLook` yet — that's G10.

**Files to touch:**
- `internal/overland/display.go` NEW.
- `internal/overland/display_test.go` NEW.

**Tests (first):**
1. `TestDisplayMapInvalidMapRecovers` — player with `Map == -1` gets the "&RYou were found on an invalid map..." blurb and is moved to continent-1 coords (499, 500). Log a `util.Bug`.
2. `TestDisplayMapHappyOcean` — player at (499,499), all-ocean grid, holylight mortal → output contains 29 rows of 75 ocean tiles (each prefixed by `&b`, symbol `~`) plus player sprite `&R@` at (499,499). Exact line count = 29.
3. `TestDisplayMapColorRunOptimization` — adjacent same-sector tiles emit color only once. Observable by counting `&` occurrences in output.
4. `TestDisplayMapPlayerSprite` — player at (100,100) in a 7-radius mortal view → `&R@` at the center; no `&P@` unless a second PC is at a nearby cell.
5. `TestDisplayMapSecondPCSprite` — second PC in the same room at (101,101) → `&P@` at that column.
6. `TestDisplayMapObjectSprite` — object with (X,Y) matching a visible tile → `&Y$`.
7. `TestDisplayMapSectorBlurbLandmarkIsDesc` — if a landmark at ch's exact coords has `IsDesc=true`, its description replaces the `impass_message[sector]`.
8. `TestDisplayMapImmortalSectorLine` — immortal gets "&GSector type: ocean. Coordinates: 499X, 499Y\r\n" appended.
9. `TestDisplayMapImmortalResetLines` — immortal with a mapreset at their coords sees "&POOject reset present: Vnum N" (quoted verbatim from C `:2306-2307`). Note: preserve the typo-free C form "%s reset present: Vnum %d".
10. `TestDisplayMapRadiusBlackout` — tile beyond `radius` (mortal + mod=7) renders as `" "` (space), no color.

**Mutation verify:**
- Delete the `lastsector = -1` reset in the sprite-emit block → test 3 fails (color spam increases).
- Swap x and y in the inner loop → test 2 fails (wrong dimensions).

### G10 — Boot wiring + registration

**Deliverable:** All loaders called from `boot.Boot`. `DoCoords`/`DoSurvey`/`DoLandmarks` registered. `overland.WorldRef` set. An integration test logs in a telnet client, teleports to overland, runs all three commands, and parses their output.

**Files to touch:**
- `internal/boot/boot.go` — loader calls + command registrations + `overland.WorldRef = w`.
- `internal/boot/boot_test.go` — post-boot non-nil assertions for `w.MapSectors`, `len(w.Entrances) == 10`, `len(w.Landmarks) == 1`.
- `internal/testclient/overland_test.go` NEW — end-to-end scenario.

**Tests (first):**
1. `TestBootLoadsMapSectors` — `w.MapSectors != nil` after `Boot(w, ...)`.
2. `TestBootLoadsTenEntrances` — matches stock data.
3. `TestBootLoadsOneLandmark` — matches stock data.
4. `TestBootRegistersCoordsSurveyLandmarks` — `command.Find("coords")` returns non-nil, `command.Find("survey")` too, `command.Find("landmarks")` too.
5. **Integration** `TestOverlandLandmarksCommandEndToEnd` — spin up `testclient.Harness`, log in a PC, force-set `PLR_ONMAP`, `ch.X = 100, ch.Y = 100, ch.Map = 0`, send `landmarks`, assert output contains "Continent | Coordinates" and "Map 1  499X 498Y".
6. **Integration** `TestOverlandCoordsCommandEndToEnd` — same setup, send `coords 200 300`, assert `ch.X == 200` and the client receives the continent banner.
7. **Integration** `TestOverlandSurveyCommandEndToEnd` — place PC at (499, 525) (26 tiles from monolith at 499,498, just inside the visibility-25 threshold — actually 25 is exclusive per `dist <= landmark.distance` → use distance ≤ 25: dist = 25 exactly lands on boundary, use y=523 for safe-in-range). Assert output contains "a pitch black monolith" and "To the north" or similar.

**Mutation verify:**
- Delete the `LoadEntrances` call → test 2 fails.
- Typo `coords` to `coordz` in registration → test 4 fails.

### G11 — Mapresets table load (deferred spawn)

**Deliverable:** `LoadMapResets` populates `w.MapResets`. `ApplyMapResets(w)` spawns the mobs/objects. Run after `LoadAreas` in `boot.go`. Preserves C behavior that stock data is empty; provides a fixture-driven test.

**Files to touch:**
- `internal/persist/overland.go`.
- `internal/overland/reset.go` NEW — `ApplyMapResets(w *World)`.
- `internal/overland/reset_test.go` NEW.
- `internal/boot/boot.go`.

**Tests (first):**
1. `TestApplyMapResetsEmpty` — empty slice, no mobs created.
2. `TestApplyMapResetsMobile` — seed one mobile reset for an existing mob vnum (use one from `db/area/limbo.are`); call `ApplyMapResets`; assert a mob exists in `get_room_index(OVERLAND_MAP1)` with `Act[ACT_ONMAP]` set and `Map/X/Y` matching the reset.
3. `TestApplyMapResetsObject` — seed one object reset; similar assertions with `ExtraFlags[ITEM_ONMAP]`.
4. `TestApplyMapResetsMap2CaseBreakPresent` — seed map-2 reset; assert it spawns. Guards against the C `:1514` missing-`break` bug. **Q1 in Open Questions** — policy is to fix in Go; if preserved, flip test to assert NO spawn.
5. `TestApplyMapResetsBadVnumLogged` — seed a reset with vnum=999999 (not in any index); assert no crash, `util.Bug` logged.

**Mutation verify:**
- Remove the `Act[ACT_ONMAP] = true` set → test 2 fails.
- Remove the Map/X/Y stamp → test 2 fails on position assertion.

### G12 — `DoLook` hook for overland rooms

**Deliverable:** `act.DoLook` detects `ROOM_MAP`-flagged rooms and calls `overland.DisplayMap(ch)` instead of the normal description emitter. Minimal change — existing DoLook path for non-map rooms is untouched.

**Files to touch:**
- `internal/act/info.go:21` (`DoLook`) — add the check.
- `internal/types/enums.go` — add `ROOM_MAP` flag if absent (§Open Questions Q2).
- `internal/act/info_test.go` — new test for the overland branch.

**Tests (first):**
1. `TestDoLookOverlandRoomShowsMap` — PC in a `ROOM_MAP`-flagged room with `PLR_ONMAP` gets the map render, not the plain description.
2. `TestDoLookNonOverlandRoomShowsDescription` — normal room still emits room name / description / exits list (pre-Phase-6 behavior unchanged).
3. `TestDoLookAutoTriggersFullRender` — `DoLook(ch, "auto")` from `DoCoords` drives the map render, confirming G6's integration with G9 through this hook.

**Mutation verify:**
- Invert the `RoomFlags[ROOM_MAP]` branch condition → test 2 fails.

### G13 — Documentation + changelog

**Deliverable:** Update `CLAUDE.md` index table with the new overland plan-doc pointers and the new package. Append one `CHANGELOG.md` entry for the loader landing. Move relevant TODO items to Done.

**Files to touch:**
- `CLAUDE.md` — add `plan-phase6-overland-loader.md` entry.
- `CHANGELOG.md` — append date-stamped line.
- `TODO.md` — move loader follow-ups.

No tests — documentation.

---

## Acceptance Criteria

All criteria mechanically verifiable by `go test ./...` or by a `telnet + interpret` scenario test.

- **A1.** `Boot` completes with `w.MapSectors != nil`, `len(w.Landmarks) == 1`, `len(w.Entrances) == 10`, `len(w.MapResets) == 0` against stock data. (G3 + G4 + G10 + G11)
- **A2.** `SECT_MAX == 34` and `SectorTable[SECT_OCEAN].Color == "&b"`. (G1)
- **A3.** `LoadMapFile` converts stock `db/maps/map1.raw` to an all-or-mostly-ocean grid, with zero panics and zero unresolved-RGB log lines. (G3)
- **A4.** `command.Find("coords")` / `command.Find("survey")` / `command.Find("landmarks")` each return non-nil after boot. (G10)
- **A5.** A telnet-driven PC with `PLR_ONMAP` set can run `coords 499 499` → `landmarks` → `survey` and receive the monolith landmark data in output. (G10 integration test)
- **A6.** `DoCoords` rejects NPCs, non-map players, and out-of-bounds coordinates with C-exact messages. (G6)
- **A7.** `DoSurvey` respects the 11 iMes bucket boundaries as defined at `overland.c:892-913`. (G8)
- **A8.** `DoLandmarks` output row format matches C's `%-10s  %-4dX %-4dY   %-4d       %s`. (G7)
- **A9.** `DisplayMap` emits at most one color code per sector run (optimization preserved). (G9)
- **A10.** `DisplayMap` correctly resolves `IsDesc` landmarks as the sector blurb. (G9)
- **A11.** `DisplayMap` invalid-map recovery path logs a bug and teleports the player to continent-1 coords. (G9)
- **A12.** `LoadLandmarks` / `LoadEntrances` / `LoadMapresets` all return nil error when the file is absent, leaving their world slices empty. (G4)
- **A13.** `Distance(a,b,c,d) == Distance(c,d,a,b)` and `CalcAngle` quadrant output matches `overland.c:564-577` for the 8 cardinal/ordinal directions. (G5)
- **A14.** `ApplyMapResets` spawns mobs with `Act[ACT_ONMAP] == true` and `Map/X/Y` stamped. (G11)
- **A15.** `DoLook` branches to `DisplayMap` for `ROOM_MAP`-flagged rooms and to the legacy description emitter for others. (G12)
- **A16.** `go test -count=3 ./...` is green across all packages after landing.
- **A17.** Stock `.are` files continue to load without sector-re-interpretation errors (guarded by `TestLegacyAreaFilesStillLoad`). (G1)

---

## Scope Cuts / Deferrals (tracked for builder plan or Phase-7 follow-up)

- **No overland movement engine.** `process_exit` (`:2585-3010`) is NOT ported. `north` / `south` / etc. on an overland room fall through to the existing `MoveChar`, which will emit "alas, you cannot go that way" for `ROOM_MAP` rooms without exits. Players move via `coords` only. **Deferred to `plan-phase6-overland-move.md` (future).**
- **No `enter_map` / `leave_map` / `find_continent` / `collect_followers`.** Teleport commands (`goto`, `transfer`) do not drop characters onto the map yet — they land in the overland bucket room with `PLR_ONMAP` unset. A testclient helper sets the flag manually. **Deferred.**
- **No `map_wander` / `check_random_mobs` / `map_scan`.** No wander-mob spawn, no `scan` enhancement. The `random_mobs[][]` table from `:199-472` is NOT ported.
- **No `atmob` / `atobj` map bookkeeping.** The immortal `at` command does not yet round-trip overland state. Already works for non-map rooms.
- **No `mapedit` / `setmark` / `setexit` / `mapresets` / `mreset` / floodfill.** All mutation → `plan-phase6-overland-builder.md`.
- **No DRAGONFLIGHT / landing sites.** Out of Phase 6 entirely.
- **No `save_map`** — builder plan.
- **No `map_index_data` / `ASCII sub-map`** — legacy feature that stock SMAUG data does not exercise. Tracked as open follow-up.
- **No telnet-scale performance test.** `DisplayMap` is hot but the baseline stock map is empty (all-ocean) — real perf work deferred until content lands.
- **No MCCP-aware output-buffer sizing.** Current `SendToChar` buffer is fine for 75x29 ANSI output; keep.

---

## Open Questions

### Q1. Map-3 `mapreset` case-break bug (C `:1514-1518`)

C's `case MAP_C3:` has no `break;`, so it falls through to `default:` and calls `bug()`. **Stock data has zero map-3 resets**, so the bug has not been exercised in production. Policy options:

- **(a) Preserve verbatim** — keep the bug; add a regression test that asserts map-3 resets are dropped.
- **(b) Fix in Go, document in plan** — port with an explicit `break` and a code comment citing the C line.

**Recommended:** (b). The bug is clearly a typo; preserving C-faithfulness here harms building. Add a `// NOTE: C src/overland.c:1514 missing break — fixed here.` inline comment.

### Q2. `ROOM_MAP` flag index

Current `internal/types/enums.go` has no `ROOM_MAP` constant (verify via `Grep`). C defines `ROOM_MAP` as bit `BV30`-ish in `mud.h` (search in G1). **Recommended:** add to enums.go with the next free bit slot AND audit every stock `.are` file to ensure no room currently sets that bit unexpectedly — `grep -l 'ROOM_MAP\b' db/area/*.are` or equivalent. If area files don't already reference this flag, we can pick any bit.

### Q3. Sector enum re-order impact on existing area files

Re-ordering `SECT_*` shifts some integer indices. Old Go order had `SECT_LAVA=14`; new overland order has `SECT_LAVA=19`. Stock `.are` files store integer sector values; a room with `"sector_type 14"` will load as `SECT_JUNGLE` (index 14 in new order) instead of `SECT_LAVA`. **Three options:**

- **(a) Bump on load — translation table.** In `persist/area.go`, detect "legacy-ordered" files (heuristic: no sectors > 15) and map `oldGoIndex → overlandIndex`. Risk: some areas legitimately use sectors 0-15 in overland order too — ambiguity.
- **(b) Area-file version field.** Introduce `#VERSION` hint at the top of `.are` files. Retroactively stamp stock files. Breaks file-format-compatibility claim in `plan.md:96-99`.
- **(c) Re-save all stock area files.** One-shot `persist/area_write.go` bulk re-save after migration. Single migration commit.

**Recommended:** (c), paired with an assertion in G1's `TestLegacyAreaFilesStillLoad` that each stock area's sectors round-trip cleanly after the re-save.

**Human input gate.** The choice affects save-file semantics and should be confirmed before G1 lands.

### Q4. `ACT_SENTINEL` / `PLR_ONMAP` / `ACT_ONMAP` / `ITEM_ONMAP` / `PLR_MAPEDIT` presence

`Grep`-verify each constant in `internal/types/enums.go` before G2. Defaults from `mud.h`:
- `PLR_ONMAP` — set in `mud.h` near line 2600.
- `ACT_ONMAP` — set in `mud.h` near line 2400.
- `PLR_MAPEDIT` — set in `mud.h` around 2650.
- `ITEM_ONMAP` — set in `mud.h` around 2100.

If absent, add them to `enums.go` at the end of each bit-table section in G1 (bundle with sector-enum extension). Recommended: bundle into G1 with a single `enums.go` Edit.

### Q5. Mount update in `DoCoords` (test G6 #6)

`plan-combat-depth.md` (Tier 6) already ships mount basics, but overland mount tracking was not covered. The C path at `:3174-3178` unconditionally copies mount coords. If Go's mount plumbing has gaps (e.g. `ch.Mount` is always nil post-load), mark the test `t.Skip("mount-overland positioning deferred")` and add a TODO. **Recommended:** port the code path; skip the test if `ch.Mount == nil` path coverage is the only concern.

### Q6. Binary-file endianness / encoding

Stock `.raw` files are byte streams of 3-byte RGB triples with no multi-byte-integer fields. Portable across platforms. **No action — document inline in G3.**

### Q7. MAP_DIR path handling on Windows

Go uses `filepath.Join` per existing boot code. Backslash / forward-slash normalization is handled by Go. No platform-specific concern.

---

## Risk Analysis

### R1. Sector enum re-order breaks stock `.are` files

**Probability:** High.
**Impact:** Area load emits wrong sector per room → movement costs, impassability, weather all mispredict.
**Mitigation:** §Q3 option (c) — re-save all stock areas in a single migration commit paired with G1. G1's `TestLegacyAreaFilesStillLoad` integration test gates the landing.

### R2. 9 MB static grid memory hit

**Probability:** Certain.
**Impact:** Baseline RSS +9 MB. Stress-tested fine (Phase 3 was at ~80 MB steady-state).
**Mitigation:** Document in CHANGELOG. Acceptable. If a test harness wants a smaller grid, add a `NewMapSectorGridForTest(mapCount, xDim, yDim)` helper with smaller bounds — tests already use this pattern for areas.

### R3. `display_map` output size exceeds output-buffer limits

**Probability:** Medium.
**Impact:** 75x29 = 2175 tiles; worst case (every tile a different sector) = ~2175 × 4 bytes (color + symbol) ≈ 9 KB per render. `DescriptorData` output buffer is typically 32-64 KB. Room-movement spam (10 moves/sec) could trigger buffer bloat.
**Mitigation:** G9's color-run optimization. Document buffer size in the file header. If flushes stall, escalate in TODO for Phase-7 perf.

### R4. `fopen` / `os.Open` binary-mode Windows differences

**Probability:** Low (project is Linux-targeted per `CLAUDE.md`).
**Impact:** Byte-for-byte raw file reads would be corrupted on Windows if opened in text mode.
**Mitigation:** Use `os.Open` (binary by default in Go). Add explicit comment.

### R5. Mount-overland coords drift

**Probability:** Medium.
**Impact:** `coords` updates `ch.Mount.X/Y` but no other command does — mount can de-sync on teleports, goto, bamf.
**Mitigation:** Out of loader-plan scope; track in TODO for "overland mount sync" follow-up. `DoCoords` mount-update code path covered by G6 test with `t.Skip` fallback.

### R6. Missing `ROOM_MAP` bit collision with existing area flag

**Probability:** Low.
**Impact:** Selecting a room-flag bit that an existing area already uses would cause `DoLook` to mis-route real rooms into `DisplayMap`.
**Mitigation:** §Q2 audit before G1; pick a free bit.

### R7. `SECT_DUNNO` interop

**Probability:** Low-medium.
**Impact:** Overland enum has `SECT_DUNNO` at the end (index 34); stock Go had it at 11. Area files using sector 11 (`DUNNO`) will be re-interpreted as `SECT_RIVER` under the new order.
**Mitigation:** §Q3 resolves this. Also: `grep -c "sector_type 11" db/area/*.are` to quantify blast radius before deciding.

### R8. `display_map` infinite recursion on invalid-map recovery

**Probability:** Low.
**Impact:** `display_map` calls `enter_map` when `ch.map == -1`. `enter_map` calls `do_look(ch, "auto")`. If the target continent room is itself invalid, we could stack-overflow.
**Mitigation:** G9 test 1; add a one-shot guard (a local bool or ch-level flag) that breaks the recursion with a "ERROR" message if re-entry is detected. Cite C's lack of guard as a preserved bug; document the Go fix.

### R9. Scanner tilde handling inconsistent with landmark `Description`

**Probability:** Low.
**Impact:** `fread_string` is tilde-terminated; `persist.Scanner.ReadString` already mirrors this. Empty string edge case (`~` immediately after key) could panic.
**Mitigation:** G4 test 1 exercises a real-description landmark; add a G4 test with empty description (`Description ~`) as a regression.

### R10. `coords` gate on `PLR_ONMAP` vs immortal override

**Probability:** Medium.
**Impact:** C code rejects immortals if not flagged `PLR_ONMAP`. Some imm workflows assume always-on `coords`. Preserve C gate to avoid building muscle memory around a non-faithful behavior, but document the shortcut: `goto <overland-vnum>` → `pedit onmap` → `coords x y`.
**Mitigation:** G6 test 2 preserves the gate. Immortal-shortcut policy noted in CHANGELOG.

### R11. Stock map-3 tiles have unrecognized RGB triples

**Probability:** Low-medium.
**Impact:** Tiles silently become SECT_OCEAN, cascading into display / movement confusion.
**Mitigation:** G3 test 5 covers the fallback. Post-load, emit a one-time log summary: "LoadMapFile map3: N tiles unresolved → ocean." Phase-7 follow-up to investigate any non-trivial count.

### R12. Command-table prefix conflicts with existing commands

**Probability:** Low.
**Impact:** `coords` does not clash with any existing command; `survey` doesn't either; `landmarks` is fresh.
**Mitigation:** G10 test 4 verifies registration via `command.Find(exact-name)`. Also: `grep -i "register.*coords\|register.*survey\|register.*landmarks" internal/boot/boot.go` before landing.

---

## Status

Planned 2026-04-18; external adversary review recommended before dispatch.
