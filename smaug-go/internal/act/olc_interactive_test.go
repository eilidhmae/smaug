package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// --- DoOedit ---

func TestDoOedit_NoArg(t *testing.T) {
	_ = setupOlcWorld()
	ch, client := makeImmTestChar("Builder")
	defer client.Close()
	DoOedit(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Usage") {
		t.Errorf("expected usage, got: %q", out)
	}
}

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
