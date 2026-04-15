package combat

import (
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// WorldRef is wired at boot so DamMessage can look up skills by sn.
// Tests may override this directly.
var WorldRef *world.World

// attackTable maps DAM_HIT..DAM_PEA (18 entries) to the singular noun used
// when the attacker is wielding a weapon of that damage type. Ports
// C src/const.c:422-428 (non-ENABLE_WEAPONPROF block).
var attackTable = [18]string{
	"hit",
	"slice", "stab", "slash", "whip", "claw",
	"blast", "pound", "crush", "grep", "bite",
	"pierce", "suction", "bolt", "arrow", "dart",
	"stone", "pea",
}

// Damage-intensity verb tables. 24 entries each, indexed by d_index (0-23).
// Ported verbatim from C src/const.c:431-476.
var sBladeMessages = [24]string{
	"miss", "barely scratch", "scratch", "nick", "cut", "hit", "tear",
	"rip", "gash", "lacerate", "hack", "maul", "rend", "decimate",
	"_mangle_", "_devastate_", "_cleave_", "_butcher_", "DISEMBOWEL",
	"DISFIGURE", "GUT", "EVISCERATE", "* SLAUGHTER *", "*** ANNIHILATE ***",
}

var pBladeMessages = [24]string{
	"misses", "barely scratches", "scratches", "nicks", "cuts", "hits",
	"tears", "rips", "gashes", "lacerates", "hacks", "mauls", "rends",
	"decimates", "_mangles_", "_devastates_", "_cleaves_", "_butchers_",
	"DISEMBOWELS", "DISFIGURES", "GUTS", "EVISCERATES", "* SLAUGHTERS *",
	"*** ANNIHILATES ***",
}

var sBluntMessages = [24]string{
	"miss", "barely scuff", "scuff", "pelt", "bruise", "strike", "thrash",
	"batter", "flog", "pummel", "smash", "maul", "bludgeon", "decimate",
	"_shatter_", "_devastate_", "_maim_", "_cripple_", "MUTILATE", "DISFIGURE",
	"MASSACRE", "PULVERIZE", "* OBLITERATE *", "*** ANNIHILATE ***",
}

var pBluntMessages = [24]string{
	"misses", "barely scuffs", "scuffs", "pelts", "bruises", "strikes",
	"thrashes", "batters", "flogs", "pummels", "smashes", "mauls",
	"bludgeons", "decimates", "_shatters_", "_devastates_", "_maims_",
	"_cripples_", "MUTILATES", "DISFIGURES", "MASSACRES", "PULVERIZES",
	"* OBLITERATES *", "*** ANNIHILATES ***",
}

var sGenericMessages = [24]string{
	"miss", "brush", "scratch", "graze", "nick", "jolt", "wound",
	"injure", "hit", "jar", "thrash", "maul", "decimate", "_traumatize_",
	"_devastate_", "_maim_", "_demolish_", "MUTILATE", "MASSACRE",
	"PULVERIZE", "DESTROY", "* OBLITERATE *", "*** ANNIHILATE ***",
	"**** SMITE ****",
}

var pGenericMessages = [24]string{
	"misses", "brushes", "scratches", "grazes", "nicks", "jolts", "wounds",
	"injures", "hits", "jars", "thrashes", "mauls", "decimates",
	"_traumatizes_",
	"_devastates_", "_maims_", "_demolishes_", "MUTILATES", "MASSACRES",
	"PULVERIZES", "DESTROYS", "* OBLITERATES *", "*** ANNIHILATES ***",
	"**** SMITES ****",
}

// sMessageTable / pMessageTable map w_index (0..17) -> verb table. Ported
// from C src/const.c:504-544 (non-ENABLE_WEAPONPROF block).
var sMessageTable = [18]*[24]string{
	&sGenericMessages, // hit
	&sBladeMessages,   // slice
	&sBladeMessages,   // stab
	&sBladeMessages,   // slash
	&sBluntMessages,   // whip
	&sBladeMessages,   // claw
	&sGenericMessages, // blast
	&sBluntMessages,   // pound
	&sBluntMessages,   // crush
	&sGenericMessages, // grep
	&sBladeMessages,   // bite
	&sBladeMessages,   // pierce
	&sBluntMessages,   // suction
	&sGenericMessages, // bolt
	&sGenericMessages, // arrow
	&sGenericMessages, // dart
	&sGenericMessages, // stone
	&sGenericMessages, // pea
}

var pMessageTable = [18]*[24]string{
	&pGenericMessages, // hit
	&pBladeMessages,   // slice
	&pBladeMessages,   // stab
	&pBladeMessages,   // slash
	&pBluntMessages,   // whip
	&pBladeMessages,   // claw
	&pGenericMessages, // blast
	&pBluntMessages,   // pound
	&pBluntMessages,   // crush
	&pGenericMessages, // grep
	&pBladeMessages,   // bite
	&pBladeMessages,   // pierce
	&pBluntMessages,   // suction
	&pGenericMessages, // bolt
	&pGenericMessages, // arrow
	&pGenericMessages, // dart
	&pGenericMessages, // stone
	&pGenericMessages, // pea
}

// topSn returns the number of registered skills/spells (C top_sn equivalent).
func topSn() int {
	if WorldRef == nil {
		return 0
	}
	return len(WorldRef.Skills)
}

// lookupSkill returns the skill at sn, or nil if out of range.
func lookupSkill(sn int) *types.SkillType {
	if WorldRef == nil {
		return nil
	}
	if sn < 0 || sn >= len(WorldRef.Skills) {
		return nil
	}
	return WorldRef.Skills[sn]
}

// DamMessage emits damage messages to ch, victim, and room based on the
// weapon type (dt) and the damage as a percentage of the victim's max HP.
// Ported from C src/fight.c:4410 (new_dam_message).
//
// dt values:
//   - TYPE_UNDEFINED (-1): treated as DAM_HIT (w_index 0).
//   - 0..top_sn-1: skill/spell sn; uses skill.NounDamage as the attack word
//     and skill's hit_*/miss_* strings when populated.
//   - TYPE_HIT..TYPE_HIT+17: weapon damage type; uses attackTable or obj's
//     ShortDescr (if obj != nil).
//
// obj (when non-nil and dt > TYPE_HIT) is substituted as the attack word
// in place of attackTable[dt-TYPE_HIT].
//
// TODO(phase5-tier4): Port the was_in_room swap in fight.c:4432-4438 so a
// cross-room attacker appears to the victim's room while messages fire.
// For now, if the attacker is in a different room we emit only to the victim.
//
// TODO(phase5-tier4): Port PCFLAG_GAG handling in fight.c:4481-4486 so gagged
// players can suppress their own miss messages.
//
// TODO(phase5-tier4): Port is_wielding_poisoned prefix in fight.c:4496-4512.
func DamMessage(ch, victim *types.CharData, dam, dt int, obj *types.ObjData) {
	if ch == nil || victim == nil {
		return
	}

	// Compute damage percentage (fight.c:4426-4430).
	var dampc int
	if dam == 0 || victim.MaxHit == 0 {
		dampc = 0
	} else {
		dampc = (dam*1000)/victim.MaxHit + (50 - (victim.Hit*50)/victim.MaxHit)
	}

	// w_index (fight.c:4441-4459). TYPE_UNDEFINED is not a C case but our
	// callers pass it through OneHit, so map it to DAM_HIT up front.
	w_index := 0
	if dt == types.TYPE_UNDEFINED {
		w_index = 0
		dt = types.TYPE_HIT
	} else if dt >= 0 && dt < topSn() {
		// Skill/spell sn. Use DAM_HIT (generic) verb table; attack word
		// comes from skill.NounDamage below. Note this diverges slightly
		// from C fight.c:4442 (`dt > 0 && dt < top_sn`) so sn=0 is valid
		// in Go — C treated sn=0 as "reserved" and fell to the bug path.
		w_index = 0
	} else if dt >= types.TYPE_HIT && dt < types.TYPE_HIT+len(attackTable) {
		w_index = dt - types.TYPE_HIT
	} else {
		vnum := 0
		if ch.InRoom != nil {
			vnum = ch.InRoom.Vnum
		}
		util.Bug("Dam_message: bad dt %d from %s in %d.", dt, ch.Name, vnum)
		dt = types.TYPE_HIT
		w_index = 0
	}

	// d_index (fight.c:4461-4473).
	var d_index int
	switch {
	case dam == 0:
		d_index = 0
	case dampc < 0:
		d_index = 1
	case dampc <= 100:
		d_index = 1 + dampc/10
	case dampc <= 200:
		d_index = 11 + (dampc-100)/20
	case dampc <= 900:
		d_index = 16 + (dampc-200)/100
	default:
		d_index = 23
	}

	vs := sMessageTable[w_index][d_index]
	vp := pMessageTable[w_index][d_index]
	punct := byte('.')
	if dampc > 30 {
		punct = '!'
	}

	// Skill lookup (fight.c:4488-4489). We intentionally include sn=0 here —
	// see w_index comment above.
	var skill *types.SkillType
	if dt >= 0 && dt < topSn() {
		skill = lookupSkill(dt)
	}

	var buf1, buf2, buf3 string

	// Cross-room guard. C's version does a char_from_room/char_to_room swap
	// so the attacker is temporarily placed with the victim for message
	// routing. We skip that dance and emit only to the victim. This keeps
	// the TO_NOTVICT broadcast in the victim's room, and skips TO_CHAR.
	crossRoom := ch.InRoom != victim.InRoom

	if dt == types.TYPE_HIT {
		buf1 = "$n " + vp + " $N" + string(punct)
		buf2 = "You " + vs + " $N" + string(punct)
		buf3 = "$n " + vp + " you" + string(punct)
	} else if skill != nil {
		attack := skill.NounDamage
		if dam == 0 {
			found := false
			if skill.MissChar != "" {
				util.Act(skill.MissChar, ch, victim, nil, nil, types.TO_CHAR)
				found = true
			}
			if skill.MissVict != "" {
				util.Act(skill.MissVict, ch, victim, nil, nil, types.TO_VICT)
				found = true
			}
			if skill.MissRoom != "" {
				if skill.MissRoom != "supress" {
					util.Act(skill.MissRoom, ch, victim, nil, nil, types.TO_NOTVICT)
				}
				found = true
			}
			if found {
				return
			}
		} else {
			// C fight.c:4549-4558: hit_* strings fire WITHOUT returning — the
			// code falls through to buf1/buf2/buf3 below and also emits the
			// generic "$n's <noun_damage> <verb> $N" line. Only the miss path
			// above short-circuits.
			if skill.HitChar != "" {
				util.Act(skill.HitChar, ch, victim, nil, nil, types.TO_CHAR)
			}
			if skill.HitVict != "" {
				util.Act(skill.HitVict, ch, victim, nil, nil, types.TO_VICT)
			}
			if skill.HitRoom != "" {
				util.Act(skill.HitRoom, ch, victim, nil, nil, types.TO_NOTVICT)
			}
		}
		// Fall through to the generic attacker-verb format.
		if attack == "" {
			attack = attackTable[0]
		}
		buf1 = "$n's " + attack + " " + vp + " $N" + string(punct)
		buf2 = "Your " + attack + " " + vs + " $N" + string(punct)
		buf3 = "$n's " + attack + " " + vp + " you" + string(punct)
	} else {
		// dt > TYPE_HIT (weapon damage type). Attack word is obj's
		// ShortDescr when present, else the generic attack table entry.
		var attack string
		if obj != nil && obj.ShortDescr != "" {
			attack = obj.ShortDescr
		} else if dt >= types.TYPE_HIT && dt < types.TYPE_HIT+len(attackTable) {
			attack = attackTable[dt-types.TYPE_HIT]
		} else {
			attack = attackTable[0]
		}
		buf1 = "$n's " + attack + " " + vp + " $N" + string(punct)
		buf2 = "Your " + attack + " " + vs + " $N" + string(punct)
		buf3 = "$n's " + attack + " " + vp + " you" + string(punct)
	}

	if crossRoom {
		// TODO: port was_in_room swap. For now, emit only to victim.
		util.Act(buf3, ch, victim, nil, nil, types.TO_VICT)
		return
	}

	util.Act(buf1, ch, victim, nil, nil, types.TO_NOTVICT)
	util.Act(buf2, ch, victim, nil, nil, types.TO_CHAR)
	util.Act(buf3, ch, victim, nil, nil, types.TO_VICT)
}
