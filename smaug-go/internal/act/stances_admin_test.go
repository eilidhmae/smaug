package act

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// G4 — DoSTstat + DoSTset. Covers A12-A15 from plan-phase6-stances-olc.md.

// saveStanceIndex clones StanceIndex and registers a cleanup restorer
// so subtests can mutate the table without leaking state.
func saveStanceIndex(t *testing.T) {
	t.Helper()
	saved := combat.StanceIndex
	t.Cleanup(func() { combat.StanceIndex = saved })
}

func newAdminChar(t *testing.T) (*types.CharData, net.Conn) {
	t.Helper()
	ch, client := makeTestChar("Admin")
	ch.Level = types.LEVEL_IMMORTAL
	ch.Trust = types.LEVEL_IMMORTAL
	return ch, client
}

func ensureSystemDirForTest(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "system"), 0o755); err != nil {
		t.Fatal(err)
	}
}

// --- DoSTstat ---

func TestDoSTstat_NoArg_ListsAllStances(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	DoSTstat(ch, "")
	out := readOutput(ch, client)
	wantNames := []string{"None", "Normal", "Viper", "Crane", "Crab", "Mongoose",
		"Bull", "Mantis", "Dragon", "Tiger", "Monkey", "Swallow"}
	for _, n := range wantNames {
		if !strings.Contains(out, n) {
			t.Errorf("DoSTstat no-arg output missing %q; output=\n%s", n, out)
		}
	}
}

func TestDoSTstat_UnknownName_ShowsErrorAndHelp(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	DoSTstat(ch, "nonesuch")
	out := readOutput(ch, client)
	if !strings.Contains(out, "That stance does not exist") {
		t.Errorf("expected error message; got:\n%s", out)
	}
}

func TestDoSTstat_KnownStance_PrintsAllFields(t *testing.T) {
	saveStanceIndex(t)
	combat.StanceIndex[types.STANCE_DRAGON] = combat.StanceInfo{
		NumAttacks: 2, DamDone: 150, DamTaken: 90,
		Dodge: 10, Parry: 5, Dual: 1,
		MaxWeight: 500, Wait: 12,
		Resist: int(types.RIS_FIRE), Immune: int(types.RIS_COLD), Suscept: int(types.RIS_ACID),
		SpecialPercent: 25, Class: 0x7, Race: 0x3,
		Prereq: [2]int{types.STANCE_TIGER, types.STANCE_SWALLOW},
		Self:   "you self", Others: "the room",
	}
	ch, client := newAdminChar(t)
	defer client.Close()
	DoSTstat(ch, "Dragon")
	out := readOutput(ch, client)

	wantSubstrings := []string{
		"Name            : Dragon",
		"Self            : you self",
		"Other           : the room",
		"Required Stances: Tiger and Swallow",
		"Dual Wield      : 1",
		"Damage Done     : 150",
		"Damage Taken   : 90",
		"Dodge           : 10",
		"Parry          : 5",
		"Attacks         : 2",
		"Max Weight     : 500",
		"Wait            : 12",
		"Resistance      : fire",
		"Susceptible     : acid",
		"Immune          : cold",
		"Special Move    : None",
		"Chance         : 25",
		"Class Restrictions: 0x7",
		"Race Restrictions : 0x3",
	}
	for _, s := range wantSubstrings {
		if !strings.Contains(out, s) {
			t.Errorf("DoSTstat output missing %q; output=\n%s", s, out)
		}
	}
}

// --- DoSTset ---

func TestDoSTset_NPCRejected(t *testing.T) {
	saveStanceIndex(t)
	server, client := net.Pipe()
	defer client.Close()
	npc := &types.CharData{}
	npc.Act.Set(types.ACT_IS_NPC)
	d := &types.DescriptorData{Conn: server, Connected: types.CON_PLAYING}
	d.Character = npc
	npc.Desc = d
	DoSTset(npc, "dragon attacks 2")
	out := readOutput(npc, client)
	if !strings.Contains(out, "Mob's can't mset") {
		t.Errorf("NPC caller should see typo'd rejection; got:\n%s", out)
	}
}

func TestDoSTset_NoDesc_Rejected(t *testing.T) {
	saveStanceIndex(t)
	ch := &types.CharData{Level: types.LEVEL_IMMORTAL}
	before := combat.StanceIndex[types.STANCE_DRAGON].NumAttacks
	DoSTset(ch, "dragon attacks 5")
	if combat.StanceIndex[types.STANCE_DRAGON].NumAttacks != before {
		t.Errorf("no-desc should early-return without mutation")
	}
}

func TestDoSTset_NoArgs_ShowsUsage(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	DoSTset(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Syntax: stset") {
		t.Errorf("no-args should print usage; got:\n%s", out)
	}
}

func TestDoSTset_QuestionMark_ShowsUsage(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	DoSTset(ch, "? ?")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Syntax: stset") {
		t.Errorf("'?' first-arg should print usage; got:\n%s", out)
	}
}

func TestDoSTset_UnknownStance_ShowsError(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	DoSTset(ch, "bogus attacks 3")
	out := readOutput(ch, client)
	if !strings.Contains(out, "No such stance name") {
		t.Errorf("unknown stance should error; got:\n%s", out)
	}
}

func TestDoSTset_Attacks_InRange(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	DoSTset(ch, "dragon attacks 2")
	if combat.StanceIndex[types.STANCE_DRAGON].NumAttacks != 2 {
		t.Errorf("attacks=%d, want 2", combat.StanceIndex[types.STANCE_DRAGON].NumAttacks)
	}
}

func TestDoSTset_Attacks_OutOfRange_Rejected(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	before := combat.StanceIndex[types.STANCE_DRAGON].NumAttacks
	DoSTset(ch, "dragon attacks 10")
	if combat.StanceIndex[types.STANCE_DRAGON].NumAttacks != before {
		t.Errorf("attacks=10 should be rejected (range -5..5); got %d",
			combat.StanceIndex[types.STANCE_DRAGON].NumAttacks)
	}
}

func TestDoSTset_Attacks_NegativeBoundary(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	DoSTset(ch, "dragon attacks -5")
	if combat.StanceIndex[types.STANCE_DRAGON].NumAttacks != -5 {
		t.Errorf("attacks=%d, want -5", combat.StanceIndex[types.STANCE_DRAGON].NumAttacks)
	}
}

func TestDoSTset_Attacks_PositiveBoundary(t *testing.T) {
	// Pins the +5 inclusive upper bound against mutation to `> 4`.
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	DoSTset(ch, "dragon attacks 5")
	if combat.StanceIndex[types.STANCE_DRAGON].NumAttacks != 5 {
		t.Errorf("attacks=%d, want 5 (upper boundary inclusive)",
			combat.StanceIndex[types.STANCE_DRAGON].NumAttacks)
	}
}

func TestDoSTset_Damage_InRange(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	DoSTset(ch, "dragon damage 175")
	if combat.StanceIndex[types.STANCE_DRAGON].DamDone != 175 {
		t.Errorf("damage=%d, want 175", combat.StanceIndex[types.STANCE_DRAGON].DamDone)
	}
}

func TestDoSTset_Damage_OutOfRange(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	before := combat.StanceIndex[types.STANCE_DRAGON].DamDone
	DoSTset(ch, "dragon damage 250")
	if combat.StanceIndex[types.STANCE_DRAGON].DamDone != before {
		t.Errorf("damage=250 should be rejected; got %d",
			combat.StanceIndex[types.STANCE_DRAGON].DamDone)
	}
}

func TestDoSTset_Dodge_Signed_InRange(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	DoSTset(ch, "dragon dodge -25")
	if combat.StanceIndex[types.STANCE_DRAGON].Dodge != -25 {
		t.Errorf("dodge=%d, want -25", combat.StanceIndex[types.STANCE_DRAGON].Dodge)
	}
}

// C bug preserved: stset dragon dual 1 → stores 0 (inverted).
func TestDoSTset_Dual_CBugPreserved(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	combat.StanceIndex[types.STANCE_DRAGON].Dual = 99 // marker
	DoSTset(ch, "dragon dual 1")
	if combat.StanceIndex[types.STANCE_DRAGON].Dual != 0 {
		t.Errorf("dual 1 should STORE 0 (C bug); got %d",
			combat.StanceIndex[types.STANCE_DRAGON].Dual)
	}
	DoSTset(ch, "dragon dual 0")
	if combat.StanceIndex[types.STANCE_DRAGON].Dual != 1 {
		t.Errorf("dual 0 should STORE 1 (C bug); got %d",
			combat.StanceIndex[types.STANCE_DRAGON].Dual)
	}
}

// C bug preserved: class setter body is absent — silent no-op.
func TestDoSTset_Class_NoOp_CBugPreserved(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	combat.StanceIndex[types.STANCE_DRAGON].Class = 7
	DoSTset(ch, "dragon class 15")
	if combat.StanceIndex[types.STANCE_DRAGON].Class != 7 {
		t.Errorf("class setter should no-op; got %d",
			combat.StanceIndex[types.STANCE_DRAGON].Class)
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "Done.") {
		t.Errorf("class no-op should still echo Done; got:\n%s", out)
	}
}

func TestDoSTset_Race_NoOp_CBugPreserved(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	combat.StanceIndex[types.STANCE_DRAGON].Race = 5
	DoSTset(ch, "dragon race 2")
	if combat.StanceIndex[types.STANCE_DRAGON].Race != 5 {
		t.Errorf("race setter should no-op; got %d",
			combat.StanceIndex[types.STANCE_DRAGON].Race)
	}
}

func TestDoSTset_Immune_TogglesBit(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	combat.StanceIndex[types.STANCE_DRAGON].Immune = 0
	DoSTset(ch, "dragon immune fire")
	if combat.StanceIndex[types.STANCE_DRAGON].Immune != 1 {
		t.Errorf("immune fire → bit 0; got %d",
			combat.StanceIndex[types.STANCE_DRAGON].Immune)
	}
	// Toggle off.
	DoSTset(ch, "dragon immune fire")
	if combat.StanceIndex[types.STANCE_DRAGON].Immune != 0 {
		t.Errorf("immune fire toggle-off; got %d",
			combat.StanceIndex[types.STANCE_DRAGON].Immune)
	}
}

func TestDoSTset_Immune_UnknownFlag_Rejected(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	combat.StanceIndex[types.STANCE_DRAGON].Immune = 0
	DoSTset(ch, "dragon immune bogus")
	if combat.StanceIndex[types.STANCE_DRAGON].Immune != 0 {
		t.Errorf("unknown flag should not mutate; got %d",
			combat.StanceIndex[types.STANCE_DRAGON].Immune)
	}
	out := readOutput(ch, client)
	if !strings.Contains(out, "Unkown flag") {
		t.Errorf("unknown flag should print typo'd error; got:\n%s", out)
	}
}

func TestDoSTset_Others_Self_StoresString(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	DoSTset(ch, "dragon others $n becomes fierce")
	if combat.StanceIndex[types.STANCE_DRAGON].Others != "$n becomes fierce" {
		t.Errorf("others not stored; got %q",
			combat.StanceIndex[types.STANCE_DRAGON].Others)
	}
	DoSTset(ch, "dragon self you feel fierce")
	if combat.StanceIndex[types.STANCE_DRAGON].Self != "you feel fierce" {
		t.Errorf("self not stored; got %q",
			combat.StanceIndex[types.STANCE_DRAGON].Self)
	}
}

func TestDoSTset_Parry_InRange(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	DoSTset(ch, "dragon parry 30")
	if combat.StanceIndex[types.STANCE_DRAGON].Parry != 30 {
		t.Errorf("parry=%d, want 30", combat.StanceIndex[types.STANCE_DRAGON].Parry)
	}
}

func TestDoSTset_Percent_InRange(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	DoSTset(ch, "dragon percent 50")
	if combat.StanceIndex[types.STANCE_DRAGON].SpecialPercent != 50 {
		t.Errorf("percent=%d, want 50", combat.StanceIndex[types.STANCE_DRAGON].SpecialPercent)
	}
}

func TestDoSTset_Protection_SetsDamTaken(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	DoSTset(ch, "dragon protection 80")
	if combat.StanceIndex[types.STANCE_DRAGON].DamTaken != 80 {
		t.Errorf("protection=%d, want 80",
			combat.StanceIndex[types.STANCE_DRAGON].DamTaken)
	}
}

func TestDoSTset_Resist_TogglesBit(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	combat.StanceIndex[types.STANCE_DRAGON].Resist = 0
	DoSTset(ch, "dragon resist cold")
	if combat.StanceIndex[types.STANCE_DRAGON].Resist != 2 {
		t.Errorf("resist cold → bit 1; got %d",
			combat.StanceIndex[types.STANCE_DRAGON].Resist)
	}
}

func TestDoSTset_Special_InvokesStub_SetsSpecialMoveToZero(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	combat.StanceIndex[types.STANCE_DRAGON].SpecialMove = 7 // non-zero marker
	DoSTset(ch, "dragon special fireball")
	if combat.StanceIndex[types.STANCE_DRAGON].SpecialMove != 0 {
		t.Errorf("special stub should set 0; got %d",
			combat.StanceIndex[types.STANCE_DRAGON].SpecialMove)
	}
}

func TestDoSTset_Susceptible_TogglesBit(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	combat.StanceIndex[types.STANCE_DRAGON].Suscept = 0
	DoSTset(ch, "dragon susceptible acid")
	if combat.StanceIndex[types.STANCE_DRAGON].Suscept != int(types.RIS_ACID) {
		t.Errorf("susceptible acid → RIS_ACID bit; got %d",
			combat.StanceIndex[types.STANCE_DRAGON].Suscept)
	}
}

func TestDoSTset_Stance1_SetsPrereq0(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	combat.StanceIndex[types.STANCE_DRAGON].Prereq = [2]int{0, 0}
	DoSTset(ch, "dragon stance1 tiger")
	if combat.StanceIndex[types.STANCE_DRAGON].Prereq[0] != types.STANCE_TIGER {
		t.Errorf("Prereq[0] = %d, want TIGER=%d",
			combat.StanceIndex[types.STANCE_DRAGON].Prereq[0], types.STANCE_TIGER)
	}
}

func TestDoSTset_Stance1_InvalidName_Rejected(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	combat.StanceIndex[types.STANCE_DRAGON].Prereq = [2]int{0, 0}
	DoSTset(ch, "dragon stance1 bogus")
	if combat.StanceIndex[types.STANCE_DRAGON].Prereq[0] != 0 {
		t.Errorf("invalid name should not mutate Prereq[0]; got %d",
			combat.StanceIndex[types.STANCE_DRAGON].Prereq[0])
	}
}

func TestDoSTset_Stance2_SetsPrereq1(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	combat.StanceIndex[types.STANCE_DRAGON].Prereq = [2]int{0, 0}
	DoSTset(ch, "dragon stance2 swallow")
	if combat.StanceIndex[types.STANCE_DRAGON].Prereq[1] != types.STANCE_SWALLOW {
		t.Errorf("Prereq[1] = %d, want SWALLOW=%d",
			combat.StanceIndex[types.STANCE_DRAGON].Prereq[1], types.STANCE_SWALLOW)
	}
}

func TestDoSTset_Wait_InRange(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	DoSTset(ch, "dragon wait 15")
	if combat.StanceIndex[types.STANCE_DRAGON].Wait != 15 {
		t.Errorf("wait=%d, want 15", combat.StanceIndex[types.STANCE_DRAGON].Wait)
	}
}

func TestDoSTset_Wait_OutOfRange(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	before := combat.StanceIndex[types.STANCE_DRAGON].Wait
	DoSTset(ch, "dragon wait 99")
	if combat.StanceIndex[types.STANCE_DRAGON].Wait != before {
		t.Errorf("wait=99 should be rejected (range 0..25); got %d",
			combat.StanceIndex[types.STANCE_DRAGON].Wait)
	}
}

func TestDoSTset_Weight_InRange(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	DoSTset(ch, "dragon weight 500")
	if combat.StanceIndex[types.STANCE_DRAGON].MaxWeight != 500 {
		t.Errorf("weight=%d, want 500", combat.StanceIndex[types.STANCE_DRAGON].MaxWeight)
	}
}

func TestDoSTset_UnknownField_ShowsUsage(t *testing.T) {
	saveStanceIndex(t)
	ch, client := newAdminChar(t)
	defer client.Close()
	DoSTset(ch, "dragon bogusfield 5")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Syntax: stset") {
		t.Errorf("unknown field should fall through to usage; got:\n%s", out)
	}
}

// --- stset save round-trip ---

func TestDoSTset_Save_InvokesWriter(t *testing.T) {
	saveStanceIndex(t)
	savedWorld := WorldRef
	savedPath := persist.StancePath
	t.Cleanup(func() {
		WorldRef = savedWorld
		persist.StancePath = savedPath
	})

	dir := t.TempDir()
	ensureSystemDirForTest(t, dir)
	WorldRef = world.New(dir)
	persist.StancePath = filepath.Join(dir, "system", "stances.dat")

	ch, client := newAdminChar(t)
	defer client.Close()
	combat.StanceIndex[types.STANCE_DRAGON].NumAttacks = 5
	DoSTset(ch, "save")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Done.") {
		t.Errorf("save should echo Done.; got:\n%s", out)
	}

	var fresh [types.MAX_STANCE]combat.StanceInfo
	if err := persist.LoadStancesInto(&fresh, persist.StancePath); err != nil {
		t.Fatalf("reload: %v", err)
	}
	if fresh[types.STANCE_DRAGON].NumAttacks != 5 {
		t.Errorf("round-trip: got %d, want 5", fresh[types.STANCE_DRAGON].NumAttacks)
	}
}

// --- Scenario tests (G5 acceptance criteria) ---

func TestScenario_StsetEditThenSaveThenReload(t *testing.T) {
	saveStanceIndex(t)
	savedWorld := WorldRef
	savedPath := persist.StancePath
	t.Cleanup(func() {
		WorldRef = savedWorld
		persist.StancePath = savedPath
	})
	dir := t.TempDir()
	ensureSystemDirForTest(t, dir)
	WorldRef = world.New(dir)
	persist.StancePath = filepath.Join(dir, "system", "stances.dat")

	admin, client := newAdminChar(t)
	defer client.Close()
	DoSTset(admin, "dragon attacks 4")
	DoSTset(admin, "save")
	var fresh [types.MAX_STANCE]combat.StanceInfo
	if err := persist.LoadStancesInto(&fresh, persist.StancePath); err != nil {
		t.Fatal(err)
	}
	if fresh[types.STANCE_DRAGON].NumAttacks != 4 {
		t.Errorf("round-trip: got %d, want 4", fresh[types.STANCE_DRAGON].NumAttacks)
	}
}

func TestScenario_MortalCannotChangeStanceWhileActive(t *testing.T) {
	saveStanceIndex(t)
	ch, _ := newSkillTestRoom()
	ch.Stance = types.STANCE_DRAGON
	DoStance(ch, "tiger")
	if ch.Stance != types.STANCE_DRAGON {
		t.Errorf("change-while-active should be refused; got %d", ch.Stance)
	}
}

func TestScenario_MortalCanUseAfterGMMastery(t *testing.T) {
	saveStanceIndex(t)
	ch, _ := newSkillTestRoom()
	ch.Stance = types.STANCE_NONE
	ch.PCData.Stances[types.STANCE_TIGER] = 200
	combat.StanceIndex[types.STANCE_DRAGON] = combat.StanceInfo{
		Prereq: [2]int{types.STANCE_TIGER, 0},
	}
	DoStance(ch, "dragon")
	if ch.Stance != types.STANCE_DRAGON {
		t.Errorf("GM-mastery on TIGER should unlock DRAGON; got %d", ch.Stance)
	}
}

func TestScenario_MortalCannotUseWithoutGMMastery(t *testing.T) {
	saveStanceIndex(t)
	ch, _ := newSkillTestRoom()
	ch.Stance = types.STANCE_NONE
	ch.PCData.Stances[types.STANCE_TIGER] = 199
	combat.StanceIndex[types.STANCE_DRAGON] = combat.StanceInfo{
		Prereq: [2]int{types.STANCE_TIGER, 0},
	}
	DoStance(ch, "dragon")
	if ch.Stance != types.STANCE_NONE {
		t.Errorf("TIGER at 199 should NOT unlock DRAGON; got %d", ch.Stance)
	}
}
