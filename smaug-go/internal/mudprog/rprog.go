package mudprog

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// buildSupermobForRoom returns a transient "supermob" CharData used as the
// driver's self-reference when running room-progs. Mirrors C's
// `rset_supermob` (`src/mud_prog.c:3424`). The supermob is bound to the given
// room so that `mpecho` and other commands which dereference `mob.InRoom`
// work correctly.
//
// The supermob is NOT appended to room.People; like its C counterpart it is
// a singleton shuttled between rooms, not a resident mob.
func buildSupermobForRoom(room *types.RoomIndexData) *types.CharData {
	if room == nil {
		return nil
	}
	sm := &types.CharData{
		Name:       "supermob",
		ShortDescr: "the room",
		Level:      60,
		Position:   types.POS_STANDING,
		Hit:        1000,
		MaxHit:     1000,
		InRoom:     room,
	}
	if room.Name != "" {
		sm.ShortDescr = room.Name
	}
	sm.Act.Set(types.ACT_IS_NPC)
	return sm
}

// RoomTrigger is the room-prog equivalent of MobTrigger / ObjTrigger. It fires
// the first matching room-prog of trigType on room, using a synthesised
// supermob as the driver's self-reference.
//
// Mirrors C's `rprog_percent_check` (`src/mud_prog.c:3700`) + supermob flow.
func RoomTrigger(trigType int64, room *types.RoomIndexData, actor *types.CharData,
	obj *types.ObjData, victim *types.CharData, target *types.ObjData, arg string) {

	if room == nil || len(room.MudProgs) == 0 {
		return
	}
	sm := buildSupermobForRoom(room)
	if sm == nil {
		return
	}
	for _, prog := range room.MudProgs {
		if prog.Type&trigType == 0 {
			continue
		}
		if !triggerMatches(trigType, prog.ArgList, arg) {
			continue
		}
		Driver(prog.ComList, sm, actor, obj, victim, target, false)
		return
	}
}

// --- Public fire helpers. One per C rprog_*_trigger. ---

// RprogEnterTrigger fires ENTER progs when ch enters a room.
// C: rprog_enter_trigger (uses ENTER_PROG ≡ ENTRY_PROG).
func RprogEnterTrigger(ch *types.CharData) {
	if ch == nil {
		return
	}
	RoomTrigger(types.MPROG_ENTER, ch.InRoom, ch, nil, nil, nil, "")
}

// RprogLeaveTrigger fires LEAVE progs when ch leaves fromRoom. Callers pass
// the room being left explicitly so the helper can run before CharFromRoom.
// C: rprog_leave_trigger.
func RprogLeaveTrigger(ch *types.CharData, fromRoom *types.RoomIndexData) {
	if ch == nil {
		return
	}
	RoomTrigger(types.MPROG_LEAVE, fromRoom, ch, nil, nil, nil, "")
}

// RprogSleepTrigger fires SLEEP progs when ch transitions to POS_SLEEPING.
// C: rprog_sleep_trigger.
func RprogSleepTrigger(ch *types.CharData) {
	if ch == nil {
		return
	}
	RoomTrigger(types.MPROG_SLEEP, ch.InRoom, ch, nil, nil, nil, "")
}

// RprogRestTrigger fires REST progs when ch transitions to POS_RESTING.
// C: rprog_rest_trigger.
func RprogRestTrigger(ch *types.CharData) {
	if ch == nil {
		return
	}
	RoomTrigger(types.MPROG_REST, ch.InRoom, ch, nil, nil, nil, "")
}

// RprogRfightTrigger fires RFIGHT progs when combat begins in ch's room.
// C: rprog_rfight_trigger (RFIGHT_PROG ≡ FIGHT_PROG).
func RprogRfightTrigger(ch *types.CharData) {
	if ch == nil {
		return
	}
	RoomTrigger(types.MPROG_RFIGHT, ch.InRoom, ch, nil, nil, nil, "")
}

// RprogDeathTrigger fires RDEATH progs when someone dies in ch's room.
// C: rprog_death_trigger (RDEATH_PROG ≡ DEATH_PROG).
func RprogDeathTrigger(ch *types.CharData) {
	if ch == nil {
		return
	}
	RoomTrigger(types.MPROG_RDEATH, ch.InRoom, ch, nil, nil, nil, "")
}

// RprogImminfoTrigger fires IMMINFO progs when an immortal inspects ch's room
// (goto, rstat). C: rprog_imminfo_trigger.
func RprogImminfoTrigger(imm *types.CharData) {
	if imm == nil {
		return
	}
	RoomTrigger(types.MPROG_IMMINFO, imm.InRoom, imm, nil, nil, nil, "")
}

// RprogSpeechTrigger fires SPEECH progs on ch's room in response to spoken
// text. C: rprog_speech_trigger.
func RprogSpeechTrigger(ch *types.CharData, message string) {
	if ch == nil {
		return
	}
	RoomTrigger(types.MPROG_SPEECH, ch.InRoom, ch, nil, nil, nil, message)
}

// RprogCommandTrigger fires CMD progs on ch's room. Returns true if any prog
// consumed the command (caller should skip normal dispatch).
// C: rprog_command_trigger.
func RprogCommandTrigger(ch *types.CharData, line string) bool {
	if ch == nil || ch.InRoom == nil || line == "" {
		return false
	}
	room := ch.InRoom
	if len(room.MudProgs) == 0 {
		return false
	}
	// C matches CMD progs by comparing the first word of the command against
	// the prog's arglist (space-separated keyword list). This mirrors the
	// C wordlist_check logic in a lightweight form.
	cmd, _ := firstWord(line)
	cmd = strings.ToLower(cmd)
	for _, prog := range room.MudProgs {
		if prog.Type&types.MPROG_CMD == 0 {
			continue
		}
		argList := strings.TrimSpace(prog.ArgList)
		matched := false
		if argList == "" {
			matched = true
		} else {
			for _, kw := range strings.Fields(strings.ToLower(argList)) {
				if kw == cmd {
					matched = true
					break
				}
			}
		}
		if !matched {
			continue
		}
		sm := buildSupermobForRoom(room)
		if sm == nil {
			return false
		}
		Driver(prog.ComList, sm, ch, nil, nil, nil, false)
		return true
	}
	return false
}

// RprogRandomTrigger fires RAND progs on the given room. Called from the
// per-pulse room-update scanner. C: rprog_random_trigger.
func RprogRandomTrigger(room *types.RoomIndexData) {
	if room == nil || len(room.MudProgs) == 0 {
		return
	}
	RoomTrigger(types.MPROG_RAND, room, nil, nil, nil, nil, "")
}

// RprogActTrigger fires ACT progs on a room in response to an act-style
// message. C: rprog_act_trigger. The MVP fires synchronously (no MPROG_ACT
// queue on the room) — that matches the existing mob ACT path.
func RprogActTrigger(text string, room *types.RoomIndexData) {
	if room == nil {
		return
	}
	RoomTrigger(types.MPROG_ACT, room, nil, nil, nil, nil, text)
}

// RprogHourTrigger fires HOUR progs on every loaded room at each in-game
// hour transition. C: rprog_hour_trigger.
func RprogHourTrigger(hour int, w *world.World) {
	if w == nil {
		return
	}
	for _, room := range w.Rooms {
		if room == nil || len(room.MudProgs) == 0 {
			continue
		}
		fireRoomTimeProg(room, hour, types.MPROG_HOUR)
	}
}

// RprogTimeTrigger fires TIME progs on every loaded room whose arglist
// matches the current hour. C: rprog_time_trigger.
func RprogTimeTrigger(hour int, w *world.World) {
	if w == nil {
		return
	}
	for _, room := range w.Rooms {
		if room == nil || len(room.MudProgs) == 0 {
			continue
		}
		fireRoomTimeProg(room, hour, types.MPROG_TIME)
	}
}

// fireRoomTimeProg is the shared HOUR/TIME prog scanner for rooms. Mirrors
// fireTimeProg but operates on room.MudProgs and a supermob.
func fireRoomTimeProg(room *types.RoomIndexData, hour int, trigType int64) {
	var sm *types.CharData
	for _, prg := range room.MudProgs {
		if prg.Type&trigType == 0 {
			continue
		}
		triggerTime := false
		argNum, ok := parseTimeArg(prg.ArgList)
		if trigType == types.MPROG_HOUR && prg.ArgList == "" {
			triggerTime = true
		} else if ok && argNum == hour {
			triggerTime = true
		}
		if !triggerTime {
			if prg.Triggered {
				prg.Triggered = false
			}
			continue
		}
		if prg.Triggered && trigType != types.MPROG_HOUR {
			continue
		}
		prg.Triggered = true
		if sm == nil {
			sm = buildSupermobForRoom(room)
			if sm == nil {
				return
			}
		}
		Driver(prg.ComList, sm, nil, nil, nil, nil, false)
	}
}
