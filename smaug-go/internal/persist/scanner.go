// Package persist handles reading and writing SMAUG data file formats.
package persist

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// Scanner reads SMAUG's custom file formats. It provides equivalents of
// the C functions fread_word, fread_string, fread_number, fread_to_eol,
// and fread_bitvector from db.c.
type Scanner struct {
	r    *bufio.Reader
	line int
	file string
}

// NewScanner creates a Scanner for the given reader.
func NewScanner(r io.Reader, filename string) *Scanner {
	return &Scanner{
		r:    bufio.NewReader(r),
		line: 1,
		file: filename,
	}
}

// ReadWord reads the next whitespace-delimited word, skipping leading
// whitespace and comments (lines starting with '#$'). Equivalent to
// fread_word() in db.c.
func (s *Scanner) ReadWord() string {
	s.skipWhitespace()

	c, err := s.readByte()
	if err != nil {
		return ""
	}

	var buf strings.Builder
	if c == '\'' || c == '"' {
		// Quoted string — read until matching quote
		delim := c
		for {
			c, err = s.readByte()
			if err != nil || c == delim {
				break
			}
			if c == '\n' {
				s.line++
			}
			buf.WriteByte(c)
		}
	} else {
		buf.WriteByte(c)
		for {
			c, err = s.readByte()
			if err != nil {
				break
			}
			if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
				if c == '\n' {
					s.line++
				}
				break
			}
			buf.WriteByte(c)
		}
	}
	return buf.String()
}

// ReadString reads a tilde-terminated string. In SMAUG file formats,
// strings are terminated by a ~ character. Equivalent to fread_string()
// in db.c.
func (s *Scanner) ReadString() string {
	s.skipWhitespace()

	var buf strings.Builder
	for {
		c, err := s.readByte()
		if err != nil {
			break
		}
		if c == '~' {
			// Consume optional trailing newline
			next, err := s.readByte()
			if err == nil && next != '\n' && next != '\r' {
				s.unreadByte()
			} else if next == '\r' {
				// consume \n after \r
				next2, err := s.readByte()
				if err == nil && next2 != '\n' {
					s.unreadByte()
				}
				s.line++
			} else if next == '\n' {
				s.line++
			}
			break
		}
		if c == '\n' {
			s.line++
			// SMAUG converts \n\r or \r\n sequences
			buf.WriteByte('\n')
			buf.WriteByte('\r')
			continue
		}
		if c == '\r' {
			// skip \r (handled above with \n)
			continue
		}
		buf.WriteByte(c)
	}
	return buf.String()
}

// ReadNumber reads an integer (possibly negative). Equivalent to
// fread_number() in db.c.
func (s *Scanner) ReadNumber() int {
	s.skipWhitespace()

	sign := 1
	c, err := s.readByte()
	if err != nil {
		return 0
	}
	if c == '+' {
		c, err = s.readByte()
		if err != nil {
			return 0
		}
	} else if c == '-' {
		sign = -1
		c, err = s.readByte()
		if err != nil {
			return 0
		}
	}

	if c < '0' || c > '9' {
		util.Bug("ReadNumber: bad format at %s:%d, char='%c'", s.file, s.line, c)
		return 0
	}

	number := 0
	for c >= '0' && c <= '9' {
		number = number*10 + int(c-'0')
		c, err = s.readByte()
		if err != nil {
			return sign * number
		}
	}

	// Check for pipe operator (used in some SMAUG formats for OR)
	if c == '|' {
		number |= s.ReadNumber()
	} else if c != ' ' && c != '\n' && c != '\r' && c != '\t' {
		s.unreadByte()
	}
	if c == '\n' {
		s.line++
	}

	return sign * number
}

// ReadToEOL reads and returns the rest of the current line.
// Equivalent to fread_to_eol() in db.c.
func (s *Scanner) ReadToEOL() string {
	var buf strings.Builder
	for {
		c, err := s.readByte()
		if err != nil || c == '\n' {
			if c == '\n' {
				s.line++
			}
			break
		}
		if c == '\r' {
			continue
		}
		buf.WriteByte(c)
	}
	return strings.TrimSpace(buf.String())
}

// ReadBitvector reads an extended bitvector (up to 4 space-separated uint32 values).
func (s *Scanner) ReadBitvector() types.BitVector {
	str := s.ReadToEOL()
	bv, err := types.ParseBitVector(str)
	if err != nil {
		util.Bug("ReadBitvector: parse error at %s:%d: %v", s.file, s.line, err)
	}
	return bv
}

// ReadLetter reads and returns the next non-whitespace character.
// Equivalent to fread_letter() in db.c.
func (s *Scanner) ReadLetter() byte {
	s.skipWhitespace()
	c, err := s.readByte()
	if err != nil {
		return 0
	}
	return c
}

// ReadStringNoHash reads a tilde-terminated string without hash-table
// deduplication (same as ReadString in Go since we don't need STRALLOC).
func (s *Scanner) ReadStringNoHash() string {
	return s.ReadString()
}

// Line returns the current line number.
func (s *Scanner) Line() int {
	return s.line
}

// File returns the filename being read.
func (s *Scanner) File() string {
	return s.file
}

// ReadFlag reads a flag value that can be in "number" or "ABCxyz" letter format.
// The letter format uses A-Z for bits 0-25 and a-z for bits 26-51.
func (s *Scanner) ReadFlag() int {
	s.skipWhitespace()
	c, err := s.readByte()
	if err != nil {
		return 0
	}

	// If it starts with a digit or sign, it's a number
	if (c >= '0' && c <= '9') || c == '-' || c == '+' {
		s.unreadByte()
		return s.ReadNumber()
	}

	// Letter-based flag format
	var flag int
	for {
		if c >= 'A' && c <= 'Z' {
			flag |= 1 << (int(c - 'A'))
		} else if c >= 'a' && c <= 'z' {
			flag |= 1 << (26 + int(c-'a'))
		} else {
			if c != '\n' && c != '\r' && c != ' ' && c != '\t' {
				s.unreadByte()
			}
			if c == '\n' {
				s.line++
			}
			break
		}
		c, err = s.readByte()
		if err != nil {
			break
		}
	}
	return flag
}

// Errorf formats and returns an error with file/line context.
func (s *Scanner) Errorf(format string, args ...any) error {
	prefix := fmt.Sprintf("%s:%d: ", s.file, s.line)
	return fmt.Errorf(prefix+format, args...)
}

// --- internal helpers ---

func (s *Scanner) skipWhitespace() {
	for {
		c, err := s.readByte()
		if err != nil {
			return
		}
		if c == '\n' {
			s.line++
			continue
		}
		if c == ' ' || c == '\t' || c == '\r' {
			continue
		}
		s.unreadByte()
		return
	}
}

func (s *Scanner) readByte() (byte, error) {
	return s.r.ReadByte()
}

func (s *Scanner) unreadByte() {
	_ = s.r.UnreadByte()
}

// ParseVnum parses a vnum string to int.
func ParseVnum(str string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(str))
}
