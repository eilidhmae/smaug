//go:build windows

// Hotboot pause/resume Windows stub. Phase-6 hotboot is Linux/macOS only
// (plan-phase6-hotboot.md §Platform — syscall.Exec absent and (*net.TCPConn).File()
// returns an FD "not usable on other processes" per Go stdlib docs).

package net

import (
	"fmt"
	"os"

	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
)

// PauseForHotboot is unsupported on Windows.
func (s *Server) PauseForHotboot(descriptors []*types.DescriptorData, port int) (*os.File, []*os.File, []persist.HotbootSession, error) {
	return nil, nil, nil, fmt.Errorf("hotboot not supported on this platform")
}

// ResumeFromPause is unsupported on Windows.
func (s *Server) ResumeFromPause(lnFile *os.File) error {
	return fmt.Errorf("hotboot not supported on this platform")
}
