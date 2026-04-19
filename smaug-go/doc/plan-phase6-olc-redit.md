# Plan: Phase 6 Interactive OLC — `CON_REDIT` Substate (Menu-Driven Room Editor)

**Status:** Authored 2026-04-18. Pending adversary review before dispatch.
**Priority:** P2 (Phase 6, Wave D — Interactive OLC, first of three editors).
**Scope:** Interactive room-editor nanny substate `CON_REDIT`. Adds a menu-driven room editor on top of the existing flat `redit <sub> <args>` command (Phase 4b). Proves the nanny-dispatch pattern that `CON_OEDIT` / `CON_MEDIT` will inherit.

---

## Problem

The constants `CON_REDIT`, `CON_OEDIT`, `CON_MEDIT` are defined in `internal/types/enums.go:89-91` (iota values 21/22/23) but **no code ever reads them**. The pulse-loop dispatch switch in `internal/game/loop.go:252-267` handles `CON_PLAYING` and `CON_EDITING` and falls through to `g.nanny(d, line)` in `default:` — there is no case for `CON_REDIT`, so input typed while in the redit menu never reaches a menu handler.

The flat command path `DoRedit` (`internal/act/olc.go:32-221`) implements 17 subcommands (`name`, `desc`, `sector`, `flags`, `exdesc`, `ed`, `rmed`, `exit`, `bexit`, `exflags`, `exname`, `exkey`, `teledelay`, `televnum`, `tunnel`, `rlist`, plus dispatch defaults). It does **not** enter a menu substate — each invocation is one-shot. A builder who wants to change multiple fields re-types `redit ...` each time.

The C flow is the opposite. `do_oredit` at `src/oredit.c:79-136` sets `d->connected = CON_REDIT`, stashes the room on `d->character->dest_buf`, and calls `redit_disp_menu`. Every subsequent line arrives at `redit_parse` (`src/oredit.c:541-1035`) and is dispatched against `OLC_MODE(d)` — a per-descriptor menu-state enum (`REDIT_MAIN_MENU`, `REDIT_NAME`, `REDIT_DESC`, `REDIT_FLAGS`, `REDIT_SECTOR`, `REDIT_TUNNEL`, `REDIT_TELEDELAY`, `REDIT_TELEVNUM`, `REDIT_EXIT_MENU`, `REDIT_EXIT_EDIT`, `REDIT_EXIT_FLAGS`, `REDIT_EXIT_ADD`, `REDIT_EXIT_ADD_VNUM`, `REDIT_EXIT_DELETE`, `REDIT_EXIT_VNUM`, `REDIT_EXIT_KEY`, `REDIT_EXIT_KEYWORD`, `REDIT_EXIT_DESC`, `REDIT_EXTRADESC_MENU`, `REDIT_EXTRADESC_CHOICE`, `REDIT_EXTRADESC_KEY`, `REDIT_EXTRADESC_DESCRIPTION`, `REDIT_EXTRADESC_DELETE`, `REDIT_CONFIRM_SAVESTRING`).

C dispatch at `src/smaug.c:1641-1652` routes `d->connected == CON_REDIT` to `redit_parse(d, argument)`. The nanny is NOT involved; CON_REDIT short-circuits it.

Consequence of the Go gap: no interactive editor exists. Builders use the flat commands or nothing. This plan adds the menu-driven editor while keeping the flat path.

## C Reference (authoritative)

All line numbers verified 2026-04-18 from `src/oredit.c` (1036 LOC).

### Entry points

- **`do_oredit`** — `src/oredit.c:79-136`. Parses optional vnum, checks `can_rmodify`, guards against double-edit (`d->connected == CON_REDIT && OLC_VNUM(d) == room->vnum`), allocates `OLC_DATA`, sets `d->connected = CON_REDIT`, `d->character->dest_buf = room`, calls `redit_disp_menu(d)`.
- **`do_redit_reset`** — `src/oredit.c:493-535`. The `last_cmd` callback invoked by `edit_buffer` on `/s`. Two substates: `SUB_ROOM_DESC` (updates `room->description`, returns to `REDIT_MAIN_MENU`), `SUB_ROOM_EXTRA` (updates extradesc, returns to `REDIT_EXTRADESC_CHOICE`). Always sets `d->connected = CON_REDIT` after `stop_editing`.
- **`redit_parse`** — `src/oredit.c:541-1035`. Top-level state machine keyed on `OLC_MODE(d)`. Every case mutates the room and either returns (stay in same mode) or `break`s (falls to the end, sets `OLC_CHANGE(d) = TRUE`, re-displays `redit_disp_menu(d)`).
- **C dispatch call site** — `src/smaug.c:1641-1652` (must verify when writing). Routes input based on `d->connected`: CON_PLAYING → command interpreter; CON_EDITING → `edit_buffer`; CON_REDIT/CON_OEDIT/CON_MEDIT → `redit_parse`/`oedit_parse`/`medit_parse`; default → `nanny`.

### Menu rendering

- **`redit_disp_menu`** — `src/oredit.c:423-477`. The main room menu. Lists vnum, area name, name, description, room flags (via `ext_flag_string`), sector name, tunnel, tele_delay, tele_vnum; options 1-7 per field, A (exit menu), B (extradesc menu), Q (quit). Sets `OLC_MODE(d) = REDIT_MAIN_MENU`.
- **`redit_disp_flag_menu`** — `src/oredit.c:366-389`. Two-column flag list (`r_flags`), current flags shown, prompt `"Enter room flags, 0 to quit"`. Sets `OLC_MODE(d) = REDIT_FLAGS`.
- **`redit_disp_sector_menu`** — `src/oredit.c:399-420`. Sector type list (`sector_names`), skips `SECT_DUNNO`. Sets `OLC_MODE(d) = REDIT_SECTOR`.
- **`redit_disp_exit_menu`** — `src/oredit.c:260-291`. Lists each exit (direction, target vnum, key, flags, keywords); A/R/Q options. Sets `OLC_MODE(d) = REDIT_EXIT_MENU`.
- **`redit_disp_exit_edit`** — `src/oredit.c:293-320`. Per-exit field menu (direction, to-vnum, key, keyword, flags, description); Q to return. Sets `OLC_MODE(d) = REDIT_EXIT_EDIT`.
- **`redit_disp_exit_flag_menu`** — `src/oredit.c:337-363`. Two-column exit-flag list (`ex_flags`, skipping reserved bits). Sets `OLC_MODE(d) = REDIT_EXIT_FLAGS`.
- **`redit_disp_exit_dirs`** — `src/oredit.c:322-334`. Direction list 0..`DIR_SOMEWHERE`.
- **`redit_disp_extradesc_menu`** — `src/oredit.c:234-257`. Lists existing extradescs, A/R/Q options. Sets `OLC_MODE(d) = REDIT_EXTRADESC_MENU`.
- **`redit_disp_extradesc_prompt_menu`** — `src/oredit.c:219-232`. Secondary prompt "Which extra description do you want to edit?"
- **`oedit_disp_extra_choice`** — `src/ooedit.c` (referenced here). Per-extradesc sub-menu with keywords and description.

### Field parsers (from `redit_parse`)

| OLC_MODE | Input | Action |
|---|---|---|
| `REDIT_MAIN_MENU` | `1` | Prompt for name; `OLC_MODE = REDIT_NAME` |
| `REDIT_MAIN_MENU` | `2` | Set `substate=SUB_ROOM_DESC`, `last_cmd=do_redit_reset`, call `start_editing(room->description)` |
| `REDIT_MAIN_MENU` | `3` | `redit_disp_flag_menu` |
| `REDIT_MAIN_MENU` | `4` | `redit_disp_sector_menu` |
| `REDIT_MAIN_MENU` | `5` | Prompt for tunnel |
| `REDIT_MAIN_MENU` | `6` | Prompt for teledelay |
| `REDIT_MAIN_MENU` | `7` | Prompt for tele_vnum |
| `REDIT_MAIN_MENU` | `A/a` | `redit_disp_exit_menu` |
| `REDIT_MAIN_MENU` | `B/b` | `redit_disp_extradesc_menu` |
| `REDIT_MAIN_MENU` | `Q/q` | `cleanup_olc(d)` — sets `d->connected = CON_PLAYING` |
| `REDIT_NAME` | any | `STRALLOC` into `room->name`, log, redisplay menu |
| `REDIT_FLAGS` | number 0 | redisplay menu (break) |
| `REDIT_FLAGS` | number 1..32 | toggle flag bit, re-display flag menu |
| `REDIT_FLAGS` | word | `get_rflag(word)` → toggle, re-display flag menu |
| `REDIT_SECTOR` | 0..SECT_MAX | set `room->sector_type` |
| `REDIT_TUNNEL` | number | `URANGE(0, n, 1000)` into `room->tunnel` |
| `REDIT_TELEDELAY` | number | `room->tele_delay = n` |
| `REDIT_TELEVNUM` | number | `URANGE(1, n, MAX_VNUM)` into `room->tele_vnum` |
| `REDIT_EXIT_MENU` | `A` | `REDIT_EXIT_ADD`, show dir list |
| `REDIT_EXIT_MENU` | `R` | `REDIT_EXIT_DELETE`, "Delete which exit?" |
| `REDIT_EXIT_MENU` | `Q` | null `spare_ptr`, fall through → redisplay main |
| `REDIT_EXIT_MENU` | number | Look up by index, stash on `spare_ptr`, `redit_disp_exit_edit` |
| `REDIT_EXIT_EDIT` | 1 | Message "can only be changed by remaking" |
| `REDIT_EXIT_EDIT` | 2 | Prompt for to-vnum → `REDIT_EXIT_VNUM` |
| `REDIT_EXIT_EDIT` | 3 | Prompt for key vnum → `REDIT_EXIT_KEY` |
| `REDIT_EXIT_EDIT` | 4 | Prompt for keyword → `REDIT_EXIT_KEYWORD` |
| `REDIT_EXIT_EDIT` | 5 | `redit_disp_exit_flag_menu` |
| `REDIT_EXIT_EDIT` | 6 | Prompt for description → `REDIT_EXIT_DESC` |
| `REDIT_EXIT_EDIT` | Q | null `spare_ptr`, redisplay exit menu |
| `REDIT_EXIT_ADD` | number or word | Validate direction, stash on `tempnum`, prompt for vnum → `REDIT_EXIT_ADD_VNUM` |
| `REDIT_EXIT_ADD_VNUM` | number | `get_room_index`, `make_exit`, initialize exit, `redit_disp_exit_edit` |
| `REDIT_EXIT_DELETE` | number | `get_exit_num`, `extract_exit` |
| `REDIT_EXIT_VNUM` | number | Set `pexit->vnum` (and implicitly `to_room`? — verify — C assigns vnum only, to_room stays old) |
| `REDIT_EXIT_KEYWORD` | string | `STRALLOC` into `pexit->keyword` |
| `REDIT_EXIT_KEY` | number | `pexit->key = n` |
| `REDIT_EXIT_DESC` | string | Set `pexit->description` (one-liner; no editor) |
| `REDIT_EXIT_FLAGS` | 0 | redisplay exit edit |
| `REDIT_EXIT_FLAGS` | number | `TOGGLE_BIT(pexit->exit_info, 1 << (n-1))`, re-display flag menu |
| `REDIT_EXTRADESC_MENU` | A | Create new ed, link, `oedit_disp_extra_choice`, `REDIT_EXTRADESC_CHOICE` |
| `REDIT_EXTRADESC_MENU` | R | `REDIT_EXTRADESC_DELETE`, prompt "Delete which" |
| `REDIT_EXTRADESC_MENU` | number | Find ed by index, `oedit_disp_extra_choice` |
| `REDIT_EXTRADESC_MENU` | Q | fall through → main redisplay |
| `REDIT_EXTRADESC_CHOICE` | 1 | Prompt "Keywords" → `REDIT_EXTRADESC_KEY` |
| `REDIT_EXTRADESC_CHOICE` | 2 | Set `substate=SUB_ROOM_EXTRA`, `last_cmd=do_redit_reset`, `start_editing(ed->description)` |
| `REDIT_EXTRADESC_CHOICE` | Q | Check ed has keyword+desc, junk if not; redisplay extradesc menu |
| `REDIT_EXTRADESC_KEY` | string | `STRALLOC` into `ed->keyword`, back to `REDIT_EXTRADESC_CHOICE` |
| `REDIT_EXTRADESC_DELETE` | number | Unlink from extradesc chain, redisplay |
| `REDIT_CONFIRM_SAVESTRING` | Y | log, `cleanup_olc`, "Room saved to memory" |
| `REDIT_CONFIRM_SAVESTRING` | N | `cleanup_olc` |

### Other referenced C functions

- `cleanup_olc(d)` — disposes `d->olc`, sets `d->connected = CON_PLAYING`, nulls `d->character->dest_buf`. Not in `oredit.c`; lives in `src/olc.c`.
- `olc_log(d, fmt, ...)` — at `src/oredit.c:171-210`. Builds structured log line, calls `log_string_plus` with `LOG_BUILD`. Go equivalent would funnel through `util.LogBuild` (new) or `util.Bug`-adjacent path.
- `is_inolc(d)` — at `src/oredit.c:143-165`. Returns TRUE if `d->connected` is any of `CON_*EDIT`. Useful as a Go helper.

## Go Current State

Verified 2026-04-18.

- `internal/types/enums.go:89-91` — `CON_REDIT`, `CON_OEDIT`, `CON_MEDIT` are DEFINED (iota 21/22/23). No code reads them.
- `internal/game/loop.go:252-267` — pulse-loop dispatch switch has `case CON_PLAYING`, `case CON_EDITING`, and `default: g.nanny(d, line)`. No CON_REDIT arm.
- `internal/types/descriptor.go` — `DescriptorData` has no `OLC` field. No `OlcMode`, `OlcVnum`, `OlcChange` fields. No `DestBuf` stash.
- `internal/types/character.go` — `CharData` has `Substate int`, `InterEditing string`, `InterEditingVnum int`, `EditorSave func(*CharData)` (landed Tier 12). No `DestBuf`, no `SparePtr`, no `TempNum`, no `LastCmd`. The `EditorSave` callback is the Go moral equivalent of C's `last_cmd` trampoline.
- `internal/act/olc.go:32-221` — `DoRedit` implements 17 flat subcommands. The `desc` and `ed` subcommands already go through `StartEditingFunc` with `EditorSave` closures that capture the target pointer. This pattern is the direct template for the menu-driven path.
- `internal/act/olc.go:20-28` — `CopyBufferFunc` / `StopEditingFunc` / `StartEditingFunc` seams wire from `boot.go` (`internal/boot/boot.go:133-137`). Kept circular-dep-free.
- `internal/game/editor.go` — `/s` handler transitions `d.Connected` to `CON_PLAYING` unconditionally (landed Tier 12). For the interactive redit case the callback must *override* that transition back to `CON_REDIT` after the save. Mirrors C `do_redit_reset:514` (`ch->desc->connected = CON_REDIT`).
- `internal/types/enums.go:97+` — `SUB_ROOM_DESC` and `SUB_ROOM_EXTRA` exist.
- `internal/types/descriptor.go` — `ScrLen int` already present (for screen-clear optimizations, if later desired).

### Gaps this plan must fill

1. `OlcData` struct on `DescriptorData` (or inline fields) tracking `Mode int`, `Vnum int`, `Change bool`, plus `Target any` stash (moral equivalent of `dest_buf`/`spare_ptr`/`tempnum`).
2. `CON_REDIT` arm in `loop.go` dispatch switch, routing to a new `ReditParse(d, line)` function.
3. `ReditParse` state-machine dispatcher.
4. Menu-display helpers (`reditDispMenu`, `reditDispFlagMenu`, `reditDispSectorMenu`, `reditDispExitMenu`, `reditDispExitEdit`, `reditDispExitFlagMenu`, `reditDispExitDirs`, `reditDispExtradescMenu`, `reditDispExtradescChoice`).
5. A new entry command (`redit` invoked with no args or with just a vnum) that sets `d.Connected = CON_REDIT` and calls `reditDispMenu`. The existing flat-subcommand path should remain as-is when an arg like `name`/`desc` is provided.
6. REDIT_MODE constants (`REDIT_MAIN_MENU`, ...).
7. `EditorSave` callbacks for menu-driven `/s` that restore `d.Connected = CON_REDIT` after `StopEditing` (override the default `CON_PLAYING` transition).
8. An `olcLog` helper (small, routes through `util.Log` with a `"BUILD"` tag).

## Go Design

### Chosen approach

**Approach A (CHOSEN): Add `OlcData` struct alongside `DescriptorData`, add a single-arm switch case in `loop.go`, dispatch to a new `game.ReditParse` function.** Matches C shape closely; minimizes cross-cutting changes; proves the pattern for future `oedit`/`medit`.

Rationale for structure choices:

- **`OlcData` as a separate struct** (not inline fields on `DescriptorData`) because future `oedit`/`medit` will share it — one allocation per descriptor, nil when no OLC in progress (matches C `d->olc` pointer).
- **`ReditParse` lives in `internal/game/`** (not `internal/act/`) because it needs to mutate `d.Connected`, which is a descriptor-level concern best kept in the game loop package. It calls helper functions in `internal/act/` for actual field mutation where possible (reuse `DoRedit`'s subcommand logic).
- **Menu-display functions live in `internal/game/redit_menu.go`** (new file) to keep `act/olc.go` focused on the flat subcommand API.
- **Entry command stays in `act/olc.go` as `DoRedit`**, extended so that calling `redit` with no subcommand (or with a numeric vnum + no subcommand) enters the menu substate rather than printing the usage line. This preserves the flat path for builders who prefer it.

### Alternatives considered and rejected

- **Approach B: Single-file monolithic `redit_parse` in `internal/act/`.** Rejected — forces `act/` to mutate `d.Connected`, breaking the layering that keeps `act/` free of network concerns.
- **Approach C: Extend `EditBuffer` / `CON_EDITING` to gate on `Substate`.** Rejected — CON_EDITING is the text-editor state, not the menu state. Conflating them would break `/a` / `/s` semantics.
- **Approach D: One mega-state-table with function pointers.** Go-idiomatic but overkill for a 24-mode state machine; plain switch is clearer and easier to adversary-review.
- **Approach E: Rewrite flat `DoRedit` to be a thin wrapper around the menu.** Rejected — flat commands are faster for experienced builders; keeping both preserves UX.

### New types

```go
// internal/types/olc.go (new file)
package types

// OlcData tracks interactive-OLC session state for a descriptor.
// Moral equivalent of C's OLC_DATA struct (src/olc.h).
// Populated when d.Connected transitions to CON_REDIT / CON_OEDIT / CON_MEDIT;
// nil otherwise.
type OlcData struct {
    Mode   int         // REDIT_MAIN_MENU, REDIT_NAME, ... (per-editor enum space)
    Vnum   int         // vnum of the room/obj/mob being edited
    Change bool        // "something has been modified" — reserved for save-confirm path
    Target any         // the room/obj/mob pointer (type-asserted by parser)
    Spare  any         // secondary stash (exit pointer, extradesc pointer, etc. — matches C spare_ptr)
    TempNum int        // scratch integer (direction during exit-add, etc. — matches C tempnum)
}
```

Add to `DescriptorData`:

```go
// internal/types/descriptor.go
Olc *OlcData          // interactive-OLC state; nil when not in CON_REDIT/OEDIT/MEDIT
```

Add REDIT_ mode constants in `internal/types/enums.go` (new block; does not disturb existing iota chains because it's a new `const` block):

```go
// REDIT_* modes for OlcData.Mode when d.Connected == CON_REDIT.
// Mirror C src/olc.h REDIT_* enum.
const (
    REDIT_MAIN_MENU = iota + 100  // offset so OEDIT/MEDIT get later ranges
    REDIT_NAME
    REDIT_DESC
    REDIT_FLAGS
    REDIT_SECTOR
    REDIT_TUNNEL
    REDIT_TELEDELAY
    REDIT_TELEVNUM
    REDIT_EXIT_MENU
    REDIT_EXIT_EDIT
    REDIT_EXIT_ADD
    REDIT_EXIT_ADD_VNUM
    REDIT_EXIT_DELETE
    REDIT_EXIT_VNUM
    REDIT_EXIT_KEY
    REDIT_EXIT_KEYWORD
    REDIT_EXIT_DESC
    REDIT_EXIT_FLAGS
    REDIT_EXTRADESC_MENU
    REDIT_EXTRADESC_CHOICE
    REDIT_EXTRADESC_KEY
    REDIT_EXTRADESC_DESCRIPTION
    REDIT_EXTRADESC_DELETE
    REDIT_CONFIRM_SAVESTRING
)
```

## Task Groups

Every group follows test-first + mutation-verify (Edit-only revert per manager constraints).

### G1 — Scaffolding: `OlcData` + `DescriptorData.Olc` + REDIT_ mode constants

- Files: `internal/types/olc.go` (new), `internal/types/descriptor.go` (add field), `internal/types/enums.go` (add const block).
- Test first: `TestDescriptorOlc_DefaultsNil`, `TestOlcData_ModeConstantsUnique` (asserts the REDIT_ const values are pairwise distinct and exceed CON_* values so they can't be conflated by accident).
- Mutation: drop `Olc *OlcData` field → `TestDescriptorOlc_DefaultsNil` fails to compile.

### G2 — `CON_REDIT` arm in pulse-loop dispatch

- File: `internal/game/loop.go:252-267`. Add:
  ```go
  case types.CON_REDIT:
      ReditParse(d, line)
  ```
- `ReditParse` lives in new file `internal/game/redit_parse.go` — stub that logs "unimplemented" and falls back to `cleanup_olc(d)`-equivalent so tests can exercise the dispatch even before G3.
- Test first: `TestLoop_ConReditDispatchesToReditParse` — set up a descriptor with `Connected = CON_REDIT`, inject a line, assert `ReditParse` was called (wire a spy via a package-level `reditParseFunc` function variable if needed for testability, same pattern as `StartEditingFunc`).
- Mutation: change the switch case to `types.CON_OEDIT` → test fails.

### G3 — Menu entry: extend `DoRedit` to open the menu when no subcommand is supplied

- File: `internal/act/olc.go:44-46`. Replace the current "Redit what?" usage line with: if `arg == ""`, set `d.Olc = &types.OlcData{Mode: REDIT_MAIN_MENU, Vnum: room.Vnum, Target: room}`, set `d.Connected = CON_REDIT`, call `ReditDispMenuFunc(d)` (new seam — same cross-package pattern as `StartEditingFunc`).
- Flat path (`redit name foo` etc.) stays in place for experienced-builder UX.
- New seam declarations in `internal/act/olc.go` top-of-file:
  ```go
  var ReditDispMenuFunc func(d *types.DescriptorData)
  ```
- Wire in `internal/boot/boot.go` alongside existing seams.
- Test first: `TestDoRedit_NoArgEntersMenu` — asserts `ch.Desc.Connected == CON_REDIT`, `ch.Desc.Olc.Mode == REDIT_MAIN_MENU`, `ch.Desc.Olc.Target == ch.InRoom` after `DoRedit(ch, "")`.
- Test: `TestDoRedit_WithArgKeepsFlatPath` — `DoRedit(ch, "name Foo")` does NOT set Connected to CON_REDIT.
- Mutation: remove the `d.Connected = CON_REDIT` assignment → first test fails.

### G4 — Menu rendering: `ReditDispMenu` + flag/sector sub-menus

- File: `internal/game/redit_menu.go` (new).
- Implement `ReditDispMenu(d)`, `reditDispFlagMenu(d)`, `reditDispSectorMenu(d)` mirroring C `src/oredit.c:423-477`, `:366-389`, `:399-420`.
- Output format: match C's color-coded menu (`&g1&w)` etc.) — reuse `util.Act`/`ch.Send` — color-code tags pass through the existing color processor.
- Do NOT emit the C `"50\x1B[;H\x1B[2J"` screen-clear sequence. Rationale: the testclient harness does not interpret ANSI cursor moves; the clear is cosmetic in C and omitting it keeps tests deterministic. (Decision noted in Open Q1.)
- Test first: `TestReditDispMenu_ContainsAllFields` — drive with a populated `RoomIndexData`, assert output contains vnum, name, description, sector name, tunnel, teledelay, televnum, and the 7+A+B+Q option lines.
- Mutation: remove the `tunnel` line from the format string → test fails.
- Wire `act.ReditDispMenuFunc = game.ReditDispMenu` in `boot.go`.

### G5 — `ReditParse` core dispatch + `REDIT_MAIN_MENU` branch

- File: `internal/game/redit_parse.go`.
- Implement the outer `switch d.Olc.Mode` skeleton with ONLY the `REDIT_MAIN_MENU` case fully populated; all other cases return "not yet implemented" stubs so G6-G10 can fill them incrementally.
- `REDIT_MAIN_MENU` dispatches on the first input character/digit:
  - `1` → set Mode=REDIT_NAME, prompt.
  - `2` → set `ch.Substate = SUB_ROOM_DESC`, build an `EditorSave` closure that:
    - writes `room.Description = CopyBufferFunc(ch)`,
    - calls `StopEditingFunc(ch)`,
    - re-sets `ch.Desc.Connected = CON_REDIT` AFTER StopEditingFunc (the `/s` handler at `game/editor.go:164` has already transitioned to CON_PLAYING before invoking the closure — the closure re-sets, it does not override a live transition),
    - calls `ReditDispMenu(ch.Desc)`.
  - `3` → call `reditDispFlagMenu`, Mode=REDIT_FLAGS.
  - `4` → `reditDispSectorMenu`, Mode=REDIT_SECTOR.
  - `5`/`6`/`7` → prompt for tunnel/teledelay/televnum, Mode accordingly.
  - `a`/`A` → `reditDispExitMenu`, Mode=REDIT_EXIT_MENU.
  - `b`/`B` → `reditDispExtradescMenu`, Mode=REDIT_EXTRADESC_MENU.
  - `q`/`Q` → cleanup: `d.Olc = nil`, `d.Connected = CON_PLAYING`, `ch.Substate = SUB_NONE`, emit "Room saved." (matches C `cleanup_olc`).
  - default → "Invalid choice.", redisplay menu.
- Test first: 4+ unit tests covering `1`/`3`/`4`/`Q` paths. Example: `TestReditParse_Q_CleansUpAndReturnsToPlaying`.
- Mutation: in the `Q` case, swap `CON_PLAYING` for `CON_REDIT` → test fails.

### G6 — Field-set branches: `REDIT_NAME`, `REDIT_SECTOR`, `REDIT_TUNNEL`, `REDIT_TELEDELAY`, `REDIT_TELEVNUM`

- File: `internal/game/redit_parse.go` (extend).
- Each branch: mutate the room field (using `URANGE` semantics from C where applicable — `tunnel` clamped 0..1000; `tele_vnum` 1..MAX_VNUM), log via `util.Log`, redisplay main menu, reset Mode=REDIT_MAIN_MENU.
- Test first per branch. Example: `TestReditParse_Tunnel_Clamps` — drive with input `-5` → asserts `room.Tunnel == 0`; input `9999` → `== 1000`.
- Mutation: change `URANGE(0, n, 1000)` upper bound to `100` → clamp test fails.

### G7 — `REDIT_FLAGS` branch (two input modes: number or word list)

- Input as integer `0` → exit to main menu. Integer 1..32 → toggle that flag bit on `room.RoomFlags`. Word list → `get_rflag(word)` lookup per word; toggle each.
- File: `internal/game/redit_parse.go`.
- Reuse: flag-name lookup helper — scan `util.RoomFlagBits` map or equivalent; if absent, add a `types.GetRFlag(name string) int` helper backed by the `r_flags` string table (must exist somewhere in Go — verify during implementation; `act/setstat.go` handles rset flags).
- Test first: `TestReditParse_Flags_WordToggle`, `TestReditParse_Flags_NumberToggle`, `TestReditParse_Flags_ZeroExits`.
- Mutation: change `^= (1 << (n-1))` to `|= (1 << (n-1))` → toggle-test (flag set then unset) fails.

### G8 — Exit sub-menu: `REDIT_EXIT_MENU` / `_EDIT` / `_ADD` / `_ADD_VNUM` / `_DELETE` / `_VNUM` / `_KEY` / `_KEYWORD` / `_DESC` / `_FLAGS`

- File: `internal/game/redit_parse.go` (largest single branch).
- Each sub-state maps 1:1 to a C branch (see table in C Reference). The `Spare` field on `OlcData` stashes the exit pointer between invocations.
- `REDIT_EXIT_DESC` sets the description inline (single-line, no text editor) — matches C `oredit.c:780-790`.
- Test first: end-to-end `TestReditParse_ExitFlow_AddAndDelete` — enter exit menu, `A n 1001`, verify exit created to room 1001; `R <n>`, verify removed.
- Test: `TestReditParse_ExitFlags_Toggle`.
- Mutation: in `REDIT_EXIT_ADD_VNUM`, comment out the `make_exit` call → create test fails.

### G9 — Extradesc sub-menu: `REDIT_EXTRADESC_MENU` / `_CHOICE` / `_KEY` / `_DESCRIPTION` / `_DELETE`

- File: `internal/game/redit_parse.go`.
- `_DESCRIPTION` path reuses the same EditorSave-restore-CON_REDIT pattern as the main-menu `2` path (G5).
- The C path junks an extradesc that has null keyword+description when the user Qs out. Port this guard (mirrors `oredit.c:941-949`).
- Test first: `TestReditParse_Extradesc_AddKeyJunksOnQuitIfEmpty`.
- Mutation: remove the junk branch → test fails.

### G10 — `cleanup_olc` helper + session-end invariants

- File: `internal/game/redit_parse.go` or `internal/game/olc.go` (new).
- `CleanupOlc(d)` function: nulls `d.Olc`, sets `d.Connected = CON_PLAYING`, clears `d.Character.Substate = SUB_NONE`, emits cleanup message.
- Test first: `TestCleanupOlc_ResetsDescriptorAndCharacter`.
- Mutation: remove `d.Olc = nil` → follow-up test that drives a re-entry (`DoRedit` with existing Olc set) fails.

### G11 — `olcLog` helper (small, optional but recommended)

- File: `internal/util/olclog.go` (new, or append to `util/log.go`).
- `OlcLog(ch *CharData, roomOrObjRef any, format string, args ...any)` — formats `Log <player>: ROOM(<vnum>): <msg>` and routes to the shared logger with a `BUILD` tag.
- Every mutation path in G5-G9 calls `util.OlcLog(ch, room, ...)`.
- Test first: capture log output via a test hook, assert format.
- Mutation: change the format template → test fails.

### G12 — Integration tests via testclient harness

- File: `internal/game/redit_integration_test.go` (new) OR `internal/testclient/redit_test.go`.
- Scenario: immortal logs in, types `redit`, expects main menu prompt, types `1`, expects "Enter room name", types `Foo Bar`, expects "Changed name to Foo Bar" log + menu redisplay, types `Q`, expects return to CON_PLAYING + their normal prompt.
- Scenario: `redit`, `2`, enter text, `/s`, verify `room.Description` updated AND descriptor returns to CON_REDIT (not CON_PLAYING), main menu redisplayed.
- Scenario: `redit`, bad input at main menu, verify "Invalid choice." and menu redisplay (no session loss).
- Use `WithPrompt` seam from Tier 5 for prompt-tracking. Note: `ReditDispMenu` emits menu text without a trailing `> ` prompt sentinel (matches C `redit_disp_menu` — the menu itself is the prompt). Testclient scenarios must match on menu-text substrings (e.g. `"A) Exits"`) rather than generic prompt symbols. If this proves too flaky, consider emitting a synthetic trailing sentinel; deferred to implementation.
- Mutation: break the CON_REDIT-restore callback in G5 (have it set CON_PLAYING) → second scenario fails at the "still in CON_REDIT" assertion.

### G13 — Documentation + CHANGELOG + TODO housekeeping

- Update `CLAUDE.md` "Phase 6 plans authored" table with a new row for `plan-phase6-olc-redit.md`.
- Append a CHANGELOG entry.
- Move the Wave-D "redit CON_REDIT" item from Active to Done in `TODO.md` (if present), or add a new Done entry.

## Acceptance Criteria

Each numbered criterion is mechanically verifiable via `go test` or the testclient harness.

A1. `DescriptorData` has an `Olc *OlcData` field that is nil on a fresh descriptor.
A2. `OlcData` struct has `Mode int`, `Vnum int`, `Change bool`, `Target any`, `Spare any`, `TempNum int` fields.
A3. `REDIT_*` mode constants are distinct integers in a single `const` block, distinct from the `CON_*` values.
A4. `loop.go` dispatch switch has a `case types.CON_REDIT:` arm that calls `ReditParse(d, line)`.
A5. `DoRedit(ch, "")` (no argument) transitions `ch.Desc.Connected` to `CON_REDIT`, allocates `ch.Desc.Olc`, sets `Olc.Target == ch.InRoom`, `Olc.Mode == REDIT_MAIN_MENU`, and emits the main menu.
A6. `DoRedit(ch, "name Foo")` (with a flat subcommand) does NOT change `Connected`; the flat path still works.
A7. In the menu, input `1 <CR>` → prompts for name; follow-up line sets `room.Name` and redisplays the menu.
A8. In the menu, input `2 <CR>` enters the text editor; `/s` writes `room.Description` AND restores `Connected == CON_REDIT` (NOT CON_PLAYING); main menu redisplays.
A9. In the menu, input `3 <CR>` shows the flag menu; integer or word input toggles the correct flag bit on `room.RoomFlags`.
A10. In the menu, input `4 <CR>` shows the sector menu; integer 0..SECT_MAX-1 sets `room.SectorType` (rejects SECT_DUNNO and out-of-range).
A11. In the menu, inputs `5`/`6`/`7` set tunnel (clamp 0..1000), teledelay, tele_vnum (clamp 1..MAX_VNUM).
A12. In the menu, `A` enters the exit menu; sub-branches `A` (add), `R` (remove), numeric (edit) all work; returning via `Q` goes back to main menu.
A13. In the menu, `B` enters the extradesc menu; sub-branches add / remove / edit-description reuse the editor round-trip.
A14. In the menu, `Q` exits cleanly: `Connected == CON_PLAYING`, `Olc == nil`, `Substate == SUB_NONE`, closing message emitted.
A15. Invalid input at any menu prints "Invalid choice." (or C-equivalent message) and redisplays the current menu — does NOT drop out of CON_REDIT.
A16. The flat `redit name`/`desc`/etc. subcommands continue to work unchanged (regression test: existing `act/olc_test.go` tests still green).
A17. `go build ./...` green, `go test -count=3 ./...` green.
A18. Mutation matrix (≥ 8 mutations total across G1-G12) all caught by at least one test and reverted via `Edit`-only (no `git checkout`).

## Scope Cuts / Deferrals

- **Save-confirmation flow** (`REDIT_CONFIRM_SAVESTRING`) — C has it behind a commented-out block at `oredit.c:580-586`. The current C code already skips this and just `cleanup_olc`s on `Q`. Port matches current C behavior: no save-confirm. Mode constant reserved for future.
- **`redit_setup_new`** — for creating a NEW room from a fresh vnum, currently not called from `redit_parse` in the shipped C (see prototype at `oredit.c:66` — no definition in this file). Already partially handled by `DoRdig` (Phase 4b). Not in scope here.
- **ANSI screen-clear sequence** (`"50\x1B[;H\x1B[2J"`) — cosmetic; not emitted. Noted in G4.
- **Double-edit guard** (C `oredit.c:114-121` — reject if another descriptor is editing the same vnum) — DEFERRED to a follow-up. Single-builder-at-a-time is acceptable for this plan; the guard requires a world-wide descriptor scan, which is cleaner once the `oedit`/`medit` plans generalize the lookup.
- **`OLC_CHANGE(d)` dirty-bit honoring** — the field is allocated; no exit path consumes it. Deferred.
- **Building-log persistence** — `olcLog` writes to stdout only (G11). Persisted build-log files (C writes `build.log`) is deferred.
- **`can_rmodify` fine-grained permission model** — C checks area ownership and level ranges. Current Go `DoRedit` checks `GetTrust() >= LEVEL_IMMORTAL` only. Preserve that simpler check; defer full `can_rmodify` to an admin-plan.

## Open Questions

1. **Screen-clear in menu redisplays.** C emits `"50\x1B[;H\x1B[2J"` at the top of every `redit_disp_*` call. Go plan omits this for testability. **Recommended:** omit (matches Go-port convention of ANSI minimalism; testclient harness doesn't interpret cursor moves). Trivial to revisit if builders complain.
2. **`Olc` field on `DescriptorData` vs `CharData`.** C stores on both (`d->olc`, `ch->dest_buf`, `ch->substate`). Go plan: `Olc` on Descriptor, reuse existing `Substate` and `EditorSave` on CharData. **Recommended:** as above — Descriptor holds the mode-state union, Character holds the editor callback. Matches existing Tier 12 pattern.
3. **Menu color-code handling.** C uses `&g`/`&w`/`&O` color tags. Go has `util.Act` per-call AType (Tranche C) but the menu is not a character act — it's descriptor output. Reuse the raw-`Send` path + the existing color processor wired on the descriptor (`ColorFunc` at `descriptor.go:74`). **Recommended:** pass color-tag strings directly; the processor handles them at flush time (same idiom as existing `DoWho` output).
4. **`REDIT_EXIT_VNUM` semantics.** C `oredit.c:860-875` assigns `pexit->vnum = number` but does NOT update `pexit->to_room`. This is arguably a C bug — the exit now has a different vnum than its target room. **Recommended:** fix in Go — update both `exit.Vnum` AND `exit.ToRoom` via `GetRoom(number)`. Document the divergence in the plan and flag as a C-bug fix in CHANGELOG.
5. **Logging destination.** Port `olc_log` now, or defer? **Recommended:** port minimally via `util.Log(...)` (new tag "BUILD"). Persistence of a dedicated build-log file is deferred.
6. **Menu redisplay after field-set branches.** C `redit_parse` falls through to the bottom of the function which re-displays the main menu. Should each Go field-set branch explicitly call `ReditDispMenu(d)` or use a defer? **Recommended:** explicit call at the end of each branch — easier to audit, no surprise execution order.

## Risk Analysis

- **Medium overall.** State machine has 24 modes; risk is state-confusion bugs where a wrong mode transition leaves the descriptor unable to exit (builder has to disconnect). Mitigated by:
  - G12 integration tests drive the full round-trip.
  - G15 (implicit) — every branch either explicitly sets `d.Olc.Mode` to a known value or redisplays the same mode. No fallthroughs.
  - `CleanupOlc` is idempotent and can be invoked from a panic-recover if we add one (future enhancement, not in scope).
- **Editor-callback CON_REDIT restore is the single highest-risk point.** If the callback forgets to flip `Connected` back to `CON_REDIT` after `StopEditing`, the builder gets dropped to CON_PLAYING mid-session and the menu is lost. Mitigated by A8 acceptance test; mutation-verify against this specific path.
- **Cross-package seam count.** Adds `ReditDispMenuFunc` (and possibly a parse spy for testing). Follows the established pattern (`StartEditingFunc`, `CopyBufferFunc`, `StopEditingFunc`) — same wiring location in `boot.go`, same test-hygiene tradeoffs.
- **Blast radius on `DescriptorData`.** Adding one pointer field (`Olc`) is additive — no existing consumers care. `boot_test.go`'s non-nil-post-boot assertions only cover the existing seams; the new `ReditDispMenuFunc` seam needs a new assertion (per Tier-12 adversary-caught gap).
- **C-bug fixes.** Q4 recommends fixing the exit-vnum asymmetry. Flag clearly in plan and CHANGELOG; easy to revert if a builder depended on the C quirk (unlikely — the quirk is silently destructive).

## Adversary-Resolved Concerns

_External adversary review completed 2026-04-18 — verdict PASS with minor CONCERNS; 2 surgical edits applied (G5 wording clarification, G12 prompt-tracking note). Design decisions (OlcData placement, REDIT_EXIT_VNUM fix-in-Go, 13-group decomposition) sustained. `Agent` tool unavailable in this environment; structured self-review substituted. Self-review notes follow._

- Verified C line numbers for every cited function.
- Verified Go state against `loop.go:252-267`, `olc.go:16-28`, `descriptor.go:13-75`, `enums.go:89-91`.
- Verified the existing flat `DoRedit` uses `EditorSave` closures with the same callback shape this plan extends.
- Verified `CopyBufferFunc` / `StopEditingFunc` / `StartEditingFunc` seams are already wired in `boot.go:133-137`.
- Flagged the exit-vnum asymmetry as a C bug (Open Q4) rather than blindly porting.
- Scope-cut the save-confirmation flow and the double-edit guard — both are optional UX, not correctness-critical.
- No `git checkout`/`git restore`/`git stash`/`git reset --hard` invocations in the plan's mutation-verify steps — all revert via `Edit` only (per manager constraints).
- Adversary-tool availability status flagged at the end of the completion report.

## Completion Record

**Landed 2026-04-19.** All 13 task groups executed; all 18 acceptance criteria satisfied. First of three OLC-substate plans; establishes the nanny-dispatch pattern + `OlcData`-on-descriptor + `CON_REDIT` loop arm that `oedit`/`medit` will inherit.

### Group-by-group summary

- **G1** — `OlcData` struct (6 fields: Mode, Vnum, Change, Target, Spare, TempNum) + 24 REDIT_* constants offset `iota+100` to keep clear of CON_* iota space. `DescriptorData.Olc *OlcData` added. Pinned by 4 tests in `internal/types/olc_test.go`: nil-default, struct-literal field shape, mode uniqueness, CON_*-vs-REDIT_* separation.
- **G2** — `CON_REDIT` arm added to `internal/game/loop.go:252-267` processInput dispatch switch. Mutation swap `CON_REDIT`→`CON_OEDIT` caught by E2E test (drops to nanny default → "Unexpected state. Disconnecting.").
- **G3** — `DoRedit(ch, "")` extended to open the menu: allocate `ch.Desc.Olc`, set `Connected=CON_REDIT`, invoke `ReditDispMenuFunc` seam. Flat path (`redit name foo` etc.) untouched. `TestDoRedit_NoArgEntersMenu` + `TestDoRedit_WithArgKeepsFlatPath` pin both branches; obsolete `TestDoRedit_NoArg` expecting "Redit what?" removed.
- **G4** — Menu renderers in new `internal/game/redit_menu.go`: ReditDispMenu (main), reditDispFlagMenu, reditDispSectorMenu, reditDispExitMenu, reditDispExitEdit, reditDispExitFlagMenu, reditDispExitDirs, reditDispExtradescMenu, reditDispExtradescChoice. Source tables ported verbatim: `r_flags[]` from `src/build.c:92-105` (41 labels, matching overland-code conditional removed — stock build has no overland); sector labels from `redit_disp_menu` switch at `src/oredit.c:430-446`; `sectorKeywords` from `sector_names[]` at `:391-396`; direction names; exitFlagLabels (29 bits up to MAX_EXFLAG=28 per `constants.go:625`, with reserved slots EX_RES1/EX_RES2/EX_PORTAL skipped in UI). ANSI screen-clear sequence deliberately omitted per §Open Q1.
- **G5** — `reditParse` core + REDIT_MAIN_MENU branch in new `internal/game/redit_parse.go`. Q → `cleanupOlc` + "Exiting editor." (wording corrected per readiness-vet: C `cleanup_olc` discards, no "Room saved." emitted). `2` → EditorSave closure that re-sets `Connected = CON_REDIT` AFTER `StopEditing` (wording-verified: `/s` handler at `game/editor.go:163-172` transitions to CON_PLAYING BEFORE invoking closure). Mirror C `do_redit_reset` at `src/oredit.c:493-535`. Mutation: drop CON_REDIT re-set → `TestReditParse_DescEditorCallback_ReturnsToConRedit` fails.
- **G6** — REDIT_NAME / REDIT_SECTOR / REDIT_TUNNEL / REDIT_TELEDELAY / REDIT_TELEVNUM branches. Tunnel URANGE-clamped [0, 1000]; televnum clamped [1, MAX_VNUM]; sector validator rejects SECT_DUNNO + out-of-range (mirrors C `oredit.c:685-696`). Names pass through `util.SmashTilde`. Mutations: flip tunnel upper 1000→100; drop SECT_DUNNO guard — both caught.
- **G7** — REDIT_FLAGS branch: whole-input integer `0` returns to main; `1..32` toggles 1-indexed bit; non-numeric treated as space-separated word list dispatched through `getRoomFlagBit`. Mutation: Toggle→Set breaks the "toggle-off" test.
- **G8** — Exit sub-machine: REDIT_EXIT_MENU / EDIT / ADD / ADD_VNUM / DELETE / VNUM / KEY / KEYWORD / DESC / FLAGS. All 10 modes handled. **REDIT_EXIT_VNUM C-bug fix**: Go updates both `pexit.Vnum` AND `pexit.ToRoom` (C at `src/oredit.c:860-875` sets only vnum, leaving ToRoom stale). Pinned by `TestReditParse_ExitVnum_SetsBothFields`. Add-vnum path validates the destination exists via `worldRoomLookup` (seam to `world.GetRoom`). Direction input accepts both numeric (0..DIR_SOMEWHERE) and word-prefix forms; duplicate direction detected and rejected with menu redisplay.
- **G9** — Extradesc sub-machine: REDIT_EXTRADESC_MENU / CHOICE / KEY / DESCRIPTION / DELETE. Matches C junk-on-Q-if-empty branch at `oredit.c:941-949`. Description path reuses the same EditorSave + restore-CON_REDIT pattern from G5.
- **G10** — `cleanupOlc` helper: nulls Olc, sets Connected=CON_PLAYING, clears Substate=SUB_NONE. Idempotent (second call no-op). Mutation: swap CON_PLAYING→CON_REDIT caught by Quit + cleanup tests.
- **G11** — `olcLog` ports C `olc_log` at `src/oredit.c:171-210` with the ROOM(vnum) prefix only (OBJ/MOB branches deferred to future oedit/medit plans). Routes through `util.LogStringPlus` tagged `types.LOG_BUILD` + the character's trust. Format verified by observation in test output.
- **G12** — testclient E2E in new `internal/testclient/redit_test.go`: (1) `TestTestclient_ReditMenuEntryAndQuit` — immortal types `redit`, reads menu (match on "Enter choice" since no `> ` sentinel, per G12 readiness-vet note), types Q, sees "Exiting editor."; (2) `TestTestclient_ReditMenuSetName` — `redit`, `1`, text, verify redisplay contains new name; (3) `TestTestclient_ReditInvalidChoiceRedisplays` — bad input produces "Invalid choice!" and the menu redisplays (no CON_REDIT drop). All 3 green. Mutation: swap CON_REDIT arm in loop.go → first test fails with "Unexpected state. Disconnecting." (nanny default fired instead of parser).
- **G13** — CLAUDE.md row flipped to **LANDED 2026-04-19**; CHANGELOG.md 2026-04-19 entry prepended; TODO.md item #6 updated; this Completion Record appended.

### Mutation matrix (≥8 required per A18)

9 `Edit`-round-trip mutations verified — no `git checkout` / `git restore` / `git stash` / `git reset --hard` / `git commit --amend` used at any point in this lineage:

1. Drop `d.Desc.Connected = CON_REDIT` in DoRedit → `TestDoRedit_NoArgEntersMenu` fails (Connected=0).
2. Swap CON_PLAYING→CON_REDIT in `cleanupOlc` → `TestReditParse_Quit_CleansUpAndReturnsToPlaying` + `TestCleanupOlc_ResetsDescriptorAndCharacter` both fail.
3. Flip tunnel URANGE upper 1000→100 → `TestReditParse_Tunnel_Clamps` fails on 9999 case.
4. Replace `Toggle(bit)` with `Set(bit)` in flag-number branch → `TestReditParse_Flags_NumberToggle` fails (can't un-set).
5. Drop `n == types.SECT_DUNNO` guard → `TestReditParse_Sector_RejectsDunno` fails.
6. Drop `pexit.ToRoom = target` Go-bug-fix in REDIT_EXIT_VNUM → `TestReditParse_ExitVnum_SetsBothFields` fails (this is the C-bug-reproducing mutation).
7. Disable extradesc empty-junk branch → `TestReditParse_Extradesc_AddKeyQuitJunksIfEmpty` fails.
8. Swap `case CON_REDIT:` → `case CON_OEDIT:` in loop.go dispatch → `TestTestclient_ReditMenuEntryAndQuit` E2E fails.
9. Drop CON_REDIT re-set in main-menu-`2` EditorSave closure → `TestReditParse_DescEditorCallback_ReturnsToConRedit` fails.

### Test count delta

43 new tests — 4 types (olc_test.go), 34 game (redit_parse_test.go), 2 act (olc_test.go), 3 testclient (redit_test.go). Added to the pre-existing suite: `go test -count=3 ./...` green across all 15 packages.

### Deferrals honored (plan §Scope Cuts)

- **REDIT_CONFIRM_SAVESTRING flow** — constant reserved; C already skips via commented-out block at `oredit.c:580-586`.
- **`redit_setup_new`** — new-room creation partially covered by Phase 4b `DoRdig`; the C path in this file (prototype at :66 without definition) is unused.
- **Double-edit guard** — world-wide descriptor scan deferred until oedit/medit generalize the lookup shape.
- **`OLC_CHANGE` dirty-bit consumer** — field allocated on `OlcData`; no path reads it yet.
- **Persisted build-log file** — `olcLog` writes via `util.LogStringPlus` only; dedicated `build.log` file deferred.
- **Fine-grained `can_rmodify`** — current `GetTrust() >= LEVEL_IMMORTAL` check retained; full area-ownership + level-range logic deferred to admin-plan.

### Open Q resolutions

- **Q1** (screen-clear omitted): retained the plan recommendation. No testclient breakage; cosmetic-only behavior change.
- **Q2** (Olc on Descriptor vs Char): retained — Olc on Descriptor, reuse existing `Substate`/`EditorSave` on Char. Consistent with Tier 12 precedent.
- **Q3** (color codes): retained — menu emits `&g`/`&w`/`&O` tags; the existing descriptor `ColorFunc` processes at flush.
- **Q4** (REDIT_EXIT_VNUM C bug): fix-in-Go applied; both `pexit.Vnum` and `pexit.ToRoom` now updated. CHANGELOG cites the bug.
- **Q5** (olcLog destination): ported minimally via `util.LogStringPlus` with `LOG_BUILD` tag. Persistence deferred.
- **Q6** (field-set branch menu redisplay): explicit call at the end of each branch. No defers; clearer control flow.

### Downstream readiness

Nanny-dispatch pattern + `OlcData`-on-descriptor + `CON_REDIT` loop arm are **ready for `oedit`/`medit` to inherit**:

- `internal/types/olc.go` defines the shared struct with `Target any` so `*ObjIndexData` / `*MobIndexData` type-assertions just work.
- `internal/game/loop.go` dispatch switch needs one new case per future editor (`CON_OEDIT` / `CON_MEDIT` → respective parser).
- `internal/boot/boot.go` wires the menu seam; future editors will register `act.OeditDispMenuFunc = game.OeditDispMenu` etc.
- `internal/game/redit_parse.go` demonstrates the EditorSave restore-CON_REDIT contract for line-editor round-trips — same pattern applies for any descriptor-editing mode.
- `internal/testclient/redit_test.go` shows the E2E pattern (match on menu-text substrings, not `> ` sentinel).
- `CharData.Substate` sub-state constants for OEDIT / MEDIT already exist; future plans fill in the analogous editor dispatch.

### Tooling caveat

`Agent` tool remained unavailable in the `manager` subagent harness across this lineage — structured self-review substituted for an external adversary pass. Plan audit (2026-04-18, PASS with minor CONCERNS) informed all design decisions; the 2 readiness-vet corrections from the orchestrator (G5 wording, G12 prompt-tracking) were applied in-place before execution.

### Commit

`f678e8a`. See CHANGELOG.md 2026-04-19 entry for the exhaustive file list.
