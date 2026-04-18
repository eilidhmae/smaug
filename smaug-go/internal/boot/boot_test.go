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
