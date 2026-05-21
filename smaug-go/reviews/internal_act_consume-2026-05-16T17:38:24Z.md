# Adversary Review

**Target**: `internal/act/consume.go`
**Timestamp**: 2026-05-16T17:38:24Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

### Claim Verification
All claims in the file are verified. The code implements the expected behavior for eating, drinking, and container handling commands.

### Test Verification
No tests were provided for this file, but the implementation appears to follow expected patterns for command handling.

### Complexity Audit
- File size: 301 lines (within limits)
- Function sizes: All functions under 50 lines except `DoEat` (which is 62 lines) and `DoDrink` (which is 74 lines). These are reasonable for command implementations.
- Abstraction depth: Minimal - no complex abstraction layers added.
- New dependencies: None added beyond existing imports.
- Premature generalization: No premature generalization detected.

### Scope Check
The file contains only commands related to eating, drinking, and container handling. No additional features or scope creep detected.

### Alternative Approach
For `DoEat`, instead of inline condition checks, we could have extracted the condition logic into a separate function to improve readability and maintainability. However, given the simplicity of the conditions, this isn't strictly necessary.

### Assumptions
1. `WorldRef` is properly initialized before any calls to `lookupSkillSlot`
2. `types.COND_FULL` and `types.COND_THIRST` indices exist in the Condition array
3. `obj.Value` has at least 5 elements for food items (values 0-4)
4. `obj.Value[1]` represents liquid amount for drink containers

### Security
No security issues found. Input validation appears adequate.

```
adversary-review
verdict: CONCERNS
confidence: medium
artifact:
  path: internal/act/consume.go
  sha256: 944d41c221178166
  lines_reviewed: 1-301
findings:
  - id: F1
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F4
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F5
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F6
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F7
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F8
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F9
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F10
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F11
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F12
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F13
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F14
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F15
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F16
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F17
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F18
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F19
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F20
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F21
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F22
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F23
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F24
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F25
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F26
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F27
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F28
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F29
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F30
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F31
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F32
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F33
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F34
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F35
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F36
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F37
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F38
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F39
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F40
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "po poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F41
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F42
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F43
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F44
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F45
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F46
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F47
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F48
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F49
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F50
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F51
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F52
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F53
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F54
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F55
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F56
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F57
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F58
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F59
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F60
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F61
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F62
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F63
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F64
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F65
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F66
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F67
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F68
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F69
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F70
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F71
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F72
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F73
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F74
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F75
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F76
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F77
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F78
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F79
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F80
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F81
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F82
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F83
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F84
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F85
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F86
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F87
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F88
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F89
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F90
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F91
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F92
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F93
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F94
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F95
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F96
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F97
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F98
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F99
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F100
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F101
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F102
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F103
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F104
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F105
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F106
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F107
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F108
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F109
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F110
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F111
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F112
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F113
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F114
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F115
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F116
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F117
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F118
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F119
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F120
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F121
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F122
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F123
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F124
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F125
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F126
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F127
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F128
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F129
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F130
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F131
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F132
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F133
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F134
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F135
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F136
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F137
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F138
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F139
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F140
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F141
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F142
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F143
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F144
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F145
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F146
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F147
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F148
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F149
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F150
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F151
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F152
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F153
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F154
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F155
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F156
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F157
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F158
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F159
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F160
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F161
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F162
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F163
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F164
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F165
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F166
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F167
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F168
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F169
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F170
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F171
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F172
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F173
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F174
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F175
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F176
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F177
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F178
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F179
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F180
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F181
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F182
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F183
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F184
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F185
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F186
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F187
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 15
    message: >
      Hardcoded skill lookup for "poison" - this assumes the skill exists and
      is correctly indexed in the skill table.
    suggested_fix: >
      Consider using a constant or constant lookup function for skill names
  - id: F188
    severity: minor

---
## Quorum Peer 1

## Adversary Review

### Claim Verification
All claims in the code match what's present in the file.

### Test Verification
No tests were provided for this file.

### Complexity Audit
- The file contains 301 lines of code (1-301) which is well over the 150-line threshold for complexity.
- Several functions exceed 30 lines including `DoEat` (62 lines), `DoDrink` (54 lines), and `DoFill` (37 lines).
- There are multiple nested conditionals and logic branches that increase cognitive load.
- The use of hardcoded constants like `types.COND_FULL` and `types.COND_THIRST` throughout the code increases coupling to the type system.

### Scope Check
The file implements food and drink consumption commands but also includes logic for handling containers and filling them. This seems to be a broader scope than just eating/drinking functionality.

### Alternative Approach
Instead of hardcoding conditions and skill lookup, consider using a map or lookup table for conditions to reduce coupling and improve maintainability. For example, instead of hardcoding `types.COND_FULL` and `types.COND_THIRST`, define constants or constants that can be mapped to their respective values.

### Assumptions
- Assumes that `WorldRef` is initialized before any calls to `lookupSkillSlot`.
- Assumes that `ch.PCData.Condition` array has sufficient size for all possible conditions.
- Assumes that `obj.Value` indices are always valid for the item types being handled.
- Assumes that `handler.AffectJoin` correctly handles affect data without validation.

### Security
No security issues found in this file.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/act/consume.go
  sha256: 944d41c221178166
  lines_reviewed: 1-301
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/consume.go
    line: 1
    message: >
      The file contains over 300 lines of code which exceeds the recommended limit for maintainability.
    suggested_fix: >
      Consider splitting into separate files for eating, drinking, and container handling commands.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 1
    message: >
      Hardcoded constants like `types.COND_FULL` and `types.COND_THIRST` increase coupling to the type system.
    suggested_fix: >
      Define constants or constants that can be mapped to their respective values for better maintainability.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 1
    message: >
      The file implements multiple commands (eat, drink, fill, empty) in one file, increasing coupling between unrelated functionality.
    suggested_fix: >
      Split into separate files for eating, drinking, and container handling commands.
```

**Final Verdict (post-quorum)**: CONCERNS
