package types

import "testing"

func TestStrApp(t *testing.T) {
	tests := []struct {
		stat                 int
		wantToHit, wantToDam int
		wantCarry, wantWield int
	}{
		{0, -5, -4, 0, 0},
		{1, -5, -4, 3, 1},
		{8, 0, 0, 100, 8},
		{15, 1, 1, 170, 15},
		{18, 2, 4, 250, 25},
		{25, 10, 12, 999, 60},
	}
	for _, tt := range tests {
		got := StrApp[tt.stat]
		if got.ToHit != tt.wantToHit {
			t.Errorf("StrApp[%d].ToHit = %d, want %d", tt.stat, got.ToHit, tt.wantToHit)
		}
		if got.ToDam != tt.wantToDam {
			t.Errorf("StrApp[%d].ToDam = %d, want %d", tt.stat, got.ToDam, tt.wantToDam)
		}
		if got.Carry != tt.wantCarry {
			t.Errorf("StrApp[%d].Carry = %d, want %d", tt.stat, got.Carry, tt.wantCarry)
		}
		if got.Wield != tt.wantWield {
			t.Errorf("StrApp[%d].Wield = %d, want %d", tt.stat, got.Wield, tt.wantWield)
		}
	}
}

func TestIntApp(t *testing.T) {
	tests := []struct {
		stat, wantLearn int
	}{
		{0, 3}, {9, 15}, {18, 40}, {25, 99},
	}
	for _, tt := range tests {
		if got := IntApp[tt.stat].Learn; got != tt.wantLearn {
			t.Errorf("IntApp[%d].Learn = %d, want %d", tt.stat, got, tt.wantLearn)
		}
	}
}

func TestWisApp(t *testing.T) {
	tests := []struct {
		stat, wantPractice int
	}{
		{0, 0}, {5, 1}, {15, 3}, {25, 7},
	}
	for _, tt := range tests {
		if got := WisApp[tt.stat].Practice; got != tt.wantPractice {
			t.Errorf("WisApp[%d].Practice = %d, want %d", tt.stat, got, tt.wantPractice)
		}
	}
}

func TestDexApp(t *testing.T) {
	tests := []struct {
		stat, wantDefensive int
	}{
		{0, 60}, {7, 0}, {15, -10}, {25, -120},
	}
	for _, tt := range tests {
		if got := DexApp[tt.stat].Defensive; got != tt.wantDefensive {
			t.Errorf("DexApp[%d].Defensive = %d, want %d", tt.stat, got, tt.wantDefensive)
		}
	}
}

func TestConApp(t *testing.T) {
	tests := []struct {
		stat, wantHitp, wantShock int
	}{
		{0, -4, 20}, {7, 0, 55}, {15, 1, 90}, {25, 8, 99},
	}
	for _, tt := range tests {
		got := ConApp[tt.stat]
		if got.Hitp != tt.wantHitp {
			t.Errorf("ConApp[%d].Hitp = %d, want %d", tt.stat, got.Hitp, tt.wantHitp)
		}
		if got.Shock != tt.wantShock {
			t.Errorf("ConApp[%d].Shock = %d, want %d", tt.stat, got.Shock, tt.wantShock)
		}
	}
}

func TestChaApp(t *testing.T) {
	tests := []struct {
		stat, wantCharm int
	}{
		{0, -60}, {8, -1}, {13, 0}, {25, 99},
	}
	for _, tt := range tests {
		if got := ChaApp[tt.stat].Charm; got != tt.wantCharm {
			t.Errorf("ChaApp[%d].Charm = %d, want %d", tt.stat, got, tt.wantCharm)
		}
	}
}

func TestLckApp(t *testing.T) {
	tests := []struct {
		stat, wantLuck int
	}{
		{0, 60}, {7, 0}, {15, -10}, {25, -120},
	}
	for _, tt := range tests {
		if got := LckApp[tt.stat].Luck; got != tt.wantLuck {
			t.Errorf("LckApp[%d].Luck = %d, want %d", tt.stat, got, tt.wantLuck)
		}
	}
}

func TestAttributeTableLength(t *testing.T) {
	if len(StrApp) != 26 {
		t.Errorf("StrApp length = %d, want 26", len(StrApp))
	}
	if len(IntApp) != 26 {
		t.Errorf("IntApp length = %d, want 26", len(IntApp))
	}
	if len(WisApp) != 26 {
		t.Errorf("WisApp length = %d, want 26", len(WisApp))
	}
	if len(DexApp) != 26 {
		t.Errorf("DexApp length = %d, want 26", len(DexApp))
	}
	if len(ConApp) != 26 {
		t.Errorf("ConApp length = %d, want 26", len(ConApp))
	}
	if len(ChaApp) != 26 {
		t.Errorf("ChaApp length = %d, want 26", len(ChaApp))
	}
	if len(LckApp) != 26 {
		t.Errorf("LckApp length = %d, want 26", len(LckApp))
	}
}
