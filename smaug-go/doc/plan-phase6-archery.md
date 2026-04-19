# Plan: Phase 6 — Archery

**Status:** Authored 2026-04-18. Pending external adversary review before execution. Self-review substituted where the `Agent` tool is unavailable.
**Priority:** Wave 4 of Phase 6 (`phase6-roadmap.md`). Medium-scope cross-cutting addition: adds 3 new `WEAR_*` slots, 3 new commands, 1 combat-hook (`projectile_hit`), and 1 ranged-targeting helper (`ranged_attack` / `scan_for_victim`).
**Scope:** New files `internal/act/archery.go` + `internal/act/archery_test.go`. Modifications to `internal/types/enums.go` (add 3 `WEAR_LODGE_*` slots + bump `MAX_WEAR`), `internal/act/skills4.go` (replace `DoFire` stub — currently a `DoThrow` passthrough at `skills4.go:333-340`), `internal/boot/boot.go` (register `draw` + `dislodge`; `fire` is already registered at `boot.go:526` and will retarget the new implementation). No new package. Combat loop does **not** gain a `WEAR_MISSILE_WIELD` branch in `OneHit` — ranged attacks are triggered by explicit `DoFire` / `mob_fire`, not by the melee-round cascade (verified against `src/archery.c` — `projectile_hit` is only called from `ranged_got_target`, which is only called from `ranged_attack`, which is only called from `do_fire` / `mob_fire`).

---

## Problem

C ships a 1362-LOC archery subsystem at `src/archery.c`. The game-play feature set:

1. **Draw**: move a projectile from a worn quiver into the held slot, locking hands for ranged use (`do_draw`, `src/archery.c:125-200`).
2. **Fire**: loose the held projectile from a worn missile weapon at a named victim or through a compass direction; if a direction, the projectile scans up to N rooms (`bow->value[4]` clamped 1..10) for the named target through the exit chain, honoring closed doors, walls, tunnels, private/solitary/no-missile rooms, and sector-type distance costs (`do_fire` → `ranged_attack` → `scan_for_victim` + `ranged_got_target` → `projectile_hit`, `src/archery.c:1252-1362`, :873-1249, :789-868, :640-783, :262-635).
3. **Dislodge**: pull an arrow embedded in one of three body-slot wear locations (`WEAR_LODGE_RIB` / `WEAR_LODGE_ARM` / `WEAR_LODGE_LEG`); deal self-damage based on the projectile's `value[1..2]` dice roll (`do_dislodge`, `src/archery.c:203-255`).
4. **Arrow-lodges mechanic**: on a successful damaging `projectile_hit`, the projectile is equipped on the victim in one of the three `WEAR_LODGE_*` slots based on the hit-zone chance table (arm=1..3, leg=4..6, chest/rib=7..10), with `ITEM_LODGED` extra-flag and the corresponding `ITEM_LODGE_*` wear-flag (`src/archery.c:541-575`).
5. **Hit-roll**: full THAC0 interpolation (`thac0_00`/`thac0_32` scaled by level), damroll, prof_bonus, dist-penalty (`+2 * dist`), no-see modifiers, RIS (+RIS_PLUSx weapon tiers), victim position modifiers (`POS_BERSERK`/`POS_AGGRESSIVE`/`POS_DEFENSIVE`/`POS_EVASIVE`), sleeping doubler, enhanced-damage, weapon-spell application (`projectile_hit`, `src/archery.c:262-635`).
6. **Mob-fire**: NPC-side autonomous fire loop (`mob_fire`, `src/archery.c:1335-1362`) — invoked from mob-AI update.

The Go port has:

- `WEAR_MISSILE_WIELD int = 21` at `internal/types/enums.go:787`. Recognized by the equip path — `internal/act/obj.go:368` has `case types.WEAR_WIELD, types.WEAR_DUAL_WIELD, types.WEAR_MISSILE_WIELD` for the wield guard.
- `ITEM_PROJECTILE` and `ITEM_QUIVER` at `internal/types/enums.go:629-630`. Recognized by at least one caller: `internal/act/obj.go:125` asserts `container.ItemType != types.ITEM_QUIVER` for generic `put` gating.
- `DoFire` (stub) at `internal/act/skills4.go:330-340` currently passes through to `DoThrow` — "MVP: acts like a throw of a named object." Registered in `internal/boot/boot.go:526`.
- `DoThrow` at `internal/act/cmds2.go:188-193` — "hurls an object from inventory; the object leaves ch's room" — not archery; this is the generic grenade-like throw.
- Combat OneHit / MultiHit at `internal/combat/combat.go:173-468` — pure melee; no `WEAR_MISSILE_WIELD` branch (confirmed), which **matches C semantics** (C's `one_hit` in `src/fight.c` likewise does not read `WEAR_MISSILE_WIELD`; the melee cascade is strictly `WEAR_WIELD` + `WEAR_DUAL_WIELD`). Per-plan decision: we do NOT add a `WEAR_MISSILE_WIELD` branch to `OneHit`. Archery fires only through the explicit `fire` command (and a future mob-fire hook out of scope).

**Missing:**

- `WEAR_LODGE_RIB` / `WEAR_LODGE_ARM` / `WEAR_LODGE_LEG` constants (C `src/mud.h:2448-2450`, gated behind `#ifdef ENABLE_ARCHERY`). Per roadmap directive, port unconditionally.
- The matching `ITEM_LODGE_RIB` / `ITEM_LODGE_ARM` / `ITEM_LODGE_LEG` wear-flag bits (C uses them at `src/archery.c:220, 232, 244, 559, 565, 572` — `REMOVE_BIT(arrow->wear_flags, ITEM_LODGE_*)` and `SET_BIT` both sides of the mechanic). These are `ITEM_*` wear-flag bits (distinct from item-type enums) that need to exist on the bitmask where Go currently tracks wear-flags. Grep-verify during G1 which Go file owns the `ITEM_WEAR_*` wear-flag bitmask (likely `internal/types/enums.go` alongside `ITEM_TAKE` / `ITEM_WEAR_FINGER` / etc.).
- `ITEM_LODGED` extra-flag (C `xSET_BIT(arrow->extra_flags, ITEM_LODGED)` at `src/archery.c:552`). Grep-verify during G1 whether Go ships this already.
- `PROJ_BOLT` / `PROJ_ARROW` / `PROJ_DART` / `PROJ_STONE` constants (referenced in C at `src/archery.c:541, 1301, 1304, 1307, 1310`). Used both for ammo-type matching (`bow->value[5] != arrow->value[4]` gate) and for the "no bolts / no arrows / no darts / no slingstones" error message switch.
- `DoDraw` / `DoDislodge` commands.
- Real `DoFire` implementation replacing the throw-passthrough.
- `FindQuiver(ch)` / `FindProjectile(ch, quiver)` helpers (the C `find_quiver` at `archery.c:94-107` iterates `ch->last_carrying` back-walk, honors `can_see_obj`, accepts only non-`CONT_CLOSED` quivers).
- `ProjectileHit(ch, victim, wield, projectile, dist)` — the damage core.
- `RangedAttack(ch, argument, weapon, projectile, dt, range)` — the exit-chain walker with same-room / scan-for-victim branches.
- `ScanForVictim(ch, pexit, name)` — the exit-chain target-finder (walks up to `max_dist` rooms, respects sector-type distance costs; `char_from_room`/`char_to_room` idiom preserved so that `get_char_room` can be used with the searcher's POV).
- `RangedGotTarget(ch, victim, weapon, projectile, dist, dt, stxt, color)` — the safe-room / skill-check gate.
- `weapon_prof_bonus_check` — already partially shipped in `internal/combat/profbonus.go` for melee; archery uses `gsn_archery` / `gsn_blowguns` / `gsn_slings` skill branch via the `projectile->value[3]` switch. May need extension.
- `ITEM_MISSILE_WEAPON` in Go — verify at G1; the `projectile_hit` path uses `wield->item_type == ITEM_MISSILE_WEAPON` (`archery.c:302`).
- No `mob_fire` triggering path in mob-AI update. Out of scope for this plan — tracked as follow-up.

**`#ifdef ENABLE_ARCHERY` in C — port policy:** the three wear slots + the three item-lodge wear-flags are in the archery ifdef. `do_fire` body is in the archery ifdef. Ship unconditionally in Go (no build tags, no runtime toggle). This matches the roadmap Non-goal "No gameplay invented" lineage — we're porting the feature as the C shipping state intends, with the ifdef serving merely as the build-time enable.

**Value-slot conventions (C archery):** for projectiles, `value[1]` = low-dam, `value[2]` = high-dam, `value[3]` = damtype (TYPE_HIT offset: 13/14 = archery, 15 = blowgun, 16 = sling), `value[4]` = ammo type (must equal bow's `value[5]`), `value[5]` = PROJ_* kind. For bows (ITEM_MISSILE_WEAPON), `value[1]` = low-proj-bonus, `value[2]` = high-proj-bonus, `value[3]` = damtype offset, `value[4]` = max-dist (1..10), `value[5]` = accepted-ammo PROJ_* kind. **Verify each during G4 implementation** — the C source has load-bearing asymmetries (note the in-source "OOPS - had these backwards" comment at `archery.c:174`).

---

## C Reference (authoritative)

All citations against `src/archery.c` and `src/mud.h` (HEAD).

### Helpers

- `find_quiver(ch)` — `archery.c:94-107`. Back-walk `ch->last_carrying`, return first visible `ITEM_QUIVER` with `!IS_SET(value[1], CONT_CLOSED)`.
- `find_projectile(ch, quiver)` — `archery.c:109-122`. Back-walk `quiver->last_content`, return first visible `ITEM_PROJECTILE`.

### `do_draw` — `archery.c:125-200`

1. Must wield `WEAR_MISSILE_WIELD` → "You are not wielding a missile weapon!".
2. Must have quiver via `find_quiver` → "You aren't wearing a quiver where you can get to it!".
3. Count occupied hands (`WEAR_LIGHT` + `WEAR_SHIELD` + `WEAR_HOLD` + `WEAR_WIELD`). `hand_count > 1` → "You need a free hand to draw with." (**`WEAR_DUAL_WIELD` is NOT counted** — C bug or feature; port verbatim.)
4. `WEAR_HOLD` occupied → "Your hand is not empty!" (redundant with step 3 since `WEAR_HOLD` counted; the second gate fires when `hand_count == 1` and the one hand is specifically HOLD).
5. `find_projectile(ch, quiver)` NULL → "Your quiver is empty!!".
6. `separate_obj(arrow)` — splits a stacked-object group; port via existing Go helper (verify G4).
7. `bow->value[5] != arrow->value[4]` → "You drew the wrong projectile type for this weapon!" + put arrow back into quiver.
8. `WAIT_STATE(ch, PULSE_VIOLENCE)`.
9. Act to room: `"$n draws %s from $p."` with arrow `short_descr`. Act to char: `"You draw %s from $p."`.
10. `obj_from_obj(arrow)` then `obj_to_char(arrow, ch)`.
11. `wear_obj(ch, arrow, TRUE, -1)` — let the wear-path figure the HOLD slot from the arrow's wear-flags.

### `do_dislodge` — `archery.c:203-255`

1. Empty arg → "Dislodge what?". (**Note C quirk**: the arg is consumed only for the empty-check; the actual arrow is located by `WEAR_LODGE_*` slot scan, not by name match. Port verbatim.)
2. Scan `WEAR_LODGE_RIB` → `WEAR_LODGE_ARM` → `WEAR_LODGE_LEG` in order (first non-nil wins):
   - RIB: `AT_CARNAGE` "With a wrenching pull, you dislodge $p from your chest." + ROOM "$N winces in pain as $e dislodges $p from $s chest.". Damage = `number_range(3*val[1], 3*val[2])`.
   - ARM: "With a tug you dislodge $p from your arm." / "$N winces in pain as $e dislodges $p from $s arm.". **Damage = `number_range(3*val[1], 2*val[2])`** — note the `3` vs `2` asymmetry; port verbatim (potential C bug but authoritative).
   - LEG: "With a tug you dislodge $p from your leg." / "$N winces in pain as $e dislodges $p from $s leg.". Damage = `number_range(2*val[1], 2*val[2])`.
3. For each: `unequip_char(ch, arrow)` + `REMOVE_BIT(arrow->wear_flags, ITEM_LODGE_*)` + `xREMOVE_BIT(arrow->extra_flags, ITEM_LODGED)` + `damage(ch, ch, dam, TYPE_UNDEFINED)`. **Arrow stays in inventory.**
4. No lodge → "You have nothing lodged in your body.".

### `do_fire` — `archery.c:1252-1329`

1. `WEAR_MISSILE_WIELD` empty → "But you are not wielding a missile weapon!!".
2. Empty arg AND no `ch->fighting` → "Fire at whom or what?".
3. Arg == "none"/"self" OR victim == ch → "How exactly did you plan on firing at yourself?". (**C bug**: `victim` is never assigned before this check — it's initialized NULL at L1255; the `victim == ch` branch is structurally dead. Port as-is or fix — recommend port-verbatim.)
4. `WEAR_HOLD` empty OR not `ITEM_PROJECTILE` → "You are not holding a projectile!".
5. `max_dist = URANGE(1, bow->value[4], 10)`.
6. `bow->value[5] != arrow->value[4]` → switch on `bow->value[5]` emit specific "You have no bolts / arrows / darts / slingstones / nothing to fire..." message.
7. `WAIT_STATE(ch, 6)`.
8. `ranged_attack(ch, argument, bow, arrow, TYPE_HIT + arrow->value[3], max_dist)`.

### `projectile_hit` — `archery.c:262-635`

The damage core. 370 LOC. Flow:

1. Null projectile → `rNONE`.
2. Compute `dt`: `TYPE_HIT + projectile->value[3]` for ITEM_PROJECTILE or ITEM_WEAPON; `TYPE_UNDEFINED` otherwise. `proj_bonus = number_range(wield->value[1], wield->value[2])` if wield, else 0.
3. Victim dead → `extract_obj(projectile)` + return `rVICT_DIED`.
4. `prof_bonus = weapon_prof_bonus_check(ch, wield, &prof_gsn)` if wield. (Wires to `gsn_archery` / `gsn_blowguns` / `gsn_slings` via `ranged_got_target`'s upstream `wtype` switch — see below.)
5. If `dt == TYPE_UNDEFINED`, coerce to `TYPE_HIT + (wield->value[3] if wield->item_type == ITEM_MISSILE_WEAPON)`.
6. THAC0 calc: `thac0_00`/`thac0_32` from class table (NPC: `mobthac0`, `thac0_32=0`); `thac0 = interpolate(level, thac0_00, thac0_32) - GET_HITROLL(ch) + (dist * 2)`. **Distance penalty = `+2 per room`** — novel vs melee.
7. `victim_ac = UMAX(-19, GET_AC(victim)/10)`.
8. `!can_see_obj(victim, projectile)` → `+1 ac`; `!can_see(ch, victim)` → `-4 ac`.
9. `victim_ac += prof_bonus`.
10. Roll: `while ((diceroll = number_bits(5)) >= 20) ;`  — reroll until <20.
11. Miss: `diceroll == 0 || (diceroll != 19 && diceroll < thac0 - victim_ac)` → learn-from-failure, 50% extract / 50% drop-to-victim-room, `damage(ch, victim, 0, dt)`, return `rNONE`.
12. Hit: pchance = `number_range(1, 10)` — hit-zone:
    - 1..3 (arm): `dam = number_range(val[1], val[2]) + proj_bonus`.
    - 4..6 (leg): `dam = number_range(2*val[1], 2*val[2]) + proj_bonus`.
    - 7..10 (chest): `dam = number_range(3*val[1], 3*val[2]) + proj_bonus`.
13. Add GET_DAMROLL. `if prof_bonus: dam += prof_bonus/4`.
14. Position modifiers: `POS_BERSERK` *1.2, `POS_AGGRESSIVE` *1.1, `POS_DEFENSIVE` *0.85, `POS_EVASIVE` *0.8.
15. Enhanced-damage: `learned[gsn_enhanced_damage] > 0 && number_percent() < learned` → `dam += dam * learned / 120`; else `learn_from_failure`.
16. Sleeping doubler: `!IS_AWAKE(victim)` → `dam *= 2`.
17. `dam = max(dam, 1)` (C: `if (dam <= 0) dam = 1`).
18. RIS: magic if `ITEM_MAGIC`, else nonmagic. Then RIS_PLUSx via `obj_hitroll(wield)` — iterate `RIS_PLUS1..RIS_PLUS6` for the victim's immune/resistant/susceptible bit, compute mod, apply.
19. Immune (`dam == -1`): dispatch to skill's `imm_char`/`imm_vict`/`imm_room` messages (skill_table[dt]) if present; then 50% extract / 50% drop-to-victim-room; return `rNONE`.
20. `retcode = damage(ch, victim, dam, dt)`. If `retcode != rNONE`:
    - If `projectile->value[5] == PROJ_STONE` → extract.
    - Else-if victim died → extract, return `rVICT_DIED`.
    - **Else: LODGE mechanic** — `obj_from_char(projectile)` + `obj_to_char(projectile, victim)` + `xSET_BIT(projectile->extra_flags, ITEM_LODGED)` + by `pchance`: `SET_BIT(wear_flags, ITEM_LODGE_ARM/LEG/RIB)` + `wear_obj(victim, projectile, TRUE, get_wflag("lodge_arm"/"lodge_leg"/"lodge_rib"))`.
    - Return `retcode`.
21. Char/victim died post-damage → extract, return corresponding ret.
22. If `dam == 0` → 50% extract / 50% drop-to-victim-room; return.
23. Weapon-spell iter (`APPLY_WEAPONSPELL` on `pIndexData->first_affect` and `first_affect`), skip if victim immune to magic or room NO_MAGIC.
24. `extract_obj(projectile)` + `tail_chain()` + return.

### `ranged_got_target` — `archery.c:640-783`

- ROOM_SAFE → projectile blasted: "godly presence smites $p" + return `rNONE`.
- Determine `wtype` by `projectile->value[3]`: 13,14=archery; 15=blowguns; 16=slings.
- Skill-check gate: `number_percent() > 50 || (projectile && weapon && can_use_skill(ch, number_percent(), wtype))` → dispatch to `projectile_hit` (or `spell_attack` if no projectile). On fail → learn-from-failure on wtype, `damage(ch, victim, 0, dt)`, 50% extract / 50% drop.
- If NPC victim, `xREMOVE_BIT(victim->act, ACT_SENTINEL)` so the mob pursues.

### `scan_for_victim` — `archery.c:789-868`

- `AFF_BLIND` OR `pexit == NULL` → NULL.
- Walk exits with level-scaled max_dist: base 8, -1 per gap below 50/40/30.
- Loop: if `EX_CLOSED` or `room_is_private()` (below override) → break. `char_from_room(ch)` + `char_to_room(ch, pexit->to_room)`. `get_char_room(ch, name)` → if found, restore and return.
- Sector-cost dist increments: INSIDE/FIELD/UNDERGROUND +=1; FOREST/CITY/DESERT/HILLS +=2; WATER_SWIM/WATER_NOSWIM +=3; MOUNTAIN/UNDERWATER/OCEANFLOOR +=4; AIR +=1 with 80% chance; default +=1.
- `pexit = get_exit(ch->in_room, dir)` — same-dir walk chain until dist exceeds max or exit gone.
- Always restore ch's original room before returning.

### `ranged_attack` — `archery.c:873-1249`

Large exit-chain walker. Key gates:

- Quote-arg swallow at arg[0]=='\''.
- `find_door(ch, arg, TRUE)` → if no exit, try `get_char_room(ch, arg)` (same-room victim); else "Aim in what direction?".
- Same-room victim = `who_fighting(ch)` → "They are too close to release that type of attack!".
- No victim + (PRIVATE | SOLITARY) → "You cannot perform a ranged attack from a private room.".
- `tunnel > 0` and room_count >= tunnel → "This room is too cramped…".
- `pexit && !pexit->to_room` → "Are you expecting to fire through a wall!?".
- `pexit && EX_CLOSED` → "fire through a door" (or "wall" if EX_SECRET/EX_DIG).
- `pexit && arg1` → `scan_for_victim`; if NULL → "You cannot see your target.". If victim's room is `ROOM_NOMISSILE` → "You can't get a clean shot off.". If victim `num_fighting > max_fight(victim)` → "There is too much activity there…".
- `vch && !NPC && !NPC && PLR_NICE` → "Your too nice to do that!" (port verbatim typo).
- `is_safe(ch, vch, TRUE)` → bail `rNONE`.
- `separate_obj(projectile)`.
- Act messages (fire-weapon vs throw, exit-direction vs same-room). See `archery.c:1028-1091` — verbatim port.
- If same-room victim: `check_illegal_pk` + `check_attacker` + `ranged_got_target(...)`.
- Else: exit-chain walk with `char_from_room`/`char_to_room`. Per room: closed-door intercept message, random-target fallback (`num_bits(1)==0` for opposite-NPC-ness alternation), `is_safe` bail, same-room-as-victim → `ranged_got_target`. End-of-range → "falls harmlessly", drop to the final room. Wall hit (no exit) → "bounces harmlessly", drop. Always restore `was_in_room` at end.

### `mob_fire` — `archery.c:1335-1362`

Out of scope for this plan. Leave a stub doc comment in Go.

---

## Go Current State

| Component | Location | Status |
|---|---|---|
| `WEAR_MISSILE_WIELD` constant | `internal/types/enums.go:787` | Shipped |
| `ITEM_PROJECTILE` / `ITEM_QUIVER` constants | `internal/types/enums.go:629-630` | Shipped |
| `WEAR_LODGE_RIB` / `_ARM` / `_LEG` constants | — | **Missing** (G1) |
| `ITEM_LODGE_*` wear-flag bits | — | **Missing** (G1, grep-verify location) |
| `ITEM_LODGED` extra-flag | — | **Grep-verify in G1**; add if absent |
| `PROJ_ARROW` / `_BOLT` / `_DART` / `_STONE` | — | **Grep-verify in G1**; add if absent |
| `ITEM_MISSILE_WEAPON` item-type | `internal/types/enums.go:628` | Shipped |
| `DoFire` | `internal/act/skills4.go:333` | Stub — passes to `DoThrow` |
| `DoDraw` | — | **Missing** (G2) |
| `DoDislodge` | — | **Missing** (G4) |
| `FindQuiver` / `FindProjectile` helpers | — | **Missing** (G2) |
| `ProjectileHit` / `RangedAttack` / `ScanForVictim` / `RangedGotTarget` | — | **Missing** (G5) |
| `WeaponProfBonusCheck` for archery gsns | `internal/combat/profbonus.go` (for melee) | Shipped for melee; verify archery `gsn_archery` / `gsn_blowguns` / `gsn_slings` wired in G5 |
| `gsn_archery`, `gsn_blowguns`, `gsn_slings` | — | Grep-verify — likely shipped via skill-table load |
| `SeparateObj` helper | — | Grep-verify |
| `TailChain` | — | Likely a no-op in Go; drop the call |
| `can_see` / `can_see_obj` | `internal/handler/` | Shipped |
| `GetEqChar` | `internal/handler/` | Shipped |
| `ObjFromObj` / `ObjToChar` / `ObjFromChar` / `ObjToRoom` | `internal/handler/` | Shipped |
| `RoomIndexData.Tunnel` | Grep-verify | Expected shipped |
| `ROOM_SAFE` / `ROOM_PRIVATE` / `ROOM_SOLITARY` / `ROOM_NOMISSILE` / `ROOM_NO_MAGIC` room flags | `internal/types/enums.go` | **Shipped** — `ROOM_NOMISSILE` at `enums.go:722` (audit-archery 2026-04-18) |
| `find_door` helper | `internal/act/move.go:12` | **Shipped** as lowercase-private `findDoor` (audit-archery 2026-04-18) |
| `get_char_room` / `get_char_world` | `internal/handler/find.go` | Shipped |
| `IsSafe(ch, vch, check_friendly)` | — | Missing — port during G5 |
| `check_illegal_pk` / `check_attacker` | Grep-verify | Grep-verify |
| `PLR_NICE` flag | `internal/types/enums.go:975` | **Shipped** (audit-archery 2026-04-18) |

Combat loop: `internal/combat/combat.go:173-468` has no `WEAR_MISSILE_WIELD` branch; this matches C semantics and is intentional — **do not modify `OneHit` or `MultiHit`.**

---

## Go Design

**Package placement:** new `internal/act/archery.go`. Helpers (`FindQuiver`, `FindProjectile`) exported from `act` (not combat) because they're inventory-scanners, not damage-calc. `ProjectileHit` and `RangedAttack` live in the same archery.go — their only Go-external caller is `DoFire` (same package) and a future `MobFire`. Keep them exported for testability.

**Why archery is NOT in `internal/combat/`:** `combat/` is the melee-round package (`OneHit`, `MultiHit`, `DamMessage`). Archery runs outside the combat round, triggered by explicit command. Mixing would require `combat` to import `handler.ObjToRoom` / `handler.GetCharRoom`, which complicates the existing dependency graph. `act` already imports both. The archery damage path calls `combat.Damage(...)` through the existing import route.

**Test seams:** the dice-rolls at `projectile_hit:343-346` (the thac0 roll) and `:379` (the hit-zone pchance) and `:357, :422, :522, :593, :768` (the 50/50 projectile-drop) all need seams to deterministically exercise miss/hit paths and hit-zone selection. Use the existing `rollD20` / `numberPercent` seams from Tier 14 — verify they cover the range we need (`number_range`, `number_bits(5)`, `number_percent`). If any are missing seams, extend in G5's test scaffolding. `IsAttackSuppressed` (Tier 11) is orthogonal and not relevant here — archery has its own `is_safe` / `ROOM_SAFE` gates.

**Act-color choice:** Tranche C shipped the per-call color parameter for `util.Act`. Use `AT_ACTION` for draw-room / draw-self messages, `AT_CARNAGE` for dislodge-self + dislodge-room, `AT_GREY` for fire/throw, `AT_MAGIC` for the spell-variant branch in `ranged_attack`, `AT_HIT`/`AT_HITME` for RIS-immune skill messages. All colors cited verbatim from C source.

**Quiver auto-reload on depletion:** the roadmap's G7 proposal. The C source does NOT auto-reload — fire requires draw-draw-fire-fire-fire, each draw moves one arrow. **We port C exactly: no auto-reload.** If the orchestrator wants auto-reload as a UX improvement, it must be a separate plan (new gameplay). Strike G7 from the group list. See Open Question 4.

**Mutation verification safety:** per `_shared.md`, workers doing mutation-verify must avoid `git checkout`, `git restore`, `git stash`, `git reset --hard`. Use `Edit` with the exact forward + reverse delta. All G1..G6 instructions below use the Edit-only revert pattern.

---

## Task Groups

### G1 — Constants: WEAR_LODGE slots + ITEM_LODGE wear-flags + ITEM_LODGED extra-flag + PROJ_ kinds

**Deliverable:** add missing constants to `internal/types/enums.go` (or the appropriate existing file for each flag-family — grep-verify during work).

**Files:**
- `internal/types/enums.go` (add `WEAR_LODGE_RIB = 26`, `WEAR_LODGE_ARM = 27`, `WEAR_LODGE_LEG = 28`; bump `MAX_WEAR = 29`).
- `internal/types/constants.go:564-589` — the wear-flag bitmask is defined here. Add `ITEM_LODGE_RIB = 1 << 22`, `ITEM_LODGE_ARM = 1 << 23`, `ITEM_LODGE_LEG = 1 << 24` (bits 22/23/24 are currently free; matches C's BV22/23/24 at `mud.h:2214-2216`). **Also bump `ITEM_WEAR_MAX = 21` at `constants.go:588` to `24`** — this constant documents the highest bit in use and must stay in sync. (audit-archery 2026-04-18)
- Wherever extra-flags live (grep `ITEM_MAGIC` to find the file) — add `ITEM_LODGED` if missing.
- Wherever projectile kinds live (grep `PROJ_` or check `internal/types/object.go`) — add `PROJ_ARROW`, `PROJ_BOLT`, `PROJ_DART`, `PROJ_STONE`. Values must match C (`src/mud.h` — search the archery section for `#define PROJ_`).

**Tests (G1):** table-driven test in `internal/types/archery_constants_test.go` asserting each new constant has the expected value and the expected range (e.g. `WEAR_LODGE_RIB < MAX_WEAR`).

**Mutation-verify:** use `Edit` to flip `WEAR_LODGE_RIB = 26` to `27` → test fails → `Edit` back → test passes.

**Acceptance G1:**
- G1-1: `WEAR_LODGE_RIB` / `_ARM` / `_LEG` are int constants; `MAX_WEAR == 29`.
- G1-2: `ITEM_LODGE_RIB` / `_ARM` / `_LEG` are wear-flag bitmask constants.
- G1-3: `ITEM_LODGED` extra-flag exists.
- G1-4: `PROJ_ARROW` / `_BOLT` / `_DART` / `_STONE` exist with C-matching values.
- G1-5: No prior constant's value shifted (grep-scan `MAX_WEAR` callers for value dependencies).

### G2 — Quiver / projectile helpers

**Deliverable:** `FindQuiver(ch *CharData) *ObjData` and `FindProjectile(ch *CharData, quiver *ObjData) *ObjData` at `internal/act/archery.go`.

Port-verbatim from `src/archery.c:94-107` and `:109-122`. Use the Go inventory linked-list conventions (verify via `internal/handler/` whether Go uses `Carrying []*ObjData` or a linked list — match C's back-walk semantics: first visible match from tail).

**Tests (G2):** `internal/act/archery_test.go`:
- `FindQuiver` returns first open quiver.
- `FindQuiver` skips closed quivers (`CONT_CLOSED` bit of `Value[1]`).
- `FindQuiver` skips non-quivers.
- `FindQuiver` skips invisible quivers (returns nil if all are hidden from ch).
- `FindProjectile` returns first visible `ITEM_PROJECTILE` in the quiver's contents.

**Mutation-verify:** flip the item-type check to `ITEM_WEAPON` → tests fail → revert.

**Acceptance G2:**
- G2-1: both helpers unit-pass.
- G2-2: visibility-gating honored.

### G3 — DoDraw

**Deliverable:** `DoDraw(ch, argument)` at `internal/act/archery.go`. Port-verbatim `src/archery.c:125-200`.

**Tests (G3):** `TestDoDraw_NoBow`, `_NoQuiver`, `_HandsBusy`, `_HandsHold` (after `WEAR_HOLD` equipped), `_EmptyQuiver`, `_WrongAmmoType`, `_Success`. Assert wait-state set (`PULSE_VIOLENCE`), arrow moved from quiver to character's `WEAR_HOLD`, message order matches ("draws" act-to-room + act-to-self).

**Mutation-verify:** remove the `bow->value[5] != arrow->value[4]` ammo-type gate → "wrong ammo" test fails → revert.

**Acceptance G3:**
- G3-1: all DoDraw tests pass.
- G3-2: unit registered in `boot.go` (new `reg.Register draw`).

### G4 — DoDislodge

**Deliverable:** `DoDislodge(ch, argument)` at `internal/act/archery.go`. Port-verbatim `src/archery.c:203-255`.

**Tests (G4):** `TestDoDislodge_NoArg`, `_NothingLodged`, `_RibSuccess` (check damage range 3*val[1]..3*val[2] with a deterministic seam), `_ArmSuccess` (3*val[1]..2*val[2] — **preserve the C asymmetry**), `_LegSuccess` (2*val[1]..2*val[2]), `_RibPriority` (rib beats arm when both lodged). Assert arrow unequipped, flags cleared, arrow stays in inventory.

**Mutation-verify:** swap the arm damage formula to `3*val[1]..3*val[2]` → test fails → revert.

**Acceptance G4:**
- G4-1: all DoDislodge tests pass.
- G4-2: `boot.go` registers `dislodge`.
- G4-3: arm's `3*val[1] / 2*val[2]` asymmetry is preserved (explicitly tested).

### G5 — DoFire and ranged-hit core

**Deliverable:** `DoFire` replaces the `skills4.go:333` stub. New helpers `ProjectileHit`, `RangedAttack`, `ScanForVictim`, `RangedGotTarget` at `internal/act/archery.go`. Port-verbatim the full `src/archery.c:262-1329` surface.

**Files:**
- `internal/act/archery.go` (new helpers).
- `internal/act/skills4.go` — delete `DoFire` stub (lines 330-340); leave a short doc-comment pointing to `archery.go`. **Preserve the function name so the `boot.go:526` registration keeps resolving.**
- `internal/boot/boot.go` — `fire` already registered (line 526); no change needed beyond `draw` + `dislodge` additions.

**Sub-groups:**

- **G5a `ProjectileHit`**: 370-LOC port. Test vectors: miss via low diceroll, hit-zone arm/leg/rib via stubbed `number_range`, lodge mechanic on success (arrow equipped in victim's `WEAR_LODGE_*`), `PROJ_STONE` extraction branch, dead-victim early-return, immune-to-damage skill-message branch, RIS_PLUSx modifier, weapon-spell application.
- **G5b `RangedGotTarget`**: `ROOM_SAFE` gate, wtype-branch (archery/blowguns/slings), skill-check pass → `ProjectileHit`, fail → learn-from-failure + drop.
- **G5c `ScanForVictim`**: exit-chain walker; level-scaled max_dist; sector-cost dist increments (all 11 sector types); closed-door + private-room breaks; always restore `in_room`.
- **G5d `RangedAttack`**: all 370 lines of `src/archery.c:873-1249`; quote-swallow arg handling; `find_door` vs `get_char_room` disambiguation; PRIVATE/SOLITARY/NOMISSILE/`tunnel`/`num_fighting>max_fight` gates; PLR_NICE gate; act-message emission (fire vs throw, exit vs same-room); exit-chain walk with closed-door intercept, random-target fallback, out-of-range fallout, wall-hit fallout.
- **G5e `DoFire`** (final): 80-LOC port of `src/archery.c:1252-1329`.

**Tests (G5):** for each sub-group, a named-test per gate. Count target: ~40 tests total covering:
- DoFire early-exits (5 tests): no-bow, no-arg, arg=self, not-holding-projectile, wrong-ammo-type.
- DoFire success (1 test): delegates to RangedAttack.
- RangedAttack gates (8 tests): private-room, tunnel-overcrowded, closed-door, wall, scan-not-found, NOMISSILE, too-close (fighting), PLR_NICE.
- RangedAttack same-room-victim (2 tests): target in same room as ch; victim is currently fighting (reject).
- RangedAttack exit-chain (4 tests): projectile flies through N rooms, hits wall, falls out of range, intercepts closed door.
- RangedGotTarget (3 tests): ROOM_SAFE, skill-check pass (→ ProjectileHit), skill-check fail (→ learn-from-failure + drop).
- ScanForVictim (6 tests): blind ch, closed-door break, private-room break, level-scaled max_dist, sector-cost arithmetic, always-restore-room.
- ProjectileHit (12 tests): null projectile, dead victim, miss (low-roll), hit arm/leg/rib (damage range), position modifiers, enhanced-damage, sleeping doubler, RIS immune message, lodge success, PROJ_STONE extract, weapon-spell iter, dam==0 drop.

**Mutation-verify:** for each sub-group, pick one conditional (e.g., the `dist*2` THAC0 penalty multiplier in ProjectileHit) and flip it (`dist*3`) to confirm tests detect.

**Acceptance G5:**
- G5-1: all 40+ tests pass.
- G5-2: `skills4.go:333 DoFire` stub replaced; function body now delegates to the archery-module impl (still exported as `act.DoFire` so `boot.go` keeps resolving).
- G5-3: `go test ./internal/act/...` all pass.
- G5-4: no melee-combat test (combat/combat_test.go) regresses.

### G6 — Boot registration + integration smoke

**Deliverable:** register `draw` + `dislodge` in `internal/boot/boot.go` in the same style as existing archery/combat commands (`Position: POS_FIGHTING, Level: 0`). `fire` is already registered.

**Files:**
- `internal/boot/boot.go` — add two `reg.Register` calls near line 526.

**Tests (G6):** integration test in `internal/boot/boot_test.go` (if one exists) or a new `internal/act/archery_integration_test.go` that invokes the boot registry and asserts `draw` / `fire` / `dislodge` resolve.

**Mutation-verify:** remove the `dislodge` registration → integration test fails → revert.

**Acceptance G6:**
- G6-1: `draw` / `fire` / `dislodge` each resolve in the command registry.
- G6-2: `go build ./...` succeeds.
- G6-3: `go test ./...` succeeds.

---

## Acceptance Criteria (overall)

Mechanically verifiable end-to-end conditions. The plan is done when:

1. **AC1**: `types.WEAR_LODGE_RIB`, `WEAR_LODGE_ARM`, `WEAR_LODGE_LEG` are defined `int` constants with consecutive values 26/27/28 and `MAX_WEAR == 29`.
2. **AC2**: `ITEM_LODGE_RIB`/`_ARM`/`_LEG` (wear-flag bits) and `ITEM_LODGED` (extra-flag) exist.
3. **AC3**: `PROJ_ARROW`/`BOLT`/`DART`/`STONE` exist with C-matching values.
4. **AC4**: `act.FindQuiver(ch)` returns the first visible open quiver from the tail of `ch.Carrying`.
5. **AC5**: `act.FindProjectile(ch, quiver)` returns the first visible projectile in the quiver.
6. **AC6**: `act.DoDraw` moves an arrow from quiver to WEAR_HOLD with the correct act messages and wait-state.
7. **AC7**: `act.DoFire` fires a held projectile through a bow at a named target; dispatch to `RangedAttack` is wired.
8. **AC8**: `RangedAttack` honors PRIVATE/SOLITARY/NOMISSILE/tunnel/closed-door/wall gates.
9. **AC9**: `ScanForVictim` walks ≤ level-scaled max_dist rooms with sector-cost arithmetic.
10. **AC10**: `ProjectileHit` honors full THAC0 + AC + prof_bonus + distance penalty + hit-zone damage + position modifiers + enhanced-damage + sleeping doubler + RIS.
11. **AC11**: On successful damaging hit, the projectile is equipped in the victim's `WEAR_LODGE_*` matching the hit-zone; `ITEM_LODGED` is set.
12. **AC12**: `DoDislodge` unequips the lodged arrow (rib / arm / leg, in that priority order), deals self-damage per the C formula (preserving the arm's `3*v[1]/2*v[2]` asymmetry), clears flags, keeps arrow in inventory.
13. **AC13**: `draw`, `fire`, `dislodge` resolve via the command registry at boot.
14. **AC14**: `go test ./...` all pass; `go build ./...` succeeds.
15. **AC15**: No melee-combat test regresses — `OneHit` / `MultiHit` / `DamMessage` test surfaces unchanged.

---

## Scope Cuts / Deferrals

- **`mob_fire` autonomous NPC firing** (`src/archery.c:1335-1362`): out of scope. Add stub doc-comment only; track as follow-up in `TODO-updates.md`. NPCs will not autonomously fire bows after this plan lands.
- **Quiver auto-reload** (G7 in the dispatch prompt): **omitted by design.** C does not auto-reload; porting auto-reload would be new gameplay. Tracked in Open Question 4 for human adjudication.
- **`gsn_archery` / `gsn_blowguns` / `gsn_slings` skill-table entries:** assumed shipped via skill-table load. If grep reveals they're missing, add as a G0 pre-req. Otherwise, wire calls only.
- **OLC support for bows/quivers/arrows:** C ships archery-aware `oedit` prompts; Go's `oset` path already supports all `Value[i]` fields generically. Not gapped.
- **`weapon_prof_bonus_check` extension for archery gsns:** if the existing melee-only implementation at `combat/profbonus.go` doesn't thread through `gsn_archery`/`blowguns`/`slings`, extend during G5; otherwise reuse.
- **`mob_fire`-triggering pulse** in mob-AI update (`update.c` equivalent): out of scope.

---

## Open Questions

**Q1 — `ITEM_LODGE_*` wear-flag bit values.** C uses `#define ITEM_LODGE_ARM` / `_LEG` / `_RIB` in `src/mud.h` at specific bit positions; need to verify they're free in Go's current `ITEM_WEAR_*` bitmask (no collision). If the Go bitmask has reallocated bits, pick fresh values and grep-scan for any callers that hard-code the numeric bit.

**Q2 — `ITEM_LODGED` extra-flag bit value.** Same as Q1 for extra-flags.

**Q3 — Combat-loop ranged-attack hook.** Dispatch prompt mentions "G5: Combat-loop ranged-attack hook (on OneHit / MultiHit)". **Recommendation: omit.** C's `one_hit` does not read `WEAR_MISSILE_WIELD`; ranged attacks are strictly explicit-command via `do_fire` / `mob_fire`. Wiring archery into the melee round would be new gameplay. Resolved in-plan as **omit**; flag for human veto if desired.

**Q4 — Quiver auto-reload after fire.** Dispatch prompt G7. C does not auto-reload — draw is a one-arrow action, fire consumes one arrow, then the hold slot is empty. **Recommendation: omit** (preserve C). If the human wants a UX-improvement auto-reload, track as a separate follow-up plan.

**Q5 — `PROJ_*` kind values.** Must match C exactly; verify against `src/mud.h` or the archery C header during G1 implementation. Wrong values break the `bow->value[5] != arrow->value[4]` ammo-match gate.

**Q6 — `ROOM_NOMISSILE` room-flag.** **RESOLVED (audit-archery 2026-04-18)** — present at `internal/types/enums.go:722`. No pre-req needed.

**Q7 — `PLR_NICE` flag.** **RESOLVED (audit-archery 2026-04-18)** — present at `internal/types/enums.go:975`. No pre-req needed.

**Q8 — `num_fighting` / `max_fight` wiring.** **PARTIALLY RESOLVED (audit-archery 2026-04-18)** — `CharData.NumFighting` already exists at `internal/types/character.go:78`. Only the `max_fight(ch)` helper is absent; port from `src/fight.c:281` during G5.

**Q9 — `separate_obj` / `tail_chain`.** Grep Go to verify `handler.SeparateObj` exists (for stacked-object splits) and whether `TailChain` is a no-op. If either is missing, stub accordingly during G3/G5.

**Q10 — `do_fire` bug: `victim == ch` check on unassigned pointer** (`src/archery.c:1272`). **Recommendation: port verbatim** (dead-code check; harmless). Flag for human if stricter fidelity desired.

**Q11 — Fixture creation for tests.** New test helpers will need bow/quiver/arrow object fixtures, plus exit/room chains for the scan-for-victim walker. Reuse `QuickLoginTwo` from Tier 9 where possible for two-char tests; most archery logic is single-char-to-NPC-target, so simpler fixtures suffice.

**Q12 — `mob_fire` value-index asymmetry.** `src/archery.c:1352` (inside `mob_fire`, deferred from scope) reads the ammo-match gate as `bow->value[4] != arrow->value[5]`, while `do_fire:1295` uses `bow->value[5] != arrow->value[4]`. The two indices are transposed between mob and PC fire paths. One of them is almost certainly a C bug. **Recommendation:** since `mob_fire` is deferred (Scope Cuts), port `DoFire` with `[5]!=[4]` verbatim (matches shipping PC behavior). When `mob_fire` eventually ports (follow-up plan), reconcile the discrepancy — preserve `[4]!=[5]` verbatim for fidelity OR fix to match `DoFire`. Flagged so the `mob_fire` follow-up author has the asymmetry on record. No action required in this plan. Tracked in CLAUDE.md Wave D follow-ups.

**Q13 — Lodged-arrow `remove` bypass.** A victim with a lodged arrow (wearing it in `WEAR_LODGE_RIB/_ARM/_LEG`) can call `remove arrow` to unequip it exactly like any other worn item — the standard `DoRemove` path honors `ITEM_NOREMOVE` but C does NOT set that flag on arrows at lodge time (`src/archery.c:455-460` equip path uses plain `equip_char`). The intended interaction is `DoDislodge` (self-damage + remove); `DoRemove` is a no-damage bypass. **Recommendation:** port verbatim (no `ITEM_NOREMOVE` set at equip time). A policy question for builders is whether to set `ITEM_NOREMOVE` in Go to close the bypass — that would be new gameplay, out of scope here. Flag as a UX gap for future review. Tracked in CLAUDE.md Wave D follow-ups.

---

## Risk Analysis

**Medium risk overall.** Subsystem is self-contained (one new `archery.go` file; one existing stub replaced; one `enums.go` extension) but cross-cuts several systems (inventory, combat, handler, room-exits).

- **R1 — Wear-flag bit collision** (Q1/Q2): if Go's bitmask reallocated bits in Phases 2-5, naive value choices will corrupt other gear. Mitigation: grep `ITEM_WEAR_FINGER` + run the full existing test suite after G1 to catch regressions.
- **R2 — ScanForVictim room-mutation idiom**: C's `char_from_room` + `char_to_room` mid-scan temporarily yanks the character out of their room. If Go has side-effects on room-enter (e.g. lighting, room-trigger progs), the scan may fire spurious triggers. Mitigation: use a non-mutating scan variant if the idiom exists in Go; otherwise, accept the fidelity (C has the same risk).
- **R3 — RIS_PLUSx bitmap ordering**: C iterates `RIS_PLUS1 << UMIN(plusris, 7)` — if Go's RIS bits are in a different order, the weapon-hitroll → RIS-plus mapping breaks silently. Mitigation: test G5-RIS path with explicit bit-value assertions.
- **R4 — Damage formula arm asymmetry** (`3*v[1] / 2*v[2]`): likely a C bug. Port verbatim to match shipping behavior; document in-test so future auditors know it's intentional.
- **R5 — `ranged_attack`'s room-walk restores `in_room`**: if ch is disconnected mid-walk (logout during fire), the restore may fire on a stale room. Low-probability; accept C-fidelity.
- **R6 — Package-dependency creep**: `internal/act/archery.go` will need imports of `handler`, `combat` (for `Damage`), `util` (for `Act`, `NumberRange`), `types`. Verify `act` already imports all of these; if not, minor boot-ordering concern.

---

## Adversary Verification Notes

*(Appended after plan adversary pass — pending.)*

---

## Completion Record

*(Appended after work lands — pending.)*
