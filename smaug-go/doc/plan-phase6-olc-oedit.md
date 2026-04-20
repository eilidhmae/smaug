# Plan: Phase 6 Interactive OLC — `CON_OEDIT` Substate (Menu-Driven Object Editor)

**Status:** Authored 2026-04-19. Pending adversary review before dispatch.
**Priority:** P2 (Phase 6, Wave D — Interactive OLC, second of three editors).
**Scope:** Interactive object-editor nanny substate `CON_OEDIT`. Adds a menu-driven object editor on top of the existing flat `oedit <vnum> <sub> [args]` command (Phase 4b / `DoOedit` at `internal/act/olc_interactive.go:16`). Inherits the nanny-dispatch pattern + `OlcData`-on-descriptor allocation + `CON_*` loop arm established by `plan-phase6-olc-redit.md` (LANDED 2026-04-19, commit `f678e8a`).

---

## Cross-Plan Dependencies

**Depends on (LANDED):**

- `plan-phase6-olc-redit.md` commit `f678e8a` 2026-04-19 — establishes `OlcData` struct, `OlcData.Mode` iota-space convention (REDIT_MAIN_MENU=iota+100; OEDIT_* will use iota+200 per §G1 below), `DescriptorData.Olc` field, `CON_REDIT` arm in `internal/game/loop.go:273-278` dispatch switch, `ReditDispMenuFunc` seam pattern in `internal/boot/boot.go:151-154`, `olcLog` helper in `internal/game/redit_parse.go:690-712`, `worldRoomLookup` seam at `:718-723`, `uRange` helper at `:725-734`, EditorSave restore-CON_REDIT trampoline pattern at `:163-184` (copy pattern for CON_OEDIT restore).
- `plan-phase5-tier12-editor-save.md` LANDED — `EditorSave` callback on `CharData` + `/s` handler in `internal/game/editor.go:163-172` that transitions `Connected=CON_PLAYING` BEFORE invoking the closure (closure re-sets CON_OEDIT after `StopEditing`, same contract as redit §G5).
- Existing flat `DoOedit` at `internal/act/olc_interactive.go:16-200` (including `oeditAffects` at :179) — flat path is preserved; menu-entry path is added when `argument == ""` and `GetTrust() >= LEVEL_IMMORTAL`.
- Existing `DoOset` at `internal/act/olc_set.go:208` — flat field-setting logic is reused inside the parser where it simplifies porting (e.g. the same `oeditSetField` helpers can back both flat + menu paths). Menu arms may call into these existing helpers rather than re-porting field mutation logic.

**Blocks (hard):**

- `plan-phase6-olc-medit.md` — third OLC plan. Will inherit every structural decision from this plan (OEDIT_* iota-space convention → MEDIT_* = iota+300; `OeditDispMenuFunc` seam shape → `MeditDispMenuFunc`; value-menu per-item-type dispatch shape is analogous to medit's per-ACT_*-class customization but not identical). No re-derivation; medit plan will reference this one the way this one references redit.
- `plan-phase6-olc-mpedit.md` — object-mudprog-editor. Hard-blocked by the `OEDIT_MPROGS*` family being scope-cut from this plan (see §Scope Cuts). mpedit must land AFTER oedit + medit establish both menu shapes; OEDIT_MPROGS / OEDIT_MPROGS_* mode constants are reserved here (iota slot allocated in §G1) so mpedit does not need to renumber.

**Soft dependencies (safe to ignore):**

- Phase 4b `DoOcreate` — already covers new-object creation via flat `oedit <vnum> create <name>` path. Menu-entry CANNOT create new objects (C `do_ooedit` at `:129-134` prints `"OEdit what?"` and returns when no argument given — ENABLE_OLC2_EXTRAS is OFF by default). Match C.
- `plan-phase6-olc-redit.md` §G12 testclient E2E pattern — reused verbatim for oedit E2E scenarios (match on menu-text substrings, no `> ` prompt sentinel).

---

## Problem

The constant `CON_OEDIT` is defined in `internal/types/enums.go:90` (iota value 22). The pulse-loop dispatch switch in `internal/game/loop.go:273-278` has a `case types.CON_REDIT:` arm (landed Phase-6 olc-redit) but **no `case types.CON_OEDIT:` arm**. Any descriptor that transitions to `CON_OEDIT` falls through to the nanny `default:` branch and is disconnected with "Unexpected state. Disconnecting.".

The flat command path `DoOedit` at `internal/act/olc_interactive.go:16` implements ~15 subcommands (`create`, `name`, `short`, `long`, `type`, `flags`, `wearflags`, `values`, `weight`, `cost`, `level`, `affects` + sub-args `add`/`del`, `ed`, `delete`, `show`). It prints a usage line when invoked without a subcommand. It does **not** enter a menu substate. A builder who wants to change multiple fields re-types `oedit <vnum> ...` each time.

The C flow is the opposite. `do_ooedit` at `src/ooedit.c:109-208` sets `d->connected = CON_OEDIT`, stashes the object on `d->character->dest_buf`, and calls `oedit_disp_menu`. Every subsequent line arrives at `oedit_parse` at `:1172-2160` and is dispatched against `OLC_MODE(d)` — a per-descriptor menu-state enum with **37 OEDIT_* values** (see `src/olc.h:114-150`).

C dispatch at `src/smaug.c:1641-1652` routes `d->connected == CON_OEDIT` to `oedit_parse(d, argument)`. The nanny is NOT involved; CON_OEDIT short-circuits it — same pattern as CON_REDIT.

Consequence of the Go gap: no interactive object editor exists. Builders use the flat `oedit <vnum> ...` path or nothing. This plan adds the menu-driven editor while keeping the flat path untouched.

## C Reference (authoritative)

All line numbers verified 2026-04-19 from `src/ooedit.c` (2160 LOC) and `src/olc.h` (293 LOC).

### Entry points

- **`do_ooedit`** — `src/ooedit.c:109-208`. Parses optional vnum, checks `IS_NPC` (reject), `can_omodify`, guards against double-edit (`d->connected == CON_OEDIT && OLC_VNUM(d) == obj->pIndexData->vnum`), allocates `OLC_DATA`, sets `d->connected = CON_OEDIT`, `d->character->dest_buf = obj`, calls `oedit_disp_menu(d)`.
- **`do_oedit_reset`** — `src/ooedit.c:1099-1167` (stub in ooedit, body lives after `oedit_parse`). The `last_cmd` callback invoked by `edit_buffer` on `/s`. Substates in C: `SUB_OBJ_LONG` (updates `obj->description`), `SUB_OBJ_EXTRA` (updates extradesc `ed->description`), `SUB_MPROG_EDIT` (updates mprog `comlist`). **NOTE: C does NOT have a SUB_OBJ_ACTION case** — option 4 (action desc) accepts inline single-line input in C, not a text editor. Go-port ADDS a text-editor path for action desc using a new `SUB_OBJ_ACTION` constant (Go-port divergence, documented). SUB_MPROG_EDIT scope-cut to mpedit.
- **`oedit_parse`** — `src/ooedit.c:1172-2160`. Top-level state machine keyed on `OLC_MODE(d)`. Every case mutates the object and either returns (stay in same mode) or `break`s (falls to the end, sets `OLC_CHANGE(d) = TRUE`, re-displays `oedit_disp_menu(d)`).
- **C dispatch call site** — `src/smaug.c:1641-1652`. Same site that routes CON_REDIT.

### OEDIT_* mode constants (from `src/olc.h:114-150`)

```
OEDIT_MAIN_MENU                1
OEDIT_EDIT_NAMELIST            2
OEDIT_SHORTDESC                3
OEDIT_LONGDESC                 4
OEDIT_ACTDESC                  5
OEDIT_TYPE                     6
OEDIT_EXTRAS                   7
OEDIT_WEAR                     8
OEDIT_WEIGHT                   9
OEDIT_COST                    10
OEDIT_COSTPERDAY              11
OEDIT_TIMER                   12
OEDIT_VALUE_1                 13
OEDIT_VALUE_2                 14
OEDIT_VALUE_3                 15
OEDIT_VALUE_4                 16
OEDIT_VALUE_5                 17
OEDIT_VALUE_6                 18
OEDIT_EXTRADESC_KEY           19
OEDIT_CONFIRM_SAVEDB          20  (unused in stock code; reserve only)
OEDIT_CONFIRM_SAVESTRING      21  (commented out at :1208-1209; reserve only)
OEDIT_EXTRADESC_DESCRIPTION   22
OEDIT_EXTRADESC_MENU          23
OEDIT_LEVEL                   24
OEDIT_LAYERS                  25
OEDIT_AFFECT_MENU             26
OEDIT_AFFECT_LOCATION         27
OEDIT_AFFECT_MODIFIER         28
OEDIT_AFFECT_REMOVE           29
OEDIT_AFFECT_RIS              30
OEDIT_EXTRADESC_CHOICE        31
OEDIT_EXTRADESC_DELETE        32
OEDIT_MPROGS                  33  (scope-cut to mpedit; reserve iota slot)
OEDIT_MPROGS_CHOICE           34  (scope-cut)
OEDIT_MPROGS_DELETE           35  (scope-cut)
OEDIT_MPROGS_TYPE             36  (scope-cut)
OEDIT_MPROGS_ARG              37  (scope-cut)
```

### Menu rendering functions

| C function | Line | Purpose |
|---|---|---|
| `oedit_disp_menu` | `:921-985` | Main menu (17 rendered options: 1-9, A-G, Q) — name, s-desc, l-desc, a-desc, type, extras, wear, weight, cost, rent, timer, level, layers, values, affect, extradesc, quit. `H` (mprog) is dispatched in the parser but NOT rendered in the menu display string (audit correction 2026-04-19 — mprog lives under `CON_MPROG_EDIT` per olc-mpedit plan). |
| `oedit_disp_type_menu` | `:852-868` | Item-type menu (0..MAX_ITEM_TYPE, `o_types[]` in 3 cols) |
| `oedit_disp_extra_menu` | `:871-890` | Extra-flags menu (MAX_ITEM_FLAG labels, 2 cols, with current `ext_flag_string`) |
| `oedit_disp_wear_menu` | `:895-917` | Wear-flags menu (ITEM_WEAR_MAX labels, 2 cols, skipping ITEM_DUAL_WIELD bit) |
| `oedit_disp_layer_menu` | `:449-466` | 9-option layer list (Nothing / Silk Shirt / Leather Vest / Light Chainmail / Leather Jacket / Light Cloak / Loose Cloak / Cape / Magical Effects) |
| `oedit_disp_val1_menu` .. `val6_menu` | `:623-849` | Per-slot value prompt dispatched on `obj->item_type` (see §Item-Type Value Dispatch below) |
| `oedit_disp_container_flags_menu` | `:407-424` | Container flags (5 labels) |
| `oedit_disp_lever_flags_menu` | `:429-444` | Lever/switch trigger flags (29 labels, `trig_flags[]`) |
| `oedit_liquid_type` | `:547-564` | Liquid-type list for drink containers (LIQ_MAX entries, 3 cols) |
| `oedit_disp_weapon_menu` | `:597-614` | Weapon-type menu (18 `attack_table` entries, 2 cols) |
| `oedit_disp_spells_menu` | `:617-620` | One-liner "Enter the name of the spell:" prompt |
| `oedit_disp_extradesc_menu` | `:469-502` | Extradesc list + A/R/Q options (iterates both `pIndexData->first_extradesc` AND `obj->first_extradesc`) |
| `oedit_disp_extra_choice` | `:504-514` | Per-extradesc sub-menu (1 keyword / 2 description / Q) |
| `oedit_disp_prompt_apply_menu` | `:517-544` | Affect list + A/R/Q options (iterates both pIndexData AND obj affects) |
| `oedit_disp_affect_menu` | `:569-591` | Apply-type menu (MAX_APPLY_TYPE entries, 3 cols, skipping 0 and APPLY_EXT_AFFECT) |
| `medit_disp_aff_flags` | *external* | Affect-flag bitmask editor (shared with medit — extern at ooedit:59) |
| `medit_disp_ris` | *external* | RIS bitmask editor (shared with medit — extern at ooedit:60) |

### Item-Type Value Dispatch (the oedit-specific complexity)

C dispatches value-slot prompts by `obj->item_type`. The table below pins each branch verbatim from `src/ooedit.c:623-849`. Each Go arm in `§G4` must cover every ITEM_TYPE_* entry and either set the corresponding prompt OR fall through to `oedit_disp_menu` (default case).

**`oedit_disp_val1_menu` (`:623-688`) → Value[0]:**

| item_type | prompt / action |
|---|---|
| ITEM_LIGHT | skip slot → call val3_menu (values 0 and 1 are unused for lights) |
| ITEM_SALVE / ITEM_PILL / ITEM_SCROLL / ITEM_WAND / ITEM_STAFF / ITEM_POTION | "Spell level : " |
| ITEM_MISSILE_WEAPON / ITEM_WEAPON | "Condition : " |
| ITEM_ARMOR | "Current AC : " |
| ITEM_PIPE / ITEM_CONTAINER / ITEM_DRINK_CON / ITEM_FOUNTAIN | "Capacity : " |
| ITEM_FOOD | "Hours to fill stomach : " |
| ITEM_MONEY | "Amount of Gold coins : " (GSC variant out of scope — stock C uses ITEM_MONEY) |
| ITEM_HERB | skip slot → call val2_menu |
| ITEM_LEVER / ITEM_SWITCH | `oedit_disp_lever_flags_menu` |
| ITEM_TRAP | "Charges: " |
| *default* | `oedit_disp_menu` (bail back to main) |

**`oedit_disp_val2_menu` (`:691-749`) → Value[1]:**

| item_type | prompt / action |
|---|---|
| ITEM_PILL / ITEM_SCROLL / ITEM_POTION | `oedit_disp_spells_menu` ("Enter the name of the spell:") |
| ITEM_SALVE / ITEM_HERB | "Charges: " |
| ITEM_PIPE | "Number of draws: " |
| ITEM_WAND / ITEM_STAFF | "Max number of charges : " |
| ITEM_WEAPON | "Number of damage dice : " |
| ITEM_FOOD | "Condition: " |
| ITEM_CONTAINER | `oedit_disp_container_flags_menu` |
| ITEM_DRINK_CON / ITEM_FOUNTAIN | "Quantity : " |
| ITEM_ARMOR | "Original AC: " |
| ITEM_LEVER / ITEM_SWITCH | if TRIG_CAST set: spells menu; else "Vnum: " |
| *default* | `oedit_disp_menu` |

**`oedit_disp_val3_menu` (`:752-784`) → Value[2]:**

| item_type | prompt / action |
|---|---|
| ITEM_LIGHT | "Number of hours (0 = burnt, -1 is infinite) : " |
| ITEM_PILL / ITEM_SCROLL / ITEM_POTION | spells menu |
| ITEM_WAND / ITEM_STAFF | "Number of charges remaining : " |
| ITEM_WEAPON | "Size of damage dice : " |
| ITEM_CONTAINER | "Vnum of key to open container (-1 for no key) : " |
| ITEM_DRINK_CON / ITEM_FOUNTAIN | `oedit_liquid_type` |
| *default* | `oedit_disp_menu` |

**`oedit_disp_val4_menu` (`:787-811`) → Value[3]:**

| item_type | prompt / action |
|---|---|
| ITEM_SCROLL / ITEM_POTION / ITEM_WAND / ITEM_STAFF | spells menu |
| ITEM_WEAPON | `oedit_disp_weapon_menu` |
| ITEM_DRINK_CON / ITEM_FOUNTAIN / ITEM_FOOD | "Poisoned (0 = not poisoned) : " |
| *default* | `oedit_disp_menu` |

**`oedit_disp_val5_menu` (`:814-833`) → Value[4]:**

| item_type | prompt / action |
|---|---|
| ITEM_SALVE | spells menu |
| ITEM_FOOD | "Food value: " |
| ITEM_MISSILE_WEAPON | "Range: " |
| *default* | `oedit_disp_menu` |

**`oedit_disp_val6_menu` (`:836-849`) → Value[5]:**

| item_type | prompt / action |
|---|---|
| ITEM_SALVE | spells menu |
| *default* | `oedit_disp_menu` |

### OEDIT_AFFECT_* flow

- **`OEDIT_AFFECT_MENU`** (`:1735-1777`): parse first character — `r`/`R` removes, `a`/`A` creates a new AFFECT_DATA stashed on `spare_ptr` and calls `oedit_disp_affect_menu`, `q`/`Q` nulls `spare_ptr` and falls through to main redisplay; digit → `edit_object_affect(d, atoi)` (lists by index, stash on spare_ptr).
- **`OEDIT_AFFECT_LOCATION`** (`:1779-1819`): input 0 → junk the pending affect (DISPOSE paf, null spare_ptr, re-show affect menu). Otherwise lookup via `get_atype(arg)` (word) or atoi (number); range-check `[0, MAX_APPLY_TYPE)` rejecting `APPLY_EXT_AFFECT`. Stash `paf->location`. Dispatch by location:
  - `APPLY_AFFECT` → `medit_disp_aff_flags` (bitmask editor)
  - `APPLY_RESISTANT` / `APPLY_IMMUNE` / `APPLY_SUSCEPTIBLE` → `medit_disp_ris`
  - `APPLY_WEAPONSPELL` / `APPLY_WEARSPELL` / `APPLY_REMOVESPELL` → spells menu
  - *else* → "Modifier: " prompt
- **`OEDIT_AFFECT_MODIFIER`** (`:1821-1914`): per `paf->location` consume input. For affect/RIS branches, consume space-separated bit tokens (numeric or word via `get_aflag`/`get_risflag`), accumulate into `tempnum`. For spells, accept number-or-name via `skill_lookup`. Otherwise plain `atoi`. On zero-valued finish, LINK a new AFFECT_DATA into `obj->first_affect` (or `pIndexData->first_affect` if ITEM_PROTOTYPE). Dispose the pending paf. Re-show the prompt-apply menu.
- **`OEDIT_AFFECT_REMOVE`** (`:1926-1931`): atoi → `remove_affect_from_obj` → re-show prompt-apply menu.
- **`OEDIT_AFFECT_RIS`** (`:1916-1924`): range-check 0..32 stub. The real bitflag accumulation happens in `OEDIT_AFFECT_MODIFIER` — OEDIT_AFFECT_RIS is a declared-but-unused mode in shipped C code. Reserve the iota slot; no code path reaches it. Document in §C Bug Catalog.

### ITEM_PROTOTYPE mirror-write pattern

Every mutation branch in `oedit_parse` mirrors changes to `obj->pIndexData->*` when `IS_OBJ_STAT(obj, ITEM_PROTOTYPE)`. In Go, `DoOedit` operates directly on `*ObjIndexData` (prototype only) and does not distinguish between instance and prototype — the existing flat `DoOset` path also only touches the prototype. The interactive menu must match this Go-port convention: menu mutations always target the prototype. Document as a C-divergence (Go simplification) in §Go Design.

### Other referenced C functions

- `cleanup_olc(d)` at `:91-104` — identical shape to redit's; disposes `d->olc`, sets `CON_PLAYING`, nulls `dest_buf`. Reuse existing Go `cleanupOlc` from `redit_parse.go:676-688`.
- `olc_log` — shared helper (the `ROOM(vnum)` prefix at `:710` is hard-coded; generalize to route by edit-target type).
- `get_otype` / `get_oflag` / `get_wflag` / `get_atype` / `get_aflag` / `get_risflag` — string→bit lookups via `build.c` tables. Verify during implementation that Go equivalents exist or need to be added.

## Go Current State

Verified 2026-04-19.

- `internal/types/enums.go:90` — `CON_OEDIT` defined (iota 22). No code reads it.
- `internal/game/loop.go:273-278` — dispatch switch has `case types.CON_REDIT:` only. **Adding `case types.CON_OEDIT: oeditParse(d, line)` is a 3-line change in §G7.**
- `internal/types/descriptor.go:77-79` — `Olc *OlcData` field exists (landed redit G1). Ready to reuse.
- `internal/types/olc.go:51+` — `OlcData` struct + 24 REDIT_* constants (iota+100). This plan ADDS a parallel OEDIT_* const block at iota+200 (see §G1).
- `internal/types/character.go` — has `Substate int`, `EditorSave func(*CharData)`, matches redit assumptions.
- `internal/types/enums.go:103-104` — `SUB_OBJ_LONG` (at :103) and `SUB_OBJ_EXTRA` (at :104) both already exist. **`SUB_OBJ_ACTION` is NOT in C** (verified: `src/mud.h:917` enum has no such entry) and is NOT in Go either. **G1 adds only `SUB_OBJ_ACTION`** as a Go-port divergence: C accepts inline single-line input for action desc (OEDIT_ACTDESC mode), but Go will use the text editor for multi-line fidelity. Document as C-divergence in CHANGELOG.
- `internal/act/olc_interactive.go:16-200` — `DoOedit` exists with flat path. No-arg branch at :22-26 prints usage. **Extend: when `argument == ""`, enter menu substate.**
- `internal/act/olc.go:30-35` — `ReditDispMenuFunc` seam declared. **G7 adds `OeditDispMenuFunc` seam of the same shape.**
- `internal/boot/boot.go:151-154` — wires `act.ReditDispMenuFunc = game.ReditDispMenu`. **G7 adds `act.OeditDispMenuFunc = game.OeditDispMenu`.**
- `internal/game/redit_parse.go:690-712` — `olcLog` helper has `ROOM(%d)` prefix hard-coded at :710. **G2 extracts to a general `olcLog` that accepts an edit-target label** (e.g. `"ROOM"` / `"OBJ"` / `"MOB"`). Backward-compatible: redit callers pass `"ROOM"`.
- `internal/game/redit_parse.go:714-723` — `worldRoomLookup` seam. oedit needs a **parallel `worldObjLookup`** seam in new `oedit_parse.go` targeting `WorldRef.ObjIndex[vnum]`.
- `internal/game/redit_parse.go:725-734` — `uRange` helper. Reuse (it's already package-level in `game`).
- `internal/game/editor.go:163-172` — `/s` handler transitions to CON_PLAYING BEFORE calling the EditorSave closure. **The oedit closures re-set CON_OEDIT after StopEditing**, same contract as redit.
- `internal/act/olc.go:799-830` — `DoSaveArea` area-save path. **Already uses `filepath.Join(WorldRef.DataDir, "area", filename)`** but has no containment guard. Post-Phase-6 audit (commit `b4c7477` 2026-04-19) flagged future-risk path traversal. **G8 adds a guard** and an explicit pin-test.

### Gaps this plan must fill

1. **OEDIT_* mode constants** in a new `const` block in `internal/types/olc.go` (iota+200, matching the redit convention).
2. **`SUB_OBJ_ACTION` constant** in `internal/types/enums.go` substate block (adjacent to existing `SUB_OBJ_EXTRA` at :104). `SUB_OBJ_LONG` pre-exists at :103. `SUB_OBJ_ACTION` is a Go-port divergence — C OEDIT_ACTDESC uses inline single-line input, Go uses the text editor for multi-line fidelity.
3. **`CON_OEDIT` arm in loop.go** dispatch switch, routing to new `oeditParse(d, line)`.
4. **`oeditParse` state-machine dispatcher** + menu-display helpers.
5. **Per-item-type value dispatch** for OEDIT_VALUE_1..OEDIT_VALUE_6 (six helper functions, each switching on `idx.ItemType`).
6. **Menu-entry extension of `DoOedit`**: `argument == ""` path now sets `Connected=CON_OEDIT`, allocates `Olc`, calls `OeditDispMenuFunc`. Preserves flat path for `oedit <vnum> <sub>` etc.
7. **`OeditDispMenuFunc` seam** declaration in `internal/act/olc.go` + wiring in `internal/boot/boot.go`.
8. **Path-containment guard** for area-save (§G8).
9. **Generalized `olcLog`** accepting a target-label parameter.
10. **`worldObjLookup` seam** in the new parser file.
11. **EditorSave restore-CON_OEDIT closures** for `SUB_OBJ_LONG` (long desc, uses pre-existing constant), `SUB_OBJ_ACTION` (action desc, new Go-port constant), `SUB_OBJ_EXTRA` (extradesc, pre-existing). Note: C `do_oedit_reset` only has SUB_OBJ_LONG + SUB_OBJ_EXTRA cases; the SUB_OBJ_ACTION closure is a Go-port addition.

## Go Design

### Chosen approach

**Approach A (CHOSEN): Mirror redit's structure.** New files `internal/game/oedit_menu.go` + `internal/game/oedit_parse.go` + `internal/game/oedit_value_menus.go` (split out because the 6 per-item-type dispatchers are ~200 LOC alone). Add OEDIT_* constants to `internal/types/olc.go`. Extend `DoOedit` no-arg branch. Add CON_OEDIT arm to loop.go.

Rationale:

- **Parallel file layout** — `oedit_menu.go`, `oedit_parse.go`, `oedit_value_menus.go` mirror `redit_menu.go`, `redit_parse.go`. A future maintainer cross-referencing redit ports will find the same shape.
- **Reuse `OlcData`** unchanged. `Target any` field already holds the edit-target pointer (type-asserted to `*ObjIndexData` in oedit); `Spare any` holds the pending affect pointer / extradesc pointer / mprog pointer (same triple-purpose as C `spare_ptr`). `TempNum int` holds the pending affect-flag bitmask (same as C `tempnum` at `:1807`).
- **Prototype-only mutation** — Go-port divergence from C. All menu mutations target `*ObjIndexData` directly. Flag as an intentional C-divergence in CHANGELOG; the flat `DoOedit` already behaves this way, so consistency.
- **Generalize `olcLog`** — add a `target string` parameter. Redit callers pass `"ROOM"`; oedit callers pass `"OBJ"`. Mpedit / medit inherit the same hook.
- **Path-containment guard** — G8 adds `filepath.Clean(path)` + `strings.HasPrefix(cleaned, filepath.Clean(areaDir))` check to `DoSaveArea` BEFORE the temp-file create. Pin-test exercises `../../etc/passwd` as the filename.

### Alternatives considered and rejected

- **Approach B: One monolithic `oedit_parse.go` with all menu renderers inline.** Rejected — the per-item-type value dispatchers alone warrant a separate file; keeps each file under 800 LOC.
- **Approach C: Port the ITEM_PROTOTYPE mirror-write logic verbatim.** Rejected — Go `DoOedit` + `DoOset` have NEVER implemented the instance-vs-prototype split. Adding it here in the menu but not in the flat path creates inconsistency. Defer the split to a separate `plan-phase6-item-prototype-split.md` if ever needed.
- **Approach D: Reuse flat `DoOset` verbatim from inside each menu arm.** Rejected for most arms — DoOset parses its own `vnum field value` argument format; the menu already has the vnum stashed on `Olc.Vnum` and the field+value split by the state machine. Pass-through would require re-serialization. Use shared helper functions below `DoOset` for field-mutation logic where they exist; otherwise port inline.
- **Approach E: Port OEDIT_MPROGS* now.** Rejected per plan charter — object-mprog UX is the `plan-phase6-olc-mpedit.md` surface.
- **Approach F: Emit ANSI screen-clear.** Rejected per redit precedent Q1. Cosmetic only; harms testclient determinism.

### OEDIT_* constant layout

Append to `internal/types/olc.go` below the REDIT_* block:

```go
// OEDIT_* modes for OlcData.Mode while d.Connected == CON_OEDIT.
// Mirror C src/olc.h OEDIT_* enum. Offset iota+200 to stay clear of
// CON_* (iota 0..~30) and REDIT_* (iota 100..124).
const (
    OEDIT_MAIN_MENU = iota + 200
    OEDIT_EDIT_NAMELIST
    OEDIT_SHORTDESC
    OEDIT_LONGDESC
    OEDIT_ACTDESC
    OEDIT_TYPE
    OEDIT_EXTRAS
    OEDIT_WEAR
    OEDIT_WEIGHT
    OEDIT_COST
    OEDIT_COSTPERDAY
    OEDIT_TIMER
    OEDIT_VALUE_1
    OEDIT_VALUE_2
    OEDIT_VALUE_3
    OEDIT_VALUE_4
    OEDIT_VALUE_5
    OEDIT_VALUE_6
    OEDIT_EXTRADESC_KEY
    OEDIT_CONFIRM_SAVEDB          // reserved; unused in stock C
    OEDIT_CONFIRM_SAVESTRING      // reserved; commented-out in C
    OEDIT_EXTRADESC_DESCRIPTION
    OEDIT_EXTRADESC_MENU
    OEDIT_LEVEL
    OEDIT_LAYERS
    OEDIT_AFFECT_MENU
    OEDIT_AFFECT_LOCATION
    OEDIT_AFFECT_MODIFIER
    OEDIT_AFFECT_REMOVE
    OEDIT_AFFECT_RIS              // declared in shipped C but unreachable code path
    OEDIT_EXTRADESC_CHOICE
    OEDIT_EXTRADESC_DELETE
    OEDIT_MPROGS                  // reserved for mpedit
    OEDIT_MPROGS_CHOICE           // reserved for mpedit
    OEDIT_MPROGS_DELETE           // reserved for mpedit
    OEDIT_MPROGS_TYPE             // reserved for mpedit
    OEDIT_MPROGS_ARG              // reserved for mpedit
)
```

Pins: `TestOEditData_ModeConstantsUnique` (pairwise distinct); `TestOEditData_ConstantsExceedRedit` (every OEDIT_* > every REDIT_*); `TestOEditData_ReservedMProgsIotaSlots` (asserts OEDIT_MPROGS = iota+200+32 so mpedit plan knows the starting offset).

## Task Groups

Every group follows test-first + mutation-verify (`Edit`-round-trip only — BANNED: `git checkout`, `git restore`, `git reset --hard`, `git stash`, `git commit --amend`).

### G1 — Schema: OEDIT_* constants + SUB_OBJ_ACTION (Go-port addition)

- Files: `internal/types/olc.go` (append OEDIT_* const block below the REDIT_* block), `internal/types/enums.go` (add ONE new substate constant `SUB_OBJ_ACTION` adjacent to `SUB_OBJ_EXTRA` at :104; `SUB_OBJ_LONG` already exists at :103 — do NOT add a duplicate).
- Test first:
  - `TestOEditData_ModeConstantsUnique` — pairwise-distinct assertion for all 37 OEDIT_* values.
  - `TestOEditData_ConstantsExceedRedit` — `OEDIT_MAIN_MENU > REDIT_EXTRADESC_DELETE` (or whatever the max REDIT_* is).
  - `TestOEditData_ReservedMProgsIotaSlots` — pins `OEDIT_MPROGS == OEDIT_MAIN_MENU + 32` (so mpedit plan can reference this guarantee).
  - `TestSubstate_ObjActionDistinct` — `SUB_OBJ_ACTION != SUB_OBJ_LONG != SUB_OBJ_EXTRA != SUB_NONE` (verifies the new constant is distinct from all existing substates).
  - `TestSubstate_ObjActionAppendsAfterExtra` — `SUB_OBJ_ACTION > SUB_OBJ_EXTRA` (pins iota ordering so new constant appends rather than shifts existing values).
- Mutation: drop `OEDIT_MAIN_MENU = iota + 200` offset (change to `iota + 100` — collision with REDIT_*) → `TestOEditData_ConstantsExceedRedit` fails.

### G2 — Generalize `olcLog` to accept target-label parameter

- File: `internal/game/redit_parse.go:690-712`. Change signature:
  ```go
  func olcLog(d *types.DescriptorData, target string, format string, args ...any)
  ```
  and update the format string from `"Log %s: ROOM(%d): %s"` to `"Log %s: %s(%d): %s"` with `target` interpolated.
- Sweep all existing call sites in `redit_parse.go` (~12 sites per earlier grep) to pass `"ROOM"` as the second arg.
- Test first: `TestOlcLog_EmitsRoomPrefix` (existing/new) confirms `ROOM(vnum)` format preserved. Add `TestOlcLog_EmitsObjPrefix` exercising `"OBJ"`.
- Mutation: swap `"%s(%d)"` → `"%d(%s)"` (argument order) → both tests fail.

### G3 — New parser skeleton + CON_OEDIT loop arm + `OeditDispMenuFunc` seam

- Files:
  - `internal/game/oedit_parse.go` (new) — package-level `oeditParse(d, line)` function (initially a stub that returns "unimplemented" + calls `cleanupOlc`), `worldObjLookup` closure seam, `SetWorldRef` already exported by redit's file (same package — reuse).
  - `internal/game/loop.go:273+` — add `case types.CON_OEDIT: oeditParse(d, line)` arm immediately below the existing CON_REDIT arm.
  - `internal/act/olc.go` — add seam declaration `var OeditDispMenuFunc func(d *types.DescriptorData)`.
  - `internal/boot/boot.go:153-154` — add wiring `act.OeditDispMenuFunc = game.OeditDispMenu`.
- Test first:
  - `TestLoop_ConOeditDispatchesToOeditParse` — spy seam on `oeditParse`, simulate descriptor with `Connected=CON_OEDIT`, push a line, assert stub called (not nanny).
  - `TestBoot_OeditDispMenuFuncWired` — post-boot, assert `act.OeditDispMenuFunc != nil`.
- Mutation: swap `case types.CON_OEDIT` → `case types.CON_REDIT` (collision with existing arm = compile error; use `case types.CON_MEDIT` instead — no arm yet, so descriptor drops to nanny "Unexpected state") → first test fails.

### G4 — Menu renderers (main + type + extras + wear + layers + extradesc + affect + liquid + weapon + spells + container-flags + lever-flags)

- File: `internal/game/oedit_menu.go` (new, ~600 LOC).
- Functions:
  - `OeditDispMenu(d)` — main menu. Mirror C `oedit_disp_menu` at `:921-985`. **17 rendered options (1-9, A-G, Q)** — `H` (mprog) is dispatched in the parser but NOT rendered. Displays vnum, name, s-desc, l-desc, a-desc, type-name, extras, wear, weight, cost, rent, timer, level, layers, 6 values, and menu-option lines for affect + extradesc. Set `OlcData.Mode = OEDIT_MAIN_MENU`.
  - `oeditDispTypeMenu(d)` — iterates 0..MAX_ITEM_TYPE using `itemTypeNames[]` table (port from `src/tables.c`-ish source; verify existing Go table or add `oTypes []string`). 3 cols. Set Mode = OEDIT_TYPE.
  - `oeditDispExtraMenu(d)` — MAX_ITEM_FLAG labels + current `ext_flag_string(obj->extra_flags, o_flags)`. 2 cols. Set Mode = OEDIT_EXTRAS.
  - `oeditDispWearMenu(d)` — ITEM_WEAR_MAX entries, skip ITEM_DUAL_WIELD bit. 2 cols. Set Mode = OEDIT_WEAR.
  - `oeditDispLayerMenu(d)` — 9 fixed options (Nothing / Silk Shirt / ... / Magical Effects). Port verbatim from `:449-466`. Set Mode = OEDIT_LAYERS.
  - `oeditDispExtradescMenu(d)` — list extradescs with numeric indices + A/R/Q. Set Mode = OEDIT_EXTRADESC_MENU.
  - `oeditDispExtraChoice(d)` — per-extradesc keyword + description sub-menu. Set Mode = OEDIT_EXTRADESC_CHOICE.
  - `oeditDispPromptApplyMenu(d)` — list affects + A/R/Q. Set Mode = OEDIT_AFFECT_MENU.
  - `oeditDispAffectMenu(d)` — MAX_APPLY_TYPE entries, 3 cols, skip 0 and APPLY_EXT_AFFECT. Set Mode = OEDIT_AFFECT_LOCATION.
  - `oeditLiquidType(d)` — LIQ_MAX entries from `liq_table[]`. 3 cols. Set Mode = OEDIT_VALUE_3.
  - `oeditDispWeaponMenu(d)` — 18 `attack_table` entries. 2 cols. Does NOT set Mode (caller is already in OEDIT_VALUE_4).
  - `oeditDispSpellsMenu(d)` — one-liner "Enter the name of the spell: " prompt.
  - `oeditDispContainerFlagsMenu(d)` — 5 `container_flags` labels.
  - `oeditDispLeverFlagsMenu(d)` — 29 `trig_flags` labels.
- Static tables to port (verify existence vs. add):
  - `oTypes []string` — item-type labels (`o_types[]` from `src/tables.c`). Check `internal/types/` for existing table.
  - `oFlags []string` — item extra-flag labels (`o_flags[]`). Likely already exists for `ext_flag_string` generator.
  - `wFlags []string` — wear-flag labels (`w_flags[]`).
  - `aTypes []string` — apply-type labels (`a_types[]`).
  - `aFlags []string` — affect-bit labels.
  - `risFlags []string` — RIS-bit labels.
  - `containerFlags []string` — 5 labels (`closeable`, `pickproof`, `closed`, `locked`, `eatkey`).
  - `trigFlags []string` — 29 labels.
  - `liqTable []struct{ LiqName string; ... }` — LIQ_MAX entries.
  - `attackTable []string` — 18 weapon-type labels.
- Test first: per-menu "contains expected strings" tests (11 tests in total, one per menu). Example: `TestOeditDispTypeMenu_ShowsAllItemTypes`.
- Mutation: drop `"Layers"` line from `OeditDispMenu` format → `TestOeditDispMenu_ContainsAllFields` fails.
- Color-tag handling: pass through `&g`/`&w`/`&O` tags; descriptor `ColorFunc` strips/translates at flush, matching redit precedent.
- No ANSI screen-clear (per Q1).

### G5 — Top-level `oeditParse` dispatcher + OEDIT_MAIN_MENU branch

- File: `internal/game/oedit_parse.go` (extend the G3 stub).
- Outer skeleton `switch d.Olc.Mode { ... }` with ONLY the OEDIT_MAIN_MENU case fully populated and all other cases returning "not yet implemented" stubs (filled in §G6-G12 incrementally).
- OEDIT_MAIN_MENU dispatches on `strings.ToUpper(first char)`:
  - `Q` → `cleanupOlc(d)` (reuse existing helper from `redit_parse.go:676-688` — it is already mode-agnostic).
  - `1` → prompt "Enter namelist : ", Mode = OEDIT_EDIT_NAMELIST.
  - `2` → prompt "Enter short desc : ", Mode = OEDIT_SHORTDESC.
  - `3` → prompt "Enter long desc :-\r\n| ", `ch.Substate = SUB_OBJ_LONG`, build EditorSave closure that writes `idx.Description`, calls `StopEditingFunc`, re-sets `Connected = CON_OEDIT` (**matching redit G5 trampoline**), calls `OeditDispMenu(ch.Desc)`. Call `StartEditingFunc(ch, idx.Description)`.
  - `4` → prompt "Enter action desc :-\r\n", `Substate = SUB_OBJ_ACTION`, EditorSave writes `idx.ActionDesc` (verify Go field name; C uses `action_desc`). Mode = OEDIT_ACTDESC (but actually enters CON_EDITING via StartEditingFunc — set mode before the call so the trampoline reliably returns).
  - `5` → `oeditDispTypeMenu`, Mode = OEDIT_TYPE.
  - `6` → `oeditDispExtraMenu`, Mode = OEDIT_EXTRAS.
  - `7` → `oeditDispWearMenu`, Mode = OEDIT_WEAR.
  - `8` → prompt "Enter weight : ", Mode = OEDIT_WEIGHT.
  - `9` → prompt "Enter cost : ", Mode = OEDIT_COST.
  - `A` → prompt "Enter cost per day : ", Mode = OEDIT_COSTPERDAY.
  - `B` → prompt "Enter timer : ", Mode = OEDIT_TIMER.
  - `C` → prompt "Enter level : ", Mode = OEDIT_LEVEL.
  - `D` → layerable-check (any of `ITEM_WEAR_BODY`/`ABOUT`/`ARMS`/`FEET`/`HANDS`/`LEGS`/`WAIST` set) — if yes call `oeditDispLayerMenu`, Mode = OEDIT_LAYERS; if no print "The wear location of this object is not layerable.\n\r" and redisplay main (don't change mode).
  - `E` → `oeditDispVal1Menu`. (Mode is set inside val1_menu.)
  - `F` → `oeditDispPromptApplyMenu`, Mode = OEDIT_AFFECT_MENU.
  - `G` → `oeditDispExtradescMenu`, Mode = OEDIT_EXTRADESC_MENU.
  - `H` → **NOT WIRED** (mpedit scope-cut). Print "Mudprog editing is not yet supported via the menu; use the flat 'mpedit' commands." and redisplay main.
  - default → redisplay main menu.
- Test first: 6+ unit tests covering `Q`, `1`, `5` (type menu entry), `D` layerable-guard (positive + negative), `H` scope-cut message.
- Mutation: swap `CON_OEDIT` → `CON_PLAYING` in the `3`-branch EditorSave closure → post-save test asserts `Connected == CON_OEDIT` fails.

### G6 — Field-set branches: OEDIT_EDIT_NAMELIST / SHORTDESC / LONGDESC (editor) / ACTDESC (editor) / WEIGHT / COST / COSTPERDAY / TIMER / LEVEL / TYPE / EXTRAS / WEAR / LAYERS

- File: `internal/game/oedit_parse.go` (extend).
- Each arm mutates `idx.*` (target is `*ObjIndexData` per §Go Design), calls `olcLog(d, "OBJ", format, args...)`, redisplays main menu (Mode = OEDIT_MAIN_MENU), and returns.
- Clamping:
  - OEDIT_LEVEL: `uRange(0, n, MAX_LEVEL)` matching C `:1487`.
  - OEDIT_WEIGHT / COST / COSTPERDAY / TIMER: no range clamp (matches C `:1450-1489`).
- OEDIT_TYPE: accept number or word (via new helper `getOtype(arg)` — table lookup on `oTypes`). Reject `number < 1 || number >= MAX_ITEM_TYPE` with "Invalid choice, try again : " and return WITHOUT changing mode.
- OEDIT_EXTRAS: loop on space-separated tokens. Per-token: if numeric `0`, redisplay main (C `:1367-1370`); if numeric `n`, `n-=1`, range-check `[0, MAX_ITEM_FLAG]`, toggle bit `n` via `ext_flag_toggle` (`xTOGGLE_BIT`); else word → `getOflag(token)` and toggle. Protection: `ITEM_PROTOTYPE` flag toggle requires `GetTrust() >= LEVEL_GREATER OR is_name("protoflag", ch.PCData.Bestowments)` (per C `:1389-1392`). If numeric input, process only ONE flag then break (C `:1401-1403`); if word input, process all. End with `oeditDispExtraMenu` redisplay (stay in OEDIT_EXTRAS).
- OEDIT_WEAR: numeric branch: `0` → fall through to save; else `n-=1`, range-check `[0, ITEM_WEAR_MAX+1)`, toggle `1<<n`. Word branch: loop tokens via `getWflag`, toggle each. End with `oeditDispWearMenu` redisplay.
- OEDIT_LAYERS: number 0 → redisplay main; 1 → `idx.Layers = 0`; 2-9 → `TOGGLE_BIT(idx.Layers, 1<<(n-2))` (bit values: 1, 2, 4, 8, 16, 32, 64, 128 per `:1500-1526`). Redisplay layer menu. Invalid → "Invalid selection, try again: ".
- Test first:
  - `TestOeditParse_Namelist_SetsName`
  - `TestOeditParse_Weight_SetsWeight`
  - `TestOeditParse_Level_ClampsToMaxLevel` (input `999` → `MAX_LEVEL`)
  - `TestOeditParse_Type_RejectsOutOfRange`
  - `TestOeditParse_Type_WordLookup` (input `"weapon"` → `idx.ItemType == ITEM_WEAPON`)
  - `TestOeditParse_Extras_WordToggle` (input `"magic"` → toggle)
  - `TestOeditParse_Extras_NumberBreaks` (input `"5 6"` → processes only 5, not 6)
  - `TestOeditParse_Extras_PrototypeGatedByTrust` (trust < LEVEL_GREATER → "cannot change the prototype flag")
  - `TestOeditParse_Wear_NumberZeroSaves`
  - `TestOeditParse_Layers_Option1ZeroesLayers`
  - `TestOeditParse_Layers_Option2TogglesBit1`
- Mutations:
  - Flip OEDIT_LEVEL clamp `uRange(0, n, MAX_LEVEL)` → `uRange(0, n, 100)` → `TestOeditParse_Level_ClampsToMaxLevel` fails (assuming MAX_LEVEL > 100).
  - Swap OEDIT_TYPE lower `number < 1` → `number < 0` → `TestOeditParse_Type_RejectsOutOfRange` fails on input `0`.
  - Drop OEDIT_EXTRAS prototype-trust guard → `TestOeditParse_Extras_PrototypeGatedByTrust` fails.
  - Swap OEDIT_LAYERS option-1 `idx.Layers = 0` → `idx.Layers = 1` → `TestOeditParse_Layers_Option1ZeroesLayers` fails.

### G7 — Per-item-type value dispatchers (OEDIT_VALUE_1 through OEDIT_VALUE_6)

- File: `internal/game/oedit_value_menus.go` (new, ~300 LOC).
- Six functions: `oeditDispVal1Menu(d)` ... `oeditDispVal6Menu(d)`. Each sets `Olc.Mode = OEDIT_VALUE_N` and dispatches on `idx.ItemType` per the tables in §C Reference above.
- Six parser arms in `oedit_parse.go`: OEDIT_VALUE_1 ... OEDIT_VALUE_6. Each:
  - Parses input (numeric or word depending on item-type).
  - Applies URANGE clamping per the C table at `:1535-1732`.
  - For skill-lookup branches (ITEM_SCROLL / POTION / PILL / SALVE / WAND / STAFF Value[3], Value[4]), lookup via `skillLookup(arg)` helper (verify existing Go equivalent; add if missing).
  - For `ITEM_LEVER`/`SWITCH` OEDIT_VALUE_1 with number 0..29 → call `oeditDispLeverFlagsMenu`, toggle `idx.Value[0]` bit `1<<(n-1)`; number 0 → advance to val2_menu.
  - For `ITEM_CONTAINER` OEDIT_VALUE_2 with number 0..MAX_OLC_ITEMS_LIST → toggle `idx.Value[1]` bit `1<<(n-1)`; number 0 → advance to val3_menu.
  - On terminal branch (simple atoi) → set value, call olcLog, advance to next val_menu. On last (val6_menu) → redisplay main menu.
- Test first (one per item-type × value-slot = lots; prioritize the following pins):
  - `TestOeditDispVal1_WeaponPromptsCondition`
  - `TestOeditDispVal1_LightSkipsToVal3`
  - `TestOeditDispVal1_HerbSkipsToVal2`
  - `TestOeditDispVal1_LeverCallsLeverFlagsMenu`
  - `TestOeditParse_Val1_WeaponSetsValue0`
  - `TestOeditParse_Val1_LeverToggleTrigFlag`
  - `TestOeditParse_Val2_ContainerTogglesFlag`
  - `TestOeditParse_Val3_LightAllowsNegativeOne` (min=-32000, max=32000)
  - `TestOeditParse_Val3_ScrollSkillLookup` ("fireball" → skill sn)
  - `TestOeditParse_Val4_WeaponRejectsOutOfRangeAttack` (n >= MAX_ATTACK_TYPE → re-prompt, no mutation)
  - `TestOeditParse_Val5_FoodMinClampedToZero`
  - `TestOeditParse_Val6_SalveSkillLookup`
- Mutations:
  - Swap ITEM_LIGHT branch in val1_menu from `oeditDispVal3Menu` → `oeditDispVal2Menu` → `TestOeditDispVal1_LightSkipsToVal3` fails.
  - Flip OEDIT_VALUE_4 WEAPON `max_val = MAX_ATTACK_TYPE - 1` → `MAX_ATTACK_TYPE` → range-reject test fails (boundary off by one).
  - Drop ITEM_CONTAINER branch in val2_parse → container toggle-test fails.

### G8 — Extradesc sub-menu: OEDIT_EXTRADESC_MENU / _CHOICE / _KEY / _DESCRIPTION / _DELETE

- File: `internal/game/oedit_parse.go` (extend).
- Mirror redit G9 semantics. Key differences from redit:
  - C iterates BOTH `pIndexData->first_extradesc` AND `obj->first_extradesc` at `:476-491`. Go-port: operate on `idx.FirstExtradesc` only (prototype-only divergence, per §Go Design). Document in CHANGELOG.
  - `OEDIT_EXTRADESC_DESCRIPTION` path: `Substate = SUB_OBJ_EXTRA`, EditorSave closure writes `ed.Description`, calls `StopEditingFunc`, re-sets `Connected = CON_OEDIT`, calls `OeditDispExtraChoice`. Match redit's trampoline pattern exactly.
  - Add/Delete: allocate new `ExtraDescData{Keyword: "", Description: ""}` and LINK to `idx.FirstExtradesc` (Go slice-push). Delete: unlink via `oeditFindExtradesc(idx, n)` helper.
  - Junk-empty-on-Q: `OEDIT_EXTRADESC_CHOICE` with Q → if both keyword+description are empty, remove the most recently added extradesc (match C `:1973` fall-through to `OEDIT_EXTRADESC_MENU` + junk logic present upstream at `:941-949` of ... wait, that's redit's guard. Verify C ooedit at `:1950-1973` — actually C oedit does NOT have the junk-on-Q guard; it simply transitions back to OEDIT_EXTRADESC_MENU. Match C: no junk-on-Q.) **Divergence from redit**: oedit does not junk empty on Q. Document.
- Test first:
  - `TestOeditParse_ExtradescAdd_ReturnsChoice`
  - `TestOeditParse_ExtradescDelete_ByIndex`
  - `TestOeditParse_ExtradescKey_StoresKeyword`
  - `TestOeditParse_ExtradescDescription_CallsStartEditing`
  - `TestOeditParse_ExtradescChoice_QReturnsMenuNoJunk` (pins the oedit-vs-redit divergence)
- Mutation: swap `Connected = CON_OEDIT` → `CON_REDIT` in the description-editor EditorSave closure → `TestOeditParse_ExtradescDescription_RestoresConOedit` fails.

### G9 — Affect sub-menu: OEDIT_AFFECT_MENU / _LOCATION / _MODIFIER / _REMOVE (+ scope-cut OEDIT_AFFECT_RIS reservation)

- File: `internal/game/oedit_parse.go` (extend).
- Mirror C `:1735-1931`.
- `OEDIT_AFFECT_MENU`: `a`/`A` → allocate `AffectData`, stash on `Olc.Spare`, call `oeditDispAffectMenu`, Mode = OEDIT_AFFECT_LOCATION; `r`/`R` → consume next arg; if numeric provided, remove-then-redisplay; else prompt "Remove which affect? ", Mode = OEDIT_AFFECT_REMOVE; `q`/`Q` → null `Spare`, fall through to main redisplay; numeric → `editObjectAffect(d, n)` (lookup by index, stash on Spare).
- `OEDIT_AFFECT_LOCATION`: numeric 0 → dispose pending Spare, redisplay prompt-apply-menu. Otherwise `getAtype(arg)` (word) or atoi (number); range-check `[0, MAX_APPLY_TYPE)` rejecting `APPLY_EXT_AFFECT`. Stash `paf.Location = n`. Dispatch:
  - `APPLY_AFFECT` → set `Olc.TempNum = 0`, call `meditDispAffFlags(d)` (external seam: **medit plan owns this renderer**; until medit lands, emit a placeholder "Affect-flag editing is a medit-plan dependency; enter 0 to cancel.". This is a documented soft-dependency — menu is functional but not fully featured until medit lands.) Mode = OEDIT_AFFECT_MODIFIER.
  - `APPLY_RESISTANT` / `APPLY_IMMUNE` / `APPLY_SUSCEPTIBLE` → set TempNum=0, call `meditDispRis(d)` (same soft-dep placeholder). Mode = OEDIT_AFFECT_MODIFIER.
  - `APPLY_WEAPONSPELL` / `APPLY_WEARSPELL` / `APPLY_REMOVESPELL` → call `oeditDispSpellsMenu(d)`. Mode = OEDIT_AFFECT_MODIFIER.
  - else → prompt "\n\rModifier: ". Mode = OEDIT_AFFECT_MODIFIER.
- `OEDIT_AFFECT_MODIFIER`: per `paf.Location`:
  - For APPLY_AFFECT/RIS branches: numeric 0 → break with `value = Olc.TempNum`; numeric n → `TOGGLE_BIT(Olc.TempNum, 1<<(n-1))`, re-display the flag menu (medit-dep placeholder), return.
  - For spells: `IS_VALID_SN(n)` or `bsearch_skill_exact` (verify Go equivalent); if invalid, reprompt.
  - else: `value = atoi(arg)`.
  - After break: if `value == 0 || Olc.Change` → set `paf.Modifier = value`, log "Modified affect", redisplay prompt-apply menu, return. Otherwise allocate a new `AffectData` (`Type=-1, Duration=-1, Location = uRange(0, paf.Location, MAX_APPLY_TYPE), Modifier = value`), LINK into `idx.FirstAffect`, dispose pending paf, null Spare, redisplay prompt-apply menu.
- `OEDIT_AFFECT_REMOVE`: atoi → `removeAffectFromObj(idx, n)` → log "Removed affect #%d" → redisplay prompt-apply menu.
- `OEDIT_AFFECT_RIS` (reserved): unreachable code path in shipped C (see §C Bug Catalog #2); Go reserves the iota slot and has the case in the switch returning `bug("Oedit_parse: OEDIT_AFFECT_RIS unreachable path"); oeditDispMenu(d); return`.
- Test first:
  - `TestOeditParse_Affect_QnullsSpare`
  - `TestOeditParse_AffectLocation_AtypeWordLookup`
  - `TestOeditParse_AffectLocation_RejectsExtAffect`
  - `TestOeditParse_AffectLocation_ZeroJunksPending`
  - `TestOeditParse_AffectModifier_SimpleAtoi`
  - `TestOeditParse_AffectModifier_AffectFlagPlaceholder` (asserts the medit-soft-dep placeholder message appears)
  - `TestOeditParse_AffectRemove_CallsHelper`
  - `TestOeditParse_AffectRis_UnreachableBugs`
- Mutation: drop the APPLY_EXT_AFFECT reject in OEDIT_AFFECT_LOCATION → `TestOeditParse_AffectLocation_RejectsExtAffect` fails.

### G10 — Menu-entry extension of `DoOedit`

- File: `internal/act/olc_interactive.go:22-26`. Replace:
  ```go
  if vnumArg == "" {
      ch.Send("Usage: oedit <vnum> <subcommand> [args]\n\r")
      ...
      return
  }
  ```
  with:
  ```go
  if vnumArg == "" {
      // Match C do_ooedit at :129-134: print the C-verbatim error when
      // the editor is invoked without a target. Creation via this
      // command requires ENABLE_OLC2_EXTRAS which is OFF in stock.
      ch.Send("OEdit what?\n\r")
      return
  }
  ```
  **Wait — this is WRONG.** The menu-entry path is invoked with `oedit <vnum>` (vnum but no subcommand). Current code at :32: `sub, args := util.OneArgument(rest)`. Change the behavior at the `sub == ""` branch (around :71 in the switch):
  ```go
  case "":
      // Enter interactive menu substate.
      if ch.Desc == nil {
          ch.Send("You have no descriptor.\n\r")
          return
      }
      ch.Desc.Olc = &types.OlcData{
          Mode:   types.OEDIT_MAIN_MENU,
          Vnum:   vnum,
          Target: idx,
      }
      ch.Desc.Connected = types.CON_OEDIT
      if OeditDispMenuFunc != nil {
          OeditDispMenuFunc(ch.Desc)
      }
      return
  ```
  (The old :71 `oeditShow(ch, idx)` is replaced by the menu-entry path. `oeditShow` becomes a flat alternative available as `oedit <vnum> show` if desired — but verify existing callers; if none rely on no-arg show, remove `oeditShow` or rebind.)
- **Note**: C `do_ooedit` requires an OBJECT (name/vnum of a live instance) not a vnum. Go-port has diverged to vnum-only (matches flat `DoOedit`). Document as a **deliberate, existing Go-port divergence** — not introduced here, just noted.
- Test first:
  - `TestDoOedit_VnumOnlyEntersMenu` — `DoOedit(ch, "4000")` sets `Connected == CON_OEDIT`, `Olc.Mode == OEDIT_MAIN_MENU`, `Olc.Vnum == 4000`, `Olc.Target == WorldRef.ObjIndex[4000]`.
  - `TestDoOedit_VnumWithSubcommandKeepsFlatPath` — `DoOedit(ch, "4000 name Foo")` does NOT set Connected to CON_OEDIT.
  - `TestDoOedit_EmptyArgPrintsUsage` — no change to existing behavior (prints usage).
  - `TestDoOedit_VnumUnknownRejects` — `DoOedit(ch, "99999")` prints "does not exist" and does NOT enter menu.
- Mutation: remove the `ch.Desc.Connected = types.CON_OEDIT` assignment → `TestDoOedit_VnumOnlyEntersMenu` fails.

### G11 — Area-save path-containment guard + pin test

- File: `internal/act/olc.go:799-830` (`DoSaveArea`). Add after line `path := filepath.Join(WorldRef.DataDir, "area", filename)`:
  ```go
  areaDir := filepath.Clean(filepath.Join(WorldRef.DataDir, "area"))
  cleaned := filepath.Clean(path)
  if !strings.HasPrefix(cleaned, areaDir+string(filepath.Separator)) && cleaned != areaDir {
      ch.Sendf("Invalid area filename (rejected path): %s\n\r", filename)
      return
  }
  ```
- Same guard logic must be added to any OTHER area-save site touched by this plan (grep to confirm there's only one).
- Test first:
  - `TestDoSaveArea_RejectsPathTraversal` — set `area.Filename = "../../etc/passwd"`, call `DoSaveArea`, assert "rejected path" message AND NO file created at `/etc/passwd` AND NO file created at `WorldRef.DataDir/area/../../etc/passwd`.
  - `TestDoSaveArea_RejectsAbsolutePath` — `area.Filename = "/tmp/pwned.are"`, assert rejection.
  - `TestDoSaveArea_RejectsDotDotSegment` — `area.Filename = "foo/../bar.are"`, verify rejected even though the net effect is a sibling directory (conservative behavior).
  - `TestDoSaveArea_AcceptsNormalFilename` — `area.Filename = "newacad.are"`, assert save proceeds (regression guard).
- Mutation: drop the `!strings.HasPrefix(...)` guard → all three rejection tests fail; accept-test still passes. Pins the exact positive-security behavior.

### G12 — Boot wire + testclient E2E

- Files:
  - `internal/boot/boot.go:153-154` — add `act.OeditDispMenuFunc = game.OeditDispMenu` immediately after the redit wiring.
  - `internal/testclient/oedit_test.go` (new) — three E2E scenarios mirroring redit G12.
- Scenarios:
  - `TestTestclient_OeditMenuEntryAndQuit` — immortal logs in, types `oedit 4000`, reads menu (match on "Enter choice :"), types `Q`, sees "Exiting editor." (or whatever cleanupOlc emits), `Connected` returns to CON_PLAYING.
  - `TestTestclient_OeditMenuSetName` — `oedit 4000`, `1`, "Foo Bar", verify menu redisplay contains `"Name     : Foo Bar"`.
  - `TestTestclient_OeditInvalidChoiceRedisplays` — `oedit 4000`, bad input like `"zzz"`, verify menu redisplays without CON_OEDIT drop.
  - `TestTestclient_OeditTypeSubmenu` — `oedit 4000`, `5`, `1` (ITEM_LIGHT or whatever ITEM_TYPE=1 is), verify `idx.ItemType` updated.
- Mutation: swap `case CON_OEDIT` → `case CON_MEDIT` in loop.go → first test fails with "Unexpected state. Disconnecting.".

### G13 — CLAUDE.md + CHANGELOG + TODO housekeeping + completion record

- Update `CLAUDE.md` "Phase 6 plans authored" table row: flip status to **LANDED YYYY-MM-DD**.
- Append CHANGELOG entry.
- Move any related TODO.md item from Active to Done.
- Append Completion Record section to this plan file.

---

## Acceptance Criteria

A1. `internal/types/olc.go` has a new `const` block with all 37 OEDIT_* values, offset `iota + 200`. Pairwise-distinct and all values > max REDIT_*.
A2. `internal/types/enums.go` has `SUB_OBJ_ACTION` constant, distinct from `SUB_OBJ_LONG` / `SUB_OBJ_EXTRA` / `SUB_NONE`, AND appended AFTER `SUB_OBJ_EXTRA` in the iota block (i.e. `SUB_OBJ_ACTION > SUB_OBJ_EXTRA`) so existing SUB_* numeric values don't shift. (`SUB_OBJ_LONG` already exists at :103 — no new entry needed for it.) Pinned by `TestSubstate_ObjActionDistinct` (value inequality) + `TestSubstate_ObjActionAppendsAfterExtra` (iota ordering).
A3. `internal/game/loop.go` dispatch switch has a `case types.CON_OEDIT: oeditParse(d, line)` arm below the existing CON_REDIT arm.
A4. `DoOedit(ch, "<vnum>")` (vnum only, no subcommand) with `ch.GetTrust() >= LEVEL_IMMORTAL` and a valid vnum transitions `ch.Desc.Connected` to `CON_OEDIT`, allocates `ch.Desc.Olc` with `Target = WorldRef.ObjIndex[vnum]`, `Mode = OEDIT_MAIN_MENU`, and emits the main menu.
A5. `DoOedit(ch, "<vnum> name Foo")` (flat subcommand) does NOT change `Connected`; the flat path regression tests still pass.
A6. `DoOedit(ch, "")` still prints the usage line (does NOT enter menu — matches C `"OEdit what?"`).
A7. In the menu, input `1 <CR>` → prompts for namelist; follow-up line sets `idx.Name` and redisplays main menu.
A8. In the menu, input `3 <CR>` enters the long-desc text editor via `StartEditingFunc`; `/s` writes `idx.Description` AND restores `Connected == CON_OEDIT` (NOT CON_PLAYING); main menu redisplays.
A9. In the menu, input `4 <CR>` enters the action-desc text editor with the same restore-CON_OEDIT contract.
A10. In the menu, input `5 <CR>` shows the item-type menu; numeric in `[1, MAX_ITEM_TYPE)` OR word (via `getOtype`) sets `idx.ItemType`. Out-of-range rejected without mode change.
A11. In the menu, input `6 <CR>` shows the extra-flags menu; numeric `1..MAX_ITEM_FLAG+1` toggles that bit on `idx.ExtraFlags`. Word-list input (space-separated) also works via `getOflag`. Numeric input processes one flag then returns; word input processes all.
A12. In the menu, `ITEM_PROTOTYPE` flag toggle rejects when `GetTrust() < LEVEL_GREATER` AND `not is_name("protoflag", bestowments)` with "You cannot change the prototype flag.".
A13. In the menu, input `7 <CR>` shows the wear-flags menu with ITEM_DUAL_WIELD bit SKIPPED; numeric-or-word input toggles `idx.WearFlags`.
A14. In the menu, `D` (layers) is gated: only offered when `idx.WearFlags` has any of `BODY|ABOUT|ARMS|FEET|HANDS|LEGS|WAIST`; otherwise prints "not layerable" without mode change.
A15. In the menu, `E` (values) dispatches `oeditDispVal1Menu` which branches on `idx.ItemType`: ITEM_LIGHT skips to val3, ITEM_HERB skips to val2, ITEM_LEVER/SWITCH → lever-flags menu, standard items → typed prompt per §C Reference tables. Every ITEM_TYPE_* has a known path.
A16. In the menu, value-slot parsers apply the URANGE clamps from §C Reference: OEDIT_VALUE_3 weapon `[0, 100]`, OEDIT_VALUE_3 drink/fountain `[0, LIQ_MAX]`, OEDIT_VALUE_4 weapon `[0, MAX_ATTACK_TYPE-1]` (reject rather than clamp), OEDIT_VALUE_5 food `[0, 32000]`. OEDIT_VALUE_3/4/5/6 with skill-lookup branches accept both sn-number and spell-name input.
A17. In the menu, `F` (affects) shows the prompt-apply menu; `a`/`A` creates a pending `AffectData`, `r`/`R` removes by index, numeric edits by index, `q`/`Q` returns to main.
A18. In the menu, APPLY_EXT_AFFECT is rejected as a location choice (cannot be set via OLC).
A19. In the menu, `G` (extradescs) shows the list; `A` adds + enters choice sub-menu; `R` deletes by index; numeric edits by index. Description path uses the editor with restore-CON_OEDIT contract.
A20. In the menu, `H` (mudprogs) prints a scope-cut message and redisplays main (does NOT enter OEDIT_MPROGS — reserved for mpedit plan).
A21. In the menu, `Q` exits cleanly: `Connected == CON_PLAYING`, `Olc == nil`, `Substate == SUB_NONE`. Same contract as redit cleanupOlc.
A22. Invalid input at any menu redisplays the current menu without dropping CON_OEDIT.
A23. Flat `oedit <vnum> <sub> [args]` regression tests (existing `olc_interactive_test.go` + `olc_set_test.go`) still pass unchanged.
A24. `DoSaveArea` rejects path-traversal filenames (absolute paths, `..` segments, paths escaping `DataDir/area/`). Pin-tests cover 3 attack vectors + 1 accept case.
A25. `olcLog` accepts a target-label parameter; redit callers pass `"ROOM"`, oedit callers pass `"OBJ"`; both emit `Log <name>: <target>(<vnum>): <msg>`.
A26. `go build ./...` green, `go vet ./...` clean, `gofmt -l` clean on all changed/new files, `go test -count=3 ./...` green across all 15 packages.
A27. Mutation matrix ≥ 12 mutations, all caught by at least one test, all reverted via `Edit`-only (no `git checkout` / `git restore` / `git stash` / `git reset --hard`).

---

## Mutation Gates (≥ 12 — Edit-round-trip only)

1. Drop the `OEDIT_MAIN_MENU = iota + 200` offset (use `iota + 100`) → `TestOEditData_ConstantsExceedRedit` fails.
2. Swap `case types.CON_OEDIT` → `case types.CON_MEDIT` in loop.go dispatch → `TestLoop_ConOeditDispatchesToOeditParse` + `TestTestclient_OeditMenuEntryAndQuit` both fail.
3. Remove the `ch.Desc.Connected = types.CON_OEDIT` assignment in `DoOedit` vnum-only branch → `TestDoOedit_VnumOnlyEntersMenu` fails.
4. Flip OEDIT_LEVEL clamp upper `MAX_LEVEL` → `100` → `TestOeditParse_Level_ClampsToMaxLevel` fails (assuming MAX_LEVEL > 100; verify).
5. Swap OEDIT_TYPE lower bound `number < 1` → `number < 0` → `TestOeditParse_Type_RejectsOutOfRange` fails on input `0`.
6. Drop the ITEM_PROTOTYPE trust guard in OEDIT_EXTRAS → `TestOeditParse_Extras_PrototypeGatedByTrust` fails.
7. Swap OEDIT_LAYERS option-1 `idx.Layers = 0` → `idx.Layers = 1` → `TestOeditParse_Layers_Option1ZeroesLayers` fails.
8. Swap ITEM_LIGHT branch in val1_menu from `oeditDispVal3Menu` → `oeditDispVal2Menu` → `TestOeditDispVal1_LightSkipsToVal3` fails.
9. Change OEDIT_VALUE_4 WEAPON `max_val = MAX_ATTACK_TYPE - 1` → `MAX_ATTACK_TYPE` → `TestOeditParse_Val4_WeaponRejectsOutOfRangeAttack` fails at the boundary.
10. Drop the APPLY_EXT_AFFECT reject in OEDIT_AFFECT_LOCATION → `TestOeditParse_AffectLocation_RejectsExtAffect` fails.
11. Swap `Connected = CON_OEDIT` → `CON_REDIT` in the long-desc EditorSave closure → `TestOeditParse_Longdesc_RestoresConOedit` fails.
12. Swap `Connected = CON_OEDIT` → `CON_PLAYING` in the extradesc-description EditorSave closure → `TestOeditParse_ExtradescDescription_RestoresConOedit` fails.
13. Drop the path-containment guard in `DoSaveArea` → `TestDoSaveArea_RejectsPathTraversal` + `TestDoSaveArea_RejectsAbsolutePath` + `TestDoSaveArea_RejectsDotDotSegment` all fail.
14. Flip `olcLog` format `"%s(%d)"` → `"%d(%s)"` → `TestOlcLog_EmitsObjPrefix` + `TestOlcLog_EmitsRoomPrefix` both fail.
15. Drop the `Substate = SUB_NONE` clearing in `cleanupOlc` → `TestOeditParse_Quit_ClearsSubstate` fails. Mutation is exercised via an oedit-specific test (not deferred to redit coverage) because oedit substates include `SUB_OBJ_LONG` / `SUB_OBJ_ACTION` that redit's cleanupOlc path never touches — the oedit test seeds `ch.Substate = SUB_OBJ_LONG` before `Q` to pin that the shared helper clears non-redit substates too.

---

## Scope Cuts / Deferrals

- **OEDIT_MPROGS / OEDIT_MPROGS_CHOICE / OEDIT_MPROGS_DELETE / OEDIT_MPROGS_TYPE / OEDIT_MPROGS_ARG** — entire 5-mode mudprog sub-machine at C `:2037-2150`. Deferred to `plan-phase6-olc-mpedit.md`. Iota slots reserved; main-menu `H` prints a scope-cut placeholder.
- **ITEM_PROTOTYPE instance-vs-prototype mirror-writes** — C `oedit_parse` mirrors every mutation to both `obj->*` and `obj->pIndexData->*` when `IS_OBJ_STAT(obj, ITEM_PROTOTYPE)`. Go `DoOedit`/`DoOset` have always mutated prototype-only. Menu matches flat-path convention. Deferred to `plan-phase6-item-prototype-split.md` (not yet authored; lowest priority).
- **OEDIT_CONFIRM_SAVESTRING / OEDIT_CONFIRM_SAVEDB** — constants reserved; C code paths are commented out (`:1208-1209`) or entirely unused. Quit goes straight to `cleanupOlc` matching C live behavior.
- **OEDIT_AFFECT_RIS** — reachable-by-name but dead-code in shipped C (`:1916-1924` range-checks 0..32 then returns without mutating). Iota slot reserved; Go case panics via `util.Bug` with "unreachable path".
- **`medit_disp_aff_flags` / `medit_disp_ris` shared renderers** — soft-dependency on `plan-phase6-olc-medit.md`. Until medit lands, oedit affect-flag editing prompts a placeholder message and accepts only `0` (cancel). Full editing unblocks when medit lands.
- **ENABLE_OLC2_EXTRAS auto-create-on-empty-arg** — C `#ifdef`-gated block at `:136-169` that auto-allocates a new object in the builder's area when `oedit` is invoked with no arg. Stock SMAUG has this OFF. Go matches stock: no-arg prints usage/error, no auto-create.
- **GSC (gold/silver/copper)** — `#ifdef ENABLE_GOLD_SILVER_COPPER` arms in val1_menu (ITEM_GOLD / ITEM_SILVER / ITEM_COPPER) and OEDIT_COST. Out of Phase 6 scope (consistent with auction plan §Scope Cuts).
- **ANSI screen-clear sequences** — all `write_to_buffer(d, "50\x1B[;H\x1B[2J", 0)` calls deliberately omitted per redit Q1 precedent.
- **Building-log file persistence** — `olcLog` routes through `util.LogStringPlus` only; dedicated `build.log` file deferred (same as redit).
- **`can_omodify` fine-grained permission model** — C checks object-ownership + level-range. Go `DoOedit` uses `GetTrust() >= LEVEL_IMMORTAL` single gate. Preserve simpler check; deferred to admin-plan.
- **Double-edit guard** — C `:186-192` rejects when another descriptor is editing the same vnum. Same deferral as redit (single-builder-at-a-time acceptable for now).
- **`oeditShow` removal** — decide during implementation: if no callers depend on `oedit <vnum>` printing a flat summary, `oeditShow` is removed. If callers exist, rebind to `oedit <vnum> show`.

---

## Open Questions

**Q1. ANSI screen-clear in menu redisplays.** Omit (matches redit Q1, precedent already established).

**Q2. `oeditShow` flat-summary behavior.** Currently `DoOedit(ch, "<vnum>")` calls `oeditShow`. This plan replaces that with menu-entry. **Recommended:** remove `oeditShow` unless test suite relies on it; if it does, rebind to `oedit <vnum> show` (flat subcommand). Verify during G10 implementation.

**Q3. `ObjIndexData.FirstExtradesc` vs instance `Obj.FirstExtradesc`.** C iterates both; Go port iterates prototype only (consistent with §Go Design). **Recommended:** prototype-only. Document in CHANGELOG.

**Q4. `medit_disp_aff_flags` / `medit_disp_ris` soft dependency.** Affect-flag / RIS editing shipped in this plan with a placeholder. **Recommended:** ship the placeholder, flag as TODO item tied to medit-plan landing. Acceptance criteria A17/A18 cover the seam but not the full-featured editor.

**Q5. `skillLookup` / `bsearch_skill_exact` Go equivalents.** Used in OEDIT_VALUE_3/4/5/6 for spell lookups. Verify during implementation — if absent, use `internal/skill` table lookups or port a minimal `skillLookup(name string) int` helper. **Recommended:** port the helper if missing (small scope), test independently.

**Q6. `ObjIndexData.ActionDesc` field naming.** C uses `obj->action_desc`. Verify Go field name in `internal/types/object.go`. **Recommended:** if named differently (e.g. `ActionDesc` vs `ActDesc`), use the Go name; document in §Go Current State.

**Q7. Prototype-flag `is_name("protoflag", bestowments)` helper.** Verify `util.IsName` handles bestowments list. **Recommended:** reuse if exists; else port from `util.IsName` pattern.

**Q8. OEDIT_EXTRAS numeric-input early-break behavior.** C at `:1401-1403` processes only ONE flag per numeric input, but processes ALL words on word-input. Preserve this asymmetry? **Recommended:** preserve verbatim. Pin with two tests: numeric-processes-one, word-processes-all.

**Q9. `MAX_OLC_ITEMS_LIST` constant.** Used in OEDIT_VALUE_2 container branch for range-check `[0, MAX_OLC_ITEMS_LIST]`. Defined in C at `src/olc.h:48` as `61` (NOT `src/mud.h` — audit correction 2026-04-19). Verify existence in Go; if absent, add as const in `internal/types/constants.go`. **Recommended:** grep + add if missing.

**Q10. `skillLookup` must be exported to `game` package.** Current Go package layout may require a seam. **Recommended:** add a package-level func var `skillLookupFunc` in `oedit_parse.go` wired from boot, same pattern as `worldObjLookup`.

**Q11. `editObjectAffect(d, n)` helper — C uses a fixed index across `pIndexData->first_affect + obj->first_affect` chain. Go iterates `idx.FirstAffect` only (prototype). Any test that previously relied on instance-affect indexing breaks. Pin with regression test.

**Q12. `oeditFindExtradesc(idx, n)` helper — same concern as Q11 for extradescs. Prototype-only.

---

## C Bug Catalog (per-policy = preserve verbatim + pin-test; fix-in-Go flagged explicitly)

**Bug 1 — OEDIT_AFFECT_RIS is declared but unreachable in shipped C.**
C at `:1916-1924` only runs a range-check and returns without mutating any state. The actual RIS bit-accumulation happens in `OEDIT_AFFECT_MODIFIER` via the `APPLY_RESISTANT|APPLY_IMMUNE|APPLY_SUSCEPTIBLE` sub-switch at `:1821-1852`. The OEDIT_AFFECT_RIS mode is never set by any other arm.
**Policy: preserve verbatim.** Reserve the iota slot. Go case emits `util.Bug("Oedit_parse: OEDIT_AFFECT_RIS unreachable path")` and redisplays main menu.
Pin: `TestOeditParse_AffectRis_UnreachableBugs` (drives the mode directly via test-only mode-setter and asserts the bug path).

**Bug 2 — OEDIT_CONFIRM_SAVESTRING commented-out in main-menu Q.**
C at `:1207-1211` has the `OEDIT_CONFIRM_SAVESTRING` transition commented out (`/* send_to_char(...); OLC_MODE(d) = OEDIT_CONFIRM_SAVESTRING; */`). Quit goes straight to `cleanup_olc`.
**Policy: preserve.** Go matches: `Q` → `cleanupOlc(d)` with no save-confirmation prompt. Constant reserved for future re-activation.

**Bug 3 — OEDIT_TYPE accepts `number == 0` as "invalid" but the prompt says "Invalid choice, try again" without returning to the main menu.**
C at `:1344-1348` rejects-and-reprompts, staying in OEDIT_TYPE. User can type `1..MAX_ITEM_TYPE-1` at the next prompt.
**Policy: preserve.** Go matches.

**Bug 4 — OEDIT_VALUE_1 ITEM_MONEY "Amount of Gold coins" prompt missing `break`.**
C at `:665-672` falls through from ITEM_MONEY to ITEM_SILVER/ITEM_COPPER (under `#ifdef ENABLE_GOLD_SILVER_COPPER`). Without the `#ifdef`, the fall-through reaches only the shared `break` at the bottom. In stock SMAUG with GSC off, behavior is correct. GSC is scope-cut regardless.
**Policy: preserve (no-op in stock build).**

**Bug 5 — OEDIT_VALUE_2 ITEM_POTION missing `break` AT C `:1578` ... actually wait, it has one.** (Self-check during authoring; no bug found here.) Skip.

**Bug 6 — ENABLE_OLC2_EXTRAS auto-create block at `:156` assigns `pArea = ch->pcdata->area` AFTER calling `make_object` which used `ch->in_room->area`. This means the save target is wrong.**
Not applicable to Go-port (block is `#ifdef`-gated and OFF in stock). **Policy: scope-cut.**

**Bug 7 — `oedit_disp_prog_choice` at `:381-404` uses `strcat(buf, ...)` on uninitialized `buf` at line 390.**
`buf[0] = '\0'` is set on the ELSE branch (`:395`) but the IF branch at `:388` calls `strcat(buf, "\n\r")` with `buf` uninitialized. Undefined behavior in C.
**Policy: scope-cut with mpedit** (this is mudprog-specific — defer to mpedit plan).

**Bug 8 — `do_ocopy` at `:210-340` leaks affect/extradesc clones when the copy is rolled back by an error.**
Out of scope for this plan (ocopy is a separate flat command). Noted only for completeness.

---

## Risk Analysis

- **High overall.** 37 modes × per-item-type dispatch = ~75 possible `(mode, item_type)` pairs to exercise. Risk is (a) state-confusion (wrong mode transition), (b) per-item-type dispatch omissions (an item_type that never gets a prompt), (c) CON_OEDIT restore-after-editor missing on one of three editor paths (long/action/extra).
- **Mitigation:**
  - G12 integration tests drive main-menu round-trips.
  - G5 + G8 + G9 unit tests pin each editor-trampoline closure independently.
  - The per-item-type tables in §C Reference are authoritative — G7 tests should iterate every entry.
  - 37 mutation gates are plausible; 15 are pinned here. Adversary may ask for more.
- **Highest-risk single point:** editor-callback CON_OEDIT restore (three closures: SUB_OBJ_LONG, SUB_OBJ_ACTION, SUB_OBJ_EXTRA). Mutation #11 + #12 specifically pin two of the three; adversary should require a third.
- **Cross-package seam count:** adds `OeditDispMenuFunc`, `worldObjLookup`, possibly `skillLookupFunc`. Follows the redit precedent. `boot_test.go` non-nil-post-boot assertion needs one new entry for `OeditDispMenuFunc`.
- **Path-traversal fix (G8)** is a SECURITY fix that lands as a side-effect of this plan. Worth highlighting in CHANGELOG.
- **medit soft-dependency** — affect-flag editing placeholder is functional but reduced-featured. Not a regression (flat `DoOedit` has no full affect-flag surface either); some builders may miss parity with stock C. Acceptable.

---

## Test Plan Summary

- **Unit tests (~90):**
  - Types/olc_test.go: +4 (OEDIT_* constants).
  - Game/oedit_menu_test.go: +11 (one per menu renderer, contains-expected-strings).
  - Game/oedit_parse_test.go: +~50 (main-menu branches × per-item-type value branches × affect/extradesc flows).
  - Game/oedit_value_menus_test.go: +~15 (one per value-menu × item-type table row).
  - Act/olc_interactive_test.go: +4 (menu-entry extension).
  - Act/olc_test.go: +4 (`DoSaveArea` path-containment — 3 reject + 1 accept).
  - Util/olclog_test.go: +1 (target-label format).
  - Boot/boot_test.go: +1 (`OeditDispMenuFunc` wired).
- **Mutation tests (15):** all via `Edit` round-trips; no banned git commands. See §Mutation Gates.
- **E2E tests (4):** testclient `oedit_test.go` — menu-entry-and-quit / set-name / invalid-choice / type-submenu.
- **Regression tests:** existing `olc_interactive_test.go` + `olc_set_test.go` continue to pass without modification.
- **Full suite:** `go test -count=3 ./...` green across all 15 packages.

---

## File Budget

**New files (estimated LOC):**

| File | LOC |
|---|---|
| `internal/game/oedit_menu.go` | ~600 |
| `internal/game/oedit_parse.go` | ~900 |
| `internal/game/oedit_value_menus.go` | ~300 |
| `internal/game/oedit_menu_test.go` | ~450 |
| `internal/game/oedit_parse_test.go` | ~1500 |
| `internal/game/oedit_value_menus_test.go` | ~500 |
| `internal/testclient/oedit_test.go` | ~200 |
| Total new | **~4450** |

**Modified files (estimated delta):**

| File | Delta LOC |
|---|---|
| `internal/types/olc.go` | +45 (OEDIT_* const block + reservation pins) |
| `internal/types/olc_test.go` | +80 |
| `internal/types/enums.go` | +1 (SUB_OBJ_ACTION only; SUB_OBJ_LONG pre-exists at :103) |
| `internal/game/loop.go` | +3 (CON_OEDIT arm) |
| `internal/game/redit_parse.go` | +10 / -5 (olcLog target-label parameter + all callers) |
| `internal/act/olc.go` | +5 (OeditDispMenuFunc seam + path guard imports) |
| `internal/act/olc.go` `DoSaveArea` | +10 (path-containment guard) |
| `internal/act/olc_test.go` | +80 (path-traversal pins) |
| `internal/act/olc_interactive.go` | +15 / -3 (vnum-only menu-entry branch) |
| `internal/act/olc_interactive_test.go` | +40 |
| `internal/boot/boot.go` | +1 (OeditDispMenuFunc wiring) |
| `internal/boot/boot_test.go` | +4 (seam-non-nil) |
| Total modified | **~295** |

**Grand total:** ~4750 LOC of new/modified code, of which ~2700 are tests (~57%).

---

## Cross-References

- **Precedent:** `smaug-go/doc/plan-phase6-olc-redit.md` (476 lines, LANDED `f678e8a` 2026-04-19).
- **Downstream:** `smaug-go/doc/plan-phase6-olc-medit.md` (TBW — will inherit OEDIT_* iota-offset convention → MEDIT_* = iota+300).
- **Downstream hard-blocked:** `smaug-go/doc/plan-phase6-olc-mpedit.md` (TBW — hard-blocked on this plan landing; OEDIT_MPROGS* iota slots reserved here).
- **C source:** `src/ooedit.c:1-2160`, `src/olc.h:51-150`, `src/mud.h:6308`, `src/smaug.c:1641-1652`.
- **Go existing:** `internal/types/olc.go`, `internal/game/loop.go:273-278`, `internal/game/redit_parse.go`, `internal/game/redit_menu.go`, `internal/act/olc_interactive.go:16-200`, `internal/act/olc_set.go:207+`, `internal/act/olc.go:799-830`.
- **Roadmap row:** `smaug-go/doc/phase6-roadmap.md` — OLC oedit.
- **Security context:** commit `b4c7477` (2026-04-19 post-Phase-6 audit) flagged `DoSaveArea` path-traversal as future-risk. This plan lands the fix as §G11.

---

## Adversary-Resolved Concerns

_Pending adversary review. `Agent` tool availability remains inconsistent per persistent environment gap tracked in TODO.md — structured self-review substituted during authoring. Self-review notes:_

- Verified C line numbers for `do_ooedit` (:109-208), `oedit_parse` (:1172-2160), each `oedit_disp_*` helper, each OEDIT_* mode switch case, and each per-item-type value-dispatch arm.
- Verified OEDIT_* constant numbering from `src/olc.h:114-150` — all 37 values.
- Verified existing Go state: `DescriptorData.Olc` present (descriptor.go:77-79); `CON_OEDIT` defined (enums.go:90); loop.go:273-278 has CON_REDIT arm; `DoOedit` flat path operational at olc_interactive.go:16-200.
- Verified `olcLog` can be generalized backward-compatibly with a target-label parameter (current call sites at redit_parse.go:68, :80, :85, etc. are all single-target "ROOM").
- Verified Q4 strategy (medit soft-dep placeholder) is consistent with flat `DoOedit` which has no affect-flag surface either.
- Flagged G11 path-containment as a SECURITY fix (not just a quality improvement).
- Scope-cut mudprog menu and GSC arms; both are separate plans / feature toggles.
- No `git checkout` / `git restore` / `git stash` / `git reset --hard` in mutation-verify steps — all revert via `Edit` round-trips.
- **Pre-execution gate:** G10 implementation must verify Q2 (does any caller depend on old `oedit <vnum>` → `oeditShow` behavior?) before modifying. If yes, rebind to `oedit <vnum> show`.

## Completion Record

*To be filled on landing.*
