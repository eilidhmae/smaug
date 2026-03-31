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
