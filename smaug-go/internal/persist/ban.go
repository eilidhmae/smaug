package persist

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// LoadBanList reads the ban list from the given file path.
// File format: blocks separated by ~ on its own line.
// Each block has 5 lines:
//
//	Site~
//	BanBy~
//	BanTime~
//	Prefix Suffix [Type] [Level]
//	~
//
// Legacy files (no Type) are read as Type=BAN_SITE. If the file does not
// exist, returns nil with no error (no bans).
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

		ban := &types.BanData{Type: types.BAN_SITE}

		// Line 1: Site~
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

		// Line 4: Prefix Suffix [Type] [Level]
		if !scanner.Scan() {
			break
		}
		parts := strings.Fields(scanner.Text())
		if len(parts) >= 2 {
			ban.Prefix = parts[0] == "1"
			ban.Suffix = parts[1] == "1"
		}
		if len(parts) >= 3 {
			fmt.Sscanf(parts[2], "%d", &ban.Type)
			if ban.Type == 0 {
				ban.Type = types.BAN_SITE
			}
		}
		if len(parts) >= 4 {
			fmt.Sscanf(parts[3], "%d", &ban.Level)
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
		banType := ban.Type
		if banType == 0 {
			banType = types.BAN_SITE
		}
		fmt.Fprintf(writer, "%s~\n", ban.Name)
		fmt.Fprintf(writer, "%s~\n", ban.BanBy)
		fmt.Fprintf(writer, "%s~\n", ban.BanTime)
		fmt.Fprintf(writer, "%d %d %d %d\n", prefixFlag, suffixFlag, banType, ban.Level)
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
