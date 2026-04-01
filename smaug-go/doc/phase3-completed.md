# Phase 3: Advanced Systems — Completed Work

## Summary

Phase 3 is complete. All 12 of 12 task groups done (G1–G12). 67 source files, 36 test files, 572 test cases across 13 packages — all passing.

## G1: Pager + String Editor ✓

### Pager (`game/pager.go`, `game/pager_test.go`)

Output paging system for long text. When a player has the pager enabled (`PCFLAG_PAGERON`), text sent via `SendToPager` is buffered and displayed one page at a time.

- **`SendToPager(ch, text)`** — routes to pager buffer if enabled, else normal output
- **`WriteToPager(d, text)`** — appends to descriptor's pager buffer
- **`PagerOutput(d)`** — displays next page, returns true when done
- **`SetPagerInput(d, cmd)`** — stores user's paging command
- Commands: Continue (default), Non-stop (dump all), Refresh, Back, Quit
- Minimum page length clamped to 5 lines
- 12 tests covering all pager commands and edge cases

### String Editor (`game/editor.go`, `game/editor_test.go`)

Line-based text editor for OLC descriptions, notes, and help files.

- **`StartEditing(ch, text)`** — initializes editor, sets `CON_EDITING`
- **`StopEditing(ch)`** — cleans up, returns to `CON_PLAYING`
- **`EditBuffer(ch, line)`** — processes editor input (commands or text append)
- **`CopyBuffer(ch)`** — converts buffer back to string
- Editor commands: `/l` list, `/c` clear, `/d` delete, `/g` goto, `/i` insert, `/r` replace, `/f` format (word wrap), `/a` abort, `/s` save, `/?` help
- 24 lines normal limit, 48 for mudprogs/help; 79 char max line length
- 18 tests covering all commands and edge cases

### Descriptor Pager Methods (`types/descriptor.go`)

Added to `DescriptorData`: `WriteToPager`, `HasPagerData`, `GetPagerData`, `GetPagePoint`, `SetPagePoint`, `GetPagerCmd`, `SetPagerCmd`, `ClearPager`.

### Game Loop Integration (`game/loop.go`)

- `processInput`: pager input takes priority; `CON_EDITING` dispatches to `EditBuffer`
- `flushOutput`: calls `PagerOutput` when pager data is pending before normal flush

### Pager Command (`act/info2.go`)

`DoPager` — toggle pager on/off, set page length (5-200).

---

## G2: Phase 2 Deferred Items ✓

### Consume Commands (`act/consume.go`, `act/consume_test.go`)

- **`DoEat`** — eat food/pills; modifies `COND_FULL`, handles poison, extracts consumed item
- **`DoDrink`** — drink from containers or fountains; modifies `COND_THIRST`/`COND_DRUNK`/`COND_FULL`, handles poison
- **`DoFill`** — fill drink container from fountain in room
- **`DoEmpty`** — empty drink container
- **`GainCondition`** — condition modifier with bounds checking; ignores NPCs and immortals; starvation/dehydration warnings
- 18 tests covering conditions, eat/drink/fill/empty, edge cases

### Communication Commands (`act/comm.go`)

- **`DoShout`** — global message to all players (same as gossip but thematic)
- **`DoPmote`** — possessive emote (`Gandalf's eyes glow brightly`)

### Examine Command (`act/info.go`)

- **`DoExamine`** — performs look + shows container contents or drink level

### Combat Additions (`act/combat.go`, `combat/combat.go`)

- **`DoMurder`** — attack player characters (safe room check)
- **`DoWimpy`** — set/show auto-flee HP threshold
- **Dual wield extra attack** — in `ViolenceUpdate`, if `WEAR_DUAL_WIELD` equipped, fires an extra `OneHit`
- **Wimpy auto-flee** — in `ViolenceUpdate`, if HP <= wimpy, auto-flee to first available exit

### Update Functions (`game/update.go`, `game/update_test.go`)

- **`aggrUpdate`** — aggressive NPCs (`ACT_AGGRESSIVE`) attack players in room; skips safe rooms, immortals, charmed mobs
- **NPC scavenging** — mobs with `ACT_SCAVENGER` pick up the most valuable takeable item in the room (25% chance per pulse)
- **Area reset on timer** — areas reset when `age >= resetFrequency` (no players) or `age >= resetFrequency*2` (forced); sends `ResetMsg` to players in area
- 4 tests for aggrUpdate and area reset timer

### Enhanced Score (`act/info.go`)

Expanded `DoScore` with: race/class names, sex name, alignment descriptor text, title, kill/death stats, item/weight carried, wimpy setting, active affects with duration.

### Weather Command (`act/info.go`)

- **`DoWeather`** — shows sky conditions based on `WeatherInfo.Sky` and time of day; blocks use indoors
- Added `WeatherInfoData` struct to `types/system.go`; `WeatherInfo` field on `world.World`

### Spells (8 new) (`magic/magic.go`, `magic/magic_test.go`)

| Spell | Effect |
|-------|--------|
| `SpellSleep` | AFF_SLEEP, forces POS_SLEEPING (save negates) |
| `SpellCharmPerson` | AFF_CHARM, sets master/leader (save negates, NPC only) |
| `SpellDetectEvil` | AFF_DETECT_EVIL for level+10 duration |
| `SpellDetectInvis` | AFF_DETECT_INVIS for level+10 duration |
| `SpellDetectMagic` | AFF_DETECT_MAGIC for level+10 duration |
| `SpellDetectHidden` | AFF_DETECT_HIDDEN for level+10 duration |
| `SpellShield` | -20 AC buff for 18+level/2 duration |
| `SpellIdentify` | Reveals item type, level, weight, cost, values, affects |

All registered in `spellRegistry`. 10 new tests.

---

## G4: Stat Commands ✓

### Immortal Inspection (`act/wiz.go`)

- **`DoMstat`** — detailed mob/character stats: identity, vitals, attributes, combat stats, room, vnum, act/affect flags, affect list, carrying, fighting target. Pager-aware via `sendToPager`.
- **`DoOstat`** — detailed object stats: identity, vnum, type, level, weight, cost, wear, extra flags, values, location (carried/room/container), affects, contents.
- **`DoRstat`** — detailed room stats: vnum, sector, light, room flags, description, area, exits (direction/dest/key/flags/keyword), people, objects, extra descriptions.

All require `LEVEL_IMMORTAL` trust.

---

## G6: Shops ✓

### Shop Commands (`act/shop.go`, `act/shop_test.go`)

- **`DoBuy`** — purchase item from shopkeeper; checks gold, transfers item
- **`DoSell`** — sell item to shopkeeper; checks shop buy-type restrictions
- **`DoList`** — display items for sale with prices
- **`DoValue`** — show what shopkeeper would pay for an item
- **`findKeeper`** — finds NPC with `Shop` data in player's room
- **`getShopCost`** — buy price = `GoldCost * ProfitBuy / 100`
- **`getSellPrice`** — sell price = `GoldCost * ProfitSell / 100`
- **`shopBuysType`** — checks if shop's `BuyType` array includes the item type
- 9 tests covering buy/sell success, insufficient gold, no keeper, type restrictions, price calculations

---

## New Constants Added

- `MAX_COND_VAL = 48` — maximum condition value for hunger/thirst/drunk

## Registered Commands (16 new)

| Command | Function | Position | Level |
|---------|----------|----------|-------|
| eat | DoEat | POS_RESTING | 0 |
| drink | DoDrink | POS_RESTING | 0 |
| fill | DoFill | POS_RESTING | 0 |
| empty | DoEmpty | POS_RESTING | 0 |
| examine | DoExamine | POS_RESTING | 0 |
| shout | DoShout | POS_RESTING | 0 |
| pmote | DoPmote | POS_RESTING | 0 |
| pager | DoPager | POS_DEAD | 0 |
| weather | DoWeather | POS_RESTING | 0 |
| murder | DoMurder | POS_FIGHTING | 0 |
| wimpy | DoWimpy | POS_DEAD | 0 |
| buy | DoBuy | POS_STANDING | 0 |
| sell | DoSell | POS_STANDING | 0 |
| list | DoList | POS_STANDING | 0 |
| value | DoValue | POS_STANDING | 0 |
| mstat | DoMstat | POS_DEAD | LEVEL_IMMORTAL |
| ostat | DoOstat | POS_DEAD | LEVEL_IMMORTAL |
| rstat | DoRstat | POS_DEAD | LEVEL_IMMORTAL |

---

## G3: Subsystem Data Loaders ✓

### Loaders (`persist/subsystems.go`, `persist/subsystems_test.go`)

- **`LoadClansFromDir`** — loads clan files from `db/clans/` via `clan.lst` index. Parses `#CLAN` sections with key-value pairs (tilde-terminated strings, plain integers). 8 clans loaded.
- **`LoadDeitiesFromDir`** — loads deity files from `db/deity/` via `deity.lst` index. Parses `#DEITY` sections. 2 deities loaded.
- **`LoadSocials`** — loads social commands from `db/system/en/socials.dat`. Parses `#SOCIAL` sections with Name/CharNoArg/OthersNoArg/CharFound/OthersFound/VictFound/CharAuto/OthersAuto fields. 496 socials loaded.
- **`LoadBoards`** — loads board definitions from `db/boards/boards.dat`. Line-based key-value parsing. 2 boards loaded.
- All loaders wired into `bootDB` in main.go.
- 4 tests verify loading from real data files.

### Social Dispatch (`act/social.go`, `command/interpret.go`)

- **`CheckSocial`** — fallback function called when no command matches. Searches world social list by name/prefix, handles no-arg/targeted/self-targeted variants.
- **`socialSub`** — variable substitution for social messages ($n, $N, $e, $m, $s, $E, $M, $S).
- `SocialFallback` function pointer added to `command.Registry`.

---

## G5: Immortal Commands — Core Set ✓

### Movement (`act/wiz.go`)

- **`DoGoto`** — teleport to room vnum or character. Uses `teleportTo` with bamf messages.
- **`DoTransfer`** — move a character to you. Supports "all" for all online players.
- **`DoAt`** — execute a command at a remote location (room vnum or character), then return.
- **`DoBamfin`/`DoBamfout`** — set custom arrival/departure messages.

### Action

- **`DoForce`** — compel a character to execute a command. Trust level check prevents forcing higher-level immortals. Supports "all".
- **`DoPeace`** — stop all combat in the room, set everyone to standing.
- **`DoPurge`** — remove all NPCs/objects from room, or a specific NPC/object by name.
- **`DoRestore`** — restore HP/mana/move to max for a character or "all".
- **`DoAdvance`** — set a player's level (trust level capped).
- **`DoSlay`** — instant kill. NPC creates corpse; PC resets to 1 HP.

### Info

- **`DoMfind`/`DoOfind`** — search mob/object templates by name. Pager-aware.
- **`DoMwhere`/`DoOwhere`** — find mob/object instances in the world with location. Pager-aware.
- **`DoUsers`** — show all connected descriptors with state, name, host.

### Control

- **`DoInvis`** — toggle immortal invisibility (PLR_WIZINVIS) with optional level argument.
- **`DoHolylight`** — toggle see-all mode (PLR_HOLYLIGHT).
- **`DoFreeze`** — toggle PLR_FREEZE on a player (prevents all actions).
- **`DoSilence`** — toggle PLR_SILENCE on a player (blocks channels).
- **`DoSnoop`** (`act/clan.go`) — watch another player's output. Self-targeted cancels all snoops.

### System

- **`DoEcho`** — send message to all connected players.
- **`DoRecho`** — send message to current room.

`CmdRegistry` variable wired from main.go for force/at command dispatch.

---

## G7: Remaining Spells and Skills ✓

### Combat Skills (`act/skills.go`, `act/skills_test.go`)

| Skill | Effect |
|-------|--------|
| backstab | Weapon damage × level/10 multiplier; requires wield, out of combat |
| bash | 1-level damage, knocks victim to sitting, 2×PULSE_VIOLENCE stun |
| kick | 1-level damage while in combat (func `DoKickSkill`) |
| disarm | Knocks weapon from victim's grasp to ground |
| rescue | Swaps fighting targets to protect an ally |
| sneak | AFF_SNEAK affect for level duration |
| hide | Sets AFF_HIDE directly on character |
| steal | Steal gold (10%) or items from victim inventory |
| pick | Pick lock on doors (clears EX_LOCKED); checks EX_PICKPROOF |
| scan | Look into adjacent rooms, show characters |
| aid | Restore incapacitated character to 1 HP |
| recall | Teleport to temple (ROOM_VNUM_TEMPLE), costs half move |

### Skill System

- **`canUseSkill`** — NPC: 85% base success. PC: roll vs `Learned[gsn]`.
- **`learnFromSuccess`** — chance to gain +2 skill points, capped at 100.
- **`learnFromFailure`** — chance to gain +1 skill point, capped at 99.

### Named Spells (12 new, `magic/magic.go`)

| Spell | Effect |
|-------|--------|
| locate object | Find first matching object in world |
| create food | Create magical mushroom (level/2 nutrition) |
| create water | Fill drink container with water |
| summon | Teleport victim to caster |
| teleport | Teleport caster to random room |
| enchant weapon | +1+(level/20) hitroll/damroll, sets ITEM_MAGIC |
| enchant armor | -(1+level/20) AC, sets ITEM_MAGIC |
| invis | AFF_INVISIBLE for 24+level duration |
| fly | AFF_FLYING for level+10 duration |
| heal | Heals max(100, level×5) HP |

16 tests for skills, 10 for new spells.

---

## G8: OLC — Online Creation ✓

### Room Editing (`act/olc.go`)

- **`DoRedit`** — edit current room: name, desc (launches string editor via `StartEditingFunc`), sector, flags (toggle), exdesc (add), exit (create/delete/set destination).

### Creation Commands

- **`DoOcreate`** — create new object template by vnum, add to `ObjIndex`, create instance in inventory.
- **`DoMcreate`** — create new mob template by vnum, add to `MobIndex`, create instance in room.
- **`DoRdig`** — create new room in a direction with bidirectional exit linking. Auto-assigns vnum if not specified.

### Listing Commands

- **`DoRlist`/`DoOlist`/`DoMlist`** — list rooms/objects/mobs in a vnum range. Defaults to current area's range. Pager-aware.

`StartEditingFunc` variable breaks circular dependency between act and game packages.

---

## G9: Clans, Councils, Deities, Boards/Notes ✓

### Clan Commands (`act/clan.go`)

- **`DoClans`** — list all clans with leader and member count.
- **`DoClanInfo`** — show detailed clan info (leader, officers, stats, motto, description).
- **`DoClantalk`** — clan-only chat channel.
- **`DoClanJoin`** — join a clan (sets PCData.Clan, increments member count).
- **`DoClanLeave`** — leave your clan.

### Deity Commands

- **`DoDeities`** — list all deities with worshipper count and alignment.
- **`DoDevote`** — devote to a deity or renounce. Tracks worshipper count.

### Board/Note Commands

- **`DoNote`** — subcommands: list, read <#>, write, post, remove <#>. Notes stored on board.

### Additional

- **`DoSnoop`** — immortal command to watch another player's output.

---

## G10: MUD Programs ✓

### Interpreter (`mudprog/driver.go`, `mudprog/driver_test.go`)

- **`Driver`** — executes a mudprog script line by line. Maintains if-level stack (MAX_IFS=20) with state tracking (execute/skip-if/skip-else). Handles if/or/else/endif/break. Prevents infinite recursion (maxProgNest=3). Translates variables before executing each command. Falls through to normal command interpreter for non-mp commands.

### If-Checks (`mudprog/ifcheck.go`, `mudprog/ifcheck_test.go`)

- **`DoIfCheck`** — evaluates if-check expressions. Parses "checkname($var) op value" format.
- ~25 check types: rand, ispc, isnpc, isevil, isgood, isneutral, isimmort, isfight, ischarmed, isflying, isinvis, isaffected, level, hp, hppcnt, mana, gold, sex, position, class, race, alignment, str/int/wis/dex/con, name, isinroom.
- Integer comparisons (==, !=, >, <, >=, <=) and string matching (==, !=, /, !/).
- 7 tests covering rand, ispc, level, isevil, hppcnt, isfight, name.

### Variable Substitution (`mudprog/translate.go`, `mudprog/translate_test.go`)

- **`Translate`** — replaces $i/$I (mob), $n/$N (actor), $t/$T (victim), $r (random), $o/$O (object), $p/$P (target), $e/$m/$s (actor pronouns), $E/$M/$S (victim pronouns), $j/$k/$l (mob pronouns).
- 3 tests for basic substitution, nil safety, object variables.

### Commands (`mudprog/commands.go`)

- **mpecho** — message to mob's room.
- **mpechoat** — message to specific character.
- **mpechoaround** — message to room except target.
- **mpgoto** — move mob to room by vnum.
- **mptransfer** — move character to mob's room (or specified vnum).
- **mpforce** — force character to execute a command.
- **mpkill** — start mob fighting a character.
- **mpdamage** — deal damage to a character (non-lethal).
- **mppurge** — remove NPCs/objects from room.

### Triggers (`mudprog/triggers.go`)

- **`MobTrigger`** — checks prog type mask, matches trigger argument, calls Driver.
- **`triggerMatches`** — type-specific matching: rand (percent), speech (keyword), act (substring), bribe (gold amount), give (object keyword).
- Convenience functions: `TrigGreet`, `TrigEntry`, `TrigSpeech`, `TrigFight`, `TrigDeath`, `TrigRand`, `TrigGive`.

---

## G11: Protocol Support ✓

### Telnet (`net/telnet.go`)

- Constants: IAC, DO/DONT/WILL/WONT, SB/SE, option codes (ECHO, TTYPE, NAWS, COMPRESS2, MSDP, MSSP).
- **`TelnetNeg`** — build 3-byte negotiation sequence.
- **`TelnetSubneg`** — build subnegotiation sequence (IAC SB option data IAC SE).

### MCCP2 (`net/mccp.go`, tested)

- **`CompressData`** — zlib compress via `compress/zlib`.
- **`DecompressData`** — zlib decompress.
- **`MCCPStartSequence`** — returns IAC SB COMPRESS2 IAC SE.

### MSDP (`net/msdp.go`, tested)

- **`BuildMSDPReport`** — builds MSDP subnegotiation with VAR/VAL pairs.
- **`MSDPCharReport`** — standard character status variables (CHARACTER_NAME, HEALTH, MANA, MOVEMENT).

### MSSP (`net/mssp.go`, tested)

- **`BuildMSSPPayload`** — builds MSSP subnegotiation with server info (NAME, PLAYERS, UPTIME, CODEBASE, PORT, etc.).

8 tests covering telnet negotiation, MSSP payload, MSDP report, compression round-trip.

---

## G12: OLC Area Save ✓

### Area Writer (`persist/area_write.go`, `persist/area_write_test.go`)

- **`SaveArea`** — writes a complete area to an io.Writer in SMAUG .are format. Sections: #AREA, #AUTHOR, #RANGES, #RESETMSG, #FLAGS, #ECONOMY, #VERSION, #MOBILES, #OBJECTS, #ROOMS, #RESETS, #SHOPS, #$.
- **`saveMobiles`** — writes mob templates in area's vnum range (sorted).
- **`saveObjects`** — writes object templates with extra descriptions, affects, mudprogs.
- **`saveRooms`** — writes rooms with exits (D sections), extra descs, mudprogs, S terminators.
- **`saveResets`** — writes reset commands.
- **`saveShops`** — writes shop data for keeper mobs.

### Save Command (`act/olc.go`)

- **`DoSaveArea`** — saves the current room's area to `db/area/<filename>`.

2 tests: basic area with room/mob/obj/reset (verifies all sections present), empty area.
