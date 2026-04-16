package testclient

import (
	"testing"
	"time"

	"github.com/eilidhmae/smaug/internal/act"
	"github.com/eilidhmae/smaug/internal/world"
)

// testDataDir is the shared fixture location used by G2 harness tests.
// G4 will replace this with a dedicated testclient fixture.
const testDataDir = "../../cmd/smaug/testdata"

func TestStart_BootsLoop(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))
	if h.Port <= 0 {
		t.Fatalf("Port = %d, want > 0", h.Port)
	}

	c := h.Dial(t)
	defer c.Close()

	got := c.ReadUntil("what name", 3*time.Second)
	if got == "" {
		t.Fatal("expected greeting, got empty")
	}
}

func TestStart_Cleanup(t *testing.T) {
	// Run Start inside a subtest so its t.Cleanup fires when the
	// subtest ends, releasing harnessMu.
	t.Run("first", func(t *testing.T) {
		h := Start(t, WithDataDir(testDataDir))
		if h.Port <= 0 {
			t.Fatalf("first Port = %d", h.Port)
		}
	})
	// Second Start proves the package-level mutex was released.
	h2 := Start(t, WithDataDir(testDataDir))
	if h2.Port <= 0 {
		t.Fatalf("second Port = %d, want > 0", h2.Port)
	}
}

func TestQuery_Serialized(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))

	c := h.Dial(t)
	defer c.Close()

	// Wait for the greeting so the descriptor is fully registered with
	// the world before we inspect it.
	c.ReadUntil("what name", 3*time.Second)

	var count int
	h.Query(func(w *world.World) {
		count = len(w.Descriptors)
	})
	if count < 1 {
		t.Errorf("descriptor count = %d, want >= 1", count)
	}
}

func TestShutdowns_Signal(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))

	// Fire the ShutdownFunc hook from inside the loop goroutine.
	h.Query(func(*world.World) {
		act.ShutdownFunc(false)
	})

	select {
	case req := <-h.Shutdowns():
		if req.Reboot {
			t.Errorf("Reboot = true, want false")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no shutdown request arrived")
	}

	// Loop must exit within a second.
	select {
	case <-h.loopDone:
		// ok
	case <-time.After(2 * time.Second):
		t.Fatal("loop did not exit after shutdown")
	}
}

func TestQuery_PanicsAfterLoopExit(t *testing.T) {
	h := Start(t, WithDataDir(testDataDir))

	// Stop the loop and wait for it to finish.
	h.loop.Cancel()
	select {
	case <-h.loopDone:
		// ok
	case <-time.After(2 * time.Second):
		t.Fatal("loop did not exit after Cancel")
	}

	// Now Query should panic — not block forever.
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Query should have panicked after loop exited")
		}
	}()
	h.Query(func(*world.World) {
		t.Error("Query fn should not run after loop exit")
	})
}
