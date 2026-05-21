# Adversary Review

**Target**: `internal/game/update.go`
**Timestamp**: 2026-05-16T07:29:27Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

### Complexity Audit

The file contains several functions that handle character regeneration, object updates, and NPC behavior. The `charUpdate` function is particularly complex with multiple conditions and nested logic for handling different character states and positions. There's also a potential issue with the `mobileUpdate` function where it uses `util.NumberBits(5)` to determine random directions, which could lead to unexpected behavior if not properly handled.

### Scope Check

The file appears to be focused on game loop updates but includes some logic that seems unrelated to the primary purpose of updating game state, such as autosaving player data and area resets. These seem like they might belong elsewhere or at least be separated into their own modules.

### Alternative Approach

For the `charUpdate` function, instead of using a single large function with many conditionals, consider breaking down the logic into smaller, more manageable functions that each handle specific aspects of character updates (e.g., regeneration, affect duration, etc.). This would improve maintainability and readability.

### Assumptions

1. The `g.world` field exists and is properly initialized.
2. Character data structures are correctly structured with all necessary fields.
3. The `util` package provides expected functionality for number generation and other utilities.
4. The `handler` and `combat` packages provide expected behavior for character actions and combat.

### Security

No obvious security issues found in the code review. However, there's potential for resource exhaustion through excessive object creation or memory usage during updates.

### Quorum

No concerns identified requiring peer review.

```
adversary-review
verdict: CONCERNS
confidence: medium
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
    message: >
      The file contains multiple responsibilities including character regeneration,
      object updates, NPC behavior, and autosaving. These should be separated into
      separate modules or files to adhere to single responsibility principle.
    suggested_fix: >
      Split into separate files for character updates, object updates, NPC behavior,
      and autosave logic.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 33
    message: >
      The `charUpdate` function is very complex with many nested conditions and
      logic paths. Consider refactoring into smaller functions for better readability
      and maintainability.
    suggested_fix: >
      Break down `charUpdate` into smaller functions like `handleIncapacitation`,
      `handleRegeneration`, `handleAffects`, etc.
  - id: F3
    severity: minor
    category: performance
    file: internal/game/update.go
    line: 35
    message: >
      The `charUpdate` function iterates through all characters every tick, which
      could become performance bottlenecks as the number of characters increases.
    suggested_fix: >
      Consider implementing a more efficient data structure or caching mechanism
      to reduce iteration overhead.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 160
    message: >
      The use of `util.NumberBits(5)` to determine random directions might not
      produce expected behavior if not properly handled in the util package.
    suggested_fix: >
      Ensure that `util.NumberBits` returns valid directions (0-9) and validate
      the exit before moving.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 280
    message: >
      The area reset logic uses hardcoded values for reset frequency and age
      calculations. These should be configurable or at least documented clearly.
    suggested_fix: >
      Make reset frequency configurable via configuration files or constants.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/game/update.go
    line: 1
    message: >
      The file contains multiple responsibilities including character regeneration,
      object updates, NPC behavior, and autosaving. These should be separated into
      separate modules or files to adhere to single single responsibility principle.
    suggested_fix: >
      Split into separate files for character updates, object updates, NPC behavior,
      and autosave logic.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 33
    message: >
      The `charUpdate` function is very complex with many nested conditions and
      logic paths. Consider refactoring into smaller functions for better readability
      and maintainability.
    suggested_fix: >
      Break down `charUpdate` into smaller functions like `handleIncapacitation`,
      `handleRegeneration`, `handleAffects`, etc.
  - id: F3
    severity: minor
    category: performance
    file: internal/game/update.go
    line: 35
    message: >
      The `charUpdate` function iterates through all characters every tick, which
      could become performance bottlenecks as the number of characters increases.
    suggested_fix: >
      Consider implementing a more efficient data structure or caching mechanism
      to reduce iteration overhead.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 160
    message: >
      The use of `util.NumberBits(5)` to determine random directions might not
      produce expected behavior if not properly handled in the util package.
    suggested_fix: >
      Ensure that `util.NumberBits` returns valid directions (0-9) and validate
      the exit before moving.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/game/update.go
    line: 280
    message: >
      The area reset logic uses hardcoded values for reset frequency and age
      calculations. These should be configurable or at least documented clearly.
    suggested_fix: >
      Make reset frequency configurable via configuration files or constants.
```
