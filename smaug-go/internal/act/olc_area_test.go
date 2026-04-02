package act

import (
	"net"
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

func setupAreaWorld() (*types.CharData, net.Conn, *types.AreaData, func()) {
	w := setupOlcWorld()
	area := &types.AreaData{
		Name:           "Test Area",
		Author:         "Tester",
		Filename:       "test.are",
		ResetMsg:       "The area resets.",
		LowRVnum:       100,
		HiRVnum:        199,
		LowMVnum:       200,
		HiMVnum:        299,
		LowOVnum:       300,
		HiOVnum:        399,
		ResetFrequency: 15,
		Age:            5,
		NPlayer:        2,
		Flags:          0,
	}
	w.Areas = append(w.Areas, area)

	ch, client := makeImmTestChar("Builder")
	room := &types.RoomIndexData{Vnum: 100, Name: "Area Room", Area: area}
	w.Rooms[100] = room
	handler.CharToRoom(ch, room)

	return ch, client, area, func() { client.Close() }
}

// --- DoAset tests ---

func TestDoAset_NoArg(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoAset(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Syntax") {
		t.Errorf("expected syntax help, got: %q", out)
	}
}

func TestDoAset_AreaNotFound(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoAset(ch, "nonexistent name Foo")
	out := readOutput(ch, client)
	if !strings.Contains(out, "No such area") {
		t.Errorf("expected 'No such area', got: %q", out)
	}
}

func TestDoAset_NoField(t *testing.T) {
	ch, client, _, cleanup := setupAreaWorld()
	defer cleanup()

	DoAset(ch, "test")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Syntax") || !strings.Contains(out, "Valid fields") {
		t.Errorf("expected syntax or valid fields, got: %q", out)
	}
}

func TestDoAset_Fields(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		check   func(area *types.AreaData) bool
		wantOut string
	}{
		{"name", "test name New Name", func(a *types.AreaData) bool { return a.Name == "New Name" }, "name set to"},
		{"author", "test author Eilidh", func(a *types.AreaData) bool { return a.Author == "Eilidh" }, "author set to"},
		{"resetmsg", "test resetmsg The wind blows.", func(a *types.AreaData) bool { return a.ResetMsg == "The wind blows." }, "resetmsg set to"},
		{"low_r_vnum", "test low_r_vnum 50", func(a *types.AreaData) bool { return a.LowRVnum == 50 }, "low_r_vnum set to"},
		{"hi_r_vnum", "test hi_r_vnum 150", func(a *types.AreaData) bool { return a.HiRVnum == 150 }, "hi_r_vnum set to"},
		{"low_m_vnum", "test low_m_vnum 60", func(a *types.AreaData) bool { return a.LowMVnum == 60 }, "low_m_vnum set to"},
		{"hi_m_vnum", "test hi_m_vnum 160", func(a *types.AreaData) bool { return a.HiMVnum == 160 }, "hi_m_vnum set to"},
		{"low_o_vnum", "test low_o_vnum 70", func(a *types.AreaData) bool { return a.LowOVnum == 70 }, "low_o_vnum set to"},
		{"hi_o_vnum", "test hi_o_vnum 170", func(a *types.AreaData) bool { return a.HiOVnum == 170 }, "hi_o_vnum set to"},
		{"resetfreq", "test resetfreq 20", func(a *types.AreaData) bool { return a.ResetFrequency == 20 }, "resetfreq set to"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ch, client, area, cleanup := setupAreaWorld()
			defer cleanup()

			DoAset(ch, tc.args)
			out := readOutput(ch, client)

			if !strings.Contains(strings.ToLower(out), tc.wantOut) {
				t.Errorf("expected output containing %q, got: %q", tc.wantOut, out)
			}
			if !tc.check(area) {
				t.Errorf("field check failed for %s", tc.name)
			}
		})
	}
}

func TestDoAset_BadField(t *testing.T) {
	ch, client, _, cleanup := setupAreaWorld()
	defer cleanup()

	DoAset(ch, "test badfield 5")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Valid fields") {
		t.Errorf("expected valid fields, got: %q", out)
	}
}

// --- DoAstat tests ---

func TestDoAstat_NoArg_CurrentArea(t *testing.T) {
	ch, client, area, cleanup := setupAreaWorld()
	defer cleanup()

	DoAstat(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, area.Name) {
		t.Errorf("expected area name %q in output, got: %q", area.Name, out)
	}
	if !strings.Contains(out, area.Author) {
		t.Errorf("expected author %q in output, got: %q", area.Author, out)
	}
	if !strings.Contains(out, area.Filename) {
		t.Errorf("expected filename %q in output, got: %q", area.Filename, out)
	}
}

func TestDoAstat_NamedArea(t *testing.T) {
	ch, client, area, cleanup := setupAreaWorld()
	defer cleanup()

	DoAstat(ch, "test")
	out := readOutput(ch, client)

	if !strings.Contains(out, area.Name) {
		t.Errorf("expected area name %q in output, got: %q", area.Name, out)
	}
}

func TestDoAstat_AreaNotFound(t *testing.T) {
	ch, client, _, cleanup := setupAreaWorld()
	defer cleanup()

	DoAstat(ch, "nonexistent")
	out := readOutput(ch, client)
	if !strings.Contains(out, "No such area") {
		t.Errorf("expected 'No such area', got: %q", out)
	}
}

func TestDoAstat_NoRoom(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.InRoom = nil

	DoAstat(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "not in") {
		t.Errorf("expected 'not in' message, got: %q", out)
	}
}

func TestDoAstat_OutputContainsVnumRanges(t *testing.T) {
	ch, client, _, cleanup := setupAreaWorld()
	defer cleanup()

	DoAstat(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "100") || !strings.Contains(out, "199") {
		t.Errorf("expected room vnum range in output, got: %q", out)
	}
	if !strings.Contains(out, "Reset Freq") {
		t.Errorf("expected 'Reset Freq' in output, got: %q", out)
	}
}
