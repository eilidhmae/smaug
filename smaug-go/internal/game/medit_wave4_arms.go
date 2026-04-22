// Package game — medit Wave 4 G11/G12/G13 arm bodies.
//
// G11 — save-throw editor (MEDIT_SAVE_MENU + 5 SAV arms) + the
//
//	MEDIT_CONFIRM_SAVESTRING Y-path that Wave 2 stubbed.
//
// G12 — class / race editors (MEDIT_CLASS, MEDIT_RACE) with numeric
//
//	and (when present) name lookup.
//
// G13 — password editor (MEDIT_PASSWORD) — security-critical; uses
//
//	bcrypt.GenerateFromPassword (NOT sha256, which C uses) for
//	Go-port parity with DoPassword and bcrypt.CompareHashAndPassword
//	round-trip in the nanny login path.
//
// C reference:
//
//	MEDIT_SAV1..SAV5     — src/omedit.c:1625-1660
//	MEDIT_SAVE_MENU      — src/omedit.c:1833-1864
//	MEDIT_CLASS          — src/omedit.c:1866-1877
//	MEDIT_RACE           — src/omedit.c:1879-1890
//	MEDIT_PASSWORD       — src/omedit.c:1601-1623
//	CONFIRM_SAVESTRING Y — src/omedit.c:1011-1057
//
// Plan: plan-phase6-olc-medit.md §G11 (A23), §G12 (A24, A25),
// §G13 (A26, A27, A28, A29), MEDIT_CONFIRM_SAVESTRING Y-path.
package game

import (
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/act"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"golang.org/x/crypto/bcrypt"
)

// --- G11: Save-throw editor ---

// meditArmSaveMenu handles MEDIT_SAVE_MENU input. Mirrors C
// omedit.c:1833-1864 — a 5-way digit dispatcher routing to MEDIT_SAV1..
// SAV5 plus Q to return to the main menu.
//
// PC-only is NOT enforced here: C medit_disp_pc_menu and
// medit_disp_npc_menu both expose digit P → MEDIT_SAVE_MENU (NPCs have
// saves too in C). The per-SAV arms write to victim.SavingPoisonDeath
// etc. on CharData (verified at internal/types/character.go:141-145) and
// mirror to MobIndexData when NPC + ACT_PROTOTYPE.
func meditArmSaveMenu(d *types.DescriptorData, victim *types.CharData, arg string) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		meditDispSaveMenu(d)
		return
	}
	switch strings.ToUpper(arg[:1]) {
	case "Q":
		d.Olc.Mode = npcOrPcMenu(victim)
		MeditDispMenu(d)
		return
	case "1":
		d.Olc.Mode = types.MEDIT_SAV1
		d.WriteToBuffer("\n\rEnter throw (-30 to 30): ")
	case "2":
		d.Olc.Mode = types.MEDIT_SAV2
		d.WriteToBuffer("\n\rEnter throw (-30 to 30): ")
	case "3":
		d.Olc.Mode = types.MEDIT_SAV3
		d.WriteToBuffer("\n\rEnter throw (-30 to 30): ")
	case "4":
		d.Olc.Mode = types.MEDIT_SAV4
		d.WriteToBuffer("\n\rEnter throw (-30 to 30): ")
	case "5":
		d.Olc.Mode = types.MEDIT_SAV5
		d.WriteToBuffer("\n\rEnter throw (-30 to 30): ")
	default:
		meditDispSaveMenu(d)
	}
}

// savIndex selects which named CharData/MobIndexData saving-throw field
// to mutate. The dispatcher passes 1..5 matching MEDIT_SAV1..SAV5; the
// helper assigns the clamped value to both the instance field and the
// prototype mirror (when NPC + ACT_PROTOTYPE), then redisplays the
// SAVE_MENU.
//
// C clamp: URANGE(-30, atoi(arg), 30). Verified at omedit.c:1626-1660.
func meditArmSav(d *types.DescriptorData, victim *types.CharData, arg string, idx int) {
	v := uRange(-30, parseInt(arg), 30)
	var label string
	switch idx {
	case 1:
		victim.SavingPoisonDeath = v
		label = "save_poison_death"
		if victim.IsNPC() && victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
			victim.IndexData.SavingPoisonDeath = v
		}
	case 2:
		victim.SavingWand = v
		label = "save_wand"
		if victim.IsNPC() && victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
			victim.IndexData.SavingWand = v
		}
	case 3:
		victim.SavingParaPetri = v
		label = "save_paralysis_petrification"
		if victim.IsNPC() && victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
			victim.IndexData.SavingParaPetri = v
		}
	case 4:
		victim.SavingBreath = v
		label = "save_breath"
		if victim.IsNPC() && victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
			victim.IndexData.SavingBreath = v
		}
	case 5:
		victim.SavingSpellStaff = v
		label = "save_spell_staff"
		if victim.IsNPC() && victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
			victim.IndexData.SavingSpellStaff = v
		}
	}
	olcLog(d, "MOB", "Changed %s to %d", label, v)
	if d.Olc != nil {
		d.Olc.Change = true
	}
	d.Olc.Mode = types.MEDIT_SAVE_MENU
	meditDispSaveMenu(d)
}

// --- G11: MEDIT_CONFIRM_SAVESTRING Y-path ---

// meditConfirmSavestringFull replaces the Wave-2 stub. Mirrors C
// omedit.c:1011-1057:
//
//	Y → "Saving...\n\r"; if NPC: best-effort fold_area equivalent
//	    (Go: invoke act.SaveFunc on victim — fold_area equivalent for
//	    mobs lives in oedit's DoSaveArea path which we don't yet wire
//	    from medit; pin a TODO follow-up). If PC: invoke act.SaveFunc
//	    (the project-wide pfile-write seam, equivalent to save_char_obj).
//	    Then cleanup_olc.
//	N → cleanup_olc only (discard).
//	default → invalid + reprompt verbatim per C :1053-1054:
//	    "Invalid choice!\n\r" + "Do you wish to save to disk? : "
//
// Plan §1011-1057 ports verbatim. Wave 2's "Please answer Y or N: "
// reprompt is replaced with the C verbatim strings as part of this fix
// (Wave-3 follow-up "medit G11 polish").
func meditConfirmSavestringFull(d *types.DescriptorData, victim *types.CharData, arg string) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		d.WriteToBuffer("Invalid choice!\n\r")
		d.WriteToBuffer("Do you wish to save to disk? : ")
		return
	}
	switch strings.ToUpper(arg[:1]) {
	case "Y":
		d.WriteToBuffer("Saving...\n\r")
		// PC: invoke pfile saver. NPC: best-effort — same seam (a future
		// fold_area mob-prototype writer would be invoked here). The C
		// branch handles NPC vs PC differently; the Go port currently
		// reuses the single SaveFunc seam for both and logs the operation
		// at OLC log level. Whether NPC fold_area lands in this seam is a
		// deferred concern (see TODO follow-up).
		if victim != nil && act.SaveFunc != nil {
			act.SaveFunc(victim)
		}
		olcLog(d, "MOB", "Confirmed save for %s", victimNameForLog(victim))
		cleanupOlc(d)
		return
	case "N":
		cleanupOlc(d)
		return
	default:
		d.WriteToBuffer("Invalid choice!\n\r")
		d.WriteToBuffer("Do you wish to save to disk? : ")
		return
	}
}

func victimNameForLog(victim *types.CharData) string {
	if victim == nil {
		return "(nil)"
	}
	return victim.Name
}

// --- G12: Class / Race editor ---

// meditArmClass handles MEDIT_CLASS input. Mirrors C omedit.c:1866-1877.
//
// Both NPC and PC are accepted (C does not reject NPC here — it has its
// own clamp range MAX_NPC_CLASS-1, vs PC MAX_CLASS). The Go port matches
// C exactly. Trust gate (LEVEL_GREATER) is applied at the PC main-menu
// digit-4 dispatch (Wave 2) AND defensively re-applied here so the arm
// is safe to enter from a future flat path or a direct mode set.
//
// Name-lookup via worldRef.Classes[i].WhoName when arg is non-numeric.
// Numeric input takes precedence (C parity — C only does numeric).
func meditArmClass(d *types.DescriptorData, victim *types.CharData, arg string) {
	arg = strings.TrimSpace(arg)
	// Defense-in-depth trust gate (Wave 2 dispatch already gates PC path).
	if !victim.IsNPC() && d.Character != nil && d.Character.GetTrust() < types.LEVEL_GREATER {
		d.WriteToBuffer("Requires Greater Immortal trust.\n\r")
		d.Olc.Mode = npcOrPcMenu(victim)
		MeditDispMenu(d)
		return
	}
	idx, ok := classLookupNumericOrName(arg)
	if !ok {
		d.WriteToBuffer("Unknown class. Try again: ")
		return
	}
	if victim.IsNPC() {
		idx = uRange(0, idx, types.MAX_NPC_CLASS-1)
		victim.Class = idx
		if victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
			victim.IndexData.Class = idx
		}
	} else {
		// PC clamp: C uses URANGE(0, number, MAX_CLASS). MAX_CLASS in Go
		// matches C MAX_CLASS (27) verified at types/constants.go:29.
		idx = uRange(0, idx, types.MAX_CLASS)
		victim.Class = idx
	}
	olcLog(d, "MOB", "Changed class to %s", classNameForLog(idx))
	if d.Olc != nil {
		d.Olc.Change = true
	}
	d.Olc.Mode = npcOrPcMenu(victim)
	MeditDispMenu(d)
}

// meditArmRace handles MEDIT_RACE input. Parallel shape to meditArmClass.
// C omedit.c:1879-1890. PC clamp 0..MAX_RACE-1; NPC clamp 0..MAX_NPC_RACE-1.
func meditArmRace(d *types.DescriptorData, victim *types.CharData, arg string) {
	arg = strings.TrimSpace(arg)
	if !victim.IsNPC() && d.Character != nil && d.Character.GetTrust() < types.LEVEL_GREATER {
		d.WriteToBuffer("Requires Greater Immortal trust.\n\r")
		d.Olc.Mode = npcOrPcMenu(victim)
		MeditDispMenu(d)
		return
	}
	idx, ok := raceLookupNumericOrName(arg)
	if !ok {
		d.WriteToBuffer("Unknown race. Try again: ")
		return
	}
	if victim.IsNPC() {
		idx = uRange(0, idx, types.MAX_NPC_RACE-1)
		victim.Race = idx
		if victim.Act.IsSet(types.ACT_PROTOTYPE) && victim.IndexData != nil {
			victim.IndexData.Race = idx
		}
	} else {
		idx = uRange(0, idx, types.MAX_RACE-1)
		victim.Race = idx
	}
	olcLog(d, "MOB", "Changed race to %s", raceNameForLog(idx))
	if d.Olc != nil {
		d.Olc.Change = true
	}
	d.Olc.Mode = npcOrPcMenu(victim)
	MeditDispMenu(d)
}

// classLookupNumericOrName parses arg as a numeric class index OR looks
// up by case-insensitive WhoName in worldRef.Classes. Returns the index
// + true on hit, 0 + false on unknown name. Empty arg → (0, false).
func classLookupNumericOrName(arg string) (int, bool) {
	if arg == "" {
		return 0, false
	}
	if n, err := atoiStrict(arg); err == nil {
		return n, true
	}
	if worldRef == nil {
		return 0, false
	}
	for i, cls := range worldRef.Classes {
		if cls != nil && strings.EqualFold(cls.WhoName, arg) {
			return i, true
		}
	}
	return 0, false
}

// raceLookupNumericOrName: same shape as classLookupNumericOrName, but
// for worldRef.Races[i].Name.
func raceLookupNumericOrName(arg string) (int, bool) {
	if arg == "" {
		return 0, false
	}
	if n, err := atoiStrict(arg); err == nil {
		return n, true
	}
	if worldRef == nil {
		return 0, false
	}
	for i, r := range worldRef.Races {
		if r != nil && strings.EqualFold(r.Name, arg) {
			return i, true
		}
	}
	return 0, false
}

// atoiStrict is strconv.Atoi with leading/trailing whitespace trimmed;
// gives a typed error for the lookup helpers to discriminate "not a
// number" from "out-of-range number" (parseInt swallows the error).
func atoiStrict(s string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(s))
}

// classNameForLog returns the class WhoName for olcLog output, or
// fmt-numeric if worldRef is nil / out-of-range. Never empty.
func classNameForLog(idx int) string {
	if worldRef != nil && idx >= 0 && idx < len(worldRef.Classes) {
		if c := worldRef.Classes[idx]; c != nil && c.WhoName != "" {
			return c.WhoName
		}
	}
	return numberLabel(idx)
}

// raceNameForLog: as above for races.
func raceNameForLog(idx int) string {
	if worldRef != nil && idx >= 0 && idx < len(worldRef.Races) {
		if r := worldRef.Races[idx]; r != nil && r.Name != "" {
			return r.Name
		}
	}
	return numberLabel(idx)
}

// --- G13: MEDIT_PASSWORD (security-critical) ---

// meditArmPassword handles MEDIT_PASSWORD input. Security-critical port
// of C omedit.c:1601-1623 with bcrypt instead of sha256 (Go-port parity
// with DoPassword + nanny login).
//
// Invariants pin-tested in medit_wave4_test.go:
//
//  1. plaintext NEVER stored — victim.PCData.Pwd is the bcrypt hash.
//  2. assignment happens — len(Pwd) > 0 after success.
//  3. round-trip via bcrypt.CompareHashAndPassword succeeds (same path
//     as DoPassword and the nanny).
//  4. olcLog entry contains "Modified password" but NOT the plaintext or
//     the hash value (defense in depth; the helper logs the victim name
//     only).
//  5. tilde-containing input is rejected BEFORE hashing — stored Pwd
//     unchanged. C checks tildes AFTER hashing, which is unreachable on
//     bcrypt's base64 alphabet; the Go port moves the check to RAW input
//     for defense in depth.
//  6. too-short input (< 5 chars) is rejected BEFORE hashing — stored
//     Pwd unchanged.
//
// Trust gate: matches C `LEVEL_SUB_IMPLEM` (omedit.c:1602). Self-edit
// always allowed (C does not differentiate self vs other-PC at this
// point, so neither do we; the LEVEL_SUB_IMPLEM gate is the only auth
// check). NPC rejection mirrors the standard PC-only pattern.
func meditArmPassword(d *types.DescriptorData, victim *types.CharData, arg string) {
	if victim.IsNPC() {
		meditRejectNpcOnly(d, victim)
		return
	}
	if victim.PCData == nil {
		d.WriteToBuffer("Victim has no PCData.\n\r")
		d.Olc.Mode = types.MEDIT_PC_MAIN_MENU
		MeditDispMenu(d)
		return
	}
	// Trust gate — C uses LEVEL_SUB_IMPLEM at omedit.c:1602. Self-edit
	// is implicitly allowed by C (no explicit override), so the gate is
	// the same regardless of whether the builder edits self or another
	// PC. Below-threshold trust silently returns to the main menu (matches
	// C `break;` on the gate fail).
	if d.Character != nil && d.Character.GetTrust() < types.LEVEL_SUB_IMPLEM {
		d.Olc.Mode = types.MEDIT_PC_MAIN_MENU
		MeditDispMenu(d)
		return
	}

	// Defense-in-depth tilde check on RAW input BEFORE hashing.
	// C checks the post-hash output for tildes (omedit.c:1610-1616) which
	// is unreachable on bcrypt; we move the check to the input side. An
	// empty arg falls through to the min-length check below.
	if strings.Contains(arg, "~") {
		d.WriteToBuffer("Unacceptable choice, try again: ")
		return // stay in MEDIT_PASSWORD
	}

	// SmashTilde is a no-op here (we already rejected on tilde) but
	// harmless — kept for parity with other string-input arms.
	clean := util.SmashTilde(arg)

	// Min-length: C uses 5 (omedit.c:1604). DoPassword uses 6 in Go. Per
	// plan §398-399 G13 matches C exactly to preserve the medit-vs-flat
	// behavior contract.
	if len(clean) < 5 {
		d.WriteToBuffer("Password too short, try again: ")
		return // stay in MEDIT_PASSWORD
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(clean), act.BcryptCost)
	if err != nil {
		util.Bug("meditArmPassword: GenerateFromPassword: %v", err)
		d.WriteToBuffer("Password hashing failed.\n\r")
		d.Olc.Mode = types.MEDIT_PC_MAIN_MENU
		MeditDispMenu(d)
		return
	}
	victim.PCData.Pwd = string(hash)

	// Persist immediately to mirror DoPassword's unconditional save (Go
	// does not gate on SV_PASSCHG sysflag; plan §420-424).
	if act.SaveFunc != nil {
		act.SaveFunc(victim)
	}

	// Log change WITHOUT exposing the plaintext or hash value.
	olcLog(d, "MOB", "Modified password for %s", victim.Name)
	d.WriteToBuffer("Password updated.\n\r")
	if d.Olc != nil {
		d.Olc.Change = true
	}
	d.Olc.Mode = types.MEDIT_PC_MAIN_MENU
	MeditDispMenu(d)
}

// numberLabel returns "%d" formatting for an int, used as a fallback in
// log labels when the world tables are absent.
func numberLabel(n int) string {
	return strconv.Itoa(n)
}
