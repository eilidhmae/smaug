package types

// ObjIndexData is the prototype/template for an object.
// Maps to C struct obj_index_data (mud.h:3101).
type ObjIndexData struct {
	ExtraDescr []*ExtraDescrData
	Affects    []*AffectData
	MudProgs   []*MProgData
	ProgTypes  BitVector

	Name        string
	ShortDescr  string
	Description string
	ActionDesc  string

	Vnum       int
	Level      int
	ItemType   int
	ExtraFlags BitVector
	MagicFlags int
	WearFlags  int
	Count      int
	Weight     int
	GoldCost   int
	SilverCost int
	CopperCost int
	Value      [6]int
	Serial     int
	Layers     int
	Rent       int
}

// ObjData is one object instance in the game world.
// Maps to C struct obj_data (mud.h:3143).
type ObjData struct {
	// Contents (if this is a container)
	Contents []*ObjData

	// Parent container
	InObj *ObjData

	// Carried by this character (nil if on ground or in container)
	CarriedBy *CharData

	ExtraDescr []*ExtraDescrData
	Affects    []*AffectData

	// Template
	IndexData *ObjIndexData

	// Room (if on the ground)
	InRoom *RoomIndexData

	// Identity
	Name        string
	ShortDescr  string
	ExtraDescr2 string // Marriage extra descr
	Description string
	ActionDesc  string
	Owner       string

	// Properties
	ItemType    int
	MpScriptPos int
	ExtraFlags  BitVector
	MagicFlags  int
	WearFlags   int

	// MUD prog
	MpAct    []*MProgActList
	MpActNum int

	WearLoc    int
	Weight     int
	GoldCost   int
	SilverCost int
	CopperCost int
	Level      int
	Timer      int
	Value      [6]int
	Count      int
	Serial     int

	// Hotboot
	RoomVnum int

	// Overland
	X   int
	Y   int
	Map int
}

// ExtraDescrData for additional examine targets.
// Maps to C struct extra_descr_data (mud.h:3085).
type ExtraDescrData struct {
	Keyword     string
	Description string
}
