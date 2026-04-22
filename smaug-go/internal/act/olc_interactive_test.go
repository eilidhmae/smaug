package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// --- DoOedit ---
// TestDoOedit_NoArg superseded by TestDoOedit_NoArgPrintsUsage below
// (G10 menu-entry extension — pins the same "no menu on empty arg"
// contract plus the Olc/Connected invariants).

func TestDoOedit_Create(t *testing.T) {
	w := setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoOedit(ch, "4000 create a magic sword")
	out := readOutput(ch, client)
	if !strings.Contains(out, "created") {
		t.Errorf("expected created, got: %q", out)
	}
	if _, ok := w.ObjIndex[4000]; !ok {
		t.Error("object 4000 should exist")
	}
}

func TestDoOedit_CreateDuplicate(t *testing.T) {
	w := setupOlcWorld()
	w.ObjIndex[4010] = &types.ObjIndexData{Vnum: 4010, Name: "existing"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoOedit(ch, "4010 create foo")
	out := readOutput(ch, client)
	if !strings.Contains(out, "already exists") {
		t.Errorf("expected already-exists, got: %q", out)
	}
}

func TestDoOedit_SetFields(t *testing.T) {
	w := setupOlcWorld()
	w.ObjIndex[4020] = &types.ObjIndexData{Vnum: 4020, Name: "obj", ShortDescr: "an obj"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	tests := []struct {
		name  string
		args  string
		check func(*types.ObjIndexData) bool
	}{
		{"name", "4020 name newname", func(o *types.ObjIndexData) bool { return o.Name == "newname" }},
		{"short", "4020 short a shiny thing", func(o *types.ObjIndexData) bool { return o.ShortDescr == "a shiny thing" }},
		{"long", "4020 long It glows.", func(o *types.ObjIndexData) bool { return o.Description == "It glows." }},
		{"type_weapon", "4020 type weapon", func(o *types.ObjIndexData) bool { return o.ItemType == types.ITEM_WEAPON }},
		{"weight", "4020 weight 25", func(o *types.ObjIndexData) bool { return o.Weight == 25 }},
		{"cost", "4020 cost 500", func(o *types.ObjIndexData) bool { return o.GoldCost == 500 }},
		{"level", "4020 level 30", func(o *types.ObjIndexData) bool { return o.Level == 30 }},
		{"value", "4020 values 2 15", func(o *types.ObjIndexData) bool { return o.Value[2] == 15 }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			DoOedit(ch, tc.args)
			_ = readOutput(ch, client)
			if !tc.check(w.ObjIndex[4020]) {
				t.Errorf("field check failed for %s", tc.name)
			}
		})
	}
}

func TestDoOedit_UnknownVnum(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	DoOedit(ch, "99999 name foo")
	out := readOutput(ch, client)
	if !strings.Contains(out, "does not exist") {
		t.Errorf("expected does-not-exist, got: %q", out)
	}
}

// --- G10: DoOedit menu-entry (CON_OEDIT) extension ---
//
// Plan-phase6-olc-oedit.md §G10 / §A4 / §A5 / §A6. These tests pin the
// split: fully-empty input keeps the existing usage line (matches C's
// "OEdit what?" behavior because ENABLE_OLC2_EXTRAS auto-create is off),
// vnum-only input enters the interactive menu, and the flat subcommand
// form (`oedit <vnum> name foo`) is preserved unchanged.

// TestDoOedit_NoArgPrintsUsage pins A6: fully empty argument does NOT
// enter the menu. Regression guard for the flat path — also pins
// mutation gate #3 (if the menu-entry branch accidentally fires on
// empty input, this test fails because Connected would flip to
// CON_OEDIT).
func TestDoOedit_NoArgPrintsUsage(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	// Seed the seam so a bug wouldn't silently hide a menu-entry.
	OeditDispMenuFunc = func(d *types.DescriptorData) {
		t.Error("OeditDispMenuFunc must NOT fire on empty argument")
	}
	defer func() { OeditDispMenuFunc = nil }()

	DoOedit(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Usage") {
		t.Errorf("expected usage, got: %q", out)
	}
	if ch.Desc != nil && ch.Desc.Connected == int(types.CON_OEDIT) {
		t.Errorf("empty arg must not transition Connected to CON_OEDIT")
	}
	if ch.Desc != nil && ch.Desc.Olc != nil {
		t.Errorf("empty arg must not allocate Olc data")
	}
}

// TestDoOedit_VnumOnlyEntersMenu pins A4: `oedit <vnum>` with no
// subcommand flips Connected to CON_OEDIT, allocates OlcData with the
// resolved *ObjIndexData as Target, and fires the menu-disp seam.
func TestDoOedit_VnumOnlyEntersMenu(t *testing.T) {
	w := setupOlcWorld()
	w.ObjIndex[4000] = &types.ObjIndexData{Vnum: 4000, Name: "a prototype", ShortDescr: "a thing"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	var dispCalled bool
	OeditDispMenuFunc = func(d *types.DescriptorData) {
		dispCalled = true
		// Verify the descriptor has the expected Olc state at the
		// moment the menu renders.
		if d == nil || d.Olc == nil {
			t.Error("expected Olc allocated when disp fires")
			return
		}
		if d.Olc.Mode != types.OEDIT_MAIN_MENU {
			t.Errorf("Mode = %d, want OEDIT_MAIN_MENU", d.Olc.Mode)
		}
		if d.Olc.Vnum != 4000 {
			t.Errorf("Vnum = %d, want 4000", d.Olc.Vnum)
		}
		idx, ok := d.Olc.Target.(*types.ObjIndexData)
		if !ok || idx == nil || idx.Vnum != 4000 {
			t.Errorf("Target not *ObjIndexData{Vnum:4000}; got %T %v", d.Olc.Target, d.Olc.Target)
		}
	}
	defer func() { OeditDispMenuFunc = nil }()

	DoOedit(ch, "4000")
	_ = readOutput(ch, client)

	if !dispCalled {
		t.Error("OeditDispMenuFunc was not invoked on vnum-only entry")
	}
	if ch.Desc.Connected != int(types.CON_OEDIT) {
		t.Errorf("Connected = %d, want CON_OEDIT (%d)", ch.Desc.Connected, types.CON_OEDIT)
	}
	if ch.Desc.Olc == nil || ch.Desc.Olc.Mode != types.OEDIT_MAIN_MENU {
		t.Errorf("post-entry Olc state wrong: %+v", ch.Desc.Olc)
	}
}

// TestDoOedit_UnknownVnumRefuses pins the no-auto-create contract: the
// vnum must exist; menu entry does NOT allocate a new prototype (C
// `do_ooedit` falls through to "OEdit what?" when ENABLE_OLC2_EXTRAS is
// off, matching Go's "does not exist" message).
func TestDoOedit_UnknownVnumRefuses(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	OeditDispMenuFunc = func(d *types.DescriptorData) {
		t.Error("OeditDispMenuFunc must NOT fire for unknown vnum")
	}
	defer func() { OeditDispMenuFunc = nil }()

	DoOedit(ch, "99999")
	out := readOutput(ch, client)
	if !strings.Contains(out, "does not exist") {
		t.Errorf("expected does-not-exist, got: %q", out)
	}
	if ch.Desc.Connected == int(types.CON_OEDIT) {
		t.Error("unknown vnum must not transition to CON_OEDIT")
	}
	if ch.Desc.Olc != nil {
		t.Error("unknown vnum must not allocate Olc")
	}
}

// TestDoOedit_NonImmortalRefused pins A6 trust gate: mortals see "Huh?"
// and do NOT enter the menu regardless of arg shape.
func TestDoOedit_NonImmortalRefused(t *testing.T) {
	w := setupOlcWorld()
	w.ObjIndex[4000] = &types.ObjIndexData{Vnum: 4000, Name: "x"}
	mortal, client := makeTestChar("Mortal")
	defer client.Close()
	mortal.Level = 1

	OeditDispMenuFunc = func(d *types.DescriptorData) {
		t.Error("OeditDispMenuFunc must NOT fire for mortal")
	}
	defer func() { OeditDispMenuFunc = nil }()

	DoOedit(mortal, "4000")
	out := readOutput(mortal, client)
	if !strings.Contains(out, "Huh?") {
		t.Errorf("expected Huh?, got %q", out)
	}
	if mortal.Desc.Connected == int(types.CON_OEDIT) {
		t.Error("mortal must not transition to CON_OEDIT")
	}
}

// TestDoOedit_NoDescriptor pins NPC defense: a CharData without a
// descriptor cannot enter the menu. (Trust-gate still blocks NPCs in
// practice because NPC Trust defaults to 0, but the defensive nil-check
// is independently verified.)
func TestDoOedit_NoDescriptor(t *testing.T) {
	w := setupOlcWorld()
	w.ObjIndex[4000] = &types.ObjIndexData{Vnum: 4000, Name: "x"}
	// Build a bare immortal with no descriptor.
	ch := &types.CharData{
		Name:  "Deskless",
		Level: types.LEVEL_IMMORTAL,
	}

	OeditDispMenuFunc = func(d *types.DescriptorData) {
		t.Error("OeditDispMenuFunc must NOT fire when Desc == nil")
	}
	defer func() { OeditDispMenuFunc = nil }()

	// Must not panic.
	DoOedit(ch, "4000")
}

// TestDoOedit_FlatPathStillWorks pins A5: `oedit <vnum> <sub> [args]`
// takes the flat subcommand branch, does NOT enter the menu, and
// Connected stays at CON_PLAYING.
func TestDoOedit_FlatPathStillWorks(t *testing.T) {
	w := setupOlcWorld()
	w.ObjIndex[4000] = &types.ObjIndexData{Vnum: 4000, Name: "old"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	OeditDispMenuFunc = func(d *types.DescriptorData) {
		t.Error("OeditDispMenuFunc must NOT fire on flat-subcommand path")
	}
	defer func() { OeditDispMenuFunc = nil }()

	DoOedit(ch, "4000 name newname")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Name set") {
		t.Errorf("expected flat-path ack, got %q", out)
	}
	if w.ObjIndex[4000].Name != "newname" {
		t.Errorf("flat name mutation did not land; got %q", w.ObjIndex[4000].Name)
	}
	if ch.Desc.Connected == int(types.CON_OEDIT) {
		t.Error("flat path must NOT transition to CON_OEDIT")
	}
	if ch.Desc.Olc != nil {
		t.Error("flat path must NOT allocate Olc")
	}
}

// TestDoOedit_ShowSubcommand preserves the old flat-summary behavior
// under the new explicit `show` subcommand (Q2 rebinding — see plan
// §Open Questions Q2).
func TestDoOedit_ShowSubcommand(t *testing.T) {
	w := setupOlcWorld()
	w.ObjIndex[4000] = &types.ObjIndexData{Vnum: 4000, Name: "item", ShortDescr: "a shiny thing", ItemType: types.ITEM_WEAPON, Level: 5, Weight: 3}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	OeditDispMenuFunc = func(d *types.DescriptorData) {
		t.Error("show subcommand must NOT enter the menu")
	}
	defer func() { OeditDispMenuFunc = nil }()

	DoOedit(ch, "4000 show")
	out := readOutput(ch, client)
	if !strings.Contains(out, "a shiny thing") {
		t.Errorf("show summary missing short desc; got %q", out)
	}
	if ch.Desc.Connected == int(types.CON_OEDIT) {
		t.Error("show subcommand must NOT transition to CON_OEDIT")
	}
}

func TestDoOedit_AffectsAddDel(t *testing.T) {
	w := setupOlcWorld()
	w.ObjIndex[4030] = &types.ObjIndexData{Vnum: 4030, Name: "ring"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoOedit(ch, "4030 affects add 1 5")
	_ = readOutput(ch, client)
	if len(w.ObjIndex[4030].Affects) != 1 {
		t.Fatalf("expected 1 affect, got %d", len(w.ObjIndex[4030].Affects))
	}
	DoOedit(ch, "4030 affects del 0")
	_ = readOutput(ch, client)
	if len(w.ObjIndex[4030].Affects) != 0 {
		t.Errorf("expected 0 affects, got %d", len(w.ObjIndex[4030].Affects))
	}
}

// --- DoMedit ---

func TestDoMedit_NoArg(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	DoMedit(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Usage") {
		t.Errorf("expected usage, got: %q", out)
	}
}

func TestDoMedit_Create(t *testing.T) {
	w := setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	DoMedit(ch, "5000 create a goblin")
	out := readOutput(ch, client)
	if !strings.Contains(out, "created") {
		t.Errorf("expected created, got: %q", out)
	}
	if _, ok := w.MobIndex[5000]; !ok {
		t.Error("mob 5000 should exist")
	}
}

func TestDoMedit_SetFields(t *testing.T) {
	w := setupOlcWorld()
	w.MobIndex[5100] = &types.MobIndexData{Vnum: 5100, PlayerName: "rat", ShortDescr: "a rat"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	tests := []struct {
		name  string
		args  string
		check func(*types.MobIndexData) bool
	}{
		{"name", "5100 name newrat", func(m *types.MobIndexData) bool { return m.PlayerName == "newrat" }},
		{"short", "5100 short a plump rat", func(m *types.MobIndexData) bool { return m.ShortDescr == "a plump rat" }},
		{"level", "5100 level 25", func(m *types.MobIndexData) bool { return m.Level == 25 }},
		{"armor", "5100 armor -50", func(m *types.MobIndexData) bool { return m.AC == -50 }},
		{"hitroll", "5100 hitroll 7", func(m *types.MobIndexData) bool { return m.Hitroll == 7 }},
		{"damroll", "5100 damroll 4", func(m *types.MobIndexData) bool { return m.Damroll == 4 }},
		{"align", "5100 align -800", func(m *types.MobIndexData) bool { return m.Alignment == -800 }},
		{"race", "5100 race 3", func(m *types.MobIndexData) bool { return m.Race == 3 }},
		{"class", "5100 class 2", func(m *types.MobIndexData) bool { return m.Class == 2 }},
		{"sex_male", "5100 sex male", func(m *types.MobIndexData) bool { return m.Sex == types.SEX_MALE }},
		{"stats_str", "5100 stats str 18", func(m *types.MobIndexData) bool { return m.PermStr == 18 }},
		{"hp", "5100 hp 5 10 3", func(m *types.MobIndexData) bool { return m.HitNoDice == 5 && m.HitSizeDice == 10 && m.HitPlus == 3 }},
		{"gold", "5100 gold 250", func(m *types.MobIndexData) bool { return m.Gold == 250 }},
		{"xp", "5100 xp 1500", func(m *types.MobIndexData) bool { return m.Exp == 1500 }},
		{"position", "5100 position 4", func(m *types.MobIndexData) bool { return m.Position == 4 && m.DefPosition == 4 }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			DoMedit(ch, tc.args)
			_ = readOutput(ch, client)
			if !tc.check(w.MobIndex[5100]) {
				t.Errorf("field check failed for %s", tc.name)
			}
		})
	}
}

func TestDoMedit_UnknownVnum(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	DoMedit(ch, "88888 name foo")
	out := readOutput(ch, client)
	if !strings.Contains(out, "does not exist") {
		t.Errorf("expected does-not-exist, got: %q", out)
	}
}

func TestDoMedit_BadStat(t *testing.T) {
	w := setupOlcWorld()
	w.MobIndex[5200] = &types.MobIndexData{Vnum: 5200, PlayerName: "x"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	DoMedit(ch, "5200 stats bogus 10")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Unknown stat") {
		t.Errorf("expected unknown-stat, got: %q", out)
	}
}

// --- G14: DoMedit menu-entry (CON_MEDIT) extension ---
//
// Plan-phase6-olc-medit.md §G14 / §A30. These tests pin the same split
// shape oedit Wave 4 landed: empty arg → usage; vnum-only → enter
// CON_MEDIT and fire MeditDispMenuFunc; flat sub-arg path unchanged;
// `show` subcommand re-binds the pre-Wave-5 flat summary; the
// double-edit guard refuses a second concurrent menu-entry.
//
// PC-by-name resolution is a documented scope cut — test
// TestDoMedit_PcNameRefusedPendingLookupSeam pins the failure-mode
// message so future workers know the gap.

// TestDoMedit_NoArgEntersMenuNpc pins A30 — `medit <vnum>` with no
// subcommand wraps the prototype, sets ACT_PROTOTYPE on the wrapper,
// flips Connected to CON_MEDIT, allocates Olc with Mode=NPC main menu
// and Target=*CharData (NOT *MobIndexData — the dispatcher type-asserts
// to *CharData per Wave-1 design), and fires the menu seam.
func TestDoMedit_NoArgEntersMenuNpc(t *testing.T) {
	w := setupOlcWorld()
	w.MobIndex[6000] = &types.MobIndexData{
		Vnum:        6000,
		PlayerName:  "rat",
		ShortDescr:  "a small rat",
		LongDescr:   "A small rat scurries here.",
		Level:       1,
		Position:    types.POS_STANDING,
		DefPosition: types.POS_STANDING,
	}
	w.MobIndex[6000].Act.Set(types.ACT_IS_NPC)

	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	var dispCalled bool
	MeditDispMenuFunc = func(d *types.DescriptorData) {
		dispCalled = true
		if d == nil || d.Olc == nil {
			t.Error("expected Olc allocated when disp fires")
			return
		}
		if d.Olc.Mode != types.MEDIT_NPC_MAIN_MENU {
			t.Errorf("Mode = %d, want MEDIT_NPC_MAIN_MENU", d.Olc.Mode)
		}
		if d.Olc.Vnum != 6000 {
			t.Errorf("Vnum = %d, want 6000", d.Olc.Vnum)
		}
		victim, ok := d.Olc.Target.(*types.CharData)
		if !ok || victim == nil {
			t.Errorf("Target not *CharData; got %T %v", d.Olc.Target, d.Olc.Target)
			return
		}
		if victim.IndexData == nil || victim.IndexData.Vnum != 6000 {
			t.Errorf("victim.IndexData not wired to mob 6000; got %+v", victim.IndexData)
		}
		if !victim.IsNPC() {
			t.Error("wrapped victim should report IsNPC()")
		}
		if !victim.Act.IsSet(types.ACT_PROTOTYPE) {
			t.Error("ACT_PROTOTYPE must be set on the wrapper so arms mirror to IndexData")
		}
	}
	defer func() { MeditDispMenuFunc = nil }()

	DoMedit(ch, "6000")
	_ = readOutput(ch, client)

	if !dispCalled {
		t.Error("MeditDispMenuFunc was not invoked on vnum-only entry")
	}
	if ch.Desc.Connected != int(types.CON_MEDIT) {
		t.Errorf("Connected = %d, want CON_MEDIT (%d)", ch.Desc.Connected, types.CON_MEDIT)
	}
	if ch.Desc.Olc == nil || ch.Desc.Olc.Mode != types.MEDIT_NPC_MAIN_MENU {
		t.Errorf("post-entry Olc state wrong: %+v", ch.Desc.Olc)
	}
}

// TestDoMedit_PcNameRefusedPendingLookupSeam pins the documented A31
// scope cut: PC-by-name (`medit Tagith`) requires a worldPcLookup seam
// that has not landed. Today non-numeric arguments emit a clear
// message naming the gap so future workers know to wire it.
func TestDoMedit_PcNameRefusedPendingLookupSeam(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	MeditDispMenuFunc = func(d *types.DescriptorData) {
		t.Error("MeditDispMenuFunc must NOT fire on non-numeric arg")
	}
	defer func() { MeditDispMenuFunc = nil }()

	DoMedit(ch, "Tagith")
	out := readOutput(ch, client)
	if !strings.Contains(out, "PC editing by name not yet supported") {
		t.Errorf("expected PC-name-not-supported message, got: %q", out)
	}
	if ch.Desc.Connected == int(types.CON_MEDIT) {
		t.Error("non-numeric arg must not transition to CON_MEDIT")
	}
	if ch.Desc.Olc != nil {
		t.Error("non-numeric arg must not allocate Olc")
	}
}

// TestDoMedit_FlatPathStillWorks pins regression: `medit <vnum> <sub>
// [args]` takes the existing flat subcommand branch unchanged and does
// NOT enter the menu.
func TestDoMedit_FlatPathStillWorks(t *testing.T) {
	w := setupOlcWorld()
	w.MobIndex[6010] = &types.MobIndexData{Vnum: 6010, PlayerName: "old"}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	MeditDispMenuFunc = func(d *types.DescriptorData) {
		t.Error("MeditDispMenuFunc must NOT fire on flat-subcommand path")
	}
	defer func() { MeditDispMenuFunc = nil }()

	DoMedit(ch, "6010 name foo")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Name set to: foo") {
		t.Errorf("expected flat-path ack, got %q", out)
	}
	if w.MobIndex[6010].PlayerName != "foo" {
		t.Errorf("flat name mutation did not land; got %q", w.MobIndex[6010].PlayerName)
	}
	if ch.Desc.Connected == int(types.CON_MEDIT) {
		t.Error("flat path must NOT transition to CON_MEDIT")
	}
	if ch.Desc.Olc != nil {
		t.Error("flat path must NOT allocate Olc")
	}
}

// TestDoMedit_DoubleEditGuard pins M16 — a second `medit <vnum>` while
// the same descriptor is already in CON_MEDIT must refuse rather than
// clobber the in-flight Olc state.
func TestDoMedit_DoubleEditGuard(t *testing.T) {
	w := setupOlcWorld()
	w.MobIndex[6020] = &types.MobIndexData{Vnum: 6020, PlayerName: "first"}
	w.MobIndex[6020].Act.Set(types.ACT_IS_NPC)
	w.MobIndex[6021] = &types.MobIndexData{Vnum: 6021, PlayerName: "second"}
	w.MobIndex[6021].Act.Set(types.ACT_IS_NPC)

	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	dispCalls := 0
	MeditDispMenuFunc = func(d *types.DescriptorData) { dispCalls++ }
	defer func() { MeditDispMenuFunc = nil }()

	DoMedit(ch, "6020")
	_ = readOutput(ch, client)
	if ch.Desc.Connected != int(types.CON_MEDIT) {
		t.Fatalf("first medit did not enter CON_MEDIT; Connected=%d", ch.Desc.Connected)
	}
	firstOlc := ch.Desc.Olc

	DoMedit(ch, "6021")
	out := readOutput(ch, client)
	if !strings.Contains(out, "already editing") {
		t.Errorf("expected double-edit refusal, got: %q", out)
	}
	if dispCalls != 1 {
		t.Errorf("MeditDispMenuFunc called %d times, want 1 (no second entry)", dispCalls)
	}
	if ch.Desc.Olc != firstOlc {
		t.Error("second medit clobbered Olc — guard failed")
	}
}

// TestDoMedit_NonImmortalRefused pins the trust gate — mortals see
// "Huh?" and do NOT enter the menu.
func TestDoMedit_NonImmortalRefused(t *testing.T) {
	w := setupOlcWorld()
	w.MobIndex[6030] = &types.MobIndexData{Vnum: 6030, PlayerName: "x"}
	mortal, client := makeTestChar("Mortal")
	defer client.Close()
	mortal.Level = 1

	MeditDispMenuFunc = func(d *types.DescriptorData) {
		t.Error("MeditDispMenuFunc must NOT fire for mortal")
	}
	defer func() { MeditDispMenuFunc = nil }()

	DoMedit(mortal, "6030")
	out := readOutput(mortal, client)
	if !strings.Contains(out, "Huh?") {
		t.Errorf("expected Huh?, got %q", out)
	}
	if mortal.Desc.Connected == int(types.CON_MEDIT) {
		t.Error("mortal must not transition to CON_MEDIT")
	}
}

// TestDoMedit_ShowSubcommand preserves the old flat-summary behavior
// under the new explicit `show` subcommand (mirrors oedit ShowSubcommand).
func TestDoMedit_ShowSubcommand(t *testing.T) {
	w := setupOlcWorld()
	w.MobIndex[6040] = &types.MobIndexData{
		Vnum: 6040, PlayerName: "rat", ShortDescr: "a tiny rat", Level: 3,
	}
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	MeditDispMenuFunc = func(d *types.DescriptorData) {
		t.Error("show subcommand must NOT enter the menu")
	}
	defer func() { MeditDispMenuFunc = nil }()

	DoMedit(ch, "6040 show")
	out := readOutput(ch, client)
	if !strings.Contains(out, "a tiny rat") {
		t.Errorf("show summary missing short desc; got %q", out)
	}
	if ch.Desc.Connected == int(types.CON_MEDIT) {
		t.Error("show subcommand must NOT transition to CON_MEDIT")
	}
}

// --- Wave 5 follow-up: HIGH #1 — bare-allocation wrapper ---
//
// The pre-fix code wrapped the prototype via handler.CreateMobile, which
// bumps idx.Count and appends to WorldRef.Characters. cleanupOlc only
// resets descriptor state, so each `medit <vnum>` permanently leaked one
// phantom mob: idx.Count++ would block area resets that gate on
// `idx.Count >= reset.Arg2`, and the orphaned CharData entries (with
// InRoom == nil) accumulated in WorldRef.Characters.
//
// The fix replaces handler.CreateMobile with a bare *types.CharData
// allocation that copies only the fields the menu renderers (medit_menu.go)
// read, plus IndexData and ACT_PROTOTYPE so the existing arms' dual-write
// path (medit_arms.go) still propagates edits to the prototype. The
// invariants are: idx.Count must NOT change, and WorldRef.Characters must
// NOT grow when DoMedit is invoked.
//
// Tests below mutation-verify the fix at the world-state level.

// TestDoMedit_NoArgDoesNotIncrementCount pins HIGH #1 invariant 1: the
// menu-entry path must not bump idx.Count. With the buggy CreateMobile
// call, this test fails because Count goes 0→1.
func TestDoMedit_NoArgDoesNotIncrementCount(t *testing.T) {
	w := setupOlcWorld()
	w.MobIndex[6000] = &types.MobIndexData{
		Vnum:        6000,
		PlayerName:  "rat",
		ShortDescr:  "a tiny rat",
		LongDescr:   "A tiny rat scurries here.",
		Level:       1,
		Position:    types.POS_STANDING,
		DefPosition: types.POS_STANDING,
	}
	w.MobIndex[6000].Act.Set(types.ACT_IS_NPC)

	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	MeditDispMenuFunc = func(d *types.DescriptorData) {}
	defer func() { MeditDispMenuFunc = nil }()

	if got := w.MobIndex[6000].Count; got != 0 {
		t.Fatalf("precondition: Count=%d, want 0", got)
	}

	DoMedit(ch, "6000")
	_ = readOutput(ch, client)

	if got := w.MobIndex[6000].Count; got != 0 {
		t.Errorf("post-DoMedit idx.Count=%d, want 0 — wrapper must not call CreateMobile", got)
	}
}

// TestDoMedit_NoArgDoesNotPolluteWorldCharacters pins HIGH #1 invariant 2:
// the menu-entry path must not append to WorldRef.Characters. Phantom
// entries with InRoom==nil are iterated by every pulse update.
func TestDoMedit_NoArgDoesNotPolluteWorldCharacters(t *testing.T) {
	w := setupOlcWorld()
	w.MobIndex[6000] = &types.MobIndexData{
		Vnum:        6000,
		PlayerName:  "rat",
		Level:       1,
		Position:    types.POS_STANDING,
		DefPosition: types.POS_STANDING,
	}
	w.MobIndex[6000].Act.Set(types.ACT_IS_NPC)

	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	MeditDispMenuFunc = func(d *types.DescriptorData) {}
	defer func() { MeditDispMenuFunc = nil }()

	before := len(w.Characters)
	DoMedit(ch, "6000")
	_ = readOutput(ch, client)
	after := len(w.Characters)

	if after != before {
		t.Errorf("len(WorldRef.Characters): before=%d after=%d (delta=%d) — wrapper must not call AddChar",
			before, after, after-before)
	}
}

// TestDoMedit_QuitDoesNotLeakAfterMultipleEntries pins HIGH #1 invariant 3:
// repeating the medit-then-quit cycle must leave both idx.Count and
// WorldRef.Characters at their starting values. cleanupOlc only resets
// descriptor state — without the bare-allocation fix, each cycle leaks a
// phantom mob even after the builder Quits. After 5 iterations the buggy
// path would leave Count=5 and 5 extra entries in Characters.
func TestDoMedit_QuitDoesNotLeakAfterMultipleEntries(t *testing.T) {
	w := setupOlcWorld()
	w.MobIndex[6000] = &types.MobIndexData{
		Vnum:        6000,
		PlayerName:  "rat",
		Level:       1,
		Position:    types.POS_STANDING,
		DefPosition: types.POS_STANDING,
	}
	w.MobIndex[6000].Act.Set(types.ACT_IS_NPC)

	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	MeditDispMenuFunc = func(d *types.DescriptorData) {}
	defer func() { MeditDispMenuFunc = nil }()

	beforeChars := len(w.Characters)

	for i := 0; i < 5; i++ {
		DoMedit(ch, "6000")
		_ = readOutput(ch, client)
		// Simulate the Q-arm cleanup that medit_parse.cleanupOlc performs
		// on a successful Quit: nil out Olc, return descriptor to
		// CON_PLAYING. We don't import game/ here (would be a cycle), so
		// we replicate the two field writes inline. The test pins that
		// these are the only post-Quit cleanup steps and that the wrapper
		// allocation must therefore not have side-effects on world state.
		ch.Desc.Olc = nil
		ch.Desc.Connected = int(types.CON_PLAYING)
	}

	if got := w.MobIndex[6000].Count; got != 0 {
		t.Errorf("after 5 medit/Q cycles idx.Count=%d, want 0", got)
	}
	if got := len(w.Characters); got != beforeChars {
		t.Errorf("after 5 medit/Q cycles len(WorldRef.Characters)=%d, want %d (leaked %d phantoms)",
			got, beforeChars, got-beforeChars)
	}
}

// TestDoMedit_ZeroVnumRejected pins LOW #4: a numeric-but-non-positive
// argument ("0", "-5") must take the "Vnum must be a positive number."
// branch, NOT the "PC editing by name" branch. The pre-fix code's
// redundant inner Atoi conflated these — the outer Atoi already
// succeeded (so err==nil) but vnum<=0; the inner re-Atoi also succeeded;
// the dispatch was correct only by accident. The split-branch fix makes
// the routing explicit on err.
func TestDoMedit_ZeroVnumRejected(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	for _, arg := range []string{"0", "-5"} {
		DoMedit(ch, arg)
		out := readOutput(ch, client)
		if !strings.Contains(out, "Vnum must be a positive number") {
			t.Errorf("DoMedit(%q): expected positive-number rejection, got %q", arg, out)
		}
		if strings.Contains(out, "PC editing by name") {
			t.Errorf("DoMedit(%q): wrongly took PC-name branch; got %q", arg, out)
		}
	}
}

// TestDoMedit_UnknownVnumRefuses pins the no-auto-create contract for
// the menu-entry path: vnum must exist; menu does NOT fire.
func TestDoMedit_UnknownVnumRefuses(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()

	MeditDispMenuFunc = func(d *types.DescriptorData) {
		t.Error("MeditDispMenuFunc must NOT fire for unknown vnum")
	}
	defer func() { MeditDispMenuFunc = nil }()

	DoMedit(ch, "9999999")
	out := readOutput(ch, client)
	if !strings.Contains(out, "does not exist") {
		t.Errorf("expected does-not-exist, got: %q", out)
	}
	if ch.Desc.Connected == int(types.CON_MEDIT) {
		t.Error("unknown vnum must not transition to CON_MEDIT")
	}
	if ch.Desc.Olc != nil {
		t.Error("unknown vnum must not allocate Olc")
	}
}

func TestOLC_EditMortalReject(t *testing.T) {
	_ = setupOlcWorld()
	mortal, client := makeTestChar("Mortal")
	defer client.Close()
	mortal.Level = 1

	DoOedit(mortal, "1 create foo")
	if out := readOutput(mortal, client); !strings.Contains(out, "Huh?") {
		t.Errorf("oedit: expected Huh?, got: %q", out)
	}
	DoMedit(mortal, "1 create foo")
	if out := readOutput(mortal, client); !strings.Contains(out, "Huh?") {
		t.Errorf("medit: expected Huh?, got: %q", out)
	}
}
