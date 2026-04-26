package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// --- test fixtures: prototype mob/obj/room with progs in a room with the builder ---

func setupMpeditFixture(name string, vnum int, progs []*types.MProgData) *types.MobIndexData {
	idx := &types.MobIndexData{
		Vnum:        vnum,
		PlayerName:  name,
		ShortDescr:  name,
		LongDescr:   name + " is here.",
		Description: "",
		MudProgs:    progs,
	}
	idx.Act.Set(types.ACT_PROTOTYPE)
	for _, p := range progs {
		idx.ProgTypes.Set(progBitIndex(p.Type))
	}
	WorldRef.MobIndex[vnum] = idx
	return idx
}

func placeMobInRoomWith(builder *types.CharData, idx *types.MobIndexData) *types.CharData {
	mob := &types.CharData{
		Name:       idx.PlayerName,
		ShortDescr: idx.ShortDescr,
		IndexData:  idx,
		Level:      10,
	}
	mob.Act.Set(types.ACT_IS_NPC)
	mob.Act.Set(types.ACT_PROTOTYPE)
	if builder.InRoom == nil {
		room := &types.RoomIndexData{Vnum: 9001, Name: "Test Room"}
		builder.InRoom = room
		room.People = append(room.People, builder)
	}
	mob.InRoom = builder.InRoom
	builder.InRoom.People = append(builder.InRoom.People, mob)
	return mob
}

func setupOpeditFixture(builder *types.CharData, vnum int, kw string, progs []*types.MProgData) *types.ObjData {
	idx := &types.ObjIndexData{
		Vnum:       vnum,
		Name:       kw,
		ShortDescr: kw,
		MudProgs:   progs,
	}
	idx.ExtraFlags.Set(types.ITEM_PROTOTYPE)
	for _, p := range progs {
		idx.ProgTypes.Set(progBitIndex(p.Type))
	}
	WorldRef.ObjIndex[vnum] = idx
	obj := &types.ObjData{
		Name:       kw,
		ShortDescr: kw,
		IndexData:  idx,
	}
	obj.ExtraFlags.Set(types.ITEM_PROTOTYPE)
	builder.Carrying = append(builder.Carrying, obj)
	return obj
}

func setupRpeditFixture(builder *types.CharData, vnum int, progs []*types.MProgData) *types.RoomIndexData {
	room := &types.RoomIndexData{
		Vnum:     vnum,
		Name:     "Rpedit Test Room",
		MudProgs: progs,
	}
	for _, p := range progs {
		room.ProgTypes.Set(progBitIndex(p.Type))
	}
	WorldRef.Rooms[vnum] = room
	builder.InRoom = room
	room.People = append(room.People, builder)
	return room
}

// --- G3: dispatcher gates ---

func TestMpeditDispatch_NpcCannotEdit(t *testing.T) {
	_ = setupOlcWorld()
	mob, client := makeTestChar("MobBuilder")
	defer client.Close()
	mob.Act.Set(types.ACT_IS_NPC)
	mob.Level = types.LEVEL_IMMORTAL

	DoMpedit(mob, "fido list")
	out := readOutput(mob, client)
	if !strings.Contains(out, "Mob's can't mpedit") {
		t.Errorf("expected NPC reject, got: %q", out)
	}
}

func TestOpeditDispatch_NpcCannotEdit(t *testing.T) {
	_ = setupOlcWorld()
	mob, client := makeTestChar("MobBuilder")
	defer client.Close()
	mob.Act.Set(types.ACT_IS_NPC)
	mob.Level = types.LEVEL_IMMORTAL

	DoOpedit(mob, "sword list")
	out := readOutput(mob, client)
	if !strings.Contains(out, "Mob's can't opedit") {
		t.Errorf("expected NPC reject, got: %q", out)
	}
}

func TestRpeditDispatch_NpcCannotEdit(t *testing.T) {
	_ = setupOlcWorld()
	mob, client := makeTestChar("MobBuilder")
	defer client.Close()
	mob.Act.Set(types.ACT_IS_NPC)
	mob.Level = types.LEVEL_IMMORTAL

	DoRpedit(mob, "list")
	out := readOutput(mob, client)
	if !strings.Contains(out, "Mob's can't rpedit") {
		t.Errorf("expected NPC reject, got: %q", out)
	}
}

func TestMpeditDispatch_NoDesc(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.Desc = nil

	DoMpedit(ch, "fido list")
	// Output goes nowhere (Desc is nil); we can only assert non-panic.
	// But we can also call Send which is a no-op without Desc — there's no
	// observable output. Use a separate descriptor-attached char to confirm.
	_ = ch
}

func TestMpeditDispatch_SyntaxHelp(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoMpedit(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Syntax:") {
		t.Errorf("expected syntax help, got: %q", out)
	}
	if !strings.Contains(out, "add delete insert edit list") {
		t.Errorf("expected command list, got: %q", out)
	}
	if !strings.Contains(out, "act speech") {
		t.Errorf("expected program list, got: %q", out)
	}
}

func TestRpeditDispatch_SyntaxHelp_NoArg(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoRpedit(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Syntax:") {
		t.Errorf("expected syntax help, got: %q", out)
	}
	if !strings.Contains(out, "standing in room") {
		t.Errorf("expected rpedit-specific footer, got: %q", out)
	}
}

func TestMpeditDispatch_VictimNotInRoom(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.Level = types.LEVEL_IMMORTAL // < LEVEL_GOD
	ch.InRoom = &types.RoomIndexData{Vnum: 1, Name: "Empty"}

	DoMpedit(ch, "fido list")
	out := readOutput(ch, client)
	if !strings.Contains(out, "They aren't here") {
		t.Errorf("expected room-only reject, got: %q", out)
	}
}

func TestMpeditDispatch_VictimNotInWorld(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.Level = types.LEVEL_GOD

	DoMpedit(ch, "nonexistent list")
	out := readOutput(ch, client)
	if !strings.Contains(out, "all the realms") {
		t.Errorf("expected world-search miss, got: %q", out)
	}
}

func TestMpeditDispatch_NotPrototype(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := &types.MobIndexData{Vnum: 1000, PlayerName: "fido", ShortDescr: "fido"}
	// NOTE: ACT_PROTOTYPE NOT set.
	WorldRef.MobIndex[1000] = idx
	mob := placeMobInRoomWith(ch, idx)
	mob.Act.Remove(types.ACT_PROTOTYPE) // clear the auto-set in helper.
	mob.IndexData.Act.Remove(types.ACT_PROTOTYPE)

	DoMpedit(ch, "fido list")
	out := readOutput(ch, client)
	if !strings.Contains(out, "prototype flag to be mpset") {
		t.Errorf("expected prototype gate reject, got: %q", out)
	}
}

func TestMpeditDispatch_StatshieldGate(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	ch.Level = types.LEVEL_IMMORTAL // < LEVEL_GREATER
	idx := setupMpeditFixture("fido", 1001, []*types.MProgData{{Type: types.MPROG_ACT, ComList: "x"}})
	mob := placeMobInRoomWith(ch, idx)
	mob.Act.Set(types.ACT_STATSHIELD)

	DoMpedit(ch, "fido list")
	out := readOutput(ch, client)
	if !strings.Contains(out, "godly glow") {
		t.Errorf("expected statshield reject, got: %q", out)
	}
}

func TestMpeditDispatch_PcVictim(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	// Place a PC (non-NPC) in the room.
	pc, pcClient := makeTestChar("Frodo")
	defer pcClient.Close()
	room := &types.RoomIndexData{Vnum: 1, Name: "Test"}
	ch.InRoom = room
	pc.InRoom = room
	room.People = append(room.People, ch, pc)

	DoMpedit(ch, "frodo list")
	out := readOutput(ch, client)
	if !strings.Contains(out, "You can't do that") {
		t.Errorf("expected PC-victim reject, got: %q", out)
	}
}

// --- G4: list subcommand ---

func TestMpeditList_Empty(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 1002, nil)
	placeMobInRoomWith(ch, idx)

	DoMpedit(ch, "fido list")
	out := readOutput(ch, client)
	if !strings.Contains(out, "No programs on mobile") {
		t.Errorf("expected empty-list message, got: %q", out)
	}
}

func TestOpeditList_Empty(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	setupOpeditFixture(ch, 2001, "sword", nil)

	DoOpedit(ch, "sword list")
	out := readOutput(ch, client)
	// C-bug verbatim wording on opedit: "mob programs".
	if !strings.Contains(out, "no mob programs") {
		t.Errorf("expected opedit empty-list message, got: %q", out)
	}
}

func TestRpeditList_Empty(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	setupRpeditFixture(ch, 3001, nil)

	DoRpedit(ch, "list")
	out := readOutput(ch, client)
	if !strings.Contains(out, "no mob programs") {
		t.Errorf("expected rpedit empty-list message, got: %q", out)
	}
}

func TestMpeditList_HeaderOnly(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 1003, []*types.MProgData{
		{Type: types.MPROG_ACT, ArgList: "p hi", ComList: "say A"},
		{Type: types.MPROG_GREET, ArgList: "100", ComList: "say B"},
		{Type: types.MPROG_DEATH, ArgList: "*", ComList: "say C"},
	})
	placeMobInRoomWith(ch, idx)

	DoMpedit(ch, "fido list 0")
	out := readOutput(ch, client)
	for _, want := range []string{"1>act", "2>greet", "3>death"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q, got: %q", want, out)
		}
	}
	// Bodies should NOT be present (header-only).
	if strings.Contains(out, "say A") || strings.Contains(out, "say B") {
		t.Errorf("header-only listing leaked bodies: %q", out)
	}
}

func TestMpeditList_Full(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 1004, []*types.MProgData{
		{Type: types.MPROG_ACT, ArgList: "p hi", ComList: "say BODY-A"},
		{Type: types.MPROG_GREET, ArgList: "100", ComList: "say BODY-B"},
	})
	placeMobInRoomWith(ch, idx)

	DoMpedit(ch, "fido list full")
	out := readOutput(ch, client)
	if !strings.Contains(out, "BODY-A") || !strings.Contains(out, "BODY-B") {
		t.Errorf("expected both bodies, got: %q", out)
	}
}

func TestMpeditList_ByIndex(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 1005, []*types.MProgData{
		{Type: types.MPROG_ACT, ArgList: "x", ComList: "BODY1"},
		{Type: types.MPROG_GREET, ArgList: "y", ComList: "BODY2"},
		{Type: types.MPROG_DEATH, ArgList: "z", ComList: "BODY3"},
	})
	placeMobInRoomWith(ch, idx)

	DoMpedit(ch, "fido list 2")
	out := readOutput(ch, client)
	if !strings.Contains(out, "BODY2") {
		t.Errorf("expected BODY2 (1-based index 2), got: %q", out)
	}
	if strings.Contains(out, "BODY1") || strings.Contains(out, "BODY3") {
		t.Errorf("expected only the 2nd prog body, got: %q", out)
	}
}

func TestMpeditList_OutOfRange(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 1006, []*types.MProgData{
		{Type: types.MPROG_ACT, ComList: "x"},
	})
	placeMobInRoomWith(ch, idx)

	DoMpedit(ch, "fido list 99")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Program not found") {
		t.Errorf("expected not-found, got: %q", out)
	}
}

func TestRpeditList_HeaderOnly(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	setupRpeditFixture(ch, 3002, []*types.MProgData{
		{Type: types.MPROG_ACT, ArgList: "x", ComList: "BODY-X"},
		{Type: types.MPROG_GREET, ArgList: "y", ComList: "BODY-Y"},
	})

	DoRpedit(ch, "list 0")
	out := readOutput(ch, client)
	if !strings.Contains(out, "1>act") || !strings.Contains(out, "2>greet") {
		t.Errorf("expected rpedit header lines, got: %q", out)
	}
	if strings.Contains(out, "BODY-X") {
		t.Errorf("rpedit header-only leaked body: %q", out)
	}
}

// Mortal-reject regression (was the inspector test).
func TestProgEditors_MortalReject(t *testing.T) {
	_ = setupOlcWorld()
	mortal, client := makeTestChar("Mortal")
	defer client.Close()
	mortal.Level = 1

	for _, fn := range []func(*types.CharData, string){DoMpedit, DoOpedit, DoRpedit} {
		fn(mortal, "anything list")
		out := readOutput(mortal, client)
		if !strings.Contains(out, "Huh?") {
			t.Errorf("expected Huh?, got: %q", out)
		}
	}
}
