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
//
// publicAll marks channels belonging to the "public" set toggled by
// `channels +all` / `channels -all` (C act_info.c:5482-5542). The
// `+all`/`-all` handler iterates this table filtered by publicAll rather
// than maintaining a separate list, so adding a new public channel to
// the table automatically enrolls it in the `+all` set.
//
// AVTALK is NOT included in the public set here because C gates it on
// `ch->level >= LEVEL_IMMORTAL` (not LEVEL_HERO); the `+all` handler
// applies that trust gate separately.
var channelToggleTable = []struct {
	name      string
	bit       int
	publicAll bool
}{
	{"auction", types.CHANNEL_AUCTION, true},
	{"traffic", types.CHANNEL_TRAFFIC, true},
	{"chat", types.CHANNEL_CHAT, true},
	{"clan", types.CHANNEL_CLAN, false},
	{"council", types.CHANNEL_COUNCIL, false},
	{"guild", types.CHANNEL_GUILD, false},
	{"quest", types.CHANNEL_QUEST, true},
	{"tells", types.CHANNEL_TELLS, false},
	{"immtalk", types.CHANNEL_IMMTALK, false},
	{"log", types.CHANNEL_LOG, false},
	{"build", types.CHANNEL_BUILD, false},
	{"high", types.CHANNEL_HIGH, false},
	{"pray", types.CHANNEL_PRAY, true},
	{"avatar", types.CHANNEL_AVTALK, false}, // immortal-gated; see loop below
	{"monitor", types.CHANNEL_MONITOR, false},
	{"death", types.CHANNEL_DEATH, false},
	{"auth", types.CHANNEL_AUTH, false},
	{"newbie", types.CHANNEL_NEWBIE, false},
	{"music", types.CHANNEL_MUSIC, true},
	{"muse", types.CHANNEL_HIGHGOD, false},
	{"ask", types.CHANNEL_ASK, true},
	{"yell", types.CHANNEL_YELL, true},
	{"comm", types.CHANNEL_COMM, false},
	{"warn", types.CHANNEL_WARN, false},
	{"bug", types.CHANNEL_BUG, false},
	{"order", types.CHANNEL_ORDER, false},
	{"wartalk", types.CHANNEL_WARTALK, true},
	{"whisper", types.CHANNEL_WHISPER, false},
	{"racetalk", types.CHANNEL_RACETALK, true},
	{"retired", types.CHANNEL_RETIRED, false},
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

// talkChannel is the shared skeleton for public/clan-family channels — the
// six Phase-6 "extra channels" (music, racetalk, wartalk, counciltalk,
// guildtalk, newbiechat). Callers pass the channel bit, a canonical verb
// (e.g. "music", "racetalk"), and a recipient-filter predicate that decides
// whether each receiver's vch receives the broadcast. The helper owns:
// empty-arg rejection, PLR_SILENCE sender gate, deaf-sender block (with
// wartalk/yell exception), ROOM_SILENCE, self-echo, and the broadcast walk.
// Per-wrapper pre-gates (is-in-council, is-in-guild-clan, etc.) run in the
// wrapper before this helper is called.
//
// The deaf-auto-clear at C act_comm.c:514 is unreachable after the :511
// return and is NOT ported here (matching C semantics and DoImmtalk).
//
// AT_ color fidelity is deferred — uses inline &Y/&G/&D per Tier 9
// convention.
//
// Wait-state, PLR_WIZINVIS preamble, ROOM_LOGSPEECH append, AFLAG_SILENCE,
// and is_ignoring filter are NOT applied here — they are not honored by any
// current Go channel command; see TODO.md for the uniform follow-up.
//
// C reference: act_comm.c:392-858 (`talk_channel`).
func talkChannel(
	ch *types.CharData,
	argument string,
	channel int,
	verb string,
	recipientFilter func(vch *types.CharData) bool,
) {
	if argument == "" {
		// capitalize first rune — guard against empty verb.
		var cap string
		if verb == "" {
			cap = ""
		} else {
			cap = strings.ToUpper(verb[:1]) + verb[1:]
		}
		ch.Sendf("%s what?\n\r", cap)
		return
	}
	// PLR_SILENCE sender gate (C act_comm.c:500-504).
	if !ch.IsNPC() && ch.Act.IsSet(types.PLR_SILENCE) {
		ch.Sendf("You can't %s.\n\r", verb)
		return
	}
	// Deaf-sender block (C act_comm.c:505-513). The C source uses `||` which
	// always evaluates true and is a typo for `&&` — we port the intent form:
	// wartalk and yell senders are never blocked by their own deaf bit on
	// this channel. The deaf bit is NOT auto-cleared — the C xREMOVE_BIT at
	// :514 is unreachable when the :511 return fires.
	if ch.Deaf.IsSet(channel) &&
		channel != types.CHANNEL_WARTALK && channel != types.CHANNEL_YELL {
		ch.Sendf("You don't have the %s channel turned on. To turn it on, use the Channels command.\n\r", verb)
		return
	}
	// ROOM_SILENCE on sender's room (C act_comm.c:471-476).
	if ch.InRoom != nil && ch.InRoom.RoomFlags.IsSet(types.ROOM_SILENCE) {
		ch.Send("You can't do that here.\n\r")
		return
	}

	// Self-echo.
	ch.Sendf("&YYou %s '&G%s&Y'&D\n\r", verb, argument)

	// Broadcast walk (C act_comm.c:671-848).
	for _, d := range WorldRef.Descriptors {
		if d == nil || d.Connected != types.CON_PLAYING || d.Character == nil {
			continue
		}
		vch := d.Character
		if vch == ch {
			continue
		}
		if vch.Deaf.IsSet(channel) {
			continue
		}
		if vch.InRoom != nil && vch.InRoom.RoomFlags.IsSet(types.ROOM_SILENCE) {
			continue
		}
		if recipientFilter != nil && !recipientFilter(vch) {
			continue
		}
		heard := translateFor(ch, vch, argument)
		vch.Sendf("&Y%s %ss '&G%s&Y'&D\n\r", ch.Name, verb, heard)
	}
}

// DoMusic — public leisure channel. C: `do_music` (act_comm.c:1012-1021).
// NOT_AUTHED branch from C is dropped (Go has no auth split).
func DoMusic(ch *types.CharData, argument string) {
	if ch.IsNPC() {
		ch.Send("Huh?\n\r")
		return
	}
	talkChannel(ch, argument, types.CHANNEL_MUSIC, "music", nil)
}

// DoRacetalk — players of the same race. C: `do_racetalk`
// (act_comm.c:4624-4632).
func DoRacetalk(ch *types.CharData, argument string) {
	if ch.IsNPC() {
		ch.Send("Huh?\n\r")
		return
	}
	filter := func(vch *types.CharData) bool {
		return vch.Race == ch.Race
	}
	talkChannel(ch, argument, types.CHANNEL_RACETALK, "racetalk", filter)
}

// DoWartalk — pkill-only strategy channel. C: `do_wartalk`
// (act_comm.c:4612-4621). Verb is "war" matching C; produces "You war '...'"
// and "X wars '...'" — that's how C reads too.
func DoWartalk(ch *types.CharData, argument string) {
	if ch.IsNPC() {
		ch.Send("Huh?\n\r")
		return
	}
	// Pkill-only sender gate (C act_comm.c:432-436).
	if ch.PCData == nil || uint32(ch.PCData.Flags)&types.PCFLAG_DEADLY == 0 {
		ch.Send("Peacefuls have no need to use wartalk.\n\r")
		return
	}
	filter := func(vch *types.CharData) bool {
		return vch.PCData != nil && uint32(vch.PCData.Flags)&types.PCFLAG_DEADLY != 0
	}
	talkChannel(ch, argument, types.CHANNEL_WARTALK, "war", filter)
}

// DoCouncilTalk — shared-council chat. C: `do_counciltalk`
// (act_comm.c:975-990). Pointer-identity on PCData.Council.
func DoCouncilTalk(ch *types.CharData, argument string) {
	if ch.IsNPC() || ch.PCData == nil || ch.PCData.Council == nil {
		ch.Send("Huh?\n\r")
		return
	}
	council := ch.PCData.Council
	filter := func(vch *types.CharData) bool {
		return vch.PCData != nil && vch.PCData.Council == council
	}
	talkChannel(ch, argument, types.CHANNEL_COUNCIL, "counciltalk", filter)
}

// DoGuildTalk — shared-guild-clan chat. C: `do_guildtalk`
// (act_comm.c:993-1009). ClanType must be CLAN_GUILD (not CLAN_ORDER).
func DoGuildTalk(ch *types.CharData, argument string) {
	if ch.IsNPC() || ch.PCData == nil ||
		ch.PCData.Clan == nil ||
		ch.PCData.Clan.ClanType != types.CLAN_GUILD {
		ch.Send("Huh?\n\r")
		return
	}
	clan := ch.PCData.Clan
	filter := func(vch *types.CharData) bool {
		return vch.PCData != nil && vch.PCData.Clan == clan
	}
	talkChannel(ch, argument, types.CHANNEL_GUILD, "guildtalk", filter)
}

// isNewbieChannelMember matches C's newbiechat sender/receiver predicate
// (act_comm.c:937-944 sender; :717-721 receiver): immortal OR member of a
// council literally named "Newbie Council" (case-insensitive match).
// C's NOT_AUTHED branch collapses into the immortal disjunct since Go has
// no unauthed state.
//
// Note: stock SMAUG data ships an empty `db/councils/council.lst`, so the
// mortal branch only triggers on servers that explicitly create a "Newbie
// Council". Matches C semantics exactly for the same data configuration.
func isNewbieChannelMember(ch *types.CharData) bool {
	if ch.IsImmortal() {
		return true
	}
	if ch.PCData != nil && ch.PCData.Council != nil &&
		strings.EqualFold(ch.PCData.Council.Name, "Newbie Council") {
		return true
	}
	return false
}

// DoNewbieChat — chat for newbies plus immortals. C: `do_newbiechat`
// (act_comm.c:935-947). Receiver filter: immortal or Newbie-Council member.
func DoNewbieChat(ch *types.CharData, argument string) {
	if ch.IsNPC() {
		ch.Send("Huh?\n\r")
		return
	}
	if !isNewbieChannelMember(ch) {
		ch.Send("Huh?\n\r")
		return
	}
	filter := func(vch *types.CharData) bool {
		return isNewbieChannelMember(vch)
	}
	talkChannel(ch, argument, types.CHANNEL_NEWBIE, "newbiechat", filter)
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

	// "all" toggles the public set (C: act_info.c:5482-5542). The set is
	// derived from channelToggleTable.publicAll so adding a new public
	// channel to the table automatically enrolls it — no parallel list
	// to keep in sync.
	if strings.EqualFold(name, "all") {
		for _, entry := range channelToggleTable {
			if !entry.publicAll {
				continue
			}
			if fClear {
				ch.Deaf.Remove(entry.bit)
			} else {
				ch.Deaf.Set(entry.bit)
			}
		}
		// AVTALK is public-set gated on immortal-level-or-higher (C
		// act_info.c:5497 `ch->level >= LEVEL_IMMORTAL`). Kept separate
		// from publicAll because the gate is a player property, not a
		// channel property.
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
