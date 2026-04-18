package mudprog

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// ---- timeskilled (pMob->killed) ----

func TestIfCheck_TimeSkilled_ChkcharResolvesIndexKilled(t *testing.T) {
	// NPC with IndexData.Killed = 5 → timeskilled($n) == 5 is true.
	mob := makeMob(100)
	mob.IndexData.Killed = 5
	if !DoIfCheck("timeskilled($n) == 5", mob, mob, nil, nil, nil) {
		t.Error("timeskilled($n)==5 should be true when IndexData.Killed=5")
	}
	if DoIfCheck("timeskilled($n) == 4", mob, mob, nil, nil, nil) {
		t.Error("timeskilled($n)==4 should be false when IndexData.Killed=5")
	}
}

func TestIfCheck_TimeSkilled_CVarVnumFallback(t *testing.T) {
	// C path: if chkchar is nil (or a PC), resolve argStr as a vnum and
	// read get_mob_index(vnum)->killed. Test the vnum-fallback via $i
	// pointing at a mob with known IndexData.
	oldWorld := WorldRef
	defer func() { WorldRef = oldWorld }()
	w := world.New("/tmp")
	WorldRef = w

	idx := &types.MobIndexData{Vnum: 3001, Killed: 3}
	w.MobIndex[3001] = idx

	// No chkchar in the normal sense; we exercise the vnum-fallback by
	// passing `atoi(cvar)` = 3001 as the argStr. The easiest way is to
	// pass `$n` with an NPC whose IndexData matches.
	mob := makeMob(3001)
	mob.IndexData.Killed = 3
	if !DoIfCheck("timeskilled($n) >= 3", mob, mob, nil, nil, nil) {
		t.Error("timeskilled($n)>=3 should be true")
	}
}

func TestIfCheck_TimeSkilled_PCTargetReturnsFalseNotPanic(t *testing.T) {
	// chkchar is a PC (no IndexData). Must return false without panic.
	pc := &types.CharData{Name: "Player"}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("timeskilled on PC panicked: %v", r)
		}
	}()
	if DoIfCheck("timeskilled($n) == 0", nil, pc, nil, nil, nil) {
		t.Error("timeskilled on PC target should be false (no IndexData)")
	}
}

// ---- objtype ----

func TestIfCheck_ObjType_Weapon(t *testing.T) {
	obj := makeObj(500, types.ITEM_WEAPON)
	if !DoIfCheck("objtype($o) == 5", nil, nil, obj, nil, nil) {
		t.Errorf("objtype($o)==%d (ITEM_WEAPON) should be true", types.ITEM_WEAPON)
	}
	if DoIfCheck("objtype($o) == 4", nil, nil, obj, nil, nil) {
		t.Error("objtype($o)==4 should be false when type is WEAPON=5")
	}
}

func TestIfCheck_ObjType_NilObjFalse(t *testing.T) {
	// No chkobj → false, no panic.
	if DoIfCheck("objtype($o) == 5", nil, nil, nil, nil, nil) {
		t.Error("objtype on nil obj should be false")
	}
}

// ---- leverpos ----

func TestIfCheck_LeverPos_UpWhenBitSet(t *testing.T) {
	obj := makeObj(600, types.ITEM_LEVER)
	obj.Value[0] = int(types.TRIG_UP)
	if !DoIfCheck("leverpos($o) == up", nil, nil, obj, nil, nil) {
		t.Error("leverpos($o)==up should be true when TRIG_UP bit set")
	}
	if DoIfCheck("leverpos($o) == down", nil, nil, obj, nil, nil) {
		t.Error("leverpos($o)==down should be false when TRIG_UP bit set")
	}
}

func TestIfCheck_LeverPos_DownWhenBitClear(t *testing.T) {
	obj := makeObj(600, types.ITEM_SWITCH)
	obj.Value[0] = 0
	if !DoIfCheck("leverpos($o) == down", nil, nil, obj, nil, nil) {
		t.Error("leverpos($o)==down should be true when TRIG_UP bit clear")
	}
}

func TestIfCheck_LeverPos_WrongItemTypeFalse(t *testing.T) {
	// Port *intent*: only switch/lever/pullchain engage. A weapon
	// returns false even if Value[0] has TRIG_UP.
	obj := makeObj(600, types.ITEM_WEAPON)
	obj.Value[0] = int(types.TRIG_UP)
	if DoIfCheck("leverpos($o) == up", nil, nil, obj, nil, nil) {
		t.Error("leverpos on ITEM_WEAPON should be false regardless of Value[0]")
	}
	if DoIfCheck("leverpos($o) == down", nil, nil, obj, nil, nil) {
		t.Error("leverpos on ITEM_WEAPON should be false for down too")
	}
}

func TestIfCheck_LeverPos_PullChain(t *testing.T) {
	obj := makeObj(600, types.ITEM_PULLCHAIN)
	obj.Value[0] = int(types.TRIG_UP)
	if !DoIfCheck("leverpos($o) == up", nil, nil, obj, nil, nil) {
		t.Error("leverpos($o)==up should work on ITEM_PULLCHAIN")
	}
}

// ---- pkadrenalized ----

func TestIfCheck_PkAdrenalized_Present(t *testing.T) {
	ch := &types.CharData{Name: "Bob"}
	handler.AddTimer(ch, types.TIMER_RECENTFIGHT, 5, "", 0)
	if !DoIfCheck("pkadrenalized($n) > 0", nil, ch, nil, nil, nil) {
		t.Error("pkadrenalized($n)>0 should be true when TIMER_RECENTFIGHT count=5")
	}
	if !DoIfCheck("pkadrenalized($n) == 5", nil, ch, nil, nil, nil) {
		t.Error("pkadrenalized($n)==5 should be true when count=5")
	}
}

func TestIfCheck_PkAdrenalized_Absent(t *testing.T) {
	ch := &types.CharData{Name: "Bob"}
	if DoIfCheck("pkadrenalized($n) > 0", nil, ch, nil, nil, nil) {
		t.Error("pkadrenalized($n)>0 should be false when no timer")
	}
	if !DoIfCheck("pkadrenalized($n) == 0", nil, ch, nil, nil, nil) {
		t.Error("pkadrenalized($n)==0 should be true when no timer")
	}
}

// ---- asupressed ----

func TestIfCheck_ASupressed_Present(t *testing.T) {
	ch := &types.CharData{Name: "Bob"}
	handler.AddTimer(ch, types.TIMER_ASUPRESSED, 3, "", 0)
	if !DoIfCheck("asupressed($n) > 0", nil, ch, nil, nil, nil) {
		t.Error("asupressed($n)>0 should be true when TIMER_ASUPRESSED count=3")
	}
}

func TestIfCheck_ASupressed_Absent(t *testing.T) {
	ch := &types.CharData{Name: "Bob"}
	if DoIfCheck("asupressed($n) > 0", nil, ch, nil, nil, nil) {
		t.Error("asupressed($n)>0 should be false when no timer")
	}
}

// ---- areamulti / multi ----

func makePCWithDesc(name, host string, area *types.AreaData, roomVnum int) *types.CharData {
	room := &types.RoomIndexData{Vnum: roomVnum, Area: area}
	pc := &types.CharData{
		Name: name,
		Desc: &types.DescriptorData{Host: host},
	}
	pc.InRoom = room
	return pc
}

func TestIfCheck_AreaMulti_SameHostSameArea(t *testing.T) {
	oldWorld := WorldRef
	defer func() { WorldRef = oldWorld }()
	w := world.New("/tmp")
	WorldRef = w

	area := &types.AreaData{Name: "TestArea"}
	a := makePCWithDesc("Alice", "10.0.0.1", area, 100)
	b := makePCWithDesc("Bob", "10.0.0.1", area, 100)
	a.InRoom = b.InRoom // same room in same area
	w.Characters = []*types.CharData{a, b}

	if !DoIfCheck("areamulti($n) == 2", nil, a, nil, nil, nil) {
		t.Error("areamulti($n)==2 should be true when two PCs share host in same area")
	}
}

func TestIfCheck_AreaMulti_DifferentAreaExcluded(t *testing.T) {
	oldWorld := WorldRef
	defer func() { WorldRef = oldWorld }()
	w := world.New("/tmp")
	WorldRef = w

	area1 := &types.AreaData{Name: "Area1"}
	area2 := &types.AreaData{Name: "Area2"}
	a := makePCWithDesc("Alice", "10.0.0.1", area1, 100)
	b := makePCWithDesc("Bob", "10.0.0.1", area2, 200)
	w.Characters = []*types.CharData{a, b}

	// Same host but different areas → only 1 (the chkchar itself).
	if !DoIfCheck("areamulti($n) == 1", nil, a, nil, nil, nil) {
		t.Error("areamulti($n)==1 expected (different areas excluded)")
	}
}

func TestIfCheck_AreaMulti_DifferentHostExcluded(t *testing.T) {
	oldWorld := WorldRef
	defer func() { WorldRef = oldWorld }()
	w := world.New("/tmp")
	WorldRef = w

	area := &types.AreaData{Name: "TestArea"}
	a := makePCWithDesc("Alice", "10.0.0.1", area, 100)
	b := makePCWithDesc("Bob", "10.0.0.2", area, 100)
	w.Characters = []*types.CharData{a, b}

	if !DoIfCheck("areamulti($n) == 1", nil, a, nil, nil, nil) {
		t.Error("areamulti($n)==1 expected (different hosts excluded)")
	}
}

func TestIfCheck_AreaMulti_NPCsExcluded(t *testing.T) {
	oldWorld := WorldRef
	defer func() { WorldRef = oldWorld }()
	w := world.New("/tmp")
	WorldRef = w

	area := &types.AreaData{Name: "TestArea"}
	a := makePCWithDesc("Alice", "10.0.0.1", area, 100)
	// NPC with matching "host" (but NPCs have no Desc) — shouldn't count.
	npc := makeMob(500)
	npc.InRoom = &types.RoomIndexData{Vnum: 100, Area: area}
	w.Characters = []*types.CharData{a, npc}

	if !DoIfCheck("areamulti($n) == 1", nil, a, nil, nil, nil) {
		t.Error("areamulti($n)==1 expected (NPC excluded)")
	}
}

func TestIfCheck_Multi_IgnoresArea(t *testing.T) {
	oldWorld := WorldRef
	defer func() { WorldRef = oldWorld }()
	w := world.New("/tmp")
	WorldRef = w

	area1 := &types.AreaData{Name: "Area1"}
	area2 := &types.AreaData{Name: "Area2"}
	a := makePCWithDesc("Alice", "10.0.0.1", area1, 100)
	b := makePCWithDesc("Bob", "10.0.0.1", area2, 200)
	w.Characters = []*types.CharData{a, b}

	// Same host, different areas → multi sees both.
	if !DoIfCheck("multi($n) == 2", nil, a, nil, nil, nil) {
		t.Error("multi($n)==2 expected (areas ignored)")
	}
}

func TestIfCheck_Multi_NPCChkcharReturnsFalse(t *testing.T) {
	// C: `!IS_NPC(chkchar)` guard; an NPC chkchar means the check
	// collapses to 0 (loop never increments).
	oldWorld := WorldRef
	defer func() { WorldRef = oldWorld }()
	w := world.New("/tmp")
	WorldRef = w

	npc := makeMob(500)
	w.Characters = []*types.CharData{npc}
	if !DoIfCheck("multi($n) == 0", nil, npc, nil, nil, nil) {
		t.Error("multi($n)==0 expected for NPC chkchar")
	}
}
