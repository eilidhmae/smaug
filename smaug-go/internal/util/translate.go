package util

import (
	"strings"
	"unicode"

	"github.com/eilidhmae/smaug/internal/types"
)

// LangName maps a LANG_ bit to its canonical lowercase name. Only the
// player-facing languages defined in VALID_LANGS are exposed to do_speak;
// others (GOD, MAGICAL, CLAN, etc.) are reserved for immortals or specific
// channels and are intentionally absent here.
var LangName = map[uint32]string{
	types.LANG_COMMON:   "common",
	types.LANG_ELVEN:    "elven",
	types.LANG_DWARVEN:  "dwarven",
	types.LANG_PIXIE:    "pixie",
	types.LANG_OGRE:     "ogre",
	types.LANG_ORCISH:   "orcish",
	types.LANG_TROLLISH: "trollish",
	types.LANG_GOBLIN:   "goblin",
	types.LANG_HALFLING: "halfling",
	types.LANG_GITH:     "gith",
	types.LANG_GNOME:    "gnome",
	types.LANG_DRAGON:   "dragon",
	types.LANG_CLAN:     "clan",
	types.LANG_MAGICAL:  "magical",
	types.LANG_GOD:      "god",
}

// LangBit returns the LANG_ bit for a named language, matching case-
// insensitively. The second return is false if the name is unknown.
func LangBit(name string) (uint32, bool) {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "" {
		return 0, false
	}
	for bit, cname := range LangName {
		if cname == n {
			return bit, true
		}
	}
	// Prefix match as a convenience ("el" → elven, "dw" → dwarven).
	for bit, cname := range LangName {
		if strings.HasPrefix(cname, n) {
			return bit, true
		}
	}
	return 0, false
}

// langAlphabet returns a deterministic scramble alphabet for a language. Each
// alphabet is a 26-rune mapping for 'a'..'z'. The mappings below are a
// lightweight Go analogue of the C LANG_DATA alphabets loaded from file on
// full servers — a running game without the language data files still gets
// visually distinct gibberish per tongue. When no alphabet is registered the
// generic fallback alphabet is used (mirrors C get_lang("default")).
var langAlphabet = map[uint32]string{
	types.LANG_ELVEN:    "aethloriendnshvquemptubfxzygcjkw",
	types.LANG_DWARVEN:  "okthdurgmonthreibalklvzjpnqyxfsw",
	types.LANG_ORCISH:   "grukznhlargthabcdefijmopqstuvwxy",
	types.LANG_TROLLISH: "brugghnomthraklgzdfijpqstuvwxyce",
	types.LANG_GOBLIN:   "snikgribzkahrtheldwumpvfjqxcyo",
	types.LANG_PIXIE:    "iloelwyphfnasbuntgdrmvckjqxz",
	types.LANG_OGRE:     "ughgrokbshntoarelmdfwpcijvxyzq",
	types.LANG_HALFLING: "mefirloakewntshbdpgquvcjzxy",
	types.LANG_DRAGON:   "xzqvktharinemolybsdfgjpucw",
	types.LANG_GITH:     "ssthkrvnzmxqpoluebidcagfjwy",
	types.LANG_GNOME:    "nfibltgarhepounsckmdvqjxywz",
}

// defaultAlphabet is used when the speaker's tongue has no registered table.
const defaultAlphabet = "abcdefghijklmnopqrstuvwxyz"

// Translate returns the argument as it would be heard by listener when
// speaker is speaking speaker.Speaking. If the listener shares the language
// (bit is in listener.Speaks) or the language is LANG_COMMON, the original
// text is returned unchanged. Otherwise letters are substituted through a
// deterministic per-language alphabet, preserving whitespace, punctuation
// and case. Color-code sequences ("&X" / "&[...]" style) are preserved
// byte-for-byte so the output still renders correctly to the client.
//
// Immortals always understand every tongue (LANG_COMMON is the "universal"
// passthrough — mirrors C's translate() shortcut at percent > 99).
func Translate(text string, speaker, listener *types.CharData) string {
	if speaker == nil || listener == nil || text == "" {
		return text
	}
	lang := uint32(speaker.Speaking)
	if lang == 0 || lang == types.LANG_COMMON {
		return text
	}
	if listener.IsImmortal() {
		return text
	}
	if uint32(listener.Speaks)&lang != 0 {
		return text
	}

	alpha := langAlphabet[lang]
	if alpha == "" {
		alpha = defaultAlphabet
	}
	// Pad or truncate to exactly 26 runes; fall back to default for missing.
	runes := []rune(alpha)
	for len(runes) < 26 {
		runes = append(runes, rune('a'+len(runes)%26))
	}
	runes = runes[:26]

	var b strings.Builder
	b.Grow(len(text))
	in := []rune(text)
	for i := 0; i < len(in); i++ {
		r := in[i]
		// Preserve "&X" colour codes verbatim.
		if r == '&' && i+1 < len(in) {
			b.WriteRune(r)
			b.WriteRune(in[i+1])
			i++
			continue
		}
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(runes[r-'a'])
		case r >= 'A' && r <= 'Z':
			m := runes[unicode.ToLower(r)-'a']
			b.WriteRune(unicode.ToUpper(m))
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
