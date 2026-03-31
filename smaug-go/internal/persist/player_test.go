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
	if ch.Position != types.POS_STANDING {
		t.Errorf("Position = %d, want POS_STANDING (%d)", ch.Position, types.POS_STANDING)
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
	if loaded.Position != ch.Position {
		t.Errorf("Position = %d, want %d", loaded.Position, ch.Position)
	}
	if loaded.Style != ch.Style {
		t.Errorf("Style = %d, want %d", loaded.Style, ch.Style)
	}
	if loaded.Height != ch.Height {
		t.Errorf("Height = %d, want %d", loaded.Height, ch.Height)
	}
	if loaded.Weight != ch.Weight {
		t.Errorf("Weight = %d, want %d", loaded.Weight, ch.Weight)
	}
	// Verify on-disk format has Position + 100 (C compatibility)
	saved := buf.String()
	if !bytes.Contains([]byte(saved), []byte("Position   112")) {
		t.Errorf("saved file should contain 'Position   112' (POS_STANDING=12 + 100), got:\n%s",
			saved)
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

func TestSaveLoadAffects(t *testing.T) {
	ch := &types.CharData{
		Name:    "Afftest",
		Level:   10,
		Hit:     100,
		MaxHit:  100,
		Mana:    50,
		MaxMana: 50,
		Move:    80,
		MaxMove: 80,
		PermStr: 15, PermInt: 13, PermWis: 12, PermDex: 14,
		PermCon: 13, PermCha: 11, PermLck: 13,
		Position: types.POS_STANDING,
		PCData:   &types.PCData{Pwd: "pass", PagerLen: 24},
	}

	// Add two affects
	ch.Affects = []*types.AffectData{
		{Type: 5, Duration: 24, Location: types.APPLY_AC, Modifier: -20},
		{Type: 8, Duration: 10, Location: types.APPLY_STR, Modifier: -2},
	}
	ch.Affects[1].BitVector.Set(types.AFF_POISON)

	var buf bytes.Buffer
	if err := SavePlayer(&buf, ch); err != nil {
		t.Fatalf("SavePlayer failed: %v", err)
	}

	// Verify raw output contains Affect lines
	saved := buf.String()
	if !bytes.Contains([]byte(saved), []byte("Affect ")) {
		t.Error("saved file should contain 'Affect' lines")
	}

	loaded, err := LoadPlayer(bytes.NewReader(buf.Bytes()), "Afftest")
	if err != nil {
		t.Fatalf("LoadPlayer failed: %v", err)
	}

	if len(loaded.Affects) != 2 {
		t.Fatalf("Affects = %d, want 2", len(loaded.Affects))
	}

	aff := loaded.Affects[0]
	if aff.Type != 5 || aff.Duration != 24 || aff.Location != types.APPLY_AC || aff.Modifier != -20 {
		t.Errorf("Affect[0] = {%d %d %d %d}, want {5 24 %d -20}",
			aff.Type, aff.Duration, aff.Location, aff.Modifier, types.APPLY_AC)
	}

	aff1 := loaded.Affects[1]
	if aff1.Type != 8 || aff1.Duration != 10 || aff1.Location != types.APPLY_STR || aff1.Modifier != -2 {
		t.Errorf("Affect[1] type/dur/loc/mod wrong")
	}
	if !aff1.BitVector.IsSet(types.AFF_POISON) {
		t.Error("Affect[1] should have AFF_POISON set")
	}
}

func TestSaveLoadObjects(t *testing.T) {
	ch := &types.CharData{
		Name:    "Objtest",
		Level:   10,
		Hit:     100,
		MaxHit:  100,
		Mana:    50,
		MaxMana: 50,
		Move:    80,
		MaxMove: 80,
		PermStr: 15, PermInt: 13, PermWis: 12, PermDex: 14,
		PermCon: 13, PermCha: 11, PermLck: 13,
		Position: types.POS_STANDING,
		PCData:   &types.PCData{Pwd: "pass", PagerLen: 24},
	}

	// Sword in inventory (not equipped)
	sword := &types.ObjData{
		Name:       "a short sword",
		ShortDescr: "a short sword",
		ItemType:   types.ITEM_WEAPON,
		WearLoc:    types.WEAR_NONE,
		Weight:     5,
		Level:      3,
		Value:      [6]int{0, 4, 6, 0, 0, 0},
		IndexData:  &types.ObjIndexData{Vnum: 1001},
	}

	// Shield equipped
	shield := &types.ObjData{
		Name:       "a wooden shield",
		ShortDescr: "a wooden shield",
		ItemType:   types.ITEM_ARMOR,
		WearLoc:    types.WEAR_SHIELD,
		Weight:     8,
		Level:      2,
		Value:      [6]int{5, 0, 0, 0, 0, 0},
		IndexData:  &types.ObjIndexData{Vnum: 1002},
	}

	ch.Carrying = []*types.ObjData{sword, shield}

	var buf bytes.Buffer
	if err := SavePlayer(&buf, ch); err != nil {
		t.Fatalf("SavePlayer failed: %v", err)
	}

	saved := buf.String()
	if !bytes.Contains([]byte(saved), []byte("#OBJECT")) {
		t.Error("saved file should contain '#OBJECT' sections")
	}
	if !bytes.Contains([]byte(saved), []byte("Vnum         1001")) {
		t.Error("saved file should contain sword vnum 1001")
	}
	if !bytes.Contains([]byte(saved), []byte("Vnum         1002")) {
		t.Error("saved file should contain shield vnum 1002")
	}
	if !bytes.Contains([]byte(saved), []byte("WearLoc      11")) {
		t.Errorf("saved file should contain 'WearLoc      11' for WEAR_SHIELD")
	}

	// Load back — objects need ObjIndex lookup so we provide a resolver
	loaded, err := LoadPlayerWithWorld(bytes.NewReader(buf.Bytes()), "Objtest", func(vnum int) *types.ObjIndexData {
		switch vnum {
		case 1001:
			return &types.ObjIndexData{
				Vnum: 1001, Name: "a short sword", ShortDescr: "a short sword",
				ItemType: types.ITEM_WEAPON, Weight: 5, Level: 3,
				Value: [6]int{0, 4, 6, 0, 0, 0},
			}
		case 1002:
			return &types.ObjIndexData{
				Vnum: 1002, Name: "a wooden shield", ShortDescr: "a wooden shield",
				ItemType: types.ITEM_ARMOR, Weight: 8, Level: 2,
				Value: [6]int{5, 0, 0, 0, 0, 0},
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("LoadPlayerWithWorld failed: %v", err)
	}

	if len(loaded.Carrying) != 2 {
		t.Fatalf("Carrying = %d, want 2", len(loaded.Carrying))
	}

	// Check sword
	obj0 := loaded.Carrying[0]
	if obj0.IndexData == nil || obj0.IndexData.Vnum != 1001 {
		t.Errorf("obj[0] vnum wrong, got %v", obj0.IndexData)
	}
	if obj0.WearLoc != types.WEAR_NONE {
		t.Errorf("obj[0] WearLoc = %d, want WEAR_NONE (%d)", obj0.WearLoc, types.WEAR_NONE)
	}

	// Check shield
	obj1 := loaded.Carrying[1]
	if obj1.IndexData == nil || obj1.IndexData.Vnum != 1002 {
		t.Errorf("obj[1] vnum wrong, got %v", obj1.IndexData)
	}
	if obj1.WearLoc != types.WEAR_SHIELD {
		t.Errorf("obj[1] WearLoc = %d, want WEAR_SHIELD (%d)", obj1.WearLoc, types.WEAR_SHIELD)
	}
}

func TestSaveLoadObjectInContainer(t *testing.T) {
	ch := &types.CharData{
		Name:    "Containertest",
		Level:   10,
		Hit:     100, MaxHit: 100, Mana: 50, MaxMana: 50, Move: 80, MaxMove: 80,
		PermStr: 15, PermInt: 13, PermWis: 12, PermDex: 14,
		PermCon: 13, PermCha: 11, PermLck: 13,
		Position: types.POS_STANDING,
		PCData:   &types.PCData{Pwd: "pass", PagerLen: 24},
	}

	gem := &types.ObjData{
		Name:       "a ruby gem",
		ShortDescr: "a ruby gem",
		ItemType:   types.ITEM_TREASURE,
		WearLoc:    types.WEAR_NONE,
		Weight:     1,
		IndexData:  &types.ObjIndexData{Vnum: 2001},
	}

	bag := &types.ObjData{
		Name:       "a leather bag",
		ShortDescr: "a leather bag",
		ItemType:   types.ITEM_CONTAINER,
		WearLoc:    types.WEAR_NONE,
		Weight:     3,
		IndexData:  &types.ObjIndexData{Vnum: 2002},
		Contents:   []*types.ObjData{gem},
	}
	gem.InObj = bag

	ch.Carrying = []*types.ObjData{bag}

	var buf bytes.Buffer
	if err := SavePlayer(&buf, ch); err != nil {
		t.Fatalf("SavePlayer failed: %v", err)
	}

	saved := buf.String()
	// Should have two #OBJECT sections (bag at nest 0, gem at nest 1)
	if count := bytes.Count([]byte(saved), []byte("#OBJECT")); count != 2 {
		t.Errorf("#OBJECT count = %d, want 2\n%s", count, saved)
	}

	loaded, err := LoadPlayerWithWorld(bytes.NewReader(buf.Bytes()), "Containertest", func(vnum int) *types.ObjIndexData {
		switch vnum {
		case 2001:
			return &types.ObjIndexData{Vnum: 2001, Name: "a ruby gem", ShortDescr: "a ruby gem", ItemType: types.ITEM_TREASURE, Weight: 1}
		case 2002:
			return &types.ObjIndexData{Vnum: 2002, Name: "a leather bag", ShortDescr: "a leather bag", ItemType: types.ITEM_CONTAINER, Weight: 3}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("LoadPlayerWithWorld failed: %v", err)
	}

	if len(loaded.Carrying) != 1 {
		t.Fatalf("Carrying = %d, want 1 (bag)", len(loaded.Carrying))
	}
	loadedBag := loaded.Carrying[0]
	if loadedBag.IndexData == nil || loadedBag.IndexData.Vnum != 2002 {
		t.Error("bag vnum wrong")
	}
	if len(loadedBag.Contents) != 1 {
		t.Fatalf("bag Contents = %d, want 1 (gem)", len(loadedBag.Contents))
	}
	loadedGem := loadedBag.Contents[0]
	if loadedGem.IndexData == nil || loadedGem.IndexData.Vnum != 2001 {
		t.Error("gem vnum wrong")
	}
	if loadedGem.InObj != loadedBag {
		t.Error("gem.InObj should point to bag")
	}
}
