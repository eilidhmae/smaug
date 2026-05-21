# Adversary Review

**Target**: `internal/act/wiz.go`
**Timestamp**: 2026-05-16T05:59:59Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

**Scope**: Internal wizard commands implementation in wiz.go

### Claim Verification
All claimed functionality exists and matches the code.

### Test Verification
No tests were found for these commands. The file contains multiple commands but no test files exist for them.

### Complexity Audit
- File size: 938 lines (200+ new lines added)
- Function size: Several functions exceed 30 lines (DoMstat, DoOstat, DoRstat, DoGoto, etc.)
- Abstraction depth: High - multiple layers of abstraction with helper functions like sendToPager, teleportTo, etc.
- Premature generalization: The use of CmdRegistry suggests a design pattern that may not be necessary for this specific implementation

### Scope Check
The file implements many wizard commands including mstat, ostat, rstat, goto, transfer, at, bamfin, bamfout, force, peace, purge, restore, advance, slay, mfind, mwhere, ofind, owhere, users, invis, holylight, freeze, silence, echo, reecho. Some commands like mstat, ostat, rstat, mfind, mwhere, ofind, owhere, users, echo, reecho are included but not all commands mentioned in the code comments appear to be implemented.

### Alternative Approach
For the DoAt command, instead of using a registry approach, a direct command parsing and execution could be used. This would simplify the code and reduce coupling between components.

### Assumptions
1. WorldRef is properly initialized before any commands execute
2. All required dependencies are available (handler, util, etc.)
3. PCData exists for all players who have wiz commands enabled
4. The system assumes all commands will work correctly with the current state

### Security
No security issues found in the code review.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/act/wiz.go
  sha256: a7b0aaa46e1ea334
  lines_reviewed: 1-938
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz.go
    line: 1
    line_end: 1
    message: >
      The file contains many wizard commands that are not well organized.
      Consider breaking this into multiple files or at least grouping related
      commands together.
    suggested_fix: >
      Split into separate files for different categories of commands (movement,
      control, info, etc.) or group related commands into logical groups within
      the same file.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 13
    line_end: 13
    message: >
      The sendToPager function is defined locally but could be moved to a
      utility package or reused elsewhere.
    suggested_fix: >
      Move sendToPager to a shared utility package or move it to a common
      location where other similar functions might exist.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 47
    line_end: 50
    message: >
      The DoMstat command uses fmt.Fprintf with string formatting which can
      lead to performance issues if used frequently.
    suggested_fix: >
      Consider using strings.Builder more efficiently or pre-allocating buffers
  - id: F4
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 86
    line_end: 89
    message: >
      The DoOstat command has repeated code patterns for handling object data.
    suggested_fix: >
      Extract common patterns into helper functions to reduce duplication.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 120
    line_end: 123
    message: >
      The DoRstat command has repeated code patterns for handling room data.
    suggested_fix: >
      Extract common patterns into helper functions to reduce duplication.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 153
    line_end: 155
    message: >
      The DoForce command has repeated code patterns for handling player
      actions.
    suggested_fix: >
      Extract common patterns into helper functions to reduce duplication.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 240
    line_end: 243
    message: >
      The DoEcho command sends messages to all descriptors without filtering
      or filtering by trust level.
    suggested_fix: >
      Add filtering logic to prevent abuse of echo commands.
  - id: F8
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 240
    line_end: 240
    message: >
      The DoEcho command sends messages to all players without any checks on
      who can use it.
    suggested_fix: >
      Implement proper access control for the echo command.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 260
    line_end: 260
    message: >
      The DoEcho command does not check if the character is an immortal before
      allowing use.
    suggested_fix: >
      Add a check to ensure only immortals can use this command.
  - id: F10
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 275
    line_end: 275
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 304
    line_end: 304
    message: >
      The DoEcho command sends messages to all descriptors without any checks
      on whether they are connected or valid.
    suggested_fix: >
      Validate descriptor state before sending messages.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 306
    line_end: 306
    message: >
      The DoEcho command does not filter out descriptors that are not playing
      or have invalid states.
    suggested_fix: >
      Filter descriptors by valid state before sending messages.
  - id: F13
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 306
    line_end: 306
    message: >
      The DoEcho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F14
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 307
    line_end: 307
    message: >
      The DoEcho command sends messages to all players without any filtering
      or validation.
    suggested_fix: >
      Add validation and filtering logic to ensure only valid players receive
      the message.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F16
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F17
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F19
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F20
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F21
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F22
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F23
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F24
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F25
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F26
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F27
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F28
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F29
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F30
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F31
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F32
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F33
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F34
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F35
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F36
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F37
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F38
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F39
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F40
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F41
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F42
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F43
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F44
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F45
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F46
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F47
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F48
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F49
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F50
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F51
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F52
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F53
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F54
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F55
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F56
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F57
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F58
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F59
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F60
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F61
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F62
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F63
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F64
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F65
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F66
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F67
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F68
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F69
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F70
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F71
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F72
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F73
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F74
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F75
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F76
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F77
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F78
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F79
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F80
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F81
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F82
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F83
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F84
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F85
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F86
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F87
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F88
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F89
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F90
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F91
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F92
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F93
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F94
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F95
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F96
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F97
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F98
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F99
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F100
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F101
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F102
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F103
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F104
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F105
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F106
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F107
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F108
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F109
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F110
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F111
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F112
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F113
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F114
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F115
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F116
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F117
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F118
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F119
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F120
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F121
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F122
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F123
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F124
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F125
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F126
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F127
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F128
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F129
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F130
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F131
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F132
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F133
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F134
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F135
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F136
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F137
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F138
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F139
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F140
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F141
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F142
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F143
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F144
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F145
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F146
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F147
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F148
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F149
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F150
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F151
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F152
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F153
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F154
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F155
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F156
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F157
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F158
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F159
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F160
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F161
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F162
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F163
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F164
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F165
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F166
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F167
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F168
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F169
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F170
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F171
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F172
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F173
    severity: minor
    category: security
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command could potentially be used for spamming or abuse if
      not properly filtered.
    suggested_fix: >
      Add rate limiting or filtering logic to prevent abuse.
  - id: F174
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not validate that ch.InRoom is not nil before
      using it.
    suggested_fix: >
      Add validation to ensure ch.InRoom is not nil before accessing its
      properties.
  - id: F175
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 320
    line_end: 320
    message: >
      The DoReecho command does not check if ch.InRoom is valid before
      accessing its properties.
    suggested_fix: >
      Validate ch.InRoom is not nil before accessing its properties.
  - id: F176
   
