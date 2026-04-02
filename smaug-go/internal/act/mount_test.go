package act

import (
	"strings"
	"testing"

	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
)

// --- DoMount ---

func TestDoMount_NoArg(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	DoMount(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "Mount what?") {
		t.Errorf("expected 'Mount what?', got: %q", out)
	}
}

func TestDoMount_AlreadyMounted(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	ch.Mount = &types.CharData{Name: "horse"}

	DoMount(ch, "horse")
	out := readOutput(ch, client)
	if !strings.Contains(out, "already mounted") {
		t.Errorf("expected 'already mounted', got: %q", out)
	}
}

func TestDoMount_TargetNotFound(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	room := &types.RoomIndexData{Vnum: 1000, Name: "Test Room"}
	handler.CharToRoom(ch, room)

	DoMount(ch, "horse")
	out := readOutput(ch, client)
	if !strings.Contains(out, "aren't here") {
		t.Errorf("expected 'aren't here', got: %q", out)
	}
}

func TestDoMount_CantMountPlayers(t *testing.T) {
	_ = setupGroupWorld()
	ch, chClient := makeTestChar("Alice")
	defer chClient.Close()
	target, targetClient := makeTestChar("Bob")
	defer targetClient.Close()

	room := &types.RoomIndexData{Vnum: 1000, Name: "Test Room"}
	handler.CharToRoom(ch, room)
	handler.CharToRoom(target, room)

	DoMount(ch, "Bob")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "can't mount players") {
		t.Errorf("expected 'can't mount players', got: %q", out)
	}
	_ = targetClient
}

func TestDoMount_NotMountable(t *testing.T) {
	_ = setupGroupWorld()
	ch, chClient := makeTestChar("Alice")
	defer chClient.Close()

	mob := &types.CharData{Name: "goblin", ShortDescr: "a goblin"}
	mob.Act.Set(types.ACT_IS_NPC)

	room := &types.RoomIndexData{Vnum: 1000, Name: "Test Room"}
	handler.CharToRoom(ch, room)
	handler.CharToRoom(mob, room)

	DoMount(ch, "goblin")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "cannot be mounted") {
		t.Errorf("expected 'cannot be mounted', got: %q", out)
	}
}

func TestDoMount_Success(t *testing.T) {
	_ = setupGroupWorld()
	ch, chClient := makeTestChar("Alice")
	defer chClient.Close()

	mount := &types.CharData{Name: "horse", ShortDescr: "a horse"}
	mount.Act.Set(types.ACT_IS_NPC)
	mount.Act.Set(types.ACT_MOUNTABLE)

	room := &types.RoomIndexData{Vnum: 1000, Name: "Test Room"}
	handler.CharToRoom(ch, room)
	handler.CharToRoom(mount, room)

	DoMount(ch, "horse")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "You mount a horse") {
		t.Errorf("expected 'You mount a horse', got: %q", out)
	}
	if ch.Mount != mount {
		t.Error("ch.Mount should be mount")
	}
	if mount.Mount != ch {
		t.Error("mount.Mount should be ch (rider link)")
	}
	if ch.Position != types.POS_MOUNTED {
		t.Errorf("expected POS_MOUNTED (%d), got: %d", types.POS_MOUNTED, ch.Position)
	}
}

// --- DoDismount ---

func TestDoDismount_NotMounted(t *testing.T) {
	_ = setupGroupWorld()
	ch, client := makeTestChar("Alice")
	defer client.Close()

	DoDismount(ch, "")
	out := readOutput(ch, client)
	if !strings.Contains(out, "aren't mounted") {
		t.Errorf("expected 'aren't mounted', got: %q", out)
	}
}

func TestDoDismount_Success(t *testing.T) {
	_ = setupGroupWorld()
	ch, chClient := makeTestChar("Alice")
	defer chClient.Close()

	mount := &types.CharData{Name: "horse", ShortDescr: "a horse"}
	ch.Mount = mount
	mount.Mount = ch
	ch.Position = types.POS_MOUNTED

	DoDismount(ch, "")
	out := readOutput(ch, chClient)
	if !strings.Contains(out, "You dismount a horse") {
		t.Errorf("expected 'You dismount a horse', got: %q", out)
	}
	if ch.Mount != nil {
		t.Error("ch.Mount should be nil")
	}
	if mount.Mount != nil {
		t.Error("mount.Mount should be nil")
	}
	if ch.Position != types.POS_STANDING {
		t.Errorf("expected POS_STANDING, got: %d", ch.Position)
	}
}
