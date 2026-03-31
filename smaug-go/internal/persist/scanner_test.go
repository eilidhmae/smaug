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
