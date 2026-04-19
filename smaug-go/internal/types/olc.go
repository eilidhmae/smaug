package types

// OlcData tracks interactive-OLC session state for a descriptor.
// Moral equivalent of C's OLC_DATA struct (src/olc.h); populated when
// d.Connected transitions to CON_REDIT / CON_OEDIT / CON_MEDIT, and nil
// otherwise. One allocation per descriptor — cleared via CleanupOlc on
// session exit.
//
// Per plan-phase6-olc-redit.md §Go Design (chosen approach A): the struct
// lives on DescriptorData (not CharData) because the mode-state union is a
// descriptor-scoped concern; the per-character editor callback continues to
// live on CharData.EditorSave (Tier 12).
//
// Target / Spare fields are `any` (not pointer types) so one struct serves
// every OLC editor — a redit session stashes *RoomIndexData; future oedit
// stashes *ObjIndexData; future medit stashes *MobIndexData. The parser
// type-asserts on entry.
type OlcData struct {
	// Mode is the current menu state (REDIT_MAIN_MENU, REDIT_NAME, ...).
	// Per-editor enum spaces are offset so values don't collide.
	Mode int

	// Vnum is the room/obj/mob vnum being edited. Cached here so the
	// double-edit guard and olc-log helper don't have to re-derive it
	// from Target.
	Vnum int

	// Change is a dirty-bit reserved for the save-confirm flow. The
	// flow itself is scope-cut (plan §Scope Cuts) but the field stays
	// for consumers that check "did anything change?" semantics.
	Change bool

	// Target holds the primary object under edit (room for CON_REDIT).
	// Equivalent to C's ch->dest_buf.
	Target any

	// Spare holds a secondary stash — the currently-selected exit or
	// extradesc. Equivalent to C's ch->spare_ptr.
	Spare any

	// TempNum is a scratch integer for two-step prompts (e.g. during
	// exit-add the direction is stashed here between the first and
	// second prompts). Equivalent to C's ch->tempnum.
	TempNum int
}

// REDIT_* modes for OlcData.Mode while d.Connected == CON_REDIT.
// Mirror C src/olc.h REDIT_* enum. Offset from zero so future OEDIT_* /
// MEDIT_* blocks can occupy later ranges without ambiguity.
const (
	REDIT_MAIN_MENU = iota + 100
	REDIT_NAME
	REDIT_DESC
	REDIT_FLAGS
	REDIT_SECTOR
	REDIT_TUNNEL
	REDIT_TELEDELAY
	REDIT_TELEVNUM
	REDIT_EXIT_MENU
	REDIT_EXIT_EDIT
	REDIT_EXIT_ADD
	REDIT_EXIT_ADD_VNUM
	REDIT_EXIT_DELETE
	REDIT_EXIT_VNUM
	REDIT_EXIT_KEY
	REDIT_EXIT_KEYWORD
	REDIT_EXIT_DESC
	REDIT_EXIT_FLAGS
	REDIT_EXTRADESC_MENU
	REDIT_EXTRADESC_CHOICE
	REDIT_EXTRADESC_KEY
	REDIT_EXTRADESC_DESCRIPTION
	REDIT_EXTRADESC_DELETE
	REDIT_CONFIRM_SAVESTRING
)
