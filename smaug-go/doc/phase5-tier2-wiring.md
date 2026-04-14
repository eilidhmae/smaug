# Phase 5 — Tier 2: Flag Honoring & Wiring Idle Data

## Goal

Many SMAUG game behaviors are controlled by flags and data that *already load* from area files and player files but are never read by game logic. This tier makes that data do its job. No new subsystems — just plugging existing data into the right guard points. After Tier 2:

- Rooms with ROOM_NOMAGIC, ROOM_NOSUMMON, ROOM_NORECALL, ROOM_SILENCE, ROOM_NOFLOOR, ROOM_DEATH, ROOM_NOMOB actually enforce those rules.
- Wearing items aligned against your alignment/class is blocked.
- Exits marked EX_SECRET or EX_HIDDEN aren't given away by `exits`/`look`.
- AFK, NO_TELL, NO_EMOTE players behave as SMAUG expects.
- Language scrambling, repair shops, class/race bans, mail targeting, and clan storerooms become functional — their data already loads.

## Gap inventory (verified)

### Flag-honoring gaps

Status determined by grep in `smaug-go/internal/` excluding `types/`, `constants.go`, `enums.go`, and `*_test.go`.

| Flag | C enforcement site | Go status | Notes |
|------|--------------------|-----------|-------|
| ROOM_SAFE | `src/fight.c` (`check_safe`) | **Honored** at `act/combat.go:121` | — |
| ROOM_PRIVATE | `src/act_move.c` | **Honored** at `magic/magic.go:515` | Teleport blocker — good. |
| ROOM_DARK | `src/handler.c` | **Honored** at `handler/handler.go:163` | — |
| ROOM_NORECALL | spell/cmd recall | **Honored once** at `magic/magic.go:515` | Need to also check in `DoRecall` (`act/skills.go`) and `SpellWordOfRecall` (Tier 4). |
| ROOM_NOMAGIC | `src/magic.c` `do_cast` gate | **Defined only** | Add guard at top of `magic.DoCast`. |
| ROOM_NOSUMMON | `src/magic.c` `spell_summon` + recall guards | **Absent** | Block `SpellSummon` and `SpellWordOfRecall` (Tier 4). |
| ROOM_SILENCE | `src/act_comm.c` say/yell gates | **Defined only** | Gate `DoSay`/`DoYell`/`DoGossip`/`DoShout`. |
| ROOM_DEATH | `src/handler.c` char_to_room | **Defined only** | Kill on entry. |
| ROOM_NOMOB | `src/handler.c` | **Defined only** | Block NPC entry in movement helper. |
| ROOM_NOFLOOR | movement gate for non-flyers | **Defined only** | Block non-AFF_FLYING descent. |
| ROOM_NODROP | `src/act_obj.c` do_drop | **Defined only** | Gate `DoDrop` — the item falls back to char. |
| ROOM_SOLITARY | `src/handler.c` | **Defined only** | Block second occupant. |
| ROOM_ARENA | PK resolution | **Defined only** | Tier 2 scope: optional — Phase 5 proper handles arena fully; for now gate `ROOM_SAFE` exemption. |
| AFF_BERSERK | combat bonus | **Defined only** | Apply hit/damroll modifier in `combat.OneHit`. |
| AFF_FEARED | movement force-flee | **Absent** | Lower priority. |
| PLR_AFK | `src/act_comm.c` do_tell | **Defined only** | Show "(afk)" or message in `DoTell`. |
| PLR_NO_TELL | `src/act_comm.c` do_tell receiver | **Defined only** | Refuse incoming. |
| PLR_NO_EMOTE | emote receiver | **Defined only** | Refuse incoming. |
| PLR_COMPACT | blank-line suppression | **Absent** | Skip blank lines before prompts. |
| ITEM_ANTI_EVIL/GOOD/NEUTRAL | `src/act_obj.c` wear gate | **Defined only** | Gate `DoWear`/`DoWield`. |
| ITEM_ANTI_MAGE/CLERIC/THIEF/WARRIOR | same | **Defined only** | Same gates; class check. |
| ITEM_POISONED | drink/eat path | **Defined only** | Gate in `DoEat` / `DoDrink` / weapon poison on hit. |
| ITEM_HIDDEN | reset-time placement | **Honored** at `handler/reset.go:247` | — |
| EX_SECRET | `exits` display | **Defined only** | Hide from `showExits` in `act/info.go`. |
| EX_HIDDEN | same | **Defined only** | Hide from `showExits`. |
| EX_NOMOB | NPC movement | **Defined only** | Gate NPC-branch of `MoveChar`. |
| EX_PASSAGE | open passage auto-close | **Defined only** | Lower priority; used by mp_open_passage (Tier 3). |
| EX_WINDOW | look-through | **Defined only** | Lower priority; used by `look <dir>`. |
| ACT_PACIFIST | can't be forced to fight | **Defined only** | Gate in aggressive and mpkill paths. |
| ACT_HUNTER | hunt on sight | **Absent** | NPC hunting code exists (`act/track.go`) but flag not checked to *start* hunt. |
| ACT_TRAIN | trainer NPC for `practice` | **Defined only** | Gate `DoPractice` trainer search. |
| ACT_PRACTICE | same | **Defined only** | Same. |
| ACT_STOP_SCRIPT | mudprog halt | **Defined only** | Honored by Tier 3's mudprog changes. |

### Data-loaded-but-idle subsystems

| Subsystem | Data present | Missing wiring |
|-----------|--------------|-----------------|
| **Languages** | `Speaks`/`Speaking` on `CharData` (`types/character.go:131–132`); `LANG_*` constants in `types/constants.go:163`; per-race `Language` loaded at `persist/races.go:111`; player file `Languages` saved and loaded at `persist/player.go:443` and `:107`. | No `DoSpeak` / `DoLearn` commands. No scrambler applied in `DoSay`/`DoTell`/`DoYell`/`DoShout`. Mob `Speaks` field loaded at `persist/area.go:413–425` but never consulted. |
| **Repair shops** | `RepairData` struct at `types/shop.go`; `loadRepairs` at `persist/area.go:1028–1052`; tests at `persist/area_test.go:685–718` confirm loading; `mob.RShop` assigned. | No `DoRepair` command. No `DoAppraise` for repair cost. No `findRepairKeeper`. |
| **Class/race bans** | `BAN_CLASS=2`, `BAN_RACE=3` in `types/constants.go:145–146`; `BanClassLevel`/`BanRaceLevel` fields on `SystemData` (`types/system.go:81–82`). | `ban.go` (audited) only branches on `"site"` / `"remove"` / `"list"`. No class/race/sex ban storage format, no `ban class`/`ban race` subcommands, no `CheckBans` variants that check class/race. |
| **Mail (targeted notes)** | `NoteData.ToList` loaded and written (`act/clan.go:266` displays it; `:291–292` defaults to "all"). | `note list` shows every note to every player at `act/clan.go:245–253`. No recipient-filter check. No "You have N unread messages" on login. |
| **Clan storerooms** | `ClanData.Storeroom int` at `types/clan.go:41`; loaded by `persist/subsystems.go:100–101`; test at `persist/subsystems_test.go:391`. | No storeroom commands. Storeroom vnum never resolved to a room pointer; no deposit/withdraw path. |

## Task groups

### G1 — Room-flag enforcement

Add guards at these sites (one line plus a send to the actor):

- `magic.DoCast` in `magic/magic.go`: before spell dispatch, check `ch.InRoom.Flags.IsSet(types.ROOM_NOMAGIC)` → "You cannot cast magic here." and return.
- `magic.SpellSummon` and `magic.SpellWordOfRecall` (Tier 4): check `ROOM_NOSUMMON` / `ROOM_NORECALL` on both source and target rooms.
- `act.DoRecall`: check `ROOM_NORECALL`.
- `act/comm.go` DoSay/DoYell/DoGossip/DoShout: check `ROOM_SILENCE` on the sender's room (not the listener's for gossip/shout — global channels bypass; see C reference).
- `act.MoveChar` in `act/info.go`: add checks for
  - ROOM_DEATH destination → call death/slay helper (minimum: HP = -10, position = POS_DEAD, reset to 1 HP like player death today).
  - ROOM_NOMOB destination + `ch.IsNPC()` → send "You can't go there." to messenger if any.
  - ROOM_NOFLOOR destination + direction == DIR_DOWN + `!ch.AffectedBy.IsSet(types.AFF_FLYING)` → send "You can't fly." and block.
  - ROOM_SOLITARY destination with another char present → block.
- `act/obj.go` DoDrop: honor ROOM_NODROP → "A magical force prevents you from dropping that." and keep in inventory.

**TDD:** one table-driven test per command with rows for flag-set vs flag-unset vs bypass.

### G2 — Item wear restrictions

In `act/obj.go` DoWear:
- If `obj.ExtraFlags` has `ITEM_ANTI_EVIL` and `ch.IsEvil()` → reject with "You are too evil to use that."
- Same for GOOD/NEUTRAL.
- If `ITEM_ANTI_MAGE` and `ch.Class == CLASS_MAGE` (etc. for CLERIC/THIEF/WARRIOR) → reject with "Your class cannot use that."

Reuse helper `ch.IsEvil/IsGood/IsNeutral` if present in `types/character.go`; if not, add them based on `ch.Alignment >= 350` / `<= -350`.

**Optional stretch:** burn on wear (C has a flag for this). Document as a follow-up, don't include in Tier 2.

**TDD:** table test at `act/obj_test.go`: each alignment × item flag combination.

### G3 — Exit secret/hidden/nomob

In `act/info.go` `showExits`:
- Skip exits where `exit.Flags.IsSet(types.EX_SECRET)` or `.IsSet(types.EX_HIDDEN)` unless viewer has `PLR_HOLYLIGHT` (already honored elsewhere).

In `act.MoveChar`:
- If `ch.IsNPC()` and the exit has `EX_NOMOB`, block with no message (NPC silently refuses).

**TDD:** update `act/info_test.go` `TestShowExits` to cover secret/hidden hidden-from-display; add movement test for NOMOB.

### G4 — Player-flag comm filters

In `act/comm.go`:
- `DoTell` sender path: after resolving the target, if target has `PLR_AFK` → still deliver but prepend "(afk)". If target has `PLR_NO_TELL` → refuse with "They are refusing tells.".
- `DoReply`: same PLR_NO_TELL check on the stored reply target.
- Emote receive loop: if recipient has `PLR_NO_EMOTE` → skip (don't send to them).
- `PLR_COMPACT`: in descriptor flush (`types/descriptor.go`) — if the flag is set, suppress the leading blank line before the prompt. Document as optional polish if descriptor flush is too risky to touch.

**TDD:** use `bufio`-backed fake descriptor in `act/comm_test.go` and assert presence/absence of the AFK prefix; refusal message; suppressed emote.

### G5 — Languages subsystem wiring

**Commands (new in `act/comm.go`):**
- `DoSpeak <language>` — set `ch.Speaking` to the matching LANG_ bit if the language is in `ch.Speaks`. Case-insensitive name match against the C `lang_names` table (porting to a Go map).
- `DoLearn <language>` — immortals teach; or skill/trainer-gated learning sets a bit in `ch.Speaks` (port from C).

**Scrambler (new in `util/translate.go`):**
```go
// Translate scrambles text that `listener` does not share a language with
// the `speaker` is speaking. Preserves color codes and whitespace.
func Translate(text string, speaker, listener *types.CharData) string { ... }
```
Port from C `translate()` — typically a deterministic per-language character substitution, preserving word lengths and spaces.

**Apply in:** `DoSay`, `DoTell`, `DoYell`, `DoShout`. Leave `DoGossip` global-channel as-is unless C scrambles gossip (check — it does not by default).

**TDD:** table-driven test in `util/translate_test.go`: speaker speaks Elven, listener speaks/doesn't speak Elven → text scrambled/preserved. Mutation check: break the bit-match → tests should fail.

### G6 — Repair shops

Create `smaug-go/internal/act/repair.go`:

```go
// DoRepair — estimate/accept repair of a carried item at a RShop keeper.
func DoRepair(ch *types.CharData, argument string) { ... }
// DoAppraise — alias for 'repair estimate' for legacy muscle memory. Also used by shops for buy-appraise.
func DoAppraise(ch *types.CharData, argument string) { ... }
```

Helpers:
- `findRepairKeeper(ch)` — scan `ch.InRoom.Contents` for an NPC whose `Mob.RShop != nil` (analogous to existing shop `findKeeper`).
- Compute cost: `cost = obj.Cost * (100 - obj.Condition*100/obj.MaxCondition) / 100 * keeper.RShop.ProfitFix / 100`. See `src/shops.c` for exact formula.
- Apply: instantly set `obj.Condition = obj.MaxCondition` on successful payment.

**Register in `cmd/smaug/main.go`** at standard mortal level.

**TDD:** tests in `act/repair_test.go` covering: no keeper; wrong item type; insufficient gold; successful repair. Use existing test fixture pattern from `act/shop_test.go`.

### G7 — Class/race ban enforcement

Extend `act/ban.go`:
- New subcommands: `ban class <name> [level]`, `ban race <name> [level]`. Store as `BanData` with a new `Type int` field on `types.BanData` (`BAN_SITE=1`, `BAN_CLASS=2`, `BAN_RACE=3` — already defined in `types/constants.go:145`).
- `CheckBans(w, site, class, race)` signature change or add `CheckClassRaceBan(w, ch)` called at character-creation time and at login.
- Enforce at `game/loop.go` in `CON_GET_NEW_CLASS` / `CON_GET_NEW_RACE` handlers to refuse class/race selection if banned.

Update `persist/ban.go` `LoadBanList` / `SaveBanList` to round-trip the new `Type` field.

**TDD:** `persist/ban_test.go` round-trip; `act/ban_test.go` for subcommand dispatch; `game/loop_test.go` for class/race rejection at nanny.

### G8 — Mail targeting

In `act/clan.go` `DoNote`:
- `note list`: filter by recipient match. A note is visible if `note.ToList == "all"` or the list contains `ch.Name` (case-insensitive tokenized). Reference C's `is_note_to` in `src/boards.c:61`.
- `note read <n>`: same filter.
- Login hook: on entering game, count unread notes and send "You have N new notes.\n\r" similar to C's behavior at `game/loop.go` enterGame.

**Optional:** persistent per-player "last read" index — C stores last-read note number in PCData. Document as stretch goal; MVP is show-don't-hide toggle.

**TDD:** `act/clan_test.go` (or `act/board_test.go` if split) — notes with ToList "all", "Alice", "Alice Bob" vs viewer "Alice"/"Carol".

### G9 — Clan storerooms

Minimal: add `DoClanDeposit <obj>` and `DoClanWithdraw <obj>` in `act/clan.go`. Storeroom is identified by `ch.PCData.Clan.Storeroom` (vnum → `WorldRef.GetRoom(vnum)`). Deposit moves obj to storeroom; withdraw the reverse. Permission check: `ch.PCData.ClanName == clan.Name`.

**Stretch:** automatic save of storeroom contents on server shutdown. Leave as follow-up; Tier 2 MVP is the commands plus volatile contents.

**TDD:** `act/clan_test.go` with a stub storeroom room.

## Critical files

**Modify:**
- `smaug-go/internal/magic/magic.go` — ROOM_NOMAGIC guard in DoCast; ROOM_NOSUMMON in SpellSummon.
- `smaug-go/internal/act/info.go` — MoveChar flag checks (ROOM_DEATH/NOMOB/NOFLOOR/SOLITARY); showExits SECRET/HIDDEN.
- `smaug-go/internal/act/comm.go` — ROOM_SILENCE gate in comm commands; AFK/NO_TELL/NO_EMOTE; DoSpeak/DoLearn; scrambler application.
- `smaug-go/internal/act/obj.go` — ROOM_NODROP in DoDrop; ANTI_* in DoWear/DoWield; ITEM_POISONED in DoEat/DoDrink.
- `smaug-go/internal/act/skills.go` — ROOM_NORECALL in DoRecall.
- `smaug-go/internal/act/ban.go` + `persist/ban.go` — class/race subcommands + persistence.
- `smaug-go/internal/act/clan.go` — note filtering; storeroom commands.
- `smaug-go/internal/game/loop.go` — class/race ban rejection during char creation; unread-notes hook on enterGame.

**Create:**
- `smaug-go/internal/util/translate.go` + `translate_test.go` — language scrambler.
- `smaug-go/internal/act/repair.go` + `repair_test.go` — repair shop commands.

**Register in `cmd/smaug/main.go`:** DoSpeak, DoLearn, DoRepair, DoAppraise.

**Reference (C):**
- `src/magic.c` for flag-gate locations in spells.
- `src/act_move.c` for ROOM_DEATH/NOFLOOR/NOMOB enforcement.
- `src/act_comm.c` for ROOM_SILENCE and PLR_AFK/NO_TELL/NO_EMOTE handling.
- `src/act_obj.c` for ITEM_ANTI_* wear checks and ROOM_NODROP.
- `src/shops.c` for repair formulas (`do_repair` and `do_appraise`).
- `src/ban.c` for ban types and CheckBans variants.
- `src/boards.c` for is_note_to.
- `src/act_comm.c` for `translate()` scrambler.

## Reused utilities

- `util.OneArgument`, `util.NumberArgument`, `util.IsName` for command argument parsing.
- `util.NumberPercent` for any roll-based mechanics.
- `WorldRef` for world lookups (clan by name, room by vnum, keeper finding).
- Existing shop helper pattern (`findKeeper`, cost calc) from `act/shop.go` is the template for repair equivalents.
- `ch.Sendf` and `ch.Send` for output — no new descriptor-side plumbing needed.
- `types.BitVector` `.IsSet` for flag checks everywhere.

## Verification

1. `go test ./...` passes. New tests: `util/translate_test.go`, `act/repair_test.go`, plus updates to `act/info_test.go`, `act/comm_test.go`, `act/obj_test.go`, `act/clan_test.go`, `persist/ban_test.go`.
2. Manual smoke test:
   - Enter a ROOM_NOMAGIC area → cast fails with the right message.
   - Wear an ITEM_ANTI_MAGE as a mage → blocked.
   - Speak a language your target doesn't share → output appears as gibberish to them, readable to you.
   - Visit a repair shop, repair a damaged weapon, see gold deducted and condition restored.
   - Ban a class, try to create a character of that class → blocked.
   - Write a targeted note, log in as a different player → not shown in `note list`.
3. Regression: tests for already-honored flags (ROOM_SAFE, AFF_POISON, AFF_SANCTUARY) remain green.

## Open questions / follow-ups

- **`Act()` dependency.** G1 adds many `ch.Sendf` calls. If Tier 1 lands first, some of these can use `Act()` for neater emote-style messages (ROOM_DEATH entry visible to bystanders, e.g., "Eilidh screams and falls dead!"). Not required for Tier 2 acceptance.
- **Storeroom persistence** deferred — document as a Phase-6 polish item.
- **`PLR_COMPACT` flush change** deferred if descriptor flush changes feel risky.
- **`ACT_HUNTER` auto-hunt-start** is a behavior change, not just a flag gate; batch with Tier 3 mudprog when hunt triggers get attention.
