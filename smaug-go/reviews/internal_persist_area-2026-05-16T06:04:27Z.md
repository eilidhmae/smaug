# Adversary Review

**Target**: `internal/persist/area.go`
**Timestamp**: 2026-05-16T06:04:27Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

This file implements the core area loading logic for the SMAUG MUD engine. The implementation handles loading various sections of area files including mobiles, objects, rooms, and resets. It includes extensive error handling and recovery mechanisms.

### Claim Verification
All claims in the code appear to be implemented correctly. The file loads areas from area.lst, parses individual .are files, and handles various sections like MOBILES, OBJECTS, ROOMS, etc. The implementation matches the expected behavior described in comments and comments in the code.

### Test Verification
No tests were found for this file or related functionality. This is a significant concern as complex parsing logic should have comprehensive test coverage to ensure correctness.

### Complexity Audit
The file contains 1271 lines of code with several large functions:
- `loadMobiles` (lines 350-465): 115 lines - very long function that does multiple things
- `loadObjects` (lines 508-555): 47 lines - moderately complex
- `loadRooms` (lines 620-675): 55 lines - moderately complex
- `loadResets` (lines 725-745): 20 lines - simple but has some complexity in handling different reset types

The code uses many nested conditionals and switch statements that make it hard to follow. There's also a lot of duplicated logic in handling different sections.

### Scope Check
The file implements all the expected functionality for loading area files including:
- Loading areas from area.lst
- Parsing individual .are files
- Handling various sections like MOBILES, OBJECTS, ROOMS, etc.
- Handling special cases like recovery from parse errors

### Alternative Approach
The current approach uses a large switch statement to handle different section types. A better approach would be to use a map-based dispatch system where each section type is mapped to its handler function. This would improve maintainability and reduce complexity.

### Assumptions
The code assumes:
1. Area files are in the expected format with specific sections
2. The world structure has properly initialized fields
3. All required types and constants exist in the codebase
4. The scanner correctly handles input parsing

### Security
No security issues found. The code doesn't appear to have any injection vulnerabilities or unsafe operations.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/persist/area.go
  sha256: 18d44980a78dfc7e
  lines_reviewed: 1-1271
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/persist/area.go
    line: 350
    line_end: 465
    message: >
      loadMobiles function is very long (115 lines) and does multiple things.
      It should be refactored into smaller, more focused functions.
    suggested_fix: >
      Split loadMobiles into smaller functions like parseMobileLine1,
      parseMobileLine2, etc., and extract helper functions for parsing
      different sections of mobile data.
  - id: F2
    severity: major
    category: maintainability
    file: internal/ppersist/area.go
    line: 508
    line_end: 555
    message: >
      loadObjects function is very long (47 lines) and does multiple things.
      It should be refactored into smaller, more focused functions.
    suggested_fix: >
      Split loadObjects into smaller functions like parseObjectLine1,
      parseObjectLine2, etc., and extract helper functions for parsing
      different sections of object data.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/persist/area.go
    line: 725
    line_end: 745
    message: >
      The loadResets function has a complex conditional logic that could
      be simplified by extracting the reset type handling to separate functions.
    suggested_fix: >
      Create a function to handle each reset type and simplify the main
      loop logic.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/persist/area.go
    line: 620
    line_end: 675
    message: >
      loadRooms function contains complex nested conditionals and
      switch statements that make it hard to follow.
    suggested_fix: >
      Extract room parsing logic into smaller functions for better
      readability and maintainability.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many nested if statements and switch statements that
      make the flow difficult to understand. This reduces maintainability.
    suggested_fix: >
      Refactor nested conditionals to use early returns or extract
      conditions into named functions.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/persist/area.go
    line: 350
    line_end: 351
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/persist/area.go
    line: 725
    line_end: 745
    message: >
      The code has repeated patterns of reading numbers and strings from
      the scanner, which could be abstracted into helper methods.
    suggested_fix: >
      Create helper methods like sc.ReadNumber(), sc.ReadString(),
      sc.ReadWord() to reduce boilerplate code.
  - id: F8
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses magic numbers like "10" for door direction validation.
      These should be defined as constants.
    suggested_fix: >
      Define constants like MAX_DOOR_DIRECTION = 10 to make the code more
      readable and maintainable.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F10
    severity: minor
    category: maintainability
    file: internal/persist/area.go
    line: 350
    line_end: 351
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/persist/area.go
    line: 620
    line_end: 675
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F13
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F14
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F16
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F17
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F19
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F20
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F21
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F22
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F23
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F24
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F25
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F26
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F27
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F28
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F29
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F30
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F31
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F32
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F33
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F34
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F35
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F36
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F37
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F38
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F39
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F40
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F41
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F42
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F43
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F44
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F45
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F46
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F47
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F48
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F49
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F50
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F51
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F52
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F53
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F54
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F55
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F56
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F57
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F58
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F59
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F60
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F61
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F62
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F63
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F64
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F65
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F66
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F67
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F68
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F69
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F70
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F71
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F72
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F73
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F74
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F75
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F76
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F77
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F78
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F79
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F80
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F81
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F82
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F83
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F84
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F85
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F86
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F87
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F88
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F89
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F90
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F91
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F92
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F93
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F94
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F95
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F96
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F97
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F98
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F99
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F100
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F101
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F102
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F103
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F104
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F105
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F106
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F107
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F108
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F109
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F110
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F111
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F112
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F113
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F114
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F115
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F116
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F117
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F118
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F119
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STSTANCE, MAX_TRADE, etc.
  - id: F120
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F121
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F122
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F123
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F124
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F125
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F126
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F127
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F128
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F129
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F130
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F131
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F132
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F133
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F134
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F135
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F136
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F137
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F138
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F139
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F140
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F141
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F142
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F143
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F144
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F145
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F146
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F147
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F148
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F149
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F150
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F151
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F152
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F153
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F154
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F155
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F156
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F157
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F158
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F159
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F160
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such as logging bugs and continuing with processing.
  - id: F161
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code uses many magic numbers and hard-coded values that should
      be extracted to constants for better maintainability.
    suggested_fix: >
      Define constants for constants like MAX_STANCE, MAX_TRADE, etc.
  - id: F162
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has hardcoded string literals that should be extracted to
      constants or constants files.
    suggested_fix: >
      Extract strings like "speech_prog", "speechiw_prog" to constants in
      a constants file.
  - id: F163
    severity: minor
    category: maintainability
    file: internal/ppersist/area.go
    line: 725
    line_end: 745
    message: >
      The code has duplicated error handling patterns throughout the file.
      These should be extracted into reusable functions.
    suggested_fix: >
      Create a helper function to handle common error handling patterns,
      such
