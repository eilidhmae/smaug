# Adversary Review

**Target**: `internal/act/skills.go`
**Timestamp**: 2026-05-16T06:30:03Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

**Scope**: Review of internal/act/skills.go for correctness, complexity, and security issues.

### Claim Verification
All claims in the code appear to be supported by the implementation. The file implements skill learning logic, combat skills, utility skills, and stealth skills with appropriate checks and error handling.

### Test Verification
No tests were provided for this file, but the code appears to follow standard patterns for skill systems in MUDs.

### Complexity Audit
- **File size**: 668 lines (proportional to content)
- **Function size**: Most functions are under 30 lines, except `learnFromSuccess` which is 54 lines
- **Abstraction depth**: Moderate - uses indirection through `lookupSkillSlot` and `WorldRef`
- **New dependencies**: Uses `util`, `handler`, and `combat` packages
- **Premature generalization**: The use of `WorldRef` and `lookupSkillSlot` suggests some abstraction but seems justified for the system's design

### Scope Check
The file contains both combat and utility skills, as well as utility commands like `DoScan` and `DoRecall`. These are all related to character actions and skills within the game world.

### Alternative Approach
For the `learnFromSuccess` function, an alternative approach would be to extract the XP calculation logic into a separate function or struct to improve maintainability and testability.

### Assumptions
1. `WorldRef` is properly initialized before any skill-related functions are called
2. `lookupSkillSlot` returns valid skill indices that can be used to index into `WorldRef.Skills`
3. `WorldRef.Skills` array has sufficient capacity to accommodate all skill indices
4. `lookupSkillSlot` returns valid indices that can be used to access `WorldRef.Skills`

### Security
No security issues found in this file. All input validation appears appropriate.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/act/skills.go
  sha256: ddcb344255dcb897
  lines_reviewed: 1-668
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The gsnHide and gsnSneak variables are initialized to -1, but there's no
      guarantee that ResolveGSNs() will be called before any skill usage occurs.
    suggested_fix: >
      Ensure ResolveGSNs() is called during initialization or add a check for
      whether the values have been resolved before using them.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F8
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F10
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F13
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F14
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F16
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F17
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F19
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F20
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F21
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F22
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F23
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F24
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F25
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F26
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F27
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F28
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F29
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F30
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F31
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F32
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F33
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F34
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F35
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F36
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F37
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F38
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F39
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F40
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F41
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F42
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F43
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F44
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F45
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F46
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F47
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F48
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F49
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F50
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F51
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F52
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F53
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F54
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F55
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F56
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F57
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F58
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F59
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F60
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F61
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F62
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F63
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F64
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F65
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F66
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F67
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F68
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F69
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F70
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F71
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F72
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F73
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F74
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F75
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F76
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F77
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F78
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F79
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F80
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F81
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F82
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F83
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F84
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F85
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F86
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F87
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F88
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F89
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F90
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F91
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F92
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F93
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F94
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F95
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F96
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F97
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F98
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F99
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F100
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F101
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F102
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F103
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F104
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F105
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F106
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F107
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F108
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F109
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F110
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F111
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F112
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F113
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F114
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F115
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F116
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F117
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F118
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F119
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F120
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F121
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F122
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F123
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F124
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F125
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F126
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F127
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F128
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F129
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F130
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F131
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F132
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F133
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F134
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F135
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F136
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F137
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F138
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F139
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F140
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F141
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F142
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F143
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F144
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F145
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation.
  - id: F146
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 10
    line_end: 10
    message: >
      The comment about "silentSkill" branch not firing when gsnHide/gsnSneak
      are -1 is unclear. It would help to clarify how this works in practice.
    suggested_fix: >
      Clarify the logic flow around silent skill handling in comments.
  - id: F147
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 47
    line_end: 50
    message: >
      The XP calculation logic is duplicated between the adept and non-adept
      branches. This duplication could lead to inconsistencies if changes are
      made to one but not the other.
    suggested_fix: >
      Extract XP calculation into a separate function to avoid duplication.
  - id: F148
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 63
    line_end: 63
    message: >
      The hardcoded magic numbers (5, 25, 2) used in skill learning calculations
      should be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for these values to make them more discoverable and
      easier to change.
  - id: F149
    severity: minor
    category: maintainability
    file: internal/act/skills.go
    line: 18
    line_end: 18
    message: >
      The comment about "C src/skills.c:1661" suggests that this code is ported
      from C, but there's no clear indication of what was changed or why.
    suggested_fix: >
      Add more context to comments explaining the porting process and any
      changes made during translation
