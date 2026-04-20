// Package game — interactive mob/character editor (CON_MEDIT substate) menu
// renderers.
//
// This file ports src/omedit.c medit_disp_menu / medit_disp_npc_menu /
// medit_disp_pc_menu into Go. The C dispatcher at :814-825 branches on
// IS_NPC(victim); the Go port mirrors that shape exactly via the method
// victim.IsNPC() on *types.CharData (confirmed at internal/types/character.go:374-375).
//
// Wave 1 scope: entry-point skeletons only. Both branches emit a placeholder
// message so the CON_MEDIT loop arm can be exercised end-to-end without
// Wave-2 menu content blocking integration testing. Wave 2 fills the real
// NPC / PC renderers from src/omedit.c:827-945.
//
// Plan: plan-phase6-olc-medit.md §G2 / §G3.
package game

import (
	"github.com/eilidhmae/smaug/internal/types"
)

// MeditDispMenu is the single entry point the rest of the codebase calls
// to redisplay the medit main menu for a descriptor. It dispatches to the
// NPC or PC sub-renderer based on the current victim pointer stashed in
// d.Olc.Target. Mirrors C medit_disp_menu at src/omedit.c:814-825.
//
// Defensive: if d, d.Olc, or the target is absent / the wrong type, the
// call is a no-op. In valid flow this cannot happen (meditParse's
// type-assertion guard fires first), but the menu entry point can be
// reached from elsewhere (e.g. the eventual DoMedit no-arg path) and so
// double-checks.
func MeditDispMenu(d *types.DescriptorData) {
	if d == nil || d.Olc == nil {
		return
	}
	victim, ok := d.Olc.Target.(*types.CharData)
	if !ok || victim == nil {
		return
	}
	if victim.IsNPC() {
		meditDispNpcMenu(d)
		return
	}
	meditDispPcMenu(d)
}

// meditDispNpcMenu renders the NPC main menu. Wave 1 is a placeholder —
// Wave 2 ports the 59-line renderer from src/omedit.c:827-885 with SPEC,
// DEFAULT_POS, ATTACK, DEFENSE, STATS, flags, RIS, parts fields. The
// placeholder line is load-bearing only insofar as it lets Wave-1
// integration tests distinguish "MeditDispMenu reached NPC branch" from
// "MeditDispMenu reached PC branch" — it is not user-facing in a landed
// build because Wave 2 overwrites the body before this ships to players.
func meditDispNpcMenu(d *types.DescriptorData) {
	d.WriteToBuffer("Medit menu: npc — Wave 2 fills the renderer.\n\r")
}

// meditDispPcMenu renders the PC main menu. Wave 1 placeholder — Wave 2
// ports the 59-line renderer from src/omedit.c:887-945 with CLASS, RACE,
// PRACTICE, PASSWORD, SAVE_MENU, PC_FLAGS, PCDATA_FLAGS, COPPER/SILVER,
// FAVOR fields.
func meditDispPcMenu(d *types.DescriptorData) {
	d.WriteToBuffer("Medit menu: pc — Wave 2 fills the renderer.\n\r")
}
