# Adversary Review

**Target**: `internal/mudprog/ifcheck.go`
**Timestamp**: 2026-05-16T06:29:26Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

**Scope**: mudprog ifcheck implementation

### Claim Verification
All claims in the code appear to be implemented correctly. The file implements various mud prog if-checks with proper handling of different character and object states.

### Test Verification
No test files were provided for review.

### Complexity Audit
- File size: 1230 lines (large but mostly due to extensive case handling)
- Function size: Several functions exceed 30 lines (e.g., DoIfCheck has over 100 lines)
- Abstraction depth: High - multiple helper functions and complex conditional logic
- New dependencies: None beyond standard library and existing types

### Scope Check
The implementation covers many more if-checks than typical mud prog implementations, including several that seem to be part of a planned feature set (like isflagged/istagged).

### Alternative Approach
The current approach uses a large switch statement with many cases. A more maintainable alternative would be to use a map of check names to handler functions, which would make adding new checks easier and reduce code duplication.

### Assumptions
- WorldRef is always initialized when needed
- Character data structures have expected fields (e.g., AffectedBy, PCData)
- String comparisons are case-insensitive as intended
- Numeric values can be parsed without error (error handling is minimal)

### Security
- No obvious injection vulnerabilities in string handling
- Error handling appears minimal (ignoring parse errors)
- No direct user input processing in this file

### Quorum
Not applicable - this is the only file being reviewed.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/mudprog/ifcheck.go
  sha256: 947141e387629450
  lines_reviewed: 1-1230
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/mudprog/ifcheck.go
    line: 1
    message: >
      The file contains over 1200 lines of code with a single large switch statement.
      This makes maintenance difficult and increases complexity significantly.
    suggested_fix: >
      Refactor into smaller functions or use a map-based approach for better maintainability.
  - id: F2
    severity: minor
    category: idiomatic
    file: internal/mudprog/ifcheck.go
    line: 100
    message: >
      The function `carryingVnumVisit` recursively walks object contents without
      any depth limit, which could lead to stack overflow on deeply nested objects.
    suggested_fix: >
      Implement iterative traversal or add depth limiting to prevent stack overflow.
  - id: F3
    severity: minor
    category: error-handling
    file: internal/mudprog/ifcheck.go
    line: 100
    message: >
      Error handling is minimal in several places where string parsing occurs,
      potentially leading to unexpected behavior if input is malformed.
    suggested_fix: >
      Add proper error handling for string-to-int conversions and validate inputs.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/mudprog/ifcheck.go
    line: 150
    message: >
      The comment mentions "TODO(tier4)" but this appears to be an incomplete implementation
      that only does exact name matching instead of the full nifty_is_name functionality.
    suggested_fix: >
      Implement full nifty_is_name functionality or remove the TODO comment.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/mudprog/ifcheck.go
    line: 1200
    message: >
      Several functions have comments indicating they're part of a planned feature set (e.g., "TODO(tier3)")
      that are not yet implemented, suggesting incomplete implementation.
    suggested_fix: >
      Remove or implement these features properly before merging.
```
