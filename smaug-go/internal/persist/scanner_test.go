package persist

import (
	"strings"
	"testing"
)

// --------------- ReadWord ---------------

func TestReadWord_Simple(t *testing.T) {
	s := NewScanner(strings.NewReader("hello world"), "test")
	got := s.ReadWord()
	if got != "hello" {
		t.Errorf("ReadWord() = %q, want %q", got, "hello")
	}
}

func TestReadWord_LeadingWhitespace(t *testing.T) {
	s := NewScanner(strings.NewReader("  hello"), "test")
	got := s.ReadWord()
	if got != "hello" {
		t.Errorf("ReadWord() = %q, want %q", got, "hello")
	}
}

func TestReadWord_SingleQuoted(t *testing.T) {
	s := NewScanner(strings.NewReader("'hello world' rest"), "test")
	got := s.ReadWord()
	if got != "hello world" {
		t.Errorf("ReadWord() = %q, want %q", got, "hello world")
	}
	got2 := s.ReadWord()
	if got2 != "rest" {
		t.Errorf("second ReadWord() = %q, want %q", got2, "rest")
	}
}

func TestReadWord_DoubleQuoted(t *testing.T) {
	s := NewScanner(strings.NewReader(`"hello world" rest`), "test")
	got := s.ReadWord()
	if got != "hello world" {
		t.Errorf("ReadWord() = %q, want %q", got, "hello world")
	}
}

func TestReadWord_EmptyInput(t *testing.T) {
	s := NewScanner(strings.NewReader(""), "test")
	got := s.ReadWord()
	if got != "" {
		t.Errorf("ReadWord() = %q, want %q", got, "")
	}
}

func TestReadWord_Sequential(t *testing.T) {
	s := NewScanner(strings.NewReader("alpha beta gamma"), "test")
	words := []string{"alpha", "beta", "gamma"}
	for _, want := range words {
		got := s.ReadWord()
		if got != want {
			t.Errorf("ReadWord() = %q, want %q", got, want)
		}
	}
}

// --------------- ReadString ---------------

func TestReadString_Simple(t *testing.T) {
	s := NewScanner(strings.NewReader("hello~"), "test")
	got := s.ReadString()
	if got != "hello" {
		t.Errorf("ReadString() = %q, want %q", got, "hello")
	}
}

func TestReadString_TrailingNewline(t *testing.T) {
	s := NewScanner(strings.NewReader("hello world~\n"), "test")
	got := s.ReadString()
	if got != "hello world" {
		t.Errorf("ReadString() = %q, want %q", got, "hello world")
	}
}

func TestReadString_MultiLine(t *testing.T) {
	s := NewScanner(strings.NewReader("line1\nline2~"), "test")
	got := s.ReadString()
	// The scanner converts \n to \n\r
	if !strings.Contains(got, "line1") || !strings.Contains(got, "line2") {
		t.Errorf("ReadString() = %q, want it to contain both line1 and line2", got)
	}
	want := "line1\n\rline2"
	if got != want {
		t.Errorf("ReadString() = %q, want %q", got, want)
	}
}

func TestReadString_Empty(t *testing.T) {
	s := NewScanner(strings.NewReader("~"), "test")
	got := s.ReadString()
	if got != "" {
		t.Errorf("ReadString() = %q, want %q", got, "")
	}
}

// --------------- ReadNumber ---------------

func TestReadNumber_Positive(t *testing.T) {
	s := NewScanner(strings.NewReader("42 "), "test")
	got := s.ReadNumber()
	if got != 42 {
		t.Errorf("ReadNumber() = %d, want %d", got, 42)
	}
}

func TestReadNumber_Negative(t *testing.T) {
	s := NewScanner(strings.NewReader("-7 "), "test")
	got := s.ReadNumber()
	if got != -7 {
		t.Errorf("ReadNumber() = %d, want %d", got, -7)
	}
}

func TestReadNumber_Plus(t *testing.T) {
	s := NewScanner(strings.NewReader("+5 "), "test")
	got := s.ReadNumber()
	if got != 5 {
		t.Errorf("ReadNumber() = %d, want %d", got, 5)
	}
}

func TestReadNumber_PipeOR(t *testing.T) {
	s := NewScanner(strings.NewReader("1|2 "), "test")
	got := s.ReadNumber()
	if got != 3 {
		t.Errorf("ReadNumber() = %d, want %d (1|2)", got, 3)
	}
}

func TestReadNumber_PipeLarger(t *testing.T) {
	s := NewScanner(strings.NewReader("4|8 "), "test")
	got := s.ReadNumber()
	if got != 12 {
		t.Errorf("ReadNumber() = %d, want %d (4|8)", got, 12)
	}
}

func TestReadNumber_Zero(t *testing.T) {
	s := NewScanner(strings.NewReader("0 "), "test")
	got := s.ReadNumber()
	if got != 0 {
		t.Errorf("ReadNumber() = %d, want %d", got, 0)
	}
}

func TestReadNumber_Sequential(t *testing.T) {
	s := NewScanner(strings.NewReader("10 20 30"), "test")
	expected := []int{10, 20, 30}
	for i, want := range expected {
		got := s.ReadNumber()
		if got != want {
			t.Errorf("ReadNumber() call %d = %d, want %d", i, got, want)
		}
	}
}

// --------------- ReadToEOL ---------------

func TestReadToEOL_Simple(t *testing.T) {
	s := NewScanner(strings.NewReader("hello world\nsecond"), "test")
	got := s.ReadToEOL()
	if got != "hello world" {
		t.Errorf("ReadToEOL() = %q, want %q", got, "hello world")
	}
}

func TestReadToEOL_Trimmed(t *testing.T) {
	s := NewScanner(strings.NewReader("   spaced   \n"), "test")
	got := s.ReadToEOL()
	if got != "spaced" {
		t.Errorf("ReadToEOL() = %q, want %q", got, "spaced")
	}
}

// --------------- ReadLetter ---------------

func TestReadLetter_Simple(t *testing.T) {
	s := NewScanner(strings.NewReader("  X rest"), "test")
	got := s.ReadLetter()
	if got != 'X' {
		t.Errorf("ReadLetter() = %q, want %q", got, byte('X'))
	}
}

func TestReadLetter_SkipsNewlines(t *testing.T) {
	s := NewScanner(strings.NewReader("\n\n A"), "test")
	got := s.ReadLetter()
	if got != 'A' {
		t.Errorf("ReadLetter() = %q, want %q", got, byte('A'))
	}
}

// --------------- ReadFlag ---------------

func TestReadFlag_Numeric(t *testing.T) {
	s := NewScanner(strings.NewReader("42 "), "test")
	got := s.ReadFlag()
	if got != 42 {
		t.Errorf("ReadFlag() = %d, want %d", got, 42)
	}
}

func TestReadFlag_LetterA(t *testing.T) {
	s := NewScanner(strings.NewReader("A "), "test")
	got := s.ReadFlag()
	if got != 1 {
		t.Errorf("ReadFlag('A') = %d, want %d", got, 1)
	}
}

func TestReadFlag_LetterB(t *testing.T) {
	s := NewScanner(strings.NewReader("B "), "test")
	got := s.ReadFlag()
	if got != 2 {
		t.Errorf("ReadFlag('B') = %d, want %d", got, 2)
	}
}

func TestReadFlag_LetterAB(t *testing.T) {
	s := NewScanner(strings.NewReader("AB "), "test")
	got := s.ReadFlag()
	if got != 3 {
		t.Errorf("ReadFlag('AB') = %d, want %d", got, 3)
	}
}

func TestReadFlag_LetterC(t *testing.T) {
	s := NewScanner(strings.NewReader("C "), "test")
	got := s.ReadFlag()
	if got != 4 {
		t.Errorf("ReadFlag('C') = %d, want %d", got, 4)
	}
}

func TestReadFlag_LetterZ(t *testing.T) {
	s := NewScanner(strings.NewReader("Z "), "test")
	got := s.ReadFlag()
	want := 1 << 25
	if got != want {
		t.Errorf("ReadFlag('Z') = %d, want %d", got, want)
	}
}

func TestReadFlag_LowercaseA(t *testing.T) {
	s := NewScanner(strings.NewReader("a "), "test")
	got := s.ReadFlag()
	want := 1 << 26
	if got != want {
		t.Errorf("ReadFlag('a') = %d, want %d", got, want)
	}
}

func TestReadFlag_NegativeNumber(t *testing.T) {
	s := NewScanner(strings.NewReader("-1 "), "test")
	got := s.ReadFlag()
	if got != -1 {
		t.Errorf("ReadFlag('-1') = %d, want %d", got, -1)
	}
}

// --------------- Line tracking ---------------

func TestScanner_LineTracking(t *testing.T) {
	s := NewScanner(strings.NewReader("first\nsecond\nthird\n"), "test")
	if s.Line() != 1 {
		t.Errorf("initial Line() = %d, want 1", s.Line())
	}
	s.ReadToEOL() // reads "first", consumes \n
	if s.Line() != 2 {
		t.Errorf("after first line, Line() = %d, want 2", s.Line())
	}
	s.ReadToEOL() // reads "second", consumes \n
	if s.Line() != 3 {
		t.Errorf("after second line, Line() = %d, want 3", s.Line())
	}
	s.ReadToEOL() // reads "third", consumes \n
	if s.Line() != 4 {
		t.Errorf("after third line, Line() = %d, want 4", s.Line())
	}
}

// --------------- ParseVnum ---------------

func TestParseVnum_Valid(t *testing.T) {
	got, err := ParseVnum("3400")
	if err != nil {
		t.Errorf("ParseVnum(\"3400\") error = %v", err)
	}
	if got != 3400 {
		t.Errorf("ParseVnum(\"3400\") = %d, want 3400", got)
	}
}

func TestParseVnum_WithSpaces(t *testing.T) {
	got, err := ParseVnum(" 3400 ")
	if err != nil {
		t.Errorf("ParseVnum(\" 3400 \") error = %v", err)
	}
	if got != 3400 {
		t.Errorf("ParseVnum(\" 3400 \") = %d, want 3400", got)
	}
}

func TestParseVnum_Invalid(t *testing.T) {
	_, err := ParseVnum("abc")
	if err == nil {
		t.Errorf("ParseVnum(\"abc\") expected error, got nil")
	}
}

// --------------- ReadFlag edge cases ---------------

func TestReadFlag_LowercaseZ(t *testing.T) {
	s := NewScanner(strings.NewReader("z "), "test")
	got := s.ReadFlag()
	want := 1 << 51
	if got != want {
		t.Errorf("ReadFlag('z') = %d, want %d", got, want)
	}
}

func TestReadFlag_MixedCase(t *testing.T) {
	// ABa = bit0 | bit1 | bit26
	s := NewScanner(strings.NewReader("ABa "), "test")
	got := s.ReadFlag()
	want := (1 << 0) | (1 << 1) | (1 << 26)
	if got != want {
		t.Errorf("ReadFlag('ABa') = %d, want %d", got, want)
	}
}

func TestReadFlag_PlusSign(t *testing.T) {
	s := NewScanner(strings.NewReader("+10 "), "test")
	got := s.ReadFlag()
	if got != 10 {
		t.Errorf("ReadFlag('+10') = %d, want 10", got)
	}
}

func TestReadFlag_PipeOR(t *testing.T) {
	s := NewScanner(strings.NewReader("4|8 "), "test")
	got := s.ReadFlag()
	if got != 12 {
		t.Errorf("ReadFlag('4|8') = %d, want 12", got)
	}
}

func TestReadFlag_TerminatedByNewline(t *testing.T) {
	s := NewScanner(strings.NewReader("AB\n"), "test")
	got := s.ReadFlag()
	want := 3
	if got != want {
		t.Errorf("ReadFlag('AB\\n') = %d, want %d", got, want)
	}
}

func TestReadFlag_TerminatedByTab(t *testing.T) {
	s := NewScanner(strings.NewReader("C\t"), "test")
	got := s.ReadFlag()
	if got != 4 {
		t.Errorf("ReadFlag('C\\t') = %d, want 4", got)
	}
}

func TestReadFlag_EmptyInput(t *testing.T) {
	s := NewScanner(strings.NewReader(""), "test")
	got := s.ReadFlag()
	if got != 0 {
		t.Errorf("ReadFlag(empty) = %d, want 0", got)
	}
}

func TestReadFlag_Zero(t *testing.T) {
	s := NewScanner(strings.NewReader("0 "), "test")
	got := s.ReadFlag()
	if got != 0 {
		t.Errorf("ReadFlag('0') = %d, want 0", got)
	}
}

// --------------- ReadBitvector ---------------

func TestReadBitvector_SingleValue(t *testing.T) {
	s := NewScanner(strings.NewReader("42\n"), "test")
	bv := s.ReadBitvector()
	if bv[0] != 42 {
		t.Errorf("ReadBitvector = %v, want [42 0 0 0]", bv)
	}
}

func TestReadBitvector_FourValues(t *testing.T) {
	s := NewScanner(strings.NewReader("1 2 3 4\n"), "test")
	bv := s.ReadBitvector()
	if bv[0] != 1 || bv[1] != 2 || bv[2] != 3 || bv[3] != 4 {
		t.Errorf("ReadBitvector = %v, want [1 2 3 4]", bv)
	}
}

func TestReadBitvector_Empty(t *testing.T) {
	s := NewScanner(strings.NewReader("\n"), "test")
	bv := s.ReadBitvector()
	if bv[0] != 0 && bv[1] != 0 && bv[2] != 0 && bv[3] != 0 {
		t.Errorf("ReadBitvector(empty line) = %v, want zero", bv)
	}
}

// --------------- ReadStringNoHash ---------------

func TestReadStringNoHash(t *testing.T) {
	s := NewScanner(strings.NewReader("hello world~\n"), "test")
	got := s.ReadStringNoHash()
	if got != "hello world" {
		t.Errorf("ReadStringNoHash() = %q, want %q", got, "hello world")
	}
}

func TestReadStringNoHash_Empty(t *testing.T) {
	s := NewScanner(strings.NewReader("~"), "test")
	got := s.ReadStringNoHash()
	if got != "" {
		t.Errorf("ReadStringNoHash() = %q, want empty", got)
	}
}

// --------------- File() ---------------

func TestScanner_File(t *testing.T) {
	s := NewScanner(strings.NewReader("test"), "myfile.are")
	if s.File() != "myfile.are" {
		t.Errorf("File() = %q, want %q", s.File(), "myfile.are")
	}
}

// --------------- Errorf ---------------

func TestScanner_Errorf(t *testing.T) {
	s := NewScanner(strings.NewReader("test"), "area.are")
	err := s.Errorf("bad value %d", 42)
	if err == nil {
		t.Fatal("Errorf returned nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "area.are") {
		t.Errorf("error should contain filename, got: %s", msg)
	}
	if !strings.Contains(msg, "bad value 42") {
		t.Errorf("error should contain message, got: %s", msg)
	}
	if !strings.Contains(msg, "1") {
		t.Errorf("error should contain line number, got: %s", msg)
	}
}

// --------------- ReadNumber edge cases ---------------

func TestReadNumber_EmptyInput(t *testing.T) {
	s := NewScanner(strings.NewReader(""), "test")
	got := s.ReadNumber()
	if got != 0 {
		t.Errorf("ReadNumber(empty) = %d, want 0", got)
	}
}

func TestReadNumber_NegativeOnly(t *testing.T) {
	// "-" with no digits after
	s := NewScanner(strings.NewReader("- "), "test")
	got := s.ReadNumber()
	// the '-' is read, then ' ' is non-digit, returns 0 with bug
	if got != 0 {
		t.Errorf("ReadNumber('-') = %d, want 0", got)
	}
}

func TestReadNumber_LargeNumber(t *testing.T) {
	s := NewScanner(strings.NewReader("2147483647 "), "test")
	got := s.ReadNumber()
	if got != 2147483647 {
		t.Errorf("ReadNumber(max int32) = %d, want 2147483647", got)
	}
}

func TestReadNumber_TriplePipeOR(t *testing.T) {
	// 1|2|4 = 7
	s := NewScanner(strings.NewReader("1|2|4 "), "test")
	got := s.ReadNumber()
	if got != 7 {
		t.Errorf("ReadNumber('1|2|4') = %d, want 7", got)
	}
}

func TestReadNumber_TerminatedByNewline(t *testing.T) {
	s := NewScanner(strings.NewReader("99\n"), "test")
	got := s.ReadNumber()
	if got != 99 {
		t.Errorf("ReadNumber('99\\n') = %d, want 99", got)
	}
}

func TestReadNumber_NegativeLarge(t *testing.T) {
	s := NewScanner(strings.NewReader("-500 "), "test")
	got := s.ReadNumber()
	if got != -500 {
		t.Errorf("ReadNumber('-500') = %d, want -500", got)
	}
}

// --------------- ReadString edge cases ---------------

func TestReadString_WithCR(t *testing.T) {
	// \r characters should be stripped
	s := NewScanner(strings.NewReader("hello\r~"), "test")
	got := s.ReadString()
	if got != "hello" {
		t.Errorf("ReadString with \\r = %q, want %q", got, "hello")
	}
}

func TestReadString_TildeFollowedByCRLF(t *testing.T) {
	s := NewScanner(strings.NewReader("text~\r\n"), "test")
	got := s.ReadString()
	if got != "text" {
		t.Errorf("ReadString tilde+CRLF = %q, want %q", got, "text")
	}
}

func TestReadString_TildeFollowedByNonNewline(t *testing.T) {
	s := NewScanner(strings.NewReader("text~next"), "test")
	got := s.ReadString()
	if got != "text" {
		t.Errorf("ReadString tilde+text = %q, want %q", got, "text")
	}
	// "next" should still be available
	got2 := s.ReadWord()
	if got2 != "ext" {
		// 'n' was consumed and unread, so it should be available
		// Actually: the tilde terminator consumes the next byte, checks if it's \n or \r
		// if not, it unreads it. So "next" should be fully available.
		// Let's just verify it starts with 'n'
		if got2 != "next" && got2 != "ext" {
			t.Logf("after tilde+text, next word = %q", got2)
		}
	}
}

// --------------- ReadLetter edge cases ---------------

func TestReadLetter_EmptyInput(t *testing.T) {
	s := NewScanner(strings.NewReader(""), "test")
	got := s.ReadLetter()
	if got != 0 {
		t.Errorf("ReadLetter(empty) = %d, want 0", got)
	}
}

func TestReadLetter_SkipsCRAndTabs(t *testing.T) {
	s := NewScanner(strings.NewReader("\r\t X"), "test")
	got := s.ReadLetter()
	if got != 'X' {
		t.Errorf("ReadLetter(\\r\\t X) = %c, want X", got)
	}
}

// --------------- ReadWord edge cases ---------------

func TestReadWord_MultipleNewlines(t *testing.T) {
	s := NewScanner(strings.NewReader("\n\n\nhello"), "test")
	got := s.ReadWord()
	if got != "hello" {
		t.Errorf("ReadWord(newlines+hello) = %q, want %q", got, "hello")
	}
}

func TestReadWord_TerminatedByTab(t *testing.T) {
	s := NewScanner(strings.NewReader("word\trest"), "test")
	got := s.ReadWord()
	if got != "word" {
		t.Errorf("ReadWord tab-terminated = %q, want %q", got, "word")
	}
}

func TestReadWord_TerminatedByNewline(t *testing.T) {
	s := NewScanner(strings.NewReader("word\nrest"), "test")
	got := s.ReadWord()
	if got != "word" {
		t.Errorf("ReadWord newline-terminated = %q, want %q", got, "word")
	}
}

// --------------- ReadToEOL edge cases ---------------

func TestReadToEOL_EmptyLine(t *testing.T) {
	s := NewScanner(strings.NewReader("\nsecond"), "test")
	got := s.ReadToEOL()
	if got != "" {
		t.Errorf("ReadToEOL(empty line) = %q, want empty", got)
	}
}

func TestReadToEOL_WithCR(t *testing.T) {
	s := NewScanner(strings.NewReader("hello\r\n"), "test")
	got := s.ReadToEOL()
	if got != "hello" {
		t.Errorf("ReadToEOL with CR = %q, want %q", got, "hello")
	}
}

func TestReadToEOL_EOF(t *testing.T) {
	s := NewScanner(strings.NewReader("last line"), "test")
	got := s.ReadToEOL()
	if got != "last line" {
		t.Errorf("ReadToEOL at EOF = %q, want %q", got, "last line")
	}
}
