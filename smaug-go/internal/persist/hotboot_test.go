//go:build !windows

package persist

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// --- fixture helpers ---------------------------------------------------

// newHotbootWorld returns a minimal world with two rooms and one mob
// prototype wired in. Additional rooms/mobs can be added by the caller.
func newHotbootWorld(t *testing.T) *world.World {
	t.Helper()
	w := world.New(t.TempDir())
	// Two stock rooms
	w.Rooms[3001] = &types.RoomIndexData{Vnum: 3001, Name: "Temple"}
	w.Rooms[3002] = &types.RoomIndexData{Vnum: 3002, Name: "Square"}
	// Limbo fallback
	w.Rooms[types.ROOM_VNUM_LIMBO] = &types.RoomIndexData{Vnum: types.ROOM_VNUM_LIMBO, Name: "Limbo"}
	// Mob prototype vnum 100
	w.MobIndex[100] = &types.MobIndexData{
		Vnum:        100,
		PlayerName:  "guard",
		ShortDescr:  "a guard",
		LongDescr:   "A guard stands here.",
		Description: "Beefy.",
	}
	// Object prototype vnum 200
	w.ObjIndex[200] = &types.ObjIndexData{
		Vnum:       200,
		Name:       "sword",
		ShortDescr: "a sword",
		Weight:     5,
		ItemType:   types.ITEM_WEAPON,
	}
	return w
}

// makeMob builds a live NPC tied to the vnum=100 prototype, level+gold+room populated.
func makeMob(w *world.World, roomVnum, level, gold int) *types.CharData {
	idx := w.MobIndex[100]
	mob := &types.CharData{
		Name:        idx.PlayerName,
		ShortDescr:  idx.ShortDescr,
		LongDescr:   idx.LongDescr,
		Description: idx.Description,
		Level:       level,
		Gold:        gold,
		Hit:         50,
		MaxHit:      50,
		Mana:        100,
		MaxMana:     100,
		Move:        150,
		MaxMove:     150,
		Position:    types.POS_STANDING,
		IndexData:   idx,
	}
	mob.Act.Set(types.ACT_IS_NPC)
	if room, ok := w.Rooms[roomVnum]; ok {
		mob.InRoom = room
		room.People = append(room.People, mob)
	}
	w.AddChar(mob)
	return mob
}

// makeObj builds a live object tied to the vnum=200 prototype.
func makeObj(w *world.World) *types.ObjData {
	idx := w.ObjIndex[200]
	obj := &types.ObjData{
		IndexData:  idx,
		Name:       idx.Name,
		ShortDescr: idx.ShortDescr,
		Weight:     idx.Weight,
		ItemType:   idx.ItemType,
		WearLoc:    types.WEAR_NONE,
		Count:      1,
	}
	w.AddObj(obj)
	return obj
}

// --- SaveMob / LoadMob round-trip -------------------------------------

func TestSaveMob_RoundTripSimple(t *testing.T) {
	w := newHotbootWorld(t)
	mob := makeMob(w, 3001, 10, 500)

	var buf bytes.Buffer
	if err := SaveMob(&buf, mob); err != nil {
		t.Fatalf("SaveMob: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("SaveMob: produced no output")
	}

	// Load into a fresh world with the same prototype.
	w2 := newHotbootWorld(t)
	sc := NewScanner(&buf, "buf")
	// SaveMob emits "#MOBILE\n" as the opener; consume it.
	word := sc.ReadWord()
	if word != "#MOBILE" {
		t.Fatalf("expected #MOBILE, got %q", word)
	}
	loaded := LoadMob(w2, sc)
	if loaded == nil {
		t.Fatal("LoadMob returned nil")
	}
	if loaded.Level != 10 {
		t.Errorf("Level = %d, want 10", loaded.Level)
	}
	if loaded.Gold != 500 {
		t.Errorf("Gold = %d, want 500", loaded.Gold)
	}
	if loaded.InRoom == nil || loaded.InRoom.Vnum != 3001 {
		t.Errorf("InRoom = %v, want vnum 3001", loaded.InRoom)
	}
	if loaded.IndexData == nil || loaded.IndexData.Vnum != 100 {
		t.Errorf("IndexData vnum = %v, want 100", loaded.IndexData)
	}
}

func TestSaveMob_WithInventory(t *testing.T) {
	w := newHotbootWorld(t)
	mob := makeMob(w, 3001, 5, 0)
	o1 := makeObj(w)
	o2 := makeObj(w)
	mob.Carrying = append(mob.Carrying, o1, o2)
	o1.CarriedBy = mob
	o2.CarriedBy = mob

	var buf bytes.Buffer
	if err := SaveMob(&buf, mob); err != nil {
		t.Fatalf("SaveMob: %v", err)
	}
	out := buf.String()
	// Two #OBJECT blocks inside.
	if cnt := strings.Count(out, "#OBJECT"); cnt != 2 {
		t.Fatalf("#OBJECT count = %d, want 2; out=\n%s", cnt, out)
	}

	// Round-trip
	w2 := newHotbootWorld(t)
	sc := NewScanner(strings.NewReader(out), "buf")
	_ = sc.ReadWord() // consume #MOBILE
	loaded := LoadMob(w2, sc)
	if loaded == nil {
		t.Fatal("LoadMob returned nil")
	}
	if len(loaded.Carrying) != 2 {
		t.Fatalf("Carrying count = %d, want 2", len(loaded.Carrying))
	}
	for _, c := range loaded.Carrying {
		if c.IndexData == nil || c.IndexData.Vnum != 200 {
			t.Errorf("carried obj vnum = %v, want 200", c.IndexData)
		}
	}
}

func TestSaveMob_WithAffects(t *testing.T) {
	w := newHotbootWorld(t)
	mob := makeMob(w, 3001, 5, 0)
	mob.Affects = append(mob.Affects, &types.AffectData{
		Type: 1, Duration: 20, Modifier: 3, Location: 4,
	})
	mob.Affects = append(mob.Affects, &types.AffectData{
		Type: 2, Duration: 30, Modifier: -2, Location: 5,
	})

	var buf bytes.Buffer
	if err := SaveMob(&buf, mob); err != nil {
		t.Fatalf("SaveMob: %v", err)
	}
	// Round-trip
	w2 := newHotbootWorld(t)
	sc := NewScanner(&buf, "buf")
	_ = sc.ReadWord()
	loaded := LoadMob(w2, sc)
	if loaded == nil {
		t.Fatal("nil")
	}
	if len(loaded.Affects) != 2 {
		t.Fatalf("affect count = %d, want 2", len(loaded.Affects))
	}
	if loaded.Affects[0].Duration != 20 || loaded.Affects[0].Modifier != 3 {
		t.Errorf("affect 0 = %+v", loaded.Affects[0])
	}
	if loaded.Affects[1].Duration != 30 || loaded.Affects[1].Modifier != -2 {
		t.Errorf("affect 1 = %+v", loaded.Affects[1])
	}
}

func TestSaveMob_SkipsPC(t *testing.T) {
	// PC has no IndexData and no ACT_IS_NPC. C uses !IS_NPC check.
	ch := &types.CharData{Name: "Alice", Level: 10, PCData: &types.PCData{}}
	var buf bytes.Buffer
	if err := SaveMob(&buf, ch); err != nil {
		t.Fatalf("SaveMob PC: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("PC produced output: %q", buf.String())
	}
}

func TestSaveMob_SkipsPrototype(t *testing.T) {
	w := newHotbootWorld(t)
	mob := makeMob(w, 3001, 5, 0)
	mob.Act.Set(types.ACT_PROTOTYPE)
	var buf bytes.Buffer
	if err := SaveMob(&buf, mob); err != nil {
		t.Fatalf("SaveMob: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("ACT_PROTOTYPE produced output: %q", buf.String())
	}
}

func TestSaveMob_SkipsPet(t *testing.T) {
	w := newHotbootWorld(t)
	mob := makeMob(w, 3001, 5, 0)
	mob.Act.Set(types.ACT_PET)
	var buf bytes.Buffer
	if err := SaveMob(&buf, mob); err != nil {
		t.Fatalf("SaveMob: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("ACT_PET produced output: %q", buf.String())
	}
}

// --- SaveWorld / LoadWorld -----------------------------------------------

// TestSaveWorld_FileMode pins 0600 on mobfile.dat and per-room *.objdat.
// Post-adversary fix (2026-04-19): these files carry NPC state + room
// contents and should not be world-readable. Prior `os.Create` landed
// them at 0644 on typical deployments (umask 0022).
func TestSaveWorld_FileMode(t *testing.T) {
	w := newHotbootWorld(t)
	m1 := makeMob(w, 3001, 10, 100)
	o1 := makeObj(w)
	m1.Carrying = append(m1.Carrying, o1)
	o1.CarriedBy = m1
	// Put an object on the room floor so an objdat file is emitted.
	rm := w.GetRoom(3001)
	if rm != nil {
		floorObj := makeObj(w)
		rm.Contents = append(rm.Contents, floorObj)
	}

	dir := t.TempDir()
	if err := SaveWorld(w, dir); err != nil {
		t.Fatalf("SaveWorld: %v", err)
	}
	hotbootDir := filepath.Join(dir, "hotboot")

	mobInfo, err := os.Stat(filepath.Join(hotbootDir, "mobfile.dat"))
	if err != nil {
		t.Fatalf("stat mobfile.dat: %v", err)
	}
	if mode := mobInfo.Mode().Perm(); mode != 0o600 {
		t.Errorf("mobfile.dat mode = %#o, want 0o600", mode)
	}

	objdat := filepath.Join(hotbootDir, "3001.objdat")
	objInfo, err := os.Stat(objdat)
	if err != nil {
		t.Fatalf("stat 3001.objdat: %v", err)
	}
	if mode := objInfo.Mode().Perm(); mode != 0o600 {
		t.Errorf("3001.objdat mode = %#o, want 0o600", mode)
	}
}

func TestSaveWorld_RoundTripTwoRoomsTwoMobs(t *testing.T) {
	w := newHotbootWorld(t)
	m1 := makeMob(w, 3001, 10, 100)
	m2 := makeMob(w, 3002, 15, 200)
	o1 := makeObj(w)
	o2 := makeObj(w)
	m1.Carrying = append(m1.Carrying, o1)
	o1.CarriedBy = m1
	m2.Carrying = append(m2.Carrying, o2)
	o2.CarriedBy = m2

	dir := t.TempDir()
	if err := SaveWorld(w, dir); err != nil {
		t.Fatalf("SaveWorld: %v", err)
	}

	// Mobfile must exist.
	mobfile := filepath.Join(dir, "hotboot", "mobfile.dat")
	if _, err := os.Stat(mobfile); err != nil {
		t.Fatalf("mobfile missing: %v", err)
	}

	// Fresh world with same prototypes.
	w2 := newHotbootWorld(t)
	if err := LoadWorld(w2, dir); err != nil {
		t.Fatalf("LoadWorld: %v", err)
	}
	if len(w2.Characters) != 2 {
		t.Fatalf("Characters count = %d, want 2", len(w2.Characters))
	}
	// Find mobs by room.
	byRoom := map[int]*types.CharData{}
	for _, c := range w2.Characters {
		if c.InRoom != nil {
			byRoom[c.InRoom.Vnum] = c
		}
	}
	if byRoom[3001] == nil || byRoom[3001].Level != 10 || byRoom[3001].Gold != 100 {
		t.Errorf("mob in 3001 = %+v", byRoom[3001])
	}
	if byRoom[3002] == nil || byRoom[3002].Level != 15 || byRoom[3002].Gold != 200 {
		t.Errorf("mob in 3002 = %+v", byRoom[3002])
	}
	// Both should have inventory.
	if len(byRoom[3001].Carrying) != 1 {
		t.Errorf("3001 mob Carrying = %d, want 1", len(byRoom[3001].Carrying))
	}
	if len(byRoom[3002].Carrying) != 1 {
		t.Errorf("3002 mob Carrying = %d, want 1", len(byRoom[3002].Carrying))
	}
}

func TestSaveWorld_SkipsClanStore(t *testing.T) {
	w := newHotbootWorld(t)
	// Flag room 3001 as clan-store
	w.Rooms[3001].RoomFlags.Set(types.ROOM_CLANSTOREROOM)
	// Drop a floor object in 3001.
	obj := makeObj(w)
	w.Rooms[3001].Contents = append(w.Rooms[3001].Contents, obj)
	obj.InRoom = w.Rooms[3001]

	// Also put a mob in 3001 — mobfile should still include it (C skip is
	// per-room for the object file, not per-mob).
	_ = makeMob(w, 3002, 5, 0)

	dir := t.TempDir()
	if err := SaveWorld(w, dir); err != nil {
		t.Fatalf("SaveWorld: %v", err)
	}
	// No objdat for 3001 (was clan-store).
	clanObj := filepath.Join(dir, "hotboot", "3001.objdat")
	if _, err := os.Stat(clanObj); !os.IsNotExist(err) {
		t.Errorf("clan-store room objdat should not exist, stat err = %v", err)
	}
}

func TestLoadWorld_MissingPrototype_Logs(t *testing.T) {
	// Manually craft a mobfile referencing an unknown vnum.
	dir := t.TempDir()
	hotbootDir := filepath.Join(dir, "hotboot")
	if err := os.MkdirAll(hotbootDir, 0o755); err != nil {
		t.Fatal(err)
	}
	raw := "#MOBILE\nVnum 99999\nLevel 5\nRoom 3001\nHpManaMove 10 10 10 10 10 10\nPosition 8\nEndMobile\n\n#END\n"
	if err := os.WriteFile(filepath.Join(hotbootDir, "mobfile.dat"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	w := newHotbootWorld(t)
	if err := LoadWorld(w, dir); err != nil {
		t.Fatalf("LoadWorld: %v", err)
	}
	if len(w.Characters) != 0 {
		t.Errorf("Characters count = %d, want 0 (bad vnum should be skipped)", len(w.Characters))
	}
}

func TestLoadWorld_MissingRoom_FallsBackToLimbo(t *testing.T) {
	dir := t.TempDir()
	hotbootDir := filepath.Join(dir, "hotboot")
	if err := os.MkdirAll(hotbootDir, 0o755); err != nil {
		t.Fatal(err)
	}
	raw := "#MOBILE\nVnum 100\nLevel 5\nRoom 99999\nHpManaMove 10 10 10 10 10 10\nPosition 8\nEndMobile\n\n#END\n"
	if err := os.WriteFile(filepath.Join(hotbootDir, "mobfile.dat"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	w := newHotbootWorld(t)
	if err := LoadWorld(w, dir); err != nil {
		t.Fatalf("LoadWorld: %v", err)
	}
	if len(w.Characters) != 1 {
		t.Fatalf("Characters count = %d, want 1", len(w.Characters))
	}
	got := w.Characters[0]
	if got.InRoom == nil || got.InRoom.Vnum != types.ROOM_VNUM_LIMBO {
		t.Errorf("InRoom = %v, want LIMBO", got.InRoom)
	}
}

func TestLoadWorld_UnlinksMobFileAfterLoad(t *testing.T) {
	w := newHotbootWorld(t)
	makeMob(w, 3001, 5, 0)
	dir := t.TempDir()
	if err := SaveWorld(w, dir); err != nil {
		t.Fatal(err)
	}
	mobfile := filepath.Join(dir, "hotboot", "mobfile.dat")
	if _, err := os.Stat(mobfile); err != nil {
		t.Fatalf("pre-load: mobfile missing: %v", err)
	}
	w2 := newHotbootWorld(t)
	if err := LoadWorld(w2, dir); err != nil {
		t.Fatalf("LoadWorld: %v", err)
	}
	if _, err := os.Stat(mobfile); !os.IsNotExist(err) {
		t.Errorf("mobfile still exists after load (stat err=%v)", err)
	}
}

func TestLoadWorld_UnlinksObjFilesAfterLoad(t *testing.T) {
	w := newHotbootWorld(t)
	obj := makeObj(w)
	w.Rooms[3001].Contents = append(w.Rooms[3001].Contents, obj)
	obj.InRoom = w.Rooms[3001]
	dir := t.TempDir()
	if err := SaveWorld(w, dir); err != nil {
		t.Fatal(err)
	}
	objdat := filepath.Join(dir, "hotboot", "3001.objdat")
	if _, err := os.Stat(objdat); err != nil {
		t.Fatalf("pre-load: objdat missing: %v", err)
	}
	w2 := newHotbootWorld(t)
	if err := LoadWorld(w2, dir); err != nil {
		t.Fatalf("LoadWorld: %v", err)
	}
	if _, err := os.Stat(objdat); !os.IsNotExist(err) {
		t.Errorf("objdat still exists after load (stat err=%v)", err)
	}
	// Verify object is in the room.
	if len(w2.Rooms[3001].Contents) != 1 {
		t.Errorf("room 3001 contents = %d, want 1", len(w2.Rooms[3001].Contents))
	}
}

func TestSaveWorld_EmptyRoomsProduceNoObjFile(t *testing.T) {
	w := newHotbootWorld(t)
	// No mobs, no floor objects in any room.
	dir := t.TempDir()
	if err := SaveWorld(w, dir); err != nil {
		t.Fatalf("SaveWorld: %v", err)
	}
	hotbootDir := filepath.Join(dir, "hotboot")
	entries, err := os.ReadDir(hotbootDir)
	if err != nil {
		// Directory may not exist at all — acceptable.
		return
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".objdat") {
			t.Errorf("unexpected objdat: %s", e.Name())
		}
	}
}
