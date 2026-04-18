# Plan — Communication Channels (immtalk / gtell / auction)

Source audit: `audit-2026-04-17.md` P1. TODO ref: `TODO.md:37`.

## Goal

Port the top-3 communication channels flagged as unblocking for multi-player coordination:

1. **`immtalk`** — immortal-only chat (alias `:`).
2. **`gtell`** — group-tell, walks `ch.Leader` graph (alias `;`).
3. **`auction`** — channel broadcast helper (`talk_auction`) + stub `DoAuction` command.

**Important scoping correction:** C's `do_auction` is the full auction-house system (list / bid / stop, ~250 LOC in `act_obj.c` + `update.c` ticker). That's a P3-sized task. What this plan actually delivers is the **`talk_auction` broadcast helper** (~20 LOC) plus a stubbed `DoAuction` that emits "The auction house is currently closed.". The helper is the reusable piece for when the full system lands in Phase 6.

## Per-channel mapping

| Channel | C function | C file:line | Alias | Trust gate | Channel bit | Audience |
|---|---|---|---|---|---|---|
| `immtalk` | `do_immtalk` → `talk_channel(..., CHANNEL_IMMTALK, "immtalk")` | `act_comm.c:1327` + helper `:392` | `:` | `LEVEL_IMMORTAL` (51) | `CHANNEL_IMMTALK` (`enums.go:914`) | all `CON_PLAYING` with `Trust >= 51` |
| `gtell` | `do_gtell` (does NOT go through `talk_channel`) | `act_comm.c:4217` | `;` | 0 | none (not a deaf-filtered channel) | `IsSameGroup(gch, ch)` — includes sleepers |
| `auction` (helper) | `talk_auction` | `act_comm.c:4309` | — | listener `Trust >= 5` | `CHANNEL_AUCTION` (`enums.go:911`) | `CON_PLAYING`, `!Deaf[AUCTION]`, `!ROOM_SILENCE` |
| `auction` (command) | `do_auction` | `act_obj.c:3775+` | — | full subsystem | — | **DEFERRED to Phase 6** |

## Shared infrastructure survey

- **`talk_channel` generic helper — NOT ported.** Nearest templates: `DoShout` (`act/comm.go:147`), `DoGossip` (`:124`), `DoClantalk` (`act/clan.go:77`). Each open-codes the loop. **Recommendation:** keep open-coding for these three — divergent audience predicates (imm-only / group-only / deaf+trust) don't share enough structure to factor yet. Revisit when `muse`/`think`/`wartalk`/`racetalk`/`council`/`guild` land.
- **Channel bits:** fully defined in `types/enums.go:910-943` (iota enum matching C order).
- **Deaf bitvector:** `CharData.Deaf BitVector` (`character.go:164`); persists correctly (`persist/player_test.go:700`).
- **`PLR_SILENCE`:** defined (`enums.go:963`), set/cleared by `DoSilence` (`wiz.go:896-901`), **but not currently checked by any Go communication command**. Latent bug — fix opportunistically in G1.
- **`PLR_NO_TELL`:** defined (`enums.go:966`), honored by `DoTell` / `DoReply`. Reuse for `gtell` sender gate.
- **`ROOM_SILENCE`:** defined (`enums.go:711`), honored by `DoYell` / `DoShout` / `DoGossip`.
- **Group helpers:** `CharData.Leader` exists (`character.go:10`), `DoGroup` uses `rch.Leader == ch` inline (`group.go:75`). **No `IsSameGroup` helper** — port one as a prereq.
- **Language / scrambling:** `translateFor(speaker, listener, text)` at `act/comm.go:13` — reuse verbatim.
- **Descriptor walk:** pattern at `info.go:307`, `wiz.go:340`, `clan.go:91`: `for _, d := range WorldRef.Descriptors { if d.Connected != CON_PLAYING || d.Character == nil { continue } … }`.
- **Character walk (for gtell):** `for _, gch := range WorldRef.Characters`. **Verified:** `game/loop.go:711` adds on login; `game/loop.go:875` removes on `closeDescriptor`. So `Characters` retains chars who are **asleep-but-still-connected** (e.g., used `sleep`), but NOT disconnected players. C's `first_char` retains *all* loaded chars until explicit extract; Go removes on disconnect. Gtell to a disconnected "sleeper" is a C-only behavior — Go will behave differently by design (non-blocking divergence; disconnected players can't receive messages anyway).
- **Color codes:** inline `&Y` / `&W` / `&G` / `&D` in format strings; expanded downstream. Match existing `DoClantalk` style.
- **`NOT_AUTHED`** from C — not in Go port. Every player is authed. Drop that C gate.
- **Testclient:** `Harness.Dial(t)` can be called repeatedly to open multiple simultaneous telnet clients — the harness is a real server with a real listener. However, **no existing test drives two simultaneously-logged-in players** (the closest, `login_test.go:118-134`, is sequential: close first, then open second). G3's integration test will be the first concurrent-login scenario and should be treated as new infrastructure. A single concurrent dial + `QuickLogin` for two different characters is the smallest proof-of-concept needed; add it as a helper in `testclient/` if absent.

## Task groups

### G1 — Shared infrastructure (S, ~1h)

- Add `handler.IsSameGroup(a, b *CharData) bool` matching C `act_comm.c:4293-4301`. `groupLeader(x) := if x.Leader != nil { x.Leader } else { x }`; equal if both resolve to the same leader. Tests: self, nil, leader-and-follower, two followers of same leader, unrelated chars, transitive (A leads B; B leads C should NOT equal A+C since leaders form a flat graph in SMAUG).
- Opportunistic: add `PLR_SILENCE` sender check to existing `DoShout`, `DoGossip`, `DoYell`, `DoTell` — `if !ch.IsNPC() && ch.Act.IsSet(PLR_SILENCE) { ch.Send("You can't do that.\n\r"); return }`. Flag in commit message as a latent-bug fix.
- **Files:** `internal/handler/group.go` (or extend existing handler file); `internal/act/comm.go`.
- **Tests:** `handler/group_test.go` new; existing comm tests extended for `PLR_SILENCE`.

### G2 — `DoImmtalk` (S, ~45min)

Body in `act/comm.go` (or new `act/channels.go`):

- Empty-arg: "Immtalk what?".
- Mortal (`!ch.IsImmortal()`): "Huh?" and return.
- `PLR_SILENCE` on sender: "You can't do that." and return.
- **Deaf-sender block (match C `act_comm.c:505-513`):** if `ch.Deaf.IsSet(CHANNEL_IMMTALK)`, send `"You don't have the immtalk channel turned on...\n\r"` and return. C does NOT auto-clear the deaf bit here — the `xREMOVE_BIT` at `:514` is unreachable when the sender is deaf (the function returned on `:511`).
- Self-echo: `ch.Sendf("&YYou %s '&G%s&Y'&D\n\r", verbForTense(...), argument)` or match C format exactly.
- Broadcast: iterate `WorldRef.Descriptors`, filter `CON_PLAYING && !nil && Trust >= 51 && d.Character != ch && !d.Character.Deaf.IsSet(CHANNEL_IMMTALK)`, emit `&Y%s&G>&Y %s&D\n\r`.

Register `immtalk` (Level 51) and `:` (Level 51) in `boot/boot.go`.

- **Tests:** mortal denied; imm-to-imm delivered; imm-to-mortal not delivered; auto-unmute on send; `PLR_SILENCE` blocks; `:` alias routes to same handler.

### G3 — `DoGtell` (S, ~45min)

Body in same file:

- Empty-arg: "Tell your group what?".
- `PLR_NO_TELL` on sender: "Your message didn't get through!" and return.
- Iterate `WorldRef.Characters`, include `gch` where `handler.IsSameGroup(gch, ch)`.
- Per-recipient: `gch.Sendf("%s tells the group '%s'\n\r", name, message)` (match C format at `act_comm.c:4251`).
- Asleep-but-connected `gch`: `gch.Send` works normally (they have a `Desc`). Disconnected players aren't in `WorldRef.Characters` so they don't receive — divergence from C noted in the seams section, non-blocking.

Register `gtell` (Level 0) and `;` (Level 0).

- **Tests:** solo player sees only own echo; A leads {A, B}, C unrelated → A's gtell reaches B, not C; NPC leader with follower → follower still hears; `PLR_NO_TELL` on sender blocks; sleeper B with `Desc == nil` doesn't crash on broadcast; `;` alias works.
- **Integration test (testclient):** two clients, `follow` + `group` to form a group, assert the non-sender receives the gtell line.

### G4 — `talk_auction` helper + stub `DoAuction` (M, ~2h)

- New file `internal/act/auction.go`. Exported `BroadcastAuction(message string)`:
  - Iterate `WorldRef.Descriptors`, filter `CON_PLAYING && !nil && Trust >= 5 && !Deaf[CHANNEL_AUCTION] && !InRoom.RoomFlags.IsSet(ROOM_SILENCE)`.
  - Emit `"Auction: %s\n\r"`.
- `DoAuction(ch, argument)`: NPC check → "Huh?"; non-empty arg → "The auction house is currently closed. (See the Phase-6 roadmap.)" and return. Do NOT broadcast — there's no real auction system yet.
- Register `auction` (Level 3 per C).

- **Tests:** `BroadcastAuction` unit — trust gate (level-4 doesn't see), `ROOM_SILENCE` skip, `Deaf` skip, happy path. `DoAuction` stub: closed-message response; NPC no-op. Add a TODO in `TODO.md` pointing at the Phase-6 full `do_auction` task.

### G5 — Full auction system (DEFERRED to Phase 6)

Port `do_auction` state machine (list/bid/stop), `noauction` list, item escrow, gold handling, auction-tick in `update.c`. Out of scope for this plan; placeholder entry only.

## Acceptance criteria

**`immtalk`:**

1. Mortal typing `immtalk hi` → "Huh?"; no other descriptor receives output.
2. Two immortals A and B online: A's `immtalk hi` → B's descriptor receives a line containing "A" and "hi".
3. Mortal M online alongside B: M sees nothing from A's immtalk.
4. `:` alias: `: hi` behaves identically to `immtalk hi`.
5. Imm A with `Deaf[IMMTALK]` set sends an immtalk → sender is **blocked** with "You don't have the immtalk channel turned on..." and broadcast does not fire; flag is NOT cleared. (Matches C `act_comm.c:505-513`; to re-enable, a separate `channels` command / toggle is needed.)
6. Imm A with `PLR_SILENCE` set: send blocked with explanation.

**`gtell`:**

7. Solo player: `gtell hi` → only self echo; no crash.
8. A leads {A, B}; C unrelated: A's `gtell plan` → B's descriptor receives the line; C does not.
9. `PLR_NO_TELL` on sender: "Your message didn't get through!".
10. Asleep group member B (no `Desc`): A's gtell does not panic.
11. `;` alias works.

**`auction`:**

12. `DoAuction whatever` → "The auction house is currently closed." (no broadcast fires).
13. Unit test: direct `BroadcastAuction("Hi")` → player with `Trust < 5` does not see; `Trust >= 5 && !Deaf[AUCTION] && !ROOM_SILENCE` sees "Auction: Hi"; `Deaf[AUCTION]` does not see; `ROOM_SILENCE` room does not see.

**Regression:**

14. Full `go test ./...` green.
15. Existing `DoTell` + AFK prefix tests still pass.

## Open questions / risks

- **Factor-vs-open-code:** I recommend open-coding. Three sites with divergent audience predicates don't merit a helper yet. If we later add 5+ more channels with uniform filtering, revisit.
- **`PLR_SILENCE` scope creep:** adding the check to existing comm commands is a behavior change beyond the P1 ask. Flag in commit message.
- **Gtell sleeper semantics:** C includes sleepers via `first_char`. Go matches automatically since `WorldRef.Characters` is sleeper-inclusive and `ch.Send` is a nil-Desc no-op. Don't special-case — just ensure no panic.
- **Auction stub vs full:** stubbing the command keeps `auction` from appearing broken in `commands` list without committing to the P3 system. If the user prefers omitting the command entirely until Phase 6, that's also defensible.
- **Alias registration / parsing:** command table uses string keys. `Interpret` extracts `cmdWord` via `util.OneArgument` which splits on whitespace. `": hello"` parses to cmdWord `":"` — works. `":hello"` (no space, a C prefix-match behavior) parses to cmdWord `":hello"` — **won't match**. For this port, require a space after `:`/`;` (document in commit); full prefix-matching is a separate concern in the interpreter. Add explicit `Interpret(ch, ": hi")` and `Interpret(ch, "; hi")` dispatch tests.
- **Color-code fidelity:** C's distinct `AT_IMMORT` / `AT_GTELL` / `AT_GOSSIP` colors aren't mapped to `&<code>` in Go's current translation layer. Match existing `DoClantalk` (`&Y`/`&G`/`&D`) for now; fidelity pass is a separate task.
- **Trust elevation:** no channel here invokes the interpreter, so the S3 trust-cap concern doesn't apply.

## Rough total effort

G1+G2+G3+G4 ≈ half a day with tests. G5 is a whole day minimum and is out of scope.

## Relevant file paths

- C: `/home/eilidh/src/smaug/src/act_comm.c` (392, 1327, 4217, 4293, 4309), `/home/eilidh/src/smaug/src/act_obj.c:3775+`, `/home/eilidh/src/smaug/src/update.c:2886-3286`, `/home/eilidh/src/smaug/db/system/en/commands.dat`
- Go new: `/home/eilidh/src/smaug/smaug-go/internal/handler/group.go` (or extension), `…/internal/act/auction.go`, corresponding `_test.go` siblings
- Go modified: `/home/eilidh/src/smaug/smaug-go/internal/act/comm.go` (add `DoImmtalk`, `DoGtell`; add `PLR_SILENCE` check to existing commands), `/home/eilidh/src/smaug/smaug-go/internal/boot/boot.go` (register five entries: `immtalk`, `:`, `gtell`, `;`, `auction`)
- Seams: `…/internal/types/enums.go:910-943` (channel bits), `…/internal/types/character.go:164,384,395` (`Deaf`, `GetTrust`, `Send`), `…/internal/types/misc.go:93` (BitVector), `…/internal/world/world.go:19,23` (Characters, Descriptors), `…/internal/testclient/login_test.go:118-134` (two-client test pattern)

---

## Completion record (2026-04-17)

All four in-scope groups landed. 15 acceptance criteria verified. `go test -count=3 ./...` green across all 15 packages.

### Delivered

- **G1 shared infra.** `handler.IsSameGroup(a, b *CharData) bool` in new `internal/handler/group.go` matching C `act_comm.c:4293-4301` exactly — normalizes each arg to its group leader (itself if `Leader == nil`) and compares by pointer identity. Nil-safe. A dedicated `TestIsSameGroup_TransitiveViaChain` pins the flat-graph semantic (A leads B, C follows B; IsSameGroup(A, C) is false because C normalizes to B, not to A — C doesn't recurse either). 6 tests total.
- **G1 opportunistic PLR_SILENCE.** Added sender-gate to `DoTell` (`"You can't do that."` per C `act_comm.c:1804`), `DoYell` / `DoShout` / `DoGossip` (`"You can't <verb>."` per C `act_comm.c:500-504`). Latent-bug fix documented in CHANGELOG and in the commit message: `PLR_SILENCE` was defined and set by `DoSilence` (`wiz.go:896-901`) but no communication command consulted it. NPC safety preserved — `ch.IsNPC()` short-circuits. 5 new tests (4 gate, 1 NPC complement).
- **G2 DoImmtalk.** In new `internal/act/channels.go`. Mortal → `"Huh?"`. Empty-arg → `"Immtalk what?"`. `PLR_SILENCE` sender gate → `"You can't immtalk."`. Deaf-sender block with the exact C diagnostic, and the deaf bit is NOT cleared (matches C — `xREMOVE_BIT` at `act_comm.c:514` is unreachable after the `:511` return). Self-echo and broadcast use `DoClantalk`-style `&Y`/`&G`/`&D` color. Broadcast filter walks `WorldRef.Descriptors` on `CON_PLAYING && Character != nil && Character != ch && Trust >= LEVEL_IMMORTAL && !Deaf[IMMTALK]`. 7 tests.
- **G3 DoGtell.** Same file. Empty-arg → `"Tell your group what?"`. `PLR_NO_TELL` sender gate → `"Your message didn't get through!"`. Iterates `WorldRef.Characters` and delivers to every `gch` where `handler.IsSameGroup(gch, ch)`. Safe against `Desc == nil` (char.Send is already nil-safe). The `QuickLoginTwo` helper added to `internal/testclient/login.go` is the first test utility in the suite to drive two concurrent clients — `TestGtell_TwoClients_Concurrent` in `internal/testclient/channels_test.go` uses it to form a group via `follow` and assert the follower receives the gtell line through the real network path. 6 unit tests + 2 alias-dispatch tests (`: hi`, `; hi`) + 1 integration test.
- **G4 auction helper + stub.** `internal/act/auction.go`. `BroadcastAuction(message string)` exported with the full C `talk_auction` filter (`CON_PLAYING && Trust >= 5 && !Deaf[AUCTION] && !InRoom.RoomFlags[ROOM_SILENCE]`). `DoAuction` emits `"The auction house is currently closed. (See the Phase-6 roadmap.)"` — no broadcast. NPC guard on `DoAuction`. `auctionMinTrust = 5` named constant for the trust threshold. 8 tests covering happy path + 4 filter predicates + nil-world safety + stub message + NPC guard.
- **Boot.** 5 new entries in `internal/boot/boot.go`: `immtalk` and `:` (both POS_DEAD, LEVEL_IMMORTAL, DoImmtalk); `gtell` and `;` (both POS_SLEEPING, Level 0, DoGtell); `auction` (POS_SLEEPING, Level 0, DoAuction).

### Acceptance criteria verdict

- **AC 1-6 (`immtalk`).** PASS. `TestDoImmtalk_MortalDenied` covers 1 and 3; `TestDoImmtalk_ImmortalToImmortal` covers 2; `TestInterpret_ImmtalkAlias_Colon` covers 4; `TestDoImmtalk_DeafSender_Blocked_NotCleared` covers 5 (both "blocked" and "not cleared" assertions); `TestDoImmtalk_PLRSilence_BlocksSender` covers 6.
- **AC 7-11 (`gtell`).** PASS. `TestDoGtell_Solo` covers 7; `TestDoGtell_DeliveresToGroupOnly` covers 8; `TestDoGtell_PLRNoTell_BlocksSender` covers 9; `TestDoGtell_NilDescNoPanic` covers 10; `TestInterpret_GtellAlias_Semicolon` covers 11. `TestDoGtell_NPCLeaderFollowerHears` adds a bonus NPC-leader-follower scenario.
- **AC 12-13 (`auction`).** PASS. `TestDoAuction_Stub_ClosedMessage` covers 12; `TestBroadcastAuction_{HappyPath,TrustGate,DeafFilter,RoomSilence,OnlyPlaying,NilWorldNoPanic}` covers 13 across six filter predicates.
- **AC 14 (regression).** PASS. `go test -count=3 ./...` green — cached package hits on the first pass, freshly-run on the mutation-verify iterations of each gate.
- **AC 15 (existing tests).** PASS. `TestDoTell` / `TestDoTell_NoArg` / `TestDoReply` / `TestDoGossip` / `TestDoEmote` / `TestDoYell` unchanged and still green. AFK prefix in `DoTell` untouched.

### Mutation verification summary

| Gate | Mutation | Expected fail | Result |
|---|---|---|---|
| `IsSameGroup` leader-normalization | Strip `if Leader != nil { ... }` branches | LeaderAndFollower + TwoFollowersOfSameLeader + TransitiveViaChain | FAIL as expected, revert restores green |
| `DoImmtalk` mortal block | `!ch.IsImmortal()` → `ch.IsImmortal()` | ImmortalToImmortal + PLRSilence + DeafSender all fail with `"Huh?"` | FAIL as expected |
| `DoImmtalk` receiver trust gate | `< LEVEL_IMMORTAL` → `< LEVEL_IMMORTAL+1` | MortalDoesNotReceive — mortal starts receiving immtalk | FAIL as expected |
| `DoGtell` same-group filter | `!IsSameGroup(...)` → `IsSameGroup(...)` | Solo + DeliveresToGroupOnly + NilDescNoPanic + NPCLeaderFollowerHears | FAIL as expected |
| `BroadcastAuction` trust gate | `< auctionMinTrust` → `< 0` | TrustGate — low-trust listener receives auction | FAIL as expected |

### Deliberate divergences from C

- **Prefix-within-token aliases.** `util.OneArgument` splits on whitespace, so `":hi"` parses to cmdWord `":hi"` and does not match the `:` registration. `": hi"` works. Plan-documented; fix is a separate interpreter change.
- **Gtell sleeper scope.** C's `first_char` retains loaded-but-disconnected characters until explicit extract. `WorldRef.Characters` drops chars on disconnect (`game/loop.go:875`). Net effect: disconnected "sleepers" don't receive group tells in Go, whereas C would enqueue to their descriptor. Non-blocking — disconnected players can't receive messages either way. Asleep-but-connected players still receive via their live `Desc`.
- **Deaf-not-cleared in DoImmtalk.** Matches C behavior exactly (the `xREMOVE_BIT` at `act_comm.c:514` is unreachable when the sender is deaf). Documented in both code comment and commit message. Requires a future `DoChannels` toggle command to re-enable — follow-up queued in TODO.md.

### Deferrals → Phase 6 (tracked in TODO.md)

- Full `do_auction` state machine (list / bid / stop / noauction / escrow / gold / tick).
- `music` / `newbiechat` / `racetalk` / `wartalk` / `counciltalk` / `guildtalk` channels. Revisit factor-out-of-`talk_channel` question once 5+ channels land.
- Per-AT_ color preservation (`AT_IMMORT` / `AT_GTELL` / `AT_GOSSIP`) — pairs with the existing `util.Act` audit finding.
- Alias prefix-matching (`":hi"` with no space) — requires interpreter change.
- `DoChannels` toggle command — required so a player who deafens `immtalk` has a way to re-enable it in the Go port.

### Files touched

- New: `internal/handler/group.go`, `internal/handler/group_test.go`, `internal/act/channels.go`, `internal/act/channels_test.go`, `internal/act/auction.go`, `internal/act/auction_test.go`, `internal/testclient/channels_test.go`.
- Modified: `internal/act/comm.go` (PLR_SILENCE gates in DoTell/DoYell/DoShout/DoGossip), `internal/act/comm_test.go` (+5 PLR_SILENCE tests), `internal/boot/boot.go` (+5 registrations), `internal/testclient/login.go` (+`QuickLoginTwo`), `smaug-go/doc/phases.md` (Tier 9 entry), `CHANGELOG.md`, `TODO.md`.
