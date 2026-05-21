# Adversary Review

**Target**: `internal/handler/handler.go`
**Timestamp**: 2026-05-16T06:28:20Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

### Claim Verification
All claims in the code match what's present in the file.

### Test Verification
No tests were provided for this file.

### Complexity Audit
- **File size**: 546 lines total, 500+ lines new code added (over threshold of 150)
- **Function size**: Several functions exceed 30 lines:
  - `CreateMobile` (28 lines)
  - `CreateObject` (25 lines)
  - `ExtractChar` (25 lines)
  - `ExtractObj` (17 lines)
  - `AffectJoin` (15 lines)
  - `AffectStrip` (9 lines)
  - `AffectRemove` (10 lines)
  - `AffectToChar` (10 lines)
  - `AffectModify` (20 lines)
  - `UnequipChar` (10 lines)
  - `CanDropObj` (3 lines)

### Scope Check
The file contains more than just the requested handler functionality. It includes:
- Interpolation function (`interpolate`)
- Affects handling logic (`AffectModify`, `AffectToChar`, `AffectRemove`, `AffectStrip`, `AffectJoin`)
- Object extraction and manipulation (`ExtractObj`, `ObjFromRoom`, etc.)
- Character manipulation (`CharToRoom`, `CharFromRoom`, `ExtractChar`, etc.)

### Alternative Approach
The code uses a large switch statement in `AffectModify` that could be replaced with a map-based lookup for better maintainability and readability.

### Assumptions
- The `util` package provides `NumberFuzzy`, `DiceRoll`, `NumberRange`, and `UMIN` functions as expected.
- `types.LEVEL_AVATAR` is defined and non-zero.
- Room lists are managed correctly by appending to slices without proper indexing management.
- No concurrent access to character or room data structures is assumed.

### Security
No security issues found in the provided code.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/handler/handler.go
  sha256: e7c3ae116dc743aa
  lines_reviewed: 1-546
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/handler/handler.go
    line: 47
    message: >
      Large switch statement in AffectModify function (lines 47-80) makes
      maintenance difficult and increases complexity. Could be refactored into
      a map lookup or similar structure.
    suggested_fix: >
      Replace switch with a map of location -> function that modifies stats,
      or use reflection if necessary.
  - id: F2
    severity: minor
    category: performance
    file: internal/handler/handler.go
    line: 139
    message: >
      In ExtractChar, character carrying list is iterated backwards to remove
      objects. This is inefficient for large lists.
    suggested_fix: >
      Use forward iteration and track indices carefully, or use a more efficient
      data structure like a set or hash table for tracking objects.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/handler/handler.go
    line: 153
    message: >
      Hardcoded constants used in affect handling (e.g., `types.WEAR_NONE`)
      should be defined as constants elsewhere for clarity and reuse.
    suggested_fix: >
      Define constants like `WEAR_NONE` in a constants file or package.
```
