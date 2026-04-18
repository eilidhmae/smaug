# Plan: `DoGag` Player Command — Standalone Toggle

**Status:** Planned (2026-04-17). Adversary-verified research: PASS with minor doc-labeling nit.
**Priority:** P2 — `PCFLAG_GAG` already honored by `dammessage.go`; command is the only missing piece.
**Scope:** `internal/act/playercfg.go`, `internal/act/playercfg_test.go`, `internal/boot/boot.go`. Doc correction to TODO.md + plan-dammessage-gaps.md.

---

## Problem

`PCFLAG_GAG` (bit 5 of `PCData.Flags`) is defined in `internal/types/constants.go:652`, honored in `internal/combat/dammessage.go:342-349` (zero-damage miss suppression from Tier 7), saved/loaded through `persist/player.go:243` (load) and `:517` (save). But **no player-facing command sets it** — the flag is effectively test-only today.

## C Reference

- **C has no standalone `do_gag`.** The toggle is embedded in `do_config` at `src/act_info.c:5585-5833`. Gag specifically dispatches at `act_info.c:5794-5795` (`bit = PCFLAG_GAG`). Adversary correctly flagged that TODO.md and `plan-dammessage-gaps.md` mislabel `act_info.c:5795` as "`do_gag`" — it is a line inside `do_config`, not a function entry point.
- **NPC gate:** `act_info.c:5589-5590` (early `IS_NPC` return).
- **Syntax:** `config +gag` / `config -gag` (explicit +/-). Help text at `act_info.c:5601-5603` documents `'config +/- <keyword>'`.
- **Echo on toggle:** `Ok.\n` at `act_info.c:5826`.

## Go Design — Deliberate Divergence from C

C bundles 15+ flags into one `do_config` with strict `+`/`-` syntax. The Go port has already shipped standalone commands for the most-used flags (`DoAfk`, `DoTitle`, etc. in `internal/act/playercfg.go`). Shipping `DoGag` as a **standalone no-args toggle** matches that pattern and is simpler UX.

Documented divergences (each justified):
- **Syntax:** C requires `config +gag` / `config -gag`; Go exposes `gag` as a no-arg toggle, mirroring `DoAfk`'s local precedent. Simpler UX.
- **Message format:** C `do_config` always echoes `"Ok.\n"` (generic flag toggler has no knowledge of which flag). Go `DoAfk` uses directional messages ("You are now afk." / "You are no longer afk."). The **local Go convention wins** — ship directional messages:
  - On: `"Combat messages will be gagged.\n\r"`
  - Off: `"Combat messages will no longer be gagged.\n\r"`
- **Line-ending convention:** every existing `ch.Send(...)` in `playercfg.go` uses `\n\r` (newline-then-CR). Match that, not `\r\n`.
- **Toggle structure:** follow `DoAfk`'s conditional `if IsSet { Remove; send off-msg; return }; Set; send on-msg` — NOT XOR. Adversary noted that XOR has a mutation-testing blind spot (`^=` → `&^=` still passes a "toggle-off" test starting from flag-set state).

## Task Groups

### G1 — Implement `DoGag`

- File: `internal/act/playercfg.go`.
- Shape (mirrors `DoAfk` at `playercfg.go:40-53`):
  ```go
  func DoGag(ch *types.CharData, argument string) {
      if ch.IsNPC() { return }
      if ch.PCData == nil { return }  // load-bearing: IsNPC checks Act, not PCData
      mask := int(types.PCFLAG_GAG)
      if ch.PCData.Flags&mask != 0 {
          ch.PCData.Flags &^= mask
          ch.Send("Combat messages will no longer be gagged.\n\r")
          return
      }
      ch.PCData.Flags |= mask
      ch.Send("Combat messages will be gagged.\n\r")
  }
  ```
- ~10 LOC. Matches `DoAfk`'s conditional pattern exactly. Line-ending `\n\r` matches every other `ch.Send` in the file.
- `IsNPC()` (`types/character.go`) only checks `Act.ACT_IS_NPC` — does NOT check `PCData`. A malformed PC with nil `PCData` would pass `IsNPC()` and crash on `.Flags`. The nil guard is load-bearing, not cosmetic.

### G2 — Register in boot

- File: `internal/boot/boot.go`, near line 385 where `DoAfk` is registered.
- Add:
  ```go
  reg.Register(&command.Command{Name: "gag", DoFun: act.DoGag, Position: types.POS_DEAD, Level: 0})
  ```
- Position `POS_DEAD` matches C (mortal player setting, no physical-state requirement).

### G3 — Tests

- File: `internal/act/playercfg_test.go`.
- Mirror `TestDoAfk_TogglesOn` / `TestDoAfk_TogglesOff` / `TestDoAfk_NPCIsNoop`.
- `TestDoGag_TogglesOn` — PC with flag clear, call `DoGag`, assert `Flags & PCFLAG_GAG != 0` AND output contains `"Combat messages will be gagged."`.
- `TestDoGag_TogglesOff` — PC with flag set, call `DoGag`, assert `Flags & PCFLAG_GAG == 0` AND output contains `"Combat messages will no longer be gagged."`.
- `TestDoGag_NPCIsNoop` — NPC, call `DoGag`, assert no output, no panic. (NPC has `PCData == nil` but guard is IsNPC first; nil guard is the secondary defense.)
- `TestDoGag_NilPCDataIsNoop` — `ch.Act` does NOT have `ACT_IS_NPC` but `PCData == nil` (corrupted PC), assert no panic, no output.
- **Mutation verification matrix:**
  - `|=` → `&^=` on the set path → `TestDoGag_TogglesOn` fails (bit never gets set).
  - `&^=` → `|=` on the clear path → `TestDoGag_TogglesOff` fails (bit never gets cleared).
  - Remove `IsNPC` guard → `TestDoGag_NPCIsNoop` fails (output emitted or panic).
  - Remove nil PCData guard → `TestDoGag_NilPCDataIsNoop` fails (panic).
  - Swap the two messages → both toggle tests fail the output-content assertion.

### G4 — Doc corrections

- `TODO.md` line 98: change `"act_info.c:5795"` to `"act_info.c:5585 do_config (gag branch at :5794)"`.
- `smaug-go/doc/plan-dammessage-gaps.md:87` — same correction.
- Mark TODO item "Port `DoGag` player command" done after landing.

## Acceptance Criteria

A1. `DoGag` exists in `internal/act/playercfg.go` matching `DoAfk`'s shape (NPC gate, nil-PCData guard, conditional if-set-clear-else-set pattern, directional messages).
A2. Registered in `boot.go` at `level 0`, `POS_DEAD`.
A3. 4 tests pass, each mutation-verified.
A4. `go test -count=3 ./internal/act/...` green.
A5. Existing `dammessage_test.go` gag tests (at lines 398, 428, 456, 483) still pass — this command only sets the flag; it does not change how `dammessage.go` reads it.
A6. A PC who runs `save` after `gag` reloads with the flag intact (proves the round-trip works with the existing persist code).

## Scope Cuts

- **No C `+gag`/`-gag` explicit-prefix syntax.** Toggle-on-invoke.
- **No state display** (C's parent `config` prints the full flag state). Not shipping `config` at all; standalone per flag.
- **No `[GAG]` indicator on `who`.** Orthogonal; tracked elsewhere in TODO.md for AFK similarly.

## Open Questions

1. Command name: `gag` vs `gagcombat`? **Answer: `gag`.** Matches C's keyword. Short.
2. Should we also register `config` as a stub that errors? **Answer: No.** Each flag will get its own standalone command as needed. `config` itself is deliberately not ported.

## Risk

- **Low.** ~10 LOC of new code + 4 tests + 1 registration line + 2 doc-line edits. No architectural change. Flag already defined, consumed, and persisted.

## Adversary-Resolved Concerns (2026-04-17 plan review)

1. **XOR pattern was a mutation blind spot** — resolved by switching to `DoAfk`-style conditional (`if IsSet { Remove; ... } else { Set; ... }`) with directional messages. This also produces richer test assertions.
2. **`\r\n` vs `\n\r` line endings** — resolved: use `\n\r` everywhere to match every existing `ch.Send` in `playercfg.go`.
3. **Redundant nil guard** — clarified: guard is load-bearing because `IsNPC()` doesn't check `PCData`. Test added.
4. **TODO.md line 98 doc-correction target** — adversary's grep miss; line 98 in `/home/eilidh/src/smaug/TODO.md` DOES read `"Port DoGag player command (act_info.c:5795 ...)"`. Correction task remains valid.
