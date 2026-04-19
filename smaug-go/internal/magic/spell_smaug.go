// Package magic — data-driven spell dispatcher (spell_smaug).
//
// Port of spell_smaug() from src/magic.c:7882. The C implementation reads
// metadata packed into skill->info (SPELL_DAMAGE/ACTION/CLASS/POWER/SAVE bits,
// non-NEWSPELLS layout) plus skill->flags (SF_*), skill->affects,
// skill->saves/save-effect, and skill->dice to route to one of several
// sub-dispatchers: spell_attack, spell_affect, spell_area_attack,
// spell_create_obj, spell_create_mob.
//
// The Go port uses methods on SkillType (SpellAction(), SpellClass(),
// SpellDamageType(), SpellPower(), SpellSave(), HasFlag()) to unpack the same
// Info bits, so no new .dat keywords or SkillType fields are introduced — Info
// and Flags alone carry the dispatch metadata, matching the on-disk format.
package magic

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// effectiveSave returns the save-effect enum. Prefers Info-bits (authoritative
// in C), falls back to the struct field when Info bits are SE_NONE.
func effectiveSave(skill *types.SkillType) int {
	if se := skill.SpellSave(); se != types.SE_NONE {
		return se
	}
	return skill.SaveEffect
}

// SpellSmaug is the data-driven spell dispatcher that reads skill metadata
// (Info-packed SD/SA/SC/SP/SSAVE enums, Flags SF_* bits, Affects, SaveType,
// DiceFormula) and routes to the appropriate sub-dispatcher. Registered in
// spellRegistry as "spell_smaug".
//
// Structure mirrors C src/magic.c:7895 — dispatch primarily on skill.Target,
// then on SpellAction/SpellClass/SpellDamageType inside each branch.
func SpellSmaug(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if w == nil || ch == nil {
		return
	}
	if sn < 0 || sn >= len(w.Skills) {
		util.Bug("SpellSmaug: bad sn %d", sn)
		return
	}
	skill := w.Skills[sn]
	if skill == nil {
		util.Bug("SpellSmaug: nil skill for sn %d", sn)
		return
	}

	switch skill.Target {
	case types.TAR_IGNORE:
		// Offensive area spell: SA_DESTROY+SC_LIFE or SA_CREATE+SC_DEATH with SF_AREA.
		if skill.HasFlag(types.SF_AREA) &&
			((skill.SpellAction() == types.SA_DESTROY && skill.SpellClass() == types.SC_LIFE) ||
				(skill.SpellAction() == types.SA_CREATE && skill.SpellClass() == types.SC_DEATH)) {
			spellAreaAttack(w, skill, sn, level, ch)
			return
		}
		// SA_CREATE branch: object (SF_OBJECT) or mob (SC_LIFE).
		if skill.SpellAction() == types.SA_CREATE {
			if skill.HasFlag(types.SF_OBJECT) {
				spellCreateObj(w, skill, sn, level, ch)
				return
			}
			if skill.SpellClass() == types.SC_LIFE {
				spellCreateMob(w, skill, sn, level, ch)
				return
			}
		}
		// SF_CHARACTER with no distinct victim → affect self (group/area flavor).
		if skill.HasFlag(types.SF_CHARACTER) && victim != nil {
			spellAffect(w, skill, sn, level, ch, victim)
			return
		}
		// Fallback: apply as a self-affect (C's spell_affect(ch, vo) path).
		spellAffect(w, skill, sn, level, ch, ch)

	case types.TAR_CHAR_OFFENSIVE:
		// Bare area flag still forces area attack even if miscategorized.
		if skill.HasFlag(types.SF_AREA) {
			spellAreaAttack(w, skill, sn, level, ch)
			return
		}
		// Damage spell: SA_DESTROY+SC_LIFE or SA_CREATE+SC_DEATH.
		if (skill.SpellAction() == types.SA_DESTROY && skill.SpellClass() == types.SC_LIFE) ||
			(skill.SpellAction() == types.SA_CREATE && skill.SpellClass() == types.SC_DEATH) {
			spellAttack(w, skill, sn, level, ch, victim)
			return
		}
		// Otherwise a nasty affect (curse, weaken, blindness, etc.).
		spellAffect(w, skill, sn, level, ch, victim)

	case types.TAR_CHAR_DEFENSIVE, types.TAR_CHAR_SELF:
		// C emits a "can't concentrate" message for SF_NOFIGHT while in
		// combat; enforce here too.
		if skill.HasFlag(types.SF_NOFIGHT) && isFighting(ch) {
			ch.Send("You can't concentrate enough for that!\n\r")
			return
		}
		// Cure-poison: SA_DESTROY + SD_POISON on a defensive target.
		if victim != nil && skill.SpellAction() == types.SA_DESTROY &&
			skill.SpellDamageType() == types.SD_POISON {
			if curePoison(w, skill, sn, level, ch, victim) {
				return
			}
			return
		}
		// Cure-blindness: SA_DESTROY + SC_ILLUSION.
		if victim != nil && skill.SpellAction() == types.SA_DESTROY &&
			skill.SpellClass() == types.SC_ILLUSION {
			if cureBlindness(w, skill, sn, level, ch, victim) {
				return
			}
			return
		}
		spellAffect(w, skill, sn, level, ch, victim)

	case types.TAR_OBJ_INV:
		// spell_obj_inv is Tier-4 scope. Log and inform, don't crash.
		util.Bug("SpellSmaug: TAR_OBJ_INV not implemented for %q", skill.Name)
		ch.Send("That spell type is not yet implemented.\n\r")

	default:
		util.Bug("SpellSmaug: unknown target %d for %q", skill.Target, skill.Name)
	}
}

// isFighting reports whether ch is in any actively-engaged combat position.
func isFighting(ch *types.CharData) bool {
	switch ch.Position {
	case types.POS_FIGHTING, types.POS_EVASIVE, types.POS_DEFENSIVE,
		types.POS_AGGRESSIVE, types.POS_BERSERK:
		return true
	}
	return false
}

// curePoison strips gsn_poison if present; returns true if the spell resolved
// (success or fail-with-message). Simplified port of the TAR_CHAR_DEFENSIVE
// SD_POISON branch in src/magic.c:7964.
func curePoison(w *world.World, skill *types.SkillType, sn, level int, ch, victim *types.CharData) bool {
	poisonSN := findSkillSN(w, "poison")
	if poisonSN < 0 {
		return false
	}
	if charIsAffected(victim, poisonSN) {
		handler.AffectStrip(victim, poisonSN)
		if skill.HitChar != "" {
			ch.Send(skill.HitChar + "\n\r")
		}
		return true
	}
	if skill.MissChar != "" {
		ch.Send(skill.MissChar + "\n\r")
	}
	return true
}

// cureBlindness strips gsn_blindness if present. Port of the SC_ILLUSION
// branch in src/magic.c:7978.
func cureBlindness(w *world.World, skill *types.SkillType, sn, level int, ch, victim *types.CharData) bool {
	blindSN := findSkillSN(w, "blindness")
	if blindSN < 0 {
		return false
	}
	if charIsAffected(victim, blindSN) {
		handler.AffectStrip(victim, blindSN)
		if skill.HitChar != "" {
			ch.Send(skill.HitChar + "\n\r")
		}
		return true
	}
	if skill.MissChar != "" {
		ch.Send(skill.MissChar + "\n\r")
	}
	return true
}

// charIsAffected returns true if ch has an affect of type sn.
func charIsAffected(ch *types.CharData, sn int) bool {
	for _, a := range ch.Affects {
		if a != nil && a.Type == sn {
			return true
		}
	}
	return false
}

// findSkillSN does a case-insensitive name lookup through world.Skills.
// Returns -1 if not found.
func findSkillSN(w *world.World, name string) int {
	for i, sk := range w.Skills {
		if sk == nil {
			continue
		}
		if strings.EqualFold(sk.Name, name) {
			return i
		}
	}
	return -1
}

// spellAffect applies AffectData derived from skill.Affects to the target.
// Honors SaveType + SE_NEGATE to block non-self affects on a made save.
// Loose port of spell_affect() in src/magic.c.
func spellAffect(w *world.World, skill *types.SkillType, sn, level int, ch, victim *types.CharData) {
	target := victim
	if target == nil {
		target = ch
	}

	// Save roll — only meaningful for victims distinct from the caster.
	if skill.SaveType != types.SS_NONE && target != ch {
		if rollSave(level, target, skill.SaveType) {
			if effectiveSave(skill) == types.SE_NEGATE {
				if skill.MissChar != "" {
					ch.Send(skill.MissChar + "\n\r")
				} else {
					ch.Send("Your spell fails to take hold.\n\r")
				}
				return
			}
		}
	}

	// Apply each affect record.
	for _, sa := range skill.Affects {
		if sa == nil {
			continue
		}
		aff := &types.AffectData{
			Type:     sn,
			Duration: parseDiceExpr(sa.Duration, level),
			Location: sa.Location,
			Modifier: parseDiceExpr(sa.Modifier, level),
		}
		// SmaugAff.BitVector stores a single bit index (from fread_skill's
		// power-of-two → index conversion in src/tables.c:4117). -1 means
		// "no bitvector"; anything in [0,127] is a valid AFF_ index.
		if sa.BitVector >= 0 {
			aff.BitVector.Set(sa.BitVector)
		}
		handler.AffectToChar(target, aff)
	}

	if skill.HitChar != "" {
		ch.Send(skill.HitChar + "\n\r")
	}
	if skill.HitVict != "" && target != ch {
		target.Send(skill.HitVict + "\n\r")
	}
}

// spellAttack rolls damage from DiceFormula, applies save modifiers, damages.
// Loose port of spell_attack() in src/magic.c:6986.
func spellAttack(w *world.World, skill *types.SkillType, sn, level int, ch, victim *types.CharData) {
	if victim == nil {
		return
	}

	// SF_PKSENSITIVE halves the effective save-level when both caster and
	// victim are PCs (src/magic.c:2343). It does NOT skip the victim.
	effLevel := level
	if skill.HasFlag(types.SF_PKSENSITIVE) && !ch.IsNPC() && !victim.IsNPC() {
		effLevel = level / 2
	}

	saved := false
	if skill.SaveType != types.SS_NONE {
		saved = rollSave(effLevel, victim, skill.SaveType)
	}

	saveEffect := effectiveSave(skill)
	if saved && saveEffect == types.SE_NEGATE {
		if skill.MissChar != "" {
			ch.Send(skill.MissChar + "\n\r")
		}
		return
	}

	dam := parseDiceExpr(skill.DiceFormula, level)
	if dam < 0 {
		dam = 0
	}

	if saved {
		switch saveEffect {
		case types.SE_3QTRDAM:
			dam = dam * 3 / 4
		case types.SE_HALFDAM:
			dam /= 2
		case types.SE_QUARTERDAM:
			dam /= 4
		case types.SE_EIGHTHDAM:
			dam /= 8
		case types.SE_ABSORB:
			// Victim heals for dam; no damage dealt.
			if victim.Hit+dam > victim.MaxHit {
				victim.Hit = victim.MaxHit
			} else if victim.Hit+dam < 0 {
				victim.Hit = 0
			} else {
				victim.Hit = victim.Hit + dam
			}
			return
		case types.SE_REFLECT:
			// Redirect to caster with roles swapped. Guard against infinite
			// recursion: a reflected spell is a straight damage call, not a
			// recursive spellAttack, to avoid the victim's own save chain
			// bouncing it back again.
			combat.Damage(w, victim, ch, dam, sn)
			return
		}
	}
	combat.Damage(w, ch, victim, dam, sn)
}

// spellAreaAttack hits every eligible character in the caster's room.
// Caster is always excluded; SF_NOMOB skips NPCs. SF_PKSENSITIVE does NOT
// skip players (it halves the effective save-level inside spellAttack).
// Port of spell_area_attack() in src/magic.c:7059.
func spellAreaAttack(w *world.World, skill *types.SkillType, sn, level int, ch *types.CharData) {
	if ch.InRoom == nil {
		return
	}
	// Snapshot occupants: Damage may remove corpses/chars, and we iterate
	// safely even if the slice mutates during the loop.
	targets := make([]*types.CharData, 0, len(ch.InRoom.People))
	for _, p := range ch.InRoom.People {
		if p == nil || p == ch {
			continue
		}
		if skill.HasFlag(types.SF_NOMOB) && p.IsNPC() {
			continue
		}
		targets = append(targets, p)
	}
	for _, t := range targets {
		spellAttack(w, skill, sn, level, ch, t)
	}
}

// spellCreateObj creates an object of vnum = skill.Value in the caster's
// inventory. Port of spell_create_obj() in src/magic.c.
func spellCreateObj(w *world.World, skill *types.SkillType, sn, level int, ch *types.CharData) {
	vnum := skill.Value
	if vnum <= 0 {
		util.Bug("spellCreateObj: skill %q has no vnum (value=%d)", skill.Name, skill.Value)
		return
	}
	idx := w.ObjIndex[vnum]
	if idx == nil {
		util.Bug("spellCreateObj: obj vnum %d not found for %q", vnum, skill.Name)
		return
	}

	// SpellPower tier scales object level (C src/magic.c:7755).
	lvl := 10
	switch skill.SpellPower() {
	case types.SP_MINOR:
		lvl = 0
	case types.SP_GREATER:
		lvl = level / 2
	case types.SP_MAJOR:
		lvl = level
	}
	obj := handler.CreateObject(w, idx, lvl)
	if obj == nil {
		return
	}
	handler.ObjToChar(obj, ch)
	if skill.HitChar != "" {
		ch.Send(skill.HitChar + "\n\r")
	}
}

// spellCreateMob creates a mobile of vnum = skill.Value in the caster's room.
// Port of spell_create_mob() in src/magic.c.
func spellCreateMob(w *world.World, skill *types.SkillType, sn, level int, ch *types.CharData) {
	vnum := skill.Value
	if vnum <= 0 {
		util.Bug("spellCreateMob: skill %q has no vnum (value=%d)", skill.Name, skill.Value)
		return
	}
	idx := w.MobIndex[vnum]
	if idx == nil {
		util.Bug("spellCreateMob: mob vnum %d not found for %q", vnum, skill.Name)
		return
	}
	mob := handler.CreateMobile(w, idx)
	if mob == nil || ch.InRoom == nil {
		return
	}
	handler.CharToRoom(mob, ch.InRoom)
	if skill.HitChar != "" {
		ch.Send(skill.HitChar + "\n\r")
	}
}

// rollSaveFunc is the package-level save-resolver hook; tests may override
// it to make save outcomes deterministic.
var rollSaveFunc = defaultRollSave

// rollSave dispatches to the correct Saves* helper based on SS_* type.
// Returns true if the victim successfully resists.
func rollSave(level int, victim *types.CharData, saveType int) bool {
	return rollSaveFunc(level, victim, saveType)
}

func defaultRollSave(level int, victim *types.CharData, saveType int) bool {
	switch saveType {
	case types.SS_POISON_DEATH:
		return SavesPoisonDeath(level, victim)
	case types.SS_ROD_WANDS:
		return SavesWands(level, victim)
	case types.SS_PARA_PETRI:
		return SavesParaPetri(level, victim)
	case types.SS_BREATH:
		return SavesBreath(level, victim)
	case types.SS_SPELL_STAFF:
		return SavesSpellStaff(level, victim)
	default:
		return false
	}
}

// parseDiceExpr evaluates a SMAUG expression string at the given level.
// Thin wrapper over util.DiceParse — the grammar and implementation now
// live there so other packages (e.g. handler/polymorph) can share it
// without creating an import cycle with the magic package.
//
// Plan plan-phase6-polymorph.md §G0.2 (promoted from package-private).
func parseDiceExpr(s string, level int) int {
	return util.DiceParse(s, level)
}

// parseDiceOrInt is an alias retained for back-compat with existing callers.
func parseDiceOrInt(s string, level int) int { return util.DiceParse(s, level) }
