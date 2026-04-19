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

	"github.com/eilidhmae/smaug/internal/act"
	"github.com/eilidhmae/smaug/internal/boot"
	"github.com/eilidhmae/smaug/internal/game"
	smaugnet "github.com/eilidhmae/smaug/internal/net"
	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func main() {
	// Parse --hotboot-recover argv BEFORE flag.Parse so its payload
	// tokens (ln_fd + fd:idx pairs) don't confuse the flag package.
	// See plan-phase6-hotboot.md §G5.
	hotbootArgs, err := ParseHotbootArgv(os.Args[1:])
	if err != nil {
		log.Fatalf("Bad hotboot argv: %v", err)
	}
	// Rewrite os.Args so flag.Parse sees a clean arg list.
	os.Args = append([]string{os.Args[0]}, StripHotbootArgv(os.Args[1:])...)

	port := flag.Int("port", 4000, "Port to listen on")
	dataDir := flag.String("data", "../db", "Path to data directory")
	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Printf("SMAUG MUD starting on port %d...", *port)

	// Create the world.
	w := world.New(*dataDir)

	var (
		server    *smaugnet.Server
		gameLoop  *game.GameLoop
		boundPort int
	)

	if hotbootArgs != nil {
		// --- Hotboot recovery path. BootRecover inherits the listener
		// FD passed via argv and re-wraps each saved descriptor FD. No
		// fresh Listen — the port is whatever the parent bound. The
		// returned *smaugnet.Server is already running against the
		// inherited listener with accept loops spawned. ---
		log.Printf("[hotboot] Recovering with ln_fd=%d and %d session(s)",
			hotbootArgs.LnFD, len(hotbootArgs.Sessions))
		var bootErr error
		_, gameLoop, server, bootErr = boot.BootRecover(
			w, *dataDir, boot.ProductionOpts(),
			hotbootArgs.LnFD, hotbootArgs.Sessions,
		)
		if bootErr != nil {
			log.Fatalf("Failed to hotboot-recover: %v", bootErr)
		}
	} else {
		// --- Normal cold boot path. ---
		// Create the network server first so we can pass its Incoming
		// channel to Boot, which wires the game loop to read from it.
		server = smaugnet.NewServer()

		_, gameLoop, err = boot.Boot(w, *dataDir, server.Incoming, boot.ProductionOpts())
		if err != nil {
			log.Fatalf("Failed to boot: %v", err)
		}

		// Bind the listener ourselves so the port is reserved before we start
		// accepting — avoids the bind race if a test harness recycles the port.
		ln, listenErr := gonet.Listen("tcp", fmt.Sprintf(":%d", *port))
		if listenErr != nil {
			log.Fatalf("Failed to listen on port %d: %v", *port, listenErr)
		}
		if tcp, ok := ln.Addr().(*gonet.TCPAddr); ok {
			boundPort = tcp.Port
		} else {
			boundPort = *port
		}
		log.Printf("[net] Listening on port %d", boundPort)
		if err := server.StartOnListener(ln); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}
	defer server.Stop()

	// Derive the actually-bound port for the subsequent log line. In the
	// recover path this comes from the inherited listener via Server.Addr();
	// in the normal path it was captured above. Required exact prefix
	// "SMAUG MUD listening on :<port>" for the G6 integration test to
	// parse from stderr.
	if hotbootArgs != nil {
		if addr, ok := server.Addr().(*gonet.TCPAddr); ok && addr != nil {
			boundPort = addr.Port
		} else {
			boundPort = *port
		}
	}
	log.Printf("SMAUG MUD listening on :%d", boundPort)

	// Hotboot seams (plan-phase6-hotboot.md §G3). Wired here because
	// internal/boot cannot import internal/net (would pull a cycle).
	act.HotbootPort = boundPort
	act.HotbootPauseFunc = func(descriptors []*types.DescriptorData, p int) (*os.File, []*os.File, []persist.HotbootSession, error) {
		return server.PauseForHotboot(descriptors, p)
	}
	act.HotbootResumeFunc = func(lnFile *os.File) error {
		return server.ResumeFromPause(lnFile)
	}

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

	log.Printf("SMAUG MUD is ready. Listening on port %d.", boundPort)

	// Run the game loop (blocks until context is cancelled).
	gameLoop.Run(ctx)

	log.Println("SMAUG MUD shut down.")
}
