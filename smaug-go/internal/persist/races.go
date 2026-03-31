package persist

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// LoadRaces loads all races listed in test_race.lst (or race.lst) from the
// given directory. Each entry names a .race file to load.
func LoadRaces(w *world.World, raceDir string) error {
	// Try test_race.lst first (for testdata), then race.lst
	listPath := filepath.Join(raceDir, "race.lst")
	if _, err := os.Stat(listPath); os.IsNotExist(err) {
		alt := filepath.Join(raceDir, "test_race.lst")
		if _, err2 := os.Stat(alt); err2 == nil {
			listPath = alt
		}
	}

	data, err := os.ReadFile(listPath)
	if err != nil {
		return fmt.Errorf("LoadRaces: cannot read %s: %w", listPath, err)
	}

	// Ensure Races slice is large enough
	if w.Races == nil || len(w.Races) < types.MAX_NPC_RACE {
		newRaces := make([]*types.RaceData, types.MAX_NPC_RACE)
		copy(newRaces, w.Races)
		w.Races = newRaces
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "$" {
			break
		}
		racePath := filepath.Join(raceDir, line)
		if err := loadRaceFile(w, racePath); err != nil {
			util.Bug("LoadRaces: error loading %s: %v", racePath, err)
		}
	}
	return nil
}

// loadRaceFile reads a single .race file and populates the corresponding
// entry in w.Races[].
func loadRaceFile(w *world.World, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("loadRaceFile: cannot open %s: %w", filename, err)
	}
	defer f.Close()

	sc := NewScanner(f, filename)
	race := &types.RaceData{}
	whereIdx := 0

	for {
		word := sc.ReadWord()
		if word == "" {
			break
		}

		switch strings.ToLower(word) {
		case "name":
			race.Name = sc.ReadString()
		case "race":
			idx := sc.ReadNumber()
			if idx >= 0 && idx < len(w.Races) {
				w.Races[idx] = race
			} else {
				return sc.Errorf("race index %d out of range", idx)
			}
		case "classes":
			race.ClassRestriction = sc.ReadNumber()
		case "str_plus":
			race.StrPlus = sc.ReadNumber()
		case "dex_plus":
			race.DexPlus = sc.ReadNumber()
		case "wis_plus":
			race.WisPlus = sc.ReadNumber()
		case "int_plus":
			race.IntPlus = sc.ReadNumber()
		case "con_plus":
			race.ConPlus = sc.ReadNumber()
		case "cha_plus":
			race.ChaPlus = sc.ReadNumber()
		case "lck_plus":
			race.LckPlus = sc.ReadNumber()
		case "hit":
			race.Hit = sc.ReadNumber()
		case "mana":
			race.Mana = sc.ReadNumber()
		case "affected":
			race.Affected = types.BitVectorFromInt(uint32(sc.ReadNumber()))
		case "resist":
			race.Resist = sc.ReadNumber()
		case "suscept":
			race.Suscept = sc.ReadNumber()
		case "language":
			race.Language = sc.ReadNumber()
		case "align":
			race.Alignment = sc.ReadNumber()
		case "min_align":
			race.MinAlign = sc.ReadNumber()
		case "max_align":
			race.MaxAlign = sc.ReadNumber()
		case "ac_plus":
			race.ACPlus = sc.ReadNumber()
		case "exp_mult":
			race.ExpMultiplier = sc.ReadNumber()
		case "attacks":
			race.Attacks = types.BitVectorFromInt(uint32(sc.ReadNumber()))
		case "defenses":
			race.Defenses = types.BitVectorFromInt(uint32(sc.ReadNumber()))
		case "height":
			race.Height = sc.ReadNumber()
		case "weight":
			race.Weight = sc.ReadNumber()
		case "hunger_mod":
			race.HungerMod = sc.ReadNumber()
		case "thirst_mod":
			race.ThirstMod = sc.ReadNumber()
		case "mana_regen":
			race.ManaRegen = sc.ReadNumber()
		case "hp_regen":
			race.HPRegen = sc.ReadNumber()
		case "race_recall":
			race.RaceRecall = sc.ReadNumber()
		case "skill":
			// Skip: word + 2 numbers
			sc.ReadWord()
			sc.ReadNumber()
			sc.ReadNumber()
		case "wherename":
			if whereIdx < types.MAX_WHERE_NAME {
				race.WhereName[whereIdx] = strings.TrimRight(sc.ReadString(), " \t")
				whereIdx++
			} else {
				sc.ReadString() // consume and discard
			}
		case "end":
			return nil
		default:
			util.Bug("loadRaceFile: %s:%d: unknown key %q", filename, sc.Line(), word)
			sc.ReadToEOL()
		}
	}

	return nil
}
