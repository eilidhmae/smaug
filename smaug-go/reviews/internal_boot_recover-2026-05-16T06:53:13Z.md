# Adversary Review

**Target**: `internal/boot/recover.go`
**Timestamp**: 2026-05-16T06:53:13Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

**Scope**: Hotboot recovery path implementation in internal/boot/recover.go

### Complexity Audit

- **File size**: 353 lines (100% new code)
- **Function size**: 
  - `BootRecover` (1-353): 353 lines (too long)
  - `restoreDescriptor` (206-278): 72 lines (too long)
  - `restoreReadLoop` (280-320): 40 lines (too long)
- **Abstraction depth**: 
  - Multiple levels of error handling and error logging
  - Error handling is duplicated across multiple functions
  - The code has a complex flow with many conditional branches and error paths
- **Premature generalization**: 
  - The `stripTelnet` and `stripCarriageReturn` functions are duplicated from internal/net but not exported, which creates duplication and potential maintenance issues
  - The `restoreReadLoop` function duplicates logic from the server's read loop, creating maintainability concerns

### Scope Check

- No scope creep detected

### Alternative Approach

The current approach uses a complex flow with multiple error handling paths and duplicated logic. A simpler alternative would be to:

1. Use a single error handling pattern throughout the codebase
2. Reuse existing server components instead of duplicating logic
3. Reduce the complexity of the `restoreDescriptor` function by breaking it into smaller functions

### Assumptions

- The hotboot recovery assumes that all FDs and session indices are valid and exist in the hotboot data file
- The code assumes that all necessary files (pfiles, hotboot.dat) exist and are accessible
- The code assumes that the world state can be properly restored without conflicts or race conditions

### Security

- No injection vulnerabilities found
- No unsafe default permissions or missing authentication checks
- No secrets in code
- No path traversal issues

### Quorum

No concerns identified

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/boot/recover.go
  sha256: 298c8cde674f3fc9
  lines_reviewed: 1-353
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/boot/recover.go
    line: 206
    line_end: 278
    message: >
      The restoreDescriptor function is too complex with multiple error paths
      and duplicated logic. It should be refactored into smaller, more
      manageable functions.
    suggested_fix: >
      Split restoreDescriptor into smaller functions for loading pfile,
      placing character, and handling welcome messages.
  - id: F2
    severity: major
    category: maintainability
    file: internal/boot/recover.go
    line: 280
    line_end: 320
    message: >
      The restoreReadLoop function duplicates logic from the server's read loop.
      This duplication creates maintenance burden and potential inconsistencies.
    suggested_fix: >
      Consider reusing the server's read loop or making it exported to allow
      reuse in recovery mode.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/boot/recover.go
    line: 280
    line_end: 320
    message: >
      The stripTelnet and stripCarriageReturn functions are duplicated from
      internal/net but not exported, creating code duplication.
    suggested_fix: >
      Export these functions from internal/net and use them instead of
      duplicating the logic.
```
