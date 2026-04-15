package types

// MProgData for mob/obj/room programs.
// Maps to C struct mob_prog_data (mud.h:1324).
//
// Type is a bitmask of MPROG_* constants. It is int64 because the Go port uses
// Go bit flags (rather than the C EXT_BV two-word bitvector) and the full
// complement of SMAUG mob/obj/room-prog triggers exceeds 32 distinct flags.
type MProgData struct {
	Type       int64
	Triggered  bool
	ResetDelay int
	ArgList    string
	ComList    string
}

// MProgActList for queued program triggers.
// Maps to C struct mob_prog_act_list (mud.h:1314).
type MProgActList struct {
	Buf    string
	Ch     *CharData
	Obj    *ObjData
	Victim *CharData
	Target *ObjData
}

// MProgSleepData for sleeping mudprogs.
// Maps to C struct mpsleep_data (mud.h:1337).
type MProgSleepData struct {
	Timer int
	Type  int // MP_MOB, MP_ROOM, MP_OBJ
	Room  *RoomIndexData

	IgnoreLevel int
	IfLevel     int
	IfState     [MAX_IFS][2]bool

	ComList    string
	Mob        *CharData
	Actor      *CharData
	Obj        *ObjData
	Victim     *CharData
	Target     *ObjData
	SingleStep bool
}

// MUD prog trigger type constants. Stored as bit flags on MProgData.Type.
const (
	MPROG_ACT       int64 = 1 << 0
	MPROG_SPEECH    int64 = 1 << 1
	MPROG_RAND      int64 = 1 << 2
	MPROG_FIGHT     int64 = 1 << 3
	MPROG_DEATH     int64 = 1 << 4
	MPROG_HITPRCNT  int64 = 1 << 5
	MPROG_ENTRY     int64 = 1 << 6
	MPROG_GREET     int64 = 1 << 7
	MPROG_ALL_GREET int64 = 1 << 8
	MPROG_GIVE      int64 = 1 << 9
	MPROG_BRIBE     int64 = 1 << 10
	MPROG_HOUR      int64 = 1 << 11
	MPROG_TIME      int64 = 1 << 12
	MPROG_WEAR      int64 = 1 << 13
	MPROG_REMOVE    int64 = 1 << 14
	MPROG_SAC       int64 = 1 << 15
	MPROG_LOOK      int64 = 1 << 16
	MPROG_EXA       int64 = 1 << 17
	MPROG_ZAP       int64 = 1 << 18
	MPROG_GET       int64 = 1 << 19
	MPROG_DROP      int64 = 1 << 20
	MPROG_DAMAGE    int64 = 1 << 21
	MPROG_REPAIR    int64 = 1 << 22
	MPROG_RANDIW    int64 = 1 << 23
	MPROG_SPEECHIW  int64 = 1 << 24
	MPROG_PULL      int64 = 1 << 25
	MPROG_PUSH      int64 = 1 << 26
	MPROG_SLEEP     int64 = 1 << 27
	MPROG_REST      int64 = 1 << 28
	MPROG_LEAVE     int64 = 1 << 29
	MPROG_SCRIPT    int64 = 1 << 30
	MPROG_USE       int64 = 1 << 31

	// Triggers that did not fit in the original 32-bit layout. These match the
	// ordering of the C mud.h prog_types enum (LOGIN_PROG, VOID_PROG, TELL_PROG,
	// ..., SELL_PROG) but the bit offsets are unique to Go.
	MPROG_LOGIN   int64 = 1 << 32
	MPROG_VOID    int64 = 1 << 33
	MPROG_TELL    int64 = 1 << 34
	MPROG_SELL    int64 = 1 << 35
	MPROG_IMMINFO int64 = 1 << 36
	MPROG_CMD     int64 = 1 << 37

	// --- Room-prog aliases for clarity at fire sites. C uses compatibility
	// #defines in mud.h (RDEATH_PROG, ENTER_PROG, RFIGHT_PROG, RGREET_PROG) so
	// the same prog-type bit is reused. The Go aliases mirror that mapping so
	// callers don't have to remember which distinct bit lives under a given
	// room-prog keyword.
	MPROG_ENTER  = MPROG_ENTRY
	MPROG_RFIGHT = MPROG_FIGHT
	MPROG_RDEATH = MPROG_DEATH
	MPROG_RGREET = MPROG_GREET
)
