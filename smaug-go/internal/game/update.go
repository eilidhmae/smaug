package game

import (
	"github.com/eilidhmae/smaug/internal/act"
	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// hitGain calculates HP regeneration per tick for a character.
func hitGain(ch *types.CharData) int {
	var gain int

	if ch.IsNPC() {
		gain = ch.Level * 3 / 2
	} else {
		gain = util.UMIN(5, ch.Level)

		switch ch.Position {
		case types.POS_DEAD:
			return 0
		case types.POS_MORTAL, types.POS_INCAP:
			return -1
		case types.POS_STUNNED:
			return 1
		case types.POS_SLEEPING:
			gain += ch.GetCurrCon() * 2
		case types.POS_RESTING:
			gain += ch.GetCurrCon()
		}

		// Hunger/thirst penalties
		if ch.PCData != nil {
			if ch.PCData.Condition[types.COND_FULL] == 0 {
				gain /= 2
			}
			if ch.PCData.Condition[types.COND_THIRST] == 0 {
				gain /= 2
			}
		}
	}

	if ch.AffectedBy.IsSet(types.AFF_POISON) {
		gain /= 4
	}

	return util.UMIN(gain, ch.MaxHit-ch.Hit)
}

// manaGain calculates mana regeneration per tick for a character.
func manaGain(ch *types.CharData) int {
	var gain int

	if ch.IsNPC() {
		gain = ch.Level
	} else {
		gain = util.UMIN(5, ch.Level/2)

		if ch.Position < types.POS_SLEEPING {
			return 0
		}

		switch ch.Position {
		case types.POS_SLEEPING:
			gain += ch.GetCurrInt() * 3
		case types.POS_RESTING:
			gain += ch.GetCurrInt() * 2
		}

		if ch.PCData != nil {
			if ch.PCData.Condition[types.COND_FULL] == 0 {
				gain /= 2
			}
			if ch.PCData.Condition[types.COND_THIRST] == 0 {
				gain /= 2
			}
		}
	}

	if ch.AffectedBy.IsSet(types.AFF_POISON) {
		gain /= 4
	}

	return util.UMIN(gain, ch.MaxMana-ch.Mana)
}

// moveGain calculates move regeneration per tick for a character.
func moveGain(ch *types.CharData) int {
	var gain int

	if ch.IsNPC() {
		gain = ch.Level
	} else {
		gain = util.UMAX(15, 2*ch.Level)

		switch ch.Position {
		case types.POS_DEAD:
			return 0
		case types.POS_MORTAL, types.POS_INCAP:
			return -1
		case types.POS_STUNNED:
			return 1
		case types.POS_SLEEPING:
			gain += ch.GetCurrDex() * 4
		case types.POS_RESTING:
			gain += ch.GetCurrDex() * 2
		}

		if ch.PCData != nil {
			if ch.PCData.Condition[types.COND_FULL] == 0 {
				gain /= 2
			}
			if ch.PCData.Condition[types.COND_THIRST] == 0 {
				gain /= 2
			}
		}
	}

	if ch.AffectedBy.IsSet(types.AFF_POISON) {
		gain /= 4
	}

	return util.UMIN(gain, ch.MaxMove-ch.Move)
}

// charUpdate runs per-tick character updates: regen, affect duration, hunger/thirst.
func (g *GameLoop) charUpdate() {
	for _, ch := range g.world.Characters {
		if ch.InRoom == nil {
			continue
		}

		// Incapacitated/mortal players slowly bleed out
		if ch.Position == types.POS_INCAP || ch.Position == types.POS_MORTAL {
			ch.Send("You are bleeding out...\n\r")
			ch.Hit--
			if ch.Hit <= -ch.MaxHit {
				ch.Position = types.POS_DEAD
				ch.Send("You have been KILLED!\n\r")
				// PC death: reset
				if !ch.IsNPC() {
					ch.Hit = 1
					ch.Mana = 1
					ch.Move = 1
					ch.Position = types.POS_RESTING
					ch.Send("You awaken in a daze...\n\r")
				}
			}
			continue
		}

		// Stunned characters auto-recover
		if ch.Position == types.POS_STUNNED {
			ch.Hit++
			if ch.Hit > 0 {
				ch.Position = types.POS_STANDING
				ch.Send("You regain consciousness.\n\r")
			}
		}

		// HP/mana/move regeneration
		if ch.Position >= types.POS_STUNNED {
			if ch.Hit < ch.MaxHit {
				ch.Hit += hitGain(ch)
			}
			if ch.Mana < ch.MaxMana {
				ch.Mana += manaGain(ch)
			}
			if ch.Move < ch.MaxMove {
				ch.Move += moveGain(ch)
			}
		}

		// Affect duration countdown
		for i := len(ch.Affects) - 1; i >= 0; i-- {
			aff := ch.Affects[i]
			if aff.Duration > 0 {
				aff.Duration--
			} else if aff.Duration == 0 {
				// Send wear-off message
				if aff.Type > 0 && aff.Type < len(g.world.Skills) {
					sk := g.world.Skills[aff.Type]
					if sk != nil && sk.MsgOff != "" {
						ch.Sendf("%s\n\r", sk.MsgOff)
					}
				}
				handler.AffectRemove(ch, aff)
			}
			// Duration -1 = permanent, skip
		}

		// Poison tick damage
		if ch.AffectedBy.IsSet(types.AFF_POISON) {
			ch.Send("You shiver and suffer.\n\r")
			ch.Hit -= 6
			if ch.Hit < 1 {
				ch.Hit = 1
			}
		}
	}
}

// objUpdate runs per-tick object updates: timers, corpse decay.
func (g *GameLoop) objUpdate() {
	for i := len(g.world.Objects) - 1; i >= 0; i-- {
		if i >= len(g.world.Objects) {
			continue
		}
		obj := g.world.Objects[i]

		if obj.Timer <= 0 {
			continue
		}

		obj.Timer--
		if obj.Timer > 0 {
			continue
		}

		// Timer expired — extract
		switch obj.ItemType {
		case types.ITEM_CORPSE_NPC:
			if obj.InRoom != nil {
				for _, rch := range obj.InRoom.People {
					if rch.Desc != nil {
						rch.Sendf("%s decays into dust and blows away.\n\r",
							util.Capitalize(obj.ShortDescr))
					}
				}
			}
		case types.ITEM_CORPSE_PC:
			if obj.InRoom != nil {
				for _, rch := range obj.InRoom.People {
					if rch.Desc != nil {
						rch.Sendf("%s is sucked into a swirling vortex of colors...\n\r",
							util.Capitalize(obj.ShortDescr))
					}
				}
			}
		}

		handler.ExtractObj(g.world, obj)
	}
}

// mobileUpdate runs per-pulse NPC behavior: wandering, scavenging.
func (g *GameLoop) mobileUpdate() {
	for _, ch := range g.world.Characters {
		if !ch.IsNPC() || ch.InRoom == nil {
			continue
		}

		// Only standing mobs act
		if ch.Position != types.POS_STANDING {
			continue
		}

		// Skip charmed/fighting mobs
		if ch.AffectedBy.IsSet(types.AFF_CHARM) || ch.Fighting != nil {
			continue
		}

		// Sentinel mobs don't wander
		if ch.Act.IsSet(types.ACT_SENTINEL) {
			continue
		}

		// Random wander: pick a random direction
		door := util.NumberBits(5)
		if door > 9 {
			continue // Invalid direction
		}

		exit := ch.InRoom.GetExit(door)
		if exit == nil || exit.ToRoom == nil {
			continue
		}

		// Don't walk through closed doors
		if exit.ExitInfo&int(types.EX_CLOSED) != 0 {
			continue
		}

		// Don't enter no-mob rooms
		if exit.ToRoom.RoomFlags.IsSet(types.ROOM_NO_MOB) {
			continue
		}

		// Stay-area mobs don't leave their area
		if ch.Act.IsSet(types.ACT_STAY_AREA) && exit.ToRoom.Area != ch.InRoom.Area {
			continue
		}

		act.MoveChar(ch, door)
	}
}

// violenceUpdate runs per-pulse combat rounds.
func (g *GameLoop) violenceUpdate() {
	combat.ViolenceUpdate(g.world)
}

// areaUpdate runs periodic area resets.
func (g *GameLoop) areaUpdate() {
	// Area resets are already run at boot via handler.ResetAllAreas.
	// Periodic re-resets will be added when area reset timers are implemented.
}
