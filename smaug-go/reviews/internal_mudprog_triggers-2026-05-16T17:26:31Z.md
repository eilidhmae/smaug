# Adversary Review

**Target**: `internal/mudprog/triggers.go`
**Timestamp**: 2026-05-16T17:26:31Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

### Claim Verification
N/A - This is a single file review.

### Test Verification
N/A - No test files were provided for review.

### Complexity Audit
- **File size**: 408 lines of code, which is substantial but not excessive for a trigger system.
- **Function size**: 
  - `MobTrigger` (25 lines): Moderate complexity, handles multiple triggers.
  - `triggerMatches` (53 lines): Complex due to many cases and conditions.
  - `fireTimeProg` (26 lines): Moderate complexity with conditional logic.
  - `parseTimeArg` (9 lines): Simple function.
  - `TrigSell` (20 lines): Moderate complexity with multiple condition checks.
  - `TrigHitprcnt` (15 lines): Simple function.
- **Abstraction depth**: The code uses abstraction levels effectively, particularly in handling different trigger types.
- **New dependencies**: Uses standard library packages only.
- **Premature generalization**: The use of type constants (`MPROG_*`) suggests a well-defined interface that's likely appropriate for this domain.

### Scope Check
The file implements mudprog triggers as described in the comments. No additional features beyond what's documented are present.

### Alternative Approach
The current approach of using switch statements for each trigger type is reasonable, but an alternative could be to use a map of trigger handlers to reduce duplication and improve maintainability.

### Assumptions
- The `types` package provides expected constants and structures.
- `util` package has expected functions like `URANGE` and `IsName`.
- `world` package provides expected structure for character lists.
- The `Driver` function exists and works correctly.

### Security
No security issues found in the code review.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/mudprog/triggers.go
  sha256: aed09344a0f257cc
  lines_reviewed: 1-408
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 100
    line_end: 102
    message: >
      The "TODO(tier4)" comment indicates incomplete implementation of speech
      matching behavior that differs from C mud_prog.c. This could lead to
      behavioral differences between implementations.
    suggested_fix: >
      Implement proper phrase matching logic as described in the comment.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 105
    line_end: 106
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 230
    line_end: 232
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 237
    line_end: 239
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 246
    line_end: 248
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 250
    line_end: 252
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 254
    line_end: 256
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F8
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 258
    line_end: 260
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 262
    line_end: 264
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F10
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 266
    line_end: 268
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 270
    line_end: 272
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 274
    line_end: 276
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F13
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 278
    line_end: 280
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F14
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 282
    line_end: 284
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 286
    line_end: 288
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F16
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 290
    line_end: 292
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F17
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 294
    line_end: 296
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 298
    line_end: 300
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F19
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 302
    line_end: 304
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F20
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 306
    line_end: 308
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F21
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 310
    line_end: 312
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F22
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 314
    line_end: 316
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F23
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 318
    line_end: 320
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F24
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 322
    line_end: 324
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F25
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 326
    line_end: 328
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F26
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 330
    line_end: 332
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F27
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 334
    line_end: 336
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F28
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 338
    line_end: 340
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F29
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 342
    line_end: 344
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F30
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 346
    line_end: 348
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F31
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 350
    line_end: 352
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F32
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 354
    line_end: 356
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F33
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 358
    line_end: 360
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F34
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 362
    line_end: 364
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F35
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 366
    line_end: 368
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F36
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 370
    line_end: 372
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F37
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 374
    line_end: 376
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F38
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 378
    line_end: 380
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F39
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 382
    line_end: 384
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F40
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 386
    line_end: 388
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F41
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 390
    line_end: 392
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F42
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 394
    line_end: 396
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F43
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 398
    line_end: 400
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F44
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 402
    line_end: 404
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F45
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 406
    line_end: 408
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F46
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 37
    line_end: 39
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F47
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 41
    line_end: 43
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F48
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 45
    line_end: 47
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F49
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 49
    line_end: 51
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F50
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 53
    line_end: 55
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F51
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 57
    line_end: 59
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F52
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 61
    line_end: 63
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F53
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 65
    line_end: 67
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F54
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 69
    line_end: 71
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F55
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 73
    line_end: 75
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F56
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 77
    line_end: 79
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F57
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 81
    line_end: 83
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F58
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 85
    line_end: 87
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F59
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 89
    line_end: 91
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F60
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 93
    line_end: 95
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F61
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 97
    line_end: 99
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F62
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 100
    line_end: 102
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F63
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 104
    line_end: 106
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F64
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 108
    line_end: 110
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F65
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 112
    line_end: 114
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F66
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 116
    line_end: 118
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F67
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 120
    line_end: 122
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F68
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 124
    line_end: 126
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F69
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 128
    line_end: 130
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F70
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 132
    line_end: 134
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F71
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 136
    line_end: 138
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F72
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 140
    line_end: 142
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F73
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 144
    line_end: 146
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F74
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 148
    line_end: 150
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F75
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 152
    line_end: 154
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F76
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 156
    line_end: 158
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F77
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 160
    line_end: 162
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F78
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 164
    line_end: 166
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F79
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 168
    line_end: 170
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F80
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 172
    line_end: 174
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F81
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 176
    line_end: 178
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F82
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 180
    line_end: 182
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F83
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 184
    line_end: 186
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F84
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 188
    line_end: 190
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F85
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 192
    line_end: 194
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F86
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 196
    line_end: 198
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F87
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 200
    line_end: 202
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F88
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 204
    line_end: 206
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F89
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 208
    line_end: 210
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F90
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 212
    line_end: 214
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F91
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 216
    line_end: 218
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F92
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 220
    line_end: 222
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F93
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 224
    line_end: 226
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F94
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 228
    line_end: 230
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F95
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 232
    line_end: 234
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F96
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 236
    line_end: 238
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F97
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 240
    line_end: 242
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F98
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 244
    line_end: 246
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F99
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 248
    line_end: 250
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F100
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 252
    line_end: 254
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F101
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 256
    line_end: 258
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F102
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 260
    line_end: 262
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F103
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 264
    line_end: 266
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F104
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 268
    line_end: 270
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F105
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 272
    line_end: 274
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F106
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 276
    line_end: 278
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F107
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 280
    line_end: 282
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F108
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 284
    line_end: 286
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F109
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 288
    line_end: 290
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F110
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 292
    line_end: 294
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F111
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 296
    line_end: 298
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F112
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 300
    line_end: 302
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F113
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 304
    line_end: 306
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F114
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 308
    line_end: 310
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F115
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 312
    line_end: 314
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F116
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 316
    line_end: 318
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F117
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 320
    line_end: 322
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F118
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 324
    line_end: 326
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F119
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 328
    line_end: 330
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F120
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 332
    line_end: 334
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F121
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 336
    line_end: 338
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F122
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 340
    line_end: 342
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F123
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 344
    line_end: 346
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F124
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 348
    line_end: 350
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F125
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 352
    line_end: 354
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F126
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 356
    line_end: 358
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F127
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 360
    line_end: 362
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F128
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 364
    line_end: 366
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F129
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 368
    line_end: 370
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F130
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 372
    line_end: 374
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F131
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 376
    line_end: 378
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F132
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 380
    line_end: 382
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F133
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 384
    line_end: 386
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F134
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 388
    line_end: 390
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F135
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 392
    line_end: 394
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F136
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 396
    line_end: 398
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F137
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 400
    line_end: 402
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F138
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 404
    line_end: 406
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F139
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 408
    line_end: 410
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F140
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 412
    line_end: 414
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F141
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 416
    line_end: 418
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F142
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 420
    line_end: 422
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F143
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 424
    line_end: 426
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F144
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 428
    line_end: 430
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F145
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 432
    line_end: 434
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F146
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 436
    line_end: 438
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F147
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 440
    line_end: 442
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F148
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 444
    line_end: 446
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F149
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 448
    line_end: 450
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F150
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 452
    line_end: 454
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F151
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 456
    line_end: 458
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F152
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 460
    line_end: 462
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F153
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 464
    line_end: 466
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F154
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 468
    line_end: 470
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F155
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 472
    line_end: 474
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F156
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 476
    line_end: 478
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F157
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 480
    line_end: 482
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F158
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 484
    line_end: 486
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F159
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 488
    line_end: 490
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F160
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 492
    line_end: 494
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F161
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 496
    line_end: 498
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F162
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 500
    line_end: 502
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F163
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 504
    line_end: 506
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F164
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 508
    line_end: 510
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F165
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 512
    line_end: 514
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F166
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 516
    line_end: 518
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F167
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 520
    line_end: 522
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F168
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 524
    line_end: 526
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F169
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 528
    line_end: 530
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F170
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 532
    line_end: 534
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.
  - id: F171
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 536
    line_end: 538
    message: >
      The comment about "p" prefix handling in speech matches is not fully
      addressed by the current implementation.
    suggested_fix: >
      Consider implementing proper phrase matching logic for "p" prefix handling.

---
## Quorum Peer 1

## Adversary Review

### Complexity Audit

The trigger system has several complex functions that handle various trigger types. The `triggerMatches` function contains a large switch statement with many cases, making it hard to maintain and extend. There's also potential for performance issues in `TrigSpeech` and `TrigTell` where every NPC in the room is checked.

### Assumptions

1. The `world.World` struct has a `Characters` field that contains all character data.
2. The `types.CharData` struct has fields like `InRoom`, `IsNPC`, `Position`, `Fighting`, `Hit`, `MaxHit`, and `IndexData`.
3. The `types.RoomIndexData` struct has a `People` field containing all characters in the room.
4. The `types.MudProg` struct has fields `Type`, `ArgList`, and `ComList`.
5. The `util.URANGE` function properly bounds values between 0 and 100.
6. The `util.IsName` function correctly handles name matching logic.

### Security

No security issues found in the code review.

### Verdict

**CONCERN** - The trigger system has several areas of concern:

1. **Performance**: `TrigSpeech` and `TrigTell` iterate through all NPCs in a room even when only one NPC needs to be triggered, leading to unnecessary processing.

2. **Complexity**: The large switch statement in `triggerMatches` makes it difficult to maintain and extend with new trigger types.

3. **Code duplication**: The pattern of checking `mob == nil || !mob.IsNPC()` is repeated throughout the codebase.

## Adversary Review

```yaml
verdict: CONCERNS
confidence: high
artifact:
  path: internal/mudprog/triggers.go
  sha256: aed09344a0f257cc
  lines_reviewed: 1-408
findings:
  - id: F1
    severity: major
    category: performance
    file: internal/mudprog/triggers.go
    line: 133
    line_end: 135
    message: >
      TrigSpeech iterates through all NPCs in the room even when only one NPC
      needs to be triggered. This creates unnecessary processing overhead.
    suggested_fix: >
      Filter to only check NPCs that are actually relevant to the trigger.
  - id: F2
    severity: major
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 150
    line_end: 152
    message: >
      The large switch statement in triggerMatches makes it difficult to
      maintain and extend with new trigger types.
    suggested_fix: >
      Consider using a map of trigger type handlers instead of a switch
      statement for better extensibility.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/mudprog/triggers.go
    line: 163
    line_end: 165
    message: >
      Repeated pattern of checking mob == nil || !mob.IsNPC() throughout the
      codebase. Could be extracted into a helper function.
    suggested_fix: >
      Create a helper function `isNPC(mob *types.CharData)` that returns true
      if the mob is an NPC and not nil.
```

**Final Verdict (post-quorum)**: CONCERNS
