package persist

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

func TestLoadPlayer(t *testing.T) {
	path := filepath.Join("testdata", "Testchar")
	f, err := os.Open(path)
	if err != nil {
		t.Skipf("testdata not found: %s", path)
	}
	defer f.Close()

	ch, err := LoadPlayer(f, "Testchar")
	if err != nil {
		t.Fatalf("LoadPlayer failed: %v", err)
	}

	if ch.Name != "Testchar" {
		t.Errorf("Name = %q, want %q", ch.Name, "Testchar")
	}
	if ch.Sex != 1 {
		t.Errorf("Sex = %d, want 1", ch.Sex)
	}
	if ch.Class != 3 {
		t.Errorf("Class = %d, want 3", ch.Class)
	}
	if ch.Race != 0 {
		t.Errorf("Race = %d, want 0", ch.Race)
	}
	if ch.Level != 10 {
		t.Errorf("Level = %d, want 10", ch.Level)
	}
	if ch.Hit != 100 {
		t.Errorf("Hit = %d, want 100", ch.Hit)
	}
	if ch.MaxHit != 100 {
		t.Errorf("MaxHit = %d, want 100", ch.MaxHit)
	}
	if ch.Mana != 50 {
		t.Errorf("Mana = %d, want 50", ch.Mana)
	}
	if ch.MaxMana != 50 {
		t.Errorf("MaxMana = %d, want 50", ch.MaxMana)
	}
	if ch.Move != 80 {
		t.Errorf("Move = %d, want 80", ch.Move)
	}
	if ch.MaxMove != 80 {
		t.Errorf("MaxMove = %d, want 80", ch.MaxMove)
	}
	if ch.Gold != 500 {
		t.Errorf("Gold = %d, want 500", ch.Gold)
	}
	if ch.Exp != 5000 {
		t.Errorf("Exp = %d, want 5000", ch.Exp)
	}
	if ch.Hitroll != 3 {
		t.Errorf("Hitroll = %d, want 3", ch.Hitroll)
	}
	if ch.Damroll != 5 {
		t.Errorf("Damroll = %d, want 5", ch.Damroll)
	}
	if ch.Armor != 50 {
		t.Errorf("Armor = %d, want 50", ch.Armor)
	}
	if ch.PermStr != 15 {
		t.Errorf("PermStr = %d, want 15", ch.PermStr)
	}
	if ch.PermLck != 13 {
		t.Errorf("PermLck = %d, want 13", ch.PermLck)
	}
	if ch.PCData == nil {
		t.Fatal("PCData is nil")
	}
	if ch.PCData.Pwd != "testpass" {
		t.Errorf("Pwd = %q, want %q", ch.PCData.Pwd, "testpass")
	}
	if ch.PCData.Title != "the Brave" {
		t.Errorf("Title = %q, want %q", ch.PCData.Title, "the Brave")
	}
	if ch.PCData.MKills != 42 {
		t.Errorf("MKills = %d, want 42", ch.PCData.MKills)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	// Create a character
	ch := &types.CharData{
		Name:      "Roundtrip",
		Sex:       2,
		Class:     1,
		Race:      2,
		Level:     5,
		Hit:       75,
		MaxHit:    80,
		Mana:      30,
		MaxMana:   40,
		Move:      60,
		MaxMove:   70,
		Gold:      1000,
		Exp:       2500,
		Hitroll:   2,
		Damroll:   3,
		Armor:     30,
		Alignment: 500,
		PermStr:   14,
		PermInt:   16,
		PermWis:   12,
		PermDex:   15,
		PermCon:   13,
		PermCha:   10,
		PermLck:   11,
		Position:  types.POS_STANDING,
		Style:     types.STYLE_FIGHTING,
		Height:    68,
		Weight:    160,
		PCData: &types.PCData{
			Pwd:      "secret",
			Title:    "the Bold",
			Prompt:   "<%hhp> ",
			PKills:   1,
			PDeaths:  2,
			MKills:   100,
			MDeaths:  10,
			PagerLen: 24,
			Flags:    2,
		},
	}

	// Save to buffer
	var buf bytes.Buffer
	err := SavePlayer(&buf, ch)
	if err != nil {
		t.Fatalf("SavePlayer failed: %v", err)
	}

	// Load back
	loaded, err := LoadPlayer(bytes.NewReader(buf.Bytes()), "Roundtrip")
	if err != nil {
		t.Fatalf("LoadPlayer failed: %v", err)
	}

	// Verify key fields
	if loaded.Name != ch.Name {
		t.Errorf("Name = %q, want %q", loaded.Name, ch.Name)
	}
	if loaded.Level != ch.Level {
		t.Errorf("Level = %d, want %d", loaded.Level, ch.Level)
	}
	if loaded.Hit != ch.Hit {
		t.Errorf("Hit = %d, want %d", loaded.Hit, ch.Hit)
	}
	if loaded.MaxHit != ch.MaxHit {
		t.Errorf("MaxHit = %d, want %d", loaded.MaxHit, ch.MaxHit)
	}
	if loaded.Gold != ch.Gold {
		t.Errorf("Gold = %d, want %d", loaded.Gold, ch.Gold)
	}
	if loaded.Exp != ch.Exp {
		t.Errorf("Exp = %d, want %d", loaded.Exp, ch.Exp)
	}
	if loaded.PermStr != ch.PermStr {
		t.Errorf("PermStr = %d, want %d", loaded.PermStr, ch.PermStr)
	}
	if loaded.PermLck != ch.PermLck {
		t.Errorf("PermLck = %d, want %d", loaded.PermLck, ch.PermLck)
	}
	if loaded.Alignment != ch.Alignment {
		t.Errorf("Alignment = %d, want %d", loaded.Alignment, ch.Alignment)
	}
	if loaded.PCData == nil {
		t.Fatal("PCData is nil after round-trip")
	}
	if loaded.PCData.Pwd != ch.PCData.Pwd {
		t.Errorf("Pwd = %q, want %q", loaded.PCData.Pwd, ch.PCData.Pwd)
	}
	if loaded.PCData.Title != ch.PCData.Title {
		t.Errorf("Title = %q, want %q", loaded.PCData.Title, ch.PCData.Title)
	}
	if loaded.PCData.MKills != ch.PCData.MKills {
		t.Errorf("MKills = %d, want %d", loaded.PCData.MKills, ch.PCData.MKills)
	}
}
