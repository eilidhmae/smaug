// Arena PvP — player-driven challenge / accept / decline / withdraw, plus
// the shared state seam (`arenaState`) exported to the game loop's tick
// handler and the combat-victory check.
//
// See plan-phase6-arena.md. C ref: src/arena.c:89-358 (under
// #ifdef ENABLE_ARENA).
//
// Package-level `arenaState` mirrors the C file-scope globals
// `is_challenge` and `arena_is_busy`. Every arena operation reads/writes
// this shared state; tests reset via ResetArenaState.
package act

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// arenaState mirrors the file-scope globals in src/arena.c:91-92.
// IsChallenge: a pending challenge exists, awaiting accept/decline.
// IsBusy:      a match is in progress (set by DoAccept, cleared by the
//
//	combat-victory branch).
//
// Both flags are reset by tests via ResetArenaState. No lock: single-
// threaded pulse-based game loop; same invariant as C.
var arenaState struct {
	IsChallenge bool
	IsBusy      bool
}

// ResetArenaState clears both arena flags. Test-only seam.
func ResetArenaState() {
	arenaState.IsChallenge = false
	arenaState.IsBusy = false
}

// SetArenaIsBusy is the exported seam used by combat.ArenaVictoryCheck to
// clear arena_is_busy from a package that cannot import `act`. Installed
// at boot via `combat.ArenaIsBusyFunc = act.SetArenaIsBusy`. See
// plan-phase6-arena.md §G6 Seams and §G7.
func SetArenaIsBusy(v bool) {
	arenaState.IsBusy = v
}

// SetArenaIsChallenge is the exported seam used by the per-char timeout
// tick (internal/game/update.go charUpdate). Keeps all package-level
// arena state writes inside this file. See plan-phase6-arena.md §G5.
func SetArenaIsChallenge(v bool) {
	arenaState.IsChallenge = v
}

// ArenaIsChallenge / ArenaIsBusy are read-only exports for tests and
// the charUpdate tick. They return the current flag values.
func ArenaIsChallenge() bool { return arenaState.IsChallenge }
func ArenaIsBusy() bool      { return arenaState.IsBusy }

// DoChallenge handles "challenge <player>". Pre-accept PvP invite —
// sets ACT_CHALLENGER on the caller, notifies the target, arms a
// TIMER_CHALLENGE timer. C ref: src/arena.c:97-193.
//
// Gate order preserved verbatim including the C typos ("wait a few
// moments the try again") and broken punctuation ("gone from the
// realms at this time.'"). See plan-phase6-arena.md §G2.
func DoChallenge(ch *types.CharData, argument string) {
	if ch == nil {
		return
	}
	if strings.TrimSpace(argument) == "" {
		ch.Send("You must specify who you want to challenge.\n\r")
		return
	}
	if arenaState.IsChallenge {
		ch.Send("Someone has been challenged wait a few moments the try again.")
		return
	}
	if arenaState.IsBusy {
		ch.Send("The arena is being used at the moment..Please wait a few minutes.\n\r")
		return
	}
	victim := handler.GetCharWorld(WorldRef, ch, argument)
	if victim == nil {
		util.Act(types.AT_ACTION, "Sorry but $t seems to be gone from the realms at this time.'",
			ch, nil, argument, nil, types.TO_CHAR)
		return
	}
	if ch.Act.IsSet(types.ACT_CHALLENGED) {
		ch.Send("You have been challenged already.  Accept or decline.\n\r")
		return
	}
	if victim.Act.IsSet(types.ACT_CHALLENGER) {
		util.Act(types.AT_RED, "$N has already challenged someone else.",
			ch, victim, nil, nil, types.TO_CHAR)
		return
	}
	if victim.Act.IsSet(types.PLR_AFK) {
		util.Act(types.AT_YELLOW, "Sorry but $N is AFK at the moment.",
			ch, victim, nil, nil, types.TO_CHAR)
		return
	}
	if victim.IsNPC() || ch.IsNPC() {
		ch.Send("You can only challenge players or yourself.\n\r")
		return
	}
	if victim.IsImmortal() && !ch.IsImmortal() {
		util.Act(types.AT_RED, "Sorry but $E is a higher being.",
			ch, victim, nil, nil, types.TO_CHAR)
		ch.Send("Besides they all would laugh at you!\n\r")
		return
	}
	if victim.Level <= 5 {
		util.Act(types.AT_WHITE, "Sorry but $E is not experienced enough.",
			ch, victim, nil, nil, types.TO_CHAR)
		return
	}
	if victim.Fighting != nil {
		util.Act(types.AT_ACTION, "Sorry but $N is in combat right now.",
			ch, victim, nil, nil, types.TO_CHAR)
		return
	}
	// C: !str_cmp(argument,"self") || !str_cmp(argument,"mirror match")
	// Both checks reference the RAW argument before one_argument; preserve.
	if strings.EqualFold(argument, "self") || strings.EqualFold(argument, "mirror match") {
		return // "I dont wanna give away all my stuff right now"
	}

	ch.Act.Set(types.ACT_CHALLENGER)
	ch.Send("Your challenge has been sent if they done accept in 3 ticks it will decline")
	ch.Sendf("\n\rIf they do not accept it will be automatically withdrawn from %s.\n\r", ch.Name)
	util.Act(types.AT_RED, "You have been challenged by $n.",
		ch, victim, nil, nil, types.TO_VICT)
	victim.Sendf("Type accept %s to accept or decline %s to decline.\n\r", ch.Name, ch.Name)
	handler.AddTimer(ch, types.TIMER_CHALLENGE, 5, "", 0)
	arenaState.IsChallenge = true
}

// DoAccept handles "accept <challenger>". Teleports both PCs to random
// arena rooms, sets ACT_CHALLENGED on the accepter, PLR_SILENCE on both,
// flips arena_is_busy. C ref: src/arena.c:198-265.
//
// Note: the C victim-first announce + ch-first teleport ordering is
// preserved, including the apparent bug at C:255 where do_look is
// called on `ch` twice (once for each plop). plan §G3 opts to preserve
// C verbatim here — a second look on the accepter after both teleports
// means the accepter sees the room they just moved to (good UX) and
// the victim doesn't get an auto-look (matches C).
func DoAccept(ch *types.CharData, argument string) {
	if ch == nil {
		return
	}
	if ch.IsNPC() {
		ch.Send("Hrmm....You must be a new kinda mobile to accept challenges eh?")
		return
	}
	if strings.TrimSpace(argument) == "" {
		ch.Send("You must be specify who challenged you I am only a MUD that can do so mud.\n\r")
		return
	}
	victim := handler.GetCharWorld(WorldRef, ch, argument)
	if victim == nil {
		util.Act(types.AT_ACTION, "Sorry but $t seems to be gone from the realms at this time.'",
			ch, nil, argument, nil, types.TO_CHAR)
		return
	}
	if victim.IsNPC() {
		ch.Send("Hrmm, I dont think I can let you die like that.\n\r")
		return
	}
	if victim == ch {
		ch.Send("I bet you think your funny eh?\n\r")
		return
	}
	if !victim.Act.IsSet(types.ACT_CHALLENGER) {
		ch.Send("Now have you drank that much ale that you can't remember who challenged you?\n\r")
		return
	}
	room1 := pickArenaRoom()
	room2 := pickArenaRoom()
	if room1 == nil || room2 == nil {
		// No arena-flagged room in the world. Clear state and fail
		// cleanly — C would crash; we surface a message instead.
		ch.Send("The arena is not configured on this server.\n\r")
		return
	}

	victim.Sendf("%s has accepted your challenge!\n\r", ch.Name)
	ch.Send("You plop onto the ground.\n\r")
	handler.CharFromRoom(ch)
	handler.CharToRoom(ch, room1)
	DoLook(ch, "auto")
	victim.Send("You plop onto the ground.\n\r")
	handler.CharFromRoom(victim)
	handler.CharToRoom(victim, room2)
	// C arena.c:255 calls do_look(ch, "auto") here — NOT victim. Port
	// verbatim: accepter gets a second look, victim gets none.
	DoLook(ch, "auto")

	ch.Act.Set(types.ACT_CHALLENGED)
	ch.Act.Set(types.PLR_SILENCE)
	victim.Act.Set(types.PLR_SILENCE)
	arenaState.IsBusy = true
	arenaState.IsChallenge = false
	handler.RemoveTimer(ch, types.TIMER_CHALLENGE)
	handler.RemoveTimer(victim, types.TIMER_CHALLENGE)
}

// DoDecline handles "decline <challenger>". C ref: src/arena.c:267-317.
//
// C bug preserved verbatim: the flag-clear at C:311 removes
// ACT_CHALLENGER from `ch` (the decliner) not from `victim` (the
// challenger). The challenger's flag stays set until the 5-tick
// timeout fires in charUpdate. See Open Q1 in the plan — preserve C.
func DoDecline(ch *types.CharData, argument string) {
	if ch == nil {
		return
	}
	if strings.TrimSpace(argument) == "" {
		// C reuses the challenge-syntax error here; verbatim.
		ch.Send("You must specify who you want to challenge.\n\r")
		return
	}
	if !arenaState.IsChallenge {
		ch.Send("Nobody hase even challenged someone wanna try again?")
		return
	}
	if arenaState.IsBusy {
		ch.Send("Alright the arena is in use so no one yet again has challenged you.\n\r")
		return
	}
	victim := handler.GetCharWorld(WorldRef, ch, argument)
	if victim == nil {
		util.Act(types.AT_ACTION, "Sorry but $t seems to be gone from the realms at this time.'",
			ch, nil, argument, nil, types.TO_CHAR)
		return
	}
	if !victim.Act.IsSet(types.ACT_CHALLENGER) {
		util.Act(types.AT_RED, "$N hasn't challenged you.",
			ch, victim, nil, nil, types.TO_CHAR)
		return
	}
	if victim.IsNPC() || victim == ch {
		ch.Send("How did that happen?\n\r")
		return
	}
	arenaState.IsChallenge = false
	// C bug: clears flag on wrong target. See method doc.
	ch.Act.Remove(types.ACT_CHALLENGER)
	victim.Sendf("%s decline your invitation to death!\n\r", ch.Name)
	ch.Sendf("You have choosen not to die by %s!\n\r", victim.Name)
}

// DoWithdraw handles "withdraw <challenger-target>". Challenger-initiated
// cancellation. C ref: src/arena.c:319-358.
func DoWithdraw(ch *types.CharData, argument string) {
	if ch == nil || ch.IsNPC() {
		return
	}
	if !ch.Act.IsSet(types.ACT_CHALLENGER) {
		ch.Send("You haven't challenged anyone!\n\r")
		return
	}
	victim := handler.GetCharWorld(WorldRef, ch, argument)
	if victim == nil {
		ch.Send("Does it look like they are here?!?\n\rNO, I didn't think so!!\n\r")
		return
	}
	if arenaState.IsBusy {
		ch.Send("Hrmm, how does that work again?!?\n\r")
		return
	}
	if victim.IsNPC() {
		ch.Send("I think you are brain dead how many drugs have you done again?\n\r")
		return
	}
	arenaState.IsChallenge = false
	ch.Act.Remove(types.ACT_CHALLENGER)
	handler.RemoveTimer(ch, types.TIMER_CHALLENGE)
	victim.Sendf("%s has withdrawn their challenge!\n\r", ch.Name)
	ch.Sendf("You withdraw your challenge to %s!\n\r", victim.Name)
}

// pickArenaRoom returns a random ROOM_ARENA-flagged room by scanning the
// vnum range ROOM_VNUM_ARENA_MIN..MAX and filtering for the flag. Plan
// Open Question #3 resolution: filter-then-random (option b) rather than
// C-verbatim (option c, may return nil for non-arena vnums in the
// range). Returns nil when no arena-flagged room exists.
func pickArenaRoom() *types.RoomIndexData {
	if WorldRef == nil {
		return nil
	}
	var arenas []*types.RoomIndexData
	for v := types.ROOM_VNUM_ARENA_MIN; v <= types.ROOM_VNUM_ARENA_MAX; v++ {
		r := WorldRef.GetRoom(v)
		if r == nil {
			continue
		}
		if r.RoomFlags.IsSet(types.ROOM_ARENA) {
			arenas = append(arenas, r)
		}
	}
	if len(arenas) == 0 {
		return nil
	}
	return arenas[util.NumberRange(0, len(arenas)-1)]
}
