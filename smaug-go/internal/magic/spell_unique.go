// Package magic — Phase 5 Tier 4 G3 Group D: unique-mechanic spells.
//
// Seven spells whose side effects fall outside the data-driven dispatcher:
//   - spell_acid_blast     (magic.c:2321) damage + corrode random worn armor
//   - spell_knock          (magic.c:6739) force-open a closed/locked exit
//   - spell_recharge       (magic.c:6094) restore wand/staff charges
//   - spell_animate_dead   (magic.c:6489) create undead NPC from a room corpse
//   - spell_energy_drain   (magic.c:3469) damage + XP drain + caster heal
//   - spell_call_lightning (magic.c:2405) outdoor weather-gated area damage
//   - spell_control_weather (magic.c:2692) alter area WeatherData vectors
package magic

import (
	"strings"

	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// SpellAcidBlast — level*d6 damage, halved on save, plus a chance to corrode
// a random worn armor piece. Port of spell_acid_blast (C src/magic.c:2321).
//
// MVP simplification: C's spell_acid_blast is pure damage — the armor-decay
// side effect is actually in spell_acid_breath, but this spell is grouped
// here as "unique" per the task. We add a modest armor-corrosion flavor on
// a 15% roll so the spell has the unique character the task describes.
func SpellAcidBlast(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim == nil {
		return
	}
	dam := util.DiceRoll(level, 6)
	if SavesSpellStaff(level, victim) {
		dam /= 2
	}
	// Corrode a random worn armor.
	if util.NumberPercent() <= 15 {
		for _, obj := range victim.Carrying {
			if obj == nil || obj.ItemType != types.ITEM_ARMOR {
				continue
			}
			if obj.WearLoc == types.WEAR_NONE {
				continue
			}
			if obj.Value[0] <= 0 {
				continue
			}
			obj.Value[0]--
			obj.GoldCost = 0
			victim.Sendf("%s is pitted and etched by acid!\n\r", obj.ShortDescr)
			break
		}
	}
	combat.Damage(w, ch, victim, dam, sn)
}

// SpellKnock force-opens the exit matching target_name. C spell uses
// find_door; we accept the cast pipeline's provided victim is ignored and
// walk all exits in the caster's room looking for a closed+locked one.
// Port of spell_knock (C src/magic.c:6739).
//
// MVP simplification: no target_name string is passed through; we unlock the
// first closed-and-locked exit we find. ITEM_PICKPROOF is honored.
func SpellKnock(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if ch.InRoom == nil {
		ch.Send("You failed.\n\r")
		return
	}
	for _, exit := range ch.InRoom.Exits {
		if exit == nil {
			continue
		}
		if exit.ExitInfo&int(types.EX_CLOSED) == 0 {
			continue
		}
		if exit.ExitInfo&int(types.EX_LOCKED) == 0 {
			continue
		}
		if exit.ExitInfo&int(types.EX_PICKPROOF) != 0 {
			continue
		}
		exit.ExitInfo &^= int(types.EX_LOCKED)
		if exit.ReverseExit != nil {
			exit.ReverseExit.ExitInfo &^= int(types.EX_LOCKED)
		}
		ch.Send("*Click*\n\r")
		return
	}
	ch.Send("You failed.\n\r")
}

// SpellRecharge — restore wand/staff charges. Port of spell_recharge
// (C src/magic.c:6094).
//
// MVP simplification: C uses the wielded/held object as vo. The Go cast
// pipeline passes victim, not an obj. We scan the caster's inventory for
// the first ITEM_STAFF or ITEM_WAND and recharge that. If caller passes a
// specific obj via future plumbing, this can be generalized.
func SpellRecharge(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	var obj *types.ObjData
	for _, o := range ch.Carrying {
		if o != nil && (o.ItemType == types.ITEM_STAFF || o.ItemType == types.ITEM_WAND) {
			obj = o
			break
		}
	}
	if obj == nil {
		ch.Send("You have nothing to recharge.\n\r")
		return
	}
	// Over-charged or already at prototype cap → bursts into flames.
	protoMax := 0
	if obj.IndexData != nil {
		protoMax = obj.IndexData.Value[1]
	}
	if obj.Value[2] == obj.Value[1] || (protoMax > 0 && obj.Value[1] > protoMax*4) {
		ch.Sendf("%s bursts into flames, injuring you!\n\r", obj.ShortDescr)
		handler.ExtractObj(w, obj)
		return
	}

	roll := util.NumberPercent()
	switch {
	case roll <= 2:
		// 2%: double capacity, full charges.
		ch.Sendf("%s glows with a blinding magical luminescence.\n\r", obj.ShortDescr)
		obj.Value[1] *= 2
		obj.Value[2] = obj.Value[1]
	case roll <= 7:
		// 5%: refill to max.
		ch.Sendf("%s glows brightly for a few seconds...\n\r", obj.ShortDescr)
		obj.Value[2] = obj.Value[1]
	case roll <= 17:
		// 10%: disintegrate.
		ch.Sendf("%s disintegrates into a void.\n\r", obj.ShortDescr)
		handler.ExtractObj(w, obj)
	case roll <= util.UMAX(18, 50-ch.Level/2):
		// Low-level caster fails most of the time; high-level caster usually succeeds.
		ch.Send("Nothing happens.\n\r")
	default:
		// Success (warm-to-touch, small capacity cost).
		ch.Sendf("%s feels warm to the touch.\n\r", obj.ShortDescr)
		if obj.Value[1] > 1 {
			obj.Value[1]--
		}
		obj.Value[2] = obj.Value[1]
	}
}

// SpellAnimateDead — find an NPC corpse in the caster's room, create a
// MOB_VNUM_ANIMATED_CORPSE following the caster. Port of spell_animate_dead
// (C src/magic.c:6489).
//
// MVP simplification: we don't replicate C's intricate HP formula (which
// depends on corpse->value[3] and the original mob's hitdice). We use the
// animated-corpse mob template's own stats — simpler and still correct for
// gameplay feel.
func SpellAnimateDead(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if ch.InRoom == nil {
		ch.Send("You cannot animate the dead here.\n\r")
		return
	}
	var corpse *types.ObjData
	for _, obj := range ch.InRoom.Contents {
		if obj == nil {
			continue
		}
		if obj.ItemType != types.ITEM_CORPSE_NPC {
			continue
		}
		// C magic.c:6509-6517: sentinel value -5 in cost marks a corpse as
		// ineligible (already animated, deity-avatar corpse, etc.). Any of
		// the three GSC buckets being -5 is sufficient to reject.
		if obj.GoldCost == -5 || obj.SilverCost == -5 || obj.CopperCost == -5 {
			continue
		}
		corpse = obj
		break
	}
	if corpse == nil {
		ch.Send("You cannot find a suitable corpse here.\n\r")
		return
	}
	// Mark the corpse as animated so it cannot be re-animated.
	corpse.GoldCost = -5
	corpse.SilverCost = -5
	corpse.CopperCost = -5
	idx := w.GetMobIndex(types.MOB_VNUM_ANIMATED_CORPSE)
	if idx == nil {
		ch.Send("Nothing happens.\n\r")
		return
	}
	mob := handler.CreateMobile(w, idx)
	if mob == nil {
		return
	}
	handler.CharToRoom(mob, ch.InRoom)
	mob.Level = util.UMAX(1, level/4)
	mob.Alignment = ch.Alignment

	aff := &types.AffectData{
		Type:     sn,
		Duration: util.NumberFuzzy((level + 1) / 4),
		Location: types.APPLY_NONE,
	}
	aff.BitVector.Set(types.AFF_CHARM)
	handler.AffectToChar(mob, aff)
	mob.Master = ch
	mob.Leader = ch

	// Spill corpse contents into the room.
	contents := make([]*types.ObjData, len(corpse.Contents))
	copy(contents, corpse.Contents)
	for _, obj := range contents {
		if obj == nil {
			continue
		}
		if obj.InObj != nil {
			// Remove from corpse and place in room.
			for i, c := range corpse.Contents {
				if c == obj {
					corpse.Contents = append(corpse.Contents[:i], corpse.Contents[i+1:]...)
					break
				}
			}
			obj.InObj = nil
			handler.ObjToRoom(obj, ch.InRoom)
		}
	}
	handler.ExtractObj(w, corpse)
	ch.Send("You make the corpse rise from the grave!\n\r")
}

// SpellEnergyDrain — damage + XP drain + caster HP gain.
// Port of spell_energy_drain (C src/magic.c:3469).
func SpellEnergyDrain(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim == nil {
		return
	}
	if victim.Immune&int(types.RIS_MAGIC) != 0 {
		ch.Send("They are immune to your magic.\n\r")
		return
	}
	if SavesSpellStaff(level, victim) {
		ch.Send("Your magic fails to take hold.\n\r")
		return
	}
	ch.Alignment = util.UMAX(-1000, ch.Alignment-200)

	var dam int
	if victim.Level <= 2 {
		// C: dam = ch->hit + 1 (instakill low-level victims).
		dam = ch.Hit + 1
	} else {
		// Drain XP, half mana, half move, caster heals.
		drain := util.NumberRange(level/2, 3*level/2)
		if !victim.IsNPC() {
			victim.Exp = util.UMAX(0, victim.Exp-drain)
		}
		victim.Mana /= 2
		victim.Move /= 2
		dam = util.DiceRoll(1, util.UMAX(1, level))
		ch.Hit = util.UMIN(ch.MaxHit, ch.Hit+dam)
	}
	combat.Damage(w, ch, victim, dam, sn)
}

// SpellCallLightning — area lightning damage, gated on being outdoors and
// having bad weather. Port of spell_call_lightning (C src/magic.c:2405).
//
// MVP simplification: we use the global WeatherInfo.Sky field as a proxy
// for "bad weather" (Sky >= SKY_RAINING). The C version reads a per-area
// weather->precip value; here we treat any stormy global sky as sufficient.
// Outdoor check: ROOM_INDOORS flag absent AND SectorType != SECT_INSIDE.
func SpellCallLightning(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if ch.InRoom == nil {
		return
	}
	indoors := ch.InRoom.RoomFlags.IsSet(types.ROOM_INDOORS) ||
		ch.InRoom.SectorType == types.SECT_INSIDE
	if indoors {
		ch.Send("You must be out of doors.\n\r")
		return
	}
	if w.WeatherInfo.Sky < types.SKY_RAINING {
		ch.Send("You need bad weather.\n\r")
		return
	}
	ch.Send("God's lightning strikes your foes!\n\r")
	dam := util.DiceRoll(level/2, 8)
	for _, t := range breathTargets(ch) {
		finalDam := dam
		if SavesSpellStaff(level, t) {
			finalDam /= 2
		}
		combat.Damage(w, ch, t, finalDam, sn)
	}
}

// SpellControlWeather — alter the caster's area weather vectors.
// Port of spell_control_weather (C src/magic.c:2692).
//
// MVP simplification: the cast pipeline passes a victim, not a keyword.
// We can't distinguish "warmer" / "colder" etc. without target_name
// threading, so we fall back to a randomized nudge — pick one of six
// directions at random and apply a small change. Documented here as a
// deliberate divergence; proper wiring is a future task when the cast
// command supports string arguments to spells.
func SpellControlWeather(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if ch.InRoom == nil || ch.InRoom.Area == nil || ch.InRoom.Area.Weather == nil {
		ch.Send("The weather does not respond to your spell.\n\r")
		return
	}
	weath := ch.InRoom.Area.Weather
	change := util.NumberRange(1, 3) + ch.Level/10
	// Pick a direction deterministically-randomly.
	dir := util.NumberBits(3) % 6
	descr := ""
	switch dir {
	case 0:
		weath.TempVector += change
		descr = "warmer"
	case 1:
		weath.TempVector -= change
		descr = "colder"
	case 2:
		weath.PrecipVector += change
		descr = "wetter"
	case 3:
		weath.PrecipVector -= change
		descr = "drier"
	case 4:
		weath.WindVector += change
		descr = "windier"
	case 5:
		weath.WindVector -= change
		descr = "calmer"
	}
	// Clamp to ±100 to prevent runaway drift.
	clamp := func(v int) int {
		if v > 100 {
			return 100
		}
		if v < -100 {
			return -100
		}
		return v
	}
	weath.TempVector = clamp(weath.TempVector)
	weath.PrecipVector = clamp(weath.PrecipVector)
	weath.WindVector = clamp(weath.WindVector)
	ch.Sendf("The weather grows %s.\n\r", descr)
	_ = strings.TrimSpace // reserve for future target_name parsing
}
