# Phase 5 — Tier 3: Mudprog Depth

## Goal

Bring the Go mudprog subsystem to content parity with C. Area authors write extensive scripts using if-checks, triggers, and mp commands; today most SMAUG areas relying on these scripts will fail silently when loaded into the Go port. After Tier 3:

1. The if-check vocabulary matches C (adds ~86 checks).
2. All missing mob trigger types fire (LOGIN, VOID, TELL, HOUR, TIME, SELL, HITPRCNT).
3. Object programs (18 trigger types) and room programs (13 trigger types) work end-to-end — parsing, storage, dispatch.
4. `mpsleep` delayed execution runs: scripts can schedule work for later ticks.
5. The ~35 missing mp commands are available to authors.

## Gap inventory (verified)

Audit summary (from Round-1 Explore + adversary re-verification):

| Dimension | C | Go | Gap |
|-----------|---|-----|-----|
| If-checks | ~115 distinct keywords in `mprog_do_ifcheck` | 29 `case` branches in `DoIfCheck` | **~86 missing** |
| Mob trigger types | ~13 fire-sites across `src/mud_prog.c` | 7 fire-helpers (`TrigGreet`, `TrigEntry`, `TrigSpeech`, `TrigFight`, `TrigDeath`, `TrigRand`, `TrigGive`) | **LOGIN, VOID, TELL, HOUR, TIME, SELL missing; HITPRCNT defined but never fires** |
| Object progs | 18 `oprog_*_trigger` functions in `src/mud_prog.c:3375–3788` | 0 (entire subsystem absent) | **All 18 missing** |
| Room progs | 13 `rprog_*_trigger` functions in `src/mud_prog.c:3944–4220` | 0 (entire subsystem absent) | **All 13 missing** |
| `mpsleep` runtime | `mpsleep_update` called from `src/update.c:2752`; queue globals in `src/mud_prog.c:131–133` | `types.MProgSleepData` struct exists (`types/mudprog.go:25`) — no queue, no update, no command | **Runtime absent** |
| `mp_*` commands | ~44 in `src/mud_comm.c` | 9 in `mudprog/commands.go` (mpEcho/EchoAt/EchoAround, mpGoto, mpTransfer, mpForce, mpKill, mpDamage, mpPurge) | **~35 missing** |

Good news: data model foundations already exist.
- `MPROG_*` bit constants for 24 trigger types (including LOOK/EXA/ZAP/GET/DROP/WEAR/REMOVE/SAC/SLEEP/REST/LEAVE/PULL/PUSH/SCRIPT/USE) are defined in `types/mudprog.go:44–77`.
- `MProgSleepData` has every field needed (Timer, Type `MP_MOB/MP_ROOM/MP_OBJ`, IfState, ComList, per-entity pointers).
- `MProgData` / `MProgActList` structs exist.
- The C `mob_index.MudProgs` field is parsed by `persist/area.go`. Object and room prog parsing likely needs extending (verify during G3).

## Task groups

### G1 — Fill out `DoIfCheck` (if-checks)

Add each of the following checks to `mudprog/ifcheck.go`. Port order grouped by player-visible impact:

**Priority A — World/object lookup (unlocks most existing scripts):**
- `mobinarea($i)` / `mobinroom($i)` / `mobinworld($i)` — count mob-vnum instances.
- `objinworld($i)` — count object-vnum instances.
- `ovnumhere($i)` / `otypehere($i)` — count in `ch.InRoom.Contents` + `ch.Carrying`.
- `ovnumroom($i)` / `otyperoom($i)` — count in `ch.InRoom.Contents`.
- `ovnumcarry($i)` / `otypecarry($i)` — count in `ch.Carrying` (not worn).
- `ovnumwear($i)` / `otypewear($i)` — count in `ch.Equipment`.
- `ovnuminv($i)` / `otypeinv($i)` — count in `ch.Carrying` (worn + unworn). (C distinguishes; follow the C semantics.)
- `objval0` through `objval5` — read specific `obj.Value[N]`.
- `wearing($i)` / `wearingvnum($i)` — by keyword / vnum.
- `carryingvnum($i)` — by vnum.

**Priority B — Character state:**
- `cansee($i, $n)` — visibility (reuse future or existing IsVisible helper).
- `ispacifist($i)`, `isriding($i)`, `ismounted($i)`, `ismorphed($i)`, `isnuisance($i)`.
- `ispkill($i)`, `canpkill($i)`, `isdevoted($i)`, `isthief($i)`, `isattacker($i)`, `iskiller($i)`.
- `isflagged($i, name)`, `istagged($i, name)` — variable data tags.
- `drunk($i)`, `hostdesc($i)`.
- `waitstate($i)`, `pkadrenalized($i)`, `asupressed($i)`.
- `favor($i)` (deity favor), `hps($i)`, `str/int/wis/dex/con/cha/lck($i)`.
- `numfighting($i)` — count of combatants.

**Priority C — Room / area / exit state:**
- `inroom($i, vnum)`, `wasinroom($i, vnum)`.
- `inarea($i, vnum)`, `areamulti($i)`, `multi($i)`.
- `indoors($i)`, `nomagic($i)`, `safe($i)`, `nosummon($i)`, `noastral($i)`, `nosupplicate($i)`, `norecall($i)`.
- `ispassage($i, dir)`, `isopen($i, dir)`, `islocked($i, dir)`.

**Priority D — Social/political:**
- `clan($i)`, `isclanleader($i)`, `isclan1($i)`, `isclan2($i)`, `isleader($i)`.
- `council($i)`, `deity($i)`, `guild($i)`, `clantype($i)`.

**Priority E — Miscellaneous:**
- `economy($i)`, `timeskilled($i, vnum)`, `leverpos($i)`, `number($i)`, `time($i)`, `rank($i)`.
- `mortinroom($i)`, `mortinarea($i)`, `mortinworld`, `mortcount`, `mobcount`, `charcount`.
- `ismobinvis($i)`, `mobinvislevel($i)`.
- `weight($i)`.

**Acceptance:** for each check, one table-driven test covering: match + non-match + missing subject + bad operator. Grouped tests in `mudprog/ifcheck_test.go` organized by priority tier above.

**TDD:** because the existing switch already has 29 cases, add tests first with expected behavior matching C semantics, watch them fail, then add the switch arm. Follow CLAUDE.md strict-TDD convention.

### G2 — Missing mob trigger fires

Target file: `mudprog/triggers.go` (adds the helpers) and call sites.

- **LOGIN (`MPROG_LOGIN`)**: fire in `game/loop.go` `enterGame()` after character placement, before MOTD flush. Iterate room + area mobs with `MudProgs` set, run matching progs.
- **VOID (`MPROG_VOID`)**: fire in `combat.StopFighting` / player disconnect path when no other characters remain in the room with the triggered mob.
- **TELL (`MPROG_TELL`)**: fire in `act/comm.go` `DoTell` after the message is delivered, on the target mob if it's an NPC with the trigger. Match on argument substring (similar to SPEECH).
- **HOUR (`MPROG_HOUR`)**: fire in `game/update.go` hourly tick — once per in-game hour boundary. Scan all mobs world-wide (cheap, one hour = ~70 real-time seconds * 4 pulses = rare event).
- **TIME (`MPROG_TIME`)**: similar to HOUR but specific hour argument.
- **SELL (`MPROG_SELL`)**: fire in `act/shop.go` `DoSell` after the shopkeeper accepts the item. The act-shop path already exists; add a trigger hook.
- **HITPRCNT (`MPROG_HITPRCNT`)**: in `combat.Damage` after HP update, iterate the victim's trigger list (if the victim is a scripted mob) and fire if new `hp * 100 / max_hp` dropped below the argument. Wire into existing fire path by extending the random/entry pattern.

**Acceptance:** each trigger has an end-to-end test: create a scripted mob with the relevant prog, perform the trigger action, observe side effect (commonly `mpecho` to room). `mudprog/triggers_test.go` gains one sub-test per trigger.

### G3 — Object-prog subsystem

**Data model:** Already fully present and loaded. `ObjIndexData.MudProgs []*MProgData` at `types/object.go:8`; `persist/area.go:747` already populates it during area parse. `MProgData.Type` already supports bit-flags including `MPROG_WEAR/REMOVE/SAC/LOOK/EXA/ZAP/GET/DROP/DAMAGE/REPAIR/PULL/PUSH/USE/SCRIPT`. What is missing is only the fire-helpers below and the call sites that invoke them.

**Fire-helper API (new in `mudprog/oprog.go`):**
```go
func OprogWearTrigger(ch *types.CharData, obj *types.ObjData) // WEAR
func OprogRemoveTrigger(ch *types.CharData, obj *types.ObjData)
func OprogSacTrigger(ch *types.CharData, obj *types.ObjData)
func OprogLookTrigger(ch *types.CharData, obj *types.ObjData)
func OprogExamineTrigger(ch *types.CharData, obj *types.ObjData)
func OprogZapTrigger(ch *types.CharData, obj *types.ObjData)
func OprogGetTrigger(ch *types.CharData, obj *types.ObjData)
func OprogDropTrigger(ch *types.CharData, obj *types.ObjData)
func OprogDamageTrigger(ch *types.CharData, obj *types.ObjData)
func OprogRepairTrigger(ch *types.CharData, obj *types.ObjData)
func OprogPullTrigger(ch *types.CharData, obj *types.ObjData)
func OprogPushTrigger(ch *types.CharData, obj *types.ObjData)
func OprogUseTrigger(ch *types.CharData, obj *types.ObjData)
func OprogGreetTrigger(ch *types.CharData) // for objects in room
func OprogSpeechTrigger(ch *types.CharData, argument string)
func OprogCommandTrigger(ch *types.CharData, argument string) bool // returns true if consumed
func OprogRandomTrigger(obj *types.ObjData)
func OprogActTrigger(actText string, obj *types.ObjData)
```
Each scans the object's progs, matches argument, translates variables, and calls `Driver` with `Type = MP_OBJ`.

**Call-site wiring:**
- WEAR/REMOVE → `act/obj.go` DoWear / DoRemove after equip/unequip.
- SAC → `act/obj.go` DoSacrifice before extraction.
- LOOK/EXA → `act/info.go` DoLook and DoExamine when the target is an object.
- ZAP/GET/DROP → `act/obj.go` (DoDrop, DoGet, and in `act/itemuse.go` DoZap/DoQuaff as relevant).
- DAMAGE → `combat/combat.go` Damage when an item worn by the victim takes damage (approximation in Go until weapon-damage modeling lands).
- REPAIR → Tier 2's `DoRepair` after successful repair.
- PULL/PUSH → future `DoPull` / `DoPush` commands; document as Tier 4 tie-in.
- USE → ItemUse (quaff/recite/brandish/zap).
- GREET → room-prog-style entry fire sites in `act.MoveChar`.
- SPEECH → all comm commands that broadcast to room.
- COMMAND → command interpreter fallback; similar shape to the existing social fallback. If an obj-prog consumes the command (returns true), skip normal dispatch.
- RANDOM → `game/update.go` objUpdate once per pulse per obj.
- ACT → whenever `Act()` (Tier 1) emits to an object's containing room.

**TDD:** `mudprog/oprog_test.go` with table: one fixture per trigger type, assert expected side effects.

### G4 — Room-prog subsystem

**Data model:** Already fully present and loaded. `RoomIndexData.MudProgs` at `types/room.go:31`; `persist/area.go:890` already populates it during area parse. C room-progs use types: `RPROG_ACT`, `RPROG_ENTER`, `RPROG_LEAVE`, `RPROG_SLEEP`, `RPROG_REST`, `RPROG_RFIGHT`, `RPROG_DEATH`, `RPROG_IMMINFO`, `RPROG_SPEECH`, `RPROG_COMMAND`, `RPROG_RANDOM`, `RPROG_TIME`, `RPROG_HOUR`. Fire-helpers and call sites are what's missing.

**Fire-helpers (new in `mudprog/rprog.go`):** same signature shape as G3; first argument is the room.

**Call-site wiring:**
- ENTER/LEAVE → `act.MoveChar` entry/exit.
- SLEEP/REST → Tier 2 position commands (`DoSleep`, `DoRest`).
- RFIGHT → `combat.StartFighting` if the room has the trigger.
- DEATH → `combat.Damage` on POS_DEAD in that room.
- IMMINFO → `act/wiz.go` DoGoto / DoRstat when an immortal enters.
- SPEECH → all comm commands.
- COMMAND → command interpreter fallback (before social, before obj-prog command).
- RANDOM → `game/update.go` per-pulse scan.
- TIME/HOUR → hourly tick.
- ACT → Act() emissions (Tier 1 integration).

**TDD:** `mudprog/rprog_test.go` per trigger.

### G5 — mpsleep runtime

Implement the delayed-execution queue in a new `mudprog/sleep.go`:

```go
// Package-level queue (single-threaded game loop, so no mutex).
var sleepQueue []*types.MProgSleepData

// SleepAdd queues a sleeping program to resume after `timer` pulses.
func SleepAdd(sd *types.MProgSleepData) { sleepQueue = append(sleepQueue, sd) }

// SleepUpdate is called once per pulse from the game loop.
// Decrements timers and resumes programs whose timers hit 0.
func SleepUpdate() {
    kept := sleepQueue[:0]
    for _, sd := range sleepQueue {
        sd.Timer--
        if sd.Timer > 0 {
            kept = append(kept, sd)
            continue
        }
        // Resume: resume Driver with IfLevel + IfState restored.
        resumeFromSleep(sd)
    }
    sleepQueue = kept
}
```

Extend `Driver` to recognize `mpsleep <ticks>` — serialize the remainder of the program into an `MProgSleepData` and add to queue, then return.

Wire `SleepUpdate()` into `game/update.go` pulse loop (alongside violenceUpdate, mobileUpdate, etc.).

**Acceptance:** a mudprog with `mpsleep 2` followed by `mpecho` causes the echo to appear two pulses later, not immediately. Test uses a fake `pulseOnce()` helper in `game/update_test.go`.

### G6 — Missing mp commands

Port the following into `mudprog/commands.go`. Group by complexity:

**Simple (echo/comm variants):**
- `mpasound <msg>` — echo to all connected players in the area.
- `mpsound` / `mpsoundat` / `mpsoundaround` — like mpecho but with MSDP sound payload. Minimum viable implementation: treat as mpecho variants until MSDP-sound is modeled.
- `mpmusic` / `mpmusicat` / `mpmusicaround` — same treatment as mpsound.
- `mpechozone <msg>` — echo to everyone in the same zone/area.

**Loading:**
- `mpmload <vnum>` — create mob instance in mob's room; use `handler.CreateMobile` + `handler.CharToRoom`.
- `mpoload <vnum> [level] [location]` — create object; use `handler.CreateObject` + `handler.ObjToRoom` or `ObjToChar`.

**State:**
- `mpinvis` — set `AFF_INVISIBLE` on mob.
- `mpat <vnum|name> <cmd>` — execute `cmd` at target room (parallel to `DoAt` from act/wiz.go).
- `mpadvance <target> <level>` — advance a player's level (restricted check).
- `mp_slay <target>` — instantly slay a character.
- `mp_log <msg>` — send to system log via `util.LogString`.
- `mp_restore <target>` — restore HP/mana/move.
- `mpfavor <target> <amount>` — adjust deity favor.
- `mpnuisance` / `mpunnuisance` — increment/decrement nuisance stages.
- `mpbodybag` — create corpse.
- `mpmorph <vnum>` / `mpunmorph` — morph/demorph target.
- `mp_practice <target> <skill>` — set a learned skill to adept.

**World:**
- `mp_open_passage <dir> <vnum>` / `mp_close_passage <dir>` / `mp_fill_in <dir>` — create/remove an exit from mob's room.
- `mppeace` — stop all combat in mob's room.
- `mppkset <target> <flag>` — set/clear PK status.
- `mpoowner <obj> <name>` — set object owner.
- `mphunt <target>` — set mob's `Hunting` to target (Tier 4 track integration).
- `mphate <target> <seconds>` — add to mob hate-list.

**Economy:**
- `mp_deposit <target> <amount>` — deposit gold into target's bank.
- `mp_withdraw <target> <amount>` — opposite.

**Affects:**
- `mpapply <target> <aff> <mod> <dur>` — apply an affect.
- `mpapplyb <target> <aff> <mod> <dur>` — broadcast version.

**Control:**
- `mpdelay <ticks>` — alias for `mpsleep`.
- `mpstrew <obj_vnum>` / `mpscatter` — place objects around area rooms.
- `mpdream <target> <text>` — wake-state-dependent message to target.
- `mpnothing` — no-op (explicit).

**Admin (`mpstat`, `opstat`, `rpstat`):** also needed as player-facing immortal commands in Tier 4. Leave the mudprog side of `mpstat` to Tier 4 — it's an inspection command, not a script action.

**Acceptance:** `mudprog/commands_test.go` gets one test per new command. Use minimal fixtures, assert the world-state mutation.

## Critical files

**Create:**
- `smaug-go/internal/mudprog/oprog.go` + `oprog_test.go`.
- `smaug-go/internal/mudprog/rprog.go` + `rprog_test.go`.
- `smaug-go/internal/mudprog/sleep.go` + `sleep_test.go`.

**Modify:**
- `smaug-go/internal/mudprog/ifcheck.go` — ~86 new case branches.
- `smaug-go/internal/mudprog/triggers.go` — add LOGIN/VOID/TELL/HOUR/TIME/SELL fire helpers; wire HITPRCNT.
- `smaug-go/internal/mudprog/commands.go` — add ~35 commands.
- `smaug-go/internal/mudprog/driver.go` — recognize `mpsleep` and intercept it before normal dispatch; export `Driver` in a form that sleep resumption can re-enter with restored if-state.
<!-- persist/area.go already parses mob, obj, and room progs at lines 506/747/890; no change needed there. -->
- `smaug-go/internal/types/object.go` / `types/room.go` — no change (MudProgs fields already present).
- `smaug-go/internal/game/update.go` — call `SleepUpdate()`, hourly HOUR/TIME triggers, random prog fires.
- `smaug-go/internal/act/obj.go`, `act/info.go`, `act/comm.go`, `act/move.go` (if it exists), `act/itemuse.go`, `act/wiz.go`, `combat/combat.go` — sprinkle call sites per G3/G4 plans.

**Reference (C):**
- `src/mud_prog.c` — full if-check and trigger code, especially `mprog_do_ifcheck` and `oprog_*_trigger` / `rprog_*_trigger` definitions.
- `src/mud_comm.c` — implementations of every mp command; port conservatively with C's behavior as the spec.
- `src/mud.h` — `MPROG_*`, `MP_*`, `RPROG_*` constants (already mirrored in `types/mudprog.go`; verify new ones if any are missing).
- `src/update.c:2752` — `mpsleep_update` invocation rhythm.

## Reused utilities

- `mudprog.Driver` already executes programs with if/or/else/endif + variable expansion — reuse for every new trigger and for sleep resumption.
- `mudprog.Translate` (`mudprog/translate.go`) handles `$n/$t/$o` token substitution.
- `handler.CreateMobile` / `handler.CreateObject` for load commands.
- `handler.CharToRoom` / `handler.ObjToRoom` / `ObjToChar`.
- `combat.Damage`, `combat.StopFighting`.
- `act.CmdRegistry` (via `mudprog.CmdRegistry`) for commands that invoke the interpreter.
- `WorldRef` (`mudprog.WorldRef`) for world lookups.

## Verification

1. `go test ./...` passes. Test counts: each G adds per-feature tests — expect ~200+ new test cases for Tier 3 alone.
2. Targeted integration: load a real `.are` area that uses mudprogs (e.g., `db/area/arachnos.are` or similar) and confirm no "unknown if-check" log entries during boot or gameplay.
3. Mutation-verify a sample of new ifchecks per CLAUDE.md convention: invert the return, tests fail, revert, tests pass.
4. Smoke-test with a crafted area file containing:
   - A mob with `RAND_PROG 25~ mpecho The sky darkens. ~` → see random messages.
   - A room with `ENTER_PROG 100~ mpecho You hear footsteps. ~` → on entry, message fires.
   - An object with `WEAR_PROG 100~ mpechoat $n You feel warmer. ~` → on wear, target-only message.
   - `mpsleep 2` + `mpecho` → echo appears 2 pulses later.

## Open questions / follow-ups

- **If-check count floor.** Adversary audit reported ~115 distinct C checks; some are compile-time-gated or area-codebase-specific (Shaddai, Gorog additions). Treat ~115 as the aspirational ceiling; MVP is to port the ~60 that appear in real area files bundled with `db/area/`.
- **MSDP sound/music variants** — until MSDP sound payloads are modeled, treat mpsound/mpmusic as mpecho aliases. Record as Phase-6 polish.
- **Object-prog command-consumption** (G3) conflicts with social-fallback dispatch: define the precedence (normal commands → obj-progs COMMAND → room-progs COMMAND → social fallback → "Huh?") and document it in `command/interpret.go` comments.
- **`mpsleep` re-entrancy** — the pattern used in C with a global current_mpsleep is fine in Go (single-threaded game loop), but be careful: if a sleeping prog spawns another `mpsleep`, ensure the new sleep record is queued post-append (not mutated while iterating). The `kept := sleepQueue[:0]` pattern in G5 handles this.
