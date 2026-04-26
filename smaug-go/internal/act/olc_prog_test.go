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

// --- G5: add subcommand ---

func TestMpeditAdd_AppendsAtTail(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 5001, []*types.MProgData{
		{Type: types.MPROG_ACT, ArgList: "x", ComList: "say A"},
		{Type: types.MPROG_SPEECH, ArgList: "y", ComList: "say B"},
	})
	placeMobInRoomWith(ch, idx)

	prevStart := StartEditingFunc
	StartEditingFunc = func(c *types.CharData, text string) {}
	defer func() { StartEditingFunc = prevStart }()

	DoMpedit(ch, "fido add greet hello there")
	_ = readOutput(ch, client)

	if len(idx.MudProgs) != 3 {
		t.Fatalf("expected 3 progs after add, got %d", len(idx.MudProgs))
	}
	added := idx.MudProgs[2]
	if added.Type != types.MPROG_GREET {
		t.Errorf("expected MPROG_GREET, got 0x%x", added.Type)
	}
	if added.ArgList != "hello there" {
		t.Errorf("expected arglist 'hello there', got %q", added.ArgList)
	}
	if !idx.ProgTypes.IsSet(progBitIndex(types.MPROG_GREET)) {
		t.Errorf("progtypes bit not set for greet")
	}
}

func TestMpeditAdd_OpensEditor(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 5002, nil)
	placeMobInRoomWith(ch, idx)

	var seedSeen string
	var saveAtCall func(*types.CharData)
	prevStart := StartEditingFunc
	StartEditingFunc = func(c *types.CharData, text string) {
		seedSeen = text
		saveAtCall = c.EditorSave
	}
	defer func() { StartEditingFunc = prevStart }()

	DoMpedit(ch, "fido add act sample args")
	_ = readOutput(ch, client)

	if seedSeen != "" {
		t.Errorf("expected empty seed for new prog, got %q", seedSeen)
	}
	if saveAtCall == nil {
		t.Fatal("EditorSave must be installed before StartEditingFunc is called")
	}
	if ch.Substate != types.SUB_MPROG_EDIT {
		t.Errorf("expected Substate=SUB_MPROG_EDIT (%d), got %d", types.SUB_MPROG_EDIT, ch.Substate)
	}
}

func TestMpeditAdd_EditorSaveWritesComList(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 5003, nil)
	placeMobInRoomWith(ch, idx)

	prevStart, prevCopy, prevStop := StartEditingFunc, CopyBufferFunc, StopEditingFunc
	defer func() {
		StartEditingFunc, CopyBufferFunc, StopEditingFunc = prevStart, prevCopy, prevStop
	}()
	StartEditingFunc = func(c *types.CharData, text string) {}
	CopyBufferFunc = func(c *types.CharData) string { return "say hi $n\n\r" }
	StopEditingFunc = func(c *types.CharData) {
		c.Substate = types.SUB_NONE
		c.EditorSave = nil
	}

	DoMpedit(ch, "fido add greet hi $n")
	_ = readOutput(ch, client)
	if ch.EditorSave == nil {
		t.Fatal("EditorSave should be set")
	}
	ch.EditorSave(ch)

	if len(idx.MudProgs) != 1 {
		t.Fatalf("expected 1 prog, got %d", len(idx.MudProgs))
	}
	if idx.MudProgs[0].ComList != "say hi $n\n\r" {
		t.Errorf("ComList not written by EditorSave: got %q", idx.MudProgs[0].ComList)
	}
}

func TestMpeditAdd_UnknownType(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 5004, nil)
	placeMobInRoomWith(ch, idx)

	DoMpedit(ch, "fido add bogus_type sample")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Unknown program type") {
		t.Errorf("expected unknown-type reject, got: %q", out)
	}
	if len(idx.MudProgs) != 0 {
		t.Errorf("progs should not have been mutated")
	}
}

func TestMpeditAdd_SmashesTilde(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 5005, nil)
	placeMobInRoomWith(ch, idx)

	prevStart := StartEditingFunc
	StartEditingFunc = func(c *types.CharData, text string) {}
	defer func() { StartEditingFunc = prevStart }()

	DoMpedit(ch, "fido add act hi~there~ok")
	_ = readOutput(ch, client)
	if len(idx.MudProgs) != 1 {
		t.Fatalf("expected 1 prog, got %d", len(idx.MudProgs))
	}
	if strings.Contains(idx.MudProgs[0].ArgList, "~") {
		t.Errorf("tildes not smashed: %q", idx.MudProgs[0].ArgList)
	}
}

func TestMpeditAdd_DoubleSetIdempotent(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 5006, []*types.MProgData{
		{Type: types.MPROG_GREET, ArgList: "x", ComList: "first"},
	})
	placeMobInRoomWith(ch, idx)

	prevStart := StartEditingFunc
	StartEditingFunc = func(c *types.CharData, text string) {}
	defer func() { StartEditingFunc = prevStart }()

	// Add a second greet — bit should remain set, not toggle off.
	DoMpedit(ch, "fido add greet second")
	_ = readOutput(ch, client)
	if !idx.ProgTypes.IsSet(progBitIndex(types.MPROG_GREET)) {
		t.Errorf("greet bit lost after second add (Set vs Toggle mutation gate)")
	}
}

// --- G6: insert subcommand ---

func TestMpeditInsert_AtHead(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 6001, []*types.MProgData{
		{Type: types.MPROG_ACT, ComList: "A"},
		{Type: types.MPROG_SPEECH, ComList: "B"},
		{Type: types.MPROG_GREET, ComList: "C"},
	})
	placeMobInRoomWith(ch, idx)

	prevStart := StartEditingFunc
	StartEditingFunc = func(c *types.CharData, text string) {}
	defer func() { StartEditingFunc = prevStart }()

	DoMpedit(ch, "fido insert 1 death")
	_ = readOutput(ch, client)
	if len(idx.MudProgs) != 4 {
		t.Fatalf("expected 4 progs, got %d", len(idx.MudProgs))
	}
	if idx.MudProgs[0].Type != types.MPROG_DEATH {
		t.Errorf("expected DEATH at head, got 0x%x", idx.MudProgs[0].Type)
	}
}

func TestMpeditInsert_InMiddle(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 6002, []*types.MProgData{
		{Type: types.MPROG_ACT, ComList: "A"},
		{Type: types.MPROG_SPEECH, ComList: "B"},
		{Type: types.MPROG_GREET, ComList: "C"},
	})
	placeMobInRoomWith(ch, idx)

	prevStart := StartEditingFunc
	StartEditingFunc = func(c *types.CharData, text string) {}
	defer func() { StartEditingFunc = prevStart }()

	DoMpedit(ch, "fido insert 2 death")
	_ = readOutput(ch, client)
	if len(idx.MudProgs) != 4 {
		t.Fatalf("expected 4 progs, got %d", len(idx.MudProgs))
	}
	// Inserted at index 1 (i.e. before old position 2).
	if idx.MudProgs[1].Type != types.MPROG_DEATH {
		t.Errorf("expected DEATH at index 1, got 0x%x", idx.MudProgs[1].Type)
	}
	// Old index 1 (SPEECH) is now at index 2.
	if idx.MudProgs[2].Type != types.MPROG_SPEECH {
		t.Errorf("expected SPEECH at index 2 (shifted), got 0x%x", idx.MudProgs[2].Type)
	}
}

func TestMpeditInsert_OutOfRange(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 6003, []*types.MProgData{
		{Type: types.MPROG_ACT, ComList: "A"},
	})
	placeMobInRoomWith(ch, idx)

	DoMpedit(ch, "fido insert 99 greet")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Program not found") {
		t.Errorf("expected not-found, got: %q", out)
	}
	if len(idx.MudProgs) != 1 {
		t.Errorf("progs should not have been mutated")
	}
}

func TestMpeditInsert_AtLastPosition(t *testing.T) {
	// C-fidelity: insert at position == len(progs)+1 hits C's `&& mprg->next`
	// guard — append requires `add`, not `insert`. (Audit F4.)
	// On a 3-prog list, value=4 means "after position 4", which would be
	// past the tail; C's loop exits without splicing.
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 6004, []*types.MProgData{
		{Type: types.MPROG_ACT, ComList: "A"},
		{Type: types.MPROG_SPEECH, ComList: "B"},
		{Type: types.MPROG_GREET, ComList: "C"},
	})
	placeMobInRoomWith(ch, idx)

	DoMpedit(ch, "fido insert 4 death")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Program not found") {
		t.Errorf("expected at-last reject, got: %q", out)
	}
	if len(idx.MudProgs) != 3 {
		t.Errorf("progs should not have been mutated, got len=%d", len(idx.MudProgs))
	}
}

func TestMpeditInsert_Empty(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 6005, nil)
	placeMobInRoomWith(ch, idx)

	DoMpedit(ch, "fido insert 1 act")
	out := readOutput(ch, client)
	if !strings.Contains(out, "No programs on mobile") {
		t.Errorf("expected empty-list reject, got: %q", out)
	}
	if len(idx.MudProgs) != 0 {
		t.Errorf("progs should not have been mutated")
	}
}

func TestRpeditInsert_GateFires(t *testing.T) {
	// Q1/Q2 fix: rpedit insert was a dead branch in C (gated on arg2 instead
	// of arg1). Go-port routes correctly through the dispatcher.
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	setupRpeditFixture(ch, 6006, []*types.MProgData{
		{Type: types.MPROG_ACT, ComList: "A"},
		{Type: types.MPROG_SPEECH, ComList: "B"},
	})

	prevStart := StartEditingFunc
	StartEditingFunc = func(c *types.CharData, text string) {}
	defer func() { StartEditingFunc = prevStart }()

	DoRpedit(ch, "insert 1 death")
	_ = readOutput(ch, client)
	if len(ch.InRoom.MudProgs) != 3 {
		t.Fatalf("expected 3 progs after rpedit insert, got %d", len(ch.InRoom.MudProgs))
	}
	if ch.InRoom.MudProgs[0].Type != types.MPROG_DEATH {
		t.Errorf("expected DEATH at head, got 0x%x", ch.InRoom.MudProgs[0].Type)
	}
}

// --- G7: edit subcommand ---

func TestMpeditEdit_KeepsType(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 7001, []*types.MProgData{
		{Type: types.MPROG_ACT, ArgList: "old args", ComList: "old body"},
	})
	placeMobInRoomWith(ch, idx)

	prevStart := StartEditingFunc
	StartEditingFunc = func(c *types.CharData, text string) {}
	defer func() { StartEditingFunc = prevStart }()

	// No progName supplied → type stays MPROG_ACT, only arglist updates.
	DoMpedit(ch, "fido edit 1  new arglist")
	_ = readOutput(ch, client)
	if idx.MudProgs[0].Type != types.MPROG_ACT {
		t.Errorf("type changed unexpectedly: 0x%x", idx.MudProgs[0].Type)
	}
	if idx.MudProgs[0].ArgList != "new arglist" {
		t.Errorf("arglist not updated: %q", idx.MudProgs[0].ArgList)
	}
}

func TestMpeditEdit_OverridesType(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 7002, []*types.MProgData{
		{Type: types.MPROG_ACT, ArgList: "old", ComList: "body"},
	})
	placeMobInRoomWith(ch, idx)

	prevStart := StartEditingFunc
	StartEditingFunc = func(c *types.CharData, text string) {}
	defer func() { StartEditingFunc = prevStart }()

	DoMpedit(ch, "fido edit 1 greet new args")
	_ = readOutput(ch, client)
	if idx.MudProgs[0].Type != types.MPROG_GREET {
		t.Errorf("type not overridden: 0x%x", idx.MudProgs[0].Type)
	}
	if idx.MudProgs[0].ArgList != "new args" {
		t.Errorf("arglist not updated: %q", idx.MudProgs[0].ArgList)
	}
}

func TestMpeditEdit_RebuildsProgtypes(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 7003, []*types.MProgData{
		{Type: types.MPROG_ACT, ComList: "A"},
		{Type: types.MPROG_SPEECH, ComList: "B"},
		{Type: types.MPROG_GREET, ComList: "C"},
	})
	placeMobInRoomWith(ch, idx)
	if !idx.ProgTypes.IsSet(progBitIndex(types.MPROG_GREET)) {
		t.Fatal("setup precondition: greet bit should be set")
	}

	prevStart, prevCopy, prevStop := StartEditingFunc, CopyBufferFunc, StopEditingFunc
	defer func() {
		StartEditingFunc, CopyBufferFunc, StopEditingFunc = prevStart, prevCopy, prevStop
	}()
	StartEditingFunc = func(c *types.CharData, text string) {}
	CopyBufferFunc = func(c *types.CharData) string { return "edited body" }
	StopEditingFunc = func(c *types.CharData) {
		c.Substate = types.SUB_NONE
		c.EditorSave = nil
	}

	// Edit the greet (3rd) prog and change its type to act.
	DoMpedit(ch, "fido edit 3 act new args")
	_ = readOutput(ch, client)
	if ch.EditorSave == nil {
		t.Fatal("EditorSave should be set")
	}
	ch.EditorSave(ch)

	// After save, greet bit should be cleared (no remaining greet progs).
	if idx.ProgTypes.IsSet(progBitIndex(types.MPROG_GREET)) {
		t.Errorf("greet bit should have been cleared by rebuild")
	}
	if !idx.ProgTypes.IsSet(progBitIndex(types.MPROG_ACT)) {
		t.Errorf("act bit should remain set after rebuild")
	}
	if !idx.ProgTypes.IsSet(progBitIndex(types.MPROG_SPEECH)) {
		t.Errorf("speech bit should remain set after rebuild")
	}
}

func TestMpeditEdit_OutOfRange(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 7004, []*types.MProgData{
		{Type: types.MPROG_ACT, ArgList: "x", ComList: "x"},
	})
	placeMobInRoomWith(ch, idx)

	DoMpedit(ch, "fido edit 99 act new")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Program not found") {
		t.Errorf("expected not-found, got: %q", out)
	}
}

// --- G8: delete subcommand ---

func TestMpeditDelete_Found(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 8001, []*types.MProgData{
		{Type: types.MPROG_ACT, ComList: "A"},
		{Type: types.MPROG_SPEECH, ComList: "B"},
		{Type: types.MPROG_GREET, ComList: "C"},
	})
	placeMobInRoomWith(ch, idx)

	DoMpedit(ch, "fido delete 2")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Program removed") {
		t.Errorf("expected removed message, got: %q", out)
	}
	if len(idx.MudProgs) != 2 {
		t.Fatalf("expected 2 remaining progs, got %d", len(idx.MudProgs))
	}
	if idx.MudProgs[0].Type != types.MPROG_ACT || idx.MudProgs[1].Type != types.MPROG_GREET {
		t.Errorf("wrong remaining progs: 0x%x, 0x%x", idx.MudProgs[0].Type, idx.MudProgs[1].Type)
	}
}

func TestMpeditDelete_OutOfRange(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 8002, []*types.MProgData{
		{Type: types.MPROG_ACT, ComList: "A"},
	})
	placeMobInRoomWith(ch, idx)

	DoMpedit(ch, "fido delete 99")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Program not found") {
		t.Errorf("expected not-found, got: %q", out)
	}
	if len(idx.MudProgs) != 1 {
		t.Errorf("progs should not have been mutated")
	}
}

func TestMpeditDelete_Empty(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 8003, nil)
	placeMobInRoomWith(ch, idx)

	DoMpedit(ch, "fido delete 1")
	out := readOutput(ch, client)
	if !strings.Contains(out, "No programs on mobile") {
		t.Errorf("expected empty-list message, got: %q", out)
	}
}

func TestMpeditDelete_ClearsBitWhenLast(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 8004, []*types.MProgData{
		{Type: types.MPROG_GREET, ComList: "A"},
	})
	placeMobInRoomWith(ch, idx)
	if !idx.ProgTypes.IsSet(progBitIndex(types.MPROG_GREET)) {
		t.Fatal("setup precondition")
	}

	DoMpedit(ch, "fido delete 1")
	_ = readOutput(ch, client)
	if idx.ProgTypes.IsSet(progBitIndex(types.MPROG_GREET)) {
		t.Errorf("greet bit should have been cleared (last of its type)")
	}
}

func TestMpeditDelete_KeepsBitWhenSiblings(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 8005, []*types.MProgData{
		{Type: types.MPROG_GREET, ComList: "A"},
		{Type: types.MPROG_GREET, ComList: "B"},
	})
	placeMobInRoomWith(ch, idx)

	DoMpedit(ch, "fido delete 1")
	_ = readOutput(ch, client)
	if !idx.ProgTypes.IsSet(progBitIndex(types.MPROG_GREET)) {
		t.Errorf("greet bit should remain set (sibling exists)")
	}
}

// --- G9: substate lifecycle ---

func TestMpeditEditor_SubstateLifecycle(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 9001, nil)
	placeMobInRoomWith(ch, idx)

	prevStart, prevCopy, prevStop := StartEditingFunc, CopyBufferFunc, StopEditingFunc
	defer func() {
		StartEditingFunc, CopyBufferFunc, StopEditingFunc = prevStart, prevCopy, prevStop
	}()
	StartEditingFunc = func(c *types.CharData, text string) {}
	CopyBufferFunc = func(c *types.CharData) string { return "body" }
	StopEditingFunc = func(c *types.CharData) {
		c.Substate = types.SUB_NONE
		c.EditorSave = nil
	}

	if ch.Substate != types.SUB_NONE {
		t.Fatalf("precondition: Substate should start at SUB_NONE")
	}
	DoMpedit(ch, "fido add act sample")
	_ = readOutput(ch, client)
	if ch.Substate != types.SUB_MPROG_EDIT {
		t.Errorf("expected SUB_MPROG_EDIT during edit, got %d", ch.Substate)
	}
	ch.EditorSave(ch)
	if ch.Substate != types.SUB_NONE {
		t.Errorf("expected SUB_NONE after StopEditingFunc, got %d", ch.Substate)
	}
}

// --- G10: end-to-end cycles ---

func TestMpedit_FullCycleAddListEdit(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	idx := setupMpeditFixture("fido", 10001, nil)
	placeMobInRoomWith(ch, idx)

	prevStart, prevCopy, prevStop := StartEditingFunc, CopyBufferFunc, StopEditingFunc
	defer func() {
		StartEditingFunc, CopyBufferFunc, StopEditingFunc = prevStart, prevCopy, prevStop
	}()
	StartEditingFunc = func(c *types.CharData, text string) {}
	editorBuf := "say HELLO_E2E\n\r"
	CopyBufferFunc = func(c *types.CharData) string { return editorBuf }
	StopEditingFunc = func(c *types.CharData) {
		c.Substate = types.SUB_NONE
		c.EditorSave = nil
	}

	// add
	DoMpedit(ch, "fido add act welcome")
	_ = readOutput(ch, client)
	ch.EditorSave(ch)
	if len(idx.MudProgs) != 1 {
		t.Fatalf("after add: expected 1 prog, got %d", len(idx.MudProgs))
	}

	// list full — should show the body
	DoMpedit(ch, "fido list full")
	out := readOutput(ch, client)
	if !strings.Contains(out, "HELLO_E2E") {
		t.Errorf("list full missing body: %q", out)
	}

	// edit — change body via closure
	editorBuf = "say UPDATED_E2E\n\r"
	DoMpedit(ch, "fido edit 1 act new args")
	_ = readOutput(ch, client)
	ch.EditorSave(ch)
	if !strings.Contains(idx.MudProgs[0].ComList, "UPDATED_E2E") {
		t.Errorf("edit did not update body, got: %q", idx.MudProgs[0].ComList)
	}
}

func TestOpedit_FullCycleAddDelete(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	obj := setupOpeditFixture(ch, 10002, "sword", nil)

	prevStart, prevCopy, prevStop := StartEditingFunc, CopyBufferFunc, StopEditingFunc
	defer func() {
		StartEditingFunc, CopyBufferFunc, StopEditingFunc = prevStart, prevCopy, prevStop
	}()
	StartEditingFunc = func(c *types.CharData, text string) {}
	CopyBufferFunc = func(c *types.CharData) string { return "echo wear" }
	StopEditingFunc = func(c *types.CharData) {
		c.Substate = types.SUB_NONE
		c.EditorSave = nil
	}

	DoOpedit(ch, "sword add wear keyword")
	_ = readOutput(ch, client)
	ch.EditorSave(ch)
	if len(obj.IndexData.MudProgs) != 1 {
		t.Fatalf("expected 1 prog after add, got %d", len(obj.IndexData.MudProgs))
	}

	DoOpedit(ch, "sword delete 1")
	_ = readOutput(ch, client)
	if len(obj.IndexData.MudProgs) != 0 {
		t.Errorf("expected 0 progs after delete, got %d", len(obj.IndexData.MudProgs))
	}

	// List on empty list → C-bug "no mob programs" wording.
	DoOpedit(ch, "sword list")
	out := readOutput(ch, client)
	if !strings.Contains(out, "no mob programs") {
		t.Errorf("expected empty-list message, got: %q", out)
	}
}

func TestRpedit_FullCycleAddInsert(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	room := setupRpeditFixture(ch, 10003, nil)

	prevStart, prevCopy, prevStop := StartEditingFunc, CopyBufferFunc, StopEditingFunc
	defer func() {
		StartEditingFunc, CopyBufferFunc, StopEditingFunc = prevStart, prevCopy, prevStop
	}()
	StartEditingFunc = func(c *types.CharData, text string) {}
	CopyBufferFunc = func(c *types.CharData) string { return "body" }
	StopEditingFunc = func(c *types.CharData) {
		c.Substate = types.SUB_NONE
		c.EditorSave = nil
	}

	DoRpedit(ch, "add act first")
	_ = readOutput(ch, client)
	ch.EditorSave(ch)
	DoRpedit(ch, "add greet second")
	_ = readOutput(ch, client)
	ch.EditorSave(ch)
	if len(room.MudProgs) != 2 {
		t.Fatalf("expected 2 progs, got %d", len(room.MudProgs))
	}

	// Insert at head — Q1/Q2 fix verified end-to-end.
	DoRpedit(ch, "insert 1 death new")
	_ = readOutput(ch, client)
	ch.EditorSave(ch)
	if len(room.MudProgs) != 3 {
		t.Fatalf("expected 3 progs after insert, got %d", len(room.MudProgs))
	}
	if room.MudProgs[0].Type != types.MPROG_DEATH {
		t.Errorf("expected DEATH at head, got 0x%x", room.MudProgs[0].Type)
	}
	if room.MudProgs[1].Type != types.MPROG_ACT {
		t.Errorf("expected ACT at index 1 (shifted), got 0x%x", room.MudProgs[1].Type)
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
