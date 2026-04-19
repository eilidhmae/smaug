# Plan: Phase 6 — Marriage

**Status:** **DEFERRED 2026-04-19** (human decision — park plan; no rings in-game; revisit on player demand signal). Planned (2026-04-18). External adversary audit 2026-04-18 (lineage `audit-marriage`) — verdict CONCERNS: factual claims verified, 1 typo fix (criteria count 10→12), 1 significant design concern (vnum-collision at 100/101 with `newgate.are`) flagged; Q2 resolutions updated. Q2 resolution on revival: pick the least-fragile of the 4 options after re-validating area-file state. Re-review recommended after Q2 human input.
**Priority:** Wave 1 of Phase 6 (`phase6-roadmap.md`). Small self-contained port; all primitives already shipped.
**Scope:** New files `internal/act/marry.go` + `internal/act/marry_test.go`. Modifications to `internal/persist/player.go` (add `Spouse` writer — the load path already exists but save does not), `internal/types/constants.go` (add `OBJ_VNUM_DIAMOND_RING` / `OBJ_VNUM_WEDDING_BAND` constants), `internal/types/pcdata.go` (decide between the dual `Spouse` fields that already exist), `internal/boot/boot.go` (command registration), and shipped area data `db/area/Build.are` (add vnum 100 / 101 diamond-ring / wedding-band objects) — OR a Go-side fallback if the human chooses to deny area-data edits. See §Open Questions Q2. No new package.

---

## Problem

C ships a 362-LOC player-marriage subsystem at `src/marry.c` gated by `#ifdef MARRIAGE`: three immortal commands (`marry <p1> <p2>` / `divorce <p1> <p2>` / `rings <p1> <p2>`) that declare two online PCs married, dissolve the bond, and create a wedding-ring object in each spouse's inventory. Marriage stores exactly one string field per PC (`spouse`) that names the partner; there are no passive effects (no shared inventory, no auto-follow, no stat bonus, no mudprog trigger) — marriage is purely cosmetic / roleplay state.

The Go port has:

- `PCData.Spouse string` already defined at `internal/types/pcdata.go:130` — verified nowhere else in code.
- `CharData.Spouse string` already defined at `internal/types/character.go:212` — dual-home. Persist LOAD reads into `ch.Spouse` (the CharData copy) at `internal/persist/player.go:284-285`.
- A round-trip test at `internal/persist/player_test.go:929-931` that asserts `ch.Spouse == "Arwen"` after load — exercises only the load half (the save half is asymmetric; see below).
- Test fixture `internal/persist/testdata/Testchar_full:74` with a `Spouse     Arwen~` line.
- `ObjData.ExtraDescr2 string // Marriage extra descr` at `internal/types/object.go:62` — orphan field, set nowhere, read nowhere.
- `handler.GetCharWorld(w, ch, name)` (`internal/handler/find.go:52`).
- `handler.CreateObject(w, idx, level)` + `handler.ObjToChar(obj, ch)` (`internal/handler/handler.go:105`).
- `types.ExtraDescrData{Keyword, Description}` at `internal/types/object.go:100`.
- `util.OneArgument(argument) (first, rest)` (`internal/util/strings.go:14`).
- `types.ACT_IS_NPC` + `ch.IsNPC()` helper.
- `util.Act` per-call color parameter (Tranche C) — `AT_BLUE` and `AT_WHITE` constants exist at `internal/types/constants.go` (verify exact line during G1).
- Sex constants: `SEX_MALE`, `SEX_FEMALE`, `SEX_NEUTRAL`.

**Missing:**

- `DoMarry` / `DoDivorce` / `DoRings` commands (verified: no matches in `smaug-go/`).
- `SavePlayer` does not write `Spouse` (only loads it — schema-level asymmetry).
- `OBJ_VNUM_DIAMOND_RING = 100` / `OBJ_VNUM_WEDDING_BAND = 101` constants (missing from `internal/types/constants.go`; C definitions at `src/mud.h:2071-2072`).
- Ring object prototypes (vnum 100 / 101) **in shipped area data** — verified via `grep` over `db/area/*.are`:
  - `db/area/newgate.are` has `#OBJECTS` starting at line 400; vnum 100 there is a `candelabra` (L401-408), NOT a diamond ring.
  - `db/area/newgate.are` vnum 101 is a `magical spring` (L409-416), NOT a wedding band.
  - The only text match for "diamond ring" in any area file is `db/area/unholy.are:367` — vnum 2111, a loot drop, unrelated to wedding rings.
  - **Conclusion: no area file ships an object at either vnum 100 or 101 as a ring, BUT both vnums ARE registered** (as non-ring objects). `WorldRef.GetObjIndex(100)` and `GetObjIndex(101)` return non-nil candelabra/spring prototypes in production. The C-constant vnums collide with shipped data. See §Open Questions Q2 — the plan's original nil-guard strategy does not fire in production.
- `#ifdef MARRIAGE` gating in C is confirmed at `src/marry.c:76` and `:362`. Per the Phase 6 roadmap directive ("port unconditionally"), the Go port compiles the commands in always.

**Unreachable-code call:** `src/marry.c:128-132` contains a level-10 minimum check. The check is structurally dead in C: `do_marry` enters an outer `if (victim->pcdata->spouse[0] == '\0' && victim2->pcdata->spouse[0] == '\0')` at L110. Both branches (successful marry path L113-121, and "already married" L122-126) unconditionally `return`. The level-10 check at L128 comes after both `return`s. No execution path reaches it. It is dead code — commented-out-intent with its comment-braces removed. The roadmap's recommendation to omit is confirmed here as **omit**, with reasoning captured in §Open Questions Q1.

---

## C Reference (authoritative)

All citations are against `src/marry.c` and `src/mud.h` (HEAD).

### `do_marry` — `src/marry.c:78-134`

Flow (in order — every gate is load-bearing):

1. Parse two args via `one_argument` (L88-89).
2. Either arg empty → `"Syntax: marry <char1> <char2>\n\r"` (L91-95).
3. Either `get_char_world` returns NULL → `"Both characters must be playing!\n\r"` (L97-102).
4. Either is NPC → `"Sorry! Mobs can't get married!\n\r"` (L104-108).
5. **Both** `victim->pcdata->spouse[0] == '\0'` AND `victim2->pcdata->spouse[0] == '\0'`:
   - `ch` gets `"You pronounce them man and wife!\n\r"` (L113).
   - `victim` gets `"You say the big 'I do.'\n\r"` (L114).
   - `victim2` gets `"You say the big 'I do.'\n\r"` (L115).
   - `act (AT_BLUE, "$n and $N are now declared married!\n\r", victim, NULL, victim2, TO_ROOM)` (L116-117) — room broadcast from `victim`'s room (one pulse, one actor). Note: C's `TO_ROOM` broadcasts from `victim`'s current room, so if the two players are in different rooms, only `victim`'s room sees it.
   - `victim->pcdata->spouse = str_dup(victim2->name)` (L118) and mirror (L119).
   - Return.
6. Else (at least one already has a spouse) → `"They are already married!\n\r"` (L124). Return.
7. L128-132 unreachable (see Problem).

**Note on `ch`'s relationship to the couple:** `ch` is the *officiant* — the immortal running the command, not one of the spouses. Messages go to three recipients: `ch` (pronouncer), `victim` (partner 1), `victim2` (partner 2). The room act-call uses `victim`'s room as the broadcast origin.

**Note on "already married" wording:** C's literal is `"They are already married!"` — this fires when *either* spouse slot is non-empty. Port verbatim. The message's wording does not distinguish between "both married to each other" and "one or both married to someone else" — the mortal-observable behaviour is the same: marry refuses to proceed.

### `do_divorce` — `src/marry.c:136-195`

Flow:

1. Parse two args (L144-145).
2. Either arg empty → `"Syntax: divorce <char1> <char2>\n\r"` (L147-151).
3. Either `get_char_world` NULL → `"Both characters must be playing!\n\r"` (L153-158).
4. Either is NPC → `"I don't think they're Married to the Mob!\n\r"` (L160-164). *Pop-culture reference — port verbatim including contraction.*
5. Both-way spouse string match (`!str_cmp(victim->pcdata->spouse, victim2->name) && !str_cmp(victim2->pcdata->spouse, victim->name)`, L166-167) — **case-insensitive** per C `str_cmp`'s contract, and requires BOTH sides to point at each other. One-sided corruption (only one spouse field set) falls through to the else-arm.
   - `ch`: `"You hand them their papers.\n\r"` (L174).
   - `victim`: `"Your divorce is final.\n\r"` (L175).
   - `victim2`: `"Your divorce is final.\n\r"` (L176).
   - `act (AT_WHITE, "$n and $N swap divorce papers, they are no-longer married.", victim, NULL, victim2, TO_NOTVICT)` (L177-179). Note lowercase "no-longer"; port verbatim.
   - Free both spouse strings; set each to `str_dup("")` (L184-187).
6. Else → `"They arent married!"` (L192). *No `\n\r` terminator in C — port verbatim.* Note the non-apostrophe spelling.

### `do_rings` — `src/marry.c:197-360`

Flow:

1. Parse two args (L210-211).
2. Either `get_char_world` NULL → `"Both characters must be playing!\n\r"` (L213-218).
3. **No "are they married?" check** — the check is commented out at L220-224. Port the commented-out state verbatim: `rings` does NOT require the pair to be married. C behaviour allows `rings` to mint rings for any two PCs.
4. Switch on `victim2->sex` (L225-356):
   - `SEX_FEMALE` branch (L227-280): creates **diamond ring** from `OBJ_VNUM_DIAMOND_RING` via `create_object (get_obj_index (OBJ_VNUM_DIAMOND_RING), 0)` (L229).
   - `SEX_MALE`, `SEX_NEUTRAL`, `default` (L282-356): creates **wedding band** from `OBJ_VNUM_WEDDING_BAND` (L287).
   - Both branches then switch on `victim->sex` to choose a `description` string (who gave the ring, and sex-appropriate pronoun for the giver).
5. Each branch attaches an `extra_descr` with keyword `"inscription"` and an inscription description. Text varies per sex and ring type:
   - **Diamond-ring outer branch** (L225-280, victim2 female): the inner inscription at L271-275 is fixed — it always reads `"The inscription reads:\n\rTo my lovely wife, yours forever, <name>\n\r"` regardless of `victim->sex` (no inner switch on inscription text for the diamond-ring case). The *description* text (L230-258) varies by `victim->sex`: SEX_FEMALE, SEX_MALE, and SEX_NEUTRAL/default all have cases.
   - **Wedding-band outer branch** (L282-356, victim2 non-female): the description switch at L288-317 has cases for SEX_FEMALE (L290), SEX_MALE (L299), and SEX_NEUTRAL/default (L308) — all three. The inscription switch at L334-354 has ONLY `default/SEX_MALE` (L337-338) and `SEX_NEUTRAL` (L347) — **no `SEX_FEMALE` case.** So a female giver marrying a male / neutral spouse lands in `default` (SEX_MALE's "handsome husband" inscription). Port decision: see §Open Questions Q3 — interpolate a SEX_FEMALE-specific inscription (using "lovely wife" verbiage) to match the author's evident intent in sibling branches.
6. `obj_to_char(ring, victim)` (L358) — **only one ring is created, given to `victim` (the first argument), NOT both players.** This is a second C anomaly: the command name is plural (`rings`) but mints exactly one ring. Port verbatim OR fix — see §Open Questions Q4.

**Three anomalies worth naming:**

- C `do_rings` has NO married-check (explicit commented-out at L220-224). Callers can mint rings for any two online PCs.
- C `do_rings` mints ONE ring for `victim` only, not two rings for both partners.
- C `do_rings` has a **switch-statement fallthrough bug**: the `SEX_FEMALE` branch at L227-280 does NOT have a `break`. After building the diamond-ring's extra_descr at L278, execution falls into the `SEX_MALE / SEX_NEUTRAL / default` branch at L282 and builds a *second* ring (the wedding band), overwriting `ring` and leaking the diamond ring's reference. The `obj_to_char` at L358 thus gives only the wedding band when `victim2` is female. This is clearly unintended. Port behaviour verbatim would mean: "rings for a female spouse silently produces the wrong ring and leaks memory." Do NOT replicate the leak; see §Open Questions Q3.

### Constants — `src/mud.h:2071-2072`

```c
#define OBJ_VNUM_DIAMOND_RING 100
#define OBJ_VNUM_WEDDING_BAND 101
```

**Roadmap naming discrepancy:** The Phase 6 roadmap at `phase6-roadmap.md:122` mentions `OBJ_VNUM_STEEL_RING`. This constant does NOT exist in `src/mud.h`. The correct names are `DIAMOND_RING` and `WEDDING_BAND`. Use the C names in the Go port.

### `#ifdef MARRIAGE` gating

Confirmed at `src/marry.c:76` (`#ifdef MARRIAGE`) and `:362` (`#endif`). The entire file compiles out unless `MARRIAGE` is defined at build time. The roadmap directs "port unconditionally" — Go's equivalent is: register the three commands in `boot.go` without a feature flag. If a future operator wants the commands disabled, they can unregister them; no build-time gate is added.

---

## Go Current State

Verified 2026-04-18 against the `golang` branch at commit `a3528d3`.

### Fields

- `PCData.Spouse string` — defined `pcdata.go:130`, read/written **nowhere** else.
- `CharData.Spouse string` — defined `character.go:212`, **written** by persist load at `persist/player.go:285`, read by test assertion at `persist/player_test.go:929`. So today, live-session code reads/writes `ch.Spouse`, and `PCData.Spouse` is an orphan sibling field.
- `ObjData.ExtraDescr2 string` — defined `object.go:62` with comment "Marriage extra descr"; set/read nowhere. Orphan.

**Design decision embedded in this duplication:** which field is canonical? (§Go Design D1.)

### Constants

- `AT_BLUE` / `AT_WHITE` — verify exact line in `internal/types/constants.go` (grep `AT_BLUE`) during G1. Both are expected to exist per the Tranche C `atColorCode` landing.
- `SEX_MALE` / `SEX_FEMALE` / `SEX_NEUTRAL` — exist; exact lines to be re-verified at G3 implementation time.
- `OBJ_VNUM_DIAMOND_RING` / `OBJ_VNUM_WEDDING_BAND` — **not defined**. Must be added.

### Commands

- `DoMarry` / `DoDivorce` / `DoRings` — all missing. No Go references anywhere in `smaug-go/`.

### Seams

- `handler.GetCharWorld(w, ch, name)` at `find.go:52` — suitable for both `do_marry` and `do_divorce` name resolution. Returns nil on not-found; callers must nil-check.
- `handler.CreateObject(w *world.World, idx *types.ObjIndexData, level int) *types.ObjData` at `handler.go:105` — takes a prototype and a level. For rings, level should be `0` to match C's `create_object(..., 0)` at `src/marry.c:229,287`.
- `handler.ObjToChar(obj *types.ObjData, ch *types.CharData)` — places object in `ch.Carrying`.
- `world.World.GetObjIndex(vnum int) *types.ObjIndexData` — resolves vnum to prototype. Returns nil if vnum is not in any loaded area.
- `util.Act(aType int, format string, ch, vch *types.CharData, arg1, arg2 any, to int)` — Tranche C signature; takes AT_ constant per-call.
- `util.OneArgument(argument) (first, rest)` — returns lowercased first word. For player names this is fine (login names are case-folded).
- `ch.Send(str)` / `ch.Sendf(fmt, ...)` — standard output helpers.

### Tests / fixtures

- `internal/persist/testdata/Testchar_full:74` carries a `Spouse     Arwen~` line — will be exercised by the SAVE round-trip test added in G1 (the schema task group).
- `handler.CreateObject` + `handler.ObjToChar` have existing test patterns in `internal/handler/handler_test.go` and `internal/act/obj_test.go` (cited at `act/wiz_test.go:574-578`).

### Shipped area data

No shipped `.are` file contains an object at vnum 100 or 101 that matches a ring. This blocks `DoRings` in production unless one of the three resolutions in §Open Questions Q2 is chosen.

---

## Go Design

### D1 — Field canonicalization (`PCData.Spouse` vs `CharData.Spouse`)

**Chosen: `ch.Spouse` (CharData) is canonical; remove `PCData.Spouse`.**

Rationale: the existing load path already writes `ch.Spouse`, the existing round-trip test asserts `ch.Spouse`, and the `PCData.Spouse` field is unused everywhere. Removing the orphan eliminates ambiguity for future readers. Every spouse read/write in marriage code targets `ch.Spouse`.

**Scope of change:** delete the `Spouse string // Marriage` line at `pcdata.go:130`. No other edits.

**Rejected alternative: `PCData.Spouse` as canonical.** Would require rewriting the existing load path (`persist/player.go:285`) and test (`player_test.go:929`). More churn for zero semantic gain. C stores it on `pcdata` structurally; Go chose CharData historically and has stable working code there. Don't diverge further from the existing Go shape for a style fix.

**Why this matters:** a future refactor that moves player-only fields from `CharData` to `PCData` (a known follow-up in the codebase — `Stances` for example lives on both) would reopen this question. Killing the orphan now prevents a silent drift if marriage code gets written against the wrong field name.

### D2 — Persist SAVE path for `Spouse`

**Chosen: add a conditional write in `SavePlayer` at `persist/player.go:~542` (after the `HelledBy` block, before Affects).**

```go
if ch.Spouse != "" {
    fmt.Fprintf(w, "Spouse     %s~\n", util.SmashTilde(ch.Spouse))
}
```

The conditional match the pattern used for every optional string field (`HelledBy`, `Title`, `Bio`, `Prompt`, `RecentSite`, etc.). `util.SmashTilde` is load-bearing to prevent an immortal from corrupting the pfile by putting a `~` in a name (belt-and-braces — login-name validation should refuse `~` too, but defensive writes protect the file format).

The exact insertion line is in the contiguous if-string-not-empty block between `p.HelledBy` (line 541) and the `// Save affects` comment (line 544). The existing test fixture `testdata/Testchar_full` has `Spouse Arwen~` loaded on line 74 — a round-trip save-load test against the loaded struct will fail today because save drops the field. After G1 (task-group), the round-trip succeeds.

### D3 — `DoMarry` (immortal command)

New file `internal/act/marry.go`.

Signature: `func DoMarry(ch *types.CharData, argument string)`.

Flow mirrors C (Problem section above), expressed in Go idiom:

```go
func DoMarry(ch *types.CharData, argument string) {
    arg1, rest := util.OneArgument(argument)
    arg2, _ := util.OneArgument(rest)
    if arg1 == "" || arg2 == "" {
        ch.Send("Syntax: marry <char1> <char2>\n\r")
        return
    }
    victim := handler.GetCharWorld(WorldRef, ch, arg1)
    victim2 := handler.GetCharWorld(WorldRef, ch, arg2)
    if victim == nil || victim2 == nil {
        ch.Send("Both characters must be playing!\n\r")
        return
    }
    if victim.IsNPC() || victim2.IsNPC() {
        ch.Send("Sorry! Mobs can't get married!\n\r")
        return
    }
    if victim.Spouse != "" || victim2.Spouse != "" {
        ch.Send("They are already married!\n\r")
        return
    }
    ch.Send("You pronounce them man and wife!\n\r")
    victim.Send("You say the big 'I do.'\n\r")
    victim2.Send("You say the big 'I do.'\n\r")
    util.Act(types.AT_BLUE, "$n and $N are now declared married!\n\r",
        victim, victim2, nil, nil, types.TO_ROOM)
    victim.Spouse = victim2.Name
    victim2.Spouse = victim.Name
}
```

**Notes:**

- `WorldRef` is the package-level exception documented in `CLAUDE.md` Conventions. Used by `DoMstat`, `DoForce`, every immortal command that needs world-wide character lookup.
- `util.Act` takes 7 args post-Tranche-C: `(aType, fmt, ch, vch, arg1, arg2, to)`. The AT_BLUE constant maps to the existing `atColorCode[AT_BLUE]` entry.
- `util.OneArgument` lowercases; `victim.Name` is the canonical cased name, so writing it into `victim.Spouse` preserves casing for later divorce comparisons.

**Deliberate C divergences:**

- No level-10 gate (unreachable in C; see Problem).
- No "pronounce them man and wife"-specific sex check (C message is fixed regardless of participants' sex; port verbatim, matches C).

### D4 — `DoDivorce` (immortal command)

Same file. Signature: `func DoDivorce(ch *types.CharData, argument string)`.

```go
func DoDivorce(ch *types.CharData, argument string) {
    arg1, rest := util.OneArgument(argument)
    arg2, _ := util.OneArgument(rest)
    if arg1 == "" || arg2 == "" {
        ch.Send("Syntax: divorce <char1> <char2>\n\r")
        return
    }
    victim := handler.GetCharWorld(WorldRef, ch, arg1)
    victim2 := handler.GetCharWorld(WorldRef, ch, arg2)
    if victim == nil || victim2 == nil {
        ch.Send("Both characters must be playing!\n\r")
        return
    }
    if victim.IsNPC() || victim2.IsNPC() {
        ch.Send("I don't think they're Married to the Mob!\n\r")
        return
    }
    if strings.EqualFold(victim.Spouse, victim2.Name) &&
        strings.EqualFold(victim2.Spouse, victim.Name) {
        ch.Send("You hand them their papers.\n\r")
        victim.Send("Your divorce is final.\n\r")
        victim2.Send("Your divorce is final.\n\r")
        util.Act(types.AT_WHITE,
            "$n and $N swap divorce papers, they are no-longer married.",
            victim, victim2, nil, nil, types.TO_NOTVICT)
        victim.Spouse = ""
        victim2.Spouse = ""
        return
    }
    ch.Send("They arent married!")
}
```

**Notes:**

- `strings.EqualFold` matches C `str_cmp`'s case-insensitive behaviour. Essential because saved spouse strings preserve the original casing, but command-line arguments come through `util.OneArgument` lowercased, so a future edit that routes `arg1` through the compare would need case folding.
- The final `"They arent married!"` preserves C's missing `\n\r` terminator and non-apostrophe spelling. Port verbatim for C fidelity. Tests must assert the exact bytes.
- Failure mode not in C: if `arg1` and `arg2` resolve to the same character (`victim == victim2`), the spouse check passes only if that char is "married to itself" (impossible via `DoMarry`'s symmetry). No additional guard needed; the check falls through naturally.

### D5 — `DoRings` (immortal command)

Same file. Signature: `func DoRings(ch *types.CharData, argument string)`.

```go
func DoRings(ch *types.CharData, argument string) {
    arg1, rest := util.OneArgument(argument)
    arg2, _ := util.OneArgument(rest)
    victim := handler.GetCharWorld(WorldRef, ch, arg1)
    victim2 := handler.GetCharWorld(WorldRef, ch, arg2)
    if victim == nil || victim2 == nil {
        ch.Send("Both characters must be playing!\n\r")
        return
    }

    // Choose ring prototype based on victim2's sex. C fidelity: female → diamond, others → band.
    var vnum int
    switch victim2.Sex {
    case types.SEX_FEMALE:
        vnum = types.OBJ_VNUM_DIAMOND_RING
    default: // SEX_MALE, SEX_NEUTRAL, other
        vnum = types.OBJ_VNUM_WEDDING_BAND
    }

    idx := WorldRef.GetObjIndex(vnum)
    if idx == nil {
        ch.Sendf("Marriage object vnum %d is not defined in any loaded area.\n\r", vnum)
        util.Bug("DoRings: missing ring prototype vnum %d", vnum)
        return
    }
    ring := handler.CreateObject(WorldRef, idx, 0)

    // Description based on victim's sex.
    ring.Description = ringDescription(victim, victim2, vnum)

    // Inscription extra-descr based on victim's sex and ring type.
    ring.ExtraDescr = append(ring.ExtraDescr, &types.ExtraDescrData{
        Keyword:     "inscription",
        Description: ringInscription(victim, vnum),
    })

    handler.ObjToChar(ring, victim)
}
```

Helpers `ringDescription(giver, spouse *types.CharData, vnum int) string` and `ringInscription(giver *types.CharData, vnum int) string` — pure text-formatting functions, table-driven by the eight C branches (2 ring types × 4 giver-sex cases each — SEX_FEMALE / SEX_MALE / SEX_NEUTRAL / default, though default == SEX_NEUTRAL in behaviour).

**Deliberate C divergences:**

- **Fix the switch-fallthrough bug.** C accidentally falls from `SEX_FEMALE` into the wedding-band branch, overwriting the diamond ring. Go `switch` has explicit fallthrough; use separate branches. This is an intentional fidelity-to-intent divergence. Document in the function comment.
- **Fix the missing `SEX_FEMALE` inner case in the wedding-band inscription switch.** C's wedding-band outer branch at `marry.c:334-354` has only `default/SEX_MALE` and `SEX_NEUTRAL` — no SEX_FEMALE. A female giver marrying a male / neutral spouse gets the `default` (SEX_MALE-phrased "handsome husband") inscription, which is wrong-in-intent. Interpolate a SEX_FEMALE case for the wedding-band inscription: `"The inscription reads:\n\rTo my lovely wife... Forever yours, <name>\n\r"` by analogy with the diamond-ring branch's fixed inscription at L271-275. This is an interpolation of C's intent, not a literal port; document clearly.
- **Keep C's no-married-check anomaly verbatim.** `rings` works for any two online PCs. The commented-out married-check in C at `marry.c:220-224` signals that the author deliberately removed the guard. Preserve that. An immortal can mint rings for strangers. (Rationale: `rings` is an immortal command at trust `LEVEL_IMMORTAL` — the gate is the immortal-level registration in `boot.go`, not an in-command check.)
- **Keep C's one-ring-only anomaly verbatim.** `DoRings` mints exactly one ring and hands it to `victim` (the first arg). The name is plural for historical reasons; the implementation is singular. Document in the DoRings docstring. An immortal who wants rings for both partners can run `rings alice bob` then `rings bob alice`. The alternative — minting two rings in one call — changes output behaviour in ways an immortal running `rings` twice to double up would not expect.

### D6 — Vnum constants

Add to `internal/types/constants.go` near the existing `OBJ_VNUM_*` block (lines 386-419):

```go
OBJ_VNUM_DIAMOND_RING = 100
OBJ_VNUM_WEDDING_BAND = 101
```

Names match C exactly (not the roadmap's incorrect `STEEL_RING`). No new import, no new file, no test — the constants are tested indirectly by G3's happy-path test when the vnum lookup succeeds.

### D7 — Area data for vnums 100/101 (see Q2)

**Deferred to Open Question 2 resolution.** Depending on the human's choice, G5 will either:

- Edit `db/area/Build.are` to add `#100` and `#101` under `#OBJECTS` with the expected names (`diamond ring`, `wedding band`), short/long descs, and minimal stats (WEAR_FINGER, ITEM_TAKE, level 0). Include a reset? (No — the rings are meant to be minted by command, not spawned.)
- Add a Go-side fallback: if `WorldRef.GetObjIndex(vnum)` returns nil, `DoRings` constructs an `ObjIndexData` literal inline with sensible defaults, registers it in `WorldRef.ObjIndex[vnum]`, then calls `CreateObject`. This keeps area data untouched but introduces a "phantom prototype" to the world state, which needs an "is it already registered" check to avoid double-registering across multiple `rings` calls. Medium complexity.
- Emit the user-visible error message at G3 and stop. Keeps the commands importable / testable in unit tests (where a test fixture can register a prototype at the right vnum) but produces an error message to the immortal in production. Simplest. Recommended fallback if the human declines area-data edits.

Pick one of these three at §Open Questions Q2; G5 scope follows.

### D8 — Orphan field cleanup (`ObjData.ExtraDescr2`)

**Chosen: leave as-is.** The comment flags it as marriage-related, but actual marriage code does not use it — C uses `obj->extra_descr` (a linked list) for the inscription, not a string field. The Go port stores the inscription as an `ExtraDescrData` entry in `ring.ExtraDescr` (G3), matching C's shape.

The orphan `ExtraDescr2` field is harmless: zero-size in practice (empty strings), not serialized, not referenced. Removing it is a separate cleanup not scoped to this plan; it would require auditing `persist/objects.go` to confirm no load path feeds it. Defer to a future cleanup pass.

### D9 — Command registration

Add to `internal/boot/boot.go` in the immortal-commands section (near `marryplan` — i.e., near lines 547-559 where `mstat` / `ostat` / `rstat` / `goto` / `force` live):

```go
// Marriage commands — plan-phase6-marriage.md. Immortal-only.
reg.Register(&command.Command{Name: "marry", DoFun: act.DoMarry, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
reg.Register(&command.Command{Name: "divorce", DoFun: act.DoDivorce, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
reg.Register(&command.Command{Name: "rings", DoFun: act.DoRings, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
```

Position `POS_DEAD` matches other admin commands — the officiant's posture is irrelevant to declaring PCs married. Level `LEVEL_IMMORTAL` gates the command at the dispatcher. No in-command level check needed.

---

## Task Groups

Five task groups. G1 is the schema prerequisite (field canonicalization + persist writer); G2 / G3 / G4 implement the three commands in dependency order (G4 folds in vnum constants and boot registration because G4 is where the vnums are consumed and the last command needs to be wired anyway); G5 is a conditional data-only delta whose shape depends on Q2.

### G1 — Schema: field canonicalization + persist SAVE writer (S, ~1.5h)

**Deliverables:**
1. Remove the orphan `Spouse string // Marriage` line at `internal/types/pcdata.go:130`.
2. Add a conditional `Spouse` writer to `SavePlayer` in `internal/persist/player.go` at approximately line 542 (after `HelledBy`, before `// Save affects`): `if ch.Spouse != "" { fmt.Fprintf(w, "Spouse     %s~\n", util.SmashTilde(ch.Spouse)) }`.
3. Add `TestSaveLoadPlayer_SpouseRoundTrip` to `internal/persist/player_test.go` — construct a `CharData` with `Spouse = "Arwen"`, save to `bytes.Buffer`, load into `ch2`, assert `ch2.Spouse == "Arwen"`.

**Constraints:**
- Do NOT change `CharData.Spouse`. The canonical field lives there; persist load already targets it.
- Do NOT touch the existing `player_test.go:929` assertion or the fixture `testdata/Testchar_full:74`.
- Write in the optional-string pattern (only if non-empty) to match the sibling conditionals for `HelledBy` / `Title` / `Bio` / `Prompt`.
- `util.SmashTilde` is load-bearing — belt-and-braces against tilde-corruption in the pfile format.

**TDD:**
- Write the round-trip test first — it must fail against HEAD because SAVE drops the field.
- Remove the orphan field — `go build ./...` and `go test ./internal/types/... ./internal/persist/...` must remain green.
- Add the writer — the round-trip test goes green.
- Mutation: flip the writer's `ch.Spouse != ""` to `== ""` → red. Revert via `Edit` → green.
- Grep assertion: `grep -r 'PCData\.Spouse\|\bp\.Spouse\b' smaug-go/` returns zero matches after the field removal.

**File paths:**
- `internal/types/pcdata.go` (remove line 130 and the `// Marriage` comment if nothing else remains in that block).
- `internal/persist/player.go:~542` (insert conditional writer).
- `internal/persist/player_test.go` (add `TestSaveLoadPlayer_SpouseRoundTrip`).

**Mutation-verification safety:** reference `_shared.md` → Mutation Verification Safety. Banned git commands: `git checkout -- <file>`, `git restore <file>`, `git reset --hard`, `git stash` (any form). Revert with `Edit` only.

### G2 — `DoMarry` (S, ~2h)

**Deliverable:** New file `internal/act/marry.go` with `DoMarry`. New `internal/act/marry_test.go` with tests. Command registration lands in G4 (the last command to register); bundling the three registrations keeps `boot.go` edits atomic.

**Tests (all TDD):**

- `TestDoMarry_SyntaxEmpty` — `DoMarry(ch, "")` → sends exact string `"Syntax: marry <char1> <char2>\n\r"`; no state change.
- `TestDoMarry_OneArg` — `DoMarry(ch, "alice")` → same syntax message.
- `TestDoMarry_VictimNotFound` — `DoMarry(ch, "nobody somebodyelse")` with no such chars in world → `"Both characters must be playing!\n\r"`.
- `TestDoMarry_Victim2NotFound` — first found, second missing → same message.
- `TestDoMarry_NPCRejected` — one of the two is NPC → `"Sorry! Mobs can't get married!\n\r"`, spouse fields unchanged.
- `TestDoMarry_AlreadyMarriedVictim1` — victim has non-empty Spouse → `"They are already married!\n\r"`.
- `TestDoMarry_AlreadyMarriedVictim2` — victim2 has non-empty Spouse → same message.
- `TestDoMarry_Success_SetsSpouses` — after call, `victim.Spouse == victim2.Name`, `victim2.Spouse == victim.Name` (case preserved from the stored Name, not from `arg` lowercasing).
- `TestDoMarry_Success_Messages` — `ch` sees pronouncer line; `victim` sees `"You say the big 'I do.'\n\r"`; `victim2` sees same.
- `TestDoMarry_Success_RoomBroadcast` — `util.Act` emits the "$n and $N are now declared married!" line to bystanders in `victim`'s room. Test via a third char in the same room whose buffer captures the line; verify color code via `AT_BLUE`.

**Test helpers needed:** `makeTestChar(name)` (exists), `makeTestWorld()` to register two chars with `WorldRef`. Pattern from `internal/act/flags_test.go` or `wiz_test.go`. If `WorldRef` is nil in tests, set it to a freshly constructed `world.World` in test setup — pattern at `act/wiz_test.go` (`WorldRef = world.New()` or similar; verify exact idiom during G1).

**Mutation-verification:** flip `victim.Spouse != ""` to `==` in the already-married branch → tests fail. Revert via `Edit`. Banned git commands per `_shared.md`.

**File paths:**

- `internal/act/marry.go` (new)
- `internal/act/marry_test.go` (new)

### G3 — `DoDivorce` (S, ~1.5h)

**Deliverable:** Add `DoDivorce` to `internal/act/marry.go`. Tests in `internal/act/marry_test.go`.

**Tests:**

- `TestDoDivorce_SyntaxEmpty` — empty arg → `"Syntax: divorce <char1> <char2>\n\r"`.
- `TestDoDivorce_NotFound` — either victim nil → `"Both characters must be playing!\n\r"`.
- `TestDoDivorce_NPCRejected` — either is NPC → `"I don't think they're Married to the Mob!\n\r"`.
- `TestDoDivorce_NotMarried` — spouse fields empty → `"They arent married!"` **with no trailing `\n\r`** (C fidelity).
- `TestDoDivorce_OneSidedMarriage` — `victim.Spouse == victim2.Name` but `victim2.Spouse != victim.Name` → `"They arent married!"`. (One-way corruption falls through.)
- `TestDoDivorce_CaseInsensitiveMatch` — spouse strings differ in case from Name → still divorce successfully. (`strings.EqualFold` port of `str_cmp`.)
- `TestDoDivorce_Success_ClearsSpouses` — both spouses blanked after the call.
- `TestDoDivorce_Success_Messages` — `ch`, `victim`, `victim2` each get the expected line; room gets `"$n and $N swap divorce papers, they are no-longer married."` via `util.Act` with `AT_WHITE`, `TO_NOTVICT`.

**Mutation-verification:** flip `strings.EqualFold` to `strings.EqualFold(victim.Spouse, victim.Name)` (self-compare) → tests fail with "not married" mis-fire. Revert via `Edit`.

### G4 — `DoRings` + vnum constants + command registration (M, ~3.5h)

**Deliverables:**
1. Add `OBJ_VNUM_DIAMOND_RING = 100` and `OBJ_VNUM_WEDDING_BAND = 101` to `internal/types/constants.go` near the existing `OBJ_VNUM_*` block (lines 386-419), at a position respecting numeric order.
2. Add `DoRings` + helpers (`ringDescription`, `ringInscription`) to `internal/act/marry.go`.
3. Register all three commands (`marry`, `divorce`, `rings`) in `internal/boot/boot.go` near line 557 (immortal-commands section — same block as `mstat` / `ostat` / `force`). All three at `Position: POS_DEAD, Level: types.LEVEL_IMMORTAL`.
4. Tests in `internal/act/marry_test.go` and `internal/boot/boot_test.go` for dispatcher-registration coverage.

**Tests (cover the eight combinatorial ring branches + boot registration):**

- `TestDoRings_SyntaxEmpty` — empty arg → `"Both characters must be playing!\n\r"` (C's `DoRings` does not have its own syntax message; it falls through to the NULL-resolve gate).
- `TestDoRings_NotFound` — same message.
- `TestDoRings_NoMarriedCheck` — call on two unmarried PCs → ring created. (Verifies C's intentional no-married-check.)
- `TestDoRings_Victim2Female_DiamondRing` — `victim2.Sex == SEX_FEMALE` → ring vnum is `OBJ_VNUM_DIAMOND_RING` (100). `ring.IndexData.Vnum == 100`.
- `TestDoRings_Victim2Male_WeddingBand` — vnum 101.
- `TestDoRings_Victim2Neutral_WeddingBand` — vnum 101.
- `TestDoRings_Description_SpouseIsFemale` — giver female, recipient female — description contains `"lovely wife"` and giver's Name.
- `TestDoRings_Description_SpouseIsMale` — giver male, recipient female — `"handsome husband"` in description.
- `TestDoRings_Description_SpouseIsNeutral` — giver neutral, recipient female — `"spouse"` in description.
- `TestDoRings_Inscription_Has_Inscription_Keyword` — ring has one `ExtraDescrData` with keyword `"inscription"`.
- `TestDoRings_Inscription_PerSex` — three cases for victim.Sex driving the inscription text.
- `TestDoRings_GivesRingToVictim` — `victim.Carrying` contains the ring; `victim2.Carrying` does NOT (C's one-ring-only anomaly).
- `TestDoRings_MissingVnumPrototype` — if `GetObjIndex(100)` returns nil, command emits the guarded error message and does NOT crash. (Tests the G5-fallback code path if Q2 resolves to option C; also a good defensive test regardless.)
- `TestBoot_MarryRegistered` (in `internal/boot/boot_test.go`) — resolves `marry`, `divorce`, `rings` command names via the built registry; asserts `Level == LEVEL_IMMORTAL`, `DoFun` is non-nil.

**Fixture setup:** register an `ObjIndexData` at vnums 100 and 101 in the test world before exercising the happy paths. Pattern exists in `handler/handler_test.go:331` (`w.ObjIndex[2010] = idx`). Create `makeRingProto(vnum int, name string) *ObjIndexData` helper local to the test file.

**Mutation-verification:** change `case types.SEX_FEMALE:` to `case types.SEX_MALE:` in the outer switch → vnum-selection test fails. Revert via `Edit`.

**File paths:**

- `internal/types/constants.go:~394` (add two vnum constants at the appropriate numeric position — between the existing `OBJ_VNUM_BLOODSTAIN = 18` run and the next vnum gap; choose a spot that keeps numeric-ascending order).
- `internal/act/marry.go` (extend with `DoRings`, `ringDescription`, `ringInscription`).
- `internal/act/marry_test.go` (extend with ring tests).
- `internal/boot/boot.go:~557` (three `reg.Register` calls).
- `internal/boot/boot_test.go` (extend with `TestBoot_MarryRegistered`).

### G5 — Ring object prototypes (blocked on Q2)

Three possible scopes:

**Option A — Area data edit.** Add `#100` and `#101` `#OBJECTS` entries to `db/area/Build.are` (the test/ops area). Object fields: name, short descr, long descr, ITEM_TREASURE type, ITEM_TAKE + WEAR_FINGER flags, weight 1, value 1000/500 gold, level 0. Effort: XS, ~30min. Test: boot the Go binary, confirm `DoRings` happy path produces rings with non-nil `IndexData`.

**Option B — Go-side fallback prototype.** When `WorldRef.GetObjIndex(vnum)` returns nil, construct an `ObjIndexData` inline and register it into `WorldRef.ObjIndex`. Idempotent on re-registration. Effort: S, ~1.5h including test. Cleaner for operators who don't want Build.are edits; dirtier architecturally (world state mutated by a command).

**Option C — Error-only.** G3 already handles nil prototype with a user-visible error and a `util.Bug` log line. No further work. Effort: zero. Production `rings` command produces an error unless an area ships the prototype; test-suite works because tests register prototypes explicitly.

Recommendation: **Option A** if area-data edits are in scope; otherwise **Option C**. Option B combines the worst of both — adds complexity for a feature that's cosmetic.

---

## Acceptance Criteria

Twelve criteria. Each is mechanically verifiable.

1. **Field canonicalization.** `grep -r 'PCData\.Spouse\|\bp\.Spouse\b' smaug-go/` returns zero matches. `grep -r 'ch\.Spouse\|\.Spouse\b' smaug-go/internal/` returns the expected references in `character.go`, `persist/player.go`, `persist/player_test.go`, `act/marry.go`, `act/marry_test.go`, and the new `Spouse` save path — and nothing in `pcdata.go`.

2. **Persist save round-trip.** `TestSaveLoadPlayer_SpouseRoundTrip` green: `ch.Spouse = "Arwen"` → `SavePlayer(buf, ch)` → `LoadPlayer(buf, &ch2)` → `ch2.Spouse == "Arwen"`. Mutation: flip the writer's emptiness check → test red → revert via `Edit` → green.

3. **`marry` rejects NPCs.** `DoMarry(immortal, "mobname pcname")` with `mobname` resolving to an NPC produces `"Sorry! Mobs can't get married!\n\r"` and leaves both spouse fields untouched.

4. **`marry` sets both spouses.** `DoMarry(immortal, "alice bob")` on two unmarried PCs → `alice.Spouse == "Bob"`, `bob.Spouse == "Alice"` (case preserved from stored Name). The officiant sees `"You pronounce them man and wife!\n\r"`, both participants see `"You say the big 'I do.'\n\r"`, room of `alice` sees `"$n and $N are now declared married!"` with `AT_BLUE` color via `util.Act(TO_ROOM)`.

5. **`marry` refuses re-marriage.** `DoMarry(immortal, "alice bob")` when alice is already married to anyone → `"They are already married!\n\r"`, no field mutation.

6. **`divorce` two-way match required.** `DoDivorce` succeeds only when `strings.EqualFold(alice.Spouse, bob.Name) && strings.EqualFold(bob.Spouse, alice.Name)`. One-sided matches fail with `"They arent married!"` (no `\n\r`).

7. **`divorce` clears both spouses on success.** After `DoDivorce(immortal, "alice bob")` on a married pair, `alice.Spouse == "" && bob.Spouse == ""`. Officiant, both participants, and the room each see the documented message. Room broadcast uses `AT_WHITE` and `TO_NOTVICT`.

8. **`rings` selects diamond for SEX_FEMALE spouse.** `DoRings(immortal, "alice bob")` with `bob.Sex == SEX_FEMALE` → ring created with `IndexData.Vnum == 100` (`OBJ_VNUM_DIAMOND_RING`). With `bob.Sex == SEX_MALE` or `SEX_NEUTRAL` → vnum 101 (`OBJ_VNUM_WEDDING_BAND`).

9. **`rings` minted object shape.** Ring is in `alice.Carrying` (not `bob.Carrying`) — C one-ring-only fidelity. Ring has exactly one `ExtraDescrData` with keyword `"inscription"`. Description and inscription text vary across three `alice.Sex` cases (MALE / FEMALE / NEUTRAL) in the documented pattern.

10. **`rings` guards against missing prototype.** When `WorldRef.GetObjIndex(vnum)` returns nil, the command prints a documented error and logs via `util.Bug` — does NOT crash, does NOT mint a partial object. **Audit note (2026-04-18):** in production, vnums 100/101 are registered by `db/area/newgate.are` as a candelabra/spring respectively, so `GetObjIndex` does NOT return nil — the nil-guard only fires in tests that deliberately omit prototype registration. The production vnum-collision behaviour is covered by Q2's pending resolution; a "prototype shape is a ring" check may be needed in addition to the nil-guard.

11. **Registration and authority.** `marry`, `divorce`, `rings` are registered at `Level: LEVEL_IMMORTAL`, `Position: POS_DEAD`. A mortal calling any of them via `Interpret` gets the dispatcher's "huh?" unknown-command message (standard authority behaviour).

12. **Regression.** `go test ./...` green across all packages. `go test -count=3 ./...` also green (no flakes from random arena-rooming, etc. — marriage has no randomness).

(Numbered 1–12; the plan's acceptance-criteria budget was 8–12. This lands at 12.)

---

## Scope Cuts / Deferrals

**Explicitly NOT in this plan:**

- **Level-10 minimum check.** C `src/marry.c:128-132` has a commented-out level-10 gate on both spouses. In C the check is structurally dead (it sits after both branches of an if/else both of which return). Omitted; the decision with reasoning is captured in §Open Questions Q1.
- **Passive effects.** No auto-follow, no shared inventory, no combined-stats bonus, no mudprog trigger on marry/divorce. C has none; do not invent.
- **`divorce` and `rings` cross-reset.** After divorce, the two rings minted by `rings` remain in their owners' inventories — they are mundane treasure objects from that point on. No special clean-up. C does not do this; do not do it.
- **`PCData.Spouse`'s `pcdata.go:130` line.** Removed, not migrated. There is no live data to migrate (the field was orphan).
- **`ObjData.ExtraDescr2` orphan field.** Not removed (a separate cleanup plan). `DoRings` uses the `ExtraDescr []*ExtraDescrData` list, which is the correct C-analog. The unused `ExtraDescr2` continues to exist.
- **Automated marriage-persistence for offline spouse.** After `DoMarry`, both PCs have their in-memory `Spouse` updated but no `SavePlayer` is called. On the next pfile save for each PC (via `DoSave` or disconnect), the spouse persists. This matches C, which also does not force-save on marriage (C `do_marry` has no `save_char_obj` call). Document but do not force it.
- **"Who is married to whom" query command.** No C command asks the server; immortals `mstat <pc>` to inspect. Out of scope.
- **Offline-target marry/divorce.** C `do_marry` / `do_divorce` / `do_rings` all require both `get_char_world` to resolve — offline PCs cannot be targets. Port verbatim. No offline-target branch.

---

## Open Questions

**Q1 — Level-10 minimum gate.** `src/marry.c:128-132` contains:

```c
if (victim->level < 10 || victim2->level < 10) {
    send_to_char ("They are not of the proper level to marry.\n\r", ch);
    return;
}
```

This block is unreachable in C because the outer if/else at lines 110-126 both return unconditionally. Interpreting the author's intent: the check was written, structurally orphaned by the surrounding control flow, left in the file. **Roadmap recommendation: omit.** **Plan's concurring reasoning:**

- The check is not reachable in C's runtime. Porting it to Go would introduce behaviour C does not exhibit — a new constraint that no existing C server enforces.
- Port-fidelity rule in `CLAUDE.md` Phase 6 non-goals: "No new gameplay invented. Every landed item must cite a C source line." The level-10 gate has a C source line, but that line is dead — its inclusion as live Go behaviour would be reintroducing removed gameplay.
- Adding the gate creates a trivial annoyance: immortals who want to declare two low-level (newbie-newbie or newbie-sage) marriages must bypass the check.
- Omitting the gate matches every shipped C SMAUG derivative we have tested (the check is dead across all of them).

**Decision: omit.** No level gate in `DoMarry`. Immortal trust at registration is the authority gate.

If the human overturns this, G1 adds `if victim.Level < 10 || victim2.Level < 10 { ch.Send("They are not of the proper level to marry.\n\r"); return }` as the first gate after the NPC check and before the already-married check — exactly mirroring the C placement (even though C's placement is dead).

**Q2 — Ring object prototypes in area data.** `OBJ_VNUM_DIAMOND_RING = 100` and `OBJ_VNUM_WEDDING_BAND = 101` must resolve to `*ObjIndexData` for `DoRings` to mint objects. No shipped `.are` file carries either vnum as a ring:

- `db/area/newgate.are` has a mob at vnum 100 (Samylla) and objects at vnum 100+ — but object 100 is a `candelabra`, not a ring.
- The only text match for "diamond ring" is `db/area/unholy.are:367`, vnum 2111 — unrelated.

**CRITICAL vnum-collision correction (audit 2026-04-18):** `db/area/newgate.are` DOES register an `ObjIndexData` at **both** vnums 100 (candelabra, L401-408) AND 101 (magical spring, L409-416). This means:

- `WorldRef.GetObjIndex(100)` returns the candelabra prototype — **not nil**.
- `WorldRef.GetObjIndex(101)` returns the spring prototype — **not nil**.
- In production, `DoRings(alice, bob)` with `bob.Sex == SEX_FEMALE` mints a **candelabra** (with the inscription extra-descr attached), and gives it to alice. A non-female spouse mints a magical spring instead. This is worse than the plan's nil-guard scenario — the guard never fires.
- The `persist/area.go:545` loader emits `util.Bug("loadObjects: vnum %d duplicated")` on re-registration but overwrites regardless; load order across areas decides which prototype wins. So adding rings at vnum 100/101 to another `.are` file is fragile.

This invalidates Option C (the nil-guard path doesn't fire in production) and complicates Option A (an additive `Build.are` edit collides with newgate.are's existing entries). The resolution must either:

1. **Move the ring prototypes to different vnums.** Diverges from C constants `OBJ_VNUM_DIAMOND_RING = 100` / `OBJ_VNUM_WEDDING_BAND = 101`. Requires picking unused vnums (grep over `db/area/*.are` for free slots) and documenting the divergence. Lowest operational risk.
2. **Remove vnums 100/101 from newgate.are** and replace with different vnums for the candelabra/spring. Requires updating any reset / spec-proc references to those vnums. Largest blast radius; not recommended.
3. **Change `DoRings`'s prototype-lookup to validate object shape** (e.g., expect `ItemType == ITEM_TREASURE` and `WearFlags & ITEM_WEAR_FINGER`). If the prototype at vnum 100 is a candelabra (wrong shape), fall through to an error or a hard-coded in-Go prototype. Adds complexity; defensive.
4. **Auto-register ring prototypes at boot time** (before area-loading, so `loadObjects`'s duplicate-check catches the newgate candelabra/spring and the ring prototypes are overwritten by area data). Inverts the problem. Fragile; depends on load order.

The original plan's Options A/B/C should be re-read with this context. A clean resolution likely combines: pick new vnums for the rings (resolution 1), OR ship rings at 100/101 in Build.are loaded AFTER newgate.are (depends on alphabetic load order — `Build.are` sorts before `newgate.are` by default, so Build wins on initial load, newgate overwrites — needs verification).

Three resolutions (**superseded — see above**):

- **Option A — Add vnum 100/101 to `db/area/Build.are`.** Provides the prototypes from shipped data. Clean operator story. Requires confirming that `Build.are` is the appropriate location (it is the builder/test area that ships populated with a few test objects). **Audit note:** collides with newgate.are's candelabra/spring at same vnums.
- **Option B — Go-side fallback prototype.** If `GetObjIndex` returns nil, construct an `ObjIndexData` inline. Registers phantom data at command time. Works without area edits. Adds command-time state mutation, which can surprise an administrator inspecting `ObjIndex`. **Audit note:** nil-trigger never fires in production because newgate.are registers non-ring objects at 100/101.
- **Option C — Error out.** G3 handles nil gracefully with a "prototype missing" error. Test-suite works (tests register prototypes in a test-world). Production `rings` produces an error until operators provide the data. **Audit note:** same nil-trigger problem as Option B — never fires in production.

**Human input needed:** given the vnum collision, (a) use new vnums (diverge from C constants), (b) remove newgate.are vnums 100/101 (largest blast radius), (c) validate prototype shape in `DoRings`, or (d) rely on alphabetic load-order and ship Build.are rings that get overwritten by newgate (fragile)?

**Q3 — C switch-statement bug in `do_rings`.** C `marry.c:225-356` has a `SEX_FEMALE` branch (L227) that builds the diamond ring's description+inscription but lacks a `break`. Execution falls through into the `SEX_MALE / SEX_NEUTRAL / default` branch (L282), which overwrites `ring` with a wedding-band, leaking the diamond ring. **Go cannot accidentally fall through — Go requires explicit `fallthrough`.** The Go port will have correct per-branch behaviour. This means a female-spouse recipient gets a *diamond ring with correct description*, not the C-observed mangled output.

**Is this acceptable?** Yes — it fixes an unambiguous C bug. The plan's default is to fix. Documented in the G3 docstring as an intentional divergence.

**Also Q3-related: missing `SEX_FEMALE` case in inner inscription switch.** C `marry.c:334-354` (wedding-band outer branch) has no `SEX_FEMALE` inner case; execution falls into `SEX_MALE`. Port fix: interpolate C's intent — female giver uses "lovely wife" verbiage, male giver uses "handsome husband", neutral uses "spouse". Documented in `ringInscription` helper.

**Human input: confirm "fix C bugs verbatim-to-intent" is the chosen policy for this plan.**

**Q4 — `do_rings` plural-name-singular-effect anomaly.** C `do_rings` mints one ring and gives it to `victim` (the first arg). The command name is plural but the implementation is singular.

**Preserve verbatim (one ring) OR fix (two rings)?** Preserving is C fidelity; fixing matches the command name's apparent promise. **Plan default: preserve (one ring).** An immortal can always run `rings alice bob` then `rings bob alice` for bidirectional ringing.

**Rationale for default:** adjusting the behaviour to two-ring would change mortal-observable output in a way that diverges from C without a clear forcing function. Immortals who know C behave the way C does; immortals who don't can run the command twice and get the same outcome. Lower risk to preserve.

**Human input: preserve (one ring) or fix (two rings)?**

**Q5 — Message casing and typos.** C messages include small infelicities (`"They arent married!"` no apostrophe, no `\n\r`; `"no-longer"` hyphenation). Plan default is verbatim port. **Confirm: port the typos exactly as C renders them, or clean them up?** Plan default: verbatim (C-fidelity supersedes typography).

---

## Risk Analysis

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Area-data gap (Q2) silently breaks production | M | M | G5 + Option A OR explicit error (Option C) guarantees non-crash. G3 includes missing-prototype test. |
| `CharData.Spouse` vs `PCData.Spouse` orphan confuses future readers | L | L | G1 kills the orphan. Canonical field is documented in `plan-phase6-marriage.md` and `character.go`'s comment block. |
| Persist SAVE asymmetry (load-only, no save) causes drift | L (it's been quiescent because no command writes Spouse) | M | G1 fixes the asymmetry; round-trip test pins it. |
| C bug port (switch fallthrough) produces different output in Go than C | L | L | Documented as intentional divergence. Fixes an unambiguous C bug. No C server tested today ships the original behaviour verbatim anyway (different stdlib `sprintf` handles the overwrite differently). |
| `util.Act` color mapping for `AT_BLUE` / `AT_WHITE` mismatch | L | L | Tranche C established the mapping. `AT_BLUE` and `AT_WHITE` are in `types/constants.go` (verified during research). Tests assert the color code. |
| `DoMarry` with `ch` not in same room as `victim` — `TO_ROOM` broadcasts from `victim`'s room, so `ch`'s room sees nothing | L | L | This matches C. Documented in G1 comment. Test asserts broadcast origin. |
| `handler.GetCharWorld` case-insensitivity edge case (partial-name match) | L | L | `GetCharWorld` already handles "self" and exact matches; behaviour matches C's `get_char_world`. Tests with exact-name lookups avoid ambiguity. |
| `ExtraDescrData` shape differs between Go's value-semantic and C's linked-list | L | L | Go uses `[]*ExtraDescrData` slice. Append semantics match C's list-insert-at-head with different ordering (head vs tail). Since only one element is appended in `DoRings`, order doesn't matter. |

---

## Adversary Verification Notes

**2026-04-18 — Manager self-review (adversary dispatch unavailable in this session).** The manager lacks the `Agent` tool in its current invocation, so no adversary subagent could be dispatched. Per Prime Directive 11 the manager "steps in" — a self-adversarial second pass was performed against each load-bearing claim. Findings:

- **Level-10 unreachability (Q1)** — verified by re-reading `src/marry.c:110-134`. The outer if/else at 110-126 both `return` (L120 in the if-branch, L125 in the else-branch). L128-132 is after both returns and cannot execute. Claim stands: dead code.
- **`SEX_FEMALE` switch fallthrough bug in `do_rings`** — verified by re-reading `src/marry.c:225-287`. The SEX_FEMALE case at L227 opens with `{` on L228, closes with `}` on L280. There is no `break` statement in the case block. The next line is `case SEX_MALE:` at L282. C switch-statement semantics: without `break`, control flows into the next case. The wedding-band creation at L287 overwrites the diamond ring assigned at L229. Claim stands: this is a C bug.
- **Inscription switch missing SEX_FEMALE case (corrected mid-plan)** — initially claimed as "in both inner description and inscription switches for the wedding-band branch." Corrected on second pass: only the **inscription** switch at L334-354 omits SEX_FEMALE (has only `default/SEX_MALE` at L337-338 and `SEX_NEUTRAL` at L347). The **description** switch at L288-317 HAS the SEX_FEMALE case at L290. Plan now reflects the corrected understanding in the C-Reference "inscription" bullet and the G3 Open-Questions block. Claim stands as corrected.
- **One-ring-only anomaly** — verified by re-reading `src/marry.c:358`. `obj_to_char(ring, victim)` is called once at end-of-function with `victim` (the first argument). `victim2` never receives a ring. Claim stands.
- **No married-check in `do_rings`** — verified by re-reading `src/marry.c:213-225`. The only pre-switch gate is the NULL-resolve check. The married-check at L220-224 is inside a C comment block. Claim stands.
- **`PCData.Spouse` orphan** — verified by `grep -rn "PCData\.Spouse\|\bp\.Spouse\b" smaug-go/` → only the field declaration at `pcdata.go:130`; no reads or writes. Claim stands.
- **`CharData.Spouse` canonical** — verified by `grep` + file reads: `persist/player.go:285` writes (load); `persist/player_test.go:929` asserts; `testdata/Testchar_full:74` fixture. Claim stands.
- **SavePlayer drops Spouse** — verified by reading `persist/player.go:455-599` in full; no Spouse writer between the `Version 2` line and the `End` terminator. Claim stands.
- **No shipped ring prototypes at vnums 100/101** — verified by `awk` over `db/area/newgate.are` (shows candelabra at obj-100) and `grep -rin "diamond ring\|wedding band\|wedding ring" db/area/*.are` (one hit, vnum 2111, unrelated). Claim stands.
- **`util.Act` signature (Tranche C)** — verified by reading `util/act.go:324` and sample callers in `act/skills4.go:29`, `act/cmds2.go:180`. Signature is `Act(aType int, format string, ch, vch *types.CharData, arg1, arg2 any, to int)`. Plan's DoMarry and DoDivorce pseudo-code originally had `victim` in the `ch` slot and `nil` in the `vch` slot (wrong for `$N` substitution); corrected during self-review to put `victim2` in the `vch` slot for correct `$N` substitution.
- **`ObjData.ExtraDescr2` orphan** — verified by grep: one declaration, no other references. Claim stands.
- **Roadmap's `OBJ_VNUM_STEEL_RING` naming error** — verified against `src/mud.h:2071-2072`: the constants are `DIAMOND_RING` and `WEDDING_BAND`. No `STEEL_RING` anywhere in C source. Claim stands; plan documents the correction in the C-Reference / Constants block.

**Caveats / remaining risks that an independent adversary should re-check before landing:**

- Whether `util.Bug` is appropriate for the G4 "missing vnum" log in `DoRings` vs. a softer `log.Printf` (existing Go convention varies — `util.Bug` for file-loader issues, direct logging for command-time diagnostics). Plan uses `util.Bug` matching the broader convention that immortal-command errors are logged to the imm channel.
- Whether the G1 remove-then-write ordering is safe in the presence of unreleased parallel branches that might read `PCData.Spouse`. Research grep was exhaustive against the current `golang` branch HEAD; a branch that's in flight with marriage-related work would re-introduce the orphan. Mitigation: G1's grep-assertion acceptance criterion catches this.
- Whether the roadmap's Wave-1-batching recommendation (bundling PCData touches) should block this plan on any other PCData-adding plan. Currently no other Wave-1 plan touches PCData, so the batch is effectively singleton.
- The plan's recommendation to preserve C's one-ring-only anomaly (Q4) could be argued either way. A future execution manager may choose to overturn it; the plan's Q4 documents both paths.

**Verdict: self-review PASS with caveats.** An external adversary pass is still recommended before executable landing, per the template's adversary-verification expectation.

---

## Completion Record

*To be appended after work lands.*

---

## Relevant file paths

- **C source:** `/home/eilidh/src/smaug/src/marry.c` (78-362), `/home/eilidh/src/smaug/src/mud.h:2071-2072` (vnums).
- **New Go files:** `/home/eilidh/src/smaug/smaug-go/internal/act/marry.go`, `.../internal/act/marry_test.go`.
- **Modified Go files:** `/home/eilidh/src/smaug/smaug-go/internal/types/pcdata.go` (G1 — remove line 130), `.../internal/types/constants.go` (G4 — add vnums), `.../internal/persist/player.go` (G1 — SavePlayer writer at ~L542), `.../internal/persist/player_test.go` (G1 — round-trip test), `.../internal/boot/boot.go` (G4 — command registration), `.../internal/boot/boot_test.go` (G4 — registration tests).
- **Shipped data (conditional on Q2):** `/home/eilidh/src/smaug/db/area/Build.are` — add vnums 100/101 under `#OBJECTS`.
- **Reference plans:** `/home/eilidh/src/smaug/smaug-go/doc/plan-tranche-b.md`, `.../plan-tranche-c.md`, `.../plan-player-config.md` (Bio-field-add pattern — closest analog), `.../plan-phase6-arena.md` (Wave 1 sibling).
- **Existing seams cited:** `/home/eilidh/src/smaug/smaug-go/internal/handler/find.go:52` (`GetCharWorld`), `.../internal/handler/handler.go:105` (`CreateObject`), `.../internal/types/object.go:100` (`ExtraDescrData`), `.../internal/persist/player.go:284-285` (existing Spouse load), `.../internal/persist/player_test.go:929-931` (existing Spouse test), `.../internal/types/pcdata.go:130` (orphan `PCData.Spouse`), `.../internal/types/character.go:212` (canonical `CharData.Spouse`), `.../internal/types/object.go:62` (orphan `ObjData.ExtraDescr2`).
