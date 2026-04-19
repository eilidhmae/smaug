//go:build !windows

// Hotboot world-state persistence. Ports src/hotboot.c:73-596.
// save_mobile / load_mobile → SaveMob / LoadMob.
// save_world / load_world → SaveWorld / LoadWorld.
//
// Files live in <dir>/hotboot/ (auto-created on save):
//   - mobfile.dat: all NPCs, one #MOBILE section each, terminated by #END.
//   - <vnum>.objdat: floor objects per non-empty, non-clanstore room,
//     each file a sequence of #OBJECT sections terminated by #END.
//
// On LoadWorld, both file kinds are unlinked after a successful parse
// (C parity — prevents a crash-during-recovery from looping).
//
// Build-tagged !windows because the broader hotboot subsystem is Unix-only.
// Sibling hotboot_windows.go stubs the same exports to hard-error.

package persist

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/eilidhmae/smaug/internal/types"
	"github.com/eilidhmae/smaug/internal/util"
	"github.com/eilidhmae/smaug/internal/world"
)

// SaveMob writes one #MOBILE section for mob. Skips PCs, ACT_PROTOTYPE,
// and ACT_PET to match src/hotboot.c:73-79,193-194.
func SaveMob(w io.Writer, mob *types.CharData) error {
	if mob == nil {
		return nil
	}
	// PC gate — C uses !IS_NPC(). In Go, a PC is any char with no IndexData
	// OR PCData attached. Match the semantics by excluding anything that
	// isn't clearly an NPC: require IndexData AND ACT_IS_NPC.
	if mob.PCData != nil || mob.IndexData == nil || !mob.Act.IsSet(types.ACT_IS_NPC) {
		return nil
	}
	if mob.Act.IsSet(types.ACT_PROTOTYPE) {
		return nil
	}
	if mob.Act.IsSet(types.ACT_PET) {
		return nil
	}

	idx := mob.IndexData
	if _, err := fmt.Fprintf(w, "#MOBILE\n"); err != nil {
		return err
	}
	fmt.Fprintf(w, "Vnum %d\n", idx.Vnum)
	fmt.Fprintf(w, "Level %d\n", mob.Level)
	fmt.Fprintf(w, "Gold %d\n", mob.Gold)

	// Room — sentinel mobs record HomeVnum (C hotboot.c:86-95).
	roomVnum := types.ROOM_VNUM_LIMBO
	if mob.InRoom != nil {
		if mob.Act.IsSet(types.ACT_SENTINEL) && mob.HomeVnum != 0 {
			roomVnum = mob.HomeVnum
		} else {
			roomVnum = mob.InRoom.Vnum
		}
	}
	fmt.Fprintf(w, "Room %d\n", roomVnum)

	// Name / Short / Long / Description — only emit when different from
	// prototype (C hotboot.c:104-111 str_cmp gate).
	if mob.Name != "" && idx.PlayerName != "" && mob.Name != idx.PlayerName {
		fmt.Fprintf(w, "Name %s~\n", util.SmashTilde(mob.Name))
	}
	if mob.ShortDescr != "" && idx.ShortDescr != "" && mob.ShortDescr != idx.ShortDescr {
		fmt.Fprintf(w, "Short %s~\n", util.SmashTilde(mob.ShortDescr))
	}
	if mob.LongDescr != "" && idx.LongDescr != "" && mob.LongDescr != idx.LongDescr {
		fmt.Fprintf(w, "Long %s~\n", util.SmashTilde(mob.LongDescr))
	}
	if mob.Description != "" && idx.Description != "" && mob.Description != idx.Description {
		fmt.Fprintf(w, "Description %s~\n", util.SmashTilde(mob.Description))
	}

	fmt.Fprintf(w, "HpManaMove %d %d %d %d %d %d\n",
		mob.Hit, mob.MaxHit, mob.Mana, mob.MaxMana, mob.Move, mob.MaxMove)
	fmt.Fprintf(w, "Position %d\n", mob.Position)
	fmt.Fprintf(w, "Flags %s\n", mob.Act.String())
	if !mob.AffectedBy.IsEmpty() {
		fmt.Fprintf(w, "AffectedBy %s\n", mob.AffectedBy.String())
	}

	for _, paf := range mob.Affects {
		if paf == nil {
			continue
		}
		fmt.Fprintf(w, "Affect %d %d %d %d %s\n",
			paf.Type, paf.Duration, paf.Modifier, paf.Location, paf.BitVector.String())
	}

	// Inventory via existing player-object writer (nest 0 at this level;
	// writePlayerObj handles recursion into containers).
	for _, obj := range mob.Carrying {
		writePlayerObj(w, obj, 0)
	}

	fmt.Fprintf(w, "EndMobile\n\n")
	return nil
}

// LoadMob reads one #MOBILE section. The opening "#MOBILE" word must
// already have been consumed by the caller (parallel to load_obj's
// letter-dispatch convention in C). Returns nil on bad vnum or
// unparseable content.
func LoadMob(w *world.World, sc *Scanner) *types.CharData {
	if w == nil || sc == nil {
		return nil
	}

	// First keyword must be Vnum.
	word := sc.ReadWord()
	if word != "Vnum" {
		util.Bug("LoadMob: expected Vnum, got %q at %s:%d", word, sc.File(), sc.Line())
		// Skip until EndMobile to stay in sync.
		skipUntilEndMobile(sc)
		return nil
	}
	vnum := sc.ReadNumber()
	idx := w.GetMobIndex(vnum)
	if idx == nil {
		util.Bug("LoadMob: no index data for vnum %d", vnum)
		skipUntilEndMobile(sc)
		return nil
	}

	mob := &types.CharData{
		Name:        idx.PlayerName,
		ShortDescr:  idx.ShortDescr,
		LongDescr:   idx.LongDescr,
		Description: idx.Description,
		IndexData:   idx,
		Level:       idx.Level,
		Sex:         idx.Sex,
		Race:        idx.Race,
		Class:       idx.Class,
		Alignment:   idx.Alignment,
		Position:    idx.Position,
		DefPosition: idx.DefPosition,
	}
	mob.Act.Set(types.ACT_IS_NPC)
	idx.Count++

	inroom := 0
	for {
		word = sc.ReadWord()
		if word == "" || word == "EndMobile" {
			break
		}
		switch word {
		case "Level":
			mob.Level = sc.ReadNumber()
		case "Gold":
			mob.Gold = sc.ReadNumber()
		case "Room":
			inroom = sc.ReadNumber()
		case "Name":
			mob.Name = sc.ReadString()
		case "Short":
			mob.ShortDescr = sc.ReadString()
		case "Long":
			mob.LongDescr = sc.ReadString()
		case "Description":
			mob.Description = sc.ReadString()
		case "HpManaMove":
			mob.Hit = sc.ReadNumber()
			mob.MaxHit = sc.ReadNumber()
			mob.Mana = sc.ReadNumber()
			mob.MaxMana = sc.ReadNumber()
			mob.Move = sc.ReadNumber()
			mob.MaxMove = sc.ReadNumber()
			if mob.MaxMove <= 0 {
				mob.MaxMove = 150
			}
		case "Position":
			mob.Position = sc.ReadNumber()
		case "Flags":
			bv, _ := types.ParseBitVector(sc.ReadToEOL())
			mob.Act = bv
			mob.Act.Set(types.ACT_IS_NPC)
		case "AffectedBy":
			bv, _ := types.ParseBitVector(sc.ReadToEOL())
			mob.AffectedBy = bv
		case "Affect", "AffectData":
			paf := &types.AffectData{}
			if word == "Affect" {
				paf.Type = sc.ReadNumber()
			} else {
				// AffectData uses a skill-name token for paf.Type; we
				// don't need perfect C parity here (Go hotboot files
				// are written by our own SaveMob which always emits
				// "Affect"), but accept the skill-name form for
				// forward compat.
				_ = sc.ReadWord()
				paf.Type = 0
			}
			paf.Duration = sc.ReadNumber()
			paf.Modifier = sc.ReadNumber()
			paf.Location = sc.ReadNumber()
			paf.BitVector = sc.ReadBitvector()
			mob.Affects = append(mob.Affects, paf)
		case "#OBJECT", "#CORPSE":
			// Inventory object. Use existing player-object reader.
			const maxNest = 100
			var nestObj [maxNest]*types.ObjData
			obj := readPlayerObject(sc, func(vnum int) *types.ObjIndexData {
				return w.GetObjIndex(vnum)
			}, nestObj[:])
			if obj != nil && obj.InObj == nil {
				mob.Carrying = append(mob.Carrying, obj)
				obj.CarriedBy = mob
			}
		case "End":
			// Stray End from nested object — ignore (C parity hotboot.c:355-356).
		default:
			sc.ReadToEOL()
		}
	}

	// Place in room. Falls back to Limbo if saved vnum unknown.
	if inroom == 0 {
		inroom = types.ROOM_VNUM_LIMBO
	}
	room := w.GetRoom(inroom)
	if room == nil {
		util.Bug("LoadMob: room vnum %d not found, falling back to LIMBO", inroom)
		room = w.GetRoom(types.ROOM_VNUM_LIMBO)
	}
	if room != nil {
		mob.InRoom = room
		room.People = append(room.People, mob)
	}
	w.AddChar(mob)
	return mob
}

func skipUntilEndMobile(sc *Scanner) {
	for {
		word := sc.ReadWord()
		if word == "" || word == "EndMobile" {
			return
		}
		// Don't call ReadToEOL on a word we already consumed — Scanner
		// ReadWord is self-terminating. Loop just reads tokens.
	}
}

// SaveWorld writes <dir>/hotboot/mobfile.dat plus one <vnum>.objdat per
// non-empty, non-clanstore room. Auto-creates <dir>/hotboot/ if missing.
func SaveWorld(w *world.World, dir string) error {
	if w == nil {
		return fmt.Errorf("SaveWorld: nil world")
	}
	hotbootDir := filepath.Join(dir, "hotboot")
	if err := os.MkdirAll(hotbootDir, 0o755); err != nil {
		return fmt.Errorf("SaveWorld: mkdir %s: %w", hotbootDir, err)
	}

	// --- Per-room object files ---
	for _, room := range w.Rooms {
		if room == nil || len(room.Contents) == 0 {
			continue
		}
		if room.RoomFlags.IsSet(types.ROOM_CLANSTOREROOM) {
			continue
		}
		path := filepath.Join(hotbootDir, fmt.Sprintf("%d.objdat", room.Vnum))
		// 0o600 — hotboot transient files contain player/NPC state and
		// should not be world-readable. os.Create would use 0o666 &
		// umask, landing at 0o644 on typical deployments.
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
		if err != nil {
			util.Bug("SaveWorld: create %s: %v", path, err)
			continue
		}
		for _, obj := range room.Contents {
			writePlayerObj(f, obj, 0)
		}
		fmt.Fprintf(f, "#END\n")
		f.Close()
	}

	// --- Mob file ---
	mobPath := filepath.Join(hotbootDir, "mobfile.dat")
	// 0o600 — see objdat block above for rationale.
	f, err := os.OpenFile(mobPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("SaveWorld: create %s: %w", mobPath, err)
	}
	defer f.Close()
	for _, c := range w.Characters {
		if c == nil {
			continue
		}
		if err := SaveMob(f, c); err != nil {
			return err
		}
	}
	fmt.Fprintf(f, "#END\n")
	return nil
}

// LoadWorld reads <dir>/hotboot/mobfile.dat (restoring NPCs) and every
// <vnum>.objdat (restoring floor objects). Unlinks each file after a
// successful read so a crash during recovery doesn't loop.
func LoadWorld(w *world.World, dir string) error {
	if w == nil {
		return fmt.Errorf("LoadWorld: nil world")
	}
	hotbootDir := filepath.Join(dir, "hotboot")

	// --- Mob file ---
	mobPath := filepath.Join(hotbootDir, "mobfile.dat")
	if f, err := os.Open(mobPath); err == nil {
		sc := NewScanner(f, mobPath)
		for {
			word := sc.ReadWord()
			if word == "" || word == "#END" {
				break
			}
			if word == "#MOBILE" {
				_ = LoadMob(w, sc)
				continue
			}
			util.Bug("LoadWorld: unexpected token %q in %s", word, mobPath)
		}
		f.Close()
		if err := os.Remove(mobPath); err != nil {
			util.Bug("LoadWorld: remove %s: %v", mobPath, err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("LoadWorld: open %s: %w", mobPath, err)
	}

	// --- Per-room object files ---
	entries, err := os.ReadDir(hotbootDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("LoadWorld: readdir: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".objdat") {
			continue
		}
		path := filepath.Join(hotbootDir, name)
		if err := loadObjFile(w, path); err != nil {
			util.Bug("LoadWorld: %s: %v", path, err)
		}
		// Unlink after load (C parity).
		if err := os.Remove(path); err != nil {
			util.Bug("LoadWorld: remove %s: %v", path, err)
		}
	}
	return nil
}

// loadObjFile parses a single <vnum>.objdat into its room's Contents.
func loadObjFile(w *world.World, path string) error {
	base := filepath.Base(path)
	vnumStr := strings.TrimSuffix(base, ".objdat")
	vnum, err := ParseVnum(vnumStr)
	if err != nil {
		return fmt.Errorf("bad filename %q: %w", base, err)
	}
	room := w.GetRoom(vnum)
	if room == nil {
		return fmt.Errorf("no room index for vnum %d", vnum)
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sc := NewScanner(f, path)
	const maxNest = 100
	var nestObj [maxNest]*types.ObjData
	lookup := func(v int) *types.ObjIndexData { return w.GetObjIndex(v) }

	for {
		word := sc.ReadWord()
		if word == "" || word == "#END" {
			break
		}
		if word == "#OBJECT" || word == "#CORPSE" {
			obj := readPlayerObject(sc, lookup, nestObj[:])
			if obj != nil && obj.InObj == nil {
				obj.InRoom = room
				room.Contents = append(room.Contents, obj)
				w.AddObj(obj)
			}
			continue
		}
		util.Bug("loadObjFile: unexpected token %q in %s", word, path)
	}
	return nil
}
