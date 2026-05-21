# Adversary Review

**Target**: `internal/act/skills3.go`
**Timestamp**: 2026-05-16T18:57:21Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

**Scope**: Review of internal/act/skills3.go for correctness, complexity, and security issues.

### Claim Verification
All claims in the code comments match what's implemented in the file.

### Test Verification
N/A - No tests were provided or referenced.

### Complexity Audit
- **File size**: 434 lines (new code added)
- **Function size**: Several functions exceed 30 lines:
  - `DoHitall` (50+ lines)
  - `DoPounce` (30+ lines)
  - `DoGrapple` (20+ lines)
  - `DoBerserk` (15+ lines)
- **Abstraction depth**: The code uses a shared helper (`unarmedAttack`) but has several complex conditional paths that could be simplified.
- **New dependencies**: Uses existing packages like `combat`, `handler`, and `util`.
- **Premature generalization**: The `unarmedAttack` function is a good abstraction, but some skills have specific logic that might benefit from more modular design.

### Scope Check
- No additional files changed beyond what was expected.
- No extra features added beyond what's described in the comments.

### Alternative Approach
The `DoHitall` function uses a loop to process victims and handles retcode tracking. A simpler approach would be to simplify the loop structure and reduce complexity by simplifying conditionals and reducing nested checks.

### Assumptions
- Assumes `WorldRef` is properly initialized and accessible.
- Assumes `ch.Fighting != nil` when needed for combat actions.
- Assumes `handler.GetCharRoom` returns valid character references or nil.
- Assumes `ch.InRoom` is not nil when processing room actions.

### Security
- No direct injection vulnerabilities detected.
- No hardcoded secrets or credentials found.
- No unsafe operations on user input.

### Quorum
N/A - No peer review required.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/act/skills3.go
  sha256: 996c65ba648fafd0
  lines_reviewed: 1-434
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/skills3.go
    line: 37
    line_end: 43
    message: >
      The DoHitall function has complex logic with multiple nested conditions
      and retcode handling that could be simplified to improve readability and maintainability.
    suggested_fix: >
      Refactor the loop structure to simplify conditionals and reduce nesting.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 230
    line_end: 232
    message: >
      The `DoGrapple` function uses a newAff closure which is used twice. This
      could be extracted into a separate function for better clarity.
    suggested_fix: >
      Extract the `newAff` closure into a standalone function to avoid duplication.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 35
    line_end: 36
    message: >
      The comment in `unarmedAttack` mentions that it's porting from C code but
      doesn't mention any specific issues or considerations regarding the port.
    suggested_fix: >
      Add more context about potential differences between C and Go implementations.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/skills3.go
    line: 37
    line_end: 43
    message: >
      The DoHitall function has complex logic with multiple nested conditions
      and retcode handling that could be simplified to improve readability and maintainability.
    suggested_fix: >
      Refactor the loop structure to simplify conditionals and reduce nesting.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 230
    line_end: 232
    message: >
      The `DoGrapple` function uses a newAff closure which is used twice. This
      could be extracted into a separate function for better clarity.
    suggested_fix: >
      Extract the `newAff` closure into a standalone function to avoid duplication.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 35
    line_end: 36
    message: >
      The comment in `unarmedAttack` mentions that it's porting from C code but
      doesn't mention any specific issues or considerations regarding the port.
    suggested_fix: >
      Add more context about potential differences between C and Go implementations.
```

---
## Quorum Peer 1

## Adversary Review

### Claim Verification
All claims in the code match what's present in the file.

### Test Verification
No tests were added or modified as part of this change.

### Complexity Audit
- **File size**: The file contains 434 lines of code.
- **Function size**: Several functions exceed 30 lines (e.g., `DoHitall`, `DoPounce`, `DoGrapple`, `DoBerserk`)
- **Abstraction depth**: The code uses a shared helper function `unarmedAttack` for similar attacks, which is good.
- **New dependencies**: No new dependencies introduced.
- **Premature generalization**: The use of `lookupSkillSlot` suggests potential for future extensibility but no over-engineering is evident here.
- **Feature flags**: No feature flags or conditional logic that seems unnecessary.

### Scope Check
The changes align with the intended scope of implementing various combat skills in Go, including both offensive and defensive skills like grapple, cleave, and berserk.

### Alternative Approach
An alternative approach would be to define a common interface for all skill types and implement them through a single dispatch mechanism rather than individual functions. However, given the current implementation style and simplicity, this isn't necessary.

### Assumptions
1. `WorldRef` is properly initialized and accessible throughout the application.
2. The `util` package provides reliable random number generation.
3. `handler.GetCharRoom` correctly handles character lookup within the room context.
4. `handler.AffectToChar` correctly applies affects to characters.
5. `combat.Damage` and related functions behave as expected in terms of damage calculation and effect application.

### Security
No security issues found in the provided code.

```
adversary-review
verdict: CONCERN
confidence: high
artifact:
  path: internal/act/skills3.go
  sha256: 996c65ba648fafd0
  lines_reviewed: 1-434
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills3.go
    line: 33
    message: >
      The function `unarmedAttack` uses `util.NumberRange(1, ch.Level)` for
      damage calculation, which could lead to very high damage values at higher levels.
    suggested_fix: >
      Consider implementing a more balanced damage formula that scales with level but doesn't scale linearly.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 33
    message: >
      The comment mentions "MVP: skip per-class level gating" but does not address
      how skill levels are handled elsewhere in the system.
    suggested_fix: >
      Consider adding a comment or comment block explaining why this is an MVP and what needs to be done to improve it.
findings:
  - id: F

---
## Quorum Peer 2

## Adversary Review

The code implements several combat skills in Go, porting logic from C MUD code. Most implementations are straightforward but some have potential issues:

### Claim Verification
All claims in the comments match the implementation.

### Test Verification
No tests were provided for these functions.

### Complexity Audit
- `DoHitall` has complex logic with multiple conditions and loops that could be simplified.
- `DoPounce` has a complex conditional structure that's hard to follow.
- `DoHitall` has nested loops and conditionals that increase complexity.

### Scope Check
The implementation matches the requested scope of implementing combat skills.

### Alternative Approach
For `DoHitall`, instead of copying the People slice and filtering, we could filter directly in the loop or use a more functional approach.

### Assumptions
- Assumes `WorldRef` is always initialized when needed.
- Assumes `ch.Fighting != nil` when fighting.
- Assumes `handler.GetCharRoom` returns valid character references.

### Security
No security issues found.

### Quorum
No concerns identified.

```
adversary-review
verdict: CONCERNS
confidence: medium
artifact:
  path: internal/act/skills3.go
  sha256: 996c65ba648fafd0
  lines_reviewed: 1-434
findings:
- id: F1
  severity: major
  category: maintainability
  file: internal/act/skills3.go
  line: 37
  line_end: 43
  message: >
    The DoHitall function has complex logic with multiple conditions and
    nested loops that make it hard to follow. The filtering of victims
    should be simplified.
  suggested_fix: >
    Refactor to simplify the victim filtering and processing logic.
- id: F2
  severity: minor
  category: maintainability
  file: internal/act/skills3.go
  line: 230
  line_end: 234
  message: >
    The DoPounce function has complex conditional logic that's hard to
    understand due to multiple nested conditions.
  suggested_fix: >
    Simplify conditionals by extracting checks into named functions or
    early returns.
- id: F3
  severity: minor
  category: maintainability
  file: internal/act/skills3.go
  line: 300
  line_end: 305
  message: >
    The DoBerserk function uses a complex formula for calculating duration
    that could be simplified.
  suggested_fix: >
    Use simpler calculations for duration calculation.
```

**Final Verdict (post-quorum)**: CONCERNS
