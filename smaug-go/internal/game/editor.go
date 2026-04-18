package game

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
)

const (
	maxBufLines   = 24 // normal editing limit
	maxBufLinesMp = 48 // mudprog/help editing limit
	maxLineLen    = 79
	formatLineLen = 75
)

// editorHelp is the help text for the string editor.
const editorHelp = "Editing commands:\n\r" +
	"-----------------------------------------\n\r" +
	"/l              list buffer\n\r" +
	"/c              clear buffer\n\r" +
	"/d [line]       delete line\n\r" +
	"/g <line>       goto line\n\r" +
	"/i [line]       insert blank line\n\r" +
	"/r <old> <new>  global replace\n\r" +
	"/f              format (word wrap)\n\r" +
	"/a              abort (discard changes)\n\r" +
	"/s              save and exit\n\r" +
	"/?              this help\n\r" +
	"-----------------------------------------\n\r"

// StartEditing initializes the line editor for a character.
// Maps to C start_editing().
func StartEditing(ch *types.CharData, text string) {
	if ch == nil || ch.Desc == nil {
		return
	}

	ed := &types.EditorData{}

	// Parse existing text into lines
	if text != "" {
		// Normalize line endings: \n\r -> \n, \r\n -> \n, \r -> \n
		text = strings.ReplaceAll(text, "\n\r", "\n")
		text = strings.ReplaceAll(text, "\r\n", "\n")
		text = strings.ReplaceAll(text, "\r", "\n")

		lines := strings.Split(text, "\n")
		for _, line := range lines {
			if ed.NumLines >= 49 {
				break
			}
			// Skip trailing empty line from split
			if line == "" && ed.NumLines > 0 {
				continue
			}
			if line == "" {
				continue
			}
			// Truncate long lines
			if len(line) > maxLineLen-1 {
				line = line[:maxLineLen-1]
			}
			ed.Lines[ed.NumLines] = line
			ed.Size += len(line)
			ed.NumLines++
		}
	}

	ed.OnLine = ed.NumLines

	ch.Editor = ed
	ch.Desc.Connected = types.CON_EDITING

	ch.Sendf("Begin entering your text now (/? = help /s = save /c = clear /l = list)\n\r")
	ch.Sendf("-----------------------------------------------------------------------\n\r")
	ch.Send("> ")
}

// StopEditing cleans up the editor and returns to playing state.
// Maps to C stop_editing().
func StopEditing(ch *types.CharData) {
	if ch == nil {
		return
	}
	ch.Editor = nil
	ch.Substate = types.SUB_NONE
	// Defensive: drop any pending callback so a leaked closure can't
	// hold a large context (plan § Open Q 1).
	ch.EditorSave = nil
	if ch.Desc != nil {
		ch.Desc.Connected = types.CON_PLAYING
	}
}

// CopyBuffer converts the editor buffer back to a string.
// Maps to C copy_buffer().
func CopyBuffer(ch *types.CharData) string {
	if ch == nil || ch.Editor == nil || ch.Editor.NumLines == 0 {
		return ""
	}
	ed := ch.Editor
	var sb strings.Builder
	for i := 0; i < ed.NumLines; i++ {
		sb.WriteString(ed.Lines[i])
		sb.WriteString("\n\r")
	}
	return sb.String()
}

// EditBuffer processes one line of editor input.
// Maps to C edit_buffer().
func EditBuffer(ch *types.CharData, line string) {
	if ch == nil || ch.Editor == nil {
		return
	}

	ed := ch.Editor

	// Check for editor commands (start with / or \)
	if len(line) >= 2 && (line[0] == '/' || line[0] == '\\') {
		cmd := line[1]
		args := ""
		if len(line) > 2 {
			args = strings.TrimSpace(line[2:])
		}

		switch cmd {
		case '?':
			ch.Send(editorHelp)

		case 'l':
			editorList(ch)

		case 'c':
			editorClear(ch)

		case 'd':
			editorDelete(ch, args)

		case 'g':
			editorGoto(ch, args)

		case 'i':
			editorInsert(ch, args)

		case 'r':
			editorReplace(ch, args)

		case 'f':
			editorFormat(ch)

		case 'a':
			ch.Send("Editing aborted.\n\r")
			StopEditing(ch)
			return

		case 's':
			// Transition the descriptor out of CON_EDITING FIRST (matches C
			// build.c:7004-7010) so the callback can re-enter StartEditing
			// safely for nested edits.
			if ch.Desc != nil {
				ch.Desc.Connected = types.CON_PLAYING
			}
			// One-shot callback: clear BEFORE invoking so a nested
			// StartEditing inside the save handler can install its own.
			if ch.EditorSave != nil {
				save := ch.EditorSave
				ch.EditorSave = nil
				save(ch)
			}
			return

		default:
			ch.Sendf("Unknown editor command '/%c'. Use /? for help.\n\r", cmd)
		}

		if ch.Editor != nil {
			ch.Send("> ")
		}
		return
	}

	// Normal text input — append a line
	maxLines := maxBufLines
	if ch.Substate == types.SUB_MPROG_EDIT || ch.Substate == types.SUB_HELP_EDIT {
		maxLines = maxBufLinesMp
	}

	if ed.NumLines >= maxLines {
		ch.Sendf("Buffer is full (%d lines max).\n\r", maxLines)
		ch.Send("> ")
		return
	}

	// Truncate if too long
	if len(line) > maxLineLen-1 {
		ch.Send("Line too long, truncated.\n\r")
		line = line[:maxLineLen-1]
	}

	ed.Lines[ed.NumLines] = line
	ed.Size += len(line)
	ed.NumLines++
	ed.OnLine = ed.NumLines

	ch.Send("> ")
}

// editorList displays the buffer with line numbers.
func editorList(ch *types.CharData) {
	ed := ch.Editor
	if ed.NumLines == 0 {
		ch.Send("Buffer is empty.\n\r")
		return
	}
	for i := 0; i < ed.NumLines; i++ {
		ch.Sendf("%2d> %s\n\r", i+1, ed.Lines[i])
	}
}

// editorClear empties the buffer.
func editorClear(ch *types.CharData) {
	ed := ch.Editor
	for i := 0; i < ed.NumLines; i++ {
		ed.Lines[i] = ""
	}
	ed.NumLines = 0
	ed.Size = 0
	ed.OnLine = 0
	ch.Send("Buffer cleared.\n\r")
}

// editorDelete removes a line from the buffer.
func editorDelete(ch *types.CharData, args string) {
	ed := ch.Editor
	lineNum := ed.NumLines // default: delete last line (1-indexed)
	if args != "" {
		n, err := strconv.Atoi(args)
		if err != nil || n < 1 || n > ed.NumLines {
			ch.Sendf("Valid range is 1 to %d.\n\r", ed.NumLines)
			return
		}
		lineNum = n
	}
	if ed.NumLines == 0 {
		ch.Send("Buffer is empty.\n\r")
		return
	}

	idx := lineNum - 1
	ed.Size -= len(ed.Lines[idx])

	// Shift lines up
	for i := idx; i < ed.NumLines-1; i++ {
		ed.Lines[i] = ed.Lines[i+1]
	}
	ed.Lines[ed.NumLines-1] = ""
	ed.NumLines--

	if ed.OnLine > ed.NumLines {
		ed.OnLine = ed.NumLines
	}

	ch.Sendf("Line %d deleted.\n\r", lineNum)
}

// editorGoto sets the current line position.
func editorGoto(ch *types.CharData, args string) {
	ed := ch.Editor
	if args == "" {
		ch.Send("Goto which line?\n\r")
		return
	}
	n, err := strconv.Atoi(args)
	if err != nil || n < 1 || n > ed.NumLines+1 {
		ch.Sendf("Valid range is 1 to %d.\n\r", ed.NumLines+1)
		return
	}
	ed.OnLine = n - 1 // 0-indexed
	ch.Sendf("Now at line %d.\n\r", n)
}

// editorInsert inserts a blank line at the given position.
func editorInsert(ch *types.CharData, args string) {
	ed := ch.Editor

	lineNum := ed.OnLine + 1 // 1-indexed default
	if args != "" {
		n, err := strconv.Atoi(args)
		if err != nil || n < 1 || n > ed.NumLines+1 {
			ch.Sendf("Valid range is 1 to %d.\n\r", ed.NumLines+1)
			return
		}
		lineNum = n
	}

	maxLines := maxBufLines
	if ch.Substate == types.SUB_MPROG_EDIT || ch.Substate == types.SUB_HELP_EDIT {
		maxLines = maxBufLinesMp
	}
	if ed.NumLines >= maxLines {
		ch.Sendf("Buffer is full (%d lines max).\n\r", maxLines)
		return
	}

	idx := lineNum - 1 // 0-indexed

	// Shift lines down
	for i := ed.NumLines; i > idx; i-- {
		ed.Lines[i] = ed.Lines[i-1]
	}
	ed.Lines[idx] = ""
	ed.NumLines++
	ed.OnLine = idx

	ch.Sendf("Line inserted before line %d.\n\r", lineNum)
}

// editorReplace performs global search and replace.
func editorReplace(ch *types.CharData, args string) {
	ed := ch.Editor

	parts := strings.SplitN(args, " ", 2)
	if len(parts) < 2 {
		ch.Send("Usage: /r <old> <new>\n\r")
		return
	}
	old := parts[0]
	newStr := parts[1]

	count := 0
	for i := 0; i < ed.NumLines; i++ {
		if strings.Contains(ed.Lines[i], old) {
			newLine := strings.ReplaceAll(ed.Lines[i], old, newStr)
			ed.Size += len(newLine) - len(ed.Lines[i])
			ed.Lines[i] = newLine
			count++
		}
	}

	ch.Sendf("Replaced %d occurrence(s).\n\r", count)
}

// editorFormat reflows text to formatLineLen characters per line.
func editorFormat(ch *types.CharData) {
	ed := ch.Editor
	if ed.NumLines == 0 {
		ch.Send("Buffer is empty.\n\r")
		return
	}

	// Concatenate all text
	var sb strings.Builder
	for i := 0; i < ed.NumLines; i++ {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(ed.Lines[i])
	}
	text := sb.String()

	// Word-wrap
	words := strings.Fields(text)
	ed.NumLines = 0
	ed.Size = 0

	var line strings.Builder
	for _, word := range words {
		if line.Len() > 0 && line.Len()+1+len(word) > formatLineLen {
			// Flush current line
			if ed.NumLines < 49 {
				s := line.String()
				ed.Lines[ed.NumLines] = s
				ed.Size += len(s)
				ed.NumLines++
			}
			line.Reset()
		}
		if line.Len() > 0 {
			line.WriteByte(' ')
		}
		_, _ = fmt.Fprint(&line, word)
	}
	// Flush remaining
	if line.Len() > 0 && ed.NumLines < 49 {
		s := line.String()
		ed.Lines[ed.NumLines] = s
		ed.Size += len(s)
		ed.NumLines++
	}

	// Clear unused lines
	for i := ed.NumLines; i < 49; i++ {
		ed.Lines[i] = ""
	}

	ed.OnLine = ed.NumLines
	ch.Sendf("Text formatted. %d line(s).\n\r", ed.NumLines)
}
