//go:build windows

// Windows stub for hotboot argv parsing. The hotboot subsystem is
// Linux/macOS only (see plan-phase6-hotboot.md §Platform). On Windows
// the presence of --hotboot-recover is itself a hard error.

package main

import (
	"fmt"

	"github.com/eilidhmae/smaug/internal/boot"
)

type HotbootArgv struct {
	LnFD     uintptr
	Sessions []boot.FDSession
}

const hotbootRecoverFlag = "--hotboot-recover"

func ParseHotbootArgv(args []string) (*HotbootArgv, error) {
	for _, tok := range args {
		if tok == hotbootRecoverFlag {
			return nil, fmt.Errorf("%s: hotboot not supported on this platform", hotbootRecoverFlag)
		}
	}
	return nil, nil
}

func StripHotbootArgv(args []string) []string { return args }
