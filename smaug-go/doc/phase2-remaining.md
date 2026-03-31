# Phase 2: Remaining Work

Phase 2 goal: "A playable game with full combat, inventory management, and basic spells."

**Status**: In progress. Handler layer and attribute tables complete. See `phase2-completed.md` for work done so far.

**Prerequisites from Phase 1 (carry-over):**
- GSN variable assignment (skill name → slot number mapping)
- System data loading (optional — defaults work)

---

## Priority Order

### ~~1. Handler Layer Completion~~ — DONE

See `phase2-completed.md`.

### ~~2. Attribute Bonus Tables~~ — DONE

See `phase2-completed.md`.

### 3. Object Commands (HIGH — item interaction)

**`act/obj.go`:**
- `get` — get from room, get from container, get all
- `drop` — drop to room
- `put` — put in container
- `give` — give to character
- `sacrifice` — destroy for gold
- `wear` — wear/wield/hold based on wear flags
- `remove` — remove worn item
- `eat`, `drink` — consume food/drink
- `examine` — detailed look at object

### 4. Game Updates (HIGH — world feels alive)

**`game/update.go`:**
- `charUpdate()` — HP/mana/move regeneration per tick, hunger/thirst drain, poison tick, affect duration countdown with wear-off messages
- `objUpdate()` — item timers, corpse decay
- `mobileUpdate()` — NPC wandering, scavenging
- `aggrUpdate()` — aggressive mob targeting
- Area reset on timer (respawn mobs/objects)

### 5. Combat System (HIGH — core gameplay)

**`combat/`:**
- `violenceUpdate()` — per-pulse combat round
- `oneHit(ch, victim, dt)` — single attack: hit roll vs AC, damage calculation, damage types
- `damage(ch, victim, dam, dt)` — apply damage, handle death
- Multi-attack (NumAttacks field)
- Dual wield support
- `kill`/`murder` commands
- Flee and wimpy
- Death handling: create corpse, transfer items, experience loss, gold drop

### 6. Basic Magic (MEDIUM — first ~20 spells)

**`magic/`:**
- `cast` command — mana check, target resolution, spell function dispatch
- Saving throws
- Spell function registry (map spell_fun names → Go functions)
- Initial spells: cure light/serious/critical, magic missile, fireball, sanctuary, blindness, poison, sleep, charm, detect evil/invis/magic/hidden, armor, shield, bless, curse, dispel magic, identify

### 7. Movement Enhancement (MEDIUM)

**`act/move.go`:**
- Sector-based movement costs (move points)
- Door commands: open, close, lock, unlock, pick
- Flying/swimming/underwater checks
- Position checks (can't walk while sitting)

### 8. Communication (MEDIUM)

**`act/comm.go`:**
- `tell`, `reply` — private messaging
- `yell`, `shout`, `gossip` — area/global channels
- `emote`, `pmote` — roleplay
- Channels system with deaf flags

### 9. Enhanced Information (LOW)

Expand `act/info.go`:
- `consider` — compare levels with mob
- `where` — find characters in area
- `time`, `weather` — game time display
- Enhanced `score` matching C format

### 10. Player Persistence Enhancement (LOW)

- Save inventory and equipment (object sections in player file)
- Save active affects
- Periodic autosave (every N ticks)

---

## Suggested Implementation Sequence

1. **Handler completion** + **attribute tables** — foundation for everything
2. **Object commands** (get/drop/wear/remove) — players can interact with items
3. **Game updates** (regen, timers, mob AI) — world feels alive
4. **Combat** (oneHit, damage, death) — core gameplay loop
5. **Magic** (cast, first 20 spells) — depth
6. **Movement/communication/info** — polish

Each step should follow TDD with fixture-based tests.
