# Plan — Damage-message gaps

Source audit: `audit-2026-04-17.md`, TODO ref: `TODO.md:71-76`, plus inline TODOs in `combat/dammessage.go:150,154,157`.

## Goal

Close the three small-but-visible gaps in `combat/dammessage.go` vs C `fight.c`'s `dam_message`:

1. **`was_in_room` swap** (`fight.c:4432-4439, 4590-4594`) — attacker's `InRoom` is temporarily moved to the victim's room so bystanders there see buf1. Currently Go bails with a "TODO" and emits only to the victim.
2. **`PCFLAG_GAG` self-suppress** (`fight.c:4481-4486`) — a gagged *attacker* doesn't see their own zero-damage miss; a gagged *victim* doesn't see someone missing them. Bystanders always see it (TO_NOTVICT never gated). Damage > 0 is never silenced.
3. **`is_wielding_poisoned` prefix** (`fight.c:4496-4512`) — weapon attacks with a poisoned wielded weapon get a "poisoned" prefix in the buffer formatting.

All three land as one tight commit — same file, same test pattern, ~1.5 days total.

## Exact C blocks (for reference)

Room swap (fight.c:4432-4439):

```c
if (ch->in_room != victim->in_room) {
    was_in_room = ch->in_room;
    char_from_room (ch);
    char_to_room (ch, victim->in_room);
} else
    was_in_room = NULL;
```

Restore (fight.c:4590-4594):

```c
if (was_in_room) {
    char_from_room (ch);
    char_to_room (ch, was_in_room);
}
```

GAG (fight.c:4481-4488):

```c
if (dam == 0 && (!IS_NPC(ch) && IS_SET(ch->pcdata->flags, PCFLAG_GAG)))
    gcflag = TRUE;
if (dam == 0 && (!IS_NPC(victim) && IS_SET(victim->pcdata->flags, PCFLAG_GAG)))
    gvflag = TRUE;
...
act (AT_ACTION, buf1, ch, NULL, victim, TO_NOTVICT);
if (!gcflag) act (AT_HIT,   buf2, ch, NULL, victim, TO_CHAR);
if (!gvflag) act (AT_HITME, buf3, ch, NULL, victim, TO_VICT);
```

Poisoned prefix (fight.c:4496-4512):

```c
else if (dt > TYPE_HIT && is_wielding_poisoned(ch)) {
    attack = attack_table[dt - TYPE_HIT];
    sprintf(buf1, "$n's poisoned %s %s $N%c",  attack, vp, punct);
    sprintf(buf2, "Your poisoned %s %s $N%c",  attack, vp, punct);
    sprintf(buf3, "$n's poisoned %s %s you%c", attack, vp, punct);
}
```

`is_wielding_poisoned` (fight.c:104-119) checks `used_weapon == eq[WEAR_WIELD] || == eq[WEAR_DUAL_WIELD]` and `IS_OBJ_STAT(obj, ITEM_POISONED)`. Go's `DamMessage` already receives the active weapon as `obj` — `used_weapon` indirection is not needed.

## Current Go state

`/home/eilidh/src/smaug/smaug-go/internal/combat/dammessage.go`:

- Top-of-function TODO block at `:150-157` calls out all three gaps.
- `:232` — `crossRoom := ch.InRoom != victim.InRoom`.
- `:297-301` — "TODO: port was_in_room swap. For now, emit only to victim." → emits `TO_VICT` and returns. This test is locked in via `TestDamMessageDifferentRoomsEmitsToVictOnly` in `dammessage_test.go` (must be flipped once G1 lands).
- `:303-305` — the three `util.Act(bufN, ch, victim, nil, nil, type)` broadcast calls. GAG suppression must wrap the TO_CHAR/TO_VICT calls here.
- Broadcast does NOT have an inline `for _, vch := range room.People` loop — delegated to `util.Act` (`util/act.go:283-318`). GAG lives in `DamMessage`, not `util.Act`, because C's GAG is self-suppress, not a universal recipient filter.

No existing GAG or poisoned-weapon test coverage.

## Missing prerequisites

**G1 (room swap):**

- `CharData.InRoom *RoomIndexData` — swappable freely.
- `handler.CharFromRoom` / `handler.CharToRoom` exist and are used from `act/skills.go:499-501` — no cycle risk adding the import to `combat/`.
- Restore-on-panic: use `defer` so the restore runs even if `util.Act` panics. Existing behavior in `util.Act` guards all current panic surfaces, but `defer` is future-proof.

**G2 (GAG):**

- `PCFLAG_GAG` constant defined at `types/constants.go:652` (`1 << 5`).
- `CharData.PCData.Flags int` (`types/pcdata.go:38`).
- No `DoGag` player command exists in Go yet. Tests set the flag directly; a follow-up task can port `do_gag` (`act_info.c:5795`) separately.

**G3 (poisoned prefix):**

- `ITEM_POISONED` flag defined at `types/enums.go:674`, set by `DoEnvenom` at `act/skills4.go:326`.
- `handler.GetEqChar` exists; `WEAR_WIELD = 16`, `WEAR_DUAL_WIELD = 18` defined.
- **C semantics (verified `fight.c:104-119`):** `is_wielding_poisoned` checks that `used_weapon` is identical to `get_eq_char(ch, WEAR_WIELD)` OR `get_eq_char(ch, WEAR_DUAL_WIELD)` AND the weapon has `ITEM_POISONED`. The identity check matters.
- **Go caveat:** `combat.go:162` inside `OneHit` only fetches `WEAR_WIELD`, so the `obj` threaded into `DamMessage` on a dual-wield **extra** hit is the primary wield, not the offhand. A naive `isWieldingPoisoned(obj)` would misread the dual-wield case. Two fixes:
  - (a) Match C exactly: `isWieldingPoisoned(ch, obj *ObjData) bool` — returns true iff `obj != nil && obj.ExtraFlags.IsSet(ITEM_POISONED) && (obj == handler.GetEqChar(ch, WEAR_WIELD) || obj == handler.GetEqChar(ch, WEAR_DUAL_WIELD))`. Safest; tolerates the combat-depth plan's G8 weapon-alternation fix.
  - (b) Tighten `OneHit` signature in combat-depth G8 to thread the actually-striking weapon in, then `isWieldingPoisoned` can take just `obj`.
- **Recommendation:** use (a) here. It's self-contained to this plan and doesn't depend on combat-depth G8 landing first.

## Task groups

### G1 — Room swap port (S, ~0.5 day)

Replace the crossRoom short-circuit at `:232, 297-301` with a C-fidelity swap using `handler.CharFromRoom` / `handler.CharToRoom`. Wrap restore in `defer`.

```go
if ch.InRoom != victim.InRoom {
    wasInRoom := ch.InRoom
    handler.CharFromRoom(ch)
    handler.CharToRoom(ch, victim.InRoom)
    defer func() {
        handler.CharFromRoom(ch)
        handler.CharToRoom(ch, wasInRoom)
    }()
}
```

Then fall through to the unified broadcast (G2 follows).

- **Tests:**
  - **Delete** the existing `TestDamMessageDifferentRoomsEmitsToVictOnly` at `dammessage_test.go:296` — it locks in the current broken behavior. Replace (not rename) with `TestDamMessage_DifferentRooms_SwapsAttackerIntoVictimRoom` with a fresh body.
  - Attacker in room A, victim + bystander in room B → bystander sees buf1.
  - After call, `ch.InRoom == A` (restored); room A's `People` contains `ch`; room B's does not.
  - **Defer-restore test (without mocking `util.Act`):** construct a bystander whose `Desc.Conn` panics on write (custom `net.Conn` whose `Write` returns an error that `util.Act` propagates as panic). Assert `ch.InRoom == A` after `DamMessage` returns. Skip this test if a simpler descriptor-error path doesn't panic; the `defer` is correctness insurance either way.

### G2 — PCFLAG_GAG self-suppress (S, ~0.5 day)

Compute locally (match C exactly):

```go
gcflag := dam == 0 && !ch.IsNPC() && ch.PCData != nil &&
          (ch.PCData.Flags & int(types.PCFLAG_GAG)) != 0
gvflag := dam == 0 && !victim.IsNPC() && victim.PCData != nil &&
          (victim.PCData.Flags & int(types.PCFLAG_GAG)) != 0
```

Then at the existing broadcast site:

```go
util.Act(buf1, ch, nil, victim, types.TO_NOTVICT)
if !gcflag { util.Act(buf2, ch, nil, victim, types.TO_CHAR) }
if !gvflag { util.Act(buf3, ch, nil, victim, types.TO_VICT) }
```

- **Tests:**
  - Attacker-gagged, `dam=0`, `dt=TYPE_HIT`: no TO_CHAR output; TO_VICT + TO_NOTVICT still deliver.
  - Victim-gagged mirror case: no TO_VICT output; TO_CHAR + TO_NOTVICT still deliver.
  - Both gagged: only TO_NOTVICT delivers.
  - `dam=1` with same flags set: all three deliver (flag only silences zero-dam misses).
  - NPC attacker/victim with `PCData == nil`: no crash.

### G3 — Poisoned weapon prefix (S, ~0.5 day)

Add `isWieldingPoisoned(ch *CharData, obj *ObjData) bool` (package-private, top of `dammessage.go` or new `combat/poison.go`):

```go
func isWieldingPoisoned(ch *types.CharData, obj *types.ObjData) bool {
    if obj == nil || !obj.ExtraFlags.IsSet(types.ITEM_POISONED) {
        return false
    }
    return obj == handler.GetEqChar(ch, types.WEAR_WIELD) ||
           obj == handler.GetEqChar(ch, types.WEAR_DUAL_WIELD)
}
```

Insert a new branch in `DamMessage`'s format switch between `TYPE_HIT` and the skill branch (aligning with C's `dt > TYPE_HIT && is_wielding_poisoned` at `:4496`). When active, buf1/buf2/buf3 become "poisoned %s" instead of plain "%s" (e.g., "Your poisoned slice mauls Bob!").

- **Tests:**
  - Wielded `ObjData{ExtraFlags: ITEM_POISONED}` in `WEAR_WIELD`, `dt = TYPE_HIT + DAM_SLICE`: TO_CHAR contains "poisoned slice".
  - Same weapon in `WEAR_DUAL_WIELD` (primary slot empty), `obj` threaded through `DamMessage` is the offhand: TO_CHAR contains "poisoned slice".
  - `obj` is a poisoned weapon but NOT equipped in either wield slot (constructed ad-hoc): TO_CHAR does NOT contain "poisoned" (identity check).
  - Same weapon without `ITEM_POISONED`: TO_CHAR does NOT contain "poisoned".
  - `dt == TYPE_HIT` (bare hands) with poisoned obj: output does NOT contain "poisoned" (guard on `dt > TYPE_HIT`).
  - Skill path (`dt` is a valid gsn): poisoned branch skipped; existing MissChar/HitChar formatting wins.
  - `obj == nil`: poisoned branch skipped.

## Acceptance criteria

1. **G1:** attacker in room A, victim in room B with a third bystander C in B. After `DamMessage`:
   - C's captured output contains `buf1` text.
   - `ch.InRoom == A` (restored).
   - `A.People` contains `ch`, `B.People` does not.
   - Passes even when a mocked `util.Act` recipient panics.
2. **G2:** three assertions over `dam=0, dt=TYPE_HIT`:
   - (i) attacker with `PCData.Flags |= PCFLAG_GAG` has empty TO_CHAR; non-gagged victim + bystander still deliver.
   - (ii) gagged victim mirror.
   - (iii) with `dam=1`, all three deliver regardless of flag state.
3. **G3:** with `wield.ExtraFlags.IsSet(ITEM_POISONED)` and `dt = TYPE_HIT + DAM_SLICE`, TO_CHAR contains "poisoned slice". Without the flag, it doesn't. With `dt = TYPE_HIT` bare-hands, poisoned prefix is absent even if flag is set. With `obj == nil`, no prefix.
4. **Regression:** `go test ./internal/combat/...` green; existing 11 `dammessage_test.go` cases unchanged except the renamed swap test.

## Open questions

- **`DoGag` command:** not required to land these — the flag is a mutable `int` and tests set it directly. Porting `do_gag` (~15 LOC) is a trivial follow-up for player-facing UX, not a blocker.
- **testclient harness for dam-message assertions:** overkill for unit-level gaps. The existing `dammessage_test.go` pipe-backed descriptor pattern is the right level. testclient becomes useful when we add a regression that plays `gag` toggle + combat round and asserts observable lines.
- **Cross-package import:** `combat/combat.go:7` already imports `handler/` — so `dammessage.go` using `handler.CharFromRoom` / `handler.CharToRoom` / `handler.GetEqChar` adds no new cycle edge. Verified.
- **Color-code loss:** C's `AT_ACTION` / `AT_HIT` / `AT_HITME` per-recipient coloring is flattened by Go's `util.Act` today. Noted in audit but **out of scope here** — all three gaps are about message content/audience, not color.

## Rough total effort

~1.5 days engineering + tests. All three pair naturally with the combat-depth plan (same package, same test infra) — strong candidate for the same PR or an adjacent commit in the same sub-phase.

## Relevant file paths

- C: `/home/eilidh/src/smaug/src/fight.c` (100-119, 4410-4596)
- Go: `/home/eilidh/src/smaug/smaug-go/internal/combat/dammessage.go` (150-157, 232, 297-305), `…/internal/combat/dammessage_test.go`
- Seams: `/home/eilidh/src/smaug/smaug-go/internal/handler/handler.go` (`CharFromRoom`, `CharToRoom`, `GetEqChar`), `/home/eilidh/src/smaug/smaug-go/internal/util/act.go:283-318` (broadcast target), `/home/eilidh/src/smaug/smaug-go/internal/types/pcdata.go` (`PCData.Flags`), `/home/eilidh/src/smaug/smaug-go/internal/types/constants.go:652` (`PCFLAG_GAG`), `/home/eilidh/src/smaug/smaug-go/internal/types/enums.go:674,782-784` (`ITEM_POISONED`, `WEAR_WIELD`, `WEAR_DUAL_WIELD`)
- Reference: `/home/eilidh/src/smaug/smaug-go/internal/act/skills4.go:326` (where `ITEM_POISONED` is set)

---

## Completion record (2026-04-17)

All three task groups landed in one pass against `combat/dammessage.go`.

### G1 — `was_in_room` swap
- Added an unconditional cross-room swap at the top of `DamMessage`: if `ch.InRoom != victim.InRoom && victim.InRoom != nil`, `handler.CharFromRoom(ch)` + `CharToRoom(ch, victim.InRoom)`, and a `defer` that restores `ch` to the original room. The restore is panic-safe via `defer`.
- Removed the `crossRoom` short-circuit at the former `:297-301` and the stale local. The unified broadcast at the bottom now handles both same-room and cross-room callers.
- Deleted `TestDamMessageDifferentRoomsEmitsToVictOnly` and replaced it with two new tests: `TestDamMessage_DifferentRooms_SwapsAttackerIntoVictimRoom` (bystander in victim room sees buf1; ch restored to room A; room-membership invariants) and `TestDamMessage_DifferentRooms_BystanderInOriginalRoomSeesNothing` (cross-room confirmation — bystander in attacker's original room stays silent).
- **Defer-restore-on-panic test deferred (flagged in task).** `util.Act` never panics through the current descriptor path (`CharData.Send` → `WriteToBuffer` is pure buffer append; `FlushOutput` returns `error`, not `panic`). Constructing a reliable panic surface would require injecting a panicking `net.Conn` AND a flush — but `Act` never flushes, only buffers. The `defer` is retained as correctness insurance, verified by the two landed G1 tests covering normal restore.

### G2 — `PCFLAG_GAG` self-suppress
- Computed `gcflag` / `gvflag` exactly per C: `dam == 0 && !IsNPC && PCData != nil && PCData.Flags & int(PCFLAG_GAG) != 0`.
- At the broadcast site, `TO_NOTVICT` is always emitted; `TO_CHAR` is gated on `!gcflag`; `TO_VICT` is gated on `!gvflag`.
- 5 new tests: `TestDamMessage_GaggedAttacker_ZeroDam_NoToChar`, `TestDamMessage_GaggedVictim_ZeroDam_NoToVict`, `TestDamMessage_BothGagged_OnlyBystanderSees`, `TestDamMessage_GaggedAttacker_PositiveDam_AllDeliver`, `TestDamMessage_NPCAttackerVictimNoGagCrash`. The NPC crash-guard test confirms `PCData == nil` never panics; per the plan this also covers the `DoGag` follow-up since tests set `PCData.Flags` directly.

### G3 — Poisoned-weapon prefix
- Added package-private `isWieldingPoisoned(ch, obj)` at the top of `dammessage.go`, implementing option (a) from the plan: obj is pointer-identical to `handler.GetEqChar(ch, WEAR_WIELD)` or `handler.GetEqChar(ch, WEAR_DUAL_WIELD)` AND `obj.ExtraFlags.IsSet(ITEM_POISONED)`. This matches C `fight.c:104-119` exactly and tolerates the combat-depth G8 dual-wield alternation that threads offhand `obj` through `oneHitFull`.
- Inserted an `else if dt > types.TYPE_HIT && isWieldingPoisoned(ch, obj)` branch BEFORE the existing weapon-damage-type branch. When active, buf1/buf2/buf3 become `"$n's poisoned <attack> <verb> ..."` etc. Per C (`fight.c:4496-4511`) the attack word is taken from `attackTable[dt - TYPE_HIT]`, NOT from `obj.ShortDescr` — C's poisoned branch deliberately uses the generic noun.
- 7 new tests: `TestDamMessage_PoisonedWield_PrimarySlot`, `TestDamMessage_PoisonedWield_DualSlot`, `TestDamMessage_PoisonedButNotEquipped_NoPrefix` (identity check), `TestDamMessage_WieldWithoutPoisonFlag_NoPrefix`, `TestDamMessage_BareHands_PoisonedObjIgnored` (`dt == TYPE_HIT` gate), `TestDamMessage_SkillPathPoisonedBranchSkipped` (sn path unaffected), `TestDamMessage_PoisonedNilObj_NoPrefix`.

### Test delta
- Added `"github.com/eilidhmae/smaug/internal/handler"` import to `combat/dammessage.go`; no new import cycle (`combat/combat.go` already imports `handler`).
- `internal/combat/dammessage_test.go`: deleted 1, added 14 (2 G1 + 5 G2 + 7 G3). 10 pre-existing tests untouched + 13 renamed-style new tests = 23 distinct dam-message test functions now, all PASS under `go test -v -run TestDamMessage`.
- Full package: `go test -count=1 ./internal/combat/...` green (all 82 tests pass). Full project: `go test -count=3 ./...` green across 15 packages.

### Acceptance criteria coverage
1. **G1:** covered by `TestDamMessage_DifferentRooms_SwapsAttackerIntoVictimRoom` (buf1 to bystander in victim's room; `ch.InRoom == roomA` after call; room membership invariants on both rooms) and `TestDamMessage_DifferentRooms_BystanderInOriginalRoomSeesNothing` (no leak to attacker's original-room bystanders).
2. **G2:** all three assertions covered by `TestDamMessage_GaggedAttacker_ZeroDam_NoToChar` (i), `TestDamMessage_GaggedVictim_ZeroDam_NoToVict` (ii), and `TestDamMessage_GaggedAttacker_PositiveDam_AllDeliver` (iii).
3. **G3:** covered by `TestDamMessage_PoisonedWield_PrimarySlot` (primary slot, flag set → "poisoned slice"), `TestDamMessage_WieldWithoutPoisonFlag_NoPrefix` (without flag), `TestDamMessage_BareHands_PoisonedObjIgnored` (`dt == TYPE_HIT` gate), `TestDamMessage_PoisonedNilObj_NoPrefix` (nil obj).
4. **Regression:** 10 unchanged pre-existing dam-message tests still green; 1 intentionally replaced (`TestDamMessageDifferentRoomsEmitsToVictOnly` → `TestDamMessage_DifferentRooms_SwapsAttackerIntoVictimRoom`).

### Deferrals / follow-ups
- `DoGag` player command still not ported (out of scope per plan; can be a trivial follow-up — gag flag is already player-mutable via tests).
- Defer-restore-on-panic test skipped per task allowance — no reachable panic surface in the current `util.Act` → `Send` → buffer path. The `defer` remains as correctness insurance.
- Color-code loss (`AT_ACTION` / `AT_HIT` / `AT_HITME` per-recipient coloring) still flattened by `util.Act` — noted in audit; out of scope.
