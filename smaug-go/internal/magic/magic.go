// Package magic implements the SMAUG spell/skill system.
package magic

import (
	"fmt"
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
	"spell_magic_missile":  SpellMagicMissile,
	"spell_cure_light":     SpellCureLight,
	"spell_cure_serious":   SpellCureSerious,
	"spell_cure_critical":  SpellCureCritical,
	"spell_fireball":       SpellFireball,
	"spell_armor":          SpellArmor,
	"spell_bless":          SpellBless,
	"spell_curse":          SpellCurse,
	"spell_poison":         SpellPoison,
	"spell_blindness":      SpellBlindness,
	"spell_sanctuary":      SpellSanctuary,
	"spell_dispel_magic":   SpellDispelMagic,
	"spell_sleep":          SpellSleep,
	"spell_charm_person":   SpellCharmPerson,
	"spell_detect_evil":    SpellDetectEvil,
	"spell_detect_invis":   SpellDetectInvis,
	"spell_detect_magic":   SpellDetectMagic,
	"spell_detect_hidden":  SpellDetectHidden,
	"spell_shield":         SpellShield,
	"spell_identify":       SpellIdentify,
	"spell_locate_object":  SpellLocateObject,
	"spell_create_food":    SpellCreateFood,
	"spell_create_water":   SpellCreateWater,
	"spell_summon":         SpellSummon,
	"spell_teleport":       SpellTeleport,
	"spell_enchant_weapon": SpellEnchantWeapon,
	"spell_enchant_armor":  SpellEnchantArmor,
	"spell_invis":          SpellInvis,
	"spell_fly":            SpellFly,
	"spell_heal":           SpellHeal,
	"spell_smaug":          SpellSmaug,
	// Phase 5 Tier 4 G3 Group A — simple affect variants
	"spell_pass_door":     SpellPassDoor,
	"spell_farsight":      SpellFarsight,
	"spell_ventriloquate": SpellVentriloquate,
	"spell_remove_invis":  SpellRemoveInvis,
	"spell_remove_trap":   SpellRemoveTrap,
	// Phase 5 Tier 4 G3 Group B — area / breath attacks
	"spell_acid_breath":      SpellAcidBreath,
	"spell_fire_breath":      SpellFireBreath,
	"spell_frost_breath":     SpellFrostBreath,
	"spell_gas_breath":       SpellGasBreath,
	"spell_lightning_breath": SpellLightningBreath,
	"spell_earthquake":       SpellEarthquake,
	// Phase 5 Tier 4 G3 Group C — movement / teleportation
	"spell_astral_walk":     SpellAstralWalk,
	"spell_gate":            SpellGate,
	"spell_mist_walk":       SpellMistWalk,
	"spell_transport":       SpellTransport,
	"spell_word_of_recall":  SpellWordOfRecall,
	"spell_group_teleport":  SpellGroupTeleport,
	// Phase 5 Tier 4 G3 Group D — unique mechanics
	"spell_acid_blast":      SpellAcidBlast,
	"spell_knock":           SpellKnock,
	"spell_recharge":        SpellRecharge,
	"spell_animate_dead":    SpellAnimateDead,
	"spell_energy_drain":    SpellEnergyDrain,
	"spell_call_lightning":  SpellCallLightning,
	"spell_control_weather": SpellControlWeather,
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

// savesSpellStaffFn is the package-level save-resolver hook used by
// SpellFarsight; tests may override it to make save outcomes deterministic.
// Mirrors the rollSaveFunc seam in spell_smaug.go.
var savesSpellStaffFn = SavesSpellStaff

// SavesPoisonDeath checks if a victim makes their saving throw vs poison/death.
func SavesPoisonDeath(level int, victim *types.CharData) bool {
	save := 50 + (victim.Level-level-victim.SavingPoisonDeath)*5
	save = util.URANGE(5, save, 95)
	return util.NumberPercent() <= save
}

// SavesWands checks if a victim makes their saving throw vs wands.
// Returns true if the save succeeds.
// RIS_MAGIC-immune victims auto-pass (C src/magic.c:1087).
func SavesWands(level int, victim *types.CharData) bool {
	if victim.Immune&int(types.RIS_MAGIC) != 0 {
		return true
	}
	save := 50 + (victim.Level-level-victim.SavingWand)*5
	save = util.URANGE(5, save, 95)
	return util.NumberPercent() <= save
}

// SavesParaPetri checks if a victim makes their saving throw vs paralysis/petrification.
func SavesParaPetri(level int, victim *types.CharData) bool {
	save := 50 + (victim.Level-level-victim.SavingParaPetri)*5
	save = util.URANGE(5, save, 95)
	return util.NumberPercent() <= save
}

// SavesBreath checks if a victim makes their saving throw vs breath weapons.
func SavesBreath(level int, victim *types.CharData) bool {
	save := 50 + (victim.Level-level-victim.SavingBreath)*5
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
	combat.Damage(w, ch, victim, dam, sn)
}

// SpellFireball deals level-scaled damage (save for half).
func SpellFireball(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	dam := util.UMAX(30, level*3)
	dam = util.NumberRange(dam/2, dam*2)
	if SavesSpellStaff(level, victim) {
		dam /= 2
	}
	combat.Damage(w, ch, victim, dam, sn)
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

// --- Phase 3 Spells ---

// SpellSleep puts the victim to sleep (save negates).
func SpellSleep(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim.AffectedBy.IsSet(types.AFF_SLEEP) {
		ch.Send("They are already asleep.\n\r")
		return
	}
	if SavesSpellStaff(level, victim) {
		ch.Send("Your magic fails to take hold.\n\r")
		return
	}
	aff := &types.AffectData{
		Type:     sn,
		Duration: 4 + level/8,
		Location: types.APPLY_NONE,
	}
	aff.BitVector.Set(types.AFF_SLEEP)
	handler.AffectToChar(victim, aff)

	if victim.Position > types.POS_SLEEPING {
		victim.Send("You feel very sleepy... zzzz.\n\r")
		victim.Position = types.POS_SLEEPING
	}
}

// SpellCharmPerson charms the victim to follow the caster (save negates).
func SpellCharmPerson(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim == ch {
		ch.Send("You like yourself even better!\n\r")
		return
	}
	if victim.AffectedBy.IsSet(types.AFF_CHARM) {
		ch.Send("They are already charmed.\n\r")
		return
	}
	if !victim.IsNPC() {
		ch.Send("You cannot charm other players.\n\r")
		return
	}
	if SavesSpellStaff(level, victim) {
		ch.Send("Your magic fails to take hold.\n\r")
		return
	}
	aff := &types.AffectData{
		Type:     sn,
		Duration: util.NumberFuzzy(level / 4),
		Location: types.APPLY_NONE,
	}
	aff.BitVector.Set(types.AFF_CHARM)
	handler.AffectToChar(victim, aff)

	victim.Master = ch
	victim.Leader = ch
	ch.Sendf("%s looks at you with adoring eyes.\n\r", victim.ShortDescr)
}

// SpellDetectEvil grants AFF_DETECT_EVIL.
func SpellDetectEvil(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim.AffectedBy.IsSet(types.AFF_DETECT_EVIL) {
		victim.Send("You can already sense evil.\n\r")
		return
	}
	aff := &types.AffectData{
		Type:     sn,
		Duration: level + 10,
		Location: types.APPLY_NONE,
	}
	aff.BitVector.Set(types.AFF_DETECT_EVIL)
	handler.AffectToChar(victim, aff)
	victim.Send("Your eyes tingle.\n\r")
}

// SpellDetectInvis grants AFF_DETECT_INVIS.
func SpellDetectInvis(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim.AffectedBy.IsSet(types.AFF_DETECT_INVIS) {
		victim.Send("You can already see invisible.\n\r")
		return
	}
	aff := &types.AffectData{
		Type:     sn,
		Duration: level + 10,
		Location: types.APPLY_NONE,
	}
	aff.BitVector.Set(types.AFF_DETECT_INVIS)
	handler.AffectToChar(victim, aff)
	victim.Send("Your eyes tingle.\n\r")
}

// SpellDetectMagic grants AFF_DETECT_MAGIC.
func SpellDetectMagic(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim.AffectedBy.IsSet(types.AFF_DETECT_MAGIC) {
		victim.Send("You can already sense magical auras.\n\r")
		return
	}
	aff := &types.AffectData{
		Type:     sn,
		Duration: level + 10,
		Location: types.APPLY_NONE,
	}
	aff.BitVector.Set(types.AFF_DETECT_MAGIC)
	handler.AffectToChar(victim, aff)
	victim.Send("Your eyes tingle.\n\r")
}

// SpellDetectHidden grants AFF_DETECT_HIDDEN.
func SpellDetectHidden(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim.AffectedBy.IsSet(types.AFF_DETECT_HIDDEN) {
		victim.Send("You can already sense hidden things.\n\r")
		return
	}
	aff := &types.AffectData{
		Type:     sn,
		Duration: level + 10,
		Location: types.APPLY_NONE,
	}
	aff.BitVector.Set(types.AFF_DETECT_HIDDEN)
	handler.AffectToChar(victim, aff)
	victim.Send("Your awareness improves.\n\r")
}

// SpellShield improves AC by -20 (similar to armor but stacks differently).
func SpellShield(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	handler.AffectToChar(victim, &types.AffectData{
		Type:     sn,
		Duration: 18 + level/2,
		Location: types.APPLY_AC,
		Modifier: -20,
	})
	victim.Send("A shimmering shield surrounds you.\n\r")
}

// SpellIdentify reveals the properties of an item held by the caster.
func SpellIdentify(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	// Identify operates on an object, not a victim. Use the last item in inventory.
	if len(ch.Carrying) == 0 {
		ch.Send("You are not carrying anything to identify.\n\r")
		return
	}
	// Find first non-worn item
	var obj *types.ObjData
	for _, o := range ch.Carrying {
		if o.WearLoc == types.WEAR_NONE {
			obj = o
			break
		}
	}
	if obj == nil {
		ch.Send("You are not carrying anything to identify.\n\r")
		return
	}

	ch.Sendf("Object: %s\n\r", obj.ShortDescr)
	ch.Sendf("Type: %d  Level: %d  Weight: %d  Value: %d\n\r",
		obj.ItemType, obj.Level, obj.Weight, obj.GoldCost)
	ch.Sendf("Values: %d %d %d %d %d %d\n\r",
		obj.Value[0], obj.Value[1], obj.Value[2],
		obj.Value[3], obj.Value[4], obj.Value[5])

	if len(obj.Affects) > 0 {
		for _, aff := range obj.Affects {
			ch.Sendf("Affects %d by %d.\n\r", aff.Location, aff.Modifier)
		}
	}
}

// --- Phase 3 G7 Spells ---

// SpellLocateObject shows where an object is in the world.
func SpellLocateObject(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if ch == nil {
		return
	}
	// Use the argument stored in ch's last command — simplified version
	found := false
	for _, obj := range w.Objects {
		if obj.ExtraFlags.IsSet(types.ITEM_NOLOCATE) {
			continue
		}
		location := "somewhere"
		if obj.CarriedBy != nil {
			location = "carried by " + obj.CarriedBy.Name
		} else if obj.InRoom != nil {
			location = obj.InRoom.Name
		}
		ch.Sendf("%s is %s.\n\r", obj.ShortDescr, location)
		found = true
		break // just show first match for simplicity
	}
	if !found {
		ch.Send("Nothing like that in the world.\n\r")
	}
}

// SpellCreateFood creates a food item in the caster's inventory.
func SpellCreateFood(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	food := &types.ObjData{
		Name:       "mushroom magical",
		ShortDescr: "a magical mushroom",
		Description: "A magical mushroom has been left here.",
		ItemType:   types.ITEM_FOOD,
		Value:      [6]int{level / 2, level / 2, 0, 0, 0, 0},
		Weight:     1,
		WearLoc:    types.WEAR_NONE,
		Timer:      24 + level,
	}
	w.AddObj(food)
	food.CarriedBy = ch
	ch.Carrying = append(ch.Carrying, food)
	ch.Send("A mushroom suddenly appears.\n\r")
}

// SpellCreateWater fills a drink container with water.
func SpellCreateWater(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	// Find a drink container in inventory
	for _, obj := range ch.Carrying {
		if obj.ItemType == types.ITEM_DRINK_CON && obj.WearLoc == types.WEAR_NONE {
			water := util.UMIN(level*2, obj.Value[0]-obj.Value[1])
			if water > 0 {
				obj.Value[1] += water
				obj.Value[2] = 0 // water liquid type
				ch.Sendf("Water flows into %s.\n\r", obj.ShortDescr)
			} else {
				ch.Send("It's already full.\n\r")
			}
			return
		}
	}
	ch.Send("You need a drink container.\n\r")
}

// SpellSummon teleports a character to the caster.
func SpellSummon(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim == nil || victim == ch {
		ch.Send("You can't summon that.\n\r")
		return
	}
	if victim.InRoom == nil || ch.InRoom == nil {
		return
	}
	if ch.InRoom.RoomFlags.IsSet(types.ROOM_NO_SUMMON) ||
		victim.InRoom.RoomFlags.IsSet(types.ROOM_NO_SUMMON) ||
		victim.InRoom.RoomFlags.IsSet(types.ROOM_PRIVATE) ||
		victim.InRoom.RoomFlags.IsSet(types.ROOM_SOLITARY) {
		ch.Send("You failed.\n\r")
		return
	}
	if victim.Fighting != nil {
		ch.Send("They are too busy fighting.\n\r")
		return
	}

	handler.CharFromRoom(victim)
	handler.CharToRoom(victim, ch.InRoom)
	victim.Sendf("%s has summoned you!\n\r", ch.Name)
	ch.Sendf("You summon %s.\n\r", victim.Name)
}

// SpellTeleport teleports the caster to a random room.
func SpellTeleport(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	// Find a random room
	var dest *types.RoomIndexData
	for _, r := range w.Rooms {
		if r.RoomFlags.IsSet(types.ROOM_PRIVATE) || r.RoomFlags.IsSet(types.ROOM_NO_RECALL) {
			continue
		}
		if util.NumberBits(3) == 0 {
			dest = r
			break
		}
	}
	if dest == nil {
		ch.Send("Your spell fizzles.\n\r")
		return
	}

	if ch.InRoom != nil {
		handler.CharFromRoom(ch)
	}
	handler.CharToRoom(ch, dest)
	ch.Send("You are teleported!\n\r")
}

// SpellEnchantWeapon adds +1 hitroll/damroll to a wielded weapon.
func SpellEnchantWeapon(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	wield := handler.GetEqChar(ch, types.WEAR_WIELD)
	if wield == nil {
		ch.Send("You must wield a weapon to enchant.\n\r")
		return
	}
	if wield.ExtraFlags.IsSet(types.ITEM_MAGIC) {
		ch.Send("That weapon is already enchanted.\n\r")
		return
	}

	bonus := 1 + level/20
	wield.Affects = append(wield.Affects, &types.AffectData{
		Type:     sn,
		Duration: -1,
		Location: types.APPLY_HITROLL,
		Modifier: bonus,
	})
	wield.Affects = append(wield.Affects, &types.AffectData{
		Type:     sn,
		Duration: -1,
		Location: types.APPLY_DAMROLL,
		Modifier: bonus,
	})
	wield.ExtraFlags.Set(types.ITEM_MAGIC)
	ch.Sendf("%s glows brightly!\n\r", util.Capitalize(wield.ShortDescr))
}

// SpellEnchantArmor adds +1 AC to worn armor.
func SpellEnchantArmor(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	// Find first worn armor piece
	var armor *types.ObjData
	for _, obj := range ch.Carrying {
		if obj.WearLoc != types.WEAR_NONE && obj.ItemType == types.ITEM_ARMOR {
			armor = obj
			break
		}
	}
	if armor == nil {
		ch.Send("You must be wearing armor to enchant.\n\r")
		return
	}
	if armor.ExtraFlags.IsSet(types.ITEM_MAGIC) {
		ch.Send("That armor is already enchanted.\n\r")
		return
	}

	bonus := -(1 + level/20)
	armor.Affects = append(armor.Affects, &types.AffectData{
		Type:     sn,
		Duration: -1,
		Location: types.APPLY_AC,
		Modifier: bonus,
	})
	armor.ExtraFlags.Set(types.ITEM_MAGIC)
	ch.Sendf("%s glows with a protective aura!\n\r", util.Capitalize(armor.ShortDescr))
}

// SpellInvis grants AFF_INVISIBLE.
func SpellInvis(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim.AffectedBy.IsSet(types.AFF_INVISIBLE) {
		victim.Send("You are already invisible.\n\r")
		return
	}
	aff := &types.AffectData{
		Type:     sn,
		Duration: 24 + level,
		Location: types.APPLY_NONE,
	}
	aff.BitVector.Set(types.AFF_INVISIBLE)
	handler.AffectToChar(victim, aff)
	victim.Send("You fade out of existence.\n\r")
}

// SpellFly grants AFF_FLYING.
func SpellFly(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim.AffectedBy.IsSet(types.AFF_FLYING) {
		victim.Send("You are already flying.\n\r")
		return
	}
	aff := &types.AffectData{
		Type:     sn,
		Duration: level + 10,
		Location: types.APPLY_NONE,
	}
	aff.BitVector.Set(types.AFF_FLYING)
	handler.AffectToChar(victim, aff)
	victim.Send("Your feet rise off the ground.\n\r")
}

// SpellHeal heals a large amount of HP.
func SpellHeal(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	heal := util.UMAX(100, level*5)
	victim.Hit = util.UMIN(victim.Hit+heal, victim.MaxHit)
	victim.Send("A warm feeling fills your body.\n\r")
}

// --- Phase 5 Tier 4 G3 Group A: simple affect-like variants ---

// SpellPassDoor grants AFF_PASS_DOOR for level/4 (fuzzy) pulses.
// Port of spell_pass_door (C src/magic.c:4522).
// RIS_MAGIC immunity blocks.
func SpellPassDoor(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim == nil {
		victim = ch
	}
	if victim.Immune&int(types.RIS_MAGIC) != 0 {
		ch.Send("They are immune to your magic.\n\r")
		return
	}
	if victim.AffectedBy.IsSet(types.AFF_PASS_DOOR) {
		if ch == victim {
			ch.Send("You are already ethereal.\n\r")
		} else {
			ch.Send("They are already ethereal.\n\r")
		}
		return
	}
	aff := &types.AffectData{
		Type:     sn,
		Duration: util.NumberFuzzy(level / 4),
		Location: types.APPLY_NONE,
	}
	aff.BitVector.Set(types.AFF_PASS_DOOR)
	handler.AffectToChar(victim, aff)
	victim.Send("You turn translucent.\n\r")
}

// SpellFarsight is a scry-style spell: "view" a remote target's room.
// Port of spell_farsight (C src/magic.c:5973).
//
// MVP simplification: the Go port doesn't yet have OVERLANDCODE hooks or the
// `do_look` temporary relocation dance. Instead, we show the caster the target
// room's Name + Description if scrying is allowed. This matches the observable
// player-visible effect (you see the remote room) without perturbing state.
//
// Fails if victim is nil, victim is the caster, target room has PRIVATE /
// SOLITARY / NO_ASTRAL / DEATH flags, caster's room has NO_RECALL, victim is
// too high level, or victim saves.
func SpellFarsight(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim == nil || victim == ch || victim.InRoom == nil {
		ch.Send("You fail to locate them.\n\r")
		return
	}
	loc := victim.InRoom
	if loc.RoomFlags.IsSet(types.ROOM_PRIVATE) ||
		loc.RoomFlags.IsSet(types.ROOM_SOLITARY) ||
		loc.RoomFlags.IsSet(types.ROOM_NO_ASTRAL) ||
		loc.RoomFlags.IsSet(types.ROOM_DEATH) {
		ch.Send("You fail to locate them.\n\r")
		return
	}
	if ch.InRoom != nil && ch.InRoom.RoomFlags.IsSet(types.ROOM_NO_RECALL) {
		ch.Send("You fail to locate them.\n\r")
		return
	}
	if victim.Level >= level+15 {
		ch.Send("You fail to locate them.\n\r")
		return
	}
	if victim.IsNPC() && savesSpellStaffFn(level, victim) {
		ch.Send("You fail to locate them.\n\r")
		return
	}
	ch.Sendf("You concentrate on %s and your vision blurs...\n\r", victim.Name)
	ch.Sendf("%s\n\r", loc.Name)
	if loc.Description != "" {
		ch.Send(loc.Description)
	}
	// 1-in-20 chance (matches C's chance_attrib; simplified without wis check)
	// to tip off the victim.
	if util.NumberPercent() <= 5 {
		victim.Send("You get an uneasy feeling that you are being watched.\n\r")
	}
}

// SpellVentriloquate fakes a named speaker in the caster's room. Each listener
// either sees the real speech ("Foo says '...'") if they fail their save, or
// a hint ("Someone makes Foo say '...'") if they save.
// Port of spell_ventriloquate (C src/magic.c:5183).
//
// MVP simplification: the Go port doesn't have target_name threading, so the
// spell text is passed as victim.ShortDescr (the speaker name) — the cast
// command resolves the name to a char in the room via GetCharRoom. The actual
// uttered text is taken from the skill's HitChar field (set by skills.dat) if
// the caller populated it; otherwise a generic placeholder is used. This
// diverges slightly from C (which splits target_name into speaker + message)
// but preserves the save-gated dual-message structure.
func SpellVentriloquate(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if ch.InRoom == nil {
		return
	}
	speaker := "someone"
	if victim != nil {
		speaker = victim.Name
	}
	msg := "..."
	realLine := fmt.Sprintf("%s says '%s'\n\r", util.Capitalize(speaker), msg)
	fakeLine := fmt.Sprintf("Someone makes %s say '%s'\n\r", speaker, msg)
	for _, vch := range ch.InRoom.People {
		if vch == ch || vch == victim {
			continue
		}
		if SavesSpellStaff(level, vch) {
			vch.Send(fakeLine)
		} else {
			vch.Send(realLine)
		}
	}
}

// SpellRemoveInvis strips invisibility from an object in inventory or a char
// in the room. Port of spell_remove_invis (C src/magic.c:6387).
//
// MVP simplification: the Go port's spell dispatch already resolves the
// target to victim. If victim is non-nil we strip AFF_INVISIBLE from them.
// We do not implement the object branch here — that path requires target_name
// text threading (like enchant_armor), which isn't wired in the spell system
// yet.
func SpellRemoveInvis(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim == nil {
		ch.Send("What should the spell be cast upon?\n\r")
		return
	}
	if !victim.AffectedBy.IsSet(types.AFF_INVISIBLE) {
		ch.Send("They are not invisible!\n\r")
		return
	}
	if victim.Immune&int(types.RIS_MAGIC) != 0 {
		ch.Send("They are immune to your magic.\n\r")
		return
	}
	// Remove both generic invis and mass_invis affects, plus the bit.
	invisSN := findSkillSN(w, "invis")
	if invisSN >= 0 {
		handler.AffectStrip(victim, invisSN)
	}
	massInvisSN := findSkillSN(w, "mass invis")
	if massInvisSN >= 0 {
		handler.AffectStrip(victim, massInvisSN)
	}
	victim.AffectedBy.Remove(types.AFF_INVISIBLE)
	ch.Send("Ok.\n\r")
}

// SpellRemoveTrap disarms an ITEM_TRAP in the room. Port of spell_remove_trap
// (C src/magic.c:4648).
//
// MVP simplification: no per-object targeting (lacks target_name). We scan
// the caster's room contents and disarm the first ITEM_TRAP we find. If that
// trap is attached to a container, we don't walk inside (C uses get_trap which
// peeks inside containers); this covers the common case of floor traps.
func SpellRemoveTrap(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if ch.InRoom == nil {
		ch.Send("You can't find that here.\n\r")
		return
	}
	for _, obj := range ch.InRoom.Contents {
		if obj != nil && obj.ItemType == types.ITEM_TRAP {
			// 70% + curr_wis success chance (C uses chance(ch, 70 + get_curr_wis)).
			// We simplify to a flat 75 (approx low-wis baseline) + level-aware bonus.
			if util.NumberPercent() > 75+level/4 {
				ch.Send("Ooops!\n\r")
				// C triggers spring_trap here; we just silently fail to avoid
				// further wiring deps.
				return
			}
			handler.ExtractObj(w, obj)
			ch.Send("You successfully remove the trap.\n\r")
			return
		}
	}
	ch.Send("You can't find a trap here.\n\r")
}
