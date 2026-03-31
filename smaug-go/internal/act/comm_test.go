package act

import (
	"net"
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func setupCommWorld() *world.World {
	w := world.New("/tmp/test")
	WorldRef = w
	return w
}

func TestDoTell(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 6100, Name: "Test Room"}
	w.Rooms[6100] = room

	ch, chClient := makeTestChar("Sender")
	defer chClient.Close()
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	victim, victimClient := makeTestChar("Receiver")
	defer victimClient.Close()
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	DoTell(ch, "Receiver hello there")

	output := readOutput(ch, chClient)
	if !strings.Contains(output, "Receiver") && !strings.Contains(output, "hello there") {
		t.Errorf("sender should see tell confirmation, got: %q", output)
	}

	victimOut := readOutput(victim, victimClient)
	if !strings.Contains(victimOut, "hello there") {
		t.Errorf("receiver should see the message, got: %q", victimOut)
	}

	// Reply should be set
	if victim.Reply != ch {
		t.Error("victim.Reply should be set to sender")
	}
}

func TestDoTell_NoArg(t *testing.T) {
	_ = setupCommWorld()
	ch, client := makeTestChar("Sender")
	defer client.Close()

	DoTell(ch, "")
	output := readOutput(ch, client)
	if !strings.Contains(strings.ToLower(output), "tell whom") {
		t.Errorf("expected 'tell whom' message, got: %q", output)
	}
}

func TestDoReply(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 6101, Name: "Test Room"}
	w.Rooms[6101] = room

	ch, chClient := makeTestChar("Alice")
	defer chClient.Close()
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	bob, bobClient := makeTestChar("Bob")
	defer bobClient.Close()
	handler.CharToRoom(bob, room)
	w.AddChar(bob)

	// Bob tells Alice first
	DoTell(bob, "Alice hey there")
	_ = readOutput(bob, bobClient)
	_ = readOutput(ch, chClient)

	// Alice replies
	DoReply(ch, "hey back!")

	bobOut := readOutput(bob, bobClient)
	if !strings.Contains(bobOut, "hey back!") {
		t.Errorf("Bob should receive reply, got: %q", bobOut)
	}
}

func TestDoGossip(t *testing.T) {
	w := setupCommWorld()
	room1 := &types.RoomIndexData{Vnum: 6102, Name: "Room 1"}
	room2 := &types.RoomIndexData{Vnum: 6103, Name: "Room 2"}
	w.Rooms[6102] = room1
	w.Rooms[6103] = room2

	ch, chClient := makeTestChar("Gossiper")
	defer chClient.Close()
	handler.CharToRoom(ch, room1)
	w.AddChar(ch)

	listener, listenerClient := makeTestChar("Listener")
	defer listenerClient.Close()
	handler.CharToRoom(listener, room2) // Different room
	w.AddChar(listener)

	DoGossip(ch, "hello world")

	listenerOut := readOutput(listener, listenerClient)
	if !strings.Contains(listenerOut, "hello world") {
		t.Errorf("listener in different room should hear gossip, got: %q", listenerOut)
	}
}

func TestDoEmote(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 6104, Name: "Test Room"}
	w.Rooms[6104] = room

	ch, chClient := makeTestChar("Actor")
	defer chClient.Close()
	handler.CharToRoom(ch, room)

	bystander, bystanderClient := makeTestChar("Watcher")
	defer bystanderClient.Close()
	handler.CharToRoom(bystander, room)

	DoEmote(ch, "dances around.")

	bystanderOut := readOutput(bystander, bystanderClient)
	if !strings.Contains(bystanderOut, "Actor dances around.") {
		t.Errorf("bystander should see emote, got: %q", bystanderOut)
	}
}

func TestDoYell(t *testing.T) {
	w := setupCommWorld()
	area := &types.AreaData{Name: "Test Area"}
	room1 := &types.RoomIndexData{Vnum: 6105, Name: "Room 1", Area: area}
	room2 := &types.RoomIndexData{Vnum: 6106, Name: "Room 2", Area: area}
	w.Rooms[6105] = room1
	w.Rooms[6106] = room2

	ch, chClient := makeTestChar("Yeller")
	defer chClient.Close()
	handler.CharToRoom(ch, room1)
	w.AddChar(ch)

	listener, listenerClient := makeTestChar("Listener")
	defer listenerClient.Close()
	handler.CharToRoom(listener, room2) // Same area, different room
	w.AddChar(listener)

	DoYell(ch, "help me!")

	listenerOut := readOutput(listener, listenerClient)
	if !strings.Contains(listenerOut, "help me!") {
		t.Errorf("listener in same area should hear yell, got: %q", listenerOut)
	}
}

// Suppress unused import
func init() {
	_ = net.Pipe
}
