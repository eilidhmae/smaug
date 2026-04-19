// Arena victory processing — splits off from the combat death path.
// When a PC dies to another PC in a ROOM_ARENA room, the arena branch
// fires INSTEAD of the normal corpse/XP processing. Loser goes to the
// altar, winner to the temple, both heal to max, both lose arena flags
// and debuffs, arena state is released.
//
// C ref: src/fight.c:2861-2915 (under #ifdef ENABLE_ARENA).
package combat

import (
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

// Seams populated at boot to break the act↔combat import cycle. See
// plan-phase6-arena.md §G6 Seams.
var (
	// ArenaIsBusyFunc writes to act.arenaState.IsBusy. Wired at boot
	// via `combat.ArenaIsBusyFunc = act.SetArenaIsBusy`.
	ArenaIsBusyFunc func(bool)
	// DoLookFunc performs the "auto" look after teleport. Wired at boot
	// via `combat.DoLookFunc = act.DoLook`.
	DoLookFunc func(*types.CharData, string)
)

// ArenaVictoryCheck handles PvP arena death: teleport loser to altar +
// winner to temple, heal both to full, strip debuffs, track
// akills/adeaths, clear arena flags. Returns true if the arena branch
// fired (caller should skip normal death processing).
//
// C ref: src/fight.c:2861-2915.
//
// Fires only when every condition holds:
//   - victim is dead (position == POS_DEAD)
//   - neither side is an NPC
//   - victim is in a ROOM_ARENA room (protects against non-arena PK)
func ArenaVictoryCheck(ch, victim *types.CharData) bool {
	if ch == nil || victim == nil {
		return false
	}
	if victim.Position != types.POS_DEAD {
		return false
	}
	if victim.IsNPC() || ch.IsNPC() {
		return false
	}
	if victim.InRoom == nil || !victim.InRoom.RoomFlags.IsSet(types.ROOM_ARENA) {
		return false
	}

	if victim.PCData != nil {
		victim.PCData.ADeaths++
	}
	if ch.PCData != nil {
		ch.PCData.AKills++
	}

	// --- Loser path ---
	StopFighting(victim, true)
	handler.CharFromRoom(victim)
	if WorldRef != nil {
		if altar := WorldRef.GetRoom(types.ROOM_VNUM_ALTAR); altar != nil {
			handler.CharToRoom(victim, altar)
		} else if temple := WorldRef.GetRoom(types.ROOM_VNUM_TEMPLE); temple != nil {
			// Fallback: if the altar vnum isn't loaded, the loser
			// still needs a destination. Temple is the only other
			// shipped post-death landing site. Plan Open Q5.
			handler.CharToRoom(victim, temple)
		}
	}
	victim.Send("You lost haha!\n\r")
	victim.Hit = victim.MaxHit
	victim.Mana = victim.MaxMana
	victim.Move = victim.MaxMove
	stripArenaDebuffs(victim)
	updatePos(victim)
	if DoLookFunc != nil {
		DoLookFunc(victim, "auto")
	}

	// --- Winner path ---
	StopFighting(ch, true)
	handler.CharFromRoom(ch)
	if WorldRef != nil {
		if temple := WorldRef.GetRoom(types.ROOM_VNUM_TEMPLE); temple != nil {
			handler.CharToRoom(ch, temple)
		}
	}
	ch.Send("You won are you sure you didn't cheat?!? heh\n\r")
	ch.Hit = ch.MaxHit
	ch.Mana = ch.MaxMana
	ch.Move = ch.MaxMove
	stripArenaDebuffs(ch)
	updatePos(ch)
	if DoLookFunc != nil {
		DoLookFunc(ch, "auto")
	}

	// --- Defensive flag cleanup on BOTH sides ---
	// C arena.c behavior is to set ACT_CHALLENGER on challenger only and
	// ACT_CHALLENGED on accepter only, so the clear needs to cover both
	// sides for both flags. See plan-phase6-arena.md §C-ref "Flag-set
	// asymmetry" + C fight.c:2900-2907.
	for _, p := range []*types.CharData{ch, victim} {
		p.Act.Remove(types.ACT_CHALLENGER)
		p.Act.Remove(types.ACT_CHALLENGED)
		p.Act.Remove(types.PLR_SILENCE)
	}
	if ArenaIsBusyFunc != nil {
		ArenaIsBusyFunc(false)
	}
	return true
}

// stripArenaDebuffs removes the 4 affect types C lists at fight.c:2876-
// 2879: poison, blindness, sleep, curse. Skips gsns resolved to -1 —
// the skill may not be present in the stock table.
func stripArenaDebuffs(ch *types.CharData) {
	for _, gsn := range []int{gsnPoison, gsnBlindness, gsnSleep, gsnCurse} {
		if gsn < 0 {
			continue
		}
		handler.AffectStrip(ch, gsn)
	}
}
