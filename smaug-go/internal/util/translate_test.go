package util

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

func makeSpeaker(lang uint32) *types.CharData {
	return &types.CharData{Name: "Speaker", Level: 10, Speaking: int(lang), Speaks: int(lang)}
}

func makeListener(speaks uint32) *types.CharData {
	return &types.CharData{Name: "Listener", Level: 10, Speaks: int(speaks)}
}

func TestLangBit_KnownName(t *testing.T) {
	cases := map[string]uint32{
		"common":   types.LANG_COMMON,
		"Elven":    types.LANG_ELVEN,
		"DWARVEN":  types.LANG_DWARVEN,
		"ORCISH":   types.LANG_ORCISH,
		"trollish": types.LANG_TROLLISH,
	}
	for name, want := range cases {
		got, ok := LangBit(name)
		if !ok || got != want {
			t.Errorf("LangBit(%q) = %v, %v; want %v, true", name, got, ok, want)
		}
	}
}

func TestLangBit_PrefixMatch(t *testing.T) {
	got, ok := LangBit("elv")
	if !ok || got != types.LANG_ELVEN {
		t.Errorf("prefix 'elv' should resolve to LANG_ELVEN, got %v, %v", got, ok)
	}
}

func TestLangBit_Unknown(t *testing.T) {
	if _, ok := LangBit("klingon"); ok {
		t.Error("'klingon' should not resolve to any language")
	}
	if _, ok := LangBit(""); ok {
		t.Error("empty name should not resolve")
	}
}

func TestTranslate_SharedLanguagePassthrough(t *testing.T) {
	sp := makeSpeaker(types.LANG_ELVEN)
	li := makeListener(types.LANG_ELVEN | types.LANG_COMMON)
	got := Translate("hello friend", sp, li)
	if got != "hello friend" {
		t.Errorf("shared language should pass through; got %q", got)
	}
}

func TestTranslate_CommonPassthrough(t *testing.T) {
	sp := makeSpeaker(types.LANG_COMMON)
	li := makeListener(types.LANG_COMMON)
	got := Translate("hi there", sp, li)
	if got != "hi there" {
		t.Errorf("common should pass through; got %q", got)
	}
}

func TestTranslate_ImmortalUnderstandsAll(t *testing.T) {
	sp := makeSpeaker(types.LANG_ELVEN)
	li := makeListener(types.LANG_COMMON)
	li.Trust = types.LEVEL_IMMORTAL
	got := Translate("secret", sp, li)
	if got != "secret" {
		t.Errorf("immortal should understand; got %q", got)
	}
}

func TestTranslate_Scrambled(t *testing.T) {
	sp := makeSpeaker(types.LANG_ELVEN)
	li := makeListener(types.LANG_COMMON) // no elven
	in := "hello friend"
	got := Translate(in, sp, li)
	if got == in {
		t.Errorf("non-shared language should be scrambled; got %q", got)
	}
	if len(got) != len(in) {
		t.Errorf("scramble should preserve length; %d != %d (%q)", len(got), len(in), got)
	}
	// Whitespace preserved positionally.
	for i := range in {
		if in[i] == ' ' && got[i] != ' ' {
			t.Errorf("space at %d lost: in=%q out=%q", i, in, got)
		}
	}
}

func TestTranslate_CasePreserved(t *testing.T) {
	sp := makeSpeaker(types.LANG_DWARVEN)
	li := makeListener(types.LANG_COMMON)
	got := Translate("Hello World!", sp, li)
	if len(got) != len("Hello World!") {
		t.Fatalf("length mismatch: %q", got)
	}
	// 'H' should map to an uppercase letter; '!' preserved.
	if got[0] < 'A' || got[0] > 'Z' {
		t.Errorf("expected upper-case leading char, got %q", got)
	}
	if !strings.HasSuffix(got, "!") {
		t.Errorf("punctuation should be preserved; got %q", got)
	}
}

func TestTranslate_ColorCodesPreserved(t *testing.T) {
	sp := makeSpeaker(types.LANG_ORCISH)
	li := makeListener(types.LANG_COMMON)
	got := Translate("&Rhi&D", sp, li)
	if !strings.HasPrefix(got, "&R") || !strings.HasSuffix(got, "&D") {
		t.Errorf("color codes should be preserved; got %q", got)
	}
}

func TestTranslate_DeterministicAcrossCalls(t *testing.T) {
	sp := makeSpeaker(types.LANG_ELVEN)
	li := makeListener(types.LANG_COMMON)
	a := Translate("mellon", sp, li)
	b := Translate("mellon", sp, li)
	if a != b {
		t.Errorf("translate should be deterministic; %q != %q", a, b)
	}
}

func TestTranslate_NilSafe(t *testing.T) {
	if got := Translate("x", nil, nil); got != "x" {
		t.Errorf("nil speaker/listener should pass through; got %q", got)
	}
}
