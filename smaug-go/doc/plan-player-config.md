# Plan — Player-Config Commands (password / title / afk / save)

Source audit: `audit-2026-04-17.md` P1 "easy wins". TODO ref: `TODO.md:31-37`.

## Goal

Close four player-visible command gaps that are each a one-session task, and optionally knock out `pagelen` as a trivial alias. `bio` and `description` are **deferred** — they require string-editor substate dispatch that doesn't exist yet (see R6).

Four commands land:

1. **`save`** — manual pfile save (wraps existing `SaveFunc`).
2. **`afk`** — toggle `PLR_AFK` (flag already exists; `DoTell`/`DoReply` already honor it).
3. **`title`** — set own title, with length/tilde/color-token scrub.
4. **`password`** — change own password (bcrypt; divergence from C noted below).

Plus optional: **`pagelen`** as an alias for the existing `DoPager`.

## Per-command mapping

| Cmd | C ref | Go status | Fields | Save | Trust | Position |
|---|---|---|---|---|---|---|
| `save` | `act_comm.c:3277` | Missing; `SaveFunc` wired in `boot/boot.go:92` | entire pfile | yes (that's the point) | 0 | `POS_DEAD` |
| `afk` | `act_info.c:6020` | Missing (flag honored in `act/comm.go:57,89`) | `ch.Act` bit `PLR_AFK` | no | 0 | `POS_SLEEPING` |
| `title` | `player.c:3138` (via `set_title` at :3112) | Missing | `ch.PCData.Title` | no | 0 | `POS_DEAD` |
| `password` | `act_info.c:5078` | Missing | `ch.PCData.Pwd` | yes | 0 | `POS_DEAD` |
| `pagelen` | `ibuild.c:1440` | Partial — `DoPager` at `act/info2.go:130` already covers | `ch.PCData.PagerLen` | no | 0 | `POS_DEAD` |
| `bio` | `player.c:3413` | Missing; also persist.Bio writer missing (see R1) | `ch.PCData.Bio` | no | 5 | `POS_DEAD` |
| `description` | `player.c:3366` | Missing; blocked on editor substate | `ch.Description` | no | 0 | `POS_DEAD` |

## Existing Go seams

- **Command registration:** `reg.Register(&command.Command{Name, DoFun, Position, Level})` in `internal/boot/boot.go` (clusters at :534, :538, :572).
- **Dispatcher:** `Interpret` lowercases only `cmdWord`; raw `argument` preserves case (so `DoPassword` can read raw args).
- **Save path:** `(*GameLoop).SavePlayer` at `internal/game/loop.go:771`; exposed as `act.SaveFunc` (`act/info.go:335`); atomic tmp+rename landed in batch 1.
- **AFK honorers (verified):** `act/comm.go:57,89` (DoTell/DoReply prefix); test at `act/flags_test.go:349`. **`ban.go` does NOT honor it today** — the TODO line saying `ban.go` does is incorrect.
- **Bcrypt:** `game.BcryptCost = bcrypt.DefaultCost` (`game/loop.go:22`); login uses `CompareHashAndPassword` + auto-migration from legacy plaintext; test pattern at `game/loop_test.go:914,1356`.
- **Arg parsing:** `util.OneArgument` lowercases — **unsafe for passwords**. Add `util.CaseArgument` (case-preserving) as a prereq for G4.
- **Tilde scrub:** `util.SmashTilde` exists. **No `SmashColorToken` exists** — port as a prereq for G3.
- **Test helpers:** `makeTestChar(name)` + `readOutput(ch, client)` in `act/info_test.go:14,50`.

## Task groups

### G1 — `DoSave` (S, ~1h)

`act/playercfg.go` (new file). Body: NPC check, level-2 gate ("You must be at least second level to save."), `ch.Wait = 2`, call `SaveFunc(ch)`, `ch.Send("Saved...\n\r")`. Register in `boot.go`.

- **Tests** (`act/playercfg_test.go`): NPC no-op; level-1 rejected; level-2 invokes `SaveFunc` (spy via test-wired hook); message format.

### G2 — `DoAfk` (S, ~1h)

Same file. NPC check, toggle `ch.Act` `PLR_AFK`, self-message, room broadcast matching C `TO_CANSEE`. Register.

- C messages:
  - On: `"You are now afk.\n\r"` + room `"$n is now afk."`
  - Off: `"You are no longer afk.\n\r"` + room `"$n is no longer afk."`
- **Tests:** toggle on/off, NPC no-op, `DoTell` prefix regression still fires when receiver AFK.

### G3 — `DoTitle` (S, ~1–2h)

**Prereq:** port `util.SmashColorToken(s string) string` matching C `smash_color_token` (character-by-character `&` scrub). Tests in `util/strings_test.go`.

Body in `act/playercfg.go`: NPC check; `PCFLAG_NOTITLE` (`types/constants.go:657`) → "The Gods prohibit you from changing your title."; empty arg → "Change your title to what?"; truncate to 50 chars; `SmashTilde` + `SmashColorToken`; leading-space rule (prepend ` ` if starts with alnum, per C `set_title` at `player.c:3122`); assign to `ch.PCData.Title`; "Your new title has been set.".

- **Tests:** set, truncate, tilde scrub, color-token scrub, `PCFLAG_NOTITLE` blocked, NPC blocked, leading-space rule.

### G4 — `DoPassword` (M, ~3–4h)

**Prereq:** port `util.CaseArgument(argument) (first, rest string)` (case-preserving OneArgument). Tests.

Body:

- NPC check.
- Parse three case-preserved args: `<old> <new> <again>`.
- `bcrypt.CompareHashAndPassword([]byte(ch.PCData.Pwd), []byte(old))` — on mismatch, emit "Wrong password." and return.
- Confirm `new == again`; on mismatch emit "Passwords don't match; password not changed."
- Length gate: `len(new) >= 6` (see R3 — confirm with user).
- `bcrypt.GenerateFromPassword([]byte(new), game.BcryptCost)` → `ch.PCData.Pwd = string(hash)`.
- `SaveFunc(ch)`, log `"<name> changing password"`.
- Send "Ok.".

Register in `boot.go`.

**Divergence from C (see R2):** C takes only `<new> <again>` (no old-password check — the check is commented out). Go port requires old password. Rationale: modern security baseline; aligns with the bcrypt migration already shipped.

- **Tests:** success (pfile updated, hash starts with `$2`); wrong old; mismatch new; too-short new; NPC no-op; no args → usage. Test harness must set `game.BcryptCost = bcrypt.MinCost` (existing pattern).

### G5 — `pagelen` alias (XS, ~30min, optional)

Register `pagelen` as a second name pointing at existing `DoPager` (`act/info2.go:130`). `DoPager` already handles numeric set + toggle + 5–200 range.

**Test (new, dispatcher-level):** `TestInterpret_PagelenAlias` → `Interpret(ch, "pagelen 40")` must dispatch to `DoPager` (assert `ch.PCData.PagerLen == 40` after). Existing `info3_test.go:470` `TestDoPager_SetLength` only calls `DoPager` directly — it does NOT exercise the command-registry wiring, so a fresh test is required to prevent regression of the alias.

### G6 — `bio` + `description` (DEFERRED)

Blocked on string-editor substate dispatch — see R6.

## Acceptance criteria

1. **`save`**: level-1 char → "You must be at least second level to save."; level-2+ → "Saved..." and `SaveFunc` invoked. Pfile on disk has updated `Played` timestamp after disconnect+reconnect.
2. **`afk`**: first `afk` → `PLR_AFK` set + self/room messages; second → cleared + messages; a concurrent `tell` from Alice → Bob-AFK delivers with `(afk) Alice tells you '...'` prefix; NPC no-op.
3. **`title`**: `title the Slayer` → `ch.PCData.Title = " the Slayer"` (leading space); empty → "Change your title to what?"; `PCFLAG_NOTITLE` → prohibit message; tildes replaced; >50 chars truncated; color tokens scrubbed; NPC no-op.
4. **`password`**: `password oldpass newpass newpass` → bcrypt hash verifies against `newpass`; wrong old → "Wrong password."; mismatch → "Passwords don't match"; `<6` chars → too-short error; NPC no-op; log line emitted; save triggered.
5. **`pagelen`** (optional): `pagelen 40` sets `ch.PCData.PagerLen = 40`; <5 or >200 → usage error via `DoPager`.
6. **Regression:** full `go test ./...` green; existing `DoTell` AFK-prefix coverage still passes.

## Open questions / risks

- **R1 — Bio saver missing.** `persist/player.go:218` reads `Bio` but nothing writes it to the pfile. Blocks G6. If we ever land G6, add `fmt.Fprintf(w, "Bio %s~\n", util.SmashTilde(p.Bio))` at ~`persist/player.go:511` and a round-trip test.
- **R2 — Old-password prompt divergence.** C `do_password` takes only `<new> <again>`. We diverge to require `<old> <new> <again>`. **Confirm with user before coding G4.**
- **R3 — Minimum password length.** C enforces `>=5`. We bump to `>=6` (or `>=8`). **Confirm with user.**
- **R4 — `SmashColorToken` behavior.** C replaces `&`-prefix color runs with neutralized chars. Port should match C exactly; test against C-generated fixtures if available.
- **R5 — `update_aris` on save.** ~~C calls it before save to flush RIS affects. Not ported to Go. Low-impact for manual save. Leave as follow-up TODO.~~ **Audit correction 2026-04-18 (plan-tranche-c.md G5):** the original "not called before save_char_obj in DoSave" framing assumes a C-style architecture that the Go port does not use. C (src/handler.c:1221-1305) clears `affected_by`/`resistant`/`immune`/`susceptible` and rebuilds from scratch on every affect change; C saves *base-minus-equipment* stats (src/save.c:213 `de_equip_char` before `fwrite_char`) and re-runs `update_aris` on load (src/save.c:1078, 1238). Go instead mutates `ch.Hitroll` / `ch.Damroll` / `ch.Armor` *incrementally* via `handler.AffectModify(ch, aff, fAdd)` (internal/handler/handler.go:386-455) — called with `fAdd=true` from `AffectToChar` and `fAdd=false` from `AffectRemove`. `SavePlayer` writes the computed values; `LoadPlayer` reads them back verbatim. Neither forcing function (rebuild-from-scratch affected_by, de-equip-before-save) applies in Go. `update_aris` is therefore architecturally unnecessary and NOT a follow-up. The regression test `TestSaveLoadPlayer_AffectRemovalMaintainsStatInvariant` (internal/persist/player_affect_test.go) pins this invariant so a future refactor cannot silently drift the stat / affect-list relationship.
- **R6 — String-editor `/s` leaves descriptor in CON_EDITING permanently.** `internal/game/editor.go:156-160` sends "Done.\n\r" and returns **without calling `StopEditing`**. The descriptor stays in `CON_EDITING` forever; the player is stuck until they disconnect. (Compare `/a` at :152-153 which does call `StopEditing`.) This is a sharper problem than a missing write-back callback: the editor is fundamentally broken for `/s`. Blocks `DoBio`, `DoDescription`, and Phase-6 OLC substates. Deferred here; a standalone fix to `editor.go` should land before any editor-backed command.
- **R7 — AFK who-list indicator.** C `do_who` shows `[AFK]` on listings. Not in Go `DoWho`. Non-blocking for this batch; add follow-up task.
- **R8 — Audit claim about `ban.go` honoring AFK is incorrect** — document in commit message for accuracy.

## Rough total effort

G1+G2+G3+G5 ≈ one afternoon. G4 ≈ half a day (plus user confirmation on R2/R3). G6 is deferred entirely. Util prereqs (`SmashColorToken`, `CaseArgument`) are each <1h with tests.

## Relevant file paths

- C: `/home/eilidh/src/smaug/src/act_info.c` (5078, 6020), `/home/eilidh/src/smaug/src/act_comm.c:3277`, `/home/eilidh/src/smaug/src/player.c` (3112, 3138), `/home/eilidh/src/smaug/src/tables.c` (315, 1163, 1335, 1513)
- New Go files: `/home/eilidh/src/smaug/smaug-go/internal/act/playercfg.go`, `…/internal/act/playercfg_test.go`
- Util prereqs: `/home/eilidh/src/smaug/smaug-go/internal/util/strings.go` (add `CaseArgument`, `SmashColorToken`), `…/util/strings_test.go`
- Wiring: `/home/eilidh/src/smaug/smaug-go/internal/boot/boot.go` (register four commands + optional pagelen alias)
- Existing seams: `/home/eilidh/src/smaug/smaug-go/internal/act/info.go:335` (`SaveFunc`), `…/internal/act/comm.go:57,89` (AFK prefix), `…/internal/act/info2.go:130` (`DoPager`), `…/internal/game/loop.go:771` (`SavePlayer`), `…/internal/types/enums.go:979` (`PLR_AFK`), `…/internal/types/constants.go:657` (`PCFLAG_NOTITLE`), `…/internal/types/pcdata.go` (fields)

---

## Completion (2026-04-17)

All five in-scope task groups landed (G1 `save`, G2 `afk`, G3 `title`, G4 `password`, G5 `pagelen` alias) plus both util prereqs (`CaseArgument`, `SmashColorToken`). `go test -count=3 ./...` green across all 15 packages. G6 (`bio` / `description`) stays deferred on R6 as planned.

### Decisions confirmed from the user at kickoff

- **R2 — password argument shape.** Go port requires `password <old> <new> <again>`. This is a deliberate divergence from C — C `do_password` at `act_info.c:5078` takes only `<new> <again>` because the old-password check is commented out in-place. Rationale: aligns with the already-shipped bcrypt migration and meets a modern security baseline. Divergence is noted inline in `internal/act/playercfg.go` `DoPassword` docstring.
- **R3 — minimum password length = 6.** Up from C's 5. Message: `"New password must be at least six characters long."`.

### What landed

- **Util prereqs** — `internal/util/strings.go`:
  - `CaseArgument(argument) (first, rest string)` — case-preserving `OneArgument`, matches C `interp.c:1170` including quote handling (`'` / `"`), leading-whitespace trim, and trailing-whitespace trim on rest. 12 test cases in `internal/util/strings_test.go`.
  - `SmashColorToken(str) string` — ports C `db.c:4462` **accurately**: `^` → `-` AND `&` → `+` (the original plan described this as "character-by-character `&` scrub" which was a slight under-specification — actual C handles both introducers with distinct replacement chars). 11 test cases.
- **G1 — `DoSave`** (`internal/act/playercfg.go`). NPC early return, level<2 gate, `ch.Wait=2`, calls `SaveFunc(ch)`, sends `"Saved...\n\r"`. 4 tests (NPC noop, level-1 rejected, level-2 saves, nil-SaveFunc safety). `update_aris` deferred per R5.
- **G2 — `DoAfk`** (same file). NPC early return, `PLR_AFK` toggle via `ch.Act.Set`/`Remove`, self-message, `util.Act(..., TO_CANSEE)` room broadcast. 5 tests including the `DoTell` AFK-prefix regression.
- **G3 — `DoTitle`** (same file). NPC + PCData-nil guards, `PCFLAG_NOTITLE` gate, empty-arg check, 50-byte truncate (byte-wise matches C's `argument[50] = '\0'`), `SmashTilde` → `SmashColorToken` → `setTitle`. `setTitle` (port of `player.c:3112`) prepends a space iff the title starts with an alnum char. 9 tests.
- **G4 — `DoPassword`** (same file). NPC + PCData-nil guards; three-arg `CaseArgument` parse; partial/missing args print `"Syntax: password <old> <new> <again>."`; bcrypt verify against stored hash; mismatch check; `len(newPwd) >= 6` check; `bcrypt.GenerateFromPassword` at `BcryptCost`; log line (`"<name> changing password"` or `"<name> changing password from site <host>"`); `SaveFunc(ch)`; `"Ok.\n\r"`. 9 tests (success, wrong-old, mismatch, too-short, no-args, partial-args, NPC, log emitted, case preserved).
- **G5 — `pagelen` alias** (`internal/boot/boot.go`). Second `reg.Register` with `Name: "pagelen"` pointing at `act.DoPager`. Verified by `TestInterpret_PagelenAlias` in `internal/act/playercfg_test.go` — dispatches `pagelen 40` through `Interpret` and asserts `ch.PCData.PagerLen == 40`.

### Cross-package seam

- **`act.BcryptCost`** — new package-local var (`var BcryptCost = bcrypt.DefaultCost`) declared in `internal/act/playercfg.go`. `internal/boot/boot.go` `Boot` syncs `act.BcryptCost = game.BcryptCost` after the game loop is constructed so that `boot.TestOpts()` (which lowers `game.BcryptCost` to `bcrypt.MinCost`) also applies to `DoPassword`. Covered by a new `TestBoot_SyncsActBcryptCost` in `internal/boot/boot_test.go` with mutation verification.

### Tests and verification

- 29 new test cases in `internal/act/playercfg_test.go` (28) + `internal/boot/boot_test.go` (1 BcryptCost sync).
- 2 new test cases in `internal/util/strings_test.go` (`TestCaseArgument`, `TestSmashColorToken` — 23 subtests combined).
- Every gate mutation-verified: flip the `<`/`>`/`!=` → tests red → revert → tests green. Mutations exercised: `DoSave` level gate, `DoAfk` set path, `DoTitle` NOTITLE inversion, `DoPassword` old-pwd-check bypass, `SmashColorToken` replacement swap, `CaseArgument` lowercasing, `BcryptCost` sync omission.
- `go test -count=3 ./...` green across all 15 packages.
- Existing `TestDoTell_AFKPrefixRegression` in `act/flags_test.go:340` still green — new AFK toggle behavior drives the same downstream prefix.

### Intentional deferrals

- **`bio` / `description` (G6)** — blocked on R6: `internal/game/editor.go:156-160` handles `/s` by sending `"Done."` and returning **without** calling `StopEditing`, so the descriptor stays in `CON_EDITING` forever. Fix must land in `editor.go` before any editor-backed command. Captured in `TODO.md` follow-ups.
- **R1 bio saver missing** — pairs with G6; no writer for `Bio` in `persist/player.go:~511`. Follow-up noted.
- **R5 `update_aris` on save** — not ported (manual save is low-impact). Follow-up noted.
- **R7 `[AFK]` who-list indicator** — not ported. Follow-up noted.
- **R8 audit-doc correction (`ban.go` AFK)** — noted in `TODO.md`.

### Files touched

- New: `internal/act/playercfg.go`, `internal/act/playercfg_test.go`.
- Modified: `internal/util/strings.go`, `internal/util/strings_test.go`, `internal/boot/boot.go`, `internal/boot/boot_test.go`.
- Doc updates: this file (completion section appended), `smaug-go/doc/phases.md` (Tier 8 entry added), `CLAUDE.md` / `TODO.md` / `CHANGELOG.md` at the repo root.
