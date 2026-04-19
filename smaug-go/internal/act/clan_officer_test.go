package act

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// setupClanOfficerWorld seeds a world with a pkill-type clan, an officer
// ch (clan.Leader), and a non-clan victim in the same room. Both PCs have
// net.Pipe-backed descriptors for output capture.
func setupClanOfficerWorld(t *testing.T, clanType int) (
	w *world.World,
	room *types.RoomIndexData,
	officer *types.CharData, officerClient net.Conn,
	victim *types.CharData, victimClient net.Conn,
	clan *types.ClanData,
) {
	t.Helper()
	w = world.New("/tmp/test")
	WorldRef = w
	clan = &types.ClanData{
		Filename: "test.clan",
		Name:     "Test Clan",
		Leader:   "Alice",
		ClanType: clanType,
		Class:    2,
	}
	w.Clans = append(w.Clans, clan)

	room = &types.RoomIndexData{Vnum: 1000, Name: "Clan Hall"}

	officer, officerClient = makeTestChar("Alice")
	officer.Level = 50
	officer.Class = 2
	officer.PCData.Flags = int(types.PCFLAG_DEADLY)
	officer.PCData.Clan = clan
	officer.PCData.ClanName = clan.Name
	officer.InRoom = room
	room.People = append(room.People, officer)

	victim, victimClient = makeTestChar("Bob")
	victim.Level = 20
	victim.Class = 2
	victim.PCData.Flags = int(types.PCFLAG_DEADLY)
	victim.Act.Set(types.PLR_NICE)
	victim.InRoom = room
	room.People = append(room.People, victim)

	return w, room, officer, officerClient, victim, victimClient, clan
}

// --- isClanOfficer ---

func TestIsClanOfficer_LeaderMatches(t *testing.T) {
	clan := &types.ClanData{Leader: "Alice"}
	ch, client := makeTestChar("Alice")
	defer client.Close()
	if !isClanOfficer(ch, clan, "induct") {
		t.Errorf("leader Alice should be officer for induct")
	}
}

func TestIsClanOfficer_Number1Matches(t *testing.T) {
	clan := &types.ClanData{Number1: "Bob"}
	ch, client := makeTestChar("Bob")
	defer client.Close()
	if !isClanOfficer(ch, clan, "induct") {
		t.Errorf("Number1 Bob should be officer")
	}
}

func TestIsClanOfficer_Number2Matches(t *testing.T) {
	clan := &types.ClanData{Number2: "Carol"}
	ch, client := makeTestChar("Carol")
	defer client.Close()
	if !isClanOfficer(ch, clan, "induct") {
		t.Errorf("Number2 Carol should be officer")
	}
}

func TestIsClanOfficer_DeityMatches(t *testing.T) {
	clan := &types.ClanData{Deity: "Mota"}
	ch, client := makeTestChar("Mota")
	defer client.Close()
	if !isClanOfficer(ch, clan, "induct") {
		t.Errorf("clan deity Mota should be officer")
	}
}

func TestIsClanOfficer_BestowmentsMatches(t *testing.T) {
	clan := &types.ClanData{Leader: "Alice"}
	ch, client := makeTestChar("Bob")
	defer client.Close()
	ch.PCData.Bestowments = "induct outcast"
	if !isClanOfficer(ch, clan, "induct") {
		t.Errorf("Bob with bestowed 'induct' should be officer")
	}
	if !isClanOfficer(ch, clan, "outcast") {
		t.Errorf("Bob with bestowed 'outcast' should be officer")
	}
}

func TestIsClanOfficer_BestowmentsDoesNotMatchOtherKeyword(t *testing.T) {
	clan := &types.ClanData{Leader: "Alice"}
	ch, client := makeTestChar("Bob")
	defer client.Close()
	ch.PCData.Bestowments = "induct"
	// "outcast" not bestowed
	if isClanOfficer(ch, clan, "outcast") {
		t.Errorf("Bob bestowed only 'induct' should NOT be officer for 'outcast'")
	}
}

func TestIsClanOfficer_NoMatch(t *testing.T) {
	clan := &types.ClanData{Leader: "Alice", Number1: "Bob", Number2: "Carol", Deity: "Mota"}
	ch, client := makeTestChar("Eve")
	defer client.Close()
	if isClanOfficer(ch, clan, "induct") {
		t.Errorf("random Eve should not be officer")
	}
}

func TestIsClanOfficer_CaseInsensitive(t *testing.T) {
	clan := &types.ClanData{Leader: "Alice"}
	ch, client := makeTestChar("alice")
	defer client.Close()
	if !isClanOfficer(ch, clan, "induct") {
		t.Errorf("lowercase 'alice' should match leader 'Alice' case-insensitively")
	}
}

func TestIsClanOfficer_NilCh(t *testing.T) {
	clan := &types.ClanData{Leader: "Alice"}
	if isClanOfficer(nil, clan, "induct") {
		t.Errorf("nil ch should not be officer")
	}
}

func TestIsClanOfficer_NilClan(t *testing.T) {
	ch, client := makeTestChar("Alice")
	defer client.Close()
	if isClanOfficer(ch, nil, "induct") {
		t.Errorf("nil clan should not be officer")
	}
}

func TestIsClanOfficer_NPC(t *testing.T) {
	clan := &types.ClanData{Leader: "Alice"}
	ch, client := makeTestChar("Alice")
	defer client.Close()
	ch.Act.Set(types.ACT_IS_NPC)
	if isClanOfficer(ch, clan, "induct") {
		t.Errorf("NPC should not be officer")
	}
}

// --- isPkill ---

func TestIsPkill_FlagSet(t *testing.T) {
	ch, client := makeTestChar("P")
	defer client.Close()
	ch.PCData.Flags = int(types.PCFLAG_DEADLY)
	if !isPkill(ch) {
		t.Errorf("PCFLAG_DEADLY should make isPkill true")
	}
}

func TestIsPkill_FlagUnset(t *testing.T) {
	ch, client := makeTestChar("P")
	defer client.Close()
	if isPkill(ch) {
		t.Errorf("no PCFLAG_DEADLY should make isPkill false")
	}
}

func TestIsPkill_NPC(t *testing.T) {
	ch, client := makeTestChar("P")
	defer client.Close()
	ch.Act.Set(types.ACT_IS_NPC)
	if isPkill(ch) {
		t.Errorf("NPC should not be pkill")
	}
}

// --- DoInduct gates ---

func TestDoInduct_NPCCallerGetsHuh(t *testing.T) {
	WorldRef = world.New("/tmp/test")
	ch, client := makeTestChar("Mob")
	defer client.Close()
	ch.Act.Set(types.ACT_IS_NPC)
	DoInduct(ch, "victim")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Huh?") {
		t.Errorf("NPC caller expected Huh?; got %q", out)
	}
}

func TestDoInduct_NoClanCallerGetsHuh(t *testing.T) {
	WorldRef = world.New("/tmp/test")
	ch, client := makeTestChar("Alice")
	defer client.Close()
	DoInduct(ch, "victim")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Huh?") {
		t.Errorf("no-clan caller expected Huh?; got %q", out)
	}
}

func TestDoInduct_NonOfficerGetsHuh(t *testing.T) {
	_, _, _, oc, _, vc, clan := setupClanOfficerWorld(t, 0 /* pkill */)
	defer oc.Close()
	defer vc.Close()
	// Re-create caller as a plain clan member, NOT officer.
	member, mc := makeTestChar("Dave")
	defer mc.Close()
	member.PCData.Clan = clan
	DoInduct(member, "victim")
	out := readOutput(member, mc)
	if !strings.Contains(out, "Huh?") {
		t.Errorf("non-officer expected Huh?; got %q", out)
	}
}

func TestDoInduct_EmptyArg(t *testing.T) {
	_, _, officer, oc, _, vc, _ := setupClanOfficerWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	DoInduct(officer, "")
	out := readOutput(officer, oc)
	if !strings.Contains(out, "Induct whom?") {
		t.Errorf("empty arg expected 'Induct whom?'; got %q", out)
	}
}

func TestDoInduct_VictimNotInRoom(t *testing.T) {
	_, _, officer, oc, _, vc, _ := setupClanOfficerWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	DoInduct(officer, "ghost")
	out := readOutput(officer, oc)
	if !strings.Contains(out, "That player is not here.") {
		t.Errorf("expected 'That player is not here.'; got %q", out)
	}
}

func TestDoInduct_VictimIsNPC(t *testing.T) {
	_, room, officer, oc, _, vc, _ := setupClanOfficerWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	mob := &types.CharData{Name: "guard", ShortDescr: "a guard", InRoom: room}
	mob.Act.Set(types.ACT_IS_NPC)
	room.People = append(room.People, mob)
	DoInduct(officer, "guard")
	out := readOutput(officer, oc)
	if !strings.Contains(out, "Not on NPC's.") {
		t.Errorf("expected 'Not on NPC's.'; got %q", out)
	}
}

func TestDoInduct_VictimIsImmortal(t *testing.T) {
	_, room, officer, oc, _, vc, _ := setupClanOfficerWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	imm, ic := makeTestChar("Godd")
	defer ic.Close()
	imm.Level = types.LEVEL_IMMORTAL
	imm.InRoom = room
	room.People = append(room.People, imm)
	DoInduct(officer, "godd")
	out := readOutput(officer, oc)
	if !strings.Contains(out, "godly presence") {
		t.Errorf("expected 'godly presence'; got %q", out)
	}
}

func TestDoInduct_PeacefulVictim_PkillClan_Rejected(t *testing.T) {
	_, _, officer, oc, victim, vc, _ := setupClanOfficerWorld(t, 0 /* pkill */)
	defer oc.Close()
	defer vc.Close()
	victim.PCData.Flags = 0 // clear PCFLAG_DEADLY
	DoInduct(officer, "bob")
	out := readOutput(officer, oc)
	if !strings.Contains(out, "cannot induct a peaceful character") {
		t.Errorf("expected peaceful rejection; got %q", out)
	}
}

func TestDoInduct_PeacefulVictim_GuildClan_ProceedsToClassCheck(t *testing.T) {
	_, _, officer, oc, victim, vc, clan := setupClanOfficerWorld(t, types.CLAN_GUILD)
	defer oc.Close()
	defer vc.Close()
	victim.PCData.Flags = 0
	victim.Class = 5 // mismatch clan.Class=2
	_ = clan
	DoInduct(officer, "bob")
	out := readOutput(officer, oc)
	// Should NOT contain peaceful rejection.
	if strings.Contains(out, "peaceful character") {
		t.Errorf("guild clan should accept peaceful inductee; got %q", out)
	}
	// SHOULD get class-mismatch message instead.
	if !strings.Contains(out, "not in accordance with your guild") {
		t.Errorf("expected guild class-mismatch; got %q", out)
	}
}

func TestDoInduct_GuildClassMismatch(t *testing.T) {
	_, _, officer, oc, victim, vc, _ := setupClanOfficerWorld(t, types.CLAN_GUILD)
	defer oc.Close()
	defer vc.Close()
	victim.Class = 5 // clan.Class=2
	DoInduct(officer, "bob")
	out := readOutput(officer, oc)
	if !strings.Contains(out, "not in accordance with your guild") {
		t.Errorf("expected guild class-mismatch; got %q", out)
	}
}

func TestDoInduct_VictimLevelTooLow(t *testing.T) {
	_, _, officer, oc, victim, vc, _ := setupClanOfficerWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	victim.Level = 5 // below 10
	DoInduct(officer, "bob")
	out := readOutput(officer, oc)
	if !strings.Contains(out, "not worthy of joining yet") {
		t.Errorf("expected low-level rejection; got %q", out)
	}
}

// TestDoInduct_VictimLevelExactly10_Allowed pins the boundary: C uses
// `victim->level < 10`, so level==10 is allowed. Protects against a
// strict-vs-non-strict off-by-one mutation.
func TestDoInduct_VictimLevelExactly10_Allowed(t *testing.T) {
	_, _, officer, oc, victim, vc, _ := setupClanOfficerWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	victim.Level = 10
	DoInduct(officer, "bob")
	out := readOutput(officer, oc)
	if strings.Contains(out, "not worthy of joining yet") {
		t.Errorf("level 10 should be accepted (C uses < 10, strict); got %q", out)
	}
}

func TestDoInduct_VictimLevelTooHigh(t *testing.T) {
	_, _, officer, oc, victim, vc, _ := setupClanOfficerWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	// Keep both below LEVEL_IMMORTAL (=51) so the IsImmortal gate doesn't fire.
	officer.Level = 30
	victim.Level = 40
	DoInduct(officer, "bob")
	out := readOutput(officer, oc)
	if !strings.Contains(out, "too powerful for you to induct") {
		t.Errorf("expected too-powerful rejection; got %q", out)
	}
}

func TestDoInduct_AlreadyInYourClan(t *testing.T) {
	_, _, officer, oc, victim, vc, clan := setupClanOfficerWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	victim.PCData.Clan = clan
	DoInduct(officer, "bob")
	out := readOutput(officer, oc)
	if !strings.Contains(out, "already belongs to your clan") {
		t.Errorf("expected 'already belongs to your clan'; got %q", out)
	}
}

func TestDoInduct_AlreadyInADifferentClan(t *testing.T) {
	_, _, officer, oc, victim, vc, _ := setupClanOfficerWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	other := &types.ClanData{Name: "Other", ClanType: 0}
	victim.PCData.Clan = other
	DoInduct(officer, "bob")
	out := readOutput(officer, oc)
	if !strings.Contains(out, "already belongs to a clan") {
		t.Errorf("expected 'already belongs to a clan'; got %q", out)
	}
}

func TestDoInduct_AlreadyInYourOrder(t *testing.T) {
	_, _, officer, oc, victim, vc, clan := setupClanOfficerWorld(t, types.CLAN_ORDER)
	defer oc.Close()
	defer vc.Close()
	clan.ClanType = types.CLAN_ORDER
	victim.PCData.Clan = clan
	DoInduct(officer, "bob")
	out := readOutput(officer, oc)
	if !strings.Contains(out, "already belongs to your order") {
		t.Errorf("expected 'already belongs to your order'; got %q", out)
	}
}

func TestDoInduct_AlreadyInAnOrder(t *testing.T) {
	_, _, officer, oc, victim, vc, _ := setupClanOfficerWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	other := &types.ClanData{Name: "Other", ClanType: types.CLAN_ORDER}
	victim.PCData.Clan = other
	DoInduct(officer, "bob")
	out := readOutput(officer, oc)
	if !strings.Contains(out, "already belongs to an order") {
		t.Errorf("expected 'already belongs to an order'; got %q", out)
	}
}

func TestDoInduct_AlreadyInYourGuild(t *testing.T) {
	_, _, officer, oc, victim, vc, clan := setupClanOfficerWorld(t, types.CLAN_GUILD)
	defer oc.Close()
	defer vc.Close()
	victim.PCData.Clan = clan
	DoInduct(officer, "bob")
	out := readOutput(officer, oc)
	if !strings.Contains(out, "already belongs to your guild") {
		t.Errorf("expected 'already belongs to your guild'; got %q", out)
	}
}

func TestDoInduct_AlreadyInAGuild(t *testing.T) {
	_, _, officer, oc, victim, vc, _ := setupClanOfficerWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	other := &types.ClanData{Name: "Other", ClanType: types.CLAN_GUILD}
	victim.PCData.Clan = other
	DoInduct(officer, "bob")
	out := readOutput(officer, oc)
	if !strings.Contains(out, "already belongs to an guild") {
		// NB: C has literal "an guild" (grammar bug preserved verbatim).
		t.Errorf("expected 'already belongs to an guild' (C verbatim); got %q", out)
	}
}

func TestDoInduct_MemberLimitHit(t *testing.T) {
	_, _, officer, oc, _, vc, clan := setupClanOfficerWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	clan.MemLimit = 1
	clan.Members = 1
	DoInduct(officer, "bob")
	out := readOutput(officer, oc)
	if !strings.Contains(out, "too big to induct anymore") {
		t.Errorf("expected member-limit rejection; got %q", out)
	}
}

func TestDoInduct_MemLimitZeroIsUnlimited(t *testing.T) {
	_, _, officer, oc, _, vc, clan := setupClanOfficerWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	clan.MemLimit = 0 // unlimited per C (if !mem_limit)
	clan.Members = 999
	DoInduct(officer, "bob")
	if clan.Members != 1000 {
		t.Errorf("MemLimit=0 should permit induct; Members=%d", clan.Members)
	}
}

// --- DoInduct happy path ---

func TestDoInduct_Success_PkillClan_SetsFields(t *testing.T) {
	_, _, officer, oc, victim, vc, clan := setupClanOfficerWorld(t, 0 /* pkill */)
	defer oc.Close()
	defer vc.Close()
	origMembers := clan.Members
	DoInduct(officer, "bob")
	if victim.PCData.Clan != clan {
		t.Errorf("victim.PCData.Clan = %v, want clan", victim.PCData.Clan)
	}
	if victim.PCData.ClanName != clan.Name {
		t.Errorf("ClanName = %q, want %q", victim.PCData.ClanName, clan.Name)
	}
	if clan.Members != origMembers+1 {
		t.Errorf("Members = %d, want %d", clan.Members, origMembers+1)
	}
	if uint32(victim.Speaks)&types.LANG_CLAN == 0 {
		t.Error("LANG_CLAN should be set on victim.Speaks after induct")
	}
	if victim.Act.IsSet(types.PLR_NICE) {
		t.Error("PLR_NICE should be cleared by pkill-clan induct")
	}
	if uint32(victim.PCData.Flags)&types.PCFLAG_DEADLY == 0 {
		t.Error("PCFLAG_DEADLY should be set by pkill-clan induct")
	}
}

func TestDoInduct_Success_GuildClan_DoesNotTouchNice(t *testing.T) {
	_, _, officer, oc, victim, vc, clan := setupClanOfficerWorld(t, types.CLAN_GUILD)
	defer oc.Close()
	defer vc.Close()
	victim.Class = 2
	victim.Act.Set(types.PLR_NICE)
	_ = clan
	DoInduct(officer, "bob")
	if !victim.Act.IsSet(types.PLR_NICE) {
		t.Error("PLR_NICE must survive a guild-clan induct (non-pkill-type)")
	}
	// LANG_CLAN NOT set for guild.
	if uint32(victim.Speaks)&types.LANG_CLAN != 0 {
		t.Error("LANG_CLAN must NOT be set on guild induct")
	}
}

func TestDoInduct_Success_OrderClan_DoesNotSetLangClan(t *testing.T) {
	_, _, officer, oc, victim, vc, _ := setupClanOfficerWorld(t, types.CLAN_ORDER)
	defer oc.Close()
	defer vc.Close()
	DoInduct(officer, "bob")
	if uint32(victim.Speaks)&types.LANG_CLAN != 0 {
		t.Error("LANG_CLAN must NOT be set on order induct")
	}
}

func TestDoInduct_Success_NokillClan_SetsLangClanButNotDeadly(t *testing.T) {
	_, _, officer, oc, victim, vc, _ := setupClanOfficerWorld(t, types.CLAN_NOKILL)
	defer oc.Close()
	defer vc.Close()
	victim.PCData.Flags = 0
	DoInduct(officer, "bob")
	if uint32(victim.Speaks)&types.LANG_CLAN == 0 {
		t.Error("LANG_CLAN must be set on nokill induct (non-order/non-guild)")
	}
	if uint32(victim.PCData.Flags)&types.PCFLAG_DEADLY != 0 {
		t.Error("PCFLAG_DEADLY must NOT be set by nokill induct")
	}
}

func TestDoInduct_Success_AwardsGuildSkills(t *testing.T) {
	w, _, officer, oc, victim, vc, _ := setupClanOfficerWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	// Register a skill with Guild matching clan.Class=2.
	sk := &types.SkillType{Name: "testskill", Guild: 2}
	sk.SkillAdept[2] = 75
	// Fill index 0 with nil and skill at 1 so sn is clearly nontrivial.
	w.Skills = []*types.SkillType{nil, sk}
	DoInduct(officer, "bob")
	if victim.PCData.Learned[1] != 75 {
		t.Errorf("Learned[1] = %d, want 75 (skill adept value)", victim.PCData.Learned[1])
	}
}

func TestDoInduct_Success_DoesNotAwardWrongGuildSkills(t *testing.T) {
	w, _, officer, oc, victim, vc, clan := setupClanOfficerWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	sk := &types.SkillType{Name: "mismatch", Guild: 99} // guild != clan.Class=2
	sk.SkillAdept[2] = 75
	w.Skills = []*types.SkillType{sk}
	_ = clan
	DoInduct(officer, "bob")
	if victim.PCData.Learned[0] != 0 {
		t.Errorf("Learned[0] should be 0 for non-matching guild skill; got %d", victim.PCData.Learned[0])
	}
}

func TestDoInduct_Success_Messages(t *testing.T) {
	_, _, officer, oc, victim, vc, _ := setupClanOfficerWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	DoInduct(officer, "bob")
	officerOut := readOutput(officer, oc)
	victimOut := readOutput(victim, vc)
	if !strings.Contains(officerOut, "You induct Bob into Test Clan") {
		t.Errorf("officer message mismatch; got %q", officerOut)
	}
	if !strings.Contains(victimOut, "Alice inducts you into Test Clan") {
		t.Errorf("victim message mismatch; got %q", victimOut)
	}
}

func TestDoInduct_Persists_SaveFuncCalled(t *testing.T) {
	_, _, officer, oc, _, vc, _ := setupClanOfficerWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	prev := SaveFunc
	var saved bool
	SaveFunc = func(c *types.CharData) { saved = true }
	defer func() { SaveFunc = prev }()
	DoInduct(officer, "bob")
	if !saved {
		t.Error("SaveFunc should be called on successful induct")
	}
}

func TestDoInduct_Persists_SaveClanFileCalled(t *testing.T) {
	_, _, officer, oc, _, vc, clan := setupClanOfficerWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	// Point ClanDir at temp dir; DoInduct will call SaveClanFile.
	prevDir := ClanDir
	ClanDir = t.TempDir()
	defer func() { ClanDir = prevDir }()
	DoInduct(officer, "bob")
	written := filepath.Join(ClanDir, clan.Filename)
	if _, err := os.Stat(written); err != nil {
		t.Errorf("expected SaveClanFile to write %s: %v", written, err)
	}
}

func TestDoInduct_Persists_NilClanDirSilent(t *testing.T) {
	_, _, officer, oc, _, vc, _ := setupClanOfficerWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	prevDir := ClanDir
	ClanDir = ""
	defer func() { ClanDir = prevDir }()
	// Must not panic even when ClanDir is empty.
	DoInduct(officer, "bob")
}

// --- DoOutcast gates ---

// helper: set up same world as induct but with victim already in clan.
func setupOutcastWorld(t *testing.T, clanType int) (
	w *world.World,
	room *types.RoomIndexData,
	officer *types.CharData, oc net.Conn,
	victim *types.CharData, vc net.Conn,
	clan *types.ClanData,
) {
	t.Helper()
	w, room, officer, oc, victim, vc, clan = setupClanOfficerWorld(t, clanType)
	victim.PCData.Clan = clan
	victim.PCData.ClanName = clan.Name
	victim.Speaks = int(types.LANG_CLAN | types.LANG_COMMON)
	victim.Speaking = int(types.LANG_CLAN)
	clan.Members = 2
	return w, room, officer, oc, victim, vc, clan
}

func TestDoOutcast_NPCGetsHuh(t *testing.T) {
	WorldRef = world.New("/tmp/test")
	ch, client := makeTestChar("Mob")
	defer client.Close()
	ch.Act.Set(types.ACT_IS_NPC)
	DoOutcast(ch, "victim")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Huh?") {
		t.Errorf("NPC expected Huh?; got %q", out)
	}
}

func TestDoOutcast_NoClanGetsHuh(t *testing.T) {
	WorldRef = world.New("/tmp/test")
	ch, client := makeTestChar("Alice")
	defer client.Close()
	DoOutcast(ch, "victim")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Huh?") {
		t.Errorf("no-clan expected Huh?; got %q", out)
	}
}

func TestDoOutcast_NonOfficerGetsHuh(t *testing.T) {
	_, _, _, oc, _, vc, clan := setupOutcastWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	nonOff, nc := makeTestChar("Dave")
	defer nc.Close()
	nonOff.PCData.Clan = clan
	DoOutcast(nonOff, "bob")
	out := readOutput(nonOff, nc)
	if !strings.Contains(out, "Huh?") {
		t.Errorf("non-officer expected Huh?; got %q", out)
	}
}

func TestDoOutcast_EmptyArg(t *testing.T) {
	_, _, officer, oc, _, vc, _ := setupOutcastWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	DoOutcast(officer, "")
	out := readOutput(officer, oc)
	if !strings.Contains(out, "Outcast whom?") {
		t.Errorf("empty arg expected 'Outcast whom?'; got %q", out)
	}
}

func TestDoOutcast_VictimNotHere(t *testing.T) {
	_, _, officer, oc, _, vc, _ := setupOutcastWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	DoOutcast(officer, "ghost")
	out := readOutput(officer, oc)
	if !strings.Contains(out, "That player is not here.") {
		t.Errorf("expected 'That player is not here.'; got %q", out)
	}
}

func TestDoOutcast_VictimIsNPC(t *testing.T) {
	_, room, officer, oc, _, vc, _ := setupOutcastWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	mob := &types.CharData{Name: "guard", InRoom: room}
	mob.Act.Set(types.ACT_IS_NPC)
	room.People = append(room.People, mob)
	DoOutcast(officer, "guard")
	out := readOutput(officer, oc)
	if !strings.Contains(out, "Not on NPC's.") {
		t.Errorf("expected 'Not on NPC's.'; got %q", out)
	}
}

// --- DoOutcast rank arithmetic ---

func TestDoOutcast_RankArithmetic_OfficerOutrankedByVictim(t *testing.T) {
	_, _, officer, oc, victim, vc, clan := setupOutcastWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	// Rewire: clan.Leader = victim (rank 3), officer = number2 (rank 1).
	clan.Leader = victim.Name
	clan.Number2 = officer.Name
	officer.PCData.Clan = clan
	// Gate: x <= y && trust <= trust. x=1, y=3, so x<=y is true.
	// Make trusts equal so trust<=trust is also true → gate fires.
	victim.Level = officer.Level
	DoOutcast(officer, "bob")
	out := readOutput(officer, oc)
	if !strings.Contains(out, "not powerful enough") {
		t.Errorf("outranked officer expected 'not powerful enough'; got %q", out)
	}
}

func TestDoOutcast_RankArithmetic_SuperiorTrustBeatsEqualRank(t *testing.T) {
	_, _, officer, oc, victim, vc, clan := setupOutcastWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	// Both are Number1 → x=y=2; officer has higher trust.
	clan.Number1 = officer.Name
	clan.Leader = "" // not leader of either
	victim.Level = 10
	officer.Level = 50 // higher trust
	// Officer is NOT also leader → isClanOfficer still true via Number1 slot.
	// Swap by bumping trust; but rank test is x<=y && trust<=trust.
	// officer x=2 (number1), victim y=0, so x<=y is false → proceed.
	DoOutcast(officer, "bob")
	// Should proceed past the power gate.
	if victim.PCData.Clan != nil {
		t.Errorf("outcast should succeed; victim still in clan %+v", victim.PCData.Clan)
	}
}

// TestDoOutcast_SelfOutcast_PowerGateBlocks documents that C's self-outcast
// messages at src/clans.c:1238-1255 are structurally dead code for any
// caller reachable via isClanOfficer: when victim == ch, x == y (both derived
// from the same name) AND trust == trust, so the power gate at
// src/clans.c:1231 (x <= y && trust <= trust) always fires first. The Go
// port matches C verbatim; the self-outcast branch is preserved in code but
// functionally unreachable. This test pins that observation.
func TestDoOutcast_SelfOutcast_PowerGateBlocks(t *testing.T) {
	_, _, officer, oc, _, vc, _ := setupOutcastWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	DoOutcast(officer, officer.Name)
	out := readOutput(officer, oc)
	if !strings.Contains(out, "not powerful enough") {
		t.Errorf("power gate should fire first on self-outcast (C fidelity); got %q", out)
	}
}

func TestDoOutcast_VictimHigherLevel(t *testing.T) {
	_, _, officer, oc, victim, vc, _ := setupOutcastWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	victim.Level = 60
	officer.Level = 50
	// Keep officer as Leader (x=3), victim rank 0 — power gate passes.
	DoOutcast(officer, "bob")
	out := readOutput(officer, oc)
	if !strings.Contains(out, "too powerful for you to outcast") {
		t.Errorf("expected too-powerful message; got %q", out)
	}
}

func TestDoOutcast_NotSameClan_Clan(t *testing.T) {
	_, _, officer, oc, victim, vc, _ := setupOutcastWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	other := &types.ClanData{Name: "Other"}
	victim.PCData.Clan = other
	DoOutcast(officer, "bob")
	out := readOutput(officer, oc)
	if !strings.Contains(out, "does not belong to your clan") {
		t.Errorf("expected not-same-clan clan variant; got %q", out)
	}
}

func TestDoOutcast_NotSameClan_Order(t *testing.T) {
	_, _, officer, oc, victim, vc, _ := setupOutcastWorld(t, types.CLAN_ORDER)
	defer oc.Close()
	defer vc.Close()
	other := &types.ClanData{Name: "Other"}
	victim.PCData.Clan = other
	DoOutcast(officer, "bob")
	out := readOutput(officer, oc)
	if !strings.Contains(out, "does not belong to your order") {
		t.Errorf("expected not-same-clan order variant; got %q", out)
	}
}

func TestDoOutcast_NotSameClan_Guild(t *testing.T) {
	_, _, officer, oc, victim, vc, _ := setupOutcastWorld(t, types.CLAN_GUILD)
	defer oc.Close()
	defer vc.Close()
	other := &types.ClanData{Name: "Other"}
	victim.PCData.Clan = other
	DoOutcast(officer, "bob")
	out := readOutput(officer, oc)
	if !strings.Contains(out, "does not belong to your guild") {
		t.Errorf("expected not-same-clan guild variant; got %q", out)
	}
}

// --- DoOutcast happy path ---

func TestDoOutcast_Success_ClearsFields(t *testing.T) {
	_, _, officer, oc, victim, vc, clan := setupOutcastWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	DoOutcast(officer, "bob")
	if victim.PCData.Clan != nil {
		t.Errorf("Clan should be nil after outcast; got %v", victim.PCData.Clan)
	}
	if victim.PCData.ClanName != "" {
		t.Errorf("ClanName should be empty; got %q", victim.PCData.ClanName)
	}
	if clan.Members != 1 {
		t.Errorf("Members should decrement to 1; got %d", clan.Members)
	}
	if uint32(victim.Speaks)&types.LANG_CLAN != 0 {
		t.Error("LANG_CLAN should be cleared on Speaks")
	}
	if victim.Speaking != int(types.LANG_COMMON) {
		t.Errorf("Speaking should reset to LANG_COMMON; got %d", victim.Speaking)
	}
}

func TestDoOutcast_Success_SkillsForgotten_PkillClan(t *testing.T) {
	w, _, officer, oc, victim, vc, clan := setupOutcastWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	sk := &types.SkillType{Name: "skill-to-forget", Guild: 2}
	w.Skills = []*types.SkillType{sk}
	victim.PCData.Learned[0] = 50
	_ = clan
	DoOutcast(officer, "bob")
	if victim.PCData.Learned[0] != 0 {
		t.Errorf("Learned[0] should be zeroed after outcast from pkill clan; got %d", victim.PCData.Learned[0])
	}
}

func TestDoOutcast_Success_SkillsKept_GuildClan(t *testing.T) {
	w, _, officer, oc, victim, vc, clan := setupOutcastWorld(t, types.CLAN_GUILD)
	defer oc.Close()
	defer vc.Close()
	sk := &types.SkillType{Name: "skill-to-keep", Guild: 2}
	w.Skills = []*types.SkillType{sk}
	victim.PCData.Learned[0] = 50
	victim.Class = clan.Class // pass class-match (but that's for induct; outcast has no class check)
	DoOutcast(officer, "bob")
	if victim.PCData.Learned[0] != 50 {
		t.Errorf("Learned[0] should survive guild outcast; got %d", victim.PCData.Learned[0])
	}
}

func TestDoOutcast_Success_BlanksNumber1(t *testing.T) {
	_, _, officer, oc, victim, vc, clan := setupOutcastWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	clan.Number1 = victim.Name
	DoOutcast(officer, "bob")
	if clan.Number1 != "" {
		t.Errorf("clan.Number1 should be blanked; got %q", clan.Number1)
	}
}

func TestDoOutcast_Success_BlanksNumber2(t *testing.T) {
	_, _, officer, oc, victim, vc, clan := setupOutcastWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	clan.Number2 = victim.Name
	DoOutcast(officer, "bob")
	if clan.Number2 != "" {
		t.Errorf("clan.Number2 should be blanked; got %q", clan.Number2)
	}
}

func TestDoOutcast_Success_LinkdeadVictim(t *testing.T) {
	_, _, officer, oc, victim, vc, _ := setupOutcastWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	victim.Desc = nil
	// Must not panic; also must still perform all field mutations.
	DoOutcast(officer, "bob")
	if victim.PCData.Clan != nil {
		t.Error("linkdead victim should still be outcasted")
	}
}

func TestDoOutcast_Success_EchoPKersFires_PlainClan(t *testing.T) {
	w, room, officer, oc, _, vc, _ := setupOutcastWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	// Seed a pkill observer.
	pk, pkc := makeTestChar("Pkay")
	defer pkc.Close()
	pk.PCData.Flags = int(types.PCFLAG_DEADLY)
	pk.InRoom = room
	w.Descriptors = append(w.Descriptors, pk.Desc)
	DoOutcast(officer, "bob")
	out := readOutput(pk, pkc)
	if !strings.Contains(out, "has been outcast from") {
		t.Errorf("pkill observer expected broadcast; got %q", out)
	}
}

func TestDoOutcast_Success_EchoPKersSuppressed_GuildClan(t *testing.T) {
	w, _, officer, oc, _, vc, _ := setupOutcastWorld(t, types.CLAN_GUILD)
	defer oc.Close()
	defer vc.Close()
	pk, pkc := makeTestChar("Pkay")
	defer pkc.Close()
	pk.PCData.Flags = int(types.PCFLAG_DEADLY)
	w.Descriptors = append(w.Descriptors, pk.Desc)
	DoOutcast(officer, "bob")
	out := readOutput(pk, pkc)
	if strings.Contains(out, "has been outcast") {
		t.Errorf("guild outcast should not broadcast to PKers; got %q", out)
	}
}

func TestDoOutcast_Success_EchoPKersSuppressed_OrderClan(t *testing.T) {
	w, _, officer, oc, _, vc, _ := setupOutcastWorld(t, types.CLAN_ORDER)
	defer oc.Close()
	defer vc.Close()
	pk, pkc := makeTestChar("Pkay")
	defer pkc.Close()
	pk.PCData.Flags = int(types.PCFLAG_DEADLY)
	w.Descriptors = append(w.Descriptors, pk.Desc)
	DoOutcast(officer, "bob")
	out := readOutput(pk, pkc)
	if strings.Contains(out, "has been outcast") {
		t.Errorf("order outcast should not broadcast to PKers; got %q", out)
	}
}

func TestDoOutcast_Success_Messages(t *testing.T) {
	_, _, officer, oc, victim, vc, _ := setupOutcastWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	DoOutcast(officer, "bob")
	offOut := readOutput(officer, oc)
	vicOut := readOutput(victim, vc)
	if !strings.Contains(offOut, "You outcast Bob from Test Clan") {
		t.Errorf("officer message mismatch; got %q", offOut)
	}
	if !strings.Contains(vicOut, "Alice outcasts you from Test Clan") {
		t.Errorf("victim message mismatch; got %q", vicOut)
	}
}

func TestDoOutcast_Persists_SaveFuncAndClanFile(t *testing.T) {
	_, _, officer, oc, _, vc, clan := setupOutcastWorld(t, 0)
	defer oc.Close()
	defer vc.Close()
	var saved bool
	prev := SaveFunc
	SaveFunc = func(c *types.CharData) { saved = true }
	defer func() { SaveFunc = prev }()

	prevDir := ClanDir
	ClanDir = t.TempDir()
	defer func() { ClanDir = prevDir }()

	DoOutcast(officer, "bob")
	if !saved {
		t.Error("SaveFunc should fire on outcast")
	}
	if _, err := os.Stat(filepath.Join(ClanDir, clan.Filename)); err != nil {
		t.Errorf("expected SaveClanFile write: %v", err)
	}
}

// --- DoBestow ---

func setupBestowWorld(t *testing.T) (imm *types.CharData, ic net.Conn,
	victim *types.CharData, vc net.Conn, w *world.World) {
	t.Helper()
	w = world.New("/tmp/test")
	WorldRef = w
	imm, ic = makeTestChar("Godd")
	imm.Level = types.LEVEL_IMMORTAL
	victim, vc = makeTestChar("Bob")
	victim.Level = 30
	// Place both into a room and add to world for GetCharWorld.
	room := &types.RoomIndexData{Vnum: 1, Name: "Temple"}
	imm.InRoom = room
	victim.InRoom = room
	room.People = append(room.People, imm, victim)
	w.Characters = append(w.Characters, imm, victim)
	return imm, ic, victim, vc, w
}

func TestDoBestow_SyntaxEmpty(t *testing.T) {
	imm, ic, _, vc, _ := setupBestowWorld(t)
	defer ic.Close()
	defer vc.Close()
	DoBestow(imm, "")
	out := readOutput(imm, ic)
	if !strings.Contains(out, "Bestow whom with what?") {
		t.Errorf("expected syntax; got %q", out)
	}
}

func TestDoBestow_VictimNotFound(t *testing.T) {
	imm, ic, _, vc, _ := setupBestowWorld(t)
	defer ic.Close()
	defer vc.Close()
	DoBestow(imm, "ghost")
	out := readOutput(imm, ic)
	if !strings.Contains(out, "They aren't here.") {
		t.Errorf("expected not-here; got %q", out)
	}
}

func TestDoBestow_VictimIsNPC(t *testing.T) {
	imm, ic, _, vc, w := setupBestowWorld(t)
	defer ic.Close()
	defer vc.Close()
	mob := &types.CharData{Name: "guard", ShortDescr: "a guard", InRoom: imm.InRoom}
	mob.Act.Set(types.ACT_IS_NPC)
	imm.InRoom.People = append(imm.InRoom.People, mob)
	w.Characters = append(w.Characters, mob)
	DoBestow(imm, "guard stuff")
	out := readOutput(imm, ic)
	if !strings.Contains(out, "give special abilities to a mob") {
		t.Errorf("expected NPC rejection; got %q", out)
	}
}

func TestDoBestow_VictimOutranksImm(t *testing.T) {
	imm, ic, victim, vc, _ := setupBestowWorld(t)
	defer ic.Close()
	defer vc.Close()
	victim.Trust = types.LEVEL_IMMORTAL + 10
	DoBestow(imm, "bob induct")
	out := readOutput(imm, ic)
	if !strings.Contains(out, "powerful enough") {
		t.Errorf("expected trust rejection; got %q", out)
	}
}

func TestDoBestow_ListNoArg(t *testing.T) {
	imm, ic, victim, vc, _ := setupBestowWorld(t)
	defer ic.Close()
	defer vc.Close()
	victim.PCData.Bestowments = " induct"
	DoBestow(imm, "bob")
	out := readOutput(imm, ic)
	if !strings.Contains(out, "Current bestowed commands on Bob") {
		t.Errorf("expected list output; got %q", out)
	}
}

func TestDoBestow_ListExplicit(t *testing.T) {
	imm, ic, victim, vc, _ := setupBestowWorld(t)
	defer ic.Close()
	defer vc.Close()
	victim.PCData.Bestowments = " induct"
	DoBestow(imm, "bob list")
	out := readOutput(imm, ic)
	if !strings.Contains(out, "Current bestowed commands on Bob") {
		t.Errorf("expected list output; got %q", out)
	}
	if !strings.Contains(out, "induct") {
		t.Errorf("list should include bestowments content; got %q", out)
	}
}

func TestDoBestow_None(t *testing.T) {
	imm, ic, victim, vc, _ := setupBestowWorld(t)
	defer ic.Close()
	defer vc.Close()
	victim.PCData.Bestowments = " induct outcast"
	DoBestow(imm, "bob none")
	if victim.PCData.Bestowments != "" {
		t.Errorf("Bestowments should be cleared; got %q", victim.PCData.Bestowments)
	}
	immOut := readOutput(imm, ic)
	if !strings.Contains(immOut, "Bestowments removed from Bob") {
		t.Errorf("imm should see removed confirmation; got %q", immOut)
	}
	vicOut := readOutput(victim, vc)
	if !strings.Contains(vicOut, "has removed your bestowed commands") {
		t.Errorf("victim should see notification; got %q", vicOut)
	}
}

func TestDoBestow_Append(t *testing.T) {
	imm, ic, victim, vc, _ := setupBestowWorld(t)
	defer ic.Close()
	defer vc.Close()
	DoBestow(imm, "bob induct")
	// C's leading-space quirk: "" + " " + "induct" = " induct".
	if victim.PCData.Bestowments != " induct" {
		t.Errorf("Bestowments = %q, want \" induct\" (C leading-space quirk)", victim.PCData.Bestowments)
	}
}

func TestDoBestow_AppendMultiple(t *testing.T) {
	imm, ic, victim, vc, _ := setupBestowWorld(t)
	defer ic.Close()
	defer vc.Close()
	DoBestow(imm, "bob induct")
	DoBestow(imm, "bob outcast")
	if !strings.Contains(victim.PCData.Bestowments, "induct") {
		t.Errorf("Bestowments should retain induct; got %q", victim.PCData.Bestowments)
	}
	if !strings.Contains(victim.PCData.Bestowments, "outcast") {
		t.Errorf("Bestowments should append outcast; got %q", victim.PCData.Bestowments)
	}
}

func TestDoBestow_IsNameDetectsLeadingSpace(t *testing.T) {
	imm, ic, victim, vc, _ := setupBestowWorld(t)
	defer ic.Close()
	defer vc.Close()
	DoBestow(imm, "bob induct")
	// Verify util.IsName (used by isClanOfficer) sees the keyword through
	// the leading-space quirk.
	if !util.IsName("induct", victim.PCData.Bestowments) {
		t.Errorf("util.IsName must parse leading-space bestowments; got %q",
			victim.PCData.Bestowments)
	}
}

func TestDoBestow_DoneMessage(t *testing.T) {
	imm, ic, _, vc, _ := setupBestowWorld(t)
	defer ic.Close()
	defer vc.Close()
	DoBestow(imm, "bob induct")
	out := readOutput(imm, ic)
	if !strings.Contains(out, "Done.") {
		t.Errorf("expected 'Done.' on append; got %q", out)
	}
}
