# Plan: Phase 6 — Combat Stances OLC completion

**Status:** Audited 2026-04-18 (CONCERNS) by `audit-stances-olc` lineage. Manager self-adversary pass complete (see §Adversary-Resolved Concerns). External audit delivered 4 surgical in-plan edits (Q9 closed, §D3 type-mismatch note added, §Q6 RIS table corrected to 22 names, this status line). 1 correctness-critical finding (type mismatch between `int` CharData fields and `uint32` RIS constants at fixture sites) addressed inline in §D3. 3 open questions remain truly open (Q2 class/race setter policy — human decision; Q6 full-range cross-check at G4; Q8 resolved — `World.DataDir` confirmed at `world.go:76`, prefer Option a).
**Priority:** Wave 2 of Phase 6 (`phase6-roadmap.md` — listed alongside Hotboot executable plan, Auction state machine, Clan-officer commands). Depends on Tranche B G1 (landed 2026-04-18).
**Scope:** Finish the non-combat half of the stance subsystem that Tranche B G1 deliberately read-and-discarded. Ships:
- `StanceInfo` struct extension (add: `Resist`, `Immune`, `Suscept`, `Dodge`, `Parry`, `Dual` (no-dual-wield flag), `MaxWeight`, `Wait`, `Prereq [2]int`, `Class`, `Race`, `SpecialPercent`, `SpecialMove`, `Self string`, `Others string`).
- `combat.CanUseStance(ch, newStance) bool` — mirror of C `can_use_stance` at `src/stances.c:236-281`.
- `combat.UpdateStances(ch, flag bool)` — mirror of C `update_stances` at `src/stances.c:131-146`.
- `combat.GetStanceMastery(ch)` and `combat.GetStanceName(idx)` — convenience helpers (read by `do_ststat` + persistence).
- `DoStance` rewrite: port the full C state machine at `stances.c:51-128` in place of the current name-only stub at `internal/act/skills4.go:180-199`.
- `DoSTstat` (`ststat` command) — stat-display at `stances.c:672-735`.
- `DoSTset` (`stset` command) — stance-table OLC at `stances.c:742-1010`, including the `stset save` subcommand that calls the new persistence writer.
- `persist.LoadStancesInto` extension — store the previously-discarded non-combat keys into the extended `StanceInfo`.
- `persist.SaveStances(target, path)` — new writer mirroring C `fwrite_stance` at `stances.c:599-670` (absent today).
- `internal/boot/boot.go` — register `ststat` + `stset` commands and replace the existing stance-placeholder registration if one exists; the loader wire already exists (Tranche B G1).

**Files:**
- Modified: `internal/combat/stance_index.go` (struct extension + helpers).
- New: `internal/combat/can_use_stance.go` + `_test.go` — separate file keeps the existing `stance_index.go` focused on data.
- New: `internal/combat/stance_update.go` + `_test.go` — `UpdateStances` bitset toggle.
- Modified: `internal/persist/stances.go` (readers for extended fields).
- New: `internal/persist/stances_save.go` + `_test.go` — `SaveStances` writer.
- Modified: `internal/act/skills4.go` — replace the `DoStance` stub.
- New: `internal/act/stances_admin.go` + `_test.go` — `DoSTstat` + `DoSTset`.
- Modified: `internal/boot/boot.go` — command registration.
- Testdata: `internal/persist/testdata/stances_full.dat` — round-trip fixture covering all extended fields.

---

## Problem

Tranche B G1 (2026-04-18; `plan-tranche-b.md`, committed at `eb8083c`) shipped the minimum viable stance loader: a `StanceInfo` struct with only the three combat-critical fields (`NumAttacks`, `DamDone`, `DamTaken`) and a `LoadStancesInto` reader that *drains* the rest of the C `fread_stance` keys (`Class`, `Dodge`, `Dual`, `Immune`, `Parry`, `Percent`, `Race`, `Resist`, `Suscept`, `Wait`, `Weight`, `Other`, `Self`, `Special`, `Stance`) without storing their values. See `internal/persist/stances.go:131-141` — the `case "Class", "Dodge", ..."Weight": _ = sc.ReadNumber()` block and the comment at `:60` explicitly defer storage to "Phase-6 `StanceInfo` extension".

As a result, six C subsystems cannot be ported cleanly until those fields land:

1. **Stance prerequisite check**. C `can_use_stance` (`src/stances.c:236-281`) reads `stance_index[new_stance].stance[0]` / `stance[1]` (grand-master prerequisite chain), `max_weight` (carry-weight gate), `dual_wield` (no-dual-wield flag), `class_restrictions`, `race_restrictions`. None are stored in Go. The current Go `DoStance` at `internal/act/skills4.go:180-199` ships as a bare name-to-index mapping with NO prerequisite check — players with `ch.Stance = STANCE_NONE` can select `Dragon` without first mastering any prerequisite stance. This is a live gameplay divergence from C.

2. **Stance resistance application**. C `update_stances` (`stances.c:131-146`) toggles `ch->stance_resistant`, `ch->stance_immune`, `ch->stance_susceptible` bitsets (load-bearing in `fight.c` R/I/S resolution) from the stance's `resist`/`immune`/`suscept` fields on entry, and reverses them on exit. The fields `StanceResistant`/`StanceImmune`/`StanceSusceptible` exist on `CharData` (character.go:128-130) but are never written by anyone — the symptom is that the full C stance-RIS mechanic is dead in production.

3. **Stance display**. C `do_ststat` (`stances.c:672-735`) dumps the stance table fields: Self/Others message strings, dual-wield flag, damage-done/taken, dodge/parry, attacks, max-weight, wait, R/I/S flags, special move + percent, class/race restrictions, prerequisite chain. No Go command.

4. **Stance-table OLC**. C `do_stset` (`stances.c:742-1010`) edits any stance-table field by name, with value-range guards (attacks `-5..5`, damage `0..200`, dodge/parry `-50..50`, dual `0..1`, percent `0..100`, wait `0..25`, weight `0..999`, R/I/S via flag names). Includes a `stset save` subcommand that writes the full table to `db/system/stances.dat` via `fwrite_stance`. No Go counterpart.

5. **Stance persistence (save path)**. C `fwrite_stance` (`stances.c:599-670`) writes all 18 per-stance keys including the conditional-write pattern (`if stance_index[i].immune > 0 fprintf(...)`). Go has no save path — edits made via OLC cannot persist, even if OLC existed.

6. **Wait state on stance change**. C `do_stance` calls `WAIT_STATE(ch, stance_index[new_stance].wait)` on entry and `WAIT_STATE(ch, stance_index[STANCE_NORMAL].wait)` on exit. Go cannot honour this without the `Wait` field.

Severity order: #1 (live gameplay divergence; any player can claim any stance) is the highest-impact. #2 (dead R/I/S mechanic) is silent — nobody notices because the stance data file is a stub. #3/#4/#5 are builder tooling. #6 is minor once #1 lands (stance switching becomes gated but also instantly-respammable without the wait).

The roadmap scope statement:

> `do_stset` full stance-table editor; `do_ststat` stat display; `can_use_stance` prerequisite check; `StanceInfo` struct extension (resist/immune/suscept/dodge/parry/max_weight/dual_wield/wait/prerequisite fields); `fwrite_stance` persistence of the non-combat fields.

This plan delivers all of that, plus the `update_stances` toggle (#2) and the `DoStance` rewrite (#1/#6) — because without them, the extended `StanceInfo` fields are dead storage. #1/#2/#6 are the "make the newly-stored data do something"; the OLC bits (#3/#4/#5) are "let immortals shape that data at runtime".

## C Reference (authoritative)

Citations against `src/stances.c`, `src/mud.h`, `src/build.c` (HEAD).

### `STANCE_DATA` struct — `src/mud.h` (search for `struct stance_data`)

Verified 2026-04-18: the C struct carries at least the following fields (names as they appear in `stances.c` accesses):

- `others` — `char *` — third-person room message on stance entry (AT_STANCE color).
- `self` — `char *` — first-person message on stance entry.
- `num_attacks` — int — extra attacks per round.
- `dam_done` — int — percent damage multiplier (attacker).
- `dam_taken` — int — percent damage divisor (victim).
- `dodge` — int — dodge percent delta (signed, -50..50).
- `parry` — int — parry percent delta (signed, -50..50).
- `dual_wield` — int — non-zero means "cannot enter this stance while dual-wielding".
- `max_weight` — int — carry-weight ceiling (0 = no ceiling).
- `wait` — int — WAIT_STATE on entry (0..25).
- `resist` — int — `RIS_*` bitset applied via `update_stances`.
- `immune` — int — same.
- `suscept` — int — same.
- `special_move` — int — index into a "special moves" table (stub in C `get_special_number` / `get_special_name` at `stances.c:1012-1022` — both return constants, feature unfinished).
- `special_percent` — int — trigger chance, 0..100.
- `class_restrictions` — int — bitmask of disallowed classes.
- `race_restrictions` — int — bitmask of disallowed races.
- `stance[2]` — int[2] — prerequisite chain: `stance[0]` must be GM before new stance allowed; `stance[1]` is an optional second requirement.

### Entry-point functions

- **`do_stance`** — `src/stances.c:51-128`. Flow:
  1. Parse one arg via `one_argument`.
  2. `ch->mount != NULL` → `"While you are mounted?\n\r"`; return.
  3. `arg[0] == '\0'` (no argument):
     - `ch->stance == STANCE_NONE` → enter `STANCE_NORMAL`; `send_stance_message(ch, TRUE)`; `update_stances(ch, TRUE)`; `WAIT_STATE(ch, stance_index[STANCE_NONE].wait)`.
     - Else → `update_stances(ch, FALSE)`; set `ch->stance = STANCE_NONE`; `send_stance_message(ch, TRUE)`; `WAIT_STATE(ch, stance_index[STANCE_NORMAL].wait)`.
     - Return.
  4. `ch->stance > STANCE_NONE` (attempting change while already in a non-NONE stance) → `"You cannot change stances until you come up from the one you are currently in.\n\r"`; return.
  5. Switch on `get_stance_number(arg)`:
     - `default:` → `send_stance_message(ch, FALSE)` (syntax); return.
     - `STANCE_VIPER..STANCE_SWALLOW`: if `can_use_stance(ch, new_stance)` → `ch->stance = new_stance`; break. Else `send_stance_message(ch, FALSE)`; return.
  6. `send_stance_message(ch, TRUE)`; `WAIT_STATE(ch, stance_index[new_stance].wait)`; `update_stances(ch, TRUE)`; return.

  **Note on exit WAIT_STATE value.** C line 89 passes `stance_index[STANCE_NORMAL].wait` on exit (not `STANCE_NONE.wait`). Intentional — exits should share the same wait regardless of stance.

- **`update_stances`** — `src/stances.c:131-146`. `SET_BIT` / `REMOVE_BIT` on `ch->stance_resistant/immune/susceptible` from the stance's `resist/immune/suscept` values. Called with `flag=TRUE` on entry, `flag=FALSE` on exit. Idempotent when entry/exit pair matches; asymmetric use is a bug.

- **`get_stance_mastery`** — `src/stances.c:151-160`. `STANCE_NONE → 0`; NPC → `pIndexData->stances[stance]`; PC → `pcdata->stances[stance]`.

- **`get_stance_name`** — `src/stances.c:163-195`. Switch on the enum → canonical-cased name; `default → NULL`. Already present in Go as a reverse lookup gap: Tranche B's `GetStanceNumber` gives name→int, but not int→name.

- **`get_stance_number`** — `src/stances.c:199-232`. Case-insensitive lookup (`str_cmp` is case-insensitive in SMAUG). **Already ported** as `persist.GetStanceNumber` at `internal/persist/stances.go:15-49`.

- **`can_use_stance`** — `src/stances.c:236-281`. Flow:
  1. `IS_NPC(ch)`: `ability = (pIndexData != NULL && pIndexData->stances[new_stance] > 0)`.
  2. PC:
     - Let `temp = stance_index[new_stance].stance[0]`.
     - If `temp <= 0` → `ability = TRUE` (no prerequisite).
     - Else if `pcdata->stances[temp] >= STANCE_GRAND_MASTER`:
       - Let `temp = stance_index[new_stance].stance[1]`.
       - If `temp <= 0` → `ability = TRUE`.
       - Else if `pcdata->stances[temp] >= STANCE_GRAND_MASTER` → `ability = TRUE`.
       - Else → `ability = FALSE`.
     - Else → `ability = FALSE`.
  3. Additional gates (applied after the above, overriding any `TRUE`):
     - `max_weight > 0 && ch->carry_weight > max_weight` → `ability = FALSE`.
     - `dual_wield && get_eq_char(ch, WEAR_DUAL_WIELD)` → `ability = FALSE`. **Note: C's `dual_wield` non-zero means "forbid dual-wielding while in this stance", not "requires dual wield" — the condition forbids entry if the flag is set AND the player has a dual-wield weapon equipped.**
     - `IS_SET(class_restrictions, ch->class)` → `ability = FALSE`. **Note: C uses `IS_SET`, which expects a bitmask. `ch->class` is a 0..MAX_CLASS index, so `IS_SET(mask, index)` actually tests `mask & index` — this is a C bug: the intent is clearly `IS_SET(mask, 1<<ch->class)`, but stock SMAUG's header defines `IS_SET(flag, bit)` as `((flag) & (bit))`. Result: if `class_restrictions == 1<<(warrior-class-index)`, the guard fires only for classes whose index shares bits with 1 — functionally broken. The Go port should preserve C's verbatim behaviour (bitmask AND index) to avoid diverging from shipped content that is already calibrated around the broken guard, with a comment flagging the bug. See §Open Q4.**
     - `IS_SET(race_restrictions, ch->race)` → same bug, same behaviour.
  4. Return `ability`.

- **`do_ststat`** — `src/stances.c:672-735`. Flow:
  1. `set_char_color(AT_PLAIN, ch)`.
  2. `one_argument(argument, arg)`; if empty: print `"STANCE list:\n\r"` then loop `0..MAX_STANCE-1` printing each `get_stance_name(i) " "`; the C code has a bug at line 687 (`!index_num % 4` parses as `(!index_num) % 4` == `0 % 4` == 0, always false except at index 0) — the newline never fires. Port preserves the spacing-no-newline behaviour OR fixes it (see §Open Q5).
  3. **The entire body at C `:692-693` is dead code:** `if (arg[0] != '\'' && arg[0] != '"' && strlen(argument) > strlen(arg)) strcpy(arg, argument);`. Go `OneArgument` already lowercases, so this quote-handling block has no Go analogue needed.
  4. `index = get_stance_number(argument)`; if `< 0 || >= MAX_STANCE` → `"That stance does not exist.\n\r"` + `do_help("STANCEFLAGS")`.
  5. Print Name / Self / Other / Required Stances / Dual Wield / Damage Done+Taken / Dodge+Parry / Attacks+MaxWeight / Wait / Resistance / Susceptible / Immune / Special Move+Chance / Class Restrictions / Race Restrictions — one line per attribute, as in the C source.

- **`do_stset`** — `src/stances.c:742-1010`. Flow:
  1. NPC → `"Mob's can't mset\n\r"`; return. (C preserves the typo — port verbatim.)
  2. `!ch->desc` → `"You have no descriptor\n\r"`; return.
  3. `set_char_color(AT_PLAIN, ch)`; `smash_tilde(argument)`.
  4. Parse three args: `arg1 = stance name`, `arg2 = field name`, `arg3 = rest (value)`.
  5. `!str_cmp("save", arg1)` → `fwrite_stance()`; `"Done.\n\r"`; return.
  6. `arg1 == "" || arg2 == "" || arg1 == "?"` → usage (6-line syntax block); return.
  7. `get_stance_number(arg1)` — if `< 0` → `"No such stance name.\n\r"` + `do_help("STANCEFLAGS")`.
  8. `value = is_number(arg3) ? atoi(arg3) : -100`.
  9. Switch on `UPPER(arg2[0])`:
     - `'A'`: `"attacks"` → range `-5..5`; `stance_index[index].num_attacks = value`.
     - `'C'`: `"class"` → `found = TRUE` **but no assignment** (C bug — setter for class_restrictions is absent; leaves it as a silent no-op). Port matches C behaviour; §Open Q2.
     - `'D'`: `"damage"` → range `0..200`; `dam_done = value`. `"dodge"` → `-50..50`; `dodge = value`. `"dual"` → `0..1`; **inverted assignment** (`if (value) stance_index[index].dual_wield = 0; else stance_index[index].dual_wield = 1;` — another C bug; §Open Q3).
     - `'I'`: `"immune"` → `value = get_risflag(arg3)` (0..31); `TOGGLE_BIT(immune, 1<<value)`.
     - `'O'`: `"others"` → free + `str_dup(arg3)`; stores the room-broadcast string.
     - `'P'`: `"parry"` → `-50..50`. `"percent"` → `0..100`; `special_percent = value`. `"protection"` → `0..200`; `dam_taken = value`.
     - `'R'`: `"race"` → `found = TRUE`, no assignment (same bug as class). `"resist"` → same as immune.
     - `'S'`: `"self"` → free + `str_dup`. `"special"` → `get_special_number(arg3)` (returns 0 in stock C). `"susceptible"` → same as immune/resist. `"stance1"` / `"stance2"` → `value = get_stance_number(arg3)`; if valid → `stance[0]` or `stance[1]`.
     - `'W'`: `"wait"` → `0..25`. `"weight"` → `0..999`.
  10. `!found` → recurse with `"?"` (show usage). Else → `"Done.\n\r"`.

  **Port decisions:**
  - `"class"` / `"race"` → Go preserves C's no-op behaviour with an explicit `// C bug: setter absent; see src/stances.c:824 / :921` comment; admins use `stset` expecting it to do something, but C leaves the field zero. Alternative: port the setter as `class_restrictions = get_risflag-equivalent(arg3)` (no such helper in C) — out of scope without a clear C referent. §Open Q2.
  - `"dual"` → Go preserves C's inverted assignment with a comment. The effect: `stset dragon dual 1` sets `dual_wield = 0` (allowing dual wield in Dragon); `stset dragon dual 0` sets `dual_wield = 1` (forbidding it). Live shipped behaviour; do not silently fix.
  - `fwrite_stance` port lives in `persist.SaveStances`; `stset save` invokes it.

- **`fwrite_stance`** — `src/stances.c:599-670`. Writer. For each stance `0..MAX_STANCE-1`: unconditionally write `StartStance\t<name>\n`, then conditionally write each key (only if non-zero / non-empty), then `EndStance\n\n`. After the loop: `End\n`. Conditional-write invariants (load-bearing — they ensure `load_stances` defaults survive round-trip):
  - `num_attacks != 0` → write.
  - `class_restrictions > 0` → write. (`> 0` not `!= 0` — negative values NOT written, but none should exist; asymmetric with num_attacks.)
  - `dam_done > 0` → write. (`> 0` not `!= 0`.)
  - `dam_taken > 0` → write.
  - `dodge != 0` → write.
  - `dual_wield` (non-zero truthy) → write.
  - `immune > 0` → write.
  - `others` (non-null) → write with trailing `~`.
  - `parry != 0` → write.
  - `special_percent > 0` → write.
  - `race_restrictions > 0` → write.
  - `resist > 0` → write.
  - `self` (non-null) → write with trailing `~`.
  - `special_move > 0` → write.
  - `stance[0] > 0` → write `Stance\t<name0> <name1>\n` (both names, space-separated; `stance[1]` passed through `get_stance_name` — if `stance[1]` is 0, C passes 0 to `get_stance_name` which returns `"None"`; safe).
  - `suscept > 0` → write.
  - `wait > 0` → write.
  - `max_weight > 0` → write (key is `"Weight"`, not `"MaxWeight"`).
  - `EndStance\n\n` — note the trailing blank line.
  - File terminator `End\n`.

### Wiring / registrations

- `src/tables.c:1463` — `"do_stance"` registered.
- `src/tables.c:1479-1482` — `"do_stset"` and `"do_ststat"` registered.
- `src/db.c` (search for `load_stances`) — called during boot.

### Constants

- `STANCE_GRAND_MASTER = 200` — `src/mud.h`. Already in Go at `types/constants.go:782`.
- `MAX_PC_STANCE = MAX_MOB_STANCE = 200` — same. Already in Go at `constants.go:89-90`.
- `STANCE_FILE` / `SYSTEM_DIR` — `src/mud.h`. Go path is `filepath.Join(dataDir, "system", "stances.dat")` — already wired at `internal/boot/boot.go:296`.
- `ris_flags` — bit-name table used by `get_risflag`. Go equivalent: string-to-bit map built from the `RIS_*` constants at `types/constants.go:196-215` (16 mapped names; values 0..15 cover the entire stance resist/immune/suscept useful range — §Open Q6).

---

## Go Current State

Verified 2026-04-18 on branch `golang` at HEAD (`26db5f9`).

### Struct — `combat.StanceInfo`

`internal/combat/stance_index.go:29-33`:

```go
type StanceInfo struct {
    NumAttacks int
    DamDone    int // percent, 0 means "no modifier"
    DamTaken   int // percent, 0 means "no modifier"
}
```

Only three fields. Package-level `var StanceIndex = [MAX_STANCE]StanceInfo{...}` with hand-picked defaults (lines 37-50). Consumers in the combat hot loop: `combat.go:319` (NPC extra-attack accumulator), `combat.go:342` (PC GM-bonus-attack loop), `combat.go:588-589` (attacker dam-done multiplier), `combat.go:601-602` (victim dam-taken divisor).

### Reader — `persist.LoadStancesInto`

At `internal/persist/stances.go:69-99`. Reads `StartStance <name>` blocks; for each key-value pair, stores `Attacks → NumAttacks`, `DamDone → DamDone`, `DamTaken → DamTaken`; discards the rest via the `case "Class", "Dodge", ..."Weight": _ = sc.ReadNumber()` block (line 131) and `case "Other", "Self": _ = sc.ReadString()` (line 134). `Special` (ReadWord) and `Stance` (two ReadWords) also discarded.

### Name lookup — `persist.GetStanceNumber`

At `internal/persist/stances.go:15-49`. Case-insensitive; returns `-1` on unknown. **Reverse lookup (`GetStanceName`) does NOT exist** — Tranche B didn't need it; we need it now for `DoSTstat` and `SaveStances`.

### `DoStance` stub — `internal/act/skills4.go:180-199`

Ships as a name-prefix matcher over a hard-coded 7-name slice (`mongoose bull mantis dragon tiger monkey swallow` — missing `viper crane crab normal`). Sets `ch.Stance` directly with no prerequisite check, no mount guard, no `ch.Stance > STANCE_NONE` change-while-active gate, no `update_stances` call, no wait state. This is the live Go gameplay divergence from C.

### Admin path — `DoMset` stance-name branch

At `internal/act/olc_set.go:158-203`. Tranche B G1b shipped the PC-mastery seed path. Writes `PCData.Stances[stanceIdx]` / `IndexData.Stances[stanceIdx]` + `CharData.Stances[stanceIdx]` for NPCs. Trust gate at `LEVEL_LESSER=57` for PC targets. Value clamp `0..200`. No stance-table field editing — that is the scope of `DoSTset`.

### Fields on `CharData`

`StanceImmune`, `StanceResistant`, `StanceSusceptible` at `character.go:128-130` — declared as `int`. **Present but never written by anyone.** `Stances [MAX_STANCE]int` at `character.go:206` (per-instance NPC copy) and `pcdata.go:98` (PC copy), written by the loader and by `DoMset`. `Wait int` at `character.go:91`. `CarryWeight int` at `character.go:115`.

### Persistence

The loader exists (Tranche B). **No writer** — no `persist.SaveStances`. `stset save` has nothing to call.

### Boot wiring

`internal/boot/boot.go:296-299` calls `persist.LoadStancesInto(&combat.StanceIndex, stancesPath)` — wired. No `ststat` / `stset` command registration exists (verified via `grep -n "ststat\|stset" internal/boot/*.go` → empty).

### Shipped data

`db/system/stances.dat` is the 4-byte `End\n` stub. First real exercise of this plan's persistence code comes from tests; production will see it once a builder runs `stset <stance> <field> <value> ; stset save`.

### Scanner

`internal/persist/scanner.go` — already offers `ReadWord`, `ReadNumber`, `ReadString` (reads to `~`), `ReadToEOL`. No `ReadFlag` or bit-name helper. C's `get_risflag` must be ported to Go (new helper; §Open Q6).

### Missing helpers

- `combat.GetStanceName(idx int) string` — reverse of `GetStanceNumber`. Used by `DoSTstat`, `SaveStances`, `stset stance1/stance2` error messages.
- `combat.GetStanceMastery(ch) int` — thin wrapper reading `PCData.Stances[ch.Stance]` or `IndexData.Stances[ch.Stance]`.
- `util.GetRisflag(name string) int` OR `types.RisflagIndex(name)` — maps a flag name (`"fire"`, `"cold"`, `"electricity"`...) to its bit position `0..15+`. C implementation lives in `src/act_wiz.c` (function `get_risflag`).

---

## Go Design

### D1 — `StanceInfo` struct extension

Append new fields to the existing `StanceInfo` struct at `combat/stance_index.go:29`. Keep the original three at the top for diff readability:

```go
type StanceInfo struct {
    // Combat-critical (Tranche B G1)
    NumAttacks int
    DamDone    int // percent
    DamTaken   int // percent

    // Phase-6 OLC extension — stances.c STANCE_DATA coverage
    Dodge         int  // signed percent, -50..50
    Parry         int  // signed percent, -50..50
    Dual          int  // non-zero = forbid dual-wield while in stance (C: dual_wield)
    MaxWeight     int  // 0 = no ceiling
    Wait          int  // WAIT_STATE on entry, 0..25
    Resist        int  // RIS_* bitset
    Immune        int  // RIS_* bitset
    Suscept       int  // RIS_* bitset
    Class         int  // class_restrictions bitmask (C bug — see CanUseStance comment)
    Race          int  // race_restrictions bitmask (same bug)
    SpecialMove    int // index; stub in both C and Go (get_special_number returns 0)
    SpecialPercent int // 0..100
    Prereq        [2]int // stance[0] = primary prereq GM; stance[1] = optional secondary
    Self          string // first-person entry message (AT_STANCE)
    Others        string // third-person room entry message
}
```

**Backward compatibility.** The default `StanceIndex` var at `stance_index.go:37-50` currently names only the three combat fields. Go struct literals with named fields are robust to added fields (new fields initialise to zero). The default table stays byte-compatible.

**Rejected alternative A — separate `StanceDetails` map keyed by `int`.** Rationale for rejection: `can_use_stance` is called on every stance-change attempt and reads five fields (prereq × 2, max_weight, dual, class, race) — separating them from the in-memory representation makes every lookup two steps (`StanceIndex[i].NumAttacks` + `StanceDetails[i].MaxWeight`). Cognitive overhead not worth the token savings.

**Rejected alternative B — retain existing three-field struct and define `FullStanceInfo` wrapping it.** Same problem plus a confusing two-tier type system for readers.

### D2 — `CanUseStance`

New file `internal/combat/can_use_stance.go`. Signature: `func CanUseStance(ch *types.CharData, newStance int) bool`. Port C `stances.c:236-281` verbatim, with the C bugs on `class`/`race` restrictions preserved and commented (§Open Q4). Read `StanceIndex[newStance]` for the new fields.

Dependencies: `handler.GetEqChar(ch, WEAR_DUAL_WIELD)` exists at `handler/handler.go:511-512` — but that creates a `combat` → `handler` import. The `combat` package currently imports only `types` / `util`. Adding `handler` is a risky circular-import invitation.

**Chosen:** inline the equivalent check. `ch.Equipment` is a `[MAX_WEAR]*ObjData` (verified) — use `ch.Equipment[WEAR_DUAL_WIELD] != nil` directly.

Validation: `grep -n "Equipment\s*\[" internal/types/character.go` — `Equipment [MAX_WEAR]*ObjData` confirmed at character.go (to re-verify exact line during G1 implementation).

### D3 — `UpdateStances`

New file `internal/combat/stance_update.go`. Signature: `func UpdateStances(ch *types.CharData, entering bool)`. Port C `stances.c:131-146` verbatim:

```go
func UpdateStances(ch *types.CharData, entering bool) {
    if ch.Stance < 0 || ch.Stance >= types.MAX_STANCE {
        return
    }
    info := StanceIndex[ch.Stance]
    if entering {
        ch.StanceResistant |= info.Resist
        ch.StanceImmune |= info.Immune
        ch.StanceSusceptible |= info.Suscept
    } else {
        ch.StanceResistant &^= info.Resist
        ch.StanceImmune &^= info.Immune
        ch.StanceSusceptible &^= info.Suscept
    }
}
```

Guard against `ch.Stance` out-of-range: C does not; Go adds it defensively because untyped int constants have no bounds check. Zero cost; covers malformed fixture data.

**Type note (added 2026-04-18 audit).** `CharData.StanceResistant / StanceImmune / StanceSusceptible` are declared `int` at `character.go:128-130`. RIS bit constants (`types.RIS_FIRE..types.RIS_PARALYSIS`) are declared **`uint32`** at `constants.go:196-217`. `StanceInfo.Resist / Immune / Suscept` will be `int` (D1 matches `CharData` side). Anywhere test code or fixture setup seeds those struct fields with RIS constants directly — e.g., `StanceIndex[DRAGON].Resist = int(types.RIS_FIRE)` — an explicit `int(...)` conversion is required. Inside `UpdateStances` itself no cast is needed (both sides are `int`). Flag this at G1/G2 implementation so test fixtures compile on first pass.

### D4 — `DoStance` rewrite

Replace the body at `internal/act/skills4.go:180-199`. Keep the signature `func DoStance(ch *types.CharData, argument string)`. Full state machine per C `stances.c:51-128`:

```go
func DoStance(ch *types.CharData, argument string) {
    arg, _ := util.OneArgument(argument)

    if ch.Mount != nil {
        ch.Send("While you are mounted?\n\r")
        return
    }

    if arg == "" {
        if ch.Stance == types.STANCE_NONE {
            ch.Stance = types.STANCE_NORMAL
            sendStanceMessage(ch, true)
            combat.UpdateStances(ch, true)
            ch.Wait = maxInt(ch.Wait, combat.StanceIndex[types.STANCE_NONE].Wait)
        } else {
            combat.UpdateStances(ch, false)
            ch.Stance = types.STANCE_NONE
            sendStanceMessage(ch, true)
            ch.Wait = maxInt(ch.Wait, combat.StanceIndex[types.STANCE_NORMAL].Wait)
        }
        return
    }

    if ch.Stance > types.STANCE_NONE {
        ch.Send("You cannot change stances until you come up from the one you are currently in.\n\r")
        return
    }

    newStance := persist.GetStanceNumber(arg)
    switch newStance {
    case types.STANCE_VIPER, types.STANCE_CRANE, types.STANCE_CRAB,
        types.STANCE_MONGOOSE, types.STANCE_BULL, types.STANCE_MANTIS,
        types.STANCE_DRAGON, types.STANCE_TIGER, types.STANCE_MONKEY,
        types.STANCE_SWALLOW:
        if !combat.CanUseStance(ch, newStance) {
            sendStanceMessage(ch, false)
            return
        }
        ch.Stance = newStance
    default:
        sendStanceMessage(ch, false)
        return
    }

    sendStanceMessage(ch, true)
    ch.Wait = maxInt(ch.Wait, combat.StanceIndex[newStance].Wait)
    combat.UpdateStances(ch, true)
}
```

Helper `sendStanceMessage(ch, flag bool)`:

```go
func sendStanceMessage(ch *types.CharData, entering bool) {
    if entering {
        info := combat.StanceIndex[ch.Stance]
        if info.Others != "" {
            util.Act(types.AT_STANCE, info.Others, ch, nil, nil, nil, types.TO_ROOM)
        }
        if info.Self != "" {
            util.Act(types.AT_STANCE, info.Self, ch, nil, nil, nil, types.TO_CHAR)
        }
    } else {
        ch.Send("Syntax is: stance <style>.\n\r")
        ch.Send("Stance being one of: Viper, Crane, Crab, Mongoose, Bull.\n\r")
    }
}
```

**C-fidelity note — `Others`/`Self` nil guard.** C `send_stance_message` at `stances.c:292-295` unconditionally calls `act` with `stance_index[ch->stance].others` — if `others` is NULL, `act` crashes in C. Go defensively skips empty strings instead of calling `util.Act` with `""`. `util.Act` would output nothing anyway, but the skip makes the nil-guard explicit. §Open Q7.

**C-fidelity note — `WAIT_STATE` semantics.** SMAUG's `WAIT_STATE(ch, pulse)` macro is `(ch)->wait = UMAX((ch)->wait, pulse)`. Go uses `ch.Wait = maxInt(ch.Wait, val)`. `maxInt` is a local helper — skip using `int` built-ins to avoid Go 1.21 version gates; a 2-line helper is trivial.

**Rejected alternative — add a `WaitState(ch, pulse)` helper in `handler`.** Would slightly centralise the pattern but adds a call site for something used 10+ times in the port via inline `UMAX`-style expressions. Not worth the churn in-scope for this plan; can be a follow-up refactor.

### D5 — `DoSTstat`

New file `internal/act/stances_admin.go`. Signature: `func DoSTstat(ch *types.CharData, argument string)`.

Port C `stances.c:672-735` verbatim for the display layout. No-argument path prints the stance list (Name1 Name2 ... NameN) — port preserving C's "newline-every-4" intent **correctly** (§Open Q5 resolves toward "fix the bug in Go, port the intent"). Argument path: `GetStanceNumber(arg)` → on fail, `"That stance does not exist.\n\r"` + a reference to the STANCEFLAGS help. On success, 13 one-line fields per the layout at C `:701-734`.

Helpers used:
- `combat.GetStanceName(idx)` — new helper in `stance_index.go`.
- `util.FlagString(bits int, names []string) string` — new util helper; C `flag_string(flags, const char **names)` port. Walks bit positions, appends any name whose bit is set. §Open Q6 covers the name table.
- `util.ClassString(mask int) string` — TBD; could also be a call into the class-table registry. Out of scope for the initial land if `class`/`race` setter is no-op anyway (§Open Q2); `DoSTstat` can print the raw integer mask as `"Class Restrictions: 0x%x\n\r"` until the richer formatter lands. Documented as a scope cut.
- `combat.GetSpecialName(idx)` — mirrors C `get_special_name` at `stances.c:1018-1022` (returns `"None"` always). Cheap port.

### D6 — `DoSTset`

Same file `internal/act/stances_admin.go`. Signature: `func DoSTset(ch *types.CharData, argument string)`.

Port C `stances.c:742-1010`. Structure:

1. NPC → `"Mob's can't mset\n\r"` (port the typo verbatim).
2. No descriptor (`ch.Desc == nil`) → `"You have no descriptor\n\r"`.
3. `util.SmashTilde(&argument)` or equivalent — C `smash_tilde` is already shimmed in Go persist.
4. Parse 3 args (`OneArgument` twice plus the rest).
5. `arg1 == "save"` → `persist.SaveStances(&combat.StanceIndex, filepath.Join(WorldRef.DataDir, "system", "stances.dat"))`; `"Done.\n\r"`.
   - **Path derivation.** The loader uses `filepath.Join(dataDir, "system", "stances.dat")` via the boot call. `WorldRef.DataDir` is required to round-trip the same path. §Open Q8 — verify the field exists.
6. `arg1 == "" || arg2 == "" || arg1 == "?"` → 6-line usage block (port verbatim).
7. `GetStanceNumber(arg1)` → on fail, `"No such stance name.\n\r"` + help reference.
8. `value = util.IsNumber(arg3) ? util.Atoi(arg3) : -100`.
9. Giant switch on `strings.ToLower(arg2)`:
    - `"attacks"` → range `-5..5`; `StanceIndex[idx].NumAttacks = value`.
    - `"class"` → matched, no-op (preserve C bug; comment).
    - `"damage"` → `0..200`; `.DamDone = value`.
    - `"dodge"` → `-50..50`.
    - `"dual"` → `0..1`; inverted assignment (preserve C bug).
    - `"immune"` → `get_risflag`-equivalent (0..31); `StanceIndex[idx].Immune ^= 1 << value`.
    - `"others"` → `.Others = arg3`.
    - `"parry"` → `-50..50`.
    - `"percent"` → `0..100`; `.SpecialPercent`.
    - `"protection"` → `0..200`; `.DamTaken`.
    - `"race"` → matched, no-op (C bug).
    - `"resist"` → same as immune.
    - `"self"` → `.Self = arg3`.
    - `"special"` → `get_special_number`-equivalent (returns 0); `.SpecialMove = value`.
    - `"susceptible"` → same as immune.
    - `"stance1"` / `"stance2"` → `get_stance_number(arg3)`; if >0, set `.Prereq[0]` or `.Prereq[1]`.
    - `"wait"` → `0..25`.
    - `"weight"` → `0..999`; `.MaxWeight`.
10. No match → recurse with `"?"`.
11. Match → `"Done.\n\r"`.

### D7 — `persist.SaveStances`

New file `internal/persist/stances_save.go`. Signature:

```go
func SaveStances(target *[types.MAX_STANCE]combat.StanceInfo, path string) error
```

Port C `fwrite_stance` at `stances.c:599-670`. Write atomically: write to a `path+".tmp"` sibling, fsync, rename (pattern used by `persist/player.go` save path — verify idiom at implementation time; if persist uses plain `os.Create` without atomic rename, match the existing convention rather than introducing a new one).

Per-stance block:
```
StartStance\t<GetStanceName(i)>\n
Attacks\t<n>\n              if NumAttacks != 0
Class\t<n>\n                if Class > 0
DamDone\t<n>\n              if DamDone > 0
DamTaken\t<n>\n             if DamTaken > 0
Dodge\t<n>\n                if Dodge != 0
Dual\t<n>\n                 if Dual != 0
Immune\t<n>\n               if Immune > 0
Other\t<s>~\n               if Others != ""
Parry\t<n>\n                if Parry != 0
Percent\t<n>\n              if SpecialPercent > 0
Race\t<n>\n                 if Race > 0
Resist\t<n>\n               if Resist > 0
Self\t<s>~\n                if Self != ""
Special\t<GetSpecialName>\n if SpecialMove > 0
Stance\t<name0> <name1>\n   if Prereq[0] > 0
Suscept\t<n>\n              if Suscept > 0
Wait\t<n>\n                 if Wait > 0
Weight\t<n>\n               if MaxWeight > 0
EndStance\n
\n                          (blank line)
```

After all stances: `End\n`.

**Tab separator — load-bearing.** C writes `\t`. Go's `ReadWord` in the scanner treats tab as whitespace so either works for round-trip, but the visible file format is tab-separated; preserve exactly.

### D8 — `persist.LoadStancesInto` extension

Modify `internal/persist/stances.go:113-146`. Replace the discard-only cases with real assignments:

```go
case "Attacks":
    target[idx].NumAttacks = sc.ReadNumber()
case "Class":
    target[idx].Class = sc.ReadNumber()
case "DamDone":
    target[idx].DamDone = sc.ReadNumber()
case "DamTaken":
    target[idx].DamTaken = sc.ReadNumber()
case "Dodge":
    target[idx].Dodge = sc.ReadNumber()
case "Dual":
    target[idx].Dual = sc.ReadNumber()
case "Immune":
    target[idx].Immune = sc.ReadNumber()
case "Other":
    target[idx].Others = sc.ReadString()
case "Parry":
    target[idx].Parry = sc.ReadNumber()
case "Percent":
    target[idx].SpecialPercent = sc.ReadNumber()
case "Race":
    target[idx].Race = sc.ReadNumber()
case "Resist":
    target[idx].Resist = sc.ReadNumber()
case "Self":
    target[idx].Self = sc.ReadString()
case "Special":
    specName := sc.ReadWord()
    target[idx].SpecialMove = combat.GetSpecialNumber(specName) // returns 0 in stock (stub)
case "Stance":
    s1 := sc.ReadWord()
    s2 := sc.ReadWord()
    if n := GetStanceNumber(s1); n >= 0 {
        target[idx].Prereq[0] = n
    }
    if n := GetStanceNumber(s2); n >= 0 {
        target[idx].Prereq[1] = n
    }
case "Suscept":
    target[idx].Suscept = sc.ReadNumber()
case "Wait":
    target[idx].Wait = sc.ReadNumber()
case "Weight":
    target[idx].MaxWeight = sc.ReadNumber()
```

Update the doc comment at line 51-68 to reflect that the non-combat fields are now stored.

### D9 — Flag-string helpers

New helpers land in `util/flags.go` (or extend an existing file; verify at impl):

```go
// RisflagNames indexes the RIS_* bits by position. Index i corresponds to
// bit (1<<i). Mirrors C's ris_flags array.
var RisflagNames = []string{
    // bits 0-7
    "fire", "cold", "electricity", "energy", "blunt", "pierce", "slash", "acid",
    // bits 8-13
    "poison", "drain", "sleep", "charm", "hold", "nonmagic",
    // bits 14-19 (plus1..plus6)
    "plus1", "plus2", "plus3", "plus4", "plus5", "plus6",
    // bits 20-21 (corrected 2026-04-18 audit — constants.go:196-217 defines 22
    // RIS_* bits; terminal names are "magic" + "paralysis", not "plus7")
    "magic", "paralysis",
}

func GetRisflag(name string) int {
    for i, n := range RisflagNames {
        if strings.EqualFold(n, name) {
            return i
        }
    }
    return -1
}

func FlagString(flags int, names []string) string {
    var parts []string
    for i, name := range names {
        if flags & (1 << i) != 0 {
            parts = append(parts, name)
        }
    }
    if len(parts) == 0 {
        return "none"
    }
    return strings.Join(parts, " ")
}
```

### D10 — `combat.GetStanceName` / `GetStanceMastery` / `GetSpecialName` / `GetSpecialNumber`

Small conveniences in `combat/stance_index.go`:

```go
func GetStanceName(idx int) string {
    switch idx {
    case types.STANCE_TIGER:    return "Tiger"
    case types.STANCE_SWALLOW:  return "Swallow"
    case types.STANCE_DRAGON:   return "Dragon"
    case types.STANCE_MONKEY:   return "Monkey"
    case types.STANCE_MANTIS:   return "Mantis"
    case types.STANCE_VIPER:    return "Viper"
    case types.STANCE_CRANE:    return "Crane"
    case types.STANCE_CRAB:     return "Crab"
    case types.STANCE_MONGOOSE: return "Mongoose"
    case types.STANCE_BULL:     return "Bull"
    case types.STANCE_NORMAL:   return "Normal"
    case types.STANCE_NONE:     return "None"
    default:                    return ""
    }
}

func GetStanceMastery(ch *types.CharData) int {
    if ch.Stance == types.STANCE_NONE {
        return 0
    }
    if ch.IsNPC() {
        if ch.IndexData == nil {
            return 0
        }
        return ch.IndexData.Stances[ch.Stance]
    }
    if ch.PCData == nil {
        return 0
    }
    return ch.PCData.Stances[ch.Stance]
}

func GetSpecialName(idx int) string { return "None" } // C stub: stances.c:1018-1022
func GetSpecialNumber(name string) int { return 0 }    // C stub: stances.c:1012-1016
```

### D11 — Command registration

Extend `internal/boot/boot.go` in the admin-command section (search for existing `Register` calls for `mset`):

```go
reg.Register(&command.Command{Name: "ststat", DoFun: act.DoSTstat, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
reg.Register(&command.Command{Name: "stset",  DoFun: act.DoSTset,  Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
```

`DoStance` is already registered (Phase 5 Tier 6 via `skills4.go` path); verify by reading `boot.go` — if absent, add `"stance"` at a player-accessible level. Position `POS_STANDING` for `stance` (fighting stance cannot be set while resting — but C `do_stance` does NOT check position, it delegates to `can_use_stance` for gates, so register at `POS_DEAD` to let the command body do its own gating per C fidelity).

---

## Task Groups

Five task groups, landed in dependency order. G1 is the schema + reader (prerequisite for everything); G2 is the core gameplay (`CanUseStance` + `UpdateStances` + rewritten `DoStance`); G3 is the writer + `stset save` round-trip; G4 is `DoSTstat` + `DoSTset` OLC; G5 is boot wiring + final integration tests. Each group ships test-first with mutation verification.

**Mutation-verification safety.** All mutation reverts use `Edit`-only. Banned: `git checkout -- <file>`, `git checkout <ref> -- <file>`, `git restore <file>`, `git reset --hard` (any form), `git stash` (any form). Reference: `_shared.md` → Mutation Verification Safety.

### G1 — `StanceInfo` extension + reader upgrade + helpers (M, ~3h)

**Deliverables:**
1. Extend `StanceInfo` per D1 in `internal/combat/stance_index.go`.
2. Add `GetStanceName`, `GetStanceMastery`, `GetSpecialName`, `GetSpecialNumber` per D10 in the same file.
3. Extend `LoadStancesInto` per D8 in `internal/persist/stances.go`.
4. Update the loader doc comment to reflect stored (not discarded) non-combat fields.
5. Add `testdata/stances_full.dat` covering all 18 key types for a single stance plus a minimal second stance.
6. New tests:
   - `TestStanceInfo_ZeroValueCompatibleWithDefaults` — build the default `StanceIndex` table; assert the three combat fields match existing values; assert new fields are zero.
   - `TestLoadStances_FullFieldRoundTripLoad` — load `stances_full.dat`; assert every extended field lands in the right slot (Dodge, Parry, Dual, MaxWeight, Wait, Resist, Immune, Suscept, Class, Race, SpecialMove, SpecialPercent, Prereq[0]/[1], Self, Others).
   - `TestGetStanceName_Exhaustive` — all 12 constants map to their canonical names; `MAX_STANCE` returns `""`.
   - `TestGetStanceMastery_NPC_UsesIndexData` — NPC with `IndexData.Stances[DRAGON]=150`; `GetStanceMastery` returns 150 when `ch.Stance = DRAGON`.
   - `TestGetStanceMastery_PC_UsesPCData` — PC analogue.
   - `TestGetStanceMastery_NoneReturnsZero` — `ch.Stance = STANCE_NONE` → 0 regardless of `.Stances` contents.
   - `TestGetStanceMastery_NilIndexDataNilPCDataSafe` — NPC with nil IndexData; PC with nil PCData — returns 0, no panic.
7. Existing combat tests continue to pass (`TestOneHit_StanceDamDoneMultiplier` etc.) — no behaviour change in the hot path.

**Mutation-verification:**
- Drop the `Dodge` case in the loader → `TestLoadStances_FullFieldRoundTripLoad` fails on Dodge assertion. `Edit`-revert.
- Flip `ch.Stance == STANCE_NONE` in `GetStanceMastery` to `!= STANCE_NONE` → `TestGetStanceMastery_NoneReturnsZero` fails.
- Remove the `nil IndexData` guard → `TestGetStanceMastery_NilIndexDataNilPCDataSafe` panics.

**File paths:**
- `internal/combat/stance_index.go` (modify struct + add helpers)
- `internal/combat/stance_helpers_test.go` (new)
- `internal/persist/stances.go` (extend loader)
- `internal/persist/stances_test.go` (extend)
- `internal/persist/testdata/stances_full.dat` (new fixture)

### G2 — `CanUseStance` + `UpdateStances` + `DoStance` rewrite (L, ~4h)

**Deliverables:**
1. `combat.CanUseStance` per D2 in `internal/combat/can_use_stance.go`.
2. `combat.UpdateStances` per D3 in `internal/combat/stance_update.go`.
3. Rewrite `DoStance` at `internal/act/skills4.go:180-199` per D4.
4. `sendStanceMessage` helper in `skills4.go` (or a new `act/stance.go` if `skills4.go` is crowded — implementation choice).
5. `maxInt` helper — prefer a package-local function over adding to `util` in this plan; keep the surface small.
6. New/updated tests:
   - `TestCanUseStance_NPC_IndexDataStancesPositive_TRUE` — NPC with `IndexData.Stances[DRAGON]=1`.
   - `TestCanUseStance_NPC_IndexDataStancesZero_FALSE`.
   - `TestCanUseStance_NPC_NilIndexData_FALSE` (defensive).
   - `TestCanUseStance_PC_NoPrereq_TRUE` — `StanceIndex[DRAGON].Prereq = [0,0]`.
   - `TestCanUseStance_PC_Prereq0_MasteredAt200_TRUE` — `Prereq[0] = TIGER`; `PCData.Stances[TIGER] = 200`.
   - `TestCanUseStance_PC_Prereq0_Below200_FALSE` — `PCData.Stances[TIGER] = 199`.
   - `TestCanUseStance_PC_Prereq0and1_BothMastered_TRUE`.
   - `TestCanUseStance_PC_Prereq0Mastered_Prereq1NotMastered_FALSE`.
   - `TestCanUseStance_MaxWeightExceeded_FALSE` — `StanceIndex[i].MaxWeight=100`, `ch.CarryWeight=200`.
   - `TestCanUseStance_MaxWeightEqual_TRUE_BoundaryOff` — `ch.CarryWeight == MaxWeight` → pass (C uses `>`, not `>=`).
   - `TestCanUseStance_DualForbidWithDualWielded_FALSE` — `StanceIndex[i].Dual=1`, `ch.Equipment[WEAR_DUAL_WIELD]` non-nil.
   - `TestCanUseStance_DualForbidWithoutDualWielded_TRUE` — `.Dual=1` but no equip.
   - `TestCanUseStance_DualAllowWithDualWielded_TRUE` — `.Dual=0`.
   - `TestCanUseStance_ClassRestriction_CBug_PreservesMaskANDIndex` — explicitly pin the broken behaviour: with `Class = 1` and `ch.Class = 1` → fires (`1 & 1 != 0`). With `Class = 2` and `ch.Class = 1` → does NOT fire (`2 & 1 == 0`).
   - `TestUpdateStances_EnteringAppliesResistImmuneSuscept`.
   - `TestUpdateStances_LeavingClearsResistImmuneSuscept_Idempotent` — enter, then leave, then leave again → bits remain cleared; no double-clear panic.
   - `TestUpdateStances_OutOfRangeSafe` — `ch.Stance = MAX_STANCE+1` → no-op.
   - `TestDoStance_MountedRejected` — `ch.Mount != nil` → mount message, no state change.
   - `TestDoStance_NoArg_FromNone_EntersNormal` — `ch.Stance=STANCE_NONE`, `DoStance(ch, "")` → `ch.Stance=STANCE_NORMAL`; `UpdateStances` applied; `ch.Wait >= StanceIndex[STANCE_NONE].Wait`.
   - `TestDoStance_NoArg_FromNonNone_ExitsToNone` — set `ch.Stance=STANCE_DRAGON`; call; post: `ch.Stance=STANCE_NONE`; stance RIS bits cleared.
   - `TestDoStance_ChangeWhileAlreadyInStance_Rejected` — `ch.Stance=STANCE_DRAGON`, `DoStance(ch, "tiger")` → "cannot change" message; `ch.Stance` unchanged.
   - `TestDoStance_Unknown_ShowsSyntax` — `DoStance(ch, "nonesuch")` → syntax block.
   - `TestDoStance_CanUseBlocked_ShowsSyntax` — prereq not met; `DoStance(ch, "tiger")` → syntax block.
   - `TestDoStance_SuccessfulEntry_SetsStance_UpdatesRIS_SetsWait`.

**Mutation-verification:**
- Flip `ch.CarryWeight > MaxWeight` to `>= MaxWeight` → `TestCanUseStance_MaxWeightEqual_TRUE_BoundaryOff` fails.
- Remove the `Stance > STANCE_NONE` change-while-active gate → `TestDoStance_ChangeWhileAlreadyInStance_Rejected` fails.
- Flip `UpdateStances` `|=` / `&^=` → round-trip test diverges.

**File paths:**
- `internal/combat/can_use_stance.go` (new)
- `internal/combat/can_use_stance_test.go` (new)
- `internal/combat/stance_update.go` (new)
- `internal/combat/stance_update_test.go` (new)
- `internal/act/skills4.go` (modify DoStance + helper)
- `internal/act/skills4_test.go` (replace/extend `TestDoStance_*` cases)

### G3 — `SaveStances` writer + round-trip (M, ~2.5h)

**Deliverables:**
1. `persist.SaveStances` per D7 in `internal/persist/stances_save.go`.
2. New tests:
   - `TestSaveStances_EmptyTable_WritesHeaderAndEnd` — all default; emits `StartStance\tNone\nEndStance\n\n...End\n` with no optional keys.
   - `TestSaveStances_FullyPopulatedStance_EmitsAllKeys` — populate all 18 keys on one stance; assert each appears in output with correct tab-key-value formatting.
   - `TestSaveStances_ConditionalWriteInvariants` — pin each of the 18 invariants (e.g., `Class=0 → not emitted`, `Class=1 → emitted`; `Dodge=-5 → emitted because != 0`; `Dam_done=0 → not emitted because > 0 threshold`; etc.). One test per invariant; data-driven using a table.
   - `TestSaveStances_RoundTripMatchesLoad` — load `stances_full.dat`, immediately save to a `bytes.Buffer`, load the buffer into a fresh `[MAX_STANCE]StanceInfo`, assert `reflect.DeepEqual` to the pre-save table.
   - `TestSaveStances_StancePreqNames_BothWritten` — `Prereq=[TIGER, DRAGON]` → writes `Stance\tTiger Dragon\n`.
   - `TestSaveStances_StancePreqSecondaryZero_WritesNone` — `Prereq=[TIGER, 0]` → writes `Stance\tTiger None\n` (matches C feeding 0 to `get_stance_name` yielding "None").
3. `stset save` integration test (in G4 stanza but forward-referenced here): after `DoSTset(ch, "save")` the file at the data path matches the in-memory table via `Load+DeepEqual`.

**Mutation-verification:**
- Flip `DamDone > 0` to `DamDone != 0` in the writer → `TestSaveStances_ConditionalWriteInvariants` detects because a `DamDone = -1` fixture flips from "not written" to "written".
- Skip the `EndStance\n\n` trailing blank line → round-trip fails on the byte-exact comparison.

**File paths:**
- `internal/persist/stances_save.go` (new)
- `internal/persist/stances_save_test.go` (new)

### G4 — `DoSTstat` + `DoSTset` + flag helpers (L, ~4h)

**Deliverables:**
1. `util.RisflagNames`, `util.GetRisflag`, `util.FlagString` per D9 in a new `internal/util/flags.go` (or extend an existing file; check at impl).
2. `DoSTstat` per D5 in `internal/act/stances_admin.go`.
3. `DoSTset` per D6 in same file.
4. Tests:
   - **DoSTstat:**
     - `TestDoSTstat_NoArg_ListsAllStances` — output contains all 12 stance names.
     - `TestDoSTstat_UnknownName_ShowsErrorAndHelp` — `DoSTstat(ch, "nonesuch")` → `"That stance does not exist."` in output.
     - `TestDoSTstat_KnownStance_PrintsAllFields` — populate one stance with recognizable values; assert each of the 13 printed lines appears with the right value (regex match per line).
   - **DoSTset:**
     - `TestDoSTset_NPCRejected` — NPC caller → `"Mob's can't mset\n\r"`.
     - `TestDoSTset_NoDesc_Rejected` — `ch.Desc == nil` → no-desc message.
     - `TestDoSTset_Save_InvokesWriter` — `DoSTset(ch, "save")` → output contains `"Done."` AND `SaveStances` was called with the expected path (use a test hook or read-back via Load and DeepEqual).
     - `TestDoSTset_NoArgs_ShowsUsage` — usage block printed.
     - `TestDoSTset_UnknownStance_ShowsError`.
     - `TestDoSTset_Attacks_InRange` — `DoSTset(ch, "dragon attacks 2")` → `StanceIndex[DRAGON].NumAttacks == 2`.
     - `TestDoSTset_Attacks_OutOfRange_Rejected` — `attacks 10` → range message; no change.
     - `TestDoSTset_Damage_InRange` / `_OutOfRange`.
     - `TestDoSTset_Dodge_Signed_InRange` — `-50..50`.
     - `TestDoSTset_Dual_CBugPreserved` — `dual 1` → `.Dual == 0`. `dual 0` → `.Dual == 1`. Pin the C bug.
     - `TestDoSTset_Class_NoOp_CBugPreserved` — `class 15` → `.Class` unchanged (= previous value).
     - `TestDoSTset_Race_NoOp_CBugPreserved`.
     - `TestDoSTset_Immune_TogglesBit` — `immune fire` → bit 0 flipped.
     - `TestDoSTset_Immune_UnknownFlag_Rejected` — `immune bogus` → "Unkown flag" (typo preserved).
     - `TestDoSTset_Others_Self_StoresString` — tilde-smashed.
     - `TestDoSTset_Parry_InRange`.
     - `TestDoSTset_Percent_InRange` (0..100).
     - `TestDoSTset_Protection_SetsDamTaken` — `protection 80` → `.DamTaken == 80`.
     - `TestDoSTset_Resist_TogglesBit`.
     - `TestDoSTset_Special_InvokesStub_SetsSpecialMoveToZero` — pin the stub behaviour.
     - `TestDoSTset_Susceptible_TogglesBit`.
     - `TestDoSTset_Stance1_SetsPrereq0`.
     - `TestDoSTset_Stance1_InvalidName_Rejected`.
     - `TestDoSTset_Stance2_SetsPrereq1`.
     - `TestDoSTset_Wait_InRange`.
     - `TestDoSTset_Weight_InRange`.
     - `TestDoSTset_UnknownField_ShowsUsage`.
5. Command registration in `boot.go` (moved to G5 for atomicity).

**Mutation-verification:**
- Flip the `value < -5 || value > 5` guard in `attacks` to `value < -4 || value > 4` → `TestDoSTset_Attacks_OutOfRange_Rejected` fails at the boundary (5).
- Remove the inverted `dual` assignment → `TestDoSTset_Dual_CBugPreserved` fails.
- Remove the `immune` bit-toggle → `TestDoSTset_Immune_TogglesBit` fails.

**File paths:**
- `internal/util/flags.go` (new)
- `internal/util/flags_test.go` (new)
- `internal/act/stances_admin.go` (new)
- `internal/act/stances_admin_test.go` (new)

### G5 — Boot wiring + end-to-end integration (S, ~1.5h)

**Deliverables:**
1. Register `"ststat"` and `"stset"` commands in `internal/boot/boot.go` at `LEVEL_IMMORTAL`.
2. Verify `"stance"` is already registered; if missing, add at `POS_DEAD` / `LEVEL_HERO` (player-accessible; C `do_stance` has no trust gate).
3. New integration test in `internal/boot/boot_test.go`:
   - `TestBoot_StancesOLCRegistered` — resolves `"ststat"`, `"stset"`, `"stance"` command names; asserts levels and non-nil `DoFun`.
4. New scenario test in `internal/act/stances_admin_test.go`:
   - `TestScenario_StsetEditThenSaveThenReload` — runs `DoSTset(admin, "dragon attacks 5")`, `DoSTset(admin, "save")`, reloads via `LoadStancesInto(&fresh, path)`, asserts `fresh[DRAGON].NumAttacks == 5`. Uses `t.TempDir()` for the data path.
   - `TestScenario_MortalCannotChangeStanceWhileActive` — enter DRAGON, try to switch to TIGER; rejected with the "cannot change" message.
   - `TestScenario_MortalCanUseAfterGMMastery` — seed `pc.PCData.Stances[TIGER] = 200`; set `StanceIndex[DRAGON].Prereq[0] = TIGER`; `DoStance(pc, "dragon")` succeeds.
   - `TestScenario_MortalCannotUseWithoutGMMastery` — seed `Stances[TIGER] = 199`; fails.

**Mutation-verification:**
- Register `ststat` at `LEVEL_HERO` instead of `LEVEL_IMMORTAL` → `TestBoot_StancesOLCRegistered` fails on level assertion.
- Disable save→load round-trip (e.g., write to wrong path) → `TestScenario_StsetEditThenSaveThenReload` fails.

**File paths:**
- `internal/boot/boot.go` (add two `reg.Register` calls)
- `internal/boot/boot_test.go` (add stances-OLC registration test)
- `internal/act/stances_admin_test.go` (extend with scenario tests)

---

## Acceptance Criteria

**G1 (struct + reader + helpers):**
- A1. `combat.StanceInfo` has all 19 C-cognate fields (3 combat + 16 OLC-facing) listed in §D1.
- A2. `persist.LoadStancesInto` stores every key from C's `fread_stance` (no silent drop).
- A3. `combat.GetStanceName` is exhaustive (12 cases + default empty).
- A4. `combat.GetStanceMastery` returns zero safely when `IndexData` or `PCData` is nil.

**G2 (gameplay):**
- A5. `combat.CanUseStance` implements C's three PC prerequisite branches + four post-gates (max_weight, dual, class, race) verbatim.
- A6. `combat.UpdateStances` toggles stance RIS bits symmetrically on entry / exit.
- A7. `DoStance` rejects mounted players, ignores `arg==""` by toggling between NONE ↔ NORMAL, rejects change-while-active, delegates to `CanUseStance` for new-stance entry, emits `sendStanceMessage` correctly for both branches, and applies `WAIT_STATE` to `ch.Wait`.
- A8. `DoStance` nil-guards `Others` / `Self` empty strings (no bare `util.Act` call with `""`).

**G3 (writer):**
- A9. `persist.SaveStances` produces byte-faithful C `fwrite_stance` output including tab separators, `StartStance` / `EndStance` wrappers, blank-line separator between stances, terminal `End\n`.
- A10. All 18 conditional-write invariants match C (see D7 table) — pinned by individual tests.
- A11. Save→load round-trip on a populated fixture yields `reflect.DeepEqual` match.

**G4 (OLC commands):**
- A12. `DoSTstat` prints 13 per-stance fields (Name, Self, Other, Required Stances, Dual Wield, Damage Done+Taken, Dodge+Parry, Attacks+MaxWeight, Wait, Resist, Suscept, Immune, Special+Chance, Class Restr, Race Restr) matching C's layout intent.
- A13. `DoSTset` supports all 18 field keys with C-faithful range guards (`-5..5` attacks, `0..200` damage, `-50..50` dodge/parry, `0..1` dual, `0..100` percent, `0..200` protection, `0..25` wait, `0..999` weight).
- A14. `DoSTset` preserves the four C bugs (class/race setters no-op; dual inverted; special stub returns 0) with explicit inline comments referring to the C line numbers.
- A15. `stset save` calls `persist.SaveStances` and the result round-trips.

**G5 (wiring):**
- A16. `ststat` and `stset` registered at `LEVEL_IMMORTAL`; `stance` registered at player-accessible level with `POS_DEAD`.
- A17. Full `stset → save → reload` scenario works end-to-end in the boot-wired world.

**Global:**
- A18. `go vet ./...` clean.
- A19. `go test -count=3 ./...` green across all packages.
- A20. No regression in Tranche B tests (`go test ./internal/persist/... ./internal/combat/... ./internal/act/...`).

---

## Scope Cuts / Deferrals

- **Stance RIS consumption in combat resolution.** `CharData.StanceResistant/Immune/Susceptible` are populated by `UpdateStances` after this plan, but the combat R/I/S resolution code at `internal/combat/combat.go` does not yet consult them — that path reads only `ch.Resistant/Immune/Susceptible`. Wiring the stance bits into the final R/I/S check is a separate tranche; tracked in TODO.md as follow-up `stance-ris-wire`. This plan makes the data live; the next plan makes it consumed.
- **`randomize_stances` (C `stances.c:389-462`).** Timer-triggered randomizer for `dam_done` / `dam_taken` / `dodge` / `parry` / `max_weight` / `num_attacks` / resist/suscept. Not currently called from anywhere in C (the function exists but has no wiring outside manual invocation). Out of scope; future work if we ever discover the intended trigger.
- **`get_special_number` / `get_special_name` full implementation.** Both are stubs in C too (return constant 0 / `"None"`). Porting as stubs matches C; the full "special moves" feature is unfinished upstream.
- **`class_string` / `race_string` helpers used by `DoSTstat`.** Per §D5, `DoSTstat` prints the raw hex mask until richer formatters land (would require cross-package reads into the class/race tables). The C output is human-friendly names; ours is `0x%x` plus a comment. Follow-up: `DoSTstat` richer class/race display.
- **Class/race restriction setters.** C `stset class N` and `stset race N` are `found = TRUE` with no body — silent no-op. The Go port preserves this to match shipped behaviour (§Open Q2). A future tranche could add real setters, but stock builders expect today's behaviour.
- **`CMD_FLAG_NO_ABORT` etc. on `stance`.** C's command-interpreter treats some commands as "safe during skill wait-state"; stance is not one of them. No action needed.
- **Help file `STANCEFLAGS`.** Referenced in error paths (`do_help ch, "STANCEFLAGS"`). Go's help system is ported; adding a `STANCEFLAGS` help entry with a flag-name list is a minor content delta, not scoped here unless the help subsystem is already wired for builder additions (verify: `grep -n STANCEFLAGS db/system/help.are` at implementation time).

---

## Open Questions (with recommended answers)

### Q1 — Dead `MAX_STANCE` values via OLC

`DoSTset` writes to `StanceIndex[idx]` where `idx = GetStanceNumber(arg1)`. `GetStanceNumber("none")` returns `STANCE_NONE = 0`. Per C `do_stset`, an immortal can `stset none attacks 5` — writing into the `STANCE_NONE` slot. In C this is harmless because `STANCE_NONE` is never selected by `do_stance` (`ch->stance = STANCE_NONE` on exit, never as a new-stance entry). In Go the same holds — but it is surprising. **Recommended:** allow the write (preserves C behaviour); add a comment. Alternative: reject `idx == STANCE_NONE` with a message — diverges from C.

### Q2 — `stset class` / `stset race` setter bodies

Both branches in C set `found = TRUE` without assignment (`stances.c:823-826` / `:920-922`) — documented at D6 as "C bug preserved". **Recommended (default):** preserve no-op behaviour. Alternative: port `value = get_risflag(arg3); if value >= 0 { Class ^= 1<<value }` — but this invents gameplay not in C. Reject.

### Q3 — `stset dual` inverted assignment

C `stances.c:858-861`:
```c
if (value)
    stance_index[index].dual_wield = 0;
else
    stance_index[index].dual_wield = 1;
```
Reversed assignment — `dual 1` stores 0, `dual 0` stores 1. **Recommended:** preserve verbatim. Pin with `TestDoSTset_Dual_CBugPreserved`. Alternative: fix it silently — would drift from shipped area data expectations.

### Q4 — `class_restrictions` / `race_restrictions` bug in `CanUseStance`

C `stances.c:276-279` uses `IS_SET(mask, index)` where `index` is `ch->class` / `ch->race` — a plain index, not a bit. Effectively `mask & index`, which is almost always `0` unless `index == 1` AND `mask & 1 != 0`, or similar coincidences. **Recommended:** preserve the bug verbatim with a comment. Alternative: fix to `mask & (1<<index)`. Rejecting the fix because (a) no shipped stance populates class/race restrictions (the stub file is empty, and the `stset class/race` setters are no-ops), so the bug is observationally silent; (b) if a future builder populates via direct file edit with the fix-assumed semantics, they can then expect working restrictions — but the C lineage this port targets has never done this, so the fix would be a forward-ported gameplay change, not a fidelity repair.

### Q5 — `do_ststat` newline-every-4 bug

C `stances.c:687`: `if (index_num != 0 && (!index_num % 4))` — operator precedence makes it `(!index_num) % 4` which is always 0 (false) for `index_num != 0`. Intent was `(index_num % 4 == 0)`. **Recommended:** port the intent (fix). The C output-layout bug produces a single long line; fixing to a 4-per-row grid is a minor cosmetic improvement that doesn't change gameplay. Add a comment: "C bug fixed; see stances.c:687".

### Q6 — `RIS_*` flag name table

C `ris_flags` is a `const char *` array parallel to the `RIS_*` bit positions. The canonical table:
```
bit 0:  fire      bit 1:  cold       bit 2:  electricity  bit 3:  energy
bit 4:  blunt     bit 5:  pierce     bit 6:  slash        bit 7:  acid
bit 8:  poison    bit 9:  drain      bit 10: sleep        bit 11: charm
bit 12: hold      bit 13: nonmagic   bit 14: plus1        bit 15: plus2
bit 16: plus3     bit 17: plus4      bit 18: plus5        bit 19: plus6
bit 20: magic     bit 21: paralysis
```

**Verified 2026-04-18 audit against `internal/types/constants.go:196-217`:** 22 RIS_* bits defined (0..21). Terminal names are `magic` (bit 20) and `paralysis` (bit 21), **not** `plus7`. Pre-audit draft table was off by two names; corrected below. **Recommended:** derive the `RisflagNames` slice from the actual constant definitions and add a test that iterates every `RIS_*` constant and confirms `RisflagNames[log2(RIS_X)] == "<expected lowercase name>"` — catches any future drift if additional RIS bits land (some SMAUG variants ship bits up to 31).

### Q7 — `sendStanceMessage` empty-string nil-guard

C unconditionally calls `act` with `others` / `self` which may be NULL (`stances.c:293-295`), which segfaults or prints garbage in practice. **Recommended:** Go skips `util.Act` when the message is `""`. Alternative: call `util.Act` with `""` and let it no-op — verify `util.Act` handles empty format gracefully; if not, the skip is required. Either is safe; skip is more explicit.

### Q8 — `WorldRef.DataDir` field for `stset save` path

`stset save` needs to derive the path the boot loader used. Options:
- a. Read `WorldRef.DataDir` or equivalent field — need to verify it exists.
- b. Hardcode `"../db/system/stances.dat"` — brittle; breaks under non-default working directories.
- c. Store the path as a side effect of the loader call (e.g., `persist.StancePath string` set by `LoadStancesInto`).
- d. Thread the path through a `DataDir` function param into `DoSTset`.

**Recommended (d):** add a package-level `var StancePath string` in `internal/persist/stances.go` that `LoadStancesInto` sets from its `path` arg on first call; `DoSTset` reads it for `save`. Self-contained, no new world-struct plumbing. Alternatively, if `world.World` already carries `DataDir` (verify during G5), use `a` — one reference to a well-established field is idiomatic.

### Q9 — Register `"stance"` command if missing

**Resolved (2026-04-18 audit):** `grep -n '"stance"' internal/boot/boot.go` returns exactly one hit at `boot.go:521`: `reg.Register(&command.Command{Name: "stance", DoFun: act.DoStance, Position: types.POS_DEAD, Level: 0})`. Registration exists at `POS_DEAD` / Level 0 (player-accessible — C has no trust gate). G5 need only add `ststat` + `stset`; `stance` is already wired. No action needed for this question.

### Q10 — `maxInt` helper home

New tiny helper. **Recommended:** inline in `act/skills4.go` as an unexported `func maxInt(a, b int) int`. Alternative: drop into `util/numbers.go` if one exists. No strong preference; inline keeps the diff tight.

---

## Risk Analysis

**G1 (struct extension):** **Low.** Extending `StanceInfo` with zero-valued fields is Go-safe; the existing default table's named-field initialisers remain valid. Risk vector: a stale copy of `StanceInfo` elsewhere (e.g., a test-local struct with only three fields) would stop compiling. Mitigation: `go build ./...` after the struct change; fix each call site.

**G2 (CanUseStance + DoStance rewrite):** **Medium.** `DoStance` rewrite is the highest-behaviour-change delta in this plan. Existing `TestDoStance_Set` (at `skills4_test.go:110`) asserts `ch.Stance = STANCE_DRAGON` after `DoStance(ch, "dragon")` — the new `CanUseStance`-gated path will block this unless the test PC is seeded with a Dragon-prereq stance at GM level OR `StanceIndex[DRAGON].Prereq[0] = 0`. Mitigation: update existing Tier-6 tests to either use `Viper`/`Crane`/`Crab` (stances with no prerequisite) OR seed the PC's prereq mastery. Prefer the former to keep test intent clear.

**G3 (SaveStances):** **Low-Medium.** Byte-faithful writer is mechanical but fragile; a missed tab or trailing newline breaks round-trip. Mitigation: test pins byte-exact output via `bytes.Equal` against a golden buffer; round-trip test catches semantic drift.

**G4 (DoSTset + DoSTstat):** **Medium.** 18 field setters × range guards × C-bug-preservation = ~30 test cases. Bug-preservation tests are essential; without them, a future innocent fix would break C-fidelity silently. Mitigation: inline comments referencing C line numbers beside each quirk.

**G5 (integration):** **Low.** Boot wiring is boilerplate; scenario test exercises the realistic flow.

**Cross-cutting — C bug preservation.** Four preserved C bugs (class no-op, race no-op, dual inverted, class/race CanUseStance mask-AND-index). Each gets a dedicated test AND a source-comment. Risk: a future refactor "cleans up" one of the bugs and breaks a test — but the test's name contains `CBug` and the comment explains why. Future maintainers see the guardrail.

**Cross-cutting — test dependency on package vars.** Several tests mutate `StanceIndex` via `withStanceIndex`-style save/restore. Tranche B established the pattern at `internal/combat/stance_test.go:14-19`. Reuse that pattern for all new stance-table-mutating tests; do NOT touch the package var without restoring.

---

## Adversary-Resolved Concerns (2026-04-18 manager self-adversary pass)

Agent-based external adversary verification was unavailable for this session. A structured self-review was conducted against authoritative C sources and the Go codebase, every line-citation spot-verified via direct Read. External adversary pass is recommended before dispatch.

1. **Existing `DoStance` stub dispatch behaviour.** The current stub at `internal/act/skills4.go:180-199` uses `strings.HasPrefix(name, arg)` on a 7-name slice missing `viper/crane/crab/normal`. Players typing `stance viper` get "No such stance." today. Verified: the slice omission is a real gap, not a test fixture. Plan G2 replaces the entire body with the full C-cognate state machine, picking up `viper/crane/crab/normal` via `GetStanceNumber` (12 names).

2. **`ch.Mount` field.** Checked `internal/types/character.go` — `Mount *CharData` exists. `DoStance`'s mount-gate translates verbatim.

3. **`ch.Equipment` shape for dual-wield check.** Grep confirms `Equipment [MAX_WEAR]*ObjData` is the shape; `ch.Equipment[WEAR_DUAL_WIELD]` is an idiomatic direct lookup. No `handler.GetEqChar` import needed from `combat` (avoids a circular-import risk).

4. **RIS bit count.** The RIS constants at `types/constants.go:196-215` run to `RIS_PLUS6 = 1 << 19` visible in the excerpt; there may be more bits through 21 or higher. Plan Q6 flags the need to verify the full range at G4 impl.

5. **`util.Act` signature.** Tranche C landed the per-call AType form: `util.Act(aType int, format string, ch, vch *types.CharData, arg1, arg2 any, to int)`. Plan uses this signature.

6. **Wait-state idiom.** Inspected `internal/act/playercfg.go:30` (`ch.Wait = 2`) — direct-assignment pattern. C's `UMAX(ch->wait, val)` pattern is preserved via the inline `maxInt(ch.Wait, val)` to match SMAUG's "take the stricter of existing or new wait" semantics.

7. **Stance `Wait` on entry uses `StanceIndex[newStance].Wait`, exit uses `StanceIndex[STANCE_NORMAL].Wait`.** Verified against C lines 82/89/125 — asymmetric between enter-from-NONE and enter-to-named-stance, with exit always using NORMAL's wait. Preserved in D4.

8. **`stance[1] = 0 → get_stance_name(0) = "None"`.** Verified by reading C `get_stance_name`: case `STANCE_NONE` returns `"None"` (line 191). The Go `SaveStances` writer invariants at D7 preserve this: `Prereq[0] > 0 → write \"Stance\\t<name0> <name1>\\n\"` where `name1 = GetStanceName(Prereq[1])` which yields `"None"` when `Prereq[1] == 0`.

9. **Tranche B compatibility.** The current `LoadStancesInto` discards non-combat keys; this plan's G1 upgrades the same file. The Tranche B fixture `stances_empty.dat` / `stances_two.dat` continue to work — `TestLoadStances_EmptyFileKeepsDefaults` and `TestLoadStances_TwoStancesOverridesDefaults` rely only on `NumAttacks` / `DamDone` / `DamTaken`, which remain stored. Regression-safe.

10. **Phase 5 Tier 6 `TestDoStance_Set`.** Risk called out in G2 Risk Analysis. Mitigation: either change the test to use `"viper"` (no prereq if `StanceIndex[VIPER].Prereq[0] == 0` — which is the default) OR seed `pcdata.Stances[prereq]=200`. Plan prefers the former to keep the test's unit scope (DoStance's entry-ok path), and will spot-check at implementation time that `StanceIndex[VIPER].Prereq == [0,0]` in the default table.

11. **`act.WorldRef` usage.** `DoSTset` reaches for the save path via `WorldRef.DataDir` or the Q8-recommended package-level var. Either is acceptable; Q8 recommends the var to avoid coupling `persist.SaveStances` into `world.World`'s shape.

12. **Scanner `ReadString` vs `ReadToEOL` for `Self`/`Others`.** C `fread_string_nohash` reads a tilde-terminated string spanning newlines. Go `Scanner.ReadString` (per Tranche B's use at `persist/stances.go:135`) reads to `~`. Verified idiom compatible.

---

## Appendix A — Fixture `testdata/stances_full.dat`

Tab-separated to match C `fwrite_stance`:

```
StartStance	Dragon
Attacks	2
Class	7
DamDone	150
DamTaken	90
Dodge	10
Dual	1
Immune	5
Other	$n enters a fierce Dragon stance.~
Parry	5
Percent	25
Race	0
Resist	3
Self	You enter a fierce Dragon stance.~
Special	None
Stance	Tiger None
Suscept	2
Wait	12
Weight	500
EndStance

StartStance	Tiger
Attacks	1
DamDone	110
DamTaken	110
EndStance

End
```

This exercises all 18 keys on Dragon + verifies minimal two-key Tiger block still loads.

---

## Appendix B — Commit phasing (suggested)

1. G1 commit: "Phase 6 Stances OLC G1: StanceInfo extension + loader upgrade + helpers".
2. G2 commit: "Phase 6 Stances OLC G2: CanUseStance + UpdateStances + DoStance rewrite".
3. G3 commit: "Phase 6 Stances OLC G3: SaveStances writer + round-trip".
4. G4 commit: "Phase 6 Stances OLC G4: DoSTstat + DoSTset OLC commands".
5. G5 commit: "Phase 6 Stances OLC G5: boot registration + end-to-end scenario tests".

Single-PR delivery; five commits for review clarity. The orchestrator may squash if desired.
