# SMAUG MUD — Go Port

A pure Go port of the SMAUG MUD (Multi-User Dungeon), a text-based MMORPG built on the Diku/Merc/SMAUG lineage. No Cgo — compiles to a single static binary.

## Requirements

- Go 1.23 or later
- The `db/` data directory from the parent SMAUG repository (area files, classes, races, etc.)

## Quick Start

```bash
make build
make run
```

Then connect with any telnet client:

```bash
telnet localhost 4000
```

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make build` | Compile the server binary (`smaug-go`) |
| `make test` | Run all tests |
| `make test-v` | Run all tests with verbose output |
| `make test-race` | Run tests with Go's race detector |
| `make test-pkg PKG=util` | Run tests for a single package |
| `make test-run RUN=TestBitVector PKG=types` | Run a specific test by name |
| `make run` | Build and start the server (port 4000) |
| `make vet` | Run `go vet` static analysis |
| `make fmt` | Format all Go files with `gofmt` |
| `make fmt-check` | Check formatting without writing (exits non-zero if unformatted) |
| `make check` | Run vet + format check + all tests |
| `make clean` | Remove binary and test cache |

### Configuration

Override defaults with environment variables:

```bash
make run PORT=9000              # listen on port 9000
make run DATADIR=/path/to/db    # custom data directory
```

## Project Structure

```
cmd/smaug/main.go           Entry point
internal/
  types/                     Core data structures (17 files)
  util/                      String, dice, logging helpers
  world/                     Game state container
  persist/                   SMAUG file format readers
  net/                       TCP server, color processing
  game/                      Game loop, login state machine
  command/                   Command registry and interpreter
  act/                       Player commands
doc/                         Project documentation
```

See `doc/plan.md` for the full architectural plan and `doc/phases.md` for implementation phases.

## Testing

```bash
make test        # fast — all tests, ~0.05s
make test-v      # verbose — see every test name
make test-race   # with race detector — catches concurrency bugs
make check       # vet + format + tests — use before committing
```

284 test cases across 7 test files covering types, utilities, networking, file parsing, command interpretation, and world state management. All tests verified non-vacuous via mutation testing.

## Status

Phase 1 in progress. See `doc/phase1-completed.md` for what's done and `doc/phase1-remaining.md` for what's next.

## License

Based on SMAUG 2.0, which descends from SMAUG 1.4, Merc 2.1, and DikuMUD. See `../doc/license.txt` for full license terms.
