# Plan: Phase 6 Housing — Player Apartments / Houses, `do_house` Family

**Status:** Planned 2026-04-18; external adversary review recommended before dispatch.
**Priority:** P2 (Phase 6, Wave D — Game systems, player-facing).
**Scope:** Full port of SMAUG housing: six player commands (`do_house`, `do_gohome`, `do_residence`, `do_accessories`, `do_homebuy`, `do_sellhouse`), per-room description customization via the `EditorSave` callback, on-disk persistence at `db/houses/<Name>` plus `db/houses/house.lst` + `db/houses/homebuy.dat` + `db/houses/homeaccessories.dat`, room-ownership mapping, automated auction-tick update, immortal admin paths (`house set` / `house remove` / `accessories add` / `accessories remove` / `accessories setprice` / `homebuy add` / `homebuy remove` / `homebuy setbid` / `homebuy bidincrement` / `homebuy timeremainder` / `homebuy clearbidder`). C reference is `src/house.c` 2853 LOC.

---

## Problem

The port currently has one trailing hook (`HomeVnum int` at `internal/types/character.go:209`) but nothing else related to housing. None of the six commands are registered. `ITEM_HOUSEKEY` (`internal/types/enums.go:638`) and `ROOM_HOUSE` (`internal/types/enums.go:699`) are defined but never set by any code path, so the flags are effectively cosmetic. No persistence exists at `db/houses/*` beyond an empty `house.lst` (verified: `db/houses/` contains only `.do_not_delete` and an empty `house.lst`). No `RoomIndexData.OwnedBy` field exists (`grep` of `internal/types/room.go` returns zero hits for `OwnedBy | Residents | HouseVnums`). No `World.Homes` list exists (`grep` of `internal/world/world.go` returns zero hits for `Homes | Housing`).

The C implementation is substantial — 22 top-level functions, three linked lists of data (`first_home`, `first_homebuy`, `first_accessory`), three persistence files, and a per-home directory entry. It is the third-largest self-contained system deferred from Phase 5 after Overland (3752 LOC) and Polymorph (2753 LOC). The command surface is smaller than the LOC count suggests — `do_house` is 411 C lines but dispatches ~7 subcommands; `do_homebuy` is 347 lines but dispatches ~10 subcommands; the remaining lines are plumbing (persistence, `set_house`/`remove_house`/`add_room`/`give_key`, the auction tick).

Housing interacts with two already-landed subsystems that require care:

1. The `EditorSave` callback pattern landed in Tier 12 (`internal/game/editor.go:159-173`) is the mandated path for `house desc`. C's `SUB_ROOM_DESC` substate at `src/house.c:100-113` is *exactly* the same shape: set `ch->substate`, call `start_editing`, on `/s` run callback that writes `room->description` and calls `save_residence`. Housing is the second consumer of the pattern after OLC `redit desc`.
2. `fold_area` — SMAUG's area-writer — is the C persistence path for `save_residence`. It has a Go equivalent at `internal/persist/area_write.go` (Phase 3). `save_residence` writes the *entire* area file back. That's already how Go `redit` persists changes to modified rooms, so this is a no-new-mechanism invocation.

Hotboot is not yet shipped (design-doc only in `plan-phase6-hotboot.md`). Housing introduces significant additional state: one JSON/text file per player, accessory/auction state, per-room flag mutations on the prototype room. **Housing is soft-blocked on hotboot** insofar as any hotboot design must account for housing persistence. This plan documents the blocker but does not wait on it — housing's files are self-contained and reload cleanly from disk after a cold restart.

## C Reference (authoritative)

All line numbers verified 2026-04-18 from `src/house.c` (2853 LOC) and `src/house.h` (165 LOC). Function inventory (22 functions):

| C entry | C line | Role |
|---|---|---|
| `in_same_house` | `house.c:60-81` | Predicate: are two chars in rooms of the same home? Immortals always return TRUE. |
| `do_house` | `house.c:83-494` | Player housing command + immortal admin. 7 subcommands. |
| `do_gohome` | `house.c:497-575` | Teleport to one's first-room vnum. Position + ROOM_NO_RECALL/NO_MAGIC gates. Mount follows. |
| `do_residence` | `house.c:577-628` | List all residences (name, area, type, room count, start vnum). |
| `do_accessories` | `house.c:630-1041` | Buy furniture (obj or mob) for a house. 6 subcommands. |
| `do_homebuy` | `house.c:1043-1391` | Auction commands (bid / show / list / syntax / admin set*). 10 subcommands. |
| `do_sellhouse` | `house.c:1394-1457` | Place one's own home on auction. |
| `save_residence` | `house.c:1460-1493` | Invoke `fold_area` on the host area (saves the whole .are file). |
| `set_house` | `house.c:1495-1631` | Create a new home: allocate `HOME_DATA`, sort-insert, mutate the prototype room (name, desc, flags, exits, lock all doors, key vnum = room vnum), create a housekey object, persist. |
| `remove_house` | `house.c:1633-1751` | Reverse of `set_house`: delete extra rooms, rename room back to "Vacant Residence", unlock doors, drop key object, delete `db/houses/<Name>` file, unlink from list. |
| `add_room` | `house.c:1753-1853` | Allocate next free vnum in `houses.are`, wire exit + reverse exit, apply ROOM_HOUSE flag set, append to `home->vnum[]`, persist. |
| `give_key` | `house.c:1855-1901` | Create or instantiate a housekey obj and `obj_to_char`. Immortal branch creates obj index on demand. |
| `fwrite_house` | `house.c:1903-1962` | Write `db/houses/<Name>` — `#HOME / Name / Vnum x N / Apartment / End` plus one `#OBJECT` block per room's inventory (for house storage). |
| `load_homedata` | `house.c:1964-2006` | Boot-time: read `db/houses/house.lst`, load each file. Fatal on missing `house.lst` (C shuts down the mud). |
| `load_house_file` | `house.c:2011-2102` | Read one `db/houses/<Name>` file: `#HOME` → `fread_house`; `#OBJECT` → `fread_obj(..., OS_VAULT)`; move carried objs from supermob to first room. |
| `update_house_list` | `house.c:2104-2125` | Rewrite `db/houses/house.lst` (one name per line, terminated with `$`). |
| `fread_house` | `house.c:2127-2200` | Parse one `#HOME` block into a `HOME_DATA`. Returns first-room vnum or `-1` on error, `2` on zero first-vnum (used as sentinel by caller). |
| `save_house_by_vnum` | `house.c:2202-2215` | Locate homes containing the vnum and rewrite them (non-apartments only). Called externally on area save. |
| `save_accessories` | `house.c:2217-2255` | Write `db/houses/homeaccessories.dat`. |
| `load_accessories` | `house.c:2257-2306` | Parse `db/houses/homeaccessories.dat`. |
| `fread_accessories` | `house.c:2308-2370` | Parse one `#ACCESSORIES` block. |
| `add_homebuy` | `house.c:2372-2417` | Add a room to auction. Default endtime `7*48` (7 days at 2 ticks/hour), default bid = apartment min or house min. |
| `remove_homebuy` | `house.c:2419-2440` | Remove auction entry, rewrite auction file. |
| `save_homebuy` | `house.c:2442-2474` | Write `db/houses/homebuy.dat`. |
| `load_homebuy` | `house.c:2476-2525` | Parse `db/houses/homebuy.dat`. |
| `fread_homebuy` | `house.c:2527-2597` | Parse one `#HOMEBUY` block. |
| `homebuy_update` | `house.c:2599-2744` | Tick: decrement endtime, on expiry transfer gold, call `remove_house`/`set_house`/`give_key`, login-message offline parties. |
| `load_player` | `house.c:2746-2815` | Offline-player load (memory-only, no descriptor): open pfile, `load_char_obj`, wire in-room. |
| `logoff` | `house.c:2817-2853` | Offline save + extract for the returned-to-disk char. |

### Struct shape (from `src/house.h:66-118`)

```c
#define MAX_HOUSE_ROOMS 5
#define MIN_HOUSE_BID 5000000
#define MIN_APARTMENT_BID 2000000
#define DEFAULT_BID_INCREMENT_PERCENTAGE 3
#define PENALTY_PERCENTAGE 20
#define ADDITIONAL_ROOM_COST 100000000
#define ADDED_ROOM_HOUSING_AREA "houses.are"
#ifdef HOUSE_MOBS
  #define DEFAULT_MOB_PRICE 50000
#endif

struct homebuy_data {
    HOMEBUY_DATA *prev, *next;
    char  *bidder;     // player name or "None"
    char  *seller;     // player name
    int    vnum;       // first-room vnum
    sh_int incpercent; // bid-increment percentage
    int    bid;        // current bid
    int    endtime;    // ticks remaining (2 ticks/hour in C update loop)
    bool   apartment;
};

struct home_data {
    HOME_DATA *prev, *next;
    char *name;
    bool  apartment;
    int   vnum[MAX_HOUSE_ROOMS];  // fixed-size array; 0 marks empty slot
};

struct accessories_data {
    ACCESSORIES_DATA *prev, *next;
    int  vnum;   // obj or mob vnum
    int  price;
    bool mob;    // discriminator
};
```

### Relevant globals and file paths

- `first_home` / `last_home` — `src/house.h:102-103` — list of homes.
- `first_homebuy` / `last_homebuy` — `src/house.h:88-89` — list of auctions.
- `first_accessory` / `last_accessory` — `src/house.h:117-118` — list of accessories.
- `HOUSE_DIR` — `src/mud.h:5413` — `"houses/"` under rundir.
- `HOUSE_LIST` — `src/mud.h:5525` — `"house.lst"`.
- `HOMEBUY_FILE` — `src/mud.h:5526` — `"houses/homebuy.dat"`.
- `ACCESSORIES_FILE` — `src/mud.h:5527` — `"houses/homeaccessories.dat"`.

### Referenced constants already in Go

- `ITEM_HOUSEKEY` — `internal/types/enums.go:638` (iota value).
- `ROOM_HOUSE` — `internal/types/enums.go:702` (iota value, ~bit in `RoomFlags` BitVector).
- `LEVEL_DEMI` — `internal/types/constants.go:51` — threshold for non-owner access to `house` commands and for the `house givekey` self-path (`src/house.c:154,164`). **NOT the gate for `house set / remove / givekey <name>` — that is `LEVEL_GREATER` (see below). Audit corrected 2026-04-19; original draft had the two thresholds swapped.**
- `LEVEL_GREATER` — `internal/types/constants.go:47` — threshold for immortal `house set / remove / givekey <name>` (`src/house.c:307`) AND for accessory add/remove/setprice and homebuy admin branches.
- `SUB_ROOM_DESC` — `internal/types/enums.go` (same iota pool as `SUB_NONE` / `SUB_OBJ_LONG`). Used in C `do_house` at `src/house.c:100-113`. **Verify the Go enum value exists before G4**.
- `HomeVnum` — `internal/types/character.go:209` (integer, already defined for a hotboot hook, not currently written anywhere).

### Referenced constants missing in Go — must add in G0

- `MAX_HOUSE_ROOMS = 5`
- `MIN_HOUSE_BID = 5000000`
- `MIN_APARTMENT_BID = 2000000`
- `DEFAULT_BID_INCREMENT_PERCENTAGE = 3`
- `PENALTY_PERCENTAGE = 20`
- `ADDITIONAL_ROOM_COST = 100000000`
- `DEFAULT_MOB_PRICE = 50000` (unconditional; the `HOUSE_MOBS` ifdef is a build toggle that is effectively always on in shipped data — `ACCESSORIES_DATA.mob` branch writes to the same file unconditionally, so the Go port must define the constant so the mob-accessory path works).
- `ADDED_ROOM_HOUSING_AREA = "houses.are"` — filename lookup for `add_room`.

---

## Go Current State — gap analysis

Files audited 2026-04-18:

- `internal/types/character.go:209` — `HomeVnum int` present (annotated "Hotboot"). **Not set by any code path** (no writer in `internal/` per grep; `internal/handler/handler.go` has the `ACT_SENTINEL` wiring noted in the hotboot plan but nothing writes `HomeVnum` for players). This field is the soft-intended attachment point for "which first-room vnum does this player own?" but it is **not** the right place — C's `HOME_DATA.name` keys the list on player name, not the other way around. The plan's G2 decision: leave `HomeVnum` alone (hotboot seat) and key housing off `World.Homes[].Name`.
- `internal/types/pcdata.go` — no housing-related fields. `Spouse` string at `:130` (Tier 15 marriage lead-in), `ClanName` at `:16`. No `HomeName` or `BidOnHome` fields. Plan G2 *may* add `PCData.BidOnHome string` only if needed by `do_homebuy bid` to short-circuit lookup — see Open Q4.
- `internal/types/room.go:5-59` — no `OwnedBy`, no `Residents`, no `HouseRooms`. Confirmed per roadmap and independent grep. **G1 does NOT add a new field to `RoomIndexData`**. Instead, house membership is determined via `World.Homes` iteration — same semantics as C `in_same_house` walking `first_home`. This avoids splitting the source of truth between `HOME_DATA.vnum[]` and a room-side mirror.
- `internal/persist/` — no `housing.go`. Existing pattern files to model: `subsystems.go` (clan/council/deity loaders, ~400 LOC), `stances.go` (Tranche B), `holidays.go` (if landed — check). None of them use the per-player file-per-home shape. **New persist idiom required.**
- `internal/world/world.go:36-37, 76, 85` — `World.Clans []*ClanData`, `World.DataDir string`. **No `World.Homes`, `World.Homebuys`, `World.Accessories`.** G1 adds all three.
- `internal/act/` — no `housing.go`. Existing pattern files: `olc.go` (400 lines, has `DoRedit` + `EditorSave` callbacks at `:63,138`), `playercfg.go` (Tier 8/13/15, 200+ lines, has `DoBio` / `DoDescription` using `EditorSave`). **Housing's `house desc` subcommand mirrors the `DoBio` pattern exactly — the plan leans on this heavily.**
- `internal/game/editor.go:159-173` — `/s` transition landed in Tier 12. Preserves the exact invariants housing needs: `StopEditing` before invoking callback, one-shot clear, `ch.Desc.Connected = CON_PLAYING` set before the save runs (so a nested start-editing inside the callback is safe). Housing inherits this with zero changes.
- `internal/boot/boot.go:297,323` — established loader-wire pattern (`LoadStancesInto`, `LoadClansFromDir`). G8 adds three more calls: `LoadHomesFromDir`, `LoadHomebuy`, `LoadAccessories`.
- `db/houses/` — contains only `.do_not_delete` and an empty 1-byte `house.lst` (stock distribution). **No `homebuy.dat`, no `homeaccessories.dat`, no per-player home files.** The Go loader must treat absent files as "empty subsystem" (boot-log warning, continue), NOT fatal. Note: C's `load_homedata` at `src/house.c:1981-1984` *shuts the mud down* on missing `house.lst`. Go diverges — bug-and-continue is the project-wide idiom (`CLAUDE.md` Conventions section: "File loaders: `util.Bug()` and continue on bad data").

---

## Go Design

### File layout

Four new Go files, one modified type-package file:

```
internal/types/housing.go                 (new) — struct definitions
internal/types/constants.go               (edit) — append 8 constants
internal/world/world.go                   (edit) — 3 new list fields + accessor helpers
internal/persist/housing.go               (new) — load/save for all three files
internal/act/housing.go                   (new) — six command handlers
internal/boot/boot.go                     (edit) — 3 loader calls + 6 command registrations
```

### Struct schemas (internal/types/housing.go)

```go
// HomeData maps to C struct home_data (src/house.h:92-100).
// Slice instead of fixed [MAX_HOUSE_ROOMS]int array — Go idiom matches the
// existing AreaData / ClanData shape and avoids zero-as-sentinel bugs in loops.
// Length is bounded by MAX_HOUSE_ROOMS at mutation sites (do_house addroom).
type HomeData struct {
    Name      string   // player name (primary key; key matches pcdata.Filename)
    Apartment bool
    Vnums     []int    // first entry is the start room; len <= MAX_HOUSE_ROOMS
}

// HomebuyData maps to C struct homebuy_data (src/house.h:74-86).
type HomebuyData struct {
    Vnum       int    // first-room vnum on auction
    Bidder     string // "None" sentinel for no bidder
    Seller     string
    IncPercent int    // bid-increment percentage
    Bid        int
    Endtime    int    // ticks remaining; C decrements per pulse in homebuy_update
    Apartment  bool
}

// AccessoryData maps to C struct accessories_data (src/house.h:107-115).
type AccessoryData struct {
    Vnum  int  // obj or mob vnum
    Price int
    Mob   bool // true = mob, false = obj
}
```

Rejected alternative: fixed-size `[5]int` for `Vnums`. Preserving C's zero-as-empty-slot convention would force every iteration to stop on first zero and every mutation to find first zero slot. Slice + `len()` is strictly safer in Go and the only observable difference is the wire format — handled at the file-loader boundary (save: skip zero entries; load: append non-zero).

### World additions (internal/world/world.go)

```go
// Under "Clans, Councils, Deities" cluster:
Homes       []*types.HomeData       // sorted by Name (see LookupHomeByName)
Homebuys    []*types.HomebuyData
Accessories []*types.AccessoryData
```

Accessor helpers (methods on `*World`):

- `LookupHomeByName(name string) *types.HomeData` — case-insensitive linear scan.
- `LookupHomebuyByVnum(vnum int) *types.HomebuyData` — used by `do_sellhouse` collision check.
- `LookupHomebuyByIndex(n int) *types.HomebuyData` — 1-based index used by `homebuy show / bid / remove / ...`; iteration order is insertion order (matches C `first_homebuy` list semantics).

### Room ownership model

**Single source of truth is `World.Homes`**. The prototype room carries the `ROOM_HOUSE` flag and a renamed `Name` / `Description` (C `set_house` at `:1541-1563` sets both). Checking "is this room part of player X's home?" is always `w.LookupHomeByName(x).Vnums` membership test — no duplicate mirror on `RoomIndexData`. This matches C (which also doesn't add a back-pointer on rooms — `in_same_house` at `:69-78` walks `first_home` too).

The `ROOM_HOUSE` bit is *advisory*: it lets generic code (mudprogs, room-stat display, arena PvP guards) cheaply ask "is this a house?" without walking the list. `set_house` and `remove_house` toggle it.

### Persistence format

Byte-for-byte identical to C. Three files:

1. **Per-home file at `db/houses/<CapitalizedName>`** (`src/house.c:1919, 1928-1938`):
   ```
   #HOME
   Name        Alice~
   Vnum    1234
   Vnum    1235
   Apartment   0
   End

   #OBJECT
   ... (object save from the first room's contents, written via same
        fwrite_obj code path as player inventory)
   End
   #END
   ```
   The `#OBJECT` blocks are the *room contents* (house storage — drop things in your house, they persist). Maps to existing `internal/persist/player.go` object-save format exactly (`fwrite_obj` output). **Audit corrected 2026-04-19:** the helpers are `writePlayerObj(w, obj, nest)` and `readPlayerObject(sc, lookup, nestObj)` — both unexported in the `persist` package. Since `housing.go` lives in the same package they are callable without export. (The earlier draft referenced non-existent `WriteObjectHierarchy` / `ReadObjectHierarchy`.)

2. **`db/houses/house.lst`** (`src/house.c:2119-2122`):
   ```
   Alice
   Bob
   Carol
   $
   ```
   One capitalized name per line, `$` terminator. Written by `update_house_list` whenever a home is created or removed.

3. **`db/houses/homebuy.dat`** (`src/house.c:2460-2472`):
   ```
   #HOMEBUY
   Apartment   0
   Bid         5000000
   Bidder      None~
   BidIncPerc  3
   Endtime     336
   Seller      Alice~
   Vnum        1234
   End

   #END
   ```

4. **`db/houses/homeaccessories.dat`** (`src/house.c:2245-2252`):
   ```
   #ACCESSORIES
   Vnum        3001
   Price       1000
   Mob         0
   End

   #END
   ```

### Editor substate wiring (house desc)

Pattern inherits from Tier 12 — no new infrastructure. In `DoHouse`:

```go
// subcommand == "desc", after validating ch is in a room belonging to their home
targetRoom := ch.InRoom
ch.EditorSave = func(c *types.CharData) {
    if act.CopyBufferFunc != nil {
        targetRoom.Description = act.CopyBufferFunc(c)
    }
    if act.StopEditingFunc != nil {
        act.StopEditingFunc(c)
    }
    persist.SaveResidence(world.WorldRef, targetRoom)  // fold_area equivalent
    c.Send("House room description set.\n\r")
}
if act.StartEditingFunc != nil {
    act.StartEditingFunc(ch, targetRoom.Description)
}
```

This is deliberately identical to `internal/act/olc.go:63-75` (redit desc). The callback is self-contained; `house desc` does not need a new substate enum. The C `SUB_ROOM_DESC` substate is a dispatch artifact of the shared `stop_editing` → `last_cmd` callback — Go's per-character `EditorSave` closure carries the dispatch information directly and makes the substate redundant.

Housing does **not** introduce a `CON_HEDIT` connection state. The six commands are all flat (argv-parsed), not menu-driven. C doesn't have `CON_HEDIT` either — `do_house` is dispatched like any other command, and the one editor substate (`SUB_ROOM_DESC`) lives inside `CON_EDITING`.

### Auction tick integration

C's `homebuy_update` runs on a game-loop hook — `src/update.c:2687-2691` uses `pulse_houseauc = 1800 * PULSE_PER_SECOND` (= 7200 game-loop iterations at 4 Hz = **30 minutes real time**; fires twice per game-hour matching the endtime 2-ticks-per-hour encoding). **Audit corrected 2026-04-19** — original draft mis-stated this as "once per game hour / `PULSE_TICK` / 30 real seconds", which would have produced a 60x-wrong cadence. Port:

1. Add `HomebuyUpdate(w *world.World, now time.Time)` in `internal/persist/housing.go`.
2. Wire it into `internal/game/update.go` alongside `charUpdate` / `objUpdate`. Frequency: **every 30 minutes real time** (use `1800 * PULSE_PER_SECOND` or the existing pulse constant equivalent; DO NOT use `PULSE_TICK`). Fires 2×/game-hour, matching C's `endtime` tick-unit encoding (`7*48 = 336` half-hours = 7 days).
3. On expiry: transfer gold using offline-player load if seller/bidder is not logged in. Offline load is NOT trivial in Go — see `load_player` at `src/house.c:2746-2815`. Go's closest existing primitive is `persist.LoadPlayer`. Needs a wrapper that does **not** create a `DescriptorData` (plan Open Q5).

### Offline player gold/housing transfer — the hard part

C does this by calling `load_player` with name → reads the pfile into a transient in-memory `CHAR_DATA` with a throwaway `DESCRIPTOR_DATA` → mutates gold / calls `set_house` / calls `save_char_obj` → calls `logoff` to free. The Go port must replicate this without the `DescriptorData` shim. Two strategies:

**Strategy A — transient CharData**: create a `*CharData` with `Desc = nil`, load via `persist.LoadPlayer`, mutate, save via existing `world.SavePlayer` (or equivalent), let GC reclaim. Risk: many Go code paths assume `ch.Desc != nil` before sending output. `homebuy_update` avoids most of these because the transient char is only used for gold math + save, but calls to `add_loginmsg` / `logoff` in C need Go equivalents. **Plan uses Strategy A** — matches the scope of what C does and avoids reinventing descriptor machinery. Gated on Open Q5.

**Strategy B — direct pfile mutation**: skip the `CharData` round-trip entirely. Read pfile, parse gold line, rewrite. Too brittle — pfile format has dependent fields. Rejected.

### Alternatives rejected

- **Room-side ownership mirror** (`RoomIndexData.OwnedBy string`): rejected (see "Room ownership model" above). Splits source of truth.
- **JSON persistence** for `db/houses/*.dat`: rejected. Breaks SMAUG area-file authoring conventions and forbids a future admin from hand-editing. All other Phase-5 subsystems kept the C text format.
- **Per-subcommand command entries** (`DoHouseName`, `DoHouseAddroom`, ...): rejected. C is a single `do_house` with an `arg` switch; Go keeps the same shape and tests the dispatch at one site. Matches `DoRedit` idiom at `internal/act/olc.go:32`.
- **Skipping accessories for the first landing**: tempting (it's 400+ C lines for a menu of furniture) but the `do_house desc` round-trip needs `accessories buy` to test a non-trivial storage scenario end-to-end. Plan keeps accessories in scope.

---

## Task Groups

All test-first. Mutation verification uses `Edit` round-trips only — **destructive git commands are banned project-wide** (`git checkout`, `git reset --hard`, `git stash`, `git restore`; see `_shared.md` → Mutation Verification Safety). Mutation-verify pattern: `Edit` to break the implementation, `go test` confirms red, `Edit` to restore, `go test` confirms green. Test target: `go test -count=3 ./...` on every group.

### G0 — Constants + fixture data (serial, foundation)

- [ ] Append the 8 constants (`MAX_HOUSE_ROOMS`, `MIN_HOUSE_BID`, `MIN_APARTMENT_BID`, `DEFAULT_BID_INCREMENT_PERCENTAGE`, `PENALTY_PERCENTAGE`, `ADDITIONAL_ROOM_COST`, `DEFAULT_MOB_PRICE`, `ADDED_ROOM_HOUSING_AREA`) to `internal/types/constants.go`.
- [ ] Verify `SUB_ROOM_DESC` is defined in `internal/types/enums.go`. If absent, add it to the SUB_* iota pool.
- [ ] Create fixture dir `internal/persist/testdata/housing/` with: `house.lst` (3 lines + `$`), one valid `Alice` home file, one `homebuy.dat` with 2 auctions, one `homeaccessories.dat` with 3 entries (2 obj + 1 mob).

Tests: `internal/types/constants_test.go` — pin all 8 values with table-driven assertions.

### G1 — Type definitions (parallel-safe with G0)

- [ ] Create `internal/types/housing.go` with `HomeData`, `HomebuyData`, `AccessoryData` structs.

Tests: `internal/types/housing_test.go` — construct each struct, assert zero-value semantics (empty `Vnums` slice, `Bidder=""` not "None" sentinel — sentinel is a load-time convention, not a struct invariant).

### G2 — World container (serial after G1)

- [ ] Add `Homes`, `Homebuys`, `Accessories` slice fields to `world.World` in `internal/world/world.go`.
- [ ] Add `LookupHomeByName`, `LookupHomebuyByVnum`, `LookupHomebuyByIndex` methods.

Tests: `internal/world/world_test.go` — add 3 cases: LookupHomeByName hit/miss/case-insensitive; LookupHomebuyByVnum hit/miss; LookupHomebuyByIndex 1-based + out-of-range returns nil.

### G3 — Persist: load path (serial after G0, G1, G2)

- [ ] Create `internal/persist/housing.go`.
- [ ] `LoadHomesFromDir(w *world.World, dir string) error` — parse `house.lst`, iterate entries, load each per-home file. **Divergence from C: missing `house.lst` logs via `util.Bug` and returns nil; C shuts down the mud.** Per-file parser errors also bug-and-continue (skip the file).
- [ ] `LoadHomebuy(w *world.World, path string) error` — parse `homebuy.dat`. Missing file → empty list, nil error.
- [ ] `LoadAccessories(w *world.World, path string) error` — parse `homeaccessories.dat`. Missing file → empty list, nil error.
- [ ] Use `persist.NewScanner` for all three (matches the clan/stance loaders).

Tests: `internal/persist/housing_test.go` — 5 new tests:
1. `LoadHomesFromDir` with fixture dir: 3 homes loaded, `Alice` has 2 vnums, `Bob` is apartment.
2. `LoadHomesFromDir` with missing `house.lst`: returns nil, `w.Homes` empty, bug log seen.
3. `LoadHomebuy` with fixture: 2 auctions, correct fields. `Bidder == "None"` preserved verbatim.
4. `LoadAccessories` with fixture: 3 entries, mob flag discriminator correct.
5. Round-trip placeholder — file created in G3, asserted-round-trip in G4.

### G4 — Persist: save path (serial after G3)

- [ ] `SaveHome(dir string, h *types.HomeData) error` — write `db/houses/<CapitalizedName>`.
- [ ] `SaveHomebuy(w *world.World, path string) error` — rewrite `homebuy.dat`.
- [ ] `SaveAccessories(w *world.World, path string) error` — rewrite `homeaccessories.dat`.
- [ ] `UpdateHouseList(w *world.World, dir string) error` — rewrite `house.lst` in alphabetical order, `$` terminator.
- [ ] `SaveResidence(w *world.World, room *types.RoomIndexData) error` — invoke `persist.SaveArea(w io.Writer, wld *world.World, area *types.AreaData)` from `internal/persist/area_write.go:13` on the room's area. (`WriteArea` does not exist — audit corrected 2026-04-19.) `SaveResidence` opens the destination file and threads `world` through.

Tests: add to `internal/persist/housing_test.go`:
6. Round-trip: Load fixture, re-save to `t.TempDir()`, re-load, deep-equal.
7. `SaveHome` includes object block when the first-room has contents (asserts `#OBJECT` marker in output).
8. `UpdateHouseList` writes alphabetical order regardless of insert order.
9. `SaveResidence` calls `persist.SaveArea` on the correct area (use a stub area with one room, assert file created).

Mutation check: flip `CapitalizedName` to raw name in `SaveHome`, assert test 6 fails.

### G5 — `DoGohome` + `DoResidence` (parallel-safe, simplest commands)

- [ ] Create `internal/act/housing.go`.
- [ ] Port `DoGohome` from `house.c:497-575`: NPC guard, no-desc guard, find-home-by-name, `GetRoomIndex` of first vnum, `ROOM_NO_RECALL` / `ROOM_NO_MAGIC` gates, position switch (POS_DEAD/MORTAL/INCAP/STUNNED reject, SLEEPING/RESTING/SITTING auto-stand, BERSERK/AGGRESSIVE/FIGHTING/DEFENSIVE/EVASIVE/SHOVE/DRAG reject "too busy"), vortex-arrival `act` calls, `CharFromRoom` / `CharToRoom`, mount follows.
- [ ] Port `DoResidence` from `house.c:577-628`: list all homes in `World.Homes`, formatted as the C table (use the exact color codes from C lines 591-627 for parity).

Tests: `internal/act/housing_test.go` — 8 cases:
1. `TestDoGohome_NoHome` — no `World.Homes` entry for player → "You do not own a home."
2. `TestDoGohome_MissingRoom` — entry exists but vnum 0 → "Your home no longer exists."
3. `TestDoGohome_NoRecallRoom` — caller in ROOM_NO_RECALL → refusal message.
4. `TestDoGohome_FromFighting` — position POS_FIGHTING → "too busy".
5. `TestDoGohome_Succeeds` — healthy PC moves, vortex broadcast seen in another PC's output in the source room.
6. `TestDoGohome_WithMount` — mount follows into the home.
7. `TestDoResidence_Empty` — empty `World.Homes` → "There are currently no residences."
8. `TestDoResidence_Lists` — 2 homes, output contains both names and area names in correct order.

Mutation check: delete the `POS_FIGHTING` case, assert test 4 fails.

### G6 — `DoHouse` subcommand dispatcher (serial after G0-G5)

- [ ] Port `DoHouse` from `house.c:83-494`. Break into helpers per subcommand for testability: `doHouseSyntax`, `doHouseGivekey`, `doHouseName`, `doHouseDesc`, `doHouseAddroom`, `doHouseSet` (immortal), `doHouseRemove` (immortal).
- [ ] `doHouseDesc` uses the `EditorSave` callback pattern exactly like `internal/act/olc.go:63-75`. No new game-loop plumbing.
- [ ] Port `SetHouse` from `house.c:1495-1631` as `persist.SetHouse(w, victim, vnum, apartment) error`. Mutates the prototype room (name, description, flags, exits — locks all doors with key vnum = room vnum). Creates a housekey obj prototype. Persists via `SaveResidence` + `SaveHome` + `UpdateHouseList`.
- [ ] Port `RemoveHouse` from `house.c:1633-1751` — reverse, deletes extra rooms in the `ADDED_ROOM_HOUSING_AREA`, removes reset entries, renames room back to "Vacant Residence", deletes per-home file on disk.
- [ ] Port `AddRoom` from `house.c:1753-1853` — allocate next-free vnum in `houses.are`, wire bidirectional exit, ROOM_HOUSE flags, append to `home.Vnums`.
- [ ] Port `GiveKey` from `house.c:1855-1901` — create `ITEM_HOUSEKEY` obj with `Value[0] = vnum`, `ObjToChar`. Immortal branch creates obj index on demand (log warning about `fold_area` needed for persistence — Go path writes the area immediately via `SaveResidence`).

Tests: `internal/act/housing_test.go` add 14 cases:
1. `TestDoHouse_NoArg_Syntax` — no-arg shows help.
2. `TestDoHouse_NPCRejected`
3. `TestDoHouse_NoHomeNonImmortal` — "You do not own a residence."
4. `TestDoHouse_Givekey_Self` — deducts `DEFAULT_MOB_PRICE` gold, key in inventory.
5. `TestDoHouse_Givekey_InsufficientGold`
6. `TestDoHouse_Givekey_ImmortalGivesToOther`
7. `TestDoHouse_Name_RequiresInHouse` — owner in wrong room → refusal.
8. `TestDoHouse_Name_Succeeds` — owner in first-room: title updates, `SaveResidence` called.
9. `TestDoHouse_Desc_InvokesEditor` — `StartEditingFunc` called with current description.
10. `TestDoHouse_Desc_SaveCallback_Persists` — drive `/s` → new description on room, `SaveResidence` called.
11. `TestDoHouse_Addroom_Apartment_Rejected` — apartments can't add rooms.
12. `TestDoHouse_Addroom_Insufficient_Gold`
13. `TestDoHouse_Set_Immortal_CreatesHome` — `set_house` equivalent, new `HomeData` appears in `World.Homes`, key appears in victim's inventory (or available to give), ROOM_HOUSE flag set on prototype room.
14. `TestDoHouse_Remove_Immortal_TearsDown` — prototype room renamed back, flag cleared, per-home file deleted.

Mutation checks (Edit round-trips):
- Delete the `get_trust(ch) < LEVEL_DEMI` gate in `doHouseSet`, assert test 13 still catches it (add non-immortal caller to test 13 variant).
- Change `ROOM_NO_RECALL` to `ROOM_NO_MAGIC` alone in `DoGohome`, assert test 3.
- Remove the `Vnums[0]` sentinel check in `doHouseSet`, assert load-after-save round-trip.

### G7 — `DoAccessories` (parallel-safe with G6)

- [ ] Port `DoAccessories` from `house.c:630-1041`. Subcommands: `list`, `buy <n>`, `show <n>`, and immortal `add <vnum> [mob]`, `remove <n>`, `setprice <n> <cost>`.
- [ ] `buy` branch: find caller's home, refuse if apartment or if caller not in a home room, deduct gold, `CreateMobile` or `CreateObject` into the room, persist.
- [ ] `show` branch: mob/obj field dump (mob: name/race/class/avghp/avgac/numattacks; obj: name/type/flags/weight).

Tests: add 11 to `internal/act/housing_test.go`:
1. `TestDoAccessories_NPCRejected`
2. `TestDoAccessories_List_Empty`
3. `TestDoAccessories_List_MobAndObjDiscriminated`
4. `TestDoAccessories_Add_RequiresLEVEL_GREATER`
5. `TestDoAccessories_Add_ObjVnum`
6. `TestDoAccessories_Add_MobVnum_AppendsMobFlag`
7. `TestDoAccessories_Add_DuplicateRejected`
8. `TestDoAccessories_Buy_NoHome_Rejected`
9. `TestDoAccessories_Buy_ApartmentRejected`
10. `TestDoAccessories_Buy_InsufficientGold`
11. `TestDoAccessories_Buy_Succeeds` — obj or mob appears in the room, gold deducted, `SaveAccessories` + `SaveHome` called.

### G8 — `DoHomebuy` + `DoSellhouse` + auction tick (serial after G6)

- [ ] Port `DoHomebuy` from `house.c:1043-1391`. Subcommands: `list` (default), `syntax`, `show <n>`, `bid <n> <amount>`, and immortal `add <vnum> [apartment]`, `remove <n>`, `setbid <n> <amt>`, `bidincrement <n> <pct>`, `timeremainder <n> <days> <hours>`, `clearbidder <n>`.
- [ ] Port `DoSellhouse` from `house.c:1394-1457`. Uses `AddHomebuy` with the caller's home vnum.
- [ ] Port `AddHomebuy`, `RemoveHomebuy` helpers into `internal/persist/housing.go`.
- [ ] Port `HomebuyUpdate` — auction tick. Wire into `internal/game/update.go` on the hourly pulse (verify pulse cadence in Open Q6).
- [ ] Offline player mutation path — see §Go Design "Offline player gold/housing transfer". Use transient `*CharData` + `persist.LoadPlayer` + `persist.SavePlayer` (confirm function names during G8; plan assumes names per `internal/persist/player.go`).

Tests: add 13 to `internal/act/housing_test.go`:
1. `TestDoHomebuy_List_Empty`
2. `TestDoHomebuy_List_TwoAuctions_DaysHoursFormat`
3. `TestDoHomebuy_Bid_NoAuctionNumber`
4. `TestDoHomebuy_Bid_AlreadyOwnsHome_Rejected`
5. `TestDoHomebuy_Bid_SelfSellerRejected`
6. `TestDoHomebuy_Bid_BelowMinRejected`
7. `TestDoHomebuy_Bid_Succeeds_UpdatesBidder`
8. `TestDoHomebuy_Add_Immortal_DuplicateVnumRejected`
9. `TestDoHomebuy_Add_Immortal_RoomAlreadyHouseRejected`
10. `TestDoHomebuy_Remove_Immortal`
11. `TestDoHomebuy_Timeremainder_Bounds` — days 0..31, hours 0..23 enforced.
12. `TestDoSellhouse_NoHome_Rejected`
13. `TestDoSellhouse_Yes_UsesMinBid` — `homebuy_data.Bid == MIN_HOUSE_BID` (or `MIN_APARTMENT_BID`).

Auction tick tests in `internal/persist/housing_test.go`:
14. `TestHomebuyUpdate_NoBidder_PenalizesSeller` — seller gold drops by 20% of bid.
15. `TestHomebuyUpdate_BidderLacksGold_Penalty_Cycle` — bidder gold drops, auction endtime reset to 144.
16. `TestHomebuyUpdate_CleanSale_TransfersHouseAndGold` — seller loses home, bidder gains home, gold transferred, key given, auction removed.

Mutation check: remove the `bidder->gold -= (bid/100) * PENALTY_PERCENTAGE` line, assert test 15 fails.

### G9 — Boot wiring + command registration (serial after G3, G5, G6, G7, G8)

- [ ] In `internal/boot/boot.go`, after the clan loader block (~line 323):
  ```go
  housesDir := filepath.Join(dataDir, "houses")
  _ = persist.LoadHomesFromDir(w, housesDir)
  _ = persist.LoadHomebuy(w, filepath.Join(housesDir, "homebuy.dat"))
  _ = persist.LoadAccessories(w, filepath.Join(housesDir, "homeaccessories.dat"))
  ```
- [ ] Register six commands in the command table (same block as existing `DoBio` / `DoAfk` / `DoTitle` registrations — follow current trust/position conventions):
  - `house` → `act.DoHouse`, POS_RESTING, level 0.
  - `gohome` → `act.DoGohome`, POS_RESTING, level 0.
  - `residence` → `act.DoResidence`, POS_DEAD, level 0.
  - `accessories` → `act.DoAccessories`, POS_RESTING, level 0.
  - `homebuy` → `act.DoHomebuy`, POS_RESTING, level 0.
  - `sellhouse` → `act.DoSellhouse`, POS_RESTING, level 0.
- [ ] Wire `persist.HomebuyUpdate` into the hourly tick in `internal/game/update.go`.
- [ ] Add command-registration assertions in `internal/boot/boot_test.go` per the existing pattern (`TestBoot_RegistersHouseCommand` etc.).

Tests: add 7 to `internal/boot/boot_test.go`:
1-6. One per command — assert `cmds.Find("house")` etc. returns non-nil with the expected handler (pointer identity).
7. `TestBoot_LoadsHomesFromFixtureDir` — pre-populate `t.TempDir()/houses/` with fixture, assert `w.Homes` has 3 entries after boot.

Mutation check: remove the `persist.LoadHomesFromDir` call, assert test 7 fails.

### G10 — End-to-end testclient scenario (serial after G9)

- [ ] In `internal/testclient/housing_test.go`, exercise: quick-login player → `house` (refused, no residence) → immortal login → `house set <name> <vnum>` → target logs back in → `gohome` → `house desc` → `/s` → quit → relogin → `look` shows new description.
- [ ] Exercise the auction path: `sellhouse yes` → `homebuy list` shows it → different player `homebuy bid 1 <amount>` → trigger tick (call `HomebuyUpdate` manually in test) → auction clears, bidder owns home, seller gained gold.

Tests: 3 cases:
1. `TestHousingE2E_CreateDescribeRelogin`
2. `TestHousingE2E_SellhouseBidTickTransfersOwnership`
3. `TestHousingE2E_Gohome_FromArena_RefusedIfNoRecall`

---

## Acceptance Criteria

| # | Criterion | Evidence |
|---|---|---|
| A1 | `house` command registered, shows syntax on no-arg. | G6 test 1 + G9 test 1. |
| A2 | `house set <name> <vnum>` (immortal) creates a home: `World.Homes` gains entry, prototype room gets ROOM_HOUSE flag, exits locked with key vnum = room vnum, per-home file on disk, `house.lst` updated. | G6 test 13 + G4 test 6 + G9 test 7. |
| A3 | `house desc` enters the line editor; `/s` persists the new description via `SaveResidence` and stays seamless (descriptor back to `CON_PLAYING`). | G6 test 9+10 + G10 test 1. |
| A4 | `house addroom <dir>` allocates a vnum from `houses.are`, wires bidirectional exit, deducts `ADDITIONAL_ROOM_COST`. Apartments are rejected. | G6 test 11 + variant for success. |
| A5 | `house givekey` self-path deducts `DEFAULT_MOB_PRICE` and places a `ITEM_HOUSEKEY` obj with `Value[0] = vnum` in the caller's inventory. Immortal path targets any player. | G6 tests 4, 5, 6. |
| A6 | `house remove <name>` (immortal) reverses `set_house`: flags cleared, extras deleted, per-home file removed, `house.lst` updated. | G6 test 14. |
| A7 | `gohome` teleports to the owner's first-room vnum; rejects ROOM_NO_RECALL / ROOM_NO_MAGIC source rooms; rejects fighting/etc. positions; auto-stands from rest/sleep; mount follows. | G5 tests 1-6. |
| A8 | `residence` lists all homes with area, type (H/A), room count, start vnum. | G5 tests 7-8. |
| A9 | `accessories list` shows all furniture with price; `buy <n>` places obj-in-room or mob-in-room and deducts gold; apartments are rejected for buy. | G7 tests 3, 9, 11. |
| A10 | `accessories add / remove / setprice` (immortal, `LEVEL_GREATER`) mutate the accessory list and persist to `homeaccessories.dat`. | G7 tests 4-7. |
| A11 | `homebuy list` shows all auctions with days/hours remaining. `homebuy bid <n> <amt>` places a bid; rejects bids when caller owns a home, is the seller, is already a bidder, or lacks gold. | G8 tests 1, 2, 4, 5, 6, 7. |
| A12 | `homebuy` immortal subcommands (`add / remove / setbid / bidincrement / timeremainder / clearbidder`) work and persist. | G8 tests 8-11. |
| A13 | `sellhouse yes` places the caller's home on auction with the minimum bid for the residence type; `sellhouse <amount>` takes an explicit minimum. | G8 tests 12, 13. |
| A14 | Auction tick: on expiry with a bidder who can pay, ownership transfers, gold transfers, key is given, auction record removed. No-bidder expiry penalizes seller 20%. Bidder-cannot-pay resets endtime to 144 and penalizes bidder 20%. | G8 tests 14, 15, 16. |
| A15 | All three load paths (`LoadHomesFromDir` / `LoadHomebuy` / `LoadAccessories`) tolerate missing files with a bug-log message and proceed with empty lists. (Divergence from C: C shuts down on missing `house.lst`.) | G3 tests 2 + variant. |
| A16 | Boot loads fixture data correctly and registers all six commands. | G9 tests 1-7. |
| A17 | End-to-end testclient: one player creates a home, describes it, quits, relogins, sees the description. Another player bids on a sellhouse-listed home, tick completes, ownership transfers. | G10 tests 1, 2. |
| A18 | `ROOM_HOUSE` flag is set on all home rooms and cleared on `remove_house`. Generic systems (mudprogs, stat display) can read the bit without walking `World.Homes`. | G6 test 13/14 assertion. |

## Scope Cuts

- **Fine-grained guest ACLs.** C exposes no per-guest lock/unlock beyond "give key" (which hands out a physical key obj). Housing's "guest list" is implicit: anyone holding a key can open the door. Plan preserves this model. A named-guest list with revoke semantics would require a new `GuestList []string` field on `HomeData` and lock/unlock machinery — out of scope.
- **Rent / tax economy.** C has no per-tick rent deduction. Plan does not add one.
- **Multi-home ownership.** C's `HOME_DATA.name` is the primary key, `do_homebuy bid` refuses callers who already own a home. Plan preserves — one home per player.
- **`house` inside an apartment with `addroom`.** C explicitly rejects; plan preserves the rejection. No courier rooms / linked apartments.
- **Per-room `OwnedBy` mirror.** See §Go Design — single source of truth is `World.Homes`.
- **MXP / MSDP exposure of housing state.** Out of Phase 6.
- **House vandalism / defense.** `ROOM_HOUSE` is advisory; combat and crime systems don't interact with it beyond what C provides (which is nothing — housing rooms are normal rooms with locked exits).
- **Hotboot integration of in-flight auctions.** Plan assumes cold restart. Hotboot plan (`plan-phase6-hotboot.md`) must include auction state in its serialization pass when it lands.
- **C's `HOUSE_MOBS` build-time ifdef.** Go port defines `DEFAULT_MOB_PRICE` unconditionally. No runtime toggle.
- **`do_auction` integration.** Housing has its own auction via `do_homebuy`; the global `do_auction` channel (Phase 6 separate plan) does NOT announce house auctions. Matches C.

## Open Questions

1. **Where does the home file live on disk — `db/houses/<Name>` or `db/system/houses/<Name>`?** — Recommended answer: `db/houses/<Name>` (matches existing empty `db/houses/` directory layout and C `HOUSE_DIR = "houses/"`). Task description prompt mentions `system/houses.dat` but stock data already has `db/houses/`. Use `db/houses/`.
2. **Does `SUB_ROOM_DESC` need to be added to `internal/types/enums.go`?** — Grep the current file. If absent, G0 adds it. Plan assumes present (C uses it at `house.c:916-917 mud.h`, Go has the equivalent iota pool).
3. **Does `persist.LoadPlayer` exist today in a form usable for offline load (no descriptor required)?** — Check `internal/persist/player.go`. If not, a helper `LoadPlayerOffline(name string) (*types.CharData, error)` must be added in G8. Plan assumes a thin wrapper is needed.
4. **Should `PCData.BidOnHome` be added as a quick-lookup for `DoHomebuy bid`?** — Recommended answer: no. `Homebuys` is at most O(20) entries in practice; linear scan is fine. Avoid mirror state.
5. **Strategy A for offline gold transfer: what handles concurrent "player logs in during auction-tick processing"?** — Recommended answer: auction tick runs in the game-loop goroutine (same as all `update.go` callers). No concurrency. Document the constraint in `HomebuyUpdate`'s doc comment.
6. **Which pulse does C's `homebuy_update` run on?** — **RESOLVED 2026-04-19:** `src/update.c:2687-2691` invokes via `pulse_houseauc = 1800 * PULSE_PER_SECOND` = every 30 minutes real time (2 firings per game-hour, matching the endtime half-hour encoding). Wire accordingly in G8. No executor grep needed.
7. **What happens to objects stored in a house that is `house remove`d?** — C deletes extra rooms in `remove_house:1662-1694` but does NOT extract objects from the first room before renaming. Objects remain in the (now public) prototype room. Plan preserves — `RemoveHouse` does not extract contents. Option B (extract to player inventory on removal) is a behavior change; reject.
8. **Should G10's end-to-end test use real area data (`db/area/houses.are`) or fixtures?** — Recommended: real. `houses.are` exists in stock data and the `ADDED_ROOM_HOUSING_AREA` lookup in `add_room` demands it. Boot tests already load real areas.
9. **`gohome` from an arena room** (Phase 6 Wave D arena just shipped / is landing): does it succeed? — `ROOM_ARENA` flag is not in C's `gohome` reject list; `ROOM_NO_RECALL` *is*. Audit stock arena rooms (10366-10382 per roadmap) for `ROOM_NO_RECALL`. If set (likely), `gohome` is refused. Plan: test case G10 test 3 verifies the refusal via the explicit flag check — does not hard-code the arena room numbers.
10. **Does `house set` need a trust-level check beyond `LEVEL_GREATER`?** — C `:307` checks `get_trust(ch) < LEVEL_GREATER` before immortal `set/remove` branches. Plan preserves. `LEVEL_DEMI` (at line 154) is the cutoff for owning *any* home — demi+ don't need to own one to use `house desc` on arbitrary rooms. Preserve both thresholds.

## Risk Analysis

1. **R1 — Area-write invalidates other builders' in-flight edits.** `SaveResidence` calls `fold_area` on the entire area file. If a builder has `redit`'d another room in the same area without saving, their in-memory changes are persisted unexpectedly (or lost if the area is re-read). Mitigation: same risk exists for the shipped `redit desc` path; the Go writer at `internal/persist/area_write.go` is single-source — there's no in-memory vs on-disk split. Test G4 test 9 asserts round-trip.

2. **R2 — Offline player load/save corrupts the pfile on mid-tick crash.** `HomebuyUpdate` reads a pfile, mutates gold, writes it back. A crash between load and save leaves the player in an inconsistent state (might still own the house but have lost the gold). Mitigation: write to `db/player/<l>/<Name>.tmp` and atomic rename (`os.Rename`). Document in `HomebuyUpdate` doc comment. Belt-and-suspenders: match C's crash-unsafety for parity, but at least log before and after the transfer.

3. **R3 — `add_room` exhausts the housing-area vnum range silently.** `houses.are` has a fixed vnum range (check `low_r_vnum`/`hi_r_vnum` in stock data). When full, `AddRoom` returns false and the player gets "contact an immortal" error. No alerting for immortals. Mitigation: `util.Bug` warning when 80% full; escalate via boot log. Cheap — one extra check in `AddRoom`.

4. **R4 — Case-folding mismatch between `load_homedata` and `house.lst`.** C's `update_house_list` writes `capitalize(home->name)`; `fwrite_house` also capitalizes. Go's `persist.CapitalizedName` equivalent must match exactly (Unicode-sensitive? stock names are ASCII). Mitigation: use `strings.Title` explicitly documented as ASCII-only; if a non-ASCII name ever lands, file-loader tests will catch it via the round-trip assertion (G3 test 1).

5. **R5 — `in_same_house` is not ported in this plan but may be called by already-ported code.** Search all of `internal/` for `in_same_house` callers. Plan: no in-Go callers today (grep confirms zero `InSameHouse` references). **Add `handler.InSameHouse(w, a, b *CharData) bool` in G6 anyway** for completeness and for the auction + mudprog paths that land later in Phase 6. Small — 20 LOC, 3 tests.

6. **R6 — Hotboot landing after housing.** Any in-progress auction, editor session on `house desc`, or player teleported mid-`gohome` must survive hotboot. The hotboot plan (`plan-phase6-hotboot.md`) explicitly names housing as a persistence consumer. Mitigation: housing file formats are already disk-round-trippable; hotboot's serialization only needs to capture editor state + descriptor state. Document in the hotboot plan's dependency list.

7. **R7 — `ITEM_HOUSEKEY` value-slot semantics drift.** C `set_house:1613` stores `key->value[0] = location->vnum`; the unlock path at `src/act_move.c` (verify in G6) reads `value[0]` to determine which door this key opens. If Go's door-unlock logic reads a different slot, the key appears to fit nothing. Mitigation: grep `ITEM_HOUSEKEY` in `internal/` before G6 — any existing unlock path must read `Value[0]`. If it reads `Value[1]`, fix the unlock path with an explicit comment citing C.

8. **R8 — Per-home file vs per-player pfile ordering on boot.** If a home file references a player who doesn't exist (e.g. pfile was manually deleted), `load_homedata` silently skips via `util.Bug`. The home's rooms remain mutated (ROOM_HOUSE flag set, description "Alice's House") — an orphan. Mitigation: on boot, post-load sweep that checks each `HomeData.Name` against existing pfile; log orphans. Optional admin command `house orphans` to list them. Deferred to Open Q handling — not in this plan's G-groups but tracked in `TODO.md` after G9.

9. **R9 — `save_residence` reentrancy during `house desc` `/s`.** If the `EditorSave` callback triggers `SaveResidence`, which re-reads and re-writes the area, concurrent players holding a pointer to the room could see stale state. Mitigation: `internal/game/editor.go:159-173` already transitions out of `CON_EDITING` before calling the callback, so there's no "middle of edit" window. But the whole area write is not atomic — document in `SaveResidence` that it must only be called from the game-loop goroutine (all command handlers satisfy this).

10. **R10 — `update_house_list` output ordering.** C writes names in `first_home` iteration order (sort-inserted in `set_house:1519-1525`). Go's `SaveHomes` using `sort.Strings` on names is byte-equivalent for ASCII. But if the load path reads the list file and reconstructs the in-memory list in file order, and a subsequent `set_house` inserts at the sorted position (not end), the next `update_house_list` write produces a different byte sequence → gratuitous diff churn. Mitigation: always `sort.Strings` on write; don't rely on insertion-order preservation. Plan test G4 test 8 pins this.

---

## Dependencies

- **Hard:** none. All primitives exist.
- **Soft:** Hotboot (`plan-phase6-hotboot.md`). Housing adds significant persistent state (one file per home, plus global auction file). Any hotboot design must capture in-flight housing state (e.g. open editor on `house desc`). Housing does *not* block on hotboot — cold restart works — but the persistence scope for hotboot grows by whatever housing writes.
- **Internal (Phase 5 tiers landed):** Tier 12 (`EditorSave` callback + `StartEditingFunc` seam), `internal/persist/area_write.go` (Phase 3), `persist.LoadPlayer` / `persist.SavePlayer` (Phase 1), `handler.CreateMobile` / `handler.CreateObject` / `handler.CharToRoom` (Phase 2), `handler.GetRoomIndex` (Phase 1).
- **Wave-D parallels:** `plan-phase6-olc-redit.md` has the same `EditorSave` consumer pattern. Housing and OLC-redit can ship in parallel; they share zero code but proof-test the same seam.

---

## Executor Protocol Reminders

- **Never run** `git checkout`, `git restore`, `git reset --hard`, or `git stash`. Mutation verification uses `Edit` round-trips. See `_shared.md` → Mutation Verification Safety.
- **All C citations must be `file:line` precise.**
- **TDD mandate**: every G-group is test-first. Write the failing test, run `go test` to confirm red, implement, run `go test -count=3 ./...` to confirm green.
- **Enqueue-before-ack**: before marking any G-group done, update `CHANGELOG.md` (or lineage draft) and `TODO.md`.
- **Adversary after each G-group**: each worker completion gets an independent adversary review before the next G-group kicks off. Escalation protocol per `manager.md`.

---

## Appendix: C Call-Graphs for Executor Orientation

### Boot-time load (matches what G3 / G9 must replicate)

```
boot_db()                                         (src/db.c)
  load_homedata()                                 (house.c:1964)
    for each filename in db/houses/house.lst:
      load_house_file(name)                       (house.c:2011)
        rset_supermob(pRoom)
        loop:
          fread_letter '#'
          fread_word:
            "HOME"   → fread_house(fp)            (house.c:2127)
                         — returns first-room vnum or -1
                         — on success: LINK into first_home
            "OBJECT" → fread_obj(..., OS_VAULT)   (save.c)
                         — adds obj to supermob inventory
            "END"    → break
        transfer supermob inventory to first-room
        release_supermob()

  load_accessories()                              (house.c:2258)
    parse db/houses/homeaccessories.dat:
      each #ACCESSORIES → fread_accessories (house.c:2309)
                           — LINK into first_accessory

  load_homebuy()                                  (house.c:2477)
    parse db/houses/homebuy.dat:
      each #HOMEBUY → fread_homebuy (house.c:2528)
                       — LINK into first_homebuy
```

Go equivalent call-graph (G3 + G9):

```
boot.Boot(w, dataDir)
  persist.LoadHomesFromDir(w, filepath.Join(dataDir, "houses"))
    open house.lst → iterate lines → for each:
      loadHomeFile(w, housesDir, name)
        NewScanner → section dispatch:
          "#HOME"   → readHomeBlock → append to w.Homes
          "#OBJECT" → readObjectBlock → attach to first-room
          "#END"    → done
  persist.LoadAccessories(w, accessoriesPath)
  persist.LoadHomebuy(w, homebuyPath)
```

### `do_house set <name> <vnum>` flow (G6 executor reference)

```
do_house (trust >= LEVEL_GREATER branch)          (house.c:313-449)
  arg parse: "set" arg2=victim arg3=vnum [apartment]
  get_char_world(arg2)
  reject IS_NPC(victim)
  scan first_home for existing home under victim's name
  if "addroom" arg3: delegate to add_room branch
  else if existing home: "They already have a house"
  else:
    get_room_index(atoi(arg3))
    set_house(victim, vnum, apt)                  (house.c:1495)
      CREATE(tmphome, HOME_DATA, 1)
      STRALLOC name, set apartment, vnum[0]=vnum
      sort-insert into first_home
      update_house_list()                         (house.c:2104)
      fwrite_house(tmphome)                       (house.c:1903)
      get_room_index(vnum) → location
      mutate location: name, description, sector=0, max_weight=2000
      set bits: ROOM_NO_SUMMON/NO_ASTRAL/INDOORS/HOUSE; clear PROTOTYPE
      for each exit in location:
        set ISDOOR/CLOSED/LOCKED/NOPASSDOOR/PICKPROOF/BASHPROOF
        key = location->vnum
        add_reset(area, 'D', 0, vnum, vdir, 2)
        if rexit: mirror on reverse side; fold_area(rexit->area)
      delete_obj(location->vnum) if obj prototype collides with key vnum
      make_object(vnum, 0, "<name> <apt|house> key")
        item_type=ITEM_HOUSEKEY, level=1, short_descr, description
        wear_flags: ITEM_TAKE|ITEM_HOLD
        clear ITEM_PROTOTYPE
      save_residence(location)                    (house.c:1460)
        fold_area(location->area, filename, FALSE)
```

Go executor note: the `make_object` step creates an obj *prototype* (index entry), not a running instance. In Go this maps to writing a new `ObjIndexData` into the area and invoking `persist.WriteArea`. The instance given to the victim is created later via `give_key` → `create_object(keyindex, 1)`. Keep the two-step separation; do not collapse into a single `CreateObject` call.

### `homebuy_update` tick flow (G8 executor reference)

```
homebuy_update()                                  (house.c:2599)
  for each home in first_homebuy:
    if --home->endtime > 0 && (bid + inc > 0):
      save_homebuy(); continue                    (not expiring)

    if home->bidder != "None":
      load_player(home->bidder)                   (house.c:2746)
      if bidder->gold < home->bid:
        penalize bidder 20%
        endtime = 144 (3 days reset)
        bidder = "None"
        save + logoff bidder; continue

    if load_player(home->seller) == NULL:
      remove_homebuy(home); log error
      if bidder online: tell; else add_loginmsg
      continue

    if !bidder:
      penalize seller 20%
      tell-or-loginmsg; remove_homebuy; logoff; continue

    if seller < LEVEL_GREATER:
      remove_house(seller)                        (house.c:1633)

    set_house(bidder, home->vnum, home->apartment)
    seller->gold += home->bid
    bidder->gold -= home->bid
    give_key(bidder, home->vnum, bidder->name, home->apartment)
    remove_homebuy(home)
    tell-or-loginmsg both parties; logoff both
```

Go executor note: this is the single most complex function in the plan (mirrors house.c's 145-line function). Break into at least six helpers for testability: `auctionExpired`, `penalizeBidder`, `penalizeSeller`, `transferOwnership`, `loadOrGetOnline`, `persistOrLogoff`. Table-drive the branch cases in the G8 auction-tick tests.

---

**Planned 2026-04-18; external adversary review recommended before dispatch.**
