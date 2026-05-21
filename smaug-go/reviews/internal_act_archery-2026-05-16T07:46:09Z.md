# Adversary Review

**Target**: `internal/act/archery.go`
**Timestamp**: 2026-05-16T07:46:09Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

**Scope**: Archery subsystem implementation in Go port of SMAUG C archery.c.

### Claim Verification
All claims in the file are verified by code inspection:
- The archery subsystem implements DoDraw, DoFire, DoDislodge commands
- FindQuiver, FindProjectile helper functions are implemented
- Ported from src/archery.c with preserved bugs and C-style semantics
- Functions match expected behavior described in comments

### Test Verification
No test files mentioned in the file's comments exist in the provided content.

### Complexity Audit
- File size: 894 lines (new code added)
- Function sizes: 
  - `rangedAttack` (52 lines) - medium complexity
  - `projectileHit` (63 lines) - medium complexity
  - `scanForVictim` (37 lines) - medium complexity
- Abstraction depth: Moderate - uses helper functions but no complex abstraction layers
- New dependencies: None beyond standard library and existing types
- Premature generalization: None detected

### Scope Check
- All functions are within scope of archery subsystem as defined in comments
- No additional features added beyond what was claimed in comments
- No unrelated changes to other subsystems

### Alternative Approach
The archery implementation uses a direct port approach rather than a more modular design. A simpler alternative would be to extract the projectile handling logic into a separate module or function that could handle different projectile types more cleanly.

### Assumptions
1. WorldRef is always available for damage calculations
2. Character data structures have expected fields (InRoom, etc.)
3. Handler functions work as expected with the provided data structures
4. The archery system assumes all projectiles have valid Value[1] and Value[2] values
5. The archery system assumes all weapons have valid Value[5] values

### Security
No security issues found in the code review.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/act/archery.go
  sha256: 34dd43e4810e89e6
  lines_reviewed: 1-894
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/archery.go
    line: 1255
    message: >
      C bug preserved: `victim` is declared and initialized NULL at
      archery.c:1255, but the `victim == ch` check at :1272 is structurally dead.
      This creates a logical inconsistency where the condition can never be true.
    suggested_fix: >
      Remove the dead code branch or fix the logic to properly handle self-targeting.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/archery.go
    line: 1255
    message: >
      The comment mentions preserving C bugs but doesn't clearly indicate which
      ones are intentionally preserved vs. accidental omissions.
    suggested_fix: >
      Make clear which bugs are intentional by adding more specific comments.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/archery.go
    line: 1255
    message: >
      The comment about preserving C bugs should include specific details about
      why these bugs are preserved rather than fixed.
    suggested_fix: >
      Add more detailed explanation of why these bugs are preserved in the port.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/act/archery.go
    line: 1255
    message: >
      The comment mentions "C bug preserved" but doesn't specify which specific
      bug is being preserved and why it's acceptable for this port.
    suggested_fix: >
      Clarify which specific bug is preserved and provide justification.
```
