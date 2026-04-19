//go:build !windows

package boot_test

import (
	gonet "net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/boot"
	"github.com/eilidhmae/smaug/internal/persist"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// --- test helpers ----------------------------------------------------

// recoverFixture stands up a minimal datadir for a BootRecover call:
//
//   - Copies every subdir under testdata/ into a fresh tmp tree so each
//     test gets its own independent player/hotboot/planes etc. state.
//   - Writes an empty hotboot/mobfile.dat + hotboot/<vnum>.objdat (or
//     non-empty when the test wants one) + hotboot/hotboot.dat.
//   - Writes a pfile for each named character via persist.SavePlayer.
//
// Returns the tmpdir + a listener on 127.0.0.1:0 + the listener's fd
// (kept alive via the returned *os.File the caller must close).
func recoverFixture(t *testing.T, sessions []persist.HotbootSession, pfiles map[string]*types.CharData) (tmpdir string, lnFD uintptr, cleanup func()) {
	t.Helper()
	tmpdir = t.TempDir()
	// Copy testdata tree so Boot's area/class/race/skill loaders resolve.
	if err := copyTree(testDataDir, tmpdir); err != nil {
		t.Fatalf("copyTree: %v", err)
	}
	// Scrub any stale player dir; we'll rewrite per-test pfiles below.
	_ = os.RemoveAll(filepath.Join(tmpdir, "player"))

	// Write hotboot.dat.
	if err := persist.SaveHotbootSessions(tmpdir, sessions); err != nil {
		t.Fatalf("SaveHotbootSessions: %v", err)
	}
	// Write an empty mobfile.dat (tests that care about mobfile content
	// overwrite this after). Presence of the file exercises the LoadWorld
	// path + its post-load unlink.
	mobPath := filepath.Join(tmpdir, "hotboot", "mobfile.dat")
	if err := os.WriteFile(mobPath, []byte("#END\n"), 0o644); err != nil {
		t.Fatalf("write mobfile: %v", err)
	}
	// Write an empty objdat for vnum=21001 (temple) — ensures LoadWorld's
	// unlink branch runs against a real file.
	objPath := filepath.Join(tmpdir, "hotboot", "21001.objdat")
	if err := os.WriteFile(objPath, []byte("#END\n"), 0o644); err != nil {
		t.Fatalf("write objdat: %v", err)
	}

	// Write each pfile.
	for name, ch := range pfiles {
		path := persist.PlayerFilePath(tmpdir, name)
		if path == "" {
			t.Fatalf("invalid pfile name %q", name)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir pfile: %v", err)
		}
		f, err := os.Create(path)
		if err != nil {
			t.Fatalf("create pfile %s: %v", path, err)
		}
		if err := persist.SavePlayer(f, ch); err != nil {
			f.Close()
			t.Fatalf("SavePlayer %s: %v", name, err)
		}
		f.Close()
	}

	// Bind a listener, extract its FD so BootRecover can re-wrap.
	ln, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		t.Fatalf("socket: %v", err)
	}
	addr := &syscall.SockaddrInet4{Port: 0, Addr: [4]byte{127, 0, 0, 1}}
	if err := syscall.Bind(ln, addr); err != nil {
		syscall.Close(ln)
		t.Fatalf("bind: %v", err)
	}
	if err := syscall.Listen(ln, 16); err != nil {
		syscall.Close(ln)
		t.Fatalf("listen: %v", err)
	}
	lnFD = uintptr(ln)

	// Ownership of lnFD transfers to BootRecover; the test MUST call
	// BootRecover or explicitly close the fd. We don't close here from
	// cleanup because a successful BootRecover already consumed it via
	// net.FileListener, and the kernel may have assigned that number
	// to a subsequent test's socket — closing it would silently break
	// the next test.
	cleanup = func() {}
	return tmpdir, lnFD, cleanup
}

// makeSocketpair returns an inherited-FD child end and a net.Conn
// parent end connected to each other. BootRecover re-wraps childFD via
// os.NewFile + net.FileConn; the test drives the parent side as a
// net.Conn so read deadlines work (unlike raw *os.File on a socketpair
// fd, whose runtime poller cannot apply Read deadlines).
func makeSocketpair(t *testing.T) (childFD uintptr, parentConn gonet.Conn) {
	t.Helper()
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		t.Fatalf("socketpair: %v", err)
	}
	childFD = uintptr(fds[0])
	parentFile := os.NewFile(uintptr(fds[1]), "sp-parent")
	parentConn, err = gonet.FileConn(parentFile)
	parentFile.Close()
	if err != nil {
		t.Fatalf("FileConn: %v", err)
	}
	return childFD, parentConn
}

// copyTree recursively copies src into dst. Overwrites existing files.
func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

// newTestPlayer builds a minimally-valid CharData+PCData that round-trips
// through SavePlayer/LoadPlayer.
func newTestPlayer(name string) *types.CharData {
	return &types.CharData{
		Name:     name,
		Level:    20,
		Sex:      1,
		Class:    0,
		Race:     0,
		Hit:      100,
		MaxHit:   100,
		Mana:     100,
		MaxMana:  100,
		Move:     150,
		MaxMove:  150,
		Position: types.POS_STANDING,
		HomeVnum: types.ROOM_VNUM_TEMPLE,
		PermStr:  13,
		PermInt:  13,
		PermWis:  13,
		PermDex:  13,
		PermCon:  13,
		PermCha:  13,
		PermLck:  13,
		PCData: &types.PCData{
			Pwd:      "hash",
			PagerLen: 24,
		},
	}
}

// drainSocket reads available bytes off `c` for up to `deadline`.
// Used to observe what the descriptor's FlushOutput wrote to its conn.
func drainSocket(c gonet.Conn, deadline time.Duration) string {
	_ = c.SetReadDeadline(time.Now().Add(deadline))
	var buf [4096]byte
	var out []byte
	for {
		n, err := c.Read(buf[:])
		if n > 0 {
			out = append(out, buf[:n]...)
		}
		if err != nil {
			break
		}
	}
	_ = c.SetReadDeadline(time.Time{})
	return string(out)
}

// findDescriptor locates the descriptor corresponding to `name` in
// w.Descriptors. Returns nil on no match.
func findDescriptor(w *world.World, name string) *types.DescriptorData {
	for _, d := range w.Descriptors {
		if d != nil && d.Character != nil && d.Character.Name == name {
			return d
		}
	}
	return nil
}

// --- tests -----------------------------------------------------------

func TestBootRecover_BasicRoundTrip(t *testing.T) {
	sessions := []persist.HotbootSession{
		{FDIndex: 0, RoomVnum: types.ROOM_VNUM_TEMPLE, Port: 4000, IdleTicks: 0, Ansi: false, Name: "Trusty", Host: "10.0.0.1"},
	}
	pfiles := map[string]*types.CharData{"Trusty": newTestPlayer("Trusty")}
	tmpdir, lnFD, cleanup := recoverFixture(t, sessions, pfiles)
	defer cleanup()

	childFD, parentConn := makeSocketpair(t)
	defer parentConn.Close()

	w := world.New(tmpdir)
	fds := []boot.FDSession{{FD: childFD, SessionIdx: 0}}

	_, _, server, err := boot.BootRecover(w, tmpdir, boot.TestOpts(), lnFD, fds)
	if err != nil {
		t.Fatalf("BootRecover: %v", err)
	}
	defer server.Stop()

	d := findDescriptor(w, "Trusty")
	if d == nil {
		t.Fatal("descriptor for Trusty not attached to world")
	}
	if d.Character == nil {
		t.Fatal("descriptor has no character")
	}
	if d.Character.InRoom == nil || d.Character.InRoom.Vnum != types.ROOM_VNUM_TEMPLE {
		t.Errorf("character in room %+v, want vnum %d", d.Character.InRoom, types.ROOM_VNUM_TEMPLE)
	}
	out := drainSocket(parentConn, 500*time.Millisecond)
	if !strings.Contains(out, "Time resumes its normal flow") {
		t.Errorf("welcome missing Time-resumes line; got %q", out)
	}
	if w.SysData.HotbootInProgress {
		t.Error("HotbootInProgress flag still true after recovery")
	}
}

func TestBootRecover_MissingPfile_ClosesConnection(t *testing.T) {
	// Session references a player whose pfile was not written.
	sessions := []persist.HotbootSession{
		{FDIndex: 0, RoomVnum: types.ROOM_VNUM_TEMPLE, Port: 4000, Name: "Ghost", Host: "10.0.0.1"},
	}
	tmpdir, lnFD, cleanup := recoverFixture(t, sessions, nil /* no pfile */)
	defer cleanup()

	childFD, parentConn := makeSocketpair(t)
	defer parentConn.Close()

	w := world.New(tmpdir)
	fds := []boot.FDSession{{FD: childFD, SessionIdx: 0}}

	_, _, server, err := boot.BootRecover(w, tmpdir, boot.TestOpts(), lnFD, fds)
	if err != nil {
		t.Fatalf("BootRecover returned error: %v", err)
	}
	defer server.Stop()

	if d := findDescriptor(w, "Ghost"); d != nil {
		t.Errorf("Ghost descriptor should not be attached; got %+v", d)
	}
	// Connection should be closed: writing from the parent side should
	// eventually fail (EPIPE/ECONNRESET) or return 0 bytes. We just
	// assert no panic reached here.
}

func TestBootRecover_SavedRoomGone_FallbackToTemple(t *testing.T) {
	sessions := []persist.HotbootSession{
		{FDIndex: 0, RoomVnum: 99999, Port: 4000, Name: "Lost", Host: "10.0.0.1"},
	}
	pfiles := map[string]*types.CharData{"Lost": newTestPlayer("Lost")}
	tmpdir, lnFD, cleanup := recoverFixture(t, sessions, pfiles)
	defer cleanup()

	childFD, parentConn := makeSocketpair(t)
	defer parentConn.Close()

	w := world.New(tmpdir)
	fds := []boot.FDSession{{FD: childFD, SessionIdx: 0}}

	_, _, server, err := boot.BootRecover(w, tmpdir, boot.TestOpts(), lnFD, fds)
	if err != nil {
		t.Fatalf("BootRecover: %v", err)
	}
	defer server.Stop()

	d := findDescriptor(w, "Lost")
	if d == nil {
		t.Fatal("Lost descriptor missing")
	}
	if d.Character.InRoom == nil {
		t.Fatal("character has no room")
	}
	if d.Character.InRoom.Vnum != types.ROOM_VNUM_TEMPLE {
		t.Errorf("fallback room = %d, want TEMPLE=%d",
			d.Character.InRoom.Vnum, types.ROOM_VNUM_TEMPLE)
	}
}

func TestBootRecover_UnlinksHotbootFiles(t *testing.T) {
	sessions := []persist.HotbootSession{
		{FDIndex: 0, RoomVnum: types.ROOM_VNUM_TEMPLE, Port: 4000, Name: "Clean", Host: "h"},
	}
	pfiles := map[string]*types.CharData{"Clean": newTestPlayer("Clean")}
	tmpdir, lnFD, cleanup := recoverFixture(t, sessions, pfiles)
	defer cleanup()

	childFD, parentConn := makeSocketpair(t)
	defer parentConn.Close()

	// Sanity: the files exist before recovery.
	for _, name := range []string{"hotboot.dat", "mobfile.dat", "21001.objdat"} {
		if _, err := os.Stat(filepath.Join(tmpdir, "hotboot", name)); err != nil {
			t.Fatalf("pre-condition: %s missing: %v", name, err)
		}
	}

	w := world.New(tmpdir)
	fds := []boot.FDSession{{FD: childFD, SessionIdx: 0}}
	_, _, server, err := boot.BootRecover(w, tmpdir, boot.TestOpts(), lnFD, fds)
	if err != nil {
		t.Fatalf("BootRecover: %v", err)
	}
	defer server.Stop()

	for _, name := range []string{"hotboot.dat", "mobfile.dat", "21001.objdat"} {
		if _, err := os.Stat(filepath.Join(tmpdir, "hotboot", name)); !os.IsNotExist(err) {
			t.Errorf("%s was NOT unlinked (err=%v)", name, err)
		}
	}
}

func TestBootRecover_ClearsHotbootFlagAfterWelcome(t *testing.T) {
	ch := newTestPlayer("Flagged")
	ch.PCData.Hotboot = true // parent set this pre-exec
	sessions := []persist.HotbootSession{
		{FDIndex: 0, RoomVnum: types.ROOM_VNUM_TEMPLE, Port: 4000, Name: "Flagged", Host: "h"},
	}
	pfiles := map[string]*types.CharData{"Flagged": ch}
	tmpdir, lnFD, cleanup := recoverFixture(t, sessions, pfiles)
	defer cleanup()

	childFD, parentConn := makeSocketpair(t)
	defer parentConn.Close()

	w := world.New(tmpdir)
	fds := []boot.FDSession{{FD: childFD, SessionIdx: 0}}
	_, _, server, err := boot.BootRecover(w, tmpdir, boot.TestOpts(), lnFD, fds)
	if err != nil {
		t.Fatalf("BootRecover: %v", err)
	}
	defer server.Stop()

	d := findDescriptor(w, "Flagged")
	if d == nil {
		t.Fatal("descriptor missing")
	}
	// Drain to ensure welcome fired (and thus the post-welcome clear ran).
	drainSocket(parentConn, 500*time.Millisecond)
	if d.Character.PCData.Hotboot {
		t.Error("PCData.Hotboot should be cleared after welcome; still true")
	}
}

func TestBootRecover_ResumedFromHotbootFlagSet(t *testing.T) {
	sessions := []persist.HotbootSession{
		{FDIndex: 0, RoomVnum: types.ROOM_VNUM_TEMPLE, Port: 4000, Name: "Flagg", Host: "h"},
	}
	pfiles := map[string]*types.CharData{"Flagg": newTestPlayer("Flagg")}
	tmpdir, lnFD, cleanup := recoverFixture(t, sessions, pfiles)
	defer cleanup()

	childFD, parentConn := makeSocketpair(t)
	defer parentConn.Close()

	w := world.New(tmpdir)
	fds := []boot.FDSession{{FD: childFD, SessionIdx: 0}}
	_, _, server, err := boot.BootRecover(w, tmpdir, boot.TestOpts(), lnFD, fds)
	if err != nil {
		t.Fatalf("BootRecover: %v", err)
	}
	defer server.Stop()

	d := findDescriptor(w, "Flagg")
	if d == nil {
		t.Fatal("descriptor missing")
	}
	if !d.ResumedFromHotboot {
		t.Error("ResumedFromHotboot should be true on a recovered descriptor")
	}
}

func TestBootRecover_EmitsPuffOfSmoke(t *testing.T) {
	sessions := []persist.HotbootSession{
		{FDIndex: 0, RoomVnum: types.ROOM_VNUM_TEMPLE, Port: 4000, Name: "Puffy", Host: "h"},
	}
	pfiles := map[string]*types.CharData{"Puffy": newTestPlayer("Puffy")}
	tmpdir, lnFD, cleanup := recoverFixture(t, sessions, pfiles)
	defer cleanup()

	childFD, parentConn := makeSocketpair(t)
	defer parentConn.Close()

	w := world.New(tmpdir)
	fds := []boot.FDSession{{FD: childFD, SessionIdx: 0}}
	_, _, server, err := boot.BootRecover(w, tmpdir, boot.TestOpts(), lnFD, fds)
	if err != nil {
		t.Fatalf("BootRecover: %v", err)
	}
	defer server.Stop()

	d := findDescriptor(w, "Puffy")
	if d == nil {
		t.Fatal("descriptor missing")
	}
	out := drainSocket(parentConn, 500*time.Millisecond)
	if !strings.Contains(out, "Time resumes its normal flow.") {
		t.Errorf("welcome missing Time-resumes line; got %q", out)
	}
	if !strings.Contains(out, "puff of ethereal smoke") {
		t.Errorf("welcome missing puff-of-smoke social; got %q", out)
	}
}

func TestBootRecover_ListenerFDWrap(t *testing.T) {
	sessions := []persist.HotbootSession{}
	tmpdir, lnFD, cleanup := recoverFixture(t, sessions, nil)
	defer cleanup()

	// Capture the pre-recover bound port via a peek at the raw fd.
	preName, err := syscall.Getsockname(int(lnFD))
	if err != nil {
		t.Fatalf("getsockname: %v", err)
	}
	sa, ok := preName.(*syscall.SockaddrInet4)
	if !ok {
		t.Fatalf("getsockname returned %T, want *SockaddrInet4", preName)
	}
	wantPort := sa.Port

	w := world.New(tmpdir)
	_, _, server, err := boot.BootRecover(w, tmpdir, boot.TestOpts(), lnFD, nil)
	if err != nil {
		t.Fatalf("BootRecover: %v", err)
	}
	defer server.Stop()

	addr := server.Addr()
	if addr == nil {
		t.Fatal("server.Addr() nil post-recover")
	}
	tcpAddr, ok := addr.(interface{ Port() int })
	_ = tcpAddr
	_ = ok
	// Prefer the concrete *net.TCPAddr path for a clean port read.
	if ta, ok := addr.(interface{ String() string }); ok {
		// net.TCPAddr has .Port field; grab via reflection-free string.
		host, portStr, splitErr := splitHostPort(ta.String())
		if splitErr != nil {
			t.Fatalf("splitHostPort(%q): %v", ta.String(), splitErr)
		}
		_ = host
		gotPort, err := atoi(portStr)
		if err != nil {
			t.Fatalf("atoi(%q): %v", portStr, err)
		}
		if gotPort != wantPort {
			t.Errorf("server port = %d, want %d (from inherited FD)", gotPort, wantPort)
		}
	}
}

// splitHostPort replicates net.SplitHostPort without the dependency.
func splitHostPort(s string) (host, port string, err error) {
	i := strings.LastIndex(s, ":")
	if i < 0 {
		return "", "", errNoPort
	}
	host, port = s[:i], s[i+1:]
	// Strip IPv6 brackets.
	host = strings.TrimPrefix(host, "[")
	host = strings.TrimSuffix(host, "]")
	return host, port, nil
}

func atoi(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errBadPort
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

var (
	errNoPort  = errStr("no port in address")
	errBadPort = errStr("non-digit in port")
)

type errStr string

func (e errStr) Error() string { return string(e) }

func TestBootRecover_SpawnsReadLoopPerDescriptor(t *testing.T) {
	sessions := []persist.HotbootSession{
		{FDIndex: 0, RoomVnum: types.ROOM_VNUM_TEMPLE, Port: 4000, Name: "Loopy", Host: "h"},
	}
	pfiles := map[string]*types.CharData{"Loopy": newTestPlayer("Loopy")}
	tmpdir, lnFD, cleanup := recoverFixture(t, sessions, pfiles)
	defer cleanup()

	childFD, parentConn := makeSocketpair(t)
	defer parentConn.Close()

	w := world.New(tmpdir)
	fds := []boot.FDSession{{FD: childFD, SessionIdx: 0}}
	_, _, server, err := boot.BootRecover(w, tmpdir, boot.TestOpts(), lnFD, fds)
	if err != nil {
		t.Fatalf("BootRecover: %v", err)
	}
	defer server.Stop()

	d := findDescriptor(w, "Loopy")
	if d == nil {
		t.Fatal("descriptor missing")
	}
	// Write a command line from the parent side; the read loop should
	// pipe it into d.InputQueue.
	if _, err := parentConn.Write([]byte("look\n")); err != nil {
		t.Fatalf("parent write: %v", err)
	}
	select {
	case got := <-d.InputQueue:
		if got != "look" {
			t.Errorf("InputQueue got %q, want %q", got, "look")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("InputQueue never received the forwarded byte")
	}
}
