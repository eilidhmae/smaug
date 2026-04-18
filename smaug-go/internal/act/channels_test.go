package act

import (
	"net"
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/command"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

// addPlayingDescriptor registers a descriptor as CON_PLAYING in world so that
// channel broadcasts which walk WorldRef.Descriptors can see the char.
func addPlayingDescriptor(ch *types.CharData) {
	if ch == nil || ch.Desc == nil {
		return
	}
	ch.Desc.Connected = types.CON_PLAYING
	ch.Desc.Character = ch
	WorldRef.Descriptors = append(WorldRef.Descriptors, ch.Desc)
}

// makeImmortalInRoom creates an immortal character, puts them in the given
// room, adds them to the world, and registers their descriptor as PLAYING.
// Returns (ch, client-side-of-pipe) — the caller uses readOutput(ch, client).
func makeImmortalInRoom(room *types.RoomIndexData, name string) (*types.CharData, net.Conn) {
	ch, client := makeTestChar(name)
	ch.Level = types.LEVEL_IMMORTAL
	handler.CharToRoom(ch, room)
	WorldRef.AddChar(ch)
	addPlayingDescriptor(ch)
	return ch, client
}

// makeMortalInRoom is the mortal counterpart of makeImmortalInRoom.
func makeMortalInRoom(room *types.RoomIndexData, name string) (*types.CharData, net.Conn) {
	ch, client := makeTestChar(name)
	handler.CharToRoom(ch, room)
	WorldRef.AddChar(ch)
	addPlayingDescriptor(ch)
	return ch, client
}

// --- DoImmtalk (plan-channels.md G2) ---

// Empty arg path: "Immtalk what?"
func TestDoImmtalk_NoArg(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 7000, Name: "Room"}
	w.Rooms[7000] = room

	ch, client := makeImmortalInRoom(room, "Zeus")
	defer client.Close()

	DoImmtalk(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(strings.ToLower(out), "immtalk what") {
		t.Errorf("expected 'Immtalk what?' prompt; got %q", out)
	}
}

// Mortal: sees "Huh?"; no other descriptor receives output.
func TestDoImmtalk_MortalDenied(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 7001, Name: "Room"}
	w.Rooms[7001] = room

	mortal, mClient := makeMortalInRoom(room, "Mort")
	defer mClient.Close()
	imm, immClient := makeImmortalInRoom(room, "Zeus")
	defer immClient.Close()

	DoImmtalk(mortal, "hi")

	mOut := readOutput(mortal, mClient)
	if !strings.Contains(mOut, "Huh?") {
		t.Errorf("mortal should see 'Huh?'; got %q", mOut)
	}
	immOut := readOutput(imm, immClient)
	if strings.Contains(immOut, "hi") {
		t.Errorf("immortal must NOT receive message from mortal's immtalk; got %q", immOut)
	}
}

// Two immortals A and B online: A's immtalk reaches B.
func TestDoImmtalk_ImmortalToImmortal(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 7002, Name: "Room"}
	w.Rooms[7002] = room

	a, aClient := makeImmortalInRoom(room, "Zeus")
	defer aClient.Close()
	b, bClient := makeImmortalInRoom(room, "Odin")
	defer bClient.Close()

	DoImmtalk(a, "hi there")

	aOut := readOutput(a, aClient)
	if !strings.Contains(aOut, "hi there") {
		t.Errorf("sender should see own echo; got %q", aOut)
	}

	bOut := readOutput(b, bClient)
	if !strings.Contains(bOut, "hi there") {
		t.Errorf("other immortal must receive the immtalk; got %q", bOut)
	}
	if !strings.Contains(bOut, "Zeus") {
		t.Errorf("receiver should see sender's name 'Zeus'; got %q", bOut)
	}
}

// Mortal online alongside an immortal: mortal sees nothing.
func TestDoImmtalk_MortalDoesNotReceive(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 7003, Name: "Room"}
	w.Rooms[7003] = room

	imm, immClient := makeImmortalInRoom(room, "Zeus")
	defer immClient.Close()
	mortal, mClient := makeMortalInRoom(room, "Mort")
	defer mClient.Close()

	DoImmtalk(imm, "secret council")

	mOut := readOutput(mortal, mClient)
	if strings.Contains(mOut, "secret council") {
		t.Errorf("mortal must NOT receive immtalk; got %q", mOut)
	}
}

// PLR_SILENCE on sender blocks.
func TestDoImmtalk_PLRSilence_BlocksSender(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 7004, Name: "Room"}
	w.Rooms[7004] = room

	imm, immClient := makeImmortalInRoom(room, "Zeus")
	defer immClient.Close()
	imm.Act.Set(types.PLR_SILENCE)

	b, bClient := makeImmortalInRoom(room, "Odin")
	defer bClient.Close()

	DoImmtalk(imm, "shhh")

	immOut := readOutput(imm, immClient)
	if !strings.Contains(immOut, "You can't immtalk") {
		t.Errorf("silenced immortal should see \"You can't immtalk\"; got %q", immOut)
	}
	bOut := readOutput(b, bClient)
	if strings.Contains(bOut, "shhh") {
		t.Errorf("receiver must NOT get silenced immtalk; got %q", bOut)
	}
}

// Deaf sender blocked with "You don't have the immtalk channel turned on..."
// AND the deaf flag is NOT cleared (C act_comm.c:505-513 — xREMOVE_BIT at
// line 514 is unreachable when the sender returns on line 511).
func TestDoImmtalk_DeafSender_Blocked_NotCleared(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 7005, Name: "Room"}
	w.Rooms[7005] = room

	imm, immClient := makeImmortalInRoom(room, "Zeus")
	defer immClient.Close()
	imm.Deaf.Set(types.CHANNEL_IMMTALK)

	b, bClient := makeImmortalInRoom(room, "Odin")
	defer bClient.Close()

	DoImmtalk(imm, "testing")

	immOut := readOutput(imm, immClient)
	if !strings.Contains(strings.ToLower(immOut), "don't have the immtalk channel turned on") {
		t.Errorf("deaf sender should see the 'channel turned off' diagnostic; got %q", immOut)
	}
	bOut := readOutput(b, bClient)
	if strings.Contains(bOut, "testing") {
		t.Errorf("other immortal must NOT receive a broadcast; got %q", bOut)
	}
	if !imm.Deaf.IsSet(types.CHANNEL_IMMTALK) {
		t.Error("deaf flag must NOT be cleared by a deaf sender (C does not auto-unmute here)")
	}
}

// --- DoGtell (plan-channels.md G3) ---

// Empty-arg: "Tell your group what?"
func TestDoGtell_NoArg(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 7100, Name: "Room"}
	w.Rooms[7100] = room

	ch, client := makeMortalInRoom(room, "Solo")
	defer client.Close()

	DoGtell(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(strings.ToLower(out), "tell your group what") {
		t.Errorf("expected 'Tell your group what?'; got %q", out)
	}
}

// Solo player: no Leader, no followers — only self gets the line.
func TestDoGtell_Solo(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 7101, Name: "Room"}
	w.Rooms[7101] = room

	ch, client := makeMortalInRoom(room, "Solo")
	defer client.Close()

	// Another char in the same world, NOT in ch's group.
	other, otherClient := makeMortalInRoom(room, "Other")
	defer otherClient.Close()

	DoGtell(ch, "anyone?")

	chOut := readOutput(ch, client)
	if !strings.Contains(chOut, "anyone?") {
		t.Errorf("solo sender should see own gtell echo (self is in own group); got %q", chOut)
	}

	otherOut := readOutput(other, otherClient)
	if strings.Contains(otherOut, "anyone?") {
		t.Errorf("unrelated char must NOT receive gtell; got %q", otherOut)
	}
}

// A leads {A, B}; C unrelated. A's gtell reaches B, not C.
func TestDoGtell_DeliveresToGroupOnly(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 7102, Name: "Room"}
	w.Rooms[7102] = room

	a, aClient := makeMortalInRoom(room, "Alpha")
	defer aClient.Close()
	b, bClient := makeMortalInRoom(room, "Beta")
	defer bClient.Close()
	c, cClient := makeMortalInRoom(room, "Charlie")
	defer cClient.Close()

	b.Leader = a // B follows A; A is its own leader

	DoGtell(a, "plan attack")

	aOut := readOutput(a, aClient)
	if !strings.Contains(aOut, "plan attack") {
		t.Errorf("sender A should see the gtell; got %q", aOut)
	}
	bOut := readOutput(b, bClient)
	if !strings.Contains(bOut, "plan attack") {
		t.Errorf("group-member B should receive; got %q", bOut)
	}
	if !strings.Contains(bOut, "Alpha") {
		t.Errorf("B should see sender name Alpha; got %q", bOut)
	}
	cOut := readOutput(c, cClient)
	if strings.Contains(cOut, "plan attack") {
		t.Errorf("unrelated C must NOT receive; got %q", cOut)
	}
}

// PLR_NO_TELL on sender blocks (C act_comm.c:4237).
func TestDoGtell_PLRNoTell_BlocksSender(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 7103, Name: "Room"}
	w.Rooms[7103] = room

	a, aClient := makeMortalInRoom(room, "Alpha")
	defer aClient.Close()
	b, bClient := makeMortalInRoom(room, "Beta")
	defer bClient.Close()
	b.Leader = a
	a.Act.Set(types.PLR_NO_TELL)

	DoGtell(a, "hi")

	aOut := readOutput(a, aClient)
	if !strings.Contains(aOut, "Your message didn't get through") {
		t.Errorf("sender with PLR_NO_TELL should see bounce message; got %q", aOut)
	}
	bOut := readOutput(b, bClient)
	if strings.Contains(bOut, "hi") {
		t.Errorf("follower must NOT receive when sender is NO_TELL; got %q", bOut)
	}
}

// Sleeper/disconnected char with Desc == nil must not panic.
func TestDoGtell_NilDescNoPanic(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 7104, Name: "Room"}
	w.Rooms[7104] = room

	a, aClient := makeMortalInRoom(room, "Alpha")
	defer aClient.Close()

	// Beta is "asleep" — Desc nil — and still in WorldRef.Characters.
	beta := &types.CharData{Name: "Beta", Leader: a}
	handler.CharToRoom(beta, room)
	w.AddChar(beta)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("DoGtell panicked on nil-Desc follower: %v", r)
		}
	}()
	DoGtell(a, "wake up")

	aOut := readOutput(a, aClient)
	if !strings.Contains(aOut, "wake up") {
		t.Errorf("sender should still get own echo; got %q", aOut)
	}
}

// NPC leader with follower: follower still hears. SMAUG permits this.
func TestDoGtell_NPCLeaderFollowerHears(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 7105, Name: "Room"}
	w.Rooms[7105] = room

	mob, _ := makeTestChar("Boss")
	mob.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(mob, room)
	w.AddChar(mob)

	follower, fClient := makeMortalInRoom(room, "Minion")
	defer fClient.Close()
	follower.Leader = mob

	DoGtell(mob, "attack on sight")

	fOut := readOutput(follower, fClient)
	if !strings.Contains(fOut, "attack on sight") {
		t.Errorf("follower must hear NPC leader's gtell; got %q", fOut)
	}
}

// Receivers with Deaf[IMMTALK] do not receive (sender is not deaf).
func TestDoImmtalk_DeafReceiverFiltered(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 7006, Name: "Room"}
	w.Rooms[7006] = room

	a, aClient := makeImmortalInRoom(room, "Zeus")
	defer aClient.Close()
	b, bClient := makeImmortalInRoom(room, "Odin")
	defer bClient.Close()
	b.Deaf.Set(types.CHANNEL_IMMTALK)

	DoImmtalk(a, "loud")

	bOut := readOutput(b, bClient)
	if strings.Contains(bOut, "loud") {
		t.Errorf("deaf receiver must NOT receive immtalk; got %q", bOut)
	}
}

// --- Alias dispatch (plan-channels.md acceptance criteria 4 and 11) ---
//
// These tests exercise the generic Register/Interpret path rather than the
// production boot registry, matching the pattern used by
// TestInterpret_PagelenAlias. `": hi"` parses via util.OneArgument into
// cmdWord=":" rest="hi"; `":hi"` (no space) would parse to cmdWord=":hi"
// and not match — the plan documents this and calls it out of scope.

func TestInterpret_ImmtalkAlias_Colon(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 7300, Name: "Room"}
	w.Rooms[7300] = room

	reg := command.NewRegistry()
	reg.Register(&command.Command{Name: "immtalk", DoFun: DoImmtalk, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: ":", DoFun: DoImmtalk, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})

	sender, sClient := makeImmortalInRoom(room, "Zeus")
	defer sClient.Close()
	receiver, rClient := makeImmortalInRoom(room, "Odin")
	defer rClient.Close()

	reg.Interpret(sender, ": via alias")

	sOut := readOutput(sender, sClient)
	if !strings.Contains(sOut, "via alias") {
		t.Errorf("': via alias' should produce the immtalk self-echo; got %q", sOut)
	}
	rOut := readOutput(receiver, rClient)
	if !strings.Contains(rOut, "via alias") {
		t.Errorf("': via alias' should be broadcast through the alias; got %q", rOut)
	}
}

func TestInterpret_GtellAlias_Semicolon(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 7301, Name: "Room"}
	w.Rooms[7301] = room

	reg := command.NewRegistry()
	reg.Register(&command.Command{Name: "gtell", DoFun: DoGtell, Position: types.POS_SLEEPING, Level: 0})
	reg.Register(&command.Command{Name: ";", DoFun: DoGtell, Position: types.POS_SLEEPING, Level: 0})

	a, aClient := makeMortalInRoom(room, "Alpha")
	defer aClient.Close()
	b, bClient := makeMortalInRoom(room, "Beta")
	defer bClient.Close()
	b.Leader = a

	reg.Interpret(a, "; via alias")

	aOut := readOutput(a, aClient)
	if !strings.Contains(aOut, "via alias") {
		t.Errorf("'; via alias' should echo to sender; got %q", aOut)
	}
	bOut := readOutput(b, bClient)
	if !strings.Contains(bOut, "via alias") {
		t.Errorf("'; via alias' should reach group member; got %q", bOut)
	}
}
