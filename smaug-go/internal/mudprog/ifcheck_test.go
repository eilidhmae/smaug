package mudprog

import (
	"strconv"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

func TestDoIfCheck_EmptyCheck(t *testing.T) {
	if DoIfCheck("", nil, nil, nil, nil, nil) {
		t.Error("empty check should return false")
	}
}

func TestDoIfCheck_NoParen(t *testing.T) {
	if DoIfCheck("ispc $n", nil, nil, nil, nil, nil) {
		t.Error("missing parens should return false")
	}
}

func TestDoIfCheck_NoCloseParen(t *testing.T) {
	if DoIfCheck("ispc($n", nil, nil, nil, nil, nil) {
		t.Error("missing close paren should return false")
	}
}

func TestDoIfCheck_UnknownCheck(t *testing.T) {
	ch := &types.CharData{Name: "Player"}
	if DoIfCheck("nonexistent($n) == 1", nil, ch, nil, nil, nil) {
		t.Error("unknown check should return false")
	}
}

func TestDoIfCheck_Rand(t *testing.T) {
	trueCount := 0
	for i := 0; i < 200; i++ {
		if DoIfCheck("rand($i) == 50", nil, nil, nil, nil, nil) {
			trueCount++
		}
	}
	if trueCount == 0 || trueCount == 200 {
		t.Errorf("rand check seems broken: %d/200 true", trueCount)
	}
}

func TestDoIfCheck_IsPC(t *testing.T) {
	pc := &types.CharData{Name: "Player"}
	npc := &types.CharData{Name: "Mob"}
	npc.Act.Set(types.ACT_IS_NPC)

	if !DoIfCheck("ispc($n)", nil, pc, nil, nil, nil) {
		t.Error("ispc should return true for PC")
	}
	if DoIfCheck("ispc($n)", nil, npc, nil, nil, nil) {
		t.Error("ispc should return false for NPC")
	}
}

func TestDoIfCheck_IsNPC(t *testing.T) {
	pc := &types.CharData{Name: "Player"}
	npc := &types.CharData{Name: "Mob"}
	npc.Act.Set(types.ACT_IS_NPC)

	if DoIfCheck("isnpc($n)", nil, pc, nil, nil, nil) {
		t.Error("isnpc should return false for PC")
	}
	if !DoIfCheck("isnpc($n)", nil, npc, nil, nil, nil) {
		t.Error("isnpc should return true for NPC")
	}
}

func TestDoIfCheck_IsEvil(t *testing.T) {
	evil := &types.CharData{Name: "Evil", Alignment: -500}
	good := &types.CharData{Name: "Good", Alignment: 500}
	neutral := &types.CharData{Name: "Neutral", Alignment: 0}

	if !DoIfCheck("isevil($n)", nil, evil, nil, nil, nil) {
		t.Error("alignment -500 should be evil")
	}
	if DoIfCheck("isevil($n)", nil, good, nil, nil, nil) {
		t.Error("alignment 500 should not be evil")
	}
	if DoIfCheck("isevil($n)", nil, neutral, nil, nil, nil) {
		t.Error("alignment 0 should not be evil")
	}
}

func TestDoIfCheck_IsGood(t *testing.T) {
	evil := &types.CharData{Name: "Evil", Alignment: -500}
	good := &types.CharData{Name: "Good", Alignment: 500}
	neutral := &types.CharData{Name: "Neutral", Alignment: 0}

	if !DoIfCheck("isgood($n)", nil, good, nil, nil, nil) {
		t.Error("alignment 500 should be good")
	}
	if DoIfCheck("isgood($n)", nil, evil, nil, nil, nil) {
		t.Error("alignment -500 should not be good")
	}
	if DoIfCheck("isgood($n)", nil, neutral, nil, nil, nil) {
		t.Error("alignment 0 should not be good")
	}
}

func TestDoIfCheck_IsNeutral(t *testing.T) {
	evil := &types.CharData{Name: "Evil", Alignment: -500}
	good := &types.CharData{Name: "Good", Alignment: 500}
	neutral := &types.CharData{Name: "Neutral", Alignment: 0}
	borderLow := &types.CharData{Name: "BorderLow", Alignment: -350}
	borderHigh := &types.CharData{Name: "BorderHigh", Alignment: 350}

	if !DoIfCheck("isneutral($n)", nil, neutral, nil, nil, nil) {
		t.Error("alignment 0 should be neutral")
	}
	if !DoIfCheck("isneutral($n)", nil, borderLow, nil, nil, nil) {
		t.Error("alignment -350 should be neutral")
	}
	if !DoIfCheck("isneutral($n)", nil, borderHigh, nil, nil, nil) {
		t.Error("alignment 350 should be neutral")
	}
	if DoIfCheck("isneutral($n)", nil, evil, nil, nil, nil) {
		t.Error("alignment -500 should not be neutral")
	}
	if DoIfCheck("isneutral($n)", nil, good, nil, nil, nil) {
		t.Error("alignment 500 should not be neutral")
	}
}

func TestDoIfCheck_IsImmort(t *testing.T) {
	mortal := &types.CharData{Name: "Mortal", Level: 10}
	immortal := &types.CharData{Name: "Immortal", Level: types.LEVEL_IMMORTAL}

	if DoIfCheck("isimmort($n)", nil, mortal, nil, nil, nil) {
		t.Error("level 10 should not be immortal")
	}
	if !DoIfCheck("isimmort($n)", nil, immortal, nil, nil, nil) {
		t.Error("immortal level should be immortal")
	}
}

func TestDoIfCheck_IsFight(t *testing.T) {
	fighting := &types.CharData{
		Name:     "Fighter",
		Fighting: &types.FightData{Who: &types.CharData{Name: "Target"}},
	}
	idle := &types.CharData{Name: "Idle"}

	if !DoIfCheck("isfight($n)", nil, fighting, nil, nil, nil) {
		t.Error("fighting char should be isfight")
	}
	if DoIfCheck("isfight($n)", nil, idle, nil, nil, nil) {
		t.Error("idle char should not be isfight")
	}
}

func TestDoIfCheck_IsCharmed(t *testing.T) {
	charmed := &types.CharData{Name: "Charmed"}
	charmed.AffectedBy.Set(types.AFF_CHARM)
	normal := &types.CharData{Name: "Normal"}

	if !DoIfCheck("ischarmed($n)", nil, charmed, nil, nil, nil) {
		t.Error("charmed char should be ischarmed")
	}
	if DoIfCheck("ischarmed($n)", nil, normal, nil, nil, nil) {
		t.Error("normal char should not be ischarmed")
	}
}

func TestDoIfCheck_IsFlying(t *testing.T) {
	flying := &types.CharData{Name: "Flyer"}
	flying.AffectedBy.Set(types.AFF_FLYING)
	grounded := &types.CharData{Name: "Grounded"}

	if !DoIfCheck("isflying($n)", nil, flying, nil, nil, nil) {
		t.Error("flying char should be isflying")
	}
	if DoIfCheck("isflying($n)", nil, grounded, nil, nil, nil) {
		t.Error("grounded char should not be isflying")
	}
}

func TestDoIfCheck_IsInvis(t *testing.T) {
	invis := &types.CharData{Name: "Invis"}
	invis.AffectedBy.Set(types.AFF_INVISIBLE)
	visible := &types.CharData{Name: "Visible"}

	if !DoIfCheck("isinvis($n)", nil, invis, nil, nil, nil) {
		t.Error("invisible char should be isinvis")
	}
	if DoIfCheck("isinvis($n)", nil, visible, nil, nil, nil) {
		t.Error("visible char should not be isinvis")
	}
}

func TestDoIfCheck_IsAffected(t *testing.T) {
	ch := &types.CharData{Name: "Test"}
	ch.AffectedBy.Set(types.AFF_CHARM)

	if !DoIfCheck("isaffected($n) == "+itoaTest(types.AFF_CHARM), nil, ch, nil, nil, nil) {
		t.Error("should detect AFF_CHARM")
	}
	if DoIfCheck("isaffected($n) == "+itoaTest(types.AFF_FLYING), nil, ch, nil, nil, nil) {
		t.Error("should not detect AFF_FLYING when not set")
	}
}

func TestDoIfCheck_Level(t *testing.T) {
	ch := &types.CharData{Name: "Test", Level: 20}

	tests := []struct {
		check    string
		expected bool
	}{
		{"level($n) > 10", true},
		{"level($n) > 30", false},
		{"level($n) == 20", true},
		{"level($n) != 20", false},
		{"level($n) != 15", true},
		{"level($n) < 25", true},
		{"level($n) < 10", false},
		{"level($n) >= 20", true},
		{"level($n) >= 21", false},
		{"level($n) <= 20", true},
		{"level($n) <= 19", false},
	}
	for _, tc := range tests {
		result := DoIfCheck(tc.check, nil, ch, nil, nil, nil)
		if result != tc.expected {
			t.Errorf("DoIfCheck(%q) = %v, want %v", tc.check, result, tc.expected)
		}
	}
}

func TestDoIfCheck_Hp(t *testing.T) {
	ch := &types.CharData{Name: "Test", Hit: 75}

	if !DoIfCheck("hp($n) > 50", nil, ch, nil, nil, nil) {
		t.Error("hp 75 > 50 should be true")
	}
	if DoIfCheck("hp($n) > 100", nil, ch, nil, nil, nil) {
		t.Error("hp 75 > 100 should be false")
	}
	if !DoIfCheck("hp($n) == 75", nil, ch, nil, nil, nil) {
		t.Error("hp 75 == 75 should be true")
	}
}

func TestDoIfCheck_HpPcnt(t *testing.T) {
	ch := &types.CharData{Name: "Test", Hit: 50, MaxHit: 100}

	if !DoIfCheck("hppcnt($n) == 50", nil, ch, nil, nil, nil) {
		t.Error("50/100 hp should be 50%")
	}
	if !DoIfCheck("hppcnt($n) < 75", nil, ch, nil, nil, nil) {
		t.Error("50% < 75% should be true")
	}

	// Test with zero MaxHit
	zeroMax := &types.CharData{Name: "Zero", Hit: 10, MaxHit: 0}
	if DoIfCheck("hppcnt($n) > 0", nil, zeroMax, nil, nil, nil) {
		t.Error("zero max hit should return false")
	}
}

func TestDoIfCheck_Mana(t *testing.T) {
	ch := &types.CharData{Name: "Test", Mana: 100}

	if !DoIfCheck("mana($n) > 50", nil, ch, nil, nil, nil) {
		t.Error("mana 100 > 50 should be true")
	}
	if DoIfCheck("mana($n) > 200", nil, ch, nil, nil, nil) {
		t.Error("mana 100 > 200 should be false")
	}
}

func TestDoIfCheck_Gold(t *testing.T) {
	ch := &types.CharData{Name: "Test", Gold: 1000}

	if !DoIfCheck("gold($n) > 500", nil, ch, nil, nil, nil) {
		t.Error("gold 1000 > 500 should be true")
	}
	if !DoIfCheck("gold($n) == 1000", nil, ch, nil, nil, nil) {
		t.Error("gold 1000 == 1000 should be true")
	}
}

func TestDoIfCheck_Sex(t *testing.T) {
	male := &types.CharData{Name: "Male", Sex: types.SEX_MALE}
	female := &types.CharData{Name: "Female", Sex: types.SEX_FEMALE}

	if !DoIfCheck("sex($n) == 1", nil, male, nil, nil, nil) {
		t.Errorf("male sex should be %d", types.SEX_MALE)
	}
	if !DoIfCheck("sex($n) == 2", nil, female, nil, nil, nil) {
		t.Errorf("female sex should be %d", types.SEX_FEMALE)
	}
}

func TestDoIfCheck_Position(t *testing.T) {
	ch := &types.CharData{Name: "Test", Position: types.POS_STANDING}

	if !DoIfCheck("position($n) == "+itoaTest(types.POS_STANDING), nil, ch, nil, nil, nil) {
		t.Error("position should match POS_STANDING")
	}
	if DoIfCheck("position($n) == "+itoaTest(types.POS_SLEEPING), nil, ch, nil, nil, nil) {
		t.Error("position should not match POS_SLEEPING")
	}
}

func TestDoIfCheck_Class(t *testing.T) {
	ch := &types.CharData{Name: "Test", Class: 3}

	if !DoIfCheck("class($n) == 3", nil, ch, nil, nil, nil) {
		t.Error("class 3 == 3 should be true")
	}
	if DoIfCheck("class($n) == 5", nil, ch, nil, nil, nil) {
		t.Error("class 3 == 5 should be false")
	}
}

func TestDoIfCheck_Race(t *testing.T) {
	ch := &types.CharData{Name: "Test", Race: 2}

	if !DoIfCheck("race($n) == 2", nil, ch, nil, nil, nil) {
		t.Error("race 2 == 2 should be true")
	}
	if DoIfCheck("race($n) == 0", nil, ch, nil, nil, nil) {
		t.Error("race 2 == 0 should be false")
	}
}

func TestDoIfCheck_Alignment(t *testing.T) {
	ch := &types.CharData{Name: "Test", Alignment: -750}

	if !DoIfCheck("alignment($n) < 0", nil, ch, nil, nil, nil) {
		t.Error("alignment -750 < 0 should be true")
	}
	if !DoIfCheck("alignment($n) == -750", nil, ch, nil, nil, nil) {
		t.Error("alignment -750 == -750 should be true")
	}
}

func TestDoIfCheck_Str(t *testing.T) {
	ch := &types.CharData{Name: "Test", PermStr: 18}

	if !DoIfCheck("str($n) > 15", nil, ch, nil, nil, nil) {
		t.Error("str 18 > 15 should be true")
	}
	if !DoIfCheck("str($n) == 18", nil, ch, nil, nil, nil) {
		t.Error("str 18 == 18 should be true")
	}
}

func TestDoIfCheck_Int(t *testing.T) {
	ch := &types.CharData{Name: "Test", PermInt: 16}

	if !DoIfCheck("int($n) == 16", nil, ch, nil, nil, nil) {
		t.Error("int 16 == 16 should be true")
	}
}

func TestDoIfCheck_Wis(t *testing.T) {
	ch := &types.CharData{Name: "Test", PermWis: 14}

	if !DoIfCheck("wis($n) == 14", nil, ch, nil, nil, nil) {
		t.Error("wis 14 == 14 should be true")
	}
}

func TestDoIfCheck_Dex(t *testing.T) {
	ch := &types.CharData{Name: "Test", PermDex: 20}

	if !DoIfCheck("dex($n) > 18", nil, ch, nil, nil, nil) {
		t.Error("dex 20 > 18 should be true")
	}
}

func TestDoIfCheck_Con(t *testing.T) {
	ch := &types.CharData{Name: "Test", PermCon: 12}

	if !DoIfCheck("con($n) == 12", nil, ch, nil, nil, nil) {
		t.Error("con 12 == 12 should be true")
	}
}

func TestDoIfCheck_IsInRoom(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	mob := &types.CharData{Name: "Guard", InRoom: room}

	if !DoIfCheck("isinroom($i) == 3001", mob, nil, nil, nil, nil) {
		t.Error("mob in room 3001 should match isinroom 3001")
	}
	if DoIfCheck("isinroom($i) == 9999", mob, nil, nil, nil, nil) {
		t.Error("mob in room 3001 should not match isinroom 9999")
	}

	// Nil room
	noRoom := &types.CharData{Name: "Lost"}
	if DoIfCheck("isinroom($i) == 3001", noRoom, nil, nil, nil, nil) {
		t.Error("mob with no room should return false")
	}
}

func TestDoIfCheck_Name(t *testing.T) {
	ch := &types.CharData{Name: "Gandalf"}

	tests := []struct {
		check    string
		expected bool
	}{
		{"name($n) == Gandalf", true},
		{"name($n) == gandalf", true}, // case insensitive
		{"name($n) == Frodo", false},
		{"name($n) != Frodo", true},
		{"name($n) != Gandalf", false},
		{"name($n) / anda", true},  // contains
		{"name($n) / xyz", false},  // not contains
		{"name($n) !/ anda", false},
		{"name($n) !/ xyz", true},
	}
	for _, tc := range tests {
		result := DoIfCheck(tc.check, nil, ch, nil, nil, nil)
		if result != tc.expected {
			t.Errorf("DoIfCheck(%q) = %v, want %v", tc.check, result, tc.expected)
		}
	}
}

func TestDoIfCheck_NilChar(t *testing.T) {
	// All checks that need a char should return false with nil
	checks := []string{
		"ispc($n)", "isnpc($n)", "isevil($n)", "isgood($n)", "isneutral($n)",
		"isimmort($n)", "isfight($n)", "ischarmed($n)", "isflying($n)", "isinvis($n)",
		"level($n) > 0", "hp($n) > 0", "hppcnt($n) > 0", "mana($n) > 0",
		"gold($n) > 0", "sex($n) == 0", "position($n) == 0",
		"class($n) == 0", "race($n) == 0", "alignment($n) == 0",
		"str($n) > 0", "int($n) > 0", "wis($n) > 0", "dex($n) > 0", "con($n) > 0",
		"name($n) == test",
	}
	for _, check := range checks {
		if DoIfCheck(check, nil, nil, nil, nil, nil) {
			t.Errorf("DoIfCheck(%q) with nil char should return false", check)
		}
	}
}

func TestDoIfCheck_VarResolution(t *testing.T) {
	mob := &types.CharData{Name: "Guard", Level: 5}
	mob.Act.Set(types.ACT_IS_NPC)
	actor := &types.CharData{Name: "Player", Level: 20}
	victim := &types.CharData{Name: "Victim", Level: 10}

	// $i resolves to mob
	if !DoIfCheck("level($i) == 5", mob, actor, nil, victim, nil) {
		t.Error("$i should resolve to mob with level 5")
	}
	// $n resolves to actor
	if !DoIfCheck("level($n) == 20", mob, actor, nil, victim, nil) {
		t.Error("$n should resolve to actor with level 20")
	}
	// $t resolves to victim
	if !DoIfCheck("level($t) == 10", mob, actor, nil, victim, nil) {
		t.Error("$t should resolve to victim with level 10")
	}
	// $r resolves to first person in room
	room := &types.RoomIndexData{Vnum: 100}
	mob.InRoom = room
	roomChar := &types.CharData{Name: "RoomPerson", Level: 30}
	room.People = []*types.CharData{roomChar}
	if !DoIfCheck("level($r) == 30", mob, actor, nil, victim, nil) {
		t.Error("$r should resolve to first person in room with level 30")
	}
	// Default var resolves to actor
	if !DoIfCheck("level($x) == 20", mob, actor, nil, victim, nil) {
		t.Error("unknown var should default to actor with level 20")
	}
}

func TestDoIfCheck_InvalidOperator(t *testing.T) {
	ch := &types.CharData{Name: "Test", Level: 20}

	// Invalid operator defaults to ==
	if !DoIfCheck("level($n) ?? 20", nil, ch, nil, nil, nil) {
		t.Error("invalid operator should default to == comparison")
	}
	if DoIfCheck("level($n) ?? 15", nil, ch, nil, nil, nil) {
		t.Error("invalid operator with non-matching value should be false")
	}
}

func TestDoIfCheck_NoOperatorValue(t *testing.T) {
	// When no operator/value is given, val defaults to 0 and op to ==
	ch := &types.CharData{Name: "Test", Level: 0}

	if !DoIfCheck("level($n)", nil, ch, nil, nil, nil) {
		t.Error("level 0 with no op/val should match (defaults to == 0)")
	}
}

func TestCompareInt(t *testing.T) {
	tests := []struct {
		lhs      int
		op       string
		rhs      int
		expected bool
	}{
		{10, "==", 10, true},
		{10, "==", 5, false},
		{10, "=", 10, true},
		{10, "!=", 5, true},
		{10, "!=", 10, false},
		{10, ">", 5, true},
		{10, ">", 10, false},
		{10, "<", 15, true},
		{10, "<", 10, false},
		{10, ">=", 10, true},
		{10, ">=", 11, false},
		{10, "<=", 10, true},
		{10, "<=", 9, false},
		{10, "??", 10, true},  // default to ==
		{10, "??", 5, false},
	}
	for _, tc := range tests {
		result := compareInt(tc.lhs, tc.op, tc.rhs)
		if result != tc.expected {
			t.Errorf("compareInt(%d, %q, %d) = %v, want %v", tc.lhs, tc.op, tc.rhs, result, tc.expected)
		}
	}
}

func TestMatchStr(t *testing.T) {
	tests := []struct {
		lhs      string
		op       string
		rhs      string
		expected bool
	}{
		{"hello", "==", "hello", true},
		{"Hello", "==", "hello", true}, // case insensitive
		{"hello", "==", "world", false},
		{"hello", "=", "hello", true},
		{"hello", "!=", "world", true},
		{"hello", "!=", "hello", false},
		{"hello world", "/", "world", true},
		{"hello world", "/", "xyz", false},
		{"hello world", "!/", "xyz", true},
		{"hello world", "!/", "world", false},
		{"hello", "??", "hello", true}, // default to ==
		{"hello", "??", "world", false},
	}
	for _, tc := range tests {
		result := matchStr(tc.lhs, tc.op, tc.rhs)
		if result != tc.expected {
			t.Errorf("matchStr(%q, %q, %q) = %v, want %v", tc.lhs, tc.op, tc.rhs, result, tc.expected)
		}
	}
}

func TestResolveChar(t *testing.T) {
	mob := &types.CharData{Name: "Mob"}
	actor := &types.CharData{Name: "Actor"}
	victim := &types.CharData{Name: "Victim"}

	tests := []struct {
		varStr   string
		expected *types.CharData
	}{
		{"$i", mob},
		{"$n", actor},
		{"$t", victim},
		{"$x", actor},    // unknown defaults to actor
		{"  $i  ", mob},   // trimmed
	}
	for _, tc := range tests {
		result := resolveChar(tc.varStr, mob, actor, victim)
		if result != tc.expected {
			t.Errorf("resolveChar(%q) = %v, want %v", tc.varStr, result, tc.expected)
		}
	}
}

func TestResolveChar_RoomPeople(t *testing.T) {
	roomPerson := &types.CharData{Name: "RoomPerson"}
	room := &types.RoomIndexData{Vnum: 100, People: []*types.CharData{roomPerson}}
	mob := &types.CharData{Name: "Mob", InRoom: room}

	result := resolveChar("$r", mob, nil, nil)
	if result != roomPerson {
		t.Errorf("$r should resolve to first person in room")
	}

	// Empty room
	emptyRoom := &types.RoomIndexData{Vnum: 200}
	mobEmpty := &types.CharData{Name: "Mob2", InRoom: emptyRoom}
	result = resolveChar("$r", mobEmpty, nil, nil)
	if result != nil {
		t.Errorf("$r with empty room should return nil")
	}

	// Nil mob
	result = resolveChar("$r", nil, nil, nil)
	if result != nil {
		t.Errorf("$r with nil mob should return nil")
	}
}

// itoaTest is a local helper to convert int to string for test expressions.
func itoaTest(n int) string {
	return strconv.Itoa(n)
}
