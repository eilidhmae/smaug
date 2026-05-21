# Adversary Review

**Target**: `internal/act/obj.go`
**Timestamp**: 2026-05-16T16:51:36Z
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
- **Abstraction depth**: The code uses a simple switch-based approach to handle object operations, which is appropriate for this type of command handler.
- **New dependencies**: No new dependencies added beyond existing ones.
- **Premature generalization**: No premature generalization detected.
- **Feature flags**: No feature flags or conditional logic that seems unnecessary.

### Scope Check
The file contains only object handling commands (`get`, `drop`, `put`, `give`, `wear`, `remove`, `sacrifice`) and related helper functions. No additional features or files were added beyond what's expected from the scope.

### Alternative Approach
The current implementation uses a series of if/else conditions to handle different object types and actions. An alternative would be to use a map-based lookup for wear locations based on item flags, but given the limited number of wear flags and their straightforward mapping, the switch statement approach is acceptable.

### Assumptions
- The `handler` package provides all necessary functions for object manipulation (`GetObjHere`, `GetObjCarry`, etc.)
- The `mudprog` package provides all necessary trigger functions (`OprogGetTrigger`, etc.)
- The `types` package contains all required constants and structures for object handling

### Security
No security issues found in the code review.

```
adversary-review
verdict: CONCERNS
confidence: medium
artifact:
  path: internal/act/obj.go
  sha256: 8326062a92155140
  lines_reviewed: 1-377
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/obj.go
    line: 1
    message: >
      The file contains multiple commands (get, drop, put, give, wear, remove,
      sacrifice) which are all handled in one file. This violates single
      responsibility principle and makes maintenance harder.
    suggested_fix: >
      Split into separate files or at least separate functions into different
      files based on command type.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/obj.go
    line: 1
    message: >
      The file is very large (377 lines) with many functions that handle different
      object operations. This makes it hard to understand and maintain.
    suggested_fix: >
      Consider splitting into smaller files or modules for better organization.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/obj.go
    line: 1
    message: >
      The code uses hardcoded constants like ITEM_TAKE, ITEM_WEAR_FINGER, etc.
      These should be defined in a constants file or imported from a shared
      constants package.
    suggested_fix: >
      Define constants in a dedicated constants file or import them from a shared
      constants package.
```

---
## Quorum Peer 1

## Adversary Review

### Complexity Audit
The code is relatively straightforward but contains some complexity issues:
1. `findWearLoc` function (lines 256-274) has 18 lines of logic that could be simplified with a lookup table or map-based approach
2. The `itemWearRestriction` function (lines 282-320) has repetitive conditional checks that could be refactored into a more maintainable structure
3. The `DoGet` function (lines 1-35) has duplicated logic in handling object movement from room vs container

### Assumptions
1. The `handler` package provides reliable object management functions (`GetObjHere`, `GetObjCarry`, etc.)
2. `WorldRef` is properly initialized elsewhere in the system
3. `types` constants like `ITEM_TAKE`, `ITEM_WEAR_FINGER` are correctly defined and consistent

### Security Scan
No security issues found in this file.

### Scope Check
All functions appear to be within scope for object manipulation commands ('get', 'drop', 'put', 'give', 'wear', 'remove', 'sacrifice').

### Alternative Approach
For `findWearLoc`, instead of a long switch statement, consider using a map lookup or array lookup based on wear flags to simplify the logic.

### Verdict
**CONCERNS**

```
adversary-review
verdict: CONCERNS
confidence: medium
artifact:
  path: internal/act/obj.go
  sha256: 8326062a92155140
  lines_reviewed: 1-377
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/obj.go
    line: 256
    line_end: 274
    message: >
      The findWearLoc function uses a long switch statement with many conditions.
      This could be simplified by using a lookup table or map-based approach.
    suggested_fix: >
      Replace the switch statement with a map lookup or array lookup based on
      wear flags to reduce complexity and improve maintainability.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/obj.go
    line: 282
    line_end: 320
    message: >
      The itemWearRestriction function has repetitive conditional checks that
      could be refactored into a more maintainable structure.
    suggested_fix: >
      Consider creating a map of class-specific restrictions or use a more
      structured approach to check restrictions.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/obj.go
    line: 1
    line_end: 35
    message: >
      The DoGet function has duplicated logic for handling object movement from
      room vs container. Could be refactored to reduce duplication.
    suggested_fix: >
      Extract common object movement logic into a helper function to avoid
      duplicate code paths.
```

**Final Verdict (post-quorum)**: CONCERNS
