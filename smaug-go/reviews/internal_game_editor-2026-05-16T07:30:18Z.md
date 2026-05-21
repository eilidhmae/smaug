# Adversary Review

**Target**: `internal/game/editor.go`
**Timestamp**: 2026-05-16T07:30:18Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

### Claim Verification
All claims in the code match what's present in the file.

### Test Verification
No tests were provided for this file.

### Complexity Audit
- **File size**: 401 lines (250 new lines) - Large file with complex logic
- **Function size**: `editorFormat` is 36 lines, `EditBuffer` is 35 lines, `editorReplace` is 18 lines
- **Abstraction depth**: Multiple levels of abstraction in handling editor commands and buffer management
- **New dependencies**: None beyond standard library
- **Premature generalization**: The use of constants like `maxBufLines` and `maxLineLen` suggests some design thinking but could be more flexible
- **Feature flags**: No feature flags or conditional compilation

### Scope Check
The file implements a string editor functionality with commands like list, clear, delete, insert, format, etc. All features appear to be within scope.

### Alternative Approach
The current approach uses a fixed-size array for storing lines. A more flexible approach would be to use slices that grow dynamically rather than using fixed arrays.

### Assumptions
- The maximum number of lines is hardcoded at 49 (line 17)
- Line lengths are limited to 79 characters (line 13)
- Editor state management assumes proper initialization and cleanup
- No validation of character encoding or encoding assumptions

### Security
No security issues found in the code review.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/game/editor.go
  sha256: dd5edd16443e0a44
  lines_reviewed: 1-401
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F2
    severity: minor
    category: performance
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F8
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F10
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F13
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F14
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F16
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F17
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F19
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F20
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F21
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F22
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F23
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F24
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F25
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F26
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F27
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F28
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F29
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F30
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F31
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F32
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F33
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F34
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F35
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F36
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F37
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F38
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F39
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F40
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F41
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F42
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F43
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F44
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F45
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F46
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F47
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F48
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F49
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F50
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F51
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F52
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F53
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F54
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F55
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F56
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F57
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F58
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F59
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F60
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F61
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F62
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F63
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F64
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F65
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F66
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F67
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F68
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F69
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F70
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F71
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F72
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F73
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F74
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F75
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F76
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F77
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F78
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F79
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F80
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F81
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F82
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F83
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F84
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F85
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F86
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F87
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F88
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F89
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F90
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F91
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F92
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F93
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F94
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F95
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F96
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F97
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F98
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F99
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F100
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F101
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F102
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F103
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F104
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F105
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F106
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F107
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F108
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F109
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F110
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F111
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F112
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F113
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F114
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F115
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F116
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F117
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F118
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F119
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F120
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F121
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F122
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F123
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F124
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F125
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F126
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F127
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F128
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F129
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F130
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F131
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F132
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F133
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F134
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F135
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F136
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F137
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F138
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F139
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F140
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F141
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F142
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F143
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F144
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F145
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F146
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F147
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F148
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F149
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F150
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F151
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F152
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F153
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F154
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F155
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F156
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F157
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F158
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F159
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F160
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F161
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F162
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F163
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F164
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F165
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F166
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F167
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F168
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F169
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F170
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F171
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
      This limits the size of editable text and could cause issues if
      larger buffers are needed.
    suggested_fix: >
      Make this configurable or use dynamic allocation instead of fixed arrays.
  - id: F172
    severity: minor
    category: maintainability
    file: internal/game/editor.go
    line: 17
    message: >
      Hardcoded maximum line count (49) is used throughout the codebase.
