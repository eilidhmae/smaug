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

// --------------------------------------------------------------------
// G1 — was_in_room swap (fight.c:4432-4439, 4590-4594)
// --------------------------------------------------------------------

// TestDamMessage_DifferentRooms_SwapsAttackerIntoVictimRoom covers C's
// was_in_room dance: the attacker is temporarily moved into the victim's
// room so TO_NOTVICT bystanders there see buf1. After the call the
// attacker is restored to their original room and membership is intact.
func TestDamMessage_DifferentRooms_SwapsAttackerIntoVictimRoom(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	WorldRef = world.New("/tmp/test")

	ch, _ := makeDamChar(t, "Alice", 100)
	victim, _ := makeDamChar(t, "Bob", 100)
	bystander, bclient := makeDamChar(t, "Carol", 100)

	roomA := &types.RoomIndexData{Vnum: 1}
	roomB := &types.RoomIndexData{Vnum: 2}
	ch.InRoom = roomA
	roomA.People = append(roomA.People, ch)
	victim.InRoom = roomB
	roomB.People = append(roomB.People, victim)
	bystander.InRoom = roomB
	roomB.People = append(roomB.People, bystander)

	DamMessage(ch, victim, 10, types.TYPE_HIT, nil)

	// Bystander in room B must have seen buf1 ("$n <verb> $N.") — i.e.
	// "Alice mauls Bob!" with punct '!' at dampc=100.
	bout := drain(t, bystander, bclient)
	if !strings.Contains(bout, "Alice") || !strings.Contains(bout, "Bob") {
		t.Errorf("bystander: expected Alice+Bob in %q", bout)
	}
	if !strings.Contains(bout, "maul") {
		t.Errorf("bystander: expected 'maul' verb in %q", bout)
	}

	// After DamMessage returns, ch.InRoom must be roomA (restored).
	if ch.InRoom != roomA {
		t.Errorf("after swap: ch.InRoom = %v, want roomA", ch.InRoom)
	}
	// roomA.People must still contain ch.
	foundInA := false
	for _, p := range roomA.People {
		if p == ch {
			foundInA = true
			break
		}
	}
	if !foundInA {
		t.Errorf("after swap: roomA.People is missing ch")
	}
	// roomB.People must NOT contain ch (only victim + bystander).
	for _, p := range roomB.People {
		if p == ch {
			t.Errorf("after swap: roomB.People still contains ch")
		}
	}
}

// TestDamMessage_DifferentRooms_BystanderInOriginalRoomSeesNothing ensures
// the cross-room swap pulls bystanders from the VICTIM's room only, not the
// attacker's. A bystander back in room A must not receive the message.
func TestDamMessage_DifferentRooms_BystanderInOriginalRoomSeesNothing(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	WorldRef = world.New("/tmp/test")

	ch, _ := makeDamChar(t, "Alice", 100)
	victim, _ := makeDamChar(t, "Bob", 100)
	bystanderA, baclient := makeDamChar(t, "Dave", 100)

	roomA := &types.RoomIndexData{Vnum: 1}
	roomB := &types.RoomIndexData{Vnum: 2}
	ch.InRoom = roomA
	roomA.People = append(roomA.People, ch)
	bystanderA.InRoom = roomA
	roomA.People = append(roomA.People, bystanderA)
	victim.InRoom = roomB
	roomB.People = append(roomB.People, victim)

	DamMessage(ch, victim, 10, types.TYPE_HIT, nil)

	out := drain(t, bystanderA, baclient)
	if out != "" {
		t.Errorf("bystander in attacker's original room should not see cross-room hit; got %q", out)
	}
}

// --------------------------------------------------------------------
// G2 — PCFLAG_GAG self-suppress (fight.c:4481-4488)
// --------------------------------------------------------------------

// TestDamMessage_GaggedAttacker_ZeroDam_NoToChar: gagged attacker with dam=0
// does not see their own TO_CHAR miss, but the victim and bystander still do.
func TestDamMessage_GaggedAttacker_ZeroDam_NoToChar(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	WorldRef = world.New("/tmp/test")

	ch, cclient := makeDamChar(t, "Alice", 100)
	ch.PCData = &types.PCData{Flags: int(types.PCFLAG_GAG)}
	victim, vclient := makeDamChar(t, "Bob", 100)
	bystander, bclient := makeDamChar(t, "Carol", 100)
	putInDamRoom(ch, victim, bystander)

	DamMessage(ch, victim, 0, types.TYPE_HIT, nil)

	cout := drain(t, ch, cclient)
	if cout != "" {
		t.Errorf("gagged attacker should receive no TO_CHAR; got %q", cout)
	}
	vout := drain(t, victim, vclient)
	if vout == "" {
		t.Errorf("victim should still receive TO_VICT; got empty")
	}
	bout := drain(t, bystander, bclient)
	if bout == "" {
		t.Errorf("bystander should still receive TO_NOTVICT; got empty")
	}
}

// TestDamMessage_GaggedVictim_ZeroDam_NoToVict: gagged victim with dam=0
// does not see TO_VICT; attacker and bystander still do.
func TestDamMessage_GaggedVictim_ZeroDam_NoToVict(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	WorldRef = world.New("/tmp/test")

	ch, cclient := makeDamChar(t, "Alice", 100)
	victim, vclient := makeDamChar(t, "Bob", 100)
	victim.PCData = &types.PCData{Flags: int(types.PCFLAG_GAG)}
	bystander, bclient := makeDamChar(t, "Carol", 100)
	putInDamRoom(ch, victim, bystander)

	DamMessage(ch, victim, 0, types.TYPE_HIT, nil)

	vout := drain(t, victim, vclient)
	if vout != "" {
		t.Errorf("gagged victim should receive no TO_VICT; got %q", vout)
	}
	cout := drain(t, ch, cclient)
	if cout == "" {
		t.Errorf("attacker should still receive TO_CHAR; got empty")
	}
	bout := drain(t, bystander, bclient)
	if bout == "" {
		t.Errorf("bystander should still receive TO_NOTVICT; got empty")
	}
}

// TestDamMessage_BothGagged_OnlyBystanderSees covers both flags set
// simultaneously with dam=0 — only TO_NOTVICT delivers.
func TestDamMessage_BothGagged_OnlyBystanderSees(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	WorldRef = world.New("/tmp/test")

	ch, cclient := makeDamChar(t, "Alice", 100)
	ch.PCData = &types.PCData{Flags: int(types.PCFLAG_GAG)}
	victim, vclient := makeDamChar(t, "Bob", 100)
	victim.PCData = &types.PCData{Flags: int(types.PCFLAG_GAG)}
	bystander, bclient := makeDamChar(t, "Carol", 100)
	putInDamRoom(ch, victim, bystander)

	DamMessage(ch, victim, 0, types.TYPE_HIT, nil)

	if cout := drain(t, ch, cclient); cout != "" {
		t.Errorf("gagged attacker: want empty, got %q", cout)
	}
	if vout := drain(t, victim, vclient); vout != "" {
		t.Errorf("gagged victim: want empty, got %q", vout)
	}
	if bout := drain(t, bystander, bclient); bout == "" {
		t.Errorf("bystander: want output, got empty")
	}
}

// TestDamMessage_GaggedAttacker_PositiveDam_AllDeliver: PCFLAG_GAG only
// silences zero-damage misses. With dam > 0 every recipient still sees.
func TestDamMessage_GaggedAttacker_PositiveDam_AllDeliver(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	WorldRef = world.New("/tmp/test")

	ch, cclient := makeDamChar(t, "Alice", 100)
	ch.PCData = &types.PCData{Flags: int(types.PCFLAG_GAG)}
	victim, vclient := makeDamChar(t, "Bob", 100)
	victim.PCData = &types.PCData{Flags: int(types.PCFLAG_GAG)}
	bystander, bclient := makeDamChar(t, "Carol", 100)
	putInDamRoom(ch, victim, bystander)

	DamMessage(ch, victim, 10, types.TYPE_HIT, nil)

	if cout := drain(t, ch, cclient); cout == "" {
		t.Errorf("dam=1 with GAG: attacker should still see TO_CHAR")
	}
	if vout := drain(t, victim, vclient); vout == "" {
		t.Errorf("dam=1 with GAG: victim should still see TO_VICT")
	}
	if bout := drain(t, bystander, bclient); bout == "" {
		t.Errorf("dam=1 with GAG: bystander should still see TO_NOTVICT")
	}
}

// TestDamMessage_NPCAttackerVictimNoGagCrash ensures PCData == nil NPCs on
// either side do not panic through the gag branch.
func TestDamMessage_NPCAttackerVictimNoGagCrash(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	WorldRef = world.New("/tmp/test")

	ch, _ := makeDamChar(t, "mob1", 100)
	ch.Act.Set(types.ACT_IS_NPC)
	victim, vclient := makeDamChar(t, "mob2", 100)
	victim.Act.Set(types.ACT_IS_NPC)
	putInDamRoom(ch, victim)

	// Zero-dam NPC-vs-NPC hit: gag branches must skip both sides safely.
	DamMessage(ch, victim, 0, types.TYPE_HIT, nil)

	out := drain(t, victim, vclient)
	if !strings.Contains(out, "miss") {
		t.Errorf("npc gag guard: expected 'miss' in %q", out)
	}
}

// --------------------------------------------------------------------
// G3 — is_wielding_poisoned prefix (fight.c:4496-4512)
// --------------------------------------------------------------------

// TestDamMessage_PoisonedWield_PrimarySlot: poisoned obj equipped in
// WEAR_WIELD and threaded through DamMessage emits the "poisoned" prefix.
func TestDamMessage_PoisonedWield_PrimarySlot(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	WorldRef = world.New("/tmp/test")

	ch, cclient := makeDamChar(t, "Alice", 100)
	victim, _ := makeDamChar(t, "Bob", 100)
	putInDamRoom(ch, victim)

	obj := &types.ObjData{WearLoc: types.WEAR_WIELD}
	obj.ExtraFlags.Set(types.ITEM_POISONED)
	ch.Carrying = append(ch.Carrying, obj)

	DamMessage(ch, victim, 10, types.TYPE_HIT+types.DAM_SLICE, obj)
	out := drain(t, ch, cclient)
	if !strings.Contains(out, "poisoned slice") {
		t.Errorf("poisoned WEAR_WIELD: expected 'poisoned slice' in %q", out)
	}
}

// TestDamMessage_PoisonedWield_DualSlot: poisoned obj equipped in
// WEAR_DUAL_WIELD (primary slot empty) threaded through DamMessage still
// matches via the identity check against the dual-wield slot.
func TestDamMessage_PoisonedWield_DualSlot(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	WorldRef = world.New("/tmp/test")

	ch, cclient := makeDamChar(t, "Alice", 100)
	victim, _ := makeDamChar(t, "Bob", 100)
	putInDamRoom(ch, victim)

	obj := &types.ObjData{WearLoc: types.WEAR_DUAL_WIELD}
	obj.ExtraFlags.Set(types.ITEM_POISONED)
	ch.Carrying = append(ch.Carrying, obj)

	DamMessage(ch, victim, 10, types.TYPE_HIT+types.DAM_SLICE, obj)
	out := drain(t, ch, cclient)
	if !strings.Contains(out, "poisoned slice") {
		t.Errorf("poisoned WEAR_DUAL_WIELD: expected 'poisoned slice' in %q", out)
	}
}

// TestDamMessage_PoisonedButNotEquipped_NoPrefix: identity check must reject
// a poisoned obj that is not equipped in either wield slot.
func TestDamMessage_PoisonedButNotEquipped_NoPrefix(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	WorldRef = world.New("/tmp/test")

	ch, cclient := makeDamChar(t, "Alice", 100)
	victim, _ := makeDamChar(t, "Bob", 100)
	putInDamRoom(ch, victim)

	// Obj is poisoned but ch has it at WEAR_NONE — not equipped.
	obj := &types.ObjData{WearLoc: types.WEAR_NONE}
	obj.ExtraFlags.Set(types.ITEM_POISONED)
	ch.Carrying = append(ch.Carrying, obj)

	DamMessage(ch, victim, 10, types.TYPE_HIT+types.DAM_SLICE, obj)
	out := drain(t, ch, cclient)
	if strings.Contains(out, "poisoned") {
		t.Errorf("unequipped poisoned: 'poisoned' should NOT appear in %q", out)
	}
}

// TestDamMessage_WieldWithoutPoisonFlag_NoPrefix: equipped in WEAR_WIELD
// but ITEM_POISONED not set — no prefix.
func TestDamMessage_WieldWithoutPoisonFlag_NoPrefix(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	WorldRef = world.New("/tmp/test")

	ch, cclient := makeDamChar(t, "Alice", 100)
	victim, _ := makeDamChar(t, "Bob", 100)
	putInDamRoom(ch, victim)

	obj := &types.ObjData{WearLoc: types.WEAR_WIELD}
	ch.Carrying = append(ch.Carrying, obj)

	DamMessage(ch, victim, 10, types.TYPE_HIT+types.DAM_SLICE, obj)
	out := drain(t, ch, cclient)
	if strings.Contains(out, "poisoned") {
		t.Errorf("wield without poison: 'poisoned' should NOT appear in %q", out)
	}
}

// TestDamMessage_BareHands_PoisonedObjIgnored: dt == TYPE_HIT (bare hands)
// must never get a poisoned prefix even if a poisoned obj is threaded.
// Guard in C is `dt > TYPE_HIT`.
func TestDamMessage_BareHands_PoisonedObjIgnored(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	WorldRef = world.New("/tmp/test")

	ch, cclient := makeDamChar(t, "Alice", 100)
	victim, _ := makeDamChar(t, "Bob", 100)
	putInDamRoom(ch, victim)

	obj := &types.ObjData{WearLoc: types.WEAR_WIELD}
	obj.ExtraFlags.Set(types.ITEM_POISONED)
	ch.Carrying = append(ch.Carrying, obj)

	DamMessage(ch, victim, 10, types.TYPE_HIT, obj)
	out := drain(t, ch, cclient)
	if strings.Contains(out, "poisoned") {
		t.Errorf("bare-hands TYPE_HIT with poisoned obj: prefix must NOT appear; got %q", out)
	}
}

// TestDamMessage_SkillPathPoisonedBranchSkipped: when dt is a valid sn (skill
// path) the poisoned-prefix branch is not entered — the existing skill noun
// formatting wins.
func TestDamMessage_SkillPathPoisonedBranchSkipped(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	w := world.New("/tmp/test")
	skill := &types.SkillType{
		Name:       "searburst",
		NounDamage: "searing energy",
	}
	w.Skills = []*types.SkillType{skill}
	WorldRef = w

	ch, cclient := makeDamChar(t, "Alice", 100)
	victim, _ := makeDamChar(t, "Bob", 100)
	putInDamRoom(ch, victim)

	// Even with a poisoned wield equipped, the skill (sn=0) path must not
	// pick up "poisoned" — C's branch is gated to `dt > TYPE_HIT`.
	obj := &types.ObjData{WearLoc: types.WEAR_WIELD}
	obj.ExtraFlags.Set(types.ITEM_POISONED)
	ch.Carrying = append(ch.Carrying, obj)

	DamMessage(ch, victim, 10, 0, obj) // sn=0
	out := drain(t, ch, cclient)
	if !strings.Contains(out, "searing energy") {
		t.Errorf("skill path: expected noun_damage 'searing energy' in %q", out)
	}
	if strings.Contains(out, "poisoned") {
		t.Errorf("skill path: 'poisoned' prefix must NOT appear in %q", out)
	}
}

// TestDamMessage_PoisonedNilObj_NoPrefix: obj == nil can never produce a
// poisoned prefix.
func TestDamMessage_PoisonedNilObj_NoPrefix(t *testing.T) {
	prev := WorldRef
	resetCombatWorldRef(t, prev)
	WorldRef = world.New("/tmp/test")

	ch, cclient := makeDamChar(t, "Alice", 100)
	victim, _ := makeDamChar(t, "Bob", 100)
	putInDamRoom(ch, victim)

	DamMessage(ch, victim, 10, types.TYPE_HIT+types.DAM_SLICE, nil)
	out := drain(t, ch, cclient)
	if strings.Contains(out, "poisoned") {
		t.Errorf("nil obj: 'poisoned' must NOT appear in %q", out)
	}
}
