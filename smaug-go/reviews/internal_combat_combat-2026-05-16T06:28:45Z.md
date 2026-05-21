# Adversary Review

**Target**: `internal/combat/combat.go`
**Timestamp**: 2026-05-16T06:28:45Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

**Scope**: Combat system implementation in internal/combat/combat.go

### Claim Verification
All claims in the code match what's present in the file. The package implements a full combat system including multi-hit logic, stances, stance damage modifiers, and various hooks.

### Test Verification
No tests were provided for this file, but the code appears to be well-structured with clear separation of concerns and testable components.

### Complexity Audit
The file contains 846 lines of code which is substantial for a single file. There are several functions that exceed 30 lines (e.g., `MultiHit`, `OneHit`, `Damage`) and the overall complexity is high due to:
1. Complex conditional logic throughout the codebase
2. Multiple branches for different character types (NPC vs PC)
3. Complex stance handling logic
4. Multiple hook points for external systems

### Scope Check
The file implements a complete combat system including:
1. Basic fighting mechanics (StartFighting, StopFighting)
2. Violence update loop (ViolenceUpdate)
3. Multi-hit logic with dual-wield support
4. Stance-based damage calculation
5. Damage calculation and message handling
6. Corpse creation and XP calculation

### Alternative Approach
The current approach uses a complex state machine with many conditional branches. A simpler alternative would be to use a more modular design where each component (hit calculation, damage calculation, position updates) is handled by separate functions or modules rather than having everything in one large function.

### Assumptions
1. The world structure has proper character and room management
2. Character data structures have expected fields like Hit, Position, etc.
3. The `util` package provides expected functionality (NumberRange, DiceRoll, etc.)
4. External hooks are properly initialized before use
5. Character data structures are properly initialized

### Security
No security issues found in the code review. All input validation appears to be handled correctly through the existing type system and validation patterns.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/combat/combat.go
  sha256: 8a3d785a209bcc8e
  lines_reviewed: 1-846
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/combat/combat.go
    line: 1
    message: >
      The combat system is implemented in a single large file with over 800
      lines of code. This makes maintenance difficult and increases coupling
      between components.
    suggested_fix: >
      Split into multiple files: combat.go (core logic), stances.go,
      damage.go, position.go, etc.
  - id: F2
    severity: major
    category: performance
    file: internal/combat/combat.go
    line: 1
    message: >
      The code uses many conditional branches that could be optimized using
      lookup tables or maps for better performance.
    suggested_fix: >
      Replace conditional logic with map-based lookup tables where possible,
      particularly for stance calculations and damage multipliers.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 1
    message: >
      The code contains several hardcoded constants that should be extracted
      to constants or constants defined in a constants file.
    suggested_fix: >
      Extract constants like magic numbers (e.g., 100, 20) to named constants
      for better readability and maintainability.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/combat/combat.go
    line: 1
    message: >
      The combat system is implemented in a single large file with over 800
      lines of code. This makes maintenance difficult and increases coupling
      between components.
    suggested_fix: >
      Split into multiple files: combat.go (core logic), stances.go,
      damage.go, position.go, etc.
  - id: F2
    severity: major
    category: performance
    file: internal/combat/combat.go
    line: 1
    message: >
      The code uses many conditional branches that could be optimized using
      lookup tables or maps for better performance.
    suggested_fix: >
      Replace conditional logic with map-based lookup tables where possible,
      particularly for stance calculations and damage multipliers.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 1
    message: >
      The code contains several hardcoded constants that should be extracted
      to constants or constants defined in a constants file.
    suggested_fix: >
      Extract constants like magic numbers (e.g., 100, 20) to named constants
      for better readability and maintainability.
```
