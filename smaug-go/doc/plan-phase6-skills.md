# Plan: Phase 6 — Missing skills (`bloodlet` / `pounce` / `broach`)

**Status:** Planned (2026-04-18). External adversary audit 2026-04-18 (lineage `audit-skills`) — verdict PASS with minor clarifications applied. Corrections: Q1 "unreachable in C" claim refined (see below — reachable for dual-identity characters with distinct race/class blood assignments, but functionally dead for all typical single-identity characters, which still justifies Option A). No design changes.
**Priority:** Content (P3) — content-gap closure; unblocks nothing. Small, independent, safe to land anytime.
**Scope:** Three `DoFun` commands + `boot.go` registrations + tests. No new types, no new schema, no cross-package changes beyond `act` package and `boot.go`. ~200 C LOC collectively; estimated ~300 Go LOC with tests.
**Lineage mode:** Orchestrated (`LINEAGE_ID=phase6-skills`). Drafts under `.claude/drafts/phase6-skills/`; orchestrator merges into canonical docs on reconciliation.

---

## Problem

Three skills listed in `db/system/en/skills.dat` dispatch (via their `Code do_<name>` metadata) to C functions that have no Go equivalent. Today typing `bloodlet`, `pounce`, or `broach` at the command line falls through to the generic "command not recognized" path — the `skills.dat` metadata is present but the dispatch table in `internal/boot/boot.go` has no `Name: "bloodlet"` / `"pounce"` / `"broach"` entries.

Per-skill specifics:

### bloodlet (src/skills.c:3517-3565, 49 LOC)

A blood-race self-harm skill. Vampire or demon PC only. On success: decrements `COND_BLOODTHIRST`, creates an `OBJ_VNUM_BLOODLET` object (vnum 26, a "pool of let blood") in the room with `timer=1` (decays fast) and `value[1]=6` (blood quantity), then self-damages for `ch->level/5` HP. On failure: flavor text only, learnFromFailure, no damage, no object.

Purpose in-game: vampires offer blood to recharge bloodthirst? No — `bloodlet` *reduces* bloodthirst by 7 and damages the caster. The object drop is flavor; other systems (specifically `do_feed`) can consume the pool. Used as deity-favor ritual or flavor RP.

**C operator-precedence bug** (flagged for human confirmation): line 3521 reads `if (IS_NPC (ch) || !IS_VAMPIRE (ch) || !IS_DEMON (ch)) return;`. With `IS_VAMPIRE`/`IS_DEMON` being distinct predicates, `!A || !B` is `!(A && B)`, so the gate requires the caster to be BOTH vampire AND demon simultaneously. Because `IS_VAMPIRE(ch)` expands to `!IS_NPC(ch) && (race==RACE_VAMPIRE || class==CLASS_VAMPIRE)` and `IS_DEMON(ch)` is the parallel form over RACE_DEMON/CLASS_DEMON, the only characters that pass are those whose race is one blood identity AND class is the other (e.g., race=RACE_VAMPIRE + class=CLASS_DEMON). Plain single-identity characters (race=VAMPIRE only, or class=DEMON only) are ALL rejected. So the command is functionally dead for every typical player of either race, with an unreachable-in-practice edge case for dual-identity characters. Intent is almost certainly `IS_NPC(ch) || (!IS_VAMPIRE(ch) && !IS_DEMON(ch))` — "player AND (vampire OR demon)". The project rule "no gameplay invented" means we could port verbatim (and ship a near-dead command), OR we can fix the bug following the same principle as Holidays plan (which fixed 2 latent C bugs as documented corrections). See Open Question 1 below.

### pounce (src/skills.c:2607-2695, 88 LOC)

A stealth opening attack. Already partially wired: `gsnPounce` is resolved at boot via `combat.LookupSkillSlotHook("pounce")` (`internal/combat/skillcheck.go:110`), and `MultiHit` already short-circuits the attack cascade when `dt == gsnPounce` (`internal/combat/combat.go:247`) — so when the command eventually fires `MultiHit(w, ch, victim, gsnPounce)`, the single-hit behavior is correctly gated. The combat side of the skill is done; only the command-entry side is missing.

Gate sequence (C :2614-2673): NPC+AFF_CHARM early-return; mount early-return; arg required; `get_char_room`; self-target reject; `is_safe` reject; weapon `value[3]` (damage type) must be in `{1, 2, 3, 5, 10, 11}` = {DAM_SLICE, DAM_STAB, DAM_SLASH, DAM_CLAW, DAM_BITE, DAM_PIERCE}; victim already fighting reject; victim hurt+awake reject. Then `check_attacker`, `WAIT_STATE(ch, skill_table[gsn_pounce]->beats)`. Success path: `learn_from_success` + `multi_hit(ch, victim, gsn_pounce)` + `check_illegal_pk`. Failure path: `learn_from_failure` + `damage(ch, victim, 0, gsn_pounce)` (0-damage, just prints miss message) + `check_illegal_pk`.

### broach (src/skills.c:3960-4011, 52 LOC)

A forceful lock-breaking skill (NOT armor-piercing as the Phase 5 Tier 4 table hinted). Distinct from `pick` in that `pick` is thief-lockpicking (uses `gsn_pick_lock`); `broach` is a forceful break-the-lock alternative, `adjust_favor` on success (deity system). Green-colored output via `set_char_color(AT_DGREEN, ch)`.

Gate sequence (C :3965-3983): color-set; NPC+AFF_CHARM early-return; arg required; mount early-return; `WAIT_STATE(ch, skill_table[gsn_broach]->beats)`. Then `find_door(ch, arg, TRUE)`. If exit present: check `!EX_CLOSED || !EX_LOCKED || EX_PICKPROOF || !can_use_skill` → fail branch (with `learn_from_failure` + `check_room_for_traps`). Else success branch: remove EX_LOCKED bit from exit AND reverse-exit, `learn_from_success`, `adjust_favor(ch, 9, 1)`, `check_room_for_traps`.

**C predicate bug** (flagged): the fail-condition at :3987-3990 reads `(!CLOSED || !LOCKED || PICKPROOF || failed-can-use)`. Because `!CLOSED || !LOCKED` is a tautology whenever EITHER is false (which is commonly true on unlocked or open doors), this condition trips for most doors, meaning `broach` ALMOST ALWAYS fails on doors that are simultaneously closed AND locked. The probably-intended predicate is `CLOSED && LOCKED && !PICKPROOF && can_use_skill` for success, mirrored by its negation for fail. See Open Question 2.

---

## C Reference (authoritative)

### Entry-point signatures
- `src/skills.c:3517 do_bloodlet(CHAR_DATA *ch, char *argument)`
- `src/skills.c:2607 do_pounce(CHAR_DATA *ch, char *argument)`
- `src/skills.c:3960 do_broach(CHAR_DATA *ch, char *argument)`

### Per-skill metadata (from `db/system/en/skills.dat`)
| Skill | Minpos | Rounds (WAIT_STATE) | Minlevel | Mana | Dammsg |
|---|---|---|---|---|---|
| bloodlet | 105 (RESTING) | 12 | 11 | 0 | empty |
| pounce | 112 (STANDING) | 12 | 5 | 15 | `pounce~` |
| broach | 112 (STANDING) | 24 | 41 | 0 | empty |

**Minpos decoding:** Go's `convertMinPos` in `internal/persist/skills.go:86` maps C's 100+ positions to Go's POS_* values. 105 = POS_RESTING, 107 = POS_FIGHTING, 112 = POS_STANDING (matching existing Go skill registrations). Boot registration `Position` field uses POS_STANDING or POS_FIGHTING as appropriate.

### Dependent macros / helpers
- `IS_VAMPIRE(ch)` / `IS_DEMON(ch)` — `src/mud.h:4029-4034`. Expands to `!IS_NPC(ch) && (ch->race == RACE_VAMPIRE || ch->class == CLASS_VAMPIRE)` and equivalent for demon. Note: C's "race OR class" disjunction means either path qualifies.
- `WAIT_STATE(ch, beats)` — `src/mud.h` (WAIT_STATE macro). Sets `ch->wait = UMAX(ch->wait, beats)`.
- `can_use_skill` — matches Go's `canUseSkill` at `internal/act/skills.go:55`.
- `learn_from_success` / `learn_from_failure` — match Go's `learnFromSuccess` / `learnFromFailure` in the same file.
- `get_char_room` — matches Go's `handler.GetCharRoom`.
- `is_safe(ch, victim, show_msg)` — checks PK legality. Go has no direct port.
- `check_attacker` — marks ch as aggressor for illegal-PK tracking. Go has no direct port.
- `check_illegal_pk` — post-combat PK legality check. Go has no direct port.
- `find_door(ch, arg, quiet)` — matches Go's `findDoor` at `internal/act/move.go:12`, but Go's version lacks the `quiet` flag and doesn't emit "You see no door there" messages itself. Broach's C `find_door(ch, arg, TRUE)` passes quiet=TRUE — Go's findDoor is already silent.
- `check_room_for_traps` — trap-trigger dispatch on room entry / lock-pick. Go has no port (`TRAP_PICK` constant exists but no dispatcher).
- `adjust_favor(ch, category, amount)` — deity-favor adjustment. Go has no port.
- `multi_hit` — matches Go's `combat.MultiHit`.
- `damage` — matches Go's `combat.Damage`.
- `gain_condition(ch, cond, value)` — matches Go's `act.GainCondition` at `internal/act/consume.go:34`.
- `create_object(idx, level)` — matches Go's `handler.CreateObject(w, idx, level)` at `internal/handler/handler.go:105`.
- `get_obj_index(vnum)` — matches Go's `world.GetObjIndex(vnum)` at `internal/world/world.go:100`.
- `obj_to_room(obj, room)` — matches Go's `handler.ObjToRoom(obj, room)` at `internal/handler/handler.go:185`.
- `set_char_color(AT_DGREEN, ch)` — no Go equivalent. Go embeds color tokens inline (`&g`, `&G`, etc.).

### Ancillary Go constants / types that already exist
- `types.OBJ_VNUM_BLOODLET = 26` at `internal/types/constants.go:404`
- `types.RACE_VAMPIRE`, `types.RACE_DEMON` at `internal/types/enums.go:238-239`
- `types.CLASS_VAMPIRE = 4`, `types.CLASS_DEMON = 13` at `internal/types/constants.go:347/356`
- `types.COND_BLOODTHIRST` at `internal/types/enums.go:884`
- `types.AT_BLOOD = 1`, `types.AT_DGREEN = 2` at `internal/types/constants.go:493-494` (not yet in `util.atColorCode` map — see Open Question 3)
- `types.DAM_SLICE=1, DAM_STAB=2, DAM_SLASH=3, DAM_CLAW=5, DAM_BITE=10, DAM_PIERCE=11` at `internal/types/enums.go:1105-1116` (pounce's weapon-type gate list)
- `types.EX_CLOSED`, `types.EX_LOCKED`, `types.EX_PICKPROOF` at `internal/types/enums.go` (existing `DoPick` uses these)
- `types.AFF_CHARM` (existing)

### Existing area data
- Object vnum 26 (OBJ_VNUM_BLOODLET) prototype exists in `db/area/en/limbo.are2:806-811` — "blood pool spill bloodlet~ / a quantity of let blood~ / A pool of let blood glistens here.~". Shipped data; no area edit needed for bloodlet.

---

## Go Current State

**What is shipped:**
- `skills.dat` loader populates `world.Skills` entries for bloodlet/pounce/broach with `Beats` (from `Rounds`), `MinimumPos` (converted from Minpos), `SkillFunName = "do_<name>"` — all present in `WorldRef.Skills[gsn]`.
- `gsnPounce` slot resolved at boot (`internal/combat/skillcheck.go:110`) and consumed by `MultiHit`'s single-hit short-circuit (`internal/combat/combat.go:247`).
- `OBJ_VNUM_BLOODLET = 26` constant and the object prototype in area data — both already present.
- `COND_BLOODTHIRST` enum, `GainCondition`, `CreateObject`, `GetObjIndex`, `ObjToRoom`, `MultiHit`, `Damage`, `GetCharRoom`, `GetEqChar`, `findDoor`, `canUseSkill`, `learnFromSuccess`, `learnFromFailure` — all shipped.
- Reference implementations to mirror: `DoBackstab` (`internal/act/skills.go:188`) is the closest structural analog to `pounce` (target+weapon gate+single-hit); `DoPick` (`:441`) is the direct analog for `broach` (find_door + exit-bit check); `DoFeed` at `internal/act/skills4.go:240` and vampire-gate pattern at `spell_teleport_test.go:94` establish the race/class-gate pattern.

**What is missing (and in scope):**
- `DoBloodlet` function in `internal/act/skills4.go` (fits alongside `DoFeed` / `DoSkin` / `DoPoisonWeapon` — vampire/flavor cluster).
- `DoPounce` function in `internal/act/skills3.go` (fits alongside `DoCircle` / `DoGouge` / `DoStun` — the Wave 2 offensive-combat cluster per `skills3.go:50`).
- `DoBroach` function in `internal/act/skills.go` (fits alongside `DoPick` — both are door/lock skills).
- `boot.go` registrations for all three.
- A helper for the C `IS_VAMPIRE`/`IS_DEMON` predicate (since two consumers — `DoBloodlet` and a future vampire-skill — will want it). Proposed: `act.isBloodRace(ch) bool` local helper in `skills4.go`, returning `!ch.IsNPC() && (ch.Race == types.RACE_VAMPIRE || ch.Class == types.CLASS_VAMPIRE || ch.Race == types.RACE_DEMON || ch.Class == types.CLASS_DEMON)`. Used only for bloodlet; extract to a shared location if a second consumer appears.

**What is missing (and OUT of scope):**
- `IsSafe` / `CheckAttacker` / `CheckIllegalPK` PK-legality helpers — not ported; `DoBackstab` / `DoCircle` don't use them either. Pounce matches the existing pattern (no PK-legality calls).
- `check_room_for_traps` — not ported; `DoPick` doesn't use it either. Broach matches (no trap dispatch).
- `adjust_favor` — not ported; `DoPick` doesn't use it either. Broach's success-path favor bump is dropped (deferred with a TODO comment).
- `set_char_color(AT_DGREEN, ch)` — no direct port. Inline `&g` color in the broach send-strings (pattern already used in `act/cmds2.go:372` et al.), or drop the color (Go default is uncolored). See Open Question 3.

---

## Go Design

The three skills are genuinely independent — no shared infrastructure, no shared helper, no ordering constraint. Per the dispatch prompt's "bundle or split" decision rule, bundle into one plan with three independent task groups. This matches the Tranche-A pattern (6 quick-win items shipped together with no shared infrastructure).

### G1 — DoPounce (the simplest; infrastructure fully in place)

Thinnest skill to ship. `gsnPounce` already resolves at boot. `MultiHit` already short-circuits on it. Registration in `boot.go` at `POS_STANDING`, level 0. Mirror `DoBackstab` structurally — target+weapon+single-hit — plus the extra victim-state gates.

**Omissions vs. C:**
- `is_safe` / `check_attacker` / `check_illegal_pk` — Go hasn't ported these. The C skill leans on them for PK-flagging; Go skills have consistently dropped them. Document the omission; it parallels `DoBackstab` and `DoCircle`.
- `check_char_room`'s "They aren't here" message — already the Go pattern.

**Why structural mirror is the right choice:** `DoBackstab` is 40 LOC; `DoPounce` will be ~55 LOC (extra gates: mount, `victim.Fighting != nil`, `victim.Hit < victim.MaxHit && victim.Position > POS_SLEEPING`, weapon-type `Value[3]` set-membership). `MultiHit` dispatch shape is identical.

### G2 — DoBroach (door-lock skill; mirror DoPick)

Door-lock skill: `findDoor`, exit-info bit check, conditionally clear EX_LOCKED on both sides. `DoPick` is the structural template.

**Decision on the C predicate bug:** Port the *intent*, not the broken expression. Use `CLOSED && LOCKED && !PICKPROOF && can_use_skill` as the positive-success condition — i.e., reject when the door is open, unlocked, pickproof, or the skill check fails. Pattern already established by Tranche B G3 (`leverpos` ported the intent not the buggy C guard). Flag clearly in the doc comment citing C line 3987 and this plan's Open Question 2.

**Omissions vs. C:**
- `adjust_favor(ch, 9, 1)` — Go has no `adjust_favor`. Drop with a one-line TODO pointing at a future deity-favor subsystem.
- `check_room_for_traps(ch, TRAP_PICK | trap_door[pexit->vdir])` — Go has no trap dispatcher. Drop with a one-line TODO.
- `set_char_color(AT_DGREEN, ch)` — see Open Question 3. Default choice: inline `&g` prefix on each `ch.Send` string. Keeps the C visual intent without porting a per-call color-prefix helper.

**Why structural mirror is the right choice:** `DoPick` is 38 LOC; `DoBroach` will be ~50 LOC (same exit-info check pattern, plus the mount gate and the color prefix).

### G3 — DoBloodlet (vampire self-harm; creates room object)

A flavor/self-harm skill. Races/class gate via `isBloodRace`. On success: consume bloodthirst, spawn the OBJ_VNUM_BLOODLET object in the room with per-C field overrides (`timer=1`, `value[1]=6`), self-damage for `level/5`.

**Decision on the C operator-precedence bug:** Port the *intent*, not the broken expression. The gate `!ch.IsNPC() && isBloodRace(ch)` matches the likely authorial intent (non-NPC AND (vampire OR demon)). Flag this in the doc comment citing C line 3521 and this plan's Open Question 1. The alternative — porting verbatim so `isBloodRace` requires BOTH race/class conditions simultaneously — ships a command that is functionally unreachable, which cuts against the "each landed item must cite a C source line" principle in the spirit of making the cited behavior actually manifest.

**Key detail: self-targeting `damage` call.** C calls `damage(ch, ch, ch->level/5, gsn_bloodlet)` — self-damage. Go's `combat.Damage(WorldRef, ch, ch, ch.Level/5, gsn)` works the same way; MultiHit's retcode handling tolerates ch==victim as tested in `combat_test.go` elsewhere. The self-damage must run LAST after the object-spawn because Damage may kill ch (improbable at level `>=11` with `/5` damage but not impossible) — if ch dies mid-function the object spawn should have already happened, matching C source order.

**Object creation detail:**
```go
idx := WorldRef.GetObjIndex(types.OBJ_VNUM_BLOODLET)
if idx == nil {
    util.Bug("DoBloodlet: OBJ_VNUM_BLOODLET (vnum 26) not found in index; skipping spawn")
} else {
    obj := handler.CreateObject(WorldRef, idx, 0)
    obj.Timer = 1
    obj.Value[1] = 6
    handler.ObjToRoom(obj, ch.InRoom)
}
```

The `util.Bug` + skip pattern handles the "area data missing" edge case gracefully (pattern established by other Go loaders). The shipped `db/area/en/limbo.are2:806` has the prototype so production code never hits the Bug path.

### Decision — no shared helper

Three skills, three different infrastructure needs. A shared helper would be artificial. Reject "factor out a `skillGate(ch, ...)` common-prologue" — the gates are 20% overlap at most (NPC+AFF_CHARM gate is common to two of three; mount gate is common to two of three; argument gate is specific to pounce and broach; different error messages in each).

---

## Task Groups

### G1 — DoPounce

**Files:**
- Modified: `internal/act/skills3.go` (new `DoPounce` function after `DoCircle` / `DoGouge` / `DoStun` cluster around `:50-173`)
- Modified: `internal/act/skills3_test.go` (new test block)
- Modified: `internal/boot/boot.go` (one-line registration near `circle` at line 507)

**Test-first (all via `Edit` round-trips only — BANNED for mutation revert: `git checkout -- <file>`, `git checkout <ref> -- <file>`, `git restore <file>`, `git reset --hard` any form, `git stash` any form. Apply the mutation with `Edit`; run test; confirm failure; call `Edit` again with the opposite change to revert. See `_shared.md` → Mutation Verification Safety):**

- `TestDoPounce_MissingArgPromptsTarget` — no arg; assert `"Pounce on whom?\n\r"` sent, no WAIT_STATE set.
- `TestDoPounce_TargetNotFound` — arg names nobody; assert `"They aren't here.\n\r"`.
- `TestDoPounce_SelfTargetRejected` — arg is caster's name; assert `"Pounce on yourself?\n\r"`.
- `TestDoPounce_MountedRejected` — set `ch.Mount != nil`; assert `"You can't get close enough while mounted.\n\r"`.
- `TestDoPounce_NPCCharmedRejected` — NPC with `AFF_CHARM` set; assert early `"You can't do that right now.\n\r"`.
- `TestDoPounce_NoWeaponRejected` — no `WEAR_WIELD`; assert `"You are not wielding an appropriate weapon type to effectively pounce.\n\r"`.
- `TestDoPounce_WrongWeaponTypeRejected` — WEAR_WIELD with `Value[3]=0` (DAM_HIT); same message.
- `TestDoPounce_EachValidWeaponTypeAccepted` — parametric test iterating `{1, 2, 3, 5, 10, 11}` (= `{DAM_SLICE, DAM_STAB, DAM_SLASH, DAM_CLAW, DAM_BITE, DAM_PIERCE}`); each passes the gate (verify no "not wielding an appropriate" message).
- `TestDoPounce_VictimAlreadyFighting` — victim `Fighting != nil`; assert `"You cannot pounce on someone who is in combat.\n\r"`.
- `TestDoPounce_HurtAwakeVictim` — `victim.Hit = victim.MaxHit - 1`, `victim.Position = POS_STANDING`; assert act-string `"$N is hurt and suspicious ... you can't sneak up."` sent to ch.
- `TestDoPounce_SleepingVictimAllowed` — `victim.Hit = victim.MaxHit - 1`, `victim.Position = POS_SLEEPING` (IS_AWAKE false); gate passes, skill resolves.
- `TestDoPounce_SuccessInvokesMultiHit` — stub `numberPercent` to return 1 (guaranteed `canUseSkill` success), install spy on `combat.MultiHit` via an `act`-level seam (see implementation note below), assert it was called with `(WorldRef, ch, victim, gsnPounce)`.
- `TestDoPounce_FailureDoesZeroDamage` — stub `numberPercent` to return 99, install spy on `combat.Damage`, assert called with `(WorldRef, ch, victim, 0, gsnPounce)`.
- `TestDoPounce_WaitStateSet` — after successful invocation, `ch.Wait` is set to `WorldRef.Skills[gsnPounce].Beats` (== 12 per `skills.dat`).

**Implementation note — MultiHit spy:** `combat.MultiHit` is a free function, not a seam. Existing Go skills (`DoBackstab`, `DoCircle`) call it directly and use integration-style tests that check *effects* (victim HP, position, Fighting state) rather than mocking `MultiHit`. Follow the existing convention: assert on the post-state (e.g., `ch.Fighting.Who == victim` after a successful pounce), not on the call. The "invokes MultiHit" wording in `TestDoPounce_SuccessInvokesMultiHit` refers to: set up a clean two-character room, call `DoPounce`, assert `ch.Fighting.Who == victim && victim.Fighting.Who == ch` (MultiHit → OneHit → Damage → StartFighting flow). If a finer-grained seam is needed, extract `combat.MultiHit` into a package-level var like `rollD20` — out of scope for this plan.

**Mutation-verify (via `Edit` round-trips):**
- Change weapon-type check from `{1,2,3,5,10,11}` to `{1,2,3,5,10}` → `TestDoPounce_EachValidWeaponTypeAccepted` fails on `DAM_PIERCE=11`.
- Invert `victim.Hit < victim.MaxHit` (to `<=`) → `TestDoPounce_HurtAwakeVictim` detection flips for boundary cases.
- Drop the `IS_AWAKE` check (remove `Position > POS_SLEEPING` conjunct) → `TestDoPounce_SleepingVictimAllowed` fails (gate rejects sleeping victim).
- Swap `learnFromSuccess` and `learnFromFailure` branches → either the success-side Learned counter fails to increment on passed gates or the failure-side does.

**Acceptance gate:** `go test -count=3 ./internal/act/...` green; `go vet ./...` clean; manually verified via `./smaug-go -port 4000`: log in at level 5, `pounce <mob>`, see attack initiate.

### G2 — DoBroach

**Files:**
- Modified: `internal/act/skills.go` (new `DoBroach` function after `DoPick` at `:441-479`)
- Modified: `internal/act/skills_test.go` (new test block)
- Modified: `internal/boot/boot.go` (one-line registration near `pick` at line 496)

**Test-first (mutation-verification: `Edit` round-trips only, same banned list as G1):**

- `TestDoBroach_MissingArgPromptsDirection` — no arg; assert `"Attempt this in which direction?\n\r"`.
- `TestDoBroach_MountedRejected` — `ch.Mount != nil`; assert `"You should really dismount first.\n\r"`.
- `TestDoBroach_NPCCharmedRejected` — NPC+AFF_CHARM; assert `"You can't concentrate enough for that.\n\r"` (per C :3969 wording).
- `TestDoBroach_NoDoor` — arg that matches no exit; assert `"Your attempt fails.\n\r"` (matches C :4009 fallthrough).
- `TestDoBroach_DoorNotClosed` — exit without `EX_CLOSED`; with stubbed `numberPercent=1` (skill pass); assert the success branch does NOT fire (door stays unlocked but no "You successfully broach" message).
- `TestDoBroach_DoorNotLocked` — exit with `EX_CLOSED` but not `EX_LOCKED`; same as above (fail branch).
- `TestDoBroach_DoorPickproof` — exit with `EX_CLOSED | EX_LOCKED | EX_PICKPROOF`; fail branch with `"Your attempt fails.\n\r"`, no lock removed, `learnFromFailure` called.
- `TestDoBroach_FailedSkillCheck` — closed+locked+!pickproof, stubbed `numberPercent=99` (skill fails); fail branch; door stays locked.
- `TestDoBroach_SuccessRemovesLockBothSides` — closed+locked+!pickproof, numberPercent=1; door's `EX_LOCKED` bit cleared; reverse-exit's `EX_LOCKED` bit also cleared (if reverse exit points back to ch's room); success send `"You successfully broach the exit...\n\r"`; `learnFromSuccess` called.
- `TestDoBroach_SuccessOneSidedWhenReverseDoesNotReciprocate` — door whose reverse-exit's `ToRoom` is NOT ch's room; only the forward lock removed; no crash.
- `TestDoBroach_WaitStateSet` — after invocation (success or fail branch), `ch.Wait` is `WorldRef.Skills[gsnBroach].Beats` (== 24 per `skills.dat`).

**Green-prefix decision (per Open Question 3):** if inline-`&g` is chosen (default), assert the success message begins with `&g` and the fail message also does. Omit those assertions if Option B (no color) is chosen instead — document the decision in the plan completion record.

**Mutation-verify (via `Edit` round-trips):**
- Swap the success predicate from `CLOSED && LOCKED && !PICKPROOF && can_use` to `CLOSED || LOCKED || PICKPROOF || !can_use` → `TestDoBroach_DoorNotClosed` should now "succeed" incorrectly; fails.
- Remove the reverse-exit lock-removal → `TestDoBroach_SuccessRemovesLockBothSides` asserting-reverse-exit-cleared fails.
- Remove the `pexit_rev.ToRoom == ch.InRoom` guard → `TestDoBroach_SuccessOneSidedWhenReverseDoesNotReciprocate` would crash or incorrectly modify a distant exit.
- Drop the `WAIT_STATE` unconditional call → `TestDoBroach_WaitStateSet` fails for the fail branch.

**Acceptance gate:** `go test -count=3 ./internal/act/...` green; `go vet ./...` clean; manually verified: log in at level 41+, lock a door with `redit`, `broach <dir>`, see the lock break after skill-check passes.

### G3 — DoBloodlet

**Files:**
- Modified: `internal/act/skills4.go` (new `DoBloodlet` function + new `isBloodRace` helper, placed near `DoFeed` at `:240`)
- Modified: `internal/act/skills4_test.go` (new test block)
- Modified: `internal/boot/boot.go` (one-line registration near `feed` at line 523)

**Test-first (mutation-verification: `Edit` round-trips only, same banned list as G1):**

- `TestIsBloodRace_NPC` — NPC (with Act.ACT_IS_NPC set): returns false regardless of race/class.
- `TestIsBloodRace_VampireRace` — PC, `Race = RACE_VAMPIRE`: returns true.
- `TestIsBloodRace_VampireClass` — PC, `Class = CLASS_VAMPIRE` (=4): returns true.
- `TestIsBloodRace_DemonRace` — PC, `Race = RACE_DEMON`: returns true.
- `TestIsBloodRace_DemonClass` — PC, `Class = CLASS_DEMON` (=13): returns true.
- `TestIsBloodRace_Plain` — PC, `Race = RACE_HUMAN`, `Class = CLASS_WARRIOR`: returns false.
- `TestDoBloodlet_NPCSilentNoop` — NPC caller; command silently returns (no error message, no damage) — matches C early-return with no user message.
- `TestDoBloodlet_NonBloodRaceSilentNoop` — PC human/warrior; silent early-return (no `learnFromFailure`, no damage).
- `TestDoBloodlet_BlockedIfFighting` — vampire PC with `ch.Fighting != nil`; assert `"You're too busy fighting...\n\r"`; no damage, no object spawn.
- `TestDoBloodlet_BlockedIfBloodthirstLow` — vampire PC, `PCData.Condition[COND_BLOODTHIRST] = 9` (< 10); assert `"You are too drained to offer any blood...\n\r"`; no damage, no object spawn.
- `TestDoBloodlet_SuccessSpawnsObject` — vampire PC, bloodthirst=20, stubbed numberPercent=1 (`can_use_skill` passes); assert new obj in `ch.InRoom.Contents` with `IndexData.Vnum == 26`, `Timer == 1`, `Value[1] == 6`.
- `TestDoBloodlet_SuccessDecrementsBloodthirst` — after success, `ch.PCData.Condition[COND_BLOODTHIRST] == 13` (was 20, -7).
- `TestDoBloodlet_SuccessSelfDamages` — ch.Level=20, ch.Hit=100; after success, `ch.Hit == 100 - 4` (level/5=4). (Verify no crash if self-damage would kill; set ch.Hit=1 → expect ch transitions to dead/dying and object STILL spawns.)
- `TestDoBloodlet_FailureNoDamageNoObject` — stubbed numberPercent=99 (skill fails); assert fail-path messages sent, no damage taken, no object in room, `learnFromFailure` invoked.
- `TestDoBloodlet_MissingObjIndexLogsBugNoSpawn` — force `WorldRef.ObjIndex[26] = nil` before calling; skill resolves, logs `util.Bug`, no panic, no object; ch still takes self-damage and loses bloodthirst.
- `TestDoBloodlet_WaitStateSet` — after invocation (success or fail branch), `ch.Wait = types.PULSE_VIOLENCE` (matches C `WAIT_STATE(ch, PULSE_VIOLENCE)` — note: not the `skills.dat` Rounds value; bloodlet is special per C :3535).

**Implementation note on `ch.Wait`:** C bloodlet uses `WAIT_STATE(ch, PULSE_VIOLENCE)` — a literal `PULSE_VIOLENCE` (~3 seconds), NOT `skill_table[gsn_bloodlet]->beats` (which is 12 per skills.dat, ~12 rounds = 36s). This divergence from the `skills.dat` metadata is intentional in C. Port verbatim: `ch.Wait = types.PULSE_VIOLENCE`.

**Mutation-verify (via `Edit` round-trips):**
- Remove the `bloodthirst < 10` guard → `TestDoBloodlet_BlockedIfBloodthirstLow` fails (skill proceeds instead of blocking).
- Change `-7` bloodthirst decrement to `-6` → `TestDoBloodlet_SuccessDecrementsBloodthirst` fails.
- Drop the object-spawn → `TestDoBloodlet_SuccessSpawnsObject` fails.
- Swap self-damage formula `ch.Level/5` to `ch.Level/4` → `TestDoBloodlet_SuccessSelfDamages` fails.
- Use `PULSE_TICK` instead of `PULSE_VIOLENCE` for Wait → `TestDoBloodlet_WaitStateSet` fails.

**Acceptance gate:** `go test -count=3 ./internal/act/...` green; `go vet ./...` clean; manually verified: log in as a vampire (or set `Race = RACE_VAMPIRE` via immortal command), bloodthirst high, `bloodlet`, see blood pool drop and HP loss.

---

## Acceptance Criteria

**G1 (DoPounce):**
- A1. `pounce <target>` command is registered at `POS_STANDING`, level 0, dispatching to `act.DoPounce`.
- A2. All 6 valid weapon damage-types (SLICE/STAB/SLASH/CLAW/BITE/PIERCE per C :2650-2654) accepted; all others rejected with the "not wielding an appropriate weapon type" message.
- A3. Successful invocation starts combat (`ch.Fighting.Who == victim` post-call) and sets `ch.Wait = WorldRef.Skills[gsnPounce].Beats` (= 12).
- A4. `MultiHit`'s single-hit short-circuit for `gsnPounce` (already in place at `combat.go:247`) fires once per successful invocation — verified by asserting ch's second-attack cascade does NOT run (stub `numberPercent` such that second_attack would succeed if entered; assert only one hit landed).

**G2 (DoBroach):**
- A5. `broach <direction>` command is registered at `POS_STANDING`, level 0, dispatching to `act.DoBroach`.
- A6. Success path (door is CLOSED && LOCKED && !PICKPROOF && skill-check passes) clears `EX_LOCKED` on both the forward exit AND its reverse-exit (when the reverse points back to ch's room).
- A7. All four fail conditions (door open / door unlocked / door pickproof / skill-check fails) reject the attempt and leave `EX_LOCKED` intact, sending `"Your attempt fails.\n\r"`.
- A8. `ch.Wait = WorldRef.Skills[gsnBroach].Beats` (= 24) is set unconditionally before the find-door branch.

**G3 (DoBloodlet):**
- A9. `bloodlet` command is registered at `POS_RESTING`, level 0, dispatching to `act.DoBloodlet`.
- A10. NPC callers and non-blood-race PC callers silently early-return (no error message, no side effects).
- A11. On success: `OBJ_VNUM_BLOODLET` (vnum 26) spawns in ch's room with `Timer=1, Value[1]=6`; ch's `COND_BLOODTHIRST` decreases by 7; ch self-damages for `Level/5` HP.
- A12. `ch.Wait = PULSE_VIOLENCE` (NOT `Beats=12` — matches C :3535 literal).
- A13. Missing `ObjIndex[26]` logs `util.Bug` and skips object spawn without crashing.

**Global:**
- A14. `go vet ./...` clean.
- A15. `go test -count=3 ./...` green across all packages.
- A16. `make install-hooks`'s pre-commit gate (gofmt + vet + unit tests + conflict markers) passes.

---

## Scope Cuts / Deferrals

- **`is_safe` / `check_attacker` / `check_illegal_pk` PK-legality helpers** — not ported in Go; all existing combat skills (backstab/circle) drop these too. Track in TODO as "PK-legality subsystem port" (would cover `illegal_pk` tracking, `ACT_NICE`/`PLR_NICE` interaction, area `AFLAG_NOPKILL`). Out of scope for this plan.
- **`adjust_favor`** — deity-favor subsystem not yet ported (deity prayer is a separate Phase 6 plan, `plan-phase6-deity-prayer.md`). Broach's favor bump drops with a TODO comment; re-add when the deity subsystem lands.
- **`check_room_for_traps`** — trap subsystem not ported (`TRAP_PICK` constant exists but no dispatcher). Broach's trap-trigger calls drop. No known ported skill triggers traps; this is a standalone follow-up.
- **Per-AT color preservation for bloodlet's `AT_BLOOD` / broach's `AT_DGREEN`** — the `util.atColorCode` map currently has no entry for `AT_BLOOD` or `AT_DGREEN`. `util.Act(types.AT_BLOOD, ...)` will send uncolored text (safe default). Bloodlet messages will therefore render without the C's blood-red coloring. Broach uses `set_char_color` not `util.Act`; inline `&g` prefix chosen to preserve green intent. Track as follow-up: "Extend `util.atColorCode` to include `AT_BLOOD` and `AT_DGREEN`" — belongs to the existing TODO.md item "Extend `internal/util/act.go:atColorCode` table to additional AT_* codes".
- **Verbatim preservation of C operator-precedence bug in bloodlet** — plan fixes the bug (interprets intent). If human review prefers verbatim fidelity, revert to `!ch.IsNPC() && isVampire(ch) && isDemon(ch)` (ships an unreachable command). See Open Question 1.
- **Verbatim preservation of C tautological guard in broach** — plan fixes the guard (uses `CLOSED && LOCKED && !PICKPROOF`). If human review prefers verbatim fidelity, ship the negated-tautology form. See Open Question 2.
- **Spell-`gsn_bloodlet` wiring in damage messages** — C's `damage` call passes `gsn_bloodlet` as the dt argument; damage-message lookup reads `skill_table[gsn_bloodlet]->noun_damage` (which is empty per skills.dat). Go will see empty and fall through to default message. Noted; no fix needed.
- **`do_stset` / `do_ststat` / other skill-metadata admin commands** — out of Phase 6 scope per `plan-tranche-b.md` G1b. Bloodlet/pounce/broach are player-facing commands, no admin surface added here.

---

## Open Questions (with recommended answers)

### Q1 — Porting policy for bloodlet's operator-precedence bug

**The question:** C's gate at `src/skills.c:3521` reads `if (IS_NPC (ch) || !IS_VAMPIRE (ch) || !IS_DEMON (ch)) return;`. Given De Morgan, `!A || !B` is `!(A && B)`, so this gates on "NPC OR not-BOTH-vampire-AND-demon". Audit correction 2026-04-18: because `IS_VAMPIRE(ch) = !IS_NPC && (race==RACE_VAMPIRE || class==CLASS_VAMPIRE)` and `IS_DEMON(ch)` is the parallel form, a character CAN hold both predicates simultaneously if race is one blood identity and class is the other (e.g., race=RACE_VAMPIRE + class=CLASS_DEMON, or race=RACE_DEMON + class=CLASS_VAMPIRE). So the command is NOT strictly unreachable — it's reachable for the narrow dual-identity edge case but unreachable for all typical single-identity vampire/demon characters. This still strongly suggests a bug (Option A is still recommended) but the plan's original "unreachable" framing is imprecise.

Options:
- **Option A (RECOMMENDED): Fix the bug.** Gate on `!ch.IsNPC() && (isVampire(ch) || isDemon(ch))`. Ship a working command. Document in a code comment citing C :3521 and this Open Question. Precedent: Tranche B G3 (leverpos fixed the buggy C guard); Holidays plan (fixed 2 latent C bugs). Rationale: the project's "no gameplay invented" rule does not require shipping unreachable dead code; it requires every behavior to have a C citation. Ported intent IS a citation.
- **Option B: Port verbatim.** Gate on the broken `!A || !B` form; ship a dead command. Rationale: strict fidelity. Downside: every future player who `help bloodlet` and tries to use it gets no feedback (silent early-return); the whole tranche's value is zero.
- **Option C: Port verbatim behind a compile-time flag.** `go:build` tag `c_bug_compat` to ship the verbatim form for auditing; default build ships fixed form. Downside: complexity without user benefit.

**Recommendation: Option A.** The bug is demonstrably unintended (every use site of `IS_VAMPIRE`/`IS_DEMON` elsewhere in C uses `||` explicitly — `src/update.c:171, :354, :463`, `src/player.c:454, :1368, :1922, :3484, :3539`). The authorial intent is unambiguous.

### Q2 — Porting policy for broach's tautological predicate

**The question:** C's fail-branch at `src/skills.c:3987-3990` reads `if (!IS_SET(pexit->exit_info, EX_CLOSED) || !IS_SET(pexit->exit_info, EX_LOCKED) || IS_SET(pexit->exit_info, EX_PICKPROOF) || can_use_skill(ch, number_percent(), gsn_broach))`. The two `!IS_SET` disjuncts are almost certainly wrong — the intent is "fail if the door is open OR unlocked OR pickproof OR skill fails". But `can_use_skill` returning true (success) is ALSO wired into the fail branch, which is plainly backwards.

Options:
- **Option A (RECOMMENDED): Fix the predicate.** Port the intent as `success if CLOSED && LOCKED && !PICKPROOF && can_use_skill`; fail otherwise. Document in a code comment. Rationale: the C form is provably non-sense (using `||` between positive and negative conditions with `can_use_skill` on the wrong side); the authorial intent is reconstructible from the branch structure.
- **Option B: Port verbatim.** Ship the broken predicate. Downside: broach will behave erratically — sometimes succeeding when it should fail, sometimes failing when it should succeed, depending on door state. Users will experience the command as flaky.

**Recommendation: Option A.** Same precedent as Q1. The fix is a judgment call but the only reasonable reading of the branch structure.

### Q3 — Color handling for broach's `set_char_color(AT_DGREEN, ch)`

**The question:** C calls `set_char_color(AT_DGREEN, ch)` before every `send_to_char` — this sets a per-descriptor "current color" that subsequent sends inherit. Go has no per-descriptor color state; `ch.Send` sends uncolored unless inline `&`-codes are embedded.

Options:
- **Option A (RECOMMENDED): Inline `&g` prefix** on every `ch.Send` call in `DoBroach`. Example: `ch.Send("&gYour attempt fails.\n\r")`. Pattern already used in `internal/act/cmds2.go:372` et al. Preserves the green-color user experience.
- **Option B: Drop the color** — send uncolored. Rationale: defer the color port until `util.atColorCode` gains `AT_DGREEN`. Downside: users familiar with the C MUD see broach messages in a different color (or no color).
- **Option C: Extend `util.atColorCode`** with `AT_DGREEN = "&g"` and use `util.Act(types.AT_DGREEN, ...)`. Downside: `util.Act` is for formatted actor/target/room messages — broach's sends are simple direct sends, not Act-formatted. Misuse of the helper.

**Recommendation: Option A.** Scope-appropriate; matches existing Go patterns for color; no new infrastructure needed. Also applies to `DoBloodlet`'s `AT_BLOOD` sends (use `&r` — dark red — inline). Alternative for bloodlet: use `util.Act(types.AT_BLOOD, ...)` which matches C's `act (AT_BLOOD, ...)` structurally; the color will drop through silently (AT_BLOOD not in atColorCode map) but the message format (actor/victim substitution) is preserved. Recommendation for bloodlet specifically: Option-A-via-util.Act — call `util.Act(types.AT_BLOOD, ...)` for all bloodlet messages (C uses `act()` not `send_to_char` here); when `atColorCode` gains an `AT_BLOOD` entry later, color appears automatically. See `plan-tranche-c.md` G1 for precedent.

### Q4 — Registration position for bloodlet

**The question:** Bloodlet's `Minpos=105` converts to `POS_RESTING` in Go. Should the command registration use `POS_RESTING` or `POS_STANDING`?

- **Option A (RECOMMENDED): POS_RESTING.** Matches the skill-metadata MinimumPos (after conversion). C's position gate is enforced by the command interpreter before dispatching to `do_bloodlet`, so the skill function body doesn't re-check position. Go registration at `POS_RESTING` matches C dispatch.
- **Option B: POS_STANDING.** Matches the "most combat/skill commands" default.

**Recommendation: Option A.** The skill metadata is authoritative. Mirrors how other skills match their Minpos values.

### Q5 — Registration minlevel

**The question:** `skills.dat` Minlevel for each skill is bloodlet=11, pounce=5, broach=41. Boot registration's `Level` field currently defaults to 0 for skill commands (`DoBackstab` is Level 0). The player-facing level gate happens via the class/race skill-learning system (only some classes get the skill at level X). Should registration override this?

- **Option A (RECOMMENDED): Level 0.** All three register at Level 0 like every other skill. The actual "you can't use this yet" gate comes from `PCData.Learned[gsn] == 0` (i.e., unlearned) returning false from `canUseSkill`. This is how Go's existing skill dispatch works.
- **Option B: Use `skills.dat` Minlevel.** Set `Level: 41` for broach et al. Downside: creates a second-tier gate that doesn't exist in C — C's Minlevel is the *Learn* gate, not the *use* gate.

**Recommendation: Option A.** Fidelity with existing Go skills.

---

## Risk Analysis

**Low risk** overall. Self-contained; no shared types; no cross-package changes; all primitives already shipped. Three parallel-safe work units.

**Specific risk points:**

1. **Bloodlet's self-damage → character death mid-function.** `combat.Damage(w, ch, ch, ch.Level/5, gsn)` could theoretically kill ch. Object spawn runs before self-damage, so the object exists in the room even if ch dies; fight.c path handles `ch == victim` case. Mitigation: tests include the edge case (ch.Hit=1 pre-damage) to confirm no panic and that the object spawns anyway.

2. **Pounce's success test relies on integration-level assertions** (ch.Fighting state after MultiHit), not a mocked `MultiHit` seam. If MultiHit's semantics change, tests may fail for reasons unrelated to DoPounce. Mitigation: tests scope post-conditions to the DoPounce-specific effects (Wait set, victim Fighting state, no cascade when gsnPounce short-circuits) — not MultiHit's internal cascade behavior.

3. **Broach's fix-the-C-bug decision** (Open Question 2) affects every test in G2. If Open Question 2 resolves to Option B (ship verbatim), tests must be rewritten. Mitigation: write tests against the *recommended* (fixed) behavior; if human reviewer demands verbatim, revise tests and prose together.

4. **Bloodlet's operator-precedence fix** (Open Question 1) affects whether the command is reachable at all. If Option B chosen, most tests become "unreachable-command" no-ops. Mitigation: same as risk 3 — proceed with recommended Option A; revise on explicit request.

5. **`AT_BLOOD` / `AT_DGREEN` color absence.** Default uncolored output for bloodlet if Option A-via-util.Act is chosen and `atColorCode` isn't extended yet. Users see the messages but without the intended red/green tint. Low risk; cosmetic.

6. **Unported `is_safe` gate on pounce means PC-vs-PC pounce has no legality check.** Matches existing Go convention (no skill in `internal/act/skills*.go` does PK-legality checks). Consistency argues for deferring; inconsistency would be the bigger risk.

7. **Broach's lack of trap-dispatch** means a locked door booby-trapped with a pick-trap won't trigger on broach attempts. Matches `DoPick` (no trap dispatch in Go). Track as shared follow-up.

8. **Object extraction on timer=1.** `OBJ_VNUM_BLOODLET` has `timer=1` meaning it decays after one tick. Go's object-timer tick already ships in `world/update.go` equivalent (or equivalent timer pulse — quick verify needed but out of scope; if the timer tick isn't running, the object becomes a permanent litter). Low risk: if tick is broken, it's a separate bug (affects every timed object).

---

## Adversary Verification Notes

*To be filled in after plan adversary pass.*

**Self-review completed 2026-04-18:**
- Plan cites every C source line referenced (skills.c:3517/2607/3960; mud.h:4029-4034/2048; skills.dat lines verified via `sed`).
- Every Go helper referenced has been verified to exist (see § Dependent macros / helpers checklist).
- Every claim about existing shipped Go state cross-referenced (e.g., `gsnPounce` line, `OBJ_VNUM_BLOODLET` constant, object prototype in limbo.are2).
- Open Questions 1-2 flag C source bugs with explicit port-intent-not-verbatim recommendations and the precedent (Tranche B G3, Holidays).
- Mutation-verify lists concrete mutations, not generic "try changing something" hand-waves.
- Scope cuts enumerate every C call that won't port (is_safe, check_attacker, check_illegal_pk, adjust_favor, check_room_for_traps, set_char_color per-descriptor, AT_BLOOD/AT_DGREEN colorization).
- Manager-harness `Agent` tool absence noted — external adversary pass is queued per Wave-A Phase 6 plan pattern.

**External adversary follow-up (queued):**
- Run 1-3 independent adversaries on this plan in a future session where `Agent` tool is available.
- Particular attention to: (a) Open-Question-1 recommendation (are we wrong to interpret the C bug as unintended?), (b) Open-Question-2 recommendation (is the broach predicate really reconstructible from branch structure?), (c) risks 1 and 3 (self-damage edge cases, test-assertion scope).

---

## Completion Record

*To be filled in after work lands.*
