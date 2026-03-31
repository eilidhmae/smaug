package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/eilidhmae/smaug/internal/act"
	"github.com/eilidhmae/smaug/internal/command"
	"github.com/eilidhmae/smaug/internal/game"
	smaugnet "github.com/eilidhmae/smaug/internal/net"
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

	// For now, create a minimal world with a starting room
	startRoom := &types.RoomIndexData{
		Vnum:        types.ROOM_VNUM_TEMPLE,
		Name:        "The Temple of Midgaard",
		Description: "You are standing in the Temple of Midgaard. Marble pillars rise\n\rto the ceiling high above, and the air smells of ancient incense.\n\rA soft light filters through stained glass windows.\n\r",
		SectorType:  types.SECT_INSIDE,
	}

	// Add a second room and connect them
	squareRoom := &types.RoomIndexData{
		Vnum:        types.ROOM_VNUM_CHAT,
		Name:        "The Town Square",
		Description: "You are standing in the town square of Midgaard. The square bustles\n\rwith activity as merchants hawk their wares and travellers rest.\n\rA large fountain stands in the center.\n\r",
		SectorType:  types.SECT_CITY,
	}

	// Connect rooms: temple is south of square
	startRoom.Exits = append(startRoom.Exits, &types.ExitData{
		Direction: types.DIR_NORTH,
		ToRoom:    squareRoom,
		Vnum:      squareRoom.Vnum,
	})
	squareRoom.Exits = append(squareRoom.Exits, &types.ExitData{
		Direction: types.DIR_SOUTH,
		ToRoom:    startRoom,
		Vnum:      startRoom.Vnum,
	})

	w.Rooms[startRoom.Vnum] = startRoom
	w.Rooms[squareRoom.Vnum] = squareRoom

	// TODO: Load from area files with persist.LoadAreas()

	log.Printf("Boot complete. %d rooms loaded.", len(w.Rooms))
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
