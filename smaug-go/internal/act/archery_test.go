package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// --- G2: FindQuiver / FindProjectile helpers ----------------------------

func setupArcheryWorld() *world.World {
	w := world.New("/tmp/test")
	WorldRef = w
	return w
}

// makeObj builds an unrooted ObjData — tests attach it to a char or container.
func makeArcheryObj(name string, itemType int) *types.ObjData {
	return &types.ObjData{
		Name:       name,
		ShortDescr: name,
		ItemType:   itemType,
		WearLoc:    types.WEAR_NONE,
	}
}

func TestFindQuiver_ReturnsFirstOpenQuiver(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()

	quiver := makeArcheryObj("quiver", types.ITEM_QUIVER)
	// Value[1] = container-flags; CONT_CLOSED bit clear = open.
	quiver.Value[1] = 0
	handler.ObjToChar(quiver, ch)

	got := FindQuiver(ch)
	if got != quiver {
		t.Fatalf("FindQuiver = %v, want %v", got, quiver)
	}
}

func TestFindQuiver_SkipsClosedQuiver(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()

	quiver := makeArcheryObj("quiver", types.ITEM_QUIVER)
	quiver.Value[1] = int(types.CONT_CLOSED)
	handler.ObjToChar(quiver, ch)

	got := FindQuiver(ch)
	if got != nil {
		t.Fatalf("FindQuiver should skip closed quiver, got %v", got)
	}
}

func TestFindQuiver_SkipsNonQuivers(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()

	bag := makeArcheryObj("bag", types.ITEM_CONTAINER)
	handler.ObjToChar(bag, ch)

	got := FindQuiver(ch)
	if got != nil {
		t.Fatalf("FindQuiver should skip non-quiver container, got %v", got)
	}
}

func TestFindQuiver_ReturnsLastOpenQuiver_BackWalkSemantics(t *testing.T) {
	// C iterates ch->last_carrying backwards. Go stores Carrying as a slice
	// appended on ObjToChar. A back-walk equivalent returns the LAST appended
	// open quiver first. With two open quivers, the one added second wins.
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()

	q1 := makeArcheryObj("old quiver", types.ITEM_QUIVER)
	q1.Value[1] = 0
	handler.ObjToChar(q1, ch)
	q2 := makeArcheryObj("new quiver", types.ITEM_QUIVER)
	q2.Value[1] = 0
	handler.ObjToChar(q2, ch)

	got := FindQuiver(ch)
	if got != q2 {
		t.Fatalf("FindQuiver should return most-recently-carried open quiver, got %v want %v", got, q2)
	}
}

func TestFindProjectile_ReturnsFirstProjectile(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()

	quiver := makeArcheryObj("quiver", types.ITEM_QUIVER)
	handler.ObjToChar(quiver, ch)
	arrow := makeArcheryObj("steel arrow", types.ITEM_PROJECTILE)
	handler.ObjToObj(arrow, quiver)

	got := FindProjectile(ch, quiver)
	if got != arrow {
		t.Fatalf("FindProjectile = %v, want %v", got, arrow)
	}
}

func TestFindProjectile_EmptyQuiverReturnsNil(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()

	quiver := makeArcheryObj("quiver", types.ITEM_QUIVER)
	handler.ObjToChar(quiver, ch)

	got := FindProjectile(ch, quiver)
	if got != nil {
		t.Fatalf("FindProjectile on empty quiver = %v, want nil", got)
	}
}

func TestFindProjectile_SkipsNonProjectiles(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()

	quiver := makeArcheryObj("quiver", types.ITEM_QUIVER)
	handler.ObjToChar(quiver, ch)
	trinket := makeArcheryObj("lucky rock", types.ITEM_TRASH)
	handler.ObjToObj(trinket, quiver)

	got := FindProjectile(ch, quiver)
	if got != nil {
		t.Fatalf("FindProjectile should skip non-projectile, got %v", got)
	}
}

// --- G3: DoDraw ---------------------------------------------------------

// makeBowArrowQuiver wires up a bow worn at WEAR_MISSILE_WIELD, a quiver
// in ch.Carrying with one arrow inside, ammo types matched. Returns
// (bow, quiver, arrow).
func makeBowArrowQuiver(ch *types.CharData) (*types.ObjData, *types.ObjData, *types.ObjData) {
	bow := makeArcheryObj("longbow", types.ITEM_MISSILE_WEAPON)
	// bow.Value[5] = accepted PROJ_* kind.
	bow.Value[5] = types.PROJ_ARROW
	bow.Value[4] = 5 // max dist
	bow.WearLoc = types.WEAR_MISSILE_WIELD
	ch.Carrying = append(ch.Carrying, bow)
	bow.CarriedBy = ch

	quiver := makeArcheryObj("quiver", types.ITEM_QUIVER)
	quiver.Value[1] = 0
	handler.ObjToChar(quiver, ch)

	arrow := makeArcheryObj("steel arrow", types.ITEM_PROJECTILE)
	// arrow.Value[4] = ammo type (must equal bow.Value[5]).
	arrow.Value[4] = types.PROJ_ARROW
	arrow.Value[1] = 2
	arrow.Value[2] = 4
	handler.ObjToObj(arrow, quiver)

	return bow, quiver, arrow
}

func TestDoDraw_NoBow(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9001, Name: "Range"}
	handler.CharToRoom(ch, room)

	DoDraw(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "not wielding a missile weapon") {
		t.Errorf("expected missile-weapon error, got: %q", out)
	}
}

func TestDoDraw_NoQuiver(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9001, Name: "Range"}
	handler.CharToRoom(ch, room)

	bow := makeArcheryObj("longbow", types.ITEM_MISSILE_WEAPON)
	bow.Value[5] = types.PROJ_ARROW
	bow.WearLoc = types.WEAR_MISSILE_WIELD
	ch.Carrying = append(ch.Carrying, bow)
	bow.CarriedBy = ch

	DoDraw(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "quiver") {
		t.Errorf("expected quiver error, got: %q", out)
	}
}

func TestDoDraw_HandsBusy(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9001, Name: "Range"}
	handler.CharToRoom(ch, room)

	_, _, _ = makeBowArrowQuiver(ch)
	// Two busy hands: WEAR_LIGHT + WEAR_SHIELD.
	light := makeArcheryObj("torch", types.ITEM_LIGHT)
	light.WearLoc = types.WEAR_LIGHT
	ch.Carrying = append(ch.Carrying, light)
	light.CarriedBy = ch
	shield := makeArcheryObj("buckler", types.ITEM_ARMOR)
	shield.WearLoc = types.WEAR_SHIELD
	ch.Carrying = append(ch.Carrying, shield)
	shield.CarriedBy = ch

	DoDraw(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "free hand") {
		t.Errorf("expected free-hand error, got: %q", out)
	}
}

func TestDoDraw_HoldOccupied(t *testing.T) {
	// Hand-count is exactly 1 (only WEAR_HOLD), which lets the first
	// gate pass; the second gate (WEAR_HOLD explicitly occupied) fires.
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9001, Name: "Range"}
	handler.CharToRoom(ch, room)

	_, _, _ = makeBowArrowQuiver(ch)
	held := makeArcheryObj("lantern", types.ITEM_LIGHT)
	held.WearLoc = types.WEAR_HOLD
	ch.Carrying = append(ch.Carrying, held)
	held.CarriedBy = ch

	DoDraw(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "not empty") {
		t.Errorf("expected hand-not-empty error, got: %q", out)
	}
}

func TestDoDraw_EmptyQuiver(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9001, Name: "Range"}
	handler.CharToRoom(ch, room)

	bow := makeArcheryObj("longbow", types.ITEM_MISSILE_WEAPON)
	bow.Value[5] = types.PROJ_ARROW
	bow.WearLoc = types.WEAR_MISSILE_WIELD
	ch.Carrying = append(ch.Carrying, bow)
	bow.CarriedBy = ch
	quiver := makeArcheryObj("quiver", types.ITEM_QUIVER)
	quiver.Value[1] = 0
	handler.ObjToChar(quiver, ch)

	DoDraw(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "quiver is empty") {
		t.Errorf("expected empty-quiver error, got: %q", out)
	}
}

func TestDoDraw_WrongAmmoType(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9001, Name: "Range"}
	handler.CharToRoom(ch, room)

	bow, quiver, arrow := makeBowArrowQuiver(ch)
	bow.Value[5] = types.PROJ_BOLT // bow expects bolts
	arrow.Value[4] = types.PROJ_ARROW

	DoDraw(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "wrong projectile") {
		t.Errorf("expected wrong-ammo error, got: %q", out)
	}
	// Arrow should be put back into the quiver after the error.
	if arrow.InObj != quiver {
		t.Errorf("arrow.InObj = %v, want back in quiver", arrow.InObj)
	}
}

func TestDoDraw_DualWieldOmittedFromHandCount_CBugPreserved(t *testing.T) {
	// C src/archery.c:142-149 counts LIGHT+SHIELD+HOLD+WIELD but NOT
	// DUAL_WIELD. A character with MISSILE_WIELD (bow) + DUAL_WIELD
	// (offhand) + empty HOLD has hand_count = 0 and draws successfully.
	// This is a known C bug / quirk; preserve verbatim.
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9003, Name: "Range"}
	handler.CharToRoom(ch, room)

	_, _, arrow := makeBowArrowQuiver(ch)
	// Equip a WIELD + a DUAL_WIELD weapon. hand_count = 1 (WIELD only)
	// when DUAL is excluded per C; mutation adding DUAL makes it 2.
	main := makeArcheryObj("sword", types.ITEM_WEAPON)
	main.WearLoc = types.WEAR_WIELD
	ch.Carrying = append(ch.Carrying, main)
	main.CarriedBy = ch
	offhand := makeArcheryObj("dagger", types.ITEM_WEAPON)
	offhand.WearLoc = types.WEAR_DUAL_WIELD
	ch.Carrying = append(ch.Carrying, offhand)
	offhand.CarriedBy = ch

	DoDraw(ch, "")
	out := readOutput(ch, client)
	// Should succeed — no 'free hand' error.
	if strings.Contains(out, "free hand") {
		t.Errorf("DUAL_WIELD should NOT block draw (C bug preserved), got: %q", out)
	}
	if arrow.WearLoc != types.WEAR_HOLD {
		t.Errorf("arrow.WearLoc = %d, want WEAR_HOLD (draw should have succeeded)", arrow.WearLoc)
	}
}

func TestDoDraw_Success(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9001, Name: "Range"}
	handler.CharToRoom(ch, room)

	_, _, arrow := makeBowArrowQuiver(ch)

	ch.Wait = 0
	DoDraw(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "You draw") {
		t.Errorf("expected draw confirmation, got: %q", out)
	}
	// Arrow now on character, WEAR_HOLD slot.
	if arrow.CarriedBy != ch {
		t.Errorf("arrow.CarriedBy = %v, want %v", arrow.CarriedBy, ch)
	}
	if arrow.InObj != nil {
		t.Errorf("arrow.InObj = %v, want nil after draw", arrow.InObj)
	}
	if arrow.WearLoc != types.WEAR_HOLD {
		t.Errorf("arrow.WearLoc = %d, want WEAR_HOLD (%d)", arrow.WearLoc, types.WEAR_HOLD)
	}
	if ch.Wait != types.PULSE_VIOLENCE {
		t.Errorf("ch.Wait = %d, want PULSE_VIOLENCE (%d)", ch.Wait, types.PULSE_VIOLENCE)
	}
}

// --- G4: DoDislodge -----------------------------------------------------

// makeLodgedArrow equips a lodged arrow at the given WEAR_LODGE_* slot
// on ch. Value[1]/Value[2] are the damage dice range.
func makeLodgedArrow(ch *types.CharData, slot, low, high int) *types.ObjData {
	arrow := makeArcheryObj("steel arrow", types.ITEM_PROJECTILE)
	arrow.Value[1] = low
	arrow.Value[2] = high
	arrow.ExtraFlags.Set(types.ITEM_LODGED)
	switch slot {
	case types.WEAR_LODGE_RIB:
		arrow.WearFlags |= int(types.ITEM_LODGE_RIB)
	case types.WEAR_LODGE_ARM:
		arrow.WearFlags |= int(types.ITEM_LODGE_ARM)
	case types.WEAR_LODGE_LEG:
		arrow.WearFlags |= int(types.ITEM_LODGE_LEG)
	}
	arrow.WearLoc = slot
	ch.Carrying = append(ch.Carrying, arrow)
	arrow.CarriedBy = ch
	return arrow
}

func TestDoDislodge_NoArg(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9001, Name: "Range"}
	handler.CharToRoom(ch, room)
	ch.Hit = 100

	DoDislodge(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Dislodge what") {
		t.Errorf("expected dislodge-what prompt, got: %q", out)
	}
}

func TestDoDislodge_NothingLodged(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9001, Name: "Range"}
	handler.CharToRoom(ch, room)
	ch.Hit = 100

	DoDislodge(ch, "arrow")
	out := readOutput(ch, client)
	if !strings.Contains(out, "nothing lodged") {
		t.Errorf("expected nothing-lodged error, got: %q", out)
	}
}

func TestDoDislodge_RibSuccess(t *testing.T) {
	// Rib damage = number_range(3*val[1], 3*val[2]). With val[1]=val[2]=2,
	// damage is deterministic = 6.
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9001, Name: "Range"}
	handler.CharToRoom(ch, room)
	ch.Hit = 100
	ch.MaxHit = 100

	arrow := makeLodgedArrow(ch, types.WEAR_LODGE_RIB, 2, 2)

	DoDislodge(ch, "arrow")
	_ = readOutput(ch, client)

	if arrow.WearLoc != types.WEAR_NONE {
		t.Errorf("arrow.WearLoc = %d, want WEAR_NONE after dislodge", arrow.WearLoc)
	}
	if arrow.WearFlags&int(types.ITEM_LODGE_RIB) != 0 {
		t.Errorf("ITEM_LODGE_RIB still set on arrow.WearFlags")
	}
	if arrow.ExtraFlags.IsSet(types.ITEM_LODGED) {
		t.Errorf("ITEM_LODGED still set on arrow")
	}
	// Arrow still in inventory.
	if arrow.CarriedBy != ch {
		t.Errorf("arrow.CarriedBy = %v, want %v (arrow should stay in inventory)", arrow.CarriedBy, ch)
	}
	// Damage taken = exactly 6 (3 * 2 .. 3 * 2).
	if ch.Hit != 94 {
		t.Errorf("ch.Hit = %d, want 94 (100 - 6 rib damage)", ch.Hit)
	}
}

func TestDoDislodge_ArmSuccess_PreservesAsymmetry(t *testing.T) {
	// Arm damage = number_range(3*val[1], 2*val[2]). C asymmetry at
	// src/archery.c:234 (rib is 3/3, leg is 2/2, but arm is 3/2).
	// With val[1]=2 val[2]=3: damage = number_range(6, 6) = 6.
	// Mutation: if anyone swaps to 3*val[2], formula becomes
	// number_range(6, 9) which varies — this deterministic check pins
	// the 2*val[2] upper bound.
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9001, Name: "Range"}
	handler.CharToRoom(ch, room)
	ch.Hit = 100
	ch.MaxHit = 100

	arrow := makeLodgedArrow(ch, types.WEAR_LODGE_ARM, 2, 3)

	DoDislodge(ch, "arrow")
	_ = readOutput(ch, client)

	damTaken := 100 - ch.Hit
	if damTaken != 6 {
		t.Errorf("arm damage = %d, want 6 (C asymmetry 3*val[1]..2*val[2] with 2,3 → range(6,6)=6)", damTaken)
	}
	if arrow.WearFlags&int(types.ITEM_LODGE_ARM) != 0 {
		t.Errorf("ITEM_LODGE_ARM still set")
	}
}

func TestDoDislodge_ArmSuccess_UpperBoundAsymmetry(t *testing.T) {
	// Probabilistic bound test: with val[1]=1 val[2]=10, arm formula
	// is number_range(3, 20). Max possible damage is 20. If the formula
	// is mutated to 3*val[2] (rib-style), max becomes 30. Run 200
	// iterations and assert damage never exceeds 20. RNG is crypto-seeded
	// in util.init(), so this is effectively non-flaky for a 200-sample
	// window — expected hits at 21..30 under a mutated formula would be
	// ≈100 per run (10-bucket uniform from [3..30], so ≈37% chance of
	// exceeding 20 on each iteration → vanishingly small false-pass).
	for iter := 0; iter < 200; iter++ {
		_ = setupArcheryWorld()
		ch, client := makeTestChar("Archer")
		room := &types.RoomIndexData{Vnum: 9002, Name: "Range"}
		handler.CharToRoom(ch, room)
		ch.Hit = 100
		ch.MaxHit = 100

		_ = makeLodgedArrow(ch, types.WEAR_LODGE_ARM, 1, 10)
		DoDislodge(ch, "arrow")
		_ = readOutput(ch, client)
		client.Close()

		damTaken := 100 - ch.Hit
		if damTaken > 20 {
			t.Fatalf("iter %d: arm damage %d > 20 upper bound (arm is 3*val[1]..2*val[2]; must not exceed 2*val[2]=20)", iter, damTaken)
		}
		if damTaken < 3 {
			t.Fatalf("iter %d: arm damage %d < 3 lower bound (arm is 3*val[1]..2*val[2]; must be >= 3*val[1]=3)", iter, damTaken)
		}
	}
}

func TestDoDislodge_LegSuccess(t *testing.T) {
	// Leg damage = number_range(2*val[1], 2*val[2]). val[1]=val[2]=2 → 4.
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9001, Name: "Range"}
	handler.CharToRoom(ch, room)
	ch.Hit = 100
	ch.MaxHit = 100

	arrow := makeLodgedArrow(ch, types.WEAR_LODGE_LEG, 2, 2)

	DoDislodge(ch, "arrow")
	_ = readOutput(ch, client)

	if ch.Hit != 96 {
		t.Errorf("ch.Hit = %d, want 96 (100 - 4 leg damage)", ch.Hit)
	}
	if arrow.WearFlags&int(types.ITEM_LODGE_LEG) != 0 {
		t.Errorf("ITEM_LODGE_LEG still set")
	}
}

// --- G5: DoFire / RangedAttack / ScanForVictim / RangedGotTarget / projectileHit ---

// makeFireFixture wires a bow wielded + arrow held. Returns bow, arrow.
// The bow accepts PROJ_ARROW ammunition at max_dist=3.
func makeFireFixture(ch *types.CharData) (*types.ObjData, *types.ObjData) {
	bow := makeArcheryObj("longbow", types.ITEM_MISSILE_WEAPON)
	bow.Value[1] = 1
	bow.Value[2] = 3
	bow.Value[3] = types.DAM_ARROW
	bow.Value[4] = 3 // max_dist
	bow.Value[5] = types.PROJ_ARROW
	bow.WearLoc = types.WEAR_MISSILE_WIELD
	ch.Carrying = append(ch.Carrying, bow)
	bow.CarriedBy = ch

	arrow := makeArcheryObj("steel arrow", types.ITEM_PROJECTILE)
	arrow.Value[1] = 1
	arrow.Value[2] = 2
	arrow.Value[3] = types.DAM_ARROW
	arrow.Value[4] = types.PROJ_ARROW
	arrow.Value[5] = types.PROJ_ARROW
	arrow.WearLoc = types.WEAR_HOLD
	ch.Carrying = append(ch.Carrying, arrow)
	arrow.CarriedBy = ch

	return bow, arrow
}

func TestDoFire_NoBow(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9100, Name: "Range"}
	handler.CharToRoom(ch, room)

	DoFire(ch, "north orc")
	out := readOutput(ch, client)
	if !strings.Contains(out, "not wielding a missile weapon") {
		t.Errorf("expected missile-weapon error, got: %q", out)
	}
}

func TestDoFire_NoArgNoFight(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9100, Name: "Range"}
	handler.CharToRoom(ch, room)
	_, _ = makeFireFixture(ch)

	DoFire(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "whom or what") {
		t.Errorf("expected whom-or-what prompt, got: %q", out)
	}
}

func TestDoFire_SelfArg(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9100, Name: "Range"}
	handler.CharToRoom(ch, room)
	_, _ = makeFireFixture(ch)

	DoFire(ch, "self")
	out := readOutput(ch, client)
	if !strings.Contains(out, "yourself") {
		t.Errorf("expected self-fire block, got: %q", out)
	}
}

func TestDoFire_NoneArg(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9100, Name: "Range"}
	handler.CharToRoom(ch, room)
	_, _ = makeFireFixture(ch)

	DoFire(ch, "none")
	out := readOutput(ch, client)
	if !strings.Contains(out, "yourself") {
		t.Errorf("expected self-fire block for 'none', got: %q", out)
	}
}

func TestDoFire_VictimEqualsCh_DeadCheck_CBugPreserved(t *testing.T) {
	// C src/archery.c:1255,1272 has a structurally-dead check:
	//   CHAR_DATA *victim = NULL;
	//   ...
	//   if (!str_cmp(arg, "none") || !str_cmp(arg, "self") || victim == ch)
	// The `victim == ch` clause never fires because victim is never
	// assigned before the check. The string-matched arms ("none"/"self")
	// carry the full weight of this gate. Port preserves the dead branch
	// verbatim. This test validates that firing at an arbitrary non-self,
	// non-"none" argument name does NOT trigger the self-fire block —
	// i.e., the dead branch stays dead in Go.
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9101, Name: "Range"}
	handler.CharToRoom(ch, room)
	_, _ = makeFireFixture(ch)

	DoFire(ch, "someotherarg")
	out := readOutput(ch, client)
	// "someotherarg" is not "none"/"self" and victim is never bound to
	// ch — so self-fire block must NOT fire. Instead we expect the
	// downstream "Aim in what direction?" error (no such exit/victim).
	if strings.Contains(out, "yourself") {
		t.Errorf("victim==ch dead check should not fire for non-self arg, got: %q", out)
	}
}

func TestDoFire_NotHoldingProjectile(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9100, Name: "Range"}
	handler.CharToRoom(ch, room)
	bow, arrow := makeFireFixture(ch)
	_ = bow
	// Unequip the arrow so HOLD is empty.
	arrow.WearLoc = types.WEAR_NONE

	DoFire(ch, "target")
	out := readOutput(ch, client)
	if !strings.Contains(out, "not holding a projectile") {
		t.Errorf("expected no-projectile error, got: %q", out)
	}
}

func TestDoFire_WrongAmmoType_Arrows(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9100, Name: "Range"}
	handler.CharToRoom(ch, room)
	bow, arrow := makeFireFixture(ch)
	bow.Value[5] = types.PROJ_ARROW
	arrow.Value[4] = types.PROJ_BOLT // mismatch

	DoFire(ch, "target")
	out := readOutput(ch, client)
	if !strings.Contains(out, "no arrows") {
		t.Errorf("expected 'no arrows' error, got: %q", out)
	}
}

func TestDoFire_WrongAmmoType_Bolts(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9100, Name: "Range"}
	handler.CharToRoom(ch, room)
	bow, arrow := makeFireFixture(ch)
	bow.Value[5] = types.PROJ_BOLT
	arrow.Value[4] = types.PROJ_ARROW

	DoFire(ch, "target")
	out := readOutput(ch, client)
	if !strings.Contains(out, "no bolts") {
		t.Errorf("expected 'no bolts' error, got: %q", out)
	}
}

// --- RangedAttack gates ---

func TestRangedAttack_EmptyArg(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9100, Name: "Range"}
	handler.CharToRoom(ch, room)
	bow, arrow := makeFireFixture(ch)

	rangedAttack(ch, "", bow, arrow, types.TYPE_HIT+types.DAM_ARROW, 3)
	out := readOutput(ch, client)
	if !strings.Contains(out, "Where") {
		t.Errorf("expected 'Where? At who?' prompt, got: %q", out)
	}
}

func TestRangedAttack_WallNoDestination(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9100, Name: "Range"}
	// An exit with ToRoom == nil → wall.
	exit := &types.ExitData{Direction: 0, Vnum: 0, ToRoom: nil}
	room.Exits = append(room.Exits, exit)
	handler.CharToRoom(ch, room)
	bow, arrow := makeFireFixture(ch)

	rangedAttack(ch, "north target", bow, arrow, types.TYPE_HIT+types.DAM_ARROW, 3)
	out := readOutput(ch, client)
	if !strings.Contains(out, "wall") {
		t.Errorf("expected 'wall' error, got: %q", out)
	}
}

func TestRangedAttack_ClosedDoor(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	here := &types.RoomIndexData{Vnum: 9100, Name: "Here"}
	there := &types.RoomIndexData{Vnum: 9101, Name: "There"}
	exit := &types.ExitData{Direction: 0, Vnum: 9101, ToRoom: there,
		ExitInfo: int(types.EX_CLOSED)}
	here.Exits = append(here.Exits, exit)
	handler.CharToRoom(ch, here)
	bow, arrow := makeFireFixture(ch)

	rangedAttack(ch, "north target", bow, arrow, types.TYPE_HIT+types.DAM_ARROW, 3)
	out := readOutput(ch, client)
	if !strings.Contains(out, "door") {
		t.Errorf("expected 'door' error, got: %q", out)
	}
}

func TestRangedAttack_SecretClosedIsWall(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	here := &types.RoomIndexData{Vnum: 9100, Name: "Here"}
	there := &types.RoomIndexData{Vnum: 9101, Name: "There"}
	exit := &types.ExitData{Direction: 0, Vnum: 9101, ToRoom: there,
		ExitInfo: int(types.EX_CLOSED | types.EX_SECRET)}
	here.Exits = append(here.Exits, exit)
	handler.CharToRoom(ch, here)
	bow, arrow := makeFireFixture(ch)

	rangedAttack(ch, "north target", bow, arrow, types.TYPE_HIT+types.DAM_ARROW, 3)
	out := readOutput(ch, client)
	if !strings.Contains(out, "wall") {
		t.Errorf("expected 'wall' error for EX_SECRET, got: %q", out)
	}
}

func TestRangedAttack_PrivateRoom(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	here := &types.RoomIndexData{Vnum: 9100, Name: "Private"}
	here.RoomFlags.Set(types.ROOM_PRIVATE)
	there := &types.RoomIndexData{Vnum: 9101, Name: "There"}
	exit := &types.ExitData{Direction: 0, Vnum: 9101, ToRoom: there}
	here.Exits = append(here.Exits, exit)
	handler.CharToRoom(ch, here)
	bow, arrow := makeFireFixture(ch)

	rangedAttack(ch, "north target", bow, arrow, types.TYPE_HIT+types.DAM_ARROW, 3)
	out := readOutput(ch, client)
	if !strings.Contains(out, "private room") {
		t.Errorf("expected private-room error, got: %q", out)
	}
}

func TestRangedAttack_NoDirectionNoVictim(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9100, Name: "Range"}
	handler.CharToRoom(ch, room)
	bow, arrow := makeFireFixture(ch)

	rangedAttack(ch, "gibberish", bow, arrow, types.TYPE_HIT+types.DAM_ARROW, 3)
	out := readOutput(ch, client)
	if !strings.Contains(out, "Aim in what direction") {
		t.Errorf("expected aim-direction error, got: %q", out)
	}
}

func TestRangedAttack_SameRoomVictim_Fires(t *testing.T) {
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9100, Name: "Range"}
	handler.CharToRoom(ch, room)
	bow, arrow := makeFireFixture(ch)

	// Add a dummy NPC target in the same room.
	target := &types.CharData{Name: "orc", ShortDescr: "an orc", Level: 5,
		Position: types.POS_STANDING, Hit: 20, MaxHit: 20, Armor: 100}
	target.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(target, room)

	// Stub archeryRollD20 to 1 — guaranteed miss to avoid Damage path
	// requiring full world state. We only care that emit happens.
	saved := archeryRollD20
	t.Cleanup(func() { archeryRollD20 = saved })
	archeryRollD20 = func() int { return 1 }

	_ = rangedAttack(ch, "orc", bow, arrow, types.TYPE_HIT+types.DAM_ARROW, 3)
	out := readOutput(ch, client)
	// Expect a "fire" act-to-char line.
	if !strings.Contains(out, "fire") && !strings.Contains(out, "Fire") {
		t.Errorf("expected fire message, got: %q", out)
	}
}

// --- ScanForVictim ---

func TestScanForVictim_Blind(t *testing.T) {
	_ = setupArcheryWorld()
	ch, _ := makeTestChar("Archer")
	here := &types.RoomIndexData{Vnum: 9200, Name: "Here"}
	there := &types.RoomIndexData{Vnum: 9201, Name: "There"}
	exit := &types.ExitData{Direction: 0, Vnum: 9201, ToRoom: there}
	here.Exits = append(here.Exits, exit)
	handler.CharToRoom(ch, here)
	ch.AffectedBy.Set(types.AFF_BLIND)

	got := scanForVictim(ch, exit, "target")
	if got != nil {
		t.Errorf("scanForVictim on blind ch = %v, want nil", got)
	}
}

func TestScanForVictim_ClosedExit(t *testing.T) {
	_ = setupArcheryWorld()
	ch, _ := makeTestChar("Archer")
	here := &types.RoomIndexData{Vnum: 9200, Name: "Here"}
	there := &types.RoomIndexData{Vnum: 9201, Name: "There"}
	exit := &types.ExitData{Direction: 0, Vnum: 9201, ToRoom: there,
		ExitInfo: int(types.EX_CLOSED)}
	here.Exits = append(here.Exits, exit)
	handler.CharToRoom(ch, here)

	got := scanForVictim(ch, exit, "target")
	if got != nil {
		t.Errorf("scanForVictim through closed exit = %v, want nil", got)
	}
	// Ch must be back where they started.
	if ch.InRoom != here {
		t.Errorf("ch.InRoom = %v, want %v (scan must restore)", ch.InRoom, here)
	}
}

func TestScanForVictim_FindsTarget(t *testing.T) {
	_ = setupArcheryWorld()
	ch, _ := makeTestChar("Archer")
	here := &types.RoomIndexData{Vnum: 9200, Name: "Here", SectorType: types.SECT_FIELD}
	there := &types.RoomIndexData{Vnum: 9201, Name: "There", SectorType: types.SECT_FIELD}
	exit := &types.ExitData{Direction: 0, Vnum: 9201, ToRoom: there}
	here.Exits = append(here.Exits, exit)
	handler.CharToRoom(ch, here)

	target := &types.CharData{Name: "orc", ShortDescr: "an orc", Level: 10}
	target.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(target, there)

	got := scanForVictim(ch, exit, "orc")
	if got != target {
		t.Errorf("scanForVictim = %v, want %v", got, target)
	}
	if ch.InRoom != here {
		t.Errorf("ch.InRoom = %v, want %v (must restore)", ch.InRoom, here)
	}
}

func TestScanForVictim_TargetBeyondRange(t *testing.T) {
	// Level-10 ch has max_dist = 8-3 = 5 (below 50, 40, 30). Sector
	// INSIDE adds +1 per room. Chain 10 rooms deep; target should not
	// be found within 5.
	_ = setupArcheryWorld()
	ch, _ := makeTestChar("Shortbow")
	ch.Level = 10
	rooms := make([]*types.RoomIndexData, 12)
	for i := range rooms {
		rooms[i] = &types.RoomIndexData{Vnum: 9300 + i, Name: "Corridor",
			SectorType: types.SECT_INSIDE}
	}
	for i := 0; i < len(rooms)-1; i++ {
		exit := &types.ExitData{Direction: 0, Vnum: rooms[i+1].Vnum, ToRoom: rooms[i+1]}
		rooms[i].Exits = append(rooms[i].Exits, exit)
	}
	handler.CharToRoom(ch, rooms[0])

	target := &types.CharData{Name: "distant orc", ShortDescr: "an orc", Level: 10}
	target.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(target, rooms[10])

	got := scanForVictim(ch, rooms[0].Exits[0], "orc")
	if got == target {
		t.Errorf("scanForVictim should not find target beyond max_dist, but got %v", got)
	}
	if ch.InRoom != rooms[0] {
		t.Errorf("ch.InRoom = %v, want rooms[0] (must restore)", ch.InRoom)
	}
}

// --- ProjectileHit ---

func makeHitFixture(t *testing.T) (*types.CharData, *types.CharData, *types.ObjData, *types.ObjData, *types.RoomIndexData) {
	t.Helper()
	w := setupArcheryWorld()
	ch, _ := makeTestChar("Archer")
	ch.Hitroll = 99 // guarantee hit unless roll is pinned to 0
	ch.Damroll = 0
	room := &types.RoomIndexData{Vnum: 9400, Name: "Range"}
	w.Rooms[9400] = room
	handler.CharToRoom(ch, room)

	target := &types.CharData{
		Name: "orc", ShortDescr: "an orc", Level: 5,
		Position: types.POS_STANDING, Hit: 100, MaxHit: 100, Armor: 100,
	}
	target.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(target, room)
	w.AddChar(target)

	bow, arrow := makeFireFixture(ch)
	return ch, target, bow, arrow, room
}

func TestProjectileHit_NilProjectile(t *testing.T) {
	ch, target, bow, _, _ := makeHitFixture(t)
	got := projectileHit(ch, target, bow, nil, 0, types.TYPE_HIT+types.DAM_ARROW)
	if got != rcNONE {
		t.Errorf("projectileHit(nil projectile) = %d, want rcNONE", got)
	}
}

func TestProjectileHit_DeadVictim(t *testing.T) {
	ch, target, bow, arrow, _ := makeHitFixture(t)
	target.Position = types.POS_DEAD
	target.Hit = 0

	saved := archeryRollD20
	t.Cleanup(func() { archeryRollD20 = saved })
	archeryRollD20 = func() int { return 10 }

	got := projectileHit(ch, target, bow, arrow, 0, types.TYPE_HIT+types.DAM_ARROW)
	if got != rcVICT_DIED {
		t.Errorf("projectileHit(dead victim) = %d, want rcVICT_DIED", got)
	}
}

func TestProjectileHit_MissOnNaturalZero(t *testing.T) {
	ch, target, bow, arrow, _ := makeHitFixture(t)
	saved := archeryRollD20
	t.Cleanup(func() { archeryRollD20 = saved })
	archeryRollD20 = func() int { return 0 } // always miss

	startHit := target.Hit
	_ = projectileHit(ch, target, bow, arrow, 0, types.TYPE_HIT+types.DAM_ARROW)

	// Target HP unchanged (miss does 0 dam).
	if target.Hit != startHit {
		t.Errorf("target.Hit changed on miss: %d → %d", startHit, target.Hit)
	}
}

func TestProjectileHit_HitLodgesProjectile(t *testing.T) {
	// With hitroll=99 and roll=19, hit is guaranteed. The arrow should
	// end up lodged in the victim's WEAR_LODGE_* slot (hit-zone depends
	// on the uncontrolled pchance; any of 3 is OK). ITEM_LODGED set.
	ch, target, bow, arrow, _ := makeHitFixture(t)
	ch.Damroll = 5
	saved := archeryRollD20
	t.Cleanup(func() { archeryRollD20 = saved })
	archeryRollD20 = func() int { return 19 } // always hit (natural 20 equiv)

	_ = projectileHit(ch, target, bow, arrow, 0, types.TYPE_HIT+types.DAM_ARROW)

	if arrow.CarriedBy != target {
		t.Errorf("arrow.CarriedBy = %v, want target (lodged in victim)", arrow.CarriedBy)
	}
	if !arrow.ExtraFlags.IsSet(types.ITEM_LODGED) {
		t.Errorf("ITEM_LODGED not set on arrow")
	}
	// Exactly one of the three WEAR_LODGE_* slots should be occupied.
	validSlots := map[int]bool{
		types.WEAR_LODGE_ARM: true,
		types.WEAR_LODGE_LEG: true,
		types.WEAR_LODGE_RIB: true,
	}
	if !validSlots[arrow.WearLoc] {
		t.Errorf("arrow.WearLoc = %d, want one of WEAR_LODGE_ARM/LEG/RIB", arrow.WearLoc)
	}
}

func TestProjectileHit_StoneExtracts(t *testing.T) {
	// PROJ_STONE projectiles are extracted on damage, never lodge.
	ch, target, bow, arrow, _ := makeHitFixture(t)
	ch.Damroll = 5
	arrow.Value[5] = types.PROJ_STONE
	saved := archeryRollD20
	t.Cleanup(func() { archeryRollD20 = saved })
	archeryRollD20 = func() int { return 19 }

	_ = projectileHit(ch, target, bow, arrow, 0, types.TYPE_HIT+types.DAM_ARROW)

	if arrow.ExtraFlags.IsSet(types.ITEM_LODGED) {
		t.Errorf("PROJ_STONE incorrectly lodged")
	}
	if arrow.WearLoc == types.WEAR_LODGE_ARM ||
		arrow.WearLoc == types.WEAR_LODGE_LEG ||
		arrow.WearLoc == types.WEAR_LODGE_RIB {
		t.Errorf("PROJ_STONE wear-slot = %d (should not be lodge)", arrow.WearLoc)
	}
}

// --- dirName helper ---

func TestDirName_KnownDirections(t *testing.T) {
	cases := []struct {
		dir  int
		want string
	}{
		{0, "north"}, {1, "east"}, {2, "south"}, {3, "west"},
		{4, "up"}, {5, "down"},
		{-1, "somewhere"}, {99, "somewhere"},
	}
	for _, tc := range cases {
		if got := dirName(tc.dir); got != tc.want {
			t.Errorf("dirName(%d) = %q, want %q", tc.dir, got, tc.want)
		}
	}
}

// Keep TestDoDislodge_RibPriority definition as-is.
func TestDoDislodge_RibPriority(t *testing.T) {
	// With arrows lodged in both rib and arm, the rib branch wins (first
	// non-nil in the scan order).
	_ = setupArcheryWorld()
	ch, client := makeTestChar("Archer")
	defer client.Close()
	room := &types.RoomIndexData{Vnum: 9001, Name: "Range"}
	handler.CharToRoom(ch, room)
	ch.Hit = 100

	rib := makeLodgedArrow(ch, types.WEAR_LODGE_RIB, 1, 1)
	arm := makeLodgedArrow(ch, types.WEAR_LODGE_ARM, 1, 1)

	DoDislodge(ch, "arrow")
	out := readOutput(ch, client)
	if !strings.Contains(out, "chest") {
		t.Errorf("rib should win: expected 'chest' in output, got: %q", out)
	}
	if rib.WearLoc != types.WEAR_NONE {
		t.Errorf("rib arrow not unequipped")
	}
	if arm.WearLoc == types.WEAR_NONE {
		t.Errorf("arm arrow incorrectly unequipped (rib should win priority)")
	}
}
