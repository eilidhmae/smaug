// Package magic — Phase 5 Tier 4 G3 Group C: movement / teleportation.
//
// Ports six teleport-style spells from src/magic.c. All honor the Go port's
// established destination-blocking flags: ROOM_PRIVATE, ROOM_SOLITARY,
// ROOM_NO_ASTRAL (C's "no teleport" flag), ROOM_DEATH, ROOM_NO_RECALL, plus
// ROOM_NO_RECALL on the caster's room for spells that displace the caster.
//
//   - spell_astral_walk    (magic.c:4911) teleport caster to named char
//   - spell_gate           (magic.c:3616) open a gate to target's room
//   - spell_mist_walk      (magic.c:6237) vampire version of astral_walk
//   - spell_transport      (magic.c:5574) targeted teleport
//   - spell_word_of_recall (magic.c:5279) teleport caster to ROOM_VNUM_TEMPLE
//   - spell_group_teleport (magic.c:5066) teleport followers to caster
//
// MVP simplifications per spell are documented inline.
package magic

import (
	"github.com/eilidhmae/smaug/internal/handler"
	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/world"
)

// canTeleportTo returns true if the destination room is eligible for
// teleport-style spells. Checks PRIVATE, SOLITARY, NO_ASTRAL, DEATH, NO_RECALL.
func canTeleportTo(room *types.RoomIndexData) bool {
	if room == nil {
		return false
	}
	if room.RoomFlags.IsSet(types.ROOM_PRIVATE) ||
		room.RoomFlags.IsSet(types.ROOM_SOLITARY) ||
		room.RoomFlags.IsSet(types.ROOM_NO_ASTRAL) ||
		room.RoomFlags.IsSet(types.ROOM_DEATH) ||
		room.RoomFlags.IsSet(types.ROOM_NO_RECALL) {
		return false
	}
	return true
}

// relocate is the common "fade out / fade in" pattern for teleports. Removes
// ch from its current room, places in dest, announces to both rooms.
func relocate(ch *types.CharData, dest *types.RoomIndexData, fadeOut, fadeIn string) {
	if ch.InRoom != nil {
		for _, p := range ch.InRoom.People {
			if p != ch && p.Desc != nil {
				p.Sendf(fadeOut, ch.Name)
			}
		}
		handler.CharFromRoom(ch)
	}
	handler.CharToRoom(ch, dest)
	for _, p := range dest.People {
		if p != ch && p.Desc != nil {
			p.Sendf(fadeIn, ch.Name)
		}
	}
}

// SpellAstralWalk teleports the caster to the victim's room.
// Port of spell_astral_walk (C src/magic.c:4911).
//
// MVP simplification: no overland-map support (no PLR_ONMAP hand-off); no
// magichell random reroute. Victim must be in a room that passes canTeleportTo,
// and the caster's room must not be NO_RECALL.
func SpellAstralWalk(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim == nil || victim.InRoom == nil {
		ch.Send("You fail to locate them.\n\r")
		return
	}
	if ch.InRoom != nil && ch.InRoom.RoomFlags.IsSet(types.ROOM_NO_RECALL) {
		ch.Send("You failed.\n\r")
		return
	}
	if !canTeleportTo(victim.InRoom) {
		ch.Send("You failed.\n\r")
		return
	}
	if victim.Level >= level+15 {
		ch.Send("You failed.\n\r")
		return
	}
	if victim.IsNPC() && SavesSpellStaff(level, victim) {
		ch.Send("You failed.\n\r")
		return
	}
	relocate(ch, victim.InRoom,
		"%s disappears in a flash of light!\n\r",
		"%s appears in a flash of light!\n\r")
}

// SpellGate opens a portal (short-range teleport) to the victim's location.
// Port of spell_gate (C src/magic.c:3616).
//
// Divergence from C: the original spell_gate in C is a one-liner that
// creates a MOB_VNUM_VAMPIRE in the caster's room (a legacy "summon vampire"
// spell — see the Diku lineage). That bears no relation to the task's
// description ("portal to target's room"). We implement the *documented*
// semantics: teleport caster to victim's room, matching the gate metaphor.
// This matches skills.dat's expected behavior for spells named "gate".
func SpellGate(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim == nil || victim.InRoom == nil || victim == ch {
		ch.Send("You cannot gate there.\n\r")
		return
	}
	if ch.InRoom != nil && ch.InRoom.RoomFlags.IsSet(types.ROOM_NO_RECALL) {
		ch.Send("You failed.\n\r")
		return
	}
	if !canTeleportTo(victim.InRoom) {
		ch.Send("You failed.\n\r")
		return
	}
	if victim.Level >= level+15 {
		ch.Send("You failed.\n\r")
		return
	}
	if victim.IsNPC() && SavesSpellStaff(level, victim) {
		ch.Send("You failed.\n\r")
		return
	}
	relocate(ch, victim.InRoom,
		"%s steps through a shimmering gate!\n\r",
		"%s arrives through a shimmering gate!\n\r")
}

// SpellMistWalk is the vampire-flavored astral walk. Port of spell_mist_walk
// (C src/magic.c:6237).
//
// MVP simplification: we skip the C vampire-class / bloodthirst / day-hour
// gating (needs vampire class support). Otherwise mechanics match astral_walk.
func SpellMistWalk(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim == nil || victim.InRoom == nil || victim == ch {
		ch.Send("You cannot sense your victim...\n\r")
		return
	}
	if ch.InRoom != nil && ch.InRoom.RoomFlags.IsSet(types.ROOM_NO_RECALL) {
		ch.Send("You failed.\n\r")
		return
	}
	if !canTeleportTo(victim.InRoom) {
		ch.Send("You failed.\n\r")
		return
	}
	if victim.Level >= level+15 {
		ch.Send("You failed.\n\r")
		return
	}
	if victim.IsNPC() && SavesSpellStaff(level, victim) {
		ch.Send("You failed.\n\r")
		return
	}
	relocate(ch, victim.InRoom,
		"%s dissolves into a cloud of glowing mist, then vanishes!\n\r",
		"A cloud of glowing mist engulfs you, then withdraws to unveil %s!\n\r")
}

// SpellTransport teleports the caster to victim's room.
// Port of spell_transport (C src/magic.c:5574).
//
// MVP simplification: C's spell_transport *also* moves a named object from
// caster's inventory into the victim's inventory (second target_name arg).
// The Go cast pipeline doesn't yet support two-arg spell targets, so we
// implement the movement-only behavior (caster→victim room) and skip the
// object relocation. Documented here rather than silently omitted.
func SpellTransport(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if victim == nil || victim.InRoom == nil || victim == ch {
		ch.Send("You cannot transport there.\n\r")
		return
	}
	if victim.InRoom == ch.InRoom {
		ch.Send("They are right beside you!\n\r")
		return
	}
	if ch.InRoom != nil && ch.InRoom.RoomFlags.IsSet(types.ROOM_NO_RECALL) {
		ch.Send("You failed.\n\r")
		return
	}
	if !canTeleportTo(victim.InRoom) {
		ch.Send("You failed.\n\r")
		return
	}
	if victim.Level >= level+15 {
		ch.Send("You failed.\n\r")
		return
	}
	if victim.IsNPC() && SavesSpellStaff(level, victim) {
		ch.Send("You failed.\n\r")
		return
	}
	relocate(ch, victim.InRoom,
		"%s dematerializes in a blaze of light!\n\r",
		"%s materializes out of nowhere!\n\r")
}

// SpellWordOfRecall teleports the caster (or victim, in C) to the temple.
// Port of spell_word_of_recall (C src/magic.c:5279). C simply delegates to
// do_recall; we inline the relevant logic here to avoid a package-level
// dependency on internal/act.
func SpellWordOfRecall(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	target := victim
	if target == nil {
		target = ch
	}
	if target.Fighting != nil {
		target.Send("You can't recall while fighting!\n\r")
		return
	}
	if target.InRoom != nil && target.InRoom.RoomFlags.IsSet(types.ROOM_NO_RECALL) {
		target.Send("You failed.\n\r")
		return
	}
	temple := w.GetRoom(types.ROOM_VNUM_TEMPLE)
	if temple == nil {
		target.Send("You are completely lost.\n\r")
		return
	}
	if target.InRoom == temple {
		target.Send("You are already there.\n\r")
		return
	}
	target.Move /= 2
	relocate(target, temple,
		"%s disappears in a flash of light.\n\r",
		"%s arrives in a flash of light.\n\r")
	target.Send("You recall to the temple!\n\r")
}

// SpellGroupTeleport teleports all followers of the caster to the caster's
// current room. Port of spell_group_teleport (C src/magic.c:5066).
//
// Divergence from C: the original spell teleports the caster's *group* to
// a random room and charges 100 mana per group member. Our interpretation
// from the task description ("teleport all followers to caster") is the
// simpler, observable behavior: collect all followers across the world and
// pull them to the caster's room. This matches the spell's name more
// intuitively and avoids needing a full group-accounting subsystem. Still
// honors ROOM_NO_RECALL on the caster's room.
func SpellGroupTeleport(w *world.World, sn int, level int, ch *types.CharData, victim *types.CharData) {
	if ch.InRoom == nil {
		return
	}
	if ch.InRoom.RoomFlags.IsSet(types.ROOM_NO_RECALL) {
		ch.Send("You failed.\n\r")
		return
	}
	dest := ch.InRoom
	moved := 0
	// Snapshot world characters so CharFromRoom/CharToRoom mutations are safe.
	snapshot := make([]*types.CharData, len(w.Characters))
	copy(snapshot, w.Characters)
	for _, f := range snapshot {
		if f == nil || f == ch {
			continue
		}
		// Is f following ch (Master or Leader)?
		if f.Master != ch && f.Leader != ch {
			continue
		}
		if f.InRoom == dest {
			continue
		}
		if f.InRoom != nil && f.InRoom.RoomFlags.IsSet(types.ROOM_NO_RECALL) {
			continue
		}
		relocate(f, dest,
			"%s is whisked away in a tornado!\n\r",
			"%s drops out of a tornado!\n\r")
		moved++
	}
	if moved == 0 {
		ch.Send("You have no followers to teleport.\n\r")
		return
	}
	ch.Sendf("You teleport %d follower(s) to your location.\n\r", moved)
}
