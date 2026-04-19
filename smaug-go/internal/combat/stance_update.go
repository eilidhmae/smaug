package combat

import "github.com/eilidhmae/smaug/internal/types"

// UpdateStances toggles ch's stance R/I/S bitsets against the current
// stance's Resist/Immune/Suscept values. Mirrors C update_stances at
// src/stances.c:131-146. `entering` = TRUE on stance entry; FALSE on
// stance exit — idempotent when entry/exit are paired.
//
// Both CharData.Stance* fields and StanceInfo.{Resist,Immune,Suscept}
// are plain `int` (CharData: character.go:128-130; StanceInfo: D1), so
// no RIS-constant cast is required inside this function. Callers that
// seed StanceInfo directly from `types.RIS_FIRE` etc. (which are
// `uint32`) must convert with `int(types.RIS_FIRE)`.
func UpdateStances(ch *types.CharData, entering bool) {
	if ch == nil {
		return
	}
	if ch.Stance < 0 || ch.Stance >= types.MAX_STANCE {
		return
	}
	info := StanceIndex[ch.Stance]
	if entering {
		ch.StanceResistant |= info.Resist
		ch.StanceImmune |= info.Immune
		ch.StanceSusceptible |= info.Suscept
	} else {
		ch.StanceResistant &^= info.Resist
		ch.StanceImmune &^= info.Immune
		ch.StanceSusceptible &^= info.Suscept
	}
}
