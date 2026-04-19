package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// -----------------------------------------------------------------------------
// G1 — pure helper tests (computePositions, precipBucket, renderStarmap).
// -----------------------------------------------------------------------------

// A1 — starMap verbatim byte-for-byte against src/starmap.c:59-68.
// Transcription errors here are the most dangerous class of bug in this port.
func TestStarMap_AllRowsAreSeventyTwoBytes(t *testing.T) {
	if len(starMap) != starmapHeight {
		t.Fatalf("starMap has %d rows, want %d", len(starMap), starmapHeight)
	}
	for i, row := range starMap {
		if len(row) != starmapWidth {
			t.Errorf("starMap row %d has %d bytes, want %d: %q", i, len(row), starmapWidth, row)
		}
	}
}

func TestStarMap_ExactBytesPerRow(t *testing.T) {
	// These must match src/starmap.c:60-67 exactly. Do not edit without also
	// editing the C source — this table IS the gameplay.
	want := []string{
		"                                               C. C.                  g*",
		"    O:       R*        G*    G.  W* W. W.          C. C.    Y* Y. Y.    ",
		"  O*.                c.          W.W.     W.            C.       Y..Y.  ",
		"O.O. O.              c.  G..G.           W:      B*                   Y.",
		"     O.    c.     c.                     W. W.                  r*    Y.",
		"     O.c.     c.      G.             P..     W.        p.      Y.   Y:  ",
		"        c.                    G*    P.  P.           p.  p:     Y.   Y. ",
		"                 b*             P.: P*                 p.p:             ",
	}
	for i, w := range want {
		if starMap[i] != w {
			t.Errorf("starMap row %d mismatch\n got: %q\nwant: %q", i, starMap[i], w)
		}
	}
}

// A1 — sunMap verbatim from src/starmap.c:75-79.
func TestSunMap_ExactBytes(t *testing.T) {
	want := []string{
		`\` + "`" + `|'/`,
		`- O -`,
		`/.|.\`,
	}
	if len(sunMap) != 3 {
		t.Fatalf("sunMap has %d rows, want 3", len(sunMap))
	}
	for i, w := range want {
		if sunMap[i] != w {
			t.Errorf("sunMap row %d: got %q, want %q", i, sunMap[i], w)
		}
		if len(sunMap[i]) != 5 {
			t.Errorf("sunMap row %d length %d, want 5", i, len(sunMap[i]))
		}
	}
}

// A1 — moonMap verbatim from src/starmap.c:81-85.
func TestMoonMap_ExactBytes(t *testing.T) {
	want := []string{" @@@ ", "@@@@@", " @@@ "}
	if len(moonMap) != 3 {
		t.Fatalf("moonMap has %d rows, want 3", len(moonMap))
	}
	for i, w := range want {
		if moonMap[i] != w {
			t.Errorf("moonMap row %d: got %q, want %q", i, moonMap[i], w)
		}
	}
}

// A2 — computePositions against hand-computed C values.

func TestComputePositions_MidnightDay0Month0(t *testing.T) {
	sp, mp, mphase, stp := computePositions(0, 0, 0)
	if sp != 72 || mp != 0 || mphase != 0 || stp != 0 {
		t.Errorf("computePositions(0,0,0)=(%d,%d,%d,%d), want (72,0,0,0)", sp, mp, mphase, stp)
	}
}

func TestComputePositions_Noon(t *testing.T) {
	sp, mp, mphase, stp := computePositions(12, 0, 0)
	if sp != 36 || mp != 36 || mphase != 0 || stp != 36 {
		t.Errorf("computePositions(12,0,0)=(%d,%d,%d,%d), want (36,36,0,36)", sp, mp, mphase, stp)
	}
}

// At day=18, hour=0: pre-clamp moonphase = 4 exactly, clamp not triggered (4 is not > 4).
func TestComputePositions_MoonphaseAtUpperBoundary(t *testing.T) {
	sp, mp, mphase, stp := computePositions(0, 18, 0)
	if sp != 72 || mp != 37 || mphase != 4 || stp != 0 {
		t.Errorf("computePositions(0,18,0)=(%d,%d,%d,%d), want (72,37,4,0)", sp, mp, mphase, stp)
	}
}

// At day=20, hour=0: pre-clamp moonphase = 5, clamp fires (-8) → -3.
func TestComputePositions_MoonphaseWanesPastFull(t *testing.T) {
	sp, mp, mphase, stp := computePositions(0, 20, 0)
	if sp != 72 || mp != 41 || mphase != -3 || stp != 0 {
		t.Errorf("computePositions(0,20,0)=(%d,%d,%d,%d), want (72,41,-3,0)", sp, mp, mphase, stp)
	}
}

func TestComputePositions_MonthAdvancesStarpos(t *testing.T) {
	_, _, _, stp := computePositions(0, 0, 8)
	if stp != 33 {
		t.Errorf("starpos at month=8: got %d, want 33", stp)
	}
}

// A3 — precipBucket C-formula parity.

func TestPrecipBucket_DryIsBucketOne(t *testing.T) {
	if got := precipBucket(-19); got != 1 {
		t.Errorf("precipBucket(-19)=%d, want 1 (sky visible)", got)
	}
}

func TestPrecipBucket_BorderlineIsBucketTwo(t *testing.T) {
	if got := precipBucket(0); got != 2 {
		t.Errorf("precipBucket(0)=%d, want 2 (cloudy)", got)
	}
}

func TestPrecipBucket_Rainy(t *testing.T) {
	if got := precipBucket(10); got != 3 {
		t.Errorf("precipBucket(10)=%d, want 3 (cloudy)", got)
	}
}

// Go integer division truncates toward zero (C on gcc likewise). -1/10 == 0.
func TestPrecipBucket_VeryDry(t *testing.T) {
	if got := precipBucket(-30); got != 0 {
		t.Errorf("precipBucket(-30)=%d, want 0 (sky visible)", got)
	}
}

// A4 / A5 — renderStarmap shape and content.

func TestRenderStarmap_CloudyShortCircuit(t *testing.T) {
	lines := renderStarmap(0, 0, 0, 0) // bucket(0)=2, cloudy
	if len(lines) != 2 {
		t.Fatalf("cloudy render has %d lines, want 2: %v", len(lines), lines)
	}
	if lines[0] != "You gaze up towards the heavens and see:" {
		t.Errorf("line[0]=%q, want header", lines[0])
	}
	if !strings.Contains(lines[1], "some clouds in the sky") {
		t.Errorf("line[1]=%q, want cloud message", lines[1])
	}
}

func TestRenderStarmap_HeaderIsFirstLine(t *testing.T) {
	lines := renderStarmap(0, 0, 0, -30)
	if lines[0] != "You gaze up towards the heavens and see:" {
		t.Errorf("header: got %q", lines[0])
	}
}

func TestRenderStarmap_NightRenders8Rows(t *testing.T) {
	lines := renderStarmap(0, 0, 0, -30)
	if len(lines) != 9 {
		t.Errorf("midnight render has %d lines (header + plotted), want 9", len(lines))
	}
}

func TestRenderStarmap_DayRenders3Rows(t *testing.T) {
	lines := renderStarmap(12, 0, 0, -30)
	if len(lines) != 4 {
		t.Errorf("noon render has %d lines (header + plotted), want 4", len(lines))
	}
}

// Daytime boundary per C (starmap.c:116): hour >= 6 && hour <= 18 is inclusive.
func TestRenderStarmap_DayRowSkipBoundary(t *testing.T) {
	cases := []struct {
		hour      int
		wantLines int
		label     string
	}{
		{5, 9, "hour=5 (night)"},
		{6, 4, "hour=6 (first daytime hour)"},
		{18, 4, "hour=18 (last daytime hour)"},
		{19, 9, "hour=19 (night)"},
	}
	for _, tc := range cases {
		lines := renderStarmap(tc.hour, 0, 0, -30)
		if len(lines) != tc.wantLines {
			t.Errorf("%s: got %d lines, want %d", tc.label, len(lines), tc.wantLines)
		}
	}
}

// Night row 0 at starpos=0 must round-trip the starMap row 0 through the
// writeStarGlyph mapping. This is the transcription guard test.
func TestRenderStarmap_NightRow0IsVerbatimTable(t *testing.T) {
	lines := renderStarmap(0, 0, 0, -30)
	// Skip header. Row 0 of the map is lines[1]. Line starts with a leading space
	// from renderStarmap's sprintf(" "); each cell of row 0 then passes through
	// writeStarGlyph.
	got := lines[1]

	// Build expected by running the same mapping over starMap[0] byte-by-byte.
	// Inner loop uses i ∈ [1, 72]; C indexing is (starmapWidth + i - starpos) % starmapWidth
	// which with starpos=0 and i ∈ [1,72] yields byte indices 1..71, then 0 (wrap).
	var expected strings.Builder
	expected.WriteByte(' ')
	for i := 1; i <= starmapWidth; i++ {
		ch := starMap[0][(starmapWidth+i)%starmapWidth]
		writeStarGlyph(&expected, ch)
	}
	want := expected.String()
	if got != want {
		t.Errorf("row 0 render mismatch\n got %q\nwant %q", got, want)
	}

	// Sanity: row should contain some color codes (row 0 has C., C., g*).
	if !strings.Contains(got, "&C ") {
		t.Errorf("expected &C color in row 0, got %q", got)
	}
	if !strings.Contains(got, "&g ") {
		t.Errorf("expected &g color in row 0, got %q", got)
	}
}

// A5 — noon sun at center. day=5 avoids eclipse (moon is offset) so the sun
// glyph renders unblocked on all three day rows.
// Plan v1 used day=0 (eclipse) for this test; that's C-incorrect — moonphase=0
// at the eclipse masks every moon cell as space, which then blocks the sun
// cells behind it. See Completion Record deviation note.
func TestRenderStarmap_NoonSunAtCenter_NonEclipse(t *testing.T) {
	// hour=12 → sunpos=36. day=5 → moonpos=46, moonphase=1. Sun cells 34-38
	// are unobstructed (moonpos-2..+2 = 44..48).
	lines := renderStarmap(12, 5, 0, -30)
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines at noon, got %d", len(lines))
	}
	// Row 3 is lines[1] (lineum 3 is first daytime row). sun_map[0] = "\\`|'/".
	row3 := lines[1]
	// The pipe char at center should render as "&Y|".
	if !strings.Contains(row3, "&Y|") {
		t.Errorf("noon sun row 3 missing &Y|: %q", row3)
	}
	// Row 4 (lines[2]) is sun_map[1] = "- O -". Center char 'O' at position 36
	// renders as "&YO".
	row4 := lines[2]
	if !strings.Contains(row4, "&YO") {
		t.Errorf("noon sun row 4 missing &YO: %q", row4)
	}
}

// Eclipse at noon: moonphase=0 and the eclipse gate admits the moon branch.
// Every moon cell falls through to the phase-mask ` ` branch — the moon
// renders as a 5x3 region of SPACES, blocking the sun behind it.
// Consequence: `&Y|` does NOT appear on row 3 at (hour=12, day=0).
// This is the faithful port of C's `look_sky` and resolves plan A11's
// ambiguity (plan said &W@ appears; C semantics say it does not at moonphase=0).
func TestRenderStarmap_EclipseAtNoonRendersBlackDisk(t *testing.T) {
	lines := renderStarmap(12, 0, 0, -30)
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines at noon eclipse, got %d", len(lines))
	}
	row3 := lines[1]
	// Under the eclipse, the sun's `|` glyph must NOT appear (moon covers it).
	if strings.Contains(row3, "&Y|") {
		t.Errorf("noon-eclipse row 3 should not contain &Y| (moon eclipses sun), got %q", row3)
	}
	// No &W@ glyphs either — at moonphase=0 every moon cell emits space.
	if strings.Contains(row3, "&W@") {
		t.Errorf("noon-eclipse should not render &W@ (moonphase=0 masks all), got %q", row3)
	}
	if strings.Contains(lines[2], "&W@") {
		t.Errorf("noon-eclipse row 4 should not render &W@, got %q", lines[2])
	}
}

// Night new moon (day=0, hour=0): every moon cell of the sky is on the
// invisible/blocked side — the moon is not visible at all. The night branch
// skips the noon-eclipse condition so the entire moon region emits spaces.
func TestRenderStarmap_NewMoonAtNightIsInvisible(t *testing.T) {
	lines := renderStarmap(0, 0, 0, -30)
	if len(lines) != 9 {
		t.Fatalf("expected 9 lines at midnight, got %d", len(lines))
	}
	// Rows 3-5 are moon band; at day=0 moonpos=0, which is LEFT of
	// moonInSky range (moonpos >= 16), so the moon branch isn't entered at all.
	// No `&W@` anywhere in the output.
	for i, ln := range lines {
		if strings.Contains(ln, "&W@") {
			t.Errorf("line %d contains unexpected moon glyph: %q", i, ln)
		}
	}
}

// Waxing moon (positive moonphase) pins the right-edge masking. At day=5 hour=0
// moonpos in band and phase=1 so the right edge is masked.
func TestRenderStarmap_MoonVisibleWhenInSky(t *testing.T) {
	// hour=0 night; day=16 → moonpos = (72 + 16*72/35) % 72 = (72 + 32) % 72 = 32.
	// 32 is in moonInSky range [16, 56]. moonphase = ((72+32-72)%72 + 4)*8/72
	//   = (32+4)*8/72 = 288/72 = 4. Not clamped. Phase=4 (full waxing at edge).
	lines := renderStarmap(0, 16, 0, -30)
	full := strings.Join(lines, "")
	if !strings.Contains(full, "&W@") {
		t.Errorf("day=16 night moon should render &W@ at least once, got %q", full)
	}
}

// -----------------------------------------------------------------------------
// writeStarGlyph color-dispatch table (A5).
// -----------------------------------------------------------------------------

func TestWriteStarGlyph_AllBranches(t *testing.T) {
	cases := []struct {
		in   byte
		want string
	}{
		{':', ":"},
		{'.', "."},
		{'*', "*"},
		{'G', "&G "},
		{'g', "&g "},
		{'R', "&R "},
		{'r', "&r "},
		{'C', "&C "},
		{'O', "&O "},
		{'B', "&B "},
		{'P', "&P "},
		{'W', "&W "},
		{'b', "&b "},
		{'p', "&p "},
		{'Y', "&Y "},
		{'c', "&c "},
		{' ', " "},
		{'x', " "}, // unmapped → space
		{0, " "},
	}
	for _, tc := range cases {
		var buf strings.Builder
		writeStarGlyph(&buf, tc.in)
		if got := buf.String(); got != tc.want {
			t.Errorf("writeStarGlyph(%q)=%q, want %q", tc.in, got, tc.want)
		}
	}
}

// -----------------------------------------------------------------------------
// G2 — DoLook `sky` branch + LookSky dispatch.
// -----------------------------------------------------------------------------

// setupSky constructs an outdoor room with a weather struct and returns
// a char placed in it, a function to read captured output, and a cleanup.
func setupSky(t *testing.T, precip int, indoors bool, sector int) (*types.CharData, func() string, func()) {
	t.Helper()
	w := world.New("/tmp/test")
	WorldRef = w
	w.TimeInfo = types.TimeInfoData{Hour: 0, Day: 0, Month: 0}

	area := &types.AreaData{
		Name: "Test Area",
		Weather: &types.WeatherData{
			Precip: precip,
		},
	}
	room := &types.RoomIndexData{
		Vnum:       8000,
		Name:       "Test Sky Room",
		Area:       area,
		SectorType: sector,
	}
	if indoors {
		room.RoomFlags.Set(types.ROOM_INDOORS)
	}
	w.Rooms[8000] = room

	ch, client := makeTestChar("SkyTester")
	handler.CharToRoom(ch, room)

	readFn := func() string {
		return readOutput(ch, client)
	}
	cleanup := func() { client.Close() }
	return ch, readFn, cleanup
}

// A7 — indoor by ROOM_INDOORS flag.
func TestDoLook_SkyIndoorsByFlag(t *testing.T) {
	ch, read, cleanup := setupSky(t, -30, true, types.SECT_CITY)
	defer cleanup()

	DoLook(ch, "sky")
	out := read()

	if !strings.Contains(out, "can't see the sky indoors") {
		t.Errorf("expected indoor-block message, got %q", out)
	}
	if strings.Contains(out, "gaze up towards") {
		t.Errorf("indoor branch should not emit sky header, got %q", out)
	}
}

// A7 — indoor by SECT_INSIDE sector even without the flag.
func TestDoLook_SkyIndoorsBySector(t *testing.T) {
	ch, read, cleanup := setupSky(t, -30, false, types.SECT_INSIDE)
	defer cleanup()

	DoLook(ch, "sky")
	out := read()

	if !strings.Contains(out, "can't see the sky indoors") {
		t.Errorf("expected indoor-block message, got %q", out)
	}
}

// A6 + A5 — outdoor clear sky produces header + color-coded map.
func TestDoLook_SkyOutdoorsClear(t *testing.T) {
	ch, read, cleanup := setupSky(t, -30, false, types.SECT_FIELD)
	defer cleanup()

	DoLook(ch, "sky")
	out := read()

	if !strings.Contains(out, "You gaze up towards the heavens") {
		t.Errorf("expected sky header, got %q", out)
	}
	// Should contain at least one color code from the constellation table.
	// Row 0 has `C.` which renders `&C `.
	if !strings.Contains(out, "&C ") && !strings.Contains(out, "&g ") && !strings.Contains(out, "&Y ") {
		t.Errorf("expected at least one color code in clear-sky output, got %q", out)
	}
}

// Outdoor cloudy → header + cloud message, no colors.
func TestDoLook_SkyOutdoorsCloudy(t *testing.T) {
	ch, read, cleanup := setupSky(t, 10, false, types.SECT_FIELD)
	defer cleanup()

	DoLook(ch, "sky")
	out := read()

	if !strings.Contains(out, "You gaze up towards the heavens") {
		t.Errorf("cloudy path missing header, got %q", out)
	}
	if !strings.Contains(out, "some clouds in the sky") {
		t.Errorf("cloudy path missing cloud message, got %q", out)
	}
	// Cloud message should be at most header + cloud; no color codes from stars.
	if strings.Contains(out, "&C ") || strings.Contains(out, "&g ") {
		t.Errorf("cloudy output should not contain star colors, got %q", out)
	}
}

// A9 — nil World is a silent no-op, no panic.
func TestDoLook_SkyNoWorldRef(t *testing.T) {
	// Don't call setupSky; LookSky directly with ch present but WorldRef nil.
	prev := WorldRef
	WorldRef = nil
	defer func() { WorldRef = prev }()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("LookSky panicked with nil WorldRef: %v", r)
		}
	}()
	// ch doesn't need to be functional; just non-nil.
	ch := &types.CharData{Name: "X"}
	LookSky(ch)
}

// A10 — nil Area falls back to cloudy; no panic.
func TestDoLook_SkyNilArea(t *testing.T) {
	ch, read, cleanup := setupSky(t, -30, false, types.SECT_FIELD)
	defer cleanup()
	// Wipe the area on the room to force the fallback branch.
	ch.InRoom.Area = nil

	DoLook(ch, "sky")
	out := read()

	// Outdoor branch still fires LookSky; LookSky sees nil Area and uses the
	// sentinel precip → cloudy render.
	if !strings.Contains(out, "some clouds in the sky") {
		t.Errorf("nil-area should produce cloudy fallback, got %q", out)
	}
}

// A8 — dispatch is case-insensitive (`strings.EqualFold`).
func TestDoLook_SkyCaseInsensitive(t *testing.T) {
	for _, cmd := range []string{"SKY", "Sky", "sKy"} {
		ch, read, cleanup := setupSky(t, -30, false, types.SECT_FIELD)
		DoLook(ch, cmd)
		out := read()
		if !strings.Contains(out, "You gaze up towards") {
			t.Errorf("cmd=%q: expected sky header, got %q", cmd, out)
		}
		cleanup()
	}
}

// Regression: generic DoLook paths still work after the sky branch insertion.
func TestDoLook_OtherArgsStillWorkAfterSkyBranch(t *testing.T) {
	// Reuse setupTestWorld from info_test.go.
	w := setupTestWorld()
	ch, client := makeTestChar("Gandalf")
	defer client.Close()

	room := w.GetRoom(21002)
	ch.InRoom = room
	room.People = append(room.People, ch)

	DoLook(ch, "fountain")
	out := readOutput(ch, client)

	if !strings.Contains(out, "marble fountain") {
		t.Errorf("DoLook on keyword should still hit extra-desc path, got %q", out)
	}
}

// -----------------------------------------------------------------------------
// G3 — integration / regression pins.
// -----------------------------------------------------------------------------

// A11 (adjusted) — end-to-end outdoor look at noon on a non-eclipse day
// produces `&Y|` (sun) in the output.
func TestDoLook_SkyContainsSunAtNoon(t *testing.T) {
	ch, read, cleanup := setupSky(t, -30, false, types.SECT_FIELD)
	defer cleanup()
	WorldRef.TimeInfo = types.TimeInfoData{Hour: 12, Day: 5, Month: 0}

	DoLook(ch, "sky")
	out := read()

	if !strings.Contains(out, "&Y|") {
		t.Errorf("noon non-eclipse should contain &Y| sun glyph, got %q", out)
	}
}

// Noon eclipse (day=0) has moonphase=0, which masks the moon as black —
// consequence: NO &W@ anywhere, and the sun glyphs are hidden by the moon's
// shadow spaces.
func TestDoLook_SkyEclipseBlocksSunAt_Day0Noon(t *testing.T) {
	ch, read, cleanup := setupSky(t, -30, false, types.SECT_FIELD)
	defer cleanup()
	WorldRef.TimeInfo = types.TimeInfoData{Hour: 12, Day: 0, Month: 0}

	DoLook(ch, "sky")
	out := read()

	if strings.Contains(out, "&Y|") {
		t.Errorf("noon eclipse should NOT emit &Y| (moon blocks sun), got %q", out)
	}
	if strings.Contains(out, "&W@") {
		t.Errorf("noon eclipse at moonphase=0 should not emit &W@, got %q", out)
	}
}

// Lines split by \n\r should produce exactly 9 segments at midnight
// (header + 8 plot rows, plus a trailing empty from the terminator).
func TestDoLook_SkyPrintsEightNightRows(t *testing.T) {
	ch, read, cleanup := setupSky(t, -30, false, types.SECT_FIELD)
	defer cleanup()
	WorldRef.TimeInfo = types.TimeInfoData{Hour: 0, Day: 0, Month: 0}

	DoLook(ch, "sky")
	out := read()

	// Split on "\n\r"; trailing terminator produces one extra empty segment.
	parts := strings.Split(out, "\n\r")
	// Filter empties.
	nonempty := 0
	for _, p := range parts {
		if p != "" {
			nonempty++
		}
	}
	if nonempty != 9 {
		t.Errorf("midnight sky should emit 9 non-empty \\n\\r-separated lines, got %d: %q", nonempty, out)
	}
}

func TestDoLook_SkyPrintsThreeDayRows(t *testing.T) {
	ch, read, cleanup := setupSky(t, -30, false, types.SECT_FIELD)
	defer cleanup()
	WorldRef.TimeInfo = types.TimeInfoData{Hour: 12, Day: 5, Month: 0}

	DoLook(ch, "sky")
	out := read()

	parts := strings.Split(out, "\n\r")
	nonempty := 0
	for _, p := range parts {
		if p != "" {
			nonempty++
		}
	}
	if nonempty != 4 {
		t.Errorf("noon sky should emit 4 non-empty \\n\\r-separated lines, got %d: %q", nonempty, out)
	}
}

// Month changes the starpos, which rotates the constellation table. The row 0
// bytes must differ between month=0 and month=8 at the same hour/day.
func TestDoLook_SkyCalendarMonthChangesConstellationPosition(t *testing.T) {
	ch1, read1, cleanup1 := setupSky(t, -30, false, types.SECT_FIELD)
	defer cleanup1()
	WorldRef.TimeInfo = types.TimeInfoData{Hour: 0, Day: 0, Month: 0}
	DoLook(ch1, "sky")
	out1 := read1()

	ch2, read2, cleanup2 := setupSky(t, -30, false, types.SECT_FIELD)
	defer cleanup2()
	WorldRef.TimeInfo = types.TimeInfoData{Hour: 0, Day: 0, Month: 8}
	DoLook(ch2, "sky")
	out2 := read2()

	if out1 == out2 {
		t.Errorf("month=0 and month=8 should produce different sky output, both = %q", out1)
	}
}
