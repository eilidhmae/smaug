package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

// --- DoKill ---

func TestDoKill_NoArg(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Fighter")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 9000, Name: "Arena"}

	DoKill(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Kill whom?") {
		t.Errorf("expected 'Kill whom?', got: %q", out)
	}
}

func TestDoKill_NotFound(t *testing.T) {
	_ = setupWizWorld()
	room := &types.RoomIndexData{Vnum: 9001, Name: "Arena"}
	ch, client := makeTestChar("Fighter")
	defer client.Close()
	handler.CharToRoom(ch, room)

	DoKill(ch, "dragon")
	out := readOutput(ch, client)
	if !strings.Contains(out, "aren't here") {
		t.Errorf("expected 'aren't here', got: %q", out)
	}
}

func TestDoKill_Self(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 9002, Name: "Arena"}
	w.Rooms[9002] = room

	ch, client := makeTestChar("Fighter")
	defer client.Close()
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	DoKill(ch, "Fighter")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Ouch") {
		t.Errorf("expected self-attack message, got: %q", out)
	}
}

func TestDoKill_Player(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 9003, Name: "Arena"}
	w.Rooms[9003] = room

	ch, client := makeTestChar("Fighter")
	defer client.Close()
	handler.CharToRoom(ch, room)

	victim, victimClient := makeTestChar("Target")
	defer victimClient.Close()
	handler.CharToRoom(victim, room)

	DoKill(ch, "Target")
	out := readOutput(ch, client)
	if !strings.Contains(out, "MURDER") {
		t.Errorf("expected 'must MURDER' message, got: %q", out)
	}
}

func TestDoKill_NPC(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 9004, Name: "Arena"}
	w.Rooms[9004] = room

	ch, client := makeTestChar("Fighter")
	defer client.Close()
	handler.CharToRoom(ch, room)

	npc := &types.CharData{
		Name:       "rat sewer",
		ShortDescr: "a sewer rat",
		Level:      2,
		Hit:        20, MaxHit: 20,
		Position: types.POS_STANDING,
		InRoom:   room,
	}
	npc.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(npc, room)

	DoKill(ch, "rat")
	out := readOutput(ch, client)

	if ch.Fighting == nil {
		t.Error("should be fighting")
	}
	if !strings.Contains(out, "sewer rat") {
		t.Errorf("expected attack message, got: %q", out)
	}

	// Cleanup
	combat.StopFighting(ch, true)
	if npc.Fighting != nil {
		combat.StopFighting(npc, true)
	}
}

func TestDoKill_AlreadyFighting(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 9005, Name: "Arena"}
	w.Rooms[9005] = room

	ch, client := makeTestChar("Fighter")
	defer client.Close()
	ch.Position = types.POS_STANDING
	handler.CharToRoom(ch, room)

	npc1 := &types.CharData{Name: "rat", Level: 2, Hit: 20, MaxHit: 20, Position: types.POS_STANDING}
	npc1.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(npc1, room)

	npc2 := &types.CharData{Name: "bat", Level: 2, Hit: 20, MaxHit: 20, Position: types.POS_STANDING}
	npc2.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(npc2, room)

	combat.StartFighting(ch, npc1)
	ch.Position = types.POS_STANDING // Keep standing to reach the Fighting check

	DoKill(ch, "bat")
	out := readOutput(ch, client)
	if !strings.Contains(out, "best you can") {
		t.Errorf("expected already fighting message, got: %q", out)
	}

	combat.StopFighting(ch, true)
}

func TestDoKill_BadCondition(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 9006, Name: "Arena"}
	w.Rooms[9006] = room

	ch, client := makeTestChar("Fighter")
	defer client.Close()
	ch.Hit = 0
	handler.CharToRoom(ch, room)

	npc := &types.CharData{Name: "rat", Level: 2, Position: types.POS_STANDING}
	npc.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(npc, room)

	DoKill(ch, "rat")
	out := readOutput(ch, client)
	if !strings.Contains(out, "no condition") {
		t.Errorf("expected 'no condition' message, got: %q", out)
	}
}

// --- DoMurder ---

func TestDoMurder_NoArg(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Fighter")
	defer client.Close()

	DoMurder(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Murder whom?") {
		t.Errorf("expected 'Murder whom?', got: %q", out)
	}
}

func TestDoMurder_Self(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 9010, Name: "Arena"}
	w.Rooms[9010] = room

	ch, client := makeTestChar("Fighter")
	defer client.Close()
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	DoMurder(ch, "Fighter")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Suicide") {
		t.Errorf("expected suicide message, got: %q", out)
	}
}

func TestDoMurder_SafeRoom(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 9011, Name: "Temple"}
	room.RoomFlags.Set(types.ROOM_SAFE)
	w.Rooms[9011] = room

	ch, client := makeTestChar("Fighter")
	defer client.Close()
	handler.CharToRoom(ch, room)

	victim, victimClient := makeTestChar("Target")
	defer victimClient.Close()
	handler.CharToRoom(victim, room)

	DoMurder(ch, "Target")
	out := readOutput(ch, client)
	if !strings.Contains(out, "cannot fight here") {
		t.Errorf("expected safe room message, got: %q", out)
	}
}

func TestDoMurder_NotFound(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 9012, Name: "Arena"}
	w.Rooms[9012] = room

	ch, client := makeTestChar("Fighter")
	defer client.Close()
	handler.CharToRoom(ch, room)

	DoMurder(ch, "nonexistent")
	out := readOutput(ch, client)
	if !strings.Contains(out, "aren't here") {
		t.Errorf("expected 'aren't here', got: %q", out)
	}
}

func TestDoMurder_Success(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 9013, Name: "Arena"}
	w.Rooms[9013] = room

	ch, client := makeTestChar("Fighter")
	defer client.Close()
	handler.CharToRoom(ch, room)

	victim, victimClient := makeTestChar("Target")
	defer victimClient.Close()
	handler.CharToRoom(victim, room)

	DoMurder(ch, "Target")
	out := readOutput(ch, client)

	if ch.Fighting == nil {
		t.Error("should be fighting")
	}
	if !strings.Contains(out, "attack") {
		t.Errorf("expected attack message, got: %q", out)
	}

	combat.StopFighting(ch, true)
	if victim.Fighting != nil {
		combat.StopFighting(victim, true)
	}
}

func TestDoMurder_AlreadyFighting(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 9014, Name: "Arena"}
	w.Rooms[9014] = room

	ch, client := makeTestChar("Fighter")
	defer client.Close()
	handler.CharToRoom(ch, room)

	npc := &types.CharData{Name: "rat", Level: 2, Position: types.POS_STANDING}
	npc.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(npc, room)

	victim, victimClient := makeTestChar("Target")
	defer victimClient.Close()
	handler.CharToRoom(victim, room)

	combat.StartFighting(ch, npc)
	ch.Position = types.POS_STANDING

	DoMurder(ch, "Target")
	out := readOutput(ch, client)
	if !strings.Contains(out, "best you can") {
		t.Errorf("expected 'best you can', got: %q", out)
	}

	combat.StopFighting(ch, true)
}

func TestDoMurder_BadCondition(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 9015, Name: "Arena"}
	w.Rooms[9015] = room

	ch, client := makeTestChar("Fighter")
	defer client.Close()
	ch.Hit = 0
	handler.CharToRoom(ch, room)

	victim, victimClient := makeTestChar("Target")
	defer victimClient.Close()
	handler.CharToRoom(victim, room)

	DoMurder(ch, "Target")
	out := readOutput(ch, client)
	if !strings.Contains(out, "no condition") {
		t.Errorf("expected 'no condition', got: %q", out)
	}
}

// --- DoWimpy ---

func TestDoWimpy_Show(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Fighter")
	defer client.Close()
	ch.Wimpy = 25

	DoWimpy(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "25") {
		t.Errorf("expected current wimpy value, got: %q", out)
	}
}

func TestDoWimpy_Set(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Fighter")
	defer client.Close()

	DoWimpy(ch, "30")
	out := readOutput(ch, client)

	if ch.Wimpy != 30 {
		t.Errorf("wimpy should be 30, got: %d", ch.Wimpy)
	}
	if !strings.Contains(out, "30") {
		t.Errorf("expected '30' in output, got: %q", out)
	}
}

func TestDoWimpy_Negative(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Fighter")
	defer client.Close()

	DoWimpy(ch, "-5")
	out := readOutput(ch, client)
	if !strings.Contains(out, "positive") {
		t.Errorf("expected 'positive' message, got: %q", out)
	}
}

func TestDoWimpy_TooHigh(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Fighter")
	defer client.Close()

	DoWimpy(ch, "999")
	out := readOutput(ch, client)
	if !strings.Contains(out, "exceed") {
		t.Errorf("expected 'exceed' message, got: %q", out)
	}
}

// --- DoFlee ---

func TestDoFlee_NotFighting(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Fighter")
	defer client.Close()

	DoFlee(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "aren't fighting") {
		t.Errorf("expected 'aren't fighting', got: %q", out)
	}
}

func TestDoFlee_NoMove(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 9020, Name: "Pit"}
	w.Rooms[9020] = room

	ch, client := makeTestChar("Fighter")
	defer client.Close()
	ch.Move = 0
	handler.CharToRoom(ch, room)

	npc := &types.CharData{Name: "rat", Level: 2, Position: types.POS_STANDING}
	npc.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(npc, room)
	combat.StartFighting(ch, npc)

	DoFlee(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "exhausted") {
		t.Errorf("expected 'exhausted', got: %q", out)
	}

	combat.StopFighting(ch, true)
}
