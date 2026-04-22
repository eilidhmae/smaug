// Package game — medit Wave 3 G7 simple-field arm bodies.
//
// Each arm reads one line of builder input, validates (clamp / name
// lookup / empty-to-clear), assigns to the victim's field, emits an
// olcLog trail, marks OlcData.Change, and redisplays the correct main
// menu (NPC or PC) via npcOrPcMenu + MeditDispMenu.
//
// NPC-only / PC-only enforcement follows plan-phase6-olc-medit.md
// §313-347. Rejection message is the Go-idiom form — the verbatim
// "NPC Only!!" / "NPCs Only!!" strings live only on the PC-main-menu
// digit-6 / digit-U direct rejects (plan §C Bug Catalog #5).
//
// C reference: src/omedit.c:1389-1700 (per-arm citations in each arm
// comment below).
//
// Plan: plan-phase6-olc-medit.md §G7 (A14, A15).
package game

import (
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// meditFinishArm is the post-toggle tail every G7/G8 simple arm calls:
// sets OlcData.Change, flips to the appropriate main-menu mode, and
// redisplays it. Mirrors C src/omedit.c:2276-2277.
func meditFinishArm(d *types.DescriptorData, victim *types.CharData) {
	if d.Olc != nil {
		d.Olc.Change = true
		d.Olc.Mode = npcOrPcMenu(victim)
	}
	MeditDispMenu(d)
}

// meditRejectNpcOnly — called when victim IS an NPC but the requested
// field is PC-only (e.g. MEDIT_PRACTICE). Emits "doesn't apply to NPCs"
// rejection and redisplays the NPC main menu.
func meditRejectNpcOnly(d *types.DescriptorData, victim *types.CharData) {
	d.WriteToBuffer("That field doesn't apply to NPCs.\n\r")
	if d.Olc != nil {
		d.Olc.Mode = types.MEDIT_NPC_MAIN_MENU
	}
	MeditDispMenu(d)
}

// meditRejectPcOnly — called when victim is a PC but the requested
// field is NPC-only (e.g. MEDIT_DAMNUMDIE). Emits "doesn't apply to
// players" rejection and redisplays the PC main menu.
func meditRejectPcOnly(d *types.DescriptorData, victim *types.CharData) {
	d.WriteToBuffer("That field doesn't apply to players.\n\r")
	if d.Olc != nil {
		d.Olc.Mode = types.MEDIT_PC_MAIN_MENU
	}
	MeditDispMenu(d)
}

// parseInt parses arg as int; returns 0 on failure (C atoi semantics).
func parseInt(arg string) int {
	n, err := strconv.Atoi(strings.TrimSpace(arg))
	if err != nil {
		return 0
	}
	return n
}

// meditArmName handles MEDIT_NAME. C omedit.c:1389-1405.
// For PC with trust > LEVEL_SUB_IMPLEM-1, C invokes do_pcrename to
// rename the pfile on disk. Go port falls back to direct name assignment
// with a TODO entry (see TODO-updates.md) because DoPcrename does not
// yet exist. SmashTilde applied before assignment.
func meditArmName(d *types.DescriptorData, victim *types.CharData, arg string) {
	arg = util.SmashTilde(strings.TrimSpace(arg))
	if arg == "" {
		d.WriteToBuffer("Name cannot be empty.\n\r")
		meditFinishArm(d, victim)
		return
	}
	victim.Name = arg
	if victim.IsNPC() && victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
		victim.IndexData.PlayerName = victim.Name
	}
	olcLog(d, "MOB", "Changed name to %s", arg)
	meditFinishArm(d, victim)
}

// meditArmShortDesc handles MEDIT_S_DESC (NPC-only per menu). C omedit.c:1407-1416.
func meditArmShortDesc(d *types.DescriptorData, victim *types.CharData, arg string) {
	if !victim.IsNPC() {
		meditRejectPcOnly(d, victim)
		return
	}
	arg = util.SmashTilde(arg)
	victim.ShortDescr = arg
	if victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
		victim.IndexData.ShortDescr = arg
	}
	olcLog(d, "MOB", "Changed short desc to %s", arg)
	meditFinishArm(d, victim)
}

// meditArmLongDesc handles MEDIT_L_DESC (NPC-only per menu). C omedit.c:1418-1429.
// C appends "\n\r" to the input unconditionally — Go preserves that.
func meditArmLongDesc(d *types.DescriptorData, victim *types.CharData, arg string) {
	if !victim.IsNPC() {
		meditRejectPcOnly(d, victim)
		return
	}
	arg = util.SmashTilde(arg)
	victim.LongDescr = arg + "\n\r"
	if victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
		victim.IndexData.LongDescr = victim.LongDescr
	}
	olcLog(d, "MOB", "Changed long desc to %s", arg)
	meditFinishArm(d, victim)
}

// meditArmDDescBugTrap handles MEDIT_D_DESC parse-arm entry. Per
// C omedit.c:1431-1435 this is an unreachable bug-trap — the editor
// trampoline (launched from main-menu digit 5 NPC / digit 3 PC) sets
// MEDIT_D_DESC as the mode marker but the /s handler invokes the
// closure directly on CharData.EditorSave; no parse-input lines should
// ever arrive here.
//
// If a line arrives anyway: bug-log and clean up (matches C behavior).
// Plan §G7 + §C Bug Catalog #1.
func meditArmDDescBugTrap(d *types.DescriptorData, victim *types.CharData, arg string) {
	util.Bug("medit_parse: reached D_DESC case (unreachable bug-trap)")
	cleanupOlc(d)
}

// meditArmSex handles MEDIT_SEX. C omedit.c:1714-1719 uses
// URANGE(0, atoi(arg), 2) — silent clamp, no rejection. Wave 3
// follow-up LOW #6 (2026-04-22): switched from rejection-on-OOB to
// C-parity silent clamp.
func meditArmSex(d *types.DescriptorData, victim *types.CharData, arg string) {
	victim.Sex = uRange(0, parseInt(arg), 2)
	if victim.IsNPC() && victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
		victim.IndexData.Sex = victim.Sex
	}
	var label string
	switch victim.Sex {
	case types.SEX_MALE:
		label = "Male"
	case types.SEX_FEMALE:
		label = "Female"
	default:
		label = "Neutral"
	}
	olcLog(d, "MOB", "Changed sex to %s", label)
	meditFinishArm(d, victim)
}

// meditArmHitroll handles MEDIT_HITROLL. C omedit.c:1721-1726. URANGE(0,x,85).
func meditArmHitroll(d *types.DescriptorData, victim *types.CharData, arg string) {
	victim.Hitroll = uRange(0, parseInt(arg), 85)
	if victim.IsNPC() && victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
		victim.IndexData.Hitroll = victim.Hitroll
	}
	olcLog(d, "MOB", "Changed hitroll to %d", victim.Hitroll)
	meditFinishArm(d, victim)
}

// meditArmDamroll handles MEDIT_DAMROLL. C omedit.c:1728-1733. URANGE(0,x,65).
func meditArmDamroll(d *types.DescriptorData, victim *types.CharData, arg string) {
	victim.Damroll = uRange(0, parseInt(arg), 65)
	if victim.IsNPC() && victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
		victim.IndexData.Damroll = victim.Damroll
	}
	olcLog(d, "MOB", "Changed damroll to %d", victim.Damroll)
	meditFinishArm(d, victim)
}

// meditArmDamNumDie handles MEDIT_DAMNUMDIE (NPC-only). C :1735-1739.
// Target is pIndexData.damnodice — only writes when prototype + NPC.
func meditArmDamNumDie(d *types.DescriptorData, victim *types.CharData, arg string) {
	if !victim.IsNPC() {
		meditRejectPcOnly(d, victim)
		return
	}
	if victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
		victim.IndexData.DamNoDice = uRange(0, parseInt(arg), 100)
		olcLog(d, "MOB", "Changed damnumdie to %d", victim.IndexData.DamNoDice)
	}
	meditFinishArm(d, victim)
}

// meditArmDamSizeDie handles MEDIT_DAMSIZEDIE. C :1741-1745.
func meditArmDamSizeDie(d *types.DescriptorData, victim *types.CharData, arg string) {
	if !victim.IsNPC() {
		meditRejectPcOnly(d, victim)
		return
	}
	if victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
		victim.IndexData.DamSizeDice = uRange(0, parseInt(arg), 100)
		olcLog(d, "MOB", "Changed damsizedie to %d", victim.IndexData.DamSizeDice)
	}
	meditFinishArm(d, victim)
}

// meditArmDamPlus handles MEDIT_DAMPLUS. C :1747-1751. URANGE(0,x,1000).
func meditArmDamPlus(d *types.DescriptorData, victim *types.CharData, arg string) {
	if !victim.IsNPC() {
		meditRejectPcOnly(d, victim)
		return
	}
	if victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
		victim.IndexData.DamPlus = uRange(0, parseInt(arg), 1000)
		olcLog(d, "MOB", "Changed damplus to %d", victim.IndexData.DamPlus)
	}
	meditFinishArm(d, victim)
}

// meditArmHitNumDie handles MEDIT_HITNUMDIE. C :1753-1757. URANGE(0,x,32767).
func meditArmHitNumDie(d *types.DescriptorData, victim *types.CharData, arg string) {
	if !victim.IsNPC() {
		meditRejectPcOnly(d, victim)
		return
	}
	if victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
		victim.IndexData.HitNoDice = uRange(0, parseInt(arg), 32767)
		olcLog(d, "MOB", "Changed hitnumdie to %d", victim.IndexData.HitNoDice)
	}
	meditFinishArm(d, victim)
}

// meditArmHitSizeDie handles MEDIT_HITSIZEDIE. C :1759-1763. URANGE(0,x,30000).
func meditArmHitSizeDie(d *types.DescriptorData, victim *types.CharData, arg string) {
	if !victim.IsNPC() {
		meditRejectPcOnly(d, victim)
		return
	}
	if victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
		victim.IndexData.HitSizeDice = uRange(0, parseInt(arg), 30000)
		olcLog(d, "MOB", "Changed hitsizedie to %d", victim.IndexData.HitSizeDice)
	}
	meditFinishArm(d, victim)
}

// meditArmHitPlus handles MEDIT_HITPLUS. C :1765-1769. URANGE(0,x,30000).
func meditArmHitPlus(d *types.DescriptorData, victim *types.CharData, arg string) {
	if !victim.IsNPC() {
		meditRejectPcOnly(d, victim)
		return
	}
	if victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
		victim.IndexData.HitPlus = uRange(0, parseInt(arg), 30000)
		olcLog(d, "MOB", "Changed hitplus to %d", victim.IndexData.HitPlus)
	}
	meditFinishArm(d, victim)
}

// meditArmArmor handles MEDIT_AC. C :1771-1774. URANGE(-300,x,300).
func meditArmArmor(d *types.DescriptorData, victim *types.CharData, arg string) {
	victim.Armor = uRange(-300, parseInt(arg), 300)
	olcLog(d, "MOB", "Changed armor to %d", victim.Armor)
	meditFinishArm(d, victim)
}

// meditArmGold handles MEDIT_GOLD. C :1776-1779. UMAX(0,x).
func meditArmGold(d *types.DescriptorData, victim *types.CharData, arg string) {
	n := parseInt(arg)
	if n < 0 {
		n = 0
	}
	victim.Gold = n
	olcLog(d, "MOB", "Changed gold to %d", victim.Gold)
	meditFinishArm(d, victim)
}

// meditArmSilver handles MEDIT_SILVER (GSC only). C :1782-1785.
func meditArmSilver(d *types.DescriptorData, victim *types.CharData, arg string) {
	n := parseInt(arg)
	if n < 0 {
		n = 0
	}
	victim.Silver = n
	olcLog(d, "MOB", "Changed silver to %d", victim.Silver)
	meditFinishArm(d, victim)
}

// meditArmCopper handles MEDIT_COPPER. C :1787-1790.
func meditArmCopper(d *types.DescriptorData, victim *types.CharData, arg string) {
	n := parseInt(arg)
	if n < 0 {
		n = 0
	}
	victim.Copper = n
	olcLog(d, "MOB", "Changed copper to %d", victim.Copper)
	meditFinishArm(d, victim)
}

// meditArmPosition handles MEDIT_POS. C :1793-1796. URANGE(0,x,POS_STANDING).
func meditArmPosition(d *types.DescriptorData, victim *types.CharData, arg string) {
	victim.Position = uRange(0, parseInt(arg), types.POS_STANDING)
	olcLog(d, "MOB", "Changed position to %d", victim.Position)
	meditFinishArm(d, victim)
}

// meditArmDefaultPosition handles MEDIT_DEFAULT_POS (NPC-only). C :1798-1801.
func meditArmDefaultPosition(d *types.DescriptorData, victim *types.CharData, arg string) {
	if !victim.IsNPC() {
		meditRejectPcOnly(d, victim)
		return
	}
	victim.DefPosition = uRange(0, parseInt(arg), types.POS_STANDING)
	olcLog(d, "MOB", "Changed default position to %d", victim.DefPosition)
	meditFinishArm(d, victim)
}

// meditArmLevel handles MEDIT_LEVEL. C :1988-1991. URANGE(1,x,MAX_LEVEL-1).
// For PC target C does NOT require LEVEL_GREATER at parse-time (only at
// PC-menu digit time — see plan §Q6); Go matches C.
func meditArmLevel(d *types.DescriptorData, victim *types.CharData, arg string) {
	victim.Level = uRange(1, parseInt(arg), types.MAX_LEVEL-1)
	olcLog(d, "MOB", "Changed level to %d", victim.Level)
	meditFinishArm(d, victim)
}

// meditArmAlignment handles MEDIT_ALIGNMENT. C :1993-1996. URANGE(-1000,x,1000).
func meditArmAlignment(d *types.DescriptorData, victim *types.CharData, arg string) {
	victim.Alignment = uRange(-1000, parseInt(arg), 1000)
	olcLog(d, "MOB", "Changed alignment to %d", victim.Alignment)
	meditFinishArm(d, victim)
}

// meditArmFavor handles MEDIT_FAVOR (PC-only). C :1828-1831. URANGE(-2500,x,2500).
// Target is pcdata.favor.
func meditArmFavor(d *types.DescriptorData, victim *types.CharData, arg string) {
	if victim.IsNPC() {
		meditRejectNpcOnly(d, victim)
		return
	}
	if victim.PCData == nil {
		d.WriteToBuffer("Victim has no PCData.\n\r")
		meditFinishArm(d, victim)
		return
	}
	victim.PCData.Favor = uRange(-2500, parseInt(arg), 2500)
	olcLog(d, "MOB", "Changed favor to %d", victim.PCData.Favor)
	meditFinishArm(d, victim)
}

// meditArmHitpoint handles MEDIT_HITPOINT. C :1581-1584. URANGE(1,x,32700).
func meditArmHitpoint(d *types.DescriptorData, victim *types.CharData, arg string) {
	victim.MaxHit = uRange(1, parseInt(arg), 32700)
	olcLog(d, "MOB", "Changed hitpoints to %d", victim.MaxHit)
	meditFinishArm(d, victim)
}

// meditArmMana handles MEDIT_MANA. C :1586-1589. URANGE(1,x,30000).
func meditArmMana(d *types.DescriptorData, victim *types.CharData, arg string) {
	victim.MaxMana = uRange(1, parseInt(arg), 30000)
	olcLog(d, "MOB", "Changed mana to %d", victim.MaxMana)
	meditFinishArm(d, victim)
}

// meditArmMove handles MEDIT_MOVE. C :1591-1594. URANGE(1,x,30000).
func meditArmMove(d *types.DescriptorData, victim *types.CharData, arg string) {
	victim.MaxMove = uRange(1, parseInt(arg), 30000)
	olcLog(d, "MOB", "Changed moves to %d", victim.MaxMove)
	meditFinishArm(d, victim)
}

// meditArmPractice handles MEDIT_PRACTICE (PC-only). C :1596-1599.
// URANGE(1,x,300) per C — plan had (0,250) and flagged the correction.
func meditArmPractice(d *types.DescriptorData, victim *types.CharData, arg string) {
	if victim.IsNPC() {
		meditRejectNpcOnly(d, victim)
		return
	}
	victim.Practice = uRange(1, parseInt(arg), 300)
	olcLog(d, "MOB", "Changed practices to %d", victim.Practice)
	meditFinishArm(d, victim)
}

// meditArmMentalState handles MEDIT_MENTALSTATE. C :1803-1806. URANGE(-100,x,100).
func meditArmMentalState(d *types.DescriptorData, victim *types.CharData, arg string) {
	victim.MentalState = uRange(-100, parseInt(arg), 100)
	olcLog(d, "MOB", "Changed mental state to %d", victim.MentalState)
	meditFinishArm(d, victim)
}

// meditArmEmotional handles MEDIT_EMOTIONAL. C :1808-1811.
func meditArmEmotional(d *types.DescriptorData, victim *types.CharData, arg string) {
	victim.EmotionalState = uRange(-100, parseInt(arg), 100)
	olcLog(d, "MOB", "Changed emotional state to %d", victim.EmotionalState)
	meditFinishArm(d, victim)
}

// meditArmThirst handles MEDIT_THIRST (PC-only). C :1813-1816. Target
// is pcdata.condition[COND_THIRST]. URANGE(0,x,100).
func meditArmThirst(d *types.DescriptorData, victim *types.CharData, arg string) {
	if victim.IsNPC() {
		meditRejectNpcOnly(d, victim)
		return
	}
	if victim.PCData == nil {
		d.WriteToBuffer("Victim has no PCData.\n\r")
		meditFinishArm(d, victim)
		return
	}
	victim.PCData.Condition[types.COND_THIRST] = uRange(0, parseInt(arg), 100)
	olcLog(d, "MOB", "Changed thirst to %d", victim.PCData.Condition[types.COND_THIRST])
	meditFinishArm(d, victim)
}

// meditArmFull handles MEDIT_FULL (PC-only). C :1818-1821.
func meditArmFull(d *types.DescriptorData, victim *types.CharData, arg string) {
	if victim.IsNPC() {
		meditRejectNpcOnly(d, victim)
		return
	}
	if victim.PCData == nil {
		d.WriteToBuffer("Victim has no PCData.\n\r")
		meditFinishArm(d, victim)
		return
	}
	victim.PCData.Condition[types.COND_FULL] = uRange(0, parseInt(arg), 100)
	olcLog(d, "MOB", "Changed hunger to %d", victim.PCData.Condition[types.COND_FULL])
	meditFinishArm(d, victim)
}

// meditArmDrunk handles MEDIT_DRUNK (PC-only). C :1823-1826.
func meditArmDrunk(d *types.DescriptorData, victim *types.CharData, arg string) {
	if victim.IsNPC() {
		meditRejectNpcOnly(d, victim)
		return
	}
	if victim.PCData == nil {
		d.WriteToBuffer("Victim has no PCData.\n\r")
		meditFinishArm(d, victim)
		return
	}
	victim.PCData.Condition[types.COND_DRUNK] = uRange(0, parseInt(arg), 100)
	olcLog(d, "MOB", "Changed drunkness to %d", victim.PCData.Condition[types.COND_DRUNK])
	meditFinishArm(d, victim)
}

// meditArmAttack handles MEDIT_ATTACK (NPC-only). C :1916-1950.
// Toggle bit in victim.Attacks (BitVector). Single-token numeric uses
// offset (arg-1), "0" exits to menu; keyword lookup via attack_flags.
// Accepts space-separated tokens (C parity).
func meditArmAttack(d *types.DescriptorData, victim *types.CharData, arg string) {
	if !victim.IsNPC() {
		meditRejectPcOnly(d, victim)
		return
	}
	arg = strings.TrimSpace(arg)
	if arg == "" || strings.EqualFold(arg, "done") || strings.EqualFold(arg, "quit") {
		meditFinishArm(d, victim)
		return
	}
	tokens := strings.Fields(arg)
	for _, tok := range tokens {
		if n, err := strconv.Atoi(tok); err == nil {
			if n == 0 {
				meditFinishArm(d, victim)
				return
			}
			idx := n - 1
			if idx < 0 || idx >= types.MAX_ATTACK_TYPE {
				d.WriteToBuffer("Invalid flag, try again: ")
				return
			}
			victim.Attacks.Toggle(idx)
			olcLog(d, "MOB", "Toggled attack bit %d", idx)
		} else {
			// No Go-side attack_flags name table yet — reject keyword
			// input with a TODO-follow-up message and ask for numeric.
			// Plan: flag as follow-up in TODO-updates.md.
			d.WriteToBuffer("Numeric attack index only (flag-name lookup pending). Try again: ")
			return
		}
	}
	if victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
		victim.IndexData.Attacks = victim.Attacks
	}
	meditFinishArm(d, victim)
}

// meditArmDefense handles MEDIT_DEFENSE (NPC-only). C :1952-1986.
func meditArmDefense(d *types.DescriptorData, victim *types.CharData, arg string) {
	if !victim.IsNPC() {
		meditRejectPcOnly(d, victim)
		return
	}
	arg = strings.TrimSpace(arg)
	if arg == "" || strings.EqualFold(arg, "done") || strings.EqualFold(arg, "quit") {
		meditFinishArm(d, victim)
		return
	}
	tokens := strings.Fields(arg)
	for _, tok := range tokens {
		if n, err := strconv.Atoi(tok); err == nil {
			if n == 0 {
				meditFinishArm(d, victim)
				return
			}
			idx := n - 1
			if idx < 0 || idx >= types.MAX_DEFENSE_TYPE {
				d.WriteToBuffer("Invalid flag, try again: ")
				return
			}
			victim.Defenses.Toggle(idx)
			olcLog(d, "MOB", "Toggled defense bit %d", idx)
		} else {
			d.WriteToBuffer("Numeric defense index only (flag-name lookup pending). Try again: ")
			return
		}
	}
	if victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
		victim.IndexData.Defenses = victim.Defenses
	}
	meditFinishArm(d, victim)
}

// meditSpecNames maps the numeric pick index to the spec-function name.
// Mirrors C omedit.c:2104-2162. Index 0 clears. Unknown entries keep
// the current spec.
var meditSpecNames = []string{
	"", // 0 = clear
	"spec_breath_any", "spec_breath_acid", "spec_breath_fire",
	"spec_breath_frost", "spec_breath_gas", "spec_breath_lightning",
	"spec_cast_adept", "spec_cast_cleric", "spec_cast_mage",
	"spec_cast_undead", "spec_executioner", "spec_fido",
	"spec_guard", "spec_janitor", "spec_poison", "spec_thief",
}

// meditArmSpec handles MEDIT_SPEC (NPC-only). C :2104-2163.
// Numeric index 0..16 picks a spec_fun name; unknown index is ignored.
func meditArmSpec(d *types.DescriptorData, victim *types.CharData, arg string) {
	if !victim.IsNPC() {
		meditRejectPcOnly(d, victim)
		return
	}
	n := parseInt(arg)
	if n < 0 || n >= len(meditSpecNames) {
		// Unknown — keep current silently (matches C's no-op default).
		meditFinishArm(d, victim)
		return
	}
	victim.SpecFun = meditSpecNames[n]
	if victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
		victim.IndexData.SpecFun = victim.SpecFun
	}
	olcLog(d, "MOB", "Changed spec_fun to %s", victim.SpecFun)
	meditFinishArm(d, victim)
}

// meditArmClan handles MEDIT_CLAN (PC-only). C :2192-2233. LEVEL_GOD-gated
// at the PC main-menu digit-Z already (see Wave 2). Parse accepts
// numeric 0 (clear) or a name (no lookup table yet — accepts the name
// verbatim as pcdata.ClanName). Follow-up: wire ClanLookup when
// util.ClanLookup lands.
func meditArmClan(d *types.DescriptorData, victim *types.CharData, arg string) {
	if victim.IsNPC() {
		meditRejectNpcOnly(d, victim)
		return
	}
	if victim.PCData == nil {
		d.WriteToBuffer("Victim has no PCData.\n\r")
		meditFinishArm(d, victim)
		return
	}
	arg = util.SmashTilde(strings.TrimSpace(arg))
	if arg == "" || arg == "0" {
		victim.PCData.ClanName = ""
		victim.PCData.Clan = nil
		olcLog(d, "MOB", "Cleared clan")
		meditFinishArm(d, victim)
		return
	}
	victim.PCData.ClanName = arg
	// No ClanLookup yet — leave pointer stale; TODO follow-up.
	olcLog(d, "MOB", "Changed clan to %s", arg)
	meditFinishArm(d, victim)
}

// meditArmDeity handles MEDIT_DEITY (PC-only). C :2165-2190.
func meditArmDeity(d *types.DescriptorData, victim *types.CharData, arg string) {
	if victim.IsNPC() {
		meditRejectNpcOnly(d, victim)
		return
	}
	if victim.PCData == nil {
		d.WriteToBuffer("Victim has no PCData.\n\r")
		meditFinishArm(d, victim)
		return
	}
	arg = util.SmashTilde(strings.TrimSpace(arg))
	if arg == "" || arg == "0" {
		victim.PCData.DeityName = ""
		victim.PCData.Deity = nil
		olcLog(d, "MOB", "Cleared deity")
		meditFinishArm(d, victim)
		return
	}
	victim.PCData.DeityName = arg
	olcLog(d, "MOB", "Changed deity to %s", arg)
	meditFinishArm(d, victim)
}

// meditArmCouncil handles MEDIT_COUNCIL (PC-only). C :2235-2262.
func meditArmCouncil(d *types.DescriptorData, victim *types.CharData, arg string) {
	if victim.IsNPC() {
		meditRejectNpcOnly(d, victim)
		return
	}
	if victim.PCData == nil {
		d.WriteToBuffer("Victim has no PCData.\n\r")
		meditFinishArm(d, victim)
		return
	}
	arg = util.SmashTilde(strings.TrimSpace(arg))
	if arg == "" || arg == "0" {
		victim.PCData.CouncilName = ""
		victim.PCData.Council = nil
		olcLog(d, "MOB", "Cleared council")
		meditFinishArm(d, victim)
		return
	}
	victim.PCData.CouncilName = arg
	olcLog(d, "MOB", "Changed council to %s", arg)
	meditFinishArm(d, victim)
}

// --- G8: stat editor (shared helper) ---

// statKey identifies which prototype attribute to mirror to. Used by
// statMirror to pick the matching MobIndexData.Perm* field. The dispatch
// stays pointer-driven (no reflection / no field-name string).
type statKey int

const (
	statStr statKey = iota
	statInt
	statWis
	statDex
	statCon
	statCha
	statLck
)

// statMirror returns a pointer to the MobIndexData.Perm* field matching
// the given stat key, or nil when victim.IndexData is nil. Caller passes
// the result to meditArmStat as the mirror argument; meditArmStat further
// gates the write on IsNPC()+ACT_PROTOTYPE so PCs and non-prototype NPCs
// pass through nil-safe.
func statMirror(victim *types.CharData, key statKey) *int {
	if victim == nil || victim.IndexData == nil {
		return nil
	}
	switch key {
	case statStr:
		return &victim.IndexData.PermStr
	case statInt:
		return &victim.IndexData.PermInt
	case statWis:
		return &victim.IndexData.PermWis
	case statDex:
		return &victim.IndexData.PermDex
	case statCon:
		return &victim.IndexData.PermCon
	case statCha:
		return &victim.IndexData.PermCha
	case statLck:
		return &victim.IndexData.PermLck
	}
	return nil
}

// meditArmStat is the shared helper for MEDIT_STRENGTH..LUCK. C
// omedit.c:998-1007 sets minattr=1,maxattr=25 for NPC and
// minattr=3,maxattr=18 for PC. Per-arm bodies at :1665-1712 write to
// victim.PermStr..PermLck (the Go "perm" slot) and also mirror to
// pIndexData->perm_str (etc.) when NPC+ACT_PROTOTYPE so subsequent mob
// instantiations from the same prototype pick up the new baseline.
// The dispatcher passes the matching MobIndexData field via mirror;
// callers pass nil for non-NPC paths or when no IndexData exists.
// Plan §G8 + Wave 3 follow-up HIGH #1 (2026-04-22).
func meditArmStat(d *types.DescriptorData, victim *types.CharData, arg string, field *int, mirror *int, label string) {
	minAttr, maxAttr := 1, 25
	if !victim.IsNPC() {
		minAttr, maxAttr = 3, 18
	}
	*field = uRange(minAttr, parseInt(arg), maxAttr)
	// NPC + ACT_PROTOTYPE: mirror the write to pIndexData so subsequent
	// mob instantiations pick up the new base attribute (C omedit.c:1665-1712).
	if mirror != nil && victim.IsNPC() && victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
		*mirror = *field
	}
	olcLog(d, "MOB", "Changed %s to %d", label, *field)
	meditFinishArm(d, victim)
}

// --- G9: bitmask arms (wires to olcBitmaskEdit helper) ---

// meditArmNpcFlags handles MEDIT_NPC_FLAGS (NPC-only). C :1437-1482.
// ACT_* bitmask on victim.Act (BitVector). Protected bits:
//
//   - ACT_IS_NPC is uneditable. C emits the verbatim string
//     "It isn't possible to change that flag.\n\r" (omedit.c:1472-1473).
//   - ACT_PROTOTYPE requires LEVEL_GREATER (or pcdata.bestowments
//     contains "protoflag" — not yet ported, see TODO). C emits
//     "You don't have permission to change the prototype flag.\n\r"
//     (omedit.c:1467-1471).
//
// Both gates apply post-resolve so they catch numeric AND keyword input
// (Wave 3 follow-up HIGH #2 + MEDIUM #3, 2026-04-22). Implementation:
// snapshot the bit before dispatch, run the helper, then restore the bit
// + emit the rejection if the bit changed.
func meditArmNpcFlags(d *types.DescriptorData, victim *types.CharData, arg string) {
	if !victim.IsNPC() {
		meditRejectPcOnly(d, victim)
		return
	}
	wasIsNpc := victim.Act.IsSet(types.ACT_IS_NPC)
	wasPrototype := victim.Act.IsSet(types.ACT_PROTOTYPE)
	trustLow := d.Character != nil && d.Character.GetTrust() < types.LEVEL_GREATER

	done := olcBitmaskEditBitVector(d, "ACT_FLAGS", &victim.Act, arg)

	// Post-toggle protection — catches both numeric and keyword forms
	// (mirrors C omedit.c:1467-1473 which gates AFTER lookup resolves).
	if wasIsNpc && !victim.Act.IsSet(types.ACT_IS_NPC) {
		victim.Act.Set(types.ACT_IS_NPC)
		d.WriteToBuffer("It isn't possible to change that flag.\n\r")
	}
	if trustLow && wasPrototype != victim.Act.IsSet(types.ACT_PROTOTYPE) {
		// Toggle would have flipped the bit either direction — restore
		// to the original value and emit the trust-rejection.
		if wasPrototype {
			victim.Act.Set(types.ACT_PROTOTYPE)
		} else {
			victim.Act.Remove(types.ACT_PROTOTYPE)
		}
		d.WriteToBuffer("You don't have permission to change the prototype flag.\n\r")
	}

	if done {
		// LOW #4: do NOT mirror on done — no toggle occurred (C wraps
		// the mirror inside the toggle block at omedit.c:1478-1479).
		meditFinishArm(d, victim)
		return
	}
	if victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
		victim.IndexData.Act = victim.Act
	}
	meditDispNpcFlagsMenu(d)
}

// meditArmPcFlags handles MEDIT_PC_FLAGS (PC-only). C :1484-1513.
// PLR_* bitmask on victim.Act (PCs and NPCs share the Act BitVector
// but the naming differs; C writes through to victim->act).
func meditArmPcFlags(d *types.DescriptorData, victim *types.CharData, arg string) {
	if victim.IsNPC() {
		meditRejectNpcOnly(d, victim)
		return
	}
	done := olcBitmaskEditBitVector(d, "PLR_FLAGS", &victim.Act, arg)
	if done {
		meditFinishArm(d, victim)
		return
	}
	meditDispPcFlagsMenu(d)
}

// meditArmAffFlags handles MEDIT_AFF_FLAGS. C :1546-1577.
// AFF_* bitmask on victim.AffectedBy (BitVector). Shared between NPC/PC.
func meditArmAffFlags(d *types.DescriptorData, victim *types.CharData, arg string) {
	done := olcBitmaskEditBitVector(d, "AFF_FLAGS", &victim.AffectedBy, arg)
	if done {
		// LOW #4: skip mirror on done — no toggle occurred.
		meditFinishArm(d, victim)
		return
	}
	if victim.IsNPC() && victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
		victim.IndexData.AffectedBy = victim.AffectedBy
	}
	meditDispAffFlagsMenu(d)
}

// meditArmPcdataFlags handles MEDIT_PCDATA_FLAGS (PC-only). C :1515-1544.
// PCFLAG_* bitmask on victim.PCData.Flags (int).
func meditArmPcdataFlags(d *types.DescriptorData, victim *types.CharData, arg string) {
	if victim.IsNPC() {
		meditRejectNpcOnly(d, victim)
		return
	}
	if victim.PCData == nil {
		d.WriteToBuffer("Victim has no PCData.\n\r")
		meditFinishArm(d, victim)
		return
	}
	done := olcBitmaskEdit(d, "PCFLAG", &victim.PCData.Flags, arg)
	if done {
		meditFinishArm(d, victim)
		return
	}
	meditDispPcdataFlagsMenu(d)
}

// meditArmParts handles MEDIT_PARTS (NPC-only). C :1892-1914.
// PART_* bitmask on victim.XFlags (int).
func meditArmParts(d *types.DescriptorData, victim *types.CharData, arg string) {
	if !victim.IsNPC() {
		meditRejectPcOnly(d, victim)
		return
	}
	done := olcBitmaskEdit(d, "PART", &victim.XFlags, arg)
	if done {
		// LOW #4: skip mirror on done — no toggle occurred.
		meditFinishArm(d, victim)
		return
	}
	if victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
		victim.IndexData.XFlags = victim.XFlags
	}
	meditDispPartsMenu(d)
}

// meditArmRis is the shared RESISTANT/IMMUNE/SUSCEPTIBLE arm.
// C :1998-2102. RIS_* bitmask on victim.Resistant/Immune/Susceptible
// (three parallel int fields, one shared name table).
func meditArmRis(d *types.DescriptorData, victim *types.CharData, arg string, field *int, label string) {
	_ = label
	done := olcBitmaskEdit(d, "RIS", field, arg)
	if done {
		// LOW #4: skip mirror on done — no toggle occurred.
		meditFinishArm(d, victim)
		return
	}
	if victim.IsNPC() && victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
		// Mirror the targeted field to the prototype. MobIndexData has
		// Resistant / Immune / Susceptible parallel int fields.
		switch field {
		case &victim.Resistant:
			victim.IndexData.Resistant = victim.Resistant
		case &victim.Immune:
			victim.IndexData.Immune = victim.Immune
		case &victim.Susceptible:
			victim.IndexData.Susceptible = victim.Susceptible
		}
	}
	meditDispRisMenu(d)
}
