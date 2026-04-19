//go:build windows

// Hotboot recovery stub for Windows. The hotboot subsystem is Linux/
// macOS only (see plan-phase6-hotboot.md §Platform). This file exists
// so GOOS=windows builds still resolve the BootRecover + FDSession
// symbols that main.go's argv-branch dispatches to.

package boot

import (
	"fmt"

	"github.com/eilidhmae/smaug/internal/command"
	"github.com/eilidhmae/smaug/internal/game"
	smaugnet "github.com/eilidhmae/smaug/internal/net"
	"github.com/eilidhmae/smaug/internal/world"
)

// FDSession is the Windows-side signature of the Linux/macOS type.
// Present only so callers can reference boot.FDSession uniformly.
type FDSession struct {
	FD         uintptr
	SessionIdx int
}

// BootRecover is not supported on Windows — errors out.
func BootRecover(
	w *world.World,
	dataDir string,
	opts BootOpts,
	lnFD uintptr,
	descFDSessions []FDSession,
) (*command.Registry, *game.GameLoop, *smaugnet.Server, error) {
	return nil, nil, nil, fmt.Errorf("hotboot not supported on this platform")
}
