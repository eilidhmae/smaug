package boot_test

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/eilidhmae/smaug/internal/act"
	"github.com/eilidhmae/smaug/internal/boot"
	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/game"
	"github.com/eilidhmae/smaug/internal/mudprog"
	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func parentCtx() context.Context {
	return context.Background()
}

// testDataDir is the relative path from internal/boot/ to the cmd/smaug testdata.
const testDataDir = "../../cmd/smaug/testdata"

func makeIncoming() chan *types.DescriptorData {
	return make(chan *types.DescriptorData, 8)
}

func TestBoot_LoadsAreaData(t *testing.T) {
	// Clean stray player saves from previous tests.
	_ = os.RemoveAll(filepath.Join(testDataDir, "player"))

	w := world.New(testDataDir)
	incoming := makeIncoming()

	_, _, err := boot.Boot(w, testDataDir, incoming, boot.ProductionOpts())
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}

	if len(w.Rooms) < 1 {
		t.Errorf("expected at least 1 room, got %d", len(w.Rooms))
	}
	if len(w.Classes) < 1 {
		t.Errorf("expected at least 1 class, got %d", len(w.Classes))
	}
	if len(w.Races) < 1 {
		t.Errorf("expected at least 1 race, got %d", len(w.Races))
	}
	// MobIndex/ObjIndex/Skills can legitimately be 0 in testdata — just make
	// sure the fields are reachable (nil-map check is implicit).
	_ = w.MobIndex
	_ = w.ObjIndex
	_ = w.Skills
}

func TestBoot_RegistersCommands(t *testing.T) {
	_ = os.RemoveAll(filepath.Join(testDataDir, "player"))
	w := world.New(testDataDir)
	incoming := makeIncoming()

	reg, _, err := boot.Boot(w, testDataDir, incoming, boot.ProductionOpts())
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}

	if reg == nil {
		t.Fatal("Boot returned nil registry")
	}
	// Use maxTrust so imm-only commands (redit) resolve.
	const maxTrust = 65535
	if reg.Find("look", maxTrust) == nil {
		t.Error("expected 'look' to be registered")
	}
	if reg.Find("cast", maxTrust) == nil {
		t.Error("expected 'cast' to be registered")
	}
	if reg.Find("redit", maxTrust) == nil {
		t.Error("expected 'redit' to be registered")
	}
	if got := len(reg.All()); got < 150 {
		t.Errorf("expected at least 150 commands, got %d", got)
	}
}

// TestBoot_WiresOlcMenuSeams pins the three OLC editor-menu seams (redit,
// oedit, medit) to their game-package backing functions after Boot. A
// mutation that breaks any of the three seam assignments fails this test
// (mutation gate M-boot-wires). Plan-phase6-olc-medit.md §G1 Wave-2 boot
// wire: act.MeditDispMenuFunc = game.MeditDispMenu.
func TestBoot_WiresOlcMenuSeams(t *testing.T) {
	_ = os.RemoveAll(filepath.Join(testDataDir, "player"))
	w := world.New(testDataDir)
	incoming := makeIncoming()
	if _, _, err := boot.Boot(w, testDataDir, incoming, boot.ProductionOpts()); err != nil {
		t.Fatalf("Boot: %v", err)
	}
	if act.ReditDispMenuFunc == nil {
		t.Error("act.ReditDispMenuFunc == nil after Boot")
	}
	if act.OeditDispMenuFunc == nil {
		t.Error("act.OeditDispMenuFunc == nil after Boot")
	}
	if act.MeditDispMenuFunc == nil {
		t.Error("act.MeditDispMenuFunc == nil after Boot (Wave-2 wire missing)")
	}
}

// TestBoot_McEditorsRegistered pins mpedit/opedit/rpedit registration at
// LEVEL_IMMORTAL. Plan-phase6-olc-mpedit.md §G11 acceptance.
func TestBoot_McEditorsRegistered(t *testing.T) {
	_ = os.RemoveAll(filepath.Join(testDataDir, "player"))
	w := world.New(testDataDir)
	incoming := makeIncoming()
	reg, _, err := boot.Boot(w, testDataDir, incoming, boot.ProductionOpts())
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}
	const maxTrust = 65535
	for _, name := range []string{"mpedit", "opedit", "rpedit"} {
		c := reg.Find(name, maxTrust)
		if c == nil {
			t.Errorf("expected %q to be registered", name)
			continue
		}
		if c.DoFun == nil {
			t.Errorf("%q.DoFun must be non-nil", name)
		}
	}
}

// TestBoot_FoldareaRegistered pins foldarea/unfoldarea registration at
// LEVEL_IMMORTAL with POS_DEAD. Plan plan-phase6-foldarea.md §G6 / §A11.
func TestBoot_FoldareaRegistered(t *testing.T) {
	_ = os.RemoveAll(filepath.Join(testDataDir, "player"))
	w := world.New(testDataDir)
	incoming := makeIncoming()
	reg, _, err := boot.Boot(w, testDataDir, incoming, boot.ProductionOpts())
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}
	const maxTrust = 65535
	for _, name := range []string{"foldarea", "unfoldarea"} {
		c := reg.Find(name, maxTrust)
		if c == nil {
			t.Errorf("expected %q to be registered", name)
			continue
		}
		if c.DoFun == nil {
			t.Errorf("%q.DoFun must be non-nil", name)
		}
		if c.Level != types.LEVEL_IMMORTAL {
			t.Errorf("%q level: got %d, want LEVEL_IMMORTAL (%d)", name, c.Level, types.LEVEL_IMMORTAL)
		}
		if c.Position != types.POS_DEAD {
			t.Errorf("%q position: got %d, want POS_DEAD (%d)", name, c.Position, types.POS_DEAD)
		}
	}
	// Trust gate: a low-trust char should not see foldarea via Find.
	if c := reg.Find("foldarea", 0); c != nil {
		t.Errorf("foldarea must be hidden from trust=0; got %v", c)
	}
}

// TestBoot_PcrenameSeamsWired pins the D4 boot wirings: RenamePlayerFileFunc
// (act seam → persist.RenamePlayerFile) + game.PcrenameFunc (game seam →
// act.DoPcrename). Plan plan-phase6-quickwins-blank-pcrename.md §G8 / §A16.
func TestBoot_PcrenameSeamsWired(t *testing.T) {
	_ = os.RemoveAll(filepath.Join(testDataDir, "player"))
	w := world.New(testDataDir)
	incoming := makeIncoming()
	_, _, err := boot.Boot(w, testDataDir, incoming, boot.ProductionOpts())
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}
	if act.RenamePlayerFileFunc == nil {
		t.Error("act.RenamePlayerFileFunc must be wired after boot")
	}
	if game.PcrenameFunc == nil {
		t.Error("game.PcrenameFunc must be wired after boot")
	}
}

// TestBoot_BlankRegistered pins blank command registration at level 0
// / POS_DEAD. Plan plan-phase6-quickwins-blank-pcrename.md §G2 / §A3.
func TestBoot_BlankRegistered(t *testing.T) {
	_ = os.RemoveAll(filepath.Join(testDataDir, "player"))
	w := world.New(testDataDir)
	incoming := makeIncoming()
	reg, _, err := boot.Boot(w, testDataDir, incoming, boot.ProductionOpts())
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}
	c := reg.Find("blank", 0)
	if c == nil {
		t.Fatalf("expected %q to be registered", "blank")
	}
	if c.DoFun == nil {
		t.Errorf("blank.DoFun must be non-nil")
	}
	if c.Level != 0 {
		t.Errorf("blank level: got %d, want 0", c.Level)
	}
	if c.Position != types.POS_DEAD {
		t.Errorf("blank position: got %d, want POS_DEAD (%d)", c.Position, types.POS_DEAD)
	}
}

// TestBoot_ClanOfficerRegistered pins induct/outcast/bestow command registration.
// induct and outcast carry Level 0 (authority is in-command via isClanOfficer);
// bestow is LEVEL_IMMORTAL. See plan-phase6-clan-officer.md §D5.
func TestBoot_ClanOfficerRegistered(t *testing.T) {
	_ = os.RemoveAll(filepath.Join(testDataDir, "player"))
	w := world.New(testDataDir)
	incoming := makeIncoming()
	reg, _, err := boot.Boot(w, testDataDir, incoming, boot.ProductionOpts())
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}
	const maxTrust = 65535

	induct := reg.Find("induct", maxTrust)
	if induct == nil {
		t.Fatal("induct must be registered")
	}
	if induct.Level != 0 {
		t.Errorf("induct.Level = %d, want 0", induct.Level)
	}
	if induct.DoFun == nil {
		t.Error("induct.DoFun must be non-nil")
	}

	outcast := reg.Find("outcast", maxTrust)
	if outcast == nil {
		t.Fatal("outcast must be registered")
	}
	if outcast.Level != 0 {
		t.Errorf("outcast.Level = %d, want 0", outcast.Level)
	}

	bestow := reg.Find("bestow", maxTrust)
	if bestow == nil {
		t.Fatal("bestow must be registered")
	}
	if bestow.Level != types.LEVEL_IMMORTAL {
		t.Errorf("bestow.Level = %d, want %d (LEVEL_IMMORTAL)",
			bestow.Level, types.LEVEL_IMMORTAL)
	}

	// ClanDir must be exported & set to <dataDir>/clans.
	wantClanDir := filepath.Join(testDataDir, "clans")
	if act.ClanDir != wantClanDir {
		t.Errorf("act.ClanDir = %q, want %q", act.ClanDir, wantClanDir)
	}
}

// TestBoot_StancesOLCRegistered pins ststat/stset command registration
// per plan-phase6-stances-olc.md A16. Both carry LEVEL_IMMORTAL and
// POS_DEAD; `stance` (already registered elsewhere) remains Level 0.
func TestBoot_StancesOLCRegistered(t *testing.T) {
	_ = os.RemoveAll(filepath.Join(testDataDir, "player"))
	w := world.New(testDataDir)
	incoming := makeIncoming()
	reg, _, err := boot.Boot(w, testDataDir, incoming, boot.ProductionOpts())
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}
	const maxTrust = 65535

	ststat := reg.Find("ststat", maxTrust)
	if ststat == nil {
		t.Fatal("ststat must be registered")
	}
	if ststat.Level != types.LEVEL_IMMORTAL {
		t.Errorf("ststat.Level = %d, want %d (LEVEL_IMMORTAL)",
			ststat.Level, types.LEVEL_IMMORTAL)
	}
	if ststat.DoFun == nil {
		t.Error("ststat.DoFun must be non-nil")
	}

	stset := reg.Find("stset", maxTrust)
	if stset == nil {
		t.Fatal("stset must be registered")
	}
	if stset.Level != types.LEVEL_IMMORTAL {
		t.Errorf("stset.Level = %d, want %d (LEVEL_IMMORTAL)",
			stset.Level, types.LEVEL_IMMORTAL)
	}

	stance := reg.Find("stance", maxTrust)
	if stance == nil {
		t.Fatal("stance must be registered")
	}
	if stance.Level != 0 {
		t.Errorf("stance.Level = %d, want 0 (player-accessible)", stance.Level)
	}
}

// TestBoot_ArcheryRegistered pins draw/fire/dislodge command registration
// per plan-phase6-archery.md G6. All three are player-accessible (Level 0).
// `fire` pre-existed at POS_FIGHTING; `draw` uses POS_FIGHTING (combat
// draw is allowed); `dislodge` uses POS_RESTING (self-surgery, sitting OK).
func TestBoot_ArcheryRegistered(t *testing.T) {
	_ = os.RemoveAll(filepath.Join(testDataDir, "player"))
	w := world.New(testDataDir)
	incoming := makeIncoming()
	reg, _, err := boot.Boot(w, testDataDir, incoming, boot.ProductionOpts())
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}
	const maxTrust = 65535

	draw := reg.Find("draw", maxTrust)
	if draw == nil {
		t.Fatal("draw must be registered")
	}
	if draw.Level != 0 {
		t.Errorf("draw.Level = %d, want 0", draw.Level)
	}
	if draw.DoFun == nil {
		t.Error("draw.DoFun must be non-nil")
	}

	fire := reg.Find("fire", maxTrust)
	if fire == nil {
		t.Fatal("fire must be registered")
	}
	if fire.Level != 0 {
		t.Errorf("fire.Level = %d, want 0", fire.Level)
	}
	if fire.DoFun == nil {
		t.Error("fire.DoFun must be non-nil")
	}

	dislodge := reg.Find("dislodge", maxTrust)
	if dislodge == nil {
		t.Fatal("dislodge must be registered")
	}
	if dislodge.Level != 0 {
		t.Errorf("dislodge.Level = %d, want 0", dislodge.Level)
	}
	if dislodge.DoFun == nil {
		t.Error("dislodge.DoFun must be non-nil")
	}
}

func TestBoot_WiresCallbacks(t *testing.T) {
	_ = os.RemoveAll(filepath.Join(testDataDir, "player"))
	// Reset globals we care about.
	act.WorldRef = nil
	act.CmdRegistry = nil
	act.SaveFunc = nil
	act.StartEditingFunc = nil
	act.CopyBufferFunc = nil
	act.StopEditingFunc = nil
	game.PromptExpBase = nil
	act.ShutdownFunc = nil
	act.DisconnectFunc = nil
	mudprog.WorldRef = nil
	mudprog.CmdRegistry = nil
	combat.WorldRef = nil
	combat.HitprcntHook = nil
	combat.VoidHook = nil
	combat.ObjDamageHook = nil
	combat.RfightHook = nil
	combat.DeathRoomHook = nil
	combat.CanUseSkillHook = nil
	combat.LearnFromSuccessHook = nil
	combat.LearnFromFailureHook = nil
	combat.LookupSkillSlotHook = nil
	combat.ArenaIsBusyFunc = nil
	combat.DoLookFunc = nil
	persist.SkillNameLookup = nil
	persist.SkillGetter = nil

	// Reset + restore BcryptCost so this test doesn't depend on, or leak
	// into, other tests' global state. Production Boot must not mutate it.
	prevCost := game.BcryptCost
	defer func() { game.BcryptCost = prevCost }()
	game.BcryptCost = bcrypt.DefaultCost

	w := world.New(testDataDir)
	incoming := makeIncoming()

	reg, _, err := boot.Boot(w, testDataDir, incoming, boot.ProductionOpts())
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}

	if game.BcryptCost != bcrypt.DefaultCost {
		t.Errorf("ProductionOpts must not mutate game.BcryptCost: got %d want %d",
			game.BcryptCost, bcrypt.DefaultCost)
	}

	if act.WorldRef != w {
		t.Error("act.WorldRef not wired")
	}
	if act.CmdRegistry == nil {
		t.Error("act.CmdRegistry not wired")
	}
	if act.SaveFunc == nil {
		t.Error("act.SaveFunc not wired")
	}
	if act.StartEditingFunc == nil {
		t.Error("act.StartEditingFunc not wired")
	}
	if act.CopyBufferFunc == nil {
		t.Error("act.CopyBufferFunc not wired")
	}
	if act.StopEditingFunc == nil {
		t.Error("act.StopEditingFunc not wired")
	}
	if game.PromptExpBase == nil {
		t.Error("game.PromptExpBase not wired")
	}
	// ShutdownFunc should be wired to the ProductionOpts default
	// (save-and-exit). Do NOT invoke it — it calls os.Exit.
	if act.ShutdownFunc == nil {
		t.Error("act.ShutdownFunc not wired")
	}
	if act.DisconnectFunc == nil {
		t.Error("act.DisconnectFunc not wired")
	}
	if mudprog.WorldRef != w {
		t.Error("mudprog.WorldRef not wired")
	}
	if mudprog.CmdRegistry == nil {
		t.Error("mudprog.CmdRegistry not wired")
	}
	if combat.WorldRef != w {
		t.Error("combat.WorldRef not wired")
	}
	if combat.HitprcntHook == nil {
		t.Error("combat.HitprcntHook not wired")
	}
	if combat.VoidHook == nil {
		t.Error("combat.VoidHook not wired")
	}
	if combat.ObjDamageHook == nil {
		t.Error("combat.ObjDamageHook not wired")
	}
	if combat.RfightHook == nil {
		t.Error("combat.RfightHook not wired")
	}
	if combat.DeathRoomHook == nil {
		t.Error("combat.DeathRoomHook not wired")
	}
	if combat.CanUseSkillHook == nil {
		t.Error("combat.CanUseSkillHook not wired")
	}
	if combat.LearnFromSuccessHook == nil {
		t.Error("combat.LearnFromSuccessHook not wired")
	}
	if combat.LearnFromFailureHook == nil {
		t.Error("combat.LearnFromFailureHook not wired")
	}
	if combat.LookupSkillSlotHook == nil {
		t.Error("combat.LookupSkillSlotHook not wired")
	}
	// Arena seams (plan-phase6-arena.md §G7).
	if combat.ArenaIsBusyFunc == nil {
		t.Error("combat.ArenaIsBusyFunc not wired")
	}
	if combat.DoLookFunc == nil {
		t.Error("combat.DoLookFunc not wired")
	}
	if persist.SkillNameLookup == nil {
		t.Error("persist.SkillNameLookup not wired")
	}
	if persist.SkillGetter == nil {
		t.Error("persist.SkillGetter not wired")
	}
	// Registry-level hooks wired onto the Registry returned by Boot.
	if reg == nil {
		t.Fatal("Boot returned nil registry")
	}
	if reg.SocialFallback == nil {
		t.Error("reg.SocialFallback not wired")
	}
	if reg.ObjCommandHook == nil {
		t.Error("reg.ObjCommandHook not wired")
	}
	if reg.RoomCommandHook == nil {
		t.Error("reg.RoomCommandHook not wired")
	}
}

func TestBoot_TestOptsShutdownRecords(t *testing.T) {
	_ = os.RemoveAll(filepath.Join(testDataDir, "player"))

	w := world.New(testDataDir)
	incoming := makeIncoming()

	opts := boot.TestOpts()
	if opts.Shutdowns == nil {
		t.Fatal("TestOpts must allocate Shutdowns channel")
	}

	_, loop, err := boot.Boot(w, testDataDir, incoming, opts)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}
	if loop == nil {
		t.Fatal("Boot returned nil loop")
	}

	// Fire the wired shutdown hook; should enqueue a ShutdownRequest and
	// cancel the loop's internal ctx.
	act.ShutdownFunc(false)

	select {
	case req := <-opts.Shutdowns:
		if req.Reboot {
			t.Errorf("expected Reboot=false, got true")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("no ShutdownRequest received within 100ms")
	}

	// Loop must be cancelled: calling Run should return promptly.
	done := make(chan struct{})
	go func() {
		// Run with a fresh ctx — the loop's internal ctx should already
		// be cancelled so Run returns immediately regardless.
		ctx := parentCtx()
		loop.Run(ctx)
		close(done)
	}()
	select {
	case <-done:
		// good — loop exited via internal cancel
	case <-time.After(500 * time.Millisecond):
		t.Fatal("loop did not exit after Cancel()")
	}
}

func TestTestOpts_LowersBcryptCost(t *testing.T) {
	// Save and restore — TestOpts mutates a process-global.
	prev := game.BcryptCost
	defer func() { game.BcryptCost = prev }()
	game.BcryptCost = bcrypt.DefaultCost

	_ = boot.TestOpts()
	if game.BcryptCost != bcrypt.MinCost {
		t.Errorf("TestOpts did not lower BcryptCost: got %d want %d",
			game.BcryptCost, bcrypt.MinCost)
	}
}

// TestBoot_SyncsActBcryptCost asserts that Boot propagates game.BcryptCost
// into act.BcryptCost so DoPassword hashes at the same cost the login flow
// does. Without this sync, tests that use TestOpts() to lower bcrypt cost
// would still pay the full DefaultCost inside DoPassword.
func TestBoot_SyncsActBcryptCost(t *testing.T) {
	_ = os.RemoveAll(filepath.Join(testDataDir, "player"))
	prevGame := game.BcryptCost
	prevAct := act.BcryptCost
	defer func() {
		game.BcryptCost = prevGame
		act.BcryptCost = prevAct
	}()

	game.BcryptCost = bcrypt.MinCost
	act.BcryptCost = bcrypt.DefaultCost // force a mismatch so the sync is observable

	w := world.New(testDataDir)
	incoming := makeIncoming()

	_, _, err := boot.Boot(w, testDataDir, incoming, boot.ProductionOpts())
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}

	if act.BcryptCost != bcrypt.MinCost {
		t.Errorf("Boot should sync act.BcryptCost from game.BcryptCost: act=%d game=%d",
			act.BcryptCost, game.BcryptCost)
	}
}

// copyFile copies a single regular file from src to dst, creating
// intermediate directories as needed. Used to assemble scoped-fixture
// data dirs for the fail-loud tests below.
func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read %s: %v", src, err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(dst), err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", dst, err)
	}
}

// makeMinimalDataDir builds a temp data dir under t.TempDir() that
// contains a valid areas subtree copied from the shared testdata fixture,
// plus whichever of classes / races / skills the caller chooses to
// include. Used to drive the "missing foundational subsystem" tests: the
// missing subsystem's source files are simply not copied, so its loader
// will fail when Boot runs.
func makeMinimalDataDir(t *testing.T, includeClasses, includeRaces, includeSkills bool) string {
	t.Helper()
	dst := t.TempDir()

	// Areas are always required — copy the whole dir.
	copyFile(t, filepath.Join(testDataDir, "area", "area.lst"),
		filepath.Join(dst, "area", "area.lst"))
	copyFile(t, filepath.Join(testDataDir, "area", "test_boot.are"),
		filepath.Join(dst, "area", "test_boot.are"))

	if includeClasses {
		copyFile(t, filepath.Join(testDataDir, "classes", "class.lst"),
			filepath.Join(dst, "classes", "class.lst"))
		copyFile(t, filepath.Join(testDataDir, "classes", "Warrior.class"),
			filepath.Join(dst, "classes", "Warrior.class"))
	}
	if includeRaces {
		copyFile(t, filepath.Join(testDataDir, "races", "race.lst"),
			filepath.Join(dst, "races", "race.lst"))
		copyFile(t, filepath.Join(testDataDir, "races", "Human.race"),
			filepath.Join(dst, "races", "Human.race"))
	}
	if includeSkills {
		copyFile(t, filepath.Join(testDataDir, "system", "en", "skills.dat"),
			filepath.Join(dst, "system", "en", "skills.dat"))
	}
	return dst
}

// TestBoot_FailsLoudWhenClassesMissing asserts that Boot returns an error
// (and no usable registry / loop) when the classes directory is missing.
// A misconfigured -data dir must not silently produce a server that then
// fails character creation at runtime.
func TestBoot_FailsLoudWhenClassesMissing(t *testing.T) {
	dir := makeMinimalDataDir(t, false /*classes*/, true /*races*/, true /*skills*/)
	w := world.New(dir)

	reg, loop, err := boot.Boot(w, dir, makeIncoming(), boot.ProductionOpts())
	if err == nil {
		t.Fatal("Boot should fail when classes are missing, got nil error")
	}
	if !regexp.MustCompile(`(?i)class`).MatchString(err.Error()) {
		t.Errorf("error should mention classes, got: %v", err)
	}
	if reg != nil {
		t.Error("Boot should return nil registry on failure")
	}
	if loop != nil {
		t.Error("Boot should return nil loop on failure")
	}
}

// TestBoot_FailsLoudWhenRacesMissing asserts that Boot returns an error
// when the races directory is missing. Same rationale as the classes test.
func TestBoot_FailsLoudWhenRacesMissing(t *testing.T) {
	dir := makeMinimalDataDir(t, true /*classes*/, false /*races*/, true /*skills*/)
	w := world.New(dir)

	reg, loop, err := boot.Boot(w, dir, makeIncoming(), boot.ProductionOpts())
	if err == nil {
		t.Fatal("Boot should fail when races are missing, got nil error")
	}
	if !regexp.MustCompile(`(?i)race`).MatchString(err.Error()) {
		t.Errorf("error should mention races, got: %v", err)
	}
	if reg != nil {
		t.Error("Boot should return nil registry on failure")
	}
	if loop != nil {
		t.Error("Boot should return nil loop on failure")
	}
}

// TestBoot_FailsLoudWhenSkillsMissing asserts that Boot returns an error
// when system/en/skills.dat is missing. Same rationale as above.
func TestBoot_FailsLoudWhenSkillsMissing(t *testing.T) {
	dir := makeMinimalDataDir(t, true /*classes*/, true /*races*/, false /*skills*/)
	w := world.New(dir)

	reg, loop, err := boot.Boot(w, dir, makeIncoming(), boot.ProductionOpts())
	if err == nil {
		t.Fatal("Boot should fail when skills are missing, got nil error")
	}
	if !regexp.MustCompile(`(?i)skill`).MatchString(err.Error()) {
		t.Errorf("error should mention skills, got: %v", err)
	}
	if reg != nil {
		t.Error("Boot should return nil registry on failure")
	}
	if loop != nil {
		t.Error("Boot should return nil loop on failure")
	}
}

func TestMainGoHasNoCallbackWires(t *testing.T) {
	data, err := os.ReadFile("../../cmd/smaug/main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	src := string(data)

	// Strip // line comments so matches inside comments are not flagged.
	lineComment := regexp.MustCompile(`(?m)//.*$`)
	src = lineComment.ReplaceAllString(src, "")

	// Also strip /* ... */ block comments (minimal, no nesting).
	blockComment := regexp.MustCompile(`(?s)/\*.*?\*/`)
	src = blockComment.ReplaceAllString(src, "")

	forbidden := []string{
		`act\.WorldRef\s*=`,
		`act\.SaveFunc\s*=`,
		`act\.CmdRegistry\s*=`,
		`act\.StartEditingFunc\s*=`,
		`act\.ShutdownFunc\s*=`,
		`act\.DisconnectFunc\s*=`,
		`mudprog\.CmdRegistry\s*=`,
		`mudprog\.WorldRef\s*=`,
		`combat\.WorldRef\s*=`,
		`combat\.HitprcntHook\s*=`,
		`combat\.VoidHook\s*=`,
		`combat\.ObjDamageHook\s*=`,
		`combat\.RfightHook\s*=`,
		`combat\.DeathRoomHook\s*=`,
		`\.SocialFallback\s*=`,
		`\.ObjCommandHook\s*=`,
		`\.RoomCommandHook\s*=`,
		`persist\.SkillNameLookup\s*=`,
		`persist\.SkillGetter\s*=`,
	}
	for _, pat := range forbidden {
		re := regexp.MustCompile(pat)
		if loc := re.FindStringIndex(src); loc != nil {
			t.Errorf("forbidden callback wire %q found in main.go at offset %d", pat, loc[0])
		}
	}
}

// --- Phase 6 Planes integration pins (plan-phase6-planes.md G3) --------

func TestBoot_RegistersPlaneCommands(t *testing.T) {
	_ = os.RemoveAll(filepath.Join(testDataDir, "player"))
	w := world.New(testDataDir)
	incoming := makeIncoming()

	reg, _, err := boot.Boot(w, testDataDir, incoming, boot.ProductionOpts())
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}
	// plist must resolve at mortal trust (Level 0).
	if reg.Find("plist", 0) == nil {
		t.Error("expected 'plist' to be registered at Level 0")
	}
	// pstat is immortal-only.
	if reg.Find("pstat", types.LEVEL_IMMORTAL) == nil {
		t.Error("expected 'pstat' to resolve at LEVEL_IMMORTAL")
	}
	if reg.Find("pstat", 0) != nil {
		t.Error("pstat should NOT resolve at mortal trust=0")
	}
	// pset is LEVEL_GREATER.
	if reg.Find("pset", types.LEVEL_GREATER) == nil {
		t.Error("expected 'pset' to resolve at LEVEL_GREATER")
	}
	if reg.Find("pset", types.LEVEL_IMMORTAL) != nil {
		t.Error("pset should NOT resolve at LEVEL_IMMORTAL (below LEVEL_GREATER)")
	}
}

func TestBoot_LoadsPlanesAndSeedsPrimeMaterial(t *testing.T) {
	_ = os.RemoveAll(filepath.Join(testDataDir, "player"))
	w := world.New(testDataDir)
	incoming := makeIncoming()

	_, _, err := boot.Boot(w, testDataDir, incoming, boot.ProductionOpts())
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}
	if len(w.Planes) != 1 {
		t.Fatalf("expected 1 plane after boot (Prime Material fallback); got %d", len(w.Planes))
	}
	if w.Planes[0].Name != "Prime Material" {
		t.Errorf("expected 'Prime Material'; got %q", w.Planes[0].Name)
	}
}

func TestBoot_AssignsEveryRoomAPlane(t *testing.T) {
	_ = os.RemoveAll(filepath.Join(testDataDir, "player"))
	w := world.New(testDataDir)
	incoming := makeIncoming()

	_, _, err := boot.Boot(w, testDataDir, incoming, boot.ProductionOpts())
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}
	if len(w.Rooms) == 0 {
		t.Skip("no rooms loaded from testdata; nothing to verify")
	}
	for vnum, room := range w.Rooms {
		if room == nil {
			continue
		}
		if room.Plane == nil {
			t.Errorf("room vnum %d has nil Plane after boot", vnum)
		}
	}
}

func TestBoot_WiresPlanesFilePath(t *testing.T) {
	_ = os.RemoveAll(filepath.Join(testDataDir, "player"))
	w := world.New(testDataDir)
	incoming := makeIncoming()

	_, _, err := boot.Boot(w, testDataDir, incoming, boot.ProductionOpts())
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}
	expected := filepath.Join(testDataDir, "system", "planes.dat")
	if act.PlanesFilePath != expected {
		t.Errorf("PlanesFilePath = %q, want %q", act.PlanesFilePath, expected)
	}
}

// TestBoot_RegistersArenaCommands pins the 4 arena commands (plan
// phase6-arena.md §G7). Level 0 / POS_RESTING is the plan's choice —
// matching C, which has no explicit command-level trust gate.
func TestBoot_RegistersArenaCommands(t *testing.T) {
	_ = os.RemoveAll(filepath.Join(testDataDir, "player"))
	w := world.New(testDataDir)
	incoming := makeIncoming()

	reg, _, err := boot.Boot(w, testDataDir, incoming, boot.ProductionOpts())
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}
	const maxTrust = 65535
	for _, name := range []string{"challenge", "accept", "decline", "withdraw"} {
		cmd := reg.Find(name, maxTrust)
		if cmd == nil {
			t.Errorf("expected %q command to be registered", name)
			continue
		}
		if cmd.Level != 0 {
			t.Errorf("%q: Level = %d, want 0", name, cmd.Level)
		}
		if cmd.Position != types.POS_RESTING {
			t.Errorf("%q: Position = %d, want POS_RESTING(%d)",
				name, cmd.Position, types.POS_RESTING)
		}
	}
}

// TestBoot_ArenaRoomsFlaggedInShipped pins the area-data edit from
// plan-phase6-arena.md §Readiness — all arena-range rooms in the
// shipped dev-data area file `db/area/newacad.are` must carry the
// ROOM_ARENA flag (bit 26) so the combat-victory branch can fire.
//
// The boot-test harness uses a separate testdata tree without the
// arena rooms; this test performs a minimal textual scan of the
// shipped file. That keeps the drift-guard cheap (~5ms) and
// independent of any loader-test wiring.
//
// The test accepts values that have bit 26 set. 3145736 (old) lacks
// the bit; 70254600 (old | (1<<26)) has it. A future OLC roundtrip
// that drops the bit would fail the test.
func TestBoot_ArenaRoomsFlaggedInShipped(t *testing.T) {
	path := filepath.Join("..", "..", "..", "db", "area", "newacad.are")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("shipped newacad.are not reachable (%v); skipping drift test", err)
	}
	// The ROOMS section is bounded by "#ROOMS" and "#$"; within it,
	// each room starts with "#<vnum>\n", has header text up to a "~"
	// on its own line, then a flag line of the shape "0 <flags>
	// <sector>". We locate each arena vnum and verify its flag line
	// has bit 26 set.
	section := string(data)
	roomsStart := regexp.MustCompile(`(?m)^#ROOMS$`).FindStringIndex(section)
	if roomsStart == nil {
		t.Fatalf("no #ROOMS marker in %s", path)
	}
	roomsEnd := regexp.MustCompile(`(?m)^#\$`).FindStringIndex(section[roomsStart[1]:])
	if roomsEnd == nil {
		t.Fatalf("no #$ marker after #ROOMS in %s", path)
	}
	body := section[roomsStart[1] : roomsStart[1]+roomsEnd[0]]

	var flagged, total int
	for v := types.ROOM_VNUM_ARENA_MIN; v <= types.ROOM_VNUM_ARENA_MAX; v++ {
		pat := regexp.MustCompile(`(?ms)^#` + itoa(v) + `\n.*?^~\n(0 (\d+) \d+)`)
		m := pat.FindStringSubmatch(body)
		if m == nil {
			continue
		}
		total++
		flags := 0
		for _, c := range m[2] {
			flags = flags*10 + int(c-'0')
		}
		if flags&(1<<types.ROOM_ARENA) != 0 {
			flagged++
		}
	}
	if total == 0 {
		t.Fatalf("no rooms loaded in ROOM_VNUM_ARENA range from %s", path)
	}
	if flagged != total {
		t.Errorf("only %d/%d rooms in arena vnum range %d-%d carry ROOM_ARENA; "+
			"rerun plan-phase6-arena.md area-edit step",
			flagged, total,
			types.ROOM_VNUM_ARENA_MIN, types.ROOM_VNUM_ARENA_MAX)
	}
}

// --- Holidays boot wiring (plan-phase6-holidays.md G4) ---------------

func TestBoot_MissingHolidaysFileIsNonFatal(t *testing.T) {
	// testDataDir has no system/holidays.dat; Boot should succeed,
	// Holidays remains nil/empty, commands still registered, and
	// SysData defaults seeded.
	_ = os.RemoveAll(filepath.Join(testDataDir, "player"))
	_ = os.Remove(filepath.Join(testDataDir, "system", "holidays.dat"))
	w := world.New(testDataDir)
	incoming := makeIncoming()

	reg, _, err := boot.Boot(w, testDataDir, incoming, boot.ProductionOpts())
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}
	if len(w.Holidays) != 0 {
		t.Errorf("expected 0 holidays on missing file; got %d", len(w.Holidays))
	}
	if w.SysData.MaxHoliday != 32 {
		t.Errorf("MaxHoliday = %d, want 32 (default)", w.SysData.MaxHoliday)
	}
	if w.SysData.MonthsPerYear == 0 {
		t.Errorf("MonthsPerYear not defaulted")
	}
	if w.SysData.DaysPerMonth == 0 {
		t.Errorf("DaysPerMonth not defaulted")
	}
	expectedPath := filepath.Join(testDataDir, "system", "holidays.dat")
	if act.HolidayFilePath != expectedPath {
		t.Errorf("HolidayFilePath = %q, want %q", act.HolidayFilePath, expectedPath)
	}
	const maxTrust = 65535
	for _, name := range []string{"holidays", "saveholiday", "setholiday"} {
		if reg.Find(name, maxTrust) == nil {
			t.Errorf("command %q not registered", name)
		}
	}
	// Level pin: holidays is Level 0; the others at LEVEL_ASCENDANT.
	if cmd := reg.Find("holidays", 0); cmd == nil {
		t.Errorf("holidays should be player-visible (Level 0)")
	}
	if cmd := reg.Find("saveholiday", 0); cmd != nil {
		t.Errorf("saveholiday should NOT resolve at trust 0")
	}
	if cmd := reg.Find("saveholiday", types.LEVEL_ASCENDANT); cmd == nil {
		t.Errorf("saveholiday should resolve at LEVEL_ASCENDANT")
	}
	if cmd := reg.Find("setholiday", types.LEVEL_ASCENDANT); cmd == nil {
		t.Errorf("setholiday should resolve at LEVEL_ASCENDANT")
	}
}

func TestBoot_LoadsHolidaysWhenFilePresent(t *testing.T) {
	// Write a scratch holidays.dat into testDataDir/system/, boot,
	// verify the holidays loaded, then remove the scratch file.
	scratchPath := filepath.Join(testDataDir, "system", "holidays.dat")
	// Make sure there's nothing left over from a previous run.
	_ = os.Remove(scratchPath)
	content := "#HOLIDAY\nName\t\tFirst~\nAnnounce\tAlpha.~\nMonth\t\t1\nDay\t\t1\nEnd\n\n#HOLIDAY\nName\t\tSecond~\nAnnounce\tBeta.~\nMonth\t\t12\nDay\t\t30\nEnd\n\n#END\n"
	if err := os.WriteFile(scratchPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write scratch holidays.dat: %v", err)
	}
	defer os.Remove(scratchPath)

	_ = os.RemoveAll(filepath.Join(testDataDir, "player"))
	w := world.New(testDataDir)
	incoming := makeIncoming()

	_, _, err := boot.Boot(w, testDataDir, incoming, boot.ProductionOpts())
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}
	if len(w.Holidays) != 2 {
		t.Fatalf("len(Holidays) = %d, want 2", len(w.Holidays))
	}
	if w.Holidays[0].Name != "First" {
		t.Errorf("Holidays[0].Name = %q, want First", w.Holidays[0].Name)
	}
	if w.Holidays[1].Name != "Second" {
		t.Errorf("Holidays[1].Name = %q, want Second", w.Holidays[1].Name)
	}
}

// itoa formats n as a base-10 ASCII string without importing strconv
// (keeps the drift-test dependency surface tiny).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
