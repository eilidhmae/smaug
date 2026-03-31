# Phase 2: Completed Work

This document records what has been implemented during Phase 2.

## Summary

Phase 2 goal: "A playable game with full combat, inventory management, and basic spells."

**Status**: In progress. Handler layer and attribute bonus tables complete.

**Stats**: 40 source files, 18 test files, 383 test cases — all passing.

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

#### BitVector Additions (types/bitvector.go)

- `Not()` — bitwise complement
- `AndNot(other)` — clear bits (`bv &^ other`), used for removing affect bitvector flags

**Tests**: `handler/handler_test.go` — 31 new test cases (was 18, now 49 total):
- ObjFromChar (2), ObjFromRoom (1), ObjFromObj (1), UnequipChar (2 — including dual-wield promotion)
- ExtractObj (3 — basic, with contents, from char), ExtractChar (2 — basic, with inventory)
- AffectToChar (1), AffectRemove (1), AffectStrip (1), AffectJoin (1)
- AffectModify BitVector (1), AffectModify multiple stats (1 — tests 13 apply types)
- GetCharRoom (1), GetCharWorld (1), GetObjCarry (1), GetObjWear (1), GetObjHere (1), GetObjWorld (1)

---

## Test Suite

**18 test files, 383 test cases — all passing.**

New/modified test files this phase:

| File | Tests Added | What's Covered |
|------|-------------|----------------|
| `types/attributes_test.go` | 8 | All 7 attribute tables + length validation |
| `handler/handler_test.go` | 31 | Object removal, extraction, affects, find functions |
