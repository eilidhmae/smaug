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
