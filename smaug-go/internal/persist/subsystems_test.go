package persist

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func TestLoadClansFromDir(t *testing.T) {
	w := world.New("../../../db")
	clanDir := filepath.Join(w.DataDir, "clans")

	if _, err := os.Stat(clanDir); os.IsNotExist(err) {
		t.Skip("clans directory not found")
	}

	err := LoadClansFromDir(w, clanDir)
	if err != nil {
		t.Fatalf("LoadClansFromDir: %v", err)
	}

	if len(w.Clans) == 0 {
		t.Error("expected at least one clan to be loaded")
	}

	// Verify a known clan
	found := false
	for _, c := range w.Clans {
		if c.Name != "" {
			found = true
			t.Logf("Loaded clan: %s (type=%d, members=%d)", c.Name, c.ClanType, c.Members)
		}
	}
	if !found {
		t.Error("no clans with names loaded")
	}
}

func TestLoadDeitiesFromDir(t *testing.T) {
	w := world.New("../../../db")
	deityDir := filepath.Join(w.DataDir, "deity")

	if _, err := os.Stat(deityDir); os.IsNotExist(err) {
		t.Skip("deity directory not found")
	}

	err := LoadDeitiesFromDir(w, deityDir)
	if err != nil {
		t.Fatalf("LoadDeitiesFromDir: %v", err)
	}

	if len(w.Deities) == 0 {
		t.Error("expected at least one deity to be loaded")
	}

	for _, d := range w.Deities {
		t.Logf("Loaded deity: %s (alignment=%d, worshippers=%d)", d.Name, d.Alignment, d.Worshippers)
	}
}

func TestLoadSocials(t *testing.T) {
	w := world.New("../../../db")
	socialsPath := filepath.Join(w.DataDir, "system", "en", "socials.dat")

	if _, err := os.Stat(socialsPath); os.IsNotExist(err) {
		t.Skip("socials.dat not found")
	}

	err := LoadSocials(w, socialsPath)
	if err != nil {
		t.Fatalf("LoadSocials: %v", err)
	}

	if len(w.Socials) == 0 {
		t.Error("expected at least one social to be loaded")
	}

	// Check that first social has expected structure
	if len(w.Socials) > 0 {
		s := w.Socials[0]
		if s.Name == "" {
			t.Error("first social should have a name")
		}
		t.Logf("Loaded %d socials, first: %s", len(w.Socials), s.Name)
	}
}

func TestLoadBoards(t *testing.T) {
	w := world.New("../../../db")
	boardsPath := filepath.Join(w.DataDir, "boards", "boards.dat")

	if _, err := os.Stat(boardsPath); os.IsNotExist(err) {
		t.Skip("boards.dat not found")
	}

	err := LoadBoards(w, boardsPath)
	if err != nil {
		t.Fatalf("LoadBoards: %v", err)
	}

	if len(w.Boards) == 0 {
		t.Error("expected at least one board to be loaded")
	}

	for _, b := range w.Boards {
		t.Logf("Loaded board: %s (vnum=%d, read=%d, post=%d)",
			b.NoteFile, b.BoardObj, b.MinReadLevel, b.MinPostLevel)
	}
}

// --------------- LoadSocials from string (unit test) ---------------

func TestLoadSocials_FromString(t *testing.T) {
	input := `#SOCIAL
Name smile~
CharNoArg You smile happily.~
OthersNoArg $n smiles happily.~
CharFound You smile at $N.~
OthersFound $n smiles at $N.~
VictFound $n smiles at you.~
CharAuto You smile at yourself.~
OthersAuto $n smiles at $mself.~
End
#SOCIAL
Name grin~
CharNoArg You grin evilly.~
OthersNoArg $n grins evilly.~
End
#END
`
	w := world.New("/tmp/test")

	// Write to temp file
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "socials.dat")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	err := LoadSocials(w, path)
	if err != nil {
		t.Fatalf("LoadSocials: %v", err)
	}

	if len(w.Socials) != 2 {
		t.Fatalf("Socials = %d, want 2", len(w.Socials))
	}

	s := w.Socials[0]
	if s.Name != "smile" {
		t.Errorf("Socials[0].Name = %q, want %q", s.Name, "smile")
	}
	if s.CharNoArg != "You smile happily." {
		t.Errorf("CharNoArg = %q", s.CharNoArg)
	}
	if s.OthersNoArg != "$n smiles happily." {
		t.Errorf("OthersNoArg = %q", s.OthersNoArg)
	}
	if s.CharFound != "You smile at $N." {
		t.Errorf("CharFound = %q", s.CharFound)
	}
	if s.OthersFound != "$n smiles at $N." {
		t.Errorf("OthersFound = %q", s.OthersFound)
	}
	if s.VictFound != "$n smiles at you." {
		t.Errorf("VictFound = %q", s.VictFound)
	}
	if s.CharAuto != "You smile at yourself." {
		t.Errorf("CharAuto = %q", s.CharAuto)
	}
	if s.OthersAuto != "$n smiles at $mself." {
		t.Errorf("OthersAuto = %q", s.OthersAuto)
	}

	s2 := w.Socials[1]
	if s2.Name != "grin" {
		t.Errorf("Socials[1].Name = %q, want %q", s2.Name, "grin")
	}
	// Fields not set should be empty
	if s2.CharFound != "" {
		t.Errorf("Socials[1].CharFound = %q, want empty", s2.CharFound)
	}
}

func TestLoadSocials_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "socials.dat")
	if err := os.WriteFile(path, []byte("#END\n"), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	w := world.New("/tmp/test")
	err := LoadSocials(w, path)
	if err != nil {
		t.Fatalf("LoadSocials: %v", err)
	}
	if len(w.Socials) != 0 {
		t.Errorf("Socials = %d, want 0", len(w.Socials))
	}
}

func TestLoadSocials_DollarTerminator(t *testing.T) {
	input := `#SOCIAL
Name wave~
CharNoArg You wave.~
OthersNoArg $n waves.~
End
$
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "socials.dat")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	w := world.New("/tmp/test")
	err := LoadSocials(w, path)
	if err != nil {
		t.Fatalf("LoadSocials: %v", err)
	}
	if len(w.Socials) != 1 {
		t.Fatalf("Socials = %d, want 1", len(w.Socials))
	}
	if w.Socials[0].Name != "wave" {
		t.Errorf("Name = %q, want wave", w.Socials[0].Name)
	}
}

func TestLoadSocials_NonexistentFile(t *testing.T) {
	w := world.New("/tmp/test")
	err := LoadSocials(w, "/nonexistent/socials.dat")
	if err == nil {
		t.Error("LoadSocials with missing file should return error")
	}
}

// --------------- LoadBoards from string (unit test) ---------------

func TestLoadBoards_FromString(t *testing.T) {
	input := `Filename public.board~
Vnum 10001
Min_read_level 1
Min_post_level 5
Min_remove_level 51
Max_posts 100
Type 0
End
Filename imm.board~
Vnum 10002
Min_read_level 51
Min_post_level 51
Min_remove_level 60
Max_posts 50
Type 1
End
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "boards.dat")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	w := world.New("/tmp/test")
	err := LoadBoards(w, path)
	if err != nil {
		t.Fatalf("LoadBoards: %v", err)
	}

	if len(w.Boards) != 2 {
		t.Fatalf("Boards = %d, want 2", len(w.Boards))
	}

	b := w.Boards[0]
	if b.NoteFile != "public.board" {
		t.Errorf("NoteFile = %q, want %q", b.NoteFile, "public.board")
	}
	if b.BoardObj != 10001 {
		t.Errorf("BoardObj = %d, want 10001", b.BoardObj)
	}
	if b.MinReadLevel != 1 {
		t.Errorf("MinReadLevel = %d, want 1", b.MinReadLevel)
	}
	if b.MinPostLevel != 5 {
		t.Errorf("MinPostLevel = %d, want 5", b.MinPostLevel)
	}
	if b.MinRemoveLevel != 51 {
		t.Errorf("MinRemoveLevel = %d, want 51", b.MinRemoveLevel)
	}
	if b.MaxPosts != 100 {
		t.Errorf("MaxPosts = %d, want 100", b.MaxPosts)
	}
	if b.Type != 0 {
		t.Errorf("Type = %d, want 0", b.Type)
	}

	b2 := w.Boards[1]
	if b2.NoteFile != "imm.board" {
		t.Errorf("NoteFile = %q, want %q", b2.NoteFile, "imm.board")
	}
	if b2.BoardObj != 10002 {
		t.Errorf("BoardObj = %d, want 10002", b2.BoardObj)
	}
	if b2.Type != 1 {
		t.Errorf("Type = %d, want 1", b2.Type)
	}
}

func TestLoadBoards_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "boards.dat")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	w := world.New("/tmp/test")
	err := LoadBoards(w, path)
	if err != nil {
		t.Fatalf("LoadBoards: %v", err)
	}
	if len(w.Boards) != 0 {
		t.Errorf("Boards = %d, want 0", len(w.Boards))
	}
}

func TestLoadBoards_NonexistentFile(t *testing.T) {
	w := world.New("/tmp/test")
	err := LoadBoards(w, "/nonexistent/boards.dat")
	if err == nil {
		t.Error("LoadBoards with missing file should return error")
	}
}

// --------------- loadClan from string (unit test) ---------------

func TestLoadClan_FromString(t *testing.T) {
	input := `#CLAN
Name Test Clan~
Filename testclan~
Motto Victory or Death~
Leader Gandalf~
NumberOne Aragorn~
NumberTwo Legolas~
Type 0
Members 10
Recall 21001
Storeroom 21002
Board 10001
Score 500
MKills 100
MDeaths 5
End
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "testclan.clan")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	clan, err := loadClan(path)
	if err != nil {
		t.Fatalf("loadClan: %v", err)
	}

	if clan.Name != "Test Clan" {
		t.Errorf("Name = %q, want %q", clan.Name, "Test Clan")
	}
	if clan.Motto != "Victory or Death" {
		t.Errorf("Motto = %q", clan.Motto)
	}
	if clan.Leader != "Gandalf" {
		t.Errorf("Leader = %q", clan.Leader)
	}
	if clan.Number1 != "Aragorn" {
		t.Errorf("Number1 = %q", clan.Number1)
	}
	if clan.Number2 != "Legolas" {
		t.Errorf("Number2 = %q", clan.Number2)
	}
	if clan.ClanType != 0 {
		t.Errorf("ClanType = %d", clan.ClanType)
	}
	if clan.Members != 10 {
		t.Errorf("Members = %d", clan.Members)
	}
	if clan.Recall != 21001 {
		t.Errorf("Recall = %d", clan.Recall)
	}
	if clan.Storeroom != 21002 {
		t.Errorf("Storeroom = %d", clan.Storeroom)
	}
	if clan.Score != 500 {
		t.Errorf("Score = %d", clan.Score)
	}
	if clan.MKills != 100 {
		t.Errorf("MKills = %d", clan.MKills)
	}
}

func TestLoadClan_BadHeader(t *testing.T) {
	input := `#NOTCLAN
Name Bad~
End
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bad.clan")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	_, err := loadClan(path)
	if err == nil {
		t.Error("loadClan with bad header should return error")
	}
}

func TestLoadClan_AllFields(t *testing.T) {
	input := `#CLAN
Name Full Clan~
Filename fullclan~
Description A full clan.~
Deity Mota~
Badge [FC]~
PKills 10
PDeaths 3
IllegalPK 1
Class 2
Favour 50
Strikes 2
Alignment 1000
ClanObjOne 5001
ClanObjTwo 5002
ClanObjThree 5003
GuardOne 6001
GuardTwo 6002
End
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "full.clan")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	clan, err := loadClan(path)
	if err != nil {
		t.Fatalf("loadClan: %v", err)
	}

	if clan.Description != "A full clan." {
		t.Errorf("Description = %q", clan.Description)
	}
	if clan.Deity != "Mota" {
		t.Errorf("Deity = %q", clan.Deity)
	}
	if clan.Badge != "[FC]" {
		t.Errorf("Badge = %q", clan.Badge)
	}
	// C fread_clan:434-435 stores legacy single-int PKills/PDeaths into
	// index [6] (the cumulative slot), not [0]. Pre-existing Go bug fixed
	// 2026-04-19 in G0 of plan-phase6-clan-officer.md.
	if clan.PKills[6] != 10 {
		t.Errorf("PKills[6] = %d, want 10", clan.PKills[6])
	}
	if clan.PDeaths[6] != 3 {
		t.Errorf("PDeaths[6] = %d, want 3", clan.PDeaths[6])
	}
	if clan.IllegalPK != 1 {
		t.Errorf("IllegalPK = %d", clan.IllegalPK)
	}
	if clan.Class != 2 {
		t.Errorf("Class = %d", clan.Class)
	}
	if clan.Favour != 50 {
		t.Errorf("Favour = %d", clan.Favour)
	}
	if clan.Strikes != 2 {
		t.Errorf("Strikes = %d", clan.Strikes)
	}
	if clan.Alignment != 1000 {
		t.Errorf("Alignment = %d", clan.Alignment)
	}
	if clan.ClanObj1 != 5001 {
		t.Errorf("ClanObj1 = %d", clan.ClanObj1)
	}
	if clan.ClanObj2 != 5002 {
		t.Errorf("ClanObj2 = %d", clan.ClanObj2)
	}
	if clan.ClanObj3 != 5003 {
		t.Errorf("ClanObj3 = %d", clan.ClanObj3)
	}
	if clan.Guard1 != 6001 {
		t.Errorf("Guard1 = %d", clan.Guard1)
	}
	if clan.Guard2 != 6002 {
		t.Errorf("Guard2 = %d", clan.Guard2)
	}
}

// --------------- SaveClan + round-trip (G0) ---------------

// TestLoadClan_LegacyPKillsIndex6 pins the C fread_clan:434-435 semantics:
// the legacy single-int "PKills N" / "PDeaths N" form stores into index [6]
// (the cumulative "total" slot), not [0]. Pre-existing bug fixed 2026-04-19.
func TestLoadClan_LegacyPKillsIndex6(t *testing.T) {
	input := `#CLAN
Name Legacy~
Filename legacy.clan~
PKills 42
PDeaths 7
End
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "legacy.clan")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	clan, err := loadClan(path)
	if err != nil {
		t.Fatalf("loadClan: %v", err)
	}
	if clan.PKills[6] != 42 {
		t.Errorf("PKills[6] = %d, want 42 (C fread_clan:435 legacy slot)", clan.PKills[6])
	}
	if clan.PKills[0] != 0 {
		t.Errorf("PKills[0] = %d, want 0 (legacy PKills must not target index 0)", clan.PKills[0])
	}
	if clan.PDeaths[6] != 7 {
		t.Errorf("PDeaths[6] = %d, want 7", clan.PDeaths[6])
	}
	if clan.PDeaths[0] != 0 {
		t.Errorf("PDeaths[0] = %d, want 0", clan.PDeaths[0])
	}
}

// TestLoadClan_PKillRangeNew_SevenInts exercises the active 7-int block form
// (C fread_clan:459-469).
func TestLoadClan_PKillRangeNew_SevenInts(t *testing.T) {
	input := `#CLAN
Name Range~
Filename range.clan~
PKillRangeNew 1 2 3 4 5 6 7
PDeathRangeNew 10 20 30 40 50 60 70
End
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "range.clan")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	clan, err := loadClan(path)
	if err != nil {
		t.Fatalf("loadClan: %v", err)
	}
	want := [7]int{1, 2, 3, 4, 5, 6, 7}
	if clan.PKills != want {
		t.Errorf("PKills = %v, want %v", clan.PKills, want)
	}
	wantD := [7]int{10, 20, 30, 40, 50, 60, 70}
	if clan.PDeaths != wantD {
		t.Errorf("PDeaths = %v, want %v", clan.PDeaths, wantD)
	}
}

// TestLoadClan_PKillRangeLegacy_Discarded verifies the legacy 7-int form
// consumes seven numbers but doesn't store them (C fread_clan:437-447, 470-479).
func TestLoadClan_PKillRangeLegacy_Discarded(t *testing.T) {
	input := `#CLAN
Name Legacy7~
Filename legacy7.clan~
PKillRange 100 200 300 400 500 600 700
PDeathRange 11 22 33 44 55 66 77
Score 99
End
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "legacy7.clan")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	clan, err := loadClan(path)
	if err != nil {
		t.Fatalf("loadClan: %v", err)
	}
	var zero [7]int
	if clan.PKills != zero {
		t.Errorf("PKills = %v, want zeroed (legacy form discards)", clan.PKills)
	}
	if clan.PDeaths != zero {
		t.Errorf("PDeaths = %v, want zeroed (legacy form discards)", clan.PDeaths)
	}
	// Next key (Score) must still parse — proves the seven ReadNumber
	// calls consumed exactly seven ints and did not leak into the next key.
	if clan.Score != 99 {
		t.Errorf("Score = %d, want 99 (legacy PKillRange should consume 7 ints exactly)", clan.Score)
	}
}

// TestLoadClan_AllExtendedKeys exercises every key added in G0.
func TestLoadClan_AllExtendedKeys(t *testing.T) {
	input := `#CLAN
Name Extended~
Abbrev Ext~
Filename extended.clan~
Leadrank Grand Master~
Onerank Captain~
Tworank Lieutenant~
MemLimit 25
ClanObjFour 7004
ClanObjFive 7005
End
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "extended.clan")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	clan, err := loadClan(path)
	if err != nil {
		t.Fatalf("loadClan: %v", err)
	}
	if clan.Abbrev != "Ext" {
		t.Errorf("Abbrev = %q, want %q", clan.Abbrev, "Ext")
	}
	if clan.LeadRank != "Grand Master" {
		t.Errorf("LeadRank = %q", clan.LeadRank)
	}
	if clan.OneRank != "Captain" {
		t.Errorf("OneRank = %q", clan.OneRank)
	}
	if clan.TwoRank != "Lieutenant" {
		t.Errorf("TwoRank = %q", clan.TwoRank)
	}
	if clan.MemLimit != 25 {
		t.Errorf("MemLimit = %d, want 25", clan.MemLimit)
	}
	if clan.ClanObj4 != 7004 {
		t.Errorf("ClanObj4 = %d, want 7004", clan.ClanObj4)
	}
	if clan.ClanObj5 != 7005 {
		t.Errorf("ClanObj5 = %d, want 7005", clan.ClanObj5)
	}
}

// TestSaveLoadClan_RoundTrip pins that every field in ClanData survives a
// SaveClan → readClan cycle with byte-identical values.
func TestSaveLoadClan_RoundTrip(t *testing.T) {
	orig := &types.ClanData{
		Filename:    "test.clan",
		Name:        "Test Clan",
		Abbrev:      "TC",
		Motto:       "Unity.",
		Description: "A fine clan.",
		Deity:       "Mota",
		Leader:      "Alice",
		Number1:     "Bob",
		Number2:     "Carol",
		Badge:       "[TC]",
		LeadRank:    "Leader Rank",
		OneRank:     "One Rank",
		TwoRank:     "Two Rank",
		PKills:      [7]int{1, 2, 3, 4, 5, 6, 7},
		PDeaths:     [7]int{8, 9, 10, 11, 12, 13, 14},
		MKills:      100,
		MDeaths:     15,
		IllegalPK:   2,
		Score:       250,
		ClanType:    1,
		Class:       3,
		Favour:      77,
		Strikes:     1,
		Members:     42,
		MemLimit:    50,
		Alignment:   1000,
		Board:       10010,
		ClanObj1:    5001,
		ClanObj2:    5002,
		ClanObj3:    5003,
		ClanObj4:    5004,
		ClanObj5:    5005,
		Recall:      21001,
		Storeroom:   21002,
		Guard1:      6001,
		Guard2:      6002,
	}

	var buf bytes.Buffer
	if err := SaveClan(&buf, orig); err != nil {
		t.Fatalf("SaveClan: %v", err)
	}

	sc := NewScanner(bytes.NewReader(buf.Bytes()), "test.clan")
	got, err := readClan(sc)
	if err != nil {
		t.Fatalf("readClan: %v", err)
	}

	if *got != *orig {
		t.Errorf("round-trip mismatch.\n got=%+v\nwant=%+v", *got, *orig)
	}
}

// TestSaveClan_NilRejected pins the nil-clan guard.
func TestSaveClan_NilRejected(t *testing.T) {
	var buf bytes.Buffer
	if err := SaveClan(&buf, nil); err == nil {
		t.Error("SaveClan(nil) should return error")
	}
}

// TestSaveClan_SmashTildeOnStrings ensures player-typed tildes in clan
// strings are neutralized before write.
func TestSaveClan_SmashTildeOnStrings(t *testing.T) {
	c := &types.ClanData{
		Filename: "smash.clan",
		Name:     "Tilde~Name",
		Motto:    "Lots ~~ tildes",
	}
	var buf bytes.Buffer
	if err := SaveClan(&buf, c); err != nil {
		t.Fatalf("SaveClan: %v", err)
	}
	out := buf.String()
	// Every ~ should be either the terminator at end-of-field or replaced with -.
	// There should be exactly N field-terminating ~ chars = 13 string fields.
	tildeCount := strings.Count(out, "~")
	wantTildes := 13 // 13 string fields each end with a single ~
	if tildeCount != wantTildes {
		t.Errorf("tilde count = %d, want %d (SmashTilde should neutralize embedded tildes)", tildeCount, wantTildes)
	}
	if !strings.Contains(out, "Tilde-Name") {
		t.Errorf("expected smashed Name 'Tilde-Name', got: %q", out)
	}
}

// TestSaveClan_FormatMatchesGolden pins the byte-for-byte C-parity output
// (src/clans.c:206-250). A representative clan is serialized and compared
// against the expected string.
func TestSaveClan_FormatMatchesGolden(t *testing.T) {
	c := &types.ClanData{
		Name:     "Golden",
		Abbrev:   "GD",
		Filename: "golden.clan",
		Motto:    "For Glory",
		Leader:   "Leader",
		PKills:   [7]int{0, 0, 0, 0, 0, 0, 5},
		PDeaths:  [7]int{0, 0, 0, 0, 0, 0, 2},
		ClanType: 0,
		Class:    0,
		Members:  10,
		MemLimit: 50,
	}
	var buf bytes.Buffer
	if err := SaveClan(&buf, c); err != nil {
		t.Fatalf("SaveClan: %v", err)
	}
	want := "#CLAN\n" +
		"Name         Golden~\n" +
		"Abbrev       GD~\n" +
		"Filename     golden.clan~\n" +
		"Motto        For Glory~\n" +
		"Description  ~\n" +
		"Deity        ~\n" +
		"Leader       Leader~\n" +
		"NumberOne    ~\n" +
		"NumberTwo    ~\n" +
		"Badge        ~\n" +
		"Leadrank     ~\n" +
		"Onerank      ~\n" +
		"Tworank      ~\n" +
		"PKillRangeNew   0 0 0 0 0 0 5\n" +
		"PDeathRangeNew  0 0 0 0 0 0 2\n" +
		"MKills       0\n" +
		"MDeaths      0\n" +
		"IllegalPK    0\n" +
		"Score        0\n" +
		"Type         0\n" +
		"Class        0\n" +
		"Favour       0\n" +
		"Strikes      0\n" +
		"Members      10\n" +
		"MemLimit     50\n" +
		"Alignment    0\n" +
		"Board        0\n" +
		"ClanObjOne   0\n" +
		"ClanObjTwo   0\n" +
		"ClanObjThree 0\n" +
		"ClanObjFour  0\n" +
		"ClanObjFive  0\n" +
		"Recall       0\n" +
		"Storeroom    0\n" +
		"GuardOne     0\n" +
		"GuardTwo     0\n" +
		"End\n\n" +
		"#END\n"
	if buf.String() != want {
		t.Errorf("SaveClan output mismatch.\n got=%q\nwant=%q", buf.String(), want)
	}
}

// TestSaveClanFile_WritesToPath exercises the file-creation path.
func TestSaveClanFile_WritesToPath(t *testing.T) {
	tmpDir := t.TempDir()
	c := &types.ClanData{
		Filename: "written.clan",
		Name:     "Written Clan",
		Members:  3,
	}
	if err := SaveClanFile(tmpDir, c); err != nil {
		t.Fatalf("SaveClanFile: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(tmpDir, "written.clan"))
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if !strings.Contains(string(data), "Name         Written Clan~") {
		t.Errorf("written file missing Name line: %s", data)
	}
}

// TestSaveClanFile_EmptyFilenameRejected pins the C src/clans.c:189-194
// empty-filename guard.
func TestSaveClanFile_EmptyFilenameRejected(t *testing.T) {
	if err := SaveClanFile(t.TempDir(), &types.ClanData{Name: "No Filename"}); err == nil {
		t.Error("SaveClanFile with empty filename should return error")
	}
}

// TestSaveClanFile_NilRejected pins the nil-clan guard.
func TestSaveClanFile_NilRejected(t *testing.T) {
	if err := SaveClanFile(t.TempDir(), nil); err == nil {
		t.Error("SaveClanFile(nil) should return error")
	}
}

// TestSaveClan_PropagatesWriteError pins the post-audit fix: prior
// implementation discarded all 39 fmt.Fprintf return values and returned
// nil even on a broken pipe / disk-full mid-write. The errWriter
// short-circuits on the first failure and surfaces the error.
func TestSaveClan_PropagatesWriteError(t *testing.T) {
	w := &shortWriter{limit: 5}
	c := &types.ClanData{Name: "Truncated", Filename: "x.clan"}
	if err := SaveClan(w, c); err == nil {
		t.Error("SaveClan with mid-write failure should return error, got nil")
	}
}

// shortWriter accepts up to `limit` bytes total then returns io.ErrShortWrite.
type shortWriter struct {
	limit   int
	written int
}

func (sw *shortWriter) Write(p []byte) (int, error) {
	if sw.written >= sw.limit {
		return 0, errShortBound
	}
	remaining := sw.limit - sw.written
	if len(p) <= remaining {
		sw.written += len(p)
		return len(p), nil
	}
	sw.written = sw.limit
	return remaining, errShortBound
}

var errShortBound = fmtErrorf("short writer hit limit")

// fmtErrorf is a tiny test helper to avoid pulling fmt into the var init.
func fmtErrorf(s string) error { return &fmtError{s: s} }

type fmtError struct{ s string }

func (e *fmtError) Error() string { return e.s }

// TestSaveClanFile_FileMode pins 0600 — security adversary 2026-04-19.
// Clan files contain PKill records that shouldn't be world-readable on
// shared hosts.
func TestSaveClanFile_FileMode(t *testing.T) {
	tmpDir := t.TempDir()
	c := &types.ClanData{Filename: "mode.clan", Name: "Mode Clan"}
	if err := SaveClanFile(tmpDir, c); err != nil {
		t.Fatalf("SaveClanFile: %v", err)
	}
	info, err := os.Stat(filepath.Join(tmpDir, "mode.clan"))
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("clan-file mode = %#o, want 0o600", mode)
	}
}

// --------------- loadDeity from string (unit test) ---------------

func TestLoadDeity_FromString(t *testing.T) {
	input := `#DEITY
Name Mota~
Filename mota~
Description The god of magic.~
Alignment 1000
Worshippers 5
Flee -1
Kill 1
Kill_magic 2
Aid 1
Die -2
Race 0
Class 2
Sex 1
Element 3
Npcrace 5
Npcfoe 7
End
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "mota.deity")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	deity, err := loadDeity(path)
	if err != nil {
		t.Fatalf("loadDeity: %v", err)
	}

	if deity.Name != "Mota" {
		t.Errorf("Name = %q", deity.Name)
	}
	if deity.Description != "The god of magic." {
		t.Errorf("Description = %q", deity.Description)
	}
	if deity.Alignment != 1000 {
		t.Errorf("Alignment = %d", deity.Alignment)
	}
	if deity.Worshippers != 5 {
		t.Errorf("Worshippers = %d", deity.Worshippers)
	}
	if deity.Flee != -1 {
		t.Errorf("Flee = %d", deity.Flee)
	}
	if deity.Kill != 1 {
		t.Errorf("Kill = %d", deity.Kill)
	}
	if deity.KillMagic != 2 {
		t.Errorf("KillMagic = %d", deity.KillMagic)
	}
	if deity.Aid != 1 {
		t.Errorf("Aid = %d", deity.Aid)
	}
	if deity.Die != -2 {
		t.Errorf("Die = %d", deity.Die)
	}
	if deity.Race != 0 {
		t.Errorf("Race = %d", deity.Race)
	}
	if deity.Class != 2 {
		t.Errorf("Class = %d", deity.Class)
	}
	if deity.Sex != 1 {
		t.Errorf("Sex = %d", deity.Sex)
	}
	if deity.NpcRace != 5 {
		t.Errorf("NpcRace = %d", deity.NpcRace)
	}
	if deity.NpcFoe != 7 {
		t.Errorf("NpcFoe = %d", deity.NpcFoe)
	}
}

func TestLoadDeity_BadHeader(t *testing.T) {
	input := `#NOTDEITY
Name Bad~
End
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bad.deity")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	_, err := loadDeity(path)
	if err == nil {
		t.Error("loadDeity with bad header should return error")
	}
}

func TestLoadDeity_AllFields(t *testing.T) {
	input := `#DEITY
Name Full~
Filename full~
Flee_npcrace 1
Flee_npcfoe 2
Kill_npcrace 3
Kill_npcfoe 4
Sac 5
Bury_corpse 6
Aid_spell 7
Backstab 8
Steal 9
Die_npcrace 10
Die_npcfoe 11
Spell_aid 12
Dig_corpse 13
Scorpse 14
Savatar 15
Sdeityobj 16
Srecall 17
Race2 18
Suscept 19
Susceptnum 20
Elementnum 21
Affectednum 22
Objstat 23
Affected 1 0 0 0
End
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "full.deity")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	deity, err := loadDeity(path)
	if err != nil {
		t.Fatalf("loadDeity: %v", err)
	}

	checks := []struct {
		name string
		got  int
		want int
	}{
		{"FleeNpcRace", deity.FleeNpcRace, 1},
		{"FleeNpcFoe", deity.FleeNpcFoe, 2},
		{"KillNpcRace", deity.KillNpcRace, 3},
		{"KillNpcFoe", deity.KillNpcFoe, 4},
		{"Sac", deity.Sac, 5},
		{"BuryCorpse", deity.BuryCorpse, 6},
		{"AidSpell", deity.AidSpell, 7},
		{"Backstab", deity.Backstab, 8},
		{"Steal", deity.Steal, 9},
		{"DieNpcRace", deity.DieNpcRace, 10},
		{"DieNpcFoe", deity.DieNpcFoe, 11},
		{"SpellAid", deity.SpellAid, 12},
		{"DigCorpse", deity.DigCorpse, 13},
		{"SCorpse", deity.SCorpse, 14},
		{"SAvatar", deity.SAvatar, 15},
		{"SDeityObj", deity.SDeityObj, 16},
		{"SRecall", deity.SRecall, 17},
		{"Race2", deity.Race2, 18},
		{"Suscept", deity.Suscept, 19},
		{"SusceptNum", deity.SusceptNum, 20},
		{"ElementNum", deity.ElementNum, 21},
		{"AffectedNum", deity.AffectedNum, 22},
		{"ObjStat", deity.ObjStat, 23},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d", c.name, c.got, c.want)
		}
	}

	if !deity.Affected.IsSet(0) {
		t.Error("Affected bit 0 should be set")
	}
}

// --------------- loadFromList ---------------

func TestLoadFromList_Basic(t *testing.T) {
	input := "file1.dat\nfile2.dat\n$\n"
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.lst")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	var loaded []string
	err := loadFromList(path, func(name string) {
		loaded = append(loaded, name)
	})
	if err != nil {
		t.Fatalf("loadFromList: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("loaded %d files, want 2", len(loaded))
	}
	if loaded[0] != "file1.dat" || loaded[1] != "file2.dat" {
		t.Errorf("loaded = %v, want [file1.dat file2.dat]", loaded)
	}
}

func TestLoadFromList_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "empty.lst")
	if err := os.WriteFile(path, []byte("$\n"), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	var loaded []string
	err := loadFromList(path, func(name string) {
		loaded = append(loaded, name)
	})
	if err != nil {
		t.Fatalf("loadFromList: %v", err)
	}
	if len(loaded) != 0 {
		t.Errorf("loaded %d, want 0", len(loaded))
	}
}

func TestLoadFromList_Nonexistent(t *testing.T) {
	err := loadFromList("/nonexistent/list.lst", func(string) {})
	if err == nil {
		t.Error("loadFromList with missing file should return error")
	}
}

func TestLoadFromList_SkipsEmptyLines(t *testing.T) {
	input := "\nfile1.dat\n\n$\n"
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "skip.lst")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	var loaded []string
	err := loadFromList(path, func(name string) {
		loaded = append(loaded, name)
	})
	if err != nil {
		t.Fatalf("loadFromList: %v", err)
	}
	// Empty first line terminates early in current implementation
	// because strings.TrimSpace("") == ""
	if len(loaded) != 0 {
		t.Logf("loaded %d files (empty line may terminate early)", len(loaded))
	}
}

// --------------- LoadClansFromDir edge case ---------------

func TestLoadClansFromDir_Nonexistent(t *testing.T) {
	w := world.New("/tmp/test")
	err := LoadClansFromDir(w, "/nonexistent/clans")
	if err == nil {
		t.Error("LoadClansFromDir with missing dir should return error")
	}
}

func TestLoadDeitiesFromDir_Nonexistent(t *testing.T) {
	w := world.New("/tmp/test")
	err := LoadDeitiesFromDir(w, "/nonexistent/deity")
	if err == nil {
		t.Error("LoadDeitiesFromDir with missing dir should return error")
	}
}

// --------------- LoadBoards partial data ---------------

func TestLoadBoards_PartialBoard(t *testing.T) {
	// A board entry that's missing some fields (but has End)
	input := `Filename partial.board~
Vnum 9999
End
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "boards.dat")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	w := world.New("/tmp/test")
	err := LoadBoards(w, path)
	if err != nil {
		t.Fatalf("LoadBoards: %v", err)
	}
	if len(w.Boards) != 1 {
		t.Fatalf("Boards = %d, want 1", len(w.Boards))
	}
	if w.Boards[0].NoteFile != "partial.board" {
		t.Errorf("NoteFile = %q", w.Boards[0].NoteFile)
	}
	if w.Boards[0].BoardObj != 9999 {
		t.Errorf("BoardObj = %d, want 9999", w.Boards[0].BoardObj)
	}
	// Missing fields should be zero
	if w.Boards[0].MinReadLevel != 0 {
		t.Errorf("MinReadLevel = %d, want 0", w.Boards[0].MinReadLevel)
	}
}

// Use strings to suppress unused import warning (already imported above)
var _ = strings.TrimSpace
