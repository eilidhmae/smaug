package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// --- DoBan ---

func TestDoBan_NoArgEmptyList(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeImmTestChar("Admin")
	defer client.Close()

	DoBan(ch, "")
	out := readOutput(ch, client)

	if !strings.Contains(out, "No bans") {
		t.Errorf("expected 'No bans' message, got: %q", out)
	}
}

func TestDoBan_ListEmpty(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeImmTestChar("Admin")
	defer client.Close()

	DoBan(ch, "list")
	out := readOutput(ch, client)

	if !strings.Contains(out, "No bans") {
		t.Errorf("expected 'No bans' message, got: %q", out)
	}
}

func TestDoBan_SiteExact(t *testing.T) {
	w := setupWizWorld()
	ch, client := makeImmTestChar("Admin")
	defer client.Close()

	DoBan(ch, "site evil.com")
	out := readOutput(ch, client)

	if !strings.Contains(out, "evil.com") {
		t.Errorf("expected confirmation with site name, got: %q", out)
	}
	if len(w.Bans) != 1 {
		t.Fatalf("expected 1 ban, got %d", len(w.Bans))
	}
	ban := w.Bans[0]
	if ban.Name != "evil.com" {
		t.Errorf("expected ban name 'evil.com', got %q", ban.Name)
	}
	if ban.Prefix || ban.Suffix {
		t.Errorf("expected neither prefix nor suffix for exact ban")
	}
	if ban.BanBy != "Admin" {
		t.Errorf("expected BanBy 'Admin', got %q", ban.BanBy)
	}
}

func TestDoBan_SiteSuffix(t *testing.T) {
	w := setupWizWorld()
	ch, client := makeImmTestChar("Admin")
	defer client.Close()

	DoBan(ch, "site *evil.com")
	_ = readOutput(ch, client)

	if len(w.Bans) != 1 {
		t.Fatalf("expected 1 ban, got %d", len(w.Bans))
	}
	ban := w.Bans[0]
	if ban.Name != "evil.com" {
		t.Errorf("expected ban name 'evil.com', got %q", ban.Name)
	}
	if !ban.Suffix {
		t.Errorf("expected Suffix=true")
	}
	if ban.Prefix {
		t.Errorf("expected Prefix=false")
	}
}

func TestDoBan_SitePrefix(t *testing.T) {
	w := setupWizWorld()
	ch, client := makeImmTestChar("Admin")
	defer client.Close()

	DoBan(ch, "site evil*")
	_ = readOutput(ch, client)

	if len(w.Bans) != 1 {
		t.Fatalf("expected 1 ban, got %d", len(w.Bans))
	}
	ban := w.Bans[0]
	if ban.Name != "evil" {
		t.Errorf("expected ban name 'evil', got %q", ban.Name)
	}
	if !ban.Prefix {
		t.Errorf("expected Prefix=true")
	}
	if ban.Suffix {
		t.Errorf("expected Suffix=false")
	}
}

func TestDoBan_SiteNoPattern(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeImmTestChar("Admin")
	defer client.Close()

	DoBan(ch, "site")
	out := readOutput(ch, client)

	if !strings.Contains(out, "Usage") && !strings.Contains(out, "usage") && !strings.Contains(out, "Syntax") {
		t.Errorf("expected usage message, got: %q", out)
	}
}

func TestDoBan_ListWithBans(t *testing.T) {
	w := setupWizWorld()
	w.Bans = append(w.Bans, &types.BanData{
		Name:  "evil.com",
		BanBy: "Admin",
	})

	ch, client := makeImmTestChar("Admin")
	defer client.Close()

	DoBan(ch, "list")
	out := readOutput(ch, client)

	if !strings.Contains(out, "evil.com") {
		t.Errorf("expected 'evil.com' in ban list, got: %q", out)
	}
}

func TestDoBan_RemoveExisting(t *testing.T) {
	w := setupWizWorld()
	w.Bans = append(w.Bans, &types.BanData{
		Name:  "evil.com",
		BanBy: "Admin",
	})

	ch, client := makeImmTestChar("Admin")
	defer client.Close()

	DoBan(ch, "remove evil.com")
	out := readOutput(ch, client)

	if len(w.Bans) != 0 {
		t.Errorf("expected 0 bans after removal, got %d", len(w.Bans))
	}
	if !strings.Contains(out, "removed") && !strings.Contains(out, "Removed") && !strings.Contains(out, "lifted") {
		t.Errorf("expected removal confirmation, got: %q", out)
	}
}

func TestDoBan_RemoveNotFound(t *testing.T) {
	_ = setupWizWorld()
	ch, client := makeImmTestChar("Admin")
	defer client.Close()

	DoBan(ch, "remove nonexistent.com")
	out := readOutput(ch, client)

	if !strings.Contains(out, "not found") {
		t.Errorf("expected 'not found' message, got: %q", out)
	}
}

// --- CheckBans ---

func TestCheckBans_NoBans(t *testing.T) {
	w := world.New("/tmp/test")
	result := CheckBans(w, "anything.com")
	if result != nil {
		t.Errorf("expected nil, got %+v", result)
	}
}

func TestCheckBans_ExactMatch(t *testing.T) {
	w := world.New("/tmp/test")
	ban := &types.BanData{Name: "evil.com"}
	w.Bans = append(w.Bans, ban)

	result := CheckBans(w, "evil.com")
	if result != ban {
		t.Errorf("expected matching ban, got %v", result)
	}
}

func TestCheckBans_ExactMatchCaseInsensitive(t *testing.T) {
	w := world.New("/tmp/test")
	ban := &types.BanData{Name: "Evil.COM"}
	w.Bans = append(w.Bans, ban)

	result := CheckBans(w, "evil.com")
	if result != ban {
		t.Errorf("expected matching ban (case-insensitive), got %v", result)
	}
}

func TestCheckBans_PrefixMatch(t *testing.T) {
	w := world.New("/tmp/test")
	ban := &types.BanData{Name: "evil", Prefix: true}
	w.Bans = append(w.Bans, ban)

	result := CheckBans(w, "evil.hacker.com")
	if result != ban {
		t.Errorf("expected prefix match, got %v", result)
	}
}

func TestCheckBans_SuffixMatch(t *testing.T) {
	w := world.New("/tmp/test")
	ban := &types.BanData{Name: "evil.com", Suffix: true}
	w.Bans = append(w.Bans, ban)

	result := CheckBans(w, "very.evil.com")
	if result != ban {
		t.Errorf("expected suffix match, got %v", result)
	}
}

func TestCheckBans_NoMatch(t *testing.T) {
	w := world.New("/tmp/test")
	w.Bans = append(w.Bans, &types.BanData{Name: "evil.com"})
	w.Bans = append(w.Bans, &types.BanData{Name: "bad", Prefix: true})
	w.Bans = append(w.Bans, &types.BanData{Name: "hacker.net", Suffix: true})

	result := CheckBans(w, "good.com")
	if result != nil {
		t.Errorf("expected nil for non-matching site, got %+v", result)
	}
}

func TestCheckBans_PrefixNoMatch(t *testing.T) {
	w := world.New("/tmp/test")
	w.Bans = append(w.Bans, &types.BanData{Name: "evil", Prefix: true})

	result := CheckBans(w, "notevil.com")
	if result == nil {
		// "notevil.com" does start with "evil"? No, "notevil" does not start with "evil".
		// This should be nil.
	}
	// Actually "notevil.com" does NOT start with "evil", so nil is correct.
	if result != nil {
		t.Errorf("expected nil for non-prefix match, got %+v", result)
	}
}

func TestCheckBans_SuffixNoMatch(t *testing.T) {
	w := world.New("/tmp/test")
	w.Bans = append(w.Bans, &types.BanData{Name: "evil.com", Suffix: true})

	result := CheckBans(w, "evil.com.au")
	if result != nil {
		t.Errorf("expected nil for non-suffix match, got %+v", result)
	}
}

// --- Class/race bans (G7) ---

func TestDoBan_ClassAddedWithLevel(t *testing.T) {
	w := setupWizWorld()
	ch, client := makeImmTestChar("Admin")
	defer client.Close()

	DoBan(ch, "class mage 10")
	_ = readOutput(ch, client)

	if len(w.Bans) != 1 {
		t.Fatalf("expected 1 ban, got %d", len(w.Bans))
	}
	b := w.Bans[0]
	if b.Type != types.BAN_CLASS {
		t.Errorf("expected BAN_CLASS, got %d", b.Type)
	}
	if b.Name != "mage" {
		t.Errorf("expected name 'mage', got %q", b.Name)
	}
	if b.Level != 10 {
		t.Errorf("expected level 10, got %d", b.Level)
	}
}

func TestDoBan_RaceAdded(t *testing.T) {
	w := setupWizWorld()
	ch, client := makeImmTestChar("Admin")
	defer client.Close()

	DoBan(ch, "race troll")
	_ = readOutput(ch, client)
	if len(w.Bans) != 1 || w.Bans[0].Type != types.BAN_RACE {
		t.Fatalf("expected one race ban, got %+v", w.Bans)
	}
}

func TestCheckClassBan_BelowLevel(t *testing.T) {
	w := world.New("/tmp/test")
	w.Bans = append(w.Bans, &types.BanData{
		Name: "mage", Type: types.BAN_CLASS, Level: 10,
	})
	if CheckClassBan(w, "Mage", 5) == nil {
		t.Errorf("level 5 mage should be banned when level threshold is 10")
	}
	if CheckClassBan(w, "Mage", 15) != nil {
		t.Errorf("level 15 mage should NOT be banned above threshold")
	}
}

func TestCheckRaceBan_Unconditional(t *testing.T) {
	w := world.New("/tmp/test")
	w.Bans = append(w.Bans, &types.BanData{
		Name: "troll", Type: types.BAN_RACE, Level: 0,
	})
	if CheckRaceBan(w, "Troll", 50) == nil {
		t.Errorf("troll should be banned at all levels when threshold is 0")
	}
}

func TestCheckBans_OnlyMatchesSiteType(t *testing.T) {
	w := world.New("/tmp/test")
	// A class ban named "evil.com" should NOT match a site-ban lookup.
	w.Bans = append(w.Bans, &types.BanData{
		Name: "evil.com", Type: types.BAN_CLASS,
	})
	if CheckBans(w, "evil.com") != nil {
		t.Errorf("class ban should not match site lookup")
	}
}
