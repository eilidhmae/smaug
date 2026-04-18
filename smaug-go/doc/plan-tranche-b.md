# Plan: Tranche B — Stances Completion + Timer Dispatch + Mudprog If-Check Wiring + Wordlist Match

**Status:** Planned (2026-04-18). Adversary-verified research: to be filled after plan adversary pass.
**Priority:** Mixed (see per-task-group notes). Collected here because all four work items are combat/mudprog completeness gaps whose infrastructure already exists in Go but whose data/dispatch glue is missing.
**Scope:** New files `internal/persist/stances.go` + `internal/persist/stances_test.go`, `internal/mudprog/wordlist.go` + `internal/mudprog/wordlist_test.go`, `internal/handler/timer_registry_test.go`. Modifications to `internal/handler/timer.go` (registry + expiry dispatch), `internal/act/olc_set.go` (`DoMset` stance branch), `internal/mudprog/ifcheck.go` (7 if-check bodies), `internal/mudprog/oprog.go` (OprogCommandTrigger), `internal/mudprog/rprog.go` (RprogCommandTrigger), `internal/boot/boot.go` (loader wire + timer-registry wire + seeded entries). Plus test additions in each touched package.

---

## Problem

Four separate completeness gaps with a shared theme: the Go port has the types, flags, and call sites wired, but a data-loader / dispatch / algorithm port is missing, so the feature is dead code in production.

### Gap 1 — `db/system/stances.dat` loader + PC practice-stance flow

`internal/combat/stance_index.go` hard-codes a 12-entry `StanceIndex` table with reasonable defaults (Tiger/Dragon aggressive, Crab defensive, Monkey suppressor, etc.). C loads this table from `db/system/stances.dat` via `load_stances` at `src/stances.c:311`. The shipped stub at `/home/eilidh/src/smaug/db/system/stances.dat` is literally `End\n` (4 bytes) — no stance data. The comment on `stance_index.go:7-11` explicitly flags this as the gap.

Separately, C's `can_use_stance` at `src/stances.c:237` and the grand-master bonus-attack loop at `src/fight.c:1058-1069` both read `ch->pcdata->stances[]` — a per-PC per-stance mastery counter (0 to `MAX_PC_STANCE = 200`). In Go, `PCData.Stances [MAX_STANCE]int` exists at `internal/types/pcdata.go:98` and is persisted (loaded at `internal/persist/player.go:135`), but nothing currently **increments** it outside admin commands. The combat-depth G7 code at `internal/combat/combat.go:324` gates the GM bonus on `ch.PCData.Stances[ch.Stance] >= STANCE_GRAND_MASTER` — so for every real player, that condition is `0 >= 200 = false`, and the GM bonus-attack code path is unreachable in production. It's only exercised by tests that inject the counter directly.

### Gap 2 — `TIMER_DO_FUN` callback dispatch

`internal/handler/timer.go` is feature-complete for storage, but `DecrementTimers` at `:117` drops expired timers silently without dispatching their `DoFun` string. C `src/fight.c:382-430` dispatches in two places:
1. **Mid-decrement intercept** (line 386-398): during combat, any `TIMER_DO_FUN` timer is fired IMMEDIATELY with `ch->substate = SUB_TIMER_DO_ABORT` — this aborts skills like `detrap`, `dig`, `mend`, spellcasting when the PC enters combat. After dispatch the timer is extracted regardless.
2. **On expiry** (line 415-427): when `--timer->count <= 0` on a `TIMER_DO_FUN`, save substate, set it to `timer->value`, invoke the do_fun, restore substate, extract.

C also calls `TIMER_DO_FUN` dispatch from the command interpreter at `src/interp.c:713-732`: if the player has a pending `TIMER_DO_FUN` and issues any command that isn't `CMD_FLAG_NO_ABORT`, the timer is invoked with `SUB_TIMER_DO_ABORT` to cancel the skill.

Callers that set a `TIMER_DO_FUN`: `src/skills.c:1965` (`do_detrap`), `:2104` (`do_dig`), `:2265` (`do_search`), `:2934` (`do_mend`), `:3045` (`do_reading`), `src/magic.c:1858` (`do_cast`). None of these commands are ported to Go yet (the mend/dig/search/detrap/reading skills are placeholders) — so the intercept path and expiry-dispatch path are both currently dead. BUT: the moment any of those skills lands (e.g., `plan-skills-timer.md` in a future tranche), the dispatch must exist. Shipping the registry + expiry dispatch now unblocks those.

### Gap 3 — Mudprog if-check bodies

`internal/mudprog/ifcheck.go:885-890` has placeholder comments for: `timeskilled`, `leverpos`, `isflagged`, `istagged`, `pkadrenalized`, `asupressed`, `areamulti`, `multi`, `objtype`.

**Now unblocked** (because of Tranche B's other work + pre-existing types):
- `timeskilled` — wants `pMob->killed` counter. Already exists: `MobIndexData.Killed int` at `types/mob_index.go:21`.
- `objtype` — wants `obj->item_type` compare. Trivial: `chkObj.IndexData.ItemType`.
- `leverpos` — wants `IS_SET(obj->value[0], TRIG_UP)`. Already defined: `TRIG_UP uint32 = 1 << 0` at `constants.go:525`, and `ITEM_SWITCH`/`ITEM_LEVER`/`ITEM_PULLCHAIN` exist at `enums.go:608-610`. **Note:** the C body at `mud_prog.c:1493-1495` has a latent bug (`!=` chain with `||`) that makes the guard tautologically true — we port the *intent* (must be switch/lever/pullchain) via `&&` with `||` of positive matches, not the buggy C.
- `pkadrenalized` — wants `get_timer(chkchar, TIMER_RECENTFIGHT)`. `handler.GetTimer` subsystem landed in plan-timer-subsystem.md.
- `asupressed` — wants `get_timer(chkchar, TIMER_ASUPRESSED)`. Same subsystem.

**Independent of Tranche B but in scope**:
- `areamulti` — wants per-descriptor host comparison across chars in the same area. Verified: `DescriptorData.Host string` at `internal/types/descriptor.go:27`.
- `multi` — same as `areamulti` but world-wide. Same requirement.

**Still blocked** (explicitly excluded):
- `isflagged` / `istagged` — C uses `get_tag(chkchar, argv[2], vnum)` which walks `ch->variables` / `VARIABLE_DATA`. Go has `VariableData` + `CharData.Variables []*VariableData` but no `GetTag` implementation. Requires new code to port C `get_tag` at `src/mud_prog.c:2236+` — this is a full `variables.go` subsystem, out of scope for Tranche B.

### Gap 4 — `OprogCommandTrigger` / `RprogCommandTrigger` — full wordlist match

`internal/mudprog/oprog.go:216-221` stubs to `return false`. Its comment (`:211-215`) incorrectly claims `MPROG_CMD` does not yet exist; the bit DOES exist at `types/mudprog.go:90` (`MPROG_CMD = 1 << 37`). So the stub is simply a deferred port.

`internal/mudprog/rprog.go:143-183` is a partial port — matches first-word-of-command against space-separated keywords in the arglist. This is incorrect vs. C `rprog_wordlist_check` at `src/mud_prog.c:4109-4178`, which does a full substring search of each keyword over the entire command string with word-boundary checks (`start == dupl || *(start - 1) == ' '`) and tail-character checks (`== ' ' || '\n' || '\r' || '\0'`). It also supports a `"p "` prefix for phrase match where the remaining arglist is treated as ONE literal phrase.

The current Go Rprog matching is therefore BOTH narrower (only the first word of input) AND broader (keyword match without word-boundary check) than C. It works by accident for simple cases but diverges in load-bearing content scripts.

## C Reference (authoritative)

### Stances
- **`load_stances`** — `src/stances.c:311-382`. Opens `STANCE_FILE`, zeroes `stance_index[]`, then reads tokens: `StartStance <name>` begins a block; `EndStance` closes it; `End` terminates the file. Inside a block, key-value pairs like `Attacks N`, `Class N`, `DamDone N`, `DamTaken N`, `Dodge N`, `Dual N`, `Immune N`, `Other <str>~`, `Parry N`, `Percent N`, `Race N`, `Resist N`, `Self <str>~`, `Special <name>`, `Stance <name> <name>`, `Suscept N`, `Wait N`, `Weight N`.
- **`fread_stance`** — `src/stances.c:469-593`. Reads one block.
- **`fwrite_stance`** — `src/stances.c:599-670`. Writes all blocks.
- **`get_stance_number`** — `src/stances.c:199-232`. Case-insensitive lookup: `Tiger`, `Swallow`, `Dragon`, `Monkey`, `Mantis`, `Viper`, `Crane`, `Crab`, `Mongoose`, `Bull`, `None`, `Normal`.
- **`can_use_stance`** — `src/stances.c:237-281`. NPC: requires `pIndexData->stances[new_stance] > 0`. PC: requires prerequisite `stance_index[new_stance].stance[0]` to be grand-master (>= 200) — see struct field.
- **`do_stance`** — `src/stances.c:51-128`. User-facing `stance <name>` — gates on mount, then requires `can_use_stance`.
- **`do_stset`** — `src/stances.c:742-1010`. Admin command `stset <stance> <field> <value>`. Immortal only.
- **`get_stance_mastery`** — `src/stances.c:151-160`. For NPC returns `pIndexData->stances[stance]`; for PC returns `pcdata->stances[stance]`.
- **Practice-stance flow (admin only):** `src/build.c:3499-3537` inside `do_mset`. `mset <victim> <stance-name> <value>` sets `pcdata->stances[temp_num] = value` (PC) or `pIndexData->stances[temp_num] = value` (NPC). Values clamped `[0, MAX_PC_STANCE=200]` / `[0, MAX_MOB_STANCE=200]`. Requires `LEVEL_LESSER` trust.
- **Important:** C has NO `practice stance <name>` grinding command. Mastery is incremented via `mset` / `stset` by immortals. PCs cannot grind their own stance mastery. `do_stset` also has a `save` subcommand that calls `fwrite_stance`.

### Timer DO_FUN dispatch
- **Decrement-loop intercept** (`src/fight.c:382-430`): for each timer:
  1. If `ch->fighting && timer->type == TIMER_DO_FUN` → save substate, set `SUB_TIMER_DO_ABORT`, call `(timer->do_fun)(ch, "")`, restore, extract.
  2. Else decrement `--timer->count`. If count <= 0:
     - `TIMER_ASUPRESSED` with `timer->value == -1` → reset count to 1000, continue (permanent).
     - `TIMER_NUISANCE` → dispose `ch->pcdata->nuisance`.
     - `TIMER_DO_FUN` → save substate, set `ch->substate = timer->value` (NORMAL substate, not ABORT), call do_fun, restore. If timer->count > 0 after the call, continue (do_fun re-extended). Else extract.
- **Interp intercept** (`src/interp.c:713-733`): if `get_timerptr(ch, TIMER_DO_FUN)` is non-nil AND the new command is NOT `CMD_FLAG_NO_ABORT`, invoke do_fun with `SUB_TIMER_DO_ABORT`. If do_fun did NOT set `ch->substate = SUB_TIMER_CANT_ABORT`, extract the timer and continue normal command dispatch. Else return without dispatching the new command.
- **Callers (all in C)**: `src/skills.c:1965` (`do_detrap`, count 3), `:2104` (`do_dig`, count `UMIN(beats/10, 3)`), `:2265` (`do_search`), `:2934` (`do_mend`), `:3045` (`do_reading`), `src/magic.c:1858` (`do_cast`, count `UMIN(beats/10, 3)`). All use `ch` (PC) as the target — no NPC uses TIMER_DO_FUN.

### Mudprog if-checks
- `timeskilled` — `src/mud_prog.c:574-586`. Returns `mprog_veval(pMob->killed, opr, atoi(rval), mob)`. When `chkchar` is set, reads `chkchar->pIndexData->killed`; else resolves `atoi(cvar)` as a vnum and reads `get_mob_index(vnum)->killed`.
- `objtype` — `src/mud_prog.c:1486-1489`. `mprog_veval(chkobj->item_type, opr, atoi(rval), mob)`.
- `leverpos` — `src/mud_prog.c:1490-1503`. Requires `chkobj->item_type ∈ {ITEM_SWITCH, ITEM_LEVER, ITEM_PULLCHAIN}` (port *intent* — see Problem above about C's buggy guard). Reads `IS_SET(obj->value[0], TRIG_UP)`. Returns `mprog_veval(wantsup, opr, isup, mob)` where `wantsup = !str_cmp(rval, "up")`.
- `pkadrenalized` — `src/mud_prog.c:1431-1435`. `mprog_veval(get_timer(chkchar, TIMER_RECENTFIGHT), opr, atoi(rval), mob)`.
- `asupressed` — `src/mud_prog.c:1436-1440`. `mprog_veval(get_timer(chkchar, TIMER_ASUPRESSED), opr, atoi(rval), mob)`.
- `areamulti` — `src/mud_prog.c:1161-1181`. Counts distinct PCs in the same area with the same descriptor host as `chkchar`. `QUICKMATCH` is a null-safe `strcmp`.
- `multi` — `src/mud_prog.c:1182-1199`. Same but world-wide (no area gate).

### Wordlist match
- **`rprog_wordlist_check`** — `src/mud_prog.c:4109-4178`. For each room prog of the given type:
  1. Lower-case the arglist AND the input text.
  2. If arglist starts with `"p "`: remaining string is ONE phrase — search it as a single substring with word-boundary checks at start and end.
  3. Else: split arglist on whitespace, iterate each keyword, search each as a substring with word-boundary checks.
  4. Word-boundary check: `(start == dupl || *(start-1) == ' ')` AND `(end char == ' ' || '\n' || '\r' || '\0')`.
  5. On first match, call `mprog_driver`, mark executed, break out of the arglist loop (but continue through other progs? — C source shows `break` inside the keyword loop, then falls through to next prog; the loop-over-all-progs is NOT broken, so multiple progs CAN fire per call). Return executed.
- **`oprog_wordlist_check`** — `src/mud_prog.c:3813-3880`. Same algorithm but iterates `iobj->pIndexData->mudprogs`; wraps `mprog_driver` in `set_supermob(iobj)` / `release_supermob()`.
- **`mprog_wordlist_check`** — `src/mud_prog.c:2780-2845`. Same but over `mob->pIndexData->mudprogs`, no supermob wrapping.

## Go Current State

See Problem section. Key verified facts:
- `StanceIndex` is a compile-time table at `combat/stance_index.go:37-50`; tests mutate it via `withStanceIndex` helper (`combat/stance_test.go:14-19`).
- `PCData.Stances` and `CharData.Stances` both exist as `[MAX_STANCE]int`. `CharData.Stances` is an NPC-oriented copy (populated in `handler/handler.go:64` as `mob.Stances = idx.Stances`); `pIndexData.Stances` reads go through `ch.IndexData.Stances[stance]` at `combat/combat.go:613`.
- `handler.AddTimer` / `GetTimer` / `GetTimerPtr` / `RemoveTimer` / `ExtractTimer` / `DecrementTimers` all landed 2026-04-18. `DecrementTimers` drops expired timers with no dispatch (`timer.go:135` comment explicit).
- `DescriptorData.Host string` at `internal/types/descriptor.go:27` — populated from `net.SplitHostPort` on connect. Used by `areamulti` / `multi` if-checks.
- `MPROG_CMD` bit exists (`types/mudprog.go:90`); `OprogCommandTrigger` stubbed in `oprog.go:216-221`; `RprogCommandTrigger` partial in `rprog.go:143-183`.
- `firstWord` helper exists in `mudprog` package (`rprog.go:154` calls it).
- `types/constants.go:782` — `STANCE_GRAND_MASTER = 200`.
- `types/enums.go:20-33` — STANCE_* enum with `MAX_STANCE`.
- `types/enums.go:125` — `SUB_TIMER_DO_ABORT = 128`; `SUB_TIMER_CANT_ABORT = 129`.

## Go Design

### G1 — Stances data file loader

**Approach A (CHOSEN): Runtime replace of `StanceIndex`.** The existing `StanceIndex` array becomes the default fallback. New `LoadStances(path string) error` in a new file `internal/persist/stances.go` reads `db/system/stances.dat` with a `Scanner` and returns a populated `[MAX_STANCE]combat.StanceInfo` plus a name-mapped `map[int]StanceDetails` for non-combat fields (resist, immune, etc.) that will be consumed by the future `can_use_stance` / `do_ststat` work.

But the stance package needs to expose setters. Since `StanceIndex` is a package-level `var [MAX_STANCE]StanceInfo`, the loader writes into slots. `boot.Boot` calls `persist.LoadStancesInto(&combat.StanceIndex, path)` at startup. When `stances.dat` is empty or absent, the defaults remain untouched (the stub-file case is a no-op).

**Rejected alternatives:**
- **Approach B (full StanceTable struct with resist/immune/etc.):** The extra fields are only consumed by `do_ststat` / OLC, neither of which is in Tranche B scope. Defer the full struct until OLC stance editing ships in Phase 6.
- **Approach C (store in World.Stances):** `StanceIndex` is read in the combat hot loop. Indirection through `world.World` adds a pointer chase per attack. Package-var with a loader is the existing pattern for boot-time data (skills, classes, races).

### G1b — Expose PC stance mastery setter for admins

`DoStset` C admin command already has a shipped stub counterpart in Go: search finds none — so `stset` is not ported. Scope: (1) admin command `stset <victim> <stance> <value>` clamps `[0, MAX_PC_STANCE]` and writes `PCData.Stances[stance] = value` (PC) or `IndexData.Stances[stance] = value` (NPC). Requires trust >= LEVEL_LESSER for PC target. Mirrors `src/build.c:3499-3537` inside `do_mset`, not the full `do_stset` (which covers the stance-table edit — out of scope).

**Deliberate scope cut: no player-facing practice-stance grind.** C has no such command. `DoPractice` in Go practices skills, not stances. Admins seed mastery with `stset` (C) / `mset` (C+Go) — NOT grinding. The combat-depth plan's G7 open-question 2 ("PC practice-stance flow — either seed admins or add a practice command") is resolved here as: seed admins via `stset`. No new grind command.

**Rejected alternative:** A custom Go-only `practice stance <name>` that increments the counter by 1 per invocation. Temptation is to give players a way to reach GM stance. This would diverge from C-fidelity without a compelling reason. Users who want GM stance can be seeded by an immortal. Future Phase 6 content decision if needed.

### G2 — `TIMER_DO_FUN` callback dispatch

**Approach A (CHOSEN): Central registry in `handler` package.** New `handler.RegisterTimerFunc(name string, fn TimerFunc)` + `handler.LookupTimerFunc(name string) TimerFunc`. `TimerFunc` signature: `func(ch *types.CharData, argument string)` — same as `CmdFunc`. Registry populated at boot by `internal/boot/boot.go` with entries for the skills that currently ship as placeholders (no dispatch yet), plus a TODO note for each unported skill.

`DecrementTimers` gains an expiry-dispatch branch: if a timer expires and `t.DoFun != ""`, look up the function and invoke it with `ch.Substate` saved, set to `t.Value`, called, and restored. If `LookupTimerFunc` returns nil, log a `util.Bug` and drop silently (don't crash).

Mid-decrement intercept (combat-aborts-skill) is intentionally deferred — requires every skill command to be ported first. Plan explicitly documents the abort path as a scope cut (see Scope Cuts / Deferrals).

**Rejected alternatives:**
- **Approach B (per-package registration):** Each package (magic, skills) would own its timer names. Too much plumbing for what is effectively one map.
- **Approach C (closure stored directly on `TimerData`):** Breaks C-format persistence: `TIMER_PKILLED` persists the name-string via `PTimer` in save.c; closures can't serialize. Even though `TIMER_PKILLED` is the only persisted type today, keeping the name-string pattern preserves forward compatibility.

### G3 — Mudprog if-check bodies

Simple additions to the existing switch in `internal/mudprog/ifcheck.go`. Each if-check lands as a new `case "<name>"` returning `compareInt(...)` or `compareStr(...)` as appropriate. No new helpers required except:
- `objTypeOf(chkObj)` — already inlinable as `chkObj.IndexData.ItemType`.
- `leverIsUp(obj)` — `obj.Value[0] & types.TRIG_UP != 0`. `Value` is `[6]int`, existing.
- `sameHost(a, b)` for areamulti/multi — `a.Desc != nil && b.Desc != nil && a.Desc.Host == b.Desc.Host`. Exact field name to verify (see G3 open question).

### G4 — Wordlist match port

**Approach A (CHOSEN): Port C's word-boundary algorithm verbatim.** New `wordlistMatch(arglist, input string) bool` helper in `internal/mudprog/triggers.go` (or a new `wordlist.go` — see G4 question). Algorithm:

```go
func wordlistMatch(arglist, input string) bool {
    list := strings.ToLower(strings.TrimSpace(arglist))
    text := strings.ToLower(input)
    if list == "" { return false }
    // "p " prefix = literal phrase
    if strings.HasPrefix(list, "p ") {
        phrase := list[2:]
        return containsAtWordBoundary(text, phrase)
    }
    for _, kw := range strings.Fields(list) {
        if containsAtWordBoundary(text, kw) {
            return true
        }
    }
    return false
}

func containsAtWordBoundary(text, needle string) bool {
    if needle == "" { return false }
    idx := 0
    for idx < len(text) {
        found := strings.Index(text[idx:], needle)
        if found < 0 { return false }
        start := idx + found
        // Left word boundary: start of text or preceded by space
        leftOk := start == 0 || text[start-1] == ' '
        // Right word boundary: end of text or followed by ' ', '\n', '\r'
        end := start + len(needle)
        rightOk := end == len(text) || text[end] == ' ' || text[end] == '\n' || text[end] == '\r'
        if leftOk && rightOk {
            return true
        }
        idx = start + 1 // advance past this non-boundary match
    }
    return false
}
```

Replace the Rprog stub at `rprog.go:143-183` with this helper call. Replace the Oprog stub at `oprog.go:216-221` with the iterate-ch-room-obj + iterate-ch-inventory pattern matching C `oprog_command_trigger` at `src/mud_prog.c:3569-3580`.

**Rejected alternative:** Regex-based. Regex compilation per call is expensive; the command interpreter runs per-keystroke.

## Task Groups

### G1 — Stances data loader

**Files:**
- New: `internal/persist/stances.go`, `internal/persist/stances_test.go`
- Modified: `internal/combat/stance_index.go` (export a setter func; remove `var` default — see G1 decision), `internal/boot/boot.go` (call loader)
- Testdata: `internal/persist/testdata/stances.dat` — a minimal 2-stance fixture for loader tests

**Test-first:**
- `TestLoadStances_EmptyFileKeepsDefaults` — load a file containing only `End\n`; assert defaults match the existing hard-coded table.
- `TestLoadStances_TwoStancesOverridesDefaults` — fixture with `StartStance Dragon\n Attacks 2\n DamDone 150\n EndStance\n End\n`, assert `StanceIndex[STANCE_DRAGON].NumAttacks == 2` and `DamDone == 150`. Unmentioned fields retain defaults.
- `TestLoadStances_UnknownKeywordIgnored` — file with `StartStance Tiger\n BogusKey 42\n EndStance\n End\n` — loader logs a bug via `util.Bug` and continues (matches C `Fread_stance: no match` path).
- `TestLoadStances_BadStanceNameSkipsBlock` — `StartStance Nonesuch\n Attacks 5\n EndStance\n End\n` — no stance overridden, bug logged.
- `TestLoadStances_FileMissingIsNotFatal` — path doesn't exist — loader returns nil error, defaults intact.
- `TestGetStanceNumber` — case-insensitive `get_stance_number` matches C exactly for all 12 names; returns -1 for unknown.

**Mutation-verify expectations (via `Edit` round-trips only — BANNED for mutation revert: `git checkout -- <file>`, `git checkout <ref> -- <file>`, `git restore <file>`, `git reset --hard` any form, `git stash` any form. Apply the mutation with `Edit`; run test; confirm failure; call `Edit` again with the opposite change to revert):**
- Drop the `Attacks` key handler → `TestLoadStances_TwoStancesOverridesDefaults` fails on NumAttacks.
- Drop `DamDone` → same test fails on dam_done.
- Swap `StanceIndex[STANCE_DRAGON]` write to `[STANCE_TIGER]` → same test fails (dragon unchanged, tiger polluted).

**Note:** `StanceIndex` is a package-level `var [MAX_STANCE]StanceInfo`. The loader mutates entries in place. Tests that stub `StanceIndex[DRAGON]` via `withStanceIndex` (combat/stance_test.go) already use save/restore — keep that pattern.

**Acceptance gate:** Loader populates `StanceIndex` at boot without breaking existing combat tests. `go test -count=3 ./internal/combat/... ./internal/persist/...` green.

### G1b — Extend `DoMset` with stance-name field (C-parity path)

**Decision:** Include because the C practice-stance flow REQUIRES this command path (`mset <victim> <stance-name> <value>` is the sole path in C to increment `pcdata->stances[]`). Without G1b, G7's GM bonus-attack loop remains unreachable for real players. Tranche B ships only the `do_mset` branch at `src/build.c:3499-3537`, NOT the full `do_stset` at `src/stances.c:742` (which also edits the stance-table metadata — damage/dodge/parry/resist/etc — and calls `fwrite_stance`). Stance-table edit + save belong with Phase-6 OLC tooling.

**Files:**
- Modified: `internal/act/olc_set.go` (extend `DoMset` with a stance-name branch after existing field-match).
- Modified: `internal/act/olc_set_test.go` (add `TestDoMset_Stance*` cases).

**Scope (narrower than C):** Extend `DoMset` only. The C branch runs AFTER all named fields (str, int, level, etc.) and before the usage-generator, using `get_stance_number(arg2)` to check if `arg2` matches a stance name; if yes, write the mastery. Mirror exactly.

**C divergence note — `STANCE_NONE` guard:** C `src/build.c:3502` guards on `get_stance_number(arg2) > 0` (strict positive), so `mset bob none <val>` is silently a no-op in C because `STANCE_NONE` resolves to index 0. The Go port MUST match this exactly — use `> 0` not `>= 0` — so area builders relying on the C behavior don't get surprised. Add `TestDoMset_StanceNoneSilentNoop` to pin this.

**Test-first:**
- `TestDoMset_StanceSetsPCMastery` — level-60 admin, target PC, `mset bob dragon 200` — assert `victim.PCData.Stances[STANCE_DRAGON] == 200` and `"Done."` sent.
- `TestDoMset_StanceNPCTarget` — target is mob, `mset mob dragon 200` — assert `victim.IndexData.Stances[STANCE_DRAGON] == 200`.
- `TestDoMset_StanceClampsToMax` — input `mset bob dragon 250` — rejected with `"Stance value is from 0 to 200\n"` (matches C `src/build.c:3510-3512` / `:3529-3531`).
- `TestDoMset_StanceNegativeRejected` — input `mset bob dragon -5` — rejected with same message.
- `TestDoMset_StanceInsufficientTrustPC` — caller trust 56 (below LEVEL_LESSER=57), target PC, expect rejection with `"You can only modify a mobile's immunities.\n\r"` (matches C `src/build.c:3524`).
- `TestDoMset_StanceUnknownName` — `mset bob nonesuch 50` — falls through to existing `DoMset` unknown-field handling (no stance match, no error specific to stance — default "bad field" path).

**Mutation (via `Edit` round-trips only):**
- Change `PCData.Stances` write to `IndexData.Stances` on PC path → `TestDoMset_StanceSetsPCMastery` fails.
- Invert the max clamp check (`> MAX_PC_STANCE` → `>= MAX_PC_STANCE`) → `TestDoMset_StanceClampsToMax` fails with boundary-at-200 test.
- Remove the trust gate → `TestDoMset_StanceInsufficientTrustPC` fails.

### G2 — `TIMER_DO_FUN` callback registry + expiry dispatch

**Files:**
- Modified: `internal/handler/timer.go` (add registry + dispatch in `DecrementTimers`)
- Modified: `internal/boot/boot.go` (register known names; initially empty map + TODO comments referencing the unported skill commands)
- New: `internal/handler/timer_registry_test.go`

**Test-first:**
- `TestRegisterTimerFunc_LookupRoundTrip` — register `"do_detrap"` with a spy function; `LookupTimerFunc("do_detrap")` returns same non-nil.
- `TestLookupTimerFunc_UnknownReturnsNil` — `LookupTimerFunc("nonexistent")` returns nil.
- `TestRegisterTimerFunc_NilFuncNoPanic` — register with a nil TimerFunc; lookup returns nil (or registered-nil — define contract and test it).
- `TestDecrementTimers_ExpiryDispatchesKnownDoFun` — install a spy function `"do_spy"`, add timer with `DoFun: "do_spy"`, Value: 17, Count: 1; decrement once; assert spy called with the expected `argument=""` and that `ch.Substate` was set to 17 during the call (spy can record ch.Substate on entry).
- `TestDecrementTimers_ExpiryRestoresSubstate` — set `ch.Substate = 99` before; after dispatch (spy does nothing with substate), `ch.Substate == 99`.
- `TestDecrementTimers_ExpiryDispatchesUnknownDoFunLogsBug` — timer with `DoFun: "nonexistent"`, Count: 1; decrement; timer dropped; no panic; `util.Bug` log observed (seam via a test logger sink).
- `TestDecrementTimers_EmptyDoFunDropsSilently` — timer with `DoFun: ""`, Count: 1; decrement; timer dropped; no dispatch; no bug log.
- `TestDecrementTimers_DoFunThatReExtends` — spy re-calls `AddTimer` for same type with Count: 5; after decrement, check the new timer is present with Count 5 (not dropped by the expiry branch — we follow C `fight.c:425-426` `if (timer->count > 0) continue`). This is the "do_fun re-extended timer" case.

**Mutation-verify (via `Edit` round-trips only — same banned list as G1):**
- Remove the DoFun dispatch line → `TestDecrementTimers_ExpiryDispatchesKnownDoFun` fails.
- Skip substate save/restore → `TestDecrementTimers_ExpiryRestoresSubstate` fails.
- Silent-drop unknown names (no Bug log) → `TestDecrementTimers_ExpiryDispatchesUnknownDoFunLogsBug` fails.
- Don't check `timer->count > 0` after do_fun → `TestDecrementTimers_DoFunThatReExtends` fails (re-extended timer dropped).

**Scope cut (explicit):** The mid-decrement intercept (combat-aborts-skill at C `fight.c:386-398`) is NOT ported in G2. It requires every skill command that sets a TIMER_DO_FUN to be ported (none are yet). When the first skill command ports, that tranche will add the intercept (one-line check before the decrement branch). Documented in the `DecrementTimers` Go doc comment.

**Similarly deferred:** The command-interpreter intercept at `src/interp.c:713-733`. Also requires ported skill commands.

**Rationale for shipping the registry now with no consumers:** When the first skill command ports (e.g., `do_detrap`), its author can focus on the skill logic alone; the timer plumbing is already in place. Landing the registry first without consumers is a cheap, fully-tested investment.

### G3 — Mudprog if-checks (7 of the 9 placeholders)

**Files:**
- Modified: `internal/mudprog/ifcheck.go` (extend the switch; update the deferral comment at `:885-890`).
- Modified: `internal/mudprog/ifcheck_test.go` (add cases).

**Unblocked:**
- `timeskilled` (reads `MobIndexData.Killed`)
- `objtype` (reads `chkObj.IndexData.ItemType`)
- `leverpos` (reads `obj.Value[0] & TRIG_UP`, compares against `"up"`/`"down"` rval)
- `pkadrenalized` (reads `handler.GetTimer(chkchar, TIMER_RECENTFIGHT)`)
- `asupressed` (reads `handler.GetTimer(chkchar, TIMER_ASUPRESSED)`)
- `areamulti` (reads descriptor host, counts PCs in same area)
- `multi` (reads descriptor host, counts PCs world-wide)

**Still deferred** (stays placeholder):
- `isflagged` / `istagged` — needs `get_tag` + `VariableData` lookup by name/vnum. This IS a self-contained subsystem port (~60 LOC + tests), but out of Tranche B scope. Plan opens a follow-up in scope cuts.

**Test-first:**
- `TestIfCheck_TimeSkilled_ChkcharResolvesIndexKilled` — chkchar=mob with IndexData.Killed=5; ifcheck `timeskilled($n) == 5` returns true.
- `TestIfCheck_TimeSkilled_CVarVnumFallback` — chkchar=nil, cvar="3001", pre-loaded index with Killed=3; compare `>= 3` true.
- `TestIfCheck_TimeSkilled_PCTargetReturnsFalseNotPanic` — chkchar is a PC (IndexData==nil); ifcheck must return false without panicking. Closes the nil-deref gap flagged by adversary 2026-04-18; exercises the `IsNPC()` gate / nil-check documented in Mutation-verify below.
- `TestIfCheck_ObjType_Weapon` — chkObj.IndexData.ItemType=5 (ITEM_WEAPON); `objtype($o) == 5` true.
- `TestIfCheck_LeverPos_UpWhenBitSet` — chkObj.ItemType=ITEM_LEVER, Value[0] & TRIG_UP != 0; `leverpos($o) == up` true; `leverpos($o) == down` false.
- `TestIfCheck_LeverPos_WrongItemTypeFalse` — chkObj is ITEM_WEAPON — always false even if Value[0] has TRIG_UP set.
- `TestIfCheck_PkAdrenalized` — install `TIMER_RECENTFIGHT` on chkchar via `handler.AddTimer(ch, TIMER_RECENTFIGHT, 5, "", 0)`; ifcheck `pkadrenalized($n) > 0` true. Count=0 → false.
- `TestIfCheck_ASupressed_PermanentViaValueMinus1` — `AddTimer(ch, TIMER_ASUPRESSED, 0, "", -1)`; `asupressed($n) > 0` — depends on semantics: `GetTimer` returns Count (0 here), so this test returns FALSE. **This is a C-fidelity question (see Open Questions G3.1).**
- `TestIfCheck_AreaMulti_SameHostSameArea` — two PCs with matching `Desc.Host`, both in the same area; `areamulti($n) == 2` true.
- `TestIfCheck_AreaMulti_DifferentAreaExcluded` — same host, different areas — counts as 1 (just chkchar).
- `TestIfCheck_Multi_IgnoresArea` — same-host PCs in different areas — `multi($n) == 2` true.

**Mutation-verify (via `Edit` round-trips only — same banned list as G1):**
- Swap `TRIG_UP` to `TRIG_DOWN` (or equivalent bit) → `TestIfCheck_LeverPos_UpWhenBitSet` fails.
- Drop the area filter in `areamulti` → `TestIfCheck_AreaMulti_DifferentAreaExcluded` fails.
- Drop the host-compare → both area/multi tests fail.
- Return `ch.IndexData.Killed` when chkchar is a PC (PCs have no IndexData) → `TestIfCheck_TimeSkilled_CVarVnumFallback` nil-panics OR returns 0 incorrectly. **G3 implementation must nil-check `chkchar.IndexData` OR gate the whole branch on `chkchar.IsNPC()` — see C `mud_prog.c:578-585` which implicitly assumes chkchar is an NPC, and panics on PC input.** Document this in the case body.

### G4 — Wordlist match port

**Files:**
- New: `internal/mudprog/wordlist.go` (exports `wordlistMatch(arglist, input string) bool` and `containsAtWordBoundary(text, needle string) bool`)
- New: `internal/mudprog/wordlist_test.go`
- Modified: `internal/mudprog/rprog.go:143-183` — replace the first-word-only match with `wordlistMatch`
- Modified: `internal/mudprog/oprog.go:216-221` — implement by iterating `ch.InRoom.Contents` ONLY. C `src/mud_prog.c:3569-3580` walks only `ch->in_room->first_content` (room floor). Items in `ch.Carrying` and `ch.Equipment` are NOT scanned by C for `CMD_PROG` — verified directly by reading the C source (room-only loop).

**Test-first for the helper:**
- `TestWordlistMatch_SingleKeywordMatches` — arglist `"hello"`, input `"hello world"` → true.
- `TestWordlistMatch_SubstringWithoutWordBoundaryFails` — arglist `"ell"`, input `"hello"` → false (substring but not word).
- `TestWordlistMatch_MultipleKeywordsAnyMatches` — arglist `"foo bar"`, input `"bar"` → true.
- `TestWordlistMatch_TrailingPunctuation` — arglist `"hello"`, input `"hello!"` → false (C `'!' != ' '/\n/\r/\0'`).
- `TestWordlistMatch_TrailingNewline` — arglist `"hello"`, input `"hello\n"` → true.
- `TestWordlistMatch_TrailingNull` — arglist `"hello"`, input `"hello"` (end-of-string) → true.
- `TestWordlistMatch_PhrasePrefix` — arglist `"p I am the king"`, input `"I am the king of the hill"` → true.
- `TestWordlistMatch_PhrasePrefixPartialFails` — arglist `"p I am the king"`, input `"I am the king of"` → true (word boundary on "of" — phrase matches at start with trailing space).
- `TestWordlistMatch_PhrasePrefixDoesNotMatchWords` — arglist `"p foo bar"`, input `"bar foo"` → false (phrase, not split).
- `TestWordlistMatch_CaseInsensitive` — arglist `"Hello"`, input `"hello world"` → true.

**Test-first for `RprogCommandTrigger`:**
- `TestRprogCommandTrigger_SubstringWithoutWordBoundaryDoesNotFire` — prog arglist `"wigg"`, input `"wiggle hard"` → returns false (regression against current broken behavior).
- `TestRprogCommandTrigger_WordMatchInMiddle` — prog arglist `"hard"`, input `"wiggle hard"` → returns true.
- Existing `TestRprogCommandTrigger_Consumes` must still pass (regression guard).

**Test-first for `OprogCommandTrigger`:**
- `TestOprogCommandTrigger_RoomObjMatches` — put an obj with MPROG_CMD arglist `"twist"` on the room floor (`ch.InRoom.Contents`); input `"twist"` → returns true.
- `TestOprogCommandTrigger_CarryingObjDoesNotMatch` — same obj moved to `ch.Carrying`; input `"twist"` → returns false (matches C room-only iteration).
- `TestOprogCommandTrigger_NonMatchingReturnsFalse` — same obj with arglist `"pull"`; input `"twist"` → returns false (and falls through to normal dispatch).
- **Replaces the existing `TestOprogCommandTrigger_Stub` test** which asserts `return false` — that test is no longer accurate once the stub is implemented.

**Mutation (via `Edit` round-trips only — banned: `git checkout`/`git reset --hard`/`git stash`):**
- Drop the word-boundary right-side check → `TestWordlistMatch_TrailingPunctuation` flips to true, fails.
- Skip the `"p "` prefix branch → `TestWordlistMatch_PhrasePrefix` fails (tries word-match on "I am the king" as four separate keywords).
- Iterate `ch.Carrying` in oprog (pulls in carried progs C excludes) → `TestOprogCommandTrigger_CarryingObjDoesNotMatch` fails.

## Acceptance Criteria

**G1 (Stances loader):**
- A1. `persist.LoadStancesInto(&combat.StanceIndex, "db/system/stances.dat")` returns nil error for the shipped stub file AND for an actual populated file (tested via fixture).
- A2. Existing combat tests still pass after boot invokes the loader (`go test -count=3 ./internal/combat/... ./internal/persist/...` green).
- A3. A populated stance in `stances.dat` OVERRIDES the defaults; unmentioned stances keep defaults.
- A4. `GetStanceNumber` resolves all 12 stance names case-insensitively.

**G1b (DoMset stance-name extension):**
- A5. `mset <victim> <stance-name> <value>` with caller trust >= LEVEL_LESSER (=57) mutates `PCData.Stances[<stance>]` for PC targets, `IndexData.Stances[<stance>]` for NPC targets.
- A6. Values clamped to `[0, MAX_PC_STANCE=200]` / `[0, MAX_MOB_STANCE=200]` — out-of-range rejected with C-faithful message.
- A7. Non-admin (trust < LEVEL_LESSER) attempting PC target is rejected with `"You can only modify a mobile's immunities.\n\r"`.

**G2 (Timer DO_FUN):**
- A8. `handler.RegisterTimerFunc(name, fn)` and `handler.LookupTimerFunc(name)` round-trip.
- A9. `DecrementTimers` on an expired `TIMER_DO_FUN` with a known DoFun name invokes the function with `ch.Substate = t.Value` during the call; substate is restored after the call.
- A10. Expired timer with unknown DoFun name logs `util.Bug` and drops.
- A11. Expired timer whose DoFun re-extends (via `AddTimer`) leaves the re-extended timer present after `DecrementTimers` returns.

**G3 (Mudprog if-checks):**
- A12. 7 of 9 placeholder if-checks (`timeskilled`, `objtype`, `leverpos`, `pkadrenalized`, `asupressed`, `areamulti`, `multi`) are implemented and tested; `isflagged` and `istagged` remain placeholders with an explicit TODO comment pointing at the variable-subsystem follow-up.
- A13. Ifcheck comment at `ifcheck.go:885-890` updated to reflect which remain.

**G4 (Wordlist match):**
- A14. `wordlistMatch` implements C's word-boundary algorithm including the `"p "` phrase prefix.
- A15. `RprogCommandTrigger` fires on word-boundary keyword matches, not bare substring.
- A16. `OprogCommandTrigger` fires on word-boundary matches against any MPROG_CMD prog attached to an obj in ch's room only (NOT carried or equipped — matches C `mud_prog.c:3569-3580` which walks `ch->in_room->first_content`). Adversary-resolved: plan v1 erroneously said "or carried by ch" — corrected 2026-04-18.

**Global:**
- A17. `go vet ./...` clean.
- A18. `go test -count=3 ./...` green across all packages.

## Scope Cuts / Deferrals

- **`isflagged` / `istagged`** — requires `get_tag` port + `VariableData` lookup by name/vnum. Self-contained follow-up, out of Tranche B.
- **`TIMER_DO_FUN` mid-decrement intercept** (combat-aborts-skill at `src/fight.c:386-398`) — no Go skill commands set `TIMER_DO_FUN` today. Ports with the first skill command.
- **`TIMER_DO_FUN` interp intercept** (new-command-aborts-skill at `src/interp.c:713-733`) — same reason.
- **Stance-table OLC** (full `do_stset` with field editing, `do_ststat`, `fwrite_stance`) — Phase 6 OLC work.
- **Stance `can_use_stance` prerequisite checks** in `DoStance` — currently `DoStance` just sets `ch.Stance` without checking `stance_index[new_stance].stance[0]` prerequisites. Port when we have the full stance table loaded (G1 is the precondition, but the prerequisite field isn't added to `StanceInfo` in G1 — it would require a struct extension). Out of scope; tracked as open question G1.3.
- **Class/race stance restrictions** — same reason.
- **Stance max_weight / dual_wield restrictions** — same reason.
- **Full C `do_stset` with Stance save-to-file** — Phase 6 tooling.
- **Player-facing `practice stance <name>` grind command** — C has no such command. Out of scope by design (see G1b rationale).
- **`mpapply` / `mpapplyb` auth state machine** — known deferral (phase5-tier3 follow-up).
- **`mpmorph` / `mpunmorph`** — needs morph subsystem.
- **Richer hate-list for `mphate`** — current single-slot `Hating`.
- **Per-target `mpPeace`** — current always room-wide.
- **`OprogDamageTrigger` per-weapon-damage vs per-worn-item divergence** — audit-flagged but small; separate tiny follow-up.
- **`MPROG_SPEECH` `"p "` phrase prefix** (`mudprog/triggers.go:49`) — the wordlist helper G4 ships supports it; wiring the speech dispatcher through the new helper is a small follow-up commit AFTER G4 lands.
- **`mortinroom` / `mortinworld` nifty-is-name prefix match** — using `strings.EqualFold` vs C's whitespace-split prefix match. Tracked in TODO.md already.

## Open Questions (with recommended answers)

### G1

**G1.1 — Loader interface:** Should `LoadStances` return a new array, or mutate `combat.StanceIndex` in place? **Recommended:** Mutate in place. Keeps the existing package-level `var` as the consumer-facing table (zero changes to `combat/combat.go`). Signature: `func LoadStancesInto(target *[MAX_STANCE]combat.StanceInfo, path string) error`. Wiring in `boot.Boot` adds one line: `_ = persist.LoadStancesInto(&combat.StanceIndex, filepath.Join(dataDir, "system", "stances.dat"))`.

**G1.2 — Fixtures file placement:** Use `internal/persist/testdata/stances.dat` or `internal/persist/testdata/stances_full.dat`? **Recommended:** Two fixtures: `testdata/stances_empty.dat` (just `End\n`) and `testdata/stances_two.dat` (Dragon + Tiger populated). Keep the real `db/system/stances.dat` untouched.

**G1.3 — Extension to `StanceInfo` struct:** The real stance table has `resist`, `immune`, `suscept`, `stance[0]`, `stance[1]`, `class_restrictions`, `race_restrictions`, `dual_wield`, `max_weight`, `wait`, etc. Load them into a separate `persist.StanceDetails` map keyed by `int`, even if nothing consumes them yet? **Recommended:** Read and discard, log nothing. Adding storage without a consumer is dead code; the loader can be extended in Phase 6 OLC.

**G1.4 — `DoStance` prerequisite check:** Should the port honor `stance_index[<new>].stance[0]` (prerequisite stance that must be GM) before allowing selection? **Recommended:** No (for Tranche B). The prerequisite fields are not stored in Go's `StanceInfo` today (G1.3). Adding them is Phase 6.

### G1b

**G1b.1 — Extend `DoMset` vs ship `DoStset`:** C has two routes to mutate stance mastery: `do_mset stance` (`src/build.c:3499-3537`) which writes mastery values, AND the full `do_stset` (`src/stances.c:742-1010`) which edits the stance-table itself plus has a `save` subcommand for `fwrite_stance`. The former is the only path that edits PCData/IndexData stances. **Recommended:** Extend `DoMset` with a stance-name branch. Ships the PC-mastery-increment path without pulling in the stance-table editor. Phase 6 OLC can add `DoStset` as the full port when stance-table OLC ships.

**G1b.2 — Ordering inside `DoMset`:** C places the stance branch AFTER all other field-name checks, just before the usage-message generator. Match the ordering so an unrecognized arg2 that is a stance name goes through the stance branch, not the unknown-field path.

### G2

**G2.1 — `util.Bug` seam for test observability:** `DecrementTimers` currently logs via `util.Bug` on unknown DoFun name. How do we test this? **Recommended:** `util.Bug` already writes to a package-level `BugLogger` or similar. The test installs a capturing sink via `util.SetBugLogger(fn)` (if that seam exists) or reads from a `bytes.Buffer` if `util.Bug` supports injecting one. Verify seam existence in G2 implementation; if absent, add one.

**G2.2 — Zero-count dispatch re-extension semantics:** If the do_fun returns without re-extending, should we still invoke ExtractTimer? C does `extract_timer(ch, timer)` after the call returns when `timer->count > 0` is false. Go `DecrementTimers` already drops expired timers via the compact-in-place filter. **Recommended:** Invoke AFTER the callback runs, then check `t.Count > 0` to decide whether to keep or drop. Document the sequencing in the Go doc comment.

**G2.3 — Empty `DoFun` dispatch:** What happens if a caller passes `AddTimer(ch, TIMER_DO_FUN, 3, "", 0)` (empty name)? **Recommended:** Drop silently on expiry. The empty string means "no callback"; no bug log. Test pins this (`TestDecrementTimers_EmptyDoFunDropsSilently`).

### G3

**G3.1 — `asupressed` semantics for permanent timers (`Value == -1`):** Verified by scanning all C callers of `add_timer(..., TIMER_ASUPRESSED, ...)`: the ONLY caller is `src/mud_comm.c:328` passing `value=0` (finite timer). The `value == -1` permanent branch in C `fight.c:404-408` is effectively dead code in stock SMAUG — no caller installs a permanent suppression. So the mudprog `asupressed` if-check in both C and Go operates on finite `Count` — no divergence in practice. Go port simply reads `handler.GetTimer(ch, TIMER_ASUPRESSED)` which returns `Count`. If a future Go caller installs a permanent ASUPRESSED (`Value == -1`), a reader of `asupressed > 0` in a mudprog will see a stale `Count` (whatever was set at install time, not a re-bumped 1000). Low-impact edge case; document in the Go doc comment.

**G3.2 — `DescriptorData.Host` field name:** Verified: `internal/types/descriptor.go:27` declares `Host string` with comment `// Host info`. Set at connect time from `net.SplitHostPort(conn.RemoteAddr().String())` (`descriptor.go:209-212`). Use `d.Host` in the `sameHost` helper.

**G3.3 — NPC descriptors:** `areamulti` / `multi` explicitly skip NPCs in C (`!IS_NPC(chkchar) && !IS_NPC(ch)`). **Port verbatim.** NPCs have no descriptors in Go anyway, so the effect is the same, but the explicit guard is idiomatic.

### G4

**G4.1 — Helper location:** `internal/mudprog/wordlist.go` (new file) or in `triggers.go`? **Recommended:** New `wordlist.go` — the helper is 40 LOC with its own test file. Mixing into `triggers.go` clutters an already-long file.

**G4.2 — Shared use:** Should `triggerMatches` (line 40, for MPROG_SPEECH / SPEECHIW / TELL) switch to this helper? **Recommended:** Not in G4 — a separate follow-up after G4 lands. The `triggerMatches` function has a `"p"` skip that is almost-but-not-quite the same as the wordlist helper; switching requires a careful audit. Tracked as a follow-up.

**G4.3 — Iteration scope in OprogCommandTrigger:** C `src/mud_prog.c:3569-3580` walks ONLY `ch->in_room->first_content` (room floor), not carrying/equipment. **Recommended:** Port verbatim — iterate `ch.InRoom.Contents` only. The Go port's plan-v1 assumption that C walks carrying is wrong — verified by direct Read of the C source.

## Risk

- **G1:** Low. Loader is straightforward; `StanceIndex` overwrite is scoped and nil-safe (if the loader crashes mid-file, defaults remain for unread stances — idempotent).
- **G1b:** Low-Medium. New admin command; impact limited to immortals; existing `DoPractice` / `DoStance` untouched.
- **G2:** Medium. Changes the hot `DecrementTimers` path. Mitigations: extensive unit tests (20+ existing + 8 new); registry is trivial (map lookup). Risk of a subtle bug in substate save/restore — pinned by `TestDecrementTimers_ExpiryRestoresSubstate`. No consumers ship in G2, so the risk is entirely theoretical until the first skill ports.
- **G3:** Low. Pure additions to a switch; no existing paths modified. Risk of mis-reading C semantics on edge cases (`leverpos` buggy C guard, `asupressed` permanent-count semantics). G3.1 and G3.3 pin the divergences.
- **G4:** Medium. `RprogCommandTrigger` is live in production — changing its semantics could break existing content scripts. Mitigations: regression test `TestRprogCommandTrigger_Consumes` must still pass; test for new word-boundary correctness; manual audit of any rooms in `db/area/*.are` that use `cmd_prog` (grep first for the feature surface).

## Adversary-Resolved Concerns (2026-04-18 manager self-adversary pass)

Agent-based adversary verification was unavailable for this session; the manager ran a structured self-review against authoritative C sources and Go codebase, with every citation spot-verified via direct Read. Any subsequent external adversary pass should re-check these resolutions.

1. **CRITICAL — `OprogCommandTrigger` iteration scope.** Plan v1 said the helper should iterate `ch.Carrying` + `ch.Equipment` to "match C". Verified by direct Read of `src/mud_prog.c:3569-3580`: C iterates ONLY `ch->in_room->first_content` (room floor). Plan corrected to iterate `ch.InRoom.Contents` only. Test case renamed `TestOprogCommandTrigger_CarryingObjMatches` → `TestOprogCommandTrigger_CarryingObjDoesNotMatch` to pin the correct C semantics. Wrong port would have ALWAYS-FIRED on carried items — a live divergence.

2. **`LEVEL_LESSER` constant value.** Plan v1 said "trust 5+" and "trust 52" in different spots. Verified: `MAX_LEVEL=65`, `LEVEL_LESSER = MAX_LEVEL - 8 = 57` (`types/constants.go:49`). Corrected in A5 and G1b.

3. **`DoStset` vs `DoMset` extension.** Plan v1 routed G1b through a new `DoStset` admin command. C has distinct `do_stset` (stances.c:742, stance-table editor with `save` subcommand) and the `do_mset stance` branch (build.c:3499, stance-mastery setter). The mastery setter is the only path Tranche B needs. Plan corrected: G1b extends `DoMset`, does NOT ship `DoStset` (Phase 6).

4. **`asupressed` permanent-timer semantics drift.** Plan v1 claimed a divergence between Go (returns Count=0 for permanent) and C (returns 1000-ish). Grep of all C callers of `add_timer(..., TIMER_ASUPRESSED, ...)` found only one — `src/mud_comm.c:328` passing `value=0` (finite). The `value == -1` branch in C is dead code. Divergence is theoretical only. Plan G3.1 updated.

5. **`DescriptorData.Host` field name.** Plan v1 had a "to verify during G3" placeholder. Verified: `internal/types/descriptor.go:27` declares `Host string`. G3.2 upgraded from "to verify" to "verified".

6. **`leverpos` C bug note.** Plan v1 correctly flagged C's tautological OR guard at `mud_prog.c:1493-1495` that makes the check always return FALSE in C. Plan's recommendation to port the *intent* (AND of positive matches) stands — otherwise `leverpos` is guaranteed-false in Go too, which nobody wants.

7. **`MPROG_CMD` bit existence.** The `OprogCommandTrigger` stub's own doc comment (`oprog.go:211-215`) claimed "Go port does not yet have a MPROG_CMD bit". That comment is stale; the bit exists at `types/mudprog.go:90`. Plan's Problem section documents this and G4 proceeds on the correct assumption.

8. **`StanceInfo` extension scope.** Plan v1 hinted at extending `StanceInfo` with `Resist`, `Immune`, etc. Verified that the combat hot path reads ONLY `NumAttacks`, `DamDone`, `DamTaken`. The other fields (class/race restrictions, resist/immune, prerequisite stance[], max_weight, dual_wield, wait, etc.) are consumed by `can_use_stance`, `update_stances`, and `do_ststat` — none of which are ported. G1.3 correctly defers the struct extension. Loader reads-and-discards the extra fields; future Phase 6 work promotes them.

9. **Stance fixture whitespace.** Plan v1 used spaces in the Appendix A fixture. C's `fwrite_stance` writes tab-separated. `fread_word` treats both as whitespace. Appendix A now documents both forms; G1 test includes a tab-separated round-trip fixture to pin the real format.

10. **`TIMER_DO_FUN` registry scope cut for mid-decrement intercept.** Plan G2 explicitly defers the combat-aborts-skill path (`fight.c:386-398`) and the command-interp abort path (`interp.c:713-733`). Justification: no Go skill currently sets `TIMER_DO_FUN`, so both intercepts are unreachable. Documented in `DecrementTimers` Go doc comment as a forward note for the first skill porter.

---

## Appendix A — Fixture for G1 stances_two.dat

C `fwrite_stance` uses tab-separated key/value pairs (`fprintf(fp, "Attacks\t%d\n", ...)` at `stances.c:626`). `fread_word` treats tab as whitespace, so space-separated also works. Fixture uses spaces for readability:

```
StartStance Dragon
Attacks 2
DamDone 150
DamTaken 90
EndStance

StartStance Tiger
Attacks 3
DamDone 120
DamTaken 110
EndStance

End
```

(If the Scanner is whitespace-sensitive, fall back to tabs — add a mutation-verify that the real C-format file round-trips via a separate `testdata/stances_tabs.dat`.)

## Appendix B — Rough order of landing (suggested commit phasing)

1. **G1 alone** — loader + fixtures + boot wire. ~90 min worker.
2. **G1b** — `DoStset` admin command. ~45 min worker.
3. **G2** — timer registry + dispatch. ~90 min worker.
4. **G3** — 7 if-check bodies. ~120 min worker (each ifcheck is ~15 min with test).
5. **G4** — wordlist helper + Rprog/Oprog rewires. ~90 min worker.

Total: ~7-8 worker-hours + 5 adversary passes. Can parallelize G1+G2+G3+G4 if workers are independent (they touch different files).

---

## External adversary pass (2026-04-18)

Parent manager spawned an external adversary after the sub-manager's self-review. Verdict: CONCERNS. Two blockers fixed in this revision:

1. **`timerskilled` → `timeskilled` (no 'r')** throughout G3, test names, and acceptance criterion. C at `src/mud_prog.c:574` uses `timeskilled` (meaning "times killed," not "timer skilled"); the existing Go placeholder comment at `ifcheck.go:885` also uses `timeskilled`. The plan v1 spelling would have registered a case string that never fires on real area files.
2. **A16 contradicted G4 design** — A16 said "or carried by ch" but G4 + `TestOprogCommandTrigger_CarryingObjDoesNotMatch` explicitly gate to room-only (matches C `mud_prog.c:3569-3580` walking `ch->in_room->first_content`). A16 corrected.

Nits also addressed:
3. Added `TestIfCheck_TimeSkilled_PCTargetReturnsFalseNotPanic` to pin the nil-safety gate that mutation-verify already expects (previously documented in prose but no positive test case).
4. Added explicit C-divergence note + `TestDoMset_StanceNoneSilentNoop` for the C `build.c:3502` `> 0` guard (so `mset bob none <val>` is a silent no-op matching C).

Adversary confirmed: all other C/Go citations sampled were accurate; complexity is proportional; scope cuts are reasonable; `util.Bug` seam assumption documented for worker to verify at implementation time.
