package game

import (
	"context"
	"crypto/subtle"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/eilidhmae/smaug/internal/command"
	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// MaxConnections is the maximum number of simultaneous connections allowed.
const MaxConnections = 256

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
	pulseSave     int
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
		g.aggrUpdate()
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

	g.pulseSave--
	if g.pulseSave <= 0 {
		g.pulseSave = types.PULSE_SAVE
		g.autosave()
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
			if len(g.world.Descriptors) >= MaxConnections {
				d.WriteToBuffer("Server is full, try again later.\n\r")
				d.Connected = -1
				log.Printf("Connection from %s rejected: server full", d.Host)
				continue
			}
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

			// Pager takes priority: if paging, handle pager input
			if d.HasPagerData() {
				SetPagerInput(d, line)
				continue
			}

			switch d.Connected {
			case types.CON_PLAYING:
				if d.Character != nil {
					g.cmdReg.Interpret(d.Character, line)
					// Send prompt after command output
					if d.Connected == types.CON_PLAYING {
						d.WriteToBuffer(FormatPrompt(d.Character))
					}
				}
			case types.CON_EDITING:
				if d.Character != nil {
					EditBuffer(d.Character, line)
				}
			default:
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
		g.nannyGetName(d, line)
	case types.CON_GET_OLD_PASSWORD:
		g.nannyGetOldPassword(d, line)
	case types.CON_CONFIRM_NEW_NAME:
		g.nannyConfirmNewName(d, line)
	case types.CON_GET_NEW_PASSWORD:
		g.nannyGetNewPassword(d, line)
	case types.CON_CONFIRM_NEW_PASSWORD:
		g.nannyConfirmNewPassword(d, line)
	case types.CON_GET_NEW_SEX:
		g.nannyGetNewSex(d, line)
	case types.CON_GET_NEW_CLASS:
		g.nannyGetNewClass(d, line)
	case types.CON_GET_NEW_RACE:
		g.nannyGetNewRace(d, line)
	case types.CON_READ_MOTD:
		g.enterGame(d)
	default:
		d.WriteToBuffer("Unexpected state. Disconnecting.\n\r")
		d.Connected = -1
	}
}

// nannyGetName handles name entry at login.
func (g *GameLoop) nannyGetName(d *types.DescriptorData, line string) {
	name := strings.TrimSpace(line)
	if name == "" {
		d.WriteToBuffer("By what name do you wish to be known? ")
		return
	}

	// Validate name: letters only, 3-12 characters
	if !isValidName(name) {
		d.WriteToBuffer("Illegal name, try another.\n\rName: ")
		return
	}

	// Capitalize
	name = strings.ToUpper(name[:1]) + strings.ToLower(name[1:])
	d.User = name

	// Check if already playing
	for _, od := range g.world.Descriptors {
		if od != d && od.Character != nil && strings.EqualFold(od.Character.Name, name) {
			d.WriteToBuffer("That character is already playing. Try another name.\n\rName: ")
			return
		}
	}

	// Check for existing player file
	playerPath := persist.PlayerFilePath(g.world.DataDir, name)
	if _, err := os.Stat(playerPath); err == nil {
		// Existing player — load and ask for password
		f, err := os.Open(playerPath)
		if err != nil {
			log.Printf("Error opening player file %s: %v", playerPath, err)
			d.WriteToBuffer("Error loading your character. Try again.\n\rName: ")
			return
		}
		ch, err := persist.LoadPlayerWithWorld(f, playerPath, g.world.GetObjIndex)
		f.Close()
		if err != nil {
			log.Printf("Error loading player %s: %v", name, err)
			d.WriteToBuffer("Error loading your character. Try again.\n\rName: ")
			return
		}
		d.Character = ch
		ch.Desc = d
		_, _ = d.Conn.Write(telnetEchoOff)
		d.WriteToBuffer("Password: ")
		d.Connected = types.CON_GET_OLD_PASSWORD
	} else {
		// New player — confirm the name
		d.WriteToBufferf("Did I get that right, %s (Y/N)? ", name)
		d.Connected = types.CON_CONFIRM_NEW_NAME
	}
}

// nannyGetOldPassword verifies password for returning players.
func (g *GameLoop) nannyGetOldPassword(d *types.DescriptorData, line string) {
	_, _ = d.Conn.Write(telnetEchoOn)
	d.WriteToBuffer("\n\r")

	if d.Character == nil || d.Character.PCData == nil {
		d.WriteToBuffer("Error: no character data. Disconnecting.\n\r")
		d.Connected = -1
		return
	}

	// Check password: bcrypt if stored hash starts with $2, otherwise legacy plaintext
	storedPwd := d.Character.PCData.Pwd
	passwordOK := false
	if strings.HasPrefix(storedPwd, "$2a$") || strings.HasPrefix(storedPwd, "$2b$") {
		// Bcrypt hash comparison (constant-time internally)
		passwordOK = bcrypt.CompareHashAndPassword([]byte(storedPwd), []byte(line)) == nil
	} else {
		// Legacy plaintext comparison using constant-time compare to prevent timing attacks
		passwordOK = subtle.ConstantTimeCompare([]byte(line), []byte(storedPwd)) == 1
		if passwordOK {
			// Migrate legacy plaintext password to bcrypt
			if hash, err := bcrypt.GenerateFromPassword([]byte(line), bcrypt.DefaultCost); err == nil {
				d.Character.PCData.Pwd = string(hash)
			}
		}
	}

	if !passwordOK {
		d.WriteToBuffer("Wrong password.\n\r")
		log.Printf("Bad password for %s from %s", d.User, d.Host)
		d.FailedAttempts++
		if d.FailedAttempts >= 3 {
			d.WriteToBuffer("Too many failed attempts. Disconnecting.\n\r")
		}
		d.Character = nil
		d.Connected = -1
		return
	}

	// Record the login site
	d.Character.PCData.RecentSite = d.Host

	d.WriteToBuffer(motd)
	d.Connected = types.CON_READ_MOTD
}

// nannyConfirmNewName handles "Did I get that right, Gandalf (Y/N)?"
func (g *GameLoop) nannyConfirmNewName(d *types.DescriptorData, line string) {
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		d.WriteToBufferf("Did I get that right, %s (Y/N)? ", d.User)
		return
	}

	switch strings.ToUpper(line[:1]) {
	case "Y":
		_, _ = d.Conn.Write(telnetEchoOff)
		d.WriteToBuffer("New character.\n\rGive me a password for this character: ")
		d.Connected = types.CON_GET_NEW_PASSWORD
	case "N":
		d.WriteToBuffer("Ok, what IS it, then? ")
		d.User = ""
		d.Connected = types.CON_GET_NAME
	default:
		d.WriteToBuffer("Please type Yes or No: ")
	}
}

// nannyGetNewPassword handles initial password entry for new characters.
func (g *GameLoop) nannyGetNewPassword(d *types.DescriptorData, line string) {
	d.WriteToBuffer("\n\r")

	if len(line) < 5 {
		d.WriteToBuffer("Password must be at least five characters long.\n\rPassword: ")
		return
	}

	if strings.Contains(line, "~") {
		d.WriteToBuffer("New password not acceptable, try again.\n\rPassword: ")
		return
	}

	// Create the character now with the password (bcrypt hashed)
	hash, err := bcrypt.GenerateFromPassword([]byte(line), bcrypt.DefaultCost)
	if err != nil {
		d.WriteToBuffer("Error hashing password. Try again.\n\rPassword: ")
		return
	}
	ch := g.createNewCharacter(d.User)
	ch.PCData.Pwd = string(hash)
	ch.Desc = d
	d.Character = ch

	d.WriteToBuffer("Please retype the password to confirm: ")
	d.Connected = types.CON_CONFIRM_NEW_PASSWORD
}

// nannyConfirmNewPassword handles password confirmation.
func (g *GameLoop) nannyConfirmNewPassword(d *types.DescriptorData, line string) {
	_, _ = d.Conn.Write(telnetEchoOn)
	d.WriteToBuffer("\n\r")

	if d.Character == nil || d.Character.PCData == nil {
		d.WriteToBuffer("Error: no character. Disconnecting.\n\r")
		d.Connected = -1
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(d.Character.PCData.Pwd), []byte(line)) != nil {
		d.WriteToBuffer("Passwords don't match.\n\rRetype password: ")
		_, _ = d.Conn.Write(telnetEchoOff)
		d.Connected = types.CON_GET_NEW_PASSWORD
		return
	}

	d.WriteToBuffer("\n\rWhich gender will your character be?\n\r (M)ale\n\r (F)emale\n\r (N)eutral\n\rPlease select: ")
	d.Connected = types.CON_GET_NEW_SEX
}

// nannyGetNewSex handles sex selection for new characters.
func (g *GameLoop) nannyGetNewSex(d *types.DescriptorData, line string) {
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		d.WriteToBuffer("Please select (M)ale, (F)emale, or (N)eutral: ")
		return
	}

	switch strings.ToUpper(line[:1]) {
	case "M":
		d.Character.Sex = types.SEX_MALE
	case "F":
		d.Character.Sex = types.SEX_FEMALE
	case "N":
		d.Character.Sex = types.SEX_NEUTRAL
	default:
		d.WriteToBuffer("That's not a valid selection.\n\rPlease select (M)ale, (F)emale, or (N)eutral: ")
		return
	}

	g.showClassMenu(d)
	d.Connected = types.CON_GET_NEW_CLASS
}

// nannyGetNewClass handles class selection for new characters.
func (g *GameLoop) nannyGetNewClass(d *types.DescriptorData, line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		g.showClassMenu(d)
		return
	}

	classIdx := -1
	for i, c := range g.world.Classes {
		if c == nil || c.WhoName == "" || strings.EqualFold(c.WhoName, "unused") {
			continue
		}
		if strings.EqualFold(c.WhoName, line) || (len(line) >= 1 && strings.HasPrefix(strings.ToLower(c.WhoName), strings.ToLower(line))) {
			classIdx = i
			break
		}
	}

	if classIdx < 0 {
		d.WriteToBuffer("That's not a valid class.\n\r")
		g.showClassMenu(d)
		return
	}

	d.Character.Class = classIdx

	g.showRaceMenu(d)
	d.Connected = types.CON_GET_NEW_RACE
}

// nannyGetNewRace handles race selection for new characters.
func (g *GameLoop) nannyGetNewRace(d *types.DescriptorData, line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		g.showRaceMenu(d)
		return
	}

	raceIdx := -1
	for i, r := range g.world.Races {
		if r == nil || r.Name == "" || strings.EqualFold(r.Name, "unused") {
			continue
		}
		if strings.EqualFold(r.Name, line) || (len(line) >= 1 && strings.HasPrefix(strings.ToLower(r.Name), strings.ToLower(line))) {
			// Check class restriction
			if r.ClassRestriction != 0 && (r.ClassRestriction&(1<<d.Character.Class)) != 0 {
				d.WriteToBuffer("That race is not available for your class.\n\r")
				g.showRaceMenu(d)
				return
			}
			raceIdx = i
			break
		}
	}

	if raceIdx < 0 {
		d.WriteToBuffer("That's not a valid race.\n\r")
		g.showRaceMenu(d)
		return
	}

	d.Character.Race = raceIdx

	// Apply race stat bonuses
	g.applyRaceBonuses(d.Character, g.world.Races[raceIdx])

	log.Printf("%s@%s new %s %s.",
		d.Character.Name, d.Host,
		g.world.Races[raceIdx].Name,
		g.world.Classes[d.Character.Class].WhoName)

	d.WriteToBuffer(motd)
	d.Connected = types.CON_READ_MOTD
}

// showClassMenu displays available classes.
func (g *GameLoop) showClassMenu(d *types.DescriptorData) {
	d.WriteToBuffer("\n\rSelect a class:\n\r")
	for _, c := range g.world.Classes {
		if c == nil || c.WhoName == "" || strings.EqualFold(c.WhoName, "unused") {
			continue
		}
		d.WriteToBufferf("  %s\n\r", c.WhoName)
	}
	d.WriteToBuffer("Choice: ")
}

// showRaceMenu displays available races (filtered by class restriction).
func (g *GameLoop) showRaceMenu(d *types.DescriptorData) {
	d.WriteToBuffer("\n\rSelect a race:\n\r")
	for _, r := range g.world.Races {
		if r == nil || r.Name == "" || strings.EqualFold(r.Name, "unused") {
			continue
		}
		// Skip races restricted for this class
		if r.ClassRestriction != 0 && (r.ClassRestriction&(1<<d.Character.Class)) != 0 {
			continue
		}
		d.WriteToBufferf("  %s\n\r", r.Name)
	}
	d.WriteToBuffer("Choice: ")
}

// createNewCharacter builds a fresh CharData for a new player.
func (g *GameLoop) createNewCharacter(name string) *types.CharData {
	ch := &types.CharData{
		Name:       name,
		ShortDescr: name,
		LongDescr:  name + " is here.\n\r",
		Level:      1,
		Position:   types.POS_STANDING,
		Hit:        20,
		MaxHit:     20,
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
		Gold:       0,
		Armor:      100,
		PCData: &types.PCData{
			Title:     "the newbie",
			Prompt:    "<%hhp %mm %vmv> ",
			Filename:  strings.ToLower(name),
			PagerLen:  24,
			Condition: [4]int{48, 48, 48, 0}, // full food/drink/blood
		},
	}
	ch.Act.Set(types.PLR_AUTOEXIT)
	ch.Act.Set(types.PLR_ANSI)
	return ch
}

// applyRaceBonuses applies racial stat modifiers to a new character.
func (g *GameLoop) applyRaceBonuses(ch *types.CharData, race *types.RaceData) {
	ch.PermStr += race.StrPlus
	ch.PermDex += race.DexPlus
	ch.PermWis += race.WisPlus
	ch.PermInt += race.IntPlus
	ch.PermCon += race.ConPlus
	ch.PermCha += race.ChaPlus
	ch.PermLck += race.LckPlus
	ch.AffectedBy = ch.AffectedBy.Or(race.Affected)
	ch.Resistant = race.Resist
	ch.Susceptible = race.Suscept
}

// enterGame places a character into the game world.
// For returning players, d.Character is already loaded from the save file.
// For new players, d.Character was built during character creation.
func (g *GameLoop) enterGame(d *types.DescriptorData) {
	ch := d.Character
	if ch == nil {
		// Should not happen, but guard against it
		d.WriteToBuffer("Error: no character data. Disconnecting.\n\r")
		d.Connected = -1
		return
	}

	ch.Desc = d
	d.Connected = types.CON_PLAYING

	// Add to world character list
	g.world.AddChar(ch)

	// Determine starting room
	startVnum := types.ROOM_VNUM_TEMPLE
	if ch.HomeVnum != 0 {
		startVnum = ch.HomeVnum
	}
	startRoom := g.world.GetRoom(startVnum)
	if startRoom == nil {
		startRoom = g.world.GetRoom(types.ROOM_VNUM_TEMPLE)
	}
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

	ch.Position = types.POS_STANDING

	log.Printf("%s has entered the game from %s", ch.Name, d.Host)

	d.WriteToBuffer("\n\rWelcome to SMAUG!\n\r\n\r")

	// Auto-look
	if ch.InRoom != nil {
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

	// Send initial prompt
	d.WriteToBuffer(FormatPrompt(ch))
}

// SavePlayer saves a character's data to disk.
func (g *GameLoop) SavePlayer(ch *types.CharData) {
	if ch == nil || ch.PCData == nil || ch.IsNPC() {
		return
	}
	playerPath := persist.PlayerFilePath(g.world.DataDir, ch.Name)
	if playerPath == "" {
		return
	}

	// Ensure the directory exists
	dir := filepath.Dir(playerPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Error creating player directory %s: %v", dir, err)
		return
	}

	tmpPath := playerPath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		log.Printf("Error saving player %s: %v", ch.Name, err)
		return
	}

	if err := persist.SavePlayer(f, ch); err != nil {
		f.Close()
		os.Remove(tmpPath)
		log.Printf("Error writing player %s: %v", ch.Name, err)
		return
	}
	f.Close()

	if err := os.Rename(tmpPath, playerPath); err != nil {
		os.Remove(tmpPath)
		log.Printf("Error renaming player file %s: %v", ch.Name, err)
	}
}

// isValidName checks if a name is acceptable for a player character.
func isValidName(name string) bool {
	if len(name) < 3 || len(name) > 12 {
		return false
	}
	for _, c := range name {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') {
			return false
		}
	}
	return true
}

// flushOutput sends buffered output for all descriptors.
func (g *GameLoop) flushOutput() {
	for _, d := range g.world.Descriptors {
		// If pager has data, process pager output instead of normal flush
		if d.HasPagerData() {
			PagerOutput(d)
		}

		if err := d.FlushOutput(); err != nil {
			log.Printf("Lost connection to %s: %v", d.Host, err)
			d.Connected = -1 // Mark for cleanup
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
		g.SavePlayer(ch)
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
