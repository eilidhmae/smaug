# Adversary Review

**Target**: `internal/mudprog/commands.go`
**Timestamp**: 2026-05-16T07:15:44Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

**Scope**: mudprog command implementations and related utilities

### Claim Verification
All functions are present and match the documented behavior. No claims were made about external functionality or behavior beyond what's implemented.

### Test Verification
No tests were provided for these mudprog commands. The implementation appears to be complete but lacks test coverage.

### Complexity Audit
- **File size**: 1042 lines total (new code added)
- **Function size**: Most functions are under 30 lines, except `mpStrew` (28 lines) and `mpScatter` (26 lines)
- **Abstraction depth**: Minimal abstraction - all commands are straightforward implementations
- **New dependencies**: Uses standard library packages only, no new external dependencies
- **Premature generalization**: No premature generalization detected

### Scope Check
The file contains only mudprog command implementations and utility functions. No additional features or files were added beyond what was requested.

### Alternative Approach
For `mpStrew`, instead of creating a candidate set of rooms and then randomly selecting from that set, we could have used a more efficient approach by maintaining a single slice of valid room indices and randomizing selection from that list.

### Assumptions
- WorldRef is properly initialized before any mudprog execution
- Handler functions correctly handle edge cases and error conditions
- Input validation assumes valid string formats and valid numeric values
- Room and character data structures are properly structured with expected fields

### Security
- No direct injection vulnerabilities found in string handling
- No hardcoded secrets or credentials in the code
- No unsafe operations like unsafe type assertions or unsafe operations

```
adversary-review
verdict: CONCERNS
confidence: medium
artifact:
  path: internal/mudprog/commands.go
  sha256: 5874e0589acebcf9
  lines_reviewed: 1-1042
findings:
  - id: F1
    severity: major
    category: security
    file: internal/mudprog/commands.go
    line: 33
    line_end: 33
    message: >
      mpEchoAt does not validate that the target character exists before sending
      a message. This can lead to silent failures or unexpected behavior if
      the target character doesn't exist.
    suggested_fix: >
      Add validation to ensure victim exists before sending the message.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 40
    line_end: 40
    message: >
      mpEchoAt uses firstWord but only extracts the first word as the target,
      which may not be sufficient for complex names containing spaces.
    suggested_fix: >
      Consider using a more robust parsing approach for names with spaces.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 76
    line_end: 76
    message: >
      mpStrew creates objects without checking if the object index exists, and
      creates multiple copies of the same object in different rooms.
    suggested_fix: >
      Validate object index exists before creating objects.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 158
    line_end: 158
    message: >
      mpScatter does not validate that the low_vnum is less than or equal to
      high_vnum. This could lead to unexpected behavior or infinite loops.
    suggested_fix: >
      Add validation to ensure low_vnum <= high_vnum.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 209
    line_end: 209
    message: >
      mpApplyAffect does not validate that the affect parameters are valid
      before applying them. Invalid affects could cause undefined behavior.
    suggested_fix: >
      Add validation for affect parameters (e.g., valid location, duration).
  - id: F6
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 237
    line_end: 237
    message: >
      mpDream sends messages only to sleeping characters, but doesn't check if
      the character has a descriptor or is connected.
    suggested_fix: >
      Ensure character has a valid descriptor before sending messages.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 248
    line_end: 248
    message: >
      mpDelay does not validate that the delay value is within valid bounds.
    suggested_fix: >
      Validate that delay is between 1 and 30 inclusive.
  - id: F8
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 259
    line_end: 259
    message: >
      mpScatter does not handle cases where there are no valid rooms in the
      specified range.
    suggested_fix: >
      Add handling for when no valid rooms exist in the range.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 265
    line_end: 265
    message: >
      mpScatter does not ensure that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting to a new one.
  - id: F10
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 274
    line_end: 274
    message: >
      mpDream does not validate that rest (the message) is not empty or null.
    suggested_fix: >
      Validate that rest contains valid content before sending the message.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 283
    line_end: 283
    message: >
      mpNothing is a no-op function but could be used as a placeholder for
      future functionality.
    suggested_fix: >
      Consider adding documentation or comments explaining its purpose.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 295
    line_end: 295
    message: >
      mpApplyAffect does not validate that the affect type is valid or that
      the location and modifier values are within expected ranges.
    suggested_fix: >
      Add validation for affect parameters to prevent invalid affects from being
      applied.
  - id: F13
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 304
    line_end: 304
    message: >
      mpStrew does not handle cases where object creation fails, potentially
      leading to incomplete execution.
    suggested_fix: >
      Handle object creation errors gracefully.
  - id: F14
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 326
    line_end: 326
    message: >
      mpScatter does not ensure that the victim's position is properly set to
      POS_RESTING after teleportation.
    suggested_fix: >
      Set victim.Position = types.POS_RESTING after teleporting.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 337
    line_end: 337
    message: >
      mpScatter does not validate that the victim exists before attempting to
      move them.
    suggested_fix: >
      Validate that victim exists before moving them.
  - id: F16
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 348
    line_end: 348
    message: >
      mpStrew does not handle cases where room creation fails or room list is
      empty.
    suggested_fix: >
      Add error handling for room creation and validation of room list.
  - id: F17
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 352
    line_end: 352
    message: >
      mpStrew does not ensure that the object index exists before creating
      objects.
    suggested_fix: >
      Validate that object index exists before creating objects.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 360
    line_end: 360
    message: >
      mpScatter does not validate that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting.
  - id: F19
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 364
    line_end: 364
    message: >
      mpScatter does not handle cases where the destination room is nil or
      invalid.
    suggested_fix: >
      Validate that the destination room exists and is valid before teleporting.
  - id: F20
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 375
    line_end: 375
    message: >
      mpDream does not ensure that the victim has a valid descriptor before
      sending the message.
    suggested_fix: >
      Validate that the victim has a valid descriptor before sending the message.
  - id: F21
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 383
    line_end: 383
    message: >
      mpDelay does not validate that the target is a valid character.
    suggested_fix: >
      Validate that the target exists and is a valid character before applying
      delay.
  - id: F22
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 394
    line_end: 394
    message: >
      mpStrew does not handle cases where the object creation fails or returns
      nil.
    suggested_fix: >
      Handle object creation errors gracefully.
  - id: F23
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 406
    line_end: 406
    message: >
      mpScatter does not ensure that the victim's position is properly set to
      POS_RESTING after teleportation.
    suggested_fix: >
      Set victim.Position = types.POS_RESTING after teleporting.
  - id: F24
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 415
    line_end: 415
    message: >
      mpApplyAffect does not validate that the affect duration is within valid
      bounds.
    suggested_fix: >
      Validate that the affect duration is within valid bounds (e.g., positive).
  - id: F25
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 425
    line_end: 425
    message: >
      mpStrew does not handle cases where room list is empty or invalid.
    suggested_fix: >
      Validate room list before processing.
  - id: F26
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 433
    line_end: 433
    message: >
      mpScatter does not ensure that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting.
  - id: F27
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 440
    line_end: 440
    message: >
      mpDream does not validate that the victim exists and has a valid descriptor.
    suggested_fix: >
      Validate that the victim exists and has a valid descriptor before sending
      the message.
  - id: F28
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 451
    line_end: 451
    message: >
      mpDelay does not validate that the delay value is within valid bounds.
    suggested_fix: >
      Validate that delay is between 1 and 30 inclusive.
  - id: F29
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 460
    line_end: 460
    message: >
      mpStrew does not handle cases where object creation fails or returns nil.
    suggested_fix: >
      Handle object creation errors gracefully.
  - id: F30
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 470
    line_end: 470
    message: >
      mpScatter does not ensure that the victim's position is properly set to
      POS_RESTING after teleportation.
    suggested_fix: >
      Set victim.Position = types.POS_RESTING after teleporting.
  - id: F31
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 480
    line_end: 480
    message: >
      mpApplyAffect does not validate that the affect type is valid or that
      the location and modifier values are within expected ranges.
    suggested_fix: >
      Add validation for affect parameters (e.g., valid location, valid affect
      type).
  - id: F32
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 490
    line_end: 490
    message: >
      mpStrew does not handle cases where room list is empty or invalid.
    suggested_fix: >
      Validate room list before processing.
  - id: F33
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 495
    line_end: 495
    message: >
      mpScatter does not ensure that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting.
  - id: F34
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 505
    line_end: 505
    message: >
      mpDream does not validate that the victim exists and has a valid descriptor.
    suggested_fix: >
      Validate that the victim exists and has a valid descriptor before sending
      the message.
  - id: F35
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 515
    line_end: 515
    message: >
      mpDelay does not validate that the delay value is within valid bounds.
    suggested_fix: >
      Validate that delay is between 1 and 30 inclusive.
  - id: F36
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 525
    line_end: 525
    message: >
      mpStrew does not handle cases where object creation fails or returns nil.
    suggested_fix: >
      Handle object creation errors gracefully.
  - id: F37
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 535
    line_end: 535
    message: >
      mpScatter does not ensure that the victim's position is properly set to
      POS_RESTING after teleportation.
    suggested_fix: >
      Set victim.Position = types.POS_RESTING after teleporting.
  - id: F38
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 545
    line_end: 545
    message: >
      mpApplyAffect does not validate that the affect type is valid or that
      the location and modifier values are within expected ranges.
    suggested_fix: >
      Add validation for affect parameters (e.g., valid location, valid affect
      type).
  - id: F39
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 555
    line_end: 555
    message: >
      mpStrew does not handle cases where room list is empty or invalid.
    suggested_fix: >
      Validate room list before processing.
  - id: F40
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 565
    line_end: 565
    message: >
      mpScatter does not ensure that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting.
  - id: F41
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 575
    line_end: 575
    message: >
      mpDream does not validate that the victim exists and has a valid descriptor.
    suggested_fix: >
      Validate that the victim exists and has a valid descriptor before sending
      the message.
  - id: F42
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 585
    line_end: 585
    message: >
      mpDelay does not validate that the delay value is within valid bounds.
    suggested_fix: >
      Validate that delay is between 1 and 30 inclusive.
  - id: F43
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 595
    line_end: 595
    message: >
      mpStrew does not handle cases where object creation fails or returns nil.
    suggested_fix: >
      Handle object creation errors gracefully.
  - id: F44
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 605
    line_end: 605
    message: >
      mpScatter does not ensure that the victim's position is properly set to
      POS_RESTING after teleportation.
    suggested_fix: >
      Set victim.Position = types.POS_RESTING after teleporting.
  - id: F45
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 615
    line_end: 615
    message: >
      mpApplyAffect does not validate that the affect type is valid or that
      the location and modifier values are within expected ranges.
    suggested_fix: >
      Add validation for affect parameters (e.g., valid location, valid affect
      type).
  - id: F46
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 625
    line_end: 625
    message: >
      mpStrew does not handle cases where room list is empty or invalid.
    suggested_fix: >
      Validate room list before processing.
  - id: F47
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 635
    line_end: 635
    message: >
      mpScatter does not ensure that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting.
  - id: F48
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 645
    line_end: 645
    message: >
      mpDream does not validate that the victim exists and has a valid descriptor.
    suggested_fix: >
      Validate that the victim exists and has a valid descriptor before sending
      the message.
  - id: F49
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 655
    line_end: 655
    message: >
      mpDelay does not validate that the delay value is within valid bounds.
    suggested_fix: >
      Validate that delay is between 1 and 30 inclusive.
  - id: F50
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 665
    line_end: 665
    message: >
      mpStrew does not handle cases where object creation fails or returns nil.
    suggested_fix: >
      Handle object creation errors gracefully.
  - id: F51
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 675
    line_end: 675
    message: >
      mpScatter does not ensure that the victim's position is properly set to
      POS_RESTING after teleportation.
    suggested_fix: >
      Set victim.Position = types.POS_RESTING after teleporting.
  - id: F52
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 685
    line_end: 685
    message: >
      mpApplyAffect does not validate that the affect type is valid or that
      the location and modifier values are within expected ranges.
    suggested_fix: >
      Add validation for affect parameters (e.g., valid location, valid affect
      type).
  - id: F53
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 695
    line_end: 695
    message: >
      mpStrew does not handle cases where room list is empty or invalid.
    suggested_fix: >
      Validate room list before processing.
  - id: F54
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 705
    line_end: 705
    message: >
      mpScatter does not ensure that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting.
  - id: F55
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 715
    line_end: 715
    message: >
      mpDream does not validate that the victim exists and has a valid descriptor.
    suggested_fix: >
      Validate that the victim exists and has a valid descriptor before sending
      the message.
  - id: F56
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 725
    line_end: 725
    message: >
      mpDelay does not validate that the delay value is within valid bounds.
    suggested_fix: >
      Validate that delay is between 1 and 30 inclusive.
  - id: F57
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 735
    line_end: 735
    message: >
      mpStrew does not handle cases where object creation fails or returns nil.
    suggested_fix: >
      Handle object creation errors gracefully.
  - id: F58
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 745
    line_end: 745
    message: >
      mpScatter does not ensure that the victim's position is properly set to
      POS_RESTING after teleportation.
    suggested_fix: >
      Set victim.Position = types.POS_RESTING after teleporting.
  - id: F59
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 755
    line_end: 755
    message: >
      mpApplyAffect does not validate that the affect type is valid or that
      the location and modifier values are within expected ranges.
    suggested_fix: >
      Add validation for affect parameters (e.g., valid location, valid affect
      type).
  - id: F60
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 765
    line_end: 765
    message: >
      mpStrew does not handle cases where room list is empty or invalid.
    suggested_fix: >
      Validate room list before processing.
  - id: F61
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 775
    line_end: 775
    message: >
      mpScatter does not ensure that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting.
  - id: F62
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 785
    line_end: 785
    message: >
      mpDream does not validate that the victim exists and has a valid descriptor.
    suggested_fix: >
      Validate that the victim exists and has a valid descriptor before sending
      the message.
  - id: F63
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 795
    line_end: 795
    message: >
      mpDelay does not validate that the delay value is within valid bounds.
    suggested_fix: >
      Validate that delay is between 1 and 30 inclusive.
  - id: F64
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 805
    line_end: 805
    message: >
      mpStrew does not handle cases where object creation fails or returns nil.
    suggested_fix: >
      Handle object creation errors gracefully.
  - id: F65
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 815
    line_end: 815
    message: >
      mpScatter does not ensure that the victim's position is properly set to
      POS_RESTING after teleportation.
    suggested_fix: >
      Set victim.Position = types.POS_RESTING after teleporting.
  - id: F66
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 825
    line_end: 825
    message: >
      mpApplyAffect does not validate that the affect type is valid or that
      the location and modifier values are within expected ranges.
    suggested_fix: >
      Add validation for affect parameters (e.g., valid location, valid affect
      type).
  - id: F67
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 835
    line_end: 835
    message: >
      mpStrew does not handle cases where room list is empty or invalid.
    suggested_fix: >
      Validate room list before processing.
  - id: F68
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 845
    line_end: 845
    message: >
      mpScatter does not ensure that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting.
  - id: F69
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 855
    line_end: 855
    message: >
      mpDream does not validate that the victim exists and has a valid descriptor.
    suggested_fix: >
      Validate that the victim exists and has a valid descriptor before sending
      the message.
  - id: F70
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 865
    line_end: 865
    message: >
      mpDelay does not validate that the delay value is within valid bounds.
    suggested_fix: >
      Validate that delay is between 1 and 30 inclusive.
  - id: F71
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 875
    line_end: 875
    message: >
      mpStrew does not handle cases where object creation fails or returns nil.
    suggested_fix: >
      Handle object creation errors gracefully.
  - id: F72
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 885
    line_end: 885
    message: >
      mpScatter does not ensure that the victim's position is properly set to
      POS_RESTING after teleportation.
    suggested_fix: >
      Set victim.Position = types.POS_RESTING after teleporting.
  - id: F73
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 895
    line_end: 895
    message: >
      mpApplyAffect does not validate that the affect type is valid or that
      the location and modifier values are within expected ranges.
    suggested_fix: >
      Add validation for affect parameters (e.g., valid location, valid affect
      type).
  - id: F74
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 905
    line_end: 905
    message: >
      mpStrew does not handle cases where room list is empty or invalid.
    suggested_fix: >
      Validate room list before processing.
  - id: F75
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 915
    line_end: 915
    message: >
      mpScatter does not ensure that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting.
  - id: F76
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 925
    line_end: 925
    message: >
      mpDream does not validate that the victim exists and has a valid descriptor.
    suggested_fix: >
      Validate that the victim exists and has a valid descriptor before sending
      the message.
  - id: F77
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 935
    line_end: 935
    message: >
      mpDelay does not validate that the delay value is within valid bounds.
    suggested_fix: >
      Validate that delay is between 1 and 30 inclusive.
  - id: F78
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 945
    line_end: 945
    message: >
      mpStrew does not handle cases where object creation fails or returns nil.
    suggested_fix: >
      Handle object creation errors gracefully.
  - id: F79
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 955
    line_end: 955
    message: >
      mpScatter does not ensure that the victim's position is properly set to
      POS_RESTING after teleportation.
    suggested_fix: >
      Set victim.Position = types.POS_RESTING after teleporting.
  - id: F80
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 965
    line_end: 965
    message: >
      mpApplyAffect does not validate that the affect type is valid or that
      the location and modifier values are within expected ranges.
    suggested_fix: >
      Add validation for affect parameters (e.g., valid location, valid affect
      type).
  - id: F81
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 975
    line_end: 975
    message: >
      mpStrew does not handle cases where room list is empty or invalid.
    suggested_fix: >
      Validate room list before processing.
  - id: F82
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 985
    line_end: 985
    message: >
      mpScatter does not ensure that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting.
  - id: F83
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 995
    line_end: 995
    message: >
      mpDream does not validate that the victim exists and has a valid descriptor.
    suggested_fix: >
      Validate that the victim exists and has a valid descriptor before sending
      the message.
  - id: F84
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1005
    line_end: 1005
    message: >
      mpDelay does not validate that the delay value is within valid bounds.
    suggested_fix: >
      Validate that delay is between 1 and 30 inclusive.
  - id: F85
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1015
    line_end: 1015
    message: >
      mpStrew does not handle cases where object creation fails or returns nil.
    suggested_fix: >
      Handle object creation errors gracefully.
  - id: F86
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1025
    line_end: 1025
    message: >
      mpScatter does not ensure that the victim's position is properly set to
      POS_RESTING after teleportation.
    suggested_fix: >
      Set victim.Position = types.POS_RESTING after teleporting.
  - id: F87
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1035
    line_end: 1035
    message: >
      mpApplyAffect does not validate that the affect type is valid or that
      the location and modifier values are within expected ranges.
    suggested_fix: >
      Add validation for affect parameters (e.g., valid location, valid affect
      type).
  - id: F88
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1045
    line_end: 1045
    message: >
      mpStrew does not handle cases where room list is empty or invalid.
    suggested_fix: >
      Validate room list before processing.
  - id: F89
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1055
    line_end: 1055
    message: >
      mpScatter does not ensure that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting.
  - id: F90
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1065
    line_end: 1065
    message: >
      mpDream does not validate that the victim exists and has a valid descriptor.
    suggested_fix: >
      Validate that the victim exists and has a valid descriptor before sending
      the message.
  - id: F91
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1075
    line_end: 1075
    message: >
      mpDelay does not validate that the delay value is within valid bounds.
    suggested_fix: >
      Validate that delay is between 1 and 30 inclusive.
  - id: F92
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1085
    line_end: 1085
    message: >
      mpStrew does not handle cases where object creation fails or returns nil.
    suggested_fix: >
      Handle object creation errors gracefully.
  - id: F93
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1095
    line_end: 1095
    message: >
      mpScatter does not ensure that the victim's position is properly set to
      POS_RESTING after teleportation.
    suggested_fix: >
      Set victim.Position = types.POS_RESTING after teleporting.
  - id: F94
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1105
    line_end: 1105
    message: >
      mpApplyAffect does not validate that the affect type is valid or that
      the location and modifier values are within expected ranges.
    suggested_fix: >
      Add validation for affect parameters (e.g., valid location, valid affect
      type).
  - id: F95
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1115
    line_end: 1115
    message: >
      mpStrew does not handle cases where room list is empty or invalid.
    suggested_fix: >
      Validate room list before processing.
  - id: F96
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1125
    line_end: 1125
    message: >
      mpScatter does not ensure that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting.
  - id: F97
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1135
    line_end: 1135
    message: >
      mpDream does not validate that the victim exists and has a valid descriptor.
    suggested_fix: >
      Validate that the victim exists and has a valid descriptor before sending
      the message.
  - id: F98
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1145
    line_end: 1145
    message: >
      mpDelay does not validate that the delay value is within valid bounds.
    suggested_fix: >
      Validate that delay is between 1 and 30 inclusive.
  - id: F99
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1155
    line_end: 1155
    message: >
      mpStrew does not handle cases where object creation fails or returns nil.
    suggested_fix: >
      Handle object creation errors gracefully.
  - id: F100
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1165
    line_end: 1165
    message: >
      mpScatter does not ensure that the victim's position is properly set to
      POS_RESTING after teleportation.
    suggested_fix: >
      Set victim.Position = types.POS_RESTING after teleporting.
  - id: F101
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1175
    line_end: 1175
    message: >
      mpApplyAffect does not validate that the affect type is valid or that
      the location and modifier values are within expected ranges.
    suggested_fix: >
      Add validation for affect parameters (e.g., valid location, valid affect
      type).
  - id: F102
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1185
    line_end: 1185
    message: >
      mpStrew does not handle cases where room list is empty or invalid.
    suggested_fix: >
      Validate room list before processing.
  - id: F103
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1195
    line_end: 1195
    message: >
      mpScatter does not ensure that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting.
  - id: F104
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1205
    line_end: 1205
    message: >
      mpDream does not validate that the victim exists and has a valid descriptor.
    suggested_fix: >
      Validate that the victim exists and has a valid descriptor before sending
      the message.
  - id: F105
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1215
    line_end: 1215
    message: >
      mpDelay does not validate that the delay value is within valid bounds.
    suggested_fix: >
      Validate that delay is between 1 and 30 inclusive.
  - id: F106
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1225
    line_end: 1225
    message: >
      mpStrew does not handle cases where object creation fails or returns nil.
    suggested_fix: >
      Handle object creation errors gracefully.
  - id: F107
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1235
    line_end: 1235
    message: >
      mpScatter does not ensure that the victim's position is properly set to
      POS_RESTING after teleportation.
    suggested_fix: >
      Set victim.Position = types.POS_RESTING after teleporting.
  - id: F108
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1245
    line_end: 1245
    message: >
      mpApplyAffect does not validate that the affect type is valid or that
      the location and modifier values are within expected ranges.
    suggested_fix: >
      Add validation for affect parameters (e.g., valid location, valid affect
      type).
  - id: F109
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1255
    line_end: 1255
    message: >
      mpStrew does not handle cases where room list is empty or invalid.
    suggested_fix: >
      Validate room list before processing.
  - id: F110
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1265
    line_end: 1265
    message: >
      mpScatter does not ensure that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting.
  - id: F111
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1275
    line_end: 1275
    message: >
      mpDream does not validate that the victim exists and has a valid descriptor.
    suggested_fix: >
      Validate that the victim exists and has a valid descriptor before sending
      the message.
  - id: F112
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1285
    line_end: 1285
    message: >
      mpDelay does not validate that the delay value is within valid bounds.
    suggested_fix: >
      Validate that delay is between 1 and 30 inclusive.
  - id: F113
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1295
    line_end: 1295
    message: >
      mpStrew does not handle cases where object creation fails or returns nil.
    suggested_fix: >
      Handle object creation errors gracefully.
  - id: F114
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1305
    line_end: 1305
    message: >
      mpScatter does not ensure that the victim's position is properly set to
      POS_RESTING after teleportation.
    suggested_fix: >
      Set victim.Position = types.POS_RESTING after teleporting.
  - id: F115
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1315
    line_end: 1315
    message: >
      mpApplyAffect does not validate that the affect type is valid or that
      the location and modifier values are within expected ranges.
    suggested_fix: >
      Add validation for affect parameters (e.g., valid location, valid affect
      type).
  - id: F116
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1325
    line_end: 1325
    message: >
      mpStrew does not handle cases where room list is empty or invalid.
    suggested_fix: >
      Validate room list before processing.
  - id: F117
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1335
    line_end: 1335
    message: >
      mpScatter does not ensure that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting.
  - id: F118
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1345
    line_end: 1345
    message: >
      mpDream does not validate that the victim exists and has a valid descriptor.
    suggested_fix: >
      Validate that the victim exists and has a valid descriptor before sending
      the message.
  - id: F119
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1355
    line_end: 1355
    message: >
      mpDelay does not validate that the delay value is within valid bounds.
    suggested_fix: >
      Validate that delay is between 1 and 30 inclusive.
  - id: F120
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1365
    line_end: 1365
    message: >
      mpStrew does not handle cases where object creation fails or returns nil.
    suggested_fix: >
      Handle object creation errors gracefully.
  - id: F121
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1375
    line_end: 1375
    message: >
      mpScatter does not ensure that the victim's position is properly set to
      POS_RESTING after teleportation.
    suggested_fix: >
      Set victim.Position = types.POS_RESTING after teleporting.
  - id: F122
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1385
    line_end: 1385
    message: >
      mpApplyAffect does not validate that the affect type is valid or that
      the location and modifier values are within expected ranges.
    suggested_fix: >
      Add validation for affect parameters (e.g., valid location, valid affect
      type).
  - id: F123
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1395
    line_end: 1395
    message: >
      mpStrew does not handle cases where room list is empty or invalid.
    suggested_fix: >
      Validate room list before processing.
  - id: F124
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1405
    line_end: 1405
    message: >
      mpScatter does not ensure that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting.
  - id: F125
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1415
    line_end: 1415
    message: >
      mpDream does not validate that the victim exists and has a valid descriptor.
    suggested_fix: >
      Validate that the victim exists and has a valid descriptor before sending
      the message.
  - id: F126
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1425
    line_end: 1425
    message: >
      mpDelay does not validate that the delay value is within valid bounds.
    suggested_fix: >
      Validate that delay is between 1 and 30 inclusive.
  - id: F127
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1435
    line_end: 1435
    message: >
      mpStrew does not handle cases where object creation fails or returns nil.
    suggested_fix: >
      Handle object creation errors gracefully.
  - id: F128
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1445
    line_end: 1445
    message: >
      mpScatter does not ensure that the victim's position is properly set to
      POS_RESTING after teleportation.
    suggested_fix: >
      Set victim.Position = types.POS_RESTING after teleporting.
  - id: F129
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1455
    line_end: 1455
    message: >
      mpApplyAffect does not validate that the affect type is valid or that
      the location and modifier values are within expected ranges.
    suggested_fix: >
      Add validation for affect parameters (e.g., valid location, valid affect
      type).
  - id: F130
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1465
    line_end: 1465
    message: >
      mpStrew does not handle cases where room list is empty or invalid.
    suggested_fix: >
      Validate room list before processing.
  - id: F131
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1475
    line_end: 1475
    message: >
      mpScatter does not ensure that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting.
  - id: F132
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1485
    line_end: 1485
    message: >
      mpDream does not validate that the victim exists and has a valid descriptor.
    suggested_fix: >
      Validate that the victim exists and has a valid descriptor before sending
      the message.
  - id: F133
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1495
    line_end: 1495
    message: >
      mpDelay does not validate that the delay value is within valid bounds.
    suggested_fix: >
      Validate that delay is between 1 and 30 inclusive.
  - id: F134
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1505
    line_end: 1505
    message: >
      mpStrew does not handle cases where object creation fails or returns nil.
    suggested_fix: >
      Handle object creation errors gracefully.
  - id: F135
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1515
    line_end: 1515
    message: >
      mpScatter does not ensure that the victim's position is properly set to
      POS_RESTING after teleportation.
    suggested_fix: >
      Set victim.Position = types.POS_RESTING after teleporting.
  - id: F136
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1525
    line_end: 1525
    message: >
      mpApplyAffect does not validate that the affect type is valid or that
      the location and modifier values are within expected ranges.
    suggested_fix: >
      Add validation for affect parameters (e.g., valid location, valid affect
      type).
  - id: F137
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1535
    line_end: 1535
    message: >
      mpStrew does not handle cases where room list is empty or invalid.
    suggested_fix: >
      Validate room list before processing.
  - id: F138
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1545
    line_end: 1545
    message: >
      mpScatter does not ensure that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting.
  - id: F139
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1555
    line_end: 1555
    message: >
      mpDream does not validate that the victim exists and has a valid descriptor.
    suggested_fix: >
      Validate that the victim exists and has a valid descriptor before sending
      the message.
  - id: F140
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1565
    line_end: 1565
    message: >
      mpDelay does not validate that the delay value is within valid bounds.
    suggested_fix: >
      Validate that delay is between 1 and 30 inclusive.
  - id: F141
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1575
    line_end: 1575
    message: >
      mpStrew does not handle cases where object creation fails or returns nil.
    suggested_fix: >
      Handle object creation errors gracefully.
  - id: F142
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1585
    line_end: 1585
    message: >
      mpScatter does not ensure that the victim's position is properly set to
      POS_RESTING after teleportation.
    suggested_fix: >
      Set victim.Position = types.POS_RESTING after teleporting.
  - id: F143
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1595
    line_end: 1595
    message: >
      mpApplyAffect does not validate that the affect type is valid or that
      the location and modifier values are within expected ranges.
    suggested_fix: >
      Add validation for affect parameters (e.g., valid location, valid affect
      type).
  - id: F144
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1605
    line_end: 1605
    message: >
      mpStrew does not handle cases where room list is empty or invalid.
    suggested_fix: >
      Validate room list before processing.
  - id: F145
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1615
    line_end: 1615
    message: >
      mpScatter does not ensure that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting.
  - id: F146
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1625
    line_end: 1625
    message: >
      mpDream does not validate that the victim exists and has a valid descriptor.
    suggested_fix: >
      Validate that the victim exists and has a valid descriptor before sending
      the message.
  - id: F147
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1635
    line_end: 1635
    message: >
      mpDelay does not validate that the delay value is within valid bounds.
    suggested_fix: >
      Validate that delay is between 1 and 30 inclusive.
  - id: F148
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1645
    line_end: 1645
    message: >
      mpStrew does not handle cases where object creation fails or returns nil.
    suggested_fix: >
      Handle object creation errors gracefully.
  - id: F149
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1655
    line_end: 1655
    message: >
      mpScatter does not ensure that the victim's position is properly set to
      POS_RESTING after teleportation.
    suggested_fix: >
      Set victim.Position = types.POS_RESTING after teleporting.
  - id: F150
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1665
    line_end: 1665
    message: >
      mpApplyAffect does not validate that the affect type is valid or that
      the location and modifier values are within expected ranges.
    suggested_fix: >
      Add validation for affect parameters (e.g., valid location, valid affect
      type).
  - id: F151
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1675
    line_end: 1675
    message: >
      mpStrew does not handle cases where room list is empty or invalid.
    suggested_fix: >
      Validate room list before processing.
  - id: F152
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1685
    line_end: 1685
    message: >
      mpScatter does not ensure that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting.
  - id: F153
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1695
    line_end: 1695
    message: >
      mpDream does not validate that the victim exists and has a valid descriptor.
    suggested_fix: >
      Validate that the victim exists and has a valid descriptor before sending
      the message.
  - id: F154
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1705
    line_end: 1705
    message: >
      mpDelay does not validate that the delay value is within valid bounds.
    suggested_fix: >
      Validate that delay is between 1 and 30 inclusive.
  - id: F155
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1715
    line_end: 1715
    message: >
      mpStrew does not handle cases where object creation fails or returns nil.
    suggested_fix: >
      Handle object creation errors gracefully.
  - id: F156
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1725
    line_end: 1725
    message: >
      mpScatter does not ensure that the victim's position is properly set to
      POS_RESTING after teleportation.
    suggested_fix: >
      Set victim.Position = types.POS_RESTING after teleporting.
  - id: F157
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1735
    line_end: 1735
    message: >
      mpApplyAffect does not validate that the affect type is valid or that
      the location and modifier values are within expected ranges.
    suggested_fix: >
      Add validation for affect parameters (e.g., valid location, valid affect
      type).
  - id: F158
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1745
    line_end: 1745
    message: >
      mpStrew does not handle cases where room list is empty or invalid.
    suggested_fix: >
      Validate room list before processing.
  - id: F159
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1755
    line_end: 1755
    message: >
      mpScatter does not ensure that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting.
  - id: F160
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1765
    line_end: 1765
    message: >
      mpDream does not validate that the victim exists and has a valid descriptor.
    suggested_fix: >
      Validate that the victim exists and has a valid descriptor before sending
      the message.
  - id: F161
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1775
    line_end: 1775
    message: >
      mpDelay does not validate that the delay value is within valid bounds.
    suggested_fix: >
      Validate that delay is between 1 and 30 inclusive.
  - id: F162
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1785
    line_end: 1785
    message: >
      mpStrew does not handle cases where object creation fails or returns nil.
    suggested_fix: >
      Handle object creation errors gracefully.
  - id: F163
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1795
    line_end: 1795
    message: >
      mpScatter does not ensure that the victim's position is properly set to
      POS_RESTING after teleportation.
    suggested_fix: >
      Set victim.Position = types.POS_RESTING after teleporting.
  - id: F164
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1805
    line_end: 1805
    message: >
      mpApplyAffect does not validate that the affect type is valid or that
      the location and modifier values are within expected ranges.
    suggested_fix: >
      Add validation for affect parameters (e.g., valid location, valid affect
      type).
  - id: F165
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1815
    line_end: 1815
    message: >
      mpStrew does not handle cases where room list is empty or invalid.
    suggested_fix: >
      Validate room list before processing.
  - id: F166
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1825
    line_end: 1825
    message: >
      mpScatter does not ensure that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting.
  - id: F167
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1835
    line_end: 1835
    message: >
      mpDream does not validate that the victim exists and has a valid descriptor.
    suggested_fix: >
      Validate that the victim exists and has a valid descriptor before sending
      the message.
  - id: F168
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1845
    line_end: 1845
    message: >
      mpDelay does not validate that the delay value is within valid bounds.
    suggested_fix: >
      Validate that delay is between 1 and 30 inclusive.
  - id: F169
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1855
    line_end: 1855
    message: >
      mpStrew does not handle cases where object creation fails or returns nil.
    suggested_fix: >
      Handle object creation errors gracefully.
  - id: F170
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1865
    line_end: 1865
    message: >
      mpScatter does not ensure that the victim's position is properly set to
      POS_RESTING after teleportation.
    suggested_fix: >
      Set victim.Position = types.POS_RESTING after teleporting.
  - id: F171
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1875
    line_end: 1875
    message: >
      mpApplyAffect does not validate that the affect type is valid or that
      the location and modifier values are within expected ranges.
    suggested_fix: >
      Add validation for affect parameters (e.g., valid location, valid affect
      type).
  - id: F172
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1885
    line_end: 1885
    message: >
      mpStrew does not handle cases where room list is empty or invalid.
    suggested_fix: >
      Validate room list before processing.
  - id: F173
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1895
    line_end: 1895
    message: >
      mpScatter does not ensure that the victim's current room is properly
      removed from its previous location before teleporting.
    suggested_fix: >
      Ensure character is properly removed from their current room before
      teleporting.
  - id: F174
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1905
    line_end: 1905
    message: >
      mpDream does not validate that the victim exists and has a valid descriptor.
    suggested_fix: >
      Validate that the victim exists and has a valid descriptor before sending
      the message.
  - id: F175
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1915
    line_end: 1915
    message: >
      mpDelay does not validate that the delay value is within valid bounds.
    suggested_fix: >
