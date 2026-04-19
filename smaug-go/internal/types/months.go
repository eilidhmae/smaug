package types

// MonthNames maps to C `month_name[]` at src/act_info.c:2455-2460 /
// src/timezone.c:101-106. 17 entries; indexed 0..16 matching the
// 0-indexed TimeInfoData.Month semantics.
//
// Callers that hold a 1-indexed value (HolidayData.Month, file-format
// month values, user input to `setholiday month <n>`) must subtract 1
// before indexing. This mirrors C `month_name[day->month - 1]` at
// src/holidays.c:95.
var MonthNames = [...]string{
	"Winter",
	"the Winter Wolf",
	"the Frost Giant",
	"the Old Forces",
	"the Grand Struggle",
	"the Spring",
	"Nature",
	"Futility",
	"the Dragon",
	"the Sun",
	"the Heat",
	"the Battle",
	"the Dark Shades",
	"the Shadows",
	"the Long Shadows",
	"the Ancient Darkness",
	"the Great Evil",
}

// MonthName returns the calendar name for a 0-indexed month. Out-of-
// range indices return "<unknown>" rather than panic — matches the
// "fail-soft, never abort" convention for Go file-data callers.
func MonthName(m int) string {
	if m < 0 || m >= len(MonthNames) {
		return "<unknown>"
	}
	return MonthNames[m]
}
