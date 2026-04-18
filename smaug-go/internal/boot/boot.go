// Package boot consolidates cross-package boot wiring so production
// (cmd/smaug) and test harnesses (internal/testclient) share one code path
// with no drift. Boot loads game data, registers commands, and wires every
// callback/hook the game's packages need to function. Nothing imports this
// package except entry points and tests.
package boot

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/eilidhmae/smaug/internal/act"
	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/command"
	"github.com/eilidhmae/smaug/internal/game"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/mudprog"
	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// ShutdownRequest is delivered on BootOpts.Shutdowns when a test-mode
// shutdown hook fires. Production ignores this channel (it exits the process
// directly).
type ShutdownRequest struct {
	Reboot bool
}

// BootOpts controls how Boot wires a couple of process-lifecycle hooks.
// Leave func fields nil to accept the per-mode default; supply your own to
// override (tests that need to inject custom behavior do this).
type BootOpts struct {
	// ShutdownFunc is invoked when an immortal issues `shutdown` or `reboot`.
	// Nil means "use the default baked into Boot for this opts flavor".
	ShutdownFunc func(reboot bool)
	// DisconnectFunc closes a descriptor's connection. Nil means use the
	// default (plain conn.Close).
	DisconnectFunc func(d *types.DescriptorData)
	// Shutdowns, when non-nil, receives a ShutdownRequest from the test-mode
	// ShutdownFunc default. Production leaves this nil (production exits
	// the process, never enqueues).
	Shutdowns chan ShutdownRequest
}

// ProductionOpts returns BootOpts that wire the real production behavior:
// shutdown saves all players and calls os.Exit; disconnect closes the conn.
func ProductionOpts() BootOpts {
	return BootOpts{}
}

// TestOpts returns BootOpts configured for testing:
//   - ShutdownFunc records to Shutdowns and cancels the loop
//   - Lowers game.BcryptCost to bcrypt.MinCost for fast login hashing
//
// Side effect: TestOpts() mutates game.BcryptCost as a process-level
// setting. Production code must call ProductionOpts() (which does not
// touch BcryptCost) so default cost remains in effect.
func TestOpts() BootOpts {
	game.BcryptCost = bcrypt.MinCost
	return BootOpts{
		Shutdowns: make(chan ShutdownRequest, 4),
	}
}

// Boot loads all game data, registers every command, and wires every
// cross-package callback. It returns the command registry and a game loop
// ready to Run. The loop is NOT started — the caller owns its lifecycle.
//
// `incoming` is the channel the network server will push new descriptors
// onto; the game loop reads from it. Callers must create the server and
// pass its Incoming channel here.
func Boot(w *world.World, dataDir string, incoming chan *types.DescriptorData, opts BootOpts) (*command.Registry, *game.GameLoop, error) {
	// WorldRef is read by act/combat/mudprog at runtime; wire early so that
	// any initialization that triggers these packages sees a non-nil world.
	act.WorldRef = w

	if err := bootDB(w, dataDir); err != nil {
		return nil, nil, err
	}

	cmdReg := registerCommands()

	loop := game.NewGameLoop(w, cmdReg, incoming)

	// Save hook used by quit and autosave paths in act.
	act.SaveFunc = func(ch *types.CharData) {
		loop.SavePlayer(ch)
	}
	act.CmdRegistry = cmdReg
	cmdReg.SocialFallback = act.CheckSocial

	// Keep act.BcryptCost in sync with game.BcryptCost so DoPassword hashes
	// at the same cost the login/nanny flow uses. TestOpts lowers the game
	// var to bcrypt.MinCost before Boot runs; this assignment propagates
	// that into act. Production leaves both at bcrypt.DefaultCost.
	act.BcryptCost = game.BcryptCost

	// mudprog needs the command registry (for CMD progs) and world ref.
	mudprog.CmdRegistry = cmdReg
	mudprog.WorldRef = w

	// Combat → mudprog hooks. combat cannot import mudprog directly
	// (mudprog already imports combat), so combat publishes func vars and
	// boot wires them here to break the cycle. DamMessage needs the skill
	// registry to dispatch skill hit/miss strings.
	combat.WorldRef = w
	combat.HitprcntHook = mudprog.TrigHitprcnt
	combat.VoidHook = mudprog.CheckVoid
	combat.ObjDamageHook = mudprog.OprogDamageTrigger
	combat.RfightHook = mudprog.RprogRfightTrigger
	combat.DeathRoomHook = mudprog.RprogDeathTrigger
	cmdReg.ObjCommandHook = mudprog.OprogCommandTrigger
	cmdReg.RoomCommandHook = mudprog.RprogCommandTrigger

	// Combat → act skill-check hooks. Same cycle-break pattern: combat
	// cannot import act, so act's canUseSkill / learnFromSuccess /
	// learnFromFailure / lookupSkillSlot are exposed through these seams.
	// Used by MultiHit's cascade (second_attack … seventh_attack, berserk,
	// dual_wield) and by the weapon-proficiency bonus in OneHit.
	combat.CanUseSkillHook = act.CanUseSkill
	combat.LearnFromSuccessHook = act.LearnFromSuccess
	combat.LearnFromFailureHook = act.LearnFromFailure
	combat.LookupSkillSlotHook = act.LookupSkillSlot
	combat.ResolveGSNs()

	// OLC editor entry point (act.DoRedit et al. call this to start editing).
	act.StartEditingFunc = game.StartEditing

	// Shutdown + disconnect hooks. If the caller supplied overrides use
	// those, otherwise fall back to the mode-appropriate default.
	if opts.ShutdownFunc != nil {
		act.ShutdownFunc = opts.ShutdownFunc
	} else {
		act.ShutdownFunc = defaultShutdownFunc(w, loop, opts)
	}
	if opts.DisconnectFunc != nil {
		act.DisconnectFunc = opts.DisconnectFunc
	} else {
		act.DisconnectFunc = defaultDisconnectFunc()
	}

	return cmdReg, loop, nil
}

// defaultShutdownFunc returns the appropriate default ShutdownFunc for the
// given opts flavor. If opts.Shutdowns is non-nil (test mode), we record
// requests and cancel the loop. Otherwise (production), we save every
// connected player and os.Exit.
func defaultShutdownFunc(w *world.World, loop *game.GameLoop, opts BootOpts) func(reboot bool) {
	if opts.Shutdowns != nil {
		// Test-mode default: record then cancel the loop.
		ch := opts.Shutdowns
		return func(reboot bool) {
			select {
			case ch <- ShutdownRequest{Reboot: reboot}:
			default:
				// Channel full — drop. Tests should size Shutdowns adequately.
			}
			loop.Cancel()
		}
	}
	// Production default: save every connected player, then exit. Exit
	// code 0 = clean shutdown, 2 = reboot (signal to supervisor to relaunch).
	// MVP: a production deployment would drain pulses first. See
	// phase5-tier4-content.md.
	return func(reboot bool) {
		for _, d := range w.Descriptors {
			if d.Character != nil && !d.Character.IsNPC() {
				loop.SavePlayer(d.Character)
			}
		}
		if reboot {
			log.Println("Reboot requested by immortal — exiting with code 2.")
			os.Exit(2)
		}
		log.Println("Shutdown requested by immortal — exiting.")
		os.Exit(0)
	}
}

// defaultDisconnectFunc returns the default DisconnectFunc. Same for both
// production and test — just close the underlying conn.
func defaultDisconnectFunc() func(d *types.DescriptorData) {
	return func(d *types.DescriptorData) {
		if d != nil && d.Conn != nil {
			_ = d.Conn.Close()
		}
	}
}

// bootDB loads areas, classes, races, skills, socials, clans, deities, and
// boards, then wires persist's skill lookup callbacks and processes area
// resets. Moved verbatim from cmd/smaug/main.go.
func bootDB(w *world.World, dataDir string) error {
	log.Println("Booting database...")

	// Load area files
	areaDir := filepath.Join(dataDir, "area")
	if err := persist.LoadAreas(w, areaDir); err != nil {
		return fmt.Errorf("loading areas: %w", err)
	}

	// Resolve exit vnums to room pointers
	w.FixExits()

	// Load class and race data. Classes, races, and skills are foundational —
	// nothing in the game works without them (character creation immediately
	// fails with no classes/races, and combat/spells need the skill table).
	// A misconfigured -data path must fail loudly at boot rather than
	// silently accepting connections and breaking at runtime.
	classDir := filepath.Join(dataDir, "classes")
	if err := persist.LoadClasses(w, classDir); err != nil {
		return fmt.Errorf("failed to load classes: %w", err)
	}
	{
		count := 0
		for _, c := range w.Classes {
			if c != nil {
				count++
			}
		}
		log.Printf("Loaded %d classes.", count)
	}

	raceDir := filepath.Join(dataDir, "races")
	if err := persist.LoadRaces(w, raceDir); err != nil {
		return fmt.Errorf("failed to load races: %w", err)
	}
	{
		count := 0
		for _, r := range w.Races {
			if r != nil {
				count++
			}
		}
		log.Printf("Loaded %d races.", count)
	}

	// Load skills data
	skillsPath := filepath.Join(dataDir, "system", "en", "skills.dat")
	if err := persist.LoadSkills(w, skillsPath); err != nil {
		return fmt.Errorf("failed to load skills: %w", err)
	}
	log.Printf("Loaded %d skills/spells.", len(w.Skills))

	// Wire skill lookups so player save/load persists learned proficiencies.
	persist.SkillNameLookup = func(name string) int {
		for i, sk := range w.Skills {
			if sk != nil && strings.EqualFold(sk.Name, name) {
				return i
			}
		}
		return -1
	}
	persist.SkillGetter = func(gsn int) *types.SkillType {
		if gsn < 0 || gsn >= len(w.Skills) {
			return nil
		}
		return w.Skills[gsn]
	}

	if len(w.Helps) > 0 {
		log.Printf("Loaded %d help entries.", len(w.Helps))
	}

	// Load subsystem data
	clanDir := filepath.Join(dataDir, "clans")
	if err := persist.LoadClansFromDir(w, clanDir); err != nil {
		log.Printf("WARNING: failed to load clans: %v", err)
	} else if len(w.Clans) > 0 {
		log.Printf("Loaded %d clans.", len(w.Clans))
	}

	deityDir := filepath.Join(dataDir, "deity")
	if err := persist.LoadDeitiesFromDir(w, deityDir); err != nil {
		log.Printf("WARNING: failed to load deities: %v", err)
	} else if len(w.Deities) > 0 {
		log.Printf("Loaded %d deities.", len(w.Deities))
	}

	socialsPath := filepath.Join(dataDir, "system", "en", "socials.dat")
	if err := persist.LoadSocials(w, socialsPath); err != nil {
		log.Printf("WARNING: failed to load socials: %v", err)
	} else if len(w.Socials) > 0 {
		log.Printf("Loaded %d socials.", len(w.Socials))
	}

	boardsPath := filepath.Join(dataDir, "boards", "boards.dat")
	if err := persist.LoadBoards(w, boardsPath); err != nil {
		log.Printf("WARNING: failed to load boards: %v", err)
	} else if len(w.Boards) > 0 {
		log.Printf("Loaded %d boards.", len(w.Boards))
	}

	log.Printf("Boot complete. %d rooms, %d mob templates, %d obj templates loaded.",
		len(w.Rooms), len(w.MobIndex), len(w.ObjIndex))

	// Process area resets to populate rooms with mobs and objects
	handler.ResetAllAreas(w)

	// Ensure we have a starting room — create a fallback if temple doesn't exist
	if w.GetRoom(types.ROOM_VNUM_TEMPLE) == nil {
		log.Printf("WARNING: Temple room (vnum %d) not found. Creating fallback.", types.ROOM_VNUM_TEMPLE)
		fallback := &types.RoomIndexData{
			Vnum:        types.ROOM_VNUM_TEMPLE,
			Name:        "The Void",
			Description: "You are floating in an empty void. The world has not been loaded.\n\r",
			SectorType:  types.SECT_INSIDE,
		}
		w.Rooms[fallback.Vnum] = fallback
	}

	return nil
}

// registerCommands builds the complete command registry. Moved verbatim
// from cmd/smaug/main.go; any command additions happen here.
func registerCommands() *command.Registry {
	reg := command.NewRegistry()

	reg.Register(&command.Command{Name: "look", DoFun: act.DoLook, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "quit", DoFun: act.DoQuit, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "say", DoFun: act.DoSay, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "score", DoFun: act.DoScore, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "who", DoFun: act.DoWho, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "commands", DoFun: act.DoCommands, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "help", DoFun: act.DoHelp, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "inventory", DoFun: act.DoInventory, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "equipment", DoFun: act.DoEquipment, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "consider", DoFun: act.DoConsider, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "examine", DoFun: act.DoExamine, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "where", DoFun: act.DoWhere, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "time", DoFun: act.DoTime, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "weather", DoFun: act.DoWeather, Position: types.POS_RESTING, Level: 0})

	// Communication commands
	reg.Register(&command.Command{Name: "tell", DoFun: act.DoTell, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "reply", DoFun: act.DoReply, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "yell", DoFun: act.DoYell, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "shout", DoFun: act.DoShout, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "gossip", DoFun: act.DoGossip, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "emote", DoFun: act.DoEmote, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "pmote", DoFun: act.DoPmote, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "speak", DoFun: act.DoSpeak, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "learn", DoFun: act.DoLearn, Position: types.POS_RESTING, Level: 0})

	// Phase 5 Tier 9 channels (plan-channels.md): immtalk (alias `:`),
	// gtell (alias `;`), auction (broadcast stub). Aliases share the
	// underlying handler per the same-name-second-entry pattern used for
	// `pager`/`pagelen`. Command parsing requires a space after the alias
	// (`: hi`) — prefix matching within the token is not supported; the
	// plan notes this as a deliberate limitation.
	reg.Register(&command.Command{Name: "immtalk", DoFun: act.DoImmtalk, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: ":", DoFun: act.DoImmtalk, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "gtell", DoFun: act.DoGtell, Position: types.POS_SLEEPING, Level: 0})
	reg.Register(&command.Command{Name: ";", DoFun: act.DoGtell, Position: types.POS_SLEEPING, Level: 0})
	reg.Register(&command.Command{Name: "auction", DoFun: act.DoAuction, Position: types.POS_SLEEPING, Level: 0})

	// Object commands
	reg.Register(&command.Command{Name: "get", DoFun: act.DoGet, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "drop", DoFun: act.DoDrop, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "put", DoFun: act.DoPut, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "give", DoFun: act.DoGive, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "wear", DoFun: act.DoWear, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "remove", DoFun: act.DoRemove, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "sacrifice", DoFun: act.DoSacrifice, Position: types.POS_RESTING, Level: 0})

	// Config commands
	reg.Register(&command.Command{Name: "pager", DoFun: act.DoPager, Position: types.POS_DEAD, Level: 0})
	// `pagelen` is a second name for `DoPager`. Relies on registry map-keyed
	// storage (internal/command/interpret.go Register), so both entries
	// resolve to the same handler. Tests in act/playercfg_test.go assert the
	// alias round-trips through Interpret.
	reg.Register(&command.Command{Name: "pagelen", DoFun: act.DoPager, Position: types.POS_DEAD, Level: 0})

	// Player-config commands — Phase 5 Tier 7 (plan-player-config.md).
	// Trust=0, mortals-only logic inside each handler.
	reg.Register(&command.Command{Name: "save", DoFun: act.DoSave, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "afk", DoFun: act.DoAfk, Position: types.POS_SLEEPING, Level: 0})
	reg.Register(&command.Command{Name: "title", DoFun: act.DoTitle, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "password", DoFun: act.DoPassword, Position: types.POS_DEAD, Level: 0})

	// Clan/deity/board commands
	reg.Register(&command.Command{Name: "clans", DoFun: act.DoClans, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "claninfo", DoFun: act.DoClanInfo, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "clantalk", DoFun: act.DoClantalk, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "join", DoFun: act.DoClanJoin, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "leave", DoFun: act.DoClanLeave, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "clandeposit", DoFun: act.DoClanDeposit, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "clanwithdraw", DoFun: act.DoClanWithdraw, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "deities", DoFun: act.DoDeities, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "devote", DoFun: act.DoDevote, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "note", DoFun: act.DoNote, Position: types.POS_RESTING, Level: 0})

	// Immortal control (additional)
	reg.Register(&command.Command{Name: "snoop", DoFun: act.DoSnoop, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "ban", DoFun: act.DoBan, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})

	// Shop commands
	reg.Register(&command.Command{Name: "buy", DoFun: act.DoBuy, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "sell", DoFun: act.DoSell, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "list", DoFun: act.DoList, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "value", DoFun: act.DoValue, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "repair", DoFun: act.DoRepair, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "appraise", DoFun: act.DoAppraise, Position: types.POS_STANDING, Level: 0})

	// Consume commands
	reg.Register(&command.Command{Name: "eat", DoFun: act.DoEat, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "drink", DoFun: act.DoDrink, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "fill", DoFun: act.DoFill, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "empty", DoFun: act.DoEmpty, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "quaff", DoFun: act.DoQuaff, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "recite", DoFun: act.DoRecite, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "brandish", DoFun: act.DoBrandish, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "zap", DoFun: act.DoZap, Position: types.POS_RESTING, Level: 0})

	// Magic commands
	reg.Register(&command.Command{Name: "cast", DoFun: act.DoCast, Position: types.POS_FIGHTING, Level: 0})

	// Combat commands
	reg.Register(&command.Command{Name: "kill", DoFun: act.DoKill, Position: types.POS_FIGHTING, Level: 0})
	reg.Register(&command.Command{Name: "flee", DoFun: act.DoFlee, Position: types.POS_FIGHTING, Level: 0})
	reg.Register(&command.Command{Name: "murder", DoFun: act.DoMurder, Position: types.POS_FIGHTING, Level: 0})
	reg.Register(&command.Command{Name: "wimpy", DoFun: act.DoWimpy, Position: types.POS_DEAD, Level: 0})

	// Skill commands
	reg.Register(&command.Command{Name: "backstab", DoFun: act.DoBackstab, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "bash", DoFun: act.DoBash, Position: types.POS_FIGHTING, Level: 0})
	reg.Register(&command.Command{Name: "kick", DoFun: act.DoKickSkill, Position: types.POS_FIGHTING, Level: 0})
	reg.Register(&command.Command{Name: "disarm", DoFun: act.DoDisarm, Position: types.POS_FIGHTING, Level: 0})
	reg.Register(&command.Command{Name: "rescue", DoFun: act.DoRescue, Position: types.POS_FIGHTING, Level: 0})
	reg.Register(&command.Command{Name: "sneak", DoFun: act.DoSneak, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "hide", DoFun: act.DoHide, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "steal", DoFun: act.DoSteal, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "pick", DoFun: act.DoPick, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "scan", DoFun: act.DoScan, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "aid", DoFun: act.DoAid, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "recall", DoFun: act.DoRecall, Position: types.POS_STANDING, Level: 0})

	// Combat/utility skills — Phase 5 Tier 4 G4
	reg.Register(&command.Command{Name: "bite", DoFun: act.DoBite, Position: types.POS_FIGHTING, Level: 0})
	reg.Register(&command.Command{Name: "claw", DoFun: act.DoClaw, Position: types.POS_FIGHTING, Level: 0})
	reg.Register(&command.Command{Name: "punch", DoFun: act.DoPunch, Position: types.POS_FIGHTING, Level: 0})
	reg.Register(&command.Command{Name: "sting", DoFun: act.DoSting, Position: types.POS_FIGHTING, Level: 0})
	reg.Register(&command.Command{Name: "tail", DoFun: act.DoTail, Position: types.POS_FIGHTING, Level: 0})
	reg.Register(&command.Command{Name: "circle", DoFun: act.DoCircle, Position: types.POS_FIGHTING, Level: 0})
	reg.Register(&command.Command{Name: "gouge", DoFun: act.DoGouge, Position: types.POS_FIGHTING, Level: 0})
	reg.Register(&command.Command{Name: "stun", DoFun: act.DoStun, Position: types.POS_FIGHTING, Level: 0})
	reg.Register(&command.Command{Name: "grapple", DoFun: act.DoGrapple, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "cleave", DoFun: act.DoCleave, Position: types.POS_FIGHTING, Level: 0})
	reg.Register(&command.Command{Name: "hitall", DoFun: act.DoHitall, Position: types.POS_FIGHTING, Level: 0})
	reg.Register(&command.Command{Name: "berserk", DoFun: act.DoBerserk, Position: types.POS_FIGHTING, Level: 0})
	reg.Register(&command.Command{Name: "meditate", DoFun: act.DoMeditate, Position: types.POS_SLEEPING, Level: 0})
	reg.Register(&command.Command{Name: "trance", DoFun: act.DoTrance, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "search", DoFun: act.DoSearch, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "detrap", DoFun: act.DoDetrap, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "dig", DoFun: act.DoDig, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "visible", DoFun: act.DoVisible, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "style", DoFun: act.DoStyle, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "stance", DoFun: act.DoStance, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "mistwalk", DoFun: act.DoMistwalk, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "feed", DoFun: act.DoFeed, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "skin", DoFun: act.DoSkin, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "poison", DoFun: act.DoPoisonWeapon, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "fire", DoFun: act.DoFire, Position: types.POS_FIGHTING, Level: 0})
	reg.Register(&command.Command{Name: "scribe", DoFun: act.DoScribe, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "cook", DoFun: act.DoCook, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "slookup", DoFun: act.DoSlookup, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "sset", DoFun: act.DoSset, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})

	// Quest
	reg.Register(&command.Command{Name: "quest", DoFun: act.DoQuest, Position: types.POS_RESTING, Level: 0})

	// Skill/spell info
	reg.Register(&command.Command{Name: "practice", DoFun: act.DoPractice, Position: types.POS_SLEEPING, Level: 0})
	reg.Register(&command.Command{Name: "skills", DoFun: act.DoSkills, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "spells", DoFun: act.DoSpells, Position: types.POS_DEAD, Level: 0})

	// Door commands
	reg.Register(&command.Command{Name: "open", DoFun: act.DoOpen, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "close", DoFun: act.DoClose, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "unlock", DoFun: act.DoUnlock, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "lock", DoFun: act.DoLock, Position: types.POS_STANDING, Level: 0})

	// Immortal stat commands
	reg.Register(&command.Command{Name: "mstat", DoFun: act.DoMstat, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "ostat", DoFun: act.DoOstat, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "rstat", DoFun: act.DoRstat, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})

	// Immortal movement
	reg.Register(&command.Command{Name: "goto", DoFun: act.DoGoto, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "transfer", DoFun: act.DoTransfer, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "at", DoFun: act.DoAt, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "bamfin", DoFun: act.DoBamfin, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "bamfout", DoFun: act.DoBamfout, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})

	// Immortal action
	reg.Register(&command.Command{Name: "force", DoFun: act.DoForce, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "peace", DoFun: act.DoPeace, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "purge", DoFun: act.DoPurge, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "restore", DoFun: act.DoRestore, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "advance", DoFun: act.DoAdvance, Position: types.POS_DEAD, Level: types.LEVEL_SUPREME})
	reg.Register(&command.Command{Name: "slay", DoFun: act.DoSlay, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})

	// Immortal info
	reg.Register(&command.Command{Name: "mfind", DoFun: act.DoMfind, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "ofind", DoFun: act.DoOfind, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "mwhere", DoFun: act.DoMwhere, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "owhere", DoFun: act.DoOwhere, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "users", DoFun: act.DoUsers, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})

	// Immortal control
	reg.Register(&command.Command{Name: "invis", DoFun: act.DoInvis, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "holylight", DoFun: act.DoHolylight, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "freeze", DoFun: act.DoFreeze, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "silence", DoFun: act.DoSilence, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})

	// Immortal system
	reg.Register(&command.Command{Name: "echo", DoFun: act.DoEcho, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "recho", DoFun: act.DoRecho, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "aecho", DoFun: act.DoAecho, Position: types.POS_DEAD, Level: types.LEVEL_TRUEIMM})

	// Possession / return — switch into an NPC body (act_wiz.c:3760).
	reg.Register(&command.Command{Name: "switch", DoFun: act.DoSwitch, Position: types.POS_DEAD, Level: types.LEVEL_CREATOR})
	reg.Register(&command.Command{Name: "return", DoFun: act.DoReturn, Position: types.POS_DEAD, Level: 0})

	// Lockout / lifecycle — wizlock/shutdown/reboot are reserved to the
	// very top tier of imms. See act_wiz.c:6075/3606/3554.
	reg.Register(&command.Command{Name: "wizlock", DoFun: act.DoWizlock, Position: types.POS_DEAD, Level: types.LEVEL_ASCENDANT})
	reg.Register(&command.Command{Name: "shutdown", DoFun: act.DoShutdown, Position: types.POS_DEAD, Level: types.LEVEL_ASCENDANT})
	reg.Register(&command.Command{Name: "reboot", DoFun: act.DoReboot, Position: types.POS_DEAD, Level: types.LEVEL_GREATER})

	// Help / discipline / mudprog inspection
	reg.Register(&command.Command{Name: "wizhelp", DoFun: act.DoWizhelp, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "hell", DoFun: act.DoHell, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "log", DoFun: act.DoLog, Position: types.POS_DEAD, Level: types.LEVEL_DEMI})
	reg.Register(&command.Command{Name: "deny", DoFun: act.DoDeny, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "pardon", DoFun: act.DoPardon, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "disconnect", DoFun: act.DoDisconnect, Position: types.POS_DEAD, Level: types.LEVEL_LESSER})
	reg.Register(&command.Command{Name: "mortalize", DoFun: act.DoMortalize, Position: types.POS_DEAD, Level: types.LEVEL_LESSER})
	reg.Register(&command.Command{Name: "mpstat", DoFun: act.DoMpstat, Position: types.POS_DEAD, Level: types.LEVEL_TRUEIMM})
	reg.Register(&command.Command{Name: "opstat", DoFun: act.DoOpstat, Position: types.POS_DEAD, Level: types.LEVEL_TRUEIMM})
	reg.Register(&command.Command{Name: "rpstat", DoFun: act.DoRpstat, Position: types.POS_DEAD, Level: types.LEVEL_TRUEIMM})

	// OLC commands
	reg.Register(&command.Command{Name: "redit", DoFun: act.DoRedit, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "ocreate", DoFun: act.DoOcreate, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "mcreate", DoFun: act.DoMcreate, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "rdig", DoFun: act.DoRdig, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "rlist", DoFun: act.DoRlist, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "olist", DoFun: act.DoOlist, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "mlist", DoFun: act.DoMlist, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "savearea", DoFun: act.DoSaveArea, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "mset", DoFun: act.DoMset, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "oset", DoFun: act.DoOset, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "rset", DoFun: act.DoRset, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "aset", DoFun: act.DoAset, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "astat", DoFun: act.DoAstat, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "oedit", DoFun: act.DoOedit, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "medit", DoFun: act.DoMedit, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "rdelete", DoFun: act.DoRdelete, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "odelete", DoFun: act.DoOdelete, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "mdelete", DoFun: act.DoMdelete, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "mpedit", DoFun: act.DoMpedit, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "opedit", DoFun: act.DoOpedit, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})
	reg.Register(&command.Command{Name: "rpedit", DoFun: act.DoRpedit, Position: types.POS_DEAD, Level: types.LEVEL_IMMORTAL})

	// Banking
	reg.Register(&command.Command{Name: "bank", DoFun: act.DoBank, Position: types.POS_STANDING, Level: 0})

	// Tracking
	reg.Register(&command.Command{Name: "track", DoFun: act.DoTrack, Position: types.POS_STANDING, Level: 0})

	// Group commands
	reg.Register(&command.Command{Name: "follow", DoFun: act.DoFollow, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "group", DoFun: act.DoGroup, Position: types.POS_SLEEPING, Level: 0})
	reg.Register(&command.Command{Name: "order", DoFun: act.DoOrder, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "assist", DoFun: act.DoAssist, Position: types.POS_STANDING, Level: 0})

	// Mount commands
	reg.Register(&command.Command{Name: "mount", DoFun: act.DoMount, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "dismount", DoFun: act.DoDismount, Position: types.POS_MOUNTED, Level: 0})

	// Position commands
	reg.Register(&command.Command{Name: "rest", DoFun: act.DoRest, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "sit", DoFun: act.DoSit, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "stand", DoFun: act.DoStand, Position: types.POS_SLEEPING, Level: 0})
	reg.Register(&command.Command{Name: "sleep", DoFun: act.DoSleep, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "wake", DoFun: act.DoWake, Position: types.POS_SLEEPING, Level: 0})

	// Movement commands
	reg.Register(&command.Command{Name: "north", DoFun: act.DoNorth, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "east", DoFun: act.DoEast, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "south", DoFun: act.DoSouth, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "west", DoFun: act.DoWest, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "up", DoFun: act.DoUp, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "down", DoFun: act.DoDown, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "northeast", DoFun: act.DoNortheast, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "northwest", DoFun: act.DoNorthwest, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "southeast", DoFun: act.DoSoutheast, Position: types.POS_STANDING, Level: 0})
	reg.Register(&command.Command{Name: "southwest", DoFun: act.DoSouthwest, Position: types.POS_STANDING, Level: 0})

	// Tier-4 G6 mortal commands
	reg.Register(&command.Command{Name: "split", DoFun: act.DoSplit, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "light", DoFun: act.DoLight, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "throw", DoFun: act.DoThrow, Position: types.POS_FIGHTING, Level: 0})
	reg.Register(&command.Command{Name: "alias", DoFun: act.DoAlias, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "unalias", DoFun: act.DoUnalias, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "areas", DoFun: act.DoAreas, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "altscore", DoFun: act.DoAltscore, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "color", DoFun: act.DoColor, Position: types.POS_DEAD, Level: 0})
	reg.Register(&command.Command{Name: "compress", DoFun: act.DoCompress, Position: types.POS_DEAD, Level: 0})

	return reg
}
