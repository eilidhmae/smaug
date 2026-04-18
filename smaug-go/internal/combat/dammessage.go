package combat

import (
	"github.com/eilidhmae/smaug/internal/handler"
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

// isWieldingPoisoned returns true iff obj is the actively-wielded poisoned
// weapon for ch. Matches C fight.c:104-119: requires obj to be non-nil,
// ITEM_POISONED set, and pointer-identical to the character's primary wield
// (WEAR_WIELD) OR dual-wield (WEAR_DUAL_WIELD) slot. The identity check is
// load-bearing — an ad-hoc poisoned obj that is not currently equipped must
// not trigger the prefix.
func isWieldingPoisoned(ch *types.CharData, obj *types.ObjData) bool {
	if ch == nil || obj == nil {
		return false
	}
	if !obj.ExtraFlags.IsSet(types.ITEM_POISONED) {
		return false
	}
	return obj == handler.GetEqChar(ch, types.WEAR_WIELD) ||
		obj == handler.GetEqChar(ch, types.WEAR_DUAL_WIELD)
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
// in place of attackTable[dt-TYPE_HIT]. When obj is also the active poisoned
// wield (primary or dual), the attack word is prefixed with "poisoned "
// (C fight.c:4496-4512).
//
// Cross-room attacker: if ch.InRoom != victim.InRoom, the attacker is
// temporarily moved into the victim's room so TO_NOTVICT bystanders there
// see the broadcast, and restored afterwards via defer (C fight.c:4432-4438
// and 4590-4594).
//
// PCFLAG_GAG: a zero-damage miss is suppressed for gagged PCs on their own
// side (attacker-gag silences TO_CHAR, victim-gag silences TO_VICT); the
// bystander broadcast (TO_NOTVICT) is never gated. Positive damage is never
// silenced. Mirrors C fight.c:4481-4488.
func DamMessage(ch, victim *types.CharData, dam, dt int, obj *types.ObjData) {
	if ch == nil || victim == nil {
		return
	}

	// Cross-room swap (C fight.c:4432-4439). Temporarily move the attacker
	// into the victim's room so the TO_NOTVICT broadcast reaches bystanders
	// there. Wrapped in defer so the restore runs even if util.Act panics.
	if ch.InRoom != victim.InRoom && victim.InRoom != nil {
		wasInRoom := ch.InRoom
		handler.CharFromRoom(ch)
		handler.CharToRoom(ch, victim.InRoom)
		defer func() {
			if wasInRoom == nil {
				return
			}
			handler.CharFromRoom(ch)
			handler.CharToRoom(ch, wasInRoom)
		}()
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
	} else if dt > types.TYPE_HIT && isWieldingPoisoned(ch, obj) {
		// Poisoned wielded weapon (C fight.c:4496-4512). C uses the
		// attack_table entry — NOT the obj's short_descr — when emitting
		// this variant, so we do the same. The prefix is literal "poisoned"
		// and the buffers become "$n's poisoned <attack> ...".
		var attack string
		if dt >= types.TYPE_HIT && dt < types.TYPE_HIT+len(attackTable) {
			attack = attackTable[dt-types.TYPE_HIT]
		} else {
			attack = attackTable[0]
		}
		buf1 = "$n's poisoned " + attack + " " + vp + " $N" + string(punct)
		buf2 = "Your poisoned " + attack + " " + vs + " $N" + string(punct)
		buf3 = "$n's poisoned " + attack + " " + vp + " you" + string(punct)
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

	// PCFLAG_GAG self-suppress (C fight.c:4481-4488). Only a zero-damage
	// miss is silenced, and only on the gagged recipient's own side. NPCs
	// (and PCs with PCData == nil) are never gagged. TO_NOTVICT is always
	// delivered.
	gcflag := dam == 0 && !ch.IsNPC() && ch.PCData != nil &&
		(ch.PCData.Flags&int(types.PCFLAG_GAG)) != 0
	gvflag := dam == 0 && !victim.IsNPC() && victim.PCData != nil &&
		(victim.PCData.Flags&int(types.PCFLAG_GAG)) != 0

	util.Act(buf1, ch, victim, nil, nil, types.TO_NOTVICT)
	if !gcflag {
		util.Act(buf2, ch, victim, nil, nil, types.TO_CHAR)
	}
	if !gvflag {
		util.Act(buf3, ch, victim, nil, nil, types.TO_VICT)
	}
}
