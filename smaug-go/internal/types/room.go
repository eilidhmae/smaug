package types

// RoomIndexData represents a room in the game world.
// Maps to C struct room_index_data (mud.h:3467).
type RoomIndexData struct {
	// People currently in the room
	People []*CharData

	// Objects on the floor
	Contents []*ObjData

	// Extra descriptions
	ExtraDescr []*ExtraDescrData

	// Area this room belongs to
	Area *AreaData

	// Exits
	Exits []*ExitData

	// Room-wide effects
	Affects []*AffectData

	// Map data
	MapData *MapData
	Plane   *PlaneData

	// MUD progs
	MpAct       []*MProgActList
	MpActNum    int
	MudProgs    []*MProgData
	MpScriptPos int
	ProgTypes   BitVector

	// Identity
	Name        string
	Description string
	Vnum        int

	// Weight/capacity
	Weight    int
	MaxWeight int

	// Flags
	RoomFlags BitVector

	// Lighting
	Light int

	// Terrain
	SectorType int

	// Teleport
	TeleVnum  int
	TeleDelay int

	// Tunnel limit
	Tunnel int
}

// ExitData represents an exit from one room to another.
// Maps to C struct exit_data (mud.h:3202).
type ExitData struct {
	// Reverse exit pointer (set during fix_exits)
	ReverseExit *ExitData

	// Destination room
	ToRoom *RoomIndexData

	Keyword     string
	Description string

	// Destination vnum (before resolving to pointer)
	Vnum int
	// Reverse exit room vnum
	RVnum int

	// Door flags (EX_ISDOOR, EX_CLOSED, etc.)
	ExitInfo int
	// Key vnum needed to unlock
	Key int

	// Direction
	Direction int
	OrigDoor  int
	Distance  int
	Pull      int
	PullType  int

	// Overland coordinates
	X int
	Y int
}

// GetExit returns the exit in the given direction, or nil.
func (r *RoomIndexData) GetExit(dir int) *ExitData {
	for _, ex := range r.Exits {
		if ex.Direction == dir {
			return ex
		}
	}
	return nil
}

// MapData for room maps.
type MapData struct {
	// Room map data - to be fleshed out
	Vnum int
}

// PlaneData for planar rooms.
type PlaneData struct {
	Name string
}
