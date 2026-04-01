package mudprog

import (
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// DoIfCheck evaluates a single if-check expression.
// Format: "ifchk( $var ) [operator value]"
// Returns true if the condition is met.
func DoIfCheck(check string, mob *types.CharData, actor *types.CharData,
	obj *types.ObjData, victim *types.CharData, target *types.ObjData) bool {

	check = strings.TrimSpace(check)
	if check == "" {
		return false
	}

	// Parse: "checkname($x) [op val]" or just "checkname($x)"
	parenIdx := strings.Index(check, "(")
	if parenIdx < 0 {
		return false
	}
	checkName := strings.TrimSpace(check[:parenIdx])
	rest := check[parenIdx+1:]

	closeIdx := strings.Index(rest, ")")
	if closeIdx < 0 {
		return false
	}
	argStr := strings.TrimSpace(rest[:closeIdx])
	rest = strings.TrimSpace(rest[closeIdx+1:])

	// Resolve the variable argument to a character
	chk := resolveChar(argStr, mob, actor, victim)

	// Parse operator and value
	op := "=="
	valStr := ""
	if rest != "" {
		parts := strings.SplitN(rest, " ", 2)
		if len(parts) >= 1 {
			op = parts[0]
		}
		if len(parts) >= 2 {
			valStr = strings.TrimSpace(parts[1])
		}
	}
	val, _ := strconv.Atoi(valStr)

	switch strings.ToLower(checkName) {
	case "rand":
		return util.NumberPercent() <= val

	case "ispc":
		return chk != nil && !chk.IsNPC()
	case "isnpc":
		return chk != nil && chk.IsNPC()
	case "isevil":
		return chk != nil && chk.Alignment < -350
	case "isgood":
		return chk != nil && chk.Alignment > 350
	case "isneutral":
		return chk != nil && chk.Alignment >= -350 && chk.Alignment <= 350
	case "isimmort":
		return chk != nil && chk.IsImmortal()
	case "isfight":
		return chk != nil && chk.Fighting != nil
	case "ischarmed":
		return chk != nil && chk.AffectedBy.IsSet(types.AFF_CHARM)
	case "isflying":
		return chk != nil && chk.AffectedBy.IsSet(types.AFF_FLYING)
	case "isinvis":
		return chk != nil && chk.AffectedBy.IsSet(types.AFF_INVISIBLE)
	case "isaffected":
		return chk != nil && chk.AffectedBy.IsSet(val)

	case "level":
		if chk == nil {
			return false
		}
		return compareInt(chk.Level, op, val)
	case "hp":
		if chk == nil {
			return false
		}
		return compareInt(chk.Hit, op, val)
	case "hppcnt":
		if chk == nil || chk.MaxHit == 0 {
			return false
		}
		return compareInt(chk.Hit*100/chk.MaxHit, op, val)
	case "mana":
		if chk == nil {
			return false
		}
		return compareInt(chk.Mana, op, val)
	case "gold":
		if chk == nil {
			return false
		}
		return compareInt(chk.Gold, op, val)
	case "sex":
		if chk == nil {
			return false
		}
		return compareInt(chk.Sex, op, val)
	case "position":
		if chk == nil {
			return false
		}
		return compareInt(chk.Position, op, val)
	case "class":
		if chk == nil {
			return false
		}
		return compareInt(chk.Class, op, val)
	case "race":
		if chk == nil {
			return false
		}
		return compareInt(chk.Race, op, val)
	case "alignment":
		if chk == nil {
			return false
		}
		return compareInt(chk.Alignment, op, val)
	case "str":
		if chk == nil {
			return false
		}
		return compareInt(chk.GetCurrStr(), op, val)
	case "int":
		if chk == nil {
			return false
		}
		return compareInt(chk.GetCurrInt(), op, val)
	case "wis":
		if chk == nil {
			return false
		}
		return compareInt(chk.GetCurrWis(), op, val)
	case "dex":
		if chk == nil {
			return false
		}
		return compareInt(chk.GetCurrDex(), op, val)
	case "con":
		if chk == nil {
			return false
		}
		return compareInt(chk.GetCurrCon(), op, val)

	case "name":
		if chk == nil {
			return false
		}
		return matchStr(chk.Name, op, valStr)

	case "isinroom":
		if mob == nil || mob.InRoom == nil {
			return false
		}
		return mob.InRoom.Vnum == val

	default:
		return false
	}
}

func resolveChar(varStr string, mob, actor, victim *types.CharData) *types.CharData {
	varStr = strings.TrimSpace(varStr)
	switch varStr {
	case "$i":
		return mob
	case "$n":
		return actor
	case "$t":
		return victim
	case "$r":
		if mob != nil && mob.InRoom != nil && len(mob.InRoom.People) > 0 {
			return mob.InRoom.People[0]
		}
		return nil
	default:
		return actor
	}
}

func compareInt(lhs int, op string, rhs int) bool {
	switch op {
	case "==", "=":
		return lhs == rhs
	case "!=":
		return lhs != rhs
	case ">":
		return lhs > rhs
	case "<":
		return lhs < rhs
	case ">=":
		return lhs >= rhs
	case "<=":
		return lhs <= rhs
	default:
		return lhs == rhs
	}
}

func matchStr(lhs, op, rhs string) bool {
	switch op {
	case "==", "=":
		return strings.EqualFold(lhs, rhs)
	case "!=":
		return !strings.EqualFold(lhs, rhs)
	case "/":
		return strings.Contains(strings.ToLower(lhs), strings.ToLower(rhs))
	case "!/":
		return !strings.Contains(strings.ToLower(lhs), strings.ToLower(rhs))
	default:
		return strings.EqualFold(lhs, rhs)
	}
}
