package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func setupClanWorld() *world.World {
	w := world.New("/tmp/test")
	WorldRef = w
	return w
}

// --- DoClans ---

func TestDoClans_NoClans(t *testing.T) {
	_ = setupClanWorld()
	ch, client := makeTestChar("Player")
	defer client.Close()

	DoClans(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "no clans") {
		t.Errorf("expected 'no clans', got: %q", out)
	}
}

func TestDoClans_ListClans(t *testing.T) {
	w := setupClanWorld()
	w.Clans = append(w.Clans,
		&types.ClanData{Name: "Warriors Guild", Leader: "Thor", Members: 5},
		&types.ClanData{Name: "Mages Circle", Leader: "Merlin", Members: 3},
	)

	ch, client := makeTestChar("Player")
	defer client.Close()

	DoClans(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Warriors Guild") {
		t.Errorf("expected 'Warriors Guild', got: %q", out)
	}
	if !strings.Contains(out, "Mages Circle") {
		t.Errorf("expected 'Mages Circle', got: %q", out)
	}
	if !strings.Contains(out, "Thor") {
		t.Errorf("expected leader 'Thor', got: %q", out)
	}
	if !strings.Contains(out, "5") {
		t.Errorf("expected member count 5, got: %q", out)
	}
}

// --- DoClanInfo ---

func TestDoClanInfo_NoArg(t *testing.T) {
	_ = setupClanWorld()
	ch, client := makeTestChar("Player")
	defer client.Close()

	DoClanInfo(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Which clan?") {
		t.Errorf("expected 'Which clan?', got: %q", out)
	}
}

func TestDoClanInfo_Found(t *testing.T) {
	w := setupClanWorld()
	w.Clans = append(w.Clans, &types.ClanData{
		Name:        "Warriors Guild",
		Leader:      "Thor",
		Number1:     "Loki",
		Number2:     "Odin",
		Members:     10,
		PKills:      [7]int{5},
		PDeaths:     [7]int{2},
		MKills:      100,
		MDeaths:     10,
		Motto:       "Victory or death!",
		Description: "The strongest fighters.\n\r",
	})

	ch, client := makeTestChar("Player")
	defer client.Close()

	DoClanInfo(ch, "warriors")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Warriors Guild") {
		t.Errorf("expected clan name, got: %q", out)
	}
	if !strings.Contains(out, "Thor") {
		t.Errorf("expected leader, got: %q", out)
	}
	if !strings.Contains(out, "Loki") {
		t.Errorf("expected number1, got: %q", out)
	}
	if !strings.Contains(out, "Victory or death!") {
		t.Errorf("expected motto, got: %q", out)
	}
}

func TestDoClanInfo_NotFound(t *testing.T) {
	w := setupClanWorld()
	w.Clans = append(w.Clans, &types.ClanData{Name: "Warriors Guild"})

	ch, client := makeTestChar("Player")
	defer client.Close()

	DoClanInfo(ch, "nonexistent")
	out := readOutput(ch, client)
	if !strings.Contains(out, "No such clan") {
		t.Errorf("expected 'No such clan', got: %q", out)
	}
}

// --- DoClantalk ---

func TestDoClantalk_NotInClan(t *testing.T) {
	_ = setupClanWorld()
	ch, client := makeTestChar("Player")
	defer client.Close()

	DoClantalk(ch, "hello")
	out := readOutput(ch, client)
	if !strings.Contains(out, "aren't in a clan") {
		t.Errorf("expected 'aren't in a clan', got: %q", out)
	}
}

func TestDoClantalk_NoArg(t *testing.T) {
	_ = setupClanWorld()
	clan := &types.ClanData{Name: "Test Clan"}
	ch, client := makeTestChar("Player")
	defer client.Close()
	ch.PCData.Clan = clan

	DoClantalk(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Clantalk what?") {
		t.Errorf("expected 'Clantalk what?', got: %q", out)
	}
}

func TestDoClantalk_SendsMessage(t *testing.T) {
	w := setupClanWorld()
	clan := &types.ClanData{Name: "Warriors"}

	ch, chClient := makeTestChar("Sender")
	defer chClient.Close()
	ch.PCData.Clan = clan
	w.Descriptors = append(w.Descriptors, ch.Desc)

	clanmate, cmClient := makeTestChar("Clanmate")
	defer cmClient.Close()
	clanmate.PCData.Clan = clan
	w.Descriptors = append(w.Descriptors, clanmate.Desc)

	outsider, outsiderClient := makeTestChar("Outsider")
	defer outsiderClient.Close()
	w.Descriptors = append(w.Descriptors, outsider.Desc)

	DoClantalk(ch, "rally at the gate!")

	chOut := readOutput(ch, chClient)
	if !strings.Contains(chOut, "rally at the gate!") {
		t.Errorf("sender should see own clantalk, got: %q", chOut)
	}

	cmOut := readOutput(clanmate, cmClient)
	if !strings.Contains(cmOut, "rally at the gate!") {
		t.Errorf("clanmate should see clantalk, got: %q", cmOut)
	}

	outsiderOut := readOutput(outsider, outsiderClient)
	if outsiderOut != "" {
		t.Errorf("outsider should not see clantalk, got: %q", outsiderOut)
	}
}

// --- DoClanJoin ---

func TestDoClanJoin_NoArg(t *testing.T) {
	_ = setupClanWorld()
	ch, client := makeTestChar("Player")
	defer client.Close()

	DoClanJoin(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Join which clan?") {
		t.Errorf("expected 'Join which clan?', got: %q", out)
	}
}

func TestDoClanJoin_Success(t *testing.T) {
	w := setupClanWorld()
	clan := &types.ClanData{Name: "Warriors Guild", Members: 5}
	w.Clans = append(w.Clans, clan)

	ch, client := makeTestChar("Player")
	defer client.Close()

	DoClanJoin(ch, "warriors")
	out := readOutput(ch, client)

	if ch.PCData.Clan != clan {
		t.Error("character should be in the clan")
	}
	if ch.PCData.ClanName != "Warriors Guild" {
		t.Errorf("ClanName should be set, got: %q", ch.PCData.ClanName)
	}
	if clan.Members != 6 {
		t.Errorf("members should be 6, got: %d", clan.Members)
	}
	if !strings.Contains(out, "Warriors Guild") {
		t.Errorf("expected clan name in output, got: %q", out)
	}
}

func TestDoClanJoin_AlreadyInClan(t *testing.T) {
	_ = setupClanWorld()
	ch, client := makeTestChar("Player")
	defer client.Close()
	ch.PCData.Clan = &types.ClanData{Name: "Old Clan"}

	DoClanJoin(ch, "warriors")
	out := readOutput(ch, client)
	if !strings.Contains(out, "already in a clan") {
		t.Errorf("expected 'already in a clan', got: %q", out)
	}
}

func TestDoClanJoin_NotFound(t *testing.T) {
	_ = setupClanWorld()
	ch, client := makeTestChar("Player")
	defer client.Close()

	DoClanJoin(ch, "nonexistent")
	out := readOutput(ch, client)
	if !strings.Contains(out, "No such clan") {
		t.Errorf("expected 'No such clan', got: %q", out)
	}
}

// --- DoClanLeave ---

func TestDoClanLeave_NotInClan(t *testing.T) {
	_ = setupClanWorld()
	ch, client := makeTestChar("Player")
	defer client.Close()

	DoClanLeave(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "aren't in a clan") {
		t.Errorf("expected 'aren't in a clan', got: %q", out)
	}
}

func TestDoClanLeave_Success(t *testing.T) {
	_ = setupClanWorld()
	clan := &types.ClanData{Name: "Warriors Guild", Members: 5}
	ch, client := makeTestChar("Player")
	defer client.Close()
	ch.PCData.Clan = clan
	ch.PCData.ClanName = "Warriors Guild"

	DoClanLeave(ch, "")
	out := readOutput(ch, client)

	if ch.PCData.Clan != nil {
		t.Error("character should not be in a clan")
	}
	if ch.PCData.ClanName != "" {
		t.Errorf("ClanName should be empty, got: %q", ch.PCData.ClanName)
	}
	if clan.Members != 4 {
		t.Errorf("members should be 4, got: %d", clan.Members)
	}
	if !strings.Contains(out, "Warriors Guild") {
		t.Errorf("expected clan name in output, got: %q", out)
	}
}

// --- DoDeities ---

func TestDoDeities_NoDeities(t *testing.T) {
	_ = setupClanWorld()
	ch, client := makeTestChar("Player")
	defer client.Close()

	DoDeities(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "no deities") {
		t.Errorf("expected 'no deities', got: %q", out)
	}
}

func TestDoDeities_ListDeities(t *testing.T) {
	w := setupClanWorld()
	w.Deities = append(w.Deities,
		&types.DeityData{Name: "Mota", Worshippers: 10, Alignment: 1000},
		&types.DeityData{Name: "Thoric", Worshippers: 5, Alignment: 0},
	)

	ch, client := makeTestChar("Player")
	defer client.Close()

	DoDeities(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Mota") {
		t.Errorf("expected 'Mota', got: %q", out)
	}
	if !strings.Contains(out, "Thoric") {
		t.Errorf("expected 'Thoric', got: %q", out)
	}
}

// --- DoDevote ---

func TestDoDevote_ShowCurrent(t *testing.T) {
	w := setupClanWorld()
	deity := &types.DeityData{Name: "Mota", Worshippers: 10}
	w.Deities = append(w.Deities, deity)

	ch, client := makeTestChar("Player")
	defer client.Close()
	ch.PCData.Deity = deity

	DoDevote(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Mota") {
		t.Errorf("expected current deity shown, got: %q", out)
	}
}

func TestDoDevote_ShowNone(t *testing.T) {
	_ = setupClanWorld()
	ch, client := makeTestChar("Player")
	defer client.Close()

	DoDevote(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "not devoted") {
		t.Errorf("expected 'not devoted', got: %q", out)
	}
}

func TestDoDevote_Devote(t *testing.T) {
	w := setupClanWorld()
	deity := &types.DeityData{Name: "Mota", Worshippers: 10}
	w.Deities = append(w.Deities, deity)

	ch, client := makeTestChar("Player")
	defer client.Close()

	DoDevote(ch, "mota")
	out := readOutput(ch, client)

	if ch.PCData.Deity != deity {
		t.Error("character should be devoted to deity")
	}
	if ch.PCData.DeityName != "Mota" {
		t.Errorf("DeityName should be 'Mota', got: %q", ch.PCData.DeityName)
	}
	if deity.Worshippers != 11 {
		t.Errorf("worshippers should be 11, got: %d", deity.Worshippers)
	}
	if !strings.Contains(out, "Mota") {
		t.Errorf("expected deity name in output, got: %q", out)
	}
}

func TestDoDevote_AlreadyDevoted(t *testing.T) {
	w := setupClanWorld()
	deity := &types.DeityData{Name: "Mota"}
	w.Deities = append(w.Deities, deity)

	ch, client := makeTestChar("Player")
	defer client.Close()
	ch.PCData.Deity = deity

	DoDevote(ch, "mota")
	out := readOutput(ch, client)
	if !strings.Contains(out, "already devoted") {
		t.Errorf("expected 'already devoted', got: %q", out)
	}
}

func TestDoDevote_Renounce(t *testing.T) {
	w := setupClanWorld()
	deity := &types.DeityData{Name: "Mota", Worshippers: 10}
	w.Deities = append(w.Deities, deity)

	ch, client := makeTestChar("Player")
	defer client.Close()
	ch.PCData.Deity = deity
	ch.PCData.DeityName = "Mota"

	DoDevote(ch, "none")
	out := readOutput(ch, client)

	if ch.PCData.Deity != nil {
		t.Error("character should not be devoted after renouncing")
	}
	if deity.Worshippers != 9 {
		t.Errorf("worshippers should be 9, got: %d", deity.Worshippers)
	}
	if !strings.Contains(out, "renounce") {
		t.Errorf("expected 'renounce' message, got: %q", out)
	}
}

func TestDoDevote_RenounceNone(t *testing.T) {
	_ = setupClanWorld()
	ch, client := makeTestChar("Player")
	defer client.Close()

	DoDevote(ch, "none")
	out := readOutput(ch, client)
	if !strings.Contains(out, "aren't devoted") {
		t.Errorf("expected 'aren't devoted', got: %q", out)
	}
}

func TestDoDevote_NotFound(t *testing.T) {
	w := setupClanWorld()
	w.Deities = append(w.Deities, &types.DeityData{Name: "Mota"})

	ch, client := makeTestChar("Player")
	defer client.Close()

	DoDevote(ch, "nonexistent")
	out := readOutput(ch, client)
	if !strings.Contains(out, "No such deity") {
		t.Errorf("expected 'No such deity', got: %q", out)
	}
}

// --- DoNote ---

func TestDoNote_NoArg(t *testing.T) {
	w := setupClanWorld()
	w.Boards = append(w.Boards, &types.BoardData{})

	ch, client := makeTestChar("Player")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 1, Name: "Test"}

	DoNote(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Note what?") {
		t.Errorf("expected 'Note what?', got: %q", out)
	}
}

func TestDoNote_ListEmpty(t *testing.T) {
	w := setupClanWorld()
	w.Boards = append(w.Boards, &types.BoardData{})

	ch, client := makeTestChar("Player")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 1, Name: "Test"}

	DoNote(ch, "list")
	out := readOutput(ch, client)
	if !strings.Contains(out, "no notes") {
		t.Errorf("expected 'no notes', got: %q", out)
	}
}

func TestDoNote_ListWithNotes(t *testing.T) {
	w := setupClanWorld()
	board := &types.BoardData{
		Notes: []*types.NoteData{
			{Sender: "Admin", Subject: "Welcome", Date: "2024-01-01", ToList: "all", Text: "Hello!\n\r"},
			{Sender: "Builder", Subject: "Bug Fix", Date: "2024-01-02", ToList: "all", Text: "Fixed a bug.\n\r"},
		},
	}
	w.Boards = append(w.Boards, board)

	ch, client := makeTestChar("Player")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 1, Name: "Test"}

	DoNote(ch, "list")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Welcome") {
		t.Errorf("expected 'Welcome' subject, got: %q", out)
	}
	if !strings.Contains(out, "Bug Fix") {
		t.Errorf("expected 'Bug Fix' subject, got: %q", out)
	}
}

func TestDoNote_Read(t *testing.T) {
	w := setupClanWorld()
	board := &types.BoardData{
		Notes: []*types.NoteData{
			{Sender: "Admin", Subject: "Welcome", Date: "2024-01-01", ToList: "all", Text: "Hello everyone!\n\r"},
		},
	}
	w.Boards = append(w.Boards, board)

	ch, client := makeTestChar("Player")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 1, Name: "Test"}

	DoNote(ch, "read 1")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Welcome") {
		t.Errorf("expected subject, got: %q", out)
	}
	if !strings.Contains(out, "Hello everyone!") {
		t.Errorf("expected note text, got: %q", out)
	}
}

func TestDoNote_ReadInvalid(t *testing.T) {
	w := setupClanWorld()
	// Empty board so any read is out of range
	board := &types.BoardData{}
	w.Boards = append(w.Boards, board)

	ch, client := makeTestChar("Player")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 1, Name: "Test"}

	DoNote(ch, "read 1")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Valid range") {
		t.Errorf("expected 'Valid range', got: %q", out)
	}
}

func TestDoNote_PostAndRemove(t *testing.T) {
	w := setupClanWorld()
	board := &types.BoardData{}
	w.Boards = append(w.Boards, board)

	ch, client := makeTestChar("Player")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 1, Name: "Test"}

	// Try posting without writing first
	DoNote(ch, "post")
	out := readOutput(ch, client)
	if !strings.Contains(out, "no note in progress") {
		t.Errorf("expected 'no note in progress', got: %q", out)
	}

	// Write and post
	ch.PNote = &types.NoteData{
		Sender:  "Player",
		Subject: "Test Note",
		ToList:  "all",
		Text:    "This is a test.\n\r",
	}
	DoNote(ch, "post")
	out = readOutput(ch, client)

	if len(board.Notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(board.Notes))
	}
	if board.Notes[0].Subject != "Test Note" {
		t.Errorf("expected subject 'Test Note', got: %q", board.Notes[0].Subject)
	}
	if !strings.Contains(out, "posted") {
		t.Errorf("expected 'posted' message, got: %q", out)
	}

	// Remove it
	DoNote(ch, "remove 1")
	out = readOutput(ch, client)

	if len(board.Notes) != 0 {
		t.Errorf("expected 0 notes after remove, got %d", len(board.Notes))
	}
	if !strings.Contains(out, "removed") {
		t.Errorf("expected 'removed' message, got: %q", out)
	}
}

func TestDoNote_NoBoard(t *testing.T) {
	_ = setupClanWorld()
	ch, client := makeTestChar("Player")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 1, Name: "Test"}

	DoNote(ch, "list")
	out := readOutput(ch, client)
	if !strings.Contains(out, "no board") {
		t.Errorf("expected 'no board', got: %q", out)
	}
}

// --- Tier 2: note recipient filtering ---

func TestIsNoteTo_Matches(t *testing.T) {
	alice := &types.CharData{Name: "Alice", Trust: 0, Level: 10}
	carol := &types.CharData{Name: "Carol", Trust: 0, Level: 10}

	cases := []struct {
		toList string
		wantA  bool
		wantC  bool
	}{
		{"all", true, true},
		{"", true, true}, // empty treated as all
		{"Alice", true, false},
		{"Alice Bob", true, false},
		{"alice", true, false}, // case-insensitive
		{"Carol", false, true},
		{"Dave", false, false},
	}
	for _, c := range cases {
		note := &types.NoteData{ToList: c.toList}
		if got := isNoteTo(alice, note); got != c.wantA {
			t.Errorf("isNoteTo(Alice, %q) = %v; want %v", c.toList, got, c.wantA)
		}
		if got := isNoteTo(carol, note); got != c.wantC {
			t.Errorf("isNoteTo(Carol, %q) = %v; want %v", c.toList, got, c.wantC)
		}
	}
}

func TestDoNote_ListFiltersByRecipient(t *testing.T) {
	w := setupClanWorld()
	board := &types.BoardData{
		Notes: []*types.NoteData{
			{Sender: "Admin", Subject: "ForAll", Date: "d", ToList: "all", Text: "x"},
			{Sender: "Alice", Subject: "ForCarol", Date: "d", ToList: "Carol", Text: "y"},
			{Sender: "Admin", Subject: "ForAlice", Date: "d", ToList: "Alice Dave", Text: "z"},
		},
	}
	w.Boards = append(w.Boards, board)

	alice, aliceClient := makeTestChar("Alice")
	defer aliceClient.Close()
	alice.InRoom = &types.RoomIndexData{Vnum: 1, Name: "Test"}

	DoNote(alice, "list")
	out := readOutput(alice, aliceClient)

	if !strings.Contains(out, "ForAll") {
		t.Errorf("Alice should see 'all' note, got: %q", out)
	}
	if !strings.Contains(out, "ForAlice") {
		t.Errorf("Alice should see note addressed to her, got: %q", out)
	}
	if strings.Contains(out, "ForCarol") {
		t.Errorf("Alice should NOT see Carol's note, got: %q", out)
	}
}

func TestDoNote_ReadRefusesOtherRecipient(t *testing.T) {
	w := setupClanWorld()
	board := &types.BoardData{
		Notes: []*types.NoteData{
			{Sender: "Admin", Subject: "Private", Date: "d", ToList: "Carol", Text: "secret\n\r"},
		},
	}
	w.Boards = append(w.Boards, board)

	alice, client := makeTestChar("Alice")
	defer client.Close()
	alice.InRoom = &types.RoomIndexData{Vnum: 1, Name: "Test"}

	DoNote(alice, "read 1")
	out := readOutput(alice, client)
	if !strings.Contains(out, "not addressed to you") {
		t.Errorf("expected refusal for foreign note, got: %q", out)
	}
	if strings.Contains(out, "secret") {
		t.Errorf("secret text leaked to wrong reader, got: %q", out)
	}
}

func TestCountNotesFor(t *testing.T) {
	w := setupClanWorld()
	w.Boards = []*types.BoardData{
		{Notes: []*types.NoteData{
			{ToList: "all"},
			{ToList: "Alice"},
			{ToList: "Bob"},
		}},
		{Notes: []*types.NoteData{
			{ToList: "Alice Bob"},
		}},
	}
	alice := &types.CharData{Name: "Alice", Trust: 0, Level: 10}
	got := CountNotesFor(alice)
	if got != 3 {
		t.Errorf("CountNotesFor(Alice) = %d; want 3 (all, Alice, Alice Bob)", got)
	}
}
