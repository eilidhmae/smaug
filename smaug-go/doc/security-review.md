# Security Review

Security audit of the SMAUG Go port. Findings categorized by severity.

## Critical

### S1: Plaintext Password Storage

- **File**: game/loop.go:280,333 + persist/player.go:474
- **Issue**: Passwords stored and compared in plaintext (`NOCRYPT` mode from C). Player files contain raw passwords.
- **Risk**: File system compromise exposes all passwords.
- **Fix**: Replace with `bcrypt.GenerateFromPassword()`/`bcrypt.CompareHashAndPassword()`. Add migration path: if stored password doesn't start with `$2a$`, treat as legacy plaintext and re-hash on successful login.

### S2: No Connection Limit

- **File**: game/loop.go:142, net/server.go:52-77
- **Issue**: No maximum connection limit. Each connection spawns a goroutine and allocates a descriptor.
- **Risk**: Memory/goroutine exhaustion via connection flood.
- **Fix**: Add `const MaxConnections = 256`. Check `len(g.world.Descriptors) >= MaxConnections` before accepting; send "Server full" and close.

### S3: Force Command Privilege Bypass

- **File**: act/wiz.go:432-471
- **Issue**: `DoForce` calls `CmdRegistry.Interpret(victim, rest)` which executes with the victim's trust level. An immortal can force a player to run commands the immortal themselves can't access if the victim has higher trust.
- **Risk**: Privilege escalation.
- **Fix**: In `DoForce`, verify `ch.GetTrust() > victim.GetTrust()` (already done) AND cap the forced command's trust to `ch.GetTrust()` rather than `victim.GetTrust()`.

## High

### S4: No Input Length Limit

- **File**: net/server.go:82-96
- **Issue**: `bufio.Scanner` allows up to 64KB per line. No explicit limit on command length.
- **Risk**: Memory exhaustion via long input lines.
- **Fix**: `scanner.Buffer(make([]byte, 1024), 1024)` to limit input to 1KB.

### S5: No Password Brute Force Protection

- **File**: game/loop.go:268-286
- **Issue**: Unlimited password attempts with no rate limiting or lockout.
- **Risk**: Brute force password cracking.
- **Fix**: Track failed attempts per IP. Disconnect after 3 failures with increasing delay. Add `FailedAttempts int` to descriptor.

### S6: Unbounded Output Buffers

- **File**: types/descriptor.go:46,50
- **Issue**: `outBuf` and `pageBuf` grow without limit.
- **Risk**: Memory exhaustion if a client stops reading but commands keep generating output.
- **Fix**: Add `const MaxOutputBuf = 1 << 20` (1MB). Disconnect if exceeded.

### S7: Player Name Path Traversal

- **File**: persist/player.go:593-599
- **Issue**: `PlayerFilePath` constructs path from player name: `dataDir/player/first/name`. While `isValidName` at login requires letters-only (3-12 chars), names loaded from disk are not re-validated.
- **Risk**: A crafted player file with a dangerous name could cause path traversal on re-save.
- **Fix**: Add defense-in-depth validation in `PlayerFilePath()`: `name = filepath.Base(name)` and verify `[a-zA-Z]{3,12}` regex.

### S8: Area.lst Path Traversal

- **File**: persist/area.go:32
- **Issue**: Area filenames from area.lst used in `filepath.Join()` for loading. However, `filepath.Base()` is already applied at line 73 when storing the filename on the area struct, so saved area filenames are safe.
- **Risk**: Reduced — a malicious area.lst could cause reads from outside the area directory during loading, but stored filenames are sanitized. Defense-in-depth would validate at load time too.
- **Fix**: Apply `filepath.Base()` at line 32 as well, before the initial file read. Validate extension is `.are`.

## Medium

### S9: Descriptor Access Race Condition

- **File**: game/loop.go:652-675
- **Issue**: `flushOutput()` iterates descriptors and calls `FlushOutput()` while network goroutines may be writing to descriptor output buffers concurrently.
- **Note**: The output buffer IS protected by `outMu` mutex in types/descriptor.go. The real race would be on the Descriptors slice itself — but it's only modified in the game loop goroutine (acceptNewConnections, cleanupDescriptors), so this is safe. The mutex on outBuf protects the actual hot path.
- **Severity**: Medium (theoretical, mitigated by design)

### S10: No Atomic Player File Writes

- **File**: game/loop.go:625
- **Issue**: `os.Create()` truncates immediately. If server crashes mid-save, player file is lost/corrupted.
- **Risk**: Player data loss.
- **Fix**: Write to `path.tmp`, then `os.Rename(path.tmp, path)` for atomic replacement.

### S11: stripTelnetIAC Bounds Check

- **File**: net/server.go:120-125
- **Issue**: `i += 3` for WILL/WONT/DO/DONT without verifying 3rd byte exists. If input is `[IAC, WILL]` (no option byte), skips valid data.
- **Fix**: Check `i+2 < len(data)` before consuming 3 bytes.

### S12: Scanner ReadString Unbounded

- **File**: persist/scanner.go:80-120
- **Issue**: `ReadString()` reads until tilde with no maximum length. Malformed file could exhaust memory.
- **Fix**: Add `const maxStringLen = 32768` limit.

### S13: Trust Field Not Capped on Load

- **File**: persist/player.go (LoadPlayer)
- **Issue**: Trust field loaded from player file without range validation.
- **Risk**: Crafted player file could set Trust to any value.
- **Fix**: Cap `ch.Trust` to `LEVEL_IMPL` after loading.

## Low

### S14: Timing Attack on Password Comparison

- **File**: game/loop.go:280
- **Issue**: `line != d.Character.PCData.Pwd` is not constant-time.
- **Risk**: Theoretical timing side-channel (minimal practical risk over network).
- **Fix**: Use `subtle.ConstantTimeCompare()` (moot if S1 bcrypt fix is applied, since bcrypt.CompareHashAndPassword is constant-time).

### S15: OLC Commands Rely on Registry Level Only

- **File**: act/olc.go
- **Issue**: OLC commands don't have internal level checks; rely on command registry `Level: LEVEL_IMMORTAL`.
- **Risk**: Low — only reachable via the registry which does check. But defense-in-depth would add guards.
- **Fix**: Add `if ch.GetTrust() < types.LEVEL_IMMORTAL { return }` at function start.

### S16: Telnet Echo-Off Not Guaranteed

- **File**: game/loop.go:258,305
- **Issue**: Sends `IAC WILL ECHO` to suppress password echo, but client may ignore it.
- **Risk**: Passwords visible on screen with non-compliant clients.
- **Fix**: Document limitation. No server-side fix possible for client behavior.

## Summary

| Severity | Count | Key Items |
|----------|-------|-----------|
| Critical | 3 | Plaintext passwords, no connection limit, force bypass |
| High | 5 | Input length, brute force, output buffers, path traversal (x2) |
| Medium | 5 | Race condition, atomic writes, telnet bounds, scanner bounds, trust cap |
| Low | 3 | Timing attack, OLC defense-in-depth, telnet echo |

## Resolution Status

All actionable findings have been addressed:

| ID | Status | Fix |
|----|--------|-----|
| S1 | **FIXED** | bcrypt hashing with legacy plaintext migration on login |
| S2 | **FIXED** | `MaxConnections = 256` check in `acceptNewConnections` |
| S3 | **FIXED** | `InterpretWithTrustCap` caps forced command trust to forcer's level |
| S4 | **FIXED** | `scanner.Buffer(1024, 1024)` limits input to 1KB |
| S5 | **FIXED** | `FailedAttempts` counter, disconnect after 3 failures |
| S6 | **FIXED** | `MaxOutputBuf = 1MB`, `OutputOverflow` flag clears buffer |
| S7 | **FIXED** | `filepath.Base()` + `^[a-zA-Z]{3,12}$` regex in `PlayerFilePath` |
| S8 | **FIXED** | `filepath.Base()` + `.are` suffix check at load time |
| S9 | N/A | Safe by design (documented, no code change needed) |
| S10 | **FIXED** | Write to `.tmp` then `os.Rename` for atomic replacement |
| S11 | **FIXED** | Bounds check `i+2 < len(data)` before consuming 3-byte sequence |
| S12 | **FIXED** | `maxStringLen = 32768` cap in `ReadString` |
| S13 | **FIXED** | Trust capped to `LEVEL_SUPREME` on player load |
| S14 | **FIXED** | Moot — bcrypt (S1) uses constant-time comparison; legacy path uses `subtle.ConstantTimeCompare` |
| S15 | **FIXED** | Trust guard `ch.GetTrust() < LEVEL_IMMORTAL` at top of all 8 OLC functions |
| S16 | N/A | Client-side limitation, documented (no server fix possible) |
