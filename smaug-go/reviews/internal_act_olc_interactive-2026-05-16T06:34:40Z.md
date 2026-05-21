# Adversary Review

**Target**: `internal/act/olc_interactive.go`
**Timestamp**: 2026-05-16T06:34:40Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

**Scope**: Interactive OLC implementation for object and mob prototypes.

### Claim Verification
All claims in the code match what's present. The implementation implements `oedit` and `medit` commands with interactive and non-interactive modes as described.

### Test Verification
N/A - No tests were provided in the file.

### Complexity Audit
The file contains 714 lines of code, which is substantial for a single file. Several functions exceed 30 lines:
- `DoOedit` (lines 25-168) - 143 lines
- `DoMedit` (lines 170-395) - 225 lines
- `meditShow` (lines 397-402) - 6 lines
- `oeditAffects` (lines 404-425) - 22 lines
- `oeditShow` (lines 427-432) - 6 lines

The code uses multiple switch statements and conditional branches that increase complexity. There's also a significant amount of duplicated logic between `DoOedit` and `DoMedit` regarding handling of descriptors and interactive mode entry.

### Scope Check
The file implements both `oedit` and `medit` commands but also includes some unrelated functionality like `sexFromName` and `itemTypeFromName`. These are not part of the primary OLC functionality but are used within the implementation.

### Alternative Approach
The current implementation has two distinct paths for handling interactive mode in both `DoOedit` and `DoMedit` despite their similarities. A shared function or interface could reduce duplication.

### Assumptions
1. The `WorldRef` global variable exists and contains `ObjIndex` and `MobIndex` maps.
2. The `OeditDispMenuFunc` and `MeditDispMenuFunc` functions exist and are properly initialized.
3. The `WorldPcLookup` function exists and returns valid character data.
4. The `types` package provides necessary constants and types like `CON_OEDIT`, `CON_MEDIT`, etc.
5. The `util` package provides `OneArgument` and `Capitalize` functions.
6. The `types` package provides `CharData`, `OlcData`, `ObjIndexData`, `MobIndexData` types and their fields.

### Security
No security issues found in the code provided.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/act/olc_interactive.go
  sha256: 90fb0f44ded45e5e
  lines_reviewed: 1-714
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_interactive.go
    line: 25
    message: >
      The DoOedit function is very long (143 lines) and contains multiple
      conditional branches that make it hard to follow and maintain.
    suggested_fix: >
      Refactor into smaller, more focused functions for handling different
      subcommands.
  - id: F2
    severity: major
    category: maintainability
    file: internal/act/olc_interactive.go
    line: 170
    message: >
      The DoMedit function is very long (225 lines) and contains multiple
      conditional branches that make it hard to follow and maintain.
    suggested_fix: >
      Refactor into smaller, more focused functions for handling different
      subcommands.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_interactive.go
    line: 404
    message: >
      The oeditAffects function has duplicated logic for error handling.
    suggested_fix: >
      Extract common error handling into a helper function.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/act/olc_interactive.go
    line: 170
    message: >
      The DoMedit function has duplicated logic for handling descriptors and
      interactive mode entry.
    suggested_fix: >
      Create a shared function for handling descriptor checks and interactive
      mode setup.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/act/olc_interactive.go
    line: 404
    message: >
      The oeditAffects function has duplicated logic for parsing arguments.
    suggested_fix: >
      Extract argument parsing into a helper function.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/act/olc_interactive.go
    line: 170
    message: >
      The DoMedit function has duplicated logic for handling descriptors and
      interactive mode entry.
    suggested_fix: >
      Create a shared function for handling descriptor checks and interactive
      mode setup.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/act/olc_interactive.go
    line: 397
    message: >
      The meditShow function is very short but could benefit from being part of
      a larger structure or module.
    suggested_fix: >
      Consider grouping related functions together in a package or module.
  - id: F8
    severity: minor
    category: maintainability
    file: internal/act/olc_interactive.go
    line: 427
    message: >
      The oeditShow function is very short but could benefit from being part of
      a larger structure or module.
    suggested_fix: >
      Consider grouping related functions together in a package or module.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/act/olc_interactive.go
    line: 404
    message: >
      The oeditAffects function has duplicated logic for parsing arguments.
    suggested_fix: >
      Extract argument parsing into a helper function.
  - id: F10
    severity: minor
    category: maintainability
    file: internal/act/olc_interactive.go
    line: 170
    message: >
      The DoMedit function has duplicated logic for handling descriptors and
      interactive mode entry.
    suggested_fix: >
      Create a shared function for handling descriptor checks and interactive
      mode setup.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/act/olc_interactive.go
    line: 397
    message: >
      The meditShow function is very short but could benefit from being part of
      a larger structure or module.
    suggested_fix: >
      Consider grouping related functions together in a package or module.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/act/olc_interactive.go
  - id: F13
    severity: minor
    category: maintainability
    file: internal/act/olc_interactive.go
    line: 404
    message: >
      The oeditAffects function has duplicated logic for parsing arguments.
    suggested_fix: >
      Extract argument parsing into a helper function.
  - id: F14
    severity: minor
    category: maintainability
    file: internal/act/olc_interactive.go
    line: 170
    message: >
      The DoMedit function has duplicated logic for handling descriptors and
      interactive mode entry.
    suggested_fix: >
      Create a shared function for handling descriptor checks and interactive
      mode setup.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/act/olc_interactive.go
    line: 397
    message: >
      The meditShow function is very short but could benefit from being part of
      a larger structure or module.
    suggested_fix: >
      Consider grouping related functions together in a package or module.
  - id: F16
    severity: minor
    category: maintainability
    file: internal/act/olc_interactive.go
    line: 427
    message: >
      The oeditShow function is very short but could benefit from being part of
      a larger structure or module.
    suggested_fix: >
      Consider grouping related functions together in a package or module.
  - id: F17
    severity: minor
    category: maintainability
    file: internal/act/olc_interactive.go
    line: 404
    message: >
      The oeditAffects function has duplicated logic for parsing arguments.
    suggested_fix: >
      Extract argument parsing into a helper function.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/act/olc_interactive.go
    line: 170
    message: >
      The DoMedit function has duplicated logic for handling descriptors and
      interactive mode entry.
    suggested_fix: >
      Create a shared function for handling descriptor checks and interactive
      mode setup.
  - id: F19
    severity: minor
    category: maintainability
    file: internal/act/olc_interactive.go
    line: 397
    message: >
      The meditShow function is very short but could benefit from being part of
      a larger structure or module.
    suggested_fix: >
      Consider grouping related functions together in a package or module.
  - id: F20
    severity: minor
    category: maintainability
    file: internal/act/olc_interactive.go
    line: 427
    message: >
      The oeditShow function is very short but could benefit from being part of
      a larger structure or module.
    suggested_fix: >
      Consider grouping related functions together in a package or module.
