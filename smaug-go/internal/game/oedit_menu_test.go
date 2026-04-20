package game

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// --- G4 menu-renderer tests ---

func TestOeditDispMenu_ContainsAllFields(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.Name = "widget"
	rig.idx.ShortDescr = "a sparkly widget"
	rig.idx.Description = "A widget lies here."
	rig.idx.ActionDesc = "* click *"
	rig.idx.ItemType = types.ITEM_WEAPON
	rig.idx.Weight = 5
	rig.idx.GoldCost = 100
	rig.idx.Rent = 10
	rig.idx.Level = 20
	rig.idx.Layers = 3
	rig.idx.Value[0] = 11
	rig.idx.Value[5] = 66

	OeditDispMenu(rig.d)
	out := rig.readBuf(t)

	// Color tags embed between digit/letter and ')', so substring matches
	// use the ') <label>' form rather than '1) Name'.
	mustContain := []string{
		"Object number",
		") Name",
		"widget",
		") Short desc",
		"a sparkly widget",
		") Long desc",
		"A widget lies here.",
		") Action desc",
		"* click *",
		") Type",
		"weapon",
		") Extra flags",
		") Wear flags",
		") Weight",
		") Cost",
		") Rent",
		") Timer",
		") Level",
		") Layers",
		") Values",
		") Affect menu",
		") Extra descriptions menu",
		") Quit",
		"Enter choice",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("main menu missing %q; got: %q", s, out)
		}
	}
	if rig.d.Olc.Mode != types.OEDIT_MAIN_MENU {
		t.Errorf("Mode = %d, want OEDIT_MAIN_MENU", rig.d.Olc.Mode)
	}
}

func TestOeditDispMenu_DoesNotRenderOptionH(t *testing.T) {
	// Audit-corrected plan §G4: H (mprog) is parser-dispatched but NOT
	// rendered in the main-menu display.
	rig := newOeditHarness(t)
	OeditDispMenu(rig.d)
	out := rig.readBuf(t)
	// Because color tags separate H from ')', match on "gH&w)" which
	// is exactly how every other option is rendered.
	if strings.Contains(out, "gH&w)") {
		t.Errorf("main menu must not render 'H)' option; got: %q", out)
	}
}

func TestOeditDispTypeMenu_ShowsAllItemTypes(t *testing.T) {
	rig := newOeditHarness(t)
	oeditDispTypeMenu(rig.d)
	out := rig.readBuf(t)
	// Spot-check a handful of types at different indices.
	for _, name := range []string{"light", "weapon", "armor", "pill", "lever", "salve"} {
		if !strings.Contains(out, name) {
			t.Errorf("type menu missing %q; got: %q", name, out)
		}
	}
	if !strings.Contains(out, "Enter type") {
		t.Errorf("type menu missing 'Enter type' prompt; got: %q", out)
	}
	if rig.d.Olc.Mode != types.OEDIT_TYPE {
		t.Errorf("Mode = %d, want OEDIT_TYPE", rig.d.Olc.Mode)
	}
}

func TestOeditDispExtraMenu_ShowsFlagsAndCurrent(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.ExtraFlags.Set(types.ITEM_MAGIC)
	oeditDispExtraMenu(rig.d)
	out := rig.readBuf(t)
	for _, name := range []string{"glow", "hum", "magic", "prototype"} {
		if !strings.Contains(out, name) {
			t.Errorf("extra menu missing %q; got: %q", name, out)
		}
	}
	if !strings.Contains(out, "Extra flags:") {
		t.Errorf("extra menu missing 'Extra flags:' summary; got: %q", out)
	}
	if !strings.Contains(out, "Enter flags, 0 to quit") {
		t.Errorf("extra menu missing prompt; got: %q", out)
	}
	if rig.d.Olc.Mode != types.OEDIT_EXTRAS {
		t.Errorf("Mode = %d, want OEDIT_EXTRAS", rig.d.Olc.Mode)
	}
}

func TestOeditDispWearMenu_SkipsDualWield(t *testing.T) {
	// ITEM_DUAL_WIELD (bit 15) must NOT appear in the menu. The label
	// "dual_wield" is only in wFlagNames for save fidelity.
	rig := newOeditHarness(t)
	oeditDispWearMenu(rig.d)
	out := rig.readBuf(t)
	if strings.Contains(out, "dual_wield") {
		t.Errorf("wear menu must skip dual_wield; got: %q", out)
	}
	// Other wear flags must be rendered.
	for _, name := range []string{"take", "finger", "body", "head", "wield"} {
		if !strings.Contains(out, name) {
			t.Errorf("wear menu missing %q; got: %q", name, out)
		}
	}
	if rig.d.Olc.Mode != types.OEDIT_WEAR {
		t.Errorf("Mode = %d, want OEDIT_WEAR", rig.d.Olc.Mode)
	}
}

func TestOeditDispLayerMenu_ShowsNineLabels(t *testing.T) {
	rig := newOeditHarness(t)
	oeditDispLayerMenu(rig.d)
	out := rig.readBuf(t)
	for _, name := range []string{
		"Nothing", "Silk Shirt", "Leather Vest", "Light Chainmail",
		"Leather Jacket", "Light Cloak", "Loose Cloak", "Cape",
		"Magical Effects",
	} {
		if !strings.Contains(out, name) {
			t.Errorf("layer menu missing %q; got: %q", name, out)
		}
	}
	if rig.d.Olc.Mode != types.OEDIT_LAYERS {
		t.Errorf("Mode = %d, want OEDIT_LAYERS", rig.d.Olc.Mode)
	}
}

func TestOeditDispExtradescMenu_ShowsListAndAR(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.ExtraDescr = []*types.ExtraDescrData{
		{Keyword: "blade", Description: "sharp"},
		{Keyword: "hilt", Description: "worn"},
	}
	oeditDispExtradescMenu(rig.d)
	out := rig.readBuf(t)
	for _, s := range []string{"blade", "hilt", ") Add a new description", ") Remove a description", ") Quit", "Enter choice"} {
		if !strings.Contains(out, s) {
			t.Errorf("extradesc menu missing %q; got: %q", s, out)
		}
	}
	if rig.d.Olc.Mode != types.OEDIT_EXTRADESC_MENU {
		t.Errorf("Mode = %d, want OEDIT_EXTRADESC_MENU", rig.d.Olc.Mode)
	}
}

// --- Wave-2 value-menu stubs ---

func TestOeditDispVal1Menu_Stub_SetsValue1Mode(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_WEAPON
	oeditDispVal1Menu(rig.d)
	out := rig.readBuf(t)
	if !strings.Contains(out, "Value 1 editing") {
		t.Errorf("expected stub text; got: %q", out)
	}
	if !strings.Contains(out, "Wave 3") {
		t.Errorf("expected 'Wave 3' marker; got: %q", out)
	}
	if rig.d.Olc.Mode != types.OEDIT_VALUE_1 {
		t.Errorf("Mode = %d, want OEDIT_VALUE_1", rig.d.Olc.Mode)
	}
}

func TestGetOtype_WordLookup(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"weapon", types.ITEM_WEAPON},
		{"Weapon", types.ITEM_WEAPON},
		{"  light  ", types.ITEM_LIGHT},
		{"", -1},
		{"notathing", -1},
	}
	for _, tc := range cases {
		if got := getOtype(tc.in); got != tc.want {
			t.Errorf("getOtype(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestGetOflag_WordLookup(t *testing.T) {
	if getOflag("glow") != 0 {
		t.Errorf("getOflag(glow) = %d, want 0", getOflag("glow"))
	}
	if getOflag("prototype") != types.ITEM_PROTOTYPE {
		t.Errorf("getOflag(prototype) = %d, want ITEM_PROTOTYPE=%d", getOflag("prototype"), types.ITEM_PROTOTYPE)
	}
	if getOflag("notathing") != -1 {
		t.Errorf("getOflag(notathing) = %d, want -1", getOflag("notathing"))
	}
}

func TestGetWflag_RejectsDualWield(t *testing.T) {
	if got := getWflag("dual_wield"); got != -1 {
		t.Errorf("getWflag(dual_wield) = %d, want -1 (OLC may not set this bit)", got)
	}
	if got := getWflag("body"); got != 3 {
		t.Errorf("getWflag(body) = %d, want 3", got)
	}
}
