# Adversary Review

**Target**: `internal/act/social.go`
**Timestamp**: 2026-05-16T08:00:49Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

### Claim Verification
This file implements social command handling logic. The implementation appears to handle social commands with proper variable substitution and room messaging.

### Test Verification
No test files were provided for review.

### Complexity Audit
- File size: 114 lines (within limit)
- Function sizes: `CheckSocial` is 35 lines, `socialSub` is 20 lines, `pronoun` is 7 lines - all acceptable
- Abstraction depth: Minimal, no unnecessary indirections
- No new dependencies added
- No premature generalization detected

### Scope Check
The code implements social command handling functionality as intended. No additional features or scope creep detected.

### Alternative Approach
The current approach uses string replacement for variable substitution. An alternative would be to use a more structured template system or regex-based substitution that could handle complex formatting more robustly.

### Assumptions
1. `WorldRef` is globally available and properly initialized
2. `ch.InRoom` exists and has valid `People` slice
3. `ch.Desc` is never nil when checking room people
4. `ch.Sex` is one of the expected values (MALE, FEMALE, OTHER)

### Security
No security issues found in the code review.

```
adversary-review
verdict: CONCERNS
confidence: medium
artifact:
  path: internal/act/social.go
  sha256: fe59aa069acc4dd7
  lines_reviewed: 1-114
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/social.go
    line: 18
    line_end: 18
    message: >
      Hardcoded string "They aren't here." returned without proper error handling.
      This could lead to inconsistent user experience if the error message format
      changes or needs localization.
    suggested_fix: >
      Consider using a constant or centralized error message handling instead of
      hardcoded strings.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/social.go
    line: 35
    line_end: 35
    message: >
      The error from util.OneArgument is ignored. If this function can return
      errors, they should be handled appropriately.
    suggested_fix: >
      Handle the error from util.OneArgument and handle it gracefully or log it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/social.go
    line: 45
    line_end: 45
    message: >
      Hardcoded string "They aren't here." returned without proper error handling.
      This could lead to inconsistent user experience if the error message format
      changes or needs localization.
    suggested_fix: >
      Consider using a constant or centralized error message handling instead of
      hardcoded strings.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/act/social.go
    line: 60
    line_end: 60
    message: >
      Hardcoded string "They aren't here." returned without proper error handling.
      This could lead to inconsistent user experience if the error message format
      changes or needs localization.
    suggested_fix: >
      Consider using a constant or centralized error message handling instead of
      hardcoded strings.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/act/social.go
    line: 72
    line_end: 72
    message: >
      Hardcoded string "They aren't here." returned without proper error handling.
      This could lead to inconsistent user experience if the error message format
      changes or needs localization.
    suggested_fix: >
      Consider using a constant or
```
