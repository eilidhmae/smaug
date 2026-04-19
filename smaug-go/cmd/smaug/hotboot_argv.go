//go:build !windows

// Hotboot child-process argv parsing. The parent `do_hotboot` re-execs
// this binary with `--hotboot-recover <ln_fd> <d1_fd>:<idx> ...` appended
// to the normal arg list. `main.go` calls ParseHotbootArgv before
// flag.Parse to branch the boot path, then StripHotbootArgv removes the
// hotboot tokens so `flag` can still parse -port / -data.
//
// Contract (plan-phase6-hotboot.md §G5):
//
//	(nil, nil) — flag absent; caller takes the normal-boot path.
//	(v,   nil) — flag present and valid; caller calls boot.BootRecover.
//	(nil, err) — flag present but malformed; caller log.Fatals.
//
// Accepted payload: one or more non-dash tokens. Scanning stops at the
// next `-`-prefixed token OR end-of-args, so it doesn't eat a trailing
// `-port 4000`. Required shape: <ln_fd> <fd>:<idx> [<fd>:<idx> ...].

package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/boot"
)

// HotbootArgv carries the parsed --hotboot-recover arguments.
type HotbootArgv struct {
	LnFD     uintptr
	Sessions []boot.FDSession
}

const hotbootRecoverFlag = "--hotboot-recover"

// ParseHotbootArgv scans args for the hotboot-recover flag and returns
// the parsed payload. args should be os.Args[1:] (no program name).
func ParseHotbootArgv(args []string) (*HotbootArgv, error) {
	flagIdx := findHotbootFlag(args)
	if flagIdx < 0 {
		return nil, nil
	}

	start := flagIdx + 1
	end := start
	for end < len(args) && !strings.HasPrefix(args[end], "-") {
		end++
	}
	payload := args[start:end]
	if len(payload) < 2 {
		return nil, fmt.Errorf("%s: expected <ln_fd> <d_fd>:<idx> [...], got %d arg(s)",
			hotbootRecoverFlag, len(payload))
	}

	lnFD64, err := strconv.ParseUint(payload[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%s: ln_fd %q: %v", hotbootRecoverFlag, payload[0], err)
	}
	if lnFD64 == 0 {
		// FD 0 is stdin — the listener can never be stdin after exec.
		return nil, fmt.Errorf("%s: ln_fd must be > 0 (got 0; stdin reserved)", hotbootRecoverFlag)
	}

	sessions := make([]boot.FDSession, 0, len(payload)-1)
	for _, tok := range payload[1:] {
		colon := strings.IndexByte(tok, ':')
		if colon < 0 {
			return nil, fmt.Errorf("%s: descriptor token %q missing ':idx'",
				hotbootRecoverFlag, tok)
		}
		fdStr, idxStr := tok[:colon], tok[colon+1:]
		fd64, err := strconv.ParseUint(fdStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%s: descriptor fd %q: %v", hotbootRecoverFlag, fdStr, err)
		}
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			return nil, fmt.Errorf("%s: descriptor idx %q: %v", hotbootRecoverFlag, idxStr, err)
		}
		sessions = append(sessions, boot.FDSession{FD: uintptr(fd64), SessionIdx: idx})
	}

	return &HotbootArgv{LnFD: uintptr(lnFD64), Sessions: sessions}, nil
}

// StripHotbootArgv returns a copy of args without the --hotboot-recover
// flag and its payload. Leaves everything else (including other flags
// and their values) intact so flag.Parse sees a clean arg list.
func StripHotbootArgv(args []string) []string {
	flagIdx := findHotbootFlag(args)
	if flagIdx < 0 {
		return args
	}
	end := flagIdx + 1
	for end < len(args) && !strings.HasPrefix(args[end], "-") {
		end++
	}
	out := make([]string, 0, len(args)-(end-flagIdx))
	out = append(out, args[:flagIdx]...)
	out = append(out, args[end:]...)
	return out
}

func findHotbootFlag(args []string) int {
	for i, tok := range args {
		if tok == hotbootRecoverFlag {
			return i
		}
	}
	return -1
}
