package types

import (
	"fmt"
	"net"
	"sync"
)

// DescriptorData represents a network connection for one player.
// Maps to C struct descriptor_data (mud.h:932).
type DescriptorData struct {
	// Network connection
	Conn net.Conn

	// Associated character
	Character *CharData
	// Original character (for switch/snoop)
	Original *CharData

	// Snoop
	SnoopBy *DescriptorData

	// Host info
	Host string
	Port int

	// Connection state machine
	Connected int // CON_PLAYING, CON_GET_NAME, etc.

	// Idle tracking
	Idle  int
	Lines int

	// Screen dimensions
	ScrLen int

	// Whether we have a pending command
	FCommand bool

	// Input handling
	InputQueue chan string // buffered channel for commands from read goroutine
	InLast     string      // last input command
	Repeat     int         // repeat count

	// Output buffering
	outBuf []byte
	outMu  sync.Mutex

	// Pager
	pageBuf   []byte
	pagePoint int
	pageCmd   byte
	pageColor byte

	// User data during login
	User     string
	NewState int

	// Previous color sent
	PrevColor byte

	// Protocol state
	TelnetState *TelnetState

	// ColorFunc processes color codes in output before sending to the client.
	// Set by the server layer. If nil, output is sent as-is.
	ColorFunc func(text string, ansiEnabled bool) string
}

// TelnetState tracks telnet protocol negotiation.
type TelnetState struct {
	// MCCP
	MCCPEnabled bool

	// MSDP
	MSDPEnabled bool

	// MSSP
	MSSPEnabled bool

	// Terminal type
	TermType string

	// Screen size from NAWS
	Width  int
	Height int
}

// WriteToBuffer appends text to the output buffer.
func (d *DescriptorData) WriteToBuffer(text string) {
	d.outMu.Lock()
	d.outBuf = append(d.outBuf, []byte(text)...)
	d.outMu.Unlock()
}

// WriteToBufferf appends formatted text to the output buffer.
func (d *DescriptorData) WriteToBufferf(format string, args ...any) {
	d.WriteToBuffer(fmt.Sprintf(format, args...))
}

// FlushOutput sends the output buffer to the connection and clears it.
// If ColorFunc is set, color codes are processed before sending.
// Returns an error if the write fails.
func (d *DescriptorData) FlushOutput() error {
	d.outMu.Lock()
	if len(d.outBuf) == 0 {
		d.outMu.Unlock()
		return nil
	}
	buf := make([]byte, len(d.outBuf))
	copy(buf, d.outBuf)
	d.outBuf = d.outBuf[:0]
	d.outMu.Unlock()

	// Process color codes if a color function is set
	if d.ColorFunc != nil {
		ansi := d.Character == nil || d.Character.Act.IsSet(PLR_ANSI)
		processed := d.ColorFunc(string(buf), ansi)
		buf = []byte(processed)
	}

	_, err := d.Conn.Write(buf)
	return err
}

// HasOutput returns true if there is pending output.
func (d *DescriptorData) HasOutput() bool {
	d.outMu.Lock()
	defer d.outMu.Unlock()
	return len(d.outBuf) > 0
}

// NewDescriptor creates a new descriptor for a connection.
func NewDescriptor(conn net.Conn) *DescriptorData {
	host, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	return &DescriptorData{
		Conn:        conn,
		Host:        host,
		Connected:   int(CON_GET_NAME),
		InputQueue:  make(chan string, 100),
		ScrLen:      24,
		TelnetState: &TelnetState{Width: 80, Height: 24},
	}
}
