# Adversary Review

**Target**: `internal/act/move.go`
**Timestamp**: 2026-05-16T16:51:20Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

### Claim Verification
N/A - Skipping steps 0-2 as requested.

### Test Verification
N/A - Skipping step 2 as requested.

### Complexity Audit
The file contains three functions (`DoOpen`, `DoClose`, `DoUnlock`) that each handle door-related commands. Each function has similar structure and logic flow, with similar error handling and door handling patterns. The code is relatively straightforward but could benefit from better abstraction or shared logic.

### Scope Check
The file implements door opening/closing/locking functionality. No additional features beyond what's described in the code are present.

### Alternative Approach
Instead of duplicating similar logic across multiple functions, a single generic function could be created to handle door operations with different operation types (open, close, lock, unlock). This would reduce duplication and make maintenance easier.

### Assumptions
1. Character data structure has `InRoom`, `Carrying`, and `Name` fields
2. Room data structure has `GetExit` method and `People` field
3. Exit data structure has `ExitInfo`, `Key`, and `Keyword` fields
4. Constants like `EX_ISDOOR`, `EX_CLOSED`, `EX_LOCKED` exist and are properly defined
5. `util.IsName` correctly handles name matching

### Security
No security issues found in the provided code.

```
adversary-review
verdict: CONCERNS
confidence: medium
artifact:
  path: internal/act/move.go
  sha256: 6cae36bf961b24b1
  lines_reviewed: 1-210
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/move.go
    line: 17
    line_end: 23
    message: >
      Duplicate logic exists across multiple door handling functions.
      The same pattern of checking door properties is repeated in DoOpen,
      DoClose, and DoUnlock functions.
    suggested_fix: >
      Create a shared helper function to handle door property checks and
      operations to reduce duplication.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/move.go
    line: 48
    line_end: 54
    message: >
      Hardcoded direction names could be extracted into constants or
      constants defined elsewhere for better maintainability.
    suggested_fix: >
      Define direction names as constants or constants in a dedicated
      constants file.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/move.go
    line: 103
    line_end: 109
    message: >
      Room notification logic is duplicated between DoOpen and DoClose
    suggested_fix: >
      Extract room notification logic into a separate function that can
      be reused by both functions.
```

---
## Quorum Peer 1

## Adversary Review

### Claim Verification
All claims in the code match what's present in the file.

### Test Verification
No test files were provided for review.

### Complexity Audit
The file contains three functions (`DoOpen`, `DoClose`, `DoUnlock`) that each handle door-related commands. Each function has a similar structure:
1. Parse input arguments
2. Find relevant door/exit based on user input
3. Validate conditions (locked, closed, etc.)
4. Perform action or send error message

Each function is relatively simple but repetitive in structure. The logic flow is straightforward but could benefit from abstraction to reduce duplication.

### Scope Check
The file implements door handling functionality as intended. No additional features beyond what's described in the code are present.

### Alternative Approach
Instead of repeating similar logic across multiple functions, a single generic function could be used to handle door operations with different operation types (open, close, lock, unlock). This would reduce code duplication and make maintenance easier.

### Assumptions
1. Character data structure includes `InRoom`, `Carrying`, and `Name` fields
2. `ExitData` struct has `ExitInfo`, `Key`, and `Keyword` fields
3. Constants like `EX_ISDOOR`, `EX_CLOSED`, `EX_LOCKED` exist in `types` package
4. Room structure has `People` field and `GetExit` method
5. `util.IsName` function exists and works correctly for name matching

### Security
No security issues found in the provided code.

```
adversary-review
verdict: CONCERNS
confidence: medium
artifact:
  path: internal/act/move.go
  sha256: 6cae36bf961b24b1
  lines_reviewed: 1-210
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/move.go
    line: 17
    line_end: 23
    message: >
      Repetitive logic pattern across DoOpen, DoClose, DoUnlock, and DoLock.
      Could be refactored into a single generic function to reduce duplication.
    suggested_fix: >
      Create a helper function that takes operation type and handles common
      validation logic, reducing repeated condition checks and error handling.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/move.go
    line: 48
    line_end: 52
    message: >
      Hardcoded string "north" used for direction lookup. This could be
      replaced with constants or constants from a constants package.
    suggested_fix: >
      Define direction constants in a constants file and use them instead of
      hardcoded strings.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/move.go
    line: 69
    line_end: 73
    message: >
      The same error handling pattern is repeated in multiple functions:
      checking for empty arguments, finding door/exit, validating conditions,
      sending error messages. Could be abstracted into a shared validation
      function.
    suggested_fix: >
      Extract common validation logic into a shared function that can be reused
      across all door-related commands.
```

**Final Verdict (post-quorum)**: CONCERNS
