// Package act clan_officer commands: induct / outcast / bestow. Ports
// C do_induct (src/clans.c:928-1097), do_outcast (src/clans.c:1168-1340),
// do_bestow (src/act_wiz.c:7079-7135). Officer-authority gating is inlined
// per-command via isClanOfficer (5-way: bestowed-keyword, deity, leader,
// number1, number2) — NOT a trust-level gate. A low-level PC who is
// leader of their clan can induct; a high-level PC who is not an officer
// cannot. Bestow is separately gated at LEVEL_IMMORTAL in boot.
package act

import (
	"fmt"
	"strings"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// ClanDir is the absolute path to the clans directory. Written once at
// boot.Boot so SaveClanFile knows where to persist. Parallel to
// PlanesFilePath (plan-phase6-planes.md §D6). Set-at-boot, read-from-
// game-loop; no synchronization needed.
var ClanDir string

// isClanOfficer replicates the 5-way caller gate shared by do_induct
// (src/clans.c:942-953) and do_outcast (src/clans.c:1184-1195):
//
//	bestowments contains the command keyword, OR
//	ch.Name matches clan.Deity, Leader, Number1, or Number2
//
// All string compares are case-insensitive (C str_cmp is
// case-insensitive; is_name splits on whitespace and compares with
// str_cmp). Returns false when ch or clan is nil, or when ch is an NPC.
func isClanOfficer(ch *types.CharData, clan *types.ClanData, command string) bool {
	if ch == nil || clan == nil || ch.IsNPC() || ch.PCData == nil {
		return false
	}
	if ch.PCData.Bestowments != "" && util.IsName(command, ch.PCData.Bestowments) {
		return true
	}
	if clan.Deity != "" && strings.EqualFold(ch.Name, clan.Deity) {
		return true
	}
	if clan.Leader != "" && strings.EqualFold(ch.Name, clan.Leader) {
		return true
	}
	if clan.Number1 != "" && strings.EqualFold(ch.Name, clan.Number1) {
		return true
	}
	if clan.Number2 != "" && strings.EqualFold(ch.Name, clan.Number2) {
		return true
	}
	return false
}

// isPkill mirrors C IS_PKILL macro: IS_SET(ch->pcdata->flags, PCFLAG_DEADLY).
// Package-private because six Go files already inline this expression;
// exporting would need a policy decision on where the method lives.
func isPkill(ch *types.CharData) bool {
	if ch == nil || ch.IsNPC() || ch.PCData == nil {
		return false
	}
	return uint32(ch.PCData.Flags)&types.PCFLAG_DEADLY != 0
}

// DoInduct ports C do_induct (src/clans.c:928-1097). Clan officer inducts
// a same-room PC into the officer's clan. Authority check is the 5-way
// isClanOfficer gate; level/class/type gates mirror C verbatim.
func DoInduct(ch *types.CharData, argument string) {
	if ch == nil || ch.IsNPC() || ch.PCData == nil || ch.PCData.Clan == nil {
		ch.Send("Huh?\n")
		return
	}
	clan := ch.PCData.Clan
	if !isClanOfficer(ch, clan, "induct") {
		ch.Send("Huh?\n")
		return
	}

	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Induct whom?\n\r")
		return
	}

	victim := handler.GetCharRoom(ch, arg)
	if victim == nil {
		ch.Send("That player is not here.\n\r")
		return
	}
	if victim.IsNPC() {
		ch.Send("Not on NPC's.\n\r")
		return
	}
	if victim.IsImmortal() {
		ch.Send("You can't induct such a godly presence.\n\r")
		return
	}

	// Peaceful-target gate: pkill-type clans require a deadly inductee.
	// Guild/order/nokill accept peaceful PCs.
	if !isPkill(victim) &&
		clan.ClanType != types.CLAN_GUILD &&
		clan.ClanType != types.CLAN_ORDER &&
		clan.ClanType != types.CLAN_NOKILL {
		ch.Send("You cannot induct a peaceful character.\n\r")
		return
	}

	// Guild class-match: victim.Class must match clan.Class.
	if clan.ClanType == types.CLAN_GUILD {
		if victim.Class != clan.Class {
			ch.Send("This player's will is not in accordance with your guild.\n\r")
			return
		}
	} else {
		// Non-guild: level-floor (10) and inductor-outranks gates.
		if victim.Level < 10 {
			ch.Send("This player is not worthy of joining yet.\n\r")
			return
		}
		if victim.Level > ch.Level {
			ch.Send("This player is too powerful for you to induct.\n\r")
			return
		}
	}

	// Already-in-clan gate — branches on victim's EXISTING clan type.
	if victim.PCData != nil && victim.PCData.Clan != nil {
		vclan := victim.PCData.Clan
		same := vclan == clan
		switch vclan.ClanType {
		case types.CLAN_ORDER:
			if same {
				ch.Send("This player already belongs to your order!\n\r")
			} else {
				ch.Send("This player already belongs to an order!\n\r")
			}
		case types.CLAN_GUILD:
			if same {
				ch.Send("This player already belongs to your guild!\n\r")
			} else {
				ch.Send("This player already belongs to an guild!\n\r")
			}
		default:
			if same {
				ch.Send("This player already belongs to your clan!\n\r")
			} else {
				ch.Send("This player already belongs to a clan!\n\r")
			}
		}
		return
	}

	// Member-limit gate.
	if clan.MemLimit > 0 && clan.Members >= clan.MemLimit {
		ch.Send("Your clan is too big to induct anymore players.\n\r")
		return
	}

	// Commit.
	clan.Members++

	// Non-order-non-guild → grant LANG_CLAN.
	if clan.ClanType != types.CLAN_ORDER && clan.ClanType != types.CLAN_GUILD {
		victim.Speaks |= int(types.LANG_CLAN)
	}

	// Pkill-type only (not nokill/order/guild) → clear PLR_NICE and set PCFLAG_DEADLY.
	if clan.ClanType != types.CLAN_NOKILL &&
		clan.ClanType != types.CLAN_ORDER &&
		clan.ClanType != types.CLAN_GUILD {
		victim.Act.Remove(types.PLR_NICE)
		victim.PCData.Flags |= int(types.PCFLAG_DEADLY)
	}

	// Pkill-type only → award all clan.Class-guild skills at per-class adept.
	if clan.ClanType != types.CLAN_GUILD &&
		clan.ClanType != types.CLAN_ORDER &&
		clan.ClanType != types.CLAN_NOKILL {
		if WorldRef != nil {
			for sn, sk := range WorldRef.Skills {
				if sk == nil || sk.Name == "" {
					continue
				}
				if sk.Guild != clan.Class {
					continue
				}
				if sn < 0 || sn >= types.MAX_SKILL {
					continue
				}
				if victim.Class >= 0 && victim.Class < types.MAX_CLASS {
					victim.PCData.Learned[sn] = sk.SkillAdept[victim.Class]
				}
				victim.Sendf("%s instructs you in the ways of %s.\n\r", ch.Name, sk.Name)
			}
		}
	}

	// Link-in.
	victim.PCData.Clan = clan
	victim.PCData.ClanName = clan.Name

	// Broadcast: C act() with $t = clan.Name.
	util.Act(types.AT_MAGIC, "You induct $N into $t", ch, victim, clan.Name, nil, types.TO_CHAR)
	util.Act(types.AT_MAGIC, "$n inducts $N into $t", ch, victim, clan.Name, nil, types.TO_NOTVICT)
	util.Act(types.AT_MAGIC, "$n inducts you into $t", ch, victim, clan.Name, nil, types.TO_VICT)

	// Persist victim pfile and clan state. SaveFunc is the act-layer
	// SavePlayer seam wired at boot.go:92 (set to GameLoop.SavePlayer).
	if SaveFunc != nil {
		SaveFunc(victim)
	}
	if ClanDir != "" {
		if err := persist.SaveClanFile(ClanDir, clan); err != nil {
			util.Bug("DoInduct: SaveClanFile(%s): %v", clan.Filename, err)
		}
	}
}

// officerRank returns the C do_outcast x/y value for a given name against
// the clan: 3=leader, 2=number1, 1=number2, 0=any other (including deity
// and bestowed-keyword officers). Matches src/clans.c:1218-1229 verbatim.
func officerRank(name string, clan *types.ClanData) int {
	if strings.EqualFold(name, clan.Leader) {
		return 3
	}
	if strings.EqualFold(name, clan.Number1) {
		return 2
	}
	if strings.EqualFold(name, clan.Number2) {
		return 1
	}
	return 0
}

// echoToPKers mirrors C echo_to_all(AT_MAGIC, msg, ECHOTAR_PK) for the
// outcast broadcast at src/clans.c:1328. Iterates WorldRef.Descriptors and
// sends msg to every PC with PCFLAG_DEADLY set. Local helper — deliberately
// not generalized into util/ per plan §D3 (narrow scope here).
func echoToPKers(msg string) {
	if WorldRef == nil {
		return
	}
	for _, d := range WorldRef.Descriptors {
		if d == nil || d.Character == nil {
			continue
		}
		vch := d.Character
		if vch.IsNPC() || vch.PCData == nil {
			continue
		}
		if uint32(vch.PCData.Flags)&types.PCFLAG_DEADLY == 0 {
			continue
		}
		vch.Sendf("%s\n\r", msg)
	}
}

// DoOutcast ports C do_outcast (src/clans.c:1168-1340). Clan officer removes
// a same-room PC from the clan. Officer-rank arithmetic gates the call; C
// preserves a deliberately non-strict "x <= y && trust <= trust" block so
// equal-rank peers cannot outcast each other.
func DoOutcast(ch *types.CharData, argument string) {
	if ch == nil || ch.IsNPC() || ch.PCData == nil || ch.PCData.Clan == nil {
		ch.Send("Huh?\n")
		return
	}
	clan := ch.PCData.Clan
	if !isClanOfficer(ch, clan, "outcast") {
		ch.Send("Huh?\n")
		return
	}

	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Outcast whom?\n\r")
		return
	}

	victim := handler.GetCharRoom(ch, arg)
	if victim == nil {
		ch.Send("That player is not here.\n\r")
		return
	}
	if victim.IsNPC() {
		ch.Send("Not on NPC's.\n\r")
		return
	}

	x := officerRank(ch.Name, clan)
	y := officerRank(victim.Name, clan)

	// Power gate: C src/clans.c:1231-1236. Deliberately non-strict — an
	// officer of equal rank with equal trust cannot outcast a peer.
	if x <= y && ch.GetTrust() <= victim.GetTrust() {
		ch.Send("You are not powerful enough to outcast this character.\n\r")
		return
	}

	// Self-outcast → type-specific message.
	if victim == ch {
		switch clan.ClanType {
		case types.CLAN_ORDER:
			ch.Send("Kick yourself out of your own order?\n\r")
		case types.CLAN_GUILD:
			ch.Send("Kick yourself out of your own guild?\n\r")
		default:
			ch.Send("Kick yourself out of your own clan?\n\r")
		}
		return
	}

	if victim.Level > ch.Level {
		ch.Send("This player is too powerful for you to outcast.\n\r")
		return
	}

	// Same-clan gate.
	if victim.PCData == nil || victim.PCData.Clan != clan {
		switch clan.ClanType {
		case types.CLAN_ORDER:
			ch.Send("This player does not belong to your order!\n\r")
		case types.CLAN_GUILD:
			ch.Send("This player does not belong to your guild!\n\r")
		default:
			ch.Send("This player does not belong to your clan!\n\r")
		}
		return
	}

	// Skill-forget loop: non-guild/order/nokill only. Zero victim.Learned[sn]
	// for every skill whose Guild == victim.PCData.Clan.Class.
	if clan.ClanType != types.CLAN_GUILD &&
		clan.ClanType != types.CLAN_ORDER &&
		clan.ClanType != types.CLAN_NOKILL {
		if WorldRef != nil {
			for sn, sk := range WorldRef.Skills {
				if sk == nil || sk.Name == "" {
					continue
				}
				if sk.Guild != victim.PCData.Clan.Class {
					continue
				}
				if sn < 0 || sn >= types.MAX_SKILL {
					continue
				}
				victim.PCData.Learned[sn] = 0
				victim.Sendf("You forget the ways of %s.\n\r", sk.Name)
			}
		}
	}

	// Language cleanup.
	if uint32(victim.Speaking)&types.LANG_CLAN != 0 {
		victim.Speaking = int(types.LANG_COMMON)
	}
	victim.Speaks &^= int(types.LANG_CLAN)

	clan.Members--

	// Blank the rank slot victim occupied (if any).
	if strings.EqualFold(victim.Name, clan.Number1) {
		clan.Number1 = ""
	}
	if strings.EqualFold(victim.Name, clan.Number2) {
		clan.Number2 = ""
	}

	// Clear victim's clan link.
	victim.PCData.Clan = nil
	victim.PCData.ClanName = ""

	// Broadcast. TO_VICT only when victim has a live descriptor; C checks
	// desc && desc->host (linkdead detection). Go has no loginmsg system
	// so the linkdead branch logs via util.Bug instead.
	util.Act(types.AT_MAGIC, "You outcast $N from $t", ch, victim, clan.Name, nil, types.TO_CHAR)
	util.Act(types.AT_MAGIC, "$n outcasts $N from $t", ch, victim, clan.Name, nil, types.TO_ROOM)
	if victim.Desc != nil {
		util.Act(types.AT_MAGIC, "$n outcasts you from $t", ch, victim, clan.Name, nil, types.TO_VICT)
	} else {
		util.Bug("DoOutcast: linkdead victim %s — loginmsg deferred", victim.Name)
	}

	// PKers echo: non-guild/non-order broadcast to all PKers.
	if clan.ClanType != types.CLAN_GUILD && clan.ClanType != types.CLAN_ORDER {
		echoToPKers(fmt.Sprintf("%s has been outcast from %s!", victim.Name, clan.Name))
	}

	// Persist.
	if SaveFunc != nil {
		SaveFunc(victim)
	}
	if ClanDir != "" {
		if err := persist.SaveClanFile(ClanDir, clan); err != nil {
			util.Bug("DoOutcast: SaveClanFile(%s): %v", clan.Filename, err)
		}
	}
}

// DoBestow ports C do_bestow (src/act_wiz.c:7079-7135). Immortal-only
// command registered at LEVEL_IMMORTAL; mutates victim.PCData.Bestowments
// (a space-separated command whitelist). C's leading-space quirk is
// preserved verbatim — a fresh bestow produces " induct" with a leading
// blank. util.IsName is whitespace-tolerant so isClanOfficer sees the
// keyword correctly.
//
// Anomaly (preserved): C do_bestow never calls save_char_obj(victim);
// bestowments persist only on the victim's next normal pfile save. The
// Go port matches — no SaveFunc(victim) call here.
func DoBestow(ch *types.CharData, argument string) {
	if ch == nil {
		return
	}

	arg, argAfter := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Bestow whom with what?\n\r")
		return
	}
	victim := handler.GetCharWorld(WorldRef, ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n")
		return
	}
	if victim.IsNPC() {
		ch.Send("You can't give special abilities to a mob!\n")
		return
	}
	if victim.GetTrust() > ch.GetTrust() {
		ch.Send("You aren't powerful enough...\n")
		return
	}
	if victim.PCData == nil {
		ch.Send("They aren't here.\n")
		return
	}

	// Empty second-arg or "list" → list current bestowments.
	if argAfter == "" || strings.EqualFold(strings.TrimSpace(argAfter), "list") {
		ch.Sendf("Current bestowed commands on %s: %s.\n\r",
			victim.Name, victim.PCData.Bestowments)
		return
	}

	// "none" → clear bestowments, notify both.
	if strings.EqualFold(strings.TrimSpace(argAfter), "none") {
		victim.PCData.Bestowments = ""
		ch.Sendf("Bestowments removed from %s.\n\r", victim.Name)
		victim.Sendf("%s has removed your bestowed commands.\n\r", ch.Name)
		return
	}

	// Append. C sprintf("%s %s", old, new) introduces a leading space when
	// old is empty; port verbatim per plan §D4.
	victim.PCData.Bestowments = victim.PCData.Bestowments + " " + argAfter
	victim.Sendf("%s has bestowed on you the command(s): %s\n\r", ch.Name, argAfter)
	ch.Send("Done.\n")
}
