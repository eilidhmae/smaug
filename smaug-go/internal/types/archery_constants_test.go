package types

import "testing"

// TestArcheryConstants_WearLodgeSlots pins the three WEAR_LODGE_* slots
// added in Phase-6 archery. Values are consecutive 26/27/28; MAX_WEAR
// bumps to 29. See plan-phase6-archery.md G1-1.
func TestArcheryConstants_WearLodgeSlots(t *testing.T) {
	cases := []struct {
		name string
		got  int
		want int
	}{
		{"WEAR_LODGE_RIB", WEAR_LODGE_RIB, 26},
		{"WEAR_LODGE_ARM", WEAR_LODGE_ARM, 27},
		{"WEAR_LODGE_LEG", WEAR_LODGE_LEG, 28},
		{"MAX_WEAR", MAX_WEAR, 29},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s = %d, want %d", tc.name, tc.got, tc.want)
		}
	}
	if WEAR_LODGE_RIB >= MAX_WEAR {
		t.Errorf("WEAR_LODGE_RIB (%d) must be < MAX_WEAR (%d)", WEAR_LODGE_RIB, MAX_WEAR)
	}
	if WEAR_LODGE_LEG >= MAX_WEAR {
		t.Errorf("WEAR_LODGE_LEG (%d) must be < MAX_WEAR (%d)", WEAR_LODGE_LEG, MAX_WEAR)
	}
}

// TestArcheryConstants_ItemLodgeWearFlags pins the three ITEM_LODGE_*
// wear-flag bitmask constants. Bits 22/23/24 match C mud.h:2214-2216
// (BV22/BV23/BV24). ITEM_WEAR_MAX bumps to 24 to stay in sync. G1-2.
func TestArcheryConstants_ItemLodgeWearFlags(t *testing.T) {
	cases := []struct {
		name string
		got  uint32
		want uint32
	}{
		{"ITEM_LODGE_RIB", ITEM_LODGE_RIB, 1 << 22},
		{"ITEM_LODGE_ARM", ITEM_LODGE_ARM, 1 << 23},
		{"ITEM_LODGE_LEG", ITEM_LODGE_LEG, 1 << 24},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s = 0x%x, want 0x%x", tc.name, tc.got, tc.want)
		}
	}
	if ITEM_WEAR_MAX != 24 {
		t.Errorf("ITEM_WEAR_MAX = %d, want 24 (highest bit in use after archery)", ITEM_WEAR_MAX)
	}
	// No collision with existing ITEM_WEAR_ANKLE (bit 21).
	if ITEM_LODGE_RIB == ITEM_WEAR_ANKLE {
		t.Errorf("ITEM_LODGE_RIB collides with ITEM_WEAR_ANKLE (0x%x)", ITEM_LODGE_RIB)
	}
}

// TestArcheryConstants_ItemLodgedExtraFlag pins ITEM_LODGED as the
// final extra-flag slot before MAX_ITEM_FLAG. C mud.h:2135. G1-3.
func TestArcheryConstants_ItemLodgedExtraFlag(t *testing.T) {
	if ITEM_LODGED == 0 {
		t.Fatal("ITEM_LODGED == 0 — missing or placed at ITEM_GLOW's slot")
	}
	if ITEM_LODGED >= MAX_ITEM_FLAG {
		t.Errorf("ITEM_LODGED (%d) must be < MAX_ITEM_FLAG (%d)", ITEM_LODGED, MAX_ITEM_FLAG)
	}
}

// TestArcheryConstants_ProjectileKinds pins PROJ_BOLT/ARROW/DART/STONE
// with values matching C mud.h:3068 enum ordering: BOLT=0, ARROW=1,
// DART=2, STONE=3. G1-4.
func TestArcheryConstants_ProjectileKinds(t *testing.T) {
	cases := []struct {
		name string
		got  int
		want int
	}{
		{"PROJ_BOLT", PROJ_BOLT, 0},
		{"PROJ_ARROW", PROJ_ARROW, 1},
		{"PROJ_DART", PROJ_DART, 2},
		{"PROJ_STONE", PROJ_STONE, 3},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s = %d, want %d (C mud.h:3068)", tc.name, tc.got, tc.want)
		}
	}
}
