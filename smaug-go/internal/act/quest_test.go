package act

import (
	"fmt"
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// setupQuestWorld creates a world with a quest hall room, a questmaster NPC,
// and some mob templates suitable for quest generation.
func setupQuestWorld() (*world.World, *types.RoomIndexData, *types.CharData) {
	w := setupWizWorld()

	room := &types.RoomIndexData{Vnum: 9200, Name: "Quest Hall"}
	w.Rooms[9200] = room

	// Create questmaster NPC
	qm := &types.CharData{
		Name:       "questmaster",
		ShortDescr: "the questmaster",
		Level:      50,
	}
	qm.Act.Set(types.ACT_IS_NPC)
	qm.Act.Set(types.ACT_QUESTMASTER)
	handler.CharToRoom(qm, room)
	w.AddChar(qm)

	// Add mob templates in level range suitable for a level-10 player
	for i := 0; i < 5; i++ {
		w.MobIndex[3000+i] = &types.MobIndexData{
			Vnum:       3000 + i,
			ShortDescr: fmt.Sprintf("mob %d", i),
			Level:      8 + i, // levels 8-12
		}
	}

	return w, room, qm
}

// --- DoQuest no argument ---

func TestDoQuest_NoArg(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 9200, Name: "Room"}

	DoQuest(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Syntax") {
		t.Errorf("expected usage syntax, got: %q", out)
	}
}

// --- DoQuest points ---

func TestDoQuest_Points(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 9200, Name: "Room"}
	ch.QuestPoints = 42

	DoQuest(ch, "points")
	out := readOutput(ch, client)
	if !strings.Contains(out, "42") {
		t.Errorf("expected '42' quest points in output, got: %q", out)
	}
}

// --- DoQuest time ---

func TestDoQuest_Time_NotOnQuest(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 9200, Name: "Room"}
	ch.Countdown = 0
	ch.NextQuest = 0

	DoQuest(ch, "time")
	out := readOutput(ch, client)
	if !strings.Contains(out, "not on a quest") {
		t.Errorf("expected 'not on a quest', got: %q", out)
	}
}

func TestDoQuest_Time_OnQuest(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 9200, Name: "Room"}
	ch.Countdown = 20

	DoQuest(ch, "time")
	out := readOutput(ch, client)
	if !strings.Contains(out, "20") {
		t.Errorf("expected countdown '20' in output, got: %q", out)
	}
}

func TestDoQuest_Time_Waiting(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 9200, Name: "Room"}
	ch.Countdown = 0
	ch.NextQuest = 10

	DoQuest(ch, "time")
	out := readOutput(ch, client)
	if !strings.Contains(out, "10") {
		t.Errorf("expected wait time '10' in output, got: %q", out)
	}
}

// --- DoQuest info ---

func TestDoQuest_Info_NotOnQuest(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 9200, Name: "Room"}
	ch.QuestMob = 0
	ch.QuestObj = 0

	DoQuest(ch, "info")
	out := readOutput(ch, client)
	if !strings.Contains(out, "not on a quest") {
		t.Errorf("expected 'not on a quest', got: %q", out)
	}
}

func TestDoQuest_Info_MobQuest(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 9200, Name: "Room"}
	ch.QuestMob = 3001

	DoQuest(ch, "info")
	out := readOutput(ch, client)
	if !strings.Contains(out, "3001") {
		t.Errorf("expected mob vnum '3001' in output, got: %q", out)
	}
}

func TestDoQuest_Info_ObjQuest(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 9200, Name: "Room"}
	ch.QuestObj = 5001

	DoQuest(ch, "info")
	out := readOutput(ch, client)
	if !strings.Contains(out, "5001") {
		t.Errorf("expected obj vnum '5001' in output, got: %q", out)
	}
}

// --- DoQuest request ---

func TestDoQuest_Request_NoQuestmaster(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 9200, Name: "Room"}

	DoQuest(ch, "request")
	out := readOutput(ch, client)
	if !strings.Contains(out, "can't do that here") {
		t.Errorf("expected 'can't do that here', got: %q", out)
	}
}

func TestDoQuest_Request_AlreadyOnQuest(t *testing.T) {
	_, room, _ := setupQuestWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	handler.CharToRoom(ch, room)
	ch.Countdown = 20

	DoQuest(ch, "request")
	out := readOutput(ch, client)
	if !strings.Contains(out, "already on a quest") {
		t.Errorf("expected 'already on a quest', got: %q", out)
	}
}

func TestDoQuest_Request_MustWait(t *testing.T) {
	_, room, _ := setupQuestWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	handler.CharToRoom(ch, room)
	ch.NextQuest = 5

	DoQuest(ch, "request")
	out := readOutput(ch, client)
	if !strings.Contains(out, "wait") {
		t.Errorf("expected wait message, got: %q", out)
	}
}

func TestDoQuest_Request_Success(t *testing.T) {
	_, room, _ := setupQuestWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	DoQuest(ch, "request")
	out := readOutput(ch, client)

	if ch.QuestMob == 0 {
		t.Errorf("expected QuestMob to be set, got 0")
	}
	if ch.Countdown != 30 {
		t.Errorf("expected Countdown=30, got %d", ch.Countdown)
	}
	if !strings.Contains(out, "slay") {
		t.Errorf("expected quest assignment message with 'slay', got: %q", out)
	}
}

// --- DoQuest complete ---

func TestDoQuest_Complete_NoQuestmaster(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 9200, Name: "Room"}

	DoQuest(ch, "complete")
	out := readOutput(ch, client)
	if !strings.Contains(out, "can't do that here") {
		t.Errorf("expected 'can't do that here', got: %q", out)
	}
}

func TestDoQuest_Complete_NotDone(t *testing.T) {
	_, room, _ := setupQuestWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	handler.CharToRoom(ch, room)
	ch.QuestMob = 3001 // still alive

	DoQuest(ch, "complete")
	out := readOutput(ch, client)
	if !strings.Contains(out, "haven't completed") {
		t.Errorf("expected 'haven't completed', got: %q", out)
	}
}

func TestDoQuest_Complete_Success(t *testing.T) {
	_, room, _ := setupQuestWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	handler.CharToRoom(ch, room)
	ch.QuestMob = -1 // mob killed
	ch.Countdown = 15
	ch.QuestPoints = 100

	DoQuest(ch, "complete")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Congratulations") {
		t.Errorf("expected congratulations message, got: %q", out)
	}
	if ch.QuestPoints <= 100 {
		t.Errorf("expected quest points to increase from 100, got %d", ch.QuestPoints)
	}
	if ch.QuestMob != 0 {
		t.Errorf("expected QuestMob reset to 0, got %d", ch.QuestMob)
	}
	if ch.Countdown != 0 {
		t.Errorf("expected Countdown reset to 0, got %d", ch.Countdown)
	}
	if ch.NextQuest != 15 {
		t.Errorf("expected NextQuest=15, got %d", ch.NextQuest)
	}
}

// --- DoQuest list ---

func TestDoQuest_List(t *testing.T) {
	_, room, _ := setupQuestWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	DoQuest(ch, "list")
	out := readOutput(ch, client)
	if !strings.Contains(out, "gold") || !strings.Contains(out, "500") {
		t.Errorf("expected reward list with gold/500, got: %q", out)
	}
}

func TestDoQuest_List_NoQuestmaster(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 9200, Name: "Room"}

	DoQuest(ch, "list")
	out := readOutput(ch, client)
	if !strings.Contains(out, "can't do that here") {
		t.Errorf("expected 'can't do that here', got: %q", out)
	}
}

// --- DoQuest buy ---

func TestDoQuest_Buy_Gold(t *testing.T) {
	_, room, _ := setupQuestWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	handler.CharToRoom(ch, room)
	ch.QuestPoints = 600
	ch.Gold = 100

	DoQuest(ch, "buy gold")
	out := readOutput(ch, client)

	if ch.QuestPoints != 100 {
		t.Errorf("expected QuestPoints=100 after buying gold, got %d", ch.QuestPoints)
	}
	if ch.Gold != 10100 {
		t.Errorf("expected Gold=10100, got %d", ch.Gold)
	}
	if !strings.Contains(out, "10000 gold") {
		t.Errorf("expected confirmation of gold purchase, got: %q", out)
	}
}

func TestDoQuest_Buy_Practices(t *testing.T) {
	_, room, _ := setupQuestWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	handler.CharToRoom(ch, room)
	ch.QuestPoints = 300
	ch.Practice = 2

	DoQuest(ch, "buy practices")
	_ = readOutput(ch, client)

	if ch.QuestPoints != 50 {
		t.Errorf("expected QuestPoints=50, got %d", ch.QuestPoints)
	}
	if ch.Practice != 7 {
		t.Errorf("expected Practice=7, got %d", ch.Practice)
	}
}

func TestDoQuest_Buy_Hp(t *testing.T) {
	_, room, _ := setupQuestWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	handler.CharToRoom(ch, room)
	ch.QuestPoints = 1200

	DoQuest(ch, "buy hp")
	_ = readOutput(ch, client)

	if ch.QuestPoints != 200 {
		t.Errorf("expected QuestPoints=200, got %d", ch.QuestPoints)
	}
	if ch.MaxHit != 110 {
		t.Errorf("expected MaxHit=110, got %d", ch.MaxHit)
	}
}

func TestDoQuest_Buy_Mana(t *testing.T) {
	_, room, _ := setupQuestWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	handler.CharToRoom(ch, room)
	ch.QuestPoints = 1200

	DoQuest(ch, "buy mana")
	_ = readOutput(ch, client)

	if ch.QuestPoints != 200 {
		t.Errorf("expected QuestPoints=200, got %d", ch.QuestPoints)
	}
	if ch.MaxMana != 60 {
		t.Errorf("expected MaxMana=60, got %d", ch.MaxMana)
	}
}

func TestDoQuest_Buy_InsufficientQP(t *testing.T) {
	_, room, _ := setupQuestWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	handler.CharToRoom(ch, room)
	ch.QuestPoints = 10

	DoQuest(ch, "buy gold")
	out := readOutput(ch, client)

	if !strings.Contains(out, "not have enough quest points") {
		t.Errorf("expected insufficient quest points message, got: %q", out)
	}
	if ch.QuestPoints != 10 {
		t.Errorf("quest points should not change, got %d", ch.QuestPoints)
	}
}

func TestDoQuest_Buy_NoQuestmaster(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 9200, Name: "Room"}

	DoQuest(ch, "buy gold")
	out := readOutput(ch, client)
	if !strings.Contains(out, "can't do that here") {
		t.Errorf("expected 'can't do that here', got: %q", out)
	}
}

func TestDoQuest_Buy_InvalidReward(t *testing.T) {
	_, room, _ := setupQuestWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	handler.CharToRoom(ch, room)
	ch.QuestPoints = 1000

	DoQuest(ch, "buy nonsense")
	out := readOutput(ch, client)
	if !strings.Contains(out, "buy") {
		t.Errorf("expected buy usage/error message, got: %q", out)
	}
}

// --- QuestUpdate ---

func TestQuestUpdate_CountdownDecrement(t *testing.T) {
	w := setupWizWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 9200, Name: "Room"}
	ch.Countdown = 5
	ch.QuestMob = 3001
	w.AddChar(ch)

	QuestUpdate(w)

	if ch.Countdown != 4 {
		t.Errorf("expected Countdown=4, got %d", ch.Countdown)
	}
}

func TestQuestUpdate_QuestFailure(t *testing.T) {
	w := setupWizWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 9200, Name: "Room"}
	ch.Countdown = 1
	ch.QuestMob = 3001
	w.AddChar(ch)

	QuestUpdate(w)

	if ch.Countdown != 0 {
		t.Errorf("expected Countdown=0, got %d", ch.Countdown)
	}
	if ch.QuestMob != 0 {
		t.Errorf("expected QuestMob reset to 0, got %d", ch.QuestMob)
	}

	out := readOutput(ch, client)
	if !strings.Contains(out, "failed") {
		t.Errorf("expected failure message, got: %q", out)
	}
}

func TestQuestUpdate_NextQuestDecrement(t *testing.T) {
	w := setupWizWorld()
	ch, client := makeTestChar("Quester")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 9200, Name: "Room"}
	ch.NextQuest = 3
	w.AddChar(ch)

	QuestUpdate(w)

	if ch.NextQuest != 2 {
		t.Errorf("expected NextQuest=2, got %d", ch.NextQuest)
	}
}

func TestQuestUpdate_SkipsNPCs(t *testing.T) {
	w := setupWizWorld()
	npc := &types.CharData{Name: "mob", Level: 5}
	npc.Act.Set(types.ACT_IS_NPC)
	npc.Countdown = 10
	w.AddChar(npc)

	QuestUpdate(w)

	if npc.Countdown != 10 {
		t.Errorf("NPC countdown should not change, got %d", npc.Countdown)
	}
}

// --- generateQuest ---

func TestGenerateQuest_NoCandidates(t *testing.T) {
	w := setupWizWorld()
	// No mob templates in range
	w.MobIndex[9999] = &types.MobIndexData{Vnum: 9999, Level: 50, ShortDescr: "high mob"}

	ch, client := makeTestChar("Quester")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 9200, Name: "Room"}

	generateQuest(ch)
	out := readOutput(ch, client)

	if ch.QuestMob != 0 {
		t.Errorf("expected QuestMob=0 when no candidates, got %d", ch.QuestMob)
	}
	if !strings.Contains(out, "No suitable") {
		t.Errorf("expected 'No suitable' message, got: %q", out)
	}
}

func TestGenerateQuest_SelectsInRange(t *testing.T) {
	w := setupWizWorld()
	// Add mobs: one in range, one out of range
	w.MobIndex[3000] = &types.MobIndexData{Vnum: 3000, Level: 10, ShortDescr: "test mob"}
	w.MobIndex[3001] = &types.MobIndexData{Vnum: 3001, Level: 50, ShortDescr: "high mob"}

	ch, client := makeTestChar("Quester")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 9200, Name: "Room"}

	generateQuest(ch)
	_ = readOutput(ch, client)

	if ch.QuestMob != 3000 {
		t.Errorf("expected QuestMob=3000 (only in-range mob), got %d", ch.QuestMob)
	}
	if ch.Countdown != 30 {
		t.Errorf("expected Countdown=30, got %d", ch.Countdown)
	}
}
