# Plan: Phase 6 OLC `mpedit` / `opedit` / `rpedit` — Mob/Obj/Room MUD-Program Editors

**Status:** Authored 2026-04-26. Pending adversary review before dispatch.
**Priority:** P2 (Phase 6, Wave D — Interactive OLC, last of three editor lineages: redit / oedit / medit / **mpedit-trio**).
**Scope:** Port C `do_mpedit` / `do_opedit` / `do_rpedit` (`src/build.c:9039-10058`, ~1015 LOC of which ~970 are the three editors plus the `mpedit` and `rpedit` body-helpers) to Go. Today the Go side at `internal/act/olc_prog.go` ships a **read-only inspector** — this plan extends it to the full **add / delete / insert / edit / list** mutator surface, integrating the string editor through the Tier-12 `EditorSave` callback pattern.

---

## Cross-Plan Dependencies

**Depends on (LANDED):**

- `plan-phase5-tier12-editor-save.md` (commit landed 2026-04-18). Provides `CharData.EditorSave func(*CharData)` and the `/s` handler at `internal/game/editor.go:159-173` that transitions `Connected=CON_PLAYING` BEFORE invoking the closure. Mpedit's "open the string editor on add/insert/edit" path is a direct moral port of C's `start_editing(ch, mprg->comlist)` via this callback.
- `internal/act/olc.go:16-28` — `StartEditingFunc` / `CopyBufferFunc` / `StopEditingFunc` cross-package seams (boot-wired in `internal/boot/boot.go:144-148`).
- `internal/persist/area.go:505-514, 750-758, 895-903` — area-loader path that calls `pIndex.ProgTypes.Set(bits.TrailingZeros64(uint64(prog.Type)))`. Establishes the Go convention: `MProgData.Type` is the **bit-flag value** (`1<<n`), `ProgTypes` BitVector indexes by **bit-position** (`bits.TrailingZeros64`). Mpedit's progtypes-rebuild (G7) must follow this convention exactly.
- `internal/act/olc_prog.go:1-253` — current read-only inspector. Extend in place; do NOT introduce a parallel surface.

**Soft dependencies (verify at G-start):**

- `handler.GetCharRoom` (`internal/handler/find.go:14`) — in-room mob lookup by name; reused for trust < LEVEL_GOD path.
- `handler.GetCharWorld` (`internal/handler/find.go:52`) — world-wide mob lookup; reused for trust >= LEVEL_GOD path.
- `handler.GetObjCarry` (`internal/handler/find.go:91`) — inventory lookup; reused for opedit's trust < LEVEL_GOD path.
- `handler.GetObjWorld` (`internal/handler/find.go:183`) — world-wide object lookup; reused for opedit's trust >= LEVEL_GOD path.
- `util.SmashTilde` (`internal/util/strings.go:46`) — strip `~` from raw arglist before storage (matches C `smash_tilde(argument)` at top of each editor).
- `util.OneArgument` — already used by the inspector at `olc_prog.go:115`.

**Blocks:** nothing yet authored. mpedit closes Phase-6 Wave-D's interactive-OLC stripe.

**Does NOT depend on:** `OlcData` / `CON_*EDIT` loop-arm machinery. **mpedit is NOT a menu substate** — see §Go Design "No CON_MPEDIT".

---

## Problem

The Go side ships a **read-only inspector** at `internal/act/olc_prog.go:96-162`. It accepts `mpedit <vnum> [trigger]`, prints existing progs, and emits the trailing line "(read-only inspector; full string-editor integration is a follow-up)". The inspector has zero mutation surface — builders cannot add, delete, insert, or edit a prog through Go today.

The C reference at `src/build.c:9039-9376` (`do_mpedit`, mob version, 338 LOC) implements a **stateless dispatcher**:

- Syntax: `mpedit <victim> <command> [number] <program> <value>`
- Subcommands: `list` / `add` / `insert` / `edit` / `delete`
- For `add` / `insert` / `edit`, the helper `mpedit(ch, mprg, mptype, argument)` at `:9018-9033` mutates the prog's `type` + `arglist` and calls `start_editing(ch, mprg->comlist)` — **the editor takes over from there**. On `/s`, the substate `SUB_MPROG_EDIT` re-enters `do_mpedit`, which copies the buffer into `mprg->comlist` and calls `stop_editing(ch)`. No menu loop; no CON_*EDIT state.

`do_opedit` (`:9379-9719`) is the same shape, swapping `victim->pIndexData->mudprogs` for `obj->pIndexData->mudprogs` and the trust-gated lookup from `get_char_room` / `get_char_world` to `get_obj_carry` / `get_obj_world`.

`do_rpedit` (`:9745-10058`) is a third near-clone. It targets `ch->in_room->mudprogs` directly (no victim arg) and shifts the argument indices: subcommand is `arg1`, value is `atoi(arg2)`. The substate path is identical (`SUB_MPROG_EDIT` → reuse `mpedit` body-helper).

The work is entirely a **mechanical port** of three near-identical dispatchers, sharing the `mpedit` body-helper (which becomes a single Go function), preserving per-editor argument shape (mob+obj take a target name; room takes none), and wiring the EditorSave callback to copy the editor buffer into `MProgData.ComList` on `/s`.

---

## C Reference (authoritative)

All line numbers verified 2026-04-26 from `src/build.c`.

### `do_mpedit` — `src/build.c:9039-9376` (338 LOC)

| Step | Lines | Action |
|---|---|---|
| NPC reject | 9051-9055 | `IS_NPC(ch)` → "Mob's can't mpedit" |
| Descriptor guard | 9057-9061 | `!ch->desc` → "You have no descriptor" |
| Substate re-entry | 9063-9081 | `SUB_MPROG_EDIT` branch: copy buffer into `mprog->comlist`, `stop_editing`. Includes a defensive `!ch->dest_buf` guard with `bug()` log. **NB: lines 9080-9081 contain the `return;` — there is NO unreachable code here despite earlier reports.** |
| Tilde-smash + arg parse | 9083-9087 | `smash_tilde(argument)`; three `one_argument` calls into `arg1` (victim), `arg2` (subcommand), `arg3` (number-or-program). `value = atoi(arg3)`. |
| Syntax help | 9088-9102 | If `arg1` or `arg2` empty, print syntax + subcommand list + program-type list (mob-prog flavors). |
| Victim resolution | 9104-9119 | `get_trust(ch) < LEVEL_GOD` → `get_char_room(ch, arg1)`; else `get_char_world(ch, arg1)`. NULL → message + return. |
| Trust + NPC gates | 9121-9133 | `get_trust(ch) < victim->level || !IS_NPC(victim)` → "You can't do that!". `< LEVEL_GREATER && IS_NPC && ACT_STATSHIELD` → "Their godly glow prevents you...". |
| ACL | 9134-9135 | `!can_mmodify(ch, victim)` → return. |
| Prototype gate | 9137-9142 | `!ACT_PROTOTYPE` → "A mobile must have a prototype flag to be mpset.". |
| `mprog = victim->pIndexData->mudprogs;` | 9144 | Set list-head pointer for subcommand bodies. |
| `list` | 9148-9201 | Empty-list message; `value < 1` branch (full list when `arg3=="full"`, header-only otherwise); `value >= 1` branch (single prog by index). Index is 1-based. |
| `edit` | 9203-9242 | Optional `arg4` = new program type via `get_mpflag`; `value < 1` reject; iterate to `value`-th prog; call `mpedit(ch, mprg, mptype, argument)` (which opens the editor); rebuild `progtypes` bitmask via `xCLEAR_BITS` + iteration. |
| `delete` | 9244-9306 | Iterate to `value`-th prog → store its type as `mptype`; count how many progs share that type (`num`); first-prog vs nth-prog unlink branches; `STRFREE`/`DISPOSE`; `if (num <= 1) xREMOVE_BIT(progtypes, mptype)`. |
| `insert` | 9308-9352 | Insert before position `value`. `arg4` = program type. `value == 1` → splice at head. `value > 1` → walk to position, splice after `value`-th. Calls `mpedit(...)` to open editor. **Note `arg4` is the prog type for insert (not `arg3`).** |
| `add` | 9354-9373 | Append. `arg3` is program type. Walks to tail via `for (; mprog->next; mprog = mprog->next);`. Allocates, links, sets progtypes bit, calls `mpedit(...)` to open editor, sets `mprg->next = NULL` (defensive — line 9371 is reachable; the only code AFTER `mpedit` returns from start_editing's setup). |
| Default | 9375 | `do_mpedit(ch, "")` — recursive call with empty args, falls into syntax-help branch. |

### `do_opedit` — `src/build.c:9378-9719` (342 LOC)

Same shape as `do_mpedit` with these divergences:

| Step | Lines | Divergence |
|---|---|---|
| NPC reject | 9391-9395 | "Mob's can't opedit". |
| Substate re-entry | 9403-9421 | **C-bug:** the `bug()` message at line 9411 says `"do_opedit: sub_oprog_edit: NULL ch->dest_buf"` — the substate is `SUB_MPROG_EDIT` (shared with mob), not `SUB_OPROG_EDIT`. Cosmetic only; preserve verbatim. |
| Syntax help | 9429-9444 | Object-prog flavor list (act/speech/rand/wear/remove/sac/zap/get/drop/damage/repair/greet/exa/use/pull/push). Adds tail line "Object should be in your inventory to edit." |
| Target resolution | 9446-9461 | `get_obj_carry(ch, arg1)` for trust < LEVEL_GOD; `get_obj_world(ch, arg1)` otherwise. |
| ACL | 9463-9464 | `!can_omodify(ch, obj)` → return. **No level/STATSHIELD gates** (objects don't have those flags). |
| Prototype gate | 9466-9471 | `IS_OBJ_STAT(obj, ITEM_PROTOTYPE)`. |
| `mprog = obj->pIndexData->mudprogs;` | 9473 | Same shape. |
| `list` / `edit` / `delete` / `insert` / `add` | 9477-9716 | Identical to mpedit modulo `victim->pIndexData → obj->pIndexData`. **C-bug at line 9411** message string preserved. |
| Default | 9718 | Recursive empty-arg call. |

### `do_rpedit` — `src/build.c:9744-10058` (315 LOC)

Same shape, room-targeted:

| Step | Lines | Divergence |
|---|---|---|
| NPC reject | 9755-9759 | "Mob's can't rpedit". |
| Substate re-entry | 9767-9785 | **C-bug:** the `bug()` message at line 9775 says `"do_opedit: sub_oprog_edit: NULL ch->dest_buf"` — copy-paste typo from `do_opedit`. Preserve verbatim with a `// C-bug` comment. |
| Argument shape | 9787-9790 | **DIFFERS:** only TWO `one_argument` calls (no victim). `arg1` = subcommand; `arg2` = number-or-program. `value = atoi(arg2)`. The `arg3` declared at line 9749 is unused at this point (read into `arg3` only inside subcommand bodies). |
| Syntax help | 9793-9807 | Room-prog flavor list (act/speech/rand/sleep/rest/rfight/enter/leave/death). Adds "You should be standing in room you wish to edit." |
| Target resolution | 9809-9810 | `!can_rmodify(ch, ch->in_room)` → return. **Target is `ch->in_room` directly — no name argument, no level/proto gate (rooms have no proto flag).** |
| `mprog = ch->in_room->mudprogs;` | 9812 | Same shape. |
| `list` / `edit` / `delete` / `add` | 9816-10055 | Same algorithm; arg index shifted (use `arg2` where mpedit/opedit use `arg3`, etc.). |
| **C-bug — `insert` test on `arg2`** | 9991 | The `insert` branch is gated by `if (!str_cmp(arg2, "insert"))` instead of `arg1`. **`arg2` is the value/number string at this point — `insert` is unreachable in stock C rpedit.** Document and preserve verbatim with a `// C-bug: dead branch` comment, OR fix in Go (recommended; Open Q1). |
| **C-bug — `insert` body uses `get_mpflag(arg2)`** | 9999 | Even if the gate fired, `arg2` is the value string, not the prog type. The argument is read into `arg3` at line 9998 but `arg2` is passed to `get_mpflag`. Document. |
| Default | 10057 | Recursive empty-arg call. |

### `mpedit` body-helper — `src/build.c:9018-9033` (16 LOC)

Shared by all five dispatch arms (mob/obj/room × edit/insert/add). The `rpedit` helper at `:9727-9742` is a verbatim duplicate of `mpedit`, called from `do_rpedit` body — but the Go port collapses to a single helper.

```c
void mpedit (CHAR_DATA *ch, MPROG_DATA *mprg, int mptype, char *argument)
{
  if (mptype != -1) {
    mprg->type = mptype;
    if (mprg->arglist) STRFREE (mprg->arglist);
    mprg->arglist = STRALLOC (argument);
  }
  ch->substate = SUB_MPROG_EDIT;
  ch->dest_buf = mprg;
  if (!mprg->comlist)
    mprg->comlist = STRALLOC ("");
  start_editing (ch, mprg->comlist);
}
```

Translation: optionally rewrite `type` + `arglist`; mark substate; stash the prog pointer on `dest_buf`; open the editor seeded with `comlist`. Go port: stash via closure capture (no `dest_buf`; the closure binds `mprg` directly), set `Substate = SUB_MPROG_EDIT`, install `EditorSave` to write `mprg.ComList = CopyBufferFunc(c)` and `StopEditingFunc(c)`, then call `StartEditingFunc(ch, mprg.ComList)`.

### `mprog_flags[]` table — `src/build.c:366-374` (52 entries)

Used by `get_mpflag` at `:721-728`. The function returns the **index** in the table (which IS the bit position for `xSET_BIT(progtypes, idx)`). The first 32 entries map to `MPROG_*` bit-flag values 0..31; entries 32+ map to higher bits per the Go-port enums in `internal/types/mudprog.go:49-90`.

```
0  act        16 look       32 load       (nb: load_prog is not ported in Go)
1  speech     17 exa        33 login
2  rand       18 zap        34 void
3  fight      19 get        35 tell
4  death      20 drop       36 imminfo
5  hitprcnt   21 damage     37 greetinfight (nb: not in Go enum)
6  entry      22 repair     38 move          (nb: not in Go enum)
7  greet      23 randiw     39 command
8  allgreet   24 speechiw   40 sell
9  give       25 pull       41 emote         (nb: not in Go enum)
10 bribe     26 push       42-50 r1..r10   (reserved)
11 hour      27 sleep
12 time      28 rest
13 wear      29 leave
14 remove    30 script
15 sac       31 use
```

Go port: introduce `util.MProgFlagNames []string` (or extend the existing `triggerBitFromName` map at `internal/act/olc_prog.go:15-57` with a parallel index-ordered slice). The lookup result is the **bit-flag value** (`1 << idx` for low bits, `1 << 32+offset` for the high band) — see `internal/types/mudprog.go:49-90`. The plan defines a single `getMpFlag(name string) (flagBit int64, ok bool)` helper that returns the bit-flag value.

---

## Go Current State

Verified 2026-04-26.

| Artifact | Location | Status |
|---|---|---|
| `DoMpedit` / `DoOpedit` / `DoRpedit` (read-only) | `internal/act/olc_prog.go:96-108` | Inspector only — no add/delete/insert/edit. **Replace in place** with full mutator surface. |
| `mudprogEdit(ch, argument, kind)` shared helper | `internal/act/olc_prog.go:110-162` | Inspector dispatch. Refactor into a per-kind dispatcher with subcommand parsing. |
| `triggerBitFromName` map | `internal/act/olc_prog.go:15-57` | Maps name → bit-flag value. Reuse for `getMpFlag`. |
| `parseProgTriggerName` | `internal/act/olc_prog.go:59-69` | Already strips `_prog` / `prog` suffix. Reuse. |
| `firstTriggerName(mask int64) string` | `internal/act/olc_prog.go:201-252` | Reverse lookup for list rendering. Reuse. |
| `MProgData{Type, Triggered, ResetDelay, ArgList, ComList}` | `internal/types/mudprog.go:9-15` | Existing shape — no new fields needed. |
| `MobIndexData.MudProgs []*MProgData` + `ProgTypes BitVector` | `internal/types/mob_index.go:11-12` | Reuse. |
| `ObjIndexData.MudProgs` + `ProgTypes` | `internal/types/object.go:8-9` | Reuse. |
| `RoomIndexData.MudProgs` + `ProgTypes` | `internal/types/room.go:31-33` | Reuse. |
| `SUB_MPROG_EDIT` constant | `internal/types/enums.go:112` | Already defined. |
| `EditorSave` callback | `internal/types/character.go:60` | Already defined. |
| `/s` handler | `internal/game/editor.go:159-173` | Already invokes the closure after `Connected → CON_PLAYING`. |
| `StartEditingFunc` / `CopyBufferFunc` / `StopEditingFunc` | `internal/act/olc.go:18-27` | Boot-wired. Reuse. |
| `WorldRef *world.World` | `internal/act/olc.go` (boot-set) | Reuse for index lookups. |
| `handler.GetCharRoom` / `GetCharWorld` | `internal/handler/find.go:14, 52` | Reuse for mpedit victim resolution. |
| `handler.GetObjCarry` / `GetObjWorld` | `internal/handler/find.go:91, 183` | Reuse for opedit object resolution. |
| `util.SmashTilde` | `internal/util/strings.go:46` | Reuse for arglist sanitization. |
| `util.OneArgument` | `internal/util/strings.go` | Reuse for arg parsing. |
| `mprog_flags[]` index table | **MISSING** | G1 deliverable: `util.MProgFlagNames []string` ordered to match C `mprog_flags[]` for keyword lookup. |
| `getMpFlag(name string)` helper | **MISSING** | G1 deliverable. Returns bit-flag value (NOT bit-index). |
| `can_mmodify` / `can_omodify` / `can_rmodify` | **MISSING** | Not ported in Go. Use `GetTrust() >= LEVEL_IMMORTAL` precedent (already enforced via command registration at `internal/boot/boot.go:792-794`). Document deferral. |
| Boot registration | `internal/boot/boot.go:792-794` | mpedit/opedit/rpedit already registered at LEVEL_IMMORTAL; **no change needed**. |
| `internal/act/olc_prog_test.go` | 148 LOC, 9 tests | Inspector-only; some tests will need to evolve as the inspector dispatch path changes. |

---

## Go Design

### No CON_MPEDIT — this is NOT a menu substate

C `do_mpedit` is **not** a menu loop. It is a one-shot dispatcher whose `add` / `insert` / `edit` arms call `start_editing` directly. Re-entry happens via `SUB_MPROG_EDIT` at the top of `do_mpedit` itself (the C interpreter's `do_last_cmd` path). There is no `CON_MPEDIT` constant in `src/olc.h`.

The Go port **does NOT** introduce:

- A new `CON_MPEDIT` enum value.
- A new `case types.CON_MPEDIT:` arm in `internal/game/loop.go`.
- An `OlcData.Mode` iota+400 block.
- A `MeditDispMenuFunc`-style seam.
- Any per-descriptor menu state.

Instead, the Go port uses the already-shipped **EditorSave callback pattern** (Tier 12). Each `add` / `insert` / `edit` arm:

1. Mutates the prog's `Type` + `ArgList` synchronously.
2. Captures the `*MProgData` pointer in a closure.
3. Sets `ch.Substate = SUB_MPROG_EDIT` (purely advisory in Go — the C role of triggering re-entry into `do_mpedit` is replaced by the closure).
4. Installs `ch.EditorSave = func(c *CharData) { mprg.ComList = CopyBufferFunc(c); StopEditingFunc(c); /* optional: rebuild progtypes */ }`.
5. Calls `StartEditingFunc(ch, mprg.ComList)`.

On `/s`, the editor handler at `internal/game/editor.go:159-173` already transitions `Connected=CON_PLAYING` and invokes the closure. The closure writes the buffer into `mprg.ComList` and calls `StopEditingFunc` (which resets `Substate=SUB_NONE` per `internal/game/editor.go:88`).

This is the same pattern `redit desc` and `redit ed` already use (`internal/act/olc.go:99-115, 174-190`).

### Victim / target resolution: match C verbatim

`do_mpedit` takes a **victim mob name** (resolved via `get_char_room` / `get_char_world`). The current Go inspector takes a **vnum integer**. The plan ships **C-faithful name resolution** as the new shape:

- Subcommand syntax: `mpedit <victim> <command> [number] <program> <value>`.
- For trust < LEVEL_GOD, restrict to in-room mobs (`handler.GetCharRoom`).
- For trust >= LEVEL_GOD, search world-wide (`handler.GetCharWorld`).

Rationale: muscle memory transfer for builders coming from any C SMAUG fork; precedent set by every other ported builder command; the inspector's `mpedit <vnum>` form is a Go-port deviation that has no C analog and sees minimal use (no usage telemetry, but the inspector's tests are the only callers).

**The vnum-form is replaced, not preserved.** A builder can still inspect via `mpedit <victim> list` (which lists all progs on the victim's prototype). Any test asserting `mpedit <vnum>` behavior is rewritten to use `mpedit <victim> list`.

opedit and rpedit follow the same C shape:

- `opedit <object> <command> [number] <program> <value>` — `get_obj_carry` for trust < LEVEL_GOD, `get_obj_world` otherwise.
- `rpedit <command> [number] <program> <value>` — operates on `ch.InRoom` directly.

### Dispatch shape

A single shared mutator helper, called by three thin per-kind dispatchers (mob / obj / room), parameterized by which field-set the subcommand mutates and which gates apply.

```go
// internal/act/olc_prog.go (replacing the inspector dispatch)

type progEditorTarget struct {
    kind        string             // "mob" | "obj" | "room"
    label       string             // "Mob" | "Object" | "Room" — for messages
    progs       *[]*types.MProgData // pointer-to-slice so mutations are visible
    progTypes   *types.BitVector   // pointer to the index's bitmask
    targetName  string             // for log + display
    targetVnum  int                // for log + display
}

func DoMpedit(ch *types.CharData, argument string) { mpedit_dispatch(ch, argument, "mob") }
func DoOpedit(ch *types.CharData, argument string) { mpedit_dispatch(ch, argument, "obj") }
func DoRpedit(ch *types.CharData, argument string) { mpedit_dispatch(ch, argument, "room") }

func mpedit_dispatch(ch *types.CharData, argument string, kind string) {
    // 1. Common gates: NPC reject, descriptor check, substate re-entry handling.
    // 2. Per-kind argument parse (mob/obj take target name; room does not).
    // 3. Per-kind target resolution + ACL + prototype gate.
    //    NB: handler.GetCharWorld and handler.GetObjWorld both require
    //    *world.World as their first arg — pass act.WorldRef (the
    //    package-level boot-set seam, same pattern as wiz.go / arena.go).
    //    GetCharRoom and GetObjCarry do NOT need WorldRef.
    // 4. Build progEditorTarget and dispatch on subcommand.
}
```

The substate re-entry path is handled by the EditorSave closure, NOT by a `switch ch.Substate` at function entry. This is the cleanest Go-idiomatic translation: the closure captures everything the C re-entry path needed `dest_buf` for. We document the divergence at the top of the file.

### Argument-shape divergence between mpedit/opedit and rpedit

C: mpedit/opedit consume `arg1=victim`, `arg2=subcommand`, `arg3=number-or-program`. rpedit consumes `arg1=subcommand`, `arg2=number-or-program` (no victim). The Go port mirrors this — the per-kind argument parser branches on `kind == "room"` for the room shape.

Rationale for not unifying: the arg shape is part of the user-visible contract; preserving it is fidelity. The Go dispatcher fans out to two parsers (mob-or-obj vs room) before reaching the subcommand body.

### Subcommand body semantics (shared across kinds)

| Subcommand | Behavior | Editor-save callback? |
|---|---|---|
| `list` | Print all progs (header-only, or `full` for body), or single prog by index. | No |
| `add <type> <args>` | Append new prog at tail; set `progTypes` bit; open editor seeded with empty `ComList`. | YES |
| `insert <pos> <type> <args>` | Splice before position; set bit; open editor. | YES |
| `edit <pos> [type] <args>` | Mutate prog at position (optional new type, always new args); rebuild full progTypes bitmask via iteration; open editor. | YES |
| `delete <pos>` | Remove prog at position; if the removed type has zero remaining progs, clear that progTypes bit. | No |

For the editor-save callback:
- The closure captures `*MProgData` (the prog being edited).
- On `/s`, `mprg.ComList = CopyBufferFunc(c)` writes the new body.
- `StopEditingFunc(c)` resets editor state.
- For `edit` (only), the closure ALSO rebuilds the progtypes bitmask after copy — matching C `:9234-9236` (`xCLEAR_BITS` then iterate). For `add` / `insert` the bit was already set synchronously before opening the editor.

### Progtypes bitmask convention

`MProgData.Type` is a **bit-flag value** (`1 << n` for low bits; `1 << 32+offset` for high bits per `internal/types/mudprog.go:85-90`). `ProgTypes` BitVector indexes by **bit-position** (`bits.TrailingZeros64(uint64(prog.Type))`). The persist loader at `internal/persist/area.go:513` already follows this convention; mpedit must too.

For `delete`, the bit-clear is conditional on "zero remaining progs of this type":

```go
removedType := prog.Type
remaining := 0
for _, p := range *target.progs {
    if p.Type == removedType {
        remaining++
    }
}
// (deletion happens here)
if remaining <= 1 {  // <=1 because we count BEFORE the delete
    target.progTypes.Remove(bits.TrailingZeros64(uint64(removedType)))
}
```

For `edit`, the C path clears all bits and re-iterates — same in Go. For `add` / `insert`, the new type's bit is set synchronously.

### EditorSave closure shape

```go
func mpeditOpenEditor(ch *types.CharData, mprg *types.MProgData, target progEditorTarget, rebuildOnSave bool) {
    if mprg.ComList == "" {
        mprg.ComList = ""  // C: STRALLOC("") — Go's zero-string is fine
    }
    ch.Substate = types.SUB_MPROG_EDIT
    captured := mprg
    capturedTarget := target
    rebuild := rebuildOnSave
    ch.EditorSave = func(c *types.CharData) {
        if CopyBufferFunc != nil {
            captured.ComList = CopyBufferFunc(c)
        }
        if StopEditingFunc != nil {
            StopEditingFunc(c)
        }
        if rebuild {
            capturedTarget.progTypes.Clear()
            for _, p := range *capturedTarget.progs {
                capturedTarget.progTypes.Set(bits.TrailingZeros64(uint64(p.Type)))
            }
        }
        c.Send("\n\r")
    }
    if StartEditingFunc != nil {
        StartEditingFunc(ch, mprg.ComList)
    }
}
```

The `rebuild` flag is only `true` on `edit` (matching C). On `add` / `insert`, the bit was set synchronously before the editor opened, so no rebuild is needed.

### Trust gating

The `command.Registry` already gates all three at `LEVEL_IMMORTAL` (`internal/boot/boot.go:792-794`). Per-arm gates inside the dispatcher:

- mpedit: `LEVEL_GOD` for world-wide victim search; `LEVEL_GREATER` for editing ACT_STATSHIELD mobs.
- opedit: `LEVEL_GOD` for world-wide object search.
- rpedit: no per-arm trust gate beyond the registry's LEVEL_IMMORTAL (rooms have no STATSHIELD).

`can_mmodify` / `can_omodify` / `can_rmodify` are NOT ported. The Go inspector already uses `GetTrust() >= LEVEL_IMMORTAL` (`olc_prog.go:111`). Document the deferral in §Scope Cuts.

### C-bug catalog (preserve verbatim unless Open-Q resolves to fix)

1. **`do_opedit` SUB_MPROG_EDIT bug message says "sub_oprog_edit"** (`build.c:9411`). Cosmetic. **Preserve verbatim.** Same in `do_rpedit` at `:9775` (also says "do_opedit: sub_oprog_edit:" — copy-paste typo). **Preserve verbatim** with a `// C-bug: copy-paste from do_opedit` comment.
2. **`do_rpedit` insert-branch gate uses `arg2` instead of `arg1`** (`build.c:9991`). At that point `arg2` is the value/number string, NOT the subcommand. The branch is effectively dead in C (the only way it fires is if a builder typed e.g. `rpedit foo insert ...` where `arg1="foo"` non-matching falls through past `list`/`edit`/`delete`, then `arg2="insert"` accidentally matches). **Open Q1: preserve dead-branch verbatim, OR fix to `arg1`.** Recommendation: **fix in Go** with a `// C-bug fix: was arg2 in C` comment. The user-visible UX is "insert works in rpedit" vs "insert silently does nothing in rpedit" — the latter is clearly unintended.
3. **`do_rpedit` insert-body uses `get_mpflag(arg2)`** (`build.c:9999`). Even if the gate fires, the lookup uses the wrong arg slot. Tied to bug #2. **Same recommendation: fix to `arg3`.**
4. **`do_mpedit` `add` body's trailing `mprg->next = NULL` is reachable but redundant** (`build.c:9371`) — `CREATE` zeroes the struct in C, so `next` is already NULL. Cosmetic. **Preserve verbatim** in Go (zero-value struct). No-op in Go regardless.
5. **`get_char_room` / `get_char_world` victim returns** — C does not check whether the victim's `pIndexData` is non-NULL. Test mobs created in-test without an `IndexData` field would crash at `:9144`. The Go port should add a defensive nil-check on `victim.IndexData`. Document as Go-divergence.

(The original dispatch prompt mentioned "unreachable code after return" in `do_mpedit` and "else obj->pIndexData->mudprogs = mprg" unreachable in `do_opedit`. After re-reading the C carefully, **neither is actually unreachable** — line 9371 is reachable and `if (mprog) mprog->next = mprg; else obj->pIndexData->mudprogs = mprg;` at `:9708-9711` is a normal branch. Catalog corrected.)

---

## Task Groups

Every group follows test-first. Mutation verification uses `Edit`-only revert per the manager's banned-command list (no `git checkout` / `git restore` / `git stash` / `git reset --hard`). Reference: `_shared.md` → Mutation Verification Safety.

### G1 — `mprog_flags` table + `getMpFlag` helper + lookup tests

**Deliverables:**

- `internal/util/mprog_flags.go` (NEW) — `MProgFlagNames []string` (52 entries — indices 0-51, `act` through `r10`) ordered to match C `mprog_flags[]` at `src/build.c:366-374`. Each entry's index IS the bit-position; the bit-flag value is `1 << index` for low bits, with the high-band entries (`load`, `login`, `void`, `tell`, `imminfo`, `greetinfight`, `move`, `command`, `sell`, `emote`) requiring per-entry mapping to the Go-port `MPROG_*` constants since the Go iota at `internal/types/mudprog.go:85-90` skipped some C entries (`load`, `greetinfight`, `move`, `emote`). **Reserved slots `r1..r10` (indices 42-51) MUST be present even though they are unused** — dropping them would misalign every reserved index for future ports.
- `getMpFlag(name string) (bit int64, ok bool)` — returns the **bit-flag value** (matching `MProgData.Type` storage convention). Strips trailing `_prog` / `prog` suffix (reuse `parseProgTriggerName` logic).

**Rationale for two-step:** the C function returns a bit-INDEX; Go stores bit-VALUEs. Returning the value directly lets callers do `mprg.Type = bit` without the index→value conversion.

**Tests (`internal/util/mprog_flags_test.go`, NEW):**

- `TestMProgFlagNames_HasAllCEntries` — table-driven; asserts `len(util.MProgFlagNames) == 52` and each of the 52 C names maps to a defined Go MPROG_* constant OR an explicit "unported" sentinel.
- `TestGetMpFlag_KnownNames` — `act` → `MPROG_ACT`, `speech` → `MPROG_SPEECH`, ..., `login` → `MPROG_LOGIN`, `cmd` (alias for `command`) → `MPROG_CMD`.
- `TestGetMpFlag_StripSuffix` — `act_prog`, `speechprog` resolve identically.
- `TestGetMpFlag_Unknown` — empty / garbage / unported (`load`, `greetinfight`, `move`, `emote`) returns `0, false`.
- `TestGetMpFlag_CaseInsensitive` — `ACT`, `Act`, `act` all match.

**Mutation:** swap `MPROG_ACT` and `MPROG_SPEECH` in the lookup map → `TestGetMpFlag_KnownNames` fails on both. Revert via `Edit`.

**Acceptance:** A1.

### G2 — `getMpFlag` integrated into existing inspector path; tests refactored

**Deliverables:**

- `internal/act/olc_prog.go` — replace `triggerBitFromName` lookups with `util.GetMpFlag`. Keep `parseProgTriggerName` as a thin wrapper that delegates to `util.GetMpFlag` (for backward compatibility with the existing `[trigger]` filter argument the inspector accepts).
- `firstTriggerName` reverse lookup: switch to consulting `util.MProgFlagNames` (so a single edit to add a new mprog flavor flows through both helpers).

**Tests (extending `internal/act/olc_prog_test.go`):**

- Existing inspector tests adapted to new arg shape (mpedit-takes-victim instead of vnum) — see G4 for the new shape.

**Mutation:** drop `_prog` suffix-strip → `TestGetMpFlag_StripSuffix` fails.

**Acceptance:** A2.

### G3 — Common dispatcher skeleton + per-kind argument parser

**Deliverables:**

- `internal/act/olc_prog.go` — `mpedit_dispatch(ch, argument, kind)` replaces `mudprogEdit`.
- Common-gates section: NPC reject, `ch.Desc` nil-guard. (Substate re-entry handled implicitly by EditorSave; the C `switch (ch->substate)` at `:9063-9081` becomes a no-op in Go because `/s` invokes the closure directly. Document the divergence with a comment.)
- Per-kind argument parser:
  - `kind == "mob"` or `kind == "obj"`: `arg1` = target name; `arg2` = subcommand; `arg3` = number-or-program. `value = atoi(arg3)`.
  - `kind == "room"`: `arg1` = subcommand; `arg2` = number-or-program. `value = atoi(arg2)`.
- Syntax-help branch (per kind, matching C help text verbatim).
- Per-kind target resolution + prototype gate (mob: ACT_PROTOTYPE; obj: ITEM_PROTOTYPE; room: no proto gate).
- Per-kind ACL: trust-gated lookup branch (mob/obj LEVEL_GOD bifurcation; room: trust >= LEVEL_IMMORTAL only — already enforced by command registration).
- Per-arm trust gate: mpedit STATSHIELD check at LEVEL_GREATER; mpedit `IS_NPC(victim)` check.
- Build `progEditorTarget` struct, dispatch on subcommand to G4-G8 helpers (which return early; G3 ships only the skeleton + syntax-help + target resolution).

**Tests (`internal/act/olc_prog_test.go`, replacing the inspector test suite):**

- `TestMpeditDispatch_NpcCannotEdit` — `IS_NPC(ch)` rejected with "Mob's can't mpedit".
- `TestMpeditDispatch_NoDesc` — `ch.Desc == nil` rejected with "You have no descriptor".
- `TestMpeditDispatch_SyntaxHelp` — empty args → syntax string contains "Syntax:" + "add delete insert edit list" + per-kind program list.
- `TestMpeditDispatch_VictimNotInRoom` — trust < LEVEL_GOD with offroom victim → "They aren't here." (mob); analogous for obj ("You aren't carrying that.").
- `TestMpeditDispatch_VictimNotInWorld` — trust >= LEVEL_GOD with bogus name → "No one like that in all the realms." / "Nothing like that in all the realms."
- `TestMpeditDispatch_NotPrototype` — non-prototype target → "A mobile must have a prototype flag to be mpset." / object-equivalent.
- `TestMpeditDispatch_StatshieldGate` — trust < LEVEL_GREATER on ACT_STATSHIELD mob → godly-glow message.
- `TestMpeditDispatch_PcVictim` — `!IS_NPC(victim)` → "You can't do that!".
- `TestRpeditDispatch_NoVictimArg` — rpedit has no victim arg; `arg1` parsed as subcommand directly.

**Mutation:** swap `LEVEL_GOD` for `LEVEL_GREATER` in trust branch → cross-room mob-find test fails because it now allows a lower trust. Revert via `Edit`.

**Acceptance:** A3, A4, A5, A6.

### G4 — `list` subcommand

**Deliverables:**

- `mpeditList(ch, target, value, isFull bool)` helper.
- Empty-list branch: "No programs on mobile: %s - #%d" / "That object has no mob programs." (sic — C uses "mob programs" for obj editors at `:9482`) / "That object has no mob programs." (rpedit at `:9821`, also sic).
- `value < 1`: full list (header-only by default, body included when `arg3 == "full"`).
- `value >= 1`: single prog by 1-based index; "Program not found." if out of range.

**Implementation note (Adversary F2):** C at `:9160` writes `if (strcmp("full", arg3))` — `strcmp` returns non-zero when strings DIFFER, so the truthy branch is header-only, the `else` is full-body. **Do NOT port literally as `if util.StrCmp(...)`.** Use Go-readable `if arg3 == "full" { full } else { header-only }` instead. A worker mechanically porting the C `strcmp` will invert the branch.

**Tests:**

- `TestMpeditList_Empty` — no progs on mob → empty-list message.
- `TestMpeditList_HeaderOnly` — three progs → output contains three "%d>%s %s" lines (no body).
- `TestMpeditList_Full` — `mpedit fido list 0 full` → output contains all three bodies.
- `TestMpeditList_ByIndex` — `mpedit fido list 2` → output contains only the second prog's body.
- `TestMpeditList_OutOfRange` — `mpedit fido list 99` → "Program not found.".
- Parallel tests for opedit and rpedit shapes.

**Mutation:** flip 1-based index to 0-based (drop `++cnt` increment timing) → `TestMpeditList_ByIndex` fails.

**Acceptance:** A7, A8.

### G5 — `add` subcommand (with EditorSave closure for prog body)

**Deliverables:**

- `mpeditAdd(ch, target, mptype int64, argument string)` helper.
- Validate `mptype != 0` (the `getMpFlag` error sentinel) → "Unknown program type.".
- Walk target.progs slice to tail; allocate new `*MProgData{Type: mptype, ArgList: util.SmashTilde(argument), ComList: ""}`; append.
- Set `target.progTypes` bit for `bits.TrailingZeros64(uint64(mptype))`.
- Call `mpeditOpenEditor(ch, mprg, target, false /* no rebuild */)`.

**EditorSave closure:** copy buffer to `mprg.ComList`; call `StopEditingFunc`. **Do NOT rebuild progtypes** (bit was set synchronously).

**Tests:**

- `TestMpeditAdd_AppendsAtTail` — start with [act, speech]; add `greet "..."` → progs is [act, speech, greet]; `progTypes` has act+speech+greet bits.
- `TestMpeditAdd_OpensEditor` — capture editor-start: assert `StartEditingFunc` was called with `mprg.ComList` (empty) seed. (Use a test-local `StartEditingFunc` spy.)
- `TestMpeditAdd_EditorSaveWritesComList` — invoke the captured closure with a fake CopyBufferFunc returning `"say hi $n\n\r"`; assert `mprg.ComList == "say hi $n\n\r"`.
- `TestMpeditAdd_UnknownType` — `mpedit fido add bogus_type "..."` → "Unknown program type."; progs unchanged.
- `TestMpeditAdd_SmashesTilde` — `mpedit fido add act "hi~there"` → `mprg.ArgList == "hi-there"`.
- Parallel tests for opedit + rpedit.

**Mutation gates:**

- Replace `append(*progs, mprg)` with prepend → `TestMpeditAdd_AppendsAtTail` fails.
- Replace `progTypes.Set(...)` with `Toggle(...)` → re-add same flavor flips bit off; pin-test `TestMpeditAdd_DoubleSetIdempotent`.

**Acceptance:** A9, A10, A11.

### G6 — `insert` subcommand

**Deliverables:**

- `mpeditInsert(ch, target, value int, mptype int64, argument string)` helper.
- Reject if `target.progs` empty: "No programs on mobile..." (per kind).
- Reject `mptype == 0` → "Unknown program type.".
- Reject `value < 1` → "Program not found.".
- `value == 1`: prepend (splice at head).
- `value > 1`: walk to position `value-1`, splice after. If position not reached: "Program not found.".
- Set `target.progTypes` bit synchronously.
- Open editor via `mpeditOpenEditor(ch, mprg, target, false)`.

**Tests:**

- `TestMpeditInsert_AtHead` — three progs; insert at 1 → new prog is at index 0.
- `TestMpeditInsert_InMiddle` — three progs; insert at 2 → new prog is at index 1.
- `TestMpeditInsert_OutOfRange` — three progs; insert at 99 → "Program not found."; progs unchanged.
- `TestMpeditInsert_AtLastPosition` — three progs; insert at 3 (== `len(progs)`) → "Program not found.", progs unchanged. Pins the C `&& mprg->next` guard at `:9340`: the loop exits without splicing when the iteration reaches the final prog whose `->next == NULL`. A builder must use `add` (not `insert`) to append. (Adversary F4.)
- `TestMpeditInsert_Empty` — zero progs; insert at 1 → "No programs on mobile..."; progs unchanged.
- `TestMpeditInsert_OpensEditor` — same shape as G5.
- Parallel tests for opedit + rpedit. **Note:** rpedit's insert path is the C-bug branch (Open Q1) — if Open Q1 resolves to "fix", rpedit insert tests pass; if "preserve", rpedit insert tests assert "the gate doesn't fire" (i.e. inserting via rpedit does nothing). **Recommendation: fix.**

**Mutation:** flip splice-index from `value-1` to `value` → off-by-one test fails on `TestMpeditInsert_InMiddle`.

**Acceptance:** A12, A13.

### G7 — `edit` subcommand (incl. progtypes bitmask rebuild)

**Deliverables:**

- `mpeditEdit(ch, target, value int, mptype int64, argument string)` helper.
- Reject empty progs / `value < 1` (per C).
- Walk to `value`-th prog. If reached: optionally update `mprg.Type = mptype` (if `mptype != -1` semantically — Go uses `mptype != 0` as "no change" sentinel; alternative: pass a `*int64` or a `(mptype int64, hasType bool)` pair to disambiguate from "no override"). **Recommendation:** use `(mptype int64, override bool)` — cleaner than overloading `0`.
- Always update `mprg.ArgList = util.SmashTilde(argument)` (matching C `:9025`).
- Open editor via `mpeditOpenEditor(ch, mprg, target, true /* rebuild on save */)`.

**EditorSave closure:** copy buffer; `StopEditingFunc`; **rebuild progtypes bitmask** by `target.progTypes.Clear()` + iterate.

**Tests:**

- `TestMpeditEdit_KeepsType` — edit with no type override → type unchanged; arglist updated.
- `TestMpeditEdit_OverridesType` — edit with new type → type changed; arglist updated.
- `TestMpeditEdit_RebuildsProgtypes` — edit-changing-type closure invocation rebuilds bitmask: original [act, speech, greet] → edit greet's type to `act` → after `/s`, `progTypes` has act+speech bits only (greet bit cleared because greet has zero remaining progs).
- `TestMpeditEdit_OpensEditor` — same shape.
- Parallel tests for opedit + rpedit.

**Mutation:** drop `target.progTypes.Clear()` in the rebuild path → `TestMpeditEdit_RebuildsProgtypes` fails (greet bit lingers).

**Acceptance:** A14, A15.

### G8 — `delete` subcommand (incl. conditional bitmask clear)

**Deliverables:**

- `mpeditDelete(ch, target, value int)` helper.
- Reject empty progs / `value < 1` (per C).
- Walk to `value`-th prog → record its type as `removedType`.
- Count remaining progs sharing `removedType` (BEFORE deletion); call this `n`.
- Splice out the `value`-th prog from `*target.progs`.
- If `n <= 1` (i.e. the deleted prog was the last of its type): `target.progTypes.Remove(bits.TrailingZeros64(uint64(removedType)))`.
- Emit "Program removed.".

**Tests:**

- `TestMpeditDelete_Found` — delete the second of three progs; `len(progs) == 2`; "Program removed.".
- `TestMpeditDelete_OutOfRange` — `mpedit fido delete 99` → "Program not found."; progs unchanged.
- `TestMpeditDelete_Empty` — zero progs → "No programs on mobile..."; progs unchanged.
- `TestMpeditDelete_ClearsBitWhenLast` — start with one greet prog; delete it → `progTypes.IsSet(greet_bit) == false`.
- `TestMpeditDelete_KeepsBitWhenSiblings` — start with two greet progs; delete one → `progTypes.IsSet(greet_bit) == true`.
- Parallel tests for opedit + rpedit.

**Mutation:** drop the `if n <= 1` guard → `TestMpeditDelete_KeepsBitWhenSiblings` fails (bit erroneously cleared while a sibling prog of the same type still exists).

**Acceptance:** A16, A17.

### G9 — Substate re-entry safety + `SUB_MPROG_EDIT` semantics

**Deliverables:**

- Document at top of `olc_prog.go` that the C `switch (ch->substate)` re-entry path is replaced by the EditorSave closure. The `SUB_MPROG_EDIT` constant remains set during editing for any consumer that gates on it (notably `internal/game/editor.go:187-189` which extends `maxBufLines` for mprog/help edits — this is the load-bearing reason to set the substate even though the closure does the actual save).
- Verify in test that `ch.Substate` returns to `SUB_NONE` after `/s` (StopEditingFunc resets it per `editor.go:88`).

**Tests:**

- `TestMpeditEditor_SubstateLifecycle` — start state SUB_NONE; after `mpedit fido add act "..."`, `ch.Substate == SUB_MPROG_EDIT`; after invoking the captured EditorSave closure (which calls StopEditingFunc), `ch.Substate == SUB_NONE`.
- `TestMpeditEditor_HonorsMprogMaxLines` — confirm `editor.go:187` continues to honor mprog-extended buffer limit (this is a regression check, not a new behavior).

**Mutation:** in `mpeditOpenEditor`, drop `ch.Substate = SUB_MPROG_EDIT` → `TestMpeditEditor_SubstateLifecycle` first-half fails (substate stays SUB_NONE during edit).

**Acceptance:** A18, A19.

### G10 — Inspector test migration + integration tests

**Deliverables:**

- Rewrite `internal/act/olc_prog_test.go` test cases that asserted the inspector's `mpedit <vnum>` shape. Each becomes `mpedit <victim_name> list` or equivalent.
- Add an in-package E2E test that drives a full add → /s → list cycle: instantiate a mob in a test world, call `DoMpedit(ch, "fido add act \"hi\"")`, manually invoke the captured `EditorSave` closure with a fake CopyBufferFunc (per the established Tier-12 test pattern at `internal/act/olc_test.go:530+`), then call `DoMpedit(ch, "fido list 0 full")` and assert the new prog body appears.

**Tests:**

- `TestMpedit_FullCycleAddListEdit` — add → list → edit → list → assert mutations applied.
- `TestOpedit_FullCycleAddDelete` — add → delete → list-empty.
- `TestRpedit_FullCycleAddInsert` — add → insert at head → assert ordering.

**Mutation:** swap the order of `*target.progs = append(...)` and `progTypes.Set(...)` → no behavioral test failure; cosmetic. Skip — not a meaningful gate.

**Acceptance:** A20, A21, A22.

### G11 — Boot regs verified, no change required (sanity check)

**Deliverables:**

- Re-verify `internal/boot/boot.go:792-794` continues to register mpedit/opedit/rpedit at LEVEL_IMMORTAL with `POS_DEAD`. **Expected: zero changes.**
- Add a boot-test assertion if missing (mirror the redit/oedit/medit pattern).

**Tests:** `TestBoot_McEditorsRegistered` — interprets `commands` for an immortal char, asserts mpedit/opedit/rpedit appear.

**Acceptance:** A23.

### G12 — Documentation: CHANGELOG + CLAUDE.md LANDED row + TODO follow-ups

**Deliverables:**

- `CHANGELOG.md` entry for the lineage.
- `CLAUDE.md` Phase-6 landing-table row updated from `unauthored / authored / pending` to `LANDED YYYY-MM-DD (commit)`.
- `phases.md` §Phase 6 board: move row from "Unauthored" to "Landed".
- `phase6-roadmap.md`: mpedit row updated.
- `TODO.md`: add follow-up entries for any deferrals (e.g. `can_mmodify` / `can_omodify` / `can_rmodify` ACL helpers; potentially a non-prototype edit-as-instance path).

**Acceptance:** A24.

---

## Acceptance Criteria (binary)

| # | Criterion | Gate |
|---|---|---|
| A1 | `util.MProgFlagNames` matches C `mprog_flags[]` ordering for ported entries | `TestMProgFlagNames_HasAllCEntries` |
| A2 | `util.GetMpFlag` returns bit-flag value (not index); strips `_prog` suffix; case-insensitive | `TestGetMpFlag_*` suite |
| A3 | `DoMpedit` / `DoOpedit` / `DoRpedit` accept C-shape arguments (mob/obj take target name; room does not) | `TestMpeditDispatch_*` + `TestRpeditDispatch_NoVictimArg` |
| A4 | NPC reject + descriptor check + syntax-help paths emit C-faithful messages | `TestMpeditDispatch_NpcCannotEdit` etc. |
| A5 | Trust < LEVEL_GOD restricts mob/obj search to room/inventory; >= LEVEL_GOD searches world | `TestMpeditDispatch_VictimNotInRoom` + `_NotInWorld` |
| A6 | Prototype gate enforces ACT_PROTOTYPE / ITEM_PROTOTYPE; STATSHIELD trust gate fires | `TestMpeditDispatch_NotPrototype` + `_StatshieldGate` |
| A7 | `list` empty-list message per kind matches C verbatim (incl. C-bug "mob programs" wording on opedit/rpedit) | `TestMpeditList_Empty` per kind |
| A8 | `list` 1-based index by-position; `full` flag emits bodies | `TestMpeditList_ByIndex` + `_Full` |
| A9 | `add` appends to tail; sets progtypes bit synchronously | `TestMpeditAdd_AppendsAtTail` |
| A10 | `add` opens string editor via StartEditingFunc seam | `TestMpeditAdd_OpensEditor` |
| A11 | `add` editor-save closure writes ComList; SmashTilde applied to ArgList | `TestMpeditAdd_EditorSaveWritesComList` + `_SmashesTilde` |
| A12 | `insert` at head and middle splices correctly | `TestMpeditInsert_AtHead` + `_InMiddle` |
| A13 | `insert` opens editor; rpedit branch fixed to `arg1` (Open Q1 resolution) | `TestMpeditInsert_OpensEditor` + `TestRpeditInsert_GateFires` |
| A14 | `edit` updates ArgList; optionally updates Type; opens editor seeded with current ComList | `TestMpeditEdit_*` |
| A15 | `edit` editor-save closure rebuilds progtypes bitmask via clear+iterate | `TestMpeditEdit_RebuildsProgtypes` |
| A16 | `delete` removes prog at 1-based position; emits "Program removed." | `TestMpeditDelete_Found` |
| A17 | `delete` clears progtypes bit only when last prog of that type | `TestMpeditDelete_ClearsBitWhenLast` + `_KeepsBitWhenSiblings` |
| A18 | `Substate = SUB_MPROG_EDIT` during edit; resets to SUB_NONE after `/s` | `TestMpeditEditor_SubstateLifecycle` |
| A19 | Mprog-extended buffer limit (`editor.go:187`) honored during edit | `TestMpeditEditor_HonorsMprogMaxLines` |
| A20 | Full add → /s → list cycle round-trips body text correctly | `TestMpedit_FullCycleAddListEdit` |
| A21 | opedit full add → delete cycle leaves empty list | `TestOpedit_FullCycleAddDelete` |
| A22 | rpedit full add → insert cycle preserves ordering | `TestRpedit_FullCycleAddInsert` |
| A23 | Boot registry continues to expose mpedit/opedit/rpedit at LEVEL_IMMORTAL | `TestBoot_McEditorsRegistered` |
| A24 | CHANGELOG + CLAUDE.md + phases.md + TODO.md updated | manual verification |

---

## Mutation Gates (≥10, all `Edit`-round-trip)

| # | Mutation | Pinning test |
|---|---|---|
| M1 | Swap `MPROG_ACT` ↔ `MPROG_SPEECH` in lookup table | `TestGetMpFlag_KnownNames` |
| M2 | Drop `_prog` suffix-strip in `GetMpFlag` | `TestGetMpFlag_StripSuffix` |
| M3 | Replace `LEVEL_GOD` with `LEVEL_GREATER` in mpedit trust branch | `TestMpeditDispatch_VictimNotInRoom` (mob outside room with low-trust passes when it shouldn't) |
| M4 | Drop `IS_NPC(victim)` check (allow PC editing) | `TestMpeditDispatch_PcVictim` |
| M5 | Drop ACT_PROTOTYPE gate | `TestMpeditDispatch_NotPrototype` |
| M6 | Replace `append(*progs, mprg)` with prepend in `add` | `TestMpeditAdd_AppendsAtTail` |
| M7 | Replace `progTypes.Set(idx)` with `Toggle(idx)` in `add` | `TestMpeditAdd_DoubleSetIdempotent` (or via repeat-add) |
| M8 | Off-by-one in `insert` splice (`value` instead of `value-1`) | `TestMpeditInsert_InMiddle` |
| M9 | Drop `progTypes.Clear()` in `edit` rebuild path | `TestMpeditEdit_RebuildsProgtypes` |
| M10 | Drop `if n <= 1` guard in `delete` | `TestMpeditDelete_KeepsBitWhenSiblings` |
| M11 | Drop `ch.Substate = SUB_MPROG_EDIT` in `mpeditOpenEditor` | `TestMpeditEditor_SubstateLifecycle` |
| M12 | Drop `util.SmashTilde` on `arglist` in `add` | `TestMpeditAdd_SmashesTilde` |
| M13 | Drop `bits.TrailingZeros64` conversion (treat flag as bit-index directly) | `TestMpeditAdd_AppendsAtTail` (high-band flags like LOGIN end up flipping wrong bits) |
| M14 | rpedit insert gate left as `arg2` (Open Q1 "preserve verbatim") instead of `arg1` (Open Q1 "fix") | `TestRpeditInsert_GateFires` (only meaningful if Q1 fixes; otherwise this M is vacuous and replaced by `TestRpeditInsert_DeadBranchPreserved`) |

Each mutation: apply via `Edit old_string → new_string`, run the test (red), revert via `Edit new_string → old_string`, re-run (green). No `git checkout` / `git restore` / `git stash` / `git reset --hard` invocations.

---

## Scope Cuts / Deferrals

Out of scope for this plan:

- **`foldarea` / `mpxset` / mudprog-syntax validator.** Out per dispatch prompt. None are referenced by the trio.
- **`can_mmodify` / `can_omodify` / `can_rmodify` ACL helpers.** Not ported in Go today; the existing inspector relies on `GetTrust() >= LEVEL_IMMORTAL` (registry-enforced). Document as TODO follow-up.
- **Edit progs on a non-prototype instance.** C `do_*pedit` rejects non-prototypes; we preserve verbatim.
- **`do_mpstat` / `do_opstat` / `do_rpstat` (read-only stat dumps).** Already shipped in Phase 5 Tier 4. Out of scope.
- **Persistence of edited prototypes to `.are` files.** Mutations are in-memory only; the next `asave` tick (already shipped) flushes to disk via `persist/area_write.go`. Mpedit does NOT force-save.
- **Full mob-prog interpreter behavioral changes.** This plan ports the editor surface only; the runtime interpreter at `internal/mudprog/` is unaffected.
- **MEDIT-style menu loop.** Explicit non-goal: see §"No CON_MPEDIT" — the Tier-12 closure pattern is sufficient.
- **Argument-shape compatibility with the old inspector (`mpedit <vnum>`).** Replaced, not preserved. Tests rewritten.
- **Mob-prog body syntax checking.** `mp` interpreter's syntax errors surface at trigger-time, not edit-time. Out of scope (matches C).
- **`load_prog` / `greetinfight` / `move` / `emote` mob-prog flavors.** Listed in C `mprog_flags[]` but not yet ported to Go's `MPROG_*` enum at `internal/types/mudprog.go`. `getMpFlag` returns "unknown program type" for these names. Document as TODO follow-up.

---

## Open Questions

| # | Question | Recommendation |
|---|---|---|
| Q1 | C `do_rpedit` `insert` gate uses `arg2` (the value/number) instead of `arg1` (the subcommand) at `build.c:9991` — dead branch. Preserve verbatim, or fix in Go? | **Fix in Go.** The branch is clearly unintended; preserving it ships a feature gap with no fidelity benefit. Document the divergence inline + in CHANGELOG. |
| Q2 | C `do_rpedit` `insert` body uses `get_mpflag(arg2)` at `:9999` — same root cause. Fix to `arg3`? | **Fix together with Q1.** Both stem from the same arg-index slip. |
| Q3 | The "rpedit" body-helper at `src/build.c:9727-9742` is a verbatim duplicate of `mpedit` (`:9018-9033`). Single Go helper or two? | **Single Go helper** (`mpeditOpenEditor`) — the C duplication is accidental, no behavioral divergence. |
| Q4 | `getMpFlag` storage convention — return bit-flag VALUE (matching `MProgData.Type`) or bit-INDEX (matching C)? | **Return bit-flag VALUE.** Callers do `mprg.Type = bit` directly without conversion. Internal bit-set conversion via `bits.TrailingZeros64` happens at the BitVector boundary, not in the lookup helper. |
| Q5 | C `mpedit(ch, mprg, mptype, argument)` signature uses `mptype = -1` to mean "no type override" (`build.c:9020`). Go signature options: `(mptype int64, override bool)` two-arg vs `mptype int64` with `0` as sentinel. | **Two-arg `(mptype int64, override bool)`** — `0` is a valid bit-pattern for an unrecognized prog (G1 `getMpFlag` returns `0, false` for unknown), so overloading it as a sentinel inside `mpedit` would muddle the API. |
| Q6 | Mpedit accepts a victim name; the existing inspector accepts a vnum. Preserve both, or replace? | **Replace.** Vnum-form has no C analog and no production usage; preserving it doubles the surface for no benefit. Inspector tests rewritten. |
| Q7 | `add` walks the linked list to its tail in C (`for (; mprog->next; mprog = mprog->next);` at `:9362`). Go has a slice — `*progs = append(*progs, mprg)`. Is the slice-append guaranteed to preserve insertion order across reallocation? | **Yes** — Go slice append is order-preserving; reallocation copies elements verbatim. No semantic divergence. |
| Q8 | Persist of in-memory edits — should mpedit force a `SaveArea` for the affected area? | **No.** C does not force-save (mutations live in `pIndexData` until `asave`). Go follows. |
| Q9 | Should the editor opening clear ComList before seeding? | **No** — C seeds with the existing body (`start_editing(ch, mprg->comlist)` at `:9031`); the builder edits in place. Go follows verbatim. |
| Q10 | EditorSave closure on mpedit `edit` rebuilds the progtypes bitmask. Should `add` and `insert` also rebuild on save (defensively)? | **No** — C only rebuilds in `edit` (`:9234-9236`). Adding rebuild to `add`/`insert` would diverge from C with no correctness benefit (the bit was set synchronously). |

---

## Risk Analysis

- **Low overall.** Mechanical port of three near-identical dispatchers; no menu-state machine; no descriptor lifecycle changes; reuses Tier-12 callback infrastructure verbatim.
- **Single moderate risk: progtypes bitmask drift.** The `delete`-conditional and `edit`-rebuild paths both manipulate `ProgTypes`. A miscount in `delete` leaves a phantom bit set, breaking the runtime mudprog dispatcher's "does this prototype have any X-progs?" gate. Mitigation: M9 + M10 mutation gates pin both.
- **Low risk: argument shape divergence between mob/obj and room.** The fan-out to a per-kind argument parser is the only structural complexity. Mitigation: per-kind tests for syntax-help and target-resolution branches.
- **Low risk: EditorSave closure capture.** The closure binds `*MProgData` and `progEditorTarget` (which itself binds `*[]*MProgData` and `*BitVector`). All pointers stable across the editor session — no races (single-goroutine game loop). Same model as redit/oedit's tested closures.
- **Low risk: rpedit `insert` C-bug fix.** Q1 recommends fixing; if the fix is wrong, builders see a working `insert` instead of a dead one — failure mode is "feature works", not "regression". The corresponding test (`TestRpeditInsert_GateFires`) pins the fixed behavior.
- **Low risk: Unported mprog flavors.** `getMpFlag` returns false for `load` / `greetinfight` / `move` / `emote`. Builders typing those names see "Unknown program type." — same UX as C if they typo. Documented as TODO follow-up.

---

## Test Plan

| Test file | Group | Approx. LOC |
|---|---|---|
| `internal/util/mprog_flags_test.go` (NEW) | G1 | ~120 |
| `internal/act/olc_prog_test.go` (REWRITE) | G2-G10 | ~600 (replaces existing 148) |
| `internal/boot/boot_test.go` (extend) | G11 | +10 |

Estimated total: ~720 new test LOC; ~30 tests; ~14 mutation gates.

**Integration tests:** in-package E2E covers the closure round-trip (G10). No `internal/testclient/` scenarios are required because mpedit does not introduce a new descriptor state — the closure invocation is already covered by Tier-12's testclient layering. Optionally add one testclient scenario per Wave-D consistency:

- `TestTestclient_MpeditAddCycle` (new) — login as immortal, instantiate a prototype mob in the test world via existing fixture loader, type `mpedit fido add act "say hi"`, type editor body `say hello world\n`, type `/s`, type `mpedit fido list 0 full`, assert "say hello world" appears.

This is OPTIONAL (test-plan §3 above is sufficient); if added, increment LOC by ~80 and add as G10b.

---

## File Budget

New files:

| Path | Approx. LOC |
|---|---|
| `internal/util/mprog_flags.go` | ~80 |
| `internal/util/mprog_flags_test.go` | ~120 |

Modified files:

| Path | Approx. Δ LOC |
|---|---|
| `internal/act/olc_prog.go` | ~+450 (replaces inspector at 162 LOC; new dispatcher + 5 subcommand helpers + EditorSave plumbing) |
| `internal/act/olc_prog_test.go` | ~+450 (148 → ~600; restructure around new arg shape + add closure-invocation tests) |
| `internal/boot/boot_test.go` | ~+10 |
| `CLAUDE.md` | ~+1 line |
| `CHANGELOG.md` | ~+15 |
| `TODO.md` | ~+10 |
| `smaug-go/doc/phases.md` | ~+1 row move |
| `smaug-go/doc/phase6-roadmap.md` | ~+1 row update |

**Total:** ~720 new LOC (code + tests); ~470 modified LOC.

---

## Cross-References

- `smaug-go/doc/plan-phase6-olc-redit.md` — first OLC plan; established `OlcData`-on-Descriptor + nanny dispatch (NOT used by mpedit).
- `smaug-go/doc/plan-phase6-olc-oedit.md` — second OLC plan; `olcBitmaskEdit` helper (NOT used by mpedit; mpedit has no bitmask-toggle UI).
- `smaug-go/doc/plan-phase6-olc-medit.md` — third OLC plan; reservation of `OEDIT_MPROGS_*` iota slots that this plan **does NOT consume** (mpedit has no `CON_*EDIT` substate; the reserved slots remain unused after mpedit lands and could be reclaimed in a follow-up cleanup).
- `smaug-go/doc/plan-phase5-tier12-editor-save.md` — `EditorSave` callback contract (the load-bearing pattern this plan reuses).
- `src/build.c:9018-9033` — `mpedit` body-helper.
- `src/build.c:9039-9376` — `do_mpedit`.
- `src/build.c:9379-9719` — `do_opedit`.
- `src/build.c:9744-10058` — `do_rpedit`.
- `src/build.c:366-374` — `mprog_flags[]` table.
- `src/build.c:721-728` — `get_mpflag`.
- `src/mud.h` — `SUB_MPROG_EDIT` constant; `MPROG_DATA` struct (already ported as `MProgData`).
- `internal/types/mudprog.go:9-15` — `MProgData` struct.
- `internal/types/mudprog.go:49-101` — `MPROG_*` bit-flag constants.
- `internal/persist/area.go:505-514` — area-loader convention for `ProgTypes` bit-index.
- `internal/act/olc.go:99-115` — `redit desc` EditorSave closure precedent.

---

## Adversary Verification Notes

**Audit date:** 2026-04-26.
**Auditor:** Adversary agent (lineage: phase6-olc-mpedit-audit).
**Verdict:** PASS-with-CONCERNS

---

### Mechanical Checks

`adversary-check.sh` ran clean: no large additions, no missing test files, no conflict markers. The only changed file is the plan document itself (the adversary section). No test runner applicable (plan-only repo state).

---

### Claim Verification

All major structural claims verified against filesystem and C source. One factual error found (mprog_flags count — see MEDIUM finding F1). All other claims match.

---

### C Citations Verified

**Count: 18 distinct spans read.**

Verified (line numbers and content match):

- `src/build.c:9018-9033` — `mpedit` body-helper. Exact as quoted. `mptype != -1` guard, `STRALLOC("")` for empty comlist, `start_editing` call. Confirmed no behavioral divergence from the `rpedit` duplicate.
- `src/build.c:9039-9081` — `do_mpedit` substate block. The `SUB_MPROG_EDIT` branch at lines 9067-9080 copies buffer, calls `stop_editing`, and returns at line 9080. Line 9081 is `}` (closes the switch). There is **no unreachable code after the return** — the plan's note "NB: lines 9080-9081 contain the `return;` — there is NO unreachable code here despite earlier reports" is **confirmed correct**.
- `src/build.c:9083-9102` — `smash_tilde`, three `one_argument` calls, `value = atoi(arg3)`, syntax-help branch. Confirmed.
- `src/build.c:9104-9119` — Victim resolution (`LEVEL_GOD` bifurcation). Confirmed.
- `src/build.c:9121-9135` — Trust/NPC/STATSHIELD/ACL gates. Confirmed.
- `src/build.c:9144` — `mprog = victim->pIndexData->mudprogs`. Confirmed.
- `src/build.c:9148-9201` — `list` subcommand. `value < 1` + `strcmp("full", arg3)` logic; note: `strcmp` returns non-zero when strings are NOT equal, so header-only is the `if (strcmp("full", arg3))` branch (strings differ = header-only), full body in the `else`. Plan does not highlight this double-negative; worker must get it right.
- `src/build.c:9203-9242` — `edit` subcommand. `arg4` parsed for optional new type; `xCLEAR_BITS` then iterate for progtypes rebuild. Confirmed.
- `src/build.c:9244-9306` — `delete` subcommand. Counts `num` progs sharing `mptype` BEFORE deletion, unlinks, then `if (num <= 1) xREMOVE_BIT`. Confirmed.
- `src/build.c:9308-9352` — `insert` subcommand (`do_mpedit`). `arg4` is the prog type; `value == 1` splices at head; `value > 1` walks to position. Confirmed.
- `src/build.c:9354-9373` — `add` subcommand. `arg3` is the prog type. `mprg->next = NULL` at line 9371 is reachable. Confirmed.
- `src/build.c:9379-9421` — `do_opedit` opening. Bug message at line 9411: `"do_opedit: sub_oprog_edit: NULL ch->dest_buf"` — confirmed. The substate is `SUB_MPROG_EDIT`; the message string is the copy-paste error the plan documents.
- `src/build.c:9446-9473` — `do_opedit` target resolution (obj-carry / obj-world). Confirmed.
- `src/build.c:9697-9718` — `do_opedit` `add` subcommand. Note the `if (mprog) mprog->next = mprg; else obj->pIndexData->mudprogs = mprg` at lines 9708-9711 is reachable (both branches). Plan correctly notes these are NOT unreachable. Confirmed.
- `src/build.c:9727-9742` — `rpedit` body-helper. **Byte-for-byte duplicate of `mpedit` at 9018-9033**, modulo function name. No behavioral divergence. Confirmed — plan's G3 consolidation is correct.
- `src/build.c:9745-9812` — `do_rpedit` opening. Only two `one_argument` calls (arg3 commented out at line 9791). `value = atoi(arg2)`. Confirmed arg-shape divergence.
- `src/build.c:9775` — `do_rpedit` substate bug message: `"do_opedit: sub_oprog_edit: NULL ch->dest_buf"` — confirmed copy-paste typo from `do_opedit`. Confirmed.
- `src/build.c:9991-10033` — `do_rpedit` `insert` branch. Gate is `!str_cmp(arg2, "insert")` at line 9991. At this point `arg2` contains the value/number string (already consumed as `value = atoi(arg2)` at line 9790), NOT the subcommand. The subcommand is in `arg1`. This branch is dead in all normal usage: for `insert` to fire here, the user would need to type something like `rpedit <n> insert ...` where `arg1` is a number matching none of the prior string comparisons. Confirmed dead branch. Also confirmed: line 9999 uses `get_mpflag(arg2)` inside the insert body even though `arg2` is the value string — the wrong arg slot. Plan documents both bugs accurately.
- `src/build.c:366-374` — `mprog_flags[]` table. **52 entries (indices 0-51)**, not 51 as stated in the plan. See F1 below.
- `src/build.c:721-728` — `get_mpflag`. Returns the **array index** (0..51), not a bit-flag value. C callers then pass this index to `xSET_BIT(progtypes, mptype)`. Plan documents that Go's `getMpFlag` should instead return the **bit-flag value** (`1 << index` for low bits, mapped constants for high bits). This design choice (Q4) is correctly documented.

---

### Go Citations Verified

**Count: 12 distinct spans read.**

Verified:

- `internal/act/olc_prog.go:96-108` — `DoMpedit` / `DoOpedit` / `DoRpedit` are thin wrappers calling `mudprogEdit(ch, argument, "mob"|"obj"|"room")`. Inspector only, no mutation surface. Confirmed.
- `internal/act/olc_prog.go:110-162` — `mudprogEdit` dispatcher: vnum-based, no name resolution. Confirmed — this is what the plan replaces.
- `internal/act/olc_prog.go:15-57` — `triggerBitFromName` map. Present and correct. Plan's G1 consolidates this into `util.MProgFlagNames`. Confirmed.
- `internal/act/olc_prog.go:59-69` — `parseProgTriggerName`. Strips `_prog` and `prog` suffixes. Confirmed.
- `internal/act/olc_prog.go:201-252` — `firstTriggerName`. Reverse lookup. Confirmed.
- `internal/game/editor.go:159-173` — `/s` handler. Transitions `Connected = CON_PLAYING` BEFORE invoking `EditorSave` closure, clears `EditorSave` to nil before calling, invokes the closure. Confirmed — the plan's "EditorSave invoked after CON_PLAYING transition" claim is correct.
- `internal/game/editor.go:187-188` — `maxBufLines` override for `SUB_MPROG_EDIT`. Sets `maxLines = maxBufLinesMp` when substate is `SUB_MPROG_EDIT` or `SUB_HELP_EDIT`. This is the load-bearing reason to set `ch.Substate = SUB_MPROG_EDIT` in the Go port even though the closure handles the save. Plan documents this correctly. Confirmed.
- `internal/types/character.go:60` — `EditorSave func(*CharData)`. Present. Confirmed.
- `internal/types/enums.go:112` — `SUB_MPROG_EDIT`. Present. Confirmed.
- `internal/act/olc.go:18-27` — `StartEditingFunc`, `CopyBufferFunc`, `StopEditingFunc`. Present. Confirmed.
- `internal/boot/boot.go:792-794` — mpedit/opedit/rpedit registered at `LEVEL_IMMORTAL`, `POS_DEAD`. Confirmed.
- `internal/persist/area.go:505-514, 750-758, 895-903` — All three spans (mob, obj, room) follow the convention: `prog.Type` is the bit-flag value; `ProgTypes.Set(bits.TrailingZeros64(uint64(prog.Type)))` converts to bit-position. The plan's progtypes convention description is **confirmed correct**.
- `internal/types/olc.go:123-127` — `OEDIT_MPROGS`, `OEDIT_MPROGS_CHOICE`, `OEDIT_MPROGS_DELETE`, `OEDIT_MPROGS_TYPE`, `OEDIT_MPROGS_ARG`. Five iota slots reserved at medit Wave 1. No other plan in the doc set references them as inputs. Plan's statement "go unused" after mpedit lands is correct — they were reserved optimistically for a CON_OEDIT menu-driven mprogs editor that was never built. The plan correctly flags them for follow-up cleanup. Confirmed.
- `internal/handler/find.go:14, 52, 91, 183` — `GetCharRoom`, `GetCharWorld`, `GetObjCarry`, `GetObjWorld`. All present. **Important:** `GetCharWorld` signature is `func GetCharWorld(w *world.World, ch *types.CharData, argument string)` and `GetObjWorld` is `func GetObjWorld(w *world.World, ch *types.CharData, argument string)` — both require a `*world.World` as first argument. The plan cites these functions but the dispatch skeleton (§Dispatch shape) does not show this parameter explicitly. The `act` package has `WorldRef *world.World` (boot-set), so the worker will pass `WorldRef` — this is the established pattern (confirmed via `internal/act/wiz.go:48` etc.). The plan does mention `WorldRef *world.World` in the Go Current State table but the sketch omits it. Minor omission; not a correctness risk given precedent.
- `internal/act/olc.go:99-115` — `redit desc` EditorSave closure. Confirmed as the precedent shape the plan follows.

---

### Findings

#### F1 — MEDIUM: `mprog_flags[]` entry count is 52, not 51

The plan states "51 entries" in two places: the C Reference section (line ~134: "51 entries") and G1 deliverables (line ~371: "51 entries"). Counting the actual C table at `src/build.c:366-374`:

```
act speech rand fight death hitprcnt entry greet allgreet give bribe hour time
wear remove sac look exa zap get drop damage repair randiw speechiw pull push
sleep rest leave script use load login void tell imminfo greetinfight move
command sell emote r1 r2 r3 r4 r5 r6 r7 r8 r9 r10
```

That is 52 entries (indices 0-51: `act`=0 through `r10`=51). The plan's "51 entries" is a counting error — off by one because it likely missed either `r10` or the reserved slots. G1's `TestMProgFlagNames_HasAllCEntries` should assert 52 entries, not 51. If the worker builds a 51-entry table they will map `r10` to the wrong index or drop it, silently misaligning all reserved slots starting at index 42.

**Fix:** Correct "51 entries" to "52 entries" in both the C Reference table header and the G1 deliverable description. Update the G1 test assertion to `len(MProgFlagNames) == 52`.

#### F2 — LOW: `list` subcommand `full` flag: plan inverts C's `strcmp` semantics

The plan's G4 description says: "`value < 1`: full list (header-only by default, body included when `arg3 == 'full'`)."

The C code at `src/build.c:9160` is `if (strcmp("full", arg3))` — which is TRUE when the strings are NOT equal (i.e., arg3 is NOT "full"), meaning header-only. The body-print path is in the `else` branch (strings equal). The plan's English description is correct in outcome ("body included when arg3 == full") but a worker reading the plan who implements it as `if arg3 == "full" { full list } else { header-only }` gets the right result. No bug here — but the plan does NOT mention the double-negative in C (`strcmp` returns non-zero for inequality), which means a worker porting the C literally might write `if !str_cmp(arg3, "full")` and get the wrong branch. Recommend adding a note: "C uses `strcmp("full", arg3)` (non-zero = not full = header-only); Go should use `arg3 == "full"` for clarity — this IS a deliberate Go-readability divergence."

#### F3 — LOW: `GetCharWorld` / `GetObjWorld` require `WorldRef` as first arg — sketch omits it

The dispatch sketch in §Dispatch shape shows `mpedit_dispatch(ch, argument, kind)` without a `w *world.World` parameter. The actual Go functions `handler.GetCharWorld` and `handler.GetObjWorld` both require `(w *world.World, ch, arg)`. The `act` package resolves this via the package-level `WorldRef` (the established pattern, confirmed in six other act commands). The plan mentions `WorldRef` in the Go Current State table but the sketch omits it.

This is a documentation gap, not a correctness risk: the worker will follow precedent. But if a new worker unfamiliar with the package reads only the sketch, they may write a version that compiles but passes `nil` or the wrong world. Recommend adding `WorldRef` explicitly to the dispatch sketch or adding a comment: "Use `WorldRef` for GetCharWorld/GetObjWorld — same pattern as wiz.go."

#### F4 — LOW: `insert` splice semantics — `mprg->next` guard vs. plan's description

C `do_mpedit` insert at line 9340: `if (++cnt == value && mprg->next)` — the `&& mprg->next` guard means the insert fails silently if the target position is the last element (no `->next`). The plan's G6 says "walk to position `value-1`, splice after" but does not document this last-position edge case. A builder typing `mpedit fido insert 3` on a 3-prog list (inserting after the last prog) would hit this guard and get "Program not found." instead of appending. The plan tests `TestMpeditInsert_OutOfRange` with position 99 but not the "insert at last element" case. This is a fidelity edge case worth pinning with a test.

**Fix:** Add a test `TestMpeditInsert_AtLastPosition` that asserts inserting at position `len(progs)` on a non-empty list yields "Program not found." (matching C). Or, if the Go port intentionally diverges (allowing append-at-end), document it explicitly.

#### F5 — LOW: Open Q10 is listed as Q10 but the plan footer says "Q1-Q9"

The plan's footer line (near line 790-791) reads: "**Open questions:** Q1-Q9." but the §Open Questions table has 10 entries (Q1-Q10). The footer is stale — Q10 was added without updating the count.

**Fix:** Update the footer to "Q1-Q10".

---

### Decision Audit

**Decision 1 — No CON_MPEDIT: BLESSED**

Verified independently. C `do_mpedit`'s `SUB_MPROG_EDIT` branch at lines 9063-9081 is the ONLY re-entry path, and it copies the buffer and returns immediately — no menu loop, no CON_*EDIT state machine. The `return;` at line 9080 is not followed by unreachable code. The Go EditorSave closure is a correct and cleaner replacement. The plan's claim is accurate.

The concern the dispatch prompt raised — "does the dispatcher need state to remember which prog the editor is currently mutating?" — is resolved by the closure capture. The closure binds `*MProgData` directly (equivalent to C's `ch->dest_buf`), so no external state is needed. This is sound.

**Decision 2 — Argument-shape change: BLESSED**

Verified: C `do_opedit` at lines 9446-9461 takes an obj NAME (not vnum), resolved via `get_obj_carry` / `get_obj_world`. C `do_rpedit` at lines 9787-9812 takes NO victim arg (operates on `ch->in_room` directly). The plan documents both divergences correctly.

The inspector's `mpedit <vnum>` replacement is an acceptable breakage: the inspector is a Go-only interim that has no C analog and is not user-visible in production. Tests rewritten, not deleted.

**Decision 3 — rpedit insert C-bug fix: BLESSED with note**

Bug confirmed: `src/build.c:9991` gates on `arg2` (the value/number string, not the subcommand). C `do_rpedit` has `arg1=subcommand`, `arg2=number`, and the insert check `!str_cmp(arg2, "insert")` is dead. The plan's recommendation to fix in Go (Q1/Q2 resolved to fix) is reasonable and the correct call. The "fix in Go" policy is consistent with the project's established precedent for dead-code fixes (per plan.md's testing strategy, which explicitly approves diverging from C when C behavior is unintended).

Note: the plan does not cite an explicit policy document for this class of fix. The dispatch prompt flagged "confirm that policy actually allows this." The author deferred to recommendation without citing the policy file. This is LOW risk — the fix is clearly correct — but the worker should add a `// C-bug fix: was arg2 in C, dead branch` comment as the plan instructs.

**Decision 4 — Single `mpeditOpenEditor` helper: BLESSED**

Read both C helpers (`mpedit` at 9018-9033 and `rpedit` at 9727-9742) side by side. They are byte-for-byte identical modulo function name. No behavioral divergence. Consolidation is correct.

**Decision 5 — C-bug catalog: BLESSED (with one additional note)**

All three documented bugs verified:
- `build.c:9411` — "do_opedit: sub_oprog_edit:" in `do_opedit`. Confirmed.
- `build.c:9775` — Same string in `do_rpedit`. Confirmed.
- `build.c:9991` — insert gate on `arg2`. Confirmed dead branch.

Additional note: the plan's C-bug catalog entry #4 says "line 9371 is reachable but redundant" (`mprg->next = NULL`). Confirmed reachable — `start_editing` starts the editor but does not block; `mpedit` returns immediately after `start_editing` returns, and line 9371 executes. In Go this is a no-op (zero-value struct). No issue.

---

### Scope Check

No scope creep detected. The plan is tightly bounded to the three dispatchers plus necessary helpers. The optional G11 boot-reg sanity check and G12 docs are appropriate bookends.

The `OEDIT_MPROGS_*` iota slots are correctly identified as dead-letter after mpedit lands; the plan does not claim to clean them up (deferred to TODO). Appropriate.

---

### Alternative Approach

The plan chooses a single shared `mpedit_dispatch` parameterized by `kind string` with internal branching for mob/obj/room. An alternative would be three fully separate functions (`DoMpedit`, `DoOpedit`, `DoRpedit` each implemented from scratch, sharing only `mpeditOpenEditor`). The chosen approach is the correct call here: the three editors are near-identical (~95% code overlap), and the kind-parameter fan-out is only two branches (mob-or-obj vs room for argument shape). The shared dispatcher reduces the surface for divergence bugs. The tradeoff is a `kind == "room"` branch in argument parsing — low complexity cost.

---

### Assumptions

1. **`WorldRef` is non-nil by the time mpedit is invoked.** True at runtime (boot-wired); test scaffolding must set `WorldRef` or the dispatch will nil-panic on world lookup. The inspector tests at `olc_prog_test.go` do set this up (via `lookupProgs` which reads `WorldRef`). Workers extending the test file must maintain this.

2. **`ch.Desc` non-nil check is sufficient for "has a connected player."** The plan gates on `ch.Desc == nil` per C. Correct — headless mobs have nil Desc.

3. **The `progEditorTarget` struct's `progs *[]*types.MProgData` pointer-to-slice approach correctly propagates mutations.** This is sound: `append` on a `*[]T` writes through the pointer. Confirmed by the `add` semantic. Workers must not pass the slice by value.

4. **`bits.TrailingZeros64(uint64(prog.Type))` is safe for all `MPROG_*` values.** True for all currently defined values (max bit index is 37 for `MPROG_CMD = 1<<37`). The `uint64` cast is safe because `int64` values with bit 63 set would represent negative numbers — but no current MPROG constant is that high. The plan's convention is sound.

5. **`util.SmashTilde` is applied to `argument` (the arglist) before the editor opens.** The plan applies it to the `ArgList` field. C does `smash_tilde(argument)` at the top of the function, which also affects the rest of the argument string being parsed. In Go, the plan applies `util.SmashTilde` only to the arglist portion passed to `mpeditOpenEditor`. This means prog-type names containing tildes would survive in C too (they can't — they're keyword lookups), so this is fine. Assumption holds.

---

### Security

No command injection or path traversal vectors. The dispatchers operate on in-memory data structures only; persistence is via the existing `asave` path. No new file I/O, no shell invocations, no trust elevation. No secrets.

The trust-gate logic (LEVEL_GOD for world search, LEVEL_GREATER for STATSHIELD, LEVEL_IMMORTAL at registry) is correctly layered and matches C.

---

### Mutation Gate Spot-Check

Four gates evaluated:

**M6 — Replace `append(*progs, mprg)` with prepend:**
`TestMpeditAdd_AppendsAtTail` starts with [act, speech] and asserts new prog is at index 2. A prepend would put it at index 0. Gate would catch the mutation. **Gate is sound.**

**M9 — Drop `target.progTypes.Clear()` in edit rebuild path:**
`TestMpeditEdit_RebuildsProgtypes` edits a greet-type prog to act-type and asserts greet bit is cleared. Without `Clear()`, the old bits persist. Gate would catch it. **Gate is sound.**

**M10 — Drop `if n <= 1` guard in delete:**
`TestMpeditDelete_KeepsBitWhenSiblings` starts with two greet progs, deletes one, asserts greet bit stays set. Without the guard, the bit would be cleared. Gate would catch it. **Gate is sound.**

**M13 — Drop `bits.TrailingZeros64` (treat flag as bit-index):**
Test asserts high-band flags (e.g., `MPROG_LOGIN = 1<<32`) end up with the correct bit set in `ProgTypes`. Without the conversion, `ProgTypes.Set(1<<32)` would try to set bit index `4294967296` — that would either panic or silently overflow. Gate would catch it. **Gate is sound.**

**M14 note:** If Q1 resolves to "fix" (plan's recommendation), `TestRpeditInsert_GateFires` becomes a meaningful pin. If Q1 resolves to "preserve", M14 is vacuous. The plan acknowledges this. **Gate is conditionally sound — acceptable.**

---

### Open Questions Assessment

Q1/Q2 (rpedit insert bug): Plan recommends fix. Verified the bug. Fix is the right call. No human input needed before dispatch.

Q3/Q4/Q5/Q7/Q8/Q9/Q10: All are design choices the plan resolves authoritatively and correctly. No human input needed.

Q6 (vnum form replacement): Replacement is correct. No human input needed.

**No open questions require human escalation before G1 dispatch.**

---

### Pre-Dispatch Edits Recommended

1. **Correct "51 entries" to "52 entries"** in two places (C Reference table header and G1 deliverable). Update `TestMProgFlagNames_HasAllCEntries` assertion to `len == 52`. (F1 — required before G1.)

2. **Add `TestMpeditInsert_AtLastPosition`** to G6's test list: inserting at `len(progs)` on a non-empty list yields "Program not found." (F4 — recommended before G6.)

3. **Add note in §Dispatch shape sketch** that `GetCharWorld`/`GetObjWorld` take `WorldRef` as first arg. One sentence. (F3 — cosmetic but useful for future readers.)

4. **Update plan footer** from "Q1-Q9" to "Q1-Q10". (F5 — trivial.)

5. **Add a clarity note in G4** that C's `strcmp("full", arg3)` is non-zero-means-not-equal; the Go port should use `arg3 == "full"` (not literal porting of `strcmp`). (F2 — recommended.)

---

**VERDICT: PASS-with-CONCERNS**

The plan is structurally sound, factually accurate on all major claims, and ready for G1 dispatch after the two pre-dispatch edits above: (1) correct the mprog_flags count from 51 to 52 — a factual error that would produce a wrong-length table in G1 and misalign reserved slots — and (2) add the insert-at-last-position test case to G6. All five locked design decisions are affirmed. No human escalation required.

---

**Plan authored:** 2026-04-26.
**Task groups:** G1-G12.
**Acceptance criteria:** A1-A24.
**Mutation gates:** M1-M14.
**Open questions:** Q1-Q10.
**C citations verified:** `mpedit` body @ :9018-9033; `do_mpedit` @ :9039-9376; `do_opedit` @ :9379-9719; `do_rpedit` @ :9744-10058; `mprog_flags[]` @ :366-374; `get_mpflag` @ :721-728.
**Pending:** adversary review before G1 dispatch.
