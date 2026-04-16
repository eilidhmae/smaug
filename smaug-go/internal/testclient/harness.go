// Package testclient provides a reusable in-process test client that drives
// a real telnet session against an ephemeral SMAUG game server. Tests in
// any _test.go file in this module can import it.
//
// # Serialization
//
// SMAUG's runtime holds cross-package globals (act.WorldRef,
// mudprog.WorldRef, combat.WorldRef, act.CmdRegistry, act.ShutdownFunc,
// act.DisconnectFunc, the combat.*Hook function vars, etc.) that are set
// during boot. Two concurrent Harness instances would race on those
// globals. To keep tests safe, Start acquires a package-level mutex that
// is released by the t.Cleanup chain. The practical effect: testclient-
// based tests are single-threaded per process. Inside a test, goroutines
// the harness spawns (the game loop, the network accept loop) are fine;
// only parallel Harness instances are disallowed.
//
// # Usage
//
//	h := testclient.Start(t)
//	c := h.Dial(t)
//	c.ReadUntil("what name", 3*time.Second)
//	c.Send("Gandalf")
//	…
//	h.Query(func(w *world.World) { … assertion … })
package testclient

import (
	"context"
	"io"
	gonet "net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/boot"
	"github.com/eilidhmae/smaug/internal/game"
	smaugnet "github.com/eilidhmae/smaug/internal/net"
	"github.com/eilidhmae/smaug/internal/world"
)

const dialTimeout = 2 * time.Second

// harnessMu serializes Harness lifetimes. See package doc comment.
var harnessMu sync.Mutex

// lockHarness acquires the package-level harness mutex and schedules its
// release on test cleanup. Anything that boots the game loop or mutates
// the cross-package globals wired by boot.Boot (act.WorldRef,
// act.SaveFunc, mudprog.WorldRef, combat hooks, …) must hold this lock,
// otherwise parallel/interleaved tests can race those globals. Shared
// between Harness.Start and the fixture-smoke tests in testdata_test.go.
func lockHarness(t *testing.T) {
	t.Helper()
	harnessMu.Lock()
	t.Cleanup(harnessMu.Unlock)
}

// defaultDataDir is the relative path used when no WithDataDir override is
// given. Resolved relative to the *caller* package's cwd (what `go test`
// sets for each package). Most harness callers live outside
// internal/testclient; for them, the correct default is two levels up to
// cmd/smaug/testdata. Internal/testclient's own tests must override.
const defaultDataDir = "../../cmd/smaug/testdata"

// harnessConfig carries the resolved option set. All fields have usable
// zero-values except dataDir (empty falls through to defaultDataDir).
type harnessConfig struct {
	dataDir   string
	port      int
	stripANSI bool
	stripIAC  bool
}

// Option mutates a harnessConfig. Returned by helpers like WithDataDir.
type Option func(*harnessConfig)

// WithDataDir overrides the data directory. Path is resolved relative to
// the caller package's cwd.
func WithDataDir(dir string) Option { return func(c *harnessConfig) { c.dataDir = dir } }

// WithPort requests a specific listen port. Default 0 (OS picks).
func WithPort(port int) Option { return func(c *harnessConfig) { c.port = port } }

// WithoutANSIStrip disables the ANSI escape stripper in Client.
func WithoutANSIStrip() Option { return func(c *harnessConfig) { c.stripANSI = false } }

// WithoutIACStrip disables the telnet IAC stripper in Client.
func WithoutIACStrip() Option { return func(c *harnessConfig) { c.stripIAC = false } }

// Harness is a running test-mode game server. Exported Port is the bound
// TCP port; Dial opens a new session against it.
type Harness struct {
	Port int

	world     *world.World
	loop      *game.GameLoop
	server    *smaugnet.Server
	dataDir   string
	stripANSI bool
	stripIAC  bool
	shutdowns chan boot.ShutdownRequest
	loopDone  chan struct{}
}

// Start boots an ephemeral game server for the current test. It loads the
// world from the configured dataDir, binds a random TCP port, starts the
// game loop goroutine, and registers cleanup handlers so the loop stops,
// the listener closes, and the package mutex releases at test end.
//
// Parallelism: Start serializes with other Start calls via a package-level
// mutex. Tests that use this harness must NOT call t.Parallel().
func Start(t *testing.T, opts ...Option) *Harness {
	t.Helper()

	cfg := harnessConfig{
		dataDir:   defaultDataDir,
		port:      0,
		stripANSI: true,
		stripIAC:  true,
	}
	for _, o := range opts {
		o(&cfg)
	}

	// Serialize with any other Harness. Released via t.Cleanup inside
	// lockHarness; all early-error paths that invoke t.Fatalf trigger
	// cleanup too, so the lock always drops.
	lockHarness(t)

	// Copy the source fixture into a per-test TempDir so parallel
	// packages (e.g. internal/boot's TestBoot_* suite, which does
	// os.RemoveAll on the same shared path) cannot delete our player
	// saves mid-test. `go test ./...` runs packages concurrently by
	// default — without isolation, one package's cleanup races against
	// another's live save file. t.TempDir is torn down automatically
	// after the test, so no extra cleanup is needed here.
	isolatedDir := copyDataDirToTempDir(t, cfg.dataDir)
	cfg.dataDir = isolatedDir

	w := world.New(cfg.dataDir)
	server := smaugnet.NewServer()

	shutdowns := make(chan boot.ShutdownRequest, 4)
	bopts := boot.TestOpts()
	bopts.Shutdowns = shutdowns

	_, loop, err := boot.Boot(w, cfg.dataDir, server.Incoming, bopts)
	if err != nil {
		t.Fatalf("boot.Boot: %v", err)
		return nil
	}

	addr := "127.0.0.1:0"
	if cfg.port != 0 {
		addr = "127.0.0.1:" + strconv.Itoa(cfg.port)
	}
	ln, err := gonet.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("listen: %v", err)
		return nil
	}
	port := ln.Addr().(*gonet.TCPAddr).Port

	if err := server.StartOnListener(ln); err != nil {
		ln.Close()
		t.Fatalf("server.StartOnListener: %v", err)
		return nil
	}

	loopDone := make(chan struct{})
	go func() {
		defer close(loopDone)
		// Boot wired a GameLoop that stops on its internalCtx, which the
		// test-mode ShutdownFunc triggers via Cancel. Run() also stops on
		// the ctx we pass — use a fresh context.Background since we rely
		// on Cancel().
		loop.Run(context.Background())
	}()

	h := &Harness{
		Port:      port,
		world:     w,
		loop:      loop,
		server:    server,
		dataDir:   cfg.dataDir,
		stripANSI: cfg.stripANSI,
		stripIAC:  cfg.stripIAC,
		shutdowns: shutdowns,
		loopDone:  loopDone,
	}

	t.Cleanup(func() {
		// Stop the loop first (so no pulses run after we tear down the
		// server) — Cancel is idempotent, safe even if a test already
		// fired the shutdown hook.
		loop.Cancel()
		<-loopDone
		server.Stop()
		// The data dir is a t.TempDir, so the test runtime tears it
		// down after cleanup finishes — no explicit player-dir purge
		// needed here.
		// harnessMu is released by lockHarness's own t.Cleanup, which
		// runs after this one (t.Cleanup is LIFO, so later registrations
		// fire first). Do not unlock here — double-unlock would panic.
	})

	return h
}

// Dial opens a new TCP connection to the server and returns a Client
// positioned before any server greeting has been read. The connection is
// registered with t.Cleanup so it closes automatically at test end.
func (h *Harness) Dial(t *testing.T) *Client {
	t.Helper()
	conn, err := gonet.DialTimeout("tcp", "127.0.0.1:"+strconv.Itoa(h.Port), dialTimeout)
	if err != nil {
		t.Fatalf("dial: %v", err)
		return nil
	}
	t.Cleanup(func() { _ = conn.Close() })
	return &Client{
		conn:      conn,
		t:         t,
		stripANSI: h.stripANSI,
		stripIAC:  h.stripIAC,
	}
}

// Query runs fn on the loop goroutine, serialized with pulse processing,
// and returns when fn completes. Use this for world-state assertions that
// must observe a consistent snapshot. fn receives the live *world.World —
// do not retain references to mutable world state after fn returns; make
// copies inside fn if you need to compare later.
//
// Query panics if the loop has already exited, if enqueueing stalls, or if
// fn does not complete within 5 seconds. Those conditions always indicate
// test bugs (calling Query after shutdown, or the pulse goroutine wedged)
// — panicking keeps the stack trace, whereas blocking would only surface
// via the outer `go test -timeout`.
func (h *Harness) Query(fn func(*world.World)) {
	done := make(chan struct{})
	select {
	case h.loop.QueryQueue() <- func() { fn(h.world); close(done) }:
	case <-h.loopDone:
		panic("testclient: Query called after loop exited")
	case <-time.After(5 * time.Second):
		panic("testclient: Query enqueue timed out after 5s — loop stalled?")
	}
	select {
	case <-done:
	case <-h.loopDone:
		panic("testclient: loop exited while Query was pending")
	case <-time.After(5 * time.Second):
		panic("testclient: Query fn timed out after 5s")
	}
}

// Shutdowns returns the channel that receives a ShutdownRequest each time
// an immortal shutdown/reboot hook fires. Tests assert via a select on
// this channel.
func (h *Harness) Shutdowns() <-chan boot.ShutdownRequest {
	return h.shutdowns
}

// copyDataDirToTempDir copies srcDir into a freshly-created t.TempDir
// and returns the absolute path to the copy. Relative srcDir paths are
// resolved against the caller package's cwd the same way boot.Boot
// treats dataDir arguments. Any I/O error here fails the test; the
// TempDir itself is scheduled for automatic cleanup by t.TempDir.
//
// Used by Start to give every Harness its own private player-save
// sandbox. Other packages (internal/boot, cmd/smaug) that do
// os.RemoveAll on their shared test-data path can no longer race with
// in-flight testclient saves.
func copyDataDirToTempDir(t *testing.T, srcDir string) string {
	t.Helper()
	absSrc, err := filepath.Abs(srcDir)
	if err != nil {
		t.Fatalf("resolve data dir %q: %v", srcDir, err)
	}
	info, err := os.Stat(absSrc)
	if err != nil {
		t.Fatalf("stat data dir %q: %v", absSrc, err)
	}
	if !info.IsDir() {
		t.Fatalf("data dir %q is not a directory", absSrc)
	}
	dstDir := t.TempDir()
	if err := copyTree(absSrc, dstDir); err != nil {
		t.Fatalf("copy data dir %q -> %q: %v", absSrc, dstDir, err)
	}
	return dstDir
}

// copyTree copies the contents of src into dst recursively. Files get
// mode 0644, directories 0755 — player saves don't need to preserve
// source-tree permissions, and hardening the dst modes here avoids
// depending on whatever umask the test runner was launched with.
func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

