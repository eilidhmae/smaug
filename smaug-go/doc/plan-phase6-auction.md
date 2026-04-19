# Plan: Phase 6 — Full `do_auction` State Machine

**Status:** Planned (2026-04-18). `Agent`-tool unavailable in this manager environment — structured self-review substituted for external adversary pass. External adversary review recommended before dispatch.
**Priority:** Wave 2 of Phase 6 (`phase6-roadmap.md`). Completes the Tier-9 `BroadcastAuction` stub.
**Scope:** Extend `internal/types/misc.go` (`AuctionData` gains `History [AUCTION_MEM]*ObjIndexData`, `HistTimer int` — current struct has 7 of the 9 C fields). Rewrite `internal/act/auction.go` (`DoAuction` from stub to full state machine). New file `internal/act/auction_tick.go` or extension of `internal/game/update.go` (auction tick). New file `internal/util/parsebet.go` (port of C `parsebet`/`advatoi`). New file `internal/world/noauction.go` OR `internal/types/noauction.go` (no-auction list scaffolding; Phase-6 optional, see §Open Questions Q3). Touches `internal/game/loop.go` (register auction-pulse slot), `internal/boot/boot.go` (no new entries — `auction` already registered at line 412; immortal `auction stop` needs no extra registration). Touches `internal/game/update.go` or sibling for the pulse-driven tick wrapper. Substantial test additions in `internal/act/auction_test.go` (extend existing 8 tests to ~25).

**Scope explicitly narrowed:** this plan ports the **non-GSC** (`#else` branch, `src/act_obj.c:3775-4240`) variant of `do_auction` and the **non-GSC** (`src/update.c:3183-3326`) variant of `auction_update`. The `ENABLE_GOLD_SILVER_COPPER` branch (GSC) is out of scope for Phase 6 per the roadmap. The Go port has never modeled separate gold/silver/copper currency — every `gold` field is a single `int`.

---

## Problem

C ships the full auction system at two entry points:

1. **Command:** `src/act_obj.c:3775-4240` `do_auction(ch, argument)` — a 470-line state machine parsing one of four user intents: (a) `auction` with no args — show current-item info or "nothing auctioned", (b) `auction stop` (immortal only) — cancel + refund, (c) `auction bid <amount> [item-keyword]` — place a bid, (d) `auction <item> [min-bet]` — start auctioning a carried item.
2. **Tick:** `src/update.c:3183-3326` `auction_update()` — a 144-line routine called every `PULSE_AUCTION` (9 seconds) from `update.c:2746-2750`. Advances `auction->going` from 0 → 1 → 2 → 3, broadcasting "going once" / "going twice" / "SOLD!" or "returned to seller" at each stage. Rewards seller 90%, taxes the area 10% on a successful sale.

The Go port has shipped (Tier 9):

- `types.AuctionData` struct with **7 of 9 C fields** (`Item`, `Seller`, `Buyer`, `Bet`, `Going`, `Pulse`, `Starting` — missing `History [3]*ObjIndexData` and `HistTimer int`).
- `world.World.Auction *types.AuctionData` field (unassigned today — no loader initializes it; `DoAuction` stub never reads it).
- `act.BroadcastAuction(message string)` — the `talk_auction` helper with full filter chain.
- `act.DoAuction` as a stub: emits `"The auction house is currently closed. (See the Phase-6 roadmap.)"` and returns. Never reads `world.Auction`.
- `types.PULSE_AUCTION = 9 * PULSE_PER_SECOND` and `types.AUCTION_MEM = 3` constants — defined but unused.
- `types.CHANNEL_AUCTION` Deaf bit — honored by `BroadcastAuction`.
- `handler.GetObjCarry(ch, name)` — maps to C `get_obj_carry`.
- `handler.ObjFromChar` / `handler.ObjToChar` — map to C `obj_from_char` / `obj_to_char`.
- `mudprog.boostEconomy` — ports C `boost_economy` (per `internal/mudprog/commands.go:752`).
- Item-type flag helpers (`IsObjStat`, bit-test on `obj.ExtraFlags`) — adequate for the `ITEM_CLANOBJECT` / `ITEM_PERMANENT` / `ITEM_POISONED` gates.

**What's missing:**

- The state machine in `DoAuction` itself.
- The per-pulse tick in `update.go` / `loop.go`.
- `parsebet(currentbet, string) int` (C `src/bet.h:188-210`) — handles bare integers, `Nk`/`Nm` multipliers, `+N` percent additions, `*N` / `xN` multiplication. Not present anywhere in Go.
- `advatoi` (C `src/bet.h:118-163`) — helper for the non-GSC `parsebet`. Not present.
- `AuctionData.History` + `AuctionData.HistTimer` fields for the "auctioned-recently" guard.
- `NOAUCTION_DATA` (C `src/mud.h:778-783`) — the immortal-managed list of vnums that cannot be auctioned. **Nothing in the Go port loads or saves this list today.** Phase 6 decision in §Open Questions Q3.
- `separate_obj` (C — splits a stacked object; Go has no object-stacking so the call becomes a no-op, see §Go Design D4).
- `can_carry_w(ch) int` / `get_obj_weight(obj) int` — used by the tick to decide whether to give the sold item to buyer directly or drop it on the floor. **Go has no carry-weight check** (audit 2026-04-18 found no `CanCarryW` / `GetCarryW` / `WeightCap` helpers). See §Open Questions Q4.
- `auction->item->short_descr` broadcast formatting — straightforward, uses `obj.ShortDescr` (already a field).

### C anomalies catalogued during this plan's research

1. **Reachable — "not sold" branch `can_carry_w` check for SELLER at `update.c:3293-3308`.** When no-bid-received, the tick calls `obj_to_char(auction->item, auction->seller)` (or `obj_to_room` on overload). The seller *gave up* the item at auction-start (`obj_from_char(obj)` at `act_obj.c:4198`) and has up to ≤27 seconds (3 × PULSE_AUCTION) to walk around freely — including picking up new items, entering shops, buying gear. The returning item's weight CAN push them over capacity. Not dead code; port verbatim. (Audit 2026-04-18 corrected an earlier mischaracterization of this as defensive/unreachable.)
2. **Seller/buyer disconnect — C has a quit gate, Go does not.** C ships an active `do_quit` gate at `src/act_comm.c:2883-2890`: `if (auction->item != NULL && ((ch == auction->buyer) || (ch == auction->seller))) { send_to_char("Wait until you have bought/sold the item on auction.\n", ch); return; }`. The `mud.h:3614-3615` comments (`"may NOT quit"`) document this contract. **The Go port currently has NO such gate** at `internal/act/info.go:343-366` `DoQuit`. Quitting leaves a dangling `*CharData` pointer in `world.Auction`. This plan's G8 ports the C gate AND adds a Go-originated defensive `closeDescriptor` clear (for network-level disconnects — SIGKILL / socket reset — that bypass `do_quit` in both C and Go). Message wording ported from C: `"Wait until you have bought/sold the item on auction."`.
3. **Time-of-day gate at `act_obj.c:3801-3808`** — auctions only start 9 AM – 6 PM game time, except immortals bypass. Port verbatim. Uses `world.TimeInfo.Hour`.
4. **Level-3 minimum at `act_obj.c:3794-3799`** — auctioneers must be level ≥3. Port verbatim.
5. **Bid-too-low threshold at `act_obj.c:4024` is hardcoded `10000` gold.** Comment at `:4020-4022` says "changed to 10000 for our high economy." Port verbatim.
6. **Maximum bid at `act_obj.c:4038-4042` is `2000000000`** (2 billion — near INT32 max). Port verbatim.
7. **Self-bid guard at `act_obj.c:3996-4000`** — seller cannot bid on their own item. Port verbatim.
8. **Item-type whitelist at `act_obj.c:4158-4231`** — only specific item types can be auctioned (PAPER, LIGHT, TREASURE, POTION, KEYRING, QUIVER, DRINK_CON, FOOD, COOK, PEN, BOAT, PILL, PIPE, HERB_CON, INCENSE, FIRE, RUNEPOUCH, MAP, BOOK, RUNE, MATCH, HERB, WEAPON, MISSILE_WEAPON, ARMOR, STAFF, WAND, SCROLL). Everything else falls to `default: "You cannot auction Xs."` at `:4161-4164`. Port verbatim (list is load-bearing, not gameplay-invented).
9. **`ch->level < auction->item->level` bid-block at `act_obj.c:3989-3994`** — bidders can't bid on items above their own level. Port verbatim.
10. **Timer-decay block at `act_obj.c:4107-4111`** — objects with `obj.Timer > 0` (decaying) cannot be auctioned. Port verbatim. Go already has `ObjData.Timer`.
11. **Modified-item block at `act_obj.c:4099-4105`** — `obj->item_type != obj->pIndexData->item_type` ⇒ "too modified to auction". Go has `obj.ItemType` and `obj.IndexData.ItemType` — port verbatim.
12. **History-queue collision at `act_obj.c:4125-4134`** — refuse to auction an item whose *prototype vnum* appears in the history ring. Prevents farm-bot spam. Port verbatim.
13. **ClanObject / Permanent gates at `act_obj.c:4113-4123`** — `ITEM_CLANOBJECT` + `ITEM_PERMANENT` forbid auctioning. Port verbatim.
14. **`ms_find_obj(ch)` call at `act_obj.c:4079-4080`** — `ms_find_obj` is defined at `handler.c:2935-2994` as a **generalized drunk/mental-state check** called from get/drop/put/quaff/recite/eat/drink/auction/pick. It applies to ALL characters and probabilistically makes the object operation fail when mental_state + drunk are high. Go has no equivalent and has not wired mental-state side-effects into any object command. Port as no-op; consistent with Go's existing omission. Flag in §Open Questions Q5. (Audit 2026-04-18 corrected earlier mischaracterization as a "mud-school quest hook.")
15. **Tax-gold-floor bug at `update.c:3317-3320`** — if the seller's gold minus tax would go negative, the seller's gold is zeroed. This is a "no-negative-balances" safety. Port verbatim but note: the "not sold" tax is `item->cost * 0.05` — a tax on the un-sold item's list price. The "sold" tax (`update.c:3266`) is `bet * 0.10` — a tax on the winning bid. Two different formulae; port both.
16. **`num_punct` dependency** — C uses `num_punct()` to format gold values as `"1,234,567"`. Go has no equivalent. Port as a new helper `util.NumPunct(int) string` OR use standard Go formatting — decision captured in §Go Design D5.
17. **WAIT_STATE on repeat-auction-attempt at `act_obj.c:4236-4238`** — if a mortal tries to auction a second item while one is already running, the C code applies `WAIT_STATE(ch, PULSE_VIOLENCE)`. Go has `types.CharData.Wait` (`int`) and `internal/types/character.go` exposes it; check + verify it's the same units as C. Port verbatim.

None of the above are gameplay-invented. Every item cites a C line.

---

## C Reference (authoritative)

All citations are against `src/act_obj.c` and `src/update.c` on HEAD. Line numbers verified 2026-04-18.

### Entry points and pulse

- `do_auction` (non-GSC) — `src/act_obj.c:3775-4240`.
- `auction_update` (non-GSC) — `src/update.c:3183-3326`.
- Tick call site — `src/update.c:2746-2750`:
  ```c
  if (--auction->pulse <= 0)
    {
      auction->pulse = PULSE_AUCTION;
      auction_update ();
    }
  ```
  Invariant: `auction->pulse` counts DOWN every master-pulse; when it hits zero, tick runs and `pulse` resets to `PULSE_AUCTION = 9 * PULSE_PER_SECOND = 36` (9 seconds). Tick runs IFF `auction->item != NULL` (the first line of `auction_update` returns early on NULL except for history decay).
- Constants: `src/mud.h:408` `PULSE_AUCTION = 9 * PULSE_PER_SECOND`; `src/mud.h:3609` `AUCTION_MEM = 3`.
- Struct: `src/mud.h:3611-3622` `struct auction_data` — 9 fields including `history[AUCTION_MEM]` and `hist_timer`.
- `NOAUCTION_DATA` — `src/mud.h:778-783` — linked list of vnums.
- `parsebet` — `src/bet.h:188-210`. `advatoi` — `src/bet.h:118-163`.

### `do_auction` flow summary

1. **Arg parse** — three one_argument calls (L3785-3787).
2. **NPC guard** — return silently (L3791).
3. **Level gate** — `< 3` refuse (L3794-3798).
4. **Time gate** — hour < 9 OR hour > 18, no current item, non-immortal → refuse (L3801-3808).
5. **Empty `arg1`** — show info block (L3810-3955):
   - If `auction->item`: show bid, full spell_identify-style breakdown, contents for containers, extra-affects, and for immortals: seller+buyer names, round, time-left.
   - If no item: `"There is nothing being auctioned right now…"`.
6. **`stop` (immortal)** — cancel, return item to seller, refund bidder, NULL item (L3957-3982).
7. **`bid`** — parse bet, validate (item exists, bidder ≤ item level, not self-seller, not less than starting, not less than current + 10k, not more than player's gold, not >2B, item keyword matches if arg3), refund previous bidder's gold, debit new bidder's gold, set buyer=ch, bet=newbet, reset `going=0` and `pulse=PULSE_AUCTION`, broadcast (L3983-4077).
8. **Start-new-auction** — `get_obj_carry`, noauc check, unmodified check, timer check, clanobject check, permanent check, history check, parse min-bet arg2, item-type switch (whitelist — all valid types do the same: `separate_obj`, `obj_from_char`, set auction state, push to history ring, broadcast) (L4078-4240).

### `auction_update` flow summary

1. **No item** — early return except for history aging: every `6 * AUCTION_MEM = 18` ticks, pop oldest non-null entry and reset timer (L3188-3206).
2. **Increment `going`:**
   - `case 1` / `case 2`: broadcast "going once/twice" (L3210-3223).
   - `case 3`: resolve sale (L3224-3324).
3. **Sold branch (L3231-3281):** broadcast "sold to X for Y", spawn-to-buyer via `AT_ACTION` act calls ("The auctioneer materializes before you…"), check `can_carry_w` — over-cap drops to room, else to inventory. Pay seller 90%, tax area 10%, ch_printf the gold breakdown to seller.
4. **Not-sold branch (L3282-3323):** broadcast "No bids received", spawn-to-seller act messages, over-cap drops to room (this is the dead-code defense — see anomaly #1), tax is `item->cost * 0.05`, floor at zero.
5. **Clear item** — `auction->item = NULL` (L3324).

---

## Go Current State

Verified 2026-04-18 against the `golang` branch at commit `26034c1` (working tree).

### Fields

- `types.AuctionData` — `internal/types/misc.go:93-102`, 7 fields:
  ```go
  type AuctionData struct {
      Item     *ObjData
      Seller   *CharData
      Buyer    *CharData
      Bet      int
      Going    int
      Pulse    int
      Starting int
  }
  ```
  **Missing:** `History [AUCTION_MEM]*ObjIndexData`, `HistTimer int`.
- `world.World.Auction *types.AuctionData` — `internal/world/world.go:60` — struct field present, never initialized. `world.New()` does NOT assign it.

### Constants

- `PULSE_AUCTION = 9 * PULSE_PER_SECOND` — `internal/types/constants.go:71`.
- `AUCTION_MEM = 3` — `internal/types/constants.go:776`.
- `CHANNEL_AUCTION` (Deaf bit) — `internal/types/enums.go:911`.
- `ROOM_SILENCE` — `internal/types/enums.go:711`.
- `ITEM_CLANOBJECT` — `internal/types/enums.go:669`.
- `ITEM_PERMANENT` — `internal/types/enums.go:682`.
- Item-type whitelist constants (ITEM_PAPER, LIGHT, TREASURE, POTION, KEYRING, QUIVER, DRINK_CON, FOOD, COOK, PEN, BOAT, PILL, PIPE, HERB_CON, INCENSE, FIRE, RUNEPOUCH, MAP, BOOK, RUNE, MATCH, HERB, WEAPON, MISSILE_WEAPON, ARMOR, STAFF, WAND, SCROLL) — confirm during G1 exact lines; all expected in `enums.go` given Phase 5 coverage.

### Commands

- `act.DoAuction` — existing stub at `internal/act/auction.go:51-60`. To be replaced.
- `act.BroadcastAuction` — existing helper at `internal/act/auction.go:25-45`. Kept verbatim; called from both the command and the tick.
- Boot registration — `internal/boot/boot.go:412`:
  ```go
  reg.Register(&command.Command{Name: "auction", DoFun: act.DoAuction, Position: types.POS_SLEEPING, Level: 0})
  ```
  **Level 0** in the Go port. C registers as level 3 via `commands.dat`; the registration is in-file here. Confirm trust gate is handled inside `DoAuction` (matches C pattern) — yes, C checks `ch->level < 3` inside the function body.

### Pulse infrastructure

- Game loop at `internal/game/loop.go:139-194` has four pulse slots: Area (via `pulseArea`), Violence (`pulseViolence`), Mobile (`pulseMobile`), Tick (`pulseTick`), plus Save (`pulseSave`). **No auction slot.**
- Pulse slot constants and counters are initialized at `loop.go:87-94`.
- `PULSE_SAVE` existence implies the pattern is extensible.
- Adding `pulseAuction` is mechanical: add the field to the `GameLoop` struct, initialize in `NewGameLoop`, decrement + reset + dispatch in `pulse()`.

### Disconnect / quit gate

- `internal/game/loop.go:866-900` `closeDescriptor` — removes char from room, removes from `world.Characters`, closes conn. **Does NOT clear `world.Auction.Seller` or `.Buyer` pointers.** A disconnect-while-auctioning leaves a dangling pointer.
- `internal/act/info.go:343+` `DoQuit` — has TIMER_RECENTFIGHT gate, no auction gate.

### Helpers

- `handler.GetObjCarry(ch, name) *ObjData` — `internal/handler/find.go:88`.
- `handler.ObjFromChar(obj)` — `internal/handler/handler.go:215`.
- `handler.ObjToChar(obj, ch)` — `internal/handler/handler.go:195`.
- `handler.ExtractObj(w, obj)` — `internal/handler/handler.go:~305`.
- `world.World.GetObjIndex(vnum)` — for history ring comparisons.
- No `CanCarryW`, `GetCarryW`, `GetObjWeight`, `WeightCap` — see §Open Questions Q4.
- No `SeparateObj` — Go does not stack objects, so `separate_obj` becomes a no-op (see §Go Design D4).
- `mudprog.boostEconomy(w, area, amount)` — `internal/mudprog/commands.go:752-800`.
- No `parsebet` / `advatoi` — to be ported.
- No `NumPunct` — see §Go Design D5.
- `util.OneArgument(s) (first, rest)` — lowercases first word.

### Tests

- `internal/act/auction_test.go` — 8 existing tests covering `BroadcastAuction` (6) and `DoAuction` stub (2).

---

## Go Design

### D1 — Auction ownership semantics

**Chosen: `world.World.Auction *types.AuctionData` is canonical; initialize via `world.New()` with an empty struct (not a nil pointer).**

Rationale: the struct field is already present (`world.go:60`); initializing it at creation time removes nil-guard boilerplate from the command and the tick. Single allocation at boot. Matches C's global `auction` pointer which is malloc'd once in `db.c:587` and reused.

Change: `world.New(dataDir)` at `internal/world/world.go:80` — add `Auction: &types.AuctionData{}` to the struct literal. Persistence of the auction state across reboots is **out of scope** (C does not persist it either — `auction->item` is always NULL at boot; C `db.c:587-600` allocates a zero-filled struct). Hotboot will need to decide per §Open Questions Q6.

### D2 — `AuctionData` struct extension

Add two fields:
```go
type AuctionData struct {
    Item      *ObjData
    Seller    *CharData
    Buyer     *CharData
    Bet       int
    Going     int
    Pulse     int
    Starting  int
    History   [AUCTION_MEM]*ObjIndexData  // NEW — prototype ring
    HistTimer int                         // NEW — idle-decay counter
}
```

`AUCTION_MEM` is already defined at `constants.go:776`. The `[3]*ObjIndexData` array is fixed-size to match C's `OBJ_INDEX_DATA *history[AUCTION_MEM]`. Zero-value works (`[3]*ObjIndexData{}` is three nils).

### D3 — Auction tick slot

**Chosen: add `pulseAuction` as a sibling to `pulseViolence` / `pulseMobile` / `pulseTick` in `GameLoop`, drive a new `auctionUpdate()` method from the main pulse loop.**

Changes:

- `internal/game/loop.go`:
  - Add field `pulseAuction int` to `GameLoop` struct (exists next to `pulseSave`).
  - In `NewGameLoop`, initialize `pulseAuction: types.PULSE_AUCTION`.
  - In `pulse()`, add the block (inserted after `pulseTick` handling, before `pulseSave`):
    ```go
    g.pulseAuction--
    if g.pulseAuction <= 0 {
        g.pulseAuction = types.PULSE_AUCTION
        g.auctionUpdate()
    }
    ```
- `internal/game/update.go`: add `auctionUpdate()` method (direct port of C `auction_update`). Keeps the package coherent (every pulse-driven update lives in `update.go`).

**Alternative rejected: `act.AuctionTick(world)` called from the existing `pulseTick` block.** C does not piggyback on `PULSE_TICK` (70 seconds); it uses a dedicated counter. The Go port already has the extensibility pattern for new pulse slots (`PULSE_SAVE` is additive). Using `PULSE_TICK` would reduce auction-resolution granularity from 9 seconds to 70 seconds — visibly degraded.

**Alternative rejected: run `auctionUpdate` directly inside `pulse()` without a counter** (i.e., decrement `world.Auction.Pulse` every pulse). Would work and matches C more literally (C decrements `auction->pulse` every master-pulse). Rejected because it mixes state-machine state (`Auction.Pulse`) with scheduler state — the Go pattern is that each subsystem's pulse interval lives on `GameLoop`, not on domain data. Keep subsystem counters in one place.

### D4 — `separate_obj` port as no-op helper

C `separate_obj` splits a stacked object: if an item has `count > 1`, it's decremented and a fresh duplicate is spawned with `count = 1`. **The Go port does not stack objects** — every `ObjData` instance is count-1 by construction (confirmed: no `Count` field on `ObjData` beyond `obj.IndexData.Count` which is the *prototype* reference count, not a per-instance stack count).

So `separate_obj(obj)` becomes a **no-op comment** in the port:
```go
// C calls separate_obj(obj) here; Go has no object stacking (every
// ObjData is count-1), so the call is a no-op.
```

No helper, no function — just the comment. Document this divergence explicitly in the task group.

### D5 — `num_punct` port

**Chosen: new `util.NumPunct(n int) string` helper returning the comma-formatted representation.**

Rationale: used in at least 5 distinct broadcast messages and the seller's "you received N gold" printout. Centralizing avoids copy-paste. Go stdlib has no built-in thousands-separator formatting; `fmt.Sprintf("%d", n)` emits raw digits.

Implementation (≤15 LOC):
```go
// NumPunct formats an integer with commas as thousands separators.
// Mirrors C num_punct() in comm.c. Negative numbers are handled.
func NumPunct(n int) string {
    if n < 0 {
        return "-" + NumPunct(-n)
    }
    s := strconv.Itoa(n)
    if len(s) <= 3 {
        return s
    }
    // Insert commas from the right every 3 digits.
    var b strings.Builder
    first := len(s) % 3
    if first > 0 {
        b.WriteString(s[:first])
        if len(s) > first {
            b.WriteByte(',')
        }
    }
    for i := first; i < len(s); i += 3 {
        b.WriteString(s[i : i+3])
        if i+3 < len(s) {
            b.WriteByte(',')
        }
    }
    return b.String()
}
```

**Alternative rejected: `golang.org/x/text/message` package.** Adds a module dependency for one formatter. Overkill.

### D6 — `parsebet` / `advatoi` port

**Chosen: new file `internal/util/parsebet.go` with two exported functions:**

```go
// ParseBet mirrors C parsebet (src/bet.h:188-210).
func ParseBet(currentBet int, s string) int { ... }

// Advatoi mirrors the non-GSC C advatoi (src/bet.h:118-163): parses
// numbers with optional 'k' (×1000) or 'm' (×1000000) multipliers,
// plus trailing digits that scale down with the multiplier.
func Advatoi(s string) int { ... }
```

`ParseBet` handles four cases: bare-digit (delegate to `Advatoi`), `+N%` (add-percent), `*N` / `xN` (multiply), empty/unknown (0). Matches C semantics including the "percent default 25", "multiply default 2" quirks.

Tests: table-driven, ≥20 cases covering the bare/+/*/x/Nk/Nm/Nk4/Nm5 permutations per C bet.h docstring at `:60-77`.

### D7 — `NOAUCTION_DATA` list scaffolding

See §Open Questions Q3. Recommended resolution: **stub the no-auction list as an empty slice on `world.World` and defer the admin CRUD command.** C ships `do_noauction` at `act_wiz.c:11120-11174` (immortal command: list current entries, toggle vnum in/out, calls `save_noauctions` after mutation — the file IS runtime-mutable in C). The Go port's Phase-6 cut is load-only (the file rarely changes in practice); `do_noauction` is tracked as a follow-up (see `TODO.md`). The loader reads `system/noauction.dat` at boot:

- Add `World.NoAuction []int` (slice of vnums).
- Loader in `internal/persist/subsystems.go` reads `system/noauction.dat` if present (one vnum per line), else empty.
- `DoAuction` new-item path filters against this slice.
- No admin command in this plan.

### D8 — Disconnect / quit gate

**Chosen: port C's existing `DoQuit` gate (`act_comm.c:2883-2890`) to Go; additionally extend `closeDescriptor` to defensively clear any dangling `Auction.Seller` / `.Buyer` that match.**

C has an active quit gate at `act_comm.c:2883-2890`. Porting it is parity, not a Go invention. The additional `closeDescriptor` clear IS Go-originated and handles socket-level disconnects (peer reset / process killed) that bypass `do_quit` in both C and Go. Two-layer defense:

1. **`DoQuit` refusal** (`internal/act/info.go`, ported from `act_comm.c:2883-2890`): if `WorldRef.Auction != nil && WorldRef.Auction.Item != nil && (ch == WorldRef.Auction.Seller || ch == WorldRef.Auction.Buyer)` → `"Wait until you have bought/sold the item on auction.\n\r"` (verbatim from C at `act_comm.c:2887`) and return.
2. **`closeDescriptor` defensive clear** (`internal/game/loop.go:867`): after `g.SavePlayer(ch)` and before `g.world.RemoveChar(ch)`:
   ```go
   if g.world.Auction != nil && g.world.Auction.Item != nil {
       // Seller dropping: cancel auction, drop item to ground.
       if g.world.Auction.Seller == ch {
           // Broadcast cancellation, return item to ch's last room or drop.
           act.BroadcastAuction("The auction is cancelled — the seller has left.")
           if ch.InRoom != nil {
               handler.ObjToRoom(g.world.Auction.Item, ch.InRoom)
           }
           // If bidder exists and != seller, refund their gold.
           if g.world.Auction.Buyer != nil && g.world.Auction.Buyer != ch {
               g.world.Auction.Buyer.Gold += g.world.Auction.Bet
           }
           *g.world.Auction = types.AuctionData{History: g.world.Auction.History, HistTimer: g.world.Auction.HistTimer}
       } else if g.world.Auction.Buyer == ch {
           // Bidder dropping: refund not possible (char leaving); reset buyer to seller.
           g.world.Auction.Buyer = g.world.Auction.Seller
           g.world.Auction.Bet = g.world.Auction.Starting
           g.world.Auction.Going = 0
           g.world.Auction.Pulse = types.PULSE_AUCTION
           act.BroadcastAuction("The top bidder has disconnected. Bidding re-opens.")
       }
   }
   ```

History ring is preserved across cancellations (item-vnum spam guard must persist).

### D9 — `DoAuction` state machine layout

New file replacing stub at `internal/act/auction.go` (or extend — stub removal in-place is fine). Decomposed into private helpers:

```go
func DoAuction(ch *types.CharData, argument string) {
    // Arg parse, NPC, level, time gates.
    // Dispatch:
    arg1, rest := util.OneArgument(argument)
    arg2, rest := util.OneArgument(rest)
    arg3, _   := util.OneArgument(rest)
    if arg1 == ""            { auctionInfo(ch); return }
    if arg1 == "stop" && ch.IsImmortal() { auctionStop(ch); return }
    if arg1 == "bid"         { auctionBid(ch, arg2, arg3); return }
    auctionStart(ch, arg1, arg2)
}
```

Each helper ≤80 LOC, maps 1:1 to a C branch with inline `// src/act_obj.c:NNNN` line comments.

### D10 — Test strategy

Unit tests go in `internal/act/auction_test.go` (extend existing). Separate `auction_tick_test.go` for the tick. Integration test in `internal/testclient/` optional — the state machine is deterministic without the pulse, so unit coverage dominates.

Test harness pattern: directly manipulate `WorldRef.Auction` in each test's Arrange phase, invoke `DoAuction` / `auctionUpdate` in Act, assert on `WorldRef.Auction.*` and broadcast output in Assert. `BroadcastAuction` already has capture helpers.

---

## Task Groups

All task groups follow TDD: write failing tests first; implement until green; mutation-verify via `Edit`-only revert.

### G0 — Helper ports (`parsebet`, `advatoi`, `NumPunct`) (S, ~1.5h)

- **Test first** (`internal/util/parsebet_test.go`):
  - `TestAdvatoi_BareInt`: `"123"` → 123; `""` → 0; `"abc"` → 0.
  - `TestAdvatoi_KSuffix`: `"14k"` → 14000; `"14k42"` → 14420; `"14k1234"` → 14123 (truncates).
  - `TestAdvatoi_MSuffix`: `"23m"` → 23000000; `"23m5"` → 23500000.
  - `TestAdvatoi_RejectBadChar`: `"14q"` → 0.
  - `TestParseBet_BareDigit`: `ParseBet(1000, "500")` → 500.
  - `TestParseBet_PercentDefault`: `ParseBet(1000, "+")` → 1250.
  - `TestParseBet_PercentN`: `ParseBet(1000, "+50")` → 1500.
  - `TestParseBet_MultiplyDefault`: `ParseBet(1000, "*")` → 2000; `ParseBet(1000, "x")` → 2000.
  - `TestParseBet_MultiplyN`: `ParseBet(1000, "x10")` → 10000.
  - `TestParseBet_Empty`: `ParseBet(1000, "")` → 0.
- **Test first** (`internal/util/numpunct_test.go`):
  - `TestNumPunct_Small`: `0` → `"0"`; `123` → `"123"`.
  - `TestNumPunct_Thousand`: `1234` → `"1,234"`; `1000000` → `"1,000,000"`.
  - `TestNumPunct_Negative`: `-1234` → `"-1,234"`.
  - `TestNumPunct_Max`: `2000000000` → `"2,000,000,000"`.
- **Implement** in `internal/util/parsebet.go` and `internal/util/numpunct.go`.
- **Mutation verify:** In `NumPunct`, change `first := len(s) % 3` to `first := 0` — tests fail; `Edit`-revert.

### G1 — `AuctionData` struct extension (S, ~30min)

- **Test first** (`internal/types/misc_test.go` or new):
  - `TestAuctionData_ZeroHistoryRing`: construct an `AuctionData{}`, assert `History[0] == nil && History[1] == nil && History[2] == nil`, `HistTimer == 0`.
  - `TestAuctionData_FixedSize`: `len(AuctionData{}.History)` equals `AUCTION_MEM` (3).
- **Implement** the two-field extension per §D2.
- **Mutation verify:** temporarily change `[AUCTION_MEM]` to `[5]` — `TestAuctionData_FixedSize` fails; revert.

### G2 — `world.New` auction initialization + no-auction list scaffold (S, ~45min)

- **Test first** (`internal/world/world_test.go`):
  - `TestNew_AuctionInitialized`: `w := world.New("/tmp")`; `if w.Auction == nil { t.Fatal(...) }`.
  - `TestNew_AuctionEmpty`: assert `w.Auction.Item == nil && w.Auction.Bet == 0 && w.Auction.Going == 0`.
  - `TestNew_NoAuctionEmpty`: assert `len(w.NoAuction) == 0`.
- **Implement:**
  - Add `NoAuction []int` field to `World` at `internal/world/world.go:~51`.
  - In `world.New`, set `Auction: &types.AuctionData{}`.
- **Mutation verify:** remove the `Auction:` assignment → `TestNew_AuctionInitialized` fails; revert.

### G3 — `auctionUpdate` tick (M, ~3h)

- **Test first** (`internal/game/update_auction_test.go` new):
  - `TestAuctionUpdate_NoItem_NoBroadcast`: empty auction, tick runs, no broadcast fired, no crash.
  - `TestAuctionUpdate_NoItem_HistoryDecay`: set `History[0] = &ObjIndexData{}`, run 18 ticks, assert `History[0] == nil` and `HistTimer == 0`.
  - `TestAuctionUpdate_Going1`: `Item` set, `Bet > Starting`, tick → `Going == 1`, broadcast contains `"going once for"`.
  - `TestAuctionUpdate_Going2`: `Going=1`, tick → `Going == 2`, broadcast contains `"going twice"`.
  - `TestAuctionUpdate_Sold_HappyPath`: `Going=2`, `Buyer != Seller`, tick → buyer receives item, seller gets 90% bet, area gets 10% tax via `boostEconomy`, `Auction.Item == nil`.
  - `TestAuctionUpdate_NotSold`: `Going=2`, `Buyer == Seller` (no bids above starting), tick → seller reclaims item, seller gets charged 5% of item.Cost as tax, tax floored at 0.
  - `TestAuctionUpdate_Sold_NilBuyer_Recovers`: pathological `Going=2`, `Buyer == nil`, `Bet > 0` — defensive: bug-log emitted, `Bet` zeroed, not-sold branch taken.
  - `TestAuctionUpdate_Sold_SellerNilRoom_NoPanic`: seller's `InRoom` nil — tax `boostEconomy` call still no-panics.
- **Implement:**
  - Add `pulseAuction int` field to `GameLoop`, initialize in `NewGameLoop`, decrement + dispatch in `pulse()`.
  - Add method `(g *GameLoop) auctionUpdate()` in `internal/game/update.go`, mirroring C L3183-3326 branch-for-branch.
  - History aging mirrors `update.c:3190-3204`: if any slot is set and `++HistTimer == 6 * AUCTION_MEM`, pop oldest (highest index) and reset timer.
  - Uses `BroadcastAuction` for channel broadcasts, `util.Act` (AT_ACTION) for room narration — AT_ACTION is already defined in `constants.go` per Tranche C.
- **Mutation verify:** flip the `Bet > Starting` check to `Bet >= Starting` → `TestAuctionUpdate_Going1` message format differs; revert.

### G4 — `DoAuction` info branch (empty arg + no current item) (S, ~1h)

- **Test first** (extend `internal/act/auction_test.go`):
  - `TestDoAuction_Empty_NoItem`: no current auction, `DoAuction(ch, "")` → output contains `"There is nothing being auctioned right now"`.
  - `TestDoAuction_Empty_WithItem_NoBids`: item set, `Bet == 0`, `DoAuction(ch, "")` → output contains `"No bids on this item have been received"` and the item's short descr.
  - `TestDoAuction_Empty_WithItem_WithBids`: item set, `Bet == 500`, → output contains `"Current bid on this item is 500 gold"`.
  - `TestDoAuction_Empty_Immortal_SeesSellerBuyer`: item set, immortal caller → output contains `"Seller:"` and `"Bidder:"` and `"Round:"`.
  - `TestDoAuction_Empty_Mortal_NoSellerBuyer`: item set, mortal caller → output does NOT contain `"Seller:"`.
  - `TestDoAuction_Empty_NPC_Silent`: NPC caller → no output (C returns silently after NPC check).
  - `TestDoAuction_Empty_LevelTooLow`: level-2 caller → `"at least level three"`.
  - `TestDoAuction_Empty_NightHours_MortalBlocked`: hour=20, no current item, mortal → `"auctioneer works between"`.
  - `TestDoAuction_Empty_NightHours_ItemExists_Allowed`: hour=20, current item exists → info shown (C bypasses time-gate if item set).
  - `TestDoAuction_Empty_NightHours_ImmortalBypass`: hour=20, immortal → shown.
- **Implement** the info branch of `DoAuction` (L3810-3955).
  - Simplification: DO NOT port the per-item-type full spell_identify breakdown (container capacity phrases, scroll-spell listings, weapon damage, etc.) in this group — defer to G4b. Emit the bid line + item short_descr + weight/value/level.
- **Mutation verify:** Flip `hour > 18 || hour < 9` to `hour > 18 && hour < 9` → `TestDoAuction_Empty_NightHours_*` fails; revert.

### G4b — Per-item-type info expansions (S, ~1h)

- **Test first:** one test per item-type case from `act_obj.c:3842-3920`: container/keyring/quiver (capacity phrase), pill/scroll/potion (spell levels + names), wand/staff (charges), weapon (damage range), armor (AC).
- **Implement** the switch on `obj.ItemType` per C L3842-3920.
- **Mutation verify:** Swap the capacity-thresholds (`< 76` → `< 75`) — breaks "have a small capacity" assertion; revert.

### G5 — `DoAuction` bid branch (M, ~2h)

- **Test first:**
  - `TestDoAuction_Bid_NoItem`: `DoAuction(ch, "bid 500")` with no current auction → `"isn't anything being auctioned"`.
  - `TestDoAuction_Bid_ItemLevelTooHigh`: bidder level 5, item.Level 50 → `"level is too high"`.
  - `TestDoAuction_Bid_SelfBid`: seller bids on own item → `"can't bid on your own item"`.
  - `TestDoAuction_Bid_NoAmount`: `"bid"` with no arg2 → `"Bid how much?"`.
  - `TestDoAuction_Bid_BelowStarting`: starting=1000, bid 500 → `"higher than the starting bet"`.
  - `TestDoAuction_Bid_BelowCurrentPlus10k`: current bet=1000, bid 5000 → `"at least bid 10000 coins over"`.
  - `TestDoAuction_Bid_ExceedsGold`: ch.Gold=100, bid 20000 → `"don't have that much money"`.
  - `TestDoAuction_Bid_ExceedsMax`: bid 2500000000 → `"can't bid over 2 billion"`.
  - `TestDoAuction_Bid_WrongKeyword`: bid includes arg3="shield" but auction item is a sword → `"not being auctioned right now"`.
  - `TestDoAuction_Bid_Success`: valid bid — prev bidder refunded, ch.Gold debited, `Auction.Buyer == ch`, `Auction.Bet == newbet`, `Auction.Going == 0`, `Auction.Pulse == PULSE_AUCTION`, broadcast emits `"A bid of <N> gold has been received on <item>"`.
  - `TestDoAuction_Bid_Success_SelfPreviousBidder_NoDoubleRefund`: scenario where prev buyer is seller — no refund issued (C gate `auction->buyer != auction->seller`).
  - `TestDoAuction_Bid_ParsesKMultiplier`: bid `"10k"` → 10000.
  - `TestDoAuction_Bid_ParsesRelativePercent`: current bet 10000, bid `"+50"` → 15000 — but blocks at "+10000 over" — so the test asserts on the block, not on acceptance.
- **Implement** the bid branch of `DoAuction` (L3983-4077).
- **Mutation verify:** Flip `newbet < auction->bet + 10000` to `newbet < auction->bet + 1000` — `TestDoAuction_Bid_BelowCurrentPlus10k` fails; revert.

### G6 — `DoAuction` start-new branch (M, ~2.5h)

- **Test first:**
  - `TestDoAuction_Start_NotCarried`: arg1 matches no carried item → `"aren't carrying that"`.
  - `TestDoAuction_Start_InNoAuctionList`: item vnum is in `World.NoAuction` → `"cannot be auctioned"`.
  - `TestDoAuction_Start_ImmortalBypassesNoauc`: immortal auctions a NoAuction vnum → succeeds.
  - `TestDoAuction_Start_TypeMismatch`: `obj.ItemType != obj.IndexData.ItemType` → `"too modified to auction"`.
  - `TestDoAuction_Start_Decaying`: obj.Timer > 0 → `"can't auction objects that are decaying"`.
  - `TestDoAuction_Start_ClanObject`: ITEM_CLANOBJECT flag → `"can't auction clan items"`.
  - `TestDoAuction_Start_Permanent`: ITEM_PERMANENT flag → `"cannot leave your possession"`.
  - `TestDoAuction_Start_HistoryCollision`: obj's IndexData is in `Auction.History` → `"been auctioned recently"`.
  - `TestDoAuction_Start_InvalidMinBet`: arg2 is non-numeric → `"must input a number"`.
  - `TestDoAuction_Start_NegativeMinBet`: arg2 = "-10" → `"less than 0 gold"`.
  - `TestDoAuction_Start_DisallowedItemType`: attempt to auction an ITEM_KEY → `"You cannot auction Xs"` (using act AT_TELL pattern).
  - `TestDoAuction_Start_HappyPath_Weapon`: valid weapon, min-bet 1000 → `Auction.Item == obj`, `Auction.Seller == ch`, `Auction.Buyer == ch`, `Auction.Bet == 0`, `Auction.Starting == 1000`, `Auction.Going == 0`, `Auction.Pulse == PULSE_AUCTION`, `Auction.History[0] == obj.IndexData`, obj removed from `ch.Carrying`, broadcast emits `"A new item is being auctioned"`.
  - `TestDoAuction_Start_StartingBetPreloadsBet`: min-bet > 0 → `Auction.Bet == Starting`.
  - `TestDoAuction_Start_WhileInProgress_MortalWaitState`: auction already has an item, mortal attempts new → act-tell `"Try again later"` + `ch.Wait` set to `PULSE_VIOLENCE`.
  - `TestDoAuction_Start_WhileInProgress_ImmortalNoWaitState`: immortal attempts → same message, no wait-state.
  - `TestDoAuction_Start_HistoryRotation`: history had `[A, B, nil]`; start auction of C → `[C, A, B]`.
- **Implement** the start-new branch (L4078-4240), including the item-type whitelist switch.
- **Mutation verify:** Swap `memmove` semantics — flip slice of copy to keep `[A, B]` → `TestDoAuction_Start_HistoryRotation` fails; revert.

### G7 — `DoAuction stop` (immortal) branch (S, ~45min)

- **Test first:**
  - `TestDoAuction_Stop_Mortal_FallsThrough`: mortal types `"stop"` — C falls through to the start-new branch since the `stop` case is guarded by `IS_IMMORTAL`. Expected: `"aren't carrying that"` (no item named "stop" in inventory).
  - `TestDoAuction_Stop_Immortal_NoActiveAuction`: immortal, no item → `"no auction to stop"`.
  - `TestDoAuction_Stop_Immortal_Active_ReturnsItem`: immortal, active auction → broadcast `"Sale of X has been stopped by an Immortal"`, seller gets item back, buyer (if distinct) gets bet refunded, `Auction.Item == nil`.
  - `TestDoAuction_Stop_Immortal_SameBuyerAndSeller_NoRefund`: Buyer == Seller — no refund issued.
- **Implement** the stop branch (L3957-3982).
- **Mutation verify:** Flip `ch.IsImmortal()` gate to always-true → `TestDoAuction_Stop_Mortal_FallsThrough` fails (mortal now gets the stop path); revert.

### G8 — Disconnect / quit gate (S, ~1h)

- **Test first:**
  - `TestDoQuit_BlockedWhileAuctioning_Seller`: set `Auction.Seller = ch`, `Auction.Item = obj`, `DoQuit(ch, "")` → quit blocked, ch still in room.
  - `TestDoQuit_BlockedWhileAuctioning_Buyer`: set `Auction.Buyer = ch`, `Auction.Item = obj` → blocked.
  - `TestDoQuit_NotBlocked_NoActiveAuction`: `Auction.Item == nil` → quit proceeds normally.
  - `TestCloseDescriptor_SellerDropout_CancelsAuction`: direct test against `g.closeDescriptor` with a seller mid-auction → `Auction.Item == nil`, broadcast `"cancelled — the seller has left"`, buyer refunded.
  - `TestCloseDescriptor_BuyerDropout_RevertsBet`: buyer mid-auction → `Auction.Buyer == Auction.Seller`, `Auction.Bet == Auction.Starting`, broadcast `"top bidder has disconnected"`.
  - `TestCloseDescriptor_PreservesHistory`: either dropout → `Auction.History` array is unchanged.
- **Implement** the two-layer gate per §D8. Requires `internal/act/info.go` edit (DoQuit) and `internal/game/loop.go` edit (closeDescriptor).
- **Mutation verify:** Remove the `if g.world.Auction.Seller == ch` branch → seller-dropout test fails; revert.

### G9 — Integration: end-to-end auction with testclient (M, ~1.5h; lower priority)

**Optional** — drop if schedule-pressured. Two-client test:

- Client A: high-level immortal, carries a sword.
- Client B: level-5 mortal, starts with 50000 gold.
- Timeline:
  - A: `auction sword 1000`.
  - Harness: advance pulse until auction broadcast arrives (`Auction: A new item is being auctioned`).
  - B: `auction bid 20000`.
  - Harness: advance pulse PULSE_AUCTION × 3 times.
  - Assert: B sees `"going once"`, `"going twice"`, `"sold to"` in order; B's inventory contains the sword; B's gold = 30000; A's gold delta = +18000.

Relies on the `QuickLoginTwo` helper from Tier 9 (`internal/testclient/login.go`).

- **Mutation verify:** Change the tick dispatch to no-op in `pulse()` → integration asserts time out on the `"going once"` line; revert.

---

## Acceptance Criteria

1. `DoAuction("")` with no current item prints `"There is nothing being auctioned right now…"`.
2. `DoAuction("")` with a current item prints current bid, item short_descr, weight, value, level, wear-location, and per-item-type info (container capacity / scroll spells / weapon damage / armor AC).
3. Immortal sees seller, buyer, round, time-left in the info block; mortal does not.
4. NPC caller returns silently (no output).
5. Level < 3 caller is refused with `"at least level three"`.
6. Non-immortal caller between hour 18 and 9 with no active auction is refused with `"auctioneer works between"`.
7. Immortal bypasses the time gate.
8. `DoAuction("stop")` from an immortal with an active auction cancels, broadcasts `"stopped by an Immortal"`, returns item to seller, refunds distinct buyer.
9. `DoAuction("bid 1000")` validates: item present, bidder-level ≥ item-level, bidder != seller, amount ≥ starting, amount ≥ current + 10000, amount ≤ player's gold, amount ≤ 2 billion, optional keyword-arg3 matches item.
10. Successful bid: previous distinct bidder refunded, new bidder's gold debited, `Bet`/`Buyer` updated, `Going=0`, `Pulse=PULSE_AUCTION`, broadcast emitted.
11. `DoAuction("<item-kwd> <min-bet>")` validates carried, not-in-noauction, not modified, not decaying, not clan, not permanent, not in history; accepts only whitelisted item types.
12. Successful start: item removed from carrying, `Auction` populated, history ring rotated (newest at index 0), broadcast emitted.
13. Second start while active: mortal gets `"Try again later"` + `PULSE_VIOLENCE` wait state; immortal gets message without wait state.
14. `auctionUpdate` called every `PULSE_AUCTION` advances `Going` 0→1→2→3 with the right broadcasts at each stage.
15. `Going=3` with bet > starting and buyer != seller: sold — buyer receives item, seller receives `Bet * 0.9` gold, area receives `Bet * 0.1` tax via `boostEconomy`, `Auction.Item = nil`.
16. `Going=3` with bet == starting or buyer == seller: returned to seller, seller charged `item.Cost * 0.05` tax (floored at zero-gold), `Auction.Item = nil`.
17. `DoQuit` refuses while the caller is the active auction's seller OR buyer.
18. `closeDescriptor` on disconnect: seller-drop cancels auction + refunds buyer + broadcasts; buyer-drop resets buyer=seller + resets bet to starting + broadcasts.
19. History ring ages out one entry per `6 * AUCTION_MEM = 18` idle ticks.
20. `ParseBet("+50")` against current 1000 returns 1500; `ParseBet("x10")` returns 10000; `Advatoi("14k42")` returns 14420.
21. `util.NumPunct(1234567)` returns `"1,234,567"`.
22. `world.New(d).Auction` is non-nil and zero-valued after construction.
23. Full `go test ./...` green; `make test` green; `go vet ./...` clean.

---

## Scope Cuts / Deferrals

- **GSC (ENABLE_GOLD_SILVER_COPPER) variant** — out of Phase 6 per roadmap.
- **`noauction.dat` admin CRUD (`do_noauction`)** — C DOES ship an admin command at `act_wiz.c:11120-11174` (list + toggle + save). This Phase-6 plan ports load-only; the admin command is tracked as a follow-up in `TODO.md`. Rarely used in practice (list size typically static post-setup).
- **Auction persistence across reboot / hotboot** — C never persists auction state either (boot resets to empty). Hotboot plan (Wave 0) must decide whether to serialize mid-auction state. Out of this plan.
- **Carry-weight check on sold-to-buyer / returned-to-seller paths** — see §Open Questions Q4. If unresolved at dispatch time, fall back to "always give to character" (skip the obj-to-room drop branch).
- **`ms_find_obj` mud-school quest hook** — Go has no mud-school subsystem; port as no-op.
- **Auction channel `toggle`** — existing `DoChannels` (Tier 10) already handles `auction` Deaf bit; no extra command.
- **Per-ticking CHAR_DATA pointer GC** — Go's GC handles unreachable chars automatically; no `recycle_char` equivalent needed. Disconnect-dropout path only needs to clear the `world.Auction` pointer, not free memory.
- **`GainExp` on buyer / seller** — C does not award XP for auction participation. No drift risk.
- **Per-item spell_identify full breakdown** — G4b ports the common cases; obscure cases (ITEM_FIRE charge counts, ITEM_HERB_CON saturation) can be simplified to `"Level N item of type Xs."` — matches the C `default:` path, not gameplay-invented.

---

## Open Questions

Each with recommended answer and rationale.

### Q1 — Tick cadence: piggyback on `PULSE_TICK` or dedicated `pulseAuction`?

**Recommended: dedicated `pulseAuction` (§D3).** Matches C's cadence (9s vs 70s). Adds one int field + one `if` block to `GameLoop` — minimal footprint. `PULSE_TICK` is already load-bearing for char/obj/weather/room updates; adding auction work there bloats a hot path.

### Q2 — `world.New` initialization vs lazy allocation?

**Recommended: eager init at `world.New` (§D1).** Single allocation, fewer nil-guards. If boot-time memory is a concern (it isn't — a 7-field struct), lazy init via `EnsureAuction()` helper is trivially refactorable later.

### Q3 — `NOAUCTION_DATA` list: port loader, skip entirely, or stub?

**Recommended: port a minimal loader (§D7).** Loader reads `system/noauction.dat` if present (one vnum per line); missing file → empty slice. No admin command. Matches C's "hardcoded list populated at boot" semantic. If the file is absent in `db/system/` (verify at boot), the behaviour is: any carryable non-clan non-permanent item of a whitelisted type can be auctioned — which is the C default for a fresh MUD with no noauction list configured.

**Verification note for the executor:** `ls /home/eilidh/src/smaug/db/system/` confirmed at audit time (2026-04-18) — `noauction.dat` is **not present**. C's own writer (`act_wiz.c:11112-11114`) emits one `%d\n` per vnum plus a literal `0\n` terminator. Loader must: (a) treat missing file as empty (no error), (b) parse integers one per line, (c) stop at vnum 0 (terminator), (d) log `util.Bug("noauction.dat not found — no auction blacklist")` at INFO when missing.

### Q4 — Carry-weight check on obj-to-char vs obj-to-room?

**Recommended: skip the check; always `ObjToChar` to the recipient.** Go's combat/inventory path has never modeled carry-weight caps (audit 2026-04-18: no `CanCarryW`, `GetCarryW`, `WeightCap` helpers anywhere). Adding the C carry-weight cap here introduces a cross-cutting concern that affects every inventory operation, not just auctions. Deferring the cap to a future "carry-weight parity" pass is cleaner than one-off-in-auction.

**Fallback if the human disagrees:** the C code at `update.c:3247-3262` (sold branch) and `:3293-3308` (unsold branch) does `obj_to_room` on overcap. Port that verbatim against a stub `handler.CanCarryObj(ch, obj) bool` that always returns true for now. The dead branch stays dead until the weight subsystem lands.

### Q5 — `ms_find_obj` drunk/mental-state hook?

**Recommended: port as `// C: ms_find_obj (no-op — Go has not wired mental-state fail-chance into any object command)`.** `ms_find_obj` (`handler.c:2941`) is a generalized check that probabilistically makes get/drop/put/quaff/recite/eat/drink/auction/pick fail when the character's `mental_state` + drunkenness are high. It is NOT mud-school-specific (earlier audit claim corrected 2026-04-18). Go has not ported this side-effect to any of the parallel commands; auction is consistent with that omission. If a future subsystem adds it, the auction call site can be restored. Zero gameplay loss in the current Go port.

### Q6 — Hotboot interaction?

**Recommended: out-of-scope; document in hotboot plan.** Hotboot (Wave 0) is a design-exploration plan; when it promotes to executable, it decides whether mid-auction state (`world.Auction`) is serialized across the process restart. C does not serialize it. Matching C = accept that hotboot cancels any active auction. This is an acceptable divergence; note in the hotboot plan's open questions.

### Q7 — `world.Auction.Seller.Gold` typing?

**Recommended: assume `Gold int` on `CharData`.** Verified: `ch.Gold` is `int` at `internal/types/character.go` (confirm exact line at G5 time). The tax math uses truncating integer division (`Bet * 10 / 100` or `int(Bet * 0.1)`); port using `Bet / 10` (sold tax is exact 10%; C uses `(int)(bet * 0.1)` which is the same for positive integers < 10M). Use integer arithmetic throughout to avoid float drift: `pay := bet * 9 / 10; tax := bet - pay`. This handles odd-bet rounding (e.g., `bet=11` → `pay=9, tax=2`, preserving the invariant `pay + tax == bet`). Note: C's `(int)(bet * 0.1) + (int)(bet * 0.9)` can lose 1 gold on certain values — see `bet=11`: C gets `pay=9, tax=1, total=10` (1 gold evaporates). Port-with-fidelity would match; port-with-fix uses `bet - pay`. **Recommendation: port-with-fix** since the 1-gold drift is obviously unintended.

### Q8 — Broadcast message wording when an auction is cancelled via disconnect?

C never emits such a message (C assumes quit-gate prevents disconnect). Go needs SOMETHING. **Recommended message:** `"Auction: The auction has been cancelled — the seller has left the game."` for seller drop; `"Auction: The top bidder has left the game. Bidding re-opens at the starting price."` for buyer drop. Document as "Go-originated wording, non-blocking divergence from C" in the commit message.

---

## Risk Analysis

**High risk:**

- **Disconnect handling (G8)** — the two-layer gate is a Go-originated safety net. C relies solely on `do_quit`. Getting the defensive `closeDescriptor` clear right matters because a leak here means a dangling `*CharData` in `world.Auction` and every subsequent `auctionUpdate` tick crashes on nil-deref. Mitigation: the defensive clear is idempotent (clearing a nil pointer is a no-op); every test in G8 has a direct assertion on `Auction.Item == nil` post-drop.

**Medium risk:**

- **Gold-type integer arithmetic (Q7)** — using `int(bet * 0.9)` introduces float in an integer path. Stick to `bet * 9 / 10`. Mitigation: one test case (`TestAuctionUpdate_Sold_OddBet_NoGoldLeak`) asserts `pay + tax == bet`.
- **History-ring rotation (G6)** — C uses `memmove` which handles overlapping regions. Go's slice semantics are different; a naive `copy(Auction.History[1:], Auction.History[:])` does the right thing but reverses one has to verify direction. Mitigation: `TestDoAuction_Start_HistoryRotation` pins the expected order.
- **Pulse ordering in `pulse()`** — the auction tick must run AFTER input processing but BEFORE flush. Current order (input → pulse-dispatches → flush) handles this naturally if the new `pulseAuction` block is inserted between `pulseTick` and `pulseSave`. Mitigation: integration test G9 verifies the full chain.

**Low risk:**

- `parsebet` / `advatoi` (G0) — well-specified in C `bet.h` with docstring; mechanical port.
- Item-type whitelist (G6) — list is copy-paste from C; no invention.
- Broadcast formatting — reuses shipped `BroadcastAuction`.

---

## Adversary-Resolved Concerns

**`Agent` tool unavailable in this manager environment.** The manager was briefed to "dispatch 1-3 adversary subagents via `Agent` tool with `subagent_type: adversary`." The function schema exposed to this manager contains only `Read`, `Glob`, `Grep`, `Bash`, `Write`, `Edit` — no `Agent` tool. This limitation is consistent with the prior Wave 1 plans (marriage / arena / etc. all note external adversary review is pending). Structured self-review follows.

### Self-review pass (mechanical cross-check against C sources)

| Claim | C citation re-verified? |
|---|---|
| `do_auction` non-GSC starts at `act_obj.c:3775` | YES — `Read act_obj.c 3770-3790` confirms `#else` on L3773, function on L3775. |
| `auction_update` non-GSC starts at `update.c:3183` | YES — `Read update.c 3180-3200` confirms `#else` on L3181, function on L3183. |
| `auction->pulse` decrement call site is `update.c:2746-2750` | YES — `Read update.c 2735-2755` confirms. |
| `PULSE_AUCTION = 9 * PULSE_PER_SECOND` | YES — `mud.h:408`. |
| `AUCTION_MEM = 3` | YES — `mud.h:3609`. |
| `auction_data` C struct has `history[AUCTION_MEM]` + `hist_timer` | YES — `mud.h:3620-3621`. |
| Level-3 gate, time-gate 9-18, 10000-over-current, 2B max, self-bid | YES — all verified against `act_obj.c:3794, 3801, 4024, 4038, 3996`. |
| Tax split 10% sold, 5% unsold | YES — `update.c:3266` (sold `0.10`), `:3311` (unsold `cost * 0.05`). |
| `talk_auction` already shipped as `BroadcastAuction` | YES — `auction.go:25-45` matches C filter. |
| 33 test-coverage count on existing stub | Re-read `auction_test.go` — 8 tests shipped (6 Broadcast + 2 stub). (The "33" figure in some notes refers to Tranche C `util.Act` sites, not this file — irrelevant to this plan.) |
| `closeDescriptor` does NOT clear `Auction.Seller` / `Auction.Buyer` today | YES — `Read loop.go 866-900` confirms no auction reference. |
| `DoQuit` TIMER_RECENTFIGHT gate exists but no auction gate | YES — `Grep` confirms no `Auction` reference in `info.go`. |

### Edge-case self-review

- **G2 `world.New` change is non-breaking.** All existing callers of `world.New` assume `Auction` may be nil today. Setting it to a non-nil zero-value struct cannot break any caller that was already nil-guarding (a nil-guard `if w.Auction != nil` now evaluates true, but the subsequent read of `.Item` returns nil as before). Low risk. Mitigation: G3 tests assert the non-nil pointer explicitly.
- **G8 defensive clear race with tick.** Concurrency model: the game loop is single-threaded (one goroutine runs `pulse()`). `closeDescriptor` is called from within `pulse()` (via `cleanupDescriptors`, `loop.go:853-863`). `auctionUpdate` is called earlier in the same `pulse()`. There is no goroutine concurrency risk. The ONLY risk is ordering within a single pulse: if the disconnect is detected on pulse N and auctionUpdate runs on pulse N (before cleanup), the dangling pointer survives one tick. Acceptable: Go's runtime does not GC a referenced char; `auctionUpdate` reads `Auction.Seller` and it's still a valid pointer (the descriptor has been marked `Connected = -1`, but the char struct still exists). Broadcast to `Seller.Desc` is a no-op (Desc was cleared). No crash. Mitigation: defensive clear runs on the NEXT pulse; verified at `TestCloseDescriptor_SellerDropout_CancelsAuction`.
- **Ring rotation direction.** C `memmove` at `act_obj.c:4212-4214`: `memmove(dst=history+1, src=history, len=(AUCTION_MEM-1)*sizeof)`. Shifts RIGHT (forward in memory). `history[0]` then becomes the new entry. Go equivalent: `copy(History[1:], History[:AUCTION_MEM-1])` then `History[0] = newEntry`. Verified correct.
- **Bet math with `bet - pay`.** Positive-gold invariant: if `bet=100, pay=90, tax=10`, sum is 100. If `bet=101, pay=90, tax=11`, sum is 101. If `bet=11, pay=9, tax=2`, sum is 11. Port-with-fix is correct.

### Remaining unresolved risks (no adversary agreement possible; flag for human)

- Carry-weight check (Q4) — human decision needed on whether to stub `CanCarryObj` or skip entirely.
- Disconnect-cancellation broadcast wording (Q8) — Go-originated text; reasonable default provided.

---

## Dispatch Readiness

**Ready to dispatch:** YES, with the following caveats:

1. External adversary pass on the plan (not done — `Agent` tool unavailable).
2. Q3 `noauction.dat` presence verification (mechanical — part of G2).
3. Q4 carry-weight policy — human input preferred; fallback (skip) documented.

Effort estimate: **G0 (1.5h) + G1 (0.5h) + G2 (0.75h) + G3 (3h) + G4 (1h) + G4b (1h) + G5 (2h) + G6 (2.5h) + G7 (0.75h) + G8 (1h) + G9 (1.5h) = ~15.5h single-session worker.** Split into two sessions recommended: foundation (G0-G3) then command+tick (G4-G8), with G9 optional.

---

## Relevant File Paths

- C: `/home/eilidh/src/smaug/src/act_obj.c:3775-4240`, `/home/eilidh/src/smaug/src/update.c:3183-3326`, `/home/eilidh/src/smaug/src/update.c:2746-2750`, `/home/eilidh/src/smaug/src/bet.h:118-210`, `/home/eilidh/src/smaug/src/mud.h:408`, `/home/eilidh/src/smaug/src/mud.h:3609-3622`, `/home/eilidh/src/smaug/src/mud.h:778-783`.
- Go new: `/home/eilidh/src/smaug/smaug-go/internal/util/parsebet.go`, `/home/eilidh/src/smaug/smaug-go/internal/util/parsebet_test.go`, `/home/eilidh/src/smaug/smaug-go/internal/util/numpunct.go`, `/home/eilidh/src/smaug/smaug-go/internal/util/numpunct_test.go`, `/home/eilidh/src/smaug/smaug-go/internal/game/update_auction_test.go`.
- Go modified: `/home/eilidh/src/smaug/smaug-go/internal/types/misc.go` (AuctionData extension), `/home/eilidh/src/smaug/smaug-go/internal/world/world.go` (init + NoAuction slice), `/home/eilidh/src/smaug/smaug-go/internal/game/loop.go` (pulseAuction slot + closeDescriptor defensive clear), `/home/eilidh/src/smaug/smaug-go/internal/game/update.go` (auctionUpdate method), `/home/eilidh/src/smaug/smaug-go/internal/act/auction.go` (stub → state machine), `/home/eilidh/src/smaug/smaug-go/internal/act/auction_test.go` (extend from 8 to ~35 tests), `/home/eilidh/src/smaug/smaug-go/internal/act/info.go` (DoQuit auction gate), `/home/eilidh/src/smaug/smaug-go/internal/persist/subsystems.go` (optional noauction.dat loader).
