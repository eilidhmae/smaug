package persist

import (
	"bytes"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

// TestSaveLoadPlayer_AffectRemovalMaintainsStatInvariant pins the
// Go architectural invariant that replaces C's update_aris rebuild
// pass: ch.Armor / Hitroll / Damroll stay correct through
// AffectToChar → AffectRemove → SavePlayer → LoadPlayer. C needs
// update_aris because it rebuilds affected_by/RIS from scratch after
// every affect change and saves base-minus-equipment stats (C
// src/handler.c:1221-1305 + src/save.c:213 de_equip_char); Go
// instead mutates ch.Armor / Hitroll / Damroll incrementally via
// handler.AffectModify (internal/handler/handler.go:386-455) on each
// AffectToChar / AffectRemove, and persists the computed values
// verbatim in the save file. This test guards against a future
// regression that might drift stats from the affect list.
//
// Scenario: player has base Armor = -10, Hitroll = 3, Damroll = 5.
// A sanctuary-style buff with Location=APPLY_AC,Modifier=-20 is applied
// (Armor should become -30). The buff expires and is removed via
// AffectRemove (Armor should restore to -10). Save, load, assert
// round-trip preserves Armor = -10 and Hitroll / Damroll unchanged.
//
// Mutation: flipping AffectRemove's AffectModify(ch, aff, false) call
// to fAdd=true would leave Armor at -50 instead of -10. This test
// would catch that drift because the saved value is written verbatim
// and loaded verbatim.
func TestSaveLoadPlayer_AffectRemovalMaintainsStatInvariant(t *testing.T) {
	ch := &types.CharData{
		Name:    "Affinv",
		Level:   10,
		Hit:     100,
		MaxHit:  100,
		Mana:    50,
		MaxMana: 50,
		Move:    80,
		MaxMove: 80,
		Hitroll: 3,
		Damroll: 5,
		Armor:   -10,
		PermStr: 13, PermInt: 13, PermWis: 13, PermDex: 13,
		PermCon: 13, PermCha: 13, PermLck: 13,
		Position: types.POS_STANDING,
		PCData:   &types.PCData{Pwd: "pw", PagerLen: 24},
	}

	// Apply a sanctuary-style affect that improves AC by 20.
	buff := &types.AffectData{
		Type:     1,
		Duration: 10,
		Location: types.APPLY_AC,
		Modifier: -20, // SMAUG convention: negative = better AC
	}
	handler.AffectToChar(ch, buff)
	if ch.Armor != -30 {
		t.Fatalf("after AffectToChar: Armor = %d, want -30 (base -10 + buff -20)", ch.Armor)
	}
	if ch.Hitroll != 3 || ch.Damroll != 5 {
		t.Fatalf("after AffectToChar: Hitroll/Damroll perturbed: %d/%d (want 3/5)",
			ch.Hitroll, ch.Damroll)
	}
	if len(ch.Affects) != 1 {
		t.Fatalf("after AffectToChar: Affects len = %d, want 1", len(ch.Affects))
	}

	// Expire + remove the affect (AffectRemove requires the pointer
	// stored on ch, not the original `buff` — AffectToChar copies).
	stored := ch.Affects[0]
	stored.Duration = 0
	handler.AffectRemove(ch, stored)
	if ch.Armor != -10 {
		t.Fatalf("after AffectRemove: Armor = %d, want -10 (modifier should be reversed)", ch.Armor)
	}
	if len(ch.Affects) != 0 {
		t.Fatalf("after AffectRemove: Affects len = %d, want 0", len(ch.Affects))
	}

	// Save + load round-trip.
	var buf bytes.Buffer
	if err := SavePlayer(&buf, ch); err != nil {
		t.Fatalf("SavePlayer: %v", err)
	}
	loaded, err := LoadPlayer(bytes.NewReader(buf.Bytes()), "Affinv")
	if err != nil {
		t.Fatalf("LoadPlayer: %v", err)
	}
	if loaded.Armor != -10 {
		t.Errorf("round-trip: loaded.Armor = %d, want -10", loaded.Armor)
	}
	if loaded.Hitroll != 3 {
		t.Errorf("round-trip: loaded.Hitroll = %d, want 3", loaded.Hitroll)
	}
	if loaded.Damroll != 5 {
		t.Errorf("round-trip: loaded.Damroll = %d, want 5", loaded.Damroll)
	}
	if len(loaded.Affects) != 0 {
		t.Errorf("round-trip: loaded.Affects len = %d, want 0 (affect was removed before save)", len(loaded.Affects))
	}
}
