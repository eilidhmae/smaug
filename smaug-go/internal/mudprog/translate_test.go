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

func TestTranslate_AllVariables(t *testing.T) {
	mob := &types.CharData{Name: "Guard", ShortDescr: "the guard", Sex: types.SEX_MALE}
	actor := &types.CharData{Name: "Gandalf", ShortDescr: "Gandalf the Grey", Sex: types.SEX_MALE}
	victim := &types.CharData{Name: "Orc", ShortDescr: "an ugly orc", Sex: types.SEX_FEMALE}
	obj := &types.ObjData{Name: "sword", ShortDescr: "a magic sword"}
	target := &types.ObjData{Name: "shield", ShortDescr: "a wooden shield"}

	tests := []struct {
		input    string
		expected string
	}{
		// Character names
		{"$i", "Guard"},
		{"$n", "Gandalf"},
		{"$t", "Orc"},
		// Short descriptions
		{"$I", "the guard"},
		{"$N", "Gandalf the Grey"},
		{"$T", "an ugly orc"},
		// Object short descriptions
		{"$o", "a magic sword"},
		{"$p", "a wooden shield"},
		// Object names
		{"$O", "sword"},
		{"$P", "shield"},
		// Actor pronouns
		{"$e", "he"},
		{"$m", "him"},
		{"$s", "his"},
		// Victim pronouns
		{"$E", "she"},
		{"$M", "her"},
		{"$S", "her"},
		// Mob pronouns
		{"$j", "he"},
		{"$k", "him"},
		{"$l", "his"},
		// Unknown variable - pass through
		{"$z", "$z"},
	}
	for _, tc := range tests {
		result := Translate(mob, actor, obj, victim, target, tc.input)
		if result != tc.expected {
			t.Errorf("Translate(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestTranslate_RoomPeople(t *testing.T) {
	roomChar := &types.CharData{Name: "Bystander"}
	room := &types.RoomIndexData{Vnum: 100, People: []*types.CharData{roomChar}}
	mob := &types.CharData{Name: "Guard", InRoom: room}

	result := Translate(mob, nil, nil, nil, nil, "$r watches")
	if result != "Bystander watches" {
		t.Errorf("$r should resolve to first person in room, got %q", result)
	}

	// Empty room
	mob2 := &types.CharData{Name: "Guard2", InRoom: &types.RoomIndexData{Vnum: 200}}
	result = Translate(mob2, nil, nil, nil, nil, "$r watches")
	if result != "someone watches" {
		t.Errorf("$r with empty room should be 'someone', got %q", result)
	}

	// Nil mob
	result = Translate(nil, nil, nil, nil, nil, "$r watches")
	if result != "someone watches" {
		t.Errorf("$r with nil mob should be 'someone', got %q", result)
	}
}

func TestTranslate_Pronouns_Female(t *testing.T) {
	female := &types.CharData{Name: "Queen", Sex: types.SEX_FEMALE}
	result := Translate(nil, female, nil, nil, nil, "$e/$m/$s")
	if result != "she/her/her" {
		t.Errorf("female pronouns: got %q", result)
	}
}

func TestTranslate_Pronouns_Neutral(t *testing.T) {
	neutral := &types.CharData{Name: "Golem", Sex: types.SEX_NEUTRAL}
	result := Translate(nil, neutral, nil, nil, nil, "$e/$m/$s")
	if result != "it/it/its" {
		t.Errorf("neutral pronouns: got %q", result)
	}
}

func TestTranslate_Pronouns_Nil(t *testing.T) {
	result := Translate(nil, nil, nil, nil, nil, "$e/$m/$s/$E/$M/$S")
	if result != "it/it/its/it/it/its" {
		t.Errorf("nil pronouns: got %q", result)
	}
}

func TestTranslate_MobPronouns(t *testing.T) {
	mob := &types.CharData{Name: "Guard", Sex: types.SEX_FEMALE}
	result := Translate(mob, nil, nil, nil, nil, "$j/$k/$l")
	if result != "she/her/her" {
		t.Errorf("mob female pronouns: got %q", result)
	}
}

func TestTranslate_NilObjects(t *testing.T) {
	result := Translate(nil, nil, nil, nil, nil, "$o/$O/$p/$P")
	if result != "something/something/something/something" {
		t.Errorf("nil objects: got %q", result)
	}
}

func TestTranslate_DollarAtEnd(t *testing.T) {
	// $ at end of string with no following char
	result := Translate(nil, nil, nil, nil, nil, "test$")
	if result != "test$" {
		t.Errorf("trailing dollar: got %q", result)
	}
}

func TestTranslate_CharShort_FallbackToName(t *testing.T) {
	// When ShortDescr is empty, charShort falls back to Name
	ch := &types.CharData{Name: "TestName", ShortDescr: ""}
	result := Translate(nil, ch, nil, nil, nil, "$N")
	if result != "TestName" {
		t.Errorf("charShort should fallback to Name when ShortDescr empty, got %q", result)
	}
}

func TestTranslate_NoSubstitution(t *testing.T) {
	// Text without $ should pass through unchanged
	result := Translate(nil, nil, nil, nil, nil, "hello world, no variables here")
	if result != "hello world, no variables here" {
		t.Errorf("got %q", result)
	}
}

func TestCharName_Nil(t *testing.T) {
	if charName(nil) != "someone" {
		t.Error("charName(nil) should return 'someone'")
	}
}

func TestCharShort_Nil(t *testing.T) {
	if charShort(nil) != "someone" {
		t.Error("charShort(nil) should return 'someone'")
	}
}

func TestObjShort_Nil(t *testing.T) {
	if objShort(nil) != "something" {
		t.Error("objShort(nil) should return 'something'")
	}
}

func TestObjName_Nil(t *testing.T) {
	if objName(nil) != "something" {
		t.Error("objName(nil) should return 'something'")
	}
}

func TestHeShe(t *testing.T) {
	tests := []struct {
		sex      int
		expected string
	}{
		{types.SEX_MALE, "he"},
		{types.SEX_FEMALE, "she"},
		{types.SEX_NEUTRAL, "it"},
		{99, "it"}, // unknown defaults to it
	}
	for _, tc := range tests {
		ch := &types.CharData{Sex: tc.sex}
		if heShe(ch) != tc.expected {
			t.Errorf("heShe(sex=%d) = %q, want %q", tc.sex, heShe(ch), tc.expected)
		}
	}
	if heShe(nil) != "it" {
		t.Error("heShe(nil) should return 'it'")
	}
}

func TestHimHer(t *testing.T) {
	tests := []struct {
		sex      int
		expected string
	}{
		{types.SEX_MALE, "him"},
		{types.SEX_FEMALE, "her"},
		{types.SEX_NEUTRAL, "it"},
	}
	for _, tc := range tests {
		ch := &types.CharData{Sex: tc.sex}
		if himHer(ch) != tc.expected {
			t.Errorf("himHer(sex=%d) = %q, want %q", tc.sex, himHer(ch), tc.expected)
		}
	}
	if himHer(nil) != "it" {
		t.Error("himHer(nil) should return 'it'")
	}
}

func TestHisHer(t *testing.T) {
	tests := []struct {
		sex      int
		expected string
	}{
		{types.SEX_MALE, "his"},
		{types.SEX_FEMALE, "her"},
		{types.SEX_NEUTRAL, "its"},
	}
	for _, tc := range tests {
		ch := &types.CharData{Sex: tc.sex}
		if hisHer(ch) != tc.expected {
			t.Errorf("hisHer(sex=%d) = %q, want %q", tc.sex, hisHer(ch), tc.expected)
		}
	}
	if hisHer(nil) != "its" {
		t.Error("hisHer(nil) should return 'its'")
	}
}
