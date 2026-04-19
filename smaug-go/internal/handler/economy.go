package handler

import "github.com/eilidhmae/smaug/internal/types"

// EconomyBillion mirrors the 1e9 chunk size used by C boost/lower_economy.
const EconomyBillion = 1000000000

// BoostEconomy mirrors C boost_economy (src/handler.c:5641): each whole
// billion becomes a +1 to HighEconomy, the remainder accumulates in
// LowEconomy, and any LowEconomy overflow rolls forward into HighEconomy.
//
// Nil-safe — a nil area is a no-op (auction code feeds seller-in-no-room
// into this path when closeDescriptor has already dropped the room
// reference).
func BoostEconomy(area *types.AreaData, gold int) {
	if area == nil {
		return
	}
	for gold >= EconomyBillion {
		area.HighEconomy++
		gold -= EconomyBillion
	}
	area.LowEconomy += gold
	for area.LowEconomy >= EconomyBillion {
		area.HighEconomy++
		area.LowEconomy -= EconomyBillion
	}
}

// LowerEconomy mirrors C lower_economy (src/handler.c:5660): each whole
// billion decrements HighEconomy, then the remainder is subtracted from
// LowEconomy, borrowing from HighEconomy when LowEconomy would go negative.
func LowerEconomy(area *types.AreaData, gold int) {
	if area == nil {
		return
	}
	for gold >= EconomyBillion {
		area.HighEconomy--
		gold -= EconomyBillion
	}
	area.LowEconomy -= gold
	for area.LowEconomy < 0 {
		area.HighEconomy--
		area.LowEconomy += EconomyBillion
	}
}
