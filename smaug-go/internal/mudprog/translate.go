// Package mudprog implements the SMAUG MUD Program scripting system.
package mudprog

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
)

// Translate performs variable substitution on a mudprog command line.
// $i = mob, $n = actor, $t = victim, $r = random char in room,
// $o = object, $p = target object
// Uppercase variants give short descriptions.
// $e/$E, $m/$M, $s/$S = pronouns for actor/victim.
func Translate(mob *types.CharData, actor *types.CharData, obj *types.ObjData,
	victim *types.CharData, target *types.ObjData, text string) string {

	if !strings.ContainsRune(text, '$') {
		return text
	}

	var sb strings.Builder
	sb.Grow(len(text) * 2)

	for i := 0; i < len(text); i++ {
		if text[i] != '$' || i+1 >= len(text) {
			sb.WriteByte(text[i])
			continue
		}
		i++
		c := text[i]
		switch c {
		case 'i':
			sb.WriteString(charName(mob))
		case 'I':
			sb.WriteString(charShort(mob))
		case 'n':
			sb.WriteString(charName(actor))
		case 'N':
			sb.WriteString(charShort(actor))
		case 't':
			sb.WriteString(charName(victim))
		case 'T':
			sb.WriteString(charShort(victim))
		case 'r':
			if mob != nil && mob.InRoom != nil && len(mob.InRoom.People) > 0 {
				sb.WriteString(charName(mob.InRoom.People[0]))
			} else {
				sb.WriteString("someone")
			}
		case 'o':
			sb.WriteString(objShort(obj))
		case 'O':
			sb.WriteString(objName(obj))
		case 'p':
			sb.WriteString(objShort(target))
		case 'P':
			sb.WriteString(objName(target))
		case 'e':
			sb.WriteString(heShe(actor))
		case 'E':
			sb.WriteString(heShe(victim))
		case 'm':
			sb.WriteString(himHer(actor))
		case 'M':
			sb.WriteString(himHer(victim))
		case 's':
			sb.WriteString(hisHer(actor))
		case 'S':
			sb.WriteString(hisHer(victim))
		case 'j':
			sb.WriteString(heShe(mob))
		case 'k':
			sb.WriteString(himHer(mob))
		case 'l':
			sb.WriteString(hisHer(mob))
		default:
			sb.WriteByte('$')
			sb.WriteByte(c)
		}
	}
	return sb.String()
}

func charName(ch *types.CharData) string {
	if ch == nil {
		return "someone"
	}
	return ch.Name
}

func charShort(ch *types.CharData) string {
	if ch == nil {
		return "someone"
	}
	if ch.ShortDescr != "" {
		return ch.ShortDescr
	}
	return ch.Name
}

func objShort(obj *types.ObjData) string {
	if obj == nil {
		return "something"
	}
	return obj.ShortDescr
}

func objName(obj *types.ObjData) string {
	if obj == nil {
		return "something"
	}
	return obj.Name
}

func heShe(ch *types.CharData) string {
	if ch == nil {
		return "it"
	}
	switch ch.Sex {
	case types.SEX_MALE:
		return "he"
	case types.SEX_FEMALE:
		return "she"
	default:
		return "it"
	}
}

func himHer(ch *types.CharData) string {
	if ch == nil {
		return "it"
	}
	switch ch.Sex {
	case types.SEX_MALE:
		return "him"
	case types.SEX_FEMALE:
		return "her"
	default:
		return "it"
	}
}

func hisHer(ch *types.CharData) string {
	if ch == nil {
		return "its"
	}
	switch ch.Sex {
	case types.SEX_MALE:
		return "his"
	case types.SEX_FEMALE:
		return "her"
	default:
		return "its"
	}
}
