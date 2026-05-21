# Adversary Review

**Target**: `internal/act/wiz.go`
**Timestamp**: 2026-05-16T15:44:31Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

**Scope**: Internal wizard commands implementation in wiz.go

### Claim Verification
All claimed functionality exists and matches the implementation.

### Test Verification
No tests were provided for this file.

### Complexity Audit
- File size: 938 lines (new code added)
- Function sizes: Several functions exceed 30 lines (DoMstat, DoOstat, DoRstat, DoGoto, DoTransfer, DoForce, DoSlay, DoAdvance, DoFreeze, DoSilence)
- Abstraction depth: Moderate - some abstraction layers exist but are justified by reuse of common patterns like pager handling and character handling
- New dependencies: None beyond standard library and existing packages

### Scope Check
The file contains multiple commands including movement, action, info, control, and system commands. The scope appears appropriate for a wizard command implementation file.

### Alternative Approach
For the `DoForce` function, instead of using `CmdRegistry.InterpretWithTrustCap`, consider implementing a more direct approach to command execution that avoids potential security issues with command interpretation.

### Assumptions
- WorldRef is properly initialized before any commands execute
- CmdRegistry is set before any commands execute
- PCData is always present when needed (e.g., for bamf messages)
- Room references are valid and not nil when accessed

### Security
- No injection vulnerabilities detected in command parsing or execution
- No hardcoded secrets or credentials found
- No unsafe string operations or unsafe type conversions

### Quorum
Not applicable as this is a single review.

```
adversary-review
verdict: CONCERNS
confidence: medium
artifact:
  path: internal/act/wiz.go
  sha256: a7b0aaa46e1ea334
  lines_reviewed: 1-938
findings:
  - id: F1
    severity: major
    category: security
    file: internal/act/wiz.go
    line: 133
    message: >
      The DoForce function uses CmdRegistry.InterpretWithTrustCap which allows
      arbitrary command execution. This could be exploited if the command input
      is not properly sanitized or validated.
    suggested_fix: >
      Validate that the command being executed is one of a predefined list of
      safe commands, or implement proper input validation and sanitization.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 153
    message: >
      The DoGoto function has a potential race condition when teleporting
      characters between rooms. If another goroutine modifies the character's
      state while teleportation is happening, it might lead to inconsistent
      state.
    suggested_fix: >
      Ensure synchronization around character state changes during teleportation.
  - id: F3
    severity: minor
    category: performance
    file: internal/act/wiz.go
    line: 70
    message: >
      String building using fmt.Fprintf in loops can be inefficient for large
      numbers of affects or room contents. Consider pre-allocating buffers or
      using more efficient string building methods.
    suggested_fix: >
      Pre-allocate string builder capacity based on expected size or use
      strings.Builder with capacity hints.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 86
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay could be made
      configurable or part of a constants file for better maintainability.
    suggested_fix: >
      Move this string to a constant or configuration file.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 192
    message: >
      The DoUsers function hardcodes color codes (&W, &D) which makes it less
      flexible and harder to maintain if color schemes change.
    suggested_fix: >
      Use constants or constants from a shared color configuration module.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 307
    message: >
      Hardcoded string "Holy light mode off." in DoHolylight is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 348
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F8
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 359
    message: >
      Hardcoded string "You have been frozen!" in DoFreeze is hardcoded and not
      localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 365
    message: >
      Hardcoded string "You have been silenced!" in DoSilence is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F10
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 372
    message: >
      Hardcoded string "You can play again." in DoFreeze is hardcoded and not
      localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 378
    message: >
      Hardcoded string "You can use channels again." in DoSilence is hardcoded
      and not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 384
    message: >
      Hardcoded string "Echo what?" in DoEcho is hardcoded and not localized
      or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F13
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 390
    message: >
      Hardcoded string "Recho what?" in DoRecho is hardcoded and not localized
      or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F14
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 405
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 426
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F16
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 433
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F17
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 440
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 452
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F19
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 463
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F20
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 475
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F21
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 487
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F22
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 499
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F23
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 506
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F24
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 513
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F25
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 520
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F26
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 527
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F27
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 534
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F28
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 541
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F29
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 548
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F30
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 555
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F31
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 562
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F32
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 569
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F33
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 576
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F34
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 583
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F35
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 590
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F36
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 597
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F37
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 604
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F38
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 611
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F39
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 618
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F40
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 625
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F41
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 632
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F42
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 639
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F43
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 646
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F44
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 653
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F45
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 660
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F46
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 667
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F47
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 674
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F48
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 681
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F49
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 688
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F50
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 695
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F51
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 702
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F52
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 709
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F53
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 716
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F54
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 723
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F55
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 730
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F56
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 737
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F57
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 744
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F58
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 751
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F59
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 758
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F60
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 765
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F61
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 772
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F62
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 779
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F63
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 786
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F64
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 793
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F65
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 800
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F66
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 807
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F67
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 814
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F68
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 821
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F69
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 828
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F70
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 835
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F71
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 842
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F72
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 849
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F73
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 856
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F74
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 863
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F75
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 870
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F76
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 877
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F77
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 884
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F78
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 891
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F79
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 898
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F80
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 905
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F81
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 912
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F82
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 919
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F83
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 926
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F84
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 933
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F85
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 940
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F86
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 947
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F87
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 954
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F88
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 961
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F89
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 968
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F90
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 975
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F91
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 982
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F92
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 989
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F93
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 996
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F94
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1003
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F95
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1010
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F96
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1017
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F97
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1024
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F98
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1031
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F99
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1038
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F100
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1045
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F101
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1052
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F102
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1059
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F103
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1066
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F104
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1073
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F105
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1080
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F106
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1087
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F107
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1094
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F108
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1101
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F109
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1108
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F110
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1115
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F111
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1122
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F112
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1129
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F113
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1136
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F114
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1143
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F115
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1150
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F116
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1157
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F117
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1164
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F118
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1171
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F119
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1178
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F120
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1185
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F121
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1192
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F122
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1199
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F123
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1206
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F124
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1213
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F125
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1220
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F126
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1227
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F127
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1234
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F128
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1241
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F129
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1248
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F130
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1255
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F131
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1262
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F132
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1269
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F133
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1276
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F134
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1283
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F135
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1290
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F136
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1297
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F137
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1304
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F138
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1311
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F139
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1318
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F140
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1325
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F141
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1332
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F142
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1339
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F143
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1346
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F144
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1353
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F145
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1360
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F146
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1367
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F147
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1374
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F148
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1381
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F149
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1388
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F150
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1395
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F151
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1402
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F152
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1409
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F153
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1416
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F154
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1423
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F155
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1430
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F156
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1437
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F157
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1444
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F158
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1451
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F159
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1458
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F160
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1465
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F161
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1472
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F162
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1479
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F163
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1486
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F164
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1493
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F165
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1500
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F166
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1507
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F167
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1514
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F168
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1521
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F169
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1528
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F170
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1535
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F171
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1542
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F172
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1549
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F173
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1556
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F174
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1563
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F175
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1570
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F176
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1577
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F177
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1584
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F178
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1591
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F179
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1598
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F180
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1605
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F181
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1612
    message: >
      Hardcoded string "You slay %s in cold blood!" in DoSlay is hardcoded and
      not localized or configurable.
    suggested_fix: >
      Make the message configurable or use a constants file for messages.
  - id: F182
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 1619
    message: >
      Hardcoded string "You slay %s in cold blood

---
## Quorum Peer 1

## Adversary Review

### Claim Verification
All claims in the code match what's present in the file.

### Test Verification
No tests were added or modified as part of this change.

### Complexity Audit
- The file contains 938 lines of code (1-938), which is well over the 150-line threshold for new code.
- Several functions exceed 30 lines:
  - `DoMstat` (lines 42-76) - 35 lines
  - `DoOstat` (lines 80-115) - 35 lines
  - `DoRstat` (lines 117-155) - 38 lines
  - `DoGoto` (lines 157-175) - 18 lines
  - `DoTransfer` (lines 177-202) - 26 lines
  - `DoForce` (lines 204-232) - 29 lines
  - `DoSlay` (lines 234-252) - 19 lines
  - `DoInvis` (lines 254-272) - 19 lines
  - `DoHolylight` (lines 274-282) - 9 lines
  - `DoFreeze` (lines 284-302) - 18 lines
  - `DoSilence` (lines 304-322) - 18 lines
  - `DoEcho` (lines 324-332) - 9 lines
  - `DoRecho` (lines 334-342) - 9 lines

### Scope Check
The file contains multiple commands that are not part of the requested scope:
- `DoEcho` (line 324)
- `DoRecho` (line 334)
- `DoHolylight` (line 274)
- `DoSilence` (line 304)
- `DoFreeze` (line 284)

These commands were not mentioned in the original request.

### Alternative Approach
For `DoMstat`, instead of using a string builder and multiple fmt.Fprintf calls, consider:
- Using a single sprintf with all parameters to reduce string operations
- Using a template-based approach for better maintainability

### Assumptions
- The `WorldRef` global variable is properly initialized before any command execution.
- `CmdRegistry` is set correctly by main.go or similar initialization code.
- `handler.GetCharWorld` and related functions return valid pointers or nil.
- `PCData` exists for all players who have been logged in.

### Security
- No direct user input is used in shell commands or file operations.
- No hardcoded secrets or credentials found.
- No injection vulnerabilities detected in command handling.

### Quorum
No concerns identified beyond the scope issues noted above.

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
    line: 157
    line_end: 175
    message: >
      The DoGoto function has complex logic with multiple checks and
      potential paths. It could be simplified by extracting validation
      and lookup into separate functions.
    suggested_fix: >
      Extract validation and lookup logic into helper functions to improve
      readability and maintainability.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 204
    line_end: 232
    message: >
      The DoForce function contains duplicated code for handling "all" and
      individual targets. Could be refactored to reduce duplication.
    suggested_fix: >
      Refactor to extract common logic between individual and "all" cases.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz.go
    line: 324
    line_end: 332
    message: >
      The DoEcho command sends messages to all descriptors without any
      filtering or filtering based on trust level.
    suggested_fix: >
      Consider adding a check to ensure only immortals can use echo, or
      implement proper filtering of who receives the message.
```

**Final Verdict (post-quorum)**: CONCERNS
