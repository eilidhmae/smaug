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
		util.Act("$n is no longer afk.", ch, nil, nil, nil, types.TO_CANSEE)
		return
	}
	ch.Act.Set(types.PLR_AFK)
	ch.Send("You are now afk.\n\r")
	util.Act("$n is now afk.", ch, nil, nil, nil, types.TO_CANSEE)
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
