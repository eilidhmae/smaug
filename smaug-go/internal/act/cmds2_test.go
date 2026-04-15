package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

// --- DoSplit ---

func TestDoSplit_NoArg(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	DoSplit(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Split") {
		t.Errorf("expected split usage, got: %q", out)
	}
}

func TestDoSplit_NotEnoughGold(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	room := &types.RoomIndexData{Vnum: 1000}
	placeInRoom(ch, room)
	ch.Gold = 10

	DoSplit(ch, "100")
	out := readOutput(ch, client)
	if !strings.Contains(out, "don't have that much") {
		t.Errorf("expected insufficient funds, got: %q", out)
	}
}

func TestDoSplit_AloneJustKeepIt(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	room := &types.RoomIndexData{Vnum: 1000}
	placeInRoom(ch, room)
	ch.Gold = 100

	DoSplit(ch, "50")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Just keep it") {
		t.Errorf("expected alone message, got: %q", out)
	}
	if ch.Gold != 100 {
		t.Errorf("gold should be unchanged, got %d", ch.Gold)
	}
}

func TestDoSplit_WithGroupMember(t *testing.T) {
	_ = setupGroupWorld()
	ch, chClient := makeTestChar("Alice")
	defer chClient.Close()
	mate, mateClient := makeTestChar("Bob")
	defer mateClient.Close()

	room := &types.RoomIndexData{Vnum: 1000}
	placeInRoom(ch, room)
	placeInRoom(mate, room)

	// Group Bob under Alice
	mate.Master = ch
	mate.Leader = ch

	ch.Gold = 100
	DoSplit(ch, "60")
	_ = readOutput(ch, chClient)
	_ = readOutput(mate, mateClient)

	// 60 split across 2 = 30 each, remainder 0. Alice: started 100, spent 60, +30. End: 70.
	if ch.Gold != 70 {
		t.Errorf("Alice gold = %d, want 70", ch.Gold)
	}
	if mate.Gold != 530 { // started 500
		t.Errorf("Bob gold = %d, want 530", mate.Gold)
	}
}

func TestDoSplit_NegativeAmount(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 1000}
	placeInRoom(ch, room)

	DoSplit(ch, "-50")
	out := readOutput(ch, client)
	if !strings.Contains(out, "wouldn't like") {
		t.Errorf("expected negative guard, got: %q", out)
	}
}

// --- DoLight ---

func TestDoLight_NoArg(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	DoLight(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Light what?") {
		t.Errorf("expected 'Light what?', got: %q", out)
	}
}

func TestDoLight_NotCarrying(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	DoLight(ch, "torch")
	out := readOutput(ch, client)
	if !strings.Contains(out, "aren't carrying") {
		t.Errorf("expected not carrying, got: %q", out)
	}
}

func TestDoLight_NotALight(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 1000}
	placeInRoom(ch, room)

	rock := &types.ObjData{
		Name:       "rock",
		ShortDescr: "a rock",
		ItemType:   types.ITEM_ARMOR,
		WearLoc:    types.WEAR_NONE,
	}
	handler.ObjToChar(rock, ch)

	DoLight(ch, "rock")
	out := readOutput(ch, client)
	if !strings.Contains(out, "can't light that") {
		t.Errorf("expected 'can't light that', got: %q", out)
	}
}

func TestDoLight_LightsIt(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 1000}
	placeInRoom(ch, room)

	torch := &types.ObjData{
		Name:       "torch",
		ShortDescr: "a torch",
		ItemType:   types.ITEM_LIGHT,
		WearLoc:    types.WEAR_NONE,
	}
	torch.Value[1] = 10 // Value[1] = fuel/hours per C misc.c:1917
	handler.ObjToChar(torch, ch)

	DoLight(ch, "torch")
	out := readOutput(ch, client)
	if !strings.Contains(out, "light") && !strings.Contains(out, "Light") {
		t.Errorf("expected light message, got: %q", out)
	}
	// Regression guard: PIPE_LIT bit on Value[3] must be set.
	if torch.Value[3]&int(types.PIPE_LIT) == 0 {
		t.Errorf("PIPE_LIT bit not set on Value[3]: %d", torch.Value[3])
	}
}

func TestDoLight_EmptyFuelRejected(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 1000}
	placeInRoom(ch, room)

	// Torch with zero fuel — Value[1]=0 per C must refuse to light.
	torch := &types.ObjData{
		Name:       "torch",
		ShortDescr: "a torch",
		ItemType:   types.ITEM_LIGHT,
		WearLoc:    types.WEAR_NONE,
	}
	torch.Value[1] = 0
	handler.ObjToChar(torch, ch)

	DoLight(ch, "torch")
	out := readOutput(ch, client)
	if !strings.Contains(out, "can't light that") {
		t.Errorf("expected 'can't light that' for empty fuel, got %q", out)
	}
	if torch.Value[3]&int(types.PIPE_LIT) != 0 {
		t.Errorf("PIPE_LIT must not be set when lighting fails")
	}
}

// DoSplit — regression: silver/copper buckets.

func TestDoSplit_Silver(t *testing.T) {
	_ = setupGroupWorld()
	ch, chClient := makeTestChar("Alice")
	defer chClient.Close()
	mate, mateClient := makeTestChar("Bob")
	defer mateClient.Close()

	room := &types.RoomIndexData{Vnum: 1000}
	placeInRoom(ch, room)
	placeInRoom(mate, room)
	mate.Master = ch
	mate.Leader = ch

	ch.Silver = 100
	bobSilver := mate.Silver
	aliceGold := ch.Gold
	DoSplit(ch, "60 silver")
	_ = readOutput(ch, chClient)
	_ = readOutput(mate, mateClient)

	if ch.Silver != 70 {
		t.Errorf("Alice silver = %d, want 70", ch.Silver)
	}
	if mate.Silver != bobSilver+30 {
		t.Errorf("Bob silver = %d, want %d", mate.Silver, bobSilver+30)
	}
	if ch.Gold != aliceGold {
		t.Errorf("Gold must not change during silver split, was %d got %d", aliceGold, ch.Gold)
	}
}

func TestDoSplit_RemainderToCh(t *testing.T) {
	_ = setupGroupWorld()
	ch, chClient := makeTestChar("Alice")
	defer chClient.Close()
	mate, mateClient := makeTestChar("Bob")
	defer mateClient.Close()

	room := &types.RoomIndexData{Vnum: 1000}
	placeInRoom(ch, room)
	placeInRoom(mate, room)
	mate.Master = ch
	mate.Leader = ch

	ch.Gold = 100
	bobGold := mate.Gold
	// 61 / 2 = 30 each, remainder 1 to ch.
	DoSplit(ch, "61")
	_ = readOutput(ch, chClient)
	_ = readOutput(mate, mateClient)

	// Alice: 100 - 61 + 30 + 1 = 70
	if ch.Gold != 70 {
		t.Errorf("Alice gold = %d, want 70 (share+remainder)", ch.Gold)
	}
	if mate.Gold != bobGold+30 {
		t.Errorf("Bob gold = %d, want %d (share only)", mate.Gold, bobGold+30)
	}
}

func TestDoSplit_InvalidCoinType(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 1000}
	placeInRoom(ch, room)
	ch.Gold = 100

	DoSplit(ch, "50 platinum")
	out := readOutput(ch, client)
	if !strings.Contains(out, "not a valid coin type") {
		t.Errorf("expected invalid-coin-type rejection, got %q", out)
	}
	if ch.Gold != 100 {
		t.Errorf("Gold must be unchanged on invalid coin, got %d", ch.Gold)
	}
}

// --- DoThrow ---

func TestDoThrow_NoArg(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	DoThrow(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Throw what?") {
		t.Errorf("expected 'Throw what?', got: %q", out)
	}
}

func TestDoThrow_NotCarrying(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 1000}
	placeInRoom(ch, room)

	DoThrow(ch, "knife")
	out := readOutput(ch, client)
	if !strings.Contains(out, "aren't carrying") {
		t.Errorf("expected not carrying, got: %q", out)
	}
}

func TestDoThrow_ClattersIntoRoom(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 1000}
	placeInRoom(ch, room)

	knife := &types.ObjData{
		Name:       "knife",
		ShortDescr: "a knife",
		ItemType:   types.ITEM_WEAPON,
		WearLoc:    types.WEAR_NONE,
	}
	handler.ObjToChar(knife, ch)

	DoThrow(ch, "knife")
	out := readOutput(ch, client)
	if !strings.Contains(out, "throw") && !strings.Contains(out, "Throw") {
		t.Errorf("expected throw message, got: %q", out)
	}
	// Must have left inventory
	for _, o := range ch.Carrying {
		if o == knife {
			t.Errorf("knife should no longer be in inventory")
		}
	}
}

// --- DoAlias / DoUnalias ---

func TestDoAlias_EmptyList(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	DoAlias(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "no aliases") {
		t.Errorf("expected empty-list message, got: %q", out)
	}
}

func TestDoAlias_Create(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	DoAlias(ch, "g get all corpse")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Created") && !strings.Contains(out, "created") {
		t.Errorf("expected created msg, got: %q", out)
	}
	found := false
	for _, a := range ch.PCData.Aliases {
		if a.Name == "g" && a.Cmd == "get all corpse" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("alias not stored: %+v", ch.PCData.Aliases)
	}
}

func TestDoAlias_List(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	ch.PCData.Aliases = append(ch.PCData.Aliases, &types.AliasData{Name: "k", Cmd: "kill"})
	DoAlias(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "k") || !strings.Contains(out, "kill") {
		t.Errorf("expected alias listed, got: %q", out)
	}
}

func TestDoAlias_RejectTilde(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	DoAlias(ch, "bad ~evil")
	out := readOutput(ch, client)
	if !strings.Contains(out, "~") {
		t.Errorf("expected tilde rejection, got: %q", out)
	}
}

func TestDoUnalias(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	ch.PCData.Aliases = append(ch.PCData.Aliases, &types.AliasData{Name: "g", Cmd: "get all"})
	DoUnalias(ch, "g")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Deleted") && !strings.Contains(out, "Removed") && !strings.Contains(out, "removed") {
		t.Errorf("expected delete msg, got: %q", out)
	}
	if len(ch.PCData.Aliases) != 0 {
		t.Errorf("alias not removed: %+v", ch.PCData.Aliases)
	}
}

func TestDoUnalias_NotFound(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	DoUnalias(ch, "nothing")
	out := readOutput(ch, client)
	if !strings.Contains(out, "does not exist") && !strings.Contains(out, "No such") {
		t.Errorf("expected not-found msg, got: %q", out)
	}
}

// --- DoAreas ---

func TestDoAreas_EmptyWorld(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	DoAreas(ch, "")
	out := readOutput(ch, client)
	// Should at least print a header even with zero areas
	if !strings.Contains(strings.ToLower(out), "area") {
		t.Errorf("expected areas output header, got: %q", out)
	}
}

func TestDoAreas_WithAreas(t *testing.T) {
	w := setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	w.Areas = append(w.Areas, &types.AreaData{
		Name:         "Midgaard",
		Filename:     "midgaard.are",
		Author:       "Furey",
		LowSoftRange: 1,
		HiSoftRange:  50,
	})

	DoAreas(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Midgaard") {
		t.Errorf("expected area 'Midgaard', got: %q", out)
	}
	if !strings.Contains(out, "Furey") {
		t.Errorf("expected author 'Furey', got: %q", out)
	}
}

// --- DoAltscore ---

func TestDoAltscore(t *testing.T) {
	w := setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()
	// Races/Classes must be valid indices.
	w.Classes = append(w.Classes, &types.ClassType{WhoName: "Mage"})
	w.Races = append(w.Races, &types.RaceData{Name: "Human"})

	ch.Class = 0
	ch.Race = 0
	ch.Hit = 42
	ch.MaxHit = 100

	DoAltscore(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Alice") {
		t.Errorf("expected name in altscore, got: %q", out)
	}
	if !strings.Contains(out, "42") {
		t.Errorf("expected HP 42, got: %q", out)
	}
}

// --- DoColor ---

func TestDoColor_Toggle(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	// Default: PLR_ANSI not set.
	DoColor(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(strings.ToLower(out), "on") {
		t.Errorf("expected color ON after first toggle, got: %q", out)
	}
	if !ch.Act.IsSet(types.PLR_ANSI) {
		t.Errorf("PLR_ANSI should be set after toggle on")
	}

	DoColor(ch, "")
	out = readOutput(ch, client)
	if !strings.Contains(strings.ToLower(out), "off") {
		t.Errorf("expected color OFF after second toggle, got: %q", out)
	}
	if ch.Act.IsSet(types.PLR_ANSI) {
		t.Errorf("PLR_ANSI should be cleared after toggle off")
	}
}

func TestDoColor_On(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	DoColor(ch, "on")
	out := readOutput(ch, client)
	if !ch.Act.IsSet(types.PLR_ANSI) {
		t.Errorf("PLR_ANSI should be set, got output: %q", out)
	}
}

func TestDoColor_Off(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()
	ch.Act.Set(types.PLR_ANSI)

	DoColor(ch, "off")
	_ = readOutput(ch, client)
	if ch.Act.IsSet(types.PLR_ANSI) {
		t.Errorf("PLR_ANSI should be cleared")
	}
}

// --- DoCompress ---

func TestDoCompress_Toggle(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()
	ch.Desc.TelnetState = &types.TelnetState{}

	DoCompress(ch, "")
	out := readOutput(ch, client)
	if !ch.Desc.TelnetState.MCCPEnabled {
		t.Errorf("MCCP should be enabled after first toggle, out=%q", out)
	}

	DoCompress(ch, "")
	_ = readOutput(ch, client)
	if ch.Desc.TelnetState.MCCPEnabled {
		t.Errorf("MCCP should be disabled after second toggle")
	}
}
