# Plan — Combat Depth (multi-attack, weapon proficiency, stance)

Source audit: `audit-2026-04-17.md` findings C1/C2 (P0). TODO refs: `TODO.md:6`, `17–28`, `109`.

## Goal

Restore three tightly-coupled combat features so the game stops feeling shallow at higher levels:

1. **PC multi-attack cascade** — `second_attack` through `seventh_attack` skills, matching `fight.c:1071-1141`. A level-30 PC warrior currently gets 1 attack/round; C gives 3–5.
2. **Weapon proficiency bonus** — `weapon_prof_bonus_check` (non-`ENABLE_WEAPONPROF` path), matching `fight.c:1256-1317`. Applies to `victim_ac` before the hit roll and to `dam` on hit.
3. **Stance application** — `ch.Stance` is stored/saved today but unused in combat. Port the GM bonus-attack loop and the per-stance `dam_done`/`dam_taken` multipliers, matching `fight.c:1058-1069, 2549-2594`.

All three share the same outer loop (C's `multi_hit`), the same `OneHit` call path, and the same skill/learn helpers, so they're naturally batched into one sub-phase.

## Gap inventory (verified)

| Feature | C ref | Go state | Gap |
|---|---|---|---|
| `multi_hit` outer loop | `fight.c:973-1174` | Inlined in `combat.go:107-122` ViolenceUpdate: only NPC `NumAttacks` loop + unconditional dual-wield extra hit | No dedicated `MultiHit`; no retcode propagation; no short-circuit on death |
| Berserk extra hit | `fight.c:1001-1008` (`LEARNED(berserk)*6/2`) | `AFF_BERSERK` set in `skills3.go:314-330`, never consumed | Missing |
| Dual-wield learned roll | `fight.c:1010-1026` | Unconditional extra OneHit (`combat.go:120-122`) | Missing learned gate, `dual_bonus` calc, learn-on-success/failure, `Move<10 → -20` penalty |
| Dual-wield weapon alternation | `fight.c:1402-1411` (`static bool dual_flip`) | `OneHit` always picks `WEAR_WIELD` (`combat.go:162`) | Offhand-weapon bonus hit currently swings primary stats (bug) |
| NPC stance attacks | `fight.c:1033-1052` | `NumAttacks` loop exists for NPCs | Missing `+= stance_index[stance].num_attacks` |
| PC stance GM bonus loop | `fight.c:1055-1069` | `Stance`, `Stances[]`, `STANCE_GRAND_MASTER` all in types; `stance_index` unloaded (`db/system/stances.dat` is a 1-line stub) | Fully missing |
| second…seventh attack cascade | `fight.c:1071-1141` | Absent | Fully missing — biggest single gap |
| Post-cascade NPC half-level bonus hit | `fight.c:1145-1147` | Absent | Missing |
| `ACT_NOATTACK` gate | `fight.c:991-992` | Constant defined (`constants.go:467`), not checked | Missing |
| `TIMER_RECENTFIGHT` on mutual-PC combat | `fight.c:980-986` | Constant defined (`enums.go:899`), not used | Missing |
| `weapon_prof_bonus_check` | `fight.c:1256-1317` (non-ENABLE path — `Value[3]` DAM_*) | No prof-skill lookup anywhere in `combat/`; prof skill names exist in `db/system/en/skills.dat:5233-5294` | Fully missing |
| Prof bonus → `victim_ac` and `dam` | `fight.c:1533, 1566-1567` | `OneHit:176,189-223` has no prof path | Missing |
| Stance `dam_done` / `dam_taken` | `fight.c:2549-2594` | Absent | Missing |

Foundations ready:

- `types/character.go:102` — `NumAttacks int`
- `types/character.go:204-205` — `Stance int`, `Stances [MAX_STANCE]int`
- `types/pcdata.go:97-98` — Stances duplicated (persisted — confirmed round-trip)
- `types/constants.go:782` — `STANCE_GRAND_MASTER = 200`
- `types/enums.go:19-33` — `STANCE_*` enum and `MAX_STANCE`
- `types/enums.go:1105-1123` — `DAM_HIT`..`DAM_PEA` (for prof switch)
- `combat/combat.go:87-143` `ViolenceUpdate`, `:156` `OneHit(w, ch, victim, dt)` — insertion points
- `util.NumberPercent`, `UMAX`, `UMIN` — RNG helpers
- Existing skill helpers live in `act/skills.go` (`canUseSkill`, `learnFromSuccess`, `learnFromFailure`) plus `act/consume.go` (`lookupSkillSlot`) — **not reachable from `combat/` without a cycle**; see G2.
- `is_attack_supressed(ch)` (C `fight.c:74-92, 988`) — checks `TIMER_ASUPRESSED`. No Go analog; silently false today. Include as a G5 gate.

`AFF_HASTE`/`AFF_SLOW` are not defined in Go types today and C `multi_hit` does not consume them — non-issue for this port.

## Task groups

### G1 — Extract `MultiHit`, return retcode from `OneHit` (S, prereq)

Change `OneHit` signature to `func OneHit(w, ch, victim, dt) int` returning `rNONE`/`rVICT_DIED` etc. Extract `MultiHit(w, ch, victim, dt) int` mirroring C `multi_hit`. Move NPC `NumAttacks` loop, dual-wield extra hit, and wimpy-flee trigger into `MultiHit`. `ViolenceUpdate` just calls `MultiHit` per fighter and keeps post-round wimpy logic.

- **Files:** `combat/combat.go`, `combat/combat_test.go`, callers in `act/skills3.go:91-94,295`.
- **Caller note:** `act/skills3.go:91-94` is `DoCircle` which today fires two `OneHit`s unconditionally for "multi" effect. After G1, that pattern is semantically wrong (MultiHit owns cascading). Review `DoCircle` as part of G1 — keep its two-hit behavior, but document the divergence. Do NOT just widen signature without understanding intent.
- **Tests:** retcode unit tests (hit/miss/kill); existing 5 OneHit tests remain green; new `TestMultiHit_ShortCircuitsOnDeath`.

### G2 — Expose skill-check helpers to `combat/` (S, prereq)

`canUseSkill` / `learnFromSuccess` / `learnFromFailure` / `lookupSkillSlot` live in `act`; `combat` cannot import `act`. Use the **existing hook pattern** (`combat.HitprcntHook`, `VoidHook`, `ObjDamageHook`, `RfightHook`, `DeathRoomHook` — 5 combat-package hooks currently wired in `boot/boot.go`). Add:

```go
// combat/hooks.go
var (
    CanUseSkillHook      func(ch *types.CharData, percent int, gsn int) bool
    LearnFromSuccessHook func(ch *types.CharData, gsn int)
    LearnFromFailureHook func(ch *types.CharData, gsn int)
    LookupSkillSlotHook  func(name string) int
)
```

Wire from `boot/boot.go` (same pattern as existing combat hooks). Lower disruption than extracting a package.

**GSN name-to-int resolution:** to avoid a per-attack map lookup in the hot path, resolve gsn names once at boot (`gsnSecondAttack`, `gsnThirdAttack`, etc.) into package-level ints in `combat/`. `LookupSkillSlotHook` is only called at init; `canUseSkill` etc. take the resolved int.

- **Tests:** existing `act/skills_test.go` green; add `combat/skillcheck_test.go` covering nil-hook fallthrough; add `combat/skillcheck_init_test.go` asserting all gsn names resolve at boot.

### G3 — PC multi-attack cascade (M)

In `MultiHit`, after NPC short-circuit and before return, port the 6-tier cascade from `fight.c:1071-1141`. For each tier: look up gsn, compute chance, check first-hit-still-alive short-circuit, fire `OneHit`, propagate retcode, learn on success/failure. Thread `dual_bonus` from G4.

- **Files:** `combat/combat.go`, `combat/combat_test.go`.
- **Tests (TDD, write-red-first, RNG-stubbed for determinism):**
  - `TestMultiHit_SecondAttackFires_Learned150` (Learned=150, stub `NumberPercent` returning 50 → second fires; assert exactly 2 OneHit calls via spy).
  - `TestMultiHit_ThirdAttack_Learned0_DoesNotFire` (stub returns 50, Learned=0 → exactly 1 OneHit call).
  - `TestMultiHit_SecondAttack_Learned75_Hit` (stub returns 40, Learned=75 → fires; stub returns 80 → doesn't fire — boundary case).
  - `TestMultiHit_NPCUsesLevel` (NPC, no Learned, level drives chance).
  - `TestMultiHit_LearnFromSuccessFires` (spy via hook).
  - `TestMultiHit_CascadeShortCircuitsOnVictimDeath`.
  - `TestMultiHit_BackstabSkipsCascade` (`dt == gsn_backstab` → single hit).
  - No loose statistical tests (±10% is wide enough to pass a broken implementation). Use the stub pattern throughout.

### G4 — Dual-wield learned roll + `dual_bonus` threading (S)

Port `fight.c:1010-1029`. Learned-check on `gsn_dual_wield`; on success compute `dual_bonus`, fire extra `OneHit` with offhand weapon; low-move penalty (`Move<10 → dual_bonus = -20`). Thread `dual_bonus` into G3 tier chances.

- **Files:** `combat/combat.go`, `combat/combat_test.go`.
- **Tests:** existing `TestViolenceUpdate_DualWield` stays green; add `TestMultiHit_DualWield_LearnedGate`, `TestMultiHit_LowMovePenalty`.

### G5 — Berserk / NOATTACK / PKill-timer / attack-suppress gates (S)

Front-of-`MultiHit` gates from `fight.c:980-1008`:

- `ACT_NOATTACK` (mob) → early return.
- `is_attack_supressed(ch)` (C `fight.c:74-92, 988`) — checks `TIMER_ASUPRESSED`. Port a minimal `IsAttackSuppressed(ch) bool` in `handler/` reading `ch.Timers` if present; return false when timer subsystem is incomplete (flag TODO).
- `TIMER_RECENTFIGHT` add on mutual PC combat (defer if `handler.AddTimer` missing — flag follow-up).
- `AFF_BERSERK` → extra hit at `gsn_berserk*6/2`% chance.

- **Tests:** `TestMultiHit_BerserkExtraHit`, `TestMultiHit_NoAttackMobSkips`, `TestMultiHit_AttackSuppressedSkipsCascade`, `TestMultiHit_PKillTimer` (conditional on handler.AddTimer).

### G6 — Weapon proficiency bonus (M)

Port the **non-`ENABLE_WEAPONPROF`** branch (`fight.c:1256-1317`) as `WeaponProfBonusCheck(ch, wield) (bonus, profGsn int)`. Switch on `wield.Value[3]` (DAM_*) → `pugilism` / `long_blades` / `short_blades` / `flexible_arms` / `talonous_arms` / `bludgeons` / `missile_weapons`. Bonus = `(Learned - 50) / 10`. PC-only, level > 5. Apply in `OneHit`: `victimAC += bonus` (before diceroll), on miss `LearnFromFailureHook(ch, profGsn)`, on hit `dam += bonus/4`.

- **Files:** new `combat/profbonus.go` (or inline), `combat/combat_test.go`.
- **Tests:**
  - `TestWeaponProfBonus_Slashing` (Learned=100 long_blades, DAM_SLASH weapon → bonus=5).
  - `TestWeaponProfBonus_BelowLevel5_NoBonus`.
  - `TestWeaponProfBonus_NPC_NoBonus`.
  - `TestWeaponProfBonus_NoWield_NoBonus`.
  - `TestOneHit_ProfBonusLowersVictimAC`.
  - `TestOneHit_ProfBonusDamageAdd`.
  - `TestOneHit_ProfMissCallsLearnFailure`.

### G7 — Stance application in combat (M)

Five pieces:

1. Introduce a minimal `stance_index` in Go. `db/system/stances.dat` is empty; hard-code SMAUG defaults (from `src/stances.c`) into `combat/stance_index.go` with fields `NumAttacks`, `DamDone`, `DamTaken` only. Document as data, not logic.
2. In `MultiHit`, **NPC branch**: port `fight.c:1033-1052` `temp_attacks += stance_index[ch.Stance].NumAttacks` so NPC stance attacks are additive to `NumAttacks`. Read NPC stance from `MobIndexData.Stance` if field exists; otherwise flag as a types extension.
3. In `MultiHit`, **PC branch**: before G3 cascade, add the GM bonus-attack loop matching `fight.c:1058-1069`: when `ch.PCData.Stances[ch.Stance] >= STANCE_GRAND_MASTER` and neither side is `STANCE_MONKEY`, fire `stance_index[ch.Stance].NumAttacks` extra `OneHit`s.
4. In `OneHit` damage calc, apply `dam_done` / `dam_taken` multipliers per `fight.c:2549-2594`. For NPCs, read from `ch.pIndexData.Stances[stance]` — if `MobIndexData.Stances` doesn't exist, add the field (tiny types extension) or flag as a documented gap.
5. (Optional, tiny) `DoStance` (`act/skills4.go:183-199`) bumps `PCData.Stances[i]` practice counter on selection.

- **Tests:** `TestMultiHit_StanceGM_BonusAttacks` (PC), `TestMultiHit_NPCStanceAdds_NumAttacks`, `TestMultiHit_StanceNonGM_NoBonus`, `TestMultiHit_MonkeyStanceSuppresses`, `TestOneHit_StanceDamDoneMultiplier` (PC + NPC), `TestOneHit_StanceDamTakenMultiplier`.

### G8 — Dual-wield weapon alternation in OneHit (S, bug fix)

When both `WEAR_WIELD` and `WEAR_DUAL_WIELD` equipped, `OneHit` must swing the correct weapon per call. C uses a `static bool dual_flip` (per-process — cross-fighter race bug). Port with a **per-character** `DualFlip bool` on `CharData`, or (cleaner) pass `wield` into `OneHit` from `MultiHit` as an argument.

Recommend argument-passing to avoid static state. Threads naturally with G1's signature change.

- **Tests:** `TestOneHit_DualWieldAlternatesWeapons`.

### G9 — Docs + audit sign-off (S)

Update `phases.md`, `audit-2026-04-17.md`, uncheck the P0 items in `TODO.md`. Append completion record to this plan file.

## Acceptance criteria

1. With RNG stubbed via a `var numberPercent = util.NumberPercent` swap: level-30 PC warrior with `Learned[second_attack]=100`, stub returning 50 → exactly 2 `OneHit` calls per round (spy). With `Learned[third_attack]=75`, stub returning 40 → 3 calls; stub returning 80 → 2 calls (boundary check).
2. Level-30 PC with no multi-attack skills learned gets exactly 1 attack/round (no regression).
3. Level-50 NPC with `NumAttacks=4` gets 4 attacks/round, independent of PC skill state.
4. PC dual-wielding with `Learned[dual_wield]=100` swings **offhand** stats on the bonus hit (G8 alternation fix).
5. PC with `Move<10` dual-wielding suffers `-20` `dual_bonus`; second_attack chance drops accordingly.
6. Level-10 PC wielding a DAM_SLASH weapon with `Learned[long_blades]=100` gets `victim_ac -= 5` and `dam += 1`; miss triggers `LearnFromFailureHook(ch, gsn_long_blades)`.
7. PC with `Stance == STANCE_DRAGON`, `PCData.Stances[STANCE_DRAGON] >= 200`, synthetic `stance_index[DRAGON].NumAttacks = 1` gets one extra attack before G3 cascade.
8. Either side in `STANCE_MONKEY` suppresses the stance bonus-attack and damage multipliers.
9. `AFF_BERSERK` with `Learned[berserk]=33` yields ~99% chance of the berserk extra hit.
10. `dt == gsn_backstab` never triggers the cascade (backstab remains one hit).
11. On victim death mid-cascade, no subsequent `OneHit` fires; retcode propagates.
12. `ACT_NOATTACK` mob never calls `OneHit`.

## Open questions / risks

- **G2 decision:** hook pattern vs new `internal/skillcheck` package. Hook matches existing `boot` wiring; package is cleaner but more churn. Default: hooks.
- **`stance_index` data source:** `db/system/stances.dat` is empty. Hard-code defaults from `src/stances.c`. Revisit with OLC/content later.
- **`PCData.Stances[]` practice counter:** not bumped today, so G7's GM loop is unreachable for existing players. Either seed admins to 200 via a one-time admin command, or add a follow-up task for a real `practice stance` flow.
- **Retcode callers:** `act/skills3.go:91-94,295` currently ignore OneHit's return; safe to widen signature.
- **`handler.AddTimer`** may not exist — check in G5; defer `TIMER_RECENTFIGHT` to follow-up if absent.
- **Devoted-clan favor penalty** (`fight.c:1245-1247, 1312-1313`) — low-impact, gate behind nil-check on `pcdata.Favor`; leave a TODO if incomplete.
- **Move cost per round** (`fight.c:1149-1171`) — out of scope; file as separate follow-up.

## TDD notes

- Use the existing `newCombatWorld` / `newFighter` builders in `combat_test.go:11-35`.
- Multi-attack statistical tests: stub `util.NumberPercent` via a `var numberPercent = util.NumberPercent` swap, or use the existing RNG-seed control pattern in `combat_test.go:73-104`.
- G6 (prof bonus) and G7 (stance) each require a test that constructs a minimal `ObjData` with `Value[3]` and a `CharData` with `Stances[]`; keep fixture builders local to the test file.
- After each G, run `go test -count=3 ./...` to shake out RNG flakes.

## Rough total effort

6 S + 3 M groups ≈ 2–3 focused TDD sessions. G1/G2 gate everything else; G3 is the tentpole; G6/G7 can land in parallel once G1 is in.

## Relevant file paths

- C: `/home/eilidh/src/smaug/src/fight.c` (973-1174, 1256-1317, 1385-1566, 2549-2594), `/home/eilidh/src/smaug/src/stances.c`
- Go implementation targets: `/home/eilidh/src/smaug/smaug-go/internal/combat/combat.go` (87-143, 156-228), new `combat/hooks.go`, new `combat/profbonus.go`, new `combat/stance_index.go`
- Go call sites: `/home/eilidh/src/smaug/smaug-go/internal/act/skills3.go` (91-94, 295), `/home/eilidh/src/smaug/smaug-go/internal/act/skills4.go:183`
- Wiring: `/home/eilidh/src/smaug/smaug-go/internal/boot/boot.go`
- Tests: `/home/eilidh/src/smaug/smaug-go/internal/combat/combat_test.go` (11-35, 73-104, 899-966)
- Data: `/home/eilidh/src/smaug/db/system/en/skills.dat:5233-5294`, `/home/eilidh/src/smaug/db/system/stances.dat`
