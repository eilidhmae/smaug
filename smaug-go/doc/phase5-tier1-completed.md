# Phase 5 Tier 1 — Completed Work

## Summary

Phase 5 Tier 1 is complete. All 6 task groups (G1–G6) landed. 81 source files, 64 test files, 1,445 top-level test cases across 13 packages — all passing via `go test -count=1 ./...`.

Process note: six implementer subagents ran mostly in parallel, each followed by an adversary audit. One adversary disagreement triggered a quorum audit on G2 (spell_smaug bit layout) — quorum confirmed the adversary, a rework subagent applied the fix, and a final adversary pass verified the rework.

---

## G1: Act() Dispatcher ✓

Pure formatter + router ported from `src/smaug.c:3171` (`act_string`) and `src/smaug.c:3404` (`act`).

**New files:**
- `smaug-go/internal/util/act.go` — pure formatter, no world deps
- `smaug-go/internal/util/act_test.go` — 32 tests (29 initial + 3 wiz-invis regression guards)

**API:**
```go
func ActFormat(format string, to, ch, vch *types.CharData, arg1, arg2 any) string
func Act(format string, ch, vch *types.CharData, arg1, arg2 any, to int)
```

**Tokens implemented:** `$n` `$N` `$e` `$E` `$m` `$M` `$s` `$S` `$p` `$P` `$t` `$T` `$d` `$q` `$Q` `$$`. NPC ShortDescr fallback; pronouns by Sex; object short with `IndexData` fallback.

**Routing:** `TO_CHAR` / `TO_VICT` / `TO_NOTVICT` / `TO_ROOM` (`TO_CANSEE` currently routes like `TO_ROOM`). Iterates `ch.InRoom.People`. `CharData.Send` is a no-op on nil Desc, so NPCs stay in the loop — ready for Tier 3's mudprog ACT_PROG triggers.

**Visibility:** `actCanSee` checks `PLR_HOLYLIGHT` bypass, `PLR_WIZINVIS` vs observer Trust, `AFF_INVISIBLE`/`AFF_HIDE` with `AFF_DETECT_*` overrides. `actCanSeeObj` handles `ITEM_INVIS`. Self always sees self.

**Known scope-limited divergences from C** (deferred to Tier 4):
- No `DONT_UPPER` flag parameter — no callers yet
- `ACT_SECRETIVE` NPC flag not honored
- `IS_AWAKE` gate on recipients not applied
- `ITEM_PAPER` short-descr override absent

---

## G2: spell_smaug Data-Driven Dispatcher ✓

Ported from `src/magic.c:7882`, using the **non-NEWSPELLS** bit layout (the layout shipped with the `db/system/en/skills.dat` file in this repo). Initial port used NEWSPELLS shifts; adversary + quorum audits confirmed the bit-layout error and a rework applied the fix.

**New files:**
- `smaug-go/internal/magic/spell_smaug.go` — dispatcher, 5 sub-dispatchers, save-dispatch, `parseDiceExpr`
- `smaug-go/internal/magic/spell_smaug_test.go` — 21 tests

**Modified:**
- `types/skill.go` — accessor methods unpacking `Info` (non-NEWSPELLS layout):
  - `SpellDamageType() = Info & 7`
  - `SpellAction() = (Info >> 3) & 7`
  - `SpellClass() = (Info >> 6) & 7`
  - `SpellPower() = (Info >> 9) & 3`
  - `SpellSave() = (Info >> 11) & 7`
  - Added `Value int` field for create-obj/create-mob vnum (C: `skill->value`).
- `persist/skills.go` — `Value` keyword now populates `skill.Value` (was read-and-discard).
- `magic/magic.go` — registers `"spell_smaug": SpellSmaug` in `spellRegistry`.

**Dispatcher structure** (mirrors `src/magic.c:7895`):
- Top-level `switch skill.Target` — `TAR_IGNORE`, `TAR_CHAR_OFFENSIVE`, `TAR_CHAR_DEFENSIVE`/`TAR_CHAR_SELF`, `TAR_OBJ_INV` (stub), default
- Inside `TAR_IGNORE`: `SF_OBJECT` → create obj, `SF_CHARACTER` → create mob, otherwise self-affect
- Inside `TAR_CHAR_OFFENSIVE`: `SA_DESTROY` → attack, else affect
- `SF_AREA` shortcut at the top for area spells

**Save effects:** `SE_NEGATE`, `SE_EIGHTHDAM`, `SE_QUARTERDAM`, `SE_HALFDAM`, `SE_3QTRDAM`, `SE_REFLECT` (swaps roles to damage the caster), `SE_ABSORB` (victim heals).

**SF_PKSENSITIVE:** halves effective save-level when both caster and victim are PCs (per `src/magic.c:2343`), does not skip the victim.

**parseDiceExpr** — recursive-descent over tokens NUMBER, `l`/`level`/`i`, `+`, `-`, `*`, `/`, `(`, `)`, `d`. Covers every expression form found across the 169 `Affect` lines in `db/system/en/skills.dat`: plain signed int, `l`, `l*N`, `(l*N)+M`, `l/N`, `NdM`, `NdM+(l-K)`, `1d8+(l/3)`, etc. Bitvector-name modifiers return 0 (handled via `SmaugAff.BitVector`).

**Known divergences** (noted in comments):
- `SE_REFLECT` uses `combat.Damage` directly instead of recursing through `spellAttack` — intentional anti-recursion protection; C gives the reflected spell a second save chain
- `spellAreaAttack` does not yet honor `ROOM_SAFE` or PKILL guards — Tier 2 scope (room-flag honoring)
- `TAR_OBJ_INV` path is a friendly-message stub — Tier 4 scope

---

## G3: Missing Saving Throws ✓

Three functions added to `smaug-go/internal/magic/magic.go` alongside existing `SavesSpellStaff` / `SavesPoisonDeath`:

```go
func SavesWands(level int, victim *types.CharData) bool
func SavesParaPetri(level int, victim *types.CharData) bool
func SavesBreath(level int, victim *types.CharData) bool
```

Same `save := 50 + (victim.Level - level - victim.Saving*) * 5; URANGE(5,95)` formula. `SavesWands` auto-passes for victims with `Immune & RIS_MAGIC` (C: `src/magic.c:1087`).

**Tests:** 11 cases in `magic_test.go` — per-function statistical, clamp-range, field-identity, and RIS_MAGIC immune.

---

## G4: Learned Proficiency Persistence ✓

Fixed the silent data-loss bug where every `Skill`/`Spell`/`Weapon`/`Tongue` line in player save files was discarded on load and never emitted on save.

**Modified** `smaug-go/internal/persist/player.go`:
- Added package-level hooks: `SkillNameLookup func(name string) int` and `SkillGetter func(gsn int) *types.SkillType`. Callers set them once at boot.
- Load: `Skill`/`Spell`/`Weapon`/`Tongue` case in `parsePlayerField` now reads the number + single-quoted name, resolves gsn via `SkillNameLookup`, stores to `ch.PCData.Learned[gsn]`.
- Save: after the existing affect loop, iterate `p.Learned`; for each non-zero entry, `SkillGetter` the skill and emit the keyword mapped by `skill.Type` (`SKILL_SPELL` → `Spell`, `SKILL_WEAPON` → `Weapon`, `SKILL_TONGUE` → `Tongue`, everything else → `Skill`, matching C's `fwrite_char` default case).

**Boot wiring** in `cmd/smaug/main.go` sets the two hooks after `LoadSkills`.

**Tests:** `TestSaveLoadSkills` round-trips four skills covering all `SKILL_*` types; `defer` restores the hook state so parallel tests aren't contaminated.

---

## G5: Silver/Copper Save ✓

Two `fmt.Fprintf` lines added after the existing `Gold` write in `SavePlayer`:
```go
fmt.Fprintf(w, "Silver     %d\n", ch.Silver)
fmt.Fprintf(w, "Copper     %d\n", ch.Copper)
```
Load side was already correct. `TestSaveLoadSilverCopper` round-trips non-zero values; mutation test (delete the Silver line) verified test failure.

---

## G6: Skill-Learning Formula ✓

Replaced flat `+20` formula in `act/skills.go` `learnFromSuccess` / `learnFromFailure` with C-parity logic (`src/skills.c:1621–1687`).

**Key behaviors:**
- `chance = learned + 5*skill.Difficulty` (was flat `+20`)
- Per-class adept cap via `skill.SkillAdept[ch.Class]` (was flat `100`)
- Guard: both functions early-return when `ch.PCData.Learned[gsn] <= 0` (C parity; prevents 0-proficiency players from improving by use alone)
- Success: `gain=2` if `roll >= chance`, `gain=1` if `chance-roll <= 25`, else no gain
- Failure: `gain=1` iff `chance - roll <= 25` (no `roll < chance` conjunct — an earlier port mistakenly added it, causing legitimate gains to be suppressed)
- Clamps: success at `adept`, failure at `adept-1`

**Test seam:** `var numberPercent = util.NumberPercent` at package level lets tests swap the RNG deterministically.

**XP-on-gain and "fully learned" messages** intentionally deferred to Tier 4 (noted in code comments).

---

## Test-Suite State

| Package | Result |
|---------|--------|
| cmd/smaug | PASS (22.7s — integration) |
| internal/act | PASS |
| internal/combat | PASS |
| internal/command | PASS |
| internal/game | PASS |
| internal/handler | PASS |
| internal/magic | PASS |
| internal/mudprog | PASS |
| internal/net | PASS |
| internal/persist | PASS |
| internal/types | PASS |
| internal/util | PASS |
| internal/world | PASS |

Total: 81 source files, 64 test files, 1,445 top-level test cases.

---

## Deferred to Later Tiers

**Tier 2 (flag honoring):**
- `spellAreaAttack` `ROOM_SAFE` + PKILL guards
- `ACT_SECRETIVE`, `IS_AWAKE`, `ITEM_PAPER` checks in `Act()`

**Tier 4 (content breadth):**
- `DONT_UPPER` / flags parameter on `Act()`
- `TAR_OBJ_INV` branch in `SpellSmaug`
- XP-on-skill-gain plumbing and "fully learned" message
- Migrating the 30 hand-coded spells onto `SpellSmaug`
- Full `SE_REFLECT` save chain recursion
- Inherited tech debt: C's `chance()` luck/mental-state/deity-favor modifiers on all save functions

**Foundation now in place for Tier 2** (room-flag enforcement + idle-data wiring). `Act()`, `spell_smaug`, `SavesWands`/`ParaPetri`/`Breath`, and correct skill learning are available to downstream tiers.
