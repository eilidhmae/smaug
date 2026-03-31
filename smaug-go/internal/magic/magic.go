// Package magic implements the SMAUG spell/skill system.
package magic

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// SpellFunc is the signature for all spell implementations.
type SpellFunc func(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData)

// spellRegistry maps spell function names (from skills.dat) to Go functions.
var spellRegistry = map[string]SpellFunc{
	"spell_magic_missile": SpellMagicMissile,
	"spell_cure_light":    SpellCureLight,
	"spell_cure_serious":  SpellCureSerious,
	"spell_cure_critical": SpellCureCritical,
	"spell_fireball":      SpellFireball,
	"spell_armor":         SpellArmor,
	"spell_bless":         SpellBless,
	"spell_curse":         SpellCurse,
	"spell_poison":        SpellPoison,
	"spell_blindness":     SpellBlindness,
	"spell_sanctuary":     SpellSanctuary,
	"spell_dispel_magic":  SpellDispelMagic,
}

// FindSpellFunc looks up a spell function by its code name.
func FindSpellFunc(name string) SpellFunc {
	return spellRegistry[name]
}

// FindSpellByName finds a spell's index in the world skill list by name.
func FindSpellByName(w *world.World, name string) int {
	name = strings.ToLower(name)
	for i, sk := range w.Skills {
		if sk != nil && sk.Type == types.SKILL_SPELL &&
			strings.ToLower(sk.Name) == name {
			return i
		}
	}
	// Prefix match
	for i, sk := range w.Skills {
		if sk != nil && sk.Type == types.SKILL_SPELL &&
			strings.HasPrefix(strings.ToLower(sk.Name), name) {
			return i
		}
	}
	return -1
}

// SavesSpellStaff checks if a victim makes their saving throw vs spells.
// Returns true if the save succeeds (spell is resisted).
func SavesSpellStaff(level int, victim *types.CharData) bool {
	save := 50 + (victim.Level-level-victim.SavingSpellStaff)*5
	save = util.URANGE(5, save, 95)
	return util.NumberPercent() <= save
}

// SavesPoisonDeath checks if a victim makes their saving throw vs poison/death.
func SavesPoisonDeath(level int, victim *types.CharData) bool {
	save := 50 + (victim.Level-level-victim.SavingPoisonDeath)*5
	save = util.URANGE(5, save, 95)
	return util.NumberPercent() <= save
}

// --- Healing Spells ---

// SpellCureLight heals 1d8 + level/3 HP.
func SpellCureLight(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	heal := util.DiceRoll(1, 8) + level/3
	victim.Hit = util.UMIN(victim.Hit+heal, victim.MaxHit)
	victim.Send("You feel a little better!\n\r")
}

// SpellCureSerious heals 2d8 + level/2 HP.
func SpellCureSerious(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	heal := util.DiceRoll(2, 8) + level/2
	victim.Hit = util.UMIN(victim.Hit+heal, victim.MaxHit)
	victim.Send("You feel better!\n\r")
}

// SpellCureCritical heals 3d8 + level HP.
func SpellCureCritical(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	heal := util.DiceRoll(3, 8) + level
	victim.Hit = util.UMIN(victim.Hit+heal, victim.MaxHit)
	victim.Send("You feel much better!\n\r")
}

// --- Damage Spells ---

// SpellMagicMissile deals level-scaled damage (no save).
func SpellMagicMissile(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	dam := util.UMAX(3, level*2/3)
	dam = util.NumberRange(dam/2, dam*2)
	if dam < 1 {
		dam = 1
	}
	combat.Damage(w, ch, victim, dam, types.TYPE_HIT+sn)
}

// SpellFireball deals level-scaled damage (save for half).
func SpellFireball(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	dam := util.UMAX(30, level*3)
	dam = util.NumberRange(dam/2, dam*2)
	if SavesSpellStaff(level, victim) {
		dam /= 2
	}
	combat.Damage(w, ch, victim, dam, types.TYPE_HIT+sn)
}

// --- Buff Spells ---

// SpellArmor improves AC by -20 for a duration.
func SpellArmor(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim.AffectedBy.IsSet(types.AFF_PROTECT) {
		victim.Send("You are already armored.\n\r")
		return
	}
	handler.AffectToChar(victim, &types.AffectData{
		Type:     sn,
		Duration: 24 + level,
		Location: types.APPLY_AC,
		Modifier: -20,
	})
	victim.Send("You feel someone protecting you.\n\r")
}

// SpellBless improves hitroll by 1+level/8 for a duration.
func SpellBless(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	handler.AffectToChar(victim, &types.AffectData{
		Type:     sn,
		Duration: 12 + level,
		Location: types.APPLY_HITROLL,
		Modifier: 1 + level/8,
	})
	victim.Send("You feel righteous.\n\r")
}

// SpellSanctuary grants AFF_SANCTUARY (damage halved).
func SpellSanctuary(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim.AffectedBy.IsSet(types.AFF_SANCTUARY) {
		victim.Send("You are already in sanctuary.\n\r")
		return
	}
	aff := &types.AffectData{
		Type:     sn,
		Duration: 16 + level/2,
		Location: types.APPLY_NONE,
	}
	aff.BitVector.Set(types.AFF_SANCTUARY)
	handler.AffectToChar(victim, aff)
	victim.Send("You are surrounded by a white aura.\n\r")
}

// --- Debuff Spells ---

// SpellCurse applies -1 hitroll/-1 damroll with save.
func SpellCurse(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim.AffectedBy.IsSet(types.AFF_CURSE) {
		ch.Send("They are already cursed.\n\r")
		return
	}
	if SavesSpellStaff(level, victim) {
		ch.Send("Your magic fails to take hold.\n\r")
		return
	}
	aff := &types.AffectData{
		Type:     sn,
		Duration: 4 + level/4,
		Location: types.APPLY_HITROLL,
		Modifier: -1,
	}
	aff.BitVector.Set(types.AFF_CURSE)
	handler.AffectToChar(victim, aff)
	victim.Send("You feel unclean.\n\r")
}

// SpellPoison applies AFF_POISON with -2 STR, with save.
func SpellPoison(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim.AffectedBy.IsSet(types.AFF_POISON) {
		ch.Send("They are already poisoned.\n\r")
		return
	}
	if SavesPoisonDeath(level, victim) {
		ch.Send("Your magic fails to take hold.\n\r")
		return
	}
	aff := &types.AffectData{
		Type:     sn,
		Duration: level,
		Location: types.APPLY_STR,
		Modifier: -2,
	}
	aff.BitVector.Set(types.AFF_POISON)
	handler.AffectToChar(victim, aff)
	victim.Send("You feel very sick.\n\r")
}

// SpellBlindness applies AFF_BLIND with -4 hitroll, with save.
func SpellBlindness(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim.AffectedBy.IsSet(types.AFF_BLIND) {
		ch.Send("They are already blinded.\n\r")
		return
	}
	if SavesSpellStaff(level, victim) {
		ch.Send("Your magic fails to take hold.\n\r")
		return
	}
	aff := &types.AffectData{
		Type:     sn,
		Duration: 1 + level/3,
		Location: types.APPLY_HITROLL,
		Modifier: -4,
	}
	aff.BitVector.Set(types.AFF_BLIND)
	handler.AffectToChar(victim, aff)
	victim.Send("You are blinded!\n\r")
}

// SpellDispelMagic removes all affects from a target (save negates).
func SpellDispelMagic(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if SavesSpellStaff(level, victim) && victim != ch {
		ch.Send("Your magic fails to take hold.\n\r")
		return
	}
	for len(victim.Affects) > 0 {
		handler.AffectRemove(victim, victim.Affects[0])
	}
	victim.Send("Your enchantments fade away.\n\r")
	if ch != victim {
		ch.Sendf("You dispel %s's magic.\n\r", victim.Name)
	}
}
