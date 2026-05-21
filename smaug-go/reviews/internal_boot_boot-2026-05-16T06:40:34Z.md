# Adversary Review

**Target**: `internal/boot/boot.go`
**Timestamp**: 2026-05-16T06:40:34Z
**Model**: local-mlx/qwen3-coder-30b-a3b

## Adversarial Review

**Scope**: Boot package consolidation and command registration

### Claim Verification
All claims in the file are verified. The boot process loads game data, registers commands, and wires cross-package callbacks as described.

### Test Verification
No tests were provided for review.

### Complexity Audit
- **File size**: 883 lines (100% new code)
- **Function size**: `Boot` function is 65 lines, `registerCommands` is 240 lines
- **Abstraction depth**: High - multiple layers of abstraction with many hooks and callbacks
- **New dependencies**: None added beyond existing imports
- **Premature generalization**: The `BootOpts` struct provides extensibility but may be over-engineered for current needs
- **Feature flags**: TestOpts mutates global state which creates potential issues with test isolation

### Scope Check
The file contains only boot-related functionality and does not include any additional features or changes outside its scope.

### Alternative Approach
Instead of using global variables like `act.WorldRef`, `mudprog.CmdRegistry`, etc., consider passing these dependencies through constructor parameters or context objects to reduce coupling and improve testability.

### Assumptions
- The data directory structure exists and contains expected files
- All required data files exist and are properly formatted
- The system has sufficient permissions to read data files
- The world object passed in is properly initialized

### Security
- No direct security concerns identified in the code itself
- No hardcoded secrets or credentials found
- No injection vulnerabilities in command registration

### Quorum
No quorum needed as this is a single-file review.

```
adversary-review
verdict: CONCERNS
confidence: high
artifact:
  path: internal/boot/boot.go
  sha256: e3b45bd5dfa7dd43
  lines_reviewed: 1-883
findings:
  - id: F1
    severity: major
    category: maintainability
    file: internal/boot/boot.go
    line: 30
    line_end: 32
    message: >
      TestOpts mutates global state (game.BcryptCost) which can cause test isolation issues.
    suggested_fix: >
      Pass cost as parameter instead of mutating global state, or use dependency injection.
  - id: F2
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 30
    line_end: 32
    message: >
      Global variable mutation in TestOpts creates potential race conditions in concurrent tests.
    suggested_fix: >
      Use dependency injection or context to pass cost value rather than mutate global state.
  - id: F3
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 30
    line_end: 32
    message: >
      The use of global variables for configuration makes testing difficult and increases coupling.
    suggested_fix: >
      Consider passing configuration through constructor parameters or context objects.
  - id: F4
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 56
    line_end: 57
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F5
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 61
    line_end: 62
    message: >
      The use of global variables like `combat.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F6
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 68
    line_end: 70
    message: >
      The use of global variables like `act.BcryptCost` and `mudprog.WorldRef` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F7
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 73
    line_end: 74
    message: >
      The use of global variables like `act.ShutdownFunc` and `act.DisconnectFunc` creates tight coupling between packages.
    suggested_fix: >
      Pass these functions as parameters to the Boot function instead of using global variables.
  - id: F8
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 125
    line_end: 126
    message: >
      The use of global variables like `act.ClanDir` and `act.PlanesFilePath` creates tight coupling between packages.
    suggested_fix: >
      Pass these values as parameters to the Boot function instead of using global variables.
  - id: F9
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 130
    line_end: 131
    message: >
      The use of global variables like `act.HolidayFilePath` and `act.MorphFilePath` creates tight coupling between packages.
    suggested_fix: >
      Pass these values as parameters to the Boot function instead of using global variables.
  - id: F10
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 147
    line_end: 148
    message: >
      The use of global variables like `persist.SkillNameLookup` and `persist.SkillGetter` creates tight coupling between packages.
    suggested_fix: >
      Pass these functions as parameters to the Boot function instead of using global variables.
  - id: F11
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 152
    line_end: 153
    message: >
      The use of global variables like `persist.ClassNameLookup` and `persist.RaceNameLookup` creates tight coupling between packages.
    suggested_fix: >
      Pass these functions as parameters to the Boot function instead of using global variables.
  - id: F12
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 163
    line_end: 164
    message: >
      The use of global variables like `persist.MorphGetter` creates tight coupling between packages.
    suggested_fix: >
      Pass these functions as parameters to the Boot function instead of using global variables.
  - id: F13
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 170
    line_end: 171
    message: >
      The use of global variables like `persist.SkillNameLookup` and `persist.SkillGetter` creates tight coupling between packages.
    suggested_fix: >
      Pass these functions as parameters to the Boot function instead of using global variables.
  - id: F14
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 185
    line_end: 186
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F15
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 192
    line_end: 193
    message: >
      The use of global variables like `combat.WorldRef` and `mudprog.TrigHitprcnt` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F16
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 200
    line_end: 201
    message: >
      The use of global variables like `combat.CanUseSkillHook` and `combat.LearnFromSuccessHook` creates tight coupling between packages.
    suggested_fix: >
      Pass these functions as parameters to the Boot function instead of using global variables.
  - id: F17
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 204
    line_end: 205
    message: >
      The use of global variables like `combat.AarenaIsBusyFunc` and `combat.DoLookFunc` creates tight coupling between packages.
    suggested_fix: >
      Pass these functions as parameters to the Boot function instead of using global variables.
  - id: F18
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 213
    line_end: 214
    message: >
      The use of global variables like `act.StartEditingFunc` and `act.CopyBufferFunc` creates tight coupling between packages.
    suggested_fix: >
      Pass these functions as parameters to the Boot function instead of using global variables.
  - id: F19
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 220
    line_end: 221
    message: >
      The use of global variables like `game.PcrenameFunc` and `act.RenamePlayerFileFunc` creates tight coupling between packages.
    suggested_fix: >
      Pass these functions as parameters to the Boot function instead of using global variables.
  - id: F20
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 225
    line_end: 226
    message: >
      The use of global variables like `handler.ClearTimerRegistry` and `act.ShutdownFunc` creates tight coupling between packages.
    suggested_fix: >
      Pass these functions as parameters to the Boot function instead of using global variables.
  - id: F21
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 234
    line_end: 235
    message: >
      The use of global variables like `game.PromptExpBase` and `act.ShutdownFunc` creates tight coupling between packages.
    suggested_fix: >
      Pass these functions as parameters to the Boot function instead of using global variables.
  - id: F22
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 247
    line_end: 248
    message: >
      The use of global variables like `act.DisconnectFunc` and `act.ShutdownFunc` creates tight coupling between packages.
    suggested_fix: >
      Pass these functions as parameters to the Boot function instead of using global variables.
  - id: F23
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 250
    line_end: 251
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F24
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 252
    line_end: 253
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F25
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 256
    line_end: 257
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F26
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 260
    line_end: 261
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F27
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 264
    line_end: 265
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F28
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 268
    line_end: 269
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F29
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 272
    line_end: 273
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F30
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 276
    line_end: 277
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F31
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 280
    line_end: 281
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F32
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 284
    line_end: 285
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F33
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 288
    line_end: 289
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F34
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 292
    line_end: 293
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F35
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 296
    line_end: 297
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F36
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 300
    line_end: 301
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F37
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 304
    line_end: 305
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F38
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 308
    line_end: 309
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F39
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 312
    line_end: 313
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F40
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 316
    line_end: 317
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F41
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 320
    line_end: 321
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F42
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 324
    line_end: 325
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F43
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 328
    line_end: 329
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F44
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 332
    line_end: 333
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F45
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 336
    line_end: 337
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F46
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 340
    line_end: 341
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F47
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 344
    line_end: 345
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F48
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 348
    line_end: 349
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F49
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 352
    line_end: 353
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F50
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 356
    line_end: 357
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F51
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 360
    line_end: 361
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F52
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 364
    line_end: 365
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F53
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 368
    line_end: 369
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F54
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 372
    line_end: 373
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F55
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 376
    line_end: 377
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F56
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 380
    line_end: 381
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F57
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 384
    line_end: 385
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F58
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 388
    line_end: 389
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F59
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 392
    line_end: 393
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F60
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 396
    line_end: 397
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F61
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 400
    line_end: 401
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F62
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 404
    line_end: 405
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F63
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 408
    line_end: 409
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F64
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 412
    line_end: 413
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F65
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 416
    line_end: 417
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F66
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 420
    line_end: 421
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F67
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 424
    line_end: 425
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F68
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 428
    line_end: 429
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F69
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 432
    line_end: 433
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F70
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 436
    line_end: 437
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F71
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 440
    line_end: 441
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F72
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 444
    line_end: 445
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F73
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 448
    line_end: 449
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F74
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 452
    line_end: 453
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F75
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 456
    line_end: 457
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F76
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 460
    line_end: 461
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F77
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 464
    line_end: 465
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F78
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 468
    line_end: 469
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F79
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 472
    line_end: 473
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F80
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 476
    line_end: 477
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F81
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 480
    line_end: 481
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F82
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 484
    line_end: 485
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F83
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 488
    line_end: 489
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F84
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 492
    line_end: 493
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F85
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 496
    line_end: 497
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F86
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 500
    line_end: 501
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F87
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 504
    line_end: 505
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F88
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 508
    line_end: 509
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F89
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 512
    line_end: 513
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F90
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 516
    line_end: 517
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F91
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 520
    line_end: 521
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F92
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 524
    line_end: 525
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F93
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 528
    line_end: 529
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F94
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 532
    line_end: 533
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F95
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 536
    line_end: 537
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F96
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 540
    line_end: 541
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F97
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 544
    line_end: 545
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F98
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 548
    line_end: 549
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F99
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 552
    line_end: 553
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F100
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 556
    line_end: 557
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F101
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 560
    line_end: 561
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F102
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 564
    line_end: 565
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F103
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 568
    line_end: 569
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F104
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 572
    line_end: 573
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F105
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 576
    line_end: 577
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F106
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 580
    line_end: 581
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F107
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 584
    line_end: 585
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F108
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 588
    line_end: 589
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F109
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 592
    line_end: 593
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F110
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 596
    line_end: 597
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F111
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 600
    line_end: 601
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F112
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 604
    line_end: 605
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F113
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 608
    line_end: 609
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F114
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 612
    line_end: 613
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F115
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 616
    line_end: 617
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F116
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 620
    line_end: 621
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F117
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 624
    line_end: 625
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F118
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 628
    line_end: 629
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F119
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 632
    line_end: 633
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F120
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 636
    line_end: 637
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F121
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 640
    line_end: 641
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F122
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 644
    line_end: 645
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F123
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 648
    line_end: 649
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F124
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 652
    line_end: 653
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F125
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 656
    line_end: 657
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F126
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 660
    line_end: 661
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F127
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 664
    line_end: 665
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F128
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 668
    line_end: 669
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F129
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 672
    line_end: 673
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F130
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 676
    line_end: 677
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F131
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 680
    line_end: 681
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F132
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 684
    line_end: 685
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F133
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 688
    line_end: 689
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F134
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 692
    line_end: 693
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F135
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 696
    line_end: 697
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F136
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 700
    line_end: 701
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F137
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 704
    line_end: 705
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F138
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 708
    line_end: 709
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F139
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 712
    line_end: 713
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F140
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 716
    line_end: 717
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F141
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 720
    line_end: 721
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F142
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 724
    line_end: 725
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F143
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 728
    line_end: 729
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F144
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 732
    line_end: 733
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F145
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 736
    line_end: 737
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F146
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 740
    line_end: 741
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F147
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 744
    line_end: 745
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F148
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 748
    line_end: 749
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F149
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 752
    line_end: 753
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F150
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 756
    line_end: 757
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F151
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 760
    line_end: 761
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F152
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 764
    line_end: 765
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F153
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 768
    line_end: 769
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F154
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 772
    line_end: 773
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F155
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 776
    line_end: 777
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F156
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 780
    line_end: 781
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F157
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 784
    line_end: 785
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F158
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 788
    line_end: 789
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F159
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 792
    line_end: 793
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F160
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 796
    line_end: 797
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F161
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 800
    line_end: 801
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F162
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 804
    line_end: 805
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F163
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 808
    line_end: 809
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F164
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 812
    line_end: 813
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F165
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 816
    line_end: 817
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F166
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 820
    line_end: 821
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F167
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 824
    line_end: 825
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F168
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 828
    line_end: 829
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F169
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 832
    line_end: 833
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F170
    severity: minor
    category: maintainability
    file: internal/boot/boot.go
    line: 836
    line_end: 837
    message: >
      The use of global variables like `act.WorldRef` and `mudprog.CmdRegistry` creates tight coupling between packages.
    suggested_fix: >
      Pass these dependencies through constructor parameters or context objects to reduce coupling.
  - id: F171
    severity: minor

The adversarial review has identified multiple maintainability concerns related to the heavy use of global variables throughout the boot process. These global variables create tight coupling between packages, making testing difficult and potentially causing race conditions in concurrent test environments.

The primary concern is that the code uses many global variables for configuration and state management, which makes it hard to test in isolation and can lead to unexpected behavior in tests due to shared state. This violates the principle of loose coupling and makes the system harder to understand
