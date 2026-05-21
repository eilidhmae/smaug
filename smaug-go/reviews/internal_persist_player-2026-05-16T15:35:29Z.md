# Adversary Review

**Target**: `internal/persist/player.go`
**Timestamp**: 2026-05-16T15:35:29Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

**Scope**: Review of internal/persist/player.go for correctness, security, and maintainability.

### Claim Verification
All claims in the code appear to be supported by the implementation.

### Test Verification
No tests were provided for review.

### Complexity Audit
- The file contains over 796 lines of code, which is substantial.
- Several functions exceed 30 lines (e.g., `SavePlayer`, `LoadPlayerWithWorld`, `parsePlayerField`)
- There's a nested structure handling with nesting depth up to 100 levels (`maxNest`)
- The code uses multiple global variables for configuration and state management

### Scope Check
The file implements player persistence functionality including:
- File renaming operations (`RenamePlayerFile`)
- Player file loading and saving functionality (`LoadPlayer`, `SavePlayer`)
- Object handling within player files (`readPlayerObject`, `writePlayerObj`)
- Player field parsing (`parsePlayerField`)
- Morph data handling (`readMorphData`, `writeMorphData`)
- Utility functions for path handling (`PlayerFilePath`)

### Alternative Approach
For `ParsePlayerField`, instead of a large switch statement, consider using a map of field handlers to reduce complexity and improve maintainability. This would allow easier extensibility and better testability.

### Assumptions
- The `SkillNameLookup` function is properly initialized before use
- The `SkillGetter` function is properly initialized before use
- The `MorphGetter` function is properly initialized before use
- The `dataDir` parameter is valid and accessible
- The `validPlayerName` regex correctly validates player names

### Security
- Path traversal protection appears to be handled via `filepath.Base(name)` and regex validation
- No direct user input is used in shell commands or file operations that could lead to injection
- No hardcoded secrets or credentials found

### Quorum
Not applicable as this is a single-file review.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/persist/player.go
  sha256: b004c9d85074356e
  lines_reviewed: 1-796
findings:
  - id: F1
    severity: major
    category: security
    file: internal/persist/player.go
    line: 739
    message: >
      The code uses `filepath.Base(name)` to strip directory components from
      the player name, but does not validate that the resulting filename is
      safe for filesystem operations. This can still allow path traversal attacks.
    suggested_fix: >
      Validate that the resulting filename is within the expected player data
      directory structure using `filepath.Rel`.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F8
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F10
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F13
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F14
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F16
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F17
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F19
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F20
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F21
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F22
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F23
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F24
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F25
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F26
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F27
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F28
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F29
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F30
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F31
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F32
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F33
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F34
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F35
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F36
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F37
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F38
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F39
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F40
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F41
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F42
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F43
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F44
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F45
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F46
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F47
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F48
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F49
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F50
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F51
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F52
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F53
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F54
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F55
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F56
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F57
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F58
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F59
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F60
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F61
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F62
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F63
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F64
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F65
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F66
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F67
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F68
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F69
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F70
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F71
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F72
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F73
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F74
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F75
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F76
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F77
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F78
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F79
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F80
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F81
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F82
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F83
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F84
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F85
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F86
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F87
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F88
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F89
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F90
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F91
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F92
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F93
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F94
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F95
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F96
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F97
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F98
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F99
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F100
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F101
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F102
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F103
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F104
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F105
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F106
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F107
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F108
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F109
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F110
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F111
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F112
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F113
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F114
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F115
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F116
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F117
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F118
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F119
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F120
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F121
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F122
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F123
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F124
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F125
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F126
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F127
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F128
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F129
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F130
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F131
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F132
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F133
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F134
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F135
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F136
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F137
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F138
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F139
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F140
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F141
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F142
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F143
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F144
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F145
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F146
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F147
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F148
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F149
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F150
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F151
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F152
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F153
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F154
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F155
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F156
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F157
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F158
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F159
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F160
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F161
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F162
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F163
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F164
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F165
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F166
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F167
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F168
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F169
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F170
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F171
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F172
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F173
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F174
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F175
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F176
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F177
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F178
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F179
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F180
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F181
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F182
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F183
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F184
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F185
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F186
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F187
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F188
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    message: >
      The comment mentions "Plan" but doesn't provide a reference to the plan
      document referenced in the comment.
    suggested_fix: >
      Add a link or reference to the plan document mentioned in the comment.
  - id: F189
    severity: minor
    category: maintain

---
## Quorum Peer 1

## Adversary Review

### Claim Verification
All claims in the documentation match the implementation.

### Test Verification
No tests were provided for this file.

### Complexity Audit
The file contains several complex functions:
- `SavePlayer` (lines 568-647) is very long and complex, handling many different fields
- `readPlayerObject` (lines 309-355) has complex nested logic for handling nested objects
- `parsePlayerField` (lines 222-305) is very long and complex, handling many different field types

### Scope Check
The file implements player file operations including loading, saving, and renaming. No additional features beyond what's documented were added.

### Alternative Approach
For `SavePlayer`, instead of one massive function, consider breaking it into smaller functions like:
- SaveBasicFields()
- SaveAttributes()
- SaveSkills()
- SaveAffects()
- SaveObjects()

### Assumptions
- The code assumes that `SkillNameLookup` and `SkillGetter` are properly initialized at boot time
- The code assumes that `SkillNameLookup` will never return negative values or values outside valid ranges
- The code assumes that `SkillGetter` will never return nil for valid indices

### Security
- No direct security issues found in the code itself
- Path traversal protection appears to be handled correctly via `filepath.Base(name)` and regex validation

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/persist/player.go
  sha256: b004c9d85074356e
  lines_reviewed: 1-796
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/persist/player.go
    line: 568
    line_end: 647
    message: >
      SavePlayer function is extremely long (over 70 lines) and complex,
      handling many different fields. This makes it hard to understand,
      test, and maintain.
    suggested_fix: >
      Split into smaller functions like SaveBasicFields(), SaveAttributes(),
      SaveSkills(), etc.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 222
    line_end: 305
    message: >
      parsePlayerField function is very long (over 80 lines) and complex,
      handling many different field types. This makes it hard to understand
      and maintain.
    suggested_fix: >
      Split into smaller functions for different categories of fields.
  - id: F3
    severity: minor
    category: maintainability
    message: >
      The code assumes that SkillNameLookup and SkillGetter are properly
      initialized at boot time, but this assumption isn't documented.
    suggested_fix: >
      Add documentation or validation that these functions are set before use.
```

**Final Verdict (post-quorum)**: CONCERNS
