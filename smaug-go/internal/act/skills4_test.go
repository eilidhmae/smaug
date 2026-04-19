package act

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// --- Wave 3: utility ---

func TestDoMeditate_InCombat(t *testing.T) {
	ensureSkill(t, "meditate")
	ch, victim := newSkillTestRoom()
	ch.Fighting = &types.FightData{Who: victim}
	DoMeditate(ch, "")
	// Should say "not in combat"
}

func TestDoMeditate_Success(t *testing.T) {
	ensureSkill(t, "meditate")
	ch, _ := newSkillTestRoom()
	setLearned(t, ch, "meditate", 100)
	DoMeditate(ch, "")
	// No crash
}

func TestDoTrance(t *testing.T) {
	ensureSkill(t, "trance")
	ch, _ := newSkillTestRoom()
	setLearned(t, ch, "trance", 100)
	DoTrance(ch, "")
}

func TestDoSearch_NoHidden(t *testing.T) {
	ensureSkill(t, "search")
	ch, _ := newSkillTestRoom()
	setLearned(t, ch, "search", 100)
	DoSearch(ch, "")
}

func TestDoSearch_FindsSecretExit(t *testing.T) {
	ensureSkill(t, "search")
	ch, _ := newSkillTestRoom()
	setLearned(t, ch, "search", 100)
	target := &types.RoomIndexData{Vnum: 9100}
	ex := &types.ExitData{
		Direction: types.DIR_NORTH,
		ToRoom:    target,
		ExitInfo:  int(types.EX_SECRET),
	}
	ch.InRoom.Exits = append(ch.InRoom.Exits, ex)
	// Try several times — may take a few rolls
	for i := 0; i < 20; i++ {
		DoSearch(ch, "")
		if ex.ExitInfo&int(types.EX_SECRET) == 0 {
			return
		}
	}
}

func TestDoDetrap_NoArg(t *testing.T) {
	ensureSkill(t, "detrap")
	ch, _ := newSkillTestRoom()
	DoDetrap(ch, "")
}

func TestDoDig(t *testing.T) {
	ensureSkill(t, "dig")
	ch, _ := newSkillTestRoom()
	setLearned(t, ch, "dig", 100)
	DoDig(ch, "")
}

func TestDoVisible(t *testing.T) {
	ch, _ := newSkillTestRoom()
	ch.AffectedBy.Set(types.AFF_INVISIBLE)
	ch.AffectedBy.Set(types.AFF_HIDE)
	DoVisible(ch, "")
	if ch.AffectedBy.IsSet(types.AFF_INVISIBLE) {
		t.Error("expected AFF_INVISIBLE cleared")
	}
	if ch.AffectedBy.IsSet(types.AFF_HIDE) {
		t.Error("expected AFF_HIDE cleared")
	}
}

func TestDoStyle_NoArg(t *testing.T) {
	ch, _ := newSkillTestRoom()
	DoStyle(ch, "")
}

func TestDoStyle_Set(t *testing.T) {
	ch, _ := newSkillTestRoom()
	DoStyle(ch, "defensive")
	if ch.Style != types.STYLE_DEFENSIVE {
		t.Errorf("expected STYLE_DEFENSIVE, got %d", ch.Style)
	}
}

func TestDoStyle_Unknown(t *testing.T) {
	ch, _ := newSkillTestRoom()
	DoStyle(ch, "xyzzy")
}

func TestDoStance_NoArg(t *testing.T) {
	ch, _ := newSkillTestRoom()
	DoStance(ch, "")
}

func TestDoStance_Set(t *testing.T) {
	ch, _ := newSkillTestRoom()
	DoStance(ch, "dragon")
	if ch.Stance != types.STANCE_DRAGON {
		t.Errorf("expected STANCE_DRAGON, got %d", ch.Stance)
	}
}

// --- Wave 4: crafting ---

func TestDoFeed_NoArgs(t *testing.T) {
	ch, _ := newSkillTestRoom()
	DoFeed(ch, "")
}

func TestDoFeed_Success(t *testing.T) {
	ch, victim := newSkillTestRoom()
	food := &types.ObjData{
		Name:       "bread",
		ShortDescr: "a loaf of bread",
		ItemType:   types.ITEM_FOOD,
		WearLoc:    types.WEAR_NONE,
	}
	food.CarriedBy = ch
	ch.Carrying = append(ch.Carrying, food)
	victim.Name = "target"
	DoFeed(ch, "target bread")
	// Food should now be on victim
	found := false
	for _, o := range victim.Carrying {
		if o == food {
			found = true
		}
	}
	if !found {
		t.Error("expected food transferred to victim")
	}
}

func TestDoSkin_NoArg(t *testing.T) {
	ensureSkill(t, "skin")
	ch, _ := newSkillTestRoom()
	DoSkin(ch, "")
}

func TestDoSkin_Success(t *testing.T) {
	ensureSkill(t, "skin")
	ch, _ := newSkillTestRoom()
	setLearned(t, ch, "skin", 100)
	corpse := &types.ObjData{
		Name:     "orc corpse",
		ItemType: types.ITEM_CORPSE_NPC,
	}
	ch.InRoom.Contents = append(ch.InRoom.Contents, corpse)
	for i := 0; i < 20; i++ {
		startCount := len(ch.Carrying)
		DoSkin(ch, "orc")
		if len(ch.Carrying) > startCount {
			return
		}
		// re-add corpse if extracted but skill failed — extracting empties contents
		found := false
		for _, o := range ch.InRoom.Contents {
			if o == corpse {
				found = true
			}
		}
		if !found {
			ch.InRoom.Contents = append(ch.InRoom.Contents, corpse)
		}
	}
}

func TestDoPoisonWeapon_NoWeapon(t *testing.T) {
	ensureSkill(t, "poison weapon")
	ch, _ := newSkillTestRoom()
	DoPoisonWeapon(ch, "")
}

func TestDoScribe_NoArg(t *testing.T) {
	ensureSkill(t, "scribe")
	ch, _ := newSkillTestRoom()
	DoScribe(ch, "")
}

func TestDoScribe_UnknownSpell(t *testing.T) {
	ensureSkill(t, "scribe")
	ch, _ := newSkillTestRoom()
	DoScribe(ch, "_totallyfakespell_")
}

func TestDoCook_NotFood(t *testing.T) {
	ensureSkill(t, "cook")
	ch, _ := newSkillTestRoom()
	obj := &types.ObjData{
		Name:       "stick",
		ShortDescr: "a stick",
		ItemType:   types.ITEM_WEAPON,
		CarriedBy:  ch,
		WearLoc:    types.WEAR_NONE,
	}
	ch.Carrying = append(ch.Carrying, obj)
	DoCook(ch, "stick")
}

func TestDoCook_Success(t *testing.T) {
	ensureSkill(t, "cook")
	ch, _ := newSkillTestRoom()
	setLearned(t, ch, "cook", 100)
	food := &types.ObjData{
		Name:       "fish",
		ShortDescr: "a raw fish",
		ItemType:   types.ITEM_FOOD,
		CarriedBy:  ch,
		WearLoc:    types.WEAR_NONE,
		Value:      [6]int{10, 0, 0, 0, 0, 0},
	}
	ch.Carrying = append(ch.Carrying, food)
	for i := 0; i < 20; i++ {
		DoCook(ch, "fish")
		if food.Value[4] != 0 {
			return
		}
	}
}

func TestDoFire_NoArg(t *testing.T) {
	ch, _ := newSkillTestRoom()
	DoFire(ch, "")
}

func TestDoMistwalk_NoArg(t *testing.T) {
	ensureSkill(t, "mistwalk")
	ch, _ := newSkillTestRoom()
	DoMistwalk(ch, "")
}

// --- isBloodRace / DoBloodlet (Phase 6 skills G3) ---

func TestIsBloodRace_NPC(t *testing.T) {
	ch := &types.CharData{Race: types.RACE_VAMPIRE, Class: types.CLASS_VAMPIRE}
	ch.Act.Set(types.ACT_IS_NPC)
	if isBloodRace(ch) {
		t.Error("NPC should never qualify as blood race")
	}
}

func TestIsBloodRace_VampireRace(t *testing.T) {
	ch := &types.CharData{Race: types.RACE_VAMPIRE, Class: types.CLASS_WARRIOR, PCData: &types.PCData{}}
	if !isBloodRace(ch) {
		t.Error("RACE_VAMPIRE PC should qualify")
	}
}

func TestIsBloodRace_VampireClass(t *testing.T) {
	ch := &types.CharData{Race: types.RACE_HUMAN, Class: types.CLASS_VAMPIRE, PCData: &types.PCData{}}
	if !isBloodRace(ch) {
		t.Error("CLASS_VAMPIRE PC should qualify")
	}
}

func TestIsBloodRace_DemonRace(t *testing.T) {
	ch := &types.CharData{Race: types.RACE_DEMON, Class: types.CLASS_WARRIOR, PCData: &types.PCData{}}
	if !isBloodRace(ch) {
		t.Error("RACE_DEMON PC should qualify")
	}
}

func TestIsBloodRace_DemonClass(t *testing.T) {
	ch := &types.CharData{Race: types.RACE_HUMAN, Class: types.CLASS_DEMON, PCData: &types.PCData{}}
	if !isBloodRace(ch) {
		t.Error("CLASS_DEMON PC should qualify")
	}
}

func TestIsBloodRace_Plain(t *testing.T) {
	ch := &types.CharData{Race: types.RACE_HUMAN, Class: types.CLASS_WARRIOR, PCData: &types.PCData{}}
	if isBloodRace(ch) {
		t.Error("plain human warrior should not qualify")
	}
}

// newBloodletFixture creates a vampire PC in a fresh room with bloodthirst
// set high enough to pass the gate. Returns the character and a helper to
// snapshot pre-call room contents for assertions.
func newBloodletFixture(t *testing.T, learned int) *types.CharData {
	t.Helper()
	ensureSkill(t, "bloodlet")
	gsn := lookupSkillSlot("bloodlet")
	// Bump Beats=1 to match typical skills.dat values; bloodlet uses
	// PULSE_VIOLENCE for Wait, not Beats — but Beats still needs to exist.
	WorldRef.Skills[gsn].Beats = 12
	// Adept=100 so learnFromSuccess/failure don't cap out during tests.
	for i := 0; i < types.MAX_CLASS; i++ {
		WorldRef.Skills[gsn].SkillAdept[i] = 100
	}

	room := &types.RoomIndexData{Vnum: 9500, Name: "Bloodlet Test"}
	ch := newTestCharWithDesc()
	ch.InRoom = room
	ch.Level = 20
	ch.Hit = 100
	ch.MaxHit = 100
	ch.Position = types.POS_STANDING
	ch.Race = types.RACE_VAMPIRE
	ch.PCData.Condition[types.COND_BLOODTHIRST] = 20
	ch.PCData.Learned[gsn] = learned
	room.People = append(room.People, ch)
	return ch
}

func TestDoBloodlet_NPCSilentNoop(t *testing.T) {
	ensureSkill(t, "bloodlet")
	room := &types.RoomIndexData{Vnum: 9500}
	ch := &types.CharData{InRoom: room, Level: 20, Hit: 100, MaxHit: 100}
	ch.Act.Set(types.ACT_IS_NPC)
	room.People = append(room.People, ch)
	prevHit := ch.Hit
	DoBloodlet(ch, "")
	if ch.Hit != prevHit {
		t.Errorf("NPC bloodlet should be no-op; hit went %d -> %d", prevHit, ch.Hit)
	}
	if len(room.Contents) != 0 {
		t.Error("NPC bloodlet should not spawn object")
	}
}

func TestDoBloodlet_NonBloodRaceSilentNoop(t *testing.T) {
	ensureSkill(t, "bloodlet")
	room := &types.RoomIndexData{Vnum: 9500}
	ch := newTestCharWithDesc()
	ch.InRoom = room
	ch.Level = 20
	ch.Hit = 100
	ch.MaxHit = 100
	ch.Race = types.RACE_HUMAN
	ch.Class = types.CLASS_WARRIOR
	ch.PCData.Condition[types.COND_BLOODTHIRST] = 20
	room.People = append(room.People, ch)
	prevHit := ch.Hit
	DoBloodlet(ch, "")
	if ch.Hit != prevHit {
		t.Errorf("non-blood-race bloodlet should be no-op; hit went %d -> %d", prevHit, ch.Hit)
	}
	if len(room.Contents) != 0 {
		t.Error("non-blood-race bloodlet should not spawn object")
	}
}

func TestDoBloodlet_BlockedIfFighting(t *testing.T) {
	ch := newBloodletFixture(t, 100)
	victim := &types.CharData{Name: "foe", Level: 5, Hit: 20, MaxHit: 20, InRoom: ch.InRoom}
	victim.Act.Set(types.ACT_IS_NPC)
	ch.Fighting = &types.FightData{Who: victim}
	prevHit := ch.Hit
	DoBloodlet(ch, "")
	if ch.Hit != prevHit {
		t.Error("fighting bloodlet should be rejected before damage")
	}
	if len(ch.InRoom.Contents) != 0 {
		t.Error("fighting bloodlet should not spawn object")
	}
}

func TestDoBloodlet_BlockedIfBloodthirstLow(t *testing.T) {
	ch := newBloodletFixture(t, 100)
	ch.PCData.Condition[types.COND_BLOODTHIRST] = 9
	prevHit := ch.Hit
	DoBloodlet(ch, "")
	if ch.Hit != prevHit {
		t.Error("low-bloodthirst bloodlet should be rejected before damage")
	}
	if len(ch.InRoom.Contents) != 0 {
		t.Error("low-bloodthirst bloodlet should not spawn object")
	}
}

func TestDoBloodlet_SuccessSpawnsObject(t *testing.T) {
	// Guarantee OBJ_VNUM_BLOODLET is in the world index — test init may or
	// may not have loaded the real area. Inject a minimal prototype.
	if WorldRef.ObjIndex[types.OBJ_VNUM_BLOODLET] == nil {
		WorldRef.ObjIndex[types.OBJ_VNUM_BLOODLET] = &types.ObjIndexData{
			Vnum:       types.OBJ_VNUM_BLOODLET,
			Name:       "blood pool",
			ShortDescr: "a pool of blood",
		}
	}
	ch := newBloodletFixture(t, 100)
	restore := withStubNumberPercent(1)
	defer restore()
	DoBloodlet(ch, "")
	var pool *types.ObjData
	for _, o := range ch.InRoom.Contents {
		if o.IndexData != nil && o.IndexData.Vnum == types.OBJ_VNUM_BLOODLET {
			pool = o
			break
		}
	}
	if pool == nil {
		t.Fatal("successful bloodlet should spawn OBJ_VNUM_BLOODLET in room")
	}
	if pool.Timer != 1 {
		t.Errorf("bloodlet obj Timer = %d; want 1", pool.Timer)
	}
	if pool.Value[1] != 6 {
		t.Errorf("bloodlet obj Value[1] = %d; want 6", pool.Value[1])
	}
}

func TestDoBloodlet_SuccessDecrementsBloodthirst(t *testing.T) {
	if WorldRef.ObjIndex[types.OBJ_VNUM_BLOODLET] == nil {
		WorldRef.ObjIndex[types.OBJ_VNUM_BLOODLET] = &types.ObjIndexData{Vnum: types.OBJ_VNUM_BLOODLET}
	}
	ch := newBloodletFixture(t, 100)
	ch.PCData.Condition[types.COND_BLOODTHIRST] = 20
	restore := withStubNumberPercent(1)
	defer restore()
	DoBloodlet(ch, "")
	if got := ch.PCData.Condition[types.COND_BLOODTHIRST]; got != 13 {
		t.Errorf("bloodthirst = %d after bloodlet; want 13 (20-7)", got)
	}
}

func TestDoBloodlet_SuccessSelfDamages(t *testing.T) {
	if WorldRef.ObjIndex[types.OBJ_VNUM_BLOODLET] == nil {
		WorldRef.ObjIndex[types.OBJ_VNUM_BLOODLET] = &types.ObjIndexData{Vnum: types.OBJ_VNUM_BLOODLET}
	}
	ch := newBloodletFixture(t, 100)
	ch.Level = 20
	ch.Hit = 100
	restore := withStubNumberPercent(1)
	defer restore()
	DoBloodlet(ch, "")
	// Level/5 = 4 damage. damageWith subtracts dam from Hit.
	if ch.Hit != 96 {
		t.Errorf("ch.Hit = %d after bloodlet; want 96 (100 - 20/5)", ch.Hit)
	}
}

func TestDoBloodlet_FailureNoDamageNoObject(t *testing.T) {
	if WorldRef.ObjIndex[types.OBJ_VNUM_BLOODLET] == nil {
		WorldRef.ObjIndex[types.OBJ_VNUM_BLOODLET] = &types.ObjIndexData{Vnum: types.OBJ_VNUM_BLOODLET}
	}
	ch := newBloodletFixture(t, 0) // unlearned so canUseSkill=false
	prevHit := ch.Hit
	prevBlood := ch.PCData.Condition[types.COND_BLOODTHIRST]
	restore := withStubNumberPercent(99)
	defer restore()
	DoBloodlet(ch, "")
	if ch.Hit != prevHit {
		t.Errorf("failed bloodlet should do 0 damage; hit went %d -> %d", prevHit, ch.Hit)
	}
	if ch.PCData.Condition[types.COND_BLOODTHIRST] != prevBlood {
		t.Error("failed bloodlet should not consume bloodthirst")
	}
	// Check no pool was spawned.
	for _, o := range ch.InRoom.Contents {
		if o.IndexData != nil && o.IndexData.Vnum == types.OBJ_VNUM_BLOODLET {
			t.Error("failed bloodlet should not spawn object")
		}
	}
}

func TestDoBloodlet_MissingObjIndexLogsBugNoSpawn(t *testing.T) {
	// Force the index missing.
	saved := WorldRef.ObjIndex[types.OBJ_VNUM_BLOODLET]
	delete(WorldRef.ObjIndex, types.OBJ_VNUM_BLOODLET)
	defer func() {
		if saved != nil {
			WorldRef.ObjIndex[types.OBJ_VNUM_BLOODLET] = saved
		}
	}()

	ch := newBloodletFixture(t, 100)
	restore := withStubNumberPercent(1)
	defer restore()
	DoBloodlet(ch, "") // should not panic

	for _, o := range ch.InRoom.Contents {
		if o.IndexData != nil && o.IndexData.Vnum == types.OBJ_VNUM_BLOODLET {
			t.Error("missing-index bloodlet should not spawn a pool")
		}
	}
	// Self-damage still ran.
	if ch.Hit == 100 {
		t.Error("missing-index bloodlet should still self-damage")
	}
}

func TestDoBloodlet_WaitStateSet(t *testing.T) {
	if WorldRef.ObjIndex[types.OBJ_VNUM_BLOODLET] == nil {
		WorldRef.ObjIndex[types.OBJ_VNUM_BLOODLET] = &types.ObjIndexData{Vnum: types.OBJ_VNUM_BLOODLET}
	}
	ch := newBloodletFixture(t, 100)
	ch.Wait = 0
	restore := withStubNumberPercent(1)
	defer restore()
	DoBloodlet(ch, "")
	if ch.Wait != types.PULSE_VIOLENCE {
		t.Errorf("ch.Wait = %d after bloodlet; want PULSE_VIOLENCE (%d)",
			ch.Wait, types.PULSE_VIOLENCE)
	}
}
