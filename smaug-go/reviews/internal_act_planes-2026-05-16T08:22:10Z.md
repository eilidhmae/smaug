# Adversary Review

**Target**: `internal/act/planes.go`
**Timestamp**: 2026-05-16T08:22:10Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

**Scope**: Internal plane management logic in `internal/act/planes.go`

### Claim Verification
N/A - No claims made about implementation or functionality.

### Test Verification
N/A - No test files mentioned or referenced.

### Complexity Audit
The file contains several issues:
1. **Race condition**: The `WorldRef` global variable is accessed and modified without synchronization, which could lead to data races if multiple goroutines access it simultaneously.
2. **Performance issue**: The `planeLookup` function iterates through the entire slice twice for each lookup operation, resulting in O(n) time complexity for lookups.
3. **Maintainability concern**: The code uses string comparisons with `strings.EqualFold` and `strings.HasPrefix` for command parsing, which is less efficient than direct string comparison when possible.

### Scope Check
No scope creep detected.

### Alternative Approach
A more efficient approach would be to use a map-based lookup structure instead of linear search, which would reduce lookup time from O(n) to O(1) on average.

### Assumptions
1. The `WorldRef` global variable is properly initialized before any operations occur.
2. The `PlanesFilePath` is correctly set at boot time.
3. The `persist.SavePlanes` function correctly persists plane data to disk.
4. The `persist.CheckPlanes` function correctly reassigns rooms to planes when planes are deleted.

### Security
No security issues found.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/act/planes.go
  sha256: 119a434c4d3a02c7
  lines_reviewed: 1-188
findings:
  - id: F1
    severity: major
    category: race-condition
    file: internal/act/planes.go
    line: 13
    line_end: 13
    message: >
      Global variable `WorldRef` is accessed and modified without synchronization.
      Concurrent access could lead to data races in multi-g Goroutine environments.
    suggested_fix: >
      Use sync.RWmutex to protect access to `WorldRef` and ensure thread safety.
  - id: F2
    severity: minor
    category: performance
    file: internal/act/planes.go
    line: 35
    line_end: 45
    message: >
      Linear search through all planes for each lookup operation results in O(n)
      time complexity. This can become inefficient as the number of planes grows.
    suggested_fix: >
      Consider using a map-based lookup structure for faster plane retrieval.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/planes.go
    line: 63
    line_end: 64
    message: >
      String comparison using `strings.EqualFold` and `strings.HasPrefix` is less
      efficient than direct string comparison when possible.
    suggested_fix: >
      Replace `strings.EqualFold` with direct string comparison where appropriate,
      or use `strings.EqualFold` only when case-insensitive matching is required.
```
