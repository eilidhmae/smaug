package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// --- G1: room-flag enforcement ---

func TestDoCast_BlockedByNoMagic(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 9001, Name: "Silenced"}
	room.RoomFlags.Set(types.ROOM_NO_MAGIC)

	ch, client := makeTestChar("Gandalf")
	defer client.Close()
	handler.CharToRoom(ch, room)

	DoCast(ch, "magic missile")
	out := readOutput(ch, client)
	if !strings.Contains(out, "failed") {
		t.Errorf("expected magic failure message; got %q", out)
	}
}

func TestDoRecall_BlockedByNoRecall(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	temple := &types.RoomIndexData{Vnum: types.ROOM_VNUM_TEMPLE, Name: "Temple"}
	w.Rooms[temple.Vnum] = temple
	room := &types.RoomIndexData{Vnum: 9002, Name: "No Recall"}
	room.RoomFlags.Set(types.ROOM_NO_RECALL)

	ch, client := makeTestChar("Gandalf")
	defer client.Close()
	handler.CharToRoom(ch, room)

	DoRecall(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "failed") {
		t.Errorf("expected recall failure; got %q", out)
	}
	if ch.InRoom != room {
		t.Errorf("char moved off no-recall room")
	}
}

func TestDoSay_BlockedBySilence(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 9003, Name: "Silent"}
	room.RoomFlags.Set(types.ROOM_SILENCE)

	ch, client := makeTestChar("Mute")
	defer client.Close()
	handler.CharToRoom(ch, room)

	DoSay(ch, "hello")
	out := readOutput(ch, client)
	if !strings.Contains(out, "can't") {
		t.Errorf("expected silence refusal; got %q", out)
	}
	if strings.Contains(out, "You say") {
		t.Errorf("say should not echo in silenced room; got %q", out)
	}
}

func TestDoDrop_BlockedByNoDrop(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 9004, Name: "Sticky"}
	room.RoomFlags.Set(types.ROOM_NODROP)

	ch, client := makeTestChar("Dropper")
	defer client.Close()
	handler.CharToRoom(ch, room)
	obj := &types.ObjData{Name: "sword", ShortDescr: "a sword", WearLoc: types.WEAR_NONE}
	handler.ObjToChar(obj, ch)

	DoDrop(ch, "sword")
	out := readOutput(ch, client)
	if !strings.Contains(out, "magical force") {
		t.Errorf("expected magical-force message; got %q", out)
	}
	if obj.CarriedBy != ch {
		t.Errorf("object should still be carried")
	}
}

func TestMoveChar_SolitaryBlocksSecondOccupant(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	start := &types.RoomIndexData{Vnum: 9005, Name: "Start"}
	dest := &types.RoomIndexData{Vnum: 9006, Name: "Solitary"}
	dest.RoomFlags.Set(types.ROOM_SOLITARY)
	start.Exits = []*types.ExitData{{Direction: types.DIR_NORTH, ToRoom: dest}}
	dest.Exits = []*types.ExitData{{Direction: types.DIR_SOUTH, ToRoom: start}}

	hermit, hc := makeTestChar("Hermit")
	defer hc.Close()
	handler.CharToRoom(hermit, dest)

	visitor, vc := makeTestChar("Visitor")
	defer vc.Close()
	handler.CharToRoom(visitor, start)

	MoveChar(visitor, types.DIR_NORTH)
	out := readOutput(visitor, vc)
	if visitor.InRoom != start {
		t.Errorf("visitor should be blocked; still at %v", visitor.InRoom.Vnum)
	}
	if !strings.Contains(out, "private") {
		t.Errorf("expected private-room message; got %q", out)
	}
}

func TestMoveChar_PrivateBlocksThirdOccupant(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	start := &types.RoomIndexData{Vnum: 9020, Name: "Start"}
	dest := &types.RoomIndexData{Vnum: 9021, Name: "Private"}
	dest.RoomFlags.Set(types.ROOM_PRIVATE)
	start.Exits = []*types.ExitData{{Direction: types.DIR_NORTH, ToRoom: dest}}

	// Fill two existing occupants — a third should be refused.
	a, ac := makeTestChar("Ada")
	defer ac.Close()
	handler.CharToRoom(a, dest)
	b, bc := makeTestChar("Bea")
	defer bc.Close()
	handler.CharToRoom(b, dest)

	c, cc := makeTestChar("Cam")
	defer cc.Close()
	handler.CharToRoom(c, start)

	MoveChar(c, types.DIR_NORTH)
	out := readOutput(c, cc)
	if c.InRoom != start {
		t.Errorf("third occupant should be blocked from ROOM_PRIVATE")
	}
	if !strings.Contains(out, "private") {
		t.Errorf("expected private message; got %q", out)
	}
}

func TestMoveChar_NoFloorRequiresFlying(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	start := &types.RoomIndexData{Vnum: 9007, Name: "Ledge"}
	dest := &types.RoomIndexData{Vnum: 9008, Name: "Chasm"}
	dest.RoomFlags.Set(types.ROOM_NOFLOOR)
	start.Exits = []*types.ExitData{{Direction: types.DIR_DOWN, ToRoom: dest}}

	ch, client := makeTestChar("Faller")
	defer client.Close()
	handler.CharToRoom(ch, start)

	MoveChar(ch, types.DIR_DOWN)
	out := readOutput(ch, client)
	if !strings.Contains(out, "fly") {
		t.Errorf("expected fly message; got %q", out)
	}
	if ch.InRoom != start {
		t.Errorf("char should not have fallen without flying")
	}

	// Now with flying
	ch.AffectedBy.Set(types.AFF_FLYING)
	MoveChar(ch, types.DIR_DOWN)
	if ch.InRoom != dest {
		t.Errorf("flying char should descend into no-floor room")
	}
}

func TestMoveChar_NoMobBlocksNPC(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	start := &types.RoomIndexData{Vnum: 9009, Name: "Start"}
	dest := &types.RoomIndexData{Vnum: 9010, Name: "No Mob"}
	dest.RoomFlags.Set(types.ROOM_NO_MOB)
	start.Exits = []*types.ExitData{{Direction: types.DIR_NORTH, ToRoom: dest}}

	mob := &types.CharData{
		Name:      "Guard",
		Position:  types.POS_STANDING,
		IndexData: &types.MobIndexData{Vnum: 3001},
	}
	mob.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(mob, start)

	MoveChar(mob, types.DIR_NORTH)
	if mob.InRoom != start {
		t.Errorf("NPC should be blocked by ROOM_NO_MOB")
	}
}

// --- G2: item wear restrictions ---

func TestDoWear_BlockedByAntiEvil(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 9100}

	ch, client := makeTestChar("Dark")
	defer client.Close()
	ch.Alignment = -500
	handler.CharToRoom(ch, room)

	obj := &types.ObjData{
		Name: "holy", ShortDescr: "a holy sword", WearLoc: types.WEAR_NONE,
		ItemType: types.ITEM_WEAPON, WearFlags: int(types.ITEM_TAKE | types.ITEM_WIELD),
	}
	obj.ExtraFlags.Set(types.ITEM_ANTI_EVIL)
	handler.ObjToChar(obj, ch)

	DoWear(ch, "holy")
	out := readOutput(ch, client)
	if !strings.Contains(out, "evil") {
		t.Errorf("expected anti-evil rejection; got %q", out)
	}
	if obj.WearLoc != types.WEAR_NONE {
		t.Errorf("item should not be equipped")
	}
}

func TestDoWear_BlockedByAntiClass(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 9101}

	ch, client := makeTestChar("Mage")
	defer client.Close()
	ch.Class = types.CLASS_MAGE
	handler.CharToRoom(ch, room)

	obj := &types.ObjData{
		Name: "battle", ShortDescr: "a battleaxe", WearLoc: types.WEAR_NONE,
		ItemType: types.ITEM_WEAPON, WearFlags: int(types.ITEM_TAKE | types.ITEM_WIELD),
	}
	obj.ExtraFlags.Set(types.ITEM_ANTI_MAGE)
	handler.ObjToChar(obj, ch)

	DoWear(ch, "battle")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Mages") {
		t.Errorf("expected anti-class rejection; got %q", out)
	}
}

func TestDoWear_AntiGoodAllowsEvil(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 9102}

	ch, client := makeTestChar("Evil")
	defer client.Close()
	ch.Alignment = -500
	handler.CharToRoom(ch, room)

	obj := &types.ObjData{
		Name: "dark", ShortDescr: "a dark blade", WearLoc: types.WEAR_NONE,
		ItemType: types.ITEM_WEAPON, WearFlags: int(types.ITEM_TAKE | types.ITEM_WIELD),
	}
	obj.ExtraFlags.Set(types.ITEM_ANTI_GOOD)
	handler.ObjToChar(obj, ch)

	DoWear(ch, "dark")
	_ = readOutput(ch, client)
	if obj.WearLoc == types.WEAR_NONE {
		t.Errorf("evil char should be allowed to wield anti-good blade")
	}
}

// --- G3: exit secret/hidden/nomob ---

func TestShowExits_HidesSecret(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 9200, Name: "Chamber"}
	other := &types.RoomIndexData{Vnum: 9201, Name: "Other"}
	room.Exits = []*types.ExitData{
		{Direction: types.DIR_NORTH, ToRoom: other},
		{Direction: types.DIR_EAST, ToRoom: other, ExitInfo: int(types.EX_SECRET)},
	}

	ch, client := makeTestChar("Looker")
	defer client.Close()
	handler.CharToRoom(ch, room)

	showExits(ch, room)
	out := readOutput(ch, client)
	if !strings.Contains(out, "north") {
		t.Errorf("normal exit missing; got %q", out)
	}
	if strings.Contains(out, "east") {
		t.Errorf("secret exit should be hidden; got %q", out)
	}

	// Holylight reveals secret exits.
	ch.Act.Set(types.PLR_HOLYLIGHT)
	showExits(ch, room)
	out = readOutput(ch, client)
	if !strings.Contains(out, "east") {
		t.Errorf("holylight should reveal secret exits; got %q", out)
	}
}

// --- G4: player-flag comm filters ---

func TestDoTell_RefusedByNoTell(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 9300}

	sender, sc := makeTestChar("Alice")
	defer sc.Close()
	handler.CharToRoom(sender, room)

	recipient, rc := makeTestChar("Bob")
	defer rc.Close()
	recipient.Act.Set(types.PLR_NO_TELL)
	handler.CharToRoom(recipient, room)
	w.Characters = []*types.CharData{sender, recipient}

	DoTell(sender, "bob hi")
	out := readOutput(sender, sc)
	if !strings.Contains(out, "refusing tells") {
		t.Errorf("expected refusal message; got %q", out)
	}
}

func TestDoTell_AFKPrefix(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 9301}

	sender, sc := makeTestChar("Alice")
	defer sc.Close()
	handler.CharToRoom(sender, room)

	recipient, rc := makeTestChar("Bob")
	defer rc.Close()
	recipient.Act.Set(types.PLR_AFK)
	handler.CharToRoom(recipient, room)
	w.Characters = []*types.CharData{sender, recipient}

	DoTell(sender, "bob hi")
	_ = readOutput(sender, sc)
	bobOut := readOutput(recipient, rc)
	if !strings.Contains(bobOut, "(afk)") {
		t.Errorf("expected (afk) prefix; got %q", bobOut)
	}
}

// --- G5: languages ---

func TestDoSpeak_SetsSpeakingIfKnown(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 9400}

	ch, client := makeTestChar("Linguist")
	defer client.Close()
	handler.CharToRoom(ch, room)
	ch.Speaks = int(types.LANG_COMMON | types.LANG_ELVEN)

	DoSpeak(ch, "elven")
	out := readOutput(ch, client)
	if uint32(ch.Speaking) != types.LANG_ELVEN {
		t.Errorf("Speaking should be LANG_ELVEN; got %d", ch.Speaking)
	}
	if !strings.Contains(out, "elven") {
		t.Errorf("expected confirmation; got %q", out)
	}
}

func TestDoSpeak_UnknownLanguageRejected(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 9401}

	ch, client := makeTestChar("Monoglot")
	defer client.Close()
	handler.CharToRoom(ch, room)
	ch.Speaks = int(types.LANG_COMMON)

	DoSpeak(ch, "dwarven")
	out := readOutput(ch, client)
	if ch.Speaking == int(types.LANG_DWARVEN) {
		t.Errorf("should not be able to speak unknown language")
	}
	if !strings.Contains(out, "do not know") {
		t.Errorf("expected refusal; got %q", out)
	}
}

func TestDoSay_ScramblesForListenerWithoutLanguage(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 9402}

	speaker, sc := makeTestChar("Elf")
	defer sc.Close()
	speaker.Speaks = int(types.LANG_COMMON | types.LANG_ELVEN)
	speaker.Speaking = int(types.LANG_ELVEN)
	handler.CharToRoom(speaker, room)

	listener, lc := makeTestChar("Dwarf")
	defer lc.Close()
	listener.Speaks = int(types.LANG_COMMON | types.LANG_DWARVEN)
	handler.CharToRoom(listener, room)

	DoSay(speaker, "hello friend")
	_ = readOutput(speaker, sc)
	lOut := readOutput(listener, lc)
	if strings.Contains(lOut, "hello friend") {
		t.Errorf("listener should hear scrambled text; got %q", lOut)
	}

	// Listener who shares language gets clear text.
	listener.Speaks = int(types.LANG_COMMON | types.LANG_ELVEN)
	DoSay(speaker, "hello friend")
	_ = readOutput(speaker, sc)
	lOut = readOutput(listener, lc)
	if !strings.Contains(lOut, "hello friend") {
		t.Errorf("sharing language should preserve text; got %q", lOut)
	}
}

// --- Adversary-follow-up coverage tests ---

func TestRoomSuppressesMagic_AreaLevelNoMagic(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	area := &types.AreaData{Name: "Silent", Flags: int(types.AFLAG_NOMAGIC)}
	room := &types.RoomIndexData{Vnum: 9050, Area: area}

	ch, client := makeTestChar("Mage")
	defer client.Close()
	handler.CharToRoom(ch, room)

	DoCast(ch, "magic missile")
	out := readOutput(ch, client)
	if !strings.Contains(out, "failed") {
		t.Errorf("area AFLAG_NOMAGIC should suppress DoCast; got %q", out)
	}
}

func TestDoQuaff_NoMagicRoomDoesNotConsumePotion(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 9051}
	room.RoomFlags.Set(types.ROOM_NO_MAGIC)

	ch, client := makeTestChar("Priest")
	defer client.Close()
	handler.CharToRoom(ch, room)

	potion := &types.ObjData{
		Name: "potion", ShortDescr: "a potion",
		ItemType: types.ITEM_POTION, WearLoc: types.WEAR_NONE,
	}
	handler.ObjToChar(potion, ch)

	DoQuaff(ch, "potion")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Nothing seems to happen") {
		t.Errorf("expected magic-suppressed message; got %q", out)
	}
	if potion.CarriedBy != ch {
		t.Errorf("potion should not be consumed in no-magic room")
	}
}

func TestDoBrandish_NoMagicRoomDoesNotDecrementCharge(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 9052}
	room.RoomFlags.Set(types.ROOM_NO_MAGIC)

	ch, client := makeTestChar("Wizard")
	defer client.Close()
	handler.CharToRoom(ch, room)

	staff := &types.ObjData{
		Name: "staff", ShortDescr: "a staff",
		ItemType: types.ITEM_STAFF, WearLoc: types.WEAR_HOLD,
	}
	staff.Value[2] = 3
	handler.ObjToChar(staff, ch)

	DoBrandish(ch, "")
	_ = readOutput(ch, client)
	if staff.Value[2] != 3 {
		t.Errorf("charge should not drop in no-magic room; charges = %d", staff.Value[2])
	}
}

func TestApplyRoomDeath_MovesPlayerToTemple(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	temple := &types.RoomIndexData{Vnum: types.ROOM_VNUM_TEMPLE, Name: "Temple"}
	w.Rooms[temple.Vnum] = temple

	start := &types.RoomIndexData{Vnum: 9060, Name: "Ledge"}
	trap := &types.RoomIndexData{Vnum: 9061, Name: "Pit"}
	trap.RoomFlags.Set(types.ROOM_DEATH)
	start.Exits = []*types.ExitData{{Direction: types.DIR_NORTH, ToRoom: trap}}

	ch, client := makeTestChar("Victim")
	defer client.Close()
	ch.Hit = 100
	handler.CharToRoom(ch, start)

	MoveChar(ch, types.DIR_NORTH)
	_ = readOutput(ch, client)

	if ch.InRoom != temple {
		t.Errorf("player should be teleported to temple; in %+v", ch.InRoom)
	}
	if ch.Hit != 1 {
		t.Errorf("player HP should be reset to 1; got %d", ch.Hit)
	}
	if ch.Position != types.POS_RESTING {
		t.Errorf("player should be resting; got %d", ch.Position)
	}
}

func TestApplyRoomDeath_NPCExempt(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	temple := &types.RoomIndexData{Vnum: types.ROOM_VNUM_TEMPLE, Name: "Temple"}
	w.Rooms[temple.Vnum] = temple

	start := &types.RoomIndexData{Vnum: 9062}
	trap := &types.RoomIndexData{Vnum: 9063}
	trap.RoomFlags.Set(types.ROOM_DEATH)
	start.Exits = []*types.ExitData{{Direction: types.DIR_NORTH, ToRoom: trap}}

	mob := &types.CharData{
		Name: "Wanderer", Position: types.POS_STANDING, Hit: 50,
		IndexData: &types.MobIndexData{Vnum: 3001},
	}
	mob.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(mob, start)

	MoveChar(mob, types.DIR_NORTH)

	if mob.InRoom != trap {
		t.Errorf("NPC should proceed into ROOM_DEATH without being killed; in %+v", mob.InRoom)
	}
	if mob.Hit != 50 {
		t.Errorf("NPC HP should be unchanged; got %d", mob.Hit)
	}
}

func TestDoEmote_SkippedByNoEmote(t *testing.T) {
	w := world.New("/tmp/test")
	WorldRef = w
	room := &types.RoomIndexData{Vnum: 9302}

	sender, sc := makeTestChar("Alice")
	defer sc.Close()
	handler.CharToRoom(sender, room)

	recipient, rc := makeTestChar("Bob")
	defer rc.Close()
	recipient.Act.Set(types.PLR_NO_EMOTE)
	handler.CharToRoom(recipient, room)

	DoEmote(sender, "smiles")
	_ = readOutput(sender, sc)
	if recipient.Desc.HasOutput() {
		bobOut := readOutput(recipient, rc)
		if strings.Contains(bobOut, "smiles") {
			t.Errorf("no-emote player should not receive emote; got %q", bobOut)
		}
	}
}
