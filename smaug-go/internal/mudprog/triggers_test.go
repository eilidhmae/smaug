package mudprog

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/types"
)

func TestTriggerMatches_Rand(t *testing.T) {
	trueCount := 0
	for i := 0; i < 200; i++ {
		if triggerMatches(types.MPROG_RAND, "50", "") {
			trueCount++
		}
	}
	// Should be roughly 50% but verify it's not always true or false
	if trueCount == 0 || trueCount == 200 {
		t.Errorf("rand trigger seems broken: %d/200 true", trueCount)
	}
}

func TestTriggerMatches_Rand_Clamped(t *testing.T) {
	// 0% should never fire
	for i := 0; i < 50; i++ {
		if triggerMatches(types.MPROG_RAND, "0", "") {
			t.Fatal("rand 0 should never fire")
		}
	}
	// 100% should always fire
	for i := 0; i < 50; i++ {
		if !triggerMatches(types.MPROG_RAND, "100", "") {
			t.Fatal("rand 100 should always fire")
		}
	}
}

func TestTriggerMatches_Speech(t *testing.T) {
	tests := []struct {
		argList  string
		actual   string
		expected bool
	}{
		{"hello", "hello world", true},
		{"hello", "goodbye", false},
		{"hello world", "I said hello to you", true},
		{"hello world", "I said goodbye", false},
		{"", "anything", true},           // empty arglist matches all
		{"Hello", "HELLO world", true},   // case insensitive
		{"p hello", "hello world", true}, // "p" prefix is skipped
	}
	for _, tc := range tests {
		result := triggerMatches(types.MPROG_SPEECH, tc.argList, tc.actual)
		if result != tc.expected {
			t.Errorf("triggerMatches(SPEECH, %q, %q) = %v, want %v",
				tc.argList, tc.actual, result, tc.expected)
		}
	}
}

func TestTriggerMatches_SpeechIW(t *testing.T) {
	// SPEECHIW should use same logic as SPEECH
	if !triggerMatches(types.MPROG_SPEECHIW, "hello", "hello world") {
		t.Error("speechiw should match keyword in speech")
	}
	if triggerMatches(types.MPROG_SPEECHIW, "hello", "goodbye") {
		t.Error("speechiw should not match when keyword absent")
	}
}

func TestTriggerMatches_Act(t *testing.T) {
	tests := []struct {
		argList  string
		actual   string
		expected bool
	}{
		{"laughs", "Player laughs at you.", true},
		{"laughs", "Player cries.", false},
		{"", "anything", true},
		{"LAUGHS", "player laughs", true}, // case insensitive
	}
	for _, tc := range tests {
		result := triggerMatches(types.MPROG_ACT, tc.argList, tc.actual)
		if result != tc.expected {
			t.Errorf("triggerMatches(ACT, %q, %q) = %v, want %v",
				tc.argList, tc.actual, result, tc.expected)
		}
	}
}

func TestTriggerMatches_Bribe(t *testing.T) {
	tests := []struct {
		argList  string
		actual   string
		expected bool
	}{
		{"100", "200", true},
		{"100", "100", true},
		{"100", "50", false},
		{"0", "0", true},
	}
	for _, tc := range tests {
		result := triggerMatches(types.MPROG_BRIBE, tc.argList, tc.actual)
		if result != tc.expected {
			t.Errorf("triggerMatches(BRIBE, %q, %q) = %v, want %v",
				tc.argList, tc.actual, result, tc.expected)
		}
	}
}

func TestTriggerMatches_Give(t *testing.T) {
	tests := []struct {
		argList  string
		actual   string
		expected bool
	}{
		{"sword", "magic sword", true},   // IsName checks word-by-word
		{"shield", "magic sword", false},
		{"", "anything", true},
	}
	for _, tc := range tests {
		result := triggerMatches(types.MPROG_GIVE, tc.argList, tc.actual)
		if result != tc.expected {
			t.Errorf("triggerMatches(GIVE, %q, %q) = %v, want %v",
				tc.argList, tc.actual, result, tc.expected)
		}
	}
}

func TestTriggerMatches_HitPrcnt(t *testing.T) {
	// HITPRCNT always returns true (caller checks HP percent)
	if !triggerMatches(types.MPROG_HITPRCNT, "50", "") {
		t.Error("hitprcnt should always return true")
	}
}

func TestTriggerMatches_Default(t *testing.T) {
	// Default triggers (GREET, ENTRY, FIGHT, DEATH, etc.) always fire
	defaultTypes := []int{
		types.MPROG_GREET,
		types.MPROG_ALL_GREET,
		types.MPROG_ENTRY,
		types.MPROG_FIGHT,
		types.MPROG_DEATH,
		types.MPROG_WEAR,
		types.MPROG_REMOVE,
		types.MPROG_LOOK,
	}
	for _, tt := range defaultTypes {
		if !triggerMatches(tt, "anything", "anything") {
			t.Errorf("trigger type %d should always match", tt)
		}
	}
}

func TestAtoi(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"100", 100},
		{"+50", 50},
		{"0", 0},
		{"abc", 0},
		{"  42  ", 42},
		{"12abc", 12},
		{"", 0},
	}
	for _, tc := range tests {
		result := atoi(tc.input)
		if result != tc.expected {
			t.Errorf("atoi(%q) = %d, want %d", tc.input, result, tc.expected)
		}
	}
}

func TestMobTrigger_NilMob(t *testing.T) {
	// Should not panic
	MobTrigger(types.MPROG_GREET, nil, nil, nil, nil, nil, "")
}

func TestMobTrigger_NilIndexData(t *testing.T) {
	mob := &types.CharData{Name: "Test"}
	mob.Act.Set(types.ACT_IS_NPC)
	// nil IndexData should bail
	MobTrigger(types.MPROG_GREET, mob, nil, nil, nil, nil, "")
}

func TestMobTrigger_NoProgs(t *testing.T) {
	mob := makeNPC("guard")
	mob.IndexData = &types.MobIndexData{}
	mob.IndexData.ProgTypes.Set(0) // has prog types set but empty list
	// Should bail on the ProgTypes.IsSet(0) && len == 0 check
	MobTrigger(types.MPROG_GREET, mob, nil, nil, nil, nil, "")
}

func TestMobTrigger_MatchingProg(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001, Name: "Test Room"}

	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	mob.IndexData = &types.MobIndexData{
		MudProgs: []*types.MProgData{
			{
				Type:    types.MPROG_GREET,
				ArgList: "",
				ComList: "mpecho Welcome, traveler!",
			},
		},
	}

	actor, client := makeDescChar("Player")
	defer client.Close()
	actor.InRoom = room
	room.People = append(room.People, actor)

	progNest = 0
	MobTrigger(types.MPROG_GREET, mob, actor, nil, nil, nil, "")

	out := readTrigOutput(actor, client)
	if !strings.Contains(out, "Welcome, traveler!") {
		t.Errorf("expected greeting output, got %q", out)
	}
}

func TestMobTrigger_NonMatchingType(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001, Name: "Test Room"}

	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	mob.IndexData = &types.MobIndexData{
		MudProgs: []*types.MProgData{
			{
				Type:    types.MPROG_DEATH, // only triggers on DEATH
				ArgList: "",
				ComList: "mpecho You should not see this",
			},
		},
	}

	actor, client := makeDescChar("Player")
	defer client.Close()
	actor.InRoom = room
	room.People = append(room.People, actor)

	progNest = 0
	MobTrigger(types.MPROG_GREET, mob, actor, nil, nil, nil, "") // trigger GREET, not DEATH

	out := readTrigOutput(actor, client)
	if strings.Contains(out, "You should not see this") {
		t.Error("non-matching trigger type should not fire")
	}
}

func TestMobTrigger_SpeechMatch(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001, Name: "Test Room"}

	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	mob.IndexData = &types.MobIndexData{
		MudProgs: []*types.MProgData{
			{
				Type:    types.MPROG_SPEECH,
				ArgList: "password",
				ComList: "mpecho The guard nods.",
			},
		},
	}

	actor, client := makeDescChar("Player")
	defer client.Close()
	actor.InRoom = room
	room.People = append(room.People, actor)

	progNest = 0
	MobTrigger(types.MPROG_SPEECH, mob, actor, nil, nil, nil, "I know the password")

	out := readTrigOutput(actor, client)
	if !strings.Contains(out, "The guard nods.") {
		t.Errorf("speech trigger should fire for matching keyword, got %q", out)
	}
}

func TestMobTrigger_SpeechNoMatch(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001, Name: "Test Room"}

	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = append(room.People, mob)

	mob.IndexData = &types.MobIndexData{
		MudProgs: []*types.MProgData{
			{
				Type:    types.MPROG_SPEECH,
				ArgList: "password",
				ComList: "mpecho Should not see this",
			},
		},
	}

	actor, client := makeDescChar("Player")
	defer client.Close()
	actor.InRoom = room
	room.People = append(room.People, actor)

	progNest = 0
	MobTrigger(types.MPROG_SPEECH, mob, actor, nil, nil, nil, "hello world")

	out := readTrigOutput(actor, client)
	if strings.Contains(out, "Should not see this") {
		t.Error("speech trigger should not fire for non-matching keyword")
	}
}

// readTrigOutput is a helper that reads output from a descriptor.
func readTrigOutput(ch *types.CharData, client net.Conn) string {
	if !ch.Desc.HasOutput() {
		return ""
	}
	result := make(chan string, 1)
	go func() {
		buf := make([]byte, 8192)
		client.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, _ := client.Read(buf)
		result <- string(buf[:n])
	}()
	_ = ch.Desc.FlushOutput()
	return <-result
}

func TestTrigGreet_NilChar(t *testing.T) {
	// Should not panic
	TrigGreet(nil)
}

func TestTrigGreet_NilRoom(t *testing.T) {
	ch := &types.CharData{Name: "Player"}
	TrigGreet(ch)
}

func TestTrigGreet_SkipsSelf(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	ch := &types.CharData{Name: "Player", Position: types.POS_STANDING, InRoom: room}
	room.People = []*types.CharData{ch}

	// Should not fire trigger on self
	TrigGreet(ch)
}

func TestTrigGreet_SkipsPC(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	ch := &types.CharData{Name: "Player", Position: types.POS_STANDING, InRoom: room}
	other := &types.CharData{Name: "OtherPlayer", Position: types.POS_STANDING, InRoom: room}
	room.People = []*types.CharData{ch, other}

	// Should not fire on non-NPCs
	TrigGreet(ch)
}

func TestTrigGreet_SkipsFightingMob(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	ch := &types.CharData{Name: "Player", Position: types.POS_STANDING, InRoom: room}
	mob := makeNPC("guard")
	mob.InRoom = room
	mob.Fighting = &types.FightData{Who: ch}
	mob.IndexData = &types.MobIndexData{
		MudProgs: []*types.MProgData{
			{Type: types.MPROG_GREET, ComList: "mpecho hello"},
		},
	}
	room.People = []*types.CharData{ch, mob}

	// Should not fire on fighting mob
	TrigGreet(ch)
}

func TestTrigGreet_SkipsNonStanding(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	ch := &types.CharData{Name: "Player", Position: types.POS_STANDING, InRoom: room}
	mob := makeNPC("guard")
	mob.InRoom = room
	mob.Position = types.POS_SLEEPING
	mob.IndexData = &types.MobIndexData{
		MudProgs: []*types.MProgData{
			{Type: types.MPROG_GREET, ComList: "mpecho hello"},
		},
	}
	room.People = []*types.CharData{ch, mob}

	// Should not fire on sleeping mob
	TrigGreet(ch)
}

func TestTrigEntry_NilMob(t *testing.T) {
	TrigEntry(nil)
}

func TestTrigEntry_NotNPC(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	ch := &types.CharData{Name: "Player", InRoom: room}
	// Not an NPC - should not fire
	TrigEntry(ch)
}

func TestTrigEntry_NilRoom(t *testing.T) {
	mob := makeNPC("guard")
	mob.InRoom = nil
	TrigEntry(mob)
}

func TestTrigSpeech_NilChar(t *testing.T) {
	TrigSpeech(nil, "hello")
}

func TestTrigSpeech_NilRoom(t *testing.T) {
	ch := &types.CharData{Name: "Player"}
	TrigSpeech(ch, "hello")
}

func TestTrigSpeech_SkipsSelf(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	mob := makeNPC("guard")
	mob.InRoom = room
	room.People = []*types.CharData{mob}

	// Speaking mob is the only one in the room - should not trigger on self
	TrigSpeech(mob, "hello")
}

func TestTrigFight_NilMob(t *testing.T) {
	TrigFight(nil)
}

func TestTrigFight_NotNPC(t *testing.T) {
	ch := &types.CharData{Name: "Player", Fighting: &types.FightData{Who: &types.CharData{}}}
	TrigFight(ch)
}

func TestTrigFight_NotFighting(t *testing.T) {
	mob := makeNPC("guard")
	TrigFight(mob)
}

func TestTrigDeath_NilMob(t *testing.T) {
	TrigDeath(nil, nil)
}

func TestTrigDeath_NotNPC(t *testing.T) {
	ch := &types.CharData{Name: "Player"}
	TrigDeath(ch, nil)
}

func TestTrigRand_NilMob(t *testing.T) {
	TrigRand(nil)
}

func TestTrigRand_NotNPC(t *testing.T) {
	ch := &types.CharData{Name: "Player", InRoom: &types.RoomIndexData{}}
	TrigRand(ch)
}

func TestTrigRand_NilRoom(t *testing.T) {
	mob := makeNPC("guard")
	mob.InRoom = nil
	TrigRand(mob)
}

func TestTrigGive_NilMob(t *testing.T) {
	TrigGive(nil, nil, nil)
}

func TestTrigGive_NotNPC(t *testing.T) {
	ch := &types.CharData{Name: "Player"}
	TrigGive(ch, nil, nil)
}
