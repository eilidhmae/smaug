// Package magic — Phase 5 Tier 4 G3 Group B: area / breath attacks.
//
// Ports six breath-style area spells from src/magic.c:
//   - spell_acid_breath    (magic.c:5290) area + random worn-armor decay
//   - spell_fire_breath    (magic.c:5358) area + burn carried potions/scrolls
//   - spell_frost_breath   (magic.c:5429) area + freeze liquid containers
//   - spell_gas_breath     (magic.c:5483) area + apply AFF_POISON to survivors
//   - spell_lightning_breath (magic.c:5529) pure area damage
//   - spell_earthquake     (magic.c:3276) area damage, immune to AFF_FLYING/FLOATING
//
// Each iterates the caster's room, applying per-target breath damage
// (hp/16..hp/8, halved on save) and a flavor side effect.
package magic

import (
	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// breathDamage returns C's `number_range(hpch/16+1, hpch/8)` with halving
// on successful breath save. hpch is `UMAX(10, ch->hit)`.
func breathDamage(level int, ch, victim *types.CharData) int {
	hpch := util.UMAX(10, ch.Hit)
	dam := util.NumberRange(hpch/16+1, hpch/8)
	if SavesBreath(level, victim) {
		dam /= 2
	}
	return dam
}

// breathTargets returns room occupants eligible to receive a breath attack
// from ch, using NPC-vs-PC exclusion (C: `IS_NPC(ch) ? !IS_NPC(vch) : IS_NPC(vch)`).
// The caster is always excluded.
func breathTargets(ch *types.CharData) []*types.CharData {
	if ch == nil || ch.InRoom == nil {
		return nil
	}
	chIsNPC := ch.IsNPC()
	out := make([]*types.CharData, 0, len(ch.InRoom.People))
	for _, p := range ch.InRoom.People {
		if p == nil || p == ch {
			continue
		}
		if chIsNPC == p.IsNPC() {
			continue
		}
		out = append(out, p)
	}
	return out
}

// SpellAcidBreath — single-victim acid damage; on failed breath save AND a
// 2*level% chance, decay random carried armor AND dissolve random containers.
// Port of spell_acid_breath (C src/magic.c:5290). C takes one victim, not
// room-wide — adversary-caught pre-land.
func SpellAcidBreath(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim == nil {
		return
	}
	if util.NumberPercent() <= 2*level && !SavesBreath(level, victim) {
		acidDecay(w, victim)
	}
	dam := breathDamage(level, ch, victim)
	combat.Damage(w, ch, victim, dam, sn)
}

// acidDecay iterates the victim's inventory (like C magic.c:5300-5345) and,
// with a 1-in-4 chance per item, either pits ITEM_ARMOR (value[0]-1) or
// dissolves ITEM_CONTAINER entirely (spilling contents to the room floor).
// C continues iterating after each hit, so multiple items can be affected.
func acidDecay(w *world.World, victim *types.CharData) {
	carrying := make([]*types.ObjData, len(victim.Carrying))
	copy(carrying, victim.Carrying)
	for _, obj := range carrying {
		if obj == nil {
			continue
		}
		if util.NumberBits(2) != 0 {
			continue
		}
		switch obj.ItemType {
		case types.ITEM_ARMOR:
			if obj.Value[0] <= 0 {
				continue
			}
			obj.Value[0]--
			obj.GoldCost = 0
			obj.SilverCost = 0
			obj.CopperCost = 0
			victim.Sendf("%s is pitted and etched!\n\r", obj.ShortDescr)
		case types.ITEM_CONTAINER:
			// C magic.c:5332-5343: dissolve container, spill its contents
			// into the room.
			victim.Sendf("%s fumes and dissolves!\n\r", obj.ShortDescr)
			if victim.InRoom != nil {
				// Move contents to room before extract.
				contents := make([]*types.ObjData, len(obj.Contents))
				copy(contents, obj.Contents)
				for _, inner := range contents {
					handler.ObjFromObj(inner)
					handler.ObjToRoom(inner, victim.InRoom)
				}
			}
			handler.ExtractObj(w, obj)
		}
	}
}

// SpellFireBreath — single-victim fire damage; on failed save, burn random
// flammables (C src/magic.c:5358). Iterates all carried items per C.
func SpellFireBreath(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim == nil {
		return
	}
	if util.NumberPercent() <= 2*level && !SavesBreath(level, victim) {
		fireBurn(w, victim)
	}
	dam := breathDamage(level, ch, victim)
	combat.Damage(w, ch, victim, dam, sn)
}

// fireBurn iterates the victim's carried items and burns each eligible one
// (1-in-4 chance per C magic.c:5368-5415 `continue` loop). Multiple items
// can burn per breath.
func fireBurn(w *world.World, victim *types.CharData) {
	carrying := make([]*types.ObjData, len(victim.Carrying))
	copy(carrying, victim.Carrying)
	for _, obj := range carrying {
		if obj == nil {
			continue
		}
		if util.NumberBits(2) != 0 {
			continue
		}
		switch obj.ItemType {
		case types.ITEM_POTION:
			victim.Sendf("%s bubbles and boils!\n\r", obj.ShortDescr)
		case types.ITEM_SCROLL:
			victim.Sendf("%s crackles and burns!\n\r", obj.ShortDescr)
		case types.ITEM_STAFF:
			victim.Sendf("%s smokes and chars!\n\r", obj.ShortDescr)
		case types.ITEM_WAND:
			victim.Sendf("%s sparks and sputters!\n\r", obj.ShortDescr)
		case types.ITEM_FOOD, types.ITEM_COOK:
			victim.Sendf("%s blackens and crisps!\n\r", obj.ShortDescr)
		case types.ITEM_PILL:
			victim.Sendf("%s melts and drips!\n\r", obj.ShortDescr)
		case types.ITEM_CONTAINER:
			victim.Sendf("%s ignites and burns!\n\r", obj.ShortDescr)
		default:
			continue
		}
		handler.ExtractObj(w, obj)
	}
}

// SpellFrostBreath — single-victim cold damage; on failed save, freeze/shatter
// random liquid containers (C src/magic.c:5429). Iterates all items per C.
func SpellFrostBreath(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim == nil {
		return
	}
	if util.NumberPercent() <= 2*level && !SavesBreath(level, victim) {
		frostShatter(w, victim)
	}
	dam := breathDamage(level, ch, victim)
	combat.Damage(w, ch, victim, dam, sn)
}

// frostShatter iterates victim's carried liquid containers and shatters each
// eligible one (1-in-4 chance per item, C magic.c:5439-5470 `continue` loop).
func frostShatter(w *world.World, victim *types.CharData) {
	carrying := make([]*types.ObjData, len(victim.Carrying))
	copy(carrying, victim.Carrying)
	for _, obj := range carrying {
		if obj == nil {
			continue
		}
		if util.NumberBits(2) != 0 {
			continue
		}
		switch obj.ItemType {
		case types.ITEM_CONTAINER, types.ITEM_DRINK_CON, types.ITEM_POTION:
			victim.Sendf("%s freezes and shatters!\n\r", obj.ShortDescr)
			handler.ExtractObj(w, obj)
		}
	}
}

// SpellGasBreath — area gas damage; unlike the other breaths this is true
// area-of-effect even without an initial victim. On failed save, also apply
// AFF_POISON (ROOM_SAFE short-circuits the whole spell). Port of
// spell_gas_breath (C src/magic.c:5483).
func SpellGasBreath(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if ch.InRoom == nil {
		return
	}
	if ch.InRoom.RoomFlags.IsSet(types.ROOM_SAFE) {
		ch.Send("You fail to breathe.\n\r")
		return
	}
	for _, t := range breathTargets(ch) {
		saved := SavesBreath(level, t)
		dam := util.NumberRange(util.UMAX(10, ch.Hit)/16+1, util.UMAX(10, ch.Hit)/8)
		if saved {
			dam /= 2
		}
		if !saved && !t.AffectedBy.IsSet(types.AFF_POISON) {
			aff := &types.AffectData{
				Type:     sn,
				Duration: level / 2,
				Location: types.APPLY_STR,
				Modifier: -2,
			}
			aff.BitVector.Set(types.AFF_POISON)
			handler.AffectToChar(t, aff)
			t.Send("You feel poisoned!\n\r")
		}
		combat.Damage(w, ch, t, dam, sn)
	}
}

// SpellLightningBreath — single-victim lightning damage. C takes one victim;
// equipment-shock side effect mirrors fire/frost pattern.
// Port of spell_lightning_breath (C src/magic.c:5529).
func SpellLightningBreath(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim == nil {
		return
	}
	dam := breathDamage(level, ch, victim)
	combat.Damage(w, ch, victim, dam, sn)
}

// SpellEarthquake — area damage level + 2d8, skipping any target that is
// AFF_FLYING or AFF_FLOATING. ROOM_SAFE short-circuits. Port of
// spell_earthquake (C src/magic.c:3276).
func SpellEarthquake(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if ch.InRoom == nil {
		return
	}
	if ch.InRoom.RoomFlags.IsSet(types.ROOM_SAFE) {
		ch.Send("The earth refuses to tremble here.\n\r")
		return
	}
	ch.Send("The earth trembles beneath your feet!\n\r")
	for _, t := range breathTargets(ch) {
		if t.AffectedBy.IsSet(types.AFF_FLYING) || t.AffectedBy.IsSet(types.AFF_FLOATING) {
			continue
		}
		dam := level + util.DiceRoll(2, 8)
		combat.Damage(w, ch, t, dam, sn)
	}
}
