package mudprog

import (
	"net"
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// makeProgRoom builds a RoomIndexData with a single MudProg of the given
// type/arglist/comlist and returns it.
func makeProgRoom(vnum int, progType int64, argList, comList string) *types.RoomIndexData {
	return &types.RoomIndexData{
		Vnum: vnum,
		Name: "Test Room",
		MudProgs: []*types.MProgData{
			{Type: progType, ArgList: argList, ComList: comList},
		},
	}
}

// watcherIn builds a PC descriptor-backed char, places them in room, and
// returns (char, client) so tests can read captured mpecho output.
func watcherIn(room *types.RoomIndexData, name string) (*types.CharData, net.Conn) {
	ch, client := makeDescChar(name)
	ch.InRoom = room
	room.People = append(room.People, ch)
	return ch, client
}

func TestBuildSupermobForRoom(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001, Name: "A Place"}
	sm := buildSupermobForRoom(room)
	if sm == nil {
		t.Fatal("expected non-nil supermob")
	}
	if !sm.IsNPC() {
		t.Error("supermob must be flagged NPC")
	}
	if sm.InRoom != room {
		t.Error("supermob must be bound to room")
	}
	if sm.Position != types.POS_STANDING {
		t.Error("supermob must be standing")
	}
	for _, p := range room.People {
		if p == sm {
			t.Error("supermob must not be added to room.People")
		}
	}
	if got := buildSupermobForRoom(nil); got != nil {
		t.Error("nil room must yield nil supermob")
	}
}

func TestRprogEnterTrigger_Fires(t *testing.T) {
	room := makeProgRoom(3001, types.MPROG_ENTER, "", "mpecho FIRED")
	ch, client := watcherIn(room, "Walker")
	defer client.Close()

	progNest = 0
	RprogEnterTrigger(ch)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("ENTER: got %q", out)
	}
}

func TestRprogLeaveTrigger_Fires(t *testing.T) {
	room := makeProgRoom(3002, types.MPROG_LEAVE, "", "mpecho FIRED")
	ch, client := watcherIn(room, "Leaver")
	defer client.Close()

	progNest = 0
	RprogLeaveTrigger(ch, room)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("LEAVE: got %q", out)
	}
}

func TestRprogSleepTrigger_Fires(t *testing.T) {
	room := makeProgRoom(3003, types.MPROG_SLEEP, "", "mpecho FIRED")
	ch, client := watcherIn(room, "Sleeper")
	defer client.Close()

	progNest = 0
	RprogSleepTrigger(ch)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("SLEEP: got %q", out)
	}
}

func TestRprogRestTrigger_Fires(t *testing.T) {
	room := makeProgRoom(3004, types.MPROG_REST, "", "mpecho FIRED")
	ch, client := watcherIn(room, "Rester")
	defer client.Close()

	progNest = 0
	RprogRestTrigger(ch)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("REST: got %q", out)
	}
}

func TestRprogRfightTrigger_Fires(t *testing.T) {
	room := makeProgRoom(3005, types.MPROG_RFIGHT, "", "mpecho FIRED")
	ch, client := watcherIn(room, "Fighter")
	defer client.Close()

	progNest = 0
	RprogRfightTrigger(ch)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("RFIGHT: got %q", out)
	}
}

func TestRprogDeathTrigger_Fires(t *testing.T) {
	room := makeProgRoom(3006, types.MPROG_RDEATH, "", "mpecho FIRED")
	ch, client := watcherIn(room, "Corpse")
	defer client.Close()

	progNest = 0
	RprogDeathTrigger(ch)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("RDEATH: got %q", out)
	}
}

func TestRprogImminfoTrigger_Fires(t *testing.T) {
	room := makeProgRoom(3007, types.MPROG_IMMINFO, "", "mpecho FIRED")
	imm, client := watcherIn(room, "Imm")
	defer client.Close()

	progNest = 0
	RprogImminfoTrigger(imm)

	out := readTrigOutput(imm, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("IMMINFO: got %q", out)
	}
}

func TestRprogSpeechTrigger_Fires(t *testing.T) {
	room := makeProgRoom(3008, types.MPROG_SPEECH, "hello", "mpecho FIRED")
	ch, client := watcherIn(room, "Speaker")
	defer client.Close()

	progNest = 0
	RprogSpeechTrigger(ch, "hello there friend")

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("SPEECH: got %q", out)
	}
}

func TestRprogSpeechTrigger_NoMatch(t *testing.T) {
	room := makeProgRoom(3008, types.MPROG_SPEECH, "password", "mpecho FIRED")
	ch, client := watcherIn(room, "Speaker")
	defer client.Close()

	progNest = 0
	RprogSpeechTrigger(ch, "hello there friend")

	out := readTrigOutput(ch, client)
	if strings.Contains(out, "FIRED") {
		t.Errorf("SPEECH should not fire on non-matching keyword, got %q", out)
	}
}

func TestRprogCommandTrigger_Consumes(t *testing.T) {
	room := makeProgRoom(3009, types.MPROG_CMD, "wiggle", "mpecho FIRED")
	ch, client := watcherIn(room, "Commander")
	defer client.Close()

	progNest = 0
	if !RprogCommandTrigger(ch, "wiggle hard") {
		t.Fatal("CMD prog should consume matching command")
	}

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("CMD: got %q", out)
	}
}

// plan-tranche-b.md G4: wordlist match with true word boundaries.
func TestRprogCommandTrigger_SubstringWithoutWordBoundaryDoesNotFire(t *testing.T) {
	// arglist "wigg" should NOT match input "wiggle hard" (no right
	// word boundary after "wigg"). Regression against the previous
	// first-word-exact-match implementation which would have... still
	// returned false since "wigg" != "wiggle" either way. Use a
	// clearer variant: arglist "wig" input "wiggle hard" → no match.
	room := makeProgRoom(3009, types.MPROG_CMD, "wig", "mpecho FIRED")
	ch, client := watcherIn(room, "Commander")
	defer client.Close()

	progNest = 0
	if RprogCommandTrigger(ch, "wiggle hard") {
		t.Fatal("CMD prog should NOT consume substring-only match")
	}
	out := readTrigOutput(ch, client)
	if strings.Contains(out, "FIRED") {
		t.Errorf("substring match should not fire; got %q", out)
	}
}

// plan-tranche-b.md G4: match on non-first word.
func TestRprogCommandTrigger_WordMatchInMiddle(t *testing.T) {
	room := makeProgRoom(3009, types.MPROG_CMD, "hard", "mpecho FIRED")
	ch, client := watcherIn(room, "Commander")
	defer client.Close()

	progNest = 0
	if !RprogCommandTrigger(ch, "wiggle hard") {
		t.Fatal("CMD prog should fire on middle-of-input keyword match")
	}
	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("CMD: got %q", out)
	}
}

// plan-tranche-b.md G4: phrase prefix "p " support.
func TestRprogCommandTrigger_PhrasePrefix(t *testing.T) {
	room := makeProgRoom(3009, types.MPROG_CMD, "p say hello", "mpecho FIRED")
	ch, client := watcherIn(room, "Commander")
	defer client.Close()

	progNest = 0
	if !RprogCommandTrigger(ch, "say hello world") {
		t.Fatal("phrase prefix should match `say hello world`")
	}
	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("CMD: got %q", out)
	}
}

func TestRprogCommandTrigger_DoesNotConsumeUnmatched(t *testing.T) {
	room := makeProgRoom(3009, types.MPROG_CMD, "wiggle", "mpecho FIRED")
	ch, client := watcherIn(room, "Commander")
	defer client.Close()

	progNest = 0
	if RprogCommandTrigger(ch, "dance") {
		t.Fatal("CMD prog should not consume non-matching command")
	}
	out := readTrigOutput(ch, client)
	if strings.Contains(out, "FIRED") {
		t.Errorf("CMD: unexpected fire %q", out)
	}
}

func TestRprogRandomTrigger_Fires(t *testing.T) {
	// arglist "100" → always fire (URANGE(0, 100, 100), percent <= 100)
	room := makeProgRoom(3010, types.MPROG_RAND, "100", "mpecho FIRED")
	ch, client := watcherIn(room, "Watcher")
	defer client.Close()

	progNest = 0
	RprogRandomTrigger(room)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("RAND: got %q", out)
	}
}

func TestRprogActTrigger_Fires(t *testing.T) {
	room := makeProgRoom(3011, types.MPROG_ACT, "laughs", "mpecho FIRED")
	ch, client := watcherIn(room, "Watcher")
	defer client.Close()

	progNest = 0
	RprogActTrigger("Someone laughs hysterically", room)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("ACT: got %q", out)
	}
}

func TestRprogHourTrigger_FiresAtMatchingHour(t *testing.T) {
	room := makeProgRoom(3012, types.MPROG_HOUR, "9", "mpecho FIRED")
	ch, client := watcherIn(room, "Watcher")
	defer client.Close()

	w := world.New("/tmp/test-rprog")
	w.Rooms[room.Vnum] = room

	progNest = 0
	RprogHourTrigger(9, w)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("HOUR: got %q", out)
	}
}

func TestRprogTimeTrigger_FiresOncePerHour(t *testing.T) {
	room := makeProgRoom(3013, types.MPROG_TIME, "12", "mpecho FIRED")
	ch, client := watcherIn(room, "Watcher")
	defer client.Close()

	w := world.New("/tmp/test-rprog")
	w.Rooms[room.Vnum] = room

	progNest = 0
	RprogTimeTrigger(12, w)
	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("TIME first fire: got %q", out)
	}

	// Second call same hour: TIME must latch (not re-fire).
	RprogTimeTrigger(12, w)
	out2 := readTrigOutput(ch, client)
	if strings.Contains(out2, "FIRED") {
		t.Errorf("TIME re-fire without hour change: got %q", out2)
	}

	// Different hour resets latch — but arglist is 12, so 13 shouldn't fire.
	RprogTimeTrigger(13, w)
	// Going back to 12 should fire again because the latch was cleared.
	RprogTimeTrigger(12, w)
	out3 := readTrigOutput(ch, client)
	if !strings.Contains(out3, "FIRED") {
		t.Errorf("TIME second fire after latch reset: got %q", out3)
	}
}

func TestRoomTrigger_NilRoom(t *testing.T) {
	RoomTrigger(types.MPROG_ENTER, nil, nil, nil, nil, nil, "")
}

func TestRoomTrigger_NoProgs(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 99}
	RoomTrigger(types.MPROG_ENTER, room, nil, nil, nil, nil, "")
}

func TestRprogEnterTrigger_NilChar(t *testing.T) {
	RprogEnterTrigger(nil)
}

func TestRprogCommandTrigger_NilRoom(t *testing.T) {
	ch, client := makeDescChar("x")
	defer client.Close()
	if RprogCommandTrigger(ch, "wiggle") {
		t.Error("nil-room CMD trigger must return false")
	}
}
