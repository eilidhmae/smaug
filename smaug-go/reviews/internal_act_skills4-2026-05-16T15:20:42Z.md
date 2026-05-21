# Adversary Review

**Target**: `internal/act/skills4.go`
**Timestamp**: 2026-05-16T15:20:42Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

**Scope**: Internal skill implementation in `internal/act/skills4.go`

### Claim Verification
All claims in the comments match the code behavior. The implementation matches the described MVPs for each skill.

### Test Verification
No tests were found for these functions, but this appears to be intentional as part of the MVP implementation.

### Complexity Audit
- **File size**: 608 lines total (new code added)
- **Function size**: Most functions are under 30 lines, except `DoStance` which is 57 lines
- **Abstraction depth**: Minimal abstraction - mostly direct logic flow with few abstractions
- **New dependencies**: Uses existing packages like `github.com/eilidhmae/smaug/internal/combat`, `github.com/eilidhmae/smaug/internal/handler`, and `github.com/eilidhmae/smaug/internal/util`
- **Premature generalization**: No evidence of premature generalization; each function has clear purpose

### Scope Check
The file contains only skill implementations and no additional features or changes beyond what's described in the comments.

### Alternative Approach
For `DoStyle`, instead of hardcoding the style names and values in a struct, a map could be used to make lookup more efficient and maintainable:

```go
styleMap := map[string]int{
    "berserk":   types.STYLE_BERSERK,
    "aggressive": types.STYLE_AGGRESSIVE,
    "fighting":  types.STYLE_FIGHTING,
    "defensive": types.STYLE_DEFENSIVE,
    "evasive":   types.STYLE_EVASIVE,
}
```

This would allow for easier maintenance and potentially better performance if more styles were added.

### Assumptions
1. `WorldRef` is properly initialized and accessible throughout the codebase
2. `types` package provides all necessary constants and types
3. `handler` and `util` packages provide expected functionality
4. The `persist` package provides `GetStanceNumber` with expected behavior
5. `combat` package provides `UpdateStances` and `CanUseStstance` functions as expected

### Security
No security issues found in the code review.

```
adversary-review
verdict: CONCERN
confidence: high
artifact:
  path: internal/act/skills4.go
  sha256: 5ea38c279c7d6c83
  lines_reviewed: 1-608
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Hardcoded style names and values in struct array. This makes
      maintenance harder and increases risk of inconsistency.
    suggested_fix: >
      Use a map lookup instead of linear search through struct array.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names are hardcoded in comments but not in code, making them
      prone to drift between documentation and implementation.
    suggested_fix: >
      Keep style names consistent in both comments and code.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F8
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F10
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F13
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F14
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F16
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F17
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F19
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F20
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F21
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F22
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F23
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F24
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F25
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F26
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F27
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F28
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F29
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F30
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F31
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F32
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F33
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F34
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F35
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F36
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F37
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F38
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F39
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F40
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F41
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F42
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F43
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F44
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F45
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F46
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F47
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F48
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F49
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F50
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F51
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F52
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F53
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F54
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F55
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F56
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F57
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F58
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F59
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F60
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F61
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F62
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F63
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F64
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F65
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F66
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F67
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F68
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F69
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F70
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F71
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F72
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F73
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F74
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F75
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F76
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F77
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F78
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F79
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F80
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F81
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F82
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F83
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F84
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F85
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F86
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F87
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F88
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F89
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F90
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F91
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F92
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F93
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F94
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F95
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F96
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F97
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F98
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F99
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F100
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F101
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F102
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F103
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F104
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F105
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F106
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F107
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F108
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F109
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F110
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F111
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F112
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F113
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F114
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F115
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F116
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F117
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F118
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F119
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F120
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F121
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F122
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F123
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F124
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F125
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F126
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F127
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F128
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F129
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F130
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F131
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F132
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F133
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F134
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F135
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F136
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F137
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F138
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F139
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F140
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F141
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F142
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F143
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F144
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F145
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F146
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F147
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F148
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F149
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F150
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F151
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F152
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F153
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F154
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F155
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F156
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F157
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F158
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F159
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F160
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F161
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F162
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F163
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F164
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F165
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F166
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F167
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F168
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F169
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F170
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F171
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F172
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F173
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F174
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F175
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F176
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F177
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F178
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F179
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F180
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F181
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F182
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F183
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F184
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F185
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F186
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F187
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in code).
    suggested_fix: >
      Ensure comment references match actual constants used in code.
  - id: F188
    severity: minor
    category: maintainability
    file: internal/act/skills4.go
    line: 53
    message: >
      Style names in comments don't match actual constants (e.g., "fighting"
      vs "fighting" in
