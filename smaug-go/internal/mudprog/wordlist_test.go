package mudprog

import "testing"

// Covers plan-tranche-b.md G4 word-boundary match helper.
//
// C reference: `rprog_wordlist_check` at src/mud_prog.c:4109-4178.
// Algorithm:
//   1. lower-case arglist and input
//   2. if arglist starts with "p " → phrase match (remaining = one phrase)
//   3. else split arglist on whitespace and keyword-match each
//   4. word-boundary: left must be start-of-text or ' ', right must be
//      end-of-text, ' ', '\n', or '\r' (NOT '\t', NOT punctuation).

func TestWordlistMatch_SingleKeywordMatches(t *testing.T) {
	if !wordlistMatch("hello", "hello world") {
		t.Error("`hello` should match in `hello world`")
	}
}

func TestWordlistMatch_SubstringWithoutWordBoundaryFails(t *testing.T) {
	if wordlistMatch("ell", "hello") {
		t.Error("`ell` should not match in `hello` (no left boundary)")
	}
}

func TestWordlistMatch_MultipleKeywordsAnyMatches(t *testing.T) {
	if !wordlistMatch("foo bar", "bar") {
		t.Error("`foo bar` keywords should match `bar`")
	}
	if !wordlistMatch("foo bar", "foo") {
		t.Error("`foo bar` keywords should match `foo`")
	}
	if wordlistMatch("foo bar", "baz") {
		t.Error("`foo bar` should not match `baz`")
	}
}

func TestWordlistMatch_TrailingPunctuationFails(t *testing.T) {
	// '!' is not a listed right-boundary char in C; `hello!` shouldn't match.
	if wordlistMatch("hello", "hello!") {
		t.Error("`hello` should not match `hello!` (right boundary '!' not listed)")
	}
}

func TestWordlistMatch_TrailingNewlineMatches(t *testing.T) {
	if !wordlistMatch("hello", "hello\n") {
		t.Error("`hello` should match `hello\\n`")
	}
	if !wordlistMatch("hello", "hello\r") {
		t.Error("`hello` should match `hello\\r`")
	}
}

func TestWordlistMatch_TrailingNullMatches(t *testing.T) {
	// "end-of-text" boundary: input terminator matches C's '\0' check.
	if !wordlistMatch("hello", "hello") {
		t.Error("`hello` should match exact `hello`")
	}
}

func TestWordlistMatch_PhrasePrefix(t *testing.T) {
	if !wordlistMatch("p I am the king", "I am the king of the hill") {
		t.Error("phrase prefix should match substring with boundaries")
	}
}

func TestWordlistMatch_PhrasePrefixAtEnd(t *testing.T) {
	// input ends with the phrase → right boundary is end-of-text.
	if !wordlistMatch("p I am the king", "Behold I am the king") {
		t.Error("phrase at end should match")
	}
}

func TestWordlistMatch_PhrasePrefixDoesNotMatchReordered(t *testing.T) {
	if wordlistMatch("p foo bar", "bar foo") {
		t.Error("phrase order matters; `foo bar` should not match `bar foo`")
	}
}

func TestWordlistMatch_CaseInsensitive(t *testing.T) {
	if !wordlistMatch("Hello", "hello world") {
		t.Error("case-insensitive match expected")
	}
	if !wordlistMatch("hello", "HELLO WORLD") {
		t.Error("case-insensitive match expected")
	}
}

func TestWordlistMatch_EmptyArglistFalse(t *testing.T) {
	// Defensive: the wordlist helper shouldn't match everything when
	// arglist is empty — callers (RprogCommandTrigger etc.) already
	// short-circuit on empty arglist before calling us.
	if wordlistMatch("", "anything") {
		t.Error("empty arglist should be false from the helper")
	}
}

func TestWordlistMatch_EmptyInputFalse(t *testing.T) {
	if wordlistMatch("hello", "") {
		t.Error("empty input should be false")
	}
}

func TestWordlistMatch_LeftBoundaryWithSpace(t *testing.T) {
	// `bar` is a standalone word in `foo bar`; boundary on both sides.
	if !wordlistMatch("bar", "foo bar") {
		t.Error("`bar` should match in `foo bar` (space boundary)")
	}
	// `bar` is inside `foobar`; no left boundary → no match.
	if wordlistMatch("bar", "foobar") {
		t.Error("`bar` should not match in `foobar` (no left space boundary)")
	}
}

// --- containsAtWordBoundary direct tests (internal helper) ---

func TestContainsAtWordBoundary_MiddleWord(t *testing.T) {
	if !containsAtWordBoundary("one two three", "two") {
		t.Error("`two` middle-of-sentence should match")
	}
}

func TestContainsAtWordBoundary_NotWordInterior(t *testing.T) {
	if containsAtWordBoundary("sweater", "eat") {
		t.Error("`eat` inside `sweater` should not match")
	}
}

func TestContainsAtWordBoundary_Empty(t *testing.T) {
	if containsAtWordBoundary("foo", "") {
		t.Error("empty needle should be false")
	}
}
