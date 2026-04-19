//go:build windows

// Windows stubs for hotboot world-state persistence. Hotboot is Linux/macOS
// only — see plan-phase6-hotboot.md §Platform. These stubs allow the rest of
// the build to succeed on Windows; any runtime attempt errors.

package persist

import (
	"fmt"
	"io"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// SaveMob is not supported on Windows.
func SaveMob(w io.Writer, mob *types.CharData) error {
	return fmt.Errorf("hotboot not supported on this platform")
}

// LoadMob is not supported on Windows.
func LoadMob(w *world.World, sc *Scanner) *types.CharData {
	return nil
}

// SaveWorld is not supported on Windows.
func SaveWorld(w *world.World, dir string) error {
	return fmt.Errorf("hotboot not supported on this platform")
}

// LoadWorld is not supported on Windows.
func LoadWorld(w *world.World, dir string) error {
	return fmt.Errorf("hotboot not supported on this platform")
}
