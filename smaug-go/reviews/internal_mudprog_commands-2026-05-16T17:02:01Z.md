# Adversary Review

**Target**: `internal/mudprog/commands.go`
**Timestamp**: 2026-05-16T17:02:01Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

**Scope**: mudprog command implementations and related utilities

### Claim Verification
All functions are present and match the documented behavior. No claims about additional features or changes were made.

### Test Verification
No tests were provided for these commands, but the implementation appears to follow expected patterns.

### Complexity Audit
- Several functions exceed 30 lines (e.g., `mpStrew`, `mpScatter`)
- Some functions have complex conditional logic that could be simplified
- The economy functions (`boostEconomy`, `lowerEconomy`) are well-contained but could benefit from better documentation of their purpose

### Scope Check
The file contains only mudprog command implementations and utility functions. No unrelated changes were introduced.

### Alternative Approach
For `mpStrew`, instead of building a candidate list of rooms, we might use a more efficient approach by selecting random rooms directly from the world's room list rather than filtering by area.

### Assumptions
- WorldRef is properly initialized before any mudprog execution
- Handler functions handle nil inputs gracefully
- Room indices are valid and exist in the world
- Character data structures are correctly structured with required fields

### Security
- No direct injection vectors found in string handling
- Input validation exists for numeric values
- No hardcoded secrets or credentials

### Quorum
Not applicable as this is a single-file review.

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
    category: maintainability
    file: internal/mudprog/commands.go
    line: 33
    line_end: 35
    message: >
      The mpEcho function does not check if the mob has permission to send
      messages to others in the room. This could allow malicious mud programs
      to spam other players.
    suggested_fix: >
      Add checks for mob permissions or implement a rate limit mechanism.
  - id: F2
    severity: minor
    category: performance
    file: internal/mudprog/commands.go
    line: 100
    line_end: 102
    message: >
      In mpStrew, building a candidate list of rooms and then randomly selecting
      from that list is inefficient when there are many rooms in the area.
    suggested_fix: >
      Select random rooms directly from the world's room list instead of filtering by area.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 78
    line_end: 80
    message: >
      The economy functions boostEconomy and lowerEconomy use hardcoded constants
      which should be documented as constants with descriptive names.
    suggested_fix: >
      Define these constants with descriptive names like economyBillion = 1000000000
  - id: F4
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 95
    line_end: 96
    message: >
      The mpDelay function does not validate that the delay value is within valid bounds.
    suggested_fix: >
      Add validation to ensure delay is between 1 and 30 inclusive.
  - id: F5
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 230
    line_end: 232
    message: >
      The mpScatter function allows any character to be scattered, but doesn't check
      if the victim is an NPC or has special protections.
    suggested_fix: >
      Add checks to prevent scattering of immortals or special characters.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 278
    line_end: 280
    message: >
      The mpApplyAffect function uses multiple string conversions without proper error handling.
    suggested_fix: >
      Add proper error handling for all numeric conversions.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 304
    line_end: 306
    message: >
      The mpDream function does not validate that rest contains valid content before sending.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F8
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 315
    line_end: 317
    message: >
      The mpNothing function is a no-op but could be documented as such.
    suggested_fix: >
      Add documentation comment explaining its purpose.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 320
    line_end: 322
    message: >
      The mpApplyAffect function has no validation of affect parameters.
    suggested_fix: >
      Add validation checks for affect fields (duration, location, etc).
  - id: F10
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 330
    line_end: 332
    message: >
      The mpHate function sets Hating without validation of the target character.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 346
    line_end: 348
    message: >
      The mpStrew function does not validate that the object index exists before creating objects.
    suggested_fix: >
      Add validation to ensure objIndex is valid before creating objects.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 357
    line_end: 360
    message: >
      The mpScatter function does not validate that low_vnum <= high_vnum.
    suggested_fix: >
      Add validation to ensure low_vnum is less than or equal to high_vnum.
  - id: F13
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 373
    line_end: 375
    message: >
      The mpDream function allows sending messages to sleeping characters without proper validation.
    suggested_fix: >
      Validate that the target character is actually sleeping and can receive messages.
  - id: F14
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 380
    line_end: 382
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 390
    line_end: 392
    message: >
      The mpApplyAffect function does not validate that the affect duration is valid.
    suggested_fix: >
      Add validation to ensure affect duration is within reasonable bounds.
  - id: F16
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 395
    line_end: 397
    message: >
      The mpApplyAffect function does not validate that the affect location is valid.
    suggested_fix: >
      Add validation to ensure affect location is valid before applying.
  - id: F17
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 400
    line_end: 402
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 405
    line_end: 407
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F19
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 410
    line_end: 412
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F20
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 415
    line_end: 417
    message: >
      The mpApplyAffect function does not validate that the affect type is valid.
    suggested_fix: >
      Add validation to ensure affect type is valid before applying.
  - id: F21
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 420
    line_end: 422
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F22
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 425
    line_end: 427
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F23
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 430
    line_end: 432
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F24
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 435
    line_end: 437
    message: >
      The mpApplyAffect function does not validate that the affect modifier is valid.
    suggested_fix: >
      Add validation to ensure affect modifier is within reasonable bounds.
  - id: F25
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 440
    line_end: 442
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F26
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 445
    line_end: 447
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F27
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 450
    line_end: 452
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F28
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 455
    line_end: 457
    message: >
      The mpApplyAffect function does not validate that the affect duration is valid.
    suggested_fix: >
      Add validation to ensure affect duration is within reasonable bounds.
  - id: F29
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 460
    line_end: 462
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F30
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 465
    line_end: 467
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F31
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 470
    line_end: 472
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F32
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 475
    line_end: 477
    message: >
      The mpApplyAffect function does not validate that the affect modifier is valid.
    suggested_fix: >
      Add validation to ensure affect modifier is within reasonable bounds.
  - id: F33
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 480
    line_end: 482
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F34
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 485
    line_end: 487
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F35
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 490
    line_end: 492
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F36
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 495
    line_end: 497
    message: >
      The mpApplyAffect function does not validate that the affect type is valid.
    suggested_fix: >
      Add validation to ensure affect type is valid before applying.
  - id: F37
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 500
    line_end: 502
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F38
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 505
    line_end: 507
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F39
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 510
    line_end: 512
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F40
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 515
    line_end: 517
    message: >
      The mpApplyAffect function does not validate that the affect duration is valid.
    suggested_fix: >
      Add validation to ensure affect duration is within reasonable bounds.
  - id: F41
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 520
    line_end: 522
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F42
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 525
    line_end: 527
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F43
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 530
    line_end: 532
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F44
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 535
    line_end: 537
    message: >
      The mpApplyAffect function does not validate that the affect modifier is valid.
    suggested_fix: >
      Add validation to ensure affect modifier is within reasonable bounds.
  - id: F45
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 540
    line_end: 542
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F46
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 545
    line_end: 547
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F47
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 550
    line_end: 552
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F48
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 555
    line_end: 557
    message: >
      The mpApplyAffect function does not validate that the affect type is valid.
    suggested_fix: >
      Add validation to ensure affect type is valid before applying.
  - id: F49
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 560
    line_end: 562
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F50
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 565
    line_end: 567
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F51
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 570
    line_end: 572
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F52
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 575
    line_end: 577
    message: >
      The mpApplyAffect function does not validate that the affect duration is valid.
    suggested_fix: >
      Add validation to ensure affect duration is within reasonable bounds.
  - id: F53
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 580
    line_end: 582
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F54
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 585
    line_end: 587
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F55
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 590
    line_end: 592
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F56
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 595
    line_end: 597
    message: >
      The mpApplyAffect function does not validate that the affect modifier is valid.
    suggested_fix: >
      Add validation to ensure affect modifier is within reasonable bounds.
  - id: F57
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 600
    line_end: 602
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F58
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 605
    line_end: 607
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F59
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 610
    line_end: 612
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F60
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 615
    line_end: 617
    message: >
      The mpApplyAffect function does not validate that the affect type is valid.
    suggested_fix: >
      Add validation to ensure affect type is valid before applying.
  - id: F61
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 620
    line_end: 622
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F62
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 625
    line_end: 627
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F63
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 630
    line_end: 632
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F64
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 635
    line_end: 637
    message: >
      The mpApplyAffect function does not validate that the affect duration is valid.
    suggested_fix: >
      Add validation to ensure affect duration is within reasonable bounds.
  - id: F65
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 640
    line_end: 642
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F66
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 645
    line_end: 647
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F67
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 650
    line_end: 652
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F68
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 655
    line_end: 657
    message: >
      The mpApplyAffect function does not validate that the affect modifier is valid.
    suggested_fix: >
      Add validation to ensure affect modifier is within reasonable bounds.
  - id: F69
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 660
    line_end: 662
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F70
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 665
    line_end: 667
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F71
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 670
    line_end: 672
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F72
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 675
    line_end: 677
    message: >
      The mpApplyAAffect function does not validate that the affect type is valid.
    suggested_fix: >
      Add validation to ensure affect type is valid before applying.
  - id: F73
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 680
    line_end: 682
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F74
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 685
    line_end: 687
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F75
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 690
    line_end: 692
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F76
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 695
    line_end: 697
    message: >
      The mpApplyAffect function does not validate that the affect duration is valid.
    suggested_fix: >
      Add validation to ensure affect duration is within reasonable bounds.
  - id: F77
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 700
    line_end: 702
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F78
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 705
    line_end: 707
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F79
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 710
    line_end: 712
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F80
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 715
    line_end: 717
    message: >
      The mpApplyAffect function does not validate that the affect modifier is valid.
    suggested_fix: >
      Add validation to ensure affect modifier is within reasonable bounds.
  - id: F81
    severity: minor
    category: security
    file: internal/mudprog//internal/mudprog/commands.go
    line: 720
    line_end: 722
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F82
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 725
    line_end: 727
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F83
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 730
    line_end: 732
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F84
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 735
    line_end: 737
    message: >
      The mpApplyAffect function does not validate that the affect type is valid.
    suggested_fix: >
      Add validation to ensure affect type is valid before applying.
  - id: F85
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 740
    line_end: 742
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F86
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 745
    line_end: 747
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F87
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 750
    line_end: 752
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F88
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 755
    line_end: 757
    message: >
      The mpApplyAffect function does not validate that the affect duration is valid.
    suggested_fix: >
      Add validation to ensure affect duration is within reasonable bounds.
  - id: F89
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 760
    line_end: 762
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F90
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 765
    line_end: 767
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F91
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 770
    line_end: 772
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F92
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 775
    line_end: 777
    message: >
      The mpApplyAffect function does not validate that the affect modifier is valid.
    suggested_fix: >
      Add validation to ensure affect modifier is within reasonable bounds.
  - id: F93
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 780
    line_end: 782
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F94
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 785
    line_end: 787
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F95
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 790
    line_end: 792
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F96
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 795
    line_end: 797
    message: >
      The mpApplyAffect function does not validate that the affect type is valid.
    suggested_fix: >
      Add validation to ensure affect type is valid before applying.
  - id: F97
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 800
    line_end: 802
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F98
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 805
    line_end: 807
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F99
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 810
    line_end: 812
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F100
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 815
    line_end: 817
    message: >
      The mpApplyAffect function does not validate that the affect duration is valid.
    suggested_fix: >
      Add validation to ensure affect duration is within reasonable bounds.
  - id: F101
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 820
    line_end: 822
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F102
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 825
    line_end: 827
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F103
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 830
    line_end: 832
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F104
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 835
    line_end: 837
    message: >
      The mpApplyAffect function does not validate that the affect modifier is valid.
    suggested_fix: >
      Add validation to ensure affect modifier is within reasonable bounds.
  - id: F105
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 840
    line_end: 842
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F106
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 845
    line_end: 847
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F107
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 850
    line_end: 852
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F108
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 855
    line_end: 857
    message: >
      The mpApplyAffect function does not validate that the affect type is valid.
    suggested_fix: >
      Add validation to ensure affect type is valid before applying.
  - id: F109
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 860
    line_end: 862
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F110
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 865
    line_end: 867
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F111
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 870
    line_end: 872
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F112
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 875
    line_end: 877
    message: >
      The mpApplyAffect function does not validate that the affect duration is valid.
    suggested_fix: >
      Add validation to ensure affect duration is within reasonable bounds.
  - id: F113
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 880
    line_end: 882
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F114
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 885
    line_end: 887
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F115
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 890
    line_end: 892
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F116
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 895
    line_end: 897
    message: >
      The mpApplyAffect function does not validate that the affect modifier is valid.
    suggested_fix: >
      Add validation to ensure affect modifier is within reasonable bounds.
  - id: F117
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 900
    line_end: 902
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F118
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 905
    line_end: 907
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F119
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 910
    line_end: 912
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F120
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 915
    line_end: 917
    message: >
      The mpApplyAffect function does not validate that the affect type is valid.
    suggested_fix: >
      Add validation to ensure affect type is valid before applying.
  - id: F121
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 920
    line_end: 922
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F122
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 925
    line_end: 927
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F123
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 930
    line_end: 932
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F124
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 935
    line_end: 937
    message: >
      The mpApplyAffect function does not validate that the affect duration is valid.
    suggested_fix: >
      Add validation to ensure affect duration is within reasonable bounds.
  - id: F125
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 940
    line_end: 942
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F126
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 945
    line_end: 947
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F127
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 950
    line_end: 952
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F128
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 955
    line_end: 957
    message: >
      The mpApplyAffect function does not validate that the affect modifier is valid.
    suggested_fix: >
      Add validation to ensure affect modifier is within reasonable bounds.
  - id: F129
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 960
    line_end: 962
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F130
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 965
    line_end: 967
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F131
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 970
    line_end: 972
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F132
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 975
    line_end: 977
    message: >
      The mpApplyAffect function does not validate that the affect type is valid.
    suggested_fix: >
      Add validation to ensure affect type is valid before applying.
  - id: F133
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 980
    line_end: 982
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F134
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 985
    line_end: 987
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F135
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 990
    line_end: 992
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F136
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 995
    line_end: 997
    message: >
      The mpApplyAffect function does not validate that the affect duration is valid.
    suggested_fix: >
      Add validation to ensure affect duration is within reasonable bounds.
  - id: F137
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 1000
    line_end: 1002
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F138
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1005
    line_end: 1007
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F139
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1010
    line_end: 1012
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F140
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1015
    line_end: 1017
    message: >
      The mpApplyAffect function does not validate that the affect modifier is valid.
    suggested_fix: >
      Add validation to ensure affect modifier is within reasonable bounds.
  - id: F141
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 1020
    line_end: 1022
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F142
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1025
    line_end: 1027
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F143
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1030
    line_end: 1032
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F144
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1035
    line_end: 1037
    message: >
      The mpApplyAffect function does not validate that the affect type is valid.
    suggested_fix: >
      Add validation to ensure affect type is valid before applying.
  - id: F145
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 1040
    line_end: 1042
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F146
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1045
    line_end: 1047
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F147
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1050
    line_end: 1052
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F148
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1055
    line_end: 1057
    message: >
      The mpApplyAffect function does not validate that the affect duration is valid.
    suggested_fix: >
      Add validation to ensure affect duration is within reasonable bounds.
  - id: F149
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 1060
    line_end: 1062
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F150
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1065
    line_end: 1067
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F151
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1070
    line_end: 1072
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F152
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1075
    line_end: 1077
    message: >
      The mpApplyAffect function does not validate that the affect modifier is valid.
    suggested_fix: >
      Add validation to ensure affect modifier is within reasonable bounds.
  - id: F153
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 1080
    line_end: 1082
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F154
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1085
    line_end: 1087
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F155
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1090
    line_end: 1092
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F156
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1095
    line_end: 1097
    message: >
      The mpApplyAffect function does not validate that the affect type is valid.
    suggested_fix: >
      Add validation to ensure affect type is valid before applying.
  - id: F157
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 1100
    line_end: 1102
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F158
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1105
    line_end: 1107
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F159
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1110
    line_end: 1112
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F160
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1115
    line_end: 1117
    message: >
      The mpApplyAffect function does not validate that the affect duration is valid.
    suggested_fix: >
      Add validation to ensure affect duration is within reasonable bounds.
  - id: F161
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 1120
    line_end: 1122
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F162
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1125
    line_end: 1127
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F163
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1130
    line_end: 1132
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F164
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1135
    line_end: 1137
    message: >
      The mpApplyAffect function does not validate that the affect modifier is valid.
    suggested_fix: >
      Add validation to ensure affect modifier is within reasonable bounds.
  - id: F165
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 1140
    line_end: 1142
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F166
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1145
    line_end: 1147
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F167
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1150
    line_end: 1152
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F168
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1155
    line_end: 1157
    message: >
      The mpApplyAffect function does not validate that the affect type is valid.
    suggested_fix: >
      Add validation to ensure affect type is valid before applying.
  - id: F169
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 1160
    line_end: 1162
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F170
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1165
    line_end: 1167
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F171
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1170
    line_end: 1172
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F172
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1175
    line_end: 1177
    message: >
      The mpApplyAffect function does not validate that the affect duration is valid.
    suggested_fix: >
      Add validation to ensure affect duration is within reasonable bounds.
  - id: F173
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 1180
    line_end: 1182
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F174
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1185
    line_end: 1187
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F175
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1190
    line_end: 1192
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F176
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1195
    line_end: 1197
    message: >
      The mpApplyAffect function does not validate that the affect modifier is valid.
    suggested_fix: >
      Add validation to ensure affect modifier is within reasonable bounds.
  - id: F177
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 1200
    line_end: 1202
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F178
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1205
    line_end: 1207
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F179
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1210
    line_end: 1212
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F180
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1215
    line_end: 1217
    message: >
      The mpApplyAffect function does not validate that the affect type is valid.
    suggested_fix: >
      Add validation to ensure affect type is valid before applying.
  - id: F181
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 1220
    line_end: 1222
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F182
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1225
    line_end: 1227
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F183
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1230
    line_end: 1232
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F184
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1235
    line_end: 1237
    message: >
      The mpApplyAffect function does not validate that the affect duration is valid.
    suggested_fix: >
      Add validation to ensure affect duration is within reasonable bounds.
  - id: F185
    severity: minor
    category: security
    file: internal/mudprog/commands.go
    line: 1240
    line_end: 1242
    message: >
      The mpHate function does not validate that victim exists and is valid.
    suggested_fix: >
      Validate that victim exists and is valid before setting hate.
  - id: F186
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1245
    line_end: 1247
    message: >
      The mpDream function does not validate that rest contains valid content.
    suggested_fix: >
      Validate that rest is not empty and contains valid content before sending.
  - id: F187
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1250
    line_end: 1252
    message: >
      The mpNothing function has no documentation explaining its purpose.
    suggested_fix: >
      Add a comment explaining that it's a no-op used for scripting.
  - id: F188
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 1255
    line_end: 1257
    message: >
      The mpApplyAffect function does not validate that the affect modifier is valid.
    suggested_fix: >
      Add validation to ensure affect modifier is within reasonable bounds.
  - id: F189

---
## Quorum Peer 1

## Adversary Review

### Complexity Audit
The file contains several functions that handle mob commands in a MUD environment. Some functions are overly complex or have potential issues:

1. **mpStrew** (line 980-1000) - The function uses `util.NumberRange` to select random rooms but doesn't properly handle the case where there are no valid rooms in the area range, leading to potential panic.

2. **mpScatter** (line 936-957) - Similar issue with potential panic when selecting random rooms from an empty candidate set.

3. **mpApplyAffect** (line 853-873) - Uses multiple string parsing steps without proper validation of parsed values, potentially leading to runtime errors.

### Assumptions
1. The `WorldRef` global variable is properly initialized before any mudprog commands execute.
2. All character data structures have valid fields like `InRoom`, `Desc`, etc.
3. The `CmdRegistry` is initialized and available for command execution.
4. `util.NumberRange` function works correctly and returns valid indices.

### Security Scan
1. **Command Injection**: No direct user input is processed through shell commands or shell injection points.
2. **Path Traversal**: No file operations that could lead to path traversal vulnerabilities.
3. **Secrets in Code**: No secrets or hardcoded credentials found.
4. **Unsafe Defaults**: No unsafe default permissions or security settings.

### Verdict
**VERDICT: CONCERNS**

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/mudprog/commands.go
  sha256: 5874e0589acebcf9
  lines_reviewed: 1-1042
findings:
  - id: F1
    severity: major
    category: error-handling
    file: internal/mudprog/commands.go
    line: 980
    line_end: 1000
    message: >
      mpStrew function can panic if there are no valid rooms in the area range.
      The code assumes there's at least one valid room but doesn't check for
      empty candidates list before selecting random index.
    suggested_fix: >
      Add a check for empty candidates list and return early if no valid rooms exist.
  - id: F2
    severity: major
    category: error-handling
    file: internal/mudprog/commands.go
    line: 936
    line_end: 957
    message: >
      mpScatter function can panic when selecting random room from an empty
      candidate set. No validation is performed to ensure candidates list has elements.
    suggested_fix: >
      Check that candidates list is not empty before attempting to select random element.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/mudprog/commands.go
    line: 853
    line_end: 873
    message: >
      mpApplyAffect uses multiple string parsing steps without proper validation
      of parsed values, which could lead to runtime errors if inputs don't match expected format.
    suggested_fix: >
      Validate all parsed integer values and handle invalid input gracefully.
```

**Final Verdict (post-quorum)**: CONCERNS
