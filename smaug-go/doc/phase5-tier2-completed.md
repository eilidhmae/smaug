# Phase 5 Tier 2 — Completed Work

## Summary

Tier 2 closes the flag-honoring and data-wiring gap identified in
`phase5-tier2-wiring.md`. All 9 task groups landed. The server now enforces
the room, item, exit, and player flags that previously loaded but did nothing,
and the languages, repair-shop, class/race-ban, note-targeting, and
clan-storeroom data that already sat in the world is now reachable through
actual commands and guard points.

83 source files, 66 test files, 1,505 top-level test cases across 13 packages
— all passing via `go test -count=1 ./...`.

---

## G1: Room-flag enforcement ✓

Guards added at the canonical command call sites:

- `act/magic.go` `DoCast` — `ROOM_NO_MAGIC` refuses the spell and does **not**
  deduct mana. Immortals bypass.
- `act/skills.go` `DoRecall` — `ROOM_NO_RECALL` refuses with a flavor
  message.
- `magic/magic.go` `SpellSummon` — caster-room `ROOM_NO_SUMMON` blocks the
  summon; target-room `ROOM_NO_SUMMON` or `ROOM_PRIVATE` makes the target
  unreachable.
- `act/info.go` `DoSay`, `act/comm.go` `DoYell`/`DoShout`/`DoEmote` —
  `ROOM_SILENCE` refuses with "You can't do that here."
- `act/obj.go` `DoDrop` — `ROOM_NODROP` keeps the item in inventory with the
  magical-force refusal message.
- `act/info.go` `MoveChar`:
  - `ROOM_DEATH`: entrant's Hit reset to 1 and they are recalled to the
    temple (when present), mirroring the "survivor with 1 HP" C behavior.
  - NPC + `ROOM_NO_MOB` destination: blocked before the room transfer.
  - `DIR_DOWN` + `ROOM_NOFLOOR` + no `AFF_FLYING`: blocked with "You can't
    fly."
  - `ROOM_SOLITARY` with another character already present: blocked.

Immortal bypass is consistent across every guard — a Trust of
`LEVEL_IMMORTAL`+ ignores the restriction.

## G2: Item wear restrictions ✓

New helper `wearRestrictMessage` in `act/obj.go`. `DoWear` consults it before
every equip:

- `ITEM_ANTI_EVIL` / `ITEM_ANTI_GOOD` / `ITEM_ANTI_NEUTRAL` checked against
  `ch.IsEvil()` / `IsGood()` / `IsNeutral()`.
- `ITEM_ANTI_MAGE` / `CLERIC` / `THIEF` / `WARRIOR` / `VAMPIRE` / `DRUID`
  checked against `ch.Class` for PCs only. NPCs skip the class check.
- Immortals bypass all restrictions.

Alignment helpers added to `types/character.go` — `IsEvil` (< -350),
`IsGood` (> 350), `IsNeutral` — matching the `mudprog/ifcheck.go` thresholds.

## G3: Exit flag filtering ✓

- `act/info.go` `showExits` hides `EX_SECRET` and `EX_HIDDEN` exits unless
  the viewer has `PLR_HOLYLIGHT`.
- `act/info.go` `MoveChar` — NPC + `EX_NOMOB` exit silently refuses the step.

## G4: Player-flag comm filters ✓

- `DoTell` and `DoReply` delegate to a new `deliverTell` helper:
  - `PLR_NO_TELL` on receiver → "not receiving tells" to sender; no
    receiver-side leak; reply pointer not updated.
  - `PLR_AFK` on receiver → both sides are tagged with an "AFK"/"(afk)"
    marker (message is still delivered).
  - Sender-side `ROOM_SILENCE` gates the outbound tell.
- `DoEmote` refuses when the sender has `PLR_NO_EMOTE` (C parity at
  `src/act_comm.c:2477`), and refuses in silenced rooms.

## G5: Languages subsystem ✓

**New** `util/translate.go` — pure translator with no world deps:

- `LangName` map from `LANG_*` bit → lowercase name; `LangBit(name)` is the
  inverse, with case-insensitive and prefix matching.
- `Translate(text, speaker, listener)` returns text unchanged when the
  listener shares the language or is an immortal, or when the speaker is on
  LANG_COMMON. Otherwise it scrambles through a deterministic 26-rune
  per-language alphabet, preserving case, whitespace, punctuation, and `&X`
  color-code byte pairs. Length is preserved per rune. Deterministic: the
  same input produces the same output on every call.

**New commands** in `act/comm.go`:

- `DoSpeak` — with no argument, reports the active tongue and all known
  languages; with a name (exact or unique prefix), switches the active
  tongue if the caller actually knows it. Immortals can switch freely.
- `DoLearn` — mortal path stubs to "find a trainer"; immortal path grants
  the language to self or to a target in the room.

**Wiring:** `DoSay`, `DoTell`, `DoYell`, `DoShout` now pass the outgoing
message through `util.Translate(…, sender, listener)` per recipient.
`DoGossip` remains unscrambled by design (global channel).

**Registration** in `cmd/smaug/main.go`: `speak`, `learn`.

**Known divergences from C:** the alphabet tables are hard-coded Go
constants rather than loaded from the per-language data files that full C
servers carry. The output pattern is still language-distinct and
deterministic; full alphabet-file loading is a Phase-6 polish item.

## G6: Repair shops ✓

**New** `act/repair.go`:

- `findFixer(ch)` scans `ch.InRoom.People` for an NPC whose
  `IndexData.RShop != nil`. Mirrors C `find_fixer` at `src/shops.c:202`.
- `computeRepairCost(keeper, obj)` returns gold price, or sentinel `-1`
  (unrepairable) / `-2` (already pristine). The formula mirrors C:
  ```
  base   = obj.GoldCost (or IndexData.GoldCost fallback)
  cost   = base * ProfitFix / 1000
  cost  *= condition delta (armor Value[1]-Value[0], weapon
           INIT_WEAPON_CONDITION-Value[0], wand/staff Value[1]-Value[2])
  ```
  Types outside `ITEM_ARMOR`, `ITEM_WEAPON`, `ITEM_WAND`, `ITEM_STAFF` are
  unrepairable.
- `restoreCondition(obj)` resets the condition fields per item type —
  mirrors the C `repair_one_obj` switch.
- `DoAppraise` — estimate without charging.
- `DoRepair` — full flow: keeper-in-room, item-in-inventory,
  `CanDropObj` check, cost calculation, gold deduction, gold transfer to
  keeper, condition restore.

**Registration**: `repair`, `appraise` commands wired in `cmd/smaug/main.go`.

## G7: Class/race ban enforcement ✓

**Type change**: `types/misc.go` `BanData` gets a `Type int` field (`BAN_SITE`
/ `BAN_CLASS` / `BAN_RACE`, already declared as constants). Legacy zero-
valued records are treated as BAN_SITE via the `BanType()` helper for
backwards compatibility.

**Command extensions** in `act/ban.go`:

- `ban class <name> [level]` and `ban race <name> [level]` new subcommands.
- `ban list` now shows a `[kind]` prefix per entry.
- `ban remove` now removes by name across all ban types (was site-only).
- New package-level helpers: `IsClassBanned(w, name, level)` and
  `IsRaceBanned(w, name, level)` return the matching ban when the caller's
  level is **below** the ban's threshold level. Immortals inherently
  bypass.
- `CheckBans` now filters to `BanType() == BAN_SITE` — class/race entries
  won't leak into IP enforcement.

**Enforcement** in `game/loop.go`:

- `nannyGetNewClass` consults `IsClassBanned` before accepting a class
  pick at character creation; refuses and re-shows the class menu.
- `nannyGetNewRace` does the same for race selection.

**Persistence** in `persist/ban.go`:

- Write format now emits four flag columns: `prefix suffix type level`.
- Read parses 4-column new format and falls back cleanly when only the
  legacy 2-column form is present — existing ban files load unchanged.

## G8: Note recipient filtering ✓

- New helper `isNoteTo(ch, note)` in `act/clan.go`:
  - `note.ToList == "all"` or empty → visible to everyone.
  - Otherwise case-insensitive tokenized match against `ch.Name`.
  - Immortals always see every note.
  Mirrors C `is_note_to` at `src/boards.c:61`.
- `DoNote` `list` now filters by `isNoteTo` and says "There are no notes
  for you." when the current reader has nothing addressed.
- `DoNote` `read <n>` refuses to display a note not addressed to the
  reader.
- New `CountNotesFor(ch)` helper used by the login hook in `game/loop.go`
  `enterGame` to emit "You have N note(s) waiting for you." on login when
  the count is non-zero.

## G9: Clan storeroom commands ✓

New helper `clanStoreroom(ch)` in `act/clan.go` resolves `ch.PCData.Clan
.Storeroom` to an actual `RoomIndexData`, with player-friendly refusal
messages when the caller isn't clanned or the storeroom isn't set.

- `DoClanDeposit <obj>` — moves a carried object into the clan
  storeroom. Can be used from anywhere (simulating a magic locker).
- `DoClanWithdraw <obj>` — requires the caller to physically be in the
  storeroom to prevent long-distance looting, then moves the object from
  the room into ch's inventory.

**Registration**: `deposit`, `withdraw` command names registered. These
don't collide with the bank teller's built-in `deposit`/`withdraw` — the
bank uses the `DoBank` dispatcher with those as sub-arguments only.

**Stretch deferred** (noted in the plan): automatic persistence of
storeroom contents on shutdown. The plan flagged this as a Phase-6 polish
item.

---

## Test-suite state

| Package               | Result |
|-----------------------|--------|
| cmd/smaug             | PASS (22.6s integration) |
| internal/act          | PASS |
| internal/combat       | PASS |
| internal/command      | PASS |
| internal/game         | PASS |
| internal/handler      | PASS |
| internal/magic        | PASS |
| internal/mudprog      | PASS |
| internal/net          | PASS |
| internal/persist      | PASS |
| internal/types        | PASS |
| internal/util         | PASS |
| internal/world        | PASS |

Totals: 83 source files, 66 test files, 1,505 top-level test cases
(+60 vs Tier 1). New files: `util/translate.go` + test,
`act/repair.go` + test. New test coverage added to
`act/comm_test.go`, `act/obj_test.go`, `act/info_test.go`,
`act/magic_test.go`, `act/skills_test.go`, `act/ban_test.go`,
`act/clan_test.go`, and `persist/ban_test.go`.

---

## Deferred to later tiers

**Phase-6 polish / documentation follow-ups:**

- Per-language alphabet data files under `db/system/en/` with runtime
  load (replace the hard-coded `langAlphabet` tables in
  `util/translate.go`).
- Clan-storeroom persistence on shutdown; currently the contents are
  volatile.
- `PLR_COMPACT` blank-line suppression before prompts (skipped in this
  tier to keep descriptor-flush changes out of scope).
- `ACT_HUNTER` auto-hunt-start trigger — batched with Tier 3 mudprog
  since it's a behavior change, not a pure flag gate.
- `EX_PASSAGE` auto-close and `EX_WINDOW` look-through; both tied to
  Tier-3 mudprog opcodes (`mp_open_passage` etc.).

**Tier 3 scope** (mudprog rework):

- `ACT_STOP_SCRIPT` honoring during `mprog_driver` nested runs.
- `ACT_SECRETIVE` NPC flag in `util.Act()` visibility.

**Tier 4 scope** (content breadth):

- `ROOM_NO_RECALL` / `ROOM_NO_SUMMON` mirror-side checks inside
  `SpellWordOfRecall`, once it exists.
- `ROOM_SAFE` + PKILL guards in `spellAreaAttack`.
- `ITEM_POISONED` triggers on `DoEat`/`DoDrink`/weapon hit.
- Storeroom persistence and Phase-6 polish items above.
