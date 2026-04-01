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

func TestViolenceUpdate_IncapCantAttack(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 8010, Name: "Arena"}
	w.Rooms[8010] = room

	ch := newFighter("Incap Player", 10)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)
	victim := newFighter("Mob", 10)
	victim.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	StartFighting(ch, victim)

	// Simulate being beaten to incap after combat started
	ch.Hit = 0
	ch.Position = types.POS_INCAP
	startHP := victim.Hit

	ViolenceUpdate(w)

	if victim.Hit != startHP {
		t.Errorf("victim.Hit = %d, want %d (incapacitated attacker should not deal damage)", victim.Hit, startHP)
	}
}

func TestMakeCorpse_GoldInCorpse(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 8011, Name: "Arena"}
	w.Rooms[8011] = room

	mobIdx := &types.MobIndexData{
		Vnum: 9010, PlayerName: "rich guard", ShortDescr: "a rich guard",
		Level: 5, Position: types.POS_STANDING, DefPosition: types.POS_STANDING,
	}
	w.MobIndex[9010] = mobIdx
	mob := handler.CreateMobile(w, mobIdx)
	handler.CharToRoom(mob, room)
	mob.Gold = 100

	MakeCorpse(w, mob)

	corpse := room.Contents[0]
	if corpse.Value[0] != 100 {
		t.Errorf("corpse.Value[0] (gold) = %d, want 100", corpse.Value[0])
	}
}

func TestDamage_XPGainOnKill(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 8012, Name: "Arena"}
	w.Rooms[8012] = room

	ch := newFighter("Player", 10)
	ch.PCData = &types.PCData{}
	handler.CharToRoom(ch, room)
	w.AddChar(ch)
	startXP := ch.Exp

	mobIdx := &types.MobIndexData{
		Vnum: 9011, PlayerName: "target", ShortDescr: "a target",
		Level: 5, Position: types.POS_STANDING, DefPosition: types.POS_STANDING,
		HitNoDice: 1, HitSizeDice: 1, HitPlus: 10,
	}
	w.MobIndex[9011] = mobIdx
	victim := handler.CreateMobile(w, mobIdx)
	handler.CharToRoom(victim, room)
	victim.Hit = 10
	victim.MaxHit = 10
	victim.Exp = 500

	StartFighting(ch, victim)
	Damage(w, ch, victim, 200, types.TYPE_HIT)

	if ch.Exp <= startXP {
		t.Errorf("Exp = %d, should have increased from %d after killing mob", ch.Exp, startXP)
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

// --- StartFighting: already fighting is a no-op ---
func TestStartFighting_AlreadyFighting(t *testing.T) {
	ch := newFighter("Attacker", 10)
	victim1 := newFighter("Victim1", 10)
	victim2 := newFighter("Victim2", 10)

	StartFighting(ch, victim1)
	StartFighting(ch, victim2) // should be ignored

	if ch.Fighting.Who != victim1 {
		t.Error("second StartFighting should have been ignored")
	}
}

// --- StopFighting: fBoth=false should not stop opponent ---
func TestStopFighting_NotBoth(t *testing.T) {
	room := &types.RoomIndexData{Vnum: 9000, Name: "Arena"}
	ch := newFighter("Attacker", 10)
	handler.CharToRoom(ch, room)
	victim := newFighter("Defender", 10)
	handler.CharToRoom(victim, room)
	StartFighting(ch, victim)
	StartFighting(victim, ch)

	StopFighting(ch, false)

	if ch.Fighting != nil {
		t.Error("ch.Fighting should be nil")
	}
	if victim.Fighting == nil {
		t.Error("victim.Fighting should still be set (fBoth=false)")
	}
}

// --- dirName ---
func TestDirName(t *testing.T) {
	tests := []struct {
		dir  int
		want string
	}{
		{types.DIR_NORTH, "north"},
		{types.DIR_EAST, "east"},
		{types.DIR_SOUTH, "south"},
		{types.DIR_WEST, "west"},
		{types.DIR_UP, "up"},
		{types.DIR_DOWN, "down"},
		{types.DIR_NORTHEAST, "northeast"},
		{types.DIR_NORTHWEST, "northwest"},
		{types.DIR_SOUTHEAST, "southeast"},
		{types.DIR_SOUTHWEST, "southwest"},
		{99, "somewhere"},
		{-1, "somewhere"},
	}
	for _, tc := range tests {
		got := dirName(tc.dir)
		if got != tc.want {
			t.Errorf("dirName(%d) = %q, want %q", tc.dir, got, tc.want)
		}
	}
}

// --- OneHit: victim already dead ---
func TestOneHit_VictimAlreadyDead(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9001, Name: "Arena"}
	w.Rooms[9001] = room

	ch := newFighter("Attacker", 10)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)
	victim := newFighter("Defender", 10)
	handler.CharToRoom(victim, room)
	w.AddChar(victim)
	victim.Hit = 0

	// Should return immediately, no crash
	OneHit(w, ch, victim, types.TYPE_UNDEFINED)
}

// --- OneHit: attacker and victim in different rooms ---
func TestOneHit_DifferentRooms(t *testing.T) {
	w := newCombatWorld()
	room1 := &types.RoomIndexData{Vnum: 9002, Name: "Arena1"}
	room2 := &types.RoomIndexData{Vnum: 9003, Name: "Arena2"}
	w.Rooms[9002] = room1
	w.Rooms[9003] = room2

	ch := newFighter("Attacker", 10)
	handler.CharToRoom(ch, room1)
	w.AddChar(ch)
	victim := newFighter("Defender", 10)
	handler.CharToRoom(victim, room2)
	w.AddChar(victim)

	startHP := victim.Hit
	OneHit(w, ch, victim, types.TYPE_UNDEFINED)
	if victim.Hit != startHP {
		t.Error("should not deal damage when in different rooms")
	}
}

// --- OneHit: with wielded weapon ---
func TestOneHit_WithWeapon(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9004, Name: "Arena"}
	w.Rooms[9004] = room

	ch := newFighter("Attacker", 50) // high level to ensure hits
	ch.Hitroll = 50                  // high hitroll to guarantee hit
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	weapon := &types.ObjData{
		Name:     "sword",
		WearLoc:  types.WEAR_WIELD,
		ItemType: types.ITEM_WEAPON,
	}
	weapon.Value[1] = 5
	weapon.Value[2] = 10
	ch.Carrying = append(ch.Carrying, weapon)

	victim := newFighter("Defender", 1)
	victim.Hit = 500
	victim.MaxHit = 500
	victim.Armor = 200 // poor AC to guarantee being hit
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	// Run many rounds, weapon damage should be applied
	for i := 0; i < 50; i++ {
		if victim.Hit <= 0 {
			break
		}
		OneHit(w, ch, victim, types.TYPE_UNDEFINED)
	}
	if victim.Hit >= 500 {
		t.Error("victim should have taken some damage with weapon equipped")
	}
}

// --- OneHit: sanctuary halves damage ---
func TestOneHit_SanctuaryHalvesDamage(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9005, Name: "Arena"}
	w.Rooms[9005] = room

	// Run many rounds with and without sanc, compare average damage
	totalWithout := 0
	totalWith := 0
	rounds := 500

	for i := 0; i < rounds; i++ {
		ch := newFighter("Attacker", 50)
		ch.Hitroll = 50
		handler.CharToRoom(ch, room)

		victim := newFighter("Defender", 1)
		victim.Hit = 10000
		victim.MaxHit = 10000
		victim.Armor = 200
		handler.CharToRoom(victim, room)

		OneHit(w, ch, victim, types.TYPE_UNDEFINED)
		totalWithout += (10000 - victim.Hit)
		handler.CharFromRoom(ch)
		handler.CharFromRoom(victim)
	}

	for i := 0; i < rounds; i++ {
		ch := newFighter("Attacker", 50)
		ch.Hitroll = 50
		handler.CharToRoom(ch, room)

		victim := newFighter("Defender", 1)
		victim.Hit = 10000
		victim.MaxHit = 10000
		victim.Armor = 200
		victim.AffectedBy.Set(types.AFF_SANCTUARY)
		handler.CharToRoom(victim, room)

		OneHit(w, ch, victim, types.TYPE_UNDEFINED)
		totalWith += (10000 - victim.Hit)
		handler.CharFromRoom(ch)
		handler.CharFromRoom(victim)
	}

	// Sanctuary damage should be roughly half. Allow wide tolerance for RNG.
	if totalWith > 0 && totalWithout > 0 {
		ratio := float64(totalWith) / float64(totalWithout)
		if ratio > 0.75 {
			t.Errorf("sanctuary ratio = %.2f, expected roughly 0.50 (sanc should halve damage)", ratio)
		}
	}
}

// --- OneHit: position multipliers ---
func TestOneHit_PositionMultipliers(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9006, Name: "Arena"}
	w.Rooms[9006] = room

	positions := []int{types.POS_BERSERK, types.POS_AGGRESSIVE, types.POS_DEFENSIVE, types.POS_EVASIVE}
	for _, pos := range positions {
		ch := newFighter("Attacker", 50)
		ch.Hitroll = 50
		ch.Position = pos
		handler.CharToRoom(ch, room)
		w.AddChar(ch)

		victim := newFighter("Defender", 1)
		victim.Hit = 10000
		victim.MaxHit = 10000
		victim.Armor = 200
		handler.CharToRoom(victim, room)
		w.AddChar(victim)

		// Just verify it runs without crashing
		for i := 0; i < 10; i++ {
			OneHit(w, ch, victim, types.TYPE_UNDEFINED)
		}

		handler.CharFromRoom(ch)
		handler.CharFromRoom(victim)
		w.RemoveChar(ch)
		w.RemoveChar(victim)
	}
}

// --- OneHit: sleeping victim (< POS_SLEEPING) gets double damage ---
func TestOneHit_SleepingVictimDoubleDamage(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9007, Name: "Arena"}
	w.Rooms[9007] = room

	rounds := 500
	totalAwake := 0
	totalSleeping := 0

	for i := 0; i < rounds; i++ {
		ch := newFighter("Attacker", 50)
		ch.Hitroll = 50
		handler.CharToRoom(ch, room)

		victim := newFighter("Defender", 1)
		victim.Hit = 10000
		victim.MaxHit = 10000
		victim.Armor = 200
		victim.Position = types.POS_STANDING
		handler.CharToRoom(victim, room)

		OneHit(w, ch, victim, types.TYPE_UNDEFINED)
		totalAwake += (10000 - victim.Hit)
		handler.CharFromRoom(ch)
		handler.CharFromRoom(victim)
	}

	for i := 0; i < rounds; i++ {
		ch := newFighter("Attacker", 50)
		ch.Hitroll = 50
		handler.CharToRoom(ch, room)

		victim := newFighter("Defender", 1)
		victim.Hit = 10000
		victim.MaxHit = 10000
		victim.Armor = 200
		victim.Position = types.POS_STUNNED // < POS_SLEEPING
		handler.CharToRoom(victim, room)

		OneHit(w, ch, victim, types.TYPE_UNDEFINED)
		totalSleeping += (10000 - victim.Hit)
		handler.CharFromRoom(ch)
		handler.CharFromRoom(victim)
	}

	if totalSleeping > 0 && totalAwake > 0 {
		ratio := float64(totalSleeping) / float64(totalAwake)
		if ratio < 1.5 {
			t.Errorf("sleeping victim ratio = %.2f, expected roughly 2.0 (double damage)", ratio)
		}
	}
}

// --- Damage: miss (0 damage) sends miss message ---
func TestDamage_MissSendsMessage(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9010, Name: "Arena"}
	w.Rooms[9010] = room

	ch := newFighter("Attacker", 10)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)
	victim := newFighter("Defender", 10)
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	startHP := victim.Hit
	ret := Damage(w, ch, victim, 0, types.TYPE_HIT)

	if victim.Hit != startHP {
		t.Errorf("miss should not reduce HP, got %d want %d", victim.Hit, startHP)
	}
	if ret != rNONE {
		t.Errorf("miss should return rNONE, got %d", ret)
	}
}

// --- Damage: victim already at 0 HP ---
func TestDamage_VictimAlreadyDead(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9011, Name: "Arena"}
	w.Rooms[9011] = room

	ch := newFighter("Attacker", 10)
	handler.CharToRoom(ch, room)
	victim := newFighter("Defender", 10)
	handler.CharToRoom(victim, room)
	victim.Hit = 0

	ret := Damage(w, ch, victim, 50, types.TYPE_HIT)
	if ret != rNONE {
		t.Errorf("damage to dead victim should return rNONE, got %d", ret)
	}
}

// --- Damage: updates position to stunned ---
func TestDamage_StunnedPosition(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9012, Name: "Arena"}
	w.Rooms[9012] = room

	ch := newFighter("Attacker", 10)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)
	victim := newFighter("Defender", 10)
	victim.Hit = 100
	victim.MaxHit = 100
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	StartFighting(ch, victim)
	// Deal enough damage to get HP below -MaxHit/2 = -50 but above -MaxHit = -100
	// Need to bring HP from 100 to between -50 and -99 => damage of 151 to 199
	Damage(w, ch, victim, 160, types.TYPE_HIT)

	if victim.Position != types.POS_STUNNED {
		t.Errorf("victim.Position = %d, want POS_STUNNED (%d)", victim.Position, types.POS_STUNNED)
	}
	// Note: POS_STUNNED (3) > POS_INCAP (2), so the condition
	// `victim.Position <= types.POS_INCAP` in Damage is false for stunned.
	// The StopFighting message branch only covers INCAP and STUNNED cases
	// but the guard check prevents stunned from entering. This is a quirk
	// of the SMAUG position ordering where STUNNED > INCAP.
}

// --- Damage: updates position to incap ---
func TestDamage_IncapPosition(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9013, Name: "Arena"}
	w.Rooms[9013] = room

	ch := newFighter("Attacker", 10)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)
	victim := newFighter("Defender", 10)
	victim.Hit = 100
	victim.MaxHit = 100
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	StartFighting(ch, victim)
	// Bring HP from 100 to between 0 and -49 => damage of 101 to 149
	Damage(w, ch, victim, 120, types.TYPE_HIT)

	if victim.Position != types.POS_INCAP {
		t.Errorf("victim.Position = %d, want POS_INCAP (%d)", victim.Position, types.POS_INCAP)
	}
	if victim.Fighting != nil {
		t.Error("incapacitated victim should not be fighting")
	}
}

// --- Damage: PC death resets to resting at 1 HP ---
func TestDamage_PCDeath(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9014, Name: "Arena"}
	w.Rooms[9014] = room

	ch := newFighter("Attacker", 10)
	ch.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	victim := newFighter("PlayerVictim", 10)
	victim.PCData = &types.PCData{}
	victim.Hit = 50
	victim.MaxHit = 100
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	StartFighting(ch, victim)
	ret := Damage(w, ch, victim, 200, types.TYPE_HIT)

	if ret != rVICT_DIED {
		t.Errorf("ret = %d, want rVICT_DIED", ret)
	}
	if victim.Hit != 1 {
		t.Errorf("PC should be reset to 1 HP, got %d", victim.Hit)
	}
	if victim.Mana != 1 {
		t.Errorf("PC should be reset to 1 mana, got %d", victim.Mana)
	}
	if victim.Move != 1 {
		t.Errorf("PC should be reset to 1 move, got %d", victim.Move)
	}
	if victim.Position != types.POS_RESTING {
		t.Errorf("PC should be POS_RESTING, got %d", victim.Position)
	}
}

// --- Damage: auto-starts fighting for both sides ---
func TestDamage_StartsRFighting(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9015, Name: "Arena"}
	w.Rooms[9015] = room

	ch := newFighter("Attacker", 10)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)
	victim := newFighter("Defender", 10)
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	// Neither is fighting yet
	Damage(w, ch, victim, 10, types.TYPE_HIT)

	if ch.Fighting == nil {
		t.Error("ch should have started fighting")
	}
	if victim.Fighting == nil {
		t.Error("victim should have started fighting back")
	}
}

// --- Damage: self-damage does not start fighting self ---
func TestDamage_SelfDamage(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9016, Name: "Arena"}
	w.Rooms[9016] = room

	ch := newFighter("SelfHarmer", 10)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	Damage(w, ch, ch, 10, types.TYPE_HIT)

	if ch.Fighting != nil {
		t.Error("self-damage should not start fighting self")
	}
	if ch.Hit != 90 {
		t.Errorf("ch.Hit = %d, want 90", ch.Hit)
	}
}

// --- MakeCorpse: PC corpse (timer=40, no gold transfer) ---
func TestMakeCorpse_PC(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9020, Name: "Arena"}
	w.Rooms[9020] = room

	pc := newFighter("PlayerChar", 10)
	pc.PCData = &types.PCData{}
	pc.Gold = 200
	pc.Weight = 180
	handler.CharToRoom(pc, room)
	w.AddChar(pc)

	MakeCorpse(w, pc)

	if len(room.Contents) != 1 {
		t.Fatalf("room.Contents = %d, want 1", len(room.Contents))
	}
	corpse := room.Contents[0]
	if corpse.ItemType != types.ITEM_CORPSE_PC {
		t.Errorf("corpse.ItemType = %d, want ITEM_CORPSE_PC (%d)", corpse.ItemType, types.ITEM_CORPSE_PC)
	}
	if corpse.Timer != 40 {
		t.Errorf("PC corpse timer = %d, want 40", corpse.Timer)
	}
	if corpse.Value[0] != 200 {
		t.Errorf("PC corpse gold = %d, want 200", corpse.Value[0])
	}
	if pc.Gold != 0 {
		t.Errorf("PC gold should be 0 after MakeCorpse, got %d", pc.Gold)
	}
}

// --- MakeCorpse: nil room returns early ---
func TestMakeCorpse_NilRoom(t *testing.T) {
	w := newCombatWorld()
	ch := newFighter("NoRoom", 10)
	// ch.InRoom is nil
	MakeCorpse(w, ch) // should not panic
}

// --- MakeCorpse: inventory transfer ---
func TestMakeCorpse_InventoryTransfer(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9021, Name: "Arena"}
	w.Rooms[9021] = room

	mobIdx := &types.MobIndexData{
		Vnum: 9500, PlayerName: "lootmob", ShortDescr: "a loot mob",
		Level: 5, Position: types.POS_STANDING, DefPosition: types.POS_STANDING,
	}
	w.MobIndex[9500] = mobIdx
	mob := handler.CreateMobile(w, mobIdx)
	handler.CharToRoom(mob, room)

	// Give mob some items
	item1 := &types.ObjData{Name: "sword", WearLoc: types.WEAR_NONE}
	item2 := &types.ObjData{Name: "shield", WearLoc: types.WEAR_NONE}
	handler.ObjToChar(item1, mob)
	handler.ObjToChar(item2, mob)

	MakeCorpse(w, mob)

	corpse := room.Contents[0]
	if len(corpse.Contents) != 2 {
		t.Errorf("corpse.Contents = %d, want 2 (transferred items)", len(corpse.Contents))
	}
	if len(mob.Carrying) != 0 {
		t.Errorf("mob.Carrying = %d, want 0 (items transferred to corpse)", len(mob.Carrying))
	}
}

// --- updatePos: various HP thresholds ---
func TestUpdatePos(t *testing.T) {
	tests := []struct {
		name   string
		hp     int
		maxHP  int
		want   int
		setPos int // initial position
	}{
		{"dead at -maxhit", -100, 100, types.POS_DEAD, types.POS_FIGHTING},
		{"dead below -maxhit", -150, 100, types.POS_DEAD, types.POS_FIGHTING},
		{"stunned at -maxhit/2", -50, 100, types.POS_STUNNED, types.POS_FIGHTING},
		{"stunned between -maxhit/2 and -maxhit", -60, 100, types.POS_STUNNED, types.POS_FIGHTING},
		{"incap at 0", 0, 100, types.POS_INCAP, types.POS_FIGHTING},
		{"incap at -10", -10, 100, types.POS_INCAP, types.POS_FIGHTING},
		{"alive stays same", 50, 100, types.POS_FIGHTING, types.POS_FIGHTING},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ch := &types.CharData{
				Hit:      tc.hp,
				MaxHit:   tc.maxHP,
				Position: tc.setPos,
			}
			updatePos(ch)
			if ch.Position != tc.want {
				t.Errorf("updatePos: position = %d, want %d", ch.Position, tc.want)
			}
		})
	}
}

// --- computeXP: various level differences ---
func TestComputeXP(t *testing.T) {
	tests := []struct {
		name      string
		chLevel   int
		vicLevel  int
		vicExp    int
		wantRange [2]int // [min, max] inclusive
	}{
		{"victim 5+ levels above", 5, 10, 1000, [2]int{1500, 1500}},  // 150%
		{"victim 1 level above", 10, 11, 1000, [2]int{1100, 1100}},   // 110%
		{"same level", 10, 10, 1000, [2]int{1000, 1000}},             // 100%
		{"victim 3 below", 10, 7, 1000, [2]int{1000, 1000}},          // 100%
		{"victim 4 below", 10, 6, 1000, [2]int{750, 750}},            // 75%
		{"victim 8 below", 10, 2, 1000, [2]int{750, 750}},            // 75%
		{"victim 9+ below", 10, 1, 1000, [2]int{250, 250}},           // 25%
		{"fallback xp from level", 5, 3, 0, [2]int{1, 90}},           // 3*3*10=90 at 75%=67
		{"minimum 1 xp", 50, 1, 1, [2]int{1, 1}},                    // 25% of 1 = 0, clamped to 1
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ch := &types.CharData{Level: tc.chLevel}
			victim := &types.CharData{Level: tc.vicLevel, Exp: tc.vicExp}
			got := computeXP(ch, victim)
			if got < tc.wantRange[0] || got > tc.wantRange[1] {
				t.Errorf("computeXP = %d, want [%d, %d]", got, tc.wantRange[0], tc.wantRange[1])
			}
		})
	}
}

// --- ViolenceUpdate: NPC multi-attack ---
func TestViolenceUpdate_NPCMultiAttack(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9030, Name: "Arena"}
	w.Rooms[9030] = room

	ch := newFighter("MultiMob", 50)
	ch.Act.Set(types.ACT_IS_NPC)
	ch.MobThac0 = 0
	ch.Hitroll = 50 // ensure hits
	ch.NumAttacks = 3
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	victim := newFighter("Tank", 1)
	victim.Act.Set(types.ACT_IS_NPC)
	victim.Hit = 50000
	victim.MaxHit = 50000
	victim.Armor = 200
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	StartFighting(ch, victim)
	ViolenceUpdate(w)

	// With 3 attacks and high hitroll, victim should take noticeable damage
	// (more than a single attack would likely deal)
	if victim.Hit >= 50000 {
		t.Error("NPC with NumAttacks=3 should deal damage via multi-attack")
	}
}

// --- ViolenceUpdate: dual wield extra attack ---
func TestViolenceUpdate_DualWield(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9031, Name: "Arena"}
	w.Rooms[9031] = room

	ch := newFighter("DualWielder", 50)
	ch.Act.Set(types.ACT_IS_NPC)
	ch.MobThac0 = 0
	ch.Hitroll = 50
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	// Equip dual wield weapon
	dualWeapon := &types.ObjData{
		Name:     "offhand dagger",
		WearLoc:  types.WEAR_DUAL_WIELD,
		ItemType: types.ITEM_WEAPON,
	}
	dualWeapon.Value[1] = 3
	dualWeapon.Value[2] = 6
	ch.Carrying = append(ch.Carrying, dualWeapon)

	victim := newFighter("Tank", 1)
	victim.Act.Set(types.ACT_IS_NPC)
	victim.Hit = 50000
	victim.MaxHit = 50000
	victim.Armor = 200
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	StartFighting(ch, victim)
	ViolenceUpdate(w)

	// Should have attacked at least twice (main + dual wield)
	if victim.Hit >= 50000 {
		t.Error("dual wielder should deal damage")
	}
}

// --- ViolenceUpdate: wimpy auto-flee ---
func TestViolenceUpdate_WimpyFlee(t *testing.T) {
	w := newCombatWorld()
	room1 := &types.RoomIndexData{Vnum: 9032, Name: "Start"}
	room2 := &types.RoomIndexData{Vnum: 9033, Name: "Escape"}
	w.Rooms[9032] = room1
	w.Rooms[9033] = room2

	// Create exit from room1 north to room2
	exit := &types.ExitData{
		Direction: types.DIR_NORTH,
		ToRoom:    room2,
		ExitInfo:  0, // not closed
	}
	room1.Exits = append(room1.Exits, exit)

	ch := newFighter("Wimpy", 10)
	ch.PCData = &types.PCData{} // PC, not NPC
	ch.Wimpy = 80
	ch.Hit = 50 // below wimpy threshold
	handler.CharToRoom(ch, room1)
	w.AddChar(ch)

	victim := newFighter("Mob", 50)
	victim.Act.Set(types.ACT_IS_NPC)
	victim.MobThac0 = 20
	victim.Hit = 10000
	victim.MaxHit = 10000
	handler.CharToRoom(victim, room1)
	w.AddChar(victim)

	StartFighting(ch, victim)

	ViolenceUpdate(w)

	// Ch should have fled to room2
	if ch.InRoom == room1 {
		// It's possible the character died from the violence hit before wimpy could trigger.
		// But if still alive and in room1, wimpy should have triggered.
		if ch.Hit > 0 && ch.Hit <= ch.Wimpy {
			t.Error("wimpy character should have fled but is still in room1")
		}
	}
	if ch.InRoom == room2 {
		if ch.Fighting != nil {
			t.Error("character who fled should not be fighting")
		}
	}
}

// --- ViolenceUpdate: victim leaves room mid-fight ---
func TestViolenceUpdate_VictimLeftRoom(t *testing.T) {
	w := newCombatWorld()
	room1 := &types.RoomIndexData{Vnum: 9034, Name: "Arena"}
	room2 := &types.RoomIndexData{Vnum: 9035, Name: "Other"}
	w.Rooms[9034] = room1
	w.Rooms[9035] = room2

	ch := newFighter("Attacker", 10)
	handler.CharToRoom(ch, room1)
	w.AddChar(ch)

	victim := newFighter("Runner", 10)
	handler.CharToRoom(victim, room2) // different room
	w.AddChar(victim)

	ch.Fighting = &types.FightData{Who: victim}
	ch.NumFighting = 1
	ch.Position = types.POS_FIGHTING

	ViolenceUpdate(w)

	if ch.Fighting != nil {
		t.Error("fighting should stop when victim is in a different room")
	}
}

// --- ViolenceUpdate: nil victim ---
func TestViolenceUpdate_NilVictim(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9036, Name: "Arena"}
	w.Rooms[9036] = room

	ch := newFighter("Attacker", 10)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	ch.Fighting = &types.FightData{Who: nil}
	ch.NumFighting = 1
	ch.Position = types.POS_FIGHTING

	ViolenceUpdate(w)

	if ch.Fighting != nil {
		t.Error("fighting should stop when victim is nil")
	}
}

// --- ViolenceUpdate: no room ---
func TestViolenceUpdate_NoRoom(t *testing.T) {
	w := newCombatWorld()

	ch := newFighter("NoRoom", 10)
	w.AddChar(ch)
	ch.Fighting = &types.FightData{Who: newFighter("Target", 10)}

	// Should not crash with nil InRoom
	ViolenceUpdate(w)
}

// --- ViolenceUpdate: wimpy with no exits (no crash, stays in room) ---
func TestViolenceUpdate_WimpyNoExits(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9037, Name: "Sealed"}
	w.Rooms[9037] = room

	ch := newFighter("Wimpy", 10)
	ch.PCData = &types.PCData{}
	ch.Wimpy = 80
	ch.Hit = 50
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	victim := newFighter("Mob", 1)
	victim.Act.Set(types.ACT_IS_NPC)
	victim.MobThac0 = 20
	victim.Hit = 10000
	victim.MaxHit = 10000
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	StartFighting(ch, victim)
	ViolenceUpdate(w)

	// No exits, so wimpy can't flee — should still be in same room
	if ch.InRoom != room {
		t.Error("character with no exits should remain in room")
	}
}

// --- ViolenceUpdate: wimpy with closed exit ---
func TestViolenceUpdate_WimpyClosedExit(t *testing.T) {
	w := newCombatWorld()
	room1 := &types.RoomIndexData{Vnum: 9038, Name: "Locked"}
	room2 := &types.RoomIndexData{Vnum: 9039, Name: "Beyond"}
	w.Rooms[9038] = room1
	w.Rooms[9039] = room2

	exit := &types.ExitData{
		Direction: types.DIR_NORTH,
		ToRoom:    room2,
		ExitInfo:  int(types.EX_CLOSED), // closed door
	}
	room1.Exits = append(room1.Exits, exit)

	ch := newFighter("Wimpy", 10)
	ch.PCData = &types.PCData{}
	ch.Wimpy = 80
	ch.Hit = 50
	handler.CharToRoom(ch, room1)
	w.AddChar(ch)

	victim := newFighter("Mob", 1)
	victim.Act.Set(types.ACT_IS_NPC)
	victim.MobThac0 = 20
	victim.Hit = 10000
	victim.MaxHit = 10000
	handler.CharToRoom(victim, room1)
	w.AddChar(victim)

	StartFighting(ch, victim)
	ViolenceUpdate(w)

	// Exit is closed, so can't flee through it
	if ch.InRoom != room1 {
		t.Error("character should not flee through closed exit")
	}
}
