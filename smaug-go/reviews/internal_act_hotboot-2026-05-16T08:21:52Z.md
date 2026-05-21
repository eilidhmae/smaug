# Adversary Review

**Target**: `internal/act/hotboot.go`
**Timestamp**: 2026-05-16T08:21:52Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

### Complexity Audit

The file contains 318 lines of code with a single function `DoHotboot` that is 205 lines long. This function performs multiple operations including:
- Gate checking for trust level and combat state
- Descriptor management (dropping mid-nanny descriptors)
- World saving
- Player session handling
- File descriptor manipulation for hotboot recovery
- Exec syscall preparation and execution

The function has multiple nested conditionals and error handling paths that increase complexity. There's also a complex error handling flow with multiple return points and resource cleanup.

### Scope Check

The file implements hotboot functionality but includes some additional features not explicitly mentioned in the design document:
- The gate checks for OLC states beyond just CON_EDITING
- The implementation handles both mid-nanny descriptors and pre-login descriptors differently
- The implementation includes detailed logging and error handling that goes beyond basic requirements

### Alternative Approach

A simpler approach would be to separate the hotboot logic into smaller, more focused functions:
1. Validate hotboot conditions (trust, combat, etc.)
2. Drop descriptors and save world state
3. Handle player sessions and prepare exec arguments
4. Execute the exec syscall

This would reduce the complexity of the main function and make it easier to test individual components.

### Assumptions

The code assumes:
1. The `execSelf` function will never return nil in production (but tests can stub this)
2. All descriptors have valid FDs that can be manipulated via syscall
3. The system has sufficient resources to handle the hotboot process
4. The `HotbootPauseFunc` is properly wired by the boot process
5. File descriptor numbers remain stable across exec

### Security

The code uses `syscall.Exec` which is inherently secure as long as the executable path is properly validated. However, there's potential for command injection if `os.Executable()` returns a path that has been compromised or if the path contains special characters that could be interpreted by shell interpreters.

### Quorum

No concerns found with the current implementation.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/act/hotboot.go
  sha256: d67734ef98489a54
  lines_reviewed: 1-318
findings:
  - id: F1
    severity: major
    category: security
    file: internal/act/hotboot.go
    line: 103
    message: >
      The code assumes that os.Executable() will return a safe path without
      special characters that could lead to shell injection vulnerabilities.
    suggested_fix: >
      Validate and sanitize the executable path before use in syscall.Exec.
  - id: F2
    severity: major
    category: maintainability
    file: internal/act/hotboot.go
    line: 205
    message: >
      The function DoHotboot is 205 lines long and performs multiple distinct
      operations. This makes it difficult to test and maintain.
    suggested_fix: >
      Refactor into smaller, more focused functions for each phase of hotboot
      (validation, descriptor management, world saving, exec preparation).
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/hotboot.go
    line: 103
    message: >
      The code uses fmt.Sprintf to convert integers to strings, which is less
      efficient than direct conversion or using strconv.Itoa.
    suggested_fix: >
      Use strconv.Itoa for integer to string conversion instead of fmt.Sprintf.
```
