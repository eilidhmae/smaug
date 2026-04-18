// Communication-channel commands: immtalk, gtell, channels.
//
// See plan-channels.md and plan-do-channels.md. Each handler open-codes its
// descriptor/character walk because audience predicates diverge too much
// (imm-only / group-only / deaf+trust) to share a helper yet.
package act

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// channelToggleTable maps the keyword accepted by `channels +<name>` /
// `channels -<name>` to the CHANNEL_* bit manipulated in ch.Deaf. Order
// matches C's str_cmp chain in act_info.c:5414-5473. Note `muse` toggles
// CHANNEL_HIGHGOD, not CHANNEL_MUSIC — C oddity preserved.
var channelToggleTable = []struct {
	name string
	bit  int
}{
	{"auction", types.CHANNEL_AUCTION},
	{"traffic", types.CHANNEL_TRAFFIC},
	{"chat", types.CHANNEL_CHAT},
	{"clan", types.CHANNEL_CLAN},
	{"council", types.CHANNEL_COUNCIL},
	{"guild", types.CHANNEL_GUILD},
	{"quest", types.CHANNEL_QUEST},
	{"tells", types.CHANNEL_TELLS},
	{"immtalk", types.CHANNEL_IMMTALK},
	{"log", types.CHANNEL_LOG},
	{"build", types.CHANNEL_BUILD},
	{"high", types.CHANNEL_HIGH},
	{"pray", types.CHANNEL_PRAY},
	{"avatar", types.CHANNEL_AVTALK},
	{"monitor", types.CHANNEL_MONITOR},
	{"death", types.CHANNEL_DEATH},
	{"auth", types.CHANNEL_AUTH},
	{"newbie", types.CHANNEL_NEWBIE},
	{"music", types.CHANNEL_MUSIC},
	{"muse", types.CHANNEL_HIGHGOD},
	{"ask", types.CHANNEL_ASK},
	{"yell", types.CHANNEL_YELL},
	{"comm", types.CHANNEL_COMM},
	{"warn", types.CHANNEL_WARN},
	{"bug", types.CHANNEL_BUG},
	{"order", types.CHANNEL_ORDER},
	{"wartalk", types.CHANNEL_WARTALK},
	{"whisper", types.CHANNEL_WHISPER},
	{"racetalk", types.CHANNEL_RACETALK},
	{"retired", types.CHANNEL_RETIRED},
}

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

// channelEntryLine emits " &G+NAME" when the deaf bit is CLEARED (channel
// ENABLED) or " &g-name" when the bit is SET. Mirrors the C
// `!xIS_SET(ch->deaf, BIT) ? " &G+NAME" : " &g-name"` pattern used throughout
// act_info.c:5289-5391.
func channelEntryLine(ch *types.CharData, name string, bit int) {
	if !ch.Deaf.IsSet(bit) {
		ch.Sendf(" &G+%s", strings.ToUpper(name))
	} else {
		ch.Sendf(" &g-%s", name)
	}
}

// DoChannels — per-player channel toggle. Alias `channels`. C: `do_channels`
// (`act_info.c:5269`). Empty arg prints a grouped status listing. A leading
// `+`/`-` toggles a single channel or the "all" public set. See
// plan-do-channels.md.
func DoChannels(ch *types.CharData, argument string) {
	if ch.IsNPC() {
		return
	}

	arg, _ := util.OneArgument(argument)

	if arg == "" {
		if ch.Act.IsSet(types.PLR_SILENCE) {
			ch.Send("You are silenced.\n\r")
			return
		}

		// Public channels header. C: act_info.c:5287-5313.
		ch.Send("\n\r &gPublic channels  (severe penalties for abuse)&G:\n\r  ")
		channelEntryLine(ch, "racetalk", types.CHANNEL_RACETALK)
		channelEntryLine(ch, "chat", types.CHANNEL_CHAT)
		if ch.GetTrust() > 4 {
			channelEntryLine(ch, "auction", types.CHANNEL_AUCTION)
		}
		channelEntryLine(ch, "traffic", types.CHANNEL_TRAFFIC)
		channelEntryLine(ch, "quest", types.CHANNEL_QUEST)
		channelEntryLine(ch, "wartalk", types.CHANNEL_WARTALK)
		if ch.Level >= types.LEVEL_HERO {
			channelEntryLine(ch, "avatar", types.CHANNEL_AVTALK)
		}
		channelEntryLine(ch, "music", types.CHANNEL_MUSIC)
		channelEntryLine(ch, "ask", types.CHANNEL_ASK)
		channelEntryLine(ch, "yell", types.CHANNEL_YELL)

		// Private channels. C: act_info.c:5316-5350.
		ch.Send("\n\r &gPrivate channels (severe penalties for abuse)&G:\n\r ")
		channelEntryLine(ch, "tells", types.CHANNEL_TELLS)
		channelEntryLine(ch, "whisper", types.CHANNEL_WHISPER)
		if ch.PCData != nil && ch.PCData.Clan != nil {
			switch ch.PCData.Clan.ClanType {
			case types.CLAN_ORDER:
				channelEntryLine(ch, "order", types.CHANNEL_ORDER)
			case types.CLAN_GUILD:
				channelEntryLine(ch, "guild", types.CHANNEL_GUILD)
			default:
				channelEntryLine(ch, "clan", types.CHANNEL_CLAN)
			}
		}
		if ch.IsImmortal() {
			channelEntryLine(ch, "newbie", types.CHANNEL_NEWBIE)
		}
		if ch.PCData != nil && ch.PCData.Council != nil {
			channelEntryLine(ch, "council", types.CHANNEL_COUNCIL)
		}
		if ch.PCData != nil && uint32(ch.PCData.Flags)&types.PCFLAG_RETIRED != 0 {
			channelEntryLine(ch, "retired", types.CHANNEL_RETIRED)
		}

		// Immortal channels. C: act_info.c:5353-5392. Per-entry trust gates
		// (sysdata.muse_level / log_level / think_level) are simplified to a
		// blanket IsImmortal() here — refinement tracked in
		// plan-do-channels.md.
		if ch.IsImmortal() {
			ch.Send("\n\r &gImmortal Channels&G:\n\r  ")
			channelEntryLine(ch, "immtalk", types.CHANNEL_IMMTALK)
			channelEntryLine(ch, "muse", types.CHANNEL_HIGHGOD)
			channelEntryLine(ch, "monitor", types.CHANNEL_MONITOR)
			channelEntryLine(ch, "death", types.CHANNEL_DEATH)
			channelEntryLine(ch, "auth", types.CHANNEL_AUTH)
			channelEntryLine(ch, "retired", types.CHANNEL_RETIRED)
			channelEntryLine(ch, "log", types.CHANNEL_LOG)
			channelEntryLine(ch, "build", types.CHANNEL_BUILD)
			channelEntryLine(ch, "comm", types.CHANNEL_COMM)
			channelEntryLine(ch, "warn", types.CHANNEL_WARN)
			channelEntryLine(ch, "high", types.CHANNEL_HIGH)
			channelEntryLine(ch, "bug", types.CHANNEL_BUG)
		}

		ch.Send("\n\r")
		return
	}

	// Non-empty arg: sign + keyword.
	sign := arg[0]
	if sign != '+' && sign != '-' {
		ch.Send("Channels -channel or +channel?\n\r")
		return
	}
	name := arg[1:]
	fClear := sign == '+' // clear deaf bit => channel enabled

	// "all" toggles the public set (C: act_info.c:5482-5542).
	if strings.EqualFold(name, "all") {
		publicAll := []int{
			types.CHANNEL_RACETALK,
			types.CHANNEL_AUCTION,
			types.CHANNEL_CHAT,
			types.CHANNEL_QUEST,
			types.CHANNEL_WARTALK,
			types.CHANNEL_PRAY,
			types.CHANNEL_TRAFFIC,
			types.CHANNEL_MUSIC,
			types.CHANNEL_ASK,
			types.CHANNEL_YELL,
		}
		for _, b := range publicAll {
			if fClear {
				ch.Deaf.Remove(b)
			} else {
				ch.Deaf.Set(b)
			}
		}
		if ch.Level >= types.LEVEL_IMMORTAL {
			if fClear {
				ch.Deaf.Remove(types.CHANNEL_AVTALK)
			} else {
				ch.Deaf.Set(types.CHANNEL_AVTALK)
			}
		}
		ch.Send("Ok.\n\r")
		return
	}

	// Exact match (case-insensitive) against the toggle table.
	for _, entry := range channelToggleTable {
		if strings.EqualFold(name, entry.name) {
			if fClear {
				ch.Deaf.Remove(entry.bit)
			} else {
				ch.Deaf.Set(entry.bit)
			}
			ch.Send("Ok.\n\r")
			return
		}
	}

	ch.Send("Set or clear which channel?\n\r")
}
