package types

import "testing"

// ---------------------------------------------------------------------------
// GetExit
// ---------------------------------------------------------------------------

func TestRoomIndexData_GetExit(t *testing.T) {
	north := &ExitData{Direction: DIR_NORTH, Keyword: "north"}
	south := &ExitData{Direction: DIR_SOUTH, Keyword: "south"}
	up := &ExitData{Direction: DIR_UP, Keyword: "up"}

	room := &RoomIndexData{
		Exits: []*ExitData{north, south, up},
	}

	tests := []struct {
		name string
		dir  int
		want *ExitData
	}{
		{"north exists", DIR_NORTH, north},
		{"south exists", DIR_SOUTH, south},
		{"up exists", DIR_UP, up},
		{"east missing", DIR_EAST, nil},
		{"west missing", DIR_WEST, nil},
		{"down missing", DIR_DOWN, nil},
		{"northeast missing", DIR_NORTHEAST, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := room.GetExit(tc.dir)
			if got != tc.want {
				t.Errorf("GetExit(%d) = %v, want %v", tc.dir, got, tc.want)
			}
		})
	}
}

func TestRoomIndexData_GetExit_NoExits(t *testing.T) {
	room := &RoomIndexData{}
	if got := room.GetExit(DIR_NORTH); got != nil {
		t.Errorf("GetExit on empty room = %v, want nil", got)
	}
}

func TestRoomIndexData_GetExit_MultipleExitsSameDirection(t *testing.T) {
	// If somehow two exits share a direction, first one wins.
	first := &ExitData{Direction: DIR_NORTH, Keyword: "first"}
	second := &ExitData{Direction: DIR_NORTH, Keyword: "second"}

	room := &RoomIndexData{
		Exits: []*ExitData{first, second},
	}

	got := room.GetExit(DIR_NORTH)
	if got != first {
		t.Errorf("GetExit should return first matching exit, got keyword=%q", got.Keyword)
	}
}
