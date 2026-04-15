package mudprog

import (
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/command"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// CmdRegistry is set from main to allow mudprogs to execute commands.
// Written once at boot before the game loop starts; read only from the game loop goroutine. Safe without synchronization.
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

		case "mpsleep":
			// Only defer if we're in an executing state. (C handles mpsleep
			// before if-state evaluation, causing it to always fire; we instead
			// gate it on if-state for consistency with every other command. This
			// is a deliberate Go-side divergence — see sleep_test.go.)
			if ifLevel != 0 && ifState[ifLevel] != 0 {
				continue
			}
			// Parse timer arg. C: missing arg -> 4; arg < 1 -> bug + 4.
			ticks := 4
			if rest != "" {
				if n, err := strconv.Atoi(strings.TrimSpace(strings.Fields(rest)[0])); err == nil {
					ticks = n
				}
			}
			if ticks < 1 {
				util.Bug("MudProg mpsleep: bad arg, using default (mob %s)", mob.Name)
				ticks = 4
			}
			// Build remaining ComList: all lines AFTER the mpsleep line.
			var remaining string
			if i+1 < len(lines) {
				remaining = strings.Join(lines[i+1:], "\n")
			}
			sd := &types.MProgSleepData{
				Timer:      ticks,
				Type:       types.MP_MOB,
				IfLevel:    ifLevel,
				ComList:    remaining,
				Mob:        mob,
				Actor:      actor,
				Obj:        obj,
				Victim:     victim,
				Target:     target,
				SingleStep: singleStep,
			}
			// Encode int ifState into [MAX_IFS][2]bool.
			// Mapping: state==1 (skip-if)   -> [i][0]=true
			//          state==2 (skip-else) -> [i][1]=true
			//          state==0 (execute)   -> both false
			for k := 0; k < types.MAX_IFS; k++ {
				switch ifState[k] {
				case 1:
					sd.IfState[k][0] = true
				case 2:
					sd.IfState[k][1] = true
				}
			}
			SleepAdd(sd)
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
	case "mpasound":
		mpAsound(mob, line[len("mpasound"):])
		return
	case "mpsound":
		mpEcho(mob, line[len("mpsound"):])
		return
	case "mpsoundat":
		mpEchoAt(mob, line[len("mpsoundat"):])
		return
	case "mpsoundaround":
		mpEchoAround(mob, line[len("mpsoundaround"):])
		return
	case "mpmusic":
		mpEcho(mob, line[len("mpmusic"):])
		return
	case "mpmusicat":
		mpEchoAt(mob, line[len("mpmusicat"):])
		return
	case "mpmusicaround":
		mpEchoAround(mob, line[len("mpmusicaround"):])
		return
	case "mpechozone":
		mpEchoZone(mob, line[len("mpechozone"):])
		return
	case "mpmload":
		mpMload(mob, line[len("mpmload"):])
		return
	case "mpoload":
		mpOload(mob, line[len("mpoload"):])
		return
	case "mpinvis":
		mpInvis(mob, line[len("mpinvis"):])
		return
	case "mpat":
		mpAt(mob, line[len("mpat"):])
		return
	case "mpadvance":
		mpAdvance(mob, line[len("mpadvance"):])
		return
	case "mpslay":
		mpSlay(mob, line[len("mpslay"):])
		return
	case "mplog":
		mpLog(mob, line[len("mplog"):])
		return
	case "mprestore":
		mpRestore(mob, line[len("mprestore"):])
		return
	case "mpfavor":
		mpFavor(mob, line[len("mpfavor"):])
		return
	case "mpnuisance":
		mpNuisance(mob, line[len("mpnuisance"):])
		return
	case "mpunnuisance":
		mpUnnuisance(mob, line[len("mpunnuisance"):])
		return
	case "mpbodybag":
		mpBodybag(mob, line[len("mpbodybag"):])
		return
	case "mpmorph":
		mpMorph(mob, line[len("mpmorph"):])
		return
	case "mpunmorph":
		mpUnmorph(mob, line[len("mpunmorph"):])
		return
	case "mppractice":
		mpPractice(mob, line[len("mppractice"):])
		return
	case "mpopenpassage":
		mpOpenPassage(mob, line[len("mpopenpassage"):])
		return
	case "mpclosepassage":
		mpClosePassage(mob, line[len("mpclosepassage"):])
		return
	case "mpfillin":
		mpFillIn(mob, line[len("mpfillin"):])
		return
	case "mppeace":
		mpPeace(mob, line[len("mppeace"):])
		return
	case "mppkset":
		mpPkset(mob, line[len("mppkset"):])
		return
	case "mpoowner":
		mpOowner(mob, line[len("mpoowner"):])
		return
	case "mphunt":
		mpHunt(mob, line[len("mphunt"):])
		return
	case "mphate":
		mpHate(mob, line[len("mphate"):])
		return
	case "mpdeposit":
		mpDeposit(mob, line[len("mpdeposit"):])
		return
	case "mpwithdraw":
		mpWithdraw(mob, line[len("mpwithdraw"):])
		return
	case "mpapply":
		mpApply(mob, line[len("mpapply"):])
		return
	case "mpapplyb":
		mpApplyB(mob, line[len("mpapplyb"):])
		return
	case "mpapplyaffect":
		mpApplyAffect(mob, line[len("mpapplyaffect"):])
		return
	case "mpdelay":
		mpDelay(mob, line[len("mpdelay"):])
		return
	case "mpstrew":
		mpStrew(mob, line[len("mpstrew"):])
		return
	case "mpscatter":
		mpScatter(mob, line[len("mpscatter"):])
		return
	case "mpdream":
		mpDream(mob, line[len("mpdream"):])
		return
	case "mpnothing":
		mpNothing(mob, line[len("mpnothing"):])
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
