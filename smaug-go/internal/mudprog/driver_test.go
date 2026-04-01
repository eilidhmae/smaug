package mudprog

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/command"
	"github.com/eilidhmae/smaug/internal/types"
)

// makeTestChar creates a CharData with a net.Pipe-backed descriptor for output capture.
func makeTestChar(name string) (*types.CharData, net.Conn) {
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
		Mana:       50,
		MaxMana:    50,
		Move:       80,
		MaxMove:    80,
		PermStr:    15,
		PermInt:    13,
		PermWis:    12,
		PermDex:    14,
		PermCon:    15,
		PermCha:    11,
		PermLck:    13,
		Gold:       500,
		Desc:       d,
		PCData:     &types.PCData{},
	}
	d.Character = ch
	return ch, client
}

// readOutput flushes the descriptor and reads the output.
func readOutput(ch *types.CharData, client net.Conn) string {
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

func makeNPC(name string) *types.CharData {
	ch := &types.CharData{
		Name:       name,
		ShortDescr: "a " + name,
		Level:      10,
		Position:   types.POS_STANDING,
		Hit:        100,
		MaxHit:     100,
	}
	ch.Act.Set(types.ACT_IS_NPC)
	return ch
}

func TestDriver_NilMob(t *testing.T) {
	// Should not panic with nil mob
	Driver("mpecho hello", nil, nil, nil, nil, nil, false)
}

func TestDriver_EmptyComList(t *testing.T) {
	mob := makeNPC("guard")
	// Should not panic with empty command list
	Driver("", mob, nil, nil, nil, nil, false)
}

func TestDriver_SimpleCommand(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 9000, Name: "Test"}
	mob := makeNPC("TestMob")
	mob.InRoom = room
	room.People = append(room.People, mob)

	actor, client := makeTestChar("Player")
	defer client.Close()
	actor.InRoom = room
	room.People = append(room.People, actor)

	Driver("mpecho Hello from mudprog!", mob, actor, nil, nil, nil, false)

	out := readOutput(actor, client)
	if !strings.Contains(out, "Hello from mudprog!") {
		t.Errorf("expected mpecho output, got %q", out)
	}
}

func TestDriver_IfTrue(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 9001, Name: "Test"}
	mob := makeNPC("TestMob")
	mob.InRoom = room
	room.People = append(room.People, mob)

	actor, client := makeTestChar("Player")
	defer client.Close()
	actor.Level = 20
	actor.InRoom = room
	room.People = append(room.People, actor)

	prog := "if level($n) > 10\nmpecho High level!\nendif"
	Driver(prog, mob, actor, nil, nil, nil, false)

	out := readOutput(actor, client)
	if !strings.Contains(out, "High level!") {
		t.Errorf("expected 'High level!' in output, got %q", out)
	}
}

func TestDriver_IfFalse(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 9001, Name: "Test"}
	mob := makeNPC("TestMob")
	mob.InRoom = room
	room.People = append(room.People, mob)

	actor, client := makeTestChar("Player")
	defer client.Close()
	actor.Level = 5
	actor.InRoom = room
	room.People = append(room.People, actor)

	prog := "if level($n) > 10\nmpecho Should not see this\nendif"
	Driver(prog, mob, actor, nil, nil, nil, false)

	out := readOutput(actor, client)
	if strings.Contains(out, "Should not see this") {
		t.Errorf("command should not have executed when condition is false, got %q", out)
	}
}

func TestDriver_Else_TrueBranch(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 9001, Name: "Test"}
	mob := makeNPC("TestMob")
	mob.InRoom = room
	room.People = append(room.People, mob)

	actor, client := makeTestChar("Player")
	defer client.Close()
	actor.Level = 20
	actor.InRoom = room
	room.People = append(room.People, actor)

	prog := "if level($n) > 10\nmpecho High level!\nelse\nmpecho Low level!\nendif"
	Driver(prog, mob, actor, nil, nil, nil, false)

	out := readOutput(actor, client)
	if !strings.Contains(out, "High level!") {
		t.Errorf("expected 'High level!' in output, got %q", out)
	}
	if strings.Contains(out, "Low level!") {
		t.Errorf("should NOT contain 'Low level!' in output, got %q", out)
	}
}

func TestDriver_Else_FalseBranch(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 9001, Name: "Test"}
	mob := makeNPC("TestMob")
	mob.InRoom = room
	room.People = append(room.People, mob)

	actor, client := makeTestChar("Player")
	defer client.Close()
	actor.Level = 5
	actor.InRoom = room
	room.People = append(room.People, actor)

	prog := "if level($n) > 10\nmpecho High level!\nelse\nmpecho Low level!\nendif"
	Driver(prog, mob, actor, nil, nil, nil, false)

	out := readOutput(actor, client)
	if strings.Contains(out, "High level!") {
		t.Errorf("should NOT contain 'High level!' in output, got %q", out)
	}
	if !strings.Contains(out, "Low level!") {
		t.Errorf("expected 'Low level!' in output, got %q", out)
	}
}

func TestDriver_NestedIf(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 9002, Name: "Test"}
	mob := makeNPC("TestMob")
	mob.InRoom = room
	room.People = append(room.People, mob)

	actor, client := makeTestChar("Player")
	defer client.Close()
	actor.Level = 20
	actor.InRoom = room
	room.People = append(room.People, actor)

	prog := "if ispc($n)\nif level($n) > 15\nmpecho High PC!\nendif\nendif"
	Driver(prog, mob, actor, nil, nil, nil, false)

	out := readOutput(actor, client)
	if !strings.Contains(out, "High PC!") {
		t.Errorf("expected 'High PC!' in output, got %q", out)
	}
}

func TestDriver_NestedIf_OuterFalse(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 9002, Name: "Test"}
	mob := makeNPC("TestMob")
	mob.InRoom = room
	room.People = append(room.People, mob)

	// NPC actor - ispc will fail
	actor := makeNPC("NpcActor")
	actor.InRoom = room
	room.People = append(room.People, actor)

	bystander, client := makeTestChar("Bystander")
	defer client.Close()
	bystander.InRoom = room
	room.People = append(room.People, bystander)

	prog := "if ispc($n)\nif level($n) > 5\nmpecho Should not see\nendif\nendif"
	Driver(prog, mob, actor, nil, nil, nil, false)

	out := readOutput(bystander, client)
	if strings.Contains(out, "Should not see") {
		t.Errorf("nested if should not execute when outer is false, got %q", out)
	}
}

func TestDriver_Or(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 9003, Name: "Test"}
	mob := makeNPC("TestMob")
	mob.InRoom = room
	room.People = append(room.People, mob)

	actor, client := makeTestChar("Player")
	defer client.Close()
	actor.Level = 5
	actor.InRoom = room
	room.People = append(room.People, actor)

	// First condition false (level > 10), OR condition true (ispc)
	prog := "if level($n) > 10\nor ispc($n)\nmpecho Or worked!\nendif"
	Driver(prog, mob, actor, nil, nil, nil, false)

	out := readOutput(actor, client)
	if !strings.Contains(out, "Or worked!") {
		t.Errorf("expected 'Or worked!' after OR branch, got %q", out)
	}
}

func TestDriver_Or_AlreadyTrue(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 9003, Name: "Test"}
	mob := makeNPC("TestMob")
	mob.InRoom = room
	room.People = append(room.People, mob)

	actor, client := makeTestChar("Player")
	defer client.Close()
	actor.Level = 20
	actor.InRoom = room
	room.People = append(room.People, actor)

	// First condition true - OR should not change state
	prog := "if level($n) > 10\nor level($n) > 100\nmpecho Still true!\nendif"
	Driver(prog, mob, actor, nil, nil, nil, false)

	out := readOutput(actor, client)
	if !strings.Contains(out, "Still true!") {
		t.Errorf("expected 'Still true!' when if was already true, got %q", out)
	}
}

func TestDriver_Break(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 9004, Name: "Test"}
	mob := makeNPC("TestMob")
	mob.InRoom = room
	room.People = append(room.People, mob)

	actor, client := makeTestChar("Player")
	defer client.Close()
	actor.InRoom = room
	room.People = append(room.People, actor)

	prog := "mpecho line1\nbreak\nmpecho should not run"
	Driver(prog, mob, actor, nil, nil, nil, false)

	out := readOutput(actor, client)
	if !strings.Contains(out, "line1") {
		t.Errorf("expected 'line1' before break, got %q", out)
	}
	if strings.Contains(out, "should not run") {
		t.Errorf("should not execute after break, got %q", out)
	}
}

func TestDriver_MaxNest(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 9005, Name: "Test"}
	mob := makeNPC("TestMob")
	mob.InRoom = room
	room.People = append(room.People, mob)

	// Reset nest counter
	progNest = 0

	// Simulate deep nesting by manually incrementing progNest
	progNest = maxProgNest
	Driver("mpecho should not run", mob, nil, nil, nil, nil, false)
	// Should bail out and restore counter
	if progNest != maxProgNest {
		t.Errorf("progNest should be %d after bail, got %d", maxProgNest, progNest)
	}
	progNest = 0
}

func TestDriver_MaxIfs(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 9006, Name: "Test"}
	mob := makeNPC("TestMob")
	mob.InRoom = room
	room.People = append(room.People, mob)

	actor := &types.CharData{Name: "Player", Level: 20}

	// Build a prog with MAX_IFS nested ifs to trigger the "too many nested ifs" path
	var lines []string
	for i := 0; i < types.MAX_IFS+1; i++ {
		lines = append(lines, "if ispc($n)")
	}
	lines = append(lines, "mpecho deep")
	for i := 0; i < types.MAX_IFS+1; i++ {
		lines = append(lines, "endif")
	}
	prog := strings.Join(lines, "\n")

	progNest = 0
	Driver(prog, mob, actor, nil, nil, nil, false)
	// Just verify it doesn't crash and progNest is restored
	if progNest != 0 {
		t.Errorf("progNest should be 0 after driver returns, got %d", progNest)
	}
}

func TestDriver_SingleStep(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 9007, Name: "Test"}
	mob := makeNPC("TestMob")
	mob.InRoom = room
	room.People = append(room.People, mob)

	actor, client := makeTestChar("Player")
	defer client.Close()
	actor.InRoom = room
	room.People = append(room.People, actor)

	prog := "mpecho first\nmpecho second"
	Driver(prog, mob, actor, nil, nil, nil, true)

	out := readOutput(actor, client)
	if !strings.Contains(out, "first") {
		t.Errorf("expected 'first' in single step, got %q", out)
	}
	if strings.Contains(out, "second") {
		t.Errorf("should not execute second line in single step, got %q", out)
	}
}

func TestDriver_SkipEmptyAndTilde(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 9008, Name: "Test"}
	mob := makeNPC("TestMob")
	mob.InRoom = room
	room.People = append(room.People, mob)

	actor, client := makeTestChar("Player")
	defer client.Close()
	actor.InRoom = room
	room.People = append(room.People, actor)

	prog := "\n~\nmpecho works\n\n~"
	Driver(prog, mob, actor, nil, nil, nil, false)

	out := readOutput(actor, client)
	if !strings.Contains(out, "works") {
		t.Errorf("expected 'works' after skipping empty/tilde lines, got %q", out)
	}
}

func TestDriver_CmdRegistryFallthrough(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 9009, Name: "Test"}
	mob := makeNPC("TestMob")
	mob.InRoom = room
	room.People = append(room.People, mob)

	// Set up a minimal command registry with a test command
	oldReg := CmdRegistry
	defer func() { CmdRegistry = oldReg }()

	var calledWith string
	reg := command.NewRegistry()
	reg.Register(&command.Command{
		Name:  "testcmd",
		Level: 1,
		DoFun: func(ch *types.CharData, argument string) {
			calledWith = argument
		},
	})
	CmdRegistry = reg

	actor := &types.CharData{Name: "Player", Level: 10, Position: types.POS_STANDING}
	Driver("testcmd hello world", mob, actor, nil, nil, nil, false)

	if calledWith != "hello world" {
		t.Errorf("expected command interpreter to be called with 'hello world', got %q", calledWith)
	}
}

func TestDriver_ElseAfterTrueIfSkips(t *testing.T) {
	// When if is true, the else block should be skipped (ifState == 2)
	room := &types.RoomIndexData{Vnum: 9010, Name: "Test"}
	mob := makeNPC("TestMob")
	mob.InRoom = room
	room.People = append(room.People, mob)

	actor, client := makeTestChar("Player")
	defer client.Close()
	actor.Level = 20
	actor.InRoom = room
	room.People = append(room.People, actor)

	prog := "if level($n) > 10\nmpecho true path\nelse\nmpecho false path\nendif\nmpecho after endif"
	Driver(prog, mob, actor, nil, nil, nil, false)

	out := readOutput(actor, client)
	if !strings.Contains(out, "true path") {
		t.Errorf("expected 'true path', got %q", out)
	}
	if strings.Contains(out, "false path") {
		t.Errorf("should not contain 'false path', got %q", out)
	}
	if !strings.Contains(out, "after endif") {
		t.Errorf("expected 'after endif' to execute, got %q", out)
	}
}

func TestFirstWord(t *testing.T) {
	tests := []struct {
		input string
		word  string
		rest  string
	}{
		{"hello world", "hello", "world"},
		{"single", "single", ""},
		{"  spaced  out  ", "spaced", "out"},
		{"", "", ""},
		{"tab\there", "tab", "here"},
	}
	for _, tc := range tests {
		word, rest := firstWord(tc.input)
		if word != tc.word || rest != tc.rest {
			t.Errorf("firstWord(%q) = (%q, %q), want (%q, %q)", tc.input, word, rest, tc.word, tc.rest)
		}
	}
}
