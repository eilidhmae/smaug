package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/command"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func setupWizWorld() *world.World {
	w := world.New("/tmp/test")
	WorldRef = w
	return w
}

// --- DoEcho ---

func TestDoEcho_NoArg(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoEcho(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Echo what?") {
		t.Errorf("expected 'Echo what?', got: %q", out)
	}
}

func TestDoEcho_SendsToAll(t *testing.T) {
	w := setupWizWorld()
	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	w.Descriptors = append(w.Descriptors, ch.Desc)

	listener, listenerClient := makeTestChar("Player")
	defer listenerClient.Close()
	w.Descriptors = append(w.Descriptors, listener.Desc)

	DoEcho(ch, "Server rebooting in 5 minutes")

	chOut := readOutput(ch, chClient)
	if !strings.Contains(chOut, "Server rebooting in 5 minutes") {
		t.Errorf("sender should see echo, got: %q", chOut)
	}

	listenerOut := readOutput(listener, listenerClient)
	if !strings.Contains(listenerOut, "Server rebooting in 5 minutes") {
		t.Errorf("listener should see echo, got: %q", listenerOut)
	}
}

// --- DoRecho ---

func TestDoRecho_NoArg(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()
	ch.InRoom = &types.RoomIndexData{Vnum: 7000, Name: "Test"}

	DoRecho(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Recho what?") {
		t.Errorf("expected 'Recho what?', got: %q", out)
	}
}

func TestDoRecho_SendsToRoom(t *testing.T) {
	_ = setupWizWorld()
	room := &types.RoomIndexData{Vnum: 7001, Name: "Test Room"}

	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	listener, listenerClient := makeTestChar("Player")
	defer listenerClient.Close()
	listener.InRoom = room
	room.People = append(room.People, listener)

	DoRecho(ch, "The walls shake violently!")

	chOut := readOutput(ch, chClient)
	if !strings.Contains(chOut, "The walls shake violently!") {
		t.Errorf("sender should see recho, got: %q", chOut)
	}

	listenerOut := readOutput(listener, listenerClient)
	if !strings.Contains(listenerOut, "The walls shake violently!") {
		t.Errorf("listener should see recho, got: %q", listenerOut)
	}
}

func TestDoRecho_NilRoom(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()
	ch.InRoom = nil

	// Should not panic
	DoRecho(ch, "test")
	_ = readOutput(ch, client)
}

// --- DoBamfin / DoBamfout ---

func TestDoBamfin_Set(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoBamfin(ch, "arrives in a burst of flame")
	out := readOutput(ch, client)

	if ch.PCData.BamfIn != "arrives in a burst of flame" {
		t.Errorf("BamfIn not set, got: %q", ch.PCData.BamfIn)
	}
	if !strings.Contains(out, "Bamfin set to") {
		t.Errorf("expected confirmation, got: %q", out)
	}
}

func TestDoBamfin_Show(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()
	ch.PCData.BamfIn = "appears in a flash"

	DoBamfin(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "appears in a flash") {
		t.Errorf("expected current bamfin shown, got: %q", out)
	}
}

func TestDoBamfout_Set(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoBamfout(ch, "vanishes in a puff of smoke")
	out := readOutput(ch, client)

	if ch.PCData.BamfOut != "vanishes in a puff of smoke" {
		t.Errorf("BamfOut not set, got: %q", ch.PCData.BamfOut)
	}
	if !strings.Contains(out, "Bamfout set to") {
		t.Errorf("expected confirmation, got: %q", out)
	}
}

func TestDoBamfout_Show(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()
	ch.PCData.BamfOut = "leaves in a swirl"

	DoBamfout(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "leaves in a swirl") {
		t.Errorf("expected current bamfout shown, got: %q", out)
	}
}

// --- DoInvis ---

func TestDoInvis_ToggleOn(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()
	ch.Level = types.LEVEL_IMMORTAL

	DoInvis(ch, "")
	out := readOutput(ch, client)

	if !ch.Act.IsSet(types.PLR_WIZINVIS) {
		t.Error("PLR_WIZINVIS should be set after toggle on")
	}
	if !strings.Contains(out, "invisible") {
		t.Errorf("expected invisibility message, got: %q", out)
	}
}

func TestDoInvis_ToggleOff(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()
	ch.Level = types.LEVEL_IMMORTAL
	ch.Act.Set(types.PLR_WIZINVIS)
	ch.PCData.WizInvis = ch.GetTrust()

	DoInvis(ch, "")
	out := readOutput(ch, client)

	if ch.Act.IsSet(types.PLR_WIZINVIS) {
		t.Error("PLR_WIZINVIS should be cleared after toggle off")
	}
	if ch.PCData.WizInvis != 0 {
		t.Errorf("WizInvis should be 0, got: %d", ch.PCData.WizInvis)
	}
	if !strings.Contains(out, "visible") {
		t.Errorf("expected visible message, got: %q", out)
	}
}

func TestDoInvis_WithLevel(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()
	ch.Level = types.LEVEL_IMMORTAL

	DoInvis(ch, "5")
	out := readOutput(ch, client)

	if !ch.Act.IsSet(types.PLR_WIZINVIS) {
		t.Error("PLR_WIZINVIS should be set")
	}
	if ch.PCData.WizInvis != 5 {
		t.Errorf("WizInvis should be 5, got: %d", ch.PCData.WizInvis)
	}
	if !strings.Contains(out, "5") {
		t.Errorf("expected level in message, got: %q", out)
	}
}

func TestDoInvis_InvalidLevel(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()
	ch.Level = types.LEVEL_IMMORTAL

	DoInvis(ch, "999")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Valid range") {
		t.Errorf("expected 'Valid range' message, got: %q", out)
	}
}

// --- DoHolylight ---

func TestDoHolylight_ToggleOn(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoHolylight(ch, "")
	out := readOutput(ch, client)

	if !ch.Act.IsSet(types.PLR_HOLYLIGHT) {
		t.Error("PLR_HOLYLIGHT should be set")
	}
	if !strings.Contains(out, "on") {
		t.Errorf("expected 'on' message, got: %q", out)
	}
}

func TestDoHolylight_ToggleOff(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()
	ch.Act.Set(types.PLR_HOLYLIGHT)

	DoHolylight(ch, "")
	out := readOutput(ch, client)

	if ch.Act.IsSet(types.PLR_HOLYLIGHT) {
		t.Error("PLR_HOLYLIGHT should be cleared")
	}
	if !strings.Contains(out, "off") {
		t.Errorf("expected 'off' message, got: %q", out)
	}
}

func TestDoHolylight_NPC(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Guard")
	defer client.Close()
	ch.Act.Set(types.ACT_IS_NPC)

	DoHolylight(ch, "")
	out := readOutput(ch, client)

	// NPC should get no output (function returns early)
	if out != "" {
		t.Errorf("NPC should get no output, got: %q", out)
	}
}

// --- DoFreeze ---

func TestDoFreeze_NoArg(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()
	ch.Level = types.LEVEL_IMMORTAL + 5

	DoFreeze(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Freeze whom?") {
		t.Errorf("expected 'Freeze whom?', got: %q", out)
	}
}

func TestDoFreeze_FreezePlayer(t *testing.T) {
	w := setupWizWorld()
	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	ch.Level = types.LEVEL_IMMORTAL + 5
	w.AddChar(ch)

	victim, victimClient := makeTestChar("Badguy")
	defer victimClient.Close()
	victim.Level = 5
	w.AddChar(victim)

	DoFreeze(ch, "Badguy")
	chOut := readOutput(ch, chClient)
	victimOut := readOutput(victim, victimClient)

	if !victim.Act.IsSet(types.PLR_FREEZE) {
		t.Error("victim should be frozen")
	}
	if !strings.Contains(chOut, "frozen") {
		t.Errorf("expected 'frozen' in imm output, got: %q", chOut)
	}
	if !strings.Contains(victimOut, "frozen") {
		t.Errorf("expected 'frozen' in victim output, got: %q", victimOut)
	}
}

func TestDoFreeze_UnfreezePlayer(t *testing.T) {
	w := setupWizWorld()
	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	ch.Level = types.LEVEL_IMMORTAL + 5
	w.AddChar(ch)

	victim, victimClient := makeTestChar("Badguy")
	defer victimClient.Close()
	victim.Level = 5
	victim.Act.Set(types.PLR_FREEZE)
	w.AddChar(victim)

	DoFreeze(ch, "Badguy")
	chOut := readOutput(ch, chClient)
	victimOut := readOutput(victim, victimClient)

	if victim.Act.IsSet(types.PLR_FREEZE) {
		t.Error("victim should be unfrozen")
	}
	if !strings.Contains(chOut, "unfrozen") {
		t.Errorf("expected 'unfrozen' in imm output, got: %q", chOut)
	}
	if !strings.Contains(victimOut, "play again") {
		t.Errorf("expected 'play again' in victim output, got: %q", victimOut)
	}
}

func TestDoFreeze_NPC(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 7010, Name: "Test"}
	w.Rooms[7010] = room

	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	ch.Level = types.LEVEL_IMMORTAL + 5
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	npc := &types.CharData{Name: "guard", Level: 5, InRoom: room}
	npc.Act.Set(types.ACT_IS_NPC)
	w.AddChar(npc)

	DoFreeze(ch, "guard")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "Not on NPCs") {
		t.Errorf("expected 'Not on NPCs', got: %q", out)
	}
}

func TestDoFreeze_EqualTrust(t *testing.T) {
	w := setupWizWorld()
	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	ch.Level = types.LEVEL_IMMORTAL
	w.AddChar(ch)

	victim, victimClient := makeTestChar("OtherImm")
	defer victimClient.Close()
	victim.Level = types.LEVEL_IMMORTAL
	w.AddChar(victim)

	DoFreeze(ch, "OtherImm")
	out := readOutput(ch, chClient)
	_ = readOutput(victim, victimClient)

	if !strings.Contains(out, "can't do that") {
		t.Errorf("expected 'can't do that', got: %q", out)
	}
}

// --- DoSilence ---

func TestDoSilence_NoArg(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoSilence(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Silence whom?") {
		t.Errorf("expected 'Silence whom?', got: %q", out)
	}
}

func TestDoSilence_SilencePlayer(t *testing.T) {
	w := setupWizWorld()
	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	ch.Level = types.LEVEL_IMMORTAL + 5
	w.AddChar(ch)

	victim, victimClient := makeTestChar("Spammer")
	defer victimClient.Close()
	victim.Level = 5
	w.AddChar(victim)

	DoSilence(ch, "Spammer")
	chOut := readOutput(ch, chClient)
	victimOut := readOutput(victim, victimClient)

	if !victim.Act.IsSet(types.PLR_SILENCE) {
		t.Error("victim should be silenced")
	}
	if !strings.Contains(chOut, "silenced") {
		t.Errorf("expected 'silenced' in imm output, got: %q", chOut)
	}
	if !strings.Contains(victimOut, "silenced") {
		t.Errorf("expected 'silenced' in victim output, got: %q", victimOut)
	}
}

func TestDoSilence_UnsilencePlayer(t *testing.T) {
	w := setupWizWorld()
	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	ch.Level = types.LEVEL_IMMORTAL + 5
	w.AddChar(ch)

	victim, victimClient := makeTestChar("Spammer")
	defer victimClient.Close()
	victim.Level = 5
	victim.Act.Set(types.PLR_SILENCE)
	w.AddChar(victim)

	DoSilence(ch, "Spammer")
	chOut := readOutput(ch, chClient)
	victimOut := readOutput(victim, victimClient)

	if victim.Act.IsSet(types.PLR_SILENCE) {
		t.Error("victim should be unsilenced")
	}
	if !strings.Contains(chOut, "unsilenced") {
		t.Errorf("expected 'unsilenced' in imm output, got: %q", chOut)
	}
	if !strings.Contains(victimOut, "channels again") {
		t.Errorf("expected 'channels again' in victim output, got: %q", victimOut)
	}
}

// --- DoMstat ---

func TestDoMstat_NoArg(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoMstat(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Stat whom?") {
		t.Errorf("expected 'Stat whom?', got: %q", out)
	}
}

func TestDoMstat_ShowsStats(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 7020, Name: "Test Room"}
	w.Rooms[7020] = room

	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	ch.Level = types.LEVEL_IMMORTAL
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	npc := &types.CharData{
		Name:       "dragon red",
		ShortDescr: "a red dragon",
		Level:      50,
		Hit:        500, MaxHit: 500,
		Mana: 200, MaxMana: 200,
		Move: 300, MaxMove: 300,
		Gold: 1000,
		PermStr: 20, PermInt: 15, PermWis: 15,
		PermDex: 12, PermCon: 20, PermCha: 10, PermLck: 10,
		Position:  types.POS_STANDING,
		Alignment: -1000,
		InRoom:    room,
	}
	npc.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(npc, room)
	w.AddChar(npc)

	DoMstat(ch, "dragon")
	out := readOutput(ch, chClient)

	if !strings.Contains(out, "dragon red") {
		t.Errorf("expected mob name in output, got: %q", out)
	}
	if !strings.Contains(out, "a red dragon") {
		t.Errorf("expected short descr in output, got: %q", out)
	}
	if !strings.Contains(out, "500/500") {
		t.Errorf("expected HP in output, got: %q", out)
	}
	if !strings.Contains(out, "Level: 50") {
		t.Errorf("expected level in output, got: %q", out)
	}
}

func TestDoMstat_NotFound(t *testing.T) {
	w := setupWizWorld()
	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	ch.Level = types.LEVEL_IMMORTAL
	w.AddChar(ch)

	DoMstat(ch, "nonexistent")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "aren't here") {
		t.Errorf("expected 'aren't here', got: %q", out)
	}
}

// --- DoOstat ---

func TestDoOstat_NoArg(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoOstat(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Stat what object?") {
		t.Errorf("expected 'Stat what object?', got: %q", out)
	}
}

func TestDoOstat_ShowsStats(t *testing.T) {
	w := setupWizWorld()
	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	ch.Level = types.LEVEL_IMMORTAL

	idx := &types.ObjIndexData{Vnum: 8000, Name: "sword flame", ShortDescr: "a flaming sword"}
	w.ObjIndex[8000] = idx
	obj := handler.CreateObject(w, idx, 10)
	obj.ItemType = types.ITEM_WEAPON
	obj.GoldCost = 500
	obj.Weight = 5
	handler.ObjToChar(obj, ch)
	w.AddChar(ch)

	DoOstat(ch, "sword")
	out := readOutput(ch, chClient)

	if !strings.Contains(out, "a flaming sword") {
		t.Errorf("expected short descr in output, got: %q", out)
	}
	if !strings.Contains(out, "8000") {
		t.Errorf("expected vnum in output, got: %q", out)
	}
}

func TestDoOstat_NotFound(t *testing.T) {
	w := setupWizWorld()
	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	w.AddChar(ch)

	DoOstat(ch, "nonexistent")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "Nothing like that") {
		t.Errorf("expected 'Nothing like that', got: %q", out)
	}
}

// --- DoRstat ---

func TestDoRstat_CurrentRoom(t *testing.T) {
	_ = setupWizWorld()
	room := &types.RoomIndexData{
		Vnum:        7030,
		Name:        "Wizard Tower",
		Description: "A tall tower.\n\r",
		SectorType:  types.SECT_INSIDE,
	}
	room.Exits = []*types.ExitData{
		{Direction: types.DIR_NORTH, ToRoom: &types.RoomIndexData{Vnum: 7031}},
	}

	ch, client := makeTestChar("Imm")
	defer client.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	DoRstat(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Wizard Tower") {
		t.Errorf("expected room name, got: %q", out)
	}
	if !strings.Contains(out, "7030") {
		t.Errorf("expected vnum, got: %q", out)
	}
	if !strings.Contains(out, "north") {
		t.Errorf("expected exit info, got: %q", out)
	}
}

func TestDoRstat_NilRoom(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()
	ch.InRoom = nil

	DoRstat(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "not in a room") {
		t.Errorf("expected 'not in a room', got: %q", out)
	}
}

// --- DoUsers ---

func TestDoUsers(t *testing.T) {
	w := setupWizWorld()
	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	w.Descriptors = append(w.Descriptors, ch.Desc)

	player, playerClient := makeTestChar("Player")
	defer playerClient.Close()
	w.Descriptors = append(w.Descriptors, player.Desc)

	DoUsers(ch, "")
	out := readOutput(ch, chClient)

	if !strings.Contains(out, "Imm") {
		t.Errorf("expected 'Imm' in users list, got: %q", out)
	}
	if !strings.Contains(out, "Player") {
		t.Errorf("expected 'Player' in users list, got: %q", out)
	}
	if !strings.Contains(out, "2 user") {
		t.Errorf("expected '2 user' count, got: %q", out)
	}
}

// --- DoMfind ---

func TestDoMfind_NoArg(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoMfind(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Find what mob?") {
		t.Errorf("expected 'Find what mob?', got: %q", out)
	}
}

func TestDoMfind_Found(t *testing.T) {
	w := setupWizWorld()
	w.MobIndex[100] = &types.MobIndexData{Vnum: 100, PlayerName: "guard city", ShortDescr: "a city guard"}
	w.MobIndex[200] = &types.MobIndexData{Vnum: 200, PlayerName: "dragon red", ShortDescr: "a red dragon"}

	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoMfind(ch, "guard")
	out := readOutput(ch, client)
	if !strings.Contains(out, "city guard") {
		t.Errorf("expected mob found, got: %q", out)
	}
}

func TestDoMfind_NotFound(t *testing.T) {
	w := setupWizWorld()
	w.MobIndex[100] = &types.MobIndexData{Vnum: 100, PlayerName: "guard", ShortDescr: "a guard"}

	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoMfind(ch, "unicorn")
	out := readOutput(ch, client)
	if !strings.Contains(out, "No mobs found") {
		t.Errorf("expected 'No mobs found', got: %q", out)
	}
}

// --- DoOfind ---

func TestDoOfind_NoArg(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoOfind(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Find what object?") {
		t.Errorf("expected 'Find what object?', got: %q", out)
	}
}

func TestDoOfind_Found(t *testing.T) {
	w := setupWizWorld()
	w.ObjIndex[100] = &types.ObjIndexData{Vnum: 100, Name: "sword flame", ShortDescr: "a flaming sword"}
	w.ObjIndex[200] = &types.ObjIndexData{Vnum: 200, Name: "shield oak", ShortDescr: "an oak shield"}

	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoOfind(ch, "sword")
	out := readOutput(ch, client)
	if !strings.Contains(out, "flaming sword") {
		t.Errorf("expected obj found, got: %q", out)
	}
}

func TestDoOfind_NotFound(t *testing.T) {
	w := setupWizWorld()
	w.ObjIndex[100] = &types.ObjIndexData{Vnum: 100, Name: "sword", ShortDescr: "a sword"}

	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoOfind(ch, "unicorn")
	out := readOutput(ch, client)
	if !strings.Contains(out, "No objects found") {
		t.Errorf("expected 'No objects found', got: %q", out)
	}
}

// --- DoMwhere ---

func TestDoMwhere_NoArg(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoMwhere(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Find what mob?") {
		t.Errorf("expected 'Find what mob?', got: %q", out)
	}
}

func TestDoMwhere_Found(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 7040, Name: "Guard Post"}
	w.Rooms[7040] = room

	npc := &types.CharData{
		Name:       "guard city",
		ShortDescr: "a city guard",
		InRoom:     room,
		IndexData:  &types.MobIndexData{Vnum: 100},
	}
	npc.Act.Set(types.ACT_IS_NPC)
	w.AddChar(npc)

	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoMwhere(ch, "guard")
	out := readOutput(ch, client)
	if !strings.Contains(out, "city guard") {
		t.Errorf("expected mob location, got: %q", out)
	}
	if !strings.Contains(out, "Guard Post") {
		t.Errorf("expected room name, got: %q", out)
	}
}

func TestDoMwhere_NotFound(t *testing.T) {
	w := setupWizWorld()
	_ = w
	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoMwhere(ch, "nonexistent")
	out := readOutput(ch, client)
	if !strings.Contains(out, "No mobs found") {
		t.Errorf("expected 'No mobs found', got: %q", out)
	}
}

// --- DoOwhere ---

func TestDoOwhere_NoArg(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoOwhere(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Find what object?") {
		t.Errorf("expected 'Find what object?', got: %q", out)
	}
}

func TestDoOwhere_FoundCarried(t *testing.T) {
	w := setupWizWorld()
	carrier := &types.CharData{Name: "Adventurer"}
	carrier.Act.Set(types.ACT_IS_NPC)
	w.AddChar(carrier)

	idx := &types.ObjIndexData{Vnum: 300, Name: "sword magic", ShortDescr: "a magic sword"}
	obj := &types.ObjData{
		Name:       "sword magic",
		ShortDescr: "a magic sword",
		CarriedBy:  carrier,
		IndexData:  idx,
	}
	w.Objects = append(w.Objects, obj)

	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoOwhere(ch, "sword")
	out := readOutput(ch, client)
	if !strings.Contains(out, "magic sword") {
		t.Errorf("expected obj in output, got: %q", out)
	}
	if !strings.Contains(out, "carried by Adventurer") {
		t.Errorf("expected carrier info, got: %q", out)
	}
}

func TestDoOwhere_NotFound(t *testing.T) {
	w := setupWizWorld()
	_ = w
	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoOwhere(ch, "nonexistent")
	out := readOutput(ch, client)
	if !strings.Contains(out, "No objects found") {
		t.Errorf("expected 'No objects found', got: %q", out)
	}
}

// --- DoPeace ---

func TestDoPeace(t *testing.T) {
	_ = setupWizWorld()
	room := &types.RoomIndexData{Vnum: 7050, Name: "Arena"}

	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	ch.InRoom = room
	room.People = append(room.People, ch)

	fighter := &types.CharData{
		Name:     "fighter",
		Position: types.POS_FIGHTING,
		InRoom:   room,
		Fighting: &types.FightData{Who: ch},
	}
	fighter.Act.Set(types.ACT_IS_NPC)
	room.People = append(room.People, fighter)

	DoPeace(ch, "")
	out := readOutput(ch, chClient)

	if fighter.Position != types.POS_STANDING {
		t.Errorf("fighter should be standing, got position: %d", fighter.Position)
	}
	if fighter.Fighting != nil {
		t.Error("fighter should no longer be fighting")
	}
	if !strings.Contains(out, "Peace has been restored") {
		t.Errorf("expected peace message, got: %q", out)
	}
}

// --- DoRestore ---

func TestDoRestore_NoArg(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoRestore(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Restore whom?") {
		t.Errorf("expected 'Restore whom?', got: %q", out)
	}
}

func TestDoRestore_Player(t *testing.T) {
	w := setupWizWorld()
	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	ch.Level = types.LEVEL_IMMORTAL
	w.AddChar(ch)

	victim, victimClient := makeTestChar("Wounded")
	defer victimClient.Close()
	victim.Hit = 10
	victim.Mana = 5
	victim.Move = 3
	w.AddChar(victim)

	DoRestore(ch, "Wounded")
	_ = readOutput(ch, chClient)
	victimOut := readOutput(victim, victimClient)

	if victim.Hit != victim.MaxHit {
		t.Errorf("HP should be max, got: %d/%d", victim.Hit, victim.MaxHit)
	}
	if victim.Mana != victim.MaxMana {
		t.Errorf("Mana should be max, got: %d/%d", victim.Mana, victim.MaxMana)
	}
	if victim.Move != victim.MaxMove {
		t.Errorf("Move should be max, got: %d/%d", victim.Move, victim.MaxMove)
	}
	if !strings.Contains(victimOut, "fully restored") {
		t.Errorf("expected 'fully restored' message, got: %q", victimOut)
	}
}

// --- DoGoto ---

func TestDoGoto_NoArg(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoGoto(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Goto where?") {
		t.Errorf("expected 'Goto where?', got: %q", out)
	}
}

func TestDoGoto_ByVnum(t *testing.T) {
	w := setupWizWorld()
	origin := &types.RoomIndexData{Vnum: 7060, Name: "Origin"}
	dest := &types.RoomIndexData{Vnum: 7061, Name: "Destination"}
	w.Rooms[7060] = origin
	w.Rooms[7061] = dest

	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	handler.CharToRoom(ch, origin)

	DoGoto(ch, "7061")
	_ = readOutput(ch, chClient)

	if ch.InRoom != dest {
		t.Error("character should have moved to destination room")
	}
}

func TestDoGoto_BadVnum(t *testing.T) {
	w := setupWizWorld()
	origin := &types.RoomIndexData{Vnum: 7062, Name: "Origin"}
	w.Rooms[7062] = origin

	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	handler.CharToRoom(ch, origin)

	DoGoto(ch, "99999")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "No room with vnum") {
		t.Errorf("expected 'No room with vnum', got: %q", out)
	}
}

// --- DoSnoop ---

func TestDoSnoop_NoArg(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoSnoop(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Snoop whom?") {
		t.Errorf("expected 'Snoop whom?', got: %q", out)
	}
}

func TestDoSnoop_Self_CancelAll(t *testing.T) {
	w := setupWizWorld()
	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	ch.Level = types.LEVEL_IMMORTAL
	w.AddChar(ch)
	w.Descriptors = append(w.Descriptors, ch.Desc)

	victim, victimClient := makeTestChar("Target")
	defer victimClient.Close()
	victim.Level = 5
	w.AddChar(victim)
	w.Descriptors = append(w.Descriptors, victim.Desc)

	// Set up a snoop
	victim.Desc.SnoopBy = ch.Desc

	DoSnoop(ch, "Imm")
	out := readOutput(ch, chClient)

	if victim.Desc.SnoopBy != nil {
		t.Error("snoop should be cancelled")
	}
	if !strings.Contains(out, "cancelled") {
		t.Errorf("expected 'cancelled' message, got: %q", out)
	}
}

func TestDoSnoop_StartSnooping(t *testing.T) {
	w := setupWizWorld()
	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	ch.Level = types.LEVEL_IMMORTAL + 5
	w.AddChar(ch)
	w.Descriptors = append(w.Descriptors, ch.Desc)

	victim, victimClient := makeTestChar("Target")
	defer victimClient.Close()
	victim.Level = 5
	w.AddChar(victim)
	w.Descriptors = append(w.Descriptors, victim.Desc)

	DoSnoop(ch, "Target")
	out := readOutput(ch, chClient)

	if victim.Desc.SnoopBy != ch.Desc {
		t.Error("victim's descriptor should be snooped by imm")
	}
	if !strings.Contains(out, "snooping") {
		t.Errorf("expected 'snooping' message, got: %q", out)
	}
}

func TestDoSnoop_AlreadySnooped(t *testing.T) {
	w := setupWizWorld()
	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	ch.Level = types.LEVEL_IMMORTAL + 5
	w.AddChar(ch)

	otherImm, otherClient := makeTestChar("OtherImm")
	defer otherClient.Close()
	otherImm.Level = types.LEVEL_IMMORTAL + 5
	w.AddChar(otherImm)

	victim, victimClient := makeTestChar("Target")
	defer victimClient.Close()
	victim.Level = 5
	w.AddChar(victim)

	// Already being snooped by someone else
	victim.Desc.SnoopBy = otherImm.Desc

	DoSnoop(ch, "Target")
	out := readOutput(ch, chClient)

	if !strings.Contains(out, "already being snooped") {
		t.Errorf("expected 'already being snooped', got: %q", out)
	}
}

// --- DoTransfer ---

func TestDoTransfer_NoArg(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoTransfer(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Transfer whom?") {
		t.Errorf("expected 'Transfer whom?', got: %q", out)
	}
}

func TestDoTransfer_Player(t *testing.T) {
	w := setupWizWorld()
	room1 := &types.RoomIndexData{Vnum: 7070, Name: "Imm Room"}
	room2 := &types.RoomIndexData{Vnum: 7071, Name: "Player Room"}
	w.Rooms[7070] = room1
	w.Rooms[7071] = room2

	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	ch.Level = types.LEVEL_IMMORTAL
	handler.CharToRoom(ch, room1)
	w.AddChar(ch)

	victim, victimClient := makeTestChar("Target")
	defer victimClient.Close()
	handler.CharToRoom(victim, room2)
	w.AddChar(victim)

	DoTransfer(ch, "Target")
	_ = readOutput(ch, chClient)
	_ = readOutput(victim, victimClient)

	if victim.InRoom != room1 {
		t.Error("victim should be transferred to imm's room")
	}
}

func TestDoTransfer_NotFound(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 7072, Name: "Imm Room"}
	w.Rooms[7072] = room

	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	handler.CharToRoom(ch, room)

	DoTransfer(ch, "nonexistent")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "aren't here") {
		t.Errorf("expected 'aren't here', got: %q", out)
	}
}

func TestDoTransfer_Self(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 7073, Name: "Imm Room"}
	w.Rooms[7073] = room

	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	DoTransfer(ch, "Imm")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "That's you") {
		t.Errorf("expected 'That's you', got: %q", out)
	}
}

// --- DoPurge ---

func TestDoPurge_Room(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 7080, Name: "Test Room"}
	w.Rooms[7080] = room

	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	handler.CharToRoom(ch, room)

	npc := &types.CharData{Name: "rat", Level: 2, Position: types.POS_STANDING}
	npc.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(npc, room)
	w.AddChar(npc)

	idx := &types.ObjIndexData{Vnum: 9500, Name: "junk", ShortDescr: "junk"}
	w.ObjIndex[9500] = idx
	obj := handler.CreateObject(w, idx, 1)
	handler.ObjToRoom(obj, room)

	DoPurge(ch, "")
	out := readOutput(ch, chClient)

	if !strings.Contains(out, "purged") {
		t.Errorf("expected 'purged', got: %q", out)
	}
	// Room should have only the imm left
	npcCount := 0
	for _, p := range room.People {
		if p.IsNPC() {
			npcCount++
		}
	}
	if npcCount > 0 {
		t.Errorf("expected 0 NPCs, got %d", npcCount)
	}
}

func TestDoPurge_SpecificNPC(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 7081, Name: "Test Room"}
	w.Rooms[7081] = room

	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	handler.CharToRoom(ch, room)

	npc := &types.CharData{Name: "rat sewer", Level: 2, Position: types.POS_STANDING}
	npc.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(npc, room)
	w.AddChar(npc)

	DoPurge(ch, "rat")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "Ok") {
		t.Errorf("expected 'Ok', got: %q", out)
	}
}

func TestDoPurge_Player(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 7082, Name: "Test Room"}
	w.Rooms[7082] = room

	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	handler.CharToRoom(ch, room)

	victim, victimClient := makeTestChar("Target")
	defer victimClient.Close()
	handler.CharToRoom(victim, room)

	DoPurge(ch, "Target")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "only purge NPCs") {
		t.Errorf("expected 'only purge NPCs', got: %q", out)
	}
}

func TestDoPurge_NotFound(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 7083, Name: "Test Room"}
	w.Rooms[7083] = room

	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	handler.CharToRoom(ch, room)

	DoPurge(ch, "nonexistent")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "Nothing like that") {
		t.Errorf("expected 'Nothing like that', got: %q", out)
	}
}

// --- DoAdvance ---

func TestDoAdvance_NoArg(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoAdvance(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Usage") {
		t.Errorf("expected 'Usage', got: %q", out)
	}
}

func TestDoAdvance_Success(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 7090, Name: "Test"}
	w.Rooms[7090] = room

	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	ch.Level = types.LEVEL_IMMORTAL + 5
	handler.CharToRoom(ch, room)

	victim, victimClient := makeTestChar("Newbie")
	defer victimClient.Close()
	victim.Level = 1
	handler.CharToRoom(victim, room)

	DoAdvance(ch, "Newbie 10")
	_ = readOutput(ch, chClient)
	victimOut := readOutput(victim, victimClient)

	if victim.Level != 10 {
		t.Errorf("victim level should be 10, got: %d", victim.Level)
	}
	if !strings.Contains(victimOut, "10") {
		t.Errorf("victim should be notified, got: %q", victimOut)
	}
}

func TestDoAdvance_NPC(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 7091, Name: "Test"}
	w.Rooms[7091] = room

	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	ch.Level = types.LEVEL_IMMORTAL + 5
	handler.CharToRoom(ch, room)

	npc := &types.CharData{Name: "rat", Level: 2, Position: types.POS_STANDING}
	npc.Act.Set(types.ACT_IS_NPC)
	handler.CharToRoom(npc, room)

	DoAdvance(ch, "rat 10")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "Not on NPCs") {
		t.Errorf("expected 'Not on NPCs', got: %q", out)
	}
}

func TestDoAdvance_BadLevel(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 7092, Name: "Test"}
	w.Rooms[7092] = room

	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	ch.Level = types.LEVEL_IMMORTAL
	handler.CharToRoom(ch, room)

	victim, victimClient := makeTestChar("Target")
	defer victimClient.Close()
	handler.CharToRoom(victim, room)

	DoAdvance(ch, "Target 0")
	out := readOutput(ch, chClient)
	_ = readOutput(victim, victimClient)
	if !strings.Contains(out, "Level must be") {
		t.Errorf("expected 'Level must be', got: %q", out)
	}
}

// --- DoSlay ---

func TestDoSlay_NoArg(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()

	DoSlay(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Slay whom?") {
		t.Errorf("expected 'Slay whom?', got: %q", out)
	}
}

func TestDoSlay_Self(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 7095, Name: "Test"}
	w.Rooms[7095] = room

	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	handler.CharToRoom(ch, room)
	w.AddChar(ch)

	DoSlay(ch, "Imm")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "Suicide") {
		t.Errorf("expected suicide message, got: %q", out)
	}
}

func TestDoSlay_NPC(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 7096, Name: "Test"}
	w.Rooms[7096] = room

	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	ch.Level = types.LEVEL_IMMORTAL + 5
	handler.CharToRoom(ch, room)

	npc := &types.CharData{
		Name: "rat", ShortDescr: "a rat",
		Level: 2, Position: types.POS_STANDING,
		Hit: 20, MaxHit: 20,
	}
	npc.Act.Set(types.ACT_IS_NPC)
	npc.IndexData = &types.MobIndexData{Vnum: 999}
	handler.CharToRoom(npc, room)
	w.AddChar(npc)

	DoSlay(ch, "rat")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "slay") {
		t.Errorf("expected slay message, got: %q", out)
	}
}

func TestDoSlay_Player(t *testing.T) {
	w := setupWizWorld()
	room := &types.RoomIndexData{Vnum: 7097, Name: "Test"}
	w.Rooms[7097] = room

	ch, chClient := makeTestChar("Imm")
	defer chClient.Close()
	ch.Level = types.LEVEL_IMMORTAL + 5
	handler.CharToRoom(ch, room)

	victim, victimClient := makeTestChar("Target")
	defer victimClient.Close()
	victim.Level = 5
	victim.Hit = 100
	handler.CharToRoom(victim, room)

	DoSlay(ch, "Target")
	_ = readOutput(ch, chClient)
	_ = readOutput(victim, victimClient)

	if victim.Hit != 1 {
		t.Errorf("player victim should have 1 HP, got: %d", victim.Hit)
	}
	if victim.Position != types.POS_RESTING {
		t.Errorf("player victim should be resting, got: %d", victim.Position)
	}
}

func TestDoForce_TrustCap(t *testing.T) {
	w := setupWizWorld()

	// Create a command at level 58 (between forcer trust 55 and victim trust 60)
	reg := command.NewRegistry()
	dispatched := false
	reg.Register(&command.Command{
		Name:     "secretwiz",
		Position: 0,
		Level:    58,
		DoFun: func(ch *types.CharData, argument string) {
			dispatched = true
		},
	})
	CmdRegistry = reg

	room := &types.RoomIndexData{Vnum: 3001, Name: "Test Room"}
	w.Rooms[3001] = room

	// Forcer: trust level 55
	forcer, forcerClient := makeTestChar("Forcer")
	defer forcerClient.Close()
	forcer.Level = 55
	handler.CharToRoom(forcer, room)
	w.Descriptors = append(w.Descriptors, forcer.Desc)

	// Victim: trust level 60 (higher than forcer)
	victim, victimClient := makeTestChar("Victim")
	defer victimClient.Close()
	victim.Level = 60
	handler.CharToRoom(victim, room)
	w.Descriptors = append(w.Descriptors, victim.Desc)

	// Force victim to run "secretwiz" — should be blocked because trust is
	// capped to forcer's 55, which is below the command's level 58
	DoForce(forcer, "Victim secretwiz")
	_ = readOutput(forcer, forcerClient)
	_ = readOutput(victim, victimClient)

	if dispatched {
		t.Fatal("secretwiz should NOT have been dispatched — trust capped to forcer's 55, command needs 58")
	}

	// Now test that a command within the forcer's trust level works
	lowVictim, lowVictimClient := makeTestChar("LowVictim")
	defer lowVictimClient.Close()
	lowVictim.Level = 52
	handler.CharToRoom(lowVictim, room)
	w.Descriptors = append(w.Descriptors, lowVictim.Desc)

	allowedCmd := false
	reg.Register(&command.Command{
		Name:     "allowed",
		Position: 0,
		Level:    50,
		DoFun: func(ch *types.CharData, argument string) {
			allowedCmd = true
		},
	})

	DoForce(forcer, "LowVictim allowed")
	_ = readOutput(forcer, forcerClient)
	_ = readOutput(lowVictim, lowVictimClient)

	if !allowedCmd {
		t.Fatal("allowed command should have been dispatched — level 50 is within forcer's trust 55")
	}
}

func TestDoForce_TrustCapAll(t *testing.T) {
	w := setupWizWorld()

	reg := command.NewRegistry()
	dispatched := 0
	reg.Register(&command.Command{
		Name:     "highcmd",
		Position: 0,
		Level:    58,
		DoFun: func(ch *types.CharData, argument string) {
			dispatched++
		},
	})
	CmdRegistry = reg

	room := &types.RoomIndexData{Vnum: 3002, Name: "Test Room"}
	w.Rooms[3002] = room

	// Forcer at trust 55
	forcer, forcerClient := makeTestChar("Forcer")
	defer forcerClient.Close()
	forcer.Level = 55
	handler.CharToRoom(forcer, room)

	// Target at trust 52 (lower than forcer, so force applies)
	target, targetClient := makeTestChar("Target")
	defer targetClient.Close()
	target.Level = 52
	handler.CharToRoom(target, room)
	w.Descriptors = []*types.DescriptorData{forcer.Desc, target.Desc}

	// "force all highcmd" — highcmd needs trust 58, forcer has 55, so capped
	DoForce(forcer, "all highcmd")
	_ = readOutput(forcer, forcerClient)
	_ = readOutput(target, targetClient)

	if dispatched != 0 {
		t.Fatalf("highcmd should NOT have been dispatched via force all, but was dispatched %d times", dispatched)
	}
}
