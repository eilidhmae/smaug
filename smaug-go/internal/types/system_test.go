package types

import "testing"

// TestSystemData_HotbootInProgressDefault pins the zero-value default.
// HotbootInProgress is a transient in-memory flag; a fresh SystemData literal
// must report false. The flag is NOT persisted to sysdata.dat.
func TestSystemData_HotbootInProgressDefault(t *testing.T) {
	sd := SystemData{}
	if sd.HotbootInProgress {
		t.Fatalf("SystemData{}.HotbootInProgress = true, want false")
	}
}

// TestSystemData_HotbootInProgressAssignable pins that the field accepts
// bool writes and round-trips (G3 DoHotboot sets, G4 BootRecover clears).
func TestSystemData_HotbootInProgressAssignable(t *testing.T) {
	sd := SystemData{}

	sd.HotbootInProgress = true
	if !sd.HotbootInProgress {
		t.Fatalf("after sd.HotbootInProgress = true, got false")
	}

	sd.HotbootInProgress = false
	if sd.HotbootInProgress {
		t.Fatalf("after sd.HotbootInProgress = false, got true")
	}
}
