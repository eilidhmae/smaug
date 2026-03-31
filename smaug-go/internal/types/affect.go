package types

// AffectData represents a temporary spell/condition effect on a character, object, or room.
// Maps to C struct affect_data (mud.h:1648).
type AffectData struct {
	Type      int       // Spell type number (skill table index)
	Duration  int       // Pulses remaining (-1 = permanent)
	Location  int       // Stat modified (APPLY_STR, APPLY_AC, etc.)
	Modifier  int       // Adjustment amount
	BitVector BitVector // Condition flags (AFF_BLIND, etc.)
}

// SmaugAff represents a SMAUG spell affect definition.
// Maps to C struct smaug_affect (mud.h:1663).
type SmaugAff struct {
	Duration  string // Dice formula for duration
	Location  int    // APPLY_* type
	Modifier  string // Dice formula for modifier
	BitVector int    // Bit number (not a full bitvector)
}
