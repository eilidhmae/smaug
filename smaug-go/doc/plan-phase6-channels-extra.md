# Plan: Phase 6 — Extra Communication Channels

**Status:** External adversary audit 2026-04-18 (lineage `audit-channels-extra`) — verdict **CONCERNS**. Factual claims largely correct; three verified factual errors (misattributed `TranslateFor` package + capitalization; `CHANNEL_MUSIC` cited line; missing `CouncilData.Name` location precision); one data-driven concern requiring human decision (newbie-council absent from stock `db/councils/council.lst`). Corrections applied in-plan. Re-audit not required before execution; open questions tracked in §Open questions.
**Priority:** Wave 1 Phase 6. Small self-contained port; template established in Phase 5 Tier 9 (`plan-channels.md`); all six target channels route through a single C function (`talk_channel`), so the port has genuine shared structure.
**Scope:** Add six commands — `music`, `newbiechat`, `racetalk`, `wartalk`, `counciltalk`, `guildtalk` — backed by a narrow shared `talkChannel` helper factored out for this tier. No changes to the four channels already shipped in Tier 9 (`DoImmtalk` / `DoGtell` / `DoClantalk` / `DoAuction` stub) — they keep their open-coded form; retrofit is explicitly out of scope and tracked as a follow-up.

---

## Problem

`src/act_comm.c:392-858` contains a single `talk_channel(ch, argument, channel, verb)` function that all "generic" chat channels dispatch to. Ten C `do_*` entry points are thin wrappers that validate a per-command predicate (authed, not-NPC, has-council, has-clan-of-right-type, pkill) and then call `talk_channel` with a `(CHANNEL_*, verb)` pair.

Tier 9 shipped `DoImmtalk` / `DoGtell` / `DoClantalk` / `DoAuction` (stub) open-coded because the audience predicates diverged sharply (immortal-only / group-walk / clan-match / auction-helper). That completion record flagged:

> `music` / `newbiechat` / `racetalk` / `wartalk` / `counciltalk` / `guildtalk` channels. **Revisit factor-out-of-`talk_channel` question once 5+ channels land.**
> — `plan-channels.md` §Deferrals, confirmed again in `TODO.md:62`

This plan lands all six. The "5+" threshold is crossed; the six new channels all share genuine `talk_channel`-derived structure (unlike the Tier 9 quartet); a narrow helper pays off.

### What is NOT in scope (C also doesn't have it, or is a deferred subsystem)

- **`PLR_WIZINVIS` "(level) " preamble** — C `act_comm.c:522-557` prepends the wizinvis level to the channel line when the sender is invis. No Go channel command honors this today (`DoImmtalk`, `DoGossip`, `DoShout`, `DoYell`, `DoClantalk` all skip it). Port behavior: match existing Go convention — skip the preamble for all six new channels. Follow-up entry added to `TODO.md` to cover all channels uniformly.
- **`ROOM_LOGSPEECH` file append** — C `act_comm.c:640-645` writes channel output to a log file when the sender's room has `ROOM_LOGSPEECH`. Go has the `ROOM_LOGSPEECH` constant (`enums.go:712`) but no channel command in the port honors it. Match existing behavior; queue uniform fix.
- **`is_ignoring` receiver filter** — C `act_comm.c:686` skips receivers who have the sender on their ignore list. Go has `PCData.Ignored []string` (`pcdata.go:104`) but no helper and no existing command consults it. Match existing behavior; queue uniform fix.
- **Language translation / `translateFor`** — already reused by `DoYell`, `DoShout`, `DoGossip`, `DoTell`, `DoReply`, `DoClantalk`, `DoSay`. The helper is **`translateFor` (unexported) in `internal/act/comm.go:13`** — NOT `util.TranslateFor` (a prior draft of this plan misattributed it). Since `talkChannel` lives in the same `act` package, the call site is `translateFor(ch, vch, argument)` — no import change required.
- **`NOT_AUTHED` gates** — Go has no auth/unauthed split; every logged-in char is authed. Drop the `NOT_AUTHED` predicate from all six wrappers (the Tier 9 plan established this). Document as a deliberate divergence.
- **`AFLAG_SILENCE`** — area-level silence flag. Defined as a constant (`constants.go:687`) but no Go command checks it. Match existing behavior (`DoYell` / `DoShout` / `DoGossip` only check `ROOM_SILENCE`); queue uniform fix.
- **`WAIT_STATE`** — the 6-pulse wait at C `:852-855` for mortals on non-{yell,wartalk,clan} channels. Go's port does not apply wait-states to any channel today; match existing behavior.
- **Full 30-branch `talk_channel` god-function port** — `talk_channel` in C covers immtalk, avtalk, wartalk, racetalk, music, traffic, quest, ask, yell, shout, highgod (muse), high (think), chat, clan, order, council, guild, newbie. Porting the whole switch would duplicate logic that Tier 9's open-coded handlers already own. The helper factored here covers **only the six new channels** that genuinely share the "clan-or-public-filter + deaf-filter + room-silence + self-echo + broadcast" skeleton. Existing Go handlers keep their open-coded form.

### Factor-vs-open-code decision

**Decision: ship a narrow shared helper (`talkChannel`) first, then port the six atop it.**

Rationale:

1. All six new channels route through C's `talk_channel` — their shared structure is real, not imagined.
2. Open-coded, six new `DoXxx` handlers would duplicate ~40 LOC each (~240 LOC total). A helper collapses that to ~80 LOC helper + ~15 LOC per wrapper = ~170 LOC total.
3. The per-channel audience predicates (wartalk→pkill, counciltalk→shared council, racetalk→shared race, guildtalk→shared guild-type clan, music→public, newbiechat→immortal-or-newbie-council-member) are small enough to pass in as a `recipientFilter func(vch) bool` closure rather than baking into the helper.
4. Non-retrofit: the existing four Tier-9 handlers keep their open-coded form. This limits blast radius. The helper is defined in terms of what the six new channels need — if the Tier 9 four are later retrofitted, the helper signature may need widening (e.g., an immortal-receiver gate); that's a separate plan.
5. The helper signature intentionally does NOT attempt to cover `DoImmtalk`'s deaf-sender-block semantics (which requires no-auto-clear), `DoGtell`'s group walk (no `talk_channel` at all in C), `DoClantalk`'s clan-name prefix formatting, or `DoAuction`'s closed-stub. Those remain open-coded.

Alternative considered: open-code all six. Rejected because (a) ~240 LOC of near-identical code is exactly the copy-paste the Tier 9 completion flag was warning against, and (b) the six channels' divergence is in a single boolean predicate (the `recipientFilter`) and a single text verb — clean parameters for a helper, not god-function arguments.

---

## C Reference (authoritative)

All citations against `src/act_comm.c` at HEAD.

### `talk_channel` shared function

**Lines 392-858**. The 30-branch god-function all public/clan channels route through. Key sections the helper ports:

- **NPC-in-organization block** — `:409-430`. NPC attempting CLAN, ORDER, COUNCIL, GUILD gets `"Mobs can't be in <X>.\n"` and returns. Applies to counciltalk, guildtalk. The wrapper also rejects NPC via the `ch.IsNPC() || ch.PCData == nil` gate so this is belt-and-suspenders.
- **Pkill gate for wartalk** — `:432-436`. `!IS_PKILL(ch) && channel == CHANNEL_WARTALK` → `"Peacefuls have no need to use wartalk.\n"`. Go: `ch.PCData.Flags & PCFLAG_DEADLY`.
- **ROOM_SILENCE / AFLAG_SILENCE** — `:471-476`. Applies to every channel via `talk_channel`. Port: check `ROOM_SILENCE` only (match existing Go channel convention; `AFLAG_SILENCE` queued as a uniform follow-up).
- **Charmed-NPC early return** — `:485-490`. Delivers `"I don't think so...\n"` to master, suppresses broadcast. The wrapper's `ch.IsNPC()` gate short-circuits before this in the Go port; include a short-circuit on `ch.Affected_by & AFF_CHARM` or equivalent is out of scope — charmed NPCs aren't expected to invoke these commands.
- **Empty-arg `"<verb> what?"`** — `:492-498`. Capitalizes the first char of `verb`. Port: literal `"%s what?\n\r"` with `strings.Title` or an upper-first helper (Go `strings.ToUpper(verb[:1])+verb[1:]` — no external helper needed).
- **`PLR_SILENCE` sender gate** — `:500-504`. `"You can't %s.\n"` — already honored by `DoYell`/`DoShout`/`DoGossip`/`DoTell` in Go (Tier 9 opportunistic fix). Mirror for the six new channels inside the helper.
- **Deaf-sender block** — `:506-513`. If `xIS_SET(ch->deaf, channel) && (channel != CHANNEL_WARTALK || channel != CHANNEL_YELL)` → `"You don't have the %s channel turned on. To turn it on, use the Channels command.\n"` and return. **The `(... || ...)` condition is always true** — the C predicate has always-true bug semantics. **Go port intent: use `channel != CHANNEL_WARTALK && channel != CHANNEL_YELL`** (i.e., `&&` not `||`) — this is what the C author meant: don't block deaf-sender on wartalk or yell specifically. The existing `DoImmtalk` (`channels.go:91-94`) doesn't apply the exception pair; preserve the exception for wartalk here.
- **Deaf-auto-clear at `:514`** — unreachable because the `:511` return fires first when the bit is set. Tier 9's `DoImmtalk` documents this precisely. Match the same no-auto-clear semantic in the helper.
- **Self-echo** — `:560-638`. Per-channel color and format. All six new channels use `"You %s '%s'\n"` format — `AT_MUSIC` for music, `AT_RACETALK` for racetalk, `AT_WARTALK` for wartalk, default `AT_GOSSIP` for newbiechat, clantalk-style for counciltalk and guildtalk. Port: use `&Y`/`&G`/`&D` inline codes (match Tier 9 fidelity; per-AT_ color is a separate uniform pass queued in TODO.md).
- **Broadcast walk** — `:671-848`. Iterate `first_descriptor`, filter `CON_PLAYING && vch != ch && !deaf`. Apply per-channel predicate. Port: iterate `WorldRef.Descriptors`.
- **Per-channel receiver predicate block** — `:691-755`. Each channel's filter:
  - `CHANNEL_IMMTALK` — `IS_IMMORTAL(och)`.
  - `CHANNEL_WARTALK` — `IS_PKILL(och)`.
  - `CHANNEL_AVTALK` — `IS_HERO(och)`.
  - `CHANNEL_NEWBIE` — `IS_IMMORTAL(och) || NOT_AUTHED(och) || (pcdata->council && !str_cmp(council->name, "Newbie Council"))`.
  - `CHANNEL_CLAN` / `CHANNEL_ORDER` / `CHANNEL_GUILD` — `!IS_NPC(vch) && vch->pcdata->clan == ch->pcdata->clan`.
  - `CHANNEL_COUNCIL` — `!IS_NPC(vch) && vch->pcdata->council == ch->pcdata->council`.
  - `CHANNEL_RACETALK` — `vch->race == ch->race`.
  - `CHANNEL_MUSIC` / `CHANNEL_TRAFFIC` / `CHANNEL_QUEST` / `CHANNEL_ASK` — no filter (public).
- **WAIT_STATE** — `:852-855`. OUT OF SCOPE for this tier (no Go channel applies wait-states).

### Per-wrapper gates (C side)

- **`do_music`** (`:1012-1021`): `NOT_AUTHED → "Huh?"`. Call `talk_channel(..., CHANNEL_MUSIC, "music")`. **Go port: drop NOT_AUTHED.** No other gates.
- **`do_newbiechat`** (`:935-947`): `IS_NPC || (!NOT_AUTHED && !IS_IMMORTAL && !(council && !str_cmp(council->name, "Newbie Council")))` → `"Huh?"`. **Go port:** `IS_NPC || (!IsImmortal && !isNewbieCouncilMember)`. Where `isNewbieCouncilMember := ch.PCData != nil && ch.PCData.Council != nil && strings.EqualFold(ch.PCData.Council.Name, "Newbie Council")`. If no such council exists in the data (likely — stock SMAUG data doesn't always ship one), only immortals can use `newbiechat` in the Go port, which matches C semantics in that configuration.
- **`do_counciltalk`** (`:975-990`): `NOT_AUTHED → "Huh?"`; `IS_NPC || !ch->pcdata->council → "Huh?"`. Call `talk_channel(..., CHANNEL_COUNCIL, "counciltalk")`. **Go port:** `if ch.IsNPC() || ch.PCData == nil || ch.PCData.Council == nil { "Huh?" }`.
- **`do_guildtalk`** (`:993-1009`): `NOT_AUTHED → "Huh?"`; `IS_NPC || !ch->pcdata->clan || clan->clan_type != CLAN_GUILD → "Huh?"`. Call `talk_channel(..., CHANNEL_GUILD, "guildtalk")`. **Go port:** `if ch.IsNPC() || ch.PCData == nil || ch.PCData.Clan == nil || ch.PCData.Clan.ClanType != types.CLAN_GUILD { "Huh?" }`.
- **`do_wartalk`** (`:4612-4621`): `NOT_AUTHED → "Huh?"`. Call `talk_channel(..., CHANNEL_WARTALK, "war")`. **Go port:** no wrapper-level gate beyond `ch.IsNPC()`. The pkill gate lives in the helper's per-channel rejection at `:432-436`.
- **`do_racetalk`** (`:4624-4632`): `NOT_AUTHED → "Huh?"`. Call `talk_channel(..., CHANNEL_RACETALK, "racetalk")`. **Go port:** `ch.IsNPC()` only (racetalk on an NPC with no race is meaningless; match the Go comm-command pattern of silencing NPC channel abuse).

### Commands file registration

`db/system/en/commands.dat` has entries for all six (verified). Boot-time registration in `internal/boot/boot.go` is the Go equivalent.

---

## Go Current State

- `CharData.Race int` at `character.go:84`. Set during chargen; persisted via `SavePlayer`. Shared-race check: `ch.Race == och.Race`.
- `CharData.PCData.Flags int` with `PCFLAG_DEADLY uint32 = 1 << 1` at `constants.go:648` — the pkill predicate. Existing mudprog ifcheck `ispkill` uses `uint32(chk.PCData.Flags) & types.PCFLAG_DEADLY != 0` (`ifcheck.go:543`). Reuse that expression in the helper.
- `CharData.PCData.Council *CouncilData` at `pcdata.go:7`. Pointer identity is the equality semantic — matches C `pcdata->council == pcdata->council`.
- `CharData.PCData.Clan *ClanData` at `pcdata.go:6`. `ClanData.ClanType int` at `clan.go:27`. Constants `CLAN_ORDER = ...`, `CLAN_GUILD = CLAN_ORDER + 1` at `enums.go:415-416`. Guild-talk gate: `ch.PCData.Clan.ClanType == types.CLAN_GUILD`.
- `CharData.IsImmortal()`, `CharData.IsNPC()`, `CharData.GetTrust()` — all on `character.go`. Used throughout.
- `CharData.Act.IsSet(PLR_SILENCE)` — existing gate on `DoYell`/`DoShout`/`DoGossip`/`DoTell`. Reuse.
- `CharData.Deaf BitVector` — already honored by `DoImmtalk` (`channels.go:91,111`), `BroadcastAuction` (`auction.go:37`). Reuse.
- `CharData.InRoom.RoomFlags.IsSet(ROOM_SILENCE)` — honored by `DoYell`/`DoShout`/`DoGossip`. Reuse.
- `translateFor(speaker, listener, text)` — package-private in `internal/act/comm.go:13` (NOT in `util/`). Wrap outgoing message text at the receiver-send site. `DoYell`/`DoShout`/`DoClantalk`/`DoSay` use it verbatim from within package `act`.
- `WorldRef.Descriptors` walk pattern — established at `channels.go:100-115`, `auction.go:29-44`, `clan.go:91-97`.
- `CouncilData.Name string` at `clan.go:50-51`. Used for the `"Newbie Council"` string match.
- Boot registration: `internal/boot/boot.go:408-414` is the Tier-9 channels block; new registrations go there.

### Gaps the port does NOT fill

| C feature | Go state | Decision |
|---|---|---|
| `PLR_WIZINVIS` level preamble | Absent from all Go channels | Match existing; queue uniform fix |
| `ROOM_LOGSPEECH` log append | Absent from all Go channels | Match existing; queue uniform fix |
| `is_ignoring` receiver filter | `Ignored []string` exists; no helper, no caller | Match existing; queue uniform fix |
| `AFLAG_SILENCE` area flag | Constant exists; no caller | Match existing; queue uniform fix |
| `NOT_AUTHED` gate | No auth split | Drop (Tier 9 precedent) |
| Per-AT_ color codes (AT_MUSIC etc.) | Constants exist; channels use inline `&Y`/`&G`/`&D` | Match Tier 9 style; separate per-AT_ pass |
| `WAIT_STATE` | Not honored on any Go channel | Match existing |
| `is_profane` filter | `#ifdef HMM` — disabled in C | Do not port |
| Charm-master redirect | No Go channel honors it | Match existing |
| `sysdata.pk_channels` AVTALK gate | AVTALK channel not in this tier | N/A |

---

## Task groups

Five groups. G0 delivers the shared helper and its tests. G1–G5 land the six channels in dependency order (parallelizable within the wave — all depend on G0 only). G6 is documentation.

### G0 — Shared `talkChannel` helper (M, ~2h)

**New file content in `internal/act/channels.go`** (extending the existing Tier-9 file; do NOT create a new file — follows the Tier 9 pattern of extending this file for each new channel handler).

Helper signature:

```go
// talkChannel is the shared skeleton for public/clan-family channels.
// Callers pass the channel bit, a canonical verb (e.g. "music", "racetalk"),
// and a recipient-filter predicate that decides whether each receiver's vch
// receives the broadcast. The helper owns: empty-arg rejection, PLR_SILENCE
// sender gate, deaf-sender block (with wartalk/yell exception), ROOM_SILENCE,
// self-echo, and the broadcast walk. Per-wrapper pre-gates (is-in-council,
// is-in-guild-clan, etc.) run before this helper is called.
//
// The deaf-auto-clear at C act_comm.c:514 is unreachable after the :511
// return and is NOT ported here (matching C semantics and DoImmtalk).
//
// AT_ color fidelity is deferred — use inline &Y/&G/&D per Tier 9 convention.
//
// Wait-state, PLR_WIZINVIS preamble, ROOM_LOGSPEECH append, AFLAG_SILENCE,
// and is_ignoring filter are NOT applied here — they are not honored by any
// current Go channel command; see the TODO queue for the uniform follow-up.
func talkChannel(
    ch *types.CharData,
    argument string,
    channel int,
    verb string,
    recipientFilter func(vch *types.CharData) bool,
)
```

Body (pseudocode; the worker writes the real Go):

```
if argument == "" {
    ch.Sendf("%s what?\n\r", capitalize(verb))   // "Music what?" / "Racetalk what?" / ...
    return
}
if !ch.IsNPC() && ch.Act.IsSet(PLR_SILENCE) {
    ch.Sendf("You can't %s.\n\r", verb)
    return
}
if ch.Deaf.IsSet(channel) &&
   channel != CHANNEL_WARTALK && channel != CHANNEL_YELL {
    ch.Sendf("You don't have the %s channel turned on. To turn it on, use the Channels command.\n\r", verb)
    return
}
if ch.InRoom != nil && ch.InRoom.RoomFlags.IsSet(ROOM_SILENCE) {
    ch.Send("You can't do that here.\n\r")
    return
}

// Self-echo.
ch.Sendf("&YYou %s '&G%s&Y'&D\n\r", verb, argument)

// Broadcast.
for _, d := range WorldRef.Descriptors {
    if d == nil || d.Connected != CON_PLAYING || d.Character == nil { continue }
    vch := d.Character
    if vch == ch { continue }
    if vch.Deaf.IsSet(channel) { continue }
    if vch.InRoom != nil && vch.InRoom.RoomFlags.IsSet(ROOM_SILENCE) { continue }
    if recipientFilter != nil && !recipientFilter(vch) { continue }
    heard := translateFor(ch, vch, argument)   // package-private in act; match DoYell/DoShout
    vch.Sendf("&Y%s %ss '&G%s&Y'&D\n\r", ch.Name, verb, heard)
}
```

`capitalize` is inline: `strings.ToUpper(verb[:1]) + verb[1:]` (guarded against empty `verb`).

**Tests (`internal/act/channels_test.go` — extend the Tier-9 test file):**

- `TestTalkChannel_EmptyArgPrintsVerbWhat` — `talkChannel(ch, "", CHANNEL_MUSIC, "music", nil)` emits `"Music what?\n\r"`.
- `TestTalkChannel_PLRSilenceBlocksSender` — player with `PLR_SILENCE` set gets `"You can't music.\n\r"`, no broadcast fires.
- `TestTalkChannel_DeafSenderBlockedWithoutAutoClear` — sender with `Deaf[CHANNEL_MUSIC]` set gets the `"You don't have..."` diagnostic, bit is **still set** after return. Mirrors the `DoImmtalk` no-auto-clear test.
- `TestTalkChannel_DeafSenderWartalkException` — sender with `Deaf[CHANNEL_WARTALK]` set is NOT blocked (C exception preserved). Self-echo fires. Broadcast fires for the one eligible listener.
- `TestTalkChannel_RoomSilenceBlocksSender` — sender in `ROOM_SILENCE` room gets `"You can't do that here.\n\r"`, no broadcast.
- `TestTalkChannel_ReceiverRoomSilenceSkipped` — listener in `ROOM_SILENCE` doesn't receive; other listeners do.
- `TestTalkChannel_ReceiverDeafSkipped` — listener with `Deaf[channel]` set doesn't receive; others do.
- `TestTalkChannel_SelfEchoNotBroadcastToSelf` — sender's own descriptor receives exactly the self-echo, NOT a second copy as a broadcast recipient.
- `TestTalkChannel_FilterClosureGatesReceivers` — pass a filter that rejects every listener; no one receives the broadcast; self-echo still fires.
- `TestTalkChannel_TranslateForIsApplied` — listener with mismatched `Speaking` sees translated (or identical) text — asserts the helper threads through `translateFor` (package-private in `act`), not that translation itself works (that's `translateFor`'s own test). A single assertion that the listener's received text is whatever `translateFor` returns for the pair, which for matched languages equals `argument`.

**Mutation verification (use `Edit` round-trip per `_shared.md` → Mutation Verification Safety):**

- Mutate `channel != CHANNEL_WARTALK && channel != CHANNEL_YELL` → `channel != CHANNEL_WARTALK || channel != CHANNEL_YELL` (the C-bug form). Assert `TestTalkChannel_DeafSenderWartalkException` fails. Revert with `Edit`.
- Mutate `continue` in the `vch == ch` guard to `// continue` (commented). Assert `TestTalkChannel_SelfEchoNotBroadcastToSelf` fails. Revert.
- Mutate the `recipientFilter != nil && !recipientFilter(vch)` guard to `false &&`. Assert `TestTalkChannel_FilterClosureGatesReceivers` fails. Revert.

**Files:** `internal/act/channels.go` (extend), `internal/act/channels_test.go` (extend). No other files touched.

---

### G1 — `DoMusic` wrapper (S, ~30min)

**Body:**

```go
func DoMusic(ch *types.CharData, argument string) {
    if ch.IsNPC() {
        ch.Send("Huh?\n\r")
        return
    }
    talkChannel(ch, argument, types.CHANNEL_MUSIC, "music", nil)
}
```

No per-wrapper gate beyond `IsNPC` (C drops only `NOT_AUTHED` which Go has no concept of). `nil` filter = public channel.

**Boot:** `reg.Register(&command.Command{Name: "music", DoFun: act.DoMusic, Position: types.POS_SLEEPING, Level: 0})`. Position matches Tier 9 `gtell`/`auction` (sleeping allowed) — music is a leisure channel.

**Tests:**

- `TestDoMusic_EmptyArgPrintsMusicWhat` — `"Music what?\n\r"`.
- `TestDoMusic_NPCGetsHuh`.
- `TestDoMusic_PublicBroadcastReachesAllPlayers` — three connected players (sender + two). Both non-senders receive. Self-echo fires.
- `TestDoMusic_DeafListenerSkipped`.
- `TestDoMusic_BootRegistered` — `Interpret(ch, "music la")` routes to `DoMusic`. Follows the Tier 9 pattern.

### G2 — `DoRacetalk` wrapper (S, ~30min)

**Body:**

```go
func DoRacetalk(ch *types.CharData, argument string) {
    if ch.IsNPC() {
        ch.Send("Huh?\n\r")
        return
    }
    filter := func(vch *types.CharData) bool {
        return vch.Race == ch.Race
    }
    talkChannel(ch, argument, types.CHANNEL_RACETALK, "racetalk", filter)
}
```

**Boot:** `reg.Register(&command.Command{Name: "racetalk", DoFun: act.DoRacetalk, Position: types.POS_SLEEPING, Level: 0})`.

**Tests:**

- `TestDoRacetalk_EmptyArgPrintsRacetalkWhat`.
- `TestDoRacetalk_NPCGetsHuh`.
- `TestDoRacetalk_SameRaceReceives` — sender race=1, listener race=1 → receives.
- `TestDoRacetalk_DifferentRaceFiltered` — sender race=1, listener race=2 → does NOT receive.
- `TestDoRacetalk_BootRegistered`.

### G3 — `DoWartalk` wrapper (S, ~30min)

**Body:**

```go
func DoWartalk(ch *types.CharData, argument string) {
    if ch.IsNPC() {
        ch.Send("Huh?\n\r")
        return
    }
    // Pkill-only sender gate (C act_comm.c:432-436).
    if ch.PCData == nil || (uint32(ch.PCData.Flags) & types.PCFLAG_DEADLY) == 0 {
        ch.Send("Peacefuls have no need to use wartalk.\n\r")
        return
    }
    filter := func(vch *types.CharData) bool {
        return vch.PCData != nil && (uint32(vch.PCData.Flags) & types.PCFLAG_DEADLY) != 0
    }
    talkChannel(ch, argument, types.CHANNEL_WARTALK, "war", filter)
}
```

The verb is `"war"` (matches C `do_wartalk` at `:4619`; yields `"You war 'foo'"` and `"X wars 'foo'"` — yes, that's how C reads too).

**Boot:** `reg.Register(&command.Command{Name: "wartalk", DoFun: act.DoWartalk, Position: types.POS_SLEEPING, Level: 0})`.

**Tests:**

- `TestDoWartalk_EmptyArgPrintsWarWhat` — `"War what?\n\r"` (verb is `"war"`, not `"wartalk"`).
- `TestDoWartalk_NPCGetsHuh`.
- `TestDoWartalk_PeacefulSenderBlocked` — PC without `PCFLAG_DEADLY` gets `"Peacefuls have no need to use wartalk.\n\r"`.
- `TestDoWartalk_PkillToPkillReceives` — both sides have `PCFLAG_DEADLY`, listener receives.
- `TestDoWartalk_PkillToPeacefulFiltered` — sender pkill, listener peaceful, listener does NOT receive.
- `TestDoWartalk_DeafSenderNotBlocked_WartalkException` — sender with `Deaf[CHANNEL_WARTALK]` set is NOT blocked; broadcast fires. **This is the only `DoXxx` that exercises the wartalk exception**; duplicate of the helper-level test but through the wrapper for regression guard.
- `TestDoWartalk_BootRegistered`.

### G4 — `DoCouncilTalk` wrapper (S, ~30min)

**Body:**

```go
func DoCouncilTalk(ch *types.CharData, argument string) {
    if ch.IsNPC() || ch.PCData == nil || ch.PCData.Council == nil {
        ch.Send("Huh?\n\r")
        return
    }
    council := ch.PCData.Council
    filter := func(vch *types.CharData) bool {
        return vch.PCData != nil && vch.PCData.Council == council
    }
    talkChannel(ch, argument, types.CHANNEL_COUNCIL, "counciltalk", filter)
}
```

Function name is `DoCouncilTalk` (matches Go camelcase convention and the shipped `DoClantalk` style — `talk` staying lowercase as one word).

**Boot:** `reg.Register(&command.Command{Name: "counciltalk", DoFun: act.DoCouncilTalk, Position: types.POS_SLEEPING, Level: 0})`.

**Tests:**

- `TestDoCouncilTalk_EmptyArgPrintsCounciltalkWhat`.
- `TestDoCouncilTalk_NPCGetsHuh`.
- `TestDoCouncilTalk_NoCouncilGetsHuh` — PC with `PCData.Council == nil` gets `"Huh?\n\r"` (NOT the "you aren't in a council" C-absent-equivalent; match the C wrapper verbatim).
- `TestDoCouncilTalk_SameCouncilReceives` — two players pointing to the same `CouncilData` pointer → listener receives.
- `TestDoCouncilTalk_DifferentCouncilFiltered` — different pointers → listener does NOT receive.
- `TestDoCouncilTalk_NilPCDataListenerSkipped` — safety: listener with `PCData == nil` doesn't panic (NPCs bypass via the filter's nil guard).
- `TestDoCouncilTalk_BootRegistered`.

### G5 — `DoGuildTalk` + `DoNewbieChat` (S, ~45min — shared file section)

**`DoGuildTalk` body:**

```go
func DoGuildTalk(ch *types.CharData, argument string) {
    if ch.IsNPC() || ch.PCData == nil ||
       ch.PCData.Clan == nil ||
       ch.PCData.Clan.ClanType != types.CLAN_GUILD {
        ch.Send("Huh?\n\r")
        return
    }
    clan := ch.PCData.Clan
    filter := func(vch *types.CharData) bool {
        return vch.PCData != nil && vch.PCData.Clan == clan
    }
    talkChannel(ch, argument, types.CHANNEL_GUILD, "guildtalk", filter)
}
```

**`DoNewbieChat` body:**

```go
func DoNewbieChat(ch *types.CharData, argument string) {
    if ch.IsNPC() {
        ch.Send("Huh?\n\r")
        return
    }
    if !isNewbieChannelMember(ch) {
        ch.Send("Huh?\n\r")
        return
    }
    filter := func(vch *types.CharData) bool {
        return isNewbieChannelMember(vch)
    }
    talkChannel(ch, argument, types.CHANNEL_NEWBIE, "newbiechat", filter)
}

// isNewbieChannelMember: immortal OR member of a council named "Newbie Council".
// Matches C `do_newbiechat` sender gate (:937-944) and `talk_channel` receiver
// filter (:717-721). C's NOT_AUTHED branch is collapsed into the "immortal"
// disjunct since Go has no unauthed state.
func isNewbieChannelMember(ch *types.CharData) bool {
    if ch.IsImmortal() {
        return true
    }
    if ch.PCData != nil && ch.PCData.Council != nil &&
       strings.EqualFold(ch.PCData.Council.Name, "Newbie Council") {
        return true
    }
    return false
}
```

**Boot:** two registrations, adjacent:

```go
reg.Register(&command.Command{Name: "guildtalk", DoFun: act.DoGuildTalk, Position: types.POS_SLEEPING, Level: 0})
reg.Register(&command.Command{Name: "newbiechat", DoFun: act.DoNewbieChat, Position: types.POS_SLEEPING, Level: 0})
```

**Tests:**

- `TestDoGuildTalk_NPCGetsHuh`.
- `TestDoGuildTalk_NoClanGetsHuh`.
- `TestDoGuildTalk_WrongClanTypeGetsHuh` — PC in a clan of `ClanType != CLAN_GUILD` (e.g. `CLAN_ORDER`) → `"Huh?"`.
- `TestDoGuildTalk_SameGuildReceives`.
- `TestDoGuildTalk_DifferentClanFiltered`.
- `TestDoGuildTalk_BootRegistered`.
- `TestDoNewbieChat_NPCGetsHuh`.
- `TestDoNewbieChat_MortalNonNewbieGetsHuh` — mortal with `Council == nil` → `"Huh?"`.
- `TestDoNewbieChat_MortalNewbieCouncilReceives` — mortal whose `Council.Name == "Newbie Council"` (case-insensitive) → broadcast succeeds.
- `TestDoNewbieChat_ImmortalReceives` — immortal (no council) receives the broadcast as a bystander.
- `TestDoNewbieChat_CaseInsensitiveCouncilName` — `"newbie council"` and `"NEWBIE COUNCIL"` both satisfy.
- `TestDoNewbieChat_BootRegistered`.

### G6 — Documentation updates (XS, ~15min)

- Append completion record to this plan when the six channels land.
- Update `TODO.md`: move "Extra channels (music/newbiechat/..." entry to the Done section (or the orchestrator's merge target); add follow-up items for the uniform fixes listed below.
- Update `CHANGELOG.md` entry citing the plan and landed files.
- `CLAUDE.md` index: add a new Tier-N row pointing at this plan.

---

## Acceptance criteria

### Helper (G0)

A1. `talkChannel` helper in `internal/act/channels.go` with the exact signature above.
A2. Empty-arg path emits `"<Verb> what?\n\r"` with first-char uppercase. Mutation-verified: changing the capitalize call to pass-through fails the test.
A3. `PLR_SILENCE` sender gate emits `"You can't <verb>.\n\r"` and does NOT broadcast.
A4. Deaf-sender block emits the `"You don't have the <verb> channel turned on..."` diagnostic and does NOT auto-clear the deaf bit.
A5. Deaf-sender block is skipped when `channel == CHANNEL_WARTALK` (exception preserved exactly per C intent). `CHANNEL_YELL` is listed in the same exception but no caller in this plan uses `CHANNEL_YELL` through this helper — the exception is carried for future callers and regression-guarded via a helper-level test.
A6. `ROOM_SILENCE` on sender's room emits `"You can't do that here.\n\r"` and does NOT broadcast.
A7. Broadcast walk skips: `d.Character == nil`, `Connected != CON_PLAYING`, `vch == ch`, `Deaf[channel]`, listener in `ROOM_SILENCE`, `recipientFilter(vch) == false`.
A8. Broadcast applies `translateFor(ch, vch, argument)` per receiver (package-private in `internal/act/comm.go:13`).
A9. No `PLR_WIZINVIS` preamble. No `ROOM_LOGSPEECH` file append. No `is_ignoring` filter. No `AFLAG_SILENCE` check. No wait-state. (All deliberate divergences matching existing Go channel commands; covered by absence-of-behavior in tests plus explicit code comment.)

### Per-channel wrappers (G1–G5)

A10. **`DoMusic`**: NPC → `"Huh?"`. Empty arg → `"Music what?"`. Public broadcast reaches every non-deaf, non-silenced, playing listener. Boot-registered as `music`.
A11. **`DoRacetalk`**: NPC → `"Huh?"`. Empty arg → `"Racetalk what?"`. Listener with matching `Race` receives; different `Race` skipped. Boot-registered as `racetalk`.
A12. **`DoWartalk`**: NPC → `"Huh?"`. Non-pkill sender → `"Peacefuls have no need to use wartalk.\n\r"`. Pkill sender with no-pkill listener: listener skipped. Pkill-to-pkill: listener receives. Deaf-sender exception: pkill sender with `Deaf[CHANNEL_WARTALK]` is NOT blocked (C exception). Boot-registered as `wartalk`.
A13. **`DoCouncilTalk`**: NPC or no-council sender → `"Huh?"`. Same-council listener receives (pointer equality). Different-council listener skipped. Boot-registered as `counciltalk`.
A14. **`DoGuildTalk`**: NPC, no-clan, or `ClanType != CLAN_GUILD` sender → `"Huh?"`. Same-guild-clan listener receives. Boot-registered as `guildtalk`.
A15. **`DoNewbieChat`**: NPC → `"Huh?"`. Mortal with no council or non-"Newbie Council" council → `"Huh?"`. Immortal sender succeeds. Mortal with `Council.Name == "Newbie Council"` (case-insensitive) succeeds. Listener filter: immortal or newbie-council member. Boot-registered as `newbiechat`.

### Regression

A16. `go test -count=3 ./...` green across all packages.
A17. No regression in Tier 9 (`DoImmtalk` / `DoGtell` / `DoClantalk` / `DoAuction` / `BroadcastAuction` / `DoChannels`) — their tests still pass unchanged.
A18. No regression in `DoYell` / `DoShout` / `DoGossip` / `DoTell` / `DoReply` — `PLR_SILENCE` gates still honored.

**Total: 18 acceptance criteria across 5 execution task groups (G0–G5) + 1 documentation group (G6) = 6 task groups.**

---

## Scope cuts

- **No retrofit of Tier 9 channels onto `talkChannel`.** `DoImmtalk`, `DoGtell`, `DoClantalk`, `DoAuction` stay open-coded. Revisit when an additional wave of channels (e.g. `quest`, `ask`, `muse`, `think`, `avtalk`) lands.
- **No `PLR_WIZINVIS` preamble.** Queued as a uniform follow-up covering all 10 channels.
- **No `ROOM_LOGSPEECH` file append.** Queued as a uniform follow-up.
- **No `is_ignoring` receiver filter.** `PCData.Ignored` exists but has no helper in the port. Queued.
- **No `AFLAG_SILENCE` check.** Constant exists; no caller. Queued.
- **No per-`AT_` color fidelity.** Tier 9 precedent: use inline `&Y`/`&G`/`&D`. Separate uniform pass queued.
- **No `WAIT_STATE` / `Lag`.** Not honored on any Go channel.
- **No `NOT_AUTHED` gate.** Go has no auth split.
- **No `DoOrdertalk` / `DoAsk` / `DoQuest` / `DoTraffic`.** Different ports, not in this plan's 6-channel scope.
- **No interpreter prefix-match for the new command names.** `music`, `racetalk` etc. are full words; prefix-matching within a single token (e.g., `mu`) is an interpreter concern out of scope for this plan.
- **No `cset` / sysdata-configurable variants** (e.g., `sysdata.pk_channels` for wartalk). `sysdata` is partially ported; wiring it to the helper is a separate concern.

---

## Open questions

1. **Verb capitalization for error messages.** `"Music what?"` with capital M is produced by `strings.ToUpper(verb[:1]) + verb[1:]`. For `"racetalk"` this yields `"Racetalk what?"`. For `"counciltalk"` it yields `"Counciltalk what?"`. Matches C's `UPPER(buf[0])` at `act_comm.c:495`. No open question — confirmed mechanical.
2. **`"newbiechat"` vs `"NewbieChat"` in display text.** The verb passed to `talkChannel` is `"newbiechat"`, yielding `"Newbiechat what?"` and `"X newbiechats 'y'"`. C produces the same. Match exactly.
3. **Does `"Newbie Council"` exist in stock SMAUG data?** **VERIFIED ABSENT (2026-04-18 audit).** `/home/eilidh/src/smaug/db/councils/council.lst` is empty — no councils are shipped in stock data. Consequence: in default installations, ONLY immortals can use `newbiechat` (the `isNewbieChannelMember` mortal branch never triggers). Matches C semantics exactly for the same data configuration. Help-text references to "Newbie Council" exist (`db/area/help.are:3955`, `:18480`, etc.) but there is no council-data entry. **Not blocking execution; feature still works for any server that creates a council named "Newbie Council".** The planned G5 test `TestDoNewbieChat_MortalNewbieCouncilReceives` must construct the council data in-test (pointer identity via a test fixture) — it cannot depend on production data.
4. **Alias registration.** C registers only the full command names; no single-character aliases (`:` is for immtalk; `;` for gtell). The six new channels have no aliases. Match.
5. **Position for `wartalk`.** Tier 9 `gtell` / `auction` use `POS_SLEEPING`. `DoClantalk` uses `POS_RESTING`. `wartalk` in active pkill combat — should it allow `POS_FIGHTING`? C's commands.dat and `tables.c` indicate position requirements via the command table; stock SMAUG position for wartalk is `POS_SLEEPING` (verified by absence of a higher requirement in the grep). **Decision:** use `POS_SLEEPING` for all six, matching Tier 9 precedent. Revisit if gameplay-testing shows pkill players need to yell war strategy mid-fight (in which case `POS_FIGHTING` for wartalk specifically).
6. **"Council members" for `counciltalk` include NPC followers?** C filter at `:744-750` explicitly rejects NPC receivers (`IS_NPC(vch) continue`). Helper's receiver predicate enforces this via `vch.PCData != nil` in the wrapper's filter closure. Verified matching.
7. **Self-ignore on `counciltalk`?** If `ch.PCData.Council` is the same pointer as `vch.PCData.Council` for `vch == ch`, the `vch == ch` early-skip in `talkChannel` fires before the filter. No double-echo risk.
8. **Should `DoNewbieChat` be level-gated for mortals?** C has no explicit level cap — any mortal in the "Newbie Council" can use it regardless of level. This contradicts the plan dispatch notes which said `newbiechat` "has a level cap in C (only low-level players)". **Grep confirms**: no level cap in C. The gating is purely by council membership. Plan spec corrects the dispatch note.
9. **`DoGuildTalk` scope.** C also gates on `!NOT_AUTHED`. Dropped per project convention. No open question.

---

## Mutation verification summary

The worker executes these mutations using `Edit` round-trips (per `_shared.md` → Mutation Verification Safety — no `git checkout`, no `git restore`, no stashes). Each mutation must cause the named test to fail; revert must restore green.

| Gate | Mutation | Expected failure | Group |
|---|---|---|---|
| Deaf-sender wartalk exception | `channel != CHANNEL_WARTALK && channel != CHANNEL_YELL` → `channel != CHANNEL_WARTALK \|\| channel != CHANNEL_YELL` | `TestTalkChannel_DeafSenderWartalkException`, `TestDoWartalk_DeafSenderNotBlocked_WartalkException` | G0, G3 |
| Self-echo not re-broadcast | Remove the `vch == ch` continue guard | `TestTalkChannel_SelfEchoNotBroadcastToSelf` | G0 |
| Recipient filter application | Collapse `recipientFilter != nil && !recipientFilter(vch)` to `false` | `TestDoRacetalk_DifferentRaceFiltered`, `TestDoCouncilTalk_DifferentCouncilFiltered`, `TestDoWartalk_PkillToPeacefulFiltered` | G0, G2, G3, G4 |
| Pkill sender gate | Flip `== 0` → `!= 0` on sender's `PCFLAG_DEADLY` check | `TestDoWartalk_PeacefulSenderBlocked` | G3 |
| Council pointer equality | Flip `==` → `!=` on `vch.PCData.Council == council` | `TestDoCouncilTalk_SameCouncilReceives` | G4 |
| Clan-type guild check | Flip `ClanType != CLAN_GUILD` → `ClanType != CLAN_ORDER` | `TestDoGuildTalk_WrongClanTypeGetsHuh` | G5 |
| Newbie-council case-insensitivity | `strings.EqualFold` → `strings.Compare ... == 0` (case-sensitive) | `TestDoNewbieChat_CaseInsensitiveCouncilName` | G5 |

Revert each with `Edit` applying the inverse change. Do NOT use `git checkout` / `git restore` / `git stash`.

---

## Risk

- **Low.** The template is established (Tier 9); all target predicates and helpers already exist; no schema changes; no persistence changes. The only judgement call is the factor-out decision, and it's bounded by an explicit non-retrofit scope.
- **Specific risk: regression on Tier 9 channels.** Mitigation — `talkChannel` is a new exported-lowercase function. Tier 9 commands do not call it. Tests A17 guard.
- **Specific risk: wartalk deaf exception drift.** The C condition `(channel != CHANNEL_WARTALK || channel != CHANNEL_YELL)` is always-true (a C-side bug). Go port uses the intent form `&&`. Two tests (helper + wrapper) guard the behavior; the mutation test proves the guard bites.
- **Specific risk: newbie-council data absence.** **Verified (audit 2026-04-18):** stock `db/councils/council.lst` is empty. In default installations only immortals can use `newbiechat`. This is a data-driven behavior matching C exactly. Not a code defect; the `newbiechat` command registers regardless, and server admins who create a council named "Newbie Council" unlock it for mortals.

---

## Relevant file paths

- **C:** `/home/eilidh/src/smaug/src/act_comm.c` — `:392-858` (`talk_channel`), `:935-947` (`do_newbiechat`), `:975-990` (`do_counciltalk`), `:993-1009` (`do_guildtalk`), `:1012-1021` (`do_music`), `:4612-4621` (`do_wartalk`), `:4624-4632` (`do_racetalk`). `/home/eilidh/src/smaug/db/system/en/commands.dat` — the six entries.
- **Go new:** none. All additions extend existing files.
- **Go modified:** `/home/eilidh/src/smaug/smaug-go/internal/act/channels.go` (+helper, +6 wrappers), `/home/eilidh/src/smaug/smaug-go/internal/act/channels_test.go` (+~35 tests), `/home/eilidh/src/smaug/smaug-go/internal/boot/boot.go` (+6 registrations), `/home/eilidh/src/smaug/CHANGELOG.md`, `/home/eilidh/src/smaug/TODO.md`, `/home/eilidh/src/smaug/CLAUDE.md` (index row).
- **Seams:** `/home/eilidh/src/smaug/smaug-go/internal/types/enums.go:910-943` (CHANNEL_*), `/home/eilidh/src/smaug/smaug-go/internal/types/constants.go:648` (`PCFLAG_DEADLY`), `/home/eilidh/src/smaug/smaug-go/internal/types/enums.go:415-416` (`CLAN_ORDER` L415, `CLAN_GUILD` L416), `/home/eilidh/src/smaug/smaug-go/internal/types/pcdata.go:6-7` (Clan/Council pointers), `/home/eilidh/src/smaug/smaug-go/internal/types/character.go:84` (`Race`), `/home/eilidh/src/smaug/smaug-go/internal/act/comm.go:13` (`translateFor` — package-private, unexported; same `act` package, no import), `/home/eilidh/src/smaug/smaug-go/internal/act/channels.go:30-65` (existing `channelToggleTable` — no changes needed; all six channel bits already enrolled).
- **Reference plans:** `/home/eilidh/src/smaug/smaug-go/doc/plan-channels.md` (authoritative for the pattern + Tier 9 completion record), `/home/eilidh/src/smaug/smaug-go/doc/plan-do-channels.md` (DoChannels toggle + `channelToggleTable`), `/home/eilidh/src/smaug/smaug-go/doc/phase6-roadmap.md:197-201` (this tier's scope line).

---

## Rough total effort

G0 + G1 + G2 + G3 + G4 + G5 + G6 ≈ half a day with tests. All groups can be dispatched as a single wave of parallel workers after G0 lands (G0 is the serial prerequisite for G1–G5).

---

## Verification (plan self-review)

No `Agent` tool available in the manager dispatch for this environment; external adversary pass deferred. Self-review against plan-authoring checklist:

- [x] All 6 channels have explicit C line-citations with file path.
- [x] All 6 channel constants verified present in `internal/types/enums.go`: `CHANNEL_MUSIC` at L915, `CHANNEL_COUNCIL` at L927, `CHANNEL_GUILD` at L928, `CHANNEL_NEWBIE` at L932, `CHANNEL_WARTALK` at L933, `CHANNEL_RACETALK` at L934 (verified by external audit — all correct).
- [x] `channelToggleTable` in `DoChannels` confirmed to already include all 6 (channels.go:30-65). No new toggle-table rows needed; deaf-bit persistence already works for all 6 post plan-do-channels.md.
- [x] Per-channel gate requirements checked and confirmed portable: `IS_PKILL` → `PCFLAG_DEADLY` (verified via mudprog ifcheck); council pointer equality (verified — `PCData.Council *CouncilData`); clan-type check (verified — `CLAN_GUILD` constant); race equality (verified — `CharData.Race int`); immortal check (verified — `ch.IsImmortal()`).
- [x] Factor-vs-open-code decision stated with rationale. Threshold ("once 5+ channels land") met. Non-retrofit scope explicit.
- [x] 6 task groups. 18 acceptance criteria. Within the 4-7 task-group range.
- [x] Mutation-verification plan with `Edit` round-trip only (no banned git commands).
- [x] Payload-by-reference throughout — every code snippet is either the proposed signature/body or a short identifier-level reference with `file:line`.
- [x] Worker prompts implied by G0-G5 are each single-file scoped and self-contained.
- [x] Non-scope items listed explicitly with the reason for deferral.
- [x] Open question 8 corrects a dispatch-prompt inaccuracy (the dispatch said "newbiechat has a level cap in C (only low-level players)" — grep of `do_newbiechat` confirms there is no level cap; only a council-name string match).

Caveats acknowledged:

- Self-review is not adversary-independent. The plan cannot PASS the "adversary verified" bar without an external run.
- The factor-out decision ships a new helper that is not used by the four Tier 9 channels. A future orchestrator may elect to retrofit — that's a separate plan.
- Verbs like `"war"` will produce self-echo `"You war 'strategy'"` which reads awkwardly in English but matches C verbatim. Editorial polish is a separate concern; fidelity wins.

---

## Completion record

*(To be filled on landing.)*
