package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

func TestCheckSocial_NotFound(t *testing.T) {
	w := setupWizWorld()
	_ = w
	ch, client := makeTestChar("Tester")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 9100, Name: "Test"}

	result := CheckSocial(ch, "nonexistentsocial", "")
	if result {
		t.Error("should return false for unknown social")
	}
}

func TestCheckSocial_NoTarget(t *testing.T) {
	w := setupWizWorld()
	w.Socials = append(w.Socials, &types.SocialType{
		Name:         "smile",
		CharNoArg:    "You smile happily.",
		OthersNoArg:  "$n smiles happily.",
		CharFound:    "You smile at $N.",
		OthersFound:  "$n smiles at $N.",
		VictFound:    "$n smiles at you.",
		CharAuto:     "You smile at yourself.",
		OthersAuto:   "$n smiles at $mself.",
	})

	room := &types.RoomIndexData{Vnum: 9101, Name: "Test"}

	ch, chClient := makeTestChar("Actor")
	defer chClient.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	listener, listenerClient := makeTestChar("Watcher")
	defer listenerClient.Close()
	listener.InRoom = room
	room.People = append(room.People, listener)

	result := CheckSocial(ch, "smile", "")
	if !result {
		t.Error("should return true for found social")
	}

	chOut := readOutput(ch, chClient)
	if !strings.Contains(chOut, "smile happily") {
		t.Errorf("sender should see char_no_arg, got: %q", chOut)
	}

	listenerOut := readOutput(listener, listenerClient)
	if !strings.Contains(listenerOut, "Actor smiles happily") {
		t.Errorf("listener should see others_no_arg, got: %q", listenerOut)
	}
}

func TestCheckSocial_WithTarget(t *testing.T) {
	w := setupWizWorld()
	w.Socials = append(w.Socials, &types.SocialType{
		Name:        "bow",
		CharNoArg:   "You bow deeply.",
		OthersNoArg: "$n bows deeply.",
		CharFound:   "You bow before $N.",
		OthersFound: "$n bows before $N.",
		VictFound:   "$n bows before you.",
		CharAuto:    "You bow to yourself.",
		OthersAuto:  "$n bows to $mself.",
	})

	room := &types.RoomIndexData{Vnum: 9102, Name: "Test"}

	ch, chClient := makeTestChar("Actor")
	defer chClient.Close()
	handler.CharToRoom(ch, room)

	victim, victimClient := makeTestChar("Target")
	defer victimClient.Close()
	handler.CharToRoom(victim, room)

	result := CheckSocial(ch, "bow", "Target")
	if !result {
		t.Error("should return true")
	}

	chOut := readOutput(ch, chClient)
	if !strings.Contains(chOut, "bow before Target") {
		t.Errorf("sender should see targeted message, got: %q", chOut)
	}

	victimOut := readOutput(victim, victimClient)
	if !strings.Contains(victimOut, "bows before you") {
		t.Errorf("victim should see vict_found, got: %q", victimOut)
	}
}

func TestCheckSocial_SelfTarget(t *testing.T) {
	w := setupWizWorld()
	w.Socials = append(w.Socials, &types.SocialType{
		Name:        "laugh",
		CharNoArg:   "You laugh.",
		OthersNoArg: "$n laughs.",
		CharFound:   "You laugh at $N.",
		OthersFound: "$n laughs at $N.",
		VictFound:   "$n laughs at you.",
		CharAuto:    "You laugh at yourself.",
		OthersAuto:  "$n laughs at $mself.",
	})

	room := &types.RoomIndexData{Vnum: 9103, Name: "Test"}
	ch, chClient := makeTestChar("Actor")
	defer chClient.Close()
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	result := CheckSocial(ch, "laugh", "Actor")
	if !result {
		t.Error("should return true for self-targeted social")
	}

	chOut := readOutput(ch, chClient)
	if !strings.Contains(chOut, "laugh at yourself") {
		t.Errorf("expected self-targeted message, got: %q", chOut)
	}
}

func TestCheckSocial_TargetNotFound(t *testing.T) {
	w := setupWizWorld()
	w.Socials = append(w.Socials, &types.SocialType{
		Name:      "wave",
		CharNoArg: "You wave.",
	})

	room := &types.RoomIndexData{Vnum: 9104, Name: "Test"}
	ch, chClient := makeTestChar("Actor")
	defer chClient.Close()
	handler.CharToRoom(ch, room)

	result := CheckSocial(ch, "wave", "nonexistent")
	if !result {
		t.Error("should return true (social found, target not)")
	}

	chOut := readOutput(ch, chClient)
	if !strings.Contains(chOut, "aren't here") {
		t.Errorf("expected 'aren't here', got: %q", chOut)
	}
}

func TestCheckSocial_NilWorldRef(t *testing.T) {
	saved := WorldRef
	WorldRef = nil
	defer func() { WorldRef = saved }()

	ch, client := makeTestChar("Tester")
	defer client.Close()

	result := CheckSocial(ch, "smile", "")
	if result {
		t.Error("should return false when WorldRef is nil")
	}
}

func TestPronoun(t *testing.T) {
	tests := []struct {
		sex      int
		expected string
	}{
		{types.SEX_MALE, "he"},
		{types.SEX_FEMALE, "she"},
		{types.SEX_NEUTRAL, "it"},
	}

	for _, tc := range tests {
		ch := &types.CharData{Sex: tc.sex}
		got := pronoun(ch, "he", "she", "it")
		if got != tc.expected {
			t.Errorf("sex %d: expected %q, got %q", tc.sex, tc.expected, got)
		}
	}
}

func TestSocialSub(t *testing.T) {
	ch := &types.CharData{Name: "Alice", Sex: types.SEX_FEMALE}
	victim := &types.CharData{Name: "Bob", Sex: types.SEX_MALE}

	result := socialSub("$n smiles at $N. $e waves.", ch, victim)
	if !strings.Contains(result, "Alice smiles at Bob") {
		t.Errorf("expected substituted names, got: %q", result)
	}
	if !strings.Contains(result, "she waves") {
		t.Errorf("expected pronoun substitution, got: %q", result)
	}
}

func TestSocialSub_Empty(t *testing.T) {
	ch := &types.CharData{Name: "Test"}
	result := socialSub("", ch, nil)
	if result != "" {
		t.Errorf("expected empty string, got: %q", result)
	}
}
