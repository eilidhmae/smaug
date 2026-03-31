package persist

import (
	"fmt"
	"io"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// LoadPlayer reads a player character from a SMAUG player save file.
func LoadPlayer(r io.Reader, filename string) (*types.CharData, error) {
	sc := NewScanner(r, filename)

	// Read #PLAYER header
	word := sc.ReadWord()
	if word != "#PLAYER" {
		return nil, fmt.Errorf("LoadPlayer: expected #PLAYER, got %q", word)
	}

	ch := &types.CharData{
		PCData: &types.PCData{
			PagerLen: 24,
		},
	}

	for {
		word = sc.ReadWord()
		if word == "" {
			break
		}

		switch word {
		case "End":
			return ch, nil
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
			ch.Trust = sc.ReadNumber()
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
		case "Skill", "Spell", "Weapon", "Tongue":
			_ = sc.ReadNumber()
			_ = sc.ReadString()
		case "Affect", "AffectData":
			sc.ReadToEOL()
		case "Killed":
			_ = sc.ReadNumber()
			_ = sc.ReadNumber()
		case "#OBJECT", "#CORPSE":
			skipPlayerObject(sc)
		case "Coordinates":
			ch.X = sc.ReadNumber()
			ch.Y = sc.ReadNumber()
			ch.Map = sc.ReadNumber()
		default:
			sc.ReadToEOL()
		}
	}

	return ch, nil
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
	fmt.Fprintf(w, "Exp        %d\n", ch.Exp)
	fmt.Fprintf(w, "Height     %d\n", ch.Height)
	fmt.Fprintf(w, "Weight     %d\n", ch.Weight)
	if !ch.Act.IsEmpty() {
		fmt.Fprintf(w, "Act        %s\n", ch.Act.String())
	}
	if !ch.AffectedBy.IsEmpty() {
		fmt.Fprintf(w, "AffectedBy %s\n", ch.AffectedBy.String())
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
	if p.Prompt != "" {
		fmt.Fprintf(w, "Prompt     %s~\n", p.Prompt)
	}
	fmt.Fprintf(w, "Pagerlen   %d\n", p.PagerLen)
	fmt.Fprintf(w, "Flags      %d\n", p.Flags)
	fmt.Fprintf(w, "PKills     %d\n", p.PKills)
	fmt.Fprintf(w, "PDeaths    %d\n", p.PDeaths)
	fmt.Fprintf(w, "MKills     %d\n", p.MKills)
	fmt.Fprintf(w, "MDeaths    %d\n", p.MDeaths)
	if p.RecentSite != "" {
		fmt.Fprintf(w, "Site       %s\n", p.RecentSite)
	}
	if ch.Trust != 0 {
		fmt.Fprintf(w, "Trust      %d\n", ch.Trust)
	}
	fmt.Fprintf(w, "End\n\n")
	return nil
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

// PlayerFilePath returns the path for a player's save file.
func PlayerFilePath(dataDir, name string) string {
	if name == "" {
		return ""
	}
	first := strings.ToLower(name[:1])
	return fmt.Sprintf("%s/player/%s/%s", dataDir, first, name)
}
