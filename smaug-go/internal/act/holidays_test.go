package act

import (
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// setupHolidaysWorld creates a fresh world, installs it at WorldRef,
// restores on cleanup, and returns the world ready for mutation.
func setupHolidaysWorld(t *testing.T) *world.World {
	t.Helper()
	w := world.New("/tmp/test-holidays")
	w.SysData.MaxHoliday = 32
	w.SysData.DaysPerMonth = 30
	w.SysData.MonthsPerYear = 17
	prev := WorldRef
	WorldRef = w
	t.Cleanup(func() { WorldRef = prev })
	return w
}

func captureActBug() (*atomic.Int64, func()) {
	var count atomic.Int64
	prev := util.BugSink
	util.BugSink = func(msg string) { count.Add(1) }
	return &count, func() { util.BugSink = prev }
}

// --- DoHolidays --------------------------------------------------------

func TestDoHolidays_EmptyListPrintsHeaderOnly(t *testing.T) {
	setupHolidaysWorld(t)
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoHolidays(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Holiday") || !strings.Contains(out, "Month") || !strings.Contains(out, "Day") {
		t.Errorf("expected header tokens, got %q", out)
	}
}

func TestDoHolidays_SingleHolidayRendersRow(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.Holidays = []*types.HolidayData{{Month: 1, Day: 1, Name: "New Year's Day", Announce: ""}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoHolidays(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "New Year's Day") {
		t.Errorf("missing name: %q", out)
	}
	if !strings.Contains(out, "Winter") {
		t.Errorf("missing month name 'Winter': %q", out)
	}
}

func TestDoHolidays_FiveHolidaysAllPresent(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.Holidays = []*types.HolidayData{
		{Month: 1, Day: 1, Name: "Alpha"},
		{Month: 2, Day: 2, Name: "Beta"},
		{Month: 3, Day: 3, Name: "Gamma"},
		{Month: 4, Day: 4, Name: "Delta"},
		{Month: 5, Day: 5, Name: "Epsilon"},
	}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoHolidays(ch, "")
	out := readOutput(ch, client)

	for _, n := range []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon"} {
		if !strings.Contains(out, n) {
			t.Errorf("missing %q in output %q", n, out)
		}
	}
}

func TestDoHolidays_MonthOutOfRangeShowsUnknown(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.Holidays = []*types.HolidayData{{Month: 99, Day: 5, Name: "Garbage"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoHolidays(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "<unknown>") {
		t.Errorf("expected <unknown> for OOB month, got %q", out)
	}
}

// --- DoSaveHoliday -----------------------------------------------------

func TestDoSaveHoliday_SendsConfirmation(t *testing.T) {
	setupHolidaysWorld(t)
	tmp := filepath.Join(t.TempDir(), "holidays.dat")
	prev := HolidayFilePath
	HolidayFilePath = tmp
	t.Cleanup(func() { HolidayFilePath = prev })

	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSaveHoliday(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Holiday chart saved.") {
		t.Errorf("missing confirmation: %q", out)
	}
	if _, err := os.Stat(tmp); err != nil {
		t.Errorf("save did not create file: %v", err)
	}
}

func TestDoSaveHoliday_SaveErrorLogsBug(t *testing.T) {
	setupHolidaysWorld(t)
	dir := filepath.Join(t.TempDir(), "readonly")
	if err := os.Mkdir(dir, 0o555); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	defer os.Chmod(dir, 0o755)
	tmp := filepath.Join(dir, "holidays.dat")
	prev := HolidayFilePath
	HolidayFilePath = tmp
	t.Cleanup(func() { HolidayFilePath = prev })

	count, restore := captureActBug()
	defer restore()

	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSaveHoliday(ch, "")
	out := readOutput(ch, client)

	// C pattern: user still gets the confirmation even if disk write fails.
	if !strings.Contains(out, "Holiday chart saved.") {
		t.Errorf("expected confirmation even on write fail, got %q", out)
	}
	if count.Load() == 0 {
		t.Errorf("expected bug log on write failure")
	}
}

// --- DoSetHoliday create ----------------------------------------------

func TestDoSetHoliday_NoArgShowsSyntax(t *testing.T) {
	setupHolidaysWorld(t)
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Syntax : setholiday") {
		t.Errorf("expected syntax help, got %q", out)
	}
}

func TestDoSetHoliday_CreateAddsToList(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.TimeInfo = types.TimeInfoData{Month: 9, Day: 30, Year: 2026}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "Halloween create")
	out := readOutput(ch, client)

	if len(w.Holidays) != 1 {
		t.Fatalf("len = %d, want 1", len(w.Holidays))
	}
	h := w.Holidays[0]
	if h.Month != 10 { // TimeInfo.Month=9 → 9+1
		t.Errorf("Month = %d, want 10 (0→1 indexed shift)", h.Month)
	}
	if h.Day != 31 { // TimeInfo.Day=30 → 30+1
		t.Errorf("Day = %d, want 31 (0→1 indexed shift)", h.Day)
	}
	if h.Announce != createDefaultAnnounce {
		t.Errorf("Announce = %q, want default", h.Announce)
	}
	if !strings.Contains(out, "Holiday created.") {
		t.Errorf("missing confirmation: %q", out)
	}
}

func TestDoSetHoliday_CreatePreservesNameCase(t *testing.T) {
	w := setupHolidaysWorld(t)
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "Halloween create")
	_ = readOutput(ch, client)

	if len(w.Holidays) != 1 {
		t.Fatalf("len = %d, want 1", len(w.Holidays))
	}
	if w.Holidays[0].Name != "Halloween" {
		t.Errorf("Name = %q, want %q (original case preserved)", w.Holidays[0].Name, "Halloween")
	}
}

func TestDoSetHoliday_LookupCaseInsensitive(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.Holidays = []*types.HolidayData{{Month: 10, Day: 31, Name: "Halloween", Announce: "Trick!"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "halloween day 15")
	_ = readOutput(ch, client)

	if w.Holidays[0].Day != 15 {
		t.Errorf("Day = %d, want 15 (case-insensitive lookup failed)", w.Holidays[0].Day)
	}
}

func TestDoSetHoliday_CreateRejectsDuplicateName(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.Holidays = []*types.HolidayData{{Month: 10, Day: 31, Name: "Halloween"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "Halloween create")
	out := readOutput(ch, client)

	if !strings.Contains(out, "exists already") {
		t.Errorf("expected duplicate-reject, got %q", out)
	}
	if len(w.Holidays) != 1 {
		t.Errorf("list grew on duplicate; len = %d", len(w.Holidays))
	}
}

func TestDoSetHoliday_CreateRejectsOverMax(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.SysData.MaxHoliday = 2
	w.Holidays = []*types.HolidayData{
		{Month: 1, Day: 1, Name: "A"},
		{Month: 2, Day: 2, Name: "B"},
	}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "C create")
	out := readOutput(ch, client)

	if !strings.Contains(out, "too many holidays") {
		t.Errorf("expected too-many reject, got %q", out)
	}
	if len(w.Holidays) != 2 {
		t.Errorf("list grew over max; len = %d", len(w.Holidays))
	}
}

// --- DoSetHoliday day -------------------------------------------------

func TestDoSetHoliday_DayUpdatesValue(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.Holidays = []*types.HolidayData{{Month: 1, Day: 1, Name: "Test"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "Test day 15")
	out := readOutput(ch, client)

	if w.Holidays[0].Day != 15 {
		t.Errorf("Day = %d, want 15", w.Holidays[0].Day)
	}
	if !strings.Contains(out, "Day changed.") {
		t.Errorf("missing confirmation: %q", out)
	}
}

func TestDoSetHoliday_DayRejectsNonNumeric(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.Holidays = []*types.HolidayData{{Month: 1, Day: 1, Name: "Test"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "Test day abc")
	out := readOutput(ch, client)

	if !strings.Contains(out, "numeric value") {
		t.Errorf("expected numeric-reject, got %q", out)
	}
	if w.Holidays[0].Day != 1 {
		t.Errorf("Day mutated on bad input: %d", w.Holidays[0].Day)
	}
}

func TestDoSetHoliday_DayRejectsZero(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.Holidays = []*types.HolidayData{{Month: 1, Day: 5, Name: "Test"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "Test day 0")
	_ = readOutput(ch, client)

	if w.Holidays[0].Day != 5 {
		t.Errorf("Day 0 accepted; now = %d", w.Holidays[0].Day)
	}
}

// Pin-against-C-bug: Go must ACCEPT day 1 (C rejects via `<= 1`).
func TestDoSetHoliday_DayAcceptsOne(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.Holidays = []*types.HolidayData{{Month: 1, Day: 5, Name: "Test"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "Test day 1")
	out := readOutput(ch, client)

	if w.Holidays[0].Day != 1 {
		t.Errorf("Day = %d, want 1 (C-bug should be fixed in Go)", w.Holidays[0].Day)
	}
	if !strings.Contains(out, "Day changed.") {
		t.Errorf("expected Day changed. confirmation, got %q", out)
	}
}

func TestDoSetHoliday_DayRejectsOverMax(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.SysData.DaysPerMonth = 30
	w.Holidays = []*types.HolidayData{{Month: 1, Day: 5, Name: "Test"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "Test day 31")
	_ = readOutput(ch, client)

	if w.Holidays[0].Day != 5 {
		t.Errorf("Day > max accepted; now = %d", w.Holidays[0].Day)
	}
}

// --- DoSetHoliday month -----------------------------------------------

func TestDoSetHoliday_MonthUpdatesValue(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.Holidays = []*types.HolidayData{{Month: 1, Day: 1, Name: "Test"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "Test month 10")
	out := readOutput(ch, client)

	if w.Holidays[0].Month != 10 {
		t.Errorf("Month = %d, want 10", w.Holidays[0].Month)
	}
	if !strings.Contains(out, "Month changed.") {
		t.Errorf("missing confirmation: %q", out)
	}
}

// Pin-against-C-bug: Go must ACCEPT month 1.
func TestDoSetHoliday_MonthAcceptsOne(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.Holidays = []*types.HolidayData{{Month: 5, Day: 1, Name: "Test"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "Test month 1")
	_ = readOutput(ch, client)

	if w.Holidays[0].Month != 1 {
		t.Errorf("Month = %d, want 1 (C-bug should be fixed in Go)", w.Holidays[0].Month)
	}
}

func TestDoSetHoliday_MonthMissingArgShowsList(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.Holidays = []*types.HolidayData{{Month: 1, Day: 1, Name: "Test"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "Test month")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Winter") {
		t.Errorf("expected month list (Winter), got %q", out)
	}
	if !strings.Contains(out, "the Great Evil") {
		t.Errorf("expected last month in list, got %q", out)
	}
}

// --- DoSetHoliday announce / name ------------------------------------

func TestDoSetHoliday_AnnounceSetsString(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.Holidays = []*types.HolidayData{{Month: 1, Day: 1, Name: "Test"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "Test announce Trick or treat!")
	out := readOutput(ch, client)

	if w.Holidays[0].Announce != "Trick or treat!" {
		t.Errorf("Announce = %q, want %q", w.Holidays[0].Announce, "Trick or treat!")
	}
	if !strings.Contains(out, "Announcement changed.") {
		t.Errorf("missing confirmation: %q", out)
	}
}

func TestDoSetHoliday_AnnounceRejectsEmpty(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.Holidays = []*types.HolidayData{{Month: 1, Day: 1, Name: "Test", Announce: "orig"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "Test announce")
	_ = readOutput(ch, client)

	if w.Holidays[0].Announce != "orig" {
		t.Errorf("Announce mutated on empty: %q", w.Holidays[0].Announce)
	}
}

func TestDoSetHoliday_NameRenames(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.Holidays = []*types.HolidayData{{Month: 1, Day: 1, Name: "Halloween"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "Halloween name AllHallows")
	_ = readOutput(ch, client)

	if w.Holidays[0].Name != "AllHallows" {
		t.Errorf("Name = %q, want %q", w.Holidays[0].Name, "AllHallows")
	}
}

// Defense-in-depth: an immortal who passes a literal `~` in the holiday
// name or announcement could otherwise inject a structurally-corrupt
// `#HOLIDAY` block into holidays.dat (the on-disk format is tilde-
// terminated). SmashTilde converts the embedded `~` to `-` before storage.
func TestDoSetHoliday_NameSmashesTilde(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.Holidays = []*types.HolidayData{{Month: 1, Day: 1, Name: "Old"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "Old name foo~injected")
	_ = readOutput(ch, client)

	if w.Holidays[0].Name != "foo-injected" {
		t.Errorf("Name = %q, want %q (tilde must be smashed)", w.Holidays[0].Name, "foo-injected")
	}
}

func TestDoSetHoliday_AnnounceSmashesTilde(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.Holidays = []*types.HolidayData{{Month: 1, Day: 1, Name: "Test"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "Test announce hello~Name injected~")
	_ = readOutput(ch, client)

	if w.Holidays[0].Announce != "hello-Name injected-" {
		t.Errorf("Announce = %q, want %q (tildes must be smashed)", w.Holidays[0].Announce, "hello-Name injected-")
	}
}

func TestDoSetHoliday_CreateSmashesTildeInName(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.TimeInfo.Month = 0
	w.TimeInfo.Day = 0
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "Foo~bar create")
	_ = readOutput(ch, client)

	if len(w.Holidays) != 1 || w.Holidays[0].Name != "Foo-bar" {
		t.Fatalf("Holidays = %+v, want one entry named %q", w.Holidays, "Foo-bar")
	}
}

// --- DoSetHoliday delete ---------------------------------------------

func TestDoSetHoliday_DeleteRequiresYes(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.Holidays = []*types.HolidayData{{Month: 1, Day: 1, Name: "Test"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "Test delete")
	out := readOutput(ch, client)

	if len(w.Holidays) != 1 {
		t.Errorf("deleted without 'yes' confirmation; len = %d", len(w.Holidays))
	}
	if !strings.Contains(out, "delete yes") {
		t.Errorf("expected 'delete yes' hint, got %q", out)
	}
}

func TestDoSetHoliday_DeleteYesRemoves(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.Holidays = []*types.HolidayData{{Month: 1, Day: 1, Name: "Test"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "Test delete yes")
	out := readOutput(ch, client)

	if len(w.Holidays) != 0 {
		t.Errorf("len = %d, want 0", len(w.Holidays))
	}
	if !strings.Contains(out, "Holiday deleted.") {
		t.Errorf("missing confirmation: %q", out)
	}
}

// --- DoSetHoliday not found + save subcommand ------------------------

func TestDoSetHoliday_NotFoundMessage(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.Holidays = []*types.HolidayData{{Month: 1, Day: 1, Name: "Test"}}
	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "NonExistent day 1")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Which holiday was that?") {
		t.Errorf("expected not-found, got %q", out)
	}
	if len(w.Holidays) != 1 {
		t.Errorf("list changed on not-found; len = %d", len(w.Holidays))
	}
}

func TestDoSetHoliday_SaveSubcommandInvokesSaver(t *testing.T) {
	setupHolidaysWorld(t)
	tmp := filepath.Join(t.TempDir(), "holidays.dat")
	prev := HolidayFilePath
	HolidayFilePath = tmp
	t.Cleanup(func() { HolidayFilePath = prev })

	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "save")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Holiday chart saved.") {
		t.Errorf("expected save confirmation, got %q", out)
	}
	if _, err := os.Stat(tmp); err != nil {
		t.Errorf("save did not create file: %v", err)
	}
}

// --- GetHoliday (A14) ------------------------------------------------

func TestGetHoliday_NewYearsDay(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.Holidays = []*types.HolidayData{{Month: 1, Day: 1, Name: "New Year's Day"}}

	// TimeInfo is 0-indexed; Month=0 Day=0 → 1/1 in 1-indexed HolidayData.
	h := GetHoliday(0, 0)
	if h == nil {
		t.Fatalf("GetHoliday(0,0) = nil, want New Year's Day")
	}
	if h.Name != "New Year's Day" {
		t.Errorf("Name = %q, want New Year's Day", h.Name)
	}
}

func TestGetHoliday_NonHolidayReturnsNil(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.Holidays = []*types.HolidayData{{Month: 1, Day: 1, Name: "New Year's Day"}}

	if h := GetHoliday(5, 5); h != nil {
		t.Errorf("GetHoliday(5,5) = %+v, want nil", h)
	}
}

func TestGetHoliday_NilWorldRef(t *testing.T) {
	prev := WorldRef
	WorldRef = nil
	t.Cleanup(func() { WorldRef = prev })
	if h := GetHoliday(0, 0); h != nil {
		t.Errorf("nil WorldRef should return nil, got %+v", h)
	}
}

// --- DoTime suffix (A15) ---------------------------------------------

func TestDoTime_EmitsHolidaySuffix(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.TimeInfo = types.TimeInfoData{Hour: 12, Day: 0, Month: 0, Year: 2026}
	w.Holidays = []*types.HolidayData{{Month: 1, Day: 1, Name: "New Year's Day"}}

	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoTime(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "It's a holiday today") {
		t.Errorf("missing holiday suffix, got %q", out)
	}
	if !strings.Contains(out, "New Year's Day") {
		t.Errorf("missing holiday name, got %q", out)
	}
}

func TestDoTime_NoSuffixOnNonHoliday(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.TimeInfo = types.TimeInfoData{Hour: 12, Day: 4, Month: 4, Year: 2026}
	w.Holidays = []*types.HolidayData{{Month: 1, Day: 1, Name: "New Year's Day"}}

	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoTime(ch, "")
	out := readOutput(ch, client)

	if strings.Contains(out, "It's a holiday today") {
		t.Errorf("unexpected holiday suffix on non-holiday: %q", out)
	}
}

// --- round-trip via persistence --------------------------------------

func TestDoSetHoliday_CreateSaveReload(t *testing.T) {
	w := setupHolidaysWorld(t)
	w.TimeInfo = types.TimeInfoData{Month: 2, Day: 4, Year: 2026}
	tmp := filepath.Join(t.TempDir(), "holidays.dat")
	prev := HolidayFilePath
	HolidayFilePath = tmp
	t.Cleanup(func() { HolidayFilePath = prev })

	ch, client := makeTestChar("Tester")
	defer client.Close()

	DoSetHoliday(ch, "MyDay create")
	_ = readOutput(ch, client)
	DoSetHoliday(ch, "save")
	_ = readOutput(ch, client)

	list, err := persist.LoadHolidays(tmp, 32)
	if err != nil {
		t.Fatalf("reload err: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len = %d, want 1", len(list))
	}
	if list[0].Name != "MyDay" {
		t.Errorf("Name = %q, want MyDay", list[0].Name)
	}
	if list[0].Month != 3 { // TimeInfo.Month=2 → 3
		t.Errorf("Month = %d, want 3", list[0].Month)
	}
	if list[0].Day != 5 { // TimeInfo.Day=4 → 5
		t.Errorf("Day = %d, want 5", list[0].Day)
	}
}
