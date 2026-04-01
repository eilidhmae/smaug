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
