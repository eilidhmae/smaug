package mudprog

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/command"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// CmdRegistry is set from main to allow mudprogs to execute commands.
var CmdRegistry *command.Registry

// maxProgNest prevents infinite recursion in mudprogs.
const maxProgNest = 3

var progNest int

// Driver executes a mudprog script for the given mob.
func Driver(comList string, mob *types.CharData, actor *types.CharData,
	obj *types.ObjData, victim *types.CharData, target *types.ObjData, singleStep bool) {

	if mob == nil || comList == "" {
		return
	}

	progNest++
	if progNest > maxProgNest {
		progNest--
		util.Bug("MudProg driver: nested too deep (mob %s)", mob.Name)
		return
	}

	lines := strings.Split(comList, "\n")

	ifLevel := 0
	var ifState [types.MAX_IFS]int // 0=execute, 1=skip-if, 2=skip-else

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || line == "~" {
			continue
		}

		// Parse the command keyword
		keyword, rest := firstWord(line)
		keyword = strings.ToLower(keyword)

		switch keyword {
		case "if":
			ifLevel++
			if ifLevel >= types.MAX_IFS {
				util.Bug("MudProg: too many nested ifs (mob %s)", mob.Name)
				progNest--
				return
			}
			if ifLevel > 1 && ifState[ifLevel-1] != 0 {
				ifState[ifLevel] = 1 // parent is skipping, skip this too
			} else if DoIfCheck(rest, mob, actor, obj, victim, target) {
				ifState[ifLevel] = 0 // condition true, execute
			} else {
				ifState[ifLevel] = 1 // condition false, skip
			}

		case "or":
			if ifLevel > 0 && ifState[ifLevel] == 1 {
				// Previous if failed, try this OR condition
				if DoIfCheck(rest, mob, actor, obj, victim, target) {
					ifState[ifLevel] = 0
				}
			}

		case "else":
			if ifLevel > 0 {
				if ifState[ifLevel] == 0 {
					ifState[ifLevel] = 2 // was executing, skip else
				} else if ifState[ifLevel] == 1 {
					ifState[ifLevel] = 0 // was skipping, execute else
				}
			}

		case "endif":
			if ifLevel > 0 {
				ifState[ifLevel] = 0
				ifLevel--
			}

		case "break":
			progNest--
			return

		default:
			// Execute command if we're in an executing state
			if ifLevel == 0 || ifState[ifLevel] == 0 {
				expanded := Translate(mob, actor, obj, victim, target, line)
				executeCommand(mob, expanded)
			}
		}

		if singleStep {
			break
		}
	}

	progNest--
}

// executeCommand runs a single command through the interpreter for the mob.
func executeCommand(mob *types.CharData, line string) {
	if mob == nil || line == "" {
		return
	}

	keyword, _ := firstWord(line)
	keyword = strings.ToLower(keyword)

	// Handle mp-specific commands
	switch keyword {
	case "mpecho":
		mpEcho(mob, line[len("mpecho"):])
		return
	case "mpechoat":
		mpEchoAt(mob, line[len("mpechoat"):])
		return
	case "mpechoaround":
		mpEchoAround(mob, line[len("mpechoaround"):])
		return
	case "mpgoto":
		mpGoto(mob, line[len("mpgoto"):])
		return
	case "mptransfer":
		mpTransfer(mob, line[len("mptransfer"):])
		return
	case "mpforce":
		mpForce(mob, line[len("mpforce"):])
		return
	case "mpkill":
		mpKill(mob, line[len("mpkill"):])
		return
	case "mpdamage":
		mpDamage(mob, line[len("mpdamage"):])
		return
	case "mppurge":
		mpPurge(mob, line[len("mppurge"):])
		return
	}

	// Fall through to normal command interpreter
	if CmdRegistry != nil {
		CmdRegistry.Interpret(mob, line)
	}
}

func firstWord(s string) (string, string) {
	s = strings.TrimSpace(s)
	idx := strings.IndexAny(s, " \t")
	if idx < 0 {
		return s, ""
	}
	return s[:idx], strings.TrimSpace(s[idx+1:])
}
