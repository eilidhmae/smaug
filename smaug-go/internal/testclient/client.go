package testclient

import (
	"bytes"
	"fmt"
	gonet "net"
	"regexp"
	"testing"
	"time"
)

// defaultPromptRe matches a trailing "> " at the end of the buffer.
var defaultPromptRe = regexp.MustCompile(`> $`)

// readScratchSize is the fillOnce read-buffer size. Large enough to absorb
// typical MUD server bursts in one Read without forcing extra syscalls.
const readScratchSize = 4096

// Client is a telnet-aware test client. Send writes go out raw; reads run
// through the IAC + ANSI strip pipeline (when those options are enabled)
// before being appended to the match buffer. Methods fatal the owning
// *testing.T on I/O errors or timeout, so scenarios stay linear.
type Client struct {
	conn      gonet.Conn
	buf       []byte // cleaned output accumulated but not yet matched
	rawTail   []byte // bytes of a partial IAC sequence carried over
	scratch   []byte // reusable read buffer (allocated once per client)
	promptRe  *regexp.Regexp
	t         *testing.T
	stripANSI bool
	stripIAC  bool
}

// Send writes line+"\n" to the server. Fatals on write error.
func (c *Client) Send(line string) {
	if c.t != nil {
		c.t.Helper()
	}
	if err := c.sendErr(line); err != nil {
		c.fatalf("Send %q: %v", line, err)
	}
}

// sendErr is the non-fatal variant of Send. Returns an error on I/O
// failure instead of invoking t.Fatalf. Used by the fatal wrapper and
// directly by tests that want to assert failure modes.
func (c *Client) sendErr(line string) error {
	_ = c.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	_, err := c.conn.Write([]byte(line + "\n"))
	return err
}

// ReadUntil reads bytes until the accumulated cleaned buffer contains
// substr (case-insensitive). Returns everything in the buffer up to and
// including the match; the matched content is consumed. Fatals on timeout
// or I/O error.
func (c *Client) ReadUntil(substr string, timeout time.Duration) string {
	if c.t != nil {
		c.t.Helper()
	}
	got, err := c.readUntilErr(substr, timeout)
	if err != nil {
		c.fatalf("ReadUntil %q: %v", substr, err)
	}
	return got
}

// readUntilErr is the non-fatal variant of ReadUntil. Returns an error on
// timeout or I/O failure instead of invoking t.Fatalf. Used by the fatal
// wrapper and directly by tests that want to assert failure modes.
func (c *Client) readUntilErr(substr string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	// Search on []byte throughout — avoids string(c.buf) allocation each loop
	// iteration. bytes.ToLower returns a fresh slice but that cost is
	// unavoidable (and identical to the old strings.ToLower path).
	lowerSub := bytes.ToLower([]byte(substr))

	for {
		if idx := bytes.Index(bytes.ToLower(c.buf), lowerSub); idx >= 0 {
			end := idx + len(substr)
			result := string(c.buf[:end])
			c.buf = c.buf[end:]
			return result, nil
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("timeout after %s; last %d bytes: %q",
				timeout, len(c.buf), truncate(string(c.buf), 500))
		}
		if err := c.fillOnce(deadline); err != nil {
			// Final check after error — maybe the closing read filled
			// us past the match.
			if idx := bytes.Index(bytes.ToLower(c.buf), lowerSub); idx >= 0 {
				end := idx + len(substr)
				result := string(c.buf[:end])
				c.buf = c.buf[end:]
				return result, nil
			}
			return "", fmt.Errorf("%w; buf %q", err, truncate(string(c.buf), 500))
		}
	}
}

// ReadMatch reads bytes until `re` matches. Returns FindStringSubmatch
// (whole match + groups). Consumes through the end of the match. Fatals
// on timeout or I/O error.
func (c *Client) ReadMatch(re *regexp.Regexp, timeout time.Duration) []string {
	if c.t != nil {
		c.t.Helper()
	}
	deadline := time.Now().Add(timeout)
	for {
		if idx := re.FindStringSubmatchIndex(string(c.buf)); idx != nil {
			groups := matchFromIndex(c.buf, idx)
			c.buf = c.buf[idx[1]:]
			return groups
		}
		if time.Now().After(deadline) {
			c.fatalf("ReadMatch %q: timeout; last %d bytes: %q",
				re.String(), len(c.buf), truncate(string(c.buf), 500))
			return nil
		}
		if err := c.fillOnce(deadline); err != nil {
			if idx := re.FindStringSubmatchIndex(string(c.buf)); idx != nil {
				groups := matchFromIndex(c.buf, idx)
				c.buf = c.buf[idx[1]:]
				return groups
			}
			c.fatalf("ReadMatch %q: %v; buf %q",
				re.String(), err, truncate(string(c.buf), 500))
			return nil
		}
	}
}

// matchFromIndex extracts capture-group strings from a buffer using index
// pairs returned by FindStringSubmatchIndex. Returns nil if idx is nil.
// Unmatched groups (start == -1) become empty strings, matching
// FindStringSubmatch semantics.
func matchFromIndex(buf []byte, idx []int) []string {
	if idx == nil {
		return nil
	}
	n := len(idx) / 2
	groups := make([]string, n)
	for i := 0; i < n; i++ {
		if idx[2*i] < 0 {
			continue
		}
		groups[i] = string(buf[idx[2*i]:idx[2*i+1]])
	}
	return groups
}

// ReadToPrompt reads until the prompt regex matches at the end of the
// buffer. Default prompt is "> " at end-of-buffer; override via WithPrompt.
// Fatals on timeout or I/O error.
func (c *Client) ReadToPrompt(timeout time.Duration) string {
	if c.t != nil {
		c.t.Helper()
	}
	re := c.promptRe
	if re == nil {
		re = defaultPromptRe
	}
	deadline := time.Now().Add(timeout)
	for {
		if loc := re.FindIndex(c.buf); loc != nil {
			end := loc[1]
			result := string(c.buf[:end])
			c.buf = c.buf[end:]
			return result
		}
		if time.Now().After(deadline) {
			c.fatalf("ReadToPrompt: timeout waiting for %q; last %d bytes: %q",
				re.String(), len(c.buf), truncate(string(c.buf), 500))
			return ""
		}
		if err := c.fillOnce(deadline); err != nil {
			if loc := re.FindIndex(c.buf); loc != nil {
				end := loc[1]
				result := string(c.buf[:end])
				c.buf = c.buf[end:]
				return result
			}
			c.fatalf("ReadToPrompt: %v; buf %q", err, truncate(string(c.buf), 500))
			return ""
		}
	}
}

// ReadFor drains all available output for the given duration.
func (c *Client) ReadFor(d time.Duration) string {
	if c.t != nil {
		c.t.Helper()
	}
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if err := c.fillOnce(deadline); err != nil {
			break
		}
	}
	result := string(c.buf)
	c.buf = c.buf[:0]
	return result
}

// WithPrompt overrides the prompt-detection regex for this client. An
// empty pattern restores the default "> " at end-of-buffer. Returns the
// client for chaining.
func (c *Client) WithPrompt(pattern string) *Client {
	if pattern == "" {
		c.promptRe = nil
		return c
	}
	c.promptRe = regexp.MustCompile(pattern)
	return c
}

// Close closes the underlying connection.
func (c *Client) Close() {
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

// fillOnce does one Read into the per-Client scratch buffer and appends
// the cleaned bytes to c.buf. Returns an error on timeout or other I/O
// failures. Small reads are fine — ReadUntil loops until its condition
// is met. The scratch buffer is allocated lazily on first use so zero-
// constructed Client{} instances (unit tests that never call fillOnce)
// stay allocation-free.
func (c *Client) fillOnce(deadline time.Time) error {
	if c.scratch == nil {
		c.scratch = make([]byte, readScratchSize)
	}
	if err := c.conn.SetReadDeadline(deadline); err != nil {
		return err
	}
	n, err := c.conn.Read(c.scratch)
	if n > 0 {
		raw := append(c.rawTail, c.scratch[:n]...)
		c.rawTail = nil
		if c.stripIAC {
			clean, consumed := stripIAC(raw)
			if consumed < len(raw) {
				c.rawTail = append(c.rawTail, raw[consumed:]...)
			}
			raw = clean
		}
		if c.stripANSI {
			raw = stripANSI(raw)
		}
		// Strip \r so matches against literal "\n" work regardless of
		// the server using "\n\r" or "\r\n".
		raw = stripCR(raw)
		c.buf = append(c.buf, raw...)
	}
	return err
}

func stripCR(data []byte) []byte {
	out := make([]byte, 0, len(data))
	for _, b := range data {
		if b != '\r' {
			out = append(out, b)
		}
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

func (c *Client) fatalf(format string, args ...any) {
	if c.t == nil {
		panic(fmt.Sprintf(format, args...))
	}
	c.t.Helper()
	c.t.Fatalf(format, args...)
}
