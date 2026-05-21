# Adversary Review

**Target**: `internal/act/comm.go`
**Timestamp**: 2026-05-16T08:01:44Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

### Claim Verification
N/A - Skipping steps 0-2 as requested.

### Test Verification
N/A - Skipping step 2 as requested.

### Complexity Audit
The file contains several functions that handle communication commands like `DoTell`, `DoYell`, `DoShout`, etc. The complexity is moderate with some repeated logic patterns but no overly complex structures or deep nesting.

### Scope Check
The file contains only communication-related functions (`DoTell`, `DoYell`, `DoShout`, `DoGossip`, `DoPmote`, `DoSpeak`, `DoLearn`, `DoEmote`) and no additional features beyond what's expected for communication handling in a MUD system.

### Alternative Approach
For the `translateFor` function, instead of checking each condition separately, we could use a map lookup to determine translation behavior based on language flags and player status. This would make the code more maintainable and easier to extend.

### Assumptions
1. The `WorldRef` global variable exists and is properly initialized elsewhere.
2. Character data structures have valid fields like `Speaking`, `Speaks`, `Act`, `InRoom`, `Name`, `Desc`, `Reply`.
3. The `util` package provides `OneArgument`, `LanguageName`, `LanguageBit`, and `Scramble` functions.
4. `PLR_SILENCE` and other flags are defined in `types`.
5. Room flags like `ROOM_SILENCE` exist and work as expected.
6. `mudprog.TrigTell` and similar functions exist and work correctly.

### Security
No security issues found in the code review.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/act/comm.go
  sha256: 1decddc55c2ae530
  lines_reviewed: 1-300
findings:
  - id: F1
    severity: major
    category: correctness
    file: internal/act/comm.go
    line: 180
    message: >
      In DoTell, there's a comment about a latent bug in the Go port where
      PLR_SILENCE is not properly checked. This could lead to silent failures
      or security issues if silence isn't enforced correctly.
    suggested_fix: >
      Ensure PLR_SILENCE is properly enforced by checking it before sending
      messages, not just in the gate logic.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/comm.go
    line: 247
    message: >
      The DoPmote function has inconsistent handling of the "your" replacement
      logic. It should replace the character name with "your" when showing
      the emote to other players.
    suggested_fix: >
      Implement proper name replacement logic similar to how DoEmote works,
      replacing the character's name with "your" for non-characters.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/comm.go
    line: 265
    message: >
      The DoLearn function has a hardcoded check for PLR_HOLYLIGHT that may
      not be appropriate for all systems or could be made more configurable.
    suggested_fix: >
      Consider making the trust requirement configurable or using a more
      flexible permission system instead of hardcoding the holylight flag.
```
