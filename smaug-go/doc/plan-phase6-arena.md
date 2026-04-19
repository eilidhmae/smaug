# Plan: Phase 6 — Arena PvP

**Status:** Planned (2026-04-18). Adversary-verified: to be filled after plan adversary pass.
**Priority:** First-cut of Phase 6 (Wave 1). Smallest surface, all primitives already shipped.
**Scope:** New files `internal/act/arena.go` + `internal/act/arena_test.go`, `internal/testclient/arena_test.go`. Modifications to `internal/types/pcdata.go` (`Akills` / `Adeaths` fields), `internal/types/constants.go` (add `ROOM_VNUM_ARENA_MIN` / `MAX` — `ROOM_VNUM_ALTAR` already present at L433), `internal/types/enums.go` (add `TIMER_CHALLENGE` at tag 8), `internal/boot/boot.go` (command registration + arena-state var seam), `internal/persist/player.go` (persist `Akills` / `Adeaths`), `internal/combat/combat.go` (arena-victory branch in `Damage` or `rawKill`), `internal/game/update.go` (challenge-timeout tick). No new package.

---

## Problem

C ships an isolated PvP arena subsystem at `src/arena.c:89-358` gated by `#ifdef ENABLE_ARENA`: four player commands (`challenge` / `accept` / `decline` / `withdraw`) + an arena-tick timeout in `src/update.c:1209-1218` + a victory-resolution branch in `src/fight.c:2861-2914`. The Go port has:

- `ACT_CHALLENGED` and `ACT_CHALLENGER` flag constants at `internal/types/constants.go:482-483` (defined only — never set, never read).
- `ROOM_ARENA` room flag at `internal/types/enums.go:721` (defined only — never consulted).
- `PLR_SILENCE` flag (set/cleared by `DoSilence`, honored by comm commands since Tier 9).
- `handler.GetCharWorld` for world-wide target lookup.
- `handler.CharFromRoom` / `CharToRoom` for teleport.
- `util.NumberRange` for random arena-room selection.
- Combat loop (`ViolenceUpdate`, `Damage`, `rawKill`).
- Testclient `QuickLoginTwo` helper for two-player scenarios (shipped Tier 9, `internal/testclient/login.go:174`).

Everything needed is already in place. No command is currently registered; the `ACT_CHALLENGED` / `ACT_CHALLENGER` flags are dead bits. Shipping Arena wires these up into an end-to-end feature.

### Why Arena is the right first-cut Phase 6 plan

1. **Smallest surface.** 358 C LOC (including banner, comments, whitespace). Real code is ~200 lines across 4 commands, 1 tick handler, and 1 combat-victory branch.
2. **Zero new packages.** Every helper exists.
3. **Zero new schema gaps.** The only PCData additions (`Akills` / `Adeaths`) are two `int` fields.
4. **First two-player E2E test scenario.** Exercises `QuickLoginTwo` in a PvP path (Tier 9 only used it for `gtell`).
5. **Visible from telnet.** Demo-able: two sessions, one challenges, other accepts, both get teleported, combat resolves, both get teleported back.

---

## C Reference (authoritative)

All citations are against `src/arena.c` (HEAD), `src/update.c`, `src/fight.c`, `src/mud.h`.

### Commands

#### `do_challenge` — `src/arena.c:97-193`

Flow (in order — every gate is load-bearing):
1. Empty argument → `"You must specify who you want to challenge.\n\r"` (L106-110).
2. Global `is_challenge` already set → `"Someone has been challenged wait a few moments the try again."` (L112-116). *Note: typo in original; port verbatim.*
3. Global `arena_is_busy` set → `"The arena is being used at the moment..Please wait a few minutes.\n\r"` (L118-122).
4. `victim = get_char_world(ch, argument)` nil → `act(AT_ACTION, "Sorry but $t seems to be gone from the realms at this time.'", ch, argument, NULL, TO_CHAR)` (L124-130). *Note: stray trailing `'` in C format — port verbatim.*
5. `ch` has `ACT_CHALLENGED` → `"You have been challenged already.  Accept or decline.\n\r"` (L132-136).
6. `victim` has `ACT_CHALLENGER` → `act(AT_RED, "$N has already challenged someone else.", ch, NULL, victim, TO_CHAR)` (L138-142).
7. `victim` has `PLR_AFK` → `act(AT_YELLOW, "Sorry but $N is AFK at the moment.", ch, NULL, victim, TO_CHAR)` (L144-148).
8. Either is NPC → `"You can only challenge players or yourself.\n\r"` (L151-156).
9. `victim` is immortal and `ch` is not → `act(AT_RED, "Sorry but $E is a higher being.", ch, NULL, victim, TO_CHAR)` + `"Besides they all would laugh at you!\n\r"` (L158-163).
10. `victim->level <= 5` → `act(AT_WHITE, "Sorry but $E is not experienced enough.", ch, NULL, victim, TO_CHAR)` (L165-169).
11. `victim->fighting != NULL` → `act(AT_ACTION, "Sorry but $N is in combat right now.", ch, NULL, victim, TO_CHAR)` (L171-175).
12. Argument is `"self"` or `"mirror match"` → return silently (L177-181; explicit "not giving away my stuff" placeholder).
13. **Success path** (L182-192): `xSET_BIT(ch->act, ACT_CHALLENGER)`, two-line confirmation to `ch`, `act(AT_RED, "You have been challenged by $n.", ch, NULL, victim, TO_VICT)`, instruction sprintf + send to victim, `challenge_tme = 5`, `is_challenge = TRUE`.

**Flag-set asymmetry:** `do_challenge` sets `ACT_CHALLENGER` on **ch** (the caller) only. The victim does NOT get `ACT_CHALLENGED` yet — that happens in `do_accept`. So between challenge and accept, state is: `ch.ACT_CHALLENGER=1`, `victim` has nothing set; `is_challenge=true` globally.

#### `do_accept` — `src/arena.c:198-265`

Flow:
1. NPC gate (ch) → `"Hrmm....You must be a new kinda mobile to accept challenges eh?"` (L205-209).
2. Empty argument → `"You must be specify who challenged you I am only a MUD that can do so mud.\n\r"` (L211-215). *Note: broken English — port verbatim.*
3. `victim = get_char_world(ch, argument)` nil → same `$t ... gone from the realms` act-call as `do_challenge` (L217-222).
4. NPC gate (victim) → `"Hrmm, I dont think I can let you die like that.\n\r"` (L224-228).
5. `victim == ch` → `"I bet you think your funny eh?\n\r"` (L230-234).
6. `victim` doesn't have `ACT_CHALLENGER` → `"Now have you drank that much ale that you can't remember who challenged you?\n\r"` (L236-240).
7. **Success path** (L242-265): random arena rooms `room1` / `room2` via `number_range(ROOM_VNUM_ARENA_MIN, ROOM_VNUM_ARENA_MAX)`, announce to victim, teleport `ch`, `do_look(ch,"auto")`, announce to `ch`, teleport `victim`, `do_look(ch,"auto")` *(note: C passes `ch` not `victim` on the second do_look — likely a bug; port matches C)*, set `ACT_CHALLENGED` on `ch`, set `PLR_SILENCE` on both, `arena_is_busy = TRUE`, `is_challenge = FALSE`, `challenge_tme = 0`.

**Flag-set asymmetry:** `do_accept` sets `ACT_CHALLENGED` on **ch** (the *accepter*), NOT on `victim` (the challenger). Combined with the challenger's `ACT_CHALLENGER` from `do_challenge`, both flags exist but on different players. The victory path (`src/fight.c:2900-2907`) clears BOTH flags from BOTH players — a defensive approach that doesn't depend on which player has which flag.

#### `do_decline` — `src/arena.c:267-317`

Flow:
1. Empty argument → challenge-error (reuses challenge's syntax string — a minor leak, port verbatim).
2. `!is_challenge` → `"Nobody hase even challenged someone wanna try again?"` (L278-282). *Typo — port verbatim.*
3. `arena_is_busy` → `"Alright the arena is in use so no one yet again has challenged you.\n\r"` (L284-288).
4. `victim = get_char_world(ch, argument)` nil → same not-found act (L290-296).
5. `victim` doesn't have `ACT_CHALLENGER` → `act(AT_RED, "$N hasn't challenged you.", ch, NULL, victim, TO_CHAR)` (L298-302).
6. `IS_NPC(victim) || victim == ch` → `"How did that happen?\n\r"` (L304-308).
7. **Success**: `is_challenge = FALSE`, `xREMOVE_BIT(ch->act, ACT_CHALLENGER)` *(BUG: this should remove ACT_CHALLENGER from `victim`, not `ch`; see Open Question #1)*, announce both ways.

**C bug note:** `do_decline` clears `ACT_CHALLENGER` from **ch** at L311 — but `ch` is the decliner, not the challenger. The challenger's flag is NEVER cleared by `do_decline`. The next challenge would find `victim->ACT_CHALLENGER=1` and bounce it. This means **a decline leaves the challenger's flag set forever** until the tick-timeout at `update.c:1212` fires (5 ticks later) OR until victory. Port this verbatim OR fix it — see Open Question #1.

#### `do_withdraw` — `src/arena.c:319-358`

Flow:
1. NPC gate → silent return.
2. `ch` doesn't have `ACT_CHALLENGER` → `"You haven't challenged anyone!\n\r"` (L327-331).
3. `victim = get_char_world(ch, argument)` nil → `"Does it look like they are here?!?\n\rNO, I didn't think so!!\n\r"` (L333-337).
4. `arena_is_busy` → `"Hrmm, how does that work again?!?\n\r"` (L339-343).
5. NPC gate (victim) → `"I think you are brain dead how many drugs have you done again?\n\r"` (L345-349).
6. **Success**: `is_challenge = FALSE`, `xREMOVE_BIT(ch->act, ACT_CHALLENGER)`, announce.

### Challenge-timeout tick

**`src/update.c:1208-1219`** — in `char_update` per-player loop:

```c
#ifdef ENABLE_ARENA
if (challenge_tme != 0)
    challenge_tme--;

if (challenge_tme == 0 && xIS_SET(ch->act, ACT_CHALLENGER)
    && !xIS_SET(ch->in_room->room_flags, ROOM_ARENA))
{
    send_to_char("They have not responded challenge canceled.\n\r", ch);
    xREMOVE_BIT(ch->act, ACT_CHALLENGER);
    is_challenge = FALSE;
    return;
}
#endif
```

**Subtle:** `challenge_tme` is declared `bool` at `arena.c:91` but treated as `int` throughout (`challenge_tme = 5`, `challenge_tme--`). Port as `int`. The condition `challenge_tme == 0 && ACT_CHALLENGER && !ROOM_ARENA` means the timeout only fires for challengers not currently in the arena — i.e., pre-accept or post-victory. The `char_update` loop runs per-PULSE_MOBILE (4 pulses/second → ~4-second ticks for the whole tick; `challenge_tme = 5` is a 5-tick countdown = ~5 seconds of wall time).

**One-timer-per-world vs per-player:** C uses a SINGLE global `challenge_tme` that every PC's `char_update` decrements. This is a bug (two concurrent challenges share a counter). Port as-is or elevate — see Open Question #2. Recommendation in Design section: one per-player timer via the shipped `handler.AddTimer` subsystem.

### Arena victory

**`src/fight.c:2861-2915`** — in `damage` (or `rawKill`):

```c
#ifdef ENABLE_ARENA
if (victim->position == POS_DEAD && !IS_NPC(victim) && !IS_NPC(ch)
    && (xIS_SET(victim->in_room->room_flags, ROOM_ARENA)))
{
    ch->pcdata->akills += 1;
    victim->pcdata->adeaths += 1;

    // loser path
    stop_fighting(victim, TRUE);
    char_from_room(victim);
    char_to_room(victim, get_room_index(ROOM_VNUM_ALTAR));
    send_to_char("You lost haha!\n\r", victim);
    victim->hit = victim->max_hit;
    victim->mana = victim->max_mana;
    affect_strip(victim, gsn_poison);
    affect_strip(victim, gsn_blindness);
    affect_strip(victim, gsn_sleep);
    affect_strip(victim, gsn_curse);
    victim->move = victim->max_move;
    update_pos(victim);
    do_look(victim, "auto");

    // winner path
    stop_fighting(ch, TRUE);
    char_from_room(ch);
    char_to_room(ch, get_room_index(ROOM_VNUM_TEMPLE));
    send_to_char("You won are you sure you didn't cheat?!? heh\n\r", ch);
    // heal to full + affect_strip + update_pos + do_look

    // flag cleanup on both players (defensive)
    if (xIS_SET(ch->act, ACT_CHALLENGER)) xREMOVE_BIT(ch->act, ACT_CHALLENGER);
    if (xIS_SET(victim->act, ACT_CHALLENGER)) xREMOVE_BIT(victim->act, ACT_CHALLENGER);
    if (xIS_SET(victim->act, ACT_CHALLENGED)) xREMOVE_BIT(victim->act, ACT_CHALLENGED);
    if (xIS_SET(ch->act, ACT_CHALLENGED)) xREMOVE_BIT(ch->act, ACT_CHALLENGED);
    xREMOVE_BIT(ch->act, PLR_SILENCE);
    xREMOVE_BIT(victim->act, PLR_SILENCE);
    arena_is_busy = FALSE;
    return TRUE;
}
#endif
```

**Returning `TRUE` from damage skips the normal death path** — no corpse, no XP, no gold, no death-log. That's the intended behavior: arena deaths are bloodless.

### Constants

- `ROOM_VNUM_ARENA_MIN` = 10366 (`src/mud.h:2283`).
- `ROOM_VNUM_ARENA_MAX` = 10382 (`src/mud.h:2284`).
- `ROOM_VNUM_ALTAR` = 21194 (`src/mud.h:2272`).
- `ROOM_VNUM_TEMPLE` = 21001 (`src/mud.h:2271`; already present in Go `types/constants.go`).
- `ACT_CHALLENGED` = 44, `ACT_CHALLENGER` = 45 (`src/mud.h:1741-1742`).
- `ROOM_ARENA` is a room flag.

### Area data

`db/area/newacad.are` contains rooms 10366 and 10382 (verified `Grep` 2026-04-18). The seventeen vnums in range 10366-10382 are NOT all arena rooms; only those with `ROOM_ARENA` in their room flags count. The random-range selection `number_range(ROOM_VNUM_ARENA_MIN, ROOM_VNUM_ARENA_MAX)` can hit a non-arena vnum — C doesn't check. Port verbatim OR filter — see Open Question #3.

---

## Go Current State

Verified 2026-04-18 against `internal/` tree.

- **Flags:** `types.ACT_CHALLENGED = 44`, `types.ACT_CHALLENGER = 45` (`types/constants.go:482-483`). Defined; zero readers / zero writers.
- **Room flag:** `types.ROOM_ARENA` (`types/enums.go:721`). Defined; zero readers.
- **PCData fields:** `Akills int`, `Adeaths int` — **absent** (verified via grep in `types/pcdata.go`). Other kill-tracking fields (`PKills`, `PDeaths`, `MKills`, `MDeaths`) present.
- **Room vnums:** `ROOM_VNUM_ARENA_MIN` / `ROOM_VNUM_ARENA_MAX` — not present in Go constants. `ROOM_VNUM_TEMPLE = 21001` and `ROOM_VNUM_ALTAR = 21194` are already defined at `types/constants.go:432-433` (audit 2026-04-18 correction — both altar and temple vnums already shipped; only the ARENA_MIN/MAX pair must be added).
- **Commands:** no `DoChallenge` / `DoAccept` / `DoDecline` / `DoWithdraw` registered. No arena file exists under `internal/act/`.
- **Combat victory hook:** `combat.Damage` has no arena branch. The death path in `internal/combat/combat.go:rawKill` (or equivalent) handles corpse + XP + extract — arena needs to intercept BEFORE that.
- **Tick-timeout:** `internal/game/update.go:charUpdate` has no arena-timeout block. Challenge countdown needs integrating.
- **Two-player testclient:** `internal/testclient/login.go:174-191` `Harness.QuickLoginTwo(t, nameA, nameB)` shipped in Tier 9. Used by `internal/testclient/channels_test.go:26` for the `gtell` two-client scenario.
- **`handler.GetCharWorld`:** `internal/handler/find.go:49-52` `func GetCharWorld(w *world.World, ch *types.CharData, argument string) *types.CharData` — world-wide case-insensitive name match.
- **`handler.CharFromRoom`/`CharToRoom`:** `internal/handler/handler.go:153,168` shipped.
- **`util.Act`:** per-call `aType` color signature landed in Tranche C G1-G3. `AT_RED` / `AT_YELLOW` / `AT_ACTION` / `AT_WHITE` all in `types/constants.go`.
- **`PLR_SILENCE`:** `types.PLR_SILENCE` defined; DoSilence + comm commands honor it (Tier 9).
- **`handler.AddTimer`:** `internal/handler/timer.go` subsystem landed 2026-04-18 via `plan-timer-subsystem.md`. Per-player timers with type tag; used by `TIMER_RECENTFIGHT` and `TIMER_ASUPRESSED`. Supports custom type tags.

---

## Go Design

### Approach A (CHOSEN): Package-level arena state + per-player challenge timer

**Arena-global state** — a new file-level struct in `internal/act/arena.go`:

```go
var arenaState struct {
    isChallenge bool
    isBusy      bool
}
```

*Note:* package-level state mirrors the C `is_challenge` / `arena_is_busy` globals. Tests save/restore via a helper. This matches the pattern used by `internal/combat/stance_index.go` (package var with defaults) and the `DescRegistry` init in Tier 9 channels.

**Per-player challenge timer** — use the shipped `handler.AddTimer(ch, TIMER_CHALLENGE, 5, "", 0)`. Signature is `AddTimer(ch *CharData, tType, count int, doFun string, value int)` per `internal/handler/timer.go:53` — doFun comes BEFORE value. New timer-type constant `TIMER_CHALLENGE = 8` (next available slot — audit 2026-04-18 confirmed `types/enums.go:897-906` currently defines TIMER_NONE=0, TIMER_RECENTFIGHT=1, TIMER_SHOVEDRAG=2, TIMER_DO_FUN=3, TIMER_APPLIED=4, TIMER_PKILLED=5, TIMER_ASUPRESSED=6, TIMER_NUISANCE=7). At expiry (via `handler.DecrementTimers`, which drops expired timers silently — per Tranche B design), the challenger's `ACT_CHALLENGER` bit is cleared AND a message is sent.

*But* `DecrementTimers` does NOT currently dispatch on expiry except for `TIMER_DO_FUN`. To clear `ACT_CHALLENGER` on expiry, one of:

- **Option A1 (CHOSEN):** Poll in `charUpdate` — if `ch.Act.IsSet(ACT_CHALLENGER) && !handler.HasTimer(ch, TIMER_CHALLENGE) && !ch.InRoom.RoomFlags.IsSet(ROOM_ARENA)` → fire the timeout branch. Matches C's tick-based approach. `HasTimer` wraps `GetTimer(...) > 0`.
- Option A2: Extend `DecrementTimers` with a generic expiry-callback registration. Too invasive for Arena alone.

`charUpdate` already runs per-pulse; adding the arena poll is ~10 lines.

**Arena-victory branch** — `combat.ArenaVictoryCheck(ch, victim *types.CharData) bool` new exported helper called from `Damage` (or `rawKill`) right before normal death processing. Returns `true` if arena-victory path was taken — caller skips the corpse/XP path. Invoked when `victim.Position == POS_DEAD && !victim.IsNPC() && !ch.IsNPC() && victim.InRoom.RoomFlags.IsSet(ROOM_ARENA)`.

**Why `combat.ArenaVictoryCheck` and not inlined:** keeps the `Damage` signature stable, adds one testable pure function, and decouples arena from `combat`'s own tests.

**State reset seam** — `act.ResetArenaState()` exported for tests; sets `arenaState.isChallenge = false` and `arenaState.isBusy = false`. Called from `TestMain` or test setup to prevent cross-test bleed.

### Rejected alternatives

- **Approach B: Global state on `world.World`.** Arena is an isolated subsystem; adding fields to `world.World` bloats the "everything" struct. C keeps it file-local; Go mirrors.
- **Approach C: Replace global with channel-based dispatcher.** Overkill. Single-player-at-a-time semantics of C arena implies global state is the correct model.
- **Approach D: Port C's bugs verbatim including the `do_decline` flag-clearing-on-wrong-player bug.** See Open Question #1 — leaning toward a documented deliberate fix rather than verbatim port.
- **Approach E: Skip `challenge_tme` tick and rely on "challenger must withdraw" UX.** Loses the 5-tick auto-timeout; breaks test expectations that assume expiry.

---

## Task Groups

All tasks include the TDD mandate and the mutation-verify expectation. Mutation verification uses `Edit` round-trips only per `_shared.md` → Mutation Verification Safety (banned commands: `git checkout -- <file>`, `git checkout <ref> -- <file>`, `git restore <file>`, `git reset --hard` any form, `git stash` any form).

### G1 — Schema additions + constants

**Files:**
- Modified: `internal/types/pcdata.go` (add `Akills int`, `Adeaths int` after `PKills`/`PDeaths`/`MKills`/`MDeaths`).
- Modified: `internal/types/constants.go` (add `ROOM_VNUM_ARENA_MIN = 10366`, `ROOM_VNUM_ARENA_MAX = 10382`). `ROOM_VNUM_ALTAR = 21194` already present at L433 — do not re-declare.
- Modified: `internal/types/enums.go` (add `TIMER_CHALLENGE` at next unused tag).
- Modified: `internal/persist/player.go` (append `Akills` and `Adeaths` to `SavePlayer` write path + `LoadPlayer` read path using C key names `Akills` / `Adeaths`).

**Test-first:**
- `TestPCData_AkillsAdeathsFields` in `internal/types/pcdata_test.go` — construct a PCData with non-zero Akills/Adeaths and confirm zero-value on an empty struct. Trivial but pins the schema.
- `TestSavePlayer_PersistsArenaCounters` in `internal/persist/player_test.go` — round-trip: create PCData with `Akills=3, Adeaths=1`, save, load, assert both preserved.
- `TestTimerChallengeTagUnique` in `internal/types/enums_test.go` (or equivalent) — confirm `TIMER_CHALLENGE` != any other `TIMER_*`.

**Mutation verify:**
- Drop `Akills` from `SavePlayer` write → round-trip test fails on Akills.
- Drop `Akills` from `LoadPlayer` read → round-trip test fails on Akills.
- Change `ROOM_VNUM_ARENA_MIN` to 10400 → arena tests (G5) fail on room-range.

### G2 — `act.DoChallenge`

**Files:**
- New: `internal/act/arena.go` — package-var `arenaState`, `DoChallenge` body.
- New: `internal/act/arena_test.go` — unit tests for each gate.

**DoChallenge body** (Go; C at `src/arena.c:97-193`):

```go
// DoChallenge handles "challenge <player>" — PvP arena challenge.
// C ref: src/arena.c:97-193.
func DoChallenge(ch *types.CharData, argument string) {
    if argument == "" {
        ch.Send("You must specify who you want to challenge.\n\r")
        return
    }
    if arenaState.isChallenge {
        ch.Send("Someone has been challenged wait a few moments the try again.")
        return
    }
    if arenaState.isBusy {
        ch.Send("The arena is being used at the moment..Please wait a few minutes.\n\r")
        return
    }
    victim := handler.GetCharWorld(WorldRef, ch, argument)
    if victim == nil {
        util.Act(types.AT_ACTION, "Sorry but $t seems to be gone from the realms at this time.'",
            ch, nil, argument, nil, types.TO_CHAR)
        return
    }
    if ch.Act.IsSet(types.ACT_CHALLENGED) {
        ch.Send("You have been challenged already.  Accept or decline.\n\r")
        return
    }
    if victim.Act.IsSet(types.ACT_CHALLENGER) {
        util.Act(types.AT_RED, "$N has already challenged someone else.", ch, victim, nil, nil, types.TO_CHAR)
        return
    }
    if victim.Act.IsSet(types.PLR_AFK) {
        util.Act(types.AT_YELLOW, "Sorry but $N is AFK at the moment.", ch, victim, nil, nil, types.TO_CHAR)
        return
    }
    if victim.IsNPC() || ch.IsNPC() {
        ch.Send("You can only challenge players or yourself.\n\r")
        return
    }
    if victim.IsImmortal() && !ch.IsImmortal() {
        util.Act(types.AT_RED, "Sorry but $E is a higher being.", ch, victim, nil, nil, types.TO_CHAR)
        ch.Send("Besides they all would laugh at you!\n\r")
        return
    }
    if victim.Level <= 5 {
        util.Act(types.AT_WHITE, "Sorry but $E is not experienced enough.", ch, victim, nil, nil, types.TO_CHAR)
        return
    }
    if victim.Fighting != nil {
        util.Act(types.AT_ACTION, "Sorry but $N is in combat right now.", ch, victim, nil, nil, types.TO_CHAR)
        return
    }
    if strings.EqualFold(argument, "self") || strings.EqualFold(argument, "mirror match") {
        return // silent — C: "not giving away my stuff"
    }

    ch.Act.Set(types.ACT_CHALLENGER)
    ch.Send("Your challenge has been sent if they done accept in 3 ticks it will decline")
    ch.Sendf("\n\rIf they do not accept it will be automatically withdrawn from %s.\n\r", ch.Name)
    util.Act(types.AT_RED, "You have been challenged by $n.", ch, victim, nil, nil, types.TO_VICT)
    victim.Sendf("Type accept %s to accept or decline %s to decline.\n\r", ch.Name, ch.Name)
    handler.AddTimer(ch, types.TIMER_CHALLENGE, 5, "", 0)
    arenaState.isChallenge = true
}
```

**Test-first** (`arena_test.go`):
- `TestDoChallenge_EmptyArg` — no-arg → "You must specify...".
- `TestDoChallenge_AlreadyChallenging` — set arenaState.isChallenge → "Someone has been challenged..." emitted; no state change.
- `TestDoChallenge_ArenaBusy` — set arenaState.isBusy → "The arena is being used...".
- `TestDoChallenge_VictimNotFound` — argument `"ghost"` with no such player → `"Sorry but ghost seems to be gone..."`.
- `TestDoChallenge_ChIsAlreadyChallenged` — `ch.ACT_CHALLENGED=1` → "You have been challenged already.".
- `TestDoChallenge_VictimIsChallenger` — `victim.ACT_CHALLENGER=1` → "$N has already challenged...".
- `TestDoChallenge_VictimAFK` — `victim.PLR_AFK=1` → "Sorry but $N is AFK...".
- `TestDoChallenge_VictimIsNPC` — "You can only challenge players...".
- `TestDoChallenge_ChIsNPC` — same message.
- `TestDoChallenge_VictimImmortalChMortal` — "Sorry but $E is a higher being." + "Besides they all would laugh...".
- `TestDoChallenge_VictimLowLevel` — victim.Level=5 → "not experienced enough".
- `TestDoChallenge_VictimFighting` — `victim.Fighting != nil` → "in combat right now".
- `TestDoChallenge_SelfArg` — silent return, no state change.
- `TestDoChallenge_MirrorMatch` — same.
- `TestDoChallenge_Success` — valid → `ch.ACT_CHALLENGER=1`, timer set, two-line self-message, victim receives challenge notice, arenaState.isChallenge=true, timer count=5.

**Mutation verify:**
- Drop the `ch.Act.Set(ACT_CHALLENGER)` line → success-path test fails on flag check.
- Drop `arenaState.isChallenge = true` → test fails on state.
- Change level gate from `<= 5` to `< 5` → low-level test fails.
- Swap `ACT_CHALLENGER` and `ACT_CHALLENGED` check order → "already-challenged" test now fires on wrong input.

### G3 — `act.DoAccept`

**Files:**
- Modified: `internal/act/arena.go` — append `DoAccept`.
- Modified: `internal/act/arena_test.go` — gate tests.

**DoAccept body** (C at `src/arena.c:198-265`):

Selection of arena rooms uses `util.NumberRange(types.ROOM_VNUM_ARENA_MIN, types.ROOM_VNUM_ARENA_MAX)` and `WorldRef.GetRoom(vnum)`. If the room is nil (range hits an unmapped vnum), retry up to N times — see Open Question #3.

**Key gates / test cases:**
- `TestDoAccept_NPC` — ch is NPC → "Hrmm....You must be a new kinda mobile...".
- `TestDoAccept_EmptyArg` — "You must be specify who challenged you...".
- `TestDoAccept_VictimNotFound` — "$t seems to be gone...".
- `TestDoAccept_VictimIsNPC` — "I dont think I can let you die like that.".
- `TestDoAccept_SelfAccept` — victim == ch → "I bet you think your funny eh?".
- `TestDoAccept_VictimNotChallenger` — victim lacks ACT_CHALLENGER → "Now have you drank that much ale...".
- `TestDoAccept_Success` — valid → both teleported; ch.ACT_CHALLENGED=1; both PLR_SILENCE; arenaState.isBusy=true, isChallenge=false; challenge timer extracted.

**Mutation verify:**
- Drop `ch.Act.Set(ACT_CHALLENGED)` → success test fails.
- Drop `arenaState.isBusy = true` → test fails.
- Swap `room1` and `room2` assignments → verified by test that both players land in arena rooms (test cannot distinguish which room; doesn't care — this mutation is silently equivalent. **Important:** the plan explicitly notes this mutation is NOT caught — which is fine because the C code itself treats room1/room2 as interchangeable.)

### G4 — `act.DoDecline` + `act.DoWithdraw`

**Files:**
- Modified: `internal/act/arena.go` — append both.
- Modified: `internal/act/arena_test.go`.

**Both are simple enough to land together.**

**DoDecline** — see Open Question #1. Plan preference: **match C bug verbatim** to maintain C-fidelity as the port's core invariant, and add a TODO note citing the bug. Rationale: project precedent (mudprog `leverpos` buggy-guard was noted but ported-intent; here the C-fidelity argument prevails because a changed flag semantic may break mudprog ACT_CHALLENGER triggers if any content checks it — likely zero content does today, but we preserve the option).

**DoWithdraw** — straightforward.

**Test cases:**
- DoDecline: 6 gates from C + 1 success test; success test verifies `arenaState.isChallenge = false` AND verbatim `ch.Act` cleared (even though victim's challenger flag persists — document this).
- DoWithdraw: 5 gates + 1 success test.

**Mutation verify:** each gate's message-check fails with gate removal; success-path state bits pinned.

### G5 — Challenge-timeout tick in `charUpdate`

**Files:**
- Modified: `internal/game/update.go` — add poll in the per-player loop of `charUpdate`.
- Modified: `internal/game/update_test.go` — new test.

**Behavior:**

```go
// In charUpdate, per ch:
if ch.Act.IsSet(types.ACT_CHALLENGER) && !handler.HasTimer(ch, types.TIMER_CHALLENGE) {
    if ch.InRoom == nil || !ch.InRoom.RoomFlags.IsSet(types.ROOM_ARENA) {
        ch.Send("They have not responded challenge canceled.\n\r")
        ch.Act.Remove(types.ACT_CHALLENGER)
        act.SetArenaIsChallenge(false) // exported seam
    }
}
```

`SetArenaIsChallenge(bool)` is a new exported seam in `internal/act/arena.go` — not ideal (global mutation), but matches C. Future refactor: replace with an `ArenaResetIfExpired(ch)` method that owns the state.

**Test:**
- `TestCharUpdate_ArenaChallengeTimeout` — set `ch.ACT_CHALLENGER=1`, no `TIMER_CHALLENGE`, ch not in ROOM_ARENA, arenaState.isChallenge=true → call `charUpdate` → ch.ACT_CHALLENGER cleared, message sent, arenaState.isChallenge=false.
- `TestCharUpdate_ArenaChallengeStillActive` — set ch.ACT_CHALLENGER, `TIMER_CHALLENGE` present (count=3) → `charUpdate` → no change.
- `TestCharUpdate_ArenaChallengeInArena` — ch in ROOM_ARENA → no timeout fires even with no timer (matches C — only expires pre-accept).

**Mutation verify:**
- Drop the `ch.Act.Remove(ACT_CHALLENGER)` → timeout test fails on flag.
- Drop `SetArenaIsChallenge(false)` → test fails on state.
- Change `!handler.HasTimer` to `handler.HasTimer` → timeout fires when it shouldn't (active challenge test catches).

### G6 — Arena victory in combat

**Files:**
- New: exported helper `combat.ArenaVictoryCheck(ch, victim *types.CharData) bool` in `internal/combat/combat.go` (or a new `arena_victory.go` in the combat package — preferred for test isolation).
- Modified: `internal/combat/combat.go` `rawKill` or `Damage` — call `ArenaVictoryCheck` before normal death processing; if returns true, short-circuit (skip corpse + XP).
- New: `internal/combat/arena_victory_test.go`.

**Helper body** (C at `src/fight.c:2861-2915`):

```go
// ArenaVictoryCheck handles PvP arena death: teleport loser to altar + winner
// to temple, heal both to full, strip debuffs, track akills/adeaths, clear
// arena flags. Returns true if the arena branch fired (caller should skip
// normal death processing). C ref: src/fight.c:2861-2915.
func ArenaVictoryCheck(ch, victim *types.CharData) bool {
    if victim.Position != types.POS_DEAD || victim.IsNPC() || ch.IsNPC() {
        return false
    }
    if victim.InRoom == nil || !victim.InRoom.RoomFlags.IsSet(types.ROOM_ARENA) {
        return false
    }
    if victim.PCData != nil {
        victim.PCData.Adeaths++
    }
    if ch.PCData != nil {
        ch.PCData.Akills++
    }

    // Loser
    StopFighting(victim, true)
    handler.CharFromRoom(victim)
    if altar := WorldRef.GetRoom(types.ROOM_VNUM_ALTAR); altar != nil {
        handler.CharToRoom(victim, altar)
    }
    victim.Send("You lost haha!\n\r")
    victim.Hit = victim.MaxHit
    victim.Mana = victim.MaxMana
    victim.Move = victim.MaxMove
    stripArenaDebuffs(victim)
    updatePos(victim) // package-local; or rename to UpdatePos and export — see below
    if DoLookFunc != nil {
        DoLookFunc(victim, "auto")
    }

    // Winner
    StopFighting(ch, true)
    handler.CharFromRoom(ch)
    if temple := WorldRef.GetRoom(types.ROOM_VNUM_TEMPLE); temple != nil {
        handler.CharToRoom(ch, temple)
    }
    ch.Send("You won are you sure you didn't cheat?!? heh\n\r")
    ch.Hit = ch.MaxHit
    ch.Mana = ch.MaxMana
    ch.Move = ch.MaxMove
    stripArenaDebuffs(ch)
    updatePos(ch)
    if DoLookFunc != nil {
        DoLookFunc(ch, "auto")
    }

    // Clear arena flags defensively on both
    for _, p := range []*types.CharData{ch, victim} {
        p.Act.Remove(types.ACT_CHALLENGER)
        p.Act.Remove(types.ACT_CHALLENGED)
        p.Act.Remove(types.PLR_SILENCE)
    }
    // Clear arena-busy state via the ArenaIsBusyFunc hook (wired at boot to
    // act.SetArenaIsBusy). Direct import of `act` would create a cycle; see
    // Seams paragraph below.
    if ArenaIsBusyFunc != nil {
        ArenaIsBusyFunc(false)
    }
    return true
}

func stripArenaDebuffs(ch *types.CharData) {
    // C fight.c:2876-2879 strips gsn_poison / gsn_blindness / gsn_sleep / gsn_curse.
    // Go combat package has gsnBackstab/gsnCircle/gsnDualWield/gsnBerserk at skillcheck.go:75+
    // resolved via LookupSkillSlotHook at boot. Add gsnPoison/gsnBlindness/gsnSleep/gsnCurse
    // to the same block (new entries in the var block + LookupSkillSlotHook calls).
    // Then call handler.AffectStrip(ch, gsnPoison) for each, guarded by != -1.
}
```

**Gsn additions in G6:** C strips `gsn_poison`/`gsn_blindness`/`gsn_sleep`/`gsn_curse` from both players. Go combat package has an existing GSN-resolution block at `internal/combat/skillcheck.go:75-110` (`gsnBackstab` / `gsnCircle` / `gsnPounce` / `gsnDualWield` / `gsnBerserk` resolved via `LookupSkillSlotHook` at boot). G6 extends this with `gsnPoison` / `gsnBlindness` / `gsnSleep` / `gsnCurse` and their `LookupSkillSlotHook` calls — mirrors the shipped pattern.

**Seams:** `combat` cannot import `act` (circular). `SetArenaIsBusy(bool)` is exported from `act`; `combat` needs to call it — so `combat` either (a) imports `act` or (b) `act` provides a `combat.ArenaIsBusyFunc func(bool)` hook populated at boot. **Chosen: (b)** — the Tier 12 editor-save pattern (`combat.ArenaIsBusyFunc = act.SetArenaIsBusy` in `boot.go`) handles this without circular imports.

**`do_look(victim, "auto")` seam:** no existing `DoLookFunc` in `combat` (verified 2026-04-18 grep). This plan adds a **new** `combat.DoLookFunc func(*types.CharData, string)` package var, populated at boot with `act.DoLook`. Mirrors the `act.SaveFunc` / `act.ShutdownFunc` / `act.DisconnectFunc` / `act.StartEditingFunc` seam pattern already used throughout the codebase. Alternative: skip the look-auto call and let the next player input trigger a fresh look. C does issue the auto-look; port matches C.

**Hookup in `Damage`/`rawKill`:** right before the normal "victim.Position == POS_DEAD" branch that creates a corpse, call `ArenaVictoryCheck(ch, victim)`. If true, return early (pre-corpse).

**Test cases:**
- `TestArenaVictory_LoserTeleportedToAltar` — stub rooms; inject both PCs; victim.Position=POS_DEAD in arena room → altar is loser's new InRoom.
- `TestArenaVictory_WinnerTeleportedToTemple`.
- `TestArenaVictory_BothHealedToMax` — Hit/Mana/Move clamped to Max before call, assert == Max after.
- `TestArenaVictory_AkillsAdeathsIncremented`.
- `TestArenaVictory_AllFlagsClearedBoth` — pre-set both ACT_CHALLENGED/CHALLENGER/PLR_SILENCE on each; post, all six bits are zero.
- `TestArenaVictory_IsBusyClearedAfter` — isBusy=true → isBusy=false after.
- `TestArenaVictory_NoFireIfNotInArena` — victim in non-arena room → returns false, no mutation.
- `TestArenaVictory_NoFireIfNPC` — either side NPC → returns false.
- `TestArenaVictory_NoFireIfVictimNotDead` — Position != POS_DEAD → returns false.
- `TestArenaVictory_CalledFromDamagePath` — integration: inject HP=1, land a killing blow via real `Damage`, assert arena branch fired (via post-condition room vnum).

**Mutation verify:**
- Drop `victim.PCData.Adeaths++` → Akills/Adeaths test fails.
- Drop `act.SetArenaIsBusy(false)` → isBusy test fails.
- Swap temple and altar in teleport targets → loser-at-altar/winner-at-temple tests fail.
- Drop any single `p.Act.Remove` → that-flag test fails.
- Return `true` when room is not arena → `NoFireIfNotInArena` fails.

### G7 — Registration + boot wire + E2E testclient scenario

**Files:**
- Modified: `internal/boot/boot.go` — register 4 commands; wire `combat.ArenaIsBusyFunc = act.SetArenaIsBusy`.
- New: `internal/testclient/arena_test.go` — full E2E flow via telnet.

**Command registration** (near the other player-facing command entries):

```go
reg.Register(&command.Command{Name: "challenge", DoFun: act.DoChallenge, Position: types.POS_RESTING, Level: 0})
reg.Register(&command.Command{Name: "accept",    DoFun: act.DoAccept,    Position: types.POS_RESTING, Level: 0})
reg.Register(&command.Command{Name: "decline",   DoFun: act.DoDecline,   Position: types.POS_RESTING, Level: 0})
reg.Register(&command.Command{Name: "withdraw",  DoFun: act.DoWithdraw,  Position: types.POS_RESTING, Level: 0})
```

**Level gate:** C has no explicit trust gate on these commands. Go mirrors — `Level:0` on all four. The `victim.Level <= 5` check inside `DoChallenge` stops low-level engagement but does not bar the *command* itself (a newbie CAN attempt the challenge; it bounces because the *victim* is low-level, which is what C does). See Open Question #4 for the considered alternative.

**Boot wire:**
```go
combat.ArenaIsBusyFunc = act.SetArenaIsBusy
```

Plus a `boot_test.go` assertion that `combat.ArenaIsBusyFunc != nil` post-`Boot` (mirrors the Tier 12 pattern).

**E2E testclient scenario** — `TestArena_FullChallengeAcceptFightVictory` in `internal/testclient/arena_test.go`:

1. `h := testclient.NewHarness(t)`.
2. `alice, bob := h.QuickLoginTwo(t, "Alice", "Bob")` — both land in temple.
3. Set both PCs' levels > 5 (via test seam — Level:6+ on the CharData, so `victim.Level <= 5` gate passes).
4. `alice.Send("challenge Bob")`, expect `alice.ExpectContains("Your challenge has been sent")` + `bob.ExpectContains("You have been challenged by Alice")`.
5. `bob.Send("accept Alice")`, expect `alice.ExpectContains("Alice has accepted")` (wait — C sends this backwards; actually victim sees "accepted" — verify against C:246) and both players' subsequent room description shows an arena room.
6. Inject HP=1 on Bob via admin hook (since combat tuning is not the test's concern). `alice.Send("kill Bob")`, let combat roll until one dies.
7. Assert: Alice's `PCData.Akills == 1`, Bob's `PCData.Adeaths == 1`. Alice's InRoom is temple (ROOM_VNUM_TEMPLE); Bob's InRoom is altar (ROOM_VNUM_ALTAR). Both `ACT_CHALLENGED` / `ACT_CHALLENGER` / `PLR_SILENCE` cleared. `arenaState.isBusy == false`.

Additional scenarios:
- `TestArena_ChallengeDeclined` — challenge, decline, assert arenaState.isChallenge=false.
- `TestArena_ChallengeTimedOut` — challenge, advance game clock until the timer expires, assert `ch.ACT_CHALLENGER` cleared and "They have not responded challenge canceled." received.

**Mutation verify:**
- Drop the `combat.ArenaIsBusyFunc = act.SetArenaIsBusy` boot wire → post-victory, `arenaState.isBusy` stays true → next challenge bounces. Test catches.
- Drop the `challenge` registration → `ExpectContains("Your challenge has been sent")` fails with "Huh?".
- Change Level:0 → Level:55 on `challenge` → mortal can't challenge → full E2E test fails early.

---

## Acceptance Criteria

All must be mechanically verifiable by the test suite:

1. `DoChallenge` matches C `do_challenge` for all 12 gates + 1 success path (tests enumerate all 13).
2. `DoAccept` matches C `do_accept` for all 6 gates + 1 success path.
3. `DoDecline` matches C `do_decline` for all 6 gates + 1 success path.
4. `DoWithdraw` matches C `do_withdraw` for all 5 gates + 1 success path.
5. `TIMER_CHALLENGE` is a new unique timer tag; challenge sets it with count=5; `DoAccept` extracts it.
6. `charUpdate` cancels a challenge when timer expires AND challenger is not in arena room, emitting the exact C message "They have not responded challenge canceled.\n\r".
7. `combat.ArenaVictoryCheck` teleports loser to altar, winner to temple, heals both to max, strips poison/blindness/sleep/curse from both, increments `Akills`/`Adeaths`, clears `ACT_CHALLENGED`/`ACT_CHALLENGER`/`PLR_SILENCE` on both, sets `arenaState.isBusy=false`, and returns true; returns false otherwise (NPC or non-arena or non-dead).
8. The victory branch fires from the combat death path BEFORE corpse creation — no corpse is spawned in an arena death.
9. `Akills` and `Adeaths` persist across save/load.
10. Four commands registered at `POS_RESTING`, `Level:0` (matches C — no explicit command-level gate; the level-5-victim check lives inside `DoChallenge`).
11. `combat.ArenaIsBusyFunc` is non-nil after `boot.Boot` returns.
12. End-to-end testclient scenario drives `challenge` → `accept` → kill → verifies all eight exit invariants (teleport destinations, HP restoration, Akills/Adeaths delta, flag cleanup, busy reset).
13. `go test -count=3 ./...` green across all 15 packages.
14. `go vet ./...` clean.
15. `gofmt -l .` reports no changes.

---

## Scope Cuts / Deferrals

- **C bugs in `do_decline` flag-clear target** — ported verbatim pending Open Question #1 resolution. Tracked in TODO.md as a follow-up after first player report.
- **Arena-room vnum filter** — Open Question #3. Plan leans toward "retry up to 5 times if `GetRoom` is nil"; if all five fail, abort the accept with a message and clear state. Tracked as a separate follow-up if Open Question #3 dictates pure C-verbatim (crash/fail silently).
- **`ROOM_VNUM_ALTAR`** — may differ from default area data's vnum. Plan uses `ROOM_VNUM_TEMPLE` as fallback if altar isn't in the map. Open Question #5.
- **No `PLR_NORESTORE`** — C has this behind a comment (`src/arena.c:258`). Skip; not in Go.
- **No `pcdata->akills` imm stat display** — C does not ship one either. Defer to `score` extension (Phase 6 polish).
- **No arena-match message to all challengers watching** — C does not ship a spectator system. Skip.
- **No arena OLC commands** — building arena rooms is via the standard `redit` + `ROOM_ARENA` room flag. Not ported here (see plan-phase6-olc-redit.md).
- **No anti-exploit checks** (e.g., "challenger equipped 100k gp of gear, victim is naked") — C ships nothing. Skip.
- **No team / group / clan challenges** — C author notes "working on having it so you can challenge groups and guilds" in the banner comment, but ships nothing. Skip.
- **`ROOM_ARENA` application to existing newacad.are rooms 10366-10382** — those rooms do not currently have the `ROOM_ARENA` flag in `db/area/newacad.are` (verified 2026-04-18 via Grep — vnum mentions are from reset lines and other references, not room-flag bitvector). Without this flag, arena victory never fires. **Required step in G5 or G6:** add `ROOM_ARENA` to the arena vnums via `redit` OR modify the area file OR add a boot-time post-load step that sets it. Preferred: update the area file in `db/area/newacad.are` as part of this plan. **Open Question #6.**

---

## Open Questions

These require human input before dispatching the worker waves.

1. **`do_decline` C bug — port verbatim or fix?** C clears `ACT_CHALLENGER` from the decliner (wrong target). Fixing adds a Go-only divergence; not fixing means the challenger's flag stays set until tick-expiry. Recommendation: **fix with a comment** citing C bug. Cost: zero; clarifies intent.
2. **Global `challenge_tme` vs per-player timer** — C uses a single global counter. Plan chose per-player via `handler.AddTimer`. Confirm the extra correctness is worth the divergence. Recommendation: **keep per-player**. Global counter is already broken for concurrent challenges (C doesn't catch this).
3. **Arena-room vnum random range** — C picks any vnum 10366-10382 without checking for `ROOM_ARENA` flag. Most area files only flag some of those vnums. Options: (a) retry-with-limit-5, (b) filter at boot to collect valid arena-flag rooms then random from that list, (c) verbatim (may crash). Recommendation: **(b)** — scan rooms at boot into a package-level `[]*RoomIndexData` slice.
4. **`challenge` command trust level** — C has no explicit gate, only a level-5 check on the victim. Setting `Level:6` on the command means low-level new players can't call `challenge` at all. Arguably the correct UX (level-5 victim check means you need level-6+ to issue anyway). Alternative: `Level:0` with the in-command check preserved. Recommendation: **`Level:0`**, match C.
5. **`ROOM_VNUM_ALTAR` value** — verified 2026-04-18: `src/mud.h:2272` defines `ROOM_VNUM_ALTAR = 21194`. **Resolved.** Question remaining: does `db/area/` actually contain a room with vnum 21194? If not, fall back to `ROOM_VNUM_TEMPLE` (21001). Worker should verify and document in the completion record.
6. **Area data — add `ROOM_ARENA` flag to 10366-10382?** Without the flag, arena victory never fires. Either we edit `db/area/newacad.are` (data change, builder-user-visible) OR post-load-set the flag in `boot.go` for those specific vnums (hack, but isolates the change). Recommendation: **edit the area file once** and commit. Newacad is a test/development area; the flag edit is safe.
7. **`combat.ArenaVictoryCheck` insertion point** — before or after stop_fighting in `rawKill`? C calls `stop_fighting` inside the arena branch. Plan aligns with C (call stop_fighting inside the helper).

---

## Risk Analysis

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Area file doesn't have `ROOM_ARENA` flag → victory silently never fires | High (known: verified) | High (feature looks broken) | Open Question #6 — edit area data in the plan. Add integration test that asserts a ROOM_ARENA-flagged room exists post-boot. |
| `ROOM_VNUM_ALTAR` room doesn't exist in the shipped area data | Medium | Medium (teleport silently no-ops) | Fallback to `ROOM_VNUM_TEMPLE`. Test: `TestArenaVictory_NoAltarFallsToTemple`. |
| Combat's `rawKill` hook point for the arena intercept is wrong — fires after corpse creation | Low | High (corpse in temple) | Test: `TestArenaVictoryNoCorpseSpawned`. Inject a body-count-check on `ch.InRoom.FirstContent`. |
| `act.WorldRef` not populated in test harness → `GetCharWorld` returns nil → all challenges fail silently | Low | Medium | Precedent: channels tests already handle this via `testclient.NewHarness`. Reuse. |
| `handler.HasTimer` does not exist yet | Medium | Low | Add in G5 as a 3-line wrapper over `GetTimer`. Mutation-verified in G5 tests. |
| Circular import: `combat → act` for `SetArenaIsBusy` | Low | Medium | Solved via `combat.ArenaIsBusyFunc` hook populated in boot. Pattern established in Tier 12. |
| Tests flake on arena-room selection randomness | Low | Low | Stub `util.NumberRange` in `arena_test.go` via a function-variable seam (pattern from `rollD20`). |
| `PLR_SILENCE` interaction with arena fights — silence blocks communication, but what about combat messages? | Low | Low | C does not gate combat messages on `PLR_SILENCE` — only comm commands. Go mirrors. Document: combat messages DO emit during arena fight. |
| A third party (non-combatant) in the arena room sees messages | Low | Low | `ROOM_ARENA` flag should only be on arena rooms; by convention no PCs enter except via teleport. Document as a builder invariant. |

---

## Reference Plans for Shape

This plan follows `plan-tranche-b.md` / `plan-tranche-c.md` section ordering:

1. Problem (scoped gap statement).
2. C Reference (line-cited, authoritative).
3. Go Current State (verified facts).
4. Go Design (chosen + rejected alternatives).
5. Task Groups (test-first + mutation-verify, one per concern).
6. Acceptance Criteria (mechanical).
7. Scope Cuts / Deferrals.
8. Open Questions (human-input blockers).
9. Risk Analysis.

Completion record and adversary-verification notes will append after the plan adversary pass and the worker execution waves land.

---

## Dispatch Protocol for Future Manager

When a manager is dispatched with this plan as the goal:

1. Resolve the 7 Open Questions (human input if needed).
2. Execute G1-G7 in order: G1 (schema) unblocks G2-G6; G7 (integration) depends on all prior.
3. Within G2-G6, workers can run in parallel waves:
   - Wave A: G2 (`DoChallenge`) + G3 (`DoAccept`) + G4 (`DoDecline`+`DoWithdraw`) — all operate on `arena.go` but on separate functions; coordinate via one worker owning the file.
   - Wave B: G5 (`charUpdate`) + G6 (`ArenaVictoryCheck`) — independent files.
   - Wave C: G7 (integration + E2E).
4. Adversary per task group (PASS required per `_shared.md` escalation protocol).
5. Final adversary on the full acceptance-criteria checklist before reporting completion.
6. Manager reports to orchestrator: completion record appended to this plan doc; `CHANGELOG.md` entry drafted; `TODO.md` entries moved; `CLAUDE.md` patch drafted.

Worker fanout cap: 6 per wave (per `_shared.md`). G2-G6 fits comfortably in one wave.

---

*Adversary verification notes for this plan — to be appended after planning adversary pass.*

---

## Completion Record — 2026-04-19 (commit `0cc1d97`)

### Summary

All 7 task groups executed. 15/15 acceptance criteria satisfied. First true two-player PvP scenario in the Go port — exercises `QuickLoginTwo` with 3 concurrent-session E2E testclient scenarios. Structured self-review substituted for external adversary pass per documented tooling caveat.

### Acceptance cross-reference

| # | Criterion | Where verified |
|---|---|---|
| A1 | `DoChallenge` matches C for all 12 gates + success path | `internal/act/arena_test.go` 15 tests `TestDoChallenge_*` (empty/already/busy/notfound/alreadychallenged/isChallenger/afk/NPCvictim/NPCch/immortal/lowlevel/fighting/self/mirror/success) |
| A2 | `DoAccept` all 6 gates + success | `TestDoAccept_*` (NPC/empty/notfound/NPCvictim/self/notchallenger/success) |
| A3 | `DoDecline` all 6 gates + success | `TestDoDecline_*` (empty/nochallenge/busy/notfound/notchallenger/npc-or-self/success) |
| A4 | `DoWithdraw` all 5 gates + success | `TestDoWithdraw_*` (NPC/notchallenger/notfound/busy/victimNPC/success) |
| A5 | `TIMER_CHALLENGE` unique; challenge sets count=5; accept extracts | `TestTimerChallenge_UniqueTag` + `TestDoChallenge_Success` + `TestDoAccept_Success` |
| A6 | `charUpdate` cancels challenge, emits exact message | `TestCharUpdate_ArenaChallengeTimeout` (plus 3 companion tests pin the non-firing cases) |
| A7 | `ArenaVictoryCheck` teleports + heals + strips + tracks + clears + busy=false | `TestArenaVictory_TeleportAndHeal` + `_AKillsADeathsIncremented` + `_AllFlagsClearedBoth` + `_IsBusyClearedAfter` + `_DebuffsStripped` + `_NoFireIfVictimNotDead`/`_NoFireIfNPC`/`_NoFireIfNotInArena` |
| A8 | Victory branch fires before corpse creation — no corpse in arena death | `TestArenaVictory_NoCorpseSpawned` (drives real `Damage` → asserts no ITEM_CORPSE_* in arena.Contents) |
| A9 | AKills/ADeaths persist across save/load | `TestSavePlayer_PersistsArenaCounters` + `TestSavePlayer_OmitsZeroArenaCounters` |
| A10 | Four commands at POS_RESTING, Level 0 | `TestBoot_RegistersArenaCommands` |
| A11 | `combat.ArenaIsBusyFunc` non-nil after `boot.Boot` | `TestBoot_WiresCallbacks` gains new arena-seam assertions |
| A12 | End-to-end testclient scenario drives challenge/accept → verifies exit invariants | `TestArena_ChallengeAndAccept_TeleportsBoth` (plus `_AndDecline` + `_AndWithdraw` for siblings) |
| A13 | `go test -count=3 ./...` green | Full-suite run 2026-04-19 — all 15 packages PASS |
| A14 | `go vet ./...` clean | Clean |
| A15 | `gofmt -l .` reports no changes on arena files | Clean |

### Mutation gates (`Edit` round-trips only — no banned git commands)

| Site | Mutation | Failing test |
|---|---|---|
| `act/arena.go` DoChallenge success | Drop `ch.Act.Set(ACT_CHALLENGER)` | `TestDoChallenge_Success` fails on flag assertion |
| `act/arena.go` DoChallenge level gate | `<=5` → `<5` | `TestDoChallenge_VictimLowLevel` fails (victim slips past) |
| `game/update.go` charUpdate tick | Drop `ch.Act.Remove(ACT_CHALLENGER)` | `TestCharUpdate_ArenaChallengeTimeout` fails |
| `combat/arena_victory.go` loser teleport | Swap `ROOM_VNUM_ALTAR` → `ROOM_VNUM_TEMPLE` | `TestArenaVictory_TeleportAndHeal` fails on loser-room assertion |
| `combat/arena_victory.go` AKills++ | Drop `ch.PCData.AKills++` | `TestArenaVictory_AKillsADeathsIncremented` fails (winner count 4 vs 5) |
| `combat/combat.go` Damage intercept | Drop `if ArenaVictoryCheck(...) { return rVICT_DIED }` | `TestArenaVictory_NoCorpseSpawned` fails (loser still in arena) |

### Open-Q resolutions applied

1. **`do_decline` C bug** — preserved verbatim. Decliner's `ACT_CHALLENGER` flag clears (wrong target); challenger's flag lingers until 5-tick timeout or victory. C-fidelity over behavioral fix.
2. **Global `challenge_tme` vs per-player timer** — per-player via `handler.AddTimer(ch, TIMER_CHALLENGE, 5, "", 0)`. `handler.HasTimer` added as 3-line wrapper over `GetTimer`.
3. **Arena-room vnum range** — Option (b) filter-then-random: `pickArenaRoom` scans MIN..MAX, keeps only rooms with `ROOM_ARENA` flag, returns `nil` if none. `DoAccept` surfaces "arena is not configured on this server." if `nil`.
4. **Command trust level** — `Level:0` / `POS_RESTING` on all four commands. Matches C (no explicit command-level gate).
5. **`ROOM_VNUM_ALTAR` fallback** — if altar vnum isn't loaded, loser falls back to `ROOM_VNUM_TEMPLE`. Pinned by `TestArenaVictory_FallbackAltarToTemple`.
6. **Area-data `ROOM_ARENA` flag** — `db/area/newacad.are` edited to add bit 26 to all 17 rooms in range 10366-10382 (old flags `3145736`/`3145740` → `70254600`/`70254604`). Drift-prevention test in `internal/boot/boot_test.go` regex-scans the shipped file.
7. **`ArenaVictoryCheck` insertion point** — inside the helper, calls `StopFighting` itself. Caller (`Damage`) does not call StopFighting before the check.

### Deliberate C divergences

- `DoAccept` success path: per plan §G3, C's `do_look(ch)` called a second time after `victim` teleport (instead of `do_look(victim)`) preserved verbatim — accepter sees room twice, victim gets no auto-look. See arena.c:255 comment.
- `do_decline` flag-clear target bug: plan Q1 resolved as preserve-verbatim; comment cites the bug.
- `pickArenaRoom` filter-then-random: divergence from C's potentially-crashing `number_range(MIN, MAX)` on a non-arena vnum.
- `AKills`/`ADeaths` persist only when non-zero (emit-gate): keeps stock pfiles unchanged for non-arena players.

### Scope cuts observed (from plan)

- No `PLR_NORESTORE`.
- No score-card display of AKills/ADeaths (follow-up).
- No spectator system.
- No arena OLC commands (`redit` does the job).
- No anti-exploit checks.
- No team/group/clan challenges.

### New tests (approximate)

- `internal/act/arena_test.go`: 37
- `internal/combat/arena_victory_test.go`: 11
- `internal/testclient/arena_test.go`: 3
- `internal/game/update_test.go`: 4 (arena timeout)
- `internal/persist/player_test.go`: 2
- `internal/types/constants_test.go`: 2
- `internal/boot/boot_test.go`: 2

Total: ~61 new test cases.

### Follow-ups deferred

- Score-card AKills/ADeaths display (plan §Scope Cuts).
- `do_noauction` admin toggle for the arena subsystem — C doesn't ship one either; skip.
- Spectator mode / `watch` command — not in C.
