// Package world holds the mutable game state, replacing C global variables.
package world

import (
	"github.com/eilidhmae/smaug/internal/types"
)

// World holds all game state. There is exactly one instance, passed explicitly
// to all game systems. This replaces the C global linked lists and hash tables.
type World struct {
	// Index data (templates loaded from area files)
	Rooms    map[int]*types.RoomIndexData
	MobIndex map[int]*types.MobIndexData
	ObjIndex map[int]*types.ObjIndexData

	// Live entities in the game
	Characters []*types.CharData
	Objects    []*types.ObjData

	// Network connections
	Descriptors []*types.DescriptorData

	// Areas
	Areas []*types.AreaData

	// Help system
	Helps []*types.HelpData

	// Data tables
	Skills  []*types.SkillType
	Classes []*types.ClassType
	Races   []*types.RaceData

	// Clans, Councils, Deities
	Clans    []*types.ClanData
	Councils []*types.CouncilData
	Deities  []*types.DeityData

	// Shops
	Shops   []*types.ShopData
	Repairs []*types.RepairData

	// Boards
	Boards []*types.BoardData

	// Socials
	Socials []*types.SocialType

	// Bans
	Bans []*types.BanData

	// Global config
	SysData  types.SystemData
	TimeInfo types.TimeInfoData

	// Auction
	Auction *types.AuctionData

	// Extraction queues (processed at end of each pulse)
	ExtractChars []*types.CharData
	ExtractObjs  []*types.ObjData

	// Counters
	TopRoom   int
	TopMob    int
	TopObj    int
	TopAffect int
	TopArea   int
	TopHelp   int
	TopReset  int

	// Data directory path
	DataDir string
}

// New creates a new World with initialized maps.
func New(dataDir string) *World {
	return &World{
		Rooms:    make(map[int]*types.RoomIndexData),
		MobIndex: make(map[int]*types.MobIndexData),
		ObjIndex: make(map[int]*types.ObjIndexData),
		DataDir:  dataDir,
	}
}

// GetRoom looks up a room by vnum.
func (w *World) GetRoom(vnum int) *types.RoomIndexData {
	return w.Rooms[vnum]
}

// GetMobIndex looks up a mob prototype by vnum.
func (w *World) GetMobIndex(vnum int) *types.MobIndexData {
	return w.MobIndex[vnum]
}

// GetObjIndex looks up an object prototype by vnum.
func (w *World) GetObjIndex(vnum int) *types.ObjIndexData {
	return w.ObjIndex[vnum]
}

// AddChar adds a character to the global character list.
func (w *World) AddChar(ch *types.CharData) {
	w.Characters = append(w.Characters, ch)
}

// RemoveChar removes a character from the global character list.
func (w *World) RemoveChar(ch *types.CharData) {
	for i, c := range w.Characters {
		if c == ch {
			w.Characters = append(w.Characters[:i], w.Characters[i+1:]...)
			return
		}
	}
}

// AddObj adds an object to the global object list.
func (w *World) AddObj(obj *types.ObjData) {
	w.Objects = append(w.Objects, obj)
}

// RemoveObj removes an object from the global object list.
func (w *World) RemoveObj(obj *types.ObjData) {
	for i, o := range w.Objects {
		if o == obj {
			w.Objects = append(w.Objects[:i], w.Objects[i+1:]...)
			return
		}
	}
}
