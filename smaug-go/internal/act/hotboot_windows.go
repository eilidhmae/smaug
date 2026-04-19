//go:build windows

// Hotboot Windows stub. Phase-6 hotboot is Linux/macOS only
// (plan-phase6-hotboot.md §Platform). syscall.Exec is absent on
// Windows and (*net.TCPConn).File() returns an FD "not usable on
// other processes" per Go stdlib docs.

package act

import (
	"os"

	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
)

// HotbootPauseFunc seam (unused on Windows — left nil).
var HotbootPauseFunc func(descriptors []*types.DescriptorData, port int) (*os.File, []*os.File, []persist.HotbootSession, error)

// HotbootResumeFunc seam (unused on Windows — left nil).
var HotbootResumeFunc func(lnFile *os.File) error

// HotbootPort is unused on Windows.
var HotbootPort int

// DoHotboot is a hard-fail stub on Windows.
func DoHotboot(ch *types.CharData, argument string) {
	if ch == nil {
		return
	}
	ch.Send("hotboot not supported on this platform\n\r")
}
