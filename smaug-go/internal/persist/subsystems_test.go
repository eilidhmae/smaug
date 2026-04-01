package persist

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eilidhmae/smaug/internal/world"
)

func TestLoadClansFromDir(t *testing.T) {
	w := world.New("../../../db")
	clanDir := filepath.Join(w.DataDir, "clans")

	if _, err := os.Stat(clanDir); os.IsNotExist(err) {
		t.Skip("clans directory not found")
	}

	err := LoadClansFromDir(w, clanDir)
	if err != nil {
		t.Fatalf("LoadClansFromDir: %v", err)
	}

	if len(w.Clans) == 0 {
		t.Error("expected at least one clan to be loaded")
	}

	// Verify a known clan
	found := false
	for _, c := range w.Clans {
		if c.Name != "" {
			found = true
			t.Logf("Loaded clan: %s (type=%d, members=%d)", c.Name, c.ClanType, c.Members)
		}
	}
	if !found {
		t.Error("no clans with names loaded")
	}
}

func TestLoadDeitiesFromDir(t *testing.T) {
	w := world.New("../../../db")
	deityDir := filepath.Join(w.DataDir, "deity")

	if _, err := os.Stat(deityDir); os.IsNotExist(err) {
		t.Skip("deity directory not found")
	}

	err := LoadDeitiesFromDir(w, deityDir)
	if err != nil {
		t.Fatalf("LoadDeitiesFromDir: %v", err)
	}

	if len(w.Deities) == 0 {
		t.Error("expected at least one deity to be loaded")
	}

	for _, d := range w.Deities {
		t.Logf("Loaded deity: %s (alignment=%d, worshippers=%d)", d.Name, d.Alignment, d.Worshippers)
	}
}

func TestLoadSocials(t *testing.T) {
	w := world.New("../../../db")
	socialsPath := filepath.Join(w.DataDir, "system", "en", "socials.dat")

	if _, err := os.Stat(socialsPath); os.IsNotExist(err) {
		t.Skip("socials.dat not found")
	}

	err := LoadSocials(w, socialsPath)
	if err != nil {
		t.Fatalf("LoadSocials: %v", err)
	}

	if len(w.Socials) == 0 {
		t.Error("expected at least one social to be loaded")
	}

	// Check that first social has expected structure
	if len(w.Socials) > 0 {
		s := w.Socials[0]
		if s.Name == "" {
			t.Error("first social should have a name")
		}
		t.Logf("Loaded %d socials, first: %s", len(w.Socials), s.Name)
	}
}

func TestLoadBoards(t *testing.T) {
	w := world.New("../../../db")
	boardsPath := filepath.Join(w.DataDir, "boards", "boards.dat")

	if _, err := os.Stat(boardsPath); os.IsNotExist(err) {
		t.Skip("boards.dat not found")
	}

	err := LoadBoards(w, boardsPath)
	if err != nil {
		t.Fatalf("LoadBoards: %v", err)
	}

	if len(w.Boards) == 0 {
		t.Error("expected at least one board to be loaded")
	}

	for _, b := range w.Boards {
		t.Logf("Loaded board: %s (vnum=%d, read=%d, post=%d)",
			b.NoteFile, b.BoardObj, b.MinReadLevel, b.MinPostLevel)
	}
}
