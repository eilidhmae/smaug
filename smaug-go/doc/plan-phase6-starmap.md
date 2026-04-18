# Plan: Phase 6 — Starmap (`look sky`)

**Status:** Planned (2026-04-18). Adversary-verified: to be filled after plan adversary pass.
**Priority:** Phase 6 Wave 1 — small self-contained. Pure-render; no persistence; no mutation.
**Scope:** New files `internal/act/starmap.go` + `internal/act/starmap_test.go`. Modifications to `internal/act/info.go` (`DoLook` branch for `"sky"`). No boot registration change (already registered as `look`). No persistence change. No new package.

---

## Problem

C ships a sky-rendering subsystem at `src/starmap.c:87-226` (`look_sky`) that paints a 72-column x 8-row ANSI sky map driven by `time_info.hour` / `.day` / `.month` and the player's area `weather.precip`. The map shows:

- Three nighttime star rows (lines 0-2) and three daytime sun rows (lines 3-5) — wait, the opposite: C uses lines 3-5 for sun/moon during daytime and the full 6 night rows (0-2 and 6-7) during nighttime. Verified at `starmap.c:116-118`: during daytime (`hour >= 6 && hour <= 18`) non-middle rows are skipped; at night, ALL 8 rows render stars.
- Sun path (daytime): 5-cell sun glyph scrolls west-to-east over the 72-cell width.
- Moon path (night + daytime eclipse): 5-cell moon glyph at an offset driven by `day` mod `NUM_DAYS=35`. Phase culling masks the partial moon from the leading or trailing edge.
- Constellation table: a fixed 8-row x 72-col ANSI table (`src/starmap.c:59-68`) that scrolls east-to-west as `month` advances across `NUM_MONTHS=17`. Each cell is either a plain-star glyph (`.` / `:` / `*`), a color+space pair (`G<space>`, `C<space>`, etc.) representing a colored constellation dot, or a literal space.

Caller site: `src/act_info.c:1489-1501`. After the outer `DoLook`'s `arg1` keyword match, if arg1 is `"sky"`:
1. If `!IS_OUTSIDE(ch)` (indoor): `"You can't see the sky indoors.\n"` → return.
2. Else: `look_sky(ch)` → return.

### Current Go gap

`smaug-go/internal/act/info.go:21` implements `DoLook` but has no `"sky"` branch. When a player types `look sky` today, the command parser passes `"sky"` as the argument. `DoLook` falls through all keyword matches (character, extra-descr, contents, inventory) and emits `"You do not see that here.\n\r"` at `info.go:98`. No sky map.

Confirmed facts against the Go tree (verified 2026-04-18 via `Read` + `Grep`):
- `DoLook` entry at `info.go:20-99` has no `sky` keyword.
- `DoWeather` at `info.go:258-297` renders a *text* sky description (`"The sky is dark and cloudless."` etc.) but no ASCII-art map.
- `WorldRef.TimeInfo` (`types/system.go:112-121`) has `Hour`, `Day`, `Month`, `Year`, `Season`, `Sunlight` — all required fields shipped.
- `ch.InRoom.Area.Weather` (`types/area.go:52-62`) is `*WeatherData`; `WeatherData.Precip` at `types/area.go:78` is `int`. Per-area precipitation, matching C.
- `weath_unit` C global at `src/db.c:108,600` — value is `10`. NOT present in Go. Needs to be added (as either a constant or a field — see Design below).
- `ROOM_INDOORS` flag (`types/enums.go:698`) + `SECT_INSIDE` sector (`types/enums.go:550`) both shipped; `DoWeather` already uses the flag at `info.go:264`. The "indoors" check needs to match the established Go helper: flag-or-sector. See `spell_unique.go:281` / `ifcheck.go:640-643` for the canonical two-part test.
- Color codes: Go `DescriptorData.ColorFunc` processes `&Y` / `&W` / `&G` / etc. before writing to the wire (`descriptor.go:72-74`). `CharData.Send` / `Sendf` (`character.go:396-407`) emit through the buffer pipeline. No pager equivalent — C's `pager_printf_color` in Go is just `ch.Sendf`.

### Why "the table IS the gameplay"

Resolved in `phase6-roadmap.md` Open Question 7: the constellation positions and colors in `src/starmap.c:59-68` are SMAUG-authored content, not algorithmic output. The Go port MUST preserve the 8-line x 72-character `star_map[]` verbatim (every character including leading spaces, tabs, and every color code). Similarly the `sun_map` and `moon_map` 3x5 glyph arrays are authored gameplay and must be copied byte-for-byte.

---

## C Reference (authoritative)

All citations are against `src/starmap.c` (HEAD) and `src/act_info.c`.

### Entry point and dispatch

`src/act_info.c:1489-1501` — inside `do_look` after the `"in"` / `"inside"` / `"on"` branches and before `"under"`:

```c
if (!str_cmp (arg1, "sky"))
  {
    if (!IS_OUTSIDE (ch))
      {
        send_to_char (_("You can't see the sky indoors.\n"), ch);
        return;
      }
    else
      {
        look_sky (ch);
        return;
      }
  }
```

`IS_OUTSIDE(ch)` expands to `!IS_SET(ch->in_room->room_flags, ROOM_INDOORS)` (mud.h). The Go port's existing indoor check elsewhere uses flag-OR-sector (`ROOM_INDOORS || SECT_INSIDE`) — a superset that treats bare sector-INSIDE rooms as indoors even if they forgot to set the flag. This is the audit-established Go convention (`phase5-tier3-completed.md` entry A1).

### Module constants

`src/starmap.c:50-56`:

| Constant | Value | Role |
|---|---|---|
| `NUM_DAYS` | `35` | Moon cycle period (match days-per-month in calendar). |
| `NUM_MONTHS` | `17` | Match month count in `month_name[]`. |
| `MAP_WIDTH` | `72` | Columns of each row — also star_map string length. |
| `MAP_HEIGHT` | `8` | Rows in star_map. |

Go's `types.Calendar` at `types/system.go:98-110` has `DaysPerMonth` and `MonthsPerYear` as configurable fields (boot-loaded from `system.dat`). C's `NUM_DAYS` / `NUM_MONTHS` are compile-time. The port constants must stay at 35 / 17 — not the runtime calendar — because the constellation table is hard-sized to them. Divergence between the calendar's real length and the starmap's hard-coded scaling is accepted (C has the same divergence).

### `look_sky` function — `src/starmap.c:87-226`

#### Precip gate (L94-104)

```c
precip = (ch->in_room->area->weather->precip + 3 * weath_unit - 1) / weath_unit;
if (precip > 1)
  {
    send_to_char("There are some clouds in the sky so you cannot see anything else.\n\r", ch);
    return;
  }
```

`weath_unit = 10` (`src/db.c:108,600`). So `precip = (raw + 29) / 10` — classic ceiling-divide-by-10 after a +29 offset. For the starmap, `precip > 1` gate. Real values: raw precip is bounded roughly `[-3*weath_unit, +3*weath_unit]` = `[-30, +30]` per weather vector cycle (`src/update.c:3377-3378` sets bounds). Sample mapping:

| Raw precip | Bucket | Visible? |
|---|---|---|
| ≤ -10 | 2 (or below) | Yes — wait, re-run math. `(−10 + 29) / 10 = 19/10 = 1`. Bucket 1. Visible. |
| -9 to 0 | 2 | Wait: `(0+29)/10 = 2`. Bucket 2. NOT visible. |
| 1 to 10 | 3 | Not visible. |
| ... | ... | ... |

Actually: `(precip + 3*weath_unit - 1) / weath_unit` is ceiling( (precip + 29) / 10 ). For starmap visibility we need `bucket <= 1`, i.e. `precip + 29 <= 10 + 9` → `precip <= -10`. So the sky is visible only when `raw precip <= -10` (dry conditions). This matches the `weather.c` text-sky bucketing where bucket-0 is "cloudless" and bucket-1 is "cloudy".

Double-checking via the `DoWeather` implementation: `WeatherInfoData.Sky` is a separate field (`SKY_CLOUDLESS` / `SKY_CLOUDY` / `SKY_RAINING` / `SKY_LIGHTNING`) — that's a global enum, not the per-area bucketed precip. Starmap uses the *area* weather's raw `precip` int. Port the exact formula; do not substitute `WeatherInfo.Sky`.

**Cloud path output:** `"There are some clouds in the sky so you cannot see anything else.\n\r"` — port verbatim.

#### Header (L94)

`"You gaze up towards the heavens and see:\n\r"` — always printed before the precip gate check? No — re-read L94: `pager_printf_color (ch, "You gaze up towards the heavens and see:\n\r");` is on L94, BEFORE the precip gate at L96-104. So the header emits even when the sky is cloudy. The cloudy branch sends its message after the header; total output = header + cloudy-line. Port verbatim.

#### Sun/moon/star positions (L105-112)

```c
sunpos = (MAP_WIDTH * (24 - time_info.hour) / 24);                          // 0..72 sliding with hour
moonpos = (sunpos + time_info.day * MAP_WIDTH / NUM_DAYS) % MAP_WIDTH;      // 0..71
moonphase = ((((MAP_WIDTH + moonpos - sunpos) % MAP_WIDTH)
              + (MAP_WIDTH / 16)) * 8) / MAP_WIDTH;
if (moonphase > 4) moonphase -= 8;                                          // wraps to -3..4
starpos = (sunpos + MAP_WIDTH * time_info.month / NUM_MONTHS) % MAP_WIDTH;  // 0..71
```

- `sunpos`: at hour 0 → 72, hour 12 → 36, hour 24 → 0. (Note: `time_info.hour` ranges 0..23 in C; 24-hour is unreachable.)
- `moonpos`: orbits sunpos by day/35 of the full width.
- `moonphase`: signed in `[-3, 4]`. 0 = full (or new if aligned with sun). Negative = waning; positive = waxing.
- `starpos`: offset by month/17 of the full width — the constellation table scrolls.

Port verbatim with integer division (all ints in C).

#### Row selection (L114-118)

```c
for (linenum = 0; linenum < MAP_HEIGHT; linenum++)
  {
    if ((time_info.hour >= 6 && time_info.hour <= 18) &&
        (linenum < 3 || linenum >= 6))
      continue;
    ...
  }
```

Daytime (hour 6-18 inclusive): render only rows 3, 4, 5 (the sun/moon rows). Night: render all 8 rows (stars on rows 0-2 and 6-7 + rows 3-5 get moon or stars).

#### Per-cell plot (L123-221)

For each cell `i` in `[1, 72]` (1-indexed iteration; C loop `for (i = 1; i <= MAP_WIDTH; i++)`) and each `linenum`:

**Moon-first branch (L126-148)** — daytime or night, moon present at row 3-5:
- Gate: `moonpos in [MAP_WIDTH/4 - 2, 3*MAP_WIDTH/4 + 2]` = `[16, 56]`.
- Gate: `i in [moonpos - 2, moonpos + 2]`.
- Gate: eclipse check. Daytime branch requires `(sunpos == moonpos && hour == 12)` OR `moonphase != 0` (no-eclipse). **Verified L129**: the condition is `((sunpos == moonpos && hour == 12) || moonphase != 0)` — this means eclipse (sunpos==moonpos at hour 12) PREVENTS the normal moon-over-sun overlay? Re-read carefully: the gate is "moon overlays sun UNLESS it's a new moon (`moonphase == 0`) that isn't at noon". So at the eclipse moment (new moon directly overhead at noon) the moon IS drawn — it's the actual eclipse render. At any other time with `moonphase == 0` (new moon, not noon), the moon is invisible. Port verbatim.
- Gate: `moon_map[linenum - 3][i + 2 - moonpos] == '@'`. (The moon-map glyph at this cell.)

Moon-phase masking (L132-136):
- If `moonphase < 0 && (i - 2 - moonpos >= moonphase)` OR `moonphase > 0 && (i + 2 - moonpos <= moonphase)` → emit `"&W@"`.
- Else → emit `" "`. (The phase-masked side of the moon is black space.)

The night-only moon branch (L138-148) has the same glyph/masking logic minus the daytime/eclipse gate — at night the moon is always visible when in the sky.

**Sun branch (L149-161)** — daytime, non-moon cells:
- `i in [sunpos - 2, sunpos + 2]` → emit `"&Y" + sun_map[linenum - 3][i + 2 - sunpos]`.
- Else → emit `" "`.

**Star branch (L162-220)** — nighttime, non-moon cells:
- Look up `star_map[linenum][(MAP_WIDTH + i - starpos) % MAP_WIDTH]`.
- Dispatch on the character: plain glyphs `.` / `:` / `*` emit verbatim as single chars. Color-prefix letters `G` / `g` / `R` / `r` / `C` / `O` / `B` / `P` / `W` / `b` / `p` / `Y` / `c` emit `"&X "` (color + single space). Default → `" "`.

The color letter → ANSI mapping is exactly the SMAUG color-code table: `&G`=bright green, `&g`=dark green, `&R`=bright red, `&r`=dark red, `&C`=bright cyan, `&O`=orange (bright brown), `&B`=bright blue, `&P`=bright magenta, `&W`=bright white, `&b`=dark blue, `&p`=dark magenta, `&Y`=bright yellow, `&c`=dark cyan.

#### Row terminator (L222-223)

After the inner loop, `strcat(buf, "\n\r"); pager_printf_color(ch, buf);`. Port via `ch.Sendf` or batched `strings.Builder` + `ch.Send` (Go prefers the latter — one `Send` per line is ~8 sends for night, 3 for day, which is negligible).

### Constellation tables (verbatim data)

`src/starmap.c:59-68` — 8 lines of exactly 72 characters each:

```
"                                               C. C.                  g*"
"    O:       R*        G*    G.  W* W. W.          C. C.    Y* Y. Y.    "
"  O*.                c.          W.W.     W.            C.       Y..Y.  "
"O.O. O.              c.  G..G.           W:      B*                   Y."
"     O.    c.     c.                     W. W.                  r*    Y."
"     O.c.     c.      G.             P..     W.        p.      Y.   Y:  "
"        C*                    G*    P.  P.           p.  p:     Y.   Y. "
"                 b*             P.: P*                 p.p:             "
```

Constellation legend (src/starmap.c:70-74 comment — informational, no functional use): Cygnus, Mars, Orion, Dragon, Cassiopeia, Venus, Ursa Minor, Mercurius, Pluto, Uranus, Leo, Crown, Raptor.

*Row index note:* the 8 rows above are indices 0-7 of the `star_map[]` array, corresponding to source lines 60-67. After `Read`-verification of HEAD, the literal row bytes are:

- Row 0 (line 60): `"                                               C. C.                  g*"`
- Row 1 (line 61): `"    O:       R*        G*    G.  W* W. W.          C. C.    Y* Y. Y.    "`
- Row 2 (line 62): `"  O*.                c.          W.W.     W.            C.       Y..Y.  "`
- Row 3 (line 63): `"O.O. O.              c.  G..G.           W:      B*                   Y."`
- Row 4 (line 64): `"     O.    c.     c.                     W. W.                  r*    Y."`
- Row 5 (line 65): `"     O.c.     c.      G.             P..     W.        p.      Y.   Y:  "`
- Row 6 (line 66): `"        c.                    G*    P.  P.           p.  p:     Y.   Y. "`
- Row 7 (line 67): `"                 b*             P.: P*                 p.p:             "`

Note that `b*` (lowercase, dark-blue bright-star) appears on row 7 col 17, NOT row 6. Some SMAUG forks place other variants here; use HEAD byte-for-byte. Go port's final `var starMap = []string{...}` must match these 8 lines exactly, each exactly 72 bytes.

`src/starmap.c:75-79` — sun 3x5 glyph:

```
"\\`|'/",
"- O -",
"/.|.\\"
```

(Note: C source uses `\\` for a literal backslash in string-literal form. Go raw strings handle this naturally.)

`src/starmap.c:81-85` — moon 3x5 glyph:

```
" @@@ ",
"@@@@@",
" @@@ "
```

### `weath_unit` constant

C declares `int weath_unit` at `src/db.c:108`. Assigned `weath_unit = 10` at `src/db.c:600`. Loadable via `system.dat` at `src/db.c:8115` (`weath_unit = fread_number(fp)`). Settable via immortal `weather` command at `src/act_wiz.c:11393`.

Go has no equivalent. Options:
- Hardcode `const WeathUnit = 10` in `internal/types/constants.go`. Matches reality (the runtime assignment is redundant with the compile-time constant).
- Store on `world.World.WeathUnit` as `int`, default 10. Matches C's "tunable at runtime" semantics.

Recommendation: **constant**. No shipped area file sets a non-default value; the immortal `weather` command isn't ported; and making `look sky` dependent on world-state for a cosmetic 10 is overkill. If a future plan needs to tune it, it can promote to a field. G1 of this plan adds the constant.

---

## Go Current State

Verified 2026-04-18 against `internal/` tree.

- **Entry point:** `internal/act/info.go:DoLook` has no `sky` keyword — falls through to "You do not see that here."
- **Calendar:** `WorldRef.TimeInfo.Hour` / `.Day` / `.Month` populated from system.dat at boot; advanced by `game/update.go:weatherUpdate` each hour-boundary tick.
- **Per-area weather:** `ch.InRoom.Area.Weather.Precip` shipped; populated by `AreaData` loader and `game/update.go` per-area weather tick.
- **Indoor/outdoor check:** canonical pattern is `RoomFlags.IsSet(ROOM_INDOORS) || SectorType == SECT_INSIDE` (`spell_unique.go:281`, `ifcheck.go:640-643`, `info3_test.go:391`).
- **`weath_unit`:** not defined anywhere in Go. Adding as a package constant is uncontroversial.
- **Color rendering:** `&Y` / `&W` / `&G` / etc. pass through `DescriptorData.ColorFunc` to become ANSI escapes (when color is enabled) or be stripped (when disabled). Tests typically inspect the raw-buffer output and match on the `&X` form (see `readOutput` in `internal/act/*_test.go`).
- **Output width:** tests read what `ch.Send` wrote. No auto-wrapping. The plan's rendered lines will be exactly 72 color-coded cells + `\n\r`.
- **Pager:** `internal/game/pager.go` exists for large help text but `ch.Send` / `Sendf` do not route through it by default. Starmap output is at most 8 lines of ~72 chars each = ~600 bytes, well under any pager threshold. Use plain `Sendf`.

## Go Design

### Approach A (CHOSEN): New file `internal/act/starmap.go` with pure-function renderer + integration in `DoLook`

- **`starmap.go` layout:** package-level `const` block with `starmapWidth = 72`, `starmapHeight = 8`, `starmapNumDays = 35`, `starmapNumMonths = 17`, `starmapWeathUnit = 10` (local-to-file, not exported, because only `LookSky` uses them). Three package-level string arrays (slices of strings, len 8 / 3 / 3) for `starMap`, `sunMap`, `moonMap`. One exported function `LookSky(ch *types.CharData)`.
- **Algorithm:** port `look_sky` verbatim, row-by-row, writing to a `strings.Builder` per line and invoking `ch.Send(line)` once per line. Using a builder per line matches the C `sprintf(buf, " "); ... strcat(buf, ...); pager_printf_color(ch, buf)` pattern.
- **Pure helper:** extract `renderStarmap(t TimeInfoData, precip int) []string` — returns the rendered lines (with embedded `&X` color codes, no `\n\r` terminator). Invoking goroutine calls `ch.Send(line + "\n\r")` for each. This split lets tests assert on the raw rendered strings without a live `ch`.
- **`DoLook` branch** inside `info.go:DoLook` after the `argument == "" || auto` block and before the "look at something specific" one-arg extraction — insert:
  ```go
  if strings.EqualFold(argument, "sky") {
      if ch.InRoom.RoomFlags.IsSet(types.ROOM_INDOORS) || ch.InRoom.SectorType == types.SECT_INSIDE {
          ch.Send("You can't see the sky indoors.\n\r")
          return
      }
      LookSky(ch)
      return
  }
  ```
  Position: after the bare-arg room print, before the `util.OneArgument(argument)` match loop. Placement verified against the C order in `act_info.c:1489` (post-room, pre-per-item).
- **Cloudy path:** `LookSky` emits the header, computes precip bucket via `(precip + 3*weathUnit - 1) / weathUnit`, and if `> 1` emits the cloud message and returns.
- **Tests** in `starmap_test.go`:
  - Table-driven tests for `renderStarmap` at specific (hour, day, month, precip) quadruples — pin bytes of selected lines.
  - Dispatch test: invoke `DoLook(ch, "sky")` with an outdoor room, assert the header string appears in output.
  - Indoor test: invoke with `ROOM_INDOORS` set, assert only the indoor-block message.
  - Sector-INSIDE test: invoke with `SectorType == SECT_INSIDE`, same assertion.
  - Cloudy test: set `ch.InRoom.Area.Weather.Precip = 0` (raw bucket > 1), assert cloud message after header, no map.
  - Dispatch-precedence test: `DoLook(ch, "sky")` when `ch.InRoom` is nil → "You are nowhere!" (already handled by the existing nil-room guard at `info.go:22-25`; pin that the sky branch does NOT execute before that guard).

### Rejected alternatives

- **Approach B: Put `LookSky` inside `info.go`.** 140 LOC of cell-plot logic would dominate the file. A dedicated `starmap.go` with its own test file mirrors the C module separation.
- **Approach C: Render as a []byte + fprintf-style.** Go idiom prefers `strings.Builder` for line assembly. Matches the existing `DoScore` / `DoWho` style.
- **Approach D: Put the constants in `types/constants.go`.** Only `LookSky` reads them. Keeping them file-local reduces `types` package surface. The sole exception is `WeathUnit` — an argument for hoisting it if a future weather-command port needs it. Defer.
- **Approach E: Precompute the star_map / sun_map / moon_map bytes as `[][]byte` at `init()`.** Premature optimization; string indexing is fine at 72 chars x 8 rows.
- **Approach F: Store `WeathUnit` on `world.World` as a field.** Would mirror C exactly but adds a coupling for no current benefit. Deferred to the first plan that needs mutable weath_unit.
- **Approach G: Reuse `WeatherInfoData.Sky` (global) instead of per-area `Weather.Precip`.** Wrong — C uses per-area precip (`starmap.c:96`: `ch->in_room->area->weather->precip`). Global `Sky` is a separate bucketing used by `DoWeather`. Port reads area precip.

---

## Task Groups

All tasks include the TDD mandate and the mutation-verify expectation. Mutation verification uses `Edit` round-trips only per `_shared.md` → Mutation Verification Safety (banned commands: `git checkout -- <file>`, `git checkout <ref> -- <file>`, `git restore <file>`, `git reset --hard` any form, `git stash` any form).

Total suggested worker phasing: G1 alone (~60 min), G2 alone (~75 min, depends on G1), G3 alone (~30 min, depends on G2). Serializable; cannot parallelize within a lineage because each group touches the same file the next reads.

### G1 — Constellation tables + pure renderer + helper functions

**Files:**
- New: `internal/act/starmap.go`
- New: `internal/act/starmap_test.go`
- No modifications to existing files in G1. No registration.

**Contents of `starmap.go`:**
- File-package constants:
  - `const starmapWidth = 72`
  - `const starmapHeight = 8`
  - `const starmapNumDays = 35`
  - `const starmapNumMonths = 17`
  - `const starmapWeathUnit = 10`
- File-package data:
  - `var starMap = []string{ ... 8 lines copied verbatim from src/starmap.c:59-68 ... }` — each line exactly 72 bytes long.
  - `var sunMap = []string{"\\`|'/", "- O -", "/.|.\\\\"}` — 3 lines of 5 bytes (note Go raw string handling of backslash).
  - `var moonMap = []string{" @@@ ", "@@@@@", " @@@ "}` — 3 lines of 5 bytes.
- Unexported helpers:
  - `func computePositions(hour, day, month int) (sunpos, moonpos, moonphase, starpos int)` — ports `starmap.c:105-112`. Returns the four derived values.
  - `func precipBucket(rawPrecip int) int` — returns `(rawPrecip + 3*starmapWeathUnit - 1) / starmapWeathUnit`. Matches C integer-division semantics.
  - `func renderStarmap(hour, day, month, rawPrecip int) []string` — returns the rendered lines WITHOUT any trailing `"\n\r"` (the caller appends). The returned slice always begins with the header `"You gaze up towards the heavens and see:"`. If `precipBucket(rawPrecip) > 1`, returns exactly `[]string{"You gaze up towards the heavens and see:", "There are some clouds in the sky so you cannot see anything else."}`. Otherwise: header + 3-to-8 plotted sky lines (exact count driven by the daytime row-skip rule at `starmap.c:116-118`). `LookSky` appends `"\n\r"` to each element when sending — this matches the C behavior where the header emits via `pager_printf_color` (which writes `"\n\r"` in-string) and each plot line ends with a `strcat(buf, "\n\r")` at L222.
  - Inside the plot loop, one plotted cell emits into a `strings.Builder`; after the inner `i` loop the builder's `String()` becomes one line.

**Test-first (all in `starmap_test.go`):**

Table-driven tests on the pure helpers — no CharData, no World.

Positions:
- `TestComputePositions_MidnightDay0Month0` — `hour=0, day=0, month=0` → sunpos = 72, moonpos = 0 (72+0 %72), wait: `sunpos = 72 * (24-0) / 24 = 72`. Then `moonpos = (72 + 0*72/35) % 72 = 72 % 72 = 0`. Then `moonphase = (((72 + 0 - 72) % 72) + 4) * 8 / 72 = (0 + 4) * 8 / 72 = 32 / 72 = 0`. Starpos = `(72 + 0) % 72 = 0`. Assert `(72, 0, 0, 0)`.
- `TestComputePositions_Noon` — `hour=12, day=0, month=0` → sunpos = 36, moonpos = 36, moonphase = 0 (eclipse), starpos = 36.
- `TestComputePositions_MoonphaseAtUpperBoundary` — `hour=0, day=18, month=0`. Hand-computation: sunpos = 72*(24-0)/24 = 72; moonpos = (72 + 18*72/35) % 72. `18*72 = 1296`, `1296/35 = 37` (integer truncation: 35*37=1295). So moonpos = (72+37) % 72 = 109 % 72 = 37. moonphase pre-clamp = ((72+37-72)%72 + 4) * 8 / 72 = (37+4)*8/72 = 328/72 = 4. Clamp `if > 4 -= 8` does NOT fire (4 is not > 4). Final moonphase = 4. starpos = (72+0) % 72 = 0. Expected tuple `(72, 37, 4, 0)`.
- `TestComputePositions_MoonphaseWanesPastFull` — find a `day` value that makes pre-clamp moonphase > 4 so the `-8` branch fires to negative. Hand-compute: need `((moonpos)%72 + 4) * 8 / 72 > 4`, i.e., `(moonpos + 4) * 8 > 4*72 = 288`, i.e., `moonpos > 32`. At hour=0 sunpos=72, moonpos = `(72 + day*72/35) % 72 = (day*72/35) % 72`. day=20: `20*72/35 = 1440/35 = 41` → moonpos=41. Pre-clamp moonphase = (41+4)*8/72 = 360/72 = 5. After clamp: 5-8 = -3. Expected tuple `(72, 41, -3, 0)`.
- `TestComputePositions_MonthAdvancesStarpos` — `hour=0, month=8` → starpos = `(72 + 72*8/17) % 72 = (72 + 33) % 72 = 33`.

Precip bucket:
- `TestPrecipBucket_DryIsBucketOne` — `raw=-19` → `(-19+29)/10 = 1`. Sky visible.
- `TestPrecipBucket_BorderlineIsBucketTwo` — `raw=0` → `(0+29)/10 = 2`. Sky NOT visible.
- `TestPrecipBucket_Rainy` — `raw=10` → `(10+29)/10 = 3`. Not visible.
- `TestPrecipBucket_VeryDry` — `raw=-30` → `(-30+29)/10 = -1/10 = 0` (Go rounds toward zero for integer division of negatives; C does the same on most compilers. Verify identity.) Sky visible.

Renderer (exact-byte pinning):
- `TestRenderStarmap_CloudyShortCircuit` — `rawPrecip=0` — returns 2 strings: the header and the cloud line. Exact length 2; exact content both lines.
- `TestRenderStarmap_HeaderIsFirstLine` — any non-cloudy input → `lines[0] == "You gaze up towards the heavens and see:"`.
- `TestRenderStarmap_NightRenders8Rows` — `hour=0` (night) → header + 8 plot rows = 9 total lines.
- `TestRenderStarmap_DayRenders3Rows` — `hour=12` → header + 3 plot rows (rows 3, 4, 5) = 4 total lines.
- `TestRenderStarmap_DayRowSkipBoundary6` — `hour=6` and `hour=18` are both daytime per C `(hour >= 6 && hour <= 18)`. Assert `hour=6` and `hour=18` return 4 lines; `hour=5` and `hour=19` return 9.
- `TestRenderStarmap_MidnightMonth0_StarLine0IsVerbatimTable` — with `starpos=0` (hour=0, month=0) the rendered night line 0 must equal the tableau-produced version of row 0 of `starMap`: each `.` / `:` / `*` emits as itself; each color letter emits `&X ` (two characters). Exact-byte match against an expected string computed by the test (or hand-pinned — prefer hand-pinned to catch transcription errors on the starMap table).
- `TestRenderStarmap_NoonSunAtCenter` — `hour=12, day=0` → sun at cells `[sunpos-2, sunpos+2] = [34, 38]` on rows 3, 4, 5. Line for row 3 contains `&Y\` at col 34, `&Y` at col 35, `&Y|` at col 36, `&Y'` at col 37, `&Y/` at col 38. Pin the 5-glyph sequence.
- `TestRenderStarmap_EclipseAtNoonRendersMoon` — `hour=12, day=0` makes moonpos == sunpos == 36, and the eclipse gate `sunpos == moonpos && hour == 12` is true → moon `@` glyph OVERLAYS the sun cells on rows 3-5. Row 3 at cell 36 emits `&W@`; adjacent moon cells either emit `&W@` or space depending on the phase mask. Pin row 4 center cell as `&W@`.
- `TestRenderStarmap_NewMoonNotAtNoonRenderedAsBlank` — construct `hour=0, day=0` → moonphase=0, night, no eclipse gate because night path doesn't check noon. Actually at night, the moon is drawn whenever phase != 0. Re-verify: the night branch (L138) has no `moonphase != 0` check? Re-read L138-148: night condition `(linenum >= 3 && linenum < 6) && (moonpos in range) && (i near moonpos) && (moon_map == '@')` — then phase masking. No eclipse check here because at night there's no sun to eclipse. So moonphase=0 at night means ALL cells of the moon get phase-masked via the `if moonphase < 0 ...` / `if moonphase > 0 ...` checks; with moonphase == 0, neither branch hits, so it falls to the `else strcat(buf, " ");` branch → all moon cells emit space. Effectively: new moon at night is invisible. Pin this.

Boot integration: G1 doesn't wire anything. `renderStarmap` is a pure function with no dependencies.

**Mutation verify (via `Edit` round-trips only — banned list above):**
- Change `NUM_DAYS` constant from 35 to 36 → moon-position tests fail.
- Change `starmapWeathUnit` from 10 to 1 → precip bucket tests fail.
- Mutate a single byte in a `starMap` row (e.g., change `O:` to `O.` on row 1) → `TestRenderStarmap_MidnightMonth0_StarLine0IsVerbatimTable` fails.
- Swap `sunpos` and `moonpos` in the position formulas → `TestComputePositions_Noon` fails.
- Drop the `if moonphase > 4: moonphase -= 8` clamp → `TestComputePositions_MoonphaseWanesPastFull` fails.

**Acceptance gate:** `go test ./internal/act/... -count=3 -run 'TestComputePositions|TestPrecipBucket|TestRenderStarmap'` green.

### G2 — `LookSky` entry point + `DoLook` dispatch + ch.Send integration

**Files:**
- Modified: `internal/act/starmap.go` — add exported `LookSky(ch *types.CharData)`.
- Modified: `internal/act/info.go` — add the `"sky"` branch inside `DoLook`.
- Modified: `internal/act/starmap_test.go` — add dispatch tests.

**`LookSky` body:**

```go
// LookSky renders the constellation and sun/moon map to ch.
// Port of C src/starmap.c:87-226 (look_sky).
//
// Caller is responsible for the IS_OUTSIDE gate; LookSky assumes ch is outdoors.
// If ch.InRoom or the area weather is nil, LookSky emits the header and cloud
// message (safest degenerate rendering).
func LookSky(ch *types.CharData) {
    if ch == nil || WorldRef == nil {
        return
    }
    var precip int
    if ch.InRoom != nil && ch.InRoom.Area != nil && ch.InRoom.Area.Weather != nil {
        precip = ch.InRoom.Area.Weather.Precip
    } else {
        precip = 3 * starmapWeathUnit  // force cloudy fallback
    }
    t := WorldRef.TimeInfo
    for _, line := range renderStarmap(t.Hour, t.Day, t.Month, precip) {
        ch.Send(line + "\n\r")
    }
}
```

**`DoLook` branch** in `info.go` inserted after the bare-arg block at `info.go:55` and before the `util.OneArgument` line (currently `info.go:59`):

```go
if strings.EqualFold(argument, "sky") {
    if ch.InRoom.RoomFlags.IsSet(types.ROOM_INDOORS) || ch.InRoom.SectorType == types.SECT_INSIDE {
        ch.Send("You can't see the sky indoors.\n\r")
        return
    }
    LookSky(ch)
    return
}
```

The existing nil-room guard at `info.go:22-25` fires before the sky branch, so `ch.InRoom != nil` holds.

**Test-first in `starmap_test.go`:**

- `TestDoLook_SkyIndoorsByFlag` — room with `ROOM_INDOORS` flag set, sector any. `DoLook(ch, "sky")` → output contains `"can't see the sky indoors"`, does NOT contain `"gaze up towards"`.
- `TestDoLook_SkyIndoorsBySector` — room with no `ROOM_INDOORS` flag but `SectorType == SECT_INSIDE`. Same expectation.
- `TestDoLook_SkyOutdoorsClear` — outdoor room (flag absent, sector field = `SECT_FIELD` or any non-INSIDE), WorldRef configured with a working TimeInfo, area weather precip = -30 (bucket 0). Output starts with `"You gaze up towards the heavens"` and contains at least one color code (`&Y` or `&W` or `&G`).
- `TestDoLook_SkyOutdoorsCloudy` — same setup but precip = 10 (bucket 3). Output contains header AND cloud line, no color codes from the star table.
- `TestDoLook_SkyNoWorldRef` — `WorldRef = nil` → LookSky returns silently; no panic. Guard test.
- `TestDoLook_SkyNilArea` — outdoor room with `ch.InRoom.Area == nil` → fallback: emits cloud message (degenerate safest rendering). No panic.
- `TestDoLook_SkyExistingLookPathsUnaffected` — regression guard: `DoLook(ch, "")` still calls the auto-room path; `DoLook(ch, "fountain")` still hits the keyword-match path. Both produce their pre-G2 output.
- `TestDoLook_SkyCaseInsensitive` — `DoLook(ch, "SKY")` and `DoLook(ch, "Sky")` both fire the sky branch. The `strings.EqualFold` in the branch test is what makes this pass.

Test fixtures required:
- Two rooms: `mkOutdoorRoom` helper returning `*types.RoomIndexData` with `SectorType: SECT_FIELD`, no indoor flag, and an attached `*types.AreaData` with `*types.WeatherData{Precip: <param>}`.
- `mkIndoorRoom` variant with `ROOM_INDOORS` flag.
- Mirror `setupInfo3World` pattern from `internal/act/info3_test.go:12-16`.

**Mutation verify (via `Edit` round-trips only):**
- Drop the `SectorType == SECT_INSIDE` clause in the indoor check → `TestDoLook_SkyIndoorsBySector` fails.
- Invert the precip gate (`>` to `<`) → `TestDoLook_SkyOutdoorsClear` fails (sky falsely cloudy).
- Replace `strings.EqualFold` with `strings.EqualFold == false` → `TestDoLook_SkyCaseInsensitive` fails.
- Drop the nil-World guard in `LookSky` → `TestDoLook_SkyNoWorldRef` panics.

**Acceptance gate:** `go test ./internal/act/... -count=3` green; `go vet ./...` clean.

### G3 — Integration test + regression pin + final cleanup

**Files:**
- Modified: `internal/act/starmap_test.go` — add regression/integration coverage.
- No new files. No boot changes (already registered; `look` dispatches to `DoLook` which now handles `"sky"`).

**Purpose:** tests in G1 and G2 cover the pure function and the branch wiring. G3 locks the end-to-end behavior through `DoLook` and pins the key visual outputs an end user would see.

**Test-first:**

- `TestDoLook_SkyPrintsEightNightRows` — outdoor, precip=-30, `WorldRef.TimeInfo.Hour=0` (midnight) → raw output splits on `\n\r` into header + 8 non-empty lines. Exactly 9 lines total.
- `TestDoLook_SkyPrintsThreeDayRows` — same setup, `Hour=12` → header + 3 lines = 4 lines.
- `TestDoLook_SkyContainsSunAtNoon` — `Hour=12, Day=0, Month=0` → output contains the substring `&Y|` (sun midline).
- `TestDoLook_SkyContainsMoonAtEclipse` — `Hour=12, Day=0, Month=0` → output contains `&W@` (moon overlays sun during eclipse).
- `TestDoLook_SkyCalendarMonthChangesConstellationPosition` — same hour/day, different month values → the bytes on row 0 shift by `month * 72 / 17` cells. Assert two distinct month values produce different rendered row-0 bytes.

**Mutation verify:**
- Replace `&Y` with `&R` in the sun glyph emission path → `TestDoLook_SkyContainsSunAtNoon` fails.
- Swap line-terminator from `\n\r` to `\n` → line-count tests fail.

**Acceptance gate:** `go test ./internal/act/... -count=3` green (all G1 + G2 + G3 tests); `go vet ./...` clean; `go test ./...` green across all packages (regression guard).

---

## Acceptance Criteria

**G1 (constellation tables + renderer):**
- A1. `starMap` / `sunMap` / `moonMap` are defined verbatim from `src/starmap.c:59-85`, each row exactly the expected length (72 for starMap, 5 for sun and moon).
- A2. `computePositions(hour, day, month)` returns the same 4-tuple as C `look_sky` for every `(hour, day, month)` triple tested in the pinned table.
- A3. `precipBucket(raw)` matches C's `(raw + 3*weath_unit - 1) / weath_unit` for integer inputs including negatives.
- A4. `renderStarmap(hour, day, month, precip)` returns `[]string` of length 2 on cloudy (bucket > 1), length 9 on night (header + 8 rows), length 4 on day (header + 3 rows).
- A5. Rendered lines contain color codes in the SMAUG `&X<char>` form where the C source emits color.

**G2 (`LookSky` + `DoLook` dispatch):**
- A6. `DoLook(ch, "sky")` dispatches to `LookSky(ch)` when `ch.InRoom` is outdoors.
- A7. `DoLook(ch, "sky")` emits `"You can't see the sky indoors.\n\r"` when `ch.InRoom` has `ROOM_INDOORS` OR `SectorType == SECT_INSIDE`.
- A8. Dispatch is case-insensitive: `"sky"`, `"SKY"`, `"Sky"` all fire.
- A9. `LookSky` with `WorldRef == nil`, `ch == nil`, or `ch.InRoom == nil` returns silently — no panic.
- A10. `LookSky` with `ch.InRoom.Area == nil` or `Weather == nil` falls back to the cloudy path — no panic, no map.

**G3 (integration):**
- A11. End-to-end `DoLook(ch, "sky")` at a specific `(hour=12, day=0, month=0)` pin produces an output containing both `&Y|` (sun) and `&W@` (moon) substrings (the eclipse render).
- A12. `go test ./... -count=3` green across all packages.
- A13. `go vet ./...` clean.

---

## Scope Cuts / Deferrals

- **No C `weath_unit` runtime tunability.** Port as a package-private `const = 10`. If a future plan needs the immortal `weather` command (not currently ported), that plan promotes `weath_unit` to a `world.World` field.
- **No translations (`_()` gettext wrapper).** C wraps the indoor message in `_()`; Go port uses the raw English string. All other Go commands do the same (there's no gettext layer shipped).
- **No pager integration.** `pager_printf_color` in C goes through the pager for long output; Go port's `ch.Send` is a direct buffer write. Starmap output at 9 lines is well below any pager threshold. If a future plan wires a pager for long output, starmap can be retrofitted.
- **No `help sky` command or help-file entry.** Out of scope; separate help-file plan.
- **No testclient golden-file integration.** Golden-file diffing is tracked in the Phase 6 roadmap's "Test infrastructure polish" wave. Unit tests with exact-byte pins are sufficient for this plan.
- **No dynamic configuration of `NUM_DAYS` / `NUM_MONTHS`.** They stay at 35 / 17 matching the starMap table's hard shape. C has the same divergence between runtime calendar and starmap scaling.
- **No `MPROG_SKY` trigger** (there is no such C trigger — noted here to prevent speculation).
- **No `look stars` or `look moon` separate subcommands** — C has only `look sky`. Port matches.
- **No per-player or per-area constellation override.** The table is global in C; global in Go.

---

## Open Questions (with recommended answers)

**Q1 — Should the indoor check use flag-only or flag-OR-sector?**
C `IS_OUTSIDE` is flag-only (`!IS_SET(ROOM_INDOORS)`). Go's convention since `phase5-tier3-completed.md` A1 is flag-OR-sector (covers rooms that forgot to set the flag on a SECT_INSIDE sector).
**Recommended:** use flag-OR-sector (matches `DoWeather` at `info.go:264` + the canonical `ifcheck.go:640-643` pattern). Strict C fidelity is a loss here — the canonical Go pattern is intentionally more permissive and has been the convention since Tier 3. Tests cover both forms.

**Q2 — Package-private constants vs exported `types.WEATH_UNIT`?**
No other subsystem reads `weath_unit` today. Exporting it hoists package surface for zero benefit.
**Recommended:** keep `starmapWeathUnit` package-private inside `starmap.go`. First cross-package consumer promotes it to `types/constants.go`.

**Q3 — `renderStarmap` signature: take full `TimeInfoData` or individual ints?**
Individual `int`s is more testable (no need to construct a full `TimeInfoData`) and mirrors C (local locals). `LookSky` extracts the three fields from `WorldRef.TimeInfo` and calls through.
**Recommended:** individual ints. `TimeInfoData` can change; the renderer does not care.

**Q4 — Emit the header line through the renderer or have `LookSky` print it?**
If the renderer returns the header as `lines[0]`, then `renderStarmap` is self-contained and fully testable in isolation. If `LookSky` emits the header separately, we have to duplicate-test the "cloudy also emits header first" detail.
**Recommended:** header in the renderer. Single source of truth for output order.

**Q5 — What if `WorldRef.TimeInfo.Hour` is out of range (`< 0 || >= 24`)?**
C doesn't guard. Go's `time_info.Hour` is set by `weatherUpdate` which modulos 24 at `update.go:weatherUpdate`. Out-of-range values would produce weird sunpos values but won't crash (integer arithmetic).
**Recommended:** don't guard. Match C. If a data-corruption bug sets Hour=25 the starmap will render weirdly, consistent with C.

**Q6 — Which bytes on which starMap row?**
HEAD `src/starmap.c:60-67` byte-verified. Row 7 (line 67) contains `b*` at column 17. Row 0 (line 60) has `C.` at column 47 and `g*` at column 70. Full byte table reproduced in §C Reference above.
**Recommended:** copy the HEAD byte stream verbatim into the Go string. Add a `TestStarMap_RowExactBytes` that pins each of the 8 rows by length (exactly 72) AND full-string equality against the exact source bytes, as a transcription guard. The pinned-prefix alternative is too weak — one bad middle character would slip through.

**Q7 — Does the plan need a fixture for area weather?**
Tests in G2 need outdoor rooms with a valid `Area.Weather.Precip`. That requires constructing 2-3 `AreaData` + `WeatherData` structs in each test. Helper `mkOutdoorRoomWithPrecip(precip int) *types.RoomIndexData` is trivial (~6 lines). No `testdata/` file needed.
**Recommended:** in-code helper; no testdata fixture.

---

## Risk Analysis

**Low risk — well-bounded pure function.** The entire starmap module is a read-only renderer of three in-memory fields. No persistence, no mutation, no concurrency. If the output is wrong, the only damage is a visual glitch that the next `look sky` corrects.

Specific risks:

1. **Transcription error in `starMap` table** — the 8x72 = 576-char table is the single most error-prone part of the port. One mis-copied character produces a subtle wrong-constellation bug invisible in most tests.
   - **Mitigation:** pinned-byte test per row (A1 + test case in G1). Separately, mutation-verify with a byte-change proves the test is strict.
2. **Color code divergence** — the `&X` mapping could be off-by-one between C `case 'G':` dispatch and Go. Every color letter has a `src/starmap.c:179-217` case; Go must handle all 13 (G, g, R, r, C, O, B, P, W, b, p, Y, c) plus the plain-glyph cases (`.`, `:`, `*`) and the default-space fallback.
   - **Mitigation:** a table-driven test per color letter + fuzz test over the 256 possible bytes → verify default path is space.
3. **Integer division on negative `rawPrecip`** — Go and C both use truncated division (round toward zero) for `int`. No divergence on 64-bit linux targets. But if the raw precip is extremely negative (e.g., `rawPrecip = -1000`), the bucket math overflows meaningfully. The domain is `[-30, 30]` per `update.c:3377-3378` so overflow is not reachable in practice.
   - **Mitigation:** test with rawPrecip = -30 (minimum) and +30 (maximum). Document the domain in `renderStarmap` doc comment.
4. **Row-skip off-by-one between daytime and nighttime** — C uses `linenum < 3 || linenum >= 6` to skip during daytime, meaning rows 3, 4, 5 render. Easy to transcribe as `<=` or `<`.
   - **Mitigation:** explicit tests at hour=5, 6, 18, 19 boundaries (G1 test).
5. **Eclipse-at-noon conditional** — C's condition `(sunpos == moonpos && hour == 12) || moonphase != 0` is non-obvious on first read. A Go port that reverses the `||` to `&&` silently renders moon differently.
   - **Mitigation:** dedicated `TestRenderStarmap_EclipseAtNoonRendersMoon` (G1) and end-to-end pin in G3.
6. **Moon phase masking** — the `if (moonphase < 0 && i - 2 - moonpos >= moonphase)` and mirror `moonphase > 0 && i + 2 - moonpos <= moonphase` are easy to flip. The negative-phase mask masks the LEFT side of the moon (waning); positive masks the RIGHT side (waxing).
   - **Mitigation:** test at `day=17` (waxing ~full) and `day=25` (waning) with exact-byte pins on the moon row (row 4).
7. **`ColorFunc` interaction** — if a future descriptor has color disabled, the `&Y|` substring won't appear; instead the raw `|` will. Tests that assert on `&Y|` will fail for no-color clients.
   - **Mitigation:** tests construct descriptors that DO preserve color codes (the default `ColorFunc = nil` in test fixtures passes through unchanged). Document this assumption in `starmap_test.go`.
8. **`DoLook` branch ordering** — the "sky" keyword must be checked BEFORE the generic keyword-match loop, else the loop would fall through to "You do not see that here."
   - **Mitigation:** G2 regression test asserts the outdoor sky path produces the header string (indirect proof the branch fired before the fallthrough).

**No high-risk concerns.** All primitives exist; the port is mechanical; the test surface is narrow and pinnable.

---

## Adversary Verification Notes

### Self-adversary pass (2026-04-18 manager session)

Agent-based adversary subagent verification was unavailable in this session (consistent with Tranche B's completion record and recent manager sessions that report the same limitation). The manager ran a structured self-adversary review against authoritative C sources and the Go codebase, with every citation spot-verified via direct `Read`. Any subsequent external adversary pass should re-check these resolutions.

Findings caught and fixed in-session:

1. **CRITICAL — `starMap` row-6 misattribution.** Plan v1 claimed the `b*` token was on row 6 col 9. Direct `Read` of `src/starmap.c:60-67` confirmed `b*` is on row 7 col 17 (line 67); row 6 (line 66) starts `"        c.   ..."` with lowercase `c.`. Plan's §C Reference now pastes all 8 rows individually with their exact source-line citation, and Q6 updated to reflect the correct placement. A transcription error would have produced a subtle wrong-constellation bug invisible to tests that did not pin row-7.

2. **`TestComputePositions_MoonphaseWaningClampsNegative` mis-named.** Plan v1 claimed day=18 produces a negative moonphase via the clamp branch. Hand-computation showed day=18 produces pre-clamp moonphase = 4 exactly, which does NOT trigger the `if > 4` clamp — final is +4. Renamed the test to `TestComputePositions_MoonphaseAtUpperBoundary` (pins +4, not clamped) and added a second test `TestComputePositions_MoonphaseWanesPastFull` using day=20 (pre-clamp = 5, post-clamp = -3) to exercise the clamp branch correctly.

3. **Header line terminator inconsistency.** Plan v1 said `renderStarmap` returns lines without `\n\r`, but the C source includes `\n\r` in the header-line string literal. Plan clarified that `renderStarmap` returns lines without trailing `\n\r` (uniform shape) and `LookSky` appends `\n\r` per line on emit. The net observable output matches C exactly; only the contract with the tests changes.

4. **Indoor check convention divergence documented.** C `IS_OUTSIDE` at `src/mud.h:4056` is flag-only: `!xIS_SET(in_room->room_flags, ROOM_INDOORS)`. The Go port's canonical pattern since Phase 5 Tier 3 (ref: `phase5-tier3-completed.md` A1) uses flag-OR-sector. The plan deliberately selects the Go convention as a documented divergence (Q1). This mirrors `DoWeather`, `ifcheck.go:640-643`, and `spell_unique.go:281`.

5. **`weath_unit` scope.** C defines it as an `int` global with a loader and an admin command. Go port intentionally ships it as a file-private `const = 10` (§Q2 / §Scope Cuts). Promotion to `world.World` field deferred until a first cross-package consumer.

6. **`renderStarmap` signature.** Plan v1 had a `TimeInfoData` parameter; self-review noted that individual `int`s are more testable and mirror C's locals. Updated to `renderStarmap(hour, day, month, rawPrecip int) []string` (Q3).

7. **Transcription guard strength.** Plan v1 proposed a "first-10-char prefix" pinning test for starMap rows; self-review judged that too weak (middle-byte transcription errors slip through). Upgraded to full-string equality per row (Q6).

Unresolved items requiring external adversary verification:

- The `precipBucket` hand-computation domain claim (that raw precip is bounded `[-30, 30]` per `src/update.c:3377-3378`). Plan spot-checked the citation but did not hand-verify the bounds.
- The eclipse-at-noon behavior (`sunpos == moonpos && hour == 12`) edge case. Plan pins it via `TestRenderStarmap_EclipseAtNoonRendersMoon` but the exact expected output row-by-row was not hand-computed — tests assert the presence of `&W@` in the output, not a full row's bytes. An external adversary may want to demand stricter pins here.
- The claim that `ColorFunc` passthrough preserves `&X` codes in test buffers. Plan relies on the default nil `ColorFunc` in test fixtures — if a test fixture elsewhere installs a stripping `ColorFunc`, this plan's assertions would fail. Spot-check by grep would confirm.

### Scope for external adversary (if dispatched)

- Verify all C citations (`src/starmap.c:*`, `src/act_info.c:*`, `src/db.c:108,600`, `src/mud.h:4056`) are accurate against HEAD.
- Verify all 8 lines of the `starMap` constellation table in §C Reference match the source bytes exactly.
- Verify `precipBucket` formula matches `src/starmap.c:96-98` including behavior on negative inputs.
- Recompute `TestComputePositions_MoonphaseAtUpperBoundary` and `TestComputePositions_MoonphaseWanesPastFull` by hand.
- Confirm the indoor-check divergence from C is the right convention for this project.
- Confirm task groups are worker-executable (one session per group) and acceptance criteria are mechanically verifiable (no judgment calls).
- Confirm scope cuts don't omit user-visible obligations.

---

## Completion Record

Left empty until the plan executes. Filled after workers complete and adversary passes close out the work unit.
