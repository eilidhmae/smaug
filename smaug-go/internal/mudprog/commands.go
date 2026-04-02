package mudprog

import (
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
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
