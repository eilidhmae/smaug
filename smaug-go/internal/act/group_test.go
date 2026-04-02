package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func setupGroupWorld() *world.World {
	w := world.New("/tmp/test")
	WorldRef = w
	return w
}

func placeInRoom(ch *types.CharData, room *types.RoomIndexData) {
	handler.CharToRoom(ch, room)
}

// --- DoFollow ---

func TestDoFollow_NoArg(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	DoFollow(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Follow whom?") {
		t.Errorf("expected 'Follow whom?', got: %q", out)
	}
}

func TestDoFollow_TargetNotFound(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	room := &types.RoomIndexData{Vnum: 1000, Name: "Test Room"}
	placeInRoom(ch, room)

	DoFollow(ch, "nobody")
	out := readOutput(ch, client)
	if !strings.Contains(out, "aren't here") {
		t.Errorf("expected 'aren't here', got: %q", out)
	}
}

func TestDoFollow_SelfNotFollowing(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	room := &types.RoomIndexData{Vnum: 1000, Name: "Test Room"}
	placeInRoom(ch, room)

	DoFollow(ch, "self")
	out := readOutput(ch, client)
	if !strings.Contains(out, "aren't following") {
		t.Errorf("expected 'aren't following', got: %q", out)
	}
}

func TestDoFollow_FollowSomeone(t *testing.T) {
	_ = setupGroupWorld()
	ch, chClient := makeTestChar("Alice")
	defer chClient.Close()
	victim, victimClient := makeTestChar("Bob")
	defer victimClient.Close()

	room := &types.RoomIndexData{Vnum: 1000, Name: "Test Room"}
	placeInRoom(ch, room)
	placeInRoom(victim, room)

	DoFollow(ch, "Bob")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "You now follow Bob") {
		t.Errorf("expected 'You now follow Bob', got: %q", out)
	}
	if ch.Master != victim {
		t.Error("ch.Master should be victim")
	}
	if ch.Leader != victim {
		t.Error("ch.Leader should be victim")
	}

	victimOut := readOutput(victim, victimClient)
	if !strings.Contains(victimOut, "Alice now follows you") {
		t.Errorf("expected 'Alice now follows you', got: %q", victimOut)
	}
}

func TestDoFollow_SelfToStop(t *testing.T) {
	_ = setupGroupWorld()
	ch, chClient := makeTestChar("Alice")
	defer chClient.Close()
	victim, victimClient := makeTestChar("Bob")
	defer victimClient.Close()

	room := &types.RoomIndexData{Vnum: 1000, Name: "Test Room"}
	placeInRoom(ch, room)
	placeInRoom(victim, room)

	ch.Master = victim
	ch.Leader = victim

	DoFollow(ch, "self")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "You stop following") {
		t.Errorf("expected 'You stop following', got: %q", out)
	}
	if ch.Master != nil {
		t.Error("ch.Master should be nil")
	}
	if ch.Leader != nil {
		t.Error("ch.Leader should be nil")
	}
	_ = victimClient
}

func TestDoFollow_AlreadyFollowing(t *testing.T) {
	_ = setupGroupWorld()
	ch, chClient := makeTestChar("Alice")
	defer chClient.Close()
	victim, victimClient := makeTestChar("Bob")
	defer victimClient.Close()

	room := &types.RoomIndexData{Vnum: 1000, Name: "Test Room"}
	placeInRoom(ch, room)
	placeInRoom(victim, room)

	ch.Master = victim

	DoFollow(ch, "Bob")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "already following") {
		t.Errorf("expected 'already following', got: %q", out)
	}
	_ = victimClient
}

func TestDoFollow_SwitchMaster(t *testing.T) {
	_ = setupGroupWorld()
	ch, chClient := makeTestChar("Alice")
	defer chClient.Close()
	oldMaster, oldClient := makeTestChar("Bob")
	defer oldClient.Close()
	newMaster, newClient := makeTestChar("Charlie")
	defer newClient.Close()

	room := &types.RoomIndexData{Vnum: 1000, Name: "Test Room"}
	placeInRoom(ch, room)
	placeInRoom(oldMaster, room)
	placeInRoom(newMaster, room)

	ch.Master = oldMaster
	ch.Leader = oldMaster

	DoFollow(ch, "Charlie")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "You now follow Charlie") {
		t.Errorf("expected 'You now follow Charlie', got: %q", out)
	}
	if ch.Master != newMaster {
		t.Error("ch.Master should be newMaster")
	}
	_ = oldClient
	_ = newClient
}

// --- DoGroup ---

func TestDoGroup_NoArg_ShowsSelf(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	room := &types.RoomIndexData{Vnum: 1000, Name: "Test Room"}
	placeInRoom(ch, room)

	DoGroup(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Alice") {
		t.Errorf("expected group display with 'Alice', got: %q", out)
	}
}

func TestDoGroup_TargetNotFollowing(t *testing.T) {
	_ = setupGroupWorld()
	ch, chClient := makeTestChar("Alice")
	defer chClient.Close()
	target, targetClient := makeTestChar("Bob")
	defer targetClient.Close()

	room := &types.RoomIndexData{Vnum: 1000, Name: "Test Room"}
	placeInRoom(ch, room)
	placeInRoom(target, room)

	DoGroup(ch, "Bob")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "not following you") {
		t.Errorf("expected 'not following you', got: %q", out)
	}
	_ = targetClient
}

func TestDoGroup_TargetJoins(t *testing.T) {
	_ = setupGroupWorld()
	ch, chClient := makeTestChar("Alice")
	defer chClient.Close()
	target, targetClient := makeTestChar("Bob")
	defer targetClient.Close()

	room := &types.RoomIndexData{Vnum: 1000, Name: "Test Room"}
	placeInRoom(ch, room)
	placeInRoom(target, room)

	target.Master = ch

	DoGroup(ch, "Bob")
	chOut := readOutput(ch, chClient)
	if !strings.Contains(chOut, "Bob joins your group") {
		t.Errorf("expected 'Bob joins your group', got: %q", chOut)
	}
	if target.Leader != ch {
		t.Error("target.Leader should be ch")
	}

	targetOut := readOutput(target, targetClient)
	if !strings.Contains(targetOut, "You join the group") {
		t.Errorf("expected 'You join the group', got: %q", targetOut)
	}
}

func TestDoGroup_AlreadyGrouped_Ungroup(t *testing.T) {
	_ = setupGroupWorld()
	ch, chClient := makeTestChar("Alice")
	defer chClient.Close()
	target, targetClient := makeTestChar("Bob")
	defer targetClient.Close()

	room := &types.RoomIndexData{Vnum: 1000, Name: "Test Room"}
	placeInRoom(ch, room)
	placeInRoom(target, room)

	target.Master = ch
	target.Leader = ch

	DoGroup(ch, "Bob")
	chOut := readOutput(ch, chClient)
	if !strings.Contains(chOut, "removed from your group") {
		t.Errorf("expected 'removed from your group', got: %q", chOut)
	}
	if target.Leader != nil {
		t.Error("target.Leader should be nil after ungroup")
	}

	targetOut := readOutput(target, targetClient)
	if !strings.Contains(targetOut, "removed from the group") {
		t.Errorf("expected 'removed from the group', got: %q", targetOut)
	}
}

func TestDoGroup_TargetNotFound(t *testing.T) {
	_ = setupGroupWorld()
	ch, chClient := makeTestChar("Alice")
	defer chClient.Close()

	room := &types.RoomIndexData{Vnum: 1000, Name: "Test Room"}
	placeInRoom(ch, room)

	DoGroup(ch, "nobody")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "aren't here") {
		t.Errorf("expected 'aren't here', got: %q", out)
	}
}

// --- DoOrder ---

func TestDoOrder_NoArg(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	DoOrder(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Order whom to do what?") {
		t.Errorf("expected 'Order whom to do what?', got: %q", out)
	}
}

func TestDoOrder_NoCommand(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	DoOrder(ch, "Bob")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Order them to do what?") {
		t.Errorf("expected 'Order them to do what?', got: %q", out)
	}
}

func TestDoOrder_TargetNotFollowing(t *testing.T) {
	_ = setupGroupWorld()
	ch, chClient := makeTestChar("Alice")
	defer chClient.Close()
	target, targetClient := makeTestChar("Bob")
	defer targetClient.Close()

	room := &types.RoomIndexData{Vnum: 1000, Name: "Test Room"}
	placeInRoom(ch, room)
	placeInRoom(target, room)

	DoOrder(ch, "Bob sit")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "not following you") {
		t.Errorf("expected 'not following you', got: %q", out)
	}
	_ = targetClient
}

func TestDoOrder_TargetFollowing_Queued(t *testing.T) {
	_ = setupGroupWorld()
	ch, chClient := makeTestChar("Alice")
	defer chClient.Close()
	target, targetClient := makeTestChar("Bob")
	defer targetClient.Close()

	room := &types.RoomIndexData{Vnum: 1000, Name: "Test Room"}
	placeInRoom(ch, room)
	placeInRoom(target, room)

	target.Master = ch

	DoOrder(ch, "Bob sit")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "Ok") {
		t.Errorf("expected 'Ok', got: %q", out)
	}

	// Check that command was queued
	select {
	case cmd := <-target.Desc.InputQueue:
		if cmd != "sit" {
			t.Errorf("expected 'sit' in input queue, got: %q", cmd)
		}
	default:
		t.Error("expected command in target's input queue")
	}
	_ = targetClient
}

// --- DoAssist ---

func TestDoAssist_NoArg(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	DoAssist(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Assist whom?") {
		t.Errorf("expected 'Assist whom?', got: %q", out)
	}
}

func TestDoAssist_TargetNotFighting(t *testing.T) {
	_ = setupGroupWorld()
	ch, chClient := makeTestChar("Alice")
	defer chClient.Close()
	target, targetClient := makeTestChar("Bob")
	defer targetClient.Close()

	room := &types.RoomIndexData{Vnum: 1000, Name: "Test Room"}
	placeInRoom(ch, room)
	placeInRoom(target, room)

	DoAssist(ch, "Bob")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "fighting anyone") {
		t.Errorf("expected 'fighting anyone', got: %q", out)
	}
	_ = targetClient
}

func TestDoAssist_AlreadyFighting(t *testing.T) {
	_ = setupGroupWorld()
	ch, chClient := makeTestChar("Alice")
	defer chClient.Close()
	target, targetClient := makeTestChar("Bob")
	defer targetClient.Close()
	mob := &types.CharData{Name: "mob", ShortDescr: "a mob", Level: 5}
	mob.Act.Set(types.ACT_IS_NPC)

	room := &types.RoomIndexData{Vnum: 1000, Name: "Test Room"}
	placeInRoom(ch, room)
	placeInRoom(target, room)
	placeInRoom(mob, room)

	target.Fighting = &types.FightData{Who: mob}
	ch.Fighting = &types.FightData{Who: mob}

	DoAssist(ch, "Bob")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "already fighting") {
		t.Errorf("expected 'already fighting', got: %q", out)
	}
	_ = targetClient
}

func TestDoAssist_Success(t *testing.T) {
	w := setupGroupWorld()
	ch, chClient := makeTestChar("Alice")
	defer chClient.Close()
	target, targetClient := makeTestChar("Bob")
	defer targetClient.Close()
	mob := &types.CharData{Name: "mob", ShortDescr: "a mob", Level: 5, Hit: 100, MaxHit: 100, Position: types.POS_STANDING}
	mob.Act.Set(types.ACT_IS_NPC)

	room := &types.RoomIndexData{Vnum: 1000, Name: "Test Room"}
	placeInRoom(ch, room)
	placeInRoom(target, room)
	placeInRoom(mob, room)

	// Need world chars for combat
	w.Characters = append(w.Characters, ch, target, mob)

	target.Fighting = &types.FightData{Who: mob}

	DoAssist(ch, "Bob")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "You assist Bob") {
		t.Errorf("expected 'You assist Bob', got: %q", out)
	}
	if ch.Fighting == nil || ch.Fighting.Who != mob {
		t.Error("ch should be fighting the mob")
	}
	_ = targetClient
	_ = combat.StartFighting // ensure import used
}

func TestDoAssist_TargetNotFound(t *testing.T) {
	_ = setupGroupWorld()
	ch, chClient := makeTestChar("Alice")
	defer chClient.Close()

	room := &types.RoomIndexData{Vnum: 1000, Name: "Test Room"}
	placeInRoom(ch, room)

	DoAssist(ch, "nobody")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "aren't here") {
		t.Errorf("expected 'aren't here', got: %q", out)
	}
}
