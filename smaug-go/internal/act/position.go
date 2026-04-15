package act

import (
	"github.com/eilidhmae/smaug/internal/mudprog"
	"github.com/eilidhmae/smaug/internal/types"
)

func isFighting(pos int) bool {
	return pos == types.POS_FIGHTING || pos == types.POS_EVASIVE ||
		pos == types.POS_DEFENSIVE || pos == types.POS_AGGRESSIVE ||
		pos == types.POS_BERSERK
}

// DoRest implements the 'rest' command.
func DoRest(ch *types.CharData, argument string) {
	oldPos := ch.Position
	switch {
	case isFighting(ch.Position):
		ch.Send("You are busy fighting!\n\r")
	case ch.Position == types.POS_MOUNTED:
		ch.Send("You'd better dismount first.\n\r")
	case ch.Position == types.POS_RESTING:
		ch.Send("You are already resting.\n\r")
	case ch.Position == types.POS_SLEEPING:
		if ch.AffectedBy.IsSet(types.AFF_SLEEP) {
			ch.Send("You can't seem to wake up!\n\r")
			return
		}
		ch.Send("You rouse from your slumber.\n\r")
		ch.Position = types.POS_RESTING
	case ch.Position == types.POS_SITTING:
		ch.Send("You lie back and sprawl out to rest.\n\r")
		ch.Position = types.POS_RESTING
	case ch.Position == types.POS_STANDING:
		ch.Send("You sprawl out haphazardly.\n\r")
		ch.Position = types.POS_RESTING
	}
	// Room-prog REST trigger fires once on the transition into POS_RESTING.
	if oldPos != types.POS_RESTING && ch.Position == types.POS_RESTING {
		mudprog.RprogRestTrigger(ch)
	}
}

// DoSit implements the 'sit' command.
func DoSit(ch *types.CharData, argument string) {
	switch {
	case isFighting(ch.Position):
		ch.Send("You are busy fighting!\n\r")
	case ch.Position == types.POS_MOUNTED:
		ch.Send("You are already sitting - on your mount.\n\r")
	case ch.Position == types.POS_SITTING:
		ch.Send("You are already sitting.\n\r")
	case ch.Position == types.POS_SLEEPING:
		if ch.AffectedBy.IsSet(types.AFF_SLEEP) {
			ch.Send("You can't seem to wake up!\n\r")
			return
		}
		ch.Send("You wake and sit up.\n\r")
		ch.Position = types.POS_SITTING
	case ch.Position == types.POS_RESTING:
		ch.Send("You stop resting and sit up.\n\r")
		ch.Position = types.POS_SITTING
	case ch.Position == types.POS_STANDING:
		ch.Send("You sit down.\n\r")
		ch.Position = types.POS_SITTING
	}
}

// DoStand implements the 'stand' command.
func DoStand(ch *types.CharData, argument string) {
	switch {
	case isFighting(ch.Position):
		ch.Send("You are already fighting!\n\r")
	case ch.Position == types.POS_STANDING:
		ch.Send("You are already standing.\n\r")
	case ch.Position == types.POS_SLEEPING:
		if ch.AffectedBy.IsSet(types.AFF_SLEEP) {
			ch.Send("You can't seem to wake up!\n\r")
			return
		}
		ch.Send("You wake and climb quickly to your feet.\n\r")
		ch.Position = types.POS_STANDING
	case ch.Position == types.POS_RESTING:
		ch.Send("You gather yourself and stand up.\n\r")
		ch.Position = types.POS_STANDING
	case ch.Position == types.POS_SITTING:
		ch.Send("You move quickly to your feet.\n\r")
		ch.Position = types.POS_STANDING
	}
}

// DoSleep implements the 'sleep' command.
func DoSleep(ch *types.CharData, argument string) {
	oldPos := ch.Position
	switch {
	case isFighting(ch.Position):
		ch.Send("You are busy fighting!\n\r")
	case ch.Position == types.POS_MOUNTED:
		ch.Send("You really should dismount first.\n\r")
	case ch.Position == types.POS_SLEEPING:
		ch.Send("You are already sleeping.\n\r")
	case ch.Position == types.POS_RESTING:
		ch.Send("You close your eyes and drift into slumber.\n\r")
		ch.Position = types.POS_SLEEPING
	case ch.Position == types.POS_SITTING:
		ch.Send("You slump over and fall dead asleep.\n\r")
		ch.Position = types.POS_SLEEPING
	case ch.Position == types.POS_STANDING:
		ch.Send("You collapse into a deep sleep.\n\r")
		ch.Position = types.POS_SLEEPING
	}
	// Room-prog SLEEP trigger fires once on the transition into POS_SLEEPING.
	if oldPos != types.POS_SLEEPING && ch.Position == types.POS_SLEEPING {
		mudprog.RprogSleepTrigger(ch)
	}
}

// DoWake implements the 'wake' command. With no argument, same as stand.
func DoWake(ch *types.CharData, argument string) {
	DoStand(ch, argument)
}
