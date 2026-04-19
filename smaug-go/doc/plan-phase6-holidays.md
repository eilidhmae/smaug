# Plan: Phase 6 — Holidays

**Status:** Planned (2026-04-18). Audited by `audit-holidays` lineage 2026-04-18 (self-review fallback — no `Agent` tool available); 3 material corrections applied (get_holiday callers, do_load claim, Q6 new). See `.claude/drafts/audit-holidays/audit-findings.md` for full report.
**Priority:** Wave 1 Phase 6. Small self-contained port (416 C LOC); no hard blockers; bundles the `month_name[]` port per the roadmap's Soft-Blocker note.
**Scope:** New files `internal/types/holiday.go`, `internal/types/months.go` (`month_name[]` + `MonthName()` helper), `internal/persist/holidays.go` + `internal/persist/holidays_test.go`, `internal/act/holidays.go` + `internal/act/holidays_test.go`. Modifications to `internal/world/world.go` (new `Holidays []*types.HolidayData` slice), `internal/boot/boot.go` (loader wire + command registration + `SysData.MaxHoliday` default), `internal/types/system.go` (no field additions — `MaxHoliday` already exists at :109). No changes to combat or any already-shipped subsystem.

---

## Problem

SMAUG ships a holiday-data subsystem at `src/holidays.c:53-416` that lets immortals author named calendar entries (e.g. `New Year's Day` on month 1, day 1; `Year's End` on month 12, day 30) and lets players list them via `do_holidays`. Data persists at `system/holidays.dat`. The shipped data file at `/home/eilidh/src/smaug/db/system/holidays.dat` contains two holidays (`New Year's Day` and `Year's End`) in C tilde-terminated `#HOLIDAY … End` block format — 195 bytes of payload, 14 lines.

The Go port currently has zero coverage:

1. No `HolidayData` / `HOLIDAY_DATA` equivalent in `internal/types/`.
2. No `month_name[]` array (C defines it at `src/act_info.c:2455-2460` / `src/timezone.c:101-106`; used by `do_time` at `act_info.c:2493`, `do_holidays` at `holidays.c:95`, `do_setholiday` at `holidays.c:352-355`, and `src/timezone.c:522`).
3. No loader or persister for `db/system/holidays.dat`; the existing file is therefore ignored at boot.
4. No `do_holidays` / `do_saveholiday` / `do_setholiday` commands registered.
5. No `do_load` / `DoLoad` admin command to force-reload the holiday chart at runtime (C wires this at `src/db.c:694` for the boot path; `do_load` admin-dispatch at runtime is in a separate C file).

Consequence: a feature that area builders and immortals expect to have on any SMAUG MUD is dead in the Go port. Content-only gap — no gameplay interaction (combat, time, etc.) depends on this.

### What is NOT in scope (C also doesn't have it)

- **Announce-on-time-tick — CORRECTED (audit-holidays 2026-04-18).** Previous draft claimed `get_holiday` has zero callers. **This is wrong.** `src/timezone.c:528` (inside `do_time`, guarded by `#ifdef ENABLE_HOLIDAYS`) appends `"It's a holiday today: <name>"` to the time output, and `src/timezone.c:617-631` (`season_update`, same guard) auto-broadcasts `day->announce` via `echo_to_all(AT_IMMORT, day->announce, ECHOTAR_ALL)` when `time_info.hour == 0` on a holiday day. That IS a real announce-on-time-tick mechanism. The cross-compile unit that houses these calls (`src/timezone.c`) is a SMAUG 2.0 / AFKMud extension; stock SMAUG 1.4 uses `src/act_info.c:2459-2498` for `do_time` WITHOUT `get_holiday` integration — so the behavior is compile-flag dependent. **Scope decision for Go port:** (a) port `GetHoliday(month, day) *HolidayData` as a one-liner lookup (field is already defined on `HolidayData`); (b) wire `DoTime` to consult it and append "&wIt's a holiday today:&W <name>" line; (c) wire the game-tick (`game.PulseUpdate` / seasonal tick — whichever layer handles `season_update`-equivalent today) to `echo_to_all` the announce on hour-0 of a holiday day. **All three gates require locating the Go-side equivalent of `season_update`**; if no such gate exists yet in the port, the announce-tick becomes a follow-up TODO and only (a) + (b) land in this plan. Open Q6 captures this decision.

### Sysdata-configurable ceiling

Both `load_holidays` (`:186`) and `do_setholiday create` (`:295`) gate on `sysdata.maxholiday`. C declares `int maxholiday;` at `mud.h:3449` and lets immortals set it via `cset max-holidays <n>` at `act_wiz.c:8141-8146`. Go already has `SystemData.MaxHoliday` at `internal/types/system.go:109` — the field is declared but never used (grep confirms). The plan wires it with a boot-time default of `32` (matching the C default seeded by `save_sysdata` templates in stock trees).

---

## C Reference (authoritative)

All citations are against `src/holidays.c`, `src/holidays.h`, `src/db.c`, `src/mud.h`, `src/act_info.c`, `src/timezone.c` at HEAD.

### Struct

**`HOLIDAY_DATA`** — `src/holidays.h:48-56`:

```c
struct holiday_data {
    HOLIDAY_DATA *next;
    HOLIDAY_DATA *prev;
    short month;      /* 1-indexed. Month the holiday falls in */
    short day;        /* 1-indexed. Day the holiday falls on */
    char *name;
    char *announce;   /* Shown in listing but not auto-broadcast */
};
```

**1-indexed month/day semantics.** Both fields are stored 1-indexed in the file AND in memory, contrary to the 0-indexed `time_info.month` / `time_info.day` used elsewhere. Evidence: the listing at `holidays.c:95` subtracts 1 to index `month_name[]` (`month_name[day->month - 1]`). The `do_setholiday create` path at `:303-304` stores `time_info.day` / `time_info.month` verbatim without +1, meaning a newly-created holiday comes out 0-indexed — a C **bug** (see Design below: Go port fixes this by emitting `+1` on create, matching the file format's 1-indexed intent).

### Entry points

- **`do_holidays`** — `src/holidays.c:86-98`. Player-facing listing. Prints header then iterates `first_holiday → last_holiday`. Format: `"&G%-21s\t&g%-11s\t%-2d\r\n"` with `(name, month_name[month-1], day)`. Sent via `send_to_pager` / `pager_printf` (routes through the pager subsystem; Go has `game.SendToPager`). **No argument parsing** — listing is unconditional.

- **`do_saveholiday`** — `src/holidays.c:243-249`. Single-purpose: calls `save_holidays()` + emits `"Holiday chart saved.\n"`. Immortal-only in practice (registered with a gated level in SMAUG command tables).

- **`do_setholiday`** — `src/holidays.c:253-416`. OLC-style CRUD dispatcher:
  - Syntax: `setholiday <name> <field> <argument>`.
  - No-arg → syntax help (`:265-270`).
  - `setholiday save ...` → calls `save_holidays()` + success message (`:274-279`). Subcommand shortcut.
  - `arg2 == "create"` → create-new branch (`:282-309`). Gates on duplicate-name, on `sysdata.maxholiday`. Sets `newday->name = str_dup(arg1)`, `newday->day = time_info.day`, `newday->month = time_info.month` (C bug — should be `+1` for file-format consistency), `newday->announce = str_dup("Today is the holiday of when some moron forgot to set the announcement for this one!")`. LINK into global list.
  - Otherwise resolve `day` by name match (`:313-317`). No match → `"Which holiday was that?\n"` (`:320-324`).
  - `arg2 == "day"` → parse `arg3` via `is_number` + range `1-sysdata.dayspermonth`. **C bug** — condition `atoi(arg3) <= 1` rejects day `1` instead of day `0`; port intent `atoi < 1` to allow day 1. (`:328-340`)
  - `arg2 == "month"` → parse `arg3` via `is_number` + range `1-sysdata.monthsperyear`. **Same C bug** `<= 1` vs `< 1`. Fallback prints month list `(1) Winter (2) the Winter Wolf …` via the while-loop at `:352-358` (that while-loop also has a C bug — compares `month_name[x] != '\0'`, i.e., `char*` pointer against null char — always true until OOB). The Go port must emit a clean month-list fallback without mimicking the C loop-bug.
  - `arg2 == "announce"` → empty or numeric arg3 rejected; else `str_dup` replace. (`:369-382`)
  - `arg2 == "name"` → same validation + `str_dup` replace. (`:385-398`)
  - `arg2 == "delete"` → requires `arg3 == "yes"`; else confirmation prompt. Then `free_holiday(day)`. (`:400-411`)
  - Unmatched `arg2` → syntax help footer. (`:413-415`)

### Persistence

- **File path:** `HOLIDAY_FILE` defined at `src/mud.h:5522` as `SYSTEM_DIR "holidays.dat"` → `db/system/holidays.dat` in default layout.

- **`load_holidays`** — `src/holidays.c:149-210`. Open `HOLIDAY_FILE`; zero head/tail pointers; loop:
  - `fread_letter` → if `'*'`, skip comment; if `'#'`, read next word; else bug + break.
  - `#HOLIDAY` → create + `fread_day`. Enforce `daycount < sysdata.maxholiday` — if exceeded, bug + early close + return. LINK into list, increment count.
  - `#END` → break (success).
  - Other → bug + continue.
  - No file present → silently skip (no bug log). Matches C's "missing file is OK" convention.

- **Shipped-file format quirk:** `/home/eilidh/src/smaug/db/system/holidays.dat` does NOT contain a trailing `#END` terminator — it ends after the last block's `End\n\n`. Verified via `od -c`:
  ```
  ...  End\n\n
  ```
  In C, `fread_letter` reads past whitespace and returns EOF (represented as `EOF` int value, which becomes `-1` cast to char). The condition `letter != '#'` at `:177` is then true and C falls through to `bug("load_holidays: # not found.")` + `break`. **C logs a spurious bug on every boot** for the stock shipped file. Go port MUST handle this gracefully:
  - Option 1: match C verbatim (log bug, stop reading). Bug-log emitted every boot. Ugly but faithful.
  - Option 2: treat EOF as implicit `#END` (no bug log). Cleaner behavior, deliberate divergence from C.
  - **Chosen: Option 2** — treat EOF at section-header position as implicit end. Pinned by `TestLoadHolidays_ShippedFileLoadsNoBug` (asserts both: correct contents AND no `util.Bug` emitted). Document in the loader's Go-doc comment.

- **`fread_day`** — `src/holidays.c:100-147`. Per-block KVP parser. Keys: `Announce`, `Day`, `Month`, `Name`, `End`. Comment `*` skipped via `fread_to_eol`. `End` defaults `announce` to `"Today is a holiday, but who the hell knows which one."` if still nil. Unknown word → `bug("fread_day: no match: %s", word)`.

- **`save_holidays`** — `src/holidays.c:212-241`. Truncate-write the full file. Format per holiday:
  ```
  #HOLIDAY
  Name\t\t<name>~
  Announce\t<announce>~
  Month\t\t<month>
  Day\t\t<day>
  End
  
  ```
  (blank line between blocks). Terminator: `#END\n`. Failure opens `perror` + `bug`; no return code propagated.

- **`free_holidays`** / **`free_holiday`** — `:65-84`. C-level cleanup. Go port doesn't need explicit free (GC); skip. Unlink helper wraps the doubly-linked list in C; Go port uses a `[]*HolidayData` slice.

### `month_name[]` + callers

- **Definition** — `src/act_info.c:2455-2460` AND `src/timezone.c:101-106` (duplicate — one or the other is compiled in per `ENABLE_TIMEZONE` ifdef; both have identical values):
  ```
  "Winter", "the Winter Wolf", "the Frost Giant", "the Old Forces",
  "the Grand Struggle", "the Spring", "Nature", "Futility", "the Dragon",
  "the Sun", "the Heat", "the Battle", "the Dark Shades", "the Shadows",
  "the Long Shadows", "the Ancient Darkness", "the Great Evil"
  ```
  17 entries, matching `sysdata.monthsperyear = 17` in default `sysdata.dat`.

- **Declaration** — `src/timezone.h:65` — `extern char *const month_name[MAX_STRING_LENGTH];`.

- **Callers:** `src/act_info.c:2493` (`do_time`), `src/holidays.c:95` (`do_holidays`), `src/holidays.c:352-355` (`do_setholiday` month-list fallback), `src/timezone.c:522` (logging boot banner), `src/starmap.c:53` (comment-only reference — no read).

### `do_load` admin-path wiring

- **`load_holidays` call at boot** — `src/db.c:692-695`:
  ```c
  #ifdef ENABLE_HOLIDAYS
     log_string(_("Loading holiday chart..."));
     load_holidays();
  #endif
  ```
  Runs once during `boot_db`.

- **`do_load` runtime reload — CORRECTED (audit-holidays 2026-04-18).** Previous draft claimed `do_load holiday` exists in `src/act_wiz.c` as an immortal runtime-reload path. **This is wrong.** Grepping C for `do_load` returns only `do_loadarea` (`src/build.c:7955`) and `do_loadup` (`src/act_wiz.c:6745`) — neither reloads the holiday chart. There is NO runtime-reload command for holidays in stock C. Admins wanting to re-read `holidays.dat` after a manual edit must either `saveholiday` (which emits the in-memory state) or copyover/hotboot. Scope cut stands (Go port does not add an umbrella `DoLoad` in this plan), but the rationale is corrected: we are matching the C gap, not papering over an existing C capability.

---

## Go Current State

- `internal/types/system.go:109` — `MaxHoliday int` field declared, never set, never read.
- `internal/types/system.go:100-102` — `DaysPerMonth`, `MonthsPerYear`, `DaysPerYear` fields declared. Currently unused by any `DoTime` / time-of-day logic (Go's `DoTime` at `internal/act/info2.go:95-126` doesn't reference them — it hard-codes "hour of the day" messaging). Holidays plan will be the first consumer.
- `internal/types/system.go:114-121` — `TimeInfoData { Hour, Day, Month, Year, Season, Sunlight int }`. `Month` is 0-indexed matching C `time_info.month` semantics.
- `internal/world/world.go:55-56` — `SysData types.SystemData` + `TimeInfo types.TimeInfoData` live on the world. No `Holidays` slice yet.
- `internal/persist/scanner.go` — line-oriented `Scanner` with `ReadWord`, `ReadString` (tilde-terminated), `ReadNumber`, `ReadToEOL`, `ReadLetter` (check for single-char operators). Identical to C's `fread_*` family. This is what the holidays loader uses.
- `internal/persist/stances.go` — the most recent `system/*.dat` loader, shipped in Tranche B (2026-04-18). Uses `Scanner.ReadWord` in a `for{}` driven by `StartStance`/`EndStance`/`End` tokens. The holidays loader follows the same shape but swaps the token vocabulary to `#HOLIDAY`/`End`/`#END`.
- `internal/game/pager.go:14` — `SendToPager(ch, text)` routes through the pager if enabled, else direct send. Matches C `send_to_pager`. Holidays listing will use it.
- `internal/boot/boot.go:241-345` — boot-time loader wiring: `LoadAreas`, `LoadClasses`, `LoadRaces`, `LoadSkills`, `LoadStancesInto`. Holidays loader will slot in after stances. Boot-fatal-vs-warn semantics: the `LoadStances` path uses `log.Printf("WARNING: ...")` on failure; holidays should follow the same convention (matching C's "missing file is non-fatal").
- `internal/boot/boot.go:388` — `DoTime` registered at `Level: 0`. Holidays commands register here too.
- `internal/util/strings.go:172` — `IsNumber` matches C `is_number`. Used for `setholiday day`/`month` arg3 validation.
- `internal/util/strings.go:14-41` — `OneArgument` splits off first whitespace-delimited token AND LOWERCASES it. Important for `setholiday` arg parsing: `arg1` (holiday name) and `arg2` (field) arrive lowercased. This means:
  - `setholiday "Halloween" create` → arg1 = `halloween`. If we store the name with original casing (from user input), later lookups via arg1 must be case-insensitive OR the stored name must be lowercased.
  - `setholiday "Halloween" Create` (user capitalizes field) → arg2 = `create`. Matches field names always, regardless of user casing. This is the intended behavior.
  - **Decision:** store user-provided names with original casing by using the un-lowercased portion of the raw `argument` string (split off manually before `OneArgument`, or copy from `argument` up to the first space). Lookup by name via `strings.EqualFold` (case-insensitive) — matches C's `str_cmp`. Pin via `TestDoSetHoliday_CreatePreservesNameCase`.
- `internal/types/constants.go:41,49,56` — `LEVEL_SUPREME`, `LEVEL_LESSER`, `LEVEL_IMMORTAL` command-level gates exist.
- `internal/act/` — standard home for user + immortal commands. No dedicated `holidays.go` file.
- **Missing:** `HolidayData` struct, `month_name[]` array, `MonthName(m int) string` helper, `LoadHolidays` / `SaveHolidays`, `DoHolidays` / `DoSaveHoliday` / `DoSetHoliday`, registration in `boot.go`, `world.World.Holidays` slice.

---

## Go Design

### D1 — `HolidayData` struct

**File:** `internal/types/holiday.go` (new).

```go
package types

// HolidayData maps to C HOLIDAY_DATA at src/holidays.h:48-56.
// Month and Day are 1-indexed to match the C file format. The 0-indexed
// TimeInfoData.Month is only translated at creation time (DoSetHoliday
// create branch calls TimeInfo.Month + 1 to emit a 1-indexed file value).
type HolidayData struct {
    Month    int    // 1-indexed. Valid: 1..SysData.MonthsPerYear.
    Day      int    // 1-indexed. Valid: 1..SysData.DaysPerMonth.
    Name     string
    Announce string
}
```

Simple PoD struct. No pointers, no identity semantics; stored in a slice on `World`.

**Rejected alternative:** Doubly-linked `Next` / `Prev` pointers matching C. Go idiom is `[]*HolidayData`; linked-list indexing is an anti-pattern (stances, clans, deities all use slices). The C linked list is a memory-management fiction (enables `LINK` / `UNLINK` macros); slice semantics round-trip through save/load identically.

### D2 — `month_name[]` + `MonthName()` helper

**File:** `internal/types/months.go` (new).

```go
package types

// MonthNames maps to C month_name[] at src/act_info.c:2455-2460 /
// src/timezone.c:101-106. 17 entries; indexed 0..16 matching the
// 0-indexed TimeInfoData.Month semantics. Consumers that hold a
// 1-indexed file-format month subtract 1 before indexing.
var MonthNames = [...]string{
    "Winter",
    "the Winter Wolf",
    "the Frost Giant",
    "the Old Forces",
    "the Grand Struggle",
    "the Spring",
    "Nature",
    "Futility",
    "the Dragon",
    "the Sun",
    "the Heat",
    "the Battle",
    "the Dark Shades",
    "the Shadows",
    "the Long Shadows",
    "the Ancient Darkness",
    "the Great Evil",
}

// MonthName returns the calendar name for a 0-indexed month. Out-of-
// range indices return "<unknown>" rather than panic — matches the
// "fail-soft, never abort" convention for Go file-data callers.
func MonthName(m int) string {
    if m < 0 || m >= len(MonthNames) {
        return "<unknown>"
    }
    return MonthNames[m]
}
```

**0-indexed API, 1-indexed file format.** `MonthName(0)` returns `"Winter"`. Callers with a 1-indexed value (`HolidayData.Month`, the `do_setholiday month-list` arg3) must subtract 1: `MonthName(holiday.Month - 1)`. This mirrors C exactly — `month_name[day->month - 1]` at `holidays.c:95`.

**Chosen convention:** `TimeInfoData.Month`, `time_info.month`, all 0-indexed. `HolidayData.Month`, `holidays.dat` on-disk, `setholiday month <arg>` user input — all 1-indexed. Conversion happens at the boundary (loader reads 1-indexed, CRUD accepts 1-indexed, listing translates via `MonthName(month-1)`).

`MonthName` is exposed so `DoTime` can also use it — not required by this plan but bundled per roadmap's "also nice-to-have for `DoTime`" note. DoTime upgrade is a scope cut (see below).

**Rejected alternative:** 1-indexed `MonthName` that accepts the file-format value directly. Would force every `time_info.month` caller to add 1, shifting a global-API wart. The convention "0-indexed in memory, 1-indexed at persistence boundary" keeps the wart local to persistence + CRUD.

### D3 — `LoadHolidays` / `SaveHolidays`

**File:** `internal/persist/holidays.go` (new).

Follows the shape of `internal/persist/stances.go` (Tranche B analog). Signature:

```go
// LoadHolidays reads db/system/holidays.dat into target. Missing file
// returns nil (not an error). Malformed blocks (unknown keyword, bad
// section) log via util.Bug and skip. Enforces maxHolidays ceiling —
// extra blocks are dropped with a bug log. Returns the populated slice.
// Mirrors C load_holidays at src/holidays.c:149-210.
func LoadHolidays(path string, maxHolidays int) ([]*types.HolidayData, error)

// SaveHolidays truncate-writes holidays to path. Failure returns error
// without mutating path. Mirrors C save_holidays at src/holidays.c:212-241.
func SaveHolidays(path string, list []*types.HolidayData) error
```

**Loader flow:**
1. `os.Open(path)`; if `os.IsNotExist`, return `nil, nil` (missing-file-OK, matches C).
2. `NewScanner`. Loop:
   - `ReadLetter` (or `ReadWord`; see implementation note below). `#` → `ReadWord` for section name. `*` → `ReadToEOL` (comment line).
   - `#HOLIDAY` → `readDay` builds one `HolidayData`; if `len(out) >= maxHolidays`, bug-log + break; else append.
   - `#END` → break (success).
   - Other → bug-log + continue to next line.
3. `readDay` reads KVP pairs until `End` keyword: `Name <str>~`, `Announce <str>~`, `Month <int>`, `Day <int>`. Unknown keyword → bug-log, continue. On `End`: if `Announce == ""` seed default `"Today is a holiday, but who the hell knows which one."`.

**Implementation note on `#`-scan and `*`-comments:** The `internal/persist/scanner.go` Scanner exposes:
- `ReadLetter() byte` (scanner.go:208) — equivalent to C `fread_letter`; returns next non-whitespace character.
- `ReadWord() string` — returns whitespace-delimited word. `#` is NOT whitespace, so `#HOLIDAY` arrives as a single word.
- `ReadToEOL() string` — line-skip for comment handling.
- `skipWhitespace()` (scanner.go:280) — skips space/tab/\r/\n only; it does NOT skip `#$`-prefixed comments despite the misleading doc comment at scanner.go:34. No callers rely on the comment-skip behavior.

Two implementation options:

- (a) Use `ReadWord` consistently; `#HOLIDAY` arrives as the whole word `"#HOLIDAY"`. Match on that. `*`-prefix comments would be returned as words starting with `*` — add a branch to skip-to-EOL on those. The shipped `db/system/holidays.dat` has no comments, so this branch is defensive.

- (b) Use `ReadLetter` for the leading character (truer to C `fread_letter` flow at `holidays.c:170`), then dispatch: `*` → `ReadToEOL`, `#` → `ReadWord` for section name, else bug.

**Chosen: option (b).** Closer to C's parse structure. The stances loader uses `ReadWord` top-level because stances use `StartStance` (full-word) section headers; holidays use `#HOLIDAY`/`#END` where the leading `#` is a semantically-distinct separator. Both options work; (b) matches C line-for-line and handles `*`-comments cleanly for free.

**Saver flow:** Open `os.Create(path)`. For each holiday, write:
```
#HOLIDAY
Name\t\t<name>~
Announce\t<announce>~
Month\t\t<month>
Day\t\t<day>
End

```
Then `#END\n` and close. Any write error propagates up; caller emits a bug log at the call site.

**Atomicity:** C uses `fopen("w")` (truncate). If the write fails mid-way, the file is corrupted. Go port **should NOT** be more defensive than C here — matches the project's "file loaders: `util.Bug()` and continue on bad data, don't abort" convention but tightens nothing beyond it. A future robustness pass can convert to `os.CreateTemp + rename` if any other persister adopts that pattern.

### D4 — `DoHolidays` / `DoSaveHoliday` / `DoSetHoliday`

**File:** `internal/act/holidays.go` (new).

#### `DoHolidays(ch, argument)` — player-facing listing.

```
&RHoliday                   &YMonth          &GDay
&g----------------------+----------------+---------------
&G<name>                 &g<month name>   <day>
...
```

Uses `game.SendToPager` + `fmt.Sprintf` loops. Reads from `act.WorldRef.Holidays` (analogous to `WorldRef.Clans`, `WorldRef.Races`, etc.). No argument parsing; no level gate (matches C — player-visible).

#### `DoSaveHoliday(ch, argument)` — immortal save.

Calls `persist.SaveHolidays(holidayPath, act.WorldRef.Holidays)` + sends `"Holiday chart saved.\n\r"`. Registered at `Level: LEVEL_IMMORTAL`.

Path resolution: `holidayPath` computed at boot as `filepath.Join(dataDir, "system", "holidays.dat")` and plumbed via a package-level var `persist.HolidayFilePath` or an act-level `act.HolidayFilePath` — either pattern works. **Chosen:** an `act`-level package var `HolidayFilePath string`, set at boot alongside `WorldRef`. Matches the project's existing "boot-wired package var" pattern (e.g., `persist.SkillNameLookup`, `persist.SkillGetter`).

#### `DoSetHoliday(ch, argument)` — immortal CRUD.

Argument pattern: `<name> <field> <value>`. Dispatch:

- Empty `arg1` → syntax help.
- `arg1 == "save"` → call `SaveHolidays` (subcommand shortcut matching C `:274-279`).
- `arg2 == "create"` → create branch. Reject on duplicate name, reject on `len(holidays) >= sysdata.maxholiday`. Seed `Month = TimeInfo.Month + 1` (Go port FIXES the C bug: C stores `time_info.month` unshifted at `:303-304`, producing a 0-indexed value in a 1-indexed file slot. The Go port emits `+1` to align with file semantics. Noted as a deliberate deviation — see Open Questions Q1).
- Else resolve `day` by case-insensitive name match; missing → `"Which holiday was that?\n\r"`.
- `arg2 == "day"` → parse arg3 via `util.IsNumber`, validate `1..SysData.DaysPerMonth`. **Go port FIXES the C bug** at `:330` (`<= 1` rejecting day 1) — Go uses `< 1` (the C intent). Noted as deliberate deviation.
- `arg2 == "month"` → parse arg3, validate `1..SysData.MonthsPerYear`. **Go port FIXES the C bug** at `:346`. Fallback on missing/bad arg3 emits a numbered month list via a clean iteration over `types.MonthNames` (NOT the C while-loop with its pointer-vs-char-null comparison bug).
- `arg2 == "announce"` → reject empty/numeric arg3; else `day.Announce = arg3`. C uses `str_dup` on the whole remainder-after-arg2; Go port should pass the post-`arg2` remainder, not just the one-word `arg3`. **Deliberate divergence from a literal port** — see Open Questions Q2.
- `arg2 == "name"` → same as announce but on `day.Name`. Uniqueness is NOT re-checked on rename in C (`:385-398`) — a rename can collide with another holiday's name, creating an ambiguous-lookup condition. Port verbatim, note as a latent issue (Open Questions Q3).
- `arg2 == "delete"` → require arg3 == `"yes"`; else confirm prompt. Remove from slice.
- Unmatched `arg2` → syntax help footer.

**Slice removal:** Standard Go-idiomatic `append(list[:i], list[i+1:]...)` as used in `world.RemoveChar`.

### D5 — Boot wiring

**File:** `internal/boot/boot.go` — modify.

1. After stances loader (around `:299`), add:
   ```go
   // Load holiday chart from db/system/holidays.dat. Missing file is
   // non-fatal (matches C load_holidays at src/holidays.c:149-210).
   if w.SysData.MaxHoliday == 0 {
       w.SysData.MaxHoliday = 32 // default ceiling; matches stock sysdata.dat seed.
   }
   holidayPath := filepath.Join(dataDir, "system", "holidays.dat")
   holidays, err := persist.LoadHolidays(holidayPath, w.SysData.MaxHoliday)
   if err != nil {
       log.Printf("WARNING: failed to load holidays: %v", err)
   }
   w.Holidays = holidays
   act.HolidayFilePath = holidayPath
   log.Printf("Loaded %d holidays.", len(w.Holidays))
   ```

2. After existing `MonthsPerYear`/`DaysPerMonth` defaults wherever they land: if `SysData.MonthsPerYear == 0`, default to `len(types.MonthNames)` (17); if `SysData.DaysPerMonth == 0`, default to 30 (matches C default). **Check current boot:** these fields may already be defaulted when reading `sysdata.dat` elsewhere; if so, no new code needed. Research confirms no Go code currently writes `SysData.MaxHoliday`/`MonthsPerYear`/`DaysPerMonth` — the plan adds the defaults at the holiday-loader site.

3. Register commands in the same block as `DoTime`, with levels matching the stock `db/system/en/commands.dat` entries at lines 1385-1391 (`holidays`, Level 0), 3864-3870 (`setholiday`, Level 60), 3872-3878 (`saveholiday`, Level 60):
   ```go
   reg.Register(&command.Command{Name: "holidays", DoFun: act.DoHolidays, Position: types.POS_DEAD, Level: 0})
   reg.Register(&command.Command{Name: "saveholiday", DoFun: act.DoSaveHoliday, Position: types.POS_DEAD, Level: 60})
   reg.Register(&command.Command{Name: "setholiday", DoFun: act.DoSetHoliday, Position: types.POS_DEAD, Level: 60})
   ```
   Using the literal `60` rather than a named constant — the SMAUG level table has no name for 60 (LEVEL_LESSER=57, LEVEL_GOD=58, LEVEL_GREATER=59, LEVEL_ASCENDANT=60 per `types/constants.go:46`). Actually `LEVEL_ASCENDANT = MAX_LEVEL - 5 = 60` — so we can use `types.LEVEL_ASCENDANT` for type-safety. **Chosen: `types.LEVEL_ASCENDANT`.**

**File:** `internal/world/world.go` — modify.

Add `Holidays []*types.HolidayData` to `World` struct (in the "Data tables" or "Subsystems" block near `Clans`, `Deities`).

---

## Task Groups

### G1 — `month_name[]` port + `MonthName()` helper + HolidayData struct

**Files:**
- New: `internal/types/months.go`, `internal/types/holiday.go`.
- New: `internal/types/months_test.go`.

**Test-first:**
- `TestMonthNames_HasSeventeenEntries` — pin `len(types.MonthNames) == 17`.
- `TestMonthNames_WinterFirst` — pin `MonthNames[0] == "Winter"`.
- `TestMonthNames_LastEntry` — pin `MonthNames[16] == "the Great Evil"`.
- `TestMonthName_ZeroIndexed` — `MonthName(0) == "Winter"`, `MonthName(16) == "the Great Evil"`.
- `TestMonthName_OutOfRange` — `MonthName(-1) == "<unknown>"`, `MonthName(17) == "<unknown>"`, `MonthName(1000) == "<unknown>"`; no panic.
- `TestMonthName_MatchesCArray` — all 17 entries, pinned exactly against C `src/act_info.c:2455-2460`. Single table-driven test.
- `TestHolidayData_ZeroValue` — `(&HolidayData{}).Month == 0 && .Day == 0 && .Name == "" && .Announce == ""`; confirms struct definition compiles and has the expected field set.

**Mutation-verify (via `Edit` round-trips only — BANNED for mutation revert: `git checkout -- <file>`, `git checkout <ref> -- <file>`, `git restore <file>`, `git reset --hard` any form, `git stash` any form. Apply the mutation with `Edit`; run test; confirm failure; call `Edit` again with the opposite change to revert):**
- Swap `MonthNames[0]` and `MonthNames[1]` → `TestMonthNames_WinterFirst` + `TestMonthName_ZeroIndexed` + `TestMonthName_MatchesCArray` fail.
- Remove the `m >= len(MonthNames)` guard in `MonthName` → `TestMonthName_OutOfRange` panics.
- Drop the `Announce` field from `HolidayData` → `TestHolidayData_ZeroValue` compile-fails.

**Acceptance gate:** `go build ./...` clean. `go test -count=3 ./internal/types/...` green.

### G2 — `LoadHolidays` / `SaveHolidays`

**Files:**
- New: `internal/persist/holidays.go`, `internal/persist/holidays_test.go`.
- New fixtures: `internal/persist/testdata/holidays_two.dat` (canonical two-holiday fixture matching `db/system/holidays.dat` contents), `internal/persist/testdata/holidays_empty.dat` (just `#END\n`), `internal/persist/testdata/holidays_malformed.dat` (unknown `#BOGUS` section, unknown `Garbage` key inside a block).

**Test-first (loader):**
- `TestLoadHolidays_MissingFileReturnsNilNil` — path that doesn't exist → `list == nil`, `err == nil`.
- `TestLoadHolidays_EmptyFileReturnsEmpty` — file = `#END\n` → `len(list) == 0`, no bug log.
- `TestLoadHolidays_TwoHolidaysRoundTrip` — fixture matching the shipped `db/system/holidays.dat` contents → `len == 2`; `list[0] == {Month:1, Day:1, Name:"New Year's Day", Announce:"Today is the first day of a new year."}`; `list[1] == {Month:12, Day:30, Name:"Year's End", Announce:"Today is the last day of the year."}`.
- `TestLoadHolidays_MaxHolidaysEnforced` — fixture with 3 blocks + `maxHolidays=2` → `len == 2`; bug log `"load_holidays: more holidays than 2"` observed via the `util.Bug` test-sink pattern.
- `TestLoadHolidays_UnknownSectionLogsBugContinues` — fixture with `#BOGUS` section between two valid `#HOLIDAY` blocks → both valid blocks loaded, bug log observed.
- `TestLoadHolidays_UnknownKeyInBlockLogsBugContinues` — block with `Garbage 42` between `Name` and `Day` → block still loaded with valid fields, bug log observed.
- `TestLoadHolidays_MissingAnnounceDefaults` — block without `Announce` key → `Announce == "Today is a holiday, but who the hell knows which one."` (matches C `:130`).
- `TestLoadHolidays_ShippedFileLoads` — path = real `/home/eilidh/src/smaug/db/system/holidays.dat` → `len == 2`, matches canonical contents. Pinning test against shipped data — guards against accidental `holidays.dat` corruption and loader drift in one assertion.
- `TestLoadHolidays_ShippedFileLoadsNoBug` — same as above AND no `util.Bug` emission observed. Pins the C-divergent "implicit `#END` at EOF" decision; the shipped file has no explicit `#END` terminator and the Go port treats that as success, not a bug log.
- `TestLoadHolidays_ExplicitEndTerminates` — fixture with `#END\n` after two blocks → `len == 2`, no extra reads after `#END`. Pins that an explicit terminator is still honored (both paths produce the same result).

**Test-first (saver):**
- `TestSaveHolidays_RoundTripStructEqual` — seed 2 holidays, save to temp, reload via `LoadHolidays`, assert struct-equal (reflect.DeepEqual or field-by-field).
- `TestSaveHolidays_ProducesExpectedBytes` — save 2 fixture holidays, read bytes, assert the file format matches `#HOLIDAY\nName\t\t<name>~\nAnnounce\t<announce>~\nMonth\t\t<m>\nDay\t\t<d>\nEnd\n\n` repeated + `#END\n` terminator. Matches C `save_holidays` at `:228-237`.
- `TestSaveHolidays_EmptyListWritesOnlyTerminator` — save `nil` → file contains exactly `#END\n`.
- `TestSaveHolidays_WriteErrorPropagates` — save to a path under a read-only directory (or `/dev/full`-equivalent on Linux) → err non-nil, struct not corrupted.
- **NOTE on shipped-file asymmetry:** the Go saver emits `#END`; the shipped `db/system/holidays.dat` does not. After a single `DoSaveHoliday` call, the file gains a `#END` terminator. This is cosmetically different but semantically identical — the loader handles both with equal results (A3 + the EOF-as-implicit-END test). Not an acceptance-criteria concern; documented here so future reviewers understand why the shipped file lacks `#END` while saved files have it.

**Mutation-verify (via `Edit` round-trips only — same banned list as G1):**
- Remove the `Announce` default in the `End`-branch → `TestLoadHolidays_MissingAnnounceDefaults` fails.
- Swap the `Month` / `Day` field writes in the KVP dispatch → `TestLoadHolidays_TwoHolidaysRoundTrip` fails.
- Drop the `maxHolidays` ceiling check → `TestLoadHolidays_MaxHolidaysEnforced` fails.
- Change `SaveHolidays` terminator from `#END\n` to `#END\r\n` → `TestSaveHolidays_RoundTripByteForByte` fails on byte assertion.
- Remove the unknown-section bug log → `TestLoadHolidays_UnknownSectionLogsBugContinues` passes on list count but fails on bug-log assertion.

**Acceptance gate:** `go build ./...` clean. `go test -count=3 ./internal/persist/...` green.

**C-bug port decisions (explicit):**
- Loader accepts `Announce` empty + seeds default — MATCH C verbatim.
- Loader silently drops over-ceiling blocks after logging — MATCH C verbatim.
- Saver emits `#HOLIDAY`, four KVP lines, `End`, blank line between blocks, terminates `#END` — MATCH C verbatim byte-for-byte.

### G3 — `DoHolidays` / `DoSaveHoliday` / `DoSetHoliday` commands

**Files:**
- New: `internal/act/holidays.go`, `internal/act/holidays_test.go`.

**Test-first (DoHolidays):**
- `TestDoHolidays_EmptyListPrintsHeaderOnly` — `WorldRef.Holidays = nil` → output contains `"&RHoliday"` header line, no holiday rows.
- `TestDoHolidays_SingleHolidayRendersRow` — one holiday `{Month:1, Day:1, Name:"New Year's Day", Announce:""}` → output row contains `"New Year's Day"` + `"Winter"` (from `MonthName(0)`) + `"1"`.
- `TestDoHolidays_FiveHolidaysAllPresent` — five holidays → all five names appear in output, in list order.
- `TestDoHolidays_MonthOutOfRangeShowsUnknown` — malformed holiday with `Month=99` → row contains `"<unknown>"` (safety guard against corrupt data; behavior divergence from C which would crash via month_name[98] OOB).
- `TestDoHolidays_PagerEnabledRoutesThroughPager` — ch with `PCFLAG_PAGERON` → output accumulated in pager buffer, not direct-send buffer.

**Test-first (DoSaveHoliday):**
- `TestDoSaveHoliday_SendsConfirmation` — stub `persist.SaveHolidays` via a seam (or write to a temp path and check existence) → output contains `"Holiday chart saved."`.
- `TestDoSaveHoliday_SaveErrorLogsBug` — temp path in read-only dir → output still confirms (C doesn't report error to ch; matches C `:245-246`), bug log observed.

**Test-first (DoSetHoliday):**

Create branch:
- `TestDoSetHoliday_NoArgShowsSyntax` — empty → `"Syntax : setholiday <name> <field> <argument>"` substring present.
- `TestDoSetHoliday_CreateAddsToList` — `setholiday "Halloween" create` with empty list + `TimeInfo={Month:9, Day:30, Year:2026}` + `SysData.MaxHoliday=32` → `len(list) == 1`, `list[0] == {Month:10 (0-indexed 9 +1 for file format), Day:31 (0-indexed 30 +1), Name:"Halloween", Announce:"Today is the holiday of when some moron forgot to set the announcement for this one!"}`, output contains `"Holiday created."`.
- `TestDoSetHoliday_CreatePreservesNameCase` — `setholiday "Halloween" create` → `list[0].Name == "Halloween"` (NOT `"halloween"`). Guards against the OneArgument-lowercases pitfall; forces the implementation to extract the name from the raw `argument` string rather than from the lowercased OneArgument return.
- `TestDoSetHoliday_LookupCaseInsensitive` — create `"Halloween"`, then `setholiday halloween day 15` (lowercased by user) → day updated. Pins that lookup uses `EqualFold`-semantics.
- `TestDoSetHoliday_CreateRejectsDuplicateName` — seed list with `"Halloween"`; retry create → output `"A holiday with that name exists already!"`, list unchanged.
- `TestDoSetHoliday_CreateRejectsOverMax` — seed list to `SysData.MaxHoliday` holidays; retry create → output `"There are already too many holidays!"`, list unchanged.

Day branch:
- `TestDoSetHoliday_DayUpdatesValue` — seed 1 holiday `Day=1`; `setholiday "Halloween" day 15` with `SysData.DaysPerMonth=30` → `list[0].Day == 15`, output `"Day changed."`.
- `TestDoSetHoliday_DayRejectsNonNumeric` — `setholiday "Halloween" day abc` → output syntax help, list unchanged.
- `TestDoSetHoliday_DayRejectsZero` — `setholiday "Halloween" day 0` (Go port FIXES C bug `<= 1` → `< 1`) → rejected, list unchanged.
- `TestDoSetHoliday_DayAcceptsOne` — `setholiday "Halloween" day 1` → accepted (this is the pin-against-C-bug test: C would reject this; Go must accept).
- `TestDoSetHoliday_DayRejectsOverMax` — `SysData.DaysPerMonth=30`; `setholiday "Halloween" day 31` → rejected.

Month branch:
- `TestDoSetHoliday_MonthUpdatesValue` — `SysData.MonthsPerYear=17`; `setholiday "Halloween" month 10` → accepted.
- `TestDoSetHoliday_MonthAcceptsOne` — pin-against-C-bug: Go accepts `month 1`.
- `TestDoSetHoliday_MonthMissingArgShowsList` — `setholiday "Halloween" month` with no arg3 → output contains month list `(1) Winter` through `(17) the Great Evil`.

Announce / name branches:
- `TestDoSetHoliday_AnnounceSetsString` — `setholiday "Halloween" announce "Trick or treat!"` → `list[0].Announce == "Trick or treat!"`.
- `TestDoSetHoliday_AnnounceRejectsEmpty` — `setholiday "Halloween" announce` → rejected.
- `TestDoSetHoliday_NameRenames` — `setholiday "Halloween" name "AllHallows"` → `list[0].Name == "AllHallows"`.

Delete branch:
- `TestDoSetHoliday_DeleteRequiresYes` — `setholiday "Halloween" delete` → confirm-prompt, list unchanged.
- `TestDoSetHoliday_DeleteYesRemoves` — `setholiday "Halloween" delete yes` → `len(list) == 0`, output `"Holiday deleted."`.

Not-found branch:
- `TestDoSetHoliday_NotFoundMessage` — `setholiday "NonExistent" day 1` → output `"Which holiday was that?"`, list unchanged.

Save subcommand:
- `TestDoSetHoliday_SaveSubcommandInvokesSaver` — `setholiday save` → same confirmation as `DoSaveHoliday`.

**Mutation-verify:**
- Swap `+1` to `+0` in create's `TimeInfo.Month + 1` → `TestDoSetHoliday_CreateAddsToList` fails on Month value.
- Change `< 1` guard back to `<= 1` in day branch → `TestDoSetHoliday_DayAcceptsOne` fails (and the C-bug re-manifests).
- Remove duplicate-name check in create → `TestDoSetHoliday_CreateRejectsDuplicateName` fails (list grows to 2).
- Remove the `len(list) >= MaxHoliday` check → `TestDoSetHoliday_CreateRejectsOverMax` fails.

**Acceptance gate:** `go build ./...` clean. `go test -count=3 ./internal/act/...` green.

### G4 — Boot wiring (loader + commands + sysdata defaults) + World slice

**Files:**
- Modified: `internal/boot/boot.go`, `internal/world/world.go`.
- Modified: `internal/boot/boot_test.go` (add post-boot assertions).

**Changes:**

1. `world.World.Holidays []*types.HolidayData` slice added.
2. Post-stances-loader block in `Boot()` loads holidays, seeds `SysData.MaxHoliday = 32` if zero, sets `act.HolidayFilePath`, registers three commands.
3. `TestBoot_LoadsHolidays` in `boot_test.go`: point the boot to a `testdata/` layout with a two-holiday `holidays.dat`; post-boot, `len(world.Holidays) == 2`, `world.SysData.MaxHoliday == 32` (default), registry contains `holidays`, `saveholiday`, `setholiday` commands with expected levels.
4. `TestBoot_MissingHolidaysFileIsNonFatal` — boot with no `holidays.dat` in the test data dir → boot succeeds, `len(world.Holidays) == 0`, commands still registered.

**Mutation-verify:**
- Remove the `MaxHoliday == 0` default-to-32 guard and zero out `SysData` in the test setup → `TestDoSetHoliday_CreateRejectsOverMax` (G3 test) fails OR create succeeds when it shouldn't. Edit revert restores.
- Unregister one of the three commands → `TestBoot_LoadsHolidays` fails on registry assertion.

**Acceptance gate:** `go build ./...` clean. `go vet ./...` clean. `go test -count=3 ./...` green across all packages. Manual smoke: `./smaug-go -port 4000 -data ../db`, login as immortal, `holidays` → lists 2 entries; `setholiday "Test" create` → creates; `saveholiday` → persists; restart, `holidays` → still shows `"Test"`.

---

## Acceptance Criteria

**A1.** `types.MonthNames` exists as `[17]string` matching C `src/act_info.c:2455-2460` verbatim. `types.MonthName(m int)` returns `MonthNames[m]` for 0 ≤ m < 17, `"<unknown>"` otherwise, no panics on any int input.

**A2.** `types.HolidayData{Month, Day int; Name, Announce string}` exists at `internal/types/holiday.go`. `Month` and `Day` are 1-indexed in-memory (match file-format semantics).

**A3.** `persist.LoadHolidays(path string, maxHolidays int) ([]*types.HolidayData, error)` reads the shipped `db/system/holidays.dat` and returns exactly 2 entries: `{1, 1, "New Year's Day", "Today is the first day of a new year."}` and `{12, 30, "Year's End", "Today is the last day of the year."}`. Missing file returns `(nil, nil)`. Blocks beyond `maxHolidays` dropped with `util.Bug` log.

**A4.** `persist.SaveHolidays(path, list) error` produces a file that round-trips identically through `LoadHolidays` (struct-equal). Empty list writes exactly `#END\n`. Write errors propagate up.

**A5.** `DoHolidays` command registered at `Level: 0`, `POS_DEAD`. Lists all entries in `WorldRef.Holidays` through the pager, with month-name translation via `MonthName(Month - 1)`.

**A6.** `DoSaveHoliday` command registered at `Level: LEVEL_ASCENDANT` (60), `POS_DEAD`. Matches stock `commands.dat` entry. Invokes `SaveHolidays(HolidayFilePath, WorldRef.Holidays)` + confirms `"Holiday chart saved."`.

**A7.** `DoSetHoliday` command registered at `Level: LEVEL_ASCENDANT` (60), `POS_DEAD`. Matches stock `commands.dat` entry. Supports: empty-arg-syntax, `save` subcommand, `create` (with duplicate + MaxHoliday gates, stores `TimeInfo.Month + 1` / `TimeInfo.Day + 1` to match 1-indexed file format), `day <n>` (valid range `1..SysData.DaysPerMonth`, accepts 1), `month <n>` (valid range `1..SysData.MonthsPerYear`, accepts 1, month-list fallback), `announce <str>`, `name <str>`, `delete yes`, not-found message.

**A8.** `world.World.Holidays []*types.HolidayData` slice exists. `boot.Boot` loads it from `data/system/holidays.dat` after the stances loader. Missing file is a non-fatal `log.Printf("WARNING: …")`. Post-boot, `act.HolidayFilePath` is set to the resolved absolute path.

**A9.** `SysData.MaxHoliday` is defaulted to 32 if zero at boot (matches stock C `sysdata.dat` default).

**A10.** `go build ./...` clean. `go vet ./...` clean. `go test -count=3 ./...` green across all packages, including the three new test files (`types/months_test.go`, `persist/holidays_test.go`, `act/holidays_test.go`) and the two updated test files (`boot_test.go`, `world/world_test.go` if Holidays field addition requires a test update).

**A11.** Every worker mutation in G1 / G2 / G3 / G4 is reverted via `Edit` tool round-trips, never destructive `git checkout` / `git restore` / `git reset --hard` / `git stash`.

**A12.** `TestLoadHolidays_ShippedFileLoads` pins the canonical contents of `/home/eilidh/src/smaug/db/system/holidays.dat` — prevents accidental data corruption or loader drift.

**A13.** Go port deliberately diverges from C in three places, each tested by a pinning test AND documented inline in the Go source:
  - `DoSetHoliday day 1` accepted (C rejects via `<= 1` bug) — pinned by `TestDoSetHoliday_DayAcceptsOne`.
  - `DoSetHoliday month 1` accepted (same C bug) — pinned by `TestDoSetHoliday_MonthAcceptsOne`.
  - `DoSetHoliday create` stores `TimeInfo.Month + 1` (C bug stores unshifted 0-indexed value in 1-indexed file slot) — pinned by `TestDoSetHoliday_CreateAddsToList`.

---

## Scope Cuts / Deferrals

- **`get_holiday(month, day)` function — CORRECTED (audit-holidays 2026-04-18).** Previous draft claimed zero callers. Wrong — see "What is NOT in scope" section above: `src/timezone.c:528` (`do_time` holiday-today suffix) and `src/timezone.c:622` (`season_update` hour-0 announce) both call `get_holiday`, guarded by `#ifdef ENABLE_HOLIDAYS`. Decision moved to Open Q6: either port `GetHoliday` + `DoTime` suffix + season-tick announce as part of this plan, or explicitly defer the timezone-integration pieces to a follow-up. Recommended path: port `GetHoliday` + `DoTime` suffix in this plan (local, cheap); defer the season-tick announce unless a season-update equivalent already exists Go-side. Audit action: grep Go for `season_update` / tick / pulse infrastructure before deciding.

- **`DoLoad` umbrella admin command.** C `do_load holiday` (in `src/act_wiz.c`) provides runtime reload of the holiday chart separate from boot-time load. Go port doesn't register `DoLoad` anywhere today, and the three holiday commands + boot-time load cover the common paths. `DoLoad` is a Phase-6 general admin-reload subsystem (potentially covering skills, races, holidays together) — out of this plan's scope.

- **`DoTime` upgrade to use `MonthName`.** Go `DoTime` at `internal/act/info2.go:95-126` prints `"Day %d of month %d, year %d"` — generic rather than using `month_name`. Could be upgraded with `MonthName(WorldRef.TimeInfo.Month)` for `"the Month of %s"` styling matching C `do_time` at `act_info.c:2493`. Not a blocker for holidays. Deferred to its own sub-plan (or bundled with a future `DoTime` polish pass). Noted in follow-up TODOs.

- **`cset max-holidays` immortal command.** C `act_wiz.c:8141-8146` exposes a `cset max-holidays N` admin command to tune `sysdata.maxholiday` at runtime. Go port does not have `DoCset` yet; `SysData.MaxHoliday` is boot-time-defaulted to 32 and can only be changed by editing source. Deferred to a future `DoCset` plan.

- **`sysdata.dat` loader integration.** The stock SMAUG tree persists `sysdata.maxholiday` / `sysdata.monthsperyear` / `sysdata.dayspermonth` in `db/system/sysdata.dat`. Go port doesn't have a `LoadSysdata` / `SaveSysdata` pair today — `SysData` is zero-initialized and never persisted. This plan hard-codes a default `MaxHoliday = 32` at the holidays boot site. A future plan can unify all `sysdata.*` fields through a proper loader; at that time the holidays plan's inline default will be replaced by the generic loader's value.

- **Non-English `month_name` translations.** C `src/timezone.c:102-106` wraps each month name in `___(...)` i18n macros. Go port doesn't have an i18n story. Hard-code English. Out of scope.

- **`do_time` calendar styling.** Currently generic; C uses `ordinal_day + " the Month of " + month_name[month]`. Bundling the calendar-style upgrade with holidays would broaden the plan without benefit — flagged as its own TODO item.

- **`free_holidays` explicit cleanup.** C-level memory management. Go GC makes this a no-op.

- **Per-player "favorite holidays" or subscriptions.** Not in C; not invented for Go. "No new gameplay invented" rule from `phase6-roadmap.md`.

---

## Open Questions

**Q1.** The C `do_setholiday create` path at `src/holidays.c:303-304` stores `time_info.day` / `time_info.month` unshifted — producing a 0-indexed value in a 1-indexed file format slot. Newly-created holidays in C therefore render one month/day earlier than reality. Should the Go port match the C bug verbatim for fidelity, or emit `+1` to match the file-format semantics?

**Recommended answer:** Emit `+1`. The C code is an obvious bug — the file format is 1-indexed (see `holidays.c:95` `month_name[day->month - 1]`) and `time_info.month` is 0-indexed. Storing without shift corrupts every immortal-created holiday. The Go port's "fix C bugs where the fix is local and obvious" precedent exists (Tranche B G3 `leverpos` ported intent not the buggy C chain; Tranche B G1b `> 0` vs `>= 0` matched C exactly because the "bug" was really intentional `STANCE_NONE` handling). For this, fix the bug; document in source comment.

**Q2.** C `do_setholiday announce` passes `arg3` (one word) to `str_dup`, truncating multi-word announcements silently. Should the Go port follow C verbatim (one-word), or take the full remainder-after-`<name> <announce>`?

**Recommended answer:** Take the full remainder. Announcements are short sentences (default is `"Today is the holiday of when some moron forgot..."` — a 13-word string). One-word-only makes the field useless in practice. Port intent — the field is meant to hold prose. Mark as deliberate divergence, pinned by `TestDoSetHoliday_AnnounceSetsString` with a multi-word arg.

**Q3.** C `do_setholiday name` does not re-check for duplicate names on rename, so two holidays can end up with identical names after `setholiday "A" name "B"` where `"B"` already exists. Should Go port mirror this or enforce uniqueness on rename?

**Recommended answer:** Mirror C. The uniqueness check in `create` is enforced; rename is a separate admin operation where the immortal presumably knows what they're doing. If rename creates a duplicate, the second-added holiday becomes unreachable by name-lookup until deleted by list-position (which we don't support). This is a latent UX issue but NOT a data-integrity one — `save_holidays` writes both entries, `load_holidays` reads both. Flag as a follow-up in `TODO.md` Active but do not block the plan.

**Q4.** Should the `SysData.MaxHoliday` default of 32 be a named constant or an inline literal?

**Recommended answer:** Named constant in `internal/types/constants.go` — `const DEFAULT_MAX_HOLIDAY = 32`. Matches the project's convention of C-name-matching constants (the C source uses `sysdata.maxholiday` with no named-constant equivalent; the Go constant name is Go-native).

**Q5.** Should the `do_holidays` listing match C's exact byte output (including tab characters, `&` color codes, `\r\n`)?

**Recommended answer:** Yes, byte-for-byte. Players and automated test harnesses can rely on predictable output. Pin via `TestDoHolidays_SingleHolidayRendersRow` asserting the full output string including the `\t` between name and month name.

**Q6 — NEW (audit-holidays 2026-04-18).** `get_holiday` has two live callers in `src/timezone.c` (guarded by `#ifdef ENABLE_HOLIDAYS`): the `do_time` holiday-today suffix (`:528`) and the `season_update` hour-0 announce (`:622`). Which of these should land in this plan vs defer?

**Recommended answer:** Land (A) `GetHoliday(month, day int) *HolidayData` as a one-liner range over `WorldRef.Holidays` checking `h.Month == month+1 && h.Day == day+1` — trivial, matches the 0-indexed TimeInfo → 1-indexed HolidayData translation visible at C `holidays.c:60`. Land (B) `DoTime` suffix: append `"&wIt's a holiday today:&W <name>\n\r"` when `GetHoliday(TimeInfo.Month, TimeInfo.Day)` returns non-nil — small diff to `internal/act/info2.go:95-126`. Defer (C) season-tick announce (`echo_to_all` on hour-0) unless a Go-side pulse/tick hook for season-update already exists — grep required before landing. If (C) defers, mark as TODO; do not block the plan on it. Add acceptance criteria A14 (`GetHoliday` defined), A15 (`DoTime` emits holiday-today suffix when applicable), and a scope-cut entry for (C) pending the tick-hook grep.

---

## Risk Analysis

**Overall risk: Low.** This is a pure content-port plan with no schema changes to existing types, no cross-cutting refactors, no combat/spell interaction. All primitives shipped.

**Per-risk:**

1. **`month_name[]` placement.** Deciding between `internal/types/months.go` vs `internal/types/calendar.go` vs bundling into `system.go` is aesthetic. Chose `months.go` for file-level clarity; if a future calendar-related port (seasons, days, week structure) consolidates into a single file, a trivial move. **Mitigation:** isolate in its own file so the move cost is zero.

- **`MonthName` 0-indexed vs 1-indexed API.** Chose 0-indexed to match `TimeInfoData.Month`. Mis-use at a call site would render the wrong month name. **Mitigation:** G1 test pins the semantics; every caller must translate from file-format 1-indexed at the boundary; source comment on `MonthName` explains the convention.

2. **C-bug-vs-intent decisions.** Three C bugs identified (`<=1` vs `<1` day/month validation, `create` stores `time_info.month` unshifted). Porting verbatim vs fixing is a judgment call per bug. **Mitigation:** Open Questions section addresses each; Acceptance Criteria A13 pins the deliberate divergences in tests; Go source comments at each divergence site cite the C line and explain the deviation.

3. **`holidays.dat` format compatibility.** Existing shipped file has 2 holidays in tilde-terminated blocks. If C and Go byte-level output diverge, admins upgrading from a C tree to the Go port can't share the file. **Mitigation:** `TestSaveHolidays_RoundTripByteForByte` asserts the output format; `TestLoadHolidays_ShippedFileLoads` pins loading the canonical file. Together these catch any drift.

4. **`MaxHoliday` default-32 regression.** If a future `sysdata.dat` loader sets `MaxHoliday` differently, the hard-coded 32 would override. **Mitigation:** the boot code uses `if w.SysData.MaxHoliday == 0 { ... = 32 }` — a proper loader that sets `MaxHoliday` first will be honored. Tested via `TestBoot_LoadsHolidays`.

5. **Command name collision.** `holidays`, `saveholiday`, `setholiday` are unlikely collisions but check the existing command registry. **Mitigation:** `TestBoot_NoCommandCollision` in existing `boot_test.go` covers this.

6. **Test fixture drift.** Canonical `holidays.dat` shipped at `db/system/holidays.dat` may change in future commits, breaking `TestLoadHolidays_ShippedFileLoads`. **Mitigation:** the test is INTENTIONALLY coupled to the shipped file — any change to the file is a semantic change and should surface via the test. Accept as a feature, not a bug.

7. **Pager test seams.** `DoHolidays` routes through `game.SendToPager`; testing requires a descriptor with pager state. **Mitigation:** existing pager tests in `game/pager_test.go` establish the test harness pattern — copy it.

---

## Adversary Verification Notes

*(Filled after plan adversary pass — see Adversary-resolved concerns section after dispatch.)*

---

## Completion Record

*(Filled after work lands.)*

---

## Relevant File Paths

**C references (authoritative):**
- `/home/eilidh/src/smaug/src/holidays.c:1-416` — full holiday subsystem (entry points, persistence, struct handling)
- `/home/eilidh/src/smaug/src/holidays.h:42-60` — `HOLIDAY_DATA` struct + extern declarations
- `/home/eilidh/src/smaug/src/db.c:692-695` — boot-time `load_holidays()` call (guarded by `ENABLE_HOLIDAYS`)
- `/home/eilidh/src/smaug/src/mud.h:3449` — `int maxholiday;` field in `struct system_data`
- `/home/eilidh/src/smaug/src/mud.h:5522` — `HOLIDAY_FILE` path macro (`SYSTEM_DIR "holidays.dat"`)
- `/home/eilidh/src/smaug/src/act_info.c:2455-2460` — `month_name[]` definition (primary, 17 entries)
- `/home/eilidh/src/smaug/src/timezone.c:101-106` — `month_name[]` definition (duplicate via `ENABLE_TIMEZONE`)
- `/home/eilidh/src/smaug/src/act_info.c:2493` — `do_time` uses `month_name[time_info.month]` directly (0-indexed)
- `/home/eilidh/src/smaug/src/tables.c:717,1340,1390` — C command-table registrations for the three holiday entry points
- `/home/eilidh/src/smaug/src/act_wiz.c:8141-8146` — `cset max-holidays` immortal-tune path (out of scope for this plan)
- `/home/eilidh/src/smaug/db/system/holidays.dat` — shipped data file (14 lines, 2 holidays)

**Go files to create:**
- `/home/eilidh/src/smaug/smaug-go/internal/types/months.go` (G1)
- `/home/eilidh/src/smaug/smaug-go/internal/types/holiday.go` (G1)
- `/home/eilidh/src/smaug/smaug-go/internal/types/months_test.go` (G1, new)
- `/home/eilidh/src/smaug/smaug-go/internal/persist/holidays.go` (G2)
- `/home/eilidh/src/smaug/smaug-go/internal/persist/holidays_test.go` (G2, new)
- `/home/eilidh/src/smaug/smaug-go/internal/persist/testdata/holidays_two.dat` (G2, new fixture)
- `/home/eilidh/src/smaug/smaug-go/internal/persist/testdata/holidays_empty.dat` (G2, new fixture)
- `/home/eilidh/src/smaug/smaug-go/internal/persist/testdata/holidays_malformed.dat` (G2, new fixture)
- `/home/eilidh/src/smaug/smaug-go/internal/act/holidays.go` (G3)
- `/home/eilidh/src/smaug/smaug-go/internal/act/holidays_test.go` (G3, new)

**Go files to modify:**
- `/home/eilidh/src/smaug/smaug-go/internal/world/world.go` (G4 — add `Holidays []*types.HolidayData`)
- `/home/eilidh/src/smaug/smaug-go/internal/boot/boot.go` (G4 — loader wire + command registrations + `MaxHoliday` default)
- `/home/eilidh/src/smaug/smaug-go/internal/boot/boot_test.go` (G4 — post-boot assertions)
- `/home/eilidh/src/smaug/smaug-go/internal/types/constants.go` (G4 — optional `DEFAULT_MAX_HOLIDAY = 32` constant per Q4)

**Go files read for context (not modified):**
- `/home/eilidh/src/smaug/smaug-go/internal/types/system.go:100-110,114-121` (`SystemData.MaxHoliday`, `TimeInfoData.Month`)
- `/home/eilidh/src/smaug/smaug-go/internal/persist/stances.go` (loader shape reference)
- `/home/eilidh/src/smaug/smaug-go/internal/persist/scanner.go` (Scanner API)
- `/home/eilidh/src/smaug/smaug-go/internal/game/pager.go:14` (`SendToPager`)
- `/home/eilidh/src/smaug/smaug-go/internal/act/info2.go:95-126` (`DoTime` current implementation)
- `/home/eilidh/src/smaug/smaug-go/internal/util/strings.go:172` (`IsNumber`)
- `/home/eilidh/src/smaug/smaug-go/internal/util/strings.go` (`OneArgument`)
- `/home/eilidh/src/smaug/smaug-go/internal/types/constants.go:41-59` (`LEVEL_*` gates)

**Doc files to update:**
- `/home/eilidh/src/smaug/smaug-go/doc/plan-phase6-holidays.md` (this file — Adversary Verification Notes + Completion Record appended after landing)
- `/home/eilidh/src/smaug/smaug-go/doc/phase6-roadmap.md` (mark holidays plan attached — orchestrator-level, may happen at merge time)
- `/home/eilidh/src/smaug/CLAUDE.md` (add tier entry for holidays under Phase 6 on landing — via lineage-draft CLAUDE-patch)
- `/home/eilidh/src/smaug/CHANGELOG.md` (landing entry via lineage-draft)
- `/home/eilidh/src/smaug/TODO.md` (deferred follow-up items: `DoTime` calendar upgrade, `DoLoad` umbrella, `DoCset` / sysdata loader, `setholiday name` uniqueness)

---

## Mutation Verification Safety (MANDATORY for worker and adversary prompts)

> Mutation verification must never use destructive git commands. Banned for mutation revert:
> - `git checkout -- <file>`
> - `git checkout <ref> -- <file>`
> - `git restore <file>`
> - `git reset --hard` (any form)
> - `git stash` (any form)
>
> Safe pattern: apply the mutation with the `Edit` tool, run the test to confirm failure, then call `Edit` again with the opposite change to revert. Never touch the working tree with git commands.

This ban applies to every worker subagent and every adversary subagent dispatched against this plan. Include it verbatim in their prompts.
