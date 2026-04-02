# Phase 4: Polish, Quality, and Remaining Systems

**Status: COMPLETE.** Phase 4a (quality, G1–G5) and Phase 4b (features, G6–G10, G12) are done. G11 (hotboot) deferred to Phase 5. See `phase4a-completed.md` and `phase4b-completed.md` for full records.

## Context

Phases 1-3 are complete: 66 source files, 36 test files, 572 test cases, 13 packages all passing. The Go port has a working game with combat, magic, skills, OLC, mud programs, shops, clans, and protocol support. Phase 4 shifts focus from feature implementation to quality, correctness, and filling remaining gaps.

**Current coverage by package:**
| Package | Coverage |
|---------|----------|
| util | 93.1% |
| command | 80.6% |
| cmd/smaug | 74.9% |
| handler | 69.7% |
| persist | 68.6% |
| combat | 65.4% |
| world | 51.9% |
| game | 50.9% |
| net | 50.0% |
| magic | 32.5% |
| types | 29.1% |
| mudprog | 28.1% |
| **act** | **27.0%** |

## Task Groups

### G1: Documentation Verification Pass
**Goal:** Read all phaseX-completed.md docs and verify every claimed function/feature actually exists in the codebase.

- Cross-reference each function listed in phase1-completed.md, phase2-completed.md, phase3-completed.md against actual source
- Flag any discrepancies (documented but missing, or present but undocumented)
- Update docs if needed
- **Output:** Verified docs or list of corrections

### G2: Live Integration Testing
**Goal:** Start the server, connect via telnet, and exercise the game end-to-end.

- Build and start the server: `go build -o smaug-go ./cmd/smaug/ && ./smaug-go -port 4000 -data ../../db`
- Write automated integration tests using programmatic TCP connections
- Run automated suite, then perform manual smoke tests via telnet
- Test: character creation, login, movement, combat, spells, shops, OLC, immortal commands
- Document and fix any bugs discovered
- **Phase 5 idea:** Build a minimal telnet test client for repeatable interactive testing

### G3: Code Coverage Push
**Goal:** Test critical paths thoroughly; raise coverage across all packages with mixed targets based on criticality.

**Coverage targets by tier:**

| Tier | Target | Packages | Rationale |
|------|--------|----------|-----------|
| Critical | 90%+ | combat, magic, persist, handler, game/update | Core gameplay + data integrity |
| Infrastructure | 80%+ | command, net, world, types, mudprog | System correctness |
| Commands | 70%+ | act | Focus on commands with logic (skills, shops, OLC, consume); accept lower coverage on pure-output commands |

Priority order (biggest gaps first):
1. **act** (27.0%) — test combat skills, shops, OLC, consume, wiz commands; skip trivial output-only commands
2. **mudprog** (28.1%) — expand driver, trigger, command tests
3. **types** (29.1%) — test descriptor methods, character helpers, struct methods
4. **magic** (32.5%) — test remaining spell functions
5. **net** (50.0%) — test telnet, MCCP, MSDP, MSSP integration
6. **game** (50.9%) — test pager, editor, nanny states, more update paths
7. **world** (51.9%) — test remaining World methods
8. **combat** (65.4%) — test dual wield, wimpy, edge cases
9. **handler** (69.7%) — test remaining find/affect paths
10. **persist** (68.6%) — test more area sections, edge cases
11. **command** (80.6%), **cmd/smaug** (74.9%) — fill remaining gaps

### G4: Go Idiom Pass
**Goal:** Audit the entire codebase for Go idiom violations and produce a document of improvements.

Areas to review:
- Error handling patterns (sentinel errors vs fmt.Errorf wrapping)
- Interface usage (io.Writer, io.Reader where appropriate)
- Naming conventions (Go style vs C-style holdovers)
- Package organization and dependency direction
- Use of `any` vs typed interfaces
- Goroutine/channel patterns
- Slice/map initialization idioms
- String building (strings.Builder vs fmt.Sprintf)
- Exported vs unexported symbols
- godoc-compatible comments on exported types/functions
- `context.Context` usage (or lack thereof)
- Test helper patterns (t.Helper(), testify vs stdlib)

**Output:** `smaug-go/doc/go-idiom-review.md` — categorized list of improvements with file:line references

### G5: Security Vulnerability Pass
**Goal:** Audit the codebase for security vulnerabilities and produce a document.

Areas to review:
- **Input validation:** telnet input sanitization, command argument parsing, buffer overflow equivalents
- **Path traversal:** player file save/load paths, area file paths
- **Injection:** any string interpolation into commands or file paths
- **DoS vectors:** unbounded allocations, goroutine leaks, connection limits
- **Authentication:** password storage (currently plaintext?), session management
- **Race conditions:** any shared state between goroutines without proper synchronization
- **Resource exhaustion:** file handle leaks, unclosed connections, memory leaks
- **Privilege escalation:** trust level checks on immortal commands, command bypasses
- **Data integrity:** area save corruption, player save corruption

**Output:** `smaug-go/doc/security-review.md` — categorized findings with severity, file:line, and recommended fix

### G6: Missing OLC Commands (mset, oset, rset + more)
**Goal:** Port the most-needed OLC property editors from C build.c.

- **do_mset** — edit mob template properties (level, stats, flags, descriptions, etc.)
- **do_oset** — edit object template properties (type, values, flags, weight, cost, etc.)
- **do_rset** — edit room properties (alternative to redit for quick flag/sector changes)
- **do_mpedit/do_opedit/do_rpedit** — edit mud program scripts on mobs/objects/rooms
- **do_rdelete/do_odelete/do_mdelete** — delete templates
- **do_aset/do_astat** — area property editing and stats

Estimated: ~20 functions, 2-3 files, tests for each.

### G7: Quest System
**Goal:** Port the automated quest system from quest.c (~685 lines C).

- Quest request/complete/list/buy/info/time/points commands
- Quest giver NPC interaction
- Quest objectives and rewards
- Quest state tracking on PCData

### G8: Banking System
**Goal:** Port banking from bank.c (~555 lines C).

- deposit, withdraw, balance commands
- Banker NPC detection (like shopkeeper)
- Gold storage on PCData

### G9: Ban System
**Goal:** Port ban management from ban.c (~1,398 lines C).

- Ban by site/name/class/race
- Ban expiry
- Ban check on login
- Immortal ban management commands

### G10: Mob Tracking/Hunting
**Goal:** Port BFS pathfinding from track.c (~485 lines C).

- BFS room search for target
- hunt_victim in mobile_update
- ACT_HUNTER flag behavior

### G11: Hotboot/Copyover
**Goal:** Seamless server restart without disconnecting players.

- Save all descriptor file descriptors + game state
- `syscall.Exec` to replace process
- Restore connections on startup
- **Note:** This is the most complex infrastructure task. Go's goroutine-per-connection model differs significantly from C's `select()` model. The approach will use `TCPConn.File()` to extract raw FDs, serialize game state, exec the new binary, then reconstruct connections. Platform-specific (Linux). Revisit design after G4 (Go idiom pass) which may reveal architectural patterns that simplify this.

### G12: Remaining Player Commands
**Goal:** Audit C command table against Go command table, identify and port missing standard commands.

- Compare `interp.c` command table with `main.go` registerCommands
- Port any missing commonly-used player commands
- Focus on commands players would expect in a standard SMAUG MUD

## Execution Order

```
Phase 4a - Quality (COMPLETE):
  G1 (doc verification) ✓ → G2 (integration test) ✓ → G4 (idiom pass) ✓ → G5 (security pass) ✓ → G3 (coverage) ✓

Phase 4b - Features (COMPLETE, G11 deferred):
  G6 (OLC) ✓ ─┐
  G7 (quests) ✓ ─┤
  G8 (banking) ✓ ─┤── all complete
  G9 (bans) ✓ ─┤
  G10 (tracking) ✓ ─┘
  G11 (hotboot) — DEFERRED to Phase 5 (Go's net.Conn doesn't support C's fd-inheritance via exec())
  G12 (remaining commands) ✓
```

## Phase 5 Candidates

These are large, self-contained optional systems. Each is a project unto itself:

- **Hotboot/copyover** — seamless server restart; needs Go-specific design (save state + graceful restart + auto-reconnect)
- **Overland maps** (3,752 lines) — 1000x1000 tile maps, ANSI rendering, landmarks
- **Player housing** (2,853 lines) — apartments, room customization, guests
- **Polymorph** (2,753 lines) — form shifting with stat mods
- **Archery** (1,362 lines) — ranged combat, arrow lodging
- **Combat stances** (1,022 lines) — 10 fighting postures with stat effects
- **Dragon flight** (945 lines) — coordinate-based flying system
- **Arena PvP** (358 lines) — challenge/accept isolated combat
- **Planes** (298 lines) — multi-planar system
- **Holidays** (416 lines) — calendar events
- **Star maps** (226 lines) — celestial display
- **Minimal telnet test client** — for repeatable interactive integration testing

## Verification (PASSED)

- All tests pass: `go test ./...` ✓
- Coverage report: `go test -cover ./...` ✓
- Documents exist: `doc/go-idiom-review.md`, `doc/security-review.md` ✓
- Server boots and accepts connections ✓
- Phase docs verified accurate ✓
- New commands functional: position, banking, bans, tracking, quests, OLC set, skills/spells, item-use, group, mount ✓
