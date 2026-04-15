package mudprog

import (
	"github.com/eilidhmae/smaug/internal/types"
)

// buildSupermob returns a transient "supermob" CharData used as the driver's
// self-reference when running object progs. Mirrors C's `set_supermob` in
// `src/mud_prog.c:3424`. The supermob inherits the obj's short description
// (used by $i/$I Translate tokens and mpecho-style output) and is placed in
// the obj's effective room: walk obj.InObj to the outermost container, then
// pick the carrier's room if held, else the obj's own InRoom.
//
// The supermob is NOT appended to room.People — it is only a self-ref for
// Driver. That matches the C model, where supermob is a singleton shuttled
// between rooms rather than a resident mob.
func buildSupermob(obj *types.ObjData) *types.CharData {
	if obj == nil {
		return nil
	}
	room := resolveObjRoom(obj)
	sm := &types.CharData{
		Name:       "supermob",
		ShortDescr: obj.ShortDescr,
		Level:      60,
		Position:   types.POS_STANDING,
		Hit:        1000,
		MaxHit:     1000,
		InRoom:     room,
	}
	sm.Act.Set(types.ACT_IS_NPC)
	return sm
}

// resolveObjRoom returns the room an obj should be considered to occupy:
// - if nested in a container, walk to the outermost obj first
// - if a carrier holds it, use the carrier's InRoom
// - else use obj.InRoom (which may still be nil for lingering references).
func resolveObjRoom(obj *types.ObjData) *types.RoomIndexData {
	if obj == nil {
		return nil
	}
	outer := obj
	for outer.InObj != nil {
		outer = outer.InObj
	}
	if outer.CarriedBy != nil {
		return outer.CarriedBy.InRoom
	}
	return outer.InRoom
}

// ObjTrigger is the object-prog equivalent of MobTrigger. It fires the first
// matching obj-prog of the given trigType on obj, using a synthesised supermob
// as the driver's self-reference. Mirrors C's `oprog_percent_check`
// (`src/mud_prog.c:3500`) + `set_supermob` / `release_supermob` flow.
func ObjTrigger(trigType int64, obj *types.ObjData, actor *types.CharData,
	victim *types.CharData, target *types.ObjData, arg string) {

	if obj == nil || obj.IndexData == nil {
		return
	}
	if len(obj.IndexData.MudProgs) == 0 {
		return
	}

	sm := buildSupermob(obj)
	if sm == nil {
		return
	}

	for _, prog := range obj.IndexData.MudProgs {
		if prog.Type&trigType == 0 {
			continue
		}
		if !triggerMatches(trigType, prog.ArgList, arg) {
			continue
		}
		Driver(prog.ComList, sm, actor, obj, victim, target, false)
		// C semantics: stop after the first match for non-GREET triggers;
		// GREET continues iterating so multiple progs on the same obj may
		// fire. See oprog_percent_check.
		if trigType != types.MPROG_GREET {
			return
		}
	}
}

// OprogWearTrigger fires WEAR progs on obj when worn. C: oprog_wear_trigger.
func OprogWearTrigger(ch *types.CharData, obj *types.ObjData) {
	ObjTrigger(types.MPROG_WEAR, obj, ch, nil, nil, "")
}

// OprogRemoveTrigger fires REMOVE progs on obj when unequipped. C: oprog_remove_trigger.
func OprogRemoveTrigger(ch *types.CharData, obj *types.ObjData) {
	ObjTrigger(types.MPROG_REMOVE, obj, ch, nil, nil, "")
}

// OprogSacTrigger fires SAC progs on obj before it is sacrificed/destroyed.
// C: oprog_sac_trigger.
func OprogSacTrigger(ch *types.CharData, obj *types.ObjData) {
	ObjTrigger(types.MPROG_SAC, obj, ch, nil, nil, "")
}

// OprogLookTrigger fires LOOK progs when ch looks at obj. C: oprog_look_trigger
// (named in mud.h, called from do_look).
func OprogLookTrigger(ch *types.CharData, obj *types.ObjData) {
	ObjTrigger(types.MPROG_LOOK, obj, ch, nil, nil, "")
}

// OprogExamineTrigger fires EXA progs when ch examines obj. C: oprog_examine_trigger.
func OprogExamineTrigger(ch *types.CharData, obj *types.ObjData) {
	ObjTrigger(types.MPROG_EXA, obj, ch, nil, nil, "")
}

// OprogZapTrigger fires ZAP progs when a zap-capable item discharges.
// C: oprog_zap_trigger.
func OprogZapTrigger(ch *types.CharData, obj *types.ObjData) {
	ObjTrigger(types.MPROG_ZAP, obj, ch, nil, nil, "")
}

// OprogGetTrigger fires GET progs when ch picks up obj. C: oprog_get_trigger.
func OprogGetTrigger(ch *types.CharData, obj *types.ObjData) {
	ObjTrigger(types.MPROG_GET, obj, ch, nil, nil, "")
}

// OprogDropTrigger fires DROP progs when ch drops obj. C: oprog_drop_trigger.
func OprogDropTrigger(ch *types.CharData, obj *types.ObjData) {
	ObjTrigger(types.MPROG_DROP, obj, ch, nil, nil, "")
}

// OprogDamageTrigger fires DAMAGE progs when obj is damaged in combat or
// otherwise. C: oprog_damage_trigger.
func OprogDamageTrigger(ch *types.CharData, obj *types.ObjData) {
	ObjTrigger(types.MPROG_DAMAGE, obj, ch, nil, nil, "")
}

// OprogRepairTrigger fires REPAIR progs when a shopkeeper restores obj's
// condition. C: oprog_repair_trigger.
func OprogRepairTrigger(ch *types.CharData, obj *types.ObjData) {
	ObjTrigger(types.MPROG_REPAIR, obj, ch, nil, nil, "")
}

// OprogPullTrigger fires PULL progs on lever-type objects. C: oprog_pull_trigger.
func OprogPullTrigger(ch *types.CharData, obj *types.ObjData) {
	ObjTrigger(types.MPROG_PULL, obj, ch, nil, nil, "")
}

// OprogPushTrigger fires PUSH progs on button-type objects. C: oprog_push_trigger.
func OprogPushTrigger(ch *types.CharData, obj *types.ObjData) {
	ObjTrigger(types.MPROG_PUSH, obj, ch, nil, nil, "")
}

// OprogUseTrigger fires USE progs when a quaff/recite/brandish/zap consumes
// or invokes obj. C: oprog_use_trigger.
func OprogUseTrigger(ch *types.CharData, obj *types.ObjData) {
	ObjTrigger(types.MPROG_USE, obj, ch, nil, nil, "")
}

// OprogGreetTrigger scans room.Contents and fires GREET progs on any obj
// present that has a GREET prog. C: oprog_greet_trigger
// (`src/mud_prog.c:3541`). Only objects on the floor greet — carried/worn
// objects are skipped, matching C which iterates only `first_content`.
func OprogGreetTrigger(ch *types.CharData) {
	if ch == nil || ch.InRoom == nil {
		return
	}
	// Snapshot so a prog that extracts obj doesn't corrupt iteration.
	contents := make([]*types.ObjData, len(ch.InRoom.Contents))
	copy(contents, ch.InRoom.Contents)
	for _, obj := range contents {
		if obj == nil {
			continue
		}
		ObjTrigger(types.MPROG_GREET, obj, ch, nil, nil, "")
	}
}

// OprogSpeechTrigger fires SPEECH progs on every obj in ch's room and every
// obj ch is carrying. C: oprog_speech_trigger (`src/mud_prog.c:3555`) only
// scans room contents; the Go port additionally scans carrying because obj-
// prog speech on a held item is useful and cheap.
func OprogSpeechTrigger(ch *types.CharData, argument string) {
	if ch == nil {
		return
	}
	if ch.InRoom != nil {
		contents := make([]*types.ObjData, len(ch.InRoom.Contents))
		copy(contents, ch.InRoom.Contents)
		for _, obj := range contents {
			if obj == nil {
				continue
			}
			ObjTrigger(types.MPROG_SPEECH, obj, ch, nil, nil, argument)
		}
	}
	carrying := make([]*types.ObjData, len(ch.Carrying))
	copy(carrying, ch.Carrying)
	for _, obj := range carrying {
		if obj == nil {
			continue
		}
		ObjTrigger(types.MPROG_SPEECH, obj, ch, nil, nil, argument)
	}
}

// OprogCommandTrigger is the command-trigger hook called from the command
// interpreter. If any obj-prog on an item in ch's room or inventory matches
// the command, returns true and callers should skip normal dispatch.
//
// NOTE: C uses a dedicated CMD_PROG enum value (src/mud.h:6546). The Go port
// does not yet have a MPROG_CMD bit, so this helper is stubbed to always
// return false. TODO(tier3): add MPROG_CMD bit + loader plumbing and enable
// keyword matching here (logic would mirror OprogSpeechTrigger + return true
// on first match).
func OprogCommandTrigger(ch *types.CharData, argument string) bool {
	if ch == nil || argument == "" {
		return false
	}
	return false
}

// OprogRandomTrigger fires RAND progs on obj. Must be called from a per-pulse
// object scan (mirrors C `oprog_random_trigger` at top of obj_update).
// C: oprog_random_trigger (`src/mud_prog.c:3590`).
func OprogRandomTrigger(obj *types.ObjData) {
	if obj == nil || obj.IndexData == nil {
		return
	}
	if len(obj.IndexData.MudProgs) == 0 {
		return
	}
	// RAND fires in-room; RANDIW fires anywhere in the world. The C port
	// distinguishes the two in obj_update itself; we fire RAND whenever obj
	// has a room (anywhere a PC might witness) and RANDIW unconditionally.
	if resolveObjRoom(obj) != nil {
		ObjTrigger(types.MPROG_RAND, obj, nil, nil, nil, "")
	}
	ObjTrigger(types.MPROG_RANDIW, obj, nil, nil, nil, "")
}

// OprogActTrigger fires ACT progs on obj in response to a room-level act
// string. C: oprog_act_trigger (`src/mud_prog.c:3788`). Typically queued
// via MpAct list, but the MVP fires synchronously.
func OprogActTrigger(text string, obj *types.ObjData, ch *types.CharData) {
	if obj == nil {
		return
	}
	ObjTrigger(types.MPROG_ACT, obj, ch, nil, nil, text)
}
