package game

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// newTestDesc creates a descriptor with a test character for pager tests.
func newTestDesc(pagerLen int, pagerOn bool) *types.DescriptorData {
	ch := &types.CharData{
		Name:  "Tester",
		Level: 1,
		PCData: &types.PCData{
			PagerLen: pagerLen,
		},
	}
	if pagerOn {
		ch.PCData.Flags |= int(types.PCFLAG_PAGERON)
	}
	d := &types.DescriptorData{
		Character: ch,
		Connected: types.CON_PLAYING,
		ScrLen:    24,
	}
	ch.Desc = d
	return d
}

func TestWriteToPager(t *testing.T) {
	d := newTestDesc(24, true)

	WriteToPager(d, "Hello World\n\r")
	if !d.HasPagerData() {
		t.Fatal("expected pager to have data after WriteToPager")
	}
}

func TestSendToPager_PagerOff(t *testing.T) {
	d := newTestDesc(24, false)
	ch := d.Character

	SendToPager(ch, "Hello World\n\r")

	// With pager off, text should go to normal output buffer
	if d.HasPagerData() {
		t.Error("pager should not have data when pager is off")
	}
	if !d.HasOutput() {
		t.Error("expected normal output when pager is off")
	}
}

func TestSendToPager_PagerOn(t *testing.T) {
	d := newTestDesc(24, true)
	ch := d.Character

	SendToPager(ch, "Hello World\n\r")

	if !d.HasPagerData() {
		t.Error("expected pager to have data when pager is on")
	}
}

func TestSendToPager_NPC(t *testing.T) {
	ch := &types.CharData{
		Name:  "Guard",
		Level: 1,
	}
	ch.Act.Set(types.ACT_IS_NPC)
	// NPCs have no descriptor — SendToPager should not panic
	SendToPager(ch, "test")
}

func TestPagerOutput_ShortText(t *testing.T) {
	// Text shorter than one page should display all at once and clear pager
	d := newTestDesc(24, true)

	WriteToPager(d, "Line one\n\rLine two\n\r")
	SetPagerInput(d, "") // default = continue

	done := PagerOutput(d)
	if !done {
		t.Error("short text should complete in one page")
	}
	if d.HasPagerData() {
		t.Error("pager data should be cleared after complete display")
	}
}

func TestPagerOutput_LongText(t *testing.T) {
	// Generate text longer than one page
	d := newTestDesc(10, true) // 10 lines per page -> 9 content lines

	var sb strings.Builder
	for i := 0; i < 20; i++ {
		sb.WriteString("Line of text\n\r")
	}
	WriteToPager(d, sb.String())
	SetPagerInput(d, "") // continue

	done := PagerOutput(d)
	if done {
		t.Error("20 lines should not fit in 10-line page")
	}

	// Should have pager prompt in output
	// Continue reading to finish
	SetPagerInput(d, "")
	done = PagerOutput(d)
	if done {
		t.Error("should need one more page")
	}

	SetPagerInput(d, "")
	done = PagerOutput(d)
	if !done {
		t.Error("should be done after 3 pages for 20 lines at 9 per page")
	}
}

func TestPagerOutput_Quit(t *testing.T) {
	d := newTestDesc(10, true)

	var sb strings.Builder
	for i := 0; i < 20; i++ {
		sb.WriteString("Line of text\n\r")
	}
	WriteToPager(d, sb.String())

	SetPagerInput(d, "q")
	done := PagerOutput(d)
	if !done {
		t.Error("quit command should end paging immediately")
	}
	if d.HasPagerData() {
		t.Error("pager data should be cleared on quit")
	}
}

func TestPagerOutput_NonStop(t *testing.T) {
	d := newTestDesc(10, true)

	var sb strings.Builder
	for i := 0; i < 20; i++ {
		sb.WriteString("Line of text\n\r")
	}
	WriteToPager(d, sb.String())

	SetPagerInput(d, "n")
	done := PagerOutput(d)
	if !done {
		t.Error("non-stop should display all remaining text")
	}
}

func TestPagerOutput_Back(t *testing.T) {
	d := newTestDesc(10, true)

	var sb strings.Builder
	for i := 0; i < 20; i++ {
		sb.WriteString("Line of text\n\r")
	}
	WriteToPager(d, sb.String())

	// Advance one page
	SetPagerInput(d, "")
	PagerOutput(d)

	// Go back
	SetPagerInput(d, "b")
	done := PagerOutput(d)
	if done {
		t.Error("going back should not end paging")
	}
}

func TestPagerOutput_Refresh(t *testing.T) {
	d := newTestDesc(10, true)

	var sb strings.Builder
	for i := 0; i < 20; i++ {
		sb.WriteString("Line of text\n\r")
	}
	WriteToPager(d, sb.String())

	// Advance one page
	SetPagerInput(d, "")
	PagerOutput(d)

	// Refresh — should re-display the same page
	SetPagerInput(d, "r")
	done := PagerOutput(d)
	if done {
		t.Error("refresh should not end paging")
	}
}

func TestSetPagerInput(t *testing.T) {
	d := newTestDesc(24, true)
	SetPagerInput(d, "q")
	if d.GetPagerCmd() != 'q' {
		t.Errorf("expected pager cmd 'q', got %c", d.GetPagerCmd())
	}
}

func TestPagerOutput_MinPageLen(t *testing.T) {
	// Page length < 5 should be clamped to 5
	d := newTestDesc(2, true)

	var sb strings.Builder
	for i := 0; i < 10; i++ {
		sb.WriteString("Line\n\r")
	}
	WriteToPager(d, sb.String())

	SetPagerInput(d, "")
	done := PagerOutput(d)
	// With min page length 5 (4 content lines), 10 lines needs 3 pages
	if done {
		t.Error("expected paging with min page length")
	}
}
