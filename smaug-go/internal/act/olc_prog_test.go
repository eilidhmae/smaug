package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

func TestDoMpedit_NoArg(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	DoMpedit(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Usage") {
		t.Errorf("expected usage, got: %q", out)
	}
}

func TestDoMpedit_UnknownVnum(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	DoMpedit(ch, "99999")
	out := readOutput(ch, client)
	if !strings.Contains(out, "does not exist") {
		t.Errorf("expected does-not-exist, got: %q", out)
	}
}

func TestDoMpedit_ListProgs(t *testing.T) {
	w := setupOlcWorld()
	idx := &types.MobIndexData{Vnum: 6100, PlayerName: "x"}
	idx.MudProgs = []*types.MProgData{
		{Type: types.MPROG_ACT, ArgList: "p hi", ComList: "say hello"},
		{Type: types.MPROG_GREET, ArgList: "100", ComList: "wave"},
	}
	w.MobIndex[6100] = idx
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoMpedit(ch, "6100")
	out := readOutput(ch, client)
	if !strings.Contains(out, "2 mob-prog") {
		t.Errorf("expected count in output, got: %q", out)
	}
	if !strings.Contains(out, "trigger=act") {
		t.Errorf("expected act trigger name, got: %q", out)
	}
	if !strings.Contains(out, "trigger=greet") {
		t.Errorf("expected greet trigger name, got: %q", out)
	}
}

func TestDoMpedit_ShowByTrigger(t *testing.T) {
	w := setupOlcWorld()
	idx := &types.MobIndexData{Vnum: 6101, PlayerName: "x"}
	idx.MudProgs = []*types.MProgData{
		{Type: types.MPROG_ACT, ArgList: "p waves", ComList: "say hi"},
		{Type: types.MPROG_GREET, ArgList: "100", ComList: "say hello"},
	}
	w.MobIndex[6101] = idx
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoMpedit(ch, "6101 greet")
	out := readOutput(ch, client)
	if !strings.Contains(out, "say hello") {
		t.Errorf("expected the greet script, got: %q", out)
	}
	if strings.Contains(out, "say hi") {
		t.Errorf("should not show act prog, got: %q", out)
	}
}

func TestDoMpedit_NoMatchingTrigger(t *testing.T) {
	w := setupOlcWorld()
	idx := &types.MobIndexData{Vnum: 6102, PlayerName: "x"}
	idx.MudProgs = []*types.MProgData{{Type: types.MPROG_ACT, ComList: "x"}}
	w.MobIndex[6102] = idx
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoMpedit(ch, "6102 greet")
	out := readOutput(ch, client)
	if !strings.Contains(out, "No mob-prog") {
		t.Errorf("expected no-match, got: %q", out)
	}
}

func TestDoMpedit_BadTrigger(t *testing.T) {
	w := setupOlcWorld()
	w.MobIndex[6103] = &types.MobIndexData{Vnum: 6103, PlayerName: "x"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoMpedit(ch, "6103 xyzzy")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Unknown trigger") {
		t.Errorf("expected unknown-trigger, got: %q", out)
	}
}

func TestDoOpedit_Inspector(t *testing.T) {
	w := setupOlcWorld()
	idx := &types.ObjIndexData{Vnum: 6200, Name: "obj"}
	idx.MudProgs = []*types.MProgData{{Type: types.MPROG_WEAR, ComList: "echo you wear"}}
	w.ObjIndex[6200] = idx
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoOpedit(ch, "6200 wear")
	out := readOutput(ch, client)
	if !strings.Contains(out, "echo you wear") {
		t.Errorf("expected wear script, got: %q", out)
	}
}

func TestDoRpedit_Inspector(t *testing.T) {
	w := setupOlcWorld()
	room := &types.RoomIndexData{Vnum: 6300, Name: "Test"}
	room.MudProgs = []*types.MProgData{{Type: types.MPROG_ENTER, ComList: "echo arrival"}}
	w.Rooms[6300] = room
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoRpedit(ch, "6300 enter")
	out := readOutput(ch, client)
	if !strings.Contains(out, "echo arrival") {
		t.Errorf("expected enter script, got: %q", out)
	}
}

func TestProgEditors_MortalReject(t *testing.T) {
	_ = setupOlcWorld()
	mortal, client := makeTestChar("Mortal")
	defer client.Close()
	mortal.Level = 1

	for _, fn := range []func(*types.CharData, string){DoMpedit, DoOpedit, DoRpedit} {
		fn(mortal, "1")
		out := readOutput(mortal, client)
		if !strings.Contains(out, "Huh?") {
			t.Errorf("expected Huh?, got: %q", out)
		}
	}
}
