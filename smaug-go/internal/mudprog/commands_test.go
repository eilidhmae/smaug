package mudprog

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/command"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// setupTestWorld creates a minimal world for command testing.
func setupTestWorld() *world.World {
	w := world.New("/tmp/test")
	room := &types.RoomIndexData{Vnum: 3001, Name: "Test Room"}
	w.Rooms[3001] = room
	room2 := &types.RoomIndexData{Vnum: 3002, Name: "Other Room"}
	w.Rooms[3002] = room2
	return w
}

// makeDescChar creates a char with a net.Pipe descriptor for output capture.
func makeDescChar(name string) (*types.CharData, net.Conn) {
	server, client := net.Pipe()
	d := &types.DescriptorData{
		Conn:       server,
		InputQueue: make(chan string, 10),
		Connected:  types.CON_PLAYING,
	}
	ch := &types.CharData{
		Name:       name,
		ShortDescr: name,
		Level:      10,
		Position:   types.POS_STANDING,
		Hit:        100,
		MaxHit:     100,
		Desc:       d,
	}
	d.Character = ch
	return ch, client
}

// flushAndRead flushes the descriptor and reads the output.
func flushAndRead(ch *types.CharData, client net.Conn) string {
	if !ch.Desc.HasOutput() {
		return ""
	}
	result := make(chan string, 1)
	go func() {
		buf := make([]byte, 8192)
		client.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, _ := client.Read(buf)
		result <- string(buf[:n])
	}()
	_ = ch.Desc.FlushOutput()
	return <-result
}

func TestMpEcho(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001, Name: "Test Room"}
	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	ch1, client1 := makeDescChar("Player1")
	defer client1.Close()
	ch1.InRoom = room
	room.People = append(room.People, ch1)

	ch2, client2 := makeDescChar("Player2")
	defer client2.Close()
	ch2.InRoom = room
	room.People = append(room.People, ch2)

	mpEcho(mob, " The ground shakes!")

	out1 := flushAndRead(ch1, client1)
	out2 := flushAndRead(ch2, client2)
	if !strings.Contains(out1, "The ground shakes!") {
		t.Errorf("player1 should see echo, got %q", out1)
	}
	if !strings.Contains(out2, "The ground shakes!") {
		t.Errorf("player2 should see echo, got %q", out2)
	}
}

func TestMpEcho_EmptyArgs(t *testing.T) {
	mob := makeNPC("guard")
	mob.InRoom = &types.RoomIndexData{Vnum: 3001}
	// Should not panic
	mpEcho(mob, "")
	mpEcho(mob, "   ")
}

func TestMpEcho_NilRoom(t *testing.T) {
	mob := makeNPC("guard")
	mob.InRoom = nil
	// Should not panic
	mpEcho(mob, "hello")
}

func TestMpEchoAt(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001, Name: "Test Room"}
	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	target, client := makeDescChar("TargetPlayer")
	defer client.Close()
	target.InRoom = room
	room.People = append(room.People, target)

	bystander, byClient := makeDescChar("Bystander")
	defer byClient.Close()
	bystander.InRoom = room
	room.People = append(room.People, bystander)

	mpEchoAt(mob, " TargetPlayer You feel a chill!")

	outTarget := flushAndRead(target, client)
	outBystander := flushAndRead(bystander, byClient)
	if !strings.Contains(outTarget, "You feel a chill!") {
		t.Errorf("target should see message, got %q", outTarget)
	}
	if strings.Contains(outBystander, "You feel a chill!") {
		t.Errorf("bystander should NOT see message, got %q", outBystander)
	}
}

func TestMpEchoAt_EmptyArgs(t *testing.T) {
	mob := makeNPC("guard")
	mob.InRoom = &types.RoomIndexData{Vnum: 3001}
	// Should not panic with empty args
	mpEchoAt(mob, "")
	mpEchoAt(mob, " player")    // no message
	mpEchoAt(mob, "  ")
}

func TestMpEchoAround(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001, Name: "Test Room"}
	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	target, targetClient := makeDescChar("TargetPlayer")
	defer targetClient.Close()
	target.InRoom = room
	room.People = append(room.People, target)

	bystander, byClient := makeDescChar("Bystander")
	defer byClient.Close()
	bystander.InRoom = room
	room.People = append(room.People, bystander)

	mpEchoAround(mob, " TargetPlayer laughs maniacally!")

	outTarget := flushAndRead(target, targetClient)
	outBystander := flushAndRead(bystander, byClient)
	if strings.Contains(outTarget, "laughs maniacally!") {
		t.Errorf("target should NOT see echoaround, got %q", outTarget)
	}
	if !strings.Contains(outBystander, "laughs maniacally!") {
		t.Errorf("bystander should see echoaround, got %q", outBystander)
	}
}

func TestMpEchoAround_EmptyArgs(t *testing.T) {
	mob := makeNPC("guard")
	mob.InRoom = &types.RoomIndexData{Vnum: 3001}
	// Should not panic
	mpEchoAround(mob, "")
	mpEchoAround(mob, " player")  // no message
}

func TestMpEchoAround_NilRoom(t *testing.T) {
	mob := makeNPC("guard")
	mob.InRoom = nil
	mpEchoAround(mob, " player message")
}

func TestMpGoto(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	room1 := w.Rooms[3001]
	room2 := w.Rooms[3002]

	mob := makeNPC("guard")
	handler.CharToRoom(mob, room1)

	mpGoto(mob, " 3002")

	if mob.InRoom != room2 {
		t.Errorf("mob should be in room 3002 after mpgoto, got vnum %d", mob.InRoom.Vnum)
	}
}

func TestMpGoto_InvalidVnum(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	room1 := w.Rooms[3001]
	mob := makeNPC("guard")
	handler.CharToRoom(mob, room1)

	mpGoto(mob, " 99999")

	if mob.InRoom != room1 {
		t.Errorf("mob should stay in room 3001 with invalid vnum")
	}
}

func TestMpGoto_NilWorld(t *testing.T) {
	oldWorld := WorldRef
	WorldRef = nil
	defer func() { WorldRef = oldWorld }()

	mob := makeNPC("guard")
	// Should not panic
	mpGoto(mob, " 3001")
}

func TestMpGoto_BadArg(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	mob := makeNPC("guard")
	mob.InRoom = w.Rooms[3001]
	// Non-numeric arg
	mpGoto(mob, " notanumber")
	if mob.InRoom.Vnum != 3001 {
		t.Errorf("mob should stay in place with bad arg")
	}
}

func TestMpTransfer(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	room1 := w.Rooms[3001]
	room2 := w.Rooms[3002]

	mob := makeNPC("guard")
	handler.CharToRoom(mob, room1)
	w.AddChar(mob)

	victim := &types.CharData{Name: "Player", Level: 10, Position: types.POS_STANDING}
	handler.CharToRoom(victim, room2)
	w.AddChar(victim)

	// Transfer to mob's room
	mpTransfer(mob, " Player")

	if victim.InRoom != room1 {
		t.Errorf("victim should be in mob's room 3001, got vnum %d", victim.InRoom.Vnum)
	}
}

func TestMpTransfer_WithDestVnum(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	room1 := w.Rooms[3001]
	room2 := w.Rooms[3002]

	mob := makeNPC("guard")
	handler.CharToRoom(mob, room1)
	w.AddChar(mob)

	victim := &types.CharData{Name: "Player", Level: 10, Position: types.POS_STANDING}
	handler.CharToRoom(victim, room1)
	w.AddChar(victim)

	// Transfer to specific room
	mpTransfer(mob, " Player 3002")

	if victim.InRoom != room2 {
		t.Errorf("victim should be in room 3002, got vnum %d", victim.InRoom.Vnum)
	}
}

func TestMpTransfer_EmptyArgs(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	mob := makeNPC("guard")
	// Should not panic
	mpTransfer(mob, "")
}

func TestMpTransfer_NilWorld(t *testing.T) {
	oldWorld := WorldRef
	WorldRef = nil
	defer func() { WorldRef = oldWorld }()

	mob := makeNPC("guard")
	mpTransfer(mob, " Player")
}

func TestMpForce(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()
	oldReg := CmdRegistry
	defer func() { CmdRegistry = oldReg }()

	var forcedCmd string
	reg := command.NewRegistry()
	reg.Register(&command.Command{
		Name:  "say",
		Level: 1,
		DoFun: func(ch *types.CharData, argument string) {
			forcedCmd = argument
		},
	})
	CmdRegistry = reg

	room := w.Rooms[3001]
	mob := makeNPC("guard")
	handler.CharToRoom(mob, room)
	w.AddChar(mob)

	victim := &types.CharData{Name: "Player", Level: 10, Position: types.POS_STANDING}
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	mpForce(mob, " Player say hello world")

	if forcedCmd != "hello world" {
		t.Errorf("expected forced command 'hello world', got %q", forcedCmd)
	}
}

func TestMpForce_EmptyArgs(t *testing.T) {
	mob := makeNPC("guard")
	// Should not panic
	mpForce(mob, "")
	mpForce(mob, " player")  // no command
}

// TestMpForce_TrustCap covers security finding S3 (mudprog variant): mpForce
// must cap the forced command's effective trust level at the caller mob's
// own trust (via GetTrust). Otherwise a crafted area file could mpforce an
// immortal PC to run a command registered at Level: LEVEL_IMMORTAL — the
// command would dispatch under the victim's trust instead of the (lower)
// mob's. With a proper trust cap, Find() returns nil for such a command and
// the dispatcher falls through without invoking it.
//
// NOTE: GetTrust() returns the NPC's Level when it's below LEVEL_IMMORTAL
// (types/character.go:384). makeNPC sets Level=10, so this mob's effective
// trust is 10, far below the immortal-only command we register below.
func TestMpForce_TrustCap(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()
	oldReg := CmdRegistry
	defer func() { CmdRegistry = oldReg }()

	var invoked bool
	reg := command.NewRegistry()
	reg.Register(&command.Command{
		Name:  "godcmd",
		Level: types.LEVEL_IMMORTAL,
		DoFun: func(ch *types.CharData, argument string) {
			invoked = true
		},
	})
	CmdRegistry = reg

	room := w.Rooms[3001]
	mob := makeNPC("guard") // Level=10 ⇒ GetTrust()=10
	handler.CharToRoom(mob, room)
	w.AddChar(mob)

	// High-trust victim whose own trust would be enough to run godcmd.
	victim, client := makePlayerInRoom("Imm", room)
	defer client.Close()
	victim.Level = types.LEVEL_IMMORTAL
	w.AddChar(victim)

	mpForce(mob, " Imm godcmd")

	if invoked {
		t.Error("mpForce must cap effective trust at mob's GetTrust() so an " +
			"immortal-only command is refused when the forcing mob is low-trust")
	}
}

// TestMpForce_PositiveDispatch is the positive-direction counterpart to
// TestMpForce_TrustCap: when the mob's own GetTrust() is high enough for the
// forced command's Level, the dispatcher MUST invoke it. Without this, a
// regression that over-restricts mpForce (e.g. caps trust to 0, never
// dispatches) would pass the suite silently.
func TestMpForce_PositiveDispatch(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()
	oldReg := CmdRegistry
	defer func() { CmdRegistry = oldReg }()

	var invoked bool
	var invokedArg string
	reg := command.NewRegistry()
	reg.Register(&command.Command{
		Name:  "say",
		Level: 1, // well under the mob's trust (10)
		DoFun: func(ch *types.CharData, argument string) {
			invoked = true
			invokedArg = argument
		},
	})
	CmdRegistry = reg

	room := w.Rooms[3001]
	mob := makeNPC("guard") // Level=10 ⇒ GetTrust()=10
	handler.CharToRoom(mob, room)
	w.AddChar(mob)

	victim := &types.CharData{Name: "Player", Level: 10, Position: types.POS_STANDING}
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	mpForce(mob, " Player say hello world")

	if !invoked {
		t.Fatal("mpForce should dispatch when mob's GetTrust() >= command Level")
	}
	if invokedArg != "hello world" {
		t.Errorf("expected forced command arg 'hello world', got %q", invokedArg)
	}
}

func TestMpKill(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001, Name: "Test Room"}
	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	victim := &types.CharData{Name: "Player", Level: 10}
	victim.InRoom = room
	room.People = append(room.People, victim)

	mpKill(mob, " Player")

	if mob.Fighting == nil {
		t.Fatal("mob should be fighting after mpkill")
	}
	if mob.Fighting.Who != victim {
		t.Error("mob should be fighting the victim")
	}
	if mob.Position != types.POS_FIGHTING {
		t.Error("mob should be in POS_FIGHTING")
	}
}

func TestMpKill_Self(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001, Name: "Test Room"}
	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	mpKill(mob, " guard")

	if mob.Fighting != nil {
		t.Error("mob should not fight itself")
	}
}

func TestMpKill_EmptyArgs(t *testing.T) {
	mob := makeNPC("guard")
	mpKill(mob, "")
	if mob.Fighting != nil {
		t.Error("mob should not fight with empty args")
	}
}

func TestMpKill_NotFound(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001, Name: "Test Room"}
	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	mpKill(mob, " nonexistent")

	if mob.Fighting != nil {
		t.Error("mob should not fight when target not found")
	}
}

func TestMpKill_AlreadyFighting(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001, Name: "Test Room"}
	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	existing := &types.CharData{Name: "Existing"}
	mob.Fighting = &types.FightData{Who: existing}

	newVictim := &types.CharData{Name: "NewVictim"}
	newVictim.InRoom = room
	room.People = append(room.People, newVictim)

	mpKill(mob, " NewVictim")

	// Should not change target since already fighting
	if mob.Fighting.Who != existing {
		t.Error("should not change target when already fighting")
	}
}

func TestMpDamage(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001, Name: "Test Room"}
	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	victim := &types.CharData{Name: "Player", Hit: 100, MaxHit: 100}
	victim.InRoom = room
	room.People = append(room.People, victim)

	mpDamage(mob, " Player 30")

	if victim.Hit != 70 {
		t.Errorf("victim HP should be 70, got %d", victim.Hit)
	}
}

func TestMpDamage_NoKill(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001, Name: "Test Room"}
	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	victim := &types.CharData{Name: "Player", Hit: 10, MaxHit: 100}
	victim.InRoom = room
	room.People = append(room.People, victim)

	mpDamage(mob, " Player 50")

	// Mudprog damage doesn't kill - HP should be clamped to 1
	if victim.Hit != 1 {
		t.Errorf("victim HP should be clamped to 1, got %d", victim.Hit)
	}
}

func TestMpDamage_EmptyArgs(t *testing.T) {
	mob := makeNPC("guard")
	// Should not panic
	mpDamage(mob, "")
	mpDamage(mob, " Player")  // no damage amount
}

func TestMpDamage_ZeroDamage(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001, Name: "Test Room"}
	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	victim := &types.CharData{Name: "Player", Hit: 100, MaxHit: 100}
	victim.InRoom = room
	room.People = append(room.People, victim)

	mpDamage(mob, " Player 0")

	if victim.Hit != 100 {
		t.Errorf("zero damage should not change HP, got %d", victim.Hit)
	}
}

func TestMpDamage_NegativeDamage(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001, Name: "Test Room"}
	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	victim := &types.CharData{Name: "Player", Hit: 100, MaxHit: 100}
	victim.InRoom = room
	room.People = append(room.People, victim)

	mpDamage(mob, " Player -10")

	if victim.Hit != 100 {
		t.Errorf("negative damage should not change HP, got %d", victim.Hit)
	}
}

func TestMpPurge_SpecificNPC(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	room := w.Rooms[3001]
	mob := makeNPC("guard")
	handler.CharToRoom(mob, room)
	w.AddChar(mob)

	target := makeNPC("target")
	handler.CharToRoom(target, room)
	w.AddChar(target)

	mpPurge(mob, " target")

	// target should have been extracted from the world
	for _, ch := range w.Characters {
		if ch == target {
			t.Error("target NPC should have been extracted from world")
		}
	}
	// target should no longer be in the room
	for _, ch := range room.People {
		if ch == target {
			t.Error("target NPC should have been removed from room")
		}
	}
}

func TestMpPurge_Self(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	room := w.Rooms[3001]
	mob := makeNPC("guard")
	handler.CharToRoom(mob, room)
	w.AddChar(mob)

	initialChars := len(w.Characters)
	mpPurge(mob, " guard")

	// Should NOT purge self - character count should not decrease
	if len(w.Characters) < initialChars {
		t.Error("mob should not purge itself")
	}
}

func TestMpPurge_Object(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	room := w.Rooms[3001]
	mob := makeNPC("guard")
	handler.CharToRoom(mob, room)
	w.AddChar(mob)

	obj := &types.ObjData{Name: "sword", ShortDescr: "a sword"}
	obj.InRoom = room
	room.Contents = append(room.Contents, obj)
	w.AddObj(obj)

	mpPurge(mob, " sword")

	// Object should have been extracted
	for _, o := range w.Objects {
		if o == obj {
			t.Error("object should have been removed from world after purge")
		}
	}
	for _, o := range room.Contents {
		if o == obj {
			t.Error("object should have been removed from room after purge")
		}
	}
}

func TestMpPurge_EmptyArgs_PurgesAll(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	room := w.Rooms[3001]
	mob := makeNPC("guard")
	handler.CharToRoom(mob, room)
	w.AddChar(mob)

	other := makeNPC("other")
	handler.CharToRoom(other, room)
	w.AddChar(other)

	obj := &types.ObjData{Name: "sword", ShortDescr: "a sword"}
	obj.InRoom = room
	room.Contents = append(room.Contents, obj)
	w.AddObj(obj)

	// Empty arg should purge all NPCs (except self) and objects.
	mpPurge(mob, "")

	// other NPC should have been extracted from world
	for _, ch := range w.Characters {
		if ch == other {
			t.Error("other NPC should have been extracted in purge all")
		}
	}
	// mob should still be in the world
	mobFound := false
	for _, ch := range w.Characters {
		if ch == mob {
			mobFound = true
		}
	}
	if !mobFound {
		t.Error("mob should not extract itself in purge all")
	}

	// Object should have been extracted
	for _, o := range w.Objects {
		if o == obj {
			t.Error("object should have been extracted in purge all")
		}
	}
}

func TestMpPurge_NilRoom_EmptyArgs(t *testing.T) {
	mob := makeNPC("guard")
	mob.InRoom = nil
	// Must not panic — nil room with empty args should be a no-op
	mpPurge(mob, "")
}

func TestMpPurge_NilRoom_WithArgs(t *testing.T) {
	mob := makeNPC("guard")
	mob.InRoom = nil
	// Must not panic — nil room with target name should be a no-op
	mpPurge(mob, " target")
}

func TestMpPurge_NilWorld(t *testing.T) {
	oldWorld := WorldRef
	WorldRef = nil
	defer func() { WorldRef = oldWorld }()

	room := &types.RoomIndexData{Vnum: 3001, Name: "Test Room"}
	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	other := makeNPC("other")
	other.InRoom = room
	room.People = append(room.People, other)

	// Must not panic — nil WorldRef should be a no-op even with valid room
	mpPurge(mob, "")

	// other should still be in the room (nothing was purged)
	found := false
	for _, ch := range room.People {
		if ch == other {
			found = true
		}
	}
	if !found {
		t.Error("other NPC should still be in room when WorldRef is nil")
	}
}

func TestExecuteCommand_NilMob(t *testing.T) {
	// Should not panic
	executeCommand(nil, "mpecho test")
}

func TestExecuteCommand_EmptyLine(t *testing.T) {
	mob := makeNPC("guard")
	// Should not panic
	executeCommand(mob, "")
}

func TestExecuteCommand_NilRegistry(t *testing.T) {
	oldReg := CmdRegistry
	CmdRegistry = nil
	defer func() { CmdRegistry = oldReg }()

	mob := makeNPC("guard")
	mob.InRoom = &types.RoomIndexData{Vnum: 3001}
	// Should not panic when registry is nil and command doesn't match mp-commands
	executeCommand(mob, "say hello")
}

// -- Tier 3 G6 tests ----------------------------------------------------

func makePlayerInRoom(name string, room *types.RoomIndexData) (*types.CharData, net.Conn) {
	ch, client := makeDescChar(name)
	ch.PCData = &types.PCData{}
	ch.InRoom = room
	room.People = append(room.People, ch)
	return ch, client
}

func TestMpAsound(t *testing.T) {
	src := &types.RoomIndexData{Vnum: 3001}
	dst := &types.RoomIndexData{Vnum: 3002}
	src.Exits = append(src.Exits, &types.ExitData{Direction: 0, ToRoom: dst})
	mob := makeNPC("guard")
	mob.InRoom = src

	player, client := makeDescChar("P")
	defer client.Close()
	player.InRoom = dst
	dst.People = append(dst.People, player)

	mpAsound(mob, " A distant roar!")
	out := flushAndRead(player, client)
	if !strings.Contains(out, "distant roar") {
		t.Errorf("expected roar in adjacent room, got %q", out)
	}
}

func TestMpEchoZone(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	area := &types.AreaData{Name: "Z"}
	w.Rooms[3001].Area = area
	w.Rooms[3002].Area = area

	mob := makeNPC("guard")
	handler.CharToRoom(mob, w.Rooms[3001])
	w.AddChar(mob)

	p1, c1 := makeDescChar("P1")
	defer c1.Close()
	handler.CharToRoom(p1, w.Rooms[3002])
	w.AddChar(p1)

	mpEchoZone(mob, " The sky darkens!")
	out := flushAndRead(p1, c1)
	if !strings.Contains(out, "sky darkens") {
		t.Errorf("expected zone echo, got %q", out)
	}
}

func TestMpMload(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	idx := &types.MobIndexData{Vnum: 999, PlayerName: "orc", ShortDescr: "an orc", Level: 5}
	w.MobIndex[999] = idx

	mob := makeNPC("guard")
	handler.CharToRoom(mob, w.Rooms[3001])

	mpMload(mob, " 999")

	found := false
	for _, ch := range w.Rooms[3001].People {
		if ch.IndexData == idx {
			found = true
		}
	}
	if !found {
		t.Error("mpmload should have placed a new mob in the room")
	}
}

func TestMpMload_BadVnum(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	mob := makeNPC("guard")
	handler.CharToRoom(mob, w.Rooms[3001])
	mpMload(mob, " 99999")
	mpMload(mob, " abc")
	mpMload(mob, "")
}

func TestMpOload_Room(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	idx := &types.ObjIndexData{Vnum: 100, Name: "sword", ShortDescr: "a sword"}
	w.ObjIndex[100] = idx

	mob := makeNPC("guard")
	mob.Level = 50
	handler.CharToRoom(mob, w.Rooms[3001])
	mpOload(mob, " 100")

	if len(w.Rooms[3001].Contents) == 0 {
		t.Error("mpoload with no ITEM_TAKE should drop obj in room")
	}
}

func TestMpOload_Char(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	idx := &types.ObjIndexData{Vnum: 100, Name: "sword", ShortDescr: "a sword", WearFlags: int(types.ITEM_TAKE)}
	w.ObjIndex[100] = idx

	mob := makeNPC("guard")
	mob.Level = 50
	handler.CharToRoom(mob, w.Rooms[3001])
	mpOload(mob, " 100")
	if len(mob.Carrying) == 0 {
		t.Error("mpoload with ITEM_TAKE should give obj to mob")
	}
}

func TestMpInvis_Toggle(t *testing.T) {
	mob := makeNPC("guard")
	mob.Level = 55
	mpInvis(mob, "")
	if !mob.Act.IsSet(types.ACT_MOBINVIS) {
		t.Error("mpinvis should set ACT_MOBINVIS")
	}
	mpInvis(mob, "")
	if mob.Act.IsSet(types.ACT_MOBINVIS) {
		t.Error("mpinvis second call should clear ACT_MOBINVIS")
	}
}

func TestMpInvis_SetLevel(t *testing.T) {
	mob := makeNPC("guard")
	mpInvis(mob, " 25")
	if mob.MobInvis != 25 {
		t.Errorf("mpinvis level should be 25, got %d", mob.MobInvis)
	}
}

func TestMpAt(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()
	oldReg := CmdRegistry
	defer func() { CmdRegistry = oldReg }()

	var seenRoom int
	reg := command.NewRegistry()
	reg.Register(&command.Command{
		Name:  "ping",
		Level: 1,
		DoFun: func(ch *types.CharData, _ string) {
			if ch.InRoom != nil {
				seenRoom = ch.InRoom.Vnum
			}
		},
	})
	CmdRegistry = reg

	mob := makeNPC("guard")
	handler.CharToRoom(mob, w.Rooms[3001])
	mpAt(mob, " 3002 ping")

	if seenRoom != 3002 {
		t.Errorf("mpat should have run ping at 3002, saw %d", seenRoom)
	}
	if mob.InRoom.Vnum != 3001 {
		t.Errorf("mpat should restore mob to 3001, now at %d", mob.InRoom.Vnum)
	}
}

func TestMpAdvance(t *testing.T) {
	// C do_mpadvance takes only <victim> and ALWAYS advances by exactly one
	// level, never beyond LEVEL_AVATAR (mud_comm.c:1388).
	room := &types.RoomIndexData{Vnum: 3001}
	mob := makeNPC("guard")
	mob.Level = 60
	mob.InRoom = room
	room.People = append(room.People, mob)

	victim, client := makePlayerInRoom("Victim", room)
	defer client.Close()
	victim.Level = 10

	mpAdvance(mob, " Victim")
	if victim.Level != 11 {
		t.Errorf("mpadvance should raise level by 1 to 11, got %d", victim.Level)
	}
}

func TestMpAdvance_NoOpAtAvatar(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	mob := makeNPC("guard")
	mob.Level = 65
	mob.InRoom = room
	room.People = append(room.People, mob)

	victim, client := makePlayerInRoom("V", room)
	defer client.Close()
	victim.Level = types.LEVEL_AVATAR

	mpAdvance(mob, " V")
	if victim.Level != types.LEVEL_AVATAR {
		t.Errorf("mpadvance must not raise victim past LEVEL_AVATAR, got %d", victim.Level)
	}
}

func TestMpAdvance_NoOpOnNPC(t *testing.T) {
	// C mud_comm.c:1416 explicitly refuses to advance NPCs.
	room := &types.RoomIndexData{Vnum: 3001}
	mob := makeNPC("guard")
	mob.Level = 60
	mob.InRoom = room
	room.People = append(room.People, mob)

	other := makeNPC("target")
	other.Level = 10
	other.InRoom = room
	room.People = append(room.People, other)

	mpAdvance(mob, " target")
	if other.Level != 10 {
		t.Errorf("mpadvance should not advance NPC, got Level=%d", other.Level)
	}
}

func TestMpAdvance_NoOpAtSupremeCap(t *testing.T) {
	// Sanity check: if a PC somehow sits at LEVEL_SUPREME, don't overflow.
	// Note the LEVEL_AVATAR gate already guards this in practice; this test
	// guards the additional defensive cap in mpAdvance.
	room := &types.RoomIndexData{Vnum: 3001}
	mob := makeNPC("guard")
	mob.Level = types.LEVEL_SUPREME
	mob.InRoom = room
	room.People = append(room.People, mob)

	victim, client := makePlayerInRoom("V", room)
	defer client.Close()
	victim.Level = types.LEVEL_SUPREME

	mpAdvance(mob, " V")
	if victim.Level != types.LEVEL_SUPREME {
		t.Errorf("mpadvance must not raise past LEVEL_SUPREME, got %d", victim.Level)
	}
}

func TestMpSlay(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	room := w.Rooms[3001]
	mob := makeNPC("guard")
	handler.CharToRoom(mob, room)
	w.AddChar(mob)

	victim := makeNPC("foe")
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	mpSlay(mob, " foe")
	for _, c := range w.Characters {
		if c == victim {
			t.Error("victim should be extracted after mpslay")
		}
	}
}

func TestMpSlay_NotImmortal(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	room := w.Rooms[3001]
	mob := makeNPC("guard")
	handler.CharToRoom(mob, room)
	w.AddChar(mob)

	victim, client := makePlayerInRoom("Imm", room)
	defer client.Close()
	victim.Level = types.LEVEL_IMMORTAL
	w.AddChar(victim)

	mpSlay(mob, " Imm")
	// still present
	found := false
	for _, c := range w.Characters {
		if c == victim {
			found = true
		}
	}
	if !found {
		t.Error("immortal should not be slain")
	}
}

func TestMpSlay_NoCorpseNoXP(t *testing.T) {
	// Documenting round-2 contract: Go's mpSlay uses ExtractChar(fPull=true),
	// which removes the victim WITHOUT creating a corpse and WITHOUT awarding
	// XP. Both omissions are intentional per mpSlay's doc comment.
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	room := w.Rooms[3001]
	mob := makeNPC("guard")
	mob.Exp = 1234
	handler.CharToRoom(mob, room)
	w.AddChar(mob)

	victim := makeNPC("foe")
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	prevContents := len(room.Contents)

	mpSlay(mob, " foe")

	if len(room.Contents) != prevContents {
		t.Errorf("mpslay must not add a corpse to the room; contents=%d, want %d",
			len(room.Contents), prevContents)
	}
	if mob.Exp != 1234 {
		t.Errorf("mpslay must not award XP to caller mob; Exp=%d, want 1234", mob.Exp)
	}
}

func TestMpSlay_SupermobGuard(t *testing.T) {
	// C mud_comm.c:2401: cannot slay supermob (vnum 3).
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	room := w.Rooms[3001]
	mob := makeNPC("guard")
	handler.CharToRoom(mob, room)
	w.AddChar(mob)

	supermob := makeNPC("supermob")
	supermob.IndexData = &types.MobIndexData{Vnum: 3}
	handler.CharToRoom(supermob, room)
	w.AddChar(supermob)

	mpSlay(mob, " supermob")

	found := false
	for _, c := range w.Characters {
		if c == supermob {
			found = true
		}
	}
	if !found {
		t.Error("supermob (vnum 3) must not be slain")
	}
}

func TestMpWithdraw_InsufficientFunds(t *testing.T) {
	// mpwithdraw asking for more than the area has must be a no-op.
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	area := &types.AreaData{Name: "A"}
	area.HighEconomy = 0
	area.LowEconomy = 500
	w.Rooms[3001].Area = area

	mob := makeNPC("guard")
	mob.Gold = 0
	handler.CharToRoom(mob, w.Rooms[3001])

	mpWithdraw(mob, " 1000000000")

	if mob.Gold != 0 {
		t.Errorf("mob should not receive gold on insufficient funds, got %d", mob.Gold)
	}
	if area.HighEconomy != 0 || area.LowEconomy != 500 {
		t.Errorf("area economy must be untouched: High=%d Low=%d",
			area.HighEconomy, area.LowEconomy)
	}
}

func TestMpLog(t *testing.T) {
	mob := makeNPC("guard")
	// Should not panic
	mpLog(mob, " something happened")
	mpLog(mob, "")
}

func TestMpRestore(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	victim, client := makePlayerInRoom("V", room)
	defer client.Close()
	victim.Hit, victim.MaxHit = 10, 100
	victim.Mana, victim.MaxMana = 0, 50
	victim.Move, victim.MaxMove = 1, 80

	mpRestore(mob, " V")
	if victim.Hit != 100 || victim.Mana != 50 || victim.Move != 80 {
		t.Errorf("mprestore should max all vitals, got %d/%d/%d", victim.Hit, victim.Mana, victim.Move)
	}
}

func TestMpFavor(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	victim, client := makePlayerInRoom("V", room)
	defer client.Close()
	victim.PCData.Favor = 100

	mpFavor(mob, " V +50")
	if victim.PCData.Favor != 150 {
		t.Errorf("mpfavor +50 failed: got %d", victim.PCData.Favor)
	}
	mpFavor(mob, " V -200")
	if victim.PCData.Favor != -50 {
		t.Errorf("mpfavor -200 failed: got %d", victim.PCData.Favor)
	}
	mpFavor(mob, " V 9999")
	if victim.PCData.Favor != 2500 {
		t.Errorf("mpfavor absolute 9999 should clamp to 2500, got %d", victim.PCData.Favor)
	}
}

func TestMpNuisance_AndUnnuisance(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	victim, client := makePlayerInRoom("V", room)
	defer client.Close()

	mpNuisance(mob, " V")
	if victim.PCData.Nuisance == nil {
		t.Fatal("mpnuisance should add nuisance struct")
	}
	mpUnnuisance(mob, " V")
	if victim.PCData.Nuisance != nil {
		t.Error("mpunnuisance should clear nuisance")
	}
}

func TestMpBodybag(t *testing.T) {
	// C do_mpbodybag iterates the WORLD object list, matching obj->in_room
	// short_descr "the corpse of <name>" with obj->pIndexData->vnum == 11
	// (mud_comm.c:1903).
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	mobRoom := w.Rooms[3001]
	farRoom := w.Rooms[3002]

	mob := makeNPC("guard")
	handler.CharToRoom(mob, mobRoom)

	// PC corpse vnum is 11 in C.
	pcCorpseIdx := &types.ObjIndexData{Vnum: 11, Name: "corpse"}
	w.ObjIndex[11] = pcCorpseIdx

	corpse := &types.ObjData{
		Name: "corpse", ShortDescr: "the corpse of Bob",
		IndexData: pcCorpseIdx,
	}
	corpse.InRoom = farRoom
	farRoom.Contents = append(farRoom.Contents, corpse)
	w.Objects = append(w.Objects, corpse)

	mpBodybag(mob, " Bob")

	if len(farRoom.Contents) != 0 {
		t.Errorf("corpse should be removed from far room, %d remain", len(farRoom.Contents))
	}
	if len(mob.Carrying) != 1 {
		t.Errorf("corpse should be on mob, got %d carried", len(mob.Carrying))
	}
	if corpse.Timer != -1 {
		t.Errorf("corpse timer should be -1, got %d", corpse.Timer)
	}
}

func TestMpBodybag_IgnoresWrongVnum(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	room := w.Rooms[3001]
	mob := makeNPC("guard")
	handler.CharToRoom(mob, room)

	// Vnum != 11 — must not be picked up.
	mobCorpseIdx := &types.ObjIndexData{Vnum: 10, Name: "corpse"}
	w.ObjIndex[10] = mobCorpseIdx
	corpse := &types.ObjData{
		Name: "corpse", ShortDescr: "the corpse of Bob",
		IndexData: mobCorpseIdx,
	}
	corpse.InRoom = room
	room.Contents = append(room.Contents, corpse)
	w.Objects = append(w.Objects, corpse)

	mpBodybag(mob, " Bob")
	if len(room.Contents) != 1 {
		t.Errorf("non-vnum-11 corpse should NOT be moved, %d in room", len(room.Contents))
	}
}

func TestMpMorphUnmorph_Stubs(t *testing.T) {
	mob := makeNPC("guard")
	mpMorph(mob, " V 10")
	mpUnmorph(mob, " V")
}

func TestMpPractice(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	sk := &types.SkillType{Name: "fireball"}
	sk.SkillAdept[0] = 85
	w.Skills = append(w.Skills, sk)

	room := w.Rooms[3001]
	mob := makeNPC("guard")
	handler.CharToRoom(mob, room)

	victim, client := makePlayerInRoom("V", room)
	defer client.Close()
	victim.Class = 0

	mpPractice(mob, " V fireball")
	if victim.PCData.Learned[0] != 85 {
		t.Errorf("mppractice should set learned to 85, got %d", victim.PCData.Learned[0])
	}
}

func TestMpOpenPassage_ClosePassage_FillIn(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	mob := makeNPC("guard")
	handler.CharToRoom(mob, w.Rooms[3001])

	mpOpenPassage(mob, " 0 3002")
	ex := w.Rooms[3001].GetExit(0)
	if ex == nil || ex.ToRoom != w.Rooms[3002] {
		t.Fatal("mpopenpassage should create exit")
	}
	if uint32(ex.ExitInfo)&types.EX_PASSAGE == 0 {
		t.Error("exit should have EX_PASSAGE flag")
	}

	mpClosePassage(mob, " 0")
	if w.Rooms[3001].GetExit(0) != nil {
		t.Error("mpclosepassage should remove passage exit")
	}

	// fillin adds EX_CLOSED to existing exit
	door := &types.ExitData{Direction: 1}
	w.Rooms[3001].Exits = append(w.Rooms[3001].Exits, door)
	mpFillIn(mob, " 1")
	if uint32(door.ExitInfo)&types.EX_CLOSED == 0 {
		t.Error("mpfillin should set EX_CLOSED")
	}
}

func TestMpPeace(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	a := &types.CharData{Name: "A"}
	b := &types.CharData{Name: "B"}
	a.InRoom = room
	b.InRoom = room
	room.People = append(room.People, a, b)
	a.Fighting = &types.FightData{Who: b}
	b.Fighting = &types.FightData{Who: a}
	a.Hating = &types.HHFData{Who: b}

	mpPeace(mob, "")
	if a.Fighting != nil || b.Fighting != nil {
		t.Error("mppeace should stop fighting")
	}
	if a.Hating != nil {
		t.Error("mppeace should clear hating")
	}
}

func TestMpPkset(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	v, client := makePlayerInRoom("V", room)
	defer client.Close()

	mpPkset(mob, " V yes")
	if v.PCData.Flags&int(types.PCFLAG_DEADLY) == 0 {
		t.Error("mppkset yes should set PCFLAG_DEADLY")
	}
	mpPkset(mob, " V no")
	if v.PCData.Flags&int(types.PCFLAG_DEADLY) != 0 {
		t.Error("mppkset no should clear PCFLAG_DEADLY")
	}
}

func TestMpOowner(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	v, client := makePlayerInRoom("Alice", room)
	defer client.Close()
	_ = v

	obj := &types.ObjData{Name: "sword", ShortDescr: "a sword"}
	obj.InRoom = room
	room.Contents = append(room.Contents, obj)

	mpOowner(mob, " sword Alice")
	if obj.Owner != "Alice" {
		t.Errorf("mpoowner should set owner to Alice, got %q", obj.Owner)
	}
	mpOowner(mob, " sword none")
	if obj.Owner != "" {
		t.Errorf("mpoowner none should clear owner, got %q", obj.Owner)
	}
}

func TestMpHunt(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	room := w.Rooms[3001]
	mob := makeNPC("guard")
	handler.CharToRoom(mob, room)
	w.AddChar(mob)

	prey := &types.CharData{Name: "prey"}
	handler.CharToRoom(prey, room)
	w.AddChar(prey)

	mpHunt(mob, " prey")
	if mob.Hunting == nil || mob.Hunting.Who != prey {
		t.Error("mphunt should set Hunting to prey")
	}
}

func TestMpHate(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	room := w.Rooms[3001]
	mob := makeNPC("guard")
	handler.CharToRoom(mob, room)
	w.AddChar(mob)

	foe := &types.CharData{Name: "foe"}
	handler.CharToRoom(foe, room)
	w.AddChar(foe)

	mpHate(mob, " foe")
	if mob.Hating == nil || mob.Hating.Who != foe {
		t.Error("mphate should set Hating to foe")
	}
}

func TestMpDepositAndWithdraw(t *testing.T) {
	// C boost_economy/lower_economy split gold across two fields:
	// LowEconomy holds the sub-billion remainder; HighEconomy counts billions
	// (handler.c:5641). Sub-billion deposits stay in LowEconomy.
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	area := &types.AreaData{Name: "A"}
	w.Rooms[3001].Area = area

	mob := makeNPC("guard")
	mob.Gold = 1000
	handler.CharToRoom(mob, w.Rooms[3001])

	mpDeposit(mob, " 400")
	if mob.Gold != 600 {
		t.Errorf("mpdeposit should decrement gold to 600, got %d", mob.Gold)
	}
	if area.LowEconomy != 400 || area.HighEconomy != 0 {
		t.Errorf("after 400 gold deposit: Low=%d High=%d, want Low=400 High=0",
			area.LowEconomy, area.HighEconomy)
	}
	mpWithdraw(mob, " 200")
	if mob.Gold != 800 {
		t.Errorf("mpwithdraw should return 200 gold, now %d", mob.Gold)
	}
	if area.LowEconomy != 200 || area.HighEconomy != 0 {
		t.Errorf("after 200 gold withdraw: Low=%d High=%d, want Low=200 High=0",
			area.LowEconomy, area.HighEconomy)
	}
}

func TestMpDeposit_OverflowToHighBucket(t *testing.T) {
	// When LowEconomy crosses 1e9, HighEconomy should increment.
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	area := &types.AreaData{Name: "A"}
	area.LowEconomy = 900000000
	w.Rooms[3001].Area = area

	mob := makeNPC("guard")
	mob.Gold = 200000000
	handler.CharToRoom(mob, w.Rooms[3001])

	mpDeposit(mob, " 200000000")
	if area.HighEconomy != 1 {
		t.Errorf("HighEconomy should be 1 after overflow, got %d", area.HighEconomy)
	}
	if area.LowEconomy != 100000000 {
		t.Errorf("LowEconomy should wrap to 100000000, got %d", area.LowEconomy)
	}
}

func TestMpWithdraw_BorrowFromHighBucket(t *testing.T) {
	// Withdraw beyond LowEconomy should borrow from HighEconomy.
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	area := &types.AreaData{Name: "A"}
	area.HighEconomy = 1
	area.LowEconomy = 100000000
	w.Rooms[3001].Area = area

	mob := makeNPC("guard")
	handler.CharToRoom(mob, w.Rooms[3001])

	mpWithdraw(mob, " 200000000")
	if mob.Gold != 200000000 {
		t.Errorf("mob should receive 200000000 gold, got %d", mob.Gold)
	}
	if area.HighEconomy != 0 {
		t.Errorf("HighEconomy should be 0 after borrow, got %d", area.HighEconomy)
	}
	if area.LowEconomy != 900000000 {
		t.Errorf("LowEconomy should be 900000000 after borrow, got %d", area.LowEconomy)
	}
}

func TestMpApplyAndApplyB_Stubs(t *testing.T) {
	mob := makeNPC("guard")
	mpApply(mob, " something")
	mpApplyB(mob, " something")
}

func TestMpDelay(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	v, client := makePlayerInRoom("V", room)
	defer client.Close()
	v.Level = 10

	mpDelay(mob, " V 3")
	if v.Wait != 3*types.PULSE_VIOLENCE {
		t.Errorf("mpdelay should set Wait, got %d", v.Wait)
	}
}

func TestMpStrew(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	area := &types.AreaData{Name: "A"}
	w.Rooms[3001].Area = area
	w.Rooms[3002].Area = area
	w.Rooms[3003] = &types.RoomIndexData{Vnum: 3003, Area: area}

	idx := &types.ObjIndexData{Vnum: 111, Name: "coin", ShortDescr: "a coin"}
	w.ObjIndex[111] = idx

	mob := makeNPC("guard")
	handler.CharToRoom(mob, w.Rooms[3001])

	mpStrew(mob, " 111 3")
	totalObjs := len(w.Rooms[3001].Contents) + len(w.Rooms[3002].Contents) + len(w.Rooms[3003].Contents)
	if totalObjs != 3 {
		t.Errorf("mpstrew should create 3 copies, got %d", totalObjs)
	}
}

func TestMpScatter(t *testing.T) {
	// C do_mpscatter takes <victim> <low_vnum> <high_vnum> and teleports the
	// victim CHARACTER to a random room in that vnum range, setting position
	// to POS_RESTING (mud_comm.c:2283).
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	source := w.Rooms[3001]
	dest := w.Rooms[3002]

	mob := makeNPC("guard")
	handler.CharToRoom(mob, source)

	victim, client := makePlayerInRoom("Bob", source)
	defer client.Close()
	victim.Position = types.POS_FIGHTING

	mpScatter(mob, " Bob 3002 3002")

	if victim.InRoom != dest {
		t.Errorf("victim should be in room 3002, in %v", victim.InRoom)
	}
	if victim.Position != types.POS_RESTING {
		t.Errorf("victim position should be POS_RESTING after scatter, got %d", victim.Position)
	}
}

func TestMpScatter_BadRange(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	mob := makeNPC("guard")
	handler.CharToRoom(mob, w.Rooms[3001])
	victim, client := makePlayerInRoom("Bob", w.Rooms[3001])
	defer client.Close()

	// Range with no rooms — victim should not move.
	mpScatter(mob, " Bob 99000 99001")
	if victim.InRoom != w.Rooms[3001] {
		t.Errorf("victim should stay when no rooms in range")
	}
}

func TestMpDream_Sleeping(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	room := w.Rooms[3001]
	mob := makeNPC("guard")
	handler.CharToRoom(mob, room)
	w.AddChar(mob)

	v, client := makePlayerInRoom("V", room)
	defer client.Close()
	v.Position = types.POS_SLEEPING
	w.AddChar(v)

	mpDream(mob, " V You hear bells chiming.")
	out := flushAndRead(v, client)
	if !strings.Contains(out, "bells chiming") {
		t.Errorf("mpdream should deliver to sleeper, got %q", out)
	}
}

func TestMpDream_Awake_Silent(t *testing.T) {
	w := setupTestWorld()
	oldWorld := WorldRef
	WorldRef = w
	defer func() { WorldRef = oldWorld }()

	room := w.Rooms[3001]
	mob := makeNPC("guard")
	handler.CharToRoom(mob, room)
	w.AddChar(mob)

	v, client := makePlayerInRoom("V", room)
	defer client.Close()
	v.Position = types.POS_STANDING
	w.AddChar(v)

	mpDream(mob, " V nothing")
	out := flushAndRead(v, client)
	if strings.Contains(out, "nothing") {
		t.Errorf("mpdream should not deliver to awake, got %q", out)
	}
}

func TestMpNothing(t *testing.T) {
	mob := makeNPC("guard")
	mpNothing(mob, "whatever")
}

func TestMpApplyAffect(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	v, client := makePlayerInRoom("V", room)
	defer client.Close()

	mpApplyAffect(mob, " V 1 2 3 4")
	if len(v.Affects) != 1 {
		t.Fatalf("mpapplyaffect should add an affect, got %d", len(v.Affects))
	}
	a := v.Affects[0]
	if a.Type != 1 || a.Location != 2 || a.Modifier != 3 || a.Duration != 4 {
		t.Errorf("affect fields incorrect: %+v", a)
	}
}
