# Adversary Review

**Target**: `internal/mudprog/commands.go`
**Timestamp**: 2026-05-16T16:32:46Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

**Scope**: mudprog command implementations and related utilities

### Claim Verification
All mudprog commands are implemented as functions in this file. The file contains 16 commands plus utility functions.

### Test Verification
No tests were found for these commands in the provided code.

### Complexity Audit
- `mpStrew` (line 980) has 20 lines of code including comments and comments
- `mpScatter` (line 937) has 15 lines of code including comments
- `mpApplyAffect` (line 965) has 15 lines of code including comments
- `mpOpenPassage` (line 885) has 15 lines of code including comments
- `mpClosePassage` (line 905) has 15 lines of code including comments
- `mpScatter` (line 937) has 15 lines of code including comments
- `mpDeposit` (line 854) has 15 lines of code including comments
- `mpWithdraw` (line 872) has 15 lines of code including comments
- `mpDelay` (line 835) has 15 lines of code including comments
- `mpStrew` (line 980) has 20 lines of code including comments

### Scope Check
All mudprog commands are implemented as expected. No additional features were added beyond what's described in the comments.

### Alternative Approach
The `mpApplyAffect` function uses a simple string-to-int conversion approach for all parameters, which could be replaced with a more robust parsing method that validates each parameter individually before conversion.

### Assumptions
- The WorldRef global is properly initialized before any mudprog command execution
- All character data structures have valid fields and methods (e.g., `InRoom`, `Desc`, etc.)
- The `CmdRegistry` is properly initialized and accessible
- The `WorldRef` is thread-safe for concurrent access during game loop execution

### Security
- No direct user input is processed without validation or sanitization
- No command injection risks identified
- No hardcoded secrets or credentials found

### Quorum
N/A - This is a single review

```
adversary-review
verdict: CONCERNS
confidence: medium
artifact:
  path: internal/mudprog/commands.go
  sha256: 5874e0589acebcf9
  lines_reviewed: 1-1042
findings:
  - id: F1
    severity: major
    category: security
    file: internal/mudprog/commands.go
    line: 835
    message: >
      mpDelay function does not validate that the target is a PC before applying
      delay. This could allow NPCs to be delayed, which may not be intended behavior.
    suggested_fix: >
      Add validation to ensure target is a PC before applying delay.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 965
    message: >
      mpApplyAffect uses simple string-to-int conversion for all parameters,
      which could lead to unexpected behavior if inputs are malformed.
    suggested_fix: >
      Implement more robust parameter validation and error handling.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 980
    message: >
      mpStrew has 20 lines of code including comments, which is relatively long
      for a single function.
    suggested_fix: >
      Consider refactoring into smaller helper functions or adding more comments
      to clarify logic flow.
```

---
## Quorum Peer 1

## Adversary Review

### Complexity Audit
The file contains several functions that handle mob commands in a MUD environment. Some functions are overly complex or have potential issues:

1. **mpStrew function** (line 980-1000) - The function uses `util.NumberRange` to select random rooms but doesn't properly handle the case where there are no valid rooms in the area range. It could lead to panic if `len(rooms) == 0`.

2. **mpScatter function** (line 956-975) - Similar issue with potential panic when `len(candidates) == 0`. Also, the function does not properly validate that low_vnum <= high_vnum.

3. **mpApplyAffect function** (line 994-1005) - This function has multiple error checks but doesn't validate that the affect parameters are within valid ranges for the game logic.

### Assumptions
1. The `WorldRef` global variable is properly initialized before any mudprog commands execute.
2. The `CmdRegistry` is initialized and available for command execution.
3. The `util.NumberRange` function works correctly and returns valid indices.
4. The `handler` package provides correct implementations for all operations like `CharToRoom`, `ExtractChar`, etc.

### Security Scan
No security issues found in the code review.

### Verdict
**VERDICT: CONCERNS**

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/mudprog/commands.go
  sha256: 5874e0589acebcf9
  lines_reviewed: 1-1042
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/mudprog/commands.go
    line: 980
    line_end: 1000
    message: >
      mpStrew function can panic if len(rooms) == 0 because it selects a random room without checking for empty rooms slice.
    suggested_fix: >
      Add check for len(rooms) == 0 before selecting random room.
  - id: F2
    severity: major
    category: correctness
    file: internal/mudprog/commands.go
    line: 956
    line_end: 975
    message: >
      mpScatter function can panic if len(candidates) == 0 because it selects a random candidate without validation.
    suggested_fix: >
      Add validation that len(candidates) > 0 before selecting random candidate.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 994
    line_end: 1005
    message: >
      mpApplyAffect function has multiple error checks but doesn't validate that affect parameters are within valid ranges.
    suggested_fix: >
      Validate that affect parameters (type, duration, location, modifier) are within valid ranges before applying.
```

**Final Verdict (post-quorum)**: CONCERNS
