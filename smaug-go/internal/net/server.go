package net

import (
	"bufio"
	"log"
	gonet "net"
	"sync"

	"github.com/eilidhmae/smaug/internal/types"
)

// Server listens for incoming TCP connections and hands them
// to the game loop via the Incoming channel.
type Server struct {
	listener gonet.Listener
	Incoming chan *types.DescriptorData
	done     chan struct{}
	// doneMu guards the close/recreate cycle on done across Stop and
	// the hotboot pause/resume seam so concurrent callers don't
	// double-close. Normal steady-state operation never contends.
	doneMu sync.Mutex
	// acceptWG tracks the accept goroutine so ResumeFromPause can wait
	// for the prior acceptLoop to exit before rewriting s.listener.
	// Without this, the rewrite and the old goroutine's s.listener read
	// race (go test -race catches it deterministically under
	// TestResumeFromPause_RestartsAccept).
	acceptWG sync.WaitGroup
}

// NewServer creates a new Server ready to be started.
func NewServer() *Server {
	return &Server{
		Incoming: make(chan *types.DescriptorData, 50),
		done:     make(chan struct{}),
	}
}

// StartOnListener begins accepting connections on the supplied listener.
// Use this when the caller already bound the port (tests, graceful rebinds).
// Ownership of ln transfers to the server — Stop will close it.
func (s *Server) StartOnListener(ln gonet.Listener) error {
	s.listener = ln
	s.acceptWG.Add(1)
	go s.acceptLoop()
	return nil
}

// Addr returns the listener's network address, or nil if the server has
// not been started yet. Used by hotboot recovery tests to confirm the
// inherited FD maps to the expected port.
func (s *Server) Addr() gonet.Addr {
	if s.listener == nil {
		return nil
	}
	return s.listener.Addr()
}

// Stop shuts down the server by closing the done channel and the listener.
// Idempotent: a second Stop (including after PauseForHotboot) is a no-op.
func (s *Server) Stop() {
	s.doneMu.Lock()
	select {
	case <-s.done:
		// already closed (hotboot pause or earlier Stop)
	default:
		close(s.done)
	}
	s.doneMu.Unlock()
	if s.listener != nil {
		s.listener.Close()
	}
}

// acceptLoop runs in its own goroutine, accepting new TCP connections
// until the server is stopped.
func (s *Server) acceptLoop() {
	defer s.acceptWG.Done()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.done:
				// Server is shutting down; expected error.
				return
			default:
				log.Printf("[net] Accept error: %v", err)
				continue
			}
		}

		desc := types.NewDescriptor(conn)
		desc.ColorFunc = ProcessColors
		log.Printf("[net] New connection from %s", desc.Host)

		go s.readLoop(desc, conn)

		select {
		case s.Incoming <- desc:
		case <-s.done:
			conn.Close()
			return
		}
	}
}

// readLoop reads line-based input from a connection, strips telnet IAC
// sequences, and feeds clean lines into the descriptor's InputQueue.
func (s *Server) readLoop(desc *types.DescriptorData, conn gonet.Conn) {
	defer close(desc.InputQueue)

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1024), 1024)
	for scanner.Scan() {
		select {
		case <-s.done:
			return
		default:
		}

		line := stripTelnetIAC(scanner.Bytes())
		// Strip trailing \r that may remain after the \n split.
		line = stripCR(line)
		trimmed := string(line)

		select {
		case desc.InputQueue <- trimmed:
		case <-s.done:
			return
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[net] Read error from %s: %v", desc.Host, err)
	}
}

// stripTelnetIAC removes in-band telnet IAC sequences from raw input.
// IAC (0xFF) is followed by a command byte; WILL/WONT/DO/DONT (0xFB-0xFE)
// have an additional option byte, while other two-byte commands (e.g. NOP,
// GA) consist of just IAC + command.
func stripTelnetIAC(data []byte) []byte {
	const iac byte = 255

	out := make([]byte, 0, len(data))
	i := 0
	for i < len(data) {
		if data[i] == iac && i+1 < len(data) {
			cmd := data[i+1]
			switch {
			case cmd >= 251 && cmd <= 254:
				// WILL (251), WONT (252), DO (253), DONT (254): 3 bytes total.
				if i+2 < len(data) {
					i += 3
				} else {
					i = len(data) // truncated sequence, skip to end
				}
			case cmd == iac:
				// Escaped 0xFF – keep one literal 0xFF.
				out = append(out, iac)
				i += 2
			default:
				// Two-byte command (NOP, GA, etc.).
				i += 2
			}
			continue
		}
		out = append(out, data[i])
		i++
	}
	return out
}

// stripCR removes all carriage-return bytes from a slice.
func stripCR(data []byte) []byte {
	out := make([]byte, 0, len(data))
	for _, b := range data {
		if b != '\r' {
			out = append(out, b)
		}
	}
	return out
}
