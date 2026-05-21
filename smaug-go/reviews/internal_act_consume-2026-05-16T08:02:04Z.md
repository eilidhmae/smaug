# Adversary Review

**Target**: `internal/act/consume.go`
**Timestamp**: 2026-05-16T08:02:04Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

### Claim Verification
N/A - This is a single file review.

### Test Verification
N/A - No test files mentioned or referenced in the code.

### Complexity Audit
- **File size**: 301 lines total, with 250 lines of actual implementation code (excluding comments and whitespace)
- **Function size**: 
  - `DoEat` (lines 64-95): 32 lines - quite large for a single function
  - `DoDrink` (lines 100-137): 38 lines - also quite large
  - `DoEmpty` (lines 140-155): 16 lines - acceptable size
  - `DoFill` (lines 157-175): 19 lines - acceptable size
- **Abstraction depth**: 
  - The code uses direct access to `WorldRef` global variable which creates tight coupling between modules
  - `GainCondition` function has some complexity in handling conditions and sending messages
- **New dependencies**: None added beyond standard Go libraries and existing internal packages

### Scope Check
- The file contains multiple commands (`eat`, `drink`, `fill`, `empty`) but only implements the first two commands fully
- There's no clear indication that the `DoEmpty` function is complete or tested
- No new features were added beyond what was described in the code itself

### Alternative Approach
The most significant design issue is the use of global variables like `WorldRef` and direct access to character data structures. A better approach would be to pass character data as parameters rather than relying on global state.

### Assumptions
- Assumes `WorldRef` is properly initialized before any usage
- Assumes `lookupSkillSlot` will return valid skill slot numbers or -1 for unknown skills
- Assumes `ch.PCData.Condition` array has sufficient length for all condition types
- Assumes `obj.Value` array has sufficient elements for all item types

### Security
- No injection vulnerabilities detected
- No sensitive information exposed in code comments or comments
- No hardcoded secrets or credentials found

### Verdict Summary
The code has several design issues including tight coupling through global variables, large functions that could benefit from refactoring, and some assumptions about data structure sizes that aren't validated.

```adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/act/consume.go
  sha256: 944d41c221178166
  lines_reviewed: 1-301
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/consume.go
    line: 1
    message: >
      Uses global variable `WorldRef` directly instead of passing it as a parameter.
      This creates tight coupling between modules and makes testing difficult.
    suggested_fix: >
      Pass `WorldRef` as a parameter to functions that need access to it.
  - id: F2
    severity: major
    category: maintainability
    file: internal/act/consume.go
    line: 64
    message: >
      `DoEat` function is very long (32 lines) and does multiple things:
      parsing arguments, checking conditions, handling different item types,
      and sending messages. This violates single responsibility principle.
    suggested_fix: >
      Split into smaller functions for argument parsing, food handling, poison handling, etc.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 100
    message: >
      `DoDrink` function is quite long (38 lines) and does multiple things:
      parsing arguments, finding drinkable objects, handling different object types,
      and sending messages. Could be refactored into smaller functions.
    suggested_fix: >
      Split into smaller functions for argument parsing, object finding, condition handling, etc.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/act/consume.go
    line: 157
    message: >
      `DoEmpty` function is short but incomplete - it only handles one type of container
      and doesn't handle other container types that might be supported.
    suggested_fix: >
      Add support for other container types or add documentation that this is intentional.
```
