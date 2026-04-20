package game

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// -----------------------------------------------------------------------------
// PRE-G — D-branch layerable precheck (Wave 2 scope miss fix)
// -----------------------------------------------------------------------------

// TestOeditParse_MainMenu_D_RefusesNonLayerable pins the C-parity
// precheck at ooedit.c:1261-1275. When none of BODY/ABOUT/ARMS/FEET/
// HANDS/LEGS/WAIST wear bits are set, D refuses with a specific message
// and does NOT transition into OEDIT_LAYERS.
func TestOeditParse_MainMenu_D_RefusesNonLayerable(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.WearFlags = (1 << 0) | (1 << 1) // TAKE + FINGER — not layerable
	oeditParse(rig.d, "D")
	if rig.d.Olc.Mode != types.OEDIT_MAIN_MENU {
		t.Errorf("Mode = %d, want OEDIT_MAIN_MENU (refused)", rig.d.Olc.Mode)
	}
	out := rig.readBuf(t)
	if !strings.Contains(out, "not layerable") {
		t.Errorf("expected 'not layerable' refusal; got: %q", out)
	}
}

// TestOeditParse_MainMenu_D_OpensForLayerableObject pins the
// positive-path: BODY set → OEDIT_LAYERS entered.
func TestOeditParse_MainMenu_D_OpensForLayerableObject(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.WearFlags = 1 << 3 // ITEM_WEAR_BODY
	oeditParse(rig.d, "D")
	if rig.d.Olc.Mode != types.OEDIT_LAYERS {
		t.Errorf("Mode = %d, want OEDIT_LAYERS", rig.d.Olc.Mode)
	}
}

// TestOeditParse_MainMenu_D_LayerableByWaist pins that WAIST alone is
// layerable (per C :1268).
func TestOeditParse_MainMenu_D_LayerableByWaist(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.WearFlags = 1 << 11 // ITEM_WEAR_WAIST
	oeditParse(rig.d, "D")
	if rig.d.Olc.Mode != types.OEDIT_LAYERS {
		t.Errorf("Mode = %d, want OEDIT_LAYERS with WAIST", rig.d.Olc.Mode)
	}
}

// -----------------------------------------------------------------------------
// G7 — Per-item-type value-menu disp tests
// -----------------------------------------------------------------------------

func TestOeditDispVal1_WeaponPromptsCondition(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_WEAPON
	oeditDispVal1Menu(rig.d)
	out := rig.readBuf(t)
	if !strings.Contains(out, "Condition") {
		t.Errorf("expected 'Condition' prompt for weapon; got: %q", out)
	}
	if rig.d.Olc.Mode != types.OEDIT_VALUE_1 {
		t.Errorf("Mode = %d, want OEDIT_VALUE_1", rig.d.Olc.Mode)
	}
}

func TestOeditDispVal1_LightSkipsToVal3(t *testing.T) {
	// C :630-633: ITEM_LIGHT jumps straight to val3 (skipping val1+val2).
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_LIGHT
	oeditDispVal1Menu(rig.d)
	if rig.d.Olc.Mode != types.OEDIT_VALUE_3 {
		t.Errorf("Mode = %d, want OEDIT_VALUE_3 (ITEM_LIGHT skip)", rig.d.Olc.Mode)
	}
}

func TestOeditDispVal1_HerbSkipsToVal2(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_HERB
	oeditDispVal1Menu(rig.d)
	if rig.d.Olc.Mode != types.OEDIT_VALUE_2 {
		t.Errorf("Mode = %d, want OEDIT_VALUE_2 (ITEM_HERB skip)", rig.d.Olc.Mode)
	}
}

func TestOeditDispVal1_LeverCallsLeverFlagsMenu(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_LEVER
	oeditDispVal1Menu(rig.d)
	out := rig.readBuf(t)
	if !strings.Contains(out, "Lever flags") {
		t.Errorf("expected lever-flags menu; got: %q", out)
	}
}

func TestOeditDispVal1_KeyFallsBackToMain(t *testing.T) {
	// ITEM_KEY has no val1 prompt in C — default case bails to main.
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_KEY
	oeditDispVal1Menu(rig.d)
	if rig.d.Olc.Mode != types.OEDIT_MAIN_MENU {
		t.Errorf("Mode = %d, want OEDIT_MAIN_MENU (ITEM_KEY default)", rig.d.Olc.Mode)
	}
}

func TestOeditDispVal2_ContainerCallsContainerFlags(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_CONTAINER
	oeditDispVal2Menu(rig.d)
	out := rig.readBuf(t)
	if !strings.Contains(out, "Container flags") {
		t.Errorf("expected container-flags menu; got: %q", out)
	}
}

func TestOeditDispVal3_DrinkConCallsLiquid(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_DRINK_CON
	oeditDispVal3Menu(rig.d)
	out := rig.readBuf(t)
	if !strings.Contains(out, "water") || !strings.Contains(out, "liquid") {
		t.Errorf("expected liquid menu; got: %q", out)
	}
}

func TestOeditDispVal4_WeaponCallsWeaponMenu(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_WEAPON
	oeditDispVal4Menu(rig.d)
	out := rig.readBuf(t)
	if !strings.Contains(out, "weapon type") {
		t.Errorf("expected weapon-type menu; got: %q", out)
	}
}

// -----------------------------------------------------------------------------
// G7 — Per-item-type value-menu PARSE tests
// -----------------------------------------------------------------------------

func TestOeditParse_Val1_WeaponSetsValue0(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_WEAPON
	rig.d.Olc.Mode = types.OEDIT_VALUE_1
	oeditParse(rig.d, "50")
	if rig.idx.Value[0] != 50 {
		t.Errorf("Value[0] = %d, want 50", rig.idx.Value[0])
	}
	// Advances to val2 menu (weapon val2 = "Number of damage dice").
	if rig.d.Olc.Mode != types.OEDIT_VALUE_2 {
		t.Errorf("Mode = %d, want OEDIT_VALUE_2", rig.d.Olc.Mode)
	}
}

func TestOeditParse_Val1_LeverToggleTrigFlag(t *testing.T) {
	// Lever: numeric 1 toggles bit 0 (TRIG_UP) in Value[0].
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_LEVER
	rig.d.Olc.Mode = types.OEDIT_VALUE_1
	oeditParse(rig.d, "1")
	if rig.idx.Value[0]&1 == 0 {
		t.Errorf("expected bit 0 toggled on; got Value[0]=%d", rig.idx.Value[0])
	}
	// After toggle, re-displays val1 menu (stays in OEDIT_VALUE_1).
	if rig.d.Olc.Mode != types.OEDIT_VALUE_1 {
		t.Errorf("Mode = %d, want stay OEDIT_VALUE_1 after toggle", rig.d.Olc.Mode)
	}
}

func TestOeditParse_Val1_LeverZeroAdvances(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_LEVER
	rig.d.Olc.Mode = types.OEDIT_VALUE_1
	oeditParse(rig.d, "0")
	if rig.d.Olc.Mode != types.OEDIT_VALUE_2 {
		t.Errorf("Mode = %d, want OEDIT_VALUE_2 on lever 0 (advance)", rig.d.Olc.Mode)
	}
}

func TestOeditParse_Val2_ContainerTogglesFlag(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_CONTAINER
	rig.d.Olc.Mode = types.OEDIT_VALUE_2
	oeditParse(rig.d, "1")
	if rig.idx.Value[1]&1 == 0 {
		t.Errorf("expected CONT_CLOSEABLE bit set; got Value[1]=%d", rig.idx.Value[1])
	}
	// After toggle, stay in OEDIT_VALUE_2 (re-display menu) per C :1602.
	if rig.d.Olc.Mode != types.OEDIT_VALUE_2 {
		t.Errorf("Mode = %d, want stay OEDIT_VALUE_2 after toggle (advancing means no toggle happened)", rig.d.Olc.Mode)
	}
}

// TestOeditParse_Val2_ContainerTogglesOff pins the XOR semantics: a
// second toggle clears the bit (distinguishes toggle from plain set).
func TestOeditParse_Val2_ContainerTogglesOff(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_CONTAINER
	rig.idx.Value[1] = 1 // CONT_CLOSEABLE pre-set
	rig.d.Olc.Mode = types.OEDIT_VALUE_2
	oeditParse(rig.d, "1")
	if rig.idx.Value[1] != 0 {
		t.Errorf("expected toggle-off; got Value[1]=%d", rig.idx.Value[1])
	}
}

func TestOeditParse_Val2_ContainerZeroAdvances(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_CONTAINER
	rig.d.Olc.Mode = types.OEDIT_VALUE_2
	oeditParse(rig.d, "0")
	if rig.d.Olc.Mode != types.OEDIT_VALUE_3 {
		t.Errorf("Mode = %d, want OEDIT_VALUE_3", rig.d.Olc.Mode)
	}
}

func TestOeditParse_Val3_LightAllowsNegativeOne(t *testing.T) {
	// Default branch (ITEM_LIGHT is default for val3 — min=-32000, max=32000).
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_LIGHT
	rig.d.Olc.Mode = types.OEDIT_VALUE_3
	oeditParse(rig.d, "-1")
	if rig.idx.Value[2] != -1 {
		t.Errorf("Value[2] = %d, want -1 (infinite light)", rig.idx.Value[2])
	}
}

func TestOeditParse_Val3_WeaponClampsTo100(t *testing.T) {
	// Weapon val3 range [0, 100].
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_WEAPON
	rig.d.Olc.Mode = types.OEDIT_VALUE_3
	oeditParse(rig.d, "999")
	if rig.idx.Value[2] != 100 {
		t.Errorf("Value[2] = %d, want 100 (clamp)", rig.idx.Value[2])
	}
}

func TestOeditParse_Val3_ScrollNumericSn(t *testing.T) {
	// Scroll val3 takes numeric sn (spell-name word input rejected per Q5).
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_SCROLL
	rig.d.Olc.Mode = types.OEDIT_VALUE_3
	oeditParse(rig.d, "42")
	if rig.idx.Value[2] != 42 {
		t.Errorf("Value[2] = %d, want 42", rig.idx.Value[2])
	}
}

func TestOeditParse_Val3_ScrollRejectsWord(t *testing.T) {
	// Spell-name word rejected per plan §Q5 Option 2.
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_SCROLL
	rig.d.Olc.Mode = types.OEDIT_VALUE_3
	oeditParse(rig.d, "fireball")
	if rig.idx.Value[2] != 0 {
		t.Errorf("Value[2] must not mutate on spell-word input; got %d", rig.idx.Value[2])
	}
	if rig.d.Olc.Mode != types.OEDIT_VALUE_3 {
		t.Errorf("Mode = %d, want stay OEDIT_VALUE_3 after rejection", rig.d.Olc.Mode)
	}
}

func TestOeditParse_Val4_WeaponRejectsOutOfRangeAttack(t *testing.T) {
	// Weapon val4 = attack-type; must range-REJECT (re-prompt, no mutation).
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_WEAPON
	rig.d.Olc.Mode = types.OEDIT_VALUE_4
	oeditParse(rig.d, "9999")
	if rig.idx.Value[3] != 0 {
		t.Errorf("Value[3] must not mutate on OOR attack; got %d", rig.idx.Value[3])
	}
	if rig.d.Olc.Mode != types.OEDIT_VALUE_4 {
		t.Errorf("Mode = %d, want stay OEDIT_VALUE_4 after reject", rig.d.Olc.Mode)
	}
}

// TestOeditParse_Val4_WeaponRejectsBoundaryAttack pins the exact boundary:
// MAX_ATTACK_TYPE itself must be rejected. Distinguishes upper=MAX-1 from
// upper=MAX (mutation gate).
func TestOeditParse_Val4_WeaponRejectsBoundaryAttack(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_WEAPON
	rig.d.Olc.Mode = types.OEDIT_VALUE_4
	oeditParse(rig.d, strconvTestItoa(types.MAX_ATTACK_TYPE))
	if rig.idx.Value[3] == types.MAX_ATTACK_TYPE {
		t.Errorf("Value[3] must not accept MAX_ATTACK_TYPE; got %d", rig.idx.Value[3])
	}
	if rig.d.Olc.Mode != types.OEDIT_VALUE_4 {
		t.Errorf("Mode = %d, want stay OEDIT_VALUE_4 at boundary", rig.d.Olc.Mode)
	}
}

func TestOeditParse_Val4_WeaponAcceptsValidAttack(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_WEAPON
	rig.d.Olc.Mode = types.OEDIT_VALUE_4
	oeditParse(rig.d, "3")
	if rig.idx.Value[3] != 3 {
		t.Errorf("Value[3] = %d, want 3", rig.idx.Value[3])
	}
	// val4 advances to val5; val5 for weapon has no prompt (default case)
	// and falls back to main menu per C :828.
	if rig.d.Olc.Mode != types.OEDIT_MAIN_MENU {
		t.Errorf("Mode = %d, want OEDIT_MAIN_MENU (weapon has no val5 prompt)",
			rig.d.Olc.Mode)
	}
}

func TestOeditParse_Val5_FoodMinClampedToZero(t *testing.T) {
	// Food val5 range [0, 32000] — negative rejected.
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_FOOD
	rig.d.Olc.Mode = types.OEDIT_VALUE_5
	oeditParse(rig.d, "-10")
	if rig.idx.Value[4] != 0 {
		t.Errorf("Value[4] = %d, want 0 (clamp)", rig.idx.Value[4])
	}
}

func TestOeditParse_Val6_SalveNumericSn(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_SALVE
	rig.d.Olc.Mode = types.OEDIT_VALUE_6
	oeditParse(rig.d, "15")
	if rig.idx.Value[5] != 15 {
		t.Errorf("Value[5] = %d, want 15", rig.idx.Value[5])
	}
	if rig.d.Olc.Mode != types.OEDIT_MAIN_MENU {
		t.Errorf("Mode = %d, want OEDIT_MAIN_MENU (val6 → main)", rig.d.Olc.Mode)
	}
}

func TestOeditParse_Val6_DefaultReturnsMain(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.ItemType = types.ITEM_TREASURE
	rig.d.Olc.Mode = types.OEDIT_VALUE_6
	oeditParse(rig.d, "42")
	if rig.idx.Value[5] != 42 {
		t.Errorf("Value[5] = %d, want 42", rig.idx.Value[5])
	}
	if rig.d.Olc.Mode != types.OEDIT_MAIN_MENU {
		t.Errorf("Mode = %d, want OEDIT_MAIN_MENU", rig.d.Olc.Mode)
	}
}

// -----------------------------------------------------------------------------
// G8 — Extradesc arms
// -----------------------------------------------------------------------------

func TestOeditParse_ExtradescAdd_ReturnsChoice(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_EXTRADESC_MENU
	oeditParse(rig.d, "A")
	if rig.d.Olc.Mode != types.OEDIT_EXTRADESC_CHOICE {
		t.Errorf("Mode = %d, want OEDIT_EXTRADESC_CHOICE after A", rig.d.Olc.Mode)
	}
	if len(rig.idx.ExtraDescr) != 1 {
		t.Errorf("ExtraDescr len = %d, want 1", len(rig.idx.ExtraDescr))
	}
	if rig.d.Olc.Spare == nil {
		t.Errorf("Spare should hold the new extradesc")
	}
}

func TestOeditParse_ExtradescDelete_ByIndex(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.ExtraDescr = []*types.ExtraDescrData{
		{Keyword: "first", Description: "desc1"},
		{Keyword: "second", Description: "desc2"},
	}
	rig.d.Olc.Mode = types.OEDIT_EXTRADESC_DELETE
	oeditParse(rig.d, "1")
	if len(rig.idx.ExtraDescr) != 1 {
		t.Errorf("ExtraDescr len = %d, want 1", len(rig.idx.ExtraDescr))
	}
	if rig.idx.ExtraDescr[0].Keyword != "second" {
		t.Errorf("survivor = %q, want 'second'", rig.idx.ExtraDescr[0].Keyword)
	}
}

func TestOeditParse_ExtradescDelete_OutOfRangeReprompts(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.ExtraDescr = []*types.ExtraDescrData{{Keyword: "a", Description: "desc"}}
	rig.d.Olc.Mode = types.OEDIT_EXTRADESC_DELETE
	oeditParse(rig.d, "99")
	if len(rig.idx.ExtraDescr) != 1 {
		t.Errorf("ExtraDescr len = %d, want 1 (untouched)", len(rig.idx.ExtraDescr))
	}
	out := rig.readBuf(t)
	if !strings.Contains(out, "not found") {
		t.Errorf("expected not-found message; got: %q", out)
	}
}

func TestOeditParse_ExtradescKey_StoresKeyword(t *testing.T) {
	rig := newOeditHarness(t)
	ed := &types.ExtraDescrData{Keyword: "old", Description: "d"}
	rig.idx.ExtraDescr = []*types.ExtraDescrData{ed}
	rig.d.Olc.Spare = ed
	rig.d.Olc.Mode = types.OEDIT_EXTRADESC_KEY
	oeditParse(rig.d, "new keyword")
	if ed.Keyword != "new keyword" {
		t.Errorf("Keyword = %q, want 'new keyword'", ed.Keyword)
	}
	if rig.d.Olc.Mode != types.OEDIT_EXTRADESC_CHOICE {
		t.Errorf("Mode = %d, want OEDIT_EXTRADESC_CHOICE", rig.d.Olc.Mode)
	}
}

func TestOeditParse_ExtradescKey_SmashesTilde(t *testing.T) {
	rig := newOeditHarness(t)
	ed := &types.ExtraDescrData{}
	rig.idx.ExtraDescr = []*types.ExtraDescrData{ed}
	rig.d.Olc.Spare = ed
	rig.d.Olc.Mode = types.OEDIT_EXTRADESC_KEY
	oeditParse(rig.d, "evil~tilde")
	if strings.Contains(ed.Keyword, "~") {
		t.Errorf("tilde should be smashed; got %q", ed.Keyword)
	}
}

func TestOeditParse_ExtradescChoice_QReturnsMenuNoJunk(t *testing.T) {
	// C ooedit does NOT junk empty extradescs on Q (divergence from redit).
	rig := newOeditHarness(t)
	ed := &types.ExtraDescrData{Keyword: "", Description: ""}
	rig.idx.ExtraDescr = []*types.ExtraDescrData{ed}
	rig.d.Olc.Spare = ed
	rig.d.Olc.Mode = types.OEDIT_EXTRADESC_CHOICE
	oeditParse(rig.d, "0")
	// ed must still be present.
	if len(rig.idx.ExtraDescr) != 1 {
		t.Errorf("ExtraDescr len = %d, want 1 (no junk-on-Q per C divergence)",
			len(rig.idx.ExtraDescr))
	}
	if rig.d.Olc.Mode != types.OEDIT_EXTRADESC_MENU {
		t.Errorf("Mode = %d, want OEDIT_EXTRADESC_MENU", rig.d.Olc.Mode)
	}
}

func TestOeditParse_ExtradescChoice_1EntersKeyPrompt(t *testing.T) {
	rig := newOeditHarness(t)
	ed := &types.ExtraDescrData{Keyword: "", Description: ""}
	rig.idx.ExtraDescr = []*types.ExtraDescrData{ed}
	rig.d.Olc.Spare = ed
	rig.d.Olc.Mode = types.OEDIT_EXTRADESC_CHOICE
	oeditParse(rig.d, "1")
	if rig.d.Olc.Mode != types.OEDIT_EXTRADESC_KEY {
		t.Errorf("Mode = %d, want OEDIT_EXTRADESC_KEY", rig.d.Olc.Mode)
	}
}

func TestOeditParse_ExtradescChoice_2EntersEditor(t *testing.T) {
	rig := newOeditHarness(t)
	ed := &types.ExtraDescrData{Keyword: "kw", Description: "original"}
	rig.idx.ExtraDescr = []*types.ExtraDescrData{ed}
	rig.d.Olc.Spare = ed
	rig.d.Olc.Mode = types.OEDIT_EXTRADESC_CHOICE
	oeditParse(rig.d, "2")
	if rig.d.Character.Substate != types.SUB_OBJ_EXTRA {
		t.Errorf("Substate = %d, want SUB_OBJ_EXTRA", rig.d.Character.Substate)
	}
	if rig.d.Character.EditorSave == nil {
		t.Fatal("EditorSave closure not registered")
	}
	if rig.d.Olc.Mode != types.OEDIT_EXTRADESC_DESCRIPTION {
		t.Errorf("Mode = %d, want OEDIT_EXTRADESC_DESCRIPTION", rig.d.Olc.Mode)
	}
}

// TestOeditParse_ExtradescDescription_RestoresConOedit is mutation gate
// #12 — swap CON_OEDIT → CON_PLAYING in the description-editor closure
// and this fails.
func TestOeditParse_ExtradescDescription_RestoresConOedit(t *testing.T) {
	rig := newOeditHarness(t)
	ed := &types.ExtraDescrData{Keyword: "kw", Description: "orig"}
	rig.idx.ExtraDescr = []*types.ExtraDescrData{ed}
	rig.d.Olc.Spare = ed
	rig.d.Olc.Mode = types.OEDIT_EXTRADESC_CHOICE
	oeditParse(rig.d, "2")
	if rig.d.Character.EditorSave == nil {
		t.Fatal("EditorSave closure missing")
	}
	rig.d.Character.Editor = &types.EditorData{NumLines: 1}
	rig.d.Character.Editor.Lines[0] = "A new description."
	rig.d.Character.EditorSave(rig.d.Character)
	if rig.d.Connected != int(types.CON_OEDIT) {
		t.Errorf("Connected = %d, want CON_OEDIT after /s", rig.d.Connected)
	}
	if ed.Description == "orig" {
		t.Errorf("Description must have been updated via EditorSave")
	}
}

func TestOeditParse_ExtradescMenu_QReturnsToMain(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_EXTRADESC_MENU
	oeditParse(rig.d, "Q")
	if rig.d.Olc.Mode != types.OEDIT_MAIN_MENU {
		t.Errorf("Mode = %d, want OEDIT_MAIN_MENU after Q", rig.d.Olc.Mode)
	}
}

func TestOeditParse_ExtradescMenu_REntersDeleteMode(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_EXTRADESC_MENU
	oeditParse(rig.d, "R")
	if rig.d.Olc.Mode != types.OEDIT_EXTRADESC_DELETE {
		t.Errorf("Mode = %d, want OEDIT_EXTRADESC_DELETE", rig.d.Olc.Mode)
	}
}

func TestOeditParse_ExtradescMenu_NumericEditsByIndex(t *testing.T) {
	rig := newOeditHarness(t)
	ed := &types.ExtraDescrData{Keyword: "pick", Description: "me"}
	rig.idx.ExtraDescr = []*types.ExtraDescrData{ed}
	rig.d.Olc.Mode = types.OEDIT_EXTRADESC_MENU
	oeditParse(rig.d, "1")
	if rig.d.Olc.Mode != types.OEDIT_EXTRADESC_CHOICE {
		t.Errorf("Mode = %d, want OEDIT_EXTRADESC_CHOICE", rig.d.Olc.Mode)
	}
	if rig.d.Olc.Spare != ed {
		t.Errorf("Spare should point at the selected ed")
	}
}

// -----------------------------------------------------------------------------
// G9 — Affect arms (placeholder per plan §Q4)
// -----------------------------------------------------------------------------

func TestOeditParse_AffectMenu_QReturnsToMain(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_AFFECT_MENU
	oeditParse(rig.d, "Q")
	if rig.d.Olc.Mode != types.OEDIT_MAIN_MENU {
		t.Errorf("Mode = %d, want OEDIT_MAIN_MENU after Q", rig.d.Olc.Mode)
	}
	if rig.d.Olc.Spare != nil {
		t.Errorf("Spare should be cleared on Q")
	}
}

func TestOeditParse_AffectMenu_AEntersLocation(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_AFFECT_MENU
	oeditParse(rig.d, "A")
	if rig.d.Olc.Mode != types.OEDIT_AFFECT_LOCATION {
		t.Errorf("Mode = %d, want OEDIT_AFFECT_LOCATION", rig.d.Olc.Mode)
	}
	if rig.d.Olc.Spare == nil {
		t.Fatal("Spare should hold pending AffectData")
	}
	if _, ok := rig.d.Olc.Spare.(*types.AffectData); !ok {
		t.Errorf("Spare should be *AffectData; got %T", rig.d.Olc.Spare)
	}
}

// TestOeditParse_AffectLocation_RejectsExtAffect pins mutation gate #10:
// drop the APPLY_EXT_AFFECT check → this test fails.
func TestOeditParse_AffectLocation_RejectsExtAffect(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_AFFECT_MENU
	oeditParse(rig.d, "A")
	// Now in OEDIT_AFFECT_LOCATION with pending paf.
	oeditParse(rig.d, strconvTestItoa(types.APPLY_EXT_AFFECT))
	// Should stay in OEDIT_AFFECT_LOCATION; paf location unchanged.
	paf := rig.d.Olc.Spare.(*types.AffectData)
	if paf.Location == types.APPLY_EXT_AFFECT {
		t.Errorf("Location must not be APPLY_EXT_AFFECT")
	}
	if rig.d.Olc.Mode != types.OEDIT_AFFECT_LOCATION {
		t.Errorf("Mode = %d, want stay OEDIT_AFFECT_LOCATION", rig.d.Olc.Mode)
	}
}

func TestOeditParse_AffectLocation_ZeroJunks(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_AFFECT_MENU
	oeditParse(rig.d, "A")
	oeditParse(rig.d, "0")
	if rig.d.Olc.Spare != nil {
		t.Errorf("Spare should be cleared when location=0")
	}
	if rig.d.Olc.Mode != types.OEDIT_AFFECT_MENU {
		t.Errorf("Mode = %d, want OEDIT_AFFECT_MENU after junk", rig.d.Olc.Mode)
	}
}

func TestOeditParse_AffectLocation_SimpleAtoi(t *testing.T) {
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_AFFECT_MENU
	oeditParse(rig.d, "A")
	// APPLY_STR=1, APPLY_HITROLL=15 — both ordinary numeric locations.
	oeditParse(rig.d, strconvTestItoa(types.APPLY_HITROLL))
	if rig.d.Olc.Mode != types.OEDIT_AFFECT_MODIFIER {
		t.Errorf("Mode = %d, want OEDIT_AFFECT_MODIFIER", rig.d.Olc.Mode)
	}
	paf := rig.d.Olc.Spare.(*types.AffectData)
	if paf.Location != types.APPLY_HITROLL {
		t.Errorf("Location = %d, want APPLY_HITROLL", paf.Location)
	}
}

func TestOeditParse_AffectModifier_AppendsAffect(t *testing.T) {
	// End-to-end: Add → location=APPLY_HITROLL → modifier=5 → appears in
	// idx.Affects.
	rig := newOeditHarness(t)
	rig.d.Olc.Mode = types.OEDIT_AFFECT_MENU
	oeditParse(rig.d, "A")
	oeditParse(rig.d, strconvTestItoa(types.APPLY_HITROLL))
	oeditParse(rig.d, "5")
	if len(rig.idx.Affects) != 1 {
		t.Fatalf("Affects len = %d, want 1", len(rig.idx.Affects))
	}
	if rig.idx.Affects[0].Location != types.APPLY_HITROLL {
		t.Errorf("Location = %d, want APPLY_HITROLL", rig.idx.Affects[0].Location)
	}
	if rig.idx.Affects[0].Modifier != 5 {
		t.Errorf("Modifier = %d, want 5", rig.idx.Affects[0].Modifier)
	}
	if rig.d.Olc.Spare != nil {
		t.Errorf("Spare should be cleared after commit")
	}
}

func TestOeditParse_AffectRemove_ByIndex(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.Affects = []*types.AffectData{
		{Location: types.APPLY_STR, Modifier: 1},
		{Location: types.APPLY_HITROLL, Modifier: 2},
	}
	rig.d.Olc.Mode = types.OEDIT_AFFECT_REMOVE
	oeditParse(rig.d, "1")
	if len(rig.idx.Affects) != 1 {
		t.Errorf("Affects len = %d, want 1", len(rig.idx.Affects))
	}
	if rig.idx.Affects[0].Location != types.APPLY_HITROLL {
		t.Errorf("survivor location = %d, want APPLY_HITROLL", rig.idx.Affects[0].Location)
	}
}

func TestOeditParse_AffectRemove_OutOfRangeNoOp(t *testing.T) {
	rig := newOeditHarness(t)
	rig.idx.Affects = []*types.AffectData{{Location: types.APPLY_STR, Modifier: 1}}
	rig.d.Olc.Mode = types.OEDIT_AFFECT_REMOVE
	oeditParse(rig.d, "99")
	if len(rig.idx.Affects) != 1 {
		t.Errorf("Affects len = %d, want 1 (untouched)", len(rig.idx.Affects))
	}
}

func TestOeditParse_AffectMenu_RemoveWithNumber(t *testing.T) {
	// "R 1" shortcut: one-shot remove + return to menu.
	rig := newOeditHarness(t)
	rig.idx.Affects = []*types.AffectData{{Location: types.APPLY_STR, Modifier: 1}}
	rig.d.Olc.Mode = types.OEDIT_AFFECT_MENU
	oeditParse(rig.d, "R 1")
	if len(rig.idx.Affects) != 0 {
		t.Errorf("Affects len = %d, want 0", len(rig.idx.Affects))
	}
	if rig.d.Olc.Mode != types.OEDIT_AFFECT_MENU {
		t.Errorf("Mode = %d, want OEDIT_AFFECT_MENU", rig.d.Olc.Mode)
	}
}
