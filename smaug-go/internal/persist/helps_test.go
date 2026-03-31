package persist

import (
	"os"
	"testing"

	"github.com/eilidhmae/smaug/internal/world"
)

func TestLoadHelps(t *testing.T) {
	w := world.New("testdata")

	f, err := os.Open("testdata/test_helps.are")
	if err != nil {
		t.Fatalf("cannot open test file: %v", err)
	}
	defer f.Close()

	sc := NewScanner(f, "test_helps.are")

	// Skip to #HELPS section header
	word := sc.ReadWord()
	if word != "#HELPS" {
		t.Fatalf("expected #HELPS, got %q", word)
	}

	loadHelps(sc, w)

	if len(w.Helps) != 3 {
		t.Fatalf("Helps count = %d, want 3", len(w.Helps))
	}

	// Test MOTD
	motd := w.Helps[0]
	if motd.Keyword != "MOTD" {
		t.Errorf("Helps[0].Keyword = %q, want %q", motd.Keyword, "MOTD")
	}
	if motd.Level != 0 {
		t.Errorf("MOTD.Level = %d, want 0", motd.Level)
	}
	if motd.Text == "" {
		t.Error("MOTD.Text is empty")
	}
	if motd.Text[0:7] != "Welcome" {
		t.Errorf("MOTD.Text starts with %q, want 'Welcome'", motd.Text[:7])
	}

	// Test RULES
	rules := w.Helps[1]
	if rules.Keyword != "RULES" {
		t.Errorf("Helps[1].Keyword = %q, want %q", rules.Keyword, "RULES")
	}
	if rules.Level != 1 {
		t.Errorf("RULES.Level = %d, want 1", rules.Level)
	}

	// Test IMOTD (level -1)
	imotd := w.Helps[2]
	if imotd.Keyword != "IMOTD" {
		t.Errorf("Helps[2].Keyword = %q, want %q", imotd.Keyword, "IMOTD")
	}
	if imotd.Level != -1 {
		t.Errorf("IMOTD.Level = %d, want -1", imotd.Level)
	}
}
