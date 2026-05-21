# Adversary Review

**Target**: `internal/act/skills2.go`
**Timestamp**: 2026-05-16T07:49:11Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

### Claim Verification
N/A - Skipping steps 0-2 as requested.

### Test Verification
N/A - Skipping step 2 as requested.

### Complexity Audit
The file contains three functions: `padName`, `DoSkills`, `DoSpells`, and `DoPractice`. The `DoPractice` function is particularly complex with multiple conditions and logic flow paths. The code uses string manipulation and formatting extensively, which could lead to performance issues if used in high-frequency scenarios.

### Scope Check
The file implements three functions related to skill and spell management for players in a game context. There's no evidence of additional features or files being modified beyond what's shown.

### Alternative Approach
Instead of using `strings.Builder` for building output buffers, the code uses `fmt.Fprintf` with `&buf`. A simpler alternative would be to use `fmt.Sprintf` directly with a single format string, though this might not be significantly simpler.

### Assumptions
1. `WorldRef.Skills` is properly initialized and populated before these functions are called.
2. `ch.PCData.Learned` has sufficient capacity to hold all skill indices.
3. `ch.Class` is valid and within bounds for accessing `SkillAdept`.
4. The `WorldRef` global variable is properly initialized and accessible.

### Security
No security issues found in the provided code.

```adversary-review
verdict: CONCERNS
confidence: medium
artifact:
  path: internal/act/skills2.go
  sha256: 8b5e1f9bd4ff3a83
  lines_reviewed: 1-150
findings:
  - id: F1
    severity: major
    category: performance
    file: internal/act/skills2.go
    line: 17
    line_end: 17
    message: >
      Using fmt.Fprintf with strings.Builder can be inefficient for large
      numbers of skills/spells. Consider pre-computing all skill strings and
      joining them once.
    suggested_fix: >
      Precompute skill strings and join them once instead of repeatedly
      appending to the buffer.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/skills2.go
    line: 35
    line_end: 35
    message: >
      Hardcoded padding width of 20 characters. This could be configurable
      or calculated dynamically based on longest skill name.
    suggested_fix: >
      Calculate the maximum skill name length and use that as the padding width.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/skills2.go
    line: 60
    line_end: 60
    message: >
      Magic number "75" used for default adept level. This should be defined
      as a constant with descriptive name.
    suggested_fix: >
      Define a constant like `DefaultAdeptLevel = 75` and use it here.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/act/skills2.go
    line: 68
    line_end: 68
    message: >
      The skill search logic uses prefix matching which might match multiple
      skills. If there are multiple matches, only the first one is chosen.
    suggested_fix: >
      Implement better error handling or user feedback when multiple matches
      exist.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/act/skills2.go
    line: 45
    line_end: 45
    message: >
      Hardcoded string "Skills/Spells:" - this could be extracted to a
      constant or variable for reuse.
    suggested_fix: >
      Define a constant like `skillsSpellsHeader = "Skills/Spells:"`
  - id: F6
    severity: minor
    category: maintainability
    file: internal/act/skills2.go
    line: 13
    line_end: 13
    message: >
      The padding function assumes fixed-width output but doesn't handle
      wide characters or Unicode properly.
    suggested_fix: >
      Use a more robust method for calculating character width that handles
      Unicode correctly.
```
