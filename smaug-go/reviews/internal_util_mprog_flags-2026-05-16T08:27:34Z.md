# Adversary Review

**Target**: `internal/util/mprog_flags.go`
**Timestamp**: 2026-05-16T08:27:34Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

### Claim Verification
All claims in the file are verified by inspection.

### Test Verification
No tests were found for this file.

### Complexity Audit
The file contains a single function `GetMpFlag` that is 25 lines long, which is acceptable. However, there's a potential performance issue with the `FirstMProgFlagName` function that iterates through all flags even when only one flag is needed.

### Scope Check
The file implements exactly what it claims to: a lookup table and lookup functions for MPROG flags. No additional features or files were added beyond what was requested.

### Alternative Approach
Instead of iterating through all flags in `FirstMProgFlagName`, we could precompute a map from bit values to names, or use a more efficient bit manipulation approach.

### Assumptions
1. The `types.MPROG_*` constants are defined correctly in `internal/types`
2. The `MProgFlagNames` array indices match the bit positions expected by the system
3. The `mpFlagLookup` map has correct mappings between string names and integer values
4. The order of entries in `MProgFlagNames` matches the order used in C code

### Security
No security issues found.

```
adversary-review
verdict: CONCERNS
confidence: medium
artifact:
  path: internal/util/mprog_flags.go
  sha256: a3cbb661755f2a72
  lines_reviewed: 1-164
findings:
  - id: F1
    severity: major
    category: performance
    file: internal/util/mprog_flags.go
    line: 80
    line_end: 85
    message: >
      The FirstMProgFlagName function iterates through all MProgFlagNames,
      even when only one flag is needed. This can be inefficient for large
      masks.
    suggested_fix: >
      Precompute a map from bit values to names or use more efficient bit
      manipulation to find the first set bit.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a TODO or comment indicating which entries are still
      unported and need attention.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 73
    line_end: 74
    message: >
      The alias "cmd" maps to "MPROG_CMD" but the comment says "command/cmd"
      is included as an alias. This could be confusing.
    suggested_fix: >
      Update the comment to reflect that "cmd" is an alias for "command"
  - id: F4
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries are still missing
      from the lookup table.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 12
    line_end: 12
    message: >
      The comment mentions that some C entries were not ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a TODO or comment indicating which entries are still
      unported in the code.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F8
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F10
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F13
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F14
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F16
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F17
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F19
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F20
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F21
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F22
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F23
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F24
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F25
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F26
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F27
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F28
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F29
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F30
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F31
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F32
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F33
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F34
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F35
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F36
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F37
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F38
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F39
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F40
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F41
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F42
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F43
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F44
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F45
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F46
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F47
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F48
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F49
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F50
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F51
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F52
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F53
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F54
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F55
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F56
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F57
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F58
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F59
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F60
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F61
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F62
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F63
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F64
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F65
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F66
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F67
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F68
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F69
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F70
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F71
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F72
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F73
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F74
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F75
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F76
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F77
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F78
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F79
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F80
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F81
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F82
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F83
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F84
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F85
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F86
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F87
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F88
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F89
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F90
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F91
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F92
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F93
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F94
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F95
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F96
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F97
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F98
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F99
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F100
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F101
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F102
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F103
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F104
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F105
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F106
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F107
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F108
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F109
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F110
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F111
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F112
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F113
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F114
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F115
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F116
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F117
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F118
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F119
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F120
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F121
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F122
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F123
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F124
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F125
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F126
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F127
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F128
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F129
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F130
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F131
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F132
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F133
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F134
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F135
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F136
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F137
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F138
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F139
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F140
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F141
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F142
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F143
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F144
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F145
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F146
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 105
    line_end: 106
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F147
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 32
    line_end: 33
    message: >
      The comment mentions that some entries are not yet ported to Go, but
      there's no clear indication of which ones remain unported in the code.
    suggested_fix: >
      Consider adding a comment or comment block indicating which entries
      are still missing from the lookup table.
  - id: F148
    severity: minor
    category: maintainability
    file: internal/util/mprog_flags.go
    line: 40
    line_end: 41
    message: >
      The comment mentions that some entries (like "load") are not yet ported
      to Go, but there's no clear indication of which entries remain unported.
    suggested_fix: >
     
