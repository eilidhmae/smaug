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
		{"victim 5+ levels above", 5, 10, 1000, [2]int{1500, 1500}}, // 150%
		{"victim 1 level above", 10, 11, 1000, [2]int{1100, 1100}},  // 110%
		{"same level", 10, 10, 1000, [2]int{1000, 1000}},            // 100%
		{"victim 3 below", 10, 7, 1000, [2]int{1000, 1000}},         // 100%
		{"victim 4 below", 10, 6, 1000, [2]int{750, 750}},           // 75%
		{"victim 8 below", 10, 2, 1000, [2]int{750, 750}},           // 75%
		{"victim 9+ below", 10, 1, 1000, [2]int{250, 250}},          // 25%
		{"fallback xp from level", 5, 3, 0, [2]int{1, 90}},          // 3*3*10=90 at 75%=67
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

// --- G1 — OneHit retcode + MultiHit extraction ---

// OneHit now returns a retcode: rNONE on hit/miss, rVICT_DIED on kill.
func TestOneHit_ReturnsRNoneOnHit(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9100, Name: "Arena"}
	w.Rooms[9100] = room

	ch := newFighter("Attacker", 10)
	ch.Hitroll = 50
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	victim := newFighter("Defender", 10)
	victim.Hit = 500
	victim.MaxHit = 500
	victim.Armor = 200
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	ret := OneHit(w, ch, victim, types.TYPE_UNDEFINED)
	if ret != rNONE {
		t.Errorf("OneHit on live victim returned %d, want rNONE", ret)
	}
}

func TestOneHit_ReturnsRVictDiedOnKill(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9101, Name: "Arena"}
	w.Rooms[9101] = room

	ch := newFighter("Attacker", 50)
	// Huge hitroll saturates the 0..19 d20 cap, guaranteeing a hit
	// (rollD20 < thac0 - victimAC always hits except on natural 0 which
	// is 1/20 ≈ 5%). Run a tight loop to de-flake the natural-0 case.
	ch.Hitroll = 9999
	ch.Damroll = 9999
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	mobIdx := &types.MobIndexData{
		Vnum: 9800, PlayerName: "lowlife", ShortDescr: "a lowlife",
		Level: 1, Position: types.POS_STANDING, DefPosition: types.POS_STANDING,
	}
	w.MobIndex[9800] = mobIdx
	victim := handler.CreateMobile(w, mobIdx)
	victim.Hit = 1
	victim.MaxHit = 10
	victim.Armor = 9999 // very poor AC to further de-flake
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	// Retry a handful of times — each OneHit has ≥19/20 hit chance, so
	// 5 attempts keeps flake probability below 1e-6.
	var ret int
	for i := 0; i < 5; i++ {
		victim.Hit = 1
		victim.MaxHit = 10
		victim.Position = types.POS_STANDING
		ret = OneHit(w, ch, victim, types.TYPE_UNDEFINED)
		if ret == rVICT_DIED {
			return
		}
	}
	t.Errorf("OneHit killing victim returned %d after 5 retries, want rVICT_DIED", ret)
}

// OneHit early-out on already-dead victim returns rVICT_DIED (mirrors C fight.c:1394).
func TestOneHit_EarlyOutReturnsRVictDied(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9102, Name: "Arena"}
	w.Rooms[9102] = room

	ch := newFighter("Attacker", 10)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)
	victim := newFighter("Defender", 10)
	victim.Hit = 0
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	if ret := OneHit(w, ch, victim, types.TYPE_UNDEFINED); ret != rVICT_DIED {
		t.Errorf("OneHit on dead victim returned %d, want rVICT_DIED", ret)
	}
}

// MultiHit exists and returns a retcode.
func TestMultiHit_SingleHitReturnsRNone(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9103, Name: "Arena"}
	w.Rooms[9103] = room

	ch := newFighter("Attacker", 10)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)
	victim := newFighter("Defender", 10)
	victim.Hit = 500
	victim.MaxHit = 500
	victim.Armor = 200
	handler.CharToRoom(victim, room)
	w.AddChar(victim)
	StartFighting(ch, victim)

	ret := MultiHit(w, ch, victim, types.TYPE_UNDEFINED)
	if ret != rNONE && ret != rVICT_DIED {
		t.Errorf("MultiHit returned %d, want rNONE or rVICT_DIED", ret)
	}
}

// MultiHit short-circuits on victim death.
func TestMultiHit_ShortCircuitsOnDeath(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9104, Name: "Arena"}
	w.Rooms[9104] = room

	ch := newFighter("Attacker", 50)
	ch.Act.Set(types.ACT_IS_NPC)
	ch.NumAttacks = 5 // would try 5 swings
	ch.Hitroll = 100
	ch.Damroll = 9999
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	mobIdx := &types.MobIndexData{
		Vnum: 9801, PlayerName: "fragile", ShortDescr: "a fragile mob",
		Level: 1, Position: types.POS_STANDING, DefPosition: types.POS_STANDING,
	}
	w.MobIndex[9801] = mobIdx
	victim := handler.CreateMobile(w, mobIdx)
	victim.Hit = 1
	victim.MaxHit = 10
	victim.Armor = 200
	handler.CharToRoom(victim, room)
	w.AddChar(victim)
	StartFighting(ch, victim)

	ret := MultiHit(w, ch, victim, types.TYPE_UNDEFINED)
	if ret != rVICT_DIED {
		t.Errorf("MultiHit returned %d when victim died on first swing, want rVICT_DIED", ret)
	}
	// Victim must be extracted (NPC died) — InRoom should be nil after ExtractChar.
	if victim.InRoom != nil && victim.Hit > 0 {
		t.Errorf("victim should be dead/extracted; has Hit=%d InRoom=%v", victim.Hit, victim.InRoom)
	}
}

// --- G3 — PC multi-attack cascade ---

// Helper: install a deterministic numberPercent stub and an OneHit spy
// that counts BOTH primary and offhand calls without mutating the
// victim. Returns the unified counter pointer and cleans up on test
// end. Most cascade tests don't care whether a swing was primary or
// offhand — they care about TOTAL OneHit invocations per round.
func stubCascade(t *testing.T, pctFn func() int) *int {
	t.Helper()
	savedPct := numberPercent
	savedOne := oneHit
	savedOff := oneHitOffhand
	t.Cleanup(func() {
		numberPercent = savedPct
		oneHit = savedOne
		oneHitOffhand = savedOff
	})
	numberPercent = pctFn
	calls := 0
	oneHit = func(w *world.World, ch, victim *types.CharData, dt int) int {
		calls++
		return rNONE
	}
	oneHitOffhand = func(w *world.World, ch, victim *types.CharData, dt int) int {
		calls++
		return rNONE
	}
	return &calls
}

// Helper: populate the gsn cache with distinct, non-(-1) values so the
// cascade path runs (no -1 early-outs). Restores on cleanup. The values
// picked here don't need to correspond to a real skill table; only the
// nonnegative check matters plus the less-than-MAX_SKILL bounds check.
// We choose values well below MAX_SKILL=600.
func stubGsnsForCascade(t *testing.T) {
	t.Helper()
	saved := struct {
		s, th, f, fi, si, se int
		bs, ci, po           int
		dw, be               int
	}{
		gsnSecondAttack, gsnThirdAttack, gsnFourthAttack,
		gsnFifthAttack, gsnSixthAttack, gsnSeventhAttack,
		gsnBackstab, gsnCircle, gsnPounce, gsnDualWield, gsnBerserk,
	}
	t.Cleanup(func() {
		gsnSecondAttack = saved.s
		gsnThirdAttack = saved.th
		gsnFourthAttack = saved.f
		gsnFifthAttack = saved.fi
		gsnSixthAttack = saved.si
		gsnSeventhAttack = saved.se
		gsnBackstab = saved.bs
		gsnCircle = saved.ci
		gsnPounce = saved.po
		gsnDualWield = saved.dw
		gsnBerserk = saved.be
	})
	gsnSecondAttack = 51
	gsnThirdAttack = 52
	gsnFourthAttack = 53
	gsnFifthAttack = 54
	gsnSixthAttack = 55
	gsnSeventhAttack = 56
	gsnBackstab = 60
	gsnCircle = 61
	gsnPounce = 62
	gsnDualWield = 57
	gsnBerserk = 58
}

// Helper: build a level-N PC fighter primed for the cascade. Sets all
// Learned[] entries to 0 so subclasses can set only the ones they want.
func newPCFighter(level int) *types.CharData {
	ch := newFighter("PC", level)
	ch.PCData = &types.PCData{}
	return ch
}

// Helper: build a scratch room + primed combat state for cascade tests.
// Returns (world, room, ch, victim); ch.Fighting is already set pointing
// at victim. Victim is an NPC with huge HP so OneHit spy doesn't kill.
func setupCascade(t *testing.T, level int) (*world.World, *types.RoomIndexData, *types.CharData, *types.CharData) {
	t.Helper()
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 10000, Name: "Cascade Arena"}
	w.Rooms[10000] = room

	ch := newPCFighter(level)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	victim := newFighter("Victim", 1)
	victim.Act.Set(types.ACT_IS_NPC)
	victim.Hit = 1_000_000
	victim.MaxHit = 1_000_000
	handler.CharToRoom(victim, room)
	w.AddChar(victim)

	StartFighting(ch, victim)
	return w, room, ch, victim
}

// Level-30 warrior with Learned[second_attack]=100, stub pct=50.
// chance = (100 + 0) * 2/3 = 66. 50 < 66 → second attack fires.
// Expected OneHit calls: 2 (primary + second).
func TestMultiHit_SecondAttackFires_Learned100(t *testing.T) {
	stubGsnsForCascade(t)
	calls := stubCascade(t, func() int { return 50 })
	_, _, ch, victim := setupCascade(t, 30)
	ch.PCData.Learned[gsnSecondAttack] = 100

	MultiHit(nil, ch, victim, types.TYPE_UNDEFINED)

	if *calls != 2 {
		t.Errorf("OneHit calls = %d, want 2 (primary + second)", *calls)
	}
}

// Stub pct=50, Learned[third]=0 → chance for third is 0, no fire.
// Learned[second]=0 → second chance 0 either. So only primary hits.
func TestMultiHit_ThirdAttack_Learned0_DoesNotFire(t *testing.T) {
	stubGsnsForCascade(t)
	calls := stubCascade(t, func() int { return 50 })
	_, _, ch, victim := setupCascade(t, 30)
	// Don't set any Learned — all 0.

	MultiHit(nil, ch, victim, types.TYPE_UNDEFINED)

	if *calls != 1 {
		t.Errorf("OneHit calls = %d, want 1 (primary only)", *calls)
	}
}

// Boundary check: Learned[second]=75, stub pct=40.
// chance = 75 * 2/3 = 50. 40 < 50 → fires.
// Then stub pct=80. 80 < 50 → false → no fire.
func TestMultiHit_SecondAttack_Learned75_Boundary(t *testing.T) {
	stubGsnsForCascade(t)

	// First: pct=40, fires.
	calls := stubCascade(t, func() int { return 40 })
	_, _, ch, victim := setupCascade(t, 30)
	ch.PCData.Learned[gsnSecondAttack] = 75
	MultiHit(nil, ch, victim, types.TYPE_UNDEFINED)
	if *calls != 2 {
		t.Errorf("pct=40 Learned=75: calls = %d, want 2", *calls)
	}

	// Second: pct=80, no fire.
	calls2 := stubCascade(t, func() int { return 80 })
	_, _, ch2, victim2 := setupCascade(t, 30)
	ch2.PCData.Learned[gsnSecondAttack] = 75
	MultiHit(nil, ch2, victim2, types.TYPE_UNDEFINED)
	if *calls2 != 1 {
		t.Errorf("pct=80 Learned=75: calls = %d, want 1", *calls2)
	}
}

// NPC with NumAttacks=1 uses level-driven path but the cascade is PC-only.
// NPCs should not enter the cascade regardless of stub pct. So with NPC
// stub, only primary fires.
func TestMultiHit_NPCDoesNotCascade(t *testing.T) {
	stubGsnsForCascade(t)
	calls := stubCascade(t, func() int { return 0 }) // always true for PC
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 10001, Name: "NPC Arena"}
	w.Rooms[10001] = room

	ch := newFighter("NpcFighter", 30)
	ch.Act.Set(types.ACT_IS_NPC)
	ch.NumAttacks = 1
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	victim := newFighter("Victim", 1)
	victim.Act.Set(types.ACT_IS_NPC)
	victim.Hit = 1_000_000
	victim.MaxHit = 1_000_000
	handler.CharToRoom(victim, room)
	w.AddChar(victim)
	StartFighting(ch, victim)

	MultiHit(w, ch, victim, types.TYPE_UNDEFINED)

	if *calls != 1 {
		t.Errorf("NPC OneHit calls = %d, want 1 (no cascade)", *calls)
	}
}

// LearnFromSuccessHook fires on a cascade success.
func TestMultiHit_CascadeFiresLearnFromSuccess(t *testing.T) {
	stubGsnsForCascade(t)
	_ = stubCascade(t, func() int { return 10 })

	savedS := LearnFromSuccessHook
	savedF := LearnFromFailureHook
	t.Cleanup(func() {
		LearnFromSuccessHook = savedS
		LearnFromFailureHook = savedF
	})
	var sHits []int
	LearnFromSuccessHook = func(_ *types.CharData, gsn int) {
		sHits = append(sHits, gsn)
	}
	LearnFromFailureHook = func(_ *types.CharData, _ int) {}

	_, _, ch, victim := setupCascade(t, 30)
	ch.PCData.Learned[gsnSecondAttack] = 100

	MultiHit(nil, ch, victim, types.TYPE_UNDEFINED)

	found := false
	for _, g := range sHits {
		if g == gsnSecondAttack {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected learnFromSuccess(gsnSecondAttack); got calls %v", sHits)
	}
}

// LearnFromFailureHook fires on a cascade miss.
func TestMultiHit_CascadeFiresLearnFromFailure(t *testing.T) {
	stubGsnsForCascade(t)
	_ = stubCascade(t, func() int { return 99 })

	savedS := LearnFromSuccessHook
	savedF := LearnFromFailureHook
	t.Cleanup(func() {
		LearnFromSuccessHook = savedS
		LearnFromFailureHook = savedF
	})
	var fHits []int
	LearnFromSuccessHook = func(_ *types.CharData, _ int) {}
	LearnFromFailureHook = func(_ *types.CharData, gsn int) {
		fHits = append(fHits, gsn)
	}

	_, _, ch, victim := setupCascade(t, 30)
	ch.PCData.Learned[gsnSecondAttack] = 50 // chance ~33, pct=99 → miss

	MultiHit(nil, ch, victim, types.TYPE_UNDEFINED)

	found := false
	for _, g := range fHits {
		if g == gsnSecondAttack {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected learnFromFailure(gsnSecondAttack); got calls %v", fHits)
	}
}

// Cascade short-circuits on victim death mid-cascade.
func TestMultiHit_CascadeShortCircuitsOnVictimDeath(t *testing.T) {
	stubGsnsForCascade(t)
	savedPct := numberPercent
	savedOne := oneHit
	t.Cleanup(func() {
		numberPercent = savedPct
		oneHit = savedOne
	})

	calls := 0
	// OneHit: first call alive, second call kills victim.
	oneHit = func(w *world.World, ch, victim *types.CharData, dt int) int {
		calls++
		if calls == 2 {
			return rVICT_DIED
		}
		return rNONE
	}
	numberPercent = func() int { return 0 } // always fires cascade tier

	_, _, ch, victim := setupCascade(t, 30)
	ch.PCData.Learned[gsnSecondAttack] = 100
	ch.PCData.Learned[gsnThirdAttack] = 100

	ret := MultiHit(nil, ch, victim, types.TYPE_UNDEFINED)
	if ret != rVICT_DIED {
		t.Errorf("MultiHit returned %d, want rVICT_DIED", ret)
	}
	if calls != 2 {
		t.Errorf("OneHit calls = %d, want 2 (primary + second; third never runs)", calls)
	}
}

// dt == gsn_backstab short-circuits the cascade (single hit only).
func TestMultiHit_BackstabSkipsCascade(t *testing.T) {
	stubGsnsForCascade(t)
	calls := stubCascade(t, func() int { return 0 })
	_, _, ch, victim := setupCascade(t, 30)
	// Even with 100% learned in everything, backstab dt suppresses the
	// cascade entirely.
	for i := range ch.PCData.Learned {
		ch.PCData.Learned[i] = 100
	}

	MultiHit(nil, ch, victim, gsnBackstab)

	if *calls != 1 {
		t.Errorf("backstab OneHit calls = %d, want 1 (no cascade)", *calls)
	}
}

// dt == gsn_circle short-circuits the cascade.
func TestMultiHit_CircleSkipsCascade(t *testing.T) {
	stubGsnsForCascade(t)
	calls := stubCascade(t, func() int { return 0 })
	_, _, ch, victim := setupCascade(t, 30)
	for i := range ch.PCData.Learned {
		ch.PCData.Learned[i] = 100
	}

	MultiHit(nil, ch, victim, gsnCircle)

	if *calls != 1 {
		t.Errorf("circle OneHit calls = %d, want 1 (no cascade)", *calls)
	}
}

// Full cascade with Learned[2..7]=100 and pct=0 fires all 6 tiers.
// Total OneHit calls = primary + 6 cascade = 7.
func TestMultiHit_AllTiersFire_AllLearned100(t *testing.T) {
	stubGsnsForCascade(t)
	calls := stubCascade(t, func() int { return 0 })
	_, _, ch, victim := setupCascade(t, 50)
	ch.PCData.Learned[gsnSecondAttack] = 100
	ch.PCData.Learned[gsnThirdAttack] = 100
	ch.PCData.Learned[gsnFourthAttack] = 100
	ch.PCData.Learned[gsnFifthAttack] = 100
	ch.PCData.Learned[gsnSixthAttack] = 100
	ch.PCData.Learned[gsnSeventhAttack] = 100

	MultiHit(nil, ch, victim, types.TYPE_UNDEFINED)

	if *calls != 7 {
		t.Errorf("AllLearned100 OneHit calls = %d, want 7 (primary + 6 tiers)", *calls)
	}
}

// --- G4 — Dual-wield learned roll + dual_bonus threading ---

// DualWield gated by learned roll: Learned[dual_wield]=100, pct=50 → fires.
// Expect 2 OneHit calls (primary + dual-wield extra). Cascade Learned[2..7]=0
// so those tiers don't contribute.
func TestMultiHit_DualWield_LearnedGate_Fires(t *testing.T) {
	stubGsnsForCascade(t)
	calls := stubCascade(t, func() int { return 50 })
	_, room, ch, victim := setupCascade(t, 30)
	ch.PCData.Learned[gsnDualWield] = 100

	// Put a dual-wield weapon on ch.
	offhand := &types.ObjData{
		Name:     "offhand dagger",
		WearLoc:  types.WEAR_DUAL_WIELD,
		ItemType: types.ITEM_WEAPON,
	}
	offhand.Value[1] = 1
	offhand.Value[2] = 4
	handler.ObjToChar(offhand, ch)
	_ = room

	MultiHit(nil, ch, victim, types.TYPE_UNDEFINED)

	if *calls != 2 {
		t.Errorf("OneHit calls = %d, want 2 (primary + dual-wield extra)", *calls)
	}
}

// DualWield gated: Learned=10, pct=50 → 50 < 10 is false → NO extra hit.
// Only primary OneHit.
func TestMultiHit_DualWield_LearnedGate_Fails(t *testing.T) {
	stubGsnsForCascade(t)
	calls := stubCascade(t, func() int { return 50 })
	_, _, ch, victim := setupCascade(t, 30)
	ch.PCData.Learned[gsnDualWield] = 10

	offhand := &types.ObjData{
		Name:     "offhand dagger",
		WearLoc:  types.WEAR_DUAL_WIELD,
		ItemType: types.ITEM_WEAPON,
	}
	offhand.Value[1] = 1
	offhand.Value[2] = 4
	handler.ObjToChar(offhand, ch)

	MultiHit(nil, ch, victim, types.TYPE_UNDEFINED)

	if *calls != 1 {
		t.Errorf("OneHit calls = %d, want 1 (dual-wield roll failed)", *calls)
	}
}

// Low move (<10) sets dual_bonus to -20 even without dual wield.
// With Learned[second]=100 and no dual-bonus, second chance = 100*2/3 = 66.
// With dual_bonus=-20, second chance = (100-20)*2/3 = 53. pct=60 → 60<53 false.
// (Without low-move, 60<66 → true.)
// So Move<10 should suppress the second attack at pct=60 Learned=100.
func TestMultiHit_LowMovePenalty_SuppressesCascade(t *testing.T) {
	stubGsnsForCascade(t)
	calls := stubCascade(t, func() int { return 60 })
	_, _, ch, victim := setupCascade(t, 30)
	ch.PCData.Learned[gsnSecondAttack] = 100
	ch.Move = 5 // below 10 threshold

	MultiHit(nil, ch, victim, types.TYPE_UNDEFINED)

	if *calls != 1 {
		t.Errorf("Move<10 + pct=60: OneHit calls = %d, want 1 (second suppressed by -20 dual_bonus)", *calls)
	}
}

// Sanity check: at normal Move, pct=60 Learned=100 → second DOES fire.
// Primary + second = 2 calls.
func TestMultiHit_LowMovePenalty_BaselineFiresAtPct60(t *testing.T) {
	stubGsnsForCascade(t)
	calls := stubCascade(t, func() int { return 60 })
	_, _, ch, victim := setupCascade(t, 30)
	ch.PCData.Learned[gsnSecondAttack] = 100
	ch.Move = 80 // normal

	MultiHit(nil, ch, victim, types.TYPE_UNDEFINED)

	if *calls != 2 {
		t.Errorf("Move=80 + pct=60 + Learned=100: calls = %d, want 2", *calls)
	}
}

// NPC dual-wield: chance = ch.Level, dualBonus = ch.Level/10.
func TestMultiHit_NPCDualWield_UsesLevel(t *testing.T) {
	stubGsnsForCascade(t)
	// Two separate scenarios in one test: level 100 (always fires) and
	// level 0 (never fires). Use the same seed by calling stubCascade
	// with a closure that returns 50.
	calls := stubCascade(t, func() int { return 50 })
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 10010, Name: "NPC Arena"}
	w.Rooms[10010] = room

	ch := newFighter("NpcDW", 100)
	ch.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	offhand := &types.ObjData{
		Name:    "offhand",
		WearLoc: types.WEAR_DUAL_WIELD,
	}
	offhand.Value[1] = 1
	offhand.Value[2] = 4
	handler.ObjToChar(offhand, ch)

	victim := newFighter("Victim", 1)
	victim.Act.Set(types.ACT_IS_NPC)
	victim.Hit = 1_000_000
	victim.MaxHit = 1_000_000
	handler.CharToRoom(victim, room)
	w.AddChar(victim)
	StartFighting(ch, victim)

	MultiHit(nil, ch, victim, types.TYPE_UNDEFINED)

	if *calls != 2 {
		t.Errorf("NPC level=100 dual-wield pct=50: calls = %d, want 2", *calls)
	}
}

// --- G5 — Gates (NOATTACK, BERSERK, PLR_NICE, attack-suppress) ---

// ACT_NOATTACK mob returns immediately, no OneHit fires.
func TestMultiHit_NoAttackMobSkips(t *testing.T) {
	stubGsnsForCascade(t)
	calls := stubCascade(t, func() int { return 0 })
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 10020, Name: "NoAttack Arena"}
	w.Rooms[10020] = room

	ch := newFighter("NoAttackMob", 30)
	ch.Act.Set(types.ACT_IS_NPC)
	ch.Act.Set(types.ACT_NOATTACK)
	ch.NumAttacks = 3
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	victim := newFighter("Victim", 1)
	victim.Act.Set(types.ACT_IS_NPC)
	victim.Hit = 1_000_000
	victim.MaxHit = 1_000_000
	handler.CharToRoom(victim, room)
	w.AddChar(victim)
	StartFighting(ch, victim)

	if ret := MultiHit(w, ch, victim, types.TYPE_UNDEFINED); ret != rNONE {
		t.Errorf("NOATTACK mob returned %d, want rNONE", ret)
	}
	if *calls != 0 {
		t.Errorf("NOATTACK mob OneHit calls = %d, want 0", *calls)
	}
}

// PLR_NICE on PC attacker suppresses MultiHit against another PC.
func TestMultiHit_PLRNiceSkipsPvP(t *testing.T) {
	stubGsnsForCascade(t)
	calls := stubCascade(t, func() int { return 0 })
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 10021, Name: "PvP Arena"}
	w.Rooms[10021] = room

	ch := newPCFighter(30)
	ch.Act.Set(types.PLR_NICE)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)
	victim := newPCFighter(30)
	handler.CharToRoom(victim, room)
	w.AddChar(victim)
	StartFighting(ch, victim)

	if ret := MultiHit(w, ch, victim, types.TYPE_UNDEFINED); ret != rNONE {
		t.Errorf("PLR_NICE PC-vs-PC returned %d, want rNONE", ret)
	}
	if *calls != 0 {
		t.Errorf("PLR_NICE OneHit calls = %d, want 0", *calls)
	}
}

// PLR_NICE does NOT suppress MultiHit against an NPC (only PC-vs-PC).
func TestMultiHit_PLRNiceDoesNotAffectPvNpc(t *testing.T) {
	stubGsnsForCascade(t)
	calls := stubCascade(t, func() int { return 100 })
	_, _, ch, victim := setupCascade(t, 30)
	ch.Act.Set(types.PLR_NICE) // victim is already NPC

	MultiHit(nil, ch, victim, types.TYPE_UNDEFINED)

	if *calls < 1 {
		t.Errorf("PLR_NICE vs NPC: OneHit calls = %d, want >= 1", *calls)
	}
}

// TIMER_ASUPRESSED active on ch skips the entire round.
func TestMultiHit_AttackSuppressedSkips(t *testing.T) {
	stubGsnsForCascade(t)
	calls := stubCascade(t, func() int { return 0 })
	_, _, ch, victim := setupCascade(t, 30)
	handler.AddTimer(ch, types.TIMER_ASUPRESSED, 5, "", 0)

	if ret := MultiHit(nil, ch, victim, types.TYPE_UNDEFINED); ret != rNONE {
		t.Errorf("attack-suppressed returned %d, want rNONE", ret)
	}
	if *calls != 0 {
		t.Errorf("attack-suppressed OneHit calls = %d, want 0", *calls)
	}
}

// AFF_BERSERK with Learned[berserk]=33 → chance = 33*6/2 = 99. pct=50 → fires.
// Primary + berserk = 2 calls (cascade tiers have Learned=0 → no fire).
func TestMultiHit_BerserkExtraHit_Fires(t *testing.T) {
	stubGsnsForCascade(t)
	calls := stubCascade(t, func() int { return 50 })
	_, _, ch, victim := setupCascade(t, 30)
	ch.AffectedBy.Set(types.AFF_BERSERK)
	ch.PCData.Learned[gsnBerserk] = 33

	MultiHit(nil, ch, victim, types.TYPE_UNDEFINED)

	if *calls != 2 {
		t.Errorf("berserk Learned=33 pct=50: OneHit calls = %d, want 2 (primary + berserk)", *calls)
	}
}

// AFF_BERSERK but Learned[berserk]=0 → chance=0 → pct=50 is not < 0 → no fire.
func TestMultiHit_BerserkExtraHit_Learned0_NoFire(t *testing.T) {
	stubGsnsForCascade(t)
	calls := stubCascade(t, func() int { return 50 })
	_, _, ch, victim := setupCascade(t, 30)
	ch.AffectedBy.Set(types.AFF_BERSERK)
	// Learned[berserk] = 0 (default)

	MultiHit(nil, ch, victim, types.TYPE_UNDEFINED)

	if *calls != 1 {
		t.Errorf("berserk Learned=0: OneHit calls = %d, want 1 (primary only)", *calls)
	}
}

// NPC with AFF_BERSERK always rolls at 100%, so it always fires.
func TestMultiHit_BerserkNPCAlwaysFires(t *testing.T) {
	stubGsnsForCascade(t)
	// pct=99 — even for NPC the chance is 100, so 99 < 100 → fires.
	calls := stubCascade(t, func() int { return 99 })
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 10022, Name: "Berserk Arena"}
	w.Rooms[10022] = room

	ch := newFighter("BerserkNPC", 30)
	ch.Act.Set(types.ACT_IS_NPC)
	ch.AffectedBy.Set(types.AFF_BERSERK)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	victim := newFighter("Victim", 1)
	victim.Act.Set(types.ACT_IS_NPC)
	victim.Hit = 1_000_000
	victim.MaxHit = 1_000_000
	handler.CharToRoom(victim, room)
	w.AddChar(victim)
	StartFighting(ch, victim)

	MultiHit(w, ch, victim, types.TYPE_UNDEFINED)

	if *calls != 2 {
		t.Errorf("berserk NPC pct=99: calls = %d, want 2 (primary + berserk)", *calls)
	}
}

// --- G8 — Dual-wield weapon alternation in OneHit ---

// The dual-wield bonus swing uses the offhand weapon, not the primary.
// We assert this by inspecting which object each oneHit/oneHitOffhand
// call sees — the test replaces both seams with a spy that records the
// wield handed to each.
func TestOneHit_DualWieldAlternatesWeapons(t *testing.T) {
	stubGsnsForCascade(t)

	savedPct := numberPercent
	savedPrimary := oneHit
	savedOffhand := oneHitOffhand
	t.Cleanup(func() {
		numberPercent = savedPct
		oneHit = savedPrimary
		oneHitOffhand = savedOffhand
	})
	numberPercent = func() int { return 0 }

	// Capture which weapon each seam saw.
	var primaryWields []*types.ObjData
	var offhandWields []*types.ObjData
	oneHit = func(w *world.World, ch, victim *types.CharData, dt int) int {
		primaryWields = append(primaryWields, handler.GetEqChar(ch, types.WEAR_WIELD))
		return rNONE
	}
	oneHitOffhand = func(w *world.World, ch, victim *types.CharData, dt int) int {
		offhandWields = append(offhandWields, handler.GetEqChar(ch, types.WEAR_DUAL_WIELD))
		return rNONE
	}

	// Build a PC with both weapons.
	_, _, ch, victim := setupCascade(t, 50)
	ch.PCData.Learned[gsnDualWield] = 100

	mainWeapon := &types.ObjData{
		Name:     "mainhand sword",
		WearLoc:  types.WEAR_WIELD,
		ItemType: types.ITEM_WEAPON,
	}
	offhandWeapon := &types.ObjData{
		Name:     "offhand dagger",
		WearLoc:  types.WEAR_DUAL_WIELD,
		ItemType: types.ITEM_WEAPON,
	}
	handler.ObjToChar(mainWeapon, ch)
	handler.ObjToChar(offhandWeapon, ch)

	MultiHit(nil, ch, victim, types.TYPE_UNDEFINED)

	if len(primaryWields) == 0 {
		t.Fatal("no primary oneHit call seen")
	}
	if primaryWields[0] != mainWeapon {
		t.Errorf("primary wield = %v, want mainhand sword", primaryWields[0])
	}
	if len(offhandWields) != 1 {
		t.Fatalf("offhand calls = %d, want 1 (dual-wield bonus swing)", len(offhandWields))
	}
	if offhandWields[0] != offhandWeapon {
		t.Errorf("offhand wield = %v, want offhand dagger", offhandWields[0])
	}
}

// oneHitFull called directly with a chosen wield uses that wield's
// damage dice (not the primary).
func TestOneHitFull_ExplicitWieldUsed(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 10300, Name: "Dual Arena"}
	w.Rooms[10300] = room

	ch := newFighter("DualWielder", 50)
	ch.Hitroll = 999 // always hit
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	// Equip a primary sword with 1..1 dmg and an offhand dagger with 20..20.
	primary := &types.ObjData{
		Name: "sword", WearLoc: types.WEAR_WIELD, ItemType: types.ITEM_WEAPON,
	}
	primary.Value[1] = 1
	primary.Value[2] = 1
	offhand := &types.ObjData{
		Name: "heavy dagger", WearLoc: types.WEAR_DUAL_WIELD, ItemType: types.ITEM_WEAPON,
	}
	offhand.Value[1] = 20
	offhand.Value[2] = 20
	handler.ObjToChar(primary, ch)
	handler.ObjToChar(offhand, ch)

	victim := newFighter("Victim", 1)
	victim.Act.Set(types.ACT_IS_NPC)
	victim.Hit = 1_000_000
	victim.MaxHit = 1_000_000
	victim.Armor = 9999 // poor AC
	handler.CharToRoom(victim, room)
	w.AddChar(victim)
	StartFighting(ch, victim)

	// Swinging with the offhand directly should deal damage in the 20..20
	// range, not 1..1. We track the HP delta for one explicit-offhand
	// call vs one primary call.
	startHP := victim.Hit
	oneHitFull(w, ch, victim, types.TYPE_UNDEFINED, offhand)
	offhandDam := startHP - victim.Hit

	startHP = victim.Hit
	oneHitFull(w, ch, victim, types.TYPE_UNDEFINED, primary)
	primaryDam := startHP - victim.Hit

	// Damroll/str bonuses are identical, but weapon dice differ by ~19.
	if offhandDam <= primaryDam {
		t.Errorf("offhand dam=%d should exceed primary dam=%d (20-dice vs 1-dice)", offhandDam, primaryDam)
	}
}

// NPC NumAttacks multi-attack now lives inside MultiHit.
func TestMultiHit_NPCNumAttacks(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 9105, Name: "Arena"}
	w.Rooms[9105] = room

	ch := newFighter("MultiMob", 50)
	ch.Act.Set(types.ACT_IS_NPC)
	ch.MobThac0 = 0
	ch.Hitroll = 50
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

	MultiHit(w, ch, victim, types.TYPE_UNDEFINED)
	if victim.Hit >= 50000 {
		t.Error("NPC with NumAttacks=3 should deal damage via MultiHit")
	}
}

// --- Timer subsystem integration (G2-G4) ---

// ViolenceUpdate decrements per-char timers on every character each
// pulse, not just fighters. Two non-fighting chars each with a timer
// Count=3 must both drop to Count=2 after one call.
func TestViolenceUpdate_DecrementsEveryChar(t *testing.T) {
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 10200, Name: "Timer Field"}
	w.Rooms[10200] = room

	a := newFighter("Alpha", 10)
	handler.CharToRoom(a, room)
	w.AddChar(a)
	b := newFighter("Beta", 10)
	handler.CharToRoom(b, room)
	w.AddChar(b)

	handler.AddTimer(a, types.TIMER_RECENTFIGHT, 3, "", 0)
	handler.AddTimer(b, types.TIMER_RECENTFIGHT, 3, "", 0)

	ViolenceUpdate(w)

	if got := handler.GetTimer(a, types.TIMER_RECENTFIGHT); got != 2 {
		t.Errorf("Alpha TIMER_RECENTFIGHT = %d, want 2", got)
	}
	if got := handler.GetTimer(b, types.TIMER_RECENTFIGHT); got != 2 {
		t.Errorf("Beta TIMER_RECENTFIGHT = %d, want 2", got)
	}
}

// PC-vs-PC MultiHit sets TIMER_RECENTFIGHT=11 on BOTH attacker and
// victim (C fight.c:982-986).
func TestMultiHit_PCvsPCSetsRecentFightOnBoth(t *testing.T) {
	stubGsnsForCascade(t)
	stubCascade(t, func() int { return 0 })
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 10201, Name: "PvP Arena 2"}
	w.Rooms[10201] = room

	ch := newPCFighter(30)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)
	victim := newPCFighter(30)
	handler.CharToRoom(victim, room)
	w.AddChar(victim)
	StartFighting(ch, victim)

	MultiHit(w, ch, victim, types.TYPE_UNDEFINED)

	if got := handler.GetTimer(ch, types.TIMER_RECENTFIGHT); got != 11 {
		t.Errorf("attacker TIMER_RECENTFIGHT = %d, want 11", got)
	}
	if got := handler.GetTimer(victim, types.TIMER_RECENTFIGHT); got != 11 {
		t.Errorf("victim TIMER_RECENTFIGHT = %d, want 11", got)
	}
}

// PC attacking an NPC does NOT set TIMER_RECENTFIGHT on either side —
// the rule only applies to mutual PC-vs-PC combat.
func TestMultiHit_PCvsNPCDoesNotSetRecentFight(t *testing.T) {
	stubGsnsForCascade(t)
	stubCascade(t, func() int { return 0 })
	_, _, ch, victim := setupCascade(t, 30)

	MultiHit(nil, ch, victim, types.TYPE_UNDEFINED)

	if got := handler.GetTimer(ch, types.TIMER_RECENTFIGHT); got != 0 {
		t.Errorf("PC attacker TIMER_RECENTFIGHT = %d, want 0", got)
	}
	if got := handler.GetTimer(victim, types.TIMER_RECENTFIGHT); got != 0 {
		t.Errorf("NPC victim TIMER_RECENTFIGHT = %d, want 0", got)
	}
}

// PLR_NICE short-circuits before the timer is set; no timer on either
// side, and MultiHit returns rNONE.
func TestMultiHit_PLR_NICE_NoTimer(t *testing.T) {
	stubGsnsForCascade(t)
	stubCascade(t, func() int { return 0 })
	w := newCombatWorld()
	room := &types.RoomIndexData{Vnum: 10202, Name: "Nice Arena"}
	w.Rooms[10202] = room

	ch := newPCFighter(30)
	ch.Act.Set(types.PLR_NICE)
	handler.CharToRoom(ch, room)
	w.AddChar(ch)
	victim := newPCFighter(30)
	handler.CharToRoom(victim, room)
	w.AddChar(victim)
	StartFighting(ch, victim)

	if ret := MultiHit(w, ch, victim, types.TYPE_UNDEFINED); ret != rNONE {
		t.Errorf("PLR_NICE PC-vs-PC returned %d, want rNONE", ret)
	}
	if got := handler.GetTimer(ch, types.TIMER_RECENTFIGHT); got != 0 {
		t.Errorf("PLR_NICE attacker TIMER_RECENTFIGHT = %d, want 0", got)
	}
	if got := handler.GetTimer(victim, types.TIMER_RECENTFIGHT); got != 0 {
		t.Errorf("PLR_NICE victim TIMER_RECENTFIGHT = %d, want 0", got)
	}
}

// IsAttackSuppressed: a permanent (Value == -1) TIMER_ASUPRESSED is
// active regardless of Count.
func TestIsAttackSuppressed_PermanentValueMinus1(t *testing.T) {
	ch := newFighter("Monk", 10)
	ch.Timers = append(ch.Timers, &types.TimerData{
		Type: types.TIMER_ASUPRESSED, Count: 0, Value: -1,
	})
	if !IsAttackSuppressed(ch) {
		t.Error("permanent TIMER_ASUPRESSED (Value=-1) should suppress")
	}
}

// IsAttackSuppressed: a timer with Count=0 and Value=0 does not
// suppress (it should have been removed, but defensively check).
func TestIsAttackSuppressed_ZeroCount(t *testing.T) {
	ch := newFighter("Monk", 10)
	ch.Timers = append(ch.Timers, &types.TimerData{
		Type: types.TIMER_ASUPRESSED, Count: 0, Value: 0,
	})
	if IsAttackSuppressed(ch) {
		t.Error("Count=0 Value=0 TIMER_ASUPRESSED should not suppress")
	}
}

// IsAttackSuppressed: a positive-count non-permanent timer suppresses.
func TestIsAttackSuppressed_PositiveCount(t *testing.T) {
	ch := newFighter("Monk", 10)
	ch.Timers = append(ch.Timers, &types.TimerData{
		Type: types.TIMER_ASUPRESSED, Count: 3, Value: 0,
	})
	if !IsAttackSuppressed(ch) {
		t.Error("Count=3 TIMER_ASUPRESSED should suppress")
	}
}

// Count=1 boundary — guards the `>= 1` (not `> 1`) semantic; adversary-demonstrated coverage gap.
func TestIsAttackSuppressed_CountOneBoundary(t *testing.T) {
	ch := newFighter("Monk", 10)
	handler.AddTimer(ch, types.TIMER_ASUPRESSED, 1, "", 0)
	if !IsAttackSuppressed(ch) {
		t.Error("Count=1 TIMER_ASUPRESSED should suppress (>= 1 boundary)")
	}
}
