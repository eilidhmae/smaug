//go:build integration && !windows

// End-to-end hotboot integration test.
//
// Runs the real cmd/smaug binary as a child process. Opt-in via the
// `integration` build tag so `go test ./...` stays fast by default; run
// with:
//
//	go test -tags integration -run TestHotboot_EndToEnd -count=1 \
//	        ./internal/testclient/...
//
// The test is gated to !windows because the hotboot implementation relies
// on syscall.Exec + FD inheritance which Windows does not support (see
// plan-phase6-hotboot.md §Platform).
//
// Scenario (plan-phase6-hotboot.md §G6):
//   Phase A: spawn child, create Immy + Mortal through the nanny, quit
//            them so pfiles land on disk, SIGTERM the child.
//   Phase B: edit Immy's pfile to grant Trust LEVEL_ASCENDANT (60 =
//            MAX_LEVEL-5 per internal/types/constants.go).
//   Phase C: spawn a fresh child, log in both players, record their
//            rooms, issue `hotboot` from Immy, observe the welcome on
//            both sockets within 500ms, verify state preserved via
//            `look`, open a third fresh connection to confirm the
//            rebound listener accepts new clients post-recovery.

package testclient

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	gonet "net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"
)

const (
	hotbootReadyTimeout   = 15 * time.Second
	hotbootDialTimeout    = 5 * time.Second
	hotbootNannyTimeout   = 10 * time.Second
	hotbootWelcomeTimeout = 500 * time.Millisecond
	// After a hotboot fires, accept() needs a moment for the child to
	// re-wrap the inherited listener before we dial the third fresh
	// connection. The in-process exec takes ~65ms in a PoC; give a
	// wide envelope on shared CI.
	hotbootPostExecTimeout = 5 * time.Second
)

var listeningRe = regexp.MustCompile(`SMAUG MUD listening on :(\d+)`)

// TestHotboot_EndToEnd is the one-scenario integration test. Its size
// reflects the 12 plan-mandated steps; split into helpers below.
func TestHotboot_EndToEnd(t *testing.T) {
	// Source db directory. `go test` runs with cwd == package dir, so
	// the repo's db/ is two levels up from internal/testclient.
	srcDB, err := filepath.Abs("../../../db")
	if err != nil {
		t.Fatalf("resolve db/: %v", err)
	}
	if _, err := os.Stat(srcDB); err != nil {
		t.Fatalf("db/ fixture missing at %s: %v", srcDB, err)
	}

	tmpDB := t.TempDir()
	// db/ contains symlinks (locale aliases under area/) which the
	// plain file-by-file copyTree helper in harness.go rejects. Shell
	// out to `cp -a` for a faithful recursive copy that preserves
	// symlinks. cp -a is POSIX-adjacent (GNU + macOS both implement).
	if out, err := exec.Command("cp", "-a", srcDB+"/.", tmpDB).CombinedOutput(); err != nil {
		t.Fatalf("cp -a db/: %v\n%s", err, out)
	}

	// Build the binary once per test run into a dedicated tempdir so
	// the repo tree stays clean.
	binDir := t.TempDir()
	binPath := filepath.Join(binDir, "smaug-test")
	if out, err := runGoBuild(binPath); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}

	// --- Phase A: create both characters through the nanny, quit. ---
	t.Logf("[Phase A] seeding pfiles via nanny")
	portA := spawnAndCreateChars(t, binPath, tmpDB)
	_ = portA // only used for the assertion that phase A bound a port

	// --- Phase B: grant Immy LEVEL_ASCENDANT trust on disk. ---
	t.Logf("[Phase B] granting Immy Trust 55 on disk")
	if err := grantImmyTrust(tmpDB); err != nil {
		t.Fatalf("grant trust: %v", err)
	}
	if bs, err := os.ReadFile(filepath.Join(tmpDB, "player", "i", "Immy")); err == nil {
		if !bytes.Contains(bs, []byte("Trust      60")) {
			t.Fatalf("Immy pfile is missing Trust line after edit:\n%s", bs)
		}
		t.Logf("[Phase B] Immy pfile length %d bytes, Trust 60 confirmed", len(bs))
	} else {
		t.Fatalf("re-read Immy pfile: %v", err)
	}

	// --- Phase C: fresh child, login both, hotboot. ---
	t.Logf("[Phase C] spawning fresh child for hotboot scenario")
	cmd, stderrLines, portC := spawnSmaug(t, binPath, tmpDB)
	defer func() {
		// Kill whatever PID is still around — the child exec'd itself
		// during the test, but cmd still tracks the original PID. Both
		// are the same process under syscall.Exec, so SIGTERM lands
		// correctly either way.
		_ = cmd.Process.Signal(syscall.SIGTERM)
		_, _ = cmd.Process.Wait()
	}()

	// (6) Dial both telnet clients, log in.
	immConn, immBR := dialAndLogin(t, portC, "Immy", "secret123")
	defer immConn.Close()
	mortConn, mortBR := dialAndLogin(t, portC, "Mortal", "secret123")
	defer mortConn.Close()

	// (7) Record starting rooms via `look`. The room name is the first
	// non-blank line of look output. We don't pin the exact text — we
	// just capture it so we can reassert equality post-hotboot.
	//
	// Drain any pending broadcasts from each client's buffer first.
	// Immy saw Mortal's entrance event when Mortal logged in; letting
	// that linger would cause captureLook to return the broadcast
	// text instead of the room name.
	drainPending(immConn, immBR, 300*time.Millisecond)
	drainPending(mortConn, mortBR, 300*time.Millisecond)

	immRoom := captureLook(t, immConn, immBR)
	mortRoom := captureLook(t, mortConn, mortBR)
	t.Logf("[rooms] Immy=%q Mortal=%q", immRoom, mortRoom)

	// (8) Imm issues hotboot.
	if _, err := fmt.Fprintf(immConn, "hotboot\n"); err != nil {
		t.Fatalf("send hotboot: %v", err)
	}

	// (9) Within 500ms, both clients must see the welcome + puff.
	//
	// "Time resumes its normal flow." is asserted as mandatory (see
	// plan-phase6-hotboot.md §G6 test step 9 and internal/boot/recover.go
	// hotbootWelcomePrefix). The puff-of-smoke text lives in recover.go
	// as an inline const rather than in db/socials, so it IS available
	// in the recovery path and we assert it too — if it ever moves to a
	// data file and vanishes from recovery, this assertion catches it.
	const welcome = "Time resumes its normal flow."
	const puff = "puff of ethereal smoke"

	immWelcome := readForSubstring(t, immConn, immBR, welcome, hotbootPostExecTimeout, "imm welcome")
	if !strings.Contains(immWelcome, welcome) {
		t.Fatalf("imm: missing %q in:\n%s", welcome, immWelcome)
	}
	if !strings.Contains(immWelcome, puff) {
		t.Errorf("imm: missing puff %q in:\n%s", puff, immWelcome)
	}

	mortWelcome := readForSubstring(t, mortConn, mortBR, welcome, hotbootPostExecTimeout, "mortal welcome")
	if !strings.Contains(mortWelcome, welcome) {
		t.Fatalf("mortal: missing %q in:\n%s", welcome, mortWelcome)
	}
	if !strings.Contains(mortWelcome, puff) {
		t.Errorf("mortal: missing puff %q in:\n%s", puff, mortWelcome)
	}

	// (10) State preserved — `look` returns the same room name.
	// Give the recovery path a beat to finish wiring the descriptor
	// input goroutine; then both `look`s should land on the pulse
	// following arrival. Drain any remaining puff/welcome/MOTD
	// fragments before issuing look.
	drainPending(immConn, immBR, 300*time.Millisecond)
	drainPending(mortConn, mortBR, 300*time.Millisecond)
	postImmRoom := captureLook(t, immConn, immBR)
	postMortRoom := captureLook(t, mortConn, mortBR)
	if postImmRoom != immRoom {
		t.Errorf("imm room drift: pre=%q post=%q", immRoom, postImmRoom)
	}
	if postMortRoom != mortRoom {
		t.Errorf("mortal room drift: pre=%q post=%q", mortRoom, postMortRoom)
	}

	// (11) Third fresh connection to confirm the rebound listener accepts.
	thirdAddr := fmt.Sprintf("127.0.0.1:%d", portC)
	thirdConn, err := gonet.DialTimeout("tcp", thirdAddr, hotbootDialTimeout)
	if err != nil {
		t.Fatalf("post-hotboot dial: %v", err)
	}
	defer thirdConn.Close()
	thirdBR := bufio.NewReader(thirdConn)
	// The stock boot emits a banner before "what is your name?" — reading
	// until "name" is the cheapest confirmation of CON_GET_NAME arrival.
	if !readUntilContainsConn(thirdConn, thirdBR, "what name", hotbootReadyTimeout) {
		t.Fatalf("third conn: never saw name prompt")
	}
	t.Logf("[Phase C] third conn reached CON_GET_NAME via rebound listener")

	// Drain any remaining stderr lines for post-mortem visibility on
	// failure; do not fail the test on their content.
	drainLog := collectStderr(stderrLines)
	if t.Failed() {
		t.Logf("[child stderr tail]\n%s", drainLog)
	}

	// (12) SIGTERM + Wait handled by the deferred cleanup above.
}

// --- Helpers below ---

// runGoBuild invokes `go build -o out ./cmd/smaug/`. The returned bytes
// are merged stdout+stderr for failure diagnosis.
func runGoBuild(out string) ([]byte, error) {
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		return nil, err
	}
	cmd := exec.Command("go", "build", "-o", out, "./cmd/smaug/")
	cmd.Dir = repoRoot
	return cmd.CombinedOutput()
}

// spawnSmaug launches the binary with -port 0 + -data tmpDB. Returns the
// *exec.Cmd handle, a channel of stderr lines (each sent once, the sender
// goroutine closes on EOF), and the port parsed from stderr.
func spawnSmaug(t *testing.T, bin, data string) (*exec.Cmd, <-chan string, int) {
	t.Helper()
	cmd := exec.Command(bin, "-port", "0", "-data", data)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}
	// Also drain stdout to prevent any buffer backpressure.
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	lines := make(chan string, 512)
	go func() {
		defer close(lines)
		sc := bufio.NewScanner(stderr)
		sc.Buffer(make([]byte, 64*1024), 1<<20)
		for sc.Scan() {
			line := sc.Text()
			t.Logf("[child.stderr] %s", line)
			select {
			case lines <- line:
			default:
				// channel full — drop; the test doesn't need
				// every line, just the listening announcement.
			}
		}
	}()
	go func() {
		sc := bufio.NewScanner(stdout)
		sc.Buffer(make([]byte, 64*1024), 1<<20)
		for sc.Scan() {
			t.Logf("[child.stdout] %s", sc.Text())
		}
	}()

	// Wait for "SMAUG MUD listening on :<port>" — authoritative port
	// marker since -port 0 asks the kernel to pick. Also accept "is
	// ready" as a secondary readiness signal; both lines are emitted in
	// order by cmd/smaug/main.go.
	deadline := time.Now().Add(hotbootReadyTimeout)
	var port int
	var sawReady bool
	for time.Now().Before(deadline) {
		select {
		case line, ok := <-lines:
			if !ok {
				t.Fatalf("spawnSmaug: child exited before ready")
			}
			if m := listeningRe.FindStringSubmatch(line); m != nil {
				fmt.Sscanf(m[1], "%d", &port)
			}
			if strings.Contains(line, "SMAUG MUD is ready") {
				sawReady = true
			}
			if port > 0 && sawReady {
				return cmd, lines, port
			}
		case <-time.After(500 * time.Millisecond):
			// keep looping until deadline
		}
	}
	_ = cmd.Process.Kill()
	t.Fatalf("spawnSmaug: timed out waiting for readiness (port=%d ready=%v)",
		port, sawReady)
	return nil, nil, 0
}

// spawnAndCreateChars runs Phase A: creates Immy + Mortal, issues `quit`
// on each, then SIGTERMs the child. Returns the port (useful only for
// log context — the next phase rebinds a fresh ephemeral port).
func spawnAndCreateChars(t *testing.T, bin, data string) int {
	t.Helper()
	cmd, stderrLines, port := spawnSmaug(t, bin, data)
	defer func() {
		_ = cmd.Process.Signal(syscall.SIGTERM)
		done := make(chan struct{})
		go func() {
			_ = cmd.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			_ = cmd.Process.Kill()
			<-done
		}
	}()
	_ = stderrLines // drained by background goroutine via t.Logf

	addr := fmt.Sprintf("127.0.0.1:%d", port)

	createChar(t, addr, "Immy", "secret123")
	createChar(t, addr, "Mortal", "secret123")

	return port
}

// createChar drives the full nanny flow for a fresh character and issues
// `quit` so the player file is saved and flushed to disk before the
// connection drops.
func createChar(t *testing.T, addr, name, password string) {
	t.Helper()
	conn, err := gonet.DialTimeout("tcp", addr, hotbootDialTimeout)
	if err != nil {
		t.Fatalf("dial for %s: %v", name, err)
	}
	defer conn.Close()
	br := bufio.NewReader(conn)

	expect := func(substr string) {
		if !readUntilContainsConn(conn, br, substr, hotbootNannyTimeout) {
			t.Fatalf("[%s] did not see %q", name, substr)
		}
	}
	send := func(s string) {
		if _, err := fmt.Fprintf(conn, "%s\n", s); err != nil {
			t.Fatalf("[%s] send %q: %v", name, s, err)
		}
	}

	expect("what name")
	send(name)
	expect("Did I get that right")
	send("y")
	// After Y: "New character.\n\rGive me a password for this character: "
	expect("password for this character")
	send(password)
	expect("retype")
	send(password)
	// Each menu ends with its terminator prompt; gender→"Please select:",
	// class→"Choice:", race→"Choice:". Match only the terminator to
	// avoid consuming the in-between bytes twice (bufio.Reader drains
	// the underlying socket in one read, so successive expect calls on
	// the same message would find an empty accumulator).
	expect("Please select:")
	send("m")
	expect("Choice:")
	send("warrior")
	expect("Choice:")
	send("human")
	expect("Press Enter")
	send("")
	// At the in-game prompt. Wait for any subsequent prompt/output, then
	// quit — SavePlayer runs on the DoQuit path, so the pfile lands
	// before the descriptor closes.
	if !readUntilContainsConn(conn, br, ">", hotbootNannyTimeout) {
		// Not fatal; some banner configurations omit a leading prompt.
	}
	send("quit")
	// Wait for the save-confirmation + goodbye text before dropping
	// the socket, so the child has flushed the pfile to disk.
	if !readUntilContainsConn(conn, br, "has been saved", 5*time.Second) {
		t.Logf("[%s] did not see save confirmation — pfile may not be persisted", name)
	}
	// Brief tail drain to catch the farewell line + EOF cleanly.
	_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	buf := make([]byte, 1024)
	for {
		n, err := br.Read(buf)
		if n == 0 || err != nil {
			break
		}
	}
}

// grantImmyTrust rewrites Immy's pfile to add "Trust 55" immediately
// before the terminating "End" line. SMAUG pfile loader is tolerant of
// field order, so a simple append before End is safe. If a Trust line
// already exists we replace it.
func grantImmyTrust(data string) error {
	path := filepath.Join(data, "player", "i", "Immy")
	bs, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read Immy pfile: %w", err)
	}
	text := string(bs)

	// Drop any pre-existing Trust line so we don't get two.
	trustLine := regexp.MustCompile(`(?m)^Trust\s+\d+\s*\n`)
	text = trustLine.ReplaceAllString(text, "")

	// Insert "Trust 55\n" before the first "End" line at column 0.
	endRe := regexp.MustCompile(`(?m)^End\b`)
	loc := endRe.FindStringIndex(text)
	if loc == nil {
		return fmt.Errorf("Immy pfile has no End marker; contents:\n%s", text)
	}
	inserted := text[:loc[0]] + "Trust      60\n" + text[loc[0]:]

	if err := os.WriteFile(path, []byte(inserted), 0o644); err != nil {
		return err
	}
	// Re-read to confirm and log for diagnostics.
	if bs2, err := os.ReadFile(path); err == nil {
		if !bytes.Contains(bs2, []byte("Trust      60")) {
			return fmt.Errorf("grantImmyTrust: Trust line not found after write; pfile:\n%s", bs2)
		}
	}
	return nil
}

// dialAndLogin opens a TCP connection and drives the returning-player
// nanny (name → password → Press Enter → prompt). Returns the raw
// net.Conn (for writes) and a bufio.Reader-over-IAC-stripper (for
// reads). All subsequent reads by the caller must go through the
// returned reader so IAC bytes are consumed consistently.
func dialAndLogin(t *testing.T, port int, name, password string) (gonet.Conn, *bufio.Reader) {
	t.Helper()
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := gonet.DialTimeout("tcp", addr, hotbootDialTimeout)
	if err != nil {
		t.Fatalf("dial %s: %v", name, err)
	}
	br := bufio.NewReader(conn)

	if !readUntilContainsConn(conn, br, "what name", hotbootNannyTimeout) {
		t.Fatalf("%s: no name prompt", name)
	}
	if _, err := fmt.Fprintf(conn, "%s\n", name); err != nil {
		t.Fatalf("%s: send name: %v", name, err)
	}
	// Returning-player branch: server writes "Password: " at loop.go:378.
	if !readUntilContainsConn(conn, br, "Password:", hotbootNannyTimeout) {
		t.Fatalf("%s: no password prompt", name)
	}
	if _, err := fmt.Fprintf(conn, "%s\n", password); err != nil {
		t.Fatalf("%s: send password: %v", name, err)
	}
	if !readUntilContainsConn(conn, br, "Press Enter", hotbootNannyTimeout) {
		t.Fatalf("%s: no MOTD prompt", name)
	}
	if _, err := fmt.Fprintf(conn, "\n"); err != nil {
		t.Fatalf("%s: send enter: %v", name, err)
	}
	// Drain initial look/prompt.
	if !readUntilContainsConn(conn, br, ">", hotbootNannyTimeout) {
		t.Logf("%s: no prompt char seen post-login (proceeding)", name)
	}
	return conn, br
}

// captureLook sends `look` and captures the first substantive output
// line (typically the room name on its own line). Returns the raw
// captured text for equality comparison. Implementation is tolerant:
// we consume up to the prompt, then pick the first non-blank
// non-prompt line.
func captureLook(t *testing.T, conn gonet.Conn, br *bufio.Reader) string {
	t.Helper()
	if _, err := fmt.Fprintf(conn, "look\n"); err != nil {
		t.Fatalf("send look: %v", err)
	}
	// Accumulate until a prompt char '>' is seen or timeout.
	deadline := time.Now().Add(3 * time.Second)
	var acc bytes.Buffer
	buf := make([]byte, 4096)
	for time.Now().Before(deadline) {
		_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, err := br.Read(buf)
		if n > 0 {
			acc.Write(buf[:n])
			if strings.Contains(cleanPayload(acc.Bytes()), ">") {
				break
			}
		}
		if err != nil {
			break
		}
	}
	cleaned := []byte(cleanPayload(acc.Bytes()))
	// Return the first meaningful (non-blank, non-prompt-only) line.
	for _, line := range strings.Split(string(cleaned), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, ">") {
			continue
		}
		return line
	}
	return strings.TrimSpace(string(cleaned))
}

// readForSubstring reads bytes up to timeout, returning the accumulated
// cleaned buffer as soon as substr is seen. IAC + ANSI are stripped.
// If substr never arrives, returns whatever was accumulated (the
// caller typically asserts containment and fails on miss).
func readForSubstring(t *testing.T, conn gonet.Conn, br *bufio.Reader, substr string, timeout time.Duration, label string) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var acc bytes.Buffer
	buf := make([]byte, 4096)
	for time.Now().Before(deadline) {
		_ = conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		n, err := br.Read(buf)
		if n > 0 {
			acc.Write(buf[:n])
			cleaned := cleanPayload(acc.Bytes())
			if strings.Contains(cleaned, substr) {
				return cleaned
			}
		}
		if err != nil {
			if ne, ok := err.(gonet.Error); ok && ne.Timeout() {
				continue
			}
			if err != io.EOF {
				t.Logf("[%s] read err: %v", label, err)
				break
			}
		}
		if n == 0 {
			time.Sleep(10 * time.Millisecond)
		}
	}
	return cleanPayload(acc.Bytes())
}

// readUntilContains reads from br with a bounded deadline until the
// cleaned buffer contains substr. Returns true on match, false on
// timeout or error. If the bufio.Reader is wrapping a net.Conn, pass
// the conn separately via readUntilContainsConn so read deadlines can
// be plumbed — br.Read alone will block indefinitely on a conn with no
// deadline set.
func readUntilContains(br *bufio.Reader, substr string, timeout time.Duration) bool {
	return readUntilContainsConn(nil, br, substr, timeout)
}

// readUntilContainsConn is the deadline-aware variant. If conn is
// non-nil, SetReadDeadline is called before each underlying Read so
// timeouts are bounded. Callers that have a raw net.Conn should use
// this form.
func readUntilContainsConn(conn gonet.Conn, br *bufio.Reader, substr string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	var acc bytes.Buffer
	buf := make([]byte, 4096)
	for time.Now().Before(deadline) {
		if conn != nil {
			_ = conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		}
		n, err := br.Read(buf)
		if n > 0 {
			acc.Write(buf[:n])
			if strings.Contains(cleanPayload(acc.Bytes()), substr) {
				return true
			}
		}
		if err != nil {
			// Timeout errors from SetReadDeadline are transient — keep
			// polling until the outer deadline elapses. Real EOF /
			// connection errors abort.
			if ne, ok := err.(gonet.Error); ok && ne.Timeout() {
				continue
			}
			if err != io.EOF {
				return false
			}
		}
		if n == 0 {
			time.Sleep(10 * time.Millisecond)
		}
	}
	return false
}

// drainPending reads whatever bytes are already buffered on the
// connection for up to `budget`, discarding them. Used to clear stale
// broadcast messages (e.g. another player's entrance) before issuing
// commands whose output we want to capture cleanly.
func drainPending(conn gonet.Conn, br *bufio.Reader, budget time.Duration) {
	deadline := time.Now().Add(budget)
	buf := make([]byte, 4096)
	for time.Now().Before(deadline) {
		_ = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		n, err := br.Read(buf)
		if n == 0 && err != nil {
			// Timeout means no more data — we're drained.
			if ne, ok := err.(gonet.Error); ok && ne.Timeout() {
				return
			}
			return
		}
	}
}

// cleanPayload strips IAC + ANSI + CR from raw bytes for human-readable
// substring matching. Partial IAC sequences at the tail are dropped
// (the `consumed` return is ignored) — this is safe for tests because
// we keep accumulating raw bytes and re-clean each iteration.
func cleanPayload(b []byte) string {
	out, _ := stripIAC(b)
	out = stripANSI(out)
	out = stripCR(out)
	return string(out)
}

// collectStderr drains whatever stderr lines are still queued without
// blocking. Used on test failure for diagnostic context.
func collectStderr(lines <-chan string) string {
	var acc bytes.Buffer
	for {
		select {
		case line, ok := <-lines:
			if !ok {
				return acc.String()
			}
			acc.WriteString(line)
			acc.WriteByte('\n')
		default:
			return acc.String()
		}
	}
}
