# Plan: Phase 6 Interactive OLC — `CON_MEDIT` Substate (Menu-Driven Mob/Character Editor)

**Status:** Authored 2026-04-19. Pending adversary review before dispatch.
**Priority:** P2 (Phase 6, Wave D — Interactive OLC, third of three editors).
**Scope:** Interactive mob/character-editor nanny substate `CON_MEDIT`. Ports `src/omedit.c` (2280 LOC) and the 64 `MEDIT_*` mode constants from `src/olc.h:198-263` into Go. Inherits the nanny-dispatch pattern + `OlcData`-on-descriptor allocation + `CON_*` loop arm established by `plan-phase6-olc-redit.md` (LANDED 2026-04-19) and `plan-phase6-olc-oedit.md` (LANDED/in-flight 2026-04-19). Unlike redit (rooms only) and oedit (objects only), medit edits BOTH non-player mobs (NPC, IS_NPC flag set) AND connected player characters (PC, IS_NPC flag clear) through two parallel main-menu arms inside the same parser.

---

## Cross-Plan Dependencies

**Depends on (LANDED or in-flight):**

- `plan-phase6-olc-redit.md` commit `f678e8a` 2026-04-19 — establishes `OlcData` struct, `OlcData.Mode` iota-space convention (REDIT_* = iota+100), `DescriptorData.Olc` field, `CON_REDIT` arm in `internal/game/loop.go:273-278` dispatch switch, `ReditDispMenuFunc` seam pattern in `internal/boot/boot.go:151-154`, `olcLog` helper, `worldRoomLookup` seam, `uRange` helper, EditorSave restore-CON_REDIT trampoline pattern (medit copies the pattern verbatim for CON_MEDIT restore).
- `plan-phase6-olc-oedit.md` — establishes OEDIT_* = iota+200 convention, `worldObjLookup` seam, bitmask-editor shape for `OEDIT_EXTRAS`/`OEDIT_WEAR`, dual-list extradesc + affect editors (iterating both `pIndexData` AND instance), dest-buf-on-CharData stash pattern, and placeholders that medit explicitly back-fills:
  - `OEDIT_AFFECT_MODIFIER` currently accepts numeric modifier only (oedit Wave-3 placeholder) — medit G14 must rewire it to call the shared bitmask helper landed in G9.
  - `oedit.APPLY_EXT_AFFECT` explicit reject branch — revisit per C parity in G14.
- `plan-phase5-tier12-editor-save.md` LANDED — `EditorSave` callback on `CharData` + `/s` handler in `internal/game/editor.go:163-172` that transitions `Connected=CON_PLAYING` BEFORE invoking the closure (closure re-sets CON_MEDIT after `StopEditing`).
- Existing flat `DoMedit` at `internal/act/olc_interactive.go` (if present — grep required at G1 start). If absent, menu-entry path is the first port of medit; no flat-path preservation required.
- `bcrypt.GenerateFromPassword` via `golang.org/x/crypto/bcrypt` + `act.BcryptCost` — confirmed at `internal/act/playercfg.go:260` and `internal/game/loop.go:492` (Q2 RESOLVED — `util.CryptPassword` does not exist; bcrypt is the actual primitive).

**Blocks (hard):**

- `plan-phase6-olc-mpedit.md` — mob-program editor. Hard-blocked by the `MEDIT_MPROGS*` family being scope-cut from this plan (see §Scope Cuts). mpedit must land AFTER medit establishes the CON_MEDIT shape. No MEDIT_MPROGS iota slot is reserved here because `src/olc.h:198-263` does not define one for mobs (C mob-mprog editing lives under a separate `CON_MPROG_EDIT` substate; mpedit plan will introduce it at iota+400).

**Soft dependencies:**

- Phase 4b `DoMcreate` — already covers new-mob creation via flat `mset <vnum> create <name>` path (grep at G1). Menu-entry CANNOT create new mobs.
- `util.ClassLookup` / `util.RaceLookup` — may not exist yet. Mirror oedit Q5 Option 2: numeric-only with TODO follow-up if absent.
- `handler.AffectModify` — affect add/remove primitive. Reused verbatim from oedit G10.
- `util.SmashTilde` — strips `~` from builder input before persistence.

---

## Problem

The constant `CON_MEDIT` is defined in `internal/types/enums.go` (iota value follows CON_OEDIT). The pulse-loop dispatch switch in `internal/game/loop.go` has arms for `CON_REDIT` (landed) and `CON_OEDIT` (landed/in-flight) but **no `case types.CON_MEDIT:` arm**. Any descriptor that transitions to `CON_MEDIT` falls through to the nanny `default:` branch and is disconnected with "Unexpected state. Disconnecting.".

The flat command path (if any exists in Go) implements only a subset of the ~64 MEDIT_* modes. A builder who wants to change multiple fields re-types `mset <victim> <field> <value>` each time. Editing PC fields from `mset` is particularly error-prone because field parsing is identical to NPC fields but semantics differ (e.g. `class` on an NPC is a mobprog-Class index; `class` on a PC is the player class selection).

The C flow is the opposite. `do_omedit` at `src/omedit.c` sets `d->connected = CON_MEDIT`, stashes the victim on `d->character->dest_buf`, inspects `IS_NPC(victim)` to choose between `medit_disp_npc_menu` and `medit_disp_pc_menu`, and transitions into the menu state. Every subsequent line arrives at `medit_parse` and is dispatched against `OLC_MODE(d)` — a per-descriptor menu-state enum with **64 MEDIT_* values** (see `src/olc.h:198-263`).

C dispatch at `src/smaug.c` (same switch that routes CON_REDIT / CON_OEDIT) routes `d->connected == CON_MEDIT` to `medit_parse(d, argument)`. The nanny is NOT involved; CON_MEDIT short-circuits it — same pattern as CON_REDIT / CON_OEDIT.

Consequence of the Go gap: no interactive mob/character editor exists. Immortals edit mobs through flat `mset` (if ported) or through direct `handler.AffectModify` calls. PCs are edited only through specialized commands (`DoBestow`, `DoAdvance`). This plan adds the menu-driven editor while keeping any existing flat path untouched.

---

## C Reference (authoritative)

All line numbers to be verified 2026-04-19 from `src/omedit.c` (2280 LOC) and `src/olc.h` (293 LOC). Per anti-stall protocol, detailed line-citations are batched — below are the 6 spot-checked C-function anchors plus a TBD-citation list for per-G verification.

### Entry points (line-verified 2026-04-19)

- **`do_omedit`** — `src/omedit.c:109-232`. [C AUDIT: actual function name is `do_omedit`, not `do_mmedit`.] Parses optional `<vnum|name>`, resolves the target (NPC prototype by vnum via `get_char_world` at `:180` OR connected PC by name). Trust gate: `!IS_NPC(victim) && get_trust(ch) < sysdata.level_modify_proto` → reject at `:196-200` ("Huh?"). Double-edit guard at `:203-210`: walks `first_descriptor`, compares `d->connected == CON_MEDIT && OLC_VNUM(d) == victim->pIndexData->vnum`. `can_mmodify` check at `:212-213` (external ACL helper). Allocates `OLC_DATA` at `:216`. Branches at `:217-223`: NPC → stash `OLC_VNUM(d) = victim->pIndexData->vnum`; PC → `medit_setup_arrays()` (array provisioning — see §Q1 resolution below for Go-port divergence). Stashes `d->character->dest_buf = victim` at `:225`, sets `d->connected = CON_MEDIT` at `:226`, clears `OLC_CHANGE(d) = FALSE`, calls `medit_disp_menu(d)` at `:228`, and emits an AT_ACTION "starts using OLC" room message at `:230`.
- **`do_medit_reset`** — `src/omedit.c:947-983`. The `last_cmd` callback invoked by `edit_buffer` on `/s`. Handles `SUB_MOB_DESC` substate for L_DESC updates. Go port: reuse the `EditorSave`-closure trampoline from redit/oedit; no new substate constants strictly required if the closure can capture which field was in flight.
- **`medit_parse`** — `src/omedit.c:985-2160`. Top-level state machine keyed on `OLC_MODE(d)`. Every case mutates the victim and either returns (stay in same mode) or `break`s (falls to the end, re-displays the appropriate NPC/PC main menu).
- **`medit_disp_menu`** — `src/omedit.c:814-825`. Thin dispatcher: `if (!IS_NPC(victim)) medit_disp_pc_menu(d); else medit_disp_npc_menu(d);` [C AUDIT: condition is negated-IS_NPC, i.e. non-NPC goes to PC menu; functionally same result].
- **`medit_disp_npc_menu`** — `src/omedit.c:827-885`. Renders NPC main menu (SPEC, DEFAULT_POS, ATTACK, DEFENSE, STATS, flags, RIS, parts, etc.).
- **`medit_disp_pc_menu`** — `src/omedit.c:887-945`. Renders PC main menu (CLASS, RACE, PRACTICE, PASSWORD, SAVE_MENU, PC_FLAGS, PCDATA_FLAGS, COPPER/SILVER, FAVOR, etc.).
- **`medit_disp_ris`** — `src/omedit.c:502-533`. RIS_* bitmask renderer (shared with oedit via extern).
- **`medit_disp_attack_menu`** — `src/omedit.c:535-596`. Attack-type picker (`attack_table`, 18 entries).
- **`medit_disp_aff_flags`** — `src/omedit.c:637-812`. AFF_* bitmask renderer (shared with oedit via extern; oedit currently has a G9 placeholder that this plan back-wires at G14).
- **MEDIT_PASSWORD arm** — `src/omedit.c:1601-1623`. Trust gate `LEVEL_SUB_IMPLEM` at `:1602`. Min-length 5 at `:1604`. Hash: `pwdnew = sha256_crypt(arg, victim->name)` at `:1609`. Post-hash tilde check at `:1610-1616` (REJECTS if hash contains `~`). Assignment: `DISPOSE(victim->pcdata->pwd); victim->pcdata->pwd = str_dup(pwdnew);` at `:1618-1619`. Persists via `save_char_obj(victim)` if `SV_PASSCHG` set at `:1620-1621`. Logs "Modified password" (no value) at `:1622`.
- **C dispatch call site** — `src/smaug.c` (same switch as CON_REDIT / CON_OEDIT).

### MEDIT_* mode constants (from `src/olc.h:198-263`)

```
MEDIT_NPC_MAIN_MENU           0   (shown for IS_NPC)
MEDIT_PC_MAIN_MENU            1   (shown for connected PC)
MEDIT_NAME                    2
MEDIT_S_DESC                  3
MEDIT_L_DESC                  4
MEDIT_D_DESC                  5
MEDIT_NPC_FLAGS               6   (ACT_* bitmask editor)
MEDIT_PC_FLAGS                7   (PLR_* bitmask editor)
MEDIT_AFF_FLAGS               8   (AFF_* bitmask editor — shared with oedit G9)
MEDIT_CONFIRM_SAVESTRING      9   (LIVE for PC-exit save-confirm; NPC Q skips per C omedit.c:1011-1057 — this DIFFERS from oedit's reserved-but-dead OEDIT_CONFIRM_SAVESTRING)
MEDIT_SEX                    10
MEDIT_HITROLL                11
MEDIT_DAMROLL                12
MEDIT_DAMNUMDIE              13
MEDIT_DAMSIZEDIE             14
MEDIT_DAMPLUS                15
MEDIT_HITNUMDIE              16
MEDIT_HITSIZEDIE             17
MEDIT_HITPLUS                18
MEDIT_AC                     19
MEDIT_GOLD                   20
MEDIT_POS                    21
MEDIT_DEFAULT_POS            22   (NPC-only — default resting posture)
MEDIT_ATTACK                 23   (NPC-only — attack-type index)
MEDIT_DEFENSE                24   (NPC-only — defense flags)
MEDIT_LEVEL                  25   (LEVEL_GREATER-gated for PC target)
MEDIT_ALIGNMENT              26
MEDIT_STRENGTH               27
MEDIT_INTELLIGENCE           28
MEDIT_WISDOM                 29
MEDIT_DEXTERITY              30
MEDIT_CONSTITUTION           31
MEDIT_CHARISMA               32
MEDIT_LUCK                   33
MEDIT_CLAN                   34
MEDIT_DEITY                  35
MEDIT_COUNCIL                36
MEDIT_SPEC                   37   (NPC-only — `spec_fun` name)
MEDIT_RESISTANT              38   (RIS_* bitmask editor)
MEDIT_IMMUNE                 39   (RIS_* bitmask editor)
MEDIT_SUSCEPTIBLE            40   (RIS_* bitmask editor)
MEDIT_PCDATA_FLAGS           41   (PCFLAG_* bitmask editor — PC-only)
MEDIT_MENTALSTATE            42
MEDIT_EMOTIONAL              43
MEDIT_THIRST                 44
MEDIT_FULL                   45
MEDIT_DRUNK                  46
MEDIT_PARTS                  47   (PART_* bitmask editor — body-part flags)
MEDIT_FAVOR                  48
MEDIT_HITPOINT               49
MEDIT_MANA                   50
MEDIT_MOVE                   51
MEDIT_PRACTICE               52   (PC-only)
MEDIT_PASSWORD               53   (PC-only; security-critical — see §G13)
MEDIT_SAVE_MENU              54   (PC-only; 5-way sub-dispatch to SAV1..SAV5)
MEDIT_SAV1                   55
MEDIT_SAV2                   56
MEDIT_SAV3                   57
MEDIT_SAV4                   58
MEDIT_SAV5                   59
MEDIT_CLASS                  60   (PC-only; LEVEL_GREATER-gated)
MEDIT_RACE                   61   (PC-only; LEVEL_GREATER-gated)
MEDIT_SILVER                 62
MEDIT_COPPER                 63
```

**Go mapping: `MEDIT_NPC_MAIN_MENU = iota + 300`.** 64 consecutive iota slots.

### Menu rendering functions (TBD — to be cited per-G during execution)

| C function (expected) | Purpose |
|---|---|
| `medit_disp_npc_menu` | NPC main menu — renders SPEC, DEFAULT_POS, ATTACK, DEFENSE, STATS block, NPC_FLAGS, AFF_FLAGS, RIS, PARTS, etc. |
| `medit_disp_pc_menu` | PC main menu — renders CLASS, RACE, PRACTICE, PASSWORD, SAVE_MENU, PC_FLAGS, PCDATA_FLAGS, COPPER/SILVER, FAVOR, etc. |
| `medit_disp_sex_menu` | Sex enum picker (0/1/2 → NEUTER/MALE/FEMALE; see §post-phase6-vision for SOGI-migration concerns) |
| `medit_disp_pos_menu` | Position enum picker (POS_DEAD..POS_STANDING) |
| `medit_disp_aff_flags` | AFF_* bitmask editor (SHARED with oedit; landed by oedit-G9 OR landed here first and back-wired) |
| `medit_disp_ris` | RIS_* bitmask editor (SHARED with oedit) |
| `medit_disp_save_menu` | SAV1..SAV5 sub-dispatch (PC only) |
| `medit_disp_attack_menu` | NPC attack-type picker (`attack_table` — 18 entries) |
| `medit_disp_defense_menu` | NPC defense-type picker |
| `medit_disp_spec_menu` | NPC spec_fun name picker (list of `spec_*` function names) |
| `medit_disp_class_menu` | PC class picker (`class_table`) |
| `medit_disp_race_menu` | PC race picker (`race_table`) |

Per-G each renderer gets a dedicated Go function (e.g. `meditDispNpcMenu(d *DescriptorData)`). All renderers live in `internal/game/medit_menu.go`.

---

## Go Current State

| Artifact | Location | Status |
|---|---|---|
| `CON_MEDIT` constant | `internal/types/enums.go` | Exists (iota value after CON_OEDIT; verify at G1) |
| Loop dispatch arm | `internal/game/loop.go` | **MISSING** — G1 deliverable |
| `MEDIT_*` iota table | `internal/types/olc.go` | **MISSING** — G1 deliverable (64 slots, iota+300) |
| `OlcData.Mode` | `internal/types/olc.go` | Exists (landed by redit); reused unchanged |
| `OlcData.Victim *CharData` | `internal/types/olc.go` | **DESIGN NOTE: existing OlcData uses `Target any` for all editor targets (redit=*RoomIndexData, oedit=*ObjIndexData). G1 must use `Target any` with type-assertion for victim — NOT add a new typed Victim field; that would break the established `any` union pattern. Remove references to OlcData.Room/OlcData.Obj (those fields do not exist either — see Target any).** |
| `meditParse` entry point | `internal/game/medit_parse.go` | **NEW FILE** — G4 deliverable |
| `MeditDispNpcMenu` / `MeditDispPcMenu` | `internal/game/medit_menu.go` | **NEW FILE** — G2 deliverable |
| `MeditDispMenuFunc` seam | `internal/act/olc.go` | **MISSING** — G1 deliverable (seam declaration) |
| Boot wire | `internal/boot/boot.go` | **MISSING** — G1 deliverable (assigns seam + `game.SetWorldRef` if new) |
| `DoMedit` no-arg menu-entry | `internal/act/olc.go` (or olc_interactive.go) | **MISSING** — G14 extends existing flat path (if any) |
| `worldMobLookup` seam | `internal/game/medit_parse.go` | **MISSING** — G1 deliverable (NPC prototype lookup by vnum) |
| `worldPcLookup` seam | `internal/game/medit_parse.go` | **MISSING** — G1 (connected-descriptor lookup by name) |
| `bcrypt.GenerateFromPassword` | `internal/act/playercfg.go:260` | **RESOLVED** — `golang.org/x/crypto/bcrypt` + `act.BcryptCost`. `util.CryptPassword` does not exist. G13 uses `bcrypt.GenerateFromPassword` directly (as shown in §Password path). |
| Existing `DoMset` flat path | `internal/act/` | **VERIFY at G1** — preserve; menu-entry is additive |

---

## Go Design

### Dispatch shape

Identical to CON_REDIT / CON_OEDIT. `internal/game/loop.go` gains a new arm:

```go
case types.CON_MEDIT:
    meditParse(d, line)
    return
```

placed immediately after the `CON_OEDIT` arm. The arm short-circuits the default nanny-fallthrough exactly as redit / oedit do.

### `meditParse` top-level

```go
func meditParse(d *types.DescriptorData, arg string) {
    if d == nil || d.Olc == nil || d.Character == nil {
        return  // defensive; should never hit in valid flow
    }
    victim := d.Olc.Victim  // stashed by DoMedit; NOT d.Character (which is the editor)
    if victim == nil {
        sendToChar(d.Character, "No victim.\n\rExiting editor.\n\r")
        cleanupOlc(d)
        return
    }
    arg = strings.TrimSpace(arg)

    switch d.Olc.Mode {
    case types.MEDIT_NPC_MAIN_MENU:
        meditDispatchNpcMain(d, victim, arg)
    case types.MEDIT_PC_MAIN_MENU:
        meditDispatchPcMain(d, victim, arg)
    case types.MEDIT_NAME:
        meditArmName(d, victim, arg)
    // ... 62 more arms
    default:
        olcLog(d, "meditParse: unknown mode %d", d.Olc.Mode)
        d.Olc.Mode = npcOrPcMenu(victim)
        MeditDispMenuFunc(d)
    }
}
```

where `npcOrPcMenu(v)` returns `MEDIT_NPC_MAIN_MENU` or `MEDIT_PC_MAIN_MENU` based on `victim.IsNPC()` (method on `*CharData` at `internal/types/character.go:374-375` — audit confirmed 2026-04-19, plan originally said `util.IsNpc(v)` which does not exist).

### PC-vs-NPC mode ownership — DERIVED FROM C DIGIT MAPPING

Verified 2026-04-19 from `src/omedit.c:1059-1215` (NPC main menu) and `:1218-1386` (PC main menu).

**NPC main menu digit → mode mapping (C `:1062-1214`):**

| Digit | C mode | Notes |
|---|---|---|
| Q | cleanup_olc | exit (no save-confirm for NPC) |
| 1 | MEDIT_SEX | |
| 2 | MEDIT_NAME | |
| 3 | MEDIT_S_DESC | inline single-line |
| 4 | MEDIT_L_DESC | inline single-line; appends `\n\r` |
| 5 | MEDIT_D_DESC | editor trampoline via `SUB_MOB_DESC` |
| 6 | MEDIT_CLASS | |
| 7 | MEDIT_RACE | |
| 8 | MEDIT_LEVEL | |
| 9 | MEDIT_ALIGNMENT | |
| A | MEDIT_STRENGTH | |
| B | MEDIT_INTELLIGENCE | |
| C | MEDIT_WISDOM | |
| D | MEDIT_DEXTERITY | |
| E | MEDIT_CONSTITUTION | |
| F | MEDIT_CHARISMA | |
| G | MEDIT_LUCK | |
| H | MEDIT_DAMNUMDIE | |
| I | MEDIT_DAMSIZEDIE | |
| J | MEDIT_DAMPLUS | |
| K | MEDIT_HITNUMDIE | |
| L | MEDIT_HITSIZEDIE | |
| M | MEDIT_HITPLUS | |
| N | MEDIT_GOLD (or MEDIT_COPPER if `ENABLE_GOLD_SILVER_COPPER`) | |
| O | MEDIT_SPEC | |
| P | MEDIT_SAVE_MENU | NPC **does** have saves in C |
| R | MEDIT_RESISTANT | |
| S | MEDIT_IMMUNE | |
| T | MEDIT_SUSCEPTIBLE | |
| U | MEDIT_POS | |
| V | MEDIT_ATTACK | |
| W | MEDIT_DEFENSE | |
| X | MEDIT_PARTS | |
| Y | MEDIT_NPC_FLAGS | |
| Z | MEDIT_AFF_FLAGS | |

**PC main menu digit → mode mapping (C `:1221-1384`):**

| Digit | C mode | Notes |
|---|---|---|
| Q | MEDIT_CONFIRM_SAVESTRING (if OLC_CHANGE) else cleanup_olc | PC save-confirm is real |
| 1 | MEDIT_SEX | |
| 2 | MEDIT_NAME | triggers `do_pcrename` if trust > LEVEL_SUB_IMPLEM-1 |
| 3 | MEDIT_D_DESC | editor trampoline |
| 4 | MEDIT_CLASS | |
| 5 | MEDIT_RACE | |
| 6 | **NPC Only** — explicit reject ("NPC Only!!") |
| 7 | MEDIT_ALIGNMENT | |
| 8 | MEDIT_STRENGTH | |
| 9 | MEDIT_INTELLIGENCE | |
| A | MEDIT_WISDOM | |
| B | MEDIT_DEXTERITY | |
| C | MEDIT_CONSTITUTION | |
| D | MEDIT_CHARISMA | |
| E | MEDIT_LUCK | |
| F | MEDIT_HITPOINT | |
| G | MEDIT_MANA | |
| H | MEDIT_MOVE | |
| I | MEDIT_GOLD (or MEDIT_COPPER if GSC) | |
| J | MEDIT_MENTALSTATE | |
| K | MEDIT_EMOTIONAL | |
| L | MEDIT_THIRST | |
| M | MEDIT_FULL | |
| N | MEDIT_DRUNK | |
| O | MEDIT_FAVOR | |
| P | MEDIT_SAVE_MENU | |
| R | MEDIT_RESISTANT | |
| S | MEDIT_IMMUNE | |
| T | MEDIT_SUSCEPTIBLE | |
| U | **NPCs Only** — explicit reject ("NPCs Only!!") |
| V | MEDIT_PC_FLAGS | |
| W | MEDIT_PCDATA_FLAGS | |
| X | MEDIT_AFF_FLAGS | |
| Y | MEDIT_DEITY | |
| Z | MEDIT_CLAN | trust-gated `LEVEL_GOD` |
| = | MEDIT_COUNCIL | trust-gated `LEVEL_SUB_IMPLEM` |

**Absent from both main menus in stock C:** MEDIT_PASSWORD, MEDIT_PRACTICE, MEDIT_DEFAULT_POS, MEDIT_NPC_FLAGS-via-PC-menu, MEDIT_HITROLL, MEDIT_DAMROLL, MEDIT_AC, MEDIT_HITNUMDIE-via-PC (PC has different HP). These modes are reachable in C only via the flat `mset` command, OR they are dispatched to but not-yet-menu-rendered (dead-code reserve).

**Go-port ownership rule** derived from the above:

| Mode | Owner | Rationale |
|---|---|---|
| MEDIT_NPC_MAIN_MENU / MEDIT_PC_MAIN_MENU | self | entry points |
| MEDIT_NAME / MEDIT_SEX / MEDIT_AFF_FLAGS / MEDIT_ALIGNMENT | shared | in both menus |
| MEDIT_D_DESC | shared | in both menus (editor trampoline) |
| MEDIT_S_DESC / MEDIT_L_DESC | NPC-only | NPC-menu-only in C |
| MEDIT_CLASS / MEDIT_RACE | shared | in both menus |
| MEDIT_LEVEL / MEDIT_DAMNUMDIE / MEDIT_DAMSIZEDIE / MEDIT_DAMPLUS / MEDIT_HITNUMDIE / MEDIT_HITSIZEDIE / MEDIT_HITPLUS / MEDIT_SPEC / MEDIT_ATTACK / MEDIT_DEFENSE / MEDIT_PARTS / MEDIT_NPC_FLAGS / MEDIT_POS | NPC-only | NPC-menu-only in C |
| MEDIT_STRENGTH..MEDIT_LUCK (7) | shared | in both menus |
| MEDIT_GOLD / MEDIT_COPPER / MEDIT_SILVER | shared | N/I digit (GSC-conditional) |
| MEDIT_RESISTANT / IMMUNE / SUSCEPTIBLE | shared | R/S/T in both |
| MEDIT_SAVE_MENU / SAV1..SAV5 | shared | P in both |
| MEDIT_HITPOINT / MANA / MOVE / MENTALSTATE / EMOTIONAL / THIRST / FULL / DRUNK / FAVOR / PC_FLAGS / PCDATA_FLAGS / DEITY / CLAN / COUNCIL | PC-only | PC-menu-only in C (MOVE/MANA/HITPOINT — NPC's HP comes from dice) |
| MEDIT_PASSWORD / MEDIT_PRACTICE / MEDIT_DEFAULT_POS | reserved | not dispatched from either menu in stock C; Go port implements arm bodies but does not yet wire a menu digit — parse arm activates only if the mode is set via a future flat-path or direct-test |
| MEDIT_CONFIRM_SAVESTRING | shared | Q-handler on PC menu only; NPC Q skips the confirm |

**Enforcement** — every arm body begins:

```go
if victim.IsNPC() && <pc-only-mode> {
    sendToChar(d.Character, "That field doesn't apply to NPCs.\n\r")
    d.Olc.Mode = types.MEDIT_NPC_MAIN_MENU
    MeditDispMenuFunc(d)
    return
}
if !victim.IsNPC() && <npc-only-mode> {
    sendToChar(d.Character, "That field doesn't apply to players.\n\r")
    d.Olc.Mode = types.MEDIT_PC_MAIN_MENU
    MeditDispMenuFunc(d)
    return
}
```

Rejection messages match C `:1255-1257` ("NPC Only!!") and `:1351-1353` ("NPCs Only!!") VERBATIM when dispatched from main-menu digits 6 (PC) / U (PC). For direct-mode-transition arms (future flat-path), use the longer Go-idiom messages above.

### Main-menu dispatch skeleton (G5 NPC, G6 PC)

Each main-menu dispatch parses `arg`, treats `Q`/`q` as quit-to-playing, treats `1`..`9`/`A`..`Z` as submenu transitions, treats unrecognized input as menu redisplay.

```go
func meditDispatchNpcMain(d *types.DescriptorData, victim *types.CharData, arg string) {
    switch strings.ToUpper(arg) {
    case "Q", "":
        cleanupOlc(d)  // transitions d.Connected to CON_PLAYING, clears d.Olc
        return
    case "1": d.Olc.Mode = types.MEDIT_NAME
    case "2": d.Olc.Mode = types.MEDIT_S_DESC
    case "3": d.Olc.Mode = types.MEDIT_L_DESC
    case "4": d.Olc.Mode = types.MEDIT_D_DESC
    // ... (exact mapping matches C medit_disp_npc_menu rendering)
    default:
        sendToChar(d.Character, "Invalid choice.\n\r")
    }
    MeditDispMenuFunc(d)
}
```

Mapping NPC-menu digits → MEDIT_* modes is TBD per C `medit_disp_npc_menu` display order (verify at G5 start). The Go port MUST match C's digit-to-field mapping so existing builder muscle memory transfers. [C AUDIT: the skeleton above is WRONG — C has digit 1=MEDIT_SEX, digit 2=MEDIT_NAME, digit 3=MEDIT_S_DESC, digit 4=MEDIT_L_DESC, digit 5=MEDIT_D_DESC per omedit.c:1065-1090. The skeleton shows 1=MEDIT_NAME which is incorrect. Worker must NOT copy the skeleton digits literally; verify at G5 start from C source.]

### Password path (G13) — C-parity

C reference `src/omedit.c:1601-1623` uses `sha256_crypt(arg, victim->name)`. Go project convention (verified at `internal/act/playercfg.go:260-266` `DoPassword`, `internal/game/loop.go:415` nanny-new-PC-creation, `internal/game/loop.go:499` nanny) is `bcrypt.GenerateFromPassword([]byte(pwd), BcryptCost)`. Medit G13 uses bcrypt to match Go-port convention (SHA256 would diverge from every other password path in the Go port and break `bcrypt.CompareHashAndPassword` verification in `DoPassword` + nanny login).

```go
func meditArmPassword(d *types.DescriptorData, victim *types.CharData, arg string) {
    if victim.IsNPC() {
        sendToChar(d.Character, "That field doesn't apply to NPCs.\n\r")
        d.Olc.Mode = types.MEDIT_NPC_MAIN_MENU
        MeditDispMenuFunc(d)
        return
    }
    if victim.PCData == nil {
        sendToChar(d.Character, "Victim has no PCData.\n\r")
        d.Olc.Mode = types.MEDIT_PC_MAIN_MENU
        MeditDispMenuFunc(d)
        return
    }
    // Trust gate: C uses LEVEL_SUB_IMPLEM at :1602.
    if d.Character.GetTrust() < types.LEVEL_SUB_IMPLEM {
        d.Olc.Mode = types.MEDIT_PC_MAIN_MENU
        MeditDispMenuFunc(d)
        return
    }
    // Min-length: C uses 5 at :1604. Keep C-parity at 5 (NOT 6, diverges from DoPassword's 6).
    if len(arg) < 5 {
        sendToChar(d.Character, "Password too short, try again: ")
        return  // stay in MEDIT_PASSWORD
    }
    // Defense-in-depth: strip tildes on RAW input before hashing.
    // C checks tildes AFTER hashing — on sha256 output this is near-impossible; on
    // bcrypt output tildes don't appear in the base64 alphabet. Move the check
    // to RAW input for Go. Document as C-divergence (safer).
    if strings.Contains(arg, "~") {
        sendToChar(d.Character, "Unacceptable choice, try again: ")
        return  // stay in MEDIT_PASSWORD
    }
    hash, err := bcrypt.GenerateFromPassword([]byte(arg), act.BcryptCost)
    if err != nil {
        util.Bug("meditArmPassword: GenerateFromPassword: %v", err)
        sendToChar(d.Character, "Password hashing failed.\n\r")
        d.Olc.Mode = types.MEDIT_PC_MAIN_MENU
        MeditDispMenuFunc(d)
        return
    }
    victim.PCData.Pwd = string(hash)
    // Persist: C calls save_char_obj if SV_PASSCHG set. Go port calls SaveFunc unconditionally
    // to match DoPassword's behavior (DoPassword does NOT gate on SV_PASSCHG at :275-277).
    if act.SaveFunc != nil {
        act.SaveFunc(victim)
    }
    olcLog(d, "Modified password for %s", victim.Name)  // NEVER log the password value
    sendToChar(d.Character, "Password updated.\n\r")
    d.Olc.Mode = types.MEDIT_PC_MAIN_MENU
    MeditDispMenuFunc(d)
}
```

**Security-critical invariants** (pin-tested):

1. `victim.PCData.Pwd != "plaintext"` after the call. Asserts bcrypt-hashing happened.
2. `len(victim.PCData.Pwd) > 0` after the call. Asserts assignment happened.
3. `bcrypt.CompareHashAndPassword([]byte(victim.PCData.Pwd), []byte("plaintext")) == nil` — round-trip through the verify path (same path as `DoPassword`).
4. olcLog entry CONTAINS `"Modified password"` but does NOT contain `"plaintext"` or the hash value.
5. Tilde-containing input (`"abc~def"`) rejected BEFORE hashing — stored password UNCHANGED.
6. Too-short input (`"abcd"`) rejected BEFORE hashing — stored password UNCHANGED.

### Shared bitmask-editor helper (G9)

Exported helper signature (placed in `internal/game/olc_bitmask.go`, new file):

```go
// olcBitmaskEdit handles a one-line toggle request for a bitmask field.
// tableName: "ACT_FLAGS", "PLR_FLAGS", "AFF_FLAGS", "PCFLAG", "PART", "RIS".
// current: pointer to the int field being mutated.
// arg: the builder's input line.
// Returns true if a toggle was applied, false if the user typed "done" or "quit".
func olcBitmaskEdit(d *types.DescriptorData, tableName string, current *int, arg string) bool
```

One shared body; 6 wire-up points (NPC_FLAGS/PC_FLAGS/AFF_FLAGS/PCDATA_FLAGS/PARTS/RIS). Each wire passes the correct lookup-table constant. RIS has three separate field pointers (Resistant/Immune/Suscept) but one shared table. `done` / `quit` / empty-line returns to the correct main menu.

### Affect editor (G10)

Full implementation (not a placeholder). Copies oedit G10 shape:

- `MEDIT_AFF_FLAGS` → toggle AFF_* bitmask (already covered by G9).
- Affect-add flow:
  - `MEDIT_AFFECT_MENU` picks Add/Remove/Quit.
  - `MEDIT_AFFECT_LOCATION` picks APPLY_* (SKIP 0 and APPLY_EXT_AFFECT — consistent with oedit `medit_disp_affect_menu:569-591`).
  - `MEDIT_AFFECT_MODIFIER` reads modifier:
    - If APPLY is a scalar (STR/HITROLL/etc.): numeric.
    - If APPLY is APPLY_AFFECT / APPLY_RESISTANT / etc.: dispatch to bitmask helper.
  - `MEDIT_AFFECT_TYPE` picks duration type (unused in stock C; reserve slot per G10 if present in omedit — TBD at G10 start).
- Affect-remove flow reuses the affect-index UI.

(Note: MEDIT_AFFECT_* constants are not in the 0-63 MEDIT_* list above; they live in the extended range `src/olc.h:260-293` or similar and are shared with oedit. Verify at G10 start — if the constants are oedit-owned, reuse; else add MEDIT_AFFECT_MENU etc. in the iota+300 range.)

### Save-throw editor (G11)

Sub-dispatch. `MEDIT_SAVE_MENU` prints:

```
1) Poison/Death     (Save[0] = %d)
2) Wands            (Save[1] = %d)
3) Paralysis/Petri  (Save[2] = %d)
4) Breath           (Save[3] = %d)
5) Spell/Staff      (Save[4] = %d)
Q) Quit

Enter choice:
```

Each digit sets `d.Olc.Mode = MEDIT_SAV1..SAV5`; each SAV arm reads an int, clamps via `uRange`, assigns the named CharData field for that slot (e.g. `victim.SavingPoisonDeath` for SAV1) [C AUDIT: Go uses named fields not an indexed array], redisplays SAVE_MENU.

### Class / Race editors (G12)

```go
func meditArmClass(d *types.DescriptorData, victim *types.CharData, arg string) {
    if victim.IsNPC() { reject }
    if d.Character.GetTrust() < types.LEVEL_GREATER { reject with trust message }
    if arg == "" { render class table; stay in MEDIT_CLASS; return }
    idx, err := strconv.Atoi(arg)
    if err != nil {
        // Try name-lookup IF util.ClassLookup exists; else reject numeric-only.
        if cl, ok := util.ClassLookup(arg); ok { idx = cl } else { reject }
    }
    if idx < 0 || idx >= len(world.ClassTable) { reject }
    victim.Class = idx
    d.Olc.Mode = types.MEDIT_PC_MAIN_MENU
    MeditDispMenuFunc(d)
}
```

Same shape for Race. TODO follow-up if `util.ClassLookup`/`util.RaceLookup` missing.

### EditorSave trampoline (for L_DESC / D_DESC / NAME-multiline)

Identical to redit/oedit G5 contract:

1. Builder chooses `3` (L_DESC).
2. Arm stashes callback: `d.Character.EditorSave = func(txt string) { victim.Description = util.SmashTilde(txt); d.Connected = types.CON_MEDIT; d.Olc.Mode = npcOrPcMenu(victim); MeditDispMenuFunc(d) }`.
3. Arm sets `d.Connected = types.CON_EDITING`, calls `StartEditing(d, victim.Description)`.
4. On `/s`, `editor.go:163-172` transitions `Connected=CON_PLAYING`, calls the closure, which RESETS `Connected=CON_MEDIT` and re-renders the main menu.

Same flow for D_DESC. NAME is inline single-line (no editor trampoline). Short-desc (`S_DESC` for NPC) is inline single-line.

### Boot wiring

`internal/boot/boot.go` gains three statements next to the oedit wire:

```go
act.MeditDispMenuFunc = game.MeditDispMenu
game.SetWorldRef(w)  // already called by redit — idempotent or no-op guard
// (no additional world refs needed unless medit introduces new lookups)
```

---

## Task Groups

### G1 — Schema + loop arm + seams

**Deliverables:**

- `internal/types/olc.go`: add 64 `MEDIT_*` constants iota+300; use existing `Target any` field (type-assert to `*types.CharData` at meditParse entry — matches redit/oedit pattern); no new field needed; `cleanupOlc` already zeroes Target via zero-value reset [C AUDIT: OlcData has no Victim field; Target any is the established pattern].
- `internal/game/loop.go`: add `case types.CON_MEDIT: meditParse(d, line); return` arm directly after `CON_OEDIT`.
- `internal/act/olc.go`: add `MeditDispMenuFunc func(*types.DescriptorData)` seam declaration.
- `internal/game/medit_parse.go`: NEW FILE — `meditParse` skeleton with nil-guard, victim nil-check, default-mode fallback; `worldMobLookup(vnum int) *CharData` seam; `worldPcLookup(name string) *CharData` seam (descriptor walk).
- `internal/game/medit_menu.go`: NEW FILE — `MeditDispMenu` dispatcher (delegates to NPC vs PC renderer based on victim).
- `internal/boot/boot.go`: wire `act.MeditDispMenuFunc = game.MeditDispMenu`.

**Tests:** `TestOlcMeditIotaContiguous` (pins 64 consecutive iota values starting at 300); `TestLoopDispatch_ConMedit_RoutesToMeditParse` (constructs fake descriptor with `Connected=CON_MEDIT`, injects `MeditDispMenuFunc` spy, asserts call); `TestMeditParse_NilGuards` (nil desc / nil Olc / nil Victim all safe).

**Acceptance:** A1 (iota contiguous), A2 (loop arm present), A3 (parse file compiles), A4 (boot wire present).

### G2 — Dual main-menu renderers

**Deliverables:**

- `internal/game/medit_menu.go`: `MeditDispNpcMenu(d)` and `MeditDispPcMenu(d)` — render verbatim from C `medit_disp_npc_menu` / `medit_disp_pc_menu` layouts.
- Shared formatting helpers (ansiHeader, two-column field renderer).
- `MeditDispMenu(d)` dispatcher chooses NPC vs PC.

**Tests:** `TestMeditDispNpcMenu_ContainsFields` (regex-scans output for "Name:", "Short desc:", "Level:", "HP:", "AC:", "Spec:", "Attack:", etc.); `TestMeditDispPcMenu_ContainsFields` (regex-scans for "Class:", "Race:", "Practice:", "Password:", "Saves:", etc.); `TestMeditDispMenu_RoutesByIsNpc` (table-driven).

**Acceptance:** A5 (NPC menu renders all 20+ expected fields), A6 (PC menu renders all 20+ expected fields), A7 (dispatcher routes correctly).

### G3 — Menu renderers (non-main)

**Deliverables:**

- `meditDispSexMenu`, `meditDispPosMenu`, `meditDispDefaultPosMenu`, `meditDispAttackMenu`, `meditDispDefenseMenu`, `meditDispSpecMenu`, `meditDispClassMenu`, `meditDispRaceMenu`, `meditDispSaveMenu`, `meditDispAffectMenu` (for affect editor G10), `meditDispNpcFlagsMenu` / `meditDispPcFlagsMenu` / `meditDispAffFlagsMenu` / `meditDispPcdataFlagsMenu` / `meditDispPartsMenu` / `meditDispRisMenu` (bitmask UIs rendering current mask + toggle hint).
- Lookup tables: `attackNames`, `defenseNames`, `partNames`, `risNames` (verify existing — RisflagNames landed by stances-olc; reuse).

**Tests:** per-renderer snapshot tests (compare against golden output strings).

**Acceptance:** A8 (all 16 submenu renderers present + tested).

### G4 — Dispatcher skeleton

**Deliverables:**

- `meditParse` full 64-case switch (each case currently calls a stub `meditArmXxx` defined in G7-G13).
- `cleanupOlc` adapted to medit (may be shared with redit/oedit; verify — if shared, confirm new field clear works; else add medit-specific cleanup).
- `olcLog` reused unchanged.

**Tests:** `TestMeditParse_EveryModeHasArm` (iterates 300..363, asserts no default hit for valid input); `TestMeditParse_QuitReturnsToPlaying` (from every sub-mode, typing Q returns to main menu; from main menu, Q transitions to CON_PLAYING).

**Acceptance:** A9 (all 64 modes dispatched without default-log), A10 (Q semantics verified).

### G5 — Main-menu dispatch NPC

**Deliverables:**

- `meditDispatchNpcMain(d, victim, arg)` — digit → mode transition table.
- Include current-value redisplay after each field change.
- Reject PC-only fields if builder types the digit that would map to one (C-compat: map only NPC-valid digits; PC-only digits don't appear in NPC menu).

**Tests:** table-driven per-digit → expected mode transition; unrecognized digit shows "Invalid choice.".

**Acceptance:** A11 (every NPC-menu digit 1..9/A..Z maps to a documented MEDIT_* mode).

### G6 — Main-menu dispatch PC

**Deliverables:**

- `meditDispatchPcMain(d, victim, arg)` — PC digit mapping.
- PC menu includes digits for CLASS, RACE, PRACTICE, PASSWORD, SAVE_MENU, PCDATA_FLAGS, COPPER, SILVER, FAVOR.
- LEVEL_GREATER gate on LEVEL / CLASS / RACE: if `d.Character.GetTrust() < LEVEL_GREATER`, reject the digit with "Requires Greater Immortal trust.".

**Tests:** table-driven; trust-gate test at LEVEL_IMMORTAL (reject) and LEVEL_GREATER (accept).

**Acceptance:** A12 (every PC-menu digit maps documented), A13 (trust gate enforced on LEVEL/CLASS/RACE).

### G7 — Simple-field arms

**Deliverables:** 25 inline arms (all C-line-cited at G7 start):

- `MEDIT_NAME` — `omedit.c:1389-1405`. Inline single-line. PC branch: `get_trust(ch) > LEVEL_SUB_IMPLEM-1` → `do_pcrename(ch, "<oldname> <newname>")`. NPC branch: direct `STRFREE/STRALLOC` on `victim->name`. Port both branches; reuse existing `DoPcrename` if present (grep at G7 start) or STRFREE-equivalent.
- `MEDIT_S_DESC` — `omedit.c:1407-1416`. Inline single-line. NPC-only per C (C does not reject PC but the menu never routes a PC to this mode). Port as NPC-only with PC-reject defensively.
- `MEDIT_L_DESC` — `omedit.c:1418-1429`. Inline single-line; appends `"\n\r"` to input (C-parity). NPC-only per menu. Port as NPC-only with PC-reject.
- `MEDIT_D_DESC` — text-editor trampoline. C at `:1082-1090` (NPC digit 5) and `:1238-1246` (PC digit 3) both set `substate = SUB_MOB_DESC` + `last_cmd = do_medit_reset` and invoke the editor. The MEDIT_D_DESC parse-arm at `:1431-1435` is a "should never get here" bug-trap (`cleanup_olc(d); bug("...");`). Go port: implement the editor entry in the main-menu digit-5 dispatch (NPC) / digit-3 dispatch (PC), NOT in the MEDIT_D_DESC parse arm. Set `d.Olc.Mode = MEDIT_D_DESC` to mark the in-flight editor state (defensive; matches C bug-trap semantics — if a line arrives here, bug-log and cleanup). Target field: `victim.Description` (same for NPC and PC per C `:1087` and `:1243`).
- `MEDIT_HITROLL`, `MEDIT_DAMROLL` (int with `uRange(-100, 100)`).
- `MEDIT_AC` (int with `uRange(-500, 500)`).
- `MEDIT_GOLD` (int ≥ 0).
- `MEDIT_LEVEL` (int `uRange(1, LEVEL_SUPREME)`; LEVEL_GREATER gate for PC target).
- `MEDIT_ALIGNMENT` (int `uRange(-1000, 1000)`).
- `MEDIT_FAVOR` (PC-only; int `uRange(-2500, 2500)` per C).
- `MEDIT_HITPOINT`, `MEDIT_MANA`, `MEDIT_MOVE` (int ≥ 0).
- `MEDIT_PRACTICE` (PC-only; int `uRange(1, 300)` — C omedit.c:1597 uses URANGE(1,atoi(arg),300); plan incorrectly stated (0,250)).
- `MEDIT_SILVER`, `MEDIT_COPPER` (int ≥ 0).
- `MEDIT_MENTALSTATE`, `MEDIT_EMOTIONAL` (int `uRange(-100, 100)`).
- `MEDIT_THIRST`, `MEDIT_FULL`, `MEDIT_DRUNK` (int `uRange(-1, 100)`).
- `MEDIT_DAMNUMDIE`, `MEDIT_DAMSIZEDIE`, `MEDIT_DAMPLUS` (NPC-only; int `uRange(0, 100)`).
- `MEDIT_HITNUMDIE`, `MEDIT_HITSIZEDIE`, `MEDIT_HITPLUS` (NPC-only; int).
- `MEDIT_POS`, `MEDIT_DEFAULT_POS` (position enum picker; DEFAULT_POS NPC-only).
- `MEDIT_SEX` (0/1/2 picker).
- `MEDIT_ATTACK`, `MEDIT_DEFENSE` (NPC-only; index into `attack_table` / defense flags).
- `MEDIT_SPEC` (NPC-only; spec_fun name lookup).
- `MEDIT_CLAN`, `MEDIT_DEITY`, `MEDIT_COUNCIL` (name lookup; empty clears).

**Tests:** one positive + one negative test per arm (valid value applies; OOB value rejects + keeps old). +1 ownership test per NPC-only / PC-only arm verifying rejection on wrong target.

**Acceptance:** A14 (all 25+ simple arms present + tested), A15 (NPC-only / PC-only rejection messages match C).

### G8 — Stat editors (STR/INT/WIS/DEX/CON/CHA/LUCK)

**Deliverables:** 7 parallel arms. All share a `meditArmStat(d, victim, arg, field *int)` helper. C `medit_parse` sets `minattr=1,maxattr=25` for NPC and `minattr=3,maxattr=18` for PC (`omedit.c:998-1007`). [C AUDIT: plan said uRange(3,25) for all; correct is NPC=uRange(1,25), PC=uRange(3,18). The helper must branch on IsNpc or accept min/max parameters.]

**Tests:** 7 positive tests (set 15, verify field), 7 clamp tests (set 99, verify clamped or rejected), 1 shared table-driven test per helper.

**Acceptance:** A16 (7 stat arms present, all clamp to URange, all match C range).

### G9 — Flag editors (bitmask shared helper + 6 wires)

**Deliverables:**

- `internal/game/olc_bitmask.go`: `olcBitmaskEdit(d, tableName, current, arg)` helper — parses `done`/`quit`/empty/number/keyword; toggles bit in `*current`; renders current mask state + remaining-bits menu.
- Lookup-table dispatch: `lookupBitmaskTable(tableName string) []string` returns the flag-name slice for ACT_FLAGS / PLR_FLAGS / AFF_FLAGS / PCFLAG / PART / RIS.
- Wire-ups: `meditArmNpcFlags` (ACT_*), `meditArmPcFlags` (PLR_*), `meditArmAffFlags` (AFF_*), `meditArmPcdataFlags` (PCFLAG_*), `meditArmParts` (PART_*), `meditArmResistant`/`meditArmImmune`/`meditArmSuscept` (RIS_* — three wires, one table).

**Tests:** table-driven toggle tests per table; `done` returns to correct main menu; numeric out-of-range rejects; keyword lookup case-insensitive.

**Acceptance:** A17 (shared helper present), A18 (6+ wires toggle correct bits), A19 (`done` / `quit` / empty semantics match across all wires).

### G10 — Affect editor (full, not placeholder)

**Deliverables:**

- `MEDIT_AFFECT_MENU` arm: parses A/R/Q. A→AFFECT_LOCATION. R→AFFECT_REMOVE. Q→main menu.
- `MEDIT_AFFECT_LOCATION` arm: APPLY_* picker (skip 0 + APPLY_EXT_AFFECT).
- `MEDIT_AFFECT_MODIFIER` arm: numeric for scalar APPLY; bitmask-helper dispatch for APPLY_AFFECT / APPLY_RESISTANT / APPLY_IMMUNE / APPLY_SUSCEPT.
- `MEDIT_AFFECT_REMOVE` arm: lists current affects with indices; builder picks number to remove.
- `handler.AffectModify` (or direct affect-list mutation; reuse oedit G10 primitive).

**Tests:** end-to-end add-STR+2 flow; add-APPLY_AFFECT with bitmask (e.g. AFF_INVISIBLE); remove-by-index.

**Acceptance:** A20 (affect-add scalar path works), A21 (affect-add bitmask path works), A22 (affect-remove path works).

### G11 — Save editor

**Deliverables:**

- `MEDIT_SAVE_MENU` arm: 5-way dispatch to SAV1..SAV5.
- `meditArmSav` shared helper (idx int): parses int, URange clamp, assigns the named Go field (`SavingPoisonDeath`/`SavingWand`/`SavingParaPetri`/`SavingBreath`/`SavingSpellStaff` on `CharData` — Go has no indexed SavingThrow array; C AUDIT: `victim->saving_poison_death` etc. per omedit.c:1626-1631; Go fields live on CharData not PCData), returns to SAVE_MENU.
- PC-only enforcement (reject on NPC).

**Tests:** 5 positive tests; NPC-reject test; clamp test.

**Acceptance:** A23 (all 5 save slots editable through menu).

### G12 — Class / Race editors

**Deliverables:**

- `MEDIT_CLASS` arm: numeric OR name-lookup via `util.ClassLookup` (if exists; TODO follow-up if not).
- `MEDIT_RACE` arm: parallel shape.
- LEVEL_GREATER trust gate.
- PC-only enforcement.

**Tests:** set class=0 by number; set class="warrior" by name (conditional on ClassLookup presence); LEVEL_GREATER gate test; NPC-reject test.

**Acceptance:** A24 (class/race editable), A25 (trust gate enforced).

### G13 — Password editor (security-critical)

**Deliverables:**

- `meditArmPassword(d, victim, arg)` per §Go Design sketch.
- Hash primitive CONFIRMED: `bcrypt.GenerateFromPassword([]byte(arg), act.BcryptCost)` at `internal/act/playercfg.go:260`. No grep needed; function is known. [C AUDIT: util.CryptPassword does not exist anywhere in the project]
- `util.SmashTilde` pre-filter before hashing.
- Empty-after-smash rejection.
- PC-only enforcement.
- LEVEL_GREATER OR self-edit check (builder may reset own password at any trust; editing another PC's password requires LEVEL_GREATER).

**Tests:**

- `TestMeditArmPassword_HashesPlaintext` — set password to `"plaintext"`; assert `victim.PcData.Pwd != "plaintext"` AND `len(victim.PcData.Pwd) > 0`.
- `TestMeditArmPassword_RoundTrips` — after set, verify `util.ComparePassword(victim.PcData.Pwd, "plaintext")` succeeds (exact compare-function name grepped at G13 start).
- `TestMeditArmPassword_EmptyRejected` — empty arg stays in MEDIT_PASSWORD; no mutation.
- `TestMeditArmPassword_TildeStripped` — input `"abc~def"` hashed as `"abcdef"` (or rejected if total-empty after smash).
- `TestMeditArmPassword_NpcRejected` — NPC target rejected with message.
- `TestMeditArmPassword_OtherPcRequiresGreater` — LEVEL_IMMORTAL editing another PC rejects; LEVEL_GREATER accepts.

**Acceptance:** A26 (plaintext never stored), A27 (round-trip compare succeeds), A28 (empty + tilde-strip semantics correct), A29 (NPC + trust gates enforced).

### G14 — DoMedit menu-entry + oedit G9 back-wire

**Deliverables:**

- `DoMedit` in `internal/act/olc.go` (or olc_interactive.go) — extend no-arg path: allocates `ch.Desc.Olc`, sets `Connected=CON_MEDIT`, stashes victim pointer in `Olc.Victim`, calls `MeditDispMenuFunc`. Flat subcommand path untouched (if it exists).
- Victim resolution: argument may be vnum (NPC prototype) OR PC name (connected descriptor). Same policy as C `do_omedit`.
- Double-edit guard: reject if `Connected == CON_MEDIT && Olc.Victim == target`.
- **Back-wire oedit G9 placeholder** (HARD DELIVERABLE — NOT a follow-up):
  - `internal/game/oedit_extras.go` `OEDIT_AFFECT_MODIFIER` arm currently accepts numeric only. Rewire to call `olcBitmaskEdit` for APPLY_AFFECT / APPLY_RESISTANT / APPLY_IMMUNE / APPLY_SUSCEPT.
  - Re-evaluate `APPLY_EXT_AFFECT` reject branch. Keep if C parity; remove with inline comment if C allows it.
  - Pin-test: `TestOeditParse_AffectModifier_UsesBitmaskEditor` — sets APPLY_AFFECT via oedit, toggles AFF_INVISIBLE via the new bitmask UI, verifies `obj.Affect[].Modifier` bit set.

**Tests:** `TestDoMedit_NoArgEntersMenuNpc`, `TestDoMedit_NoArgEntersMenuPc`, `TestDoMedit_WithArgKeepsFlatPath` (if flat path exists), `TestDoMedit_DoubleEditGuard`, plus the oedit back-wire test above.

**Acceptance:** A30 (menu-entry works for NPC vnum arg), A31 (menu-entry works for PC name arg), A32 (oedit bitmask back-wire landed + tested).

### G15 — Boot / CHANGELOG / CLAUDE.md / E2E

**Deliverables:**

- `internal/boot/boot.go`: wire `act.MeditDispMenuFunc = game.MeditDispMenu`.
- `internal/testclient/medit_test.go`: 3 E2E scenarios — (1) NPC menu entry + set-name + quit; (2) PC menu entry + set-stat + quit; (3) invalid-digit-redisplays.
- `CHANGELOG.md` entry.
- `CLAUDE.md` LANDED row update with commit hash.
- `TODO.md` housekeeping if ClassLookup/RaceLookup missing.

**Acceptance:** A33 (boot wire present + test), A34 (3 E2E scenarios green), A35 (docs updated).

---

## Acceptance Criteria (binary)

| # | Criterion | Gate (how to verify) |
|---|---|---|
| A1 | 64 `MEDIT_*` iota constants contiguous starting at 300 | `TestOlcMeditIotaContiguous` passes |
| A2 | `loop.go` dispatches `CON_MEDIT` to `meditParse` | grep for `case types.CON_MEDIT` |
| A3 | `meditParse` compiles and handles nil/empty defenses | `TestMeditParse_NilGuards` passes |
| A4 | `MeditDispMenuFunc` seam wired in boot | grep `boot.go` for assignment |
| A5 | NPC main menu renders ≥20 field labels | `TestMeditDispNpcMenu_ContainsFields` passes |
| A6 | PC main menu renders ≥20 field labels | `TestMeditDispPcMenu_ContainsFields` passes |
| A7 | Main-menu dispatcher routes by `IS_NPC` | `TestMeditDispMenu_RoutesByIsNpc` passes |
| A8 | All 16 submenu renderers present + snapshot-tested | Per-renderer test set passes |
| A9 | All 64 modes have arm + no default-log path fires | `TestMeditParse_EveryModeHasArm` passes |
| A10 | `Q` from any sub-mode returns to main; `Q` from main transitions to `CON_PLAYING` | `TestMeditParse_QuitReturnsToPlaying` passes |
| A11 | Every NPC-menu digit maps to documented mode | table-driven test passes |
| A12 | Every PC-menu digit maps to documented mode | table-driven test passes |
| A13 | LEVEL_GREATER gate enforced on LEVEL/CLASS/RACE for PC | trust-gate tests pass |
| A14 | 25+ simple-field arms apply values + pass URange clamps | per-arm tests pass |
| A15 | NPC-only / PC-only rejection messages match C verbatim | rejection-message tests pass |
| A16 | 7 stat arms (STR..LUCK) clamp to C range: NPC uRange(1,25) PC uRange(3,18) per omedit.c:998-1007 | stat tests pass [C AUDIT: original A16 had wrong bounds] |
| A17 | Shared `olcBitmaskEdit` helper present | `TestOlcBitmaskEdit_ShapeMatch` passes |
| A18 | 6+ bitmask wires toggle correct bits | per-wire toggle tests pass |
| A19 | `done`/`quit`/empty semantics consistent across bitmask wires | shared-semantics test passes |
| A20 | Affect-add scalar path functional | affect-add scalar test passes |
| A21 | Affect-add bitmask path functional (APPLY_AFFECT etc.) | affect-add bitmask test passes |
| A22 | Affect-remove by index functional | affect-remove test passes |
| A23 | All 5 save slots (SAV1..SAV5) editable through menu | save-menu test set passes |
| A24 | Class / Race editable (numeric + name if lookup exists) | class/race tests pass |
| A25 | LEVEL_GREATER gate enforced on class/race | trust-gate tests pass |
| A26 | MEDIT_PASSWORD never stores plaintext in `PcData.Pwd` | `TestMeditArmPassword_HashesPlaintext` passes |
| A27 | Password round-trips through verify path | `TestMeditArmPassword_RoundTrips` passes |
| A28 | Empty-arg + tilde-strip semantics correct | empty + tilde tests pass |
| A29 | Password rejects NPC + enforces trust for other-PC | npc + trust tests pass |
| A30 | `DoMedit` no-arg on NPC vnum enters menu | `TestDoMedit_NoArgEntersMenuNpc` passes |
| A31 | `DoMedit` no-arg on PC name enters menu | `TestDoMedit_NoArgEntersMenuPc` passes |
| A32 | oedit `OEDIT_AFFECT_MODIFIER` placeholder replaced with bitmask UI | `TestOeditParse_AffectModifier_UsesBitmaskEditor` passes |
| A33 | Boot wire present + idempotent test | boot test passes |
| A34 | 3 E2E testclient scenarios green | testclient test suite passes |
| A35 | CHANGELOG + CLAUDE.md LANDED row updated | manual verification |

---

## Mutation Gates (≥15, all Edit-round-trip)

All mutations are `Edit`-round-trip only — banned-command list respected (no `git checkout`/`git restore`/`git stash`/`git reset --hard`).

| # | Mutation | Pinning test that MUST fail |
|---|---|---|
| M1 | Remove `case types.CON_MEDIT` from `loop.go` dispatch | `TestLoopDispatch_ConMedit_RoutesToMeditParse` |
| M2 | Flip `iota + 300` to `iota + 200` | `TestOlcMeditIotaContiguous` |
| M3 | Drop NPC-only reject branch in `meditArmSpec` | `TestMeditArmSpec_RejectedOnPc` |
| M4 | Drop PC-only reject branch in `meditArmPassword` | `TestMeditArmPassword_NpcRejected` |
| M5 | Replace `bcrypt.GenerateFromPassword([]byte(arg), act.BcryptCost)` with `arg` directly (plaintext store) [C AUDIT: M5 previously referenced non-existent util.CryptPassword; actual call is bcrypt.GenerateFromPassword] | `TestMeditArmPassword_HashesPlaintext` |
| M6 | Drop `util.SmashTilde` pre-filter in MEDIT_PASSWORD | `TestMeditArmPassword_TildeStripped` |
| M7 | Flip PC STR clamp from `uRange(3, 18)` to `uRange(0, 100)` | `TestMeditArmStr_PcClampTo18` [C AUDIT: PC max is 18 not 25; NPC max is 25 with min=1; plan had wrong clamp] |
| M8 | Drop LEVEL_GREATER gate in MEDIT_CLASS | `TestMeditArmClass_RequiresGreater` |
| M9 | Flip bitmask `done` string match from exact to prefix (e.g. `"d"` matches) | `TestOlcBitmaskEdit_DoneExactMatch` |
| M10 | Drop MEDIT_AFFECT_MODIFIER bitmask branch (revert to numeric-only) | `TestOeditParse_AffectModifier_UsesBitmaskEditor` (via back-wire) |
| M11 | Drop cleanupOlc Victim clear | `TestCleanupOlc_ClearsVictim` |
| M12 | Swap MeditDispNpcMenu and MeditDispPcMenu routing predicate | `TestMeditDispMenu_RoutesByIsNpc` |
| M13 | Drop SAVE_MENU PC-only reject | `TestMeditArmSavMenu_NpcRejected` |
| M14 | Drop MEDIT_LEVEL LEVEL_GREATER gate on PC target | `TestMeditArmLevel_PcRequiresGreater` |
| M15 | Swap MEDIT_SEX picker to accept 3 (out-of-range) | `TestMeditArmSex_RejectsThree` |
| M16 | Drop double-edit guard in `DoMedit` | `TestDoMedit_DoubleEditGuard` |

Each mutation is applied via `Edit old_string → new_string`, tested (fails), then reverted via a second `Edit new_string → old_string`. The reverse edit is the canonical banned-command workaround.

---

## Scope Cuts

Kept OUT of this plan to bound execution:

- **Mob-program editing** — MEDIT_MPROGS* constants do NOT exist in `src/olc.h:198-263`. C mob-mprog editing lives under `CON_MPROG_EDIT` substate; port in `plan-phase6-olc-mpedit.md`.
- **`MEDIT_CONFIRM_SAVESTRING` (=9)** — **LIVE in C** (`omedit.c:1011-1057`), unlike oedit where it's dead. PC Q-exit enters this mode with prompt "Do you wish to save to disk? : "; Y→`save_char_obj` or `fold_area`; N→`cleanupOlc`; invalid→re-prompt. NPC Q skips the confirm (direct `cleanupOlc`). Plan was initially wrong to call this reserved-but-not-dispatched (audit correction 2026-04-19). **G-start decision**: port verbatim (simplest, matches C), OR land without it and let the global autosave tick handle persistence (Go divergence — simpler UX but drift from C). **Recommended: port verbatim** so builders get the familiar prompt.
- **PC self-edit policy** — C `do_omedit` at `:109-232` (to verify) has an explicit PC-editing-PC ACL (typically LEVEL_GREATER for other PC, self-edit always allowed). Port the ACL but do NOT introduce a configurable policy switch.
- **APPLY_EXT_AFFECT** — If C skips it (oedit parity), skip. Defer any decision to enable it to mpedit or a later plan.
- **SOGI gender migration** — `post-phase6-vision.md` captures direction to modernize `Sex` + pronouns. MEDIT_SEX ports the stock 0/1/2 enum VERBATIM. Migration is a separate plan.
- **Password-change audit log** — G13 does not emit an `olcLog` entry containing the old or new hash. Logging that password was changed (no hash value) is acceptable.
- **Online OLC persistence** — A changed mob prototype flushes to disk on the next `asave` tick. MEDIT does NOT force-save on `Q`. PC changes persist through normal save paths.
- **Race-lookup + Class-lookup by name** — if `util.ClassLookup` / `util.RaceLookup` absent, numeric-only with TODO follow-up. NOT a blocker.
- **Reverse-field rendering (short-form current-value in main menu)** — C renders short current values for every field in the main menu. Go port may render a subset at G2 and expand in a follow-up; NOT a blocker (A5/A6 threshold is ≥20 labels).
- **Mass-mob operations** (`mset all <flag>`, `mset race <name>`) — those are `mset` concerns, not `medit` menu concerns.

---

## Open Questions

| # | Question | Resolution strategy |
|---|---|---|
| Q1 | Does a flat `DoMedit` / `DoMset` already exist in Go? | **RESOLVED** 2026-04-19 grep: `DoMset` lives at `internal/act/olc_set.go:14`. `DoMedit` does NOT exist — menu-entry path is the first medit port. G14 extends `DoMedit` as a NEW command (registered alongside DoMset). Flat subcommand path NOT required (no existing flat `medit` — builders use `mset`). |
| Q2 | Exact name of password-hash primitive? | **RESOLVED** 2026-04-19 grep: `bcrypt.GenerateFromPassword([]byte(pwd), act.BcryptCost)`. Uses `golang.org/x/crypto/bcrypt` + project-defined `act.BcryptCost`. See `internal/act/playercfg.go:260`, `internal/game/loop.go:415,499`. G13 uses bcrypt for Go-port parity (diverges from C's sha256_crypt — documented in §Password path). |
| Q3 | Does `util.ClassLookup` / `util.RaceLookup` exist? | **RESOLVED** 2026-04-19 grep: neither symbol exists in `internal/util/` or `internal/world/`. G12 ships numeric-only with TODO.md follow-up entries for both lookups. |
| Q4 | Which field does D_DESC (death description) target on CharData — `DeathCry`, `DeathMsg`, or a new field? | grep `internal/types/char.go` at G7 start. Likely exists per C `fread_mobile` parity. |
| Q5 | Does C `medit_disp_npc_menu` render digit 1 as Name or something else? | verify at G5 start by reading `src/omedit.c` main-menu render function. |
| Q6 | Does C enforce LEVEL_GREATER on LEVEL (not just CLASS/RACE)? | verify at G7 MEDIT_LEVEL arm. |
| Q7 | Does C allow editing another connected PC, or only prototype + self? | verify at G14 `do_omedit` entry-point port. Policy: port C verbatim; document divergence if any. |
| Q8 | Does `handler.AffectModify` support direct-mutation on mobs (vs oedit objects)? | verify at G10 start; reuse if so, fork if not. |
| Q9 | Do we need a separate "edit PC" LEVEL cap (e.g. cannot edit LEVEL_IMPLEMENTOR victim)? | verify at G1 `can_mmodify` port. |
| Q10 | Does stat clamp go up to 25 or higher? | verify at G8 against `src/omedit.c` MEDIT_STR arm and `src/mud.h` MAX_STAT. |
| Q11 | Does MEDIT_FAVOR apply to NPC or PC-only? | verify at G7 — C omedit likely PC-only since favor is deity-tracker in pcdata. |
| Q12 | Does `victim.PcData` nil-check need to be pervasive in PC-only arms, or does the top-level dispatcher guarantee non-nil? | verify at G4 dispatcher shape — if PC-only arms are gated by `!IS_NPC` + dispatcher guarantees `PcData != nil` for non-NPC, simplify. |

---

## C Bug Catalog (port verbatim unless fix documented)

Identified 2026-04-19 from C read. Additional bugs may surface during per-G execution; this catalog is the minimum verified set.

1. **`MEDIT_D_DESC` parse-arm is an unreachable bug-trap** (`omedit.c:1431-1435`). The editor uses `substate = SUB_MOB_DESC` + `last_cmd = do_medit_reset` path; if a line ever arrives at the MEDIT_D_DESC case, C calls `cleanup_olc(d); bug("OLC: medit_parse(): Reached D_DESC case!", 0);`. **Go port:** preserve — if the arm fires, call `cleanupOlc(d)` + `util.Bug("medit_parse: reached D_DESC case")`. Pin-test `TestMeditParse_DDescArmCallsBugAndCleanup`.
2. **NAME on PC triggers `do_pcrename`** (`omedit.c:1390-1395`) instead of direct `name =` assignment. Trust gate is `get_trust > LEVEL_SUB_IMPLEM - 1` (i.e. >= LEVEL_SUB_IMPLEM). **Go port:** if `DoPcrename` doesn't exist, fall back to direct name assignment with a TODO.md entry. If it exists, invoke it to match C semantics (rename also renames the pfile on disk). Pin with test.
3. **Post-hash tilde check on password** (`omedit.c:1610-1616`). C checks `~` in the sha256 output, not the raw input. On sha256 the check is effectively-unreachable; on bcrypt it is entirely unreachable (bcrypt base64 has no `~`). **Go port:** move the tilde check to raw input (safer). Documented in §Password path. This is the ONE C-divergence in G13; pin-tested.
4. **NPC Q skips save-confirm; PC Q triggers it** (`omedit.c:1062-1064` vs `:1222-1229`). No OLC_CHANGE gate on NPC exit. **Go port:** same asymmetry — NPC Q calls `cleanupOlc`; PC Q enters MEDIT_CONFIRM_SAVESTRING if `d.Olc.Dirty` (Go equivalent of `OLC_CHANGE`).
5. **`PC digit 6` and `PC digit U` explicit "NPC Only!!" rejects** (`omedit.c:1255-1257`, `:1351-1353`). Dead slots reserved in the digit grid; builder-training UX confirmation. **Go port:** match verbatim with those exact strings in the PC-main-menu dispatch.
6. **`PC digit Z` CLAN trust-gated to `LEVEL_GOD`, `PC digit =` COUNCIL trust-gated to `LEVEL_SUB_IMPLEM`** (`omedit.c:1370-1381`). Divergent gates within the same menu. **Go port:** match verbatim.
7. **Default-case in both menus redisplays `medit_disp_npc_menu`, not PC menu** (`omedit.c:1212-1214` NPC, `:1382-1384` PC). The PC-menu default calls the NPC-menu renderer — a likely C bug. **Go port:** fix this — PC-menu default redisplays PC menu. Document divergence.
8. **Stat clamps NOT visible in the digit-dispatch** — clamps live in the per-mode parse arms. Must cite at G8 start by reading the MEDIT_STRENGTH arm (likely `URANGE(0, atoi(arg), 25)` per `omedit.c:1625` precedent for SAV1).
9. **SAV slot store pattern** (`omedit.c:1625-1631`). Each SAV arm stores into `victim->saving_*` (named field, not indexed array), then also writes to `pIndexData->saving_*` if `IS_NPC(victim) && ACT_PROTOTYPE`. **Go port:** use named `CharData` fields (`SavingPoisonDeath`, `SavingWand`, `SavingParaPetri`, `SavingBreath`, `SavingSpellStaff` — verified at `internal/types/character.go:141-145`; NOT on PCData and NOT an indexed array); NPC proto write only if `victim.IsNPC()` AND ACT_PROTOTYPE flag.

Additional bugs will be catalogued during per-G execution.

---

## Risk Analysis

- **High risk:** MEDIT_PASSWORD. Plaintext leak is a security CVE-class bug. §G13 pin-tests are the mitigation. Secondary mitigation: adversary review called out explicitly for G13.
- **Medium risk:** PC-vs-NPC dispatcher correctness. 64 modes + ownership table is combinatorial. Mitigation: table-driven ownership test at G4 enumerates every mode × target-type pair.
- **Medium risk:** oedit back-wire (G14). If the bitmask helper shape differs from what oedit G9 placeholdered, back-wire may require refactor. Mitigation: G9 designs the helper signature to be drop-in compatible; G14 is additive.
- **Low risk:** Menu-text snapshot drift. If C output differs from our port by ANSI code sequence or spacing, tests may need fuzzy-match. Mitigation: precedent from redit/oedit (regex-scan for label substrings, NOT byte-exact).
- **Low risk:** EditorSave trampoline. Well-established contract; same as redit + oedit.
- **Low risk:** iota collision. `iota+300` is outside the ranges used by REDIT (100) and OEDIT (200). Pinned by A1.

---

## Test Plan

**Unit tests** (per G):

| Test file | G | ~LOC |
|---|---|---|
| `internal/types/olc_test.go` (extend) | G1 | +30 |
| `internal/game/medit_parse_test.go` (NEW) | G4, G7-G13 | ~1500 |
| `internal/game/medit_menu_test.go` (NEW) | G2, G3 | ~400 |
| `internal/game/olc_bitmask_test.go` (NEW) | G9 | ~300 |
| `internal/act/olc_test.go` (extend) | G14 | +100 |
| `internal/testclient/medit_test.go` (NEW) | G15 | ~150 |
| `internal/boot/boot_test.go` (extend) | G15 | +20 |

Estimated total: ~2500 new test LOC, ~150 tests.

**Integration tests:** 3 E2E testclient scenarios in `internal/testclient/medit_test.go`:

1. `TestTestclient_MeditNpcMenuEntryAndQuit` — dig into menu, quit, verify CON_PLAYING.
2. `TestTestclient_MeditPcSetStat` — enter PC menu, set STR=18, verify mutation persists.
3. `TestTestclient_MeditInvalidChoiceRedisplays` — type unrecognized digit, verify menu redisplays.

**Mutation verification:** 16 gates per §Mutation Gates. Each applied via `Edit` round-trip only.

**Regression:** `go test -count=3 ./...` green across all 15 packages post-G15.

---

## File Budget

New files:

| Path | Est. LOC |
|---|---|
| `internal/game/medit_parse.go` | ~1400 |
| `internal/game/medit_menu.go` | ~500 |
| `internal/game/olc_bitmask.go` | ~250 |
| `internal/game/medit_parse_test.go` | ~1500 |
| `internal/game/medit_menu_test.go` | ~400 |
| `internal/game/olc_bitmask_test.go` | ~300 |
| `internal/testclient/medit_test.go` | ~150 |

Modified files:

| Path | Est. Δ LOC |
|---|---|
| `internal/types/olc.go` | +80 (64 consts + Victim field) |
| `internal/game/loop.go` | +4 (dispatch arm) |
| `internal/game/oedit_extras.go` | ±30 (back-wire G14) |
| `internal/act/olc.go` | +30 (seam + DoMedit extension) |
| `internal/boot/boot.go` | +3 (wire) |
| `internal/types/olc_test.go` | +30 |
| `internal/act/olc_test.go` | +100 |
| `internal/boot/boot_test.go` | +20 |
| `CLAUDE.md` | +60 (LANDED row) |
| `CHANGELOG.md` | +10 |
| `TODO.md` | +10 (follow-ups) |

**Total:** ~4500 new LOC (code + tests), ~350 modified LOC.

---

## Cross-References

- `smaug-go/doc/plan-phase6-olc-redit.md` — §Cross-Plan Dependencies, iota convention, EditorSave trampoline pattern.
- `smaug-go/doc/plan-phase6-olc-oedit.md` — §Bitmask helper placeholder + affect editor shape + back-wire target (G14).
- `smaug-go/doc/plan-phase5-tier12-editor-save.md` — EditorSave callback contract.
- `smaug-go/doc/post-phase6-vision.md` — SOGI-gender migration (MEDIT_SEX ports stock 0/1/2 verbatim; modernization is a separate plan).
- `src/omedit.c` (2280 LOC) — authoritative C reference.
- `src/olc.h:198-263` — MEDIT_* mode constants.
- `src/smaug.c` (~:1641-1652) — C dispatch call site (same switch as CON_REDIT/CON_OEDIT).

---

## Final Note

This plan is structured to match the redit/oedit precedent exactly. Execution should begin at G1 (schema + seams + loop arm) as an atomic first wave; G2-G4 as a second wave (renderers + parse skeleton); G5-G13 as the bulk field-port wave; G14 as the back-wire + menu-entry wave; G15 as the landing wave (boot + E2E + docs). Adversary review is called out explicitly for G13 (password security) and G14 (oedit back-wire correctness).

**Blockers for execution** (post-authoring 2026-04-19 verification):

- **RESOLVED** Q1: `DoMset` exists at `internal/act/olc_set.go:14`; `DoMedit` does NOT — G14 adds as a new command.
- **RESOLVED** Q2: bcrypt via `golang.org/x/crypto/bcrypt` + `act.BcryptCost`. See §Password path.
- **RESOLVED** Q3: `ClassLookup`/`RaceLookup` absent — G12 numeric-only + TODO.md follow-ups.
- **RESOLVED** Q4: D_DESC target is `victim.Description` per C `:1087` and `:1243`.
- **VERIFY AT G-START** Q5-Q12: menu digits verified above in the PC-vs-NPC table; remaining Qs are per-arm clamps and field presence — verify-as-you-go.
- **NO HARD BLOCKERS** — plan is executable as-is.

No cross-plan blockers outstanding. oedit G9 bitmask is back-wired here (additive); no shared persistence schema (PC mutations persist through existing save path via `SaveFunc`; NPC prototype persists through existing `asave`).

---

**Plan authored:** 2026-04-19.
**Line count:** ~980.
**Task groups:** G1-G15.
**Acceptance criteria:** A1-A35.
**Mutation gates:** M1-M16.
**C citations verified:** `do_omedit` @ :109-232, `medit_disp_menu` @ :814-825, `medit_parse` @ :985-2160, `medit_disp_aff_flags` @ :637-812, `medit_disp_ris` @ :502-533, `MEDIT_PASSWORD` arm @ :1601-1623, NPC main menu dispatch @ :1059-1215, PC main menu dispatch @ :1218-1386.
**Pending:** adversary review before G1 dispatch.
