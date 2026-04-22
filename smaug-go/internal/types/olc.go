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

// OEDIT_* modes for OlcData.Mode while d.Connected == CON_OEDIT.
// Mirror C src/olc.h OEDIT_* enum (:114-150). Offset iota+200 to stay
// clear of CON_* (iota 0..~30) and REDIT_* (iota 100..124). Every OEDIT_*
// value is strictly greater than every REDIT_* value.
//
// The 37-entry block preserves C's ordering including the reserved
// iota slots (OEDIT_CONFIRM_SAVEDB / OEDIT_CONFIRM_SAVESTRING are
// declared-but-unused in stock C; OEDIT_AFFECT_RIS is declared-but-
// unreachable; OEDIT_MPROGS* are scope-cut to plan-phase6-olc-mpedit.md
// but their iota slots are reserved here so mpedit does not have to
// renumber — the plan guarantees OEDIT_MPROGS == OEDIT_MAIN_MENU + 32).
//
// Plan: plan-phase6-olc-oedit.md §G1.
const (
	OEDIT_MAIN_MENU = iota + 200
	OEDIT_EDIT_NAMELIST
	OEDIT_SHORTDESC
	OEDIT_LONGDESC
	OEDIT_ACTDESC
	OEDIT_TYPE
	OEDIT_EXTRAS
	OEDIT_WEAR
	OEDIT_WEIGHT
	OEDIT_COST
	OEDIT_COSTPERDAY
	OEDIT_TIMER
	OEDIT_VALUE_1
	OEDIT_VALUE_2
	OEDIT_VALUE_3
	OEDIT_VALUE_4
	OEDIT_VALUE_5
	OEDIT_VALUE_6
	OEDIT_EXTRADESC_KEY
	OEDIT_CONFIRM_SAVEDB     // reserved; unused in stock C
	OEDIT_CONFIRM_SAVESTRING // reserved; commented-out in C
	OEDIT_EXTRADESC_DESCRIPTION
	OEDIT_EXTRADESC_MENU
	OEDIT_LEVEL
	OEDIT_LAYERS
	OEDIT_AFFECT_MENU
	OEDIT_AFFECT_LOCATION
	OEDIT_AFFECT_MODIFIER
	OEDIT_AFFECT_REMOVE
	OEDIT_AFFECT_RIS // declared in shipped C but unreachable code path
	OEDIT_EXTRADESC_CHOICE
	OEDIT_EXTRADESC_DELETE
	OEDIT_MPROGS        // reserved for plan-phase6-olc-mpedit.md
	OEDIT_MPROGS_CHOICE // reserved for mpedit
	OEDIT_MPROGS_DELETE // reserved for mpedit
	OEDIT_MPROGS_TYPE   // reserved for mpedit
	OEDIT_MPROGS_ARG    // reserved for mpedit
)

// MEDIT_* modes for OlcData.Mode while d.Connected == CON_MEDIT.
// Mirror C src/olc.h MEDIT_* enum at :198-263. Offset iota+300 so every
// MEDIT_* value is strictly greater than every OEDIT_* value (which live
// at iota+200, max at OEDIT_MPROGS_ARG). REDIT_* live at iota+100.
//
// 64 consecutive iota slots. Plan: plan-phase6-olc-medit.md §G1.
//
// Mode ownership (NPC-only / PC-only / shared) and digit mapping from C
// main menus is documented in the plan §"PC-vs-NPC mode ownership" table
// — this block is purely the value enumeration.
const (
	MEDIT_NPC_MAIN_MENU      = iota + 300 // 0 in C — shown for IS_NPC
	MEDIT_PC_MAIN_MENU                    // 1 — shown for connected PC
	MEDIT_NAME                            // 2
	MEDIT_S_DESC                          // 3
	MEDIT_L_DESC                          // 4
	MEDIT_D_DESC                          // 5
	MEDIT_NPC_FLAGS                       // 6 — ACT_* bitmask editor
	MEDIT_PC_FLAGS                        // 7 — PLR_* bitmask editor
	MEDIT_AFF_FLAGS                       // 8 — AFF_* bitmask editor
	MEDIT_CONFIRM_SAVESTRING              // 9 — PC exit save-confirm
	MEDIT_SEX                             // 10
	MEDIT_HITROLL                         // 11
	MEDIT_DAMROLL                         // 12
	MEDIT_DAMNUMDIE                       // 13
	MEDIT_DAMSIZEDIE                      // 14
	MEDIT_DAMPLUS                         // 15
	MEDIT_HITNUMDIE                       // 16
	MEDIT_HITSIZEDIE                      // 17
	MEDIT_HITPLUS                         // 18
	MEDIT_AC                              // 19
	MEDIT_GOLD                            // 20
	MEDIT_POS                             // 21
	MEDIT_DEFAULT_POS                     // 22 — NPC-only default posture
	MEDIT_ATTACK                          // 23 — NPC-only attack index
	MEDIT_DEFENSE                         // 24 — NPC-only defense flags
	MEDIT_LEVEL                           // 25 — LEVEL_GREATER for PC target
	MEDIT_ALIGNMENT                       // 26
	MEDIT_STRENGTH                        // 27
	MEDIT_INTELLIGENCE                    // 28
	MEDIT_WISDOM                          // 29
	MEDIT_DEXTERITY                       // 30
	MEDIT_CONSTITUTION                    // 31
	MEDIT_CHARISMA                        // 32
	MEDIT_LUCK                            // 33
	MEDIT_CLAN                            // 34
	MEDIT_DEITY                           // 35
	MEDIT_COUNCIL                         // 36
	MEDIT_SPEC                            // 37 — NPC-only spec_fun name
	MEDIT_RESISTANT                       // 38 — RIS_* bitmask editor
	MEDIT_IMMUNE                          // 39 — RIS_* bitmask editor
	MEDIT_SUSCEPTIBLE                     // 40 — RIS_* bitmask editor
	MEDIT_PCDATA_FLAGS                    // 41 — PCFLAG_* PC-only
	MEDIT_MENTALSTATE                     // 42
	MEDIT_EMOTIONAL                       // 43
	MEDIT_THIRST                          // 44
	MEDIT_FULL                            // 45
	MEDIT_DRUNK                           // 46
	MEDIT_PARTS                           // 47 — PART_* bitmask editor
	MEDIT_FAVOR                           // 48
	MEDIT_HITPOINT                        // 49
	MEDIT_MANA                            // 50
	MEDIT_MOVE                            // 51
	MEDIT_PRACTICE                        // 52 — PC-only
	MEDIT_PASSWORD                        // 53 — PC-only (bcrypt path)
	MEDIT_SAVE_MENU                       // 54 — PC-only 5-way sub-dispatch
	MEDIT_SAV1                            // 55
	MEDIT_SAV2                            // 56
	MEDIT_SAV3                            // 57
	MEDIT_SAV4                            // 58
	MEDIT_SAV5                            // 59
	MEDIT_CLASS                           // 60 — PC-only LEVEL_GREATER
	MEDIT_RACE                            // 61 — PC-only LEVEL_GREATER
	MEDIT_SILVER                          // 62
	MEDIT_COPPER                          // 63

	// Wave 4 additions — affect-list editor (G10). C medit has no
	// MEDIT_AFFECT_* constants of its own (the AFF_FLAGS bitmask landed in
	// Wave 3 covers affected_by toggles); the affect-list ADD/REMOVE flow
	// is a Go-port enhancement, mirroring the OEDIT_AFFECT_* family in the
	// iota+200 range. Allocated in iota+300 so the medit dispatcher stays
	// cleanly typed and so per-arm bodies operate on *CharData.Affects
	// rather than *ObjIndexData.Affects.
	MEDIT_AFFECT_MENU     // 64
	MEDIT_AFFECT_LOCATION // 65
	MEDIT_AFFECT_MODIFIER // 66
	MEDIT_AFFECT_REMOVE   // 67
)
