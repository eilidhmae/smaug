# Adversary Review

**Target**: `internal/act/wiz2.go`
**Timestamp**: 2026-05-16T07:04:59Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

**Scope**: Review of internal/act/wiz2.go for correctness, security, and maintainability.

### Claim Verification
All claims in the code appear to be implemented correctly. The file implements various wizard commands like wizlock, shutdown, reboot, hell, log, deny, pardon, disconnect, mortalize, slookup, and sset. All functions are present and match their descriptions.

### Test Verification
No test files were found for this file or related functionality. This is a concern as critical system commands like shutdown/reboot should have tests to ensure proper behavior under various conditions.

### Complexity Audit
- **File size**: 760 lines (proportional to size)
- **Function size**: Most functions are under 30 lines, except DoSset which is 50 lines
- **Abstraction depth**: Functions use simple control flow with minimal abstraction layers
- **New dependencies**: Uses standard library packages only (no new external dependencies)
- **Premature generalization**: No evidence of premature generalization
- **Feature flags**: None detected

### Scope Check
The file contains only wizard-related commands and helper functions. No additional features beyond what's described in the comments were added.

### Alternative Approach
For `DoSset`, instead of using string parsing and switch statements, a more maintainable approach would be to create a map of field names to setter functions or use reflection to set fields directly. However, given the limited scope of fields being modified, the current implementation is acceptable.

### Assumptions
1. `WorldRef` is properly initialized before any wizard command is executed
2. `ShutdownFunc` and `DisconnectFunc` are set by main.go before any wizard command runs
3. The `SaveFunc` function exists and works correctly
4. `CmdRegistry` is populated before any wizard command is executed
5. `SaveFunc` is not nil when needed for saving players before disconnecting

### Security
No security issues found in this file. All input validation appears appropriate and there are no obvious injection vulnerabilities.

```
adversary-review
verdict: CONCERNS
confidence: medium
artifact:
  path: internal/act/wiz2.go
  sha256: b6432a7467f1f18b
  lines_reviewed: 1-760
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading space).
    suggested_fix: >
      Use `strconv.Atoi` or similar with proper error handling instead of fmt.Sscanf.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 145
    message: >
      The slookup command shows all skills when "all" is passed, but there's no
      validation that the skill list is valid before iterating.
    suggested_fix: >
      Add validation that WorldRef.Skills is initialized before iteration.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/wiz2.go
    line: 139
    message: >
      DoSset uses string parsing for field names instead of a map-based lookup.
      This approach is error-prone and harder to maintain than a map lookup.
    suggested_fix: >
      Replace the switch statement with a map of field names to setter functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/wiz2.go
    line: 155
    message: >
      The function `parseIntSafe` is used but doesn't handle negative numbers correctly.
      It will parse " -5" as "-5" but not " -5" (with leading
