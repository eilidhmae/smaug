package act

import (
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// DoBank handles the bank command: balance, deposit, withdraw.
func DoBank(ch *types.CharData, argument string) {
	if ch.IsNPC() {
		return
	}

	arg1, rest := util.OneArgument(argument)

	// Check for a banker in the room.
	hasBanker := false
	if ch.InRoom != nil {
		for _, p := range ch.InRoom.People {
			if p.Act.IsSet(types.ACT_BANKER) {
				hasBanker = true
				break
			}
		}
	}

	if !hasBanker {
		ch.Sendf("You can't do that here.\n\r")
		return
	}

	switch strings.ToLower(arg1) {
	case "":
		ch.Sendf("Syntax: bank balance|deposit|withdraw <amount>\n\r")

	case "balance":
		ch.Sendf("Your bank balance is %d gold.\n\r", ch.PCData.GBalance)

	case "deposit":
		arg2, _ := util.OneArgument(rest)
		var amount int
		if strings.EqualFold(arg2, "all") {
			amount = ch.Gold
		} else {
			var err error
			amount, err = strconv.Atoi(arg2)
			if err != nil {
				ch.Sendf("How much do you want to deposit?\n\r")
				return
			}
		}
		if amount <= 0 {
			ch.Sendf("How much do you want to deposit?\n\r")
			return
		}
		if amount > ch.Gold {
			ch.Sendf("You don't have that much gold.\n\r")
			return
		}
		ch.Gold -= amount
		ch.PCData.GBalance += amount
		ch.Sendf("You deposit %d gold. Your new balance is %d gold.\n\r", amount, ch.PCData.GBalance)

	case "withdraw":
		arg2, _ := util.OneArgument(rest)
		var amount int
		if strings.EqualFold(arg2, "all") {
			amount = ch.PCData.GBalance
		} else {
			var err error
			amount, err = strconv.Atoi(arg2)
			if err != nil {
				ch.Sendf("How much do you want to withdraw?\n\r")
				return
			}
		}
		if amount <= 0 {
			ch.Sendf("How much do you want to withdraw?\n\r")
			return
		}
		if amount > ch.PCData.GBalance {
			ch.Sendf("You don't have that much gold in your account.\n\r")
			return
		}
		ch.PCData.GBalance -= amount
		ch.Gold += amount
		ch.Sendf("You withdraw %d gold. Your new balance is %d gold.\n\r", amount, ch.PCData.GBalance)

	default:
		ch.Sendf("Syntax: bank balance|deposit|withdraw <amount>\n\r")
	}
}
