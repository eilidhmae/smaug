package types

// MProgData for mob/obj/room programs.
// Maps to C struct mob_prog_data (mud.h:1324).
type MProgData struct {
	Type       int
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

// MUD prog trigger type constants
const (
	MPROG_ACT       = 1 << 0
	MPROG_SPEECH    = 1 << 1
	MPROG_RAND      = 1 << 2
	MPROG_FIGHT     = 1 << 3
	MPROG_DEATH     = 1 << 4
	MPROG_HITPRCNT  = 1 << 5
	MPROG_ENTRY     = 1 << 6
	MPROG_GREET     = 1 << 7
	MPROG_ALL_GREET = 1 << 8
	MPROG_GIVE      = 1 << 9
	MPROG_BRIBE     = 1 << 10
	MPROG_HOUR      = 1 << 11
	MPROG_TIME      = 1 << 12
	MPROG_WEAR      = 1 << 13
	MPROG_REMOVE    = 1 << 14
	MPROG_SAC       = 1 << 15
	MPROG_LOOK      = 1 << 16
	MPROG_EXA       = 1 << 17
	MPROG_ZAP       = 1 << 18
	MPROG_GET       = 1 << 19
	MPROG_DROP      = 1 << 20
	MPROG_DAMAGE    = 1 << 21
	MPROG_REPAIR    = 1 << 22
	MPROG_RANDIW    = 1 << 23
	MPROG_SPEECHIW  = 1 << 24
	MPROG_PULL      = 1 << 25
	MPROG_PUSH      = 1 << 26
	MPROG_SLEEP     = 1 << 27
	MPROG_REST      = 1 << 28
	MPROG_LEAVE     = 1 << 29
	MPROG_SCRIPT    = 1 << 30
	MPROG_USE       = 1 << 31
)
