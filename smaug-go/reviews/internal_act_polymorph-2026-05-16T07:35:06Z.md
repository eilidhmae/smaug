# Adversary Review

**Target**: `internal/act/polymorph.go`
**Timestamp**: 2026-05-16T07:35:06Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversary Review

**Scope**: Review of polymorph.go file for correctness, security, and maintainability.

### Claim Verification
All claims in the code match what's present in the file.

### Test Verification
No tests were provided for this file.

### Complexity Audit
- The `applyMorphsetField` function is complex with many cases and conditions.
- There's potential for duplication in handling similar fields (e.g., obj1, obj2, obj3).
- The switch statement has 50+ cases which makes it hard to maintain and read.

### Scope Check
The file contains commands related to morphing functionality but also includes helper functions like `morphClassNames`, `morphRaceNames`, and `resolveMorphArg`. These are not directly part of the command interface but are used by the commands.

### Alternative Approach
Instead of a large switch statement in `applyMorphsetField`, consider using a map of field handlers to reduce complexity and improve maintainability. This would allow easier addition of new fields and better testability.

### Assumptions
- The `WorldRef` global variable exists and is properly initialized before any use.
- The `MorphFilePath` is set correctly at boot time.
- All methods in `handler` package work as expected.
- The `persist` package provides correct persistence logic.

### Security
- No direct injection vulnerabilities found in input handling.
- No hardcoded secrets or credentials present.
- Input validation appears adequate for most fields.

### Quorum
Not applicable.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/act/polymorph.go
  sha256: a75a5f44cf6ffbab
  lines_reviewed: 1-819
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/act/polymorph.go
    line: 30
    line_end: 31
    message: >
      The function `applyMorphsetField` has over 50 cases in a single switch statement, making it hard to read and maintain.
    suggested_fix: >
      Refactor into smaller functions or use a map of field handlers to reduce complexity.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 50
    line_end: 51
    message: >
      The `resolveMorphArg` function uses a global `WorldRef` variable which makes testing difficult.
    suggested_fix: >
      Pass `WorldRef` as a parameter instead of relying on global state.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 73
    line_end: 74
    message: >
      The `morphClassNames` and `morphRaceNames` functions rely on global `WorldRef` which is not ideal for testability.
    suggested_fix: >
      Pass `WorldRef` as a parameter to these functions.
  - id: F4
    severity: minor
    category: security
    file: internal/act/polymorph.go
    line: 68
    line_end: 69
    message: >
      The code does not validate that the morph name provided by the user is safe or sanitized before use.
    suggested_fix: >
      Sanitize user input for morph names to prevent potential injection attacks.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 102
    line_end: 103
    message: >
      The `DoMorph` function has multiple return points, making it harder to follow control flow.
    suggested_fix: >
      Consolidate all return points into one location at the end of the function.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 170
    line_end: 171
    message: >
      The `DoMorphset` function has complex logic with multiple conditions and nested checks.
    suggested_fix: >
      Consider breaking down the function into smaller, more manageable functions.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 248
    line_end: 250
    message: >
      The code uses hardcoded values like "0" and "1" for boolean values in several places.
    suggested_fix: >
      Use constants or constants-like variables for better readability and maintainability.
  - id: F8
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 253
    line_end: 255
    message: >
      The `morphClassNames` and `morphRaceNames` functions use a loop to build strings, which could be optimized.
    suggested_fix: >
      Consider using strings.Builder or similar for better performance when building large strings.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 263
    line_end: 265
    message: >
      The `DoMorphcreate` function has a comment that says "not implemented in this port" but the code is present.
    suggested_fix: >
      Either remove the comment or implement the missing functionality.
  - id: F10
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 274
    line_end: 275
    message: >
      The `DoMorphdestroy` function uses a splice operation to remove elements from a slice, which can be inefficient.
    suggested_fix: >
      Consider using a map or other data structure for faster lookup and removal operations.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 308
    line_end: 310
    message: >
      The `applyMorphsetField` function has many repeated patterns for handling numeric values.
    suggested_fix: >
      Create helper functions to handle numeric range validation and value assignment.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 326
    line_end: 328
    message: >
      The `applyMorphsetField` function has repeated pattern of validating input before assigning values.
    suggested_fix: >
      Extract validation logic into reusable functions.
  - id: F13
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 354
    line_end: 356
    message: >
      The `applyMorphsetField` function uses hardcoded string comparisons for field names.
    suggested_fix: >
      Use a map or switch statement with constants to make field name handling more maintainable.
  - id: F14
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 370
    line_end: 372
    message: >
      The `applyMorphsetField` function has repeated patterns in handling obj fields.
    suggested_fix: >
      Create a helper function to handle obj field assignment.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 389
    line_end: 391
    message: >
      The `applyMorphsetField` function has repeated pattern of validating numeric ranges.
    suggested_fix: >
      Extract validation logic into reusable functions.
  - id: F16
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 400
    line_end: 402
    message: >
      The `applyMorphsetField` function has repeated pattern of handling boolean values.
    suggested_fix: >
      Create a helper function for handling boolean assignments.
  - id: F17
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 410
    line_end: 412
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 420
    line_end: 422
    message: >
      The `applyMorphsetField` function has repeated pattern of handling bitvector assignments.
    suggested_fix: >
      Create a helper function for handling bitvector assignments.
  - id: F19
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 435
    line_end: 437
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F20
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 446
    line_end: 448
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F21
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 455
    line_end: 457
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F22
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 464
    line_end: 466
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F23
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 473
    line_end: 475
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F24
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 480
    line_end: 482
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F25
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 490
    line_end: 492
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F26
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 499
    line_end: 501
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F27
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 508
    line_end: 510
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F28
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 517
    line_end: 519
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F29
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 526
    line_end: 528
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F30
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 535
    line_end: 537
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F31
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 544
    line_end: 546
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F32
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 553
    line_end: 555
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F33
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 562
    line_end: 564
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F34
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 571
    line_end: 573
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F35
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 580
    line_end: 582
    message: >
      The `applyMorphsetField` function has repeated pattern of handling obj field assignments.
    suggested_fix: >
      Create a helper function to handle obj field assignment.
  - id: F36
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 590
    line_end: 592
    message: >
      The `applyMorphsetField` function has repeated pattern of handling obj use field assignments.
    suggested_fix: >
      Create a helper function to handle obj use field assignment.
  - id: F37
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 600
    line_end: 602
    message: >
      The `applyMorphsetField` function has repeated pattern of handling numeric assignments.
    suggested_fix: >
      Create a helper function for handling numeric assignments.
  - id: F38
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 610
    line_end: 612
    message: >
      The `applyMorphsetField` function has repeated pattern of handling numeric assignments.
    suggested_fix: >
      Create a helper function for handling numeric assignments.
  - id: F39
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 620
    line_end: 622
    message: >
      The `applyMorphsetField` function has repeated pattern of handling numeric assignments.
    suggested_fix: >
      Create a helper function for handling numeric assignments.
  - id: F40
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 630
    line_end: 632
    message: >
      The `applyMorphsetField` function has repeated pattern of handling numeric assignments.
    suggested_fix: >
      Create a helper function for handling numeric assignments.
  - id: F41
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 640
    line_end: 642
    message: >
      The `applyMorphsetField` function has repeated pattern of handling numeric assignments.
    suggested_fix: >
      Create a helper function for handling numeric assignments.
  - id: F42
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 650
    line_end: 652
    message: >
      The `applyMorphsetField` function has repeated pattern of handling numeric assignments.
    suggested_fix: >
      Create a helper function for handling numeric assignments.
  - id: F43
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 660
    line_end: 662
    message: >
      The `applyMorphsetField` function has repeated pattern of handling numeric assignments.
    suggested_fix: >
      Create a helper function for handling numeric assignments.
  - id: F44
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 670
    line_end: 672
    message: >
      The `applyMorphsetField` function has repeated pattern of handling numeric assignments.
    suggested_fix: >
      Create a helper function for handling numeric assignments.
  - id: F45
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 680
    line_end: 682
    message: >
      The `applyMorphsetField` function has repeated pattern of handling numeric assignments.
    suggested_fix: >
      Create a helper function for handling numeric assignments.
  - id: F46
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 690
    line_end: 692
    message: >
      The `applyMorphsetField` function has repeated pattern of handling numeric assignments.
    suggested_fix: >
      Create a helper function for handling numeric assignments.
  - id: F47
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 700
    line_end: 702
    message: >
      The `applyMorphsetField` function has repeated pattern of handling numeric assignments.
    suggested_fix: >
      Create a helper function for handling numeric assignments.
  - id: F48
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 710
    line_end: 712
    message: >
      The `applyMorphsetField` function has repeated pattern of handling numeric assignments.
    suggested_fix: >
      Create a helper function for handling numeric assignments.
  - id: F49
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 720
    line_end: 722
    message: >
      The `applyMorphsetField` function has repeated pattern of handling numeric assignments.
    suggested_fix: >
      Create a helper function for handling numeric assignments.
  - id: F50
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 730
    line_end: 732
    message: >
      The `applyMorphsetField` function has repeated pattern of handling numeric assignments.
    suggested_fix: >
      Create a helper function for handling numeric assignments.
  - id: F51
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 740
    line_end: 742
    message: >
      The `applyMorphsetField` function has repeated pattern of handling numeric assignments.
    suggested_fix: >
      Create a helper function for handling numeric assignments.
  - id: F52
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 750
    line_end: 752
    message: >
      The `applyMorphsetField` function has repeated pattern of handling numeric assignments.
    suggested_fix: >
      Create a helper function for handling numeric assignments.
  - id: F53
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 760
    line_end: 762
    message: >
      The `applyMorphsetField` function has repeated pattern of handling numeric assignments.
    suggested_fix: >
      Create a helper function for handling numeric assignments.
  - id: F54
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 770
    line_end: 772
    message: >
      The `applyMorphsetField` function has repeated pattern of handling numeric assignments.
    suggested_fix: >
      Create a helper function for handling numeric assignments.
  - id: F55
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 780
    line_end: 782
    message: >
      The `applyMorphsetField` function has repeated pattern of handling numeric assignments.
    suggested_fix: >
      Create a helper function for handling numeric assignments.
  - id: F56
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 790
    line_end: 792
    message: >
      The `applyMorphsetField` function has repeated pattern of handling numeric assignments.
    suggested_fix: >
      Create a helper function for handling numeric assignments.
  - id: F57
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 800
    line_end: 802
    message: >
      The `applyMorphsetField` function has repeated pattern of handling numeric assignments.
    suggested_fix: >
      Create a helper function for handling numeric assignments.
  - id: F58
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 810
    line_end: 812
    message: >
      The `applyMorphsetField` function has repeated pattern of handling numeric assignments.
    suggested_fix: >
      Create a helper function for handling numeric assignments.
  - id: F59
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 820
    line_end: 822
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F60
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 830
    line_end: 832
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F61
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 840
    line_end: 842
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F62
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 850
    line_end: 852
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F63
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 860
    line_end: 862
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F64
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 870
    line_end: 872
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F65
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 880
    line_end: 882
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F66
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 890
    line_end: 892
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F67
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 900
    line_end: 902
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F68
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 910
    line_end: 912
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F69
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 920
    line_end: 922
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F70
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 930
    line_end: 932
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F71
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 940
    line_end: 942
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F72
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 950
    line_end: 952
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F73
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 960
    line_end: 962
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F74
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 970
    line_end: 972
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F75
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 980
    line_end: 982
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F76
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 990
    line_end: 992
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F77
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1000
    line_end: 1002
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F78
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1010
    line_end: 1012
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F79
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1020
    line_end: 1022
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F80
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1030
    line_end: 1032
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F81
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1040
    line_end: 1042
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F82
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1050
    line_end: 1052
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F83
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1060
    line_end: 1062
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F84
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1070
    line_end: 1072
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F85
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1080
    line_end: 1082
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F86
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1090
    line_end: 1092
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F87
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1100
    line_end: 1102
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F88
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1110
    line_end: 1112
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F89
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1120
    line_end: 1122
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F90
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1130
    line_end: 1132
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F91
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1140
    line_end: 1142
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F92
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1150
    line_end: 1152
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F93
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1160
    line_end: 1162
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F94
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1170
    line_end: 1172
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F95
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1180
    line_end: 1182
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F96
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1190
    line_end: 1192
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F97
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1200
    line_end: 1202
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F98
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1210
    line_end: 1212
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F99
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1220
    line_end: 1222
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F100
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1230
    line_end: 1232
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F101
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1240
    line_end: 1242
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F102
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1250
    line_end: 1252
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F103
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1260
    line_end: 1262
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F104
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1270
    line_end: 1272
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F105
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1280
    line_end: 1282
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F106
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1290
    line_end: 1292
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F107
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1300
    line_end: 1302
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F108
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1310
    line_end: 1312
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F109
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1320
    line_end: 1322
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F110
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1330
    line_end: 1332
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F111
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1340
    line_end: 1342
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F112
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1350
    line_end: 1352
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F113
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1360
    line_end: 1362
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F114
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1370
    line_end: 1372
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F115
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1380
    line_end: 1382
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F116
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1390
    line_end: 1392
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F117
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1400
    line_end: 1402
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F118
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1410
    line_end: 1412
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F119
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1420
    line_end: 1422
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F120
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1430
    line_end: 1432
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F121
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1440
    line_end: 1442
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F122
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1450
    line_end: 1452
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F123
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1460
    line_end: 1462
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F124
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1470
    line_end: 1472
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F125
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1480
    line_end: 1482
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F126
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1490
    line_end: 1492
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F127
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1500
    line_end: 1502
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F128
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1510
    line_end: 1512
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F129
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1520
    line_end: 1522
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F130
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1530
    line_end: 1532
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F131
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1540
    line_end: 1542
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F132
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1550
    line_end: 1552
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F133
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1560
    line_end: 1562
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F134
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1570
    line_end: 1572
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F135
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1580
    line_end: 1582
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F136
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1590
    line_end: 1592
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F137
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1600
    line_end: 1602
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F138
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1610
    line_end: 1612
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F139
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1620
    line_end: 1622
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F140
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1630
    line_end: 1632
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F141
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1640
    line_end: 1642
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F142
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1650
    line_end: 1652
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F143
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1660
    line_end: 1662
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F144
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1670
    line_end: 1672
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F145
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1680
    line_end: 1682
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F146
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1690
    line_end: 1692
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F147
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1700
    line_end: 1702
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F148
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1710
    line_end: 1712
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F149
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1720
    line_end: 1722
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F150
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1730
    line_end: 1732
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F151
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1740
    line_end: 1742
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F152
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1750
    line_end: 1752
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F153
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1760
    line_end: 1762
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F154
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1770
    line_end: 1772
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F155
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1780
    line_end: 1782
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F156
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1790
    line_end: 1792
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F157
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1800
    line_end: 1802
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F158
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1810
    line_end: 1812
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F159
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1820
    line_end: 1822
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F160
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1830
    line_end: 1832
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F161
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1840
    line_end: 1842
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F162
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1850
    line_end: 1852
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F163
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1860
    line_end: 1862
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F164
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1870
    line_end: 1872
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F165
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1880
    line_end: 1882
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F166
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1890
    line_end: 1892
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F167
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1900
    line_end: 1902
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F168
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1910
    line_end: 1912
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F169
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1920
    line_end: 1922
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F170
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1930
    line_end: 1932
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F171
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1940
    line_end: 1942
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F172
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1950
    line_end: 1952
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F173
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1960
    line_end: 1962
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F174
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1970
    line_end: 1972
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F175
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1980
    line_end: 1982
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F176
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 1990
    line_end: 1992
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F177
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 2000
    line_end: 2002
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F178
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 2010
    line_end: 2012
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F179
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 2020
    line_end: 2022
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F180
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 2030
    line_end: 2032
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F181
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 2040
    line_end: 2042
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F182
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 2050
    line_end: 2052
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F183
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 2060
    line_end: 2062
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F184
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 2070
    line_end: 2072
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F185
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 2080
    line_end: 2082
    message: >
      The `applyMorphsetField` function has repeated pattern of handling string assignments.
    suggested_fix: >
      Create a helper function for handling string assignments.
  - id: F186
    severity: minor
    category: maintainability
    file: internal/act/polymorph.go
    line: 2090
    line_end: 2092
    message: >
      The
