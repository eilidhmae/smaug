package persist

import (
	"os"
	"path/filepath"
	"strings"
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

func TestLoadAreaFile_SpellObjects(t *testing.T) {
	w := world.New("")
	path := filepath.Join("testdata", "test_spellobj.are")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("testdata not found: %s", path)
	}

	err := loadAreaFile(w, path)
	if err != nil {
		t.Fatalf("loadAreaFile failed: %v", err)
	}

	// --- Object 21000: potion with spell names ---
	potion := w.ObjIndex[21000]
	if potion == nil {
		t.Fatal("object 21000 (potion) not loaded")
	}
	if potion.ItemType != types.ITEM_POTION {
		t.Errorf("potion item_type = %d, want %d (ITEM_POTION)", potion.ItemType, types.ITEM_POTION)
	}
	// Spell names: 'heal' 'NONE' 'NONE' → stored in SpellNames[1], [2], [3]
	if potion.SpellNames[1] != "heal" {
		t.Errorf("potion SpellNames[1] = %q, want %q", potion.SpellNames[1], "heal")
	}
	if potion.SpellNames[2] != "NONE" {
		t.Errorf("potion SpellNames[2] = %q, want %q", potion.SpellNames[2], "NONE")
	}
	if potion.SpellNames[3] != "NONE" {
		t.Errorf("potion SpellNames[3] = %q, want %q", potion.SpellNames[3], "NONE")
	}

	// --- Object 21001: wand with spell name ---
	wand := w.ObjIndex[21001]
	if wand == nil {
		t.Fatal("object 21001 (wand) not loaded")
	}
	if wand.ItemType != types.ITEM_WAND {
		t.Errorf("wand item_type = %d, want %d (ITEM_WAND)", wand.ItemType, types.ITEM_WAND)
	}
	// Wand: spell name in value[3] → SpellNames[3]
	if wand.SpellNames[3] != "identify" {
		t.Errorf("wand SpellNames[3] = %q, want %q", wand.SpellNames[3], "identify")
	}

	// --- Object 21002: scroll with spell names ---
	scroll := w.ObjIndex[21002]
	if scroll == nil {
		t.Fatal("object 21002 (scroll) not loaded")
	}
	if scroll.ItemType != types.ITEM_SCROLL {
		t.Errorf("scroll item_type = %d, want %d (ITEM_SCROLL)", scroll.ItemType, types.ITEM_SCROLL)
	}
	if scroll.SpellNames[1] != "magic missile" {
		t.Errorf("scroll SpellNames[1] = %q, want %q", scroll.SpellNames[1], "magic missile")
	}
	if scroll.SpellNames[2] != "armor" {
		t.Errorf("scroll SpellNames[2] = %q, want %q", scroll.SpellNames[2], "armor")
	}
}

// --------------- Full area integration test ---------------

func TestLoadAreaFile_Full(t *testing.T) {
	w := world.New("")
	path := filepath.Join("testdata", "test_full.are")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("testdata not found: %s", path)
	}

	err := loadAreaFile(w, path)
	if err != nil {
		t.Fatalf("loadAreaFile failed: %v", err)
	}

	// Verify area
	if len(w.Areas) != 1 {
		t.Fatalf("expected 1 area, got %d", len(w.Areas))
	}
	area := w.Areas[0]
	if area.Name != "Full Test Area" {
		t.Errorf("area name = %q", area.Name)
	}
	if area.Author != "Tester" {
		t.Errorf("author = %q", area.Author)
	}
	if area.ResetMsg != "The area shimmers and resets." {
		t.Errorf("resetMsg = %q", area.ResetMsg)
	}
	if area.LowSoftRange != 5 || area.HiSoftRange != 15 {
		t.Errorf("soft range = %d-%d, want 5-15", area.LowSoftRange, area.HiSoftRange)
	}
	if area.LowHardRange != 1 || area.HiHardRange != 65 {
		t.Errorf("hard range = %d-%d, want 1-65", area.LowHardRange, area.HiHardRange)
	}
	if area.HighEconomy != 5000 || area.LowEconomy != 100 {
		t.Errorf("economy = %d/%d, want 5000/100", area.HighEconomy, area.LowEconomy)
	}

	// Verify mobs
	guard := w.MobIndex[5000]
	if guard == nil {
		t.Fatal("mob 5000 not loaded")
	}
	if guard.PlayerName != "guard" {
		t.Errorf("guard name = %q", guard.PlayerName)
	}
	if guard.Level != 10 {
		t.Errorf("guard level = %d", guard.Level)
	}
	if guard.Gold != 100 {
		t.Errorf("guard gold = %d", guard.Gold)
	}

	shopkeep := w.MobIndex[5001]
	if shopkeep == nil {
		t.Fatal("mob 5001 not loaded")
	}

	// Verify objects
	sword := w.ObjIndex[5000]
	if sword == nil {
		t.Fatal("obj 5000 not loaded")
	}
	if sword.ItemType != types.ITEM_WEAPON {
		t.Errorf("sword type = %d, want %d", sword.ItemType, types.ITEM_WEAPON)
	}

	// Verify rooms
	temple := w.Rooms[5000]
	if temple == nil {
		t.Fatal("room 5000 not loaded")
	}
	if temple.Name != "The Temple" {
		t.Errorf("temple name = %q", temple.Name)
	}
	if len(temple.Exits) != 1 {
		t.Errorf("temple exits = %d, want 1", len(temple.Exits))
	}

	// Verify resets
	if len(area.Resets) < 4 {
		t.Errorf("resets = %d, want >= 4", len(area.Resets))
	}

	// Verify shops
	if len(w.Shops) != 1 {
		t.Errorf("shops = %d, want 1", len(w.Shops))
	}
	if shopkeep.Shop == nil {
		t.Error("shopkeeper should have shop attached")
	}

	// Fix exits and verify linking
	w.FixExits()
	if temple.Exits[0].ToRoom == nil {
		t.Error("temple north exit not linked after FixExits")
	}
}

// --------------- loadResets ---------------

func TestLoadResets(t *testing.T) {
	// Note: G and R resets skip arg3 (only read extra, arg1, arg2 + ReadToEOL)
	// Others read extra, arg1, arg2, arg3 + ReadToEOL
	// ReadNumber consumes up to and including the terminating whitespace/newline,
	// so for G/R the last ReadNumber eats the newline and ReadToEOL eats the NEXT line.
	// In real files, each line typically has trailing content after the last number.
	// We put all args on lines with trailing text to avoid eating the next line.
	input := "M 0 100 1 200 load guard to temple\nO 0 300 1 200 load sword to room\nG 0 400 1 give to mob\nE 0 500 1 16 equip on mob\nD 0 200 0 1 set door\n* this is a comment\nS\n"
	sc := NewScanner(strings.NewReader(input), "test")
	area := &types.AreaData{}
	loadResets(sc, area)

	if len(area.Resets) != 5 {
		t.Fatalf("Resets = %d, want 5", len(area.Resets))
	}

	// M reset
	r := area.Resets[0]
	if r.Command != 'M' || r.Arg1 != 100 || r.Arg2 != 1 || r.Arg3 != 200 {
		t.Errorf("M reset = cmd=%c extra=%d arg1=%d arg2=%d arg3=%d", r.Command, r.Extra, r.Arg1, r.Arg2, r.Arg3)
	}

	// O reset
	r = area.Resets[1]
	if r.Command != 'O' || r.Arg1 != 300 || r.Arg2 != 1 || r.Arg3 != 200 {
		t.Errorf("O reset = cmd=%c extra=%d arg1=%d arg2=%d arg3=%d", r.Command, r.Extra, r.Arg1, r.Arg2, r.Arg3)
	}

	// G reset (no arg3)
	r = area.Resets[2]
	if r.Command != 'G' || r.Arg1 != 400 || r.Arg2 != 1 || r.Arg3 != 0 {
		t.Errorf("G reset = cmd=%c extra=%d arg1=%d arg2=%d arg3=%d", r.Command, r.Extra, r.Arg1, r.Arg2, r.Arg3)
	}

	// E reset
	r = area.Resets[3]
	if r.Command != 'E' || r.Arg1 != 500 || r.Arg2 != 1 || r.Arg3 != 16 {
		t.Errorf("E reset = cmd=%c extra=%d arg1=%d arg2=%d arg3=%d", r.Command, r.Extra, r.Arg1, r.Arg2, r.Arg3)
	}

	// D reset
	r = area.Resets[4]
	if r.Command != 'D' || r.Arg1 != 200 || r.Arg2 != 0 || r.Arg3 != 1 {
		t.Errorf("D reset = cmd=%c extra=%d arg1=%d arg2=%d arg3=%d", r.Command, r.Extra, r.Arg1, r.Arg2, r.Arg3)
	}
}

func TestLoadResets_Empty(t *testing.T) {
	input := "S\n"
	sc := NewScanner(strings.NewReader(input), "test")
	area := &types.AreaData{}
	loadResets(sc, area)

	if len(area.Resets) != 0 {
		t.Errorf("Resets = %d, want 0", len(area.Resets))
	}
}

// --------------- loadShops ---------------

func TestLoadShops(t *testing.T) {
	// Create a mob index first
	w := world.New("")
	mob := &types.MobIndexData{Vnum: 100, PlayerName: "shopkeeper"}
	w.MobIndex[100] = mob

	input := "100 5 9 0 0 0 120 80 6 20\n0\n"
	sc := NewScanner(strings.NewReader(input), "test")
	loadShops(w, sc)

	if len(w.Shops) != 1 {
		t.Fatalf("Shops = %d, want 1", len(w.Shops))
	}
	shop := w.Shops[0]
	if shop.Keeper != 100 {
		t.Errorf("Keeper = %d, want 100", shop.Keeper)
	}
	if shop.BuyType[0] != 5 {
		t.Errorf("BuyType[0] = %d, want 5", shop.BuyType[0])
	}
	if shop.BuyType[1] != 9 {
		t.Errorf("BuyType[1] = %d, want 9", shop.BuyType[1])
	}
	if shop.OpenHour != 6 {
		t.Errorf("OpenHour = %d, want 6", shop.OpenHour)
	}
	if shop.CloseHour != 20 {
		t.Errorf("CloseHour = %d, want 20", shop.CloseHour)
	}
	// Verify mob got the shop reference
	if mob.Shop != shop {
		t.Error("mob.Shop should point to loaded shop")
	}
}

func TestLoadShops_ProfitClamping(t *testing.T) {
	w := world.New("")
	mob := &types.MobIndexData{Vnum: 200, PlayerName: "badshop"}
	w.MobIndex[200] = mob

	// profitBuy=50, profitSell=80 -> buy < sell+5, clamped
	input := "200 0 0 0 0 0 50 80 0 24\n0\n"
	sc := NewScanner(strings.NewReader(input), "test")
	loadShops(w, sc)

	if len(w.Shops) != 1 {
		t.Fatalf("Shops = %d, want 1", len(w.Shops))
	}
	shop := w.Shops[0]
	// profitSell=80, profitBuy should be at least sell+5=85
	if shop.ProfitBuy < shop.ProfitSell+5 {
		t.Errorf("ProfitBuy=%d should be >= ProfitSell+5=%d", shop.ProfitBuy, shop.ProfitSell+5)
	}
}

func TestLoadShops_MissingMob(t *testing.T) {
	w := world.New("")
	// No mob 999 exists — shop should still load, just with a bug message
	input := "999 0 0 0 0 0 120 80 6 20\n0\n"
	sc := NewScanner(strings.NewReader(input), "test")
	loadShops(w, sc)

	if len(w.Shops) != 1 {
		t.Fatalf("Shops = %d, want 1", len(w.Shops))
	}
}

// --------------- loadSpecials ---------------

func TestLoadSpecials(t *testing.T) {
	w := world.New("")
	mob := &types.MobIndexData{Vnum: 100, PlayerName: "guard"}
	w.MobIndex[100] = mob

	input := "M 100 spec_guard rest\n* comment\nS\n"
	sc := NewScanner(strings.NewReader(input), "test")
	loadSpecials(w, sc)

	if mob.SpecFun != "spec_guard" {
		t.Errorf("SpecFun = %q, want %q", mob.SpecFun, "spec_guard")
	}
}

func TestLoadSpecials_MissingMob(t *testing.T) {
	w := world.New("")
	input := "M 999 spec_thief rest\nS\n"
	sc := NewScanner(strings.NewReader(input), "test")
	loadSpecials(w, sc) // should not panic
}

func TestLoadSpecials_UnknownLetter(t *testing.T) {
	w := world.New("")
	input := "X unknown line\nS\n"
	sc := NewScanner(strings.NewReader(input), "test")
	loadSpecials(w, sc) // should not panic
}

func TestLoadSpecials_Empty(t *testing.T) {
	w := world.New("")
	input := "S\n"
	sc := NewScanner(strings.NewReader(input), "test")
	loadSpecials(w, sc)
}

// --------------- convertPosition ---------------

func TestConvertPosition(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{100, 0},   // >= 100: subtract 100
		{105, 5},
		{112, 12},
		{5, 6},     // old position mappings
		{6, 8},
		{7, 9},
		{8, 12},
		{9, 13},
		{10, 14},
		{11, 15},
		{0, 0},     // default: return as-is
		{1, 1},
		{2, 2},
		{3, 3},
		{4, 4},
	}
	for _, tt := range tests {
		got := convertPosition(tt.input)
		if got != tt.want {
			t.Errorf("convertPosition(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

// --------------- capitalizeFirst ---------------

func TestCapitalizeFirst(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "Hello"},
		{"world", "World"},
		{"", ""},
		{"Hello", "Hello"},  // already capitalized
		{"a", "A"},
		{"123", "123"},       // non-letter
	}
	for _, tt := range tests {
		got := capitalizeFirst(tt.input)
		if got != tt.want {
			t.Errorf("capitalizeFirst(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --------------- mprogNameToType ---------------

func TestMprogNameToType(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{"act_prog", types.MPROG_ACT},
		{"speech_prog", types.MPROG_SPEECH},
		{"rand_prog", types.MPROG_RAND},
		{"fight_prog", types.MPROG_FIGHT},
		{"death_prog", types.MPROG_DEATH},
		{"hitprcnt_prog", types.MPROG_HITPRCNT},
		{"entry_prog", types.MPROG_ENTRY},
		{"greet_prog", types.MPROG_GREET},
		{"all_greet_prog", types.MPROG_ALL_GREET},
		{"give_prog", types.MPROG_GIVE},
		{"bribe_prog", types.MPROG_BRIBE},
		{"hour_prog", types.MPROG_HOUR},
		{"time_prog", types.MPROG_TIME},
		{"wear_prog", types.MPROG_WEAR},
		{"remove_prog", types.MPROG_REMOVE},
		{"sac_prog", types.MPROG_SAC},
		{"look_prog", types.MPROG_LOOK},
		{"exa_prog", types.MPROG_EXA},
		{"zap_prog", types.MPROG_ZAP},
		{"get_prog", types.MPROG_GET},
		{"drop_prog", types.MPROG_DROP},
		{"damage_prog", types.MPROG_DAMAGE},
		{"repair_prog", types.MPROG_REPAIR},
		{"randiw_prog", types.MPROG_RANDIW},
		{"speechiw_prog", types.MPROG_SPEECHIW},
		{"pull_prog", types.MPROG_PULL},
		{"push_prog", types.MPROG_PUSH},
		{"sleep_prog", types.MPROG_SLEEP},
		{"rest_prog", types.MPROG_REST},
		{"leave_prog", types.MPROG_LEAVE},
		{"script_prog", types.MPROG_SCRIPT},
		{"use_prog", types.MPROG_USE},
	}
	for _, tt := range tests {
		got := mprogNameToType(tt.name)
		if got != tt.want {
			t.Errorf("mprogNameToType(%q) = %d, want %d", tt.name, got, tt.want)
		}
	}
}

func TestMprogNameToType_CaseInsensitive(t *testing.T) {
	got := mprogNameToType("SPEECH_PROG")
	if got != types.MPROG_SPEECH {
		t.Errorf("mprogNameToType(SPEECH_PROG) = %d, want %d", got, types.MPROG_SPEECH)
	}
}

func TestMprogNameToType_WithSpaces(t *testing.T) {
	got := mprogNameToType("  greet_prog  ")
	if got != types.MPROG_GREET {
		t.Errorf("mprogNameToType with spaces = %d, want %d", got, types.MPROG_GREET)
	}
}

func TestMprogNameToType_Unknown(t *testing.T) {
	got := mprogNameToType("unknown_prog")
	if got != 0 {
		t.Errorf("mprogNameToType(unknown) = %d, want 0", got)
	}
}

// --------------- loadRepairs ---------------

func TestLoadRepairs(t *testing.T) {
	w := world.New("")
	mob := &types.MobIndexData{Vnum: 100, PlayerName: "repairman"}
	w.MobIndex[100] = mob

	input := "100 5 9 0 120 0 6 20\n0\n"
	sc := NewScanner(strings.NewReader(input), "test")
	loadRepairs(w, sc)

	if len(w.Repairs) != 1 {
		t.Fatalf("Repairs = %d, want 1", len(w.Repairs))
	}
	rshop := w.Repairs[0]
	if rshop.Keeper != 100 {
		t.Errorf("Keeper = %d, want 100", rshop.Keeper)
	}
	if rshop.FixType[0] != 5 {
		t.Errorf("FixType[0] = %d, want 5", rshop.FixType[0])
	}
	if rshop.FixType[1] != 9 {
		t.Errorf("FixType[1] = %d, want 9", rshop.FixType[1])
	}
	if mob.RShop != rshop {
		t.Error("mob.RShop should point to loaded repair shop")
	}
}

func TestLoadRepairs_MissingMob(t *testing.T) {
	w := world.New("")
	input := "999 0 0 0 120 0 6 20\n0\n"
	sc := NewScanner(strings.NewReader(input), "test")
	loadRepairs(w, sc) // should not panic
	if len(w.Repairs) != 1 {
		t.Fatalf("Repairs = %d, want 1", len(w.Repairs))
	}
}

// --------------- skipPlayerObject ---------------

func TestSkipPlayerObject(t *testing.T) {
	// Nested #OBJECT sections should all be skipped
	input := `Vnum 100
Name test object~
WearLoc -1
End
`
	sc := NewScanner(strings.NewReader(input), "test")
	skipPlayerObject(sc) // should consume everything up to End
}

func TestSkipPlayerObject_Nested(t *testing.T) {
	input := `Vnum 100
Name container~
#OBJECT
Vnum 200
Name inner item~
End
End
`
	sc := NewScanner(strings.NewReader(input), "test")
	skipPlayerObject(sc) // should consume both nested objects
}

func TestLoadAreas_PathTraversal(t *testing.T) {
	// Create a temp directory with an area.lst that contains a path traversal entry
	tmpDir := t.TempDir()
	lst := "../evil.are\n$\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "area.lst"), []byte(lst), 0644); err != nil {
		t.Fatal(err)
	}

	w := world.New("")
	// Should not panic and should skip the traversal entry (no .are file exists)
	err := LoadAreas(w, tmpDir)
	if err != nil {
		t.Fatalf("LoadAreas should not return error: %v", err)
	}
	// No areas should have been loaded
	if len(w.Areas) != 0 {
		t.Errorf("Areas = %d, want 0", len(w.Areas))
	}
}

func TestLoadAreas_NonAreFileSkipped(t *testing.T) {
	tmpDir := t.TempDir()
	lst := "malicious.txt\n$\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "area.lst"), []byte(lst), 0644); err != nil {
		t.Fatal(err)
	}

	w := world.New("")
	err := LoadAreas(w, tmpDir)
	if err != nil {
		t.Fatalf("LoadAreas should not return error: %v", err)
	}
	if len(w.Areas) != 0 {
		t.Errorf("Areas = %d, want 0 (non-.are file should be skipped)", len(w.Areas))
	}
}
