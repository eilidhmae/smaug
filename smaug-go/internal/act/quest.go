package act

import (
	"math/rand"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// DoQuest handles the quest command: request, complete, list, buy, info, time, points.
func DoQuest(ch *types.CharData, argument string) {
	cmd, rest := util.OneArgument(argument)

	switch cmd {
	case "":
		ch.Send("Syntax: quest request|complete|list|buy|info|time|points\n\r")

	case "points":
		ch.Sendf("You have %d quest points.\n\r", ch.QuestPoints)

	case "time":
		if ch.Countdown > 0 {
			ch.Sendf("You have %d ticks to complete your quest.\n\r", ch.Countdown)
		} else if ch.NextQuest > 0 {
			ch.Sendf("You must wait %d ticks before requesting a new quest.\n\r", ch.NextQuest)
		} else {
			ch.Send("You are not on a quest and may request one.\n\r")
		}

	case "info":
		if ch.QuestMob > 0 {
			ch.Sendf("You are hunting mob vnum %d.\n\r", ch.QuestMob)
		} else if ch.QuestObj > 0 {
			ch.Sendf("You are seeking object vnum %d.\n\r", ch.QuestObj)
		} else {
			ch.Send("You are not on a quest.\n\r")
		}

	case "request":
		qm := findQuestmaster(ch)
		if qm == nil {
			ch.Send("You can't do that here.\n\r")
			return
		}
		if ch.Countdown > 0 {
			ch.Send("You are already on a quest.\n\r")
			return
		}
		if ch.NextQuest > 0 {
			ch.Send("You must wait before requesting another quest.\n\r")
			return
		}
		generateQuest(ch)

	case "complete":
		qm := findQuestmaster(ch)
		if qm == nil {
			ch.Send("You can't do that here.\n\r")
			return
		}
		if ch.QuestMob > 0 {
			ch.Send("You haven't completed your quest yet.\n\r")
			return
		}
		if ch.QuestMob == -1 {
			reward := 30 + rand.Intn(20)
			ch.QuestPoints += reward
			ch.Sendf("Congratulations! You receive %d quest points.\n\r", reward)
			ch.QuestMob = 0
			ch.QuestObj = 0
			ch.Countdown = 0
			ch.NextQuest = 15
			ch.QuestGiver = nil
			return
		}
		ch.Send("You are not on a quest.\n\r")

	case "list":
		qm := findQuestmaster(ch)
		if qm == nil {
			ch.Send("You can't do that here.\n\r")
			return
		}
		ch.Send("Available quest rewards:\n\r" +
			"  500 qp - 10000 gold\n\r" +
			"  250 qp - 5 practices\n\r" +
			"  1000 qp - +10 max hp\n\r" +
			"  1000 qp - +10 max mana\n\r")

	case "buy":
		qm := findQuestmaster(ch)
		if qm == nil {
			ch.Send("You can't do that here.\n\r")
			return
		}
		rewardName, _ := util.OneArgument(rest)
		questBuy(ch, rewardName)

	default:
		ch.Send("Syntax: quest request|complete|list|buy|info|time|points\n\r")
	}
}

// findQuestmaster looks for an NPC with ACT_QUESTMASTER in the character's room.
func findQuestmaster(ch *types.CharData) *types.CharData {
	if ch.InRoom == nil {
		return nil
	}
	for _, p := range ch.InRoom.People {
		if p.Act.IsSet(types.ACT_QUESTMASTER) {
			return p
		}
	}
	return nil
}

// questBuy handles purchasing rewards with quest points.
func questBuy(ch *types.CharData, rewardName string) {
	type reward struct {
		cost int
		apply func()
		msg  string
	}

	rewards := map[string]reward{
		"gold": {
			cost:  500,
			apply: func() { ch.Gold += 10000 },
			msg:   "You receive 10000 gold coins!\n\r",
		},
		"practices": {
			cost:  250,
			apply: func() { ch.Practice += 5 },
			msg:   "You receive 5 practices!\n\r",
		},
		"hp": {
			cost:  1000,
			apply: func() { ch.MaxHit += 10; ch.Hit += 10 },
			msg:   "Your maximum hit points increase by 10!\n\r",
		},
		"mana": {
			cost:  1000,
			apply: func() { ch.MaxMana += 10; ch.Mana += 10 },
			msg:   "Your maximum mana increases by 10!\n\r",
		},
	}

	r, ok := rewards[rewardName]
	if !ok {
		ch.Send("You can buy: gold, practices, hp, mana.\n\r")
		return
	}

	if ch.QuestPoints < r.cost {
		ch.Sendf("You do not have enough quest points. You need %d.\n\r", r.cost)
		return
	}

	ch.QuestPoints -= r.cost
	r.apply()
	ch.Send(r.msg)
}

// generateQuest picks a random mob within level range and assigns a kill quest.
func generateQuest(ch *types.CharData) {
	w := WorldRef
	if w == nil {
		ch.Send("No suitable quests available right now.\n\r")
		return
	}

	minLevel := ch.Level - 5
	maxLevel := ch.Level + 5

	var candidates []*types.MobIndexData
	for _, mob := range w.MobIndex {
		if mob.Level >= minLevel && mob.Level <= maxLevel {
			candidates = append(candidates, mob)
		}
	}

	if len(candidates) == 0 {
		ch.Send("No suitable quests available right now.\n\r")
		return
	}

	target := candidates[rand.Intn(len(candidates))]
	ch.QuestMob = target.Vnum
	ch.Countdown = 30
	ch.Sendf("Your quest: slay %s (vnum %d)! You have 30 ticks.\n\r", target.ShortDescr, target.Vnum)
}

// QuestUpdate decrements quest countdown timers for all players.
// Called once per tick from the game update loop.
func QuestUpdate(w *world.World) {
	for _, ch := range w.Characters {
		if ch.IsNPC() {
			continue
		}
		if ch.Countdown > 0 {
			ch.Countdown--
			if ch.Countdown == 0 {
				ch.Send("You have failed your quest!\n\r")
				ch.QuestMob = 0
				ch.QuestObj = 0
				ch.QuestGiver = nil
			}
		}
		if ch.NextQuest > 0 {
			ch.NextQuest--
		}
	}
}
