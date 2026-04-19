package act

import (
	"bytes"
	"net"
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/command"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/persist"
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

// --- DoChannels (plan-do-channels.md G2) ---

// makeChannelsPC is a minimal PC with a descriptor — no room, no world
// registration needed because DoChannels does not broadcast.
func makeChannelsPC(name string) (*types.CharData, net.Conn) {
	ch, client := makeTestChar(name)
	return ch, client
}

func TestDoChannels_NPCIsNoop(t *testing.T) {
	_ = setupCommWorld()
	ch, client := makeChannelsPC("Mob")
	defer client.Close()
	ch.Act.Set(types.ACT_IS_NPC)

	DoChannels(ch, "")
	out := readOutput(ch, client)
	if out != "" {
		t.Errorf("NPC DoChannels must be a silent no-op; got %q", out)
	}
}

func TestDoChannels_SilencedPrintsMessage(t *testing.T) {
	_ = setupCommWorld()
	ch, client := makeChannelsPC("Gagged")
	defer client.Close()
	ch.Act.Set(types.PLR_SILENCE)

	DoChannels(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "You are silenced.") {
		t.Errorf("silenced PC should see 'You are silenced.'; got %q", out)
	}
}

func TestDoChannels_NoArgDisplaysGroups(t *testing.T) {
	_ = setupCommWorld()
	ch, client := makeChannelsPC("Norm")
	defer client.Close()

	DoChannels(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Public channels") {
		t.Errorf("no-arg display must include 'Public channels' header; got %q", out)
	}
	if !strings.Contains(out, "Private channels") {
		t.Errorf("no-arg display must include 'Private channels' header; got %q", out)
	}
}

func TestDoChannels_NoArgShowsEnabledWithPlus(t *testing.T) {
	_ = setupCommWorld()
	ch, client := makeChannelsPC("Norm")
	defer client.Close()

	DoChannels(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "+CHAT") {
		t.Errorf("enabled CHAT channel should render as '+CHAT'; got %q", out)
	}
}

func TestDoChannels_NoArgShowsDisabledWithMinus(t *testing.T) {
	_ = setupCommWorld()
	ch, client := makeChannelsPC("Norm")
	defer client.Close()
	ch.Deaf.Set(types.CHANNEL_CHAT)

	DoChannels(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "-chat") {
		t.Errorf("disabled CHAT channel should render as '-chat'; got %q", out)
	}
	if strings.Contains(out, "+CHAT") {
		t.Errorf("CHAT is disabled so '+CHAT' must NOT appear; got %q", out)
	}
}

func TestDoChannels_ToggleOffSetsBit(t *testing.T) {
	_ = setupCommWorld()
	ch, client := makeChannelsPC("Norm")
	defer client.Close()

	DoChannels(ch, "-chat")
	if !ch.Deaf.IsSet(types.CHANNEL_CHAT) {
		t.Error("'-chat' must SET the CHANNEL_CHAT deaf bit")
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "Ok.") {
		t.Errorf("'-chat' should confirm with 'Ok.'; got %q", out)
	}
}

func TestDoChannels_ToggleOnClearsBit(t *testing.T) {
	_ = setupCommWorld()
	ch, client := makeChannelsPC("Norm")
	defer client.Close()
	ch.Deaf.Set(types.CHANNEL_CHAT)

	DoChannels(ch, "+chat")
	if ch.Deaf.IsSet(types.CHANNEL_CHAT) {
		t.Error("'+chat' must CLEAR the CHANNEL_CHAT deaf bit")
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "Ok.") {
		t.Errorf("'+chat' should confirm with 'Ok.'; got %q", out)
	}
}

func TestDoChannels_InvalidFormat(t *testing.T) {
	_ = setupCommWorld()
	ch, client := makeChannelsPC("Norm")
	defer client.Close()

	DoChannels(ch, "chat")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Channels -channel or +channel?") {
		t.Errorf("missing sign should produce format error; got %q", out)
	}
}

func TestDoChannels_UnknownChannel(t *testing.T) {
	_ = setupCommWorld()
	ch, client := makeChannelsPC("Norm")
	defer client.Close()

	DoChannels(ch, "-foobar")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Set or clear which channel?") {
		t.Errorf("unknown channel should produce 'Set or clear...'; got %q", out)
	}
}

// `gossip` is not a valid keyword — C uses `chat`. Guards against a future
// refactor that accidentally adds `gossip` as an alias.
func TestDoChannels_GossipKeywordRejected(t *testing.T) {
	_ = setupCommWorld()
	ch, client := makeChannelsPC("Norm")
	defer client.Close()

	DoChannels(ch, "-gossip")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Set or clear which channel?") {
		t.Errorf("`gossip` is not a channel keyword — should be rejected; got %q", out)
	}
	if ch.Deaf.IsSet(types.CHANNEL_CHAT) {
		t.Error("'-gossip' must NOT toggle CHANNEL_CHAT")
	}
}

// C oddity preserved: `muse` toggles CHANNEL_HIGHGOD, not CHANNEL_MUSIC.
func TestDoChannels_MuseTogglesHighgodNotMusic(t *testing.T) {
	_ = setupCommWorld()
	ch, client := makeChannelsPC("Norm")
	defer client.Close()

	DoChannels(ch, "-muse")
	if !ch.Deaf.IsSet(types.CHANNEL_HIGHGOD) {
		t.Error("'-muse' must SET CHANNEL_HIGHGOD")
	}
	if ch.Deaf.IsSet(types.CHANNEL_MUSIC) {
		t.Error("'-muse' must NOT touch CHANNEL_MUSIC")
	}
}

// `pray` is commented out in C's no-arg display (act_info.c:5358).
func TestDoChannels_PrayNotInNoArgDisplay(t *testing.T) {
	_ = setupCommWorld()
	ch, client := makeChannelsPC("Norm")
	defer client.Close()

	DoChannels(ch, "")
	out := readOutput(ch, client)
	if strings.Contains(strings.ToLower(out), "pray") {
		t.Errorf("pray must not appear in the no-arg display; got %q", out)
	}
}

// But `-pray` / `+pray` still toggle individually via the table.
func TestDoChannels_PrayIndividualToggleWorks(t *testing.T) {
	_ = setupCommWorld()
	ch, client := makeChannelsPC("Norm")
	defer client.Close()

	DoChannels(ch, "-pray")
	if !ch.Deaf.IsSet(types.CHANNEL_PRAY) {
		t.Error("'-pray' must SET CHANNEL_PRAY")
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "Ok.") {
		t.Errorf("'-pray' should confirm with 'Ok.'; got %q", out)
	}
}

// Mortal `+all` does NOT clear AVTALK; immortal `+all` does.
// Mortal is at LEVEL_HERO so a `>= LEVEL_HERO` mutation in the gate is
// caught — only `>= LEVEL_IMMORTAL` preserves the adversary's catch.
func TestDoChannels_AllAvtalkGatedOnLevelImmortal(t *testing.T) {
	_ = setupCommWorld()

	// Hero-level mortal (still mortal — LEVEL_HERO < LEVEL_IMMORTAL).
	mortal, mClient := makeChannelsPC("Mort")
	defer mClient.Close()
	mortal.Level = types.LEVEL_HERO
	mortal.Deaf.Set(types.CHANNEL_AVTALK)

	DoChannels(mortal, "+all")
	if !mortal.Deaf.IsSet(types.CHANNEL_AVTALK) {
		t.Error("mortal '+all' must NOT clear CHANNEL_AVTALK (level < LEVEL_IMMORTAL)")
	}

	// Also verify on a low-level mortal.
	lowly, lClient := makeChannelsPC("Low")
	defer lClient.Close()
	lowly.Level = 10
	lowly.Deaf.Set(types.CHANNEL_AVTALK)

	DoChannels(lowly, "+all")
	if !lowly.Deaf.IsSet(types.CHANNEL_AVTALK) {
		t.Error("low-level mortal '+all' must NOT clear CHANNEL_AVTALK")
	}

	// Immortal path.
	imm, iClient := makeChannelsPC("Imm")
	defer iClient.Close()
	imm.Level = types.LEVEL_IMMORTAL
	imm.Deaf.Set(types.CHANNEL_AVTALK)

	DoChannels(imm, "+all")
	if imm.Deaf.IsSet(types.CHANNEL_AVTALK) {
		t.Error("immortal '+all' must clear CHANNEL_AVTALK")
	}
}

// Derivation test: manually append a fake entry with publicAll=true to
// the table; the `+all` handler must include it without any change to
// DoChannels' logic. Guards against regressing to an ad-hoc duplicate
// slice.
func TestDoChannels_PublicAllDerivedFromToggleTable(t *testing.T) {
	_ = setupCommWorld()
	// Pick an otherwise-unused channel bit the existing table does NOT
	// already flag publicAll=true: CHANNEL_TELLS (publicAll=false).
	origTable := channelToggleTable
	t.Cleanup(func() { channelToggleTable = origTable })
	channelToggleTable = append([]struct {
		name      string
		bit       int
		publicAll bool
	}{}, origTable...)
	// Flip TELLS to publicAll=true for this one test.
	for i := range channelToggleTable {
		if channelToggleTable[i].name == "tells" {
			channelToggleTable[i].publicAll = true
			break
		}
	}

	ch, client := makeChannelsPC("Deriv")
	defer client.Close()
	ch.Level = 10

	// Fresh character — TELLS bit clear.
	if ch.Deaf.IsSet(types.CHANNEL_TELLS) {
		t.Fatal("precondition: CHANNEL_TELLS should start clear")
	}
	DoChannels(ch, "-all")
	if !ch.Deaf.IsSet(types.CHANNEL_TELLS) {
		t.Error("'-all' must set CHANNEL_TELLS once its table entry is publicAll=true")
	}

	DoChannels(ch, "+all")
	if ch.Deaf.IsSet(types.CHANNEL_TELLS) {
		t.Error("'+all' must clear CHANNEL_TELLS once its table entry is publicAll=true")
	}
}

// Negative direction: a channel whose table entry has publicAll=false
// MUST NOT be touched by `+all`/`-all`. Guards the other mutation
// direction (accidentally iterating every entry in the table).
func TestDoChannels_PrivateChannelsNotToggledByAll(t *testing.T) {
	_ = setupCommWorld()
	ch, client := makeChannelsPC("Priv")
	defer client.Close()
	ch.Level = 10

	// WHISPER has publicAll=false in the table.
	if ch.Deaf.IsSet(types.CHANNEL_WHISPER) {
		t.Fatal("precondition: CHANNEL_WHISPER should start clear")
	}
	DoChannels(ch, "-all")
	if ch.Deaf.IsSet(types.CHANNEL_WHISPER) {
		t.Error("'-all' must NOT touch CHANNEL_WHISPER (publicAll=false)")
	}
}

func TestDoChannels_AllTogglePublicSet(t *testing.T) {
	_ = setupCommWorld()
	ch, client := makeChannelsPC("Norm")
	defer client.Close()
	ch.Level = 10

	DoChannels(ch, "-all")

	publicSet := []int{
		types.CHANNEL_RACETALK, types.CHANNEL_AUCTION, types.CHANNEL_CHAT,
		types.CHANNEL_QUEST, types.CHANNEL_WARTALK, types.CHANNEL_PRAY,
		types.CHANNEL_TRAFFIC, types.CHANNEL_MUSIC, types.CHANNEL_ASK,
		types.CHANNEL_YELL,
	}
	for _, bit := range publicSet {
		if !ch.Deaf.IsSet(bit) {
			t.Errorf("'-all' must SET public-set bit %d", bit)
		}
	}
	if ch.Deaf.IsSet(types.CHANNEL_TELLS) {
		t.Error("'-all' must NOT touch CHANNEL_TELLS (private)")
	}
	if ch.Deaf.IsSet(types.CHANNEL_AVTALK) {
		t.Error("mortal '-all' must NOT touch CHANNEL_AVTALK")
	}
}

// End-to-end proof that G1 persistence + G2 command round-trip a Deaf bit.
func TestDoChannels_RoundTripPersistsViaSave(t *testing.T) {
	_ = setupCommWorld()
	ch, client := makeChannelsPC("Rover")
	defer client.Close()

	DoChannels(ch, "-chat")
	if !ch.Deaf.IsSet(types.CHANNEL_CHAT) {
		t.Fatal("precondition: '-chat' should set CHANNEL_CHAT deaf bit")
	}

	var buf bytes.Buffer
	if err := persist.SavePlayer(&buf, ch); err != nil {
		t.Fatalf("SavePlayer failed: %v", err)
	}

	loaded, err := persist.LoadPlayer(&buf, "rover")
	if err != nil {
		t.Fatalf("LoadPlayer failed: %v", err)
	}
	if !loaded.Deaf.IsSet(types.CHANNEL_CHAT) {
		t.Error("CHANNEL_CHAT deaf bit must survive Save/Load round trip")
	}
}

// --- Phase 6 Extra Channels (plan-phase6-channels-extra.md) ---
//
// G0 talkChannel helper + G1-G5 wrappers.

// makeMortalPCInRoom is like makeMortalInRoom but also ensures PCData is
// present (makeTestChar already creates PCData). Keeps intent explicit for
// the extra-channel tests that exercise PCData-dependent predicates.
func makeMortalPCInRoom(room *types.RoomIndexData, name string) (*types.CharData, net.Conn) {
	ch, client := makeMortalInRoom(room, name)
	if ch.PCData == nil {
		ch.PCData = &types.PCData{}
	}
	return ch, client
}

// --- G0: talkChannel helper ---

func TestTalkChannel_EmptyArgPrintsVerbWhat(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8000, Name: "Room"}
	w.Rooms[8000] = room
	ch, client := makeMortalInRoom(room, "Piper")
	defer client.Close()

	talkChannel(ch, "", types.CHANNEL_MUSIC, "music", nil)
	out := readOutput(ch, client)
	if !strings.Contains(out, "Music what?") {
		t.Errorf("empty arg should yield 'Music what?'; got %q", out)
	}
}

func TestTalkChannel_PLRSilenceBlocksSender(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8001, Name: "Room"}
	w.Rooms[8001] = room
	sender, sClient := makeMortalInRoom(room, "Piper")
	defer sClient.Close()
	sender.Act.Set(types.PLR_SILENCE)
	listener, lClient := makeMortalInRoom(room, "Ear")
	defer lClient.Close()

	talkChannel(sender, "la la", types.CHANNEL_MUSIC, "music", nil)

	sOut := readOutput(sender, sClient)
	if !strings.Contains(sOut, "You can't music") {
		t.Errorf("silenced sender should see \"You can't music\"; got %q", sOut)
	}
	lOut := readOutput(listener, lClient)
	if strings.Contains(lOut, "la la") {
		t.Errorf("silenced sender must not broadcast; listener got %q", lOut)
	}
}

func TestTalkChannel_DeafSenderBlockedWithoutAutoClear(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8002, Name: "Room"}
	w.Rooms[8002] = room
	sender, sClient := makeMortalInRoom(room, "Piper")
	defer sClient.Close()
	sender.Deaf.Set(types.CHANNEL_MUSIC)
	listener, lClient := makeMortalInRoom(room, "Ear")
	defer lClient.Close()

	talkChannel(sender, "la", types.CHANNEL_MUSIC, "music", nil)

	sOut := readOutput(sender, sClient)
	if !strings.Contains(strings.ToLower(sOut), "don't have the music channel turned on") {
		t.Errorf("deaf-sender diagnostic missing; got %q", sOut)
	}
	if !sender.Deaf.IsSet(types.CHANNEL_MUSIC) {
		t.Error("deaf bit must NOT be auto-cleared (C :514 unreachable)")
	}
	lOut := readOutput(listener, lClient)
	if strings.Contains(lOut, "la") {
		t.Errorf("deaf sender must not broadcast; listener got %q", lOut)
	}
}

func TestTalkChannel_DeafSenderWartalkException(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8003, Name: "Room"}
	w.Rooms[8003] = room
	sender, sClient := makeMortalInRoom(room, "Warlord")
	defer sClient.Close()
	sender.Deaf.Set(types.CHANNEL_WARTALK)
	listener, lClient := makeMortalInRoom(room, "Ally")
	defer lClient.Close()

	talkChannel(sender, "strat", types.CHANNEL_WARTALK, "war", nil)

	sOut := readOutput(sender, sClient)
	if !strings.Contains(sOut, "strat") {
		t.Errorf("wartalk-deaf sender must NOT be blocked; expected self-echo; got %q", sOut)
	}
	if strings.Contains(strings.ToLower(sOut), "don't have the war channel") {
		t.Errorf("wartalk exception must suppress diagnostic; got %q", sOut)
	}
	lOut := readOutput(listener, lClient)
	if !strings.Contains(lOut, "strat") {
		t.Errorf("listener should receive the wartalk broadcast; got %q", lOut)
	}
}

func TestTalkChannel_DeafSenderYellException(t *testing.T) {
	// CHANNEL_YELL has the same exception per C's intent (act_comm.c:505-513).
	// This test is a regression guard for future callers.
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8013, Name: "Room"}
	w.Rooms[8013] = room
	sender, sClient := makeMortalInRoom(room, "Yeller")
	defer sClient.Close()
	sender.Deaf.Set(types.CHANNEL_YELL)
	listener, lClient := makeMortalInRoom(room, "Ear")
	defer lClient.Close()

	talkChannel(sender, "loud", types.CHANNEL_YELL, "yell", nil)

	sOut := readOutput(sender, sClient)
	if !strings.Contains(sOut, "loud") {
		t.Errorf("yell-deaf sender must NOT be blocked (exception); got %q", sOut)
	}
	lOut := readOutput(listener, lClient)
	if !strings.Contains(lOut, "loud") {
		t.Errorf("listener should receive the yell broadcast; got %q", lOut)
	}
}

func TestTalkChannel_RoomSilenceBlocksSender(t *testing.T) {
	w := setupCommWorld()
	silent := &types.RoomIndexData{Vnum: 8004, Name: "Quiet"}
	silent.RoomFlags.Set(types.ROOM_SILENCE)
	w.Rooms[8004] = silent
	open := &types.RoomIndexData{Vnum: 8005, Name: "Open"}
	w.Rooms[8005] = open

	sender, sClient := makeMortalInRoom(silent, "Piper")
	defer sClient.Close()
	listener, lClient := makeMortalInRoom(open, "Ear")
	defer lClient.Close()

	talkChannel(sender, "la", types.CHANNEL_MUSIC, "music", nil)

	sOut := readOutput(sender, sClient)
	if !strings.Contains(sOut, "You can't do that here") {
		t.Errorf("room-silence diagnostic missing; got %q", sOut)
	}
	lOut := readOutput(listener, lClient)
	if strings.Contains(lOut, "la") {
		t.Errorf("silenced-room sender must not broadcast; got %q", lOut)
	}
}

func TestTalkChannel_ReceiverRoomSilenceSkipped(t *testing.T) {
	w := setupCommWorld()
	open := &types.RoomIndexData{Vnum: 8006, Name: "Open"}
	w.Rooms[8006] = open
	silent := &types.RoomIndexData{Vnum: 8007, Name: "Quiet"}
	silent.RoomFlags.Set(types.ROOM_SILENCE)
	w.Rooms[8007] = silent

	sender, sClient := makeMortalInRoom(open, "Piper")
	defer sClient.Close()
	silenced, xClient := makeMortalInRoom(silent, "Mute")
	defer xClient.Close()
	hearer, hClient := makeMortalInRoom(open, "Ear")
	defer hClient.Close()

	talkChannel(sender, "song", types.CHANNEL_MUSIC, "music", nil)

	xOut := readOutput(silenced, xClient)
	if strings.Contains(xOut, "song") {
		t.Errorf("listener in ROOM_SILENCE must not receive; got %q", xOut)
	}
	hOut := readOutput(hearer, hClient)
	if !strings.Contains(hOut, "song") {
		t.Errorf("listener in open room should receive; got %q", hOut)
	}
}

func TestTalkChannel_ReceiverDeafSkipped(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8008, Name: "Room"}
	w.Rooms[8008] = room
	sender, sClient := makeMortalInRoom(room, "Piper")
	defer sClient.Close()
	deaf, dClient := makeMortalInRoom(room, "Deaf")
	defer dClient.Close()
	deaf.Deaf.Set(types.CHANNEL_MUSIC)
	hearer, hClient := makeMortalInRoom(room, "Ear")
	defer hClient.Close()

	talkChannel(sender, "song", types.CHANNEL_MUSIC, "music", nil)

	dOut := readOutput(deaf, dClient)
	if strings.Contains(dOut, "song") {
		t.Errorf("Deaf[MUSIC] listener must not receive; got %q", dOut)
	}
	hOut := readOutput(hearer, hClient)
	if !strings.Contains(hOut, "song") {
		t.Errorf("non-deaf listener should receive; got %q", hOut)
	}
}

func TestTalkChannel_SelfEchoNotBroadcastToSelf(t *testing.T) {
	// Sender's descriptor is on WorldRef.Descriptors; the vch==ch guard must
	// prevent a second copy of the line from arriving as a broadcast.
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8009, Name: "Room"}
	w.Rooms[8009] = room
	sender, sClient := makeMortalInRoom(room, "Solo")
	defer sClient.Close()

	talkChannel(sender, "hum", types.CHANNEL_MUSIC, "music", nil)
	out := readOutput(sender, sClient)

	// Self-echo format: "You music 'hum'". A second broadcast copy would
	// read "Solo musics 'hum'".
	if !strings.Contains(out, "You music") {
		t.Errorf("self-echo missing; got %q", out)
	}
	if strings.Contains(out, "Solo musics") {
		t.Errorf("sender must NOT receive a second broadcast copy; got %q", out)
	}
}

func TestTalkChannel_FilterClosureGatesReceivers(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8010, Name: "Room"}
	w.Rooms[8010] = room
	sender, sClient := makeMortalInRoom(room, "Speaker")
	defer sClient.Close()
	a, aClient := makeMortalInRoom(room, "Rejected")
	defer aClient.Close()
	b, bClient := makeMortalInRoom(room, "AlsoRejected")
	defer bClient.Close()

	// Filter rejects every listener.
	talkChannel(sender, "hi", types.CHANNEL_MUSIC, "music", func(*types.CharData) bool { return false })

	sOut := readOutput(sender, sClient)
	if !strings.Contains(sOut, "hi") {
		t.Errorf("self-echo should still fire; got %q", sOut)
	}
	if strings.Contains(readOutput(a, aClient), "hi") {
		t.Errorf("filter-rejected listener must not receive")
	}
	if strings.Contains(readOutput(b, bClient), "hi") {
		t.Errorf("filter-rejected listener must not receive")
	}
}

func TestTalkChannel_TranslateForIsApplied(t *testing.T) {
	// Matched-language listeners see argument verbatim — asserts threading
	// through translateFor, not the scrambler itself.
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8011, Name: "Room"}
	w.Rooms[8011] = room
	sender, sClient := makeMortalInRoom(room, "Speaker")
	defer sClient.Close()
	sender.Speaking = int(types.LANG_COMMON)
	listener, lClient := makeMortalInRoom(room, "Ear")
	defer lClient.Close()
	listener.Speaks = int(types.LANG_COMMON)

	talkChannel(sender, "hello", types.CHANNEL_MUSIC, "music", nil)

	lOut := readOutput(listener, lClient)
	if !strings.Contains(lOut, "hello") {
		t.Errorf("matched-language listener should hear verbatim; got %q", lOut)
	}
	_ = readOutput(sender, sClient)
}

// --- G1: DoMusic ---

func TestDoMusic_EmptyArgPrintsMusicWhat(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8100, Name: "Room"}
	w.Rooms[8100] = room
	ch, client := makeMortalInRoom(room, "Piper")
	defer client.Close()

	DoMusic(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Music what?") {
		t.Errorf("expected 'Music what?'; got %q", out)
	}
}

func TestDoMusic_NPCGetsHuh(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8101, Name: "Room"}
	w.Rooms[8101] = room
	mob, mClient := makeTestChar("Mob")
	defer mClient.Close()
	mob.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(mob, room)
	w.AddChar(mob)
	addPlayingDescriptor(mob)

	DoMusic(mob, "la")
	out := readOutput(mob, mClient)
	if !strings.Contains(out, "Huh?") {
		t.Errorf("NPC should see 'Huh?'; got %q", out)
	}
}

func TestDoMusic_PublicBroadcastReachesAllPlayers(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8102, Name: "Room"}
	w.Rooms[8102] = room
	sender, sClient := makeMortalInRoom(room, "Piper")
	defer sClient.Close()
	l1, l1Client := makeMortalInRoom(room, "Listener1")
	defer l1Client.Close()
	l2, l2Client := makeMortalInRoom(room, "Listener2")
	defer l2Client.Close()

	DoMusic(sender, "song")

	if !strings.Contains(readOutput(sender, sClient), "song") {
		t.Error("sender should see own echo")
	}
	if !strings.Contains(readOutput(l1, l1Client), "song") {
		t.Error("listener1 should receive")
	}
	if !strings.Contains(readOutput(l2, l2Client), "song") {
		t.Error("listener2 should receive")
	}
}

func TestDoMusic_DeafListenerSkipped(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8103, Name: "Room"}
	w.Rooms[8103] = room
	sender, sClient := makeMortalInRoom(room, "Piper")
	defer sClient.Close()
	deaf, dClient := makeMortalInRoom(room, "Deaf")
	defer dClient.Close()
	deaf.Deaf.Set(types.CHANNEL_MUSIC)

	DoMusic(sender, "song")
	_ = readOutput(sender, sClient)
	if strings.Contains(readOutput(deaf, dClient), "song") {
		t.Error("deaf[MUSIC] listener must not receive")
	}
}

func TestDoMusic_BootRegistered(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8104, Name: "Room"}
	w.Rooms[8104] = room

	reg := command.NewRegistry()
	reg.Register(&command.Command{Name: "music", DoFun: DoMusic, Position: types.POS_SLEEPING, Level: 0})

	sender, sClient := makeMortalInRoom(room, "Piper")
	defer sClient.Close()
	listener, lClient := makeMortalInRoom(room, "Ear")
	defer lClient.Close()

	reg.Interpret(sender, "music la la")

	if !strings.Contains(readOutput(sender, sClient), "la la") {
		t.Error("interpret 'music' should self-echo")
	}
	if !strings.Contains(readOutput(listener, lClient), "la la") {
		t.Error("interpret 'music' should broadcast to listeners")
	}
}

// --- G2: DoRacetalk ---

func TestDoRacetalk_EmptyArgPrintsRacetalkWhat(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8200, Name: "Room"}
	w.Rooms[8200] = room
	ch, client := makeMortalInRoom(room, "Elf")
	defer client.Close()

	DoRacetalk(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Racetalk what?") {
		t.Errorf("expected 'Racetalk what?'; got %q", out)
	}
}

func TestDoRacetalk_NPCGetsHuh(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8201, Name: "Room"}
	w.Rooms[8201] = room
	mob, mClient := makeTestChar("Mob")
	defer mClient.Close()
	mob.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(mob, room)
	w.AddChar(mob)
	addPlayingDescriptor(mob)

	DoRacetalk(mob, "hi")
	if !strings.Contains(readOutput(mob, mClient), "Huh?") {
		t.Error("NPC should get Huh?")
	}
}

func TestDoRacetalk_SameRaceReceives(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8202, Name: "Room"}
	w.Rooms[8202] = room
	sender, sClient := makeMortalInRoom(room, "Elf1")
	defer sClient.Close()
	sender.Race = 1
	listener, lClient := makeMortalInRoom(room, "Elf2")
	defer lClient.Close()
	listener.Race = 1

	DoRacetalk(sender, "hi kin")
	_ = readOutput(sender, sClient)
	if !strings.Contains(readOutput(listener, lClient), "hi kin") {
		t.Error("same-race listener should receive")
	}
}

func TestDoRacetalk_DifferentRaceFiltered(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8203, Name: "Room"}
	w.Rooms[8203] = room
	sender, sClient := makeMortalInRoom(room, "Elf")
	defer sClient.Close()
	sender.Race = 1
	listener, lClient := makeMortalInRoom(room, "Orc")
	defer lClient.Close()
	listener.Race = 2

	DoRacetalk(sender, "hi kin")
	_ = readOutput(sender, sClient)
	if strings.Contains(readOutput(listener, lClient), "hi kin") {
		t.Error("different-race listener must NOT receive")
	}
}

func TestDoRacetalk_BootRegistered(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8204, Name: "Room"}
	w.Rooms[8204] = room
	reg := command.NewRegistry()
	reg.Register(&command.Command{Name: "racetalk", DoFun: DoRacetalk, Position: types.POS_SLEEPING, Level: 0})
	sender, sClient := makeMortalInRoom(room, "Elf1")
	defer sClient.Close()
	sender.Race = 3
	listener, lClient := makeMortalInRoom(room, "Elf2")
	defer lClient.Close()
	listener.Race = 3

	reg.Interpret(sender, "racetalk greetings")
	_ = readOutput(sender, sClient)
	if !strings.Contains(readOutput(listener, lClient), "greetings") {
		t.Error("interpret 'racetalk' should route to DoRacetalk")
	}
}

// --- G3: DoWartalk ---

func TestDoWartalk_EmptyArgPrintsWarWhat(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8300, Name: "Room"}
	w.Rooms[8300] = room
	ch, client := makeMortalPCInRoom(room, "Warlord")
	defer client.Close()
	ch.PCData.Flags = int(types.PCFLAG_DEADLY)

	DoWartalk(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "War what?") {
		t.Errorf("expected 'War what?'; got %q", out)
	}
}

func TestDoWartalk_NPCGetsHuh(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8301, Name: "Room"}
	w.Rooms[8301] = room
	mob, mClient := makeTestChar("Mob")
	defer mClient.Close()
	mob.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(mob, room)
	w.AddChar(mob)
	addPlayingDescriptor(mob)

	DoWartalk(mob, "hi")
	if !strings.Contains(readOutput(mob, mClient), "Huh?") {
		t.Error("NPC should get Huh?")
	}
}

func TestDoWartalk_PeacefulSenderBlocked(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8302, Name: "Room"}
	w.Rooms[8302] = room
	ch, client := makeMortalPCInRoom(room, "Peaceful")
	defer client.Close()
	// Flags deliberately NOT set — peaceful.

	DoWartalk(ch, "strat")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Peacefuls have no need to use wartalk") {
		t.Errorf("expected peaceful block message; got %q", out)
	}
}

func TestDoWartalk_PkillToPkillReceives(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8303, Name: "Room"}
	w.Rooms[8303] = room
	sender, sClient := makeMortalPCInRoom(room, "Warlord")
	defer sClient.Close()
	sender.PCData.Flags = int(types.PCFLAG_DEADLY)
	listener, lClient := makeMortalPCInRoom(room, "Ally")
	defer lClient.Close()
	listener.PCData.Flags = int(types.PCFLAG_DEADLY)

	DoWartalk(sender, "attack east")
	_ = readOutput(sender, sClient)
	if !strings.Contains(readOutput(listener, lClient), "attack east") {
		t.Error("pkill listener should receive wartalk")
	}
}

func TestDoWartalk_PkillToPeacefulFiltered(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8304, Name: "Room"}
	w.Rooms[8304] = room
	sender, sClient := makeMortalPCInRoom(room, "Warlord")
	defer sClient.Close()
	sender.PCData.Flags = int(types.PCFLAG_DEADLY)
	listener, lClient := makeMortalPCInRoom(room, "Peaceful")
	defer lClient.Close()
	// listener NOT pkill.

	DoWartalk(sender, "attack east")
	_ = readOutput(sender, sClient)
	if strings.Contains(readOutput(listener, lClient), "attack east") {
		t.Error("peaceful listener must NOT receive wartalk")
	}
}

func TestDoWartalk_DeafSenderNotBlocked_WartalkException(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8305, Name: "Room"}
	w.Rooms[8305] = room
	sender, sClient := makeMortalPCInRoom(room, "Warlord")
	defer sClient.Close()
	sender.PCData.Flags = int(types.PCFLAG_DEADLY)
	sender.Deaf.Set(types.CHANNEL_WARTALK)
	listener, lClient := makeMortalPCInRoom(room, "Ally")
	defer lClient.Close()
	listener.PCData.Flags = int(types.PCFLAG_DEADLY)

	DoWartalk(sender, "strat")
	sOut := readOutput(sender, sClient)
	if !strings.Contains(sOut, "strat") {
		t.Errorf("wartalk-deaf pkill sender must NOT be blocked; got %q", sOut)
	}
	if !strings.Contains(readOutput(listener, lClient), "strat") {
		t.Error("listener should receive despite sender's wartalk-deaf bit")
	}
}

func TestDoWartalk_BootRegistered(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8306, Name: "Room"}
	w.Rooms[8306] = room
	reg := command.NewRegistry()
	reg.Register(&command.Command{Name: "wartalk", DoFun: DoWartalk, Position: types.POS_SLEEPING, Level: 0})

	sender, sClient := makeMortalPCInRoom(room, "Warlord")
	defer sClient.Close()
	sender.PCData.Flags = int(types.PCFLAG_DEADLY)
	listener, lClient := makeMortalPCInRoom(room, "Ally")
	defer lClient.Close()
	listener.PCData.Flags = int(types.PCFLAG_DEADLY)

	reg.Interpret(sender, "wartalk ready")
	_ = readOutput(sender, sClient)
	if !strings.Contains(readOutput(listener, lClient), "ready") {
		t.Error("interpret 'wartalk' should broadcast to pkill listener")
	}
}

// --- G4: DoCouncilTalk ---

func TestDoCouncilTalk_EmptyArgPrintsCounciltalkWhat(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8400, Name: "Room"}
	w.Rooms[8400] = room
	council := &types.CouncilData{Name: "Mages"}
	ch, client := makeMortalPCInRoom(room, "Mage")
	defer client.Close()
	ch.PCData.Council = council

	DoCouncilTalk(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Counciltalk what?") {
		t.Errorf("expected 'Counciltalk what?'; got %q", out)
	}
}

func TestDoCouncilTalk_NPCGetsHuh(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8401, Name: "Room"}
	w.Rooms[8401] = room
	mob, mClient := makeTestChar("Mob")
	defer mClient.Close()
	mob.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(mob, room)
	w.AddChar(mob)
	addPlayingDescriptor(mob)

	DoCouncilTalk(mob, "hi")
	if !strings.Contains(readOutput(mob, mClient), "Huh?") {
		t.Error("NPC should get Huh?")
	}
}

func TestDoCouncilTalk_NoCouncilGetsHuh(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8402, Name: "Room"}
	w.Rooms[8402] = room
	ch, client := makeMortalPCInRoom(room, "Loner")
	defer client.Close()
	// ch.PCData.Council is nil.

	DoCouncilTalk(ch, "hello")
	if !strings.Contains(readOutput(ch, client), "Huh?") {
		t.Error("no-council sender should get Huh?")
	}
}

func TestDoCouncilTalk_SameCouncilReceives(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8403, Name: "Room"}
	w.Rooms[8403] = room
	council := &types.CouncilData{Name: "Mages"}
	sender, sClient := makeMortalPCInRoom(room, "Magus1")
	defer sClient.Close()
	sender.PCData.Council = council
	listener, lClient := makeMortalPCInRoom(room, "Magus2")
	defer lClient.Close()
	listener.PCData.Council = council

	DoCouncilTalk(sender, "convene")
	_ = readOutput(sender, sClient)
	if !strings.Contains(readOutput(listener, lClient), "convene") {
		t.Error("same-council listener should receive")
	}
}

func TestDoCouncilTalk_DifferentCouncilFiltered(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8404, Name: "Room"}
	w.Rooms[8404] = room
	mages := &types.CouncilData{Name: "Mages"}
	thieves := &types.CouncilData{Name: "Thieves"}
	sender, sClient := makeMortalPCInRoom(room, "Magus")
	defer sClient.Close()
	sender.PCData.Council = mages
	listener, lClient := makeMortalPCInRoom(room, "Thief")
	defer lClient.Close()
	listener.PCData.Council = thieves

	DoCouncilTalk(sender, "convene")
	_ = readOutput(sender, sClient)
	if strings.Contains(readOutput(listener, lClient), "convene") {
		t.Error("different-council listener must NOT receive")
	}
}

func TestDoCouncilTalk_NilPCDataListenerSkipped(t *testing.T) {
	// Safety: NPC listeners (PCData nil) must not panic and must be filtered.
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8405, Name: "Room"}
	w.Rooms[8405] = room
	council := &types.CouncilData{Name: "Mages"}
	sender, sClient := makeMortalPCInRoom(room, "Magus")
	defer sClient.Close()
	sender.PCData.Council = council

	mob, _ := makeTestChar("Pet")
	mob.Act.Set(types.ACT_IS_NPC)
	mob.PCData = nil
	handler.CharToRoom(mob, room)
	w.AddChar(mob)
	addPlayingDescriptor(mob)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panicked on nil-PCData listener: %v", r)
		}
	}()
	DoCouncilTalk(sender, "convene")
	_ = readOutput(sender, sClient)
}

func TestDoCouncilTalk_BootRegistered(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8406, Name: "Room"}
	w.Rooms[8406] = room
	reg := command.NewRegistry()
	reg.Register(&command.Command{Name: "counciltalk", DoFun: DoCouncilTalk, Position: types.POS_SLEEPING, Level: 0})

	council := &types.CouncilData{Name: "Mages"}
	sender, sClient := makeMortalPCInRoom(room, "Magus1")
	defer sClient.Close()
	sender.PCData.Council = council
	listener, lClient := makeMortalPCInRoom(room, "Magus2")
	defer lClient.Close()
	listener.PCData.Council = council

	reg.Interpret(sender, "counciltalk meet")
	_ = readOutput(sender, sClient)
	if !strings.Contains(readOutput(listener, lClient), "meet") {
		t.Error("interpret 'counciltalk' should broadcast to same-council listener")
	}
}

// --- G5: DoGuildTalk + DoNewbieChat ---

func TestDoGuildTalk_NPCGetsHuh(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8500, Name: "Room"}
	w.Rooms[8500] = room
	mob, mClient := makeTestChar("Mob")
	defer mClient.Close()
	mob.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(mob, room)
	w.AddChar(mob)
	addPlayingDescriptor(mob)

	DoGuildTalk(mob, "hi")
	if !strings.Contains(readOutput(mob, mClient), "Huh?") {
		t.Error("NPC should get Huh?")
	}
}

func TestDoGuildTalk_NoClanGetsHuh(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8501, Name: "Room"}
	w.Rooms[8501] = room
	ch, client := makeMortalPCInRoom(room, "Solo")
	defer client.Close()

	DoGuildTalk(ch, "hi")
	if !strings.Contains(readOutput(ch, client), "Huh?") {
		t.Error("no-clan sender should get Huh?")
	}
}

func TestDoGuildTalk_WrongClanTypeGetsHuh(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8502, Name: "Room"}
	w.Rooms[8502] = room
	order := &types.ClanData{Name: "Order", ClanType: types.CLAN_ORDER}
	ch, client := makeMortalPCInRoom(room, "Cleric")
	defer client.Close()
	ch.PCData.Clan = order

	DoGuildTalk(ch, "hi")
	if !strings.Contains(readOutput(ch, client), "Huh?") {
		t.Error("wrong-clan-type sender should get Huh?")
	}
}

func TestDoGuildTalk_SameGuildReceives(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8503, Name: "Room"}
	w.Rooms[8503] = room
	guild := &types.ClanData{Name: "Thieves", ClanType: types.CLAN_GUILD}
	sender, sClient := makeMortalPCInRoom(room, "Thief1")
	defer sClient.Close()
	sender.PCData.Clan = guild
	listener, lClient := makeMortalPCInRoom(room, "Thief2")
	defer lClient.Close()
	listener.PCData.Clan = guild

	DoGuildTalk(sender, "meet 12th")
	_ = readOutput(sender, sClient)
	if !strings.Contains(readOutput(listener, lClient), "meet 12th") {
		t.Error("same-guild listener should receive")
	}
}

func TestDoGuildTalk_DifferentClanFiltered(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8504, Name: "Room"}
	w.Rooms[8504] = room
	thieves := &types.ClanData{Name: "Thieves", ClanType: types.CLAN_GUILD}
	mages := &types.ClanData{Name: "Mages", ClanType: types.CLAN_GUILD}
	sender, sClient := makeMortalPCInRoom(room, "Thief")
	defer sClient.Close()
	sender.PCData.Clan = thieves
	listener, lClient := makeMortalPCInRoom(room, "Mage")
	defer lClient.Close()
	listener.PCData.Clan = mages

	DoGuildTalk(sender, "meet 12th")
	_ = readOutput(sender, sClient)
	if strings.Contains(readOutput(listener, lClient), "meet 12th") {
		t.Error("different-clan listener must NOT receive")
	}
}

func TestDoGuildTalk_BootRegistered(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8505, Name: "Room"}
	w.Rooms[8505] = room
	reg := command.NewRegistry()
	reg.Register(&command.Command{Name: "guildtalk", DoFun: DoGuildTalk, Position: types.POS_SLEEPING, Level: 0})

	guild := &types.ClanData{Name: "Thieves", ClanType: types.CLAN_GUILD}
	sender, sClient := makeMortalPCInRoom(room, "Thief1")
	defer sClient.Close()
	sender.PCData.Clan = guild
	listener, lClient := makeMortalPCInRoom(room, "Thief2")
	defer lClient.Close()
	listener.PCData.Clan = guild

	reg.Interpret(sender, "guildtalk heist")
	_ = readOutput(sender, sClient)
	if !strings.Contains(readOutput(listener, lClient), "heist") {
		t.Error("interpret 'guildtalk' should broadcast to guild listener")
	}
}

func TestDoNewbieChat_NPCGetsHuh(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8600, Name: "Room"}
	w.Rooms[8600] = room
	mob, mClient := makeTestChar("Mob")
	defer mClient.Close()
	mob.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(mob, room)
	w.AddChar(mob)
	addPlayingDescriptor(mob)

	DoNewbieChat(mob, "hi")
	if !strings.Contains(readOutput(mob, mClient), "Huh?") {
		t.Error("NPC should get Huh?")
	}
}

func TestDoNewbieChat_MortalNonNewbieGetsHuh(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8601, Name: "Room"}
	w.Rooms[8601] = room
	ch, client := makeMortalPCInRoom(room, "Mort")
	defer client.Close()
	// No council set.

	DoNewbieChat(ch, "hi")
	if !strings.Contains(readOutput(ch, client), "Huh?") {
		t.Error("non-newbie mortal should get Huh?")
	}
}

func TestDoNewbieChat_MortalNewbieCouncilReceives(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8602, Name: "Room"}
	w.Rooms[8602] = room
	newbieCouncil := &types.CouncilData{Name: "Newbie Council"}
	sender, sClient := makeMortalPCInRoom(room, "Newbie1")
	defer sClient.Close()
	sender.PCData.Council = newbieCouncil
	listener, lClient := makeMortalPCInRoom(room, "Newbie2")
	defer lClient.Close()
	listener.PCData.Council = newbieCouncil

	DoNewbieChat(sender, "hi all")
	_ = readOutput(sender, sClient)
	if !strings.Contains(readOutput(listener, lClient), "hi all") {
		t.Error("newbie-council mortal should receive")
	}
}

func TestDoNewbieChat_ImmortalReceives(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8603, Name: "Room"}
	w.Rooms[8603] = room
	newbieCouncil := &types.CouncilData{Name: "Newbie Council"}
	sender, sClient := makeMortalPCInRoom(room, "Newbie")
	defer sClient.Close()
	sender.PCData.Council = newbieCouncil
	imm, iClient := makeImmortalInRoom(room, "Zeus")
	defer iClient.Close()
	// Immortal has no council — still receives per filter.

	DoNewbieChat(sender, "help please")
	_ = readOutput(sender, sClient)
	if !strings.Contains(readOutput(imm, iClient), "help please") {
		t.Error("immortal should receive newbiechat")
	}
}

func TestDoNewbieChat_ImmortalCanSend(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8604, Name: "Room"}
	w.Rooms[8604] = room
	imm, iClient := makeImmortalInRoom(room, "Zeus")
	defer iClient.Close()
	newbieCouncil := &types.CouncilData{Name: "Newbie Council"}
	listener, lClient := makeMortalPCInRoom(room, "Newbie")
	defer lClient.Close()
	listener.PCData.Council = newbieCouncil

	DoNewbieChat(imm, "welcome")
	_ = readOutput(imm, iClient)
	if !strings.Contains(readOutput(listener, lClient), "welcome") {
		t.Error("newbie listener should receive from immortal sender")
	}
}

func TestDoNewbieChat_CaseInsensitiveCouncilName(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8605, Name: "Room"}
	w.Rooms[8605] = room
	// Mixed case and different cases both satisfy.
	lcase := &types.CouncilData{Name: "newbie council"}
	ucase := &types.CouncilData{Name: "NEWBIE COUNCIL"}

	sender, sClient := makeMortalPCInRoom(room, "Newb1")
	defer sClient.Close()
	sender.PCData.Council = lcase
	listener, lClient := makeMortalPCInRoom(room, "Newb2")
	defer lClient.Close()
	listener.PCData.Council = ucase

	DoNewbieChat(sender, "hi")
	_ = readOutput(sender, sClient)
	if !strings.Contains(readOutput(listener, lClient), "hi") {
		t.Error("case-insensitive match should let upper-case council listener receive")
	}
}

func TestDoNewbieChat_BootRegistered(t *testing.T) {
	w := setupCommWorld()
	room := &types.RoomIndexData{Vnum: 8606, Name: "Room"}
	w.Rooms[8606] = room
	reg := command.NewRegistry()
	reg.Register(&command.Command{Name: "newbiechat", DoFun: DoNewbieChat, Position: types.POS_SLEEPING, Level: 0})

	newbieCouncil := &types.CouncilData{Name: "Newbie Council"}
	sender, sClient := makeMortalPCInRoom(room, "Newb1")
	defer sClient.Close()
	sender.PCData.Council = newbieCouncil
	listener, lClient := makeMortalPCInRoom(room, "Newb2")
	defer lClient.Close()
	listener.PCData.Council = newbieCouncil

	reg.Interpret(sender, "newbiechat hi")
	_ = readOutput(sender, sClient)
	if !strings.Contains(readOutput(listener, lClient), "hi") {
		t.Error("interpret 'newbiechat' should broadcast")
	}
}
