package net

import (
	"fmt"
)

// MSSP (Mud Server Status Protocol) constants.
const (
	MSSP_VAR byte = 1
	MSSP_VAL byte = 2
)

// MSSPInfo holds server information for MSSP reporting.
type MSSPInfo struct {
	Name    string
	Players int
	Uptime  int64
	Port    int
}

// BuildMSSPPayload builds the MSSP subnegotiation data.
func BuildMSSPPayload(info *MSSPInfo) []byte {
	var buf []byte

	addVar := func(name, value string) {
		buf = append(buf, MSSP_VAR)
		buf = append(buf, []byte(name)...)
		buf = append(buf, MSSP_VAL)
		buf = append(buf, []byte(value)...)
	}

	addVar("NAME", info.Name)
	addVar("PLAYERS", fmt.Sprintf("%d", info.Players))
	addVar("UPTIME", fmt.Sprintf("%d", info.Uptime))
	addVar("CODEBASE", "SMAUG 1.8 (Go)")
	addVar("FAMILY", "DikuMUD")
	addVar("PORT", fmt.Sprintf("%d", info.Port))
	addVar("HOSTNAME", "localhost")
	addVar("ANSI", "1")
	addVar("MCCP", "1")

	return TelnetSubneg(TELOPT_MSSP, buf)
}
