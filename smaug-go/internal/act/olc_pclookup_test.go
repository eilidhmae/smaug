package act

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// Tests for WorldPcLookup mirror the game-package worldPcLookup tests
// (game/medit_parse_test.go) — both implementations share the same
// shape and must behave identically. Plan §G3 / §A4.

func TestWorldPcLookup_FindsConnectedPc(t *testing.T) {
	w := world.New("/tmp/test")
	prev := WorldRef
	WorldRef = w
	defer func() { WorldRef = prev }()

	pc := &types.CharData{Name: "Eilidh"}
	d := &types.DescriptorData{Connected: int(types.CON_PLAYING), Character: pc}
	w.Descriptors = append(w.Descriptors, d)

	if got := WorldPcLookup("Eilidh"); got != pc {
		t.Errorf("WorldPcLookup(Eilidh) = %v, want %v", got, pc)
	}
}

func TestWorldPcLookup_IgnoresLinkdead(t *testing.T) {
	w := world.New("/tmp/test")
	prev := WorldRef
	WorldRef = w
	defer func() { WorldRef = prev }()

	pc := &types.CharData{Name: "Eilidh"}
	d := &types.DescriptorData{Connected: int(types.CON_GET_NAME), Character: pc}
	w.Descriptors = append(w.Descriptors, d)

	if got := WorldPcLookup("Eilidh"); got != nil {
		t.Errorf("linkdead descriptor should not match; got %v", got)
	}
}

func TestWorldPcLookup_IgnoresNilCharacter(t *testing.T) {
	w := world.New("/tmp/test")
	prev := WorldRef
	WorldRef = w
	defer func() { WorldRef = prev }()

	d := &types.DescriptorData{Connected: int(types.CON_PLAYING), Character: nil}
	w.Descriptors = append(w.Descriptors, d)

	if got := WorldPcLookup("Eilidh"); got != nil {
		t.Errorf("nil-Character descriptor should not match; got %v", got)
	}
}

func TestWorldPcLookup_CaseInsensitive(t *testing.T) {
	w := world.New("/tmp/test")
	prev := WorldRef
	WorldRef = w
	defer func() { WorldRef = prev }()

	pc := &types.CharData{Name: "Eilidh"}
	d := &types.DescriptorData{Connected: int(types.CON_PLAYING), Character: pc}
	w.Descriptors = append(w.Descriptors, d)

	if got := WorldPcLookup("eilidh"); got != pc {
		t.Errorf("lowercase lookup failed; got %v", got)
	}
	if got := WorldPcLookup("EILIDH"); got != pc {
		t.Errorf("uppercase lookup failed; got %v", got)
	}
}

func TestWorldPcLookup_NoMatch(t *testing.T) {
	w := world.New("/tmp/test")
	prev := WorldRef
	WorldRef = w
	defer func() { WorldRef = prev }()

	pc := &types.CharData{Name: "Eilidh"}
	d := &types.DescriptorData{Connected: int(types.CON_PLAYING), Character: pc}
	w.Descriptors = append(w.Descriptors, d)

	if got := WorldPcLookup("Bob"); got != nil {
		t.Errorf("non-matching name should return nil; got %v", got)
	}
}

func TestWorldPcLookup_NilWorld(t *testing.T) {
	prev := WorldRef
	WorldRef = nil
	defer func() { WorldRef = prev }()

	if got := WorldPcLookup("Eilidh"); got != nil {
		t.Errorf("nil WorldRef should return nil; got %v", got)
	}
}
