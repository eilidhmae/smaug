// Package combat implements the SMAUG combat system.
package combat

import (
	"fmt"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// Return codes from damage/combat functions.
const (
	rNONE      = 0
	rVICT_DIED = 2
	rCHAR_DIED = 1
	rBOTH_DIED = 3
)

// StartFighting initiates combat between ch and victim.
func StartFighting(ch *types.CharData, victim *types.CharData) {
	if ch.Fighting != nil {
		return
	}
	ch.Fighting = &types.FightData{
		Who: victim,
	}
	ch.NumFighting = 1
	ch.Position = types.POS_FIGHTING
}

// StopFighting ends combat for a character. If fBoth is true, also stops
// anyone fighting ch.
func StopFighting(ch *types.CharData, fBoth bool) {
	ch.Fighting = nil
	ch.NumFighting = 0
	if ch.Position == types.POS_FIGHTING {
		ch.Position = types.POS_STANDING
	}

	if fBoth {
		// Find anyone fighting ch and stop them too
		if ch.InRoom != nil {
			for _, rch := range ch.InRoom.People {
				if rch.Fighting != nil && rch.Fighting.Who == ch {
					rch.Fighting = nil
					rch.NumFighting = 0
					if rch.Position == types.POS_FIGHTING {
						rch.Position = types.POS_STANDING
					}
				}
			}
		}
	}
}

// ViolenceUpdate processes one combat round for all fighting characters.
func ViolenceUpdate(w *world.World) {
	for _, ch := range w.Characters {
		if ch.Fighting == nil || ch.InRoom == nil {
			continue
		}

		// Can't fight if incapacitated or worse, or at 0 HP
		if ch.Hit <= 0 || ch.Position <= types.POS_INCAP {
			StopFighting(ch, false)
			continue
		}

		victim := ch.Fighting.Who
		if victim == nil || victim.InRoom != ch.InRoom {
			StopFighting(ch, false)
			continue
		}

		// One attack per violence pulse
		OneHit(w, ch, victim, types.TYPE_UNDEFINED)

		// Multi-attack for NPCs
		if ch.Fighting != nil && ch.IsNPC() && ch.NumAttacks > 1 {
			for i := 1; i < ch.NumAttacks; i++ {
				if ch.Fighting == nil {
					break
				}
				OneHit(w, ch, victim, types.TYPE_UNDEFINED)
			}
		}

		// Dual wield extra attack
		if ch.Fighting != nil && handler.GetEqChar(ch, types.WEAR_DUAL_WIELD) != nil {
			OneHit(w, ch, victim, types.TYPE_UNDEFINED)
		}

		// Wimpy auto-flee check
		if ch.Fighting != nil && !ch.IsNPC() && ch.Wimpy > 0 && ch.Hit <= ch.Wimpy {
			ch.Send("You wimp out and attempt to flee!\n\r")
			// Try each direction
			for dir := 0; dir <= types.DIR_DOWN; dir++ {
				if ch.InRoom == nil {
					break
				}
				exit := ch.InRoom.GetExit(dir)
				if exit != nil && exit.ToRoom != nil && exit.ExitInfo&int(types.EX_CLOSED) == 0 {
					StopFighting(ch, true)
					handler.CharFromRoom(ch)
					handler.CharToRoom(ch, exit.ToRoom)
					ch.Sendf("You flee %s!\n\r", dirName(dir))
					break
				}
			}
		}
	}
}

// dirName returns the name of a direction.
func dirName(dir int) string {
	names := []string{"north", "east", "south", "west", "up", "down",
		"northeast", "northwest", "southeast", "southwest"}
	if dir >= 0 && dir < len(names) {
		return names[dir]
	}
	return "somewhere"
}

// OneHit resolves a single attack from ch against victim.
func OneHit(w *world.World, ch *types.CharData, victim *types.CharData, dt int) {
	if victim.Hit <= 0 || ch.InRoom != victim.InRoom {
		return
	}

	// Find wielded weapon
	wield := handler.GetEqChar(ch, types.WEAR_WIELD)

	// Calculate thac0
	var thac0 int
	if ch.IsNPC() {
		thac0 = ch.MobThac0
	} else {
		// Player thac0: interpolate from class table (simplified)
		thac0 = 20 - ch.Level
	}
	thac0 -= ch.Hitroll
	thac0 -= types.StrApp[ch.GetCurrStr()].ToHit

	// Victim AC (capped at -19)
	victimAC := util.UMAX(-19, victim.Armor/10)

	// Roll d20
	diceroll := rollD20()

	// Hit/miss check
	if diceroll == 0 || (diceroll != 19 && diceroll < thac0-victimAC) {
		// Miss
		Damage(w, ch, victim, 0, dt)
		return
	}

	// Calculate damage
	var dam int
	if wield != nil {
		dam = util.NumberRange(wield.Value[1], wield.Value[2])
	} else {
		if ch.BareNumDie > 0 && ch.BareSizeDie > 0 {
			dam = util.DiceRoll(ch.BareNumDie, ch.BareSizeDie)
		} else {
			dam = util.NumberRange(1, 4)
		}
	}

	// Add damroll and str bonus
	dam += ch.Damroll
	dam += types.StrApp[ch.GetCurrStr()].ToDam

	// Position multipliers (attacker)
	switch ch.Position {
	case types.POS_BERSERK:
		dam = dam * 6 / 5
	case types.POS_AGGRESSIVE:
		dam = dam * 11 / 10
	case types.POS_DEFENSIVE:
		dam = dam * 85 / 100
	case types.POS_EVASIVE:
		dam = dam * 4 / 5
	}

	// Sleeping victim = double damage
	if victim.Position < types.POS_SLEEPING {
		dam *= 2
	}

	if dam <= 0 {
		dam = 1
	}

	// Sanctuary halves damage
	if victim.AffectedBy.IsSet(types.AFF_SANCTUARY) {
		dam /= 2
	}

	Damage(w, ch, victim, dam, dt)
}

// Damage applies damage to a victim. Returns a retcode indicating if
// the victim died.
func Damage(w *world.World, ch *types.CharData, victim *types.CharData, dam int, dt int) int {
	if victim.Hit <= 0 {
		return rNONE
	}

	// Ensure fighting is set
	if ch.Fighting == nil && ch != victim {
		StartFighting(ch, victim)
	}
	if victim.Fighting == nil && victim != ch {
		StartFighting(victim, ch)
	}

	// Apply damage
	victim.Hit -= dam

	// Send damage messages
	if dam == 0 {
		ch.Sendf("You miss %s.\n\r", victim.Name)
		victim.Sendf("%s misses you.\n\r", ch.Name)
	} else {
		ch.Sendf("You hit %s for %d damage.\n\r", victim.Name, dam)
		victim.Sendf("%s hits you for %d damage.\n\r", ch.Name, dam)
	}

	// Update position and send status messages
	oldPos := victim.Position
	updatePos(victim)

	if victim.Position != oldPos && victim.Position <= types.POS_INCAP {
		switch victim.Position {
		case types.POS_INCAP:
			victim.Send("You are incapacitated and will slowly die, if not aided.\n\r")
			StopFighting(victim, true)
		case types.POS_STUNNED:
			victim.Send("You are stunned, but will probably recover.\n\r")
			StopFighting(victim, true)
		}
	}

	// Death check
	if victim.Position == types.POS_DEAD {
		StopFighting(ch, true)

		// XP gain for killer (player killing NPC)
		if !ch.IsNPC() && victim.IsNPC() {
			xpGain := computeXP(ch, victim)
			ch.Exp += xpGain
			ch.Sendf("You receive %d experience points.\n\r", xpGain)
		}

		if victim.IsNPC() {
			MakeCorpse(w, victim)
			handler.ExtractChar(w, victim, true)
		} else {
			// PC death: reset to resting with 1 HP
			victim.Hit = 1
			victim.Mana = 1
			victim.Move = 1
			victim.Position = types.POS_RESTING
			victim.Send("You have been KILLED!\n\r")
			// Full death handling (corpse, XP loss, etc.) will be expanded later
		}

		return rVICT_DIED
	}

	return rNONE
}

// MakeCorpse creates a corpse object from a dead character and transfers
// their inventory into it.
func MakeCorpse(w *world.World, ch *types.CharData) {
	if ch.InRoom == nil {
		return
	}

	var corpse *types.ObjData

	if ch.IsNPC() {
		corpse = &types.ObjData{
			Name:        fmt.Sprintf("corpse %s", ch.Name),
			ShortDescr:  fmt.Sprintf("the corpse of %s", ch.ShortDescr),
			Description: fmt.Sprintf("The corpse of %s is lying here.", ch.ShortDescr),
			ItemType:    types.ITEM_CORPSE_NPC,
			Timer:       6,
			Weight:      ch.Weight,
		}
	} else {
		corpse = &types.ObjData{
			Name:        fmt.Sprintf("corpse %s", ch.Name),
			ShortDescr:  fmt.Sprintf("the corpse of %s", ch.Name),
			Description: fmt.Sprintf("The corpse of %s is lying here.", ch.Name),
			ItemType:    types.ITEM_CORPSE_PC,
			Timer:       40,
			Weight:      ch.Weight,
		}
	}

	w.AddObj(corpse)

	// Transfer inventory to corpse
	for len(ch.Carrying) > 0 {
		obj := ch.Carrying[len(ch.Carrying)-1]
		handler.ObjFromChar(obj)
		handler.ObjToObj(obj, corpse)
	}

	// Store gold in corpse value[0] for looting
	if ch.Gold > 0 {
		corpse.Value[0] = ch.Gold
		ch.Gold = 0
	}

	handler.ObjToRoom(corpse, ch.InRoom)
}

// updatePos sets position based on current HP.
func updatePos(ch *types.CharData) {
	if ch.Hit <= -ch.MaxHit {
		ch.Position = types.POS_DEAD
	} else if ch.Hit <= -ch.MaxHit/2 {
		ch.Position = types.POS_STUNNED
	} else if ch.Hit <= 0 {
		ch.Position = types.POS_INCAP
	}
}

// rollD20 returns a value 0-19 (equivalent to C's number_bits(5) clamped < 20).
func rollD20() int {
	for {
		v := util.NumberBits(5)
		if v < 20 {
			return v
		}
	}
}

// computeXP calculates XP gained for killing a victim.
// Based on victim's base XP scaled by level difference.
func computeXP(ch *types.CharData, victim *types.CharData) int {
	xp := victim.Exp
	if xp <= 0 {
		// Fallback: base XP from level
		xp = victim.Level * victim.Level * 10
	}

	// Level difference scaling
	diff := victim.Level - ch.Level
	switch {
	case diff >= 5:
		xp = xp * 3 / 2 // 150%
	case diff >= 1:
		xp = xp * 11 / 10 // 110%
	case diff >= -3:
		// Same range, no modifier
	case diff >= -8:
		xp = xp * 3 / 4 // 75%
	default:
		xp = xp / 4 // 25% for very low level mobs
	}

	if xp < 1 {
		xp = 1
	}
	return xp
}

// AFF_SANCTUARY constant check.
func init() {
	// Verify AFF_SANCTUARY exists
	_ = types.AFF_SANCTUARY
}
