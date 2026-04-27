// Package game — interactive mob/character editor (CON_MEDIT substate).
//
// Wave 2 ports the top-level CON_MEDIT dispatcher + NPC/PC main-menu arms +
// MEDIT_CONFIRM_SAVESTRING stub. Simple-field arms, stat editors, flag
// editors, affect editor, save editor, class/race editors, and password
// editor land in later waves per plan-phase6-olc-medit.md §G7-§G13.
//
// C reference: src/omedit.c medit_parse at :985-2160 (NPC main menu
// :1059-1215, PC main menu :1218-1386). The pulse loop dispatches input
// addressed to a descriptor in CON_MEDIT directly here (see loop.go
// processInput); the nanny is NOT invoked for menu input.
//
// Plan: plan-phase6-olc-medit.md §G4 / §G5 / §G6.
package game

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
)

// worldMobLookup is the seam the parser uses to resolve mob prototype
// vnums. Points at worldRef.GetMobIndex at boot time (worldRef /
// SetWorldRef are shared with redit_parse.go / oedit_parse.go); tests
// can override directly without standing up a full world.
var worldMobLookup = func(vnum int) *types.MobIndexData {
	if worldRef == nil {
		return nil
	}
	return worldRef.GetMobIndex(vnum)
}

// worldPcLookup resolves a connected-PC name to its *CharData by walking
// world.Descriptors. Mirrors C get_char_world() (handler.c:1894-) for the
// PC subset only — NPCs use worldMobLookup. Returns the first descriptor
// whose Connected == CON_PLAYING and whose Character.Name matches case-
// insensitively. Linkdead (Connected != CON_PLAYING) and nil-Character
// descriptors are skipped.
//
// Tests can override directly without standing up a full world.
//
// Plan: plan-phase6-quickwins-blank-pcrename.md §D3 / §G3.
var worldPcLookup = func(name string) *types.CharData {
	if worldRef == nil {
		return nil
	}
	for _, d := range worldRef.Descriptors {
		if d == nil || d.Character == nil {
			continue
		}
		if d.Connected != int(types.CON_PLAYING) {
			continue
		}
		if strings.EqualFold(d.Character.Name, name) {
			return d.Character
		}
	}
	return nil
}

// meditParse is the top-level CON_MEDIT dispatcher. Keyed on d.Olc.Mode,
// it mutates the victim (an NPC prototype OR a connected PC) and either
// returns (stay in same mode) or transitions to a new mode + redisplays
// the appropriate NPC / PC main menu.
//
// The caller (processInput) has already validated d.Connected ==
// CON_MEDIT. If d.Olc is nil, d.Character is nil, or d.Olc.Target is not
// a *CharData (stale state bug elsewhere), cleanupOlc fires defensively
// and the descriptor drops back to CON_PLAYING.
//
// Wave 2 fills in main-menu dispatch + confirm-savestring stub. Every
// other MEDIT_* mode falls through to the default arm which redisplays
// the main menu appropriate to the victim — this matches C's behavior
// for uninitialized modes and lets Waves 3+ fill in arm bodies without
// any dispatcher surgery.
func meditParse(d *types.DescriptorData, arg string) {
	if d == nil {
		return
	}
	if d.Olc == nil || d.Character == nil {
		// Defensive: impossible state; restore playable.
		d.Connected = int(types.CON_PLAYING)
		return
	}
	victim, ok := d.Olc.Target.(*types.CharData)
	if !ok || victim == nil {
		// Target lost / wrong type (should not happen). Restore playable.
		cleanupOlc(d)
		return
	}

	switch d.Olc.Mode {
	case types.MEDIT_NPC_MAIN_MENU:
		meditDispatchNpcMain(d, victim, arg)
	case types.MEDIT_PC_MAIN_MENU:
		meditDispatchPcMain(d, victim, arg)
	case types.MEDIT_CONFIRM_SAVESTRING:
		meditDispatchConfirmSavestring(d, victim, arg)

	// --- G7 simple-field arms (plan-phase6-olc-medit.md §G7) ---
	case types.MEDIT_NAME:
		meditArmName(d, victim, arg)
	case types.MEDIT_S_DESC:
		meditArmShortDesc(d, victim, arg)
	case types.MEDIT_L_DESC:
		meditArmLongDesc(d, victim, arg)
	case types.MEDIT_D_DESC:
		meditArmDDescBugTrap(d, victim, arg)
	case types.MEDIT_SEX:
		meditArmSex(d, victim, arg)
	case types.MEDIT_HITROLL:
		meditArmHitroll(d, victim, arg)
	case types.MEDIT_DAMROLL:
		meditArmDamroll(d, victim, arg)
	case types.MEDIT_DAMNUMDIE:
		meditArmDamNumDie(d, victim, arg)
	case types.MEDIT_DAMSIZEDIE:
		meditArmDamSizeDie(d, victim, arg)
	case types.MEDIT_DAMPLUS:
		meditArmDamPlus(d, victim, arg)
	case types.MEDIT_HITNUMDIE:
		meditArmHitNumDie(d, victim, arg)
	case types.MEDIT_HITSIZEDIE:
		meditArmHitSizeDie(d, victim, arg)
	case types.MEDIT_HITPLUS:
		meditArmHitPlus(d, victim, arg)
	case types.MEDIT_AC:
		meditArmArmor(d, victim, arg)
	case types.MEDIT_GOLD:
		meditArmGold(d, victim, arg)
	case types.MEDIT_SILVER:
		meditArmSilver(d, victim, arg)
	case types.MEDIT_COPPER:
		meditArmCopper(d, victim, arg)
	case types.MEDIT_POS:
		meditArmPosition(d, victim, arg)
	case types.MEDIT_DEFAULT_POS:
		meditArmDefaultPosition(d, victim, arg)
	case types.MEDIT_LEVEL:
		meditArmLevel(d, victim, arg)
	case types.MEDIT_ALIGNMENT:
		meditArmAlignment(d, victim, arg)
	case types.MEDIT_FAVOR:
		meditArmFavor(d, victim, arg)
	case types.MEDIT_HITPOINT:
		meditArmHitpoint(d, victim, arg)
	case types.MEDIT_MANA:
		meditArmMana(d, victim, arg)
	case types.MEDIT_MOVE:
		meditArmMove(d, victim, arg)
	case types.MEDIT_PRACTICE:
		meditArmPractice(d, victim, arg)
	case types.MEDIT_MENTALSTATE:
		meditArmMentalState(d, victim, arg)
	case types.MEDIT_EMOTIONAL:
		meditArmEmotional(d, victim, arg)
	case types.MEDIT_THIRST:
		meditArmThirst(d, victim, arg)
	case types.MEDIT_FULL:
		meditArmFull(d, victim, arg)
	case types.MEDIT_DRUNK:
		meditArmDrunk(d, victim, arg)
	case types.MEDIT_ATTACK:
		meditArmAttack(d, victim, arg)
	case types.MEDIT_DEFENSE:
		meditArmDefense(d, victim, arg)
	case types.MEDIT_SPEC:
		meditArmSpec(d, victim, arg)
	case types.MEDIT_CLAN:
		meditArmClan(d, victim, arg)
	case types.MEDIT_DEITY:
		meditArmDeity(d, victim, arg)
	case types.MEDIT_COUNCIL:
		meditArmCouncil(d, victim, arg)

	// --- G8 stat editors (plan-phase6-olc-medit.md §G8) ---
	// Mirror pointers are nil-safe: helpers below return nil when
	// victim.IndexData is nil (PC path or NPC w/o prototype index).
	// meditArmStat itself also gates the mirror write on
	// IsNPC()+ACT_PROTOTYPE+IndexData!=nil per HIGH #1 (Wave 3 follow-up).
	case types.MEDIT_STRENGTH:
		meditArmStat(d, victim, arg, &victim.PermStr, statMirror(victim, statStr), "strength")
	case types.MEDIT_INTELLIGENCE:
		meditArmStat(d, victim, arg, &victim.PermInt, statMirror(victim, statInt), "intelligence")
	case types.MEDIT_WISDOM:
		meditArmStat(d, victim, arg, &victim.PermWis, statMirror(victim, statWis), "wisdom")
	case types.MEDIT_DEXTERITY:
		meditArmStat(d, victim, arg, &victim.PermDex, statMirror(victim, statDex), "dexterity")
	case types.MEDIT_CONSTITUTION:
		meditArmStat(d, victim, arg, &victim.PermCon, statMirror(victim, statCon), "constitution")
	case types.MEDIT_CHARISMA:
		meditArmStat(d, victim, arg, &victim.PermCha, statMirror(victim, statCha), "charisma")
	case types.MEDIT_LUCK:
		meditArmStat(d, victim, arg, &victim.PermLck, statMirror(victim, statLck), "luck")

	// --- G9 bitmask arms (plan-phase6-olc-medit.md §G9) ---
	case types.MEDIT_NPC_FLAGS:
		meditArmNpcFlags(d, victim, arg)
	case types.MEDIT_PC_FLAGS:
		meditArmPcFlags(d, victim, arg)
	case types.MEDIT_AFF_FLAGS:
		meditArmAffFlags(d, victim, arg)
	case types.MEDIT_PCDATA_FLAGS:
		meditArmPcdataFlags(d, victim, arg)
	case types.MEDIT_PARTS:
		meditArmParts(d, victim, arg)
	case types.MEDIT_RESISTANT:
		meditArmRis(d, victim, arg, &victim.Resistant, "Resistant")
	case types.MEDIT_IMMUNE:
		meditArmRis(d, victim, arg, &victim.Immune, "Immune")
	case types.MEDIT_SUSCEPTIBLE:
		meditArmRis(d, victim, arg, &victim.Susceptible, "Susceptible")

	// --- G11 save editor (plan-phase6-olc-medit.md §G11) ---
	case types.MEDIT_SAVE_MENU:
		meditArmSaveMenu(d, victim, arg)
	case types.MEDIT_SAV1:
		meditArmSav(d, victim, arg, 1)
	case types.MEDIT_SAV2:
		meditArmSav(d, victim, arg, 2)
	case types.MEDIT_SAV3:
		meditArmSav(d, victim, arg, 3)
	case types.MEDIT_SAV4:
		meditArmSav(d, victim, arg, 4)
	case types.MEDIT_SAV5:
		meditArmSav(d, victim, arg, 5)

	// --- G12 class / race editors (plan-phase6-olc-medit.md §G12) ---
	case types.MEDIT_CLASS:
		meditArmClass(d, victim, arg)
	case types.MEDIT_RACE:
		meditArmRace(d, victim, arg)

	// --- G13 password editor (plan-phase6-olc-medit.md §G13) ---
	case types.MEDIT_PASSWORD:
		meditArmPassword(d, victim, arg)

	// --- G10 affect-list editor (plan-phase6-olc-medit.md §G10) ---
	case types.MEDIT_AFFECT_MENU:
		meditArmAffectMenu(d, victim, arg)
	case types.MEDIT_AFFECT_LOCATION:
		meditArmAffectLocation(d, victim, arg)
	case types.MEDIT_AFFECT_MODIFIER:
		meditArmAffectModifier(d, victim, arg)
	case types.MEDIT_AFFECT_REMOVE:
		meditArmAffectRemove(d, victim, arg)

	default:
		// Unimplemented arm. Redisplay the main menu appropriate to
		// the victim so the session stays navigable.
		d.Olc.Mode = npcOrPcMenu(victim)
		MeditDispMenu(d)
	}
}

// npcOrPcMenu returns the main-menu mode constant for the given victim.
// Centralizes the NPC-vs-PC branch so the many redisplay sites stay
// consistent.
func npcOrPcMenu(victim *types.CharData) int {
	if victim != nil && victim.IsNPC() {
		return types.MEDIT_NPC_MAIN_MENU
	}
	return types.MEDIT_PC_MAIN_MENU
}

// firstUpper returns the uppercase first byte of arg as a 1-char string,
// or "" if arg is empty. Mirrors C's UPPER(*arg).
func firstUpper(arg string) string {
	s := strings.TrimSpace(arg)
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1])
}

// meditDispatchNpcMain implements the NPC main-menu digit dispatcher.
// Mirrors C omedit.c:1059-1215. Each recognized digit sets d.Olc.Mode to
// the appropriate MEDIT_* target and writes a submenu-prompt (the real
// PARSE arm lands in a later wave; Wave 2 only routes to the mode and
// shows the submenu header).
//
// Digit table pinned in plan §230-269.
func meditDispatchNpcMain(d *types.DescriptorData, victim *types.CharData, arg string) {
	switch firstUpper(arg) {
	case "Q":
		// NPC Q is an unconditional cleanup — no save-confirm per C :1062-1064.
		d.WriteToBuffer("Exiting editor.\n\r")
		cleanupOlc(d)
		return

	case "1":
		d.Olc.Mode = types.MEDIT_SEX
		meditDispSexMenu(d)
	case "2":
		d.Olc.Mode = types.MEDIT_NAME
		d.WriteToBuffer("\n\rEnter name: ")
	case "3":
		d.Olc.Mode = types.MEDIT_S_DESC
		d.WriteToBuffer("\n\rEnter short description: ")
	case "4":
		d.Olc.Mode = types.MEDIT_L_DESC
		d.WriteToBuffer("\n\rEnter long description: ")
	case "5":
		// D_DESC uses the line editor; full trampoline wiring lands in
		// G7. Wave 2 only sets the mode so the dispatcher is complete.
		d.Olc.Mode = types.MEDIT_D_DESC
		d.WriteToBuffer("Enter new mob description (use the editor):\n\r")
	case "6":
		d.Olc.Mode = types.MEDIT_CLASS
		meditDispClassMenu(d)
	case "7":
		d.Olc.Mode = types.MEDIT_RACE
		meditDispRaceMenu(d)
	case "8":
		d.Olc.Mode = types.MEDIT_LEVEL
		d.WriteToBuffer("\n\rEnter level: ")
	case "9":
		d.Olc.Mode = types.MEDIT_ALIGNMENT
		d.WriteToBuffer("\n\rEnter alignment: ")
	case "A":
		d.Olc.Mode = types.MEDIT_STRENGTH
		d.WriteToBuffer("\n\rEnter strength: ")
	case "B":
		d.Olc.Mode = types.MEDIT_INTELLIGENCE
		d.WriteToBuffer("\n\rEnter intelligence: ")
	case "C":
		d.Olc.Mode = types.MEDIT_WISDOM
		d.WriteToBuffer("\n\rEnter wisdom: ")
	case "D":
		d.Olc.Mode = types.MEDIT_DEXTERITY
		d.WriteToBuffer("\n\rEnter dexterity: ")
	case "E":
		d.Olc.Mode = types.MEDIT_CONSTITUTION
		d.WriteToBuffer("\n\rEnter constitution: ")
	case "F":
		d.Olc.Mode = types.MEDIT_CHARISMA
		d.WriteToBuffer("\n\rEnter charisma: ")
	case "G":
		d.Olc.Mode = types.MEDIT_LUCK
		d.WriteToBuffer("\n\rEnter luck: ")
	case "H":
		d.Olc.Mode = types.MEDIT_DAMNUMDIE
		d.WriteToBuffer("\n\rEnter number of damage dice: ")
	case "I":
		d.Olc.Mode = types.MEDIT_DAMSIZEDIE
		d.WriteToBuffer("\n\rEnter size of damage dice: ")
	case "J":
		d.Olc.Mode = types.MEDIT_DAMPLUS
		d.WriteToBuffer("\n\rEnter amount to add to damage: ")
	case "K":
		d.Olc.Mode = types.MEDIT_HITNUMDIE
		d.WriteToBuffer("\n\rEnter number of hitpoint dice: ")
	case "L":
		d.Olc.Mode = types.MEDIT_HITSIZEDIE
		d.WriteToBuffer("\n\rEnter size of hitpoint dice: ")
	case "M":
		d.Olc.Mode = types.MEDIT_HITPLUS
		d.WriteToBuffer("\n\rEnter amount to add to hitpoints: ")
	case "N":
		// C branches on ENABLE_GOLD_SILVER_COPPER; Go-port port mirrors
		// the default (non-GSC) build — MEDIT_GOLD. If GSC is later
		// activated this arm flips to MEDIT_COPPER per plan §257.
		d.Olc.Mode = types.MEDIT_GOLD
		d.WriteToBuffer("\n\rEnter amount of gold mobile carries: ")
	case "O":
		d.Olc.Mode = types.MEDIT_SPEC
		meditDispSpecMenu(d)
	case "P":
		d.Olc.Mode = types.MEDIT_SAVE_MENU
		meditDispSaveMenu(d)
	case "R":
		d.Olc.Mode = types.MEDIT_RESISTANT
		meditDispRisMenu(d)
	case "S":
		d.Olc.Mode = types.MEDIT_IMMUNE
		meditDispRisMenu(d)
	case "T":
		d.Olc.Mode = types.MEDIT_SUSCEPTIBLE
		meditDispRisMenu(d)
	case "U":
		d.Olc.Mode = types.MEDIT_POS
		meditDispPosMenu(d)
	case "V":
		d.Olc.Mode = types.MEDIT_ATTACK
		meditDispAttackMenu(d)
	case "W":
		d.Olc.Mode = types.MEDIT_DEFENSE
		meditDispDefenseMenu(d)
	case "X":
		d.Olc.Mode = types.MEDIT_PARTS
		meditDispPartsMenu(d)
	case "Y":
		d.Olc.Mode = types.MEDIT_NPC_FLAGS
		meditDispNpcFlagsMenu(d)
	case "Z":
		d.Olc.Mode = types.MEDIT_AFF_FLAGS
		meditDispAffFlagsMenu(d)

	default:
		// C :1212-1214 default: silently redisplay the NPC main menu.
		meditDispNpcMenu(d)
	}
}

// meditDispatchPcMain implements the PC main-menu digit dispatcher.
// Mirrors C omedit.c:1218-1386. PC Q enters MEDIT_CONFIRM_SAVESTRING when
// OLC_CHANGE is set; otherwise unconditional cleanup. Digits 6 and U are
// explicit "NPC Only" / "NPCs Only" rejects per C :1255-1257 / :1351-1353.
// CLASS and RACE are LEVEL_GREATER-gated per plan §597-607 + A13.
//
// Digit table pinned in plan §270-310.
func meditDispatchPcMain(d *types.DescriptorData, victim *types.CharData, arg string) {
	ch := d.Character
	switch firstUpper(arg) {
	case "Q":
		if d.Olc.Change {
			d.WriteToBuffer("Do you wish to save changes to disk? (y/n): ")
			d.Olc.Mode = types.MEDIT_CONFIRM_SAVESTRING
			return
		}
		d.WriteToBuffer("Exiting editor.\n\r")
		cleanupOlc(d)
		return

	case "1":
		d.Olc.Mode = types.MEDIT_SEX
		meditDispSexMenu(d)
	case "2":
		d.Olc.Mode = types.MEDIT_NAME
		// C :1234-1236 triggers do_pcrename only for trust>LEVEL_SUB_IMPLEM-1;
		// Go-port G7 ports that; Wave 2 only routes the digit.
		d.WriteToBuffer("\n\rEnter new name: ")
	case "3":
		d.Olc.Mode = types.MEDIT_D_DESC
		d.WriteToBuffer("Enter new player description (use the editor):\n\r")
	case "4":
		if ch.GetTrust() < types.LEVEL_GREATER {
			d.WriteToBuffer("Requires Greater Immortal trust.\n\r")
			meditDispPcMenu(d)
			return
		}
		d.Olc.Mode = types.MEDIT_CLASS
		meditDispClassMenu(d)
	case "5":
		if ch.GetTrust() < types.LEVEL_GREATER {
			d.WriteToBuffer("Requires Greater Immortal trust.\n\r")
			meditDispPcMenu(d)
			return
		}
		d.Olc.Mode = types.MEDIT_RACE
		meditDispRaceMenu(d)
	case "6":
		// C :1255-1257 verbatim. Mode unchanged — builder stays on PC menu.
		d.WriteToBuffer("\n\rNPC Only!!")
		// C falls through to the main-menu redisplay via the switch's
		// trailing `break` — mirror that by redisplaying PC menu.
		meditDispPcMenu(d)
	case "7":
		d.Olc.Mode = types.MEDIT_ALIGNMENT
		d.WriteToBuffer("\n\rEnter alignment: ")
	case "8":
		d.Olc.Mode = types.MEDIT_STRENGTH
		d.WriteToBuffer("\n\rEnter strength: ")
	case "9":
		d.Olc.Mode = types.MEDIT_INTELLIGENCE
		d.WriteToBuffer("\n\rEnter intelligence: ")
	case "A":
		d.Olc.Mode = types.MEDIT_WISDOM
		d.WriteToBuffer("\n\rEnter wisdom: ")
	case "B":
		d.Olc.Mode = types.MEDIT_DEXTERITY
		d.WriteToBuffer("\n\rEnter dexterity: ")
	case "C":
		d.Olc.Mode = types.MEDIT_CONSTITUTION
		d.WriteToBuffer("\n\rEnter constitution: ")
	case "D":
		d.Olc.Mode = types.MEDIT_CHARISMA
		d.WriteToBuffer("\n\rEnter charisma: ")
	case "E":
		d.Olc.Mode = types.MEDIT_LUCK
		d.WriteToBuffer("\n\rEnter luck: ")
	case "F":
		d.Olc.Mode = types.MEDIT_HITPOINT
		d.WriteToBuffer("\n\rEnter hitpoints: ")
	case "G":
		d.Olc.Mode = types.MEDIT_MANA
		d.WriteToBuffer("\n\rEnter mana: ")
	case "H":
		d.Olc.Mode = types.MEDIT_MOVE
		d.WriteToBuffer("\n\rEnter moves: ")
	case "I":
		// Non-GSC path. (GSC would map to MEDIT_COPPER per plan §292.)
		d.Olc.Mode = types.MEDIT_GOLD
		d.WriteToBuffer("\n\rEnter amount of gold player carries: ")
	case "J":
		d.Olc.Mode = types.MEDIT_MENTALSTATE
		d.WriteToBuffer("\n\rEnter players mentalstate: ")
	case "K":
		d.Olc.Mode = types.MEDIT_EMOTIONAL
		d.WriteToBuffer("\n\rEnter players emotional state: ")
	case "L":
		d.Olc.Mode = types.MEDIT_THIRST
		d.WriteToBuffer("\n\rEnter player's thirst (0 = dehydrated): ")
	case "M":
		d.Olc.Mode = types.MEDIT_FULL
		d.WriteToBuffer("\n\rEnter player's fullness (0 = starving): ")
	case "N":
		d.Olc.Mode = types.MEDIT_DRUNK
		d.WriteToBuffer("\n\rEnter player's drunkeness (0 = sober): ")
	case "O":
		d.Olc.Mode = types.MEDIT_FAVOR
		d.WriteToBuffer("\n\rEnter player's favor (-2500 to 2500): ")
	case "P":
		d.Olc.Mode = types.MEDIT_SAVE_MENU
		meditDispSaveMenu(d)
	case "R":
		d.Olc.Mode = types.MEDIT_RESISTANT
		meditDispRisMenu(d)
	case "S":
		d.Olc.Mode = types.MEDIT_IMMUNE
		meditDispRisMenu(d)
	case "T":
		d.Olc.Mode = types.MEDIT_SUSCEPTIBLE
		meditDispRisMenu(d)
	case "U":
		// C :1351-1353 verbatim. Mode unchanged.
		d.WriteToBuffer("NPCs Only!!\n\r")
		meditDispPcMenu(d)
	case "V":
		d.Olc.Mode = types.MEDIT_PC_FLAGS
		meditDispPcFlagsMenu(d)
	case "W":
		d.Olc.Mode = types.MEDIT_PCDATA_FLAGS
		meditDispPcdataFlagsMenu(d)
	case "X":
		d.Olc.Mode = types.MEDIT_AFF_FLAGS
		meditDispAffFlagsMenu(d)
	case "Y":
		d.Olc.Mode = types.MEDIT_DEITY
		d.WriteToBuffer("\n\rEnter deity name (blank to clear): ")
	case "Z":
		// Clan — LEVEL_GOD-gated per C :1371-1374.
		if ch.GetTrust() < types.LEVEL_GOD {
			meditDispPcMenu(d)
			return
		}
		d.Olc.Mode = types.MEDIT_CLAN
		d.WriteToBuffer("\n\rEnter clan name (blank to clear): ")
	case "=":
		// Council — LEVEL_SUB_IMPLEM-gated per C :1377-1380.
		if ch.GetTrust() < types.LEVEL_SUB_IMPLEM {
			meditDispPcMenu(d)
			return
		}
		d.Olc.Mode = types.MEDIT_COUNCIL
		d.WriteToBuffer("\n\rEnter council name (blank to clear): ")

	default:
		// C :1382-1384 default: redisplay the PC main menu. Go port
		// deliberately preserves C's apparent bug — the C default calls
		// medit_disp_npc_menu even from the PC dispatcher; the Go port
		// instead calls meditDispPcMenu so a PC session keeps rendering
		// the PC menu. Plan §352 documents this as an intentional
		// divergence from C (C bug).
		meditDispPcMenu(d)
	}
}

// meditDispatchConfirmSavestring handles the Y/N save-confirm prompt
// after a PC-menu Q when OLC_CHANGE is set. Wave 4 / G11 promotes this
// from the Wave-2 stub to the full C-parity port:
//
//	Y → "Saving...\n\r" + invoke act.SaveFunc(victim) + cleanupOlc.
//	    For NPC the C path runs fold_area on the prototype's owning area;
//	    the Go port reuses act.SaveFunc as a single seam (a future
//	    fold_area mob-prototype writer can land behind it without a call-
//	    site change).
//	N → cleanupOlc (discard).
//	default → reprompt verbatim per C omedit.c:1053-1054 — "Invalid
//	    choice!\n\r" + "Do you wish to save to disk? : ". Wave 2's Go-
//	    idiom "Please answer Y or N: " reprompt is replaced here.
//
// Plan §G11 + Wave 3 follow-up "medit G11 polish".
func meditDispatchConfirmSavestring(d *types.DescriptorData, victim *types.CharData, arg string) {
	meditConfirmSavestringFull(d, victim, arg)
}
