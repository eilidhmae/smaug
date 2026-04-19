package persist

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// ObjIndexLookup is a function that resolves a vnum to an object template.
type ObjIndexLookup func(vnum int) *types.ObjIndexData

// SkillNameLookup, if non-nil, resolves a skill/spell/weapon/tongue name
// to its gsn (index into ch.PCData.Learned). Returns -1 when not found.
// Set at boot after skills are loaded; left nil in standalone unit tests.
var SkillNameLookup func(name string) int

// SkillGetter, if non-nil, returns the SkillType for a given gsn, or nil
// when the gsn is out of range. Set at boot after skills are loaded.
var SkillGetter func(gsn int) *types.SkillType

// MorphGetter, if non-nil, resolves a morph vnum to its MorphData pointer.
// Set at boot after morphs are loaded; consumed by readMorphData so the
// pfile's #MorphData block can re-attach ch.Morph.Morph. Returns nil when
// the morph does not exist (e.g. vnum was deleted between sessions — Go
// degrades gracefully per plan §G6 defensive path).
var MorphGetter func(vnum int) *types.MorphData

// LoadPlayer reads a player character from a SMAUG player save file.
// Objects in the file are skipped (use LoadPlayerWithWorld to load objects).
func LoadPlayer(r io.Reader, filename string) (*types.CharData, error) {
	return LoadPlayerWithWorld(r, filename, nil)
}

// LoadPlayerWithWorld reads a player character, resolving objects via the lookup function.
// If lookup is nil, objects are skipped.
func LoadPlayerWithWorld(r io.Reader, filename string, lookup ObjIndexLookup) (*types.CharData, error) {
	sc := NewScanner(r, filename)

	word := sc.ReadWord()
	if word != "#PLAYER" {
		return nil, fmt.Errorf("LoadPlayer: expected #PLAYER, got %q", word)
	}

	ch := &types.CharData{
		PCData: &types.PCData{
			PagerLen: 24,
		},
	}

	// For nesting objects in containers (like C's rgObjNest)
	const maxNest = 100
	var nestObj [maxNest]*types.ObjData

	playerDone := false
	for {
		word = sc.ReadWord()
		if word == "" {
			break
		}

		if word == "End" && !playerDone {
			playerDone = true
			// Do NOT early-return on lookup==nil: the pfile may still
			// carry a trailing #MorphData block (plan-phase6-polymorph
			// §G6) that binds to ch.Morph rather than ch.Carrying.
			// Object sections (#OBJECT / #CORPSE) handle the nil case
			// via skipPlayerObject.
			continue
		}

		switch word {
		case "#OBJECT", "#CORPSE":
			if lookup == nil {
				skipPlayerObject(sc)
			} else {
				obj := readPlayerObject(sc, lookup, nestObj[:])
				if obj != nil && obj.InObj == nil {
					ch.Carrying = append(ch.Carrying, obj)
					obj.CarriedBy = ch
				}
			}
		case "#MorphData":
			// Morph subsystem — plan-phase6-polymorph.md §G6. Parse
			// the attached CharMorph block, re-linking ch.Morph.Morph
			// via MorphGetter. C fread_morph_data at
			// src/polymorph.c:2487-2590.
			readMorphData(ch, sc)
		case "Affect":
			if !playerDone {
				aff := readAffect(sc)
				if aff != nil {
					ch.Affects = append(ch.Affects, aff)
				}
			} else {
				sc.ReadToEOL()
			}
		case "AffectData":
			sc.ReadToEOL()
		default:
			if !playerDone {
				parsePlayerField(ch, word, sc)
			}
		}
	}

	return ch, nil
}

// parsePlayerField handles a single keyword-value pair for the #PLAYER section.
func parsePlayerField(ch *types.CharData, word string, sc *Scanner) {
	switch word {
	case "Version":
		_ = sc.ReadNumber()
	case "Name":
		ch.Name = sc.ReadString()
		ch.ShortDescr = ch.Name
	case "Description":
		ch.Description = sc.ReadString()
	case "Sex":
		ch.Sex = sc.ReadNumber()
	case "Class":
		ch.Class = sc.ReadNumber()
	case "Race":
		ch.Race = sc.ReadNumber()
	case "Languages":
		ch.Speaks = sc.ReadNumber()
		ch.Speaking = sc.ReadNumber()
	case "Level":
		ch.Level = sc.ReadNumber()
	case "Played":
		ch.Played = sc.ReadNumber()
	case "Room":
		ch.HomeVnum = sc.ReadNumber()
	case "HpManaMove":
		ch.Hit = sc.ReadNumber()
		ch.MaxHit = sc.ReadNumber()
		ch.Mana = sc.ReadNumber()
		ch.MaxMana = sc.ReadNumber()
		ch.Move = sc.ReadNumber()
		ch.MaxMove = sc.ReadNumber()
	case "Stance":
		ch.Stance = sc.ReadNumber()
	case "Stances":
		for i := 0; i < types.MAX_STANCE; i++ {
			ch.Stances[i] = sc.ReadNumber()
		}
	case "Gold":
		ch.Gold = sc.ReadNumber()
	case "Silver":
		ch.Silver = sc.ReadNumber()
	case "Copper":
		ch.Copper = sc.ReadNumber()
	case "Exp":
		ch.Exp = sc.ReadNumber()
	case "Height":
		ch.Height = sc.ReadNumber()
	case "Weight":
		ch.Weight = sc.ReadNumber()
	case "Act":
		bv, _ := types.ParseBitVector(sc.ReadToEOL())
		ch.Act = bv
	case "AffectedBy":
		bv, _ := types.ParseBitVector(sc.ReadToEOL())
		ch.AffectedBy = bv
	case "NoAffectedBy":
		bv, _ := types.ParseBitVector(sc.ReadToEOL())
		ch.NoAffectedBy = bv
	case "Position":
		pos := sc.ReadNumber()
		if pos >= 100 {
			ch.Position = pos - 100
		} else {
			ch.Position = pos
		}
	case "Style":
		ch.Style = sc.ReadNumber()
	case "Practice":
		ch.Practice = sc.ReadNumber()
	case "Alignment":
		ch.Alignment = sc.ReadNumber()
	case "SavingThrows":
		ch.SavingPoisonDeath = sc.ReadNumber()
		ch.SavingWand = sc.ReadNumber()
		ch.SavingParaPetri = sc.ReadNumber()
		ch.SavingBreath = sc.ReadNumber()
		ch.SavingSpellStaff = sc.ReadNumber()
	case "Favor":
		ch.PCData.Favor = sc.ReadNumber()
	case "Honour":
		ch.PCData.Honour = sc.ReadNumber()
	case "Hitroll":
		ch.Hitroll = sc.ReadNumber()
	case "Damroll":
		ch.Damroll = sc.ReadNumber()
	case "Armor":
		ch.Armor = sc.ReadNumber()
	case "Wimpy":
		ch.Wimpy = sc.ReadNumber()
	case "Deaf":
		bv, _ := types.ParseBitVector(sc.ReadToEOL())
		ch.Deaf = bv
	case "Resistant":
		ch.Resistant = sc.ReadNumber()
	case "Immune":
		ch.Immune = sc.ReadNumber()
	case "Susceptible":
		ch.Susceptible = sc.ReadNumber()
	case "NoResistant":
		ch.NoResistant = sc.ReadNumber()
	case "NoImmune":
		ch.NoImmune = sc.ReadNumber()
	case "NoSusceptible":
		ch.NoSusceptible = sc.ReadNumber()
	case "Mentalstate":
		ch.MentalState = sc.ReadNumber()
	case "Password":
		ch.PCData.Pwd = sc.ReadString()
	case "Rank":
		ch.PCData.Rank = sc.ReadString()
	case "Bestowments":
		ch.PCData.Bestowments = sc.ReadString()
	case "Title":
		ch.PCData.Title = sc.ReadString()
	case "Homepage":
		ch.PCData.Homepage = sc.ReadString()
	case "Email":
		ch.PCData.Email = sc.ReadString()
	case "Bio":
		ch.PCData.Bio = sc.ReadString()
	case "AuthedBy":
		ch.PCData.AuthedBy = sc.ReadString()
	case "Minsnoop":
		ch.PCData.MinSnoop = sc.ReadNumber()
	case "Prompt":
		ch.PCData.Prompt = sc.ReadString()
	case "FPrompt":
		ch.PCData.FPrompt = sc.ReadString()
	case "Pagerlen":
		ch.PCData.PagerLen = sc.ReadNumber()
	case "Trust":
		trust := sc.ReadNumber()
		if trust > types.LEVEL_SUPREME {
			trust = types.LEVEL_SUPREME
		}
		ch.Trust = trust
	case "WizInvis":
		ch.PCData.WizInvis = sc.ReadNumber()
	case "Bamfin":
		ch.PCData.BamfIn = sc.ReadString()
	case "Bamfout":
		ch.PCData.BamfOut = sc.ReadString()
	case "Flags":
		ch.PCData.Flags = sc.ReadNumber()
	case "PKills":
		ch.PCData.PKills = sc.ReadNumber()
	case "PDeaths":
		ch.PCData.PDeaths = sc.ReadNumber()
	case "MKills":
		ch.PCData.MKills = sc.ReadNumber()
	case "MDeaths":
		ch.PCData.MDeaths = sc.ReadNumber()
	case "Hotboot":
		// Hotboot flag — set by DoHotboot pre-exec, carried through the
		// pfile, cleared post-recovery by BootRecover. See
		// plan-phase6-hotboot.md §Chosen Design step 4.
		ch.PCData.Hotboot = sc.ReadNumber() != 0
	case "AKills":
		// Arena wins. Not cross-referenced against a C pfile yet
		// because ENABLE_ARENA was off in Go predecessors; key
		// name mirrors the Go field, not C's lowercase "Akills".
		ch.PCData.AKills = sc.ReadNumber()
	case "ADeaths":
		ch.PCData.ADeaths = sc.ReadNumber()
	case "IllegalPK":
		ch.PCData.IllegalPK = sc.ReadNumber()
	case "AttrPerm":
		ch.PermStr = sc.ReadNumber()
		ch.PermInt = sc.ReadNumber()
		ch.PermWis = sc.ReadNumber()
		ch.PermDex = sc.ReadNumber()
		ch.PermCon = sc.ReadNumber()
		ch.PermCha = sc.ReadNumber()
		ch.PermLck = sc.ReadNumber()
	case "AttrMod":
		ch.ModStr = sc.ReadNumber()
		ch.ModInt = sc.ReadNumber()
		ch.ModWis = sc.ReadNumber()
		ch.ModDex = sc.ReadNumber()
		ch.ModCon = sc.ReadNumber()
		ch.ModCha = sc.ReadNumber()
		ch.ModLck = sc.ReadNumber()
	case "Condition":
		for i := 0; i < types.MAX_CONDS && i < 5; i++ {
			ch.PCData.Condition[i] = sc.ReadNumber()
		}
	case "Site":
		ch.PCData.RecentSite = sc.ReadToEOL()
	case "Clan":
		ch.PCData.ClanName = sc.ReadString()
	case "Council":
		ch.PCData.CouncilName = sc.ReadString()
	case "Deity":
		ch.PCData.DeityName = sc.ReadString()
	case "Locale":
		ch.PCData.Lang = sc.ReadString()
	case "Spouse":
		ch.Spouse = sc.ReadString()
	case "Hell":
		ch.PCData.Hell = int64(sc.ReadNumber())
	case "HelledBy":
		ch.PCData.HelledBy = sc.ReadString()
	case "Skill", "Spell", "Weapon", "Tongue":
		pct := sc.ReadNumber()
		name := sc.ReadWord()
		if SkillNameLookup != nil && ch.PCData != nil {
			gsn := SkillNameLookup(name)
			if gsn >= 0 && gsn < types.MAX_SKILL {
				ch.PCData.Learned[gsn] = pct
			}
		}
	case "Killed":
		_ = sc.ReadNumber()
		_ = sc.ReadNumber()
	case "Alias":
		// Format: Alias <name>~ <expansion>~
		name := sc.ReadString()
		cmd := sc.ReadString()
		if name != "" && cmd != "" && ch.PCData != nil {
			ch.PCData.Aliases = append(ch.PCData.Aliases,
				&types.AliasData{Name: name, Cmd: cmd})
		}
	case "Coordinates":
		ch.X = sc.ReadNumber()
		ch.Y = sc.ReadNumber()
		ch.Map = sc.ReadNumber()
	default:
		sc.ReadToEOL()
	}
}

func skipPlayerObject(sc *Scanner) {
	for {
		word := sc.ReadWord()
		if word == "" || word == "End" {
			return
		}
		if word == "#OBJECT" || word == "#CORPSE" {
			skipPlayerObject(sc)
			continue
		}
		sc.ReadToEOL()
	}
}

// readPlayerObject reads one #OBJECT section from a player file.
func readPlayerObject(sc *Scanner, lookup ObjIndexLookup, nestObj []*types.ObjData) *types.ObjData {
	obj := &types.ObjData{
		WearLoc: types.WEAR_NONE,
		Count:   1,
		Weight:  1,
	}
	nest := 0
	gotVnum := false

	for {
		word := sc.ReadWord()
		if word == "" || word == "End" {
			break
		}
		if word == "#OBJECT" || word == "#CORPSE" {
			readPlayerObject(sc, lookup, nestObj)
			continue
		}

		switch word {
		case "Nest":
			nest = sc.ReadNumber()
			if nest < 0 || nest >= len(nestObj) {
				nest = 0
			}
		case "Vnum":
			vnum := sc.ReadNumber()
			if lookup != nil {
				idx := lookup(vnum)
				if idx != nil {
					obj.IndexData = idx
					obj.Name = idx.Name
					obj.ShortDescr = idx.ShortDescr
					obj.Description = idx.Description
					obj.ItemType = idx.ItemType
					obj.Weight = idx.Weight
					obj.Level = idx.Level
					obj.Value = idx.Value
					obj.WearFlags = idx.WearFlags
					obj.ExtraFlags = idx.ExtraFlags
					gotVnum = true
				}
			}
		case "Name":
			obj.Name = sc.ReadString()
		case "ShortDescr":
			obj.ShortDescr = sc.ReadString()
		case "Description":
			obj.Description = sc.ReadString()
		case "ActionDesc":
			obj.ActionDesc = sc.ReadString()
		case "Owner":
			obj.Owner = sc.ReadString()
		case "ExtraFlags":
			bv, _ := types.ParseBitVector(sc.ReadToEOL())
			obj.ExtraFlags = bv
		case "WearFlags":
			obj.WearFlags = sc.ReadNumber()
		case "WearLoc":
			obj.WearLoc = sc.ReadNumber()
		case "ItemType":
			obj.ItemType = sc.ReadNumber()
		case "Weight":
			obj.Weight = sc.ReadNumber()
		case "Level":
			obj.Level = sc.ReadNumber()
		case "Timer":
			obj.Timer = sc.ReadNumber()
		case "Cost":
			obj.GoldCost = sc.ReadNumber()
		case "Count":
			obj.Count = sc.ReadNumber()
		case "Values":
			for i := 0; i < 6; i++ {
				obj.Value[i] = sc.ReadNumber()
			}
		case "Affect":
			aff := readAffect(sc)
			if aff != nil {
				obj.Affects = append(obj.Affects, aff)
			}
		case "AffectData":
			sc.ReadToEOL()
		default:
			sc.ReadToEOL()
		}
	}

	if !gotVnum {
		return nil
	}

	// Place in nest hierarchy
	nestObj[nest] = obj
	if nest > 0 {
		parent := nestObj[nest-1]
		if parent != nil {
			parent.Contents = append(parent.Contents, obj)
			obj.InObj = parent
		}
	}

	return obj
}

// readAffect reads one "Affect" line: type duration modifier location bitvector
func readAffect(sc *Scanner) *types.AffectData {
	aff := &types.AffectData{}
	aff.Type = sc.ReadNumber()
	aff.Duration = sc.ReadNumber()
	aff.Modifier = sc.ReadNumber()
	aff.Location = sc.ReadNumber()
	bvStr := sc.ReadToEOL()
	if bvStr != "" {
		bv, _ := types.ParseBitVector(strings.TrimSpace(bvStr))
		aff.BitVector = bv
	}
	return aff
}

// SavePlayer writes a player character to the SMAUG player save format.
func SavePlayer(w io.Writer, ch *types.CharData) error {
	p := ch.PCData
	if p == nil {
		return fmt.Errorf("SavePlayer: character has no PCData")
	}

	fmt.Fprintf(w, "#PLAYER\n")
	fmt.Fprintf(w, "Version    2\n")
	fmt.Fprintf(w, "Name       %s~\n", ch.Name)
	if ch.Description != "" {
		fmt.Fprintf(w, "Description %s~\n", ch.Description)
	}
	fmt.Fprintf(w, "Sex        %d\n", ch.Sex)
	fmt.Fprintf(w, "Class      %d\n", ch.Class)
	fmt.Fprintf(w, "Race       %d\n", ch.Race)
	fmt.Fprintf(w, "Languages  %d %d\n", ch.Speaks, ch.Speaking)
	fmt.Fprintf(w, "Level      %d\n", ch.Level)
	fmt.Fprintf(w, "Played     %d\n", ch.Played)
	fmt.Fprintf(w, "Room       %d\n", roomVnum(ch))
	fmt.Fprintf(w, "HpManaMove %d %d %d %d %d %d\n",
		ch.Hit, ch.MaxHit, ch.Mana, ch.MaxMana, ch.Move, ch.MaxMove)
	fmt.Fprintf(w, "Gold       %d\n", ch.Gold)
	fmt.Fprintf(w, "Silver     %d\n", ch.Silver)
	fmt.Fprintf(w, "Copper     %d\n", ch.Copper)
	fmt.Fprintf(w, "Exp        %d\n", ch.Exp)
	fmt.Fprintf(w, "Height     %d\n", ch.Height)
	fmt.Fprintf(w, "Weight     %d\n", ch.Weight)
	if !ch.Act.IsEmpty() {
		fmt.Fprintf(w, "Act        %s\n", ch.Act.String())
	}
	if !ch.AffectedBy.IsEmpty() {
		fmt.Fprintf(w, "AffectedBy %s\n", ch.AffectedBy.String())
	}
	if !ch.Deaf.IsEmpty() {
		fmt.Fprintf(w, "Deaf       %s\n", ch.Deaf.String())
	}
	fmt.Fprintf(w, "Position   %d\n", ch.Position+100)
	fmt.Fprintf(w, "Style      %d\n", ch.Style)
	fmt.Fprintf(w, "Practice   %d\n", ch.Practice)
	fmt.Fprintf(w, "Alignment  %d\n", ch.Alignment)
	fmt.Fprintf(w, "SavingThrows %d %d %d %d %d\n",
		ch.SavingPoisonDeath, ch.SavingWand, ch.SavingParaPetri,
		ch.SavingBreath, ch.SavingSpellStaff)
	fmt.Fprintf(w, "Hitroll    %d\n", ch.Hitroll)
	fmt.Fprintf(w, "Damroll    %d\n", ch.Damroll)
	fmt.Fprintf(w, "Armor      %d\n", ch.Armor)
	if ch.Wimpy != 0 {
		fmt.Fprintf(w, "Wimpy      %d\n", ch.Wimpy)
	}
	fmt.Fprintf(w, "AttrPerm   %d %d %d %d %d %d %d\n",
		ch.PermStr, ch.PermInt, ch.PermWis, ch.PermDex,
		ch.PermCon, ch.PermCha, ch.PermLck)
	fmt.Fprintf(w, "AttrMod    %d %d %d %d %d %d %d\n",
		ch.ModStr, ch.ModInt, ch.ModWis, ch.ModDex,
		ch.ModCon, ch.ModCha, ch.ModLck)
	fmt.Fprintf(w, "Condition  %d %d %d %d\n",
		p.Condition[0], p.Condition[1], p.Condition[2], p.Condition[3])
	fmt.Fprintf(w, "Password   %s~\n", p.Pwd)
	if p.Title != "" {
		fmt.Fprintf(w, "Title      %s~\n", util.SmashTilde(p.Title))
	}
	if p.Bio != "" {
		// Strip tildes so the field's terminator-tilde convention is
		// not corrupted by user input. C save.c writes bio via fwrite
		// which does the same sanitization; our Scanner matches.
		fmt.Fprintf(w, "Bio        %s~\n", util.SmashTilde(p.Bio))
	}
	if p.Prompt != "" {
		fmt.Fprintf(w, "Prompt     %s~\n", p.Prompt)
	}
	fmt.Fprintf(w, "Pagerlen   %d\n", p.PagerLen)
	fmt.Fprintf(w, "Flags      %d\n", p.Flags)
	fmt.Fprintf(w, "PKills     %d\n", p.PKills)
	fmt.Fprintf(w, "PDeaths    %d\n", p.PDeaths)
	fmt.Fprintf(w, "MKills     %d\n", p.MKills)
	fmt.Fprintf(w, "MDeaths    %d\n", p.MDeaths)
	if p.Hotboot {
		// Hotboot flag — carried across the pfile save/exec/load cycle
		// so BootRecover can clear it after the welcome. Only emitted
		// when true so non-hotboot saves don't bloat pfiles.
		fmt.Fprintf(w, "Hotboot    1\n")
	}
	if p.AKills != 0 {
		fmt.Fprintf(w, "AKills     %d\n", p.AKills)
	}
	if p.ADeaths != 0 {
		fmt.Fprintf(w, "ADeaths    %d\n", p.ADeaths)
	}
	if p.RecentSite != "" {
		fmt.Fprintf(w, "Site       %s\n", p.RecentSite)
	}
	if ch.Trust != 0 {
		fmt.Fprintf(w, "Trust      %d\n", ch.Trust)
	}
	if p.Hell != 0 {
		fmt.Fprintf(w, "Hell       %d\n", p.Hell)
	}
	if p.HelledBy != "" {
		fmt.Fprintf(w, "HelledBy   %s~\n", p.HelledBy)
	}

	// Save affects
	for _, aff := range ch.Affects {
		fmt.Fprintf(w, "Affect       %d %d %d %d %s\n",
			aff.Type, aff.Duration, aff.Modifier, aff.Location,
			aff.BitVector.String())
	}

	// Save aliases (one per line). Tildes are stripped by util.SmashTilde —
	// DoAlias already refuses literal ~ input so this is belt-and-braces.
	for _, a := range p.Aliases {
		if a == nil || a.Name == "" || a.Cmd == "" {
			continue
		}
		fmt.Fprintf(w, "Alias      %s~ %s~\n",
			util.SmashTilde(a.Name), util.SmashTilde(a.Cmd))
	}

	// Save learned skill/spell/weapon/tongue proficiencies (bug G4).
	// C emits one line per non-zero Learned entry, keyed by skill type.
	if SkillGetter != nil {
		for gsn, pct := range p.Learned {
			if pct <= 0 {
				continue
			}
			sk := SkillGetter(gsn)
			if sk == nil {
				continue
			}
			var keyword string
			switch sk.Type {
			case types.SKILL_SPELL:
				keyword = "Spell"
			case types.SKILL_WEAPON:
				keyword = "Weapon"
			case types.SKILL_TONGUE:
				keyword = "Tongue"
			default:
				// C fwrite_char's default case writes everything else under
				// "Skill" — covers SKILL_SKILL, SKILL_RACIAL, SKILL_HERB,
				// SKILL_DISEASE, SKILL_UNKNOWN. Dropping these silently loses
				// player proficiency (e.g. venomshield at gsn 0).
				keyword = "Skill"
			}
			fmt.Fprintf(w, "%-10s %d '%s'\n", keyword, pct, sk.Name)
		}
	}

	fmt.Fprintf(w, "End\n\n")

	// #MorphData block — plan-phase6-polymorph.md §G6. Emitted
	// unconditionally on ch.Morph != nil (mirrors C save.c:287-288).
	// Stat deltas are already baked into base stats, so the pfile
	// stores the player "mid-morph"; DoUnmorphChar on return subtracts
	// the delta stored on this block.
	if err := writeMorphData(w, ch); err != nil {
		return err
	}

	// Save carried/equipped objects (after End, matching C format)
	for _, obj := range ch.Carrying {
		writePlayerObj(w, obj, 0)
	}

	return nil
}

// writePlayerObj writes one #OBJECT section, recursing into contents.
func writePlayerObj(w io.Writer, obj *types.ObjData, nest int) {
	vnum := 0
	if obj.IndexData != nil {
		vnum = obj.IndexData.Vnum
	}
	if vnum == 0 {
		return
	}

	fmt.Fprintf(w, "#OBJECT\n")
	if nest > 0 {
		fmt.Fprintf(w, "Nest         %d\n", nest)
	}
	fmt.Fprintf(w, "Vnum         %d\n", vnum)
	if obj.Name != "" && (obj.IndexData == nil || obj.Name != obj.IndexData.Name) {
		fmt.Fprintf(w, "Name         %s~\n", obj.Name)
	}
	if obj.ShortDescr != "" && (obj.IndexData == nil || obj.ShortDescr != obj.IndexData.ShortDescr) {
		fmt.Fprintf(w, "ShortDescr   %s~\n", obj.ShortDescr)
	}
	if !obj.ExtraFlags.IsEmpty() && (obj.IndexData == nil || obj.ExtraFlags != obj.IndexData.ExtraFlags) {
		fmt.Fprintf(w, "ExtraFlags   %s\n", obj.ExtraFlags.String())
	}
	if obj.WearLoc != types.WEAR_NONE {
		fmt.Fprintf(w, "WearLoc      %d\n", obj.WearLoc)
	}
	if obj.IndexData == nil || obj.ItemType != obj.IndexData.ItemType {
		fmt.Fprintf(w, "ItemType     %d\n", obj.ItemType)
	}
	if obj.IndexData == nil || obj.Weight != obj.IndexData.Weight {
		fmt.Fprintf(w, "Weight       %d\n", obj.Weight)
	}
	if obj.Level > 0 {
		fmt.Fprintf(w, "Level        %d\n", obj.Level)
	}
	if obj.Timer > 0 {
		fmt.Fprintf(w, "Timer        %d\n", obj.Timer)
	}
	if obj.Count > 1 {
		fmt.Fprintf(w, "Count        %d\n", obj.Count)
	}

	hasValues := false
	for _, v := range obj.Value {
		if v != 0 {
			hasValues = true
			break
		}
	}
	if hasValues {
		fmt.Fprintf(w, "Values       %d %d %d %d %d %d\n",
			obj.Value[0], obj.Value[1], obj.Value[2],
			obj.Value[3], obj.Value[4], obj.Value[5])
	}

	// Object affects
	for _, aff := range obj.Affects {
		fmt.Fprintf(w, "Affect       %d %d %d %d %s\n",
			aff.Type, aff.Duration, aff.Modifier, aff.Location,
			aff.BitVector.String())
	}

	fmt.Fprintf(w, "End\n\n")

	// Recurse into contents
	for _, inner := range obj.Contents {
		writePlayerObj(w, inner, nest+1)
	}
}

func roomVnum(ch *types.CharData) int {
	if ch.InRoom != nil {
		return ch.InRoom.Vnum
	}
	if ch.HomeVnum != 0 {
		return ch.HomeVnum
	}
	return types.ROOM_VNUM_TEMPLE
}

// validPlayerName matches names that are 3-12 alphabetic characters.
var validPlayerName = regexp.MustCompile(`^[a-zA-Z]{3,12}$`)

// PlayerFilePath returns the path for a player's save file.
// It validates the name to prevent path traversal attacks.
func PlayerFilePath(dataDir, name string) string {
	if name == "" {
		return ""
	}
	name = filepath.Base(name)
	if !validPlayerName.MatchString(name) {
		return ""
	}
	first := strings.ToLower(name[:1])
	return fmt.Sprintf("%s/player/%s/%s", dataDir, first, name)
}
