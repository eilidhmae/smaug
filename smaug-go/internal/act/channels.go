// Communication-channel commands: immtalk, gtell.
//
// See plan-channels.md. Each handler open-codes its descriptor/character walk
// because audience predicates diverge too much (imm-only / group-only /
// deaf+trust) to share a helper yet.
package act

import (
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

// DoImmtalk — the immortal-only chat channel. Alias `:`. C: `do_immtalk`
// (`act_comm.c:1327`) dispatches to `talk_channel(..., CHANNEL_IMMTALK,
// "immtalk")`. Level 51 gate applied here and at registry time.
func DoImmtalk(ch *types.CharData, argument string) {
	if argument == "" {
		ch.Send("Immtalk what?\n\r")
		return
	}
	// Mortal block. Match C's `NOT_AUTHED` fallthrough path plus the
	// implicit level gate on the broadcast (receivers must be Trust >= 51,
	// but senders must also be immortal — a mortal who wandered in via
	// alias/prefix match here just gets "Huh?").
	if !ch.IsImmortal() {
		ch.Send("Huh?\n\r")
		return
	}
	// PLR_SILENCE sender gate (C `act_comm.c:500`).
	if !ch.IsNPC() && ch.Act.IsSet(types.PLR_SILENCE) {
		ch.Send("You can't immtalk.\n\r")
		return
	}
	// Deaf sender: C `act_comm.c:505-513` blocks the sender with the
	// diagnostic and returns. The `xREMOVE_BIT` at line 514 is unreachable
	// when this branch fires — the deaf flag is NOT cleared here.
	if ch.Deaf.IsSet(types.CHANNEL_IMMTALK) {
		ch.Send("You don't have the immtalk channel turned on. To turn it on, use the Channels command.\n\r")
		return
	}

	// Self-echo and broadcast. Color codes match DoClantalk style
	// (&Y/&G/&D); per-AT_ fidelity is a separate follow-up.
	ch.Sendf("&YYou immtalk '&G%s&Y'&D\n\r", argument)

	for _, d := range WorldRef.Descriptors {
		if d == nil || d.Connected != types.CON_PLAYING || d.Character == nil {
			continue
		}
		other := d.Character
		if other == ch {
			continue
		}
		if other.GetTrust() < types.LEVEL_IMMORTAL {
			continue
		}
		if other.Deaf.IsSet(types.CHANNEL_IMMTALK) {
			continue
		}
		other.Sendf("&Y%s&G>&Y %s&D\n\r", ch.Name, argument)
	}
}

// DoGtell — group tell. Alias `;`. C: `do_gtell` (`act_comm.c:4217`).
// Walks the character list (sleepers included) and delivers to everyone for
// whom `IsSameGroup(gch, ch)` returns true. No trust gate, no deaf filter —
// group is the access control here.
func DoGtell(ch *types.CharData, argument string) {
	if argument == "" {
		ch.Send("Tell your group what?\n\r")
		return
	}
	// C gates on PLR_NO_TELL (act_comm.c:4237) for the sender.
	if !ch.IsNPC() && ch.Act.IsSet(types.PLR_NO_TELL) {
		ch.Send("Your message didn't get through!\n\r")
		return
	}

	for _, gch := range WorldRef.Characters {
		if gch == nil {
			continue
		}
		if !handler.IsSameGroup(gch, ch) {
			continue
		}
		// ch.Send / gch.Send are nil-Desc safe (see CharData.Send), so
		// disconnected-but-loaded chars won't panic. In Go's port
		// WorldRef.Characters already drops chars on disconnect, so
		// this is a no-op for that branch — left in for safety.
		gch.Sendf("%s tells the group '%s'\n\r", ch.Name, argument)
	}
}
