package persist

import (
	"bytes"
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func TestSaveArea_Basic(t *testing.T) {
	w := world.New("/tmp/test")

	area := &types.AreaData{
		Name:           "Test Area",
		Author:         "Builder",
		Filename:       "test.are",
		ResetMsg:       "The area resets.",
		LowRVnum:       100,
		HiRVnum:        199,
		LowMVnum:       100,
		HiMVnum:        199,
		LowOVnum:       100,
		HiOVnum:        199,
		ResetFrequency: 15,
	}
	w.Areas = append(w.Areas, area)

	// Add a room
	room := &types.RoomIndexData{
		Vnum:        100,
		Name:        "A test room",
		Description: "This is a test room.\n\r",
		SectorType:  types.SECT_INSIDE,
		Area:        area,
	}
	w.Rooms[100] = room

	// Add a mob
	mobIdx := &types.MobIndexData{
		Vnum:        100,
		PlayerName:  "guard",
		ShortDescr:  "a guard",
		LongDescr:   "A guard stands here.\n\r",
		Level:       10,
		Position:    types.POS_STANDING,
		DefPosition: types.POS_STANDING,
	}
	mobIdx.Act.Set(types.ACT_IS_NPC)
	w.MobIndex[100] = mobIdx

	// Add an object
	objIdx := &types.ObjIndexData{
		Vnum:        100,
		Name:        "sword steel",
		ShortDescr:  "a steel sword",
		Description: "A steel sword lies here.",
		ItemType:    types.ITEM_WEAPON,
		Weight:      5,
		GoldCost:    100,
		Value:       [6]int{0, 4, 8, 0, 0, 0},
	}
	w.ObjIndex[100] = objIdx

	// Add a reset
	area.Resets = append(area.Resets, &types.ResetData{
		Command: 'M',
		Arg1:    100,
		Arg2:    1,
		Arg3:    100,
	})

	var buf bytes.Buffer
	err := SaveArea(&buf, w, area)
	if err != nil {
		t.Fatalf("SaveArea: %v", err)
	}

	output := buf.String()

	// Verify key sections are present
	checks := []string{
		"#AREA", "Test Area~",
		"#AUTHOR", "Builder~",
		"#MOBILES", "#100", "guard~",
		"#OBJECTS", "sword steel~",
		"#ROOMS", "A test room~",
		"#RESETS", "M",
		"#$",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("output should contain %q", check)
		}
	}
}

func TestSaveArea_EmptyArea(t *testing.T) {
	w := world.New("/tmp/test")
	area := &types.AreaData{
		Name:     "Empty",
		Filename: "empty.are",
	}

	var buf bytes.Buffer
	err := SaveArea(&buf, w, area)
	if err != nil {
		t.Fatalf("SaveArea: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "#AREA") {
		t.Error("should contain #AREA")
	}
	if !strings.Contains(output, "#$") {
		t.Error("should contain end marker")
	}
}

func TestSaveArea_WithVersion(t *testing.T) {
	w := world.New("/tmp/test")
	area := &types.AreaData{
		Name:     "Versioned",
		Filename: "ver.are",
		Version:  3,
	}

	var buf bytes.Buffer
	err := SaveArea(&buf, w, area)
	if err != nil {
		t.Fatalf("SaveArea: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "#VERSION") {
		t.Error("should contain #VERSION section")
	}
	if !strings.Contains(output, "3") {
		t.Error("should contain version number 3")
	}
}

func TestSaveArea_NoOptionalSections(t *testing.T) {
	w := world.New("/tmp/test")
	area := &types.AreaData{
		Name:     "Minimal",
		Filename: "min.are",
	}

	var buf bytes.Buffer
	err := SaveArea(&buf, w, area)
	if err != nil {
		t.Fatalf("SaveArea: %v", err)
	}

	output := buf.String()
	// No author set, so #AUTHOR should not appear
	if strings.Contains(output, "#AUTHOR") {
		t.Error("should not contain #AUTHOR when Author is empty")
	}
	// No reset message set
	if strings.Contains(output, "#RESETMSG") {
		t.Error("should not contain #RESETMSG when ResetMsg is empty")
	}
	// No version set
	if strings.Contains(output, "#VERSION") {
		t.Error("should not contain #VERSION when Version is 0")
	}
}

func TestSaveArea_RoomWithExits(t *testing.T) {
	w := world.New("/tmp/test")
	area := &types.AreaData{
		Name:     "Exits",
		Filename: "exits.are",
		LowRVnum: 100,
		HiRVnum:  199,
		LowMVnum: 100,
		HiMVnum:  199,
		LowOVnum: 100,
		HiOVnum:  199,
	}

	destRoom := &types.RoomIndexData{Vnum: 101, Name: "Dest Room", Area: area}
	room := &types.RoomIndexData{
		Vnum:        100,
		Name:        "Start Room",
		Description: "The starting room.",
		SectorType:  types.SECT_CITY,
		Area:        area,
		Exits: []*types.ExitData{
			{
				Direction:   0, // north
				Description: "A door to the north.",
				Keyword:     "door",
				ToRoom:      destRoom,
				Key:         200,
				ExitInfo:    1,
			},
			nil, // south (no exit)
		},
	}
	w.Rooms[100] = room
	w.Rooms[101] = destRoom

	var buf bytes.Buffer
	err := SaveArea(&buf, w, area)
	if err != nil {
		t.Fatalf("SaveArea: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "D0") {
		t.Error("should contain D0 for north exit")
	}
	if !strings.Contains(output, "door~") {
		t.Error("should contain exit keyword")
	}
	if !strings.Contains(output, "1 200 101") {
		t.Error("should contain exit info, key, and destination vnum")
	}
}

func TestSaveArea_RoomWithExtraDescr(t *testing.T) {
	w := world.New("/tmp/test")
	area := &types.AreaData{
		Name:     "ED",
		Filename: "ed.are",
		LowRVnum: 100,
		HiRVnum:  199,
		LowMVnum: 100,
		HiMVnum:  199,
		LowOVnum: 100,
		HiOVnum:  199,
	}

	room := &types.RoomIndexData{
		Vnum: 100,
		Name: "Desc Room",
		Area: area,
		ExtraDescr: []*types.ExtraDescrData{
			{Keyword: "painting", Description: "A beautiful oil painting."},
		},
	}
	w.Rooms[100] = room

	var buf bytes.Buffer
	err := SaveArea(&buf, w, area)
	if err != nil {
		t.Fatalf("SaveArea: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "E\npainting~") {
		t.Error("should contain room extra description")
	}
	if !strings.Contains(output, "A beautiful oil painting.~") {
		t.Error("should contain extra description text")
	}
}

func TestSaveArea_MobWithMudProgs(t *testing.T) {
	w := world.New("/tmp/test")
	area := &types.AreaData{
		Name:     "MudProg",
		Filename: "mp.are",
		LowMVnum: 100,
		HiMVnum:  199,
		LowOVnum: 100,
		HiOVnum:  199,
	}

	mob := &types.MobIndexData{
		Vnum:        100,
		PlayerName:  "prog_mob",
		ShortDescr:  "a programmed mob",
		LongDescr:   "A mob with scripts.\n\r",
		Level:       5,
		Position:    types.POS_STANDING,
		DefPosition: types.POS_STANDING,
		MudProgs: []*types.MProgData{
			{Type: 1, ArgList: "100", ComList: "mpecho Hello!"},
		},
	}
	mob.Act.Set(types.ACT_IS_NPC)
	w.MobIndex[100] = mob

	var buf bytes.Buffer
	err := SaveArea(&buf, w, area)
	if err != nil {
		t.Fatalf("SaveArea: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, ">1 100~") {
		t.Error("should contain mudprog trigger line")
	}
	if !strings.Contains(output, "mpecho Hello!~") {
		t.Error("should contain mudprog command list")
	}
}

func TestSaveArea_ObjectWithAffects(t *testing.T) {
	w := world.New("/tmp/test")
	area := &types.AreaData{
		Name:     "ObjAff",
		Filename: "oa.are",
		LowMVnum: 100,
		HiMVnum:  199,
		LowOVnum: 100,
		HiOVnum:  199,
	}

	obj := &types.ObjIndexData{
		Vnum:        100,
		Name:        "magic ring",
		ShortDescr:  "a magic ring",
		Description: "A glowing ring.",
		ItemType:    types.ITEM_ARMOR,
		Weight:      1,
		GoldCost:    500,
		Affects: []*types.AffectData{
			{Location: types.APPLY_AC, Modifier: -5},
		},
		ExtraDescr: []*types.ExtraDescrData{
			{Keyword: "ring", Description: "It glows faintly."},
		},
	}
	w.ObjIndex[100] = obj

	var buf bytes.Buffer
	err := SaveArea(&buf, w, area)
	if err != nil {
		t.Fatalf("SaveArea: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "A\n") {
		t.Error("should contain affect marker 'A'")
	}
	if !strings.Contains(output, "E\nring~") {
		t.Error("should contain object extra description")
	}
}

func TestSaveArea_MultipleResets(t *testing.T) {
	w := world.New("/tmp/test")
	area := &types.AreaData{
		Name:     "Resets",
		Filename: "resets.are",
		Resets: []*types.ResetData{
			{Command: 'M', Extra: 0, Arg1: 100, Arg2: 1, Arg3: 100},
			{Command: 'O', Extra: 0, Arg1: 200, Arg2: 1, Arg3: 100},
			{Command: 'G', Extra: 0, Arg1: 300, Arg2: 1, Arg3: 0},
			{Command: 'E', Extra: 0, Arg1: 400, Arg2: 1, Arg3: 16},
			{Command: 'D', Extra: 0, Arg1: 100, Arg2: 0, Arg3: 1},
		},
	}

	var buf bytes.Buffer
	err := SaveArea(&buf, w, area)
	if err != nil {
		t.Fatalf("SaveArea: %v", err)
	}

	output := buf.String()
	// Verify each reset command type appears
	for _, cmd := range []string{"M ", "O ", "G ", "E ", "D "} {
		if !strings.Contains(output, cmd) {
			t.Errorf("should contain reset command '%s'", cmd)
		}
	}
	if !strings.Contains(output, "#RESETS") {
		t.Error("should contain #RESETS section")
	}
}

func TestSaveArea_ShopSection(t *testing.T) {
	w := world.New("/tmp/test")
	area := &types.AreaData{
		Name:     "Shop",
		Filename: "shop.are",
		LowMVnum: 100,
		HiMVnum:  199,
		LowOVnum: 100,
		HiOVnum:  199,
	}

	shop := &types.ShopData{
		Keeper:     100,
		ProfitBuy:  120,
		ProfitSell: 80,
		OpenHour:   6,
		CloseHour:  20,
	}
	shop.BuyType[0] = types.ITEM_WEAPON
	shop.BuyType[1] = types.ITEM_ARMOR

	mob := &types.MobIndexData{
		Vnum:        100,
		PlayerName:  "shopkeeper",
		ShortDescr:  "a shopkeeper",
		LongDescr:   "A shopkeeper stands here.\n\r",
		Level:       20,
		Position:    types.POS_STANDING,
		DefPosition: types.POS_STANDING,
		Shop:        shop,
	}
	mob.Act.Set(types.ACT_IS_NPC)
	w.MobIndex[100] = mob

	var buf bytes.Buffer
	err := SaveArea(&buf, w, area)
	if err != nil {
		t.Fatalf("SaveArea: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "#SHOPS") {
		t.Error("should contain #SHOPS section")
	}
	if !strings.Contains(output, "100 ") {
		t.Error("should contain shop keeper vnum")
	}
	if !strings.Contains(output, "120 80 6 20") {
		t.Error("should contain shop profit/hours")
	}
}

func TestSaveArea_RoomWithMudProgs(t *testing.T) {
	w := world.New("/tmp/test")
	area := &types.AreaData{
		Name:     "RoomProg",
		Filename: "rp.are",
		LowRVnum: 100,
		HiRVnum:  199,
		LowMVnum: 100,
		HiMVnum:  199,
		LowOVnum: 100,
		HiOVnum:  199,
	}

	room := &types.RoomIndexData{
		Vnum: 100,
		Name: "Prog Room",
		Area: area,
		MudProgs: []*types.MProgData{
			{Type: 2, ArgList: "hello", ComList: "mpecho Welcome!"},
		},
	}
	w.Rooms[100] = room

	var buf bytes.Buffer
	err := SaveArea(&buf, w, area)
	if err != nil {
		t.Fatalf("SaveArea: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, ">2 hello~") {
		t.Error("should contain room mudprog trigger")
	}
}

func TestSaveArea_RoomWithRVnumExit(t *testing.T) {
	// Exit with no ToRoom pointer, uses RVnum
	w := world.New("/tmp/test")
	area := &types.AreaData{
		Name:     "RVnum",
		Filename: "rv.are",
		LowRVnum: 100,
		HiRVnum:  199,
		LowMVnum: 100,
		HiMVnum:  199,
		LowOVnum: 100,
		HiOVnum:  199,
	}

	room := &types.RoomIndexData{
		Vnum: 100,
		Name: "RVnum Room",
		Area: area,
		Exits: []*types.ExitData{
			{Direction: 2, RVnum: 999, Description: "", Keyword: ""},
		},
	}
	w.Rooms[100] = room

	var buf bytes.Buffer
	err := SaveArea(&buf, w, area)
	if err != nil {
		t.Fatalf("SaveArea: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "0 0 999") {
		t.Error("should contain RVnum 999 as destination")
	}
}

func TestSaveArea_Ranges(t *testing.T) {
	w := world.New("/tmp/test")
	area := &types.AreaData{
		Name:          "Ranges",
		Filename:      "ranges.are",
		LowSoftRange:  5,
		HiSoftRange:   15,
		LowHardRange:  1,
		HiHardRange:   65,
	}

	var buf bytes.Buffer
	err := SaveArea(&buf, w, area)
	if err != nil {
		t.Fatalf("SaveArea: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "#RANGES") {
		t.Error("should contain #RANGES section")
	}
	if !strings.Contains(output, "5 15 1 65") {
		t.Error("should contain range values")
	}
}

func TestSaveArea_Economy(t *testing.T) {
	w := world.New("/tmp/test")
	area := &types.AreaData{
		Name:        "Economy",
		Filename:    "econ.are",
		HighEconomy: 5000,
		LowEconomy:  100,
	}

	var buf bytes.Buffer
	err := SaveArea(&buf, w, area)
	if err != nil {
		t.Fatalf("SaveArea: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "#ECONOMY") {
		t.Error("should contain #ECONOMY section")
	}
	if !strings.Contains(output, "5000 100") {
		t.Error("should contain economy values")
	}
}
