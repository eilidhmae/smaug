# Plan: Phase 6 — Planes

**Status:** Planned (2026-04-18). External adversary audit 2026-04-18 (lineage `audit-planes`) — verdict **CONCERNS**. Factual C/Go citations accurate with minor line-offset fixes; one design-level algorithmic bug in `DoPset delete` slice-splice pattern (corrected in-plan); one factual error re: C `one_argument` quote support (corrected in-plan).
**Priority:** Phase 6 Wave 1 — small self-contained port; C source is ~150 LOC of real code inside a 298-line file (most of the rest is banner + whitespace). Roadmap Open Question 5 flags the feature as "mostly cosmetic" (named groupings of rooms); the orchestrator has committed to Wave-1 planning regardless. See Scope Cuts / Deferrals for how the cosmetic-only nature shapes the boundary of this plan.
**Scope:** New files `internal/persist/planes.go` + `internal/persist/planes_test.go`, `internal/act/planes.go` + `internal/act/planes_test.go`. Modifications to `internal/world/world.go` (new `Planes []*types.PlaneData` slice), `internal/boot/boot.go` (loader wire + `CheckPlanes` call + three command registrations + package-var wire for `act.PlanesFilePath`). No changes to `internal/types/room.go` — both `RoomIndexData.Plane *PlaneData` (line 26) and `PlaneData` struct (line 112) already exist; a prior roadmap mistake claiming these were missing was corrected in the 2026-04-18 audit.

---

## Problem

SMAUG ships a plane-data subsystem at `src/planes.c:51-298` that lets immortals author named "planes" (a soft grouping of rooms with a single `name string` payload) via `do_pset` CRUD, lets any char list them via `do_plist`, lets immortals stat one via `do_pstat`, and assigns a default "prime plane" back-reference to every room at boot via `check_planes`. Data persists at `system/planes.dat`.

The Go port has **the struct scaffolding in place** but nothing else:

1. `internal/types/room.go:26` — `Plane *PlaneData` field on `RoomIndexData` (dormant; never read, never written).
2. `internal/types/room.go:111-114` — `PlaneData` struct with a single `Name string` field.
3. `db/system/planes.dat` — 5-byte stock file containing exactly `#END\n`. No populated plane data ships.

Missing today:

- No `world.World.Planes` slice. The 2026-04-18 audit of `internal/world/world.go` confirms the "Clans, Councils, Deities" block at lines 36-39 has no planes entry.
- No `LoadPlanes` / `SavePlanes` in `internal/persist/`.
- No `DoPlist` / `DoPstat` / `DoPset` in `internal/act/`.
- No `CheckPlanes` orphan-assignment. Every room in the port's 1909 loaded rooms has `RoomIndexData.Plane == nil`. Today that's harmless because nothing reads the field; as soon as any consumer reads it (this plan's `do_pstat` for symmetry with `do_rstat`, or any future subsystem that keys off plane membership), a nil-deref will hit without the `CheckPlanes` pass.
- No boot wire for the planes loader.
- No registration of the three commands.

Consequence today: a builder who types `pset "Astral" create` sees the "Huh?" response because no such command is registered. A feature that is a documented SMAUG builder primitive is dead in the Go port. **Purely cosmetic — no gameplay interaction (combat, movement, persistence beyond planes.dat itself) depends on plane membership in C either.**

### Why this plan ships despite the feature being cosmetic

1. The roadmap's Open Question 5 ("is this feature worth porting at all?") resolves to "yes, ship it" because the schema scaffolding already lives in `types/room.go` and leaving it dormant forever signals an incomplete port. Three C commands + one loader + one saver + one boot-pass are the minimum-viable close.
2. Shipping the cosmetic CRUD now avoids the larger diff later if/when a Phase-6 builder adds a plane-membership-dependent feature (e.g., plane-scoped announce broadcast, plane-scoped area listing in `do_areas`). The back-reference `room.Plane` becomes populated by `check_planes` after this plan lands; any future reader gets a non-nil pointer for free.
3. The deferred extensions (per-plane room-listing in `do_pstat`, plane-flag on `room.RoomFlags`, area-level plane binding) are tracked in Scope Cuts — future plans can land them without revisiting this work.

---

## C Reference (authoritative)

All citations are against `src/planes.c` (HEAD) + `src/mud.h` + `src/db.c` + `src/build.c` at HEAD.

### Struct

**`PLANE_DATA`** — declared at `src/mud.h:3455-3460` (3-field struct: `next`, `prev`, `name`; verified via direct Read during audit). The C struct has `next` / `prev` linked-list pointers — the Go port replaces these with slice membership on `World.Planes`, matching the existing pattern for clans / deities / bans. Go `PlaneData` at `types/room.go:111-114` ships only `Name string` — no next/prev equivalents (slice handles ordering).

### File path

`src/mud.h:5491` — `#define PLANE_FILE SYSTEM_DIR "planes.dat"` → `db/system/planes.dat` in default layout.

### Entry points

- **`do_plist`** — `src/planes.c:50-59`. Player-facing listing (no trust gate). Sends the header `"Planes:\n-------\n"` then iterates `first_plane` → `last_plane` printing `"%s\n\r"` per name. No argument parsing.

- **`do_pstat`** — `src/planes.c:61-75`. Immortal stat of a single plane. `one_argument` → `arg`. `plane_lookup(arg)` → nil triggers `"Stat which plane?\n\r"`. Hit prints `"Name: %s\n"` and returns. **That's the entire stat output** — the C PLANE_DATA struct holds only `name`, so there's nothing else to show. No trust gate in the function body; registration-level enforcement only.

- **`do_pset`** — `src/planes.c:77-147`. Immortal CRUD dispatcher. Algorithm:
  1. `one_argument(argument, arg)` — first token is the plane-name OR the literal `save`.
  2. Empty `arg` → 6-line syntax help (`pset <plane> create` / `pset save` / `pset <plane> delete` / `pset <plane> <field> <value>` / blank / "Where <field> is one of:" / "    name").
  3. `str_cmp(arg, "save") == 0` → `save_planes()` + `"Planes saved.\n\r"`. Subcommand shortcut.
  4. Else `one_argument(argument, mod)` — second token is the operation (`create` / `delete` / `name`). Plane resolved via `plane_lookup(arg)` — may be nil.
  5. `!str_prefix(mod, "create")` (C: `mod` is a prefix of the word `"create"`, i.e., `"c"`, `"cr"`, `"cre"`, `"crea"`, `"creat"`, `"create"` all match; longer inputs like `"creates"` do NOT match because `str_prefix("creates", "create")` walks through `astr="creates"` and returns TRUE at `astr[6]='s'` vs `bstr[6]='\0'` — the "not-a-prefix" branch. Negated = FALSE = no match. `str_prefix` at `src/db.c:4563-4585` returns FALSE when `astr` IS a prefix of `bstr`, TRUE otherwise. Go idiom: `strings.HasPrefix(bstr, astr)` — mapping is `strings.HasPrefix("create", mod)`, which is FALSE for `mod="creates"`. Semantics match.):
     - Plane already exists → `"Plane already exists.\n\r"`.
     - Else CREATE → link into list + `"Plane created.\n\r"`.
  6. Plane nil (after the create branch, so any non-create op on a non-existent plane) → `"Plane doesn't exist.\n\r"`.
  7. `!str_prefix(mod, "delete")`:
     - UNLINK from list + STRFREE name + DISPOSE struct + call `check_planes(p)` (which **reassigns** every room that pointed at the just-deleted plane back to `first_plane`) + `"Plane deleted.\n\r"`.
     - **No confirm prompt** — unlike `do_setholiday delete`, planes delete is one-shot.
  8. `!str_prefix(mod, "name")`:
     - Remainder of `argument` (whatever follows `<plane> name `) is the new name.
     - Duplicate-check: `plane_lookup(argument) != nil` → `"Another plane has that name.\n\r"`.
     - Else STRFREE old + STRALLOC new + `"Name changed.\n\r"`.
  9. Fallthrough → recursive `do_pset(ch, "")` → re-emit syntax help.

- **`plane_lookup`** — `src/planes.c:149-161`. Two-pass case-insensitive match:
  1. First pass: exact `str_cmp` match. Returns first hit.
  2. Second pass: `!str_prefix(name, p->name)` — first plane whose `name` has `name` as a prefix. This means `plane_lookup("prim")` resolves to `"Prime Material"`. Identical pattern to the skill-lookup double pass.

- **`save_planes`** — `src/planes.c:163-188`. Truncate-write. Per plane:
  ```
  #PLANE
  Name      <name>
  End
  
  ```
  (two-space indent; four trailing spaces after `Name`; blank line between blocks). Terminator: `#END\n`. Failure opens `perror` + `bug`. Note: **`Name` is NOT tilde-terminated** in C's output — `fprintf(fp, "Name      %s\n", p->name)`. This is a divergence from the holidays/clans convention where string fields ARE tilde-terminated. The Go loader MUST handle the C output format.

- **`read_plane`** — `src/planes.c:190-230`. Per-block KVP parser. Keys: `Name` (via `KEY("Name", p->name, fread_string(fp))` — fread_string IS tilde-terminated), `End` (terminator). Unknown word → `bug("read_plane: unknown field '%s'", word)` + `fread_to_eol`.

  **Format mismatch between saver and loader (C bug, or deliberate tilde-tolerance?):** the saver writes `Name      prime\n` (space-delimited, no tilde). `fread_string` (`src/db.c:4370-4431`) reads until tilde. On a line like `Name      Prime Material\n#PLANE\n`, `fread_string` would consume everything up to the next tilde — which would be inside the next plane block or at EOF. This is likely a latent C bug; the shipped `db/system/planes.dat` is `#END\n` only, so the bug has never surfaced in practice.

  **Go port resolution (see D3 below):** write tilde-terminated on save AND accept tilde-terminated on load. The shipped stub file has zero plane blocks so round-tripping the stock tree is invariant under this convention change. If someone has a hand-authored C-format `planes.dat` with space-delimited names, document the migration (see Open Questions Q1).

- **`load_planes`** — `src/planes.c:232-269`. Open `PLANE_FILE`; loop until EOF:
  - `fread_letter` — must be `#` or bug + break.
  - `fread_word` — `"END"` → break (success). `"PLANE"` → `read_plane`. Other → bug + break.
  - Missing file → bug + return (non-fatal). Matches C's "missing file is OK" convention.

- **`build_prime_plane`** — `src/planes.c:271-281`. Creates one plane with `name = "Prime Material"` and LINKs it into the empty list. Called only by `check_planes` when `first_plane == nil`.

- **`check_planes`** — `src/planes.c:283-298`. Algorithm:
  1. If `first_plane == nil`, call `build_prime_plane` to ensure at least one plane exists.
  2. For every room in the global `room_index_hash[MAX_KEY_HASH = 2048]` hash: if `r->plane == nil OR r->plane == p`, set `r->plane = first_plane`.

  Called from three places:
  - `src/db.c:860` — end of boot, after `load_planes`.
  - `src/planes.c:129` — inside `do_pset delete`, with the about-to-be-freed plane as `p` so rooms pointing at it get reassigned before the free.
  - `src/build.c:8010` / `src/build.c:8050` — inside `do_foldarea` and `do_makerooms` (area-vnum-repack + area-mass-create); passed NULL so it just handles the `first_plane == nil` defensive init. (These callers are NOT in scope for Tranche/Phase-6-Planes — `foldarea` is deferred in the roadmap.)

### Dispatch wiring in C

`src/tables.c:1175-1190` registers the three `do_*` names for the function-pointer lookup table. `src/mud.h:4899,4907,4908` declares the prototypes. `src/db.c:826` calls `load_planes();` during `boot_db`. `src/db.c:860` calls `check_planes(NULL);` at the end of boot.

The shipped `db/system/en/commands.dat` does NOT contain entries for `plist`, `pstat`, or `pset` (verified 2026-04-18 via `grep -c "do_plist\|do_pstat\|do_pset" commands.dat` — zero matches). For reference the stock `aset` entry lives at `commands.dat` lines 167-173 (Level 59). In C, these commands are therefore unreachable via the player interpreter in the stock tree unless a builder has manually added `cmdedit` entries. **The Go port will register them unconditionally at boot** (see D5) to make the feature usable on a stock data tree, consistent with the Go convention (clans, deities, holidays all register unconditionally regardless of commands.dat).

---

## Go Current State

Verified 2026-04-18 via direct Read + Grep:

- `internal/types/room.go:26` — `Plane *PlaneData` field on `RoomIndexData`. Dormant: `grep -rn "\.Plane" internal/` returns zero runtime reads (the only hits are in docs and the field declaration itself).
- `internal/types/room.go:111-114` — `PlaneData` struct with `Name string`. No other fields. Matches the C payload.
- `db/system/planes.dat` — 5 bytes, content `#END\n`. Mirrors the stances.dat / morph.dat pattern: shipped empty, loader must be a no-op.
- `internal/world/world.go:36-39` — "Clans, Councils, Deities" block (`[]*types.ClanData`, etc.). Planes slice needs to slot alongside these. `World.New` at line 80 initializes the maps but not the slices (slices default to nil, which `append` handles fine); adding `Planes []*types.PlaneData` to the struct needs no change to `New`.
- `internal/world/world.go:90` — `GetRoom` returns a room by vnum. The room-by-vnum iteration needed by `CheckPlanes` walks `w.Rooms` (the `map[int]*RoomIndexData`), not a `room_index_hash` array. See D4 below.
- `internal/persist/stances.go` — the most recent `system/*.dat` loader (Tranche B, 2026-04-18). Uses `NewScanner` + `ReadWord` + `ReadString` + `ReadNumber`. Pattern for planes loader: top-level `#`-dispatched section loop over `#PLANE` / `#END`, per-block `Name`/`End` KVP parser. Closer in shape to the holidays loader (plan-phase6-holidays.md D3, not yet landed) than to stances — but holidays' draft plan is not yet in the tree, so this plan references stances.go as the shipped analog.
- `internal/persist/scanner.go` exposes `ReadLetter() byte` (fread_letter equivalent), `ReadWord() string`, `ReadString() string` (tilde-terminated), `ReadToEOL() string`. All four are needed.
- `internal/persist/subsystems.go` — contains `LoadClansFromDir`, `LoadDeitiesFromDir`, `LoadBoards`, `LoadSocials`. Pattern for per-subsystem loader signatures. Planes is simpler (single file, not a directory), so a dedicated `internal/persist/planes.go` is cleaner than extending subsystems.go. Same file-placement choice stances.go made.
- `internal/boot/boot.go:290-349` — the loader invocation block. Planes loader slots in after stances / before clans. Boot-fatal-vs-warn semantics: stances / clans / deities / socials / boards all use `log.Printf("WARNING: ...")` on failure. Planes follows the same convention — missing file is non-fatal.
- `internal/boot/boot.go:81-134` — the package-var wiring block. `act.WorldRef = w`, `act.SaveFunc`, etc. `act.PlanesFilePath string` is the new package var this plan adds. Matches the already-landed holidays-plan convention.
- `internal/boot/boot.go:374+` — `registerCommands` func. Three new `reg.Register(...)` lines. Levels resolved below.
- `internal/act/` — standard home for commands. No dedicated `planes.go` file exists today. New file follows naming convention of `channels.go` / `playercfg.go`.
- `internal/act/WorldRef` — package-level `*world.World`, set at boot. Reads from this in `DoPlist` / `DoPstat` / `DoPset` follow the exact pattern of `DoClans` at `internal/act/channels.go` / `subsystems` / etc. Nil-guard required (`if WorldRef == nil { return }`), matching Tier-9 auction pattern.
- `internal/types/constants.go:41-60` — level constants. `LEVEL_GREATER = MAX_LEVEL - 6 = 59`. `LEVEL_ASCENDANT = MAX_LEVEL - 5 = 60`. `LEVEL_IMMORTAL = MAX_LEVEL - 14 = 51`. For admin-CRUD like `pset`, SMAUG's `aset` in `commands.dat` lines 87-93 uses level 59; `cset` at 61 = `LEVEL_INFINITE - 1`. **Chosen:** `DoPlist` at Level 0 (player-visible — matches C no-gate convention), `DoPstat` at `LEVEL_IMMORTAL` (read-only admin inspection), `DoPset` at `LEVEL_GREATER` (writes; matches `aset` precedent).
- `internal/util/strings.go:14-41` — `OneArgument(argument string) (first, rest string)` — splits off first whitespace-delimited token AND LOWERCASES it. **Supports quoted arguments** (single or double quotes at `strings.go:22-25`) — `OneArgument(`"Prime Material" delete`)` returns first=`"prime material"`, rest=`"delete"`. Semantics: `OneArgument("Prime Material Plane create")` (no quotes) → first=`"prime"`, rest=`"Material Plane create"`.
  - **Critical implication for `DoPset`:** the first token gets lowercased. If `pset save` arrives, `arg == "save"` matches. Multi-word plane names ARE addressable via quoting: `pset "Prime Material" delete` → arg1=`"prime material"`, arg2=`"delete"`. Without quotes, `pset Prime Material delete` splits on the first space and fails the lookup. **C `one_argument` at `src/interp.c:1126-1163` has the SAME quote-handling logic** (lines 1143-1144: `if (*argument == '\'' || *argument == '"') cEnd = *argument++;`). Both C and Go accept `pset "Prime Material" delete`. Audit 2026-04-18 corrected an earlier plan claim that C lacked quote support. Document the quoted-form requirement in `DoPset` syntax help.
- `internal/util/strings.go:194` — `NumberArgument` for the `N.keyword` form — not needed here.
- `internal/util/strings.go` has NO `StrPrefix` helper. C's `str_prefix(a, b)` returns TRUE when `a` is a (case-insensitive) prefix of `b`, else FALSE. Go idiom: `strings.HasPrefix(strings.ToLower(b), strings.ToLower(a))`. Used inline in `DoPset` mod-dispatch. Document the polarity (C returns-FALSE-on-match, Go `HasPrefix` returns-TRUE-on-match — don't fat-finger the negation).

---

## Go Design

### D1 — No type additions

**Verified:** `RoomIndexData.Plane *PlaneData` and `PlaneData { Name string }` already exist. This plan adds zero types. `PlaneData` remains pointer-compatible with the back-reference on `RoomIndexData` so `check_planes` can set `r.Plane = w.Planes[0]` without allocating.

### D2 — `World.Planes` slice

**File:** `internal/world/world.go` — modify.

Add a single line in the "Clans, Councils, Deities" block (lines 36-39):

```go
Planes []*types.PlaneData
```

No method additions. No `AddPlane`/`RemovePlane` helpers — the command handler uses `append` and slice-delete directly, matching the `DoSetHoliday` and `DoBanSite` patterns. Zero-valued (`nil`) slices are valid; `append(nil, ...)` works; `len(nil) == 0`.

**Rejected alternative:** Doubly-linked `Next`/`Prev` pointers matching C's `first_plane`/`last_plane` + LINK/UNLINK macros. Go idiom is a slice; linked-list indexing is an anti-pattern; clans/deities/bans all use slices. Round-tripping through save/load is invariant.

### D3 — `LoadPlanes` / `SavePlanes`

**File:** `internal/persist/planes.go` (new).

Follows the shape of `internal/persist/stances.go`. Signatures:

```go
// LoadPlanes reads db/system/planes.dat into a slice. Missing file
// returns nil, nil. Malformed sections (non-'#' leader, unknown
// section name, unknown key inside a block) log via util.Bug and skip.
// Mirrors C load_planes at src/planes.c:232-269.
func LoadPlanes(path string) ([]*types.PlaneData, error)

// SavePlanes truncate-writes planes to path in the SMAUG #PLANE block
// format. Failure returns error without mutating path. Mirrors C
// save_planes at src/planes.c:163-188.
func SavePlanes(path string, list []*types.PlaneData) error
```

**Loader flow:**

1. `os.Open(path)`. `os.IsNotExist` → return `nil, nil` (missing-file-OK, matches C).
2. `NewScanner(f, path)`. Loop:
   - `ReadLetter` → must be `#` OR EOF. Any other non-whitespace character → `util.Bug("LoadPlanes: %s: expected '#', got %q", path, c)` + return (break out of the loop with whatever's been accumulated).
   - On EOF as the first letter (stub file `#END\n` has `#END\n` — so letter=`#`, word=`END`), or on any read that finds no more content, return cleanly.
   - `ReadWord` for section name:
     - `"END"` → return (clean terminator).
     - `"PLANE"` → call `readPlaneBlock(sc)`; if it returns a non-nil `*PlaneData`, append to the out slice.
     - Other → `util.Bug("LoadPlanes: %s: unknown section %q", path, word)` + return (match C `break` on invalid section).
3. `readPlaneBlock` — loops:
   - `ReadWord` for key name.
   - `"End"` → return the in-progress `*PlaneData`. If Name is still empty, log `util.Bug("LoadPlanes: %s: plane with no Name", path)` and return nil (caller skips the append). Matches intent of C (though C would accept and LINK a nil-name plane — a latent bug; Go fixes by rejecting).
   - `"Name"` → read the value. **Tilde-tolerant:** if the next byte is a newline / whitespace-then-newline, use `ReadToEOL` trimmed. If the value has a tilde, use `ReadString`. See "Format mismatch" note below — the implementation reads one word then checks if the line tail has a tilde.

   **Simpler correct implementation:** always use `ReadString` (tilde-terminated). The shipped C saver emits NO tilde, but the shipped stub is empty (no blocks), so there's no real-world C-format data to round-trip. All Go-generated planes.dat files will be tilde-terminated (see saver below). A hand-authored C-format file without tildes would cause `ReadString` to over-read — the Go loader should log a bug in that case. **Chosen:** use `ReadString`; document in a Go-doc comment on `LoadPlanes` that C's `fprintf(fp, "Name      %s\n", ...)` is a C-bug-format that this loader does NOT accept (Open Question Q1).

     Actually, **better choice:** use `ReadToEOL()` trimmed, and strip a trailing tilde if present. This accepts both formats (C's space-terminated and SMAUG's tilde-terminated) and round-trips cleanly under both. See D3-alt below.

   - Unknown key → `util.Bug("LoadPlanes: %s: unknown key %q", path, word)` + `ReadToEOL`. Match C's "skip to end of line" recovery.

**D3-alt — Name-reading policy:** use `ReadToEOL()` + `strings.TrimSpace` + strip trailing tilde. Accepts both C's space-terminated and SMAUG's tilde-terminated formats. This is the most forgiving and matches the "loaders fail soft" convention. **Chosen.** Document the policy in the `LoadPlanes` doc comment.

**Saver flow:**

1. `os.Create(path)`.
2. For each plane:
   ```
   #PLANE
   Name      <name>~
   End
   
   ```
   Use `\t\t` between `Name` and value OR align-spaces (the existing `persist/subsystems.go` clan-writer uses tabs). The C saver uses six spaces (`"Name      %s\n"`); tabs render the same for users; either is accepted by `fread_word` on re-read. **Chosen:** match C's six spaces for visual fidelity, but emit a trailing tilde (divergence from C saver, convergence with SMAUG's other string-field convention). Line pattern: `"Name      %s~\n"`. Blank line between blocks. Terminator: `"#END\n"`.
3. Close. Any write error propagates up; caller logs `util.Bug`.

**Atomicity:** Match the project convention (`SaveClans`, `SaveBoards`, `SaveHolidays` per the draft plan) — direct truncate-write; no temp-file-rename. If a future plan adopts atomic writes, planes can join.

### D4 — `CheckPlanes` orphan-assignment

**File:** `internal/persist/planes.go` (same file as loader).

Go function signature:

```go
// CheckPlanes ensures at least one plane exists and reassigns every
// room whose Plane pointer is nil (or points at the about-to-be-deleted
// `deleted` argument) to the first plane. Mirrors C check_planes at
// src/planes.c:283-298. Called at boot (after LoadPlanes) and inside
// DoPset delete (before unlinking the deleted plane).
//
// If w.Planes is empty, a "Prime Material" plane is appended (matches
// C build_prime_plane at src/planes.c:271-281). The zero-valued slice
// is not mutated in place by the caller; CheckPlanes writes back via
// the slice receiver. The caller must pass `&w.Planes` so the append
// is observable. See call-site wiring in boot.go / DoPset.
//
// Passing `deleted = nil` is valid (boot path); every nil-Plane room
// gets first_plane. Passing a non-nil `deleted` reassigns rooms that
// already pointed at that plane.
func CheckPlanes(w *world.World, deleted *types.PlaneData)
```

Algorithm:

1. If `len(w.Planes) == 0`, append `&types.PlaneData{Name: "Prime Material"}` to `w.Planes`. (This is the only place where `"Prime Material"` is hard-coded — the same string as C `build_prime_plane`.)
2. `firstPlane := w.Planes[0]`.
3. For every room in `w.Rooms` (iterating the `map[int]*RoomIndexData`): if `room.Plane == nil || room.Plane == deleted`, set `room.Plane = firstPlane`.

**Import cycle concern:** `persist` imports `types` already; does `persist` import `world`? Let me re-verify: `internal/persist/subsystems.go` signature `LoadClansFromDir(w *world.World, clanDir string) error` — YES, `persist` already imports `world`. No new cycle introduced.

**Why not put `CheckPlanes` in `internal/world`?** `world` is the pure-state package; no logic lives there beyond CRUD on the state. Orphan-assignment is a boot/CRUD concern. Placing `CheckPlanes` in `persist` keeps it near `LoadPlanes` / `SavePlanes` (same cohesion principle that put `ResetAllAreas` in `handler`, not `world`). **Chosen: persist.**

**Alternative: `internal/boot/`.** Boot imports persist; persist does not import boot. `check_planes` is called from boot AND from `DoPset delete` (in act). If it lived in boot, `act` would need to import boot — and boot already imports act, creating a cycle. **Rejected.**

### D5 — Commands

**File:** `internal/act/planes.go` (new).

Three handlers. All three nil-guard `WorldRef` per established convention.

#### `DoPlist(ch *CharData, argument string)` — player-facing listing.

```
Planes:
-------
Prime Material
Astral
Ethereal
```

Iterate `WorldRef.Planes` printing `p.Name` per line. No argument parsing. No trust gate. Matches C `do_plist` at `src/planes.c:50-59` exactly.

Send via `ch.Send`. Port C's `"Planes:\n-------\n"` verbatim (note: C uses `\n` then `\n-------\n`; the `-------` is 7 hyphens; Go port ships the same 7 hyphens).

#### `DoPstat(ch *CharData, argument string)` — immortal stat.

```
Name: Prime Material
```

`OneArgument` → first. Lookup via case-insensitive + prefix match (`planeLookup` helper). Nil → `"Stat which plane?\n\r"`. Hit → `"Name: %s\n\r"`. Port C `do_pstat` at `src/planes.c:61-75`. **Deliberate C-divergence:** C line 73 uses `ch_printf(ch, _("Name: %s\n"), p->name)` — single `\n`. The Go port emits `\n\r` to match every other Go command output; departure would break line-ending handling downstream. Cosmetic; no user-observable difference in Telnet.

#### `DoPset(ch *CharData, argument string)` — immortal CRUD.

Argument pattern: `<plane-or-save> [<op> [<remainder>]]`. Dispatch:

1. `arg1, rest := util.OneArgument(argument)`. Empty → emit 7-line syntax help verbatim from C:

   ```
   Syntax: pset <plane> create
           pset save
           pset <plane> delete
           pset <plane> <field> <value>

     Where <field> is one of:
       name
   ```

   Port C output byte-for-byte (matching the C `\n\r` line endings; Go ships `\n\r`).

2. `strings.EqualFold(arg1, "save")` → call `persist.SavePlanes(act.PlanesFilePath, WorldRef.Planes)`. Log `util.Bug` on error. Emit `"Planes saved.\n\r"` regardless of error (C matches — no error-reporting to the user). Return.

3. `arg2, rest := util.OneArgument(rest)`.
4. `plane := planeLookup(WorldRef.Planes, arg1)` (case-insensitive match, exact then prefix).

5. `strings.HasPrefix("create", strings.ToLower(arg2))` — i.e. `arg2` is a prefix of `"create"`:
   - `plane != nil` → `"Plane already exists.\n\r"` + return.
   - Else append `&types.PlaneData{Name: arg1}` to `WorldRef.Planes`. **Note:** `arg1` arrives lowercased from `OneArgument`. If user typed `pset Astral create`, stored name is `"astral"` (lowercase). This is a C-divergence: C stores `STRALLOC(arg)` after `one_argument` which ALSO lowercases — same behavior. **Verified matching C semantics.** Emit `"Plane created.\n\r"` + return.

6. `plane == nil` (after create-branch exhausted) → `"Plane doesn't exist.\n\r"` + return.

7. `strings.HasPrefix("delete", strings.ToLower(arg2))`:
   - Find index of `plane` in `WorldRef.Planes` via linear scan (`idx := -1; for i, p := range w.Planes { if p == plane { idx = i; break } }`).
   - **C reference (`src/planes.c:126-129`):** UNLINK, STRFREE, DISPOSE, `check_planes(p)`. C passes `p` as a dangling pointer to `check_planes`; the pointer-identity compare `r->plane == p` still works because the compiler does not dereference `p` and the malloc block, though freed, retains its address. Go cannot and should not replicate this pattern.

     **Correct Go order (AUDIT 2026-04-18: earlier "swap-to-end" pattern was broken for mid-index deletions in 3+ element slices — e.g., deleting index 1 of `[a, plane, c]` would remove `c` instead of `plane`. Replaced with standard splice):**
     1. Splice-delete `plane` at `idx`: `w.Planes = append(w.Planes[:idx], w.Planes[idx+1:]...)`. This is O(n) in tail length but `w.Planes` is bounded to tens of entries in practice. Works for any index (first, middle, last).
     2. If `len(w.Planes) == 0`, call `CheckPlanes(WorldRef, nil)` — rebuilds Prime Material and reassigns every room.
     3. Else call `CheckPlanes(WorldRef, plane)` — reassigns any room that still points at the deleted struct to `w.Planes[0]`. Because `plane` has been removed from the slice, `w.Planes[0]` is guaranteed to be a different pointer (safe even if `plane` was originally at index 0).

     This differs from C's UNLINK-then-free-then-check, but produces the same end-state deterministically without relying on dangling-pointer identity. Document in the `DoPset delete` branch's Go-doc comment.

   - Emit `"Plane deleted.\n\r"`.

8. `strings.HasPrefix("name", strings.ToLower(arg2))`:
   - `rest` holds the remainder after `<plane> <op>`. This is NOT OneArgument-split — it's the full-line remainder.
     - **Whitespace gotcha:** `OneArgument` returns `rest` with leading whitespace stripped. Confirmed via `internal/util/strings.go:14-41`. So `"astral name Prime Material"` → arg1="astral", rest after first OA="name Prime Material", arg2="name", rest after second OA="Prime Material". Use that `rest` as the new name.
   - If `planeLookup(WorldRef.Planes, rest) != nil` → `"Another plane has that name.\n\r"` + return.
   - Else `plane.Name = rest` + `"Name changed.\n\r"`.

   **SmashTilde policy:** C does NOT `smash_tilde` the new name before STRALLOC. But SMAUG's convention is that tildes in string fields break file I/O (tilde is the terminator for `fread_string`). Go port should `util.SmashTilde(rest)` before assigning — matches the `DoTitle` / `DoBio` / `DoDescription` convention. Document as a deliberate C-divergence (protects against a corrupt-file class of bug).

9. Fallthrough (unmatched `arg2`) → re-emit syntax help (recursive call with empty `argument` — OR inline the help emitter as a helper to avoid stack recursion). **Chosen:** inline helper `pSetSyntax(ch *CharData)` called from both the empty-arg branch and the fallthrough branch. No recursion.

#### `planeLookup` helper

```go
func planeLookup(list []*types.PlaneData, name string) *types.PlaneData {
    // Pass 1: exact case-insensitive match.
    for _, p := range list {
        if p != nil && strings.EqualFold(p.Name, name) {
            return p
        }
    }
    // Pass 2: case-insensitive prefix match.
    nameLower := strings.ToLower(name)
    for _, p := range list {
        if p != nil && strings.HasPrefix(strings.ToLower(p.Name), nameLower) {
            return p
        }
    }
    return nil
}
```

Unexported (lowercase). Lives in `internal/act/planes.go` — the only consumer. If a future `internal/persist/planes.go` consumer needs it, promote to exported then.

### D6 — Boot wiring

**File:** `internal/boot/boot.go` — modify.

1. After stances loader (around line 299), add:
   ```go
   // Load plane chart from db/system/planes.dat. Missing file is
   // non-fatal (matches C load_planes at src/planes.c:232-269).
   planesPath := filepath.Join(dataDir, "system", "planes.dat")
   planes, err := persist.LoadPlanes(planesPath)
   if err != nil {
       log.Printf("WARNING: failed to load planes: %v", err)
   }
   w.Planes = planes
   act.PlanesFilePath = planesPath
   // Orphan-assignment: every room whose Plane is nil gets the first
   // plane (Prime Material if no file provided one). Matches C
   // check_planes at src/planes.c:283-298, called from src/db.c:860.
   persist.CheckPlanes(w, nil)
   log.Printf("Loaded %d planes.", len(w.Planes))
   ```

   **Ordering:** AFTER area resets (`handler.ResetAllAreas(w)` at line 354) or BEFORE? C calls `load_planes()` before `reset_area` runs in boot_db — verified by reading `src/db.c:826` (`load_planes`) vs `:860` (`check_planes`) vs the reset-area call site (elsewhere). For Go, `ResetAllAreas` creates mob/obj instances, not rooms — rooms are created by `LoadAreas`. So either order works. **Chosen:** insert after the existing stances block (line 299) and before the cross-subsystem loaders (clans/deities/socials/boards). This puts plane-loading early in boot, before any subsystem could theoretically read `room.Plane`, and keeps the `CheckPlanes` call against a fully-populated `w.Rooms` map (which is populated by `LoadAreas` much earlier in `bootDB`, lines 260-ish).

2. Register the three commands in `registerCommands` (line 373+):
   ```go
   reg.Register(&command.Command{Name: "plist", DoFun: act.DoPlist, Position: types.POS_DEAD, Level: 0})
   reg.Register(&command.Command{Name: "pstat", DoFun: act.DoPstat, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
   reg.Register(&command.Command{Name: "pset", DoFun: act.DoPset, Position: types.POS_DEAD, Level: types.LEVEL_GREATER})
   ```

   Levels justified:
   - `plist`: Level 0 (player-visible). C `do_plist` has no gate in the function body; stock commands.dat doesn't register it. Go ships with Level 0 — builders showed planes to players in some MUDs.
   - `pstat`: Level 51 (`LEVEL_IMMORTAL`). Read-only admin stat.
   - `pset`: Level 59 (`LEVEL_GREATER`). Write/CRUD; matches `aset` precedent in stock commands.dat line 87-93.

   **Open Question Q3:** confirm levels are right for SMAUG lore. `LEVEL_GREATER = 59` aligns with `aset`, `mset`, `oset`, `rset`. Defer decision to orchestrator review.

3. Register `act.PlanesFilePath string` as a new package var in `internal/act/planes.go`. Wired at boot (step 1 above).

---

## Task Groups

### G1 — Scanner / Persister: `LoadPlanes` + `SavePlanes` + `CheckPlanes`

**Files:**
- New: `internal/persist/planes.go`, `internal/persist/planes_test.go`.
- New fixtures: `internal/persist/testdata/planes_empty.dat` (`#END\n`), `internal/persist/testdata/planes_two.dat` (two tilde-terminated blocks), `internal/persist/testdata/planes_cformat.dat` (two space-terminated blocks — hand-authored C-format for tolerance test), `internal/persist/testdata/planes_malformed.dat` (unknown section + unknown key).

**Test-first (loader):**

- `TestLoadPlanes_MissingFileReturnsNilNil` — path that does not exist → `list == nil`, `err == nil`, no bug log.
- `TestLoadPlanes_EmptyFileReturnsEmpty` — file `#END\n` → `len(list) == 0`, no bug log (round-trips the shipped stub).
- `TestLoadPlanes_TildeTerminatedRoundTrip` — fixture `planes_two.dat` with two blocks using `Name      Astral~` / `Name      Ethereal~` → `len == 2`; `list[0].Name == "Astral"`, `list[1].Name == "Ethereal"`.
- `TestLoadPlanes_CFormatTolerance` — fixture `planes_cformat.dat` with space-only format `Name      Astral\nEnd\n` → loader accepts without tilde; `list[0].Name == "Astral"` (whitespace trimmed). This pins the D3-alt policy.
- `TestLoadPlanes_BlockWithNoNameDropped` — block with only `End` keyword (no Name line) → dropped with bug-log; subsequent valid block loads normally.
- `TestLoadPlanes_UnknownSectionStopsParsing` — `#BOGUS` after `#PLANE Astral End` → first plane loaded, bug-logged; `#BOGUS` stops the outer loop; no subsequent blocks loaded. Matches C `break` behavior.
- `TestLoadPlanes_UnknownKeyInBlockSkipsLine` — block with `#PLANE\nName Astral~\nGarbage\nEnd\n` → plane loaded as `"Astral"`, `Garbage` bug-logged, `ReadToEOL` recovery.
- `TestLoadPlanes_NonHashLeaderBugsAndStops` — file starts with `XYZ` (not `#`) → bug-log + return empty.
- `TestLoadPlanes_ShippedStubIsNoOp` — load the real `/home/eilidh/src/smaug/db/system/planes.dat` → `len == 0`, no bug log. Pins that the production stub is honored.

**Test-first (saver):**

- `TestSavePlanes_EmptyListWritesTerminatorOnly` — `SavePlanes(path, nil)` → file content `#END\n` (5 bytes). Reloading gives `len == 0`. Round-trip.
- `TestSavePlanes_TwoPlanesWritesTwoBlocks` — `SavePlanes(path, [Astral, Ethereal])` → file contains `#PLANE\nName      Astral~\nEnd\n\n#PLANE\nName      Ethereal~\nEnd\n\n#END\n`. Reloading gives `len == 2` with matching names.
- `TestSavePlanes_NameWithSpacesPreserved` — plane name `"Prime Material"` (C stub says this is the default). Save + reload → exact match `"Prime Material"` after tilde-strip.
- `TestSavePlanes_FailureReturnsError` — path in a nonexistent directory → returns non-nil error; no panic.

**Test-first (CheckPlanes):**

- `TestCheckPlanes_EmptySliceAppendsPrimeMaterial` — `w.Planes = nil`; `w.Rooms[3001] = &RoomIndexData{}` (Plane nil); call `CheckPlanes(w, nil)` → `len(w.Planes) == 1`, `w.Planes[0].Name == "Prime Material"`, `w.Rooms[3001].Plane == w.Planes[0]`.
- `TestCheckPlanes_NilOrphansAssigned` — `w.Planes = [Astral]`; `w.Rooms[3001].Plane = nil`, `w.Rooms[3002].Plane = Astral`; call `CheckPlanes(w, nil)` → both rooms point at Astral.
- `TestCheckPlanes_DeletedMatchReassigned` — `w.Planes = [Astral, Ethereal]`; `w.Rooms[3001].Plane = Astral`, `w.Rooms[3002].Plane = Ethereal`; call `CheckPlanes(w, Ethereal)` (simulating mid-delete) → room 3002 gets reassigned to `w.Planes[0]` (Astral); room 3001 unchanged.
- `TestCheckPlanes_NoopWhenEverythingAssigned` — `w.Planes = [Astral]`; all rooms already point at Astral; call `CheckPlanes(w, nil)` → no changes; no alloc.
- `TestCheckPlanes_StableAcrossMultipleCalls` — idempotent: two calls produce the same state.

**Mutation-verify (via `Edit` round-trips only — BANNED for mutation revert: `git checkout -- <file>`, `git checkout <ref> -- <file>`, `git restore <file>`, `git reset --hard` any form, `git stash` any form. Apply the mutation with `Edit`; run test; confirm failure; call `Edit` again with the opposite change to revert):**

- Drop the trailing-tilde strip in `LoadPlanes` → `TestLoadPlanes_TildeTerminatedRoundTrip` fails (name comes back as `"Astral~"`).
- Swap `firstPlane := w.Planes[0]` to `w.Planes[len(w.Planes)-1]` → `TestCheckPlanes_DeletedMatchReassigned` fails (rooms end up on the wrong plane).
- Skip the `Name: ""` rejection in `readPlaneBlock` → `TestLoadPlanes_BlockWithNoNameDropped` fails (empty-name plane gets appended).
- Change the `strings.EqualFold` in `planeLookup` pass 1 to `==` → `TestDoPlist_CaseInsensitiveLookup` (see G2) fails when the request has different casing.
- Remove the `append(&types.PlaneData{Name: "Prime Material"})` fallback in `CheckPlanes` → `TestCheckPlanes_EmptySliceAppendsPrimeMaterial` fails.
- Skip the `room.Plane == deleted` branch → `TestCheckPlanes_DeletedMatchReassigned` fails.

**Acceptance gate:** `go build ./internal/persist/...` clean. `go test -count=3 ./internal/persist/...` green.

### G2 — Commands: `DoPlist` + `DoPstat` + `DoPset`

**Files:**
- New: `internal/act/planes.go`, `internal/act/planes_test.go`.
- Modified: `internal/boot/boot.go` (three register calls + `act.PlanesFilePath` wire).

**Test-first (`DoPlist`):**

- `TestDoPlist_EmptySlicePrintsHeaderOnly` — `WorldRef.Planes = nil`; call `DoPlist(ch, "")` → output `"Planes:\n-------\n"` (no plane lines).
- `TestDoPlist_TwoPlanesLists` — `WorldRef.Planes = [Astral, Ethereal]`; call `DoPlist(ch, "")` → output ends with `"Astral\n\rEthereal\n\r"`.
- `TestDoPlist_NilWorldRefNoPanic` — `WorldRef = nil` (save/restore); call `DoPlist(ch, "")` → no panic; no output (or bug log — match the channels-pattern convention).
- `TestDoPlist_IgnoresArgument` — `DoPlist(ch, "anyarg")` produces the same output as `DoPlist(ch, "")`.

**Test-first (`DoPstat`):**

- `TestDoPstat_EmptyArgPromptsStatWhich` — `DoPstat(ch, "")` → `"Stat which plane?\n\r"`.
- `TestDoPstat_UnknownPlaneSamePrompt` — `DoPstat(ch, "nonexistent")` → `"Stat which plane?\n\r"`.
- `TestDoPstat_ExactMatchPrintsName` — `WorldRef.Planes = [Astral]`; `DoPstat(ch, "astral")` → `"Name: Astral\n\r"`.
- `TestDoPstat_PrefixMatchPrintsName` — `WorldRef.Planes = [Astral, Ethereal]`; `DoPstat(ch, "eth")` → `"Name: Ethereal\n\r"`.
- `TestDoPstat_CaseInsensitive` — `DoPstat(ch, "ASTRAL")` → `"Name: Astral\n\r"`.
- `TestDoPstat_ExactBeforePrefix` — `WorldRef.Planes = [Astra, Astral]`; `DoPstat(ch, "astra")` → exact match `Astra`, not prefix `Astral`.

**Test-first (`DoPset` — syntax / save branches):**

- `TestDoPset_EmptyArgPrintsSyntax` — `DoPset(ch, "")` → 7-line syntax help verbatim.
- `TestDoPset_SaveSubcommand` — `WorldRef.Planes = [Astral]`; `act.PlanesFilePath = t.TempDir() + "/planes.dat"`; call `DoPset(ch, "save")` → file exists with `#PLANE\nName      Astral~\nEnd\n\n#END\n`; user sees `"Planes saved.\n\r"`.
- `TestDoPset_SaveErrorStillEmitsMessage` — `act.PlanesFilePath = "/nonexistent/dir/planes.dat"`; call `DoPset(ch, "save")` → user sees `"Planes saved.\n\r"` (matches C "no error propagation"); `util.Bug` called with the I/O error (test seam via `util.BugLogger`).

**Test-first (`DoPset` — create):**

- `TestDoPset_CreateNewPlane` — `WorldRef.Planes = nil`; `DoPset(ch, "astral create")` → `len(WorldRef.Planes) == 1`, `WorldRef.Planes[0].Name == "astral"` (lowercased by OneArgument, matching C); user sees `"Plane created.\n\r"`.
- `TestDoPset_CreateDuplicate` — `WorldRef.Planes = [Astral]`; `DoPset(ch, "astral create")` → no append; user sees `"Plane already exists.\n\r"`.
- `TestDoPset_CreatePrefixMatch` — input `"astral c"` (prefix of `create`) → equivalent to `"astral create"`. Pin via this test.
- `TestDoPset_CreateCaseInsensitiveOp` — input `"astral CREATE"` → same as lowercase create. Pin via this test.

**Test-first (`DoPset` — delete):**

- `TestDoPset_DeleteNoOp` — `WorldRef.Planes = nil`; `DoPset(ch, "astral delete")` → user sees `"Plane doesn't exist.\n\r"`; no change to slice.
- `TestDoPset_DeleteRemovesAndReassigns` — `WorldRef.Planes = [Astral, Ethereal]`; `w.Rooms[3001].Plane = Ethereal`; `DoPset(ch, "ethereal delete")` → `len(WorldRef.Planes) == 1`, `WorldRef.Planes[0].Name == "astral"`; `w.Rooms[3001].Plane == Astral` (reassigned); user sees `"Plane deleted.\n\r"`.
- `TestDoPset_DeleteLastPlaneRebuildsPrime` — `WorldRef.Planes = [Astral]`; `w.Rooms[3001].Plane = Astral`; `DoPset(ch, "astral delete")` → `len == 1`, `WorldRef.Planes[0].Name == "Prime Material"`; `w.Rooms[3001].Plane.Name == "Prime Material"`; user sees `"Plane deleted.\n\r"`.
- `TestDoPset_DeleteFirstPlane` — `WorldRef.Planes = [Astral, Ethereal]`; `DoPset(ch, "astral delete")` → `len == 1`; remaining plane is `Ethereal`. Pins that index-0 deletion keeps the surviving plane at index 0. (Post-audit: replaces the earlier "swap-before-remove" test name; the algorithm now uses a standard slice splice.)
- `TestDoPset_DeleteMiddlePlane` — **added post-audit 2026-04-18.** `WorldRef.Planes = [Astral, Ethereal, Astral2]`; `w.Rooms[3001].Plane = Ethereal`; `DoPset(ch, "ethereal delete")` → `len == 2`; `w.Planes == [Astral, Astral2]`; `w.Rooms[3001].Plane == Astral`. Pins that splice-delete works for mid-index targets. This case would have silently passed the broken "swap-to-end + remove-last" algorithm (which would have deleted `Astral2` instead of `Ethereal`, leaving a dangling reference in room 3001 to the still-alive `Ethereal` struct).
- `TestDoPset_DeleteLastPlane` — **added post-audit.** `WorldRef.Planes = [Astral, Ethereal]`; `DoPset(ch, "ethereal delete")` → `len == 1`; remaining `Astral`. Symmetric to the first-plane case.

**Test-first (`DoPset` — name):**

- `TestDoPset_RenameChangesName` — `WorldRef.Planes = [Astral]`; `DoPset(ch, "astral name Astrolabe")` → `WorldRef.Planes[0].Name == "Astrolabe"` (SmashTilde applied, but this value has no tildes); user sees `"Name changed.\n\r"`.
- `TestDoPset_RenameToDuplicate` — `WorldRef.Planes = [Astral, Ethereal]`; `DoPset(ch, "astral name Ethereal")` → name unchanged; user sees `"Another plane has that name.\n\r"`.
- `TestDoPset_RenameSmashTilde` — `DoPset(ch, "astral name Funky~Name")` → `WorldRef.Planes[0].Name == "Funky-Name"` (tilde smashed to dash). Deliberate C-divergence (C doesn't smash); documents Go-convention protection.
- `TestDoPset_RenameMultiWord` — `DoPset(ch, "astral name Prime Material")` → `WorldRef.Planes[0].Name == "Prime Material"` (full remainder used, not just first word of `rest`).

**Test-first (`DoPset` — fallthrough):**

- `TestDoPset_UnknownOpRePrintsSyntax` — `DoPset(ch, "astral wombat")` → same 7-line syntax-help output as `DoPset(ch, "")`. Pin that the fallthrough path DOES the inline helper, not an unbounded recursive call.

**Mutation-verify (via `Edit` round-trips only — same banned list as G1):**

- Remove the `planeLookup` pass-2 (prefix) branch → `TestDoPstat_PrefixMatchPrintsName` fails.
- Swap the `strings.HasPrefix("create", arg2)` to `strings.HasPrefix(arg2, "create")` → `TestDoPset_CreatePrefixMatch` fails (user input `"c"` no longer matches because `HasPrefix("c", "create")` is false).
- Remove the `util.SmashTilde` call in the rename branch → `TestDoPset_RenameSmashTilde` fails.
- Call `CheckPlanes` AFTER the slice-remove in delete → `TestDoPset_DeleteLastPlaneRebuildsPrime` still passes (because `CheckPlanes(w, nil)` with empty `w.Planes` still seeds Prime Material), but `TestDoPset_DeleteRemovesAndReassigns` fails for the reassignment — `w.Rooms[3001].Plane` still points at the removed `Ethereal` struct because `CheckPlanes` sees `len(w.Planes)==1, deleted=Ethereal` and `w.Planes[0]=Astral`, which happens to work. **Re-verify:** this mutation may NOT flip the test. A stronger mutation: skip the swap-before-remove AND call `CheckPlanes` after. Document in the mutation table.
- Flip the lookup ordering (prefix first, then exact) → `TestDoPstat_ExactBeforePrefix` fails.
- Remove the nil-guard `if WorldRef == nil { return }` → `TestDoPlist_NilWorldRefNoPanic` panics.

**Acceptance gate:** `go build ./internal/act/...` clean. `go test -count=3 ./internal/act/...` green.

### G3 — Boot wiring + end-to-end round-trip

**Files:**
- Modified: `internal/boot/boot.go` (add loader invocation block; add three register calls; wire `act.PlanesFilePath`).
- Modified: `internal/boot/boot_test.go` (add post-boot assertions).
- New: `internal/testclient/planes_test.go` — E2E test via the testclient harness, driving `plist` / `pstat` / `pset` through the real login flow.

**Test-first (boot):**

- `TestBoot_RegistersPlaneCommands` — after `boot.Boot`, the command registry has `"plist"`, `"pstat"`, `"pset"` resolvable via `reg.Find` at the expected levels (0, 51, 59). Pins the G3 registration lines didn't get reverted by a future mutation.
- `TestBoot_LoadsStockStubFile` — boot with real `db/system/planes.dat` (5-byte `#END\n`) → `len(w.Planes) == 1`, `w.Planes[0].Name == "Prime Material"` (from `CheckPlanes` fallback). Every room in `w.Rooms` has `room.Plane == w.Planes[0]`.
- `TestBoot_AssignsEveryRoomAPlane` — after boot, iterate every room in `w.Rooms`; every one has `room.Plane != nil`. Pins A9.
- `TestBoot_WiresPlanesFilePath` — after boot, `act.PlanesFilePath == filepath.Join(dataDir, "system", "planes.dat")`. Pins that the SaveFunc path is ready.

**Test-first (testclient E2E):**

- `TestTestclient_PlistShowsPrimeMaterial` — login as admin via `Harness.QuickLogin`; type `plist`; expect `"Prime Material"` in the output.
- `TestTestclient_PsetCreateAndListRoundTrip` — login; `pset astral create`; `plist`; expect `"astral"` in output. Then `pset astral delete`; `plist`; expect no `"astral"` in output.
- `TestTestclient_PsetSavePersistsToDisk` — login; `pset astral create`; `pset save`; assert `act.PlanesFilePath` contains a block with `Name      astral~`.

**Mutation-verify (via `Edit` round-trips only — same banned list as G1):**

- Remove the three `reg.Register` lines → `TestBoot_RegistersPlaneCommands` fails.
- Remove the `persist.CheckPlanes(w, nil)` call → `TestBoot_AssignsEveryRoomAPlane` fails.
- Swap `act.PlanesFilePath = planesPath` to an unrelated value → `TestBoot_WiresPlanesFilePath` fails AND `TestTestclient_PsetSavePersistsToDisk` fails.

**Acceptance gate:** `go build ./...` clean. `go vet ./...` clean. `go test -count=3 ./...` green across all packages (no regression).

---

## Acceptance Criteria

**G1 (loader / saver / CheckPlanes):**
- A1. `persist.LoadPlanes("/home/eilidh/src/smaug/db/system/planes.dat")` on the shipped `#END\n`-only stub file returns a non-nil empty slice (`len(list) == 0`, `list != nil`, `err == nil`) — the stub parses cleanly as "terminator reached at first letter." Missing-file returns `(nil, nil)` (distinct case: nil slice AND nil error, no bug log). The D3 test `TestLoadPlanes_EmptyFileReturnsEmpty` pins the empty-slice branch; a sibling `TestLoadPlanes_MissingFileReturnsNil` pins the missing-file branch. These two cases must be distinguishable by callers because `CheckPlanes` keys its default-seed decision on the empty-AND-non-nil branch (stub loaded cleanly, no data) vs the nil branch (file absent, treat as first-boot).
- A2. `persist.SavePlanes` round-trips a slice of plane names through load → save → load identically, including multi-word names and tilde-terminated output.
- A3. `persist.CheckPlanes(w, nil)` guarantees at least one plane exists and `room.Plane != nil` for every room in `w.Rooms`.
- A4. `persist.CheckPlanes(w, deleted)` reassigns every room pointing at `deleted` to `w.Planes[0]` without nil-deref even when `deleted` is the last plane.

**G2 (commands):**
- A5. `DoPlist` prints `"Planes:\n-------\n"` followed by one line per plane, matching C `do_plist` at `src/planes.c:50-59`.
- A6. `DoPstat <name>` resolves `<name>` case-insensitively with exact-then-prefix matching and emits `"Name: %s\n\r"`, or `"Stat which plane?\n\r"` when unresolvable.
- A7. `DoPset` dispatches: empty → syntax help; `save` → persist; `<plane> create` → create+append; `<plane> delete` → unlink + CheckPlanes + reassign; `<plane> name <remainder>` → rename with SmashTilde guard + duplicate-check. Matches C `do_pset` at `src/planes.c:77-147`.

**G3 (integration):**
- A8. `boot.Boot` on a stock data tree registers `plist` / `pstat` / `pset` and sets `act.PlanesFilePath`.
- A9. After `boot.Boot`, every room in `w.Rooms` has `room.Plane != nil` (empty-file case: every room points at the default `"Prime Material"` plane).
- A10. End-to-end via testclient: `pset astral create` + `plist` produces output containing `"astral"`; subsequent `pset astral delete` + `plist` produces output without `"astral"`; subsequent `pset save` persists to disk.

**Global:**
- A11. `go vet ./...` clean.
- A12. `go test -count=3 ./...` green across all packages (no regression in already-landed subsystems).
- A13. Every work-unit mutation documented in G1 / G2 / G3 mutation-verify tables was exercised via `Edit` round-trips with confirmed test failure and revert to green.

---

## Scope Cuts / Deferrals

Items **explicitly NOT** in this plan. Documented so the next reader knows why, and so the roadmap's "planes is cosmetic" framing is enforced rather than drifting into scope creep.

### Cosmetic-scope cuts (feature stays cosmetic)

- **Per-plane room listing in `DoPstat`.** C `do_pstat` prints only `Name: %s` (zero room-membership info). A Go extension could iterate `w.Rooms` and print rooms-on-this-plane. **Deferred:** follow-up plan if any builder-facing request materializes. Not in this plan because it diverges from C.

- **Plane-flag on `RoomFlags` or area-level plane binding.** C assigns `room.Plane` at boot via `CheckPlanes` and allows it to be mutated indirectly by plane deletion; there is NO per-room `rset plane <name>` command in C. Adding one would be novel gameplay (violates roadmap's "no new gameplay invented"). **Deferred — likely indefinitely unless a builder asks.**

- **Plane-scoped broadcasts.** No C subsystem reads `room.Plane` to scope a message. **Permanent cut** unless a future content plan proposes plane-announce as a new feature (new gameplay = out of Phase 6).

### C-feature cuts (also absent in C)

- **Plane inheritance / hierarchy.** C has no notion of a plane-tree. Go port matches — flat list.
- **Plane-scoped reset events.** Planes have no reset metadata in C (`RESET_DATA` is area-scoped).
- **`foldarea` / `makerooms` call-sites of `CheckPlanes(NULL)`.** These are `src/build.c:8010` / `:8050` — both in `do_foldarea` / `do_makerooms`. The roadmap's "foldarea is deprioritized" disposition applies. If those commands ever port, they call `persist.CheckPlanes(w, nil)` at the appropriate point — no new code to plan now.

### Go-idiom cuts (deliberate divergence)

- **No `DoPlaneSave` dedicated command.** C has `save_planes` only inside `do_pset save`. Go follows suit: no standalone `saveplane` command; the `save` subcommand lives inside `DoPset`.
- **No interactive confirm for delete.** C's `do_pset delete` is single-shot (no `yes` confirmation like `do_setholiday delete`). Go port matches — no confirm prompt.

### Technical follow-ups (tracked, not blocking)

- **Atomic save-to-file via `os.CreateTemp` + `os.Rename`.** Matches the project-wide convention of truncate-write. If a future plan upgrades any persister to atomic writes, planes can join. Tracked in TODO.md as a cross-cutting concern.

- **`util.StrPrefix` helper.** This plan inlines `strings.HasPrefix(strings.ToLower(b), strings.ToLower(a))`. If a future plan finds that three+ call sites need the same helper, promote to `util.StrPrefix` in `internal/util/strings.go`.

- **Tilde-tolerant string parsing as a Scanner feature.** This plan's D3-alt adds tilde-tolerance via `ReadToEOL + TrimSpace + TrimSuffix "~"` inline. If a future plan needs the same pattern, add a Scanner method. For now, one call site.

---

## Open Questions (with recommended answers)

### G1

**Q1 — Hand-authored C-format file migration.** If a user has a `planes.dat` file generated by the C server (space-delimited, no tilde terminator), the Go loader's `ReadToEOL + TrimSpace + TrimSuffix "~"` policy accepts it (the `TrimSuffix "~"` is a no-op on C format). The saver always emits tilde-terminated. So after one save pass, a C-format file becomes a Go-format file. **Recommended:** document this in the `LoadPlanes` / `SavePlanes` Go-doc comment. No migration tool needed.

**Q2 — Multi-word plane names.** **AUDIT 2026-04-18 CORRECTION:** both C and Go `one_argument` / `OneArgument` **support quoted arguments**. `OneArgument(`"Prime Material" delete`)` returns arg1=`"prime material"`, arg2=`"delete"` — the quoted multi-word name IS addressable. C source at `src/interp.c:1143-1144` (`if (*argument == '\'' || *argument == '"') cEnd = *argument++;`) and Go source at `internal/util/strings.go:22-25` both treat `'` or `"` as delimiters.

Unquoted multi-word: `pset Prime Material delete` → arg1=`"prime"`, arg2=`"Material"` (both lowercased) — falls through to the syntax-help path. **Recommended:** document the quoted-form requirement in the `pset` syntax-help output (add line `<plane> may be quoted if it contains spaces: pset "Prime Material" delete`). The seeded `"Prime Material"` default is thus editable (via quoting) in both C and Go — matches C verbatim. No divergence. Plan's earlier claim that `"Prime Material"` was un-editable was factually wrong on the C side; corrected.

### G2

**Q3 — Command levels.** `DoPstat` at Level 51 (`LEVEL_IMMORTAL`), `DoPset` at Level 59 (`LEVEL_GREATER`), `DoPlist` at Level 0. **Recommended:** ship as above. Rationale: `aset` / `mset` / `oset` / `rset` / `cset` cluster at 59-61; planes is world-structural CRUD, symmetry with `aset`. Orchestrator can raise/lower if SMAUG lore differs.

**Q4 — `DoPlist` visible to all players or immortal-only?** C registers no commands.dat entry so neither behavior is canonical-stock. The function body has no gate. SMAUG-derivative MUDs show inconsistent behavior (some expose `plist` to players, some gate at 51). **Recommended:** Level 0. Rationale: the information is lore-trivial (named groupings); no PvP / griefing vector; matches the "show the world structure" design of `areas` (Level 0) and `socials` (Level 0).

**Q5 — `DoPstat` minimum output.** C emits only `"Name: %s\n"`. Should Go add anything extra — e.g., a count of rooms on this plane? **Recommended:** NO. Match C exactly. Scope-cut list tracks this as an optional follow-up (per-plane room count would require a room iteration, which is O(n_rooms) per `pstat` call — cheap but unneeded).

### G3

**Q6 — Boot ordering vs `ResetAllAreas`.** Should `LoadPlanes` + `CheckPlanes` run BEFORE or AFTER `ResetAllAreas`? C runs `load_planes` at `src/db.c:826` and `check_planes` at `:860`, with `boot_reset` in between. **Recommended:** Run `LoadPlanes` + `CheckPlanes` between existing boot-db phases — specifically after `LoadAreas` (line 260-ish, which populates `w.Rooms`) and before `ResetAllAreas` (line 354, which only touches mob/obj state not room struct fields). This ensures `w.Rooms` is fully populated when `CheckPlanes` iterates it, and ensures `room.Plane` is non-nil before any mobprog or obj reset could theoretically read it (none currently do, but defense-in-depth).

**Q7 — Testclient admin harness.** `Harness.QuickLogin` returns a mortal character. Admin-gated tests need `LEVEL_GREATER`. **Recommended:** extend `QuickLoginTwo`'s pattern with a `QuickLoginAdmin` or `harness.SetCharLevel(ch, 59)` post-login setter. Since G2 unit tests can directly manipulate `ch.Trust` in-process, the testclient E2E can instead call `pset` via a character promoted post-login via a test-only `world.World.CharByName(name).Trust = 59` poke. **Chosen:** use an existing testclient helper if one exists; if none, add a trivial `setTrust` helper to `internal/testclient/harness.go` (one-line change, doesn't need its own plan). Document in A10 and in the testclient README.

### Boot / wiring

**Q8 — `act.PlanesFilePath` as package var vs `persist.PlanesFilePath` vs `world.PlanesFilePath`.** Matches the pattern from holidays-plan D4 which chose `act.HolidayFilePath`. **Recommended:** `act.PlanesFilePath string`. Wired at boot. Consistent with the established pattern; the only consumer (`DoPset save`) lives in `act`.

---

## Risk Analysis

**Low risk:**

- **G1 (persist layer).** Pure text-file I/O with no in-memory mutation. `CheckPlanes` mutates `w.Rooms` in place, but only assigns the `Plane` field (no structural changes). Nil-safe. The loader is a 20-line pattern already proven by stances.

- **G2 (command handlers).** No existing command touches `w.Planes` or `room.Plane`; the new commands cannot regress any shipped command. The nil-guard for `WorldRef` is identical to the channels pattern; the SmashTilde-on-rename is identical to `DoTitle` / `DoBio`.

**Medium risk:**

- **G3 (boot ordering).** `CheckPlanes` iterates `w.Rooms` and mutates every room's `Plane` field. A fresh `boot.Boot` on a production data tree iterates ~1909 rooms. Each iteration is a pointer-assign — O(1) work per room — so total cost is trivial (< 1ms). The risk is **ordering correctness**: `CheckPlanes` must run AFTER `LoadAreas` (which creates the rooms). Confirmed by inspection: the plan's D6 places `LoadPlanes` + `CheckPlanes` after the `bootDB` call that includes `LoadAreas` — so `w.Rooms` is fully populated. Test: `TestBoot_AssignsEveryRoomAPlane` pins this.

- **Mutation-verify ambiguity in the "CheckPlanes ordering" mutation for `DoPset delete`.** The plan's G2 mutation-verify table flags this: a naive "call `CheckPlanes` after the slice-remove" mutation may not flip any test because the end-state is the same (`CheckPlanes` handles the empty-slice case). **Mitigation:** add a stronger mutation — "skip swap-before-remove AND skip `CheckPlanes` entirely" — and confirm `TestDoPset_DeleteRemovesAndReassigns` fails. Document in the mutation table.

**No high-risk items.** The feature is cosmetic, self-contained, and all primitives are shipped.

---

## Adversary-Resolved Concerns (2026-04-18 manager self-adversary pass + external audit)

Agent-based adversary dispatch was unavailable in both the original manager session AND the external audit session (CLAUDE.md L67-68 environmental constraint). Both passes used structured self-review against authoritative C sources and Go codebase, with every citation spot-verified via direct Read.

**External audit (lineage `audit-planes`, 2026-04-18) applied four in-place corrections:**
- Status header rewritten to record CONCERNS verdict.
- PLANE_DATA line numbers corrected `3456-3462` → `3455-3460` (mud.h struct is a 3-field declaration spanning 5 lines, not 7).
- `str_prefix` commentary clarified — original wording at the create-dispatch case inverted TRUE/FALSE polarity (conclusion was correct).
- C `one_argument` quote-support claim reversed — both C (`src/interp.c:1143-1144`) and Go (`util/strings.go:22-25`) accept `'` and `"` quotes. Multi-word plane names ARE addressable via quoted form in both. Q2 rewritten.
- `DoPset delete` algorithm bug fixed — "swap-to-end + remove-last" pattern silently corrupted mid-index deletions in 3+ element slices. Replaced with standard `append(slice[:idx], slice[idx+1:]...)` splice. Two new test cases added (`TestDoPset_DeleteMiddlePlane`, `TestDoPset_DeleteLastPlane`).
- `commands.dat` cross-reference corrected `aset at lines 87-93` → `lines 167-173`.
- `DoPstat` `\n\r` rationale de-confused (two contradictory sentences merged into one).

1. **CRITICAL — Back-reference field existence verified.** Before writing the plan, this manager directly Read `/home/eilidh/src/smaug/smaug-go/internal/types/room.go` and confirmed:
   - Line 26: `Plane *PlaneData` IS present on `RoomIndexData`.
   - Lines 111-114: `PlaneData` struct IS defined with `Name string`.
   The earlier roadmap mistake (now-corrected in `phase6-roadmap.md:277`) claiming these were missing is resolved. Plan's D1 ships ZERO new types; this is a key boundary that saves scope.

2. **Shipped `planes.dat` is `#END\n` only (5 bytes).** Verified via `xxd` — file is pure ASCII `23 45 4e 44 0a` = `#END\n`. Round-tripping the stock tree is a no-op by design; the loader returns `nil, nil` and `CheckPlanes` seeds `"Prime Material"` as the sole entry.

3. **`commands.dat` does NOT register plist/pstat/pset.** Verified via `grep -B1 -A7 "Code.*do_plist\|..." commands.dat` — zero matches. Go port registers unconditionally in `registerCommands` (matches holidays-plan precedent). Without the unconditional Go registration, the feature would ship dead even after this plan.

4. **`str_prefix` polarity.** C `!str_prefix(mod, "create")` means "`mod` is a prefix of `create`" — TRUE when `mod ∈ {"c", "cr", "cre", ..., "create"}`. Go idiom: `strings.HasPrefix("create", strings.ToLower(mod))`. Plan's D5 documents this — the arguments to `HasPrefix` are (big, small), so the literal constant `"create"` is the FIRST arg. Easy to mis-code; mutation test in G2 pins this.

5. **`plane_lookup` double-pass semantics.** Verified C implementation at `src/planes.c:149-161` — first pass exact `str_cmp`, second pass `str_prefix` (user input is prefix of stored name). Plan's `planeLookup` helper in D5 ports this exactly; `TestDoPstat_ExactBeforePrefix` pins the ordering.

6. **Delete ordering in C has a latent dangling-pointer bug.** Verified via `src/planes.c:126-129`: UNLINK, STRFREE, DISPOSE, then `check_planes(p)`. The freed `p` is passed by identity. Works in C by accident (pointer identity compares unchanged). Go port CANNOT rely on dangling pointers (GC keeps the struct alive while `*PlaneData` is live on the stack, so identity comparison IS still well-defined). Nonetheless, the plan's D5 specifies a cleaner Go ordering (swap-to-end, remove, then CheckPlanes) so the semantics are independent of pointer identity. `TestDoPset_DeleteFirstPlaneSwapsBeforeRemove` pins the swap behavior.

7. **Save format divergence.** C's `save_planes` writes `Name      <name>\n` (no tilde). C's `read_plane` uses `fread_string` (tilde-terminated). This is a latent C format mismatch — the shipped stub has no blocks so it's never exercised. Plan's D3-alt chooses tilde-terminated output + tilde-tolerant input. Accepts both formats on load; emits tilde-terminated on save. `TestLoadPlanes_CFormatTolerance` + `TestLoadPlanes_TildeTerminatedRoundTrip` pin both sides.

8. **`CheckPlanes` location in `persist`.** Rejected `world` (pure-state, no logic) and `boot` (would create `act → boot` cycle because `DoPset delete` calls `CheckPlanes`). `persist` already imports `types` and `world`, and is already where `LoadXxx` / `SaveXxx` live. No cycles. Plan's D4 documents.

9. **`OneArgument` lowercases.** Plan's D5 notes this applies to `arg1` (plane name) and `arg2` (op). Stored plane names are lowercased (matches C `STRALLOC(arg)` after `one_argument`). `TestDoPset_CreateNewPlane` pins `"astral"` (lowercase) as the stored name.

10. **Room-iteration iteration order.** `w.Rooms` is `map[int]*RoomIndexData` — Go map iteration order is randomized. `CheckPlanes` doesn't care about order (assigns every room the same value), so randomization is irrelevant. Mutation concern: if a future version scopes assignment by vnum range, map iteration would need ordering. Plan documents this in the `CheckPlanes` Go-doc comment: "order-independent for this implementation; if scoping is added, switch to a sorted-vnum iteration."

---

## Completion Record

### 2026-04-19 — LANDED

**Problem restated.** SMAUG's plane-data subsystem (`src/planes.c:51-298`) — admin CRUD for named room groupings via `do_plist` / `do_pstat` / `do_pset`, persisted at `db/system/planes.dat`, with `check_planes` orphan-assignment on every room — had zero runtime presence in the Go port despite the struct scaffolding existing at `types/room.go:26,111-114`. A builder typing `pset "Astral" create` saw "Huh?" because no command was registered, and every room's `RoomIndexData.Plane` field was dormant nil. This plan shipped the minimum-viable close: three commands, one loader, one saver, one boot-pass seed.

**Groups shipped (all 3 of 3).**
- **G1 — persist layer.** `internal/persist/planes.go` (170 LOC) + `internal/persist/planes_test.go` (19 tests). `LoadPlanes` returns `(nil, nil)` for missing file and non-nil empty slice for the `#END\n` stub (A1 distinction). `SavePlanes` truncate-writes tilde-terminated blocks. `CheckPlanes` seeds "Prime Material" when slice is empty and reassigns every room whose `Plane == nil || Plane == deleted` to `w.Planes[0]`. Four testdata fixtures (`planes_empty.dat`, `planes_two.dat`, `planes_cformat.dat`, `planes_malformed.dat`). BugSink-capture helper for empty-Name and unknown-key branches.
- **G2 — commands.** `internal/act/planes.go` (188 LOC) + `internal/act/planes_test.go` (23 tests). `DoPlist` / `DoPstat` / `DoPset` plus unexported `planeLookup` (exact-then-prefix, case-insensitive) and `pSetSyntax` helper (inlined to avoid stack recursion on fallthrough). `PlanesFilePath string` package var wired at boot. Splice-delete algorithm uses `append(s[:i], s[i+1:]...)` — works uniformly for first/middle/last index. SmashTilde on rename (deliberate C-divergence).
- **G3 — boot integration.** `internal/boot/boot.go` gains loader invocation + `act.PlanesFilePath` wire + `persist.CheckPlanes(w, nil)` seed, plus three `reg.Register` calls (plist Level 0, pstat LEVEL_IMMORTAL, pset LEVEL_GREATER). Four boot-level tests in `boot_test.go`. Three testclient E2E tests in new `internal/testclient/planes_test.go`.

**Acceptance criteria — all 13 satisfied.**
- A1 ✓ `LoadPlanes` distinguishes missing file `(nil, nil)` from stub `#END\n` `(non-nil empty slice, nil err)`. Tests `TestLoadPlanes_MissingFileReturnsNilNil` + `TestLoadPlanes_EmptyFileReturnsEmpty` + `TestLoadPlanes_ShippedStubIsNoOp`.
- A2 ✓ `SavePlanes` round-trips. Tests `TestSavePlanes_TwoPlanesWritesTwoBlocks` + `TestSavePlanes_NameWithSpacesPreserved` + `TestDoPset_CreateSaveLoadRoundTrip`.
- A3 ✓ `CheckPlanes(w, nil)` seeds Prime Material. Test `TestCheckPlanes_EmptySliceAppendsPrimeMaterial`.
- A4 ✓ `CheckPlanes(w, deleted)` reassigns. Test `TestCheckPlanes_DeletedMatchReassigned`.
- A5 ✓ `DoPlist` emits header + one line per plane. Tests `TestDoPlist_EmptySlicePrintsHeaderOnly` + `TestDoPlist_TwoPlanesLists`.
- A6 ✓ `DoPstat` exact-then-prefix case-insensitive. Tests `TestDoPstat_ExactMatchPrintsName` + `TestDoPstat_PrefixMatchPrintsName` + `TestDoPstat_CaseInsensitive` + `TestDoPstat_ExactBeforePrefix`.
- A7 ✓ `DoPset` dispatches across empty/save/create/delete/name. 17 tests across `TestDoPset_*`.
- A8 ✓ `boot.Boot` registers plist/pstat/pset at correct levels. Test `TestBoot_RegistersPlaneCommands`.
- A9 ✓ Every room has non-nil Plane after boot. Test `TestBoot_AssignsEveryRoomAPlane`.
- A10 ✓ End-to-end testclient: create+plist+delete and save+reload. Tests `TestTestclient_PsetCreateAndListRoundTrip` + `TestTestclient_PsetSavePersistsToDisk`.
- A11 ✓ `go vet ./...` clean.
- A12 ✓ `go test -count=3 ./...` green across all 15 packages.
- A13 ✓ 10 mutations exercised via `Edit` round-trips with confirmed failure and revert to green.

**Tests added:** 45 new tests total (19 persist + 23 act + 4 boot + 3 testclient + 1 act-level integration round-trip). `go test -count=3 ./...` goes from the pre-landing baseline to all-green with the new surface included.

**Mutation-verify summary.** 10 mutations applied via `Edit` tool only (project-banned: `git checkout`/`git restore`/`git reset --hard`/`git stash`). Every mutation flipped its pinned test(s); every revert restored green.

| # | Mutation | Pinned test(s) that fail |
|---|---|---|
| 1 | Drop `TrimSuffix "~"` in `readPlaneBlock` | `TestLoadPlanes_TildeTerminatedRoundTrip` |
| 2 | Skip empty-Name rejection | `TestLoadPlanes_BlockWithNoNameDropped` |
| 3 | Drop Prime Material seed in `CheckPlanes` | `TestCheckPlanes_EmptySliceAppendsPrimeMaterial` |
| 4 | Drop `room.Plane == deleted` branch | `TestCheckPlanes_DeletedMatchReassigned` |
| 5 | Drop `planeLookup` pass-2 prefix match | `TestDoPstat_PrefixMatchPrintsName` |
| 6 | Swap `HasPrefix("create", arg)` args | `TestDoPset_CreatePrefixMatch` |
| 7 | Remove `util.SmashTilde` in rename | `TestDoPset_RenameSmashTilde` |
| 8 | Drop `CheckPlanes` call in `DoPset delete` | `TestDoPset_DeleteRemovesAndReassigns` + `TestDoPset_DeleteLastPlaneRebuildsPrime` + `TestDoPset_DeleteMiddlePlane` |
| 9 | Drop 3 `reg.Register` calls in boot | `TestBoot_RegistersPlaneCommands` |
| 10 | Drop boot-level `CheckPlanes(w, nil)` call | `TestBoot_AssignsEveryRoomAPlane` + `TestBoot_LoadsPlanesAndSeedsPrimeMaterial` |

**Readiness-vet caveat confirmed.** Per orchestrator readiness vet note: the planned weaker mutation "move `CheckPlanes` call AFTER slice-remove" indeed does not flip any existing test (both orderings produce the correct end state because `CheckPlanes(w, deleted)` handles the post-splice case correctly when `deleted` is no longer in the slice). The stronger mutation — **drop** the `CheckPlanes` call entirely — is what demonstrates test coverage of the reassignment pass (mutation #8 above).

**Deliberate C-divergences.**
- `DoPstat` emits `\n\r` (C emits `\n` at `src/planes.c:73`). Cosmetic only; consistent with every other Go command.
- `SavePlanes` emits tilde-terminated names (C emits space-terminated at `src/planes.c:181`). Fixes a latent C format mismatch — C's `read_plane` uses `fread_string` which expects a tilde (`:220`), so the C saver and C loader don't actually round-trip through any populated file. Go loader is tilde-tolerant so both formats work.
- `DoPset name` applies `util.SmashTilde` to the new name (C does not). Matches `DoTitle`/`DoBio`/`DoDescription` precedent; tildes in stored names would break file I/O on re-save.
- `DoPset delete` uses standard splice (`append(s[:i], s[i+1:]...)`) instead of C's UNLINK+STRFREE+DISPOSE+check_planes(freed_pointer) sequence. Same end-state; avoids C's dangling-pointer-identity reliance.

**Preserved from C.**
- `planeLookup` two-pass exact-then-prefix semantics (`src/planes.c:149-161`).
- `OneArgument` lowercases `arg1` so stored names are lowercase — matches C `STRALLOC(arg)` after `one_argument` (C also lowercases the token).
- Rename duplicate-check uses `planeLookup` on the new name, which resolves to the current plane itself as a prefix match when renaming to a same-or-prefix name. Same C false-positive (`src/planes.c:135-138`).
- No trust gate in `DoPlist` body; gating is registration-level (Level 0 = visible to all).

**Scope slippage: none.** All in-plan task groups shipped. SmashTilde-ordering follow-up noted in plan §D5 — resolved in implementation by applying SmashTilde before the duplicate-check (so a tilde-bearing name that would collide with an existing plane's smashed form is rejected early; consistent with the protect-against-corrupt-file motivation).

**Followups queued.** None beyond what's already tracked in `TODO.md`. SmashTilde-ordering note removed from TODO (resolved in-plan). Plan §Scope Cuts items remain deferred as planned (per-plane room listing, plane-scoped broadcasts, plane-flag on RoomFlags).

**Adversary pass.** Structured self-review substituted per the CLAUDE.md tooling caveat (manager-subagent `Agent`-tool availability remains inconsistent). The review catalogued 11 adversary concerns including `HasPrefix("create", "")` empty-prefix semantics (gated with `arg2Lower != ""`), rename-to-self false-positive (verified matches C), map-iteration randomness in `CheckPlanes` (safe because every room gets the same value), and DoPset nil-plane access in the `name` branch (plane != nil by control flow).

**Verdict: LANDED.** Commit `<pending>`.

---

## Appendix A — Fixture for G1 `planes_two.dat` (tilde-terminated)

```
#PLANE
Name      Astral~
End

#PLANE
Name      Ethereal~
End

#END
```

## Appendix B — Fixture for G1 `planes_cformat.dat` (C saver output, space-terminated, no tilde)

```
#PLANE
Name      Astral
End

#PLANE
Name      Ethereal
End

#END
```

## Appendix C — Fixture for G1 `planes_malformed.dat`

```
#PLANE
Name      Astral~
Garbage   this should be skipped
End

#BOGUS
should not be reached

#END
```

Expected load: `[Astral]` with two bug-logs (one for `Garbage`, one for `#BOGUS` — loader returns after the bogus section).

## Appendix D — Boot-order diagram

```
bootDB (internal/boot/boot.go)
  ├── LoadAreas        → populates w.Rooms
  ├── LoadClasses
  ├── LoadRaces
  ├── LoadSkills
  ├── LoadStancesInto  → mutates combat.StanceIndex
  ├── LoadPlanes       ← NEW (this plan): populates w.Planes
  ├── CheckPlanes(w, nil) ← NEW (this plan): seeds Prime Material + assigns every room
  ├── LoadClansFromDir
  ├── LoadDeitiesFromDir
  ├── LoadSocials
  ├── LoadBoards
  └── ResetAllAreas    → populates mobs/objs (does NOT touch room.Plane)

registerCommands
  └── Register plist / pstat / pset ← NEW (this plan)
```

## Appendix E — Rough order of landing (suggested commit phasing)

1. Commit G1: `persist/planes.go` + `persist/planes_test.go` + three testdata fixtures. ~120 Go LOC + ~150 test LOC. Lands `LoadPlanes` / `SavePlanes` / `CheckPlanes`.
2. Commit G2: `act/planes.go` + `act/planes_test.go` + `act.PlanesFilePath` package var. ~100 Go LOC + ~250 test LOC. Lands `DoPlist` / `DoPstat` / `DoPset`.
3. Commit G3: `boot/boot.go` modifications + `boot/boot_test.go` additions + `testclient/planes_test.go`. ~30 lines of boot wire + ~80 LOC of tests.

Total surface: ~250 Go LOC + ~500 test LOC across 3 commits. No cross-package refactors; additive only.
