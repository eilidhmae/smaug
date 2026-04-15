package act

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// ShutdownFunc is set from main. When called with reboot=true the process
// should exit with a status that signals a restart supervisor to relaunch;
// otherwise a graceful shutdown. MVP: after saving all players the caller
// simply cancels the context / calls os.Exit(0).
var ShutdownFunc func(reboot bool)

// DisconnectFunc is set from main to force-close a descriptor (save the
// player first, then tear down the socket). The loop package owns
// descriptor lifecycle, so this is wired at boot to avoid an import cycle.
var DisconnectFunc func(d *types.DescriptorData)

// HellRoomVnum is the destination for the 'hell' command. C uses vnum 8
// (room 8 in hell.are). If the room isn't loaded the command is a no-op
// with an error message — test suites don't load hell.are so we guard.
const HellRoomVnum = 8

// -- Possession ---------------------------------------------------------

// DoSwitch implements the 'switch' command: an immortal inhabits an NPC
// body, redirecting their descriptor to the NPC. See act_wiz.c:3760.
func DoSwitch(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Switch into whom?\n\r")
		return
	}
	if ch.Desc == nil {
		return
	}
	if ch.Desc.Original != nil {
		ch.Send("You are already switched.\n\r")
		return
	}

	victim := handler.GetCharWorld(WorldRef, ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}
	if victim == ch {
		ch.Send("Ok.\n\r")
		return
	}
	if victim.Desc != nil {
		ch.Send("Character in use.\n\r")
		return
	}
	if !victim.IsNPC() && ch.Level < types.LEVEL_GREATER {
		ch.Send("You cannot switch into a player!\n\r")
		return
	}
	if victim.Switched != nil {
		ch.Send("You can't switch into a player that is switched!\n\r")
		return
	}
	if !victim.IsNPC() && victim.Act.IsSet(types.PLR_FREEZE) {
		ch.Send("You shouldn't switch into a player that is frozen!\n\r")
		return
	}
	// C act_wiz.c:3793-3799: stat-shielded NPCs block switch below LEVEL_GREATER.
	if victim.IsNPC() && victim.Act.IsSet(types.ACT_STATSHIELD) && ch.GetTrust() < types.LEVEL_GREATER {
		ch.Send("Their godly glow prevents you from getting close enough.\n\r")
		return
	}

	ch.Desc.Character = victim
	ch.Desc.Original = ch
	victim.Desc = ch.Desc
	ch.Desc = nil
	ch.Switched = victim
	victim.Send("Ok.\n\r")
}

// DoReturn implements the 'return' command: exit possession and return
// to the immortal's own body. See act_wiz.c:3830. ch here is the NPC
// puppet; ch.Desc points at the descriptor that currently drives it, and
// ch.Desc.Original is the real immortal CharData to restore.
func DoReturn(ch *types.CharData, argument string) {
	if !ch.IsNPC() && ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}
	if ch.Desc == nil {
		return
	}
	if ch.Desc.Original == nil {
		ch.Send("You aren't switched.\n\r")
		return
	}

	d := ch.Desc
	orig := d.Original
	// Redirect the descriptor back to the immortal.
	d.Character = orig
	d.Original = nil
	orig.Desc = d
	orig.Switched = nil
	ch.Desc = nil
	orig.Send("You return to your original body.\n\r")
}

// -- Wizlock / shutdown / reboot ---------------------------------------

// DoWizlock toggles SysData.Wizlock. See act_wiz.c:6075.
func DoWizlock(ch *types.CharData, argument string) {
	if WorldRef == nil {
		return
	}
	WorldRef.SysData.Wizlock = !WorldRef.SysData.Wizlock
	if WorldRef.SysData.Wizlock {
		ch.Send("Wizlock is now ON.\n\r")
		for _, d := range WorldRef.Descriptors {
			if d.Character != nil && d.Character != ch {
				d.Character.Sendf("%s has wizlocked the game.\n\r", ch.Name)
			}
		}
	} else {
		ch.Send("Wizlock is now OFF.\n\r")
		for _, d := range WorldRef.Descriptors {
			if d.Character != nil && d.Character != ch {
				d.Character.Sendf("%s has removed the wizlock.\n\r", ch.Name)
			}
		}
	}
}

// DoShutdown triggers graceful shutdown. See act_wiz.c:3606.
func DoShutdown(ch *types.CharData, argument string) {
	ch.Sendf("%s has shut down the game.\n\r", ch.Name)
	if WorldRef != nil {
		for _, d := range WorldRef.Descriptors {
			if d.Character != nil && d.Character != ch {
				d.Character.Sendf("%s has shut down the game.\n\r", ch.Name)
			}
		}
	}
	if ShutdownFunc != nil {
		ShutdownFunc(false)
	}
}

// DoReboot triggers graceful reboot. See act_wiz.c:3554. MVP: same as
// shutdown but passes reboot=true so a supervisor can distinguish.
func DoReboot(ch *types.CharData, argument string) {
	ch.Sendf("%s has rebooted the game.\n\r", ch.Name)
	if WorldRef != nil {
		for _, d := range WorldRef.Descriptors {
			if d.Character != nil && d.Character != ch {
				d.Character.Sendf("%s has rebooted the game.\n\r", ch.Name)
			}
		}
	}
	if ShutdownFunc != nil {
		ShutdownFunc(true)
	}
}

// -- Help listing ------------------------------------------------------

// DoWizhelp lists all commands an immortal can use, alphabetical.
// See act_wiz.c:610.
func DoWizhelp(ch *types.CharData, argument string) {
	if CmdRegistry == nil {
		ch.Send("No commands registered.\n\r")
		return
	}
	trust := ch.GetTrust()
	var names []string
	for _, cmd := range CmdRegistry.All() {
		if cmd.Level >= types.LEVEL_IMMORTAL && trust >= cmd.Level {
			names = append(names, cmd.Name)
		}
	}
	if len(names) == 0 {
		ch.Send("You have no immortal commands available.\n\r")
		return
	}
	sort.Strings(names)

	var sb strings.Builder
	sb.WriteString("&WImmortal commands available to you:&D\n\r")
	col := 0
	for _, n := range names {
		fmt.Fprintf(&sb, "%-14s", n)
		col++
		if col >= 5 {
			sb.WriteString("\n\r")
			col = 0
		}
	}
	if col != 0 {
		sb.WriteString("\n\r")
	}
	sendToPager(ch, sb.String())
}

// -- Area echo ---------------------------------------------------------

// DoAecho sends a message to every char in the same area as the imm.
// See act_wiz.c:1342.
func DoAecho(ch *types.CharData, argument string) {
	if argument == "" {
		ch.Send("Aecho what?\n\r")
		return
	}
	if ch.InRoom == nil || ch.InRoom.Area == nil {
		ch.Send("You are nowhere.\n\r")
		return
	}
	area := ch.InRoom.Area
	for _, other := range WorldRef.Characters {
		if other.InRoom != nil && other.InRoom.Area == area && other.Desc != nil {
			other.Sendf("%s\n\r", argument)
		}
	}
}

// -- Player discipline -------------------------------------------------

// DoHell sends a player to hell. See act_wiz.c:8498. MVP: sets Hell
// timestamp and teleports to HellRoomVnum if loaded.
func DoHell(ch *types.CharData, argument string) {
	arg, rest := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Usage: hell <player> <hours>\n\r")
		return
	}
	victim := handler.GetCharWorld(WorldRef, ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}
	if victim.IsNPC() {
		ch.Send("Not on NPCs.\n\r")
		return
	}
	// C act_wiz.c:8518: blocks helling another immortal regardless of trust.
	if victim.GetTrust() >= types.LEVEL_IMMORTAL {
		ch.Send("There is no point in helling an immortal.\n\r")
		return
	}
	if victim.PCData == nil {
		return
	}
	// C act_wiz.c:8526-8530: already-in-hell gets a status report, not overwrite.
	if victim.PCData.Hell != 0 {
		ch.Sendf("They are already in hell until %s, by %s.\n\r",
			time.Unix(victim.PCData.Hell, 0).Format("2006-01-02 15:04"),
			victim.PCData.HelledBy)
		return
	}
	hours := 1
	if rest != "" {
		h, err := parseIntSafe(strings.TrimSpace(rest))
		if err == nil && h > 0 {
			hours = h
		}
	}
	// Verify the hell room exists BEFORE writing hell state, or a failed
	// teleport would leave the victim permanently flagged with no room to
	// go to. Adversary-caught pre-land.
	hellRoom := WorldRef.GetRoom(HellRoomVnum)
	if hellRoom == nil {
		ch.Send("No hell room defined.\n\r")
		return
	}
	victim.PCData.Hell = time.Now().Add(time.Duration(hours) * time.Hour).Unix()
	victim.PCData.HelledBy = ch.Name
	if victim.InRoom != nil {
		handler.CharFromRoom(victim)
	}
	handler.CharToRoom(victim, hellRoom)
	victim.Sendf("%s has sent you to hell!\n\r", ch.Name)
	ch.Sendf("%s sent to hell for %d hour(s).\n\r", victim.Name, hours)
}

// DoLog toggles PLR_LOG on a player. See act_wiz.c:5677. The command
// interpreter already logs all player input; this flag is read by the
// interpreter log hook (if wired) to filter which players are logged.
func DoLog(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Log whom?\n\r")
		return
	}
	if strings.EqualFold(arg, "all") {
		ch.Send("Log all toggle not supported.\n\r")
		return
	}
	victim := handler.GetCharWorld(WorldRef, ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}
	if victim.IsNPC() {
		ch.Send("Not on NPCs.\n\r")
		return
	}
	if victim.Act.IsSet(types.PLR_LOG) {
		victim.Act.Toggle(types.PLR_LOG)
		ch.Sendf("%s is no longer being logged.\n\r", victim.Name)
	} else {
		victim.Act.Set(types.PLR_LOG)
		ch.Sendf("%s is now being logged.\n\r", victim.Name)
	}
}

// DoDeny sets PLR_DENY on a player. See act_wiz.c:1083.
func DoDeny(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Deny whom?\n\r")
		return
	}
	victim := handler.GetCharWorld(WorldRef, ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}
	if victim.IsNPC() {
		ch.Send("Not on NPCs.\n\r")
		return
	}
	if victim.GetTrust() >= ch.GetTrust() {
		ch.Send("You can't do that.\n\r")
		return
	}
	victim.Act.Set(types.PLR_DENY)
	victim.Send("You are denied access!\n\r")
	ch.Sendf("%s denied.\n\r", victim.Name)
	// C act_wiz.c:1110-1115: deny kicks the victim immediately. We reuse the
	// disconnect hook wired at boot to save-and-close; if it isn't wired the
	// player will remain in-game until they manually quit (MVP behaviour).
	if DisconnectFunc != nil && victim.Desc != nil {
		DisconnectFunc(victim.Desc)
	}
}

// DoPardon clears disciplinary flags on a player. See act_wiz.c:1241.
// Syntax: pardon <name> <killer|thief|deny>
func DoPardon(ch *types.CharData, argument string) {
	arg, rest := util.OneArgument(argument)
	flagArg, _ := util.OneArgument(rest)
	if arg == "" || flagArg == "" {
		ch.Send("Syntax: pardon <name> <killer|thief|attacker|deny>\n\r")
		return
	}
	victim := handler.GetCharWorld(WorldRef, ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}
	if victim.IsNPC() {
		ch.Send("Not on NPCs.\n\r")
		return
	}
	switch strings.ToLower(flagArg) {
	case "killer":
		if victim.Act.IsSet(types.PLR_KILLER) {
			victim.Act.Toggle(types.PLR_KILLER)
			ch.Send("Killer flag removed.\n\r")
			victim.Send("You are no longer a KILLER.\n\r")
		} else {
			ch.Send("They aren't a killer.\n\r")
		}
	case "thief":
		if victim.Act.IsSet(types.PLR_THIEF) {
			victim.Act.Toggle(types.PLR_THIEF)
			ch.Send("Thief flag removed.\n\r")
			victim.Send("You are no longer a THIEF.\n\r")
		} else {
			ch.Send("They aren't a thief.\n\r")
		}
	case "deny":
		if victim.Act.IsSet(types.PLR_DENY) {
			victim.Act.Toggle(types.PLR_DENY)
			ch.Send("Deny flag removed.\n\r")
		} else {
			ch.Send("They aren't denied.\n\r")
		}
	case "attacker":
		// C act_wiz.c:1269-1285
		if victim.Act.IsSet(types.PLR_ATTACKER) {
			victim.Act.Toggle(types.PLR_ATTACKER)
			ch.Send("Attacker flag removed.\n\r")
			victim.Send("You are no longer an ATTACKER.\n\r")
		} else {
			ch.Send("They aren't an attacker.\n\r")
		}
	default:
		ch.Send("Syntax: pardon <name> <killer|thief|attacker|deny>\n\r")
	}
}

// DoDisconnect force-closes a player's descriptor. See act_wiz.c:1123.
func DoDisconnect(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Disconnect whom?\n\r")
		return
	}
	victim := handler.GetCharWorld(WorldRef, ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}
	if victim == ch {
		ch.Send("Yourself?\n\r")
		return
	}
	if victim.Desc == nil {
		ch.Send("They have no descriptor.\n\r")
		return
	}
	if !victim.IsNPC() && victim.GetTrust() >= ch.GetTrust() {
		ch.Send("You can't do that.\n\r")
		return
	}
	d := victim.Desc
	if SaveFunc != nil {
		SaveFunc(victim)
	}
	if DisconnectFunc != nil {
		DisconnectFunc(d)
	} else if d.Conn != nil {
		_ = d.Conn.Close()
	}
	ch.Sendf("%s disconnected.\n\r", victim.Name)
}

// DoMortalize strips immortal status from a character. See act_wiz.c:6594.
// MVP: zero trust and clear holylight / wizinvis bits.
func DoMortalize(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Mortalize whom?\n\r")
		return
	}
	victim := handler.GetCharWorld(WorldRef, ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}
	if victim.IsNPC() {
		ch.Send("Not on NPCs.\n\r")
		return
	}
	if victim.GetTrust() >= ch.GetTrust() && victim != ch {
		ch.Send("You can't do that.\n\r")
		return
	}
	victim.Trust = 0
	if victim.Act.IsSet(types.PLR_HOLYLIGHT) {
		victim.Act.Toggle(types.PLR_HOLYLIGHT)
	}
	if victim.Act.IsSet(types.PLR_WIZINVIS) {
		victim.Act.Toggle(types.PLR_WIZINVIS)
	}
	if victim.PCData != nil {
		victim.PCData.WizInvis = 0
	}
	victim.Sendf("%s has made you mortal!\n\r", ch.Name)
	ch.Sendf("%s mortalized.\n\r", victim.Name)
}

// -- MudProg inspection -----------------------------------------------

// DoMpstat shows an NPC's mudprogs. See mud_comm.c:148.
func DoMpstat(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("MProg stat whom?\n\r")
		return
	}
	victim := handler.GetCharWorld(WorldRef, ch, arg)
	if victim == nil {
		ch.Send("They aren't here.\n\r")
		return
	}
	if !victim.IsNPC() {
		ch.Send("Only Mobiles can have MobPrograms!\n\r")
		return
	}
	if victim.IndexData == nil || len(victim.IndexData.MudProgs) == 0 {
		ch.Sendf("No programs on mobile: %s - #%d\n\r", victim.Name, mobIndexVnum(victim))
		return
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Name: %s.  Vnum: %d.\n\r", victim.Name, mobIndexVnum(victim))
	fmt.Fprintf(&sb, "Short: %s\n\r", victim.ShortDescr)
	for i, mprg := range victim.IndexData.MudProgs {
		fmt.Fprintf(&sb, "%d >%s %s\n\r%s\n\r",
			i+1, progTypeNames(mprg.Type), mprg.ArgList, firstLine(mprg.ComList))
	}
	sendToPager(ch, sb.String())
}

// DoOpstat shows an object's mudprogs. See mud_comm.c:222.
func DoOpstat(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("OProg stat what?\n\r")
		return
	}
	obj := handler.GetObjWorld(WorldRef, ch, arg)
	if obj == nil {
		ch.Send("You cannot find that.\n\r")
		return
	}
	if obj.IndexData == nil || len(obj.IndexData.MudProgs) == 0 {
		ch.Send("That object has no programs set.\n\r")
		return
	}
	vnum := 0
	if obj.IndexData != nil {
		vnum = obj.IndexData.Vnum
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Name: %s.  Vnum: %d.\n\r", obj.Name, vnum)
	fmt.Fprintf(&sb, "Short: %s\n\r", obj.ShortDescr)
	for i, mprg := range obj.IndexData.MudProgs {
		fmt.Fprintf(&sb, "%d >%s %s\n\r%s\n\r",
			i+1, progTypeNames(mprg.Type), mprg.ArgList, firstLine(mprg.ComList))
	}
	sendToPager(ch, sb.String())
}

// DoRpstat shows the current room's mudprogs. See mud_comm.c:265.
func DoRpstat(ch *types.CharData, argument string) {
	if ch.InRoom == nil {
		return
	}
	if len(ch.InRoom.MudProgs) == 0 {
		ch.Send("This room has no programs set.\n\r")
		return
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Name: %s.  Vnum: %d.\n\r", ch.InRoom.Name, ch.InRoom.Vnum)
	for i, mprg := range ch.InRoom.MudProgs {
		fmt.Fprintf(&sb, "%d >%s %s\n\r%s\n\r",
			i+1, progTypeNames(mprg.Type), mprg.ArgList, firstLine(mprg.ComList))
	}
	sendToPager(ch, sb.String())
}

// -- helpers -----------------------------------------------------------

func mobIndexVnum(ch *types.CharData) int {
	if ch != nil && ch.IndexData != nil {
		return ch.IndexData.Vnum
	}
	return 0
}

// firstLine returns the first line (up to 60 chars) of a prog comlist
// so mpstat/opstat/rpstat output stays compact. C prints the whole
// script; the Go port shortens it to match pager-friendly width.
func firstLine(s string) string {
	if s == "" {
		return ""
	}
	// Trim leading whitespace/newlines.
	s = strings.TrimLeft(s, " \t\r\n")
	if idx := strings.IndexAny(s, "\r\n"); idx >= 0 {
		s = s[:idx]
	}
	if len(s) > 60 {
		s = s[:60] + "..."
	}
	return s
}

// progTypeNames renders a bitmask of MPROG_* flags as a space-separated
// list of keyword names, e.g. "act speech". Mirrors C mprog_type_to_name
// which returns a single name per type; we emit all set bits.
func progTypeNames(mask int64) string {
	if mask == 0 {
		return "none"
	}
	type entry struct {
		bit  int64
		name string
	}
	table := []entry{
		{types.MPROG_ACT, "act"},
		{types.MPROG_SPEECH, "speech"},
		{types.MPROG_RAND, "rand"},
		{types.MPROG_FIGHT, "fight"},
		{types.MPROG_DEATH, "death"},
		{types.MPROG_HITPRCNT, "hitprcnt"},
		{types.MPROG_ENTRY, "entry"},
		{types.MPROG_GREET, "greet"},
		{types.MPROG_ALL_GREET, "all_greet"},
		{types.MPROG_GIVE, "give"},
		{types.MPROG_BRIBE, "bribe"},
		{types.MPROG_HOUR, "hour"},
		{types.MPROG_TIME, "time"},
		{types.MPROG_WEAR, "wear"},
		{types.MPROG_REMOVE, "remove"},
		{types.MPROG_SAC, "sac"},
		{types.MPROG_LOOK, "look"},
		{types.MPROG_EXA, "exa"},
		{types.MPROG_ZAP, "zap"},
		{types.MPROG_GET, "get"},
		{types.MPROG_DROP, "drop"},
		{types.MPROG_DAMAGE, "damage"},
		{types.MPROG_REPAIR, "repair"},
		{types.MPROG_RANDIW, "randiw"},
		{types.MPROG_SPEECHIW, "speechiw"},
		{types.MPROG_PULL, "pull"},
		{types.MPROG_PUSH, "push"},
		{types.MPROG_SLEEP, "sleep"},
		{types.MPROG_REST, "rest"},
		{types.MPROG_LEAVE, "leave"},
		{types.MPROG_SCRIPT, "script"},
		{types.MPROG_USE, "use"},
		{types.MPROG_LOGIN, "login"},
		{types.MPROG_VOID, "void"},
		{types.MPROG_TELL, "tell"},
		{types.MPROG_SELL, "sell"},
		{types.MPROG_IMMINFO, "imminfo"},
		{types.MPROG_CMD, "cmd"},
	}
	var parts []string
	for _, e := range table {
		if mask&e.bit != 0 {
			parts = append(parts, e.name)
		}
	}
	if len(parts) == 0 {
		return fmt.Sprintf("0x%x", mask)
	}
	return strings.Join(parts, " ")
}

// parseIntSafe is a small wrapper that tolerates leading/trailing spaces.
func parseIntSafe(s string) (int, error) {
	s = strings.TrimSpace(s)
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

// DoSlookup implements 'slookup' — immortal skill inspector. Shows the
// metadata for one skill (name, type, level/adept per class, slot, mana,
// beats). Port of src/skills.c:615 do_slookup.
func DoSlookup(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("Syntax: slookup <skill|spell>\n\r")
		return
	}
	sn := -1
	if arg == "all" {
		for i, sk := range WorldRef.Skills {
			if sk == nil {
				continue
			}
			ch.Sendf("%3d: %s\n\r", i, sk.Name)
		}
		return
	}
	arg = strings.ToLower(arg)
	for i, sk := range WorldRef.Skills {
		if sk != nil && strings.HasPrefix(strings.ToLower(sk.Name), arg) {
			sn = i
			break
		}
	}
	if sn < 0 {
		ch.Send("No such skill or spell.\n\r")
		return
	}
	sk := WorldRef.Skills[sn]
	ch.Sendf("Sn: %d  Name: %s\n\r", sn, sk.Name)
	ch.Sendf("Type: %d  Slot: %d  MinMana: %d  Beats: %d\n\r",
		sk.Type, sk.Slot, sk.MinMana, sk.Beats)
	ch.Sendf("Target: %d  MinPos: %d  Difficulty: %d\n\r",
		sk.Target, sk.MinimumPos, sk.Difficulty)
	ch.Sendf("SpellFun: %s  SkillFun: %s\n\r", sk.SpellFunName, sk.SkillFunName)
	if sk.DiceFormula != "" {
		ch.Sendf("Dice: %s\n\r", sk.DiceFormula)
	}
	if sk.NounDamage != "" {
		ch.Sendf("Noun: %s\n\r", sk.NounDamage)
	}
}

// DoSset implements 'sset' — immortal skill modifier. Syntax:
//
//	sset <skill> <field> <value>
//
// Supported fields (MVP): slot, minmana, beats, difficulty, target, minpos,
// name. Port of src/skills.c:862 do_sset.
func DoSset(ch *types.CharData, argument string) {
	arg1, rest := util.OneArgument(argument)
	arg2, rest := util.OneArgument(rest)
	arg3, _ := util.OneArgument(rest)
	if arg1 == "" || arg2 == "" || arg3 == "" {
		ch.Send("Syntax: sset <skill> <field> <value>\n\r")
		ch.Send("Fields: slot, minmana, beats, difficulty, target, minpos, name\n\r")
		return
	}
	arg1 = strings.ToLower(arg1)
	sn := -1
	for i, sk := range WorldRef.Skills {
		if sk != nil && strings.HasPrefix(strings.ToLower(sk.Name), arg1) {
			sn = i
			break
		}
	}
	if sn < 0 {
		ch.Send("No such skill or spell.\n\r")
		return
	}
	sk := WorldRef.Skills[sn]
	field := strings.ToLower(arg2)
	if field == "name" {
		sk.Name = arg3
		ch.Sendf("Set %s name = %s\n\r", sk.Name, arg3)
		return
	}
	val, err := parseIntSafe(arg3)
	if err != nil {
		ch.Send("Value must be an integer.\n\r")
		return
	}
	switch field {
	case "slot":
		sk.Slot = val
	case "minmana":
		sk.MinMana = val
	case "beats":
		sk.Beats = val
	case "difficulty":
		sk.Difficulty = val
	case "target":
		sk.Target = val
	case "minpos":
		sk.MinimumPos = val
	default:
		ch.Send("Unknown field.\n\r")
		return
	}
	ch.Sendf("Set %s %s = %d\n\r", sk.Name, field, val)
}
