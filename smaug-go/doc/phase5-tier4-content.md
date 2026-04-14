# Phase 5 — Tier 4: Content Breadth

## Goal

With Tier 1's foundations (`Act()`, `spell_smaug`, saves), Tier 2's flag enforcement, and Tier 3's mudprog depth in place, Tier 4 is the content fill: the long tail of spells, combat skills, damage messages, missing player/immortal commands, and OLC interactivity that brings a Go builder and player experience up to C parity. After Tier 4:

1. The spellbook goes from 30 → ~100+ spells by combining `spell_smaug` metadata fills with a smaller set of custom Go ports (breath attacks, polymorph, possess, animate dead, gate, pass door, etc.).
2. Combat/utility skills roughly double (15 → 43+): unarmed attacks (bite/claw/punch/sting/tail), advanced combat (circle, gouge, stun, grapple, cleave, hitall, berserk), utility (meditate, scribe, cook, search, detrap, dig, feed).
3. Damage-message lookup dispatcher renders weapon-specific hit text ("Your claw cleaves Fido!") instead of generic output.
4. 14 missing immortal commands and 9 missing mortal commands are registered.
5. OLC gains the subcommand depth and interactive editors C builders expect.

## Gap inventory (verified)

### Spells — summary

Existing registry: 30 entries in `magic/magic.go:18–49`. C has ~101 distinct `spell_*` functions in `src/magic.c` (ignoring `spell_notfound`, `spell_null`, `spell_smaug` itself).

**Data-driven (handled by `spell_smaug` once Tier 1 ships — no Go code needed, only `skills.dat` metadata):** the majority of simple affect/attack spells — armor, bless, curse, poison (already ported but eligible), blindness, colour spray, chill touch, weaken, faerie fire, faerie fog, shocking grasp, flamestrike, harm, cause light/serious/critical, cure blindness, detect poison, dispel evil, remove curse/trap/invis, change sex, know alignment, pass door, etc. Estimated ~55 spells.

**Require custom Go ports (unique mechanics):**

| Spell | C ref | Mechanic |
|-------|-------|----------|
| spell_acid_blast | magic.c:2321 | Damage + corrosion of worn armor |
| spell_acid_breath | magic.c:5290 | Breath attack + random equipment decay |
| spell_animate_dead | magic.c:6489 | Creates undead NPC from corpse |
| spell_astral_walk | magic.c:4911 | Teleport to named target if reachable |
| spell_burning_hands | magic.c:2379 | Cone damage |
| spell_call_lightning | magic.c:2405 | Outdoor-weather-gated damage |
| spell_charm_person | *already ported* | — |
| spell_control_weather | magic.c:2692 | Alter `WeatherInfo.Sky` / pressure |
| spell_create_mob | magic.c:7806 | Create NPC of vnum |
| spell_create_obj | magic.c:7747 | Create obj of vnum |
| spell_dream | magic.c:6767 | Message to sleeping target |
| spell_earthquake | magic.c:3276 | Area damage; blocked by AFF_FLYING |
| spell_energy_drain | magic.c:3469 | Damage + XP drain |
| spell_expurgation | magic.c:2856 | Strip multiple evil affects |
| spell_farsight | magic.c:5973 | See into remote rooms |
| spell_fire_breath | magic.c:5358 | Breath + burn carried items |
| spell_frost_breath | magic.c:5429 | Breath + freeze liquid containers |
| spell_gas_breath | magic.c:5483 | Breath + poison |
| spell_lightning_breath | magic.c:5529 | Breath + random equipment shock |
| spell_gate | magic.c:3616 | Open portal to target |
| spell_group_teleport | magic.c:5066 | Teleport group |
| spell_knock | magic.c:6739 | Unlock/open door |
| spell_mist_walk | magic.c:6237 | Short-range teleport |
| spell_pass_door | magic.c:4522 | AFF_PASS_DOOR affect |
| spell_plant_pass | magic.c:6167 | Plane-specific travel |
| spell_polymorph | magic.c:3256 | Swap morph data; stat mods |
| spell_portal | (various) | Create portal object |
| spell_possess | magic.c:6656 | Switch into NPC body |
| spell_recharge | magic.c:6094 | Restore wand/staff charges |
| spell_remove_invis | magic.c:6387 | Strip AFF_INVIS from target |
| spell_remove_trap | magic.c:4648 | Defuse trap |
| spell_revive | magic.c:8119 | Resurrect from death |
| spell_sacral_divinity | magic.c:2824 | Level-3 holy attack |
| spell_solar_flight | magic.c:6318 | Outdoor daylight-gated travel |
| spell_transport | magic.c:5574 | Targeted teleport |
| spell_ventriloquate | magic.c:5183 | Fake speech source |
| spell_word_of_recall | magic.c:5279 | Recall to ROOM_VNUM_TEMPLE |
| "chaos tree" (8) | magic.c:7882–8280 | Class-specific high-level: black_fist, black_hand, black_lightning, caustic_fount, disruption, ethereal_fist, galvanic_whip, grasp_suspiria, hand_of_chaos, helical_flow, magnetic_thrust, midas_touch, mind_wrack, mind_wrench, quantum_spike, sonic_resonance, spectral_furor, spiral_blast, sulfurous_spray |

Estimated custom ports: ~25–30 spells.

### Combat / utility skills

Go has 15 skills in `act/skills.go` + `act/skills2.go`. Missing (C functions live in `src/skills.c`, some in `src/fight.c`):

| Skill | C ref | Category |
|-------|-------|----------|
| bite | skills.c:3220 | Unarmed attack |
| claw | skills.c:3260 | Unarmed attack |
| punch | skills.c:3180 | Unarmed attack |
| sting | skills.c:3295 | Unarmed attack |
| tail | skills.c:3335 | Unarmed attack |
| circle | skills.c:5136 | Combat — reposition behind target |
| gouge | skills.c:1810 | Blinding strike |
| stun | skills.c:3425 | Stun target |
| grapple | skills.c:1690 | Grapple + immobilize |
| cleave | skills.c:3831 | Heavy weapon double-hit |
| hitall | skills.c:5291 | Hit everyone in room |
| berserk | skills.c:5247 | Self-buff rage state |
| bloodlet | skills.c:3517 | DoT |
| pounce | skills.c:2607 | Charge attack |
| broach | skills.c:3960 | Armor-piercing |
| trance | skills.c:3028 | Meditative state |
| meditate | skills.c:2917 | HP/mana regen state |
| poison_weapon | skills.c:4681 | Apply poison to wielded weapon |
| throw | skills.c:6355 | Throw item |
| fire | skills.c:6223 | Ranged (archery primer) |
| detrap | skills.c:1906 | Disarm traps |
| dig | skills.c:2030 | Excavate |
| search | skills.c:2223 | Find hidden |
| skin | skills.c:523 | Skin corpse |
| scribe | skills.c:4836 | Write scroll |
| cook | skills.c:6751 | Cook food |
| stance | (special.c) | Combat stance toggle (Phase 6 full system; MVP is the skill stub) |
| style | skills.c:6525 | Style toggle |
| mistwalk | skills.c:3887 | Short teleport |
| feed | (various) | Feed food |
| slookup | skills.c:615 | Immortal skill inspect |
| sset | skills.c:862 | Immortal skill set |
| visible | skills.c:4235 | Cancel invis/hide |

Total ≈ 28 skills.

### Damage-message dispatcher

C `new_dam_message` at `src/fight.c:4410` selects verb/intensity strings from `s_message_table` / `p_message_table` indexed by weapon type × damage percentage. Go has `combat.Damage` but no equivalent renderer — attacks print generic output or none.

### Missing immortal commands (14 confirmed)

| Command | C ref | Purpose |
|---------|-------|---------|
| switch | act_wiz.c:3760 | Possess an NPC |
| return | act_wiz.c:3830 | Exit possession |
| wizlock | act_wiz.c:6075 | Block new logins at immortal level |
| shutdown | act_wiz.c:3606 | Shut down the MUD |
| reboot | act_wiz.c:3554 | Reboot the MUD |
| wizhelp | act_wiz.c:610 | List immortal commands |
| aecho | act_wiz.c:1342 | Area-wide echo |
| hell | act_wiz.c:8498 | Put player in "hell" room |
| log | act_wiz.c:5677 | Toggle command logging for player |
| mpstat | mud_comm.c:148 | Inspect mob progs |
| opstat | mud_comm.c:222 | Inspect obj progs |
| rpstat | mud_comm.c:265 | Inspect room progs |
| deny | act_wiz.c:1083 | Deny player entry |
| pardon | act_wiz.c:1241 | Remove deny |
| disconnect | act_wiz.c:1123 | Force-disconnect a descriptor |
| mortalize | act_wiz.c:6594 | Strip immortal status |

(`unsilence`, `unfreeze`, `wizify` were in the initial audit but either are handled as toggles in Go's existing Silence/Freeze, or don't exist in C.)

### Missing mortal commands (9 confirmed)

| Command | C ref | Purpose |
|---------|-------|---------|
| split | tables.c:1459 (→ act_comm.c) | Share gold with group |
| light | tables.c:796 | Light a torch/lantern |
| throw | skills.c:6355 | Throwing skill (also in skills list above) |
| appraise | tables.c:337 | Appraise item value at shop |
| alias | alias.c:57 | Create command alias (requires load/save of aliases, tie to Tier 1 skill-persistence fix pattern) |
| areas | act_info.c:5846 | List all areas with level ranges and authors |
| altscore | player.c:788 | Alternate score display |
| color / colorscheme | tables.c:475, color.c:307 | Per-AT_ color customization (depends on color expansion; see Tier 1 open questions) |
| compress | mccp.c:187 | Toggle MCCP compression |

### OLC interactivity

Go has `redit` with ~6 subcommands plus flat `mset`/`oset`/`rset`/`aset`. C has:
- `redit` with ~20 subcommands (`src/build.c:5400`): `name`, `desc`, `ed`, `rmed`, `exit`, `bexit`, `exdesc`, `exflags`, `exname`, `exkey`, `flags`, `sector`, `teledelay`, `televnum`, `tunnel`, `rlist`, `exdistance`, `pulltype`, `pull`, `push` — Go missing: `ed`, `rmed`, `bexit`, `exflags`, `exname`, `exkey`, `teledelay`, `televnum`, `tunnel`, `rlist`, `exdistance`, `pulltype`, `pull`, `push` (~14).
- `oedit` (build.c) — interactive editor with `set`, `addaffect`, `delaffect`, `addextra`, `delextra`, `addmprog`, `delmprog`, `type`, `values`, `flags`, `level`, `wearflags`, `weight`, `cost`. Go has only flat `oset`.
- `medit` (build.c) — interactive editor with the analogous set. Go has only flat `mset`.
- `mpedit` / `opedit` / `rpedit` (build.c:9039+) — mudprog script editors.
- `rdelete` / `odelete` / `mdelete` (build.c:10061/10107/10160) — delete with confirmation.
- `foldarea` (build.c:8056) — repack vnum gaps.

## Task groups

### G1 — Damage-message dispatcher (foundation for content quality)

**Why first:** nearly every content skill and spell emits messages; having the dispatcher in place means every subsequent port uses it.

Create `smaug-go/internal/combat/dammessage.go`:
- Port `s_message_table` and `p_message_table` from `src/fight.c`. These are 2D arrays indexed by `[damage-type-index][damage-percent-index]` with verbs like "scratch", "bruise", "hit", "pulverize", etc.
- `func NewDamMessage(ch, victim *types.CharData, dam int, dt int)` — selects the verb, composes messages via Tier 1's `Act()`, emits to char/vict/room.
- `func DamMessage(...)` — shorter public API used by `OneHit` and `Damage`.

Refactor `combat.Damage` and `combat.OneHit` to call `DamMessage` when dam is applied.

**TDD:** `combat/dammessage_test.go` table-driven: dam-type × dam-pct → expected verb; assert string output via a captured descriptor.

### G2 — `spell_smaug` metadata fill (unlocks 55 spells cheap)

Once Tier 1 ships `spell_smaug`, most missing spells become a `skills.dat` metadata exercise rather than a Go code exercise.

- Identify every entry in `db/system/en/skills.dat` whose `Code` is `spell_smaug`. For each, confirm required metadata fields (target/action/class/damage/flag) are populated correctly.
- Add test coverage in `magic/spell_smaug_test.go` with several representative cases (affect, attack, area).
- Track per-spell status in a punch-list table at the bottom of this tier doc when the work lands.

**Acceptance:** at boot, no skill loads with Code=`spell_smaug` that fails at first cast. Measured via automated test that iterates all loaded skills, attempts cast on a fake target, asserts no nil-dispatch panic.

### G3 — Custom-port spells (~25)

Group by complexity:

**Group A — simple affect variants (5):**
- pass_door, farsight, ventriloquate, remove_invis, remove_trap.

**Group B — area / breath attacks (6):**
- acid_breath, fire_breath, frost_breath, gas_breath, lightning_breath, earthquake.
- Reuse `spellAreaAttack` sub-dispatcher from Tier 1's `spell_smaug`.
- Breath-specific side effects: fire melts carried potions, acid decays random worn armor, frost freezes liquid containers, gas poisons, lightning damages random equipment. Port each helper verbatim from C.

**Group C — movement / teleportation (6):**
- astral_walk, gate, portal (as object), mist_walk, transport, word_of_recall, group_teleport, solar_flight, plant_pass.
- All respect ROOM_NORECALL / ROOM_NOSUMMON (Tier 2 guards).

**Group D — unique mechanics (8–10):**
- polymorph (needs morph data; Phase 6 system handles full polymorph — MVP is form-swap stub).
- possess (switch-like — tie to Tier 4 G4 switch/return).
- animate_dead (create undead NPC from corpse object).
- energy_drain (damage + XP drain).
- knock (force-open door).
- recharge (restore wand/staff charges).
- revive (resurrect).
- call_lightning / control_weather (WeatherInfo interaction).
- acid_blast (damage + corrode worn armor).

**TDD per spell:** one test each in `magic/magic_test.go`. Before implementing, write a test for the expected side effect; follow Tier 1's per-save and per-affect test pattern.

### G4 — Combat / utility skills (~28)

Port in priority order:

1. **Unarmed attacks (5):** bite, claw, punch, sting, tail — share a common `unarmedAttack` helper. Hook into combat as alternate attacks when an NPC with `Attacks` bits set uses them.
2. **Offensive combat (7):** circle, gouge, stun, grapple, cleave, hitall, berserk.
3. **Ranged primer (2):** throw, fire. (Full archery is a Phase-6 candidate; MVP wires the skill to existing throw mechanics.)
4. **Utility (7):** meditate, trance, skin, search, detrap, dig, visible.
5. **Crafting (3):** scribe, cook, feed.
6. **Admin (2):** slookup, sset (immortal commands).
7. **Combat modifiers (2):** stance, style (lightweight toggles for now).
8. **Niche (1):** mistwalk (short-range teleport; shares mechanics with Group C spells).

Every new skill uses:
- Tier 1's corrected `learnFromSuccess`/`learnFromFailure`.
- Tier 1's `Act()` for messages.
- Existing skill registration pattern from `cmd/smaug/main.go`.

**Acceptance:** per skill: one functional test, one failure-path test.

### G5 — Immortal commands (14)

Port into `act/wiz.go` (or `act/wiz2.go` if the file gets large). Each has a small, well-understood shape in C:

- **switch / return:** possess an NPC by redirecting the immortal's descriptor to the NPC's `CharData`. Return restores the original `CharData`. Single global per-descriptor swap suffices (single-threaded game loop).
- **wizlock:** toggle `SysData.Wizlock` — checked on new logins in `game/loop.go`.
- **shutdown / reboot:** set a graceful-shutdown flag on the game loop; loop drains pulses then exits after saving all players.
- **wizhelp:** list commands with `Level >= LEVEL_IMMORTAL` from `CmdRegistry`, pager-aware.
- **aecho:** area-wide echo (similar to existing `echo` but filtered by area).
- **hell:** set a "hell" timer on a player, move them to a hell room.
- **log:** toggle a `PLR_LOG` flag (defined in types/constants.go) — consumed by command interpreter to log each command for that player.
- **mpstat / opstat / rpstat:** display a mob/obj/room's mudprog list; pager-aware.
- **deny / pardon:** set/clear PLR_DENY and refuse login.
- **disconnect:** close a named player's descriptor, preserving their save.
- **mortalize:** set `ch.Trust = 0`, strip immortal flags.

**TDD:** `act/wiz_test.go` per command. Smoke each end-to-end via descriptor buffer capture.

### G6 — Mortal commands (9)

- **split:** divide gold among group members (iterate `ch` + followers, transfer share).
- **light:** find `ITEM_LIGHT` in inventory, set Value[2] > 0, set `ch.InRoom.Light` +1 via existing room-light tracking.
- **throw:** port from C — range 1-room throw, object to target.
- **appraise:** shop-side query. Tier 2's `DoAppraise` may already cover the repair case; generalize to show shop buy price for carried items.
- **alias:** persistent per-player aliases stored in `PCData.Aliases` (requires Tier 1's save-path extension to cover alias lines — fold into the skills-persistence work or do separately here).
- **areas:** list `WorldRef.Areas` with name, author, min/max vnum, reset freq. Pager-aware.
- **altscore:** compact/alt-format score display — simple display variant of `DoScore`.
- **color / colorscheme:** per-AT_ color customization. Requires Tier 1's color-slot expansion. If Tier 1's color work stays minimal, this is a follow-up.
- **compress:** toggle MCCP on descriptor; `net/protocol.go` has `MCCPEnabled` — expose via command.

**TDD:** per command in `act/*_test.go`.

### G7 — OLC interactivity

**`redit` subcommand expansion:** add to `act/olc.go` dispatch:
- `ed <keyword>` / `rmed <keyword>` — add/remove extra descriptions on the room (String Editor integration — reuses `StartEditingFunc` from `act/olc.go:17`).
- `bexit <dir>` — bidirectional exit create (digs and links reverse).
- `exflags <dir> <flag>` — set exit flags (EX_CLOSED/LOCKED/PICKPROOF/HIDDEN/SECRET/NOMOB).
- `exname <dir> <name>` — set exit keyword.
- `exkey <dir> <vnum>` — set key vnum on exit.
- `teledelay <n>` / `televnum <vnum>` — teleport room config.
- `tunnel <n>` — capacity limit.
- `rlist` — inline area room listing (delegate to existing `DoRlist`).
- `exdistance <dir> <n>` — distance for overland overflow rooms.
- `pulltype / pull / push <dir>` — lever/switch metadata.

**Interactive `oedit` / `medit`:** create `act/olc_oedit.go` and `act/olc_medit.go`. These are command-loops that take a target vnum and enter an editing substate (analogous to existing `CON_EDITING` — introduce `CON_OEDITING`/`CON_MEDITING` in `types/enums.go` for state).

Subcommand vocabulary (per C build.c):
- oedit: `name`, `short`, `long`, `type`, `flags`, `wearflags`, `values <0-5> <n>`, `weight`, `cost`, `level`, `affects <add/rem>`, `ed <kw>`, `mpedit <trig> <args>`.
- medit: `name`, `short`, `long`, `desc`, `level`, `stats <str/int/wis/dex/con/cha/lck> <n>`, `hp/mana/mv`, `armor`, `hitroll`, `damroll`, `align`, `flags`, `race`, `class`, `sex`, `attacks`, `defenses`, `gold`, `xp`, `position`, `affected`, `mpedit <trig> <args>`.

**`mpedit` / `opedit` / `rpedit`:** mudprog script editors. New `act/olc_prog.go`. Each takes a mob/obj/room vnum + trigger type and enters the string editor seeded with the existing script; save writes back.

**`rdelete` / `odelete` / `mdelete`:** delete with confirmation prompt. Extract from existing mob/obj/room index maps; also scrub any live instances (extract via existing handler APIs).

**`foldarea`:** repack vnum gaps into a contiguous range. Low priority — document as a stretch goal.

**TDD:** extend `act/olc_test.go` + new `olc_oedit_test.go` / `olc_medit_test.go` / `olc_prog_test.go`. Each subcommand gets a round-trip test that mutates state and verifies via `astat`/`rstat`/`ostat`/`mstat`.

## Critical files

**Create:**
- `smaug-go/internal/combat/dammessage.go` + `dammessage_test.go`.
- `smaug-go/internal/magic/spell_breath.go` (+test).
- `smaug-go/internal/magic/spell_teleport_ext.go` or similar for Group C (+test).
- `smaug-go/internal/magic/spell_unique.go` for Group D (+test).
- `smaug-go/internal/act/skills3.go` (if `skills.go` + `skills2.go` gets crowded) + `skills3_test.go`.
- `smaug-go/internal/act/olc_oedit.go`, `olc_medit.go`, `olc_prog.go` (+tests).
- `smaug-go/internal/act/wiz2.go` if needed for new immortal commands (+test).

**Modify:**
- `smaug-go/internal/combat/combat.go` — wire `DamMessage` into `OneHit` / `Damage`.
- `smaug-go/internal/magic/magic.go` — spellRegistry registrations for each new custom spell.
- `smaug-go/internal/persist/skills.go` — ensure `spell_smaug` metadata parses every required field (Tier 1 began this — verify coverage here).
- `smaug-go/internal/act/olc.go` — new `redit` subcommand arms.
- `smaug-go/internal/game/loop.go` — nanny states for `CON_OEDITING` / `CON_MEDITING` / `CON_MPEDIT` etc.
- `smaug-go/internal/act/wiz.go` — if extending in place.
- `smaug-go/internal/types/enums.go` — new CON_ states; `PLR_DENY`, `PLR_LOG` flag aliases if missing.
- `smaug-go/cmd/smaug/main.go` — register ~23 new commands (14 imm + 9 mort) + `DoSkin/DoScribe/DoCook/DoThrow/...` etc.

**Reference (C):**
- `src/magic.c` per-spell implementations.
- `src/skills.c` + `src/fight.c` for skill implementations.
- `src/act_wiz.c` for immortal commands.
- `src/build.c` for OLC subcommands and interactive editor structure.
- `src/tables.c` / `src/interp.c` for argument parsing conventions.

## Reused utilities

- Tier 1's `Act()` — mandatory for every new message-emitting path.
- Tier 1's `spell_smaug` — carries most metadata-driven spells.
- Tier 1's fixed `learnFromSuccess/Failure` — used by every new skill.
- Tier 3's oprog/rprog trigger helpers — OLC `mpedit` / `opedit` / `rpedit` edit the same data structures.
- Existing `StartEditingFunc` (`act/olc.go:17`) — reuse for string-editor subcommands in OLC.
- Existing pager (`game/pager.go`) — reuse for `areas`, `wizhelp`, `mstat`/`ostat`/`rstat`.
- Existing `combat.Damage`, `handler.CharToRoom`/`CharFromRoom` for switch/return and damage paths.
- Existing `CmdRegistry.InterpretWithTrustCap` (Tier 1 security hardening) for `DoForce` analogs.

## Verification

1. `go test ./...` passes. Tier 4 alone should add ≥400 test cases.
2. Smoke: boot with an unmodified `db/` and confirm:
   - Cast any `spell_smaug`-backed spell from `skills.dat` → works.
   - Use `bite`, `cleave`, `gouge` against an NPC → damage messages are flavored.
   - `wizhelp` lists immortal commands with the correct level gates.
   - `alias g get all` + `g` → resolves to `get all`.
   - `areas` shows area names with their vnum ranges.
   - Enter `oedit` on a carried object, change a flag, save → inspect with `ostat`, confirm change.
3. Mutation verify representative skills and spells per CLAUDE.md convention.

## Open questions / follow-ups

- **Archery, stances, polymorph full systems** remain Phase-6 scope. Tier 4 ships the *skill* entries that hook into eventual full systems (throw/fire skill, stance toggle, polymorph spell stub).
- **Alias persistence** pattern can piggyback on Tier 1's skill-persistence loop in `SavePlayer`.
- **Color per-AT_ customization (`color` command)** depends on whether Tier 1 expands color-slot support. If not, ship `color` as a `compact`-style on/off toggle for MVP and file full slot customization as Phase-6 polish.
- **Switch/return + possess spell (G3 Group D)** share core plumbing — port switch first, then possess layers on top.
- **`foldarea`** (area vnum repack) is low-reward and high-risk; defer to Phase 6 builder tooling.
