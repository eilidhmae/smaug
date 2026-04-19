package testclient

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/act"
	"github.com/eilidhmae/smaug/internal/types"
)

// TestTestclient_PlistShowsPrimeMaterial — any player can `plist` and
// sees the auto-seeded "Prime Material" plane. Pins plan §A10 mortal path.
func TestTestclient_PlistShowsPrimeMaterial(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))

	c := h.QuickLogin(t, "Plisty")
	defer c.Close()

	c.Send("plist")
	got := c.ReadUntil("Prime Material", 3*time.Second)
	if !strings.Contains(got, "Prime Material") {
		t.Errorf("expected 'Prime Material' in plist output; got %q", got)
	}
}

// TestTestclient_PsetCreateAndListRoundTrip — an admin can create a
// plane, see it via plist, delete it, see it absent. End-to-end
// exercise of DoPset create / DoPlist / DoPset delete.
func TestTestclient_PsetCreateAndListRoundTrip(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))

	// Need LEVEL_GREATER (59) for `pset`.
	c := h.NewCharacter(t, CharSpec{
		Name:     "Psetadm",
		Password: defaultQuickLoginPassword,
		Trust:    types.LEVEL_GREATER,
	})
	defer c.Close()

	// Create a plane and verify the ack arrives.
	c.Send("pset astrotest create")
	ack := c.ReadUntil("Plane created", 3*time.Second)
	if !strings.Contains(ack, "Plane created") {
		t.Fatalf("expected 'Plane created.'; got %q", ack)
	}

	// plist must show it.
	c.Send("plist")
	listed := c.ReadUntil("astrotest", 3*time.Second)
	if !strings.Contains(listed, "astrotest") {
		t.Errorf("expected 'astrotest' in plist output; got %q", listed)
	}

	// Delete and verify absent.
	c.Send("pset astrotest delete")
	delAck := c.ReadUntil("Plane deleted", 3*time.Second)
	if !strings.Contains(delAck, "Plane deleted") {
		t.Fatalf("expected 'Plane deleted.'; got %q", delAck)
	}

	// plist again — astrotest should be gone. Use Prime Material as the
	// "plist completed" sentinel (CheckPlanes guarantees it's always listed).
	c.Send("plist")
	afterDel := c.ReadUntil("Prime Material", 3*time.Second)
	if strings.Contains(afterDel, "astrotest") {
		t.Errorf("astrotest still present after delete; got %q", afterDel)
	}
}

// TestTestclient_PsetSavePersistsToDisk — `pset save` writes the
// in-memory slice to act.PlanesFilePath. The file must contain the
// tilde-terminated block for our new plane.
func TestTestclient_PsetSavePersistsToDisk(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))

	// Capture existence of the planes file BEFORE boot so we know how to
	// clean up — testdata does not ship planes.dat, so the save step
	// will create one and we must remove it.
	preexisting := false
	if act.PlanesFilePath != "" {
		if _, err := os.Stat(act.PlanesFilePath); err == nil {
			preexisting = true
		}
	}

	c := h.NewCharacter(t, CharSpec{
		Name:     "Psetsav",
		Password: defaultQuickLoginPassword,
		Trust:    types.LEVEL_GREATER,
	})
	defer c.Close()

	c.Send("pset savetest create")
	_ = c.ReadUntil("Plane created", 3*time.Second)

	c.Send("pset save")
	_ = c.ReadUntil("Planes saved", 3*time.Second)

	// The planes file path should be set by boot wire.
	if act.PlanesFilePath == "" {
		t.Fatal("act.PlanesFilePath not wired")
	}
	data, err := os.ReadFile(act.PlanesFilePath)
	if err != nil {
		t.Fatalf("read planes.dat: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "savetest~") {
		t.Errorf("expected 'savetest~' block in saved file; got:\n%s", content)
	}

	// Cleanup: restore pre-existing or delete the newly-created file so
	// we don't pollute the testdata tree.
	path := act.PlanesFilePath
	t.Cleanup(func() {
		if !preexisting {
			_ = os.Remove(path)
		}
	})
}
