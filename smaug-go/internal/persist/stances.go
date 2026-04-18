package persist

import (
	"os"
	"strings"

	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// GetStanceNumber resolves a stance name (case-insensitive) to its
// STANCE_* index, or -1 when the name is unknown. Mirrors C's
// get_stance_number at src/stances.c:199-232.
func GetStanceNumber(name string) int {
	name = strings.TrimSpace(name)
	// C `one_argument` takes only the first word; our call sites don't
	// pass multi-word strings, but mirror the semantics for safety.
	if i := strings.IndexAny(name, " \t"); i >= 0 {
		name = name[:i]
	}
	switch strings.ToLower(name) {
	case "tiger":
		return types.STANCE_TIGER
	case "swallow":
		return types.STANCE_SWALLOW
	case "dragon":
		return types.STANCE_DRAGON
	case "monkey":
		return types.STANCE_MONKEY
	case "mantis":
		return types.STANCE_MANTIS
	case "viper":
		return types.STANCE_VIPER
	case "crane":
		return types.STANCE_CRANE
	case "crab":
		return types.STANCE_CRAB
	case "mongoose":
		return types.STANCE_MONGOOSE
	case "bull":
		return types.STANCE_BULL
	case "none":
		return types.STANCE_NONE
	case "normal":
		return types.STANCE_NORMAL
	}
	return -1
}

// LoadStancesInto reads a SMAUG `stances.dat` file and overwrites the
// combat-relevant fields (NumAttacks, DamDone, DamTaken) of entries in
// target for every StartStance…EndStance block whose name resolves via
// GetStanceNumber. Unmentioned stances keep whatever target already held.
//
// Non-combat keys (Class, Immune, Resist, Suscept, Dodge, Dual, Parry,
// Percent, Race, Stance, Special, Other, Self, Wait, Weight) are
// consumed so the scanner advances past them; their payloads are
// discarded pending the Phase-6 `StanceInfo` extension that will store
// them (see plan-tranche-b.md G1.3).
//
// Behaviour when the file is absent: returns nil without mutating
// target. The shipped stub file contains just "End\n" and is likewise
// a no-op. A nil target is also a silent no-op. Malformed blocks
// (unknown stance name, unknown key) log a BUG line via util.Bug and
// skip — matching the C `fread_stance` path.
//
// Mirrors C `load_stances` at src/stances.c:311-382.
func LoadStancesInto(target *[types.MAX_STANCE]combat.StanceInfo, path string) error {
	if target == nil {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		util.Bug("LoadStancesInto: open %s: %v", path, err)
		return nil
	}
	defer f.Close()

	sc := NewScanner(f, path)
	for {
		word := sc.ReadWord()
		if word == "" {
			break
		}
		switch word {
		case "End":
			return nil
		case "StartStance":
			readStanceBlock(sc, target)
		default:
			util.Bug("LoadStancesInto: %s: unexpected word %q", path, word)
		}
	}
	return nil
}

// readStanceBlock reads one StartStance…EndStance block. The next word
// after "StartStance" is the stance name; subsequent key/value pairs
// run until "EndStance" (or "End" defensively). Invalid stance names
// drain the block's keys without mutating target.
func readStanceBlock(sc *Scanner, target *[types.MAX_STANCE]combat.StanceInfo) {
	name := sc.ReadWord()
	idx := GetStanceNumber(name)
	if idx < 0 || idx >= types.MAX_STANCE {
		util.Bug("LoadStancesInto: bad stance name %q", name)
		drainStanceBlock(sc)
		return
	}
	for {
		word := sc.ReadWord()
		if word == "" {
			return
		}
		switch word {
		case "EndStance":
			return
		case "End":
			return
		case "Attacks":
			target[idx].NumAttacks = sc.ReadNumber()
		case "DamDone":
			target[idx].DamDone = sc.ReadNumber()
		case "DamTaken":
			target[idx].DamTaken = sc.ReadNumber()
		// Known but not-yet-stored keys — consume payload so the
		// scanner advances correctly. See G1.3 in plan-tranche-b.md.
		case "Class", "Dodge", "Dual", "Immune", "Parry", "Percent",
			"Race", "Resist", "Suscept", "Wait", "Weight":
			_ = sc.ReadNumber()
		case "Other", "Self":
			_ = sc.ReadString()
		case "Special":
			_ = sc.ReadWord()
		case "Stance":
			// Two word payload (prerequisite pair).
			_ = sc.ReadWord()
			_ = sc.ReadWord()
		default:
			util.Bug("LoadStancesInto: unknown key %q in %s block", word, name)
		}
	}
}

// drainStanceBlock consumes tokens until an EndStance/End keyword is seen.
// Used when we cannot resolve the stance name; we still have to keep the
// scanner in sync with the file.
func drainStanceBlock(sc *Scanner) {
	for {
		word := sc.ReadWord()
		if word == "" || word == "EndStance" || word == "End" {
			return
		}
		// Skip this key's argument — best-effort. Most payloads are one
		// token; for the multi-token ones (Stance, Other, Self) the
		// scanner's read-to-tilde / single-word behaviour diverges.
		// Since the block is being discarded anyway, a safe no-op here
		// is to read-and-discard the rest of the line.
		_ = sc.ReadToEOL()
	}
}
