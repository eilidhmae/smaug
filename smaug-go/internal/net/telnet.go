package net

// Telnet protocol constants.
const (
	IAC  byte = 255
	DONT byte = 254
	DO   byte = 253
	WONT byte = 252
	WILL byte = 251
	SB   byte = 250
	SE   byte = 240
	GA   byte = 249

	// Telnet options
	TELOPT_ECHO    byte = 1
	TELOPT_TTYPE   byte = 24
	TELOPT_NAWS    byte = 31
	TELOPT_COMPRESS2 byte = 86  // MCCP2
	TELOPT_MSDP    byte = 69
	TELOPT_MSSP    byte = 70
)

// TelnetNeg builds a 3-byte telnet negotiation sequence.
func TelnetNeg(verb byte, option byte) []byte {
	return []byte{IAC, verb, option}
}

// TelnetSubneg builds a subnegotiation sequence: IAC SB <option> <data...> IAC SE
func TelnetSubneg(option byte, data []byte) []byte {
	buf := make([]byte, 0, 5+len(data))
	buf = append(buf, IAC, SB, option)
	buf = append(buf, data...)
	buf = append(buf, IAC, SE)
	return buf
}
