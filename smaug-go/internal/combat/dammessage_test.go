package combat

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// makeDamChar returns a named character with a net.Pipe descriptor so that
// DamMessage output can be captured and asserted on. The server half of the
// pipe is held by the descriptor; the client half is returned for reading.
func makeDamChar(t *testing.T, name string, maxHit int) (*types.CharData, net.Conn) {
	t.Helper()
	server, client := net.Pipe()
	d := &types.DescriptorData{Conn: server, InputQueue: make(chan string, 4)}
	ch := &types.CharData{
		Name:     name,
		Sex:      types.SEX_MALE,
		Desc:     d,
		Position: types.POS_STANDING,
		Hit:      maxHit,
		MaxHit:   maxHit,
	}
	d.Character = ch
	return ch, client
}

// drain flushes ch's descriptor and reads whatever is in the pipe.
func drain(t *testing.T, ch *types.CharData, client net.Conn) string {
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

// putInDamRoom links chars into a shared room.
func putInDamRoom(chs ...*types.CharData) *types.RoomIndexData {
	room := &types.RoomIndexData{Vnum: 1}
	for _, c := range chs {
		c.InRoom = room
		room.People = append(room.People, c)
	}
	return room
}

// resetCombatWorldRef lets tests that mutate WorldRef restore it.
func resetCombatWorldRef(t *testing.T, prev *world.World) {
	t.Helper()
	t.Cleanup(func() { WorldRef = prev })
}

func TestDamMessageDIndex(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	WorldRef = world.New("/tmp/test")

	// MaxHit=100, Hit=100, so dampc = (dam*1000)/100 + (50 - 50) = dam*10.
	// With w_index=0 (DAM_HIT -> generic table):
	//   dam=0  -> d_index=0 -> "misses"
	//   dam=1  -> dampc=10 -> 1 + 10/10 = 2 -> "scratches"
	//   dam=10 -> dampc=100 -> 1 + 100/10 = 11 -> "mauls"
	//   dam=15 -> dampc=150 -> 11 + 50/20 = 13 -> "_traumatizes_"
	//   dam=25 -> dampc=250 -> 16 + 50/100 = 16 -> "_demolishes_"
	//   dam=100 -> dampc=1000 -> 23 -> "**** SMITES ****"
	cases := []struct {
		dam  int
		verb string
	}{
		{0, "misses"},
		{1, "scratches"},
		{10, "mauls"},
		{15, "_traumatizes_"},
		{25, "_demolishes_"},
		{100, "**** SMITES ****"},
	}
	for _, tc := range cases {
		ch, _ := makeDamChar(t, "Alice", 100)
		victim, vclient := makeDamChar(t, "Bob", 100)
		putInDamRoom(ch, victim)

		DamMessage(ch, victim, tc.dam, types.TYPE_HIT, nil)
		out := drain(t, victim, vclient)
		if !strings.Contains(out, tc.verb) {
			t.Errorf("dam=%d: verb %q missing from victim output %q", tc.dam, tc.verb, out)
		}
	}
}

func TestDamMessageTypeHit(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	WorldRef = world.New("/tmp/test")

	ch, cclient := makeDamChar(t, "Alice", 100)
	victim, _ := makeDamChar(t, "Bob", 100)
	putInDamRoom(ch, victim)

	// dam=10 => dampc=100 => d_index=11 => generic: "maul" (singular) / punct '!'
	DamMessage(ch, victim, 10, types.TYPE_HIT, nil)

	out := drain(t, ch, cclient)
	// TO_CHAR: "You <vs> $N<punct>" -> "You maul Bob!"
	if !strings.Contains(out, "You maul Bob!") {
		t.Errorf("TO_CHAR: got %q, want contains %q", out, "You maul Bob!")
	}
}

func TestDamMessageWeaponType(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	WorldRef = world.New("/tmp/test")

	ch, cclient := makeDamChar(t, "Alice", 100)
	victim, _ := makeDamChar(t, "Bob", 100)
	putInDamRoom(ch, victim)

	// DAM_SLICE => blade message table. dam=10 => d_index=11 => "mauls"
	// attack word = attack_table[DAM_SLICE] = "slice".
	// TO_CHAR: "Your slice mauls Bob!"
	DamMessage(ch, victim, 10, types.TYPE_HIT+types.DAM_SLICE, nil)
	out := drain(t, ch, cclient)
	// TO_CHAR uses the singular verb (vs). With DAM_SLICE and dampc=100,
	// d_index=11 -> blade "maul" (singular).
	if !strings.Contains(out, "slice") {
		t.Errorf("weapon-type: expected 'slice' in %q", out)
	}
	if !strings.Contains(out, "maul") {
		t.Errorf("weapon-type: expected 'maul' in %q", out)
	}
}

func TestDamMessageZeroDam(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	WorldRef = world.New("/tmp/test")

	ch, cclient := makeDamChar(t, "Alice", 100)
	victim, _ := makeDamChar(t, "Bob", 100)
	putInDamRoom(ch, victim)

	DamMessage(ch, victim, 0, types.TYPE_HIT, nil)
	out := drain(t, ch, cclient)
	// TO_CHAR verb is singular "miss" from s_generic_messages[0]
	if !strings.Contains(out, "miss") {
		t.Errorf("zero-dam: expected 'miss' in %q", out)
	}
}

func TestDamMessageWithObjShortDescr(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	WorldRef = world.New("/tmp/test")

	ch, cclient := makeDamChar(t, "Alice", 100)
	victim, _ := makeDamChar(t, "Bob", 100)
	putInDamRoom(ch, victim)

	obj := &types.ObjData{ShortDescr: "a glowing sword"}
	// dt > TYPE_HIT and obj != nil should swap obj.ShortDescr in as the attack.
	DamMessage(ch, victim, 10, types.TYPE_HIT+types.DAM_SLICE, obj)
	out := drain(t, ch, cclient)
	if !strings.Contains(out, "a glowing sword") {
		t.Errorf("obj-short: expected 'a glowing sword' in %q", out)
	}
}

func TestDamMessageSkillWithNounDamage(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	w := world.New("/tmp/test")
	// Register a fake skill at sn=1 (sn=0 reserved for "reserved"). Skill[0]
	// is also fine, but give one a recognisable noun.
	skill := &types.SkillType{
		Name:       "searburst",
		NounDamage: "searing energy",
	}
	w.Skills = []*types.SkillType{skill}
	WorldRef = w

	ch, cclient := makeDamChar(t, "Alice", 100)
	victim, _ := makeDamChar(t, "Bob", 100)
	putInDamRoom(ch, victim)

	DamMessage(ch, victim, 10, 0, nil) // sn=0
	out := drain(t, ch, cclient)
	if !strings.Contains(out, "searing energy") {
		t.Errorf("skill noun_damage: expected 'searing energy' in %q", out)
	}
}

func TestDamMessageSkillMissChar(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	w := world.New("/tmp/test")
	skill := &types.SkillType{
		Name:       "swipe",
		NounDamage: "swipe",
		MissChar:   "You swing and whiff.",
	}
	w.Skills = []*types.SkillType{skill}
	WorldRef = w

	ch, cclient := makeDamChar(t, "Alice", 100)
	victim, _ := makeDamChar(t, "Bob", 100)
	putInDamRoom(ch, victim)

	DamMessage(ch, victim, 0, 0, nil) // sn=0, miss
	out := drain(t, ch, cclient)
	if !strings.Contains(out, "You swing and whiff.") {
		t.Errorf("skill miss_char: expected exact miss message in %q", out)
	}
}

func TestDamMessageSkillHitChar(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	w := world.New("/tmp/test")
	skill := &types.SkillType{
		Name:       "fireball",
		NounDamage: "fireball",
		HitChar:    "Your fireball scorches $N!",
	}
	w.Skills = []*types.SkillType{skill}
	WorldRef = w

	ch, cclient := makeDamChar(t, "Alice", 100)
	victim, _ := makeDamChar(t, "Bob", 100)
	putInDamRoom(ch, victim)

	DamMessage(ch, victim, 20, 0, nil)
	out := drain(t, ch, cclient)
	if !strings.Contains(out, "Your fireball scorches Bob!") {
		t.Errorf("skill hit_char: expected custom string, got %q", out)
	}
	// C fight.c:4549-4558 falls through after hit_* — the generic
	// "Your <noun_damage> <verb> $N" line must also appear.
	if !strings.Contains(out, "Your fireball") || !strings.Contains(out, "Bob") {
		t.Errorf("skill hit_char: expected generic verb line to also fire, got %q", out)
	}
	// Count occurrences — we expect at least two distinct "Your fireball" lines:
	// the custom hit_char and the generic noun_damage-based line.
	if strings.Count(out, "Your fireball") < 2 {
		t.Errorf("skill hit_char: expected both custom AND generic line (>=2 occurrences of 'Your fireball'), got %q", out)
	}
}

func TestDamMessageInvalidDtNoPanic(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	WorldRef = world.New("/tmp/test")

	ch, cclient := makeDamChar(t, "Alice", 100)
	victim, _ := makeDamChar(t, "Bob", 100)
	putInDamRoom(ch, victim)

	// dt 9999 is out of range: should bug-log and fall back to DAM_HIT/TYPE_HIT.
	DamMessage(ch, victim, 10, 9999, nil)
	out := drain(t, ch, cclient)
	// Should still emit something ("You maul Bob!" after fallback).
	if out == "" {
		t.Errorf("invalid-dt: expected fallback output, got empty")
	}
}

func TestDamMessageZeroMaxHitGuard(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	WorldRef = world.New("/tmp/test")

	ch, _ := makeDamChar(t, "Alice", 100)
	victim, vclient := makeDamChar(t, "Bob", 0) // MaxHit=0
	victim.Hit = 0
	putInDamRoom(ch, victim)

	// Should not panic on division by zero. dam=0 drives d_index=0 -> "misses"
	DamMessage(ch, victim, 0, types.TYPE_HIT, nil)
	out := drain(t, victim, vclient)
	if !strings.Contains(out, "miss") {
		t.Errorf("zero-maxhit: expected 'miss' in %q", out)
	}
}

func TestDamMessageDifferentRoomsEmitsToVictOnly(t *testing.T) {
	// Attacker in one room, victim in another. Per spec we skip the
	// was_in_room C dance and emit only to the victim.
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	WorldRef = world.New("/tmp/test")

	ch, _ := makeDamChar(t, "Alice", 100)
	victim, vclient := makeDamChar(t, "Bob", 100)
	roomA := &types.RoomIndexData{Vnum: 1}
	roomB := &types.RoomIndexData{Vnum: 2}
	ch.InRoom = roomA
	roomA.People = append(roomA.People, ch)
	victim.InRoom = roomB
	roomB.People = append(roomB.People, victim)

	DamMessage(ch, victim, 10, types.TYPE_HIT, nil)
	out := drain(t, victim, vclient)
	if out == "" {
		t.Errorf("cross-room: expected some output to victim, got empty")
	}
}
