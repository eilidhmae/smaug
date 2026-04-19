package types

import (
	"net"
	"testing"
)

// TestDescriptorOlc_DefaultsNil pins the invariant that a fresh descriptor
// has Olc == nil — the menu-state pointer is allocated on demand when
// DoRedit / DoOedit / DoMedit enter the menu substate, not at descriptor
// creation. Plan §A1.
func TestDescriptorOlc_DefaultsNil(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	d := NewDescriptor(server)
	if d.Olc != nil {
		t.Errorf("new descriptor Olc = %+v, want nil", d.Olc)
	}
}

// TestOlcData_AllocatesFields verifies the struct literal can name every
// field called out in the plan §A2. If a field is removed or renamed the
// compiler fails this test.
func TestOlcData_AllocatesFields(t *testing.T) {
	var room RoomIndexData
	od := &OlcData{
		Mode:    REDIT_MAIN_MENU,
		Vnum:    1234,
		Change:  true,
		Target:  &room,
		Spare:   nil,
		TempNum: 3,
	}
	if od.Mode != REDIT_MAIN_MENU {
		t.Errorf("Mode = %d, want REDIT_MAIN_MENU", od.Mode)
	}
	if od.Vnum != 1234 {
		t.Errorf("Vnum = %d, want 1234", od.Vnum)
	}
	if !od.Change {
		t.Error("Change = false, want true")
	}
	if od.Target == nil {
		t.Error("Target unexpectedly nil")
	}
	if od.Spare != nil {
		t.Errorf("Spare = %v, want nil", od.Spare)
	}
	if od.TempNum != 3 {
		t.Errorf("TempNum = %d, want 3", od.TempNum)
	}
}

// TestOlcData_ModeConstantsUnique pins the REDIT_* constants as pairwise
// distinct integers. Prevents a silent collision if someone inserts a new
// constant in the middle of the block. Plan §A3.
func TestOlcData_ModeConstantsUnique(t *testing.T) {
	modes := map[string]int{
		"REDIT_MAIN_MENU":             REDIT_MAIN_MENU,
		"REDIT_NAME":                  REDIT_NAME,
		"REDIT_DESC":                  REDIT_DESC,
		"REDIT_FLAGS":                 REDIT_FLAGS,
		"REDIT_SECTOR":                REDIT_SECTOR,
		"REDIT_TUNNEL":                REDIT_TUNNEL,
		"REDIT_TELEDELAY":             REDIT_TELEDELAY,
		"REDIT_TELEVNUM":              REDIT_TELEVNUM,
		"REDIT_EXIT_MENU":             REDIT_EXIT_MENU,
		"REDIT_EXIT_EDIT":             REDIT_EXIT_EDIT,
		"REDIT_EXIT_ADD":              REDIT_EXIT_ADD,
		"REDIT_EXIT_ADD_VNUM":         REDIT_EXIT_ADD_VNUM,
		"REDIT_EXIT_DELETE":           REDIT_EXIT_DELETE,
		"REDIT_EXIT_VNUM":             REDIT_EXIT_VNUM,
		"REDIT_EXIT_KEY":              REDIT_EXIT_KEY,
		"REDIT_EXIT_KEYWORD":          REDIT_EXIT_KEYWORD,
		"REDIT_EXIT_DESC":             REDIT_EXIT_DESC,
		"REDIT_EXIT_FLAGS":            REDIT_EXIT_FLAGS,
		"REDIT_EXTRADESC_MENU":        REDIT_EXTRADESC_MENU,
		"REDIT_EXTRADESC_CHOICE":      REDIT_EXTRADESC_CHOICE,
		"REDIT_EXTRADESC_KEY":         REDIT_EXTRADESC_KEY,
		"REDIT_EXTRADESC_DESCRIPTION": REDIT_EXTRADESC_DESCRIPTION,
		"REDIT_EXTRADESC_DELETE":      REDIT_EXTRADESC_DELETE,
		"REDIT_CONFIRM_SAVESTRING":    REDIT_CONFIRM_SAVESTRING,
	}
	seen := make(map[int]string)
	for name, val := range modes {
		if prior, dup := seen[val]; dup {
			t.Errorf("mode collision: %s and %s both = %d", name, prior, val)
		}
		seen[val] = name
	}
}

// TestReditModes_DistinctFromConstates pins the cross-enum gap that keeps
// REDIT_* values from colliding with CON_* values (CON_COPYOVER_RECOVER is
// the largest at iota 23). REDIT_MAIN_MENU is set to iota+100 so the two
// enum spaces can't be confused when casting to int. Plan §A3.
func TestReditModes_DistinctFromConstates(t *testing.T) {
	conMax := int(CON_COPYOVER_RECOVER)
	if REDIT_MAIN_MENU <= conMax {
		t.Errorf("REDIT_MAIN_MENU (%d) must exceed max CON_* (%d) to avoid enum confusion",
			REDIT_MAIN_MENU, conMax)
	}
}
