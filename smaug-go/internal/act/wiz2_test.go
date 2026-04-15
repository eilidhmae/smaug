package act

import (
	"bytes"
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/command"
	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
)

// npcActFlag returns a BitVector with ACT_IS_NPC set.
func npcActFlag() types.BitVector {
	var bv types.BitVector
	bv.Set(types.ACT_IS_NPC)
	return bv
}

// --- Switch / Return round-trip ---

func TestDoSwitch_NoArg(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()
	DoSwitch(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Switch into whom?") {
		t.Errorf("got %q", out)
	}
}

func TestDoSwitchReturn_RoundTrip(t *testing.T) {
	w := setupWizWorld()
	imm, immClient := makeTestChar("Imm")
	defer immClient.Close()
	imm.Level = types.LEVEL_IMMORTAL
	imm.Trust = types.LEVEL_IMMORTAL

	// An NPC victim (same room) — no descriptor of its own.
	room := &types.RoomIndexData{Vnum: 1001, Name: "Arena"}
	imm.InRoom = room
	room.People = append(room.People, imm)

	mob := &types.CharData{
		Name:       "goblin",
		ShortDescr: "a goblin",
		Act:        npcActFlag(),
		InRoom:     room,
	}
	room.People = append(room.People, mob)
	w.Characters = append(w.Characters, imm, mob)

	origDesc := imm.Desc
	DoSwitch(imm, "goblin")

	if imm.Desc != nil {
		t.Fatalf("imm.Desc should be nil after switch")
	}
	if mob.Desc != origDesc {
		t.Fatalf("mob should have taken over the descriptor")
	}
	if origDesc.Original != imm {
		t.Fatalf("descriptor.Original should be set to imm")
	}
	if origDesc.Character != mob {
		t.Fatalf("descriptor.Character should be mob")
	}
	if imm.Switched != mob {
		t.Fatalf("imm.Switched should point to mob")
	}

	// Now return. Call with ch=mob (the puppet).
	DoReturn(mob, "")

	if imm.Desc != origDesc {
		t.Fatalf("imm.Desc should be restored")
	}
	if mob.Desc != nil {
		t.Fatalf("mob.Desc should be nil")
	}
	if origDesc.Original != nil {
		t.Fatalf("descriptor.Original should be nil")
	}
	if origDesc.Character != imm {
		t.Fatalf("descriptor.Character should be imm")
	}
	if imm.Switched != nil {
		t.Fatalf("imm.Switched should be nil")
	}
}

func TestDoSwitch_AlreadySwitched(t *testing.T) {
	_ = setupWizWorld()
	imm, immClient := makeTestChar("Imm")
	defer immClient.Close()
	imm.Desc.Original = imm // pretend already switched
	DoSwitch(imm, "someone")
	// readOutput uses imm.Desc, which is still present
	out := readOutput(imm, immClient)
	if !strings.Contains(out, "already switched") {
		t.Errorf("got %q", out)
	}
}

// --- Wizlock ---

func TestDoWizlock_Toggles(t *testing.T) {
	w := setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()
	if w.SysData.Wizlock {
		t.Fatalf("precondition: wizlock should start false")
	}
	DoWizlock(ch, "")
	if !w.SysData.Wizlock {
		t.Fatalf("wizlock should be on")
	}
	DoWizlock(ch, "")
	if w.SysData.Wizlock {
		t.Fatalf("wizlock should be off again")
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "Wizlock is now OFF") {
		t.Errorf("got %q", out)
	}
}

// --- Shutdown / Reboot ---

func TestDoShutdown_CallsHook(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()
	called := false
	gotReboot := false
	ShutdownFunc = func(r bool) { called = true; gotReboot = r }
	defer func() { ShutdownFunc = nil }()
	DoShutdown(ch, "")
	if !called || gotReboot {
		t.Errorf("shutdown should fire with reboot=false")
	}
}

func TestDoReboot_CallsHook(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()
	gotReboot := false
	ShutdownFunc = func(r bool) { gotReboot = r }
	defer func() { ShutdownFunc = nil }()
	DoReboot(ch, "")
	if !gotReboot {
		t.Errorf("reboot should fire with reboot=true")
	}
}

// --- Wizhelp ---

func TestDoWizhelp_ListsImmortalCommands(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()
	ch.Trust = types.LEVEL_IMMORTAL
	ch.PCData.Flags = int(types.PCFLAG_PAGERON) // force pager path

	reg := command.NewRegistry()
	reg.Register(&command.Command{Name: "look", DoFun: func(*types.CharData, string) {}, Position: 0, Level: 0})
	reg.Register(&command.Command{Name: "goto", DoFun: func(*types.CharData, string) {}, Position: 0, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "slay", DoFun: func(*types.CharData, string) {}, Position: 0, Level: types.LEVEL_IMMORTAL})
	CmdRegistry = reg

	DoWizhelp(ch, "")
	pager := ch.Desc.GetPagerData()
	if !strings.Contains(pager, "goto") || !strings.Contains(pager, "slay") {
		t.Errorf("expected immortal commands in pager: %q", pager)
	}
	if strings.Contains(pager, "look") {
		t.Errorf("should NOT contain mortal commands: %q", pager)
	}
}

// --- Aecho ---

func TestDoAecho_BroadcastsInArea(t *testing.T) {
	w := setupWizWorld()
	area := &types.AreaData{Name: "TestArea"}
	rIn := &types.RoomIndexData{Vnum: 100, Area: area}
	rIn2 := &types.RoomIndexData{Vnum: 101, Area: area}
	rOther := &types.RoomIndexData{Vnum: 200, Area: &types.AreaData{Name: "Other"}}

	imm, immC := makeTestChar("Imm")
	defer immC.Close()
	imm.InRoom = rIn

	inSameArea, ac := makeTestChar("Ally")
	defer ac.Close()
	inSameArea.InRoom = rIn2

	outside, oc := makeTestChar("Stranger")
	defer oc.Close()
	outside.InRoom = rOther

	w.Characters = append(w.Characters, imm, inSameArea, outside)

	DoAecho(imm, "A rumbling echo.")

	if !strings.Contains(readOutput(inSameArea, ac), "A rumbling echo.") {
		t.Errorf("same-area char should hear aecho")
	}
	if strings.Contains(readOutput(outside, oc), "A rumbling echo.") {
		t.Errorf("outside char should NOT hear aecho")
	}
}

func TestDoAecho_NoArg(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeTestChar("Imm")
	defer client.Close()
	DoAecho(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Aecho what?") {
		t.Errorf("got %q", out)
	}
}

// --- Hell ---

func TestDoHell_SetsTimestamp(t *testing.T) {
	w := setupWizWorld()
	hell := &types.RoomIndexData{Vnum: HellRoomVnum, Name: "Hell"}
	start := &types.RoomIndexData{Vnum: 50, Name: "Start"}
	w.Rooms[HellRoomVnum] = hell
	w.Rooms[50] = start

	imm, ic := makeTestChar("Imm")
	defer ic.Close()
	imm.Trust = types.LEVEL_IMMORTAL
	imm.InRoom = start
	start.People = append(start.People, imm)

	victim, vc := makeTestChar("Sinner")
	defer vc.Close()
	victim.Trust = 10
	victim.InRoom = start
	start.People = append(start.People, victim)
	w.Characters = append(w.Characters, imm, victim)

	DoHell(imm, "Sinner 2")

	if victim.PCData.Hell == 0 {
		t.Errorf("Hell timestamp should be set")
	}
	if victim.PCData.HelledBy != "Imm" {
		t.Errorf("HelledBy should be 'Imm', got %q", victim.PCData.HelledBy)
	}
	if victim.InRoom != hell {
		t.Errorf("victim should be in hell room")
	}
}

func TestDoHell_NoHellRoom(t *testing.T) {
	w := setupWizWorld()
	// Do NOT register hell room
	imm, ic := makeTestChar("Imm")
	defer ic.Close()
	imm.Trust = types.LEVEL_IMMORTAL

	victim, _ := makeTestChar("Sinner")
	victim.Trust = 10
	w.Characters = append(w.Characters, imm, victim)

	DoHell(imm, "Sinner")
	out := readOutput(imm, ic)
	if !strings.Contains(out, "No hell room defined") {
		t.Errorf("expected missing-hell message, got %q", out)
	}
	// Regression guard: when the hell room is missing, the command must NOT
	// leave the victim flagged as helled. Pre-fix bug wrote Hell+HelledBy
	// before the room check, stranding players in ghost-hell forever.
	if victim.PCData.Hell != 0 {
		t.Errorf("Hell must not be set when hell room missing, got %d", victim.PCData.Hell)
	}
	if victim.PCData.HelledBy != "" {
		t.Errorf("HelledBy must not be set when hell room missing, got %q", victim.PCData.HelledBy)
	}
}

func TestDoHell_RejectsImmortal(t *testing.T) {
	w := setupWizWorld()
	hell := &types.RoomIndexData{Vnum: HellRoomVnum, Name: "Hell"}
	w.Rooms[HellRoomVnum] = hell

	imm, ic := makeTestChar("Imm")
	defer ic.Close()
	imm.Trust = types.LEVEL_SUPREME

	peer, _ := makeTestChar("Peer")
	peer.Trust = types.LEVEL_IMMORTAL
	w.Characters = append(w.Characters, imm, peer)

	DoHell(imm, "Peer 2")
	out := readOutput(imm, ic)
	if !strings.Contains(out, "no point in helling an immortal") {
		t.Errorf("expected immortal-guard message, got %q", out)
	}
	if peer.PCData != nil && peer.PCData.Hell != 0 {
		t.Errorf("immortal should not be helled")
	}
}

func TestDoHell_RejectsAlreadyHelled(t *testing.T) {
	w := setupWizWorld()
	hell := &types.RoomIndexData{Vnum: HellRoomVnum, Name: "Hell"}
	w.Rooms[HellRoomVnum] = hell

	imm, ic := makeTestChar("Imm")
	defer ic.Close()
	imm.Trust = types.LEVEL_IMMORTAL

	victim, _ := makeTestChar("Sinner")
	victim.Trust = 10
	victim.PCData.Hell = 9999999999
	victim.PCData.HelledBy = "SomeoneElse"
	w.Characters = append(w.Characters, imm, victim)

	DoHell(imm, "Sinner")
	out := readOutput(imm, ic)
	if !strings.Contains(out, "already in hell") {
		t.Errorf("expected already-in-hell message, got %q", out)
	}
	if victim.PCData.HelledBy != "SomeoneElse" {
		t.Errorf("HelledBy must not be overwritten, got %q", victim.PCData.HelledBy)
	}
}

// --- Hell persistence round-trip ---

func TestHell_Persists(t *testing.T) {
	ch := &types.CharData{
		Name:    "Someone",
		Trust:   0,
		Level:   1,
		Hit:     1, MaxHit: 1, Mana: 1, MaxMana: 1, Move: 1, MaxMove: 1,
		PCData: &types.PCData{
			Pwd:      "$2a$10$abcdef",
			Hell:     1713000000,
			HelledBy: "Judge",
			PagerLen: 24,
		},
	}

	var buf bytes.Buffer
	if err := persist.SavePlayer(&buf, ch); err != nil {
		t.Fatalf("save: %v", err)
	}
	if !strings.Contains(buf.String(), "Hell       1713000000") {
		t.Errorf("Hell not persisted:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "HelledBy   Judge~") {
		t.Errorf("HelledBy not persisted:\n%s", buf.String())
	}

	reloaded, err := persist.LoadPlayer(&buf, "mem")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if reloaded.PCData.Hell != 1713000000 {
		t.Errorf("Hell = %d, want 1713000000", reloaded.PCData.Hell)
	}
	if reloaded.PCData.HelledBy != "Judge" {
		t.Errorf("HelledBy = %q", reloaded.PCData.HelledBy)
	}
}

// --- Log flag ---

func TestDoLog_TogglesFlag(t *testing.T) {
	w := setupWizWorld()
	imm, _ := makeTestChar("Imm")
	imm.Trust = types.LEVEL_IMMORTAL

	victim, _ := makeTestChar("Mortal")
	victim.Trust = 5
	w.Characters = append(w.Characters, imm, victim)

	if victim.Act.IsSet(types.PLR_LOG) {
		t.Fatalf("PLR_LOG should start clear")
	}
	DoLog(imm, "Mortal")
	if !victim.Act.IsSet(types.PLR_LOG) {
		t.Errorf("PLR_LOG should be set")
	}
	DoLog(imm, "Mortal")
	if victim.Act.IsSet(types.PLR_LOG) {
		t.Errorf("PLR_LOG should be cleared")
	}
}

// --- Deny / Pardon ---

func TestDoDeny_SetsFlag(t *testing.T) {
	w := setupWizWorld()
	imm, _ := makeTestChar("Imm")
	imm.Trust = types.LEVEL_IMMORTAL
	victim, _ := makeTestChar("Baddie")
	victim.Trust = 2
	w.Characters = append(w.Characters, imm, victim)
	DoDeny(imm, "Baddie")
	if !victim.Act.IsSet(types.PLR_DENY) {
		t.Errorf("PLR_DENY should be set")
	}
}

func TestDoPardon_ClearsDeny(t *testing.T) {
	w := setupWizWorld()
	imm, _ := makeTestChar("Imm")
	imm.Trust = types.LEVEL_IMMORTAL
	victim, _ := makeTestChar("Baddie")
	victim.Trust = 2
	victim.Act.Set(types.PLR_DENY)
	w.Characters = append(w.Characters, imm, victim)
	DoPardon(imm, "Baddie deny")
	if victim.Act.IsSet(types.PLR_DENY) {
		t.Errorf("PLR_DENY should be cleared")
	}
}

func TestDoPardon_ClearsKiller(t *testing.T) {
	w := setupWizWorld()
	imm, _ := makeTestChar("Imm")
	imm.Trust = types.LEVEL_IMMORTAL
	victim, _ := makeTestChar("Baddie")
	victim.Trust = 2
	victim.Act.Set(types.PLR_KILLER)
	w.Characters = append(w.Characters, imm, victim)
	DoPardon(imm, "Baddie killer")
	if victim.Act.IsSet(types.PLR_KILLER) {
		t.Errorf("PLR_KILLER should be cleared")
	}
}

// --- Disconnect ---

func TestDoDisconnect_CallsHook(t *testing.T) {
	w := setupWizWorld()
	imm, ic := makeTestChar("Imm")
	defer ic.Close()
	imm.Trust = types.LEVEL_IMMORTAL

	victim, vc := makeTestChar("Target")
	defer vc.Close()
	victim.Trust = 2
	w.Characters = append(w.Characters, imm, victim)

	disconnected := false
	DisconnectFunc = func(d *types.DescriptorData) { disconnected = true }
	defer func() { DisconnectFunc = nil }()

	DoDisconnect(imm, "Target")
	if !disconnected {
		t.Errorf("DisconnectFunc should fire")
	}
}

// --- Mortalize ---

func TestDoMortalize_ClearsTrust(t *testing.T) {
	w := setupWizWorld()
	imm, _ := makeTestChar("Imm")
	imm.Trust = types.LEVEL_SUPREME

	victim, _ := makeTestChar("Godling")
	victim.Trust = types.LEVEL_IMMORTAL
	victim.Act.Set(types.PLR_HOLYLIGHT)
	victim.Act.Set(types.PLR_WIZINVIS)
	w.Characters = append(w.Characters, imm, victim)

	DoMortalize(imm, "Godling")
	if victim.Trust != 0 {
		t.Errorf("trust should be zero, got %d", victim.Trust)
	}
	if victim.Act.IsSet(types.PLR_HOLYLIGHT) {
		t.Errorf("holylight should be cleared")
	}
	if victim.Act.IsSet(types.PLR_WIZINVIS) {
		t.Errorf("wizinvis should be cleared")
	}
}

// --- Mpstat / Opstat / Rpstat ---

func TestDoMpstat_ShowsProgs(t *testing.T) {
	w := setupWizWorld()
	imm, ic := makeTestChar("Imm")
	defer ic.Close()
	imm.Trust = types.LEVEL_IMMORTAL

	idx := &types.MobIndexData{
		Vnum:       501,
		PlayerName: "goblin",
		ShortDescr: "a goblin",
		MudProgs: []*types.MProgData{
			{Type: types.MPROG_SPEECH, ArgList: "p hello", ComList: "say hi"},
		},
	}
	mob := &types.CharData{
		Name:       "goblin",
		ShortDescr: "a goblin",
		Act:        npcActFlag(),
		IndexData:  idx,
	}
	w.Characters = append(w.Characters, imm, mob)

	DoMpstat(imm, "goblin")
	out := readOutput(imm, ic)
	pagerOut := imm.Desc.GetPagerData()
	combined := out + pagerOut
	if !strings.Contains(combined, "speech") {
		t.Errorf("mpstat output missing prog-type: %q", combined)
	}
	if !strings.Contains(combined, "say hi") {
		t.Errorf("mpstat output missing comlist first line: %q", combined)
	}
}

func TestDoMpstat_OnPlayer(t *testing.T) {
	w := setupWizWorld()
	imm, ic := makeTestChar("Imm")
	defer ic.Close()
	imm.Trust = types.LEVEL_IMMORTAL
	other, _ := makeTestChar("Pc")
	w.Characters = append(w.Characters, imm, other)
	DoMpstat(imm, "Pc")
	out := readOutput(imm, ic)
	if !strings.Contains(out, "Only Mobiles") {
		t.Errorf("got %q", out)
	}
}

func TestDoRpstat_ShowsRoomProgs(t *testing.T) {
	_ = setupWizWorld()
	imm, ic := makeTestChar("Imm")
	defer ic.Close()
	imm.Trust = types.LEVEL_IMMORTAL
	room := &types.RoomIndexData{
		Vnum: 777, Name: "Haunted Hall",
		MudProgs: []*types.MProgData{
			{Type: types.MPROG_ENTER, ArgList: "100", ComList: "echo boo"},
		},
	}
	imm.InRoom = room

	DoRpstat(imm, "")
	combined := readOutput(imm, ic) + imm.Desc.GetPagerData()
	if !strings.Contains(combined, "entry") {
		t.Errorf("rpstat missing entry keyword: %q", combined)
	}
	if !strings.Contains(combined, "echo boo") {
		t.Errorf("rpstat missing comlist: %q", combined)
	}
}

func TestDoRpstat_NoProgs(t *testing.T) {
	_ = setupWizWorld()
	imm, ic := makeTestChar("Imm")
	defer ic.Close()
	imm.InRoom = &types.RoomIndexData{Vnum: 1}
	DoRpstat(imm, "")
	out := readOutput(imm, ic)
	if !strings.Contains(out, "no programs") {
		t.Errorf("got %q", out)
	}
}

// --- slookup / sset (Phase 5 Tier 4 G4) ---

func TestDoSlookup_NoArg(t *testing.T) {
	_ = setupWizWorld()
	imm, ic := makeTestChar("Imm")
	defer ic.Close()
	DoSlookup(imm, "")
	out := readOutput(imm, ic)
	if !strings.Contains(out, "Syntax") {
		t.Errorf("expected syntax help, got %q", out)
	}
}

func TestDoSlookup_Unknown(t *testing.T) {
	_ = setupWizWorld()
	imm, ic := makeTestChar("Imm")
	defer ic.Close()
	DoSlookup(imm, "_absolutely_not_a_skill_")
	out := readOutput(imm, ic)
	if !strings.Contains(out, "No such") {
		t.Errorf("expected not-found, got %q", out)
	}
}

func TestDoSlookup_KnownSkill(t *testing.T) {
	_ = setupWizWorld()
	// Ensure we have a known skill registered
	idx := ensureSkill(t, "bite")
	sk := WorldRef.Skills[idx]
	sk.Slot = 42
	sk.MinMana = 5
	imm, ic := makeTestChar("Imm")
	defer ic.Close()
	DoSlookup(imm, "bite")
	out := readOutput(imm, ic)
	if !strings.Contains(out, "bite") {
		t.Errorf("expected skill name in output, got %q", out)
	}
}

func TestDoSset_NoArg(t *testing.T) {
	_ = setupWizWorld()
	imm, ic := makeTestChar("Imm")
	defer ic.Close()
	DoSset(imm, "")
	out := readOutput(imm, ic)
	if !strings.Contains(out, "Syntax") {
		t.Errorf("expected syntax help, got %q", out)
	}
}

func TestDoSset_SetBeats(t *testing.T) {
	_ = setupWizWorld()
	idx := ensureSkill(t, "punch")
	imm, ic := makeTestChar("Imm")
	defer ic.Close()
	DoSset(imm, "punch beats 12")
	_ = readOutput(imm, ic)
	if WorldRef.Skills[idx].Beats != 12 {
		t.Errorf("expected Beats=12, got %d", WorldRef.Skills[idx].Beats)
	}
}

func TestDoSset_SetMinMana(t *testing.T) {
	_ = setupWizWorld()
	idx := ensureSkill(t, "cleave")
	imm, ic := makeTestChar("Imm")
	defer ic.Close()
	DoSset(imm, "cleave minmana 7")
	_ = readOutput(imm, ic)
	if WorldRef.Skills[idx].MinMana != 7 {
		t.Errorf("expected MinMana=7, got %d", WorldRef.Skills[idx].MinMana)
	}
}

