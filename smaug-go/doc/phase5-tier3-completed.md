# Phase 5 — Tier 3: Mudprog Depth (Completed)

**Landed 2026-04-14.** All 13 packages pass `go test -count=1 ./...` (2,205 top-level cases, up from 2,007 baseline). Two rounds of three-agent adversary quorum; all 10 load-bearing findings addressed before land.

## Summary by task group

### G1 — DoIfCheck: +77 if-checks
`internal/mudprog/ifcheck.go` grew from 29 to 106 keywords. Priority A (object/mob lookup, 20), B (char state, 21), C (room/exit state, 15), D (social/political, 11), E (misc, 10) — all ported. Ported-but-stubbed with `// TODO(tier3)`: `timeskilled`, `leverpos`, `isflagged`, `istagged`, `pkadrenalized`, `asupressed`, `areamulti`, `multi` — all need fields or helpers not yet present. Reworked the argument parser to disambiguate operator tokens (`==`, `/`, `!/`) from word arguments (`north`, `Ravens`).

### G2 — Mob trigger fires
New helpers in `internal/mudprog/triggers.go`: `TrigLogin`, `CheckVoid` + `TrigVoid`, `TrigTell`, `TrigHour`, `TrigTime`, `TrigSell`, `TrigHitprcnt`. New constants `MPROG_LOGIN/VOID/TELL/SELL` (bits 32–35). `MProgData.Type` widened from `int` to `int64`. Wired:
- `game/loop.go:enterGame` → TrigLogin after auto-look.
- `game/loop.go:closeDescriptor` → CheckVoid after room removal.
- `combat/combat.go:Damage` → HitprcntHook (wired from `cmd/smaug/main.go`).
- `combat/combat.go:StopFighting` → VoidHook.
- `act/comm.go:DoTell` → TrigTell.
- `act/shop.go:DoSell` → TrigSell.
- `game/update.go:weatherUpdate` → TrigHour/TrigTime on hour-boundary crossing (tracked via package-level `lastHour`).

### G3 — Object-prog subsystem
New `internal/mudprog/oprog.go` (250 lines). 19 fire helpers over a shared `ObjTrigger` dispatcher. Synthesizes a "supermob" per call (mirrors C `set_supermob`) to satisfy Driver's `mob != nil` requirement without mutating global room state. Call sites:
- `act/obj.go` — WEAR/REMOVE/SAC/GET/DROP.
- `act/info.go` — LOOK/EXA/Greet (objects in room); DoSay SPEECH.
- `act/comm.go` — Yell/Gossip/Shout SPEECH.
- `act/itemuse.go` — USE (Quaff/Recite/Brandish/Zap) + ZAP (Zap only).
- `act/repair.go:DoRepair` — REPAIR after successful repair.
- `combat/combat.go:Damage` → ObjDamageHook for each worn item (MVP approximation).
- `command/interpret.go` → ObjCommandHook (fires only when normal Find() returns nil — see fix 2 below).
- `game/update.go:objUpdate` — per-pulse RAND sweep over `world.Objects`.

### G4 — Room-prog subsystem
New `internal/mudprog/rprog.go` (265 lines). 13 fire helpers over `RoomTrigger`. Same supermob pattern, bound to the target room. New constants `MPROG_IMMINFO=1<<36`, `MPROG_CMD=1<<37`; aliases `MPROG_ENTER/RFIGHT/RDEATH/RGREET` reuse mob-prog bits per C's compat #defines. Keyword→bit mapping in `persist/area.go:mprogNameToType` extended for `enter_prog`, `leave_prog`, `sleep_prog`, `rest_prog`, `rfight_prog`, `rdeath_prog`, `imminfo_prog`, `cmd_prog`. Call sites:
- `act/info.go:MoveChar` — RprogLeave before CharFromRoom, RprogEnter after CharToRoom.
- `act/position.go:DoRest/DoSleep` — on position transition.
- `combat/combat.go:StartFighting` → RfightHook.
- `combat/combat.go:Damage` (POS_DEAD branch) → DeathRoomHook before ExtractChar.
- `act/wiz.go:DoRstat`, `teleportTo` (DoGoto/DoTransfer) → RprogImminfo.
- `act/info.go:DoSay`, `act/comm.go` → SPEECH alongside obj SPEECH.
- `command/interpret.go` → RoomCommandHook (fires only when normal Find() returns nil).
- `game/update.go:roomRandomUpdate` — per-pulse RAND scan of `world.Rooms`.
- `game/update.go` hourly boundary — RprogHour/RprogTime alongside TrigHour/TrigTime.

### G5 — mpsleep runtime
New `internal/mudprog/sleep.go` (75 lines). Package-level `sleepQueue`, `SleepAdd`, `SleepUpdate`, `SleepReset`. Driver recognizes `mpsleep <ticks>`, serializes post-sleep ComList + driver args + singleStep + ifLevel/ifState into `MProgSleepData`. Re-entrant: resumed progs may call `mpsleep` again without clobbering. Wired into `game/loop.go` pulse. Intentional divergences: `mpsleep` in a false-if branch does not queue (simpler than C's pre-if intercept); `MProgSleepData.Type` always `MP_MOB` since obj/room progs currently don't thread a caller-type.

### G6 — +35 mp commands
`internal/mudprog/commands.go` grew from 185 to 968 lines. Commands added: `mpasound`, `mpsound`/`mpsoundat`/`mpsoundaround`, `mpmusic`/`mpmusicat`/`mpmusicaround`, `mpechozone`, `mpmload`, `mpoload`, `mpinvis`, `mpat`, `mpadvance` (C-correct: 1-arg, always +1 level), `mpslay`, `mplog`, `mprestore`, `mpfavor`, `mpnuisance`/`mpunnuisance`, `mpbodybag` (world-wide corpse scan, vnum 11), `mppractice`, `mpopenpassage`/`mpclosepassage`/`mpfillin`, `mppeace`, `mppkset`, `mpoowner`, `mphunt`, `mpdeposit`/`mpwithdraw` (two-bucket economy with 1e9 chunk size), `mpapplyaffect`, `mpdelay` (WAIT_STATE on victim — distinct from `mpsleep`), `mpstrew`, `mpscatter` (C-correct: victim teleport), `mpdream`, `mpnothing`. Stubs left with `TODO(tier3)`: `mpmorph`/`mpunmorph` (no morph subsystem), `mpapply`/`mpapplyb` (auth state machine), `mphate` (rich hate list).

## Adversary quorum — findings resolved

Three parallel adversary subagents reviewed Tier 3 across correctness (A1), call-site wiring (A2), and architecture (A3). 7 load-bearing findings raised; all fixed with TDD before land:

1. **A3: `ProgTypes.Set(int(prog.Type))` silently dropped all Tier-3 bits** (area.go:507/748/891). `BitVector.Set` wants a bit-index 0–127; `prog.Type` is a bit-flag up to `1<<37`. Fixed by converting flag→index via `bits.TrailingZeros64`. The broken `IsSet(0) && len==0` short-circuit in `triggers.go` was simplified to `len==0`.
2. **A2: Obj/Room CMD hooks fired before normal command Find()** (`command/interpret.go`). Would let any room with a broad CMD prog swallow `quit`/`kill`/directional commands. Moved to the `cmd == nil` path.
3. **A1: `indoors` if-check missed ROOM_INDOORS flag.** Fixed: now returns true if either the flag is set or sector is SECT_INSIDE.
4. **A1: `mpScatter` scattered objects (wrong op).** C teleports a victim character to a random room in a vnum range. Rewrote to take `<victim> <low> <high>`, `CharFromRoom` + `CharToRoom` + `POS_RESTING`.
5. **A1: `mpAdvance` took invented 2-arg form.** C takes only `<victim>` and always advances by 1 level, clamped at LEVEL_AVATAR. Rewrote.
6. **A1: `mpDeposit/Withdraw` mutated HighEconomy raw.** C's `boost_economy`/`lower_economy` split at 1e9 boundary. Implemented two-bucket math with helpers.
7. **A1: `mpBodybag` scanned only mob's room.** C scans world object list for corpse vnum 11. Rewrote to iterate `WorldRef.Objects`.

Minor finding also fixed: `ischarmed` now matches `AFF_CHARM || AFF_POSSESS` (C parity).

### Round 2 — 3 additional load-bearing fixes

A second quorum pass raised further findings:

8. **`level` if-check read `Level` not `GetTrust`** (ifcheck.go). C `mud_prog.c:1128` uses `get_trust(chkchar)`. Fixed: now uses `chk.GetTrust()`. Immortal trust-level gates (`if level($n) >= 60`) now work correctly.
9. **`class` / `race` if-checks did numeric compare, C does string compare** (ifcheck.go). C `mud_prog.c:1144-1150` uses `mprog_seval(class_table[ch->class]->who_name, ...)`. Fixed: resolves class/race name via `WorldRef.Classes/Races` and calls `matchStr`. Preserved a numeric-fallback path so legacy progs that pass an integer still work. This was a critical find — every real area prog that gates on `if class($n) == mage` was silently failing.
10. **`cansee` missing blind + room-darkness checks** (ifcheck.go). C `handler.c:3394-3409` checks AFF_BLIND and `room_is_dark && !AFF_INFRARED`. Fixed: added both guards in `canSeeIfCheck`; ported `roomIsDark` helper mirroring C `handler.c:3183` (Light>0 / ROOM_LIGHT / ROOM_DARK / INSIDE+CITY exempt / SUN_SET+SUN_DARK darkness).

Two round-2 findings turned out to match C correctly on closer inspection of the C source:

- **mpAdvance at LEVEL_AVATAR**: A1 claimed C advances past AVATAR with special messages. `mud_comm.c:1422` explicitly returns at `>= LEVEL_AVATAR`. Go code was already correct.
- **mpSlay raw_kill**: A2 claimed C's `raw_kill` skips corpse creation. `fight.c:4081` shows `raw_kill` DOES call `make_corpse`. The Go implementation's ExtractChar-with-fPull path diverges from C (no corpse, no XP) but matches the spirit of "instant administrative kill". Documented as TODO(tier4) for full C fidelity if needed.

Additional latents flagged in round 2 and documented as TODO(tier4):
- `mortinroom` / `mortinworld` use exact-match (C uses `nifty_is_name` prefix match).
- `MPROG_SPEECH` "p " prefix for phrase-match is dropped in current matching.
- Sleep queue holds raw mob pointers; a partial defensive guard skips entries where `sd.Mob == nil || (InRoom == nil && IndexData == nil)`, but a generation counter is still the proper fix.

## Known deferrals (not load-bearing)

- `mpPeace` ignores per-target argument (always room-wide). C supports `mppeace all` and `mppeace <name>`.
- `mpStrew` invented semantics — C's version is commented-out dead code; Go version scatters object copies across an area (not a port of anything, but harmless).
- `OprogCommandTrigger` / `RprogCommandTrigger` return-false stubs. Scripts loaded with `cmd_prog` parse correctly but do not consume input until the command-matching logic is ported (needs MPROG_CMD keyword list semantics from C `rprog_wordlist_check`).
- `OprogDamageTrigger` fires for every worn item on each hit; C fires once per weapon-damage event on the specific item.
- `mpsleep` in a false-if branch does not queue (minor divergence from C).

## Files touched

**Created:** `internal/mudprog/oprog.go`, `oprog_test.go`, `rprog.go`, `rprog_test.go`, `sleep.go`, `sleep_test.go`, `ifcheck_tier3_test.go`.

**Modified:** `internal/mudprog/ifcheck.go`, `commands.go`, `commands_test.go`, `driver.go`, `triggers.go`, `triggers_test.go`; `internal/types/mudprog.go` (int64 widening, new constants); `internal/persist/area.go` (prog-bit fix, new keyword mapping); `internal/command/interpret.go` (command hooks); `internal/combat/combat.go` (5 new hook vars); `internal/game/loop.go`, `internal/game/update.go`; `internal/act/obj.go`, `info.go`, `comm.go`, `itemuse.go`, `repair.go`, `position.go`, `wiz.go`, `shop.go`; `cmd/smaug/main.go` (hook wiring).

## What's next (Tier 4)

Content breadth: missing spells, combat/utility skills, damage-message dispatcher, missing immortal/mortal commands, OLC interactivity. See `smaug-go/doc/phase5-tier4-content.md`.

Mudprog follow-ups that landed as deferrals can be picked up in Tier 4 as time permits: C-correct `OprogCommandTrigger`/`RprogCommandTrigger` wordlist matching, `mpapply`/`mpapplyb` auth state machine, `mpmorph`/`mpunmorph` (once the morph subsystem exists), richer hate-list for `mphate`, per-target `mppeace`.
