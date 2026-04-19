package types

import "testing"

// TestATColorTypes_PinEnumValues pins the at_color_types enum values to
// match the C enumeration order at src/mud.h:1107-1166. AT_COLORBASE is
// 1024; AT_PLAIN is the first entry at 1024; every subsequent entry is
// the previous value + 1.
//
// This test is a tripwire: if a future edit inserts or reorders AT_*
// constants, this test fails. The ordering MUST match C so that future
// ports referencing C line numbers (e.g., src/color.c:at_color_table
// indexing) interoperate.
func TestATColorTypes_PinEnumValues(t *testing.T) {
	tests := []struct {
		name string
		got  int
		want int
	}{
		{"AT_COLORBASE", AT_COLORBASE, 1024},
		{"AT_PLAIN", AT_PLAIN, 1024},
		{"AT_ACTION", AT_ACTION, 1025},
		{"AT_SAY", AT_SAY, 1026},
		{"AT_GOSSIP", AT_GOSSIP, 1027},
		{"AT_YELL", AT_YELL, 1028},
		{"AT_TELL", AT_TELL, 1029},
		{"AT_WHISPER", AT_WHISPER, 1030},
		{"AT_HIT", AT_HIT, 1031},
		{"AT_HITME", AT_HITME, 1032},
		{"AT_IMMORT", AT_IMMORT, 1033},
		{"AT_HURT", AT_HURT, 1034},
		{"AT_FALLING", AT_FALLING, 1035},
		{"AT_DANGER", AT_DANGER, 1036},
		{"AT_MAGIC", AT_MAGIC, 1037},
		{"AT_CONSIDER", AT_CONSIDER, 1038},
		{"AT_REPORT", AT_REPORT, 1039},
		{"AT_POISON", AT_POISON, 1040},
		{"AT_SOCIAL", AT_SOCIAL, 1041},
		{"AT_DYING", AT_DYING, 1042},
		{"AT_DEAD", AT_DEAD, 1043},
		{"AT_SKILL", AT_SKILL, 1044},
		{"AT_CARNAGE", AT_CARNAGE, 1045},
		{"AT_DAMAGE", AT_DAMAGE, 1046},
		{"AT_FLEE", AT_FLEE, 1047},
		{"AT_RMNAME", AT_RMNAME, 1048},
		{"AT_RMDESC", AT_RMDESC, 1049},
		{"AT_OBJECT", AT_OBJECT, 1050},
		{"AT_PERSON", AT_PERSON, 1051},
		{"AT_LIST", AT_LIST, 1052},
		{"AT_BYE", AT_BYE, 1053},
		{"AT_GOLD", AT_GOLD, 1054},
		{"AT_GTELL", AT_GTELL, 1055},
		{"AT_NOTE", AT_NOTE, 1056},
		{"AT_HUNGRY", AT_HUNGRY, 1057},
		{"AT_THIRSTY", AT_THIRSTY, 1058},
		{"AT_FIRE", AT_FIRE, 1059},
		{"AT_SOBER", AT_SOBER, 1060},
		{"AT_WEAROFF", AT_WEAROFF, 1061},
		{"AT_EXITS", AT_EXITS, 1062},
		{"AT_SCORE", AT_SCORE, 1063},
		{"AT_RESET", AT_RESET, 1064},
		{"AT_LOG", AT_LOG, 1065},
		{"AT_DIEMSG", AT_DIEMSG, 1066},
		{"AT_WARTALK", AT_WARTALK, 1067},
		{"AT_RACETALK", AT_RACETALK, 1068},
		{"AT_IGNORE", AT_IGNORE, 1069},
		{"AT_DIVIDER", AT_DIVIDER, 1070},
		{"AT_MORPH", AT_MORPH, 1071},
		{"AT_SHOUT", AT_SHOUT, 1072},
		{"AT_MUSE", AT_MUSE, 1073},
		{"AT_QUEST", AT_QUEST, 1074},
		{"AT_ASK", AT_ASK, 1075},
		{"AT_THINK", AT_THINK, 1076},
		{"AT_STANCE", AT_STANCE, 1077},
		{"AT_AVATAR", AT_AVATAR, 1078},
		{"AT_MUSIC", AT_MUSIC, 1079},
		{"AT_TOPCOLOR", AT_TOPCOLOR, 1080},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %d, want %d (C mud.h:1107-1166 enum order)", tt.name, tt.got, tt.want)
		}
	}
}

// TestATMaxColor pins the derived AT_MAXCOLOR = AT_TOPCOLOR - AT_COLORBASE
// constant (matches C mud.h:1167 `#define AT_MAXCOLOR`).
func TestATMaxColor(t *testing.T) {
	if AT_MAXCOLOR != AT_TOPCOLOR-AT_COLORBASE {
		t.Errorf("AT_MAXCOLOR = %d, want AT_TOPCOLOR-AT_COLORBASE = %d",
			AT_MAXCOLOR, AT_TOPCOLOR-AT_COLORBASE)
	}
}

// TestTimerChallenge_UniqueTag pins TIMER_CHALLENGE at iota slot 8,
// distinct from every other TIMER_* constant. Added by
// plan-phase6-arena.md §G1. If a future insertion reorders the iota
// block, this test catches the shift so arena pfiles / timer-test
// harnesses can be migrated deliberately.
func TestTimerChallenge_UniqueTag(t *testing.T) {
	all := map[string]int{
		"TIMER_NONE":        TIMER_NONE,
		"TIMER_RECENTFIGHT": TIMER_RECENTFIGHT,
		"TIMER_SHOVEDRAG":   TIMER_SHOVEDRAG,
		"TIMER_DO_FUN":      TIMER_DO_FUN,
		"TIMER_APPLIED":     TIMER_APPLIED,
		"TIMER_PKILLED":     TIMER_PKILLED,
		"TIMER_ASUPRESSED":  TIMER_ASUPRESSED,
		"TIMER_NUISANCE":    TIMER_NUISANCE,
		"TIMER_CHALLENGE":   TIMER_CHALLENGE,
	}
	seen := map[int]string{}
	for name, v := range all {
		if prior, clash := seen[v]; clash {
			t.Errorf("TIMER tag collision: %s and %s both == %d", prior, name, v)
		}
		seen[v] = name
	}
	if TIMER_CHALLENGE != 8 {
		t.Errorf("TIMER_CHALLENGE = %d, want 8", TIMER_CHALLENGE)
	}
}

// TestArenaVnumConstants pins ROOM_VNUM_ARENA_MIN / MAX at the C
// values from src/mud.h:2283-2284.
func TestArenaVnumConstants(t *testing.T) {
	if ROOM_VNUM_ARENA_MIN != 10366 {
		t.Errorf("ROOM_VNUM_ARENA_MIN = %d, want 10366", ROOM_VNUM_ARENA_MIN)
	}
	if ROOM_VNUM_ARENA_MAX != 10382 {
		t.Errorf("ROOM_VNUM_ARENA_MAX = %d, want 10382", ROOM_VNUM_ARENA_MAX)
	}
	if ROOM_VNUM_ARENA_MAX <= ROOM_VNUM_ARENA_MIN {
		t.Error("ROOM_VNUM_ARENA_MAX must exceed ROOM_VNUM_ARENA_MIN")
	}
}
