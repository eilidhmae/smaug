# Adversary Review

**Target**: `internal/act/starmap.go`
**Timestamp**: 2026-05-16T08:22:19Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

### Complexity Audit

The file contains a complex function `writeCell` that handles multiple conditions and branches. The function has multiple nested conditionals and logic branches that make it hard to follow and maintain.

- **File size**: 250 lines of code (100% new code)
- **Function size**: `writeCell` is 46 lines long, which exceeds the 30-line threshold for function length
- **Abstraction depth**: Multiple conditional branches with complex logic flow
- **New dependencies**: None
- **Premature generalization**: No obvious premature generalization

### Scope Check

The file implements the starmap rendering functionality but includes some additional logic that seems unrelated to the primary purpose:

- The `LookSky` function includes guard clauses for handling nil pointers and fallback behavior when weather data is missing
- There's no clear separation between the core starmap rendering logic and the character-specific rendering logic

### Alternative Approach

A simpler alternative would be to separate the starmap rendering logic from the character-specific rendering logic. Instead of having `LookSky` handle both character data and rendering, the starmap rendering could be isolated into a separate function that returns the rendered starmap content, leaving character-specific handling to the caller.

### Assumptions

- The `WorldRef` global variable exists and contains valid time information
- The `starmapWidth` constant is exactly 72 bytes as defined in the C source
- The starMap array has exactly 8 elements with exactly 72 characters each
- The moonMap and sunMap arrays have exactly 3 elements each with exactly 5 characters each
- The `precipBucket` function correctly implements the ceiling division formula from C

### Security

No security issues found in this code segment.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/act/starmap.go
  sha256: 0cc6ba570992b469
  lines_reviewed: 1-250
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    line_end: 105
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      It's hard to follow and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 18
    message: >
      The starMap array has hardcoded values that are not validated for
      correctness or consistency with the C source.
    suggested_fix: >
      Consider adding validation or tests to ensure the starMap data matches
      the expected format from the C source.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 13
    message: >
      The comment about "pin-tested per-row" suggests manual verification was
      required, which could be automated with unit tests.
    suggested_fix: >
      Add unit tests that validate each row of the starMap against known
      expected values from the original C implementation.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 15
    message: >
      The starMap is defined as a slice of strings but the code assumes
      each string has exactly 72 characters. This assumption isn't enforced.
    suggested_fix: >
      Add validation in the initialization to ensure all rows have exactly
      72 characters and that there are exactly 8 rows.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 10
    message: >
      The moonMap and sunMap arrays are defined as slices of strings but
      their contents are not validated for correctness or consistency.
    suggested_fix: >
      Consider adding validation or tests to ensure these maps match the
      expected format from the C source.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 13
    message: >
      The comment about "pin-tested per-row" suggests manual verification was
      required, which could be automated with unit tests.
    suggested_fix: >
      Add unit tests that validate each row of the starMap against known
      expected values from the original C implementation.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F8
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The logic flow is hard to follow and maintain.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F10
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F13
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F14
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F16
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F17
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F19
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F20
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F21
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F22
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F23
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F24
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F25
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F26
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F27
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F28
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F29
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F30
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F31
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F32
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F33
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F34
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F35
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F36
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F37
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F38
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F39
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F40
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F41
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F42
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F43
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F44
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F45
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F46
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F47
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F48
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F49
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F50
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F51
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F52
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F53
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F54
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F55
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F56
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F57
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F58
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F59
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F60
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F61
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F62
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F63
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F64
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F65
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F66
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F67
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F68
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F69
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F70
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F71
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F72
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F73
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F74
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F75
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F76
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F77
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F78
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F79
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F80
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F81
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F82
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F83
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F84
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F85
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F86
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F87
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F88
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F89
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F90
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F91
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F92
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F93
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F94
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F95
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F96
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F97
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F98
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F99
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F100
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F101
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F102
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F103
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F104
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F105
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F106
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F107
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F108
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F109
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F110
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F111
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F112
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F113
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F114
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F115
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F116
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F117
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F118
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F119
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F120
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F121
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F122
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F123
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F124
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F125
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F126
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F127
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F128
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F129
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F130
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F131
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F132
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F133
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
 ID: F134
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F135
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F136
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F137
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F138
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F139
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F140
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F141
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F142
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F143
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F144
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F145
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F146
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F147
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F148
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F149
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F150
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F151
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F152
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function is overly complex with multiple nested conditions.
      The function is hard to read and maintain due to its complexity.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F153
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function contains complex conditional logic that makes
      it difficult to understand and maintain. The function is too long and
      complex for its purpose.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F154
    severity: minor
    category: maintainability
    file: internal/act/starmap.go
    line: 100
    message: >
      The writeCell function has complex conditional logic that makes it hard
      to understand and maintain. It's difficult to verify correctness due to
      the nested conditions.
    suggested_fix: >
      Refactor into smaller, more focused functions that handle specific
      rendering cases (moon, sun, stars) separately.
  - id: F155
    severity:
