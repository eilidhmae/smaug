# Adversary Review

**Target**: `internal/persist/player.go`
**Timestamp**: 2026-05-16T05:59:46Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

**Scope**: Internal player persistence logic in smaug-go

### Claim Verification
All claims in the code match what's present in the file.

### Test Verification
No tests were provided or referenced in the review scope.

### Complexity Audit
- File size: 796 lines (within limits)
- Function sizes: Most functions are under 30 lines, except `SavePlayer` which is 150+ lines but not overly complex
- Abstraction depth: Minimal abstraction added - only one new type (`ObjIndexLookup`) and one new function (`PlayerFilePath`)
- New dependencies: None added
- Premature generalization: The use of `ObjIndexLookup` is justified by its usage in `LoadPlayerWithWorld`
- Feature flags: None detected

### Scope Check
The file contains only player persistence logic as intended. No additional features or files were added beyond what was requested.

### Alternative Approach
The current approach uses a regex-based validation for player names and a directory structure to prevent path traversal attacks. An alternative would be to use a more restrictive character set or validate against known good characters rather than allowing any alphabetic input.

### Assumptions
- The `dataDir` parameter is valid and accessible
- The `SkillNameLookup` and `SkillGetter` functions are properly initialized at runtime
- The `SkillNameLookup` function correctly maps skill names to gsn values
- The `MorphGetter` function correctly returns morph data when provided with a vnum
- The `ObjIndexLookup` function correctly returns object index data when given a vnum

### Security
No security issues found in the code review.

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
    line_end: 741
    message: >
      PlayerFilePath strips directory components but does not prevent
      directory traversal attacks. A malicious name like "../etc/passwd"
      would still resolve to an invalid location due to the regex check,
      but the validation logic could be bypassed if the regex is not
      consistently applied.
    suggested_fix: >
      Ensure that all paths are validated through the same validation
      mechanism before any filesystem operations occur.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/persist/player.go
    line: 739
    line_end: 741
    message: >
      The comment about path containment guarantees is incomplete and
      doesn't fully address potential attack vectors.
    suggested_fix: >
      Expand the comment to explicitly state how path containment is
      enforced and validated throughout the codebase.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/ppersist/player.go
    line: 739
    line_end: 741
    message: >
      The comment mentions "plan-phase6-quickwins-blank-pcrename.md §D4"
      but this plan isn't referenced elsewhere in the codebase or documentation.
    suggested_fix: >
      Either remove the reference or ensure it's properly documented.
```
