package persist

import (
	"fmt"
	"io"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// writeMorphData emits the #MorphData pfile block for a morphed player.
// Mirrors C fwrite_morph_data at src/polymorph.c:2405-2484. Only
// non-zero / non-default fields are written.
//
// Ordering constraint (plan §G6, C :2415-2416): Vnum MUST be emitted
// before Name. C fread_morph_data validates Name against
// morph.morph.name, but only after the Vnum line has resolved the
// morph pointer (:2543 guards on `if (morph->morph)`). Reordering
// silently disables validation.
//
// Called from SavePlayer when ch.Morph != nil. Mirrors C save.c:287-288.
func writeMorphData(w io.Writer, ch *types.CharData) error {
	if ch == nil || ch.Morph == nil {
		return nil
	}
	cm := ch.Morph
	if _, err := fmt.Fprintf(w, "#MorphData\n"); err != nil {
		return err
	}
	if cm.Morph != nil {
		if _, err := fmt.Fprintf(w, "Vnum %d\n", cm.Morph.Vnum); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "Name %s~\n", cm.Morph.Name); err != nil {
			return err
		}
	}
	if !cm.AffectedBy.IsEmpty() {
		fmt.Fprintf(w, "Affect          %s\n", cm.AffectedBy.String())
	}
	if cm.Immune != 0 {
		fmt.Fprintf(w, "Immune          %d\n", cm.Immune)
	}
	if cm.Resistant != 0 {
		fmt.Fprintf(w, "Resistant       %d\n", cm.Resistant)
	}
	if cm.Suscept != 0 {
		fmt.Fprintf(w, "Suscept         %d\n", cm.Suscept)
	}
	if !cm.NoAffectedBy.IsEmpty() {
		fmt.Fprintf(w, "NoAffect        %s\n", cm.NoAffectedBy.String())
	}
	if cm.NoImmune != 0 {
		fmt.Fprintf(w, "NoImmune        %d\n", cm.NoImmune)
	}
	if cm.NoResistant != 0 {
		fmt.Fprintf(w, "NoResistant     %d\n", cm.NoResistant)
	}
	if cm.NoSuscept != 0 {
		fmt.Fprintf(w, "NoSuscept       %d\n", cm.NoSuscept)
	}
	if cm.AC != 0 {
		fmt.Fprintf(w, "Armor           %d\n", cm.AC)
	}
	if cm.Blood != 0 {
		fmt.Fprintf(w, "Blood           %d\n", cm.Blood)
	}
	if cm.Cha != 0 {
		fmt.Fprintf(w, "Charisma        %d\n", cm.Cha)
	}
	if cm.Con != 0 {
		fmt.Fprintf(w, "Constitution    %d\n", cm.Con)
	}
	if cm.Damroll != 0 {
		fmt.Fprintf(w, "Damroll\t %d\n", cm.Damroll)
	}
	if cm.Dex != 0 {
		fmt.Fprintf(w, "Dexterity       %d\n", cm.Dex)
	}
	if cm.Dodge != 0 {
		fmt.Fprintf(w, "Dodge           %d\n", cm.Dodge)
	}
	if cm.Hit != 0 {
		fmt.Fprintf(w, "Hit             %d\n", cm.Hit)
	}
	if cm.Hitroll != 0 {
		fmt.Fprintf(w, "Hitroll         %d\n", cm.Hitroll)
	}
	if cm.Int != 0 {
		fmt.Fprintf(w, "Intelligence    %d\n", cm.Int)
	}
	if cm.Lck != 0 {
		fmt.Fprintf(w, "Luck            %d\n", cm.Lck)
	}
	if cm.Mana != 0 {
		fmt.Fprintf(w, "Mana            %d\n", cm.Mana)
	}
	if cm.Move != 0 {
		fmt.Fprintf(w, "Move            %d\n", cm.Move)
	}
	if cm.Parry != 0 {
		fmt.Fprintf(w, "Parry           %d\n", cm.Parry)
	}
	if cm.SavingBreath != 0 {
		fmt.Fprintf(w, "Save1           %d\n", cm.SavingBreath)
	}
	if cm.SavingParaPetri != 0 {
		fmt.Fprintf(w, "Save2           %d\n", cm.SavingParaPetri)
	}
	if cm.SavingPoisonDeath != 0 {
		fmt.Fprintf(w, "Save3           %d\n", cm.SavingPoisonDeath)
	}
	if cm.SavingSpellStaff != 0 {
		fmt.Fprintf(w, "Save4           %d\n", cm.SavingSpellStaff)
	}
	if cm.SavingWand != 0 {
		fmt.Fprintf(w, "Save5           %d\n", cm.SavingWand)
	}
	if cm.Str != 0 {
		fmt.Fprintf(w, "Strength        %d\n", cm.Str)
	}
	if cm.Timer != -1 {
		fmt.Fprintf(w, "Timer\t       %d\n", cm.Timer)
	}
	if cm.Tumble != 0 {
		fmt.Fprintf(w, "Tumble          %d\n", cm.Tumble)
	}
	if cm.Wis != 0 {
		fmt.Fprintf(w, "Wisdom          %d\n", cm.Wis)
	}
	_, err := fmt.Fprintf(w, "End\n")
	return err
}

// readMorphData parses a #MorphData pfile block into ch.Morph. Mirrors
// C fread_morph_data at src/polymorph.c:2487-2590. The Vnum key resolves
// the template pointer via MorphGetter; subsequent Name key validates
// (but does not re-assign) that pointer against the stored name.
//
// Defensive behavior (plan §G6): if the morph vnum no longer resolves
// (e.g. morphdestroy happened between sessions), ch.Morph is still
// allocated with Morph==nil, a warn is logged, and parsing continues.
// C blindly assumes resolution succeeds.
func readMorphData(ch *types.CharData, sc *Scanner) {
	if ch == nil {
		return
	}
	cm := &types.CharMorph{Timer: -1}
	ch.Morph = cm
	for {
		word := sc.ReadWord()
		if word == "" {
			return
		}
		if word == "End" {
			return
		}
		if !applyMorphDataField(cm, word, sc) {
			util.Bug("Fread_morph_data: no match: %s", word)
		}
	}
}

// applyMorphDataField dispatches one KV pair into the CharMorph. Returns
// true if the key was recognized. Mirrors src/polymorph.c:2504-2581.
func applyMorphDataField(cm *types.CharMorph, word string, sc *Scanner) bool {
	switch word {
	case "Affect":
		cm.AffectedBy = sc.ReadBitvector()
	case "Armor":
		cm.AC = sc.ReadNumber()
	case "Blood":
		cm.Blood = sc.ReadNumber()
	case "Charisma":
		cm.Cha = sc.ReadNumber()
	case "Constitution":
		cm.Con = sc.ReadNumber()
	case "Damroll":
		cm.Damroll = sc.ReadNumber()
	case "Dexterity":
		cm.Dex = sc.ReadNumber()
	case "Dodge":
		cm.Dodge = sc.ReadNumber()
	case "Hit":
		cm.Hit = sc.ReadNumber()
	case "Hitroll":
		cm.Hitroll = sc.ReadNumber()
	case "Immune":
		cm.Immune = sc.ReadNumber()
	case "Intelligence":
		cm.Int = sc.ReadNumber()
	case "Luck":
		cm.Lck = sc.ReadNumber()
	case "Mana":
		cm.Mana = sc.ReadNumber()
	case "Move":
		cm.Move = sc.ReadNumber()
	case "Name":
		// C validates but does not reassign — we mirror that exactly.
		stored := sc.ReadString()
		if cm.Morph != nil && cm.Morph.Name != stored {
			util.Bug("Morph Name doesn't match vnum %d (stored %q, have %q)",
				cm.Morph.Vnum, stored, cm.Morph.Name)
		}
	case "NoAffect":
		cm.NoAffectedBy = sc.ReadBitvector()
	case "NoImmune":
		cm.NoImmune = sc.ReadNumber()
	case "NoResistant":
		cm.NoResistant = sc.ReadNumber()
	case "NoSuscept":
		cm.NoSuscept = sc.ReadNumber()
	case "Parry":
		cm.Parry = sc.ReadNumber()
	case "Resistant":
		cm.Resistant = sc.ReadNumber()
	case "Save1":
		cm.SavingBreath = sc.ReadNumber()
	case "Save2":
		cm.SavingParaPetri = sc.ReadNumber()
	case "Save3":
		cm.SavingPoisonDeath = sc.ReadNumber()
	case "Save4":
		cm.SavingSpellStaff = sc.ReadNumber()
	case "Save5":
		cm.SavingWand = sc.ReadNumber()
	case "Strength":
		cm.Str = sc.ReadNumber()
	case "Suscept":
		cm.Suscept = sc.ReadNumber()
	case "Timer":
		cm.Timer = sc.ReadNumber()
	case "Tumble":
		cm.Tumble = sc.ReadNumber()
	case "Vnum":
		v := sc.ReadNumber()
		if MorphGetter != nil {
			cm.Morph = MorphGetter(v)
		}
		if cm.Morph == nil {
			util.Bug("readMorphData: unknown morph vnum %d (morph may have been destroyed between sessions)", v)
		}
	case "Wisdom":
		cm.Wis = sc.ReadNumber()
	default:
		return false
	}
	return true
}
