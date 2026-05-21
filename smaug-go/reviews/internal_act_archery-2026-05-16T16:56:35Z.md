# Adversary Review

**Target**: `internal/act/archery.go`
**Timestamp**: 2026-05-16T16:56:35Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

### Claim Verification
All claims in the file are verified. The port matches the C source code structure and preserves known bugs as specified.

### Test Verification
No tests were added or modified in this file. The file contains no new test cases or test functions.

### Complexity Audit
- **File size**: 894 lines (new code added)
- **Function size**: 
  - `rangedAttack` (lines 367-455): 88 lines
  - `projectileHit` (lines 500-635): 135 lines
  - `lodgeProjectile` (lines 637-642): 6 lines
  - `scanForVictim` (lines 644-675): 32 lines
  - `archeryRollD20` (lines 677-680): 4 lines
- **Abstraction depth**: Moderate, with helper functions like `archeryCanSeeObj`, `archeryCanSee`, and `archeryRollD20` providing abstraction layers.
- **New dependencies**: None; all dependencies are already present in the project.
- **Premature generalization**: The use of `archeryRollD20` function variable is a good pattern for testability.

### Scope Check
The file implements archery functionality as specified in the plan. No additional features were added beyond what was requested.

### Alternative Approach
The approach of porting directly from C is appropriate for this phase of development, but an alternative would be to refactor the code to reduce complexity by breaking down large functions into smaller units or using more idiomatic Go patterns.

### Assumptions
- The `WorldRef` global variable exists and is properly initialized.
- The `util.NumberRange` and `util.NumberBits` functions work as expected.
- The `handler` package provides expected functions like `GetEqChar`, `GetCharRoom`, etc.
- The `types` package contains expected constants and structures.

### Security
No security issues found in the code.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/act/archery.go
  sha256: 34dd43e4810e89e6
  lines_reviewed: 1-894
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/archery.go
    line: 137
    line_end: 139
    message: >
      The `victim == ch` check at line 137 is structurally dead because victim is
      declared as nil at line 135 and never assigned a value before the check.
    suggested_fix: >
      Remove the check or assign a value to victim before the check.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/archery.go
    line: 142
    line_end: 144
    message: >
      The comment about preserving C bug should be updated to reflect that this
      bug has been preserved in the port.
    suggested_fix: >
      Update the comment to clarify that this is a known bug in the port.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/archery.go
    line: 203
    line_end: 205
    message: >
      The comment mentions preservation of a C bug but doesn't specify which one.
    suggested_fix: >
      Clarify which specific bug is being preserved.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/act/archery.go
    line: 677
    line_end: 680
    message: >
      The archeryRollD20 function uses an infinite loop with a break condition,
      which could potentially cause performance issues or infinite loops if
      NumberBits returns values >= 20.
    suggested_fix: >
      Consider using a more predictable approach for generating random numbers.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/act/archery.go
    line: 682
    message: >
      The dirName function has hardcoded direction names, but there's no validation
      that the direction index is valid.
    suggested_fix: >
      Add validation to ensure direction index is within bounds before accessing
      the names array.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/archery.go
    line: 137
    line_end: 139
    message: >
      The `victim == ch` check at line 137 is structurally dead because victim is
      declared as nil at line 135 and never assigned a value before the check.
    suggested_fix: >
      Remove the check or assign a value to victim before the check.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/archery.go
    line: 142
    line_end: 144
    message: >
      The comment mentions preservation of a C bug but doesn't specify which one.
    suggested_fix: >
      Clarify which specific bug is being preserved.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/archery.go
    line: 203
    line_end: 205
    message: >
      The comment mentions preservation of a C bug but doesn't specify which one.
    suggested_fix: >
      Clarify which specific bug is being preserved.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/act/archery.go
    line: 677
    line_end: 680
    message: >
      The archeryRollD20 function uses an infinite loop with a break condition,
      which could potentially cause performance issues or infinite loops if
      NumberBits returns values >= 20.
    suggested_fix: >
      Consider using a more predictable approach for generating random numbers.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/act/archery.go
    line: 682
    message: >
      The dirName function has hardcoded direction names, but there's no validation
      that the direction index is valid.
    suggested_fix: >
      Add validation to ensure direction index is within bounds before accessing
      the names array.
```

---
## Quorum Peer 1

## Adversary Review

### Claim Verification
All claims in the file are verified by the code content.

### Test Verification
No tests were added or modified as part of this change.

### Complexity Audit
- The file contains over 800 lines of code, which exceeds the threshold for complexity review.
- Several functions exceed 30 lines (e.g., `rangedAttack`, `projectileHit`, `scanForVictim`)
- There's significant abstraction depth with helper functions like `archeryRollD20`, `scanForVictim`, and `lodgeProjectile`
- The code uses multiple conditional branches and complex logic flow that could benefit from simplification

### Scope Check
The code implements archery functionality including:
- DoDraw - drawing projectiles from quivers
- DoDislodge - removing lodged projectiles
- DoFire - firing ranged weapons
- Supporting functions like FindQuiver, FindProjectile, etc.

### Alternative Approach
A simpler alternative would be to reduce the number of conditional branches and simplify the logic flow in `rangedAttack` and `projectileHit`. The current implementation has many nested conditionals that make it harder to follow and maintain.

### Assumptions
1. The `WorldRef` global variable is properly initialized elsewhere in the system
2. The `util.NumberBits(5)` function returns values between 0-19 inclusive
3. The `handler` package provides expected behavior for object manipulation functions
4. The `combinat` package provides expected combat behavior

### Security
No security issues found in this file.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/act/archery.go
  sha256: 34dd43e4810e89e6
  lines_reviewed: 1-894
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/archery.go
    line: 1
    message: >
      File contains over 800 lines of code which exceeds the typical limit for maintainability.
    suggested_fix: >
      Consider splitting into smaller files or modules for better organization.
  - id: F2
    severity: major
    category: maintainability
    file: internal/act/archery.go
    line: 73
    message: >
      Function `scanForVictim` is over 70 lines and contains complex conditional logic that could be simplified.
    suggested_fix: >
      Refactor to reduce nesting and simplify conditional branches.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/archery.go
    line: 73
    message: >
      The function `archeryRollD20` uses a loop with a break condition that may not be intuitive.
    suggested_fix: >
      Consider using a simpler random number generation approach or document the intent more clearly.
```

**Final Verdict (post-quorum)**: CONCERNS
