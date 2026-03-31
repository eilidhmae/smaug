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

// LoadClasses loads all class definitions listed in class.lst from the given directory.
// Equivalent to load_classes() in db.c.
func LoadClasses(w *world.World, classDir string) error {
	listPath := filepath.Join(classDir, "class.lst")
	if _, err := os.Stat(listPath); os.IsNotExist(err) {
		listPath = filepath.Join(classDir, "test_class.lst")
	}
	data, err := os.ReadFile(listPath)
	if err != nil {
		return fmt.Errorf("LoadClasses: cannot read %s: %w", listPath, err)
	}

	// Ensure Classes slice is initialized
	if w.Classes == nil {
		w.Classes = make([]*types.ClassType, types.MAX_CLASS)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "$" {
			break
		}
		classPath := filepath.Join(classDir, line)
		if err := loadClassFile(w, classPath); err != nil {
			util.Bug("LoadClasses: error loading %s: %v", classPath, err)
		}
	}
	return nil
}

// loadClassFile reads a single .class file and stores the result in w.Classes.
func loadClassFile(w *world.World, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("loadClassFile: cannot open %s: %w", filename, err)
	}
	defer f.Close()

	sc := NewScanner(f, filename)
	cls := &types.ClassType{}
	classIndex := -1

	for {
		word := sc.ReadWord()
		if word == "" {
			break
		}

		switch word {
		case "Name":
			cls.WhoName = sc.ReadString()
		case "Class":
			classIndex = sc.ReadNumber()
		case "AttrPrime":
			cls.AttrPrime = sc.ReadNumber()
		case "AttrSecond":
			cls.AttrSecond = sc.ReadNumber()
		case "AttrDeficient":
			cls.AttrDeficient = sc.ReadNumber()
		case "Weapon":
			cls.Weapon = sc.ReadNumber()
		case "Guild":
			cls.Guild = sc.ReadNumber()
		case "Skilladept":
			cls.SkillAdept = sc.ReadNumber()
		case "Thac0":
			cls.Thac0_00 = sc.ReadNumber()
		case "Thac32":
			cls.Thac0_32 = sc.ReadNumber()
		case "Hpmin":
			cls.HPMin = sc.ReadNumber()
		case "Hpmax":
			cls.HPMax = sc.ReadNumber()
		case "Mana":
			n := sc.ReadNumber()
			cls.FMana = n != 0
		case "Expbase":
			cls.ExpBase = sc.ReadNumber()
		case "Affected":
			cls.Affected = sc.ReadBitvector()
		case "Resist":
			cls.Resist = sc.ReadNumber()
		case "Suscept":
			cls.Suscept = sc.ReadNumber()
		case "Skill":
			// Read skill name (quoted), level, and adeptness — skip for now
			sc.ReadWord()   // skill name
			sc.ReadNumber() // level
			sc.ReadNumber() // adeptness
		case "Title":
			// Two tilde-terminated strings per title entry
			sc.ReadString()
			sc.ReadString()
		case "Login":
			cls.Login = sc.ReadString()
		case "Login_other":
			cls.LoginOther = sc.ReadString()
		case "Logout":
			cls.Logout = sc.ReadString()
		case "Logout_other":
			cls.LogoutOther = sc.ReadString()
		case "Reconnect":
			cls.Reconnect = sc.ReadString()
		case "Reconnect_other":
			cls.ReconnectOther = sc.ReadString()
		case "End":
			if classIndex < 0 || classIndex >= len(w.Classes) {
				return fmt.Errorf("loadClassFile: %s: invalid class index %d", filename, classIndex)
			}
			w.Classes[classIndex] = cls
			return nil
		default:
			util.Bug("loadClassFile: %s:%d: unknown keyword %q", filename, sc.Line(), word)
			sc.ReadToEOL()
		}
	}

	return fmt.Errorf("loadClassFile: %s: unexpected EOF before End keyword", filename)
}
