package act

import (
	"fmt"
	"math/bits"
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// Phase 6 OLC mpedit/opedit/rpedit dispatchers.
//
// Ports C do_mpedit (src/build.c:9039-9376), do_opedit (:9379-9719) and
// do_rpedit (:9744-10058) — three near-identical builder commands that
// add/delete/insert/edit/list mud-programs on a mob/obj/room prototype.
//
// Architectural divergence from C: C re-enters do_mpedit on /s via the
// SUB_MPROG_EDIT substate machinery (interp.c's do_last_cmd path). The Go
// port replaces that re-entry path with the EditorSave-callback pattern
// (Phase 5 Tier 12 — internal/types/character.go:60). The closure captures
// the *MProgData pointer (and the *progEditorTarget for progtypes rebuild
// on edit) so no descriptor-level state is needed. ch.Substate is still
// set to SUB_MPROG_EDIT during edit because internal/game/editor.go:187
// extends maxBufLines for that substate — load-bearing despite the
// re-entry path being a no-op.

// parseProgTriggerName delegates to util.GetMpFlag (single source of truth).
func parseProgTriggerName(s string) (int64, bool) {
	return util.GetMpFlag(s)
}

// firstTriggerName returns a readable name for the first bit set in mask,
// or a hex fallback when no known bit is set.
func firstTriggerName(mask int64) string {
	if name := util.FirstMProgFlagName(mask); name != "" {
		return name
	}
	return fmt.Sprintf("0x%x", mask)
}

// progEditorTarget bundles the per-kind state a subcommand body needs:
// the prog list (pointer-to-slice so append/splice mutations are visible
// to the caller) and the progtypes BitVector pointer.
type progEditorTarget struct {
	kind       string // "mob" | "obj" | "room"
	label      string // "Mob" | "Object" | "Room" — for error/list messages
	progs      *[]*types.MProgData
	progTypes  *types.BitVector
	targetName string // mob/obj name OR room vnum-string, for display
	targetVnum int
}

// DoMpedit implements the 'mpedit' command. Syntax:
//
//	mpedit <victim> <command> [number] <program> <value>
//
// where <command> is one of list / add / insert / edit / delete.
func DoMpedit(ch *types.CharData, argument string) {
	progEditDispatch(ch, argument, "mob")
}

// DoOpedit implements 'opedit' — same shape with object victim.
func DoOpedit(ch *types.CharData, argument string) {
	progEditDispatch(ch, argument, "obj")
}

// DoRpedit implements 'rpedit' — operates on ch.InRoom (no victim arg).
func DoRpedit(ch *types.CharData, argument string) {
	progEditDispatch(ch, argument, "room")
}

func progEditDispatch(ch *types.CharData, argument, kind string) {
	// --- common gates: NPC reject + descriptor guard ---
	if ch.IsNPC() {
		ch.Sendf("Mob's can't %sedit\n\r", progEditCmdPrefix(kind))
		return
	}
	if ch.Desc == nil {
		ch.Send("You have no descriptor.\n\r")
		return
	}

	// Trust gate (registry already enforces LEVEL_IMMORTAL but defend in depth).
	if ch.GetTrust() < types.LEVEL_IMMORTAL {
		ch.Send("Huh?\n\r")
		return
	}

	argument = util.SmashTilde(argument)

	// --- per-kind argument parsing ---
	// mpedit/opedit: arg1=target, arg2=subcommand, arg3=number-or-program, arg4=program-or-arglist; rest = remaining arglist after arg3 (for add) or after arg4 (for insert/edit).
	// rpedit:        arg1=subcommand, arg2=number-or-program, arg3=program-or-arglist
	var arg1, arg2, arg3, arg4 string
	var afterArg2, afterArg3, afterArg4 string
	arg1, afterArg2 = util.OneArgument(argument)
	arg2, afterArg3 = util.OneArgument(afterArg2)
	arg3, afterArg4 = util.OneArgument(afterArg3)
	arg4, _ = util.OneArgument(afterArg4)

	// --- syntax help branch ---
	if kind == "room" {
		if arg1 == "" {
			progEditSyntaxHelp(ch, kind)
			return
		}
	} else {
		if arg1 == "" || arg2 == "" {
			progEditSyntaxHelp(ch, kind)
			return
		}
	}

	// --- per-kind target resolution ---
	target, ok := resolveProgEditTarget(ch, kind, arg1)
	if !ok {
		return
	}

	// --- subcommand + value parsing (per-kind argument index shift) ---
	var subcmd, valueStr, progStr, addArglist, insArglist string
	// Compute arglist tail for add (after type) and insert/edit (after type).
	// Mpedit/opedit consume one extra arg (target name) before subcommand,
	// shifting all slots by one relative to rpedit. After OneArgument has
	// peeled arg1..arg4 off the front, the remaining `tailAfterArgN` strings
	// hold the residual arglist for the corresponding subcommand shape.
	_, _ = arg4, afterArg4
	tailAfterArg3 := strings.TrimSpace(afterArg3)
	tailAfterArg4 := strings.TrimSpace(afterArg4)
	if kind == "room" {
		subcmd = strings.ToLower(arg1)
		valueStr = arg2
		// rpedit add:    arg1=add,    arg2=type,   arglist=tailAfterArg2 (i.e. after type=arg2 → starts at arg3-position)
		// rpedit insert/edit: arg1=insert/edit, arg2=number, arg3=type, arglist=tailAfterArg3
		progStr = arg3
		// "tailAfterArg2" = residual after arg2 parse = afterArg3.
		addArglist = tailAfterArg3
		// "tailAfterArg3" = residual after arg3 parse = afterArg4.
		insArglist = tailAfterArg4
	} else {
		subcmd = strings.ToLower(arg2)
		valueStr = arg3
		progStr = arg4
		// mpedit add: arg1=victim, arg2=add, arg3=type, arglist starts after arg3 = tailAfterArg4.
		// mpedit insert/edit: arg1=victim, arg2=insert/edit, arg3=number, arg4=type, arglist starts after arg4.
		addArglist = tailAfterArg4
		// For mpedit insert/edit the arglist is even further along — read one
		// more word past arg4 to get the residual.
		_, after5 := util.OneArgument(afterArg4)
		insArglist = strings.TrimSpace(after5)
	}
	value, _ := strconv.Atoi(valueStr) // C atoi returns 0 on parse failure

	// --- dispatch on subcommand ---
	switch subcmd {
	case "list":
		progEditList(ch, target, value, valueStr)
	case "add":
		// add: <type> <arglist>
		// mpedit/opedit: type=arg3 (valueStr), arglist = after arg3 (addArglist).
		// rpedit:        type=arg2 (valueStr), arglist = after arg2 (addArglist).
		progEditAdd(ch, target, valueStr, addArglist)
	case "insert":
		// insert: <pos> <type> <arglist>
		progEditInsert(ch, target, value, progStr, insArglist)
	case "edit":
		// edit: <pos> [type] <arglist> — type optional; if progStr is unknown, treat as no override.
		progEditEdit(ch, target, value, progStr, insArglist)
	case "delete":
		progEditDelete(ch, target, value)
	default:
		progEditSyntaxHelp(ch, kind)
	}
}

func progEditCmdPrefix(kind string) string {
	switch kind {
	case "mob":
		return "mp"
	case "obj":
		return "op"
	case "room":
		return "rp"
	}
	return ""
}

func progEditSyntaxHelp(ch *types.CharData, kind string) {
	prefix := progEditCmdPrefix(kind)
	switch kind {
	case "mob":
		ch.Sendf("Syntax: %sedit <victim> <command> [number] <program> <value>\n\r", prefix)
	case "obj":
		ch.Sendf("Syntax: %sedit <object> <command> [number] <program> <value>\n\r", prefix)
		ch.Send("Object should be in your inventory to edit.\n\r")
	case "room":
		ch.Sendf("Syntax: %sedit <command> [number] <program> <value>\n\r", prefix)
		ch.Send("You should be standing in room you wish to edit.\n\r")
	}
	ch.Send("Commands: add delete insert edit list\n\r")
	ch.Sendf("Programs: %s\n\r", progEditFlavorList(kind))
}

func progEditFlavorList(kind string) string {
	switch kind {
	case "mob":
		// Mob-prog flavors per C build.c mpedit syntax help.
		return "act speech rand fight death hitprcnt entry greet allgreet give bribe hour time wear remove sac script use login void tell imminfo command sell"
	case "obj":
		return "act speech rand wear remove sac zap get drop damage repair greet exa use pull push"
	case "room":
		return "act speech rand sleep rest rfight enter leave death"
	}
	return ""
}

// resolveProgEditTarget performs per-kind lookup + ACL/prototype gates.
// Returns (target, true) on success; on failure emits the appropriate
// message and returns (_, false).
func resolveProgEditTarget(ch *types.CharData, kind, name string) (*progEditorTarget, bool) {
	switch kind {
	case "mob":
		// Trust < LEVEL_GOD: in-room only; >= LEVEL_GOD: world-wide.
		var victim *types.CharData
		if ch.GetTrust() < types.LEVEL_GOD {
			// handler.GetCharRoom takes (ch, name) — uses ch.InRoom.
			victim = lookupCharInRoom(ch, name)
			if victim == nil {
				ch.Send("They aren't here.\n\r")
				return nil, false
			}
		} else {
			victim = lookupCharInWorld(ch, name)
			if victim == nil {
				ch.Send("No one like that in all the realms.\n\r")
				return nil, false
			}
		}
		// Trust + NPC + STATSHIELD gates.
		if ch.GetTrust() < victim.Level || !victim.IsNPC() {
			ch.Send("You can't do that!\n\r")
			return nil, false
		}
		if ch.GetTrust() < types.LEVEL_GREATER && victim.IsNPC() && victim.Act.IsSet(types.ACT_STATSHIELD) {
			ch.Send("Their godly glow prevents you from getting a good look.\n\r")
			return nil, false
		}
		if victim.IndexData == nil {
			ch.Send("Mobile has no prototype data.\n\r")
			return nil, false
		}
		if !victim.IndexData.Act.IsSet(types.ACT_PROTOTYPE) {
			ch.Send("A mobile must have a prototype flag to be mpset.\n\r")
			return nil, false
		}
		return &progEditorTarget{
			kind:       "mob",
			label:      "Mob",
			progs:      &victim.IndexData.MudProgs,
			progTypes:  &victim.IndexData.ProgTypes,
			targetName: victim.ShortDescr,
			targetVnum: victim.IndexData.Vnum,
		}, true

	case "obj":
		var obj *types.ObjData
		if ch.GetTrust() < types.LEVEL_GOD {
			obj = lookupObjCarry(ch, name)
			if obj == nil {
				ch.Send("You aren't carrying that.\n\r")
				return nil, false
			}
		} else {
			obj = lookupObjInWorld(ch, name)
			if obj == nil {
				ch.Send("Nothing like that in all the realms.\n\r")
				return nil, false
			}
		}
		if obj.IndexData == nil {
			ch.Send("Object has no prototype data.\n\r")
			return nil, false
		}
		if !obj.IndexData.ExtraFlags.IsSet(types.ITEM_PROTOTYPE) {
			ch.Send("An object must have a prototype flag to be opset.\n\r")
			return nil, false
		}
		return &progEditorTarget{
			kind:       "obj",
			label:      "Object",
			progs:      &obj.IndexData.MudProgs,
			progTypes:  &obj.IndexData.ProgTypes,
			targetName: obj.ShortDescr,
			targetVnum: obj.IndexData.Vnum,
		}, true

	case "room":
		// rpedit operates on ch.InRoom directly. The 'name' arg here is
		// unused — it was actually arg1 = subcommand (rpedit's arg shape
		// has no target name). Caller passed it for symmetry; ignore.
		_ = name
		if ch.InRoom == nil {
			ch.Send("You aren't in a room.\n\r")
			return nil, false
		}
		room := ch.InRoom
		// can_rmodify / room-prototype gate: rooms have no proto flag in C;
		// the registry-level LEVEL_IMMORTAL trust gate is sufficient.
		return &progEditorTarget{
			kind:       "room",
			label:      "Room",
			progs:      &room.MudProgs,
			progTypes:  &room.ProgTypes,
			targetName: room.Name,
			targetVnum: room.Vnum,
		}, true
	}
	return nil, false
}

// lookupCharInRoom / lookupCharInWorld / lookupObjCarry / lookupObjInWorld
// are thin shims over the handler package — kept here to localise the
// WorldRef dependency at the dispatch layer (matches the wiz.go pattern).

func lookupCharInRoom(ch *types.CharData, name string) *types.CharData {
	if ch.InRoom == nil {
		return nil
	}
	for _, c := range ch.InRoom.People {
		if c == ch {
			continue
		}
		if c.IsNPC() {
			if strings.EqualFold(name, c.ShortDescr) ||
				strings.HasPrefix(strings.ToLower(c.ShortDescr), strings.ToLower(name)) ||
				strings.EqualFold(name, c.Name) ||
				strings.HasPrefix(strings.ToLower(c.Name), strings.ToLower(name)) {
				return c
			}
			// Match against IndexData.PlayerName keyword list (space-separated).
			if c.IndexData != nil && c.IndexData.PlayerName != "" {
				for _, kw := range strings.Fields(c.IndexData.PlayerName) {
					if strings.EqualFold(name, kw) || strings.HasPrefix(strings.ToLower(kw), strings.ToLower(name)) {
						return c
					}
				}
			}
		} else {
			if strings.EqualFold(name, c.Name) || strings.HasPrefix(strings.ToLower(c.Name), strings.ToLower(name)) {
				return c
			}
		}
	}
	return nil
}

func lookupCharInWorld(ch *types.CharData, name string) *types.CharData {
	if WorldRef == nil {
		return nil
	}
	// First try the in-room lookup (cheap, common case).
	if c := lookupCharInRoom(ch, name); c != nil {
		return c
	}
	for _, c := range WorldRef.Characters {
		if c == ch {
			continue
		}
		if c.IsNPC() {
			if strings.EqualFold(name, c.ShortDescr) ||
				strings.HasPrefix(strings.ToLower(c.ShortDescr), strings.ToLower(name)) ||
				strings.EqualFold(name, c.Name) {
				return c
			}
			if c.IndexData != nil && c.IndexData.PlayerName != "" {
				for _, kw := range strings.Fields(c.IndexData.PlayerName) {
					if strings.EqualFold(name, kw) || strings.HasPrefix(strings.ToLower(kw), strings.ToLower(name)) {
						return c
					}
				}
			}
		} else {
			if strings.EqualFold(name, c.Name) || strings.HasPrefix(strings.ToLower(c.Name), strings.ToLower(name)) {
				return c
			}
		}
	}
	return nil
}

func lookupObjCarry(ch *types.CharData, name string) *types.ObjData {
	for _, o := range ch.Carrying {
		if matchObjKeyword(o, name) {
			return o
		}
	}
	return nil
}

func lookupObjInWorld(ch *types.CharData, name string) *types.ObjData {
	if WorldRef == nil {
		return nil
	}
	if o := lookupObjCarry(ch, name); o != nil {
		return o
	}
	for _, o := range WorldRef.Objects {
		if matchObjKeyword(o, name) {
			return o
		}
	}
	return nil
}

func matchObjKeyword(o *types.ObjData, name string) bool {
	if o == nil {
		return false
	}
	lname := strings.ToLower(name)
	if strings.EqualFold(name, o.ShortDescr) || strings.HasPrefix(strings.ToLower(o.ShortDescr), lname) {
		return true
	}
	if strings.EqualFold(name, o.Name) {
		return true
	}
	for _, kw := range strings.Fields(o.Name) {
		if strings.EqualFold(name, kw) || strings.HasPrefix(strings.ToLower(kw), lname) {
			return true
		}
	}
	return false
}

// --- list subcommand (G4) ---

func progEditList(ch *types.CharData, t *progEditorTarget, value int, valueStr string) {
	progs := *t.progs
	if len(progs) == 0 {
		switch t.kind {
		case "mob":
			ch.Sendf("No programs on mobile: %s - #%d\n\r", t.targetName, t.targetVnum)
		case "obj":
			// C-bug verbatim: "mob programs" wording on opedit (build.c:9482).
			ch.Send("That object has no mob programs.\n\r")
		case "room":
			// C-bug verbatim: "mob programs" wording on rpedit (build.c:9821).
			ch.Send("That object has no mob programs.\n\r")
		}
		return
	}

	if value < 1 {
		// Header-only list, with optional 'full' bodies.
		// C uses `if (strcmp("full", arg3))` — non-zero (i.e. NOT equal) =
		// header only. Go-readable form: explicit equality check.
		full := strings.EqualFold(valueStr, "full")
		var sb strings.Builder
		fmt.Fprintf(&sb, "Programs on %s %s (#%d):\n\r", strings.ToLower(t.label), t.targetName, t.targetVnum)
		for i, p := range progs {
			fmt.Fprintf(&sb, "%d>%s %s\n\r", i+1, firstTriggerName(p.Type), p.ArgList)
			if full {
				fmt.Fprintf(&sb, "%s\n\r", p.ComList)
			}
		}
		ch.Send(sb.String())
		return
	}

	// 1-based index into the list.
	if value > len(progs) {
		ch.Send("Program not found.\n\r")
		return
	}
	p := progs[value-1]
	ch.Sendf("%d>%s %s\n\r", value, firstTriggerName(p.Type), p.ArgList)
	ch.Sendf("%s\n\r", p.ComList)
}

// --- add subcommand (G5) ---

// progEditAdd appends a new mud-prog at the tail of the prog list. Mirrors C
// build.c:9354-9373 (do_mpedit add arm). Sets the progtypes bit synchronously
// (the C path also calls xSET_BIT before start_editing) and opens the string
// editor seeded with empty ComList. The EditorSave closure copies the buffer
// into mprg.ComList on /s — no progtypes rebuild (the bit was set above).
func progEditAdd(ch *types.CharData, t *progEditorTarget, progName, argument string) {
	mptype, ok := util.GetMpFlag(progName)
	if !ok {
		ch.Send("Unknown program type.\n\r")
		return
	}
	mprg := &types.MProgData{
		Type:    mptype,
		ArgList: util.SmashTilde(argument),
		ComList: "",
	}
	*t.progs = append(*t.progs, mprg)
	t.progTypes.Set(progBitIndex(mptype))
	progEditOpenEditor(ch, mprg, t, false /* no rebuild */)
}

// --- insert subcommand (G6) ---

// progEditInsert splices a new mud-prog at 1-based position `value`. Mirrors C
// build.c:9308-9352. C's `&& mprg->next` guard at :9340 means insertion at
// the last-element position is rejected ("Program not found.") — append
// requires the `add` subcommand. Q1/Q2 fix applied for rpedit (callers route
// here regardless of arg-shape; the dispatcher already normalised).
func progEditInsert(ch *types.CharData, t *progEditorTarget, value int, progName, argument string) {
	if len(*t.progs) == 0 {
		switch t.kind {
		case "mob":
			ch.Sendf("No programs on mobile: %s - #%d\n\r", t.targetName, t.targetVnum)
		default:
			ch.Send("That object has no mob programs.\n\r")
		}
		return
	}
	mptype, ok := util.GetMpFlag(progName)
	if !ok {
		ch.Send("Unknown program type.\n\r")
		return
	}
	if value < 1 {
		ch.Send("Program not found.\n\r")
		return
	}
	mprg := &types.MProgData{
		Type:    mptype,
		ArgList: util.SmashTilde(argument),
		ComList: "",
	}
	progs := *t.progs
	if value == 1 {
		// Splice at head.
		newProgs := make([]*types.MProgData, 0, len(progs)+1)
		newProgs = append(newProgs, mprg)
		newProgs = append(newProgs, progs...)
		*t.progs = newProgs
	} else {
		// Splice after the (value-1)-th prog (1-based). C's `&& mprg->next`
		// guard rejects insertion at-end-of-list (use `add` instead).
		if value-1 >= len(progs) {
			ch.Send("Program not found.\n\r")
			return
		}
		newProgs := make([]*types.MProgData, 0, len(progs)+1)
		newProgs = append(newProgs, progs[:value-1]...)
		newProgs = append(newProgs, mprg)
		newProgs = append(newProgs, progs[value-1:]...)
		*t.progs = newProgs
	}
	t.progTypes.Set(progBitIndex(mptype))
	progEditOpenEditor(ch, mprg, t, false /* no rebuild */)
}

// --- edit / delete stubs (filled in by Wave 4) ---

func progEditEdit(ch *types.CharData, t *progEditorTarget, value int, progName, argument string) {
	_ = t
	_ = value
	_ = progName
	_ = argument
	ch.Send("mpedit edit: not yet implemented.\n\r")
}

func progEditDelete(ch *types.CharData, t *progEditorTarget, value int) {
	_ = t
	_ = value
	ch.Send("mpedit delete: not yet implemented.\n\r")
}

// --- editor closure (shared by add / insert / edit) ---

// progEditOpenEditor sets ch.Substate = SUB_MPROG_EDIT, installs an
// EditorSave closure that writes the buffer into mprg.ComList on /s, and
// invokes StartEditingFunc seeded with the existing ComList. When
// rebuildOnSave is true (edit path only, per C build.c:9234-9236), the
// closure also clears + rebuilds the progtypes BitVector by iterating
// the post-edit prog list — matching C `xCLEAR_BITS; for (...) xSET_BIT`.
func progEditOpenEditor(ch *types.CharData, mprg *types.MProgData, t *progEditorTarget, rebuildOnSave bool) {
	ch.Substate = types.SUB_MPROG_EDIT
	captured := mprg
	target := t
	rebuild := rebuildOnSave
	ch.EditorSave = func(c *types.CharData) {
		if CopyBufferFunc != nil {
			captured.ComList = CopyBufferFunc(c)
		}
		if StopEditingFunc != nil {
			StopEditingFunc(c)
		}
		if rebuild && target != nil && target.progTypes != nil && target.progs != nil {
			target.progTypes.Clear()
			for _, p := range *target.progs {
				target.progTypes.Set(progBitIndex(p.Type))
			}
		}
		c.Send("\n\r")
	}
	if StartEditingFunc != nil {
		StartEditingFunc(ch, mprg.ComList)
	}
}

// progBitIndex converts a bit-flag value (1<<n) to its bit-index n.
// Used at the BitVector boundary (ProgTypes is indexed by bit-position,
// not by bit-flag value).
func progBitIndex(flag int64) int {
	return bits.TrailingZeros64(uint64(flag))
}
