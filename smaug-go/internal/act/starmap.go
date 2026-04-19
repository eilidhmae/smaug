package act

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
)

// Constellation, sun, and moon data for the `look sky` command.
// Ports C src/starmap.c:50-85 + look_sky at src/starmap.c:87-226.
//
// Constants preserved from C compile-time defines at src/starmap.c:50-55.
// They intentionally diverge from the runtime calendar (world.Calendar.DaysPerMonth /
// MonthsPerYear may differ) because the constellation table is hard-sized to them.
// C has the same divergence.
const (
	starmapWidth     = 72
	starmapHeight    = 8
	starmapNumDays   = 35 // moon cycle period; matches C NUM_DAYS
	starmapNumMonths = 17 // matches C NUM_MONTHS
	starmapWeathUnit = 10 // matches C weath_unit (src/db.c:108,600); package-private const.
)

// starMap is the 8x72 constellation table, byte-for-byte from src/starmap.c:59-68.
// Each row MUST be exactly 72 bytes. Transcription errors here are the single
// highest-risk class of bug in this port — pin-tested per-row.
var starMap = []string{
	"                                               C. C.                  g*",
	"    O:       R*        G*    G.  W* W. W.          C. C.    Y* Y. Y.    ",
	"  O*.                c.          W.W.     W.            C.       Y..Y.  ",
	"O.O. O.              c.  G..G.           W:      B*                   Y.",
	"     O.    c.     c.                     W. W.                  r*    Y.",
	"     O.c.     c.      G.             P..     W.        p.      Y.   Y:  ",
	"        c.                    G*    P.  P.           p.  p:     Y.   Y. ",
	"                 b*             P.: P*                 p.p:             ",
}

// sunMap is the 3x5 sun glyph, byte-for-byte from src/starmap.c:75-79.
// C literals use "\\`|'/" (backslash, backtick, pipe, apostrophe, slash) and
// "/.|.\\" (slash, period, pipe, period, backslash). Raw-string Go literals
// let us render them without escape ambiguity.
var sunMap = []string{
	`\` + "`" + `|'/`,
	`- O -`,
	`/.|.\`,
}

// moonMap is the 3x5 moon glyph, byte-for-byte from src/starmap.c:81-85.
var moonMap = []string{
	" @@@ ",
	"@@@@@",
	" @@@ ",
}

// computePositions ports the position math at src/starmap.c:105-112.
// Returns (sunpos, moonpos, moonphase, starpos).
//
// sunpos   ∈ [0, 72]         — at hour 0 → 72, hour 12 → 36, hour 24 would be 0 (unreachable).
// moonpos  ∈ [0, 71]         — (sunpos + day*72/35) mod 72.
// moonphase ∈ [-3, 4]         — negative = waning, 0 = new/full-at-noon-aligned, positive = waxing.
// starpos  ∈ [0, 71]         — (sunpos + month*72/17) mod 72.
func computePositions(hour, day, month int) (sunpos, moonpos, moonphase, starpos int) {
	sunpos = starmapWidth * (24 - hour) / 24
	moonpos = (sunpos + day*starmapWidth/starmapNumDays) % starmapWidth
	moonphase = ((((starmapWidth + moonpos - sunpos) % starmapWidth) + (starmapWidth / 16)) * 8) / starmapWidth
	if moonphase > 4 {
		moonphase -= 8
	}
	starpos = (sunpos + starmapWidth*month/starmapNumMonths) % starmapWidth
	return
}

// precipBucket ports the ceiling-divide precip formula at src/starmap.c:96-97:
//
//	(raw + 3*weath_unit - 1) / weath_unit
//
// Bucket ≤ 1 means "sky visible"; bucket > 1 triggers the cloudy short-circuit.
// Raw precip is bounded [-30, +30] in the Go port per the URANGE at
// src/update.c:3446-3452; the formula is well-behaved over that domain.
//
// Go integer division truncates toward zero, matching C semantics on all
// 64-bit linux/macos targets this project supports.
func precipBucket(rawPrecip int) int {
	return (rawPrecip + 3*starmapWeathUnit - 1) / starmapWeathUnit
}

// renderStarmap produces the rendered sky output, one line per slice element.
// Returned lines do NOT have trailing "\n\r" — LookSky appends that per line.
// The first element is always the header "You gaze up towards the heavens and see:".
//
// If the area precip is cloudy (bucket > 1), returns exactly two elements:
// header + cloud line. Otherwise returns header + 3 lines (daytime, hour in
// 6..18 inclusive) or header + 8 lines (nighttime).
//
// Ports C look_sky at src/starmap.c:87-226. Individual-int signature (rather
// than a full TimeInfoData) keeps the function test-friendly and mirrors C's
// locals. See plan Open Q3.
func renderStarmap(hour, day, month, rawPrecip int) []string {
	out := []string{"You gaze up towards the heavens and see:"}
	if precipBucket(rawPrecip) > 1 {
		out = append(out, "There are some clouds in the sky so you cannot see anything else.")
		return out
	}

	sunpos, moonpos, moonphase, starpos := computePositions(hour, day, month)
	daytime := hour >= 6 && hour <= 18

	for linenum := 0; linenum < starmapHeight; linenum++ {
		// During daytime, only rows 3, 4, 5 render (sun/moon rows).
		// At night, all 8 rows render.
		if daytime && (linenum < 3 || linenum >= 6) {
			continue
		}
		var buf strings.Builder
		// C `sprintf(buf, " ")` at starmap.c:119 — one leading space on every rendered line.
		buf.WriteByte(' ')
		// Inner loop i ∈ [1, 72] — C is 1-indexed at starmap.c:123.
		for i := 1; i <= starmapWidth; i++ {
			writeCell(&buf, linenum, i, hour, sunpos, moonpos, moonphase, starpos, daytime)
		}
		out = append(out, buf.String())
	}
	return out
}

// writeCell plots one (linenum, i) cell to buf. Ports the per-cell branch
// at src/starmap.c:125-220.
func writeCell(buf *strings.Builder, linenum, i, hour, sunpos, moonpos, moonphase, starpos int, daytime bool) {
	// Common moon visibility predicates (starmap.c:127-128, :139-140).
	moonInSky := moonpos >= starmapWidth/4-2 && moonpos <= 3*starmapWidth/4+2
	nearMoon := i >= moonpos-2 && i <= moonpos+2
	// In moon row band (rows 3-5) — same as (linenum - 3) is a valid moon_map index.
	inMoonRow := linenum >= 3 && linenum < 6

	// Daytime moon branch — starmap.c:126-137.
	// "Plot moon on top of anything else ...unless new moon & no eclipse".
	// Eclipse at noon: sunpos == moonpos && hour == 12; at that moment moonphase == 0
	// and the moon IS drawn. For any other new-moon (moonphase == 0 & not eclipse)
	// the branch is skipped.
	if daytime && inMoonRow && moonInSky && nearMoon &&
		((sunpos == moonpos && hour == 12) || moonphase != 0) &&
		moonMap[linenum-3][i+2-moonpos] == '@' {
		writeMoonGlyph(buf, i, moonpos, moonphase)
		return
	}

	// Night moon branch — starmap.c:138-148.
	// At night there is no sun to eclipse, so no noon check.
	if !daytime && inMoonRow && moonInSky && nearMoon &&
		moonMap[linenum-3][i+2-moonpos] == '@' {
		writeMoonGlyph(buf, i, moonpos, moonphase)
		return
	}

	// Daytime sun branch — starmap.c:149-161.
	if daytime {
		if i >= sunpos-2 && i <= sunpos+2 {
			// sun_map indexed by (linenum-3, i+2-sunpos). inMoonRow guarantees
			// linenum-3 in [0,2]; the i-range guard gives i+2-sunpos in [0,4].
			buf.WriteString("&Y")
			buf.WriteByte(sunMap[linenum-3][i+2-sunpos])
			return
		}
		buf.WriteByte(' ')
		return
	}

	// Night star branch — starmap.c:162-219.
	ch := starMap[linenum][(starmapWidth+i-starpos)%starmapWidth]
	writeStarGlyph(buf, ch)
}

// writeMoonGlyph handles the moon-phase masking at starmap.c:132-136 and :143-147.
// Waning (moonphase < 0) masks the trailing edge; waxing (moonphase > 0) masks
// the leading edge. A new moon (moonphase == 0) masks everything — every cell
// falls through to the space branch at night; daytime only reaches this path
// during the noon eclipse (when the gate permits despite moonphase == 0).
func writeMoonGlyph(buf *strings.Builder, i, moonpos, moonphase int) {
	if (moonphase < 0 && i-2-moonpos >= moonphase) ||
		(moonphase > 0 && i+2-moonpos <= moonphase) {
		buf.WriteString("&W@")
		return
	}
	// new-moon-at-eclipse (daytime) or masked side (night) → blank space.
	buf.WriteByte(' ')
}

// writeStarGlyph dispatches on a single char from starMap. Ports the switch
// at src/starmap.c:166-218. 13 color letters map to `&X ` (space); the 3
// plain glyphs `.` `:` `*` pass through verbatim; default is space.
func writeStarGlyph(buf *strings.Builder, ch byte) {
	switch ch {
	case ':':
		buf.WriteByte(':')
	case '.':
		buf.WriteByte('.')
	case '*':
		buf.WriteByte('*')
	case 'G':
		buf.WriteString("&G ")
	case 'g':
		buf.WriteString("&g ")
	case 'R':
		buf.WriteString("&R ")
	case 'r':
		buf.WriteString("&r ")
	case 'C':
		buf.WriteString("&C ")
	case 'O':
		buf.WriteString("&O ")
	case 'B':
		buf.WriteString("&B ")
	case 'P':
		buf.WriteString("&P ")
	case 'W':
		buf.WriteString("&W ")
	case 'b':
		buf.WriteString("&b ")
	case 'p':
		buf.WriteString("&p ")
	case 'Y':
		buf.WriteString("&Y ")
	case 'c':
		buf.WriteString("&c ")
	default:
		buf.WriteByte(' ')
	}
}

// LookSky renders the constellation and sun/moon map to ch.
// Port of C src/starmap.c:87-226 (look_sky).
//
// Caller is responsible for the IS_OUTSIDE gate; LookSky assumes ch is outdoors.
// If WorldRef or ch is nil, LookSky returns silently — caller invariants should
// prevent this but the guard avoids panics during tests and dev-loop crashes.
// If ch.InRoom / .Area / .Weather is nil, LookSky falls back to the cloudy
// rendering (header + cloud line) rather than panicking.
func LookSky(ch *types.CharData) {
	if ch == nil || WorldRef == nil {
		return
	}
	precip := 3 * starmapWeathUnit // sentinel triggering the cloudy short-circuit
	if ch.InRoom != nil && ch.InRoom.Area != nil && ch.InRoom.Area.Weather != nil {
		precip = ch.InRoom.Area.Weather.Precip
	}
	t := WorldRef.TimeInfo
	for _, line := range renderStarmap(t.Hour, t.Day, t.Month, precip) {
		ch.Send(line + "\n\r")
	}
}
