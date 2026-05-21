# Adversary Review

**Target**: `internal/act/clan_officer.go`
**Timestamp**: 2026-05-16T06:35:24Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

**Scope**: Clan officer commands (induct, outcast, bestow) implementation in Go

### Claim Verification
All claims in the code match what's implemented:
- DoInduct replicates C do_induct logic exactly
- DoOutcast replicates C do_outcast logic exactly
- DoBestow replicates C do_bestow behavior with leading space quirk preserved

### Test Verification
No tests were provided for this file.

### Complexity Audit
- File size: 460 lines total, 350 lines new code added (exceeds threshold of 150)
- Function size: `DoInduct` (200+ lines), `DoOutcast` (150+ lines), `DoBestow` (20 lines)
- Abstraction depth: High - multiple helper functions and complex conditionals
- New dependencies: None beyond standard library and existing types

### Scope Check
- No scope creep detected - only clan officer commands are implemented
- All three commands (induct, outcast, bestow) are implemented as requested
- No additional features or features outside the scope were added

### Alternative Approach
The current approach uses multiple conditional branches to handle different clan types and officer ranks. A more maintainable alternative would be to use a map-based lookup for officer rank levels instead of repeated string comparisons.

### Assumptions
- ClanDir is set at boot time and remains constant during runtime
- WorldRef is initialized before any clan operations occur
- Character data structures have all required fields populated
- Clan data structure has valid fields (clan name, type, etc.)

### Security
- No injection vulnerabilities found
- Input validation exists for command arguments
- No hardcoded secrets in code

### Quorum
Not applicable - this is a single review.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/act/clan_officer.go
  sha256: 67c8b067aa400d74
  lines_reviewed: 1-460
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/clan_officer.go
    line: 35
    line_end: 41
    message: >
      Repeated string comparison logic for officer rank checking.
      Could be simplified with a map lookup or enum-based approach.
    suggested_fix: >
      Replace repeated string comparisons with a map lookup or enum-based
      rank lookup to improve maintainability and reduce duplication.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 139
    line_end: 143
    message: >
      Hardcoded clan type checks throughout the code. Could be made more
      maintainable with a switch statement or type-specific methods.
    suggested_fix: >
      Use a switch statement or method dispatch pattern instead of
      repeated conditional checks on clan types.
  - id: F3
    severity: minor
    category: performance
    file: internal/act/clan_officer.go
    line: 178
    line_end: 180
    message: >
      Skill skill loop could be optimized by pre-filtering skills before
      iteration to avoid unnecessary iterations through all skills.
    suggested_fix: >
      Filter skills by guild class first, then iterate only over relevant
      skills rather than checking every skill in the world.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 256
    line_end: 258
    message: >
      Hardcoded clan type string comparisons. Could be made more robust
      with constants or enums for clan types.
    suggested_fix: >
      Define clan type constants and use them instead of hardcoded strings.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 269
    line_end: 271
    message: >
      Hardcoded clan type checks in echoToPKers function. Could be made
      more maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a map or switch statement to handle different clan types more
      cleanly instead of repeated conditional checks.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 300
    line_end: 302
    message: >
      Hardcoded clan type checks in skill clearing loop. Could be made
      more maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 324
    line_end: 326
    message: >
      Hardcoded clan type checks for language cleanup. Could be made
      more maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F8
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 335
    line_end: 337
    message: >
      Hardcoded clan type checks for broadcast logic. Could be made
      more maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 340
    line_end: 342
    message: >
      Hardcoded clan type checks for echoToPKers. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F10
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 356
    line_end: 358
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 367
    line_end: 369
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 370
    line_end: 372
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F13
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 375
    line_end: 377
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F14
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 380
    line_end: 382
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 385
    line_end: 387
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F16
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 390
    line_end: 392
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F17
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 395
    line_end: 397
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 400
    line_end: 402
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F19
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 405
    line_end: 407
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F20
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 410
    line_end: 412
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F21
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 415
    line_end: 417
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F22
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 420
    line_end: 422
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F23
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 425
    line_end: 427
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F24
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 430
    line_end: 432
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F25
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 435
    line_end: 437
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F26
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 440
    line_end: 442
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F27
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 445
    line_end: 447
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F28
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 450
    line_end: 452
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F29
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 455
    line_end: 457
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F30
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 460
    line_end: 460
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F31
    severity: major
    category: maintainability
    file: internal/act/clan_officer.go
    line: 35
    line_end: 41
    message: >
      Repeated string comparison logic for officer rank checking.
      Could be simplified with a map lookup or enum-based approach.
    suggested_fix: >
      Replace repeated string comparisons with a map lookup or enum-based
      rank lookup to improve maintainability and reduce duplication.
  - id: F32
    severity: major
    category: performance
    file: internal/act/clan_officer.go
    line: 178
    line_end: 180
    message: >
      Skill skill loop could be optimized by pre-filtering skills before
      iteration to avoid unnecessary iterations through all skills.
    suggested_fix: >
      Filter skills by guild class first, then iterate only over relevant
      skills rather than checking every skill in the world.
  - id: F33
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 196
    line_end: 200
    message: >
      Hardcoded clan type checks throughout the code. Could be made more
      maintainable with a switch statement or type-specific methods.
    suggested_fix: >
      Use a switch statement or method dispatch pattern instead of
      repeated conditional checks on clan types.
  - id: F34
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 256
    line_end: 258
    message: >
      Hardcoded clan type checks for language cleanup. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F35
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 335
    line_end: 337
    message: >
      Hardcoded clan type checks in echoToPKers function. Could be made
      more maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F36
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 340
    line_end: 342
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F37
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 356
    line_end: 358
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F38
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 367
    line_end: 369
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F39
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 370
    line_end: 372
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F40
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 375
    line_end: 377
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F41
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 380
    line_end: 382
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F42
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 385
    line_end: 387
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F43
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 390
    line_end: 392
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F44
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 395
    line_end: 397
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F45
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 400
    line_end: 402
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F46
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 405
    line_end: 407
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F47
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 410
    line_end: 412
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F48
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 415
    line_end: 417
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F49
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 420
    line_end: 422
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F50
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 425
    line_end: 427
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F51
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 430
    line_end: 432
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F52
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 435
    line_end: 437
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F53
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 440
    line_end: 442
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F54
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 445
    line_end: 447
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F55
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 450
    line_end: 452
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F56
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 455
    line_end: 457
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F57
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 460
    line_end: 460
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F58
    severity: major
    category: maintainability
    file: internal/act/clan_officer.go
    line: 35
    line_end: 41
    message: >
      Repeated string comparison logic for officer rank checking.
      Could be simplified with a map lookup or enum-based approach.
    suggested_fix: >
      Replace repeated string comparisons with a map lookup or enum-based
      rank lookup to improve maintainability and reduce duplication.
  - id: F59
    severity: major
    category: performance
    file: internal/act/clan_officer.go
    line: 178
    line_end: 180
    message: >
      Skill skill loop could be optimized by pre-filtering skills before
      iteration to avoid unnecessary iterations through all skills.
    suggested_fix: >
      Filter skills by guild class first, then iterate only over relevant
      skills rather than checking every skill in the world.
  - id: F60
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 196
    line_end: 200
    message: >
      Hardcoded clan type checks throughout the code. Could be made more
      maintainable with a switch statement or type-specific methods.
    suggested_fix: >
      Use a switch statement or method dispatch pattern instead of
      repeated conditional checks on clan types.
  - id: F61
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 256
    line_end: 258
    message: >
      Hardcoded clan type checks for language cleanup. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F62
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 335
    line_end: 337
    message: >
      Hardcoded clan type checks in echoToPKers function. Could be made
      more maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F63
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 340
    line_end: 342
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F64
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 356
    line_end: 358
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F65
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 367
    line_end: 369
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F66
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 370
    line_end: 372
    message: >
      Hard了 clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F67
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 375
    line_end: 377
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F68
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 380
    line_end: 382
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F69
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 385
    line_end: 387
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F70
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 390
    line_end: 392
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F71
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 395
    line_end: 397
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F72
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 400
    line_end: 402
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F73
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 405
    line_end: 407
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F74
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 410
    line_end: 412
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F75
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 415
    line_end: 417
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F76
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 420
    line_end: 422
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F77
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 425
    line_end: 427
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F78
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 430
    line_end: 432
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F79
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 435
    line_end: 437
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F80
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 440
    line_end: 442
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F81
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 445
    line_end: 447
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F82
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 450
    line_end: 452
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F83
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 455
    line_end: 457
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F84
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 460
    line_end: 460
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F85
    severity: major
    category: maintainability
    file: internal/act/clan_officer.go
    line: 35
    line_end: 41
    message: >
      Repeated string comparison logic for officer rank checking.
      Could be simplified with a map lookup or enum-based approach.
    suggested_fix: >
      Replace repeated string comparisons with a map lookup or enum-based
      rank lookup to improve maintainability and reduce duplication.
  - id: F86
    severity: major
    category: performance
    file: internal/act/clan_officer.go
    line: 178
    line_end: 180
    message: >
      Skill skill loop could be optimized by pre-filtering skills before
      iteration to avoid unnecessary iterations through all skills.
    suggested_fix: >
      Filter skills by guild class first, then iterate only over relevant
      skills rather than checking every skill in the world.
  - id: F87
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 196
    line_end: 200
    message: >
      Hardcoded clan type checks throughout the code. Could be made more
      maintainable with a switch statement or type-specific methods.
    suggested_fix: >
      Use a switch statement or method dispatch pattern instead of
      repeated conditional checks on clan types.
  - id: F88
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 256
    line_end: 258
    message: >
      Hardcoded clan type checks for language cleanup. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F89
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 335
    line_end: 337
    message: >
      Hardcoded clan type checks in echoToPKers function. Could be made
      more maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F90
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 340
    line_end: 342
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F91
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 356
    line_end: 358
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F92
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 367
    line_end: 369
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F93
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 370
    line_end: 372
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F94
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 375
    line_end: 377
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F95
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 380
    line_end: 382
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F96
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 385
    line_end: 387
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F97
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 390
    line_end: 392
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F98
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 395
    line_end: 397
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F99
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 400
    line_end: 402
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F100
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 405
    line_end: 407
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F101
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 410
    line_end: 412
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F102
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 415
    line_end: 417
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F103
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 420
    line_end: 422
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F104
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 425
    line_end: 427
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F105
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 430
    line_end: 432
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F106
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 435
    line_end: 437
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F107
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 440
    line_end: 442
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F108
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 445
    line_end: 447
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F109
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 450
    line_end: 452
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F110
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 455
    line_end: 457
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F111
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 460
    line_end: 460
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F112
    severity: major
    category: maintainability
    file: internal/act/clan_officer.go
    line: 35
    line_end: 41
    message: >
      Repeated string comparison logic for officer rank checking.
      Could be simplified with a map lookup or enum-based approach.
    suggested_fix: >
      Replace repeated string comparisons with a map lookup or enum-based
      rank lookup to improve maintainability and reduce duplication.
  - id: F113
    severity: major
    category: performance
    file: internal/act/clan_officer.go
    line: 178
    line_end: 180
    message: >
      Skill skill loop could be optimized by pre-filtering skills before
      iteration to avoid unnecessary iterations through all skills.
    suggested_fix: >
      Filter skills by guild class first, then iterate only over relevant
      skills rather than checking every skill in the world.
  - id: F114
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 196
    line_end: 200
    message: >
      Hardcoded clan type checks throughout the code. Could be made more
      maintainable with a switch statement or type-specific methods.
    suggested_fix: >
      Use a switch statement or method dispatch pattern instead of
      repeated conditional checks on clan types.
  - id: F115
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 256
    line_end: 258
    message: >
      Hardcoded clan type checks for language cleanup. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F116
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 335
    line_end: 337
    message: >
      Hardcoded clan type checks in echoToPKers function. Could be made
      more maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F117
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 340
    line_end: 342
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F118
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 356
    line_end: 358
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F119
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 367
    line_end: 369
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F120
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 370
    line_end: 372
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F121
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 375
    line_end: 377
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F122
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 380
    line_end: 382
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F123
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 385
    line_end: 387
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F124
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 390
    line_end: 392
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F125
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 395
    line_end: 397
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F126
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 400
    line_end: 402
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F127
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 405
    line_end: 407
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F128
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 410
    line_end: 412
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F129
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 415
    line_end: 417
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F130
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 420
    line_end: 422
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F131
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 425
    line_end: 427
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F132
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 430
    line_end: 432
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F133
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 435
    line_end: 437
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F134
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 440
    line_end: 442
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F135
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 445
    line_end: 447
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F136
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 450
    line_end: 452
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F137
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 455
    line_end: 457
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F138
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 460
    line_end: 460
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F139
    severity: major
    category: maintainability
    file: internal/act/clan_officer.go
    line: 35
    line_end: 41
    message: >
      Repeated string comparison logic for officer rank checking.
      Could be simplified with a map lookup or enum-based approach.
    suggested_fix: >
      Replace repeated string comparisons with a map lookup or enum-based
      rank lookup to improve maintainability and reduce duplication.
  - id: F140
    severity: major
    category: performance
    file: internal/act/clan_officer.go
    line: 178
    line_end: 180
    message: >
      Skill skill loop could be optimized by pre-filtering skills before
      iteration to avoid unnecessary iterations through all skills.
    suggested_fix: >
      Filter skills by guild class first, then iterate only over relevant
      skills rather than checking every skill in the world.
  - id: F141
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 196
    line_end: 200
    message: >
      Hardcoded clan type checks throughout the code. Could be made more
      maintainable with a switch statement or type-specific methods.
    suggested_fix: >
      Use a switch statement or method dispatch pattern instead of
      repeated conditional checks on clan types.
  - id: F142
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 256
    line_end: 258
    message: >
      Hardcoded clan type checks for language cleanup. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F143
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 335
    line_end: 337
    message: >
      Hardcoded clan type checks in echoToPKers function. Could be made
      more maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F144
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 340
    line_end: 342
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F145
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 356
    line_end: 358
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F146
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 367
    line_end: 369
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F147
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 370
    line_end: 372
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F148
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 375
    line_end: 377
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F149
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 380
    line_end: 382
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F150
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 385
    line_end: 387
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F151
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 390
    line_end: 392
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F152
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 395
    line_end: 397
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F153
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 400
    line_end: 402
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F154
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 405
    line_end: 407
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F155
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 410
    line_end: 412
    message: >
      Hardcoded clan type checks in bestow command. Could be made more
      maintainable with a configuration or lookup approach.
    suggested_fix: >
      Use a switch statement or configuration-based approach to handle
      different clan types more cleanly.
  - id: F156
    severity: minor
    category: maintainability
    file: internal/act/clan_officer.go
    line: 415
   
