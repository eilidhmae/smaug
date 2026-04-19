package types

import "testing"

func TestMonthNames_HasSeventeenEntries(t *testing.T) {
	if len(MonthNames) != 17 {
		t.Fatalf("len(MonthNames) = %d, want 17", len(MonthNames))
	}
}

func TestMonthNames_WinterFirst(t *testing.T) {
	if MonthNames[0] != "Winter" {
		t.Errorf("MonthNames[0] = %q, want %q", MonthNames[0], "Winter")
	}
}

func TestMonthNames_LastEntry(t *testing.T) {
	if MonthNames[16] != "the Great Evil" {
		t.Errorf("MonthNames[16] = %q, want %q", MonthNames[16], "the Great Evil")
	}
}

func TestMonthName_ZeroIndexed(t *testing.T) {
	if got := MonthName(0); got != "Winter" {
		t.Errorf("MonthName(0) = %q, want %q", got, "Winter")
	}
	if got := MonthName(16); got != "the Great Evil" {
		t.Errorf("MonthName(16) = %q, want %q", got, "the Great Evil")
	}
}

func TestMonthName_OutOfRange(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("MonthName OOB panicked: %v", r)
		}
	}()
	if got := MonthName(-1); got != "<unknown>" {
		t.Errorf("MonthName(-1) = %q, want <unknown>", got)
	}
	if got := MonthName(17); got != "<unknown>" {
		t.Errorf("MonthName(17) = %q, want <unknown>", got)
	}
	if got := MonthName(1000); got != "<unknown>" {
		t.Errorf("MonthName(1000) = %q, want <unknown>", got)
	}
}

func TestMonthName_MatchesCArray(t *testing.T) {
	// Verbatim pin against C src/act_info.c:2455-2460 /
	// src/timezone.c:101-106. Any drift here is an intentional semantic
	// change that needs review.
	want := []string{
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
	if len(want) != len(MonthNames) {
		t.Fatalf("C array len %d, Go len %d", len(want), len(MonthNames))
	}
	for i, name := range want {
		if MonthNames[i] != name {
			t.Errorf("MonthNames[%d] = %q, want %q", i, MonthNames[i], name)
		}
		if got := MonthName(i); got != name {
			t.Errorf("MonthName(%d) = %q, want %q", i, got, name)
		}
	}
}

func TestHolidayData_ZeroValue(t *testing.T) {
	h := &HolidayData{}
	if h.Month != 0 || h.Day != 0 || h.Name != "" || h.Announce != "" {
		t.Errorf("HolidayData zero value = %+v, want all zeroes", h)
	}
}
