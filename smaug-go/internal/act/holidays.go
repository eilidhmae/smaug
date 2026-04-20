package act

import (
	"fmt"
	"strings"

	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// HolidayFilePath is the absolute path to the holiday chart on disk.
// Written once at boot (parity with act.PlanesFilePath); read by
// DoSaveHoliday and the `setholiday save` shortcut.
var HolidayFilePath string

// createDefaultAnnounce mirrors the C default at src/holidays.c:305:
// a placeholder announcement seeded on holiday creation.
const createDefaultAnnounce = "Today is the holiday of when some moron forgot to set the announcement for this one!"

// GetHoliday returns the first-matching holiday for the given 0-indexed
// month/day (TimeInfo semantics). Callers convert 0→1 by matching
// HolidayData's 1-indexed storage. Returns nil on no match or when
// WorldRef is unset. Port of C get_holiday at src/holidays.c lookups —
// see src/timezone.c:528 for a live caller (do_time holiday suffix).
func GetHoliday(month, day int) *types.HolidayData {
	if WorldRef == nil {
		return nil
	}
	// TimeInfo is 0-indexed, HolidayData is 1-indexed. Translate here.
	m := month + 1
	d := day + 1
	for _, h := range WorldRef.Holidays {
		if h != nil && h.Month == m && h.Day == d {
			return h
		}
	}
	return nil
}

// DoHolidays lists every known holiday. Player-visible (Level 0 at
// registration). Ports C do_holidays at src/holidays.c:86-98. Routes
// through the pager matching C send_to_pager / pager_printf.
func DoHolidays(ch *types.CharData, argument string) {
	if ch == nil {
		return
	}
	sendToPager(ch, "&RHoliday		       &YMonth	        &GDay\n\r")
	sendToPager(ch, "&g----------------------+----------------+---------------\r\n")
	if WorldRef == nil {
		return
	}
	for _, h := range WorldRef.Holidays {
		if h == nil {
			continue
		}
		sendToPager(ch, fmt.Sprintf("&G%-21s	&g%-11s	%-2d\r\n",
			h.Name, types.MonthName(h.Month-1), h.Day))
	}
}

// DoSaveHoliday writes the holiday chart to disk. Immortal-gated at
// registration (LEVEL_ASCENDANT per plan D6). Ports C do_saveholiday at
// src/holidays.c:243-249.
func DoSaveHoliday(ch *types.CharData, argument string) {
	if ch == nil || WorldRef == nil {
		return
	}
	if err := persist.SaveHolidays(HolidayFilePath, WorldRef.Holidays); err != nil {
		util.Bug("DoSaveHoliday: %v", err)
	}
	ch.Send("Holiday chart saved.\n\r")
}

// holidaySyntax emits the setholiday syntax help twice-used blurb.
func holidaySyntax(ch *types.CharData) {
	ch.Send("Syntax : setholiday <name> <field> <argument>\n\r")
	ch.Send("Field can be : day name create announce save delete\n\r")
}

// firstToken returns the first whitespace-delimited (or quote-delimited)
// token of argument WITHOUT lowercasing it, along with the remainder.
// DoSetHoliday uses this to preserve case when storing a holiday name;
// util.OneArgument always lowercases, which would clobber "Halloween"
// into "halloween".
func firstToken(argument string) (first, rest string) {
	argument = strings.TrimLeft(argument, " \t")
	if argument == "" {
		return "", ""
	}
	if argument[0] == '\'' || argument[0] == '"' {
		delim := argument[0]
		end := strings.IndexByte(argument[1:], delim)
		if end < 0 {
			return argument[1:], ""
		}
		return argument[1 : 1+end], strings.TrimLeft(argument[2+end:], " \t")
	}
	i := strings.IndexAny(argument, " \t")
	if i < 0 {
		return argument, ""
	}
	return argument[:i], strings.TrimLeft(argument[i:], " \t")
}

// findHoliday performs a case-insensitive match over the world's
// holiday slice by name. Returns index and pointer, or -1 / nil.
func findHoliday(name string) (int, *types.HolidayData) {
	if WorldRef == nil {
		return -1, nil
	}
	for i, h := range WorldRef.Holidays {
		if h != nil && strings.EqualFold(h.Name, name) {
			return i, h
		}
	}
	return -1, nil
}

// DoSetHoliday implements the OLC-style CRUD dispatcher for holidays.
// Immortal-gated at registration (LEVEL_ASCENDANT per plan D6). Ports C
// do_setholiday at src/holidays.c:253-416.
//
// Deliberate divergences from C (all pinned by tests):
//   - `day 1` / `month 1` are ACCEPTED (C's `<= 1` rejection is a bug;
//     intent is `< 1`, i.e. "reject 0").
//   - `create` stores TimeInfo.Month+1 / TimeInfo.Day+1 to match the
//     1-indexed file-format slot (C's unshifted store is a bug).
//   - Holiday name preserves original case on create (C uses str_dup on
//     the already-parsed arg1 which was OneArgument-lowercased by C's
//     one_argument... though C's one_argument does NOT lowercase — Go's
//     util.OneArgument DOES, so we re-extract via firstToken to preserve
//     user casing).
//   - `announce <str>` takes the full remainder after `<name> announce`
//     rather than the single word C uses — the field is meant to hold
//     prose (default is a 13-word sentence).
func DoSetHoliday(ch *types.CharData, argument string) {
	if ch == nil {
		return
	}
	if WorldRef == nil {
		holidaySyntax(ch)
		return
	}

	// Extract arg1 preserving case; subsequent args can use OneArgument.
	arg1, rest := firstToken(argument)
	if arg1 == "" {
		holidaySyntax(ch)
		return
	}

	// `setholiday save` subcommand shortcut (C :274-279).
	if strings.EqualFold(arg1, "save") {
		if err := persist.SaveHolidays(HolidayFilePath, WorldRef.Holidays); err != nil {
			util.Bug("DoSetHoliday save: %v", err)
		}
		ch.Send("Holiday chart saved.\n\r")
		return
	}

	arg2, rest := util.OneArgument(rest)
	arg3, _ := util.OneArgument(rest)

	// `create` branch — must happen BEFORE name lookup, since the
	// holiday being created won't exist yet.
	if arg2 == "create" {
		if _, existing := findHoliday(arg1); existing != nil {
			ch.Send("A holiday with that name exists already!\n\r")
			return
		}
		if WorldRef.SysData.MaxHoliday > 0 && len(WorldRef.Holidays) >= WorldRef.SysData.MaxHoliday {
			ch.Send("There are already too many holidays!\n\r")
			return
		}
		h := &types.HolidayData{
			Name: util.SmashTilde(arg1),
			// Fix C bug: C stores time_info.month unshifted (0-indexed
			// value in 1-indexed slot). Go emits +1.
			Month:    WorldRef.TimeInfo.Month + 1,
			Day:      WorldRef.TimeInfo.Day + 1,
			Announce: createDefaultAnnounce,
		}
		WorldRef.Holidays = append(WorldRef.Holidays, h)
		ch.Send("Holiday created.\n\r")
		return
	}

	// Non-create ops: find the target holiday by name.
	_, day := findHoliday(arg1)
	if day == nil {
		ch.Send("Which holiday was that?\n\r")
		return
	}

	switch arg2 {
	case "day":
		// Fix C bug: C uses `<= 1` which rejects the valid day 1.
		// Go uses `< 1` per intent.
		if arg3 == "" || !util.IsNumber(arg3) {
			ch.Sendf("You must specify a numeric value : %d - %d\n\r",
				1, maxOrDefault(WorldRef.SysData.DaysPerMonth, 30))
			return
		}
		n := atoi(arg3)
		dpm := maxOrDefault(WorldRef.SysData.DaysPerMonth, 30)
		if n < 1 || n > dpm {
			ch.Sendf("You must specify a numeric value : %d - %d\n\r", 1, dpm)
			return
		}
		day.Day = n
		ch.Send("Day changed.\n\r")
		return

	case "month":
		mpy := maxOrDefault(WorldRef.SysData.MonthsPerYear, len(types.MonthNames))
		if arg3 == "" || !util.IsNumber(arg3) {
			// Print the clean month list (no C while-loop bug).
			ch.Send("You must specify a valid month number:\n\r")
			for i, name := range types.MonthNames {
				if i >= mpy {
					break
				}
				ch.Sendf("&R(&W%d&R)&Y%s\r\n", i+1, name)
			}
			return
		}
		n := atoi(arg3)
		// Fix C bug: `<= 1` rejects month 1; Go accepts.
		if n < 1 || n > mpy {
			ch.Send("You must specify a valid month number:\n\r")
			for i, name := range types.MonthNames {
				if i >= mpy {
					break
				}
				ch.Sendf("&R(&W%d&R)&Y%s\r\n", i+1, name)
			}
			return
		}
		day.Month = n
		ch.Send("Month changed.\n\r")
		return

	case "announce":
		// Go divergence: take full remainder including arg3 and beyond,
		// since announcements are prose. Preserve case by re-extracting
		// from the original argument (util.OneArgument lowercased arg3).
		full := preserveCaseRest(argument, arg2)
		if full == "" || util.IsNumber(full) {
			ch.Send("Set the annoucement to what?\n\r")
			return
		}
		day.Announce = util.SmashTilde(full)
		ch.Send("Announcement changed.\n\r")
		return

	case "name":
		full := preserveCaseRest(argument, arg2)
		if full == "" || util.IsNumber(full) {
			ch.Send("Set the name to what?\n\r")
			return
		}
		day.Name = util.SmashTilde(full)
		ch.Send("Name changed.\n\r")
		return

	case "delete":
		if arg3 != "yes" {
			ch.Send("If you are sure, use 'delete yes'.\n\r")
			return
		}
		idx := -1
		for i, h := range WorldRef.Holidays {
			if h == day {
				idx = i
				break
			}
		}
		if idx >= 0 {
			WorldRef.Holidays = append(WorldRef.Holidays[:idx], WorldRef.Holidays[idx+1:]...)
		}
		ch.Send("&RHoliday deleted.\n\r")
		return
	}

	// Unknown arg2.
	holidaySyntax(ch)
}

// preserveCaseRest pulls the portion of the original argument string
// that follows `<name> <field>` and returns it verbatim (original
// casing). Used by the `name` and `announce` branches so user text is
// not clobbered by util.OneArgument's lowercasing.
func preserveCaseRest(argument, field string) string {
	// Skip past first token (the existing name).
	a := strings.TrimLeft(argument, " \t")
	if a == "" {
		return ""
	}
	if a[0] == '\'' || a[0] == '"' {
		delim := a[0]
		end := strings.IndexByte(a[1:], delim)
		if end < 0 {
			return ""
		}
		a = strings.TrimLeft(a[2+end:], " \t")
	} else {
		if i := strings.IndexAny(a, " \t"); i < 0 {
			return ""
		} else {
			a = strings.TrimLeft(a[i:], " \t")
		}
	}
	// Skip past the field token (case-insensitive length match).
	if !strings.HasPrefix(strings.ToLower(a), strings.ToLower(field)) {
		return ""
	}
	a = strings.TrimLeft(a[len(field):], " \t")
	return strings.TrimSpace(a)
}

// maxOrDefault returns v if positive, else fallback.
func maxOrDefault(v, fallback int) int {
	if v > 0 {
		return v
	}
	return fallback
}

// atoi parses a decimal integer, returning 0 on failure. Only called
// after util.IsNumber confirmation so errors are impossible.
func atoi(s string) int {
	n := 0
	neg := false
	for i, c := range s {
		if i == 0 && c == '-' {
			neg = true
			continue
		}
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	if neg {
		return -n
	}
	return n
}
