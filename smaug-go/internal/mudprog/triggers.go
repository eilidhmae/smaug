package mudprog

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// MobTrigger checks and fires a mudprog trigger on a mob.
// trigType is a MPROG_* constant. arg is the trigger argument (e.g., speech text).
func MobTrigger(trigType int, mob *types.CharData, actor *types.CharData,
	obj *types.ObjData, victim *types.CharData, target *types.ObjData, arg string) {

	if mob == nil || mob.IndexData == nil {
		return
	}
	if mob.IndexData.ProgTypes.IsSet(0) && len(mob.IndexData.MudProgs) == 0 {
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
func triggerMatches(trigType int, argList string, actual string) bool {
	switch trigType {
	case types.MPROG_RAND:
		// ArgList is a percent chance
		chance := util.URANGE(0, atoi(argList), 100)
		return util.NumberPercent() <= chance

	case types.MPROG_SPEECH, types.MPROG_SPEECHIW:
		// ArgList is keyword(s) to match in speech
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
