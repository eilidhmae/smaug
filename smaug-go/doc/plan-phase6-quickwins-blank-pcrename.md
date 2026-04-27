# Plan: Phase 6 quick-wins — `pulseSave` init / `DoBlank` / `worldPcLookup` / `do_pcrename`

**Status:** PLANNED 2026-04-26.
**Priority:** P3 (Phase 6, four small follow-ups bundled into one lineage).
**Lineage:** `phase6-quickwins-blank-pcrename`.
**Scope:** Four independent deliverables, each its own task-group block:

- **D1** — Initialize `pulseSave` field in `NewGameLoop` (autosave-first-pulse bug).
- **D2** — `DoBlank` user-facing toggle for `PLR_BLANK` (rendering already shipped).
- **D3** — `worldPcLookup` seam + `DoMedit` PC-by-name argument resolution.
- **D4** — `do_pcrename` port (in-memory rename + on-disk pfile rename) wired into the medit MEDIT_NAME PC arm.

Each deliverable is independently shippable. D4 depends on D3 (`worldPcLookup`). D1/D2 are independent of everything else.

---

## Cross-Plan Dependencies

**Depends on (LANDED):**

- Phase 6 medit Wave 5 (`plan-phase6-olc-medit.md` LANDED 2026-04-26, commit `853c5b6`) — interactive CON_MEDIT menu + MEDIT_NAME PC arm at `internal/game/medit_arms.go:74-87`.
- Phase 6 Tranche C G6 (2026-04-18) — `PLR_BLANK` rendering on descriptor flush. The flag persists via `ch.Act` and is honored by the renderer; only the user-facing toggle is missing.
- `internal/persist/player.go:735-745` `PlayerFilePath(dataDir, name)` — pfile path resolver. Returns `<data>/player/<lowercase-first-letter>/<name>` and validates the name against `validPlayerName = ^[a-zA-Z]{3,12}$` (`:731`). Reused by D4.
- `internal/act/playercfg.go:68-83` `DoGag` — toggle pattern reused by D2 (NPC reject + nil-PCData guard + directional messages).
- `internal/game/medit_parse.go:22-31` `worldMobLookup` — seam pattern reused by D3 (`worldPcLookup`).
- `internal/act/olc_interactive.go:286-307` `DoMedit` — non-numeric arg currently rejects with "PC editing by name not yet supported" at `:305`. D3 replaces this rejection with a PC lookup + entry into the PC menu.
- `internal/game/loop.go:64-69, 81-97` `GameLoop.pulseSave` + `NewGameLoop` — D1 inserts the missing field initializer.
- `internal/types/enums.go:964` `PLR_BLANK` — already defined.
- `internal/types/constants.go:78` `PULSE_SAVE = 5*60*PULSE_PER_SECOND` — referenced by D1.

**Blocks:** nothing.

**Does NOT depend on:** auction, foldarea, mpedit, oedit, redit, holidays, hotboot, or any other Phase-6 lineage.

---

## C Reference (authoritative)

All line numbers verified 2026-04-26 from `src/`.

### `do_pcrename` — `src/act_wiz.c:12665-12770` (~106 LOC)

The full SMAUG rename. Key behaviors carried over to Go (D4):

| Lines | Behavior | Go equivalent |
|---|---|---|
| 12676-12677 | `one_argument` twice → `arg1` (old), `arg2` (new). | `util.OneArgument` chained. |
| 12678 | `smash_tilde(arg2)`. | `util.SmashTilde`. |
| 12681-12682 | NPC caller → silent return. | `if ch.IsNPC() return`. |
| 12684-12688 | Either arg empty → "Syntax: rename <victim> <new name>". | Same message. |
| 12690-12694 | `check_parse_name(arg2, 1)` — illegal name reject. | `validPlayerName.MatchString` from `persist/player.go:731`. |
| 12698-12702 | **`get_char_room(ch, arg1)`** — room-only lookup ("security precaution so you don't rename someone you don't mean to"). | **DIVERGENCE:** Go uses `worldPcLookup` (D3) per user directive. Documented in §Open Questions. |
| 12703-12707 | NPC victim → "You can't rename NPC's." | `if victim.IsNPC()` reject. |
| 12709-12713 | `get_trust(ch) < get_trust(victim)` → "I don't think they would like that!" | `ch.GetTrust() < victim.GetTrust()` reject. |
| 12714-12721 | Build `newname` / `oldname` / `backname` paths under `PLAYER_DIR` / `BACKUP_DIR`. | `persist.PlayerFilePath(dataDir, oldName)` + `(dataDir, newName)`. **No `BACKUP_DIR` in Go port today** — the `backname` removal step is dropped (documented in §Open Questions). |
| 12722-12726 | `access(newname, F_OK) == 0` → "That name already exists." | `os.Stat(newPath); err == nil` reject. |
| 12729-12735 | Immortal: remove `GOD_DIR/<name>` entry. | **DEFER** — wizlist port pending; see §Scope Cuts. |
| 12738-12749 | If `ch->pcdata->area` (the renaming immortal owns an area), rename `<BUILD_DIR><victim->name>.are` → `<BUILD_DIR><arg2>.are` and `.bak`. | **DEFER** — Go has no per-immortal area assignment yet; see §Scope Cuts. |
| 12751-12754 | `victim->name = arg2`; `victim->pcdata->filename = arg2`. | `victim.Name = newName` (Go has no separate `pcdata.filename` field — reuses `Name` for pfile). |
| 12755-12761 | `remove(backname)`; `remove(oldname)` with bug-log on failure. | `persist.RenamePlayerFile(dataDir, oldName, newName)` does `os.Rename(oldPath, newPath)`. The `.bak` remove is dropped per above. |
| 12763 | `save_char_obj(victim)` — force a fresh save under the new name. | Optional: existing `act.SaveFunc(victim)` available; documented in D4 design. |
| 12766-12767 | `make_wizlist()` if immortal. | **DEFER** with rest of wizlist. |
| 12768 | "Character was renamed.\n\r". | Same message. |

**Key C behaviors NOT carried over (intentional):**

- `BACKUP_DIR/<name>` removal — Go pfile layout has no `.bak` analog yet.
- Per-immortal area-file rename — Go has no `pcdata.area` ownership.
- Wizlist rebuild — wizlist port deferred independently.
- `get_char_room` — replaced with `worldPcLookup` (world-wide) per user directive.

### `do_omedit` — `src/omedit.c:110-232`

| Lines | Behavior | Go relevance |
|---|---|---|
| 180 | `victim = get_char_world(ch, arg)` — world-wide name lookup. | D3 `worldPcLookup` mirrors this (PC subset only — NPCs use `worldMobLookup`). |
| 196-200 | Trust gate: `if (!IS_NPC(victim) && get_trust(ch) < sysdata.level_modify_proto) "Huh?"`. | D3 trust gate: `LEVEL_GREATER` per user directive. **Divergence note:** C uses `sysdata.level_modify_proto` (configurable, default LEVEL_IMMORTAL); Go hardcodes `LEVEL_GREATER`. Recorded in §Open Questions Q3. |

**User prompt cited `omedit.c:1247-1253` as the trust precedent** — that location is the PC-main-menu digit dispatch (no trust gate). The actual PC-trust gate for `do_omedit` is at `:196`. The `LEVEL_SUB_IMPLEM-1` gate at `:1390` is specifically for the `do_pcrename` invocation inside MEDIT_NAME, not for menu entry.

### `omedit.c:1389-1405` — MEDIT_NAME PC arm (D4 wire-up site)

```c
case MEDIT_NAME:
    if ( !IS_NPC(victim) && get_trust( d->character ) > LEVEL_SUB_IMPLEM-1 )
    {
        sprintf( buf, "%s %s", victim->name, arg );
        do_pcrename( d->character, buf );
        olc_log( d, "Changes name to %s", arg );
        return;
    }
    STRFREE( victim->name );
    victim->name = STRALLOC( arg );
    ...
```

C invokes `do_pcrename(d->character, "<old> <new>")` only when victim is PC AND caller is `> LEVEL_SUB_IMPLEM-1` (i.e. `>= LEVEL_SUB_IMPLEM`). Below that threshold the PC name is silently changed in-memory only — pfile remains under the old name (latent C bug).

**Go port:** invoke `act.DoPcrename` for PC victims unconditionally (any caller already at `LEVEL_GREATER` to enter the medit PC arm — see D3 trust gate). Documented in §Open Questions Q4.

---

## Go Current State (verified 2026-04-26)

| Artifact | Location | Status |
|---|---|---|
| `GameLoop.pulseSave` | `internal/game/loop.go:68` | Field declared but NOT initialized in `NewGameLoop` at `:83-97`. **D1 fixes.** |
| `GameLoop` other pulse fields | `internal/game/loop.go:64-69, 88-92` | `pulseArea`, `pulseViolence`, `pulseMobile`, `pulseTick`, `pulseAuction` all initialized. `pulseSave` is the only outlier. |
| `types.PULSE_SAVE` | `internal/types/constants.go:78` | `5 * 60 * PULSE_PER_SECOND` (= 6000 pulses @ 4 PPS = 5 min). |
| `PLR_BLANK` flag | `internal/types/enums.go:964` | Defined. Honored by descriptor flush (Tranche C G6 LANDED 2026-04-18). |
| `DoGag` reference pattern | `internal/act/playercfg.go:68-83` | Standalone toggle, NPC reject, nil-PCData guard, directional messages. **D2 mirrors.** |
| `DoBlank` | — | **MISSING.** D2 adds. |
| `worldMobLookup` | `internal/game/medit_parse.go:22-31` | NPC-by-vnum seam. **D3 adds parallel `worldPcLookup`.** |
| `worldPcLookup` | — | **MISSING.** D3 adds. |
| `DoMedit` non-numeric arg | `internal/act/olc_interactive.go:298-306` | Currently rejects with "PC editing by name not yet supported". **D3 replaces with PC lookup + menu entry.** |
| `meditArmName` PC branch | `internal/game/medit_arms.go:74-87` | Currently does `victim.Name = arg` only — pfile is stale. **D4 wires `act.DoPcrename` here.** |
| `persist.PlayerFilePath` | `internal/persist/player.go:735-745` | Path resolver with charset validation. **D4 reuses for both old and new path.** |
| `validPlayerName` regex | `internal/persist/player.go:731` | `^[a-zA-Z]{3,12}$`. Reused by `D4` for charset validation. |
| `world.World.Descriptors` | `internal/world/world.go:23` | `[]*types.DescriptorData`. Iteration target for `worldPcLookup`. |
| `act.WorldRef` | `internal/act/olc.go` (boot-set) | Available globally for lookup. |
| `DescriptorData.Connected` | `internal/types/descriptor.go:31` | `int`, with `CON_PLAYING` at `enums.go:68`. |
| `DescriptorData.Character` | `internal/types/descriptor.go:19` | `*CharData`. |

---

## Go Design

### D1 — `pulseSave` field initializer

Insert one line into the struct literal in `NewGameLoop` at `internal/game/loop.go:84-96`:

```go
return &GameLoop{
    world:          w,
    cmdReg:         cmdReg,
    incoming:       incoming,
    pulseArea:      types.PULSE_AREA,
    pulseViolence:  types.PULSE_VIOLENCE,
    pulseMobile:    types.PULSE_MOBILE,
    pulseTick:      types.PULSE_TICK,
    pulseSave:      types.PULSE_SAVE,   // NEW — D1
    pulseAuction:   types.PULSE_AUCTION,
    internalCtx:    ctx,
    internalCancel: cancel,
    queryQueue:     make(chan func(), 16),
}
```

**Rationale:** every other pulse counter is initialized to its corresponding `PULSE_*` constant so the first decrement-to-zero fires at the correct interval. `pulseSave` zero-init means `g.pulseSave--` reaches its trigger on the first pulse, autosaving immediately at boot.

### D2 — `DoBlank`

```go
// DoBlank implements the 'blank' command — a standalone no-arg toggle
// for PLR_BLANK, which causes the renderer to emit an extra "\n\r"
// before each prompt (honored by descriptor flush, Tranche C G6 LANDED
// 2026-04-18 via plan-tranche-c.md).
//
// C divergence: SMAUG embeds blank toggling inside `do_config` at
// src/act_info.c:5585 (blank branch). The Go port instead ships
// standalone per-flag toggles (matching DoGag/DoAfk's local precedent)
// with directional messages. See plan-do-gag.md / plan-tranche-c.md
// for the divergence rationale.
//
// The nil-PCData guard mirrors DoGag at playercfg.go:72-74: IsNPC()
// only checks Act.ACT_IS_NPC and does not look at PCData, so a
// malformed non-NPC with nil PCData would pass IsNPC() and crash
// dereferencing .Act.Set without this guard. (PLR_BLANK lives on
// ch.Act, not PCData.Flags, but the nil-PCData guard is kept for
// consistency with the family pattern — DoBlank without it diverges
// from the local idiom even if the specific crash vector differs.)
func DoBlank(ch *types.CharData, argument string) {
    if ch.IsNPC() {
        return
    }
    if ch.PCData == nil {
        return
    }
    if ch.Act.IsSet(types.PLR_BLANK) {
        ch.Act.Remove(types.PLR_BLANK)
        ch.Send("Blank lines will no longer be inserted before each prompt.\n\r")
        return
    }
    ch.Act.Set(types.PLR_BLANK)
    ch.Send("Blank lines will be inserted before each prompt.\n\r")
}
```

**Boot reg** at `internal/boot/boot.go` adjacent to `DoGag`:

```go
reg.Register(&command.Command{Name: "blank", DoFun: act.DoBlank, Position: types.POS_DEAD, Level: 0})
```

### D3 — `worldPcLookup` + `DoMedit` PC-by-name resolution

#### D3a — `worldPcLookup` seam in `internal/game/medit_parse.go`

Place adjacent to existing `worldMobLookup`:

```go
// worldPcLookup resolves a player name to a connected *CharData by
// walking world.Descriptors. Mirrors C get_char_world() (handler.c:1894-)
// for the PC subset only — NPCs use worldMobLookup. Returns the first
// descriptor whose Connected == CON_PLAYING and whose Character.Name
// matches case-insensitively. Linkdead descriptors (Connected !=
// CON_PLAYING) and nil-Character descriptors are skipped.
//
// Tests can override directly without standing up a full world.
//
// Plan: plan-phase6-quickwins-blank-pcrename.md §D3.
var worldPcLookup = func(name string) *types.CharData {
    if worldRef == nil {
        return nil
    }
    for _, d := range worldRef.Descriptors {
        if d == nil {
            continue
        }
        if d.Connected != int(types.CON_PLAYING) {
            continue
        }
        if d.Character == nil {
            continue
        }
        if strings.EqualFold(d.Character.Name, name) {
            return d.Character
        }
    }
    return nil
}
```

The seam needs to be reachable from `internal/act/DoMedit`. Two options:

1. **Export from `game`** — but `act` cannot import `game` (cycle).
2. **Mirror in `act` package** with the same logic.

**Decision:** add a parallel `WorldPcLookup` (exported) in `act` package alongside `WorldRef`. The `game`-package `worldPcLookup` stays for medit-internal use; `act.WorldPcLookup` is what `DoMedit` calls. Both implementations are tiny — duplication is fine.

`internal/act/olc_pclookup.go` (NEW):

```go
// WorldPcLookup resolves a connected PC by name via WorldRef.Descriptors.
// Public so DoMedit (and future consumers) can invoke it without
// importing the game package. Test-overridable via assignment.
//
// See internal/game/medit_parse.go:worldPcLookup for the parallel
// game-package seam used by medit's interactive arms.
var WorldPcLookup = func(name string) *types.CharData {
    if WorldRef == nil {
        return nil
    }
    for _, d := range WorldRef.Descriptors {
        if d == nil || d.Character == nil {
            continue
        }
        if d.Connected != int(types.CON_PLAYING) {
            continue
        }
        if strings.EqualFold(d.Character.Name, name) {
            return d.Character
        }
    }
    return nil
}
```

#### D3b — `DoMedit` PC-by-name extension

Replace the rejection at `internal/act/olc_interactive.go:298-306`:

```go
vnum, err := strconv.Atoi(vnumArg)
if err != nil {
    // Non-numeric arg → PC-by-name lookup. Mirrors C do_omedit's
    // get_char_world(ch, arg) at omedit.c:180 for the PC subset.
    // Trust gate: LEVEL_GREATER for editing other PCs (Go-port
    // hardening; C uses sysdata.level_modify_proto at omedit.c:196,
    // configurable but defaulting to LEVEL_IMMORTAL — Go pins to
    // LEVEL_GREATER per plan §Open Questions Q3).
    victim := WorldPcLookup(vnumArg)
    if victim == nil {
        ch.Send("No such player connected.\n\r")
        return
    }
    if ch.GetTrust() < types.LEVEL_GREATER {
        ch.Send("Huh?\n\r")
        return
    }
    if ch.Desc == nil {
        ch.Send("No descriptor.\n\r")
        return
    }
    if ch.Desc.Connected == int(types.CON_MEDIT) && ch.Desc.Olc != nil {
        ch.Send("You are already editing a mob. Type Q to exit first.\n\r")
        return
    }
    ch.Desc.Olc = &types.OlcData{
        Mode:   types.MEDIT_PC_MAIN_MENU,
        Vnum:   0,        // PCs have no prototype vnum
        Target: victim,
    }
    ch.Desc.Connected = int(types.CON_MEDIT)
    if MeditDispMenuFunc != nil {
        MeditDispMenuFunc(ch.Desc)
    }
    return
}
```

**Trust gate ordering:** PC lookup happens BEFORE the trust gate so a builder typing `medit ghost` (where "ghost" is neither a connected PC nor numeric) gets "No such player connected." rather than a misleading "Huh?". This is a deliberate divergence from the existing `DoMedit` shape (which trust-gates first) — rationale: the lookup is read-only and side-effect-free, and "No such player" is the more informative message.

### D4 — `act.DoPcrename` + `persist.RenamePlayerFile`

#### D4a — `persist.RenamePlayerFile`

Place in `internal/persist/player.go` adjacent to `PlayerFilePath`:

```go
// RenamePlayerFile renames a player file on disk from oldName to
// newName. Uses PlayerFilePath for both source and destination
// resolution — both names go through the same charset validation
// (validPlayerName regex, ^[a-zA-Z]{3,12}$) and lowercase-first-letter
// subdirectory layout. Returns an error if either name is invalid,
// the source does not exist, the destination already exists, or the
// destination subdirectory cannot be created.
//
// Path-containment guarantee: PlayerFilePath rejects empty names,
// applies filepath.Base (strips any traversal), and matches against
// the validPlayerName regex (alphabetic only, 3-12 chars). A name
// containing "/", "..", or any non-alphabetic character returns
// "" from PlayerFilePath, which RenamePlayerFile maps to ErrInvalidName.
//
// Plan: plan-phase6-quickwins-blank-pcrename.md §D4.
var (
    ErrInvalidName       = errors.New("invalid player name")
    ErrSourcePfileNotFound = errors.New("source pfile does not exist")
    ErrDestPfileExists    = errors.New("destination pfile already exists")
)

func RenamePlayerFile(dataDir, oldName, newName string) error {
    oldPath := PlayerFilePath(dataDir, oldName)
    newPath := PlayerFilePath(dataDir, newName)
    if oldPath == "" || newPath == "" {
        return ErrInvalidName
    }
    if _, err := os.Stat(oldPath); err != nil {
        if os.IsNotExist(err) {
            return ErrSourcePfileNotFound
        }
        return err
    }
    if _, err := os.Stat(newPath); err == nil {
        return ErrDestPfileExists
    }
    // Ensure destination subdir exists (e.g. "Eilidh" → player/e/,
    // "Bob" → player/b/ — different first letter).
    if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
        return err
    }
    return os.Rename(oldPath, newPath)
}
```

**Security analysis:**

- `PlayerFilePath` calls `filepath.Base(name)` at `:739` before regex validation, stripping any directory components. A `newName` of `"../etc/passwd"` becomes `"passwd"`, then fails the `^[a-zA-Z]{3,12}$` regex (contains no path-separator after Base, but length is 6 OK; `"passwd"` is alphabetic — wait, `"passwd"` would PASS the regex). However, `filepath.Base("../etc/passwd")` returns `"passwd"`, which is a legal player name. The traversal is neutralized; the resulting `newPath` is `<dataDir>/player/p/passwd` — still inside the player tree. Confirmed safe.
- A `newName` of `"foo/bar"` → `Base` returns `"bar"` (3 chars, alphabetic) → resolves to `<dataDir>/player/b/bar`. Safe; the slash is stripped.
- A `newName` of `"a"` (too short) → fails regex → `PlayerFilePath` returns `""` → `RenamePlayerFile` returns `ErrInvalidName`.
- A `newName` of `"AB"` (too short) → fails regex.
- A `newName` of `"ThirteenCharsX"` (>12) → fails regex.
- A `newName` containing `'.'` (e.g. `"al.ice"`) → fails regex (non-alpha).
- A `newName` of `""` → `PlayerFilePath` returns `""` early → `ErrInvalidName`.

#### D4b — `act.DoPcrename`

Place in `internal/act/playercfg.go` (alongside `DoGag`):

```go
// DoPcrename implements the 'pcrename' command (C act_wiz.c:12665-12770
// do_pcrename). Renames a connected PC both in-memory (victim.Name) and
// on disk (the pfile is renamed via persist.RenamePlayerFile).
//
// Divergences from C (documented in plan-phase6-quickwins-blank-pcrename.md
// §C Reference table):
//   - Lookup: WorldPcLookup (world-wide via Descriptors) instead of
//     get_char_room. C's room-only lookup was a Shaddai-era security
//     precaution; the Go port relies on the LEVEL_GREATER trust gate
//     for the same protection (an immortal can already snoop / goto
//     any victim, so room-locality adds no real safety).
//   - No backname removal — Go pfile layout has no .bak analog.
//   - No god-dir cleanup — wizlist port deferred.
//   - No build-area rename — Go has no per-immortal area assignment.
//   - Trust precedence: ch.GetTrust() >= victim.GetTrust() (matches C).
//
// C-bug-not-preserved: C silently changes the in-memory PC name without
// renaming the pfile when caller's trust is below LEVEL_SUB_IMPLEM
// (omedit.c:1390 gate). The Go port routes ALL PC name changes through
// DoPcrename when the medit MEDIT_NAME PC arm fires (D4 wire-up); the
// caller has already passed the LEVEL_GREATER gate at DoMedit entry (D3).
func DoPcrename(ch *types.CharData, argument string) {
    if ch.IsNPC() {
        return
    }
    // CaseArgument preserves capitalization — OneArgument lowercases,
    // which would corrupt the stored victim.Name (player names are
    // canonicalized to capitalize-first by the loader at boot).
    arg1, rest := util.CaseArgument(argument)
    arg2, _ := util.CaseArgument(rest)
    arg2 = util.SmashTilde(arg2)
    if arg1 == "" || arg2 == "" {
        ch.Send("Syntax: pcrename <victim> <new name>\n\r")
        return
    }
    victim := WorldPcLookup(arg1)
    if victim == nil {
        ch.Send("No such player connected.\n\r")
        return
    }
    if victim.IsNPC() {
        ch.Send("You can't rename NPCs.\n\r")
        return
    }
    if ch.GetTrust() < victim.GetTrust() {
        ch.Send("I don't think they would like that!\n\r")
        return
    }
    // Charset / length validation happens inside RenamePlayerFile via
    // PlayerFilePath's validPlayerName regex. Same-name case (arg1
    // EqualFold arg2) is treated as a no-op — matches C's silent
    // "rename to self" behavior (the rename(2) syscall succeeds and the
    // file is unchanged).
    if strings.EqualFold(arg1, arg2) {
        ch.Send("Old and new names are identical.\n\r")
        return
    }
    if RenamePlayerFileFunc == nil {
        ch.Send("Pfile rename not wired.\n\r")
        return
    }
    if err := RenamePlayerFileFunc(arg1, arg2); err != nil {
        switch err {
        case persist.ErrInvalidName:
            ch.Send("Illegal name.\n\r")
        case persist.ErrSourcePfileNotFound:
            ch.Send("Source pfile not found.\n\r")
            util.Bug("DoPcrename: source pfile missing for %q", arg1)
        case persist.ErrDestPfileExists:
            ch.Send("That name already exists.\n\r")
        default:
            ch.Send("Couldn't rename the pfile.\n\r")
            util.Bug("DoPcrename: rename failed: %v", err)
        }
        return
    }
    victim.Name = arg2
    if SaveFunc != nil {
        SaveFunc(victim)
    }
    ch.Send("Character was renamed.\n\r")
}
```

**Wire-up:** the `act` package can't import `persist` directly (already established pattern via seams — see `SaveFunc`, `StartEditingFunc`). Add a new function-variable seam:

```go
// internal/act/olc.go (alongside SaveFunc)
//
// RenamePlayerFileFunc is the seam to persist.RenamePlayerFile, wired
// at boot. Tests can override to assert calls without touching disk.
var RenamePlayerFileFunc func(oldName, newName string) error
```

Boot wires it at `internal/boot/boot.go` adjacent to existing `act.SaveFunc` assignment:

```go
act.RenamePlayerFileFunc = func(oldName, newName string) error {
    return persist.RenamePlayerFile(dataDir, oldName, newName)
}
```

#### D4c — Wire `DoPcrename` into `meditArmName` PC branch

`internal/game/medit_arms.go:74-87`:

```go
func meditArmName(d *types.DescriptorData, victim *types.CharData, arg string) {
    arg = util.SmashTilde(strings.TrimSpace(arg))
    if arg == "" {
        d.WriteToBuffer("Name cannot be empty.\n\r")
        meditFinishArm(d, victim)
        return
    }
    // PC name change → route through DoPcrename for pfile rename.
    // Mirrors C omedit.c:1389-1395. The DoMedit entry trust gate (D3,
    // LEVEL_GREATER) already gates access to this arm for PC victims;
    // we DO NOT replicate the C `LEVEL_SUB_IMPLEM-1` gate here because
    // it would split routing (some PC name edits would skip the pfile
    // rename and silently desync) — see §Open Questions Q4.
    if !victim.IsNPC() {
        if PcrenameFunc != nil {
            // Call the seam directly with the raw old+new names (no
            // sprintf round-trip). Behavior matches DoPcrename's
            // public API; trust gates inside DoPcrename re-verify
            // d.Character vs victim.
            PcrenameFunc(d.Character, victim.Name+" "+arg)
        }
        olcLog(d, "MOB", "Changed name to %s", arg)
        meditFinishArm(d, victim)
        return
    }
    // NPC path unchanged.
    victim.Name = arg
    if victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
        victim.IndexData.PlayerName = victim.Name
    }
    olcLog(d, "MOB", "Changed name to %s", arg)
    meditFinishArm(d, victim)
}
```

`PcrenameFunc` is a new game-package seam (game can't import act):

```go
// internal/game/medit_parse.go (or alongside worldPcLookup)
// PcrenameFunc is wired at boot to act.DoPcrename.
var PcrenameFunc func(ch *types.CharData, argument string)
```

Boot wiring:

```go
game.PcrenameFunc = act.DoPcrename
```

---

## Task Groups

Every group follows test-first. Mutation verification uses `Edit`-only revert per `_shared.md` → Mutation Verification Safety. **NO** `git checkout` / `git restore` / `git stash` / `git reset --hard`.

### G1 — D1 `pulseSave` field initializer

**Deliverables:**

- Add `pulseSave: types.PULSE_SAVE,` to the struct literal in `NewGameLoop` at `internal/game/loop.go:84-96`.

**Tests** (`internal/game/loop_test.go`, NEW test):

- `TestNewGameLoop_PulseSaveInitialized` — construct via `NewGameLoop(w, reg, ch)`, assert `g.pulseSave == types.PULSE_SAVE`. Use unexported field access (test is in `package game`).

**Mutation gate:**

- M1: drop the new line from the struct literal → `TestNewGameLoop_PulseSaveInitialized` fails (expected `PULSE_SAVE`, got `0`). Revert via `Edit`.

**Acceptance:** A1.

### G2 — D2 `DoBlank` command

**Deliverables:**

- New file `internal/act/playercfg.go` (extending) with `DoBlank` per §D2 design.
- Boot reg in `internal/boot/boot.go` adjacent to DoGag's row.

**Tests** (`internal/act/playercfg_test.go`, extending):

- `TestDoBlank_TogglesOn` — fresh PCData, no `PLR_BLANK` → call DoBlank → `ch.Act.IsSet(PLR_BLANK)` true; output contains "will be inserted".
- `TestDoBlank_TogglesOff` — `PLR_BLANK` already set → call DoBlank → bit cleared; output contains "no longer be inserted".
- `TestDoBlank_NPCIsNoop` — `ch.Act.Set(ACT_IS_NPC)` → call DoBlank → no flag change, no output.
- `TestDoBlank_NilPCDataIsNoop` — non-NPC ch with `ch.PCData = nil` → call DoBlank → no panic, no flag change, no output.
- `TestBoot_BlankRegistered` — registry lookup for "blank" returns Command with `Level == 0`, `Position == POS_DEAD`, `DoFun != nil`.

**Mutation gates:**

- M2: replace `ch.Act.Set(...)` with `ch.Act.Remove(...)` → `TestDoBlank_TogglesOn` fails (bit not set).
- M3: replace `ch.Act.Remove(...)` with `ch.Act.Set(...)` → `TestDoBlank_TogglesOff` fails (bit still set).
- M4: drop the `IsNPC()` guard → `TestDoBlank_NPCIsNoop` fails (output present, bit toggled).
- M5: drop the `nil PCData` guard → would not panic on the test (ch.Act lives directly on CharData, not PCData) but a follow-up test with explicitly-nil PCData asserts the early-return path; covered by absence of "will be inserted" output. (Acknowledge in plan: M5 is weak — the guard is defense-in-depth for family-pattern consistency, not a load-bearing crash-blocker for THIS specific flag. Keep the guard; recognize the mutation gate is soft.)

**Acceptance:** A2, A3.

### G3 — D3a `worldPcLookup` + `WorldPcLookup` seams

**Deliverables:**

- `internal/game/medit_parse.go` — add `worldPcLookup` adjacent to `worldMobLookup`.
- `internal/act/olc_pclookup.go` (NEW) — add `WorldPcLookup` exported parallel.

**Tests** (`internal/game/medit_parse_test.go` AND `internal/act/olc_pclookup_test.go`):

For each lookup (game + act variants):

- `TestWorldPcLookup_FindsConnectedPc` — descriptor with `Connected == CON_PLAYING` + Character.Name "Eilidh" → lookup("Eilidh") returns that *CharData.
- `TestWorldPcLookup_IgnoresLinkdead` — descriptor with `Connected == CON_GET_NAME` (or any non-CON_PLAYING) → lookup returns nil.
- `TestWorldPcLookup_IgnoresNilCharacter` — descriptor with `Character == nil` → skipped.
- `TestWorldPcLookup_CaseInsensitive` — character name "Eilidh", lookup("eilidh") matches.
- `TestWorldPcLookup_NoMatch` — empty Descriptors slice or no name match → returns nil.
- `TestWorldPcLookup_NilWorld` — `worldRef == nil` (game) / `WorldRef == nil` (act) → returns nil (no panic).

**Mutation gates:**

- M6: replace `strings.EqualFold` with `==` → `TestWorldPcLookup_CaseInsensitive` fails.
- M7: drop the `Connected != CON_PLAYING` skip → `TestWorldPcLookup_IgnoresLinkdead` fails.
- M8: drop the `Character == nil` skip → `TestWorldPcLookup_IgnoresNilCharacter` fails (panic).
- M9: drop the `worldRef == nil` / `WorldRef == nil` skip → `TestWorldPcLookup_NilWorld` fails (panic).

**Acceptance:** A4.

### G4 — D3b `DoMedit` PC-by-name extension

**Deliverables:**

- Replace the rejection at `internal/act/olc_interactive.go:298-306` with the PC lookup + menu entry per §D3b.

**Tests** (`internal/act/olc_interactive_test.go`):

- `TestDoMedit_PcNameLooksUpConnectedPc` — install a fake PC via `act.WorldPcLookup` override → `DoMedit(ch, "Eilidh")` → `ch.Desc.Connected == CON_MEDIT`, `ch.Desc.Olc.Mode == MEDIT_PC_MAIN_MENU`, `ch.Desc.Olc.Target == victim`.
- `TestDoMedit_PcNameLookupNoMatch` — override returns nil → "No such player connected." message; descriptor state unchanged.
- `TestDoMedit_PcNameLookupBelowTrust` — caller is `LEVEL_HERO` (below LEVEL_GREATER) → "Huh?" rejection; descriptor state unchanged.
- `TestDoMedit_PcNameLookupNoDescriptor` — `ch.Desc == nil` → "No descriptor." rejection.
- `TestDoMedit_PcNameLookupAlreadyEditing` — `ch.Desc.Connected == CON_MEDIT` already → "already editing" rejection.
- E2E equivalent: extend `internal/testclient/medit_test.go` if mock-installable; otherwise rely on the unit-level coverage above and document.

**Mutation gates:**

- M10: drop the trust gate → `TestDoMedit_PcNameLookupBelowTrust` fails (no rejection).
- M11: drop the double-edit guard → `TestDoMedit_PcNameLookupAlreadyEditing` fails.
- M12: replace `MEDIT_PC_MAIN_MENU` with `MEDIT_NPC_MAIN_MENU` → `TestDoMedit_PcNameLooksUpConnectedPc` fails (wrong mode).

**Acceptance:** A5, A6.

### G5 — D4a `persist.RenamePlayerFile`

**Deliverables:**

- Add `RenamePlayerFile`, `ErrInvalidName`, `ErrSourcePfileNotFound`, `ErrDestPfileExists` to `internal/persist/player.go`.

**Tests** (`internal/persist/player_test.go`, extending):

- `TestRenamePlayerFile_HappyPath` — write fixture pfile at `<tmp>/player/e/Eilidh` → call rename → `<tmp>/player/b/Bob` exists with same content; `<tmp>/player/e/Eilidh` does not exist.
- `TestRenamePlayerFile_HappyPathSameSubdir` — rename "Eilidh" → "Edward" (both 'e'-prefix) → file moves within same subdir.
- `TestRenamePlayerFile_DifferentSubdirCreatesIt` — rename "Alice" → "Bob" when `<tmp>/player/b/` does not yet exist → MkdirAll succeeds + file lands.
- `TestRenamePlayerFile_SourceMissing` → `ErrSourcePfileNotFound`.
- `TestRenamePlayerFile_DestExists` — pre-create destination → `ErrDestPfileExists`; original source unchanged.
- `TestRenamePlayerFile_InvalidOldName` — old="A" (too short) → `ErrInvalidName`.
- `TestRenamePlayerFile_InvalidNewName` — new="bad name" (space) → `ErrInvalidName`.
- `TestRenamePlayerFile_InvalidNewNameTraversal` — new="../etc/passwd" → after `filepath.Base` becomes "passwd" (legal alphabetic 6-char name); rename succeeds and lands at `<tmp>/player/p/passwd`. **This test PINS the documented behavior** — the traversal is neutralized by `Base` + regex; the result is contained inside the player tree. Assert that no file was created OUTSIDE `<tmp>/player/`.
- `TestRenamePlayerFile_InvalidNewNameAbsolute` — new="/etc/passwd" → `Base` returns "passwd"; same containment as above.
- `TestRenamePlayerFile_InvalidNewNameTooLong` — new="ThirteenCharsX" (13 chars) → `ErrInvalidName`.
- `TestRenamePlayerFile_InvalidNewNameNonAlpha` — new="bo.b" (dot) → `ErrInvalidName`.
- `TestRenamePlayerFile_EmptyNames` — old="" → `ErrInvalidName`; new="" → `ErrInvalidName`.

**Mutation gates:**

- M13: drop the `os.Stat(oldPath)` source-existence check → `TestRenamePlayerFile_SourceMissing` fails (test asserts specific sentinel; OS would return a different error from os.Rename).
- M14: drop the `os.Stat(newPath)` destination-existence check → `TestRenamePlayerFile_DestExists` fails (rename would overwrite silently).
- M15: drop the `os.MkdirAll` for destination subdir → `TestRenamePlayerFile_DifferentSubdirCreatesIt` fails (rename to nonexistent dir errors).
- M16: replace `PlayerFilePath(dataDir, oldName)` source with literal `oldName` (skipping path resolution) → multiple tests fail.

**Acceptance:** A7, A8, A9.

### G6 — D4b `act.DoPcrename` + seam wiring

**Deliverables:**

- New `DoPcrename` in `internal/act/playercfg.go`.
- New `RenamePlayerFileFunc` seam in `internal/act/olc.go` (alongside other seams).
- Boot wiring in `internal/boot/boot.go`.

**Tests** (`internal/act/playercfg_test.go`, extending):

- `TestDoPcrename_HappyPath` — connected PC "Eilidh" via `WorldPcLookup` override; `RenamePlayerFileFunc` recorder asserts called with ("Eilidh", "Bob"); after call `victim.Name == "Bob"`; "Character was renamed." emitted; `SaveFunc` recorder asserts called once with victim.
- `TestDoPcrename_NPCCallerIsNoop` — `ch.IsNPC()` true → silent return.
- `TestDoPcrename_EmptyArgs` → "Syntax: pcrename ..." (test "Eilidh" alone → still syntax message because arg2 missing).
- `TestDoPcrename_VictimNotFound` — `WorldPcLookup` returns nil → "No such player connected.".
- `TestDoPcrename_VictimIsNPC` — overridden lookup returns NPC → "You can't rename NPCs.".
- `TestDoPcrename_BelowTrust` — `ch.GetTrust() < victim.GetTrust()` → "I don't think they would like that!".
- `TestDoPcrename_SameName` — old=new (case-insensitive) → "Old and new names are identical."; rename function NOT called.
- `TestDoPcrename_RenameFuncErrInvalidName` — RenamePlayerFileFunc returns ErrInvalidName → "Illegal name." emitted; `victim.Name` unchanged; `SaveFunc` NOT called.
- `TestDoPcrename_RenameFuncErrSourceMissing` → "Source pfile not found." + `util.Bug` recorded.
- `TestDoPcrename_RenameFuncErrDestExists` → "That name already exists.".
- `TestDoPcrename_RenameFuncGenericError` → "Couldn't rename the pfile.".
- `TestDoPcrename_NoSeamWired` — `RenamePlayerFileFunc == nil` → "Pfile rename not wired.".
- `TestDoPcrename_E2EOnDisk` — full integration: stand up tmp dir + write fixture pfile + wire real `persist.RenamePlayerFile` → `DoPcrename(ch, "Eilidh Bob")` → assert pfile moved on disk; in-memory `victim.Name == "Bob"`. **This is the load-bearing on-disk round-trip test.**

**Mutation gates:**

- M17: drop the `ch.IsNPC()` guard → `TestDoPcrename_NPCCallerIsNoop` fails.
- M18: drop the `victim.IsNPC()` reject → `TestDoPcrename_VictimIsNPC` fails.
- M19: drop the trust comparison → `TestDoPcrename_BelowTrust` fails.
- M20: replace `victim.Name = arg2` with `victim.Name = arg1` → `TestDoPcrename_HappyPath` fails (in-memory not updated to new name).
- M21: move `victim.Name = arg2` BEFORE `RenamePlayerFileFunc(...)` → `TestDoPcrename_RenameFuncErrInvalidName` fails (in-memory updated even though rename failed → desync).
- M22: drop the `SaveFunc(victim)` call → `TestDoPcrename_HappyPath` fails (SaveFunc recorder asserts called).
- M23: drop the `SmashTilde(arg2)` → add a tilde-containing test that asserts the saved name has no tilde; mutation skips the scrub. **(This test is in the test list above as part of HappyPath if extended; or split out as `TestDoPcrename_SmashTilde`.)** Plan: add explicit `TestDoPcrename_SmashTilde`.

**Acceptance:** A10, A11, A12, A13, A14.

### G7 — D4c medit MEDIT_NAME PC arm wire-up

**Deliverables:**

- Modify `meditArmName` at `internal/game/medit_arms.go:74-87` per §D4c.
- New `PcrenameFunc` seam in `internal/game/medit_parse.go` (alongside `worldMobLookup`).
- Boot wiring `game.PcrenameFunc = act.DoPcrename`.

**Tests** (`internal/game/medit_arms_test.go` or `medit_wave3_test.go` extending):

- `TestMeditArmName_PcRoutesToPcrename` — PC victim, `PcrenameFunc` recorder set → `meditArmName(d, victim, "Bob")` → recorder called once with `(d.Character, "Eilidh Bob")` (assuming victim.Name was "Eilidh"); olcLog fired.
- `TestMeditArmName_NpcUnchanged` — NPC victim → recorder NOT called; `victim.Name == "Bob"` direct assignment; `victim.IndexData.PlayerName == "Bob"` if prototype.
- `TestMeditArmName_PcrenameFuncNil` — PcrenameFunc nil → no panic; olcLog still fires; meditFinishArm completes.

**Mutation gates:**

- M24: invert the `victim.IsNPC()` branch → `TestMeditArmName_PcRoutesToPcrename` and `_NpcUnchanged` both fail (wrong path taken).
- M25: replace `d.Character` with `victim` in the PcrenameFunc call → recorder asserts caller identity → fails.
- M26: replace `victim.Name+" "+arg` with `arg+" "+victim.Name` → recorder asserts argument format → fails.

**Acceptance:** A15.

### G8 — Boot wire-up for D2 + D4 seams

**Deliverables:**

- `internal/boot/boot.go` additions:
  - `act.RenamePlayerFileFunc = func(o, n string) error { return persist.RenamePlayerFile(dataDir, o, n) }`
  - `game.PcrenameFunc = act.DoPcrename`
  - Command registry row for `blank` → `act.DoBlank`.

**Tests** (`internal/boot/boot_test.go`):

- `TestBoot_BlankRegistered` (already in G2 deliverables — single test serves both groups).
- `TestBoot_PcrenameFuncWired` — after boot, assert `game.PcrenameFunc != nil` and is `act.DoPcrename` (function-identity check via reflect.ValueOf().Pointer()).
- `TestBoot_RenamePlayerFileFuncWired` — assert `act.RenamePlayerFileFunc != nil`.

**Mutation:**

- Drop the `game.PcrenameFunc` assignment → `TestBoot_PcrenameFuncWired` fails.
- Drop the `act.RenamePlayerFileFunc` assignment → `TestBoot_RenamePlayerFileFuncWired` fails.

**Acceptance:** A16.

### G9 — Documentation closeout

**Deliverables:**

- `CHANGELOG.md` entry for the lineage (single date-grouped block).
- This plan's §Completion Record appended (commit hashes, mutation gates exercised, file deltas).
- `TODO.md`:
  - Move to Done: line ~72 (medit MEDIT_NAME PC branch — do_pcrename), line ~92 (worldPcLookup seam), line ~201 (DoBlank toggle), line ~320 (pulseSave init bug).
- `phases.md`: append landing row for this lineage in the §Phase 6 board.

**Acceptance:** A17.

---

## Acceptance Criteria (binary)

| # | Criterion | Gate |
|---|---|---|
| A1 | `pulseSave` initialized to `PULSE_SAVE` after `NewGameLoop()` | `TestNewGameLoop_PulseSaveInitialized` |
| A2 | `DoBlank` toggles `PLR_BLANK` with directional messages; NPC + nil-PCData no-op | `TestDoBlank_*` (4 tests) |
| A3 | `blank` registered at level 0 / POS_DEAD | `TestBoot_BlankRegistered` |
| A4 | `worldPcLookup` / `WorldPcLookup` resolve connected PCs by case-insensitive name; skip linkdead/nil-Character | `TestWorldPcLookup_*` (12 tests, 6 per package) |
| A5 | `DoMedit` non-numeric arg looks up PC and enters MEDIT_PC_MAIN_MENU | `TestDoMedit_PcNameLooksUpConnectedPc` |
| A6 | `DoMedit` PC lookup enforces LEVEL_GREATER, double-edit guard, descriptor presence | `TestDoMedit_PcNameLookup{NoMatch,BelowTrust,NoDescriptor,AlreadyEditing}` |
| A7 | `RenamePlayerFile` happy path moves pfile across subdirs; happy path same subdir | `TestRenamePlayerFile_HappyPath{,SameSubdir,DifferentSubdirCreatesIt}` |
| A8 | `RenamePlayerFile` returns sentinel errors for source-missing / dest-exists / invalid names | `TestRenamePlayerFile_{SourceMissing,DestExists,Invalid*}` |
| A9 | `RenamePlayerFile` containment: traversal/absolute names neutralized by `Base` + regex | `TestRenamePlayerFile_InvalidNewName{Traversal,Absolute}` |
| A10 | `DoPcrename` happy path renames in-memory + on-disk; emits "Character was renamed." | `TestDoPcrename_HappyPath` + `TestDoPcrename_E2EOnDisk` |
| A11 | `DoPcrename` rejects NPC caller, NPC victim, missing victim, below-trust, same-name | `TestDoPcrename_{NPCCallerIsNoop,VictimNotFound,VictimIsNPC,BelowTrust,SameName,EmptyArgs}` |
| A12 | `DoPcrename` does NOT update `victim.Name` when rename fails | `TestDoPcrename_RenameFuncErrInvalidName` |
| A13 | `DoPcrename` calls `SaveFunc` after successful rename | `TestDoPcrename_HappyPath` (SaveFunc recorder) |
| A14 | `DoPcrename` smash-tildes the new name | `TestDoPcrename_SmashTilde` |
| A15 | medit MEDIT_NAME PC arm routes through `PcrenameFunc`; NPC arm unchanged | `TestMeditArmName_{PcRoutesToPcrename,NpcUnchanged,PcrenameFuncNil}` |
| A16 | Boot wires `RenamePlayerFileFunc` + `game.PcrenameFunc` | `TestBoot_{PcrenameFuncWired,RenamePlayerFileFuncWired}` |
| A17 | Docs updated (CHANGELOG + TODO + plan §Completion Record + phases.md) | manual verification |

---

## Mutation Gates (≥20, all `Edit`-round-trip)

| # | Mutation | Pinning test |
|---|---|---|
| M1 | Drop `pulseSave: types.PULSE_SAVE,` from `NewGameLoop` | `TestNewGameLoop_PulseSaveInitialized` |
| M2 | Replace `Set(PLR_BLANK)` with `Remove` in DoBlank | `TestDoBlank_TogglesOn` |
| M3 | Replace `Remove(PLR_BLANK)` with `Set` in DoBlank | `TestDoBlank_TogglesOff` |
| M4 | Drop `IsNPC()` guard in DoBlank | `TestDoBlank_NPCIsNoop` |
| M5 | Drop nil-PCData guard in DoBlank | (soft — see G2) |
| M6 | Replace `EqualFold` with `==` in WorldPcLookup | `TestWorldPcLookup_CaseInsensitive` |
| M7 | Drop `Connected != CON_PLAYING` skip | `TestWorldPcLookup_IgnoresLinkdead` |
| M8 | Drop nil-Character skip | `TestWorldPcLookup_IgnoresNilCharacter` |
| M9 | Drop nil-WorldRef guard | `TestWorldPcLookup_NilWorld` |
| M10 | Drop trust gate in DoMedit PC-arg branch | `TestDoMedit_PcNameLookupBelowTrust` |
| M11 | Drop double-edit guard in DoMedit PC-arg branch | `TestDoMedit_PcNameLookupAlreadyEditing` |
| M12 | Set MEDIT_NPC_MAIN_MENU instead of MEDIT_PC_MAIN_MENU | `TestDoMedit_PcNameLooksUpConnectedPc` |
| M13 | Drop source-existence check in RenamePlayerFile | `TestRenamePlayerFile_SourceMissing` |
| M14 | Drop dest-existence check | `TestRenamePlayerFile_DestExists` |
| M15 | Drop MkdirAll | `TestRenamePlayerFile_DifferentSubdirCreatesIt` |
| M16 | Replace `PlayerFilePath(...)` with literal name | multiple |
| M17 | Drop `IsNPC()` caller guard in DoPcrename | `TestDoPcrename_NPCCallerIsNoop` |
| M18 | Drop `victim.IsNPC()` reject | `TestDoPcrename_VictimIsNPC` |
| M19 | Drop trust comparison in DoPcrename | `TestDoPcrename_BelowTrust` |
| M20 | `victim.Name = arg2` → `victim.Name = arg1` | `TestDoPcrename_HappyPath` |
| M21 | Move `victim.Name = arg2` BEFORE rename call | `TestDoPcrename_RenameFuncErrInvalidName` |
| M22 | Drop `SaveFunc(victim)` | `TestDoPcrename_HappyPath` |
| M23 | Drop `SmashTilde` | `TestDoPcrename_SmashTilde` |
| M24 | Invert NPC branch in meditArmName | `TestMeditArmName_PcRoutesToPcrename` |
| M25 | Pass `victim` instead of `d.Character` to PcrenameFunc | `TestMeditArmName_PcRoutesToPcrename` |
| M26 | Swap arg order to `arg+" "+victim.Name` | `TestMeditArmName_PcRoutesToPcrename` |

26 mutation gates total — well above the ≥10 floor.

---

## Scope Cuts / Deferrals

- **`do_pcrename` god-dir cleanup** (`act_wiz.c:12731-12734`) — wizlist/god-dir port deferred independently.
- **`do_pcrename` per-immortal area-file rename** (`act_wiz.c:12738-12749`) — Go has no `pcdata.area` ownership.
- **`do_pcrename` `.bak` file cleanup** (`act_wiz.c:12755`) — Go pfile has no `.bak` analog.
- **`make_wizlist` rebuild on rename** (`act_wiz.c:12766-12767`) — wizlist port deferred.
- **`get_char_room` security precaution** — replaced with `WorldPcLookup` per user directive; recorded as divergence.
- **Linkdead victim rename** — explicitly out of scope. The Go `WorldPcLookup` only finds CON_PLAYING descriptors. Renaming a linkdead PC requires walking pfiles on disk (no in-memory victim) and is a separate feature. A linkdead PC's pfile will be renamed by a sysop using filesystem tools or a future offline-rename utility.
- **Standalone `pcrename` command at command line** — `DoPcrename` is exported and wired as a seam, but NOT registered as a top-level command in this lineage. The C `do_pcrename` is registered (`commands.dat`) but the Go port reaches it solely via the medit MEDIT_NAME PC arm. Adding a top-level `pcrename` command is a single boot-reg row and could be queued as a follow-up.
- **C-bug-not-preserved at `omedit.c:1390`** — C silently desyncs in-memory PC name from pfile when caller is below `LEVEL_SUB_IMPLEM`; Go always routes through `DoPcrename` (see §Open Questions Q4).

---

## Open Questions

| # | Question | Recommendation |
|---|---|---|
| Q1 | Should `DoBlank` use `ch.Act` (PLR_BLANK) or `ch.PCData.Flags` (some other PCFLAG_)? | **`ch.Act`** — verified `PLR_BLANK` is at `enums.go:964` in the PLR_* family, written through `ch.Act` (BitVector). Tranche C G6 renderer reads from the same location. The nil-PCData guard is kept for family consistency even though `Act` is on `CharData` directly. |
| Q2 | Should `WorldPcLookup` live in `act` or `game` or both? | **Both.** Tiny duplication is cheaper than a third package or an act→game seam. `act` cannot import `game` (cycle); a single `game` location forces a function-pointer seam (extra wiring) for every external caller. Two co-equal definitions, each tested, is simpler. |
| Q3 | Should `DoMedit` PC-by-name use `LEVEL_GREATER` or mirror C's configurable `sysdata.level_modify_proto`? | **`LEVEL_GREATER`** per user directive. The configurable sysdata isn't ported yet; hardcoded `LEVEL_GREATER` matches Go-port hardening tone (existing `DoMedit` entry uses `LEVEL_IMMORTAL`; PC editing tightens further). When sysdata lands, this can be re-pointed. |
| Q4 | Should `meditArmName` PC arm replicate C's `LEVEL_SUB_IMPLEM-1` gate at `omedit.c:1390`? | **No.** C's gate creates a known desync (PC name changes in-memory but pfile stays under old name) when caller is below SUB_IMPLEM. Below LEVEL_SUB_IMPLEM in C, the rename "succeeds" cosmetically but corrupts the pfile mapping. The Go port routes ALL PC name changes through `DoPcrename` so the in-memory and on-disk names stay in sync. The trust gate at `DoMedit` entry (D3, LEVEL_GREATER) already protects; LEVEL_SUB_IMPLEM is a lower bar than LEVEL_GREATER (`MAX_LEVEL-4` vs `MAX_LEVEL-6` — wait, `LEVEL_SUB_IMPLEM = MAX_LEVEL-4` and `LEVEL_GREATER = MAX_LEVEL-6`, so SUB_IMPLEM is HIGHER. The user-stated gate at DoMedit entry is LOWER than the C in-arm gate. Recheck.) — **CORRECTION:** with LEVEL_GREATER < LEVEL_SUB_IMPLEM, a Go caller at LEVEL_GREATER could enter the PC menu but in C would NOT have hit the rename path. Two responses: (a) raise the DoMedit PC-arg trust gate to LEVEL_SUB_IMPLEM (matches C's effective gate); or (b) keep LEVEL_GREATER and accept that Go ports more rename privilege earlier. Recommendation: **(a) — raise to LEVEL_SUB_IMPLEM at the DoMedit PC-arg trust gate.** This pins the gate to C's effective threshold for PC renaming. Update §D3b accordingly. |
| Q5 | What happens if `DoPcrename` is called with a victim currently being edited in another descriptor's medit session? | **Out of scope.** The in-memory `victim.Name` mutation will affect both descriptors atomically (single goroutine model); the editing descriptor's stale `Olc.Target` pointer remains valid and the next menu redisplay reads the fresh name. No corruption. Document, do not block. |
| Q6 | Should `DoPcrename` reject if victim has unsaved changes or open editor session? | **No.** The rename writes through `os.Rename` which is atomic at the syscall level; subsequent saves go to the new path. An editor session continues to write its buffer to victim.Description (or wherever) — fine. |
| Q7 | Should the `.bak` rotation (foldarea-style) apply to pfile rename? | **No.** D4 is a rename, not a write. The source pfile becomes the destination — no overwrite, no need for a backup. (If dest already exists we abort with `ErrDestPfileExists`.) |

**Open Question Q4 changes the trust gate in §D3b** — implementation must use `LEVEL_SUB_IMPLEM` instead of `LEVEL_GREATER` at the DoMedit PC-arg branch. Test `TestDoMedit_PcNameLookupBelowTrust` updated to use a caller at `LEVEL_GREATER` (which would now be below the new gate).

---

## File Budget

| File | Change | Estimate |
|---|---|---|
| `internal/game/loop.go` | +1 LOC (D1) | +1 LOC |
| `internal/game/loop_test.go` | New test | +25 LOC |
| `internal/act/playercfg.go` | `DoBlank` + `DoPcrename` | +120 LOC |
| `internal/act/playercfg_test.go` | DoBlank + DoPcrename tests | +400 LOC |
| `internal/act/olc.go` | +`RenamePlayerFileFunc` seam | +5 LOC |
| `internal/act/olc_pclookup.go` (NEW) | `WorldPcLookup` | +30 LOC |
| `internal/act/olc_pclookup_test.go` (NEW) | 6 tests | +200 LOC |
| `internal/act/olc_interactive.go` | DoMedit non-numeric arg branch | +35 LOC, -5 LOC |
| `internal/act/olc_interactive_test.go` | 5 PC-name tests | +200 LOC |
| `internal/game/medit_parse.go` | `worldPcLookup` + `PcrenameFunc` seam | +40 LOC |
| `internal/game/medit_parse_test.go` | 6 lookup tests | +200 LOC |
| `internal/game/medit_arms.go` | `meditArmName` PC branch | +20 LOC, -5 LOC |
| `internal/game/medit_arms_test.go` (or wave3_test) | 3 medit-arm tests | +120 LOC |
| `internal/persist/player.go` | `RenamePlayerFile` + sentinels | +60 LOC |
| `internal/persist/player_test.go` | ~12 rename tests | +400 LOC |
| `internal/boot/boot.go` | 3 wirings (blank reg + 2 seams) | +6 LOC |
| `internal/boot/boot_test.go` | 3 boot-reg tests | +60 LOC |
| `CHANGELOG.md` (lineage draft) | 1 entry | +10 LOC |
| `TODO-updates.md` (lineage draft) | 4 Done moves | +6 LOC |
| `smaug-go/doc/plan-phase6-quickwins-blank-pcrename.md` | This plan + §Completion Record | +700 LOC (this commit) + ~80 LOC (landing) |

Total implementation churn (excluding plan + docs): ~+1,800 LOC across 9 source files (incl. tests; production code is ~+260 LOC).

---

## Wave Plan

Per dispatch directive:

| Wave | Groups | Workers | Notes |
|---|---|---|---|
| 1 | G1 | 1 | `pulseSave` field initializer — smallest, closes the bug first. |
| 2 | G2 | 1 | `DoBlank` + boot reg. Independent of D3/D4. |
| 3 | G3 | 1 | `worldPcLookup` / `WorldPcLookup` — D4 depends on this. |
| 4 | G4, G5, G6, G7, G8 | 1-3 (sequential within wave; some parallelism possible across non-overlapping files) | `DoMedit` PC-arg branch + `RenamePlayerFile` + `DoPcrename` + medit-arm wire-up + boot wirings. |
| 5 | G9 | Manager (docs only) | Closeout. |

Per-wave gates:
- `cd smaug-go && go test -count=3 ./...` must be green across all 15 packages.
- Pre-commit hook stays enabled. `SMAUG_SKIP_TESTS=1` is the only sanctioned escape valve; full suite re-run manually after commit.
- Commit messages follow `Phase 6 quickwins Wave N: <summary>` shape.
- NO `git checkout` / `git restore` / `git stash` / `git reset --hard` during mutation verification.

---

## Completion Record

(Populated at landing. Wave commit hashes, mutation gates exercised, file deltas, LOC, test counts, deferrals queued.)
