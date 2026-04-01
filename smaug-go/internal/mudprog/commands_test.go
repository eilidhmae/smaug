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
		Connected:  int(types.CON_PLAYING),
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
