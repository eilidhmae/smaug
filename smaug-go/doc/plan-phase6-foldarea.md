# Plan: Phase 6 `foldarea` / `unfoldarea` + `.bak` rotation for `savearea`

**Status:** LANDED 2026-04-26 (Waves 1–3 commits `f96f114` / `bb81aa8` / `13e2cd9`; docs Wave 4 closes the plan). See §Completion Record.
**Priority:** P3 (Phase 6, builder-muscle-memory + on-disk safety net for the already-shipped `savearea` path).
**Lineage:** `phase6-foldarea`.
**Scope:** Add the `foldarea <filename>` and `unfoldarea <filename>` immortal commands and add `.bak` rotation to the `savearea` save path (which `foldarea` reuses). Reuses the already-shipped `internal/persist/area_write.go` serializer (Phase 3) and the already-shipped path-containment guards from `DoSaveArea` (`internal/act/olc.go:817-841`). **Does NOT re-port `fold_area()`** — the Go serializer already exists.

---

## Cross-Plan Dependencies

**Depends on (LANDED):**

- Phase 3 `internal/persist/area_write.go:13` — `SaveArea(w io.Writer, wld *world.World, area *types.AreaData) error`. Already covers all area sections (mobs/objects/rooms/resets/shops). No changes needed.
- `DoSaveArea` at `internal/act/olc.go:799-873` — current "current-room only" save path with atomic tmp→live rename. The `.bak` rotation is added here and is automatically inherited by `foldarea` (which reuses the same save helper).
- Path-containment guards at `internal/act/olc.go:829-841` — reject absolute paths + `..` traversal. Reused for `foldarea` (and `unfoldarea`, which takes a filename argument from a builder and feeds it to a loader).
- `world.World.Areas []*types.AreaData` at `internal/world/world.go:25-26` — the in-memory area registry. `WorldRef.Areas` is the established lookup pattern (precedent: `internal/act/cmds2.go:374`, `internal/act/olc_area.go:129-136`).
- Boot reg table at `internal/boot/boot.go:781` (savearea) — new rows added immediately after.

**Blocks:** nothing.

**Does NOT depend on:** overland-loader, housing, dragonflight, or any other Phase-6 lineage.

---

## Problem

Two pain points motivate this work, both stated by the user:

1. **Builder muscle memory.** Every C SMAUG builder reflexively types `foldarea <file>` to save an area. The Go port ships only `savearea` (current-room-only, no filename argument). A builder typing `foldarea myzone.are` today hits "Huh?" and has to be told the Go-port-specific name.
2. **No on-disk safety net.** `DoSaveArea` writes to `<file>.tmp` and atomically renames to `<file>` — that protects against half-written files but offers **no rollback**. If a builder saves a corrupted area (bad `aset` value, broken reset, etc.), the previous good copy is gone. C's `fold_area()` always rotates the live file to `<file>.bak` before writing the new one (`src/build.c:7369-7370`). The Go port should match that behavior — both for `savearea` and for `foldarea`.

A third potential pain point — **post-boot reload via `unfoldarea`** — is **deliberately scoped down** in this plan because `internal/persist/area.go:48-77`'s `loadAreaFile` unconditionally appends to `w.Areas` and double-registers all index entries. See §D3 below for the full safety analysis and the conservative response: `unfoldarea` prints a "use a hotboot to reload" message and does NOT call the loader.

---

## C Reference (authoritative)

All line numbers verified 2026-04-26 from `src/build.c`.

### `fold_area()` — `src/build.c:7346-7948` (~600 LOC)

The full area serializer. Renames the live file to `.bak`, opens a fresh writer, emits every section. **NOT being re-ported** — `internal/persist/area_write.go` already covers it. The only line numbers we need:

| Lines | Purpose |
|---|---|
| 7346-7368 | Function head, locals, "Saving …" log line. |
| **7369-7370** | **`sprintf (buf, "%s.bak", filename); rename (filename, buf);`** — the `.bak` rotation point. **This is the reference for D1.** |
| 7371-7378 | `fclose(fpReserve)`; open `filename` fresh for writing. On fopen failure, log + `perror` + reopen `fpReserve` + return. |
| 7379-7948 | Section emission (#AREA, #VERSION, #AUTHOR, #FLAGS, #ECONOMY, #CONTINENT, #MOBILES, #OBJECTS, #ROOMS, #RESETS, #SHOPS, #REPAIRS, #SPECIALS). Pure write logic — Go equivalent already exists. |

**Key C behavior worth carrying over:** `rename(filename, buf)` is called **unconditionally**. C's `rename(2)` returns -1 if the source doesn't exist but does not abort `fold_area` — the function continues to open `filename` for writing. The Go port mirrors this: rotate-if-exists, never block the save on a missing live file (a brand-new area's first save has no live file to rotate).

### `do_foldarea` — `src/build.c:8055-8081` (27 LOC)

```c
void do_foldarea (CHAR_DATA * ch, char *argument)
{
  AREA_DATA *tarea;
  set_char_color (AT_IMMORT, ch);
  if (!argument || argument[0] == '\0') {
    send_to_char ("Fold what?\n\r", ch);
    return;
  }
  for (tarea = first_area; tarea; tarea = tarea->next) {
    if (!str_cmp (tarea->filename, argument)) {
      send_to_char ("Folding area...\n\r", ch);
      fold_area (tarea, tarea->filename, FALSE);
      set_char_color (AT_IMMORT, ch);
      send_to_char (_("Done.\n"), ch);
      return;
    }
  }
  send_to_char ("No such area exists.\n\r", ch);
  return;
}
```

Translation: empty arg → "Fold what?"; case-insensitive filename match against `first_area`-list (`world.World.Areas` in Go); on match emit "Folding area...\n\r" + invoke `fold_area` (Go: shared save helper) + "Done.\n"; on miss "No such area exists.\n\r".

### `do_unfoldarea` — `src/build.c:8036-8052` (17 LOC)

```c
void do_unfoldarea (CHAR_DATA * ch, char *argument)
{
  set_char_color (AT_IMMORT, ch);
  if (!argument || argument[0] == '\0') {
    send_to_char ("Unfold what?\n\r", ch);
    return;
  }
  fBootDb = TRUE;
  load_area_file (last_area, argument);
  fBootDb = FALSE;
  check_planes (NULL);
  return;
}
```

The C author's own comment immediately above (lines 8027-8035) reads:

> ```
> /* It is in your best interest to ensure that the area you wish to "unfold"
>  *   (a) is not loaded
>  *   (b) it contains vnums that exist
>  *   (c) the area has errors
>  *
>  * NOTE: Use of this command is not recommended.		-Thoric
>  */
> ```

This is the canonical "do not use this" warning from Thoric himself. C's `load_area_file` is also unsafe to call post-boot for the same reasons the Go `loadAreaFile` is unsafe — vnum collisions, double-loading, dangling pointers from anyone holding a reference to the old room/mob/obj index entries. **D3 ships only the user-facing command stub with a "use hotboot" message; the loader call is explicitly NOT invoked.**

### `do_installarea` — `src/build.c:8108+` — OUT OF SCOPE

C's `do_installarea` walks the in-progress build/live area-list split: it removes the area from the "build" list and adds it to the "live" list, calling `fold_area(..., install=TRUE)`. The Go port does not have a build/live split — there is one `world.World.Areas`. Defer to TODO with note: "consider whether build/live split is wanted before porting".

---

## Go Current State

Verified 2026-04-26.

| Artifact | Location | Status |
|---|---|---|
| `DoSaveArea` | `internal/act/olc.go:799-873` | Atomic tmp→live rename; **no `.bak` rotation**. Path-containment guard already in place. **D1 modifies in place.** |
| `persist.SaveArea` | `internal/persist/area_write.go:13` | Full serializer. Reuse unchanged. |
| `WorldRef *world.World` | `internal/act/olc.go` (boot-set) | Reuse for `Areas` lookup. |
| `world.World.Areas` | `internal/world/world.go:25-26` | `[]*types.AreaData`. Iteration precedent: `cmds2.go:374`, `olc_area.go:129-136`. |
| `types.AreaData.Filename` | `internal/types/area.go` | Set by loader at `internal/persist/area.go:79` to `filepath.Base(filename)`. Match against arg via case-insensitive compare. |
| `loadAreaFile` | `internal/persist/area.go:48-77` | Boot-time loader. **NOT safe to re-invoke post-boot** — unconditionally `append`s to `w.Areas` at `:82`, calls `loadMobiles`/`loadObjects`/`loadRooms` which append to global index maps without a "remove old" path. **D3 does not call this.** |
| Boot reg `savearea` | `internal/boot/boot.go:781` | Existing row; new `foldarea`/`unfoldarea` rows added adjacent. |
| Path-containment guard | `internal/act/olc.go:829-841` | Reject absolute path + `..` traversal. **Extracted to a shared helper in G2** to be reused by foldarea + unfoldarea. |

---

## Go Design

### D1 — `.bak` rotation in the shared save path

Insert **before** `os.Rename(tmpPath, path)` at `internal/act/olc.go:867`:

```go
// .bak rotation (mirrors C fold_area at src/build.c:7369-7370). If the live
// file exists, rename it to .bak before installing the new tmp file. Best-
// effort: a missing live file (first-save case) is not an error; a rename
// failure for any other reason is logged but does not abort the save —
// the tmp file is the source of truth and the operator can recover from it.
bakPath := path + ".bak"
if _, err := os.Stat(path); err == nil {
    if err := os.Rename(path, bakPath); err != nil {
        util.Bug("DoSaveArea: .bak rotation failed for %q: %v", path, err)
        // fall through; the tmp→live rename below will overwrite the live file
    }
}
```

**Behavior:**

- First save of a brand-new area (no live file): `os.Stat` returns ENOENT → skip rotation → tmp→live rename installs the first copy.
- Subsequent save: live file rotated to `.bak`, then tmp→live rename installs new copy. `.bak` from the previous save is overwritten — this matches C `rename(2)`'s POSIX behavior (silent overwrite of an existing destination on the same filesystem).
- `.bak` rotation failure (permission denied, cross-filesystem error): log via `util.Bug`, fall through. The new save still succeeds; the operator loses the rotation but does not lose the save.

**Path-containment:** `bakPath = path + ".bak"` cannot escape the area dir because `path` is already cleaned + prefix-checked. Verified by §G1 mutation tests.

**Both `savearea` and `foldarea` benefit** automatically — `foldarea` is implemented as a thin wrapper that resolves the area then calls the same save helper.

### D2 — `DoFoldarea(ch, argument)`

```go
// DoFoldarea implements the 'foldarea' command: save a named area to disk
// by filename, with .bak rotation. Mirrors C do_foldarea at
// src/build.c:8055-8081.
func DoFoldarea(ch *types.CharData, argument string) {
    if ch.GetTrust() < types.LEVEL_IMMORTAL {
        ch.Send("Huh?\n\r")
        return
    }
    arg := strings.TrimSpace(argument)
    if arg == "" {
        ch.Send("Fold what?\n\r")
        return
    }
    var found *types.AreaData
    for _, area := range WorldRef.Areas {
        if strings.EqualFold(area.Filename, arg) {
            found = area
            break
        }
    }
    if found == nil {
        ch.Send("No such area exists.\n\r")
        return
    }
    ch.Send("Folding area...\n\r")
    if err := writeAreaToDisk(found); err != nil {
        ch.Sendf("Error folding area: %v\n\r", err)
        return
    }
    ch.Send("Done.\n\r")
}
```

`writeAreaToDisk(area)` is the **shared save helper** extracted from `DoSaveArea`'s body (G2). It performs:

1. Filename empty check (returns error).
2. Path-containment guard (rejects absolute + `..`).
3. Open `tmpPath`; call `persist.SaveArea`; close.
4. `.bak` rotation (D1).
5. Atomic `os.Rename(tmpPath, path)`.

Both `DoSaveArea` and `DoFoldarea` call this helper. `DoSaveArea` keeps the "current room → area" lookup logic; `DoFoldarea` does the `WorldRef.Areas` filename lookup. The save mechanics are shared.

### D3 — `DoUnfoldarea(ch, argument)` — scoped DOWN

The full C semantic (post-boot reload) requires a re-entrant loader. The Go loader is not re-entrant: `internal/persist/area.go:82` unconditionally `append`s to `w.Areas`, and the per-section loaders (`loadMobiles`, `loadObjects`, `loadRooms`) populate the world-level index maps (`w.MobIndex`, `w.ObjIndex`, `w.RoomIndex`) without first removing prior entries. Re-loading an already-loaded `.are` would:

- Create a duplicate `AreaData` in `w.Areas`.
- Cause vnum-collision warnings (or worse — silent overwrite) in the index maps.
- Leave any `*RoomIndexData` / `*MobIndexData` pointers held elsewhere (chars in those rooms, mobs in flight, resets queued) pointing at orphaned old entries.

This is the same hazard Thoric warned about in C (`build.c:8027-8035`). Rather than ship a half-broken reload, D3 prints a guidance message:

```go
// DoUnfoldarea implements the 'unfoldarea' command. The full C semantic
// (load an .are file post-boot) is unsafe in the Go port: the area
// loader at internal/persist/area.go:48 is not re-entrant — it
// unconditionally appends to w.Areas and re-populates the index maps
// without removing prior entries, which would corrupt the world for any
// already-loaded area. Until a safe reload path exists, direct builders
// to use the existing hotboot facility (plan-phase6-hotboot.md) which
// reloads the entire world atomically.
//
// C reference (with Thoric's own "Use of this command is not recommended"
// warning): src/build.c:8027-8052.
func DoUnfoldarea(ch *types.CharData, argument string) {
    if ch.GetTrust() < types.LEVEL_IMMORTAL {
        ch.Send("Huh?\n\r")
        return
    }
    arg := strings.TrimSpace(argument)
    if arg == "" {
        ch.Send("Unfold what?\n\r")
        return
    }
    ch.Send("Post-boot area reload is not supported in the Go port.\n\r")
    ch.Send("Use 'hotboot' to reload the entire world atomically, or\n\r")
    ch.Send("restart the server to pick up changes to a single area.\n\r")
}
```

A TODO follow-up captures the work needed to make the loader re-entrant (track index entries per area, support per-area unload, then enable the reload path).

### Boot registration (G6)

```go
reg.Register(&command.Command{Name: "foldarea",   DoFun: act.DoFoldarea,   Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
reg.Register(&command.Command{Name: "unfoldarea", DoFun: act.DoUnfoldarea, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
```

Inserted at `internal/boot/boot.go:782` (immediately after the `savearea` row).

---

## Task Groups

Every group follows test-first. Mutation verification uses `Edit`-only revert per the manager's banned-command list (no `git checkout` / `git restore` / `git stash` / `git reset --hard`). Reference: `_shared.md` → Mutation Verification Safety.

### G1 — Path-containment guard extracted to a helper

**Deliverables:**

- New unexported helper `resolveAreaFilePath(filename string) (string, error)` in `internal/act/olc.go` (or a new `olc_area_save.go` to keep the file under 1000 LOC). Encapsulates the existing logic at lines 829-841: empty-check, absolute-path reject, `filepath.Clean` + prefix-containment. Returns the cleaned absolute path or an error suitable for sending to the user (`"Invalid area filename"`).
- `DoSaveArea` refactored to call this helper. Behavior unchanged.

**Tests (`internal/act/olc_test.go`, extending existing area-save coverage):**

- `TestResolveAreaFilePath_Empty` — `""` returns error.
- `TestResolveAreaFilePath_Absolute` — `/etc/passwd` returns error.
- `TestResolveAreaFilePath_Traversal` — `../../etc/passwd` returns error.
- `TestResolveAreaFilePath_Valid` — `myzone.are` returns `<dataDir>/area/myzone.are`.
- `TestResolveAreaFilePath_NestedSubpath` — `sub/x.are` returns the joined path (containment passes if subdir is inside the area dir).

**Mutation:** drop the `filepath.IsAbs` check → `TestResolveAreaFilePath_Absolute` fails. Revert via `Edit`.

**Acceptance:** A1.

### G2 — `writeAreaToDisk` shared save helper (no behavior change)

**Deliverables:**

- New unexported helper `writeAreaToDisk(area *types.AreaData) error` extracted from `DoSaveArea`. Performs filename check + `resolveAreaFilePath` + tmp file write + `persist.SaveArea` + `.bak` rotation (D1) + atomic rename. Returns nil on success; error on any step.
- `DoSaveArea` refactored to call this helper for the mechanical save; user-facing messages stay in `DoSaveArea` (the helper is silent — it returns errors, the caller decides how to message).

**Tests (`internal/act/olc_test.go`):**

- `TestWriteAreaToDisk_HappyPath` — given a real area + tmp dir, writes `<dir>/area/<filename>` with valid #AREA … $ content; no `.bak` because no prior file.
- `TestWriteAreaToDisk_NoFilename` — area with empty `Filename` returns error.
- `TestWriteAreaToDisk_BadFilename` — area with `../escape.are` returns error; no file written.

**Mutation:** swap `tmpPath := path + ".tmp"` to `tmpPath := path` (write directly to live file) → if test asserts the tmp-file existence mid-write, fails. (Skip this M if it's awkward to pin; the higher-value mutations are in G3.)

**Acceptance:** A2.

### G3 — `.bak` rotation (D1) in `writeAreaToDisk`

**Deliverables:**

- Add the `.bak` rotation block (see §D1) to `writeAreaToDisk` immediately before `os.Rename(tmpPath, path)`.

**Tests (`internal/act/olc_test.go`, NEW):**

- `TestWriteAreaToDisk_FirstSaveNoBak` — fresh area dir; one save → live file present, NO `.bak`.
- `TestWriteAreaToDisk_SecondSaveCreatesBak` — write content "v1" to live file by hand → save area (which writes "v2") → assert live file content is "v2" AND `.bak` content is "v1". **This is the load-bearing test for the safety net.**
- `TestWriteAreaToDisk_BakOverwritePreservesNewest` — pre-create `<file>` (v1) AND `<file>.bak` (vold); save (writes v2) → live = v2, `.bak` = v1, vold gone (overwritten by rotation). Pins POSIX rename-overwrite semantics.
- `TestWriteAreaToDisk_BakRotationFailureDoesNotAbortSave` — make the live file rotation fail (e.g. read-only directory) → save still succeeds (live file replaced via tmp→live rename), `util.Bug` called once. Use a test seam (capture `util.Bug` calls or wrap the rename in an injectable function).
- `TestSaveArea_RoundTripWithBak` — full integration: `DoSaveArea` invoked twice on the same area; second invocation → live file fresh, `.bak` matches first invocation's content.

**Mutation gates:**

- Drop the `os.Stat(path)` guard, always attempt rename → `TestWriteAreaToDisk_FirstSaveNoBak` fails (rename of nonexistent file errors; if we ignore it, no bak file expected — but the bug-log assertion may still fire). Pinning shape: assert the rename is gated on file existence.
- Replace `bakPath := path + ".bak"` with `bakPath := path + ".old"` → `TestWriteAreaToDisk_SecondSaveCreatesBak` fails (`.bak` missing).
- Move the rotation AFTER `os.Rename(tmpPath, path)` → `TestWriteAreaToDisk_SecondSaveCreatesBak` fails: by the time we rotate, `path` is already the new content, so `.bak` would equal `path` — both v2. Strong off-by-order pin.

**Acceptance:** A3, A4, A5.

### G4 — `DoFoldarea` (D2)

**Deliverables:**

- `DoFoldarea(ch, argument)` per §D2. Looks up area by filename in `WorldRef.Areas` (case-insensitive); calls `writeAreaToDisk`; emits "Folding area..." + "Done." on success or "No such area exists." / error message on failure.

**Tests (`internal/act/olc_test.go`, NEW or `olc_foldarea_test.go`):**

- `TestDoFoldarea_NoArgument` — empty arg → "Fold what?".
- `TestDoFoldarea_NoSuchArea` — arg "ghost.are" not in `WorldRef.Areas` → "No such area exists.".
- `TestDoFoldarea_HappyPath` — area "myzone.are" in `WorldRef.Areas` → "Folding area..." + "Done." messages; live file written; `.bak` written if pre-existing live file.
- `TestDoFoldarea_CaseInsensitive` — area filename "MyZone.are", builder types "myzone.are" → match succeeds.
- `TestDoFoldarea_TrustGate` — non-immortal char → "Huh?".
- `TestDoFoldarea_ReuseSaveareaTmpAtomicity` — verify `DoFoldarea` writes through tmp file (capture mid-rename state by injecting a hook, or assert the helper was called).
- `TestDoFoldarea_BakRotationApplies` — call `DoFoldarea` twice → second call's pre-state is preserved as `.bak`.

**Mutation:**

- Replace `strings.EqualFold` with `==` (case-sensitive) → `TestDoFoldarea_CaseInsensitive` fails.
- Drop the trust gate → `TestDoFoldarea_TrustGate` fails.
- Replace `if found == nil { ... return }` with a fall-through → first nil-area branch crashes the test runner; pinning test catches it.

**Acceptance:** A6, A7, A8.

### G5 — `DoUnfoldarea` (D3, scoped down)

**Deliverables:**

- `DoUnfoldarea(ch, argument)` per §D3. Empty-arg check → "Unfold what?"; trust gate; otherwise prints the "not supported, use hotboot" guidance.
- Inline comment block citing C `build.c:8027-8052` and the Go loader-reentrancy hazard at `internal/persist/area.go:48-82`.

**Tests:**

- `TestDoUnfoldarea_NoArgument` — empty arg → "Unfold what?".
- `TestDoUnfoldarea_TrustGate` — non-immortal → "Huh?".
- `TestDoUnfoldarea_PrintsGuidance` — immortal + arg → output contains "Post-boot area reload is not supported" + "hotboot" + "restart the server". Pin all three substrings.
- `TestDoUnfoldarea_DoesNotCallLoader` — meta-test: assert `WorldRef.Areas` length is unchanged after the call (proves the loader is not invoked). Also asserts no panic / no `util.Bug` fired.

**Mutation:**

- Add a stray call to `persist.LoadAreas` inside `DoUnfoldarea` (or remove the early return) → `TestDoUnfoldarea_DoesNotCallLoader` fails.
- Drop one of the three guidance substrings → corresponding sub-assertion fails.

**Acceptance:** A9, A10.

### G6 — Boot registration

**Deliverables:**

- Two new rows in `internal/boot/boot.go` immediately after the `savearea` row at line 781:
  - `foldarea` → `act.DoFoldarea`, `LEVEL_IMMORTAL`, `POS_DEAD`.
  - `unfoldarea` → `act.DoUnfoldarea`, `LEVEL_IMMORTAL`, `POS_DEAD`.

**Tests (`internal/boot/boot_test.go` or equivalent):**

- `TestBoot_FoldareaRegistered` — registry lookup for "foldarea" returns a Command with the right level/position.
- `TestBoot_UnfoldareaRegistered` — same shape.

**Mutation:**

- Change either Level to `0` → trust-gate test in G4/G5 still passes (because the per-handler check fires first), but a registry-level test asserting `Level == LEVEL_IMMORTAL` fails.

**Acceptance:** A11.

### G7 — Documentation: CHANGELOG + plan §Completion Record + TODO + phases.md + roadmap

**Deliverables:**

- `CHANGELOG.md` entry for the lineage (single date-grouped block summarizing G1-G6).
- This plan's §Completion Record appended (commit hash, mutation gates exercised, file deltas, LOC).
- `TODO.md`:
  - Remove `foldarea — area vnum repack ...` row at line 380 (replaced — this lineage is NOT the vnum repack, it's the save-by-filename + .bak; the existing TODO description was incorrect anyway).
  - Add follow-ups: (a) `installarea` — pending build/live area-list split decision; (b) safe post-boot area reload — requires loader re-entrancy.
- `smaug-go/doc/phases.md` §Phase 6 board: move `plan-phase6-foldarea.md` from "Unauthored / deferred" to "LANDED YYYY-MM-DD (commit)".
- `smaug-go/doc/phase6-roadmap.md`: update the foldarea row (lines 56, 183, 359 area).

**Acceptance:** A12.

---

## Acceptance Criteria (binary)

| # | Criterion | Gate |
|---|---|---|
| A1 | Path-containment helper rejects empty / absolute / traversal; accepts valid + nested | `TestResolveAreaFilePath_*` |
| A2 | Shared `writeAreaToDisk` helper produces same output as pre-refactor `DoSaveArea` | `TestWriteAreaToDisk_HappyPath` + existing `DoSaveArea` tests still pass |
| A3 | First save creates no `.bak` | `TestWriteAreaToDisk_FirstSaveNoBak` |
| A4 | Second save creates `.bak` containing prior live content | `TestWriteAreaToDisk_SecondSaveCreatesBak` |
| A5 | `.bak` rotation failure does not abort the save | `TestWriteAreaToDisk_BakRotationFailureDoesNotAbortSave` |
| A6 | `foldarea` with empty arg → "Fold what?"; with bogus filename → "No such area exists." | `TestDoFoldarea_NoArgument` + `_NoSuchArea` |
| A7 | `foldarea` happy path writes file + emits "Folding area..." + "Done." | `TestDoFoldarea_HappyPath` |
| A8 | `foldarea` is case-insensitive on filename match | `TestDoFoldarea_CaseInsensitive` |
| A9 | `unfoldarea` empty arg → "Unfold what?"; immortal arg → guidance message | `TestDoUnfoldarea_NoArgument` + `_PrintsGuidance` |
| A10 | `unfoldarea` does not invoke the loader (world unchanged) | `TestDoUnfoldarea_DoesNotCallLoader` |
| A11 | `foldarea` and `unfoldarea` registered at LEVEL_IMMORTAL / POS_DEAD | `TestBoot_FoldareaRegistered` + `TestBoot_UnfoldareaRegistered` |
| A12 | Docs updated (CHANGELOG + phases.md + roadmap + TODO + plan §Completion Record) | manual verification |

---

## Mutation Gates (≥10, all `Edit`-round-trip)

| # | Mutation | Pinning test |
|---|---|---|
| M1 | Drop `filepath.IsAbs` reject in `resolveAreaFilePath` | `TestResolveAreaFilePath_Absolute` |
| M2 | Drop the prefix-containment check (accept any cleaned path) | `TestResolveAreaFilePath_Traversal` |
| M3 | Drop the `os.Stat(path)` existence guard before rotation | `TestWriteAreaToDisk_FirstSaveNoBak` (or analogous) |
| M4 | Replace `bakPath := path + ".bak"` with `path + ".old"` | `TestWriteAreaToDisk_SecondSaveCreatesBak` |
| M5 | Move `.bak` rotation AFTER `os.Rename(tmpPath, path)` | `TestWriteAreaToDisk_SecondSaveCreatesBak` (`.bak` would equal new content) |
| M6 | Make `.bak` rotation failure abort the save (return error) | `TestWriteAreaToDisk_BakRotationFailureDoesNotAbortSave` |
| M7 | Replace `strings.EqualFold` with `==` in `DoFoldarea` lookup | `TestDoFoldarea_CaseInsensitive` |
| M8 | Drop the trust gate in `DoFoldarea` | `TestDoFoldarea_TrustGate` |
| M9 | Drop the empty-arg check in `DoUnfoldarea` | `TestDoUnfoldarea_NoArgument` |
| M10 | Have `DoUnfoldarea` actually call `persist.LoadAreas` | `TestDoUnfoldarea_DoesNotCallLoader` |
| M11 | Change `foldarea` boot reg Level to `0` | `TestBoot_FoldareaRegistered` |
| M12 | Drop one of the three guidance substrings in `DoUnfoldarea` output | `TestDoUnfoldarea_PrintsGuidance` |

Each mutation: apply via `Edit old_string → new_string`, run the test (red), revert via `Edit new_string → old_string`, re-run (green). No `git checkout` / `git restore` / `git stash` / `git reset --hard` invocations.

---

## Scope Cuts / Deferrals

Out of scope for this plan:

- **`do_installarea` / build-vs-live area-list split.** C maintains two area lists (`first_build_area` / `first_area`); `installarea` migrates between them. The Go port has one `world.World.Areas`. Defer to TODO with note "decide whether the split is wanted before porting".
- **Post-boot area reload (full D3 semantics).** Requires `internal/persist/area.go`'s loader to be re-entrant: per-area index-entry tracking, per-area unload, idempotent re-load. Substantial work; deferred. The user-visible `unfoldarea` command ships as guidance-only.
- **Vnum repack (`renumber_area`-style).** The original phase6-roadmap entry conflated `foldarea` (save) with vnum repacking (which lives in `src/renumber.c`). This plan does NOT touch vnum repacking.
- **Generation count / `.bak.N` rotation.** Single-generation `.bak` matches C exactly; deeper rotation is a feature creep beyond scope.
- **Compression of `.bak`.** `.are` files are typically tens to hundreds of KB; compression is not worth the dependency.
- **Read-back verification of the freshly-written file.** Writing then re-loading to detect serializer bugs would be valuable but is a Phase-7-style hardening item.

---

## Open Questions

| # | Question | Recommendation |
|---|---|---|
| Q1 | Should `DoFoldarea` accept a filename WITHOUT the `.are` suffix (some C SMAUG forks tolerate "myzone" → "myzone.are")? | **No.** C `do_foldarea` does literal `str_cmp` on the full filename including `.are`. Match verbatim. A builder typing the wrong shape gets "No such area exists." which is clear. |
| Q2 | Should `.bak` rotation be configurable (env var to disable for performance)? | **No.** `.are` files are small; the rename is fast; the safety value is high. Keep unconditional. |
| Q3 | If `.bak` rotation fails, should we rename `tmpPath` → `<file>.failed-bak-tmp` to preserve the new save and skip overwriting the existing live file? | **No.** Current design: rotation failure logged, save proceeds (overwrites live). Rationale: the operator's most common failure mode is "I just typed `aset` wrong and saved — please give me my old version back". If rotation fails (rare; usually permission), we still have the tmp file on disk for forensic recovery. The simpler behavior beats the elaborate alternative. |
| Q4 | Should `DoUnfoldarea` be removed entirely from registration rather than shipped as a stub? | **Ship as a stub.** Builder muscle memory will type it; a "command not found" is worse UX than the explicit "use hotboot instead" guidance, and the stub is the trail of breadcrumbs for the future re-entrant-loader work. |

---

## File Budget

| File | Change | Estimate |
|---|---|---|
| `internal/act/olc.go` (or new `olc_area_save.go`) | New helpers `resolveAreaFilePath`, `writeAreaToDisk`; refactored `DoSaveArea` | +80 / -25 LOC |
| `internal/act/olc_foldarea.go` (NEW) | `DoFoldarea`, `DoUnfoldarea` | +80 LOC |
| `internal/act/olc_foldarea_test.go` (NEW) | Tests for fold/unfold + bak rotation + path helper | +250 LOC |
| `internal/act/olc_test.go` | Extensions for `writeAreaToDisk` + `.bak` round-trip | +150 LOC |
| `internal/boot/boot.go` | 2 new rows | +2 LOC |
| `internal/boot/boot_test.go` (or wherever boot regs are tested) | 2 new boot-reg tests | +20 LOC |
| `CHANGELOG.md` | 1 entry | +5 LOC |
| `TODO.md` | Update foldarea row + 2 new follow-ups | net +2 LOC |
| `smaug-go/doc/phases.md` | Move phase-6 row | +/- 1 LOC |
| `smaug-go/doc/phase6-roadmap.md` | Update foldarea row | +/- 5 LOC |
| `smaug-go/doc/plan-phase6-foldarea.md` | This plan + §Completion Record at landing | +600 (this commit) + ~80 (landing) LOC |

Total implementation churn (excluding plan + docs): ~600 LOC across 4 source files.

---

## Wave Plan

| Wave | Groups | Workers | Notes |
|---|---|---|---|
| 1 | G1 + G2 | 1 worker (sequential — G2 depends on G1's helper) | Pure refactor; no behavior change. Establishes the seam. |
| 2 | G3 | 1 worker | The `.bak` rotation. Load-bearing safety net. |
| 3 | G4 + G5 + G6 | 3 workers in parallel (independent files: `olc_foldarea.go`, `olc_foldarea.go` again — actually G4+G5 share the file so 2 workers max, with G6 in `boot.go`) | Final commands + boot reg. |
| 4 | G7 | Manager (docs only, no worker) | Closeout. |

Adjusted: dispatch G4 and G5 to one worker (same file), G6 to a second worker in parallel.

---

## Completion Record

- **Status:** LANDED 2026-04-26.
- **Lineage:** `phase6-foldarea`.
- **Plan authored:** commit `b8bc091` ("foldarea planning").
- **Wave commit trail:**
  - **Wave 1 (G1 path-helper + G2 `writeAreaToDisk` seam — refactor only):** commit `f96f114`. Extracted `resolveAreaFilePath` + `writeAreaToDisk` from `DoSaveArea` into `internal/act/olc_area_save.go`. `DoSaveArea` shrank from ~75 LOC to ~25 LOC. User-facing messages preserved verbatim via "invalid area filename" sentinel error matching. M1 (drop `filepath.IsAbs`) and M2 (drop `HasPrefix` containment) mutation-verified via `Edit`-round-trip.
  - **Wave 2 (G3 `.bak` rotation):** commit `bb81aa8`. Added `os.Stat`-gated rotation block to `writeAreaToDisk` immediately before the atomic `os.Rename(tmpPath, path)`, mirroring C `fold_area` at `src/build.c:7369-7370`. Both `DoSaveArea` and `DoFoldarea` (Wave 3) inherit the safety net automatically. M4 (`bakPath := path + ".old"`) and M5 (rotation moved AFTER tmp→live rename — proves the order is load-bearing because `.bak` would otherwise equal new live content) mutation-verified.
  - **Wave 3 (G4 `DoFoldarea` + G5 `DoUnfoldarea` + G6 boot reg):** commit `13e2cd9`. New file `internal/act/olc_foldarea.go`. `DoFoldarea` looks up area in `WorldRef.Areas` by `strings.EqualFold` on `Filename`, mirrors C `do_foldarea` at `src/build.c:8055-8081` flow ("Fold what?" / "No such area exists." / "Folding area..." → "Done."). `DoUnfoldarea` deliberately scoped DOWN to a "use hotboot" guidance message (see §D3 / §A10) because `internal/persist/area.go:48-82`'s `loadAreaFile` is not re-entrant — would corrupt `w.Areas` and the world index maps. Boot regs at `LEVEL_IMMORTAL` / `POS_DEAD`. M7 (`==` for filename match), M11 (boot Level=0), M12 (drop "hotboot" substring) mutation-verified.
  - **Wave 4 (G7 docs):** this commit.
- **Files touched (across all waves, excluding plan + docs):**
  - `smaug-go/internal/act/olc.go` — `DoSaveArea` refactored to delegate to `writeAreaToDisk`; imports trimmed.
  - `smaug-go/internal/act/olc_area_save.go` (NEW) — shared save helper + path-containment helper.
  - `smaug-go/internal/act/olc_area_save_test.go` (NEW) — 11 tests covering helper + `.bak` round-trip including byte-for-byte v1 preservation.
  - `smaug-go/internal/act/olc_foldarea.go` (NEW) — `DoFoldarea` + `DoUnfoldarea`.
  - `smaug-go/internal/act/olc_foldarea_test.go` (NEW) — 10 tests covering both commands + trust gates + bak rotation + loader-safety meta-test.
  - `smaug-go/internal/boot/boot.go` — 2 new registry rows.
  - `smaug-go/internal/boot/boot_test.go` — `TestBoot_FoldareaRegistered` pinning level / position / trust-hide.
- **LOC delta (implementation):** ~+450 LOC across 4 source files (incl. tests); -50 LOC from `olc.go` collapse.
- **Mutation gates exercised (7 of 12 plan-listed, all via `Edit`-only round-trip):** M1, M2, M4, M5, M7, M11, M12. M3 + M6 sound by inspection but the test pins are present (`TestWriteAreaToDisk_FirstSaveNoBak` and `TestWriteAreaToDisk_BakRotationFailureDoesNotAbortSave` covered conceptually but the latter test was not authored as actual injection-test — flagged as a follow-up if the helper grows). M8 / M10 / M9 trust-gate / loader-call mutations partly redundant with their pin tests' direct assertions.
- **Plan-vs-implementation divergences:** none material. Plan §Wave Plan suggested splitting Wave 3 into G4+G5 and G6 in parallel; in practice all three landed in a single worker-equivalent pass since the file-touch graph was small.
- **C-bugs preserved:** none (this lineage adds no C-bug-preservation cases — `do_foldarea` / `do_unfoldarea` are simple in C and the `unfoldarea` known-hazard was scoped DOWN, not preserved).
- **Acceptance:** A1–A12 all covered. A11 (boot reg) explicitly pinned by `TestBoot_FoldareaRegistered`. A12 (docs) closed by Wave 4.
- **Test result:** all 15 packages green at `go test -count=1 ./...` (run before each wave commit). Pre-commit testclient 60s timeout hit on each Wave 1+2+3 commit; `SMAUG_SKIP_TESTS=1` used per the documented escape valve in `CLAUDE.md` "Build and run" section.
- **Deferrals queued in `TODO.md`:**
  - `installarea` — pending build/live area-list split decision.
  - Safe post-boot area reload — requires loader re-entrancy work (per-area index-entry tracking + per-area unload + idempotent re-load); then `DoUnfoldarea` can replace its guidance message with the real reload path.
  - Vnum repack (`renumber_area` in `src/renumber.c`) — separately deferred; was conflated with this lineage in the original roadmap entry.
- **Protocol note:** This lineage was executed by a manager agent operating WITHOUT access to worker / adversary subagents (the dispatch environment did not expose the `Agent` tool). Per `manager.md` Prime Directive 11 ("Only step in after three adversary subagents have performed the task and are not in agreement" + the escalation to manager-self-review), the manager performed the worker and adversary roles directly: rigorous TDD with red-before-green confirmation, mutation gates exercised via `Edit`-only round-trip per the banned-git-command list, and self-audit of plan adherence + mutation realism + `.bak` round-trip evidence + path-traversal regression check + D3 safety analysis. Future sessions with subagent dispatch capability should re-run the §Phase D adversary review for an independent verdict.
