package act

import (
	"net"
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func makeObjTestChar(name string) (*types.CharData, net.Conn) {
	ch, client := makeTestChar(name)
	ch.Gold = 0
	return ch, client
}

func setupObjWorld() *world.World {
	w := world.New("/tmp/test")
	WorldRef = w
	return w
}

func TestDoGet_FromRoom(t *testing.T) {
	w := setupObjWorld()
	room := &types.RoomIndexData{Vnum: 5100, Name: "Test Room"}
	w.Rooms[5100] = room
	ch, client := makeObjTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	idx := &types.ObjIndexData{Vnum: 3000, Name: "gold coin", ShortDescr: "a gold coin", WearFlags: int(types.ITEM_TAKE)}
	w.ObjIndex[3000] = idx
	obj := handler.CreateObject(w, idx, 1)
	handler.ObjToRoom(obj, room)

	DoGet(ch, "coin")
	output := readOutput(ch, client)

	if len(ch.Carrying) != 1 {
		t.Errorf("ch.Carrying = %d, want 1", len(ch.Carrying))
	}
	if len(room.Contents) != 0 {
		t.Errorf("room.Contents = %d, want 0", len(room.Contents))
	}
	if !strings.Contains(output, "gold coin") {
		t.Errorf("output should mention item: %q", output)
	}
}

func TestDoGet_NotFound(t *testing.T) {
	_ = setupObjWorld()
	room := &types.RoomIndexData{Vnum: 5101, Name: "Test Room"}
	ch, client := makeObjTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	DoGet(ch, "sword")
	output := readOutput(ch, client)

	if !strings.Contains(strings.ToLower(output), "see") || !strings.Contains(strings.ToLower(output), "here") {
		t.Errorf("expected 'not found' message, got: %q", output)
	}
}

func TestDoGet_FromContainer(t *testing.T) {
	w := setupObjWorld()
	room := &types.RoomIndexData{Vnum: 5102, Name: "Test Room"}
	w.Rooms[5102] = room
	ch, client := makeObjTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	cIdx := &types.ObjIndexData{Vnum: 3001, Name: "chest", ShortDescr: "a chest",
		ItemType: types.ITEM_CONTAINER, WearFlags: int(types.ITEM_TAKE)}
	w.ObjIndex[3001] = cIdx
	container := handler.CreateObject(w, cIdx, 1)
	handler.ObjToRoom(container, room)

	iIdx := &types.ObjIndexData{Vnum: 3002, Name: "ruby gem", ShortDescr: "a ruby",
		WearFlags: int(types.ITEM_TAKE)}
	w.ObjIndex[3002] = iIdx
	item := handler.CreateObject(w, iIdx, 1)
	handler.ObjToObj(item, container)

	DoGet(ch, "ruby chest")
	_ = readOutput(ch, client)

	if len(ch.Carrying) != 1 {
		t.Errorf("ch.Carrying = %d, want 1", len(ch.Carrying))
	}
	if len(container.Contents) != 0 {
		t.Errorf("container.Contents = %d, want 0", len(container.Contents))
	}
}

func TestDoGet_NoArg(t *testing.T) {
	_ = setupObjWorld()
	room := &types.RoomIndexData{Vnum: 5103, Name: "Test Room"}
	ch, client := makeObjTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	DoGet(ch, "")
	output := readOutput(ch, client)

	if !strings.Contains(strings.ToLower(output), "get what") {
		t.Errorf("expected 'get what' message, got: %q", output)
	}
}

func TestDoDrop(t *testing.T) {
	w := setupObjWorld()
	room := &types.RoomIndexData{Vnum: 5110, Name: "Test Room"}
	w.Rooms[5110] = room
	ch, client := makeObjTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	idx := &types.ObjIndexData{Vnum: 3003, Name: "iron sword", ShortDescr: "an iron sword",
		WearFlags: int(types.ITEM_TAKE)}
	w.ObjIndex[3003] = idx
	obj := handler.CreateObject(w, idx, 1)
	handler.ObjToChar(obj, ch)

	DoDrop(ch, "sword")
	_ = readOutput(ch, client)

	if len(ch.Carrying) != 0 {
		t.Errorf("ch.Carrying = %d, want 0", len(ch.Carrying))
	}
	if len(room.Contents) != 1 {
		t.Errorf("room.Contents = %d, want 1", len(room.Contents))
	}
}

func TestDoDrop_NoDrop(t *testing.T) {
	w := setupObjWorld()
	room := &types.RoomIndexData{Vnum: 5111, Name: "Test Room"}
	w.Rooms[5111] = room
	ch, client := makeObjTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	idx := &types.ObjIndexData{Vnum: 3004, Name: "cursed ring", ShortDescr: "a cursed ring",
		WearFlags: int(types.ITEM_TAKE)}
	w.ObjIndex[3004] = idx
	obj := handler.CreateObject(w, idx, 1)
	obj.ExtraFlags.Set(types.ITEM_NODROP)
	handler.ObjToChar(obj, ch)

	DoDrop(ch, "ring")
	_ = readOutput(ch, client)

	if len(ch.Carrying) != 1 {
		t.Errorf("ch.Carrying = %d, want 1 (nodrop item)", len(ch.Carrying))
	}
}

func TestDoWear(t *testing.T) {
	w := setupObjWorld()
	room := &types.RoomIndexData{Vnum: 5112, Name: "Test Room"}
	w.Rooms[5112] = room
	ch, client := makeObjTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	idx := &types.ObjIndexData{Vnum: 3005, Name: "iron helmet", ShortDescr: "an iron helmet",
		WearFlags: int(types.ITEM_TAKE | types.ITEM_WEAR_HEAD)}
	w.ObjIndex[3005] = idx
	obj := handler.CreateObject(w, idx, 1)
	handler.ObjToChar(obj, ch)

	DoWear(ch, "helmet")
	_ = readOutput(ch, client)

	if obj.WearLoc != types.WEAR_HEAD {
		t.Errorf("WearLoc = %d, want WEAR_HEAD (%d)", obj.WearLoc, types.WEAR_HEAD)
	}
}

func TestDoWear_Weapon(t *testing.T) {
	w := setupObjWorld()
	room := &types.RoomIndexData{Vnum: 5113, Name: "Test Room"}
	w.Rooms[5113] = room
	ch, client := makeObjTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	idx := &types.ObjIndexData{Vnum: 3011, Name: "steel sword", ShortDescr: "a steel sword",
		ItemType: types.ITEM_WEAPON, WearFlags: int(types.ITEM_TAKE | types.ITEM_WIELD)}
	w.ObjIndex[3011] = idx
	obj := handler.CreateObject(w, idx, 1)
	handler.ObjToChar(obj, ch)

	DoWear(ch, "sword")
	_ = readOutput(ch, client)

	if obj.WearLoc != types.WEAR_WIELD {
		t.Errorf("WearLoc = %d, want WEAR_WIELD (%d)", obj.WearLoc, types.WEAR_WIELD)
	}
}

func TestDoRemove(t *testing.T) {
	w := setupObjWorld()
	room := &types.RoomIndexData{Vnum: 5114, Name: "Test Room"}
	w.Rooms[5114] = room
	ch, client := makeObjTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	idx := &types.ObjIndexData{Vnum: 3006, Name: "iron helmet", ShortDescr: "an iron helmet",
		WearFlags: int(types.ITEM_TAKE | types.ITEM_WEAR_HEAD)}
	w.ObjIndex[3006] = idx
	obj := handler.CreateObject(w, idx, 1)
	handler.EquipChar(ch, obj, types.WEAR_HEAD)

	DoRemove(ch, "helmet")
	_ = readOutput(ch, client)

	if obj.WearLoc != types.WEAR_NONE {
		t.Errorf("WearLoc = %d, want WEAR_NONE", obj.WearLoc)
	}
}

func TestDoPut(t *testing.T) {
	w := setupObjWorld()
	room := &types.RoomIndexData{Vnum: 5115, Name: "Test Room"}
	w.Rooms[5115] = room
	ch, client := makeObjTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	cIdx := &types.ObjIndexData{Vnum: 3007, Name: "leather bag", ShortDescr: "a leather bag",
		ItemType: types.ITEM_CONTAINER, WearFlags: int(types.ITEM_TAKE)}
	w.ObjIndex[3007] = cIdx
	container := handler.CreateObject(w, cIdx, 1)
	handler.ObjToChar(container, ch)

	iIdx := &types.ObjIndexData{Vnum: 3008, Name: "red potion", ShortDescr: "a red potion",
		WearFlags: int(types.ITEM_TAKE)}
	w.ObjIndex[3008] = iIdx
	item := handler.CreateObject(w, iIdx, 1)
	handler.ObjToChar(item, ch)

	DoPut(ch, "potion bag")
	_ = readOutput(ch, client)

	if item.InObj != container {
		t.Error("item should be inside container")
	}
}

func TestDoGive(t *testing.T) {
	w := setupObjWorld()
	room := &types.RoomIndexData{Vnum: 5116, Name: "Test Room"}
	w.Rooms[5116] = room
	ch, client := makeObjTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	victim := &types.CharData{Name: "guard soldier", Position: types.POS_STANDING}
	victim.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(victim, room)

	idx := &types.ObjIndexData{Vnum: 3009, Name: "gold coin", ShortDescr: "a gold coin",
		WearFlags: int(types.ITEM_TAKE)}
	w.ObjIndex[3009] = idx
	obj := handler.CreateObject(w, idx, 1)
	handler.ObjToChar(obj, ch)

	DoGive(ch, "coin guard")
	_ = readOutput(ch, client)

	if len(ch.Carrying) != 0 {
		t.Errorf("ch.Carrying = %d, want 0", len(ch.Carrying))
	}
	if len(victim.Carrying) != 1 {
		t.Errorf("victim.Carrying = %d, want 1", len(victim.Carrying))
	}
}

func TestDoSacrifice(t *testing.T) {
	w := setupObjWorld()
	room := &types.RoomIndexData{Vnum: 5117, Name: "Test Room"}
	w.Rooms[5117] = room
	ch, client := makeObjTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	idx := &types.ObjIndexData{Vnum: 3010, Name: "broken sword", ShortDescr: "a broken sword",
		WearFlags: int(types.ITEM_TAKE)}
	w.ObjIndex[3010] = idx
	handler.CreateObject(w, idx, 1)
	handler.ObjToRoom(w.Objects[len(w.Objects)-1], room)

	goldBefore := ch.Gold
	DoSacrifice(ch, "sword")
	_ = readOutput(ch, client)

	if len(room.Contents) != 0 {
		t.Errorf("room.Contents = %d, want 0", len(room.Contents))
	}
	if ch.Gold != goldBefore+1 {
		t.Errorf("Gold = %d, want %d", ch.Gold, goldBefore+1)
	}
}

// --- Tier 2: flag-enforcement tests ---

func TestDoDrop_RoomNoDrop(t *testing.T) {
	w := setupObjWorld()
	room := &types.RoomIndexData{Vnum: 5200, Name: "No Drop"}
	room.RoomFlags.Set(types.ROOM_NODROP)
	w.Rooms[5200] = room

	ch, client := makeObjTestChar("Tester")
	defer client.Close()
	handler.CharToRoom(ch, room)

	idx := &types.ObjIndexData{Vnum: 3100, Name: "iron sword", ShortDescr: "an iron sword",
		WearFlags: int(types.ITEM_TAKE)}
	w.ObjIndex[3100] = idx
	obj := handler.CreateObject(w, idx, 1)
	handler.ObjToChar(obj, ch)

	DoDrop(ch, "sword")
	out := readOutput(ch, client)

	if len(ch.Carrying) != 1 {
		t.Errorf("ROOM_NODROP should keep item; Carrying=%d", len(ch.Carrying))
	}
	if len(room.Contents) != 0 {
		t.Errorf("ROOM_NODROP should leave room empty; Contents=%d", len(room.Contents))
	}
	if !strings.Contains(out, "magical force") {
		t.Errorf("expected NODROP message, got: %q", out)
	}
}

func TestDoWear_AntiEvilBlocksEvil(t *testing.T) {
	w := setupObjWorld()
	room := &types.RoomIndexData{Vnum: 5201, Name: "Room"}
	w.Rooms[5201] = room

	ch, client := makeObjTestChar("Evildoer")
	defer client.Close()
	ch.Alignment = -800 // Evil
	handler.CharToRoom(ch, room)

	idx := &types.ObjIndexData{Vnum: 3101, Name: "holy helm", ShortDescr: "a holy helm",
		WearFlags: int(types.ITEM_TAKE | types.ITEM_WEAR_HEAD)}
	w.ObjIndex[3101] = idx
	obj := handler.CreateObject(w, idx, 1)
	obj.ExtraFlags.Set(types.ITEM_ANTI_EVIL)
	handler.ObjToChar(obj, ch)

	DoWear(ch, "helm")
	out := readOutput(ch, client)

	if obj.WearLoc != types.WEAR_NONE {
		t.Errorf("ANTI_EVIL should block evil wearer; WearLoc=%d", obj.WearLoc)
	}
	if !strings.Contains(out, "too evil") {
		t.Errorf("expected 'too evil' refusal, got: %q", out)
	}
}

func TestDoWear_AntiGoodBlocksGood(t *testing.T) {
	w := setupObjWorld()
	room := &types.RoomIndexData{Vnum: 5202, Name: "Room"}
	w.Rooms[5202] = room

	ch, client := makeObjTestChar("Paladin")
	defer client.Close()
	ch.Alignment = 800 // Good
	handler.CharToRoom(ch, room)

	idx := &types.ObjIndexData{Vnum: 3102, Name: "evil helm", ShortDescr: "an evil helm",
		WearFlags: int(types.ITEM_TAKE | types.ITEM_WEAR_HEAD)}
	w.ObjIndex[3102] = idx
	obj := handler.CreateObject(w, idx, 1)
	obj.ExtraFlags.Set(types.ITEM_ANTI_GOOD)
	handler.ObjToChar(obj, ch)

	DoWear(ch, "helm")
	out := readOutput(ch, client)
	if obj.WearLoc != types.WEAR_NONE {
		t.Errorf("ANTI_GOOD should block good wearer")
	}
	if !strings.Contains(out, "too good") {
		t.Errorf("expected 'too good' refusal, got: %q", out)
	}
}

func TestDoWear_AntiClassBlocksMage(t *testing.T) {
	w := setupObjWorld()
	room := &types.RoomIndexData{Vnum: 5203, Name: "Room"}
	w.Rooms[5203] = room

	ch, client := makeObjTestChar("Wizard")
	defer client.Close()
	ch.Class = types.CLASS_MAGE
	handler.CharToRoom(ch, room)

	idx := &types.ObjIndexData{Vnum: 3103, Name: "fighter helm", ShortDescr: "a fighter's helm",
		WearFlags: int(types.ITEM_TAKE | types.ITEM_WEAR_HEAD)}
	w.ObjIndex[3103] = idx
	obj := handler.CreateObject(w, idx, 1)
	obj.ExtraFlags.Set(types.ITEM_ANTI_MAGE)
	handler.ObjToChar(obj, ch)

	DoWear(ch, "helm")
	out := readOutput(ch, client)
	if obj.WearLoc != types.WEAR_NONE {
		t.Errorf("ANTI_MAGE should block mage")
	}
	if !strings.Contains(out, "mage") {
		t.Errorf("expected class refusal, got: %q", out)
	}
}

func TestDoWear_AntiClassAllowsDifferentClass(t *testing.T) {
	w := setupObjWorld()
	room := &types.RoomIndexData{Vnum: 5204, Name: "Room"}
	w.Rooms[5204] = room

	ch, client := makeObjTestChar("Fighter")
	defer client.Close()
	ch.Class = types.CLASS_WARRIOR
	handler.CharToRoom(ch, room)

	idx := &types.ObjIndexData{Vnum: 3104, Name: "fighter helm", ShortDescr: "a fighter's helm",
		WearFlags: int(types.ITEM_TAKE | types.ITEM_WEAR_HEAD)}
	w.ObjIndex[3104] = idx
	obj := handler.CreateObject(w, idx, 1)
	obj.ExtraFlags.Set(types.ITEM_ANTI_MAGE)
	handler.ObjToChar(obj, ch)

	DoWear(ch, "helm")
	_ = readOutput(ch, client)
	if obj.WearLoc != types.WEAR_HEAD {
		t.Errorf("warrior should be allowed to wear ANTI_MAGE item; WearLoc=%d", obj.WearLoc)
	}
}
