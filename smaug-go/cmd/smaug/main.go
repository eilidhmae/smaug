package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"path/filepath"

	"github.com/eilidhmae/smaug/internal/act"
	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/command"
	"github.com/eilidhmae/smaug/internal/game"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/mudprog"
	smaugnet "github.com/eilidhmae/smaug/internal/net"
	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func main() {
	port := flag.Int("port", 4000, "Port to listen on")
	dataDir := flag.String("data", "../db", "Path to data directory")
	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Printf("SMAUG MUD starting on port %d...", *port)

	// Create the world
	w := world.New(*dataDir)

	// Set world reference for commands
	act.WorldRef = w

	// Boot the game world
	if err := bootDB(w); err != nil {
		log.Fatalf("Failed to boot database: %v", err)
	}

	// Create command registry
	cmdReg := registerCommands()

	// Create network server
	server := smaugnet.NewServer()

	// Create game loop
	gameLoop := game.NewGameLoop(w, cmdReg, server.Incoming)

	// Wire save function for quit command
	act.SaveFunc = func(ch *types.CharData) {
		gameLoop.SavePlayer(ch)
	}

	// Wire command registry for force/at commands
	act.CmdRegistry = cmdReg

	// Wire social fallback for command interpreter
	cmdReg.SocialFallback = act.CheckSocial

	// Wire mudprog system
	mudprog.CmdRegistry = cmdReg
	mudprog.WorldRef = w

	// Wire combat → mudprog hooks (combat cannot import mudprog directly:
	// mudprog already imports combat, so combat publishes func vars and main
	// wires them to break the cycle).
	combat.HitprcntHook = mudprog.TrigHitprcnt
	combat.VoidHook = mudprog.CheckVoid
	combat.ObjDamageHook = mudprog.OprogDamageTrigger
	combat.RfightHook = mudprog.RprogRfightTrigger
	combat.DeathRoomHook = mudprog.RprogDeathTrigger
	cmdReg.ObjCommandHook = mudprog.OprogCommandTrigger
	cmdReg.RoomCommandHook = mudprog.RprogCommandTrigger

	// Wire OLC editor function
	act.StartEditingFunc = game.StartEditing

	// Start network server
	if err := server.Start(*port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	log.Printf("SMAUG MUD is ready. Listening on port %d.", *port)

	// Run the game loop (blocks until context is cancelled)
	gameLoop.Run(ctx)

	log.Println("SMAUG MUD shut down.")
}

func bootDB(w *world.World) error {
	log.Println("Booting database...")

	// Load area files
	areaDir := filepath.Join(w.DataDir, "area")
	if err := persist.LoadAreas(w, areaDir); err != nil {
		return fmt.Errorf("loading areas: %w", err)
	}

	// Resolve exit vnums to room pointers
	w.FixExits()

	// Load class and race data
	classDir := filepath.Join(w.DataDir, "classes")
	if err := persist.LoadClasses(w, classDir); err != nil {
		log.Printf("WARNING: failed to load classes: %v", err)
	} else {
		count := 0
		for _, c := range w.Classes {
			if c != nil {
				count++
			}
		}
		log.Printf("Loaded %d classes.", count)
	}

	raceDir := filepath.Join(w.DataDir, "races")
	if err := persist.LoadRaces(w, raceDir); err != nil {
		log.Printf("WARNING: failed to load races: %v", err)
	} else {
		count := 0
		for _, r := range w.Races {
			if r != nil {
				count++
			}
		}
		log.Printf("Loaded %d races.", count)
	}

	// Load skills data
	skillsPath := filepath.Join(w.DataDir, "system", "en", "skills.dat")
	if err := persist.LoadSkills(w, skillsPath); err != nil {
		log.Printf("WARNING: failed to load skills: %v", err)
	} else {
		log.Printf("Loaded %d skills/spells.", len(w.Skills))
	}

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
	clanDir := filepath.Join(w.DataDir, "clans")
	if err := persist.LoadClansFromDir(w, clanDir); err != nil {
		log.Printf("WARNING: failed to load clans: %v", err)
	} else if len(w.Clans) > 0 {
		log.Printf("Loaded %d clans.", len(w.Clans))
	}

	deityDir := filepath.Join(w.DataDir, "deity")
	if err := persist.LoadDeitiesFromDir(w, deityDir); err != nil {
		log.Printf("WARNING: failed to load deities: %v", err)
	} else if len(w.Deities) > 0 {
		log.Printf("Loaded %d deities.", len(w.Deities))
	}

	socialsPath := filepath.Join(w.DataDir, "system", "en", "socials.dat")
	if err := persist.LoadSocials(w, socialsPath); err != nil {
		log.Printf("WARNING: failed to load socials: %v", err)
	} else if len(w.Socials) > 0 {
		log.Printf("Loaded %d socials.", len(w.Socials))
	}

	boardsPath := filepath.Join(w.DataDir, "boards", "boards.dat")
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

	return reg
}
