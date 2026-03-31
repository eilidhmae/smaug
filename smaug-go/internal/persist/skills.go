package persist

import (
	"fmt"
	"os"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// LoadSkills loads skills/spells from a SMAUG skills.dat file.
func LoadSkills(w *world.World, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("LoadSkills: cannot open %s: %w", filename, err)
	}
	defer f.Close()

	sc := NewScanner(f, filename)

	for {
		word := sc.ReadWord()
		if word == "" {
			break
		}

		if word == "#END" || word == "$" {
			break
		}

		if word == "#SKILL" {
			skill := readSkill(sc)
			if skill != nil {
				w.Skills = append(w.Skills, skill)
			}
			continue
		}

		// Skip unknown sections
		util.Bug("LoadSkills: %s: unexpected word %q", filename, word)
		sc.ReadToEOL()
	}

	return nil
}

// readSkill reads a single skill definition from the scanner.
func readSkill(sc *Scanner) *types.SkillType {
	skill := &types.SkillType{
		Guild: -1,
	}

	// Initialize skill/race levels to LEVEL_IMMORTAL with default adept
	for i := 0; i < types.MAX_CLASS; i++ {
		skill.SkillLevel[i] = types.LEVEL_IMMORTAL
		skill.SkillAdept[i] = 95
	}
	for i := 0; i < types.MAX_RACE; i++ {
		skill.RaceLevel[i] = types.LEVEL_IMMORTAL
		skill.RaceAdept[i] = 95
	}

	for {
		word := sc.ReadWord()
		if word == "" {
			break
		}

		switch word {
		case "End":
			return skill
		case "Name":
			skill.Name = sc.ReadString()
		case "Type":
			skill.Type = parseSkillType(sc.ReadWord())
		case "Info":
			skill.Info = sc.ReadNumber()
		case "Flags":
			skill.Flags = sc.ReadNumber()
		case "Target":
			skill.Target = sc.ReadNumber()
		case "Minpos":
			pos := sc.ReadNumber()
			skill.MinimumPos = convertMinPos(pos)
		case "Saves":
			skill.SaveType = sc.ReadNumber()
		case "Slot":
			skill.Slot = sc.ReadNumber()
		case "Mana":
			skill.MinMana = sc.ReadNumber()
		case "Rounds":
			skill.Beats = sc.ReadNumber()
		case "Range":
			skill.Range = sc.ReadNumber()
		case "Code":
			codeName := sc.ReadWord()
			if strings.HasPrefix(codeName, "spell_") {
				skill.SpellFunName = codeName
			} else if strings.HasPrefix(codeName, "do_") {
				skill.SkillFunName = codeName
			} else {
				// Could be either — store as spell by default
				skill.SpellFunName = codeName
			}
		case "Dammsg":
			skill.NounDamage = sc.ReadString()
		case "Dice":
			skill.DiceFormula = sc.ReadString()
		case "Wearoff":
			skill.MsgOff = sc.ReadString()
		case "Hitchar":
			skill.HitChar = sc.ReadString()
		case "Hitvict":
			skill.HitVict = sc.ReadString()
		case "Hitroom":
			skill.HitRoom = sc.ReadString()
		case "Misschar":
			skill.MissChar = sc.ReadString()
		case "Missvict":
			skill.MissVict = sc.ReadString()
		case "Missroom":
			skill.MissRoom = sc.ReadString()
		case "Diechar":
			skill.DieChar = sc.ReadString()
		case "Dievict":
			skill.DieVict = sc.ReadString()
		case "Dieroom":
			skill.DieRoom = sc.ReadString()
		case "Immchar":
			skill.ImmChar = sc.ReadString()
		case "Immvict":
			skill.ImmVict = sc.ReadString()
		case "Immroom":
			skill.ImmRoom = sc.ReadString()
		case "Abschar":
			skill.AbsChar = sc.ReadString()
		case "Absvict":
			skill.AbsVict = sc.ReadString()
		case "Absroom":
			skill.AbsRoom = sc.ReadString()
		case "Alignment":
			skill.Alignment = sc.ReadNumber()
		case "Difficulty":
			skill.Difficulty = sc.ReadNumber()
		case "Guild":
			skill.Guild = sc.ReadNumber()
		case "Components":
			skill.Components = sc.ReadString()
		case "Teachers":
			skill.Teachers = sc.ReadString()
		case "Participants":
			skill.Participants = sc.ReadNumber()
		case "Value":
			// Object value for create-type spells; read and discard
			sc.ReadNumber()
		case "Class":
			classIdx := sc.ReadNumber()
			level := sc.ReadNumber()
			adept := sc.ReadNumber()
			if classIdx >= 0 && classIdx < types.MAX_CLASS {
				skill.SkillLevel[classIdx] = level
				skill.SkillAdept[classIdx] = adept
			}
		case "Race":
			raceIdx := sc.ReadNumber()
			level := sc.ReadNumber()
			adept := sc.ReadNumber()
			if raceIdx >= 0 && raceIdx < types.MAX_RACE {
				skill.RaceLevel[raceIdx] = level
				skill.RaceAdept[raceIdx] = adept
			}
		case "Affect":
			aff := readSmaugAffect(sc)
			if aff != nil {
				skill.Affects = append(skill.Affects, aff)
			}
		case "Minlevel":
			// Legacy field — read and discard
			sc.ReadToEOL()
		default:
			util.Bug("readSkill: unknown keyword %q in skill %q", word, skill.Name)
			sc.ReadToEOL()
		}
	}

	return skill
}

// readSmaugAffect reads a SMAUG affect line: 'duration' location 'modifier' bitvector
// The duration and modifier are single-quoted words. ReadWord handles quote stripping.
func readSmaugAffect(sc *Scanner) *types.SmaugAff {
	aff := &types.SmaugAff{}

	aff.Duration = sc.ReadWord()
	aff.Location = sc.ReadNumber()
	aff.Modifier = sc.ReadWord()
	aff.BitVector = sc.ReadNumber()

	return aff
}

// parseSkillType converts a type name to the SKILL_* constant.
func parseSkillType(name string) int {
	switch strings.ToLower(name) {
	case "spell":
		return types.SKILL_SPELL
	case "skill":
		return types.SKILL_SKILL
	case "weapon":
		return types.SKILL_WEAPON
	case "tongue":
		return types.SKILL_TONGUE
	case "herb":
		return types.SKILL_HERB
	case "racial":
		return types.SKILL_RACIAL
	case "disease":
		return types.SKILL_DISEASE
	default:
		return types.SKILL_UNKNOWN
	}
}

// convertMinPos converts legacy position values to modern ones.
func convertMinPos(pos int) int {
	if pos >= 100 {
		return pos - 100
	}
	// Legacy conversion table (from C fread_skill)
	switch pos {
	case 5:
		return 6
	case 6:
		return 8
	case 7:
		return 9
	case 8:
		return 12
	case 9:
		return 13
	case 10:
		return 14
	case 11:
		return 15
	default:
		return pos
	}
}
