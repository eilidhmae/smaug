package game

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// ============================================================
// Wave 3 tests — G7 simple-field arms, G8 stat helper, G9
// bitmask helper + wires, plus the oedit back-wire regression.
//
// Harness reuse: newMeditHarness / newPcHarness (medit_wave2_test.go).
//
// Plan: plan-phase6-olc-medit.md §§G7 (A14, A15), §G8 (A16),
// §G9 (A17, A18, A19), §G14 back-wire (A32).
// ============================================================

// meditRigWithNpc returns an NPC-victim rig with prototype + indexdata
// so the mirror-to-prototype branches in the arms fire.
func meditRigWithNpcProto(t *testing.T) *meditTestRig {
	rig := newMeditHarness(t)
	// ensure IndexData is present + ACT_PROTOTYPE is set (prototype
	// writes activate the mirror branch per C's `xIS_SET(ACT_PROTOTYPE)`
	// gates).
	rig.victim.IndexData = &types.MobIndexData{Vnum: 1234}
	rig.victim.Act.Set(types.ACT_PROTOTYPE)
	// Sufficient trust for the ACT_FLAGS prototype-toggle gate.
	rig.d.Character.Trust = types.LEVEL_IMPLEMENTOR
	return rig
}

// --- G7 simple-field positive / negative tests (A14, A15) ---

func TestMeditArm_Sex_AssignsAndClamps(t *testing.T) {
	rig := meditRigWithNpcProto(t)
	rig.d.Olc.Mode = types.MEDIT_SEX
	meditParse(rig.d, "2")
	if rig.victim.Sex != 2 {
		t.Errorf("Sex=%d, want 2", rig.victim.Sex)
	}
	if rig.d.Olc.Mode != types.MEDIT_NPC_MAIN_MENU {
		t.Errorf("Mode=%d, want MEDIT_NPC_MAIN_MENU", rig.d.Olc.Mode)
	}
	if !rig.d.Olc.Change {
		t.Errorf("Change=false after mutation; want true")
	}
	// Prototype mirror.
	if rig.victim.IndexData.Sex != 2 {
		t.Errorf("IndexData.Sex=%d, want 2 (prototype mirror)", rig.victim.IndexData.Sex)
	}
}

func TestMeditArm_Sex_ClampsHigh(t *testing.T) {
	// Wave 3 follow-up LOW #6: C-parity silent clamp via URANGE(0,x,2)
	// (omedit.c:1715). OOB input is silently coerced into range.
	rig := meditRigWithNpcProto(t)
	rig.d.Olc.Mode = types.MEDIT_SEX
	meditParse(rig.d, "3") // C-parity: clamp to 2 (female)
	if rig.victim.Sex != 2 {
		t.Errorf("Sex=%d, want 2 (clamp high per URANGE(0,x,2))", rig.victim.Sex)
	}
	if rig.d.Olc.Mode != types.MEDIT_NPC_MAIN_MENU {
		t.Errorf("Mode=%d, want NPC_MAIN_MENU (clamp completes the arm)", rig.d.Olc.Mode)
	}
	// Prototype mirror also clamped.
	if rig.victim.IndexData.Sex != 2 {
		t.Errorf("IndexData.Sex=%d, want 2", rig.victim.IndexData.Sex)
	}
}

func TestMeditArm_Sex_ClampsLow(t *testing.T) {
	rig := newMeditHarness(t)
	rig.victim.Sex = 1
	rig.d.Olc.Mode = types.MEDIT_SEX
	meditParse(rig.d, "-5") // C-parity: clamp to 0 (neutral)
	if rig.victim.Sex != 0 {
		t.Errorf("Sex=%d, want 0 (clamp low per URANGE(0,x,2))", rig.victim.Sex)
	}
}

func TestMeditArm_Hitroll_ClampsTo85(t *testing.T) {
	rig := meditRigWithNpcProto(t)
	rig.d.Olc.Mode = types.MEDIT_HITROLL
	meditParse(rig.d, "999")
	if rig.victim.Hitroll != 85 {
		t.Errorf("Hitroll=%d, want 85 (clamp)", rig.victim.Hitroll)
	}
}

func TestMeditArm_Damroll_ClampsTo65(t *testing.T) {
	rig := meditRigWithNpcProto(t)
	rig.d.Olc.Mode = types.MEDIT_DAMROLL
	meditParse(rig.d, "-5")
	if rig.victim.Damroll != 0 {
		t.Errorf("Damroll=%d, want 0 (clamp)", rig.victim.Damroll)
	}
}

func TestMeditArm_AC_ClampsRange(t *testing.T) {
	rig := newMeditHarness(t)
	rig.d.Olc.Mode = types.MEDIT_AC
	meditParse(rig.d, "-500")
	if rig.victim.Armor != -300 {
		t.Errorf("Armor=%d, want -300 (clamp low)", rig.victim.Armor)
	}
	rig.d.Olc.Mode = types.MEDIT_AC
	meditParse(rig.d, "500")
	if rig.victim.Armor != 300 {
		t.Errorf("Armor=%d, want 300 (clamp high)", rig.victim.Armor)
	}
}

func TestMeditArm_Gold_NonNegative(t *testing.T) {
	rig := newMeditHarness(t)
	rig.d.Olc.Mode = types.MEDIT_GOLD
	meditParse(rig.d, "-100")
	if rig.victim.Gold != 0 {
		t.Errorf("Gold=%d, want 0 (clamp to >=0)", rig.victim.Gold)
	}
	rig.d.Olc.Mode = types.MEDIT_GOLD
	meditParse(rig.d, "1234567")
	if rig.victim.Gold != 1234567 {
		t.Errorf("Gold=%d, want 1234567", rig.victim.Gold)
	}
}

func TestMeditArm_Level_ClampsWithinMax(t *testing.T) {
	rig := newMeditHarness(t)
	rig.d.Olc.Mode = types.MEDIT_LEVEL
	meditParse(rig.d, "0")
	if rig.victim.Level != 1 {
		t.Errorf("Level=%d, want 1 (clamp low)", rig.victim.Level)
	}
	rig.d.Olc.Mode = types.MEDIT_LEVEL
	meditParse(rig.d, "999")
	if rig.victim.Level != types.MAX_LEVEL-1 {
		t.Errorf("Level=%d, want MAX_LEVEL-1=%d", rig.victim.Level, types.MAX_LEVEL-1)
	}
}

func TestMeditArm_Alignment_ClampsRange(t *testing.T) {
	rig := newMeditHarness(t)
	rig.d.Olc.Mode = types.MEDIT_ALIGNMENT
	meditParse(rig.d, "-5000")
	if rig.victim.Alignment != -1000 {
		t.Errorf("Alignment=%d, want -1000", rig.victim.Alignment)
	}
}

func TestMeditArm_Practice_PcOnly_Range1To300(t *testing.T) {
	// NPC rejected.
	rig := newMeditHarness(t)
	rig.d.Olc.Mode = types.MEDIT_PRACTICE
	meditParse(rig.d, "50")
	out := rig.readBuf(t)
	if !strings.Contains(out, "doesn't apply to NPCs") {
		t.Errorf("NPC practice expected reject; got: %q", out)
	}

	// PC path: clamp.
	pc := newPcHarness(t)
	pc.d.Olc.Mode = types.MEDIT_PRACTICE
	meditParse(pc.d, "500")
	if pc.victim.Practice != 300 {
		t.Errorf("PC practice=%d, want 300 (clamp)", pc.victim.Practice)
	}
	pc2 := newPcHarness(t)
	pc2.d.Olc.Mode = types.MEDIT_PRACTICE
	meditParse(pc2.d, "0")
	if pc2.victim.Practice != 1 {
		t.Errorf("PC practice=%d, want 1 (clamp low, C URANGE(1,x,300))", pc2.victim.Practice)
	}
}

func TestMeditArm_Favor_PcOnlyClamp(t *testing.T) {
	// NPC rejected.
	rig := newMeditHarness(t)
	rig.d.Olc.Mode = types.MEDIT_FAVOR
	meditParse(rig.d, "100")
	out := rig.readBuf(t)
	if !strings.Contains(out, "doesn't apply to NPCs") {
		t.Errorf("NPC favor expected reject; got: %q", out)
	}

	pc := newPcHarness(t)
	pc.d.Olc.Mode = types.MEDIT_FAVOR
	meditParse(pc.d, "3000")
	if pc.victim.PCData.Favor != 2500 {
		t.Errorf("Favor=%d, want 2500 clamp", pc.victim.PCData.Favor)
	}
}

func TestMeditArm_Hitpoint_Mana_Move_ClampPositive(t *testing.T) {
	for _, tc := range []struct {
		mode int
		get  func(v *types.CharData) int
		max  int
	}{
		{types.MEDIT_HITPOINT, func(v *types.CharData) int { return v.MaxHit }, 32700},
		{types.MEDIT_MANA, func(v *types.CharData) int { return v.MaxMana }, 30000},
		{types.MEDIT_MOVE, func(v *types.CharData) int { return v.MaxMove }, 30000},
	} {
		rig := newMeditHarness(t)
		rig.d.Olc.Mode = tc.mode
		meditParse(rig.d, "99999")
		if g := tc.get(rig.victim); g != tc.max {
			t.Errorf("mode=%d got=%d want=%d", tc.mode, g, tc.max)
		}
	}
}

func TestMeditArm_DamDice_NpcOnly(t *testing.T) {
	// PC rejected for DAMNUMDIE.
	pc := newPcHarness(t)
	pc.d.Olc.Mode = types.MEDIT_DAMNUMDIE
	meditParse(pc.d, "5")
	out := pc.readBuf(t)
	if !strings.Contains(out, "doesn't apply to players") {
		t.Errorf("PC damnumdie expected reject; got: %q", out)
	}

	// NPC + prototype writes through to IndexData.
	rig := meditRigWithNpcProto(t)
	rig.d.Olc.Mode = types.MEDIT_DAMNUMDIE
	meditParse(rig.d, "150")
	if rig.victim.IndexData.DamNoDice != 100 {
		t.Errorf("DamNoDice=%d, want 100 clamp", rig.victim.IndexData.DamNoDice)
	}
}

func TestMeditArm_Thirst_Full_Drunk_PcOnly(t *testing.T) {
	// NPC reject.
	rig := newMeditHarness(t)
	rig.d.Olc.Mode = types.MEDIT_THIRST
	meditParse(rig.d, "50")
	out := rig.readBuf(t)
	if !strings.Contains(out, "doesn't apply to NPCs") {
		t.Errorf("expected NPC-reject on THIRST, got: %q", out)
	}

	// PC applies & clamps.
	pc := newPcHarness(t)
	pc.d.Olc.Mode = types.MEDIT_THIRST
	meditParse(pc.d, "999")
	if pc.victim.PCData.Condition[types.COND_THIRST] != 100 {
		t.Errorf("thirst=%d, want 100 clamp", pc.victim.PCData.Condition[types.COND_THIRST])
	}
	pc.d.Olc.Mode = types.MEDIT_FULL
	meditParse(pc.d, "25")
	if pc.victim.PCData.Condition[types.COND_FULL] != 25 {
		t.Errorf("full=%d, want 25", pc.victim.PCData.Condition[types.COND_FULL])
	}
	pc.d.Olc.Mode = types.MEDIT_DRUNK
	meditParse(pc.d, "-10")
	if pc.victim.PCData.Condition[types.COND_DRUNK] != 0 {
		t.Errorf("drunk=%d, want 0 clamp", pc.victim.PCData.Condition[types.COND_DRUNK])
	}
}

func TestMeditArm_Position_ClampsToStanding(t *testing.T) {
	rig := newMeditHarness(t)
	rig.d.Olc.Mode = types.MEDIT_POS
	meditParse(rig.d, "99")
	if rig.victim.Position != types.POS_STANDING {
		t.Errorf("Position=%d, want POS_STANDING clamp", rig.victim.Position)
	}
}

func TestMeditArm_DefaultPos_NpcOnly(t *testing.T) {
	pc := newPcHarness(t)
	pc.d.Olc.Mode = types.MEDIT_DEFAULT_POS
	meditParse(pc.d, "5")
	out := pc.readBuf(t)
	if !strings.Contains(out, "doesn't apply to players") {
		t.Errorf("PC default-pos expected reject; got: %q", out)
	}
	rig := newMeditHarness(t)
	rig.d.Olc.Mode = types.MEDIT_DEFAULT_POS
	meditParse(rig.d, "6")
	if rig.victim.DefPosition != 6 {
		t.Errorf("DefPosition=%d, want 6", rig.victim.DefPosition)
	}
}

func TestMeditArm_Name_AssignsAndSmashTildes(t *testing.T) {
	rig := newMeditHarness(t)
	rig.d.Olc.Mode = types.MEDIT_NAME
	meditParse(rig.d, "goblin~warrior")
	if strings.Contains(rig.victim.Name, "~") {
		t.Errorf("Name=%q contains tilde; SmashTilde not applied", rig.victim.Name)
	}
	if rig.victim.Name == "" {
		t.Errorf("Name empty after set")
	}
}

func TestMeditArm_ShortDesc_NpcOnly(t *testing.T) {
	pc := newPcHarness(t)
	pc.d.Olc.Mode = types.MEDIT_S_DESC
	meditParse(pc.d, "a short description")
	out := pc.readBuf(t)
	if !strings.Contains(out, "doesn't apply to players") {
		t.Errorf("PC shortdesc expected reject; got: %q", out)
	}

	rig := meditRigWithNpcProto(t)
	rig.d.Olc.Mode = types.MEDIT_S_DESC
	meditParse(rig.d, "a grumpy shopkeeper")
	if rig.victim.ShortDescr != "a grumpy shopkeeper" {
		t.Errorf("ShortDescr=%q, want 'a grumpy shopkeeper'", rig.victim.ShortDescr)
	}
	// Prototype mirror.
	if rig.victim.IndexData.ShortDescr != "a grumpy shopkeeper" {
		t.Errorf("IndexData.ShortDescr not mirrored")
	}
}

func TestMeditArm_LongDesc_AppendsCrLf(t *testing.T) {
	rig := meditRigWithNpcProto(t)
	rig.d.Olc.Mode = types.MEDIT_L_DESC
	meditParse(rig.d, "A grumpy shopkeeper watches you")
	if !strings.HasSuffix(rig.victim.LongDescr, "\n\r") {
		t.Errorf("LongDescr=%q does not end with \\n\\r", rig.victim.LongDescr)
	}
}

func TestMeditArm_DDescBugTrap_CleansUpAndLogs(t *testing.T) {
	// C omedit.c:1431-1435 bug-trap. Direct input on MEDIT_D_DESC
	// should clean up the session (and log via util.Bug).
	rig := newMeditHarness(t)
	rig.d.Olc.Mode = types.MEDIT_D_DESC
	meditParse(rig.d, "anything")
	if rig.d.Olc != nil {
		t.Errorf("Olc not cleared after D_DESC bug-trap; got=%+v", rig.d.Olc)
	}
	if rig.d.Connected != int(types.CON_PLAYING) {
		t.Errorf("Connected=%d, want CON_PLAYING after bug-trap cleanup", rig.d.Connected)
	}
}

func TestMeditArm_Spec_NpcOnly_AssignsByIndex(t *testing.T) {
	pc := newPcHarness(t)
	pc.d.Olc.Mode = types.MEDIT_SPEC
	meditParse(pc.d, "1")
	out := pc.readBuf(t)
	if !strings.Contains(out, "doesn't apply to players") {
		t.Errorf("PC spec expected reject; got: %q", out)
	}

	rig := meditRigWithNpcProto(t)
	rig.d.Olc.Mode = types.MEDIT_SPEC
	meditParse(rig.d, "3") // index 3 = spec_breath_fire
	if rig.victim.SpecFun != "spec_breath_fire" {
		t.Errorf("SpecFun=%q, want spec_breath_fire", rig.victim.SpecFun)
	}
	if rig.victim.IndexData.SpecFun != "spec_breath_fire" {
		t.Errorf("IndexData.SpecFun not mirrored")
	}
}

func TestMeditArm_ClanDeityCouncil_PcOnlyEmptyClears(t *testing.T) {
	pc := newPcHarness(t)
	pc.victim.PCData.DeityName = "zeus"
	pc.d.Olc.Mode = types.MEDIT_DEITY
	meditParse(pc.d, "")
	if pc.victim.PCData.DeityName != "" {
		t.Errorf("DeityName=%q, want empty after blank clear", pc.victim.PCData.DeityName)
	}
	// NPC reject.
	rig := newMeditHarness(t)
	rig.d.Olc.Mode = types.MEDIT_CLAN
	meditParse(rig.d, "dragons")
	out := rig.readBuf(t)
	if !strings.Contains(out, "doesn't apply to NPCs") {
		t.Errorf("NPC CLAN expected reject; got: %q", out)
	}
}

// --- G8 stat helper tests (A16) ---

func TestMeditArm_Stat_NpcClamp_1To25(t *testing.T) {
	rig := meditRigWithNpcProto(t)
	rig.d.Olc.Mode = types.MEDIT_STRENGTH
	meditParse(rig.d, "0") // below NPC min=1
	if rig.victim.PermStr != 1 {
		t.Errorf("NPC PermStr=%d, want 1 clamp", rig.victim.PermStr)
	}
	rig.d.Olc.Mode = types.MEDIT_STRENGTH
	meditParse(rig.d, "99") // above NPC max=25
	if rig.victim.PermStr != 25 {
		t.Errorf("NPC PermStr=%d, want 25 clamp", rig.victim.PermStr)
	}
}

func TestMeditArm_Stat_PcClamp_3To18(t *testing.T) {
	// PC uses uRange(3, 18) per C omedit.c:998-1007. A mutation that
	// flips this to (0,100) or to NPC's (1,25) fails here — pins M7.
	pc := newPcHarness(t)
	pc.d.Olc.Mode = types.MEDIT_STRENGTH
	meditParse(pc.d, "25")
	if pc.victim.PermStr != 18 {
		t.Errorf("PC PermStr=%d, want 18 clamp (C URANGE(3,x,18))", pc.victim.PermStr)
	}
	pc.d.Olc.Mode = types.MEDIT_STRENGTH
	meditParse(pc.d, "1")
	if pc.victim.PermStr != 3 {
		t.Errorf("PC PermStr=%d, want 3 clamp low", pc.victim.PermStr)
	}
}

func TestMeditArm_AllSevenStats_Assign(t *testing.T) {
	// Exercise each of the 7 arms end-to-end.
	cases := []struct {
		mode int
		get  func(v *types.CharData) int
		name string
	}{
		{types.MEDIT_STRENGTH, func(v *types.CharData) int { return v.PermStr }, "str"},
		{types.MEDIT_INTELLIGENCE, func(v *types.CharData) int { return v.PermInt }, "int"},
		{types.MEDIT_WISDOM, func(v *types.CharData) int { return v.PermWis }, "wis"},
		{types.MEDIT_DEXTERITY, func(v *types.CharData) int { return v.PermDex }, "dex"},
		{types.MEDIT_CONSTITUTION, func(v *types.CharData) int { return v.PermCon }, "con"},
		{types.MEDIT_CHARISMA, func(v *types.CharData) int { return v.PermCha }, "cha"},
		{types.MEDIT_LUCK, func(v *types.CharData) int { return v.PermLck }, "lck"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rig := meditRigWithNpcProto(t)
			rig.d.Olc.Mode = tc.mode
			meditParse(rig.d, "15")
			if g := tc.get(rig.victim); g != 15 {
				t.Errorf("NPC %s=%d, want 15", tc.name, g)
			}
		})
	}
}

// --- G9 bitmask helper + wire tests (A17, A18, A19) ---

func TestOlcBitmaskEdit_DoneQuitEmpty_ReturnsTrue(t *testing.T) {
	// A19 — exit semantics. A mutation that flips "done" to "d" (prefix
	// match) fails TestOlcBitmaskEdit_PartialWordRejected below.
	rig := newMeditHarness(t)
	var cur int
	for _, in := range []string{"", "done", "DONE", "quit", "QUIT"} {
		if done := olcBitmaskEdit(rig.d, "PART", &cur, in); !done {
			t.Errorf("olcBitmaskEdit(%q) = false, want true (exit)", in)
		}
	}
}

func TestOlcBitmaskEdit_PartialWordRejected(t *testing.T) {
	// M9 pin: "d" must NOT match "done" (prefix match would be a bug).
	rig := newMeditHarness(t)
	var cur int
	done := olcBitmaskEdit(rig.d, "PART", &cur, "d")
	if done {
		t.Errorf("olcBitmaskEdit(\"d\") returned true — prefix match incorrectly accepted")
	}
	// Also verify "d" was treated as invalid (rejected with message).
	out := rig.readBuf(t)
	if !strings.Contains(out, "Invalid flag") {
		t.Errorf("expected 'Invalid flag' for unknown keyword 'd'; got %q", out)
	}
}

func TestOlcBitmaskEdit_NumericIndex_TogglesCorrectBit(t *testing.T) {
	rig := newMeditHarness(t)
	var cur int
	// PART table, input "1" → toggle bit 0 (C offset: number - 1).
	olcBitmaskEdit(rig.d, "PART", &cur, "1")
	if cur != 1 {
		t.Errorf("cur=%d after toggle bit 0, want 1", cur)
	}
	// Toggle again -> clear.
	olcBitmaskEdit(rig.d, "PART", &cur, "1")
	if cur != 0 {
		t.Errorf("cur=%d after second toggle, want 0", cur)
	}
}

func TestOlcBitmaskEdit_Numeric0_IsExit(t *testing.T) {
	rig := newMeditHarness(t)
	var cur int = 0xFF
	done := olcBitmaskEdit(rig.d, "PART", &cur, "0")
	if !done {
		t.Errorf("numeric 0 should be treated as exit (C parity)")
	}
	if cur != 0xFF {
		t.Errorf("cur mutated by exit: got %d want 255", cur)
	}
}

func TestOlcBitmaskEdit_OutOfRangeRejected(t *testing.T) {
	rig := newMeditHarness(t)
	var cur int
	done := olcBitmaskEdit(rig.d, "PART", &cur, "999")
	if done {
		t.Errorf("OOB numeric should stay in mode (done=false)")
	}
	if cur != 0 {
		t.Errorf("OOB should not mutate; cur=%d", cur)
	}
	out := rig.readBuf(t)
	if !strings.Contains(out, "Invalid flag") {
		t.Errorf("expected 'Invalid flag' message; got %q", out)
	}
}

func TestOlcBitmaskEdit_KeywordCaseInsensitive(t *testing.T) {
	rig := newMeditHarness(t)
	var cur int
	// RIS table has "fire" at index 0.
	olcBitmaskEdit(rig.d, "RIS", &cur, "FIRE")
	if cur != 1 {
		t.Errorf("cur=%d after 'FIRE' keyword, want 1 (bit 0)", cur)
	}
	olcBitmaskEdit(rig.d, "RIS", &cur, "Fire")
	if cur != 0 {
		t.Errorf("second 'Fire' should clear bit 0; got %d", cur)
	}
}

func TestOlcBitmaskEdit_TokenStream(t *testing.T) {
	rig := newMeditHarness(t)
	var cur int
	olcBitmaskEdit(rig.d, "RIS", &cur, "fire cold electricity")
	want := (1 << 0) | (1 << 1) | (1 << 2) // fire=0, cold=1, electricity=2
	if cur != want {
		t.Errorf("cur=0x%x, want 0x%x (3 flags toggled)", cur, want)
	}
}

func TestOlcBitmaskEdit_UnknownTableIsExit(t *testing.T) {
	rig := newMeditHarness(t)
	var cur int
	if done := olcBitmaskEdit(rig.d, "NONEXISTENT_TABLE", &cur, "fire"); !done {
		t.Errorf("unknown table should return true (exit)")
	}
}

// --- G9 wires — six arms toggle correct bits (A18) ---

func TestMeditArm_NpcFlags_TogglesBit(t *testing.T) {
	rig := meditRigWithNpcProto(t)
	// ACT_AGGRESSIVE = 5, 1-based index = 6.
	rig.d.Olc.Mode = types.MEDIT_NPC_FLAGS
	meditParse(rig.d, "6")
	if !rig.victim.Act.IsSet(types.ACT_AGGRESSIVE) {
		t.Errorf("ACT_AGGRESSIVE not set after toggle via digit 6")
	}
	// IndexData.Act mirrored.
	if !rig.victim.IndexData.Act.IsSet(types.ACT_AGGRESSIVE) {
		t.Errorf("IndexData.Act not mirrored")
	}
}

func TestMeditArm_NpcFlags_RejectsPcVictim(t *testing.T) {
	pc := newPcHarness(t)
	pc.d.Olc.Mode = types.MEDIT_NPC_FLAGS
	meditParse(pc.d, "aggressive")
	out := pc.readBuf(t)
	if !strings.Contains(out, "doesn't apply to players") {
		t.Errorf("PC NPC_FLAGS expected reject; got: %q", out)
	}
}

func TestMeditArm_NpcFlags_ProtectsIsNpc(t *testing.T) {
	// ACT_IS_NPC = 0, 1-based index = "1". Must be refused.
	// Wave 3 follow-up HIGH #2: emits the verbatim C string from
	// omedit.c:1473 ("It isn't possible to change that flag.\n\r").
	rig := meditRigWithNpcProto(t)
	rig.d.Olc.Mode = types.MEDIT_NPC_FLAGS
	meditParse(rig.d, "1")
	out := rig.readBuf(t)
	if !strings.Contains(out, "It isn't possible to change that flag.") {
		t.Errorf("expected C-verbatim protection string; got %q", out)
	}
	// ACT_IS_NPC bit must remain set.
	if !rig.victim.Act.IsSet(types.ACT_IS_NPC) {
		t.Errorf("ACT_IS_NPC cleared by numeric '1' — protection failed")
	}
}

func TestMeditArm_PcFlags_TogglesBit(t *testing.T) {
	pc := newPcHarness(t)
	pc.d.Olc.Mode = types.MEDIT_PC_FLAGS
	// PLR_ANSI = 25 (per util.PlrflagNames), 1-based = "26".
	meditParse(pc.d, "26")
	if !pc.victim.Act.IsSet(25) {
		t.Errorf("bit 25 not toggled via digit 26 on PC_FLAGS")
	}
}

func TestMeditArm_PcFlags_NpcRejected(t *testing.T) {
	rig := newMeditHarness(t)
	rig.d.Olc.Mode = types.MEDIT_PC_FLAGS
	meditParse(rig.d, "1")
	out := rig.readBuf(t)
	if !strings.Contains(out, "doesn't apply to NPCs") {
		t.Errorf("NPC PC_FLAGS expected reject; got: %q", out)
	}
}

func TestMeditArm_AffFlags_TogglesBit(t *testing.T) {
	rig := meditRigWithNpcProto(t)
	rig.d.Olc.Mode = types.MEDIT_AFF_FLAGS
	// AFF_INVISIBLE = 1, 1-based = "2".
	meditParse(rig.d, "2")
	if !rig.victim.AffectedBy.IsSet(types.AFF_INVISIBLE) {
		t.Errorf("AFF_INVISIBLE not set")
	}
	if !rig.victim.IndexData.AffectedBy.IsSet(types.AFF_INVISIBLE) {
		t.Errorf("IndexData.AffectedBy not mirrored")
	}
}

func TestMeditArm_PcdataFlags_TogglesBit(t *testing.T) {
	pc := newPcHarness(t)
	pc.d.Olc.Mode = types.MEDIT_PCDATA_FLAGS
	meditParse(pc.d, "2") // PCFLAG_DEADLY at index 1 (1-based = "2")
	if pc.victim.PCData.Flags&(1<<1) == 0 {
		t.Errorf("PCData.Flags bit 1 not set; got=0x%x", pc.victim.PCData.Flags)
	}
}

func TestMeditArm_PcdataFlags_NpcRejected(t *testing.T) {
	rig := newMeditHarness(t)
	rig.d.Olc.Mode = types.MEDIT_PCDATA_FLAGS
	meditParse(rig.d, "1")
	out := rig.readBuf(t)
	if !strings.Contains(out, "doesn't apply to NPCs") {
		t.Errorf("NPC PCDATA_FLAGS expected reject; got %q", out)
	}
}

func TestMeditArm_Parts_TogglesBit(t *testing.T) {
	rig := meditRigWithNpcProto(t)
	rig.d.Olc.Mode = types.MEDIT_PARTS
	// "head" is index 0, 1-based = "1".
	meditParse(rig.d, "1")
	if rig.victim.XFlags&(1<<0) == 0 {
		t.Errorf("XFlags bit 0 (head) not set")
	}
	if rig.victim.IndexData.XFlags&(1<<0) == 0 {
		t.Errorf("IndexData.XFlags not mirrored")
	}
}

func TestMeditArm_Parts_PcRejected(t *testing.T) {
	pc := newPcHarness(t)
	pc.d.Olc.Mode = types.MEDIT_PARTS
	meditParse(pc.d, "head")
	out := pc.readBuf(t)
	if !strings.Contains(out, "doesn't apply to players") {
		t.Errorf("PC PARTS expected reject; got %q", out)
	}
}

func TestMeditArm_Ris_ThreeFields_OneTable(t *testing.T) {
	// RESISTANT, IMMUNE, SUSCEPTIBLE share util.RisflagNames. Toggle
	// "fire" in each and verify only the targeted field mutates.
	for _, tc := range []struct {
		mode int
		get  func(v *types.CharData) int
		name string
	}{
		{types.MEDIT_RESISTANT, func(v *types.CharData) int { return v.Resistant }, "resistant"},
		{types.MEDIT_IMMUNE, func(v *types.CharData) int { return v.Immune }, "immune"},
		{types.MEDIT_SUSCEPTIBLE, func(v *types.CharData) int { return v.Susceptible }, "suscept"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rig := meditRigWithNpcProto(t)
			rig.d.Olc.Mode = tc.mode
			meditParse(rig.d, "fire")
			if tc.get(rig.victim)&1 == 0 {
				t.Errorf("%s: bit 0 (fire) not set", tc.name)
			}
			// Mirror verification.
			switch tc.mode {
			case types.MEDIT_RESISTANT:
				if rig.victim.IndexData.Resistant&1 == 0 {
					t.Errorf("IndexData.Resistant not mirrored")
				}
			case types.MEDIT_IMMUNE:
				if rig.victim.IndexData.Immune&1 == 0 {
					t.Errorf("IndexData.Immune not mirrored")
				}
			case types.MEDIT_SUSCEPTIBLE:
				if rig.victim.IndexData.Susceptible&1 == 0 {
					t.Errorf("IndexData.Susceptible not mirrored")
				}
			}
		})
	}
}

// Flag name tables must exist and be non-empty (A18 prerequisite).
func TestFlagNameTables_AllPresent(t *testing.T) {
	tables := []struct {
		name string
		need []string
	}{
		{"ACT_FLAGS", []string{"npc", "aggressive", "prototype"}},
		{"PLR_FLAGS", []string{"wizinvis", "autogold", "afk"}},
		{"AFF_FLAGS", []string{"blind", "invisible", "sanctuary"}},
		{"PCFLAG", []string{"deadly", "unauthed", "beckon"}},
		{"PART", []string{"head", "arms", "legs"}},
		{"RIS", []string{"fire", "cold", "poison"}},
	}
	for _, tc := range tables {
		tbl := lookupBitmaskTable(tc.name)
		if tbl == nil {
			t.Errorf("lookupBitmaskTable(%q) = nil; expected populated table", tc.name)
			continue
		}
		for _, want := range tc.need {
			found := false
			for _, n := range tbl {
				if n == want {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("table %q missing required flag %q", tc.name, want)
			}
		}
	}
}

// --- A19 done/quit/empty consistent across wires ---

func TestMeditArms_DoneReturnsToMainMenu(t *testing.T) {
	// Verify each bitmask wire exits cleanly on "done" and lands back on
	// the correct main menu (NPC for NPC victim, PC for PC victim).
	npcModes := []int{
		types.MEDIT_NPC_FLAGS, types.MEDIT_PARTS,
		types.MEDIT_RESISTANT, types.MEDIT_IMMUNE, types.MEDIT_SUSCEPTIBLE,
	}
	for _, m := range npcModes {
		rig := meditRigWithNpcProto(t)
		rig.d.Olc.Mode = m
		meditParse(rig.d, "done")
		if rig.d.Olc == nil {
			t.Errorf("mode=%d: Olc cleared on 'done' (should stay in session)", m)
			continue
		}
		if rig.d.Olc.Mode != types.MEDIT_NPC_MAIN_MENU {
			t.Errorf("mode=%d: after 'done' Mode=%d, want NPC_MAIN_MENU", m, rig.d.Olc.Mode)
		}
	}

	pcModes := []int{
		types.MEDIT_PC_FLAGS, types.MEDIT_PCDATA_FLAGS,
		types.MEDIT_RESISTANT, types.MEDIT_IMMUNE, types.MEDIT_SUSCEPTIBLE,
	}
	for _, m := range pcModes {
		rig := newPcHarness(t)
		rig.d.Olc.Mode = m
		meditParse(rig.d, "done")
		if rig.d.Olc == nil {
			t.Errorf("pc mode=%d: Olc cleared", m)
			continue
		}
		if rig.d.Olc.Mode != types.MEDIT_PC_MAIN_MENU {
			t.Errorf("pc mode=%d: after 'done' Mode=%d, want PC_MAIN_MENU", m, rig.d.Olc.Mode)
		}
	}

	// Shared AFF_FLAGS stays in each victim's main menu.
	rigAffNpc := meditRigWithNpcProto(t)
	rigAffNpc.d.Olc.Mode = types.MEDIT_AFF_FLAGS
	meditParse(rigAffNpc.d, "done")
	if rigAffNpc.d.Olc.Mode != types.MEDIT_NPC_MAIN_MENU {
		t.Errorf("AFF_FLAGS on NPC done: Mode=%d, want NPC_MAIN_MENU", rigAffNpc.d.Olc.Mode)
	}
	rigAffPc := newPcHarness(t)
	rigAffPc.d.Olc.Mode = types.MEDIT_AFF_FLAGS
	meditParse(rigAffPc.d, "done")
	if rigAffPc.d.Olc.Mode != types.MEDIT_PC_MAIN_MENU {
		t.Errorf("AFF_FLAGS on PC done: Mode=%d, want PC_MAIN_MENU", rigAffPc.d.Olc.Mode)
	}
}

// --- A32 oedit back-wire regression ---

// TestOeditParse_AffectModifier_UsesBitmaskEditor pins the Wave-3
// back-wire: oedit OEDIT_AFFECT_MODIFIER with APPLY_AFFECT now
// dispatches through olcBitmaskEdit so bit-toggles land in
// paf.Modifier. A mutation that reverts the arm to numeric-only fails
// this test (M10).
func TestOeditParse_AffectModifier_UsesBitmaskEditor(t *testing.T) {
	// Stand up a minimal oedit rig inline: we don't need the full
	// oedit harness since we only exercise oeditHandleAffectModifier
	// directly.
	rig := newMeditHarness(t)
	// Replace medit's target with an ObjIndexData for this test.
	idx := &types.ObjIndexData{Vnum: 9999}
	rig.d.Olc.Target = idx
	// Stage a pending affect with APPLY_AFFECT location.
	paf := &types.AffectData{Location: types.APPLY_AFFECT}
	rig.d.Olc.Spare = paf

	// Back-wire dispatch via oeditHandleAffectModifier.
	// "invisible" (AFF_INVISIBLE=1) → toggle bit 1.
	oeditHandleAffectModifier(rig.d, idx, "invisible")

	// Committed only after "done" — still pending.
	if len(idx.Affects) != 0 {
		t.Errorf("Affect committed before 'done'; len(Affects)=%d", len(idx.Affects))
	}
	if paf.Modifier&(1<<types.AFF_INVISIBLE) == 0 {
		t.Errorf("paf.Modifier bit AFF_INVISIBLE not set after bitmask toggle; got 0x%x", paf.Modifier)
	}

	// Now "done" commits.
	oeditHandleAffectModifier(rig.d, idx, "done")
	if len(idx.Affects) != 1 {
		t.Errorf("Affect not committed on 'done'; len(Affects)=%d", len(idx.Affects))
	}
	if idx.Affects[0].Modifier&(1<<types.AFF_INVISIBLE) == 0 {
		t.Errorf("committed affect modifier missing AFF_INVISIBLE bit")
	}
}

func TestOeditParse_AffectModifier_ScalarPathStillNumeric(t *testing.T) {
	// Regression: APPLY_STR (scalar) still takes numeric modifier.
	rig := newMeditHarness(t)
	idx := &types.ObjIndexData{Vnum: 9999}
	rig.d.Olc.Target = idx
	paf := &types.AffectData{Location: types.APPLY_STR}
	rig.d.Olc.Spare = paf
	oeditHandleAffectModifier(rig.d, idx, "3")
	if len(idx.Affects) != 1 {
		t.Fatalf("scalar affect not committed immediately; len=%d", len(idx.Affects))
	}
	if idx.Affects[0].Modifier != 3 {
		t.Errorf("scalar modifier=%d, want 3", idx.Affects[0].Modifier)
	}
}

func TestOeditParse_AffectModifier_BitmaskCancelDropsAffect(t *testing.T) {
	rig := newMeditHarness(t)
	idx := &types.ObjIndexData{Vnum: 9999}
	rig.d.Olc.Target = idx
	paf := &types.AffectData{Location: types.APPLY_RESISTANT}
	rig.d.Olc.Spare = paf
	oeditHandleAffectModifier(rig.d, idx, "0") // cancel
	if len(idx.Affects) != 0 {
		t.Errorf("cancel with '0' should not commit; len=%d", len(idx.Affects))
	}
	if rig.d.Olc.Spare != nil {
		t.Errorf("cancel should clear Olc.Spare; got %+v", rig.d.Olc.Spare)
	}
}

// ============================================================
// Wave 3 follow-up tests (2026-04-22) — adversary findings.
// ============================================================

// HIGH #1 — meditArmStat must mirror to MobIndexData.Perm* on
// NPC+ACT_PROTOTYPE. Pre-fix: comment claimed the prototype fields did
// not exist; in fact MobIndexData has PermStr..PermLck (mob_index.go:67-73).
// Mutation gate: dropping `*mirror = *field` in meditArmStat fails this.
func TestMeditArm_Stat_MirrorsToPrototype(t *testing.T) {
	cases := []struct {
		mode   int
		victim func(v *types.CharData) int
		index  func(idx *types.MobIndexData) int
		name   string
	}{
		{types.MEDIT_STRENGTH, func(v *types.CharData) int { return v.PermStr }, func(i *types.MobIndexData) int { return i.PermStr }, "str"},
		{types.MEDIT_INTELLIGENCE, func(v *types.CharData) int { return v.PermInt }, func(i *types.MobIndexData) int { return i.PermInt }, "int"},
		{types.MEDIT_WISDOM, func(v *types.CharData) int { return v.PermWis }, func(i *types.MobIndexData) int { return i.PermWis }, "wis"},
		{types.MEDIT_DEXTERITY, func(v *types.CharData) int { return v.PermDex }, func(i *types.MobIndexData) int { return i.PermDex }, "dex"},
		{types.MEDIT_CONSTITUTION, func(v *types.CharData) int { return v.PermCon }, func(i *types.MobIndexData) int { return i.PermCon }, "con"},
		{types.MEDIT_CHARISMA, func(v *types.CharData) int { return v.PermCha }, func(i *types.MobIndexData) int { return i.PermCha }, "cha"},
		{types.MEDIT_LUCK, func(v *types.CharData) int { return v.PermLck }, func(i *types.MobIndexData) int { return i.PermLck }, "lck"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rig := meditRigWithNpcProto(t)
			rig.d.Olc.Mode = tc.mode
			meditParse(rig.d, "17")
			if got := tc.victim(rig.victim); got != 17 {
				t.Errorf("victim %s=%d, want 17", tc.name, got)
			}
			if got := tc.index(rig.victim.IndexData); got != 17 {
				t.Errorf("IndexData.Perm%s=%d, want 17 (prototype mirror missing)", tc.name, got)
			}
		})
	}
}

// HIGH #1 — non-prototype NPC must NOT mirror (gates on ACT_PROTOTYPE).
func TestMeditArm_Stat_NoMirrorWithoutPrototype(t *testing.T) {
	rig := newMeditHarness(t) // NPC but no ACT_PROTOTYPE
	rig.victim.IndexData = &types.MobIndexData{Vnum: 1234, PermStr: 5}
	rig.d.Olc.Mode = types.MEDIT_STRENGTH
	meditParse(rig.d, "12")
	if rig.victim.PermStr != 12 {
		t.Errorf("victim.PermStr=%d, want 12", rig.victim.PermStr)
	}
	if rig.victim.IndexData.PermStr != 5 {
		t.Errorf("IndexData.PermStr=%d, want 5 (mirror should NOT fire without ACT_PROTOTYPE)", rig.victim.IndexData.PermStr)
	}
}

// HIGH #1 — PC stat editing must not crash when IndexData is nil and
// must not write through to a non-existent prototype.
func TestMeditArm_Stat_PcNoIndexDataIsSafe(t *testing.T) {
	pc := newPcHarness(t) // PC, no IndexData
	pc.d.Olc.Mode = types.MEDIT_STRENGTH
	meditParse(pc.d, "10")
	if pc.victim.PermStr != 10 {
		t.Errorf("PC PermStr=%d, want 10", pc.victim.PermStr)
	}
}

// HIGH #2 — keyword form "npc" must hit the ACT_IS_NPC guard.
// Pre-fix: protectedToggleRejected only caught numeric input; "npc"
// passed through to olcBitmaskEditBitVector which toggled bit 0 off,
// corrupting the mob. Mutation gate: dropping the post-toggle restore
// in meditArmNpcFlags fails this.
func TestMeditArm_NpcFlags_ProtectsIsNpc_KeywordForm(t *testing.T) {
	rig := meditRigWithNpcProto(t)
	rig.d.Olc.Mode = types.MEDIT_NPC_FLAGS
	meditParse(rig.d, "npc")
	out := rig.readBuf(t)
	if !strings.Contains(out, "It isn't possible to change that flag.") {
		t.Errorf("expected verbatim C protection string; got %q", out)
	}
	if !rig.victim.Act.IsSet(types.ACT_IS_NPC) {
		t.Errorf("ACT_IS_NPC cleared by keyword 'npc' — keyword-form protection failed")
	}
}

// HIGH #2 — case-insensitive keyword form should also be caught.
func TestMeditArm_NpcFlags_ProtectsIsNpc_KeywordCaseInsensitive(t *testing.T) {
	rig := meditRigWithNpcProto(t)
	rig.d.Olc.Mode = types.MEDIT_NPC_FLAGS
	meditParse(rig.d, "NPC")
	if !rig.victim.Act.IsSet(types.ACT_IS_NPC) {
		t.Errorf("ACT_IS_NPC cleared by uppercase 'NPC' — protection must be case-insensitive")
	}
}

// MEDIUM #3 — keyword form "prototype" with insufficient trust must be
// rejected. C omedit.c:1467-1471 applies the trust gate after keyword
// resolution; pre-fix Go only caught numeric form.
func TestMeditArm_NpcFlags_ProtectsPrototype_KeywordForm(t *testing.T) {
	rig := meditRigWithNpcProto(t)
	rig.d.Character.Trust = types.LEVEL_IMMORTAL // < LEVEL_GREATER
	if !rig.victim.Act.IsSet(types.ACT_PROTOTYPE) {
		t.Fatalf("rig precondition: ACT_PROTOTYPE must start set")
	}
	rig.d.Olc.Mode = types.MEDIT_NPC_FLAGS
	meditParse(rig.d, "prototype")
	out := rig.readBuf(t)
	if !strings.Contains(out, "You don't have permission to change the prototype flag.") {
		t.Errorf("expected trust-rejection string; got %q", out)
	}
	// Bit must remain set (trust-low builder cannot clear it).
	if !rig.victim.Act.IsSet(types.ACT_PROTOTYPE) {
		t.Errorf("ACT_PROTOTYPE cleared by keyword 'prototype' at trust < LEVEL_GREATER")
	}
}

// MEDIUM #3 — sufficient trust still allows the keyword toggle.
func TestMeditArm_NpcFlags_PrototypeKeyword_HighTrustToggles(t *testing.T) {
	rig := meditRigWithNpcProto(t)
	// Wave 3 harness defaults trust to LEVEL_IMPLEMENTOR, well above LEVEL_GREATER.
	rig.d.Olc.Mode = types.MEDIT_NPC_FLAGS
	meditParse(rig.d, "prototype")
	if rig.victim.Act.IsSet(types.ACT_PROTOTYPE) {
		t.Errorf("ACT_PROTOTYPE should have toggled OFF at high trust; still set")
	}
}

// LOW #5 — flag-name tables must match C build.c entry counts so
// rendering and lookup stay bit-for-bit aligned with the C codebase.
// Pre-fix: AffflagNames had 44 (missing trailing r8/r9/r10), PcflagNames
// had 33 (missing trailing r1..r10). Mutation gate: dropping a trailing
// reserved entry from either table fails this.
func TestFlagNameTables_MatchCConstC(t *testing.T) {
	// C build.c sizes (see src/build.c:216-283 — #ifdef-gated entries
	// counted as enabled per the Go-side comment in flags.go).
	// Note: act_flags / plr_flags are gated by #ifdef macros that the
	// Go port enables, so the totals include all conditional entries.
	cases := []struct {
		name  string
		table []string
		want  int
	}{
		{"AffflagNames", util.AffflagNames, 47},
		{"PcflagNames", util.PcflagNames, 43},
	}
	for _, tc := range cases {
		if got := len(tc.table); got != tc.want {
			t.Errorf("%s len=%d, want %d (C parity per build.c)", tc.name, got, tc.want)
		}
	}
}
