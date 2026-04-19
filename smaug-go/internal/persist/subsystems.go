package persist

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
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

	return readClan(NewScanner(f, path))
}

// readClan parses a single clan block from an already-open Scanner. Shared
// by loadClan (file path) and tests (bytes.Reader). Signature matches C
// fread_clan at src/clans.c:297-510.
func readClan(sc *Scanner) (*types.ClanData, error) {
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
		case "Abbrev":
			clan.Abbrev = sc.ReadString()
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
		case "Leadrank":
			clan.LeadRank = sc.ReadString()
		case "Onerank":
			clan.OneRank = sc.ReadString()
		case "Tworank":
			clan.TwoRank = sc.ReadString()
		case "PKills":
			// Legacy single-int form: C fread_clan:435 reads into index [6],
			// NOT [0]. Pre-existing Go bug fixed 2026-04-19.
			clan.PKills[6] = sc.ReadNumber()
		case "PDeaths":
			// Legacy single-int form: C fread_clan:434 reads into [6].
			clan.PDeaths[6] = sc.ReadNumber()
		case "PKillRange":
			// Legacy 7-int block: C fread_clan:470-479 reads-and-discards.
			for i := 0; i < 7; i++ {
				sc.ReadNumber()
			}
		case "PDeathRange":
			// Legacy 7-int block: C fread_clan:437-447 reads-and-discards.
			for i := 0; i < 7; i++ {
				sc.ReadNumber()
			}
		case "PKillRangeNew":
			// Active 7-int block: C fread_clan:459-469.
			for i := 0; i < 7; i++ {
				clan.PKills[i] = sc.ReadNumber()
			}
		case "PDeathRangeNew":
			// Active 7-int block: C fread_clan:448-458.
			for i := 0; i < 7; i++ {
				clan.PDeaths[i] = sc.ReadNumber()
			}
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
		case "MemLimit":
			clan.MemLimit = sc.ReadNumber()
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
		case "ClanObjFour":
			clan.ClanObj4 = sc.ReadNumber()
		case "ClanObjFive":
			clan.ClanObj5 = sc.ReadNumber()
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

// SaveClan writes a clan record in the SMAUG save_clan format. Byte-for-byte
// parity with C src/clans.c:206-250 save_clan. Strings get tilde terminators
// and two-space-padded key columns matching C's fprintf format strings.
func SaveClan(w io.Writer, c *types.ClanData) error {
	if c == nil {
		return fmt.Errorf("SaveClan: nil clan")
	}

	// Apply SmashTilde to every string field (belt-and-braces vs. builder-
	// typed tildes, matching C smash_tilde on read-in).
	smash := util.SmashTilde

	fmt.Fprintf(w, "#CLAN\n")
	fmt.Fprintf(w, "Name         %s~\n", smash(c.Name))
	fmt.Fprintf(w, "Abbrev       %s~\n", smash(c.Abbrev))
	fmt.Fprintf(w, "Filename     %s~\n", smash(c.Filename))
	fmt.Fprintf(w, "Motto        %s~\n", smash(c.Motto))
	fmt.Fprintf(w, "Description  %s~\n", smash(c.Description))
	fmt.Fprintf(w, "Deity        %s~\n", smash(c.Deity))
	fmt.Fprintf(w, "Leader       %s~\n", smash(c.Leader))
	fmt.Fprintf(w, "NumberOne    %s~\n", smash(c.Number1))
	fmt.Fprintf(w, "NumberTwo    %s~\n", smash(c.Number2))
	fmt.Fprintf(w, "Badge        %s~\n", smash(c.Badge))
	fmt.Fprintf(w, "Leadrank     %s~\n", smash(c.LeadRank))
	fmt.Fprintf(w, "Onerank      %s~\n", smash(c.OneRank))
	fmt.Fprintf(w, "Tworank      %s~\n", smash(c.TwoRank))
	fmt.Fprintf(w, "PKillRangeNew   %d %d %d %d %d %d %d\n",
		c.PKills[0], c.PKills[1], c.PKills[2],
		c.PKills[3], c.PKills[4], c.PKills[5], c.PKills[6])
	fmt.Fprintf(w, "PDeathRangeNew  %d %d %d %d %d %d %d\n",
		c.PDeaths[0], c.PDeaths[1], c.PDeaths[2],
		c.PDeaths[3], c.PDeaths[4], c.PDeaths[5], c.PDeaths[6])
	fmt.Fprintf(w, "MKills       %d\n", c.MKills)
	fmt.Fprintf(w, "MDeaths      %d\n", c.MDeaths)
	fmt.Fprintf(w, "IllegalPK    %d\n", c.IllegalPK)
	fmt.Fprintf(w, "Score        %d\n", c.Score)
	fmt.Fprintf(w, "Type         %d\n", c.ClanType)
	fmt.Fprintf(w, "Class        %d\n", c.Class)
	fmt.Fprintf(w, "Favour       %d\n", c.Favour)
	fmt.Fprintf(w, "Strikes      %d\n", c.Strikes)
	fmt.Fprintf(w, "Members      %d\n", c.Members)
	fmt.Fprintf(w, "MemLimit     %d\n", c.MemLimit)
	fmt.Fprintf(w, "Alignment    %d\n", c.Alignment)
	fmt.Fprintf(w, "Board        %d\n", c.Board)
	fmt.Fprintf(w, "ClanObjOne   %d\n", c.ClanObj1)
	fmt.Fprintf(w, "ClanObjTwo   %d\n", c.ClanObj2)
	fmt.Fprintf(w, "ClanObjThree %d\n", c.ClanObj3)
	fmt.Fprintf(w, "ClanObjFour  %d\n", c.ClanObj4)
	fmt.Fprintf(w, "ClanObjFive  %d\n", c.ClanObj5)
	fmt.Fprintf(w, "Recall       %d\n", c.Recall)
	fmt.Fprintf(w, "Storeroom    %d\n", c.Storeroom)
	fmt.Fprintf(w, "GuardOne     %d\n", c.Guard1)
	fmt.Fprintf(w, "GuardTwo     %d\n", c.Guard2)
	fmt.Fprintf(w, "End\n\n")
	fmt.Fprintf(w, "#END\n")
	return nil
}

// SaveClanFile writes the clan to `<dir>/<clan.Filename>` using SaveClan
// serialization. Matches C sprintf(filename, "%s%s", CLAN_DIR, clan->filename)
// at src/clans.c:196. Returns an error on empty filename or I/O failure.
func SaveClanFile(dir string, c *types.ClanData) error {
	if c == nil {
		return fmt.Errorf("SaveClanFile: nil clan")
	}
	if c.Filename == "" {
		return fmt.Errorf("SaveClanFile: clan %q has no filename", c.Name)
	}
	path := filepath.Join(dir, c.Filename)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("SaveClanFile: create %s: %w", path, err)
	}
	defer f.Close()
	if err := SaveClan(f, c); err != nil {
		return fmt.Errorf("SaveClanFile: write %s: %w", path, err)
	}
	return nil
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
