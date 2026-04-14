package persist

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// LoadBanList reads the ban list from the given file path.
// File format: blocks separated by ~ on its own line.
// Each block is either the legacy 5-line site block:
//
//	Site~
//	BanBy~
//	BanTime~
//	Prefix Suffix
//	~
//
// …or the extended block with trailing "Type Level" on the flags line:
//
//	Name~
//	BanBy~
//	BanTime~
//	Prefix Suffix Type Level
//	~
//
// Unknown Type values default to BAN_SITE so legacy files parse cleanly.
// If the file does not exist, returns nil with no error (no bans).
func LoadBanList(w *world.World, path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || line == "~" {
			continue
		}

		ban := &types.BanData{}

		// Line 1: Name~
		ban.Name = strings.TrimSuffix(line, "~")

		// Line 2: BanBy~
		if !scanner.Scan() {
			break
		}
		ban.BanBy = strings.TrimSuffix(scanner.Text(), "~")

		// Line 3: BanTime~
		if !scanner.Scan() {
			break
		}
		ban.BanTime = strings.TrimSuffix(scanner.Text(), "~")

		// Line 4: Prefix Suffix [Type Level] (extended columns are optional).
		if !scanner.Scan() {
			break
		}
		flags := scanner.Text()
		parts := strings.Fields(flags)
		if len(parts) >= 2 {
			ban.Prefix = parts[0] == "1"
			ban.Suffix = parts[1] == "1"
		}
		if len(parts) >= 3 {
			if v, err := strconv.Atoi(parts[2]); err == nil {
				ban.Type = v
			}
		}
		if len(parts) >= 4 {
			if v, err := strconv.Atoi(parts[3]); err == nil {
				ban.Level = v
			}
		}

		// Line 5: block terminator ~
		if scanner.Scan() {
			// consume the ~ separator
		}

		w.Bans = append(w.Bans, ban)
	}

	return scanner.Err()
}

// SaveBanList writes the ban list to the given file path using atomic write.
func SaveBanList(w *world.World, path string) error {
	dir := filepath.Dir(path)
	tmpPath := filepath.Join(dir, ".ban.tmp")

	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(f)
	for _, ban := range w.Bans {
		prefixFlag := boolToInt(ban.Prefix)
		suffixFlag := boolToInt(ban.Suffix)
		fmt.Fprintf(writer, "%s~\n", ban.Name)
		fmt.Fprintf(writer, "%s~\n", ban.BanBy)
		fmt.Fprintf(writer, "%s~\n", ban.BanTime)
		fmt.Fprintf(writer, "%d %d %d %d\n", prefixFlag, suffixFlag, ban.BanType(), ban.Level)
		fmt.Fprintf(writer, "~\n")
	}

	if err := writer.Flush(); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return os.Rename(tmpPath, path)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
