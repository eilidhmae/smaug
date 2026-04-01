package mudprog

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

func TestTranslate_Basic(t *testing.T) {
	mob := &types.CharData{Name: "Guard", Sex: types.SEX_MALE}
	actor := &types.CharData{Name: "Gandalf", Sex: types.SEX_MALE}
	victim := &types.CharData{Name: "Orc", Sex: types.SEX_NEUTRAL}

	tests := []struct {
		input    string
		expected string
	}{
		{"$n attacks $t!", "Gandalf attacks Orc!"},
		{"$i says hello", "Guard says hello"},
		{"$e draws $s sword", "he draws his sword"},
		{"$E growls", "it growls"},
		{"no variables", "no variables"},
		{"", ""},
	}

	for _, tc := range tests {
		result := Translate(mob, actor, nil, victim, nil, tc.input)
		if result != tc.expected {
			t.Errorf("Translate(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestTranslate_NilChars(t *testing.T) {
	result := Translate(nil, nil, nil, nil, nil, "$n does something")
	if result != "someone does something" {
		t.Errorf("got %q", result)
	}
}

func TestTranslate_Objects(t *testing.T) {
	obj := &types.ObjData{Name: "sword", ShortDescr: "a magic sword"}
	result := Translate(nil, nil, obj, nil, nil, "$o glows")
	if result != "a magic sword glows" {
		t.Errorf("got %q", result)
	}
}
