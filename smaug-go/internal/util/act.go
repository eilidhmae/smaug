package util

import (
	"strings"
	"unicode"

	"github.com/eilidhmae/smaug/internal/types"
)

// Act pronoun tables, indexed by Sex (SEX_NEUTRAL=0, SEX_MALE=1, SEX_FEMALE=2).
// Mirrors the static tables in C src/smaug.c act_string (line 3174).
var (
	actHeShe  = [3]string{"it", "he", "she"}
	actHimHer = [3]string{"it", "him", "her"}
	actHisHer = [3]string{"its", "his", "her"}
)

// actName returns the name/short-descr to use for a character token.
// NPCs with a ShortDescr use that; otherwise Name (matching C NAME/PERS macros).
func actName(ch *types.CharData) string {
	if ch == nil {
		return ""
	}
	if ch.IsNPC() && ch.ShortDescr != "" {
		return ch.ShortDescr
	}
	return ch.Name
}

// actCanSee returns true if observer can see target. There is no dedicated
// CanSee helper yet in the Go port; this is a reduced visibility check that
// handles AFF_INVISIBLE / AFF_HIDE / PLR_WIZINVIS, with AFF_DETECT_*,
// PLR_HOLYLIGHT, and Trust overrides. Self always sees self. Nil observer
// (broadcast/log contexts) always "sees".
// TODO: replace with a shared CanSee when it is ported.
func actCanSee(observer, target *types.CharData) bool {
	if observer == nil || target == nil {
		return true
	}
	if observer == target {
		return true
	}
	// PLR_HOLYLIGHT observers bypass all visibility gates.
	if observer.Act.IsSet(types.PLR_HOLYLIGHT) {
		return true
	}
	// Wiz-invisible PCs are invisible to anyone whose Trust is below their
	// WizInvis level. Critical: without this check wiz-invis immortals leak
	// their names to mortals in act() output.
	if target.PCData != nil && target.PCData.WizInvis > 0 {
		if observer.GetTrust() < target.PCData.WizInvis {
			return false
		}
	}
	if target.AffectedBy.IsSet(types.AFF_INVISIBLE) &&
		!observer.AffectedBy.IsSet(types.AFF_DETECT_INVIS) {
		return false
	}
	if target.AffectedBy.IsSet(types.AFF_HIDE) &&
		!observer.AffectedBy.IsSet(types.AFF_DETECT_HIDDEN) &&
		target.Fighting == nil {
		return false
	}
	return true
}

// actCanSeeObj returns true if observer can see obj. Mirrors can_see_obj for
// the limited set of flags relevant to act() substitution.
// TODO: replace with a shared CanSeeObj when it is ported.
func actCanSeeObj(observer *types.CharData, obj *types.ObjData) bool {
	if observer == nil || obj == nil {
		return true
	}
	if obj.ExtraFlags.IsSet(types.ITEM_INVIS) &&
		!observer.AffectedBy.IsSet(types.AFF_DETECT_INVIS) {
		return false
	}
	return true
}

// actPronoun returns the correct pronoun from tbl by ch.Sex, logging a bug and
// falling back to neutral (index 0) on out-of-range sex.
func actPronoun(tbl [3]string, ch *types.CharData) string {
	if ch == nil {
		return tbl[0]
	}
	s := ch.Sex
	if s < 0 || s > 2 {
		Bug("Act: bad sex %d on %s", s, ch.Name)
		return tbl[0]
	}
	return tbl[s]
}

// objShort returns the best short description for obj, preferring the instance
// ShortDescr and falling back to the index/prototype.
func objShort(obj *types.ObjData) string {
	if obj == nil {
		return ""
	}
	if obj.ShortDescr != "" {
		return obj.ShortDescr
	}
	if obj.IndexData != nil {
		return obj.IndexData.ShortDescr
	}
	return ""
}

// asObj returns arg as *types.ObjData if it is one, else nil.
func asObj(arg any) *types.ObjData {
	if o, ok := arg.(*types.ObjData); ok {
		return o
	}
	return nil
}

// asString returns arg as a string if it is one, else "".
func asString(arg any) string {
	if s, ok := arg.(string); ok {
		return s
	}
	return ""
}

// firstWord returns the first whitespace-delimited token of s (for $d).
func firstWord(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if i := strings.IndexAny(s, " \t"); i >= 0 {
		return s[:i]
	}
	return s
}

// ActFormat expands act()-style $-tokens in format against the given recipient.
// Returns the expanded string (capitalized, with a trailing "\n\r").
// - to is the recipient whose visibility governs substitution; may be nil.
// - ch is the actor referenced by $n/$e/$m/$s/$q.
// - vch is the victim referenced by $N/$E/$M/$S/$Q.
// - arg1 / arg2 may be *types.ObjData (for $p/$P) or string (for $t/$T/$d).
// See C src/smaug.c act_string (line 3171) for the reference implementation.
func ActFormat(format string, to, ch, vch *types.CharData, arg1, arg2 any) string {
	var b strings.Builder
	i := 0
	for i < len(format) {
		c := format[i]
		if c != '$' {
			b.WriteByte(c)
			i++
			continue
		}
		// '$' at end of string — drop it.
		if i+1 >= len(format) {
			break
		}
		tok := format[i+1]
		i += 2

		switch tok {
		case '$':
			b.WriteByte('$')
		case 'n':
			if ch == nil {
				b.WriteString(" <@@@> ")
				break
			}
			if !actCanSee(to, ch) {
				b.WriteString("someone")
			} else {
				b.WriteString(actName(ch))
			}
		case 'N':
			if vch == nil {
				Bug("Act: missing arg2 for code N: %s", format)
				b.WriteString(" <@@@> ")
				break
			}
			if !actCanSee(to, vch) {
				b.WriteString("someone")
			} else {
				b.WriteString(actName(vch))
			}
		case 'e':
			b.WriteString(actPronoun(actHeShe, ch))
		case 'E':
			if vch == nil {
				Bug("Act: missing arg2 for code E: %s", format)
				b.WriteString(" <@@@> ")
				break
			}
			b.WriteString(actPronoun(actHeShe, vch))
		case 'm':
			b.WriteString(actPronoun(actHimHer, ch))
		case 'M':
			if vch == nil {
				Bug("Act: missing arg2 for code M: %s", format)
				b.WriteString(" <@@@> ")
				break
			}
			b.WriteString(actPronoun(actHimHer, vch))
		case 's':
			b.WriteString(actPronoun(actHisHer, ch))
		case 'S':
			if vch == nil {
				Bug("Act: missing arg2 for code S: %s", format)
				b.WriteString(" <@@@> ")
				break
			}
			b.WriteString(actPronoun(actHisHer, vch))
		case 'q':
			if to == ch {
				// empty
			} else {
				b.WriteString("s")
			}
		case 'Q':
			if to == ch {
				b.WriteString("your")
			} else {
				b.WriteString(actPronoun(actHisHer, ch))
			}
		case 'p':
			if obj := asObj(arg1); obj != nil {
				if actCanSeeObj(to, obj) {
					b.WriteString(objShort(obj))
				} else {
					b.WriteString("something")
				}
			} else {
				Bug("Act: bad arg1 for $p in %s", format)
				b.WriteString(" <@@@> ")
			}
		case 'P':
			if obj := asObj(arg2); obj != nil {
				if actCanSeeObj(to, obj) {
					b.WriteString(objShort(obj))
				} else {
					b.WriteString("something")
				}
			} else {
				Bug("Act: bad arg2 for $P in %s", format)
				b.WriteString(" <@@@> ")
			}
		case 't':
			b.WriteString(asString(arg1))
		case 'T':
			b.WriteString(asString(arg2))
		case 'd':
			s := asString(arg2)
			if s == "" {
				b.WriteString("door")
			} else {
				b.WriteString(firstWord(s))
			}
		default:
			Bug("Act: bad code %c in format %s", tok, format)
			b.WriteString(" <@@@> ")
		}
	}

	b.WriteString("\n\r")
	out := b.String()
	if len(out) > 0 {
		r := []rune(out)
		r[0] = unicode.ToUpper(r[0])
		out = string(r)
	}
	return out
}

// atColorCode maps AT_* color-type constants to the project's `&X`
// color tokens (expanded to ANSI by net/color.go). Empty string means
// "no color prefix / suffix". Only entries used by current Go callers
// are populated; unknown AT values produce no color.
//
// Default values below are Go conventions that match existing hardcoded
// `&-codes` in the current Go callers (DoImmtalk uses &Y, poisoned-weapon
// prefix uses &G, etc.). They deliberately diverge from C's
// src/color.c:at_color_table default_set[] where those differ — a
// future plan can revisit the mapping if a Colorize[] port ships. See
// smaug-go/doc/plan-tranche-c.md § Design for rationale.
var atColorCode = map[int]string{
	types.AT_PLAIN:   "",   // no color — preserves current behavior
	types.AT_ACTION:  "&G", // green (Go convention; C default is AT_BLOOD)
	types.AT_HIT:     "&R", // bright red (Go convention; C default is AT_GREY)
	types.AT_HITME:   "&r", // dark red (Go convention; C default is AT_DGREY)
	types.AT_IMMORT:  "&Y", // yellow (matches DoImmtalk; C default is AT_RED)
	types.AT_GTELL:   "&P", // magenta (group tell)
	types.AT_GOSSIP:  "&P", // magenta (gossip channel)
	types.AT_SAY:     "&G", // green (say)
	types.AT_TELL:    "&G", // green (tell)
	types.AT_WHISPER: "&G", // green (whisper)
	types.AT_YELL:    "&Y", // yellow (yell)
	types.AT_SHOUT:   "&y", // dark yellow (shout)
	types.AT_MAGIC:   "&C", // cyan (magic effects)
	types.AT_POISON:  "&g", // dark green (poison)
	types.AT_SKILL:   "&C", // cyan (skill success)
	types.AT_SOCIAL:  "&P", // magenta (social commands)
	types.AT_REPORT:  "&C", // cyan (status reports)
	types.AT_QUEST:   "&Y", // yellow (quest hints)
}

// Act dispatches a formatted message to recipients per the `to` routing.
// aType is one of the AT_* color-type constants (see types/enums.go and
// C src/mud.h:1107-1166). It applies uniformly to every recipient of
// this call — callers that want different colors for TO_CHAR vs TO_VICT
// make two Act() calls with different aType values.
// to is one of types.TO_CHAR, TO_VICT, TO_NOTVICT, TO_ROOM (TO_CANSEE
// treated like TO_ROOM minus the actor).
// ch is the actor (required).
// vch may be nil when not applicable.
// arg1/arg2 fill $p/$P (object) and $t/$T/$d (string) tokens.
// Messages are written to each recipient's descriptor via CharData.Send; NPCs
// without descriptors are silently skipped.
//
// Migration note: this signature was changed from (format, ch, vch, a1, a2, to)
// to (aType, format, ch, vch, a1, a2, to) in 2026-04-18 as part of
// plan-tranche-c.md G2. The 6-arg form is gone; all callers pass aType.
//
// See C src/smaug.c act (line 3404) for the reference implementation.
func Act(aType int, format string, ch, vch *types.CharData, arg1, arg2 any, to int) {
	if format == "" {
		return
	}
	if ch == nil {
		Bug("Act: null ch. (%s)", format)
		return
	}

	switch to {
	case types.TO_CHAR:
		sendActTo(aType, ch, format, ch, vch, arg1, arg2)
	case types.TO_VICT:
		if vch == nil {
			// Match C: log and drop.
			Bug("Act: null vch with TO_VICT. %s (%s)", ch.Name, format)
			return
		}
		sendActTo(aType, vch, format, ch, vch, arg1, arg2)
	case types.TO_ROOM, types.TO_NOTVICT, types.TO_CANSEE:
		if ch.InRoom == nil {
			return
		}
		for _, p := range ch.InRoom.People {
			if p == ch {
				continue
			}
			if to == types.TO_NOTVICT && p == vch {
				continue
			}
			sendActTo(aType, p, format, ch, vch, arg1, arg2)
		}
	default:
		Bug("Act: bad `to` value %d in format %s", to, format)
	}
}

// sendActTo formats and writes the message to one recipient, prepending
// the color code for aType (if any) and appending a reset (&D) so the
// next line is not tinted. CharData.Send is a no-op when Desc is nil,
// so NPCs are handled safely. This lets future mudprog ACT_PROG triggers
// observe the formatted message even for NPC recipients (C
// src/smaug.c:3556 keeps NPCs-with-progs in the loop for this reason).
func sendActTo(aType int, recipient *types.CharData, format string, ch, vch *types.CharData, arg1, arg2 any) {
	if recipient == nil {
		return
	}
	msg := ActFormat(format, recipient, ch, vch, arg1, arg2)
	if code := atColorCode[aType]; code != "" {
		// ActFormat appends "\n\r"; splice the reset just before it so
		// the trailing newline stays at the end of the line where the
		// net/color pipeline expects it.
		if len(msg) >= 2 && msg[len(msg)-2:] == "\n\r" {
			msg = code + msg[:len(msg)-2] + "&D" + "\n\r"
		} else {
			msg = code + msg + "&D"
		}
	}
	recipient.Send(msg)
}
