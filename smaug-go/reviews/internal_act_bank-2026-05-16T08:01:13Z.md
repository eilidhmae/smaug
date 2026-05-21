# Adversary Review

**Target**: `internal/act/bank.go`
**Timestamp**: 2026-05-16T08:01:13Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

### Claim Verification
N/A - This is a single file review.

### Test Verification
N/A - This is a single file review.

### Complexity Audit
The code has several issues:

1. **Duplicate logic**: The deposit and withdraw sections both contain nearly identical code blocks for parsing the amount argument and validating it.
2. **Magic numbers**: Hardcoded values like "0" and "1" are used without explanation.
3. **Code duplication**: Both deposit and withdraw sections handle input validation and error handling in nearly identical ways.

### Scope Check
This file contains only one function (`DoBank`) which handles bank operations. No additional features or files were added beyond what's shown.

### Alternative Approach
A better approach would be to extract common validation logic into helper functions and reduce code duplication. For example:
- Extract amount parsing logic into a shared function
- Use a single validation block instead of duplicate validation blocks
- Consider using a map-based switch for command handling

### Assumptions
1. `ch` is never nil (assumes proper initialization)
2. `ch.PCData.GBalance` and `ch.Gold` are initialized properly
3. `ch.InRoom` is properly set when in a room
4. `ch.Act.IsSet()` correctly checks for ACT_BANKER flag
5. `util.OneArgument()` works as expected with empty strings

### Security
No security issues found, but the code assumes proper input validation from util.OneArgument() without checking its return values.

```adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/act/bank.go
  sha256: 7886b150f88ac619
  lines_reviewed: 1-95
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/bank.go
    line: 23
    line_end: 25
    message: >
      Duplicate validation logic between deposit and withdraw sections.
      Both sections have nearly identical validation code that could be extracted into a helper function.
    suggested_fix: >
      Extract amount parsing and validation into a shared function to reduce duplication.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/bank.go
    line: 33
    line_end: 34
    message: >
      Magic number "0" used for validation check instead of named constant.
    suggested_fix: >
      Use a named constant like "MIN_AMOUNT = 0" or similar for better clarity.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/bank.go
    line: 43
    line_end: 44
    message: >
      Duplicate validation logic in both deposit and withdraw sections.
    suggested_fix: >
      Extract common validation into a helper function to avoid repetition.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/act/bank.go
    line: 51
    line_end: 52
    message: >
      Duplicate validation logic in both deposit and withdraw sections.
    suggested_fix: >
      Extract common validation into a helper function to avoid repetition.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/act/bank.go
    line: 60
    line_end: 61
    message: >
      Hardcoded string "You can't do that here." - consider using constants or constants file.
    suggested_fix: >
      Define this as a constant string to allow for easier localization or future changes.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/act/bank.go
    line: 67
    line_end: 68
    message: >
      Hardcoded string "Syntax: bank balance|deposit|withdraw <amount>" - consider using constants.
    suggested_fix: >
      Define this as a constant string to allow for easier localization or future changes.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/act/bank.go
    line: 73
    line_end: 74
    message: >
      Hardcoded string "How much do you want to deposit?" - consider using constants.
    suggested_fix: >
      Define this as a constant string to allow for easier localization or future changes.
  - id: F8
    severity: minor
    category: maintainability
    file: internal/act/bank.go
    line: 79
    line_end: 80
    message: >
      Hardcoded string "You don't have that much gold." - consider using constants.
    suggested_fix: >
      Define this as a constant string to allow for easier localization or future changes.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/act/bank.go
    line: 85
    line_end: 86
    message: >
      Hardcoded string "You don't have that much gold in your account." - consider using constants.
    suggested_fix: >
      Define this as a constant string to allow for easier localization or future changes.
  - id: F10
    severity: minor
    category: maintainability
    file: internal/act/bank.go
    line: 92
    line_end: 93
    message: >
      Hardcoded string "You withdraw %d gold. Your new balance is %d gold." - consider using constants.
    suggested_fix: >
      Define this as a constant string to allow for easier localization or future changes.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/act/bank.go
    line: 94
    line_end: 95
    message: >
      Hardcoded string "Your bank balance is %d gold." - consider using constants.
    suggested_fix: >
      Define this as a constant string to allow for easier localization or future changes.
```
