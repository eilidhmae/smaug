// Loader for db/system/noauction.dat — the vnum blacklist consulted
// by do_auction's start-new branch. Mirrors C's load path used at boot
// (C calls `load_auction_list` indirectly via init; the writer
// `save_noauctions` at src/act_wiz.c:11099-11118 emits one `%d\n` per
// vnum plus a literal `0\n` terminator).
//
// This Phase-6 port is LOAD-ONLY; the immortal `do_noauction` admin
// CRUD at src/act_wiz.c:11120-11174 is tracked as a follow-up in
// TODO.md per plan-phase6-auction.md §Scope Cuts.
package persist

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/util"
)

// LoadNoAuction reads `path` line-by-line, collecting non-zero integers
// until a `0` sentinel (or EOF) is seen. Returns an empty slice for
// missing files — matching C's permissive behavior on a fresh MUD.
//
// Lines that fail integer parse are logged via util.Bug and skipped;
// the loader continues in the same tolerant spirit as other
// persist/* readers.
func LoadNoAuction(path string) ([]int, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No noauction list shipped — empty blacklist. Per plan
			// audit 2026-04-18 this is the stock tree state.
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	return parseNoAuction(f, path), nil
}

func parseNoAuction(r io.Reader, path string) []int {
	var list []int
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		n, err := strconv.Atoi(line)
		if err != nil {
			util.Bug("LoadNoAuction: %s: malformed vnum %q", path, line)
			continue
		}
		// C emits `0\n` as the terminator; a `0` vnum would otherwise
		// be a sentinel. Stop reading at the first `0` encountered.
		if n == 0 {
			break
		}
		list = append(list, n)
	}
	return list
}
