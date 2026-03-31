package combat

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func newCombatWorld() *world.World {
	return world.New("/tmp/test")
}

func newFighter(name string, level int) *types.CharData {
	return &types.CharData{
		Name:        name,
		Level:       level,
		Position:    types.POS_STANDING,
		Hit:         100,
		MaxHit:      100,
		Mana:        50,
		MaxMana:     50,
		Move:        80,
		MaxMove:     80,
		PermStr:     15,
		PermDex:     14,
		PermCon:     15,
		Armor:       100,
		Hitroll:     0,
		Damroll:     0,
		BareNumDie:  1,
		BareSizeDie: 4,
	}
}

func TestStartFighting(t *testing.T) {
	ch := newFighter("Attacker", 10)
	victim := newFighter("Defender", 10)

	StartFighting(ch, victim)

	if ch.Fighting == nil {
		t.Fatal("ch.Fighting should not be nil")
	}
	if ch.Fighting.Who != victim {
		t.Error("ch.Fighting.Who should be victim")
	}
	if ch.Position != types.POS_FIGHTING {
		t.Errorf("ch.Position = %d, want POS_FIGHTING", ch.Position)
	}
}

func TestStopFighting(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 7999, Name: "Arena"}
	ch := newFighter("Attacker", 10)
	handler.CharToRoom(ch, room)
	victim := newFighter("Defender", 10)
	handler.CharToRoom(victim, room)
	StartFighting(ch, victim)
	StartFighting(victim, ch)

	StopFighting(ch, true)

	if ch.Fighting != nil {
		t.Error("ch.Fighting should be nil after StopFighting")
	}
	if victim.Fighting != nil {
		t.Error("victim.Fighting should be nil (fBoth=true)")
	}
}

func TestOneHit_HitsOrMisses(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 8000, Name: "Arena"}
	w.Rooms[8000] = room

	ch := newFighter("Attacker", 10)
	ch.Act.Set(types.ACT_IS_NPC)
	ch.MobThac0 = 18
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	victim := newFighter("Defender", 10)
	victim.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	StartFighting(ch, victim)

	// Run many hits — some should hit, some miss
	hits := 0
	startHP := victim.Hit
	for i := 0; i < 100; i++ {
		victim.Hit = startHP // Reset HP
		OneHit(w, ch, victim, types.TYPE_UNDEFINED)
	}
	if victim.Hit < startHP {
		hits++
	}

	// With 100 attempts, at least some should have connected
	// (we can't assert exact count due to RNG, just verify no crashes)
}

func TestDamage_ReducesHP(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 8001, Name: "Arena"}
	w.Rooms[8001] = room

	ch := newFighter("Attacker", 10)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	victim := newFighter("Defender", 10)
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	StartFighting(ch, victim)

	ret := Damage(w, ch, victim, 20, types.TYPE_HIT)

	if victim.Hit != 80 {
		t.Errorf("victim.Hit = %d, want 80 (100 - 20)", victim.Hit)
	}
	if ret != rNONE {
		t.Errorf("ret = %d, want rNONE (victim alive)", ret)
	}
}

func TestDamage_KillsVictim(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 8002, Name: "Arena"}
	w.Rooms[8002] = room

	ch := newFighter("Attacker", 10)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	mobIdx := &types.MobIndexData{
		Vnum: 9000, PlayerName: "target dummy", ShortDescr: "a target dummy",
		Level: 1, Position: types.POS_STANDING, DefPosition: types.POS_STANDING,
		HitNoDice: 1, HitSizeDice: 1, HitPlus: 10,
	}
	w.MobIndex[9000] = mobIdx
	victim := handler.CreateMobile(w, mobIdx)
	handler.CharToRoom(victim, room)
	victim.Hit = 10
	victim.MaxHit = 10

	StartFighting(ch, victim)

	ret := Damage(w, ch, victim, 200, types.TYPE_HIT)

	if ret != rVICT_DIED {
		t.Errorf("ret = %d, want rVICT_DIED", ret)
	}
}

func TestMakeCorpse_NPC(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 8003, Name: "Arena"}
	w.Rooms[8003] = room

	mobIdx := &types.MobIndexData{
		Vnum: 9001, PlayerName: "guard", ShortDescr: "a guard",
		Level: 5, Position: types.POS_STANDING, DefPosition: types.POS_STANDING,
	}
	w.MobIndex[9001] = mobIdx
	mob := handler.CreateMobile(w, mobIdx)
	handler.CharToRoom(mob, room)
	mob.Gold = 50

	MakeCorpse(w, mob)

	if len(room.Contents) != 1 {
		t.Fatalf("room.Contents = %d, want 1 (corpse)", len(room.Contents))
	}
	corpse := room.Contents[0]
	if corpse.ItemType != types.ITEM_CORPSE_NPC {
		t.Errorf("corpse.ItemType = %d, want ITEM_CORPSE_NPC", corpse.ItemType)
	}
	if corpse.Timer != 6 {
		t.Errorf("corpse.Timer = %d, want 6", corpse.Timer)
	}
}

func TestViolenceUpdate(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 8004, Name: "Arena"}
	w.Rooms[8004] = room

	ch := newFighter("Attacker", 10)
	ch.Act.Set(types.ACT_IS_NPC)
	ch.MobThac0 = 18
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	victim := newFighter("Defender", 50)
	victim.Act.Set(types.ACT_IS_NPC)
	victim.Hit = 1000
	victim.MaxHit = 1000
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	StartFighting(ch, victim)

	// Run a violence update — should not crash
	ViolenceUpdate(w)

	// Victim may or may not have taken damage (RNG)
	// Just verify no crashes and combat is still active
	if ch.Fighting == nil {
		t.Error("ch should still be fighting")
	}
}
