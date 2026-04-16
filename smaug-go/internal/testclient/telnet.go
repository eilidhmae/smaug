package testclient

// Telnet protocol byte constants. See RFC 854 and RFC 855.
const (
	iacByte  byte = 255 // IAC  — interpret-as-command
	dontByte byte = 254
	doByte   byte = 253
	wontByte byte = 252
	willByte byte = 251
	sbByte   byte = 250 // SB   — subnegotiation begin
	seByte   byte = 240 // SE   — subnegotiation end
)

// stripIAC consumes telnet IAC sequences from a byte slice.
//
// Handles:
//   - IAC IAC (0xFF 0xFF)           → literal 0xFF in output
//   - IAC WILL/WONT/DO/DONT option  → 3-byte sequence, stripped
//   - IAC SB ... IAC SE             → subnegotiation, stripped
//   - IAC <other cmd>               → 2-byte sequence, stripped
//
// Returns the cleaned slice and the number of bytes consumed from input.
// If the input ends mid-sequence (e.g. bare trailing IAC, or an unclosed
// SB), the cleaned slice contains only the fully-processed prefix and
// `consumed` reflects that prefix length, so the caller can preserve the
// unconsumed tail and concatenate it with the next read.
func stripIAC(data []byte) (clean []byte, consumed int) {
	out := make([]byte, 0, len(data))
	i := 0
	for i < len(data) {
		if data[i] != iacByte {
			out = append(out, data[i])
			i++
			continue
		}
		// data[i] == IAC. Need at least one more byte to decide.
		if i+1 >= len(data) {
			// Truncated: bare IAC at end. Preserve for next read.
			return out, i
		}
		cmd := data[i+1]
		switch {
		case cmd == iacByte:
			// IAC IAC → literal 0xFF.
			out = append(out, iacByte)
			i += 2
		case cmd >= willByte && cmd <= dontByte:
			// WILL/WONT/DO/DONT: need 3 bytes total.
			if i+2 >= len(data) {
				// Truncated: option byte missing. Preserve from IAC.
				return out, i
			}
			i += 3
		case cmd == sbByte:
			// Subnegotiation: scan for IAC SE.
			j := i + 2
			foundEnd := false
			for j+1 < len(data) {
				if data[j] == iacByte && data[j+1] == seByte {
					foundEnd = true
					j += 2
					break
				}
				j++
			}
			if !foundEnd {
				// Unterminated SB. Preserve the whole SB sequence for
				// next read.
				return out, i
			}
			i = j
		default:
			// Two-byte command (NOP, GA, etc.): strip both bytes.
			i += 2
		}
	}
	return out, i
}
