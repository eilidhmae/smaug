package game

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
)

// pagerPrompt is displayed at the bottom of each page.
const pagerPrompt = "\n\r&P(C)ontinue, (N)on-stop, (R)efresh, (B)ack, (Q)uit: [C] &D"

// SendToPager sends text through the pager if enabled, else to normal output.
// Maps to C send_to_pager().
func SendToPager(ch *types.CharData, text string) {
	if ch == nil || ch.Desc == nil {
		return
	}
	if ch.IsNPC() || ch.PCData == nil || (ch.PCData.Flags&int(types.PCFLAG_PAGERON)) == 0 {
		ch.Send(text)
		return
	}
	WriteToPager(ch.Desc, text)
}

// WriteToPager appends text to the descriptor's pager buffer.
// Maps to C write_to_pager().
func WriteToPager(d *types.DescriptorData, text string) {
	d.WriteToPager(text)
}

// SetPagerInput stores the user's paging command for the next PagerOutput call.
// Maps to C set_pager_input().
func SetPagerInput(d *types.DescriptorData, cmd string) {
	cmd = strings.TrimSpace(cmd)
	if len(cmd) == 0 {
		d.SetPagerCmd(0)
	} else {
		d.SetPagerCmd(strings.ToLower(cmd)[0])
	}
}

// PagerOutput displays the next page of paged output.
// Returns true when all output has been displayed (or user quits).
// Maps to C pager_output().
func PagerOutput(d *types.DescriptorData) bool {
	data := d.GetPagerData()
	if len(data) == 0 {
		d.ClearPager()
		return true
	}

	// Determine page length (minimum 5)
	pageLen := 24
	if d.Character != nil && d.Character.PCData != nil && d.Character.PCData.PagerLen > 0 {
		pageLen = d.Character.PCData.PagerLen
	}
	if pageLen < 5 {
		pageLen = 5
	}
	contentLines := pageLen - 1 // reserve one line for the pager prompt

	pagePoint := d.GetPagePoint()
	cmd := d.GetPagerCmd()

	switch cmd {
	case 'q', 's':
		// Quit paging
		d.ClearPager()
		return true

	case 'n':
		// Non-stop: dump everything remaining
		if pagePoint < len(data) {
			d.WriteToBuffer(data[pagePoint:])
		}
		d.ClearPager()
		return true

	case 'b', 'v':
		// Back: move pagePoint back by 2 pages worth of lines
		pagePoint = goBackLines(data, pagePoint, contentLines*2)

	case 'r':
		// Refresh: move pagePoint back by 1 page worth of lines
		pagePoint = goBackLines(data, pagePoint, contentLines)
	}

	// Display contentLines lines from pagePoint
	pos := pagePoint
	linesShown := 0
	for pos < len(data) && linesShown < contentLines {
		// Find next line break (\n)
		nlIdx := strings.Index(data[pos:], "\n")
		if nlIdx == -1 {
			// No more newlines — output the rest as the last line
			d.WriteToBuffer(data[pos:])
			pos = len(data)
			linesShown++
		} else {
			end := pos + nlIdx + 1
			// Include \r if present after \n... but SMAUG uses \n\r
			if end < len(data) && data[end] == '\r' {
				end++
			}
			d.WriteToBuffer(data[pos:end])
			pos = end
			linesShown++
		}
	}

	d.SetPagePoint(pos)

	if pos >= len(data) {
		// All content displayed
		d.ClearPager()
		return true
	}

	// Show pager prompt
	d.WriteToBuffer(pagerPrompt)
	return false
}

// goBackLines moves backward in the text by the specified number of lines
// from the given position. Returns the new position.
func goBackLines(data string, pos int, lines int) int {
	if pos <= 0 {
		return 0
	}
	// Start just before current position
	p := pos - 1
	count := 0
	for p > 0 && count < lines {
		p--
		if data[p] == '\n' {
			count++
		}
	}
	// If we hit the beginning, start from 0
	if p <= 0 {
		return 0
	}
	// Advance past the \n we stopped on
	p++
	// Skip \r after \n
	if p < len(data) && data[p] == '\r' {
		p++
	}
	return p
}
