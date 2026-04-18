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
