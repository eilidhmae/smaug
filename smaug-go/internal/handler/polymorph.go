package handler

// Polymorph primitives — the stat-application path for the morph subsystem.
//
// This file intentionally lives in `handler` (not `act` or `polymorph`) so
// both the command layer (internal/act/polymorph.go) and the mudprog
// layer (internal/mudprog/commands.go mpMorph/mpUnmorph) can reach these
// helpers without creating an import cycle. See plan-phase6-polymorph.md
// §4.7.
//
// The apply path (DoMorph) mutates CharData fields DIRECTLY, matching C
// do_morph at src/polymorph.c:1473-1531. Not routed through AffectModify:
// C stores the exact applied delta on ch.Morph (CharMorph) so DoUnmorph
// can subtract it back out. Reimplementing this through ~25 synthetic
// AffectData entries would diverge from C semantics (see §4.2 of the plan).

import (
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// MaxMorphVital is the signed-short clamp the C code applies to
// hit/mana/move ceilings at src/polymorph.c:1501-1506, :1516-1517.
// 32700 keeps the value below SH_INT_MAX and leaves headroom for
// transient arithmetic.
const MaxMorphVital = 32700

// MakeCharMorph allocates a fresh CharMorph bound to the given MorphData.
// Mirrors C make_char_morph at src/polymorph.c:2269-2307. Only the
// "static" copy-from-morph fields are populated here; the "delta" fields
// (Hit, Damroll, Hitroll, Mana, Move, Blood) are left zero — DoMorph
// fills them after evaluating the morph's dice-string expressions at
// apply time.
func MakeCharMorph(m *types.MorphData) *types.CharMorph {
	if m == nil {
		return nil
	}
	return &types.CharMorph{
		Morph:             m,
		AC:                m.AC,
		Str:               m.Str,
		Int:               m.Int,
		Wis:               m.Wis,
		Dex:               m.Dex,
		Cha:               m.Cha,
		Lck:               m.Lck,
		SavingBreath:      m.SavingBreath,
		SavingParaPetri:   m.SavingParaPetri,
		SavingPoisonDeath: m.SavingPoisonDeath,
		SavingSpellStaff:  m.SavingSpellStaff,
		SavingWand:        m.SavingWand,
		Timer:             m.Timer,
		AffectedBy:        m.AffectedBy,
		Immune:            m.Immune,
		Resistant:         m.Resistant,
		Suscept:           m.Suscept,
		NoAffectedBy:      m.NoAffectedBy,
		NoImmune:          m.NoImmune,
		NoResistant:       m.NoResistant,
		NoSuscept:         m.NoSuscept,
	}
}

// ClearCharMorph zeroes out all numeric + bit fields on a CharMorph and
// nils the template pointer. Mirrors C clear_char_morph at
// src/polymorph.c:2593-2627.
func ClearCharMorph(cm *types.CharMorph) {
	if cm == nil {
		return
	}
	*cm = types.CharMorph{Timer: -1}
}

// isVampirePC returns true for living PC vampires (class or race). Mirrors
// C IS_VAMPIRE(ch) semantics as used at polymorph.c:1335/1508/1567 —
// the check gates the bloodthirst path instead of mana.
func isVampirePC(ch *types.CharData) bool {
	if ch == nil || ch.IsNPC() {
		return false
	}
	return ch.Race == types.RACE_VAMPIRE || ch.Class == types.CLASS_VAMPIRE
}

// DoMorph applies a morph's stat overrides to ch. Mirrors C do_morph at
// src/polymorph.c:1473-1531. If ch already has an active morph, it is
// removed first (C behavior at :1480-1481).
//
// The direct-mutation pattern is deliberate (plan §4.2). Every delta is
// stored on ch.Morph (CharMorph) so DoUnmorph can subtract it back out
// exactly.
func DoMorph(ch *types.CharData, m *types.MorphData) {
	if ch == nil || m == nil {
		return
	}
	if ch.Morph != nil {
		DoUnmorphChar(ch)
	}
	cm := MakeCharMorph(m)

	ch.Armor += m.AC
	ch.ModStr += m.Str
	ch.ModInt += m.Int
	ch.ModWis += m.Wis
	ch.ModDex += m.Dex
	ch.ModCha += m.Cha
	ch.ModLck += m.Lck

	ch.SavingBreath += m.SavingBreath
	ch.SavingParaPetri += m.SavingParaPetri
	ch.SavingPoisonDeath += m.SavingPoisonDeath
	ch.SavingSpellStaff += m.SavingSpellStaff
	ch.SavingWand += m.SavingWand

	cm.Hitroll = util.DiceParse(m.Hitroll, m.Level)
	ch.Hitroll += cm.Hitroll
	cm.Damroll = util.DiceParse(m.Damroll, m.Level)
	ch.Damroll += cm.Damroll

	cm.Hit = util.DiceParse(m.Hit, m.Level)
	if ch.Hit+cm.Hit > MaxMorphVital {
		cm.Hit = MaxMorphVital - ch.Hit
	}
	ch.Hit += cm.Hit

	cm.Move = util.DiceParse(m.Move, m.Level)
	if ch.Move+cm.Move > MaxMorphVital {
		cm.Move = MaxMorphVital - ch.Move
	}
	ch.Move += cm.Move

	if isVampirePC(ch) && ch.PCData != nil {
		cm.Blood = util.DiceParse(m.Blood, m.Level)
		ch.PCData.Condition[types.COND_BLOODTHIRST] += cm.Blood
	} else {
		cm.Mana = util.DiceParse(m.Mana, m.Level)
		if ch.Mana+cm.Mana > MaxMorphVital {
			cm.Mana = MaxMorphVital - ch.Mana
		}
		ch.Mana += cm.Mana
	}

	// Bitvector affect OR / scalar mask OR.
	ch.AffectedBy = ch.AffectedBy.Or(m.AffectedBy)
	ch.Immune |= m.Immune
	ch.Resistant |= m.Resistant
	ch.Susceptible |= m.Suscept
	// Scalar "no_*" mask removal.
	ch.AffectedBy = ch.AffectedBy.AndNot(m.NoAffectedBy)
	ch.Immune &^= m.NoImmune
	ch.Resistant &^= m.NoResistant
	ch.Susceptible &^= m.NoSuscept

	ch.Morph = cm
	m.Used++
}

// DoUnmorph reverses every delta DoMorph applied and detaches ch.Morph.
// Mirrors C do_unmorph at src/polymorph.c:1545-1579.
//
// Go-specific note: C's do_unmorph calls update_aris(ch) as the last
// step. The Go port intentionally omits this — `internal/persist/
// player_affect_test.go:11-22` documents that Go's invariant is
// incremental direct-mutation via AffectModify on each affect op.
// Because DoMorph directly mutates the base fields (not AffectData
// synthesis), DoUnmorph's subtraction is sufficient to restore state.
func DoUnmorph(ch *types.CharData) {
	if ch == nil || ch.Morph == nil {
		return
	}
	cm := ch.Morph
	ch.Armor -= cm.AC
	ch.ModStr -= cm.Str
	ch.ModInt -= cm.Int
	ch.ModWis -= cm.Wis
	ch.ModDex -= cm.Dex
	ch.ModCha -= cm.Cha
	ch.ModLck -= cm.Lck

	ch.SavingBreath -= cm.SavingBreath
	ch.SavingParaPetri -= cm.SavingParaPetri
	ch.SavingPoisonDeath -= cm.SavingPoisonDeath
	ch.SavingSpellStaff -= cm.SavingSpellStaff
	ch.SavingWand -= cm.SavingWand

	ch.Hitroll -= cm.Hitroll
	ch.Damroll -= cm.Damroll
	ch.Hit -= cm.Hit
	ch.Move -= cm.Move
	if isVampirePC(ch) && ch.PCData != nil {
		ch.PCData.Condition[types.COND_BLOODTHIRST] -= cm.Blood
	} else {
		ch.Mana -= cm.Mana
	}

	ch.AffectedBy = ch.AffectedBy.AndNot(cm.AffectedBy)
	ch.Immune &^= cm.Immune
	ch.Resistant &^= cm.Resistant
	ch.Susceptible &^= cm.Suscept

	ch.Morph = nil
}

// SendMorphMessage emits the AT_MORPH-colored act line for morph/unmorph
// events — to the room (MorphOther / UnmorphOther) and to the char
// themself (MorphSelf / UnmorphSelf). Mirrors C send_morph_message at
// src/polymorph.c:1402-1418.
//
// Go divergence: C calls act() with potentially-empty strings; util.Act
// returns early when format == "", so empty-string morph messages are
// silently dropped (not crashed). Matches user-observed behavior on a
// fresh / partially-configured morph.
func SendMorphMessage(ch *types.CharData, m *types.MorphData, isMorph bool) {
	if ch == nil || m == nil {
		return
	}
	if isMorph {
		util.Act(types.AT_MORPH, m.MorphOther, ch, nil, nil, nil, types.TO_ROOM)
		util.Act(types.AT_MORPH, m.MorphSelf, ch, nil, nil, nil, types.TO_CHAR)
	} else {
		util.Act(types.AT_MORPH, m.UnmorphOther, ch, nil, nil, nil, types.TO_ROOM)
		util.Act(types.AT_MORPH, m.UnmorphSelf, ch, nil, nil, nil, types.TO_CHAR)
	}
}

// CanMorph is the prerequisite gate for player-initiated morphs. Mirrors
// C can_morph at src/polymorph.c:1427-1459. Immortals and NPCs can always
// morph (C :1432-1434). isCast=true is the spell_polymorph path (deferred
// per plan §7); callers today pass isCast=false.
func CanMorph(ch *types.CharData, m *types.MorphData, isCast bool) bool {
	if m == nil {
		return false
	}
	if ch == nil {
		return false
	}
	if ch.IsImmortal() || ch.IsNPC() {
		return true
	}
	if m.NoCast && isCast {
		return false
	}
	if ch.Level < m.Level {
		return false
	}
	if m.PKill == types.ONLY_PKILL && !isPkill(ch) {
		return false
	}
	if m.PKill == types.ONLY_PEACEFULL && isPkill(ch) {
		return false
	}
	if m.Sex != -1 && m.Sex != ch.Sex {
		return false
	}
	// Class: morph.Class is a bitmask of allowed classes. If set and the
	// player's class bit is NOT set, deny.
	if m.Class != 0 && m.Class&(1<<uint(ch.Class)) == 0 {
		return false
	}
	// Race: morph.Race is a bitmask of DISALLOWED races (C inverts the
	// check at :1447). If set and the player's race bit IS set, deny.
	if m.Race != 0 && m.Race&(1<<uint(ch.Race)) != 0 {
		return false
	}
	if m.Deity != "" {
		if ch.PCData == nil || ch.PCData.Deity == nil {
			return false
		}
	}
	return true
}

// isPkill returns whether the character is flagged for PvP. Matches C's
// IS_PKILL(ch) at mud.h; checks PLR_PKILL on the PCData flags.
func isPkill(ch *types.CharData) bool {
	if ch == nil || ch.IsNPC() || ch.PCData == nil {
		return false
	}
	return uint32(ch.PCData.Flags)&types.PCFLAG_DEADLY != 0
}

// DoMorphChar is the resource-gated wrapper around DoMorph. Mirrors C
// do_morph_char at src/polymorph.c:1262-1384. Returns true on success.
//
// C-fidelity decision (readiness-vet): resources are consumed as each
// check runs, NOT atomically after all checks pass (C :1319-1373). If
// hpused succeeds but gloryused fails, hp deduction stays and any
// already-extracted objuse[0] items are NOT refunded. Ported verbatim
// — do not refactor to atomic reservation.
func DoMorphChar(w *world.World, ch *types.CharData, m *types.MorphData) bool {
	if ch == nil || m == nil {
		return false
	}
	canmorph := true
	if ch.Morph != nil {
		canmorph = false
	}
	for i := 0; i < 3; i++ {
		if m.Obj[i] == 0 {
			continue
		}
		obj := GetObjVnumCarry(ch, m.Obj[i])
		if obj == nil {
			canmorph = false
			continue
		}
		if !m.ObjUse[i] {
			continue
		}
		util.Act(types.AT_OBJECT,
			"$p disappears in a whisp of smoke!",
			ch, nil, obj, nil, types.TO_CHAR)
		// If the consumed obj was the wield weapon and a dual-wield
		// exists, promote the dual to primary. Mirrors C :1278-1281.
		if obj.WearLoc == types.WEAR_WIELD {
			if dual := GetEqChar(ch, types.WEAR_DUAL_WIELD); dual != nil {
				dual.WearLoc = types.WEAR_WIELD
			}
		}
		SeparateObj(obj)
		ExtractObj(w, obj)
	}

	if m.HpUsed != 0 {
		if ch.Hit < m.HpUsed {
			canmorph = false
		} else {
			ch.Hit -= m.HpUsed
		}
	}
	if m.MoveUsed != 0 {
		if ch.Move < m.MoveUsed {
			canmorph = false
		} else {
			ch.Move -= m.MoveUsed
		}
	}
	if m.ManaUsed != 0 && !isVampirePC(ch) {
		if ch.Mana < m.ManaUsed {
			canmorph = false
		} else {
			ch.Mana -= m.ManaUsed
		}
	}
	if m.BloodUsed != 0 && isVampirePC(ch) && ch.PCData != nil {
		if ch.PCData.Condition[types.COND_BLOODTHIRST] < m.BloodUsed {
			canmorph = false
		} else {
			ch.PCData.Condition[types.COND_BLOODTHIRST] -= m.BloodUsed
		}
	}
	if m.FavourUsed != 0 {
		if ch.IsNPC() || ch.PCData == nil || ch.PCData.Deity == nil ||
			ch.PCData.Favor < m.FavourUsed {
			canmorph = false
		} else {
			ch.PCData.Favor -= m.FavourUsed
			// C :1359-1362: if favor drops below deity.susceptnum,
			// OR the deity's suscept flag onto ch.Susceptible. This
			// is an intentional penalty for low-favor morphing.
			if ch.PCData.Favor < ch.PCData.Deity.SusceptNum {
				ch.Susceptible |= ch.PCData.Deity.Suscept
			}
		}
	}
	if m.GloryUsed != 0 {
		if !ch.IsNPC() && ch.PCData != nil {
			if ch.PCData.QuestCurr < m.GloryUsed {
				canmorph = false
			} else {
				ch.PCData.QuestCurr -= m.GloryUsed
			}
		}
	}

	if !canmorph {
		ch.Send("You begin to transform, but something goes wrong.\n\r")
		return false
	}
	SendMorphMessage(ch, m, true)
	DoMorph(ch, m)
	return true
}

// DoUnmorphChar unmorphs ch and emits the unmorph message. Mirrors C
// do_unmorph_char at src/polymorph.c:1387-1396.
func DoUnmorphChar(ch *types.CharData) {
	if ch == nil || ch.Morph == nil {
		return
	}
	template := ch.Morph.Morph
	DoUnmorph(ch)
	SendMorphMessage(ch, template, false)
}

// UnmorphAll force-unmorphs every PC currently wearing the given morph.
// Called before DoMorphdestroy removes a morph from the world table, so
// no dangling ch.Morph.Morph pointer exists after the splice. Mirrors C
// unmorph_all at src/polymorph.c:2654-2669.
func UnmorphAll(w *world.World, m *types.MorphData) {
	if w == nil || m == nil {
		return
	}
	for _, vch := range w.Characters {
		if vch == nil || vch.IsNPC() {
			continue
		}
		if vch.Morph == nil || vch.Morph.Morph != m {
			continue
		}
		DoUnmorphChar(vch)
	}
}

// GetObjVnumCarry walks a character's inventory for the first object
// whose IndexData.Vnum equals the requested vnum. Mirrors C get_obj_vnum
// at src/handler.c (invoked from polymorph.c:1272).
func GetObjVnumCarry(ch *types.CharData, vnum int) *types.ObjData {
	if ch == nil || vnum <= 0 {
		return nil
	}
	for _, obj := range ch.Carrying {
		if obj == nil || obj.IndexData == nil {
			continue
		}
		if obj.IndexData.Vnum == vnum {
			return obj
		}
	}
	return nil
}

// SeparateObj is the Go stand-in for C's separate_obj (handler.c), which
// peels one item off a stack of identical objects so it can be extracted
// alone. Go has no object stacking today (each OBJ_DATA is a distinct
// instance), so SeparateObj is a no-op. Kept as a named call site so
// future stacking work can be tracked by grep. Shared with the archery
// lineage — same treatment at internal/act/archery.go:101.
func SeparateObj(obj *types.ObjData) {
	_ = obj
}
