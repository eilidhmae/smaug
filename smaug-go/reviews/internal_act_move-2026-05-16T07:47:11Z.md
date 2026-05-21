# Adversary Review

**Target**: `internal/act/move.go`
**Timestamp**: 2026-05-16T07:47:11Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

### Claim Verification
N/A - Skipping steps 0-2 as requested.

### Test Verification
N/A - Skipping step 2 as requested.

### Complexity Audit
The file contains three functions (`DoOpen`, `DoClose`, `DoUnlock`) that each handle door-related commands. Each function has similar structure and logic flow, with similar error handling and door handling code. The complexity is acceptable for the functionality provided.

### Scope Check
The file implements door opening/closing/locking functionality. No additional features or files were added beyond what's expected from the name and content of this file.

### Alternative Approach
A single function could handle all door operations (open, close, lock, unlock) by using a command parameter to determine operation type. This would reduce duplication and make the code more maintainable.

### Assumptions
1. The `types.EX_ISDOOR` flag correctly identifies doors in the game world.
2. Door information is stored in `ExitData` structures with proper bit flags.
3. Character data includes valid room references and carrying lists.
4. The `util.IsName` function properly handles name matching for keywords.
5. Room people list is properly maintained and accessible.

### Security
No security issues found in the code review.

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
      Duplicate logic exists in DoOpen and DoClose functions for handling
      door operations. The same pattern of checking door properties and
      sending messages exists in both functions.
    suggested_fix: >
      Extract common door handling logic into a shared helper function.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/move.go
    line: 48
    line_end: 54
    message: >
      The key check in hasKey function could be simplified by using a range
      loop with early return instead of explicit index checking.
    suggested_fix: >
      Replace the loop with range loop that returns immediately when key is found.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/move.go
    line: 103
    line_end: 109
    message: >
      The notification logic for room people is repeated in both DoOpen and
      DoClose functions. This duplication can be extracted to a helper function.
    suggested_fix: >
      Create a helper function to notify room people about door actions.
```
