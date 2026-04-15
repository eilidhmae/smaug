package persist

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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

func TestPlayerFilePath(t *testing.T) {
	tests := []struct {
		name    string
		dataDir string
		pname   string
		want    string
	}{
		{"normal", "/data", "Gandalf", "/data/player/g/Gandalf"},
		{"uppercase", "/data", "Aragorn", "/data/player/a/Aragorn"},
		{"empty name", "/data", "", ""},
		{"three char", "/db", "Abc", "/db/player/a/Abc"},
		{"twelve char", "/db", "Abcdefghijkl", "/db/player/a/Abcdefghijkl"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PlayerFilePath(tt.dataDir, tt.pname)
			if got != tt.want {
				t.Errorf("PlayerFilePath(%q, %q) = %q, want %q", tt.dataDir, tt.pname, got, tt.want)
			}
		})
	}
}

func TestPlayerFilePath_PathTraversal(t *testing.T) {
	// These should all return empty because after filepath.Base the result
	// either fails the regex or the traversal is stripped.
	rejectCases := []string{
		"a",                                // too short
		"ab",                               // too short
		"a123bcdef",                         // contains digits
		"thisnameiswaytoolongforvalidation", // too long (>12)
		"name.with.dots",                    // contains dots
		"name with spaces",                  // contains spaces
		"../",                               // Base returns "."
		"",                                  // empty
	}
	for _, name := range rejectCases {
		got := PlayerFilePath("/data", name)
		if got != "" {
			t.Errorf("PlayerFilePath(%q) = %q, want empty", name, got)
		}
	}

	// Path traversal attempts: filepath.Base strips the directory part,
	// so these resolve to the base name which is a valid player name.
	// The path traversal is neutralised by filepath.Base.
	traversalCases := []struct {
		input string
		want  string
	}{
		{"../../../etc/passwd", "/data/player/p/passwd"},
		{"../../evil", "/data/player/e/evil"},
		{"/absolute/path", "/data/player/p/path"},
	}
	for _, tc := range traversalCases {
		got := PlayerFilePath("/data", tc.input)
		if got != tc.want {
			t.Errorf("PlayerFilePath(%q) = %q, want %q (traversal stripped)", tc.input, got, tc.want)
		}
	}
}

func TestRoomVnum_InRoom(t *testing.T) {
	ch := &types.CharData{
		InRoom:   &types.RoomIndexData{Vnum: 3001},
		HomeVnum: 21001,
	}
	got := roomVnum(ch)
	if got != 3001 {
		t.Errorf("roomVnum with InRoom = %d, want 3001", got)
	}
}

func TestRoomVnum_HomeVnum(t *testing.T) {
	ch := &types.CharData{HomeVnum: 21001}
	got := roomVnum(ch)
	if got != 21001 {
		t.Errorf("roomVnum with HomeVnum = %d, want 21001", got)
	}
}

func TestRoomVnum_Default(t *testing.T) {
	ch := &types.CharData{}
	got := roomVnum(ch)
	if got != types.ROOM_VNUM_TEMPLE {
		t.Errorf("roomVnum default = %d, want ROOM_VNUM_TEMPLE (%d)", got, types.ROOM_VNUM_TEMPLE)
	}
}

func TestSavePlayer_NilPCData(t *testing.T) {
	ch := &types.CharData{Name: "NoPCData"}
	var buf bytes.Buffer
	err := SavePlayer(&buf, ch)
	if err == nil {
		t.Error("SavePlayer with nil PCData should return error")
	}
}

func TestSaveLoadPlayer_AllFields(t *testing.T) {
	ch := &types.CharData{
		Name:                "Fulltest",
		Description:         "A fully loaded test character.",
		Sex:                 1,
		Class:               5,
		Race:                3,
		Speaks:              7,
		Speaking:            1,
		Level:               50,
		Played:              7200,
		HomeVnum:            3001,
		Hit:                 500,
		MaxHit:              600,
		Mana:                200,
		MaxMana:             300,
		Move:                150,
		MaxMove:             200,
		Gold:                10000,
		Exp:                 999999,
		Height:              70,
		Weight:              175,
		Position:            types.POS_STANDING,
		Style:               types.STYLE_FIGHTING,
		Practice:            15,
		Alignment:           1000,
		SavingPoisonDeath:   -5,
		SavingWand:          -3,
		SavingParaPetri:     -2,
		SavingBreath:        -4,
		SavingSpellStaff:    -1,
		Hitroll:             12,
		Damroll:             15,
		Armor:               -50,
		Wimpy:               50,
		PermStr:             18,
		PermInt:             16,
		PermWis:             14,
		PermDex:             20,
		PermCon:             15,
		PermCha:             12,
		PermLck:             17,
		ModStr:              2,
		ModInt:              1,
		ModWis:              0,
		ModDex:              3,
		ModCon:              -1,
		ModCha:              0,
		ModLck:              1,
		Trust:               60,
		PCData: &types.PCData{
			Pwd:        "s3cret",
			Title:      "the Magnificent",
			Prompt:     "<%hhp %mmana> ",
			PagerLen:   30,
			Flags:      5,
			PKills:     3,
			PDeaths:    1,
			MKills:     500,
			MDeaths:    20,
			RecentSite: "10.0.0.1",
		},
	}
	ch.PCData.Condition[0] = 48
	ch.PCData.Condition[1] = 48
	ch.PCData.Condition[2] = 48
	ch.PCData.Condition[3] = 0

	// Save
	var buf bytes.Buffer
	err := SavePlayer(&buf, ch)
	if err != nil {
		t.Fatalf("SavePlayer: %v", err)
	}

	// Load back
	loaded, err := LoadPlayer(bytes.NewReader(buf.Bytes()), "Fulltest")
	if err != nil {
		t.Fatalf("LoadPlayer: %v", err)
	}

	// Verify fields
	checks := []struct {
		name string
		got  int
		want int
	}{
		{"Sex", loaded.Sex, ch.Sex},
		{"Class", loaded.Class, ch.Class},
		{"Race", loaded.Race, ch.Race},
		{"Level", loaded.Level, ch.Level},
		{"Played", loaded.Played, ch.Played},
		{"Hit", loaded.Hit, ch.Hit},
		{"MaxHit", loaded.MaxHit, ch.MaxHit},
		{"Mana", loaded.Mana, ch.Mana},
		{"MaxMana", loaded.MaxMana, ch.MaxMana},
		{"Move", loaded.Move, ch.Move},
		{"MaxMove", loaded.MaxMove, ch.MaxMove},
		{"Gold", loaded.Gold, ch.Gold},
		{"Exp", loaded.Exp, ch.Exp},
		{"Height", loaded.Height, ch.Height},
		{"Weight", loaded.Weight, ch.Weight},
		{"Position", loaded.Position, ch.Position},
		{"Style", loaded.Style, ch.Style},
		{"Practice", loaded.Practice, ch.Practice},
		{"Alignment", loaded.Alignment, ch.Alignment},
		{"SavingPoisonDeath", loaded.SavingPoisonDeath, ch.SavingPoisonDeath},
		{"SavingWand", loaded.SavingWand, ch.SavingWand},
		{"SavingParaPetri", loaded.SavingParaPetri, ch.SavingParaPetri},
		{"SavingBreath", loaded.SavingBreath, ch.SavingBreath},
		{"SavingSpellStaff", loaded.SavingSpellStaff, ch.SavingSpellStaff},
		{"Hitroll", loaded.Hitroll, ch.Hitroll},
		{"Damroll", loaded.Damroll, ch.Damroll},
		{"Armor", loaded.Armor, ch.Armor},
		{"Wimpy", loaded.Wimpy, ch.Wimpy},
		{"PermStr", loaded.PermStr, ch.PermStr},
		{"PermInt", loaded.PermInt, ch.PermInt},
		{"PermWis", loaded.PermWis, ch.PermWis},
		{"PermDex", loaded.PermDex, ch.PermDex},
		{"PermCon", loaded.PermCon, ch.PermCon},
		{"PermCha", loaded.PermCha, ch.PermCha},
		{"PermLck", loaded.PermLck, ch.PermLck},
		{"ModStr", loaded.ModStr, ch.ModStr},
		{"ModInt", loaded.ModInt, ch.ModInt},
		{"ModWis", loaded.ModWis, ch.ModWis},
		{"ModDex", loaded.ModDex, ch.ModDex},
		{"ModCon", loaded.ModCon, ch.ModCon},
		{"ModCha", loaded.ModCha, ch.ModCha},
		{"ModLck", loaded.ModLck, ch.ModLck},
		{"Trust", loaded.Trust, ch.Trust},
		{"Speaks", loaded.Speaks, ch.Speaks},
		{"Speaking", loaded.Speaking, ch.Speaking},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d", c.name, c.got, c.want)
		}
	}

	// String fields
	strChecks := []struct {
		name string
		got  string
		want string
	}{
		{"Name", loaded.Name, ch.Name},
		{"Pwd", loaded.PCData.Pwd, ch.PCData.Pwd},
		{"Title", loaded.PCData.Title, ch.PCData.Title},
		{"Prompt", loaded.PCData.Prompt, ch.PCData.Prompt},
		{"RecentSite", loaded.PCData.RecentSite, ch.PCData.RecentSite},
	}
	for _, c := range strChecks {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}

	// PCData int fields
	pcChecks := []struct {
		name string
		got  int
		want int
	}{
		{"PagerLen", loaded.PCData.PagerLen, ch.PCData.PagerLen},
		{"Flags", loaded.PCData.Flags, ch.PCData.Flags},
		{"PKills", loaded.PCData.PKills, ch.PCData.PKills},
		{"PDeaths", loaded.PCData.PDeaths, ch.PCData.PDeaths},
		{"MKills", loaded.PCData.MKills, ch.PCData.MKills},
		{"MDeaths", loaded.PCData.MDeaths, ch.PCData.MDeaths},
	}
	for _, c := range pcChecks {
		if c.got != c.want {
			t.Errorf("PCData.%s = %d, want %d", c.name, c.got, c.want)
		}
	}

	// Conditions
	for i := 0; i < 4; i++ {
		if loaded.PCData.Condition[i] != ch.PCData.Condition[i] {
			t.Errorf("Condition[%d] = %d, want %d", i, loaded.PCData.Condition[i], ch.PCData.Condition[i])
		}
	}
}

func TestLoadPlayer_Full(t *testing.T) {
	path := filepath.Join("testdata", "Testchar_full")
	f, err := os.Open(path)
	if err != nil {
		t.Skipf("testdata not found: %s", path)
	}
	defer f.Close()

	ch, err := LoadPlayer(f, "Testchar_full")
	if err != nil {
		t.Fatalf("LoadPlayer failed: %v", err)
	}

	if ch.Name != "Testfull" {
		t.Errorf("Name = %q", ch.Name)
	}
	if ch.ShortDescr != "Testfull" {
		t.Errorf("ShortDescr = %q, want Name echo", ch.ShortDescr)
	}
	if ch.Description == "" {
		t.Error("Description should not be empty")
	}
	if ch.Silver != 500 {
		t.Errorf("Silver = %d, want 500", ch.Silver)
	}
	if ch.Copper != 200 {
		t.Errorf("Copper = %d, want 200", ch.Copper)
	}
	if ch.Stance != 2 {
		t.Errorf("Stance = %d, want 2", ch.Stance)
	}
	if ch.Stances[0] != 10 || ch.Stances[11] != 120 {
		t.Errorf("Stances[0]=%d [11]=%d", ch.Stances[0], ch.Stances[11])
	}
	if !ch.AffectedBy.IsSet(1) {
		t.Error("AffectedBy bit 1 should be set")
	}
	if !ch.NoAffectedBy.IsSet(2) {
		t.Error("NoAffectedBy bit 2 should be set")
	}
	if !ch.Deaf.IsSet(3) {
		t.Error("Deaf bit 3 should be set")
	}
	if ch.Resistant != 4 {
		t.Errorf("Resistant = %d, want 4", ch.Resistant)
	}
	if ch.Immune != 2 {
		t.Errorf("Immune = %d, want 2", ch.Immune)
	}
	if ch.Susceptible != 8 {
		t.Errorf("Susceptible = %d, want 8", ch.Susceptible)
	}
	if ch.NoResistant != 16 {
		t.Errorf("NoResistant = %d", ch.NoResistant)
	}
	if ch.NoImmune != 32 {
		t.Errorf("NoImmune = %d", ch.NoImmune)
	}
	if ch.NoSusceptible != 64 {
		t.Errorf("NoSusceptible = %d", ch.NoSusceptible)
	}
	if ch.MentalState != -10 {
		t.Errorf("MentalState = %d", ch.MentalState)
	}

	p := ch.PCData
	if p == nil {
		t.Fatal("PCData is nil")
	}
	if p.Favor != 100 {
		t.Errorf("Favor = %d", p.Favor)
	}
	if p.Honour != 50 {
		t.Errorf("Honour = %d", p.Honour)
	}
	if p.Rank != "Lord" {
		t.Errorf("Rank = %q", p.Rank)
	}
	if p.Bestowments != "all" {
		t.Errorf("Bestowments = %q", p.Bestowments)
	}
	if p.Homepage != "http://mud.org" {
		t.Errorf("Homepage = %q", p.Homepage)
	}
	if p.Email != "test@mud.org" {
		t.Errorf("Email = %q", p.Email)
	}
	if p.Bio != "A test character bio." {
		t.Errorf("Bio = %q", p.Bio)
	}
	if p.AuthedBy != "Admin" {
		t.Errorf("AuthedBy = %q", p.AuthedBy)
	}
	if p.MinSnoop != 55 {
		t.Errorf("MinSnoop = %d", p.MinSnoop)
	}
	if p.FPrompt != "<%hhp %mmana %vmv> " {
		t.Errorf("FPrompt = %q", p.FPrompt)
	}
	if p.WizInvis != 60 {
		t.Errorf("WizInvis = %d", p.WizInvis)
	}
	if p.BamfIn != "appears in a flash of light" {
		t.Errorf("BamfIn = %q", p.BamfIn)
	}
	if p.BamfOut != "vanishes in a puff of smoke" {
		t.Errorf("BamfOut = %q", p.BamfOut)
	}
	if p.IllegalPK != 1 {
		t.Errorf("IllegalPK = %d", p.IllegalPK)
	}
	if p.ClanName != "Guild of Heroes" {
		t.Errorf("ClanName = %q", p.ClanName)
	}
	if p.CouncilName != "Council of Elders" {
		t.Errorf("CouncilName = %q", p.CouncilName)
	}
	if p.DeityName != "Mota" {
		t.Errorf("DeityName = %q", p.DeityName)
	}
	if p.Lang != "en" {
		t.Errorf("Lang = %q", p.Lang)
	}
	if ch.Spouse != "Arwen" {
		t.Errorf("Spouse = %q", ch.Spouse)
	}
	if ch.X != 100 || ch.Y != 200 || ch.Map != 1 {
		t.Errorf("Coordinates = %d,%d,%d, want 100,200,1", ch.X, ch.Y, ch.Map)
	}

	// Should have loaded one affect
	if len(ch.Affects) != 1 {
		t.Fatalf("Affects = %d, want 1", len(ch.Affects))
	}
	aff := ch.Affects[0]
	if aff.Type != 5 || aff.Duration != 24 || aff.Modifier != -20 || aff.Location != 17 {
		t.Errorf("Affect = type=%d dur=%d mod=%d loc=%d", aff.Type, aff.Duration, aff.Modifier, aff.Location)
	}
}

func TestLoadPlayer_SkipObjectsWithNilLookup(t *testing.T) {
	// A player file with an object section; LoadPlayer (nil lookup) should skip it
	input := `#PLAYER
Name       Skipper~
Sex        1
Class      0
Race       0
Level      1
HpManaMove 10 10 10 10 10 10
AttrPerm   10 10 10 10 10 10 10
AttrMod    0 0 0 0 0 0 0
Condition  48 48 48 0
Password   test~
Position   112
Pagerlen   24
End

#OBJECT
Vnum         100
Name         test object~
WearLoc      -1
End

`
	ch, err := LoadPlayer(strings.NewReader(input), "Skipper")
	if err != nil {
		t.Fatalf("LoadPlayer: %v", err)
	}
	if ch.Name != "Skipper" {
		t.Errorf("Name = %q, want Skipper", ch.Name)
	}
	if len(ch.Carrying) != 0 {
		t.Errorf("Carrying = %d, want 0 (objects should be skipped)", len(ch.Carrying))
	}
}

func TestLoadPlayer_BadHeader(t *testing.T) {
	input := "#BADHEADER\nName Test~\nEnd\n"
	_, err := LoadPlayer(strings.NewReader(input), "test")
	if err == nil {
		t.Error("LoadPlayer with bad header should return error")
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

func TestLoadPlayer_TrustCapped(t *testing.T) {
	// Trust 999 in the player file should be capped to LEVEL_SUPREME (65)
	input := `#PLAYER
Name       Hacker~
Sex        1
Class      0
Race       0
Level      50
Trust      999
HpManaMove 100 100 50 50 80 80
AttrPerm   15 13 12 14 13 11 13
AttrMod    0 0 0 0 0 0 0
Condition  48 48 48 0
Password   test~
Position   112
Pagerlen   24
End

`
	ch, err := LoadPlayer(strings.NewReader(input), "Hacker")
	if err != nil {
		t.Fatalf("LoadPlayer: %v", err)
	}
	if ch.Trust != types.LEVEL_SUPREME {
		t.Errorf("Trust = %d, want %d (LEVEL_SUPREME)", ch.Trust, types.LEVEL_SUPREME)
	}
}

func TestLoadPlayer_TrustNormal(t *testing.T) {
	// Trust within valid range should be preserved
	input := `#PLAYER
Name       Admin~
Sex        1
Class      0
Race       0
Level      50
Trust      60
HpManaMove 100 100 50 50 80 80
AttrPerm   15 13 12 14 13 11 13
AttrMod    0 0 0 0 0 0 0
Condition  48 48 48 0
Password   test~
Position   112
Pagerlen   24
End

`
	ch, err := LoadPlayer(strings.NewReader(input), "Admin")
	if err != nil {
		t.Fatalf("LoadPlayer: %v", err)
	}
	if ch.Trust != 60 {
		t.Errorf("Trust = %d, want 60", ch.Trust)
	}
}

// TestSaveLoadSilverCopper verifies that Silver and Copper coin stashes
// round-trip through SavePlayer/LoadPlayer (bug G5).
func TestSaveLoadSilverCopper(t *testing.T) {
	ch := &types.CharData{
		Name:     "Coinage",
		Level:    5,
		Hit:      50,
		MaxHit:   50,
		Mana:     10,
		MaxMana:  10,
		Move:     40,
		MaxMove:  40,
		Gold:     111,
		Silver:   222,
		Copper:   333,
		Position: types.POS_STANDING,
		PermStr:  13, PermInt: 13, PermWis: 13, PermDex: 13,
		PermCon: 13, PermCha: 13, PermLck: 13,
		PCData: &types.PCData{Pwd: "pass", PagerLen: 24},
	}

	var buf bytes.Buffer
	if err := SavePlayer(&buf, ch); err != nil {
		t.Fatalf("SavePlayer: %v", err)
	}

	saved := buf.String()
	if !strings.Contains(saved, "Silver") {
		t.Errorf("saved file missing Silver line:\n%s", saved)
	}
	if !strings.Contains(saved, "Copper") {
		t.Errorf("saved file missing Copper line:\n%s", saved)
	}

	loaded, err := LoadPlayer(bytes.NewReader(buf.Bytes()), "Coinage")
	if err != nil {
		t.Fatalf("LoadPlayer: %v", err)
	}

	if loaded.Gold != 111 {
		t.Errorf("Gold = %d, want 111", loaded.Gold)
	}
	if loaded.Silver != 222 {
		t.Errorf("Silver = %d, want 222", loaded.Silver)
	}
	if loaded.Copper != 333 {
		t.Errorf("Copper = %d, want 333", loaded.Copper)
	}
}

// TestSaveLoadSkills verifies that learned skill/spell/weapon/tongue
// proficiencies round-trip through SavePlayer/LoadPlayer (bug G4).
func TestSaveLoadSkills(t *testing.T) {
	// Build a tiny skill table: gsn 0 = dodge (SKILL_SKILL),
	// gsn 1 = fireball (SKILL_SPELL), gsn 2 = sword (SKILL_WEAPON),
	// gsn 3 = common (SKILL_TONGUE).
	skills := []*types.SkillType{
		{Name: "dodge", Type: types.SKILL_SKILL},
		{Name: "fireball", Type: types.SKILL_SPELL},
		{Name: "sword", Type: types.SKILL_WEAPON},
		{Name: "common", Type: types.SKILL_TONGUE},
	}

	prevLookup := SkillNameLookup
	prevGetter := SkillGetter
	SkillNameLookup = func(name string) int {
		for i, sk := range skills {
			if strings.EqualFold(sk.Name, name) {
				return i
			}
		}
		return -1
	}
	SkillGetter = func(gsn int) *types.SkillType {
		if gsn < 0 || gsn >= len(skills) {
			return nil
		}
		return skills[gsn]
	}
	defer func() {
		SkillNameLookup = prevLookup
		SkillGetter = prevGetter
	}()

	ch := &types.CharData{
		Name:     "Skillful",
		Level:    5,
		Hit:      50,
		MaxHit:   50,
		Mana:     10,
		MaxMana:  10,
		Move:     40,
		MaxMove:  40,
		Position: types.POS_STANDING,
		PermStr:  13, PermInt: 13, PermWis: 13, PermDex: 13,
		PermCon: 13, PermCha: 13, PermLck: 13,
		PCData: &types.PCData{Pwd: "pass", PagerLen: 24},
	}
	ch.PCData.Learned[0] = 42 // dodge
	ch.PCData.Learned[1] = 77 // fireball
	ch.PCData.Learned[2] = 55 // sword
	ch.PCData.Learned[3] = 33 // common

	var buf bytes.Buffer
	if err := SavePlayer(&buf, ch); err != nil {
		t.Fatalf("SavePlayer: %v", err)
	}

	saved := buf.String()
	if !strings.Contains(saved, "'dodge'") {
		t.Errorf("saved file missing dodge entry:\n%s", saved)
	}
	if !strings.Contains(saved, "'fireball'") {
		t.Errorf("saved file missing fireball entry:\n%s", saved)
	}
	if !strings.Contains(saved, "Skill") {
		t.Errorf("saved file missing Skill keyword:\n%s", saved)
	}
	if !strings.Contains(saved, "Spell") {
		t.Errorf("saved file missing Spell keyword:\n%s", saved)
	}
	if !strings.Contains(saved, "Weapon") {
		t.Errorf("saved file missing Weapon keyword:\n%s", saved)
	}
	if !strings.Contains(saved, "Tongue") {
		t.Errorf("saved file missing Tongue keyword:\n%s", saved)
	}

	loaded, err := LoadPlayer(bytes.NewReader(buf.Bytes()), "Skillful")
	if err != nil {
		t.Fatalf("LoadPlayer: %v", err)
	}
	if loaded.PCData == nil {
		t.Fatal("loaded PCData is nil")
	}
	if loaded.PCData.Learned[0] != 42 {
		t.Errorf("Learned[dodge]=%d, want 42", loaded.PCData.Learned[0])
	}
	if loaded.PCData.Learned[1] != 77 {
		t.Errorf("Learned[fireball]=%d, want 77", loaded.PCData.Learned[1])
	}
	if loaded.PCData.Learned[2] != 55 {
		t.Errorf("Learned[sword]=%d, want 55", loaded.PCData.Learned[2])
	}
	if loaded.PCData.Learned[3] != 33 {
		t.Errorf("Learned[common]=%d, want 33", loaded.PCData.Learned[3])
	}
}

func TestSaveLoadAliases(t *testing.T) {
	ch := &types.CharData{
		Name:     "Aliased",
		Level:    3,
		Hit:      30,
		MaxHit:   30,
		Position: types.POS_STANDING,
		PCData: &types.PCData{
			Pwd:      "pw",
			PagerLen: 24,
			Aliases: []*types.AliasData{
				{Name: "g", Cmd: "get all corpse"},
				{Name: "k", Cmd: "kill"},
			},
		},
	}

	var buf bytes.Buffer
	if err := SavePlayer(&buf, ch); err != nil {
		t.Fatalf("SavePlayer: %v", err)
	}

	loaded, err := LoadPlayer(bytes.NewReader(buf.Bytes()), "Aliased")
	if err != nil {
		t.Fatalf("LoadPlayer: %v", err)
	}
	if loaded.PCData == nil {
		t.Fatal("PCData nil after load")
	}
	if len(loaded.PCData.Aliases) != 2 {
		t.Fatalf("expected 2 aliases, got %d: %+v",
			len(loaded.PCData.Aliases), loaded.PCData.Aliases)
	}
	want := map[string]string{"g": "get all corpse", "k": "kill"}
	for _, a := range loaded.PCData.Aliases {
		if w, ok := want[a.Name]; !ok {
			t.Errorf("unexpected alias %q", a.Name)
		} else if a.Cmd != w {
			t.Errorf("alias %q = %q, want %q", a.Name, a.Cmd, w)
		}
	}
}
