package util

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
)

// MProgFlagNames mirrors the C mprog_flags[] table at src/build.c:366-374.
// Index = bit position. Used by both the Go-port keyword lookup (GetMpFlag)
// and reverse-display helpers. The table contains 52 entries (indices 0..51).
//
// Entries 32+ correspond to triggers that did not fit in the original 32-bit
// C bitvector. Some C entries (load, greetinfight, move, emote, r1..r10) are
// not yet ported to the Go MPROG_* enum at internal/types/mudprog.go — those
// remain in this table to preserve index alignment, but GetMpFlag returns
// (0, false) for them. Document deferral in TODO.md.
var MProgFlagNames = []string{
	0:  "act",
	1:  "speech",
	2:  "rand",
	3:  "fight",
	4:  "death",
	5:  "hitprcnt",
	6:  "entry",
	7:  "greet",
	8:  "allgreet",
	9:  "give",
	10: "bribe",
	11: "hour",
	12: "time",
	13: "wear",
	14: "remove",
	15: "sac",
	16: "look",
	17: "exa",
	18: "zap",
	19: "get",
	20: "drop",
	21: "damage",
	22: "repair",
	23: "randiw",
	24: "speechiw",
	25: "pull",
	26: "push",
	27: "sleep",
	28: "rest",
	29: "leave",
	30: "script",
	31: "use",
	32: "load", // unported in Go
	33: "login",
	34: "void",
	35: "tell",
	36: "imminfo",
	37: "greetinfight", // unported in Go
	38: "move",         // unported in Go
	39: "command",
	40: "sell",
	41: "emote", // unported in Go
	42: "r1",
	43: "r2",
	44: "r3",
	45: "r4",
	46: "r5",
	47: "r6",
	48: "r7",
	49: "r8",
	50: "r9",
	51: "r10",
}

// mpFlagLookup maps a normalized keyword to the Go MPROG_* bit-flag VALUE
// (matching MProgData.Type storage). C's get_mpflag returns the array index
// 0..51 — but Go callers want the bit-flag value directly (mprg.Type = bit)
// so the conversion happens here, not at every call site. Per Q4 of the plan.
//
// Aliases (allgreet/all_greet, exa/examine, command/cmd, enter/entry) are
// included so builders typing common variants get a hit.
var mpFlagLookup = map[string]int64{
	"act":       types.MPROG_ACT,
	"speech":    types.MPROG_SPEECH,
	"rand":      types.MPROG_RAND,
	"fight":     types.MPROG_FIGHT,
	"death":     types.MPROG_DEATH,
	"hitprcnt":  types.MPROG_HITPRCNT,
	"entry":     types.MPROG_ENTRY,
	"enter":     types.MPROG_ENTER, // alias = MPROG_ENTRY
	"greet":     types.MPROG_GREET,
	"allgreet":  types.MPROG_ALL_GREET,
	"all_greet": types.MPROG_ALL_GREET,
	"give":      types.MPROG_GIVE,
	"bribe":     types.MPROG_BRIBE,
	"hour":      types.MPROG_HOUR,
	"time":      types.MPROG_TIME,
	"wear":      types.MPROG_WEAR,
	"remove":    types.MPROG_REMOVE,
	"sac":       types.MPROG_SAC,
	"look":      types.MPROG_LOOK,
	"exa":       types.MPROG_EXA,
	"examine":   types.MPROG_EXA,
	"zap":       types.MPROG_ZAP,
	"get":       types.MPROG_GET,
	"drop":      types.MPROG_DROP,
	"damage":    types.MPROG_DAMAGE,
	"repair":    types.MPROG_REPAIR,
	"randiw":    types.MPROG_RANDIW,
	"speechiw":  types.MPROG_SPEECHIW,
	"pull":      types.MPROG_PULL,
	"push":      types.MPROG_PUSH,
	"sleep":     types.MPROG_SLEEP,
	"rest":      types.MPROG_REST,
	"leave":     types.MPROG_LEAVE,
	"script":    types.MPROG_SCRIPT,
	"use":       types.MPROG_USE,
	"login":     types.MPROG_LOGIN,
	"void":      types.MPROG_VOID,
	"tell":      types.MPROG_TELL,
	"sell":      types.MPROG_SELL,
	"imminfo":   types.MPROG_IMMINFO,
	"command":   types.MPROG_CMD,
	"cmd":       types.MPROG_CMD,
}

// FirstMProgFlagName returns a readable canonical name for the lowest set bit
// in mask, or "" if no known bit is set. Used by inspector / list output to
// reverse a stored MProgData.Type into a builder-friendly keyword.
//
// Iterates MProgFlagNames in index order so the result is deterministic and
// matches C mprog_flags[] ordering. Skips unported names (those not in
// mpFlagLookup) so stale bits never resolve to unported keywords.
func FirstMProgFlagName(mask int64) string {
	for _, name := range MProgFlagNames {
		bit, ok := mpFlagLookup[name]
		if !ok {
			continue
		}
		if mask&bit != 0 {
			return name
		}
	}
	return ""
}

// GetMpFlag looks up an mprog trigger keyword and returns its bit-flag VALUE
// (NOT bit-index; see Q4 in plan-phase6-olc-mpedit.md). Strips trailing
// "_prog" / "prog" suffixes and is case-insensitive. Returns (0, false) for
// unknown or unported names.
func GetMpFlag(name string) (int64, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return 0, false
	}
	name = strings.TrimSuffix(name, "_prog")
	name = strings.TrimSuffix(name, "prog")
	if name == "" {
		return 0, false
	}
	v, ok := mpFlagLookup[name]
	if !ok {
		return 0, false
	}
	return v, true
}
