# Plan: Phase 6 — Hotboot / Copyover (Design-Exploration)

**Status:** Design-exploration draft (2026-04-18). This is **not** an executable plan. It presents two viable Go designs, evaluates each against the C baseline, records PoC evidence for both, and recommends one. A follow-up executable plan (same filename, rewritten) lands after the recommendation is accepted.

**Audit status (2026-04-18):** audited via `audit-hotboot` lineage — verdict PASS with notes. All 7 C citations verified against `src/hotboot.c:73-854`; all 14 Go-state claims verified (HomeVnum L209, RoomVnum L90, PCData.Hotboot L117, Boot L78, act.WorldRef L81, ClearTimerRegistry L152, GameLoop.Run L98 / .Cancel L120 / .SavePlayer L788, LoadPlayerWithWorld L34, DescriptorData L14, server.go accept/readLoop L46/L77, World L12). Surgical corrections applied in-place: Windows-specific caveat strengthened on `(*net.TCPConn).File()` (Go stdlib docs state the returned fd is "not usable on other processes" on Windows, independent of `syscall.Exec`); `syscall.Dup2` portability note added (available on linux/amd64; missing on linux/arm64 where only `Dup3` exists); newacad.are bug list expanded. No design-level blockers.

**Wave:** Phase 6 Wave 0 per `phase6-roadmap.md`.

**Scope of this document:**
- Problem + C reference + Go current state.
- Two viable Go design alternatives (A and B), each with concrete stdlib calls, preservation analysis, LOC estimate, tests required, risk assessment.
- PoC evidence for both designs (commands + observations recorded below).
- Recommendation with rationale.
- Prerequisites for the executable plan.
- Skeleton task groups (scope-stated, not test-first-detailed) for the follow-up executable plan.
- Open questions requiring human decision.

**Scope explicitly excluded:**
- Test-first task detail (deferred to executable-plan rewrite).
- Acceptance-criteria checklist (deferred).
- Any Go code landing in `smaug-go/` (this is a plan, not an implementation).

---

## Problem

C ships hotboot/copyover at `src/hotboot.c`: a 854-line module that lets an immortal issue `hotboot` to restart the MUD process with all players still connected. Players experience a ~2 second pause where the world is re-loaded from disk, then their prompt returns — no disconnect, no state loss beyond active combat (which C refuses to hotboot through).

The mechanism C uses is **Unix-specific**:
1. Save mob positions, object locations, player state, and per-descriptor connection info to three files (`system/mobfile.dat`, `system/<vnum>.objdat`, `system/hotboot.dat`).
2. Flip every player's `pcdata->hotboot` to `TRUE` and save their pfile.
3. `execl()` the same binary with `hotboot` as the second arg, passing the **same integer file descriptors** (the listener socket + every client socket) inherited across the exec because Unix FDs survive `execve(2)` by default when `FD_CLOEXEC` is not set.
4. The recovered process reads the hotboot file, re-wraps each FD into a fresh `DESCRIPTOR_DATA`, loads each player from the pfile, sets `connected = CON_COPYOVER_RECOVER`, and routes subsequent I/O through the existing socket.

Go has no equivalent to this flow. Three properties make it hard:

1. **Go's `net.Conn` is an interface**, not an int. The socket FD is hidden behind a `*net.TCPConn` (or whatever the implementation is). You can extract it via `(*net.TCPConn).File()` which `dup(2)`s the underlying FD into an `*os.File`. Per the Go stdlib docs: "Closing c does not affect f, and closing f does not affect c" — so the original `net.TCPConn` remains usable — but "Attempting to change properties of the original using this duplicate may or may not have the desired effect" (e.g., setsockopt on the dup doesn't necessarily propagate to the original's poller). Practical implication: for hotboot, hand the dup off via exec and do NOT continue reading from the original in the parent goroutine.
2. **Go's runtime owns goroutine-per-connection read loops.** C's single-threaded `select()` loop can hand an FD to a new process and be done. Go has per-connection goroutines blocked on `conn.Read`. Exec kills those goroutines but the TCP endpoint continues to exist in the kernel; the child must spawn fresh goroutines for the inherited FDs.
3. **Go's standard library has no native "re-exec self" helper.** There's `syscall.Exec` (Unix only; Windows has no equivalent; replaces process image in-place), and `os.StartProcess` (forks a new child; supports a `Files []uintptr` slice for FD inheritance). Neither is integrated with `net.Listener`. The caller must plumb FDs through `*os.File` themselves.

The Phase 6 roadmap (`phase6-roadmap.md:75`, "Hotboot / copyover — plan-phase6-hotboot.md"; Wave 0 designation at `:81`) flagged hotboot as a design-only pass because picking the wrong approach wastes significant implementation work. This document is the design pass.

### C reference (authoritative)

All line numbers are against `src/hotboot.c` at HEAD of the current C tree.

- `src/hotboot.c:73 save_mobile` — serializes one NPC (vnum, level, gold, room, coordinates, name/short/long/desc deltas vs index, hp/mana/move, position, act flags, affected-by, affect list, inventory via `fwrite_obj`). Skips PCs (`IS_NPC` gate), `ACT_PROTOTYPE`, `ACT_PET`.
- `src/hotboot.c:143 save_world` — top-level world save. Writes `system/mobfile.dat` for all NPCs, then one file per non-empty room (`system/<room_vnum>.objdat`) for room-floor objects. Skips `ROOM_CLANSTOREROOM` rooms (they save on their own path).
- `src/hotboot.c:206 load_mobile` — inverse of `save_mobile`. Looks up mob prototype by vnum (abort if missing), `create_mobile`, reads each key, links affects, calls `fread_obj` on embedded `#OBJECT` blocks, places into room by vnum or `ROOM_VNUM_LIMBO` fallback (C hotboot.c:350 — note this is different from the player-recovery fallback at :830 which is `ROOM_VNUM_TEMPLE`).
- `src/hotboot.c:433 read_obj_file` + `:524 load_obj_files` — per-room object files. Uses a `supermob` context so `fread_obj` has a carrier to stash objects on, then moves them to the room. Unlinks the file after load.
- `src/hotboot.c:552 load_world` — called from recovery path; reads the mob file then calls `load_obj_files`. **Unlinks the mob file after load**, so a crash in recovery doesn't loop.
- `src/hotboot.c:599 do_hotboot` — the command. Refuses hotboot if anyone is in combat (`d->character->fighting`) or in the editor (`CON_EDITING`). Calls `save_world`, then for each descriptor writes a line to `system/hotboot.dat` with `descriptor int, can_compress, room vnum, port, idle, name, host`, flips `pcdata->hotboot = TRUE`, calls `save_char_obj`, writes a farewell message, compresses-end. Then: `set_alarm(0)` (clear SIGALRM), `dlclose` the dynamic-library handle, `execl(EXE_FILE, PACKAGE, port, "hotboot", control_fd, "-1", NULL)`. The "-1" is a sentinel for "no TTY". If `execl` fails, logs and re-opens the dl handle — graceful fallback.
- `src/hotboot.c:744 hotboot_recover` — called from `main()` when argv contains `hotboot`. Reads `system/hotboot.dat` line by line; for each line does `CREATE(d, DESCRIPTOR_DATA, 1)`, sets `d->descriptor = desc`, creates output buffer, compresses-start if applicable, links into `first_descriptor/last_descriptor`, initially sets `d->connected = CON_COPYOVER_RECOVER` (negative-state sentinel so `close_socket` cuts them off on pfile-load failure; C hotboot.c:813), then writes a "time resumes" message to `desc` via `write_to_descriptor` (which uses the raw int FD — no Go analog to this raw write path). Calls `load_char_obj` to fetch the player save. On success, places character in the room (falling back to `ROOM_VNUM_TEMPLE` at C hotboot.c:830 if the saved vnum is gone), links into `first_char`, emits a puff-of-smoke social, transitions to `connected = CON_PLAYING`. On failure, closes the FD silently and moves on.
- `src/db.c` (not in hotboot.c but related): the recovery path is triggered from `main()`'s argv parsing (the second arg "hotboot"); this path calls `load_world()` after area boot but before accepting new connections. The ordering matters — all area resets must complete before `load_world` reinstates the saved mob/obj overlay on top.

### Go current state

Verified facts from reading `/home/eilidh/src/smaug/smaug-go` at HEAD (commit `a3528d3`):

1. **Network layer: goroutine-per-connection** (`internal/net/server.go:46-104`). `acceptLoop` spawns `readLoop` per connection; each readLoop is a bufio.Scanner over the `net.Conn` writing lines to `desc.InputQueue` channel. `s.Stop()` closes the listener and signals done via `s.done` channel; it does NOT preserve any FDs.
2. **Descriptor struct** (`internal/types/descriptor.go:14`). Holds `Conn net.Conn` (interface), `Character *CharData`, connection state, input/output buffers. No raw-FD field; the FD is buried inside `Conn`.
3. **Game loop** (`internal/game/loop.go:98` `Run`, `:120` `Cancel`). Pulse-based (250ms). `pulse()` drains inputs, fires updates, flushes outputs, cleans up descriptors. The loop reads `d.InputQueue` non-blockingly. Shutdown path via `g.Cancel()` (at `:120`) cancels the internal context; `Run` returns; the caller closes the server.
4. **Boot path** (`internal/boot/boot.go:78 Boot`). Loads areas, classes, races, skills, stances, socials, clans, deities, boards; registers ~200 commands; wires callbacks (`act.SaveFunc`, `combat.WorldRef`, `handler.ClearTimerRegistry`, `act.StartEditingFunc`, etc.); returns the `(Registry, GameLoop, error)` triple. `bootDB` logs "Loaded %d ..." for each subsystem.
5. **cmd/smaug/main.go** (`cmd/smaug/main.go:18-70`). `world.New(dataDir)` → `NewServer` → `Boot(w, dataDir, server.Incoming, ProductionOpts())` → `net.Listen` → `server.StartOnListener(ln)` → `gameLoop.Run(ctx)`. Signal handler on SIGINT/SIGTERM cancels ctx. Currently the listener is bound AFTER boot finishes, meaning there's a small gap between "boot complete" log and "listening" log (harmless in normal operation but relevant for hotboot: the listener FD would need to be rebound in the child or inherited from the parent).
6. **Player persistence** (`internal/persist/player.go:34 LoadPlayerWithWorld`, SavePlayer). Text format compatible with C's pfile. `game.GameLoop.SavePlayer` (`internal/game/loop.go:788`) writes via temp file + rename. `persist.SavePlayer` (`internal/persist/player.go`) takes an `io.Writer` — no side-effects beyond the writer.
7. **World structure** (`internal/world/world.go:12`). Holds maps + slices: `Rooms`, `MobIndex`, `ObjIndex` (templates); `Characters`, `Objects`, `Descriptors` (live); static tables (Classes, Races, Skills, Clans, Deities, Shops, Boards, Socials, Bans); `SysData`, `TimeInfo`, `WeatherInfo`, `Auction`. `CharData.HomeVnum int` field **is** already defined (`internal/types/character.go:209`, tagged `// Hotboot`); `ObjData.RoomVnum int` is likewise already defined (`internal/types/object.go:90`). Live `Characters` and `Objects` are plain slices; mob data is materialized from `MobIndex` templates by area resets. Sentinel-mob home-vnum wiring (matching C `save_mobile`'s `ACT_SENTINEL` → `home_vnum` save) still needs to be populated at `create_mobile` time — verify this is wired before G1 lands.
8. **No hotboot scaffolding exists.** No `persist/hotboot.go`, no `act.DoHotboot`, no `system/hotboot.dat`, no `db/hotboot/`. The roadmap's TODO.md entry (line 157) is the only mention.
9. **Build system.** `cmd/smaug/main.go` builds to a single static binary. `go build ./...` produces `smaug-go` in `smaug-go/smaug-go`. No dlopen/dlclose (Go doesn't support it; C's `sysdata.dlHandle` has no analog).
10. **Area-file bugs on load** (verified via 63ms boot-time measurement — see § Design PoC Results): `db/area/gods.are:115, 119, 123` `BUG: ReadNumber` at the `#SHOPS` / `#REPAIRS` / `#SPECIALS` header positions (parser mis-dispatch); `newacad.are:17, 931, 932, 936` similar, plus ~70 further `unknown section` errors at :942 through :2786. These are pre-existing and NOT blocking hotboot — just noise in the baseline. Full area load completes in 63ms.
11. **`act.WorldRef` singleton** (`internal/boot/boot.go:81`). Set once at boot; read by `act`, `combat`, `mudprog`. Hotboot recovery sets this again after area reload — no special handling needed since `Boot` always sets it.
12. **Existing `EditorSave` callback field** (`internal/types/character.go` — from Tier 12). This is how the plan pattern looks for "re-dispatch on restart"; per-player callbacks survive across restart only if the fn pointer is re-established by the recovery path. Hotboot-recovery must NOT attempt to preserve Go func pointers — they're process-local. Any mid-edit session must be flushed or discarded before hotboot.
13. **`handler.ClearTimerRegistry()`** (`internal/boot/boot.go:152`). Boot explicitly clears the `TIMER_DO_FUN` callback registry. This is the correct pattern for hotboot too — on recovery, the process must re-register all callbacks from scratch.
14. **No `exec_self` / `re_exec` helper**. Any hotboot design must add one.

### Key Go stdlib primitives verified as usable for hotboot

- `(*net.TCPListener).File() (*os.File, error)` — returns a `*os.File` whose `Fd()` is a dup of the listener socket. The original listener continues to work in Go-level accept loops, but the new FD is CLOEXEC-clear (subject to verification per-platform).
- `(*net.TCPConn).File() (*os.File, error)` — same story for a connection. **Windows caveat (verified in Go stdlib docs 2026-04-18):** "On Windows, the returned os.File's file descriptor is not usable on other processes." This is an independent obstacle to hotboot on Windows, on top of `syscall.Exec` being Unix-only — Design A is doubly infeasible on Windows.
- `syscall.Exec(argv0, argv, envv)` — Unix `execve` wrapper. Replaces the current process image; PID is preserved; open FDs without `FD_CLOEXEC` survive. Not available on Windows. Go's stdlib does not export a direct Windows equivalent that preserves FDs.
- `syscall.Dup2(oldfd, newfd int) error` — used by the PoC to pin FDs at specific numbers before exec. **Portability note:** present on linux/amd64 and darwin; **absent on linux/arm64** (which only exposes `syscall.Dup3(oldfd, newfd int, flags int)`). The executable plan must abstract this behind a build-tagged helper or switch to `syscall.Dup3(oldfd, newfd, 0)` which is equivalent when flags==0 and is portable across Linux architectures.
- `net.FileListener(*os.File)` — re-wraps a file into a `net.Listener`. **Verified in PoC A (below).**
- `net.FileConn(*os.File)` — re-wraps a file into a `net.Conn`. **Verified in PoC A.**
- `os.StartProcess(argv0, argv, &os.ProcAttr{Files: []*os.File{...}})` — forks a child, passes files via inheritance. The files slice's position in the slice becomes the child's fd: `Files[0]` → child fd 0, etc.
- `exec.Command(argv0).Start()` — higher-level `StartProcess` wrapper. Does NOT preserve arbitrary FDs by default; use `cmd.ExtraFiles` for post-stderr FDs (they land at fd 3, 4, ...).

---

## Go design alternatives

Two viable designs are presented. I considered a third (goroutine handoff via shared memory) and rejected it as infeasible for cross-process work in Go.

### Design A — `syscall.Exec` + FD inheritance (C-equivalent)

**Technical description.** On `do_hotboot`:
1. Refuse if anyone is fighting (C's gate — match).
2. Refuse if anyone is in the line editor (C's gate — match).
3. Save world state (NPCs + room-floor objects) to `db/hotboot/mobfile.dat` + `db/hotboot/<vnum>.objdat`.
4. For each playing descriptor, save the player to the usual pfile, append a line to `db/hotboot/hotboot.dat` with `{sequence, room_vnum, port, idle, name, host, user_chose_ansi_bool}`.
5. Extract FDs via `(*net.TCPListener).File()` and `(*net.TCPConn).File()`. Clear `FD_CLOEXEC` on each. The server's `Stop()` is NOT called (that would close the listener).
6. Build an argv: `["smaug-go", "-port", "<port>", "-data", "<dir>", "--hotboot-recover", "<ln_fd>", "<d1_fd>", "<d2_fd>", ...]`.
7. `syscall.Exec(selfExe, argv, os.Environ())`. On failure, log and return — server keeps running (C's fallback too).

On startup, `cmd/smaug/main.go` checks for `--hotboot-recover`. If set:
1. Parse the listener FD from argv. Wrap via `net.FileListener(os.NewFile(lnFD, "listener"))`.
2. Boot areas (fresh load).
3. Read `db/hotboot/hotboot.dat` and `mobfile.dat` and `*.objdat`. For each hotboot.dat line:
   - Wrap the FD via `net.FileConn(os.NewFile(dFD, "conn"))`.
   - Build a fresh `DescriptorData`, `Conn = resumed conn`, `Host = saved host`, etc.
   - Spawn the read goroutine (`server.readLoop`-style) to pipe into `d.InputQueue`.
   - Load the player via `persist.LoadPlayerWithWorld`, set `Connected = CON_PLAYING`, place in saved room.
   - Emit "Time resumes its normal flow.\n\rA puff of ethereal smoke dissipates around you!\n\r".
4. Unlink the hotboot files after recovery (C does this).
5. Pass the recovered `server` + descriptors into the game loop. Game loop proceeds normally.

**What is preserved vs C:**
- TCP 5-tuple (client, port, listener endpoint) — ✅ preserved (same as C).
- Client never sees disconnect — ✅ preserved (same as C).
- World state (mob positions, loot on floor) — ✅ preserved via save/reload (C does the same).
- Player state — ✅ saved to pfile on hotboot; reloaded on recovery (C does the same).
- Active combat — ❌ refused by C and Go (same).
- Line editor session — ❌ refused by C and Go (same).
- MCCP zlib compression state — ❌ LOST in Go (same as C — C calls `compressEnd(d)` before exec).
- MSDP / MSSP negotiation state — ❌ LOST (same as C; client renegotiates).
- Mid-mudprog sleeping programs — ⚠️ NOT PRESERVED by Go (C doesn't preserve either — `MProgSleepData` queue is process-local in both). Acceptable loss.
- Timers on characters (`TIMER_RECENTFIGHT`, etc.) — partially preserved via `save_mobile`'s not-included timer list. **Neither C nor Go persists timers through hotboot today.** Document as a known non-loss, not a regression.
- `PCData.Hotboot` flag — set pre-hotboot so recovery can detect "came through hotboot" (C does this; Go mirrors).
- Process PID — ✅ preserved (syscall.Exec reuses PID).
- Listener socket → ✅ same fd, rebound in child.

**What is NEW vs C:**
- Goroutine-per-connection readLoops are restarted fresh in the child. The `InputQueue` is recreated. A client that sent bytes between "save state" and "exec" MAY have those bytes buffered in the kernel receive queue — the new readLoop consumes them normally (FIFO preserved at the OS layer). No data loss.

**Implementation surface estimate:**
- `internal/act/hotboot.go` — `DoHotboot` command: ~80 Go LOC (gates + save-call + exec). Mirrors `src/hotboot.c:599-741`.
- `internal/persist/hotboot.go` — save/load world: ~250 Go LOC (`SaveWorld`, `LoadWorld`, `SaveMob`, `LoadMob`). Mirrors `src/hotboot.c:73-596`.
- `internal/boot/recover.go` (new file) OR extension of `boot/boot.go` — `BootRecover(...)`: ~120 Go LOC. Parses argv, wraps FDs, calls bootDB, calls loadHotbootState, repopulates descriptors.
- `internal/net/server.go` — add `ServerFromListener(ln net.Listener) *Server` and `ResumeDescriptor(d *types.DescriptorData) (start readLoop)`: ~40 Go LOC.
- `cmd/smaug/main.go` — detect `--hotboot-recover` argv; branch into `BootRecover`: ~30 Go LOC.
- `internal/types/descriptor.go` — add `Cookie`, `ResumedFromHotboot bool` fields for diagnostic: ~5 LOC.
- `internal/types/pcdata.go` — add `Hotboot bool` transient flag (not persisted): ~2 LOC.
- `internal/game/loop.go` — add graceful `Stop()` for hotboot that does NOT close descriptors or listener: ~30 LOC.

Total: **~555 Go LOC** across 7 files; 1 new file (`internal/persist/hotboot.go`).

**Tests required:**
- Unit: SaveMob / LoadMob round-trip (one NPC with inventory + affects).
- Unit: SaveWorld / LoadWorld round-trip (two NPCs in two rooms with carried objects).
- Unit: hotboot.dat serialization round-trip.
- Unit: `DoHotboot` gate tests — combat-refuse, editor-refuse, imm-only.
- Integration: two-client hotboot via `testclient` harness — requires a new in-process re-exec shim (see § PoC Results Design A for the FD-preservation mechanism). This is the hard test. Candidate approach: a helper that does NOT call `syscall.Exec` but instead tears down the game loop, rebuilds world state from hotboot files, re-wraps FDs via a unix-socketpair (testclient shim). Keeps the test in-process and deterministic. **Acceptable cut for executable plan:** unit-level round-trip coverage + one `os/exec.Command`-based integration test that actually spawns a child binary and reconnects (end-to-end fidelity, but slower test; opt-in with a build tag).
- Manual/accept: `telnet localhost 4000`, `hotboot`, observe prompt reappears with "Time resumes".

**Risk assessment:**

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| `(*net.TCPListener).File()` leaks the Go runtime's poll state — listener "works" in the child but has subtle goroutine-blocking behavior | Low | High | PoC A validated accept + read + write on the resumed listener. Still: add an integration test that post-hotboot the server can accept a NEW connection after the existing ones are recovered. |
| `FD_CLOEXEC` default on Go's sockets — inherited FDs show up closed in the child | Low (verified set-to-0 works) | High | PoC A explicitly clears CLOEXEC on FDs 3 and 4 via `F_SETFD`. Document this invariant. Add a startup check in the child: if the expected FD is closed, log and abort. |
| Windows has no `syscall.Exec` — hotboot is Linux/macOS only | High (Windows is a stated target per `plan.md`?) | Unclear | **Open Question #1.** If Windows matters, Design A is infeasible there; Windows must fall back to Design B. Current status: `plan.md` mentions "single static binary" with no OS-specific code (no Cgo), but doesn't explicitly claim Windows is a production target. |
| Zombie goroutines after exec — the parent's readLoop goroutines are destroyed by exec, but if any are mid-write to channels with no reader, they may deadlock before exec completes | Medium | Low (exec happens regardless; goroutines die with process image) | syscall.Exec is atomic — no window where goroutines "try to recover." Not a real risk, just a source of log noise. |
| Area-load time dominates pause window (63ms measured) — during this window the kernel accept queue buffers new connections | Low | Low | Existing clients see no disconnect; new clients queue and connect after recovery. 63ms is imperceptible. |
| Inherited FD has data in receive buffer that pre-dates hotboot — child's readLoop consumes stale data | Low | Low | stale data by construction is a complete command the user typed; processing it normally is correct behavior. Document. |
| A descriptor is mid-flush on the parent's output buffer at exec time — bytes in `outBuf` are lost (they were never written to the socket) | Medium | Low | Flush all descriptors synchronously **before** calling `syscall.Exec`. `DoHotboot` adds this step. |
| `dlclose` equivalent — C closes its dl handle pre-exec. Go has no dl handles. | None | None | Skip. |
| `set_alarm(0)` — C clears SIGALRM. Go uses `time.Ticker`; goroutines handle signals via os/signal. No alarm to clear. | None | None | Skip. |

### Design B — save + graceful restart + reconnect-cookie

**Technical description.** On `do_hotboot`:
1. Refuse in combat / editor (same gates as A).
2. Save world state (same format as A — SaveWorld/SaveMob/SavePlayer).
3. Generate a per-descriptor `cookie` (random UUID). Write `db/hotboot/reconnect.dat` with `{cookie → name, host, port, room_vnum, idle, ansi_prefs, ...}`.
4. Send each player: `RECONNECT <cookie> localhost 4000\n` (or equivalent in-band protocol marker; design choice — see Open Question #2). Client parses this line, disconnects, awaits child readiness, dials port 4000, issues `RESUME <cookie>`.
5. Cleanly close all connections + listener.
6. `os.StartProcess` a child (or just `os.Exit(2)` and rely on a supervisor to relaunch — see Open Question #3).
7. Child boots normally; on accept, it checks for a `RESUME <cookie>` line instead of `CON_GET_NAME`. If cookie matches `reconnect.dat`, load the player and enter game. If cookie doesn't match, reject.

**What is preserved vs C:**
- TCP 5-tuple — ❌ NOT preserved; client sees a new connection.
- Client never sees disconnect — ❌ NOT preserved; client MUST disconnect and reconnect.
- Unmodified telnet works — ❌ NO. `telnet` sees "Connection closed by foreign host", user types `telnet localhost 4000` again manually. Can present a `RESUME <cookie>` instruction but the user would have to type it.
- World state — ✅ preserved via save/reload (same as A).
- Player state — ✅ saved to pfile (same as A).
- Active combat — ❌ refused (same as A).
- Line editor session — ❌ refused (same as A).
- MCCP state — ❌ LOST (same as A).
- MSDP / MSSP — ❌ LOST (same as A).
- PID — ❌ new PID (new process).
- Port stability — ✅ preserved; child rebinds same port.

**What is NEW vs C:**
- Reconnect-cookie protocol — out-of-spec for vanilla telnet. Requires client-side cooperation. Modern MUD clients (mushclient, tintin++, mudlet) can usually script reconnect; telnet cannot.

**Implementation surface estimate:**
- `internal/persist/hotboot.go` — save/load world (identical to Design A).
- `internal/persist/reconnect.go` — cookie file format: ~60 Go LOC.
- `internal/act/hotboot.go` — `DoHotboot` command: ~60 Go LOC.
- `internal/game/loop.go` — extend nanny with `CON_GET_RECONNECT_COOKIE` state: ~80 Go LOC.
- `cmd/smaug/main.go` — if reconnect.dat exists on boot, load it (child mode): ~20 Go LOC.
- `internal/types/descriptor.go` — `Cookie` field: ~2 LOC.

Total: **~470 Go LOC** across 6 files (slightly less than A because no FD-plumbing).

**Tests required:**
- Same unit tests as A for world save/load.
- Unit: reconnect-cookie file format round-trip.
- Unit: `CON_GET_RECONNECT_COOKIE` nanny state transitions.
- Integration: via testclient — spawn two clients, trigger hotboot, handle the `RECONNECT` message, reconnect with cookie, verify in-room state preserved.
- Manual: `telnet` then `hotboot` → see `Connection closed by foreign host` — confirmed expected behavior.

**Risk assessment:**

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| User experience: vanilla `telnet` users must manually reconnect every hotboot | Certain | Medium | Document; print instructions clearly on reconnect. Advise users to use a scripted client. |
| Cookie leak — attacker reads `reconnect.dat` and hijacks a session | Low | High | 128-bit random cookie + file mode 0600 + short TTL (5min). Same discipline as `session_id` cookies. |
| Cookie replay — attacker who captures the cookie reuses it | Low | High | One-shot cookies — reject on second use. |
| Bind race — child starts before parent releases the port | Low | Low | Parent does `ln.Close()` then exits BEFORE child's `os.StartProcess`. OR use `SO_REUSEADDR` / `SO_REUSEPORT`. |
| Client disconnects permanently because reconnect dialog is unclear | Medium | Medium | Requires client cooperation. Not fixable server-side. |
| No signaling protocol — client sees `RECONNECT` as a play-content string | Medium | Low | Use an out-of-band delimiter (`\x00\x00RECONNECT ...\x00\x00`)? Telnet IAC sub-negotiation? In-band plain text works for script-aware clients; fails for naive telnet. |

---

## Design PoC Results

Two PoCs were built to verify technical feasibility. Both are in `/tmp/` and are NOT committed.

### PoC A — `syscall.Exec` + FD inheritance

**Source:** `/tmp/hotboot-poc-a/main.go` (~180 Go LOC); `/tmp/hotboot-poc-a/test-client.go` (~80 Go LOC).

**Build + run:**
```
cd /tmp/hotboot-poc-a
go build -o poc-a main.go
go build -o test-client test-client.go
(./poc-a > server.log 2>&1 &)
sleep 0.3
./test-client
```

**Observed output (client side):**
```
16:28:29.760575 client: connected local=127.0.0.1:53484 remote=127.0.0.1:8765
16:28:29.760895 client: <- greeting  "PRIMARY-READY"
16:28:29.760938 client: -> HELLO
16:28:29.761343 client: <- ack  "HELLO-ACK"
16:28:29.761367 client: -> HOTBOOT
16:28:29.761727 client: <- rebooting  "REBOOTING"
16:28:29.765715 client: <- post-exec-resumed  "RESUMED"
16:28:29.765742 client: -> AFTER-REBOOT
16:28:29.766125 client: <- post-exec-ack  "ACK-AFTER-REBOOT"

=== PoC A RESULT ===
Local addr stable:   true (was 127.0.0.1:53484)
Remote addr stable:  true (was 127.0.0.1:8765)
Post-exec greeting:  "RESUMED"
Final ack:           "ACK-AFTER-REBOOT"
VERDICT: PASS -- inherited FDs survived syscall.Exec; TCP session preserved.
```

**Observed output (server side):** server pid `581704` logs both "primary: execing" and "post-exec: up (pid 581704)" — **same PID**, confirming `syscall.Exec` replaced the image in place.

**Evidence:**
- `(*net.TCPListener).File()` works.
- `(*net.TCPConn).File()` works.
- `syscall.Dup2` + clear `FD_CLOEXEC` + `syscall.Exec` passes both FDs to the child.
- `net.FileListener(os.NewFile(3, "listener"))` re-wraps the listener.
- `net.FileConn(os.NewFile(4, "client"))` re-wraps the conn.
- Client's TCP 5-tuple (local `53484`, remote `8765`) is unchanged.
- Bidirectional I/O works after exec (the client read "RESUMED" and then wrote "AFTER-REBOOT" and read back "ACK-AFTER-REBOOT").
- Round-trip pause (HOTBOOT send → RESUMED receive): **~4ms** in the minimal PoC.

**Scaled estimate for smaug-go:** measured boot-to-ready of real `smaug-go -data ../db` = **63ms** (reproducible via `time` + poll for "SMAUG MUD is ready"). So expected real-world hotboot pause with Design A is **~65-70ms**. This is imperceptible to a telnet user (human reaction ~200ms).

**Conclusion:** Design A is technically feasible in Go 1.26 on linux/amd64. No blockers found. Note that the PoC uses `syscall.Dup2` which is not exported on linux/arm64 — the executable plan must use `syscall.Dup3(old, new, 0)` instead (same semantics when flags==0, portable across Linux architectures). Darwin is also supported via `syscall.Dup2`.

**Note on log ports:** the `server.log` block above shows ports `55736`/`8765` from a re-run of the PoC; the client log block higher up shows `53484`/`8765` from the original run. Both runs produced identical `VERDICT: PASS` output — ephemeral client ports vary per invocation; the stability assertion ("local addr stable") is about same-port-before-and-after-exec, not the specific port number.

### PoC B — save + restart + reconnect-cookie

**Source:** `/tmp/hotboot-poc-b/main.go` (~170 Go LOC); `/tmp/hotboot-poc-b/test-client.go` (~110 Go LOC).

**Build + run:**
```
cd /tmp/hotboot-poc-b
go build -o poc-b main.go
go build -o test-client test-client.go
(./poc-b > server.log 2>&1 &)
sleep 0.3
./test-client
```

**Observed output (client side):**
```
16:30:18.834040 client: <- "PRIMARY-READY cookie-1776555018833778307"
16:30:18.834728 client: <- "HELLO-ACK"
16:30:18.834761 client: -> HOTBOOT at 2026-04-18 16:30:18.834743585
16:30:18.835308 client: <- "RECONNECT cookie-1776555018833778307"
16:30:18.858861 client: <- "RESUMED-IN-ROOM TempleOfMota"
16:30:18.859297 client: <- "ACK-PING"

=== PoC B RESULT ===
HOTBOOT -> RECONNECT instr: 590.327µs
reconnect dial duration:    22.798864ms
dial-ok -> resumed:         727.532µs
total blackout window:      23.526396ms (RECONNECT instr -> resume ack)
VERDICT: PASS -- save-state + reconnect-cookie round-trip works.
```

**Evidence:**
- JSON state file round-trip works.
- `os.Executable()` + `exec.Command("child", "--resume").Start()` works.
- Child rebinds port after parent closes listener and exits (22.8ms from dial → success).
- Cookie handshake authenticates the reconnect.
- Full round-trip (disconnect → reconnect → resume ack): **23.5ms** in the minimal PoC.

**Scaled estimate for smaug-go:** add real `smaug-go` boot time (63ms) + client-side dial retry backoff (10ms first retry if unlucky) + `RESUME <cookie>` exchange (negligible) = **~80-100ms**. Still fast — BUT the client MUST reconnect; vanilla telnet breaks.

**Conclusion:** Design B is technically feasible. The UX gap is the issue, not the technology.

---

## Recommendation

**Recommend Design A** (`syscall.Exec` + FD inheritance).

### Rationale

1. **UX parity with C SMAUG.** C hotboot's defining property is "players experience 2s pause instead of disconnect." Design A preserves this; Design B does not. A player base that expects `hotboot` to be seamless will be disappointed by a design that forces reconnect. The whole point of the feature is to avoid the reconnect.

2. **Works with unmodified telnet clients.** Many SMAUG users connect via `telnet`, `nc`, or web-based telnet proxies. None of these can script a cookie-reconnect protocol. Design A works universally.

3. **~65ms vs ~100ms pause** — practically indistinguishable, but the UX difference (seamless vs reconnect) is massive.

4. **Implementation cost comparable.** Design A is ~555 Go LOC; Design B is ~470 Go LOC. The ~85 Go LOC cost for Design A's FD-plumbing (Files slice, Dup2, CLOEXEC clear) is not material.

5. **PoC A passed end-to-end** with no surprises. The Go stdlib exposes every primitive needed (`*net.TCPListener.File`, `*net.TCPConn.File`, `net.FileListener`, `net.FileConn`, `syscall.Exec`, `syscall.Dup2`).

6. **C-fidelity.** The project's "no gameplay invented" and "match C behavior where it's specified" conventions (see `phase6-roadmap.md` "explicit non-goals") argue for matching C hotboot's observable behavior.

### When Design B becomes the right answer

- **Windows support becomes a hard requirement.** `syscall.Exec` is Unix-only; no Go stdlib on Windows preserves FDs across a new process image. Design B works on Windows.
- **Docker/Kubernetes deployment** where the process must not replace its image (container orchestrators may treat syscall.Exec as a crash+restart depending on PID 1 semantics). Design B's explicit `os.Exit(2)` + supervisor-relaunch is easier to reason about in a container.

These are out-of-scope for current Phase 6 (project ships Unix-only today per `plan.md`). Revisit if requirements change.

### Hybrid option (not recommended, noted for completeness)

A "dual-stack" hotboot that uses Design A on Linux/macOS and Design B on Windows/containers is implementable but adds significant complexity (conditional compilation via `//go:build`, two code paths to test, two sets of tests, two UX contracts). Only pursue if Windows support becomes required AND seamless Unix hotboot must be preserved. Tracked as Open Question #1.

---

## Prerequisites (must land before executable plan)

The following are prerequisites. Each is a small piece of plumbing; none is a full Phase 6 tier.

1. **Graceful-shutdown seam in `net/server.go`.** Currently `Server.Stop()` closes the listener and kills the accept loop. Hotboot needs a method that stops the accept loop but does **not** close the listener FD (because that FD is about to be passed to the child via exec). Proposed: `Server.PauseForHotboot() (ln net.Listener, descConns []net.Conn)` — returns the still-open listener + the still-open per-descriptor conns, and shuts down the accept/read goroutines without closing.

2. **`Stop()` decomposition in `GameLoop`.** Currently `g.Cancel()` cancels the context; the loop exits but does not flush remaining output. Hotboot needs `g.StopForHotboot()` that flushes all descriptors synchronously, then returns. Mirror pattern of `ClearTimerRegistry()` — explicit tear-down ordering.

3. **`db/hotboot/` directory creation + permission plumbing.** Mirror `db/player/` — created on demand by `os.MkdirAll(..., 0755)`. Add to boot in `bootDB` under an "ensure writable" pass.

4. **`PCData.Hotboot bool` transient flag.** Set pre-exec; cleared post-recovery. Used as a diagnostic — "this player came back via hotboot" triggers a different welcome message (`"Time resumes its normal flow.\n\r"` vs motd). **Already declared** at `internal/types/pcdata.go:117` (with a leading `// Hotboot` comment) from earlier scaffolding work. No schema change required; only the set/clear wiring is new.

5. **`DescriptorData.ResumedFromHotboot bool` diagnostic flag.** Same purpose, descriptor-scoped.

6. **NPC persistence format decision.** C uses `Vnum/Level/Gold/Room/Coordinates/HpManaMove/Position/Flags/AffectedBy/AffectData + embedded #OBJECT`. Go needs the same format — or a Go-native decision, but fidelity is cheaper. Recommendation: match C. Follow-up: will `save.go`'s existing `fwrite_obj` equivalent be reusable, or does NPC-save need its own writer? Answer: `persist/player.go`'s object-save helpers are reusable for the `#OBJECT` block within each `#MOBILE` entry.

7. **Boot-time detection of `--hotboot-recover` argv flag.** Branches boot into recovery path (loads hotboot files in addition to area resets) or normal path (just area resets).

8. **Integration test harness for hotboot.** The testclient (`internal/testclient`) is in-process. Hotboot is cross-process by definition. Option: make the testclient fork a real `smaug-go` binary on ephemeral ports for a subset of tests tagged `//go:build integration`. Executable plan: decide whether to land this infra in the hotboot plan or punt to a Test-Infra sub-plan.

9. **MCCP/MSDP/MSSP reset semantics documented.** Hotboot loses negotiation state; the client renegotiates on recovery. Document this in the welcome message. No code change needed (Design A's readLoop restart already does fresh negotiation).

10. **`do_hotboot` gate: combat-refuse + editor-refuse.** These require iterating `w.Descriptors` and checking `Character.Fighting` / `Connected == CON_EDITING`. Trivial.

None of these prerequisites is a hard blocker for a standalone executable plan — all can be landed as part of the first Task Group of the executable plan. Listing them here so the follow-up plan author knows the surface.

---

## Task groups — skeleton for executable follow-up plan

These are **scope statements only**. The executable-plan rewrite will expand each into test-first task detail with acceptance criteria, file paths, mutation-verify steps, and adversary-verification notes. Follows the `plan-tranche-b.md` / `plan-phase6-arena.md` shape.

### G1 — World-state persistence (`internal/persist/hotboot.go`)

Scope: `SaveWorld(w *world.World, dir string) error` + `LoadWorld(w *world.World, dir string) error`. Matches C `save_world` / `load_world`. Mob file = `db/hotboot/mobfile.dat`; per-room object file = `db/hotboot/<vnum>.objdat`. Format per C `src/hotboot.c:73-204, 433-596`. Round-trip tests.

**Depends on:** `persist/player.go` object writer helpers (reusable).
**Blocks:** G3, G4.

### G2 — Hotboot descriptor state file (`internal/persist/hotboot_sessions.go`)

Scope: `SaveHotbootSessions` / `LoadHotbootSessions` covering the hotboot.dat format — per-descriptor record with `fd_index, room_vnum, port, idle, name, host, ansi`. Round-trip tests with empty list, one entry, multiple entries.

**Depends on:** nothing.
**Blocks:** G3, G5.

### G3 — `DoHotboot` command (`internal/act/hotboot.go`)

Scope: immortal-only command (level ≥ `LEVEL_ASCENDANT`). Gates: no combat, no editor sessions, no idle connections in mid-nanny. Saves world (G1), saves player files, writes hotboot.dat (G2), extracts FDs via `ServerFromLoop().PauseForHotboot()`, sets `CloseOnExec=false` on each FD, builds argv, calls `syscall.Exec`. Fallback on exec failure: log, reopen. Mirrors `src/hotboot.c:599-741`.

**Depends on:** G1, G2, prerequisites 1-2.
**Blocks:** G6.

### G4 — Recovery path (`internal/boot/recover.go`)

Scope: new file. `BootRecover(w *world.World, dataDir string, opts BootOpts, lnFD uintptr, descFDs []uintptr, descMeta []HotbootSession) (*command.Registry, *game.GameLoop, *net.Server, error)`. Wraps the listener FD, wraps each descriptor FD, loads each player from pfile + their session row, places in saved room, emits puff-of-smoke social + "Time resumes its normal flow" welcome.

**Depends on:** G1, G2, prerequisite 7.
**Blocks:** G5, G6.

### G5 — `cmd/smaug/main.go` argv parsing + boot branch

Scope: `main.go` reads os.Args; if `--hotboot-recover <ln_fd> <d1_fd>[:<cookie_idx>] ...` is present, call `BootRecover(...)` instead of the normal `Boot(...)`. Passes through the rest of the args (`-port`, `-data`) for consistency.

**Depends on:** G4.
**Blocks:** G6.

### G6 — Integration test (`internal/testclient/hotboot_test.go` OR a new `cmd/smaug/hotboot_integration_test.go`)

Scope: spawn real `smaug-go` on an ephemeral port. Connect two telnet clients. Log in, walk to different rooms. Have an immortal client issue `hotboot`. Observe both clients receive the puff-of-smoke social and their prompt returns. Run a final command on each to confirm the session is live. Tagged `//go:build integration` so it doesn't run on every `go test ./...`.

**Depends on:** G1-G5 complete.
**Blocks:** acceptance.

### G7 — Documentation + CHANGELOG

Scope: append completion record to this plan doc; append CHANGELOG entry; update `TODO.md` with completed item; `phases.md` Phase 6 table — mark Hotboot `Complete`.

**Depends on:** G6 passing.
**Blocks:** nothing.

---

## Open Questions

These require human decision before dispatching the executable plan.

1. **Windows support.** Is Windows a production target for smaug-go? `plan.md`'s "no Cgo, pure Go" discipline suggests yes; current build is Linux/macOS-only in practice (no Windows CI, no Windows-specific code). Design A requires `syscall.Exec` which is Unix-only, AND the returned FD from `(*net.TCPConn).File()` is documented as "not usable on other processes" on Windows — so even if a Windows-specific exec primitive were added, the inherited FDs could not be handed to a child process. Design A is therefore doubly infeasible on Windows. If Windows matters, the executable plan must add Design B as a Windows-specific fallback behind `//go:build windows` — roughly doubles implementation cost for a platform with no known users.
   - **Recommendation:** Defer Windows support. Design A Linux/macOS only; document it as a known limitation. Revisit when a user shows up.

2. **Reconnect-cookie in-band delimiter (Design B — not used, but for future reference).** If Design B ever ships (Windows fallback, container deployment), what's the in-band signal for "disconnect and reconnect with this cookie"? Plain text? Telnet IAC sub-negotiation (IAC SB 104 ... IAC SE, out-of-band)? MSDP variable push?
   - **Recommendation:** Plain text line starting with `\x1b[HOTBOOT-RECONNECT:<cookie>]\x1b[` — discoverable by scripted clients, shows as visible text to naive clients. Non-blocking; revisit if Design B is adopted.

3. **Supervisor-relaunch vs `syscall.Exec` for the executable plan.** Design A uses `syscall.Exec`. Some deployments run smaug-go under a supervisor (systemd, supervisord, Docker). `syscall.Exec` keeps the same PID; supervisors may or may not handle this correctly (systemd is fine; PID-1 containers may barf).
   - **Recommendation:** Document `syscall.Exec` behavior with respect to systemd; add a deployment note. Not a blocker.

4. **Idle-connection gate — refuse hotboot for connections in mid-login?** C's `do_hotboot:673` drops mid-login descriptors (writes "Sorry, we are rebooting, come back later" then `close_socket(d, FALSE)`). Go should mirror.
   - **Recommendation:** Mirror C. Simple.

5. **World-state save format: exactly C-faithful or Go-native?** C uses a text format with custom `fread_word` parsing. Go's existing `persist/area.go` + `persist/player.go` use the same Scanner. Format-faithful is cheapest; a Go-native (JSON or gob) format would diverge from the existing persistence infrastructure.
   - **Recommendation:** Format-faithful. Reuse `Scanner` + existing object writer helpers. Minimal novelty.

6. **Mob timer preservation on hotboot.** Neither C nor Go preserves character timers (`TIMER_RECENTFIGHT`, `TIMER_DO_FUN`, etc.) through hotboot. Players return with cleared timers. Is this acceptable?
   - **Recommendation:** Mirror C (don't preserve). Document as a known non-loss. Timer preservation is a follow-up if players complain.

7. **Sleeping mudprogs on hotboot.** `MProgSleepData` queue is in-process state. Neither C nor Go persists it through hotboot. Scheduled mud-prog resume times are lost; any mudprog in `mpsleep` at hotboot time loses its remaining script.
   - **Recommendation:** Mirror C (don't preserve). Document.

8. **`hotboot` command name** — match C or add a friendlier alias?
   - **Recommendation:** Match C. `hotboot` is the established MUD-admin verb.

9. **Post-hotboot welcome message.** C emits `"Time resumes its normal flow.\n\r"` + the puff-of-smoke social. Go should mirror. Open: should Go additionally show "Last updated: <timestamp>" for diagnostic?
   - **Recommendation:** Mirror C exactly. Diagnostic is a Phase-6 polish follow-up.

---

## Risk Analysis (roll-up)

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Design A technically infeasible on Linux | Low (PoC passed) | High (redesign forced) | PoC A validated on Go 1.26 Linux 6.6 WSL2. Real-world Linux kernels ≥ 3.10 support all primitives used. |
| Hotboot corrupts player state (partial save) | Medium | High | Use `tempfile + rename` pattern for every save (already used by `game.GameLoop.SavePlayer`). Hotboot test suite covers "kill -9 in the middle" via mutation testing. |
| FD leak across multiple hotboots | Medium | Medium | Explicit CLOEXEC-clear on each exec; the inherited FDs are the ONLY FDs not CLOEXEC-default. Child process starts fresh. |
| Area-file reload drift: a room vnum existed before hotboot but not after (area file deleted/renamed) | Medium | Low | Mirror C: fall back to `ROOM_VNUM_TEMPLE`. Log the event. |
| Player `pcdata->hotboot` flag never clears (bug on restore path) | Low | Low | Clear immediately after welcome message. Unit test. |
| Hotboot is invoked while a player is in CON_EDITING — C refuses; Go must refuse | Low | Low | G3 implements the gate. Unit test. |
| Argv length limit — passing N FDs as argv hits `ARG_MAX` (typically 128KB) | Low (would need 10000+ descriptors) | Low | Not a problem at MUD scale. Document the bound. |
| File locks on Windows / pfile open during exec | N/A | N/A | Not a Windows target (Open Q #1). |
| `os.Executable()` returns the wrong path (symlinks, `$PATH` shadowing) | Low | High (exec fails) | Resolve to absolute path pre-exec via `filepath.EvalSymlinks`. Log on failure. |
| Concurrent hotboot request — two immortals type `hotboot` at the same pulse | Low | Medium | `w.SysData.HotbootInProgress bool` mutex-protected (or single-threaded pulse semantics already prevent this). Mirror C's implicit single-threading. |
| Fresh process init time dominates pause (63ms measured; could grow) | Low | Low | 63ms is current baseline. Large area additions may grow it. Budget: ≤ 500ms to match C's perceived-instant pause. Monitor. |

---

## Reference Plans for Shape

- `plan-tranche-b.md` — 5 task groups, 18 acceptance criteria (reference size for G1-G5 executable plan).
- `plan-phase6-arena.md` — full Phase-6 plan with Open Questions, Risk Analysis, Reference Plans; 7 task groups, 15 acceptance criteria.
- `plan-combat-depth.md` — largest-scope precedent (9 task groups, 12 acceptance criteria).

The executable follow-up plan should size between `plan-tranche-b.md` and `plan-combat-depth.md` — 6-7 task groups with ~15 acceptance criteria.

---

## Dispatch Protocol for Future Manager

When a manager is dispatched to **rewrite this document as an executable plan**:

1. Read this design doc in full. Treat § "Recommendation" as authoritative unless Open Question #1 (Windows) resolves against it.
2. Re-verify PoC A command sequences run green on the current tree. If `syscall.Exec` behavior has changed in a newer Go release, flag and escalate.
3. Write the executable plan as a full rewrite of this file (same filename, different content). Keep the problem statement and C reference. Replace the design-alternatives section with the chosen design's full detail. Expand each task group from scope-only (§ Task Groups) to test-first with file paths, mutation-verify steps, and acceptance criteria (per `plan-phase6-arena.md` shape). Add an acceptance-criteria checklist (15-20 items). Keep § Open Questions, Risk Analysis, Reference Plans.
4. Dispatch the adversary BEFORE worker dispatch to verify the executable plan shape, C citations, Go-surface facts, and task-group coverage — same pattern as `plan-phase6-arena.md`'s pre-execution adversary pass.
5. Execute G1→G6 in waves (G1+G2 parallel; G3 after both; G4 after G2; G5 after G4; G6 after all).

When a manager is dispatched to **execute the executable plan** (once it exists):
1. Resolve remaining Open Questions via human input if needed.
2. Dispatch workers per the task-group graph.
3. Adversary-verify each task group.
4. Integration test as G6.
5. Append completion record to the rewritten plan; CHANGELOG + TODO + CLAUDE.md patch draft.

---

*Design-exploration doc complete. Adversary-verification notes for this design pass — to be appended after planning adversary review.*
