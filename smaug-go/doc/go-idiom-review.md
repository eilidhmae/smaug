# Go Idiom Review

Audit of the SMAUG Go port codebase for Go idiom violations, maintainability issues, and potential bugs.

**Status: All actionable findings resolved.** See resolution notes on each item below.

## Critical Issues

### 1. Package-Level Mutable Variables Lack Documentation — RESOLVED

These exported globals are set during init but were undocumented for thread-safety:

| Variable | File:Line | Purpose |
|----------|-----------|---------|
| `act.WorldRef` | act/info.go:16 | Game world reference for commands |
| `act.SaveFunc` | act/info.go:325 | Player save callback |
| `act.CmdRegistry` | act/wiz.go:16 | Command registry for force/at |
| `act.StartEditingFunc` | act/olc.go:17 | String editor callback |
| `mudprog.CmdRegistry` | mudprog/driver.go:12 | Command registry for mp commands |
| `mudprog.WorldRef` | mudprog/commands.go:12 | World ref for mp commands |

**Resolution**: Added godoc documenting thread-safety: written once at boot, read only from game loop goroutine, safe without synchronization.

### 2. Unnecessary int Casts on Constants — RESOLVED

Multiple instances of `int(types.CON_PLAYING)` and similar casts where the constants are already untyped `int`.

**Resolution**: Removed ~50 redundant `int()` casts across 7 files (loop.go, loop_test.go, info.go, info_test.go, magic_test.go, driver_test.go, commands_test.go).

### 3. Logic Bug in mpPurge — RESOLVED

mudprog/commands.go — `mpPurge` had an inverted condition: `if arg == "" || mob.InRoom == nil || WorldRef == nil` entered the purge-all block even when `mob.InRoom` was nil, causing a nil-pointer panic.

**Resolution**: Restructured into two guards — nil check returns early, empty-arg check enters purge-all path. Added 3 new tests (nil room + empty args, nil room + target, nil WorldRef).

## Medium Issues

### 4. Error Handling: Swallowed Errors — NO ACTION (by design)

| File:Line | Issue |
|-----------|-------|
| persist/scanner.go:299 | `unreadByte()` swallows error |
| persist/player.go:92, 272-276 | Blank assignment discards read errors |
| persist/subsystems.go:22, 122 | Loader callbacks use log.Printf instead of returning errors |
| persist/subsystems.go:238, 293 | Bare error returns without context wrapping |
| game/loop.go:258, 270, 305 | `d.Conn.Write(telnetEchoOff/On)` errors discarded |
| act/olc.go:353, 391, 428 | `strconv.Atoi` errors silently ignored |

**Note**: These follow the CLAUDE.md convention: "Error handling in file loaders: log with util.Bug() and continue." Telnet write errors are acceptable (connection may be closing). OLC strconv errors default to 0 which is fine.

### 5. Descriptor Mutex Inconsistency — RESOLVED

types/descriptor.go — `GetPagerCmd()` and `SetPagerCmd()` were not mutex-protected while all other pager methods used `outMu`.

**Resolution**: Added `outMu.Lock()/Unlock()` to both methods for consistency.

### 6. Dead Code — RESOLVED

- combat/combat.go — `init()` that only verified `types.AFF_SANCTUARY` exists. Removed.
- magic/magic.go — Unreachable `if found { break }` in `SpellLocateObject`. Simplified to unconditional `break`.
- mudprog/commands.go — mpPurge dead code path resolved by issue #3 fix above.

## Low Issues

### 7. Missing Godoc Comments on Exported Symbols — RESOLVED

**Resolution**: Added godoc to the 7 attribute table exports in types/attributes.go (`StrApp`, `IntApp`, `WisApp`, `DexApp`, `ConApp`, `ChaApp`, `LckApp`). The remaining symbols (`PlayerFilePath`, `MSDPVariable`, `FindSpellFunc`, `Driver`, `DoIfCheck`, `Translate`, `MobTrigger`) already had godoc comments.

### 8. C-Style Naming Holdovers — NO ACTION (by design)

| Current | Suggested | File |
|---------|-----------|------|
| `LowRVnum`/`HiRVnum` | `LowRoomVNum`/`HiRoomVNum` | types/area.go:21-26 |
| `LowOVnum`/`HiOVnum` | `LowObjVNum`/`HiObjVNum` | types/area.go:21-26 |
| `LowMVnum`/`HiMVnum` | `LowMobVNum`/`HiMobVNum` | types/area.go:21-26 |
| `BashPlrVsPlr` etc. | `BashPlayerVsPlayer` | types/system.go:46-58 |

**Note**: Per CLAUDE.md convention, "constant names match C names for cross-reference." Leaving as-is until the port is complete.

### 9. Linear Search in World Collections — NO ACTION (acceptable)

- world/world.go:110-116 — `RemoveChar` and `RemoveObj` use O(n) linear scan + copy
- handler/find.go:36-84 — `GetCharWorld`/`GetObjWorld` iterate twice (exact then prefix)

**Note**: Acceptable for current MUD scale (~500 entities). Would matter at 10,000+.

### 10. Close Errors Silently Discarded — NO ACTION (standard practice)

- net/server.go:45, 74 — `listener.Close()` and `conn.Close()` errors ignored

**Note**: Standard Go practice for network cleanup.

## Not Issues (Confirmed Good Patterns)

- **strings.Builder** used correctly in net/color.go, game/editor.go, game/prompt.go, act/wiz.go
- **io.Reader/io.Writer** interfaces used properly in persist/player.go, persist/area_write.go
- **defer** patterns correct throughout
- **Channel ownership** in net/server.go correctly closes InputQueue via defer in readLoop
- **select + done channel** shutdown pattern in server.go is idiomatic
- **crypto/rand seeded RNG** in util/dice.go is correct
- **Untyped int constants** per CLAUDE.md convention for C interop
