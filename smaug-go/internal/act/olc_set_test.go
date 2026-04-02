package act

import (
	"net"
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

// --- DoMset tests ---

func setupMsetWorld() (*types.CharData, net.Conn, *types.CharData, func()) {
	w := setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	room := &types.RoomIndexData{Vnum: 9100, Name: "OLC Room"}
	w.Rooms[9100] = room
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	mob := &types.CharData{
		Name:       "testmob",
		ShortDescr: "a test mob",
		Level:      1,
		Hit:        10,
		MaxHit:     10,
		Mana:       50,
		MaxMana:    50,
		Move:       100,
		MaxMove:    100,
		Gold:       0,
	}
	mob.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(mob, room)
	w.AddChar(mob)

	return ch, client, mob, func() { client.Close() }
}

func TestDoMset_NoArg(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoMset(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Syntax") {
		t.Errorf("expected syntax help, got: %q", out)
	}
}

func TestDoMset_TargetNotFound(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 1, Name: "Test"}

	DoMset(ch, "nobody level 5")
	out := readOutput(ch, client)
	if !strings.Contains(out, "aren't here") {
		t.Errorf("expected 'aren't here', got: %q", out)
	}
}

func TestDoMset_NoField(t *testing.T) {
	ch, client, _, cleanup := setupMsetWorld()
	defer cleanup()

	DoMset(ch, "testmob")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Syntax") {
		t.Errorf("expected syntax help, got: %q", out)
	}
}

func TestDoMset_Fields(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		check   func(mob *types.CharData) bool
		wantOut string
	}{
		{"level", "testmob level 50", func(m *types.CharData) bool { return m.Level == 50 }, "level set to"},
		{"str", "testmob str 18", func(m *types.CharData) bool { return m.PermStr == 18 }, "str set to"},
		{"int", "testmob int 20", func(m *types.CharData) bool { return m.PermInt == 20 }, "int set to"},
		{"wis", "testmob wis 15", func(m *types.CharData) bool { return m.PermWis == 15 }, "wis set to"},
		{"dex", "testmob dex 22", func(m *types.CharData) bool { return m.PermDex == 22 }, "dex set to"},
		{"con", "testmob con 25", func(m *types.CharData) bool { return m.PermCon == 25 }, "con set to"},
		{"cha", "testmob cha 12", func(m *types.CharData) bool { return m.PermCha == 12 }, "cha set to"},
		{"lck", "testmob lck 10", func(m *types.CharData) bool { return m.PermLck == 10 }, "lck set to"},
		{"hp", "testmob hp 200", func(m *types.CharData) bool { return m.MaxHit == 200 && m.Hit == 200 }, "hp set to"},
		{"mana", "testmob mana 300", func(m *types.CharData) bool { return m.MaxMana == 300 && m.Mana == 300 }, "mana set to"},
		{"move", "testmob move 500", func(m *types.CharData) bool { return m.MaxMove == 500 && m.Move == 500 }, "move set to"},
		{"hitroll", "testmob hitroll 15", func(m *types.CharData) bool { return m.Hitroll == 15 }, "hitroll set to"},
		{"damroll", "testmob damroll 20", func(m *types.CharData) bool { return m.Damroll == 20 }, "damroll set to"},
		{"gold", "testmob gold 1000", func(m *types.CharData) bool { return m.Gold == 1000 }, "gold set to"},
		{"align", "testmob align -500", func(m *types.CharData) bool { return m.Alignment == -500 }, "align set to"},
		{"sex_female", "testmob sex female", func(m *types.CharData) bool { return m.Sex == 2 }, "sex set to"},
		{"sex_male", "testmob sex male", func(m *types.CharData) bool { return m.Sex == 1 }, "sex set to"},
		{"sex_neutral", "testmob sex neutral", func(m *types.CharData) bool { return m.Sex == 0 }, "sex set to"},
		{"race", "testmob race 5", func(m *types.CharData) bool { return m.Race == 5 }, "race set to"},
		{"class", "testmob class 3", func(m *types.CharData) bool { return m.Class == 3 }, "class set to"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ch, client, mob, cleanup := setupMsetWorld()
			defer cleanup()

			DoMset(ch, tc.args)
			out := readOutput(ch, client)

			if !strings.Contains(strings.ToLower(out), tc.wantOut) {
				t.Errorf("expected output containing %q, got: %q", tc.wantOut, out)
			}
			if !tc.check(mob) {
				t.Errorf("field check failed for %s", tc.name)
			}
		})
	}
}

func TestDoMset_Name(t *testing.T) {
	ch, client, mob, cleanup := setupMsetWorld()
	defer cleanup()

	DoMset(ch, "testmob name big dragon")
	out := readOutput(ch, client)

	if mob.Name != "big dragon" {
		t.Errorf("expected name 'big dragon', got: %q", mob.Name)
	}
	if !strings.Contains(strings.ToLower(out), "name set to") {
		t.Errorf("expected confirmation, got: %q", out)
	}
}

func TestDoMset_Short(t *testing.T) {
	ch, client, mob, cleanup := setupMsetWorld()
	defer cleanup()

	DoMset(ch, "testmob short a fierce dragon")
	out := readOutput(ch, client)

	if mob.ShortDescr != "a fierce dragon" {
		t.Errorf("expected short 'a fierce dragon', got: %q", mob.ShortDescr)
	}
	if !strings.Contains(strings.ToLower(out), "short set to") {
		t.Errorf("expected confirmation, got: %q", out)
	}
}

func TestDoMset_Long(t *testing.T) {
	ch, client, mob, cleanup := setupMsetWorld()
	defer cleanup()

	DoMset(ch, "testmob long A fierce dragon stands here, breathing fire.")
	out := readOutput(ch, client)

	if mob.LongDescr != "A fierce dragon stands here, breathing fire." {
		t.Errorf("expected long descr, got: %q", mob.LongDescr)
	}
	if !strings.Contains(strings.ToLower(out), "long set to") {
		t.Errorf("expected confirmation, got: %q", out)
	}
}

func TestDoMset_BadField(t *testing.T) {
	ch, client, _, cleanup := setupMsetWorld()
	defer cleanup()

	DoMset(ch, "testmob badfield 5")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Valid fields") {
		t.Errorf("expected valid fields list, got: %q", out)
	}
}

func TestDoMset_StrClamp(t *testing.T) {
	ch, client, mob, cleanup := setupMsetWorld()
	defer cleanup()

	DoMset(ch, "testmob str 99")
	_ = readOutput(ch, client)
	if mob.PermStr != 25 {
		t.Errorf("expected str clamped to 25, got: %d", mob.PermStr)
	}
}

func TestDoMset_AlignClamp(t *testing.T) {
	ch, client, mob, cleanup := setupMsetWorld()
	defer cleanup()

	DoMset(ch, "testmob align 5000")
	_ = readOutput(ch, client)
	if mob.Alignment != 1000 {
		t.Errorf("expected align clamped to 1000, got: %d", mob.Alignment)
	}
}

// --- DoOset tests ---

func setupOsetWorld() (*types.CharData, net.Conn, *types.ObjData, func()) {
	w := setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	room := &types.RoomIndexData{Vnum: 9200, Name: "OLC Obj Room"}
	w.Rooms[9200] = room
	handler.CharToRoom(ch, room)

	obj := &types.ObjData{Name: "sword", ShortDescr: "a sword", Level: 1}
	handler.ObjToRoom(obj, room)

	return ch, client, obj, func() { client.Close() }
}

func TestDoOset_NoArg(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoOset(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Syntax") {
		t.Errorf("expected syntax help, got: %q", out)
	}
}

func TestDoOset_NotFound(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 1, Name: "Test"}

	DoOset(ch, "nothing type 5")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Nothing like that") {
		t.Errorf("expected 'Nothing like that', got: %q", out)
	}
}

func TestDoOset_Fields(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		check   func(obj *types.ObjData) bool
		wantOut string
	}{
		{"type", "sword type 5", func(o *types.ObjData) bool { return o.ItemType == 5 }, "type set to"},
		{"weight", "sword weight 10", func(o *types.ObjData) bool { return o.Weight == 10 }, "weight set to"},
		{"cost", "sword cost 500", func(o *types.ObjData) bool { return o.GoldCost == 500 }, "cost set to"},
		{"level", "sword level 20", func(o *types.ObjData) bool { return o.Level == 20 }, "level set to"},
		{"value0", "sword value0 42", func(o *types.ObjData) bool { return o.Value[0] == 42 }, "value0 set to"},
		{"value1", "sword value1 7", func(o *types.ObjData) bool { return o.Value[1] == 7 }, "value1 set to"},
		{"value2", "sword value2 3", func(o *types.ObjData) bool { return o.Value[2] == 3 }, "value2 set to"},
		{"value3", "sword value3 9", func(o *types.ObjData) bool { return o.Value[3] == 9 }, "value3 set to"},
		{"value4", "sword value4 1", func(o *types.ObjData) bool { return o.Value[4] == 1 }, "value4 set to"},
		{"value5", "sword value5 8", func(o *types.ObjData) bool { return o.Value[5] == 8 }, "value5 set to"},
		{"wearflags", "sword wearflags 3", func(o *types.ObjData) bool { return o.WearFlags == 3 }, "wearflags set to"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ch, client, obj, cleanup := setupOsetWorld()
			defer cleanup()

			DoOset(ch, tc.args)
			out := readOutput(ch, client)

			if !strings.Contains(strings.ToLower(out), tc.wantOut) {
				t.Errorf("expected output containing %q, got: %q", tc.wantOut, out)
			}
			if !tc.check(obj) {
				t.Errorf("field check failed for %s", tc.name)
			}
		})
	}
}

func TestDoOset_Name(t *testing.T) {
	ch, client, obj, cleanup := setupOsetWorld()
	defer cleanup()

	DoOset(ch, "sword name shiny sword")
	out := readOutput(ch, client)

	if obj.Name != "shiny sword" {
		t.Errorf("expected name 'shiny sword', got: %q", obj.Name)
	}
	if !strings.Contains(strings.ToLower(out), "name set to") {
		t.Errorf("expected confirmation, got: %q", out)
	}
}

func TestDoOset_Short(t *testing.T) {
	ch, client, obj, cleanup := setupOsetWorld()
	defer cleanup()

	DoOset(ch, "sword short a gleaming blade")
	out := readOutput(ch, client)

	if obj.ShortDescr != "a gleaming blade" {
		t.Errorf("expected short 'a gleaming blade', got: %q", obj.ShortDescr)
	}
	if !strings.Contains(strings.ToLower(out), "short set to") {
		t.Errorf("expected confirmation, got: %q", out)
	}
}

func TestDoOset_Long(t *testing.T) {
	ch, client, obj, cleanup := setupOsetWorld()
	defer cleanup()

	DoOset(ch, "sword long A gleaming blade lies on the floor.")
	out := readOutput(ch, client)

	if obj.Description != "A gleaming blade lies on the floor." {
		t.Errorf("expected long descr, got: %q", obj.Description)
	}
	if !strings.Contains(strings.ToLower(out), "long set to") {
		t.Errorf("expected confirmation, got: %q", out)
	}
}

func TestDoOset_BadField(t *testing.T) {
	ch, client, _, cleanup := setupOsetWorld()
	defer cleanup()

	DoOset(ch, "sword badfield 5")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Valid fields") {
		t.Errorf("expected valid fields list, got: %q", out)
	}
}

// --- DoRset tests ---

func TestDoRset_NoArg(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 1, Name: "Test"}

	DoRset(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Syntax") || !strings.Contains(out, "Valid fields") {
		t.Errorf("expected syntax or valid fields, got: %q", out)
	}
}

func TestDoRset_Sector(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9300, Name: "Test Room", SectorType: 0}
	ch.InRoom = room

	DoRset(ch, "sector 3")
	out := readOutput(ch, client)

	if room.SectorType != 3 {
		t.Errorf("expected sector 3, got: %d", room.SectorType)
	}
	if !strings.Contains(strings.ToLower(out), "sector set to") {
		t.Errorf("expected confirmation, got: %q", out)
	}
}

func TestDoRset_Flags(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9300, Name: "Test Room"}
	ch.InRoom = room

	DoRset(ch, "flags 2")
	out := readOutput(ch, client)

	if !room.RoomFlags.IsSet(2) {
		t.Errorf("expected flag 2 to be set")
	}
	if !strings.Contains(strings.ToLower(out), "flag") {
		t.Errorf("expected flag confirmation, got: %q", out)
	}
}

func TestDoRset_NoRoom(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = nil

	DoRset(ch, "sector 3")
	out := readOutput(ch, client)
	if !strings.Contains(out, "not in a room") {
		t.Errorf("expected 'not in a room', got: %q", out)
	}
}

func TestDoRset_BadField(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 1, Name: "Test"}

	DoRset(ch, "badfield 5")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Valid fields") {
		t.Errorf("expected valid fields, got: %q", out)
	}
}
