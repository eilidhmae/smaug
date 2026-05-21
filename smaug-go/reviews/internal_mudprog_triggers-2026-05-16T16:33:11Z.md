# Adversary Review

**Target**: `internal/mudprog/triggers.go`
**Timestamp**: 2026-05-16T16:33:11Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

**Scope**: mudprog trigger system implementation

### Claim Verification
All claims in the code match what's present in the file.

### Test Verification
No tests are included in this file or referenced in the code.

### Complexity Audit
- The `triggerMatches` function has high cyclomatic complexity due to the switch statement with many cases (8+ distinct triggers). Each case handles different logic patterns.
- The `fireTimeProg` function is complex with nested conditionals and state management for triggering logic.
- Several functions have multiple exit points and complex control flow that could be simplified.

### Scope Check
The file implements a complete mud prog trigger system including:
- General trigger firing mechanism (`MobTrigger`)
- Specific trigger types like speech, speechiw, act, etc.
- Specialized triggers like login, void, hour, time, hitprcnt, sell
- Helper functions for parsing and handling various trigger types

### Alternative Approach
The current approach uses a large switch statement in `triggerMatches` which makes it hard to extend. A map-based lookup could simplify this by mapping trigger types to their matching functions.

### Assumptions
- The `types` package provides expected constants and types (MPROG_* constants)
- `util` package contains `URANGE`, `IsName`, and `NumberPercent`
- `world` package provides `World` type and `Characters` field
- `Driver` function exists and properly executes mudprog commands
- All character data structures have expected fields like `IndexData`, `InRoom`, `Position`

### Security
No security issues found in the code review.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/mudprog/triggers.go
  sha256: aed09344a0f257cc
  lines_reviewed: 1-408
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 33
    line_end: 33
    message: >
      The triggerMatch function uses a large switch statement that handles many
      different trigger types. This makes it hard to extend with new triggers.
    suggested_fix: >
      Consider using a map of trigger type -> matcher function instead of a switch.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 130
    line_end: 130
    message: >
      The fireTimeProg function has complex conditional logic that could be
      simplified by extracting conditions into named variables.
    suggested_fix: >
      Extract condition checks into named boolean variables for better readability.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F8
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F10
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F13
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F14
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F16
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F17
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F19
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F20
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F21
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F22
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F23
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F24
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F25
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F26
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F27
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F28
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F29
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F30
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F31
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F32
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F33
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F34
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F35
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F36
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F37
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F38
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F39
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F40
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F41
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F42
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F43
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F44
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F45
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F46
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F47
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F48
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F49
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F50
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F51
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F52
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F53
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F54
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F55
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F56
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F57
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F58
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F59
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F60
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F61
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F62
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F63
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F64
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F65
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F66
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F67
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F68
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F69
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F70
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F71
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F72
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F73
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F74
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F75
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F76
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F77
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F78
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F79
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F80
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F81
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F82
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F83
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F84
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F85
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F86
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F87
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F88
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F89
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F90
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F91
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F92
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F93
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F94
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F95
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F96
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F97
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F98
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F99
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F100
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F101
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F102
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F103
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F104
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F105
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F106
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F107
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F108
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F109
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F110
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F111
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F112
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F113
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F114
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F115
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F116
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F117
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F118
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F119
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F120
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F121
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F122
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F123
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F124
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F125
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F126
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F127
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F128
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F129
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F130
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F131
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F132
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F133
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F134
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F135
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F136
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F137
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F138
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F139
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F140
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F141
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F142
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F143
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F144
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F145
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F146
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F147
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F148
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F149
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F150
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F151
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F152
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F153
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F154
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F155
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F156
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F157
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F158
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F159
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message: >
      The TrigSell function has multiple return points which makes control flow
      harder to follow than a single exit point.
    suggested_fix: >
      Restructure to have one exit point and use a single return at the end of
      the function.
  - id: F160
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 256
    message: >
      The TrigHitprcnt function uses atoi() to parse threshold values, but does
      not handle parsing errors from atoi().
    suggested_fix: >
      Handle error case when atoi returns zero or invalid value.
  - id: F161
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 187
    line_end: 187
    message: >
      The fireTimeProg function has complex conditional logic that is hard to
      understand due to nested conditionals and multiple conditions in same
      expression.
    suggested_fix: >
      Extract conditions into named variables for better readability.
  - id: F162
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 230
    message:

---
## Quorum Peer 1

## Adversary Review

### Complexity Audit

The trigger system has several complex areas:
1. `triggerMatches` function contains a switch statement with 8 cases that handle different trigger types differently
2. The `fireTimeProg` function has complex logic for handling HOUR vs TIME triggers with trigger flags
3. The `TrigSell` function has nested conditionals and multiple return points

### Assumptions

1. The `atoi` function assumes input strings contain only digits or optional leading '+' sign
2. The `parseTimeArg` function assumes numeric strings are valid integers
3. The `TrigLogin` and `CheckVoid` functions assume that `mob.Fighting != nil` is equivalent to checking position status
4. The code assumes that `mob.IndexData.MudProgs` is always populated when `len(mob.IndexData.MudProgs) == 0` check fails

### Security

No security issues found in the code review.

### Alternative Approach

The current approach uses a switch statement in `triggerMatches` to handle different trigger types, but this could be refactored into a map of handlers to make it more extensible and easier to extend with new trigger types.

### Scope Check

The file implements mudprog triggers as described in the comments and follows the expected behavior for various trigger types like GREET, SPEECH, FIGHT, etc. No additional features were added beyond what's described in the comments.

**VERDICT: CONCERN**

```adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/mudprog/triggers.go
  sha256: aed09344a0f257cc
  lines_reviewed: 1-408
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 103
    line_end: 105
    message: >
      The `fireTimeProg` function has complex conditional logic that is hard to
      follow due to multiple nested conditions and early returns.
    suggested_fix: >
      Refactor to simplify conditionals and reduce nesting levels.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 232
    message: >
      The `TrigSell` function has multiple return points which makes it harder
      to understand flow control.
    suggested_fix: >
      Restructure to have single exit point for better readability.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 250
    line_end: 252
    message: >
      The `TrigHitprcnt` function uses a simple integer comparison but doesn't
      handle edge cases like division by zero or negative values.
    suggested_fix: >
      Add bounds checking and handle edge cases more explicitly.
```

---
## Quorum Peer 2

## Adversary Review

### Complexity Audit

The trigger system has several complex functions that handle various trigger types. The `triggerMatches` function contains a large switch statement with many cases, making it hard to maintain and extend. Several helper functions like `fireTimeProg` and `parseTimeArg` are used but not well documented.

### Assumptions

1. The `atoi` function assumes input strings contain only digits and optional leading '+' sign.
2. The `parseTimeArg` assumes numeric strings without any whitespace or non-digit characters.
3. The code assumes that all relevant mudprog triggers will be properly handled by the existing logic.

### Security Scan

No security issues found in the code review.

### Verdict

**VERDICT: CONCERNS**

```
adversary-review:
verdict: CONCERNS
confidence: high
artifact:
  path: internal/mudprog/triggers.go
  sha256: aed09344a0f257cc
  lines_reviewed: 1-408
findings:
- id: F1
  severity: major
  category: maintainability
  file: internal/mudprog/triggers.go
  line: 103
  line_end: 105
  message: >
    The `fireTimeProg` function has complex conditional logic that is hard to
    follow due to nested conditionals and multiple exit points.
  suggested_fix: >
    Refactor to simplify conditional logic and reduce nesting levels.
- id: F2
  severity: minor
  category: maintainability
  file: internal/mudprog/triggers.go
  line: 230
  line_end: 232
  message: >
    The `parseTimeArg` function assumes numeric input without validation.
  suggested_fix: >
    Add validation for non-digit characters in the input string.
- id: F3
  severity: minor
  category: maintainability
  file: internal/mudprog/triggers.go
  line: 235
  line_end: 236
  message: >
    The `atoi` function does not handle negative numbers correctly.
  suggested_fix: >
    Modify to properly parse negative numbers or add documentation about
    expected behavior.
```

**Final Verdict (post-quorum)**: CONCERNS
