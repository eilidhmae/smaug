package game

import (
	"github.com/eilidhmae/smaug/internal/act"
	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/mudprog"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// lastHour caches the last in-game hour so TrigHour/TrigTime only fire on
// hour transitions. Package-level so the game loop can advance it.
var lastHour = -1

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

// weatherUpdate advances the in-game clock by one hour per PULSE_TICK and
// fires HOUR/TIME mudprogs on hour boundaries. Mirrors the tick-based clock
// advancement in C's weather_update/time_update.
func (g *GameLoop) weatherUpdate() {
	g.world.TimeInfo.Hour++
	if g.world.TimeInfo.Hour >= 24 {
		g.world.TimeInfo.Hour = 0
		g.world.TimeInfo.Day++
	}
	hour := g.world.TimeInfo.Hour
	if hour == lastHour {
		return
	}
	lastHour = hour
	mudprog.TrigHour(hour, g.world)
	mudprog.TrigTime(hour, g.world)
	mudprog.RprogHourTrigger(hour, g.world)
	mudprog.RprogTimeTrigger(hour, g.world)
}

// roomRandomUpdate fires RAND progs on every loaded room that carries any
// MudProgs. Cheap guard: skip rooms with no progs entirely. Called once per
// violence pulse by the game loop.
func (g *GameLoop) roomRandomUpdate() {
	if g.world == nil {
		return
	}
	for _, room := range g.world.Rooms {
		if room == nil || len(room.MudProgs) == 0 {
			continue
		}
		mudprog.RprogRandomTrigger(room)
	}
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

// objUpdate runs per-tick object updates: timers, corpse decay, obj-prog
// RAND/RANDIW triggers.
func (g *GameLoop) objUpdate() {
	// Fire RAND/RANDIW obj-progs first (C matches: oprog_random_trigger is
	// at the very top of obj_update, before timer decay). Use a snapshot in
	// case a prog extracts the obj mid-iteration.
	objs := make([]*types.ObjData, len(g.world.Objects))
	copy(objs, g.world.Objects)
	for _, obj := range objs {
		if obj == nil {
			continue
		}
		mudprog.OprogRandomTrigger(obj)
	}

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

		// Hunting: move toward target
		if ch.Hunting != nil && ch.Hunting.Who != nil {
			act.HuntVictim(g.world, ch)
			continue
		}

		// NPC scavenging: pick up valuable items from room
		if ch.Act.IsSet(types.ACT_SCAVENGER) && ch.InRoom != nil && len(ch.InRoom.Contents) > 0 {
			var bestObj *types.ObjData
			bestCost := 0
			for _, obj := range ch.InRoom.Contents {
				if obj.WearFlags&int(types.ITEM_TAKE) != 0 && obj.GoldCost > bestCost {
					bestObj = obj
					bestCost = obj.GoldCost
				}
			}
			if bestObj != nil && util.NumberBits(2) == 0 {
				handler.ObjFromRoom(bestObj)
				bestObj.CarriedBy = ch
				bestObj.WearLoc = types.WEAR_NONE
				ch.Carrying = append(ch.Carrying, bestObj)
			}
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

// aggrUpdate checks for aggressive NPCs that should attack nearby players.
func (g *GameLoop) aggrUpdate() {
	for _, ch := range g.world.Characters {
		if !ch.IsNPC() || ch.InRoom == nil || ch.Fighting != nil {
			continue
		}
		if !ch.Act.IsSet(types.ACT_AGGRESSIVE) {
			continue
		}
		if ch.AffectedBy.IsSet(types.AFF_CHARM) {
			continue
		}
		if ch.Position < types.POS_STANDING {
			continue
		}
		if ch.InRoom.RoomFlags.IsSet(types.ROOM_SAFE) {
			continue
		}

		// Find a victim in the room
		for _, victim := range ch.InRoom.People {
			if victim == ch || victim.IsNPC() {
				continue
			}
			if victim.Level >= types.LEVEL_IMMORTAL {
				continue
			}
			if victim.Fighting != nil {
				continue
			}
			// Random chance to not attack every pulse
			if util.NumberBits(1) == 0 {
				continue
			}

			combat.StartFighting(ch, victim)
			break
		}
	}
}

// violenceUpdate runs per-pulse combat rounds.
func (g *GameLoop) violenceUpdate() {
	combat.ViolenceUpdate(g.world)
}

// autosave saves all connected player characters periodically.
func (g *GameLoop) autosave() {
	for _, d := range g.world.Descriptors {
		if d.Character != nil && !d.Character.IsNPC() && d.Character.PCData != nil && d.Character.Level >= 2 {
			g.SavePlayer(d.Character)
		}
	}
}

// areaUpdate runs periodic area resets.
func (g *GameLoop) areaUpdate() {
	for _, area := range g.world.Areas {
		if area == nil {
			continue
		}

		area.Age++

		resetFreq := area.ResetFrequency
		if resetFreq == 0 {
			resetFreq = 15 // default 15 minutes
		}

		// Don't reset if age hasn't reached frequency and players are present
		if area.Age < resetFreq {
			continue
		}

		// Reset when: no players and age >= freq, OR age >= freq*2 (forced)
		if (area.NPlayer == 0 && area.Age >= resetFreq) || area.Age >= resetFreq*2 {
			handler.ResetArea(g.world, area)
			area.Age = 0

			// Send reset message to players in the area
			if area.ResetMsg != "" {
				for _, d := range g.world.Descriptors {
					if d.Character != nil && d.Character.InRoom != nil &&
						d.Character.InRoom.Area == area {
						d.Character.Sendf("%s\n\r", area.ResetMsg)
					}
				}
			}
		}
	}
}
