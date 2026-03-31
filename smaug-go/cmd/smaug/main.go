package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"path/filepath"

	"github.com/eilidhmae/smaug/internal/act"
	"github.com/eilidhmae/smaug/internal/command"
	"github.com/eilidhmae/smaug/internal/game"
	"github.com/eilidhmae/smaug/internal/handler"
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

	if len(w.Helps) > 0 {
		log.Printf("Loaded %d help entries.", len(w.Helps))
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

	// Object commands
	reg.Register(&command.Command{Name: "get", DoFun: act.DoGet, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "drop", DoFun: act.DoDrop, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "put", DoFun: act.DoPut, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "give", DoFun: act.DoGive, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "wear", DoFun: act.DoWear, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "remove", DoFun: act.DoRemove, Position: types.POS_RESTING, Level: 0})
	reg.Register(&command.Command{Name: "sacrifice", DoFun: act.DoSacrifice, Position: types.POS_RESTING, Level: 0})

	// Combat commands
	reg.Register(&command.Command{Name: "kill", DoFun: act.DoKill, Position: types.POS_FIGHTING, Level: 0})
	reg.Register(&command.Command{Name: "flee", DoFun: act.DoFlee, Position: types.POS_FIGHTING, Level: 0})

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
