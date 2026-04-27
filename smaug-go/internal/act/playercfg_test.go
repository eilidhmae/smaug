package act

import (
	"bytes"
	"log"
	"net"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/eilidhmae/smaug/internal/command"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// -------------- G1: DoSave --------------

func TestDoSave_NPCIsNoop(t *testing.T) {
	ch, client := makeTestChar("Grizzly")
	defer client.Close()
	ch.Act.Set(types.ACT_IS_NPC)

	called := false
	prev := SaveFunc
	SaveFunc = func(c *types.CharData) { called = true }
	defer func() { SaveFunc = prev }()

	DoSave(ch, "")
	if called {
		t.Errorf("NPC DoSave must not invoke SaveFunc")
	}
	if out := readOutput(ch, client); out != "" {
		t.Errorf("NPC DoSave must emit nothing; got %q", out)
	}
}

func TestDoSave_LevelOneRejected(t *testing.T) {
	ch, client := makeTestChar("Newbie")
	defer client.Close()
	ch.Level = 1

	called := false
	prev := SaveFunc
	SaveFunc = func(c *types.CharData) { called = true }
	defer func() { SaveFunc = prev }()

	DoSave(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "second level") {
		t.Errorf("level-1 should get rejection message; got %q", out)
	}
	if called {
		t.Errorf("level-1 DoSave must not invoke SaveFunc")
	}
}

func TestDoSave_LevelTwoSaves(t *testing.T) {
	ch, client := makeTestChar("Veteran")
	defer client.Close()
	ch.Level = 2

	called := 0
	var sawCh *types.CharData
	prev := SaveFunc
	SaveFunc = func(c *types.CharData) {
		called++
		sawCh = c
	}
	defer func() { SaveFunc = prev }()

	DoSave(ch, "")
	out := readOutput(ch, client)
	if called != 1 {
		t.Errorf("SaveFunc should be invoked exactly once; got %d", called)
	}
	if sawCh != ch {
		t.Errorf("SaveFunc should receive ch; got %v", sawCh)
	}
	if !strings.Contains(out, "Saved") {
		t.Errorf("expected Saved... message; got %q", out)
	}
	if ch.Wait != 2 {
		t.Errorf("ch.Wait should be 2 after save; got %d", ch.Wait)
	}
}

func TestDoSave_NilSaveFuncDoesNotPanic(t *testing.T) {
	ch, client := makeTestChar("NoHook")
	defer client.Close()
	ch.Level = 5

	prev := SaveFunc
	SaveFunc = nil
	defer func() { SaveFunc = prev }()

	// Must not panic even when SaveFunc is unset.
	DoSave(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Saved") {
		t.Errorf("DoSave with nil SaveFunc should still emit message; got %q", out)
	}
}

// -------------- G2: DoAfk --------------

func TestDoAfk_TogglesOn(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 10001}

	ch, client := makeTestChar("Awake")
	defer client.Close()
	handler.CharToRoom(ch, room)

	if ch.Act.IsSet(types.PLR_AFK) {
		t.Fatal("PLR_AFK should start unset")
	}

	DoAfk(ch, "")
	if !ch.Act.IsSet(types.PLR_AFK) {
		t.Errorf("PLR_AFK should be set after toggle")
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "You are now afk") {
		t.Errorf("expected self-message 'You are now afk'; got %q", out)
	}
}

func TestDoAfk_TogglesOff(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 10002}

	ch, client := makeTestChar("Returning")
	defer client.Close()
	handler.CharToRoom(ch, room)
	ch.Act.Set(types.PLR_AFK)

	DoAfk(ch, "")
	if ch.Act.IsSet(types.PLR_AFK) {
		t.Errorf("PLR_AFK should be cleared after toggle")
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "no longer afk") {
		t.Errorf("expected self-message 'no longer afk'; got %q", out)
	}
}

func TestDoAfk_NPCIsNoop(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 10003}

	ch, client := makeTestChar("Mob")
	defer client.Close()
	ch.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(ch, room)

	DoAfk(ch, "")
	// NPC should neither get PLR_AFK set nor see a message.
	if ch.Act.IsSet(types.PLR_AFK) {
		t.Errorf("NPC DoAfk must not set PLR_AFK")
	}
	if out := readOutput(ch, client); out != "" {
		t.Errorf("NPC DoAfk must emit nothing; got %q", out)
	}
}

func TestDoAfk_BroadcastsToRoom(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 10004}

	afker, ac := makeTestChar("Drifter")
	defer ac.Close()
	handler.CharToRoom(afker, room)

	observer, oc := makeTestChar("Observer")
	defer oc.Close()
	handler.CharToRoom(observer, room)

	DoAfk(afker, "")
	_ = readOutput(afker, ac)
	got := readOutput(observer, oc)
	if !strings.Contains(got, "is now afk") {
		t.Errorf("observer should see 'is now afk' broadcast; got %q", got)
	}
	if !strings.Contains(got, "Drifter") {
		t.Errorf("broadcast should include AFKer's name; got %q", got)
	}
}

// Regression — the existing DoTell AFK-prefix behavior must still fire.
func TestDoAfk_DoTellPrefixRegression(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 10005}

	sender, sc := makeTestChar("Alice")
	defer sc.Close()
	handler.CharToRoom(sender, room)

	recipient, rc := makeTestChar("Bob")
	defer rc.Close()
	handler.CharToRoom(recipient, room)
	w.Characters = []*types.CharData{sender, recipient}

	// Flip Bob into AFK via DoAfk, not by direct flag manipulation.
	DoAfk(recipient, "")
	_ = readOutput(recipient, rc)

	DoTell(sender, "bob hi")
	_ = readOutput(sender, sc)
	bobOut := readOutput(recipient, rc)
	if !strings.Contains(bobOut, "(afk)") {
		t.Errorf("expected (afk) prefix from DoTell; got %q", bobOut)
	}
}

// -------------- G3: DoTitle --------------

func TestDoTitle_SetsTitleWithLeadingSpace(t *testing.T) {
	ch, client := makeTestChar("Hero")
	defer client.Close()

	DoTitle(ch, "the Slayer")
	out := readOutput(ch, client)
	if ch.PCData.Title != " the Slayer" {
		t.Errorf("Title = %q, want %q", ch.PCData.Title, " the Slayer")
	}
	if !strings.Contains(out, "Your new title has been set") {
		t.Errorf("expected confirmation; got %q", out)
	}
}

func TestDoTitle_PunctuationStartNoLeadingSpace(t *testing.T) {
	ch, client := makeTestChar("Hero")
	defer client.Close()

	DoTitle(ch, ", Scourge of Evil")
	_ = readOutput(ch, client)
	if ch.PCData.Title != ", Scourge of Evil" {
		t.Errorf("punctuation-first title should not get leading space; got %q", ch.PCData.Title)
	}
}

func TestDoTitle_DigitStartGetsLeadingSpace(t *testing.T) {
	ch, client := makeTestChar("Hero")
	defer client.Close()

	DoTitle(ch, "1st of the 1st")
	_ = readOutput(ch, client)
	if ch.PCData.Title != " 1st of the 1st" {
		t.Errorf("digit-first title should get leading space; got %q", ch.PCData.Title)
	}
}

func TestDoTitle_EmptyArgument(t *testing.T) {
	ch, client := makeTestChar("Hero")
	defer client.Close()
	ch.PCData.Title = "the Original"

	DoTitle(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Change your title to what?") {
		t.Errorf("expected empty-arg error; got %q", out)
	}
	if ch.PCData.Title != "the Original" {
		t.Errorf("empty arg should not modify title; got %q", ch.PCData.Title)
	}
}

func TestDoTitle_NOTITLEBlocked(t *testing.T) {
	ch, client := makeTestChar("Gagged")
	defer client.Close()
	ch.PCData.Flags |= int(types.PCFLAG_NOTITLE)
	ch.PCData.Title = "the Silent"

	DoTitle(ch, "the Bold")
	out := readOutput(ch, client)
	if !strings.Contains(out, "prohibit") {
		t.Errorf("PCFLAG_NOTITLE should block with prohibit msg; got %q", out)
	}
	if ch.PCData.Title != "the Silent" {
		t.Errorf("NOTITLE must not mutate title; got %q", ch.PCData.Title)
	}
}

func TestDoTitle_NPCNoop(t *testing.T) {
	ch, client := makeTestChar("Mob")
	defer client.Close()
	ch.Act.Set(types.ACT_IS_NPC)
	originalTitle := ch.PCData.Title

	DoTitle(ch, "the Wizard")
	if ch.PCData.Title != originalTitle {
		t.Errorf("NPC title must not change; got %q", ch.PCData.Title)
	}
	if out := readOutput(ch, client); out != "" {
		t.Errorf("NPC DoTitle must emit nothing; got %q", out)
	}
}

func TestDoTitle_Truncates50Chars(t *testing.T) {
	ch, client := makeTestChar("Verbose")
	defer client.Close()

	long := strings.Repeat("a", 75)
	DoTitle(ch, long)
	_ = readOutput(ch, client)
	// The 'a'-starting title gets a leading space prepended by setTitle.
	// Truncation happens BEFORE SmashColorToken/setTitle, so the stored
	// title is " " + first-50-of-argument.
	wantContent := strings.Repeat("a", 50)
	if ch.PCData.Title != " "+wantContent {
		t.Errorf("Title truncation wrong; got %q (len=%d)", ch.PCData.Title, len(ch.PCData.Title))
	}
}

func TestDoTitle_TildeScrubbed(t *testing.T) {
	ch, client := makeTestChar("Knight")
	defer client.Close()

	DoTitle(ch, "the~Warlord")
	_ = readOutput(ch, client)
	if strings.Contains(ch.PCData.Title, "~") {
		t.Errorf("tildes should be smashed; got %q", ch.PCData.Title)
	}
	if !strings.Contains(ch.PCData.Title, "the-Warlord") {
		t.Errorf("tilde should become dash; got %q", ch.PCData.Title)
	}
}

func TestDoTitle_ColorTokenScrubbed(t *testing.T) {
	ch, client := makeTestChar("Colorist")
	defer client.Close()

	DoTitle(ch, "&Rthe ^BRed Baron")
	_ = readOutput(ch, client)
	if strings.Contains(ch.PCData.Title, "&") {
		t.Errorf("& color tokens should be scrubbed; got %q", ch.PCData.Title)
	}
	if strings.Contains(ch.PCData.Title, "^") {
		t.Errorf("^ color tokens should be scrubbed; got %q", ch.PCData.Title)
	}
	// + and - are the expected replacements.
	if !strings.Contains(ch.PCData.Title, "+R") || !strings.Contains(ch.PCData.Title, "-B") {
		t.Errorf("color tokens should become +/-; got %q", ch.PCData.Title)
	}
}

// -------------- G4: DoPassword --------------

// setupPwChar creates a test PC and stores the bcrypt hash of `currentPwd` in
// ch.PCData.Pwd so DoPassword's verify path has real data to compare against.
func setupPwChar(t *testing.T, name, currentPwd string) (*types.CharData, net.Conn) {
	t.Helper()
	ch, client := makeTestChar(name)
	hash, err := bcrypt.GenerateFromPassword([]byte(currentPwd), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("seed GenerateFromPassword: %v", err)
	}
	ch.PCData.Pwd = string(hash)
	return ch, client
}

// lowerCostForTest sets BcryptCost to MinCost for the duration of the test
// and restores it afterward. Mirrors the game.BcryptCost pattern in loop_test.
func lowerCostForTest(t *testing.T) {
	t.Helper()
	prev := BcryptCost
	BcryptCost = bcrypt.MinCost
	t.Cleanup(func() { BcryptCost = prev })
}

func TestDoPassword_Success(t *testing.T) {
	lowerCostForTest(t)
	ch, client := setupPwChar(t, "Alice", "oldpass123")
	defer client.Close()

	saved := 0
	prev := SaveFunc
	SaveFunc = func(c *types.CharData) { saved++ }
	defer func() { SaveFunc = prev }()

	DoPassword(ch, "oldpass123 newpass456 newpass456")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Ok.") {
		t.Errorf("expected Ok.; got %q", out)
	}
	if !strings.HasPrefix(ch.PCData.Pwd, "$2") {
		t.Errorf("Pwd should be a bcrypt hash ($2...); got %q", ch.PCData.Pwd)
	}
	// Verify the new hash matches newpass456.
	if err := bcrypt.CompareHashAndPassword([]byte(ch.PCData.Pwd), []byte("newpass456")); err != nil {
		t.Errorf("new hash doesn't verify against new password: %v", err)
	}
	if saved != 1 {
		t.Errorf("SaveFunc should be called once; got %d", saved)
	}
}

func TestDoPassword_WrongOld(t *testing.T) {
	lowerCostForTest(t)
	ch, client := setupPwChar(t, "Alice", "oldpass123")
	defer client.Close()
	origHash := ch.PCData.Pwd

	DoPassword(ch, "WRONGold newpass456 newpass456")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Wrong password") {
		t.Errorf("expected 'Wrong password'; got %q", out)
	}
	if ch.PCData.Pwd != origHash {
		t.Errorf("Pwd should be unchanged on wrong old; got %q", ch.PCData.Pwd)
	}
}

func TestDoPassword_MismatchNew(t *testing.T) {
	lowerCostForTest(t)
	ch, client := setupPwChar(t, "Alice", "oldpass123")
	defer client.Close()
	origHash := ch.PCData.Pwd

	DoPassword(ch, "oldpass123 newpass456 differentA")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Passwords don't match") {
		t.Errorf("expected mismatch message; got %q", out)
	}
	if ch.PCData.Pwd != origHash {
		t.Errorf("Pwd should be unchanged on mismatch; got %q", ch.PCData.Pwd)
	}
}

func TestDoPassword_TooShortNew(t *testing.T) {
	lowerCostForTest(t)
	ch, client := setupPwChar(t, "Alice", "oldpass123")
	defer client.Close()
	origHash := ch.PCData.Pwd

	DoPassword(ch, "oldpass123 short short")
	out := readOutput(ch, client)
	if !strings.Contains(out, "six characters") {
		t.Errorf("expected six-character minimum error; got %q", out)
	}
	if ch.PCData.Pwd != origHash {
		t.Errorf("Pwd should be unchanged on too-short; got %q", ch.PCData.Pwd)
	}
}

func TestDoPassword_NoArgs(t *testing.T) {
	lowerCostForTest(t)
	ch, client := setupPwChar(t, "Alice", "oldpass123")
	defer client.Close()
	origHash := ch.PCData.Pwd

	DoPassword(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Syntax: password") {
		t.Errorf("expected syntax message; got %q", out)
	}
	if ch.PCData.Pwd != origHash {
		t.Errorf("Pwd should be unchanged on no args; got %q", ch.PCData.Pwd)
	}
}

func TestDoPassword_PartialArgs(t *testing.T) {
	lowerCostForTest(t)
	ch, client := setupPwChar(t, "Alice", "oldpass123")
	defer client.Close()

	DoPassword(ch, "oldpass123 newpass456")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Syntax: password") {
		t.Errorf("partial args should print syntax; got %q", out)
	}
}

func TestDoPassword_NPCIsNoop(t *testing.T) {
	lowerCostForTest(t)
	ch, client := setupPwChar(t, "Mob", "oldpass123")
	defer client.Close()
	ch.Act.Set(types.ACT_IS_NPC)
	origHash := ch.PCData.Pwd

	DoPassword(ch, "oldpass123 newpass456 newpass456")
	if ch.PCData.Pwd != origHash {
		t.Errorf("NPC Pwd must not change; got %q", ch.PCData.Pwd)
	}
}

func TestDoPassword_LogsChangeEvent(t *testing.T) {
	lowerCostForTest(t)
	ch, client := setupPwChar(t, "Alice", "oldpass123")
	defer client.Close()
	prev := SaveFunc
	SaveFunc = func(c *types.CharData) {}
	defer func() { SaveFunc = prev }()

	// Redirect the default logger so we can assert on the emitted line.
	var buf bytes.Buffer
	prevW := log.Writer()
	prevFlags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	defer func() {
		log.SetOutput(prevW)
		log.SetFlags(prevFlags)
	}()

	DoPassword(ch, "oldpass123 newpass456 newpass456")
	_ = readOutput(ch, client)
	logged := buf.String()
	if !strings.Contains(logged, "Alice") || !strings.Contains(logged, "changing password") {
		t.Errorf("expected log line to mention 'Alice' + 'changing password'; got %q", logged)
	}
}

// Case-preservation: a capital letter in the new password must survive the
// case_argument parse AND verify against the stored bcrypt hash afterwards.
func TestDoPassword_CasePreserved(t *testing.T) {
	lowerCostForTest(t)
	ch, client := setupPwChar(t, "Alice", "OldPassCaps")
	defer client.Close()

	DoPassword(ch, "OldPassCaps NewPassCaps NewPassCaps")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Ok.") {
		t.Errorf("expected Ok. with case-preserved args; got %q", out)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(ch.PCData.Pwd), []byte("NewPassCaps")); err != nil {
		t.Errorf("case-preserved new hash must verify; err=%v", err)
	}
}

// -------------- DoGag (plan-do-gag.md) --------------

func TestDoGag_TogglesOn(t *testing.T) {
	ch, client := makeTestChar("Silent")
	defer client.Close()

	if ch.PCData.Flags&int(types.PCFLAG_GAG) != 0 {
		t.Fatal("PCFLAG_GAG should start unset")
	}

	DoGag(ch, "")
	if ch.PCData.Flags&int(types.PCFLAG_GAG) == 0 {
		t.Errorf("PCFLAG_GAG should be set after toggle")
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "Combat messages will be gagged.") {
		t.Errorf("expected self-message 'Combat messages will be gagged.'; got %q", out)
	}
}

func TestDoGag_TogglesOff(t *testing.T) {
	ch, client := makeTestChar("Chatty")
	defer client.Close()
	ch.PCData.Flags |= int(types.PCFLAG_GAG)

	DoGag(ch, "")
	if ch.PCData.Flags&int(types.PCFLAG_GAG) != 0 {
		t.Errorf("PCFLAG_GAG should be cleared after toggle")
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "Combat messages will no longer be gagged.") {
		t.Errorf("expected self-message 'Combat messages will no longer be gagged.'; got %q", out)
	}
}

func TestDoGag_NPCIsNoop(t *testing.T) {
	ch, client := makeTestChar("Mob")
	defer client.Close()
	ch.Act.Set(types.ACT_IS_NPC)

	// NPC should neither panic nor emit output. Running DoGag on an NPC with
	// PCData still present proves the IsNPC gate fires first.
	DoGag(ch, "")
	if out := readOutput(ch, client); out != "" {
		t.Errorf("NPC DoGag must emit nothing; got %q", out)
	}
}

func TestDoGag_NilPCDataIsNoop(t *testing.T) {
	ch, client := makeTestChar("Corrupt")
	defer client.Close()
	// Non-NPC (ACT_IS_NPC not set) but PCData == nil: simulates a malformed PC.
	// The nil-PCData guard is load-bearing because IsNPC() only checks Act.
	ch.PCData = nil

	// Must not panic.
	DoGag(ch, "")
	if out := readOutput(ch, client); out != "" {
		t.Errorf("nil-PCData DoGag must emit nothing; got %q", out)
	}
}

// -------------- DoBlank (plan-phase6-quickwins-blank-pcrename.md §G2) --------------

func TestDoBlank_TogglesOn(t *testing.T) {
	ch, client := makeTestChar("Spaced")
	defer client.Close()

	if ch.Act.IsSet(types.PLR_BLANK) {
		t.Fatal("PLR_BLANK should start unset")
	}
	DoBlank(ch, "")
	if !ch.Act.IsSet(types.PLR_BLANK) {
		t.Errorf("PLR_BLANK should be set after toggle")
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "Blank lines will be inserted before each prompt.") {
		t.Errorf("expected toggle-on message; got %q", out)
	}
}

func TestDoBlank_TogglesOff(t *testing.T) {
	ch, client := makeTestChar("Crowded")
	defer client.Close()
	ch.Act.Set(types.PLR_BLANK)

	DoBlank(ch, "")
	if ch.Act.IsSet(types.PLR_BLANK) {
		t.Errorf("PLR_BLANK should be cleared after toggle")
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "Blank lines will no longer be inserted before each prompt.") {
		t.Errorf("expected toggle-off message; got %q", out)
	}
}

func TestDoBlank_NPCIsNoop(t *testing.T) {
	ch, client := makeTestChar("Mob")
	defer client.Close()
	ch.Act.Set(types.ACT_IS_NPC)

	DoBlank(ch, "")
	// PLR_BLANK and ACT_IS_NPC share the same Act bitvector but different
	// bit positions, so we can verify the NPC gate by checking no toggle
	// happened on PLR_BLANK and no output emitted.
	if ch.Act.IsSet(types.PLR_BLANK) {
		t.Errorf("NPC DoBlank must not toggle PLR_BLANK")
	}
	if out := readOutput(ch, client); out != "" {
		t.Errorf("NPC DoBlank must emit nothing; got %q", out)
	}
}

func TestDoBlank_NilPCDataIsNoop(t *testing.T) {
	ch, client := makeTestChar("Hollow")
	defer client.Close()
	ch.PCData = nil

	// Must not panic and must not toggle.
	DoBlank(ch, "")
	if ch.Act.IsSet(types.PLR_BLANK) {
		t.Errorf("nil-PCData DoBlank must not toggle PLR_BLANK")
	}
	if out := readOutput(ch, client); out != "" {
		t.Errorf("nil-PCData DoBlank must emit nothing; got %q", out)
	}
}

// -------------- DoBio --------------

// TestDoBio_SetsSubstateAndEditorSave verifies DoBio installs the
// SUB_PERSONAL_BIO substate and an EditorSave closure, then invokes
// StartEditingFunc with the current PCData.Bio as seed text.
func TestDoBio_SetsSubstateAndEditorSave(t *testing.T) {
	ch, client := makeTestChar("Wanderer")
	defer client.Close()
	ch.Level = 10
	ch.PCData.Bio = "Old bio text"

	var seedText string
	var editorSaveAtCall func(*types.CharData)
	prev := StartEditingFunc
	StartEditingFunc = func(c *types.CharData, text string) {
		seedText = text
		editorSaveAtCall = c.EditorSave
	}
	defer func() { StartEditingFunc = prev }()

	DoBio(ch, "")

	if ch.Substate != types.SUB_PERSONAL_BIO {
		t.Errorf("Substate = %d, want SUB_PERSONAL_BIO (%d)", ch.Substate, types.SUB_PERSONAL_BIO)
	}
	if editorSaveAtCall == nil {
		t.Fatal("EditorSave must be non-nil when StartEditingFunc is called")
	}
	if seedText != "Old bio text" {
		t.Errorf("seed text = %q, want %q", seedText, "Old bio text")
	}
}

// TestDoBio_NPCIsNoop — NPCs cannot set a bio (C player.c:3415-3419).
func TestDoBio_NPCIsNoop(t *testing.T) {
	ch, client := makeTestChar("MobBio")
	defer client.Close()
	ch.Act.Set(types.ACT_IS_NPC)
	ch.Level = 10

	called := false
	prev := StartEditingFunc
	StartEditingFunc = func(c *types.CharData, text string) { called = true }
	defer func() { StartEditingFunc = prev }()

	DoBio(ch, "")
	if called {
		t.Error("NPC DoBio must not invoke StartEditingFunc")
	}
}

// TestDoBio_NODescFlagBlocks — PCFLAG_NOBIO blocks the command (C
// player.c:3428-3433).
func TestDoBio_PCFLAG_NOBIO_Blocked(t *testing.T) {
	ch, client := makeTestChar("Blocked")
	defer client.Close()
	ch.Level = 10
	ch.PCData.Flags |= int(types.PCFLAG_NOBIO)

	called := false
	prev := StartEditingFunc
	StartEditingFunc = func(c *types.CharData, text string) { called = true }
	defer func() { StartEditingFunc = prev }()

	DoBio(ch, "")
	if called {
		t.Error("PCFLAG_NOBIO DoBio must not invoke StartEditingFunc")
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "gods") && !strings.Contains(out, "allow") {
		t.Errorf("expected gods-deny message; got %q", out)
	}
}

// TestDoBio_EditorSaveClosureUpdatesBio — verify the closure written by
// DoBio actually mutates PCData.Bio when invoked by the editor /s path.
func TestDoBio_EditorSaveClosureUpdatesBio(t *testing.T) {
	ch, client := makeTestChar("Scribe")
	defer client.Close()
	ch.Level = 10
	ch.PCData.Bio = "stale"

	expected := "A daring wanderer from distant lands.\n\r"
	prevCopy := CopyBufferFunc
	prevStop := StopEditingFunc
	prevStart := StartEditingFunc
	defer func() {
		CopyBufferFunc = prevCopy
		StopEditingFunc = prevStop
		StartEditingFunc = prevStart
	}()
	CopyBufferFunc = func(c *types.CharData) string { return expected }
	StopEditingFunc = func(c *types.CharData) {
		c.EditorSave = nil
		if c.Desc != nil {
			c.Desc.Connected = types.CON_PLAYING
		}
	}
	StartEditingFunc = func(c *types.CharData, text string) {
		if c.Desc != nil {
			c.Desc.Connected = types.CON_EDITING
		}
	}

	DoBio(ch, "")
	if ch.EditorSave == nil {
		t.Fatal("EditorSave must be set after DoBio")
	}

	// Simulate `/s` path.
	ch.EditorSave(ch)

	if ch.PCData.Bio != expected {
		t.Errorf("PCData.Bio = %q, want %q", ch.PCData.Bio, expected)
	}
}

// -------------- DoDescription --------------

func TestDoDescription_SetsSubstateAndEditorSave(t *testing.T) {
	ch, client := makeTestChar("Traveler")
	defer client.Close()
	ch.Description = "An old description"

	var seedText string
	var editorSaveAtCall func(*types.CharData)
	prev := StartEditingFunc
	StartEditingFunc = func(c *types.CharData, text string) {
		seedText = text
		editorSaveAtCall = c.EditorSave
	}
	defer func() { StartEditingFunc = prev }()

	DoDescription(ch, "")

	if ch.Substate != types.SUB_PERSONAL_DESC {
		t.Errorf("Substate = %d, want SUB_PERSONAL_DESC (%d)", ch.Substate, types.SUB_PERSONAL_DESC)
	}
	if editorSaveAtCall == nil {
		t.Fatal("EditorSave must be non-nil when StartEditingFunc is called")
	}
	if seedText != "An old description" {
		t.Errorf("seed text = %q, want %q", seedText, "An old description")
	}
}

func TestDoDescription_NPCIsNoop(t *testing.T) {
	ch, client := makeTestChar("MobDesc")
	defer client.Close()
	ch.Act.Set(types.ACT_IS_NPC)

	called := false
	prev := StartEditingFunc
	StartEditingFunc = func(c *types.CharData, text string) { called = true }
	defer func() { StartEditingFunc = prev }()

	DoDescription(ch, "")
	if called {
		t.Error("NPC DoDescription must not invoke StartEditingFunc")
	}
}

// TestDoDescription_PCFLAG_NODESC_Blocked — C player.c:3374-3378.
func TestDoDescription_PCFLAG_NODESC_Blocked(t *testing.T) {
	ch, client := makeTestChar("BlockedDesc")
	defer client.Close()
	ch.PCData.Flags |= int(types.PCFLAG_NODESC)

	called := false
	prev := StartEditingFunc
	StartEditingFunc = func(c *types.CharData, text string) { called = true }
	defer func() { StartEditingFunc = prev }()

	DoDescription(ch, "")
	if called {
		t.Error("PCFLAG_NODESC DoDescription must not invoke StartEditingFunc")
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "cannot") {
		t.Errorf("expected 'cannot' message; got %q", out)
	}
}

func TestDoDescription_EditorSaveClosureUpdatesDescription(t *testing.T) {
	ch, client := makeTestChar("Narrator")
	defer client.Close()
	ch.Description = "placeholder"

	expected := "Tall, with a scar across the left cheek.\n\r"
	prevCopy := CopyBufferFunc
	prevStop := StopEditingFunc
	prevStart := StartEditingFunc
	defer func() {
		CopyBufferFunc = prevCopy
		StopEditingFunc = prevStop
		StartEditingFunc = prevStart
	}()
	CopyBufferFunc = func(c *types.CharData) string { return expected }
	StopEditingFunc = func(c *types.CharData) {
		c.EditorSave = nil
		if c.Desc != nil {
			c.Desc.Connected = types.CON_PLAYING
		}
	}
	StartEditingFunc = func(c *types.CharData, text string) {
		if c.Desc != nil {
			c.Desc.Connected = types.CON_EDITING
		}
	}

	DoDescription(ch, "")
	if ch.EditorSave == nil {
		t.Fatal("EditorSave must be set after DoDescription")
	}

	ch.EditorSave(ch)

	if ch.Description != expected {
		t.Errorf("Description = %q, want %q", ch.Description, expected)
	}
}

// -------------- G5: pagelen alias (dispatcher-level) --------------

// TestInterpret_PagelenAlias checks that registering `pagelen` alongside
// `pager` in a command registry dispatches `pagelen 40` into DoPager with
// the expected side-effect on PCData.PagerLen. The registry wiring happens
// in internal/boot, so this test exercises the generic Register() +
// Interpret() path rather than the production registry.
func TestInterpret_PagelenAlias(t *testing.T) {
	reg := command.NewRegistry()
	reg.Register(&command.Command{Name: "pager", DoFun: DoPager, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "pagelen", DoFun: DoPager, Position: types.POS_DEAD, Level: 0})

	ch, client := makeTestChar("Reader")
	defer client.Close()
	// Start with a distinct value so "set to 40" is observable.
	ch.PCData.PagerLen = 25

	reg.Interpret(ch, "pagelen 40")
	if ch.PCData.PagerLen != 40 {
		t.Errorf("after pagelen 40, PagerLen = %d, want 40", ch.PCData.PagerLen)
	}
}
