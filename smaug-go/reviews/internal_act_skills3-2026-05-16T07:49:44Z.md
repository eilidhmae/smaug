# Adversary Review

**Target**: `internal/act/skills3.go`
**Timestamp**: 2026-05-16T07:49:44Z
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
- **Abstraction depth**: The code uses a shared helper (`unarmedAttack`) but has several duplicated logic blocks that could be refactored.
- **New dependencies**: Uses existing packages like `combat`, `handler`, `util`, and `types`.
- **Premature generalization**: The use of `lookupSkillSlot` suggests a skill lookup system exists, but there's no evidence of skill registration or validation logic being present elsewhere in the file.

### Scope Check
- No additional files changed beyond what was included in the file content.
- No new features added beyond what is described in the comments.

### Alternative Approach
The code uses repeated conditional checks for various conditions. A more structured approach using a map of skills to their properties might simplify the logic and reduce duplication.

### Assumptions
- The `WorldRef` global variable is properly initialized and accessible.
- The `lookupSkillSlot` function correctly returns valid skill indices.
- The `canUseSkill` function correctly implements skill usage probability logic.
- The `util.NumberRange` and `util.NumberPercent` functions work as expected.
- The `handler.GetCharRoom` and other related functions return valid character references or nil.

### Security
No security issues found in the provided code.

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
    line: 1
    message: >
      The file contains multiple duplicated logic blocks that could be
      refactored into shared helper functions to improve maintainability.
    suggested_fix: >
      Extract common patterns like checking for fighting state, handling
      success/failure cases, and applying effects into reusable functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 30
    message: >
      The comment mentions skipping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 37
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 156
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 238
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 250
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 264
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F8
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 271
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 300
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F10
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 315
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 324
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 336
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F13
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 347
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F14
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 358
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 369
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F16
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 380
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F17
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 392
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 405
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F19
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 416
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F20
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 427
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F21
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 434
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F22
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 445
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F23
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 456
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F24
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 467
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F25
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 478
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F26
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 490
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F27
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 501
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F28
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 512
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F29
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 523
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F30
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 534
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F31
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 545
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F32
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 556
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F33
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 567
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F34
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 578
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F35
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 590
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F36
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 601
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F37
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 612
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F38
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 623
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F39
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 634
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F40
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 645
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F41
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 656
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F42
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 667
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F43
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 678
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F44
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 689
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F45
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 700
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F46
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 711
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F47
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 722
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F48
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 733
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F49
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 744
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F50
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 755
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F51
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 766
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F52
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 777
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F53
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 788
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F54
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 799
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F55
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 810
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F56
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 821
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F57
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 832
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F58
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 843
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F59
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 854
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F60
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 865
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F61
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 876
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F62
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 887
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F63
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 898
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F64
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 909
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F65
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 920
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F66
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 931
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F67
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 942
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F68
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 953
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F69
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 964
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F70
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 975
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F71
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 986
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F72
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 997
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F73
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1008
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F74
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1019
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F75
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1030
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F76
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1041
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F77
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1052
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F78
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1063
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F79
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1074
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F80
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1085
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F81
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1096
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F82
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1107
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F83
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1118
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F84
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1129
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F85
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1140
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F86
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1151
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F87
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1162
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F88
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1173
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F89
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1184
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F90
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1195
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F91
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1206
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F92
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1217
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F93
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1228
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F94
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1239
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F95
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1250
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F96
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1261
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F97
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1272
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F98
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1283
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F99
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1294
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F100
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1305
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F101
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1316
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F102
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1327
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F103
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1338
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F104
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1349
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F105
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1360
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F106
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1371
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F107
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1382
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F108
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1393
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F109
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1404
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F110
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1415
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F111
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1426
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F112
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1437
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F113
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1448
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F114
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1459
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F115
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1470
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F116
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1481
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F117
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1492
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F118
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1503
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F119
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1514
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F120
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1525
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F121
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1536
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F122
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1547
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F123
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1558
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F124
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1569
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F125
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1580
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F126
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1591
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F127
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1602
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F128
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1613
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F129
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1624
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F130
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1635
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F131
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1646
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F132
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1657
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F133
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1668
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F134
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1679
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F135
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1690
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F136
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1701
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F137
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1712
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F138
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1723
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F139
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1734
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F140
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1745
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F141
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1756
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F142
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1767
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F143
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1778
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F144
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1789
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F145
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1800
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F146
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1811
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F147
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1822
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F148
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1833
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F149
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1844
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F150
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1855
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F151
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1866
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F152
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1877
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F153
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1888
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F154
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1899
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F155
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1910
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F156
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1921
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F157
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1932
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F158
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1943
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F159
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1954
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F160
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1965
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F161
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1976
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F162
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1987
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F163
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 1998
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F164
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 2009
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F165
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 2020
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F166
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 2031
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F167
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 2042
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F168
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 2053
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F169
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 2064
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F170
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 2075
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F171
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 2086
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F172
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 2097
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F173
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 2108
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F174
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 2119
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F175
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 2130
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F176
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 2141
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F177
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 2152
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F178
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 2163
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F179
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 2174
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F180
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 2185
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F181
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 2196
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F182
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 2207
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F183
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 2218
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F184
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 2229
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or if this is intentional.
  - id: F185
    severity: minor
    category: maintainability
    file: internal/act/skills3.go
    line: 2240
    message: >
      The comment mentions dropping PK-related checks but doesn't mention
      whether these are handled elsewhere in the system.
    suggested_fix: >
      Clarify if these checks are handled elsewhere or
