package persist

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func TestLoadAreaFile_Objects(t *testing.T) {
	w := world.New("")
	path := filepath.Join("testdata", "test_objects.are")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("testdata not found: %s", path)
	}

	err := loadAreaFile(w, path)
	if err != nil {
		t.Fatalf("loadAreaFile failed: %v", err)
	}

	// Verify area was loaded
	if len(w.Areas) != 1 {
		t.Fatalf("expected 1 area, got %d", len(w.Areas))
	}
	if w.Areas[0].Name != "Test Area" {
		t.Errorf("area name = %q, want %q", w.Areas[0].Name, "Test Area")
	}

	// --- Object 3400: candlestick ---
	obj1 := w.ObjIndex[3400]
	if obj1 == nil {
		t.Fatal("object 3400 not loaded")
	}
	if obj1.Name != "candlestick" {
		t.Errorf("obj 3400 name = %q, want %q", obj1.Name, "candlestick")
	}
	if obj1.ShortDescr != "a candlestick" {
		t.Errorf("obj 3400 short = %q, want %q", obj1.ShortDescr, "a candlestick")
	}
	if obj1.ItemType != types.ITEM_LIGHT {
		t.Errorf("obj 3400 item_type = %d, want %d (ITEM_LIGHT)", obj1.ItemType, types.ITEM_LIGHT)
	}
	if obj1.WearFlags != 16385 {
		t.Errorf("obj 3400 wear_flags = %d, want 16385", obj1.WearFlags)
	}
	if obj1.Weight != 5 {
		t.Errorf("obj 3400 weight = %d, want 5", obj1.Weight)
	}
	if obj1.GoldCost != 150 {
		t.Errorf("obj 3400 gold_cost = %d, want 150", obj1.GoldCost)
	}
	// Extra descriptions
	if len(obj1.ExtraDescr) != 1 {
		t.Errorf("obj 3400 extra_descr count = %d, want 1", len(obj1.ExtraDescr))
	} else if obj1.ExtraDescr[0].Keyword != "candlestick" {
		t.Errorf("obj 3400 extra keyword = %q, want %q", obj1.ExtraDescr[0].Keyword, "candlestick")
	}

	// --- Object 3402: tickler ---
	obj2 := w.ObjIndex[3402]
	if obj2 == nil {
		t.Fatal("object 3402 not loaded")
	}
	if obj2.Name != "tickler" {
		t.Errorf("obj 3402 name = %q, want %q", obj2.Name, "tickler")
	}
	if obj2.ItemType != types.ITEM_WEAPON {
		t.Errorf("obj 3402 item_type = %d, want %d (ITEM_WEAPON)", obj2.ItemType, types.ITEM_WEAPON)
	}
	if obj2.ActionDesc != "tickle" {
		t.Errorf("obj 3402 action_desc = %q, want %q", obj2.ActionDesc, "tickle")
	}
	if obj2.Weight != 8 {
		t.Errorf("obj 3402 weight = %d, want 8", obj2.Weight)
	}
	// Extra descriptions
	if len(obj2.ExtraDescr) != 1 {
		t.Errorf("obj 3402 extra_descr count = %d, want 1", len(obj2.ExtraDescr))
	}
	// Affects
	if len(obj2.Affects) != 2 {
		t.Errorf("obj 3402 affects count = %d, want 2", len(obj2.Affects))
	} else {
		if obj2.Affects[0].Location != 18 || obj2.Affects[0].Modifier != 1 {
			t.Errorf("obj 3402 affect[0] = loc=%d mod=%d, want loc=18 mod=1",
				obj2.Affects[0].Location, obj2.Affects[0].Modifier)
		}
		if obj2.Affects[1].Location != 19 || obj2.Affects[1].Modifier != 2 {
			t.Errorf("obj 3402 affect[1] = loc=%d mod=%d, want loc=19 mod=2",
				obj2.Affects[1].Location, obj2.Affects[1].Modifier)
		}
	}
}

func TestLoadAreaFile_Rooms(t *testing.T) {
	w := world.New("")
	path := filepath.Join("testdata", "test_objects.are")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("testdata not found: %s", path)
	}

	err := loadAreaFile(w, path)
	if err != nil {
		t.Fatalf("loadAreaFile failed: %v", err)
	}

	// --- Room 3405 ---
	room1 := w.Rooms[3405]
	if room1 == nil {
		t.Fatal("room 3405 not loaded")
	}
	if room1.Name != "Inside the Chapel" {
		t.Errorf("room 3405 name = %q, want %q", room1.Name, "Inside the Chapel")
	}
	// In the test data, sector is 1 (SECT_CITY — "0 0 1" = unused, flags, sector)
	if room1.SectorType != types.SECT_CITY {
		t.Errorf("room 3405 sector = %d, want %d (SECT_CITY)", room1.SectorType, types.SECT_CITY)
	}
	// Exits
	if len(room1.Exits) != 1 {
		t.Fatalf("room 3405 exits count = %d, want 1", len(room1.Exits))
	}
	if room1.Exits[0].Direction != types.DIR_NORTH {
		t.Errorf("room 3405 exit dir = %d, want %d (DIR_NORTH)", room1.Exits[0].Direction, types.DIR_NORTH)
	}
	if room1.Exits[0].Vnum != 3406 {
		t.Errorf("room 3405 exit vnum = %d, want 3406", room1.Exits[0].Vnum)
	}

	// --- Room 3406 ---
	room2 := w.Rooms[3406]
	if room2 == nil {
		t.Fatal("room 3406 not loaded")
	}
	if room2.Name != "The Courtyard" {
		t.Errorf("room 3406 name = %q, want %q", room2.Name, "The Courtyard")
	}
	// In the test data, sector is 2 (SECT_FIELD — "0 0 2")
	if room2.SectorType != types.SECT_FIELD {
		t.Errorf("room 3406 sector = %d, want %d (SECT_FIELD)", room2.SectorType, types.SECT_FIELD)
	}

	// FixExits should link them
	w.FixExits()
	if room1.Exits[0].ToRoom != room2 {
		t.Error("room 3405 north exit not linked to room 3406 after FixExits")
	}
}

func TestLoadAreaFile_Mobs(t *testing.T) {
	w := world.New("")

	// Use chapel.are which has known mob data
	path := filepath.Join("testdata", "test_mobs.are")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("testdata not found: %s", path)
	}

	err := loadAreaFile(w, path)
	if err != nil {
		t.Fatalf("loadAreaFile failed: %v", err)
	}

	mob := w.MobIndex[3400]
	if mob == nil {
		t.Fatal("mob 3400 not loaded")
	}
	if mob.PlayerName != "skeleton bony" {
		t.Errorf("mob name = %q, want %q", mob.PlayerName, "skeleton bony")
	}
	if mob.Level != 1 {
		t.Errorf("mob level = %d, want 1", mob.Level)
	}
	if mob.Alignment != 0 {
		t.Errorf("mob alignment = %d, want 0", mob.Alignment)
	}
	if mob.Gold != 5 {
		t.Errorf("mob gold = %d, want 5", mob.Gold)
	}
	if mob.Exp != 100 {
		t.Errorf("mob exp = %d, want 100", mob.Exp)
	}
	if !mob.Act.IsSet(types.ACT_IS_NPC) {
		t.Error("mob ACT_IS_NPC not set")
	}
}

func TestLoadAreaFile_VMob(t *testing.T) {
	w := world.New("")
	path := filepath.Join("testdata", "test_vmob.are")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("testdata not found: %s", path)
	}

	err := loadAreaFile(w, path)
	if err != nil {
		t.Fatalf("loadAreaFile failed: %v", err)
	}

	mob := w.MobIndex[21000]
	if mob == nil {
		t.Fatal("mob 21000 not loaded")
	}
	if mob.PlayerName != "healer cleric priestess" {
		t.Errorf("mob name = %q", mob.PlayerName)
	}
	if mob.Level != 50 {
		t.Errorf("mob level = %d, want 50", mob.Level)
	}
	if mob.Gold != 0 {
		t.Errorf("mob gold = %d, want 0", mob.Gold)
	}
	if mob.Exp != 0 {
		t.Errorf("mob exp = %d, want 0", mob.Exp)
	}
	if mob.Sex != 2 {
		t.Errorf("mob sex = %d, want 2", mob.Sex)
	}
	if mob.PermStr != 18 {
		t.Errorf("mob perm_str = %d, want 18", mob.PermStr)
	}

	room := w.Rooms[21001]
	if room == nil {
		t.Fatal("room 21001 not loaded")
	}
	if room.Name != "The Temple" {
		t.Errorf("room name = %q, want %q", room.Name, "The Temple")
	}
}
