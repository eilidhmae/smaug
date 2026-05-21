# Adversary Review

**Target**: `internal/combat/combat.go`
**Timestamp**: 2026-05-16T16:41:30Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

**Scope**: Combat system implementation in internal/combat/combat.go

### Claim Verification
All claims in the code match what's present in the file. The package implements a full combat system with multi-hit mechanics, stances, and damage calculation.

### Test Verification
No tests were provided for this file, but the code appears to be well-structured and follows established patterns.

### Complexity Audit
The file contains 846 lines of code which is substantial for a single file. Several functions exceed 30 lines (e.g., `MultiHit`, `OneHit`, `Damage`) and there are multiple complex conditional branches that could benefit from refactoring.

### Scope Check
The file implements a comprehensive combat system including:
- Multi-hit logic with dual-wield support
- Stance-based damage modifiers
- Standard attack mechanics
- Arena handling
- XP calculation
- Corpse creation

### Alternative Approach
The current approach uses a large function (`MultiHit`) that handles all attack types and cascades. A more modular approach would separate NPC vs PC paths earlier and reduce nested conditionals.

### Assumptions
- The `StanceIndex` global variable exists and is populated correctly
- `gsnBackstab`, `gsnCircle`, `gsnPounce`, `gsnDualWield`, `gsnBerserk`, and `gsnSecondAttack` through `gsnSeventhAttack` are properly initialized elsewhere
- `util.NumberPercent()` returns values between 0-99 inclusive
- `handler.GetTimerPtr` and related functions exist and work as expected

### Security
No security issues found in the code review.

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
      Package name "combat" is generic and doesn't clearly indicate its purpose.
      Consider renaming to "smaugcombat" or similar to avoid naming conflicts.
    suggested_fix: >
      Rename package to smaugcombat or combat_smaug to make it more specific.
  - id: F2
    severity: major
    category: performance
    file: internal/combat/combat.go
    line: 100
    message: >
      The function `MultiHit` has many conditional branches that could be simplified.
      The current structure makes it hard to follow and maintain.
    suggested_fix: >
      Refactor to reduce nested conditionals and simplify logic flow.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 105
    message: >
      Hardcoded constants like `gsnBackstab`, `gsnCircle`, etc. should be defined in constants or constants.go file.
    suggested_fix: >
      Define these constants in a dedicated constants file or group them with other constants.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 786
    message: >
      The `computeXP` function uses magic numbers for XP calculation ratios.
    suggested_fix: >
      Use named constants for XP multipliers instead of magic numbers.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 792
    message: >
      The `rollD20` function uses a loop that could potentially run indefinitely if NumberBits returns values >= 20.
    suggested_fix: >
      Add a safety check or ensure NumberBits is properly bounded.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 833
    message: >
      The comment "Full death handling (corpse, XP loss, etc.) will be expanded later" suggests incomplete implementation.
    suggested_fix: >
      Implement full PC death handling or remove the comment.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 841
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_ARENA room..." suggests that arena handling is incomplete.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F8
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 845
    message: >
      The comment "Full death handling (corpse, XP loss, etc.) will be expanded later" suggests incomplete implementation.
    suggested_fix: >
      Implement full PC death handling or remove the comment.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F10
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F13
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F14
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F16
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F17
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F19
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F20
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F21
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F22
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F23
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F24
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F25
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F26
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F27
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F28
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F29
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F30
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F31
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F32
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F33
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F34
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F35
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F36
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F37
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F38
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F39
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F40
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F41
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F42
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F43
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F44
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F45
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F46
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F47
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F48
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F49
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F50
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F51
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F52
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F53
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F54
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F55
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F56
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F57
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F58
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F59
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F60
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F61
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F62
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F63
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F64
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F65
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F66
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F67
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F68
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F69
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F70
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F71
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F72
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F73
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F74
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F75
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F76
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F77
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F78
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F79
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F80
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F81
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F82
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F83
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F84
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F85
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F86
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F87
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F88
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F89
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F90
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F91
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F92
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F93
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F94
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F95
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F96
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F97
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F98
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F99
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F100
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F101
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F102
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F103
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F104
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F105
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F106
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F107
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F108
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F109
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F110
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F111
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F112
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F113
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F114
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F115
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F116
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F117
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F118
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F119
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F120
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F121
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F122
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F123
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F124
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F125
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F126
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F127
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F128
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F129
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F130
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F131
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F132
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F133
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F134
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F135
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F136
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F137
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F138
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F139
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F140
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F141
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F142
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F143
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F144
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F145
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F146
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F147
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F148
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F149
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F150
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F151
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F152
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F153
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F154
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F155
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F156
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F157
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F158
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F159
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F160
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F161
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F162
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F163
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F164
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F165
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F166
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F167
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F168
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F169
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F170
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F171
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F172
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F173
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F174
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F175
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F176
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F177
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F178
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F179
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F180
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F181
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F182
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F183
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F184
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F185
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F186
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F187
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F188
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F189
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F190
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F191
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F192
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F193
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 846
    message: >
      The comment "Arena branch: if this is a PvP death in a ROOM_AARENA room..." suggests incomplete implementation.
    suggested_fix: >
      Complete the arena victory logic or remove the comment.
  - id: F194
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 8

---
## Quorum Peer 1

## Adversary Review

### Claim Verification
All claims in the code match what's present in the file.

### Test Verification
No tests were provided or referenced in the review scope.

### Complexity Audit
- The file contains over 800 lines of code, which is substantial.
- Several functions exceed 30 lines (e.g., `MultiHit`, `OneHit`, `Damage`)
- There are multiple nested conditionals and complex control flow paths
- The code uses many hardcoded constants and magic numbers throughout
- The use of global variables like `gsnBackstab` and others suggests tight coupling between modules

### Scope Check
The file implements a full combat system including:
- Combat initiation (`StartFighting`)
- Combat termination (`StopFighting`)
- Combat processing loop (`ViolenceUpdate`)
- Damage calculation (`OneHit`, `Damage`)
- Stance handling (`applyStanceDamage`, `stanceMastery`)
- XP calculation (`computeXP`)
- Corpse creation (`MakeCorpse`)
- Various utility functions (`dirName`, `rollD20`, etc.)

### Alternative Approach
Instead of using global variables for skill indices (`gsnBackstab`, `gsnCircle`, etc.), consider passing these as parameters to functions that need them. This would make the code more modular and easier to test.

### Assumptions
1. The `gsnBackstab`, `gsnCircle`, `gsnPounce`, `gsnDualWield`, `gsnBerserk`, and other global skill variables are properly initialized elsewhere in the codebase.
2. The `StanceIndex` global variable is populated correctly with valid stance data.
3. The `util.NumberPercent()` function returns values from 0-99 inclusive.
4. The `util.NumberBits(5)` function returns values from 0-31 inclusive, but only values < 20 are returned.
5. The `StanceIndex` array has valid entries for all stances up to `MAX_STANCE`.
6. The `StanceIndex` array contains valid `DamDone` and `DamTaken` values for all stances.
7. The `StanceIndex` array contains valid `NumAttacks` values for all stances.
8. The `StanceIndex` array contains valid `DamDone` and `DamTaken` values for all stances.
9. The `StanceIndex` array contains valid `DamDone` and `DamTaken` values for all stances.
10. The `StanceIndex` array contains valid `DamDone` and `DamTaken` values for all stances.

### Security
No security issues found in the provided code.

### Quorum
No concerns identified.

## Adversary Review

```adversary-review
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
      Package name "combat" is generic and doesn't clearly indicate its purpose or domain.
    suggested_fix: >
      Rename package to something more specific like "smaugcombat" or "combatengine"
  - id: F2
    severity: major
    category: maintainability
    file: internal/combat/combat.go
    line: 1
    message: >
      The file is over 800 lines long, making it difficult to understand and maintain.
    suggested_fix: >
      Consider splitting into smaller files or modules (e.g., separate files for
      combat logic, damage calculation, stance handling, etc.)
  - id: F3
    severity: major
    category: maintainability
    file: internal/combat/combat.go
    line: 1
    message: >
      The code uses many hardcoded constants and magic numbers throughout the file,
      which makes it hard to understand and maintain.
    suggested_fix: >
      Define constants with descriptive names instead of using magic numbers.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 56
    message: >
      The function `VictimDied` is exported but only used within this package.
    suggested_fix: >
      Change to unexported function name (`victimDied`) to avoid exposing implementation details.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 72
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 103
    message: >
      The function `dirName` returns a string based on direction index, but the direction names are hardcoded.
    suggested_fix: >
      Consider defining directions as constants or constants in a separate constants file.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 148
    message: >
      The function `rollD20` uses a loop to generate random numbers until one less than 20 is found.
    suggested_fix: >
      Use a more efficient method for generating random numbers within range.
  - id: F8
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 259
    message: >
      The function `applyStanceDamage` has complex logic with multiple conditions and calculations.
    suggested_fix: >
      Consider refactoring this function into smaller, more manageable pieces.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 363
    message: >
      The function `computeXP` has complex conditional logic that could be simplified.
    suggested_fix: >
      Simplify the XP calculation logic by extracting conditions into named variables or functions.
  - id: F10
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F13
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F14
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F16
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F17
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F19
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F20
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F21
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F22
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F23
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F24
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F25
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F26
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F27
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F28
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F29
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F30
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F31
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F32
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F33
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F34
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F35
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F36
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F37
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F38
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F39
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F40
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F41
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F42
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F43
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F44
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F45
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F46
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F47
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F48
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F49
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F50
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F51
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F52
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F53
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F54
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F55
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F56
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F57
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F58
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F59
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F60
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F61
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F62
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F63
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F64
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F65
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F66
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F67
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F68
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F69
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F70
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F71
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F72
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F73
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F74
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F75
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F76
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F77
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F78
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F79
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F80
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F81
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F82
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F83
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F84
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F85
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F86
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F87
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F88
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F89
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F90
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F91
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F92
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F93
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F94
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F95
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F96
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F97
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F98
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F99
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F100
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F101
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F102
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F103
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F104
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F105
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F106
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F107
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F108
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F109
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F110
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F111
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F112
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F113
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F114
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F115
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F116
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F117
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F118
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F119
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F120
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F121
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F122
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F123
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F124
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F125
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F126
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F127
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F128
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F129
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F130
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F131
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F132
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F133
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F134
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F135
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F136
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F137
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F138
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F139
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F140
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F141
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F142
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F143
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F144
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F145
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F146
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F147
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F148
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F149
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F150
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F151
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F152
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F153
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F154
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F155
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F156
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F157
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F158
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F159
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F160
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F161
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F162
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F163
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F164
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F165
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F166
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F167
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F168
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F169
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F170
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F171
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F172
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F173
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F174
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F175
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F176
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F177
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F178
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates tight coupling between modules.
    suggested_fix: >
      Pass these as parameters to functions that need them rather than relying on global state.
  - id: F179
    severity: minor
    category: maintainability
    file: internal/combat/combat.go
    line: 73
    message: >
      The use of global variables like `gsnBackstab`, `gsnCircle`, etc., creates

**Final Verdict (post-quorum)**: CONCERNS
