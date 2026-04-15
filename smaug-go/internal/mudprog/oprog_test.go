package mudprog

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// makeTestObj builds an ObjData + matching ObjIndexData carrying a single
// MudProg of the given type/arglist/comlist. The obj is placed on the floor
// of room if non-nil.
func makeTestObj(room *types.RoomIndexData, progType int64, argList, comList string) *types.ObjData {
	idx := &types.ObjIndexData{
		Vnum:       8000,
		Name:       "widget",
		ShortDescr: "a small widget",
		MudProgs: []*types.MProgData{
			{Type: progType, ArgList: argList, ComList: comList},
		},
	}
	obj := &types.ObjData{
		IndexData:  idx,
		Name:       idx.Name,
		ShortDescr: idx.ShortDescr,
	}
	if room != nil {
		obj.InRoom = room
		room.Contents = append(room.Contents, obj)
	}
	return obj
}

func TestObjTrigger_NilObj(t *testing.T) {
	// Must not panic.
	ObjTrigger(types.MPROG_WEAR, nil, nil, nil, nil, "")
}

func TestObjTrigger_NilIndex(t *testing.T) {
	obj := &types.ObjData{}
	ObjTrigger(types.MPROG_WEAR, obj, nil, nil, nil, "")
}

func TestObjTrigger_NoMatchingType(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001, Name: "Test"}
	obj := makeTestObj(room, types.MPROG_DROP, "", "mpecho FIRED")

	ch, client := makeDescChar("Watcher")
	defer client.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	progNest = 0
	ObjTrigger(types.MPROG_WEAR, obj, ch, nil, nil, "")

	out := readTrigOutput(ch, client)
	if strings.Contains(out, "FIRED") {
		t.Errorf("unrelated prog type should not fire, got %q", out)
	}
}

func TestResolveObjRoom(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	// Direct: on the floor.
	obj := &types.ObjData{InRoom: room}
	if got := resolveObjRoom(obj); got != room {
		t.Errorf("direct: expected room, got %v", got)
	}
	// Carried: follow CarriedBy.InRoom.
	carrier := &types.CharData{Name: "Carrier", InRoom: room}
	held := &types.ObjData{CarriedBy: carrier}
	if got := resolveObjRoom(held); got != room {
		t.Errorf("carried: expected carrier room, got %v", got)
	}
	// Nested: follow InObj chain, then CarriedBy.
	outer := &types.ObjData{CarriedBy: carrier}
	inner := &types.ObjData{InObj: outer}
	if got := resolveObjRoom(inner); got != room {
		t.Errorf("nested: expected outer carrier room, got %v", got)
	}
	// Nil: returns nil.
	if got := resolveObjRoom(nil); got != nil {
		t.Errorf("nil obj: expected nil, got %v", got)
	}
}

func TestBuildSupermob(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	obj := &types.ObjData{ShortDescr: "a glowing orb", InRoom: room,
		IndexData: &types.ObjIndexData{}}
	sm := buildSupermob(obj)
	if sm == nil {
		t.Fatal("supermob should be non-nil for valid obj")
	}
	if !sm.IsNPC() {
		t.Error("supermob must be flagged NPC for driver semantics")
	}
	if sm.InRoom != room {
		t.Error("supermob should inherit obj's effective room")
	}
	if sm.ShortDescr != obj.ShortDescr {
		t.Errorf("supermob should inherit obj short descr, got %q", sm.ShortDescr)
	}
	if sm.Position != types.POS_STANDING {
		t.Error("supermob should be standing so position gates pass")
	}
	// supermob must not be inserted into room.People.
	for _, p := range room.People {
		if p == sm {
			t.Error("supermob should not be added to room.People")
		}
	}
}

// --- Per-helper fire tests. Each constructs a room with a watcher PC, an
// obj whose prog mpechoes "FIRED" on the matching trigger, then calls the
// helper and verifies the echo landed in the PC's output buffer. ---

func TestOprogWearTrigger(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	obj := makeTestObj(room, types.MPROG_WEAR, "", "mpecho FIRED")
	ch, client := makeDescChar("Watcher")
	defer client.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	progNest = 0
	OprogWearTrigger(ch, obj)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("WEAR: expected FIRED in output, got %q", out)
	}
}

func TestOprogRemoveTrigger(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	obj := makeTestObj(room, types.MPROG_REMOVE, "", "mpecho FIRED")
	ch, client := makeDescChar("Watcher")
	defer client.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	progNest = 0
	OprogRemoveTrigger(ch, obj)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("REMOVE: got %q", out)
	}
}

func TestOprogSacTrigger(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	obj := makeTestObj(room, types.MPROG_SAC, "", "mpecho FIRED")
	ch, client := makeDescChar("Watcher")
	defer client.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	progNest = 0
	OprogSacTrigger(ch, obj)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("SAC: got %q", out)
	}
}

func TestOprogLookTrigger(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	obj := makeTestObj(room, types.MPROG_LOOK, "", "mpecho FIRED")
	ch, client := makeDescChar("Watcher")
	defer client.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	progNest = 0
	OprogLookTrigger(ch, obj)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("LOOK: got %q", out)
	}
}

func TestOprogExamineTrigger(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	obj := makeTestObj(room, types.MPROG_EXA, "", "mpecho FIRED")
	ch, client := makeDescChar("Watcher")
	defer client.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	progNest = 0
	OprogExamineTrigger(ch, obj)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("EXA: got %q", out)
	}
}

func TestOprogZapTrigger(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	obj := makeTestObj(room, types.MPROG_ZAP, "", "mpecho FIRED")
	ch, client := makeDescChar("Watcher")
	defer client.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	progNest = 0
	OprogZapTrigger(ch, obj)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("ZAP: got %q", out)
	}
}

func TestOprogGetTrigger(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	obj := makeTestObj(room, types.MPROG_GET, "", "mpecho FIRED")
	ch, client := makeDescChar("Watcher")
	defer client.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	progNest = 0
	OprogGetTrigger(ch, obj)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("GET: got %q", out)
	}
}

func TestOprogDropTrigger(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	obj := makeTestObj(room, types.MPROG_DROP, "", "mpecho FIRED")
	ch, client := makeDescChar("Watcher")
	defer client.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	progNest = 0
	OprogDropTrigger(ch, obj)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("DROP: got %q", out)
	}
}

func TestOprogDamageTrigger(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	obj := makeTestObj(room, types.MPROG_DAMAGE, "", "mpecho FIRED")
	ch, client := makeDescChar("Watcher")
	defer client.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	progNest = 0
	OprogDamageTrigger(ch, obj)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("DAMAGE: got %q", out)
	}
}

func TestOprogRepairTrigger(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	obj := makeTestObj(room, types.MPROG_REPAIR, "", "mpecho FIRED")
	ch, client := makeDescChar("Watcher")
	defer client.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	progNest = 0
	OprogRepairTrigger(ch, obj)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("REPAIR: got %q", out)
	}
}

func TestOprogPullTrigger(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	obj := makeTestObj(room, types.MPROG_PULL, "", "mpecho FIRED")
	ch, client := makeDescChar("Watcher")
	defer client.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	progNest = 0
	OprogPullTrigger(ch, obj)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("PULL: got %q", out)
	}
}

func TestOprogPushTrigger(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	obj := makeTestObj(room, types.MPROG_PUSH, "", "mpecho FIRED")
	ch, client := makeDescChar("Watcher")
	defer client.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	progNest = 0
	OprogPushTrigger(ch, obj)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("PUSH: got %q", out)
	}
}

func TestOprogUseTrigger(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	obj := makeTestObj(room, types.MPROG_USE, "", "mpecho FIRED")
	ch, client := makeDescChar("Watcher")
	defer client.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	progNest = 0
	OprogUseTrigger(ch, obj)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("USE: got %q", out)
	}
}

func TestOprogGreetTrigger_FiresOnRoomContents(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	_ = makeTestObj(room, types.MPROG_GREET, "", "mpecho FIRED")
	ch, client := makeDescChar("Visitor")
	defer client.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	progNest = 0
	OprogGreetTrigger(ch)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("GREET: got %q", out)
	}
}

func TestOprogGreetTrigger_NilInRoom(t *testing.T) {
	ch := &types.CharData{Name: "Floater"}
	OprogGreetTrigger(ch) // must not panic
}

func TestOprogSpeechTrigger_FiresOnRoomContents(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	_ = makeTestObj(room, types.MPROG_SPEECH, "hello", "mpecho FIRED")
	ch, client := makeDescChar("Speaker")
	defer client.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	progNest = 0
	OprogSpeechTrigger(ch, "I say hello to you")

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("SPEECH in room: got %q", out)
	}
}

func TestOprogSpeechTrigger_FiresOnCarrying(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	ch, client := makeDescChar("Speaker")
	defer client.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	// Held obj (not in room.Contents) with speech prog.
	idx := &types.ObjIndexData{
		Vnum:       8000,
		Name:       "amulet",
		ShortDescr: "a chatty amulet",
		MudProgs: []*types.MProgData{
			{Type: types.MPROG_SPEECH, ArgList: "hello", ComList: "mpecho FIRED"},
		},
	}
	held := &types.ObjData{IndexData: idx, Name: idx.Name, ShortDescr: idx.ShortDescr,
		CarriedBy: ch}
	ch.Carrying = append(ch.Carrying, held)

	progNest = 0
	OprogSpeechTrigger(ch, "hello there")

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("SPEECH on carried: got %q", out)
	}
}

func TestOprogSpeechTrigger_NoMatch(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	_ = makeTestObj(room, types.MPROG_SPEECH, "password", "mpecho FIRED")
	ch, client := makeDescChar("Speaker")
	defer client.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	progNest = 0
	OprogSpeechTrigger(ch, "hello there")

	out := readTrigOutput(ch, client)
	if strings.Contains(out, "FIRED") {
		t.Errorf("SPEECH no-match should not fire, got %q", out)
	}
}

func TestOprogCommandTrigger_Stub(t *testing.T) {
	// Current Go port has no MPROG_CMD bit; helper is a stub that never
	// consumes commands. Will flip to true once Tier 3 adds the bit.
	ch, client := makeDescChar("Commander")
	defer client.Close()
	if OprogCommandTrigger(ch, "wiggle") {
		t.Error("command trigger should return false until MPROG_CMD is defined")
	}
}

func TestOprogRandomTrigger_FiresInRoom(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	// RAND arglist "100" = always fire.
	obj := makeTestObj(room, types.MPROG_RAND, "100", "mpecho FIRED")
	ch, client := makeDescChar("Watcher")
	defer client.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	progNest = 0
	OprogRandomTrigger(obj)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("RAND: got %q", out)
	}
}

func TestOprogActTrigger(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	obj := makeTestObj(room, types.MPROG_ACT, "attacks", "mpecho FIRED")
	ch, client := makeDescChar("Watcher")
	defer client.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	progNest = 0
	OprogActTrigger("A warrior attacks the dummy.", obj, ch)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("ACT: got %q", out)
	}
}

func TestObjTrigger_SkipsProgsWithWrongType(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	idx := &types.ObjIndexData{
		Vnum:       8001,
		Name:       "statue",
		ShortDescr: "a statue",
		MudProgs: []*types.MProgData{
			{Type: types.MPROG_WEAR, ArgList: "", ComList: "mpecho WRONG"},
			{Type: types.MPROG_GET, ArgList: "", ComList: "mpecho FIRED"},
		},
	}
	obj := &types.ObjData{IndexData: idx, Name: idx.Name, ShortDescr: idx.ShortDescr,
		InRoom: room}
	room.Contents = append(room.Contents, obj)

	ch, client := makeDescChar("Watcher")
	defer client.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	progNest = 0
	OprogGetTrigger(ch, obj)

	out := readTrigOutput(ch, client)
	if !strings.Contains(out, "FIRED") {
		t.Errorf("expected GET prog to fire, got %q", out)
	}
	if strings.Contains(out, "WRONG") {
		t.Errorf("WEAR prog must not fire on GET trigger, got %q", out)
	}
}
