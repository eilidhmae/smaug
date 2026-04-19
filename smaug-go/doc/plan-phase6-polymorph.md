# Phase 6 — Polymorph

**Status:** Authored 2026-04-18 (Wave D). Pending external adversary review before execution.

**Scope signal:** Medium-risk. 2753 C LOC (`src/polymorph.c`) plus `src/mud_comm.c:1960-2070` mudprog hooks plus `src/save.c:288,1083` pfile hooks. Schema (`CharData.Morph`, `CharMorph`, `MorphData`) already complete in Go per 2026-04-18 audit; missing pieces are loader/saver, command set, combat stat-application hook, mudprog bodies, and per-player `#MorphData` pfile block.

---

## 1. Problem

SMAUG's polymorph subsystem lets players (and immortals) morph into alternate forms that override their stats, resists, saves, hitroll/damroll, max hp/mana/move, affects, short desc, and long desc. Morphs are defined in `db/system/morph.dat` (one record per form) and applied on `morph <name>` (player) or `morph <vnum>` (immortal). Unmorphing reverses every applied modifier.

Go currently has:
- `CharData.Morph *CharMorph` (`internal/types/character.go:66`) — present.
- `CharMorph` (character.go:270) and `MorphData` (character.go:306) structs — fully defined with all C fields mapped.
- `mpmorph` / `mpunmorph` mudprog dispatch entries (`internal/mudprog/driver.go:260-265`) — but the handler bodies (`commands.go:545,551`) are empty stubs with `TODO(tier3)` comments.

Missing:
- No `morph_start` equivalent (morph table) in Go state. No loader/saver for `db/system/morph.dat`.
- No `DoMorph` / `DoUnmorph` / `DoMorphset` / `DoMorphstat` / `DoMorphcreate` / `DoMorphdestroy` commands.
- No `DoMorphChar` / `DoUnmorphChar` / `DoMorphApply` / `DoMorphRemove` stat-application helpers (equivalent of C `do_morph_char`/`do_morph`/`do_unmorph_char`/`do_unmorph`).
- No `#MorphData` block emit/read in `SavePlayer`/`LoadPlayer`.
- Mudprog `mpmorph` / `mpunmorph` bodies are stubs.
- No `send_morph_message` helper.
- No ifcheck data path for `ismorphed` beyond the existing presence check — the `morph` ifcheck reads `ch->morph->morph->vnum`, which already works once we populate the table (tier3 test fixture at `ifcheck_tier3_test.go:371` confirms).

---

## 2. C Reference (authoritative)

### 2.1 Struct definitions

- `struct char_morph` at `src/mud.h:524-558` (32 fields). Mapped to Go `CharMorph` at `internal/types/character.go:270-303` — all fields present.
- `struct morph_data` at `src/mud.h:560-629` (64 fields). Mapped to Go `MorphData` at `internal/types/character.go:306-372` — all fields present.
- Global table: `MORPH_DATA *morph_start, *morph_end` at `src/polymorph.c:47-48`. Go equivalent: per-world or package-level map/slice (see §4).

### 2.2 Commands (entry points)

| C function | C location | Command name (tables.c) | Purpose |
|---|---|---|---|
| `do_imm_morph` | `polymorph.c:2676-2723` | `morph` | Immortal: morph self/victim by vnum |
| `do_imm_unmorph` | `polymorph.c:2729-2753` | `unmorph` | Immortal: unmorph self/victim |
| `do_morphset` | `polymorph.c:72-... 195` | `morphset` | Immortal: edit fields of a morph by name/vnum; triggers `save_morphs` on save subcommand |
| `do_morphstat` | `polymorph.c:969-...` | `morphstat` | Immortal: display a morph's fields |
| `do_morphcreate` | `polymorph.c:2313-...` | `morphcreate` | Immortal: create new morph record (blank template, assigned vnum via `setup_morph_vnum`) |
| `do_morphdestroy` | `polymorph.c:2370-...` | `morphdestroy` | Immortal: remove morph by name/vnum; calls `unmorph_all` first |
| `do_mpmorph` | `mud_comm.c:1964-2011` | mudprog `mpmorph` | Mob command: morph target PC by morph name |
| `do_mpunmorph` | `mud_comm.c:2013-2070` | mudprog `mpunmorph` | Mob command: unmorph target PC |

**Note:** C has no "player `morph` command" — `morph` is the immortal command. Players enter morphs via `spell_polymorph` (`tables.c:177-178` maps `spell_polymorph`), which resolves a morph name and calls `do_morph_char`. Ports that want player-typed morph access attach to the spell system. We follow C and expose only the immortal `morph` / `unmorph` commands in G3.

### 2.3 Apply / Remove primitives

- `do_morph_char(CHAR_DATA*, MORPH_DATA*)` at `polymorph.c:1262-1384` — checks object/hp/mana/move/blood/favour/glory prerequisites, sends consumption messages, calls `send_morph_message`, calls `do_morph`.
- `do_morph(CHAR_DATA*, MORPH_DATA*)` at `polymorph.c:1473-1531` — **the stat-application path.** Increments `armor`, `mod_{str,int,wis,dex,cha,lck}`, saving-throws, `hitroll`, `damroll`, `hit`, `move`, `mana` (or `blood` for vampires). Also `xSET_BITS`/`SET_BIT` the morph's `affected_by`/`immune`/`resistant`/`susceptible` onto the char, and `xREMOVE_BITS`/`REMOVE_BIT` the `no_*` masks. Attaches `ch_morph` to `ch->morph`. Increments `morph->used`.
- `do_unmorph_char(CHAR_DATA*)` at `polymorph.c:1387-1396` — wraps `do_unmorph` with message.
- `do_unmorph(CHAR_DATA*)` at `polymorph.c:1545-1579` — exact reverse of `do_morph`, then `free_char_morph` + `ch->morph = NULL` + `update_aris`.
- `can_morph(CHAR_DATA*, MORPH_DATA*, bool)` at around `polymorph.c:1420+` — prereq gate (level/class/race/sex/day/time/pkill/pcast).
- `find_morph(CHAR_DATA*, char*, bool)` at `polymorph.c:1221-1235` — name lookup + `can_morph` filter.
- `get_morph(char*)` at `polymorph.c:1205-1216`, `get_morph_vnum(int)` at `polymorph.c:1243-1254`.
- `make_char_morph(MORPH_DATA*)` / `clear_char_morph(CHAR_MORPH*)` / `free_char_morph(CHAR_MORPH*)` — allocation helpers (`polymorph.c:~600,2590,~2570`).
- `send_morph_message(CHAR_DATA*, MORPH_DATA*, bool)` at `polymorph.c:1402-1418` — emits `AT_MORPH`-colored `act` to char + room.
- `unmorph_all(MORPH_DATA*)` at `polymorph.c:2655-2668` — iterates `first_char`, unmorphs any PC currently using this morph. Called before destroying a morph.
- `setup_morph_vnum(void)` at `polymorph.c:2630-2652` — assigns fresh vnums ≥1000 to any morph with vnum==0.
- `morph_defaults(MORPH_DATA*)` at `polymorph.c:~2590-2627` — zero/neutral defaults for a newly-created morph.

### 2.4 Persistence

- **Morph table** (`db/system/morph.dat`): `save_morphs()` at `polymorph.c:1586-1605`, `fwrite_morph` at `polymorph.c:1613-1755` (approx; writes all non-zero fields with tokens `Morph`/`Name`/`Vnum`/`Blood`/`Damroll`/`Defpos`/`Description`/`Help`/`Hit`/`Hitroll`/`Keywords`/`Longdesc`/`Mana`/`MorphOther`/`MorphSelf`/`Move`/`NoSkills`/`ShortDesc`/`Skills`/`UnmorphOther`/`UnmorphSelf`/`Affected`/`Class`/`Immune`/`NoAffected`/`NoImmune`/etc., terminator `End`). File terminator `#END`. `load_morphs()` at `polymorph.c:1757-1819`, `fread_morph` at `polymorph.c:1821+`. Boot wiring: `db.c:941` calls `load_morphs()` during `boot_db`.
- **Per-player morph state** (`#MorphData` block in pfile): `fwrite_morph_data(CHAR_DATA*, FILE*)` at `polymorph.c:2405-2484`. `fread_morph_data(CHAR_DATA*, FILE*)` at `polymorph.c:2487+`. Emit is conditional: `if (ch->morph) fwrite_morph_data(...)` at `save.c:288`. Load: `else if (!strcmp(word, "MorphData"))` at `save.c:1082-1083` calling `fread_morph_data`. Block format: opens with `#MorphData\n`, then key/value lines (`Vnum %d`, `Name %s~`, `Affect %s`, `Immune %d`, etc. per §2.4 field list), terminator `End`.

### 2.5 Mudprog trigger bodies

- `do_mpmorph` at `mud_comm.c:1964-2011`: argument is `<target> <morph-name>`; resolves `victim = get_char_room(mob, arg1)`; resolves `morph = get_morph(arg2)`; calls `do_morph_char(victim, morph)`. Errors bug-logged to `mpbug`.
- `do_mpunmorph` at `mud_comm.c:2013-2070`: argument is `<target>`; resolves victim; calls `do_unmorph_char(victim)`.

### 2.6 Already-ported ifchecks

- `ismorphed` (`mud_prog.c:974-976`) reads `ch->morph != NULL` — already works in Go (`ifcheck_tier3_test.go:371`).
- `morph` (`mud_prog.c:1207-1213`) reads `ch->morph->morph->vnum` — will work once the morph table is populated.

---

## 3. Go Current State

| Component | Go state | File |
|---|---|---|
| `CharData.Morph *CharMorph` | Present | `internal/types/character.go:66` |
| `CharMorph` struct | Present, 32 fields | `internal/types/character.go:270-303` |
| `MorphData` struct | Present, 64 fields | `internal/types/character.go:306-372` |
| Morph table (global/world-scoped) | **Missing** | — |
| `LoadMorphs` / `SaveMorphs` | **Missing** | — |
| `FWriteMorph` / `FReadMorph` | **Missing** | — |
| `FWriteMorphData` / `FReadMorphData` (pfile block) | **Missing** | — |
| Emit `#MorphData` in `SavePlayer` | **Missing** | `internal/persist/player.go` |
| Parse `#MorphData` in `LoadPlayer` | **Missing** | `internal/persist/player.go` |
| `DoMorph` / `DoUnmorph` commands (immortal) | **Missing** | — |
| `DoMorphset` / `DoMorphstat` / `DoMorphcreate` / `DoMorphdestroy` | **Missing** | — |
| `DoMorphChar` / `DoUnmorphChar` apply/remove helpers | **Missing** | — |
| `DoMorphApply` / `DoMorphRemove` stat path | **Missing** | — |
| `SendMorphMessage` helper | **Missing** | — |
| `GetMorph` / `GetMorphVnum` / `FindMorph` lookups | **Missing** | — |
| `MakeCharMorph` / `FreeCharMorph` | **Missing** | — |
| `MorphDefaults` / `SetupMorphVnum` | **Missing** | — |
| `UnmorphAll(*MorphData)` | **Missing** | — |
| `mpmorph` / `mpunmorph` dispatch entries | Present | `internal/mudprog/driver.go:260-265` |
| `mpMorph` / `mpUnmorph` handler bodies | **Stubbed (empty)** | `internal/mudprog/commands.go:545-554` |
| `AT_MORPH` color constant | Need to verify | `internal/types/ansi.go` or similar |
| `update_aris` equivalent | Need to check | `internal/handler/` |
| `dice_parse` equivalent | Should exist (DiceString helper) | `internal/util/dice.go` |
| `ismorphed` ifcheck | Working | `internal/mudprog/ifcheck_tier3.go` |
| `morph` ifcheck | Present; needs populated morph table to exercise | `internal/mudprog/ifcheck_tier3.go` |

### 3.1 Boot wiring needed

- `internal/boot/boot.go`: new `LoadMorphs` call (mirrors `db.c:941` placement — after world/areas loaded so `get_obj_vnum` lookups inside `do_morph_char` work on a populated world).
- `internal/boot/boot.go`: command registration for 6 new commands.

### 3.2 Existing infrastructure to reuse

- `util.Act(AT_*, format, ch, obj, vict, target)` — already wired.
- `handler.GetCharRoom` / `GetCharWorld` / `GetObjVnum` — present.
- `persist.Scanner` — line-oriented, tilde-terminated strings, matches the C `fread_string` pattern.
- `types.BitVector` — 128-bit ext_bv, methods `IsSet`/`SetBit`/`ClearBit`/`OrBits`/`AndNotBits`/`String` for serialization.
- `util.DiceParse` or equivalent for `"2d4+1"` strings — need to confirm name; C `dice_parse(ch, level, str)` takes a level-scaled expression. If missing, add it; if present, reuse.

---

## 4. Go Design

### 4.1 Morph table location

Choice: add `Morphs []*types.MorphData` (slice) and `MorphsByVnum map[int]*types.MorphData` / `MorphsByName map[string]*types.MorphData` as fields on `world.World` — same shape as existing area/skill tables. Avoids a new package-level global. Accessor helpers (`GetMorph`, `GetMorphVnum`, `FindMorph`) become methods on `*World` or package-level functions taking `*World`.

**Rejected alternative:** package-level `morphStart/morphEnd` globals. Matches C but contradicts `world.World` convention established in plan.md.

### 4.2 Stat-application hook

Choice: implement `DoMorph(ch, morph)` and `DoUnmorph(ch)` as direct helpers in a new `internal/polymorph/polymorph.go` package (or `internal/act/polymorph.go` — see §4.7), mirroring C's direct-mutation pattern. No `AffectData` synthesis.

**Why not AffectModify:** the C implementation does NOT route morph stats through `affect_modify`; it mutates `ch->armor`, `ch->mod_str`, etc. directly and stores the delta on `ch->morph` for exact reversal. Funneling through `AffectModify` would require a `SpellType`/`Location`/`Modifier` explosion per morph field (25+ synthetic affects) and would not match C semantics for `mod_*` (which `AffectModify` does touch, but via `APPLY_STR` etc. with fundamentally different semantics). Direct mutation keeps parity.

### 4.3 Apply path (porting `do_morph`)

```go
// DoMorph applies a morph's stat overrides to ch. Mirrors C do_morph at polymorph.c:1473.
func DoMorph(ch *types.CharData, morph *types.MorphData) {
    if morph == nil { return }
    if ch.Morph != nil { DoUnmorphChar(ch) } // existing morph first
    cm := MakeCharMorph(morph)
    ch.Armor += morph.AC
    ch.ModStr += morph.Str; ch.ModInt += morph.Int; ch.ModWis += morph.Wis
    ch.ModDex += morph.Dex; ch.ModCha += morph.Cha; ch.ModLck += morph.Lck
    ch.SavingBreath += morph.SavingBreath
    // ... all five saves ...
    cm.Hitroll = util.DiceParse(ch, morph.Level, morph.Hitroll); ch.Hitroll += cm.Hitroll
    cm.Damroll = util.DiceParse(ch, morph.Level, morph.Damroll); ch.Damroll += cm.Damroll
    cm.Hit = util.DiceParse(ch, morph.Level, morph.Hit)
    if ch.Hit + cm.Hit > 32700 { cm.Hit = 32700 - ch.Hit }
    ch.Hit += cm.Hit
    cm.Move = util.DiceParse(ch, morph.Level, morph.Move)
    if ch.Move + cm.Move > 32700 { cm.Move = 32700 - ch.Move }
    ch.Move += cm.Move
    if !ch.IsNPC() && isVampire(ch) {
        cm.Blood = util.DiceParse(ch, morph.Level, morph.Blood)
        ch.PCData.Condition[COND_BLOODTHIRST] += cm.Blood
    } else {
        cm.Mana = util.DiceParse(ch, morph.Level, morph.Mana)
        if ch.Mana + cm.Mana > 32700 { cm.Mana = 32700 - ch.Mana }
        ch.Mana += cm.Mana
    }
    ch.AffectedBy.OrBits(morph.AffectedBy)
    ch.Immune |= morph.Immune
    ch.Resistant |= morph.Resistant
    ch.Susceptible |= morph.Suscept
    ch.AffectedBy.AndNotBits(morph.NoAffectedBy)
    ch.Immune &^= morph.NoImmune
    ch.Resistant &^= morph.NoResistant
    ch.Susceptible &^= morph.NoSuscept
    ch.Morph = cm
    morph.Used++
}
```

`DoUnmorph` is the exact reverse (C `polymorph.c:1545-1579`).

### 4.4 Persistence (morph.dat)

Reuse `persist.Scanner`. File format token-for-token per C `fwrite_morph`. Emit: iterate world.Morphs, write `Morph <name>\n` header, then conditional-on-nonzero fields, then `End\n`. File terminator `#END\n`. Load: scan for section headers (C uses `$` and `#`; in practice each morph is introduced by a single leading `Morph <name>` — confirm during G1 by reading `polymorph.c:1757-1819`).

### 4.5 Persistence (pfile `#MorphData` block)

Add `writeMorphData(ch, buf)` inside `persist/player.go` SavePlayer. Emit `#MorphData\n` header, all non-default `CharMorph` fields (per `fwrite_morph_data` at `polymorph.c:2405-2484`), `End\n`. Parse in LoadPlayer when section token `#MorphData` seen: allocate blank `CharMorph`, populate fields, attach to `ch.Morph`. **Reapply path on login:** after LoadPlayer attaches `ch.Morph`, the delta stats (AC, ModStr, saves, hp, move, mana, affects) are NOT re-added — the pfile already stored them baked into the base stats (that's how C works: `ch->armor` on disk already includes the morph delta). So on unmorph (`DoUnmorph`), the delta stored on `CharMorph` is subtracted out. This means `SavePlayer` must emit the pfile with the morph still active — matching C `save.c:288` unconditionally.

### 4.6 Commands

All 6 new commands live in a new `internal/act/polymorph.go` (or `internal/act/morph.go`). Immortal-level gate via `ch.GetTrust() >= LEVEL_IMMORTAL` check at the top of each.

- `DoMorph(ch, arg)` — mirrors `do_imm_morph`. Parses `<vnum>` or `<vnum> <target>`. Calls `DoMorphChar`.
- `DoUnmorph(ch, arg)` — mirrors `do_imm_unmorph`.
- `DoMorphset(ch, arg)` — field editor. **Largest of the six (~700 C LOC).** Subcommands: `help`, `desc`, `save`, plus per-field setters for every MorphData field. C uses two substate callbacks (`sub_morph_desc`, `sub_morph_help`) that invoke the editor and persist via `EditorSave` pattern — reuse the Tier 12 `EditorSave` callback seam (already shipped).
- `DoMorphstat(ch, arg)` — display a morph by name.
- `DoMorphcreate(ch, arg)` — allocate blank MorphData, append to world.Morphs, call `SetupMorphVnum`, save.
- `DoMorphdestroy(ch, arg)` — find morph, `UnmorphAll(morph)`, remove from world.Morphs, save.

### 4.7 Package placement

Trade-off: `internal/polymorph/` (new package) vs `internal/act/polymorph.go` (existing act package). The `DoMorph`/`DoUnmorph` apply/remove primitives are called from both `mpmorph` (inside mudprog package) and the command layer (inside act package). To avoid import cycles, put the primitives (`DoMorph`, `DoUnmorph`, `MakeCharMorph`, `GetMorph`, `FindMorph`, `DoMorphChar`, `DoUnmorphChar`, `SendMorphMessage`) in `internal/handler/polymorph.go` — `handler` is already imported by both `act` and `mudprog`. Put the command entry points (`DoMorph <arg>` the immortal command, `DoMorphset`, etc.) in `internal/act/polymorph.go`.

### 4.8 AT_MORPH constant

Verify presence at `internal/types/ansi.go` or equivalent. C defines `AT_MORPH` in `src/color.h`. If missing, add it with C's value. G0 task.

---

## 5. Task Groups

Execution order G0 → G1 → G2 → G3 → G4 → G5 → G6. G1 and G2 can be parallelized after G0 since they touch different files.

### G0 — Prerequisites verification + missing primitives

1. Verify `AT_MORPH` color constant exists in `internal/types/ansi.go` (or equivalent). If missing, add it. Write a test that `util.Act(AT_MORPH, ...)` doesn't panic.
2. Expose dice-parse helper for morph code. Go already has a grammar-aware evaluator at `internal/magic/spell_smaug.go:455` — `parseDiceExpr(s string, level int) int` — but it is package-private in `internal/magic`. Options: (a) add an exported wrapper `magic.ParseDiceExpr`, or (b) move the helper to `internal/util/dice.go` and import from magic. Note: C source is `src/magic.c:1059-1066` (`dice_parse(ch, level, exp)` → `rd_parse(ch, level, buf)`); the `ch` parameter is used only for random-number seeding in `rd_parse` and can be dropped in the Go port since Go's helper already uses the package `rand` source. Test: `"1d1+3"` → 4; `"0"` → 0; `""` → 0.
3. ~~Verify `update_aris(ch)` equivalent~~ — **Resolved: omit the call.** `internal/persist/player_affect_test.go:11-22` documents that Go does not implement `update_aris` because the architectural invariant is incremental direct-mutation via `AffectModify` on each `AffectToChar`/`AffectRemove`. Since `DoMorph`/`DoUnmorph` already direct-add/subtract every field they touch, a Go `UpdateAris` call is both absent and redundant. Drop the `UpdateAris(ch)` line from §4.3 and §G2.3 step 3. (This resolves Open Q2.)
4. Verify `COND_BLOODTHIRST` constant exists for vampire blood path.
5. Verify `LEVEL_IMMORTAL` gate pattern matches other immortal commands (`DoMset` is a good template).

**Mutation verify:** introduce a test that `util.DiceParse(ch, 30, "2d4+1")` returns a value in `[3,9]`; corrupt the dice parse to always return 0; test fails. Revert via `Edit`.

**Files touched:** `internal/types/ansi.go` (possibly), `internal/util/dice.go` (possibly).

### G1 — Morph-table schema + loader/saver

1. Add to `world.World`: `Morphs []*types.MorphData`. Optionally add `MorphsByVnum map[int]*types.MorphData` derived helper.
2. Implement `GetMorph(world, name)` / `GetMorphVnum(world, vnum)` / `FindMorph(ch, name, isCast)` lookups.
3. Implement `persist.LoadMorphs(path) ([]*types.MorphData, error)` in `internal/persist/morphs.go`. Token dispatch per C `fread_morph` (`polymorph.c:1821+`).
4. Implement `persist.SaveMorphs(path, []*types.MorphData) error` — token-for-token per C `fwrite_morph` (`polymorph.c:1613+`).
5. Wire `LoadMorphs` into `internal/boot/boot.go` after area load (mirrors `db.c:941`).
6. Add stock `db/system/morph.dat` test fixture (copy one small morph from real data if present; otherwise synthesize a minimal record with one non-zero field per category).
7. Unit tests: round-trip load→save→reload produces identical morph slice. Test load of empty file. Test load of file with missing optional fields.

**Mutation verify:** corrupt the `fwrite_morph` field for `"Damroll"` to always emit `"0"`; round-trip test fails. Revert via `Edit`.

**Files touched:** `internal/world/world.go` (field), `internal/persist/morphs.go` (new), `internal/persist/morphs_test.go` (new), `internal/boot/boot.go` (wire), `db/system/morph.dat` (fixture; check if already present).

### G2 — Apply / remove primitives + send_morph_message

1. Implement `handler.MakeCharMorph(m *MorphData) *CharMorph` (initializes with `Morph: m`, `Timer: -1`, zero everything else).
2. Implement `handler.DoMorph(ch, morph)` per §4.3.
3. Implement `handler.DoUnmorph(ch)` — exact reverse of DoMorph, then `ch.Morph = nil`. No `UpdateAris` call (see G0.3; Go invariant). Do not call `AffectRemove`/`AffectModify` on the morph deltas — they were never registered as affects.
4. Implement `handler.DoMorphChar(ch, morph) bool` per C `polymorph.c:1262-1384`: prereq gate (obj consumption, hp/mana/move/blood/favour/glory costs) → `SendMorphMessage(ch, morph, true)` → `DoMorph(ch, morph)` → return true. Bails with "You begin to transform, but something goes wrong." on any prereq fail. **C-fidelity note: resources are deducted and objs extracted as each check runs, NOT atomically after all checks pass** (polymorph.c:1319-1373). If `hpused` succeeds but `gloryused` fails, hp deduction stays and `objuse[0]` items already extracted are NOT refunded. Preserve this check-and-spend ordering in Go; do not reorder to atomic reservation.
5. Implement `handler.DoUnmorphChar(ch)` per C `polymorph.c:1387-1396`: capture `morph := ch.Morph.Morph`, call `DoUnmorph(ch)`, call `SendMorphMessage(ch, morph, false)`.
6. Implement `handler.SendMorphMessage(ch, morph, isMorph)` per C `polymorph.c:1402-1418`: emit `AT_MORPH` act to room and to char using `morph.MorphOther`/`MorphSelf` or `UnmorphOther`/`UnmorphSelf`.
7. Implement `handler.CanMorph(ch, morph, isCast) bool` per C `polymorph.c:~1420+` (level, class, race, sex, day/time, pkill/peaceful gates).
8. Implement `handler.UnmorphAll(world, morph)` per C `polymorph.c:2655-2668`.

**Tests (G2):**
- Apply sets the expected hitroll/damroll/hit/move/ac/mod_* deltas on `ch`.
- Unmorph fully reverses the deltas (round-trip equality on all fields).
- Applying a morph when one is already active first unmorphs the prior one.
- `DoMorphChar` with insufficient hp bails with the exact C message.
- `SendMorphMessage` emits to room on true/false.
- `UnmorphAll` clears morph on every PC using the target morph.

**Mutation verify:** corrupt DoMorph to double-add `morph.AC`; unmorph round-trip test fails. Revert via `Edit`.

**Files touched:** `internal/handler/polymorph.go` (new), `internal/handler/polymorph_test.go` (new).

### G3 — Player-visible commands (immortal `morph` / `unmorph`)

1. Implement `act.DoMorph(ch, arg)` per C `do_imm_morph` (`polymorph.c:2676-2723`).
2. Implement `act.DoUnmorph(ch, arg)` per C `do_imm_unmorph` (`polymorph.c:2729-2753`).
3. Register both in `internal/boot/boot.go` command table. Immortal-level gate.
4. Tests: immortal types `morph 1000` → world morph vnum 1000 applied to self. `morph 1000 victim` → applied to victim. Non-immortal rejected. `unmorph` clears self. Non-existent vnum reports "No such morph N exists."

**Mutation verify:** flip the immortal-level check to `<`; non-immortal can now morph; test fails. Revert via `Edit`.

**Files touched:** `internal/act/polymorph.go` (new), `internal/act/polymorph_test.go` (new), `internal/boot/boot.go`.

### G4 — Immortal admin commands (morphset / morphstat / morphcreate / morphdestroy)

1. Implement `act.DoMorphstat(ch, arg)` — pretty-print a morph's fields. Mirror C `do_morphstat` (`polymorph.c:969+`).
2. Implement `act.DoMorphcreate(ch, arg)` — allocate blank, call `MorphDefaults`, append to `world.Morphs`, run `SetupMorphVnum` (assigns vnum ≥ 1000). C `polymorph.c:2313+`.
3. Implement `act.DoMorphdestroy(ch, arg)` — find morph, `UnmorphAll`, splice out, save. C `polymorph.c:2370+`.
4. Implement `act.DoMorphset(ch, arg)` — field editor. Subcommands per C `do_morphset` (`polymorph.c:72-195`): `help`, `desc`, `save`, plus setters for every `MorphData` field. `save` calls `persist.SaveMorphs`. `help` and `desc` use editor substate via Tier 12 `EditorSave` callback.
5. Tests: create → edit one field → save → reload → field persisted. Destroy → reload → gone. Stat displays all fields without crashing.

**Scope note (Morphset size):** C `do_morphset` is **889 LOC** (`polymorph.c:72-960`) of if/else-if per-field dispatch, not ~700 as originally estimated. Port is mechanical but tedious. **Adopt G4a/G4b split as mandatory:** G4a = structure + create/destroy/stat + save subcommand; G4b = full field coverage (~70+ fields). G4a ships a working morph-admin surface with a minimal field set (name, vnum, short_desc, level, ac); G4b extends to all fields.

**Mutation verify:** corrupt morphcreate to skip SetupMorphVnum; new morph has vnum 0; test fails. Revert via `Edit`.

**Files touched:** `internal/act/polymorph.go` (extended), `internal/act/polymorph_test.go`, `internal/boot/boot.go`.

### G5 — Mudprog mpmorph / mpunmorph bodies

1. Replace the stub at `internal/mudprog/commands.go:545-554` with real implementations per C `do_mpmorph` / `do_mpunmorph` at `mud_comm.c:1964-2070`.
2. `mpMorph(mob, "<target> <morphname>")`: resolves victim via `handler.GetCharRoom(mob, arg1)`; resolves morph via `GetMorph(world, arg2)`; calls `handler.DoMorphChar(victim, morph)`. Errors call `mpbug`.
3. `mpUnmorph(mob, "<target>")`: resolves victim; calls `handler.DoUnmorphChar(victim)`.
4. Remove `TODO(tier3)` comments.
5. Tests: mobprog trigger fires `mpmorph player testform` → victim is morphed. `mpunmorph player` → victim is unmorphed.

**Mutation verify:** swap arg1 and arg2 parsing in mpMorph; test fails. Revert via `Edit`.

**Files touched:** `internal/mudprog/commands.go`, `internal/mudprog/commands_test.go` (or `mpmorph_test.go` new).

### G6 — Per-player `#MorphData` pfile block

1. Implement `persist.writeMorphData(ch, w)` inside `player.go`. Emit per C `fwrite_morph_data` (`polymorph.c:2405-2484`). **Ordering constraint: `Vnum` line MUST be emitted before `Name` line.** C `fread_morph_data` validates Name against the already-resolved `morph->morph->name` (polymorph.c:2540-2547), guarded by `if (morph->morph)` which is only non-nil after the Vnum line has been processed. Reordering (e.g. from a map iteration) silently disables validation.
2. Implement `persist.readMorphData(ch, scanner)` inside `player.go`. Parse per C `fread_morph_data` (`polymorph.c:2487+`). Insertion point: new `case "#MorphData"` alongside `case "#OBJECT", "#CORPSE"` at `internal/persist/player.go:68`.
3. Hook `writeMorphData` into SavePlayer — emit when `ch.Morph != nil` (mirrors `save.c:287-288`).
4. Hook `readMorphData` into LoadPlayer — on `#MorphData` section token (mirrors `save.c:1082-1083`).
5. Tests:
   - Save player mid-morph → reload → `ch.Morph != nil`, `ch.Morph.Morph` resolves to the named morph in `world.Morphs`.
   - Save player post-unmorph → reload → `ch.Morph == nil`, no `#MorphData` block in pfile.
   - Load player whose morph name no longer exists in `world.Morphs` → `ch.Morph.Morph` is nil; LoadPlayer does not panic (log a warn). **Add this defensive path explicitly** — C's `fread_morph_data` assumes the morph is found; Go should degrade gracefully.

**Mutation verify:** corrupt writeMorphData to skip the `Vnum` line; reload fails to find the morph reference; test fails. Revert via `Edit`.

**Files touched:** `internal/persist/player.go`, `internal/persist/player_test.go`.

---

## 6. Acceptance Criteria

1. `world.Morphs` is a populated slice after boot (≥1 morph loaded from `db/system/morph.dat`).
2. `GetMorph(world, "<name>")` returns the expected `*MorphData`.
3. `persist.LoadMorphs` → `persist.SaveMorphs` → reload produces a morph slice equal to the original (field-by-field).
4. `handler.DoMorph(ch, m)` applies all 32 stat-override fields (armor, mod_*, saves, hitroll, damroll, hit, move, mana or blood, affects, resists).
5. `handler.DoUnmorph(ch)` reverses every delta applied by `DoMorph`: `ch` equals its pre-morph state field-by-field (excluding `Morph` which is now nil).
6. `DoMorphChar` prereq gates: insufficient hp rejects with "You begin to transform, but something goes wrong."; missing required object rejects; sufficient resources succeed.
7. `DoMorphChar` on an already-morphed char first unmorphs prior form.
8. `SendMorphMessage` emits `AT_MORPH`-colored act text to both char and room using correct morph_self/morph_other or unmorph_self/unmorph_other.
9. Immortal command `morph <vnum>` applies morph to self; `morph <vnum> <target>` applies to target; `unmorph` reverses. Non-immortal rejected.
10. Immortal `morphcreate <name>` assigns vnum ≥ 1000 (SetupMorphVnum rule).
11. Immortal `morphdestroy <name>` unmorphs all current users, removes from table, saves file.
12. Immortal `morphset <morph> <field> <value>` edits the field and `morphset <morph> save` persists.
13. Immortal `morphstat <name>` pretty-prints fields without crashing.
14. Mudprog `mpmorph <target> <morphname>` from mob applies morph to target PC in same room.
15. Mudprog `mpunmorph <target>` from mob unmorphs target.
16. Player saved mid-morph reloads with `ch.Morph` populated, stat deltas still present in base stats (C semantics).
17. Player whose morph name is deleted from `morph.dat` between sessions loads without panic; `ch.Morph` is nil or `ch.Morph.Morph` is nil with a log warn.
18. `ismorphed` and `morph` mudprog ifchecks continue to pass.
19. No regression in existing `ifcheck_tier3_test.go`.
20. `go test ./...` and `go vet ./...` pass.

---

## 7. Scope Cuts / Deferrals

- **`spell_polymorph`** — C has a spell that wraps `do_morph_char` with a mage/druid cast path. Phase 6 Polymorph ships the subsystem; wiring `spell_polymorph` is deferred. Rationale: players enter morphs today via `morph` (immortal) or `mpmorph` (mudprog). Spell wrapping is a small add-on once G1-G5 ship, and does not belong in the morph subsystem plan.
- **Morph editor menu (full interactive CON_OEDIT-style)** — Morphset uses flat subcommand dispatch per C, not an interactive menu. Menu-driven morph editor is deferrable to Phase 6 Wave 3 (Interactive OLC) if desired.
- **Vampire blood accounting** — G2 ports the blood path but does not test it in depth; full vampire regression is outside this plan.
- **OLC `mpedit` awareness of morph-driven mudprog bodies** — mpmorph/mpunmorph strings in mudprog programs are already parsable as free-form text; no editor integration needed beyond the dispatch wiring in G5.
- **`find_morph(..., isCast=true)` spell-cast filter** — CanMorph respects the flag, but since no caller passes `isCast=true` in this plan (only `mpmorph` which uses `false`), the isCast-specific gates (level-cast-required checks) are structurally reachable only via spell_polymorph and thus deferred with §7.1.
- **Morph list UI (`morphlist` / `moprlist`)** — C has no such command. Not ported.
- **Morph-triggered mob programs (a separate MPTRIG_MORPH / MPTRIG_UNMORPH trigger type)** — C has `do_mpmorph`/`do_mpunmorph` as mob *actions* (things mobs do), not as triggers fired when a PC morphs. No trigger type port needed.

---

## 8. Open Questions

1. **~~`util.DiceParse` name/location.~~** **Resolved (audit 2026-04-18):** Go has `parseDiceExpr(s, level)` at `internal/magic/spell_smaug.go:455`, package-private. G0.2 promotes it (exported wrapper or move to `util`). See G0.2 for details.
2. **~~`update_aris` equivalent.~~** **Resolved (audit 2026-04-18):** No equivalent exists or is needed. Go uses incremental direct-mutation (invariant documented in `internal/persist/player_affect_test.go:11-22`). `DoUnmorph` does not call any ARIS-recompute. See G0.3.
3. **`MorphData.Obj[3]` and `.ObjUse[3]` interaction with G2 DoMorphChar obj consumption.** Item extraction calls `separate_obj`/`extract_obj` — do Go equivalents exist under those names? Grep during G2.
4. **`db/system/morph.dat` stock data.** Does the existing `db/system/morph.dat` file parse cleanly with the G1 loader? If it uses any tokens not present in the C `fwrite_morph` list (unlikely but possible), G1 must handle them with a non-fatal warn-and-skip.
5. **`COND_BLOODTHIRST` array index.** Confirm Go PCData has a `Condition [N]int` array indexed by `COND_*` constants matching C (`COND_DRUNK`/`COND_FULL`/`COND_THIRST`/`COND_BLOODTHIRST`).
6. **Two-phase morph load.** C loads morphs *before* areas (`db.c` sequence — verify precise order). Go's boot sequence in `internal/boot/boot.go` should mirror this: morphs must be present before any mob's mudprog references them. Confirm ordering.
7. **`SetupMorphVnum` collision semantics.** If DoMorphcreate assigns vnum 1001 but `morph.dat` shipped with 1001, does C collide or increment? Answer by reading `polymorph.c:2630-2652`: C starts from `max(existing_vnums, 999)+1`. Port as-is.

---

## 9. Risk Analysis

**Medium risk:**
- **Morphset field coverage (G4).** 70+ mutable fields on MorphData. Port is mechanical but ~700 Go LOC. Primary risk: missing one field means the "save after edit" path silently drops that edit on reload. Mitigation: a generated test that exercises every field via round-trip (create morph → morphset each field → save → reload → assert).
- **pfile `#MorphData` backward compatibility.** Existing pfiles without `#MorphData` sections must still load (they would be pre-Polymorph saves). Mitigation: the section is optional — LoadPlayer only parses it if the token appears. Verified matches C behavior (`save.c:1082` is in an `else if` branch).

**Low risk:**
- Loader/saver parity (G1). Standard Scanner pattern.
- Apply/remove primitives (G2). Direct mutation is easy to test via round-trip.
- Mudprog bodies (G5). Dispatch entries already exist; only the body needs writing.

**High-leverage bug class:**
- **Asymmetric apply/remove.** If `DoMorph` adds a field but `DoUnmorph` forgets to subtract it, players accumulate stats across morph/unmorph cycles. Mitigation: **pair-test every field** — a single table-driven test that, for each MorphData field, sets a non-zero value, morphs, records delta, unmorphs, asserts ch is pre-morph.

---

## 10. Adversary Verification Notes

*(To be appended after plan-adversary pass.)*

Manager note: Agent-tool dispatch for adversary review is not available in this session. Self-review substituted. Key self-review checks performed:

- **Schema completeness:** cross-checked every field in `struct char_morph` (mud.h:524-558) and `struct morph_data` (mud.h:560-629) against the Go `CharMorph` and `MorphData` struct definitions. All fields present; no gaps identified. C `sh_int` → Go `int` is the established convention.
- **C LOC breakdown:** 2753 total. Commands: do_imm_morph (47, polymorph.c:2676-2723), do_imm_unmorph (24, polymorph.c:2729-2753), do_morphset (**889**, polymorph.c:72-960), do_morphstat (191, polymorph.c:969-1160), do_morphcreate (49, polymorph.c:2313-2362), do_morphdestroy (22, polymorph.c:2370-2392). Primitives: do_morph (~60), do_unmorph (~40), do_morph_char (~130), do_unmorph_char (~10), can_morph (~80), find_morph (~15), get_morph (~12), get_morph_vnum (~12), make_char_morph (~50), clear_char_morph (~35), free_char_morph (~20), send_morph_message (~18), morph_defaults (~55), setup_morph_vnum (~22), unmorph_all (~15). Persistence: save_morphs (~20), fwrite_morph (~140), load_morphs (~60), fread_morph (~500), fwrite_morph_data (~80), fread_morph_data (~200). Banner/comments/whitespace: remainder.
- **Open Q 1-7 are advisory, not blocking.** G0 resolves them as reconnaissance before G2/G1 execute.
- **Deliberate C-fidelity decision:** direct stat mutation (not AffectModify). Documented in §4.2.

---

## 11. Completion Record

*(To be appended after work lands.)*
