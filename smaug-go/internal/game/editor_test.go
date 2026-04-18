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

func TestEditBuffer_DeleteInvalidLine(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "Line one\n\rLine two\n\r")

	// Delete line 0 (invalid, 1-indexed)
	EditBuffer(ch, "/d 0")
	if ch.Editor.NumLines != 2 {
		t.Errorf("expected 2 lines (invalid delete), got %d", ch.Editor.NumLines)
	}

	// Delete line 99 (out of range)
	EditBuffer(ch, "/d 99")
	if ch.Editor.NumLines != 2 {
		t.Errorf("expected 2 lines (out of range delete), got %d", ch.Editor.NumLines)
	}

	// Delete with non-numeric arg
	EditBuffer(ch, "/d abc")
	if ch.Editor.NumLines != 2 {
		t.Errorf("expected 2 lines (non-numeric delete), got %d", ch.Editor.NumLines)
	}
}

func TestEditBuffer_DeleteEmptyBuffer(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "")

	// Deleting from empty buffer
	EditBuffer(ch, "/d")
	if ch.Editor.NumLines != 0 {
		t.Errorf("expected 0 lines after deleting from empty, got %d", ch.Editor.NumLines)
	}
}

func TestEditBuffer_GotoEmptyArg(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "Line one\n\r")

	// Goto with no args should show error
	EditBuffer(ch, "/g")
	// OnLine should not change (was set to numlines=1 by StartEditing)
	if ch.Editor.OnLine != 1 {
		t.Errorf("expected OnLine=1 (unchanged), got %d", ch.Editor.OnLine)
	}
}

func TestEditBuffer_GotoInvalidLine(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "Line one\n\rLine two\n\r")

	// Goto 0 is invalid (1-indexed, max numlines+1)
	EditBuffer(ch, "/g 0")
	if ch.Editor.OnLine != 2 {
		t.Errorf("expected OnLine=2 (unchanged), got %d", ch.Editor.OnLine)
	}

	// Goto beyond range
	EditBuffer(ch, "/g 99")
	if ch.Editor.OnLine != 2 {
		t.Errorf("expected OnLine=2 (unchanged), got %d", ch.Editor.OnLine)
	}

	// Non-numeric
	EditBuffer(ch, "/g abc")
	if ch.Editor.OnLine != 2 {
		t.Errorf("expected OnLine=2 (unchanged), got %d", ch.Editor.OnLine)
	}
}

func TestEditBuffer_InsertWithSpecificLine(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "Line one\n\rLine three\n\r")

	// Insert at line 2 (between line one and three)
	EditBuffer(ch, "/i 2")

	if ch.Editor.NumLines != 3 {
		t.Fatalf("expected 3 lines after insert, got %d", ch.Editor.NumLines)
	}
	if ch.Editor.Lines[0] != "Line one" {
		t.Errorf("line 1 = %q, want 'Line one'", ch.Editor.Lines[0])
	}
	if ch.Editor.Lines[1] != "" {
		t.Errorf("line 2 = %q, want empty (inserted)", ch.Editor.Lines[1])
	}
	if ch.Editor.Lines[2] != "Line three" {
		t.Errorf("line 3 = %q, want 'Line three'", ch.Editor.Lines[2])
	}
}

func TestEditBuffer_InsertInvalidLine(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "Line one\n\r")

	EditBuffer(ch, "/i 0")
	if ch.Editor.NumLines != 1 {
		t.Errorf("expected 1 line (invalid insert), got %d", ch.Editor.NumLines)
	}

	EditBuffer(ch, "/i 99")
	if ch.Editor.NumLines != 1 {
		t.Errorf("expected 1 line (out of range insert), got %d", ch.Editor.NumLines)
	}
}

func TestEditBuffer_InsertBufferFull(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "")

	// Fill buffer to max
	for i := 0; i < 24; i++ {
		EditBuffer(ch, "a line")
	}

	// Insert should fail when full
	EditBuffer(ch, "/i 1")
	if ch.Editor.NumLines != 24 {
		t.Errorf("expected 24 lines (insert on full), got %d", ch.Editor.NumLines)
	}
}

func TestEditBuffer_ReplaceNoArgs(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "Hello world\n\r")

	// Not enough args
	EditBuffer(ch, "/r hello")
	// Should not crash, line unchanged
	if ch.Editor.Lines[0] != "Hello world" {
		t.Errorf("line should be unchanged, got %q", ch.Editor.Lines[0])
	}
}

func TestEditBuffer_ReplaceNoMatch(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "Hello world\n\r")

	EditBuffer(ch, "/r xyz abc")
	if ch.Editor.Lines[0] != "Hello world" {
		t.Errorf("line should be unchanged, got %q", ch.Editor.Lines[0])
	}
}

func TestEditBuffer_FormatEmptyBuffer(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "")

	// Format empty buffer should not crash
	EditBuffer(ch, "/f")
	if ch.Editor.NumLines != 0 {
		t.Errorf("expected 0 lines after formatting empty, got %d", ch.Editor.NumLines)
	}
}

func TestEditBuffer_FormatMultipleLines(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "short\n\rwords\n\r")

	EditBuffer(ch, "/f")

	// Should merge into a single line since combined text is short
	result := CopyBuffer(ch)
	if !strings.Contains(result, "short words") {
		t.Errorf("format should merge short lines, got %q", result)
	}
}

func TestEditBuffer_UnknownCommand(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "")

	// Unknown command should not crash
	EditBuffer(ch, "/z")
	if ch.Editor.NumLines != 0 {
		t.Errorf("expected 0 lines after unknown command, got %d", ch.Editor.NumLines)
	}
}

func TestEditBuffer_SaveCommand(t *testing.T) {
	// /s must transition the descriptor to CON_PLAYING and invoke the
	// caller's save callback. A bare /s with no callback still transitions.
	d, ch := newEditorDesc()
	StartEditing(ch, "Some text\n\r")
	invocations := 0
	var received *types.CharData
	ch.EditorSave = func(c *types.CharData) {
		invocations++
		received = c
		StopEditing(c)
	}

	EditBuffer(ch, "/s")

	if d.Connected != types.CON_PLAYING {
		t.Errorf("expected CON_PLAYING after /s, got %d", d.Connected)
	}
	if invocations != 1 {
		t.Errorf("expected callback invoked exactly once, got %d", invocations)
	}
	if received != ch {
		t.Errorf("callback received wrong char: %p want %p", received, ch)
	}
	if ch.EditorSave != nil {
		t.Error("EditorSave must be cleared before invocation (one-shot)")
	}
	if ch.Editor != nil {
		t.Error("callback called StopEditing; Editor should be nil")
	}
}

func TestEditBuffer_BackslashCommand(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "")

	// Backslash should also work as command prefix
	EditBuffer(ch, "\\l")
	// Should not crash (lists empty buffer)
}

func TestEditBuffer_ListEmptyBuffer(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "")

	// List empty buffer should output "Buffer is empty."
	EditBuffer(ch, "/l")
	// Verify no crash
}

func TestEditBuffer_MprogEditMaxLines(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "")
	ch.Substate = types.SUB_MPROG_EDIT

	// Mudprog edit allows 48 lines
	for i := 0; i < 48; i++ {
		EditBuffer(ch, "a line")
	}

	if ch.Editor.NumLines != 48 {
		t.Errorf("expected 48 lines for mprog edit, got %d", ch.Editor.NumLines)
	}

	// 49th should fail
	EditBuffer(ch, "one more")
	if ch.Editor.NumLines != 48 {
		t.Errorf("expected 48 lines (capped), got %d", ch.Editor.NumLines)
	}
}

func TestStopEditing_NilChar(t *testing.T) {
	// Should not panic
	StopEditing(nil)
}

func TestStopEditing_NilDesc(t *testing.T) {
	ch := &types.CharData{
		Editor: &types.EditorData{},
	}
	// Should not panic with nil Desc
	StopEditing(ch)
	if ch.Editor != nil {
		t.Error("editor should be nil after StopEditing")
	}
}

func TestCopyBuffer_NilChar(t *testing.T) {
	result := CopyBuffer(nil)
	if result != "" {
		t.Errorf("CopyBuffer(nil) = %q, want empty", result)
	}
}

func TestCopyBuffer_NilEditor(t *testing.T) {
	ch := &types.CharData{}
	result := CopyBuffer(ch)
	if result != "" {
		t.Errorf("CopyBuffer(no editor) = %q, want empty", result)
	}
}

func TestEditBuffer_NilChar(t *testing.T) {
	// Should not panic
	EditBuffer(nil, "test")
}

func TestEditBuffer_NilEditor(t *testing.T) {
	ch := &types.CharData{}
	// Should not panic
	EditBuffer(ch, "test")
}

func TestSendToPager_NilChar(t *testing.T) {
	// Should not panic
	SendToPager(nil, "test")
}

func TestStartEditing_NilChar(t *testing.T) {
	// Should not panic
	StartEditing(nil, "test")
}

func TestStartEditing_NilDesc(t *testing.T) {
	ch := &types.CharData{}
	// Should not panic
	StartEditing(ch, "test")
	if ch.Editor != nil {
		t.Error("editor should not be set with nil desc")
	}
}

func TestStartEditing_LongLines(t *testing.T) {
	_, ch := newEditorDesc()
	long := strings.Repeat("x", 100)
	StartEditing(ch, long+"\n\r")

	if ch.Editor.NumLines != 1 {
		t.Fatalf("expected 1 line, got %d", ch.Editor.NumLines)
	}
	if len(ch.Editor.Lines[0]) > 78 {
		t.Errorf("line should be truncated to 78 chars, got %d", len(ch.Editor.Lines[0]))
	}
}

func TestEditBuffer_DeleteLastLineUpdatesOnLine(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "Line one\n\rLine two\n\rLine three\n\r")

	// Goto last line
	EditBuffer(ch, "/g 3")
	if ch.Editor.OnLine != 2 {
		t.Fatalf("expected OnLine=2 after /g 3, got %d", ch.Editor.OnLine)
	}

	// Delete last line
	EditBuffer(ch, "/d 3")
	if ch.Editor.OnLine > ch.Editor.NumLines {
		t.Errorf("OnLine=%d should not exceed NumLines=%d after delete",
			ch.Editor.OnLine, ch.Editor.NumLines)
	}
}

func TestEditBuffer_DeleteUpdatesSize(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "Hello\n\rWorld\n\r")
	origSize := ch.Editor.Size

	EditBuffer(ch, "/d 1") // Delete "Hello"
	if ch.Editor.Size != origSize-5 {
		t.Errorf("size should be %d after deleting 'Hello', got %d", origSize-5, ch.Editor.Size)
	}
}

// --- /s callback and transition ---

// TestEditBuffer_SaveInvokesCallbackAndTransitions asserts that /s transitions
// the descriptor back to CON_PLAYING and invokes the caller's save callback
// exactly once with the correct char.
func TestEditBuffer_SaveInvokesCallbackAndTransitions(t *testing.T) {
	d, ch := newEditorDesc()
	StartEditing(ch, "")
	EditBuffer(ch, "Line A")

	invocations := 0
	var received *types.CharData
	ch.EditorSave = func(c *types.CharData) {
		invocations++
		received = c
	}

	EditBuffer(ch, "/s")

	if d.Connected != types.CON_PLAYING {
		t.Errorf("expected CON_PLAYING, got %d", d.Connected)
	}
	if invocations != 1 {
		t.Errorf("callback should be invoked exactly once, got %d", invocations)
	}
	if received != ch {
		t.Errorf("callback received wrong char: got %p want %p", received, ch)
	}
}

// TestEditBuffer_SaveCallsCallback_ClearsState verifies a realistic save flow:
// callback grabs buffer text, calls StopEditing, and all editor state is clean.
func TestEditBuffer_SaveCallsCallback_ClearsState(t *testing.T) {
	d, ch := newEditorDesc()
	StartEditing(ch, "")
	EditBuffer(ch, "First line")
	EditBuffer(ch, "Second line")

	var saved string
	ch.EditorSave = func(c *types.CharData) {
		saved = CopyBuffer(c)
		StopEditing(c)
	}
	EditBuffer(ch, "/s")

	if ch.Editor != nil {
		t.Error("Editor should be nil after StopEditing")
	}
	if ch.Substate != types.SUB_NONE {
		t.Errorf("Substate should be SUB_NONE, got %d", ch.Substate)
	}
	if d.Connected != types.CON_PLAYING {
		t.Errorf("expected CON_PLAYING, got %d", d.Connected)
	}
	if !strings.Contains(saved, "First line") || !strings.Contains(saved, "Second line") {
		t.Errorf("expected both lines captured, got %q", saved)
	}
}

// TestEditBuffer_SaveNoCallback_StillTransitions protects the no-callback
// edge case: /s must still transition CON_PLAYING and must not panic.
func TestEditBuffer_SaveNoCallback_StillTransitions(t *testing.T) {
	d, ch := newEditorDesc()
	StartEditing(ch, "")
	// No EditorSave assigned.

	EditBuffer(ch, "/s") // must not panic

	if d.Connected != types.CON_PLAYING {
		t.Errorf("expected CON_PLAYING even with no callback, got %d", d.Connected)
	}
}

// TestEditBuffer_AbortDoesNotInvokeCallback verifies /a discards pending save.
func TestEditBuffer_AbortDoesNotInvokeCallback(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "")

	invocations := 0
	ch.EditorSave = func(c *types.CharData) {
		invocations++
	}

	EditBuffer(ch, "/a")

	if invocations != 0 {
		t.Errorf("/a must NOT invoke save callback, got %d invocations", invocations)
	}
}

// TestEditBuffer_SaveDoubleFireUsesOneShotClear protects against a buggy
// re-invoke: if the caller's save fails to clear EditorSave and the user
// somehow re-enters the editor, a second /s must not fire the SAME original
// callback again. Enforced by the one-shot clear in the /s handler.
func TestEditBuffer_SaveDoubleFireUsesOneShotClear(t *testing.T) {
	_, ch := newEditorDesc()
	StartEditing(ch, "")

	invocations := 0
	ch.EditorSave = func(c *types.CharData) {
		invocations++
		// Intentionally do NOT re-assign EditorSave from within.
	}

	EditBuffer(ch, "/s")
	// EditorSave was cleared before invocation; a subsequent StartEditing
	// with no re-assignment, followed by /s, must not re-fire the prior save.
	StartEditing(ch, "")
	EditBuffer(ch, "/s")

	if invocations != 1 {
		t.Errorf("one-shot clear failed: callback fired %d times, want 1", invocations)
	}
}
