# Phase 2: Completed Work

This document records what has been implemented during Phase 2.

## Summary

Phase 2 goal: "A playable game with full combat, inventory management, and basic spells."

**Status**: In progress. Handler layer, attribute tables, object commands, game updates, and combat system complete.

**Stats**: 46 source files, 21 test files, 414 test cases — all passing.

---

## Completed Tasks

### 1. Attribute Bonus Tables (types/attributes.go)

Ported all 7 attribute bonus tables from C `const.c`, indexed by stat value 0-25:

| Table | Fields | Usage |
|-------|--------|-------|
| `StrApp[26]` | ToHit, ToDam, Carry, Wield | Combat hit/dam bonuses, carry capacity, weapon weight limits |
| `IntApp[26]` | Learn | Spell learning rate |
| `WisApp[26]` | Practice | Practice session gain |
| `DexApp[26]` | Defensive | AC bonus from dexterity |
| `ConApp[26]` | Hitp, Shock | HP gain per level, shock survival % |
| `ChaApp[26]` | Charm | Charm spell resistance modifier |
| `LckApp[26]` | Luck | Luck modifier for various checks |

**Tests**: `types/attributes_test.go` — 8 test functions covering all 7 tables with spot-checks at key indices (0, mid-range, max) plus table length validation.

### 2. Handler Layer Completion (handler/handler.go, handler/find.go)

#### Object Removal Functions

| Function | Signature | Purpose |
|----------|-----------|---------|
| `ObjFromChar` | `(obj)` | Remove object from character inventory; unequips first if worn |
| `ObjFromRoom` | `(obj)` | Remove object from room contents |
| `ObjFromObj` | `(obj)` | Remove object from container |
| `UnequipChar` | `(ch, obj)` | Unequip item; handles dual-wield→wield promotion |

#### Extraction Functions

| Function | Signature | Purpose |
|----------|-----------|---------|
| `ExtractObj` | `(w, obj)` | Full object removal — removes from location, recursively extracts contents, removes from world, decrements template count |
| `ExtractChar` | `(w, ch, fPull)` | Full character removal — stops combat, extracts inventory, removes from room/world, clears reply/retell refs, disconnects |

#### Affect Management Functions

| Function | Signature | Purpose |
|----------|-----------|---------|
| `AffectModify` | `(ch, aff, fAdd)` | Apply/reverse stat modifications for an affect (STR/DEX/INT/WIS/CON/CHA/LCK, AC, hitroll, damroll, HP/mana/move, saves, bitvector flags) |
| `AffectToChar` | `(ch, aff)` | Add new affect to character (copies affect, applies modifiers) |
| `AffectRemove` | `(ch, aff)` | Remove specific affect (reverses modifiers, removes from list) |
| `AffectStrip` | `(ch, sn)` | Remove all affects of a given skill type |
| `AffectJoin` | `(ch, aff)` | Combine with existing affect of same type (duration + modifier stacking with caps) |

#### Find Functions (handler/find.go)

| Function | Signature | Purpose |
|----------|-----------|---------|
| `GetCharRoom` | `(ch, arg)` | Find character in room by name; supports `N.name` syntax, "self", two-phase search (exact then prefix) |
| `GetCharWorld` | `(w, ch, arg)` | Find character in world; checks room first, then world list |
| `GetObjCarry` | `(ch, arg)` | Find object in inventory (not equipped) |
| `GetObjWear` | `(ch, arg)` | Find object in equipment (equipped only) |
| `GetObjHere` | `(ch, arg)` | Find object in room, inventory, or equipment (precedence order) |
| `GetObjWorld` | `(w, ch, arg)` | Find object anywhere; checks local first, then world list |

#### Additional Handler Helpers

| Function | Signature | Purpose |
|----------|-----------|---------|
| `GetEqChar` | `(ch, wearLoc)` | Get equipped item at specific wear location |
| `CanDropObj` | `(obj)` | Check if object can be dropped (ITEM_NODROP) |
| `GetObjList` | `(list, arg)` | Find object in a list by name (public, used by DoGet for containers) |

### 3. Object Commands (act/obj.go)

| Command | Function | Purpose |
|---------|----------|---------|
| `get` | `DoGet` | Pick up objects from room or containers; supports "get item", "get item container" |
| `drop` | `DoDrop` | Drop objects to room; respects ITEM_NODROP |
| `put` | `DoPut` | Put objects in containers; checks container type |
| `give` | `DoGive` | Give objects to characters in room |
| `wear` | `DoWear` | Wear/wield/hold objects; auto-detects wear location from flags, auto-removes conflicting equipment |
| `remove` | `DoRemove` | Remove worn equipment; respects ITEM_NOREMOVE |
| `sacrifice` | `DoSacrifice` | Destroy room objects for 1 gold |

Helper functions:
- `findWearLoc(obj)` — maps object wear flags to wear location constants
- `wearVerb(wearLoc)` — returns "wear"/"wield"/"hold"/"light" based on location

All 7 commands registered in `cmd/smaug/main.go`.

#### BitVector Additions (types/bitvector.go)

- `Not()` — bitwise complement
- `AndNot(other)` — clear bits (`bv &^ other`), used for removing affect bitvector flags

**Tests**: `handler/handler_test.go` — 33 new test cases (was 18, now 51 total):
- ObjFromChar (2), ObjFromRoom (1), ObjFromObj (1), UnequipChar (2 — including dual-wield promotion)
- ExtractObj (3 — basic, with contents, from char), ExtractChar (2 — basic, with inventory)
- AffectToChar (1), AffectRemove (1), AffectStrip (1), AffectJoin (1)
- AffectModify BitVector (1), AffectModify multiple stats (1 — tests 13 apply types)
- GetCharRoom (1), GetCharWorld (1), GetObjCarry (1), GetObjWear (1), GetObjHere (1), GetObjWorld (1)

### 4. Game Updates (game/update.go)

| Function | Frequency | Purpose |
|----------|-----------|---------|
| `charUpdate()` | Per tick (~70s) | HP/mana/move regeneration, affect duration countdown with wear-off messages, poison tick damage |
| `objUpdate()` | Per tick (~70s) | Object timer countdown, corpse decay with room messages |
| `mobileUpdate()` | Per mobile pulse (~4s) | NPC random wandering (respects sentinel, stay-area, closed doors, no-mob rooms) |
| `violenceUpdate()` | Per violence pulse (~3s) | Delegates to `combat.ViolenceUpdate` for per-pulse combat rounds |
| `areaUpdate()` | Per area pulse (~60s) | Stub for periodic area resets |

Regeneration formulas ported from C `update.c`:
- `hitGain(ch)` — NPC: level*3/2; Player: base 5 + position bonus (sleeping: con*2, resting: con*1), halved by hunger/thirst, quartered by poison
- `manaGain(ch)` — NPC: level; Player: base 5 + position bonus (sleeping: int*3, resting: int*2)
- `moveGain(ch)` — NPC: level; Player: base max(15, 2*level) + position bonus (sleeping: dex*4, resting: dex*2)

`MoveChar` made public in `act` package for NPC wandering.

### 5. Combat System (combat/combat.go, act/combat.go)

| Function | Package | Purpose |
|----------|---------|---------|
| `StartFighting` | combat | Initiate combat, set FightData, position to POS_FIGHTING |
| `StopFighting` | combat | End combat, clear FightData; fBoth=true stops both sides |
| `ViolenceUpdate` | combat | Per-pulse combat loop: iterate fighting chars, call OneHit (+ multi-attack for NPCs) |
| `OneHit` | combat | Single attack: thac0 calc, d20 roll vs AC, damage dice + bonuses, sanctuary halving |
| `Damage` | combat | Apply damage, send messages, update position, handle death (corpse for NPCs, reset for PCs) |
| `MakeCorpse` | combat | Create corpse object, transfer inventory, place in room |
| `DoKill` | act | Player 'kill' command — target NPC, initiate combat |
| `DoFlee` | act | Player 'flee' command — try 8 random exits, stop fighting on success |

Combat formulas:
- **Thac0**: NPC uses `MobThac0`; Player uses `20 - level` (simplified)
- **Hit roll**: d20, roll 0 = always miss, roll 19 = always hit, else must exceed thac0 - victimAC
- **Damage**: weapon dice or bare-hand dice + damroll + str bonus, position multipliers, sleeping = 2x
- **Death**: HP <= -MaxHP → POS_DEAD → corpse creation (NPC) or reset to 1 HP (PC)
- **Corpse timer**: NPC = 6 ticks, PC = 40 ticks

Both `kill` and `flee` commands registered in main.go.

---

## Test Suite

**21 test files, 414 test cases — all passing.**

New/modified test files this phase:

| File | Tests Added | What's Covered |
|------|-------------|----------------|
| `types/attributes_test.go` | 8 | All 7 attribute tables + length validation |
| `handler/handler_test.go` | 33 | Object removal, extraction, affects, find functions, GetEqChar, CanDropObj |
| `act/obj_test.go` | 12 | DoGet (room, container, not found, no arg), DoDrop (normal, nodrop), DoWear (armor, weapon), DoRemove, DoPut, DoGive, DoSacrifice |
| `game/update_test.go` | 10 | hitGain (NPC, standing, sleeping, poisoned), manaGain, moveGain, charUpdate regen, affect expiry, objUpdate corpse decay, mobileUpdate wander |
| `combat/combat_test.go` | 7 | StartFighting, StopFighting, OneHit, Damage (reduces HP, kills victim), MakeCorpse, ViolenceUpdate |
| `persist/player_test.go` | +3 | Position round-trip, on-disk +100 format, Style/Height/Weight round-trip |
