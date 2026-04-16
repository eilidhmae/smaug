package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	gonet "net"
	"os"
	"os/signal"
	"syscall"

	"github.com/eilidhmae/smaug/internal/boot"
	smaugnet "github.com/eilidhmae/smaug/internal/net"
	"github.com/eilidhmae/smaug/internal/world"
)

func main() {
	port := flag.Int("port", 4000, "Port to listen on")
	dataDir := flag.String("data", "../db", "Path to data directory")
	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Printf("SMAUG MUD starting on port %d...", *port)

	// Create the world.
	w := world.New(*dataDir)

	// Create the network server first so we can pass its Incoming channel to
	// boot, which wires the game loop to read from it.
	server := smaugnet.NewServer()

	// Load data, register commands, wire all cross-package callbacks.
	_, gameLoop, err := boot.Boot(w, *dataDir, server.Incoming, boot.ProductionOpts())
	if err != nil {
		log.Fatalf("Failed to boot: %v", err)
	}

	// Bind the listener ourselves so the port is reserved before we start
	// accepting — avoids the bind race if a test harness recycles the port.
	ln, err := gonet.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", *port, err)
	}
	log.Printf("[net] Listening on port %d", *port)
	if err := server.StartOnListener(ln); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Set up signal handling for graceful shutdown.
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

	// Run the game loop (blocks until context is cancelled).
	gameLoop.Run(ctx)

	log.Println("SMAUG MUD shut down.")
}
