# Plan: Phase 6 — Clan Officer Commands

**Status:** Planned (2026-04-18). Self-reviewed (manager subagents in this environment cannot dispatch `Agent` — external adversary pass still pending before execution).
**Priority:** Wave 2 of Phase 6 (`phase6-roadmap.md`). Depends on `SaveClan` — bundled here as G0.
**Scope:** New file `internal/act/clan_officer.go` + `internal/act/clan_officer_test.go`. New `SaveClan` function in `internal/persist/subsystems.go` + `internal/persist/subsystems_test.go`. Modifications to `internal/persist/subsystems.go` `loadClan` (extend with missing keys for round-trip fidelity), `internal/boot/boot.go` (command registration), and a small note in `internal/types/pcdata.go` (none expected — `Bestowments` already present at line 25).

---

## Problem

C ships three clan-officer commands in `src/clans.c` totalling ~500 C LOC (not 300; the roadmap estimate was low):

- `do_induct` at `src/clans.c:928-1097` (170 LOC) — clan officer adds a PC to the clan.
- `do_outcast` at `src/clans.c:1168-1340` (173 LOC) — clan officer removes a PC from the clan.
- `do_bestow` at `src/act_wiz.c:7079-7135` (57 LOC) — immortal grants a PC access to specific commands via `PCData.Bestowments` string (this is the mechanism `do_induct` / `do_outcast` check to decide "is this PC an officer?").

**Roadmap scope correction (2026-04-18).** The `phase6-roadmap.md:205` entry lists `promote` / `demote` / `induct` / `outcast` / `bestow`. **There is no `do_promote` or `do_demote` in shipped SMAUG** — verified by `grep -rn "do_promote\|do_demote"` across `src/` returning zero function definitions. Clan-rank management (promoting a member to `number1` / `number2`, demoting them back) happens via `do_setclan <clan> number1 <playername>` at `src/clans.c:1472-1487`. Porting `do_setclan`'s full surface is a larger deferred job (it edits 30+ clan fields including pkill ranges, object vnums, guard mobs, align, memlimit — most of which are builder-only not officer-facing). This plan ports only the three officer-facing commands (`induct` / `outcast` / `bestow`) plus a minimal `do_clan_setleader`-style surface if human review at Q1 wants the rank-management holes filled. Open Question Q1 captures the decision.

The Go port has:

- `ClanData` struct at `internal/types/clan.go:5` with all C fields (`Leader`, `Number1`, `Number2`, `Deity`, `Members`, `MemLimit`, `ClanType`, `Class`, etc.).
- `PCData.Bestowments string` at `internal/types/pcdata.go:25` — **read and written** by `SavePlayer`/`LoadPlayer` (`persist/player.go:210-211`), exercised by the existing fixture `Testchar_full:49`.
- `PCData.Clan *ClanData` at `pcdata.go:6`, `PCData.ClanName string` at `pcdata.go:16`.
- `LoadClansFromDir` at `internal/persist/subsystems.go:16` — reads `clan.lst` and dispatches `loadClan` for each `.clan` file.
- Scaffolding Go commands `DoClans`/`DoClanInfo`/`DoClanJoin`/`DoClanLeave`/`DoClanDeposit`/`DoClanWithdraw` in `internal/act/clan.go`. None correspond to the officer-facing surface — `DoClanJoin` for example lets any PC join any clan with zero gating, which is a stub (not a port of `do_induct`).
- `CLAN_NOKILL` / `CLAN_ORDER` / `CLAN_GUILD` constants at `internal/types/enums.go:414-416`.
- `PCFLAG_DEADLY` at `internal/types/constants.go:649`. `PLR_NICE` at `internal/types/enums.go:975`. `LANG_CLAN` at `internal/types/constants.go:183`.
- `CharData.Speaks` / `CharData.Speaking` as `int` (bitfield — `constants.go` line range around 180 holds `LANG_*` constants). `ch.Act.IsSet(...)` / `ch.Act.Set(...)` / `ch.Act.Clear(...)` API on `BitVector`.
- `handler.GetCharRoom(ch, arg)` at `handler/find.go:14` — matches C `get_char_room`. `handler.GetCharWorld(w, ch, arg)` at `find.go:52` — matches C `get_char_world`.
- `util.Act` (Tranche C signature, per-call `aType`) at `util/act.go:324`. `util.IsName` at `util/strings.go:128` — matches C `is_name`. `util.Bug` at `util/log.go:24`. `util.SmashTilde` at `util/strings.go:46`. `util.OneArgument` at `util/strings.go:14`.
- `ch.IsNPC()` / `ch.IsImmortal()` / `ch.GetTrust()` helpers. `types.LEVEL_IMMORTAL`, `types.LEVEL_LESSER`, `types.LEVEL_IMPLEMENTOR` defined.
- `types.AT_MAGIC` at `enums.go:174` (= `AT_COLORBASE + 13`), `types.ECHOTAR_PK = 3` at `constants.go:130`.
- `WorldRef.Skills` slice; each entry `*SkillType` with `SkillAdept [MAX_CLASS]int` field and `Guild int` field — verified at `types/skill.go:52` (`Guild`) and `types/skill.go:8` (`SkillAdept`). Iterating `for i, sk := range WorldRef.Skills` yields the C `sn` value; `sk.SkillAdept[ch.Class]` replaces `GET_ADEPT(ch, sn)`.

**Missing:**

- `DoInduct` / `DoOutcast` / `DoBestow` commands (verified: zero Go matches for `DoInduct`/`DoOutcast`/`DoBestow`/`DoPromote`/`DoDemote`).
- `SaveClan` — **does not exist in Go.** Verified via `grep -rn "SaveClan\b" smaug-go/` returning zero hits. Every officer command in C ends with `save_clan(clan);` (`src/clans.c:1095`, `:1338`), so the Go port cannot land officer commands that persist rank changes without this function. Bundled into G0 of this plan.
- `save_member_lists` / `add_member` / `remove_member` — C `src/clans.c` has helpers that maintain a per-clan member roster file. **Deliberately out of scope (§Scope Cuts)** — the in-memory `clan.Members` counter is sufficient for officer-command mechanics; the separate member-list file is a display nicety whose absence does not break the commands. Follow-up.
- `add_loginmsg` — C `do_outcast` queues an offline-player notice (`src/clans.c:1323`). Go has no loginmsg system. **Deferred** — the `if victim->desc && victim->desc->host` fallback in C matches what the port will do (skip the loginmsg silently when the victim is linkdead; log via `util.Bug` instead).
- `loadClan`'s `Bestowments` field handling — not relevant to clan load (`Bestowments` is a PCData field, not ClanData). The existing `loadClan` is **incomplete**: it does not read `Abbrev`/`LeadRank`/`OneRank`/`TwoRank`/`MemLimit`/`ClanObjFour`/`ClanObjFive` keys that C writes (audit-clan-officer 2026-04-18 found `Abbrev` missing in addition to the 6 originally listed). For a round-trip `SaveClan` test to succeed on all fields the loader must be extended. Bundled into G0.

---

## C Reference (authoritative)

All citations against `src/clans.c` and `src/act_wiz.c` (HEAD).

### `save_clan` — `src/clans.c:177-255`

Writes a `#CLAN` block to `CLAN_DIR/<clan.filename>`. 40-odd `fprintf` calls emit every string field with a `~` terminator and every numeric field bare. Key ordering:

1. `#CLAN` header.
2. Strings (all `~`-terminated): `Name`, `Abbrev`, `Filename`, `Motto`, `Description`, `Deity`, `Leader`, `NumberOne`, `NumberTwo`, `Badge`, `Leadrank`, `Onerank`, `Tworank`.
3. `PKillRangeNew` / `PDeathRangeNew` — 7 space-separated ints each, no tilde.
4. Numeric fields: `MKills`, `MDeaths`, `IllegalPK`, `Score`, `Type`, `Class`, `Favour`, `Strikes`, `Members`, `MemLimit`, `Alignment`, `Board`, `ClanObjOne`-`ClanObjFive`, `Recall`, `Storeroom`, `GuardOne`, `GuardTwo`.
5. `End\n\n` followed by `#END\n`.

C uses `fpReserve` swap before/after the write to manage a tight FD budget. Go has no equivalent; the `os.Create` / defer `f.Close()` pattern is the idiom.

### `do_induct` — `src/clans.c:928-1097` (verbatim flow)

1. **Caller gate** (L934-953): NPC OR no-clan → `_("Huh?\n")`. Caller must also be one of:
   - `is_name("induct", ch->pcdata->bestowments)` — bestowed the `induct` command.
   - `!str_cmp(ch->name, clan->deity)` — clan deity.
   - `!str_cmp(ch->name, clan->leader)`.
   - `!str_cmp(ch->name, clan->number1)`.
   - `!str_cmp(ch->name, clan->number2)`.
   - None match → `_("Huh?\n")`.
2. **Parse** (L955) — one arg.
3. **Empty arg** (L957-961) → `"Induct whom?\n\r"`.
4. **Target resolve** (L963-967) — `get_char_room(ch, arg)` (same-room only, not world-wide; induction is face-to-face). NULL → `"That player is not here.\n\r"`.
5. **NPC check** (L969-973) → `"Not on NPC's.\n\r"` (note apostrophe-s).
6. **Immortal check** (L975-979) → `"You can't induct such a godly presence.\n\r"`.
7. **Peaceful-target gate** (L981-986): non-`IS_PKILL(victim)` AND clan type is NOT `CLAN_GUILD`/`CLAN_ORDER`/`CLAN_NOKILL` → `"You cannot induct a peaceful character.\n\r"`. (Guilds, orders, and nokill clans accept peaceful inductees; pkill clans require a deadly target.)
8. **Guild class-match** (L996-1005): if `clan->clan_type == CLAN_GUILD` and `victim->class != clan->class` → `"This player's will is not in accordance with your guild.\n\r"`.
9. **Level gates** (L1008-1019): non-guild only — victim < 10 → `"This player is not worthy of joining yet.\n\r"`. Victim level > inductor level → `"This player is too powerful for you to induct.\n\r"`.
10. **Already-in-clan gate** (L1022-1051) — branches on the *victim's existing* clan type (not the inducting clan's type). Three sub-messages per type (guild/order/clan), each with a sub-sub-distinction "your X" vs "an/a X". Six total messages. Port verbatim.
11. **Member-limit gate** (L1052-1057) → `"Your clan is too big to induct anymore players.\n\r"`.
12. **Commit** (L1058-1090): `clan->members++`. Non-order-non-guild → `SET_BIT(victim->speaks, LANG_CLAN)`. Pkill-type only (not nokill/order/guild) → `xREMOVE_BIT(victim->act, PLR_NICE)` + `SET_BIT(victim->pcdata->flags, PCFLAG_DEADLY)`. Pkill-type only → award all skills whose `skill_table[sn]->guild == clan->class` at `GET_ADEPT(victim, sn)` and print "$instructor instructs you in the ways of $skillname."
13. **Link-in** (L1086-1090) — set `victim->pcdata->clan = clan`, replace `clan_name`, `add_member(victim, clan_name)`, `save_member_lists()`.
14. **Broadcast** (L1091-1093): three `act(AT_MAGIC, ..., TO_CHAR/TO_NOTVICT/TO_VICT)` calls — "You induct $N into $t", "$n inducts $N into $t", "$n inducts you into $t" — `$t` is `clan->name`.
15. **Persist** (L1094-1095): `save_char_obj(victim); save_clan(clan);`.

### `do_outcast` — `src/clans.c:1168-1340` (verbatim flow)

1. **Caller gate** (L1176-1195) — same shape as `do_induct` except bestowment check uses `"outcast"` keyword.
2. **Parse** (L1197) — one arg.
3. **Empty arg** (L1199-1203) → `"Outcast whom?\n\r"`.
4. **Target resolve** (L1205-1209) — same-room only. NULL → `"That player is not here.\n\r"`.
5. **NPC check** (L1211-1215) → `"Not on NPC's.\n\r"`.
6. **Officer-rank arithmetic** (L1218-1229): assign `x = 0/1/2/3` based on ch being Number2/Number1/Leader, and `y = 0/1/2/3` for victim's rank. Rank 0 means "ordinary member or the clan deity or a bestowed-`outcast` officer."
7. **Power gate** (L1231-1236): `x <= y && get_trust(ch) <= get_trust(victim)` → `"You are not powerful enough to outcast this character.\n\r"`. (Deliberately non-strict — an officer of equal rank with equal trust cannot outcast a peer; a superior rank beats equal trust; higher trust beats equal rank.)
8. **Self-outcast** (L1238-1255): `victim == ch` → three type-specific messages ("Kick yourself out of your own order/guild/clan?"). Return.
9. **Level gate** (L1257-1262): `victim->level > ch->level` → `"This player is too powerful for you to outcast.\n\r"`.
10. **Same-clan gate** (L1264-1281): `victim->pcdata->clan != ch->pcdata->clan` → three type-specific "does not belong to your X" messages.
11. **Skill-forget** (L1283-1296): non-guild/non-order/non-nokill only — iterate `skill_table`, for `guild == victim->pcdata->clan->class && name != NULL`: `victim->pcdata->learned[sn] = 0` + `"You forget the ways of %s.\n\r"`.
12. **Language + membership cleanup** (L1298-1316): if victim's `speaking & LANG_CLAN`, reset `speaking = LANG_COMMON`. `REMOVE_BIT(victim->speaks, LANG_CLAN)`. Decrement `clan->members`. Blank `Number1`/`Number2` strings if victim was one of them. Clear `victim->pcdata->clan`. Blank `victim->pcdata->clan_name`.
13. **Broadcast** (L1317-1323): three `act` calls to TO_CHAR/TO_ROOM/TO_VICT ("You outcast $N from $t", "$n outcasts $N from $t", "$n outcasts you from $t"). If victim has no `desc->host` (linkdead) → `add_loginmsg(victim->name, 6, NULL)` instead of the TO_VICT act-call.
14. **Echo to PKers** (L1325-1329): non-guild/non-order → `echo_to_all(AT_MAGIC, "%s has been outcast from %s!", ECHOTAR_PK)`.
15. **Persist** (L1337-1338): `save_char_obj(victim); save_clan(clan);`.

### `do_bestow` — `src/act_wiz.c:7079-7135`

Immortal command that edits `victim->pcdata->bestowments` string (a space-separated command whitelist).

1. `set_char_color(AT_IMMORT, ch)` — cosmetic; Go `util.Act` handles color.
2. Parse first arg.
3. Empty → `"Bestow whom with what?\n\r"`.
4. `get_char_world(ch, arg)` NULL → `_("They aren't here.\n")`.
5. NPC → `_("You can't give special abilities to a mob!\n")`.
6. `get_trust(victim) > get_trust(ch)` → `_("You aren't powerful enough...\n")`.
7. If `!victim->pcdata->bestowments` (nil) → `str_dup("")`. (In Go, string default is `""` already; no-op.)
8. Empty remaining arg OR `str_cmp(argument, "list") == 0` → `"Current bestowed commands on %s: %s.\n\r"` with the current list. Return.
9. `str_cmp(argument, "none") == 0` → clear `bestowments`, notify both. Return.
10. Else → `sprintf("%s %s", bestowments, argument)` (append with space prefix — the leading space is a quirk; C dumps a leading blank into the first assignment because the initial empty string concatenates `" <arg>"`; port verbatim). Notify victim + `"Done.\n"`.

**Anomaly:** `do_bestow` never calls `save_char_obj(victim)`. Bestowments persist only on the victim's next pfile save (disconnect, `save` command, quit). This matches C — do NOT add a `g.SavePlayer(victim)` at the end.

### Missing from this plan: `do_bestowarea` (`act_wiz.c:7004-7076`)

Adjacent to `do_bestow`, edits the same `bestowments` field but validates the `.are` suffix to gate builder access. Functionally distinct from officer-facing `bestow`. **Scope-cut to a follow-up plan** — a 4-line Open Question on whether immortal areafile bestowal is needed on the port. Captured in §Scope Cuts.

### Wiring (C tables)

- `tables.c:752-753` registers `do_induct`.
- `tables.c:1151-1152` registers `do_outcast`.
- `tables.c:385-388` registers `do_bestow` and `do_bestowarea`.

These are reflection-lookup entries used by the C `do_fun` resolver. Go's equivalent is `boot.go` `reg.Register(...)`.

---

## Go Current State

Verified 2026-04-18 against the `golang` branch at commit `26db5f9`.

### Fields / structs

- `ClanData` (`internal/types/clan.go:5-45`) has every field C writes. No schema change needed.
- `PCData.Bestowments string` already present and round-trips through SavePlayer/LoadPlayer.
- `PCData.Clan *ClanData` + `PCData.ClanName string` both present.
- `CharData.Speaks int` / `CharData.Speaking int` both present, loaded from pfile at `persist/player.go:116-117`.
- `PCFLAG_DEADLY = 1<<1` at `constants.go:649`. `PLR_NICE` at `enums.go:975`. `LANG_CLAN = 1<<18` at `constants.go:183`.

### Persist

- `loadClan` at `subsystems.go:29-113` reads most fields BUT **misses**: `LeadRank`, `OneRank`, `TwoRank`, `ClanObjFour`, `ClanObjFive`, `MemLimit`, plus does not handle `PKillRange[New]` / `PDeathRange[New]` block-read (7 ints on one line). Round-trip of `SaveClan → loadClan` will lose data unless these are added. G0 extends the loader.
- `SaveClan` — **missing entirely.**
- `saveClansToDir` — missing. Even after adding `SaveClan`, the officer commands call `save_clan` on exactly one clan (the officer's); no "save all clans" helper needed at this tier. Follow-up: a "save clans on shutdown" wiring in `internal/game/loop.go` belongs to a future persistence-polish plan.

### Commands

- `DoClans` / `DoClanInfo` / `DoClantalk` / `DoClanJoin` / `DoClanLeave` / `DoClanDeposit` / `DoClanWithdraw` / `DoDeities` / `DoDevote` / `DoNote` / `DoSnoop` in `internal/act/clan.go`. Note `DoClanJoin` is NOT a port of `do_induct` — it is a zero-gate stub where any PC joins any clan directly. Leave it (distinct command, registered as `join`), but document the divergence in §Scope Cuts — in production `induct` is the sanctioned flow; `join` is a developer convenience.
- `DoBestow` / `DoInduct` / `DoOutcast` — **absent.**
- `DoSetclan` — absent. Not in scope here; Q1 captures whether to minimally expose Leader/Number1/Number2 setters for rank promotion.

### Seams

- `handler.GetCharRoom(ch, arg)` — suitable for `DoInduct` / `DoOutcast` (C `get_char_room`).
- `handler.GetCharWorld(w, ch, arg)` — suitable for `DoBestow` (C `get_char_world`).
- `util.Act(AT_MAGIC, "...", ch, vch, arg1, arg2, TO_X)` — Tranche-C signature. `$t` in the format string — verify at G2 implementation that `util.Act`'s template engine handles `$t` (it does per Tranche C; `arg1` slot when set as a `string` prints as the `$t` substitution).
- `ch.Send(str)` / `ch.Sendf(fmt, ...)`.
- `util.IsName("induct", ch.PCData.Bestowments)` replaces C `is_name("induct", ...)`.
- `strings.EqualFold` replaces C `str_cmp` (case-insensitive compare).
- `ch.Act.IsSet(PLR_NICE)` / `ch.Act.Clear(PLR_NICE)` replace `xREMOVE_BIT(ch->act, PLR_NICE)`.
- `ch.Speaks |= int(LANG_CLAN)` / `ch.Speaks &^= int(LANG_CLAN)` replace `SET_BIT`/`REMOVE_BIT` — both fields are untyped `int` in Go, `LANG_CLAN` is `uint32`.

### Tests / fixtures

- `persist/subsystems_test.go:12` `TestLoadClansFromDir` exercises the loader with `db/clans/`. Test world has at least one loadable clan (check `db/clans/clan.lst` at G0 setup).
- `act/clan_test.go` — existing tests for DoClans/DoClanJoin — pattern to follow for the new tests.

---

## Go Design

### D0 — `SaveClan` (format-parity writer)

**Chosen:** new function `SaveClan(w io.Writer, c *types.ClanData) error` in `internal/persist/subsystems.go`. Mirrors the 40-ish `fprintf` calls of C `save_clan` (`src/clans.c:206-250`). Signature matches the `SavePlayer(w io.Writer, ch *types.CharData) error` pattern already established at `persist/player.go:455`; callers open the file and pass the writer. A second helper `SaveClanFile(dir string, c *types.ClanData) error` handles `os.Create(filepath.Join(dir, c.Filename))` and calls `SaveClan`. Keeping the two-layered split lets tests exercise the format-write with a `bytes.Buffer` while production opens a real file.

**Rationale over Option B (deferred `plan-phase6-saveclan.md`):**

- `save_clan` is pure text serialization — no sibling data touched, no subsystem cross-wires. ~60 Go LOC including error handling.
- Every officer command ends with `save_clan(clan)`. Shipping induct/outcast without persistence means a reboot loses every membership decision — the commands would be half-ported.
- The symmetric test pattern (write to `bytes.Buffer`, re-`loadClan`, assert field equality) mirrors the existing `TestSaveLoadPlayer_*` pattern and adds <30 lines.
- A separate plan doc would add ~4 hours of orchestrator overhead (author → adversary → merge) for work that fits in one task group here.

**Rejected Option B (separate plan).** Would make sense only if `SaveClan` had nontrivial design surface — it does not.

**File format — SaveClan output exactly matches C `save_clan` byte-for-byte** so a C-side server can still read the file if operators ever cross-deploy. This is a port-fidelity constraint: matching byte output avoids a data-migration scenario on C→Go switchover. Fixed spacing (`%-12s` style not used in C — use the two-space alignment C uses: `Name         %s~\n`, `PKillRangeNew   %d %d %d %d %d %d %d\n`, etc.). Verified against `src/clans.c:207-249`.

### D1 — `loadClan` completeness

Extend the existing loader at `subsystems.go:29-113` to handle every key C writes. Missing keys: `Abbrev`, `LeadRank`, `OneRank`, `TwoRank`, `ClanObjFour`, `ClanObjFive`, `MemLimit`, `PKillRangeNew` (7-int block), `PDeathRangeNew` (7-int block), legacy `PKillRange` / `PDeathRange` (same shape — preserved for backward compat, per C `fread_clan:437-479`).

**Latent bug to fix while here (audit-clan-officer finding, 2026-04-18):** the existing legacy `"PKills"` / `"PDeaths"` case at `subsystems.go:66-69` reads into `clan.PKills[0]` / `clan.PDeaths[0]`. C `fread_clan:434-435` reads legacy single-int `PKills`/`PDeaths` into **index 6** (`clan->pkills[6]`, `clan->pdeaths[6]`), not index 0. Fix while reworking the loader — change target index to [6] to match C semantics. This is pre-existing; not introduced by this plan.

Block-read pattern: when the key is `PKillRangeNew`, read 7 consecutive numbers into `clan.PKills[0..6]`. `Scanner.ReadNumber()` is the seam; verify at G0 that 7 sequential `ReadNumber()` calls work across a single source line.

The `Score` field (line 78 of current loader) already reads correctly. The `Strikes` field is read (line 85 current) but `SaveClan` must emit it.

### D2 — `DoInduct`

New file `internal/act/clan_officer.go`. Signature: `func DoInduct(ch *types.CharData, argument string)`.

Flow mirrors C step-for-step (see §C Reference `do_induct`). Go expression:

```go
func DoInduct(ch *types.CharData, argument string) {
    if ch.IsNPC() || ch.PCData == nil || ch.PCData.Clan == nil {
        ch.Send("Huh?\n")
        return
    }
    clan := ch.PCData.Clan
    if !isClanOfficer(ch, clan, "induct") {
        ch.Send("Huh?\n")
        return
    }
    arg, _ := util.OneArgument(argument)
    if arg == "" {
        ch.Send("Induct whom?\n\r")
        return
    }
    victim := handler.GetCharRoom(ch, arg)
    if victim == nil {
        ch.Send("That player is not here.\n\r")
        return
    }
    if victim.IsNPC() {
        ch.Send("Not on NPC's.\n\r")
        return
    }
    if victim.IsImmortal() {
        ch.Send("You can't induct such a godly presence.\n\r")
        return
    }
    // ... remaining gates verbatim per §C Reference.
}
```

Helper `isClanOfficer(ch *types.CharData, clan *types.ClanData, command string) bool` consolidates the 5-way officer check (bestow-keyword OR deity OR leader OR number1 OR number2). Used by both `DoInduct` and `DoOutcast`.

Helper `isPkill(ch *types.CharData) bool` — verify whether this exists; if not, inline via `!ch.IsNPC() && ch.PCData != nil && (ch.PCData.Flags & int(PCFLAG_DEADLY)) != 0` matching C `IS_PKILL` which is `IS_SET(ch->pcdata->flags, PCFLAG_DEADLY)`. Grep `IsPkill\b|IS_PKILL` at G2 start; add to `types/` if missing. **Found (2026-04-18):** `mudprog/ifcheck.go:543` computes `(uint32(chk.PCData.Flags) & types.PCFLAG_DEADLY) != 0`; same expression inlined in `DoInduct` is fine, or add a `CharData.IsPkill()` method — Q2 captures the choice.

**Skill-award loop** (step 12): iterate `for sn, sk := range WorldRef.Skills` and check `sk != nil && sk.Guild == clan.Class && sk.Name != ""`. On match: `victim.PCData.Learned[sn] = sk.SkillAdept[victim.Class]` + `victim.Sendf("%s instructs you in the ways of %s.\n\r", ch.Name, sk.Name)`.

**Persist** (step 15): call `g.SavePlayer(victim)` — but `DoInduct` has no reference to the GameLoop. Resolution: invoke `persist.SavePlayerFile` if it exists, else use the same "deferred-save" pattern other commands use (mark the character dirty; save on next tick). **Verify at G2 implementation** which seam matches the house style. Followed by `persist.SaveClanFile(boot.ClanDir, clan)` — requires exporting the clan-dir path from boot, or re-deriving via a `world.World.DataDir` seam. G2 resolves.

### D3 — `DoOutcast`

Same file. Follows C `do_outcast` step-for-step. Notable translations:

- **Officer-rank arithmetic** (C step 6) → switch statement matching `strings.EqualFold(ch.Name, clan.Leader/Number1/Number2)`; default 0.
- **Skill-forget loop** → iterate `WorldRef.Skills`, zero `learned[sn]` when `sk.Guild == victim.PCData.Clan.Class`.
- **Language cleanup** (C step 12):
  - `if uint32(victim.Speaking) & LANG_CLAN != 0 { victim.Speaking = int(LANG_COMMON) }`.
  - `victim.Speaks &^= int(LANG_CLAN)`.
- **Rank clear**: `if strings.EqualFold(victim.Name, clan.Number1) { clan.Number1 = "" }` (same for Number2).
- **Broadcast** (step 13): the TO_VICT branch guarded by linkdead check: `if victim.Desc != nil { util.Act(TO_VICT) } else { util.Bug("DoOutcast: linkdead victim %s — loginmsg deferred", victim.Name) }`. See §Scope Cuts for the `add_loginmsg` deferral.
- **Echo to PKers** (step 14): calls `echoToPKers(msg)` helper. Go has no `echo_to_all(ECHOTAR_PK)` — port a minimal version that iterates `WorldRef.Descriptors`, filters `d.Character != nil && !d.Character.IsNPC() && (d.Character.PCData.Flags & int(PCFLAG_DEADLY)) != 0`, sends `msg`. Local helper in `clan_officer.go`; DO NOT extend `util.Act` for this — the cut is narrower.

### D4 — `DoBestow`

Same file. Port matches C `do_bestow` exactly. The leading-space quirk in bestowments (step 10) is preserved — `bestowments = strings.TrimSpace(bestowments + " " + arg)` keeps the list visually tidy; but C preserves the leading space. Port C's shape verbatim: `victim.PCData.Bestowments = victim.PCData.Bestowments + " " + argument` — no trim. A fresh bestow produces `" induct"` with a leading blank; `util.IsName` is whitespace-tolerant per its contract so the leading blank is invisible to `DoInduct`'s `isClanOfficer` check. Tested at G4.

### D5 — Command registration

Add three lines to `internal/boot/boot.go` in the immortal-commands block near line 600 (where `hell`/`deny`/`pardon` live — ordinary immortal-trust level):

```go
reg.Register(&command.Command{Name: "induct", DoFun: act.DoInduct, Position: types.POS_RESTING, Level: 0})
reg.Register(&command.Command{Name: "outcast", DoFun: act.DoOutcast, Position: types.POS_RESTING, Level: 0})
reg.Register(&command.Command{Name: "bestow", DoFun: act.DoBestow, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
```

**Rationale — Level 0 on induct/outcast.** These are *clan-officer* commands, not immortal commands. The authority check is the in-command `isClanOfficer` gate, not a trust-level gate. A low-level PC who happens to be their clan's leader CAN `induct`. Matches C `do_induct` being registered with `Level 0` in `tables.c`.

**Position POS_RESTING** matches other player-facing clan commands (`join` is `POS_RESTING` in the Go `DoClanJoin` — verify at G5). `bestow` is `POS_DEAD` (it's an immortal command).

### D6 — Clan directory seam

`SaveClanFile` needs the clan directory path. Two options:

- **A** (chosen): export `boot.ClanDir` as a package-level `string`, set during `boot.Boot()` (it's already computed at `boot.go:322`); `SaveClanFile` takes the dir as a parameter, called from `DoInduct`/`DoOutcast` via `persist.SaveClanFile(boot.ClanDir, clan)`.
- B: stash the dir on `world.World.ClanDir string`; commands pull from `WorldRef.ClanDir`.

Option A is smaller and matches how `skillGetter` and similar are wired. Option B mingles config with world state.

---

## Task Groups

Five task groups. G0 is the persistence prerequisite; G1 is the officer-helper + `DoInduct` happy path; G2 is `DoInduct` edge cases; G3 is `DoOutcast`; G4 is `DoBestow` + command registration.

### G0 — `SaveClan` + `loadClan` completeness (M, ~3.5h)

**Deliverables:**

1. Extend `loadClan` in `internal/persist/subsystems.go` with cases for `LeadRank`, `OneRank`, `TwoRank`, `ClanObjFour`, `ClanObjFive`, `MemLimit`, `Strikes`, `Type` (verify — `"Type"` already handled), `PKillRangeNew` (7-int block), `PDeathRangeNew` (7-int block), legacy `PKillRange` / `PDeathRange` (read-and-discard per C for backward compat — same 7-int shape but unused by Go since `PKills[0..6]` is the canonical target).
2. Add `SaveClan(w io.Writer, c *types.ClanData) error` in the same file. 40-line series of `fmt.Fprintf` calls, matching C `save_clan:206-250` byte-for-byte on the emitted text. Use `util.SmashTilde` on every string field (belt-and-braces against builder-typed tildes in clan names).
3. Add `SaveClanFile(dir string, c *types.ClanData) error` — `os.Create(filepath.Join(dir, c.Filename))`, defer-close, call `SaveClan`. Returns wrapped error on either failure.
4. Add `TestSaveLoadClan_RoundTrip` and `TestSaveClan_FormatMatchesGolden` in `internal/persist/subsystems_test.go`. The round-trip test populates a `ClanData` with non-zero values for every field, calls `SaveClan`, re-parses with `loadClan` (modifying the test helper to accept a `bytes.Reader`), and asserts every field equals.
5. Add `TestLoadClan_MissingKeys` — verify that existing `db/clans/*.clan` files still load cleanly after the loader extensions (no regression).

**Constraints:**

- Do NOT change C clan file format; `SaveClan` must emit identical bytes to C's output for a given `ClanData`.
- Do NOT touch `LoadClansFromDir`'s signature — only its inner `loadClan`.
- `loadClan` signature is currently `loadClan(path string)` — file-based. For the round-trip test, extract a `readClan(sc *Scanner) (*types.ClanData, error)` helper called by both the file-path and `bytes.Reader` entry points. This refactor is invisible to external callers.

**TDD:**

- Write `TestSaveLoadClan_RoundTrip` first — it must fail because `SaveClan` does not exist (compile error). Drive until green.
- Mutation-verify: change `fmt.Fprintf(w, "Members      %d\n", c.Members)` to `"Members      %d\n", c.MemLimit` → test red → revert via `Edit` → green.
- `TestSaveClan_FormatMatchesGolden` holds a golden string (20-line representative clan) and asserts byte-for-byte output. Caught by mutation: reorder two `Fprintf` calls → test red.

**File paths:**

- `internal/persist/subsystems.go` (extend `loadClan`, add `SaveClan` + `SaveClanFile`).
- `internal/persist/subsystems_test.go` (add two test functions).

**Mutation-verification safety:** per `_shared.md` → Mutation Verification Safety. Banned git commands: `git checkout -- <file>`, `git restore <file>`, `git reset --hard`, `git stash` (any form). Revert via `Edit` only.

### G1 — `isClanOfficer` + `DoInduct` happy path (M, ~3h)

**Deliverables:**

1. New file `internal/act/clan_officer.go` with `isClanOfficer(ch *types.CharData, clan *types.ClanData, command string) bool` helper. Returns true if any of: `util.IsName(command, ch.PCData.Bestowments)`, `strings.EqualFold(ch.Name, clan.Deity)`, `strings.EqualFold(ch.Name, clan.Leader)`, `strings.EqualFold(ch.Name, clan.Number1)`, `strings.EqualFold(ch.Name, clan.Number2)`.
2. `DoInduct(ch, argument)` — happy path only: caller is clan leader, target same-room PC of appropriate type/level/class, no already-in-clan, no member-limit hit, pkill clan (so skills get awarded).
3. Tests covering the happy path + `isClanOfficer`'s 5 positive cases.

**Tests (TDD):**

- `TestIsClanOfficer_LeaderMatches` — ch.Name == clan.Leader → true.
- `TestIsClanOfficer_Number1Matches` — ch.Name == clan.Number1 → true.
- `TestIsClanOfficer_Number2Matches` — ch.Name == clan.Number2 → true.
- `TestIsClanOfficer_DeityMatches` — ch.Name == clan.Deity → true.
- `TestIsClanOfficer_BestowmentsMatches` — `bestowments = "induct outcast"`, command = "induct" → true.
- `TestIsClanOfficer_NoMatch` — random PC → false.
- `TestIsClanOfficer_CaseInsensitive` — clan.Leader = "Alice", ch.Name = "alice" → true.
- `TestDoInduct_Success_PkillClan_SetsFields` — after call, `victim.PCData.Clan == clan`, `victim.PCData.ClanName == clan.Name`, `clan.Members` incremented, `victim.Speaks & int(LANG_CLAN) != 0`, `victim.Act.IsSet(PLR_NICE) == false`, `victim.PCData.Flags & int(PCFLAG_DEADLY) != 0`.
- `TestDoInduct_Success_AwardsGuildSkills` — register a fake skill with `Guild = clan.Class`; after induct, `victim.Learned[sn] == sk.SkillAdept[victim.Class]`.
- `TestDoInduct_Success_Messages` — ch, victim, and a bystander in the room each see the documented messages with `AT_MAGIC` color.

**Constraints:**

- Test fixture pattern follows `internal/act/clan_test.go` — uses `WorldRef` set to a fresh `world.World`, seeded with two PCs in one room.
- Do NOT call `g.SavePlayer(victim)` in this task group (deferred to G2's persist-seam decision); allow the in-memory state to be the verified shape. Add a `// TODO G2: persist victim + clan` comment.

**Mutation-verification:** change `victim.Speaks |= int(LANG_CLAN)` to `victim.Speaks &^= int(LANG_CLAN)` → test red; revert via `Edit`.

**File paths:**

- `internal/act/clan_officer.go` (new).
- `internal/act/clan_officer_test.go` (new).

### G2 — `DoInduct` edge cases + persistence seam (M, ~3h)

**Deliverables:**

1. All 10 edge-case branches of `DoInduct` (caller-gate, syntax, not-here, NPC, immortal, peaceful, guild-class-mismatch, level-too-low, level-too-high, already-in-clan × 6 variants, member-limit).
2. Persist seam — decide A (direct `g.SavePlayer(victim)`) or B (deferred save via dirty flag) based on local grep of existing commands. Whichever pattern `DoClanJoin` uses, match it. Add `persist.SaveClanFile(boot.ClanDir, clan)` as the final step.
3. Tests for every edge case (one test per branch).

**Tests (edge cases):**

- `TestDoInduct_GateNPC` — ch is NPC → "Huh?" no mutation.
- `TestDoInduct_GateNoClan` — ch.PCData.Clan == nil → "Huh?".
- `TestDoInduct_GateNotOfficer` — ch in clan but not an officer → "Huh?".
- `TestDoInduct_SyntaxEmpty` → "Induct whom?\n\r".
- `TestDoInduct_VictimNotInRoom` → "That player is not here.\n\r".
- `TestDoInduct_VictimIsNPC` → "Not on NPC's.\n\r".
- `TestDoInduct_VictimIsImmortal` → "You can't induct such a godly presence.\n\r".
- `TestDoInduct_PeacefulVictim_PkillClan_Rejected` — victim not PCFLAG_DEADLY, clan is pkill type → "You cannot induct a peaceful character.\n\r".
- `TestDoInduct_PeacefulVictim_GuildClan_Accepted` — victim peaceful, clan is CLAN_GUILD → proceed to class check.
- `TestDoInduct_GuildClassMismatch` → "This player's will is not in accordance with your guild.\n\r".
- `TestDoInduct_VictimLevelTooLow` — victim.Level < 10, non-guild clan → "This player is not worthy of joining yet.\n\r".
- `TestDoInduct_VictimLevelTooHigh` — victim.Level > ch.Level → "This player is too powerful for you to induct.\n\r".
- `TestDoInduct_AlreadyInOrder_SameOrder` — victim already in this clan which is CLAN_ORDER → "This player already belongs to your order!\n\r".
- `TestDoInduct_AlreadyInOrder_DifferentOrder` → "This player already belongs to an order!\n\r".
- (Same pattern for guild, clan — 6 total already-in-clan tests.)
- `TestDoInduct_MemberLimitHit` — `clan.MemLimit > 0 && clan.Members >= clan.MemLimit` → "Your clan is too big to induct anymore players.\n\r".
- `TestDoInduct_Persists_SaveClanFileCalled` — stub `persist.SaveClanFile`; verify it was invoked with the expected clan.

**Mutation-verification:** change `victim.Level < 10` to `victim.Level <= 10` → one test goes red; revert.

**File paths:**

- `internal/act/clan_officer.go` (extend DoInduct).
- `internal/act/clan_officer_test.go` (extend tests).
- `internal/boot/boot.go` (export `ClanDir` package var).

### G3 — `DoOutcast` + echo-to-PKers helper (M, ~3.5h)

**Deliverables:**

1. `DoOutcast` full port — 14 C-step flow.
2. `echoToPKers(msg string)` local helper (matches C `echo_to_all(AT_MAGIC, msg, ECHOTAR_PK)`).
3. Tests for all branches.

**Tests:**

- `TestDoOutcast_GateNPC` / `_GateNoClan` / `_GateNotOfficer` — same shape as induct.
- `TestDoOutcast_SyntaxEmpty` → "Outcast whom?\n\r".
- `TestDoOutcast_VictimNotInRoom` → "That player is not here.\n\r".
- `TestDoOutcast_VictimIsNPC` → "Not on NPC's.\n\r".
- `TestDoOutcast_RankArithmetic_OfficerOutrankedByVictim` — ch is Number2 (x=1), victim is Leader (y=3), ch.Trust <= victim.Trust → "You are not powerful enough..." .
- `TestDoOutcast_RankArithmetic_SuperiorTrustBeatsEqualRank` — both are Number1 (x=y=2), ch.Trust > victim.Trust → proceed.
- `TestDoOutcast_SelfOutcast_Order` — victim == ch, clan.Type == CLAN_ORDER → "Kick yourself out of your own order?\n\r".
- `TestDoOutcast_SelfOutcast_Guild` / `_Clan` — same shape.
- `TestDoOutcast_VictimHigherLevel` — victim.Level > ch.Level → "This player is too powerful for you to outcast.\n\r".
- `TestDoOutcast_NotSameClan_Clan` — victim.PCData.Clan != ch.PCData.Clan, ch's clan type == plain clan → "This player does not belong to your clan!\n\r".
- (Same for order, guild.)
- `TestDoOutcast_Success_ClearsFields` — victim.PCData.Clan == nil, victim.PCData.ClanName == "", clan.Members decremented, LANG_CLAN removed, PLR_NICE unchanged (outcast does NOT flip NICE — C does not set it on outcast).
- `TestDoOutcast_Success_SkillsForgotten_PkillClan` — learned[sn] == 0 for all sn where `sk.Guild == clan.Class`.
- `TestDoOutcast_Success_BlanksNumber1` — victim.Name == clan.Number1 before → clan.Number1 == "" after.
- `TestDoOutcast_Success_BlanksNumber2` — same for Number2.
- `TestDoOutcast_Success_LinkdeadVictim` — victim.Desc == nil → no TO_VICT act-call, `util.Bug` invoked (capture via existing test-logger seam or assert no panic).
- `TestDoOutcast_Success_EchoPKersFires_NonGuildNonOrder` — test world with a third PC who IS_PKILL → sees the `"X has been outcast from Y!"` broadcast.
- `TestDoOutcast_Success_EchoPKersSuppressed_GuildClan` — clan.Type == CLAN_GUILD → no echo.

**Constraints:**

- `echoToPKers` lives as a private function in `clan_officer.go` — do NOT generalize into `util/` in this plan.
- Reuse `isClanOfficer` from G1.

**Mutation-verification:** change `clan.Members--` to `clan.Members++` → multiple tests red; revert.

**File paths:**

- `internal/act/clan_officer.go` (extend).
- `internal/act/clan_officer_test.go` (extend).

### G4 — `DoBestow` + all three command registrations (S, ~2h)

**Deliverables:**

1. `DoBestow` full port — 10 C-step flow.
2. Register `induct`, `outcast`, `bestow` in `internal/boot/boot.go` near line 600.
3. Tests.

**Tests:**

- `TestDoBestow_SyntaxEmpty` → "Bestow whom with what?\n\r".
- `TestDoBestow_VictimNotFound` → "They aren't here.\n".
- `TestDoBestow_VictimIsNPC` → "You can't give special abilities to a mob!\n".
- `TestDoBestow_VictimOutranksCh` — victim.Trust > ch.Trust → "You aren't powerful enough...\n".
- `TestDoBestow_List_NoArg` — after first arg, no second arg → "Current bestowed commands on %s: %s.\n\r" with current list.
- `TestDoBestow_List_Explicit` — second arg is "list" → same output.
- `TestDoBestow_None` — second arg is "none" → bestowments cleared, both notified.
- `TestDoBestow_Append` — second arg is "induct" → `victim.PCData.Bestowments` now contains "induct" (with C's leading-space quirk).
- `TestDoBestow_AppendMultiple` — two calls, "induct" then "outcast" → both in the list.
- `TestDoBestow_IsNameDetectsLeadingSpace` — after `DoBestow(imm, "victim induct")`, `util.IsName("induct", victim.PCData.Bestowments) == true` (verifies the quirk doesn't break `isClanOfficer`).
- `TestBoot_ClanOfficerRegistered` in `internal/boot/boot_test.go` — resolves "induct", "outcast", "bestow"; asserts DoFun non-nil and Level is 0/0/LEVEL_IMMORTAL respectively.

**Constraints:**

- `DoBestow` does NOT call `g.SavePlayer(victim)` — matches C (persistence defers to next pfile save).
- Leading-space quirk preserved verbatim.

**Mutation-verification:** change `victim.PCData.Bestowments = ""` (the "none" branch) to `victim.PCData.Bestowments = argument` → test red; revert.

**File paths:**

- `internal/act/clan_officer.go` (add DoBestow).
- `internal/act/clan_officer_test.go` (extend).
- `internal/boot/boot.go` (add three `reg.Register` lines).
- `internal/boot/boot_test.go` (add registration test).

### (Conditional) G5 — Clan-rank management (pending Q1)

**Scope (if human opts in):** minimal `DoClanSetRank(ch, argument)` taking `leader|number1|number2 <playername>`. Same gating as `isClanOfficer` but requires `clan.Leader` specifically (not just any officer) for leader changes; number1/number2 changes require leader-or-deity. ~40 Go LOC, 6 tests. Not a full `do_setclan` port — only the rank-management subset.

**If human opts out:** document in completion record that rank-management remains gated behind a future `plan-phase6-setclan.md`. Officers can still `induct`/`outcast` members; they just can't promote each other.

---

## Acceptance Criteria

Fourteen criteria. Each is mechanically verifiable.

1. **`SaveClan` round-trip.** `TestSaveLoadClan_RoundTrip` green: a fully-populated `ClanData` survives `SaveClan → loadClan` with byte-identical field values. Mutation: transpose two `Fprintf` calls → test red → revert via `Edit` → green.

2. **`SaveClan` byte-for-byte format parity.** `TestSaveClan_FormatMatchesGolden` green — emitted bytes match the golden string. This is the port-fidelity check against C's `save_clan` output; a future C↔Go cross-deploy works because clan files remain interchangeable.

3. **`loadClan` handles legacy keys.** Loading the shipped `db/clans/*.clan` files succeeds with no `log.Printf` warnings. `TestLoadClansFromDir` remains green (no regression from G0's loader extension).

4. **`isClanOfficer` 5-way gate.** All five positive cases (leader, number1, number2, deity, bestow-keyword) return true; unrelated PC returns false; case-insensitive on name compares.

5. **`DoInduct` caller gate.** NPC caller, no-clan caller, or non-officer caller each get `"Huh?\n"` and no state mutation. Matches C's terse `_("Huh?\n")` output verbatim.

6. **`DoInduct` happy path (pkill clan).** After `DoInduct(leader, "target")` where target is same-room PC, pkill-flagged, level >= 10, not immortal, not in a clan, class matches a non-guild clan: `victim.PCData.Clan == clan`, `clan.Members++`, `victim.Speaks & LANG_CLAN != 0`, `PLR_NICE` cleared, `PCFLAG_DEADLY` set, all `clan.Class`-matched guild skills awarded at the per-class adept value. Three `AT_MAGIC` messages broadcast.

7. **`DoInduct` edge-case coverage.** All 15 documented edge-case tests (G2) green. Every C-visible error path reproduced verbatim (exact string match on the trailing `\n\r` and punctuation).

8. **`DoInduct` persistence.** `persist.SaveClanFile` invoked with the modified clan after a successful induct. Verified by test stub.

9. **`DoOutcast` rank arithmetic.** Test matrix covers: officer-outranked-by-victim (blocked), superior-trust-beats-equal-rank (allowed), equal-rank-equal-trust (blocked per C's `x <= y && trust <= trust`).

10. **`DoOutcast` happy path.** Victim removed from clan — `PCData.Clan == nil`, `ClanName == ""`, `clan.Members--`, LANG_CLAN cleared from `Speaks`, `Speaking` reset to LANG_COMMON if it was LANG_CLAN. If victim was Number1/Number2, clan's rank slot blanked. Skills forgotten for non-guild/order/nokill types.

11. **`DoOutcast` broadcast branches.** Linkdead victim (`Desc == nil`) does not receive TO_VICT act-call — `util.Bug` logs instead. PKers echo fires for plain clan, suppressed for guild/order. All 3 C-visible act() calls match exact format strings.

12. **`DoBestow` append and none.** Append grows `Bestowments` with C's leading-space quirk. "none" clears the field and notifies both parties. `util.IsName` correctly parses the leading-space field — `isClanOfficer` accepts a fresh-bestowed "induct" as a valid keyword.

13. **Command registration + authority.** `induct`/`outcast` registered at `Level: 0` (authority is in-command, not dispatcher); `bestow` at `LEVEL_IMMORTAL`. `DoInduct` called by a non-officer low-level PC returns "Huh?\n" via the in-command gate, not a dispatcher "huh?" from the registry.

14. **Regression.** `go test ./...` green. `go test -count=3 ./...` green (no flakes — clan commands are deterministic given fixed skill tables and no randomness).

---

## Scope Cuts / Deferrals

**Explicitly NOT in this plan:**

- **`do_promote` / `do_demote`.** These commands do not exist in shipped SMAUG. Rank-management via `do_setclan number1 <player>` is partially covered by optional G5 (Q1). If Q1 defers, the hole is documented as a follow-up `plan-phase6-setclan.md`.
- **`do_setclan` full surface.** 30+ clan-field setters. Builder-only; not officer-facing. Separate plan.
- **`do_bestowarea`.** Immortal-areafile bestowal. Adjacent to `do_bestow` but functionally distinct. Separate 50-LOC follow-up plan.
- **`save_member_lists` / `add_member` / `remove_member` / per-clan roster file.** C maintains a separate `clans/<filename>.members` file listing every member. The in-memory `clan.Members` counter is the only field officer commands mutate; the separate list file is a display nicety (enables `claninfo` to show the roster). Leave as a `TODO G3` comment; tracked in `TODO.md` Active.
- **`add_loginmsg` for linkdead outcast victim.** C queues a message that fires on next login. Go has no loginmsg system. `DoOutcast` logs via `util.Bug` instead; the victim learns of their outcast status by seeing `PCData.Clan == nil` on next login. Follow-up.
- **`save_char_obj(victim)` inside `DoBestow`.** C does NOT call it; neither does the Go port. Bestowments persist on next normal pfile save.
- **"Your clan has no members at all" cleanup.** C does not handle the last-member-leaves case (clan stays registered with Members==0). Neither does the port.
- **Cross-clan member-list integrity audit.** `fread_clan` in C assumes member roster and clan struct agree; they can drift. Out of scope; neither port nor C patches this.
- **`DoClanJoin` stub replacement.** The existing `DoClanJoin` at `internal/act/clan.go:101` lets any PC join any clan with zero gating. It is a developer convenience for test setup and NOT a port of `do_induct`. Leave registered (as `join`); `induct` is the sanctioned clan-officer flow. `DoClanLeave` similarly stays.

---

## Open Questions

**Q1 — Port rank management in this plan (via a minimal `setrank` command) or defer?** The roadmap listed "promote/demote" as in-scope but those don't exist in C. The equivalent C surface is `do_setclan <clan> number1 <player>` / `number2` / `leader`. Three options:

1. **Include G5** — port a minimal `DoClanSetRank(ch, "number1 alice")` that wraps the leader+deity-only authority check. ~40 Go LOC, 6 tests. Keeps the clan-officer command set complete.
2. **Defer to `plan-phase6-setclan.md`** — ship only induct/outcast/bestow now; rank management waits for the full `do_setclan` port.
3. **Expose via the existing scaffolded `DoSetclan` stub if one exists.** (Verified: no such stub. Option 3 == Option 2 effectively.)

**Recommendation:** either path is defensible.

- Option 1 (include G5) closes the rank-management gap with minimal surface (~40 Go LOC, 6 tests). Authority check extends `isClanOfficer` with a leader/deity-only restriction.
- Option 2 (defer to `plan-phase6-setclan.md`) keeps this plan's scope clean. Officers can `induct`/`outcast` but not promote; a half-functional rank system ships. Fine as an interim state since rank-promotion is a rare operation.

Neither choice blocks the rest of the plan. Prefer Option 2 if Wave 2 wants tightly-scoped plans; prefer Option 1 if "clan officer commands" should ship feature-complete in one plan.

**Decision for human:** Option 1 (include G5) or Option 2 (defer).

**Q2 — `IsPkill()` method on `CharData` or inline check?** Six call sites across this plan need `(ch.PCData.Flags & int(PCFLAG_DEADLY)) != 0`. Matches C `IS_PKILL(ch)` macro. Add as `func (ch *CharData) IsPkill() bool` in `internal/types/character.go` (similar to existing `IsNPC()`, `IsImmortal()`)? Or inline?

**Recommendation:** add the method. Five other Go files already inline the expression (`mudprog/ifcheck.go:543` etc.); centralizing reduces drift risk. Small change, cleaner reads.

**Q3 — Where does `persist.SaveClanFile` get the clan-directory path from?** Options A (export `boot.ClanDir` package var) and B (store `ClanDir` on `world.World`) from §Go Design D6. Recommendation: Option A for consistency with other boot-time paths. Decision low-stakes; flag for awareness.

**Q4 — Should `DoInduct` save the victim's pfile after a successful induct?** C calls `save_char_obj(victim)`. The Go port has two paths:

- Call `g.SavePlayer(victim)` — requires a GameLoop reference the command doesn't have.
- Set a dirty flag and let the next tick save.
- Rely on `persist.SavePlayerFile(dataDir, victim)` — verify this seam exists at G2.

**Recommendation:** defer the decision to G2 implementation when we see which pattern other commands use (likely the `persist.SavePlayerFile` direct-call path matching how `DoSave` works). Not a design-level question; mechanical.

---

## Risk Analysis

**Low risk:**

- `DoBestow` — minimal logic, no cross-subsystem writes, well-bounded.
- `SaveClan` format-writer — pure text serialization, symmetric with existing loader.

**Medium risk:**

- `DoInduct` / `DoOutcast` — 14-step state machines with many branches. Test coverage per-branch mitigates; mutation-verify on each.
- Skill-award / skill-forget loops — depend on `WorldRef.Skills[sn].Guild` being populated correctly for the test clan's Class. Verify test fixture at G1 start.
- `loadClan` extension — risk of breaking existing shipped clan files. Mitigation: `TestLoadClansFromDir` regression check in G0.

**High risk:** none. The clan-officer surface is self-contained; no combat-loop or mudprog hooks touched.

**Soft-shared footprint with other Wave 2 plans:**

- Every Wave 2 plan registers commands in `internal/boot/boot.go`. Three one-line additions here; converges cleanly with other additive merges.
- No other Wave 2 plan modifies `internal/persist/subsystems.go` (`SaveClan` / `loadClan` changes) — zero conflict risk.
- No Wave 2 plan touches `internal/act/clan.go` or `clan_test.go` — `clan_officer.go` is a new file.

**Dependencies on external Wave 2 decisions:**

- If the orchestrator ships `plan-phase6-setclan.md` (a future rank-management plan) before this plan's G5, G5 becomes a no-op and should be dropped.

---

## Self-Review Notes (2026-04-18)

Adversary dispatch attempted — the manager subagent in this environment has no `Agent` tool (confirmed: tool catalog lacks `Agent`). Structured self-review substituted, matching the pattern established in `plan-phase6-marriage.md` and sibling Wave 1 plans. External adversary pass recommended before execution.

**Self-review findings (severity: minor):**

- **F1 (resolved in-plan).** Initial draft stated scope as "5 commands (promote/demote/induct/outcast/bestow)" per roadmap. `grep` verified `do_promote`/`do_demote` do not exist in SMAUG; scope corrected to 3 commands in §Problem and Q1 added for optional rank-management surface.
- **F2 (resolved in-plan).** Initial draft assumed `SaveClan` was a ~20-line deferred stub; inspection of C `save_clan:177-255` showed ~45 `fprintf` calls emitting 30+ fields, and the existing Go `loadClan` is missing 6 keys. Plan expanded G0 to a ~3.5h M-task that covers both writer and loader-extension work.
- **F3 (resolved in-plan).** Roadmap estimate of "~300 C LOC total" was low. Actual: 170 (induct) + 173 (outcast) + 57 (bestow) = 400 C LOC, plus 78 C LOC for `save_clan`. Plan's ~15h total estimate reflects this.
- **F4 (open).** The `IsImmortal()` check in `DoInduct` step 6 uses C's `IS_IMMORTAL(victim)` macro which is `get_trust(victim) >= LEVEL_IMMORTAL`. Go's `ch.IsImmortal()` helper must match that semantics exactly. Verify at G2 implementation; not a design-blocker.
- **F5 (open, captured in Q3).** `persist.SaveClanFile`'s clan-directory seam — small wiring question, deferred to G2.

**Items explicitly NOT self-reviewed (require external adversary):**

- Byte-for-byte SaveClan golden-file content (G0's `TestSaveClan_FormatMatchesGolden`). Self-review verified structural match against C; a paired-adversary run against `src/clans.c:206-250` with diff would validate exact spacing.
- Clan-type branching in `DoInduct` / `DoOutcast` — the 9 branches on `clan.ClanType × victim-type` are mechanical but easy to typo. Adversary should re-derive from `src/clans.c:981-1050` and compare.
- Test coverage completeness against C error paths — self-review found ~35 distinct `send_to_char` strings in the three C functions; plan specifies ~30 tests. Adversary should pair each test to a C line.
