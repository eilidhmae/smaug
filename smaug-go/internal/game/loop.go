package game

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/eilidhmae/smaug/internal/command"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// greeting is the SMAUG banner sent to new connections.
const greeting = "\n\r" +
	"  ________                                   \n\r" +
	" /   ___  \\   M   U   D                      \n\r" +
	" \\  (__/  /              ___   __  __   ____  \n\r" +
	"  \\_____  \\   /\\/\\     /   /  / / / /  / __ \\ \n\r" +
	"  /  \\__/  | / /\\ \\  / /_/  / /_/ /  / /_/ / \n\r" +
	" /________/|/_/  \\_\\/_/ /_ /____,_| /___  /  \n\r" +
	"                                   /______/   \n\r" +
	"\n\r" +
	"               SMAUG 1.8 (Go)\n\r" +
	"     Based on SMAUG by Thoric et al.\n\r" +
	"\n\r" +
	"By what name do you wish to be known? "

// Telnet echo-off sequence: IAC WILL ECHO
var telnetEchoOff = []byte{255, 251, 1}

// Telnet echo-on sequence: IAC WONT ECHO
var telnetEchoOn = []byte{255, 252, 1}

const motd = "\n\r" +
	"Welcome to SMAUG!\n\r" +
	"This is a development version of the Go port.\n\r" +
	"\n\r" +
	"[Press Enter to continue]\n\r"

// GameLoop is the main game loop, running at 4 pulses per second.
type GameLoop struct {
	world    *world.World
	cmdReg   *command.Registry
	incoming chan *types.DescriptorData // new connections from server

	// Pulse counters
	pulseArea     int
	pulseViolence int
	pulseMobile   int
	pulseTick     int
}

// NewGameLoop creates a new game loop.
func NewGameLoop(w *world.World, cmdReg *command.Registry, incoming chan *types.DescriptorData) *GameLoop {
	return &GameLoop{
		world:         w,
		cmdReg:        cmdReg,
		incoming:      incoming,
		pulseArea:     types.PULSE_AREA,
		pulseViolence: types.PULSE_VIOLENCE,
		pulseMobile:   types.PULSE_MOBILE,
		pulseTick:     types.PULSE_TICK,
	}
}

// Run starts the game loop, blocking until the context is cancelled.
func (g *GameLoop) Run(ctx context.Context) {
	ticker := time.NewTicker(250 * time.Millisecond) // PULSE_PER_SECOND = 4
	defer ticker.Stop()

	log.Println("Game loop started.")

	for {
		select {
		case <-ctx.Done():
			log.Println("Game loop stopping.")
			return
		case <-ticker.C:
			g.pulse()
		}
	}
}

// pulse handles one tick of the game loop.
func (g *GameLoop) pulse() {
	// 1. Accept new connections
	g.acceptNewConnections()

	// 2. Process input from all descriptors
	g.processInput()

	// 3. Decrement pulse counters and fire updates
	g.pulseArea--
	if g.pulseArea <= 0 {
		g.pulseArea = types.PULSE_AREA
		g.areaUpdate()
	}

	g.pulseViolence--
	if g.pulseViolence <= 0 {
		g.pulseViolence = types.PULSE_VIOLENCE
		g.violenceUpdate()
	}

	g.pulseMobile--
	if g.pulseMobile <= 0 {
		g.pulseMobile = types.PULSE_MOBILE
		g.mobileUpdate()
	}

	g.pulseTick--
	if g.pulseTick <= 0 {
		g.pulseTick = types.PULSE_TICK
		g.charUpdate()
		g.objUpdate()
	}

	// 4. Flush output for all descriptors
	g.flushOutput()

	// 5. Clean up dead descriptors
	g.cleanupDescriptors()
}

// acceptNewConnections drains the incoming channel and adds new descriptors.
func (g *GameLoop) acceptNewConnections() {
	for {
		select {
		case d := <-g.incoming:
			g.world.Descriptors = append(g.world.Descriptors, d)
			d.WriteToBuffer(greeting)
			log.Printf("New connection from %s", d.Host)
		default:
			return
		}
	}
}

// processInput reads one command from each descriptor and dispatches it.
func (g *GameLoop) processInput() {
	for _, d := range g.world.Descriptors {
		select {
		case line := <-d.InputQueue:
			line = strings.TrimRight(line, "\r\n")
			if d.Connected == int(types.CON_PLAYING) {
				if d.Character != nil {
					g.cmdReg.Interpret(d.Character, line)
				}
			} else {
				g.nanny(d, line)
			}
		default:
			// No input waiting
		}
	}
}

// nanny handles the login state machine for descriptors not yet playing.
func (g *GameLoop) nanny(d *types.DescriptorData, line string) {
	switch d.Connected {
	case types.CON_GET_NAME:
		name := strings.TrimSpace(line)
		if name == "" {
			d.WriteToBuffer("By what name do you wish to be known? ")
			return
		}
		// Capitalize the name
		if len(name) > 0 {
			name = strings.ToUpper(name[:1]) + strings.ToLower(name[1:])
		}
		d.User = name

		// TODO: check for existing player file, load it
		// For now, skip password and go straight to MOTD
		_, _ = d.Conn.Write(telnetEchoOff)
		d.WriteToBuffer("Password: ")
		d.Connected = int(types.CON_GET_OLD_PASSWORD)

	case types.CON_GET_OLD_PASSWORD:
		// Turn echo back on
		_, _ = d.Conn.Write(telnetEchoOn)
		d.WriteToBuffer("\n\r")

		// TODO: validate password against saved player file
		// For now, accept any password and show MOTD
		d.WriteToBuffer(motd)
		d.Connected = int(types.CON_READ_MOTD)

	case types.CON_READ_MOTD:
		// Player pressed enter after MOTD — enter the game
		g.enterGame(d)

	default:
		d.WriteToBuffer("Unexpected state. Disconnecting.\n\r")
		d.Connected = -1
	}
}

// enterGame creates a character and places them in the world.
func (g *GameLoop) enterGame(d *types.DescriptorData) {
	// Create a new character for this connection
	ch := &types.CharData{
		Name:       d.User,
		ShortDescr: d.User,
		LongDescr:  d.User + " is here.\n\r",
		Level:      1,
		Sex:        types.SEX_NEUTRAL,
		Class:      types.CLASS_WARRIOR,
		Race:       types.RACE_HUMAN,
		Position:   types.POS_STANDING,
		Hit:        100,
		MaxHit:     100,
		Mana:       100,
		MaxMana:    100,
		Move:       100,
		MaxMove:    100,
		PermStr:    13,
		PermInt:    13,
		PermWis:    13,
		PermDex:    13,
		PermCon:    13,
		PermCha:    13,
		PermLck:    13,
		Gold:       500,
		Desc:       d,
		PCData: &types.PCData{
			Title:    "the newbie",
			Prompt:   "<%hhp %mm %vmv> ",
			Filename: strings.ToLower(d.User),
		},
	}

	// Set ACT flags for player (not NPC)
	// PLR_IS_NPC is bit 0 — do NOT set it for players
	ch.Act.Set(types.PLR_AUTOEXIT)
	ch.Act.Set(types.PLR_ANSI)

	d.Character = ch
	ch.Desc = d
	d.Connected = int(types.CON_PLAYING)

	// Add to world character list
	g.world.AddChar(ch)

	// Place in starting room
	startRoom := g.world.GetRoom(types.ROOM_VNUM_TEMPLE)
	if startRoom == nil {
		// Fallback: find any room
		for _, r := range g.world.Rooms {
			startRoom = r
			break
		}
	}
	if startRoom != nil {
		ch.InRoom = startRoom
		startRoom.People = append(startRoom.People, ch)
	}

	log.Printf("%s has entered the game from %s", ch.Name, d.Host)

	d.WriteToBuffer("\n\rWelcome to SMAUG!\n\r\n\r")

	// Auto-look
	if ch.InRoom != nil {
		// Inline a basic look since we can't easily call act.DoLook from here
		d.WriteToBufferf("%s\n\r", ch.InRoom.Name)
		if ch.InRoom.Description != "" {
			d.WriteToBuffer(ch.InRoom.Description)
		}
		d.WriteToBuffer("\n\r")
	}

	// Notify others in room
	if ch.InRoom != nil {
		for _, rch := range ch.InRoom.People {
			if rch != ch && rch.Desc != nil {
				rch.Sendf("%s has entered the game.\n\r", ch.Name)
			}
		}
	}
}

// flushOutput sends buffered output for all descriptors.
func (g *GameLoop) flushOutput() {
	for _, d := range g.world.Descriptors {
		if err := d.FlushOutput(); err != nil {
			log.Printf("Flush error for %s: %v", d.Host, err)
		}
	}
}

// cleanupDescriptors removes closed/quitting connections from the descriptor list.
func (g *GameLoop) cleanupDescriptors() {
	alive := make([]*types.DescriptorData, 0, len(g.world.Descriptors))
	for _, d := range g.world.Descriptors {
		if d.Conn == nil || d.Connected == -1 {
			g.closeDescriptor(d)
			continue
		}
		alive = append(alive, d)
	}
	g.world.Descriptors = alive
}

// closeDescriptor disconnects and cleans up a descriptor.
func (g *GameLoop) closeDescriptor(d *types.DescriptorData) {
	if d.Character != nil {
		ch := d.Character
		log.Printf("%s has left the game.", ch.Name)

		// Remove from room
		if ch.InRoom != nil {
			for _, rch := range ch.InRoom.People {
				if rch != ch && rch.Desc != nil {
					rch.Sendf("%s has left the game.\n\r", ch.Name)
				}
			}
			for i, p := range ch.InRoom.People {
				if p == ch {
					ch.InRoom.People = append(ch.InRoom.People[:i], ch.InRoom.People[i+1:]...)
					break
				}
			}
		}

		// Remove from world
		g.world.RemoveChar(ch)
		d.Character = nil
	}

	if d.Conn != nil {
		_ = d.Conn.Close()
		d.Conn = nil
	}
}

// Stub update functions -- implementations will follow in later modules.

func (g *GameLoop) areaUpdate()     {}
func (g *GameLoop) violenceUpdate() {}
func (g *GameLoop) mobileUpdate()   {}
func (g *GameLoop) charUpdate()     {}
func (g *GameLoop) objUpdate()      {}
