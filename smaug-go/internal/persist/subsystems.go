package persist

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// LoadClansFromDir loads all clan files from the clans directory.
func LoadClansFromDir(w *world.World, dir string) error {
	listPath := filepath.Join(dir, "clan.lst")
	return loadFromList(listPath, func(filename string) {
		path := filepath.Join(dir, filename)
		clan, err := loadClan(path)
		if err != nil {
			log.Printf("Error loading clan %s: %v", filename, err)
			return
		}
		w.Clans = append(w.Clans, clan)
	})
}

func loadClan(path string) (*types.ClanData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sc := NewScanner(f, path)
	clan := &types.ClanData{}

	// Read past #CLAN
	word := sc.ReadWord()
	if word != "#CLAN" {
		return nil, sc.Errorf("expected #CLAN, got %s", word)
	}

	for {
		word = sc.ReadWord()
		switch word {
		case "Name":
			clan.Name = sc.ReadString()
		case "Filename":
			clan.Filename = sc.ReadString()
		case "Motto":
			clan.Motto = sc.ReadString()
		case "Description":
			clan.Description = sc.ReadString()
		case "Deity":
			clan.Deity = sc.ReadString()
		case "Leader":
			clan.Leader = sc.ReadString()
		case "NumberOne":
			clan.Number1 = sc.ReadString()
		case "NumberTwo":
			clan.Number2 = sc.ReadString()
		case "Badge":
			clan.Badge = sc.ReadString()
		case "PKills":
			clan.PKills[0] = sc.ReadNumber()
		case "PDeaths":
			clan.PDeaths[0] = sc.ReadNumber()
		case "MKills":
			clan.MKills = sc.ReadNumber()
		case "MDeaths":
			clan.MDeaths = sc.ReadNumber()
		case "IllegalPK":
			clan.IllegalPK = sc.ReadNumber()
		case "Score":
			clan.Score = sc.ReadNumber()
		case "Type":
			clan.ClanType = sc.ReadNumber()
		case "Class":
			clan.Class = sc.ReadNumber()
		case "Favour":
			clan.Favour = sc.ReadNumber()
		case "Strikes":
			clan.Strikes = sc.ReadNumber()
		case "Members":
			clan.Members = sc.ReadNumber()
		case "Alignment":
			clan.Alignment = sc.ReadNumber()
		case "Board":
			clan.Board = sc.ReadNumber()
		case "ClanObjOne":
			clan.ClanObj1 = sc.ReadNumber()
		case "ClanObjTwo":
			clan.ClanObj2 = sc.ReadNumber()
		case "ClanObjThree":
			clan.ClanObj3 = sc.ReadNumber()
		case "Recall":
			clan.Recall = sc.ReadNumber()
		case "Storeroom":
			clan.Storeroom = sc.ReadNumber()
		case "GuardOne":
			clan.Guard1 = sc.ReadNumber()
		case "GuardTwo":
			clan.Guard2 = sc.ReadNumber()
		case "End", "#END":
			return clan, nil
		default:
			// Skip unknown field
			sc.ReadToEOL()
		}
	}
}

// LoadDeitiesFromDir loads all deity files from the deity directory.
func LoadDeitiesFromDir(w *world.World, dir string) error {
	listPath := filepath.Join(dir, "deity.lst")
	return loadFromList(listPath, func(filename string) {
		path := filepath.Join(dir, filename)
		deity, err := loadDeity(path)
		if err != nil {
			log.Printf("Error loading deity %s: %v", filename, err)
			return
		}
		w.Deities = append(w.Deities, deity)
	})
}

func loadDeity(path string) (*types.DeityData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sc := NewScanner(f, path)
	d := &types.DeityData{}

	word := sc.ReadWord()
	if word != "#DEITY" {
		return nil, sc.Errorf("expected #DEITY, got %s", word)
	}

	for {
		word = sc.ReadWord()
		switch word {
		case "Filename":
			d.Filename = sc.ReadString()
		case "Name":
			d.Name = sc.ReadString()
		case "Description":
			d.Description = sc.ReadString()
		case "Alignment":
			d.Alignment = sc.ReadNumber()
		case "Worshippers":
			d.Worshippers = sc.ReadNumber()
		case "Flee":
			d.Flee = sc.ReadNumber()
		case "Flee_npcrace":
			d.FleeNpcRace = sc.ReadNumber()
		case "Flee_npcfoe":
			d.FleeNpcFoe = sc.ReadNumber()
		case "Kill":
			d.Kill = sc.ReadNumber()
		case "Kill_magic":
			d.KillMagic = sc.ReadNumber()
		case "Kill_npcrace":
			d.KillNpcRace = sc.ReadNumber()
		case "Kill_npcfoe":
			d.KillNpcFoe = sc.ReadNumber()
		case "Sac":
			d.Sac = sc.ReadNumber()
		case "Bury_corpse":
			d.BuryCorpse = sc.ReadNumber()
		case "Aid_spell":
			d.AidSpell = sc.ReadNumber()
		case "Aid":
			d.Aid = sc.ReadNumber()
		case "Backstab":
			d.Backstab = sc.ReadNumber()
		case "Steal":
			d.Steal = sc.ReadNumber()
		case "Die":
			d.Die = sc.ReadNumber()
		case "Die_npcrace":
			d.DieNpcRace = sc.ReadNumber()
		case "Die_npcfoe":
			d.DieNpcFoe = sc.ReadNumber()
		case "Spell_aid":
			d.SpellAid = sc.ReadNumber()
		case "Dig_corpse":
			d.DigCorpse = sc.ReadNumber()
		case "Scorpse":
			d.SCorpse = sc.ReadNumber()
		case "Savatar":
			d.SAvatar = sc.ReadNumber()
		case "Sdeityobj":
			d.SDeityObj = sc.ReadNumber()
		case "Srecall":
			d.SRecall = sc.ReadNumber()
		case "Race":
			d.Race = sc.ReadNumber()
		case "Race2":
			d.Race2 = sc.ReadNumber()
		case "Class":
			d.Class = sc.ReadNumber()
		case "Sex":
			d.Sex = sc.ReadNumber()
		case "Element":
			d.Element = sc.ReadNumber()
		case "Npcrace":
			d.NpcRace = sc.ReadNumber()
		case "Npcfoe":
			d.NpcFoe = sc.ReadNumber()
		case "Suscept":
			d.Suscept = sc.ReadNumber()
		case "Susceptnum":
			d.SusceptNum = sc.ReadNumber()
		case "Elementnum":
			d.ElementNum = sc.ReadNumber()
		case "Affectednum":
			d.AffectedNum = sc.ReadNumber()
		case "Objstat":
			d.ObjStat = sc.ReadNumber()
		case "Affected":
			d.Affected = sc.ReadBitvector()
		case "End", "#END":
			return d, nil
		default:
			sc.ReadToEOL()
		}
	}
}

// LoadSocials loads social commands from the socials.dat file.
func LoadSocials(w *world.World, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	sc := NewScanner(f, path)

	for {
		word := sc.ReadWord()
		if word == "" || word == "#END" || word == "$" {
			break
		}
		if word != "#SOCIAL" {
			continue
		}

		social := &types.SocialType{}
		for {
			key := sc.ReadWord()
			switch key {
			case "Name":
				social.Name = sc.ReadString()
			case "CharNoArg":
				social.CharNoArg = sc.ReadString()
			case "OthersNoArg":
				social.OthersNoArg = sc.ReadString()
			case "CharFound":
				social.CharFound = sc.ReadString()
			case "OthersFound":
				social.OthersFound = sc.ReadString()
			case "VictFound":
				social.VictFound = sc.ReadString()
			case "CharAuto":
				social.CharAuto = sc.ReadString()
			case "OthersAuto":
				social.OthersAuto = sc.ReadString()
			case "End":
				w.Socials = append(w.Socials, social)
				goto nextSocial
			default:
				sc.ReadToEOL()
			}
		}
	nextSocial:
	}

	return nil
}

// LoadBoards loads board definitions from boards.dat.
func LoadBoards(w *world.World, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var board *types.BoardData

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 1 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := ""
		if len(parts) > 1 {
			val = strings.TrimSpace(parts[1])
		}

		switch key {
		case "Filename":
			board = &types.BoardData{}
			board.NoteFile = strings.TrimSuffix(val, "~")
		case "Vnum":
			if board != nil {
				board.BoardObj, _ = strconv.Atoi(val)
			}
		case "Min_read_level":
			if board != nil {
				board.MinReadLevel, _ = strconv.Atoi(val)
			}
		case "Min_post_level":
			if board != nil {
				board.MinPostLevel, _ = strconv.Atoi(val)
			}
		case "Min_remove_level":
			if board != nil {
				board.MinRemoveLevel, _ = strconv.Atoi(val)
			}
		case "Max_posts":
			if board != nil {
				board.MaxPosts, _ = strconv.Atoi(val)
			}
		case "Type":
			if board != nil {
				board.Type, _ = strconv.Atoi(val)
			}
		case "End":
			if board != nil {
				w.Boards = append(w.Boards, board)
				board = nil
			}
		}
	}

	return scanner.Err()
}

// loadFromList reads a .lst file and calls fn for each filename until $.
func loadFromList(listPath string, fn func(string)) error {
	f, err := os.Open(listPath)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line == "$" {
			break
		}
		fn(line)
	}
	return scanner.Err()
}
