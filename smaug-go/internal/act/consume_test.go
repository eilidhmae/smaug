package act

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func init() {
	if WorldRef == nil {
		WorldRef = world.New("../../db")
	}
}

// newTestCharWithDesc creates a test character with a descriptor for consume tests.
func newTestCharWithDesc() *types.CharData {
	ch := &types.CharData{
		Name:     "Tester",
		Level:    10,
		Position: types.POS_STANDING,
		PCData: &types.PCData{
			PagerLen:  24,
			Condition: [4]int{24, 24, 24, 0}, // half full
		},
		InRoom: &types.RoomIndexData{
			Vnum: 3001,
			Name: "Test Room",
		},
	}
	d := &types.DescriptorData{
		Character: ch,
		Connected: types.CON_PLAYING,
	}
	ch.Desc = d
	return ch
}

// makeCarriedObj creates an object carried by ch (not worn).
func makeCarriedObj(ch *types.CharData, name, shortDescr string, itemType int, values [6]int) *types.ObjData {
	obj := &types.ObjData{
		Name:       name,
		ShortDescr: shortDescr,
		ItemType:   itemType,
		Value:      values,
		CarriedBy:  ch,
		WearLoc:    types.WEAR_NONE,
	}
	ch.Carrying = append(ch.Carrying, obj)
	return obj
}

func TestGainCondition(t *testing.T) {
	tests := []struct {
		name     string
		initial  int
		gain     int
		cond     int
		expected int
	}{
		{"increase full", 20, 5, types.COND_FULL, 25},
		{"clamp at max", 45, 10, types.COND_FULL, types.MAX_COND_VAL},
		{"clamp at zero", 2, -5, types.COND_THIRST, 0},
		{"no change", 24, 0, types.COND_FULL, 24},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ch := newTestCharWithDesc()
			ch.PCData.Condition[tc.cond] = tc.initial
			GainCondition(ch, tc.cond, tc.gain)
			if ch.PCData.Condition[tc.cond] != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, ch.PCData.Condition[tc.cond])
			}
		})
	}
}

func TestGainCondition_IgnoresNPC(t *testing.T) {
	ch := newTestCharWithDesc()
	ch.Act.Set(types.ACT_IS_NPC)
	ch.PCData.Condition[types.COND_FULL] = 10
	GainCondition(ch, types.COND_FULL, 5)
	if ch.PCData.Condition[types.COND_FULL] != 10 {
		t.Error("should not modify NPC conditions")
	}
}

func TestGainCondition_IgnoresImmortal(t *testing.T) {
	ch := newTestCharWithDesc()
	ch.Level = types.LEVEL_IMMORTAL
	ch.PCData.Condition[types.COND_FULL] = 10
	GainCondition(ch, types.COND_FULL, -20)
	if ch.PCData.Condition[types.COND_FULL] != 10 {
		t.Error("should not modify immortal conditions")
	}
}

func TestDoEat_NoArg(t *testing.T) {
	ch := newTestCharWithDesc()
	DoEat(ch, "")
}

func TestDoEat_Food(t *testing.T) {
	ch := newTestCharWithDesc()
	makeCarriedObj(ch, "bread loaf", "a loaf of bread", types.ITEM_FOOD, [6]int{10, 10, 0, 0, 0, 0})

	initial := ch.PCData.Condition[types.COND_FULL]
	DoEat(ch, "bread")

	if ch.PCData.Condition[types.COND_FULL] <= initial {
		t.Error("eating should increase fullness")
	}
	if len(ch.Carrying) != 0 {
		t.Error("food should be removed from inventory after eating")
	}
}

func TestDoEat_NotEdible(t *testing.T) {
	ch := newTestCharWithDesc()
	makeCarriedObj(ch, "sword", "a sword", types.ITEM_WEAPON, [6]int{})

	DoEat(ch, "sword")
	if len(ch.Carrying) != 1 {
		t.Error("non-food items should not be consumed")
	}
}

func TestDoEat_TooFull(t *testing.T) {
	ch := newTestCharWithDesc()
	ch.PCData.Condition[types.COND_FULL] = 45
	makeCarriedObj(ch, "bread loaf", "a loaf of bread", types.ITEM_FOOD, [6]int{10, 10, 0, 0, 0, 0})

	DoEat(ch, "bread")
	if len(ch.Carrying) != 1 {
		t.Error("should not consume food when too full")
	}
}

func TestDoDrink_NoArg_NoFountain(t *testing.T) {
	ch := newTestCharWithDesc()
	DoDrink(ch, "")
}

func TestDoDrink_Fountain(t *testing.T) {
	ch := newTestCharWithDesc()
	fountain := &types.ObjData{
		Name:       "fountain",
		ShortDescr: "a marble fountain",
		ItemType:   types.ITEM_FOUNTAIN,
		InRoom:     ch.InRoom,
	}
	ch.InRoom.Contents = append(ch.InRoom.Contents, fountain)

	ch.PCData.Condition[types.COND_THIRST] = 10
	DoDrink(ch, "")

	if ch.PCData.Condition[types.COND_THIRST] <= 10 {
		t.Error("drinking from fountain should increase thirst satisfaction")
	}
}

func TestDoDrink_Container(t *testing.T) {
	ch := newTestCharWithDesc()
	flask := makeCarriedObj(ch, "flask water", "a flask of water", types.ITEM_DRINK_CON, [6]int{10, 5, 0, 0, 0, 0})

	DoDrink(ch, "flask")
	if flask.Value[1] != 4 {
		t.Errorf("expected liquid to decrease to 4, got %d", flask.Value[1])
	}
}

func TestDoDrink_EmptyContainer(t *testing.T) {
	ch := newTestCharWithDesc()
	makeCarriedObj(ch, "flask", "an empty flask", types.ITEM_DRINK_CON, [6]int{10, 0, 0, 0, 0, 0})

	DoDrink(ch, "flask")
}

func TestDoFill_Success(t *testing.T) {
	ch := newTestCharWithDesc()
	flask := makeCarriedObj(ch, "flask", "a flask", types.ITEM_DRINK_CON, [6]int{10, 0, 0, 0, 0, 0})

	fountain := &types.ObjData{
		Name:       "fountain",
		ShortDescr: "a fountain",
		ItemType:   types.ITEM_FOUNTAIN,
		Value:      [6]int{100, 100, 0, 0, 0, 0},
		InRoom:     ch.InRoom,
	}
	ch.InRoom.Contents = append(ch.InRoom.Contents, fountain)

	DoFill(ch, "flask")
	if flask.Value[1] != flask.Value[0] {
		t.Errorf("expected flask to be full (%d), got %d", flask.Value[0], flask.Value[1])
	}
}

func TestDoFill_AlreadyFull(t *testing.T) {
	ch := newTestCharWithDesc()
	makeCarriedObj(ch, "flask", "a flask", types.ITEM_DRINK_CON, [6]int{10, 10, 0, 0, 0, 0})

	DoFill(ch, "flask")
}

func TestDoEmpty_DrinkCon(t *testing.T) {
	ch := newTestCharWithDesc()
	flask := makeCarriedObj(ch, "flask water", "a flask of water", types.ITEM_DRINK_CON, [6]int{10, 5, 0, 0, 0, 0})

	DoEmpty(ch, "flask")
	if flask.Value[1] != 0 {
		t.Errorf("expected flask to be empty (0), got %d", flask.Value[1])
	}
}

func TestDoEmpty_AlreadyEmpty(t *testing.T) {
	ch := newTestCharWithDesc()
	makeCarriedObj(ch, "flask", "an empty flask", types.ITEM_DRINK_CON, [6]int{10, 0, 0, 0, 0, 0})

	DoEmpty(ch, "flask")
}
