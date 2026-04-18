package act

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// BcryptCost is the cost parameter for hashing new passwords written from
// the act package (currently only DoPassword). Tests may lower this to
// bcrypt.MinCost for speed; production leaves it at bcrypt.DefaultCost.
// Kept in sync with game.BcryptCost by internal/boot.
// Written once at boot before the game loop starts; read only from the
// game loop goroutine. Safe without synchronization.
var BcryptCost = bcrypt.DefaultCost

// DoSave implements the 'save' command (C act_comm.c:3277 do_save).
// Writes the current player state to disk via SaveFunc.
func DoSave(ch *types.CharData, argument string) {
	if ch.IsNPC() {
		return
	}
	if ch.Level < 2 {
		ch.Send("You must be at least second level to save.\n\r")
		return
	}
	ch.Wait = 2 // WAIT_STATE(ch, 2) — tempers save-happy players.
	if SaveFunc != nil {
		SaveFunc(ch)
	}
	ch.Send("Saved...\n\r")
}

// DoAfk implements the 'afk' command (C act_info.c:6020 do_afk). Toggles
// PLR_AFK on ch; emits a self-message and a room broadcast via Act with
// TO_CANSEE so only observers who can see ch hear the status change.
func DoAfk(ch *types.CharData, argument string) {
	if ch.IsNPC() {
		return
	}
	if ch.Act.IsSet(types.PLR_AFK) {
		ch.Act.Remove(types.PLR_AFK)
		ch.Send("You are no longer afk.\n\r")
		util.Act(types.AT_ACTION, "$n is no longer afk.", ch, nil, nil, nil, types.TO_CANSEE)
		return
	}
	ch.Act.Set(types.PLR_AFK)
	ch.Send("You are now afk.\n\r")
	util.Act(types.AT_ACTION, "$n is now afk.", ch, nil, nil, nil, types.TO_CANSEE)
}

// DoGag implements the 'gag' command — a standalone no-arg toggle for
// PCFLAG_GAG, which suppresses zero-damage combat-miss messages (honored in
// internal/combat/dammessage.go:342-349).
//
// C divergence: the C port embeds gag toggling inside `do_config` at
// src/act_info.c:5585 (gag branch at :5794), dispatched via `config +gag` /
// `config -gag` syntax with a generic `"Ok.\n"` echo. The Go port instead
// ships standalone per-flag toggles (matching DoAfk's local precedent) with
// directional messages. See smaug-go/doc/plan-do-gag.md for rationale.
//
// The nil-PCData guard is load-bearing: IsNPC() only checks Act.ACT_IS_NPC
// and does not look at PCData, so a malformed non-NPC with nil PCData would
// pass IsNPC() and crash dereferencing .Flags without this guard.
func DoGag(ch *types.CharData, argument string) {
	if ch.IsNPC() {
		return
	}
	if ch.PCData == nil {
		return
	}
	mask := int(types.PCFLAG_GAG)
	if ch.PCData.Flags&mask != 0 {
		ch.PCData.Flags &^= mask
		ch.Send("Combat messages will no longer be gagged.\n\r")
		return
	}
	ch.PCData.Flags |= mask
	ch.Send("Combat messages will be gagged.\n\r")
}

// setTitle is the Go port of C player.c:3112 set_title. If title starts with
// an alphanumeric character it is prefixed with a space (so titles like "the
// Wizard" read naturally after the player's name); otherwise the title is
// assigned as-is (letting players start titles with punctuation).
func setTitle(ch *types.CharData, title string) {
	if ch.IsNPC() {
		util.Bug("Set_title: NPC.")
		return
	}
	if title == "" {
		ch.PCData.Title = ""
		return
	}
	first := title[0]
	if (first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') ||
		(first >= '0' && first <= '9') {
		ch.PCData.Title = " " + title
	} else {
		ch.PCData.Title = title
	}
}

// DoTitle implements the 'title' command (C player.c:3138 do_title).
// Sets the caller's title after scrubbing tildes and color-tokens, enforcing
// a 50-char max, and gating on PCFLAG_NOTITLE.
func DoTitle(ch *types.CharData, argument string) {
	if ch.IsNPC() {
		return
	}
	if ch.PCData == nil {
		return
	}
	if (uint32(ch.PCData.Flags) & types.PCFLAG_NOTITLE) != 0 {
		ch.Send("The Gods prohibit you from changing your title.\n\r")
		return
	}
	if argument == "" {
		ch.Send("Change your title to what?\n\r")
		return
	}
	// C truncates with argument[50] = '\0' — cut to 50 bytes (matches C).
	if len(argument) > 50 {
		argument = argument[:50]
	}
	argument = util.SmashTilde(argument)
	argument = util.SmashColorToken(argument)
	setTitle(ch, argument)
	ch.Send("Your new title has been set.\n\r")
}

// DoBio implements the 'bio' command (C player.c:3413 do_bio). Launches
// the string editor seeded with the player's current PCData.Bio; on `/s`
// the EditorSave closure copies the buffer back to PCData.Bio.
//
// Requires editor plumbing from internal/game via the EditorSave /
// CopyBufferFunc / StopEditingFunc seams — see olc.go for the pattern.
// NPCs, nil-PCData, and PCFLAG_NOBIO are hard blocks. Level-5 gate from
// C (player.c:3420) is skipped here to match the local convention of
// deferring content gates to the player-config wiring (DoTitle / DoAfk
// take the same approach — no level gate).
func DoBio(ch *types.CharData, argument string) {
	if ch.IsNPC() {
		ch.Send("Mobs cannot set a bio.\n\r")
		return
	}
	if ch.PCData == nil {
		return
	}
	if (uint32(ch.PCData.Flags) & types.PCFLAG_NOBIO) != 0 {
		ch.Send("The gods won't allow you to do that!\n\r")
		return
	}
	if ch.Desc == nil {
		// Matches C's bug("do_bio: no descriptor"), but silent — no
		// broadcast channel for the log in the Go port.
		return
	}
	ch.Substate = types.SUB_PERSONAL_BIO
	ch.EditorSave = func(c *types.CharData) {
		if c.PCData == nil {
			return
		}
		if CopyBufferFunc != nil {
			c.PCData.Bio = CopyBufferFunc(c)
		}
		if StopEditingFunc != nil {
			StopEditingFunc(c)
		}
		c.Send("\n\r")
	}
	if StartEditingFunc != nil {
		StartEditingFunc(ch, ch.PCData.Bio)
	}
}

// DoDescription implements the 'description' command (C player.c:3366
// do_description). Same shape as DoBio but targets ch.Description
// (the CharData field, not a PCData field — C has description on
// char_data itself, mud.h:2733).
//
// PCFLAG_NODESC is a hard block (C player.c:3374-3378).
func DoDescription(ch *types.CharData, argument string) {
	if ch.IsNPC() {
		ch.Send("Monsters are too dumb to do that!\n\r")
		return
	}
	if ch.PCData == nil {
		return
	}
	if (uint32(ch.PCData.Flags) & types.PCFLAG_NODESC) != 0 {
		ch.Send("You cannot set your description.\n\r")
		return
	}
	if ch.Desc == nil {
		return
	}
	ch.Substate = types.SUB_PERSONAL_DESC
	ch.EditorSave = func(c *types.CharData) {
		if CopyBufferFunc != nil {
			c.Description = CopyBufferFunc(c)
		}
		if StopEditingFunc != nil {
			StopEditingFunc(c)
		}
		c.Send("\n\r")
	}
	if StartEditingFunc != nil {
		StartEditingFunc(ch, ch.Description)
	}
}

// DoPassword implements the 'password' command (C act_info.c:5078 do_password).
//
// Divergence from C: the C port takes `<new> <again>` only (the old-password
// check is commented out in src/act_info.c:5146-5152). This Go port enforces
// an additional old-password arg — `password <old> <new> <again>` — to align
// with the bcrypt migration already shipped and meet a modern security
// baseline. Length minimum is 6 (up from C's 5).
func DoPassword(ch *types.CharData, argument string) {
	if ch.IsNPC() {
		return
	}
	if ch.PCData == nil {
		return
	}

	// Case-preserving parse — OneArgument lowercases, which would corrupt
	// passwords containing capital letters.
	oldPwd, rest := util.CaseArgument(argument)
	newPwd, rest := util.CaseArgument(rest)
	againPwd, _ := util.CaseArgument(rest)

	if oldPwd == "" || newPwd == "" || againPwd == "" {
		ch.Send("Syntax: password <old> <new> <again>.\n\r")
		return
	}

	// Verify old password against the stored bcrypt hash. If Pwd is empty
	// treat it as a wrong-password mismatch rather than granting access.
	if ch.PCData.Pwd == "" ||
		bcrypt.CompareHashAndPassword([]byte(ch.PCData.Pwd), []byte(oldPwd)) != nil {
		ch.Send("Wrong password.\n\r")
		return
	}

	if newPwd != againPwd {
		ch.Send("Passwords don't match; password not changed.\n\r")
		return
	}

	if len(newPwd) < 6 {
		ch.Send("New password must be at least six characters long.\n\r")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPwd), BcryptCost)
	if err != nil {
		util.Bug("DoPassword: GenerateFromPassword: %v", err)
		ch.Send("Password change failed.\n\r")
		return
	}
	ch.PCData.Pwd = string(hash)

	// Match C: log the change (with site if available) and persist.
	if ch.Desc != nil && ch.Desc.Host != "" {
		util.LogString(fmt.Sprintf("%s changing password from site %s", ch.Name, ch.Desc.Host))
	} else {
		util.LogString(fmt.Sprintf("%s changing password", ch.Name))
	}

	if SaveFunc != nil {
		SaveFunc(ch)
	}

	ch.Send("Ok.\n\r")
}
