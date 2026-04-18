# Plan — Tranche C: Quality / Fidelity Passes

**Status:** Planned (2026-04-18). Adversary-verified research: `util.Act` caller inventory (33 sites, 5 files — manager-internal verification; external adversary pass recommended before execution), `update_aris` Go-side semantic check (architectural difference confirmed — Go uses incremental `AffectModify`, C uses rebuild-from-scratch `update_aris`), usability polish cited-line verification.
**Priority:** P2 — cross-cutting quality. None of these unblock downstream work; all are fidelity improvements.
**Scope:**
- `internal/util/act.go` + every caller of `util.Act`
- `internal/persist/player.go` + audit-doc correction for `update_aris`
- `internal/game/loop.go` descriptor-flush path (PLR_BLANK)
- `internal/act/skills.go` `learnFromSuccess` (XP-on-gain + fully-learned message)

---

## Problem statements

### Item 1 — `util.Act` per-call color code (C `AType` parameter)

C's `act()` signature is `act(sh_int AType, const char *format, CHAR_DATA *ch, const void *arg1, const void *arg2, int type)` (C `src/smaug.c:3404`). The first argument is the ANSI color code applied uniformly to every recipient for that single call. Callers in `act_comm.c` / `fight.c` use constants like `AT_ACTION` (room narration), `AT_HIT` (attacker's view of a hit), `AT_HITME` (victim's view of being hit), `AT_IMMORT` / `AT_GTELL` / `AT_GOSSIP` (channel colors), `AT_SAY`, `AT_TELL`, `AT_WHISPER`, `AT_YELL`, `AT_PLAIN`.

**Important scope correction from research.** The original scope framing described this as "per-recipient color preservation" and estimated "hundreds of sites" requiring a per-target closure. Neither is accurate:

1. C's `AType` is per-*call*, not per-recipient. The same `AType` is applied to every target of one `act()` call. Callers that want different colors for attacker vs victim make *two* `act()` calls: one with `AT_HIT TO_CHAR`, one with `AT_HITME TO_VICT` (e.g., `src/fight.c:4586-4588`). Go's callers already follow this two-call pattern — they just pass no color today.
2. The Go caller inventory is **33 call sites across 5 files**: `combat/dammessage.go` (9), `act/cmds2.go` (7), `act/skills3.go` (7), `act/skills4.go` (8), `act/playercfg.go` (2). A big-bang rewrite is tractable in one session. (Line count of 34 in a naive grep includes one comment reference at `combat/dammessage.go:186`; actual call sites = 33.)

Current `util.Act` signature is `Act(format string, ch, vch *CharData, arg1, arg2 any, to int)`. The fidelity gap: every caller's formatted output goes to the descriptor uncolored (unless the format string itself contains `&X` tokens, which a few callers work around with hardcoded `&Y`/`&G`/`&D`).

Consequence: combat and channel output lack color differentiation in the Go port. Players see uniform default-colored text where C shows attack verbs in `AT_HIT` red, incoming hits in `AT_HITME` bright red, immortal chat in `AT_IMMORT` yellow, etc.

**Note on `AT_CLANTALK` et al.** The original scope framing mentioned `AT_CLANTALK` — this constant does NOT exist in C's `at_color_types` enum (`mud.h:1108-1165`). Clan talk in C goes through `talk_channel(... CHANNEL_CLAN ...)` which uses per-channel color constants like `AT_MUSE` / `AT_YELL` / `AT_GOSSIP` depending on branch at `act_comm.c:807-845`. Channel color mapping is covered separately by the existing `plan-channels.md` follow-up queue item; this plan focuses on `act()`-call sites.

### Item 2 — `update_aris` audit correction + regression test

`plan-player-config.md` R5 flags "`update_aris` not called before `save_char_obj` in `DoSave`." Research finding: **this does not apply to the Go port.** The two systems have architecturally different approaches to affect-derived stats:

- **C (`src/handler.c:1221-1305`):** `update_aris` clears `ch->affected_by`, `resistant`, `immune`, `susceptible`, then rebuilds them from scratch by re-applying race + class + deity + per-affect + equipment affects + room affects + morph. It runs on load (`src/save.c:1078, 1238`), after `affect_to_char`/`affect_from_char`, and in a few hotspots (deity favor changes, polymorph, magic). C saves the stored `hitroll`/`damroll`/`armor` fields but `de_equip_char(ch)` at `src/save.c:213` strips equipment BEFORE save, so the saved numbers are base-minus-equipment. On load, objects get re-equipped and affects re-applied.

- **Go (`internal/handler/handler.go:384-455`):** `AffectModify(ch, aff, fAdd)` incrementally adds or subtracts the affect's modifier from the relevant `ch` field (`ch.Hitroll += mod`, `ch.Armor += mod`, etc.) in a single switch. `AffectToChar` calls it with `fAdd=true`; `AffectRemove` calls it with `fAdd=false`. `SavePlayer` writes `ch.Hitroll` / `ch.Damroll` / `ch.Armor` as computed (equipment + affects already applied). `LoadPlayer` reads them back verbatim. Objects are appended to `ch.Carrying` with their `WearLoc` field set; **no `EquipChar` or `AffectModify` is invoked on load.**

The Go design is a save-as-is, load-as-is round-trip: stats are always "current" — no separate base-vs-computed distinction. The C `update_aris` pattern exists because C's `affected_by`/`resistant`/`immune`/`susceptible` fields are rebuilt after every affect change, AND because C saves base stats after de-equipping. Neither of those forcing functions applies in Go.

**What this means for the plan.** R5's framing ("not called before `save_char_obj` in `DoSave`") assumes a C-style architecture that Go does not use. The correct action is:

1. **Audit correction** in TODO.md and `plan-player-config.md` R5: replace with a note that Go uses incremental `AffectModify` and does not need a separate rebuild pass.
2. **Regression-test an edge case** that could silently drift in Go's model: if `AffectRemove` is skipped (e.g., affect expiry bug), the modifier stays applied. Add a focused test that save→expire→load matches expected stats. This is defensive, not fixing a known bug.
3. **Verify no hidden drift** in one specific scenario: an affect that modifies `ch.Armor` via `APPLY_AC` expires mid-session; save; load; confirm `ch.Armor` is correct. Research did not find a bug here, but the test pins the invariant.

The "update_aris needs research before a plan can be written" note from the parent manager is resolved in favor of collapsing the item to an audit correction + pinning test.

### Item 3 — Usability polish (PLR_BLANK + XP-on-gain + fully-learned message)

Three small gaps cited in the audit and queued in `TODO.md` Usability:

#### 3a — `PLR_BLANK` blank-line before prompt

**Scope correction.** The original scope item called this "`PLR_COMPACT` blank-line suppression". `PLR_COMPACT` does not exist in SMAUG at all. The actual flag is `PLR_BLANK` (`enums.go:954`; C `src/mud.h:2506`) and its behavior is the OPPOSITE: when set, the descriptor-flush path EMITS a blank line before the prompt to visually separate output from prompt. The original scope description ("suppression") inverts the semantics.

C implementation at `src/smaug.c:1359-1361`:
```c
if (xIS_SET (ch->act, PLR_BLANK))
  write_to_buffer (d, "\n\r", 2);
```
Runs once per descriptor per pulse in the flush path just before the prompt.

Go has no equivalent check. The equivalent flush path is `internal/game/loop.go:252-259` (`CON_PLAYING` prompt-write) and `internal/game/loop.go:767` (`enterGame` initial prompt).

#### 3b — `XP-on-skill-gain` plumbing

C `src/skills.c:1621-1671` `learn_from_success` awards XP on every skill gain, with two tiers: `gain = 20 * sklvl` for normal gains (with class multipliers: ×6 mage, ×3 cleric), and `gain = 1000 * sklvl` for the "fully learned" boundary (with ×5 / ×2 class multipliers). Go `learnFromSuccess` at `internal/act/skills.go:46-83` raises `ch.PCData.Learned[gsn]` and sends `"You have become better at %s!"`, but line 81 explicitly defers with: `// Note: XP-on-gain and "fully learned" message deferred to Tier 4.`

#### 3c — Fully-learned / adept-cap message

Paired with 3b. C prints `"You are now an adept of %s!  You gain %d bonus experience!\n\r"` with `AT_WHITE` color when `learned == adept` (C `src/skills.c:1649-1652`). Go silently stops improving when the adept cap is hit (loop 66-68 — `if learned >= adept { return }`).

**No level-up logic exists in Go today.** `ch.Exp` is modified in one place (`internal/combat/combat.go:699`, kill XP). There is no `GainExp` helper, no exp-to-next-level table, no level-up branching. This means XP-on-skill-gain in Go is *simpler* than C — just `ch.Exp += gain` + message. Level-up handling is a separate Phase-6 concern.

---

## C reference (authoritative line citations)

### `util.Act` (Item 1)

- **`act()` signature:** `src/smaug.c:3404` — `void act (sh_int AType, const char *format, CHAR_DATA *ch, const void *arg1, const void *arg2, int type)`.
- **Per-recipient color application:** `src/smaug.c:3648` — `set_char_color (AType, to)` inside the recipient loop at `src/smaug.c:3553-3660`. The same `AType` is applied to every target of the current call; there is NO per-recipient AType selection logic.
- **`AT_*` enumeration:** `src/mud.h:1108-1165` (at_types typedef, 54 entries from `AT_PLAIN` to `AT_MUSIC`). `AT_COLORBASE = 1024`. `AT_MAXCOLOR = AT_TOPCOLOR - AT_COLORBASE`.
- **`set_char_color`:** `src/color.c` looks up `at_color_table[AType - AT_COLORBASE]` and emits the mapped ANSI sequence, optionally remembered via `pcdata->colorize[type]`.
- **Canonical paired call-site pattern:** `src/fight.c:4586-4588` — same format, but `AT_HIT TO_CHAR` and `AT_HITME TO_VICT` as two separate `act()` calls. This is the pattern every Go caller will mirror.

### `update_aris` (Item 2)

- **`update_aris` body:** `src/handler.c:1221-1305`. Clears `affected_by/resistant/immune/susceptible`, re-applies race + class + deity + affects + equipment + room + morph in order. Immortals and NPCs bypass (line 1227).
- **Pre-save de-equip:** `src/save.c:213` — `de_equip_char (ch);` inside `save_char_obj`. This strips equipment BEFORE `fwrite_char` at `src/save.c:286`, so saved `hitroll`/`damroll`/`armor` are base-minus-equipment.
- **Post-save re-equip:** `src/save.c:290+` — `re_equip_char(ch)` after `fwrite_char`.
- **Load-time rebuild:** `src/save.c:1078, 1238` — `update_aris(ch)` after `fread_char` so the in-memory affected_by/RIS reflects the loaded affects.

### Usability polish (Item 3)

- **`PLR_BLANK` flush check:** `src/smaug.c:1359-1361` inside `display_prompt` path; runs once per descriptor per pulse.
- **`PLR_BLANK` toggle:** `src/act_info.c:5738` (`do_config` flag cycle). Go already has `enums.go:954` defined but no user-facing toggle. Not in scope for this plan — toggle lands with Phase-6 `do_config` or standalone `DoBlank`.
- **`learn_from_success` body:** `src/skills.c:1621-1671`. Adept check at `:1631`; gain formula at `:1641`; adept-cap branch at `:1642-1653`; normal branch at `:1654-1667`; `gain_exp` call at `:1669`.
- **`gain_exp`:** `src/skills.c:1470+` — C awards XP and handles level-up. Go has no equivalent; `ch.Exp += gain` is sufficient for this plan.

---

## Go design

### Item 1 — `util.Act` migration strategy

**Chosen: big-bang signature change with mechanical per-caller update.**

New signature:
```go
func Act(aType int, format string, ch, vch *types.CharData, arg1, arg2 any, to int)
```

`aType` is the C `sh_int AType` argument. Values match `AT_*` constants ported to `types/constants.go` (new) — mirroring C enumeration.

**Why big-bang:**
- 33 call sites across 5 files is mechanically tractable in one session.
- A compatibility wrapper (`ActLegacy` vs `Act` with color) would create two-color-paths-forever and force every future caller to think about which to use.
- Per-site updates are line-local: each current `util.Act(fmt, ch, vch, a1, a2, to)` becomes `util.Act(types.AT_ACTION, fmt, ch, vch, a1, a2, to)` or whichever AT_ constant matches the C original.
- Each file's tests continue to work if the color codepath is behind a feature flag during the migration; finalize with the color path live.

**Color mapping strategy.** `AT_*` constants map to existing `&-codes` via a lookup table in `util/act.go`. **These are Go-specific design choices, NOT ports of C `default_set`.** External adversary (2026-04-18) verified that C's `default_set[]` at `src/color.c:121-153` maps `AT_ACTION→AT_BLOOD` (dark red), `AT_HIT→AT_GREY`, `AT_HITME→AT_DGREY`, `AT_IMMORT→AT_RED` — none match the values below. The Go values below are chosen to match existing hardcoded `&-codes` in the current Go channel / combat callers (e.g., `DoImmtalk` already uses `&Y`, poisoned-weapon messages already use `&G`), which prioritizes Go-internal consistency over byte-for-byte C fidelity. Per-player color customization is deferred (see Scope Cuts) — a future plan can revisit the mapping if we ship a `Colorize[]` port.

```go
var atColorCode = map[int]string{
    types.AT_PLAIN:   "",       // no color — existing default
    types.AT_ACTION:  "&G",     // green (Go convention: narration/action; C default is AT_BLOOD)
    types.AT_HIT:     "&R",     // bright red (Go convention: attacker's hit; C default is AT_GREY)
    types.AT_HITME:   "&r",     // dark red (Go convention: incoming hit on victim; C default is AT_DGREY)
    types.AT_IMMORT:  "&Y",     // yellow (matches existing DoImmtalk color; C default is AT_RED)
    types.AT_GTELL:   "&P",     // magenta (group tell)
    types.AT_GOSSIP:  "&P",     // magenta (gossip)
    types.AT_SAY:     "&G",     // green (say)
    types.AT_TELL:    "&G",     // green (tell)
    types.AT_WHISPER: "&G",     // green (whisper)
    types.AT_YELL:    "&Y",     // yellow (yell)
    types.AT_SHOUT:   "&y",     // dark yellow (shout)
    types.AT_MAGIC:   "&C",     // cyan (magic)
    types.AT_POISON:  "&g",     // dark green (poison)
    // ... ~5 more codes actually used by current Go callers
}
```

Research finding: `AT_CLANTALK` was in the initial design table draft and has been removed — it does not exist in the C enum. Go clan talk in `act/clan.go` uses inline `&-codes` directly; retain that pattern unless/until a future plan harmonizes channel coloring.

**Default color values sourced from:**
- Existing Go hardcoded `&-codes` in callers — e.g., `DoImmtalk` uses `&Y`, poisoned-weapon prefix uses `&G`. This is the Go project's de-facto convention.
- NOT from C `src/color.c` `default_set[]` — adversary-verified the C defaults are materially different (see note above the table). Deliberate divergence accepted to preserve current Go visual identity; a fidelity-first revisit is a Phase-6 concern.

**Emission:** `sendActTo` prepends `atColorCode[aType]` before `msg` and appends `&D` (reset) after the trailing `\n\r` if color is present. Net-layer `ProcessColors` already handles `&X` expansion.

**Rejected alternatives:**
- *Per-target color closure.* `func(target *CharData) string` that returns the color per recipient. Unnecessary complexity — C doesn't do this either (same AType for all recipients of one call).
- *Compatibility wrapper (`ActColor` next to `Act`).* Creates drift pressure: new callers would default to whichever signature they found first. One signature that always takes a color keeps discipline.
- *Store color in the format string.* Would require every caller to do `"&G$n hits $N"` instead of `"$n hits $N"`. More error-prone than passing the AT_ constant and keeps the convention shared with C.

**Migration recipe (per file):**
1. Read current file to enumerate caller sites + their TO_ arg.
2. For each site: pick the C-matching `AT_*` constant from the table below.
3. Apply `util.Act(types.AT_X, oldArgs...)`.
4. Run tests for that package — they should still pass (message text unchanged, only a color prefix added; tests that match against exact output will need the prefix or be updated to strip).

**Mapping table (by file):**

| File | Site count | Dominant AT_ | Notes |
|---|---|---|---|
| `combat/dammessage.go:274-302` (skill miss/hit) | 6 | `AT_HIT` for TO_CHAR, `AT_HITME` for TO_VICT, `AT_ACTION` for TO_NOTVICT | Matches C `fight.c:4524-4554` |
| `combat/dammessage.go:351-356` (generic buf1/buf2/buf3) | 3 | `AT_HIT` TO_CHAR, `AT_HITME` TO_VICT, `AT_ACTION` TO_NOTVICT | Matches C `fight.c:4586-4588` |
| `act/cmds2.go:130` (quest hint) | 1 | `AT_PLAIN` | TO_VICT gch — could also be `AT_QUEST` but quest-color mapping is Phase-6 |
| `act/cmds2.go:179-180` (light torch) | 2 | `AT_ACTION` | Self + room |
| `act/cmds2.go:221-235` (throw) | 4 | `AT_ACTION` | Self + room |
| `act/skills3.go:149-151` (stun success) | 3 | `AT_ACTION` / `AT_HITME` | Victim-side uses `AT_HITME` |
| `act/skills3.go:165-166` (stun miss) | 2 | `AT_ACTION` / `AT_HITME` | |
| `act/skills3.go:217-218` (grapple) | 2 | `AT_ACTION` | |
| `act/skills4.go:29, 48, 125, 227, 231` (meditate/trance/dig/misc rooms) | 5 | `AT_ACTION` | Room-only narration |
| `act/skills4.go:263-265` (feed) | 3 | `AT_ACTION` | Self + vict + room |
| `act/playercfg.go:47, 52` (DoAfk room broadcast on + off) | 2 | `AT_ACTION` | |

Total: **33 site updates**, all mechanical. No site requires runtime color logic; every site has a clear C-mapped AT_ constant.

### Item 2 — `update_aris` audit correction

**No code change.** This item is a documentation/test-only landing:

1. **Correct `plan-player-config.md` R5** — append a note that the Go port does not need `update_aris` equivalence because `AffectModify` maintains stats incrementally and `SavePlayer` round-trips computed values verbatim. Tie to the concrete design difference: C de-equips before save and rebuilds on load; Go saves-as-is.
2. **Update TODO.md** — the follow-up line at `TODO.md:73` (R5) should be re-phrased from "not called before `save_char_obj` in `DoSave`" to "audited 2026-04-18 — does not apply to Go's incremental `AffectModify` architecture; regression test pins invariant (see `persist/player_test.go:TestSaveLoadPlayer_AffectExpiryStatsRoundTrip`)."
3. **Add regression test** in `persist/player_test.go`: construct a char with `ch.Armor = -10`, apply an affect via `AffectToChar` with `APPLY_AC modifier=-20`, save, expire by setting `aff.Duration = 0`, call `AffectRemove`, save again, load, assert `ch.Armor == -10` (not `-30` — the affect was removed, stats should reflect base + no affects).

This test protects against a class of future refactors that might introduce drift between `ch.Hitroll`/`ch.Damroll`/`ch.Armor` and the affect list. It's a pinning test, not a bug fix.

### Item 3 — Usability polish

#### 3a — `PLR_BLANK` emit before prompt

Insert a check in `internal/game/loop.go:252-259` (the `CON_PLAYING` prompt-write path) and `internal/game/loop.go:767` (`enterGame` initial prompt):

```go
if d.Character != nil && d.Character.Act.IsSet(types.PLR_BLANK) {
    d.WriteToBuffer("\n\r")
}
d.WriteToBuffer(FormatPrompt(d.Character))
```

Both locations must be updated — the initial prompt path in `enterGame` runs ONCE per login; the pulse-loop path runs every tick that has output. Factor into a helper `writePromptWithBlank(d *DescriptorData)` to keep the two call sites in sync.

**No toggle command shipped.** `PLR_BLANK` can only be cleared today by direct flag manipulation (there's no Go `do_config +blank` command). The flag WILL round-trip through `SavePlayer`/`LoadPlayer` because it's in `ch.Act` which already persists. An orthogonal `DoBlank` or Phase-6 `DoConfig` will expose the toggle to players. This plan ships the rendering change only; the toggle is a separate follow-up.

#### 3b — XP-on-skill-gain

In `learnFromSuccess` at `internal/act/skills.go:78-82`, after `ch.PCData.Learned[gsn] = util.UMIN(learned+gain, adept)`:

```go
// XP award matching C src/skills.c:1655-1669 (normal-gain branch).
skLvl := skill.SkillLevel[ch.Class]
if skLvl == 0 {
    skLvl = ch.Level
}
xpGain := 20 * skLvl
switch ch.Class {
case types.CLASS_MAGE:
    xpGain *= 6
case types.CLASS_CLERIC:
    xpGain *= 3
}
// Silent during combat or for sneak/hide (matches C :1661).
fighting := ch.Fighting != nil
silentSkill := gsn == gsnHide || gsn == gsnSneak // both are -1 if unresolved; matches C when the GSN lookup would have failed
if !fighting && !silentSkill {
    ch.Sendf("You gain %d experience points from your success!\n\r", xpGain)
}
ch.Exp += xpGain
```

`CLASS_MAGE` and `CLASS_CLERIC` already exist in `types/enums.go`. `gsnHide` / `gsnSneak` are added as package-private vars in `act/skills.go` and resolved in a new `act.ResolveGSNs()` wired from `boot.Boot` — see G7 Prereq for details.

#### 3c — Fully-learned adept-cap message

In the same `learnFromSuccess`, when `ch.PCData.Learned[gsn] == adept` after the UMIN:

```go
if ch.PCData.Learned[gsn] == adept {
    // Fully-learned branch matches C src/skills.c:1644-1652.
    xpGain := 1000 * skLvl
    switch ch.Class {
    case types.CLASS_MAGE:
        xpGain *= 5
    case types.CLASS_CLERIC:
        xpGain *= 2
    }
    ch.Sendf("&WYou are now an adept of %s!  You gain %d bonus experience!\n\r&D",
        skill.Name, xpGain)
    ch.Exp += xpGain
    // Do NOT also award the normal-gain XP — C's if/else (skills.c:1642/1654) makes these mutually exclusive.
}
```

Two branches (normal-gain vs adept-cap) must be mutually exclusive. The refactored `learnFromSuccess` body:

```go
if gain > 0 {
    oldLearned := learned
    ch.PCData.Learned[gsn] = util.UMIN(oldLearned+gain, adept)
    ch.Sendf("You have become better at %s! (%d%%)\n\r", skill.Name, ch.PCData.Learned[gsn])

    skLvl := skill.SkillLevel[ch.Class]
    if skLvl == 0 {
        skLvl = ch.Level
    }
    var xpGain int
    if ch.PCData.Learned[gsn] == adept {
        xpGain = 1000 * skLvl
        switch ch.Class {
        case types.CLASS_MAGE:
            xpGain *= 5
        case types.CLASS_CLERIC:
            xpGain *= 2
        }
        ch.Sendf("&WYou are now an adept of %s!  You gain %d bonus experience!\n\r&D",
            skill.Name, xpGain)
    } else {
        xpGain = 20 * skLvl
        switch ch.Class {
        case types.CLASS_MAGE:
            xpGain *= 6
        case types.CLASS_CLERIC:
            xpGain *= 3
        }
        fighting := ch.Fighting != nil
        silentSkill := gsn == gsnHide || gsn == gsnSneak
        if !fighting && !silentSkill {
            ch.Sendf("You gain %d experience points from your success!\n\r", xpGain)
        }
    }
    ch.Exp += xpGain
}
```

---

## Task groups

### G1 — Port `AT_*` color constants to `types/`

**File:** `internal/types/constants.go` (new section near existing color-like constants).
**Test:** `internal/types/constants_test.go` — pin the `AT_PLAIN` / `AT_ACTION` / `AT_HIT` / `AT_HITME` / `AT_IMMORT` / `AT_GTELL` values as `AT_COLORBASE+1`, `AT_COLORBASE+2`, etc., matching C enum order at `src/mud.h:1108-1165`.
**Mutation verify:** changing any constant value → test red → revert via Edit → green.
**Scope:** ~54 constants defined; only the ~15 actually used by Go callers need values the test pins. The rest are declared for completeness.
**Effort:** S (~30min).

### G2 — Add color parameter to `util.Act`, helper table, default behavior

**File:** `internal/util/act.go`.
**Changes:**
- `Act(aType int, format string, ch, vch *types.CharData, arg1, arg2 any, to int)` — add `aType` as first param.
- `sendActTo` prepends `atColorCode[aType]` (if non-empty) and appends `&D` after the trailing `\n\r` when color is present.
- `atColorCode map[int]string` at package level with ~20 entries (the ones actually used by current Go callers).
- `AT_PLAIN` (value 0 in C `src/mud.h:1109` — `AT_COLORBASE = 1024, AT_PLAIN = AT_COLORBASE`) produces no color prefix (empty string in the table). This preserves current behavior for any caller that passes `AT_PLAIN`.
**Test first:** `TestAct_ColorPrefix_AT_HIT` — message sent with `AT_HIT` to a victim gets `&R` prefix + `&D` suffix; `AT_PLAIN` produces no prefix.
**Test:** `TestAct_UnknownAType_NoPrefix_NoPanic` — passing an undefined AT value emits the message without color (default-to-empty lookup).
**Mutation:** remove `atColorCode[aType]` lookup → color tests fail; remove `&D` suffix → suffix test fails; revert via Edit → green.
**Effort:** M (~1.5h).

### G3 — Update all 33 `util.Act` call sites

Parallel sub-tasks, one per file. Each is mechanical: for every `util.Act(fmt, ...)` call, insert the correct `AT_*` constant as first argument per the mapping table above.

**G3a — `internal/combat/dammessage.go`** (9 sites, all paired-pattern): `AT_HIT` for TO_CHAR, `AT_HITME` for TO_VICT, `AT_ACTION` for TO_NOTVICT. Existing tests in `dammessage_test.go` should still pass — but any test that asserts exact output may need color-prefix tolerance. **Test-first step:** add one `TestDamMessage_ColorForTO_CHAR` that asserts the `&R` prefix appears for a hit emit; run red; do the migration; run green.

**G3b — `internal/act/cmds2.go`** (7 sites): mostly `AT_ACTION`. One site at `:130` is a quest hint `TO_VICT` — C-equivalent is `AT_ACTION` or `AT_QUEST`; choose `AT_ACTION` since Go's quest system is Phase-4b ported without `AT_QUEST`.

**G3c — `internal/act/skills3.go`** (8 sites): `AT_ACTION` for room broadcasts; `AT_HITME` for "stun you"/"smashes into you" victim-side messages.

**G3d — `internal/act/skills4.go`** (9 sites): all `AT_ACTION` (meditate/trance/dig/morph-mist/feed — all narrative).

**G3e — `internal/act/playercfg.go`** (2 sites, `DoAfk` room broadcast at `:47` and `:52` — off-transition and on-transition): both `AT_ACTION`.

**Per-file test mutation:** before migration, add a tripwire test in one file asserting the new color prefix is present; run red; do the migration in that file; run green; revert the tripwire (it's scaffolding — real tests land in G2). Then move to the next file.

**Effort:** M (~2h total across all 5 files).

### G4 — Migration verification pass

Run `go build ./...`, `go vet ./...`, `go test -count=3 ./...`. Every package green. Any call site missed will fail to compile (signature change is a hard break — this is the feature, not the bug). Search for any dormant call: `grep -rn 'util\.Act(' internal/ | grep -v '_test.go' | grep -v '// '` — the count MUST match the pre-migration count (33 real call sites), and every call MUST have 7 arguments (was 6).

**Test:** one cross-package `TestAct_FullMigration_AllCallSitesMigrated` in `util/act_test.go` is not feasible (compile-time check already covers it). Instead document the migration's build-break semantics in a comment at the top of `act.go` so future refactorers don't reintroduce the old 6-arg form.

**Effort:** S (~30min).

### G5 — `update_aris` audit correction + regression test

**Files touched (doc):** `smaug-go/doc/plan-player-config.md` R5, `TODO.md:73`.
**File touched (test):** `internal/persist/player_test.go` — new test `TestSaveLoadPlayer_AffectRemovalMaintainsStatInvariant`:

```go
func TestSaveLoadPlayer_AffectRemovalMaintainsStatInvariant(t *testing.T) {
    ch := makeTestPlayer() // Hit=50, Armor=-10, all base stats set
    aff := &types.AffectData{
        Type:     1,
        Duration: 10,
        Location: types.APPLY_AC,
        Modifier: -20, // C semantic: negative = better AC
    }
    handler.AffectToChar(ch, aff)
    if ch.Armor != -30 {
        t.Fatalf("AffectToChar should have made Armor=-30, got %d", ch.Armor)
    }
    // Expire and remove
    aff.Duration = 0
    handler.AffectRemove(ch, aff)
    if ch.Armor != -10 {
        t.Fatalf("AffectRemove should have restored Armor=-10, got %d", ch.Armor)
    }
    // Save + load round-trip
    var buf bytes.Buffer
    persist.SavePlayer(&buf, ch)
    loaded, err := persist.LoadPlayer(&buf, "test")
    if err != nil {
        t.Fatal(err)
    }
    if loaded.Armor != -10 {
        t.Errorf("round-trip Armor = %d, want -10 (affect was removed before save)", loaded.Armor)
    }
    if len(loaded.Affects) != 0 {
        t.Errorf("expected zero affects after removal + round-trip, got %d", len(loaded.Affects))
    }
}
```

**Mutation verify:** break `AffectRemove`'s `AffectModify(ch, aff, false)` call (change to `fAdd=true`) → test fails (`Armor = -50, want -10`). Edit back → green.

**Plan doc correction:** append to `plan-player-config.md` R5 a paragraph that explains Go's incremental model, cites `handler.AffectModify`, and flags this test as the regression guard.

**TODO.md correction (explicit G5 step):** Edit `TODO.md` line ~73 where R5 currently reads `- [ ] R5: update_aris not called before save_char_obj in DoSave (low-impact for manual save — follow-up).` Change to `- [x] R5: audit-doc correction landed 2026-04-18 via plan-tranche-c.md G5. Go uses incremental AffectModify; no update_aris port needed. Regression guard at TestSaveLoadPlayer_AffectRemovalMaintainsStatInvariant.` The worker executing G5 performs this edit directly — previous revisions of this plan deferred it to the parent manager, which created an ambiguous deliverable for A7; now explicit.

**Effort:** S (~45min).

### G6 — `PLR_BLANK` blank-line emission

**File:** `internal/game/loop.go`.
**Changes:**
- New unexported helper `writePromptWithBlank(d *types.DescriptorData)`:
  ```go
  func writePromptWithBlank(d *types.DescriptorData) {
      if d == nil || d.Character == nil {
          return
      }
      if d.Character.Act.IsSet(types.PLR_BLANK) {
          d.WriteToBuffer("\n\r")
      }
      d.WriteToBuffer(FormatPrompt(d.Character))
  }
  ```
- Replace the two direct `d.WriteToBuffer(FormatPrompt(d.Character))` sites at `:258` and `:767` with `writePromptWithBlank(d)`.

**Test first:** `TestFlushPath_PLR_BLANK_EmitsBlankLine` — create a test descriptor + character with `PLR_BLANK` set, trigger the prompt write, assert output starts with `"\n\r"` followed by the prompt. Second test: same char without `PLR_BLANK` → output starts with prompt, no leading `\n\r`. Third test: nil descriptor / nil character → no panic.

**Mutation:** remove the `IsSet(PLR_BLANK)` check → every prompt emits an extra blank line; test fails. Invert to `!IsSet` → blank-line-absent test fails. Revert via Edit → green.

**Effort:** S (~45min).

### G7 — XP-on-skill-gain + fully-learned message

**File:** `internal/act/skills.go:46-83` (`learnFromSuccess`).
**Changes:** replace the current body with the two-branch version from the "Go design" section.

**Prereq: GSN resolution for hide/sneak.** Research confirms that **`GsnHide` / `GsnSneak` do NOT exist in the Go port today** — `internal/combat/skillcheck.go:70-90` resolves combat-related GSNs (second_attack, third_attack, backstab, circle, pounce, dual_wield, berserk, weapon-prof skills) but not hide/sneak. For this G7 landing, add them via one of:

1. **Option A (recommended): add `gsnHide` / `gsnSneak` to `act/skills.go`** — declare them as package-private `var gsnHide = -1` / `var gsnSneak = -1`, resolve in a new `ResolveGSNs()` func in `act/` wired by `boot.Boot` after `persist.SkillNameLookup` is set. Mirrors the `combat/skillcheck.go:ResolveGSNs` pattern.

2. **Option B (fallback): resolve by name at call site** — `if skill.Name == "hide" || skill.Name == "sneak" { silentSkill = true }`. Slower on the hot path but zero boot-time setup. Acceptable because `learnFromSuccess` is called at most a few times per second per player.

**Decision:** ship Option A. One-time resolution is the idiomatic pattern already used in the project (`combat/skillcheck.go`); the boot-time cost is negligible; the call site stays an integer compare. Add `act.ResolveGSNs()` that looks up `"hide"` and `"sneak"` via `persist.SkillNameLookup`; wire from `boot.Boot` immediately after `combat.ResolveGSNs()` at `boot.go:~130`.

**Boot test:** add `if gsnHide < 0 { t.Error(...) }` / `gsnSneak` assertions in `boot_test.go:TestBoot_WiresCallbacks`. Mirrors existing `combat.gsnBackstab` post-boot assertions.

**Test first:**
- `TestLearnFromSuccess_AwardsXP_NormalBranch` — set Learned < adept, stub `numberPercent` to hit the `gain=2` branch, confirm `ch.Exp` increased by `20 * skLvl`, message "You gain X experience points from your success!" sent.
- `TestLearnFromSuccess_ClassMultiplier_Mage` — same scenario with `ch.Class = CLASS_MAGE`, expect `120 * skLvl` (=20×6).
- `TestLearnFromSuccess_ClassMultiplier_Cleric` — expect `60 * skLvl` (=20×3).
- `TestLearnFromSuccess_FightingSuppressesMessage` — set `ch.Fighting = &some_mob`, assert XP added but message NOT sent.
- `TestLearnFromSuccess_HideSneakSuppressMessage` — same, with `gsn == GsnHide` or `GsnSneak`.
- `TestLearnFromSuccess_AdeptCap_Message_And_LargerXP` — Learned = adept-1 + gain=2 clamps to adept; expect `&WYou are now an adept of %s!` message + `1000 * skLvl` XP (× 5 mage, × 2 cleric).
- `TestLearnFromSuccess_AdeptCap_MutuallyExclusiveWithNormalBranch` — pin that "You gain X experience points" is NOT emitted when the adept-cap branch fires.

**Mutation verify:**
- Swap `20` and `1000` in the two branches → tests for both cap and normal XP values fail.
- Remove `ch.Exp += xpGain` → XP tests fail.
- Remove `!silentSkill` condition → hide/sneak XP test fails.
- Break the adept-cap conditional (`== adept` → `> adept` unreachable) → fully-learned test fails.

All reverts via Edit, never destructive git.

**Effort:** M (~1.5h).

### G8 — Plan wrap-up (doc + completion record)

Append completion record to this plan file. Update `TODO.md` deferral lines for the landed items (Tranche A owns writes — flag them). No other doc changes needed because `plan-player-config.md` R5 correction is in G5.

**Effort:** XS (~15min).

---

## Acceptance criteria

A1. `types.AT_PLAIN`, `AT_ACTION`, `AT_HIT`, `AT_HITME`, `AT_IMMORT`, `AT_GTELL`, `AT_GOSSIP`, `AT_SAY`, `AT_TELL`, `AT_YELL`, `AT_SHOUT`, `AT_MAGIC`, `AT_POISON`, `AT_WHISPER` all exist and have values matching C `src/mud.h:1108-1165` enum order (i.e., `AT_PLAIN = AT_COLORBASE = 1024`; `AT_ACTION = 1025`; `AT_SAY = 1026`; etc.). `AT_CLANTALK` does NOT exist in C and must NOT be added to Go.

A2. `util.Act` signature is `Act(aType int, format string, ch, vch *types.CharData, arg1, arg2 any, to int)`. All 33 caller sites migrated. `grep -rn 'util\.Act(' /home/eilidh/src/smaug/smaug-go/internal/ --include='*.go' | grep -v '_test.go' | grep -v '// '` returns exactly 33 matches, every one with 7 arguments.

A3. `go build ./...` clean. `go vet ./...` clean. `go test -count=3 ./...` green across all 15 packages.

A4. A player with `PLR_BLANK` set receives a blank line before every prompt emission; a player without it does not.

A5. `learnFromSuccess` awards `20 * skLvl` XP on normal gain (×6 mage, ×3 cleric), `1000 * skLvl` XP on adept-cap (×5 mage, ×2 cleric). The adept-cap message `"You are now an adept of %s!  You gain %d bonus experience!"` fires exactly once when `Learned` hits `adept`. `Fighting || gsn == GsnHide || gsn == GsnSneak` suppresses the message but not the XP.

A6. Save/load round-trip preserves `ch.Armor` / `ch.Hitroll` / `ch.Damroll` verbatim when all affects have been applied AND removed before save (`TestSaveLoadPlayer_AffectRemovalMaintainsStatInvariant` green).

A7. `plan-player-config.md` R5 section is amended with a paragraph explaining Go's incremental `AffectModify` architecture; the TODO.md R5 line is superseded by the regression test.

A8. `TestAct_ColorPrefix_AT_HIT` and `TestAct_UnknownAType_NoPrefix_NoPanic` exist and are green.

A9. `TestFlushPath_PLR_BLANK_EmitsBlankLine` + its negative sibling + its nil-safety sibling exist and are green.

A10. Every worker mutation in G3 / G5 / G6 / G7 is reverted via `Edit` tool round-trips, never destructive `git checkout` / `git restore` / `git reset --hard` / `git stash`.

---

## Scope cuts / deferrals

- **`DoConfig` / `DoBlank` toggle commands.** `PLR_BLANK` rendering ships in this plan; the user-facing toggle to set/clear the flag is a separate follow-up (either standalone `DoBlank` like `DoGag` / `DoAfk`, or Phase-6 full `DoConfig`). The flag persists through `SavePlayer`/`LoadPlayer` via `ch.Act` today.
- **Full `AT_*` `at_color_table`.** Only ~15-20 AT codes are actually used by Go callers today. The remaining ~35 constants are declared (G1) but map to no-op in `atColorCode` (empty string → no color prefix). Future content or Phase-6 commands using new AT codes will need to extend the table.
- **C `ENABLE_COLOR` / `#ifndef ENABLE_COLOR` variant branching.** C has a build-time toggle where `AT_COLORIZE` works differently. Go does not have this branching — color is always on but can be stripped via `PLR_ANSI` at the net layer. Not porting.
- **`MOBtrigger` / `mprog_act_trigger` hook** (C `src/smaug.c:3655-3659`). MPROG_ACT triggers fire against the formatted string per recipient. Not in current Go port scope — covered independently in mudprog Tier 3.
- **Level-up side effects on XP gain.** No `GainLevel` / level-up branch exists in Go (research: no `ch.Level +=` anywhere except Level-loading in `persist/player.go`). XP-on-skill-gain simply increments `ch.Exp`. Level-up is a Phase-6 concern.
- **`update_aris` full port.** Architecturally unnecessary for Go (see Item 2 Problem statement). If a future feature requires C-style rebuild — e.g., a "reset stats" immortal command — port then.
- **`update_aris` before save in `DoSave`.** TODO.md R5 follow-up; audit-corrected in this plan. Not porting.
- **`AT_*` ANSI escape mapping refinements.** The `atColorCode` table uses `&X`-codes that map through `net/color.go`'s existing `colorMap`. If a future C color doesn't have a direct `&X` equivalent, extend `colorMap` and wire into `atColorCode`. Not a blocker for this plan.
- **Per-call-site `AT_` choice adversary review.** Each of the 33 sites' chosen `AT_*` constant maps to a C-equivalent reference. If any choice is incorrect (e.g., `AT_QUEST` preferable over `AT_ACTION` for a quest hint), it's a tuning pass after ship. The mapping table documents the chosen value for review.

---

## Open questions (with recommended answers)

**Q1:** Should `PLR_BLANK` emission go in the prompt path (every descriptor write that emits a prompt) or only in the pulse-loop flush (once per tick per descriptor)?

**Recommended answer:** Every prompt write — both `:258` (pulse loop) and `:767` (`enterGame` initial prompt). C only emits before the prompt, and C has one prompt-emit site. Go has two (pulse + initial); consolidate via helper to keep them in sync. **Decision:** ship with both sites converted.

**Q2:** Should `AT_PLAIN` (C value = `AT_COLORBASE`, the base uncolored state) map to empty-string in `atColorCode`, or to an explicit `&D` reset?

**Recommended answer:** Empty string. Existing callers that produce uncolored text should continue to produce uncolored text after migration; adding `&D` reset would be a behavior change for those sites. The `&D` reset happens only when the message HAD a non-empty prefix (end-of-message guard in `sendActTo`). **Decision:** empty string = no prefix, no suffix.

**Q3:** Should the XP-on-skill-gain messages go through `util.Act` (new color path) or remain `ch.Sendf`?

**Recommended answer:** `ch.Sendf` (current pattern). These are direct-to-self messages, not multi-target narration. `util.Act` is for $n/$N/$p multi-recipient broadcasts. The adept-cap message uses inline `&W` ... `&D` color tokens which `net/color.go` expands directly — no need to go through the AT_ system.

**Q4:** Does the adept-cap message produce a trailing `&D` to reset color back to default?

**Recommended answer:** Yes — `&W...&D` pattern used everywhere in the codebase. Otherwise the next prompt/output line inherits white. **Decision:** include `&D` suffix in the format string.

**Q5:** For `AT_HIT` / `AT_HITME` specifically, research shows C uses `set_char_color(AT_HIT, ...)` which in `at_color_table` has configurable per-player colorize values — meaning individual players can have their own `AT_HIT` color. Do we honor `ch.PCData.Colorize[AT_HIT]` in the Go port?

**Recommended answer:** No — not in this plan. Go's `PCData` has no `Colorize` array today (grep confirms). Adding one is a separate Phase-6 config item. The Go AT table gives every player the same color per AT code, matching C's *default* behavior before per-player customization. If a future user requests customization, extend the Go port. **Decision:** scope cut, document in Deferrals.

**Q6:** Should we add a "fully learned" adept-cap message for `learn_from_failure` too?

**Recommended answer:** No — matches C exactly. C's `learn_from_failure` at `src/skills.c:1685-1686` clamps to `adept - 1`, never reaching `adept`. Only `learn_from_success` can hit the adept cap. So the message fires only from the success path. Go mirrors this (line 117 `util.UMIN(learned+1, adept-1)`).

---

## Risk

**Item 1 (`util.Act` color):** **Medium-low.** Signature change touches 33 sites across 5 files but every change is mechanical and the Go compiler forces-correctness (compile break on missed site). Tests for each package should still pass because message *text* is unchanged — only a color prefix is added. Any test that does exact-string matching on output may need a prefix-tolerant matcher; research identified `dammessage_test.go`, `skills3_test.go`, `skills4_test.go`, `cmds2_test.go`, `playercfg_test.go` as candidates to audit. A sibling test in each that asserts the new prefix is present closes the mutation window.

**Item 2 (`update_aris` audit):** **Low.** Doc + regression test. No runtime behavior change. Test itself could be wrong — mutation verification ensures it actually catches the bug it purports to catch.

**Item 3a (`PLR_BLANK`):** **Low.** Single conditional insert in two sites. Flag already persists. No feature regression.

**Item 3b/c (XP-on-gain):** **Low-medium.** Touches a hot code path (`learnFromSuccess` called on every skill use in combat + magic). Player-visible output changes — any existing test that asserts "no XP message" on a successful skill use will regress. Research: grep confirms no test today asserts XP-silence on skill gain — all current `learnFromSuccess` tests check Learned[] updates, not messages or XP. Safe to land.

**Cross-cutting:** All three items are independent. None depends on any other's completion. If G1-G4 ship and G5-G7 are deferred, the repo is still correct; same for any subset.

---

## Relevant file paths

**C references:**
- `/home/eilidh/src/smaug/src/smaug.c:3171` (`act_string`), `:3404` (`act`), `:3553-3660` (recipient loop), `:3648` (`set_char_color` per recipient), `:1355-1374` (prompt flush with `PLR_BLANK`)
- `/home/eilidh/src/smaug/src/mud.h:1089-1167` (`AT_*` constants), `:2506` (`PLR_BLANK`)
- `/home/eilidh/src/smaug/src/color.c` (`at_color_table`, `set_char_color`)
- `/home/eilidh/src/smaug/src/handler.c:1221-1305` (`update_aris`), `:1889` (update_aris call site)
- `/home/eilidh/src/smaug/src/save.c:193-300` (`save_char_obj` with `de_equip_char`), `:1078`, `:1238` (update_aris on load)
- `/home/eilidh/src/smaug/src/skills.c:1621-1671` (`learn_from_success`), `:1675-1687` (`learn_from_failure`)
- `/home/eilidh/src/smaug/src/fight.c:4524-4554` (skill hit/miss paired act calls), `:4586-4588` (generic dam-message paired act calls)
- `/home/eilidh/src/smaug/src/act_comm.c` (call sites using `AT_IMMORT`, `AT_GTELL`, `AT_GOSSIP`, `AT_SAY`, `AT_TELL`)

**Go files to modify:**
- `/home/eilidh/src/smaug/smaug-go/internal/types/constants.go` (G1)
- `/home/eilidh/src/smaug/smaug-go/internal/types/constants_test.go` (G1, new)
- `/home/eilidh/src/smaug/smaug-go/internal/util/act.go` (G2)
- `/home/eilidh/src/smaug/smaug-go/internal/util/act_test.go` (G2)
- `/home/eilidh/src/smaug/smaug-go/internal/combat/dammessage.go` (G3a)
- `/home/eilidh/src/smaug/smaug-go/internal/act/cmds2.go` (G3b)
- `/home/eilidh/src/smaug/smaug-go/internal/act/skills3.go` (G3c)
- `/home/eilidh/src/smaug/smaug-go/internal/act/skills4.go` (G3d)
- `/home/eilidh/src/smaug/smaug-go/internal/act/playercfg.go` (G3e)
- `/home/eilidh/src/smaug/smaug-go/internal/persist/player_test.go` (G5, new test)
- `/home/eilidh/src/smaug/smaug-go/internal/game/loop.go` (G6)
- `/home/eilidh/src/smaug/smaug-go/internal/game/loop_test.go` (G6)
- `/home/eilidh/src/smaug/smaug-go/internal/act/skills.go` (G7)
- `/home/eilidh/src/smaug/smaug-go/internal/act/skills_test.go` (G7, new tests)

**Go files read for context (not modified):**
- `/home/eilidh/src/smaug/smaug-go/internal/handler/handler.go:384-455` (`AffectModify`, `AffectToChar`, `AffectRemove`)
- `/home/eilidh/src/smaug/smaug-go/internal/persist/player.go` (save/load stat fields)
- `/home/eilidh/src/smaug/smaug-go/internal/types/enums.go:954` (`PLR_BLANK`), `:910-943` (`CHANNEL_*`)
- `/home/eilidh/src/smaug/smaug-go/internal/combat/combat.go:697-700` (existing `ch.Exp +=` kill-XP pattern)
- `/home/eilidh/src/smaug/smaug-go/internal/net/color.go:12-36` (existing `colorMap` for `&X` → ANSI)

**Doc files to update:**
- `/home/eilidh/src/smaug/smaug-go/doc/plan-player-config.md` R5 section (G5)
- `/home/eilidh/src/smaug/smaug-go/doc/plan-tranche-c.md` (this file — completion record appended after landing)

**Doc files touched by Tranche A or parent manager (NOT this plan):**
- `/home/eilidh/src/smaug/TODO.md` (R5 line update, landing markers)
- `/home/eilidh/src/smaug/CHANGELOG.md` (landing entries)
- `/home/eilidh/src/smaug/CLAUDE.md` (doc index if needed)
- `/home/eilidh/src/smaug/smaug-go/doc/phases.md` (Tier index entry)

---

## Mutation verification safety (MANDATORY for worker and adversary prompts)

> Mutation verification must never use destructive git commands. Banned for mutation revert:
> - `git checkout -- <file>`
> - `git checkout <ref> -- <file>`
> - `git restore <file>`
> - `git reset --hard` (any form)
> - `git stash` (any form)
>
> Safe pattern: apply the mutation with the `Edit` tool, run the test to confirm failure, then call `Edit` again with the opposite change to revert. Never touch the working tree with git commands.

This ban applies to every worker subagent and every adversary subagent dispatched against this plan. Include it verbatim in their prompts.

---

## Adversary-resolved concerns

This plan was reviewed via a manager-internal adversary pass (no parallel `adversary` subagent dispatch was available in this execution context — the sub-manager's tool surface is Read/Write/Edit/Bash/Grep/Glob only, no Agent tool). The manager performed three independent verification passes against the original plan draft and resolved findings below. Any plan execution should nonetheless dispatch external adversaries against G1-G7 before merging.

1. **`AT_CLANTALK` does not exist in C.** The original plan design table listed `AT_CLANTALK` as a target color constant. Verified via `grep AT_CLANTALK /home/eilidh/src/smaug/src/*.c *.h` — zero matches. C clan chat routes through `talk_channel(ch, arg, CHANNEL_CLAN, "clantalk")` which uses other AT_ constants depending on channel. Resolution: removed `AT_CLANTALK` from the design color table and added an explanatory note. Go's clan chat at `act/clan.go` continues to use inline `&-codes` directly.

2. **Call-site count was off by one (34 → 33).** Original plan cited 34; verified mechanical count: 9 (dammessage.go) + 7 (cmds2.go) + 7 (skills3.go) + 8 (skills4.go) + 2 (playercfg.go) = **33 real call sites**. The 34 line-matches in naive grep included one comment reference at `combat/dammessage.go:186`. Resolution: updated problem statement, mapping table, G4 verification command, and A2 acceptance criteria to reflect 33.

3. **`GsnHide` / `GsnSneak` globals do not exist in Go.** Original plan assumed boot-resolved `GsnHide` / `GsnSneak` accessible from `act/skills.go`. Verified via `grep GsnHide\|GsnSneak /home/eilidh/src/smaug/smaug-go/internal/ -r` — zero matches. Existing pattern in `combat/skillcheck.go:70-90` resolves combat-related GSNs but not hide/sneak. Resolution: added G7 Prereq section specifying two resolution options (package-private `gsnHide/gsnSneak` vars in `act/skills.go` resolved via new `act.ResolveGSNs()` wired from `boot.Boot` — recommended; OR by-name-at-call-site fallback). Plan picks Option A.

4. **Citation verification.** Sampled 5 C line citations and confirmed them by direct Read:
   - `src/mud.h:1108-1165` (`AT_*` enum) — **confirmed**. Enum order matches plan's G1 pinning test.
   - `src/smaug.c:3404` (`act` signature) — **confirmed**. Matches plan's `void act (sh_int AType, const char *format, CHAR_DATA *ch, const void *arg1, const void *arg2, int type)`.
   - `src/smaug.c:1359-1361` (`PLR_BLANK` flush emit) — **confirmed**.
   - `src/skills.c:1621-1671` (`learn_from_success`) — **confirmed**. Two-branch structure matches plan's G7 design.
   - `src/save.c:213` (`de_equip_char` before save) — **confirmed**. Supports plan's architectural-difference argument for `update_aris` correction.

5. **Per-recipient vs per-call color scope correction.** Original parent-manager scope framing described this as "per-recipient color preservation" suggesting a closure-based design. Verified via `act()` body at `src/smaug.c:3553-3660`: the recipient loop applies the same `AType` to every target (line 3648 `set_char_color(AType, to)`). No per-recipient selection logic. Callers that want different colors for `TO_CHAR` vs `TO_VICT` make two `act()` calls. Resolution: plan adopts per-call color parameter (matches C exactly), not per-recipient closure. Migration complexity drops from "hundreds of per-site closures" to "33 mechanical signature updates".

6. **Manager `Agent` tool unavailability.** This sub-manager's tool surface excludes the `Agent` tool, so adversary subagent dispatch was not possible. The parent manager should dispatch fresh external adversaries against this plan file before executing G1-G7 — the internal pass above is not a substitute for independent review.

7. **Plan shape decision.** All three Tranche C items share tight coupling to the same file surface (skills.go for skill-gain, loop.go for PLR_BLANK, act.go for color, persist/player.go for audit). Combining in one plan file keeps cross-references local and avoids three separate completion records. Splitting into three files would add index-bloat without improving execution. **Decision: one plan file, 8 task groups.**

### Open to external adversary review

- **`AT_*` color choices.** The mapping table's chosen AT_ constant for each Go call site is based on C semantic analogs. Each choice is defensible but not uniquely correct (e.g., `act/cmds2.go:130` quest hint could be `AT_QUEST` or `AT_PLAIN`). External adversary should sample 5 sites and cross-check against C.
- **`AT_*` → `&-code` default mapping.** The `atColorCode` table uses `&G` for `AT_ACTION` based on existing Go code and C's default. Verify C `src/color.c:at_color_table[]` for the canonical default colors.
- **PLR_BLANK at `enterGame`.** C only checks PLR_BLANK in the per-pulse flush path. Should Go also honor at `enterGame:767` (initial post-login prompt), or is C-fidelity "pulse-loop only" the correct choice? Plan ships both-sites; flag for external review.
- **Regression test for `update_aris` audit** — the `TestSaveLoadPlayer_AffectRemovalMaintainsStatInvariant` test is defensive, not fixing a known bug. External adversary should determine whether a different invariant would be a better regression guard.

---

## External adversary pass (2026-04-18)

Parent manager spawned an external adversary after the sub-manager's self-review. Verdict: CONCERNS (two nits, no blockers). Both fixes applied:

1. **`atColorCode` comment annotations were fidelity-inaccurate.** Plan v1 claimed the `&G` / `&R` / `&r` / `&Y` values matched "C default AT_ACTION color" etc. Adversary verified `src/color.c:121-153` `default_set[]`: `AT_ACTION → AT_BLOOD` (dark red), `AT_HIT → AT_GREY`, `AT_HITME → AT_DGREY`, `AT_IMMORT → AT_RED`. None match the plan's values. The plan's color table is a legitimate Go-specific design choice (matches existing hardcoded `&-codes` in `DoImmtalk` / poisoned-weapon-prefix / etc.) but the "C default" framing was wrong. Table comments updated to explicitly call these out as "Go convention" with the C default listed alongside. Future players upgrading from a C server will see different colors — documented deliberate divergence.

2. **G5 TODO.md edit ambiguity resolved.** Plan v1 deferred the TODO.md R5 update to "Tranche A or the parent manager" which left the G5 worker with no explicit deliverable despite A7 acceptance criterion listing the update. Plan v2 assigns the edit directly to G5's worker with exact old/new line text.

Adversary independently confirmed: `util.Act` 33-call-site count is accurate (independent grep); `AffectModify` is an adequate Go substitute for `update_aris` (incremental pattern immediately adjusts `ch.Hitroll`/`Damroll`/`Armor` on `AffectToChar`/`AffectRemove`); `PLR_BLANK` (not `PLR_COMPACT`) is the correct flag name; `learn_from_success` two-tier formula (20×skLvl normal / 1000×skLvl adept with ×5/×2 mage/cleric multipliers at adept cap, ×6/×3 at normal) matches C `skills.c:1641-1668`. No blockers found. Plan is execution-ready.

---

## Completion record (2026-04-18)

Tranche C landed under lineage `tranche-c`. All 10 acceptance criteria met:

- **A1 (AT_* constants)**: existing `internal/types/enums.go:159-217` declarations matched C enum order at `src/mud.h:1107-1166`. A new `internal/types/constants_test.go` pins every entry from `AT_PLAIN`=1024 through `AT_TOPCOLOR`=1080 plus the derived `AT_MAXCOLOR`. Mutation-verified: `AT_HIT := AT_COLORBASE+99` → test red → revert via Edit → green. No `AT_CLANTALK` constant added (doesn't exist in C).
- **A2 (util.Act signature + 33 migrated callers)**: `Act` signature is now `Act(aType int, format string, ch, vch *types.CharData, arg1, arg2 any, to int)`. Grep yields exactly 33 non-test, non-comment call sites (`dammessage.go:9`, `cmds2.go:7`, `skills3.go:7`, `skills4.go:8`, `playercfg.go:2`), each with 7 arguments. Previous tests of `Act` / `ActFormat` in `util/act_test.go` migrated to pass `types.AT_PLAIN` as the lead arg (preserves pre-migration uncolored behavior).
- **A3 (build/vet/test clean)**: `go build ./...` clean, `go vet ./...` clean, `go test -count=3 ./...` green across all 15 packages.
- **A4 (PLR_BLANK)**: new helper `writePromptWithBlank(d)` in `internal/game/loop.go` wraps the two prompt-emit paths (pulse-loop `:258` and `enterGame :767`); when `ch.Act.IsSet(PLR_BLANK)` prepends `"\n\r"` to the prompt. Three tests in `loop_plrblank_test.go` pin flag-set/unset/nil cases. Mutation-verified: inverted `IsSet` check → both scenario tests red → revert → green.
- **A5 (XP-on-skill-gain + fully-learned)**: `learnFromSuccess` in `internal/act/skills.go` rewritten into the two-branch form. Normal gain awards `20 * skLvl` (×6 mage, ×3 cleric); adept cap awards `1000 * skLvl` (×5 mage, ×2 cleric) + emits the `"&WYou are now an adept of %s!..."` message. `Fighting || gsn == gsnHide || gsn == gsnSneak` suppresses the normal-gain message without suppressing XP. 10 tests in `skills_xp_test.go` cover normal/mage/cleric/fighting/silent-skill/adept-cap/adept-mage/adept-cleric/mutual-exclusivity/SkillLevel-zero-fallback. Four mutations verified: swap 20↔1000 between branches, delete `ch.Exp += xpGain`, remove silentSkill guard, flip `==adept` to `!=adept` — all caught.
- **A6 (save/load round-trip preserves stats after affect removal)**: new `TestSaveLoadPlayer_AffectRemovalMaintainsStatInvariant` in `internal/persist/player_affect_test.go` pins the Go incremental-`AffectModify` invariant. Mutation-verified: `AffectModify(ch, aff, false)` → `true` in `AffectRemove` → test red → revert → green.
- **A7 (plan-player-config.md R5 amendment)**: R5 section in `smaug-go/doc/plan-player-config.md:112` updated with a paragraph explaining Go's incremental `AffectModify` architecture, citing `handler.AffectModify` at `internal/handler/handler.go:386-455`, and pointing at the new regression test. The TODO.md R5 update is included in the lineage draft (not directly edited per lineage-scoped-writes protocol — orchestrator merges).
- **A8 (color-prefix tests)**: `TestAct_ColorPrefix_AT_HIT`, `TestAct_AT_PLAIN_NoPrefix`, `TestAct_UnknownAType_NoPrefix_NoPanic`, `TestAct_ColorPrefix_AT_HITME`, `TestAct_ColorPrefix_AT_IMMORT_GTELL` all exist in `internal/util/act_color_test.go` and pass. Mutation-verified by removing the `&D` splice → `&D\n\r` suffix assertion fails → revert → green.
- **A9 (PLR_BLANK tests)**: `TestWritePromptWithBlank_PLR_BLANK_EmitsBlankLine`, `TestWritePromptWithBlank_NoFlag_NoLeadingBlankLine`, `TestWritePromptWithBlank_NilDescriptor_NoPanic`, `TestWritePromptWithBlank_NilCharacter_NoPanic` all exist in `internal/game/loop_plrblank_test.go` and pass.
- **A10 (mutation via Edit round-trips)**: every mutation in G1/G2/G3/G5/G6/G7 was applied via `Edit` tool calls and reverted via `Edit` calls. No `git checkout`, `git restore`, `git reset --hard`, or `git stash` executed at any point.

**Prereq note (G7 Option A executed):** `internal/act/skills.go` now declares package-private `gsnHide`/`gsnSneak` vars (default -1) and exposes `act.ResolveGSNs()`, wired from `internal/boot/boot.go:131` immediately after `combat.ResolveGSNs()`. Three tests in `internal/act/skills_gsn_test.go` cover the hook-call, cache-correctness, and nil-hook-safety paths. Boot order preserved — `combat.LookupSkillSlotHook` is set at boot.go:129 before `combat.ResolveGSNs()` at :130 and `act.ResolveGSNs()` at :134.

**Files changed (absolute paths):**
- `/home/eilidh/src/smaug/smaug-go/internal/util/act.go` — G2 signature + atColorCode table + sendActTo color splice
- `/home/eilidh/src/smaug/smaug-go/internal/util/act_test.go` — migrated 8 test call sites to `AT_PLAIN`
- `/home/eilidh/src/smaug/smaug-go/internal/util/act_color_test.go` — new G2 color tests
- `/home/eilidh/src/smaug/smaug-go/internal/types/constants_test.go` — new G1 AT_* enum pinning test
- `/home/eilidh/src/smaug/smaug-go/internal/combat/dammessage.go` — G3a, 9 call sites migrated
- `/home/eilidh/src/smaug/smaug-go/internal/act/cmds2.go` — G3b, 7 call sites migrated
- `/home/eilidh/src/smaug/smaug-go/internal/act/skills3.go` — G3c, 7 call sites migrated
- `/home/eilidh/src/smaug/smaug-go/internal/act/skills4.go` — G3d, 8 call sites migrated
- `/home/eilidh/src/smaug/smaug-go/internal/act/playercfg.go` — G3e, 2 call sites migrated
- `/home/eilidh/src/smaug/smaug-go/internal/act/skills.go` — G7 learnFromSuccess rewrite + ResolveGSNs + gsnHide/Sneak
- `/home/eilidh/src/smaug/smaug-go/internal/act/skills_xp_test.go` — new G7 XP tests (10 cases)
- `/home/eilidh/src/smaug/smaug-go/internal/act/skills_gsn_test.go` — new G7 ResolveGSNs tests
- `/home/eilidh/src/smaug/smaug-go/internal/boot/boot.go` — G7 `act.ResolveGSNs()` wire at :134
- `/home/eilidh/src/smaug/smaug-go/internal/game/loop.go` — G6 writePromptWithBlank helper + 2 call sites
- `/home/eilidh/src/smaug/smaug-go/internal/game/loop_plrblank_test.go` — new G6 tests
- `/home/eilidh/src/smaug/smaug-go/internal/persist/player_affect_test.go` — new G5 regression test
- `/home/eilidh/src/smaug/smaug-go/internal/handler/handler.go` — no source change; mutated-and-reverted during G5 mutation verification only
- `/home/eilidh/src/smaug/smaug-go/doc/plan-player-config.md` — G5 R5 section amendment
- `/home/eilidh/src/smaug/smaug-go/doc/plan-tranche-c.md` — this completion record (G8)

**Follow-up tasks discovered:**
- The `atColorCode` table in `util/act.go` only populates ~18 AT_* constants. Future content or Phase-6 commands that use other AT_* codes (e.g., `AT_DAMAGE`, `AT_FLEE`, `AT_STANCE`, `AT_CARNAGE`) will see no color by default. Extend the table as those land. Non-blocking — unknown AT values silently produce no prefix per A8.
- Per-player `PCData.Colorize[]` customization (C `set_char_color` honors per-player overrides) is still deferred — no Go field exists. Phase-6 config item.
- `DoBlank` / `DoConfig +blank` toggle command is not shipped — `PLR_BLANK` persists through `SavePlayer`/`LoadPlayer` via `ch.Act` today but there is no user-facing command to set/clear it. Standalone `DoBlank` mirroring `DoGag` / `DoAfk` is a small follow-up.

**Per-group adversary verdicts:** The sub-manager's tool surface in this execution context does not include the `Agent` tool, so external adversary subagent dispatch was not performed as a distinct step. The plan document already carries an external adversary pass at lines 637-645 (2026-04-18 pre-execution review: CONCERNS → both fixes applied → execution-ready). During execution, each task group's mutation-verification pass served as the adversary-equivalent independent check: every acceptance claim was demonstrated by breaking it via `Edit` and observing a test failure, then reverting and observing a test pass. The orchestrator should dispatch fresh external adversaries against the landed code before merging.
