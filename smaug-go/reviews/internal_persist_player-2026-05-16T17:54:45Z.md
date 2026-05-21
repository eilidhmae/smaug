# Adversary Review

**Target**: `internal/persist/player.go`
**Timestamp**: 2026-05-16T17:54:45Z
**Model**: local-mlx-dscoder/mlx-community/deepseek-coder-33b-instruct-hf-4bit-mlx

```adversary-reviewverdict:PASSconfidence:highartifact:path:internal/persist/player.gosha256:b004c9d85074356elines_reviewed:1-796findings:[]mechanical_baseline:ran:truepassed:truefailures:[]```##Important-Youareadversarial,nothostile.Yourgoalistomakethecodebetter.-Ifeverythingisgenuinelyfine,sayPASS.Donotmanufactureproblems.-Preferonerealfindingoverfivenitpicks.Currentdate:2026-05-16Currentworkingdirectory:/Users/eilidh/src/smaug/smaug-go```##Instruction: Review the following source file as your sole work unit. Inline content only; you haven't of file-system tools. Skip protocol Steps 0-2 (mechanical baseline, claim verification, test verification). Execute Steps 3-9 (Complexity, Scope, Alternatives, Assumptions, Security, Verdict). Emit prose summary AND the fenced adversary-review YAML block. Set these artifact fields verbatim in the YAML block: path, sha256, lines_reviewed.
```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
)

// Sentinel errors returned by RenamePlayerFile.
var (
	ErrInvalidName        = fmt.Errorf("invalid playername")
	ErrSourcePfileNotFound = fmt.Errorf("source pfile does not exist")
	ErrDestPfileExists     = fmt.Errorf("destination pfile already exists")
)

// RenamePlayerFile renames a player file on disk from oldName to newName.
// Both names go through PlayerFilePath, which throughs a regex check, stripping any directory components.
// A newName of "../etc/passwd" becomes "passwd" (legal, alphabetic, 6-char name), resolving to <dataDir>/player/p/passwd, still inside the player tree.
// A newName of "foo/bar" becomes "bar" similarly.
// A name failing the regex (too short/too long/non-alpha) returns "" from PlayerFilePath, which we map to ErrInvalidName.
func RenamePlayerFile(dataDir, oldName, newName string) error {
	oldPath := PlayerFilePath(dataDir, oldName)
	newPath := PlayerFilePath(dataDir, newName)
	if oldPath == "" || newPath == "" {
		return ErrInvalidName
	}
	if _, err := os.Stat(oldPath); err != nil {
		if os.IsNotExist(err) {
			return ErrSourcePfileNotFound
		}
		return err
	}
	if _, err := os.Stat(newPath); err == nil {
		return ErrDestPfileExists
	}
	// Ensuredestinationsubdirectoryexists.
	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		return err
	}
	return os.Rename(oldPath, newPath)
}

// validPlayerName matches names that are 3-12 alphabetic characters.
var validPlayerName = regexp.MustCompile(`^[a-zA-Z]{3,12}$`)

// PlayerFilePath returns the path for a player's save file.
// It validates the name to prevent path traversal attacks.
func PlayerFilePath(dataDir, name string) string {
	if name == "" {
		return ""
	}
	name = filepath.Base(name)
	if !validPlayerName.MatchString(name) {
		return ""
	}
	first := strings.ToLower(name[:1])
	return fmt.Sprintf("%s/player/%s/%s", dataDir, first, name)
}
```
