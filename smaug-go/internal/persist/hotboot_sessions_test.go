//go:build !windows

package persist

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestSaveHotbootSessions_FileMode pins the 0600 permission on hotboot.dat.
// Post-adversary fix (2026-04-19): prior `os.Create` landed files at 0644
// (umask-masked default), making player-name + host data readable by any
// local user. Security adversary flagged as MEDIUM.
func TestSaveHotbootSessions_FileMode(t *testing.T) {
	dir := t.TempDir()
	if err := SaveHotbootSessions(dir, nil); err != nil {
		t.Fatalf("SaveHotbootSessions: %v", err)
	}
	info, err := os.Stat(filepath.Join(dir, "hotboot", "hotboot.dat"))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("hotboot.dat mode = %#o, want 0o600", mode)
	}
}

func TestSaveHotbootSessions_Empty(t *testing.T) {
	dir := t.TempDir()
	if err := SaveHotbootSessions(dir, nil); err != nil {
		t.Fatalf("SaveHotbootSessions: %v", err)
	}
	path := filepath.Join(dir, "hotboot", "hotboot.dat")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "$\n" {
		t.Errorf("empty file = %q, want %q", string(data), "$\n")
	}
}

func TestSaveHotbootSessions_OneEntry(t *testing.T) {
	dir := t.TempDir()
	sess := []HotbootSession{{FDIndex: 4, RoomVnum: 3001, Port: 4000, IdleTicks: 7, Ansi: true, Name: "Alice", Host: "10.0.0.1"}}
	if err := SaveHotbootSessions(dir, sess); err != nil {
		t.Fatalf("SaveHotbootSessions: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "hotboot", "hotboot.dat"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(data), "\n")
	// Expect: 1 session line, 1 sentinel, 1 trailing empty due to terminal newline.
	if len(lines) != 3 {
		t.Fatalf("line count = %d, want 3; data=%q", len(lines), string(data))
	}
	if lines[1] != "$" {
		t.Errorf("sentinel = %q, want %q", lines[1], "$")
	}
}

func TestSaveHotbootSessions_MultipleEntries(t *testing.T) {
	dir := t.TempDir()
	sess := []HotbootSession{
		{FDIndex: 4, RoomVnum: 3001, Port: 4000, IdleTicks: 0, Ansi: true, Name: "Alice", Host: "h1"},
		{FDIndex: 5, RoomVnum: 3002, Port: 4000, IdleTicks: 3, Ansi: false, Name: "Bob", Host: "h2"},
		{FDIndex: 6, RoomVnum: 3003, Port: 4000, IdleTicks: 12, Ansi: true, Name: "Carol", Host: "h3"},
	}
	if err := SaveHotbootSessions(dir, sess); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "hotboot", "hotboot.dat"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	// 3 sessions + 1 sentinel = 4
	if len(lines) != 4 {
		t.Fatalf("line count = %d, want 4; data=%q", len(lines), string(data))
	}
	if lines[3] != "$" {
		t.Errorf("sentinel = %q, want %q", lines[3], "$")
	}
}

func TestLoadHotbootSessions_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := []HotbootSession{
		{FDIndex: 4, RoomVnum: 3001, Port: 4000, IdleTicks: 0, Ansi: true, Name: "Alice", Host: "host-a"},
		{FDIndex: 5, RoomVnum: 3002, Port: 4000, IdleTicks: 3, Ansi: false, Name: "Bob", Host: "host-b"},
	}
	if err := SaveHotbootSessions(dir, want); err != nil {
		t.Fatal(err)
	}
	got, err := LoadHotbootSessions(dir)
	if err != nil {
		t.Fatalf("LoadHotbootSessions: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("round-trip mismatch:\n got=%#v\nwant=%#v", got, want)
	}
}

func TestLoadHotbootSessions_MissingFile(t *testing.T) {
	dir := t.TempDir()
	got, err := LoadHotbootSessions(dir)
	if err == nil {
		t.Fatal("LoadHotbootSessions: want error, got nil")
	}
	if !os.IsNotExist(err) {
		t.Errorf("err = %v, want os.IsNotExist", err)
	}
	if got != nil {
		t.Errorf("got = %#v, want nil", got)
	}
}

func TestLoadHotbootSessions_MalformedLine_Bugs(t *testing.T) {
	dir := t.TempDir()
	hotbootDir := filepath.Join(dir, "hotboot")
	if err := os.MkdirAll(hotbootDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// First line good, second line has wrong field count.
	raw := "4\t3001\t4000\t0\t1\tAlice\thost-a\nNOTENOUGHFIELDS\n$\n"
	if err := os.WriteFile(filepath.Join(hotbootDir, "hotboot.dat"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadHotbootSessions(dir)
	if err != nil {
		t.Fatalf("LoadHotbootSessions: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("got %d sessions, want 1 (malformed line should be skipped)", len(got))
	}
}

func TestLoadHotbootSessions_NoSentinel_Bugs(t *testing.T) {
	dir := t.TempDir()
	hotbootDir := filepath.Join(dir, "hotboot")
	if err := os.MkdirAll(hotbootDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// No $ sentinel.
	raw := "4\t3001\t4000\t0\t1\tAlice\thost-a\n"
	if err := os.WriteFile(filepath.Join(hotbootDir, "hotboot.dat"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadHotbootSessions(dir)
	if err != nil {
		t.Fatalf("LoadHotbootSessions: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("got %d sessions, want 1 (partial parse)", len(got))
	}
}

func TestSaveHotbootSessions_TildeInHost(t *testing.T) {
	dir := t.TempDir()
	sess := []HotbootSession{{FDIndex: 4, RoomVnum: 3001, Port: 4000, Name: "Alice", Host: "evil~host"}}
	if err := SaveHotbootSessions(dir, sess); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "hotboot", "hotboot.dat"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "~") {
		t.Errorf("tilde not smashed: %q", string(data))
	}
}
