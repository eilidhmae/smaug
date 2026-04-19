//go:build windows

package persist

import "fmt"

// HotbootSession is defined here so Windows builds see the type, but all
// read/write operations error.
type HotbootSession struct {
	FDIndex   int
	RoomVnum  int
	Port      int
	IdleTicks int
	Ansi      bool
	Name      string
	Host      string
}

// SaveHotbootSessions is not supported on Windows.
func SaveHotbootSessions(dir string, sessions []HotbootSession) error {
	return fmt.Errorf("hotboot not supported on this platform")
}

// LoadHotbootSessions is not supported on Windows.
func LoadHotbootSessions(dir string) ([]HotbootSession, error) {
	return nil, fmt.Errorf("hotboot not supported on this platform")
}
