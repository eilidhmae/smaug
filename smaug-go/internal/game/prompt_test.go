package game

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

func TestFormatPrompt(t *testing.T) {
	ch := &types.CharData{
		Hit:       85,
		MaxHit:    100,
		Mana:      42,
		MaxMana:   50,
		Move:      75,
		MaxMove:   80,
		Gold:      500,
		Exp:       12345,
		Alignment: 350,
		PCData: &types.PCData{
			Prompt: "<%hhp %mm %vmv> ",
		},
		InRoom: &types.RoomIndexData{Name: "Temple"},
	}

	out := FormatPrompt(ch)
	if out != "<85hp 42m 75mv> " {
		t.Errorf("FormatPrompt = %q, want %q", out, "<85hp 42m 75mv> ")
	}
}

func TestFormatPrompt_AllTokens(t *testing.T) {
	ch := &types.CharData{
		Hit:       85,
		MaxHit:    100,
		Mana:      42,
		MaxMana:   50,
		Move:      75,
		MaxMove:   80,
		Gold:      500,
		Exp:       12345,
		Alignment: 350,
		PCData: &types.PCData{
			Prompt: "%h/%H %m/%M %v/%V g%g a%a x%x r%r %%",
		},
		InRoom: &types.RoomIndexData{Name: "Temple"},
	}

	out := FormatPrompt(ch)
	expected := "85/100 42/50 75/80 g500 a350 x12345 rTemple %"
	if out != expected {
		t.Errorf("FormatPrompt = %q, want %q", out, expected)
	}
}

func TestFormatPrompt_Default(t *testing.T) {
	ch := &types.CharData{
		Hit:  50,
		Mana: 30,
		Move: 40,
		PCData: &types.PCData{
			Prompt: "",
		},
	}

	out := FormatPrompt(ch)
	if out != "<50hp 30m 40mv> " {
		t.Errorf("FormatPrompt default = %q, want %q", out, "<50hp 30m 40mv> ")
	}
}

func TestFormatPrompt_NoPCData(t *testing.T) {
	ch := &types.CharData{
		Hit:  50,
		Mana: 30,
		Move: 40,
	}

	out := FormatPrompt(ch)
	if out != "<50hp 30m 40mv> " {
		t.Errorf("FormatPrompt no PCData = %q, want %q", out, "<50hp 30m 40mv> ")
	}
}

func TestFormatPrompt_UnknownToken(t *testing.T) {
	ch := &types.CharData{
		PCData: &types.PCData{
			Prompt: "test%zend",
		},
	}

	out := FormatPrompt(ch)
	if out != "test%zend" {
		t.Errorf("FormatPrompt unknown = %q, want %q", out, "test%zend")
	}
}

func TestFormatPrompt_XPNextLevel(t *testing.T) {
	ch := &types.CharData{
		Exp: 5000,
		PCData: &types.PCData{
			Prompt: "XP:%x TNL:%X",
		},
	}
	out := FormatPrompt(ch)
	if out != "XP:5000 TNL:0" {
		t.Errorf("FormatPrompt XP/TNL = %q, want %q", out, "XP:5000 TNL:0")
	}
}

func TestFormatPrompt_RoomNoRoom(t *testing.T) {
	ch := &types.CharData{
		PCData: &types.PCData{
			Prompt: "Room:%r",
		},
	}
	out := FormatPrompt(ch)
	if out != "Room:Nowhere" {
		t.Errorf("FormatPrompt no room = %q, want %q", out, "Room:Nowhere")
	}
}

func TestFormatPrompt_GoldToken(t *testing.T) {
	ch := &types.CharData{
		Gold: 1234,
		PCData: &types.PCData{
			Prompt: "Gold:%g",
		},
	}
	out := FormatPrompt(ch)
	if out != "Gold:1234" {
		t.Errorf("FormatPrompt gold = %q, want %q", out, "Gold:1234")
	}
}

func TestFormatPrompt_AlignmentToken(t *testing.T) {
	ch := &types.CharData{
		Alignment: -750,
		PCData: &types.PCData{
			Prompt: "Align:%a",
		},
	}
	out := FormatPrompt(ch)
	if out != "Align:-750" {
		t.Errorf("FormatPrompt alignment = %q, want %q", out, "Align:-750")
	}
}

func TestFormatPrompt_MaxHPMaxManaMaxMove(t *testing.T) {
	ch := &types.CharData{
		MaxHit:  200,
		MaxMana: 150,
		MaxMove: 300,
		PCData: &types.PCData{
			Prompt: "%H/%M/%V",
		},
	}
	out := FormatPrompt(ch)
	if out != "200/150/300" {
		t.Errorf("FormatPrompt max values = %q, want %q", out, "200/150/300")
	}
}

func TestFormatPrompt_LiteralPercent(t *testing.T) {
	ch := &types.CharData{
		PCData: &types.PCData{
			Prompt: "100%%",
		},
	}
	out := FormatPrompt(ch)
	if out != "100%" {
		t.Errorf("FormatPrompt literal percent = %q, want %q", out, "100%%")
	}
}

func TestFormatPrompt_TrailingPercent(t *testing.T) {
	ch := &types.CharData{
		PCData: &types.PCData{
			Prompt: "end%",
		},
	}
	out := FormatPrompt(ch)
	if out != "end%" {
		t.Errorf("FormatPrompt trailing percent = %q, want %q", out, "end%%")
	}
}

func TestFormatPrompt_EachTokenIndividually(t *testing.T) {
	ch := &types.CharData{
		Hit:       10,
		MaxHit:    20,
		Mana:      30,
		MaxMana:   40,
		Move:      50,
		MaxMove:   60,
		Gold:      70,
		Alignment: 80,
		Exp:       90,
		PCData:    &types.PCData{},
		InRoom:    &types.RoomIndexData{Name: "Hall"},
	}

	tests := []struct {
		token string
		want  string
	}{
		{"%h", "10"},
		{"%H", "20"},
		{"%m", "30"},
		{"%M", "40"},
		{"%v", "50"},
		{"%V", "60"},
		{"%g", "70"},
		{"%a", "80"},
		{"%x", "90"},
		{"%X", "0"},
		{"%r", "Hall"},
		{"%%", "%"},
	}

	for _, tt := range tests {
		t.Run(tt.token, func(t *testing.T) {
			ch.PCData.Prompt = tt.token
			got := FormatPrompt(ch)
			if got != tt.want {
				t.Errorf("FormatPrompt(%q) = %q, want %q", tt.token, got, tt.want)
			}
		})
	}
}
