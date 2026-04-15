package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
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

// --- G8: mail targeting ---

func TestIsNoteTo_AllRecipient(t *testing.T) {
	ch := &types.CharData{Name: "Alice"}
	note := &types.NoteData{ToList: "all"}
	if !isNoteTo(ch, note) {
		t.Errorf("note 'all' should be visible to everyone")
	}
}

func TestIsNoteTo_RecipientMatch(t *testing.T) {
	alice := &types.CharData{Name: "Alice"}
	bob := &types.CharData{Name: "Bob"}
	note := &types.NoteData{ToList: "Alice Carol"}
	if !isNoteTo(alice, note) {
		t.Errorf("Alice should receive the note")
	}
	if isNoteTo(bob, note) {
		t.Errorf("Bob should not receive the note")
	}
}

func TestIsNoteTo_SenderAlwaysSees(t *testing.T) {
	ch := &types.CharData{Name: "Alice"}
	note := &types.NoteData{Sender: "Alice", ToList: "Bob"}
	if !isNoteTo(ch, note) {
		t.Errorf("sender should see their own note")
	}
}

func TestIsNoteTo_HolylightSeesAll(t *testing.T) {
	ch := &types.CharData{Name: "Imm", PCData: &types.PCData{}}
	ch.Act.Set(types.PLR_HOLYLIGHT)
	note := &types.NoteData{Sender: "Bob", ToList: "Carol"}
	if !isNoteTo(ch, note) {
		t.Errorf("holylight should see every note")
	}
}

func TestDoNote_ListFiltersByRecipient(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	board := &types.BoardData{
		Notes: []*types.NoteData{
			{Sender: "Bob", Subject: "Hi Alice", ToList: "Alice"},
			{Sender: "Carol", Subject: "Hi Everyone", ToList: "all"},
			{Sender: "Dave", Subject: "Hi Bob", ToList: "Bob"},
		},
	}
	w.Boards = append(w.Boards, board)
	room := &types.RoomIndexData{Vnum: 1000}

	ch, client := makeTestChar("Alice")
	defer client.Close()
	handler.CharToRoom(ch, room)

	DoNote(ch, "list")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Hi Alice") {
		t.Errorf("should see note to Alice; got %q", out)
	}
	if !strings.Contains(out, "Hi Everyone") {
		t.Errorf("should see note to all; got %q", out)
	}
	if strings.Contains(out, "Hi Bob") {
		t.Errorf("should NOT see note addressed only to Bob; got %q", out)
	}
}

func TestUnreadNotesFor_Counts(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	board := &types.BoardData{
		Notes: []*types.NoteData{
			{Sender: "Bob", ToList: "Alice"},
			{Sender: "Carol", ToList: "all"},
			{Sender: "Dave", ToList: "Bob Frank"},
		},
	}
	w.Boards = append(w.Boards, board)

	alice := &types.CharData{Name: "Alice"}
	if got := UnreadNotesFor(alice); got != 2 {
		t.Errorf("alice should see 2 notes; got %d", got)
	}
	bob := &types.CharData{Name: "Bob"}
	// Bob sent note 1, reads note 2 (all), and is named in note 3's list.
	if got := UnreadNotesFor(bob); got != 3 {
		t.Errorf("bob should see 3 notes (one sent, two addressed); got %d", got)
	}
}

// --- G9: clan storerooms ---

func TestDoClanDeposit_RequiresMembership(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	ch, client := makeTestChar("Loner")
	defer client.Close()
	DoClanDeposit(ch, "sword")
	out := readOutput(ch, client)
	if !strings.Contains(out, "aren't in a clan") {
		t.Errorf("expected clan requirement; got %q", out)
	}
}

func TestDoClanDepositWithdraw_RoundTrip(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	store := &types.RoomIndexData{Vnum: 9600, Name: "Clan Vault"}
	w.Rooms[store.Vnum] = store

	clan := &types.ClanData{Name: "Nightwatch", Storeroom: 9600, Leader: "Ranger"}
	w.Clans = append(w.Clans, clan)

	hall := &types.RoomIndexData{Vnum: 9601}
	ch, client := makeTestChar("Ranger")
	defer client.Close()
	ch.PCData.Clan = clan
	ch.PCData.ClanName = clan.Name
	handler.CharToRoom(ch, hall)

	obj := &types.ObjData{Name: "bow", ShortDescr: "a longbow", WearLoc: types.WEAR_NONE}
	handler.ObjToChar(obj, ch)

	DoClanDeposit(ch, "bow")
	_ = readOutput(ch, client)
	if obj.InRoom != store {
		t.Fatalf("bow should be in storeroom; is at %+v / carried=%v", obj.InRoom, obj.CarriedBy)
	}

	DoClanWithdraw(ch, "bow")
	_ = readOutput(ch, client)
	if obj.CarriedBy != ch {
		t.Errorf("bow should be back in inventory; carriedby=%v", obj.CarriedBy)
	}
}

func TestDoClanWithdraw_RecruitBlocked(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	store := &types.RoomIndexData{Vnum: 9604, Name: "Vault"}
	w.Rooms[store.Vnum] = store
	clan := &types.ClanData{Name: "Order", Storeroom: 9604, Leader: "Captain"}

	recruit, client := makeTestChar("Footman")
	defer client.Close()
	recruit.PCData.Clan = clan
	handler.CharToRoom(recruit, &types.RoomIndexData{Vnum: 9605})
	// Put a bow in the storeroom so the non-permission failure is load-bearing.
	bow := &types.ObjData{Name: "bow", ShortDescr: "a bow"}
	handler.ObjToRoom(bow, store)

	DoClanWithdraw(recruit, "bow")
	out := readOutput(recruit, client)
	if !strings.Contains(out, "Only clan leaders") {
		t.Errorf("recruit should be refused; got %q", out)
	}
	if bow.InRoom != store {
		t.Errorf("bow should stay in storeroom")
	}
}

func TestDoClanWithdraw_EmptyStoreroom(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	store := &types.RoomIndexData{Vnum: 9602, Name: "Vault"}
	w.Rooms[store.Vnum] = store
	clan := &types.ClanData{Name: "Dawn", Storeroom: 9602, Leader: "Cleric"}

	ch, client := makeTestChar("Cleric")
	defer client.Close()
	ch.PCData.Clan = clan
	handler.CharToRoom(ch, &types.RoomIndexData{Vnum: 9603})

	DoClanWithdraw(ch, "sword")
	out := readOutput(ch, client)
	if !strings.Contains(out, "No sword") {
		t.Errorf("expected not-found message; got %q", out)
	}
}

// --- Adversary-follow-up coverage tests ---

func TestClanWithdrawAllowed_OfficersAndHolylight(t *testing.T) {
	clan := &types.ClanData{
		Name: "Keep", Leader: "Anne", Number1: "Beth", Number2: "Cora",
	}

	cases := []struct {
		name    string
		charFn  func(*types.CharData)
		allowed bool
	}{
		{"leader", func(c *types.CharData) { c.Name = "Anne" }, true},
		{"number1", func(c *types.CharData) { c.Name = "Beth" }, true},
		{"number2", func(c *types.CharData) { c.Name = "Cora" }, true},
		{"recruit", func(c *types.CharData) { c.Name = "Dawn" }, false},
		{"holylight immortal", func(c *types.CharData) {
			c.Name = "Imm"
			c.Act.Set(types.PLR_HOLYLIGHT)
		}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ch := &types.CharData{PCData: &types.PCData{Clan: clan}}
			tc.charFn(ch)
			if got := clanWithdrawAllowed(ch, clan); got != tc.allowed {
				t.Errorf("allowed = %v, want %v", got, tc.allowed)
			}
		})
	}
}
