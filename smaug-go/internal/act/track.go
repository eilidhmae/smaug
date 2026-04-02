package act

import (
	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

const (
	BFS_ERROR        = -1
	BFS_ALREADY_THERE = -2
	BFS_NO_PATH      = -3
)

var trackDirNames = []string{
	"north", "east", "south", "west", "up", "down",
	"northeast", "northwest", "southeast", "southwest",
}

type bfsNode struct {
	room     *types.RoomIndexData
	firstDir int // direction taken from the start room
}

// BFSFindPath finds the first step direction from src to target using BFS.
// Returns a direction (0-9), BFS_ALREADY_THERE, BFS_NO_PATH, or BFS_ERROR.
func BFSFindPath(src, target *types.RoomIndexData, maxDist int) int {
	if src == nil || target == nil {
		return BFS_ERROR
	}
	if src == target {
		return BFS_ALREADY_THERE
	}
	if src.Area != target.Area {
		return BFS_NO_PATH
	}

	visited := map[int]bool{src.Vnum: true}
	var queue []bfsNode

	// Enqueue first steps
	for _, ex := range src.Exits {
		if ex.ToRoom != nil && !visited[ex.ToRoom.Vnum] {
			visited[ex.ToRoom.Vnum] = true
			if ex.ToRoom == target {
				return ex.Direction
			}
			queue = append(queue, bfsNode{room: ex.ToRoom, firstDir: ex.Direction})
		}
	}

	count := 0
	for len(queue) > 0 {
		count++
		if count > maxDist {
			return BFS_NO_PATH
		}

		node := queue[0]
		queue = queue[1:]

		for _, ex := range node.room.Exits {
			if ex.ToRoom != nil && !visited[ex.ToRoom.Vnum] {
				visited[ex.ToRoom.Vnum] = true
				if ex.ToRoom == target {
					return node.firstDir
				}
				queue = append(queue, bfsNode{room: ex.ToRoom, firstDir: node.firstDir})
			}
		}
	}

	return BFS_NO_PATH
}

// DoTrack implements the 'track' command.
func DoTrack(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Whom are you trying to track?\n\r")
		return
	}

	victim := handler.GetCharWorld(WorldRef, ch, arg)
	if victim == nil {
		ch.Send("You can't find a trail of anyone like that.\n\r")
		return
	}

	maxDist := 100 + ch.Level*30

	dir := BFSFindPath(ch.InRoom, victim.InRoom, maxDist)

	switch dir {
	case BFS_ERROR:
		ch.Send("Hmm... something seems to be wrong.\n\r")
	case BFS_ALREADY_THERE:
		ch.Send("You're already in the same room!\n\r")
	case BFS_NO_PATH:
		ch.Send("You can't sense a trail from here.\n\r")
	default:
		if dir >= 0 && dir < len(trackDirNames) {
			ch.Sendf("You sense a trail %s from here...\n\r", trackDirNames[dir])
		}
	}
}

// HuntVictim moves an NPC one room toward its hunting target.
func HuntVictim(w *world.World, ch *types.CharData) {
	if ch == nil || ch.Hunting == nil || ch.Hunting.Who == nil || ch.Position < types.POS_RESTING {
		return
	}

	victim := ch.Hunting.Who

	// Verify victim still exists in world
	found := false
	for _, c := range w.Characters {
		if c == victim {
			found = true
			break
		}
	}
	if !found {
		ch.Hunting = nil
		return
	}

	// Already in same room — attack
	if ch.InRoom == victim.InRoom {
		if ch.Fighting != nil {
			return
		}
		ch.Hunting = nil
		combat.StartFighting(ch, victim)
		if victim.Fighting == nil {
			combat.StartFighting(victim, ch)
		}
		return
	}

	dir := BFSFindPath(ch.InRoom, victim.InRoom, 500+ch.Level*25)
	if dir < 0 {
		ch.Hunting = nil
		return
	}

	MoveChar(ch, dir)
}
