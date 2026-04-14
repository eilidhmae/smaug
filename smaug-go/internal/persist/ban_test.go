package persist

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func TestBanRoundTrip(t *testing.T) {
	w := world.New("/tmp/test")
	w.Bans = []*types.BanData{
		{
			Name:   "evil.com",
			BanBy:  "Admin",
			BanTime: "2026-04-01",
		},
		{
			Name:   "hacker",
			BanBy:  "Imm",
			BanTime: "2026-03-15",
			Prefix: true,
		},
		{
			Name:   "spam.net",
			BanBy:  "Admin",
			BanTime: "2026-02-20",
			Suffix: true,
		},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "ban.lst")

	err := SaveBanList(w, path)
	if err != nil {
		t.Fatalf("SaveBanList failed: %v", err)
	}

	w2 := world.New("/tmp/test")
	err = LoadBanList(w2, path)
	if err != nil {
		t.Fatalf("LoadBanList failed: %v", err)
	}

	if len(w2.Bans) != 3 {
		t.Fatalf("expected 3 bans, got %d", len(w2.Bans))
	}

	// Check first ban (exact)
	b := w2.Bans[0]
	if b.Name != "evil.com" {
		t.Errorf("ban[0].Name = %q, want 'evil.com'", b.Name)
	}
	if b.BanBy != "Admin" {
		t.Errorf("ban[0].BanBy = %q, want 'Admin'", b.BanBy)
	}
	if b.BanTime != "2026-04-01" {
		t.Errorf("ban[0].BanTime = %q, want '2026-04-01'", b.BanTime)
	}
	if b.Prefix || b.Suffix {
		t.Errorf("ban[0] should not have prefix or suffix")
	}

	// Check second ban (prefix)
	b = w2.Bans[1]
	if b.Name != "hacker" {
		t.Errorf("ban[1].Name = %q, want 'hacker'", b.Name)
	}
	if !b.Prefix {
		t.Errorf("ban[1].Prefix should be true")
	}
	if b.Suffix {
		t.Errorf("ban[1].Suffix should be false")
	}

	// Check third ban (suffix)
	b = w2.Bans[2]
	if b.Name != "spam.net" {
		t.Errorf("ban[2].Name = %q, want 'spam.net'", b.Name)
	}
	if !b.Suffix {
		t.Errorf("ban[2].Suffix should be true")
	}
	if b.Prefix {
		t.Errorf("ban[2].Prefix should be false")
	}
}

func TestBanLoadEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ban.lst")
	os.WriteFile(path, []byte{}, 0644)

	w := world.New("/tmp/test")
	err := LoadBanList(w, path)
	if err != nil {
		t.Fatalf("LoadBanList on empty file should not error: %v", err)
	}
	if len(w.Bans) != 0 {
		t.Errorf("expected 0 bans from empty file, got %d", len(w.Bans))
	}
}

func TestBanLoadFileNotFound(t *testing.T) {
	w := world.New("/tmp/test")
	err := LoadBanList(w, "/tmp/nonexistent_ban_file_test.lst")
	// File not found is not an error — just means no bans.
	if err != nil {
		t.Fatalf("LoadBanList on missing file should not error: %v", err)
	}
	if len(w.Bans) != 0 {
		t.Errorf("expected 0 bans from missing file, got %d", len(w.Bans))
	}
}

// --- Tier 2: class/race ban persistence ---

func TestBanRoundTrip_ClassRaceTypes(t *testing.T) {
	w := world.New("/tmp/test")
	w.Bans = []*types.BanData{
		{Name: "mage", BanBy: "Admin", BanTime: "2026-04-01", Type: types.BAN_CLASS, Level: 5},
		{Name: "troll", BanBy: "Admin", BanTime: "2026-04-02", Type: types.BAN_RACE, Level: 50},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "ban.lst")

	if err := SaveBanList(w, path); err != nil {
		t.Fatalf("SaveBanList failed: %v", err)
	}

	w2 := world.New("/tmp/test")
	if err := LoadBanList(w2, path); err != nil {
		t.Fatalf("LoadBanList failed: %v", err)
	}
	if len(w2.Bans) != 2 {
		t.Fatalf("expected 2 bans, got %d", len(w2.Bans))
	}
	if w2.Bans[0].BanType() != types.BAN_CLASS || w2.Bans[0].Level != 5 {
		t.Errorf("class ban round-trip lost data: %+v", w2.Bans[0])
	}
	if w2.Bans[1].BanType() != types.BAN_RACE || w2.Bans[1].Level != 50 {
		t.Errorf("race ban round-trip lost data: %+v", w2.Bans[1])
	}
}
