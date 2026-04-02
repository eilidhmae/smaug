package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

func TestDoRest(t *testing.T) {
	tests := []struct {
		name     string
		startPos int
		affSleep bool
		fighting bool
		wantPos  int
		wantMsg  string
	}{
		{"from standing", types.POS_STANDING, false, false, types.POS_RESTING, "sprawl out"},
		{"from sitting", types.POS_SITTING, false, false, types.POS_RESTING, "sprawl out to rest"},
		{"from sleeping", types.POS_SLEEPING, false, false, types.POS_RESTING, "rouse from"},
		{"already resting", types.POS_RESTING, false, false, types.POS_RESTING, "already resting"},
		{"while fighting", types.POS_FIGHTING, false, true, types.POS_FIGHTING, "busy fighting"},
		{"sleep affected", types.POS_SLEEPING, true, false, types.POS_SLEEPING, "can't seem to wake"},
		{"mounted", types.POS_MOUNTED, false, false, types.POS_MOUNTED, "dismount"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch, client := makeTestChar("Tester")
			defer client.Close()
			ch.Position = tt.startPos
			if tt.affSleep {
				ch.AffectedBy.Set(types.AFF_SLEEP)
			}
			if tt.fighting {
				ch.Fighting = &types.FightData{Who: &types.CharData{Name: "Mob"}}
			}

			DoRest(ch, "")

			out := readOutput(ch, client)
			if ch.Position != tt.wantPos {
				t.Errorf("position = %d, want %d", ch.Position, tt.wantPos)
			}
			if !strings.Contains(strings.ToLower(out), strings.ToLower(tt.wantMsg)) {
				t.Errorf("output %q doesn't contain %q", out, tt.wantMsg)
			}
		})
	}
}

func TestDoSit(t *testing.T) {
	tests := []struct {
		name     string
		startPos int
		affSleep bool
		wantPos  int
		wantMsg  string
	}{
		{"from standing", types.POS_STANDING, false, types.POS_SITTING, "sit down"},
		{"from resting", types.POS_RESTING, false, types.POS_SITTING, "stop resting"},
		{"from sleeping", types.POS_SLEEPING, false, types.POS_SITTING, "wake and sit"},
		{"already sitting", types.POS_SITTING, false, types.POS_SITTING, "already sitting"},
		{"while fighting", types.POS_FIGHTING, false, types.POS_FIGHTING, "busy fighting"},
		{"sleep affected", types.POS_SLEEPING, true, types.POS_SLEEPING, "can't seem to wake"},
		{"mounted", types.POS_MOUNTED, false, types.POS_MOUNTED, "already sitting"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch, client := makeTestChar("Tester")
			defer client.Close()
			ch.Position = tt.startPos
			if tt.affSleep {
				ch.AffectedBy.Set(types.AFF_SLEEP)
			}

			DoSit(ch, "")

			out := readOutput(ch, client)
			if ch.Position != tt.wantPos {
				t.Errorf("position = %d, want %d", ch.Position, tt.wantPos)
			}
			if !strings.Contains(strings.ToLower(out), strings.ToLower(tt.wantMsg)) {
				t.Errorf("output %q doesn't contain %q", out, tt.wantMsg)
			}
		})
	}
}

func TestDoStand(t *testing.T) {
	tests := []struct {
		name     string
		startPos int
		affSleep bool
		wantPos  int
		wantMsg  string
	}{
		{"from sleeping", types.POS_SLEEPING, false, types.POS_STANDING, "wake and climb"},
		{"from resting", types.POS_RESTING, false, types.POS_STANDING, "stand up"},
		{"from sitting", types.POS_SITTING, false, types.POS_STANDING, "move quickly to your feet"},
		{"already standing", types.POS_STANDING, false, types.POS_STANDING, "already standing"},
		{"while fighting", types.POS_FIGHTING, false, types.POS_FIGHTING, "already fighting"},
		{"sleep affected", types.POS_SLEEPING, true, types.POS_SLEEPING, "can't seem to wake"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch, client := makeTestChar("Tester")
			defer client.Close()
			ch.Position = tt.startPos
			if tt.affSleep {
				ch.AffectedBy.Set(types.AFF_SLEEP)
			}

			DoStand(ch, "")

			out := readOutput(ch, client)
			if ch.Position != tt.wantPos {
				t.Errorf("position = %d, want %d", ch.Position, tt.wantPos)
			}
			if !strings.Contains(strings.ToLower(out), strings.ToLower(tt.wantMsg)) {
				t.Errorf("output %q doesn't contain %q", out, tt.wantMsg)
			}
		})
	}
}

func TestDoSleep(t *testing.T) {
	tests := []struct {
		name     string
		startPos int
		wantPos  int
		wantMsg  string
	}{
		{"from standing", types.POS_STANDING, types.POS_SLEEPING, "collapse into a deep sleep"},
		{"from resting", types.POS_RESTING, types.POS_SLEEPING, "drift into slumber"},
		{"from sitting", types.POS_SITTING, types.POS_SLEEPING, "slump over"},
		{"already sleeping", types.POS_SLEEPING, types.POS_SLEEPING, "already sleeping"},
		{"while fighting", types.POS_FIGHTING, types.POS_FIGHTING, "busy fighting"},
		{"mounted", types.POS_MOUNTED, types.POS_MOUNTED, "dismount"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch, client := makeTestChar("Tester")
			defer client.Close()
			ch.Position = tt.startPos

			DoSleep(ch, "")

			out := readOutput(ch, client)
			if ch.Position != tt.wantPos {
				t.Errorf("position = %d, want %d", ch.Position, tt.wantPos)
			}
			if !strings.Contains(strings.ToLower(out), strings.ToLower(tt.wantMsg)) {
				t.Errorf("output %q doesn't contain %q", out, tt.wantMsg)
			}
		})
	}
}

func TestDoWake(t *testing.T) {
	tests := []struct {
		name     string
		startPos int
		affSleep bool
		wantPos  int
		wantMsg  string
	}{
		{"from sleeping", types.POS_SLEEPING, false, types.POS_STANDING, "wake and climb"},
		{"already standing", types.POS_STANDING, false, types.POS_STANDING, "already standing"},
		{"sleep affected", types.POS_SLEEPING, true, types.POS_SLEEPING, "can't seem to wake"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch, client := makeTestChar("Tester")
			defer client.Close()
			ch.Position = tt.startPos
			if tt.affSleep {
				ch.AffectedBy.Set(types.AFF_SLEEP)
			}

			DoWake(ch, "")

			out := readOutput(ch, client)
			if ch.Position != tt.wantPos {
				t.Errorf("position = %d, want %d", ch.Position, tt.wantPos)
			}
			if !strings.Contains(strings.ToLower(out), strings.ToLower(tt.wantMsg)) {
				t.Errorf("output %q doesn't contain %q", out, tt.wantMsg)
			}
		})
	}
}
