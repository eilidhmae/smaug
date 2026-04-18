# Plan: `DoChannels` Toggle Command + `Deaf` Persistence Fix

**Status:** Planned (2026-04-17). Adversary-verified research: FAIL verdict (the "fail" is agreement — confirms a live persistence bug and a missing command, not a dispute with the research).
**Priority:** P1 — a persistence bug that silently drops player preferences on every logout, plus a missing command that locks players out of deafened channels with no recovery.
**Scope:** `internal/persist/player.go` (persistence fix), `internal/act/channels.go` (new `DoChannels`), `internal/boot/boot.go`, tests. Two landing order: persistence fix ships first, then the command.

---

## Problem

### Bug 1 — `Deaf` bitvector is NEVER written to the player save file

`internal/persist/player.go` `SavePlayer` (~lines 454-589) does not emit a `Deaf` line. The load path at line 189-191 reads it:

```go
case "Deaf":
    bv, _ := scanner.ReadBitVector()
    ch.Deaf = bv
```

So the field is expected in the save format — the save path is simply missing. Any player who runs `channels -immtalk` (or any channel toggle) silently loses their setting on next logout. Confirmed by adversary: "no `fmt.Fprintf` anywhere in `SavePlayer` writes `Deaf`."

### Bug 2 — No `DoChannels` command at all

Players can be **put into** a deaf state (e.g., if `DoImmtalk` or any other channel command had a deaf-set path — most don't currently), but there is **no way to re-enable** a deaf channel. C `do_channels` is the canonical toggle UI. Go has zero implementation.

### Why this matters now

Tier 9 (2026-04-17) landed `DoImmtalk` / `DoGtell` / `BroadcastAuction` / `DoAuction` with the C-faithful behavior that `DoImmtalk` does NOT auto-clear the deaf bit on a deaf sender. C relies on `do_channels` to provide the recovery path; Go inherits the behavior without the recovery. This is a live UX hole.

## C Reference

- `src/act_info.c:5269-5557` — `do_channels`. No-arg path prints grouped status; `+<name>`/`-<name>` toggles one; `+all`/`-all` toggles the "public" set.
- **Channel name matching uses `str_cmp` (exact), NOT prefix match.**
- **Channel keyword is `chat`, not `gossip`.** The Go command is `DoGossip`; the channel constant is `CHANNEL_CHAT`; users type `channels -chat` to deafen.
- **NPC gate:** line 5274. **`PLR_SILENCE` no-arg early-return:** lines 5279-5283.
- **`+all` / `-all` public-channel set** (C `act_info.c:5484-5541`): RACETALK, AUCTION, CHAT, QUEST, WARTALK, PRAY, TRAFFIC, MUSIC, ASK, YELL, and AVTALK **gated on `LEVEL_IMMORTAL`** (NOT `IS_HERO` — adversary caught this; C uses `ch->level >= LEVEL_IMMORTAL` at lines 5504 and 5535). NOT included: TELLS, WHISPER, CLAN, COUNCIL, GUILD, MONITOR, DEATH, AUTH, RETIRED, LOG, BUILD, COMM, WARN, BUG, ORDER, HIGHGOD, HIGH. **PRAY IS included** — adversary caught the plan v1 contradiction where PRAY was listed here but then excluded below.
- **Error strings** (exact wording):
  - Invalid format: `"Channels -channel or +channel?\n\r"` (line 5410).
  - Unknown channel: `"Set or clear which channel?\n\r"` (line 5478).
  - Success: `"Ok.\n\r"` (line 5553).
  - Silenced: `"You are silenced.\n\r"` (line 5283).

## Go Current State

- `CharData.Deaf` at `internal/types/character.go:164` — type `BitVector` (4×uint32, matches C `EXT_BV`).
- 32 `CHANNEL_*` constants at `internal/types/enums.go:910-943` (iota-derived, matches C order).
- `Deaf` bit readers: `internal/act/channels.go:37,57` (DoImmtalk), `internal/act/auction.go:37` (BroadcastAuction). TELL/YELL/SHOUT/GOSSIP do NOT check Deaf (matches C — those bypass channel filtering).
- No `Deaf.String()` verified pattern exists yet — but `Act` and `AffectedBy` use the same `BitVector.String()` / `ParseBitVector` round-trip at `player.go:482` and `:485`. Reuse.

## Channel Scope for `DoChannels`

Per the adversary's channel universe table, many channels listed in C's `do_channels` don't have Go command implementations yet (music, newbie, racetalk, wartalk, quest, ask, traffic, whisper, pray, plus all immortal-only channels). For the Go `DoChannels`:

- **Accept toggles for any channel whose constant exists in Go** (all 32). Toggling deaf on `music` is harmless — it just sets a bit no command reads.
- **Display only channels whose commands are implemented**, PLUS the standard C-grouped list, to avoid silently ignoring what users type. **Decision:** Print the full C-grouped list. Users who see "music" in the display but find it doesn't work can infer it's not yet implemented (the command itself would say "Huh?"). This matches the existing Go convention of defining all the constants but only implementing commands incrementally.
- **`pray`:** C's no-arg display has the `"+PRAY"` line commented out (act_info.c:5358), but the `+all`/`-all` path DOES toggle PRAY. **Resolution (adversary-flagged contradiction):** Match C exactly — omit PRAY from the no-arg display output, but DO include it in the `+all`/`-all` set. Also accept `channels +pray` / `channels -pray` as individual toggles (C's toggle-table resolver at `act_info.c:5440-5441` includes `pray`).

## Task Groups

### G1 — Fix `Deaf` persistence in `SavePlayer` (P0; must ship first)

- File: `internal/persist/player.go`, `SavePlayer` function.
- Add adjacent to existing `Act` / `AffectedBy` / `Resistant` writes (e.g., after `AffectedBy`):
  ```go
  if !ch.Deaf.IsEmpty() {
      fmt.Fprintf(w, "Deaf       %s\n", ch.Deaf.String())
  }
  ```
- **Verify `BitVector.String()` exists and round-trips with `ParseBitVector`** (already used by Act/AffectedBy — if they work, Deaf works).
- **Test first:**
  - `TestSaveLoadPlayer_DeafRoundTrip` — set Deaf bits for multiple channels, save, reload, assert Deaf matches byte-for-byte.
  - `TestSaveLoadPlayer_EmptyDeafNotEmitted` — no Deaf bits set, save file has no `Deaf` line, load defaults to empty bitvector.
  - Mutation-verify: remove the write, confirm round-trip test fails; revert, confirm passes.
- **This task is independent of DoChannels and can ship alone. It fixes a live data-loss bug.**

### G2 — New file `internal/act/channels.go` additions: `DoChannels`

The existing `channels.go` has `DoImmtalk` and `DoGtell`. Extend it (don't create a new file).

Helper: a private channel name→bit table. **Full enumeration required** (adversary caught plan v1 only showing two placeholder entries). C's `do_channels` at `act_info.c:5414-5473` has 30 named branches; Go table must match exactly, in the same order:

```go
var channelToggleTable = []struct {
    name string
    bit  int
}{
    {"auction",  types.CHANNEL_AUCTION},
    {"traffic",  types.CHANNEL_TRAFFIC},
    {"chat",     types.CHANNEL_CHAT},     // NOT "gossip" — match C
    {"clan",     types.CHANNEL_CLAN},
    {"council",  types.CHANNEL_COUNCIL},
    {"guild",    types.CHANNEL_GUILD},
    {"quest",    types.CHANNEL_QUEST},
    {"tells",    types.CHANNEL_TELLS},
    {"immtalk",  types.CHANNEL_IMMTALK},
    {"log",      types.CHANNEL_LOG},
    {"build",    types.CHANNEL_BUILD},
    {"high",     types.CHANNEL_HIGH},
    {"pray",     types.CHANNEL_PRAY},
    {"avatar",   types.CHANNEL_AVTALK},
    {"monitor",  types.CHANNEL_MONITOR},
    {"death",    types.CHANNEL_DEATH},
    {"auth",     types.CHANNEL_AUTH},
    {"newbie",   types.CHANNEL_NEWBIE},
    {"music",    types.CHANNEL_MUSIC},
    {"muse",     types.CHANNEL_HIGHGOD}, // NB: "muse" → HIGHGOD, NOT MUSIC
    {"ask",      types.CHANNEL_ASK},
    {"yell",     types.CHANNEL_YELL},
    {"comm",     types.CHANNEL_COMM},
    {"warn",     types.CHANNEL_WARN},
    {"bug",      types.CHANNEL_BUG},
    {"order",    types.CHANNEL_ORDER},
    {"wartalk",  types.CHANNEL_WARTALK},
    {"whisper",  types.CHANNEL_WHISPER},
    {"racetalk", types.CHANNEL_RACETALK},
    {"retired",  types.CHANNEL_RETIRED},
}
```

`"muse" → CHANNEL_HIGHGOD` is the non-obvious mapping; a test asserting `-muse` sets `CHANNEL_HIGHGOD` (not `CHANNEL_MUSIC`) will guard regressions.

#### G2a — No-arg display

- Print `"You are silenced.\n\r"` (NB: `\n\r`, matching C and existing Go channel code) if `PLR_SILENCE` set, return.
- NPC guard: return silently.
- Print the grouped display mirroring C's four sections (Public, Private, Immortal, Restricted-trust). For each entry, `+CHANNEL` if deaf bit CLEARED (channel ENABLED), `-channel` if deaf bit SET.
- **Scope cut:** Port the exact C grouping; use color codes `&g` / `&G` as in C. Do not attempt to reorder or alphabetize.
- **Immortal section gate:** the entire "Immortal Channels" section (header + entries) is wrapped in `IS_IMMORTAL(ch)` — mortals don't see the section header at all. Use `ch.IsImmortal()`.
- Trust/membership gates (adversary corrections marked):
  - Auction entry visibility: `ch.GetTrust() > 4` (matches C).
  - AVTALK no-arg display: `ch.Level >= types.LEVEL_HERO` (`IS_HERO`; C `act_info.c:5305`). **NOTE:** `IsHero()` method does NOT exist on `CharData` — use the inline `Level` comparison. (Adversary caught plan v1 assuming the helper existed.)
  - Clan/Council/Guild entries: guard on `ch.Clan != nil`, `ch.Council != nil`, membership-equivalent helpers (check `internal/types/character.go` and clan-related code for existing predicates).
  - Immortal-trust gates (LOG, BUILD, HIGH, MUSE/HIGHGOD, BUG): use existing `ch.GetTrust()` thresholds matching C.
  - `pray`: omit from display (C has the line commented out at `act_info.c:5358`).

#### G2b — `+<name>` / `-<name>` toggle

- Parse first argument; reject if first char is not `+` or `-`: print `"Channels -channel or +channel?\n\r"` (NB: `\n\r`; adversary caught `\r\n` in plan v1).
- Strip the sign, lookup in `channelToggleTable` via **exact** `strings.EqualFold` (case-insensitive; matches C's `str_cmp`).
- Handle the `all` case first: if `name == "all"`, toggle the C-listed public set and return.
- On match: `+` → `ch.Deaf.RemoveBit(bit)`; `-` → `ch.Deaf.SetBit(bit)`. Print `"Ok.\n\r"`.
- On no match: print `"Set or clear which channel?\n\r"`.

#### G2c — `+all` / `-all` meta-toggle

- Enumerate the public-channel set exactly as C does (`act_info.c:5484-5510` for `+all`; `5515-5541` for `-all`).
- Set: RACETALK, AUCTION, CHAT, QUEST, WARTALK, PRAY, TRAFFIC, MUSIC, ASK, YELL. Plus AVTALK only if `ch.Level >= types.LEVEL_IMMORTAL` (NOT `IS_HERO`; adversary correction — C `5504`/`5535` uses `IS_IMMORTAL`).
- Single `"Ok.\n\r"` after processing.

### G3 — Register in `boot.go`

- Add:
  ```go
  reg.Register(&command.Command{Name: "channels", DoFun: act.DoChannels, Position: types.POS_DEAD, Level: 0})
  ```
- No alias (C has none).

### G4 — Tests

File: `internal/act/channels_test.go` (extends the existing Tier 9 tests).

- `TestDoChannels_NPCIsNoop` — NPC, no output, no panic.
- `TestDoChannels_SilencedPrintsMessage` — PLR_SILENCE, output contains "You are silenced.".
- `TestDoChannels_NoArgDisplaysGroups` — verify output contains "Public channels", "Private channels", "Immortal Channels" section headers.
- `TestDoChannels_NoArgShowsEnabledWithPlus` — clean Deaf, output has `+CHAT` (uppercased).
- `TestDoChannels_NoArgShowsDisabledWithMinus` — set CHAT deaf bit, output has `-chat` (lowercased).
- `TestDoChannels_ToggleOffSetsBit` — `-chat`, assert bit set, output "Ok".
- `TestDoChannels_ToggleOnClearsBit` — start with CHAT deaf, `+chat`, assert bit cleared.
- `TestDoChannels_InvalidFormat` — `chat` (no +/-), output "Channels -channel or +channel?".
- `TestDoChannels_UnknownChannel` — `-foobar`, output "Set or clear which channel?".
- `TestDoChannels_GossipKeywordRejected` — `-gossip` (C uses "chat"), output unknown-channel error. Documents the naming fidelity. UX note: players using `DoGossip` routinely will be confused that `channels -gossip` doesn't work — document in a help entry or landing commit note that the channel name for toggling is `chat` (matches C).
- `TestDoChannels_MuseTogglesHighgodNotMusic` — `-muse` sets `CHANNEL_HIGHGOD`, not `CHANNEL_MUSIC`. Guard against the non-obvious mapping.
- `TestDoChannels_PrayNotInNoArgDisplay` — PC clean, output does NOT include "pray" entry (C comments it out at 5358).
- `TestDoChannels_PrayIndividualToggleWorks` — `-pray` DOES set `CHANNEL_PRAY` (C toggle table includes it).
- `TestDoChannels_AllAvtalkGatedOnLevelImmortal` — mortal runs `+all`, CHANNEL_AVTALK NOT affected. Immortal runs `+all`, CHANNEL_AVTALK cleared. Adversary-critical mutation guard.
- `TestDoChannels_AllTogglePublicSet` — `-all`, verify each public channel bit set, verify TELLS bit NOT set (not in C's all-set).
- `TestDoChannels_RoundTripPersistsViaSave` — toggle chat off, call SavePlayer, LoadPlayer, `DoChannels` no-arg shows `-chat`. **This is the end-to-end proof that G1 + G2 work together.**
- Mutation-verify: flip each bit op, confirm relevant test fails; revert.

### G5 — Update docs

- `TODO.md`: move "DoChannels toggle command" from Active follow-ups to Done on landing.
- `plan-channels.md`: reference this plan in the follow-up section; note that DoChannels + persistence landed as a bundle.
- Append completion record to this plan.

## Acceptance Criteria

A1. `Deaf` bit is written by `SavePlayer` in the exact format parsed by `LoadPlayer`, conditional on non-empty.
A2. Save/load round-trip is byte-exact for any combination of Deaf bits.
A3. `DoChannels` NPC gate silently returns.
A4. `DoChannels` no-arg path prints the C-grouped list with correct `+`/`-` prefixes.
A5. `DoChannels +<name>` / `-<name>` exact-match toggles; invalid format and unknown-channel paths emit C-exact error strings.
A6. `+all` / `-all` toggles exactly the C public set.
A7. End-to-end test (G4 last item) proves `-chat` survives save/load.
A8. `go test -count=3 ./...` green across all packages.
A9. No regression in Tier 9 channel commands (DoImmtalk, DoGtell, BroadcastAuction, DoAuction).

## Scope Cuts

- **No new channel COMMAND implementations** (music, newbie, racetalk, etc.). `DoChannels` lets users toggle the bits; the commands themselves remain Phase-6.
- **No `pray` toggle** (C leaves it commented-out in the display but handles it in the toggle table — match C's display exclusion).
- **No prefix matching** on channel names within the toggle argument (e.g., `-imm` should NOT match `immtalk`). C uses exact match; match it.
- **No `ch` shortcut alias** for `channels`.

## Open Questions

1. **Trust-gated display of immortal channels.** Mortals shouldn't see "Immortal Channels" section. Easy: gate the section header + entries on `ch.IsImmortal()`. Match C lines 5356-5373.
2. **Channel order on the display.** C emits in a fixed hand-written order. Mirror exactly — do not sort. Adversary flagged this.
3. **Should `BitVector.IsEmpty()` exist?** If not, inline a check `ch.Deaf[0]|ch.Deaf[1]|ch.Deaf[2]|ch.Deaf[3] != 0`. Small enough to decide during implementation.
4. **`CHANNEL_SHOUT` anomaly.** Go has `DoShout` but C's `do_channels` does not list `shout`. **Decision:** Omit "shout" from the Go `DoChannels` table (matches C). Tradeoff: Shout cannot be deafened via `channels`. Acceptable.

## Risk

- **Medium.** G1 (persistence fix) is small but critical — regressions in the player-file format break saves for everyone. G4's round-trip test is the guardrail.
- **G2 is ~200 LOC with many branches** (C's `do_channels` is a big switch). Risk is UX divergence. Mitigation: ~15 tests covering each branch; mutation-verify each bit operation.
- Adversary explicitly called out the persistence bug as the highest-impact finding. Shipping G1 first (alone, if necessary) is the right phasing.

## Adversary-Resolved Concerns (2026-04-17 plan review)

1. **`IsHero()` method does not exist.** Plan v1 assumed it. Corrected to inline `ch.Level >= types.LEVEL_HERO`.
2. **`+all`/`-all` AVTALK gate uses wrong constant.** Plan v1 said `IS_HERO`; C uses `LEVEL_IMMORTAL`. Corrected.
3. **PRAY contradiction** (in `+all` set but excluded from "implement toggling"). Resolved: omitted from display (matches C commented-out line), included in `+all`/`-all` and individual toggles (matches C toggle table).
4. **Line endings** — all error strings and echoes switched from `\r\n` to `\n\r` to match existing Go channel code + C source.
5. **`channelToggleTable` incomplete** — full 30-entry enumeration now in the plan body; `muse → CHANNEL_HIGHGOD` mapping called out.
6. **G1 format spacing** — the scanner is keyword-based (verified: `player.go:189-191` reads `case "Deaf"` then the value); column alignment is cosmetic. Use any reasonable padding.
7. **Ship-G1-alone is forward-compatible** — `LoadPlayer` uses `case "Deaf":` which simply doesn't fire when the line is absent. Old save files without `Deaf` load unchanged.

## Suggested Phasing

1. Ship G1 (persistence fix) as a standalone commit. Adversary-verify.
2. Ship G2 + G3 + G4 + G5 as a second commit.

This lets G1 land immediately (unblocks any player who has deafened any channel today), with G2 following as the UX completion.

---

## Completion record (2026-04-17)

Landed in two commits per the suggested phasing. Adversary reviewed each commit independently; both PASS.

### Commit `ab47893` — G1 `Deaf` persistence fix (standalone)

- `internal/persist/player.go`: added `if !ch.Deaf.IsEmpty() { fmt.Fprintf(w, "Deaf       %s\n", ch.Deaf.String()) }` inside `SavePlayer`, adjacent to the existing `Act` / `AffectedBy` writers. The load path at `player.go:189-191` was already reading `case "Deaf":` via `ParseBitVector` — the writer was the sole missing half of the round-trip.
- `internal/persist/player_test.go`: added `TestSaveLoadPlayer_DeafRoundTrip` (two distinct channel bits through real `bytes.Buffer`) and `TestSaveLoadPlayer_EmptyDeafNotEmitted` (forward-compat — no `"Deaf"` substring in output when bitvector is empty; old saves without the line load cleanly).
- Mutation-verified: wrapping the conditional in `if false` fails `TestSaveLoadPlayer_DeafRoundTrip`; revert → green.
- Note: a prior worker session lost the tests during a git operation; a follow-up worker re-added them. A preceding `2a200f2` cleanup commit ran `gofmt -w` on the pre-existing file to unblock the pre-commit hook.

### Commit (next) — G2/G3/G4/G5

- `internal/act/channels.go`:
  - 30-entry `channelToggleTable` in C order (matches `act_info.c:5414-5473`). Non-obvious mappings pinned by tests: `muse → CHANNEL_HIGHGOD` (NOT `CHANNEL_MUSIC`); `chat` (NOT `gossip`).
  - `DoChannels(ch, argument)` with: NPC gate first (silent no-op); empty arg → `PLR_SILENCE` check then grouped display; `+<name>`/`-<name>` → exact match via `strings.EqualFold` through the table; `+all`/`-all` → public-channel set including AVTALK gated on `Level >= LEVEL_IMMORTAL` (adversary-corrected from plan v1's `IS_HERO`). Error strings are C-exact with `\n\r` line endings.
  - No-arg display mirrors C's four sections (Public / Private / Immortal / trust-gated) with `&g`/`&G` colors. `+NAME` uppercase = enabled; `-name` lowercase = deafened. Per-section gates: auction > trust 4; clan/council/guild on `PCData` predicates (guild uses `ClanType == CLAN_GUILD`); avatar on `Level >= LEVEL_HERO`; immortal section entirely gated on `IsImmortal()`.
  - `pray` omitted from display (matches C comment-out at `5358`) but present in the toggle table, so `channels -pray` works individually.
  - `shout` omitted from the toggle table (C doesn't include it) — tradeoff documented in the plan scope cuts.
- `internal/act/channels_test.go`: 16 new tests covering every branch + an end-to-end `TestDoChannels_RoundTripPersistsViaSave` that composes G1 and G2 through real `persist.SavePlayer` → `bytes.Buffer` → `persist.LoadPlayer`. Five worker-applied mutations all caught by specific named tests; adversary independently applied a sixth mutation (inverting `fClear` at `channels.go:275`) with expected failure.
- `internal/boot/boot.go`: registered `channels` once, no alias, adjacent to the Tier 9 channel registrations.
- `go vet ./...` clean. `gofmt -l` silent on the three touched files. `go test -count=3 ./...` green across all 15 packages. No regression in Tier 9 (`DoImmtalk` / `DoGtell` / `BroadcastAuction` / `DoAuction`) or any other package.

### Deliberate simplifications (flagged for follow-up in `TODO.md`)

- **Immortal section trust gates collapsed to `IsImmortal()`.** C uses per-entry trust thresholds (`sysdata.muse_level` / `sysdata.log_level` / `sysdata.think_level` + hardcoded 57 for bug). Refinement requires a `sysdata` port first.
- **`publicAll` slice duplicates part of `channelToggleTable`** — `channels.go:247`. Low drift risk today (public set hasn't changed in decades) but worth deriving one from the other when the next channel gets added.
- **Display AVTALK gate uses `Level >= LEVEL_HERO` (not trust-based `IS_HERO`).** Consistent with port convention; adversary noted as undocumented but correct.

### Adversary findings resolved during implementation

- Plan v1's `IsHero()` assumption → replaced by inline `Level >= LEVEL_HERO`.
- Plan v1's `+all`/`-all` AVTALK gate using `IS_HERO` → corrected to `LEVEL_IMMORTAL`, mutation-guarded with a new LEVEL_HERO-level mortal subcase in `TestDoChannels_AllAvtalkGatedOnLevelImmortal`.
- Line endings `\r\n` → `\n\r` across all error strings.
- `channelToggleTable` documented as 30 entries with the `muse → CHANNEL_HIGHGOD` non-obvious row pinned by a dedicated test.

### Acceptance criteria status

- A1 ✅ Deaf written by SavePlayer in parseable format.
- A2 ✅ Round-trip byte-exact for any combination of bits.
- A3 ✅ NPC gate silent.
- A4 ✅ No-arg grouped display with `+`/`-` prefixes.
- A5 ✅ Exact-match toggles; invalid-format + unknown-channel errors byte-exact.
- A6 ✅ `+all`/`-all` toggles the exact C public set.
- A7 ✅ `TestDoChannels_RoundTripPersistsViaSave` proves `-chat` survives save/load.
- A8 ✅ `go test -count=3 ./...` green.
- A9 ✅ No Tier 9 regression.
