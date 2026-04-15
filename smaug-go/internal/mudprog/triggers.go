package mudprog

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// MobTrigger checks and fires a mudprog trigger on a mob.
// trigType is a MPROG_* constant. arg is the trigger argument (e.g., speech text).
func MobTrigger(trigType int64, mob *types.CharData, actor *types.CharData,
	obj *types.ObjData, victim *types.CharData, target *types.ObjData, arg string) {

	if mob == nil || mob.IndexData == nil {
		return
	}
	// Short-circuit when the index has no progs at all.
	if len(mob.IndexData.MudProgs) == 0 {
		return
	}

	for _, prog := range mob.IndexData.MudProgs {
		if prog.Type&trigType == 0 {
			continue
		}

		// Check if the trigger argument matches
		if !triggerMatches(trigType, prog.ArgList, arg) {
			continue
		}

		Driver(prog.ComList, mob, actor, obj, victim, target, false)
		return // Only fire one matching prog per trigger
	}
}

// triggerMatches checks if the trigger argument matches the prog's arglist.
func triggerMatches(trigType int64, argList string, actual string) bool {
	switch trigType {
	case types.MPROG_RAND:
		// ArgList is a percent chance
		chance := util.URANGE(0, atoi(argList), 100)
		return util.NumberPercent() <= chance

	case types.MPROG_SPEECH, types.MPROG_SPEECHIW, types.MPROG_TELL:
		// ArgList is keyword(s) to match in speech.
		// TODO(tier4): MPROG_SPEECH "p " prefix means the rest of the arglist
		// is a phrase to match verbatim (C mud_prog.c:speech dispatch). The
		// Go port currently just discards the "p" token and keyword-matches
		// the rest, which will over-fire on partial matches vs the C phrase
		// behaviour. Fix when porting the speech dispatch trio.
		if argList == "" {
			return true
		}
		keywords := strings.Fields(argList)
		for _, kw := range keywords {
			if strings.EqualFold(kw, "p") {
				continue // "p" prefix for phrase match
			}
			if strings.Contains(strings.ToLower(actual), strings.ToLower(kw)) {
				return true
			}
		}
		return false

	case types.MPROG_ACT:
		// ArgList contains keywords to match in the act string
		if argList == "" {
			return true
		}
		return strings.Contains(strings.ToLower(actual), strings.ToLower(argList))

	case types.MPROG_HITPRCNT:
		// ArgList is a percent threshold
		return true // Caller should check HP percent

	case types.MPROG_BRIBE:
		// ArgList is a gold amount
		return atoi(actual) >= atoi(argList)

	case types.MPROG_GIVE:
		// ArgList is an object keyword
		if argList == "" {
			return true
		}
		return util.IsName(argList, actual)

	default:
		// Most triggers (GREET, ENTRY, FIGHT, DEATH, etc.) always fire
		return true
	}
}

func atoi(s string) int {
	s = strings.TrimSpace(s)
	v, _ := strings.CutPrefix(s, "+")
	n := 0
	for _, c := range v {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			break
		}
	}
	return n
}

// --- Convenience trigger functions ---

// TrigGreet fires greet/all_greet triggers when a character enters a room.
func TrigGreet(ch *types.CharData) {
	if ch == nil || ch.InRoom == nil {
		return
	}
	for _, mob := range ch.InRoom.People {
		if mob == ch || !mob.IsNPC() || mob.Fighting != nil {
			continue
		}
		if mob.Position < types.POS_STANDING {
			continue
		}
		MobTrigger(types.MPROG_GREET, mob, ch, nil, nil, nil, "")
		MobTrigger(types.MPROG_ALL_GREET, mob, ch, nil, nil, nil, "")
	}
}

// TrigEntry fires entry triggers when a mob enters a room.
func TrigEntry(mob *types.CharData) {
	if mob == nil || mob.InRoom == nil || !mob.IsNPC() {
		return
	}
	MobTrigger(types.MPROG_ENTRY, mob, mob, nil, nil, nil, "")
}

// TrigSpeech fires speech triggers from a spoken message.
func TrigSpeech(ch *types.CharData, message string) {
	if ch == nil || ch.InRoom == nil {
		return
	}
	for _, mob := range ch.InRoom.People {
		if mob == ch || !mob.IsNPC() {
			continue
		}
		MobTrigger(types.MPROG_SPEECH, mob, ch, nil, nil, nil, message)
	}
}

// TrigFight fires fight triggers during combat.
func TrigFight(mob *types.CharData) {
	if mob == nil || !mob.IsNPC() || mob.Fighting == nil {
		return
	}
	MobTrigger(types.MPROG_FIGHT, mob, mob.Fighting.Who, nil, nil, nil, "")
}

// TrigDeath fires death triggers when a mob dies.
func TrigDeath(mob *types.CharData, killer *types.CharData) {
	if mob == nil || !mob.IsNPC() {
		return
	}
	MobTrigger(types.MPROG_DEATH, mob, killer, nil, nil, nil, "")
}

// TrigRand fires random triggers (called periodically).
func TrigRand(mob *types.CharData) {
	if mob == nil || !mob.IsNPC() || mob.InRoom == nil {
		return
	}
	MobTrigger(types.MPROG_RAND, mob, nil, nil, nil, nil, "")
}

// TrigGive fires give triggers when an object is given to a mob.
func TrigGive(mob *types.CharData, ch *types.CharData, obj *types.ObjData) {
	if mob == nil || !mob.IsNPC() {
		return
	}
	MobTrigger(types.MPROG_GIVE, mob, ch, obj, nil, nil, obj.Name)
}

// TrigLogin fires login triggers when a character enters the world. It scans
// every NPC in ch's current room and fires any MPROG_LOGIN prog they carry.
// Mirrors C's mprog_login_trigger in src/mud_prog.c.
func TrigLogin(ch *types.CharData) {
	if ch == nil || ch.InRoom == nil {
		return
	}
	// Snapshot the slice so that MobTrigger mutations cannot reorder iteration.
	people := make([]*types.CharData, len(ch.InRoom.People))
	copy(people, ch.InRoom.People)
	for _, mob := range people {
		if mob == ch || !mob.IsNPC() {
			continue
		}
		// Must be awake (C IS_AWAKE: pos > POS_SLEEPING) and not fighting.
		if mob.Fighting != nil || mob.Position <= types.POS_SLEEPING {
			continue
		}
		// Avoid self-instance triggering (two of the same vnum in one room).
		if ch.IsNPC() && ch.IndexData != nil && mob.IndexData == ch.IndexData {
			continue
		}
		MobTrigger(types.MPROG_LOGIN, mob, ch, nil, nil, nil, "")
	}
}

// CheckVoid fires void triggers if the given room is now empty of PCs but
// still contains one or more NPCs. Called whenever a PC leaves a room (quit,
// death, disconnect, zone transfer). C equivalent: mprog_void_trigger.
func CheckVoid(room *types.RoomIndexData) {
	if room == nil {
		return
	}
	// If any PC is still present the room isn't void.
	for _, p := range room.People {
		if !p.IsNPC() {
			return
		}
	}
	// Fire VOID on any awake, non-fighting NPC still in the room.
	people := make([]*types.CharData, len(room.People))
	copy(people, room.People)
	for _, mob := range people {
		if mob == nil || !mob.IsNPC() {
			continue
		}
		// Must be awake (C IS_AWAKE: pos > POS_SLEEPING) and not fighting.
		if mob.Fighting != nil || mob.Position <= types.POS_SLEEPING {
			continue
		}
		TrigVoid(mob)
	}
}

// TrigVoid fires a void trigger on a single mob.
func TrigVoid(mob *types.CharData) {
	if mob == nil || !mob.IsNPC() {
		return
	}
	MobTrigger(types.MPROG_VOID, mob, nil, nil, nil, nil, "")
}

// TrigTell fires tell triggers on an NPC target when it receives a tell.
// Mirrors C's mprog_tell_trigger.
func TrigTell(ch *types.CharData, target *types.CharData, message string) {
	if ch == nil || target == nil || !target.IsNPC() {
		return
	}
	// Avoid self-triggering on the same mob index.
	if ch.IsNPC() && ch.IndexData != nil && target.IndexData == ch.IndexData {
		return
	}
	MobTrigger(types.MPROG_TELL, target, ch, nil, nil, nil, message)
}

// TrigHour fires MPROG_HOUR on all loaded NPC instances when the in-game hour
// changes. HOUR is the "every hour tick" trigger: it fires each hour the arg
// list matches (or is empty), re-triggering every hour.
func TrigHour(hour int, w *world.World) {
	if w == nil {
		return
	}
	for _, mob := range w.Characters {
		if mob == nil || !mob.IsNPC() || mob.IndexData == nil {
			continue
		}
		if len(mob.IndexData.MudProgs) == 0 {
			continue
		}
		fireTimeProg(mob, hour, types.MPROG_HOUR)
	}
}

// TrigTime fires MPROG_TIME triggers whose arg list matches the current hour.
// Unlike HOUR, each TIME prog only fires once per hour transition. Subsequent
// hour changes reset the "triggered" flag and allow re-firing.
func TrigTime(hour int, w *world.World) {
	if w == nil {
		return
	}
	for _, mob := range w.Characters {
		if mob == nil || !mob.IsNPC() || mob.IndexData == nil {
			continue
		}
		if len(mob.IndexData.MudProgs) == 0 {
			continue
		}
		fireTimeProg(mob, hour, types.MPROG_TIME)
	}
}

// fireTimeProg is the shared HOUR/TIME prog scanner. For each MudProg on the
// mob that matches the requested trigType, compare the arglist against hour:
//   - empty arglist on HOUR fires unconditionally every hour.
//   - numeric arglist fires if it equals hour.
//
// HOUR progs can re-fire each hour; TIME progs latch via Triggered until the
// arglist hour is left.
func fireTimeProg(mob *types.CharData, hour int, trigType int64) {
	for _, prg := range mob.IndexData.MudProgs {
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
		// HOUR re-fires every matching hour; TIME fires once per latch.
		if prg.Triggered && trigType != types.MPROG_HOUR {
			continue
		}
		prg.Triggered = true
		Driver(prg.ComList, mob, nil, nil, nil, nil, false)
	}
}

// parseTimeArg parses a HOUR/TIME prog arglist. Empty → (0, false).
func parseTimeArg(s string) (int, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	n := 0
	seen := false
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
		seen = true
	}
	if !seen {
		return 0, false
	}
	return n, true
}

// TrigSell fires sell triggers on a shopkeeper after a successful sell. C
// equivalent: mprog_sell_trigger. The prog's ArgList is an object vnum (0
// matches any vnum).
func TrigSell(mob *types.CharData, ch *types.CharData, obj *types.ObjData) {
	if mob == nil || ch == nil || obj == nil || !mob.IsNPC() {
		return
	}
	if mob.IndexData == nil || len(mob.IndexData.MudProgs) == 0 {
		return
	}
	if ch.IsNPC() && ch.IndexData != nil && mob.IndexData == ch.IndexData {
		return
	}
	vnum := 0
	if obj.IndexData != nil {
		vnum = obj.IndexData.Vnum
	}
	for _, prg := range mob.IndexData.MudProgs {
		if prg.Type&types.MPROG_SELL == 0 {
			continue
		}
		arg := strings.TrimSpace(prg.ArgList)
		// "0" or empty matches any; otherwise numeric match on vnum.
		if arg == "" || arg == "0" {
			Driver(prg.ComList, mob, ch, obj, nil, nil, false)
			return
		}
		n, ok := parseTimeArg(arg)
		if ok && n == vnum {
			Driver(prg.ComList, mob, ch, obj, nil, nil, false)
			return
		}
	}
}

// TrigHitprcnt fires MPROG_HITPRCNT progs on a wounded mob. For each matching
// prog, parse ArgList as a percent threshold (0-100); fire when the victim's
// current HP percent has dropped below the threshold. Mirrors C's
// mprog_hitprcnt_trigger behavior. Unlike MobTrigger's generic HITPRCNT
// matcher (which always returns true), this helper does its own per-prog
// percent comparison so each prog's threshold is honoured.
func TrigHitprcnt(mob *types.CharData, ch *types.CharData) {
	if mob == nil || !mob.IsNPC() || mob.IndexData == nil {
		return
	}
	if mob.MaxHit <= 0 {
		return
	}
	pct := 100 * mob.Hit / mob.MaxHit
	for _, prg := range mob.IndexData.MudProgs {
		if prg.Type&types.MPROG_HITPRCNT == 0 {
			continue
		}
		threshold := atoi(prg.ArgList)
		if pct < threshold {
			Driver(prg.ComList, mob, ch, nil, nil, nil, false)
			return
		}
	}
}

