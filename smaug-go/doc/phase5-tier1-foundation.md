# Phase 5 — Tier 1: Foundation & Correctness

## Goal

Fix the architectural voids and correctness bugs that cap the ceiling on every other tier. After Tier 1:

1. A central `Act()` dispatcher exists, so new combat/magic/damage-message code can emit per-recipient, visibility-aware messages with `$`-token expansion — the way SMAUG content expects.
2. A data-driven `spell_smaug` dispatcher exists, so skills.dat metadata can drive spell behavior. This alone unlocks most of the ~70 unimplemented spells without per-spell Go code.
3. The three missing saving throws (`SavesWands`, `SavesParaPetri`, `SavesBreath`) are present, so spells/attacks roll against the correct save instead of misusing `SavesSpellStaff`.
4. Two data-loss bugs in the player save path are fixed: learned skill proficiencies actually persist; on-hand Silver and Copper actually persist.
5. Skill-learning rate matches C's difficulty-scaled formula instead of the current flat `+20` approximation.

Scope note: per the plan, `Act()` is used for new and hot-path code (combat, magic, damage messages). Existing hand-formatted sites (emote, social dispatch, etc.) are not refactored in this tier.

## Gap inventory (verified)

| Item | C reference | Go status | Go reference |
|------|-------------|-----------|--------------|
| Central `act()` dispatcher with `$n/$N/$e/$m/$s/$p/$P/$t/$T` + `TO_CHAR/TO_VICT/TO_NOTVICT/TO_ROOM` routing | `src/smaug.c` `act()` + `act_string()` around line 3171–3663 | **Absent.** Grep `func Act(`, `TO_ROOM`, `TO_VICT`, `TO_NOTVICT` returns no dispatcher — only enum constants in `types/enums.go`. Commands hand-format. | Example hand-format: `act/comm.go:132` (`DoEmote` iterates room occupants). |
| `spell_smaug` data-driven dispatcher | `src/magic.c:7882` (reads `target` × `SPELL_ACTION` × `SPELL_CLASS` × `SPELL_FLAG` from skill record; routes to `spell_affect`, `spell_attack`, `spell_area_attack`, `spell_create_obj`, `spell_create_mob`) | **Absent.** No `SpellSmaug`, no `SpellAction`/`SpellClass`/`SpellDamageType`/`SpellFlag` fields on `SkillType`. Every spell is a hand-coded function in `spellRegistry`. | Registry: `magic/magic.go:18–49` (30 entries). Skill record: `types/skill.go` (has `Target`, `Flags`, `SaveType`, `Difficulty` but no action/class/damage-type enums). |
| `SavesWands` | `src/magic.c:1083` | **Absent.** | `magic/magic.go:77` has `SavesSpellStaff`, `:84` has `SavesPoisonDeath` only. |
| `SavesParaPetri` | `src/magic.c:1096` | **Absent.** | Same. |
| `SavesBreath` | `src/magic.c:1106` | **Absent.** | Same. |
| Learned skill/spell/weapon/tongue proficiency persistence | `src/save.c` `fwrite_char` writes each as `Skill <value> <name>` etc.; `fread_char` reads and stores into `ch->pcdata->learned[sn]`. | **Broken.** Load path discards with `_`; save path never writes. Effect: every player's skill proficiencies silently reset to 0 on every save/load cycle. | Discard site: `persist/player.go:277–279` (`case "Skill", "Spell", "Weapon", "Tongue": _ = sc.ReadNumber(); _ = sc.ReadString()`). Save site: `persist/player.go` SavePlayer (lines 440–515) emits no Skill/Spell/Weapon/Tongue lines. Backing store `PCData.Learned[gsn]` exists (`act/skills.go:19`). |
| Silver/Copper on-hand persistence | `src/save.c` writes Gold/Silver/Copper | **Broken.** Load reads Silver (player.go:131) and Copper (player.go:133); save writes only `Gold       %d` at player.go:449. Silent zero-out every save cycle. | `persist/player.go:449`. |
| Skill-learning formula matches C | `src/skills.c:1621–1687` `learn_from_success` uses `chance = learned + 5 * skill_table[sn]->difficulty`; awards `+2` when `percent >= chance`, `+1` when `chance - percent <= 25`; grants XP; fires "fully learned" message on adept cap. `learn_from_failure` uses same `chance`, awards `+1` only when `chance - percent <= 25`. | **Diverges.** Go uses flat `chance = learned + 20`; success always awards `+2`; failure always awards `+1`; no XP reward; no adept threshold. Hard skills improve as fast as easy ones. | `act/skills.go:23–54`. `Difficulty` field exists on `SkillType` (`types/skill.go:50`) but is never read. |

## Task groups

### G1 — `Act()` dispatcher (new)

Port `act()` + `act_string()` from `src/smaug.c`. Place in `smaug-go/internal/util/act.go` (pure formatter, no world deps) and re-export convenience wrappers from `game/act.go` for dispatching to descriptors.

**API (proposed):**
```go
// in util/act.go
func Act(format string, ch *types.CharData, arg1, arg2 any, to int) string
func ActToChar(format string, ch, victim *types.CharData, arg1, arg2 any) // TO_CHAR
func ActToVict(format string, ch, victim *types.CharData, arg1, arg2 any) // TO_VICT
func ActToRoom(format string, ch, victim *types.CharData, arg1, arg2 any) // TO_NOTVICT+TO_ROOM
```
`to` constants reuse `types.TO_CHAR/TO_VICT/TO_NOTVICT/TO_ROOM` from `types/enums.go`.

**Tokens to expand (verified in C at src/smaug.c act_string):**
- `$n` actor short (morph-aware); `$N` victim short
- `$e/$E` subjective pronoun; `$m/$M` objective; `$s/$S` possessive
- `$p` arg1-as-object short (visibility-aware); `$P` arg2-as-object short
- `$t` arg1-as-string raw; `$T` arg2-as-string raw
- `$d` door direction name derived from context
- Verb agreement helpers `$q/$Q` (C adds "s" to third-person verb)

**Acceptance:**
- `Act()` returns correctly expanded strings for each token × recipient relationship (self vs observer).
- Visibility gating: invisible `$n` shows as "someone" if observer can't see; invisible `$p` hides as "something" if observer can't see.
- `TO_NOTVICT` excludes actor and victim; `TO_ROOM` excludes actor only.
- New combat hits, spell effects, and damage messages (Tier 4) emit via `Act()`.

**TDD:**
1. Unit tests in `util/act_test.go` covering every token (table-driven: one row per token × recipient).
2. Unit tests for visibility cases (actor invisible to observer; object invisible; sleeping observer).
3. Integration check: one new combat hit message uses `Act()` and appears correctly to attacker, victim, and a bystander via descriptor buffers.

### G2 — `spell_smaug` dispatcher + metadata (new)

Port `spell_smaug` from `src/magic.c:7882`. Add the missing enum fields to `SkillType` and parse them from `skills.dat`.

**Sub-steps:**
1. Extend `types/skill.go` `SkillType` with:
   - `SpellAction int` (SA_CREATE, SA_DESTROY, SA_RESIST, SA_SUSCEPT, SA_DIVINATE, SA_OBSCURE, SA_CHANGE) — enum values in `types/enums.go`
   - `SpellClass int` (SC_NONE, SC_LUNAR, SC_SOLAR, SC_TRAVEL, SC_SUMMON, SC_LIFE, SC_DEATH, SC_ILLUSION, SC_ENHANCE)
   - `SpellDamageType int` (SD_NONE, SD_FIRE, SD_COLD, SD_ELECTRICITY, SD_ENERGY, SD_ACID, SD_POISON)
   - `SpellPower int` and additional `SpellFlag` bits beyond existing `Flags` (SF_PKSENSITIVE, SF_WATER, SF_EARTH, SF_AIR, SF_ASTRAL, SF_AREA, SF_DISTANT, SF_REVERSE, SF_NOFIGHT, SF_NODISPEL, SF_RECASTABLE, SF_NOSCRIBE, SF_NOBREW, SF_ACCUMULATIVE, SF_RECASTABLE, SF_GROUPSPELL, SF_OBJECT, SF_CHARACTER, SF_SECRETSKILL). Consult `src/mud.h` for the authoritative enum.

2. Extend `persist/skills.go` `LoadSkills` to parse the `SPELL_ACTION`, `SPELL_CLASS`, `SPELL_DAMAGE`, `SPELL_POWER`, `SPELL_FLAG` keywords.

3. Add `SpellSmaug` and the sub-dispatchers in a new `magic/spell_smaug.go`:
   - `SpellSmaug(w, sn, level, ch, vo)` — reads metadata, chooses sub-dispatcher
   - `spellAffect(skill, ch, victim, level)` — applies AffectData derived from `skill.Affects` and related fields
   - `spellAttack(skill, ch, victim, level)` — rolls damage from `skill.DiceFormula`, applies save per `skill.SaveType` and effect per `skill.SaveEffect`
   - `spellAreaAttack(skill, ch, level)` — iterates room; excludes SF_PKSENSITIVE-immune targets
   - `spellCreateObj(skill, ch, level)` — creates object of class `skill.Info` (C uses `skill->value`)
   - `spellCreateMob(skill, ch, level)` — creates mobile of vnum `skill.Info`

4. Register `"spell_smaug"` in `spellRegistry`.

**Acceptance:**
- `FindSpellFunc("spell_smaug")` returns a non-nil function.
- Loading real `db/system/en/skills.dat` does not drop spells that reference `spell_smaug`.
- A unit test casts a `spell_smaug`-backed affect spell (e.g., `faerie fire`) and verifies the affect lands; another casts an attack spell and verifies damage + save roll.
- No regression in the 30 existing hand-coded spells (they remain registered explicitly).

**TDD:**
- Tests live in `magic/spell_smaug_test.go`.
- Cover each sub-dispatcher with one positive + one "save succeeds" row.
- Fixture `skills.dat` snippets under `magic/testdata/` rather than loading the full production file.

### G3 — Missing saving throws

Port three functions next to the existing ones in `magic/magic.go`:

```go
// SavesWands — src/magic.c:1083
func SavesWands(level int, victim *types.CharData) bool { ... }
// SavesParaPetri — src/magic.c:1096
func SavesParaPetri(level int, victim *types.CharData) bool { ... }
// SavesBreath — src/magic.c:1106
func SavesBreath(level int, victim *types.CharData) bool { ... }
```

Each mirrors the existing pattern: `save := 50 + (victim.Level - level - victim.Saving<X>)*5; clamp 5..95`. The per-save `Saving*` fields already exist on `CharData` (they're loaded in `SavingThrows` at `persist/player.go:463`).

**Fix callers:** audit uses of `SavesSpellStaff` in `magic/magic.go`. Spells whose damage type is poison should call `SavesPoisonDeath`; breath attacks (Tier 4) should call `SavesBreath`; wand-delivered spells should call `SavesWands`. Document the current misuses in this doc as a punch-list when Tier 4 lands.

**TDD:** per-function unit tests in `magic/magic_test.go`.

### G4 — Skill/spell/weapon/tongue persistence (data-loss bug)

Fix `parsePlayerField` (`persist/player.go:277`) and `SavePlayer` (~line 510).

**Load:**
```go
case "Skill", "Spell", "Weapon", "Tongue":
    pct := sc.ReadNumber()
    name := sc.ReadString()
    gsn := worldSkillIndex(name) // resolve by name via World.Skills
    if gsn >= 0 && ch.PCData != nil {
        ch.PCData.Learned[gsn] = pct
    }
```
(`LoadPlayerWithWorld` already passes a world reference — use it; `LoadPlayer` (no world) can skip silently as today for compatibility.)

**Save:** after the existing `Affect` loop, iterate `ch.PCData.Learned`. For each non-zero entry, look up the skill by gsn and emit the correct keyword based on `skill.Type`:
- `SKILL_SPELL` → `Spell`
- `SKILL_SKILL` → `Skill`
- `SKILL_WEAPON` → `Weapon`
- `SKILL_TONGUE` → `Tongue`
- `SKILL_RACIAL` → `Skill` (C collapses this)

Format: `Skill         %d '%s'\n` matching C's emitter in `src/save.c`.

**Acceptance:**
- Round-trip test: create char, set three skills to 37/50/85, save, reload → values match.
- Mutation check: break save (skip the loop) → test fails; restore → test passes.

**TDD:** `persist/player_test.go` round-trip with skill proficiencies.

### G5 — Silver/Copper save bug

Add two lines to `SavePlayer` at `persist/player.go:449` immediately after the `Gold` write:

```go
fmt.Fprintf(w, "Silver     %d\n", ch.Silver)
fmt.Fprintf(w, "Copper     %d\n", ch.Copper)
```

Confirm load parses both (it does — `player.go:131` and `:133` cases `"Silver"` and `"Copper"`).

**TDD:** extend the player round-trip test to set non-zero Silver/Copper → verify preserved.

### G6 — Skill-learning formula

Change `learnFromSuccess` / `learnFromFailure` in `act/skills.go:23–54` to match C.

**Target code (matches `src/skills.c:1621–1687`):**

```go
func learnFromSuccess(ch *types.CharData, gsn int) {
    if ch.IsNPC() || ch.PCData == nil || gsn < 0 || gsn >= types.MAX_SKILL {
        return
    }
    skill := WorldRef.Skills[gsn]
    if skill == nil { return }
    learned := ch.PCData.Learned[gsn]
    adept := skill.SkillAdept[ch.Class]
    if learned >= adept { return }
    chance := learned + 5*skill.Difficulty
    roll := util.NumberPercent()
    gain := 0
    switch {
    case roll >= chance:
        gain = 2
    case chance-roll <= 25:
        gain = 1
    }
    if gain > 0 {
        ch.PCData.Learned[gsn] = util.UMIN(learned+gain, adept)
        ch.Sendf("You have become better at %s! (%d%%)\n\r",
            skill.Name, ch.PCData.Learned[gsn])
        // Tier 1 note: XP award on skill gain deferred;
        // tracked in Tier 4 when damage messages / XP plumbing get a broader pass.
    }
}

func learnFromFailure(ch *types.CharData, gsn int) {
    if ch.IsNPC() || ch.PCData == nil || gsn < 0 || gsn >= types.MAX_SKILL {
        return
    }
    skill := WorldRef.Skills[gsn]
    if skill == nil { return }
    learned := ch.PCData.Learned[gsn]
    adept := skill.SkillAdept[ch.Class]
    if learned >= adept-1 { return }
    chance := learned + 5*skill.Difficulty
    roll := util.NumberPercent()
    if chance-roll <= 25 && roll < chance {
        ch.PCData.Learned[gsn] = util.UMIN(learned+1, adept-1)
    }
}
```

**Balance call-out:** this meaningfully slows skill gain for high-difficulty skills. If `skills.dat` does not populate `Difficulty` for every entry, zero-difficulty skills will never improve (chance = learned + 0 = learned, so `roll >= chance` is rare). Either populate the field from C's `skill_table` equivalent or floor difficulty at 1 during load.

**TDD:**
- Test with `skill.Difficulty = 5`, `learned = 10`: on `NumberPercent()` returning 99, expect `gain = 2`; on 30 (chance=35), expect `gain = 0`; on 20 (chance-roll=15), expect `gain = 0` in success path but `gain = 1` in failure path at the right percent.
- Use a deterministic RNG fixture.

## Critical files

**Create:**
- `smaug-go/internal/util/act.go` — `Act()` formatter.
- `smaug-go/internal/util/act_test.go`.
- `smaug-go/internal/magic/spell_smaug.go` — dispatcher + sub-dispatchers.
- `smaug-go/internal/magic/spell_smaug_test.go`.
- `smaug-go/internal/magic/testdata/skills_smaug_fixture.dat`.

**Modify:**
- `smaug-go/internal/types/skill.go` — add `SpellAction`, `SpellClass`, `SpellDamageType`, `SpellPower` fields.
- `smaug-go/internal/types/enums.go` — add SA_*, SC_*, SD_*, expanded SF_* constants.
- `smaug-go/internal/persist/skills.go` — parse the new fields.
- `smaug-go/internal/magic/magic.go` — register `spell_smaug`; add `SavesWands`, `SavesParaPetri`, `SavesBreath`.
- `smaug-go/internal/persist/player.go` — `parsePlayerField` Skill/Spell/Weapon/Tongue cases (wire to `Learned`); `SavePlayer` emits Silver, Copper, and learned-skill lines.
- `smaug-go/internal/act/skills.go` — `learnFromSuccess` / `learnFromFailure`.
- `smaug-go/internal/act/skills_test.go` — update existing tests to reflect formula change; add difficulty table.
- `smaug-go/internal/persist/player_test.go` — round-trip skills, Silver/Copper.

**Reference (C):**
- `src/smaug.c` `act()` (~3404) and `act_string()` (~3171).
- `src/magic.c` `spell_smaug` (~7882); saves at 1073/1083/1096/1106/1116.
- `src/skills.c` `learn_from_success` (~1621) and `learn_from_failure` (~1675).
- `src/save.c` `fwrite_char` / `fread_char` for player file format truth.

## Reused utilities

- `util.OneArgument`, `util.SmashTilde`, `util.UMIN/UMAX/URANGE`, `util.NumberPercent`, `util.DiceRoll` already exist.
- `types.BitVector` for the expanded `SpellFlag` bits.
- `types.PCData.Learned` already backs skill proficiency.
- `world.World.Skills` gives gsn→name→skill lookup; a helper `worldSkillIndex(name)` can live in `persist/skills.go` next to existing loaders.
- Existing saving-throw pattern at `magic/magic.go:77–88` is the template for three new saves.

## Verification

1. `cd smaug-go && go test ./...` passes. Expected regressions to fix as part of the tier: any existing skill-gain tests that assumed the old `+20` formula.
2. New tests: `util/act_test.go`, `magic/spell_smaug_test.go`, player save round-trip for skills/silver/copper.
3. Manual smoke test:
   - Build and boot: `go build -o smaug-go ./cmd/smaug/ && ./smaug-go -port 4000 -data ../db`
   - Connect via telnet, create character, practice a skill (if trainer available) or use a skill successfully several times → see proficiency increment.
   - Save/quit → reload → `skills` or `practice` reports preserved proficiencies.
   - Carry Silver/Copper (if shops or reset grant them) → save/quit → reload → coinage preserved.
4. Spot-check: pick one spell that uses `spell_smaug` in real `skills.dat` and confirm it now casts without an "that spell hasn't been implemented" error.

## Open questions / follow-ups

- **XP-on-skill-gain** is deferred from G6 to Tier 4's messaging pass; note it on the Tier 4 punch-list.
- **`Act()` migration of existing sites** is explicitly out of scope for this tier. When Tier 4 damage messages land, expand Act() adoption to the combat emission paths.
- **`SpellAction/Class/Damage` bitvector vs int**: choose one representation up front. C uses int enums; keep int for interop.
