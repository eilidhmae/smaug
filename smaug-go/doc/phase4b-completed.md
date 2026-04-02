# Phase 4b: Features — Completed Work

## Summary

Phase 4b is complete. 7 feature task groups implemented (G6–G10, G12). G11 (hotboot/copyover) moved to Phase 5 — Go's `net.Conn` model doesn't support C's fd-inheritance via `exec()`. 78 source files, 62 test files, 1,372 test cases across 13 packages — all passing.

---

## G6: OLC Set Commands (mset/oset/rset/aset/astat)

**New files:** `act/olc_set.go` (318 lines), `act/olc_area.go` (142 lines), `act/olc_set_test.go` (411 lines), `act/olc_area_test.go` (194 lines)

**Commands:**
- `mset <target> <field> <value>` — set mob properties (level, str/int/wis/dex/con/cha/lck, hp/mana/move, hitroll/damroll, gold, align, name/short/long, sex, race, class)
- `oset <target> <field> <value>` — set object properties (type, name/short/long, value0-5, weight, cost, level, flags, wearflags)
- `rset <field> <value>` — set room properties (flags, sector)
- `aset <area> <field> <value>` — set area properties (name, author, resetmsg, vnum ranges, resetfreq)
- `astat [area]` — display area statistics

**Tests:** 32 tests covering all commands, fields, error paths.

---

## G7: Quest System

**New files:** `act/quest.go` (214 lines), `act/quest_test.go` (551 lines)

**Commands:**
- `quest request` — get a random mob-slay quest from questmaster NPC (ACT_QUESTMASTER)
- `quest complete` — turn in completed quest for quest point reward
- `quest list` — show available quest rewards (gold, practices, HP, mana)
- `quest buy <reward>` — spend quest points on rewards
- `quest info` — show current quest target
- `quest time` — show countdown/cooldown
- `quest points` — show accumulated points

**Game loop integration:** `QuestUpdate` called per tick — decrements countdown, fails quest on timeout, decrements next-quest cooldown.

**Tests:** 30 tests covering all subcommands, error paths, quest generation, completion, rewards, update timer.

---

## G8: Banking System

**New files:** `act/bank.go` (95 lines), `act/bank_test.go` (208 lines)

**Commands:**
- `bank balance` — show gold bank balance
- `bank deposit <amount|all>` — deposit gold
- `bank withdraw <amount|all>` — withdraw gold

**Requires:** ACT_BANKER NPC in room. Gold-only (single currency).

**Tests:** 14 test cases covering no banker, NPC caller, each subcommand, insufficient funds, "all" keyword.

---

## G9: Ban System

**New files:** `act/ban.go` (101 lines), `act/ban_test.go` (265 lines), `persist/ban.go` (120 lines), `persist/ban_test.go` (117 lines)

**Commands:**
- `ban list` — show all active bans
- `ban site <pattern>` — add site ban (supports `*prefix` and `suffix*` wildcards)
- `ban remove <site>` — remove ban

**Functions:**
- `CheckBans(w, site)` — returns matching BanData for login checking (exact, prefix, suffix match)
- `persist.LoadBanList` / `persist.SaveBanList` — file persistence

**Tests:** 20 tests covering add/remove/list, exact/prefix/suffix matching, persistence round-trip.

---

## G10: Mob Tracking/Hunting

**New files:** `act/track.go` (151 lines), `act/track_test.go` (311 lines)

**Commands:**
- `track <target>` — BFS pathfinding to find direction toward target character

**Functions:**
- `BFSFindPath(src, target, maxDist)` — breadth-first search returning first-step direction. Uses `map[int]bool` for visited tracking (no global room flag pollution like C version).
- `HuntVictim(w, ch)` — NPC hunting behavior: BFS toward target, move one room, attack on arrival.

**Game loop integration:** `HuntVictim` called in `mobileUpdate` for NPCs with `ch.Hunting` set.

**Tests:** 15 tests covering BFS (nil rooms, already there, direct neighbor, multi-hop, max distance, disconnected, different areas, branching), DoTrack, HuntVictim.

---

## G11: Hotboot/Copyover — Moved to Phase 5

Moved to Phase 5. Go's `net.Conn` model doesn't support C's `exec()` + fd-inheritance approach (from `src/hotboot.c`). A Go-native design is needed: save player state to disk, graceful shutdown with "please reconnect" message, auto-place reconnecting players on restart. See `phases.md` Phase 5 candidates.

---

## G12: Remaining Player Commands

### Position Commands
**New files:** `act/position.go` (108 lines), `act/position_test.go` (197 lines)

- `rest`, `sit`, `stand`, `sleep`, `wake` — full state machine with fighting/mounted/AFF_SLEEP guards
- **Tests:** 27 cases (5 test functions × multiple table-driven cases)

### Skill/Spell Info
**New files:** `act/skills2.go` (150 lines), `act/skills2_test.go` (166 lines)

- `practice [skill]` — list skills or spend practice session to improve
- `skills` — list learned non-spell skills with proficiency percentages
- `spells` — list learned spells with proficiency percentages
- **Tests:** 9 tests

### Item-Based Spellcasting
**New files:** `act/itemuse.go` (199 lines), `act/itemuse_test.go` (316 lines)

- `quaff <potion>` — drink potion, trigger up to 3 spells on self, consume item
- `recite <scroll> [target]` — read scroll, cast on target (default self), consume
- `brandish` — wave held staff, cast spell on all in room, decrement charges
- `zap [target]` — use held wand, cast on target, decrement charges
- Wands/staves crumble when charges reach 0
- **Tests:** 23 tests

### Group System
**New files:** `act/group.go` (202 lines), `act/group_test.go` (460 lines)

- `follow <target>` — follow another character, `follow self` to stop
- `group [target]` — show group info or add/remove follower from group
- `order <target|all> <command>` — order followers to execute a command
- `assist <target>` — join combat alongside an ally
- **Tests:** 21 tests

### Mount System
**New files:** `act/mount.go` (56 lines), `act/mount_test.go` (159 lines)

- `mount <creature>` — mount NPC with ACT_MOUNTABLE flag, set POS_MOUNTED
- `dismount` — dismount, restore POS_STANDING
- **Tests:** 8 tests

---

## Statistics

| Metric | Phase 4a End | Phase 4b End |
|--------|-------------|-------------|
| Source files | 67 | 78 |
| Test files | 50 | 62 |
| Test cases | 1,208 | 1,372 |
| Packages | 13 | 13 |
| Registered commands | ~110 | ~137 |
| New implementation lines | — | 1,856 |
| New test lines | — | 3,355 |

### New Commands Added (27)
rest, sit, stand, sleep, wake, bank, ban, track, quest, practice, skills, spells, quaff, recite, brandish, zap, follow, group, order, assist, mount, dismount, mset, oset, rset, aset, astat
