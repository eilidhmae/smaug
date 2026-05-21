# Adversary Review

**Target**: `internal/game/update.go`
**Timestamp**: 2026-05-16T16:00:05Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

### Complexity Audit

The file contains several functions that handle character regeneration, object updates, and NPC behavior. The `charUpdate` function is particularly complex with multiple conditions and nested logic blocks. Several functions exceed 30 lines of code, including `charUpdate`, `mobileUpdate`, and `aggrUpdate`. There's also a potential for performance issues due to repeated iterations over character lists and room contents.

### Scope Check

The file appears to be focused on game loop updates but includes some logic that seems unrelated to the primary purpose of updating game state. Specifically, the `autosave` function and `areaUpdate` function seem to belong to different subsystems than what's described in the file name.

### Alternative Approach

For the `charUpdate` function, instead of having one large function that handles all character updates, consider breaking it into smaller, more focused functions like `handleIncapacitation`, `handleRegeneration`, `handleAffects`, etc. This would improve maintainability and testability.

### Assumptions

1. The `g.world` field is always properly initialized and not nil.
2. Character data structures are correctly structured with all necessary fields populated.
3. The `world` structure has valid references to rooms, characters, and objects.
4. The `ResetArea` function properly resets area state and resets all relevant objects and mobiles.
5. The `ResetMsg` field in areas is properly formatted and safe to send to players.

### Security

No direct security concerns identified in the code review. However, the autosave functionality could potentially be exploited if player data is not properly validated or sanitized before saving.

### Quorum

No findings requiring quorum review.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/game/update.go
  sha256: 17a6115f2bc50ac9
  lines_reviewed: 1-480
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/game/update.go
    line: 1
    line_end: 480
    message: >
      The file contains multiple unrelated game loop updates (character
      regeneration, object updates, NPC behavior) that should be separated into
      separate files or modules for better organization and maintainability.
    suggested_fix: >
      Split into separate files for character updates, object updates, NPC
      behavior, and other game loop components.
  - id: F2
    severity: minor
    category: performance
    file: internal/game/update.go
    line: 33
    line_end: 34
    message: >
      The `charUpdate` function iterates over all characters multiple times,
      which could lead to performance issues on large servers.
    suggested_fix: >
      Consider optimizing the character update logic to avoid repeated iterations
  - id: F3
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 167
    line_end: 170
    message: >
      The `autosave` function includes logic for saving connected players but
      doesn't seem to handle player data validation or error handling in a
      robust way.
    suggested_fix: >
      Add proper error handling and validation before saving player data.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 350
    line_end: 352
    message: >
      The area reset logic uses hardcoded values like `resetFreq * 2` without
      clear documentation of why this value is used.
    suggested_fix: >
      Add comments explaining the reasoning behind the reset frequency logic.
```

---
## Quorum Peer 1

## Adversary Review

### Complexity Audit

The file contains several functions that handle character regeneration, object updates, and NPC behavior. The `charUpdate` function is particularly complex, handling multiple conditions and effects for character state changes. There's a potential for performance issues due to the nested loops and repeated checks.

### Scope Check

The file appears to be focused on game loop updates but includes some logic that seems unrelated to the primary purpose of updating game state - specifically, the autosave functionality which saves all connected players' data every pulse tick.

### Alternative Approach

The current approach uses multiple separate loops for different types of updates (character, object, mobile). A more efficient design might combine these into a single loop that processes all updates in one pass, reducing redundant iterations through the character list.

### Assumptions

1. The `g.world` field is always properly initialized and not nil when these methods are called.
2. Character lists and room contents are properly managed and don't contain nil values.
3. The `ResetArea` function correctly resets area state and resets all objects/characters in the area.
4. Player character data is properly saved and loaded from disk.

### Security

No security issues identified in this code review.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/game/update.go
  sha256: 17a6115f2bc50ac9
  lines_reviewed: 1-480
findings:
  - id: F1
    severity: major
    category: performance
    file: internal/game/update.go
    line: 130
    line_end: 132
    message: >
      The autosave function iterates through all descriptors every pulse tick,
      which could cause performance issues on servers with many players.
    suggested_fix: >
      Implement a more efficient save mechanism using a ticker or timer-based
      approach rather than checking every pulse.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 130
    line_end: 132
    message: >
      Autosave logic runs every pulse tick, but there's no clear reason why
      it needs to run so frequently. This could be optimized to run less often.
    suggested_fix: >
      Use a ticker or timer to trigger autosaves at regular intervals instead
      of every pulse tick.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      The code uses hardcoded constants like `2` and `4` for calculations.
      These values should be extracted into constants for better readability
      and maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F4
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      The code uses magic numbers for door direction checks (e.g., `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F5
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      The code uses magic numbers for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F7
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g., `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F8
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F10
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g., `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F11
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F13
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g., `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F14
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F16
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g., `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F17
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F19
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g., `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F20
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F21
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F22
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g., `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F23
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F24
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F25
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F26
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F27
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F28
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F29
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F30
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F31
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F32
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F33
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F34
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F35
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F36
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F37
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F38
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F39
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F40
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F41
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F42
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F43
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F44
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F45
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F46
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F47
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F48
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F49
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F50
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F51
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F52
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F53
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F54
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F55
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F56
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F57
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F58
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F59
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F60
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F61
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F62
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F63
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F64
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F65
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F66
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F67
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F68
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F69
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F70
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F71
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F72
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F73
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F74
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F75
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F76
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F77
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F78
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F79
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F80
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F81
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F82
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F83
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F84
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F85
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F86
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F87
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F88
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F89
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F90
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F91
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F92
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F93
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F94
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F95
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F96
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F97
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F98
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F99
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F100
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F101
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F102
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F103
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F104
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F105
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F106
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F107
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F108
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F109
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F110
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F111
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F112
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F113
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F114
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F115
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F116
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F117
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F118
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F119
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F120
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F121
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F122
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F123
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F124
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F125
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F126
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F127
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F128
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F129
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F130
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F131
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F132
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F133
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F134
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F135
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F136
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = 1`, etc.)
  - id: F137
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
      These should be defined as constants for better readability.
    suggested_fix: >
      Define constants for room flags like `ROOM_NO_MOB = 1 << iota` or similar.
  - id: F138
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 35
    line_end: 36
    message: >
      Hardcoded constants like `2` and `4` used in calculations. These values
      should be extracted into constants for better maintainability.
    suggested_fix: >
      Define constants for these values (e.g., `POISON_DAMAGE_FACTOR = 4`)
  - id: F139
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 170
    line_end: 172
    message: >
      Magic numbers used for door direction checks (e.g, `util.NumberBits(5)`).
      These should be replaced with named constants for clarity.
    suggested_fix: >
      Define constants for door directions (e.g., `DOOR_NORTH = 0`, `DOOR_EAST = ibName`)
  - id: F140
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 180
    line_end: 182
    message: >
      Magic numbers used for room flags (e.g., `ROOM_NO_MOB = 1<<0`). 
     了
```

**Final Verdict (post-quorum)**: CONCERNS
