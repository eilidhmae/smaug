# Go Idiom Review

Audit of the SMAUG Go port codebase for Go idiom violations, maintainability issues, and potential bugs.

## Critical Issues

### 1. Package-Level Mutable Variables Lack Documentation

These exported globals are set during init but undocumented for thread-safety:

| Variable | File:Line | Purpose |
|----------|-----------|---------|
| `act.WorldRef` | act/info.go:16 | Game world reference for commands |
| `act.SaveFunc` | act/info.go:325 | Player save callback |
| `act.CmdRegistry` | act/wiz.go:16 | Command registry for force/at |
| `act.StartEditingFunc` | act/olc.go:17 | String editor callback |
| `mudprog.CmdRegistry` | mudprog/driver.go:12 | Command registry for mp commands |
| `mudprog.WorldRef` | mudprog/commands.go:12 | World ref for mp commands |

**Fix**: Add godoc comments documenting: when initialized, thread-safety guarantees, required synchronization.

**Note**: These are only written once at boot before the game loop starts, and only read from the game loop goroutine. They are safe in practice due to the single-threaded game state model, but should be documented.

### 2. Unnecessary int Casts on Constants

Multiple instances of `int(types.CON_PLAYING)` and similar casts where the constants are already untyped `int`:

- game/loop.go:169, 260, 264, 307, 338, 354, 360, 554

**Fix**: Remove redundant `int()` casts.

### 3. Missing Nil Guards in MUD Programs

- mudprog/commands.go:22 — `mpEcho` accesses `mob.InRoom.People` after checking but before guard
- mudprog/commands.go:43 — `mpEchoAround` similar pattern
- mudprog/driver.go:160 — `mpPurge` accesses `mob.InRoom.People` without nil check

**Fix**: Add nil checks for `mob.InRoom` before accessing `People`.

## Medium Issues

### 4. Error Handling: Swallowed Errors

| File:Line | Issue |
|-----------|-------|
| persist/scanner.go:299 | `unreadByte()` swallows error |
| persist/player.go:92, 272-276 | Blank assignment discards read errors |
| persist/subsystems.go:22, 122 | Loader callbacks use log.Printf instead of returning errors |
| persist/subsystems.go:238, 293 | Bare error returns without context wrapping |
| game/loop.go:258, 270, 305 | `d.Conn.Write(telnetEchoOff/On)` errors discarded |
| act/olc.go:353, 391, 428 | `strconv.Atoi` errors silently ignored |

**Note**: Many of these follow the CLAUDE.md convention: "Error handling in file loaders: log with util.Bug() and continue." The telnet write errors are acceptable (connection may be closing). The OLC strconv errors default to 0 which is fine.

### 5. Descriptor Mutex Inconsistency

types/descriptor.go:46-47 — `outMu` protects `outBuf` and `pageBuf`, but `pagePoint`, `pageCmd`, `pageColor` are accessed via separate getter/setter methods. All pager fields should be consistently protected.

**Note**: In practice, pager fields are only accessed from the game loop goroutine, so this is safe. The mutex exists for the output buffer which is written from the game loop but flushed from the network goroutine.

### 6. Dead Code

- combat/combat.go:375-378 — `init()` function only verifies `types.AFF_SANCTUARY` exists. Can be removed.
- magic/magic.go:445-448 — Unreachable `if found { break }` in `SpellLocateObject` (always breaks on first iteration).
- mudprog/commands.go:155-170 — In `mpPurge`, logic after empty arg check is unreachable.

## Low Issues

### 7. Missing Godoc Comments on Exported Symbols

| Symbol | File |
|--------|------|
| `StrApp`, `IntApp`, etc. (7 tables) | types/attributes.go |
| `ObjIndexLookup` | persist/player.go:13 |
| `PlayerFilePath` | persist/player.go:593 |
| `MSSPInfo` | net/mssp.go:14 |
| `MSDPVariable` | net/msdp.go:16 |
| `FindSpellFunc` | magic/magic.go:52 |
| `Driver` | mudprog/driver.go:20 |
| `DoIfCheck` | mudprog/ifcheck.go:14 |
| `Translate` | mudprog/translate.go:15 |
| `MobTrigger` | mudprog/triggers.go:12 |

### 8. C-Style Naming Holdovers

| Current | Suggested | File |
|---------|-----------|------|
| `LowRVnum`/`HiRVnum` | `LowRoomVNum`/`HiRoomVNum` | types/area.go:21-26 |
| `LowOVnum`/`HiOVnum` | `LowObjVNum`/`HiObjVNum` | types/area.go:21-26 |
| `LowMVnum`/`HiMVnum` | `LowMobVNum`/`HiMobVNum` | types/area.go:21-26 |
| `BashPlrVsPlr` etc. | `BashPlayerVsPlayer` | types/system.go:46-58 |

**Note**: Per CLAUDE.md convention, "constant names match C names for cross-reference." Struct field names use PascalCase but abbreviations follow C patterns. Changing these would break cross-referencing with the C source — recommend leaving as-is until the port is complete.

### 9. Linear Search in World Collections

- world/world.go:110-116 — `RemoveChar` and `RemoveObj` use O(n) linear scan + copy
- handler/find.go:36-84 — `GetCharWorld`/`GetObjWorld` iterate twice (exact then prefix)

**Note**: Acceptable for current MUD scale (~500 entities). Would matter at 10,000+.

### 10. Close Errors Silently Discarded

- net/server.go:45, 74 — `listener.Close()` and `conn.Close()` errors ignored

**Note**: Standard Go practice for network cleanup. Logging might be useful for debugging but isn't critical.

## Not Issues (Confirmed Good Patterns)

- **strings.Builder** used correctly in net/color.go, game/editor.go, game/prompt.go, act/wiz.go
- **io.Reader/io.Writer** interfaces used properly in persist/player.go, persist/area_write.go
- **defer** patterns correct throughout
- **Channel ownership** in net/server.go correctly closes InputQueue via defer in readLoop
- **select + done channel** shutdown pattern in server.go is idiomatic
- **crypto/rand seeded RNG** in util/dice.go is correct
- **Untyped int constants** per CLAUDE.md convention for C interop
