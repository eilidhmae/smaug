# Phase 2: Completed Work

This document records what has been implemented during Phase 2.

## Summary

Phase 2 goal: "A playable game with full combat, inventory management, and basic spells."

**Status**: In progress. 9 of 10 task groups complete. Only player persistence enhancement remains.

**Stats**: 55 source files, 27 test files, 450 test cases — all passing.

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

### 5b. Combat Polish

- **Position check in ViolenceUpdate**: Characters at `POS_INCAP` or below can no longer attack
- **XP gain on NPC kill**: `computeXP(ch, victim)` — level-difference scaled XP from victim's base XP
- **Gold in corpses**: NPC gold stored in `corpse.Value[0]` for looting
- **Bug fix**: `OneHit` was using `ch.Armor` for victim AC instead of `victim.Armor`

### 6. Communication Commands (act/comm.go)

| Command | Function | Purpose |
|---------|----------|---------|
| `tell` | `DoTell` | Private message to any character in world; sets Reply pointer |
| `reply` | `DoReply` | Reply to last tell sender |
| `yell` | `DoYell` | Message to all players in same area |
| `gossip` | `DoGossip` | Global channel message to all connected players |
| `emote` | `DoEmote` | Roleplay action visible to room |

All 5 commands registered in main.go.

### 7. Movement Enhancement (act/move.go, act/info.go)

#### Door Commands

| Command | Function | Purpose |
|---------|----------|---------|
| `open` | `DoOpen` | Open closed doors and containers; checks locked state |
| `close` | `DoClose` | Close open doors |
| `unlock` | `DoUnlock` | Unlock doors with matching key in inventory |
| `lock` | `DoLock` | Lock closed doors with matching key |

Helper functions: `findDoor(ch, arg)` (keyword/direction match), `hasKey(ch, vnum)`.

#### Position Check

`MoveChar` now requires `POS_STANDING` — sitting/resting/fighting characters get appropriate error messages instead of moving.

All 4 door commands registered in main.go.

### 8. Enhanced Information (act/info2.go)

| Command | Function | Purpose |
|---------|----------|---------|
| `consider` | `DoConsider` | Compare levels with mob (7 messages from "not worth" to "death wish") |
| `where` | `DoWhere` | Find characters in same area by name, or list all visible players |
| `time` | `DoTime` | Display game time (hour, time of day, day/month/year) |

### 9. Basic Magic System (magic/magic.go, act/magic.go)

**Spell function registry** with 12 spells:

| Spell | Type | Effect |
|-------|------|--------|
| `cure light` | Heal | 1d8 + level/3 HP |
| `cure serious` | Heal | 2d8 + level/2 HP |
| `cure critical` | Heal | 3d8 + level HP |
| `magic missile` | Damage | Level-scaled, no save |
| `fireball` | Damage | Level-scaled, save for half |
| `armor` | Buff | -20 AC for 24+level ticks |
| `bless` | Buff | +hitroll for 12+level ticks |
| `sanctuary` | Buff | AFF_SANCTUARY (damage halved) for 16+level/2 ticks |
| `curse` | Debuff | -1 hitroll + AFF_CURSE, save negates |
| `poison` | Debuff | -2 STR + AFF_POISON, save negates |
| `blindness` | Debuff | -4 hitroll + AFF_BLIND, save negates |
| `dispel magic` | Utility | Remove all affects from target, save negates |

**Cast command** (`DoCast`): spell lookup by name, mana cost check, target resolution by TAR_* type, spell function dispatch, room notification.

**Saving throws**: `SavesSpellStaff`, `SavesPoisonDeath` — formula: 50 + (victim_level - caster_level - victim_saves) * 5, clamped 5-95%.

---

## Test Suite

**27 test files, 450 test cases — all passing.**

New/modified test files this phase:

| File | Tests Added | What's Covered |
|------|-------------|----------------|
| `types/attributes_test.go` | 8 | All 7 attribute tables + length validation |
| `handler/handler_test.go` | 33 | Object removal, extraction, affects, find functions, GetEqChar, CanDropObj |
| `act/obj_test.go` | 12 | DoGet (room, container, not found, no arg), DoDrop (normal, nodrop), DoWear (armor, weapon), DoRemove, DoPut, DoGive, DoSacrifice |
| `game/update_test.go` | 10 | hitGain (NPC, standing, sleeping, poisoned), manaGain, moveGain, charUpdate regen, affect expiry, objUpdate corpse decay, mobileUpdate wander |
| `combat/combat_test.go` | 10 | StartFighting, StopFighting, OneHit, Damage (reduce/kill), MakeCorpse (basic/gold), XP gain, ViolenceUpdate (basic/incap) |
| `act/comm_test.go` | 6 | DoTell (normal/no-arg), DoReply, DoGossip, DoEmote, DoYell |
| `act/move_test.go` | 6 | DoOpen (normal/locked), DoClose, DoUnlock, DoLock, MoveChar sitting |
| `act/info2_test.go` | 6 | DoConsider (weaker/stronger/no-arg), DoWhere (found/not-found), DoTime |
| `magic/magic_test.go` | 8 | SpellRegistry, CureLight, MagicMissile, Armor, Poison, Sanctuary, SavesSpellStaff, FindSpellByName |
| `persist/player_test.go` | +3 | Position round-trip, on-disk +100 format, Style/Height/Weight round-trip |
