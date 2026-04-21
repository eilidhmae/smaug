package game

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// --- G2 NPC + PC main-menu renderer tests (A5, A6, A7) ---

// newNpcVictimForRender builds an NPC victim with minimal prototype fields
// so the NPC renderer can exercise IndexData->dice paths without segfault.
func newNpcVictimForRender() *types.CharData {
	idx := &types.MobIndexData{
		Vnum:        1234,
		ShortDescr:  "a test shopkeeper",
		LongDescr:   "A shopkeeper stands here.",
		Description: "It looks ordinary.",
		HitNoDice:   5,
		HitSizeDice: 8,
		HitPlus:     10,
		DamNoDice:   2,
		DamSizeDice: 6,
		DamPlus:     1,
	}
	v := &types.CharData{
		Name:        "shopkeeper",
		Description: "It looks ordinary.",
		Level:       15,
		Alignment:   100,
		Gold:        250,
		Position:    types.POS_STANDING,
		DefPosition: types.POS_STANDING,
		IndexData:   idx,
		PermStr:     13,
		PermInt:     12,
		PermWis:     11,
		PermDex:     14,
		PermCon:     15,
		PermCha:     10,
		PermLck:     9,
	}
	v.Act.Set(types.ACT_IS_NPC)
	return v
}

// newPcVictimForRender builds a PC victim (no ACT_IS_NPC; PCData populated).
func newPcVictimForRender() *types.CharData {
	v := &types.CharData{
		Name:        "Alice",
		Description: "Alice is here.",
		Level:       40,
		Alignment:   0,
		Hit:         200, MaxHit: 200,
		Mana: 100, MaxMana: 100,
		Move: 150, MaxMove: 150,
		Gold:           500,
		MentalState:    10,
		EmotionalState: 5,
		Position:       types.POS_STANDING,
		PermStr:        16, PermInt: 15, PermWis: 13, PermDex: 14, PermCon: 12, PermCha: 11, PermLck: 10,
		PCData: &types.PCData{
			Favor: 0,
		},
	}
	// PC: ACT_IS_NPC deliberately NOT set.
	return v
}

func TestMeditDispNpcMenu_RendersAllFieldLabels(t *testing.T) {
	rig := newMeditHarness(t)
	rig.d.Olc.Target = newNpcVictimForRender()

	MeditDispMenu(rig.d)
	out := rig.readBuf(t)

	// The plan §A5 gate: ≥20 documented field labels present. Matches C
	// medit_disp_npc_menu at omedit.c:827-884.
	required := []string{
		"Mob Number",
		"Sex:", "Name:", "Shortdesc:", "Longdesc",
		"Description",
		"Class:", "Race:", "Level:",
		"Alignment:", "Strength:", "Intelligence:",
		"Dexterity:", "Constitution:", "Charisma:", "Luck:",
		"DamNumDice", "DamSizeDice", "DamPlus",
		"HitNumDice", "HitSizeDice", "HitPlus",
		"Spec:", "Saving Throws",
		"Resistant", "Immune", "Susceptible",
		"Position", "Attacks", "Defenses",
		"Body Parts", "Act Flags", "Affected",
		"Quit", "Enter choice",
	}
	for _, s := range required {
		if !strings.Contains(out, s) {
			t.Errorf("NPC menu missing %q; got: %q", s, out)
		}
	}
	// Mode is pinned back to NPC_MAIN_MENU after render (C olc_mode convention).
	if rig.d.Olc.Mode != types.MEDIT_NPC_MAIN_MENU {
		t.Errorf("Mode after NPC render = %d, want MEDIT_NPC_MAIN_MENU", rig.d.Olc.Mode)
	}
}

func TestMeditDispPcMenu_RendersAllFieldLabels(t *testing.T) {
	rig := newMeditHarness(t)
	rig.d.Olc.Target = newPcVictimForRender()
	// High-trust builder exercises the optional Deity/Clan/Council rows.
	rig.d.Character.Trust = types.LEVEL_IMPLEMENTOR

	MeditDispMenu(rig.d)
	out := rig.readBuf(t)

	required := []string{
		"Sex:", "Name:", "Description",
		"Class:", "Race:", "Level:", "Alignment:",
		"Strength:", "Intelligence:", "Dexterity:",
		"Constitution:", "Charisma:", "Luck:",
		"Hps:", "Mana:", "Move",
		"Mentalstate", "Emotional",
		"Thirst", "Full", "Drunk", "Favor",
		"Saving Throws",
		"Resistant", "Immune", "Susceptible",
		"Position", "Act Flags", "PC Flags", "Affected",
		"Deity", "Quit", "Enter choice",
	}
	for _, s := range required {
		if !strings.Contains(out, s) {
			t.Errorf("PC menu missing %q; got: %q", s, out)
		}
	}
	if rig.d.Olc.Mode != types.MEDIT_PC_MAIN_MENU {
		t.Errorf("Mode after PC render = %d, want MEDIT_PC_MAIN_MENU", rig.d.Olc.Mode)
	}
}

// A7: MeditDispMenu routes by victim.IsNPC() — a mutation flipping the
// branch (routing NPC→PC renderer) fails this test.
func TestMeditDispMenu_RoutesByIsNpc(t *testing.T) {
	t.Run("NPC victim calls NPC renderer", func(t *testing.T) {
		rig := newMeditHarness(t)
		rig.d.Olc.Target = newNpcVictimForRender()
		MeditDispMenu(rig.d)
		out := rig.readBuf(t)
		if !strings.Contains(out, "Mob Number") {
			t.Errorf("expected NPC-specific 'Mob Number' label, got: %q", out)
		}
	})
	t.Run("PC victim calls PC renderer", func(t *testing.T) {
		rig := newMeditHarness(t)
		rig.d.Olc.Target = newPcVictimForRender()
		MeditDispMenu(rig.d)
		out := rig.readBuf(t)
		// PC menu has "Hps:" which NPC does not.
		if !strings.Contains(out, "Hps:") {
			t.Errorf("expected PC-specific 'Hps:' label, got: %q", out)
		}
		// And it must NOT contain the NPC-only "Mob Number" label.
		if strings.Contains(out, "Mob Number") {
			t.Errorf("PC render leaked NPC 'Mob Number' label: %q", out)
		}
	})
}

// --- G3 submenu renderer snapshot tests (A8) ---

func TestMeditDispSubMenus_AllPresentAndEmitLabels(t *testing.T) {
	victim := newNpcVictimForRender()
	cases := []struct {
		name string
		fn   func(d *types.DescriptorData)
		want []string
	}{
		{"sex", meditDispSexMenu, []string{"Sex", "Male", "Female", "Neutral"}},
		{"pos", meditDispPosMenu, []string{"Position", "standing"}},
		{"defaultPos", meditDispDefaultPosMenu, []string{"Default Position", "standing"}},
		{"attack", meditDispAttackMenu, []string{"Attacks"}},
		{"defense", meditDispDefenseMenu, []string{"Defenses"}},
		{"spec", meditDispSpecMenu, []string{"Spec"}},
		{"class", meditDispClassMenu, []string{"Class"}},
		{"race", meditDispRaceMenu, []string{"Race"}},
		{"save", meditDispSaveMenu, []string{"Poison", "Wand", "Paralysis", "Breath", "Spell"}},
		{"affect", meditDispAffectMenu, []string{"Affect"}},
		{"npcFlags", meditDispNpcFlagsMenu, []string{"Act Flags"}},
		{"pcFlags", meditDispPcFlagsMenu, []string{"PC Flags"}},
		{"affFlags", meditDispAffFlagsMenu, []string{"Affect"}},
		{"pcdataFlags", meditDispPcdataFlagsMenu, []string{"PCData Flags"}},
		{"parts", meditDispPartsMenu, []string{"Body Parts"}},
		{"ris", meditDispRisMenu, []string{"Resistant"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rig := newMeditHarness(t)
			rig.d.Olc.Target = victim
			tc.fn(rig.d)
			out := rig.readBuf(t)
			for _, s := range tc.want {
				if !strings.Contains(out, s) {
					t.Errorf("%s submenu missing %q; got: %q", tc.name, s, out)
				}
			}
		})
	}
}

// --- G5 NPC main-menu dispatch (A11) ---

// npcDigitTable mirrors plan §230-269 and C omedit.c:1059-1215. The
// expected MEDIT_* mode is what the dispatcher MUST land on after reading
// the digit — verified by directly asserting d.Olc.Mode.
var npcDigitTable = []struct {
	in   string
	mode int
}{
	{"1", types.MEDIT_SEX},
	{"2", types.MEDIT_NAME},
	{"3", types.MEDIT_S_DESC},
	{"4", types.MEDIT_L_DESC},
	{"5", types.MEDIT_D_DESC},
	{"6", types.MEDIT_CLASS},
	{"7", types.MEDIT_RACE},
	{"8", types.MEDIT_LEVEL},
	{"9", types.MEDIT_ALIGNMENT},
	{"A", types.MEDIT_STRENGTH},
	{"B", types.MEDIT_INTELLIGENCE},
	{"C", types.MEDIT_WISDOM},
	{"D", types.MEDIT_DEXTERITY},
	{"E", types.MEDIT_CONSTITUTION},
	{"F", types.MEDIT_CHARISMA},
	{"G", types.MEDIT_LUCK},
	{"H", types.MEDIT_DAMNUMDIE},
	{"I", types.MEDIT_DAMSIZEDIE},
	{"J", types.MEDIT_DAMPLUS},
	{"K", types.MEDIT_HITNUMDIE},
	{"L", types.MEDIT_HITSIZEDIE},
	{"M", types.MEDIT_HITPLUS},
	{"N", types.MEDIT_GOLD},
	{"O", types.MEDIT_SPEC},
	{"P", types.MEDIT_SAVE_MENU},
	{"R", types.MEDIT_RESISTANT},
	{"S", types.MEDIT_IMMUNE},
	{"T", types.MEDIT_SUSCEPTIBLE},
	{"U", types.MEDIT_POS},
	{"V", types.MEDIT_ATTACK},
	{"W", types.MEDIT_DEFENSE},
	{"X", types.MEDIT_PARTS},
	{"Y", types.MEDIT_NPC_FLAGS},
	{"Z", types.MEDIT_AFF_FLAGS},
}

func TestMeditDispatchNpcMain_EveryDigitMapsToDocumentedMode(t *testing.T) {
	for _, tc := range npcDigitTable {
		t.Run(tc.in, func(t *testing.T) {
			rig := newMeditHarness(t)
			rig.d.Olc.Target = newNpcVictimForRender()
			meditParse(rig.d, tc.in)
			if rig.d.Olc == nil {
				t.Fatalf("digit %q: Olc cleared (Q path hit by mistake)", tc.in)
			}
			if rig.d.Olc.Mode != tc.mode {
				t.Errorf("digit %q: Mode = %d, want %d", tc.in, rig.d.Olc.Mode, tc.mode)
			}
		})
	}
}

// NPC Q exits immediately (cleanup) — no save-confirm for NPC per C :1062-1064.
func TestMeditDispatchNpcMain_QuitCleansUp(t *testing.T) {
	rig := newMeditHarness(t)
	rig.d.Olc.Target = newNpcVictimForRender()
	meditParse(rig.d, "Q")
	if rig.d.Connected != int(types.CON_PLAYING) {
		t.Errorf("NPC Q Connected = %d, want CON_PLAYING", rig.d.Connected)
	}
	if rig.d.Olc != nil {
		t.Errorf("NPC Q Olc = %+v, want nil (cleanup)", rig.d.Olc)
	}
}

// Unknown NPC-menu digit redisplays (mirrors C default: medit_disp_npc_menu).
func TestMeditDispatchNpcMain_UnknownDigitRedisplays(t *testing.T) {
	rig := newMeditHarness(t)
	rig.d.Olc.Target = newNpcVictimForRender()
	meditParse(rig.d, "`") // not a valid digit
	if rig.d.Olc == nil {
		t.Fatal("Olc nil after unknown digit — should NOT cleanup")
	}
	if rig.d.Olc.Mode != types.MEDIT_NPC_MAIN_MENU {
		t.Errorf("Mode after unknown NPC digit = %d, want MEDIT_NPC_MAIN_MENU", rig.d.Olc.Mode)
	}
	out := rig.readBuf(t)
	if !strings.Contains(out, "Mob Number") {
		t.Errorf("expected NPC main menu redisplay, got: %q", out)
	}
}

// --- G6 PC main-menu dispatch (A12, A13) ---

// pcDigitTable mirrors plan §270-310 and C omedit.c:1218-1386.
var pcDigitTable = []struct {
	in   string
	mode int
}{
	{"1", types.MEDIT_SEX},
	{"2", types.MEDIT_NAME},
	{"3", types.MEDIT_D_DESC},
	{"4", types.MEDIT_CLASS},
	{"5", types.MEDIT_RACE},
	// "6" rejected — NPC Only. Tested separately.
	{"7", types.MEDIT_ALIGNMENT},
	{"8", types.MEDIT_STRENGTH},
	{"9", types.MEDIT_INTELLIGENCE},
	{"A", types.MEDIT_WISDOM},
	{"B", types.MEDIT_DEXTERITY},
	{"C", types.MEDIT_CONSTITUTION},
	{"D", types.MEDIT_CHARISMA},
	{"E", types.MEDIT_LUCK},
	{"F", types.MEDIT_HITPOINT},
	{"G", types.MEDIT_MANA},
	{"H", types.MEDIT_MOVE},
	{"I", types.MEDIT_GOLD},
	{"J", types.MEDIT_MENTALSTATE},
	{"K", types.MEDIT_EMOTIONAL},
	{"L", types.MEDIT_THIRST},
	{"M", types.MEDIT_FULL},
	{"N", types.MEDIT_DRUNK},
	{"O", types.MEDIT_FAVOR},
	{"P", types.MEDIT_SAVE_MENU},
	{"R", types.MEDIT_RESISTANT},
	{"S", types.MEDIT_IMMUNE},
	{"T", types.MEDIT_SUSCEPTIBLE},
	// "U" rejected — NPCs Only. Tested separately.
	{"V", types.MEDIT_PC_FLAGS},
	{"W", types.MEDIT_PCDATA_FLAGS},
	{"X", types.MEDIT_AFF_FLAGS},
	{"Y", types.MEDIT_DEITY},
	{"Z", types.MEDIT_CLAN},
	{"=", types.MEDIT_COUNCIL},
}

// newPcHarness installs a PC victim; the Wave-1 harness defaults to NPC.
// PC Mode is initialized to MEDIT_PC_MAIN_MENU.
func newPcHarness(t *testing.T) *meditTestRig {
	rig := newMeditHarness(t)
	rig.victim = newPcVictimForRender()
	rig.d.Olc.Target = rig.victim
	rig.d.Olc.Mode = types.MEDIT_PC_MAIN_MENU
	// Trust high enough for LEVEL_GREATER gate plus LEVEL_GOD/SUB_IMPLEM gates.
	rig.d.Character.Trust = types.LEVEL_IMPLEMENTOR
	return rig
}

func TestMeditDispatchPcMain_EveryDigitMapsToDocumentedMode(t *testing.T) {
	for _, tc := range pcDigitTable {
		t.Run(tc.in, func(t *testing.T) {
			rig := newPcHarness(t)
			meditParse(rig.d, tc.in)
			if rig.d.Olc == nil {
				t.Fatalf("digit %q: Olc cleared unexpectedly", tc.in)
			}
			if rig.d.Olc.Mode != tc.mode {
				t.Errorf("digit %q: Mode = %d, want %d", tc.in, rig.d.Olc.Mode, tc.mode)
			}
		})
	}
}

// PC digit 6 → "NPC Only!!" verbatim per C :1255-1257. Mode unchanged.
func TestMeditDispatchPcMain_Digit6RejectsNpcOnly(t *testing.T) {
	rig := newPcHarness(t)
	meditParse(rig.d, "6")
	out := rig.readBuf(t)
	if !strings.Contains(out, "NPC Only!!") {
		t.Errorf("expected 'NPC Only!!' rejection, got: %q", out)
	}
	if rig.d.Olc.Mode != types.MEDIT_PC_MAIN_MENU {
		t.Errorf("Mode = %d, want MEDIT_PC_MAIN_MENU (digit 6 on PC must not transition)", rig.d.Olc.Mode)
	}
}

// PC digit U → "NPCs Only!!" verbatim per C :1351-1353.
func TestMeditDispatchPcMain_DigitURejectsNpcsOnly(t *testing.T) {
	rig := newPcHarness(t)
	meditParse(rig.d, "U")
	out := rig.readBuf(t)
	if !strings.Contains(out, "NPCs Only!!") {
		t.Errorf("expected 'NPCs Only!!' rejection, got: %q", out)
	}
	if rig.d.Olc.Mode != types.MEDIT_PC_MAIN_MENU {
		t.Errorf("Mode = %d, want MEDIT_PC_MAIN_MENU", rig.d.Olc.Mode)
	}
}

// A13 trust gate: LEVEL_GREATER required on LEVEL/CLASS/RACE for PC.
// LEVEL_IMMORTAL rejects; LEVEL_GREATER accepts.
func TestMeditDispatchPcMain_TrustGateOnLevelClassRace(t *testing.T) {
	for _, digit := range []struct {
		in   string
		mode int
	}{
		{"4", types.MEDIT_CLASS},
		{"5", types.MEDIT_RACE},
		// Note: PC menu has no LEVEL digit in stock C (digit "6" is NPC-only reject).
		// The plan §G6 says LEVEL_GREATER gate applies on LEVEL/CLASS/RACE; since PC
		// menu does not route to MEDIT_LEVEL, only CLASS/RACE digit-routes are trust-gated.
	} {
		t.Run("reject_below_greater/"+digit.in, func(t *testing.T) {
			rig := newPcHarness(t)
			rig.d.Character.Trust = types.LEVEL_IMMORTAL // below LEVEL_GREATER
			meditParse(rig.d, digit.in)
			if rig.d.Olc.Mode == digit.mode {
				t.Errorf("digit %q at LEVEL_IMMORTAL transitioned to %d; trust gate failed", digit.in, digit.mode)
			}
		})
		t.Run("accept_at_greater/"+digit.in, func(t *testing.T) {
			rig := newPcHarness(t)
			rig.d.Character.Trust = types.LEVEL_GREATER
			meditParse(rig.d, digit.in)
			if rig.d.Olc.Mode != digit.mode {
				t.Errorf("digit %q at LEVEL_GREATER did not transition to %d (Mode=%d)", digit.in, digit.mode, rig.d.Olc.Mode)
			}
		})
	}
}

// A12 trust gates on PC menu: Z (Clan) requires LEVEL_GOD,
// = (Council) requires LEVEL_SUB_IMPLEM.
// Mirrors C omedit.c:1371-1380.
func TestMeditDispatchPcMain_TrustGateOnClanCouncil(t *testing.T) {
	t.Run("Z_rejected_below_LEVEL_GOD", func(t *testing.T) {
		rig := newPcHarness(t)
		rig.d.Character.Trust = types.LEVEL_GOD - 1
		meditParse(rig.d, "Z")
		if rig.d.Olc.Mode == types.MEDIT_CLAN {
			t.Errorf("digit Z at trust=LEVEL_GOD-1 transitioned to MEDIT_CLAN; gate failed")
		}
	})
	t.Run("Z_accepted_at_LEVEL_GOD", func(t *testing.T) {
		rig := newPcHarness(t)
		rig.d.Character.Trust = types.LEVEL_GOD
		meditParse(rig.d, "Z")
		if rig.d.Olc.Mode != types.MEDIT_CLAN {
			t.Errorf("digit Z at LEVEL_GOD did not transition to MEDIT_CLAN (Mode=%d)", rig.d.Olc.Mode)
		}
	})
	t.Run("Equals_rejected_below_LEVEL_SUB_IMPLEM", func(t *testing.T) {
		rig := newPcHarness(t)
		rig.d.Character.Trust = types.LEVEL_SUB_IMPLEM - 1
		meditParse(rig.d, "=")
		if rig.d.Olc.Mode == types.MEDIT_COUNCIL {
			t.Errorf("digit = at trust=LEVEL_SUB_IMPLEM-1 transitioned to MEDIT_COUNCIL; gate failed")
		}
	})
	t.Run("Equals_accepted_at_LEVEL_SUB_IMPLEM", func(t *testing.T) {
		rig := newPcHarness(t)
		rig.d.Character.Trust = types.LEVEL_SUB_IMPLEM
		meditParse(rig.d, "=")
		if rig.d.Olc.Mode != types.MEDIT_COUNCIL {
			t.Errorf("digit = at LEVEL_SUB_IMPLEM did not transition to MEDIT_COUNCIL (Mode=%d)", rig.d.Olc.Mode)
		}
	})
}

// PC Q with no change → cleanupOlc (no save prompt fires because
// d.Olc.Change is zero in Wave 2).
func TestMeditDispatchPcMain_QuitNoChangeCleansUp(t *testing.T) {
	rig := newPcHarness(t)
	meditParse(rig.d, "Q")
	if rig.d.Connected != int(types.CON_PLAYING) {
		t.Errorf("PC Q Connected = %d, want CON_PLAYING", rig.d.Connected)
	}
	if rig.d.Olc != nil {
		t.Errorf("PC Q Olc = %+v, want nil", rig.d.Olc)
	}
}

// PC Q WITH Change=true enters MEDIT_CONFIRM_SAVESTRING. Wave-2 stub arm
// then routes both Y and N to cleanupOlc (per plan scope: save pathway is
// G11; confirm-savestring arm is a stub).
func TestMeditDispatchPcMain_QuitWithChangeEntersConfirm(t *testing.T) {
	rig := newPcHarness(t)
	rig.d.Olc.Change = true
	meditParse(rig.d, "Q")
	if rig.d.Olc == nil {
		t.Fatal("Olc cleared — should have entered MEDIT_CONFIRM_SAVESTRING")
	}
	if rig.d.Olc.Mode != types.MEDIT_CONFIRM_SAVESTRING {
		t.Errorf("Mode after PC Q with change = %d, want MEDIT_CONFIRM_SAVESTRING", rig.d.Olc.Mode)
	}
}

func TestMeditConfirmSavestring_YAndNBothCleanUp(t *testing.T) {
	for _, in := range []string{"Y", "y", "N", "n"} {
		t.Run(in, func(t *testing.T) {
			rig := newPcHarness(t)
			rig.d.Olc.Change = true
			rig.d.Olc.Mode = types.MEDIT_CONFIRM_SAVESTRING
			meditParse(rig.d, in)
			if rig.d.Connected != int(types.CON_PLAYING) {
				t.Errorf("confirm %q Connected = %d, want CON_PLAYING", in, rig.d.Connected)
			}
		})
	}
}

// Unknown PC digit redisplays the PC main menu (not NPC).
func TestMeditDispatchPcMain_UnknownDigitRedisplaysPcMenu(t *testing.T) {
	rig := newPcHarness(t)
	meditParse(rig.d, "`")
	if rig.d.Olc == nil {
		t.Fatal("Olc nil after unknown digit")
	}
	if rig.d.Olc.Mode != types.MEDIT_PC_MAIN_MENU {
		t.Errorf("Mode = %d, want MEDIT_PC_MAIN_MENU", rig.d.Olc.Mode)
	}
	out := rig.readBuf(t)
	// PC menu has "Hps:" — NPC does not.
	if !strings.Contains(out, "Hps:") {
		t.Errorf("expected PC main menu redisplay, got: %q", out)
	}
}

// Boot wire: the seam's public identity does not depend on where it gets
// assigned; we assert that game.MeditDispMenu is a valid assignment
// target and non-nil when used as a function value.
func TestMeditDispMenu_SeamSymbolResolves(t *testing.T) {
	var f func(*types.DescriptorData) = MeditDispMenu
	if f == nil {
		t.Fatal("MeditDispMenu resolved to nil")
	}
}

// TestMeditParse_UnimplementedArmRedisplaysCorrectMenu pins the Wave-2
// fallthrough default arm: when input arrives on a mode that Wave 2 does
// not yet handle (e.g. MEDIT_NAME — will get its body in G7), meditParse
// redisplays the main menu appropriate to the victim. A mutation that
// swaps the NPC/PC branch in npcOrPcMenu fails this test because an NPC
// victim would get the PC menu (no "Mob Number" label) and vice versa.
// Covers mutation gate M-npcOrPcMenu-branch.
func TestMeditParse_UnimplementedArmRedisplaysCorrectMenu(t *testing.T) {
	t.Run("NPC victim on unimplemented arm gets NPC menu", func(t *testing.T) {
		rig := newMeditHarness(t)
		rig.d.Olc.Target = newNpcVictimForRender()
		rig.d.Olc.Mode = types.MEDIT_NAME // Wave 2 has no arm body yet
		meditParse(rig.d, "some input")
		out := rig.readBuf(t)
		if !strings.Contains(out, "Mob Number") {
			t.Errorf("NPC fallthrough expected NPC menu (Mob Number label); got: %q", out)
		}
		if rig.d.Olc.Mode != types.MEDIT_NPC_MAIN_MENU {
			t.Errorf("NPC fallthrough: Mode = %d, want MEDIT_NPC_MAIN_MENU", rig.d.Olc.Mode)
		}
	})
	t.Run("PC victim on unimplemented arm gets PC menu", func(t *testing.T) {
		rig := newPcHarness(t)
		rig.d.Olc.Mode = types.MEDIT_NAME
		meditParse(rig.d, "some input")
		out := rig.readBuf(t)
		if !strings.Contains(out, "Hps:") {
			t.Errorf("PC fallthrough expected PC menu (Hps: label); got: %q", out)
		}
		if rig.d.Olc.Mode != types.MEDIT_PC_MAIN_MENU {
			t.Errorf("PC fallthrough: Mode = %d, want MEDIT_PC_MAIN_MENU", rig.d.Olc.Mode)
		}
	})
}
