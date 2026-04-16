package testclient

// stripANSI removes ANSI CSI escape sequences (ESC [ … final-byte) and OSC
// sequences (ESC ] … BEL | ESC \). Simple byte-oriented state machine; no
// dependencies. Non-escape bytes are preserved verbatim.
//
// Truncation policy: a bare trailing ESC (or an unclosed CSI/OSC that runs
// off the end of input) is DROPPED. MUD output doesn't split escape
// sequences across socket reads in practice, so we accept this rather than
// carrying partial-escape state across calls.
func stripANSI(data []byte) []byte {
	const (
		stNormal = iota
		stEsc       // just saw ESC
		stCSI       // inside ESC [ … (ends at byte 0x40-0x7E)
		stOSC       // inside ESC ] … (ends at BEL or ESC \)
		stOSCEsc    // inside OSC, just saw ESC (waiting for backslash)
	)

	out := make([]byte, 0, len(data))
	state := stNormal
	for i := 0; i < len(data); i++ {
		b := data[i]
		switch state {
		case stNormal:
			if b == 0x1b { // ESC
				state = stEsc
				continue
			}
			out = append(out, b)
		case stEsc:
			switch b {
			case '[':
				state = stCSI
			case ']':
				state = stOSC
			default:
				// Two-byte escape (ESC + intermediate like ESC 7, ESC =) —
				// swallow the follow-up byte and return to normal.
				state = stNormal
			}
		case stCSI:
			// CSI final byte is 0x40-0x7E (@–~).
			if b >= 0x40 && b <= 0x7E {
				state = stNormal
			}
			// Else: parameter/intermediate byte — keep consuming.
		case stOSC:
			if b == 0x07 { // BEL terminates OSC
				state = stNormal
			} else if b == 0x1b { // ESC — possibly start of ESC \ terminator
				state = stOSCEsc
			}
		case stOSCEsc:
			// After ESC inside OSC: anything (typically '\\') ends OSC.
			state = stNormal
		}
	}
	return out
}
