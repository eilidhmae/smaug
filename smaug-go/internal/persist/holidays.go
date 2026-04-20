package persist

import (
	"fmt"
	"os"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// DefaultHolidayAnnounce is the fallback announce string used when a
// block omits the Announce field. Mirrors C `fread_day` at
// src/holidays.c:130.
const DefaultHolidayAnnounce = "Today is a holiday, but who the hell knows which one."

// LoadHolidays reads a SMAUG `holidays.dat` file into a slice. Missing
// file returns `(nil, nil)` (matches C `load_holidays` at
// src/holidays.c:149-210: missing file is non-fatal).
//
// Malformed blocks (unknown top-level sections, unknown keys within a
// block) log via util.Bug and skip, matching the C "fail-soft" loader
// convention. When the number of successfully-loaded holidays reaches
// maxHolidays, further `#HOLIDAY` blocks are dropped with a bug log
// (mirrors C `:188-192`).
//
// C divergence: the shipped stock `db/system/holidays.dat` lacks a
// trailing `#END` terminator. Under C, EOF at section-header position
// triggers a spurious `load_holidays: # not found.` bug log on every
// boot. This Go port treats EOF at section-header position as an
// implicit `#END` — cleaner behavior, pinned by
// TestLoadHolidays_ShippedFileLoadsNoBug.
func LoadHolidays(path string, maxHolidays int) ([]*types.HolidayData, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	sc := NewScanner(f, path)
	list := make([]*types.HolidayData, 0)
	for {
		c := sc.ReadLetter()
		if c == 0 {
			// EOF at section-header position — implicit #END.
			return list, nil
		}
		if c == '*' {
			// Line comment.
			_ = sc.ReadToEOL()
			continue
		}
		if c != '#' {
			util.Bug("LoadHolidays: %s: expected '#', got %q", path, c)
			return list, nil
		}
		word := sc.ReadWord()
		switch word {
		case "END":
			return list, nil
		case "HOLIDAY":
			if maxHolidays > 0 && len(list) >= maxHolidays {
				util.Bug("load_holidays: more holidays than %d", maxHolidays)
				return list, nil
			}
			if h := readHolidayBlock(sc, path); h != nil {
				list = append(list, h)
			}
		default:
			util.Bug("LoadHolidays: %s: unknown section %q", path, word)
			// Drain the unknown section until we see `End` or EOF so
			// that the next iteration can start at a section header.
			drainHolidayBlock(sc)
		}
	}
}

// readHolidayBlock reads one `#HOLIDAY` block's KVPs up to `End`.
// Mirrors C `fread_day` at src/holidays.c:100-147. Unknown keys log via
// util.Bug and are drained. If Announce is missing, the default string
// is seeded (matches C `:130`).
func readHolidayBlock(sc *Scanner, path string) *types.HolidayData {
	h := &types.HolidayData{}
	for {
		word := sc.ReadWord()
		if word == "" {
			util.Bug("LoadHolidays: %s: unexpected EOF inside #HOLIDAY block", path)
			if h.Announce == "" {
				h.Announce = DefaultHolidayAnnounce
			}
			return h
		}
		switch word {
		case "End":
			if h.Announce == "" {
				h.Announce = DefaultHolidayAnnounce
			}
			return h
		case "Name":
			h.Name = sc.ReadString()
		case "Announce":
			h.Announce = sc.ReadString()
		case "Month":
			h.Month = sc.ReadNumber()
		case "Day":
			h.Day = sc.ReadNumber()
		case "*":
			// Comment line inside a block.
			_ = sc.ReadToEOL()
		default:
			util.Bug("fread_day: no match: %s", word)
		}
	}
}

// drainHolidayBlock consumes tokens until `End` or EOF. Used when the
// top-level section header is unknown; keeps the scanner aligned with
// the next section header.
func drainHolidayBlock(sc *Scanner) {
	for {
		word := sc.ReadWord()
		if word == "" || word == "End" {
			return
		}
	}
}

// SaveHolidays truncate-writes holidays to path in the C-compatible
// block format. Mirrors C `save_holidays` at src/holidays.c:212-241.
//
// Format (per holiday):
//
//	#HOLIDAY
//	Name\t\t<name>~
//	Announce\t<announce>~
//	Month\t\t<month>
//	Day\t\t<day>
//	End
//	<blank line>
//
// Terminator: `#END\n`.
//
// Write errors propagate up; caller logs via util.Bug and emits the
// user-facing confirmation unconditionally (matches C's no-error-to-user
// convention).
func SaveHolidays(path string, list []*types.HolidayData) error {
	// 0o600 — game-data files are kept private to the smaug user. Matches
	// the hotboot precedent (persist/hotboot.go) that landed 2026-04-19.
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, h := range list {
		if h == nil {
			continue
		}
		if _, err := fmt.Fprintf(f,
			"#HOLIDAY\nName\t\t%s~\nAnnounce\t%s~\nMonth\t\t%d\nDay\t\t%d\nEnd\n\n",
			h.Name, h.Announce, h.Month, h.Day); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(f, "#END"); err != nil {
		return err
	}
	return nil
}
