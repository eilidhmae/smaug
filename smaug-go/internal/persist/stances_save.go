package persist

import (
	"fmt"
	"io"
	"os"

	"github.com/eilidhmae/smaug/internal/combat"
	"github.com/eilidhmae/smaug/internal/types"
)

// SaveStances writes the stance table to path in C fwrite_stance format
// (src/stances.c:599-670). Each stance is emitted as a StartStance /
// EndStance block with only the fields whose values are non-default
// ("conditional-write" invariants below); the loader's defaults repopulate
// everything else on read.
//
// Conditional-write invariants (C verbatim):
//
//	Attacks     : NumAttacks     != 0
//	Class       : Class          >  0
//	DamDone     : DamDone        >  0
//	DamTaken    : DamTaken       >  0
//	Dodge       : Dodge          != 0
//	Dual        : Dual           != 0
//	Immune      : Immune         >  0
//	Other       : Others         != ""
//	Parry       : Parry          != 0
//	Percent     : SpecialPercent >  0
//	Race        : Race           >  0
//	Resist      : Resist         >  0
//	Self        : Self           != ""
//	Special     : SpecialMove    >  0
//	Stance      : Prereq[0]      >  0
//	Suscept     : Suscept        >  0
//	Wait        : Wait           >  0
//	Weight      : MaxWeight      >  0
//
// File shape:
//
//	StartStance\t<name>\n
//	<k>\t<v>\n      ... (per invariant)
//	EndStance\n
//	\n              <-- trailing blank line between stances
//	...
//	End\n
//
// Tabs are load-bearing for C parity — the scanner treats tab as
// whitespace, so round-tripping with spaces would still load, but the
// on-disk file must match C's format.
func SaveStances(target *[types.MAX_STANCE]combat.StanceInfo, path string) error {
	if target == nil {
		return fmt.Errorf("SaveStances: nil target")
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("SaveStances: create %s: %w", path, err)
	}
	defer f.Close()
	if err := writeStances(f, target); err != nil {
		return err
	}
	return nil
}

// writeStances is the io.Writer-facing core, used by SaveStances and by
// tests that want to capture output into a bytes.Buffer for byte-exact
// comparisons.
func writeStances(w io.Writer, target *[types.MAX_STANCE]combat.StanceInfo) error {
	for i := 0; i < types.MAX_STANCE; i++ {
		info := target[i]
		name := combat.GetStanceName(i)
		if name == "" {
			name = "None"
		}
		if _, err := fmt.Fprintf(w, "StartStance\t%s\n", name); err != nil {
			return err
		}
		if info.NumAttacks != 0 {
			if _, err := fmt.Fprintf(w, "Attacks\t%d\n", info.NumAttacks); err != nil {
				return err
			}
		}
		if info.Class > 0 {
			if _, err := fmt.Fprintf(w, "Class\t%d\n", info.Class); err != nil {
				return err
			}
		}
		if info.DamDone > 0 {
			if _, err := fmt.Fprintf(w, "DamDone\t%d\n", info.DamDone); err != nil {
				return err
			}
		}
		if info.DamTaken > 0 {
			if _, err := fmt.Fprintf(w, "DamTaken\t%d\n", info.DamTaken); err != nil {
				return err
			}
		}
		if info.Dodge != 0 {
			if _, err := fmt.Fprintf(w, "Dodge\t%d\n", info.Dodge); err != nil {
				return err
			}
		}
		if info.Dual != 0 {
			if _, err := fmt.Fprintf(w, "Dual\t%d\n", info.Dual); err != nil {
				return err
			}
		}
		if info.Immune > 0 {
			if _, err := fmt.Fprintf(w, "Immune\t%d\n", info.Immune); err != nil {
				return err
			}
		}
		if info.Others != "" {
			if _, err := fmt.Fprintf(w, "Other\t%s~\n", info.Others); err != nil {
				return err
			}
		}
		if info.Parry != 0 {
			if _, err := fmt.Fprintf(w, "Parry\t%d\n", info.Parry); err != nil {
				return err
			}
		}
		if info.SpecialPercent > 0 {
			if _, err := fmt.Fprintf(w, "Percent\t%d\n", info.SpecialPercent); err != nil {
				return err
			}
		}
		if info.Race > 0 {
			if _, err := fmt.Fprintf(w, "Race\t%d\n", info.Race); err != nil {
				return err
			}
		}
		if info.Resist > 0 {
			if _, err := fmt.Fprintf(w, "Resist\t%d\n", info.Resist); err != nil {
				return err
			}
		}
		if info.Self != "" {
			if _, err := fmt.Fprintf(w, "Self\t%s~\n", info.Self); err != nil {
				return err
			}
		}
		if info.SpecialMove > 0 {
			if _, err := fmt.Fprintf(w, "Special\t%s\n", combat.GetSpecialName(info.SpecialMove)); err != nil {
				return err
			}
		}
		if info.Prereq[0] > 0 {
			// C passes Prereq[1] (which may be 0) to get_stance_name; 0
			// returns "None" in GetStanceName, matching C.
			if _, err := fmt.Fprintf(w, "Stance\t%s %s\n",
				combat.GetStanceName(info.Prereq[0]),
				combat.GetStanceName(info.Prereq[1]),
			); err != nil {
				return err
			}
		}
		if info.Suscept > 0 {
			if _, err := fmt.Fprintf(w, "Suscept\t%d\n", info.Suscept); err != nil {
				return err
			}
		}
		if info.Wait > 0 {
			if _, err := fmt.Fprintf(w, "Wait\t%d\n", info.Wait); err != nil {
				return err
			}
		}
		if info.MaxWeight > 0 {
			if _, err := fmt.Fprintf(w, "Weight\t%d\n", info.MaxWeight); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprint(w, "EndStance\n\n"); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprint(w, "End\n"); err != nil {
		return err
	}
	return nil
}
