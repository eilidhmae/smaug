package game

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// newEditorDesc creates a descriptor with a character for editor tests.
func newEditorDesc() (*types.DescriptorData, *types.CharData) {
	ch := &types.CharData{
		Name:  "Builder",
		Level: types.LEVEL_IMMORTAL,
		PCData: &types.PCData{
			PagerLen: 24,
		},
	}
	d := &types.DescriptorData{
		Character: ch,
		Connected: types.CON_PLAYING,
	}
	ch.Desc = d
	return d, ch
}

func TestStartEditing_Empty(t *testing.T) {
	d, ch := newEditorDesc()

	StartEditing(ch, "")

	if ch.Editor == nil {
		t.Fatal("editor should be initialized")
	}
	if d.Connected != types.CON_EDITING {
		t.Errorf("expected CON_EDITING, got %d", d.Connected)
	}
	if ch.Editor.NumLines != 0 {
		t.Errorf("expected 0 lines, got %d", ch.Editor.NumLines)
	}
}

func TestStartEditing_WithExistingText(t *testing.T) {
	_, ch := newEditorDesc()

	StartEditing(ch, "Line one\n\rLine two\n\r")

	if ch.Editor == nil {
		t.Fatal("editor should be initialized")
	}
	if ch.Editor.NumLines != 2 {
		t.Errorf("expected 2 lines, got %d", ch.Editor.NumLines)
	}
	if ch.Editor.Lines[0] != "Line one" {
		t.Errorf("expected 'Line one', got %q", ch.Editor.Lines[0])
	}
	if ch.Editor.Lines[1] != "Line two" {
		t.Errorf("expected 'Line two', got %q", ch.Editor.Lines[1])
	}
}

func TestStopEditing(t *testing.T) {
	d, ch := newEditorDesc()

	StartEditing(ch, "test")
	StopEditing(ch)

	if ch.Editor != nil {
		t.Error("editor should be nil after StopEditing")
	}
	if d.Connected != types.CON_PLAYING {
		t.Errorf("expected CON_PLAYING, got %d", d.Connected)
	}
	if ch.Substate != types.SUB_NONE {
		t.Errorf("expected SUB_NONE, got %d", ch.Substate)
	}
}

func TestCopyBuffer_Empty(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "")

	result := CopyBuffer(ch)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestCopyBuffer_WithContent(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "Line one\n\rLine two\n\r")

	result := CopyBuffer(ch)
	if result != "Line one\n\rLine two\n\r" {
		t.Errorf("expected 'Line one\\n\\rLine two\\n\\r', got %q", result)
	}
}

func TestEditBuffer_AppendLine(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "")

	EditBuffer(ch, "Hello world")

	if ch.Editor.NumLines != 1 {
		t.Errorf("expected 1 line, got %d", ch.Editor.NumLines)
	}
	if ch.Editor.Lines[0] != "Hello world" {
		t.Errorf("expected 'Hello world', got %q", ch.Editor.Lines[0])
	}
}

func TestEditBuffer_ListCommand(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "Line one\n\rLine two\n\r")

	EditBuffer(ch, "/l")

	// Verify it doesn't crash — output goes to descriptor buffer
}

func TestEditBuffer_ClearCommand(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "Line one\n\rLine two\n\r")

	EditBuffer(ch, "/c")

	if ch.Editor.NumLines != 0 {
		t.Errorf("expected 0 lines after clear, got %d", ch.Editor.NumLines)
	}
	if ch.Editor.Size != 0 {
		t.Errorf("expected 0 size after clear, got %d", ch.Editor.Size)
	}
}

func TestEditBuffer_DeleteLine(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "Line one\n\rLine two\n\rLine three\n\r")

	EditBuffer(ch, "/d 2")

	if ch.Editor.NumLines != 2 {
		t.Errorf("expected 2 lines after delete, got %d", ch.Editor.NumLines)
	}
	if ch.Editor.Lines[0] != "Line one" {
		t.Errorf("expected 'Line one', got %q", ch.Editor.Lines[0])
	}
	if ch.Editor.Lines[1] != "Line three" {
		t.Errorf("expected 'Line three', got %q", ch.Editor.Lines[1])
	}
}

func TestEditBuffer_DeleteCurrentLine(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "Line one\n\rLine two\n\r")

	// Current line should be after last line = line 3 (1-indexed), but clamped
	EditBuffer(ch, "/d")

	// Should delete the last line (current = numlines)
	if ch.Editor.NumLines != 1 {
		t.Errorf("expected 1 line after deleting current, got %d", ch.Editor.NumLines)
	}
}

func TestEditBuffer_GotoLine(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "Line one\n\rLine two\n\rLine three\n\r")

	EditBuffer(ch, "/g 2")

	if ch.Editor.OnLine != 1 { // 0-indexed
		t.Errorf("expected OnLine=1 (line 2), got %d", ch.Editor.OnLine)
	}
}

func TestEditBuffer_InsertLine(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "Line one\n\rLine three\n\r")

	EditBuffer(ch, "/g 2")
	EditBuffer(ch, "/i")
	EditBuffer(ch, "Line two")

	// After insert at line 2, "Line two" should replace the blank line
	result := CopyBuffer(ch)
	if !strings.Contains(result, "Line one") || !strings.Contains(result, "Line three") {
		t.Errorf("unexpected result: %q", result)
	}
}

func TestEditBuffer_Replace(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "The quick brown fox\n\rThe quick brown cat\n\r")

	EditBuffer(ch, "/r quick slow")

	if ch.Editor.Lines[0] != "The slow brown fox" {
		t.Errorf("expected 'The slow brown fox', got %q", ch.Editor.Lines[0])
	}
	if ch.Editor.Lines[1] != "The slow brown cat" {
		t.Errorf("expected 'The slow brown cat', got %q", ch.Editor.Lines[1])
	}
}

func TestEditBuffer_Format(t *testing.T) {
	_, ch := newEditorDesc()
	// Create one long line that should be wrapped
	long := strings.Repeat("word ", 20) // 100 chars
	StartEditing(ch, long+"\n\r")

	EditBuffer(ch, "/f")

	// After formatting, text should be wrapped to ~75 chars
	result := CopyBuffer(ch)
	lines := strings.Split(result, "\n\r")
	for _, line := range lines {
		if line == "" {
			continue
		}
		if len(line) > 79 {
			t.Errorf("line too long after format: %d chars", len(line))
		}
	}
}

func TestEditBuffer_Abort(t *testing.T) {
	d, ch := newEditorDesc()
	ch.Substate = types.SUB_ROOM_DESC
	StartEditing(ch, "original text")

	EditBuffer(ch, "/a")

	if ch.Editor != nil {
		t.Error("editor should be nil after abort")
	}
	if d.Connected != types.CON_PLAYING {
		t.Errorf("expected CON_PLAYING after abort, got %d", d.Connected)
	}
}

func TestEditBuffer_Help(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "")

	// Just verify /? doesn't crash
	EditBuffer(ch, "/?")
}

func TestEditBuffer_LineTooLong(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "")

	long := strings.Repeat("x", 100)
	EditBuffer(ch, long)

	if ch.Editor.NumLines != 1 {
		t.Fatalf("expected 1 line, got %d", ch.Editor.NumLines)
	}
	if len(ch.Editor.Lines[0]) > 79 {
		t.Errorf("line should be truncated to 79 chars, got %d", len(ch.Editor.Lines[0]))
	}
}

func TestEditBuffer_MaxLines(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "")

	// Fill to max (49 lines for standard, 24 for normal editing)
	for i := 0; i < 25; i++ {
		EditBuffer(ch, "a line")
	}

	// Should be capped at max_buf_lines (24)
	if ch.Editor.NumLines > 24 {
		t.Errorf("expected max 24 lines, got %d", ch.Editor.NumLines)
	}
}

func TestStartEditing_ParsesNewlineVariants(t *testing.T) {
	_, ch := newEditorDesc()

	// Test with just \n (no \r)
	StartEditing(ch, "Line one\nLine two\n")

	if ch.Editor.NumLines != 2 {
		t.Errorf("expected 2 lines, got %d", ch.Editor.NumLines)
	}
	if ch.Editor.Lines[0] != "Line one" {
		t.Errorf("expected 'Line one', got %q", ch.Editor.Lines[0])
	}
}
