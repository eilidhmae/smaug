# Phase 2: Remaining Work

Phase 2 goal: "A playable game with full combat, inventory management, and basic spells."

**Status**: In progress. 9 of 10 task groups complete. See `phase2-completed.md` for work done so far.

**Prerequisites from Phase 1 (carry-over):**
- GSN variable assignment (skill name → slot number mapping)
- System data loading (optional — defaults work)

---

## Priority Order

### ~~1. Handler Layer Completion~~ — DONE

See `phase2-completed.md`.

### ~~2. Attribute Bonus Tables~~ — DONE

See `phase2-completed.md`.

### ~~3. Object Commands~~ — DONE (core commands)

See `phase2-completed.md`. Remaining: `eat`, `drink`, `examine` (lower priority).

### ~~4. Game Updates~~ — DONE

See `phase2-completed.md`. Remaining: `aggrUpdate` (aggressive mob targeting), scavenging, area reset on timer.

### ~~5. Combat System~~ — DONE (core + polish)

See `phase2-completed.md`. Remaining: dual wield attacks, murder command, wimpy, group XP, PC corpse.

### ~~6. Basic Magic~~ — DONE (12 spells)

See `phase2-completed.md`. Remaining: sleep, charm, detect spells, shield, identify (~8 more spells).

### ~~7. Movement Enhancement~~ — DONE (core)

See `phase2-completed.md`. Remaining: sector-based move costs, flying/swimming/underwater checks, pick lock.

### ~~8. Communication~~ — DONE (core)

See `phase2-completed.md`. Remaining: shout, pmote, channels system with deaf flags.

### ~~9. Enhanced Information~~ — DONE (core)

See `phase2-completed.md`. Remaining: weather, enhanced score matching C format.

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
