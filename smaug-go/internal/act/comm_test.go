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

// --- PLR_SILENCE opportunistic sender-gate tests (plan-channels.md G1) ---
//
// C `talk_channel` (act_comm.c:500-504) blocks PLR_SILENCE'd senders from
// yell/shout/gossip with "You can't <verb>." C `do_tell` (act_comm.c:1804)
// blocks with "You can't do that." The Go port previously defined
// PLR_SILENCE (set/cleared by DoSilence) but never consulted it in these
// commands. These tests pin the newly-added sender gate.

func TestDoShout_PLRSilence_BlocksSender(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 6201, Name: "Room"}
	w.Rooms[6201] = room

	ch, chClient := makeTestChar("Silenced")
	defer chClient.Close()
	handler.CharToRoom(ch, room)
	w.AddChar(ch)
	ch.Act.Set(types.PLR_SILENCE)

	listener, listenerClient := makeTestChar("Listener")
	defer listenerClient.Close()
	handler.CharToRoom(listener, room)
	w.AddChar(listener)

	DoShout(ch, "hello")

	chOut := readOutput(ch, chClient)
	if !strings.Contains(chOut, "You can't shout") {
		t.Errorf("silenced sender should see \"You can't shout\", got: %q", chOut)
	}
	if strings.Contains(chOut, "You shout") {
		t.Errorf("silenced sender must not see self-echo; got: %q", chOut)
	}
	listenerOut := readOutput(listener, listenerClient)
	if strings.Contains(listenerOut, "hello") {
		t.Errorf("listener must NOT receive the silenced shout; got: %q", listenerOut)
	}
}

func TestDoGossip_PLRSilence_BlocksSender(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 6202, Name: "Room"}
	w.Rooms[6202] = room

	ch, chClient := makeTestChar("Silenced")
	defer chClient.Close()
	handler.CharToRoom(ch, room)
	w.AddChar(ch)
	ch.Act.Set(types.PLR_SILENCE)

	listener, listenerClient := makeTestChar("Listener")
	defer listenerClient.Close()
	handler.CharToRoom(listener, room)
	w.AddChar(listener)

	DoGossip(ch, "hello")

	chOut := readOutput(ch, chClient)
	if !strings.Contains(chOut, "You can't gossip") {
		t.Errorf("silenced sender should see \"You can't gossip\", got: %q", chOut)
	}
	listenerOut := readOutput(listener, listenerClient)
	if strings.Contains(listenerOut, "hello") {
		t.Errorf("listener must NOT receive the silenced gossip; got: %q", listenerOut)
	}
}

func TestDoYell_PLRSilence_BlocksSender(t *testing.T) {
	w := setupCommWorld()
	area := &types.AreaData{Name: "Test Area"}
	room := &types.RoomIndexData{Vnum: 6203, Name: "Room", Area: area}
	w.Rooms[6203] = room

	ch, chClient := makeTestChar("Silenced")
	defer chClient.Close()
	handler.CharToRoom(ch, room)
	w.AddChar(ch)
	ch.Act.Set(types.PLR_SILENCE)

	listener, listenerClient := makeTestChar("Listener")
	defer listenerClient.Close()
	handler.CharToRoom(listener, room)
	w.AddChar(listener)

	DoYell(ch, "help me")

	chOut := readOutput(ch, chClient)
	if !strings.Contains(chOut, "You can't yell") {
		t.Errorf("silenced sender should see \"You can't yell\", got: %q", chOut)
	}
	listenerOut := readOutput(listener, listenerClient)
	if strings.Contains(listenerOut, "help me") {
		t.Errorf("listener must NOT receive the silenced yell; got: %q", listenerOut)
	}
}

func TestDoTell_PLRSilence_BlocksSender(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 6204, Name: "Room"}
	w.Rooms[6204] = room

	ch, chClient := makeTestChar("Silenced")
	defer chClient.Close()
	handler.CharToRoom(ch, room)
	w.AddChar(ch)
	ch.Act.Set(types.PLR_SILENCE)

	victim, victimClient := makeTestChar("Target")
	defer victimClient.Close()
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	DoTell(ch, "Target secret plan")

	chOut := readOutput(ch, chClient)
	if !strings.Contains(chOut, "You can't do that") {
		t.Errorf("silenced sender should see \"You can't do that\", got: %q", chOut)
	}
	victimOut := readOutput(victim, victimClient)
	if strings.Contains(victimOut, "secret plan") {
		t.Errorf("target must NOT receive the silenced tell; got: %q", victimOut)
	}
}

// Sanity-check the complement: an NPC "silenced" flag on an NPC (same bit
// position is ACT_IS_NPC-inclusive; PLR_* flags only meaningful for PCs)
// must not accidentally fire the gate. IsNPC() short-circuits.
func TestDoShout_NPCNotGatedBySilence(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 6205, Name: "Room"}
	w.Rooms[6205] = room

	ch, chClient := makeTestChar("Mob")
	defer chClient.Close()
	ch.Act.Set(types.ACT_IS_NPC) // NPC; PLR_SILENCE should not apply
	ch.Act.Set(types.PLR_SILENCE)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	DoShout(ch, "roar")
	chOut := readOutput(ch, chClient)
	if strings.Contains(chOut, "You can't shout") {
		t.Errorf("NPC flagged with PLR_SILENCE-bit must not trigger the sender gate; got: %q", chOut)
	}
	if !strings.Contains(chOut, "You shout 'roar'") {
		t.Errorf("NPC shout should self-echo; got: %q", chOut)
	}
}

// Suppress unused import
func init() {
	_ = net.Pipe
}
