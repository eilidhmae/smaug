package persist

import (
	"path/filepath"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

func TestLoadClasses(t *testing.T) {
	w := world.New("")
	// Initialize Classes slice with enough capacity
	w.Classes = make([]*types.ClassType, types.MAX_CLASS)

	err := LoadClasses(w, filepath.Join("testdata"))
	if err != nil {
		t.Fatalf("LoadClasses failed: %v", err)
	}

	cls := w.Classes[3] // Warrior is class 3
	if cls == nil {
		t.Fatal("class 3 (Warrior) not loaded")
	}
	if cls.WhoName != "Warrior" {
		t.Errorf("WhoName = %q, want %q", cls.WhoName, "Warrior")
	}
	if cls.AttrPrime != 1 {
		t.Errorf("AttrPrime = %d, want 1", cls.AttrPrime)
	}
	if cls.AttrSecond != 0 {
		t.Errorf("AttrSecond = %d, want 0", cls.AttrSecond)
	}
	if cls.AttrDeficient != 0 {
		t.Errorf("AttrDeficient = %d, want 0", cls.AttrDeficient)
	}
	if cls.Weapon != 10313 {
		t.Errorf("Weapon = %d, want 10313", cls.Weapon)
	}
	if cls.Guild != 3022 {
		t.Errorf("Guild = %d, want 3022", cls.Guild)
	}
	if cls.SkillAdept != 85 {
		t.Errorf("SkillAdept = %d, want 85", cls.SkillAdept)
	}
	if cls.Thac0_00 != 18 {
		t.Errorf("Thac0_00 = %d, want 18", cls.Thac0_00)
	}
	if cls.Thac0_32 != 6 {
		t.Errorf("Thac0_32 = %d, want 6", cls.Thac0_32)
	}
	if cls.HPMin != 11 {
		t.Errorf("HPMin = %d, want 11", cls.HPMin)
	}
	if cls.HPMax != 15 {
		t.Errorf("HPMax = %d, want 15", cls.HPMax)
	}
	if cls.FMana != false {
		t.Errorf("FMana = %v, want false", cls.FMana)
	}
	if cls.ExpBase != 1150 {
		t.Errorf("ExpBase = %d, want 1150", cls.ExpBase)
	}
}
