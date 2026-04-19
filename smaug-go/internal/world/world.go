// Package world holds the mutable game state, replacing C global variables.
package world

import (
	"log"

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

	// Planes — named groupings of rooms (src/planes.c:48). Written by
	// DoPset / SavePlanes; read by DoPlist / DoPstat + CheckPlanes
	// orphan-assignment on every room.Plane back-reference.
	Planes []*types.PlaneData

	// Holidays — named calendar entries (src/holidays.h:48-56).
	// Loaded from db/system/holidays.dat at boot; read by DoHolidays,
	// DoTime's holiday-today suffix, and DoSetHoliday CRUD.
	Holidays []*types.HolidayData

	// Morphs — polymorph templates loaded from db/system/morph.dat
	// (plan-phase6-polymorph.md §4.1). Read by GetMorph/GetMorphVnum
	// in handler/polymorph.go; persisted via SaveMorphs on morphset
	// save / morphcreate / morphdestroy. MorphVnumCounter tracks the
	// next available vnum handed out by SetupMorphVnum — mirrors C
	// morph_vnum at src/polymorph.c:49.
	Morphs           []*types.MorphData
	MorphVnumCounter int

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
	SysData     types.SystemData
	TimeInfo    types.TimeInfoData
	WeatherInfo types.WeatherInfoData

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

// GetMorphs returns the morph slice. Satisfies persist.worldMorphs so
// SetupMorphVnum can touch the morph table without an import cycle.
func (w *World) GetMorphs() []*types.MorphData { return w.Morphs }

// SetMorphVnumCounter updates the vnum counter after SetupMorphVnum.
func (w *World) SetMorphVnumCounter(v int) { w.MorphVnumCounter = v }

// GetMorphVnumCounter returns the current vnum counter.
func (w *World) GetMorphVnumCounter() int { return w.MorphVnumCounter }

// GetMorph returns the morph with the given case-insensitive name, or
// nil. Mirrors C get_morph at src/polymorph.c:1205-1216.
func (w *World) GetMorph(name string) *types.MorphData {
	if name == "" {
		return nil
	}
	for _, m := range w.Morphs {
		if m != nil && equalFold(m.Name, name) {
			return m
		}
	}
	return nil
}

// GetMorphVnum returns the morph with the given vnum, or nil. Mirrors
// C get_morph_vnum at src/polymorph.c:1243-1254.
func (w *World) GetMorphVnum(vnum int) *types.MorphData {
	if vnum < 1 {
		return nil
	}
	for _, m := range w.Morphs {
		if m != nil && m.Vnum == vnum {
			return m
		}
	}
	return nil
}

// equalFold is a light shim over strings.EqualFold to keep this file
// free of a strings import solely for one helper; the caller uses it
// twice.
func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 32
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 32
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// FixExits resolves exit vnum fields to actual room pointers after all
// areas have been loaded. Equivalent to fix_exits() in db.c.
func (w *World) FixExits() {
	fixed := 0
	broken := 0
	for _, room := range w.Rooms {
		for _, exit := range room.Exits {
			if exit.ToRoom == nil && exit.Vnum > 0 {
				dest := w.Rooms[exit.Vnum]
				if dest != nil {
					exit.ToRoom = dest
					fixed++
				} else {
					broken++
				}
			}
		}
	}
	if broken > 0 {
		log.Printf("FixExits: %d exits resolved, %d broken (destination room not found)", fixed, broken)
	} else {
		log.Printf("FixExits: %d exits resolved.", fixed)
	}
}
