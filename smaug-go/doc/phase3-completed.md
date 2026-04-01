# Phase 3: Advanced Systems — Completed Work

## Summary

Phase 3 is in progress. 4 of 12 task groups complete (G1, G2, G4, G6). 51 source files, 29 test files, 527 test cases — all passing.

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
