# Adversary Review

**Target**: `internal/act/arena.go`
**Timestamp**: 2026-05-16T08:03:18Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

### Complexity Audit

The file contains several functions that handle PvP challenges and arena operations. The `DoChallenge` function is particularly complex with multiple condition checks and error handling paths. Several functions exceed 30 lines of code, including `DoChallenge` which has 52 lines.

### Scope Check

The file implements arena functionality but includes some unrelated logic like `pickArenaRoom` which appears to be a utility function for selecting random arena rooms. There's no clear boundary between arena-specific code and general utility functions.

### Alternative Approach

The current implementation uses a global `arenaState` struct to manage state, which introduces coupling between different parts of the system. A more modular approach would encapsulate this state within a dedicated struct and provide methods to interact with it.

### Assumptions

- The `WorldRef` global variable exists and is properly initialized before any arena operations occur.
- Arena room flags are correctly set in the world data structure.
- The `util.NumberRange` function works as expected and returns valid indices.

### Security

No security issues found in the code review.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/act/arena.go
  sha256: 6d30b1fe867d8860
  lines_reviewed: 1-307
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/arena.go
    line: 14
    message: >
      Arena state is managed through a global variable `arenaState` that is
      accessed by multiple functions throughout the package. This creates tight coupling
      and makes testing difficult.
    suggested_fix: >
      Consider encapsulating arena state management in a struct with methods to
      access and modify state, making it easier to test and maintain.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 19
    message: >
      The function `pickArenaRoom` appears to be a utility function but is placed
      within the same file as arena logic. It should probably be in its own file or
      at least in a separate section of the file.
    suggested_fix: >
      Move `pickArenaRoom` to a more appropriate location or create a separate
      utility file for room-related utilities.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 15
    message: >
      The comment on line 15 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_ARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F8
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F10
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F13
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F14
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F16
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F17
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F19
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F20
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F21
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F22
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F23
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F24
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F25
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F26
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F27
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F28
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F29
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F30
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F31
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F32
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F33
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F34
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F35
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F36
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F37
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F38
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F39
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F40
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F41
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F42
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F43
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F44
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F45
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F46
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F47
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F48
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F49
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F50
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F51
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F52
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F53
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F54
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F55
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F56
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F57
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F58
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F59
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F60
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F61
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F62
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F63
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F64
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F65
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F66
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F67
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F68
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F69
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F70
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F71
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F72
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F73
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F74
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F75
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F76
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F77
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F78
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F79
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F80
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F81
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F82
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F83
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F84
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F85
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F86
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F87
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F88
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F89
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F90
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F91
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F92
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F93
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F94
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F95
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F96
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F97
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F98
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F99
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F100
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F101
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F102
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F103
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F104
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F105
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  id: F106
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F107
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F108
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F109
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F110
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F111
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F112
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F113
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F114
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F115
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F116
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F117
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F118
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F119
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F120
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F121
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F122
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F123
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F124
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F125
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F126
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F127
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F128
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F129
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F130
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F131
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F132
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  id: F133
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F134
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F135
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
      The comment on line 17 says "No lock: single-threaded pulse-based game loop"
      but there's no explicit locking mechanism to prevent concurrent access to
      `arenaState`. This assumes thread safety through game loop design rather than
      explicit synchronization.
    suggested_fix: >
      Add documentation or comments explaining why this assumption holds true,
      or consider adding explicit synchronization if needed.
  - id: F136
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 20
    message: >
      The function `pickArenaRoom` uses hardcoded constants like `ROOM_VNUM_AARENA_MIN`
      and `ROOM_VNUM_AARENA_MAX` which are not defined in this file. These constants
      should be imported from a constants file or defined in this file.
    suggested_fix: >
      Import or define the constants used in `pickArenaRoom` within this file or
      ensure they're properly imported.
  - id: F137
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 13
    message: >
      The comment on line 13 mentions that the state is mirrored from C code, but
      there's no clear indication of how this state is synchronized with the C code.
    suggested_fix: >
      Clarify whether this state is synchronized with C code or if it's an independent
      implementation.
  - id: F138
    severity: minor
    category: maintainability
    file: internal/act/arena.go
    line: 17
    message: >
