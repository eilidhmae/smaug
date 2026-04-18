package util

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/types"
)

// --- test helpers ---

// makeChar returns a character attached to a net.Pipe-backed descriptor so the
// act() output can be captured and asserted on. The client side of the pipe is
// returned so callers can read what was sent.
func makeChar(t *testing.T, name string, sex int) (*types.CharData, net.Conn) {
	t.Helper()
	server, client := net.Pipe()
	d := &types.DescriptorData{Conn: server, InputQueue: make(chan string, 4)}
	ch := &types.CharData{
		Name:     name,
		Sex:      sex,
		Desc:     d,
		Position: types.POS_STANDING,
	}
	d.Character = ch
	return ch, client
}

// readOutput flushes ch's output buffer, reading concurrently from the pipe
// (net.Pipe is synchronous). Returns whatever was sent in a short window.
func readOutput(t *testing.T, ch *types.CharData, client net.Conn) string {
	t.Helper()
	if !ch.Desc.HasOutput() {
		return ""
	}
	result := make(chan string, 1)
	go func() {
		buf := make([]byte, 4096)
		_ = client.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		n, _ := client.Read(buf)
		result <- string(buf[:n])
	}()
	_ = ch.Desc.FlushOutput()
	return <-result
}

// putInRoom links chars into a shared room so TO_ROOM-style routing works.
// Returns the room so the caller can inspect/modify it.
func putInRoom(chs ...*types.CharData) *types.RoomIndexData {
	room := &types.RoomIndexData{Vnum: 1}
	for _, c := range chs {
		c.InRoom = room
		room.People = append(room.People, c)
	}
	return room
}

// ---------- ActFormat: token expansion ----------

func TestActFormat_N_TokenCapitalized(t *testing.T) {
	alice, _ := makeChar(t, "Alice", types.SEX_FEMALE)
	got := ActFormat("$n says hi", alice, alice, nil, nil, nil)
	want := "Alice says hi\n\r"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestActFormat_N_ToVictRecipientSees(t *testing.T) {
	alice, _ := makeChar(t, "Alice", types.SEX_FEMALE)
	bob, _ := makeChar(t, "Bob", types.SEX_MALE)
	got := ActFormat("$n says hi", bob, alice, bob, nil, nil)
	want := "Alice says hi\n\r"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestActFormat_BigN_VictimSubstitution(t *testing.T) {
	alice, _ := makeChar(t, "Alice", types.SEX_FEMALE)
	bob, _ := makeChar(t, "Bob", types.SEX_MALE)
	got := ActFormat("$N smiles", alice, alice, bob, nil, nil)
	want := "Bob smiles\n\r"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestActFormat_PronounE_Male(t *testing.T) {
	alice, _ := makeChar(t, "Alice", types.SEX_MALE)
	got := ActFormat("$e waves", nil, alice, nil, nil, nil)
	want := "He waves\n\r"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestActFormat_PronounM_Female(t *testing.T) {
	eve, _ := makeChar(t, "Eve", types.SEX_FEMALE)
	got := ActFormat("$m bites", nil, eve, nil, nil, nil)
	want := "Her bites\n\r"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestActFormat_PronounS_Male(t *testing.T) {
	alice, _ := makeChar(t, "Alice", types.SEX_MALE)
	got := ActFormat("$s sword", nil, alice, nil, nil, nil)
	want := "His sword\n\r"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestActFormat_PronounNeutral(t *testing.T) {
	thing, _ := makeChar(t, "thing", types.SEX_NEUTRAL)
	got := ActFormat("$e stands there.", nil, thing, nil, nil, nil)
	want := "It stands there.\n\r"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestActFormat_ObjectP(t *testing.T) {
	alice, _ := makeChar(t, "Alice", types.SEX_FEMALE)
	obj := &types.ObjData{ShortDescr: "a rusty sword"}
	got := ActFormat("$p glows", alice, alice, nil, obj, nil)
	want := "A rusty sword glows\n\r"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestActFormat_ObjectP_UsesIndexFallback(t *testing.T) {
	alice, _ := makeChar(t, "Alice", types.SEX_FEMALE)
	idx := &types.ObjIndexData{ShortDescr: "a generic thing"}
	obj := &types.ObjData{IndexData: idx}
	got := ActFormat("$p glows", alice, alice, nil, obj, nil)
	want := "A generic thing glows\n\r"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestActFormat_StringT(t *testing.T) {
	alice, _ := makeChar(t, "Alice", types.SEX_FEMALE)
	got := ActFormat("$t wobbles", alice, alice, nil, "the table", nil)
	want := "The table wobbles\n\r"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestActFormat_DirectionD(t *testing.T) {
	alice, _ := makeChar(t, "Alice", types.SEX_FEMALE)
	got := ActFormat("$n leaves $d", alice, alice, nil, nil, "north")
	want := "Alice leaves north\n\r"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestActFormat_DirectionD_EmptyFallback(t *testing.T) {
	alice, _ := makeChar(t, "Alice", types.SEX_FEMALE)
	got := ActFormat("$n leaves $d", alice, alice, nil, nil, "")
	want := "Alice leaves door\n\r"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestActFormat_LiteralDollar(t *testing.T) {
	got := ActFormat("$$ dollar", nil, nil, nil, nil, nil)
	want := "$ dollar\n\r"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestActFormat_QToken_SelfIsActor(t *testing.T) {
	alice, _ := makeChar(t, "Alice", types.SEX_FEMALE)
	// to == ch: "s" should become "" → "you wave"
	got := ActFormat("you wave$q", alice, alice, nil, nil, nil)
	want := "You wave\n\r"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestActFormat_QToken_Observer(t *testing.T) {
	alice, _ := makeChar(t, "Alice", types.SEX_FEMALE)
	bob, _ := makeChar(t, "Bob", types.SEX_MALE)
	// to != ch: "s" appended → "Alice waves"
	got := ActFormat("$n wave$q", bob, alice, nil, nil, nil)
	want := "Alice waves\n\r"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestActFormat_NPCShortDescr(t *testing.T) {
	mob := &types.CharData{
		Name:       "orc warrior guard",
		ShortDescr: "an orc warrior",
		Sex:        types.SEX_MALE,
	}
	mob.Act.Set(types.ACT_IS_NPC)
	got := ActFormat("$n growls", nil, mob, nil, nil, nil)
	want := "An orc warrior growls\n\r"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// ---------- ActFormat: visibility ----------

func TestActFormat_InvisibleActor(t *testing.T) {
	alice, _ := makeChar(t, "Alice", types.SEX_FEMALE)
	alice.AffectedBy.Set(types.AFF_INVISIBLE)
	bob, _ := makeChar(t, "Bob", types.SEX_MALE)
	got := ActFormat("$n smiles", bob, alice, nil, nil, nil)
	want := "Someone smiles\n\r"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestActFormat_InvisibleActor_DetectInvis(t *testing.T) {
	alice, _ := makeChar(t, "Alice", types.SEX_FEMALE)
	alice.AffectedBy.Set(types.AFF_INVISIBLE)
	bob, _ := makeChar(t, "Bob", types.SEX_MALE)
	bob.AffectedBy.Set(types.AFF_DETECT_INVIS)
	got := ActFormat("$n smiles", bob, alice, nil, nil, nil)
	want := "Alice smiles\n\r"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestActFormat_InvisibleObject(t *testing.T) {
	alice, _ := makeChar(t, "Alice", types.SEX_FEMALE)
	obj := &types.ObjData{ShortDescr: "a glowing gem"}
	obj.ExtraFlags.Set(types.ITEM_INVIS)
	got := ActFormat("$p sparkles", alice, alice, nil, obj, nil)
	want := "Something sparkles\n\r"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// TestActFormat_WizInvisLeakToMortal regression-guards the security fix:
// wiz-invisible immortals must not have their name leaked to lower-trust
// observers in act() output (C can_see PLR_WIZINVIS check).
func TestActFormat_WizInvisLeakToMortal(t *testing.T) {
	imm, _ := makeChar(t, "Immortal", types.SEX_FEMALE)
	imm.PCData = &types.PCData{WizInvis: 55}
	imm.Trust = 55
	mortal, _ := makeChar(t, "Mortal", types.SEX_MALE)
	mortal.Trust = 1

	got := ActFormat("$n casts a spell", mortal, imm, nil, nil, nil)
	want := "Someone casts a spell\n\r"
	if got != want {
		t.Fatalf("wiz-invis name leaked: got %q, want %q", got, want)
	}
}

// TestActFormat_WizInvisVisibleToHigherTrust confirms the complement: an
// observer whose Trust meets or exceeds WizInvis still sees the name.
func TestActFormat_WizInvisVisibleToHigherTrust(t *testing.T) {
	imm, _ := makeChar(t, "Immortal", types.SEX_FEMALE)
	imm.PCData = &types.PCData{WizInvis: 55}
	imm.Trust = 55
	peer, _ := makeChar(t, "Peer", types.SEX_MALE)
	peer.Trust = 60

	got := ActFormat("$n casts a spell", peer, imm, nil, nil, nil)
	want := "Immortal casts a spell\n\r"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// TestActFormat_HolyLightBypassesWizInvis confirms PLR_HOLYLIGHT observers
// see wiz-invis characters regardless of trust.
func TestActFormat_HolyLightBypassesWizInvis(t *testing.T) {
	imm, _ := makeChar(t, "Immortal", types.SEX_FEMALE)
	imm.PCData = &types.PCData{WizInvis: 99}
	imm.Trust = 99
	seer, _ := makeChar(t, "Seer", types.SEX_MALE)
	seer.Trust = 1
	seer.Act.Set(types.PLR_HOLYLIGHT)

	got := ActFormat("$n casts a spell", seer, imm, nil, nil, nil)
	want := "Immortal casts a spell\n\r"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// ---------- Act: routing ----------

// actRoutingSetup builds a room with ch, vch, and two bystanders, each with a
// net.Pipe descriptor. Returns the chars and their client ends.
type routingFixture struct {
	ch, vch, by1, by2 *types.CharData
	chClient          net.Conn
	vchClient         net.Conn
	by1Client         net.Conn
	by2Client         net.Conn
}

func setupRouting(t *testing.T) *routingFixture {
	t.Helper()
	ch, cc := makeChar(t, "Alice", types.SEX_FEMALE)
	vch, vc := makeChar(t, "Bob", types.SEX_MALE)
	by1, bc1 := makeChar(t, "Carol", types.SEX_FEMALE)
	by2, bc2 := makeChar(t, "Dave", types.SEX_MALE)
	putInRoom(ch, vch, by1, by2)
	return &routingFixture{
		ch: ch, vch: vch, by1: by1, by2: by2,
		chClient: cc, vchClient: vc, by1Client: bc1, by2Client: bc2,
	}
}

func TestAct_TO_CHAR_OnlyChReceives(t *testing.T) {
	f := setupRouting(t)
	Act(types.AT_PLAIN, "$n tests", f.ch, f.vch, nil, nil, types.TO_CHAR)

	if got := readOutput(t, f.ch, f.chClient); !strings.Contains(got, "Alice tests") {
		t.Errorf("TO_CHAR: ch should receive, got %q", got)
	}
	if got := readOutput(t, f.vch, f.vchClient); got != "" {
		t.Errorf("TO_CHAR: vch should NOT receive, got %q", got)
	}
	if got := readOutput(t, f.by1, f.by1Client); got != "" {
		t.Errorf("TO_CHAR: by1 should NOT receive, got %q", got)
	}
	if got := readOutput(t, f.by2, f.by2Client); got != "" {
		t.Errorf("TO_CHAR: by2 should NOT receive, got %q", got)
	}
}

func TestAct_TO_VICT_OnlyVchReceives(t *testing.T) {
	f := setupRouting(t)
	Act(types.AT_PLAIN, "$n tests", f.ch, f.vch, nil, nil, types.TO_VICT)

	if got := readOutput(t, f.ch, f.chClient); got != "" {
		t.Errorf("TO_VICT: ch should NOT receive, got %q", got)
	}
	if got := readOutput(t, f.vch, f.vchClient); !strings.Contains(got, "Alice tests") {
		t.Errorf("TO_VICT: vch should receive, got %q", got)
	}
	if got := readOutput(t, f.by1, f.by1Client); got != "" {
		t.Errorf("TO_VICT: by1 should NOT receive, got %q", got)
	}
	if got := readOutput(t, f.by2, f.by2Client); got != "" {
		t.Errorf("TO_VICT: by2 should NOT receive, got %q", got)
	}
}

func TestAct_TO_NOTVICT_OnlyBystandersReceive(t *testing.T) {
	f := setupRouting(t)
	Act(types.AT_PLAIN, "$n tests", f.ch, f.vch, nil, nil, types.TO_NOTVICT)

	if got := readOutput(t, f.ch, f.chClient); got != "" {
		t.Errorf("TO_NOTVICT: ch should NOT receive, got %q", got)
	}
	if got := readOutput(t, f.vch, f.vchClient); got != "" {
		t.Errorf("TO_NOTVICT: vch should NOT receive, got %q", got)
	}
	if got := readOutput(t, f.by1, f.by1Client); !strings.Contains(got, "Alice tests") {
		t.Errorf("TO_NOTVICT: by1 should receive, got %q", got)
	}
	if got := readOutput(t, f.by2, f.by2Client); !strings.Contains(got, "Alice tests") {
		t.Errorf("TO_NOTVICT: by2 should receive, got %q", got)
	}
}

func TestAct_TO_ROOM_AllExceptActorReceive(t *testing.T) {
	f := setupRouting(t)
	Act(types.AT_PLAIN, "$n tests", f.ch, f.vch, nil, nil, types.TO_ROOM)

	if got := readOutput(t, f.ch, f.chClient); got != "" {
		t.Errorf("TO_ROOM: ch should NOT receive, got %q", got)
	}
	if got := readOutput(t, f.vch, f.vchClient); !strings.Contains(got, "Alice tests") {
		t.Errorf("TO_ROOM: vch should receive, got %q", got)
	}
	if got := readOutput(t, f.by1, f.by1Client); !strings.Contains(got, "Alice tests") {
		t.Errorf("TO_ROOM: by1 should receive, got %q", got)
	}
	if got := readOutput(t, f.by2, f.by2Client); !strings.Contains(got, "Alice tests") {
		t.Errorf("TO_ROOM: by2 should receive, got %q", got)
	}
}

func TestAct_NilCh_NoPanic(t *testing.T) {
	// Just verify no crash and no side effects.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Act(nil ch) panicked: %v", r)
		}
	}()
	Act(types.AT_PLAIN, "$n tests", nil, nil, nil, nil, types.TO_ROOM)
}

func TestAct_TO_VICT_NilVch_NoPanic(t *testing.T) {
	ch, _ := makeChar(t, "Alice", types.SEX_FEMALE)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Act(TO_VICT, nil vch) panicked: %v", r)
		}
	}()
	Act(types.AT_PLAIN, "$n tests", ch, nil, nil, nil, types.TO_VICT)
}

func TestActFormat_BadSexFallsBackToNeutral(t *testing.T) {
	weird := &types.CharData{Name: "glitch", Sex: 42}
	got := ActFormat("$e stands", nil, weird, nil, nil, nil)
	want := "It stands\n\r"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestActFormat_BigE_NilVch_EmitsPlaceholder(t *testing.T) {
	alice, _ := makeChar(t, "Alice", types.SEX_FEMALE)
	got := ActFormat("$E laughs", nil, alice, nil, nil, nil)
	if !strings.Contains(got, "@@@") {
		t.Fatalf("expected placeholder for missing vch, got %q", got)
	}
}

func TestActFormat_BadToken_EmitsPlaceholder(t *testing.T) {
	alice, _ := makeChar(t, "Alice", types.SEX_FEMALE)
	got := ActFormat("$z zaps", nil, alice, nil, nil, nil)
	if !strings.Contains(got, "@@@") {
		t.Fatalf("expected placeholder for unknown token, got %q", got)
	}
}

func TestAct_EmptyFormat_Noop(t *testing.T) {
	f := setupRouting(t)
	Act(types.AT_PLAIN, "", f.ch, f.vch, nil, nil, types.TO_ROOM)
	if got := readOutput(t, f.vch, f.vchClient); got != "" {
		t.Errorf("empty format should not send, got %q", got)
	}
}
