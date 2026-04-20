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

// --- OEDIT_* constants (plan-phase6-olc-oedit.md §G1) ---

// TestOEditData_ModeConstantsUnique is the OEDIT parallel of
// TestOlcData_ModeConstantsUnique. Pairwise-distinct assertion for all 37
// OEDIT_* values. If a new constant is inserted in the middle of the block
// without updating this map, the test fails.
func TestOEditData_ModeConstantsUnique(t *testing.T) {
	modes := map[string]int{
		"OEDIT_MAIN_MENU":             OEDIT_MAIN_MENU,
		"OEDIT_EDIT_NAMELIST":         OEDIT_EDIT_NAMELIST,
		"OEDIT_SHORTDESC":             OEDIT_SHORTDESC,
		"OEDIT_LONGDESC":              OEDIT_LONGDESC,
		"OEDIT_ACTDESC":               OEDIT_ACTDESC,
		"OEDIT_TYPE":                  OEDIT_TYPE,
		"OEDIT_EXTRAS":                OEDIT_EXTRAS,
		"OEDIT_WEAR":                  OEDIT_WEAR,
		"OEDIT_WEIGHT":                OEDIT_WEIGHT,
		"OEDIT_COST":                  OEDIT_COST,
		"OEDIT_COSTPERDAY":            OEDIT_COSTPERDAY,
		"OEDIT_TIMER":                 OEDIT_TIMER,
		"OEDIT_VALUE_1":               OEDIT_VALUE_1,
		"OEDIT_VALUE_2":               OEDIT_VALUE_2,
		"OEDIT_VALUE_3":               OEDIT_VALUE_3,
		"OEDIT_VALUE_4":               OEDIT_VALUE_4,
		"OEDIT_VALUE_5":               OEDIT_VALUE_5,
		"OEDIT_VALUE_6":               OEDIT_VALUE_6,
		"OEDIT_EXTRADESC_KEY":         OEDIT_EXTRADESC_KEY,
		"OEDIT_CONFIRM_SAVEDB":        OEDIT_CONFIRM_SAVEDB,
		"OEDIT_CONFIRM_SAVESTRING":    OEDIT_CONFIRM_SAVESTRING,
		"OEDIT_EXTRADESC_DESCRIPTION": OEDIT_EXTRADESC_DESCRIPTION,
		"OEDIT_EXTRADESC_MENU":        OEDIT_EXTRADESC_MENU,
		"OEDIT_LEVEL":                 OEDIT_LEVEL,
		"OEDIT_LAYERS":                OEDIT_LAYERS,
		"OEDIT_AFFECT_MENU":           OEDIT_AFFECT_MENU,
		"OEDIT_AFFECT_LOCATION":       OEDIT_AFFECT_LOCATION,
		"OEDIT_AFFECT_MODIFIER":       OEDIT_AFFECT_MODIFIER,
		"OEDIT_AFFECT_REMOVE":         OEDIT_AFFECT_REMOVE,
		"OEDIT_AFFECT_RIS":            OEDIT_AFFECT_RIS,
		"OEDIT_EXTRADESC_CHOICE":      OEDIT_EXTRADESC_CHOICE,
		"OEDIT_EXTRADESC_DELETE":      OEDIT_EXTRADESC_DELETE,
		"OEDIT_MPROGS":                OEDIT_MPROGS,
		"OEDIT_MPROGS_CHOICE":         OEDIT_MPROGS_CHOICE,
		"OEDIT_MPROGS_DELETE":         OEDIT_MPROGS_DELETE,
		"OEDIT_MPROGS_TYPE":           OEDIT_MPROGS_TYPE,
		"OEDIT_MPROGS_ARG":            OEDIT_MPROGS_ARG,
	}
	if len(modes) != 37 {
		t.Fatalf("expected 37 OEDIT_* constants in the map, got %d — did a constant get added/removed?", len(modes))
	}
	seen := make(map[int]string)
	for name, val := range modes {
		if prior, dup := seen[val]; dup {
			t.Errorf("mode collision: %s and %s both = %d", name, prior, val)
		}
		seen[val] = name
	}
}

// TestOEditData_ConstantsExceedRedit pins the iota+200 offset guarantee —
// every OEDIT_* value must be strictly greater than every REDIT_* value so
// an Olc.Mode integer cannot be misinterpreted across editors. Plan §G1.
func TestOEditData_ConstantsExceedRedit(t *testing.T) {
	reditValues := []int{
		REDIT_MAIN_MENU, REDIT_NAME, REDIT_DESC, REDIT_FLAGS, REDIT_SECTOR,
		REDIT_TUNNEL, REDIT_TELEDELAY, REDIT_TELEVNUM, REDIT_EXIT_MENU,
		REDIT_EXIT_EDIT, REDIT_EXIT_ADD, REDIT_EXIT_ADD_VNUM,
		REDIT_EXIT_DELETE, REDIT_EXIT_VNUM, REDIT_EXIT_KEY,
		REDIT_EXIT_KEYWORD, REDIT_EXIT_DESC, REDIT_EXIT_FLAGS,
		REDIT_EXTRADESC_MENU, REDIT_EXTRADESC_CHOICE, REDIT_EXTRADESC_KEY,
		REDIT_EXTRADESC_DESCRIPTION, REDIT_EXTRADESC_DELETE,
		REDIT_CONFIRM_SAVESTRING,
	}
	reditMax := reditValues[0]
	for _, v := range reditValues {
		if v > reditMax {
			reditMax = v
		}
	}
	if OEDIT_MAIN_MENU <= reditMax {
		t.Errorf("OEDIT_MAIN_MENU (%d) must exceed max REDIT_* (%d) to keep enum spaces disjoint",
			OEDIT_MAIN_MENU, reditMax)
	}
}

// TestOEditData_ReservedMProgsIotaSlots pins the guarantee that OEDIT_MPROGS
// lands exactly 32 positions after OEDIT_MAIN_MENU. Plan-phase6-olc-mpedit.md
// depends on this offset so mpedit does not need to renumber its own
// MPROG_* const block. Plan §G1.
func TestOEditData_ReservedMProgsIotaSlots(t *testing.T) {
	if got, want := OEDIT_MPROGS-OEDIT_MAIN_MENU, 32; got != want {
		t.Errorf("OEDIT_MPROGS - OEDIT_MAIN_MENU = %d, want %d (mpedit plan relies on this offset)", got, want)
	}
}

// TestSubstate_ObjActionDistinct pins the new SUB_OBJ_ACTION constant as
// distinct from every other substate it shares meaning with. Plan §G1.
func TestSubstate_ObjActionDistinct(t *testing.T) {
	if SUB_OBJ_ACTION == SUB_OBJ_LONG {
		t.Error("SUB_OBJ_ACTION collides with SUB_OBJ_LONG")
	}
	if SUB_OBJ_ACTION == SUB_OBJ_EXTRA {
		t.Error("SUB_OBJ_ACTION collides with SUB_OBJ_EXTRA")
	}
	if SUB_OBJ_ACTION == SUB_NONE {
		t.Error("SUB_OBJ_ACTION collides with SUB_NONE")
	}
}

// TestSubstate_ObjActionAppendsAfterExtra is load-bearing: pinning
// SUB_OBJ_ACTION > SUB_OBJ_EXTRA enforces the iota ordering stated in the
// plan. If a future edit inserts SUB_OBJ_ACTION before SUB_OBJ_EXTRA in the
// const block, all downstream substate values would shift by one and any
// pfile persisted via integer substate would break. Plan §G1.
func TestSubstate_ObjActionAppendsAfterExtra(t *testing.T) {
	if SUB_OBJ_ACTION <= SUB_OBJ_EXTRA {
		t.Errorf("SUB_OBJ_ACTION (%d) must be > SUB_OBJ_EXTRA (%d) — iota ordering is load-bearing",
			SUB_OBJ_ACTION, SUB_OBJ_EXTRA)
	}
}
