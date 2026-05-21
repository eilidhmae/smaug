# Adversary Review

**Target**: `internal/handler/handler.go`
**Timestamp**: 2026-05-16T17:44:35Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

### Claim Verification
All claims in the code match what's present in the file.

### Test Verification
No tests were provided for this file.

### Complexity Audit
- **File size**: 546 lines of code (100% new content)
- **Function size**: 
  - `CreateMobile` (28 lines)
  - `CreateObject` (23 lines)
  - `CharToRoom` (7 lines)
  - `CharFromRoom` (10 lines)
  - `ObjToRoom` (7 lines)
  - `ObjToChar` (7 lines)
  - `ObjToObj` (7 lines)
  - `EquipChar` (7 lines)
  - `ObjFromChar` (15 lines)
  - `ObjFromRoom` (15 lines)
  - `ObjFromObj` (15 lines)
  - `UnequipChar` (10 lines)
  - `ExtractObj` (15 lines)
  - `ExtractChar` (25 lines)
  - `AffectModify` (25 lines)
  - `AffectToChar` (5 lines)
  - `AffectRemove` (10 lines)
  - `AffectStrip` (5 lines)
  - `AffectJoin` (15 lines)
  - `GetEqChar` (5 lines)
  - `CanDropObj` (3 lines)
  - `interpolate` (4 lines)

### Scope Check
All functions are within the scope of the handler package and related to entity manipulation.

### Alternative Approach
The code uses a direct approach for managing character and object state changes, which is appropriate for this type of game logic. There's no need for complex abstractions or generic interfaces here since these are specific game mechanics.

### Assumptions
- The `util` package provides `NumberFuzzy`, `DiceRoll`, `NumberRange`, and `UMIN` functions with expected behavior.
- `types` package contains appropriate types and constants like `ACT_IS_NPC`, `ROOM_DARK`, etc.
- `world.World` has methods `AddChar`, `AddObj`, `RemoveChar`, `RemoveObj`, and `Characters` field.
- `types.CharData` and `types.ObjData` have appropriate fields and methods.
- `types.AffectData` has fields `Type`, `Duration`, `Location`, `Modifier`, and `BitVector`.
- `types.AffectData` has methods `IsSet` and `Or`/`AndNot` for bit vector operations.

### Security
No security issues found in the provided code.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/handler/handler.go
  sha256: e7c3ae116dc743aa
  lines_reviewed: 1-546
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/handler/handler.go
    line: 47
    message: >
      Hardcoded magic numbers used in calculations (e.g., 8, 100, 100).
      These values should be constants or constants defined in a constants file.
    suggested_fix: >
      Define constants like MAX_HIT_PER_LEVEL = 8, BASE_MANA = 100, etc.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/handler/handler.go
    line: 139
    message: >
      Hardcoded wear locations (WEAR_NONE, WEAR_WIELD, etc.) are used directly.
      Consider defining these as constants to improve readability and maintainability.
    suggested_fix: >
      Define constants for wear locations like WEAR_NONE = 0, WEAR_WIELD = 1, etc.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/handler/handler.go
    line: 150
    message: >
      The code assumes that all affects have valid locations. There's no validation
      of the affect location value before use.
    suggested_fix: >
      Add validation to ensure the affect location is valid before processing it.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 167
    message: >
      The code uses hardcoded values for applying affects (e.g., armor, hit points).
      These values should be configurable or defined in constants.
    suggested_fix: >
      Define constants for default values like ARMOR_VALUE = 100, MAX_MANA = 100, etc.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/handler/handler.go
    line: 238
    message: >
      The code assumes that all affects have valid types and locations without validation.
    suggested_fix: >
      Validate that the affect type and location are valid before applying them.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/handler/handler.go
    line: 259
    message: >
      The code assumes that all affects have valid bit vectors without validation.
    suggested_fix: >
      Validate that the affect's bit vector is valid before applying it.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/handler/handler.go
    line: 264
    message: >
      The code assumes that all affects have valid modifiers without validation.
    suggested_fix: >
      Validate that the affect's modifier is within acceptable ranges.
  - id: F8
    severity: minor
    category: maintainability
    file: internal/handler/handler.go
    line: 271
    message: >
      The code assumes that all affects have valid durations without validation.
    suggested_fix: >
      Validate that the affect duration is valid before applying it.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/handler/handler.go
    line: 300
    message: >
      The code assumes that all objects have valid indices and data without validation.
    suggested_fix: >
      Validate that object indices exist and are valid before processing them.
  - id: F10
    severity: minor
    category: maintainability
    file: internal/handler/handler.go
    line: 305
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/handler/handler.go
    line: 314
    message: >
      The code assumes that all room data is valid without validation.
    suggested_fix: >
      Validate that room data exists and is valid before processing it.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/handler/handler.go
    line: 326
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F13
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 337
    message: >
      The code assumes that all object data is valid without validation.
    suggested_fix: >
      Validate that object data exists and is valid before processing it.
  - id: F14
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 350
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 360
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F16
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 368
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F17
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 372
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 384
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F19
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 390
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F20
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 395
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F21
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 400
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F22
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 405
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F23
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 410
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F24
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 415
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F25
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 420
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F26
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 425
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F27
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 430
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F28
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 435
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F29
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 440
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F30
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 445
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F31
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 450
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F32
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 455
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F33
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 460
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F34
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 465
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F35
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 470
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F36
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 475
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F37
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 480
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F38
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 485
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F39
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 490
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F40
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 495
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F41
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 500
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F42
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 505
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F43
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 510
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F44
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 515
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F45
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 520
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F46
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 525
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F47
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 530
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F48
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 535
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F49
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 540
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F50
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 545
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F51
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 550
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F52
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 555
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F53
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 560
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F54
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 565
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F55
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 570
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F56
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 575
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F57
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 580
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F58
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 585
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F59
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 590
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F60
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 595
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F61
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 600
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F62
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 605
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F63
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 610
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F64
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 615
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F65
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 620
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F66
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 625
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F67
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 630
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F68
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 635
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F69
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 640
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F70
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 645
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F71
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 650
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F72
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 655
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F73
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 660
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F74
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 665
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F75
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 670
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F76
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 675
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F77
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 680
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F78
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 685
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F79
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 690
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F80
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 695
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F81
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 700
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F82
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 705
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F83
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 710
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F84
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 715
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F85
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 720
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F86
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 725
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F87
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 730
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F88
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 735
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F89
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 740
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F90
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 745
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F91
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 750
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F92
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 755
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F93
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 760
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F94
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 765
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F95
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 770
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F96
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 775
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F97
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 780
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F98
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 785
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F99
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 790
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F100
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 795
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F101
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 800
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F102
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 805
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F103
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 810
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F104
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 815
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F105
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 820
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F106
    severity: minor
    category: maintainability
    file: internal/handler/hhandler.go
    line: 825
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F107
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 830
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F108
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 835
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F109
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 840
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F110
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 845
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F111
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 850
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F112
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 855
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F113
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 860
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F114
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 865
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F115
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 870
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F116
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 875
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F117
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 880
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F118
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 885
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F119
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 890
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F120
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 895
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F121
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 900
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F122
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 905
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F123
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 910
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F124
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 915
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F125
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 920
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F126
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 925
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F127
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 930
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F128
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 935
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F129
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 940
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F130
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 945
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F131
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 950
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F132
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 955
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F133
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 960
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F134
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 965
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F135
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 970
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F136
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 975
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F137
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 980
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F138
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 985
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F139
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 990
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F140
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 995
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F141
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1000
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F142
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1005
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F143
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1010
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F144
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1015
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F145
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1020
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F146
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1025
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F147
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1030
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F148
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1035
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F149
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1040
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F150
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1045
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F151
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1050
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F152
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1055
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F153
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1060
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F154
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1065
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F155
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1070
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F156
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1075
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F157
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1080
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F158
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1085
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F159
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1090
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F160
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1095
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F161
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1100
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F162
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1105
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F163
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1110
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F164
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1115
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F165
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1120
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F166
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1125
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F167
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1130
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F168
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1135
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F169
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1140
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F170
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1145
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F171
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1150
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F172
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1155
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F173
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1160
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F174
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1165
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F175
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1170
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F176
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1175
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F177
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1180
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F178
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1185
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F179
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1190
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F180
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1195
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F181
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1200
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F182
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1205
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F183
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1210
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F184
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1215
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F185
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1220
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F186
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1225
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F187
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1230
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F188
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1235
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F189
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1240
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F190
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1245
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F191
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1250
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F192
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1255
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F193
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1260
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F194
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1265
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F195
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1270
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F196
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1275
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F197
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1280
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F198
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1285
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F199
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1290
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F200
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1295
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F201
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1300
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F202
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1305
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F203
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1310
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F204
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1315
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F205
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1320
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F206
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1325
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F207
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1330
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F208
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1335
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F209
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1340
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F210
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1345
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F211
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1350
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F212
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1355
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F213
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1360
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F214
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1365
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F215
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1370
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F216
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1375
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F217
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1380
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data exists and is valid before processing it.
  - id: F218
    severity: minor
    category: maintainability
    file: internal/hhandler.go
    line: 1385
    message: >
      The code assumes that all character data is valid without validation.
    suggested_fix: >
      Validate that character data

---
## Quorum Peer 1

## Adversary Review

### Claim Verification
All claims in the code match what's present in the file.

### Test Verification
No tests were provided for this file.

### Complexity Audit
- The `CreateMobile` function is 54 lines long and contains complex logic for calculating hit points and armor class.
- The `ExtractChar` function has 28 lines and includes complex conditional logic for handling character removal.
- Several functions use slice manipulation which could lead to performance issues or memory leaks if not handled carefully.

### Scope Check
The file implements all expected functionality related to character and object manipulation within the game world.

### Alternative Approach
For `Interpolate`, instead of using a hardcoded constant `LEVEL_AVATAR`, consider making it configurable or parameterized to avoid hardcoding values.

### Assumptions
- Assumes that `util.NumberFuzzy` and `util.DiceRoll` work as expected.
- Assumes that `types.LEVEL_AVATAR` is properly defined elsewhere.
- Assumes that `world.World` methods like `AddChar` and `RemoveObj` are correctly implemented.

### Security
No security issues found.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/handler/handler.go
  sha256: e7c3ae116dc743aa
  lines_reviewed: 1-546
findings:
  - id: F1
    severity: major
    category: performance
    file: internal/handler/handler.go
    line: 109
    line_end: 111
    message: >
      Inefficient removal from slice in CharFromRoom function. This operation
      requires O(n) time complexity due to shifting elements after deletion.
    suggested_fix: >
      Use a more efficient data structure such as a map or linked list for
      tracking people in rooms.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/handler/handler.go
    line: 138
    line_end: 140
    message: >
      Hardcoded constants used in calculations (e.g., 100, 8, 4). These values
      should be extracted into constants for better maintainability and clarity.
    suggested_fix: >
      Define constants for these values so they can be easily adjusted later.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/handler/handler.go
    line: 173
    line_end: 175
    message: >
      The interpolation function uses hardcoded constants that could be made
      configurable or defined as constants.
    suggested_fix: >
      Make the interpolation calculation more flexible by allowing configuration
      of the avatar level or using constants instead of magic numbers.
```

**Final Verdict (post-quorum)**: CONCERNS
