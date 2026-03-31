# Phase 3: Advanced Systems — Implementation Plan

## Context

Phase 2 is complete (55 source files, 27 test files, 455 tests). Phase 3 brings feature parity for world interaction and building: MUD Progs, OLC, immortal commands, 100+ spells, shops, clans, quests, boards, pager, and protocol support. The C source for these systems totals ~64K lines across 14 files.

## Task Groups (execution order)

### G1: Pager + String Editor
Unlocks stat commands, OLC, and any long-output command.
- `internal/net/pager.go` — page long output with `[Hit Return to Continue]`
- `internal/game/editor.go` — line-based text editor for OLC descriptions
- Modify `game/loop.go` nanny for `CON_PAGING` and `CON_EDITING` states
- ~18 functions, 2 new files, 2 test files

### G2: Phase 2 Deferred Items
Completes the core gameplay layer.
- **Commands:** eat, drink, fill, empty, examine, shout, pmote
- **Combat:** dual wield attacks, wimpy, murder, sector move costs, fly/swim checks, pick lock
- **Updates:** aggrUpdate, NPC scavenging, area reset on timer
- **Info:** enhanced score, weather
- **Spells (8):** sleep, charm, detect evil/invis/magic/hidden, shield, identify
- Modify existing act/combat/magic/game files, ~30 functions

### G3: Subsystem Data Loaders
Populates world state for clans, councils, deities, boards, socials.
- `persist/clans.go`, `persist/councils.go`, `persist/deities.go`, `persist/boards.go`, `persist/socials.go`
- Social dispatch fallback in command interpreter
- Wire into bootDB in main.go
- ~20 functions, 5 new files, 5 test files

### G4: Stat Commands (mstat, ostat, rstat)
Essential immortal inspection commands. Heavy pager users.
- `act/wiz.go` — mstat, ostat, rstat, astat
- ~6 functions, 1 new file, 1 test file
- **Depends on:** G1 (pager)

### G5: Immortal Commands — Core Set
Movement, manipulation, and control commands for admins/builders.
- **Movement:** goto, transfer, at, bamfin/bamfout
- **Action:** force, peace, purge, restore, advance, slay
- **Info:** mfind, ofind, mwhere, owhere, users, wizhelp, invis, holylight
- **Control:** snoop, switch/return, freeze, silence, wizlock
- **System:** echo/recho/aecho, shutdown/reboot
- Extend `act/wiz.go`, ~35 functions
- **Depends on:** G4

### G6: Shops and Repair Shops
Player-facing economy system.
- `act/shop.go` — buy, sell, list, value, repair, estimate
- Helpers: FindKeeper, GetCost with profit margins
- ~10 functions, 1 new file, 1 test file

### G7: Remaining Spells and Skills
Bulk content — generic spell engine + combat/utility skills.
- **Generic engine:** SpellSmaug (data-driven dispatcher), SpellAffect, SpellAffectChar
- **Named spells (~10):** locate object, portal, create food/water, summon, teleport, recall, enchant weapon/armor, gate
- **Combat skills (~15):** backstab, bash, kick, disarm, rescue, circle, gouge, stun, trip, berserk, hitall, sneak, hide, steal, pick
- **Utility skills (~8):** recall, scan, aid, mount/dismount, brew, scribe, cook
- **Skill improvement on use**
- `magic/spells.go`, `act/skills.go`, ~50 functions

### G8: OLC — Online Creation
Builder interface for rooms, mobs, objects.
- `act/olc_room.go` — redit (name, desc, flags, sector, exits, extra descs, progs)
- `act/olc_obj.go` — oedit (type, values, flags, affects, descs, progs)
- `act/olc_mob.go` — medit (stats, flags, attacks, defenses, progs)
- `act/olc_area.go` — aset, ocreate, mcreate, rdig, rlist/olist/mlist
- ~55 functions, 4 new files
- **Depends on:** G1 (string editor, pager), G4, G5 (goto)

### G9: Clans, Councils, Deities, Boards/Notes
Player social/organizational subsystems.
- `act/clan.go` — clantalk, clans, claninfo, join/leave/promote/demote (~12 functions)
- `act/council.go` — council commands (~6 functions)
- `act/deity.go` — devote, deities, supplicate, favor tracking (~8 functions)
- `act/board.go` — note read/list/write/post/remove (~10 functions)
- **Depends on:** G3 (data loaders)

### G10: MUD Programs
Scripting system for interactive NPCs, objects, rooms.
- `mudprog/driver.go` — interpreter with if/or/else/endif, nested ifs
- `mudprog/ifcheck.go` — ~40 if-check types (rand, level, hp, class, race, etc.)
- `mudprog/translate.go` — variable substitution ($n, $t, $o, $p, etc.)
- `mudprog/triggers.go` — 25+ trigger types (act, greet, speech, fight, death, give, etc.)
- `mudprog/commands.go` — mpgoto, mptransfer, mpforce, mpecho, mpdamage, etc.
- `mudprog/sleep.go` — delayed execution timers
- Wire triggers into existing act/combat/game code
- ~60 functions, 6 new files, 3 test files
- **Depends on:** G5 (goto/transfer), G7 (damage commands)

### G11: Protocol Support (MCCP2, MSDP, MSSP)
Network layer enhancements — independent of game logic.
- `net/mccp.go` — zlib compression via compress/zlib
- `net/msdp.go` — variable reporting (health, mana, room, etc.)
- `net/mssp.go` — server status for MUD listing sites
- `net/telnet.go` — telnet negotiation state machine
- ~22 functions, 4 new files, 3 test files

### G12: OLC Area Save
Persist builder changes to .are files.
- `persist/area_write.go` — write rooms, mobs, objects, resets, shops, mudprogs
- savearea/foldarea commands
- Round-trip test: load .are → save → reload → compare
- ~15 functions, 1 new file, 1 test file
- **Depends on:** G8

## Summary

| Group | Functions | New Files | Depends On |
|-------|-----------|-----------|------------|
| G1 Pager+Editor | ~18 | 2+2 test | — |
| G2 Deferred | ~30 | 1+3 test | — |
| G3 Loaders | ~20 | 5+5 test | — |
| G4 Stat Cmds | ~6 | 1+1 test | G1 |
| G5 Imm Cmds | ~35 | extend G4 | G4 |
| G6 Shops | ~10 | 1+1 test | — |
| G7 Spells/Skills | ~50 | 2+2 test | — |
| G8 OLC | ~55 | 4+1 test | G1,G4,G5 |
| G9 Clans/Boards | ~36 | 4+4 test | G3 |
| G10 MUD Progs | ~60 | 6+3 test | G5,G7 |
| G11 Protocols | ~22 | 4+3 test | — |
| G12 Area Save | ~15 | 1+1 test | G8 |
| **Total** | **~357** | **~31+26 test** | |

## Verification

After each group: `go test ./...` — all tests pass. After all groups: telnet in, builder can create areas with OLC, add mob progs, test them; players can use shops, join clans, interact with all game systems.
