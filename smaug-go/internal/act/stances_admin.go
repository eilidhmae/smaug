// Stances OLC commands — DoSTstat (display) + DoSTset (editor).
// Ports C do_ststat (src/stances.c:672-735) and do_stset (:742-1010).
package act

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// DoSTstat displays one stance's full field set, or the stance-list
// header when called with no argument. Mirrors C do_ststat
// (src/stances.c:672-735).
//
// No-argument: prints "STANCE list:" then walks 0..MAX_STANCE-1 printing
// each stance name. Go fixes C's operator-precedence bug at line 687
// (`!index_num % 4` parses as `(!index_num) % 4` == 0 always-false) and
// emits a newline every 4 stances as the C comment clearly intended.
//
// With argument: resolves via GetStanceNumber; on unknown, prints an
// error + help reference. On success, emits 13 display lines matching
// the C layout intent (some lines condensed for brevity, same field
// coverage). Class/Race restrictions render as raw hex masks — a
// human-friendly formatter is a follow-up (plan §Scope Cuts).
func DoSTstat(ch *types.CharData, argument string) {
	arg, _ := util.OneArgument(argument)
	if arg == "" {
		ch.Send("STANCE list:\n\r")
		for i := 0; i < types.MAX_STANCE; i++ {
			name := combat.GetStanceName(i)
			if name == "" {
				continue
			}
			ch.Sendf("%s ", name)
			// Fixed-from-C: newline every 4 stances. C bug at :687.
			if i != 0 && i%4 == 0 {
				ch.Send("\n\r")
			}
		}
		ch.Send("\n\r")
		return
	}

	idx := persist.GetStanceNumber(arg)
	if idx < 0 || idx >= types.MAX_STANCE {
		ch.Send("That stance does not exist.\n\r")
		ch.Send("(See help STANCEFLAGS.)\n\r")
		return
	}
	info := combat.StanceIndex[idx]
	ch.Sendf("Name            : %s\n\r", combat.GetStanceName(idx))
	ch.Sendf("Self            : %s\n\r", info.Self)
	ch.Sendf("Other           : %s\n\r", info.Others)
	ch.Sendf("Required Stances: %s and %s\n\r",
		combat.GetStanceName(info.Prereq[0]),
		combat.GetStanceName(info.Prereq[1]))
	ch.Sendf("Dual Wield      : %d\n\r", info.Dual)
	ch.Sendf("Damage Done     : %d     Damage Taken   : %d\n\r", info.DamDone, info.DamTaken)
	ch.Sendf("Dodge           : %d     Parry          : %d\n\r", info.Dodge, info.Parry)
	ch.Sendf("Attacks         : %d     Max Weight     : %d\n\r", info.NumAttacks, info.MaxWeight)
	ch.Sendf("Wait            : %d\n\r", info.Wait)
	ch.Sendf("Resistance      : %s\n\r", util.FlagString(info.Resist, util.RisflagNames))
	ch.Sendf("Susceptible     : %s\n\r", util.FlagString(info.Suscept, util.RisflagNames))
	ch.Sendf("Immune          : %s\n\r", util.FlagString(info.Immune, util.RisflagNames))
	ch.Sendf("Special Move    : %s     Chance         : %d\n\r",
		combat.GetSpecialName(info.SpecialMove), info.SpecialPercent)
	ch.Sendf("Class Restrictions: 0x%x\n\r", info.Class)
	ch.Sendf("Race Restrictions : 0x%x\n\r", info.Race)
}

// stsetUsage prints the C-equivalent "?" / no-args syntax block.
func stsetUsage(ch *types.CharData) {
	ch.Send("Syntax: stset <stance> <field> <value>\n\r")
	ch.Send("Syntax: stset save\n\r")
	ch.Send("Field being one of:\n\r")
	ch.Send("  attacks class damage dodge dual immune others parry\n\r")
	ch.Send("  percent protection race resist self special susceptible\n\r")
	ch.Send("  stance1 stance2 wait weight\n\r")
}

// DoSTset edits a stance-table field at runtime. Mirrors C do_stset
// (src/stances.c:742-1010).
//
// Structure:
//  1. NPC → "Mob's can't mset" (C-typo preserved).
//  2. No descriptor → reject.
//  3. SmashTilde on argument.
//  4. `stset save` → persist.SaveStances.
//  5. No/"?" args → usage.
//  6. Unknown stance name → error + help reference.
//  7. Field switch (18 branches). Preserves 4 C bugs with inline refs.
//
// Preserved C bugs (plan-phase6-stances-olc.md §Q2-Q4):
//   - `class <N>`  : matched but no assignment (stances.c:823-826).
//   - `race <N>`   : matched but no assignment (stances.c:918-921).
//   - `dual 0|1`   : stores inverted value (stances.c:858-861).
//   - `special`    : get_special_number stub returns 0 (stances.c:1012).
func DoSTset(ch *types.CharData, argument string) {
	if ch.IsNPC() {
		ch.Send("Mob's can't mset\n\r")
		return
	}
	if ch.Desc == nil {
		ch.Send("You have no descriptor\n\r")
		return
	}
	argument = util.SmashTilde(argument)

	arg1, rest := util.OneArgument(argument)
	arg2, arg3 := util.OneArgument(rest)

	if strings.EqualFold(arg1, "save") {
		// Resolve the path the boot-time loader used.
		path := persist.StancePath
		if path == "" && WorldRef != nil && WorldRef.DataDir != "" {
			path = filepath.Join(WorldRef.DataDir, "system", "stances.dat")
		}
		if path == "" {
			ch.Send("Cannot determine stances file path.\n\r")
			return
		}
		if err := persist.SaveStances(&combat.StanceIndex, path); err != nil {
			ch.Sendf("Save failed: %v\n\r", err)
			return
		}
		ch.Send("Done.\n\r")
		return
	}

	if arg1 == "" || arg2 == "" || arg1 == "?" {
		stsetUsage(ch)
		return
	}

	idx := persist.GetStanceNumber(arg1)
	if idx < 0 || idx >= types.MAX_STANCE {
		ch.Send("No such stance name.\n\r")
		ch.Send("(See help STANCEFLAGS.)\n\r")
		return
	}

	// C: value = is_number(arg3) ? atoi(arg3) : -100. The sentinel -100
	// guarantees out-of-range on any guard that uses `< min || > max`.
	value := -100
	if util.IsNumber(arg3) {
		// Local parse — util.IsNumber already validated the shape.
		if _, err := fmt.Sscanf(arg3, "%d", &value); err != nil {
			value = -100
		}
	}

	field := strings.ToLower(arg2)
	found := true
	switch field {
	case "attacks":
		if value < -5 || value > 5 {
			ch.Send("Attacks range is -5 to 5.\n\r")
			return
		}
		combat.StanceIndex[idx].NumAttacks = value
	case "class":
		// C bug preserved: stances.c:823-826 sets `found = TRUE` with
		// no assignment. Silent no-op.
		_ = value
	case "damage":
		if value < 0 || value > 200 {
			ch.Send("Damage range is 0 to 200.\n\r")
			return
		}
		combat.StanceIndex[idx].DamDone = value
	case "dodge":
		if value < -50 || value > 50 {
			ch.Send("Dodge range is -50 to 50.\n\r")
			return
		}
		combat.StanceIndex[idx].Dodge = value
	case "dual":
		if value < 0 || value > 1 {
			ch.Send("Dual range is 0 or 1.\n\r")
			return
		}
		// C bug preserved: stances.c:858-861. `value=1` stores 0,
		// `value=0` stores 1. Inverted on purpose (to match shipped
		// data already calibrated around the bug).
		if value != 0 {
			combat.StanceIndex[idx].Dual = 0
		} else {
			combat.StanceIndex[idx].Dual = 1
		}
	case "immune":
		bit := util.GetRisflag(arg3)
		if bit < 0 {
			ch.Send("Unkown flag\n\r") // C-typo preserved.
			return
		}
		combat.StanceIndex[idx].Immune ^= 1 << bit
	case "others":
		combat.StanceIndex[idx].Others = arg3
	case "parry":
		if value < -50 || value > 50 {
			ch.Send("Parry range is -50 to 50.\n\r")
			return
		}
		combat.StanceIndex[idx].Parry = value
	case "percent":
		if value < 0 || value > 100 {
			ch.Send("Percent range is 0 to 100.\n\r")
			return
		}
		combat.StanceIndex[idx].SpecialPercent = value
	case "protection":
		if value < 0 || value > 200 {
			ch.Send("Protection range is 0 to 200.\n\r")
			return
		}
		combat.StanceIndex[idx].DamTaken = value
	case "race":
		// C bug preserved: stances.c:918-921 silent no-op.
		_ = value
	case "resist":
		bit := util.GetRisflag(arg3)
		if bit < 0 {
			ch.Send("Unkown flag\n\r") // C-typo preserved.
			return
		}
		combat.StanceIndex[idx].Resist ^= 1 << bit
	case "self":
		combat.StanceIndex[idx].Self = arg3
	case "special":
		// C stub: get_special_number returns 0 unconditionally.
		combat.StanceIndex[idx].SpecialMove = combat.GetSpecialNumber(arg3)
	case "susceptible":
		bit := util.GetRisflag(arg3)
		if bit < 0 {
			ch.Send("Unkown flag\n\r")
			return
		}
		combat.StanceIndex[idx].Suscept ^= 1 << bit
	case "stance1":
		n := persist.GetStanceNumber(arg3)
		if n < 0 {
			ch.Send("No such stance name.\n\r")
			return
		}
		combat.StanceIndex[idx].Prereq[0] = n
	case "stance2":
		n := persist.GetStanceNumber(arg3)
		if n < 0 {
			ch.Send("No such stance name.\n\r")
			return
		}
		combat.StanceIndex[idx].Prereq[1] = n
	case "wait":
		if value < 0 || value > 25 {
			ch.Send("Wait range is 0 to 25.\n\r")
			return
		}
		combat.StanceIndex[idx].Wait = value
	case "weight":
		if value < 0 || value > 999 {
			ch.Send("Weight range is 0 to 999.\n\r")
			return
		}
		combat.StanceIndex[idx].MaxWeight = value
	default:
		found = false
	}

	if !found {
		stsetUsage(ch)
		return
	}
	ch.Send("Done.\n\r")
}
