# Post-Phase 6 Vision — Modernizations

**Status:** Capture only (2026-04-19). Not scheduled. These are user-stated directions for work that starts *after* Phase 6 completes. Each item is a major overhaul — schema rewrites, not ports — and should not be attempted mid-phase.

This doc is a faithful record of intent. It is the seed for future planning; the full design work (open questions, migration, UX, commands) happens when each item is scheduled.

---

## 1. Gender / Sex — SOGI framework

The SMAUG C baseline (and the current Go port) uses a 3-value `Sex int` — `SEX_NEUTRAL` / `SEX_MALE` / `SEX_FEMALE` — on `CharData`, `PCData`, `MobIndexData`, `ClanData`, etc. (`internal/types/enums.go:248-250`; fields at `character.go:82`, `mob_index.go:23`, `clan.go:99`). Pronoun resolution across the port (damage messages, `util.Act`, spell targeting strings, flee/drop messages, `tagline`) reads this single enum.

**Target model (SOGI):** separate the four dimensions — **S**exuality, **O**rientation, **G**ender, **I**dentity — and treat presentation as orthogonal to identity.

**Principles (verbatim from user direction):**
- A player can **present** any number of ways without requiring Identity alignment.
- The player is in control of their **pronouns** and **gender label**.
- Many common choices are available AND the player can specify their own.
- Custom pronouns and gender labels can be added to the community list; other players can choose the same label.
- The UX must never feel like a (Male/Female/Other) choice, or a three-gender choice.
- Players can change pronouns and gender/identity labels as often as they see fit.
- The player can choose to only share their **pronouns** in `stats` — gender and identity labels may be hidden and revealed through conversation/interaction.
- The player chooses how and when to share **Orientation**. The game does not need to track Orientation data server-side (no persisted field). The player may set their own Orientation label for display purposes only.

**Scope signals for future planning (not design decisions yet):**
- Schema: replace `Sex int` on `CharData`/`PCData` with a richer struct (e.g. `Pronouns` set, `GenderLabel` string, `IdentityLabel` optional, `PresentationTags`). Mobs keep `Sex` as the simple pronoun-resolution source (NPCs don't have agency).
- Community list: a persisted global table of custom gender/pronoun labels, first-write-wins, other players can reuse.
- Privacy: per-field visibility flags (`hidden` / `stats-only` / `public`). Gender and identity labels default hidden; pronouns default stats-only.
- `util.Act` pronoun resolution: reads Pronouns set, not `Sex`, when formatting `$e`/`$m`/`$s`/`$N` etc. Fallback rules for unknown/neutral speakers.
- Commands: new `pronouns` / `gender` / `identity` / `present` / `reveal` / `orientation` commands. Old `sex` setter becomes compatibility shim or disappears.
- Migration: existing pfiles carry `Sex int` — backfill to pronoun set on load, keep writing the new schema.
- Clan / mob data keeps simple `Sex int` (no agency there); clan officer commands may need to re-examine pronoun usage.

**Blast radius estimate:** every file that reads `Sex` directly. Grep shows `character.go`, `pcdata.go`, `mob_index.go`, `clan.go`, plus all callers of `util.Act` pronoun tokens (which is most `act/` files). Dozens of test fixtures. This is a multi-phase effort.

---

## 2. Marriage — poly-capable with per-coupling agency

Supersedes the deferred `plan-phase6-marriage.md` (parked 2026-04-19 on the vnum-100/101 collision question). When marriage is reintroduced, it is a **rewrite**, not a port of the C two-person `Spouse` field.

**Principles (verbatim from user direction):**
- A player can marry and divorce whoever they like.
- The two players getting married are the only two who must consent to that marriage.
- A player can divorce or end any marriage they are involved in.
- Groups cannot decide the fate of a marriage, even if they are also marriage partners.
- Worked example:
  - Bob and Sally are married.
  - Carla and Sally decide to get married.
  - Bob cannot object (except to divorce Sally).
  - Only Carla and Sally decide how long the Carla-Sally marriage continues.
- No restrictions on who can marry whom, or how many marriages a player participates in.
- A group of marriages can be collectively labeled (triad, thruple, polycule, poly family, etc.). All members of the group must agree to the label. Each individual marriage coupling in the group retains its own agency.

**Scope signals:**
- Schema: `Spouse string` (`character.go:212`, `pcdata.go:130`) becomes a slice/set of `MarriageData` — each entry is a distinct two-person contract. `Spouse` field goes away (or becomes a legacy pfile-load backfill).
- Commands: `marry <player>` requires two-party consent handshake (challenge/accept style, similar to arena `challenge`/`accept`). `divorce <player>` is unilateral against a specific partner. `marriages` lists active contracts.
- Group labels: separate data type (`PolyGroupData` or similar). Group `join` / `leave` / `label` commands. Group labels are metadata only — they do not constrain the individual contracts.
- Rings (if kept): minted per-marriage, not per-player. Resolves the old C `do_rings` plural-name-singular-effect anomaly organically.
- Persistence: per-marriage records in a new `db/marriages/` directory (or embedded in pfiles on both sides). Needs a reconciliation pass on load for divergent pfiles.
- Visibility: a marriage's existence may be private, stats-visible, or public at each participant's choosing (intersects with the SOGI privacy model above).

---

## 3. Race → Heritage + Community (Daggerheart-style)

Existing `Race int` enum at `enums.go:226-242` has 18 entries (HUMAN, ELF, DWARF, HALFLING, PIXIE, HALF_OGRE, HALF_ORC, HALF_TROLL, HALF_ELF, GITH, DROW, SEA_ELF, VAMPIRE, DEMON, LIZARDMAN, GNOME, ANGEL, + `RACE_DRAGON = 31` at `constants.go:367`). `Race` field appears on `CharData` (`character.go:84,337`), `MobIndexData` (`:62`), `ClanData` (`:96`), plus clan race-specific kill/flee/die counters.

**Principles (verbatim from user direction):**
- In the interest of avoiding stereotypes, renovate Race similar to Daggerheart.
- Daggerheart uses **Heritage** and **Community** instead of Race.
  - **Heritage** defines where you came from — your genetic material.
  - **Community** defines who influenced you as you grew up, and who influences you now.

**Scope signals:**
- Schema: `Race int` on `CharData`/`PCData` splits into `Heritage int` (or richer) + `Community` (possibly a slice — multiple influences). Stat baselines, affect maps, language starters, and ability modifiers currently keyed off `Race` re-key onto `Heritage`. Community layers cultural bonuses/skills on top.
- Content: `db/race.dat` (if present) reshapes. Mob prototypes keep a `Heritage int` for pronoun/pronoun-like resolution but probably don't need Community (NPCs' "community" is their area).
- Commands: creation flow teaches both choices. `stats` shows both. `setrace` dies or becomes a migration tool.
- Clan race-targeted counters (`KillNpcRace`, `FleeNpcRace`, `DieNpcRace`) re-key to Heritage or are replaced with a broader tag system.
- Migration: existing players on disk get a Heritage backfill from their old `Race`; Community defaults to "Unknown" or a prompt on next login.
- Mob racial attacks / innate abilities re-anchor to Heritage — most current C logic is already heritage-shaped, just labeled "race."

**Blast radius estimate:** comparable to SOGI — every `Race` read in combat, spells, skills, racial stats, language tables, clan counters. Dozens of fixtures.

---

## Cross-cutting considerations

- **All three items require pfile migration.** Schedule a pfile version bump + a one-time migration pass before the first of these lands.
- **All three intersect.** Heritage + Community may inform how pronouns default; marriage contracts may display heritage tags; gender presentation is visible on `stats` alongside heritage.
- **Order is not pre-committed.** Gender/SOGI is the broadest refactor but the lowest-risk (most of the work is in `util.Act` and command surface). Race → Heritage/Community touches combat balance and is the highest-risk. Marriage is the smallest but depends on the consent/contract machinery that doesn't exist yet.
- **Non-goal:** these are not gameplay rebalances. Keep combat numbers stable through the refactors; the naming and data model change, not the damage calculation.

---

## What this document is NOT

- Not a plan. No task groups, no acceptance criteria, no mutation-verify gates.
- Not a commitment to any timeline.
- Not a schema proposal. The field shapes named above are examples for scope-signaling, not decisions.
- Not a complete list of affected files. A real plan starts with a grep audit of every `Sex` / `Race` / `Spouse` read site.

When any of these are scheduled, the first step is to write a dedicated `plan-postphase6-<topic>.md` that turns the principles here into open questions, design choices, and task groups — in the same shape as the Phase 6 plan docs.
