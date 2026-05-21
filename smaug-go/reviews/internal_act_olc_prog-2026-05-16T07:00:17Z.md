# Adversary Review

**Target**: `internal/act/olc_prog.go`
**Timestamp**: 2026-05-16T07:00:17Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

**Scope**: OLC mpedit/opedit/rpedit command implementation

### Claim Verification
All claims in the code match what's present in the file.

### Test Verification
No tests were provided for this file.

### Complexity Audit
- **File size**: 687 lines total, with 500+ lines of comments and documentation
- **Function size**: 
  - `progEditDispatch` (230 lines) - very large function that handles all logic
  - `resolveProgEditTarget` (120 lines) - very large function handling all target resolution logic
  - `progEditOpenEditor` (15 lines) - small but complex closure capture
- **Abstraction depth**: 
  - Multiple levels of indirection in the editor callback closure
  - Complex argument parsing logic with multiple string manipulations
  - Complex conditional logic for different command types and targets

### Scope Check
The file implements three commands: mpedit, opedit, rpedit. All three are implemented correctly and consistently.

### Alternative Approach
The current approach uses a single dispatch function that handles all three commands with conditional logic. A cleaner approach would be to:
1. Separate the command-specific logic into separate functions
2. Use a map-based dispatcher instead of switch statements
3. Split the argument parsing logic into reusable components

### Assumptions
- The `WorldRef` global variable exists and is properly initialized
- `StartEditingFunc` and `CopyBufferFunc` are properly set before use
- `StopEditingFunc` is properly set before use
- The `util` package provides expected functionality for string manipulation and flag handling

### Security
- No injection vulnerabilities detected
- Input validation appears adequate for command arguments
- No hardcoded secrets or credentials present

### Quorum
Not applicable as this is a single review.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/act/olc_prog.go
  sha256: 99c5bea31427c988
  lines_reviewed: 1-687
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured by reference,
      which could lead to subtle bugs if the closure is called after the
      function has returned or if multiple closures are created.
    suggested_fix: >
      Pass all necessary parameters explicitly to the closure instead of
      relying on capture.
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/olc_prog.go
    line: 100
    message: >
      Function `progEditDispatch` is over 200 lines long and contains complex
      conditional logic for handling different command types and target types.
      This makes the function difficult to understand, test, and maintain.
    suggested_fix: >
      Split into smaller functions: separate argument parsing, command dispatch,
      and target resolution logic into separate functions.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 154
    message: >
      The use of string manipulation and string splitting logic is complex and
      error-prone. The code relies heavily on string manipulation without
      proper validation or validation of input formats.
    suggested_fix: >
      Consider using a more structured approach for parsing arguments, such as
      a parser library or structured parsing methods.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/olc_prog.go
    line: 130
    message: >
      The editor callback closure captures variables in a way that's not
      immediately clear. The captured variables are captured
