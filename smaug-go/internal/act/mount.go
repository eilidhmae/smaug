package act

import (
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// DoMount implements the 'mount' command.
func DoMount(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Mount what?\n\r")
		return
	}

	if ch.Mount != nil {
		ch.Send("You are already mounted.\n\r")
		return
	}

	mount := handler.GetCharRoom(ch, arg)
	if mount == nil {
		ch.Send("They aren't here.\n\r")
		return
	}

	if !mount.IsNPC() {
		ch.Send("You can't mount players.\n\r")
		return
	}

	if !mount.Act.IsSet(types.ACT_MOUNTABLE) {
		ch.Send("That creature cannot be mounted.\n\r")
		return
	}

	ch.Mount = mount
	mount.Mount = ch
	ch.Position = types.POS_MOUNTED
	ch.Sendf("You mount %s.\n\r", mount.ShortDescr)
}

// DoDismount implements the 'dismount' command.
func DoDismount(ch *types.CharData, argument string) {
	if ch.Mount == nil {
		ch.Send("You aren't mounted.\n\r")
		return
	}

	mount := ch.Mount
	ch.Mount = nil
	mount.Mount = nil
	ch.Position = types.POS_STANDING
	ch.Sendf("You dismount %s.\n\r", mount.ShortDescr)
}
