package mudprog

import (
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// WorldRef is set from main to give mudprogs access to the world.
// Written once at boot before the game loop starts; read only from the game loop goroutine. Safe without synchronization.
var WorldRef *world.World

// mpEcho sends a message to the mob's room.
func mpEcho(mob *types.CharData, args string) {
	msg := strings.TrimSpace(args)
	if mob.InRoom == nil || msg == "" {
		return
	}
	for _, rch := range mob.InRoom.People {
		if rch.Desc != nil {
			rch.Sendf("%s\n\r", msg)
		}
	}
}

// mpEchoAt sends a message to a specific character.
func mpEchoAt(mob *types.CharData, args string) {
	arg, msg := firstWord(strings.TrimSpace(args))
	if arg == "" || msg == "" {
		return
	}
	victim := handler.GetCharRoom(mob, arg)
	if victim != nil {
		victim.Sendf("%s\n\r", msg)
	}
}

// mpEchoAround sends a message to everyone in the room except the target.
func mpEchoAround(mob *types.CharData, args string) {
	arg, msg := firstWord(strings.TrimSpace(args))
	if arg == "" || msg == "" || mob.InRoom == nil {
		return
	}
	victim := handler.GetCharRoom(mob, arg)
	for _, rch := range mob.InRoom.People {
		if rch != victim && rch.Desc != nil {
			rch.Sendf("%s\n\r", msg)
		}
	}
}

// mpGoto moves the mob to a room by vnum.
func mpGoto(mob *types.CharData, args string) {
	arg := strings.TrimSpace(args)
	vnum, err := strconv.Atoi(strings.TrimSpace(arg))
	if err != nil || WorldRef == nil {
		return
	}
	room := WorldRef.GetRoom(vnum)
	if room == nil {
		return
	}
	if mob.InRoom != nil {
		handler.CharFromRoom(mob)
	}
	handler.CharToRoom(mob, room)
}

// mpTransfer moves a character to the mob's room.
func mpTransfer(mob *types.CharData, args string) {
	arg, rest := firstWord(strings.TrimSpace(args))
	if arg == "" || WorldRef == nil {
		return
	}

	var dest *types.RoomIndexData
	if rest != "" {
		vnum, err := strconv.Atoi(strings.TrimSpace(rest))
		if err == nil {
			dest = WorldRef.GetRoom(vnum)
		}
	}
	if dest == nil {
		dest = mob.InRoom
	}
	if dest == nil {
		return
	}

	victim := handler.GetCharWorld(WorldRef, mob, arg)
	if victim == nil {
		return
	}
	if victim.InRoom != nil {
		handler.CharFromRoom(victim)
	}
	handler.CharToRoom(victim, dest)
}

// mpForce forces a character to execute a command.
func mpForce(mob *types.CharData, args string) {
	arg, cmd := firstWord(strings.TrimSpace(args))
	if arg == "" || cmd == "" {
		return
	}

	victim := handler.GetCharWorld(WorldRef, mob, arg)
	if victim == nil {
		return
	}
	if CmdRegistry != nil {
		CmdRegistry.Interpret(victim, cmd)
	}
}

// mpKill starts the mob fighting a character.
func mpKill(mob *types.CharData, args string) {
	arg := strings.TrimSpace(args)
	if arg == "" {
		return
	}
	victim := handler.GetCharRoom(mob, arg)
	if victim == nil || victim == mob {
		return
	}
	// Simplified: just start fighting
	if mob.Fighting == nil {
		mob.Fighting = &types.FightData{Who: victim}
		mob.Position = types.POS_FIGHTING
	}
}

// mpDamage deals damage to a character.
func mpDamage(mob *types.CharData, args string) {
	arg, rest := firstWord(strings.TrimSpace(args))
	if arg == "" || rest == "" {
		return
	}
	victim := handler.GetCharRoom(mob, arg)
	if victim == nil {
		return
	}
	dam, err := strconv.Atoi(strings.TrimSpace(rest))
	if err != nil || dam <= 0 {
		return
	}
	victim.Hit -= dam
	if victim.Hit < 1 {
		victim.Hit = 1 // mudprog damage doesn't kill
	}
}

// mpPurge removes an NPC or object from the room.
func mpPurge(mob *types.CharData, args string) {
	arg := strings.TrimSpace(args)
	if mob.InRoom == nil || WorldRef == nil {
		return
	}
	if arg == "" {
		// Purge all NPCs (except self) and objects
		for i := len(mob.InRoom.People) - 1; i >= 0; i-- {
			rch := mob.InRoom.People[i]
			if rch != mob && rch.IsNPC() {
				handler.ExtractChar(WorldRef, rch, true)
			}
		}
		for i := len(mob.InRoom.Contents) - 1; i >= 0; i-- {
			handler.ExtractObj(WorldRef, mob.InRoom.Contents[i])
		}
		return
	}

	victim := handler.GetCharRoom(mob, arg)
	if victim != nil && victim.IsNPC() && victim != mob {
		handler.ExtractChar(WorldRef, victim, true)
		return
	}
	obj := handler.GetObjHere(mob, arg)
	if obj != nil {
		handler.ExtractObj(WorldRef, obj)
	}
}

// mpAsound echoes a message to all adjacent rooms (connected via exits).
func mpAsound(mob *types.CharData, args string) {
	msg := strings.TrimSpace(args)
	if msg == "" || mob.InRoom == nil {
		return
	}
	for _, ex := range mob.InRoom.Exits {
		dst := ex.ToRoom
		if dst == nil || dst == mob.InRoom {
			continue
		}
		for _, rch := range dst.People {
			if rch.Desc != nil {
				rch.Sendf("%s\n\r", msg)
			}
		}
	}
}

// mpEchoZone echoes a message to all players in the mob's area.
func mpEchoZone(mob *types.CharData, args string) {
	msg := strings.TrimSpace(args)
	if mob.InRoom == nil || mob.InRoom.Area == nil || WorldRef == nil {
		return
	}
	area := mob.InRoom.Area
	for _, vch := range WorldRef.Characters {
		if vch.InRoom != nil && vch.InRoom.Area == area && vch.Desc != nil {
			if msg == "" {
				vch.Sendf(" \n\r")
			} else {
				vch.Sendf("%s\n\r", msg)
			}
		}
	}
}

// mpMload creates a mob from a vnum and places it in the mob's room.
func mpMload(mob *types.CharData, args string) {
	arg := strings.TrimSpace(args)
	vnum, err := strconv.Atoi(arg)
	if err != nil || WorldRef == nil || mob.InRoom == nil {
		return
	}
	idx := WorldRef.GetMobIndex(vnum)
	if idx == nil {
		return
	}
	victim := handler.CreateMobile(WorldRef, idx)
	if victim == nil {
		return
	}
	handler.CharToRoom(victim, mob.InRoom)
}

// mpOload creates an object from a vnum. Placed on mob (if ITEM_TAKE set) or on
// the mob's room floor. Supports optional level and timer arguments.
func mpOload(mob *types.CharData, args string) {
	arg1, rest := firstWord(strings.TrimSpace(args))
	vnum, err := strconv.Atoi(arg1)
	if err != nil || WorldRef == nil || mob.InRoom == nil {
		return
	}
	level := mob.GetTrust()
	timer := 0
	if rest != "" {
		arg2, rest2 := firstWord(rest)
		if arg2 != "" {
			lvl, err := strconv.Atoi(arg2)
			if err != nil || lvl < 0 || lvl > mob.GetTrust() {
				return
			}
			level = lvl
			if rest2 != "" {
				t, err := strconv.Atoi(strings.TrimSpace(rest2))
				if err != nil || t < 0 {
					return
				}
				timer = t
			}
		}
	}
	idx := WorldRef.GetObjIndex(vnum)
	if idx == nil {
		return
	}
	obj := handler.CreateObject(WorldRef, idx, level)
	if obj == nil {
		return
	}
	obj.Timer = timer
	if obj.WearFlags&int(types.ITEM_TAKE) != 0 {
		handler.ObjToChar(obj, mob)
	} else {
		handler.ObjToRoom(obj, mob.InRoom)
	}
}

// mpInvis toggles ACT_MOBINVIS on the mob. Optional numeric argument sets mobinvis level.
func mpInvis(mob *types.CharData, args string) {
	arg := strings.TrimSpace(args)
	if arg != "" {
		level, err := strconv.Atoi(arg)
		if err != nil || level < 2 || level > types.LEVEL_IMMORTAL {
			return
		}
		mob.MobInvis = level
		return
	}
	if mob.MobInvis < 2 {
		mob.MobInvis = mob.Level
	}
	if mob.Act.IsSet(types.ACT_MOBINVIS) {
		mob.Act.Remove(types.ACT_MOBINVIS)
	} else {
		mob.Act.Set(types.ACT_MOBINVIS)
	}
}

// mpAt runs a command as the mob at another room.
func mpAt(mob *types.CharData, args string) {
	arg, cmd := firstWord(strings.TrimSpace(args))
	if arg == "" || cmd == "" || WorldRef == nil {
		return
	}
	var location *types.RoomIndexData
	if vnum, err := strconv.Atoi(arg); err == nil {
		location = WorldRef.GetRoom(vnum)
	} else if v := handler.GetCharWorld(WorldRef, mob, arg); v != nil {
		location = v.InRoom
	}
	if location == nil {
		return
	}
	original := mob.InRoom
	if original == location {
		if CmdRegistry != nil {
			CmdRegistry.Interpret(mob, cmd)
		}
		return
	}
	if original != nil {
		handler.CharFromRoom(mob)
	}
	handler.CharToRoom(mob, location)
	if CmdRegistry != nil {
		CmdRegistry.Interpret(mob, cmd)
	}
	if mob.InRoom == location && original != nil {
		handler.CharFromRoom(mob)
		handler.CharToRoom(mob, original)
	}
}

// mpAdvance advances the victim's level by exactly one, never past
// LEVEL_AVATAR. Mirrors C do_mpadvance (mud_comm.c:1388). Syntax:
// mpadvance <victim>. Refuses to act if the victim already outranks the
// caller mob or is an NPC (C mud_comm.c:1416-1420). The >= LEVEL_AVATAR
// guard mirrors C's own `return` at mud_comm.c:1422 — the god-level help
// path in C is dead code after that return. Cap at LEVEL_SUPREME defends
// against any future change that removes the AVATAR gate.
func mpAdvance(mob *types.CharData, args string) {
	arg, _ := firstWord(strings.TrimSpace(args))
	if arg == "" {
		return
	}
	victim := handler.GetCharRoom(mob, arg)
	if victim == nil || victim.IsNPC() {
		return
	}
	if victim.Level >= types.LEVEL_AVATAR {
		return
	}
	if victim.Level > mob.Level {
		return
	}
	if victim.Level+1 > types.LEVEL_SUPREME {
		return
	}
	victim.Level = victim.Level + 1
}

// mpSlay instantly kills a mortal victim. Mirrors C do_mp_slay
// (mud_comm.c:2367): cannot slay self, supermob (vnum 3), or an immortal PC.
// C routes through raw_kill which in turn calls make_corpse; the Go port
// intentionally uses ExtractChar(fPull=true) here, which removes the victim
// WITHOUT creating a corpse and WITHOUT awarding XP (verified: handler.go:328).
// That mirrors the simplest interpretation of "mob just deletes the victim"
// — the corpse+cry+XP path is reserved for real combat death in combat.Damage.
// TODO(tier4): decide whether mpslay should leave a corpse for parity with C.
func mpSlay(mob *types.CharData, args string) {
	arg := strings.TrimSpace(args)
	if arg == "" {
		return
	}
	victim := handler.GetCharRoom(mob, arg)
	if victim == nil || victim == mob || WorldRef == nil {
		return
	}
	// Supermob guard (vnum 3) — C mud_comm.c:2401.
	if victim.IsNPC() && victim.IndexData != nil && victim.IndexData.Vnum == 3 {
		return
	}
	if !victim.IsNPC() && victim.Level >= types.LEVEL_IMMORTAL {
		return
	}
	handler.ExtractChar(WorldRef, victim, true)
}

// mpLog writes an entry to the mudprog log.
func mpLog(mob *types.CharData, args string) {
	msg := strings.TrimSpace(args)
	if msg == "" {
		return
	}
	util.LogString("Mudprog log (" + mob.ShortDescr + "): " + msg)
}

// mpRestore restores a victim's HP/Mana/Move to max.
func mpRestore(mob *types.CharData, args string) {
	arg, _ := firstWord(strings.TrimSpace(args))
	if arg == "" {
		return
	}
	victim := handler.GetCharRoom(mob, arg)
	if victim == nil {
		return
	}
	victim.Hit = victim.MaxHit
	victim.Mana = victim.MaxMana
	victim.Move = victim.MaxMove
}

// mpFavor adjusts a PC's favor.
func mpFavor(mob *types.CharData, args string) {
	arg, rest := firstWord(strings.TrimSpace(args))
	if arg == "" || rest == "" {
		return
	}
	victim := handler.GetCharRoom(mob, arg)
	if victim == nil || victim.IsNPC() || victim.PCData == nil {
		return
	}
	rest = strings.TrimSpace(rest)
	plus, minus := false, false
	if strings.HasPrefix(rest, "+") {
		plus = true
		rest = rest[1:]
	} else if strings.HasPrefix(rest, "-") {
		minus = true
		rest = rest[1:]
	}
	amt, err := strconv.Atoi(rest)
	if err != nil {
		return
	}
	cur := victim.PCData.Favor
	var nv int
	switch {
	case plus:
		nv = cur + amt
	case minus:
		nv = cur - amt
	default:
		nv = amt
	}
	if nv < -2500 {
		nv = -2500
	}
	if nv > 2500 {
		nv = 2500
	}
	victim.PCData.Favor = nv
}

// mpNuisance turns on or escalates a nuisance on a PC.
func mpNuisance(mob *types.CharData, args string) {
	arg := strings.TrimSpace(args)
	if arg == "" {
		return
	}
	victim := handler.GetCharRoom(mob, arg)
	if victim == nil || victim.IsNPC() || victim.PCData == nil {
		return
	}
	if victim.PCData.Nuisance == nil {
		victim.PCData.Nuisance = &types.NuisanceData{Flags: 1, Power: 2}
	} else {
		victim.PCData.Nuisance.Power++
		if victim.PCData.Nuisance.Power > types.MAX_NUISANCE_STAGE {
			victim.PCData.Nuisance.Power = types.MAX_NUISANCE_STAGE
		}
	}
}

// mpUnnuisance clears a PC's nuisance.
func mpUnnuisance(mob *types.CharData, args string) {
	arg := strings.TrimSpace(args)
	if arg == "" {
		return
	}
	victim := handler.GetCharRoom(mob, arg)
	if victim == nil || victim.IsNPC() || victim.PCData == nil {
		return
	}
	victim.PCData.Nuisance = nil
}

// mpBodybag scans the WORLD object list (not just the mob's room) for any
// PC corpse (vnum 11) matching "the corpse of <name>" lying in any room and
// teleports it to the mob's inventory. Mirrors C do_mpbodybag
// (mud_comm.c:1903).
func mpBodybag(mob *types.CharData, args string) {
	arg := strings.TrimSpace(args)
	if arg == "" || WorldRef == nil {
		return
	}
	target := "the corpse of " + arg
	// Collect first to avoid mutating the world list mid-iteration.
	var matches []*types.ObjData
	for _, obj := range WorldRef.Objects {
		if obj.InRoom == nil || obj.IndexData == nil {
			continue
		}
		if obj.IndexData.Vnum != 11 {
			continue
		}
		if !strings.EqualFold(obj.ShortDescr, target) {
			continue
		}
		matches = append(matches, obj)
	}
	for _, obj := range matches {
		room := obj.InRoom
		// Remove from room contents.
		for i, o := range room.Contents {
			if o == obj {
				room.Contents = append(room.Contents[:i], room.Contents[i+1:]...)
				break
			}
		}
		obj.InRoom = nil
		handler.ObjToChar(obj, mob)
		obj.Timer = -1
	}
}

// mpMorph: morph subsystem not wired. TODO(tier3): morph/unmorph wiring.
func mpMorph(mob *types.CharData, args string) {
	_ = mob
	_ = args
}

// mpUnmorph: morph subsystem not wired. TODO(tier3): morph/unmorph wiring.
func mpUnmorph(mob *types.CharData, args string) {
	_ = mob
	_ = args
}

// mpPractice sets a PC victim's learned proficiency in a skill to adept.
func mpPractice(mob *types.CharData, args string) {
	arg1, rest := firstWord(strings.TrimSpace(args))
	arg2, _ := firstWord(rest)
	if arg1 == "" || arg2 == "" || WorldRef == nil {
		return
	}
	victim := handler.GetCharRoom(mob, arg1)
	if victim == nil || victim.IsNPC() || victim.PCData == nil {
		return
	}
	sn := -1
	for i, sk := range WorldRef.Skills {
		if sk == nil {
			continue
		}
		if strings.EqualFold(sk.Name, arg2) {
			sn = i
			break
		}
	}
	if sn < 0 {
		return
	}
	adept := 0
	if victim.Class >= 0 && victim.Class < types.MAX_CLASS {
		adept = WorldRef.Skills[sn].SkillAdept[victim.Class]
	}
	if adept <= 0 {
		adept = 100
	}
	victim.PCData.Learned[sn] = adept
}

// mpOpenPassage creates a one-way passage exit from the mob's room.
// Syntax: mpopenpassage <dir> <dest-vnum>.
func mpOpenPassage(mob *types.CharData, args string) {
	dirArg, rest := firstWord(strings.TrimSpace(args))
	vnumArg, _ := firstWord(rest)
	if dirArg == "" || vnumArg == "" || mob.InRoom == nil || WorldRef == nil {
		return
	}
	dir, err := strconv.Atoi(dirArg)
	if err != nil || dir < 0 || dir > types.MAX_DIR {
		return
	}
	vnum, err := strconv.Atoi(vnumArg)
	if err != nil {
		return
	}
	dest := WorldRef.GetRoom(vnum)
	if dest == nil {
		return
	}
	if ex := mob.InRoom.GetExit(dir); ex != nil {
		// Existing exit — C version only allows overwrite of prior EX_PASSAGE
		// and otherwise returns.
		return
	}
	passage := &types.ExitData{
		ToRoom:    dest,
		Vnum:      vnum,
		Direction: dir,
		Key:       -1,
		ExitInfo:  int(types.EX_PASSAGE),
	}
	mob.InRoom.Exits = append(mob.InRoom.Exits, passage)
}

// mpClosePassage removes an EX_PASSAGE exit from the mob's room.
// Syntax: mpclosepassage <dir>.
func mpClosePassage(mob *types.CharData, args string) {
	dirArg, _ := firstWord(strings.TrimSpace(args))
	if dirArg == "" || mob.InRoom == nil {
		return
	}
	dir, err := strconv.Atoi(dirArg)
	if err != nil || dir < 0 || dir > types.MAX_DIR {
		return
	}
	for i, ex := range mob.InRoom.Exits {
		if ex.Direction != dir {
			continue
		}
		if uint32(ex.ExitInfo)&types.EX_PASSAGE == 0 {
			return
		}
		mob.InRoom.Exits = append(mob.InRoom.Exits[:i], mob.InRoom.Exits[i+1:]...)
		return
	}
}

// mpFillIn closes an existing door.
func mpFillIn(mob *types.CharData, args string) {
	dirArg, _ := firstWord(strings.TrimSpace(args))
	if dirArg == "" || mob.InRoom == nil {
		return
	}
	dir, err := strconv.Atoi(dirArg)
	if err != nil || dir < 0 || dir > types.MAX_DIR {
		return
	}
	ex := mob.InRoom.GetExit(dir)
	if ex == nil {
		return
	}
	ex.ExitInfo |= int(types.EX_CLOSED)
}

// mpPeace stops all combat in the mob's room.
func mpPeace(mob *types.CharData, args string) {
	_ = args
	if mob.InRoom == nil {
		return
	}
	for _, rch := range mob.InRoom.People {
		if rch.Fighting != nil {
			combat.StopFighting(rch, true)
		}
		rch.Hunting = nil
		rch.Hating = nil
		rch.Fearing = nil
	}
}

// mpPkset flips PCFLAG_DEADLY on a PC. Syntax: mppkset <player> <yes|no>.
func mpPkset(mob *types.CharData, args string) {
	arg, rest := firstWord(strings.TrimSpace(args))
	rest = strings.TrimSpace(rest)
	if arg == "" || rest == "" {
		return
	}
	victim := handler.GetCharRoom(mob, arg)
	if victim == nil || victim.IsNPC() || victim.PCData == nil {
		return
	}
	switch strings.ToLower(rest) {
	case "yes", "y":
		victim.PCData.Flags |= int(types.PCFLAG_DEADLY)
	case "no", "n":
		victim.PCData.Flags &^= int(types.PCFLAG_DEADLY)
	}
}

// mpOowner sets Owner on an object in the mob's room.
// Syntax: mpoowner <object> <name|none>.
func mpOowner(mob *types.CharData, args string) {
	objArg, rest := firstWord(strings.TrimSpace(args))
	nameArg, _ := firstWord(rest)
	if objArg == "" || nameArg == "" {
		return
	}
	obj := handler.GetObjHere(mob, objArg)
	if obj == nil {
		return
	}
	if strings.EqualFold(nameArg, "none") {
		obj.Owner = ""
		return
	}
	victim := handler.GetCharRoom(mob, nameArg)
	if victim == nil || victim.IsNPC() {
		return
	}
	obj.Owner = victim.Name
}

// mpHunt starts the mob hunting a character.
func mpHunt(mob *types.CharData, args string) {
	arg := strings.TrimSpace(args)
	if arg == "" || WorldRef == nil {
		return
	}
	victim := handler.GetCharWorld(WorldRef, mob, arg)
	if victim == nil {
		return
	}
	mob.Hunting = &types.HHFData{Name: victim.Name, Who: victim}
}

// mpHate sets Hating (single-slot). TODO(tier3): richer hate list.
func mpHate(mob *types.CharData, args string) {
	arg, _ := firstWord(strings.TrimSpace(args))
	if arg == "" || WorldRef == nil {
		return
	}
	victim := handler.GetCharWorld(WorldRef, mob, arg)
	if victim == nil {
		return
	}
	mob.Hating = &types.HHFData{Name: victim.Name, Who: victim}
}

// economyBillion mirrors the 1e9 chunk size used by C boost/lower_economy.
const economyBillion = 1000000000

// boostEconomy mirrors C boost_economy (handler.c:5641): each whole billion
// becomes a +1 to HighEconomy, the remainder accumulates in LowEconomy, and
// any LowEconomy overflow rolls forward into HighEconomy.
func boostEconomy(area *types.AreaData, gold int) {
	for gold >= economyBillion {
		area.HighEconomy++
		gold -= economyBillion
	}
	area.LowEconomy += gold
	for area.LowEconomy >= economyBillion {
		area.HighEconomy++
		area.LowEconomy -= economyBillion
	}
}

// lowerEconomy mirrors C lower_economy (handler.c:5660): each whole billion
// taken decrements HighEconomy, then the remainder is subtracted from
// LowEconomy, borrowing from HighEconomy when LowEconomy would go negative.
func lowerEconomy(area *types.AreaData, gold int) {
	for gold >= economyBillion {
		area.HighEconomy--
		gold -= economyBillion
	}
	area.LowEconomy -= gold
	for area.LowEconomy < 0 {
		area.HighEconomy--
		area.LowEconomy += economyBillion
	}
}

// mpDeposit moves gold from the mob into the area economy split fields
// (LowEconomy/HighEconomy). Mirrors C do_mpdeposit semantics by routing
// through boost_economy.
func mpDeposit(mob *types.CharData, args string) {
	arg := strings.TrimSpace(args)
	gold, err := strconv.Atoi(arg)
	if err != nil || gold <= 0 || mob.InRoom == nil || mob.InRoom.Area == nil {
		return
	}
	if mob.Gold < gold {
		return
	}
	mob.Gold -= gold
	boostEconomy(mob.InRoom.Area, gold)
}

// mpWithdraw moves gold from the area economy split fields into the mob.
// Mirrors C do_mpwithdraw via lower_economy. Refuses to act when the area
// does not have enough total gold (HighEconomy*1e9 + LowEconomy).
func mpWithdraw(mob *types.CharData, args string) {
	arg := strings.TrimSpace(args)
	gold, err := strconv.Atoi(arg)
	if err != nil || gold <= 0 || mob.InRoom == nil || mob.InRoom.Area == nil {
		return
	}
	area := mob.InRoom.Area
	// economy_has check (handler.c:5679): treat presence of any HighEconomy
	// as one billion of headroom (matches C bool semantics).
	hasHigh := 0
	if area.HighEconomy > 0 {
		hasHigh = 1
	}
	available := hasHigh*economyBillion + area.LowEconomy
	if available < gold {
		return
	}
	lowerEconomy(area, gold)
	mob.Gold += gold
}

// mpApply is a stub.
// TODO(tier3): mpapply auth state machine not ported.
func mpApply(mob *types.CharData, args string) {
	_ = mob
	_ = args
}

// mpApplyB is a stub.
// TODO(tier3): mpapplyb auth state machine not ported.
func mpApplyB(mob *types.CharData, args string) {
	_ = mob
	_ = args
}

// mpDelay adds a WAIT_STATE to a victim. Syntax: mpdelay <target> <rounds 1-30>.
func mpDelay(mob *types.CharData, args string) {
	arg, rest := firstWord(strings.TrimSpace(args))
	if arg == "" || rest == "" {
		return
	}
	victim := handler.GetCharRoom(mob, arg)
	if victim == nil {
		return
	}
	delay, err := strconv.Atoi(strings.TrimSpace(rest))
	if err != nil || delay < 1 || delay > 30 {
		return
	}
	if !victim.IsNPC() && victim.Level >= types.LEVEL_IMMORTAL {
		return
	}
	victim.Wait = delay * types.PULSE_VIOLENCE
}

// mpStrew spreads N copies of an object vnum across random rooms in the mob's
// area. Syntax: mpstrew <vnum> [count]. Default count 5.
func mpStrew(mob *types.CharData, args string) {
	arg, rest := firstWord(strings.TrimSpace(args))
	vnum, err := strconv.Atoi(arg)
	if err != nil || WorldRef == nil || mob.InRoom == nil || mob.InRoom.Area == nil {
		return
	}
	idx := WorldRef.GetObjIndex(vnum)
	if idx == nil {
		return
	}
	count := 5
	if rest != "" {
		if c, err := strconv.Atoi(strings.TrimSpace(rest)); err == nil && c > 0 {
			count = c
		}
	}
	area := mob.InRoom.Area
	var rooms []*types.RoomIndexData
	for _, r := range WorldRef.Rooms {
		if r.Area == area {
			rooms = append(rooms, r)
		}
	}
	if len(rooms) == 0 {
		return
	}
	if count > len(rooms) {
		count = len(rooms)
	}
	for i := 0; i < count; i++ {
		r := rooms[util.NumberRange(0, len(rooms)-1)]
		obj := handler.CreateObject(WorldRef, idx, mob.GetTrust())
		if obj != nil {
			handler.ObjToRoom(obj, r)
		}
	}
}

// mpScatter teleports a victim to a random room in [low_vnum, high_vnum].
// Syntax: mpscatter <victim> <low_vnum> <high_vnum>. Mirrors C do_mpscatter
// (mud_comm.c:2283): picks a random vnum in range that has a real room,
// moves the victim there, sets POS_RESTING, and stops any current fight.
func mpScatter(mob *types.CharData, args string) {
	if WorldRef == nil {
		return
	}
	arg1, rest := firstWord(strings.TrimSpace(args))
	arg2, rest := firstWord(rest)
	arg3, _ := firstWord(rest)
	if arg1 == "" || arg2 == "" || arg3 == "" {
		return
	}
	victim := handler.GetCharRoom(mob, arg1)
	if victim == nil {
		return
	}
	low, errL := strconv.Atoi(arg2)
	high, errH := strconv.Atoi(arg3)
	if errL != nil || errH != nil || low < 1 || high < low {
		return
	}
	// Build the candidate set of real rooms in [low, high].
	var candidates []*types.RoomIndexData
	for vnum := low; vnum <= high; vnum++ {
		if r := WorldRef.GetRoom(vnum); r != nil {
			candidates = append(candidates, r)
		}
	}
	if len(candidates) == 0 {
		return
	}
	dest := candidates[util.NumberRange(0, len(candidates)-1)]

	if victim.Fighting != nil {
		combat.StopFighting(victim, true)
	}
	if victim.InRoom != nil {
		handler.CharFromRoom(victim)
	}
	handler.CharToRoom(victim, dest)
	victim.Position = types.POS_RESTING
}

// mpDream sends a message to a sleeping target anywhere in the world.
func mpDream(mob *types.CharData, args string) {
	arg, rest := firstWord(strings.TrimSpace(args))
	if arg == "" || rest == "" || WorldRef == nil {
		return
	}
	victim := handler.GetCharWorld(WorldRef, mob, arg)
	if victim == nil {
		return
	}
	if victim.Position <= types.POS_SLEEPING {
		victim.Sendf("%s\n\r", rest)
	}
}

// mpNothing is an explicit no-op used for scripts.
func mpNothing(mob *types.CharData, args string) {
	_ = mob
	_ = args
}

// mpApplyAffect attaches an affect to a victim.
// Syntax: mpapplyaffect <target> <spell-slot> <location> <modifier> <duration>.
// NOTE: C's do_mpapply is an auth-state helper (stubbed above as mpApply).
// Per task spec, mpapply here is an affect-creator variant.
func mpApplyAffect(mob *types.CharData, args string) {
	arg1, rest := firstWord(strings.TrimSpace(args))
	arg2, rest := firstWord(rest)
	arg3, rest := firstWord(rest)
	arg4, rest := firstWord(rest)
	arg5, _ := firstWord(rest)
	if arg1 == "" || arg2 == "" || arg3 == "" || arg4 == "" || arg5 == "" {
		return
	}
	victim := handler.GetCharRoom(mob, arg1)
	if victim == nil {
		return
	}
	sn, err1 := strconv.Atoi(arg2)
	loc, err2 := strconv.Atoi(arg3)
	mod, err3 := strconv.Atoi(arg4)
	dur, err4 := strconv.Atoi(arg5)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return
	}
	aff := &types.AffectData{
		Type:     sn,
		Duration: dur,
		Location: loc,
		Modifier: mod,
	}
	handler.AffectToChar(victim, aff)
}
