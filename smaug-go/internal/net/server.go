package net

import (
	"bufio"
	"fmt"
	"log"
	gonet "net"

	"github.com/eilidhmae/smaug/internal/types"
)

// Server listens for incoming TCP connections and hands them
// to the game loop via the Incoming channel.
type Server struct {
	listener gonet.Listener
	Incoming chan *types.DescriptorData
	done     chan struct{}
}

// NewServer creates a new Server ready to be started.
func NewServer() *Server {
	return &Server{
		Incoming: make(chan *types.DescriptorData, 50),
		done:     make(chan struct{}),
	}
}

// Start begins listening on the given TCP port and accepting connections.
func (s *Server) Start(port int) error {
	var err error
	s.listener, err = gonet.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", port, err)
	}
	log.Printf("[net] Listening on port %d", port)

	go s.acceptLoop()
	return nil
}

// Stop shuts down the server by closing the done channel and the listener.
func (s *Server) Stop() {
	close(s.done)
	if s.listener != nil {
		s.listener.Close()
	}
}

// acceptLoop runs in its own goroutine, accepting new TCP connections
// until the server is stopped.
func (s *Server) acceptLoop() {
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
