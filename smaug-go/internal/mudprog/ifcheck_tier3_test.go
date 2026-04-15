package mudprog

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// ---------- Priority A: object / mob lookup counts ----------

func makeMob(vnum int) *types.CharData {
	m := &types.CharData{Name: "mob"}
	m.Act.Set(types.ACT_IS_NPC)
	m.IndexData = &types.MobIndexData{Vnum: vnum}
	return m
}

func makeObj(vnum int, itype int) *types.ObjData {
	o := &types.ObjData{
		ItemType: itype,
		Count:    1,
		WearLoc:  types.WEAR_NONE,
		IndexData: &types.ObjIndexData{
			Vnum:     vnum,
			ItemType: itype,
		},
	}
	return o
}

func TestDoIfCheck_MobInRoom(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 100}
	mob := makeMob(200)
	mob.InRoom = room
	m1 := makeMob(300)
	m1.InRoom = room
	m2 := makeMob(300)
	m2.InRoom = room
	room.People = []*types.CharData{mob, m1, m2}

	if !DoIfCheck("mobinroom(300) == 2", mob, nil, nil, nil, nil) {
		t.Error("expected 2 mobs vnum 300 in room")
	}
	if DoIfCheck("mobinroom(400) > 0", mob, nil, nil, nil, nil) {
		t.Error("expected 0 mobs vnum 400 in room")
	}
	if !DoIfCheck("mobinroom(400) == 0", mob, nil, nil, nil, nil) {
		t.Error("expected 0 mobs vnum 400 in room with ==")
	}
	// default op
	if !DoIfCheck("mobinroom(400)", mob, nil, nil, nil, nil) {
		t.Error("default op should match 0")
	}
}

func TestDoIfCheck_MobInArea(t *testing.T) {
	oldWorld := WorldRef
	defer func() { WorldRef = oldWorld }()
	w := world.New("/tmp")
	WorldRef = w

	area := &types.AreaData{Name: "A"}
	room1 := &types.RoomIndexData{Vnum: 100, Area: area}
	room2 := &types.RoomIndexData{Vnum: 101, Area: area}
	otherArea := &types.AreaData{Name: "B"}
	room3 := &types.RoomIndexData{Vnum: 200, Area: otherArea}

	mob := makeMob(500)
	mob.InRoom = room1

	m1 := makeMob(300)
	m1.InRoom = room1
	m2 := makeMob(300)
	m2.InRoom = room2
	m3 := makeMob(300)
	m3.InRoom = room3

	w.Characters = []*types.CharData{mob, m1, m2, m3}

	if !DoIfCheck("mobinarea(300) == 2", mob, nil, nil, nil, nil) {
		t.Error("expected 2 mobs vnum 300 in area")
	}
	if DoIfCheck("mobinarea(300) == 3", mob, nil, nil, nil, nil) {
		t.Error("should not match 3")
	}
}

func TestDoIfCheck_MobInWorld(t *testing.T) {
	oldWorld := WorldRef
	defer func() { WorldRef = oldWorld }()
	w := world.New("/tmp")
	WorldRef = w

	m1 := makeMob(300)
	m2 := makeMob(300)
	m3 := makeMob(400)
	w.Characters = []*types.CharData{m1, m2, m3}
	mob := makeMob(100)
	w.Characters = append(w.Characters, mob)

	if !DoIfCheck("mobinworld(300) == 2", mob, nil, nil, nil, nil) {
		t.Error("expected 2 mobs vnum 300 in world")
	}
	if !DoIfCheck("mobinworld(400) == 1", mob, nil, nil, nil, nil) {
		t.Error("expected 1 mob vnum 400 in world")
	}
	if !DoIfCheck("mobinworld(999) == 0", mob, nil, nil, nil, nil) {
		t.Error("expected 0 mobs vnum 999 in world")
	}
}

func TestDoIfCheck_ObjInWorld(t *testing.T) {
	oldWorld := WorldRef
	defer func() { WorldRef = oldWorld }()
	w := world.New("/tmp")
	WorldRef = w

	o1 := makeObj(500, types.ITEM_WEAPON)
	o2 := makeObj(500, types.ITEM_WEAPON)
	o3 := makeObj(600, types.ITEM_ARMOR)
	w.Objects = []*types.ObjData{o1, o2, o3}
	mob := makeMob(100)

	if !DoIfCheck("objinworld(500) == 2", mob, nil, nil, nil, nil) {
		t.Error("expected 2 obj vnum 500")
	}
	if !DoIfCheck("objinworld(600) == 1", mob, nil, nil, nil, nil) {
		t.Error("expected 1 obj vnum 600")
	}
	if !DoIfCheck("objinworld(999) == 0", mob, nil, nil, nil, nil) {
		t.Error("expected 0 for unknown vnum")
	}
}

func TestDoIfCheck_OvnumHere(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 100}
	mob := makeMob(1)
	mob.InRoom = room
	carried := makeObj(500, types.ITEM_LIGHT)
	carried.CarriedBy = mob
	mob.Carrying = []*types.ObjData{carried}
	onFloor := makeObj(500, types.ITEM_LIGHT)
	room.Contents = []*types.ObjData{onFloor}

	if !DoIfCheck("ovnumhere(500) == 2", mob, nil, nil, nil, nil) {
		t.Error("expected 2 objs vnum 500 here")
	}
	if !DoIfCheck("ovnumhere(999) == 0", mob, nil, nil, nil, nil) {
		t.Error("expected 0 for missing vnum")
	}
}

func TestDoIfCheck_OtypeHere(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 100}
	mob := makeMob(1)
	mob.InRoom = room
	mob.Carrying = []*types.ObjData{makeObj(500, types.ITEM_WEAPON)}
	room.Contents = []*types.ObjData{makeObj(600, types.ITEM_WEAPON), makeObj(601, types.ITEM_ARMOR)}

	if !DoIfCheck("otypehere(5) == 2", mob, nil, nil, nil, nil) {
		t.Error("expected 2 ITEM_WEAPON here")
	}
	if !DoIfCheck("otypehere(9) == 1", mob, nil, nil, nil, nil) {
		t.Error("expected 1 ITEM_ARMOR here")
	}
}

func TestDoIfCheck_OvnumOtypeRoom(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 100}
	mob := makeMob(1)
	mob.InRoom = room
	room.Contents = []*types.ObjData{makeObj(500, types.ITEM_WEAPON), makeObj(500, types.ITEM_WEAPON)}

	if !DoIfCheck("ovnumroom(500) == 2", mob, nil, nil, nil, nil) {
		t.Error("expected 2 vnum 500 in room")
	}
	if !DoIfCheck("otyperoom(5) == 2", mob, nil, nil, nil, nil) {
		t.Error("expected 2 weapons in room")
	}
}

func TestDoIfCheck_OvnumOtypeCarry(t *testing.T) {
	mob := makeMob(1)
	mob.Carrying = []*types.ObjData{
		makeObj(500, types.ITEM_WEAPON),
		makeObj(500, types.ITEM_WEAPON),
		makeObj(600, types.ITEM_ARMOR),
	}

	if !DoIfCheck("ovnumcarry(500) == 2", mob, nil, nil, nil, nil) {
		t.Error("expected 2 vnum 500 carried")
	}
	if !DoIfCheck("otypecarry(9) == 1", mob, nil, nil, nil, nil) {
		t.Error("expected 1 armor carried")
	}
}

func TestDoIfCheck_OvnumOtypeWear(t *testing.T) {
	mob := makeMob(1)
	worn := makeObj(500, types.ITEM_ARMOR)
	worn.WearLoc = types.WEAR_BODY
	inv := makeObj(500, types.ITEM_ARMOR)
	inv.WearLoc = types.WEAR_NONE
	mob.Carrying = []*types.ObjData{worn, inv}

	if !DoIfCheck("ovnumwear(500) == 1", mob, nil, nil, nil, nil) {
		t.Error("expected 1 worn vnum 500")
	}
	if !DoIfCheck("otypewear(9) == 1", mob, nil, nil, nil, nil) {
		t.Error("expected 1 worn armor")
	}
}

func TestDoIfCheck_OvnumOtypeInv(t *testing.T) {
	mob := makeMob(1)
	worn := makeObj(500, types.ITEM_ARMOR)
	worn.WearLoc = types.WEAR_BODY
	inv := makeObj(500, types.ITEM_ARMOR)
	inv.WearLoc = types.WEAR_NONE
	mob.Carrying = []*types.ObjData{worn, inv}

	if !DoIfCheck("ovnuminv(500) == 1", mob, nil, nil, nil, nil) {
		t.Error("expected 1 inv vnum 500")
	}
	if !DoIfCheck("otypeinv(9) == 1", mob, nil, nil, nil, nil) {
		t.Error("expected 1 inv armor")
	}
}

func TestDoIfCheck_ObjValN(t *testing.T) {
	obj := makeObj(500, types.ITEM_WEAPON)
	obj.Value[0] = 10
	obj.Value[1] = 20
	obj.Value[2] = 30
	obj.Value[3] = 40
	obj.Value[4] = 50
	obj.Value[5] = 60

	tests := []struct {
		check    string
		expected bool
	}{
		{"objval0($o) == 10", true},
		{"objval1($o) == 20", true},
		{"objval2($o) == 30", true},
		{"objval3($o) == 40", true},
		{"objval4($o) == 50", true},
		{"objval5($o) == 60", true},
		{"objval0($o) == 99", false},
		{"objval3($o) > 30", true},
	}
	for _, tc := range tests {
		got := DoIfCheck(tc.check, nil, nil, obj, nil, nil)
		if got != tc.expected {
			t.Errorf("%q = %v, want %v", tc.check, got, tc.expected)
		}
	}

	// Missing obj
	if DoIfCheck("objval0($o) == 10", nil, nil, nil, nil, nil) {
		t.Error("missing obj should return false")
	}
}

func TestDoIfCheck_Wearing(t *testing.T) {
	ch := &types.CharData{Name: "Test"}
	worn := makeObj(500, types.ITEM_ARMOR)
	worn.WearLoc = types.WEAR_BODY
	worn.CarriedBy = ch
	ch.Carrying = []*types.ObjData{worn}

	if !DoIfCheck("wearing($n) == body", nil, ch, nil, nil, nil) {
		t.Error("expected wearing body to match")
	}
	if DoIfCheck("wearing($n) == head", nil, ch, nil, nil, nil) {
		t.Error("expected wearing head to not match")
	}
}

func TestDoIfCheck_WearingVnum(t *testing.T) {
	ch := &types.CharData{Name: "Test"}
	worn := makeObj(500, types.ITEM_ARMOR)
	worn.WearLoc = types.WEAR_BODY
	worn.CarriedBy = ch
	ch.Carrying = []*types.ObjData{worn}

	if !DoIfCheck("wearingvnum($n) == 500", nil, ch, nil, nil, nil) {
		t.Error("expected wearingvnum 500 to match")
	}
	if DoIfCheck("wearingvnum($n) == 999", nil, ch, nil, nil, nil) {
		t.Error("expected wearingvnum 999 to not match")
	}
	if DoIfCheck("wearingvnum($n) == abc", nil, ch, nil, nil, nil) {
		t.Error("non-numeric rval should return false")
	}
}

func TestDoIfCheck_CarryingVnum(t *testing.T) {
	ch := &types.CharData{Name: "Test"}
	carried := makeObj(500, types.ITEM_ARMOR)
	carried.CarriedBy = ch
	ch.Carrying = []*types.ObjData{carried}

	if !DoIfCheck("carryingvnum($n) == 500", nil, ch, nil, nil, nil) {
		t.Error("expected carryingvnum 500 to match")
	}
	if DoIfCheck("carryingvnum($n) == 999", nil, ch, nil, nil, nil) {
		t.Error("expected carryingvnum 999 to not match")
	}
}

// ---------- Priority B: char state ----------

func TestDoIfCheck_CanSee(t *testing.T) {
	mob := &types.CharData{Name: "Watcher", Level: 50}
	visible := &types.CharData{Name: "Alice"}
	invis := &types.CharData{Name: "Bob"}
	invis.AffectedBy.Set(types.AFF_INVISIBLE)

	if !DoIfCheck("cansee($n)", mob, visible, nil, nil, nil) {
		t.Error("should see visible char")
	}
	// Can-see details with AFF invis depend on viewer; at min we want the call not to crash.
	_ = DoIfCheck("cansee($n)", mob, invis, nil, nil, nil)
}

func TestDoIfCheck_IsPacifist(t *testing.T) {
	npc := makeMob(1)
	npc.Act.Set(types.ACT_PACIFIST)
	other := makeMob(2)

	if !DoIfCheck("ispacifist($n)", nil, npc, nil, nil, nil) {
		t.Error("npc with ACT_PACIFIST should be pacifist")
	}
	if DoIfCheck("ispacifist($n)", nil, other, nil, nil, nil) {
		t.Error("npc w/o flag should not be pacifist")
	}
	pc := &types.CharData{Name: "Player"}
	if DoIfCheck("ispacifist($n)", nil, pc, nil, nil, nil) {
		t.Error("pc should not be pacifist")
	}
}

func TestDoIfCheck_IsRiding(t *testing.T) {
	mob := &types.CharData{Name: "Horse"}
	rider := &types.CharData{Name: "Rider", Mount: mob}
	other := &types.CharData{Name: "Other"}

	if !DoIfCheck("isriding($n)", mob, rider, nil, nil, nil) {
		t.Error("rider mounting mob should be isriding")
	}
	if DoIfCheck("isriding($n)", mob, other, nil, nil, nil) {
		t.Error("non-rider should not be isriding")
	}
}

func TestDoIfCheck_IsMounted(t *testing.T) {
	mounted := &types.CharData{Name: "Rider", Position: types.POS_MOUNTED}
	standing := &types.CharData{Name: "Other", Position: types.POS_STANDING}

	if !DoIfCheck("ismounted($n)", nil, mounted, nil, nil, nil) {
		t.Error("POS_MOUNTED should be ismounted")
	}
	if DoIfCheck("ismounted($n)", nil, standing, nil, nil, nil) {
		t.Error("POS_STANDING should not be ismounted")
	}
}

func TestDoIfCheck_IsMorphed(t *testing.T) {
	morphed := &types.CharData{Name: "Shifter", Morph: &types.CharMorph{}}
	normal := &types.CharData{Name: "Normal"}

	if !DoIfCheck("ismorphed($n)", nil, morphed, nil, nil, nil) {
		t.Error("morphed should return true")
	}
	if DoIfCheck("ismorphed($n)", nil, normal, nil, nil, nil) {
		t.Error("normal should return false")
	}
}

func TestDoIfCheck_IsNuisance(t *testing.T) {
	pc := &types.CharData{Name: "P", PCData: &types.PCData{Nuisance: &types.NuisanceData{Flags: 1}}}
	clean := &types.CharData{Name: "Clean", PCData: &types.PCData{}}
	npc := makeMob(1)

	if !DoIfCheck("isnuisance($n)", nil, pc, nil, nil, nil) {
		t.Error("nuisance player should match")
	}
	if DoIfCheck("isnuisance($n)", nil, clean, nil, nil, nil) {
		t.Error("clean player should not match")
	}
	if DoIfCheck("isnuisance($n)", nil, npc, nil, nil, nil) {
		t.Error("npc should not match")
	}
}

func TestDoIfCheck_IsPkill(t *testing.T) {
	pkill := &types.CharData{Name: "P", PCData: &types.PCData{Flags: int(types.PCFLAG_DEADLY)}}
	peaceful := &types.CharData{Name: "PP", PCData: &types.PCData{}}
	npc := makeMob(1)

	if !DoIfCheck("ispkill($n)", nil, pkill, nil, nil, nil) {
		t.Error("pkill player should match")
	}
	if DoIfCheck("ispkill($n)", nil, peaceful, nil, nil, nil) {
		t.Error("non-pkill should not match")
	}
	if DoIfCheck("ispkill($n)", nil, npc, nil, nil, nil) {
		t.Error("npc should not match")
	}
}

func TestDoIfCheck_IsThiefIsAttackerIsKiller(t *testing.T) {
	thief := &types.CharData{Name: "T", PCData: &types.PCData{}}
	thief.Act.Set(types.PLR_THIEF)
	attacker := &types.CharData{Name: "A", PCData: &types.PCData{}}
	attacker.Act.Set(types.PLR_ATTACKER)
	killer := &types.CharData{Name: "K", PCData: &types.PCData{}}
	killer.Act.Set(types.PLR_KILLER)
	clean := &types.CharData{Name: "Clean", PCData: &types.PCData{}}

	if !DoIfCheck("isthief($n)", nil, thief, nil, nil, nil) {
		t.Error("PLR_THIEF should match isthief")
	}
	if DoIfCheck("isthief($n)", nil, clean, nil, nil, nil) {
		t.Error("clean should not match isthief")
	}
	if !DoIfCheck("isattacker($n)", nil, attacker, nil, nil, nil) {
		t.Error("PLR_ATTACKER should match")
	}
	if !DoIfCheck("iskiller($n)", nil, killer, nil, nil, nil) {
		t.Error("PLR_KILLER should match")
	}
}

func TestDoIfCheck_Drunk(t *testing.T) {
	pc := &types.CharData{Name: "P", PCData: &types.PCData{}}
	pc.PCData.Condition[types.COND_DRUNK] = 10
	npc := makeMob(1)

	if !DoIfCheck("drunk($n) == 10", nil, pc, nil, nil, nil) {
		t.Error("drunk 10 should match")
	}
	if !DoIfCheck("drunk($n) > 5", nil, pc, nil, nil, nil) {
		t.Error("drunk > 5 should match")
	}
	if DoIfCheck("drunk($n) > 0", nil, npc, nil, nil, nil) {
		t.Error("npc should not match drunk")
	}
}

func TestDoIfCheck_HostDesc(t *testing.T) {
	pc := &types.CharData{Name: "P", PCData: &types.PCData{}, Desc: &types.DescriptorData{Host: "example.com"}}
	pcNoHost := &types.CharData{Name: "P2", PCData: &types.PCData{}, Desc: &types.DescriptorData{}}
	npc := makeMob(1)

	if !DoIfCheck("hostdesc($n) == example.com", nil, pc, nil, nil, nil) {
		t.Error("hostdesc should match example.com")
	}
	if DoIfCheck("hostdesc($n) == other.com", nil, pc, nil, nil, nil) {
		t.Error("hostdesc should not match other.com")
	}
	if DoIfCheck("hostdesc($n)", nil, npc, nil, nil, nil) {
		t.Error("npc should return false")
	}
	if DoIfCheck("hostdesc($n)", nil, pcNoHost, nil, nil, nil) {
		t.Error("no-host should return false")
	}
}

func TestDoIfCheck_WaitState(t *testing.T) {
	pc := &types.CharData{Name: "P", Wait: 5, PCData: &types.PCData{}}
	pcZero := &types.CharData{Name: "P2", Wait: 0, PCData: &types.PCData{}}
	npc := makeMob(1)
	npc.Wait = 5

	if !DoIfCheck("waitstate($n) == 5", nil, pc, nil, nil, nil) {
		t.Error("wait 5 should match")
	}
	if DoIfCheck("waitstate($n) > 0", nil, pcZero, nil, nil, nil) {
		t.Error("zero wait should not match > 0 (C returns false when wait is 0)")
	}
	if DoIfCheck("waitstate($n) > 0", nil, npc, nil, nil, nil) {
		t.Error("npc should return false (C returns false for npc)")
	}
}

func TestDoIfCheck_Favor(t *testing.T) {
	pc := &types.CharData{Name: "P", PCData: &types.PCData{Favor: 1000}}
	zero := &types.CharData{Name: "Z", PCData: &types.PCData{Favor: 0}}
	npc := makeMob(1)

	if !DoIfCheck("favor($n) == 1000", nil, pc, nil, nil, nil) {
		t.Error("favor 1000 should match")
	}
	if DoIfCheck("favor($n) > 0", nil, zero, nil, nil, nil) {
		t.Error("zero favor returns false (C returns false when favor is 0)")
	}
	if DoIfCheck("favor($n) > 0", nil, npc, nil, nil, nil) {
		t.Error("npc should return false")
	}
}

func TestDoIfCheck_Hps(t *testing.T) {
	ch := &types.CharData{Name: "C", Hit: 75}
	if !DoIfCheck("hps($n) == 75", nil, ch, nil, nil, nil) {
		t.Error("hps 75 should match")
	}
	if !DoIfCheck("hps($n) > 50", nil, ch, nil, nil, nil) {
		t.Error("hps 75 > 50")
	}
}

func TestDoIfCheck_Lck(t *testing.T) {
	ch := &types.CharData{Name: "C", PermLck: 17}
	if !DoIfCheck("lck($n) == 17", nil, ch, nil, nil, nil) {
		t.Error("lck 17 should match")
	}
}

func TestDoIfCheck_NumFighting(t *testing.T) {
	ch := &types.CharData{Name: "C", NumFighting: 3}
	// C does num_fighting - 1
	if !DoIfCheck("numfighting($n) == 2", nil, ch, nil, nil, nil) {
		t.Error("numfighting 3 -> 2")
	}
}

// ---------- Priority C: room / area / exit state ----------

func TestDoIfCheck_InRoom(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 3001}
	ch := &types.CharData{Name: "C", InRoom: room}
	if !DoIfCheck("inroom($n) == 3001", nil, ch, nil, nil, nil) {
		t.Error("inroom 3001 should match")
	}
	if DoIfCheck("inroom($n) == 9999", nil, ch, nil, nil, nil) {
		t.Error("inroom 9999 should not match")
	}
}

func TestDoIfCheck_WasInRoom(t *testing.T) {
	prev := &types.RoomIndexData{Vnum: 2000}
	ch := &types.CharData{Name: "C", WasInRoom: prev}
	if !DoIfCheck("wasinroom($n) == 2000", nil, ch, nil, nil, nil) {
		t.Error("wasinroom 2000 should match")
	}
	chNone := &types.CharData{Name: "C2"}
	if DoIfCheck("wasinroom($n) == 2000", nil, chNone, nil, nil, nil) {
		t.Error("no was_in_room should return false")
	}
}

func TestDoIfCheck_InArea(t *testing.T) {
	area := &types.AreaData{Filename: "midgaard.are"}
	room := &types.RoomIndexData{Vnum: 3001, Area: area}
	ch := &types.CharData{Name: "C", InRoom: room}
	if !DoIfCheck("inarea($n) == midgaard.are", nil, ch, nil, nil, nil) {
		t.Error("inarea midgaard should match")
	}
	if DoIfCheck("inarea($n) == other.are", nil, ch, nil, nil, nil) {
		t.Error("inarea other.are should not match")
	}
}

func TestDoIfCheck_Indoors(t *testing.T) {
	inside := &types.RoomIndexData{Vnum: 1, SectorType: types.SECT_INSIDE}
	outside := &types.RoomIndexData{Vnum: 2, SectorType: types.SECT_FIELD}
	chIn := &types.CharData{Name: "I", InRoom: inside}
	chOut := &types.CharData{Name: "O", InRoom: outside}

	if !DoIfCheck("indoors($n)", nil, chIn, nil, nil, nil) {
		t.Error("inside sector should be indoors")
	}
	if DoIfCheck("indoors($n)", nil, chOut, nil, nil, nil) {
		t.Error("field sector should not be indoors")
	}

	// C IS_OUTSIDE checks ROOM_INDOORS flag, not just SECT_INSIDE.
	flagged := &types.RoomIndexData{Vnum: 3, SectorType: types.SECT_FIELD}
	flagged.RoomFlags.Set(types.ROOM_INDOORS)
	chFlagged := &types.CharData{Name: "F", InRoom: flagged}
	if !DoIfCheck("indoors($n)", nil, chFlagged, nil, nil, nil) {
		t.Error("ROOM_INDOORS flag should make non-INSIDE sector count as indoors")
	}
}

func TestDoIfCheck_RoomFlags(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 1}
	room.RoomFlags.Set(types.ROOM_NO_MAGIC)
	room.RoomFlags.Set(types.ROOM_SAFE)
	room.RoomFlags.Set(types.ROOM_NO_SUMMON)
	room.RoomFlags.Set(types.ROOM_NO_ASTRAL)
	room.RoomFlags.Set(types.ROOM_NOSUPPLICATE)
	room.RoomFlags.Set(types.ROOM_NO_RECALL)
	ch := &types.CharData{Name: "C", InRoom: room}

	if !DoIfCheck("nomagic($n)", nil, ch, nil, nil, nil) {
		t.Error("ROOM_NO_MAGIC should match nomagic")
	}
	if !DoIfCheck("safe($n)", nil, ch, nil, nil, nil) {
		t.Error("ROOM_SAFE should match safe")
	}
	if !DoIfCheck("nosummon($n)", nil, ch, nil, nil, nil) {
		t.Error("ROOM_NO_SUMMON should match nosummon")
	}
	if !DoIfCheck("noastral($n)", nil, ch, nil, nil, nil) {
		t.Error("ROOM_NO_ASTRAL should match noastral")
	}
	if !DoIfCheck("nosupplicate($n)", nil, ch, nil, nil, nil) {
		t.Error("ROOM_NOSUPPLICATE should match nosupplicate")
	}
	if !DoIfCheck("norecall($n)", nil, ch, nil, nil, nil) {
		t.Error("ROOM_NO_RECALL should match norecall")
	}

	// Non-matching room
	plain := &types.RoomIndexData{Vnum: 2}
	chP := &types.CharData{Name: "P", InRoom: plain}
	if DoIfCheck("nomagic($n)", nil, chP, nil, nil, nil) {
		t.Error("plain room should not be nomagic")
	}
}

func TestDoIfCheck_ExitChecks(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 1}
	openExit := &types.ExitData{Direction: types.DIR_NORTH, Keyword: "door"}
	closedExit := &types.ExitData{Direction: types.DIR_SOUTH, Keyword: "gate", ExitInfo: int(types.EX_CLOSED)}
	lockedExit := &types.ExitData{Direction: types.DIR_EAST, Keyword: "portcullis", ExitInfo: int(types.EX_CLOSED | types.EX_LOCKED)}
	room.Exits = []*types.ExitData{openExit, closedExit, lockedExit}
	ch := &types.CharData{Name: "C", InRoom: room}

	// ispassage (find_door -> any direction)
	if !DoIfCheck("ispassage($n) north", nil, ch, nil, nil, nil) {
		t.Error("north should be a passage")
	}
	if DoIfCheck("ispassage($n) up", nil, ch, nil, nil, nil) {
		t.Error("up has no exit")
	}

	// isopen
	if !DoIfCheck("isopen($n) north", nil, ch, nil, nil, nil) {
		t.Error("north should be open")
	}
	if DoIfCheck("isopen($n) south", nil, ch, nil, nil, nil) {
		t.Error("south (closed) should not be isopen")
	}

	// islocked
	if !DoIfCheck("islocked($n) east", nil, ch, nil, nil, nil) {
		t.Error("east should be locked")
	}
	if DoIfCheck("islocked($n) south", nil, ch, nil, nil, nil) {
		t.Error("south (closed but unlocked) should not be islocked")
	}
}

// ---------- Priority D: social / political ----------

func TestDoIfCheck_Clan(t *testing.T) {
	clan := &types.ClanData{Name: "Ravens"}
	pc := &types.CharData{Name: "C", PCData: &types.PCData{Clan: clan}}
	pcNoClan := &types.CharData{Name: "N", PCData: &types.PCData{}}
	npc := makeMob(1)

	if !DoIfCheck("clan($n) == Ravens", nil, pc, nil, nil, nil) {
		t.Error("clan Ravens should match")
	}
	if DoIfCheck("clan($n) == Other", nil, pc, nil, nil, nil) {
		t.Error("clan Other should not match")
	}
	if DoIfCheck("clan($n)", nil, pcNoClan, nil, nil, nil) {
		t.Error("no-clan should return false")
	}
	if DoIfCheck("clan($n)", nil, npc, nil, nil, nil) {
		t.Error("npc should return false")
	}
}

func TestDoIfCheck_IsClanLeader(t *testing.T) {
	oldWorld := WorldRef
	defer func() { WorldRef = oldWorld }()
	w := world.New("/tmp")
	WorldRef = w
	clan := &types.ClanData{Name: "Ravens", Leader: "Alice", Number1: "Bob", Number2: "Carol"}
	w.Clans = []*types.ClanData{clan}

	alice := &types.CharData{Name: "Alice", PCData: &types.PCData{}}
	bob := &types.CharData{Name: "Bob", PCData: &types.PCData{}}
	dave := &types.CharData{Name: "Dave", PCData: &types.PCData{}}

	if !DoIfCheck("isclanleader($n) == Ravens", nil, alice, nil, nil, nil) {
		t.Error("Alice is leader of Ravens")
	}
	if DoIfCheck("isclanleader($n) == Ravens", nil, bob, nil, nil, nil) {
		t.Error("Bob is not leader")
	}
	// isleader includes number1/number2
	if !DoIfCheck("isleader($n) == Ravens", nil, alice, nil, nil, nil) {
		t.Error("Alice is leader")
	}
	if !DoIfCheck("isleader($n) == Ravens", nil, bob, nil, nil, nil) {
		t.Error("Bob is number1 so isleader")
	}
	if DoIfCheck("isleader($n) == Ravens", nil, dave, nil, nil, nil) {
		t.Error("Dave is not any leader")
	}
	// isclan1
	if !DoIfCheck("isclan1($n) == Ravens", nil, bob, nil, nil, nil) {
		t.Error("Bob is number1")
	}
	if DoIfCheck("isclan1($n) == Ravens", nil, alice, nil, nil, nil) {
		t.Error("Alice is not number1")
	}
	// isclan2
	carol := &types.CharData{Name: "Carol", PCData: &types.PCData{}}
	if !DoIfCheck("isclan2($n) == Ravens", nil, carol, nil, nil, nil) {
		t.Error("Carol is number2")
	}
}

func TestDoIfCheck_Council(t *testing.T) {
	council := &types.CouncilData{Name: "Sages"}
	pc := &types.CharData{Name: "C", PCData: &types.PCData{Council: council}}
	pcNo := &types.CharData{Name: "C2", PCData: &types.PCData{}}

	if !DoIfCheck("council($n) == Sages", nil, pc, nil, nil, nil) {
		t.Error("Sages should match council")
	}
	if DoIfCheck("council($n)", nil, pcNo, nil, nil, nil) {
		t.Error("no council should return false")
	}
}

func TestDoIfCheck_Deity(t *testing.T) {
	deity := &types.DeityData{Name: "Thor"}
	pc := &types.CharData{Name: "C", PCData: &types.PCData{Deity: deity}}
	pcNo := &types.CharData{Name: "C2", PCData: &types.PCData{}}

	if !DoIfCheck("deity($n) == Thor", nil, pc, nil, nil, nil) {
		t.Error("Thor should match deity")
	}
	if DoIfCheck("deity($n)", nil, pcNo, nil, nil, nil) {
		t.Error("no deity should return false")
	}
}

func TestDoIfCheck_ClanType(t *testing.T) {
	clan := &types.ClanData{Name: "Ravens", ClanType: 2}
	pc := &types.CharData{Name: "C", PCData: &types.PCData{Clan: clan}}
	pcNo := &types.CharData{Name: "C2", PCData: &types.PCData{}}

	if !DoIfCheck("clantype($n) == 2", nil, pc, nil, nil, nil) {
		t.Error("clantype 2 should match")
	}
	if DoIfCheck("clantype($n) == 2", nil, pcNo, nil, nil, nil) {
		t.Error("no clan should return false")
	}
}

// ---------- Priority E: misc ----------

func TestDoIfCheck_Economy(t *testing.T) {
	area := &types.AreaData{Name: "A", LowEconomy: 5000}
	room := &types.RoomIndexData{Vnum: 100, Area: area}
	mob := makeMob(1)
	mob.InRoom = room

	// HighEconomy == 0 -> bonus is 0, so lhs = 5000
	if !DoIfCheck("economy(0) == 5000", mob, nil, nil, nil, nil) {
		t.Error("economy should be 5000")
	}
}

func TestDoIfCheck_Time(t *testing.T) {
	oldWorld := WorldRef
	defer func() { WorldRef = oldWorld }()
	w := world.New("/tmp")
	w.TimeInfo.Hour = 14
	WorldRef = w

	if !DoIfCheck("time($i) == 14", nil, nil, nil, nil, nil) {
		t.Error("time hour 14 should match")
	}
	if !DoIfCheck("time($i) > 12", nil, nil, nil, nil, nil) {
		t.Error("hour 14 > 12")
	}
}

func TestDoIfCheck_Rank(t *testing.T) {
	pc := &types.CharData{Name: "C", PCData: &types.PCData{Rank: "Captain"}}
	npc := makeMob(1)

	if !DoIfCheck("rank($n) == Captain", nil, pc, nil, nil, nil) {
		t.Error("rank Captain should match")
	}
	if DoIfCheck("rank($n)", nil, npc, nil, nil, nil) {
		t.Error("npc rank should be false")
	}
}

func TestDoIfCheck_MortInRoom(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 1}
	mob := makeMob(1)
	mob.InRoom = room

	alice := &types.CharData{Name: "Alice", Level: 10, InRoom: room}
	bob := &types.CharData{Name: "Bob", Level: 10, InRoom: room}
	imm := &types.CharData{Name: "Imm", Level: types.LEVEL_IMMORTAL, InRoom: room}
	room.People = []*types.CharData{mob, alice, bob, imm}

	if !DoIfCheck("mortinroom(Alice)", mob, nil, nil, nil, nil) {
		t.Error("Alice is mort in room")
	}
	if DoIfCheck("mortinroom(Imm)", mob, nil, nil, nil, nil) {
		t.Error("Imm is not mort")
	}
	if DoIfCheck("mortinroom(Zardoz)", mob, nil, nil, nil, nil) {
		t.Error("Zardoz is not present")
	}
}

func TestDoIfCheck_MortInArea(t *testing.T) {
	oldWorld := WorldRef
	defer func() { WorldRef = oldWorld }()
	w := world.New("/tmp")
	WorldRef = w
	area := &types.AreaData{Name: "A"}
	room1 := &types.RoomIndexData{Vnum: 1, Area: area}
	room2 := &types.RoomIndexData{Vnum: 2, Area: area}
	other := &types.AreaData{Name: "Other"}
	room3 := &types.RoomIndexData{Vnum: 3, Area: other}

	mob := makeMob(1)
	mob.InRoom = room1
	alice := &types.CharData{Name: "Alice", Level: 10, InRoom: room2}
	carl := &types.CharData{Name: "Carl", Level: 10, InRoom: room3}
	w.Characters = []*types.CharData{mob, alice, carl}

	if !DoIfCheck("mortinarea(Alice)", mob, nil, nil, nil, nil) {
		t.Error("Alice in same area")
	}
	if DoIfCheck("mortinarea(Carl)", mob, nil, nil, nil, nil) {
		t.Error("Carl is in different area")
	}
}

func TestDoIfCheck_MortInWorld(t *testing.T) {
	oldWorld := WorldRef
	defer func() { WorldRef = oldWorld }()
	w := world.New("/tmp")
	WorldRef = w
	d := &types.DescriptorData{Connected: types.CON_PLAYING}
	alice := &types.CharData{Name: "Alice", Level: 10, Desc: d}
	d.Character = alice
	w.Descriptors = []*types.DescriptorData{d}

	mob := makeMob(1)
	if !DoIfCheck("mortinworld(Alice)", mob, nil, nil, nil, nil) {
		t.Error("Alice should be in world")
	}
	if DoIfCheck("mortinworld(Zardoz)", mob, nil, nil, nil, nil) {
		t.Error("Zardoz is not online")
	}
}

func TestDoIfCheck_MortCount(t *testing.T) {
	oldWorld := WorldRef
	defer func() { WorldRef = oldWorld }()
	w := world.New("/tmp")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 100}
	w.Rooms[100] = room
	mob := makeMob(1)
	mob.InRoom = room

	alice := &types.CharData{Name: "Alice", Level: 10, InRoom: room}
	bob := &types.CharData{Name: "Bob", Level: 10, InRoom: room}
	imm := &types.CharData{Name: "Imm", Level: types.LEVEL_IMMORTAL, InRoom: room}
	room.People = []*types.CharData{mob, alice, bob, imm}

	// mortcount uses rvnum from cvar; 0 => current room. Count excludes imms and npcs
	if !DoIfCheck("mortcount(0) == 2", mob, nil, nil, nil, nil) {
		t.Error("expected 2 mortals in room")
	}
}

func TestDoIfCheck_MobCount(t *testing.T) {
	oldWorld := WorldRef
	defer func() { WorldRef = oldWorld }()
	w := world.New("/tmp")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 100}
	w.Rooms[100] = room
	mob := makeMob(1)
	mob.InRoom = room
	m2 := makeMob(2)
	m2.InRoom = room
	m3 := makeMob(3)
	m3.InRoom = room
	room.People = []*types.CharData{mob, m2, m3}

	// C starts count at -1, so 3 mobs => count 2
	if !DoIfCheck("mobcount(0) == 2", mob, nil, nil, nil, nil) {
		t.Error("expected mobcount 2 (C starts at -1)")
	}
}

func TestDoIfCheck_CharCount(t *testing.T) {
	oldWorld := WorldRef
	defer func() { WorldRef = oldWorld }()
	w := world.New("/tmp")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 100}
	w.Rooms[100] = room
	mob := makeMob(1)
	mob.InRoom = room
	p1 := &types.CharData{Name: "P1", Level: 10, InRoom: room}
	p2 := &types.CharData{Name: "P2", Level: 10, InRoom: room}
	imm := &types.CharData{Name: "I", Level: types.LEVEL_IMMORTAL, InRoom: room}
	room.People = []*types.CharData{mob, p1, p2, imm}

	// C starts at -1; counts mobs + mortals (excludes imms)
	// mob(1) + p1 + p2 = 3, -1 => 2
	if !DoIfCheck("charcount(0) == 2", mob, nil, nil, nil, nil) {
		t.Error("expected charcount 2")
	}
}

func TestDoIfCheck_Number(t *testing.T) {
	mob := makeMob(1234)
	mob.Gold = 99
	// For $i (chkchar == mob): C returns gold, not vnum.
	if !DoIfCheck("number($i) == 99", mob, nil, nil, nil, nil) {
		t.Error("number for $i should return mob.Gold")
	}
	// Non-self npc: returns vnum.
	other := makeMob(5678)
	if !DoIfCheck("number($n) == 5678", mob, other, nil, nil, nil) {
		t.Error("number of other npc should be its vnum")
	}
	// PC: returns false
	pc := &types.CharData{Name: "P"}
	if DoIfCheck("number($n) == 0", mob, pc, nil, nil, nil) {
		t.Error("pc should not match number (C returns FALSE for pc)")
	}
}

func TestDoIfCheck_IsMobInvis(t *testing.T) {
	npc := makeMob(1)
	npc.Act.Set(types.ACT_MOBINVIS)
	other := makeMob(2)

	if !DoIfCheck("ismobinvis($n)", nil, npc, nil, nil, nil) {
		t.Error("ACT_MOBINVIS should match")
	}
	if DoIfCheck("ismobinvis($n)", nil, other, nil, nil, nil) {
		t.Error("without flag should not match")
	}
}

func TestDoIfCheck_MobInvisLevel(t *testing.T) {
	npc := makeMob(1)
	npc.MobInvis = 25
	if !DoIfCheck("mobinvislevel($n) == 25", nil, npc, nil, nil, nil) {
		t.Error("mobinvis level should match")
	}
	pc := &types.CharData{Name: "P"}
	if DoIfCheck("mobinvislevel($n)", nil, pc, nil, nil, nil) {
		t.Error("pc should return false")
	}
}

func TestDoIfCheck_Weight(t *testing.T) {
	ch := &types.CharData{Name: "C", CarryWeight: 120}
	if !DoIfCheck("weight($n) == 120", nil, ch, nil, nil, nil) {
		t.Error("weight 120 should match")
	}
	if !DoIfCheck("weight($n) > 100", nil, ch, nil, nil, nil) {
		t.Error("weight > 100")
	}
}

// ---------- Round-2 fixes ----------

func TestDoIfCheck_LevelUsesGetTrust(t *testing.T) {
	// C mud_prog.c:1128 uses get_trust(chkchar). A PC with explicit Trust
	// override should evaluate against Trust, not their raw Level.
	ch := &types.CharData{Name: "Test", Level: 50, Trust: 60}
	ch.PCData = &types.PCData{}

	if !DoIfCheck("level($n) >= 55", nil, ch, nil, nil, nil) {
		t.Error("trust 60 >= 55 should be true even though Level=50")
	}
	if !DoIfCheck("level($n) < 61", nil, ch, nil, nil, nil) {
		t.Error("trust 60 < 61 should be true")
	}
	if DoIfCheck("level($n) == 50", nil, ch, nil, nil, nil) {
		t.Error("GetTrust returns 60, not raw Level 50")
	}
}

func TestDoIfCheck_LevelFallsBackToLevel(t *testing.T) {
	// With Trust==0, GetTrust() returns Level (for non-capped NPCs) or Level
	// (for PCs) — so the level ifcheck must still compare against Level.
	ch := &types.CharData{Name: "T", Level: 20, Trust: 0}
	if !DoIfCheck("level($n) == 20", nil, ch, nil, nil, nil) {
		t.Error("no-trust PC at Level 20 should compare as 20")
	}
}

func TestDoIfCheck_ClassByName(t *testing.T) {
	// C mud_prog.c:1144 compares class_table[chkchar->class]->who_name as a
	// STRING. Builders write `if class($n) == mag`.
	oldWorld := WorldRef
	defer func() { WorldRef = oldWorld }()
	w := world.New("/tmp")
	w.Classes = []*types.ClassType{
		{WhoName: "Warrior"},
		{WhoName: "Mag"},
		{WhoName: "Cleric"},
	}
	WorldRef = w

	ch := &types.CharData{Name: "Elara", Class: 1}
	if !DoIfCheck("class($n) == mag", nil, ch, nil, nil, nil) {
		t.Error("Class index 1 WhoName=Mag should match `class($n) == mag`")
	}
	if !DoIfCheck("class($n) == Mag", nil, ch, nil, nil, nil) {
		t.Error("class compare should be case-insensitive")
	}
	if DoIfCheck("class($n) == war", nil, ch, nil, nil, nil) {
		t.Error("Class index 1 should not match `war`")
	}
	// Numeric fallback: legacy numeric progs still work.
	if !DoIfCheck("class($n) == 1", nil, ch, nil, nil, nil) {
		t.Error("numeric class compare should still work as fallback")
	}
}

func TestDoIfCheck_RaceByName(t *testing.T) {
	oldWorld := WorldRef
	defer func() { WorldRef = oldWorld }()
	w := world.New("/tmp")
	w.Races = []*types.RaceData{
		{Name: "Human"},
		{Name: "Elf"},
		{Name: "Dwarf"},
	}
	WorldRef = w

	ch := &types.CharData{Name: "Thrand", Race: 2}
	if !DoIfCheck("race($n) == dwarf", nil, ch, nil, nil, nil) {
		t.Error("Race index 2 Name=Dwarf should match `race($n) == dwarf`")
	}
	if DoIfCheck("race($n) == elf", nil, ch, nil, nil, nil) {
		t.Error("Race index 2 should not match `elf`")
	}
	if !DoIfCheck("race($n) == 2", nil, ch, nil, nil, nil) {
		t.Error("numeric race compare should still work as fallback")
	}
}

func TestDoIfCheck_CanSee_BlindObserver(t *testing.T) {
	// C handler.c:3394 — blind observer cannot see anyone.
	observer := &types.CharData{Name: "Seer"}
	observer.AffectedBy.Set(types.AFF_BLIND)
	target := &types.CharData{Name: "T"}

	if DoIfCheck("cansee($n)", observer, target, nil, nil, nil) {
		t.Error("blind observer should not see target")
	}
}

func TestDoIfCheck_CanSee_DarkRoomNoInfrared(t *testing.T) {
	// C handler.c:3397 — dark room without infrared blocks sight.
	oldWorld := WorldRef
	defer func() { WorldRef = oldWorld }()
	w := world.New("/tmp")
	WorldRef = w

	darkRoom := &types.RoomIndexData{Vnum: 1, SectorType: types.SECT_FOREST}
	darkRoom.RoomFlags.Set(types.ROOM_DARK)
	observer := &types.CharData{Name: "Seer", InRoom: darkRoom}
	target := &types.CharData{Name: "T"}

	if DoIfCheck("cansee($n)", observer, target, nil, nil, nil) {
		t.Error("dark room, no infrared → should not see")
	}
}

func TestDoIfCheck_CanSee_DarkRoomWithInfrared(t *testing.T) {
	oldWorld := WorldRef
	defer func() { WorldRef = oldWorld }()
	w := world.New("/tmp")
	WorldRef = w

	darkRoom := &types.RoomIndexData{Vnum: 1, SectorType: types.SECT_FOREST}
	darkRoom.RoomFlags.Set(types.ROOM_DARK)
	observer := &types.CharData{Name: "Seer", InRoom: darkRoom}
	observer.AffectedBy.Set(types.AFF_INFRARED)
	target := &types.CharData{Name: "T"}

	if !DoIfCheck("cansee($n)", observer, target, nil, nil, nil) {
		t.Error("dark room + infrared → should see")
	}
}

func TestDoIfCheck_CanSee_LitRoom(t *testing.T) {
	oldWorld := WorldRef
	defer func() { WorldRef = oldWorld }()
	w := world.New("/tmp")
	WorldRef = w

	litRoom := &types.RoomIndexData{Vnum: 1, SectorType: types.SECT_INSIDE}
	observer := &types.CharData{Name: "Seer", InRoom: litRoom}
	target := &types.CharData{Name: "T"}

	if !DoIfCheck("cansee($n)", observer, target, nil, nil, nil) {
		t.Error("lit room, no impairment → should see")
	}
}
