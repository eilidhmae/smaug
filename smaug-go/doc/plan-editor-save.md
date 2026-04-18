# Plan: Editor `/s` Save — Callback Mechanism + `CON_PLAYING` Transition

**Status:** Planned (2026-04-17). Adversary-verified research: PASS (one fabricated fact flagged — "81 callers" was ~30).
**Priority:** P2 — unblocks `DoBio` / `DoDescription` and all Phase-6 OLC substates.
**Scope:** `internal/game/editor.go`, `internal/types/character.go`, `internal/act/olc.go`, tests.

---

## Problem

`internal/game/editor.go:156-160` handles the `/s` command by sending `"Done.\n\r"` and returning. It does **not** transition `ch.Desc.Connected` out of `CON_EDITING`, and it does **not** invoke the caller's save logic. The descriptor is stuck: every subsequent line the player types is routed back to `EditBuffer` (see `loop.go:261-264`).

The only editor test for `/s` — `TestEditBuffer_SaveCommand` at `editor_test.go:472-479` — is an empty stub. It asserts nothing, so the bug shipped untested.

Downstream consequences:
- `DoBio` / `DoDescription` (C `player.c:3366-3464`) cannot ship.
- Interactive `CON_OEDITING` / `CON_MEDITING` substates in OLC cannot ship.
- `redit` currently appears to work (C's `redit` flow DOES land in the editor), but after `/s` the builder's session is silently wedged — confirm during implementation.

## C Reference (authoritative)

- **`start_editing`** — `src/build.c:938-1008`. Caller sets `ch->last_cmd` (a `DO_FUN *` function pointer) **before** calling `start_editing`, then allocates the editor buffer, sets `ch->desc->connected = CON_EDITING`, and returns.
- **`stop_editing`** — `src/build.c:1047-1063`. Disposes the editor, clears `ch->dest_buf` / `ch->spare_ptr`, sets `ch->substate = SUB_NONE`, sets `ch->desc->connected = CON_PLAYING`.
- **`edit_buffer` `/s` handler** — `src/build.c:7004-7010`:
  ```c
  if (!str_cmp (cmd + 1, "s"))
  {
      d->connected = CON_PLAYING;     // transition FIRST
      if (!ch->last_cmd) return;
      (*ch->last_cmd)(ch, "");        // re-invoke original command
      return;
  }
  ```
- **Example callback** — `src/player.c:3413-3464` (`do_bio` with `SUB_PERSONAL_BIO`):
  ```c
  case SUB_PERSONAL_BIO:
      STRFREE(ch->pcdata->bio);
      ch->pcdata->bio = copy_buffer(ch);
      stop_editing(ch);
      return;
  ```

Key invariant: `/s` transitions the descriptor, then re-invokes the original command. The command dispatches on `ch->substate` to take the "save and clean up" branch, which calls `copy_buffer` to extract the text and `stop_editing` to release the editor.

## Go Design

Three approaches considered:

**A. Function-pointer closure on `CharData`, threaded via `StartEditing` signature change.** Add `EditorSave func(*types.CharData)` to `CharData`; update `StartEditing(ch, text, save)` and the boot-wired function-variable type. Cleanest match to C but forces coordinated changes in `olc.go`, `boot.go`, and `boot_test.go`.

**B. Structured callback context.** Add `EditorCtx interface{}` + `EditorSave func(*types.CharData)` for multi-field contexts. Rejected — closures already capture context.

**C. (ADVERSARY-RECOMMENDED) Function-pointer closure on `CharData`, set at call site before invoking `StartEditingFunc`.** Add `EditorSave` to `CharData`, but leave `StartEditing` signature unchanged. Calling command (`DoRedit`) does `ch.Substate = ...; ch.EditorSave = func(ch) { ... }; StartEditingFunc(ch, text)` — zero signature churn, zero boot-wiring change, `boot_test.go:97` and `:439` stay green without modification.

**Recommendation: C.** Adversary flagged that Option A forces coordinated changes across `olc.go:18` (`var StartEditingFunc func(ch *types.CharData, text string)`), `boot.go:133` (`act.StartEditingFunc = game.StartEditing`), `boot_test.go:97` (nil-reset), and `boot_test.go:439` (regex assertion). Option C avoids all four coordination points. Two extra lines at each call site is worth it.

## Task Groups

### G1 — Add `EditorSave` callback field to `CharData`

- File: `internal/types/character.go`.
- Add field near existing editor fields (`Editor`, `Substate`, `InterEditing`, `InterEditingVnum`):
  ```go
  EditorSave func(*CharData) // invoked on /s after CON_PLAYING transition
  ```
- **Test first:** Add `TestCharData_EditorSaveField` asserting the field exists and is nilable. Mutation-verify by removing the field → test fails to compile.

### G2 — Callers set `EditorSave` before invoking `StartEditingFunc` (signature unchanged)

- File: `internal/act/olc.go` (and any future editor user).
- Do NOT change `StartEditing` signature. Do NOT change `StartEditingFunc` type declaration.
- Caller pattern:
  ```go
  ch.Substate = types.SUB_ROOM_DESC
  ch.EditorSave = func(ch *types.CharData) {
      room.Description = game.CopyBuffer(ch)
      game.StopEditing(ch)
  }
  StartEditingFunc(ch, room.Description)
  ```
- `StartEditing` may optionally clear `ch.EditorSave` at session end (defensive; prevents leaks if someone uses `StartEditing` without setting one).
- **Test first:** `TestDoRedit_SetsEditorSave` — verify after calling the redit entry point, `ch.EditorSave` is non-nil.
- **Unchanged files:** `internal/game/editor.go` signatures, `internal/boot/boot.go:133`, `internal/boot/boot_test.go:97`, `internal/boot/boot_test.go:439`. All stay green.

### G3 — Fix `/s` handler: transition + invoke callback

- File: `internal/game/editor.go`, lines 156-160.
- New body:
  ```go
  case 's':
      if ch.Desc != nil {
          ch.Desc.Connected = types.CON_PLAYING
      }
      if ch.EditorSave != nil {
          save := ch.EditorSave
          ch.EditorSave = nil  // one-shot; prevents double-invoke
          save(ch)
      }
      return
  ```
- Do NOT call `StopEditing` here — the callback is responsible (matches C). But DO clear `EditorSave` first so a callback that re-enters `StartEditing` (nested edit) works correctly.
- **Test first:** `TestEditBuffer_SaveInvokesCallbackAndTransitions` — start an edit with a spy callback, send `/s`, assert:
  1. `ch.Desc.Connected == CON_PLAYING`
  2. Callback was invoked exactly once
  3. Callback received the same char
- **Mutation:** delete the transition line → test fails. Delete the callback invocation → test fails.

### G4 — Wire `DoRedit` call sites to set `EditorSave` before calling `StartEditingFunc`

- File: `internal/act/olc.go` at lines 53 and 116 (the two `StartEditingFunc` call sites — grep verified these are the only two).
- Immediately before each `StartEditingFunc(ch, ...)`, set `ch.EditorSave` to a closure capturing the room (or extra-description entry). The closure calls `game.CopyBuffer(ch)`, writes to the captured target, and calls `game.StopEditing(ch)`.
- **Flag this as a pre-existing bug fix.** `redit desc` currently wedges any builder session — this is a regression fix, not a new feature. Any builder who has used `redit desc` has hit it.
- **Post-transition prompt caveat.** After `/s` transitions to `CON_PLAYING`, the pulse loop's `CON_PLAYING` branch (`loop.go:252-259`) has already run for this tick — the next prompt will not fire until the next input. **Decision:** G3's callback should emit a trailing newline or explicit prompt after `StopEditing` to avoid a blank line after `Done.` Concretely: callbacks end with `ch.Send("\n\r")` after `StopEditing`. Alternative: `StopEditing` itself emits a prompt. Simpler to keep it in the callback for this plan.
- **Test first:** `TestDoRedit_DescSaveRoundTrip` — `redit desc` → enter editor → `/s` → verify `room.Description` contains the entered text and descriptor is back to `CON_PLAYING`.

### G5 — Add editor integration tests beyond `/a`

- File: `internal/game/editor_test.go`.
- Replace the empty `TestEditBuffer_SaveCommand` stub with real assertions. Note: the stub's existing comment ("Editor should still exist (caller clears it)") encodes the OLD/broken assumption; replace both comment and body.
- Add `TestEditBuffer_SaveCallsCallback_ClearsState` — end-to-end: start, write text, `/s`, verify `ch.Editor == nil`, `ch.Substate == SUB_NONE`, `ch.Desc.Connected == CON_PLAYING`, text was saved.
- Add `TestEditBuffer_SaveNoCallback_StillTransitions` — protects against the edge case where no callback was set (no panic; `CON_PLAYING` set).
- Add `TestEditBuffer_AbortDoesNotInvokeCallback` — `/a` must not call the save.
- Mutation-verify each: flip the assertion target, confirm test fails, revert.

### G6 — Update `TODO.md`, `plan-player-config.md` R6, and this plan

- `TODO.md`: move "R6 editor.go /s" follow-up to the Done section with 2026-MM-DD completion date.
- Unblock the `bio` / `description` follow-up by changing its status from blocked to ready.
- Append a completion record to this plan file.

## Acceptance Criteria

A1. `TestEditBuffer_SaveCommand` passes with real assertions (not a stub).
A2. After `/s`, `ch.Desc.Connected == CON_PLAYING` for every test case.
A3. `EditorSave` callback is invoked exactly once per `/s`.
A4. `EditorSave` is cleared before invocation (prevents double-invoke on re-enter).
A5. `redit desc` round-trip works: edit, `/s`, room description reflects the change, builder can issue the next command without reconnecting.
A6. `/a` still aborts without invoking the callback.
A7. `go test -count=3 ./internal/game/... ./internal/act/...` green.
A8. `go build ./...` green (no broken callers of `StartEditing`).

## Scope Cuts / Deferrals

- `dest_buf` / `spare_ptr` generic context fields — not added. Closures capture context; add fields only if a specific caller needs them.
- Editor color-state reset (C's `set_char_color(AT_PLAIN)`) — not required; Go doesn't track per-descriptor color the same way.
- `/q` command — does not exist in C either; do not add.

## Open Questions

1. Should `StopEditing` still nil the `EditorSave` field (defensive), or trust that `/s` / `/a` already cleared it? **Recommendation:** Have `StopEditing` nil it defensively — cheap, prevents a leaked closure from holding a large context.
2. What if the callback itself starts a nested edit (`StartEditing` inside the save handler)? The one-shot pattern in G3 (clear before invoke) handles this correctly — the nested `StartEditing` overwrites the cleared field.
3. Should we add a dedicated save-only callback alias (`OnSave`) vs the generic `EditorSave`? **Recommendation:** Just `EditorSave` — single purpose, no renaming.

## Risk

- **Medium.** Editor is a core subsystem used by `redit` today. `redit desc` is currently BROKEN end-to-end (pre-existing regression) — this plan fixes it. G5 integration tests must prove the redit round-trip works before this lands.
- Adversary noted: the research's "81 C callers of `start_editing`" figure was inflated — actual count is ~30 across ~15 files. This does not affect the plan, but narrows the Phase-6 deferred work scope (mainly: `player.c` for bio/desc; `omedit.c`/`oredit.c` for OLC; `boards.c`, `ban.c`, `news.c`, `deity.c`, `overland.c` for content editing).
- Panic in callback leaves `ch.Editor` non-nil even though `CON_PLAYING` was set. Go's single-goroutine game loop would panic-crash the whole server, so this is academic — but the one-shot clear + defensive `StopEditing` nil-pattern prevents leaked closures holding large contexts.

## Adversary-Resolved Concerns (2026-04-17 plan review)

1. **Signature change blast radius** — resolved by switching to Option C (call-site assignment). `StartEditing` / `StartEditingFunc` signatures unchanged. Adversary called this out as the primary concern.
2. **Post-transition prompt gap** — addressed in G4 by having callbacks send a trailing prompt after `StopEditing`.
3. **Existing test comment misleading** — addressed in G5 (rewrite comment + body).
