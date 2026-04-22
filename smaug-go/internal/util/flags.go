package util

import "strings"

// RisflagNames maps bit position to the RIS_* flag name used by
// `stset immune/resist/susceptible`. Derived from C's ris_flags table
// in src/stances.c (variant) and types/constants.go:196-217 which
// defines 22 RIS_* bits (0..21) terminating in "magic"+"paralysis".
//
// Used by FlagString for rendering bitsets to name lists and by
// GetRisflag for parsing stset input. Order MUST match the bit
// layout — index i corresponds to `1 << i`.
var RisflagNames = []string{
	// bits 0-7
	"fire", "cold", "electricity", "energy",
	"blunt", "pierce", "slash", "acid",
	// bits 8-13
	"poison", "drain", "sleep", "charm",
	"hold", "nonmagic",
	// bits 14-19 (plus1..plus6)
	"plus1", "plus2", "plus3", "plus4", "plus5", "plus6",
	// bits 20-21 (audit-corrected 2026-04-18 — constants.go:196-217
	// defines 22 RIS_* bits; terminal names are "magic" + "paralysis",
	// not "plus7" as an earlier draft had it)
	"magic", "paralysis",
}

// ActflagNames maps bit position to the ACT_* flag name for NPC act
// flags. Ported verbatim from C build.c:227-251 `act_flags[]`. The
// `#ifdef`-gated C entries (BANK_INSTALLED banker, ENABLE_UNDERTAKER
// undertaker, ENABLE_ARENA challenged/challenger, ENABLE_QUEST
// questmaster, OVERLANDCODE onmap) are included unconditionally here
// because the Go port enables those subsystems — index positions must
// match the ACT_* constant assignments in types/constants.go:447-495.
//
// Used by the medit G9 bitmask editor (plan-phase6-olc-medit.md §G9)
// and by future flag-name lookups on mobs. Order MUST match
// the bit layout in types/constants.go — index i corresponds to
// `1 << i` in an int field or bit i in a BitVector.
var ActflagNames = []string{
	// 0-7
	"npc", "sentinel", "scavenger", "nolocate", "r2", "aggressive",
	"stayarea", "wimpy",
	// 8-15
	"pet", "train", "practice", "immortal", "deadly", "polyself",
	"meta_aggr", "guardian",
	// 16-23
	"running", "nowander", "mountable", "mounted", "scholar",
	"secretive", "hardhat", "mobinvis",
	// 24-31
	"noassist", "autonomous", "pacifist", "noattack", "annoying",
	"statshield", "prototype", "nosummon",
	// 32-39
	"nosteal", "r2", "infested", "free2", "blocking", "is_clone",
	"is_dreamform", "is_spiritform",
	// 40-47
	"is_projection", "stopscript", "banker", "undertaker", "challenged",
	"challenger", "questmaster", "onmap",
}

// PlrflagNames maps bit position to the PLR_* flag name for connected
// players. Ported verbatim from C build.c:267-283 `plr_flags[]` with
// the `#ifdef`-gated entries (ENABLE_QUEST questmaster, OVERLANDCODE
// onmap/mapedit) included unconditionally to match the Go port's
// enabled subsystems. Order MUST match the PLR_* constant assignments.
var PlrflagNames = []string{
	// 0-7
	"npc", "boughtpet", "shovedrag", "autoexits", "autoloot", "autosac",
	"blank", "outcast",
	// 8-15
	"brief", "combine", "prompt", "telnet_ga", "holylight", "wizinvis",
	"roomvnum", "silence",
	// 16-23
	"noemote", "attacker", "notell", "log", "deny", "freeze", "thief",
	"killer",
	// 24-31
	"litterbug", "ansi", "rip", "nice", "flee", "autogold", "automap",
	"afk",
	// 32-39
	"invisprompt", "roomvis", "nofollow", "landed", "blocking",
	"is_clone", "is_dreamform", "is_spiritform",
	// 40-47
	"is_projection", "cloak", "compass", "nohomepage", "questmaster",
	"onmap", "mapedit", "r1",
}

// AffflagNames maps bit position to the AFF_* flag name used by the
// medit/oedit affect-flag bitmask editor. Ported verbatim from C
// build.c:216-225 `a_flags[]` (47 entries — terminal r8/r9/r10 reserved
// slots restored 2026-04-22 per Wave 3 follow-up LOW #5). Order MUST
// match AFF_* constants in types/enums.go:257-302.
var AffflagNames = []string{
	// 0-7
	"blind", "invisible", "detect_evil", "detect_invis", "detect_magic",
	"detect_hidden", "hold", "sanctuary",
	// 8-15
	"faerie_fire", "infrared", "curse", "_flaming", "poison", "protect",
	"_paralysis", "sneak",
	// 16-23
	"hide", "sleep", "charm", "flying", "pass_door", "floating",
	"truesight", "detect_traps",
	// 24-31
	"scrying", "fireshield", "shockshield", "r1", "iceshield", "possess",
	"berserk", "aqua_breath",
	// 32-39
	"recurringspell", "contagious", "acidmist", "venomshield", "grapple",
	"r1", "r2", "r3",
	// 40-46 (terminal reserved slots match C build.c:223-224)
	"r4", "r5", "r6", "r7", "r8", "r9", "r10",
}

// PcflagNames maps bit position to the PCFLAG_* flag name used by the
// medit PC-data flag bitmask editor. Ported verbatim from C
// build.c:253-265 `pc_flags[]` (43 entries — terminal r1..r10 reserved
// slots restored 2026-04-22 per Wave 3 follow-up LOW #5). Order MUST
// match PCFLAG_* constants.
var PcflagNames = []string{
	// 0-7
	"r1", "deadly", "unauthed", "norecall", "nointro", "gag", "retired",
	"guest",
	// 8-15
	"nosummon", "pager", "notitled", "groupwho", "diagnose", "highgag",
	"", "nstart",
	// 16-23
	"dnd", "idle", "nobio", "nodesc", "beckon", "noexp", "nobeckon",
	"hints",
	// 24-31 (includes ENABLE_BUILDWALK=buildwalk + r20..r25 reserved)
	"nohttp", "freekill", "buildwalk", "r20", "r21", "r22", "r23", "r24",
	// 32
	"r25",
	// 33-42 — terminal reserved slots match C build.c:264
	"r1", "r2", "r3", "r4", "r5", "r6", "r7", "r8", "r9", "r10",
}

// PartflagNames maps bit position to the PART_* body-part flag name
// used by the medit parts bitmask editor. Ported verbatim from C
// build.c:325-332 `part_flags[]`. Order MUST match PART_* constants.
var PartflagNames = []string{
	// 0-7
	"head", "arms", "legs", "heart", "brains", "guts", "hands", "feet",
	// 8-15
	"fingers", "ear", "eye", "long_tongue", "eyestalks", "tentacles",
	"fins", "wings",
	// 16-23
	"tail", "scales", "claws", "fangs", "horns", "tusks", "tailattack",
	"sharpscales",
	// 24-31
	"beak", "haunches", "hooves", "paws", "forelegs", "feathers",
	"shell", "r2",
}

// GetRisflag returns the bit index (0..21) for a RIS_* flag name, or -1
// if the name is unrecognized. Case-insensitive. Mirrors C's
// `get_risflag` in src/act_wiz.c.
func GetRisflag(name string) int {
	for i, n := range RisflagNames {
		if strings.EqualFold(n, name) {
			return i
		}
	}
	return -1
}

// GetActflag returns the bit index for an ACT_* flag name, or -1 if
// unrecognized. Case-insensitive. Empty / reserved slots ("r2", etc.)
// are matchable; callers that should reject them can post-check.
// Mirrors C's `get_actflag` in src/build.c.
func GetActflag(name string) int {
	for i, n := range ActflagNames {
		if n != "" && strings.EqualFold(n, name) {
			return i
		}
	}
	return -1
}

// GetPlrflag returns the bit index for a PLR_* flag name, or -1 if
// unrecognized. Case-insensitive.
func GetPlrflag(name string) int {
	for i, n := range PlrflagNames {
		if n != "" && strings.EqualFold(n, name) {
			return i
		}
	}
	return -1
}

// GetAffflag returns the bit index for an AFF_* flag name, or -1 if
// unrecognized. Case-insensitive.
func GetAffflag(name string) int {
	for i, n := range AffflagNames {
		if n != "" && strings.EqualFold(n, name) {
			return i
		}
	}
	return -1
}

// GetPcflag returns the bit index for a PCFLAG_* flag name, or -1 if
// unrecognized. Case-insensitive.
func GetPcflag(name string) int {
	for i, n := range PcflagNames {
		if n != "" && strings.EqualFold(n, name) {
			return i
		}
	}
	return -1
}

// GetPartflag returns the bit index for a PART_* flag name, or -1 if
// unrecognized. Case-insensitive.
func GetPartflag(name string) int {
	for i, n := range PartflagNames {
		if n != "" && strings.EqualFold(n, name) {
			return i
		}
	}
	return -1
}

// FlagString renders a bitset to a space-separated list of flag names.
// Returns "none" when no bit is set. Mirrors C's `flag_string` in db.c.
func FlagString(flags int, names []string) string {
	var parts []string
	for i, name := range names {
		if flags&(1<<i) != 0 {
			parts = append(parts, name)
		}
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, " ")
}
