package game

import (
	"fmt"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
)

// formatPrompt expands prompt tokens for a character.
// Supported tokens:
//
//	%h = current hp,     %H = max hp
//	%m = current mana,   %M = max mana
//	%v = current move,   %V = max move
//	%g = gold
//	%a = alignment
//	%x = experience
//	%X = xp to next level (stub: shows 0)
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
			b.WriteString("0") // TODO: XP to next level
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
