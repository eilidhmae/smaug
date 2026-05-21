# Adversary Review

**Target**: `internal/act/obj.go`
**Timestamp**: 2026-05-16T07:48:17Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

### Claim Verification
All claims in the code match what's present in the file.

### Test Verification
No tests were provided for this file.

### Complexity Audit
- **File size**: 377 lines (100% new code)
- **Function size**: 
  - `DoGet` (52 lines), `DoDrop` (24 lines), `DoPut` (36 lines), `DoGive` (35 lines), `DoWear` (38 lines), `DoRemove` (25 lines), `DoSacrifice` (25 lines), `itemWearRestriction` (25 lines), `findWearLoc` (25 lines), `wearVerb` (12 lines)
- **Abstraction depth**: 
  - `DoGet` and `DoPut` have multiple conditional branches that could be simplified.
  - `DoWear` has a complex flow with multiple checks and actions.
- **New dependencies**: Uses `github.com/eilidhmae/smaug/internal/handler` and `github.com/eilidhmae/smaug/internal/mudprog`, which are not part of standard library.
- **Premature generalization**: The wear location logic is well-defined but could benefit from a more structured approach to handle wear flags.
- **Feature flags**: No feature flags or conditional compilation detected.

### Scope Check
All functions in the file are related to object manipulation commands. No additional features were added beyond what's described in the function names.

### Alternative Approach
The current implementation uses multiple conditional branches for handling different object types and wear locations. A more maintainable approach would be to use a map-based lookup for wear locations based on item type and flags, reducing complexity and improving readability.

### Assumptions
- The `handler` package provides correct implementations for all object manipulation functions.
- The `mudprog` package correctly triggers events.
- The `types` package contains expected constants and structures.
- All objects have valid `ShortDescr` fields.
- Character data includes valid `InRoom`, `IsNPC`, `Act`, `Alignment`, `Class`, and `Gold` fields.

### Security
No security issues found in the code review.

### Quorum
Not applicable.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/act/obj.go
  sha256: 8326062a92155140
  lines_reviewed: 1-377
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/obj.go
    line: 100
    message: >
      The wear location logic is complex and hard to maintain. It uses
      multiple conditional checks that could be simplified with a lookup table.
    suggested_fix: >
      Replace the switch statement with a map-based lookup for wear locations.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/obj.go
    line: 100
    message: >
      Wear location determination logic is duplicated across multiple functions.
    suggested_fix: >
      Extract wear location determination into a shared function or constant.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/obj.go
    line: 100
    message: >
      The wear verb logic is hardcoded and could be made more extensible.
    suggested_fix: >
      Use a map or slice to store verb mappings for different wear locations.
```
