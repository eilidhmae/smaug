package persist

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadNoAuction_MissingFile(t *testing.T) {
	got, err := LoadNoAuction("/does/not/exist/noauction.dat")
	if err != nil {
		t.Fatalf("expected nil err on missing file; got %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice; got %v", got)
	}
}

func TestLoadNoAuction_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "noauction.dat")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadNoAuction(path)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice; got %v", got)
	}
}

func TestLoadNoAuction_Populated(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "noauction.dat")
	body := "100\n200\n3001\n0\n"
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadNoAuction(path)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	want := []int{100, 200, 3001}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d (got %v)", len(got), len(want), got)
	}
	for i, v := range want {
		if got[i] != v {
			t.Errorf("got[%d] = %d, want %d", i, got[i], v)
		}
	}
}

// A `0` is the C-format sentinel; anything after should be ignored.
func TestLoadNoAuction_StopsAtZeroSentinel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "noauction.dat")
	body := "42\n0\n99\n"
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadNoAuction(path)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got) != 1 || got[0] != 42 {
		t.Errorf("expected [42]; got %v", got)
	}
}

// Malformed lines logged via util.Bug but parse continues.
func TestLoadNoAuction_MalformedLineSkipped(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "noauction.dat")
	body := "100\nxyz\n200\n0\n"
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadNoAuction(path)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	want := []int{100, 200}
	if len(got) != 2 || got[0] != 100 || got[1] != 200 {
		t.Errorf("expected %v; got %v", want, got)
	}
}
