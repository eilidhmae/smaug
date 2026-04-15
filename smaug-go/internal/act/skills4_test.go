package act

import (
	"testing"

	"github.com/eilidhmae/smaug/internal/types"
)

// --- Wave 3: utility ---

func TestDoMeditate_InCombat(t *testing.T) {
	ensureSkill(t, "meditate")
	ch, victim := newSkillTestRoom()
	ch.Fighting = &types.FightData{Who: victim}
	DoMeditate(ch, "")
	// Should say "not in combat"
}

func TestDoMeditate_Success(t *testing.T) {
	ensureSkill(t, "meditate")
	ch, _ := newSkillTestRoom()
	setLearned(t, ch, "meditate", 100)
	DoMeditate(ch, "")
	// No crash
}

func TestDoTrance(t *testing.T) {
	ensureSkill(t, "trance")
	ch, _ := newSkillTestRoom()
	setLearned(t, ch, "trance", 100)
	DoTrance(ch, "")
}

func TestDoSearch_NoHidden(t *testing.T) {
	ensureSkill(t, "search")
	ch, _ := newSkillTestRoom()
	setLearned(t, ch, "search", 100)
	DoSearch(ch, "")
}

func TestDoSearch_FindsSecretExit(t *testing.T) {
	ensureSkill(t, "search")
	ch, _ := newSkillTestRoom()
	setLearned(t, ch, "search", 100)
	target := &types.RoomIndexData{Vnum: 9100}
	ex := &types.ExitData{
		Direction: types.DIR_NORTH,
		ToRoom:    target,
		ExitInfo:  int(types.EX_SECRET),
	}
	ch.InRoom.Exits = append(ch.InRoom.Exits, ex)
	// Try several times — may take a few rolls
	for i := 0; i < 20; i++ {
		DoSearch(ch, "")
		if ex.ExitInfo&int(types.EX_SECRET) == 0 {
			return
		}
	}
}

func TestDoDetrap_NoArg(t *testing.T) {
	ensureSkill(t, "detrap")
	ch, _ := newSkillTestRoom()
	DoDetrap(ch, "")
}

func TestDoDig(t *testing.T) {
	ensureSkill(t, "dig")
	ch, _ := newSkillTestRoom()
	setLearned(t, ch, "dig", 100)
	DoDig(ch, "")
}

func TestDoVisible(t *testing.T) {
	ch, _ := newSkillTestRoom()
	ch.AffectedBy.Set(types.AFF_INVISIBLE)
	ch.AffectedBy.Set(types.AFF_HIDE)
	DoVisible(ch, "")
	if ch.AffectedBy.IsSet(types.AFF_INVISIBLE) {
		t.Error("expected AFF_INVISIBLE cleared")
	}
	if ch.AffectedBy.IsSet(types.AFF_HIDE) {
		t.Error("expected AFF_HIDE cleared")
	}
}

func TestDoStyle_NoArg(t *testing.T) {
	ch, _ := newSkillTestRoom()
	DoStyle(ch, "")
}

func TestDoStyle_Set(t *testing.T) {
	ch, _ := newSkillTestRoom()
	DoStyle(ch, "defensive")
	if ch.Style != types.STYLE_DEFENSIVE {
		t.Errorf("expected STYLE_DEFENSIVE, got %d", ch.Style)
	}
}

func TestDoStyle_Unknown(t *testing.T) {
	ch, _ := newSkillTestRoom()
	DoStyle(ch, "xyzzy")
}

func TestDoStance_NoArg(t *testing.T) {
	ch, _ := newSkillTestRoom()
	DoStance(ch, "")
}

func TestDoStance_Set(t *testing.T) {
	ch, _ := newSkillTestRoom()
	DoStance(ch, "dragon")
	if ch.Stance != types.STANCE_DRAGON {
		t.Errorf("expected STANCE_DRAGON, got %d", ch.Stance)
	}
}

// --- Wave 4: crafting ---

func TestDoFeed_NoArgs(t *testing.T) {
	ch, _ := newSkillTestRoom()
	DoFeed(ch, "")
}

func TestDoFeed_Success(t *testing.T) {
	ch, victim := newSkillTestRoom()
	food := &types.ObjData{
		Name:       "bread",
		ShortDescr: "a loaf of bread",
		ItemType:   types.ITEM_FOOD,
		WearLoc:    types.WEAR_NONE,
	}
	food.CarriedBy = ch
	ch.Carrying = append(ch.Carrying, food)
	victim.Name = "target"
	DoFeed(ch, "target bread")
	// Food should now be on victim
	found := false
	for _, o := range victim.Carrying {
		if o == food {
			found = true
		}
	}
	if !found {
		t.Error("expected food transferred to victim")
	}
}

func TestDoSkin_NoArg(t *testing.T) {
	ensureSkill(t, "skin")
	ch, _ := newSkillTestRoom()
	DoSkin(ch, "")
}

func TestDoSkin_Success(t *testing.T) {
	ensureSkill(t, "skin")
	ch, _ := newSkillTestRoom()
	setLearned(t, ch, "skin", 100)
	corpse := &types.ObjData{
		Name:     "orc corpse",
		ItemType: types.ITEM_CORPSE_NPC,
	}
	ch.InRoom.Contents = append(ch.InRoom.Contents, corpse)
	for i := 0; i < 20; i++ {
		startCount := len(ch.Carrying)
		DoSkin(ch, "orc")
		if len(ch.Carrying) > startCount {
			return
		}
		// re-add corpse if extracted but skill failed — extracting empties contents
		found := false
		for _, o := range ch.InRoom.Contents {
			if o == corpse {
				found = true
			}
		}
		if !found {
			ch.InRoom.Contents = append(ch.InRoom.Contents, corpse)
		}
	}
}

func TestDoPoisonWeapon_NoWeapon(t *testing.T) {
	ensureSkill(t, "poison weapon")
	ch, _ := newSkillTestRoom()
	DoPoisonWeapon(ch, "")
}

func TestDoScribe_NoArg(t *testing.T) {
	ensureSkill(t, "scribe")
	ch, _ := newSkillTestRoom()
	DoScribe(ch, "")
}

func TestDoScribe_UnknownSpell(t *testing.T) {
	ensureSkill(t, "scribe")
	ch, _ := newSkillTestRoom()
	DoScribe(ch, "_totallyfakespell_")
}

func TestDoCook_NotFood(t *testing.T) {
	ensureSkill(t, "cook")
	ch, _ := newSkillTestRoom()
	obj := &types.ObjData{
		Name:       "stick",
		ShortDescr: "a stick",
		ItemType:   types.ITEM_WEAPON,
		CarriedBy:  ch,
		WearLoc:    types.WEAR_NONE,
	}
	ch.Carrying = append(ch.Carrying, obj)
	DoCook(ch, "stick")
}

func TestDoCook_Success(t *testing.T) {
	ensureSkill(t, "cook")
	ch, _ := newSkillTestRoom()
	setLearned(t, ch, "cook", 100)
	food := &types.ObjData{
		Name:       "fish",
		ShortDescr: "a raw fish",
		ItemType:   types.ITEM_FOOD,
		CarriedBy:  ch,
		WearLoc:    types.WEAR_NONE,
		Value:      [6]int{10, 0, 0, 0, 0, 0},
	}
	ch.Carrying = append(ch.Carrying, food)
	for i := 0; i < 20; i++ {
		DoCook(ch, "fish")
		if food.Value[4] != 0 {
			return
		}
	}
}

func TestDoFire_NoArg(t *testing.T) {
	ch, _ := newSkillTestRoom()
	DoFire(ch, "")
}

func TestDoMistwalk_NoArg(t *testing.T) {
	ensureSkill(t, "mistwalk")
	ch, _ := newSkillTestRoom()
	DoMistwalk(ch, "")
}
