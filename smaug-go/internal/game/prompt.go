package game

import (
	"fmt"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
)

// PromptExpBase, when non-nil, returns the per-character exp-per-level
// base used by the %X prompt token (XP-to-next-level). It mirrors C's
// `get_exp_base(ch)` (handler.c:107-112): 1000 for NPCs, class table's
// ExpBase for PCs. Set at boot from the world's Classes slice; unset
// in standalone unit tests so %X falls back to 0.
//
// Written once at boot before the game loop starts; read only from the
// game loop goroutine. Safe without synchronization.
var PromptExpBase func(ch *types.CharData) int

// expToLevel returns how much XP is required to reach `level` for this
// character. Mirrors C handler.c:117-124:
//
//	lvl = UMAX(0, level - 1);
//	return lvl * lvl * lvl * get_exp_base(ch);
//
// If PromptExpBase is nil (tests), returns 0 — the %X fallback case.
func expToLevel(ch *types.CharData, level int) int {
	if PromptExpBase == nil {
		return 0
	}
	lvl := level - 1
	if lvl < 0 {
		lvl = 0
	}
	return lvl * lvl * lvl * PromptExpBase(ch)
}

// formatPrompt expands prompt tokens for a character.
// Supported tokens:
//
//	%h = current hp,     %H = max hp
//	%m = current mana,   %M = max mana
//	%v = current move,   %V = max move
//	%g = gold
//	%a = alignment
//	%x = current experience
//	%X = xp to next level (0 if PromptExpBase seam is not wired)
//	%r = room name
//	%% = literal %
func FormatPrompt(ch *types.CharData) string {
	if ch.PCData == nil || ch.PCData.Prompt == "" {
		return fmt.Sprintf("<%dhp %dm %dmv> ", ch.Hit, ch.Mana, ch.Move)
	}

	prompt := ch.PCData.Prompt
	var b strings.Builder
	b.Grow(len(prompt) * 2)

	for i := 0; i < len(prompt); i++ {
		if prompt[i] != '%' {
			b.WriteByte(prompt[i])
			continue
		}
		i++
		if i >= len(prompt) {
			b.WriteByte('%')
			break
		}
		switch prompt[i] {
		case '%':
			b.WriteByte('%')
		case 'h':
			fmt.Fprintf(&b, "%d", ch.Hit)
		case 'H':
			fmt.Fprintf(&b, "%d", ch.MaxHit)
		case 'm':
			fmt.Fprintf(&b, "%d", ch.Mana)
		case 'M':
			fmt.Fprintf(&b, "%d", ch.MaxMana)
		case 'v':
			fmt.Fprintf(&b, "%d", ch.Move)
		case 'V':
			fmt.Fprintf(&b, "%d", ch.MaxMove)
		case 'g':
			fmt.Fprintf(&b, "%d", ch.Gold)
		case 'a':
			fmt.Fprintf(&b, "%d", ch.Alignment)
		case 'x':
			fmt.Fprintf(&b, "%d", ch.Exp)
		case 'X':
			// exp_level(ch, level+1) - ch.Exp, clamped >= 0.
			// C smaug.c:4193-4194.
			tnl := expToLevel(ch, ch.Level+1) - ch.Exp
			if tnl < 0 {
				tnl = 0
			}
			fmt.Fprintf(&b, "%d", tnl)
		case 'r':
			if ch.InRoom != nil {
				b.WriteString(ch.InRoom.Name)
			} else {
				b.WriteString("Nowhere")
			}
		default:
			b.WriteByte('%')
			b.WriteByte(prompt[i])
		}
	}

	return b.String()
}
