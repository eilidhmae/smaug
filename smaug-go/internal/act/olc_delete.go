package act

import (
	"strconv"
	"strings"
	"time"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

// Deletion confirmation window. If the immortal issues the same <kind>/<vnum>
// delete command twice within this window, the delete goes through. Otherwise
// the first invocation just arms the confirmation.
const deleteConfirmWindow = 10 * time.Second

// confirmDelete returns true if this invocation matches a prior pending
// delete-confirmation for the same kind+vnum; otherwise it arms the state and
// returns false. NowFunc is swappable for tests.
var nowFunc = func() time.Time { return time.Now() }

func confirmDelete(ch *types.CharData, kind string, vnum int) bool {
	now := nowFunc()
	if ch.LastDeleteKind == kind && ch.LastDeleteVnum == vnum {
		if now.Sub(ch.LastDeleteTime) <= deleteConfirmWindow {
			// confirmed; clear state
			ch.LastDeleteKind = ""
			ch.LastDeleteVnum = 0
			return true
		}
	}
	ch.LastDeleteKind = kind
	ch.LastDeleteVnum = vnum
	ch.LastDeleteTime = now
	return false
}

// DoRdelete implements the 'rdelete' command: delete a room index entry.
// Requires two invocations within deleteConfirmWindow to commit the delete.
func DoRdelete(ch *types.CharData, argument string) {
	if ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}
	arg := strings.TrimSpace(argument)
	if arg == "" {
		ch.Send("Usage: rdelete <vnum>\n\r")
		return
	}
	vnum, err := strconv.Atoi(arg)
	if err != nil || vnum <= 0 {
		ch.Send("Vnum must be a positive number.\n\r")
		return
	}
	room := WorldRef.GetRoom(vnum)
	if room == nil {
		ch.Sendf("Room %d does not exist.\n\r", vnum)
		return
	}
	if ch.InRoom == room {
		ch.Send("You can't delete the room you are standing in.\n\r")
		return
	}
	if !confirmDelete(ch, "room", vnum) {
		ch.Sendf("Type 'rdelete %d' again within 10 seconds to confirm.\n\r", vnum)
		return
	}

	// Extract all characters and objects in this room.
	for len(room.People) > 0 {
		victim := room.People[len(room.People)-1]
		if victim.IsNPC() {
			handler.ExtractChar(WorldRef, victim, true)
		} else {
			// Transfer players to a safe temple (1001 or 3001) if available.
			safe := WorldRef.GetRoom(types.ROOM_VNUM_TEMPLE)
			if safe == nil || safe == room {
				handler.CharFromRoom(victim)
				break
			}
			handler.CharFromRoom(victim)
			handler.CharToRoom(victim, safe)
		}
	}
	for len(room.Contents) > 0 {
		handler.ExtractObj(WorldRef, room.Contents[len(room.Contents)-1])
	}

	// Remove exits pointing to this room from all other rooms.
	for _, other := range WorldRef.Rooms {
		if other == room {
			continue
		}
		filtered := make([]*types.ExitData, 0, len(other.Exits))
		for _, ex := range other.Exits {
			if ex.ToRoom == room {
				continue
			}
			filtered = append(filtered, ex)
		}
		other.Exits = filtered
	}

	delete(WorldRef.Rooms, vnum)
	ch.Sendf("Room %d deleted.\n\r", vnum)
}

// DoOdelete implements the 'odelete' command: delete an object index entry,
// scrubbing all live instances.
func DoOdelete(ch *types.CharData, argument string) {
	if ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}
	arg := strings.TrimSpace(argument)
	if arg == "" {
		ch.Send("Usage: odelete <vnum>\n\r")
		return
	}
	vnum, err := strconv.Atoi(arg)
	if err != nil || vnum <= 0 {
		ch.Send("Vnum must be a positive number.\n\r")
		return
	}
	idx, ok := WorldRef.ObjIndex[vnum]
	if !ok {
		ch.Sendf("Object %d does not exist.\n\r", vnum)
		return
	}
	if !confirmDelete(ch, "obj", vnum) {
		ch.Sendf("Type 'odelete %d' again within 10 seconds to confirm.\n\r", vnum)
		return
	}

	// Extract all live instances by walking a snapshot of the objects list.
	snapshot := make([]*types.ObjData, len(WorldRef.Objects))
	copy(snapshot, WorldRef.Objects)
	for _, obj := range snapshot {
		if obj.IndexData == idx {
			handler.ExtractObj(WorldRef, obj)
		}
	}

	delete(WorldRef.ObjIndex, vnum)
	ch.Sendf("Object %d deleted.\n\r", vnum)
}

// DoMdelete implements the 'mdelete' command: delete a mob index entry,
// scrubbing all live instances.
func DoMdelete(ch *types.CharData, argument string) {
	if ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}
	arg := strings.TrimSpace(argument)
	if arg == "" {
		ch.Send("Usage: mdelete <vnum>\n\r")
		return
	}
	vnum, err := strconv.Atoi(arg)
	if err != nil || vnum <= 0 {
		ch.Send("Vnum must be a positive number.\n\r")
		return
	}
	idx, ok := WorldRef.MobIndex[vnum]
	if !ok {
		ch.Sendf("Mob %d does not exist.\n\r", vnum)
		return
	}
	if !confirmDelete(ch, "mob", vnum) {
		ch.Sendf("Type 'mdelete %d' again within 10 seconds to confirm.\n\r", vnum)
		return
	}

	// Extract all live NPC instances.
	snapshot := make([]*types.CharData, len(WorldRef.Characters))
	copy(snapshot, WorldRef.Characters)
	for _, mob := range snapshot {
		if mob.IsNPC() && mob.IndexData == idx {
			handler.ExtractChar(WorldRef, mob, true)
		}
	}

	delete(WorldRef.MobIndex, vnum)
	ch.Sendf("Mob %d deleted.\n\r", vnum)
}

