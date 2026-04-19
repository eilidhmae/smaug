package persist

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// --- LoadHolidays -----------------------------------------------------

func TestLoadHolidays_MissingFileReturnsNilNil(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nope.dat")
	list, err := LoadHolidays(path, 32)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if list != nil {
		t.Fatalf("list = %v, want nil", list)
	}
}

func TestLoadHolidays_EmptyFileReturnsEmpty(t *testing.T) {
	list, err := LoadHolidays("testdata/holidays_empty.dat", 32)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if list == nil || len(list) != 0 {
		t.Fatalf("list = %v, want non-nil len 0", list)
	}
}

func TestLoadHolidays_TwoHolidaysRoundTrip(t *testing.T) {
	list, err := LoadHolidays("testdata/holidays_two.dat", 32)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len = %d, want 2", len(list))
	}
	want0 := &types.HolidayData{Month: 1, Day: 1, Name: "New Year's Day", Announce: "Today is the first day of a new year."}
	want1 := &types.HolidayData{Month: 12, Day: 30, Name: "Year's End", Announce: "Today is the last day of the year."}
	if !reflect.DeepEqual(list[0], want0) {
		t.Errorf("list[0] = %+v, want %+v", list[0], want0)
	}
	if !reflect.DeepEqual(list[1], want1) {
		t.Errorf("list[1] = %+v, want %+v", list[1], want1)
	}
}

func TestLoadHolidays_MaxHolidaysEnforced(t *testing.T) {
	count, restore := captureBug()
	defer restore()

	list, err := LoadHolidays("testdata/holidays_three.dat", 2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len = %d, want 2 (ceiling)", len(list))
	}
	if count.Load() == 0 {
		t.Errorf("expected bug log for over-ceiling block, got none")
	}
}

func TestLoadHolidays_UnknownSectionLogsBugContinues(t *testing.T) {
	count, restore := captureBug()
	defer restore()

	list, err := LoadHolidays("testdata/holidays_malformed.dat", 32)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len = %d, want 2 (both valid blocks despite the #BOGUS)", len(list))
	}
	if list[0].Name != "First" {
		t.Errorf("list[0].Name = %q, want First", list[0].Name)
	}
	if list[1].Name != "Second" {
		t.Errorf("list[1].Name = %q, want Second", list[1].Name)
	}
	if count.Load() == 0 {
		t.Errorf("expected bug log, got none")
	}
}

func TestLoadHolidays_UnknownKeyInBlockLogsBugContinues(t *testing.T) {
	count, restore := captureBug()
	defer restore()

	list, err := LoadHolidays("testdata/holidays_malformed.dat", 32)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	// First block has `Garbage 42` between Day and End — still loads
	// with Name/Month/Day intact.
	if len(list) < 1 {
		t.Fatalf("got no holidays")
	}
	if list[0].Month != 2 || list[0].Day != 5 {
		t.Errorf("list[0] Month/Day = %d/%d, want 2/5", list[0].Month, list[0].Day)
	}
	if count.Load() == 0 {
		t.Errorf("expected bug log for Garbage key, got none")
	}
}

func TestLoadHolidays_MissingAnnounceDefaults(t *testing.T) {
	list, err := LoadHolidays("testdata/holidays_no_announce.dat", 32)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len = %d, want 1", len(list))
	}
	if list[0].Announce != DefaultHolidayAnnounce {
		t.Errorf("Announce = %q, want default %q", list[0].Announce, DefaultHolidayAnnounce)
	}
}

func TestLoadHolidays_ShippedFileLoads(t *testing.T) {
	// Pinning test against the repo-shipped holidays.dat.
	list, err := LoadHolidays("/home/eilidh/src/smaug/db/system/holidays.dat", 32)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len = %d, want 2", len(list))
	}
	want0 := &types.HolidayData{Month: 1, Day: 1, Name: "New Year's Day", Announce: "Today is the first day of a new year."}
	want1 := &types.HolidayData{Month: 12, Day: 30, Name: "Year's End", Announce: "Today is the last day of the year."}
	if !reflect.DeepEqual(list[0], want0) {
		t.Errorf("list[0] = %+v, want %+v", list[0], want0)
	}
	if !reflect.DeepEqual(list[1], want1) {
		t.Errorf("list[1] = %+v, want %+v", list[1], want1)
	}
}

func TestLoadHolidays_ShippedFileLoadsNoBug(t *testing.T) {
	// The shipped file lacks `#END`; Go port treats EOF at section-
	// header position as implicit #END — no bug log should fire.
	count, restore := captureBug()
	defer restore()

	_, err := LoadHolidays("/home/eilidh/src/smaug/db/system/holidays.dat", 32)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got := count.Load(); got != 0 {
		t.Errorf("expected 0 bug logs, got %d", got)
	}
}

func TestLoadHolidays_ExplicitEndTerminates(t *testing.T) {
	list, err := LoadHolidays("testdata/holidays_three.dat", 32)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("len = %d, want 3", len(list))
	}
}

// --- SaveHolidays -----------------------------------------------------

func TestSaveHolidays_RoundTripStructEqual(t *testing.T) {
	seed := []*types.HolidayData{
		{Month: 1, Day: 1, Name: "New Year's Day", Announce: "Today is the first day of a new year."},
		{Month: 12, Day: 30, Name: "Year's End", Announce: "Today is the last day of the year."},
	}
	path := filepath.Join(t.TempDir(), "out.dat")
	if err := SaveHolidays(path, seed); err != nil {
		t.Fatalf("save err: %v", err)
	}
	got, err := LoadHolidays(path, 32)
	if err != nil {
		t.Fatalf("load err: %v", err)
	}
	if !reflect.DeepEqual(seed, got) {
		t.Errorf("round-trip mismatch:\n  seed=%+v\n  got =%+v", seed, got)
	}
}

func TestSaveHolidays_ProducesExpectedBytes(t *testing.T) {
	seed := []*types.HolidayData{
		{Month: 1, Day: 1, Name: "A", Announce: "B"},
	}
	path := filepath.Join(t.TempDir(), "out.dat")
	if err := SaveHolidays(path, seed); err != nil {
		t.Fatalf("save err: %v", err)
	}
	bs, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read err: %v", err)
	}
	want := "#HOLIDAY\nName\t\tA~\nAnnounce\tB~\nMonth\t\t1\nDay\t\t1\nEnd\n\n#END\n"
	if string(bs) != want {
		t.Errorf("bytes mismatch:\n got=%q\nwant=%q", string(bs), want)
	}
}

func TestSaveHolidays_EmptyListWritesOnlyTerminator(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.dat")
	if err := SaveHolidays(path, nil); err != nil {
		t.Fatalf("save err: %v", err)
	}
	bs, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read err: %v", err)
	}
	if string(bs) != "#END\n" {
		t.Errorf("bytes = %q, want %q", string(bs), "#END\n")
	}
}

func TestSaveHolidays_WriteErrorPropagates(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "readonly")
	if err := os.Mkdir(dir, 0o555); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	defer os.Chmod(dir, 0o755)

	path := filepath.Join(dir, "out.dat")
	err := SaveHolidays(path, []*types.HolidayData{{Name: "x", Announce: "y"}})
	if err == nil {
		t.Errorf("expected err for read-only dir, got nil")
	}
}

// --- ensure captureBug is reachable from this file ---

func TestHolidays_BugSinkPlumbingSmoke(t *testing.T) {
	count, restore := captureBug()
	defer restore()
	util.Bug("test %d", 1)
	if count.Load() != 1 {
		t.Errorf("bug sink count = %d, want 1", count.Load())
	}
}

// captureBugMessages installs a BugSink that records every message so
// tests can assert on specific substrings (not just count).
func captureBugMessages() (*[]string, func()) {
	var msgs []string
	prev := util.BugSink
	util.BugSink = func(m string) { msgs = append(msgs, m) }
	return &msgs, func() { util.BugSink = prev }
}

func TestLoadHolidays_UnknownSectionEmitsSpecificBug(t *testing.T) {
	msgs, restore := captureBugMessages()
	defer restore()

	_, err := LoadHolidays("testdata/holidays_malformed.dat", 32)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	found := false
	for _, m := range *msgs {
		if strings.Contains(m, "unknown section") && strings.Contains(m, "BOGUS") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected unknown-section bug mentioning BOGUS, got %q", *msgs)
	}
}

// --- shipped file edge: confirm tilde handling doesn't trip on `\n\r`
// bake-in by ReadString.

func TestLoadHolidays_NameSurvivesRoundTrip(t *testing.T) {
	seed := []*types.HolidayData{{Month: 5, Day: 10, Name: "With Spaces And 'quotes'", Announce: "Body with punctuation! And apostrophes'."}}
	path := filepath.Join(t.TempDir(), "out.dat")
	if err := SaveHolidays(path, seed); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := LoadHolidays(path, 32)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d", len(got))
	}
	// ReadString preserves internal whitespace.
	if !strings.Contains(got[0].Name, "With Spaces") {
		t.Errorf("Name lost: %q", got[0].Name)
	}
	if !strings.Contains(got[0].Announce, "apostrophes") {
		t.Errorf("Announce lost: %q", got[0].Announce)
	}
}
