package persist

import (
	"path/filepath"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func TestLoadRaces(t *testing.T) {
	w := world.New("")
	w.Races = make([]*types.RaceData, types.MAX_RACE)

	err := LoadRaces(w, filepath.Join("testdata"))
	if err != nil {
		t.Fatalf("LoadRaces failed: %v", err)
	}

	race := w.Races[0] // Human is race 0
	if race == nil {
		t.Fatal("race 0 (Human) not loaded")
	}
	if race.Name != "Human" {
		t.Errorf("Name = %q, want %q", race.Name, "Human")
	}
	if race.Language != 1 {
		t.Errorf("Language = %d, want 1", race.Language)
	}
	if race.MinAlign != -1000 {
		t.Errorf("MinAlign = %d, want -1000", race.MinAlign)
	}
	if race.MaxAlign != 1000 {
		t.Errorf("MaxAlign = %d, want 1000", race.MaxAlign)
	}
	if race.ExpMultiplier != 100 {
		t.Errorf("ExpMultiplier = %d, want 100", race.ExpMultiplier)
	}
	if race.Height != 66 {
		t.Errorf("Height = %d, want 66", race.Height)
	}
	if race.Weight != 150 {
		t.Errorf("Weight = %d, want 150", race.Weight)
	}
	// WhereName
	if race.WhereName[0] != "<used as light>" {
		t.Errorf("WhereName[0] = %q, want %q", race.WhereName[0], "<used as light>")
	}
	if race.WhereName[1] != "<worn on finger>" {
		t.Errorf("WhereName[1] = %q, want %q", race.WhereName[1], "<worn on finger>")
	}
}
