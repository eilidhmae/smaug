# Adversary Review

**Target**: `internal/act/info2.go`
**Timestamp**: 2026-05-16T19:14:54Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

### Claim Verification
All claims in the file match what's present in the code.

### Test Verification
No tests were provided for these functions.

### Complexity Audit
- `DoConsider` function is 18 lines long and has a clear switch statement with 6 cases.
- `DoWhere` function is 23 lines long and uses a loop with conditionals.
- `DoTime` function is 24 lines long and includes a switch statement.
- `DoPager` function is 33 lines long and includes multiple conditionals and switch cases.

### Scope Check
The file contains four functions that implement commands related to character interaction and game information display. All functions are within scope of the act package and follow the pattern of handling user input and sending responses to characters.

### Alternative Approach
For `DoWhere`, instead of using a nested loop to find matching characters, we could pre-filter the character list by area or use a map lookup for better performance if the character list grows large.

### Assumptions
- The `WorldRef` global variable exists and is properly initialized.
- Character data structures have expected fields like `Level`, `Name`, `InRoom`, etc.
- The `util` package provides `OneArgument` and `IsName` functions as expected.
- `GetCharRoom` returns a valid character pointer or nil.

### Security
No security issues found in the code.

```
adversary-review
verdict: CONCERNS
confidence: medium
artifact:
  path: internal/act/info2.go
  sha256: 9318bfcba1b9afda
  lines_reviewed: 1-175
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/info2.go
    line: 10
    line_end: 10
    message: >
      Hardcoded string "Consider killing whom?" is used without proper
      internationalization support or configuration options.
    suggested_fix: >
      Consider using a constants file or configuration for such messages.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/info2.go
    line: 14
    line_end: 14
    message: >
      Hardcoded string "They aren't here." is used without proper
      internationalization support or configuration options.
    suggested_fix: >
      Consider using a constants file or configuration for such messages.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/info2.go
    line: 17
    line_end: 17
    message: >
      Hardcoded string "You think about yourself for a moment." is used
      without proper internationalization support or configuration options.
    suggested_fix: >
      Consider using a constants file or configuration for such messages.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/act/info2.go
    line: 25
    line_end: 25
    message: >
      Hardcoded string "Not worth the effort." is used without proper
      internationalization support or configuration options.
    suggested_fix: >
      Consider using a constants file or configuration for such messages.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/act/info2.go
    line: 29
    line_end: 29
    message: >
      Hardcoded string "Should be no contest." is used without proper
      internationalization support or configuration options.
    suggested_fix: >
      Consider using a constants file or configuration for such messages.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/act/info2.go
    line: 33
    line_end: 33
    message: >
      Hardcoded string "Easy." is used without proper internationalization
      support or configuration options.
    suggested_fix: >
      Consider using a constants file or configuration for such messages.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/act/info2.go
    line: 37
    line_end: 37
    message: >
      Hardcoded string "A fair fight." is used without proper
      internationalization support or configuration options.
    suggested_fix: >
      Consider using a constants file or configuration for such messages.
  - id: F8
    severity: minor
    category: maintainability
    file: internal/act/info2.go
    line: 41
    line_end: 41
    message: >
      Hardcoded string "You would need some luck." is used without proper
      internationalization support or configuration options.
    suggested_fix: >
      Consider using a constants file or configuration for such messages.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/act/info2.go
    line: 45
    line_end: 45
    message: >
      Hardcoded string "You would need a lot of luck!" is used without proper
      internationalization support or configuration options.
    suggested_fix: >
      Consider using a constants file or configuration for such messages.
  - id: F10
    severity: minor
    category: maintainability
    file: internal/act/info2.go
    line: 49
    line_end: 49
    message: >
      Hardcoded string "Do you have a death wish?" is used without proper
      internationalization support or configuration options.
    suggested_fix: >
      Consider using a constants file or configuration for such messages.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/act/info2.go
    line: 56
    line_end: 56
    message: >
      Hardcoded string "Players near you:" is used without proper
      internationalization support or configuration options.
    suggested_fix: >
      Consider using a constants file or configuration for such messages.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/act/info2.go
    line: 63
    line_end: 63
    message: >
      Hardcoded string "No one." is used without proper internationalization
      support or configuration options.
    suggested_fix: >
      Consider using a constants file or configuration for such messages.
  - id: F13
    severity: minor
    category: maintainability
    file: internal/act/info2.go
    line: 70
    line_end: 70
    message: >
      Hardcoded string "No one by that name around here." is used without
      proper internationalization support or configuration options.
    suggested_fix: >
      Consider using a constants file or configuration for such messages.
  - id: F14
    severity: minor
    category: maintainability
    file: internal/act/info2.go
    line: 85
    line_end: 85
    message: >
      Hardcoded string "Time is meaningless." is used without proper
      internationalization support or configuration options.
    suggested_fix: >
      Consider using a constants file or configuration for such messages.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/act/info2.go
    line: 96
    line_end: 96
    message: >
      Hardcoded string "Pager disabled." is used without proper
      internationalization support or configuration options.
    suggested_fix: >
      Consider using a constants file or configuration for such messages.
  - id: F16
    severity: minor
    category: maintainability
    file: internal/act/info2.go
    line: 99
    line_end: 99
    message: >
      Hardcoded string "Pager enabled (%d lines)." is used without proper
      internationalization support or configuration options.
    suggested_fix: >
      Consider using a constants file or configuration for such messages.
  - id: F17
    severity: minor
    category: maintainability
    file: internal/act/info2.go
    line: 104
    line_end: 104
    message: >
      Hardcoded string "Usage: pager [on|off|<5-200>]" is used without proper
      internationalization support or configuration options.
    suggested_fix: >
      Consider using a constants file or configuration for such messages.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/act/info2.go
    line: 107
    line_end: 107
    message: >
      Hardcoded string "Pager set to %d lines." is used without proper
      internationalization support or configuration options.
    suggested_fix: >
      Consider using a constants file or configuration for such messages.
  - id: F19
    severity: minor
    category: maintainability
    file: internal/act/info2.go
    line: 113
    line_end: 113
    message: >
      Hardcoded string "It is hour %d of the day, %s." is used without proper
      internationalization support or configuration options.
    suggested_fix: >
      Consider using a constants file or configuration for such messages.
  - id: F20
    severity: minor
    category: maintainability
    file: internal/act/info2.go
    line: 116
    line_end: 116
    message: >
      Hardcoded string "Day %d of month %d, year %d." is used without proper
      internationalization support or configuration options.
    suggested_fix: >
      Consider using a constants file or configuration for such messages.
  - id: F21
    severity: minor
    category: maintainability
    file: internal/act/info2.go
    line: 118
    line_end: 118
    message: >
      Hardcoded string "It's a holiday today: %s" is used without proper
      internationalization support or configuration options.
    suggested_fix: >
      Consider using a constants file or configuration for such messages.
```

---
## Quorum Peer 1

## Adversary Review

### Claim Verification
All claims in the code match what's present in the file.

### Test Verification
No tests were provided for these functions, but the logic appears straightforward.

### Complexity Audit
The code is relatively simple and readable, with no overly complex structures or abstractions.

### Scope Check
All functions are within scope of their intended functionality.

### Alternative Approach
For the `DoConsider` function, a more maintainable approach would be to extract the switch cases into constants or constants with descriptive names rather than magic numbers.

### Assumptions
- The `WorldRef` global variable exists and is properly initialized.
- `ch.IsNPC()` and `ch.IsPC()` methods exist and work as expected.
- `ch.PCData` is never nil for PC characters.
- `ch.PCData.PagerLen` has a default value when not explicitly set.

### Security
No security issues found in this code segment.

```
adversary-review
verdict: PASS
confidence: high
artifact:
  path: internal/act/info2.go
  sha256: 9318bfcba1b9afda
  lines_reviewed: 1-175
findings: []
```

---
## Quorum Peer 2

## Adversary Review

### Complexity Audit
The code is relatively straightforward but has some areas of concern:
- `DoWhere` function has a nested loop that could be optimized with a map lookup for better performance
- The switch statements in `DoConsider` and `DoTime` are well structured but could benefit from constants or constants-like constants for better maintainability
- `DoPager` has some redundant checks and logic that could be simplified

### Scope Check
No scope creep detected - all functions implement their intended functionality correctly.

### Alternative Approach
For `DoWhere`, instead of iterating through all characters every time, we could maintain a map of characters by area for faster lookups. This would improve performance as the number of characters grows.

### Assumptions
- `WorldRef` is always initialized when needed (assumes global state management)
- Character data structures have valid fields like `InRoom`, `Name`, etc.
- `PCData` is never nil for player characters
- `GetHoliday` function exists and works correctly

### Security
No security issues found in the provided code.

```
adversary-review
verdict: CONCERNS
confidence: medium
artifact:
  path: internal/act/info2.go
  sha256: 9318bfcba1b9afda
  lines_reviewed: 1-175
findings:
  - id: F1
    severity: major
    category: performance
    file: internal/act/info2.go
    line: 60
    line_end: 64
    message: >
      In DoWhere, character lookup uses linear search through all characters
      rather than using a map or hash table for better performance as the
      number of characters increases.
    suggested_fix: >
      Maintain a character index by area to enable O(1) character lookup.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/info2.go
    line: 100
    line_end: 102
    message: >
      Hardcoded constants like "5" and "200" in DoPager should be named constants
      for better readability and maintainability.
    suggested_fix: >
      Define constants like MIN_PAGER_LENGTH = 5 and MAX_PAGER_LENGTH = 200
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/info2.go
    line: 38
    line_end: 44
    message: >
      The switch statement in DoConsider uses magic numbers for level differences
      instead of named constants, making the code less readable and maintainable.
    suggested_fix: >
      Define constants like DIFF_LEVEL_10, DIFF_LEVEL_5, etc. to make the logic
      more explicit and easier to understand.
```

**Final Verdict (post-quorum)**: CONCERNS
